package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdaxf/iac/databases/orm"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type AIEmbeddingService struct {
	db *gorm.DB
}

// NewAIEmbeddingService creates a new AI embedding service
func NewAIEmbeddingService(db *gorm.DB) *AIEmbeddingService {
	return &AIEmbeddingService{db: db}
}

// GetEmbeddingConfigurations retrieves all embedding configurations
func (s *AIEmbeddingService) GetEmbeddingConfigurations(iLog *logger.Log) ([]models.AIEmbeddingConfiguration, error) {
	var configs []models.AIEmbeddingConfiguration

	if err := s.db.Order("createdon DESC").Find(&configs).Error; err != nil {
		iLog.Error(fmt.Sprintf("Error fetching embedding configurations: %v", err))
		return nil, err
	}

	return configs, nil
}

// GetEmbeddingConfigStats retrieves embedding configuration statistics
func (s *AIEmbeddingService) GetEmbeddingConfigStats(iLog *logger.Log) ([]models.EmbeddingConfigStats, error) {
	var stats []models.EmbeddingConfigStats

	// Query the view directly using GORM
	if err := s.db.Table("v_embedding_configurations_stats").Find(&stats).Error; err != nil {
		iLog.Error(fmt.Sprintf("Error fetching embedding config stats: %v", err))
		return nil, err
	}

	return stats, nil
}

// GetDatabaseSchemaMetadata retrieves schema metadata for a database
func (s *AIEmbeddingService) GetDatabaseSchemaMetadata(configID int, databaseAlias string, iLog *logger.Log) ([]models.DatabaseSchemaEmbedding, error) {
	var embeddings []models.DatabaseSchemaEmbedding

	query := s.db.Where("config_id = ?", configID)

	if databaseAlias != "" {
		query = query.Where("database_alias = ?", databaseAlias)
	}

	if err := query.Order("schema_name, table_name, column_name").Find(&embeddings).Error; err != nil {
		iLog.Error(fmt.Sprintf("Error fetching database schema metadata: %v", err))
		return nil, err
	}

	return embeddings, nil
}

// GetDatabasesWithEmbeddings retrieves list of databases that have embeddings
func (s *AIEmbeddingService) GetDatabasesWithEmbeddings(configID int, iLog *logger.Log) ([]string, error) {
	var aliases []string

	if err := s.db.Model(&models.DatabaseSchemaEmbedding{}).
		Where("config_id = ?", configID).
		Distinct("database_alias").
		Pluck("database_alias", &aliases).Error; err != nil {
		iLog.Error(fmt.Sprintf("Error fetching databases with embeddings: %v", err))
		return nil, err
	}

	return aliases, nil
}

// GenerateSchemaEmbeddings generates vector embeddings for database schema metadata
func (s *AIEmbeddingService) GenerateSchemaEmbeddings(ctx context.Context, req models.SchemaMetadataRequest, embeddingFunc func(string) ([]float32, error), iLog *logger.Log) error {
	// Create embedding generation job
	now := time.Now()
	databaseAlias := req.DatabaseAlias
	job := models.EmbeddingGenerationJob{
		ConfigID:      req.ConfigID,
		JobType:       "schema_metadata",
		DatabaseAlias: &databaseAlias,
		Status:        "running",
		StartedAt:     &now,
		CreatedBy:     "system",
		CreatedOn:     now,
	}

	if err := s.db.WithContext(ctx).Create(&job).Error; err != nil {
		return fmt.Errorf("error creating job: %v", err)
	}

	jobID := job.ID

	// Get target database connection
	targetDB, err := orm.GetGormDB(req.DatabaseAlias)
	if err != nil {
		s.updateJobStatus(jobID, "failed", fmt.Sprintf("Error connecting to database: %v", err))
		return err
	}

	// Get schema name if not provided
	schemaName := req.SchemaName
	if schemaName == "" {
		schemaName = s.getCurrentSchema(targetDB)
	}

	// Get tables
	tables := req.Tables
	if len(tables) == 0 {
		tables, err = s.getAllTables(targetDB, schemaName, iLog)
		if err != nil {
			s.updateJobStatus(jobID, "failed", fmt.Sprintf("Error getting tables: %v", err))
			return err
		}
	}

	// Update job with total items
	if err := s.db.Model(&models.EmbeddingGenerationJob{}).
		Where("id = ?", jobID).
		Update("total_items", len(tables)).Error; err != nil {
		iLog.Error(fmt.Sprintf("Error updating job total items: %v", err))
	}

	processedItems := 0
	failedItems := 0

	// Process each table
	for _, table := range tables {
		// Get table metadata
		columns, err := s.getTableColumns(targetDB, schemaName, table, iLog)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error getting columns for table %s: %v", table, err))
			failedItems++
			continue
		}

		// Generate embedding for table
		tableDesc := fmt.Sprintf("Table: %s.%s", schemaName, table)
		embedding, err := embeddingFunc(tableDesc)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error generating embedding for table %s: %v", table, err))
			failedItems++
			continue
		}

		hash := s.computeHash(tableDesc)

		// Save table embedding (insert or update using GORM's Clauses)
		tableEmbedding := models.DatabaseSchemaEmbedding{
			ConfigID:      req.ConfigID,
			DatabaseAlias: req.DatabaseAlias,
			SchemaName:    schemaName,
			MappedTableName:  table,
			Description:   tableDesc,
			Embedding:     pgvector.NewVector(embedding),
			EmbeddingHash: hash,
			GeneratedAt:   time.Now(),
			CreatedBy:     "system",
			CreatedOn:     time.Now(),
			Active:        true,
		}

		// Using Save which does upsert-like behavior with GORM
		if err := s.db.Where("config_id = ? AND database_alias = ? AND schema_name = ? AND table_name = ? AND column_name IS NULL",
			req.ConfigID, req.DatabaseAlias, schemaName, table).
			Assign(tableEmbedding).
			FirstOrCreate(&tableEmbedding).Error; err != nil {
			iLog.Error(fmt.Sprintf("Error saving table embedding: %v", err))
		}

		// Process columns
		for _, column := range columns {
			colDesc := fmt.Sprintf("Column: %s.%s.%s (Type: %s)", schemaName, table, column.Name, column.Type)
			colEmbedding, err := embeddingFunc(colDesc)
			if err != nil {
				iLog.Error(fmt.Sprintf("Error generating embedding for column %s: %v", column.Name, err))
				continue
			}

			colHash := s.computeHash(colDesc)

			// Save column embedding
			columnEmbedding := models.DatabaseSchemaEmbedding{
				ConfigID:      req.ConfigID,
				DatabaseAlias: req.DatabaseAlias,
				SchemaName:    schemaName,
				MappedTableName:  table,
				ColumnName:    &column.Name,
				Description:   colDesc,
				Embedding:     pgvector.NewVector(colEmbedding),
				EmbeddingHash: colHash,
				GeneratedAt:   time.Now(),
				CreatedBy:     "system",
				CreatedOn:     time.Now(),
				Active:        true,
			}

			if err := s.db.Where("config_id = ? AND database_alias = ? AND schema_name = ? AND table_name = ? AND column_name = ?",
				req.ConfigID, req.DatabaseAlias, schemaName, table, column.Name).
				Assign(columnEmbedding).
				FirstOrCreate(&columnEmbedding).Error; err != nil {
				iLog.Error(fmt.Sprintf("Error saving column embedding: %v", err))
			}
		}

		processedItems++
		// Update job progress
		if err := s.db.Model(&models.EmbeddingGenerationJob{}).
			Where("id = ?", jobID).
			Updates(map[string]interface{}{
				"processed_items": processedItems,
				"failed_items":    failedItems,
			}).Error; err != nil {
			iLog.Error(fmt.Sprintf("Error updating job progress: %v", err))
		}
	}

	// Complete job
	s.updateJobStatus(jobID, "completed", "")

	return nil
}

// SearchSchema performs vector similarity search on schema embeddings
func (s *AIEmbeddingService) SearchSchema(ctx context.Context, req models.SearchRequest, embeddingFunc func(string) ([]float32, error), iLog *logger.Log) (*models.SearchResponse, error) {
	startTime := time.Now()

	// Generate query embedding
	queryEmbedding, err := embeddingFunc(req.Query)
	if err != nil {
		return nil, fmt.Errorf("error generating query embedding: %v", err)
	}

	limit := req.Limit
	if limit == 0 {
		limit = 10
	}

	var results []models.SearchResult

	switch req.SearchType {
	case "schema":
		results, err = s.searchSchemaEmbeddings(req.ConfigID, queryEmbedding, limit, iLog)
	case "entity":
		results, err = s.searchBusinessEntities(req.ConfigID, queryEmbedding, limit, iLog)
	case "query":
		results, err = s.searchQueryTemplates(req.ConfigID, queryEmbedding, limit, iLog)
	default:
		return nil, fmt.Errorf("invalid search type: %s", req.SearchType)
	}

	if err != nil {
		return nil, err
	}

	searchTime := int(time.Since(startTime).Milliseconds())

	// Log search
	topResultsJSON, _ := json.Marshal(results)

	searchLog := models.EmbeddingSearchLog{
		ConfigID:         req.ConfigID,
		SearchType:       req.SearchType,
		SearchQuery:      req.Query,
		SearchVector:     pgvector.NewVector(queryEmbedding),
		ResultsCount:     len(results),
		TopResults:       topResultsJSON,
		SearchDurationMs: searchTime,
		CreatedBy:        "system",
		CreatedOn:        time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(&searchLog).Error; err != nil {
		iLog.Error(fmt.Sprintf("Error logging search: %v", err))
	}

	return &models.SearchResponse{
		Results:      results,
		TotalResults: len(results),
		SearchTime:   searchTime,
	}, nil
}

// GetBusinessEntities retrieves all business entities
func (s *AIEmbeddingService) GetBusinessEntities(configID int, iLog *logger.Log) ([]models.AIBusinessEntity, error) {
	var entities []models.AIBusinessEntity

	if err := s.db.Where("config_id = ? AND active = ?", configID, true).
		Order("entity_name").
		Find(&entities).Error; err != nil {
		iLog.Error(fmt.Sprintf("Error fetching business entities: %v", err))
		return nil, err
	}

	// Embeddings are already loaded by GORM through the Vector type's Scan method

	return entities, nil
}

// UpdateBusinessEntity updates an existing business entity
func (s *AIEmbeddingService) UpdateBusinessEntity(ctx context.Context, id int, req models.BusinessEntityRequest, embeddingFunc func(string) ([]float32, error), iLog *logger.Log) (*models.AIBusinessEntity, error) {
	// Get existing entity
	var entity models.AIBusinessEntity
	if err := s.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("entity not found")
		}
		return nil, fmt.Errorf("error fetching entity: %v", err)
	}

	// Generate new embedding if description changed
	updates := make(map[string]interface{})

	if entity.Description != req.Description {
		embedding, err := embeddingFunc(req.Description)
		if err != nil {
			return nil, fmt.Errorf("error generating embedding: %v", err)
		}
		hash := s.computeHash(req.Description)

		updates["embedding"] = embedding
		updates["embedding_hash"] = hash
	}

	// Marshal JSON fields
	if req.FieldMappings != nil {
		fieldMappingsJSON, _ := json.Marshal(req.FieldMappings)
		updates["field_mappings"] = fieldMappingsJSON
	}
	if req.Relationships != nil {
		relationshipsJSON, _ := json.Marshal(req.Relationships)
		updates["relationships"] = relationshipsJSON
	}
	if req.BusinessRules != nil {
		businessRulesJSON, _ := json.Marshal(req.BusinessRules)
		updates["business_rules"] = businessRulesJSON
	}

	// Update other fields
	updates["entity_name"] = req.EntityName
	updates["entity_type"] = req.EntityType
	updates["description"] = req.Description
	updates["database_alias"] = req.DatabaseAlias
	updates["schema_name"] = req.SchemaName
	updates["table_name"] = req.TableName
	updates["modifiedby"] = "system"
	updates["modifiedon"] = time.Now()

	if err := s.db.WithContext(ctx).Model(&entity).Updates(updates).Error; err != nil {
		iLog.Error(fmt.Sprintf("Error updating business entity: %v", err))
		return nil, err
	}

	// Fetch updated entity
	if err := s.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		return nil, fmt.Errorf("error fetching updated entity: %v", err)
	}

	return &entity, nil
}

// DeleteBusinessEntity soft deletes a business entity
func (s *AIEmbeddingService) DeleteBusinessEntity(id int, iLog *logger.Log) error {
	result := s.db.Model(&models.AIBusinessEntity{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"active":     false,
			"modifiedby": "system",
			"modifiedon": time.Now(),
		})

	if result.Error != nil {
		iLog.Error(fmt.Sprintf("Error deleting business entity: %v", result.Error))
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("business entity not found")
	}

	return nil
}

// CreateBusinessEntity creates a new business entity with embedding
func (s *AIEmbeddingService) CreateBusinessEntity(ctx context.Context, req models.BusinessEntityRequest, embeddingFunc func(string) ([]float32, error), iLog *logger.Log) (*models.AIBusinessEntity, error) {
	// Generate embedding
	embedding, err := embeddingFunc(req.Description)
	if err != nil {
		return nil, fmt.Errorf("error generating embedding: %v", err)
	}

	// Marshal JSON fields
	fieldMappingsJSON, _ := json.Marshal(req.FieldMappings)
	relationshipsJSON, _ := json.Marshal(req.Relationships)
	businessRulesJSON, _ := json.Marshal(req.BusinessRules)

	hash := s.computeHash(req.Description)
	now := time.Now()

	entity := models.AIBusinessEntity{
		ConfigID:        req.ConfigID,
		EntityName:      req.EntityName,
		EntityType:      req.EntityType,
		Description:     req.Description,
		DatabaseAlias:   req.DatabaseAlias,
		SchemaName:      req.SchemaName,
		MappedTableName: req.TableName,
		FieldMappings:   fieldMappingsJSON,
		Relationships:   relationshipsJSON,
		BusinessRules:   businessRulesJSON,
		Embedding:       pgvector.NewVector(embedding),
		EmbeddingHash:   hash,
		GeneratedAt:     now,
		Active:          true,
		CreatedBy:       "system",
		CreatedOn:       now,
	}

	if err := s.db.WithContext(ctx).Create(&entity).Error; err != nil {
		iLog.Error(fmt.Sprintf("Error creating business entity: %v", err))
		return nil, err
	}

	return &entity, nil
}

// GetQueryTemplates retrieves all query templates
func (s *AIEmbeddingService) GetQueryTemplates(configID int, iLog *logger.Log) ([]models.AIQueryTemplate, error) {
	var templates []models.AIQueryTemplate

	if err := s.db.Where("config_id = ? AND active = ?", configID, true).
		Order("template_category, template_name").
		Find(&templates).Error; err != nil {
		iLog.Error(fmt.Sprintf("Error fetching query templates: %v", err))
		return nil, err
	}

	// Embeddings are already loaded by GORM through the Vector type's Scan method

	return templates, nil
}

// UpdateQueryTemplate updates an existing query template
func (s *AIEmbeddingService) UpdateQueryTemplate(ctx context.Context, id int, req models.QueryTemplateRequest, embeddingFunc func(string) ([]float32, error), iLog *logger.Log) (*models.AIQueryTemplate, error) {
	// Get existing template
	var template models.AIQueryTemplate
	if err := s.db.WithContext(ctx).First(&template, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("error fetching template: %v", err)
	}

	// Generate new embedding if NL query changed
	updates := make(map[string]interface{})

	if template.NaturalLanguageQuery != req.NaturalLanguageQuery {
		embedding, err := embeddingFunc(req.NaturalLanguageQuery)
		if err != nil {
			return nil, fmt.Errorf("error generating embedding: %v", err)
		}
		hash := s.computeHash(req.NaturalLanguageQuery)

		updates["embedding"] = embedding
		updates["embedding_hash"] = hash
	}

	// Marshal JSON fields
	if req.Parameters != nil {
		parametersJSON, _ := json.Marshal(req.Parameters)
		updates["parameters"] = parametersJSON
	}
	if req.EntitiesUsed != nil {
		entitiesUsedJSON, _ := json.Marshal(req.EntitiesUsed)
		updates["entities_used"] = entitiesUsedJSON
	}
	if req.ExampleQueries != nil {
		exampleQueriesJSON, _ := json.Marshal(req.ExampleQueries)
		updates["example_queries"] = exampleQueriesJSON
	}
	if req.ExpectedResultsSchema != nil {
		expectedResultsJSON, _ := json.Marshal(req.ExpectedResultsSchema)
		updates["expected_results_schema"] = expectedResultsJSON
	}

	// Update other fields
	updates["template_name"] = req.TemplateName
	updates["template_category"] = req.TemplateCategory
	updates["natural_language_query"] = req.NaturalLanguageQuery
	updates["sql_template"] = req.SQLTemplate
	updates["database_alias"] = req.DatabaseAlias
	updates["modifiedby"] = "system"
	updates["modifiedon"] = time.Now()

	if err := s.db.WithContext(ctx).Model(&template).Updates(updates).Error; err != nil {
		iLog.Error(fmt.Sprintf("Error updating query template: %v", err))
		return nil, err
	}

	// Fetch updated template
	if err := s.db.WithContext(ctx).First(&template, id).Error; err != nil {
		return nil, fmt.Errorf("error fetching updated template: %v", err)
	}

	return &template, nil
}

// DeleteQueryTemplate soft deletes a query template
func (s *AIEmbeddingService) DeleteQueryTemplate(id int, iLog *logger.Log) error {
	result := s.db.Model(&models.AIQueryTemplate{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"active":     false,
			"modifiedby": "system",
			"modifiedon": time.Now(),
		})

	if result.Error != nil {
		iLog.Error(fmt.Sprintf("Error deleting query template: %v", result.Error))
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("query template not found")
	}

	return nil
}

// CreateQueryTemplate creates a new query template with embedding
func (s *AIEmbeddingService) CreateQueryTemplate(ctx context.Context, req models.QueryTemplateRequest, embeddingFunc func(string) ([]float32, error), iLog *logger.Log) (*models.AIQueryTemplate, error) {
	// Generate embedding
	embedding, err := embeddingFunc(req.NaturalLanguageQuery)
	if err != nil {
		return nil, fmt.Errorf("error generating embedding: %v", err)
	}

	// Marshal JSON fields
	parametersJSON, _ := json.Marshal(req.Parameters)
	entitiesUsedJSON, _ := json.Marshal(req.EntitiesUsed)
	exampleQueriesJSON, _ := json.Marshal(req.ExampleQueries)
	expectedResultsJSON, _ := json.Marshal(req.ExpectedResultsSchema)

	hash := s.computeHash(req.NaturalLanguageQuery)
	now := time.Now()

	template := models.AIQueryTemplate{
		ConfigID:              req.ConfigID,
		TemplateName:          req.TemplateName,
		TemplateCategory:      req.TemplateCategory,
		NaturalLanguageQuery:  req.NaturalLanguageQuery,
		SQLTemplate:           req.SQLTemplate,
		Parameters:            parametersJSON,
		DatabaseAlias:         req.DatabaseAlias,
		EntitiesUsed:          entitiesUsedJSON,
		ExampleQueries:        exampleQueriesJSON,
		ExpectedResultsSchema: expectedResultsJSON,
		Embedding:             pgvector.NewVector(embedding),
		EmbeddingHash:         hash,
		GeneratedAt:           now,
		Active:                true,
		CreatedBy:             "system",
		CreatedOn:             now,
	}

	if err := s.db.WithContext(ctx).Create(&template).Error; err != nil {
		iLog.Error(fmt.Sprintf("Error creating query template: %v", err))
		return nil, err
	}

	return &template, nil
}

// Helper functions

func (s *AIEmbeddingService) computeHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

func (s *AIEmbeddingService) updateJobStatus(jobID int, status, errorMsg string) {
	updates := map[string]interface{}{
		"status":     status,
		"modifiedon": time.Now(),
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	if status == "completed" || status == "failed" {
		now := time.Now()
		updates["completed_at"] = now
	}

	s.db.Model(&models.EmbeddingGenerationJob{}).Where("id = ?", jobID).Updates(updates)
}

func (s *AIEmbeddingService) getCurrentSchema(db *gorm.DB) string {
	var schema string
	if err := db.Raw("SELECT current_schema()").Scan(&schema).Error; err != nil || schema == "" {
		return "public"
	}
	return schema
}

func (s *AIEmbeddingService) getAllTables(db *gorm.DB, schemaName string, iLog *logger.Log) ([]string, error) {
	var tables []string

	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = ?
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	if err := db.Raw(query, schemaName).Scan(&tables).Error; err != nil {
		return nil, err
	}

	return tables, nil
}

type TableColumnInfo struct {
	Name string
	Type string
}

func (s *AIEmbeddingService) getTableColumns(db *gorm.DB, schemaName, tableName string, iLog *logger.Log) ([]TableColumnInfo, error) {
	var columns []TableColumnInfo

	query := `
		SELECT column_name as name, data_type as type
		FROM information_schema.columns
		WHERE table_schema = ? AND table_name = ?
		ORDER BY ordinal_position
	`

	if err := db.Raw(query, schemaName, tableName).Scan(&columns).Error; err != nil {
		return nil, err
	}

	return columns, nil
}

func (s *AIEmbeddingService) searchSchemaEmbeddings(configID int, queryVector []float32, limit int, iLog *logger.Log) ([]models.SearchResult, error) {
	// This is a placeholder - actual implementation would use pgvector similarity search
	// Example: SELECT *, embedding <-> $1 AS distance FROM database_schema_embeddings WHERE config_id = $2 ORDER BY distance LIMIT $3
	var results []models.SearchResult
	// TODO: Implement actual vector similarity search
	return results, nil
}

func (s *AIEmbeddingService) searchBusinessEntities(configID int, queryVector []float32, limit int, iLog *logger.Log) ([]models.SearchResult, error) {
	var results []models.SearchResult
	// TODO: Implement actual vector similarity search
	return results, nil
}

func (s *AIEmbeddingService) searchQueryTemplates(configID int, queryVector []float32, limit int, iLog *logger.Log) ([]models.SearchResult, error) {
	var results []models.SearchResult
	// TODO: Implement actual vector similarity search
	return results, nil
}
