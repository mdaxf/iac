package services

import (
	"context"
	"fmt"

	"github.com/mdaxf/iac/codegen"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// GeneralQuestionHandler handles general Q&A questions
type GeneralQuestionHandler struct {
	OpenAIKey   string
	OpenAIModel string
	iLog        logger.Log
}

func (h *GeneralQuestionHandler) GetName() string {
	return "GeneralQuestionHandler"
}

func (h *GeneralQuestionHandler) CanHandle(ctx context.Context, intent string, editorType string) bool {
	return intent == "general_question" || intent == "clarification"
}

func (h *GeneralQuestionHandler) Handle(ctx context.Context, request *AIAgencyRequest, conversation *ConversationContext) (*AIAgencyResponse, error) {
	// Use existing assistant response generator
	answer, err := codegen.GenerateAssistantResponse(
		request.Question,
		h.OpenAIKey,
		h.OpenAIModel,
		request.PageContext,
		toConversationHistory(conversation.ConversationHistory),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate assistant response: %w", err)
	}

	return &AIAgencyResponse{
		Answer:       answer,
		ResponseType: "answer",
		IntentType:   "general",
		Confidence:   0.8,
	}, nil
}

// BPMGenerationHandler handles BPM flow generation
type BPMGenerationHandler struct {
	OpenAIKey   string
	OpenAIModel string
	iLog        logger.Log
}

func (h *BPMGenerationHandler) GetName() string {
	return "BPMGenerationHandler"
}

func (h *BPMGenerationHandler) CanHandle(ctx context.Context, intent string, editorType string) bool {
	return intent == "bpm_generation" && editorType == "bpm"
}

func (h *BPMGenerationHandler) Handle(ctx context.Context, request *AIAgencyRequest, conversation *ConversationContext) (*AIAgencyResponse, error) {
	// Check if we need clarification
	if !h.hasEnoughContext(request, conversation) {
		return &AIAgencyResponse{
			Answer:         "I'd be happy to help generate a BPM flow. To create the most accurate flow, could you provide more details?",
			ResponseType:   "clarification",
			IntentType:     "bpm_generation",
			RequiresAction: true,
			Questions: []ClarificationQuestion{
				{
					Question: "What is the main business logic you want to implement?",
					Type:     "text",
					Required: true,
				},
				{
					Question: "What type of functions do you need?",
					Type:     "choice",
					Options:  []string{"Database queries", "Data transformation", "External API calls", "Business rules", "All of the above"},
					Required: false,
				},
				{
					Question: "Do you need conditional routing (branching logic)?",
					Type:     "confirm",
					Required: false,
				},
			},
			Confidence: 0.9,
		}, nil
	}

	// Generate BPM flow using multi-step generation
	result, err := codegen.GenerateFlowMultiStep(
		ctx,
		request.Question,
		h.OpenAIKey,
		h.OpenAIModel,
		request.EntityContext,
	)

	if err != nil {
		h.iLog.Error(fmt.Sprintf("Multi-step generation failed: %v", err))
		// Fallback to single-step generation
		h.iLog.Info("Falling back to single-step generation")
		result, err = codegen.GenerateFlowFromDescription(
			request.Question,
			h.OpenAIKey,
			h.OpenAIModel,
			request.EntityContext,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate BPM flow: %w", err)
		}
	}

	// Extract flow data from result
	flowData, ok := result["flow"].(map[string]interface{})
	if !ok || flowData == nil {
		return nil, fmt.Errorf("BPM generation failed: invalid flow data returned")
	}

	// Extract plan if available
	var plan *codegen.GenerationPlan
	if planData, ok := result["plan"]; ok {
		if planPtr, ok := planData.(*codegen.GenerationPlan); ok {
			plan = planPtr
		}
	}

	// Generate summary with plan information
	summary := h.generateFlowSummary(flowData, plan)

	return &AIAgencyResponse{
		Answer:         summary,
		ResponseType:   "code_generation",
		IntentType:     "bpm_generation",
		Data:           flowData,
		RequiresAction: true,
		NextStep:       "apply_to_editor",
		Confidence:     0.95,
	}, nil
}

func (h *BPMGenerationHandler) hasEnoughContext(request *AIAgencyRequest, conversation *ConversationContext) bool {
	// Check if question contains enough detail
	questionLen := len(request.Question)
	if questionLen < 20 {
		return false
	}

	// Check if there's enough conversation history or entity context
	if len(conversation.ConversationHistory) >= 2 || len(request.EntityContext) > 0 {
		return true
	}

	return questionLen > 50
}

func (h *BPMGenerationHandler) generateFlowSummary(flowData map[string]interface{}, plan *codegen.GenerationPlan) string {
	// Generate a summary of the created flow
	summary := "✅ I've generated a BPM flow for you using multi-step AI generation.\n\n"

	// Add plan information if available
	if plan != nil {
		summary += "Generation Plan:\n"
		summary += fmt.Sprintf("• Function Count: %d\n", plan.FunctionCount)
		summary += fmt.Sprintf("• Function Groups: %d\n", plan.FunctionGroups)
		if plan.HasRouting {
			summary += "• Routing: Yes (conditional branching included)\n"
		} else {
			summary += "• Routing: No (sequential flow)\n"
		}
		summary += "\n"
	}

	summary += "The flow includes:\n"

	// Extract function groups count
	if functionGroups, ok := flowData["functiongroups"].([]interface{}); ok {
		summary += fmt.Sprintf("• %d Function Group(s)\n", len(functionGroups))

		totalFunctions := 0
		for _, fg := range functionGroups {
			if fgMap, ok := fg.(map[string]interface{}); ok {
				if functions, ok := fgMap["functions"].([]interface{}); ok {
					totalFunctions += len(functions)
				}
			}
		}
		summary += fmt.Sprintf("• %d Function(s) total\n", totalFunctions)
	}

	summary += "\nYou can review the flow details and apply it to your BPM editor."

	return summary
}

// QueryGenerationHandler handles SQL query generation
type QueryGenerationHandler struct {
	AIReportService        *AIReportService
	SchemaMetadataService  *SchemaMetadataService
	SchemaEmbeddingService *SchemaEmbeddingService
	OpenAIKey              string
	OpenAIModel            string
	iLog                   logger.Log
}

func (h *QueryGenerationHandler) GetName() string {
	return "QueryGenerationHandler"
}

func (h *QueryGenerationHandler) CanHandle(ctx context.Context, intent string, editorType string) bool {
	return intent == "query_generation"
}

func (h *QueryGenerationHandler) Handle(ctx context.Context, request *AIAgencyRequest, conversation *ConversationContext) (*AIAgencyResponse, error) {
	// Get database alias - default to "it" if not provided
	databaseAlias := request.DatabaseAlias
	if databaseAlias == "" {
		databaseAlias = conversation.DatabaseAlias
	}
	if databaseAlias == "" {
		// Default to "it" (the default database)
		databaseAlias = "it"
		h.iLog.Info("No database alias provided, using default 'it'")
	}

	// Get schema information using existing AI report service logic
	schemaInfo, err := h.getSchemaContext(ctx, databaseAlias, request.Question)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema context: %w", err)
	}

	// Generate SQL using AI report service
	sqlRequest := Text2SQLRequest{
		Question:   request.Question,
		DatabaseID: databaseAlias,
	}

	sqlResponse, err := h.AIReportService.GenerateSQL(ctx, sqlRequest, schemaInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SQL: %w", err)
	}

	// Format response
	answer := fmt.Sprintf("I've generated a SQL query for you:\n\n```sql\n%s\n```\n\n**Explanation:** %s\n\n**Tables Used:** %v\n**Confidence:** %.0f%%",
		sqlResponse.SQL,
		sqlResponse.Explanation,
		sqlResponse.TablesUsed,
		sqlResponse.Confidence*100,
	)

	return &AIAgencyResponse{
		Answer:       answer,
		ResponseType: "code_generation",
		IntentType:   "query_generation",
		Data: map[string]interface{}{
			"sql":          sqlResponse.SQL,
			"explanation":  sqlResponse.Explanation,
			"tables_used":  sqlResponse.TablesUsed,
			"columns_used": sqlResponse.ColumnsUsed,
			"query_type":   sqlResponse.QueryType,
		},
		RequiresAction: true,
		NextStep:       "execute_query",
		Confidence:     sqlResponse.Confidence,
	}, nil
}

func (h *QueryGenerationHandler) getSchemaContext(ctx context.Context, databaseAlias string, question string) (string, error) {
	// Get relevant schema information using vector search
	relevantTables, err := h.SchemaEmbeddingService.SearchSimilarTables(ctx, databaseAlias, question, 10)
	if err != nil {
		h.iLog.Warn(fmt.Sprintf("Failed to search similar tables: %v, falling back to metadata", err))

		// Fallback: Get all metadata
		metadata, err := h.SchemaMetadataService.GetDatabaseMetadata(ctx, databaseAlias)
		if err != nil {
			return "", fmt.Errorf("failed to get database metadata: %w", err)
		}

		// Build schema context from metadata
		return h.buildSchemaInfoFromMetadata(metadata), nil
	}

	// Build schema context from relevant tables
	return h.buildSchemaInfoFromEmbeddings(relevantTables), nil
}

func (h *QueryGenerationHandler) buildSchemaInfoFromMetadata(metadata []models.DatabaseSchemaMetadata) string {
	schemaInfo := "Database Schema:\n\n"

	// Group metadata by table
	tableMap := make(map[string][]models.DatabaseSchemaMetadata)
	for _, meta := range metadata {
		tableMap[meta.Table] = append(tableMap[meta.Table], meta)
	}

	// Build schema info
	for tableName, columns := range tableMap {
		schemaInfo += fmt.Sprintf("Table: %s\n", tableName)

		// Find table description (from first entry with type='table')
		for _, meta := range columns {
			if meta.MetadataType == models.MetadataTypeTable && meta.Description != "" {
				schemaInfo += fmt.Sprintf("Description: %s\n", meta.Description)
				break
			}
		}

		schemaInfo += "Columns:\n"
		for _, col := range columns {
			if col.MetadataType == models.MetadataTypeColumn && col.Column != "" {
				nullable := ""
				if col.IsNullable != nil && *col.IsNullable {
					nullable = " (nullable)"
				}
				pk := ""
				if col.IsPrimaryKey != nil && *col.IsPrimaryKey {
					pk = " [PK]"
				}
				schemaInfo += fmt.Sprintf("  - %s %s%s%s\n", col.Column, col.DataType, nullable, pk)
			}
		}
		schemaInfo += "\n"
	}

	return schemaInfo
}

func (h *QueryGenerationHandler) buildSchemaInfoFromEmbeddings(tables []models.DatabaseSchemaMetadata) string {
	schemaInfo := "Relevant Database Schema:\n\n"

	// Group metadata by table
	tableMap := make(map[string][]models.DatabaseSchemaMetadata)
	for _, meta := range tables {
		tableMap[meta.Table] = append(tableMap[meta.Table], meta)
	}

	// Build schema info
	for tableName, columns := range tableMap {
		schemaInfo += fmt.Sprintf("Table: %s\n", tableName)

		// Find table description
		for _, meta := range columns {
			if meta.MetadataType == models.MetadataTypeTable && meta.Description != "" {
				schemaInfo += fmt.Sprintf("Description: %s\n", meta.Description)
				break
			}
		}

		// List columns
		schemaInfo += "Columns:\n"
		for _, col := range columns {
			if col.MetadataType == models.MetadataTypeColumn && col.Column != "" {
				schemaInfo += fmt.Sprintf("  - %s (%s)\n", col.Column, col.DataType)
			}
		}
		schemaInfo += "\n"
	}

	return schemaInfo
}

// PageGenerationHandler handles page structure generation
type PageGenerationHandler struct {
	OpenAIKey   string
	OpenAIModel string
	iLog        logger.Log
}

func (h *PageGenerationHandler) GetName() string {
	return "PageGenerationHandler"
}

func (h *PageGenerationHandler) CanHandle(ctx context.Context, intent string, editorType string) bool {
	return intent == "page_generation" && editorType == "page"
}

func (h *PageGenerationHandler) Handle(ctx context.Context, request *AIAgencyRequest, conversation *ConversationContext) (*AIAgencyResponse, error) {
	// Generate page using multi-step generation
	result, err := codegen.GeneratePageMultiStep(
		ctx,
		request.Question,
		h.OpenAIKey,
		h.OpenAIModel,
		request.EntityContext,
	)

	if err != nil {
		h.iLog.Error(fmt.Sprintf("Multi-step generation failed: %v", err))
		// Fallback to single-step generation
		h.iLog.Info("Falling back to single-step generation")
		result, err = codegen.GeneratePageFromDescription(
			request.Question,
			h.OpenAIKey,
			h.OpenAIModel,
			request.EntityContext,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate page: %w", err)
		}
	}

	// Extract page data and plan
	pageData, _ := result["page"].(map[string]interface{})
	var plan *codegen.PageGenerationPlan
	if planData, ok := result["plan"]; ok {
		if planPtr, ok := planData.(*codegen.PageGenerationPlan); ok {
			plan = planPtr
		}
	}

	// Generate enhanced summary
	summary := h.generatePageSummary(pageData, plan)

	return &AIAgencyResponse{
		Answer:         summary,
		ResponseType:   "code_generation",
		IntentType:     "page_generation",
		Data:           result,
		RequiresAction: true,
		NextStep:       "apply_to_editor",
		Confidence:     0.9,
	}, nil
}

func (h *PageGenerationHandler) generatePageSummary(pageData map[string]interface{}, plan *codegen.PageGenerationPlan) string {
	summary := "✅ I've generated a page structure for you using multi-step AI generation.\n\n"

	// Add plan information if available
	if plan != nil {
		summary += "Generation Plan:\n"
		summary += fmt.Sprintf("• Page Type: %s\n", plan.PageType)
		summary += fmt.Sprintf("• Layout Type: %s\n", plan.LayoutType)
		summary += fmt.Sprintf("• Component Count: %d\n", plan.ComponentCount)
		summary += fmt.Sprintf("• Section Count: %d\n", plan.SectionCount)
		summary += "\n"
	}

	// Add page structure information
	if sections, ok := pageData["sections"].([]interface{}); ok {
		summary += fmt.Sprintf("The page includes:\n")
		summary += fmt.Sprintf("• %d Section(s)\n", len(sections))

		totalComponents := 0
		for _, section := range sections {
			if sectionMap, ok := section.(map[string]interface{}); ok {
				if components, ok := sectionMap["components"].([]interface{}); ok {
					totalComponents += len(components)
				}
			}
		}
		summary += fmt.Sprintf("• %d Component(s) total\n\n", totalComponents)
	}

	summary += "You can review the page structure and apply it to your editor."

	return summary
}

// ViewGenerationHandler handles view structure generation
type ViewGenerationHandler struct {
	OpenAIKey   string
	OpenAIModel string
	iLog        logger.Log
}

func (h *ViewGenerationHandler) GetName() string {
	return "ViewGenerationHandler"
}

func (h *ViewGenerationHandler) CanHandle(ctx context.Context, intent string, editorType string) bool {
	return intent == "view_generation" && editorType == "view"
}

func (h *ViewGenerationHandler) Handle(ctx context.Context, request *AIAgencyRequest, conversation *ConversationContext) (*AIAgencyResponse, error) {
	result, err := codegen.GenerateViewFromDescription(
		request.Question,
		h.OpenAIKey,
		h.OpenAIModel,
		request.EntityContext,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate view: %w", err)
	}

	return &AIAgencyResponse{
		Answer:         "✅ I've generated a view structure for you. You can review and apply it to your view editor.",
		ResponseType:   "code_generation",
		IntentType:     "view_generation",
		Data:           result,
		RequiresAction: true,
		NextStep:       "apply_to_editor",
		Confidence:     0.9,
	}, nil
}

// WorkflowGenerationHandler handles workflow structure generation
type WorkflowGenerationHandler struct {
	OpenAIKey   string
	OpenAIModel string
	iLog        logger.Log
}

func (h *WorkflowGenerationHandler) GetName() string {
	return "WorkflowGenerationHandler"
}

func (h *WorkflowGenerationHandler) CanHandle(ctx context.Context, intent string, editorType string) bool {
	return intent == "workflow_generation" && editorType == "workflow"
}

func (h *WorkflowGenerationHandler) Handle(ctx context.Context, request *AIAgencyRequest, conversation *ConversationContext) (*AIAgencyResponse, error) {
	result, err := codegen.GenerateWorkflowFromDescription(
		request.Question,
		h.OpenAIKey,
		h.OpenAIModel,
		request.EntityContext,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate workflow: %w", err)
	}

	return &AIAgencyResponse{
		Answer:         "✅ I've generated a workflow structure for you. You can review and apply it to your workflow editor.",
		ResponseType:   "code_generation",
		IntentType:     "workflow_generation",
		Data:           result,
		RequiresAction: true,
		NextStep:       "apply_to_editor",
		Confidence:     0.9,
	}, nil
}

// WhiteboardGenerationHandler handles whiteboard structure generation
type WhiteboardGenerationHandler struct {
	OpenAIKey   string
	OpenAIModel string
	iLog        logger.Log
}

func (h *WhiteboardGenerationHandler) GetName() string {
	return "WhiteboardGenerationHandler"
}

func (h *WhiteboardGenerationHandler) CanHandle(ctx context.Context, intent string, editorType string) bool {
	return intent == "whiteboard_generation" && editorType == "whiteboard"
}

func (h *WhiteboardGenerationHandler) Handle(ctx context.Context, request *AIAgencyRequest, conversation *ConversationContext) (*AIAgencyResponse, error) {
	result, err := codegen.GenerateWhiteboardFromDescription(
		request.Question,
		h.OpenAIKey,
		h.OpenAIModel,
		request.EntityContext,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate whiteboard: %w", err)
	}

	return &AIAgencyResponse{
		Answer:         "✅ I've generated a whiteboard structure for you. You can review and apply it to your whiteboard editor.",
		ResponseType:   "code_generation",
		IntentType:     "whiteboard_generation",
		Data:           result,
		RequiresAction: true,
		NextStep:       "apply_to_editor",
		Confidence:     0.9,
	}, nil
}

// Helper function to convert conversation history to the format expected by codegen
func toConversationHistory(messages []ConversationMessage) []map[string]interface{} {
	history := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		history[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}
	return history
}

// CodeModificationHandler handles code/query/script modifications for existing functions
type CodeModificationHandler struct {
	OpenAIKey   string
	OpenAIModel string
	iLog        logger.Log
}

func (h *CodeModificationHandler) GetName() string {
	return "CodeModificationHandler"
}

func (h *CodeModificationHandler) CanHandle(ctx context.Context, intent string, editorType string) bool {
	// Handle code_modification intent for BPM editor
	return intent == "code_modification" && editorType == "bpm"
}

func (h *CodeModificationHandler) Handle(ctx context.Context, request *AIAgencyRequest, conversation *ConversationContext) (*AIAgencyResponse, error) {
	// Check if we have entity context (selected function/query)
	// EntityContext is a top-level field in AIAgencyRequest
	if request.EntityContext == nil || len(request.EntityContext) == 0 {
		h.iLog.Warn("No entity_context provided for code modification")
		return nil, fmt.Errorf("code modification requires a selected function/query")
	}

	entityContext := request.EntityContext

	// Get the system prompt override if provided (from Options)
	systemPrompt := ""
	if request.Options != nil {
		if sp, ok := request.Options["system_prompt_override"].(string); ok {
			systemPrompt = sp
		}
	}

	// Build the full context for code generation including current function data
	fullContext := map[string]interface{}{
		"entity_context":   entityContext,
		"page_context":     request.PageContext,
		"conversation":     toConversationHistory(conversation.ConversationHistory),
		"system_prompt":    systemPrompt,
	}

	// Merge entity context into the full context for flow generation
	if data, ok := entityContext["data"].(map[string]interface{}); ok {
		for k, v := range data {
			fullContext[k] = v
		}
	}

	// Use GenerateFlowFromDescription with full context
	result, err := codegen.GenerateFlowFromDescription(
		request.Question,
		h.OpenAIKey,
		h.OpenAIModel,
		fullContext,
	)

	if err != nil {
		h.iLog.Error(fmt.Sprintf("Failed to generate code modification: %v", err))
		return nil, fmt.Errorf("failed to generate code modification: %w", err)
	}

	// Extract the specific modified function from the result
	// The result is a flow structure, but we need just the function for context-aware modification
	modifiedFunction := extractModifiedFunction(result, entityContext)

	if modifiedFunction == nil {
		h.iLog.Warn("Could not extract modified function from AI response")
		// Fall back to returning the whole result
		modifiedFunction = result
	}

	// CRITICAL: Ensure the function has the correct ID for updating
	// Even if AI didn't include it, we need to set it from the entity context
	if metadata, ok := entityContext["metadata"].(map[string]interface{}); ok {
		if targetID, ok := metadata["id"].(string); ok {
			// Set the ID to match the target function
			modifiedFunction["id"] = targetID
			h.iLog.Info(fmt.Sprintf("Ensured function ID is set to: %s", targetID))

			// Also set name if available
			if targetName, ok := metadata["name"].(string); ok {
				if _, hasName := modifiedFunction["name"]; !hasName {
					modifiedFunction["name"] = targetName
				}
			}

			// Set functype if available
			if funcType, ok := metadata["functionType"].(float64); ok {
				if _, hasFuncType := modifiedFunction["functype"]; !hasFuncType {
					modifiedFunction["functype"] = funcType
				}
			}
		}
	}

	h.iLog.Info(fmt.Sprintf("Code modification successful, returning modified function"))

	return &AIAgencyResponse{
		Answer:         "✅ I've updated the code based on your request. You can review and apply the changes.",
		ResponseType:   "code_generation",
		IntentType:     "code_modification",
		Data:           modifiedFunction,
		RequiresAction: true,
		NextStep:       "apply_to_editor",
		Confidence:     0.9,
	}, nil
}

// extractModifiedFunction extracts the specific function from the AI-generated flow
func extractModifiedFunction(flowResult map[string]interface{}, entityContext map[string]interface{}) map[string]interface{} {
	// Get the target function ID from entity context
	metadata, ok := entityContext["metadata"].(map[string]interface{})
	if !ok {
		return nil
	}

	targetFunctionID, ok := metadata["id"].(string)
	if !ok {
		return nil
	}

	// Check if result has a "flow" key
	var flow map[string]interface{}
	if flowData, ok := flowResult["flow"].(map[string]interface{}); ok {
		flow = flowData
	} else {
		flow = flowResult
	}

	// Look for function groups in the flow
	functionGroups, ok := flow["functiongroups"].([]interface{})
	if !ok {
		return nil
	}

	// Search for the modified function in the function groups
	for _, fg := range functionGroups {
		fgMap, ok := fg.(map[string]interface{})
		if !ok {
			continue
		}

		functions, ok := fgMap["functions"].([]interface{})
		if !ok {
			continue
		}

		for _, fn := range functions {
			fnMap, ok := fn.(map[string]interface{})
			if !ok {
				continue
			}

			if fnID, ok := fnMap["id"].(string); ok && fnID == targetFunctionID {
				// Found the modified function!
				return fnMap
			}
		}
	}

	return nil
}
