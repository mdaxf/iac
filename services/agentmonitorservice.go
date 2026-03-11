package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/signalrserver"
)

// ─── Channel Monitor Runner ───────────────────────────────────────────────────

// channelMonitorEntry represents one running agent-channel binding.
type channelMonitorEntry struct {
	agentID      string
	agentName    string
	channelID    string
	channelName  string
	channelType  string
	config       map[string]string // resolved channel_config
	startMessage string            // broadcast when poller starts (empty = use default)
	stopMessage  string            // broadcast when poller stops (empty = use default)
	startedAt    time.Time
	stopFunc     context.CancelFunc // stops the channel-type polling goroutine
}

// activeMonitorKey is the composite key for (agentID, channelID).
type activeMonitorKey struct {
	agentID   string
	channelID string
}

// sessionKey is the composite map key for tracking channel-user sessions.
type sessionKey struct {
	channelID string
	senderID  string
}

// internalSession holds the mutable live state for one channel-user session.
type internalSession struct {
	mu sync.Mutex
	models.MonitorSessionStatus
}

// internalChannelStats holds running totals for a single channel.
type internalChannelStats struct {
	mu           sync.Mutex
	meta         models.ChannelMessageStats
	totalRespMs  int64
	respCount    int64
}

// AgentMonitorService manages the lifecycle and observability of agent-channel chat sessions.
//
// It acts as a transparent wrapper around the actual agent execution callback:
// AgentChannelService registers HandleMessage as its runAgentInConv callback,
// so every channel message flows through the monitor, which records timing,
// updates session status, and aggregates per-channel statistics.
//
// Additionally, a background goroutine periodically:
//   - Evicts stale sessions
//   - Scans for agent_definitions with run_instance=channel_monitor and their
//     linked agent_channels, registering them as active channel monitors.
//
// The active monitor registry is authoritative for which agents are "running"
// as channel handlers. Channels whose linked agent has run_instance=channel_monitor
// are processed in interactive/conversation mode (RunSyncInConversation).
type AgentMonitorService struct {
	// runAgentInConv is the actual execution backend (AgentRunnerService.RunSyncInConversation).
	// Injected via SetRunAgentInConvFunc to avoid import cycles.
	runAgentInConv func(ctx context.Context, convID, agentID, agentName, prompt string, timeoutSecs int) (string, error)

	// defSvc and channelSvc are injected so the monitor can scan for channel_monitor agents.
	defSvc     *AgentDefinitionService
	channelSvc *AgentChannelService

	// Session registry — keyed by (channelID, senderID)
	sessionsMu sync.RWMutex
	sessions   map[sessionKey]*internalSession

	// Per-channel stats
	statsMu sync.RWMutex
	stats   map[string]*internalChannelStats // key = channelID

	// Active channel monitor registry — keyed by (agentID, channelID).
	// Populated at startup and refreshed every 60 s by the background loop.
	monitorsMu sync.RWMutex
	monitors   map[activeMonitorKey]*channelMonitorEntry

	// pollerWg tracks all running poller goroutines so StopAll can wait for them.
	pollerWg sync.WaitGroup

	startedAt time.Time
	stopChan  chan struct{}
	iLog      logger.Log
}

var globalAgentMonitorService *AgentMonitorService

func GetGlobalAgentMonitorService() *AgentMonitorService { return globalAgentMonitorService }
func SetGlobalAgentMonitorService(s *AgentMonitorService) { globalAgentMonitorService = s }

// NewAgentMonitorService creates a ready-to-use (but not yet started) monitor service.
// Call Start() after dependency injection to begin the background loop.
func NewAgentMonitorService() *AgentMonitorService {
	return &AgentMonitorService{
		sessions:  make(map[sessionKey]*internalSession),
		stats:     make(map[string]*internalChannelStats),
		monitors:  make(map[activeMonitorKey]*channelMonitorEntry),
		startedAt: time.Now(),
		stopChan:  make(chan struct{}),
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "AgentMonitorService",
		},
	}
}

// monitorBroadcast is the payload pushed to all SignalR UI clients on topic
// "agent.monitor.update" after each registry refresh or message-processing cycle.
type monitorBroadcast struct {
	Status   *models.MonitorStatus          `json:"status"`
	Sessions []*models.MonitorSessionStatus `json:"sessions"`
	Monitors []*models.ActiveMonitorInfo    `json:"monitors"`
}

// broadcastUpdate pushes the current monitor snapshot to all connected UI clients.
// Called in a goroutine after refreshChannelMonitors and after each HandleMessage.
func (s *AgentMonitorService) broadcastUpdate() {
	srv := signalrserver.GetGlobalSignalRServer()
	if srv == nil || !srv.IsRunning() {
		return
	}
	payload := monitorBroadcast{
		Status:   s.GetStatus(),
		Sessions: s.GetActiveSessions(),
		Monitors: s.GetActiveMonitorInfos(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("broadcastUpdate: marshal error: %v", err))
		return
	}
	srv.BroadcastToHub("agent.monitor.update", string(data))
}

// startPoller is implemented in agentmonitorservice_telegram.go
// (dispatches to channel-type specific polling goroutines).

// SetRunAgentInConvFunc injects the actual execution callback.
// Must be called before HandleMessage is used.
func (s *AgentMonitorService) SetRunAgentInConvFunc(
	fn func(ctx context.Context, convID, agentID, agentName, prompt string, timeoutSecs int) (string, error),
) {
	s.runAgentInConv = fn
}

// SetDefinitionService injects the AgentDefinitionService so the monitor can
// scan for channel_monitor agents on startup and during periodic refresh.
func (s *AgentMonitorService) SetDefinitionService(svc *AgentDefinitionService) {
	s.defSvc = svc
}

// SetChannelService injects the AgentChannelService so the monitor can query
// which channels are bound to each channel_monitor agent.
func (s *AgentMonitorService) SetChannelService(svc *AgentChannelService) {
	s.channelSvc = svc
}

// Start launches the background monitoring goroutine and performs an initial
// scan of channel_monitor agents.
func (s *AgentMonitorService) Start() {
	// Initial scan — populate the active monitor registry before the first message.
	s.refreshChannelMonitors()
	go s.monitorLoop()
	s.iLog.Info("AgentMonitorService started")
}

// Stop signals the background loop to exit gracefully.
// For a full shutdown including all channel pollers, use StopAll instead.
func (s *AgentMonitorService) Stop() {
	select {
	case <-s.stopChan:
		// already closed
	default:
		close(s.stopChan)
	}
}

// StopAll performs a graceful full shutdown:
//  1. Cancels the context of every active channel poller (triggering their stop messages).
//  2. Waits for all pollers to finish, up to the supplied timeout.
//  3. Stops the background monitor loop.
//
// Call this from the process-level shutdown handler to ensure stop messages are
// delivered before os.Exit is called.
func (s *AgentMonitorService) StopAll(timeout time.Duration) {
	s.iLog.Info("AgentMonitorService: initiating graceful shutdown of all channel pollers")

	// Cancel every active poller context — this triggers case <-ctx.Done() in each
	// poller, which sends the stop message and then returns.
	s.monitorsMu.RLock()
	for _, entry := range s.monitors {
		if entry.stopFunc != nil {
			entry.stopFunc()
		}
	}
	s.monitorsMu.RUnlock()

	// Wait for all poller goroutines to finish, with a deadline.
	done := make(chan struct{})
	go func() {
		s.pollerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.iLog.Info("AgentMonitorService: all channel pollers stopped cleanly")
	case <-time.After(timeout):
		s.iLog.Warn(fmt.Sprintf("AgentMonitorService: shutdown timed out after %v — some stop messages may not have been delivered", timeout))
	}

	// Stop the background loop last.
	s.Stop()
}

// monitorLoop runs cleanup and registry refresh tasks periodically.
func (s *AgentMonitorService) monitorLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.cleanupStaleSessions()
			s.refreshChannelMonitors()
		case <-s.stopChan:
			s.iLog.Info("AgentMonitorService stopped")
			return
		}
	}
}

// refreshChannelMonitors scans for all enabled agents with run_instance=channel_monitor,
// loads their linked agent_channels, and updates the active monitor registry.
//
// This is equivalent to the integration_hub role starting up its HubExecutors —
// each channel_monitor agent "starts" by registering its channels here.
func (s *AgentMonitorService) refreshChannelMonitors() {
	if s.defSvc == nil {
		s.iLog.Warn("AgentMonitorService: refreshChannelMonitors skipped — defSvc is nil")
		return
	}
	if s.channelSvc == nil {
		s.iLog.Warn("AgentMonitorService: refreshChannelMonitors skipped — channelSvc is nil")
		return
	}

	agents, err := s.defSvc.ListByRunInstance(models.RunInstanceChannelMonitor)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("AgentMonitorService: failed to list channel_monitor agents: %v", err))
		return
	}
	s.iLog.Info(fmt.Sprintf("AgentMonitorService: found %d channel_monitor agent(s)", len(agents)))

	// Build a new registry snapshot
	newMonitors := make(map[activeMonitorKey]*channelMonitorEntry)
	for _, agent := range agents {
		s.iLog.Info(fmt.Sprintf("AgentMonitorService: scanning channels for agent id=%s name=%s run_instance=%s enabled=%v", agent.ID, agent.Name, agent.RunInstance, agent.Enabled))
		channels, err := s.channelSvc.GetByAgentID(agent.ID)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("AgentMonitorService: failed to load channels for agent %s: %v", agent.ID, err))
			continue
		}
		s.iLog.Info(fmt.Sprintf("AgentMonitorService: agent %s has %d channel binding(s)", agent.Name, len(channels)))
		for _, ch := range channels {
			s.iLog.Info(fmt.Sprintf("AgentMonitorService: channel id=%s name=%s type=%s enabled=%v active=%v", ch.ID, ch.DisplayName, ch.ChannelType, ch.Enabled, ch.Active))
			if !ch.Enabled || !ch.Active {
				s.iLog.Warn(fmt.Sprintf("AgentMonitorService: skipping channel %s (enabled=%v active=%v)", ch.DisplayName, ch.Enabled, ch.Active))
				continue
			}
			key := activeMonitorKey{agentID: agent.ID, channelID: ch.ID}

			// Resolve secrets from channel_config so pollers get the real values.
			resolvedConfig := s.channelSvc.ResolveChannelConfig(ch)

			// Preserve existing startedAt and poller if the entry already exists.
			s.monitorsMu.RLock()
			existing, exists := s.monitors[key]
			s.monitorsMu.RUnlock()

			var entry *channelMonitorEntry
			if exists {
				entry = existing
			} else {
				s.iLog.Info(fmt.Sprintf("AgentMonitorService: registering channel monitor agent=%s channel=%s (%s)", agent.Name, ch.DisplayName, ch.ChannelType))
				ctx, cancel := context.WithCancel(context.Background())
				startMsg := ch.StartMessage
				if startMsg == "" {
					startMsg = fmt.Sprintf("✅ Agent [%s] is now online and ready to chat.", agent.Name)
				}
				stopMsg := ch.StopMessage
				if stopMsg == "" {
					stopMsg = fmt.Sprintf("⚠️ Agent [%s] has gone offline. Communication is temporarily unavailable.", agent.Name)
				}
				entry = &channelMonitorEntry{
					agentID:      agent.ID,
					agentName:    agent.Name,
					channelID:    ch.ID,
					channelName:  ch.DisplayName,
					channelType:  ch.ChannelType,
					config:       resolvedConfig,
					startMessage: startMsg,
					stopMessage:  stopMsg,
					startedAt:    time.Now(),
					stopFunc:     cancel,
				}
				// Start a channel-type-specific poller goroutine.
				s.startPoller(ctx, entry)
			}
			newMonitors[key] = entry
		}
	}

	// Stop pollers for channels that are no longer in the registry.
	s.monitorsMu.RLock()
	for k, old := range s.monitors {
		if _, still := newMonitors[k]; !still && old.stopFunc != nil {
			s.iLog.Info(fmt.Sprintf("AgentMonitorService: stopping removed channel monitor channel=%s", old.channelName))
			old.stopFunc()
		}
	}
	s.monitorsMu.RUnlock()

	s.monitorsMu.Lock()
	s.monitors = newMonitors
	s.monitorsMu.Unlock()

	s.iLog.Info(fmt.Sprintf("AgentMonitorService: %d channel monitor binding(s) active", len(newMonitors)))
	go s.broadcastUpdate()
}

// IsChannelMonitored returns true if the given channelID is bound to a running
// channel_monitor agent. Callers can use this to enforce the run_instance contract.
func (s *AgentMonitorService) IsChannelMonitored(channelID string) bool {
	s.monitorsMu.RLock()
	defer s.monitorsMu.RUnlock()
	for k := range s.monitors {
		if k.channelID == channelID {
			return true
		}
	}
	return false
}

// GetActiveMonitors returns a snapshot of all currently active channel monitor bindings.
func (s *AgentMonitorService) GetActiveMonitors() []*channelMonitorEntry {
	s.monitorsMu.RLock()
	defer s.monitorsMu.RUnlock()
	result := make([]*channelMonitorEntry, 0, len(s.monitors))
	for _, m := range s.monitors {
		cp := *m
		result = append(result, &cp)
	}
	return result
}

// GetActiveMonitorInfos returns the active monitors as public API structs.
func (s *AgentMonitorService) GetActiveMonitorInfos() []*models.ActiveMonitorInfo {
	s.monitorsMu.RLock()
	defer s.monitorsMu.RUnlock()
	result := make([]*models.ActiveMonitorInfo, 0, len(s.monitors))
	for _, m := range s.monitors {
		result = append(result, &models.ActiveMonitorInfo{
			AgentID:     m.agentID,
			AgentName:   m.agentName,
			ChannelID:   m.channelID,
			ChannelName: m.channelName,
			ChannelType: m.channelType,
			StartedAt:   m.startedAt,
		})
	}
	return result
}

// cleanupStaleSessions removes sessions idle for >24 hours from the in-memory map.
func (s *AgentMonitorService) cleanupStaleSessions() {
	cutoff := time.Now().Add(-24 * time.Hour)
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	removed := 0
	for k, sess := range s.sessions {
		sess.mu.Lock()
		idle := sess.LastMessageAt.Before(cutoff) && sess.Status != "processing"
		sess.mu.Unlock()
		if idle {
			delete(s.sessions, k)
			removed++
		}
	}
	if removed > 0 {
		s.iLog.Info(fmt.Sprintf("AgentMonitorService: evicted %d stale sessions", removed))
	}
}

// ─── Core: HandleMessage ──────────────────────────────────────────────────────

// HandleMessage is registered as the runAgentInConv callback on AgentChannelService.
// It decorates every channel message execution with:
//   - Session status tracking (idle → processing → idle/error)
//   - Response time measurement
//   - Per-channel message statistics
//
// Context keys injected by AgentChannelService.HandleInbound are used to
// identify which channel and sender triggered this execution.
func (s *AgentMonitorService) HandleMessage(
	ctx context.Context,
	convID, agentID, agentName, text string,
	timeoutSecs int,
) (string, error) {
	if s.runAgentInConv == nil {
		return "", fmt.Errorf("AgentMonitorService: runAgentInConv callback not set")
	}

	// Extract channel/sender identity from context (injected by AgentChannelService.HandleInbound)
	channelID, _ := ctx.Value(models.MonitorChannelIDKey{}).(string)
	channelName, _ := ctx.Value(models.MonitorChannelNameKey{}).(string)
	channelType, _ := ctx.Value(models.MonitorChannelTypeKey{}).(string)
	senderID, _ := ctx.Value(models.MonitorSenderIDKey{}).(string)
	senderName, _ := ctx.Value(models.MonitorSenderNameKey{}).(string)

	// Mark session as processing
	if channelID != "" && senderID != "" {
		s.upsertSession(channelID, channelName, channelType, senderID, senderName, agentID, agentName, convID, "processing")
	}

	start := time.Now()

	response, err := s.runAgentInConv(ctx, convID, agentID, agentName, text, timeoutSecs)

	elapsed := time.Since(start).Milliseconds()

	// Record per-channel statistics
	if channelID != "" {
		s.recordStat(channelID, channelName, channelType, agentName, elapsed, err != nil)
	}

	// Update session status
	if channelID != "" && senderID != "" {
		status := "idle"
		if err != nil {
			status = "error"
		}
		s.upsertSessionFinished(channelID, senderID, status, elapsed)
	}

	// Push updated monitor state to all connected UI clients.
	go s.broadcastUpdate()

	return response, err
}

// ─── Session helpers ──────────────────────────────────────────────────────────

func (s *AgentMonitorService) upsertSession(
	channelID, channelName, channelType, senderID, senderName, agentID, agentName, convID, status string,
) {
	key := sessionKey{channelID: channelID, senderID: senderID}
	now := time.Now()

	s.sessionsMu.Lock()
	sess, exists := s.sessions[key]
	if !exists {
		sess = &internalSession{
			MonitorSessionStatus: models.MonitorSessionStatus{
				ChannelID:   channelID,
				ChannelName: channelName,
				ChannelType: channelType,
				SenderID:    senderID,
				SenderName:  senderName,
				AgentID:     agentID,
				AgentName:   agentName,
				ConvID:      convID,
				StartedAt:   now,
			},
		}
		s.sessions[key] = sess
	}
	s.sessionsMu.Unlock()

	sess.mu.Lock()
	sess.Status = status
	sess.LastMessageAt = now
	if exists {
		sess.MessageCount++
	} else {
		sess.MessageCount = 1
	}
	if convID != "" {
		sess.ConvID = convID
	}
	sess.mu.Unlock()
}

func (s *AgentMonitorService) upsertSessionFinished(channelID, senderID, status string, elapsedMs int64) {
	key := sessionKey{channelID: channelID, senderID: senderID}
	s.sessionsMu.RLock()
	sess, ok := s.sessions[key]
	s.sessionsMu.RUnlock()
	if !ok {
		return
	}
	sess.mu.Lock()
	sess.Status = status
	sess.LastResponseMs = elapsedMs
	sess.mu.Unlock()
}

// ─── Stats helpers ────────────────────────────────────────────────────────────

func (s *AgentMonitorService) getOrCreateChannelStats(channelID, channelName, channelType, agentName string) *internalChannelStats {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	cs, ok := s.stats[channelID]
	if !ok {
		cs = &internalChannelStats{
			meta: models.ChannelMessageStats{
				ChannelID:   channelID,
				ChannelName: channelName,
				ChannelType: channelType,
				AgentName:   agentName,
			},
		}
		s.stats[channelID] = cs
	}
	return cs
}

func (s *AgentMonitorService) recordStat(channelID, channelName, channelType, agentName string, elapsedMs int64, isErr bool) {
	cs := s.getOrCreateChannelStats(channelID, channelName, channelType, agentName)
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.meta.TotalMessages++
	cs.meta.LastActivityAt = time.Now().UTC().Format(time.RFC3339)
	if isErr {
		cs.meta.ErrorCount++
	}
	cs.totalRespMs += elapsedMs
	cs.respCount++
	if cs.respCount > 0 {
		cs.meta.AvgResponseMs = cs.totalRespMs / cs.respCount
	}
}

// ─── Status / Query API ────────────────────────────────────────────────────────

// GetStatus returns the top-level AgentMonitor status summary.
func (s *AgentMonitorService) GetStatus() *models.MonitorStatus {
	s.sessionsMu.RLock()
	allSessions := make([]*internalSession, 0, len(s.sessions))
	for _, sess := range s.sessions {
		allSessions = append(allSessions, sess)
	}
	s.sessionsMu.RUnlock()

	processing := 0
	for _, sess := range allSessions {
		sess.mu.Lock()
		if sess.Status == "processing" {
			processing++
		}
		sess.mu.Unlock()
	}

	s.statsMu.RLock()
	channelStats := make([]*models.ChannelMessageStats, 0, len(s.stats))
	var totalMessages int64
	activeChannels := 0
	for _, cs := range s.stats {
		cs.mu.Lock()
		snapshot := cs.meta // shallow copy — all fields are value types
		cs.mu.Unlock()
		channelStats = append(channelStats, &snapshot)
		totalMessages += snapshot.TotalMessages
		if snapshot.LastActivityAt != "" {
			activeChannels++
		}
	}
	s.statsMu.RUnlock()

	// Count active sessions per channel for stats
	sessionsByChannel := make(map[string]int)
	for _, sess := range allSessions {
		sess.mu.Lock()
		cid := sess.ChannelID
		sess.mu.Unlock()
		sessionsByChannel[cid]++
	}
	for _, cs := range channelStats {
		cs.ActiveSessions = sessionsByChannel[cs.ChannelID]
	}

	s.monitorsMu.RLock()
	registeredMonitors := len(s.monitors)
	s.monitorsMu.RUnlock()

	return &models.MonitorStatus{
		TotalChannels:      registeredMonitors,
		ActiveChannels:     activeChannels,
		TotalSessions:      len(allSessions),
		ProcessingSessions: processing,
		TotalMessages:      totalMessages,
		ChannelStats:       channelStats,
		UptimeSecs:         int64(time.Since(s.startedAt).Seconds()),
	}
}

// GetActiveSessions returns a snapshot of all currently tracked sessions.
func (s *AgentMonitorService) GetActiveSessions() []*models.MonitorSessionStatus {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()
	result := make([]*models.MonitorSessionStatus, 0, len(s.sessions))
	for _, sess := range s.sessions {
		sess.mu.Lock()
		snap := sess.MonitorSessionStatus // shallow copy
		sess.mu.Unlock()
		result = append(result, &snap)
	}
	return result
}

// GetSessionsByChannel returns sessions for a specific channel.
func (s *AgentMonitorService) GetSessionsByChannel(channelID string) []*models.MonitorSessionStatus {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()
	var result []*models.MonitorSessionStatus
	for k, sess := range s.sessions {
		if k.channelID != channelID {
			continue
		}
		sess.mu.Lock()
		snap := sess.MonitorSessionStatus
		sess.mu.Unlock()
		result = append(result, &snap)
	}
	return result
}

// ResetSession ends the tracked session for a given channel user, causing the
// next message to start a fresh conversation via AgentChannelService.
// The in-memory session entry is removed; the AgentChannelSession document
// in MongoDB is deactivated via channelSvc.DeactivateSession.
func (s *AgentMonitorService) ResetSession(channelSvc *AgentChannelService, channelID, senderID string) error {
	key := sessionKey{channelID: channelID, senderID: senderID}

	s.sessionsMu.Lock()
	_, exists := s.sessions[key]
	if exists {
		delete(s.sessions, key)
	}
	s.sessionsMu.Unlock()

	// Deactivate the MongoDB session doc so the next message creates a new conversation.
	if channelSvc != nil {
		if err := channelSvc.DeactivateSession(channelID, senderID); err != nil {
			s.iLog.Error(fmt.Sprintf("ResetSession: failed to deactivate MongoDB session: %v", err))
			return err
		}
	}

	s.iLog.Info(fmt.Sprintf("ResetSession: session reset for channel=%s sender=%s", channelID, senderID))
	return nil
}
