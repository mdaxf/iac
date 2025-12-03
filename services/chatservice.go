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
	"github.com/mdaxf/iac/databases/orm"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	openai "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

// SchemaDiscoveryService is an interface for services that can discover database schema
type SchemaDiscoveryService interface {
	GetDatabaseMetadata(ctx context.Context, databaseAlias string) ([]models.DatabaseSchemaMetadata, error)
}

// ChatService handles chat and AI conversation features
type ChatService struct {
	DB                     *gorm.DB
	OpenAIKey              string
	OpenAIModel            string
	SchemaMetadataService  SchemaDiscoveryService
	SchemaEmbeddingService *SchemaEmbeddingService
	iLog                   logger.Log
}

// NewChatService creates a new chat service
func NewChatService(db *gorm.DB, openAIKey, openAIModel string, schemaService SchemaDiscoveryService) *ChatService {
	embeddingService := NewSchemaEmbeddingService(db, openAIKey)
	return &ChatService{
		DB:                     db,
		OpenAIKey:              openAIKey,
		OpenAIModel:            openAIModel,
		SchemaMetadataService:  schemaService,
		SchemaEmbeddingService: embeddingService,
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

// GetMessage retrieves a single message by ID
func (s *ChatService) GetMessage(messageID string) (*models.ChatMessage, error) {
	var message models.ChatMessage
	if err := s.DB.First(&message, "id = ?", messageID).Error; err != nil {
		return nil, fmt.Errorf("message not found: %w", err)
	}
	return &message, nil
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

	// 3. Create assistant response placeholder immediately
	s.iLog.Debug("Step 3: Creating assistant response placeholder")
	assistantMsg := &models.ChatMessage{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		MessageType:    models.MessageTypeAssistant,
		Text:           "Generating SQL query...",
		Provenance: map[string]interface{}{
			"sql_status": "generating",
		},
	}

	// 4. Save assistant message placeholder
	s.iLog.Debug("Step 4: Saving assistant message placeholder to database")
	if err := s.DB.Create(assistantMsg).Error; err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to save assistant message: %v", err))
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}
	s.iLog.Debug(fmt.Sprintf("Assistant message placeholder saved with ID: %s", assistantMsg.ID))

	// 5. Launch async SQL generation and execution
	s.iLog.Info("Step 5: Launching async SQL generation and execution")
	go s.generateAndExecuteSQL(conversationID, assistantMsg.ID, userMessage, databaseAlias, schemaContext, autoExecute)

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
			// TODO: CalculationFormula field removed from BusinessEntity model
			// if entity.CalculationFormula != "" {
			// 	contextBuilder.WriteString(fmt.Sprintf("  Formula: %s\n", entity.CalculationFormula))
			// }
		}
		contextBuilder.WriteString("\n")
	} else {
		s.iLog.Warn("No business entities found for context")
	}

	if len(tableMetadata) > 0 {
		// Count tables vs columns
		tableCount := 0
		columnCount := 0
		tableNames := make(map[string]bool) // Track unique table names
		for _, meta := range tableMetadata {
			if meta.MetadataType == models.MetadataTypeTable {
				tableCount++
				tableNames[meta.Table] = true
			} else {
				columnCount++
			}
		}
		s.iLog.Info(fmt.Sprintf("Building context with %d tables and %d columns", tableCount, columnCount))

		contextBuilder.WriteString("TABLES, COLUMNS, AND SAMPLE DATA:\n\n")

		// Group metadata by table for better organization
		tableMap := make(map[string][]models.DatabaseSchemaMetadata)
		for _, meta := range tableMetadata {
			if meta.MetadataType == models.MetadataTypeTable {
				tableMap[meta.Table] = []models.DatabaseSchemaMetadata{meta}
			}
		}
		for _, meta := range tableMetadata {
			if meta.MetadataType == models.MetadataTypeColumn {
				tableMap[meta.Table] = append(tableMap[meta.Table], meta)
			}
		}

		// Build context for each table with columns and sample data
		for tableName, metaList := range tableMap {
			// Table description
			for _, meta := range metaList {
				if meta.MetadataType == models.MetadataTypeTable {
					contextBuilder.WriteString(fmt.Sprintf("Table: %s\n", tableName))
					if meta.Description != "" {
						contextBuilder.WriteString(fmt.Sprintf("Description: %s\n", meta.Description))
					}
					break
				}
			}

			// Columns
			contextBuilder.WriteString("Columns:\n")
			for _, meta := range metaList {
				if meta.MetadataType == models.MetadataTypeColumn {
					contextBuilder.WriteString(fmt.Sprintf("  - %s (%s)", meta.Column, meta.DataType))
					if meta.ColumnComment != "" {
						contextBuilder.WriteString(fmt.Sprintf(": %s", meta.ColumnComment))
					}
					contextBuilder.WriteString("\n")
				}
			}

			// Fetch and include sample data (1-3 records)
			sampleData, err := s.getSampleDataForTable(databaseAlias, tableName)
			if err == nil && len(sampleData) > 0 {
				contextBuilder.WriteString(fmt.Sprintf("Sample Data (%d rows):\n", len(sampleData)))
				sampleJSON, _ := json.MarshalIndent(sampleData, "  ", "  ")
				contextBuilder.WriteString(fmt.Sprintf("  %s\n", string(sampleJSON)))
			} else if err != nil {
				s.iLog.Debug(fmt.Sprintf("Could not fetch sample data for %s: %v", tableName, err))
			}

			contextBuilder.WriteString("\n")
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
	s.iLog.Debug("Attempting vector-based business entity search")

	// Use vector similarity search from vector database
	// Business entities are ONLY stored in vector database, not in main database
	ctx := context.Background()
	entities, err := s.SchemaEmbeddingService.SearchSimilarBusinessEntities(ctx, databaseAlias, question, 5)

	if err != nil {
		s.iLog.Warn(fmt.Sprintf("Vector search failed: %v - returning empty result", err))
		return []models.BusinessEntity{}, nil // Return empty result instead of error
	}

	if len(entities) == 0 {
		s.iLog.Debug("Vector search returned 0 results - no business entities found")
		return []models.BusinessEntity{}, nil
	}

	s.iLog.Info(fmt.Sprintf("Vector search found %d business entities", len(entities)))
	return entities, nil
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

	// Try vector similarity search first (if embeddings exist)
	s.iLog.Debug("Attempting vector-based schema metadata search")
	ctx := context.Background()

	// Search for relevant tables
	tables, err := s.SchemaEmbeddingService.SearchSimilarTables(ctx, databaseAlias, question, 5)
	if err != nil {
		s.iLog.Warn(fmt.Sprintf("Vector table search failed: %v", err))
	} else if len(tables) > 0 {
		s.iLog.Info(fmt.Sprintf("Vector search found %d tables", len(tables)))
		metadata = append(metadata, tables...)
	}

	// Search for relevant columns
	columns, err := s.SchemaEmbeddingService.SearchSimilarColumns(ctx, databaseAlias, question, 10)
	if err != nil {
		s.iLog.Warn(fmt.Sprintf("Vector column search failed: %v", err))
	} else if len(columns) > 0 {
		s.iLog.Info(fmt.Sprintf("Vector search found %d columns", len(columns)))
		metadata = append(metadata, columns...)
	}

	// If vector search returned useful results, use them
	if len(metadata) > 0 {
		metadata = filterValidMetadata(metadata)
		if isMetadataUseful(metadata) {
			s.iLog.Info(fmt.Sprintf("Vector search returned %d useful metadata entries", len(metadata)))
			return metadata, nil
		} else {
			s.iLog.Warn("Vector search returned metadata but it's incomplete, will trigger auto-discovery")
			metadata = []models.DatabaseSchemaMetadata{}
		}
	} else {
		s.iLog.Debug("Vector search returned 0 results, will trigger auto-discovery")
	}

	// Schema metadata is ONLY in vector database, not main database
	// Don't fall back to querying main database - instead use auto-discovery
	// which queries the actual target database's INFORMATION_SCHEMA
	hasUsefulMetadata := false

	// Log the metadata usefulness check
	if len(metadata) > 0 && !hasUsefulMetadata {
		tableCount := 0
		columnCount := 0
		for _, meta := range metadata {
			if meta.MetadataType == models.MetadataTypeTable {
				tableCount++
			} else if meta.MetadataType == models.MetadataTypeColumn {
				columnCount++
			}
		}
		s.iLog.Warn(fmt.Sprintf("Metadata is incomplete - found %d tables and %d columns (need at least 1 table with columns)", tableCount, columnCount))
	} else if len(metadata) > 0 && hasUsefulMetadata {
		tableCount := 0
		columnCount := 0
		for _, meta := range metadata {
			if meta.MetadataType == models.MetadataTypeTable {
				tableCount++
			} else if meta.MetadataType == models.MetadataTypeColumn {
				columnCount++
			}
		}
		s.iLog.Info(fmt.Sprintf("Using metadata with %d tables and %d columns (limited to avoid OpenAI context overflow)", tableCount, columnCount))
	}

	// Trigger auto-discovery if no metadata OR metadata is not useful
	if (!hasUsefulMetadata) && s.SchemaMetadataService != nil {
		if len(metadata) == 0 {
			s.iLog.Warn(fmt.Sprintf("No valid metadata found in schemameta table for alias '%s', triggering auto-discovery", databaseAlias))
		} else {
			s.iLog.Warn("Metadata exists but is incomplete (missing columns), triggering auto-discovery to get full schema")
		}

		// Use auto-discovery fallback from SchemaMetadataService
		ctx := context.Background()
		metadata, err = s.SchemaMetadataService.GetDatabaseMetadata(ctx, databaseAlias)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Auto-discovery failed: %v", err))
			return nil, fmt.Errorf("failed to discover schema metadata: %w", err)
		}
		s.iLog.Info(fmt.Sprintf("Auto-discovery completed, found %d metadata entries", len(metadata)))
	} else if !hasUsefulMetadata && s.SchemaMetadataService == nil {
		s.iLog.Error("Metadata is incomplete and SchemaMetadataService is nil - cannot perform auto-discovery")
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

// isMetadataUseful checks if metadata contains useful schema information
// Returns true if we have both tables AND columns, regardless of count
// This allows using limited metadata (from LIMIT queries) instead of forcing full schema discovery
func isMetadataUseful(metadata []models.DatabaseSchemaMetadata) bool {
	if len(metadata) == 0 {
		return false
	}

	tableCount := 0
	columnCount := 0

	for _, meta := range metadata {
		if meta.MetadataType == models.MetadataTypeTable {
			tableCount++
		} else if meta.MetadataType == models.MetadataTypeColumn {
			columnCount++
		}
	}

	// Need at least some columns to be useful for SQL generation
	if columnCount == 0 {
		return false
	}

	// Need at least one table
	if tableCount == 0 {
		return false
	}

	// As long as we have both tables and columns, it's useful
	// Don't require minimum counts to avoid triggering full schema discovery
	// which can overflow OpenAI's context window
	return true
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

AGGREGATION DETECTION - CRITICAL:
When the question contains words like:
- "by [column]" → Use GROUP BY on that column
- "count", "total", "sum", "average", "max", "min" → Use appropriate aggregate function
- "show X by Y" → SELECT Y, COUNT(*) FROM table GROUP BY Y
- "statistics", "breakdown", "summary" → Use GROUP BY and COUNT/SUM
- "each", "per", "for each" → Use GROUP BY

Examples:
- "show menus by page type" → SELECT pagetype, COUNT(*) as count FROM menus GROUP BY pagetype
- "total sales by category" → SELECT category, SUM(amount) as total FROM sales GROUP BY category
- "average price per product" → SELECT product, AVG(price) as avg_price FROM products GROUP BY product

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
- Add LIMIT clause if not specified (default: 100 for detail queries, no limit for aggregations)
- Use clear aliases for readability
- Pay special attention to "by" keyword - it almost always means GROUP BY`

	userPrompt := fmt.Sprintf(`%s

### QUESTION ###
User's Question: %s

### INSTRUCTIONS ###
1. Analyze the question to understand what data is being requested
2. Check for aggregation keywords ("by", "count", "total", "sum", "average", etc.)
3. If question contains "by [something]", you MUST use GROUP BY on that column
4. Identify the relevant tables and columns from the schema
5. Determine the appropriate joins, filters, and aggregations
6. Generate the SQL query following best practices
7. Provide reasoning for your choices

CRITICAL: If the question says "show X by Y", the SQL MUST include GROUP BY Y and COUNT(*).

Respond with JSON only, no additional text.`, schemaContext, question)

	s.iLog.Debug("Creating OpenAI client using go-openai library")

	// Create OpenAI client using go-openai library for better error handling and type safety
	client := openai.NewClient(s.OpenAIKey)

	s.iLog.Info(fmt.Sprintf("Sending request to OpenAI API (model: %s)", s.OpenAIModel))
	startTime := time.Now()

	// Create chat completion request
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: s.OpenAIModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			Temperature: 0.1,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	elapsed := time.Since(startTime)

	if err != nil {
		s.iLog.Error(fmt.Sprintf("OpenAI API request failed after %v: %v", elapsed, err))
		return nil, fmt.Errorf("OpenAI API request failed: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("OpenAI API responded in %v", elapsed))

	if len(resp.Choices) == 0 {
		s.iLog.Error("OpenAI response contains no choices")
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	s.iLog.Debug(fmt.Sprintf("OpenAI returned %d choices", len(resp.Choices)))

	// Parse SQL response
	var result Text2SQLResponse
	rawContent := resp.Choices[0].Message.Content
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

// getSampleDataForTable fetches 1-3 sample records from a table
func (s *ChatService) getSampleDataForTable(databaseAlias, tableName string) ([]map[string]interface{}, error) {
	s.iLog.Debug(fmt.Sprintf("Fetching sample data for table: %s", tableName))

	// Get database connection
	db, err := orm.GetDB(databaseAlias)
	if err != nil {
		s.iLog.Warn(fmt.Sprintf("Error getting database for sample data: %v", err))
		return nil, err
	}

	// Build safe SELECT query with LIMIT 3
	query := fmt.Sprintf("SELECT * FROM %s LIMIT 3", tableName)
	s.iLog.Debug(fmt.Sprintf("Sample data query: %s", query))

	rows, err := db.Query(query)
	if err != nil {
		s.iLog.Warn(fmt.Sprintf("Error fetching sample data from %s: %v", tableName, err))
		return nil, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Read sample rows
	var sampleData []map[string]interface{}
	for rows.Next() {
		// Create slice for scanning
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			s.iLog.Warn(fmt.Sprintf("Error scanning row: %v", err))
			continue
		}

		// Convert to map
		rowData := make(map[string]interface{})
		for i, colName := range columns {
			val := values[i]
			// Convert byte arrays to strings
			if b, ok := val.([]byte); ok {
				rowData[colName] = string(b)
			} else {
				rowData[colName] = val
			}
		}
		sampleData = append(sampleData, rowData)
	}

	s.iLog.Debug(fmt.Sprintf("Fetched %d sample rows from %s", len(sampleData), tableName))
	return sampleData, nil
}

// executeSQL executes SQL query on the database
func (s *ChatService) executeSQL(databaseAlias, sql string) (map[string]interface{}, int, error) {
	s.iLog.Debug(fmt.Sprintf("executeSQL START - Alias: %s, SQL: %s", databaseAlias, sql))

	// Get database connection for alias
	db, err := orm.GetDB(databaseAlias)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Error getting database for alias '%s': %v", databaseAlias, err))
		return nil, 0, fmt.Errorf("error getting database connection: %v", err)
	}

	// Safety check - only allow SELECT queries
	normalizedSQL := strings.ToUpper(strings.TrimSpace(sql))
	if !strings.HasPrefix(normalizedSQL, "SELECT") {
		s.iLog.Error("Non-SELECT query attempted")
		return nil, 0, fmt.Errorf("only SELECT queries are allowed")
	}

	// Add LIMIT clause if not present (safety measure)
	if !strings.Contains(normalizedSQL, "LIMIT") && !strings.Contains(normalizedSQL, "TOP") {
		sql = fmt.Sprintf("%s LIMIT 1000", sql)
		s.iLog.Debug("Added LIMIT 1000 to query")
	}

	s.iLog.Debug(fmt.Sprintf("Executing SQL: %s", sql))

	// Execute query
	rows, err := db.Query(sql)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Error executing query: %v", err))
		return nil, 0, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Error getting columns: %v", err))
		return nil, 0, fmt.Errorf("error getting columns: %v", err)
	}

	// Read all rows
	var resultRows []map[string]interface{}
	for rows.Next() {
		// Create slice of interface{} to hold each column
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			s.iLog.Error(fmt.Sprintf("Error scanning row: %v", err))
			continue
		}

		// Create a map for this row
		rowData := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			rowData[col] = v
		}
		resultRows = append(resultRows, rowData)
	}

	rowCount := len(resultRows)
	s.iLog.Info(fmt.Sprintf("executeSQL COMPLETE - Returned %d rows", rowCount))

	// Build response with rows array
	result := map[string]interface{}{
		"columns": columns,
		"rows":    resultRows,
	}

	return result, rowCount, nil
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

// DataInsights represents AI-generated insights from query results
type DataInsights struct {
	Summary        string                 `json:"summary"`
	KeyInsights    []string               `json:"key_insights"`
	Visualizations []VisualizationSpec    `json:"visualizations"`
	DataAnalysis   DataStatistics         `json:"data_analysis"`
}

// VisualizationSpec defines a recommended chart/visualization
type VisualizationSpec struct {
	Type        string            `json:"type"`  // bar, line, pie, table, area, scatter
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	XAxis       string            `json:"x_axis,omitempty"`
	YAxis       string            `json:"y_axis,omitempty"`
	YAxisMulti  []string          `json:"y_axis_multi,omitempty"`
	GroupBy     string            `json:"group_by,omitempty"`
	Config      map[string]string `json:"config,omitempty"`
}

// DataStatistics contains basic statistics about the data
type DataStatistics struct {
	RowCount        int                    `json:"row_count"`
	ColumnCount     int                    `json:"column_count"`
	NumericColumns  []string               `json:"numeric_columns"`
	DateColumns     []string               `json:"date_columns"`
	TextColumns     []string               `json:"text_columns"`
	TopValues       map[string]interface{} `json:"top_values,omitempty"`
}

// generateAndUpdateInsights generates insights asynchronously and updates the message in database
func (s *ChatService) generateAndUpdateInsights(messageID, question, sql string, data []map[string]interface{}) {
	iLog := logger.Log{
		ModuleName:     logger.Framework,
		User:           "System",
		ControllerName: "ChatService-Async",
	}
	iLog.Info(fmt.Sprintf("Async insights generation START - MessageID: %s", messageID))

	// Create a context with timeout for the AI call
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Generate insights
	insights, err := s.generateDataInsights(ctx, question, sql, data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to generate insights asynchronously: %v", err))
		// Update message with error status
		s.DB.Model(&models.ChatMessage{}).
			Where("id = ?", messageID).
			Update("provenance", map[string]interface{}{
				"insights_status": "error",
				"insights_error":  err.Error(),
			})
		return
	}

	if insights == nil {
		iLog.Warn("No insights generated")
		return
	}

	iLog.Info(fmt.Sprintf("Generated %d insights and %d visualizations asynchronously",
		len(insights.KeyInsights), len(insights.Visualizations)))

	// Get the current message
	var message models.ChatMessage
	if err := s.DB.First(&message, "id = ?", messageID).Error; err != nil {
		iLog.Error(fmt.Sprintf("Failed to fetch message for update: %v", err))
		return
	}

	// Update provenance with insights
	if message.Provenance == nil {
		message.Provenance = make(map[string]interface{})
	}
	message.Provenance["ai_insights"] = insights.KeyInsights
	message.Provenance["visualizations"] = insights.Visualizations
	message.Provenance["report_summary"] = insights.Summary
	message.Provenance["data_analysis"] = insights.DataAnalysis
	message.Provenance["insights_status"] = "completed"

	// Enhance text with insights
	if len(insights.KeyInsights) > 0 {
		insightsText := "\n\n📊 **Key Insights:**\n"
		for i, insight := range insights.KeyInsights {
			insightsText += fmt.Sprintf("%d. %s\n", i+1, insight)
		}
		message.Text += insightsText
	}

	// Debug: Log what we're about to save
	iLog.Info(fmt.Sprintf("About to save insights - ai_insights count: %d, visualizations count: %d",
		len(insights.KeyInsights), len(insights.Visualizations)))

	// Marshal provenance to JSON to check structure
	provenanceJSON, err := json.Marshal(message.Provenance)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to marshal provenance: %v", err))
		return
	}
	iLog.Debug(fmt.Sprintf("Provenance JSON: %s", string(provenanceJSON)))

	// Save updated message using Update with JSON string to ensure proper serialization
	if err := s.DB.Model(&models.ChatMessage{}).Where("id = ?", messageID).Updates(map[string]interface{}{
		"text":       message.Text,
		"provenance": provenanceJSON, // Pass JSON bytes instead of map
	}).Error; err != nil {
		iLog.Error(fmt.Sprintf("Failed to update message with insights: %v", err))
		return
	}

	iLog.Info(fmt.Sprintf("Async insights generation COMPLETE - MessageID: %s", messageID))
}

// generateAndExecuteSQL generates SQL and executes it asynchronously, then updates the message
func (s *ChatService) generateAndExecuteSQL(conversationID, messageID, userMessage, databaseAlias, schemaContext string, autoExecute bool) {
	iLog := logger.Log{
		ModuleName:     logger.Framework,
		User:           "System",
		ControllerName: "ChatService-AsyncSQL",
	}
	iLog.Info(fmt.Sprintf("Async SQL generation START - MessageID: %s", messageID))

	// Create a context with timeout for SQL generation
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get the message to update
	var message models.ChatMessage
	if err := s.DB.First(&message, "id = ?", messageID).Error; err != nil {
		iLog.Error(fmt.Sprintf("Failed to fetch message: %v", err))
		return
	}

	// Update status to generating SQL
	if message.Provenance == nil {
		message.Provenance = make(map[string]interface{})
	}
	message.Provenance["sql_status"] = "generating"
	s.DB.Save(&message)

	// Generate SQL
	iLog.Info("Generating SQL query using AI")
	sqlResponse, err := s.generateSQLWithContext(ctx, userMessage, schemaContext)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to generate SQL: %v", err))
		message.Provenance["sql_status"] = "error"
		message.Provenance["sql_error"] = err.Error()
		message.Text = "Failed to generate SQL query: " + err.Error()
		s.DB.Save(&message)
		return
	}

	iLog.Info(fmt.Sprintf("SQL generated successfully, confidence: %.2f", sqlResponse.Confidence))

	// Update message with SQL
	confidence := sqlResponse.Confidence
	message.SQLQuery = sqlResponse.SQL
	message.SQLConfidence = &confidence
	message.Text = sqlResponse.Explanation
	message.Provenance["tables_used"] = sqlResponse.TablesUsed
	message.Provenance["columns_used"] = sqlResponse.ColumnsUsed
	message.Provenance["reasoning"] = sqlResponse.Reasoning
	message.Provenance["query_type"] = sqlResponse.QueryType
	message.Provenance["sql_status"] = "completed"

	// Save SQL generation results
	if err := s.DB.Save(&message).Error; err != nil {
		iLog.Error(fmt.Sprintf("Failed to update message with SQL: %v", err))
		return
	}

	// Log AI generation
	s.logAIGeneration(conversationID, messageID, sqlResponse)

	// Execute SQL if auto-execute is enabled
	if autoExecute && sqlResponse.SQL != "" {
		iLog.Info("Auto-executing SQL query")

		// Update status to executing
		message.Provenance["execution_status"] = "executing"
		s.DB.Save(&message)

		startTime := time.Now()
		resultData, rowCount, execErr := s.executeSQL(databaseAlias, sqlResponse.SQL)
		executionTime := int(time.Since(startTime).Milliseconds())
		message.ExecutionTimeMs = &executionTime

		if execErr != nil {
			iLog.Error(fmt.Sprintf("SQL execution failed: %v", execErr))
			message.ErrorMessage = execErr.Error()
			message.Provenance["execution_status"] = "error"
		} else {
			iLog.Info(fmt.Sprintf("SQL executed successfully, returned %d rows in %dms", rowCount, executionTime))
			message.ResultData = resultData
			message.RowCount = &rowCount
			message.Provenance["execution_status"] = "completed"

			// Schedule async insights generation if data was returned
			if rowCount > 0 && resultData != nil {
				iLog.Info("Scheduling async AI BI report generation")

				// Convert resultData to []map[string]interface{} format
				var dataArray []map[string]interface{}
				if rows, ok := resultData["rows"].([]interface{}); ok {
					for _, row := range rows {
						if rowMap, ok := row.(map[string]interface{}); ok {
							dataArray = append(dataArray, rowMap)
						}
					}
				} else if len(resultData) > 0 {
					dataArray = []map[string]interface{}{resultData}
				}

				if len(dataArray) > 0 {
					message.Provenance["insights_status"] = "generating"
					// Launch another goroutine for insights
					go s.generateAndUpdateInsights(messageID, userMessage, sqlResponse.SQL, dataArray)
				}
			}
		}

		// Save execution results
		if err := s.DB.Save(&message).Error; err != nil {
			iLog.Error(fmt.Sprintf("Failed to update message with execution results: %v", err))
			return
		}
	}

	iLog.Info(fmt.Sprintf("Async SQL generation and execution COMPLETE - MessageID: %s", messageID))
}

// generateDataInsights generates AI-powered insights and visualization recommendations
func (s *ChatService) generateDataInsights(ctx context.Context, question, sql string, data []map[string]interface{}) (*DataInsights, error) {
	if len(data) == 0 {
		return nil, nil
	}

	// Analyze data structure
	stats := s.analyzeDataStructure(data)

	// Build AI prompt
	sampleData, _ := json.MarshalIndent(data[:min(3, len(data))], "", "  ")

	systemPrompt := `You are an expert data analyst and business intelligence specialist. Analyze query results and provide actionable insights with visualization recommendations.

RESPONSE FORMAT (JSON only, no markdown):
{
  "summary": "Brief 1-2 sentence summary of what the data shows",
  "key_insights": ["Insight 1", "Insight 2", "Insight 3"],
  "visualizations": [
    {
      "type": "bar|line|pie|area|scatter|table",
      "title": "Chart Title",
      "description": "What this shows",
      "x_axis": "column_name",
      "y_axis": "column_name",
      "group_by": "column_name"
    }
  ]
}

Guidelines:
- Focus on trends, patterns, and anomalies
- Recommend 1-3 most useful visualizations
- Keep insights concise and actionable
- Choose appropriate chart types for the data
- Use actual column names from the data`

	userPrompt := fmt.Sprintf(`Question: %s

SQL: %s

Data Summary:
- Rows: %d
- Numeric columns: %v
- Date columns: %v
- Text columns: %v

Sample Data (first 3 rows):
%s

Analyze this data and provide insights with visualization recommendations.`,
		question, sql, stats.RowCount, stats.NumericColumns, stats.DateColumns, stats.TextColumns, string(sampleData))

	// Create OpenAI client using go-openai library
	client := openai.NewClient(s.OpenAIKey)

	// Create chat completion request
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: s.OpenAIModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			Temperature: 0.3, // Lower temperature for more focused analysis
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	// Parse insights
	content := cleanJSONResponse(resp.Choices[0].Message.Content)
	var insights DataInsights
	if err := json.Unmarshal([]byte(content), &insights); err != nil {
		s.iLog.Warn(fmt.Sprintf("Failed to parse insights JSON: %v", err))
		return nil, err
	}

	insights.DataAnalysis = stats
	return &insights, nil
}

// analyzeDataStructure analyzes the structure and types of data
func (s *ChatService) analyzeDataStructure(data []map[string]interface{}) DataStatistics {
	stats := DataStatistics{
		RowCount:       len(data),
		NumericColumns: []string{},
		DateColumns:    []string{},
		TextColumns:    []string{},
	}

	if len(data) == 0 {
		return stats
	}

	// Analyze first row to determine column types
	firstRow := data[0]
	stats.ColumnCount = len(firstRow)

	for colName, val := range firstRow {
		if val == nil {
			continue
		}

		switch v := val.(type) {
		case int, int32, int64, float32, float64:
			stats.NumericColumns = append(stats.NumericColumns, colName)
		case string:
			// Check if it looks like a date
			if s.looksLikeDate(v) {
				stats.DateColumns = append(stats.DateColumns, colName)
			} else {
				stats.TextColumns = append(stats.TextColumns, colName)
			}
		case time.Time:
			stats.DateColumns = append(stats.DateColumns, colName)
		default:
			stats.TextColumns = append(stats.TextColumns, colName)
		}
	}

	return stats
}

// looksLikeDate checks if a string looks like a date
func (s *ChatService) looksLikeDate(str string) bool {
	dateFormats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"2006-01-02T15:04:05Z",
	}

	for _, format := range dateFormats {
		if _, err := time.Parse(format, str); err == nil {
			return true
		}
	}
	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ExecuteMessageSQL executes or re-executes SQL for a specific message
func (s *ChatService) ExecuteMessageSQL(ctx context.Context, messageID string, modifiedSQL *string) (*models.ChatMessage, error) {
	s.iLog.Info(fmt.Sprintf("ExecuteMessageSQL START - MessageID: %s, HasModifiedSQL: %v", messageID, modifiedSQL != nil))

	// 1. Get the existing message
	var message models.ChatMessage
	if err := s.DB.First(&message, "id = ?", messageID).Error; err != nil {
		s.iLog.Error(fmt.Sprintf("Message not found: %v", err))
		return nil, fmt.Errorf("message not found: %w", err)
	}

	// 2. Get conversation to find database alias
	var conversation models.Conversation
	if err := s.DB.First(&conversation, "id = ?", message.ConversationID).Error; err != nil {
		s.iLog.Error(fmt.Sprintf("Conversation not found: %v", err))
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	// Use default database alias if not set in conversation
	databaseAlias := conversation.DatabaseAlias
	if databaseAlias == "" {
		databaseAlias = "default"
		s.iLog.Info("Using default database alias (conversation has empty alias)")
	}

	// 3. Determine which SQL to execute
	sqlToExecute := message.SQLQuery
	if modifiedSQL != nil && *modifiedSQL != "" {
		sqlToExecute = *modifiedSQL
		s.iLog.Info("Using modified SQL from request")
	}

	if sqlToExecute == "" {
		return nil, fmt.Errorf("no SQL available to execute")
	}

	// 4. Execute the SQL
	s.iLog.Info(fmt.Sprintf("Executing SQL on database '%s': %s", databaseAlias, sqlToExecute))
	startTime := time.Now()
	resultData, rowCount, execErr := s.executeSQL(databaseAlias, sqlToExecute)
	executionTime := int(time.Since(startTime).Milliseconds())

	// 5. Update the message with new results
	message.ExecutionTimeMs = &executionTime
	if modifiedSQL != nil && *modifiedSQL != "" {
		message.SQLQuery = *modifiedSQL // Update with modified SQL
	}

	if execErr != nil {
		s.iLog.Error(fmt.Sprintf("SQL execution failed: %v", execErr))
		message.ErrorMessage = execErr.Error()
		message.ResultData = nil
		message.RowCount = nil
	} else {
		s.iLog.Info(fmt.Sprintf("SQL executed successfully, returned %d rows in %dms", rowCount, executionTime))
		message.ResultData = resultData
		message.RowCount = &rowCount
		message.ErrorMessage = "" // Clear any previous error

		// 5a. Generate AI BI Report if data was returned
		if rowCount > 0 && resultData != nil {
			s.iLog.Info("Generating AI BI report from execution results")

			// Convert resultData to []map[string]interface{} format
			var dataArray []map[string]interface{}
			if rows, ok := resultData["rows"].([]interface{}); ok {
				for _, row := range rows {
					if rowMap, ok := row.(map[string]interface{}); ok {
						dataArray = append(dataArray, rowMap)
					}
				}
			} else if len(resultData) > 0 {
				dataArray = []map[string]interface{}{resultData}
			}

			if len(dataArray) > 0 {
				// Get the original user question from the conversation
				userQuestion := message.Text // Fallback
				var userMsg models.ChatMessage
				if err := s.DB.Where("conversationid = ? AND messagetype = ?", message.ConversationID, models.MessageTypeUser).
					Order("createdon DESC").First(&userMsg).Error; err == nil {
					userQuestion = userMsg.Text
				}

				// Create a new context with longer timeout for AI insights (60 seconds)
				insightsCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				reportInsights, reportErr := s.generateDataInsights(insightsCtx, userQuestion, sqlToExecute, dataArray)
				if reportErr != nil {
					s.iLog.Warn(fmt.Sprintf("Failed to generate data insights: %v", reportErr))
				} else if reportInsights != nil {
					s.iLog.Info(fmt.Sprintf("Generated %d insights and %d visualizations",
						len(reportInsights.KeyInsights), len(reportInsights.Visualizations)))

					// Add insights to provenance
					if message.Provenance == nil {
						message.Provenance = make(map[string]interface{})
					}
					message.Provenance["ai_insights"] = reportInsights.KeyInsights
					message.Provenance["visualizations"] = reportInsights.Visualizations
					message.Provenance["report_summary"] = reportInsights.Summary
					message.Provenance["data_analysis"] = reportInsights.DataAnalysis

					// Enhance explanation with insights
					if len(reportInsights.KeyInsights) > 0 {
						insightsText := "\n\n📊 **Key Insights:**\n"
						for i, insight := range reportInsights.KeyInsights {
							insightsText += fmt.Sprintf("%d. %s\n", i+1, insight)
						}
						message.Text += insightsText
					}
				}
			}
		}
	}

	// 6. Save updated message
	if err := s.DB.Save(&message).Error; err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to update message: %v", err))
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("ExecuteMessageSQL COMPLETE - MessageID: %s", messageID))
	return &message, nil
}
