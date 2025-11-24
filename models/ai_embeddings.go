package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// AIEmbeddingConfiguration stores AI configuration metadata (IAC Standard)
type AIEmbeddingConfiguration struct {
	ID                   int             `json:"id" gorm:"primaryKey"`
	UUID                 string          `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	ReferenceID          *string         `json:"referenceid" gorm:"uniqueIndex"`
	ConfigName           string          `json:"config_name" gorm:"uniqueIndex;not null"`
	EmbeddingModel       string          `json:"embedding_model" gorm:"not null"`
	EmbeddingDimensions  int             `json:"embedding_dimensions" gorm:"not null"`
	VectorDatabaseType   string          `json:"vector_database_type" gorm:"default:postgresql"`
	VectorDatabaseConfig json.RawMessage `json:"vector_database_config" gorm:"type:jsonb"`
	Active               bool            `json:"active" gorm:"default:true"`
	CreatedBy            string          `json:"createdby" gorm:"not null"`
	CreatedOn            time.Time       `json:"createdon" gorm:"default:CURRENT_TIMESTAMP"`
	ModifiedBy           *string         `json:"modifiedby"`
	ModifiedOn           *time.Time      `json:"modifiedon"`
	RowVersionStamp      int             `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name for GORM
func (AIEmbeddingConfiguration) TableName() string {
	return "ai_embedding_configurations"
}

// AIBusinessEntity stores business entity definitions with embeddings (IAC Standard)
type AIBusinessEntity struct {
	ID              int             `json:"id" gorm:"primaryKey"`
	UUID            string          `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	ReferenceID     *string         `json:"referenceid" gorm:"uniqueIndex"`
	ConfigID        int             `json:"config_id" gorm:"not null"`
	EntityName      string          `json:"entity_name" gorm:"not null"`
	EntityType      string          `json:"entity_type"`
	Description     string          `json:"description" gorm:"not null"`
	DatabaseAlias   *string         `json:"database_alias"`
	SchemaName      *string         `json:"schema_name"`
	MappedTableName *string         `json:"table_name" gorm:"column:table_name"`
	FieldMappings   json.RawMessage `json:"field_mappings" gorm:"type:jsonb"`
	Relationships   json.RawMessage `json:"relationships" gorm:"type:jsonb"`
	BusinessRules   json.RawMessage `json:"business_rules" gorm:"type:jsonb"`
	Metadata        json.RawMessage `json:"metadata" gorm:"type:jsonb"`
	Embedding       Vector          `json:"embedding,omitempty" gorm:"type:vector"`
	EmbeddingHash   string          `json:"embedding_hash"`
	GeneratedAt     time.Time       `json:"generated_at"`
	Active          bool            `json:"active" gorm:"default:true"`
	CreatedBy       string          `json:"createdby" gorm:"not null"`
	CreatedOn       time.Time       `json:"createdon" gorm:"default:CURRENT_TIMESTAMP"`
	ModifiedBy      *string         `json:"modifiedby"`
	ModifiedOn      *time.Time      `json:"modifiedon"`
	RowVersionStamp int             `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name for GORM
func (AIBusinessEntity) TableName() string {
	return "business_entities"
}

// AIQueryTemplate stores SQL query templates with embeddings (IAC Standard)
type AIQueryTemplate struct {
	ID                    int             `json:"id" gorm:"primaryKey"`
	UUID                  string          `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	ReferenceID           *string         `json:"referenceid" gorm:"uniqueIndex"`
	ConfigID              int             `json:"config_id" gorm:"not null"`
	TemplateName          string          `json:"template_name" gorm:"not null"`
	TemplateCategory      string          `json:"template_category"`
	NaturalLanguageQuery  string          `json:"natural_language_query" gorm:"not null"`
	SQLTemplate           string          `json:"sql_template" gorm:"not null"`
	Parameters            json.RawMessage `json:"parameters" gorm:"type:jsonb"`
	DatabaseAlias         *string         `json:"database_alias"`
	EntitiesUsed          json.RawMessage `json:"entities_used" gorm:"type:jsonb"`
	ExampleQueries        json.RawMessage `json:"example_queries" gorm:"type:jsonb"`
	ExpectedResultsSchema json.RawMessage `json:"expected_results_schema" gorm:"type:jsonb"`
	UsageCount            int             `json:"usage_count" gorm:"default:0"`
	LastUsedAt            *time.Time      `json:"last_used_at"`
	Embedding             Vector          `json:"embedding,omitempty" gorm:"type:vector"`
	EmbeddingHash         string          `json:"embedding_hash"`
	GeneratedAt           time.Time       `json:"generated_at"`
	Active                bool            `json:"active" gorm:"default:true"`
	CreatedBy             string          `json:"createdby" gorm:"not null"`
	CreatedOn             time.Time       `json:"createdon" gorm:"default:CURRENT_TIMESTAMP"`
	ModifiedBy            *string         `json:"modifiedby"`
	ModifiedOn            *time.Time      `json:"modifiedon"`
	RowVersionStamp       int             `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name for GORM
func (AIQueryTemplate) TableName() string {
	return "query_templates"
}

// DatabaseSchemaEmbedding stores vector embeddings for database schema metadata (IAC Standard)
type DatabaseSchemaEmbedding struct {
	ID              int             `json:"id" gorm:"primaryKey"`
	UUID            string          `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	ReferenceID     *string         `json:"referenceid" gorm:"uniqueIndex"`
	ConfigID        int             `json:"config_id" gorm:"not null"`
	DatabaseAlias   string          `json:"database_alias" gorm:"not null"`
	SchemaName      string          `json:"schema_name" gorm:"not null"`
	MappedTableName string          `json:"table_name" gorm:"column:table_name;not null"`
	ColumnName      *string         `json:"column_name"`
	Description     string          `json:"description"`
	Metadata        json.RawMessage `json:"metadata" gorm:"type:jsonb"`
	Embedding       Vector          `json:"embedding,omitempty" gorm:"type:vector"`
	EmbeddingHash   string          `json:"embedding_hash"`
	GeneratedAt     time.Time       `json:"generated_at"`
	Active          bool            `json:"active" gorm:"default:true"`
	CreatedBy       string          `json:"createdby" gorm:"not null"`
	CreatedOn       time.Time       `json:"createdon" gorm:"default:CURRENT_TIMESTAMP"`
	ModifiedBy      *string         `json:"modifiedby"`
	ModifiedOn      *time.Time      `json:"modifiedon"`
	RowVersionStamp int             `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name for GORM
func (DatabaseSchemaEmbedding) TableName() string {
	return "database_schema_embeddings"
}

// EmbeddingGenerationJob tracks batch embedding generation (IAC Standard)
type EmbeddingGenerationJob struct {
	ID              int        `json:"id" gorm:"primaryKey"`
	UUID            string     `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	ReferenceID     *string    `json:"referenceid" gorm:"uniqueIndex"`
	ConfigID        int        `json:"config_id" gorm:"not null"`
	JobType         string     `json:"job_type" gorm:"not null"`
	DatabaseAlias   *string    `json:"database_alias"`
	Status          string     `json:"status" gorm:"default:pending"`
	TotalItems      int        `json:"total_items"`
	ProcessedItems  int        `json:"processed_items" gorm:"default:0"`
	FailedItems     int        `json:"failed_items" gorm:"default:0"`
	ErrorMessage    *string    `json:"error_message"`
	StartedAt       *time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	Active          bool       `json:"active" gorm:"default:true"`
	CreatedBy       string     `json:"createdby" gorm:"not null"`
	CreatedOn       time.Time  `json:"createdon" gorm:"default:CURRENT_TIMESTAMP"`
	ModifiedBy      *string    `json:"modifiedby"`
	ModifiedOn      *time.Time `json:"modifiedon"`
	RowVersionStamp int        `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name for GORM
func (EmbeddingGenerationJob) TableName() string {
	return "embedding_generation_jobs"
}

// EmbeddingSearchLog logs vector similarity searches (IAC Standard)
type EmbeddingSearchLog struct {
	ID               int             `json:"id" gorm:"primaryKey"`
	UUID             string          `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	ReferenceID      *string         `json:"referenceid" gorm:"uniqueIndex"`
	ConfigID         int             `json:"config_id" gorm:"not null"`
	SearchType       string          `json:"search_type" gorm:"not null"`
	SearchQuery      string          `json:"search_query" gorm:"not null"`
	SearchVector     Vector          `json:"search_vector,omitempty" gorm:"type:vector"`
	ResultsCount     int             `json:"results_count"`
	TopResults       json.RawMessage `json:"top_results" gorm:"type:jsonb"`
	SearchDurationMs int             `json:"search_duration_ms"`
	UserFeedback     *string         `json:"user_feedback"`
	Active           bool            `json:"active" gorm:"default:true"`
	CreatedBy        string          `json:"createdby" gorm:"not null"`
	CreatedOn        time.Time       `json:"createdon" gorm:"default:CURRENT_TIMESTAMP"`
	ModifiedBy       *string         `json:"modifiedby"`
	ModifiedOn       *time.Time      `json:"modifiedon"`
	RowVersionStamp  int             `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name for GORM
func (EmbeddingSearchLog) TableName() string {
	return "embedding_search_logs"
}

// Vector represents a pgvector embedding
type Vector []float32

// Value implements the driver.Valuer interface
func (v Vector) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

// Scan implements the sql.Scanner interface
func (v *Vector) Scan(value interface{}) error {
	if value == nil {
		*v = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, v)
}

// EmbeddingConfigStats holds statistics for embedding configurations (IAC Standard)
type EmbeddingConfigStats struct {
	ID                           int        `json:"id"`
	UUID                         string     `json:"uuid"`
	ConfigName                   string     `json:"config_name"`
	EmbeddingModel               string     `json:"embedding_model"`
	EmbeddingDimensions          int        `json:"embedding_dimensions"`
	VectorDatabaseType           string     `json:"vector_database_type"`
	Active                       bool       `json:"active"`
	CreatedOn                    time.Time  `json:"createdon"`
	DatabasesWithEmbeddings      int        `json:"databases_with_embeddings"`
	TablesWithEmbeddings         int        `json:"tables_with_embeddings"`
	TotalSchemaEmbeddings        int        `json:"total_schema_embeddings"`
	BusinessEntitiesCount        int        `json:"business_entities_count"`
	QueryTemplatesCount          int        `json:"query_templates_count"`
	LastSchemaEmbeddingGenerated *time.Time `json:"last_schema_embedding_generated"`
	LastEntityGenerated          *time.Time `json:"last_entity_generated"`
	LastTemplateGenerated        *time.Time `json:"last_template_generated"`
}

// DatabaseSchemaMetadataSummary holds summary of database schema metadata
type DatabaseSchemaMetadataSummary struct {
	DatabaseAlias           string     `json:"database_alias"`
	SchemaName              string     `json:"schema_name"`
	TableName               string     `json:"table_name"`
	HasTableDescription     int        `json:"has_table_description"`
	ColumnsWithDescriptions int        `json:"columns_with_descriptions"`
	LastUpdated             *time.Time `json:"last_updated"`
}

// SchemaMetadataRequest represents request to generate schema embeddings
type SchemaMetadataRequest struct {
	ConfigID      int    `json:"config_id" binding:"required"`
	DatabaseAlias string `json:"database_alias" binding:"required"`
	SchemaName    string `json:"schema_name"`
	Tables        []string `json:"tables"` // Empty means all tables
}

// BusinessEntityRequest represents request to create/update business entity
type BusinessEntityRequest struct {
	ConfigID      int                    `json:"config_id" binding:"required"`
	EntityName    string                 `json:"entity_name" binding:"required"`
	EntityType    string                 `json:"entity_type"`
	Description   string                 `json:"description" binding:"required"`
	DatabaseAlias *string                `json:"database_alias"`
	SchemaName    *string                `json:"schema_name"`
	TableName     *string                `json:"table_name"`
	FieldMappings map[string]interface{} `json:"field_mappings"`
	Relationships map[string]interface{} `json:"relationships"`
	BusinessRules map[string]interface{} `json:"business_rules"`
}

// QueryTemplateRequest represents request to create/update query template
type QueryTemplateRequest struct {
	ConfigID             int                    `json:"config_id" binding:"required"`
	TemplateName         string                 `json:"template_name" binding:"required"`
	TemplateCategory     string                 `json:"template_category"`
	NaturalLanguageQuery string                 `json:"natural_language_query" binding:"required"`
	SQLTemplate          string                 `json:"sql_template" binding:"required"`
	Parameters           map[string]interface{} `json:"parameters"`
	DatabaseAlias        *string                `json:"database_alias"`
	EntitiesUsed         []string               `json:"entities_used"`
	ExampleQueries       []string               `json:"example_queries"`
	ExpectedResultsSchema map[string]interface{} `json:"expected_results_schema"`
}

// SearchRequest represents a vector similarity search request
type SearchRequest struct {
	ConfigID   int    `json:"config_id" binding:"required"`
	SearchType string `json:"search_type" binding:"required"` // 'schema', 'entity', 'query'
	Query      string `json:"query" binding:"required"`
	Limit      int    `json:"limit"` // Default 10
}

// SearchResult represents a single search result
type SearchResult struct {
	ID          int                    `json:"id"`
	Score       float32                `json:"score"` // Similarity score
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Results      []SearchResult `json:"results"`
	TotalResults int            `json:"total_results"`
	SearchTime   int            `json:"search_time_ms"`
}
