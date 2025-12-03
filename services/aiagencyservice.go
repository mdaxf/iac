package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/codegen"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// AIAgencyService provides unified AI assistant functionality across all editors
// It acts as an AI agency that can route questions to specialized handlers
type AIAgencyService struct {
	DB                     *gorm.DB
	OpenAIKey              string
	OpenAIModel            string
	ChatService            *ChatService
	AIReportService        *AIReportService
	SchemaMetadataService  *SchemaMetadataService
	SchemaEmbeddingService *SchemaEmbeddingService
	iLog                   logger.Log
}

// NewAIAgencyService creates a new AI agency service
func NewAIAgencyService(
	db *gorm.DB,
	openAIKey, openAIModel string,
	chatService *ChatService,
	aiReportService *AIReportService,
	schemaService *SchemaMetadataService,
	embeddingService *SchemaEmbeddingService,
) *AIAgencyService {
	iLog := logger.Log{
		ModuleName:     logger.Framework,
		User:           "System",
		ControllerName: "AIAgencyService",
	}

	// Auto-migrate the conversation sessions table
	if err := db.AutoMigrate(&models.AIConversationSession{}); err != nil {
		iLog.Error(fmt.Sprintf("Failed to auto-migrate AIConversationSession table: %v", err))
	} else {
		iLog.Info("AIConversationSession table migrated successfully")
	}

	return &AIAgencyService{
		DB:                     db,
		OpenAIKey:              openAIKey,
		OpenAIModel:            openAIModel,
		ChatService:            chatService,
		AIReportService:        aiReportService,
		SchemaMetadataService:  schemaService,
		SchemaEmbeddingService: embeddingService,
		iLog:                   iLog,
	}
}

// ConversationContext holds context about the current conversation
type ConversationContext struct {
	SessionID      string                   `json:"session_id"`
	UserID         string                   `json:"user_id"`
	EditorType     string                   `json:"editor_type"`      // "bpm", "page", "view", "workflow", "whiteboard", "report", "general"
	DatabaseAlias  string                   `json:"database_alias"`
	EntityID       string                   `json:"entity_id"`        // ID of BPM/Page/View being edited
	EntityContext  map[string]interface{}   `json:"entity_context"`   // Current state of the entity
	PageContext    map[string]interface{}   `json:"page_context"`     // Current page context
	ConversationHistory []ConversationMessage `json:"conversation_history"`
	Metadata       map[string]interface{}   `json:"metadata"`
}

// ConversationMessage represents a single message in the conversation
type ConversationMessage struct {
	Role      string                 `json:"role"`      // "user", "assistant", "system"
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// AIAgencyRequest represents a request to the AI agency
type AIAgencyRequest struct {
	SessionID           string                   `json:"session_id"`
	UserID              string                   `json:"user_id"`
	Question            string                   `json:"question"`
	EditorType          string                   `json:"editor_type"`
	DatabaseAlias       string                   `json:"database_alias,omitempty"`
	EntityID            string                   `json:"entity_id,omitempty"`
	EntityContext       map[string]interface{}   `json:"entity_context,omitempty"`
	PageContext         map[string]interface{}   `json:"page_context,omitempty"`
	ConversationHistory []map[string]interface{} `json:"conversation_history,omitempty"`
	Options             map[string]interface{}   `json:"options,omitempty"`
}

// AIAgencyResponse represents a response from the AI agency
type AIAgencyResponse struct {
	SessionID      string                 `json:"session_id"`
	Answer         string                 `json:"answer"`
	ResponseType   string                 `json:"response_type"`   // "answer", "question", "code_generation", "clarification"
	IntentType     string                 `json:"intent_type"`     // "general", "bpm_generation", "query_generation", etc.
	Data           map[string]interface{} `json:"data,omitempty"`  // Generated code, SQL, etc.
	Questions      []ClarificationQuestion `json:"questions,omitempty"` // Follow-up questions if needed
	Confidence     float64                `json:"confidence"`
	RequiresAction bool                   `json:"requires_action"` // If user needs to approve/provide more info
	NextStep       string                 `json:"next_step,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ClarificationQuestion represents a question the AI needs answered
type ClarificationQuestion struct {
	Question string   `json:"question"`
	Type     string   `json:"type"`     // "text", "choice", "confirm"
	Options  []string `json:"options,omitempty"`
	Required bool     `json:"required"`
}

// TransactionHandler interface for specialized handlers
type TransactionHandler interface {
	CanHandle(ctx context.Context, intent string, editorType string) bool
	Handle(ctx context.Context, request *AIAgencyRequest, conversation *ConversationContext) (*AIAgencyResponse, error)
	GetName() string
}

// HandleRequest is the main entry point for AI agency requests
func (s *AIAgencyService) HandleRequest(ctx context.Context, request *AIAgencyRequest) (*AIAgencyResponse, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		s.iLog.PerformanceWithDuration("AIAgencyService.HandleRequest", elapsed)
	}()

	// Get or create conversation session
	conversation, err := s.getOrCreateSession(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation session: %w", err)
	}

	// Add user message to conversation history
	conversation.ConversationHistory = append(conversation.ConversationHistory, ConversationMessage{
		Role:      "user",
		Content:   request.Question,
		Timestamp: time.Now(),
	})

	// Detect intent from question and context
	intent, confidence := s.detectIntent(ctx, request, conversation)
	s.iLog.Info(fmt.Sprintf("🔍 Intent Detection - Question: %s", request.Question))
	s.iLog.Info(fmt.Sprintf("🔍 Intent Detection - Editor Type: %s", request.EditorType))
	s.iLog.Info(fmt.Sprintf("🔍 Intent Detection - Detected Intent: %s (confidence: %.2f)", intent, confidence))

	// Find appropriate transaction handler
	handler := s.findHandler(ctx, intent, request.EditorType)
	if handler == nil {
		s.iLog.Info(fmt.Sprintf("⚠️  No specific handler found for intent '%s', using GeneralQuestionHandler", intent))
		// Fall back to general handler
		handler = &GeneralQuestionHandler{
			OpenAIKey:   s.OpenAIKey,
			OpenAIModel: s.OpenAIModel,
			iLog:        s.iLog,
		}
	} else {
		s.iLog.Info(fmt.Sprintf("✅ Using handler: %s for intent: %s", handler.GetName(), intent))
	}

	// Execute handler
	response, err := handler.Handle(ctx, request, conversation)
	if err != nil {
		return nil, fmt.Errorf("handler failed: %w", err)
	}

	// Add assistant response to conversation history
	conversation.ConversationHistory = append(conversation.ConversationHistory, ConversationMessage{
		Role:      "assistant",
		Content:   response.Answer,
		Timestamp: time.Now(),
		Metadata:  response.Metadata,
	})

	// Save conversation state
	if err := s.saveSession(ctx, conversation); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to save session: %v", err))
	}

	response.SessionID = conversation.SessionID
	return response, nil
}

// getOrCreateSession retrieves or creates a conversation session
func (s *AIAgencyService) getOrCreateSession(ctx context.Context, request *AIAgencyRequest) (*ConversationContext, error) {
	if request.SessionID != "" {
		// Try to load existing session
		var sessionData models.AIConversationSession
		err := s.DB.WithContext(ctx).
			Where("session_id = ? AND user_id = ? AND active = ?", request.SessionID, request.UserID, true).
			First(&sessionData).Error

		if err == nil {
			// Session found, deserialize context
			var context ConversationContext
			if err := json.Unmarshal([]byte(sessionData.ContextData), &context); err != nil {
				s.iLog.Error(fmt.Sprintf("Failed to deserialize session context: %v", err))
			} else {
				return &context, nil
			}
		}
	}

	// Create new session
	sessionID := uuid.New().String()
	context := &ConversationContext{
		SessionID:           sessionID,
		UserID:              request.UserID,
		EditorType:          request.EditorType,
		DatabaseAlias:       request.DatabaseAlias,
		EntityID:            request.EntityID,
		EntityContext:       request.EntityContext,
		PageContext:         request.PageContext,
		ConversationHistory: []ConversationMessage{},
		Metadata:            make(map[string]interface{}),
	}

	// Convert history from request
	for _, msg := range request.ConversationHistory {
		role, _ := msg["role"].(string)
		content, _ := msg["content"].(string)
		context.ConversationHistory = append(context.ConversationHistory, ConversationMessage{
			Role:      role,
			Content:   content,
			Timestamp: time.Now(),
		})
	}

	return context, nil
}

// saveSession persists the conversation session
func (s *AIAgencyService) saveSession(ctx context.Context, conversation *ConversationContext) error {
	contextData, err := json.Marshal(conversation)
	if err != nil {
		return fmt.Errorf("failed to serialize session context: %w", err)
	}

	session := models.AIConversationSession{
		SessionID:   conversation.SessionID,
		UserID:      conversation.UserID,
		EditorType:  conversation.EditorType,
		ContextData: string(contextData),
		Active:      true,
		CreatedBy:   conversation.UserID,
		ModifiedBy:  conversation.UserID,
	}

	// Upsert session
	result := s.DB.WithContext(ctx).
		Where("session_id = ? AND user_id = ?", conversation.SessionID, conversation.UserID).
		Assign(map[string]interface{}{
			"context_data": session.ContextData,
			"modifiedby":   session.ModifiedBy,
			"modifiedon":   gorm.Expr("CURRENT_TIMESTAMP"),
		}).
		FirstOrCreate(&session)

	return result.Error
}

// detectIntent analyzes the question to determine user intent using AI
func (s *AIAgencyService) detectIntent(ctx context.Context, request *AIAgencyRequest, conversation *ConversationContext) (string, float64) {
	// Use AI model to detect intent for better accuracy
	intentPrompt := fmt.Sprintf(`Analyze the following user question and determine the intent. The user is currently in the %s editor.

User Question: "%s"

Available Intents:
1. bpm_generation - User wants to generate/create BPM flow, functions, or business logic code
2. page_generation - User wants to generate/create a page or UI component
3. view_generation - User wants to generate/create a view or form
4. workflow_generation - User wants to generate/create a workflow
5. whiteboard_generation - User wants to generate/create whiteboard elements or diagrams
6. query_generation - User wants to generate SQL queries or database queries
7. code_modification - User wants to modify/update existing code
8. clarification - User is asking how to do something or wants explanation/help/tutorial
9. general_question - General question about the system or features

Respond with ONLY the intent name and a confidence score (0.0-1.0) in this exact format:
INTENT: intent_name
CONFIDENCE: 0.95

Important:
- If user asks "how to", "explain", "help me understand" -> clarification
- If user says "generate", "create", "build" with actual requirements -> code generation for the current editor type
- Consider the editor context - if in BPM editor and asking to create/generate something, it's likely bpm_generation`,
		request.EditorType, request.Question)

	// Call OpenAI for intent classification
	response, err := codegen.CallOpenAI(
		s.OpenAIKey,
		s.OpenAIModel,
		[]map[string]interface{}{
			{
				"role":    "system",
				"content": "You are an intent classifier for a low-code development platform. Classify user requests accurately.",
			},
			{
				"role":    "user",
				"content": intentPrompt,
			},
		},
		0.3, // Low temperature for consistent classification
	)

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to detect intent using AI: %v, falling back to keyword matching", err))
		return s.detectIntentKeywordBased(ctx, request, conversation)
	}

	// Parse the AI response
	intent, confidence := s.parseIntentResponse(response)

	// If confidence is too low, fall back to keyword matching
	if confidence < 0.5 {
		s.iLog.Debug("AI intent detection confidence too low, falling back to keyword matching")
		return s.detectIntentKeywordBased(ctx, request, conversation)
	}

	return intent, confidence
}

// detectIntentKeywordBased is the fallback keyword-based intent detection
func (s *AIAgencyService) detectIntentKeywordBased(ctx context.Context, request *AIAgencyRequest, conversation *ConversationContext) (string, float64) {
	question := strings.ToLower(request.Question)

	// Check for code generation keywords
	codeGenKeywords := []string{"generate", "create", "build", "design", "make", "add", "implement", "write", "develop"}
	clarificationKeywords := []string{"what", "how", "why", "when", "where", "explain", "help", "understand", "tutorial", "steps"}

	// Score different intents
	intentScores := make(map[string]float64)

	// Check if this is a clarification/help request first
	if containsAny(question, clarificationKeywords) {
		intentScores["clarification"] = 0.75
	}

	// Editor-specific generation (only if not asking for help/explanation)
	if containsAny(question, codeGenKeywords) && !containsAny(question, clarificationKeywords) {
		switch request.EditorType {
		case "bpm":
			intentScores["bpm_generation"] = 0.85
		case "page":
			intentScores["page_generation"] = 0.85
		case "view":
			intentScores["view_generation"] = 0.85
		case "workflow":
			intentScores["workflow_generation"] = 0.85
		case "whiteboard":
			intentScores["whiteboard_generation"] = 0.85
		}
	}

	// General question (default)
	intentScores["general_question"] = 0.5

	// Find highest scoring intent
	maxIntent := "general_question"
	maxScore := 0.5

	for intent, score := range intentScores {
		if score > maxScore {
			maxScore = score
			maxIntent = intent
		}
	}

	return maxIntent, maxScore
}

// parseIntentResponse parses the AI's intent classification response
func (s *AIAgencyService) parseIntentResponse(response string) (string, float64) {
	lines := strings.Split(response, "\n")
	intent := "general_question"
	confidence := 0.5

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "INTENT:") {
			intent = strings.TrimSpace(strings.TrimPrefix(line, "INTENT:"))
		} else if strings.HasPrefix(line, "CONFIDENCE:") {
			confStr := strings.TrimSpace(strings.TrimPrefix(line, "CONFIDENCE:"))
			if conf, err := parseFloat(confStr); err == nil {
				confidence = conf
			}
		}
	}

	return intent, confidence
}

// parseFloat safely parses a float from a string
func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	// Handle both formats: "0.95" and ".95"
	if strings.HasPrefix(s, ".") {
		s = "0" + s
	}
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// findHandler finds the appropriate transaction handler for the intent
func (s *AIAgencyService) findHandler(ctx context.Context, intent string, editorType string) TransactionHandler {
	handlers := s.getHandlers()

	for _, handler := range handlers {
		if handler.CanHandle(ctx, intent, editorType) {
			return handler
		}
	}

	return nil
}

// getHandlers returns all available transaction handlers
func (s *AIAgencyService) getHandlers() []TransactionHandler {
	return []TransactionHandler{
		// Code modification should be checked first (more specific)
		&CodeModificationHandler{
			OpenAIKey:   s.OpenAIKey,
			OpenAIModel: s.OpenAIModel,
			iLog:        s.iLog,
		},
		&BPMGenerationHandler{
			OpenAIKey:   s.OpenAIKey,
			OpenAIModel: s.OpenAIModel,
			iLog:        s.iLog,
		},
		&QueryGenerationHandler{
			AIReportService:        s.AIReportService,
			SchemaMetadataService:  s.SchemaMetadataService,
			SchemaEmbeddingService: s.SchemaEmbeddingService,
			OpenAIKey:              s.OpenAIKey,
			OpenAIModel:            s.OpenAIModel,
			iLog:                   s.iLog,
		},
		&PageGenerationHandler{
			OpenAIKey:   s.OpenAIKey,
			OpenAIModel: s.OpenAIModel,
			iLog:        s.iLog,
		},
		&ViewGenerationHandler{
			OpenAIKey:   s.OpenAIKey,
			OpenAIModel: s.OpenAIModel,
			iLog:        s.iLog,
		},
		&WorkflowGenerationHandler{
			OpenAIKey:   s.OpenAIKey,
			OpenAIModel: s.OpenAIModel,
			iLog:        s.iLog,
		},
		&WhiteboardGenerationHandler{
			OpenAIKey:   s.OpenAIKey,
			OpenAIModel: s.OpenAIModel,
			iLog:        s.iLog,
		},
	}
}

// Helper function to check if string contains any of the keywords
func containsAny(str string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(str, keyword) {
			return true
		}
	}
	return false
}
