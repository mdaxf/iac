package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac-signalr/signalr"
	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/signalrserver"
	openai "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

// AgentRunRequest is the input to RunAsync / RunFromJob
type AgentRunRequest struct {
	AgentDefinitionID string
	AgentScheduleID   string // empty for on-demand runs
	ConversationID    string // existing conversation ID (pre-created by caller)
	TaskPrompt        string
	RequestedBy       string
}

// runState tracks an in-flight agent run for abort support and dashboard status
type runState struct {
	cancel            context.CancelFunc
	agentDefinitionID string
	startedAt         time.Time
	sandbox           *AgentSandbox // nil when sandbox manager not available
}

// AgentRunnerService executes agent tasks using a ReAct-style LLM tool-call loop
type AgentRunnerService struct {
	db           *gorm.DB
	sqlDB        *sql.DB
	defService   *AgentDefinitionService
	chatService  *AgentChatService
	toolRegistry *AgentToolRegistryService
	mcpService   *MCPServerService
	memorySvc    *AgentMemoryService
	sandboxMgr   *AgentSandboxManager
	mu           sync.RWMutex
	activeRuns   map[string]*runState // runID -> state
	iLog         logger.Log
}

var (
	globalAgentRunnerService *AgentRunnerService
)

func GetGlobalAgentRunnerService() *AgentRunnerService {
	return globalAgentRunnerService
}

func SetGlobalAgentRunnerService(s *AgentRunnerService) {
	globalAgentRunnerService = s
}

func NewAgentRunnerService(
	db *gorm.DB,
	sqlDB *sql.DB,
	defService *AgentDefinitionService,
	chatService *AgentChatService,
	toolRegistry *AgentToolRegistryService,
	mcpService *MCPServerService,
	memorySvc *AgentMemoryService,
	sandboxMgr *AgentSandboxManager,
) *AgentRunnerService {
	return &AgentRunnerService{
		db:           db,
		sqlDB:        sqlDB,
		defService:   defService,
		chatService:  chatService,
		toolRegistry: toolRegistry,
		mcpService:   mcpService,
		memorySvc:    memorySvc,
		sandboxMgr:   sandboxMgr,
		activeRuns:   make(map[string]*runState),
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "AgentRunnerService",
		},
	}
}

// GetActiveRunsStatus returns a map of agentDefinitionID -> "running" for all in-flight runs.
func (s *AgentRunnerService) GetActiveRunsStatus() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]string, len(s.activeRuns))
	for _, state := range s.activeRuns {
		if state.agentDefinitionID != "" {
			result[state.agentDefinitionID] = "running"
		}
	}
	return result
}

// RunAsync starts an agent run in a goroutine and returns the run ID immediately
func (s *AgentRunnerService) RunAsync(req AgentRunRequest) (string, error) {
	runID := uuid.New().String()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	// Register sandbox before goroutine starts so callers can send messages immediately
	var sb *AgentSandbox
	if s.sandboxMgr != nil {
		sb = s.sandboxMgr.Register(runID, req.AgentDefinitionID)
	}

	s.mu.Lock()
	s.activeRuns[runID] = &runState{
		cancel:            cancel,
		agentDefinitionID: req.AgentDefinitionID,
		startedAt:         time.Now(),
		sandbox:           sb,
	}
	s.mu.Unlock()

	s.broadcastAgentStatus(req.AgentDefinitionID, "running")

	go func() {
		defer func() {
			cancel()
			if s.sandboxMgr != nil {
				s.sandboxMgr.Unregister(runID)
			}
			s.mu.Lock()
			delete(s.activeRuns, runID)
			s.mu.Unlock()
		}()
		if err := s.run(ctx, runID, req, sb); err != nil {
			s.iLog.Error(fmt.Sprintf("Agent run %s failed: %v", runID, err))
			s.broadcastError(req.ConversationID, err.Error())
			s.broadcastAgentStatus(req.AgentDefinitionID, "error")
		}
	}()

	return runID, nil
}

// AbortRun cancels an in-flight agent run
func (s *AgentRunnerService) AbortRun(runID string) {
	s.mu.RLock()
	state, ok := s.activeRuns[runID]
	s.mu.RUnlock()
	if ok {
		state.cancel()
	}
}

// RunSync executes an agent synchronously and returns the final response text.
// It is used by the workflow engine (AIAgent node) to trigger a real agent run
// inline and wait for its result without SignalR coordination.
//
// Resolution order: agentID (exact UUID) → agentName (case-insensitive lookup).
// timeoutSecs overrides the context deadline when > 0.
func (s *AgentRunnerService) RunSync(ctx context.Context, agentID, agentName, prompt string, timeoutSecs int) (string, error) {
	// Resolve agentID from name when only name is provided
	if agentID == "" && agentName != "" {
		agents, err := s.defService.ListAgents()
		if err != nil {
			return "", fmt.Errorf("failed to list agents: %w", err)
		}
		for _, a := range agents {
			if strings.EqualFold(a.Name, agentName) {
				agentID = a.ID
				break
			}
		}
		if agentID == "" {
			return "", fmt.Errorf("agent '%s' not found", agentName)
		}
	}
	if agentID == "" {
		return "", fmt.Errorf("agentID and agentName are both empty")
	}

	// Apply caller-supplied timeout
	if timeoutSecs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
		defer cancel()
	}

	// Create a temporary conversation for this workflow run
	conv, err := s.chatService.CreateAgentConversation(agentID, "workflow", "workflow-run-"+uuid.New().String()[:8], "workflow")
	if err != nil {
		return "", fmt.Errorf("failed to create conversation: %w", err)
	}
	convID := conv.ID

	runID := uuid.New().String()
	s.iLog.Info(fmt.Sprintf("RunSync start — agent=%s conv=%s runID=%s", agentID, convID, runID))

	// Run synchronously (blocks until the ReAct loop completes)
	if err := s.run(ctx, runID, AgentRunRequest{
		AgentDefinitionID: agentID,
		ConversationID:    convID,
		TaskPrompt:        prompt,
		RequestedBy:       "workflow",
	}, nil); err != nil {
		return "", err
	}

	// Retrieve the final assistant message saved by run()
	history, err := s.chatService.GetMessageHistory(convID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve response: %w", err)
	}
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].MessageType == models.MessageTypeAssistant {
			return history[i].Text, nil
		}
	}
	return "", fmt.Errorf("no assistant response found for run %s", runID)
}

// RunSyncInConversation executes an agent synchronously inside an EXISTING conversation.
// Unlike RunSync (which always creates a fresh conversation), this method loads the full
// message history of the supplied conversationID so the agent has complete context of the
// prior exchange — enabling true multi-turn chat via external channels.
//
// Used by AgentChannelService.HandleInbound so each inbound message continues the same
// conversation thread that was started when the channel user first made contact.
func (s *AgentRunnerService) RunSyncInConversation(ctx context.Context, conversationID, agentID, agentName, prompt string, timeoutSecs int) (string, error) {
	// Resolve agentID from name when only name is provided
	if agentID == "" && agentName != "" {
		agents, err := s.defService.ListAgents()
		if err != nil {
			return "", fmt.Errorf("failed to list agents: %w", err)
		}
		for _, a := range agents {
			if strings.EqualFold(a.Name, agentName) {
				agentID = a.ID
				break
			}
		}
		if agentID == "" {
			return "", fmt.Errorf("agent '%s' not found", agentName)
		}
	}
	if agentID == "" {
		return "", fmt.Errorf("agentID and agentName are both empty")
	}
	if conversationID == "" {
		return "", fmt.Errorf("conversationID is required for RunSyncInConversation")
	}

	// Apply caller-supplied timeout
	if timeoutSecs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
		defer cancel()
	}

	runID := uuid.New().String()
	s.iLog.Info(fmt.Sprintf("RunSyncInConversation start — agent=%s conv=%s runID=%s", agentID, conversationID, runID))

	// Run synchronously using the caller-supplied conversationID.
	// run() will load the full history of this conversation so the agent has prior context.
	if err := s.run(ctx, runID, AgentRunRequest{
		AgentDefinitionID: agentID,
		ConversationID:    conversationID,
		TaskPrompt:        prompt,
		RequestedBy:       "channel",
	}, nil); err != nil {
		return "", err
	}

	// Retrieve the last assistant message appended during this run
	history, err := s.chatService.GetMessageHistory(conversationID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve response: %w", err)
	}
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].MessageType == models.MessageTypeAssistant {
			return history[i].Text, nil
		}
	}
	return "", fmt.Errorf("no assistant response found for run %s", runID)
}

// RunFromJob is the entry point called by the job system handler
func (s *AgentRunnerService) RunFromJob(ctx context.Context, payload models.AgentJobPayload) error {
	runID := uuid.New().String()
	var sb *AgentSandbox
	if s.sandboxMgr != nil {
		sb = s.sandboxMgr.Register(runID, payload.AgentDefinitionID)
		defer s.sandboxMgr.Unregister(runID)
	}
	return s.run(ctx, runID, AgentRunRequest{
		AgentDefinitionID: payload.AgentDefinitionID,
		AgentScheduleID:   payload.AgentScheduleID,
		ConversationID:    payload.ConversationID,
		TaskPrompt:        payload.TaskPrompt,
		RequestedBy:       payload.RequestedBy,
	}, sb)
}

// run is the core LLM tool-call loop
func (s *AgentRunnerService) run(ctx context.Context, runID string, req AgentRunRequest, sandbox *AgentSandbox) error {
	s.iLog.Info(fmt.Sprintf("Agent run %s START - agent=%s conv=%s", runID, req.AgentDefinitionID, req.ConversationID))
	startTime := time.Now()

	// Load agent definition
	agent, err := s.defService.GetAgent(req.AgentDefinitionID)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	// Inject per-agent notification config into context so notification tools
	// can resolve credentials without requiring the LLM to supply them explicitly.
	if len(agent.NotificationConfig) > 0 {
		ctx = WithAgentNotifConfig(ctx, map[string]string(agent.NotificationConfig))
	}

	// Build OpenAI client from AI config
	oaiClient, modelName, err := s.buildOpenAIClient(agent)
	if err != nil {
		return fmt.Errorf("failed to build OpenAI client: %w", err)
	}

	// Get tool definitions from the agent's assigned skills.
	// GetToolDefinitionsForSkills merges any skill Content (schema/instructions stored in
	// MongoDB) directly into the LLM-visible tool description.
	tools := s.toolRegistry.GetToolDefinitionsForSkills(agent.Skills)

	// Inject MCP tools from assigned MCP servers
	mcpToolServerMap := make(map[string]models.MCPServer) // toolName -> server
	if s.mcpService != nil && len(agent.MCPServers) > 0 {
		for _, srv := range agent.MCPServers {
			if !srv.Enabled || !srv.Active {
				continue
			}
			mcpDefs, fetchErr := s.mcpService.FetchTools(srv)
			if fetchErr != nil {
				s.iLog.Error(fmt.Sprintf("Failed to fetch MCP tools from server %s: %v", srv.Name, fetchErr))
				continue
			}
			for _, td := range mcpDefs {
				tools = append(tools, td)
				mcpToolServerMap[td.Function.Name] = srv
			}
		}
	}

	// Inject memory tools
	tools = append(tools, s.memoryToolDefinitions()...)
	// Inject send_to_agent tool (agent-to-agent collaboration)
	tools = append(tools, s.sandboxToolDefinitions()...)
	// Inject notification tools (send_email, send_telegram, etc.) — always available
	// regardless of which skills are assigned, so agents can send results without
	// requiring the notification skill to be explicitly configured.
	tools = append(tools, s.toolRegistry.GetToolDefinitions(notificationToolNames())...)
	// Inject ui_render, ui_save_view, ui_compose_page — always available so agents can
	// produce rich interactive UI panels and persist them as reusable views/pages.
	tools = append(tools, s.toolRegistry.GetToolDefinitions([]string{"ui_render", "ui_save_view", "ui_compose_page"})...)

	// Determine iteration cap from agent definition (may be overridden by channel policy below).
	maxIter := agent.MaxIterations
	if maxIter <= 0 {
		maxIter = 10
	}

	// Apply channel security policy if this run originates from a channel inbound message.
	// The policy is injected into the context by AgentChannelService.HandleInbound().
	if policy, ok := ctx.Value(models.AgentChannelPolicyKey{}).(models.AgentChannelSecurityPolicy); ok {
		tools = filterTools(tools, policy.AllowedTools, policy.BlockedTools)
		if policy.MaxIterations > 0 {
			maxIter = policy.MaxIterations
		}
	}

	oaiTools := toOpenAITools(tools)

	// Load conversation history
	history, err := s.chatService.GetMessageHistory(req.ConversationID)
	if err != nil {
		history = []models.ChatMessage{}
	}

	// Save user message
	_, _ = s.chatService.SaveUserMessage(req.ConversationID, req.TaskPrompt, req.RequestedBy)

	// Build system prompt — prepend coordinator prompt, channel policy override,
	// and L0 memory index (in that order, outermost context first).
	systemPrompt := agent.SystemPrompt
	if coordinatorPrompt, ok := ctx.Value(models.AgentCoordinatorPromptKey{}).(string); ok && coordinatorPrompt != "" {
		systemPrompt = coordinatorPrompt + "\n\n" + systemPrompt
	}
	if policy, ok := ctx.Value(models.AgentChannelPolicyKey{}).(models.AgentChannelSecurityPolicy); ok {
		if policy.SystemPromptOverride != "" {
			systemPrompt = policy.SystemPromptOverride + "\n\n" + systemPrompt
		}
	}
	if s.memorySvc != nil {
		if memCtx := s.buildMemoryContext(req.AgentDefinitionID); memCtx != "" {
			systemPrompt = systemPrompt + "\n\n" + memCtx
		}
		// Inject similar past-run summaries (auto-memory pre-run retrieval).
		if simCtx := s.buildSimilarMemoryContext(req.AgentDefinitionID, req.TaskPrompt); simCtx != "" {
			systemPrompt = systemPrompt + "\n\n" + simCtx
		}
	}

	// Build initial messages
	messages := s.buildMessages(systemPrompt, history, req.TaskPrompt)

	var finalText string
	iterationsUsed := 0
	toolsUsedSet := make(map[string]struct{})
	// richToolResults collects outputs of tools that produce renderable content
	// (ui_render → ui_json, generate_html_report → html). Saved with the final message.
	var richToolResults []map[string]interface{}

	for iter := 0; iter < maxIter; iter++ {
		iterationsUsed = iter + 1

		// Check for context cancellation or pending inbox messages
		select {
		case <-ctx.Done():
			return fmt.Errorf("agent run aborted")
		case inboxMsg, ok := <-s.drainInbox(sandbox):
			if ok {
				// Inject inbox message as a user turn and continue the loop
				inboxContent := fmt.Sprintf("[INBOX] topic: %s\npayload: %s", inboxMsg.Topic, inboxMsg.Payload)
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: inboxContent,
				})
			}
			// fall through to LLM call
		default:
			// no inbox message, proceed normally
		}

		resp, err := s.callLLM(ctx, oaiClient, modelName, messages, oaiTools, agent.Temperature)
		if err != nil {
			return fmt.Errorf("LLM call failed: %w", err)
		}

		choice := resp.Choices[0]
		msg := choice.Message

		// Append assistant message to loop
		messages = append(messages, msg)

		if len(msg.ToolCalls) == 0 {
			// No tool calls — final response
			finalText = msg.Content
			s.broadcastChunk(req.ConversationID, msg.Content)
			break
		}

		// Process tool calls
		for _, tc := range msg.ToolCalls {
			toolName := tc.Function.Name
			toolArgs := tc.Function.Arguments
			toolsUsedSet[toolName] = struct{}{}

			s.broadcastToolUse(req.ConversationID, toolName, toolArgs)

			var result string
			var toolErr error

			switch toolName {
			case "memory_search":
				result, toolErr = s.execMemorySearch(req.AgentDefinitionID, toolArgs)
			case "memory_read":
				result, toolErr = s.execMemoryRead(toolArgs)
			case "memory_save":
				result, toolErr = s.execMemorySave(req.AgentDefinitionID, toolArgs)
			case "send_to_agent":
				result, toolErr = s.execSendToAgent(runID, req.AgentDefinitionID, toolArgs)
			default:
				// Check if this is an MCP tool
				if mcpSrv, isMCP := mcpToolServerMap[toolName]; isMCP && s.mcpService != nil {
					result, toolErr = s.mcpService.CallTool(ctx, mcpSrv, toolName, toolArgs)
				} else {
					result, toolErr = s.toolRegistry.ExecuteTool(ctx, toolName, toolArgs)
				}
			}
			if toolErr != nil {
				result = fmt.Sprintf("error: %v", toolErr)
			}

			s.broadcastToolResult(req.ConversationID, toolName, result)

			// Collect rich renderable outputs for history persistence
			if toolErr == nil {
				var parsed map[string]interface{}
				if json.Unmarshal([]byte(result), &parsed) == nil {
					_, hasUI := parsed["ui_json"]
					_, hasHTML := parsed["html"]
					if hasUI || hasHTML {
						richToolResults = append(richToolResults, map[string]interface{}{
							"tool":   toolName,
							"result": parsed,
						})
					}
				}
			}

			// Append tool result for next iteration
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	// Build tools used list
	toolsUsed := make([]string, 0, len(toolsUsedSet))
	for t := range toolsUsedSet {
		toolsUsed = append(toolsUsed, t)
	}

	// Save final assistant message together with any rich tool results (ui_render, html_report)
	if finalText != "" {
		_, _ = s.chatService.SaveAgentMessageWithTools(req.ConversationID, finalText, richToolResults, req.RequestedBy)
	}

	// Auto-save run to memory (L2→L1→L0) in the background — never blocks the run.
	if s.memorySvc != nil && finalText != "" {
		historyJSON, _ := json.Marshal(messages)
		capturedRunID := runID
		capturedTask := req.TaskPrompt
		capturedFinal := finalText
		capturedAgent := req.AgentDefinitionID
		capturedConv := req.ConversationID
		capturedHistory := string(historyJSON)
		go func() {
			if err := s.memorySvc.AutoSaveRunMemory(
				context.Background(),
				capturedAgent, capturedRunID, capturedConv,
				capturedTask, capturedFinal, capturedHistory,
			); err != nil {
				s.iLog.Warn(fmt.Sprintf("AutoSaveRunMemory failed: %v", err))
			}
		}()
	}

	// Update schedule last run status if applicable
	if req.AgentScheduleID != "" {
		_ = s.updateScheduleLastRun(req.AgentScheduleID, "completed", finalText)
	}

	duration := time.Since(startTime)
	s.broadcastDone(req.ConversationID, finalText, duration, iterationsUsed, toolsUsed)
	s.broadcastAgentStatus(req.AgentDefinitionID, "idle")
	s.iLog.Info(fmt.Sprintf("Agent run %s COMPLETE in %v, %d iters, tools=%v", runID, duration, iterationsUsed, toolsUsed))
	return nil
}

// drainInbox returns a channel that yields at most one inbox message without blocking.
// Returns a nil channel when sandbox is nil (so the select case never fires).
func (s *AgentRunnerService) drainInbox(sandbox *AgentSandbox) <-chan SandboxMessage {
	if sandbox == nil {
		return nil
	}
	return sandbox.Inbox
}

// buildMemoryContext loads L0 memory items and formats them as a context block.
func (s *AgentRunnerService) buildMemoryContext(agentID string) string {
	if s.memorySvc == nil {
		return ""
	}
	items, err := s.memorySvc.ListMemory(agentID, ListMemoryOptions{Layer: "L0"})
	if err != nil || len(items) == 0 {
		items2, err2 := s.memorySvc.ListMemory(agentID, ListMemoryOptions{Layer: "L1"})
		if err2 != nil || len(items2) == 0 {
			return ""
		}
		items = items2
	}
	var sb strings.Builder
	sb.WriteString("## Memory Index\nThe following memory items are available (use memory_read to load full content):\n")
	for _, m := range items {
		sb.WriteString(fmt.Sprintf("- [%s] id=%s priority=%s | %s", m.Layer, m.ID, m.Priority, m.Title))
		if m.Summary != "" {
			sb.WriteString(": " + m.Summary)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// buildSimilarMemoryContext searches the auto-memory L0 index for past runs similar
// to taskPrompt and injects their L1 summaries into the system prompt when similarity
// is >= 80%.  At most 3 results are injected to keep the prompt concise.
func (s *AgentRunnerService) buildSimilarMemoryContext(agentID, taskPrompt string) string {
	if s.memorySvc == nil || taskPrompt == "" {
		return ""
	}
	results, err := s.memorySvc.FindSimilarMemory(agentID, taskPrompt, 0.8)
	if err != nil || len(results) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## Relevant Past Experience (Auto-retrieved)\n")
	sb.WriteString("The following summaries are from similar past tasks (similarity ≥ 80%). Use them to avoid re-doing work:\n")
	for i, r := range results {
		if i >= 3 {
			break
		}
		sb.WriteString(fmt.Sprintf("\n### Past Task %d (similarity: %.0f%%)\n", i+1, r.Score*100))
		sb.WriteString(fmt.Sprintf("**Topic:** %s\n", r.L1.Title))
		sb.WriteString(fmt.Sprintf("**Summary:** %s\n", r.L1.Content))
		// Link to full history via L2 ID stored in L1 tags.
		for _, tag := range r.L1.Tags {
			if strings.HasPrefix(tag, "l2:") {
				sb.WriteString(fmt.Sprintf("**Full history:** use `memory_read` with id=%s\n", strings.TrimPrefix(tag, "l2:")))
				break
			}
		}
	}
	return sb.String()
}

// ─── Memory tool definitions ──────────────────────────────────────────────────

func (s *AgentRunnerService) memoryToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Type: "function",
			Function: ToolFunctionDef{
				Name:        "memory_search",
				Description: "Search agent memory by title, summary, or tags. Returns matching memory items with IDs for further reading.",
				Parameters: ToolParameterSchema{
					Type: "object",
					Properties: map[string]ToolPropertySchema{
						"query": {Type: "string", Description: "Search term to match against title, summary, and tags"},
						"layer": {Type: "string", Description: "Memory layer to search: L0, L1, L2 or empty for all"},
					},
					Required: []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunctionDef{
				Name:        "memory_read",
				Description: "Load the full content of a memory item by ID. Increments the access counter.",
				Parameters: ToolParameterSchema{
					Type: "object",
					Properties: map[string]ToolPropertySchema{
						"id": {Type: "string", Description: "Memory item ID returned by memory_search"},
					},
					Required: []string{"id"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunctionDef{
				Name:        "memory_save",
				Description: "Save a new memory item for future use. Use L0 for quick index entries, L1 for summaries, L2 for full content.",
				Parameters: ToolParameterSchema{
					Type: "object",
					Properties: map[string]ToolPropertySchema{
						"title":          {Type: "string", Description: "Short descriptive title"},
						"summary":        {Type: "string", Description: "Brief summary (L0/L1)"},
						"content":        {Type: "string", Description: "Full content (L2)"},
						"layer":          {Type: "string", Description: "L0 | L1 | L2"},
						"priority":       {Type: "string", Description: "P0 (critical) | P1 (normal) | P2 (low)"},
						"retention_type": {Type: "string", Description: "permanent | interval | temporary"},
						"retention_days": {Type: "string", Description: "Number of days for interval retention (optional)"},
					},
					Required: []string{"title", "layer", "priority"},
				},
			},
		},
	}
}

// notificationToolNames returns the names of notification tools that are injected
// into every agent run regardless of skill assignment.
func notificationToolNames() []string {
	return []string{"send_email", "send_telegram", "send_webhook", "send_slack"}
}

// filterTools applies an allow/block list from an AgentChannelSecurityPolicy to a tool list.
// - If allowedTools is non-empty, only those tools are kept (whitelist).
// - blockedTools are always removed (blacklist takes priority over whitelist).
// Memory tools (memory_search, memory_read, memory_save, send_to_agent) are never filtered
// because they are internal, non-destructive capabilities always needed for proper operation.
func filterTools(tools []ToolDefinition, allowed, blocked []string) []ToolDefinition {
	// Build lookup maps
	allowMap := make(map[string]bool, len(allowed))
	for _, t := range allowed {
		allowMap[t] = true
	}
	blockMap := make(map[string]bool, len(blocked))
	for _, t := range blocked {
		blockMap[t] = true
	}

	// Always-exempt tool names (internal memory + agent comms)
	exempt := map[string]bool{
		"memory_search": true,
		"memory_read":   true,
		"memory_save":   true,
		"send_to_agent": true,
	}

	filtered := make([]ToolDefinition, 0, len(tools))
	for _, tool := range tools {
		name := tool.Function.Name
		// Never filter exempt tools
		if exempt[name] {
			filtered = append(filtered, tool)
			continue
		}
		// Block list takes priority
		if blockMap[name] {
			continue
		}
		// Apply whitelist if set
		if len(allowMap) > 0 && !allowMap[name] {
			continue
		}
		filtered = append(filtered, tool)
	}
	return filtered
}

// ─── Sandbox tool definitions ─────────────────────────────────────────────────

func (s *AgentRunnerService) sandboxToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Type: "function",
			Function: ToolFunctionDef{
				Name:        "send_to_agent",
				Description: "Send a message to another agent's inbox for asynchronous collaboration.",
				Parameters: ToolParameterSchema{
					Type: "object",
					Properties: map[string]ToolPropertySchema{
						"agent_id": {Type: "string", Description: "Target agent definition ID"},
						"topic":    {Type: "string", Description: "Message topic / intent"},
						"payload":  {Type: "string", Description: "Message payload (JSON or text)"},
					},
					Required: []string{"agent_id", "topic", "payload"},
				},
			},
		},
	}
}

// ─── Tool executors ───────────────────────────────────────────────────────────

func (s *AgentRunnerService) execMemorySearch(agentID, argsJSON string) (string, error) {
	var args struct {
		Query string `json:"query"`
		Layer string `json:"layer"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	if s.memorySvc == nil {
		return "[]", nil
	}
	items, err := s.memorySvc.SearchMemory(agentID, args.Query, args.Layer)
	if err != nil {
		return "", err
	}
	out, _ := json.Marshal(items)
	return string(out), nil
}

func (s *AgentRunnerService) execMemoryRead(argsJSON string) (string, error) {
	var args struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	if s.memorySvc == nil {
		return "{}", nil
	}
	mem, err := s.memorySvc.ReadMemoryItem(args.ID)
	if err != nil {
		return "", err
	}
	out, _ := json.Marshal(mem)
	return string(out), nil
}

func (s *AgentRunnerService) execMemorySave(agentID, argsJSON string) (string, error) {
	var args struct {
		Title         string          `json:"title"`
		Summary       string          `json:"summary"`
		Content       string          `json:"content"`
		Layer         string          `json:"layer"`
		Priority      string          `json:"priority"`
		RetentionType string          `json:"retention_type"`
		RetentionDays json.RawMessage `json:"retention_days"` // LLMs send this as a number or a string
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	if s.memorySvc == nil {
		return `{"error":"memory service unavailable"}`, nil
	}
	var retDays *int
	if len(args.RetentionDays) > 0 && string(args.RetentionDays) != "null" {
		// Try JSON number first (e.g. 30), then JSON string (e.g. "30")
		var d int
		if err := json.Unmarshal(args.RetentionDays, &d); err == nil {
			retDays = &d
		} else {
			var s string
			if err := json.Unmarshal(args.RetentionDays, &s); err == nil {
				if _, scanErr := fmt.Sscanf(s, "%d", &d); scanErr == nil {
					retDays = &d
				}
			}
		}
	}
	mem, err := s.memorySvc.SaveMemoryItem(agentID, args.Title, args.Summary, args.Content, args.Layer, args.Priority, args.RetentionType, retDays)
	if err != nil {
		return "", err
	}
	out, _ := json.Marshal(map[string]string{"id": mem.ID, "status": "saved"})
	return string(out), nil
}

func (s *AgentRunnerService) execSendToAgent(fromRunID, fromAgentID, argsJSON string) (string, error) {
	var args struct {
		AgentID string `json:"agent_id"`
		Topic   string `json:"topic"`
		Payload string `json:"payload"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	if s.sandboxMgr == nil {
		return `{"error":"sandbox manager unavailable"}`, nil
	}
	err := s.sandboxMgr.SendToAgent(args.AgentID, SandboxMessage{
		FromAgentID: fromAgentID,
		FromRunID:   fromRunID,
		Topic:       args.Topic,
		Payload:     args.Payload,
	})
	if err != nil {
		return fmt.Sprintf(`{"error":"%s"}`, err.Error()), nil
	}
	return `{"status":"sent"}`, nil
}

// buildMessages constructs the OpenAI messages array
func (s *AgentRunnerService) buildMessages(systemPrompt string, history []models.ChatMessage, userInput string) []openai.ChatCompletionMessage {
	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
	}
	for _, h := range history {
		role := openai.ChatMessageRoleUser
		if h.MessageType == models.MessageTypeAssistant {
			role = openai.ChatMessageRoleAssistant
		}
		msgs = append(msgs, openai.ChatCompletionMessage{Role: role, Content: h.Text})
	}
	msgs = append(msgs, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: userInput})
	return msgs
}

// buildOpenAIClient creates a go-openai client for the agent's configured vendor.
//
// Google Gemini and Ollama expose OpenAI-compatible endpoints, so they work with
// the same go-openai library.  The function normalises the base URL per vendor so
// the caller never has to worry about vendor-specific path differences.
//
//   google     → https://generativelanguage.googleapis.com/v1beta/openai
//                (Gemini OpenAI-compatible endpoint; key sent as Bearer token)
//   ollama     → http://localhost:11434/v1  (or configured base + /v1)
//   openai     → https://api.openai.com/v1  (default)
//   azure_openai → caller-supplied base URL
func (s *AgentRunnerService) buildOpenAIClient(agent *models.AgentDefinition) (*openai.Client, string, error) {
	aiCfg := config.GetAIConfig()
	if aiCfg == nil {
		return nil, "", fmt.Errorf("AI configuration not loaded")
	}

	vendorName := agent.ModelProvider
	if vendorName == "" {
		vendorName = "openai"
	}
	// Normalise known aliases
	if vendorName == "azure" {
		vendorName = "azure_openai"
	}

	vendorCfg, ok := aiCfg.AIVendors[vendorName]
	if !ok || !vendorCfg.Enabled {
		// Fall back to the first enabled OpenAI-compatible vendor
		for _, name := range []string{"openai", "azure_openai", "google", "ollama"} {
			if v, exists := aiCfg.AIVendors[name]; exists && v.Enabled && v.APIKey != "" {
				vendorCfg = v
				vendorName = name
				ok = true
				break
			}
		}
		if !ok {
			return nil, "", fmt.Errorf("vendor %q not found or disabled in aiconfig.json and no OpenAI-compatible fallback is available", agent.ModelProvider)
		}
	}

	if vendorCfg.APIKey == "" && vendorName != "ollama" {
		return nil, "", fmt.Errorf("API key for vendor %q is empty — set it in aiconfig.local.json (ai_vendors.%s.api_key)", vendorName, vendorName)
	}

	cfg := openai.DefaultConfig(vendorCfg.APIKey)

	// Set the correct base URL for each vendor.
	// go-openai appends /chat/completions (and similar) to BaseURL, so BaseURL
	// must be the root path of the v1-compatible API — no trailing slash.
	switch vendorName {
	case "google":
		// Google Gemini exposes an OpenAI-compatible API at /v1beta/openai/.
		// Reference: https://ai.google.dev/gemini-api/docs/openai
		// The API key is passed as a Bearer token (go-openai does this by default).
		base := strings.TrimRight(vendorCfg.APIBaseURL, "/")
		if base == "" || base == "https://generativelanguage.googleapis.com" {
			base = "https://generativelanguage.googleapis.com/v1beta/openai"
		}
		cfg.BaseURL = base
	case "ollama":
		// Ollama serves an OpenAI-compatible API at <base>/v1/.
		base := strings.TrimRight(vendorCfg.APIBaseURL, "/")
		if base == "" {
			base = "http://localhost:11434"
		}
		if !strings.HasSuffix(base, "/v1") && !strings.Contains(base, "/v1/") {
			base += "/v1"
		}
		cfg.BaseURL = base
	default:
		if vendorCfg.APIBaseURL != "" {
			cfg.BaseURL = strings.TrimRight(vendorCfg.APIBaseURL, "/")
		}
	}

	modelName := agent.ModelName
	if modelName == "" {
		switch vendorName {
		case "google":
			modelName = "gemini-2.0-flash"
		case "ollama":
			modelName = "llama3"
		default:
			modelName = "gpt-4o"
		}
	}

	s.iLog.Info(fmt.Sprintf("buildOpenAIClient: vendor=%s model=%s baseURL=%s", vendorName, modelName, cfg.BaseURL))
	return openai.NewClientWithConfig(cfg), modelName, nil
}

// callLLM calls the OpenAI/Gemini chat completion API.
// It normalises Google Gemini's non-standard array error format so callers
// always receive a meaningful error message.
func (s *AgentRunnerService) callLLM(
	ctx context.Context,
	client *openai.Client,
	model string,
	messages []openai.ChatCompletionMessage,
	tools []openai.Tool,
	temperature float64,
) (*openai.ChatCompletionResponse, error) {
	req := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: float32(temperature),
	}
	if len(tools) > 0 {
		req.Tools = tools
	}
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, extractLLMError(err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from LLM")
	}
	return &resp, nil
}

// extractLLMError converts go-openai errors into readable messages.
// Google Gemini returns errors as a JSON array [{"error":{...}}] which
// go-openai cannot parse (it expects {"error":{...}}), so we extract the
// real message from RequestError.Body manually.
func extractLLMError(err error) error {
	var reqErr *openai.RequestError
	if !errors.As(err, &reqErr) || len(reqErr.Body) == 0 {
		return err
	}

	// Try Google's array error format: [{"error":{"code":N,"message":"...","status":"..."}}]
	var googleErrs []struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	if jsonErr := json.Unmarshal(reqErr.Body, &googleErrs); jsonErr == nil && len(googleErrs) > 0 {
		e := googleErrs[0].Error
		if e.Message != "" {
			return fmt.Errorf("Google API error (HTTP %d, %s): %s", e.Code, e.Status, e.Message)
		}
	}

	// Fall back to showing the raw body so the real error is never hidden
	return fmt.Errorf("LLM error (HTTP %d): %s", reqErr.HTTPStatusCode, string(reqErr.Body))
}

// toOpenAITools converts ToolDefinition slice to openai.Tool slice
func toOpenAITools(defs []ToolDefinition) []openai.Tool {
	tools := make([]openai.Tool, 0, len(defs))
	for _, d := range defs {
		props := make(map[string]interface{})
		for k, v := range d.Function.Parameters.Properties {
			props[k] = map[string]string{
				"type":        v.Type,
				"description": v.Description,
			}
		}
		params := map[string]interface{}{
			"type":       "object",
			"properties": props,
		}
		if len(d.Function.Parameters.Required) > 0 {
			params["required"] = d.Function.Parameters.Required
		}

		paramsJSON, _ := json.Marshal(params)
		tools = append(tools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        d.Function.Name,
				Description: d.Function.Description,
				Parameters:  json.RawMessage(paramsJSON),
			},
		})
	}
	return tools
}

// ---- SignalR broadcast helpers ----

func (s *AgentRunnerService) broadcast(topic, payload string) {
	client := com.IACMessageBusClient
	if client == nil {
		// Fall back to the embedded SignalR server when no external client is configured.
		if embSrv := signalrserver.GetGlobalSignalRServer(); embSrv != nil && embSrv.IsRunning() {
			s.iLog.Info(fmt.Sprintf("broadcast: using embedded SignalR server for topic=%s", topic))
			embSrv.BroadcastToHub(topic, payload)
			return
		}
		s.iLog.Info(fmt.Sprintf("broadcast: no SignalR transport available, skipping topic=%s", topic))
		return
	}
	if client.State() != signalr.ClientConnected {
		s.iLog.Info(fmt.Sprintf("broadcast: IACMessageBusClient not connected (state=%v), skipping topic=%s", client.State(), topic))
		return
	}
	s.iLog.Info(fmt.Sprintf("broadcast: sending topic=%s", topic))
	// Use "sendtoui" hub method so iac-signalr broadcasts only to the IAC_UI_MessageBus
	// group (frontend clients). Backend Go clients (iac-main itself) never join that group,
	// so they never receive these topic-specific messages — no "Unknown method" errors.
	<-client.Send("sendtoui", topic, payload, "")
}

func (s *AgentRunnerService) broadcastChunk(convID, chunk string) {
	msg, _ := json.Marshal(map[string]string{"conversationid": convID, "chunk": chunk})
	s.broadcast("agent.chat.chunk."+convID, string(msg))
}

func (s *AgentRunnerService) broadcastToolUse(convID, toolName, toolArgs string) {
	msg, _ := json.Marshal(map[string]string{"conversationid": convID, "tool": toolName, "args": toolArgs})
	s.broadcast("agent.tool.use."+convID, string(msg))
}

func (s *AgentRunnerService) broadcastToolResult(convID, toolName, result string) {
	msg, _ := json.Marshal(map[string]string{"conversationid": convID, "tool": toolName, "result": result})
	s.broadcast("agent.tool.result."+convID, string(msg))
}

func (s *AgentRunnerService) broadcastDone(convID, finalText string, duration time.Duration, iterations int, toolsUsed []string) {
	payload := map[string]interface{}{
		"conversationid": convID,
		"status":         "done",
		"text":           finalText,
		"duration_ms":    duration.Milliseconds(),
		"iterations":     iterations,
		"tools_used":     toolsUsed,
	}
	msg, _ := json.Marshal(payload)
	s.broadcast("agent.chat.done."+convID, string(msg))
}

func (s *AgentRunnerService) broadcastError(convID, errMsg string) {
	msg, _ := json.Marshal(map[string]string{"conversationid": convID, "error": errMsg})
	s.broadcast("agent.chat.error."+convID, string(msg))
}

func (s *AgentRunnerService) broadcastAgentStatus(agentDefinitionID, status string) {
	if agentDefinitionID == "" {
		return
	}
	msg, _ := json.Marshal(map[string]string{"agentid": agentDefinitionID, "status": status})
	s.broadcast("agent.status."+agentDefinitionID, string(msg))
}

// ---- Schedule management ----

// CreateScheduleRequest contains parameters for creating an agent schedule
type CreateScheduleRequest struct {
	AgentDefinitionID string `json:"agentdefinitionid"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	TaskPrompt        string `json:"taskprompt"`
	CronExpression    string `json:"cronexpression"`
	CreatedBy         string `json:"createdby"`
}

// CreateSchedule creates a new agent schedule
func (s *AgentRunnerService) CreateSchedule(req CreateScheduleRequest) (*models.AgentSchedule, error) {
	sched := &models.AgentSchedule{
		ID:                uuid.New().String(),
		AgentDefinitionID: req.AgentDefinitionID,
		Name:              req.Name,
		Description:       req.Description,
		TaskPrompt:        req.TaskPrompt,
		CronExpression:    req.CronExpression,
		Enabled:           true,
		Active:            true,
		CreatedBy:         req.CreatedBy,
		ModifiedBy:        req.CreatedBy,
	}
	if err := s.db.Create(sched).Error; err != nil {
		return nil, err
	}
	return sched, nil
}

// ListSchedules returns schedules for an agent
func (s *AgentRunnerService) ListSchedules(agentDefinitionID string) ([]models.AgentSchedule, error) {
	var schedules []models.AgentSchedule
	query := s.db.Where("active = ?", true)
	if agentDefinitionID != "" {
		query = query.Where("agent_definition_id = ?", agentDefinitionID)
	}
	err := query.Order("name").Find(&schedules).Error
	return schedules, err
}

// ToggleSchedule enables or disables a schedule
func (s *AgentRunnerService) ToggleSchedule(scheduleID string, enabled bool, modifiedBy string) error {
	return s.db.Model(&models.AgentSchedule{}).
		Where("id = ?", scheduleID).
		Updates(map[string]interface{}{
			"enabled":    enabled,
			"modifiedby": modifiedBy,
			"modifiedon": time.Now(),
		}).Error
}

// DeleteSchedule soft-deletes a schedule
func (s *AgentRunnerService) DeleteSchedule(scheduleID, deletedBy string) error {
	return s.db.Model(&models.AgentSchedule{}).
		Where("id = ?", scheduleID).
		Updates(map[string]interface{}{
			"active":     false,
			"modifiedby": deletedBy,
			"modifiedon": time.Now(),
		}).Error
}

func (s *AgentRunnerService) updateScheduleLastRun(scheduleID, status, output string) error {
	now := time.Now()
	return s.db.Model(&models.AgentSchedule{}).
		Where("id = ?", scheduleID).
		Updates(map[string]interface{}{
			"last_run_at":     now,
			"last_run_status": status,
			"last_run_output": output,
		}).Error
}
