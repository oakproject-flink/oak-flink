package grpc

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	oakv1 "github.com/oakproject-flink/oak-flink/api/proto/oak/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AgentInfo holds information about a connected agent
type AgentInfo struct {
	AgentID       string
	ClusterID     string
	ClusterName   string
	AgentVersion  string
	K8sVersion    string
	Capabilities  *oakv1.AgentCapabilities
	Labels        map[string]string
	ConnectedAt   time.Time
	LastHeartbeat time.Time
	Status        oakv1.AgentStatus
	ActiveJobs    int32

	// Communication channel
	SendChan chan *oakv1.ServerMessage

	// Channel state protection
	mu     sync.Mutex
	closed bool
}

// Registry manages connected agents
type Registry struct {
	mu     sync.RWMutex
	agents map[string]*AgentInfo // agentID -> AgentInfo
}

// NewRegistry creates a new agent registry
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]*AgentInfo),
	}
}

// Register adds a new agent to the registry
func (r *Registry) Register(agentID string, info *AgentInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	info.AgentID = agentID
	info.ConnectedAt = time.Now()
	info.LastHeartbeat = time.Now()

	// Only create SendChan if not provided (for tests)
	if info.SendChan == nil {
		info.SendChan = make(chan *oakv1.ServerMessage, 100) // Buffered channel
	}

	r.agents[agentID] = info
}

// Unregister removes an agent from the registry
func (r *Registry) Unregister(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if info, exists := r.agents[agentID]; exists {
		// Mark as closed and close channel safely
		info.mu.Lock()
		if !info.closed {
			info.closed = true
			close(info.SendChan)
		}
		info.mu.Unlock()

		delete(r.agents, agentID)
	}
}

// copyAgentInfo creates a safe copy of AgentInfo for external use
// Note: SendChan and mu fields are not copied (left as nil/zero)
func copyAgentInfo(info *AgentInfo) *AgentInfo {
	return &AgentInfo{
		AgentID:       info.AgentID,
		ClusterID:     info.ClusterID,
		ClusterName:   info.ClusterName,
		AgentVersion:  info.AgentVersion,
		K8sVersion:    info.K8sVersion,
		Capabilities:  info.Capabilities,
		Labels:        info.Labels,
		ConnectedAt:   info.ConnectedAt,
		LastHeartbeat: info.LastHeartbeat,
		Status:        info.Status,
		ActiveJobs:    info.ActiveJobs,
		// SendChan and mu are intentionally not copied
	}
}

// Get retrieves agent information
// Returns a copy of the AgentInfo to prevent data races.
func (r *Registry) Get(agentID string) (*AgentInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.agents[agentID]
	if !exists {
		return nil, false
	}

	return copyAgentInfo(info), true
}

// UpdateHeartbeat updates the last heartbeat time for an agent
func (r *Registry) UpdateHeartbeat(agentID string, heartbeat *oakv1.Heartbeat) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if info, exists := r.agents[agentID]; exists {
		info.LastHeartbeat = time.Now()
		if heartbeat != nil {
			info.ActiveJobs = heartbeat.ActiveJobs
			info.Status = heartbeat.Status
		}
	}
}

// List returns all registered agents
// Returns copies of AgentInfo to prevent data races.
func (r *Registry) List() []*AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*AgentInfo, 0, len(r.agents))
	for _, info := range r.agents {
		agents = append(agents, copyAgentInfo(info))
	}
	return agents
}

// Count returns the number of registered agents
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.agents)
}

// GetByCluster returns all agents for a specific cluster
// Returns copies of AgentInfo to prevent data races.
func (r *Registry) GetByCluster(clusterID string) []*AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*AgentInfo, 0)
	for _, info := range r.agents {
		if info.ClusterID == clusterID {
			agents = append(agents, copyAgentInfo(info))
		}
	}
	return agents
}

// sendMessage is a helper that safely sends a message to an agent's channel
func (r *Registry) sendMessage(agentID string, msg *oakv1.ServerMessage) error {
	r.mu.RLock()
	info, exists := r.agents[agentID]
	r.mu.RUnlock()

	if !exists {
		return ErrAgentNotFound
	}

	// Check if channel is closed before sending
	info.mu.Lock()
	defer info.mu.Unlock()

	if info.closed {
		return ErrAgentDisconnected
	}

	select {
	case info.SendChan <- msg:
		return nil
	default:
		return ErrSendChannelFull
	}
}

// SendCommand sends a command to a specific agent
func (r *Registry) SendCommand(agentID string, cmd *oakv1.Command) error {
	msg := &oakv1.ServerMessage{
		MessageId: generateMessageID(),
		Timestamp: timestampNow(),
		Payload: &oakv1.ServerMessage_Command{
			Command: cmd,
		},
	}

	return r.sendMessage(agentID, msg)
}

// SendConfigUpdate sends a config update to an agent
func (r *Registry) SendConfigUpdate(agentID string, config *oakv1.AgentConfig) error {
	msg := &oakv1.ServerMessage{
		MessageId: generateMessageID(),
		Timestamp: timestampNow(),
		Payload: &oakv1.ServerMessage_ConfigUpdate{
			ConfigUpdate: &oakv1.ConfigUpdate{
				Config: config,
			},
		},
	}

	return r.sendMessage(agentID, msg)
}

// CheckHealth checks all agents for stale heartbeats
func (r *Registry) CheckHealth(timeout time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, info := range r.agents {
		if now.Sub(info.LastHeartbeat) > timeout {
			info.Status = oakv1.AgentStatus_AGENT_STATUS_UNHEALTHY
		}
	}
}

// Error definitions
var (
	ErrAgentNotFound      = errors.New("agent not found")
	ErrSendChannelFull    = errors.New("agent send channel is full")
	ErrAgentDisconnected  = errors.New("agent is disconnected")
)

// Helper functions
func generateMessageID() string {
	return uuid.New().String()
}

func timestampNow() *timestamppb.Timestamp {
	return timestamppb.Now()
}
