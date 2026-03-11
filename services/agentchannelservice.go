package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// AgentChannelService manages agent-channel bindings and routes inbound messages
// from external messaging platforms (via OpenClaw bridge or native adapters) to IAC agents.
type AgentChannelService struct {
	docDB    documents.DocumentDB
	chatSvc  *AgentChatService
	// runAgent is kept for compatibility (used by AgentGateway pattern), but channel
	// runs use runAgentInConv to maintain conversation history across messages.
	runAgent        func(ctx context.Context, agentID, agentName, prompt string, timeoutSecs int) (string, error)
	runAgentInConv  func(ctx context.Context, conversationID, agentID, agentName, prompt string, timeoutSecs int) (string, error)
	iLog     logger.Log

	// In-memory rate limit counters: key = channelID + ":" + senderID
	rateMu   sync.Mutex
	rateHits map[string]rateEntry

	// senderMu serialises HandleInbound calls per (channelID, senderID) pair.
	// This prevents concurrent agent runs on the same conversation, which would
	// cause history ordering bugs and race conditions in getOrCreateConversation.
	// Different senders (and different channels) run concurrently as normal.
	senderMu sync.Map // key: "channelID:senderID" → *sync.Mutex
}

type rateEntry struct {
	count     int
	windowEnd time.Time
}

var globalAgentChannelService *AgentChannelService

func GetGlobalAgentChannelService() *AgentChannelService { return globalAgentChannelService }
func SetGlobalAgentChannelService(s *AgentChannelService) { globalAgentChannelService = s }

// NewAgentChannelService creates the service and initialises MongoDB indexes.
func NewAgentChannelService(docDB documents.DocumentDB) *AgentChannelService {
	svc := &AgentChannelService{
		docDB: docDB,
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "AgentChannelService",
		},
		rateHits: make(map[string]rateEntry),
	}
	if docDB != nil {
		svc.initIndexes()
	}
	return svc
}

// SetChatService injects the AgentChatService (avoids import cycle).
func (s *AgentChannelService) SetChatService(svc *AgentChatService) {
	s.chatSvc = svc
}

// SetRunAgentFunc injects the stateless agent execution callback (avoids import cycle).
func (s *AgentChannelService) SetRunAgentFunc(fn func(ctx context.Context, agentID, agentName, prompt string, timeoutSecs int) (string, error)) {
	s.runAgent = fn
}

// SetRunAgentInConvFunc injects the conversation-aware execution callback.
// This is what HandleInbound uses so each channel message continues the same
// conversation thread and the agent sees the full prior message history.
func (s *AgentChannelService) SetRunAgentInConvFunc(fn func(ctx context.Context, conversationID, agentID, agentName, prompt string, timeoutSecs int) (string, error)) {
	s.runAgentInConv = fn
}

func (s *AgentChannelService) initIndexes() {
	ctx := context.Background()
	_ = s.docDB.CreateIndex(ctx, models.CollectionAgentChannels,
		map[string]int{"agent_id": 1},
		&documents.IndexOptions{Name: "idx_channel_agent"})
	_ = s.docDB.CreateIndex(ctx, models.CollectionAgentChannels,
		map[string]int{"enabled": 1, "active": 1},
		&documents.IndexOptions{Name: "idx_channel_enabled"})
	_ = s.docDB.CreateIndex(ctx, models.CollectionChannelUserMappings,
		map[string]int{"agent_channel_id": 1, "channel_user_id": 1},
		&documents.IndexOptions{Name: "idx_mapping_channel_user"})
	_ = s.docDB.CreateIndex(ctx, models.CollectionAgentChannelSessions,
		map[string]int{"agent_channel_id": 1, "channel_user_id": 1},
		&documents.IndexOptions{Name: "idx_session_channel_user"})
}

func (s *AgentChannelService) ensureDocDB() error {
	if s.docDB == nil {
		return fmt.Errorf("document database not initialized for AgentChannelService")
	}
	return nil
}

// ─── CRUD ─────────────────────────────────────────────────────────────────────

func (s *AgentChannelService) Create(doc *models.AgentChannelDoc) error {
	if err := s.ensureDocDB(); err != nil {
		return err
	}
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	doc.CreatedOn = now
	doc.ModifiedOn = now
	doc.Active = true
	if doc.ChannelConfig == nil {
		doc.ChannelConfig = map[string]string{}
	}
	if doc.SecurityPolicy.AllowedTools == nil {
		doc.SecurityPolicy.AllowedTools = []string{}
	}
	if doc.SecurityPolicy.BlockedTools == nil {
		doc.SecurityPolicy.BlockedTools = []string{}
	}
	_, err := s.docDB.InsertOne(context.Background(), models.CollectionAgentChannels, doc)
	return err
}

func (s *AgentChannelService) GetAll() ([]*models.AgentChannelDoc, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	filter := map[string]interface{}{"active": true}
	opts := &documents.FindOptions{Sort: map[string]int{"display_name": 1}}
	rows, err := s.docDB.FindMany(context.Background(), models.CollectionAgentChannels, filter, opts)
	if err != nil {
		return nil, err
	}
	result := make([]*models.AgentChannelDoc, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapToChannelDoc(row))
	}
	return result, nil
}

func (s *AgentChannelService) GetByID(id string) (*models.AgentChannelDoc, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	row, err := s.docDB.FindOne(context.Background(), models.CollectionAgentChannels, map[string]interface{}{"_id": id})
	if err != nil {
		return nil, err
	}
	return mapToChannelDoc(row), nil
}

func (s *AgentChannelService) GetByAgentID(agentID string) ([]*models.AgentChannelDoc, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	filter := map[string]interface{}{"agent_id": agentID, "active": true}
	rows, err := s.docDB.FindMany(context.Background(), models.CollectionAgentChannels, filter, nil)
	if err != nil {
		return nil, err
	}
	result := make([]*models.AgentChannelDoc, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapToChannelDoc(row))
	}
	return result, nil
}

func (s *AgentChannelService) Update(id string, doc *models.AgentChannelDoc) error {
	if err := s.ensureDocDB(); err != nil {
		return err
	}
	doc.ModifiedOn = time.Now().UTC().Format(time.RFC3339)
	if doc.ChannelConfig == nil {
		doc.ChannelConfig = map[string]string{}
	}
	if doc.SecurityPolicy.AllowedTools == nil {
		doc.SecurityPolicy.AllowedTools = []string{}
	}
	if doc.SecurityPolicy.BlockedTools == nil {
		doc.SecurityPolicy.BlockedTools = []string{}
	}

	// Convert channel_config to map[string]interface{} for safe BSON serialization
	channelConfigI := make(map[string]interface{}, len(doc.ChannelConfig))
	for k, v := range doc.ChannelConfig {
		channelConfigI[k] = v
	}

	fields := map[string]interface{}{
		"agent_id":           doc.AgentID,
		"agent_name":         doc.AgentName,
		"channel_type":       doc.ChannelType,
		"display_name":       doc.DisplayName,
		"channel_config":     channelConfigI,
		"welcome_message":    doc.WelcomeMessage,
		"start_message":      doc.StartMessage,
		"stop_message":       doc.StopMessage,
		"coordinator_mode":   doc.CoordinatorMode,
		"coordinator_prompt": doc.CoordinatorPrompt,
		"enabled":            doc.Enabled,
		"modifiedby":         doc.ModifiedBy,
		"modifiedon":         doc.ModifiedOn,
		"security_policy": map[string]interface{}{
			"require_user_auth":      doc.SecurityPolicy.RequireUserAuth,
			"allowed_tools":          doc.SecurityPolicy.AllowedTools,
			"blocked_tools":          doc.SecurityPolicy.BlockedTools,
			"max_iterations":         doc.SecurityPolicy.MaxIterations,
			"max_response_length":    doc.SecurityPolicy.MaxResponseLength,
			"rate_limit_per_hour":    doc.SecurityPolicy.RateLimitPerHour,
			"system_prompt_override": doc.SecurityPolicy.SystemPromptOverride,
		},
	}
	filter := map[string]interface{}{"_id": id}
	update := map[string]interface{}{"$set": fields}
	return s.docDB.UpdateOne(context.Background(), models.CollectionAgentChannels, filter, update)
}

func (s *AgentChannelService) Delete(id string) error {
	if err := s.ensureDocDB(); err != nil {
		return err
	}
	filter := map[string]interface{}{"_id": id}
	update := map[string]interface{}{"$set": map[string]interface{}{
		"active":     false,
		"enabled":    false,
		"modifiedon": time.Now().UTC().Format(time.RFC3339),
	}}
	return s.docDB.UpdateOne(context.Background(), models.CollectionAgentChannels, filter, update)
}

// ─── User Mapping ─────────────────────────────────────────────────────────────

func (s *AgentChannelService) ListMappings(channelID string) ([]*models.ChannelUserMappingDoc, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	filter := map[string]interface{}{"agent_channel_id": channelID, "active": true}
	rows, err := s.docDB.FindMany(context.Background(), models.CollectionChannelUserMappings, filter, nil)
	if err != nil {
		return nil, err
	}
	result := make([]*models.ChannelUserMappingDoc, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapToUserMappingDoc(row))
	}
	return result, nil
}

func (s *AgentChannelService) AuthorizeUser(channelID, channelUserID, channelUsername, iacUserID, iacUsername, authorizedBy string) error {
	if err := s.ensureDocDB(); err != nil {
		return err
	}
	ch, err := s.GetByID(channelID)
	if err != nil {
		return fmt.Errorf("channel not found: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)

	// Check if mapping already exists
	filter := map[string]interface{}{
		"agent_channel_id": channelID,
		"channel_user_id":  channelUserID,
	}
	existing, _ := s.docDB.FindOne(context.Background(), models.CollectionChannelUserMappings, filter)
	if existing != nil {
		update := map[string]interface{}{"$set": map[string]interface{}{
			"authorized":      true,
			"authorized_by":   authorizedBy,
			"authorized_on":   now,
			"iac_user_id":     iacUserID,
			"iac_username":    iacUsername,
			"channel_username": channelUsername,
			"active":          true,
		}}
		return s.docDB.UpdateOne(context.Background(), models.CollectionChannelUserMappings, filter, update)
	}

	doc := &models.ChannelUserMappingDoc{
		ID:              uuid.New().String(),
		AgentChannelID:  channelID,
		ChannelType:     ch.ChannelType,
		ChannelUserID:   channelUserID,
		ChannelUsername: channelUsername,
		IACUserID:       iacUserID,
		IACUsername:     iacUsername,
		Authorized:      true,
		AuthorizedBy:    authorizedBy,
		AuthorizedOn:    now,
		Active:          true,
		CreatedOn:       now,
	}
	_, err = s.docDB.InsertOne(context.Background(), models.CollectionChannelUserMappings, doc)
	return err
}

func (s *AgentChannelService) RevokeUser(channelID, channelUserID string) error {
	if err := s.ensureDocDB(); err != nil {
		return err
	}
	filter := map[string]interface{}{
		"agent_channel_id": channelID,
		"channel_user_id":  channelUserID,
	}
	update := map[string]interface{}{"$set": map[string]interface{}{
		"authorized": false,
	}}
	return s.docDB.UpdateOne(context.Background(), models.CollectionChannelUserMappings, filter, update)
}

// ─── Core: HandleInbound ──────────────────────────────────────────────────────

// HandleInbound processes a single inbound message from an external channel user.
//
// Full flow:
//  1. Load channel config + security policy
//  2. (HMAC verification already done in controller before this is called)
//  3. Upsert ChannelUserMapping — auto-registers any new sender
//  4. If RequireUserAuth && sender not authorized → return polite refusal
//  5. Rate-limit check (in-memory per-user hourly window)
//  6. Get or create AgentChannelSession → persistent ConversationID for this sender
//  7. Inject AgentChannelSecurityPolicy into context (tool allow/block, maxIterations)
//  8. Call RunSyncInConversation(convID, agentID, text, ...) — loads full history so agent
//     has complete context of the prior exchange before generating its response
//  9. Increment session message count; update last_seen_on
//  10. Truncate response to MaxResponseLength if set; return to caller
func (s *AgentChannelService) HandleInbound(ctx context.Context, channelID, senderID, senderName, text string) (string, error) {
	if s.runAgentInConv == nil && s.runAgent == nil {
		return "", fmt.Errorf("no runAgent callback set on AgentChannelService")
	}

	channel, err := s.GetByID(channelID)
	if err != nil {
		return "", fmt.Errorf("channel %s not found: %w", channelID, err)
	}
	if !channel.Enabled || !channel.Active {
		return "", fmt.Errorf("channel %s is disabled", channelID)
	}

	// Register this sender on first contact (creates a pending-authorization mapping)
	s.upsertUserMapping(channelID, channel.ChannelType, senderID, senderName)

	// Authorization gate
	if channel.SecurityPolicy.RequireUserAuth {
		if !s.isUserAuthorized(channelID, senderID) {
			return "Sorry, you are not authorized to use this agent. Please contact the administrator.", nil
		}
	}

	// Per-user rate limiting
	if channel.SecurityPolicy.RateLimitPerHour > 0 {
		if limited, msg := s.checkRateLimit(channelID, senderID, channel.SecurityPolicy.RateLimitPerHour); limited {
			return msg, nil
		}
	}

	// Serialise all processing for this (channelID, senderID) pair.
	// Without this, rapid successive messages from the same user launch concurrent
	// agent runs on the same conversationID — causing history ordering bugs and
	// race conditions in getOrCreateConversation (two goroutines both see no session
	// and both create new PostgreSQL conversations). Messages from different senders
	// or different channels are unaffected and run concurrently as normal.
	senderKey := channelID + ":" + senderID
	muI, _ := s.senderMu.LoadOrStore(senderKey, &sync.Mutex{})
	senderLock := muI.(*sync.Mutex)
	senderLock.Lock()
	defer senderLock.Unlock()

	// Resolve or create the persistent conversation for this (channel, sender) pair.
	// This is the key to conversation continuity: every message from the same sender
	// re-uses the same conversationID so the agent loads all previous turns as history.
	convID, err := s.getOrCreateConversation(channelID, channel, senderID, senderName)
	if err != nil {
		return "", fmt.Errorf("failed to get/create conversation: %w", err)
	}

	// Inject security policy into context so AgentRunnerService applies
	// tool filtering, iteration cap, and system prompt override for this run.
	policy := channel.SecurityPolicy
	ctx = context.WithValue(ctx, models.AgentChannelPolicyKey{}, policy)

	// Coordinator mode: inject a system-prompt prefix that directs the agent to
	// analyse the message, pick the right tools (possibly several), and aggregate
	// results before replying. Uses the custom prompt when set, otherwise the
	// built-in default.
	if channel.CoordinatorMode {
		coordinatorPrompt := channel.CoordinatorPrompt
		if coordinatorPrompt == "" {
			coordinatorPrompt = models.DefaultCoordinatorPrompt
		}
		ctx = context.WithValue(ctx, models.AgentCoordinatorPromptKey{}, coordinatorPrompt)
	}

	// Inject channel/sender identity into context so AgentMonitorService can
	// attribute stats to the correct channel and user without changing the
	// runAgentInConv callback signature.
	ctx = context.WithValue(ctx, models.MonitorChannelIDKey{}, channelID)
	ctx = context.WithValue(ctx, models.MonitorChannelNameKey{}, channel.DisplayName)
	ctx = context.WithValue(ctx, models.MonitorChannelTypeKey{}, channel.ChannelType)
	ctx = context.WithValue(ctx, models.MonitorSenderIDKey{}, senderID)
	ctx = context.WithValue(ctx, models.MonitorSenderNameKey{}, senderName)

	// Timeout budget: 30s per allowed iteration, minimum 60s
	timeoutSecs := 120
	if policy.MaxIterations > 0 {
		timeoutSecs = policy.MaxIterations * 30
		if timeoutSecs < 60 {
			timeoutSecs = 60
		}
	}

	var response string

	// Prefer the conversation-aware runner so the agent sees full prior history.
	// Fall back to stateless RunSync only if RunSyncInConversation was not wired.
	if s.runAgentInConv != nil {
		response, err = s.runAgentInConv(ctx, convID, channel.AgentID, channel.AgentName, text, timeoutSecs)
	} else {
		// Fallback path (no conversation continuity) — should not happen in normal deployment
		s.iLog.Warn(fmt.Sprintf("HandleInbound: runAgentInConv not set for channel %s, falling back to stateless RunSync (no conversation history)", channelID))
		response, err = s.runAgent(ctx, channel.AgentID, channel.AgentName, text, timeoutSecs)
	}

	if err != nil {
		s.iLog.Error(fmt.Sprintf("HandleInbound agent run error channel=%s sender=%s: %v", channelID, senderID, err))
		return "Sorry, I encountered an error processing your request. Please try again.", nil
	}

	// Enforce response length cap
	if policy.MaxResponseLength > 0 && len(response) > policy.MaxResponseLength {
		response = response[:policy.MaxResponseLength] + "…"
	}

	// Update activity timestamps and message count
	s.updateLastSeen(channelID, senderID)
	s.incrementSessionMessageCount(channelID, senderID)

	return response, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (s *AgentChannelService) upsertUserMapping(channelID, channelType, senderID, senderName string) {
	if s.docDB == nil {
		return
	}
	filter := map[string]interface{}{
		"agent_channel_id": channelID,
		"channel_user_id":  senderID,
	}
	now := time.Now().UTC().Format(time.RFC3339)
	existing, _ := s.docDB.FindOne(context.Background(), models.CollectionChannelUserMappings, filter)
	if existing != nil {
		// Update channel_username if it changed
		_ = s.docDB.UpdateOne(context.Background(), models.CollectionChannelUserMappings, filter,
			map[string]interface{}{"$set": map[string]interface{}{
				"channel_username": senderName,
			}})
		return
	}
	doc := &models.ChannelUserMappingDoc{
		ID:              uuid.New().String(),
		AgentChannelID:  channelID,
		ChannelType:     channelType,
		ChannelUserID:   senderID,
		ChannelUsername: senderName,
		Authorized:      false,
		Active:          true,
		CreatedOn:       now,
	}
	_, _ = s.docDB.InsertOne(context.Background(), models.CollectionChannelUserMappings, doc)
}

func (s *AgentChannelService) isUserAuthorized(channelID, senderID string) bool {
	if s.docDB == nil {
		return false
	}
	filter := map[string]interface{}{
		"agent_channel_id": channelID,
		"channel_user_id":  senderID,
		"authorized":       true,
		"active":           true,
	}
	row, err := s.docDB.FindOne(context.Background(), models.CollectionChannelUserMappings, filter)
	return err == nil && row != nil
}

func (s *AgentChannelService) checkRateLimit(channelID, senderID string, maxPerHour int) (bool, string) {
	key := channelID + ":" + senderID
	now := time.Now()

	s.rateMu.Lock()
	defer s.rateMu.Unlock()

	entry, exists := s.rateHits[key]
	if !exists || now.After(entry.windowEnd) {
		// New window
		s.rateHits[key] = rateEntry{count: 1, windowEnd: now.Add(time.Hour)}
		return false, ""
	}
	entry.count++
	s.rateHits[key] = entry
	if entry.count > maxPerHour {
		return true, fmt.Sprintf("Rate limit exceeded. You can send at most %d messages per hour.", maxPerHour)
	}
	return false, ""
}

func (s *AgentChannelService) getOrCreateConversation(channelID string, channel *models.AgentChannelDoc, senderID, senderName string) (string, error) {
	if s.docDB == nil || s.chatSvc == nil {
		return "", fmt.Errorf("docDB or chatSvc not initialized")
	}
	filter := map[string]interface{}{
		"agent_channel_id": channelID,
		"channel_user_id":  senderID,
		"active":           true,
	}
	row, _ := s.docDB.FindOne(context.Background(), models.CollectionAgentChannelSessions, filter)
	if row != nil {
		if convID, ok := row["conversation_id"].(string); ok && convID != "" {
			return convID, nil
		}
	}

	// Create a new conversation
	title := fmt.Sprintf("%s — %s (%s)", channel.DisplayName, senderName, channel.ChannelType)
	conv, err := s.chatSvc.CreateAgentConversation(channel.AgentID, senderID, title, "channel")
	if err != nil {
		return "", err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	session := &models.AgentChannelSessionDoc{
		ID:             uuid.New().String(),
		AgentChannelID: channelID,
		ChannelUserID:  senderID,
		ConversationID: conv.ID,
		MessageCount:   0,
		LastMessageOn:  now,
		Active:         true,
		CreatedOn:      now,
	}
	_, _ = s.docDB.InsertOne(context.Background(), models.CollectionAgentChannelSessions, session)

	// Send welcome message if configured
	if channel.WelcomeMessage != "" && s.runAgent != nil {
		go func() {
			_, _ = s.chatSvc.SaveUserMessage(conv.ID, "[SYSTEM: User has started a new conversation via "+channel.ChannelType+"]", "channel")
		}()
	}

	return conv.ID, nil
}

func (s *AgentChannelService) updateLastSeen(channelID, senderID string) {
	if s.docDB == nil {
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	filter := map[string]interface{}{
		"agent_channel_id": channelID,
		"channel_user_id":  senderID,
	}
	_ = s.docDB.UpdateOne(context.Background(), models.CollectionChannelUserMappings, filter,
		map[string]interface{}{"$set": map[string]interface{}{"last_seen_on": now}})
	_ = s.docDB.UpdateOne(context.Background(), models.CollectionAgentChannelSessions, filter,
		map[string]interface{}{"$set": map[string]interface{}{"last_message_on": now}})
}

// incrementSessionMessageCount atomically increments the message counter for a channel session.
func (s *AgentChannelService) incrementSessionMessageCount(channelID, senderID string) {
	if s.docDB == nil {
		return
	}
	filter := map[string]interface{}{
		"agent_channel_id": channelID,
		"channel_user_id":  senderID,
	}
	_ = s.docDB.UpdateOne(context.Background(), models.CollectionAgentChannelSessions, filter,
		map[string]interface{}{"$inc": map[string]interface{}{"message_count": 1}})
}

// resolveSecret resolves secret references in channel config values.
// - "$env:MY_VAR"   → os.Getenv("MY_VAR")
// - "$vault:path"   → Vault HTTP API GET /v1/{path} → data.data.value
// - "plain text"    → returned as-is
func (s *AgentChannelService) resolveSecret(value string) string {
	if strings.HasPrefix(value, "$env:") {
		return os.Getenv(strings.TrimPrefix(value, "$env:"))
	}
	if strings.HasPrefix(value, "$vault:") {
		vaultPath := strings.TrimPrefix(value, "$vault:")
		vaultAddr := os.Getenv("VAULT_ADDR")
		vaultToken := os.Getenv("VAULT_TOKEN")
		if vaultAddr == "" || vaultToken == "" {
			s.iLog.Error("resolveSecret: VAULT_ADDR or VAULT_TOKEN not set")
			return ""
		}
		url := fmt.Sprintf("%s/v1/%s", vaultAddr, vaultPath)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("resolveSecret: failed to build vault request: %v", err))
			return ""
		}
		req.Header.Set("X-Vault-Token", vaultToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			s.iLog.Error(fmt.Sprintf("resolveSecret: vault request failed for path %s", vaultPath))
			return ""
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		// Parse "data.data.value" from Vault KV v2 response
		// Simple string extraction to avoid pulling in encoding/json dependency here
		// (this package already imports fmt, strings, os — keep it minimal)
		bodyStr := string(body)
		const marker = `"value":"`
		idx := strings.Index(bodyStr, marker)
		if idx < 0 {
			return ""
		}
		start := idx + len(marker)
		end := strings.Index(bodyStr[start:], `"`)
		if end < 0 {
			return ""
		}
		return bodyStr[start : start+end]
	}
	return value
}

// ResolveChannelConfig returns a copy of channel_config with all secrets resolved.
func (s *AgentChannelService) ResolveChannelConfig(ch *models.AgentChannelDoc) map[string]string {
	resolved := make(map[string]string, len(ch.ChannelConfig))
	for k, v := range ch.ChannelConfig {
		resolved[k] = s.resolveSecret(v)
	}
	return resolved
}

// DeactivateSession marks the MongoDB AgentChannelSession document for a sender
// as inactive, so the next inbound message creates a fresh conversation.
// Called by AgentMonitorService.ResetSession.
func (s *AgentChannelService) DeactivateSession(channelID, senderID string) error {
	if err := s.ensureDocDB(); err != nil {
		return err
	}
	filter := map[string]interface{}{
		"agent_channel_id": channelID,
		"channel_user_id":  senderID,
	}
	update := map[string]interface{}{"$set": map[string]interface{}{"active": false}}
	return s.docDB.UpdateOne(context.Background(), models.CollectionAgentChannelSessions, filter, update)
}

// ─── Map helpers ─────────────────────────────────────────────────────────────

func mapToChannelDoc(row map[string]interface{}) *models.AgentChannelDoc {
	doc := &models.AgentChannelDoc{}
	if v, ok := row["_id"].(string); ok {
		doc.ID = v
	}
	if v, ok := row["agent_id"].(string); ok {
		doc.AgentID = v
	}
	if v, ok := row["agent_name"].(string); ok {
		doc.AgentName = v
	}
	if v, ok := row["channel_type"].(string); ok {
		doc.ChannelType = v
	}
	if v, ok := row["display_name"].(string); ok {
		doc.DisplayName = v
	}
	if v, ok := row["welcome_message"].(string); ok {
		doc.WelcomeMessage = v
	}
	if v, ok := row["enabled"].(bool); ok {
		doc.Enabled = v
	}
	if v, ok := row["active"].(bool); ok {
		doc.Active = v
	}
	if v, ok := row["createdby"].(string); ok {
		doc.CreatedBy = v
	}
	if v, ok := row["createdon"].(string); ok {
		doc.CreatedOn = v
	}
	if v, ok := row["modifiedby"].(string); ok {
		doc.ModifiedBy = v
	}
	if v, ok := row["modifiedon"].(string); ok {
		doc.ModifiedOn = v
	}
	if v, ok := row["channel_config"]; ok {
		doc.ChannelConfig = toStringMap(v)
	}
	if v, ok := row["coordinator_mode"].(bool); ok {
		doc.CoordinatorMode = v
	}
	if v, ok := row["coordinator_prompt"].(string); ok {
		doc.CoordinatorPrompt = v
	}
	if v, ok := row["security_policy"]; ok {
		if policyMap, ok := v.(map[string]interface{}); ok {
			doc.SecurityPolicy = mapToSecurityPolicy(policyMap)
		}
	}
	return doc
}

func mapToSecurityPolicy(m map[string]interface{}) models.AgentChannelSecurityPolicy {
	p := models.AgentChannelSecurityPolicy{}
	if v, ok := m["require_user_auth"].(bool); ok {
		p.RequireUserAuth = v
	}
	if v, ok := m["max_iterations"]; ok {
		p.MaxIterations = toInt(v)
	}
	if v, ok := m["max_response_length"]; ok {
		p.MaxResponseLength = toInt(v)
	}
	if v, ok := m["rate_limit_per_hour"]; ok {
		p.RateLimitPerHour = toInt(v)
	}
	if v, ok := m["system_prompt_override"].(string); ok {
		p.SystemPromptOverride = v
	}
	if v, ok := m["allowed_tools"]; ok {
		p.AllowedTools = toStringSlice(v)
	}
	if v, ok := m["blocked_tools"]; ok {
		p.BlockedTools = toStringSlice(v)
	}
	return p
}

func mapToUserMappingDoc(row map[string]interface{}) *models.ChannelUserMappingDoc {
	doc := &models.ChannelUserMappingDoc{}
	if v, ok := row["_id"].(string); ok {
		doc.ID = v
	}
	if v, ok := row["agent_channel_id"].(string); ok {
		doc.AgentChannelID = v
	}
	if v, ok := row["channel_type"].(string); ok {
		doc.ChannelType = v
	}
	if v, ok := row["channel_user_id"].(string); ok {
		doc.ChannelUserID = v
	}
	if v, ok := row["channel_username"].(string); ok {
		doc.ChannelUsername = v
	}
	if v, ok := row["iac_user_id"].(string); ok {
		doc.IACUserID = v
	}
	if v, ok := row["iac_username"].(string); ok {
		doc.IACUsername = v
	}
	if v, ok := row["authorized"].(bool); ok {
		doc.Authorized = v
	}
	if v, ok := row["authorized_by"].(string); ok {
		doc.AuthorizedBy = v
	}
	if v, ok := row["authorized_on"].(string); ok {
		doc.AuthorizedOn = v
	}
	if v, ok := row["last_seen_on"].(string); ok {
		doc.LastSeenOn = v
	}
	if v, ok := row["active"].(bool); ok {
		doc.Active = v
	}
	if v, ok := row["createdon"].(string); ok {
		doc.CreatedOn = v
	}
	return doc
}

func toInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int32:
		return int(n)
	case int64:
		return int(n)
	case float64:
		return int(n)
	case float32:
		return int(n)
	}
	return 0
}
