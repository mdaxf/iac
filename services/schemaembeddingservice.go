package services

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// SchemaEmbeddingService handles vector embedding generation and storage for schema elements
type SchemaEmbeddingService struct {
	DB                    *gorm.DB
	VectorDB              *gorm.DB // Separate vector database connection
	OpenAIKey             string
	ModelName             string
	ConfigID              int // Default config ID for embeddings in vector tables
	SchemaMetadataService *SchemaMetadataService
	iLog                  logger.Log
}

// NewSchemaEmbeddingService creates a new schema embedding service
func NewSchemaEmbeddingService(db *gorm.DB, openAIKey string) *SchemaEmbeddingService {
	service := &SchemaEmbeddingService{
		DB:                    db,
		OpenAIKey:             openAIKey,
		ModelName:             "text-embedding-ada-002",
		SchemaMetadataService: NewSchemaMetadataService(db),
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "SchemaEmbeddingService",
		},
	}

	// Get vector database connection based on aiconfig.json
	vectorDB, err := GetVectorDB(db)
	if err != nil {
		service.iLog.Error(fmt.Sprintf("Failed to get vector database connection: %v, using main DB", err))
		service.VectorDB = db
	} else {
		service.VectorDB = vectorDB
		service.iLog.Info("Using configured vector database for embeddings")
	}

	// Get or create default config
	service.ConfigID = service.getOrCreateDefaultConfig()

	return service
}

// getOrCreateDefaultConfig gets or creates a default embedding configuration
func (s *SchemaEmbeddingService) getOrCreateDefaultConfig() int {
	var config models.AIEmbeddingConfiguration

	// Try to find existing default config in vector database
	err := s.VectorDB.Where("config_name = ?", "default").First(&config).Error
	if err == nil {
		s.iLog.Info(fmt.Sprintf("Using existing embedding config ID: %d", config.ID))
		return config.ID
	}

	// Create new default config in vector database
	config = models.AIEmbeddingConfiguration{
		ConfigName:          "default",
		EmbeddingModel:      s.ModelName,
		EmbeddingDimensions: 1536,
		VectorDatabaseType:  "postgresql",
		Active:              true,
		CreatedBy:           "System",
	}

	if err := s.VectorDB.Create(&config).Error; err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to create default config: %v, using ID 1 as fallback", err))
		return 1
	}

	s.iLog.Info(fmt.Sprintf("Created new embedding config ID: %d in vector database", config.ID))
	return config.ID
}

// EmbeddingRequest represents OpenAI embedding API request
type EmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// EmbeddingResponse represents OpenAI embedding API response
type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// GenerateEmbedding generates a vector embedding for text using OpenAI
func (s *SchemaEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	s.iLog.Debug(fmt.Sprintf("Generating embedding for text (length: %d)", len(text)))

	// Truncate text if too long (max ~8000 tokens for ada-002)
	if len(text) > 8000 {
		text = text[:8000]
		s.iLog.Warn("Text truncated to 8000 characters for embedding generation")
	}

	reqBody := EmbeddingRequest{
		Input: text,
		Model: s.ModelName,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to marshal embedding request: %v", err))
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonBody))
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to create HTTP request: %v", err))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.OpenAIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	startTime := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(startTime)

	if err != nil {
		s.iLog.Error(fmt.Sprintf("OpenAI embedding API request failed after %v: %v", elapsed, err))
		return nil, err
	}
	defer resp.Body.Close()

	s.iLog.Debug(fmt.Sprintf("OpenAI embedding API responded in %v with status: %d", elapsed, resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		// Read the error response body to get detailed error message
		bodyBytes, readErr := io.ReadAll(resp.Body)
		errorDetail := "unable to read error response"
		if readErr == nil {
			errorDetail = string(bodyBytes)
		}
		s.iLog.Error(fmt.Sprintf("OpenAI embedding API returned error status %d: %s", resp.StatusCode, errorDetail))
		return nil, fmt.Errorf("OpenAI embedding API error: status %d - %s", resp.StatusCode, errorDetail)
	}

	var embResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to decode embedding response: %v", err))
		return nil, err
	}

	if len(embResp.Data) == 0 {
		s.iLog.Error("No embedding data in response")
		return nil, fmt.Errorf("no embedding data in response")
	}

	s.iLog.Debug(fmt.Sprintf("Embedding generated successfully, dimension: %d, tokens used: %d",
		len(embResp.Data[0].Embedding), embResp.Usage.TotalTokens))

	return embResp.Data[0].Embedding, nil
}

// GenerateTableEmbedding generates and stores embedding for a table metadata entry
func (s *SchemaEmbeddingService) GenerateTableEmbedding(ctx context.Context, metadata *models.DatabaseSchemaMetadata) error {
	s.iLog.Info(fmt.Sprintf("Generating embedding for table: %s.%s", metadata.DatabaseAlias, metadata.Table))

	// Build text representation for embedding
	var textBuilder strings.Builder
	textBuilder.WriteString(fmt.Sprintf("Table: %s\n", metadata.Table))
	if metadata.SchemaName != "" {
		textBuilder.WriteString(fmt.Sprintf("Schema: %s\n", metadata.SchemaName))
	}
	if metadata.Description != "" {
		textBuilder.WriteString(fmt.Sprintf("Description: %s\n", metadata.Description))
	}
	// Include business terms if available
	if len(metadata.BusinessTerms) > 0 {
		if terms, ok := metadata.BusinessTerms["terms"].([]interface{}); ok {
			textBuilder.WriteString("Business Terms: ")
			for i, term := range terms {
				if i > 0 {
					textBuilder.WriteString(", ")
				}
				textBuilder.WriteString(fmt.Sprintf("%v", term))
			}
			textBuilder.WriteString("\n")
		}
	}

	text := textBuilder.String()
	s.iLog.Debug(fmt.Sprintf("Table embedding text: %s", text))

	// Generate embedding
	embedding, err := s.GenerateEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert to JSON for MySQL VECTOR storage
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	// Update the metadata record with embedding
	now := time.Now()
	err = s.VectorDB.Model(metadata).Updates(map[string]interface{}{
		"embedding":              gorm.Expr("CAST(? AS JSON)", string(embeddingJSON)),
		"embedding_model":        s.ModelName,
		"embedding_generated_at": now,
	}).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to save table embedding: %v", err))
		return fmt.Errorf("failed to save embedding: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Table embedding saved successfully for %s", metadata.Table))
	return nil
}

// GenerateColumnEmbedding generates and stores embedding for a column metadata entry
func (s *SchemaEmbeddingService) GenerateColumnEmbedding(ctx context.Context, metadata *models.DatabaseSchemaMetadata) error {
	s.iLog.Info(fmt.Sprintf("Generating embedding for column: %s.%s.%s", metadata.DatabaseAlias, metadata.Table, metadata.Column))

	// Build text representation for embedding
	var textBuilder strings.Builder
	textBuilder.WriteString(fmt.Sprintf("Table: %s\n", metadata.Table))
	textBuilder.WriteString(fmt.Sprintf("Column: %s\n", metadata.Column))
	if metadata.DataType != "" {
		textBuilder.WriteString(fmt.Sprintf("Data Type: %s\n", metadata.DataType))
	}
	if metadata.ColumnComment != "" {
		textBuilder.WriteString(fmt.Sprintf("Comment: %s\n", metadata.ColumnComment))
	}
	if metadata.Description != "" {
		textBuilder.WriteString(fmt.Sprintf("Description: %s\n", metadata.Description))
	}
	// Include sample values if available
	if len(metadata.SampleValues) > 0 {
		if samples, ok := metadata.SampleValues["values"].([]interface{}); ok && len(samples) > 0 {
			textBuilder.WriteString("Sample Values: ")
			for i, sample := range samples {
				if i > 0 && i < 5 { // Limit to 5 samples
					textBuilder.WriteString(", ")
				}
				if i >= 5 {
					break
				}
				textBuilder.WriteString(fmt.Sprintf("%v", sample))
			}
			textBuilder.WriteString("\n")
		}
	}

	text := textBuilder.String()
	s.iLog.Debug(fmt.Sprintf("Column embedding text: %s", text))

	// Generate embedding
	embedding, err := s.GenerateEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert to JSON for MySQL VECTOR storage
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	// Update the metadata record with embedding
	now := time.Now()
	err = s.VectorDB.Model(metadata).Updates(map[string]interface{}{
		"embedding":              gorm.Expr("CAST(? AS JSON)", string(embeddingJSON)),
		"embedding_model":        s.ModelName,
		"embedding_generated_at": now,
	}).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to save column embedding: %v", err))
		return fmt.Errorf("failed to save embedding: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Column embedding saved successfully for %s.%s", metadata.Table, metadata.Column))
	return nil
}

// GenerateBusinessEntityEmbedding generates and stores embedding for a business entity
func (s *SchemaEmbeddingService) GenerateBusinessEntityEmbedding(ctx context.Context, entity *models.BusinessEntity) error {
	s.iLog.Info(fmt.Sprintf("Generating embedding for business entity: %s", entity.EntityName))

	// Build text representation for embedding
	var textBuilder strings.Builder
	textBuilder.WriteString(fmt.Sprintf("Entity: %s\n", entity.EntityName))
	textBuilder.WriteString(fmt.Sprintf("Type: %s\n", entity.EntityType))
	if entity.Description != "" {
		textBuilder.WriteString(fmt.Sprintf("Description: %s\n", entity.Description))
	}
	// TODO: CalculationFormula field removed from BusinessEntity model
	// if entity.CalculationFormula != "" {
	// 	textBuilder.WriteString(fmt.Sprintf("Formula: %s\n", entity.CalculationFormula))
	// }
	// TODO: Synonyms field removed from BusinessEntity model
	// Include synonyms if available
	// if entity.Synonyms != nil {
	// 	var synonyms []string
	// 	if err := json.Unmarshal(entity.Synonyms, &synonyms); err == nil && len(synonyms) > 0 {
	// 		textBuilder.WriteString(fmt.Sprintf("Synonyms: %s\n", strings.Join(synonyms, ", ")))
	// 	}
	// }
	// TODO: Examples field removed from BusinessEntity model
	// Include examples if available
	// if entity.Examples != nil {
	// 	var examples []string
	// 	if err := json.Unmarshal(entity.Examples, &examples); err == nil && len(examples) > 0 {
	// 		textBuilder.WriteString(fmt.Sprintf("Examples: %s\n", strings.Join(examples, ", ")))
	// 	}
	// }

	text := textBuilder.String()
	s.iLog.Debug(fmt.Sprintf("Business entity embedding text: %s", text))

	// Generate embedding
	embedding, err := s.GenerateEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert to JSON for MySQL VECTOR storage
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	// Update the entity record with embedding
	now := time.Now()
	err = s.VectorDB.Model(entity).Updates(map[string]interface{}{
		"embedding":              gorm.Expr("CAST(? AS JSON)", string(embeddingJSON)),
		"embedding_model":        s.ModelName,
		"embedding_generated_at": now,
	}).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to save business entity embedding: %v", err))
		return fmt.Errorf("failed to save embedding: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Business entity embedding saved successfully for %s", entity.EntityName))
	return nil
}

// GenerateEmbeddingsForDatabase generates embeddings for all schema metadata in a database
// Stores embeddings in the database_schema_embeddings vector table
func (s *SchemaEmbeddingService) GenerateEmbeddingsForDatabase(ctx context.Context, databaseAlias string) error {
	s.iLog.Info(fmt.Sprintf("Starting batch embedding generation for database: %s", databaseAlias))

	// STEP 0: First discover the schema from the actual database
	// This populates databaseschemametadata table with current schema
	s.iLog.Info(fmt.Sprintf("Discovering schema for database alias: %s", databaseAlias))

	// For MySQL/PostgreSQL, the database name is often the same as the alias
	// In a production system, you'd look up the actual database name from configuration
	dbName := databaseAlias

	err := s.SchemaMetadataService.DiscoverDatabaseSchema(ctx, databaseAlias, dbName)
	if err != nil {
		s.iLog.Warn(fmt.Sprintf("Schema discovery failed: %v (continuing with existing metadata)", err))
		// Continue anyway - maybe metadata already exists
	} else {
		s.iLog.Info(fmt.Sprintf("Schema discovery completed for database: %s", databaseAlias))
	}

	// STEP 1: Get all tables from databaseschemametadata
	var tableMetadata []models.DatabaseSchemaMetadata
	err = s.DB.Where("databasealias = ? AND metadatatype = ?",
		databaseAlias, models.MetadataTypeTable).Find(&tableMetadata).Error
	if err != nil {
		return fmt.Errorf("failed to fetch table metadata: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Found %d tables to process", len(tableMetadata)))

	if len(tableMetadata) == 0 {
		s.iLog.Warn(fmt.Sprintf("No tables found for database alias '%s'. Schema discovery may have failed or database is empty.", databaseAlias))
		return fmt.Errorf("no tables found for database alias '%s'", databaseAlias)
	}

	// STEP 2: Generate embeddings for tables in batches
	const batchSize = 10 // Process 10 tables at a time to avoid rate limits
	tablesProcessed := 0
	tablesSkipped := 0
	tablesFailed := 0

	for i, meta := range tableMetadata {
		// Check if embedding already exists in database_schema_embeddings table (in vector DB)
		var existingEmbedding models.DatabaseSchemaEmbedding
		err := s.VectorDB.Where("config_id = ? AND database_alias = ? AND table_name = ? AND column_name IS NULL",
			s.ConfigID, databaseAlias, meta.Table).First(&existingEmbedding).Error

		if err == gorm.ErrRecordNotFound {
			// Generate new embedding and store in vector table
			s.iLog.Info(fmt.Sprintf("Generating embedding for table %d/%d: %s", i+1, len(tableMetadata), meta.Table))
			if err := s.GenerateAndStoreTableEmbedding(ctx, databaseAlias, &meta); err != nil {
				s.iLog.Error(fmt.Sprintf("❌ FAILED to generate embedding for table %s: %v", meta.Table, err))
				// Log first failure with more details for debugging
				if tablesFailed == 0 {
					s.iLog.Error(fmt.Sprintf("First failure details - OpenAI Key present: %v, Model: %s", s.OpenAIKey != "", s.ModelName))
				}
				tablesFailed++
				continue
			}
			tablesProcessed++

			// Rate limiting: wait between API calls
			time.Sleep(100 * time.Millisecond)

			// Add longer pause every batch to respect rate limits
			if (i+1)%batchSize == 0 {
				s.iLog.Info(fmt.Sprintf("Processed batch of %d tables, pausing for 2 seconds...", batchSize))
				time.Sleep(2 * time.Second)
			}
		} else if err != nil {
			s.iLog.Error(fmt.Sprintf("Error checking existing embedding for table %s: %v", meta.Table, err))
			tablesFailed++
			continue
		} else {
			s.iLog.Debug(fmt.Sprintf("Table %s already has embedding (ID: %d), skipping", meta.Table, existingEmbedding.ID))
			tablesSkipped++
		}
	}

	s.iLog.Info(fmt.Sprintf("Table embeddings: %d processed, %d skipped, %d failed", tablesProcessed, tablesSkipped, tablesFailed))

	// STEP 3: Get all columns from databaseschemametadata
	var columnMetadata []models.DatabaseSchemaMetadata
	err = s.DB.Where("databasealias = ? AND metadatatype = ?",
		databaseAlias, models.MetadataTypeColumn).Find(&columnMetadata).Error
	if err != nil {
		return fmt.Errorf("failed to fetch column metadata: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Found %d columns to process", len(columnMetadata)))

	// STEP 4: Generate embeddings for columns in batches
	columnsProcessed := 0
	columnsSkipped := 0
	columnsFailed := 0

	for i, meta := range columnMetadata {
		// Check if embedding already exists (in vector DB)
		var existingEmbedding models.DatabaseSchemaEmbedding
		err := s.VectorDB.Where("config_id = ? AND database_alias = ? AND table_name = ? AND column_name = ?",
			s.ConfigID, databaseAlias, meta.Table, meta.Column).First(&existingEmbedding).Error

		if err == gorm.ErrRecordNotFound {
			s.iLog.Info(fmt.Sprintf("Generating embedding for column %d/%d: %s.%s", i+1, len(columnMetadata), meta.Table, meta.Column))
			if err := s.GenerateAndStoreColumnEmbedding(ctx, databaseAlias, &meta); err != nil {
				s.iLog.Error(fmt.Sprintf("Failed to generate embedding for column %s.%s: %v", meta.Table, meta.Column, err))
				columnsFailed++
				continue
			}
			columnsProcessed++

			// Rate limiting: wait between API calls
			time.Sleep(100 * time.Millisecond)

			// Add longer pause every batch to respect rate limits
			if (i+1)%batchSize == 0 {
				s.iLog.Info(fmt.Sprintf("Processed batch of %d columns, pausing for 2 seconds...", batchSize))
				time.Sleep(2 * time.Second)
			}
		} else if err != nil {
			s.iLog.Error(fmt.Sprintf("Error checking existing embedding for column %s.%s: %v", meta.Table, meta.Column, err))
			columnsFailed++
			continue
		} else {
			s.iLog.Debug(fmt.Sprintf("Column %s.%s already has embedding (ID: %d), skipping", meta.Table, meta.Column, existingEmbedding.ID))
			columnsSkipped++
		}
	}

	s.iLog.Info(fmt.Sprintf("Column embeddings: %d processed, %d skipped, %d failed", columnsProcessed, columnsSkipped, columnsFailed))

	// STEP 5: Summary
	totalProcessed := tablesProcessed + columnsProcessed
	totalSkipped := tablesSkipped + columnsSkipped
	totalFailed := tablesFailed + columnsFailed

	s.iLog.Info(fmt.Sprintf("=== Embedding Generation Summary for '%s' ===", databaseAlias))
	s.iLog.Info(fmt.Sprintf("Tables: %d processed, %d skipped, %d failed", tablesProcessed, tablesSkipped, tablesFailed))
	s.iLog.Info(fmt.Sprintf("Columns: %d processed, %d skipped, %d failed", columnsProcessed, columnsSkipped, columnsFailed))
	s.iLog.Info(fmt.Sprintf("Total: %d processed, %d skipped, %d failed", totalProcessed, totalSkipped, totalFailed))

	if totalFailed > 0 {
		return fmt.Errorf("embedding generation completed with %d failures", totalFailed)
	}

	return nil
}

// GenerateAndStoreTableEmbedding generates and stores embedding for a table in database_schema_embeddings
func (s *SchemaEmbeddingService) GenerateAndStoreTableEmbedding(ctx context.Context, databaseAlias string, meta *models.DatabaseSchemaMetadata) error {
	// Build description text for embedding
	text := fmt.Sprintf("Table: %s", meta.Table)
	if meta.Description != "" {
		text += fmt.Sprintf("\nDescription: %s", meta.Description)
	}
	if meta.BusinessName != "" {
		text += fmt.Sprintf("\nBusiness Name: %s", meta.BusinessName)
	}

	// Generate embedding using OpenAI
	embedding, err := s.GenerateEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert float64 slice to float32 slice for pgvector
	float32Slice := make([]float32, len(embedding))
	for i, v := range embedding {
		float32Slice[i] = float32(v)
	}
	vectorData := pgvector.NewVector(float32Slice)

	// Create embedding record in database_schema_embeddings table (in vector DB)
	embeddingRecord := models.DatabaseSchemaEmbedding{
		ConfigID:        s.ConfigID,
		DatabaseAlias:   databaseAlias,
		SchemaName:      meta.SchemaName,
		MappedTableName: meta.Table,
		ColumnName:      nil, // NULL for table-level embeddings
		Description:     meta.Description,
		Embedding:       vectorData,
		GeneratedAt:     time.Now(),
		Active:          true,
		CreatedBy:       "System",
	}

	if err := s.VectorDB.Create(&embeddingRecord).Error; err != nil {
		return fmt.Errorf("failed to store embedding in vector DB: %w", err)
	}

	s.iLog.Debug(fmt.Sprintf("Stored embedding for table %s (ID: %d) in vector database", meta.Table, embeddingRecord.ID))
	return nil
}

// GenerateAndStoreColumnEmbedding generates and stores embedding for a column in database_schema_embeddings
func (s *SchemaEmbeddingService) GenerateAndStoreColumnEmbedding(ctx context.Context, databaseAlias string, meta *models.DatabaseSchemaMetadata) error {
	// Build description text for embedding
	text := fmt.Sprintf("Table: %s, Column: %s, Type: %s", meta.Table, meta.Column, meta.DataType)
	if meta.Description != "" {
		text += fmt.Sprintf("\nDescription: %s", meta.Description)
	}
	if meta.ColumnComment != "" {
		text += fmt.Sprintf("\nComment: %s", meta.ColumnComment)
	}
	if meta.BusinessName != "" {
		text += fmt.Sprintf("\nBusiness Name: %s", meta.BusinessName)
	}

	// Generate embedding using OpenAI
	embedding, err := s.GenerateEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert float64 slice to float32 slice for pgvector
	float32Slice := make([]float32, len(embedding))
	for i, v := range embedding {
		float32Slice[i] = float32(v)
	}
	vectorData := pgvector.NewVector(float32Slice)

	// Create embedding record in database_schema_embeddings table (in vector DB)
	columnName := meta.Column
	embeddingRecord := models.DatabaseSchemaEmbedding{
		ConfigID:        s.ConfigID,
		DatabaseAlias:   databaseAlias,
		SchemaName:      meta.SchemaName,
		MappedTableName: meta.Table,
		ColumnName:      &columnName, // Pointer to column name
		Description:     meta.Description,
		Embedding:       vectorData,
		GeneratedAt:     time.Now(),
		Active:          true,
		CreatedBy:       "System",
	}

	if err := s.VectorDB.Create(&embeddingRecord).Error; err != nil {
		return fmt.Errorf("failed to store embedding in vector DB: %w", err)
	}

	s.iLog.Debug(fmt.Sprintf("Stored embedding for column %s.%s (ID: %d) in vector database", meta.Table, meta.Column, embeddingRecord.ID))
	return nil
}

// SearchSimilarTables finds tables similar to the query using vector search
func (s *SchemaEmbeddingService) SearchSimilarTables(ctx context.Context, databaseAlias, query string, limit int) ([]models.DatabaseSchemaMetadata, error) {
	s.iLog.Info(fmt.Sprintf("Searching for similar tables in %s with query: %s (using VectorDB: %v)", databaseAlias, query, s.VectorDB != nil))

	// Generate embedding for query
	queryEmbedding, err := s.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert to float32 vector for PostgreSQL pgvector
	float32Slice := make([]float32, len(queryEmbedding))
	for i, v := range queryEmbedding {
		float32Slice[i] = float32(v)
	}
	queryVector := pgvector.NewVector(float32Slice)

	// Use PostgreSQL pgvector search with cosine distance
	type SearchResult struct {
		ID            int
		DatabaseAlias string
		SchemaName    string
		TableName     string
		Description   string
		Distance      float64
	}

	var searchResults []SearchResult

	// Use the provided database_alias, or 'default' if empty
	dbAlias := databaseAlias
	if dbAlias == "" {
		dbAlias = "default"
	}

	err = s.VectorDB.Raw(`
		SELECT id, database_alias, schema_name, table_name, description,
		       (embedding <=> ?::vector) AS distance
		FROM database_schema_embeddings
		WHERE database_alias = ?
		  AND column_name IS NULL
		  AND embedding IS NOT NULL
		ORDER BY distance ASC
		LIMIT ?
	`, queryVector, dbAlias, limit).Scan(&searchResults).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Vector search query failed: %v", err))
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Convert search results to DatabaseSchemaMetadata
	results := make([]models.DatabaseSchemaMetadata, len(searchResults))
	for i, sr := range searchResults {
		results[i] = models.DatabaseSchemaMetadata{
			DatabaseAlias: sr.DatabaseAlias,
			SchemaName:    sr.SchemaName,
			Table:         sr.TableName,
			MetadataType:  models.MetadataTypeTable,
			Description:   sr.Description,
		}
	}

	s.iLog.Info(fmt.Sprintf("Found %d similar tables", len(results)))
	return results, nil
}

// SearchSimilarColumns finds columns by first doing vector search on tables, then retrieving all columns for those tables
// This approach ensures we don't miss any columns that might not have embeddings
func (s *SchemaEmbeddingService) SearchSimilarColumns(ctx context.Context, databaseAlias, query string, limit int) ([]models.DatabaseSchemaMetadata, error) {
	s.iLog.Info(fmt.Sprintf("Searching for similar columns in %s with query: %s", databaseAlias, query))

	// Use the provided database_alias, or 'default' if empty
	dbAlias := databaseAlias
	if dbAlias == "" {
		dbAlias = "default"
	}

	// Step 1: Use vector search to find relevant TABLES first
	s.iLog.Debug(fmt.Sprintf("Step 1: Finding relevant tables via vector search (using VectorDB: %v)", s.VectorDB != nil))

	// Generate embedding for query
	queryEmbedding, err := s.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert to float32 vector for PostgreSQL pgvector
	float32Slice := make([]float32, len(queryEmbedding))
	for i, v := range queryEmbedding {
		float32Slice[i] = float32(v)
	}
	queryVector := pgvector.NewVector(float32Slice)

	// Search for relevant tables using vector similarity
	type TableSearchResult struct {
		DatabaseAlias string
		SchemaName    string
		TableName     string
		Distance      float64
	}

	var tableResults []TableSearchResult

	// Limit tables to a reasonable number (e.g., top 5 tables)
	tableLimit := 5
	if limit < 5 {
		tableLimit = limit
	}

	err = s.VectorDB.Raw(`
		SELECT database_alias, schema_name, table_name,
		       (embedding <=> ?::vector) AS distance
		FROM database_schema_embeddings
		WHERE database_alias = ?
		  AND column_name IS NULL
		  AND embedding IS NOT NULL
		ORDER BY distance ASC
		LIMIT ?
	`, queryVector, dbAlias, tableLimit).Scan(&tableResults).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Vector table search query failed: %v", err))
		return nil, fmt.Errorf("vector table search failed: %w", err)
	}

	if len(tableResults) == 0 {
		s.iLog.Info("No relevant tables found via vector search")
		return []models.DatabaseSchemaMetadata{}, nil
	}

	s.iLog.Debug(fmt.Sprintf("Found %d relevant tables via vector search", len(tableResults)))

	// Step 2: Retrieve ALL columns for the found tables from vector database
	s.iLog.Debug("Step 2: Retrieving all columns for found tables from vector database")

	var results []models.DatabaseSchemaMetadata

	for _, tr := range tableResults {
		// Fetch all columns for this table from vector database
		type ColumnSearchResult struct {
			DatabaseAlias string
			SchemaName    string
			TableName     string
			ColumnName    string
			Description   string
		}

		var columnResults []ColumnSearchResult
		err := s.VectorDB.Raw(`
			SELECT database_alias, schema_name, table_name, column_name, description
			FROM database_schema_embeddings
			WHERE database_alias = ?
			  AND table_name = ?
			  AND column_name IS NOT NULL
		`, tr.DatabaseAlias, tr.TableName).Scan(&columnResults).Error

		if err != nil {
			s.iLog.Warn(fmt.Sprintf("Failed to fetch columns for table %s from vector DB: %v", tr.TableName, err))
			continue
		}

		s.iLog.Debug(fmt.Sprintf("Retrieved %d columns for table %s from vector DB", len(columnResults), tr.TableName))

		// Convert to DatabaseSchemaMetadata
		for _, col := range columnResults {
			results = append(results, models.DatabaseSchemaMetadata{
				DatabaseAlias: col.DatabaseAlias,
				Table:         col.TableName,
				Column:        col.ColumnName,
				MetadataType:  models.MetadataTypeColumn,
				Description:   col.Description,
				// DataType is not stored in database_schema_embeddings
				// It will be filled in by auto-discovery if needed
			})
		}

		// Stop if we've reached the limit
		if len(results) >= limit {
			results = results[:limit]
			break
		}
	}

	s.iLog.Info(fmt.Sprintf("Found %d columns from %d relevant tables", len(results), len(tableResults)))
	return results, nil
}

// SearchSimilarBusinessEntities finds business entities similar to the query using vector search
func (s *SchemaEmbeddingService) SearchSimilarBusinessEntities(ctx context.Context, databaseAlias, query string, limit int) ([]models.BusinessEntity, error) {
	s.iLog.Info(fmt.Sprintf("Searching for similar business entities in %s with query: %s", databaseAlias, query))

	// Generate embedding for query
	queryEmbedding, err := s.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert to float32 vector for PostgreSQL pgvector
	float32Slice := make([]float32, len(queryEmbedding))
	for i, v := range queryEmbedding {
		float32Slice[i] = float32(v)
	}
	queryVector := pgvector.NewVector(float32Slice)

	// Use PostgreSQL pgvector search with cosine distance from vector database
	type EntitySearchResult struct {
		ID            int
		DatabaseAlias string
		EntityName    string
		EntityType    string
		Description   string
		// TODO: CalculationFormula field removed from BusinessEntity model
		// CalculationFormula string
		Distance float64
	}

	var searchResults []EntitySearchResult
	err = s.VectorDB.Raw(`
		SELECT id, database_alias, entity_name, entity_type, description,
		       (embedding <=> ?::vector) AS distance
		FROM business_entities
		WHERE database_alias = ?
		  AND embedding IS NOT NULL
		ORDER BY distance ASC
		LIMIT ?
	`, queryVector, databaseAlias, limit).Scan(&searchResults).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Vector search query failed: %v", err))
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Convert search results to BusinessEntity
	// Note: Business entities are ONLY stored in vector database, not in main database
	results := make([]models.BusinessEntity, 0, len(searchResults))
	for _, sr := range searchResults {
		// Use data directly from vector database search results
		results = append(results, models.BusinessEntity{
			DatabaseAlias: sr.DatabaseAlias,
			EntityName:    sr.EntityName,
			EntityType:    sr.EntityType,
			Description:   sr.Description,
			// TODO: CalculationFormula field removed from BusinessEntity model
			// CalculationFormula: sr.CalculationFormula,
		})
	}

	s.iLog.Info(fmt.Sprintf("Found %d similar business entities", len(results)))
	return results, nil
}

// EmbeddingField is a custom type for handling vector embeddings
type EmbeddingField []float64

// Scan implements the sql.Scanner interface
func (e *EmbeddingField) Scan(value interface{}) error {
	if value == nil {
		*e = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan EmbeddingField: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, e)
}

// Value implements the driver.Valuer interface
func (e EmbeddingField) Value() (driver.Value, error) {
	if e == nil {
		return nil, nil
	}
	return json.Marshal(e)
}
