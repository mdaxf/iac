package models

import "time"

// RunInstance constants for AgentDefinition.RunInstance.
const (
	RunInstanceApp            = "app"
	RunInstanceJobExecutor    = "job_executor"
	RunInstanceIntegrationHub = "integration_hub"
	// RunInstanceChannelMonitor marks agents that are dedicated to handling
	// bidirectional channel chat (Telegram, Slack, iMessage, etc.).
	// These agents are managed by AgentMonitorService and appear in the
	// Agent Monitor dashboard.
	RunInstanceChannelMonitor = "channel_monitor"
)

// ─── Context keys injected by AgentChannelService.HandleInbound ────────────────
// AgentMonitorService.HandleMessage reads these to attribute stats to the right channel.

type MonitorChannelIDKey struct{}
type MonitorChannelNameKey struct{}
type MonitorChannelTypeKey struct{}
type MonitorSenderIDKey struct{}
type MonitorSenderNameKey struct{}

// ─── In-memory status structs (not persisted) ────────────────────────────────

// MonitorSessionStatus represents the live state of one channel-user session
// as tracked by AgentMonitorService.
type MonitorSessionStatus struct {
	ChannelID     string    `json:"channel_id"`
	ChannelName   string    `json:"channel_name"`
	ChannelType   string    `json:"channel_type"`
	AgentID       string    `json:"agent_id"`
	AgentName     string    `json:"agent_name"`
	SenderID      string    `json:"sender_id"`
	SenderName    string    `json:"sender_name"`
	ConvID        string    `json:"conversation_id"`
	// Status: "idle" | "processing" | "error"
	Status        string    `json:"status"`
	LastMessageAt time.Time `json:"last_message_at"`
	MessageCount  int       `json:"message_count"`
	StartedAt     time.Time `json:"started_at"`
	LastResponseMs int64    `json:"last_response_ms"`
}

// ChannelMessageStats holds aggregated statistics for one channel binding.
type ChannelMessageStats struct {
	ChannelID      string `json:"channel_id"`
	ChannelName    string `json:"channel_name"`
	ChannelType    string `json:"channel_type"`
	AgentName      string `json:"agent_name"`
	TotalMessages  int64  `json:"total_messages"`
	ActiveSessions int    `json:"active_sessions"`
	AvgResponseMs  int64  `json:"avg_response_ms"`
	ErrorCount     int64  `json:"error_count"`
	LastActivityAt string `json:"last_activity_at"` // RFC 3339
}

// MonitorStatus is the top-level payload returned by GET /api/agentmonitor/status.
type MonitorStatus struct {
	TotalChannels      int                    `json:"total_channels"`
	ActiveChannels     int                    `json:"active_channels"`
	TotalSessions      int                    `json:"total_sessions"`
	ProcessingSessions int                    `json:"processing_sessions"`
	TotalMessages      int64                  `json:"total_messages"`
	ChannelStats       []*ChannelMessageStats `json:"channel_stats"`
	UptimeSecs         int64                  `json:"uptime_secs"`
}

// ActiveMonitorInfo is returned by GET /api/agentmonitor/monitors and lists all
// running channel_monitor (agent, channel) bindings.
type ActiveMonitorInfo struct {
	AgentID     string    `json:"agent_id"`
	AgentName   string    `json:"agent_name"`
	ChannelID   string    `json:"channel_id"`
	ChannelName string    `json:"channel_name"`
	ChannelType string    `json:"channel_type"`
	StartedAt   time.Time `json:"started_at"`
}
