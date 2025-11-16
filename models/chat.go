package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Conversation represents a chat conversation thread
type Conversation struct {
	ID               string `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	Title            string `json:"title" gorm:"type:varchar(255);not null"`
	UserID           string `json:"userid" gorm:"type:varchar(36);not null"`
	DatabaseAlias    string `json:"databasealias" gorm:"type:varchar(100)"`
	AutoExecuteQuery bool   `json:"autoexecutequery" gorm:"default:true"`

	// Relationships
	Messages []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
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
	ConversationID  string      `json:"conversationid" gorm:"type:varchar(36);not null"`
	MessageType     MessageType `json:"messagetype" gorm:"type:enum('user','assistant');default:'user'"`
	Text            string      `json:"text" gorm:"type:text;not null"`
	SQLQuery        string      `json:"sqlquery" gorm:"type:text"`
	SQLConfidence   *float64    `json:"sqlconfidence" gorm:"type:decimal(3,2)"`
	ResultData      JSONMap     `json:"resultdata" gorm:"type:json"`
	RowCount        *int        `json:"rowcount"`
	ExecutionTimeMs *int        `json:"executiontimems"`
	ErrorMessage    string      `json:"errormessage" gorm:"type:text"`
	ChartMetadata   JSONMap     `json:"chartmetadata" gorm:"type:json"`
	Provenance      JSONMap     `json:"provenance" gorm:"type:json"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (ChatMessage) TableName() string {
	return "chatmessages"
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
	DatabaseAlias string       `json:"databasealias" gorm:"type:varchar(100);not null"`
	SchemaName    string       `json:"schemaname" gorm:"type:varchar(100)"`
	Table         string       `json:"tablename" gorm:"column:tablename;type:varchar(100);not null"`
	Column        string       `json:"columnname" gorm:"column:columnname;type:varchar(100)"`
	DataType      string       `json:"datatype" gorm:"type:varchar(50)"`
	IsNullable    *bool        `json:"isnullable"`
	ColumnComment string       `json:"columncomment" gorm:"type:text"`
	SampleValues  JSONMap      `json:"samplevalues" gorm:"type:json"`
	MetadataType  MetadataType `json:"metadatatype" gorm:"type:enum('table','column');not null"`
	Description   string       `json:"description" gorm:"type:text"`
	BusinessTerms JSONMap      `json:"businessterms" gorm:"type:json"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (DatabaseSchemaMetadata) TableName() string {
	return "databaseschemametadata"
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
	DatabaseAlias string     `json:"databasealias" gorm:"type:varchar(100);not null"`
	EntityType    EntityType `json:"entitytype" gorm:"type:enum('table','column','business_entity','query_template');not null"`
	EntityID      string     `json:"entityid" gorm:"type:varchar(36);not null"`
	EntityText    string     `json:"entitytext" gorm:"type:text;not null"`
	Embedding     JSONMap    `json:"embedding" gorm:"type:json;not null"`
	ModelName     string     `json:"modelname" gorm:"type:varchar(100);default:'text-embedding-ada-002'"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (SchemaEmbedding) TableName() string {
	return "schemaembeddings"
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
	DatabaseAlias      string             `json:"databasealias" gorm:"type:varchar(100);not null"`
	EntityName         string             `json:"entityname" gorm:"type:varchar(255);not null"`
	EntityType         BusinessEntityType `json:"entitytype" gorm:"type:enum('entity','metric','dimension');not null"`
	Description        string             `json:"description" gorm:"type:text"`
	TableMappings      JSONMap            `json:"tablemappings" gorm:"type:json"`
	ColumnMappings     JSONMap            `json:"columnmappings" gorm:"type:json"`
	CalculationFormula string             `json:"calculationformula" gorm:"type:text"`
	Synonyms           JSONMap            `json:"synonyms" gorm:"type:json"`
	Examples           JSONMap            `json:"examples" gorm:"type:json"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (BusinessEntity) TableName() string {
	return "businessentities"
}

// QueryTemplate represents reusable query patterns
type QueryTemplate struct {
	ID                     string   `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	DatabaseAlias          string   `json:"databasealias" gorm:"type:varchar(100);not null"`
	TemplateName           string   `json:"templatename" gorm:"type:varchar(255);not null"`
	Description            string   `json:"description" gorm:"type:text"`
	NaturalLanguagePattern string   `json:"naturallanguagepattern" gorm:"type:text"`
	SQLTemplate            string   `json:"sqltemplate" gorm:"type:text;not null"`
	ExampleQuestions       JSONMap  `json:"examplequestions" gorm:"type:json"`
	Parameters             JSONMap  `json:"parameters" gorm:"type:json"`
	UsageCount             int      `json:"usagecount" gorm:"default:0"`
	SuccessRate            *float64 `json:"successrate" gorm:"type:decimal(3,2)"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (QueryTemplate) TableName() string {
	return "querytemplates"
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
	ConversationID  *string        `json:"conversationid" gorm:"type:varchar(36)"`
	MessageID       *string        `json:"messageid" gorm:"type:varchar(36)"`
	GenerationType  GenerationType `json:"generationtype" gorm:"type:enum('sql','report','narrative','chart');not null"`
	InputPrompt     string         `json:"inputprompt" gorm:"type:text"`
	SystemPrompt    string         `json:"systemprompt" gorm:"type:text"`
	AIResponse      string         `json:"airesponse" gorm:"type:text"`
	ModelName       string         `json:"modelname" gorm:"type:varchar(100)"`
	TokensUsed      *int           `json:"tokensused"`
	LatencyMs       *int           `json:"latencyms"`
	ConfidenceScore *float64       `json:"confidencescore" gorm:"type:decimal(3,2)"`
	WasSuccessful   bool           `json:"wassuccessful" gorm:"default:true"`
	ErrorMessage    string         `json:"errormessage" gorm:"type:text"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (AIGenerationLog) TableName() string {
	return "aigenerationlog"
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
	ID           string      `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	SettingKey   string      `json:"settingkey" gorm:"type:varchar(100);not null;uniqueIndex"`
	SettingValue string      `json:"settingvalue" gorm:"type:text"`
	SettingType  SettingType `json:"settingtype" gorm:"type:enum('string','number','boolean','json');default:'string'"`
	Description  string      `json:"description" gorm:"type:text"`
	IsEncrypted  bool        `json:"isencrypted" gorm:"default:false"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (SystemSetting) TableName() string {
	return "systemsettings"
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
