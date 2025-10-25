package grpc

import (
	"context"
	"testing"
	"time"

	oakv1 "github.com/oakproject-flink/oak-flink/api/proto/oak/v1"
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

func TestNewService(t *testing.T) {
	service := NewService()
	defer service.Shutdown()

	if service == nil {
		t.Fatal("NewService() returned nil")
	}

	if service.registry == nil {
		t.Error("Service registry should not be nil")
	}

	if service.logger == nil {
		t.Error("Service logger should not be nil")
	}
}

func TestHealthCheck(t *testing.T) {
	service := NewService()
	defer service.Shutdown()

	ctx := context.Background()

	resp, err := service.HealthCheck(ctx, &oakv1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("HealthCheck() error = %v", err)
	}

	if resp.Status != oakv1.HealthCheckResponse_SERVING {
		t.Errorf("Status = %v, want SERVING", resp.Status)
	}
}

func TestGetRegistry(t *testing.T) {
	service := NewService()
	defer service.Shutdown()

	registry := service.GetRegistry()
	if registry == nil {
		t.Error("GetRegistry() should not return nil")
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := NewRegistry()

	agentID := "test-agent-001"
	info := &AgentInfo{
		AgentID:      agentID,
		ClusterID:    "cluster-001",
		ClusterName:  "Test Cluster",
		AgentVersion: "1.0.0",
		K8sVersion:   "1.28.0",
		ConnectedAt:  time.Now(),
	}

	// Register agent
	registry.Register(agentID, info)

	// Get agent
	retrieved, ok := registry.Get(agentID)
	if !ok {
		t.Fatal("Expected to retrieve registered agent")
	}

	if retrieved.AgentID != agentID {
		t.Errorf("AgentID = %s, want %s", retrieved.AgentID, agentID)
	}

	if retrieved.ClusterID != "cluster-001" {
		t.Errorf("ClusterID = %s, want cluster-001", retrieved.ClusterID)
	}

	if retrieved.ClusterName != "Test Cluster" {
		t.Errorf("ClusterName = %s, want Test Cluster", retrieved.ClusterName)
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	// Register multiple agents
	agents := []struct {
		agentID   string
		clusterID string
	}{
		{"agent-001", "cluster-001"},
		{"agent-002", "cluster-002"},
		{"agent-003", "cluster-001"}, // Same cluster as agent-001
	}

	for _, a := range agents {
		info := &AgentInfo{
			AgentID:     a.agentID,
			ClusterID:   a.clusterID,
			ClusterName: "Test Cluster",
			ConnectedAt: time.Now(),
		}
		registry.Register(a.agentID, info)
	}

	// List all agents
	all := registry.List()
	if len(all) != 3 {
		t.Errorf("Expected 3 agents, got %d", len(all))
	}
}

func TestRegistry_Count(t *testing.T) {
	registry := NewRegistry()

	if registry.Count() != 0 {
		t.Errorf("Initial count = %d, want 0", registry.Count())
	}

	// Register agents
	for i := 0; i < 5; i++ {
		info := &AgentInfo{
			AgentID:   string(rune('a' + i)),
			ClusterID: string(rune('A' + i)),
		}
		registry.Register(info.AgentID, info)
	}

	if registry.Count() != 5 {
		t.Errorf("Count = %d, want 5", registry.Count())
	}
}

func TestRegistry_GetByCluster(t *testing.T) {
	registry := NewRegistry()

	// Register agents in different clusters
	clusterAgents := map[string][]string{
		"cluster-A": {"agent-1", "agent-2", "agent-3"},
		"cluster-B": {"agent-4", "agent-5"},
	}

	for clusterID, agentIDs := range clusterAgents {
		for _, agentID := range agentIDs {
			info := &AgentInfo{
				AgentID:   agentID,
				ClusterID: clusterID,
			}
			registry.Register(agentID, info)
		}
	}

	// Get agents by cluster
	clusterA := registry.GetByCluster("cluster-A")
	if len(clusterA) != 3 {
		t.Errorf("cluster-A has %d agents, want 3", len(clusterA))
	}

	clusterB := registry.GetByCluster("cluster-B")
	if len(clusterB) != 2 {
		t.Errorf("cluster-B has %d agents, want 2", len(clusterB))
	}

	// Non-existent cluster
	clusterC := registry.GetByCluster("cluster-C")
	if len(clusterC) != 0 {
		t.Errorf("cluster-C has %d agents, want 0", len(clusterC))
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()

	agentID := "test-agent"
	info := &AgentInfo{
		AgentID:   agentID,
		ClusterID: "cluster-001",
	}

	registry.Register(agentID, info)

	// Verify registered
	if registry.Count() != 1 {
		t.Errorf("Count = %d, want 1", registry.Count())
	}

	// Unregister
	registry.Unregister(agentID)

	// Verify unregistered
	if registry.Count() != 0 {
		t.Errorf("Count after unregister = %d, want 0", registry.Count())
	}

	_, ok := registry.Get(agentID)
	if ok {
		t.Error("Agent should not exist after unregister")
	}
}

func TestRegistry_UpdateHeartbeat(t *testing.T) {
	registry := NewRegistry()

	agentID := "test-agent"
	info := &AgentInfo{
		AgentID:   agentID,
		ClusterID: "cluster-001",
		SendChan:  make(chan *oakv1.ServerMessage, 10),
	}

	registry.Register(agentID, info)

	// Get initial heartbeat time
	initial, _ := registry.Get(agentID)
	initialTime := initial.LastHeartbeat

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Update heartbeat
	heartbeat := &oakv1.Heartbeat{
		ActiveJobs: 3,
		Status:     oakv1.AgentStatus_AGENT_STATUS_HEALTHY,
	}
	registry.UpdateHeartbeat(agentID, heartbeat)

	// Verify heartbeat time updated
	updated, _ := registry.Get(agentID)
	if !updated.LastHeartbeat.After(initialTime) {
		t.Error("LastHeartbeat should have been updated")
	}
}

func TestHealthChecker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping health checker test in short mode")
	}

	service := NewService()
	defer service.Shutdown()

	registry := service.GetRegistry()

	// Register an agent
	agentID := "test-agent"
	info := &AgentInfo{
		AgentID:       agentID,
		ClusterID:     "cluster-test",
		ClusterName:   "Test",
		ConnectedAt:   time.Now(),
		LastHeartbeat: time.Now(),
		Status:        oakv1.AgentStatus_AGENT_STATUS_HEALTHY,
		SendChan:      make(chan *oakv1.ServerMessage, 10),
	}
	registry.Register(agentID, info)

	// Start health checker with short timeout and fast check interval
	service.StartHealthCheckerWithInterval(100*time.Millisecond, 50*time.Millisecond)

	// Wait for health checker to run and detect timeout
	time.Sleep(200 * time.Millisecond)

	// Check agent status
	retrieved, ok := registry.Get(agentID)
	if !ok {
		t.Fatal("Agent should still exist")
	}

	if retrieved.Status != oakv1.AgentStatus_AGENT_STATUS_UNHEALTHY {
		t.Errorf("Agent status = %v, want AGENT_STATUS_UNHEALTHY after timeout", retrieved.Status)
	}
}

func TestConcurrentRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	registry := NewRegistry()

	const numAgents = 100
	done := make(chan bool, numAgents)

	// Register agents concurrently
	for i := 0; i < numAgents; i++ {
		go func(id int) {
			agentID := string(rune('a' + (id % 26))) + string(rune('0' + (id / 26)))
			info := &AgentInfo{
				AgentID:   agentID,
				ClusterID: "cluster",
			}
			registry.Register(agentID, info)
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < numAgents; i++ {
		<-done
	}

	// Verify count
	if registry.Count() != numAgents {
		t.Errorf("Count = %d, want %d", registry.Count(), numAgents)
	}
}

func TestSendCommand(t *testing.T) {
	registry := NewRegistry()

	agentID := "test-agent"
	sendChan := make(chan *oakv1.ServerMessage, 10)
	info := &AgentInfo{
		AgentID:   agentID,
		ClusterID: "cluster-001",
		SendChan:  sendChan,
	}

	registry.Register(agentID, info)

	// Send command
	cmd := &oakv1.Command{
		CommandId: "cmd-001",
		Command: &oakv1.Command_ScaleJob{
			ScaleJob: &oakv1.ScaleJobCommand{
				JobId:          "job-001",
				NewParallelism: 4,
			},
		},
	}

	err := registry.SendCommand(agentID, cmd)
	if err != nil {
		t.Fatalf("SendCommand() error = %v", err)
	}

	// Verify command was sent to channel
	select {
	case msg := <-sendChan:
		cmdPayload := msg.GetCommand()
		if cmdPayload == nil {
			t.Fatal("Expected command in message")
		}
		if cmdPayload.CommandId != "cmd-001" {
			t.Errorf("CommandId = %s, want cmd-001", cmdPayload.CommandId)
		}
		scaleCmd := cmdPayload.GetScaleJob()
		if scaleCmd == nil {
			t.Fatal("Expected scale job command")
		}
		if scaleCmd.JobId != "job-001" {
			t.Errorf("JobId = %s, want job-001", scaleCmd.JobId)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Command not received in channel")
	}
}

func TestSendConfigUpdate(t *testing.T) {
	registry := NewRegistry()

	agentID := "test-agent"
	sendChan := make(chan *oakv1.ServerMessage, 10)
	info := &AgentInfo{
		AgentID:   agentID,
		ClusterID: "cluster-001",
		SendChan:  sendChan,
	}

	registry.Register(agentID, info)

	// Send config update
	config := &oakv1.AgentConfig{
		HeartbeatIntervalSeconds: 60,
		MetricsIntervalSeconds:   120,
	}

	err := registry.SendConfigUpdate(agentID, config)
	if err != nil {
		t.Fatalf("SendConfigUpdate() error = %v", err)
	}

	// Verify config was sent to channel
	select {
	case msg := <-sendChan:
		cfgUpdatePayload := msg.GetConfigUpdate()
		if cfgUpdatePayload == nil {
			t.Fatal("Expected config update in message")
		}
		cfg := cfgUpdatePayload.Config
		if cfg == nil {
			t.Fatal("Expected config in config update")
		}
		if cfg.HeartbeatIntervalSeconds != 60 {
			t.Errorf("HeartbeatInterval = %d, want 60", cfg.HeartbeatIntervalSeconds)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Config update not received in channel")
	}
}
// TestRegistry_SendToClosedChannel tests that sending to an unregistered agent returns proper error
func TestRegistry_SendToClosedChannel(t *testing.T) {
	registry := NewRegistry()

	agentID := "test-agent"
	info := &AgentInfo{
		AgentID:   agentID,
		ClusterID: "cluster-001",
	}

	registry.Register(agentID, info)

	// Unregister the agent
	registry.Unregister(agentID)

	// Try to send command to closed channel - should return error, not panic
	cmd := &oakv1.Command{
		CommandId: "cmd-001",
		Command: &oakv1.Command_ScaleJob{
			ScaleJob: &oakv1.ScaleJobCommand{
				JobId:          "job-001",
				NewParallelism: 4,
			},
		},
	}

	err := registry.SendCommand(agentID, cmd)
	if err != ErrAgentNotFound {
		t.Errorf("SendCommand() error = %v, want %v", err, ErrAgentNotFound)
	}

	// Try config update too
	config := &oakv1.AgentConfig{
		HeartbeatIntervalSeconds: 60,
	}

	err = registry.SendConfigUpdate(agentID, config)
	if err != ErrAgentNotFound {
		t.Errorf("SendConfigUpdate() error = %v, want %v", err, ErrAgentNotFound)
	}
}

// TestRegistry_ConcurrentSendAndUnregister tests concurrent sends and unregisters don't panic
func TestRegistry_ConcurrentSendAndUnregister(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	registry := NewRegistry()

	agentID := "test-agent"
	info := &AgentInfo{
		AgentID:   agentID,
		ClusterID: "cluster-001",
	}

	registry.Register(agentID, info)

	// Start goroutine that continuously tries to send
	done := make(chan struct{})
	errors := make(chan error, 100)

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				cmd := &oakv1.Command{
					CommandId: "cmd-001",
					Command: &oakv1.Command_ScaleJob{
						ScaleJob: &oakv1.ScaleJobCommand{
							JobId:          "job-001",
							NewParallelism: 4,
						},
					},
				}
				err := registry.SendCommand(agentID, cmd)
				if err != nil && err != ErrAgentNotFound && err != ErrAgentDisconnected && err != ErrSendChannelFull {
					errors <- err
				}
			}
		}
	}()

	// Let it send for a bit
	time.Sleep(10 * time.Millisecond)

	// Unregister the agent
	registry.Unregister(agentID)

	// Stop the sender
	close(done)

	// Check for unexpected errors
	select {
	case err := <-errors:
		t.Errorf("Unexpected error during concurrent send: %v", err)
	default:
		// No errors - good!
	}
}

// TestRegistry_MultipleUnregister tests that multiple unregisters don't panic
func TestRegistry_MultipleUnregister(t *testing.T) {
	registry := NewRegistry()

	agentID := "test-agent"
	info := &AgentInfo{
		AgentID:   agentID,
		ClusterID: "cluster-001",
	}

	registry.Register(agentID, info)

	// Unregister multiple times - should not panic
	registry.Unregister(agentID)
	registry.Unregister(agentID)
	registry.Unregister(agentID)

	// Agent should not exist
	_, exists := registry.Get(agentID)
	if exists {
		t.Error("Agent should not exist after unregister")
	}
}
