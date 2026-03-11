package services

import (
	"fmt"
	"sync"
	"time"
)

// ─── Types ────────────────────────────────────────────────────────────────────

// SandboxMessage is a message sent between agent sandboxes or from external callers.
type SandboxMessage struct {
	FromAgentID string `json:"fromagentid"`
	FromRunID   string `json:"fromrunid"`
	Topic       string `json:"topic"`
	Payload     string `json:"payload"` // JSON
	ReplyTo     string `json:"replyto"` // optional runID for response routing
}

// AgentSandbox is the runtime context for a single agent run.
type AgentSandbox struct {
	RunID     string
	AgentID   string
	StartedAt time.Time
	Inbox     chan SandboxMessage // buffered channel (cap=100)
}

// SandboxInfo is a lightweight view returned by ListActive.
type SandboxInfo struct {
	RunID     string    `json:"runid"`
	AgentID   string    `json:"agentid"`
	StartedAt time.Time `json:"startedat"`
}

// ─── Manager ──────────────────────────────────────────────────────────────────

// AgentSandboxManager manages the set of active agent sandboxes.
type AgentSandboxManager struct {
	mu        sync.RWMutex
	sandboxes map[string]*AgentSandbox // runID → sandbox
}

var (
	globalAgentSandboxManager *AgentSandboxManager
)

func GetGlobalAgentSandboxManager() *AgentSandboxManager { return globalAgentSandboxManager }
func SetGlobalAgentSandboxManager(m *AgentSandboxManager) { globalAgentSandboxManager = m }

func NewAgentSandboxManager() *AgentSandboxManager {
	return &AgentSandboxManager{
		sandboxes: make(map[string]*AgentSandbox),
	}
}

// Register creates a new sandbox for a run and returns it.
// Must be called before the goroutine starts.
func (m *AgentSandboxManager) Register(runID, agentID string) *AgentSandbox {
	sb := &AgentSandbox{
		RunID:     runID,
		AgentID:   agentID,
		StartedAt: time.Now(),
		Inbox:     make(chan SandboxMessage, 100),
	}
	m.mu.Lock()
	m.sandboxes[runID] = sb
	m.mu.Unlock()
	return sb
}

// Unregister removes the sandbox when the run completes.
func (m *AgentSandboxManager) Unregister(runID string) {
	m.mu.Lock()
	if sb, ok := m.sandboxes[runID]; ok {
		close(sb.Inbox)
		delete(m.sandboxes, runID)
	}
	m.mu.Unlock()
}

// Send delivers a message to the sandbox inbox identified by runID.
func (m *AgentSandboxManager) Send(toRunID string, msg SandboxMessage) error {
	m.mu.RLock()
	sb, ok := m.sandboxes[toRunID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("no active sandbox for run %s", toRunID)
	}
	select {
	case sb.Inbox <- msg:
		return nil
	default:
		return fmt.Errorf("sandbox inbox full for run %s", toRunID)
	}
}

// SendToAgent delivers a message to the first active sandbox for the given agentID.
func (m *AgentSandboxManager) SendToAgent(toAgentID string, msg SandboxMessage) error {
	m.mu.RLock()
	var target *AgentSandbox
	for _, sb := range m.sandboxes {
		if sb.AgentID == toAgentID {
			target = sb
			break
		}
	}
	m.mu.RUnlock()
	if target == nil {
		return fmt.Errorf("no active sandbox for agent %s", toAgentID)
	}
	select {
	case target.Inbox <- msg:
		return nil
	default:
		return fmt.Errorf("sandbox inbox full for agent %s", toAgentID)
	}
}

// Broadcast sends a message to all active sandboxes matching the given topic prefix.
// Pass topic="" to broadcast to all active sandboxes.
func (m *AgentSandboxManager) Broadcast(topic string, msg SandboxMessage) {
	m.mu.RLock()
	targets := make([]*AgentSandbox, 0, len(m.sandboxes))
	for _, sb := range m.sandboxes {
		targets = append(targets, sb)
	}
	m.mu.RUnlock()
	for _, sb := range targets {
		select {
		case sb.Inbox <- msg:
		default:
			// drop on full inbox — non-blocking broadcast
		}
	}
}

// ListActive returns info on all currently active sandboxes.
func (m *AgentSandboxManager) ListActive() []SandboxInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]SandboxInfo, 0, len(m.sandboxes))
	for _, sb := range m.sandboxes {
		result = append(result, SandboxInfo{
			RunID:     sb.RunID,
			AgentID:   sb.AgentID,
			StartedAt: sb.StartedAt,
		})
	}
	return result
}
