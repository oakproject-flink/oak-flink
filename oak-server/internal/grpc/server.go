package grpc

import (
	"fmt"
	"net"
	"time"

	oakv1 "github.com/oakproject-flink/oak-flink/api/proto/oak/v1"
	oakgrpc "github.com/oakproject-flink/oak-flink/oak-lib/grpc"
	"github.com/oakproject-flink/oak-flink/oak-lib/certs"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// Server wraps the gRPC server
type Server struct {
	grpcServer       *grpclib.Server
	service          *Service
	agentMgmtService *AgentManagementService
	listener         net.Listener
	port             string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port             string
	CertManager      *certs.Manager
	HeartbeatTimeout time.Duration
}

// NewServer creates a new gRPC server with mTLS
func NewServer(config ServerConfig) (*Server, error) {
	// Get server certificates
	serverCertPEM, serverKeyPEM := config.CertManager.GetServerCertAndKey()
	caCertPEM := config.CertManager.GetCACert()

	// Create TLS credentials with mTLS (client cert optional for AgentManagement service)
	creds, err := oakgrpc.NewServerCredentialsWithOptionalClient(
		serverCertPEM,
		serverKeyPEM,
		caCertPEM,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials: %w", err)
	}

	// Create gRPC server with options
	grpcServer := grpclib.NewServer(
		grpclib.Creds(creds),
		grpclib.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     5 * time.Minute,
			MaxConnectionAge:      2 * time.Hour,
			MaxConnectionAgeGrace: 5 * time.Minute,
			Time:                  30 * time.Second,
			Timeout:               10 * time.Second,
		}),
		grpclib.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             15 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	// Create services
	service := NewService()
	agentMgmtService := NewAgentManagementService(config.CertManager)

	// Register services
	oakv1.RegisterOakServiceServer(grpcServer, service)
	oakv1.RegisterAgentManagementServer(grpcServer, agentMgmtService)

	// Start health checker
	if config.HeartbeatTimeout == 0 {
		config.HeartbeatTimeout = 90 * time.Second // 3x heartbeat interval
	}
	service.StartHealthChecker(config.HeartbeatTimeout)

	return &Server{
		grpcServer:       grpcServer,
		service:          service,
		agentMgmtService: agentMgmtService,
		port:             config.Port,
	}, nil
}

// Start starts the gRPC server
func (s *Server) Start() error {
	// Create listener
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	s.listener = lis

	// Start serving (blocking)
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// GetService returns the gRPC service (for accessing registry, etc.)
func (s *Server) GetService() *Service {
	return s.service
}

// GetAgentManagementService returns the agent management service (for admin API)
func (s *Server) GetAgentManagementService() *AgentManagementService {
	return s.agentMgmtService
}

// Stop gracefully stops the server
func (s *Server) Stop() {
	s.service.Shutdown()
	s.grpcServer.GracefulStop()
}
