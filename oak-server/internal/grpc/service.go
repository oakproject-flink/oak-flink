package grpc

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	oakv1 "github.com/oakproject-flink/oak-flink/api/proto/oak/v1"
	"github.com/oakproject-flink/oak-flink/oak-lib/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service implements the OakService gRPC server
type Service struct {
	oakv1.UnimplementedOakServiceServer

	registry *Registry
	logger   *logger.Logger

	// Cleanup goroutines
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// NewService creates a new gRPC service
func NewService() *Service {
	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		registry: NewRegistry(),
		logger:   logger.NewComponent("agents"),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// AgentStream handles bidirectional streaming with agents
func (s *Service) AgentStream(stream oakv1.OakService_AgentStreamServer) error {
	// Get peer information
	peerInfo, ok := peer.FromContext(stream.Context())
	if !ok {
		return status.Error(codes.Internal, "failed to get peer info")
	}

	s.logger.Infof("New agent connection from: %s", peerInfo.Addr.String())

	// Wait for registration message
	msg, err := stream.Recv()
	if err != nil {
		s.logger.Errorf("Failed to receive registration: %v", err)
		return status.Error(codes.InvalidArgument, "registration required")
	}

	// Validate registration message
	reg, ok := msg.Payload.(*oakv1.AgentMessage_Registration)
	if !ok {
		return status.Error(codes.InvalidArgument, "first message must be registration")
	}

	registration := reg.Registration

	// Generate server-side agent ID
	agentID := generateMessageID()

	// Register agent
	agentInfo := &AgentInfo{
		AgentID:      agentID,
		ClusterID:    registration.ClusterId,
		ClusterName:  registration.ClusterName,
		AgentVersion: registration.AgentVersion,
		K8sVersion:   registration.KubernetesVersion,
		Capabilities: registration.Capabilities,
		Labels:       registration.Labels,
	}

	s.registry.Register(agentID, agentInfo)
	defer s.registry.Unregister(agentID)

	s.logger.Infof("Agent registered: id=%s, cluster=%s (%s), version=%s",
		agentID, registration.ClusterName, registration.ClusterId, registration.AgentVersion)

	// Send registration acknowledgment
	ack := &oakv1.ServerMessage{
		MessageId: generateMessageID(),
		Timestamp: timestampNow(),
		Payload: &oakv1.ServerMessage_RegistrationAck{
			RegistrationAck: &oakv1.RegistrationAck{
				AgentId:        agentID,
				WelcomeMessage: fmt.Sprintf("Welcome %s!", registration.ClusterName),
				ServerTime:     timestampNow(),
				Config: &oakv1.AgentConfig{
					HeartbeatIntervalSeconds: 30,
					MetricsIntervalSeconds:   60,
				},
			},
		},
	}

	if err := stream.Send(ack); err != nil {
		s.logger.Errorf("Failed to send registration ack: %v", err)
		return err
	}

	// Start bidirectional communication
	errChan := make(chan error, 2)

	// Receive goroutine (agent -> server)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		errChan <- s.receiveMessages(stream, agentID)
	}()

	// Send goroutine (server -> agent)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		errChan <- s.sendMessages(stream, agentInfo)
	}()

	// Wait for either goroutine to finish
	err = <-errChan

	if err != nil && err != io.EOF {
		s.logger.Errorf("Agent %s stream error: %v", agentID, err)
		return err
	}

	s.logger.Infof("Agent %s disconnected", agentID)
	return nil
}

// receiveMessages handles incoming messages from agent
func (s *Service) receiveMessages(stream oakv1.OakService_AgentStreamServer, agentID string) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return io.EOF
		}
		if err != nil {
			return fmt.Errorf("receive error: %w", err)
		}

		// Handle message based on type
		switch payload := msg.Payload.(type) {
		case *oakv1.AgentMessage_Heartbeat:
			s.handleHeartbeat(agentID, payload.Heartbeat)

		case *oakv1.AgentMessage_Metrics:
			s.handleMetrics(agentID, payload.Metrics)

		case *oakv1.AgentMessage_Event:
			s.handleEvent(agentID, payload.Event)

		case *oakv1.AgentMessage_CommandResult:
			s.handleCommandResult(agentID, payload.CommandResult)

		default:
			s.logger.Warnf("Unknown message type from agent %s", agentID)
		}
	}
}

// sendMessages handles outgoing messages to agent
func (s *Service) sendMessages(stream oakv1.OakService_AgentStreamServer, agentInfo *AgentInfo) error {
	for {
		select {
		case msg, ok := <-agentInfo.SendChan:
			if !ok {
				// Channel closed, agent unregistered
				return io.EOF
			}

			if err := stream.Send(msg); err != nil {
				return fmt.Errorf("send error: %w", err)
			}

		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

// handleHeartbeat processes heartbeat messages
func (s *Service) handleHeartbeat(agentID string, heartbeat *oakv1.Heartbeat) {
	s.registry.UpdateHeartbeat(agentID, heartbeat)
	s.logger.Debugf("Heartbeat from agent %s: status=%s, jobs=%d",
		agentID, heartbeat.Status, heartbeat.ActiveJobs)
}

// handleMetrics processes metrics reports
func (s *Service) handleMetrics(agentID string, metrics *oakv1.MetricsReport) {
	s.logger.Infof("Received metrics from agent %s: %d jobs",
		agentID, len(metrics.Jobs))

	// TODO: Store metrics in database or time-series DB
	// For now, just log
	for _, jobMetric := range metrics.Jobs {
		s.logger.Debugf("Job %s: state=%s, parallelism=%d",
			jobMetric.JobId, jobMetric.State, jobMetric.Parallelism)
	}
}

// handleEvent processes event reports
func (s *Service) handleEvent(agentID string, event *oakv1.EventReport) {
	s.logger.Infof("Event from agent %s: type=%s, severity=%s, message=%s",
		agentID, event.Type, event.Severity, event.Message)

	// TODO: Store events in database
	// TODO: Trigger alerts based on severity
}

// handleCommandResult processes command execution results
func (s *Service) handleCommandResult(agentID string, result *oakv1.CommandResult) {
	s.logger.Infof("Command result from agent %s: cmd=%s, success=%v",
		agentID, result.CommandId, result.Success)

	if !result.Success {
		s.logger.Errorf("Command %s failed: %s", result.CommandId, result.Message)
	}

	// TODO: Update command status in database
	// TODO: Notify waiting API requests
}

// HealthCheck implements the health check RPC
func (s *Service) HealthCheck(ctx context.Context, req *oakv1.HealthCheckRequest) (*oakv1.HealthCheckResponse, error) {
	return &oakv1.HealthCheckResponse{
		Status: oakv1.HealthCheckResponse_SERVING,
	}, nil
}

// GetAgentStatus returns the connection status of an agent
func (s *Service) GetAgentStatus(ctx context.Context, req *oakv1.AgentStatusRequest) (*oakv1.AgentStatusResponse, error) {
	agent, ok := s.registry.Get(req.ClusterId)
	if !ok {
		// Agent not found
		return &oakv1.AgentStatusResponse{
			Status: oakv1.AgentStatusResponse_DISCONNECTED,
		}, nil
	}

	// Agent is connected
	return &oakv1.AgentStatusResponse{
		Status:       oakv1.AgentStatusResponse_CONNECTED,
		AgentId:      agent.AgentID,
		LastSeen:     timestamppb.New(agent.LastHeartbeat),
		HealthStatus: agent.Status,
	}, nil
}

// GetRegistry returns the agent registry (for API handlers)
func (s *Service) GetRegistry() *Registry {
	return s.registry
}

// StartHealthChecker starts a background goroutine to check agent health
// Check interval is set to 30 seconds by default
func (s *Service) StartHealthChecker(timeout time.Duration) {
	s.StartHealthCheckerWithInterval(timeout, 30*time.Second)
}

// StartHealthCheckerWithInterval starts a background goroutine to check agent health
// with a custom check interval (useful for testing)
func (s *Service) StartHealthCheckerWithInterval(timeout time.Duration, checkInterval time.Duration) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.registry.CheckHealth(timeout)

			case <-s.ctx.Done():
				return
			}
		}
	}()
}

// Shutdown gracefully shuts down the service
func (s *Service) Shutdown() {
	s.logger.Infof("Shutting down gRPC service...")
	s.cancel()
	s.wg.Wait()
	s.logger.Infof("gRPC service shutdown complete")
}
