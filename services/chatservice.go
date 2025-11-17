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
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// ChatService handles chat and AI conversation features
type ChatService struct {
	DB          *gorm.DB
	OpenAIKey   string
	OpenAIModel string
}

// NewChatService creates a new chat service
func NewChatService(db *gorm.DB, openAIKey, openAIModel string) *ChatService {
	return &ChatService{
		DB:          db,
		OpenAIKey:   openAIKey,
		OpenAIModel: openAIModel,
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
	// 1. Save user message
	userMsg := &models.ChatMessage{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		MessageType:    models.MessageTypeUser,
		Text:           userMessage,
	}

	if err := s.DB.Create(userMsg).Error; err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// 2. Get schema context
	schemaContext, err := s.getSchemaContext(databaseAlias, userMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema context: %w", err)
	}

	// 3. Generate SQL using AI
	sqlResponse, err := s.generateSQLWithContext(ctx, userMessage, schemaContext)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SQL: %w", err)
	}

	// 4. Create assistant response
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
		startTime := time.Now()
		resultData, rowCount, execErr := s.executeSQL(databaseAlias, sqlResponse.SQL)
		executionTime := int(time.Since(startTime).Milliseconds())
		assistantMsg.ExecutionTimeMs = &executionTime

		if execErr != nil {
			assistantMsg.ErrorMessage = execErr.Error()
		} else {
			assistantMsg.ResultData = resultData
			assistantMsg.RowCount = &rowCount
		}
	}

	// 6. Save assistant message
	if err := s.DB.Create(assistantMsg).Error; err != nil {
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	// 7. Log AI generation
	s.logAIGeneration(conversationID, assistantMsg.ID, sqlResponse)

	return assistantMsg, nil
}

// getSchemaContext retrieves relevant schema information using vector search
func (s *ChatService) getSchemaContext(databaseAlias, question string) (string, error) {
	// 1. Get business entities
	businessEntities, err := s.getRelevantBusinessEntities(databaseAlias, question)
	if err != nil {
		return "", err
	}

	// 2. Get table metadata
	tableMetadata, err := s.getRelevantTableMetadata(databaseAlias, question)
	if err != nil {
		return "", err
	}

	// 3. Build context string
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
	}

	if len(tableMetadata) > 0 {
		contextBuilder.WriteString("TABLES AND COLUMNS:\n")
		for _, meta := range tableMetadata {
			if meta.MetadataType == models.MetadataTypeTable {
				contextBuilder.WriteString(fmt.Sprintf("Table: %s - %s\n", meta.Table, meta.Description))
			} else {
				contextBuilder.WriteString(fmt.Sprintf("  - %s.%s (%s): %s\n",
					meta.Table, meta.Column, meta.DataType, meta.ColumnComment))
			}
		}
	}

	return contextBuilder.String(), nil
}

// getRelevantBusinessEntities finds business entities relevant to the question
func (s *ChatService) getRelevantBusinessEntities(databaseAlias, question string) ([]models.BusinessEntity, error) {
	var entities []models.BusinessEntity

	// Use full-text search for now (vector search would be better)
	err := s.DB.Where("databasealias = ?", databaseAlias).
		Where("MATCH(entity_name, description) AGAINST(? IN NATURAL LANGUAGE MODE)", question).
		Limit(5).
		Find(&entities).Error

	return entities, err
}

// getRelevantTableMetadata finds table/column metadata relevant to the question
func (s *ChatService) getRelevantTableMetadata(databaseAlias, question string) ([]models.DatabaseSchemaMetadata, error) {
	var metadata []models.DatabaseSchemaMetadata

	// Use full-text search for now
	err := s.DB.Where("databasealias = ?", databaseAlias).
		Where("MATCH(description, column_comment) AGAINST(? IN NATURAL LANGUAGE MODE)", question).
		Limit(10).
		Find(&metadata).Error

	return metadata, err
}

// generateSQLWithContext generates SQL using AI with schema context
func (s *ChatService) generateSQLWithContext(ctx context.Context, question, schemaContext string) (*Text2SQLResponse, error) {
	systemPrompt := `You are an expert SQL query generator that converts natural language questions into accurate SQL queries.

CORE PRINCIPLES:
1. Generate syntactically correct SQL for MySQL
2. Use proper table and column names from the provided schema
3. Apply appropriate filters, joins, and aggregations
4. Ensure queries are efficient and follow best practices
5. Provide clear explanations for your reasoning

RESPONSE FORMAT (JSON):
{
  "sql": "The generated SQL query",
  "explanation": "Clear explanation of what the query does",
  "confidence": 0.95,
  "tables_used": ["table1", "table2"],
  "columns_used": ["column1", "column2"],
  "reasoning": "Step-by-step reasoning process",
  "query_type": "SELECT"
}

IMPORTANT:
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
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.OpenAIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: status %d", resp.StatusCode)
	}

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	// Parse SQL response
	var result Text2SQLResponse
	cleanedContent := cleanJSONResponse(openAIResp.Choices[0].Message.Content)
	if err := json.Unmarshal([]byte(cleanedContent), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w (content: %s)", err, cleanedContent)
	}

	return &result, nil
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
