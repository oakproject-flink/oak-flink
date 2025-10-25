// +build integration

package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	oakv1 "github.com/oakproject-flink/oak-flink/api/proto/oak/v1"
	oakgrpc "github.com/oakproject-flink/oak-flink/oak-lib/grpc"
	"github.com/oakproject-flink/oak-flink/oak-lib/certs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func setupTestServer(t *testing.T) (*Server, *grpc.ClientConn, func()) {
	// Create cert manager
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	// Create server
	config := ServerConfig{
		Port:             "0", // Random port
		CertManager:      certManager,
		HeartbeatTimeout: 5 * time.Second,
	}

	server, err := NewServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create in-memory listener
	lis := bufconn.Listen(bufSize)

	// Start server in background
	go func() {
		if err := server.grpcServer.Serve(lis); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Create client connection with bufconn dialer
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get server cert for client
	clientCertPEM, clientKeyPEM, err := certManager.GenerateClientCert("test-agent")
	if err != nil {
		t.Fatalf("Failed to generate client cert: %v", err)
	}

	caCertPEM := certManager.GetCACert()

	// Create client credentials
	clientCreds, err := oakgrpc.NewClientCredentials(clientCertPEM, clientKeyPEM, caCertPEM, "localhost")
	if err != nil {
		t.Fatalf("Failed to create client credentials: %v", err)
	}

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(clientCreds),
	)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}

	cleanup := func() {
		conn.Close()
		server.Stop()
		lis.Close()
	}

	return server, conn, cleanup
}

func TestIntegration_HealthCheck(t *testing.T) {
	_, conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := oakv1.NewOakServiceClient(conn)
	ctx := context.Background()

	resp, err := client.HealthCheck(ctx, &oakv1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	if resp.Status != oakv1.HealthCheckResponse_SERVING {
		t.Errorf("Status = %v, want SERVING", resp.Status)
	}
}

func TestIntegration_AgentCredentialRequest(t *testing.T) {
	server, conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := oakv1.NewAgentManagementClient(conn)
	ctx := context.Background()

	// Test auto-approval with API token
	t.Run("auto_approve_with_token", func(t *testing.T) {
		req := &oakv1.CredentialsRequest{
			ClusterId:         "integration-cluster-1",
			ClusterName:       "Integration Test Cluster 1",
			ApiToken:          "test-token",
			AgentVersion:      "1.0.0",
			KubernetesVersion: "1.28.0",
		}

		resp, err := client.RequestCredentials(ctx, req)
		if err != nil {
			t.Fatalf("RequestCredentials failed: %v", err)
		}

		approved := resp.GetApproved()
		if approved == nil {
			t.Fatal("Expected approved response")
		}

		if approved.AgentId == "" {
			t.Error("AgentId should not be empty")
		}
		if approved.AgentSecret == "" {
			t.Error("AgentSecret should not be empty")
		}
		if len(approved.ClientCertPem) == 0 {
			t.Error("ClientCertPem should not be empty")
		}
	})

	// Test pending without token
	t.Run("pending_without_token", func(t *testing.T) {
		req := &oakv1.CredentialsRequest{
			ClusterId:   "integration-cluster-2",
			ClusterName: "Integration Test Cluster 2",
		}

		resp, err := client.RequestCredentials(ctx, req)
		if err != nil {
			t.Fatalf("RequestCredentials failed: %v", err)
		}

		pending := resp.GetPending()
		if pending == nil {
			t.Fatal("Expected pending response")
		}

		if pending.Message == "" {
			t.Error("Pending message should not be empty")
		}
	})

	// Test manual approval
	t.Run("manual_approval", func(t *testing.T) {
		clusterID := "integration-cluster-3"

		// Request credentials
		req := &oakv1.CredentialsRequest{
			ClusterId:   clusterID,
			ClusterName: "Integration Test Cluster 3",
		}

		resp, err := client.RequestCredentials(ctx, req)
		if err != nil {
			t.Fatalf("RequestCredentials failed: %v", err)
		}

		if resp.GetPending() == nil {
			t.Fatal("Expected pending response")
		}

		// Manually approve via server
		agentMgmt := server.GetAgentManagementService()
		if err := agentMgmt.ManualApprove(clusterID); err != nil {
			t.Fatalf("ManualApprove failed: %v", err)
		}

		// Check status
		statusResp, err := client.CheckStatus(ctx, &oakv1.StatusRequest{
			ClusterId: clusterID,
		})
		if err != nil {
			t.Fatalf("CheckStatus failed: %v", err)
		}

		if statusResp.Status != oakv1.StatusResponse_STATUS_APPROVED {
			t.Errorf("Status = %v, want APPROVED", statusResp.Status)
		}

		if statusResp.Credentials == nil {
			t.Error("Credentials should be present after approval")
		}
	})
}

func TestIntegration_AgentStatusCheck(t *testing.T) {
	_, conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := oakv1.NewOakServiceClient(conn)
	ctx := context.Background()

	// Check status of non-existent agent
	resp, err := client.GetAgentStatus(ctx, &oakv1.AgentStatusRequest{
		ClusterId: "non-existent",
	})
	if err != nil {
		t.Fatalf("GetAgentStatus failed: %v", err)
	}

	if resp.Status != oakv1.AgentStatusResponse_DISCONNECTED {
		t.Errorf("Status = %v, want DISCONNECTED for non-existent agent", resp.Status)
	}
}

func TestIntegration_FullAgentLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full lifecycle test in short mode")
	}

	server, conn, cleanup := setupTestServer(t)
	defer cleanup()

	credClient := oakv1.NewAgentManagementClient(conn)
	ctx := context.Background()

	clusterID := "lifecycle-cluster"

	// 1. Request credentials
	credReq := &oakv1.CredentialsRequest{
		ClusterId:   clusterID,
		ClusterName: "Lifecycle Test",
		ApiToken:    "token",
	}

	credResp, err := credClient.RequestCredentials(ctx, credReq)
	if err != nil {
		t.Fatalf("RequestCredentials failed: %v", err)
	}

	approved := credResp.GetApproved()
	if approved == nil {
		t.Fatal("Expected approved response")
	}

	// 2. Verify credentials are valid (check cert format)
	if len(approved.ClientCertPem) == 0 {
		t.Fatal("Client cert should not be empty")
	}
	if len(approved.ClientKeyPem) == 0 {
		t.Fatal("Client key should not be empty")
	}
	if len(approved.CaCertPem) == 0 {
		t.Fatal("CA cert should not be empty")
	}

	// 3. Check status shows approved
	statusResp, err := credClient.CheckStatus(ctx, &oakv1.StatusRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		t.Fatalf("CheckStatus failed: %v", err)
	}

	if statusResp.Status != oakv1.StatusResponse_STATUS_APPROVED {
		t.Errorf("Status = %v, want APPROVED", statusResp.Status)
	}

	// 4. Revoke credentials
	agentMgmt := server.GetAgentManagementService()
	if err := agentMgmt.Revoke(clusterID); err != nil {
		t.Fatalf("Revoke failed: %v", err)
	}

	// 5. Check status shows revoked
	statusResp2, err := credClient.CheckStatus(ctx, &oakv1.StatusRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		t.Fatalf("CheckStatus failed: %v", err)
	}

	if statusResp2.Status != oakv1.StatusResponse_STATUS_REVOKED {
		t.Errorf("Status = %v, want REVOKED", statusResp2.Status)
	}
}

func TestIntegration_ConcurrentAgentRegistrations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	_, conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := oakv1.NewAgentManagementClient(conn)
	ctx := context.Background()

	const numAgents = 10
	errChan := make(chan error, numAgents)

	// Register multiple agents concurrently
	for i := 0; i < numAgents; i++ {
		go func(id int) {
			req := &oakv1.CredentialsRequest{
				ClusterId:   string(rune('a' + id)),
				ClusterName: string(rune('A' + id)),
				ApiToken:    "token",
			}

			resp, err := client.RequestCredentials(ctx, req)
			if err != nil {
				errChan <- err
				return
			}

			if resp.GetApproved() == nil {
				errChan <- err
				return
			}

			errChan <- nil
		}(i)
	}

	// Check all succeeded
	for i := 0; i < numAgents; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent registration failed: %v", err)
		}
	}
}

// TestIntegration_MetricsReporting tests metrics reporting functionality
// TODO: Implement metrics storage in registry before enabling this test
/*
func TestIntegration_MetricsReporting(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	// Register agent via registry
	clusterID := "metrics-test-cluster"
	service := server.GetService()

	agentID := "test-agent-001"
	info := &AgentInfo{
		AgentID:     agentID,
		ClusterID:   clusterID,
		ClusterName: "Metrics Test",
	}
	service.registry.Register(agentID, info)

	// Create metrics report
	metricsReport := &oakv1.MetricsReport{
		Jobs: []*oakv1.JobMetrics{
			{
				JobId:               "job-001",
				JobName:             "Test Job",
				State:               oakv1.JobState_JOB_STATE_RUNNING,
				Parallelism:         4,
				RecordsInPerSecond:  1000,
				RecordsOutPerSecond: 900,
				BackpressureLevel:   0.15,
			},
		},
	}

	// TODO: Implement UpdateMetrics method in registry
	// service.registry.UpdateMetrics(agentID, metricsReport)

	// Verify metrics stored
	agent, ok := service.registry.Get(agentID)
	if !ok {
		t.Fatal("Agent not found")
	}

	// TODO: Add LastMetrics field to AgentInfo and implement storage
	_ = agent
	_ = metricsReport
}
*/