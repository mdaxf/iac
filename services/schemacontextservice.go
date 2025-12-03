package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// SchemaContextService provides schema context for AI chat and report generation
// Uses vector search to find relevant tables, columns, business entities, and query templates
type SchemaContextService struct {
	DB                     *gorm.DB
	VectorDB               *gorm.DB
	SchemaEmbeddingService *SchemaEmbeddingService
	iLog                   logger.Log
}

// NewSchemaContextService creates a new schema context service
func NewSchemaContextService(db *gorm.DB, openAIKey string) *SchemaContextService {
	// Get vector database connection
	vectorDB, err := GetVectorDB(db)
	if err != nil {
		vectorDB = db // Fallback to main DB
	}

	return &SchemaContextService{
		DB:                     db,
		VectorDB:               vectorDB,
		SchemaEmbeddingService: NewSchemaEmbeddingService(db, openAIKey),
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "SchemaContextService",
		},
	}
}

// SchemaContext represents the schema context for AI
type SchemaContext struct {
	Tables           []AITableInfo          `json:"tables"`
	BusinessEntities []BusinessEntityInfo  `json:"business_entities"`
	QueryTemplates   []QueryTemplateInfo   `json:"query_templates"`
	TotalTables      int                   `json:"total_tables"`
	TotalEntities    int                   `json:"total_entities"`
	TotalTemplates   int                   `json:"total_templates"`
}

// AITableInfo contains table and its columns information for AI context
type AITableInfo struct {
	TableName   string         `json:"table_name"`
	SchemaName  string         `json:"schema_name"`
	Description string         `json:"description"`
	Similarity  float64        `json:"similarity,omitempty"`
	Columns     []AIColumnInfo `json:"columns"`
}

// AIColumnInfo contains column information for AI context
type AIColumnInfo struct {
	ColumnName  string  `json:"column_name"`
	DataType    string  `json:"data_type"`
	Description string  `json:"description"`
	Similarity  float64 `json:"similarity,omitempty"`
}

// BusinessEntityInfo contains business entity information
type BusinessEntityInfo struct {
	EntityName  string  `json:"entity_name"`
	EntityType  string  `json:"entity_type"`
	Description string  `json:"description"`
	Formula     string  `json:"formula,omitempty"`
	Similarity  float64 `json:"similarity,omitempty"`
}

// QueryTemplateInfo contains query template information
type QueryTemplateInfo struct {
	TemplateName string  `json:"template_name"`
	Category     string  `json:"category"`
	Description  string  `json:"description"`
	SQLTemplate  string  `json:"sql_template"`
	Similarity   float64 `json:"similarity,omitempty"`
}

// GetSchemaContextByQuery retrieves relevant schema context using natural language query
// Uses vector search to find the most relevant tables, columns, entities, and templates
func (s *SchemaContextService) GetSchemaContextByQuery(ctx context.Context, databaseAlias, naturalLanguageQuery string, maxTables int) (*SchemaContext, error) {
	s.iLog.Info(fmt.Sprintf("Getting schema context for database '%s' with query: %s", databaseAlias, naturalLanguageQuery))

	if maxTables <= 0 {
		maxTables = 10 // Default to 10 tables
	}
	if maxTables > 20 {
		maxTables = 20 // Cap at 20 tables
	}

	context := &SchemaContext{
		Tables:           []AITableInfo{},
		BusinessEntities: []BusinessEntityInfo{},
		QueryTemplates:   []QueryTemplateInfo{},
	}

	// 1. Search for relevant tables using vector search
	s.iLog.Debug("Searching for relevant tables...")
	tables, err := s.searchRelevantTables(ctx, databaseAlias, naturalLanguageQuery, maxTables)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to search tables: %v", err))
		// Continue anyway, don't fail the entire operation
	} else {
		context.Tables = tables
		context.TotalTables = len(tables)
	}

	// 2. Search for relevant business entities
	s.iLog.Debug("Searching for relevant business entities...")
	entities, err := s.searchRelevantBusinessEntities(ctx, databaseAlias, naturalLanguageQuery, 10)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to search business entities: %v", err))
	} else {
		context.BusinessEntities = entities
		context.TotalEntities = len(entities)
	}

	// 3. Search for relevant query templates
	s.iLog.Debug("Searching for relevant query templates...")
	templates, err := s.searchRelevantQueryTemplates(ctx, databaseAlias, naturalLanguageQuery, 5)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to search query templates: %v", err))
	} else {
		context.QueryTemplates = templates
		context.TotalTemplates = len(templates)
	}

	s.iLog.Info(fmt.Sprintf("Schema context retrieved: %d tables, %d entities, %d templates",
		context.TotalTables, context.TotalEntities, context.TotalTemplates))

	return context, nil
}

// searchRelevantTables searches for relevant tables and their columns using vector search
func (s *SchemaContextService) searchRelevantTables(ctx context.Context, databaseAlias, query string, limit int) ([]AITableInfo, error) {
	// Generate embedding for the query
	queryEmbedding, err := s.SchemaEmbeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert to float32 for vector comparison
	float32Slice := make([]float32, len(queryEmbedding))
	for i, v := range queryEmbedding {
		float32Slice[i] = float32(v)
	}
	queryVector := pgvector.NewVector(float32Slice)

	// Search for similar tables using vector similarity
	type TableResult struct {
		ID              int
		DatabaseAlias   string
		SchemaName      string
		TableName       string
		Description     string
		Distance        float64
	}

	var tableResults []TableResult

	// Use cosine distance for similarity search
	err = s.VectorDB.Raw(`
		SELECT
			id,
			database_alias,
			schema_name,
			table_name,
			description,
			(embedding <=> ?::vector) AS distance
		FROM database_schema_embeddings
		WHERE database_alias = ?
		  AND column_name IS NULL
		  AND embedding IS NOT NULL
		ORDER BY distance ASC
		LIMIT ?
	`, queryVector, databaseAlias, limit).Scan(&tableResults).Error

	if err != nil {
		return nil, fmt.Errorf("vector search query failed: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Found %d relevant tables", len(tableResults)))

	// For each table, get its columns
	tables := make([]AITableInfo, 0, len(tableResults))
	for _, tableResult := range tableResults {
		tableInfo := AITableInfo{
			TableName:   tableResult.TableName,
			SchemaName:  tableResult.SchemaName,
			Description: tableResult.Description,
			Similarity:  1.0 - tableResult.Distance, // Convert distance to similarity
			Columns:     []AIColumnInfo{},
		}

		// Get columns for this table (limit to top 20 columns)
		columns, err := s.searchRelevantColumnsForTable(ctx, databaseAlias, tableResult.TableName, query, 20)
		if err != nil {
			s.iLog.Warn(fmt.Sprintf("Failed to get columns for table %s: %v", tableResult.TableName, err))
		} else {
			tableInfo.Columns = columns
		}

		tables = append(tables, tableInfo)
	}

	return tables, nil
}

// searchRelevantColumnsForTable searches for relevant columns of a specific table
func (s *SchemaContextService) searchRelevantColumnsForTable(ctx context.Context, databaseAlias, tableName, query string, limit int) ([]AIColumnInfo, error) {
	// Generate embedding for the query
	queryEmbedding, err := s.SchemaEmbeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		// If embedding generation fails, fall back to getting all columns
		return s.getAllColumnsForTable(databaseAlias, tableName, limit)
	}

	// Convert to float32 for vector comparison
	float32Slice := make([]float32, len(queryEmbedding))
	for i, v := range queryEmbedding {
		float32Slice[i] = float32(v)
	}
	queryVector := pgvector.NewVector(float32Slice)

	type ColumnResult struct {
		ColumnName  string
		Description string
		Distance    float64
	}

	var columnResults []ColumnResult

	err = s.VectorDB.Raw(`
		SELECT
			column_name,
			description,
			(embedding <=> ?::vector) AS distance
		FROM database_schema_embeddings
		WHERE database_alias = ?
		  AND table_name = ?
		  AND column_name IS NOT NULL
		  AND embedding IS NOT NULL
		ORDER BY distance ASC
		LIMIT ?
	`, queryVector, databaseAlias, tableName, limit).Scan(&columnResults).Error

	if err != nil {
		s.iLog.Warn(fmt.Sprintf("Vector search for columns failed: %v, falling back to metadata query", err))
		return s.getAllColumnsForTable(databaseAlias, tableName, limit)
	}

	columns := make([]AIColumnInfo, 0, len(columnResults))
	for _, col := range columnResults {
		// Get additional column metadata from databaseschemametadata
		var metadata models.DatabaseSchemaMetadata
		s.DB.Where("databasealias = ? AND tablename = ? AND columnname = ? AND metadatatype = ?",
			databaseAlias, tableName, col.ColumnName, models.MetadataTypeColumn).First(&metadata)

		columns = append(columns, AIColumnInfo{
			ColumnName:  col.ColumnName,
			DataType:    metadata.DataType,
			Description: col.Description,
			Similarity:  1.0 - col.Distance,
		})
	}

	return columns, nil
}

// getAllColumnsForTable gets all columns for a table (fallback when vector search fails)
func (s *SchemaContextService) getAllColumnsForTable(databaseAlias, tableName string, limit int) ([]AIColumnInfo, error) {
	var metadata []models.DatabaseSchemaMetadata
	err := s.DB.Where("databasealias = ? AND tablename = ? AND metadatatype = ?",
		databaseAlias, tableName, models.MetadataTypeColumn).
		Limit(limit).
		Find(&metadata).Error

	if err != nil {
		return nil, err
	}

	columns := make([]AIColumnInfo, 0, len(metadata))
	for _, meta := range metadata {
		columns = append(columns, AIColumnInfo{
			ColumnName:  meta.Column,
			DataType:    meta.DataType,
			Description: meta.Description,
		})
	}

	return columns, nil
}

// searchRelevantBusinessEntities searches for relevant business entities using vector search
func (s *SchemaContextService) searchRelevantBusinessEntities(ctx context.Context, databaseAlias, query string, limit int) ([]BusinessEntityInfo, error) {
	// Use the SchemaEmbeddingService's search method
	entities, err := s.SchemaEmbeddingService.SearchSimilarBusinessEntities(ctx, databaseAlias, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search business entities: %w", err)
	}

	entityInfos := make([]BusinessEntityInfo, 0, len(entities))
	for _, entity := range entities {
		entityInfos = append(entityInfos, BusinessEntityInfo{
			EntityName:  entity.EntityName,
			EntityType:  string(entity.EntityType),
			Description: entity.Description,
			// TODO: CalculationFormula field removed from BusinessEntity model
			// Formula:     entity.CalculationFormula,
		})
	}

	return entityInfos, nil
}

// searchRelevantQueryTemplates searches for relevant query templates using vector search
func (s *SchemaContextService) searchRelevantQueryTemplates(ctx context.Context, databaseAlias, query string, limit int) ([]QueryTemplateInfo, error) {
	// Generate embedding for the query
	queryEmbedding, err := s.SchemaEmbeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert to float32 for vector comparison
	float32Slice := make([]float32, len(queryEmbedding))
	for i, v := range queryEmbedding {
		float32Slice[i] = float32(v)
	}
	queryVector := pgvector.NewVector(float32Slice)

	type TemplateResult struct {
		TemplateName         string
		TemplateCategory     string
		NaturalLanguageQuery string
		SQLTemplate          string
		Distance             float64
	}

	var templateResults []TemplateResult

	err = s.VectorDB.Raw(`
		SELECT
			template_name,
			template_category,
			natural_language_query,
			sql_template,
			(embedding <=> ?::vector) AS distance
		FROM query_templates
		WHERE database_alias = ?
		  AND embedding IS NOT NULL
		  AND active = true
		ORDER BY distance ASC
		LIMIT ?
	`, queryVector, databaseAlias, limit).Scan(&templateResults).Error

	if err != nil {
		s.iLog.Warn(fmt.Sprintf("Vector search for query templates failed: %v", err))
		return []QueryTemplateInfo{}, nil // Return empty slice, not an error
	}

	templates := make([]QueryTemplateInfo, 0, len(templateResults))
	for _, tmpl := range templateResults {
		templates = append(templates, QueryTemplateInfo{
			TemplateName: tmpl.TemplateName,
			Category:     tmpl.TemplateCategory,
			Description:  tmpl.NaturalLanguageQuery,
			SQLTemplate:  tmpl.SQLTemplate,
			Similarity:   1.0 - tmpl.Distance,
		})
	}

	return templates, nil
}

// FormatSchemaContextForAI formats the schema context as a string for AI prompt
func (s *SchemaContextService) FormatSchemaContextForAI(context *SchemaContext) string {
	var builder strings.Builder

	builder.WriteString("=== DATABASE SCHEMA CONTEXT ===\n\n")

	// Format tables and columns
	if len(context.Tables) > 0 {
		builder.WriteString(fmt.Sprintf("TABLES (%d relevant tables found):\n\n", context.TotalTables))
		for i, table := range context.Tables {
			builder.WriteString(fmt.Sprintf("%d. Table: %s", i+1, table.TableName))
			if table.SchemaName != "" {
				builder.WriteString(fmt.Sprintf(" (Schema: %s)", table.SchemaName))
			}
			if table.Description != "" {
				builder.WriteString(fmt.Sprintf("\n   Description: %s", table.Description))
			}
			builder.WriteString(fmt.Sprintf("\n   Relevance: %.1f%%", table.Similarity*100))
			builder.WriteString("\n   Columns:\n")

			for _, col := range table.Columns {
				builder.WriteString(fmt.Sprintf("   - %s", col.ColumnName))
				if col.DataType != "" {
					builder.WriteString(fmt.Sprintf(" (%s)", col.DataType))
				}
				if col.Description != "" {
					builder.WriteString(fmt.Sprintf(" - %s", col.Description))
				}
				builder.WriteString("\n")
			}
			builder.WriteString("\n")
		}
	}

	// Format business entities
	if len(context.BusinessEntities) > 0 {
		builder.WriteString(fmt.Sprintf("\nBUSINESS ENTITIES (%d relevant entities found):\n\n", context.TotalEntities))
		for i, entity := range context.BusinessEntities {
			builder.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, entity.EntityName, entity.EntityType))
			if entity.Description != "" {
				builder.WriteString(fmt.Sprintf("   Description: %s\n", entity.Description))
			}
			if entity.Formula != "" {
				builder.WriteString(fmt.Sprintf("   Formula: %s\n", entity.Formula))
			}
			builder.WriteString("\n")
		}
	}

	// Format query templates
	if len(context.QueryTemplates) > 0 {
		builder.WriteString(fmt.Sprintf("\nQUERY TEMPLATES (%d relevant examples found):\n\n", context.TotalTemplates))
		for i, tmpl := range context.QueryTemplates {
			builder.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, tmpl.TemplateName, tmpl.Category))
			if tmpl.Description != "" {
				builder.WriteString(fmt.Sprintf("   Intent: %s\n", tmpl.Description))
			}
			builder.WriteString(fmt.Sprintf("   SQL: %s\n", tmpl.SQLTemplate))
			builder.WriteString(fmt.Sprintf("   Relevance: %.1f%%\n", tmpl.Similarity*100))
			builder.WriteString("\n")
		}
	}

	builder.WriteString("=== END OF SCHEMA CONTEXT ===\n")

	return builder.String()
}
