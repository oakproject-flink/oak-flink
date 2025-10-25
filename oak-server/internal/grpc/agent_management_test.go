package grpc

import (
	"context"
	"testing"

	oakv1 "github.com/oakproject-flink/oak-flink/api/proto/oak/v1"
	"github.com/oakproject-flink/oak-flink/oak-lib/certs"
	"github.com/oakproject-flink/oak-flink/oak-lib/logger"
)

func init() {
	// Configure logger to not create files during tests
	logger.SetGlobalConfig(&logger.Config{
		LogDir:   "logs",
		Format:   logger.FormatText,
		Debug:    false,
		Fields:   []string{"timestamp", "level", "component", "message"},
		ToStdout: false, // Quiet during tests
		ToFile:   false, // Don't create files
		BufSize:  1000,
	})
}

func TestRequestCredentials_WithAPIToken(t *testing.T) {
	// Setup
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	service := NewAgentManagementService(certManager)
	ctx := context.Background()

	tests := []struct {
		name        string
		request     *oakv1.CredentialsRequest
		wantStatus  oakv1.StatusResponse_Status
		wantApprove bool
		wantErr     bool
	}{
		{
			name: "new agent with api token - auto approve",
			request: &oakv1.CredentialsRequest{
				ClusterId:         "cluster-001",
				ClusterName:       "Production Cluster",
				ApiToken:          "valid-token",
				AgentVersion:      "1.0.0",
				KubernetesVersion: "1.28.0",
			},
			wantStatus:  oakv1.StatusResponse_STATUS_APPROVED,
			wantApprove: true,
			wantErr:     false,
		},
		{
			name: "new agent without api token - pending",
			request: &oakv1.CredentialsRequest{
				ClusterId:         "cluster-002",
				ClusterName:       "Dev Cluster",
				AgentVersion:      "1.0.0",
				KubernetesVersion: "1.28.0",
			},
			wantStatus:  oakv1.StatusResponse_STATUS_PENDING,
			wantApprove: false,
			wantErr:     false,
		},
		{
			name: "missing cluster_id",
			request: &oakv1.CredentialsRequest{
				ClusterName: "Test Cluster",
			},
			wantErr: true,
		},
		{
			name: "missing cluster_name",
			request: &oakv1.CredentialsRequest{
				ClusterId: "cluster-003",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.RequestCredentials(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("RequestCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check response type
			if tt.wantApprove {
				approved := resp.GetApproved()
				if approved == nil {
					t.Errorf("Expected approved response, got: %v", resp)
					return
				}

				// Verify credentials are populated
				if approved.AgentId == "" {
					t.Error("AgentId should not be empty")
				}
				if approved.AgentSecret == "" {
					t.Error("AgentSecret should not be empty")
				}
				if len(approved.ClientCertPem) == 0 {
					t.Error("ClientCertPem should not be empty")
				}
				if len(approved.ClientKeyPem) == 0 {
					t.Error("ClientKeyPem should not be empty")
				}
				if len(approved.CaCertPem) == 0 {
					t.Error("CaCertPem should not be empty")
				}
			} else {
				pending := resp.GetPending()
				if pending == nil {
					t.Errorf("Expected pending response, got: %v", resp)
					return
				}

				if pending.Message == "" {
					t.Error("Pending message should not be empty")
				}
				if pending.PollIntervalSeconds == 0 {
					t.Error("Poll interval should be set")
				}
			}
		})
	}
}

func TestRequestCredentials_ExistingAgent(t *testing.T) {
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	service := NewAgentManagementService(certManager)
	ctx := context.Background()

	// Create an approved agent
	req := &oakv1.CredentialsRequest{
		ClusterId:   "cluster-001",
		ClusterName: "Test Cluster",
		ApiToken:    "valid-token",
	}

	resp1, err := service.RequestCredentials(ctx, req)
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}

	approved1 := resp1.GetApproved()
	if approved1 == nil {
		t.Fatal("Expected approved response")
	}

	// Request again with same cluster_id - should return same credentials
	resp2, err := service.RequestCredentials(ctx, req)
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}

	approved2 := resp2.GetApproved()
	if approved2 == nil {
		t.Fatal("Expected approved response")
	}

	// Verify same credentials returned
	if approved1.AgentId != approved2.AgentId {
		t.Errorf("AgentId changed: %s != %s", approved1.AgentId, approved2.AgentId)
	}
	if approved1.AgentSecret != approved2.AgentSecret {
		t.Errorf("AgentSecret changed")
	}
}

func TestCheckStatus(t *testing.T) {
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	service := NewAgentManagementService(certManager)
	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func() string // Returns cluster_id
		wantStatus oakv1.StatusResponse_Status
		wantCreds  bool
	}{
		{
			name: "unknown cluster",
			setup: func() string {
				return "unknown-cluster"
			},
			wantStatus: oakv1.StatusResponse_STATUS_UNKNOWN,
			wantCreds:  false,
		},
		{
			name: "approved agent",
			setup: func() string {
				req := &oakv1.CredentialsRequest{
					ClusterId:   "cluster-approved",
					ClusterName: "Approved Cluster",
					ApiToken:    "token",
				}
				service.RequestCredentials(ctx, req)
				return "cluster-approved"
			},
			wantStatus: oakv1.StatusResponse_STATUS_APPROVED,
			wantCreds:  true,
		},
		{
			name: "pending agent",
			setup: func() string {
				req := &oakv1.CredentialsRequest{
					ClusterId:   "cluster-pending",
					ClusterName: "Pending Cluster",
				}
				service.RequestCredentials(ctx, req)
				return "cluster-pending"
			},
			wantStatus: oakv1.StatusResponse_STATUS_PENDING,
			wantCreds:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterID := tt.setup()

			resp, err := service.CheckStatus(ctx, &oakv1.StatusRequest{
				ClusterId: clusterID,
			})

			if err != nil {
				t.Fatalf("CheckStatus() error = %v", err)
			}

			if resp.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", resp.Status, tt.wantStatus)
			}

			if tt.wantCreds {
				if resp.Credentials == nil {
					t.Error("Expected credentials, got nil")
				}
			} else {
				if resp.Credentials != nil {
					t.Error("Expected no credentials, got some")
				}
			}

			if resp.Message == "" {
				t.Error("Message should not be empty")
			}
		})
	}
}

func TestManualApprove(t *testing.T) {
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	service := NewAgentManagementService(certManager)
	ctx := context.Background()

	// Create pending agent
	req := &oakv1.CredentialsRequest{
		ClusterId:   "cluster-manual",
		ClusterName: "Manual Cluster",
	}
	service.RequestCredentials(ctx, req)

	// Manually approve
	err = service.ManualApprove("cluster-manual")
	if err != nil {
		t.Fatalf("ManualApprove() error = %v", err)
	}

	// Check status
	resp, err := service.CheckStatus(ctx, &oakv1.StatusRequest{
		ClusterId: "cluster-manual",
	})
	if err != nil {
		t.Fatalf("CheckStatus() error = %v", err)
	}

	if resp.Status != oakv1.StatusResponse_STATUS_APPROVED {
		t.Errorf("Status = %v, want APPROVED", resp.Status)
	}

	if resp.Credentials == nil {
		t.Error("Expected credentials after approval")
	}
}

func TestManualReject(t *testing.T) {
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	service := NewAgentManagementService(certManager)
	ctx := context.Background()

	// Create pending agent
	req := &oakv1.CredentialsRequest{
		ClusterId:   "cluster-reject",
		ClusterName: "Reject Cluster",
	}
	service.RequestCredentials(ctx, req)

	// Manually reject
	err = service.ManualReject("cluster-reject")
	if err != nil {
		t.Fatalf("ManualReject() error = %v", err)
	}

	// Check status
	resp, err := service.CheckStatus(ctx, &oakv1.StatusRequest{
		ClusterId: "cluster-reject",
	})
	if err != nil {
		t.Fatalf("CheckStatus() error = %v", err)
	}

	if resp.Status != oakv1.StatusResponse_STATUS_REJECTED {
		t.Errorf("Status = %v, want REJECTED", resp.Status)
	}
}

func TestRevoke(t *testing.T) {
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	service := NewAgentManagementService(certManager)
	ctx := context.Background()

	// Create and approve agent
	req := &oakv1.CredentialsRequest{
		ClusterId:   "cluster-revoke",
		ClusterName: "Revoke Cluster",
		ApiToken:    "token",
	}
	service.RequestCredentials(ctx, req)

	// Revoke credentials
	err = service.Revoke("cluster-revoke")
	if err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}

	// Check status
	resp, err := service.CheckStatus(ctx, &oakv1.StatusRequest{
		ClusterId: "cluster-revoke",
	})
	if err != nil {
		t.Fatalf("CheckStatus() error = %v", err)
	}

	if resp.Status != oakv1.StatusResponse_STATUS_REVOKED {
		t.Errorf("Status = %v, want REVOKED", resp.Status)
	}

	// Try to request credentials again - should return revoked status
	resp2, err := service.RequestCredentials(ctx, req)
	if err != nil {
		t.Fatalf("RequestCredentials() error = %v", err)
	}

	rejected := resp2.GetRejected()
	if rejected == nil {
		t.Error("Expected rejected response for revoked agent")
	}
}

func TestListPending(t *testing.T) {
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	service := NewAgentManagementService(certManager)
	ctx := context.Background()

	// Create mix of agents
	requests := []*oakv1.CredentialsRequest{
		{ClusterId: "pending-1", ClusterName: "Pending 1"},
		{ClusterId: "pending-2", ClusterName: "Pending 2"},
		{ClusterId: "approved-1", ClusterName: "Approved 1", ApiToken: "token"},
	}

	for _, req := range requests {
		service.RequestCredentials(ctx, req)
	}

	// Get pending list
	pending := service.ListPending()

	if len(pending) != 2 {
		t.Errorf("Expected 2 pending agents, got %d", len(pending))
	}

	// Verify they're actually pending
	for _, agent := range pending {
		if agent.Status != oakv1.StatusResponse_STATUS_PENDING {
			t.Errorf("Agent %s has status %v, want PENDING", agent.ClusterID, agent.Status)
		}
	}
}
