package node

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/init/logger"
	pb "github.com/raphaelCamblong/Luminous-Mesh/control-plane/shared/proto"
	"go.uber.org/zap"
)

type Node struct {
	ID           string
	BasicInfo    *pb.NodeBasicInfo
	Capabilities *pb.NodeCapabilities
	Status       *pb.NodeStatus
	Sessions     map[string]time.Time
	LastSeen     time.Time
}

type Manager struct {
	nodes sync.Map
	mu    sync.RWMutex
}

func NewManager() (*Manager, error) {
	return &Manager{}, nil
}

func (m *Manager) GenerateNodeID() string {
	return uuid.New().String()
}

func (m *Manager) RegisterNode(nodeID string, info *pb.NodeBasicInfo) error {
	node := &Node{
		ID:        nodeID,
		BasicInfo: info,
		Sessions:  make(map[string]time.Time),
		LastSeen:  time.Now(),
	}

	m.nodes.Store(nodeID, node)
	logger.L().Info("Node registered",
		zap.String("node_id", nodeID),
		zap.String("hostname", info.Hostname),
	)
	return nil
}

// CreateSession creates a new session for a node
func (m *Manager) CreateSession(nodeID string) (string, error) {
	nodeIface, ok := m.nodes.Load(nodeID)
	if !ok {
		return "", fmt.Errorf("node not found")
	}

	node := nodeIface.(*Node)
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clean up old sessions
	now := time.Now()
	for sessionID, lastActivity := range node.Sessions {
		if now.Sub(lastActivity) > 24*time.Hour {
			delete(node.Sessions, sessionID)
		}
	}

	// Create new session
	sessionID := uuid.New().String()
	node.Sessions[sessionID] = now

	return sessionID, nil
}

// ValidateSession validates a session
func (m *Manager) ValidateSession(nodeID, sessionID string) error {
	nodeIface, ok := m.nodes.Load(nodeID)
	if !ok {
		return fmt.Errorf("node not found")
	}

	node := nodeIface.(*Node)
	m.mu.RLock()
	lastActivity, exists := node.Sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found")
	}

	if time.Since(lastActivity) > 24*time.Hour {
		return fmt.Errorf("session expired")
	}

	return nil
}

// UpdateNodeInfo updates a node's information
func (m *Manager) UpdateNodeInfo(nodeID string, info *pb.NodeBasicInfo, capabilities *pb.NodeCapabilities) error {
	nodeIface, ok := m.nodes.Load(nodeID)
	if !ok {
		return fmt.Errorf("node not found")
	}

	node := nodeIface.(*Node)
	m.mu.Lock()
	node.BasicInfo = info
	node.Capabilities = capabilities
	node.LastSeen = time.Now()
	m.mu.Unlock()

	return nil
}

// UpdateNodeStatus updates a node's status
func (m *Manager) UpdateNodeStatus(nodeID string, status *pb.NodeStatus) error {
	nodeIface, ok := m.nodes.Load(nodeID)
	if !ok {
		return fmt.Errorf("node not found")
	}

	node := nodeIface.(*Node)
	m.mu.Lock()
	node.Status = status
	node.LastSeen = time.Now()
	m.mu.Unlock()

	return nil
}

// GetNodeConfiguration returns a node's configuration
func (m *Manager) GetNodeConfiguration(nodeID string) *pb.NodeConfiguration {
	// In a production environment, this would load from a configuration store
	return &pb.NodeConfiguration{
		Settings: map[string]string{
			"log_level": "info",
			"mode":      "normal",
		},
		EnabledFeatures: []string{"metrics", "health_check"},
		ResourceLimits: &pb.ResourceLimits{
			MaxConcurrentTasks: 10,
			MaxMemoryMb:        1024,
			MaxCpuUsage:        0.8,
		},
	}
}

// GetNode returns a node by ID
func (m *Manager) GetNode(nodeID string) (*Node, error) {
	nodeIface, ok := m.nodes.Load(nodeID)
	if !ok {
		return nil, fmt.Errorf("node not found")
	}
	return nodeIface.(*Node), nil
}

// ListNodes returns all registered nodes
func (m *Manager) ListNodes() []*Node {
	var nodes []*Node
	m.nodes.Range(func(key, value interface{}) bool {
		nodes = append(nodes, value.(*Node))
		return true
	})
	return nodes
}

// RemoveNode removes a node
func (m *Manager) RemoveNode(nodeID string) {
	m.nodes.Delete(nodeID)
	logger.L().Info("Node removed", zap.String("node_id", nodeID))
}
