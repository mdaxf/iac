package services

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// SchemaEmbeddingService handles vector embedding generation and storage for schema elements
type SchemaEmbeddingService struct {
	DB        *gorm.DB
	OpenAIKey string
	ModelName string
	iLog      logger.Log
}

// NewSchemaEmbeddingService creates a new schema embedding service
func NewSchemaEmbeddingService(db *gorm.DB, openAIKey string) *SchemaEmbeddingService {
	return &SchemaEmbeddingService{
		DB:        db,
		OpenAIKey: openAIKey,
		ModelName: "text-embedding-ada-002",
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "SchemaEmbeddingService",
		},
	}
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
		s.iLog.Error(fmt.Sprintf("OpenAI embedding API returned error status: %d", resp.StatusCode))
		return nil, fmt.Errorf("OpenAI embedding API error: status %d", resp.StatusCode)
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
	err = s.DB.Model(metadata).Updates(map[string]interface{}{
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
	err = s.DB.Model(metadata).Updates(map[string]interface{}{
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
	if entity.CalculationFormula != "" {
		textBuilder.WriteString(fmt.Sprintf("Formula: %s\n", entity.CalculationFormula))
	}
	// Include synonyms if available
	if entity.Synonyms != nil {
		var synonyms []string
		if err := json.Unmarshal(entity.Synonyms, &synonyms); err == nil && len(synonyms) > 0 {
			textBuilder.WriteString(fmt.Sprintf("Synonyms: %s\n", strings.Join(synonyms, ", ")))
		}
	}
	// Include examples if available
	if entity.Examples != nil {
		var examples []string
		if err := json.Unmarshal(entity.Examples, &examples); err == nil && len(examples) > 0 {
			textBuilder.WriteString(fmt.Sprintf("Examples: %s\n", strings.Join(examples, ", ")))
		}
	}

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
	err = s.DB.Model(entity).Updates(map[string]interface{}{
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

// GenerateEmbeddingsForDatabase generates embeddings for all schema metadata and business entities in a database
func (s *SchemaEmbeddingService) GenerateEmbeddingsForDatabase(ctx context.Context, databaseAlias string) error {
	s.iLog.Info(fmt.Sprintf("Starting batch embedding generation for database: %s", databaseAlias))

	// Generate embeddings for tables
	var tableMetadata []models.DatabaseSchemaMetadata
	err := s.DB.Where("databasealias = ? AND metadatatype = ? AND (embedding IS NULL OR embedding_generated_at IS NULL)",
		databaseAlias, models.MetadataTypeTable).Find(&tableMetadata).Error
	if err != nil {
		return fmt.Errorf("failed to fetch table metadata: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Found %d tables without embeddings", len(tableMetadata)))
	for i, meta := range tableMetadata {
		s.iLog.Debug(fmt.Sprintf("Processing table %d/%d: %s", i+1, len(tableMetadata), meta.Table))
		if err := s.GenerateTableEmbedding(ctx, &meta); err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to generate embedding for table %s: %v", meta.Table, err))
			// Continue with next table instead of failing entirely
			continue
		}
		// Rate limiting: wait 100ms between API calls to avoid rate limits
		time.Sleep(100 * time.Millisecond)
	}

	// Generate embeddings for columns
	var columnMetadata []models.DatabaseSchemaMetadata
	err = s.DB.Where("databasealias = ? AND metadatatype = ? AND (embedding IS NULL OR embedding_generated_at IS NULL)",
		databaseAlias, models.MetadataTypeColumn).Find(&columnMetadata).Error
	if err != nil {
		return fmt.Errorf("failed to fetch column metadata: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Found %d columns without embeddings", len(columnMetadata)))
	for i, meta := range columnMetadata {
		s.iLog.Debug(fmt.Sprintf("Processing column %d/%d: %s.%s", i+1, len(columnMetadata), meta.Table, meta.Column))
		if err := s.GenerateColumnEmbedding(ctx, &meta); err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to generate embedding for column %s.%s: %v", meta.Table, meta.Column, err))
			// Continue with next column instead of failing entirely
			continue
		}
		// Rate limiting: wait 100ms between API calls
		time.Sleep(100 * time.Millisecond)
	}

	// Generate embeddings for business entities
	var entities []models.BusinessEntity
	err = s.DB.Where("databasealias = ? AND (embedding IS NULL OR embedding_generated_at IS NULL)",
		databaseAlias).Find(&entities).Error
	if err != nil {
		return fmt.Errorf("failed to fetch business entities: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Found %d business entities without embeddings", len(entities)))
	for i, entity := range entities {
		s.iLog.Debug(fmt.Sprintf("Processing entity %d/%d: %s", i+1, len(entities), entity.EntityName))
		if err := s.GenerateBusinessEntityEmbedding(ctx, &entity); err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to generate embedding for entity %s: %v", entity.EntityName, err))
			// Continue with next entity instead of failing entirely
			continue
		}
		// Rate limiting: wait 100ms between API calls
		time.Sleep(100 * time.Millisecond)
	}

	s.iLog.Info(fmt.Sprintf("Batch embedding generation completed for database: %s", databaseAlias))
	return nil
}

// SearchSimilarTables finds tables similar to the query using vector search
func (s *SchemaEmbeddingService) SearchSimilarTables(ctx context.Context, databaseAlias, query string, limit int) ([]models.DatabaseSchemaMetadata, error) {
	s.iLog.Info(fmt.Sprintf("Searching for similar tables in %s with query: %s", databaseAlias, query))

	// Generate embedding for query
	queryEmbedding, err := s.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert to JSON for MySQL
	queryJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}

	// Use MySQL vector search with cosine distance
	var results []models.DatabaseSchemaMetadata
	err = s.DB.Raw(`
		SELECT id, databasealias, schemaname, tablename, metadatatype, description,
		       embedding <-> CAST(? AS JSON) AS distance
		FROM databaseschemametadata
		WHERE databasealias = ?
		  AND metadatatype = ?
		  AND embedding IS NOT NULL
		ORDER BY distance ASC
		LIMIT ?
	`, string(queryJSON), databaseAlias, models.MetadataTypeTable, limit).Scan(&results).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Vector search query failed: %v", err))
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Found %d similar tables", len(results)))
	return results, nil
}

// SearchSimilarColumns finds columns similar to the query using vector search
func (s *SchemaEmbeddingService) SearchSimilarColumns(ctx context.Context, databaseAlias, query string, limit int) ([]models.DatabaseSchemaMetadata, error) {
	s.iLog.Info(fmt.Sprintf("Searching for similar columns in %s with query: %s", databaseAlias, query))

	// Generate embedding for query
	queryEmbedding, err := s.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert to JSON for MySQL
	queryJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}

	// Use MySQL vector search with cosine distance
	var results []models.DatabaseSchemaMetadata
	err = s.DB.Raw(`
		SELECT id, databasealias, schemaname, tablename, columnname, datatype,
		       columncomment, description, metadatatype,
		       embedding <-> CAST(? AS JSON) AS distance
		FROM databaseschemametadata
		WHERE databasealias = ?
		  AND metadatatype = ?
		  AND embedding IS NOT NULL
		ORDER BY distance ASC
		LIMIT ?
	`, string(queryJSON), databaseAlias, models.MetadataTypeColumn, limit).Scan(&results).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Vector search query failed: %v", err))
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Found %d similar columns", len(results)))
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

	// Convert to JSON for MySQL
	queryJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}

	// Use MySQL vector search with cosine distance
	var results []models.BusinessEntity
	err = s.DB.Raw(`
		SELECT id, databasealias, entityname, entitytype, description, calculationformula,
		       embedding <-> CAST(? AS JSON) AS distance
		FROM businessentities
		WHERE databasealias = ?
		  AND embedding IS NOT NULL
		ORDER BY distance ASC
		LIMIT ?
	`, string(queryJSON), databaseAlias, limit).Scan(&results).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Vector search query failed: %v", err))
		return nil, fmt.Errorf("vector search failed: %w", err)
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
