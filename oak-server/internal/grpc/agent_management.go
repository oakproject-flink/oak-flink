package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	oakv1 "github.com/oakproject-flink/oak-flink/api/proto/oak/v1"
	"github.com/oakproject-flink/oak-flink/oak-lib/logger"
	"github.com/oakproject-flink/oak-flink/oak-lib/certs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AgentManagementService implements the AgentManagement gRPC service
type AgentManagementService struct {
	oakv1.UnimplementedAgentManagementServer

	certManager *certs.Manager
	logger      *logger.Logger

	// In-memory storage for now (TODO: replace with database)
	// Key: cluster_id, Value: agent state
	agents map[string]*AgentState
}

// AgentState represents the current state of an agent
type AgentState struct {
	ClusterID         string
	ClusterName       string
	AgentID           string
	AgentSecret       string
	Status            oakv1.StatusResponse_Status
	ClientCertPEM     []byte
	ClientKeyPEM      []byte
	AgentVersion      string
	KubernetesVersion string
}

// NewAgentManagementService creates a new agent management service
func NewAgentManagementService(certManager *certs.Manager) *AgentManagementService {
	return &AgentManagementService{
		certManager: certManager,
		logger:      logger.NewComponent("registration"),
		agents:      make(map[string]*AgentState),
	}
}

// RequestCredentials handles agent credential requests
func (s *AgentManagementService) RequestCredentials(ctx context.Context, req *oakv1.CredentialsRequest) (*oakv1.CredentialsResponse, error) {
	s.logger.Infof("Credential request from cluster: %s (%s)", req.ClusterName, req.ClusterId)

	// Validate request
	if req.ClusterId == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id is required")
	}
	if req.ClusterName == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_name is required")
	}

	// Check if agent already exists
	if agent, exists := s.agents[req.ClusterId]; exists {
		switch agent.Status {
		case oakv1.StatusResponse_STATUS_APPROVED:
			// Already approved, return existing credentials
			s.logger.Infof("Agent %s already approved, returning existing credentials", req.ClusterId)
			return &oakv1.CredentialsResponse{
				Result: &oakv1.CredentialsResponse_Approved{
					Approved: &oakv1.ApprovedCredentials{
						AgentId:       agent.AgentID,
						AgentSecret:   agent.AgentSecret,
						ClientCertPem: agent.ClientCertPEM,
						ClientKeyPem:  agent.ClientKeyPEM,
						CaCertPem:     s.certManager.GetCACert(),
					},
				},
			}, nil

		case oakv1.StatusResponse_STATUS_PENDING:
			// Still pending approval
			s.logger.Infof("Agent %s still pending approval", req.ClusterId)
			return &oakv1.CredentialsResponse{
				Result: &oakv1.CredentialsResponse_Pending{
					Pending: &oakv1.PendingApproval{
						Message:              "Your request is pending admin approval. Please check back later.",
						PollIntervalSeconds:  30,
					},
				},
			}, nil

		case oakv1.StatusResponse_STATUS_REJECTED:
			// Rejected by admin
			s.logger.Warnf("Agent %s was rejected", req.ClusterId)
			return &oakv1.CredentialsResponse{
				Result: &oakv1.CredentialsResponse_Rejected{
					Rejected: &oakv1.RejectedRequest{
						Reason: "Your request was rejected by an administrator.",
					},
				},
			}, nil

		case oakv1.StatusResponse_STATUS_REVOKED:
			// Credentials were revoked
			s.logger.Warnf("Agent %s credentials were revoked", req.ClusterId)
			return &oakv1.CredentialsResponse{
				Result: &oakv1.CredentialsResponse_Rejected{
					Rejected: &oakv1.RejectedRequest{
						Reason: "Your credentials were revoked. Contact an administrator.",
					},
				},
			}, nil
		}
	}

	// New agent - check if API token provided
	if req.ApiToken != "" {
		// TODO: Validate API token against database
		// For now, accept any token as valid (DEV MODE)
		s.logger.Warnf("DEV MODE: Auto-approving agent with api_token: %s", req.ClusterId)
		return s.approveAgent(req)
	}

	// No API token - create pending entry
	s.logger.Infof("Creating pending approval entry for agent: %s", req.ClusterId)
	s.agents[req.ClusterId] = &AgentState{
		ClusterID:         req.ClusterId,
		ClusterName:       req.ClusterName,
		Status:            oakv1.StatusResponse_STATUS_PENDING,
		AgentVersion:      req.AgentVersion,
		KubernetesVersion: req.KubernetesVersion,
	}

	// TODO: Trigger notification to admins (webhook, email, etc.)

	return &oakv1.CredentialsResponse{
		Result: &oakv1.CredentialsResponse_Pending{
			Pending: &oakv1.PendingApproval{
				Message:              fmt.Sprintf("Agent '%s' is awaiting admin approval. Please check back in 30 seconds.", req.ClusterName),
				PollIntervalSeconds:  30,
			},
		},
	}, nil
}

// CheckStatus allows agents to poll for approval status
func (s *AgentManagementService) CheckStatus(ctx context.Context, req *oakv1.StatusRequest) (*oakv1.StatusResponse, error) {
	if req.ClusterId == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id is required")
	}

	agent, exists := s.agents[req.ClusterId]
	if !exists {
		return &oakv1.StatusResponse{
			Status:  oakv1.StatusResponse_STATUS_UNKNOWN,
			Message: "No record found for this cluster_id",
		}, nil
	}

	resp := &oakv1.StatusResponse{
		Status:  agent.Status,
	}

	switch agent.Status {
	case oakv1.StatusResponse_STATUS_APPROVED:
		resp.Credentials = &oakv1.ApprovedCredentials{
			AgentId:       agent.AgentID,
			AgentSecret:   agent.AgentSecret,
			ClientCertPem: agent.ClientCertPEM,
			ClientKeyPem:  agent.ClientKeyPEM,
			CaCertPem:     s.certManager.GetCACert(),
		}
		resp.Message = "Agent approved and ready to connect"

	case oakv1.StatusResponse_STATUS_PENDING:
		resp.Message = "Awaiting admin approval"

	case oakv1.StatusResponse_STATUS_REJECTED:
		resp.Message = "Agent request was rejected"

	case oakv1.StatusResponse_STATUS_REVOKED:
		resp.Message = "Agent credentials were revoked"
	}

	return resp, nil
}

// approveAgent generates credentials and approves the agent
func (s *AgentManagementService) approveAgent(req *oakv1.CredentialsRequest) (*oakv1.CredentialsResponse, error) {
	// Generate agent ID and secret
	agentID := uuid.New().String()
	agentSecret := uuid.New().String() // TODO: Use crypto/rand for production

	// Generate client certificate for mTLS
	clientCertPEM, clientKeyPEM, err := s.certManager.GenerateClientCert(agentID)
	if err != nil {
		s.logger.Errorf("Failed to generate client cert for %s: %v", req.ClusterId, err)
		return nil, status.Error(codes.Internal, "failed to generate client certificate")
	}

	// Store agent state
	s.agents[req.ClusterId] = &AgentState{
		ClusterID:         req.ClusterId,
		ClusterName:       req.ClusterName,
		AgentID:           agentID,
		AgentSecret:       agentSecret,
		Status:            oakv1.StatusResponse_STATUS_APPROVED,
		ClientCertPEM:     clientCertPEM,
		ClientKeyPEM:      clientKeyPEM,
		AgentVersion:      req.AgentVersion,
		KubernetesVersion: req.KubernetesVersion,
	}

	s.logger.Infof("Agent approved: cluster=%s, agent_id=%s", req.ClusterId, agentID)

	return &oakv1.CredentialsResponse{
		Result: &oakv1.CredentialsResponse_Approved{
			Approved: &oakv1.ApprovedCredentials{
				AgentId:       agentID,
				AgentSecret:   agentSecret,
				ClientCertPem: clientCertPEM,
				ClientKeyPem:  clientKeyPEM,
				CaCertPem:     s.certManager.GetCACert(),
			},
		},
	}, nil
}

// ManualApprove allows manual approval from UI/API (to be called by admin handlers)
func (s *AgentManagementService) ManualApprove(clusterID string) error {
	agent, exists := s.agents[clusterID]
	if !exists {
		return fmt.Errorf("agent not found: %s", clusterID)
	}

	if agent.Status != oakv1.StatusResponse_STATUS_PENDING {
		return fmt.Errorf("agent is not in pending state: %s", clusterID)
	}

	// Create credentials request from stored state
	req := &oakv1.CredentialsRequest{
		ClusterId:         agent.ClusterID,
		ClusterName:       agent.ClusterName,
		AgentVersion:      agent.AgentVersion,
		KubernetesVersion: agent.KubernetesVersion,
	}

	// Approve the agent
	_, err := s.approveAgent(req)
	if err != nil {
		return fmt.Errorf("failed to approve agent: %w", err)
	}

	s.logger.Infof("Agent manually approved by admin: %s", clusterID)
	return nil
}

// ManualReject allows manual rejection from UI/API
func (s *AgentManagementService) ManualReject(clusterID string) error {
	agent, exists := s.agents[clusterID]
	if !exists {
		return fmt.Errorf("agent not found: %s", clusterID)
	}

	agent.Status = oakv1.StatusResponse_STATUS_REJECTED
	s.logger.Infof("Agent manually rejected by admin: %s", clusterID)
	return nil
}

// Revoke revokes an agent's credentials
func (s *AgentManagementService) Revoke(clusterID string) error {
	agent, exists := s.agents[clusterID]
	if !exists {
		return fmt.Errorf("agent not found: %s", clusterID)
	}

	agent.Status = oakv1.StatusResponse_STATUS_REVOKED
	s.logger.Infof("Agent credentials revoked: %s", clusterID)
	return nil
}

// ListPending returns all agents in pending state (for admin UI)
func (s *AgentManagementService) ListPending() []*AgentState {
	pending := []*AgentState{}
	for _, agent := range s.agents {
		if agent.Status == oakv1.StatusResponse_STATUS_PENDING {
			pending = append(pending, agent)
		}
	}
	return pending
}

// GetAgent returns agent state (for admin UI)
func (s *AgentManagementService) GetAgent(clusterID string) (*AgentState, error) {
	agent, exists := s.agents[clusterID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", clusterID)
	}
	return agent, nil
}