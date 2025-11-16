package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Conversation represents a chat conversation thread
type Conversation struct {
	ID               string    `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	Title            string    `json:"title" gorm:"type:varchar(255);not null"`
	UserID           string    `json:"user_id" gorm:"type:varchar(36);not null"`
	DatabaseAlias    string    `json:"database_alias" gorm:"type:varchar(100)"`
	AutoExecuteQuery bool      `json:"auto_execute_query" gorm:"default:true"`
	IsActive         bool      `json:"is_active" gorm:"default:true"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Messages []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name
func (Conversation) TableName() string {
	return "conversations"
}

// MessageType represents the type of chat message
type MessageType string

const (
	MessageTypeUser      MessageType = "user"
	MessageTypeAssistant MessageType = "assistant"
)

// ChatMessage represents a message in a conversation
type ChatMessage struct {
	ID              string      `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ConversationID  string      `json:"conversation_id" gorm:"type:varchar(36);not null"`
	MessageType     MessageType `json:"message_type" gorm:"type:enum('user','assistant');default:'user'"`
	Text            string      `json:"text" gorm:"type:text;not null"`
	SQLQuery        string      `json:"sql_query" gorm:"type:text"`
	SQLConfidence   *float64    `json:"sql_confidence" gorm:"type:decimal(3,2)"`
	ResultData      JSONMap     `json:"result_data" gorm:"type:json"`
	RowCount        *int        `json:"row_count"`
	ExecutionTimeMs *int        `json:"execution_time_ms"`
	ErrorMessage    string      `json:"error_message" gorm:"type:text"`
	ChartMetadata   JSONMap     `json:"chart_metadata" gorm:"type:json"`
	Provenance      JSONMap     `json:"provenance" gorm:"type:json"`
	CreatedAt       time.Time   `json:"created_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name
func (ChatMessage) TableName() string {
	return "chat_messages"
}

// MetadataType represents the type of schema metadata
type MetadataType string

const (
	MetadataTypeTable  MetadataType = "table"
	MetadataTypeColumn MetadataType = "column"
)

// DatabaseSchemaMetadata represents database schema information
type DatabaseSchemaMetadata struct {
	ID            string       `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	DatabaseAlias string       `json:"database_alias" gorm:"type:varchar(100);not null"`
	SchemaName    string       `json:"schema_name" gorm:"type:varchar(100)"`
	Table         string       `json:"table_name" gorm:"column:table_name;type:varchar(100);not null"`
	Column        string       `json:"column_name" gorm:"column:column_name;type:varchar(100)"`
	DataType      string       `json:"data_type" gorm:"type:varchar(50)"`
	IsNullable    *bool        `json:"is_nullable"`
	ColumnComment string       `json:"column_comment" gorm:"type:text"`
	SampleValues  JSONMap      `json:"sample_values" gorm:"type:json"`
	MetadataType  MetadataType `json:"metadata_type" gorm:"type:enum('table','column');not null"`
	Description   string       `json:"description" gorm:"type:text"`
	BusinessTerms JSONMap      `json:"business_terms" gorm:"type:json"`
	UpdatedAt     time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (DatabaseSchemaMetadata) TableName() string {
	return "database_schema_metadata"
}

// EntityType represents the type of entity for embeddings
type EntityType string

const (
	EntityTypeTable           EntityType = "table"
	EntityTypeColumn          EntityType = "column"
	EntityTypeBusinessEntity  EntityType = "business_entity"
	EntityTypeQueryTemplate   EntityType = "query_template"
)

// SchemaEmbedding represents vector embeddings for semantic search
type SchemaEmbedding struct {
	ID            string     `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	DatabaseAlias string     `json:"database_alias" gorm:"type:varchar(100);not null"`
	EntityType    EntityType `json:"entity_type" gorm:"type:enum('table','column','business_entity','query_template');not null"`
	EntityID      string     `json:"entity_id" gorm:"type:varchar(36);not null"`
	EntityText    string     `json:"entity_text" gorm:"type:text;not null"`
	Embedding     JSONMap    `json:"embedding" gorm:"type:json;not null"`
	ModelName     string     `json:"model_name" gorm:"type:varchar(100);default:'text-embedding-ada-002'"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (SchemaEmbedding) TableName() string {
	return "schema_embeddings"
}

// BusinessEntityType represents the type of business entity
type BusinessEntityType string

const (
	BusinessEntityTypeEntity    BusinessEntityType = "entity"
	BusinessEntityTypeMetric    BusinessEntityType = "metric"
	BusinessEntityTypeDimension BusinessEntityType = "dimension"
)

// BusinessEntity represents business-level entity mappings
type BusinessEntity struct {
	ID                 string             `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	DatabaseAlias      string             `json:"database_alias" gorm:"type:varchar(100);not null"`
	EntityName         string             `json:"entity_name" gorm:"type:varchar(255);not null"`
	EntityType         BusinessEntityType `json:"entity_type" gorm:"type:enum('entity','metric','dimension');not null"`
	Description        string             `json:"description" gorm:"type:text"`
	TableMappings      JSONMap            `json:"table_mappings" gorm:"type:json"`
	ColumnMappings     JSONMap            `json:"column_mappings" gorm:"type:json"`
	CalculationFormula string             `json:"calculation_formula" gorm:"type:text"`
	Synonyms           JSONMap            `json:"synonyms" gorm:"type:json"`
	Examples           JSONMap            `json:"examples" gorm:"type:json"`
	CreatedBy          string             `json:"created_by" gorm:"type:varchar(36)"`
	CreatedAt          time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time          `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (BusinessEntity) TableName() string {
	return "business_entities"
}

// QueryTemplate represents reusable query patterns
type QueryTemplate struct {
	ID                     string    `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	DatabaseAlias          string    `json:"database_alias" gorm:"type:varchar(100);not null"`
	TemplateName           string    `json:"template_name" gorm:"type:varchar(255);not null"`
	Description            string    `json:"description" gorm:"type:text"`
	NaturalLanguagePattern string    `json:"natural_language_pattern" gorm:"type:text"`
	SQLTemplate            string    `json:"sql_template" gorm:"type:text;not null"`
	ExampleQuestions       JSONMap   `json:"example_questions" gorm:"type:json"`
	Parameters             JSONMap   `json:"parameters" gorm:"type:json"`
	UsageCount             int       `json:"usage_count" gorm:"default:0"`
	SuccessRate            *float64  `json:"success_rate" gorm:"type:decimal(3,2)"`
	CreatedBy              string    `json:"created_by" gorm:"type:varchar(36)"`
	CreatedAt              time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (QueryTemplate) TableName() string {
	return "query_templates"
}

// GenerationType represents the type of AI generation
type GenerationType string

const (
	GenerationTypeSQL       GenerationType = "sql"
	GenerationTypeReport    GenerationType = "report"
	GenerationTypeNarrative GenerationType = "narrative"
	GenerationTypeChart     GenerationType = "chart"
)

// AIGenerationLog represents audit log for AI-generated content
type AIGenerationLog struct {
	ID              string         `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ConversationID  *string        `json:"conversation_id" gorm:"type:varchar(36)"`
	MessageID       *string        `json:"message_id" gorm:"type:varchar(36)"`
	GenerationType  GenerationType `json:"generation_type" gorm:"type:enum('sql','report','narrative','chart');not null"`
	InputPrompt     string         `json:"input_prompt" gorm:"type:text"`
	SystemPrompt    string         `json:"system_prompt" gorm:"type:text"`
	AIResponse      string         `json:"ai_response" gorm:"type:text"`
	ModelName       string         `json:"model_name" gorm:"type:varchar(100)"`
	TokensUsed      *int           `json:"tokens_used"`
	LatencyMs       *int           `json:"latency_ms"`
	ConfidenceScore *float64       `json:"confidence_score" gorm:"type:decimal(3,2)"`
	WasSuccessful   bool           `json:"was_successful" gorm:"default:true"`
	ErrorMessage    string         `json:"error_message" gorm:"type:text"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name
func (AIGenerationLog) TableName() string {
	return "ai_generation_log"
}

// SettingType represents the type of system setting
type SettingType string

const (
	SettingTypeString  SettingType = "string"
	SettingTypeNumber  SettingType = "number"
	SettingTypeBoolean SettingType = "boolean"
	SettingTypeJSON    SettingType = "json"
)

// SystemSetting represents system-wide settings
type SystemSetting struct {
	ID          string      `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	SettingKey  string      `json:"setting_key" gorm:"type:varchar(100);not null;uniqueIndex"`
	SettingValue string     `json:"setting_value" gorm:"type:text"`
	SettingType SettingType `json:"setting_type" gorm:"type:enum('string','number','boolean','json');default:'string'"`
	Description string      `json:"description" gorm:"type:text"`
	IsEncrypted bool        `json:"is_encrypted" gorm:"default:false"`
	UpdatedBy   string      `json:"updated_by" gorm:"type:varchar(36)"`
	UpdatedAt   time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (SystemSetting) TableName() string {
	return "system_settings"
}

// VectorArray is a helper type for vector operations
type VectorArray []float64

// Scan implements the sql.Scanner interface for vector arrays
func (v *VectorArray) Scan(value interface{}) error {
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

// Value implements the driver.Valuer interface for vector arrays
func (v VectorArray) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}
