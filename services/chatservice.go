package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// SchemaDiscoveryService is an interface for services that can discover database schema
type SchemaDiscoveryService interface {
	GetDatabaseMetadata(ctx context.Context, databaseAlias string) ([]models.DatabaseSchemaMetadata, error)
}

// ChatService handles chat and AI conversation features
type ChatService struct {
	DB                    *gorm.DB
	OpenAIKey             string
	OpenAIModel           string
	SchemaMetadataService SchemaDiscoveryService
	iLog                  logger.Log
}

// NewChatService creates a new chat service
func NewChatService(db *gorm.DB, openAIKey, openAIModel string, schemaService SchemaDiscoveryService) *ChatService {
	return &ChatService{
		DB:                    db,
		OpenAIKey:             openAIKey,
		OpenAIModel:           openAIModel,
		SchemaMetadataService: schemaService,
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "ChatService",
		},
	}
}

// CreateConversation creates a new conversation
func (s *ChatService) CreateConversation(userID, databaseAlias, title string, autoExecute bool) (*models.Conversation, error) {
	conversation := &models.Conversation{
		ID:               uuid.New().String(),
		Title:            title,
		UserID:           userID,
		DatabaseAlias:    databaseAlias,
		AutoExecuteQuery: autoExecute,
		Active:           true,
	}

	if err := s.DB.Create(conversation).Error; err != nil {
		return nil, err
	}

	return conversation, nil
}

// GetConversation retrieves a conversation by ID
func (s *ChatService) GetConversation(id string) (*models.Conversation, error) {
	var conversation models.Conversation
	err := s.DB.Preload("Messages").First(&conversation, "id = ? AND active = ?", id, true).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

// ListConversations retrieves conversations for a user
func (s *ChatService) ListConversations(userID string, limit int) ([]models.Conversation, error) {
	var conversations []models.Conversation
	err := s.DB.Where("userid = ? AND active = ?", userID, true).
		Order("modifiedon DESC").
		Limit(limit).
		Find(&conversations).Error
	return conversations, err
}

// DeleteConversation soft deletes a conversation
func (s *ChatService) DeleteConversation(id string) error {
	return s.DB.Model(&models.Conversation{}).
		Where("id = ?", id).
		Update("active", false).Error
}

// ProcessMessage processes a user message and generates AI response
func (s *ChatService) ProcessMessage(ctx context.Context, conversationID, userMessage, databaseAlias string, autoExecute bool) (*models.ChatMessage, error) {
	s.iLog.Info(fmt.Sprintf("ProcessMessage START - ConversationID: %s, DatabaseAlias: %s, Message: %s", conversationID, databaseAlias, userMessage))

	// 1. Save user message
	s.iLog.Debug("Step 1: Saving user message to database")
	userMsg := &models.ChatMessage{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		MessageType:    models.MessageTypeUser,
		Text:           userMessage,
	}

	if err := s.DB.Create(userMsg).Error; err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to save user message: %v", err))
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}
	s.iLog.Debug(fmt.Sprintf("User message saved with ID: %s", userMsg.ID))

	// 2. Get schema context
	s.iLog.Info(fmt.Sprintf("Step 2: Retrieving schema context for database alias: %s", databaseAlias))
	schemaContext, err := s.getSchemaContext(databaseAlias, userMessage)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get schema context: %v", err))
		return nil, fmt.Errorf("failed to get schema context: %w", err)
	}
	s.iLog.Info(fmt.Sprintf("Schema context retrieved, length: %d characters", len(schemaContext)))
	s.iLog.Debug(fmt.Sprintf("Schema context content:\n%s", schemaContext))

	// 3. Generate SQL using AI
	s.iLog.Info("Step 3: Generating SQL query using AI")
	sqlResponse, err := s.generateSQLWithContext(ctx, userMessage, schemaContext)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to generate SQL: %v", err))
		return nil, fmt.Errorf("failed to generate SQL: %w", err)
	}
	s.iLog.Info(fmt.Sprintf("SQL generated successfully, confidence: %.2f", sqlResponse.Confidence))
	s.iLog.Debug(fmt.Sprintf("Generated SQL: %s", sqlResponse.SQL))

	// 4. Create assistant response
	s.iLog.Debug("Step 4: Creating assistant response message")
	confidence := sqlResponse.Confidence
	assistantMsg := &models.ChatMessage{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		MessageType:    models.MessageTypeAssistant,
		Text:           sqlResponse.Explanation,
		SQLQuery:       sqlResponse.SQL,
		SQLConfidence:  &confidence,
		Provenance: map[string]interface{}{
			"tables_used":  sqlResponse.TablesUsed,
			"columns_used": sqlResponse.ColumnsUsed,
			"reasoning":    sqlResponse.Reasoning,
			"query_type":   sqlResponse.QueryType,
		},
	}

	// 5. Execute SQL if auto-execute is enabled
	if autoExecute && sqlResponse.SQL != "" {
		s.iLog.Info("Step 5: Auto-executing SQL query")
		startTime := time.Now()
		resultData, rowCount, execErr := s.executeSQL(databaseAlias, sqlResponse.SQL)
		executionTime := int(time.Since(startTime).Milliseconds())
		assistantMsg.ExecutionTimeMs = &executionTime

		if execErr != nil {
			s.iLog.Error(fmt.Sprintf("SQL execution failed: %v", execErr))
			assistantMsg.ErrorMessage = execErr.Error()
		} else {
			s.iLog.Info(fmt.Sprintf("SQL executed successfully, returned %d rows in %dms", rowCount, executionTime))
			assistantMsg.ResultData = resultData
			assistantMsg.RowCount = &rowCount
		}
	} else {
		s.iLog.Debug("Step 5: Skipping SQL execution (auto-execute disabled or no SQL generated)")
	}

	// 6. Save assistant message
	s.iLog.Debug("Step 6: Saving assistant message to database")
	if err := s.DB.Create(assistantMsg).Error; err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to save assistant message: %v", err))
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}
	s.iLog.Debug(fmt.Sprintf("Assistant message saved with ID: %s", assistantMsg.ID))

	// 7. Log AI generation
	s.iLog.Debug("Step 7: Logging AI generation metadata")
	s.logAIGeneration(conversationID, assistantMsg.ID, sqlResponse)

	s.iLog.Info(fmt.Sprintf("ProcessMessage COMPLETE - ConversationID: %s, AssistantMsgID: %s", conversationID, assistantMsg.ID))
	return assistantMsg, nil
}

// getSchemaContext retrieves relevant schema information using vector search
func (s *ChatService) getSchemaContext(databaseAlias, question string) (string, error) {
	s.iLog.Info(fmt.Sprintf("getSchemaContext START - DatabaseAlias: %s", databaseAlias))

	// 1. Get business entities
	s.iLog.Debug("Retrieving relevant business entities")
	businessEntities, err := s.getRelevantBusinessEntities(databaseAlias, question)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get business entities: %v", err))
		return "", err
	}
	s.iLog.Info(fmt.Sprintf("Found %d business entities", len(businessEntities)))

	// 2. Get table metadata
	s.iLog.Debug("Retrieving relevant table metadata")
	tableMetadata, err := s.getRelevantTableMetadata(databaseAlias, question)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get table metadata: %v", err))
		return "", err
	}
	s.iLog.Info(fmt.Sprintf("Found %d metadata entries (tables and columns)", len(tableMetadata)))

	// 3. Build context string
	s.iLog.Debug("Building schema context string")
	var contextBuilder strings.Builder

	contextBuilder.WriteString("=== DATABASE SCHEMA CONTEXT ===\n\n")

	if len(businessEntities) > 0 {
		contextBuilder.WriteString("BUSINESS ENTITIES:\n")
		for _, entity := range businessEntities {
			contextBuilder.WriteString(fmt.Sprintf("- %s (%s): %s\n", entity.EntityName, entity.EntityType, entity.Description))
			if entity.CalculationFormula != "" {
				contextBuilder.WriteString(fmt.Sprintf("  Formula: %s\n", entity.CalculationFormula))
			}
		}
		contextBuilder.WriteString("\n")
	} else {
		s.iLog.Warn("No business entities found for context")
	}

	if len(tableMetadata) > 0 {
		// Count tables vs columns
		tableCount := 0
		columnCount := 0
		for _, meta := range tableMetadata {
			if meta.MetadataType == models.MetadataTypeTable {
				tableCount++
			} else {
				columnCount++
			}
		}
		s.iLog.Info(fmt.Sprintf("Building context with %d tables and %d columns", tableCount, columnCount))

		contextBuilder.WriteString("TABLES AND COLUMNS:\n")
		for _, meta := range tableMetadata {
			if meta.MetadataType == models.MetadataTypeTable {
				contextBuilder.WriteString(fmt.Sprintf("Table: %s - %s\n", meta.Table, meta.Description))
			} else {
				contextBuilder.WriteString(fmt.Sprintf("  - %s.%s (%s): %s\n",
					meta.Table, meta.Column, meta.DataType, meta.ColumnComment))
			}
		}
	} else {
		s.iLog.Warn("WARNING: No table metadata found - schema context will be empty!")
	}

	contextStr := contextBuilder.String()
	s.iLog.Info(fmt.Sprintf("getSchemaContext COMPLETE - Context length: %d characters", len(contextStr)))
	return contextStr, nil
}

// getRelevantBusinessEntities finds business entities relevant to the question
func (s *ChatService) getRelevantBusinessEntities(databaseAlias, question string) ([]models.BusinessEntity, error) {
	var entities []models.BusinessEntity

	// Try full-text search first (requires FULLTEXT index)
	err := s.DB.Where("databasealias = ?", databaseAlias).
		Where("MATCH(entityname, description) AGAINST(? IN NATURAL LANGUAGE MODE)", question).
		Limit(5).
		Find(&entities).Error

	// If FULLTEXT search fails (no index or other error), fall back to simple query
	if err != nil || len(entities) == 0 {
		err = s.DB.Where("databasealias = ?", databaseAlias).
			Limit(10).
			Find(&entities).Error
	}

	return entities, err
}

// getRelevantTableMetadata finds table/column metadata relevant to the question
func (s *ChatService) getRelevantTableMetadata(databaseAlias, question string) ([]models.DatabaseSchemaMetadata, error) {
	s.iLog.Info(fmt.Sprintf("getRelevantTableMetadata START - DatabaseAlias: %s", databaseAlias))

	// Handle empty database alias
	if databaseAlias == "" {
		s.iLog.Warn("Database alias is empty, defaulting to 'default'")
		databaseAlias = "default"
	}

	var metadata []models.DatabaseSchemaMetadata

	// Try full-text search first (requires FULLTEXT index)
	s.iLog.Debug("Attempting full-text search on schema metadata")
	err := s.DB.Where("databasealias = ?", databaseAlias).
		Where("MATCH(description, columncomment) AGAINST(? IN NATURAL LANGUAGE MODE)", question).
		Limit(10).
		Find(&metadata).Error

	if err != nil {
		s.iLog.Warn(fmt.Sprintf("Full-text search failed (likely no FULLTEXT index): %v", err))
	} else if len(metadata) > 0 {
		// Filter out invalid metadata entries
		metadata = filterValidMetadata(metadata)
		if len(metadata) > 0 {
			s.iLog.Info(fmt.Sprintf("Full-text search found %d valid metadata entries", len(metadata)))
			return metadata, nil
		} else {
			s.iLog.Debug("Full-text search returned only invalid entries")
		}
	} else {
		s.iLog.Debug("Full-text search returned 0 results")
	}

	// If FULLTEXT search fails (no index or other error), fall back to get all tables
	if err != nil || len(metadata) == 0 {
		s.iLog.Debug("Falling back to simple query for all tables")
		// Get all tables and some columns for the database
		err = s.DB.Where("databasealias = ?", databaseAlias).
			Order("tablename, metadatatype DESC").
			Limit(50).
			Find(&metadata).Error

		if err != nil {
			s.iLog.Error(fmt.Sprintf("Simple query failed: %v", err))
		} else {
			// Filter out invalid metadata entries
			originalCount := len(metadata)
			metadata = filterValidMetadata(metadata)
			if originalCount != len(metadata) {
				s.iLog.Warn(fmt.Sprintf("Filtered out %d invalid metadata entries (empty table names)", originalCount-len(metadata)))
			}
			s.iLog.Info(fmt.Sprintf("Simple query found %d valid metadata entries", len(metadata)))
		}
	}

	// If still no metadata found and SchemaMetadataService is available, use auto-discovery
	if len(metadata) == 0 && s.SchemaMetadataService != nil {
		s.iLog.Warn(fmt.Sprintf("No valid metadata found in schemameta table for alias '%s', triggering auto-discovery", databaseAlias))
		// Use auto-discovery fallback from SchemaMetadataService
		ctx := context.Background()
		metadata, err = s.SchemaMetadataService.GetDatabaseMetadata(ctx, databaseAlias)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Auto-discovery failed: %v", err))
			return nil, fmt.Errorf("failed to discover schema metadata: %w", err)
		}
		s.iLog.Info(fmt.Sprintf("Auto-discovery completed, found %d metadata entries", len(metadata)))
	} else if len(metadata) == 0 && s.SchemaMetadataService == nil {
		s.iLog.Error("No metadata found and SchemaMetadataService is nil - cannot perform auto-discovery")
	}

	s.iLog.Info(fmt.Sprintf("getRelevantTableMetadata COMPLETE - Returning %d metadata entries", len(metadata)))
	return metadata, err
}

// filterValidMetadata filters out invalid metadata entries (empty table names, etc.)
func filterValidMetadata(metadata []models.DatabaseSchemaMetadata) []models.DatabaseSchemaMetadata {
	var valid []models.DatabaseSchemaMetadata
	for _, meta := range metadata {
		// Skip entries with empty table names
		if meta.Table == "" {
			continue
		}
		// Skip column entries with empty column names
		if meta.MetadataType == models.MetadataTypeColumn && meta.Column == "" {
			continue
		}
		valid = append(valid, meta)
	}
	return valid
}

// generateSQLWithContext generates SQL using AI with schema context
func (s *ChatService) generateSQLWithContext(ctx context.Context, question, schemaContext string) (*Text2SQLResponse, error) {
	s.iLog.Info(fmt.Sprintf("generateSQLWithContext START - Question: %s", question))
	s.iLog.Debug(fmt.Sprintf("Schema context length: %d characters", len(schemaContext)))

	systemPrompt := `You are an expert SQL query generator that converts natural language questions into accurate SQL queries.

CORE PRINCIPLES:
1. Generate syntactically correct SQL for MySQL
2. Use proper table and column names from the provided schema
3. Apply appropriate filters, joins, and aggregations
4. Ensure queries are efficient and follow best practices
5. Provide clear explanations for your reasoning

RESPONSE FORMAT (JSON):
You MUST ALWAYS respond with valid JSON in this exact format, even if you cannot generate a perfect query:
{
  "sql": "The generated SQL query (or empty string if cannot generate)",
  "explanation": "Clear explanation of what the query does or why it cannot be generated",
  "confidence": 0.95,
  "tables_used": ["table1", "table2"],
  "columns_used": ["column1", "column2"],
  "reasoning": "Step-by-step reasoning process",
  "query_type": "SELECT"
}

If schema information is limited or missing, make reasonable assumptions based on common database patterns and still return valid JSON with low confidence.

IMPORTANT:
- ALWAYS return valid JSON, never plain text
- Only generate SELECT queries (read-only)
- Never generate INSERT, UPDATE, DELETE, DROP, ALTER, or other dangerous operations
- Add LIMIT clause if not specified (default: 100)
- Use clear aliases for readability`

	userPrompt := fmt.Sprintf(`%s

### QUESTION ###
User's Question: %s

### INSTRUCTIONS ###
1. Analyze the question to understand what data is being requested
2. Identify the relevant tables and columns from the schema
3. Determine the appropriate joins, filters, and aggregations
4. Generate the SQL query following best practices
5. Provide reasoning for your choices

Respond with JSON only, no additional text.`, schemaContext, question)

	s.iLog.Debug("Building OpenAI API request")
	// Call OpenAI
	reqBody := OpenAIRequest{
		Model: s.OpenAIModel,
		Messages: []OpenAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.1,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to marshal OpenAI request: %v", err))
		return nil, err
	}
	s.iLog.Debug(fmt.Sprintf("OpenAI request body size: %d bytes", len(jsonBody)))

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to create HTTP request: %v", err))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.OpenAIKey)

	s.iLog.Info(fmt.Sprintf("Sending request to OpenAI API (model: %s)", s.OpenAIModel))
	client := &http.Client{Timeout: 60 * time.Second}
	startTime := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(startTime)

	if err != nil {
		s.iLog.Error(fmt.Sprintf("OpenAI API request failed after %v: %v", elapsed, err))
		return nil, err
	}
	defer resp.Body.Close()

	s.iLog.Info(fmt.Sprintf("OpenAI API responded in %v with status: %d", elapsed, resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		s.iLog.Error(fmt.Sprintf("OpenAI API returned error status: %d", resp.StatusCode))
		return nil, fmt.Errorf("OpenAI API error: status %d", resp.StatusCode)
	}

	s.iLog.Debug("Parsing OpenAI response")
	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to decode OpenAI response: %v", err))
		return nil, err
	}

	if len(openAIResp.Choices) == 0 {
		s.iLog.Error("OpenAI response contains no choices")
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	s.iLog.Debug(fmt.Sprintf("OpenAI returned %d choices", len(openAIResp.Choices)))

	// Parse SQL response
	var result Text2SQLResponse
	rawContent := openAIResp.Choices[0].Message.Content
	s.iLog.Debug(fmt.Sprintf("Raw AI response (first 500 chars): %s", truncateString(rawContent, 500)))

	cleanedContent := cleanJSONResponse(rawContent)
	s.iLog.Debug("Attempting to parse AI response as JSON")

	if err := json.Unmarshal([]byte(cleanedContent), &result); err != nil {
		// If JSON parsing fails and content looks like plain text, provide helpful error
		if !strings.HasPrefix(strings.TrimSpace(cleanedContent), "{") {
			s.iLog.Error(fmt.Sprintf("AI returned plain text instead of JSON: %s", truncateString(cleanedContent, 200)))
			return nil, fmt.Errorf("AI returned plain text instead of JSON. This usually means insufficient schema information. Response: %s", cleanedContent)
		}
		s.iLog.Error(fmt.Sprintf("Failed to parse JSON response: %v", err))
		s.iLog.Debug(fmt.Sprintf("Cleaned content that failed to parse: %s", cleanedContent))
		return nil, fmt.Errorf("failed to parse JSON response: %w (content: %s)", err, cleanedContent)
	}

	s.iLog.Info(fmt.Sprintf("Successfully parsed AI response - SQL generated: %v, Confidence: %.2f", result.SQL != "", result.Confidence))
	s.iLog.Debug(fmt.Sprintf("Generated SQL: %s", result.SQL))
	s.iLog.Info("generateSQLWithContext COMPLETE")

	return &result, nil
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// executeSQL executes SQL query on the database
func (s *ChatService) executeSQL(databaseAlias, sql string) (map[string]interface{}, int, error) {
	// TODO: Implement actual SQL execution against the specified database
	// This is a placeholder that returns mock data
	// In production, you would:
	// 1. Get database connection from alias
	// 2. Execute query with timeout
	// 3. Convert results to JSON
	// 4. Return data and row count

	return map[string]interface{}{
		"message": "SQL execution not yet implemented - requires database connection pool",
		"sql":     sql,
	}, 0, nil
}

// logAIGeneration logs AI generation for audit purposes
func (s *ChatService) logAIGeneration(conversationID, messageID string, response *Text2SQLResponse) {
	log := &models.AIGenerationLog{
		ID:              uuid.New().String(),
		ConversationID:  &conversationID,
		MessageID:       &messageID,
		GenerationType:  models.GenerationTypeSQL,
		AIResponse:      response.SQL,
		ModelName:       s.OpenAIModel,
		ConfidenceScore: &response.Confidence,
		WasSuccessful:   true,
	}

	s.DB.Create(log) // Ignore errors for logging
}

// CreateEmbedding creates a vector embedding for text using OpenAI
func (s *ChatService) CreateEmbedding(ctx context.Context, text string) ([]float64, error) {
	type EmbeddingRequest struct {
		Input string `json:"input"`
		Model string `json:"model"`
	}

	type EmbeddingResponse struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	reqBody := EmbeddingRequest{
		Input: text,
		Model: "text-embedding-ada-002",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.OpenAIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: status %d", resp.StatusCode)
	}

	var embResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, err
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return embResp.Data[0].Embedding, nil
}

// SaveSchemaEmbedding saves a schema embedding to the database
func (s *ChatService) SaveSchemaEmbedding(databaseAlias string, entityType models.EntityType, entityID, entityText string, embedding []float64) error {
	// Convert embedding to JSONMap
	embeddingJSON := make(map[string]interface{})
	embeddingArray := make([]interface{}, len(embedding))
	for i, v := range embedding {
		embeddingArray[i] = v
	}
	embeddingJSON["vector"] = embeddingArray

	schemaEmbedding := &models.SchemaEmbedding{
		ID:            uuid.New().String(),
		DatabaseAlias: databaseAlias,
		EntityType:    entityType,
		EntityID:      entityID,
		EntityText:    entityText,
		Embedding:     embeddingJSON,
		ModelName:     "text-embedding-ada-002",
	}

	return s.DB.Create(schemaEmbedding).Error
}

// SearchSimilarSchemaElements finds similar schema elements using vector search
func (s *ChatService) SearchSimilarSchemaElements(ctx context.Context, databaseAlias, query string, limit int) ([]models.SchemaEmbedding, error) {
	// 1. Create embedding for query
	queryEmbedding, err := s.CreateEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	// 2. Get all embeddings for the database
	var embeddings []models.SchemaEmbedding
	if err := s.DB.Where("databasealias = ?", databaseAlias).Find(&embeddings).Error; err != nil {
		return nil, err
	}

	// 3. Calculate cosine similarity for each embedding
	type ScoredEmbedding struct {
		Embedding *models.SchemaEmbedding
		Score     float64
	}

	scored := make([]ScoredEmbedding, 0, len(embeddings))
	for i := range embeddings {
		// Extract vector from JSON
		embVector, ok := embeddings[i].Embedding["vector"].([]interface{})
		if !ok {
			continue
		}

		vec := make([]float64, len(embVector))
		for j, v := range embVector {
			if fv, ok := v.(float64); ok {
				vec[j] = fv
			}
		}

		// Calculate cosine similarity
		score := cosineSimilarity(queryEmbedding, vec)
		scored = append(scored, ScoredEmbedding{
			Embedding: &embeddings[i],
			Score:     score,
		})
	}

	// 4. Sort by score descending
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].Score > scored[i].Score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// 5. Return top N results
	if limit > len(scored) {
		limit = len(scored)
	}

	results := make([]models.SchemaEmbedding, limit)
	for i := 0; i < limit; i++ {
		results[i] = *scored[i].Embedding
	}

	return results, nil
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, magnitudeA, magnitudeB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		magnitudeA += a[i] * a[i]
		magnitudeB += b[i] * b[i]
	}

	if magnitudeA == 0 || magnitudeB == 0 {
		return 0.0
	}

	return dotProduct / (sqrt(magnitudeA) * sqrt(magnitudeB))
}

// sqrt calculates square root
func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}

	z := x
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

// cleanJSONResponse removes markdown code blocks and cleans the JSON response
func cleanJSONResponse(content string) string {
	// Trim whitespace
	content = strings.TrimSpace(content)

	// Remove markdown code blocks (```json...``` or ```...```)
	if strings.HasPrefix(content, "```") {
		// Find the first newline after ```
		startIdx := strings.Index(content, "\n")
		if startIdx == -1 {
			// No newline found, try removing just the ```
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimPrefix(content, "```")
		} else {
			// Remove everything up to and including the first newline
			content = content[startIdx+1:]
		}

		// Remove trailing ```
		if strings.HasSuffix(content, "```") {
			content = content[:len(content)-3]
		}

		content = strings.TrimSpace(content)
	}

	return content
}
