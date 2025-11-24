package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

// Conversation represents a chat conversation thread
type Conversation struct {
	ID               string `json:"id" gorm:"column:id;primaryKey;type:varchar(36);default:(UUID())"`
	Title            string `json:"title" gorm:"column:title;type:varchar(255);not null"`
	UserID           string `json:"userid" gorm:"column:userid;type:varchar(36);not null"`
	DatabaseAlias    string `json:"databasealias" gorm:"column:databasealias;type:varchar(100)"`
	AutoExecuteQuery bool   `json:"autoexecutequery" gorm:"column:autoexecutequery;default:true"`

	// Relationships
	Messages []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE"`

	// Standard IAC audit fields (must be at end)
	Active          bool         `json:"active" gorm:"column:active;default:true"`
	ReferenceID     string       `json:"referenceid" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string       `json:"createdby" gorm:"column:createdby;type:varchar(45)"`
	CreatedOn       sql.NullTime `json:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string       `json:"modifiedby" gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn      sql.NullTime `json:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int          `json:"rowversionstamp" gorm:"column:rowversionstamp;default:1"`
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
	ID              string      `json:"id" gorm:"column:id;primaryKey;type:varchar(36);default:(UUID())"`
	ConversationID  string      `json:"conversationid" gorm:"column:conversationid;type:varchar(36);not null"`
	MessageType     MessageType `json:"messagetype" gorm:"column:messagetype;type:enum('user','assistant');default:'user'"`
	Text            string      `json:"text" gorm:"column:text;type:text;not null"`
	SQLQuery        string      `json:"sqlquery" gorm:"column:sqlquery;type:text"`
	SQLConfidence   *float64    `json:"sqlconfidence" gorm:"column:sqlconfidence;type:decimal(3,2)"`
	ResultData      JSONMap     `json:"resultdata" gorm:"column:resultdata;type:json"`
	RowCount        *int        `json:"rowcount" gorm:"column:rowcount"`
	ExecutionTimeMs *int        `json:"executiontimems" gorm:"column:executiontimems"`
	ErrorMessage    string      `json:"errormessage" gorm:"column:errormessage;type:text"`
	ChartMetadata   JSONMap     `json:"chartmetadata" gorm:"column:chartmetadata;type:json"`
	Provenance      JSONMap     `json:"provenance" gorm:"column:provenance;type:json"`

	// Standard IAC audit fields (must be at end)
	Active          bool         `json:"active" gorm:"column:active;default:true"`
	ReferenceID     string       `json:"referenceid" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string       `json:"createdby" gorm:"column:createdby;type:varchar(45)"`
	CreatedOn       sql.NullTime `json:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string       `json:"modifiedby" gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn      sql.NullTime `json:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int          `json:"rowversionstamp" gorm:"column:rowversionstamp;default:1"`
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
	ID            string       `json:"id" gorm:"column:id;primaryKey;type:varchar(36);default:(UUID())"`
	DatabaseAlias string       `json:"database_alias" gorm:"column:databasealias;type:varchar(100);not null"`
	SchemaName    string       `json:"schema_name" gorm:"column:schemaname;type:varchar(100)"`
	Table         string       `json:"table_name" gorm:"column:tablename;type:varchar(100);not null"`
	Column        string       `json:"column_name" gorm:"column:columnname;type:varchar(100)"`
	DataType      string       `json:"data_type" gorm:"column:datatype;type:varchar(50)"`
	IsNullable    *bool        `json:"is_nullable" gorm:"column:isnullable"`
	IsPrimaryKey  *bool        `json:"is_primary_key" gorm:"column:is_primary_key"`
	IsForeignKey  *bool        `json:"is_foreign_key" gorm:"column:is_foreign_key"`
	ColumnComment string       `json:"column_comment" gorm:"column:columncomment;type:text"`
	SampleValues  JSONMap      `json:"sample_values" gorm:"column:samplevalues;type:json"`
	MetadataType  MetadataType `json:"entity_type" gorm:"column:metadatatype;type:enum('table','column');not null"`
	Description   string       `json:"description" gorm:"column:description;type:text"`
	BusinessName  string       `json:"business_name" gorm:"column:business_name;type:varchar(255)"`
	BusinessTerms JSONMap      `json:"business_terms" gorm:"column:businessterms;type:json"`

	// Vector embedding fields for semantic search
	Embedding            VectorArray  `json:"embedding,omitempty" gorm:"column:embedding;type:json"`
	EmbeddingModel       string       `json:"embedding_model,omitempty" gorm:"column:embedding_model;type:varchar(100)"`
	EmbeddingGeneratedAt sql.NullTime `json:"embedding_generated_at,omitempty" gorm:"column:embedding_generated_at"`

	// Standard IAC audit fields (must be at end)
	Active          bool         `json:"active" gorm:"column:active;default:true"`
	ReferenceID     string       `json:"reference_id" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string       `json:"created_by" gorm:"column:createdby;type:varchar(45)"`
	CreatedOn       sql.NullTime `json:"created_at" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string       `json:"modified_by" gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn      sql.NullTime `json:"updated_at" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int          `json:"row_version_stamp" gorm:"column:rowversionstamp;default:1"`
}

// TableName specifies the table name
func (DatabaseSchemaMetadata) TableName() string {
	return "databaseschemametadata"
}

// EntityType represents the type of entity for embeddings
type EntityType string

const (
	EntityTypeTable          EntityType = "table"
	EntityTypeColumn         EntityType = "column"
	EntityTypeBusinessEntity EntityType = "business_entity"
	EntityTypeQueryTemplate  EntityType = "query_template"
)

// SchemaEmbedding represents vector embeddings for semantic search
type SchemaEmbedding struct {
	ID            string     `json:"id" gorm:"column:id;primaryKey;type:varchar(36);default:(UUID())"`
	DatabaseAlias string     `json:"database_alias" gorm:"column:databasealias;type:varchar(100);not null"`
	EntityType    EntityType `json:"entity_type" gorm:"column:entitytype;type:enum('table','column','business_entity','query_template');not null"`
	EntityID      string     `json:"entity_id" gorm:"column:entityid;type:varchar(36);not null"`
	EntityText    string     `json:"entity_text" gorm:"column:entitytext;type:text;not null"`
	Embedding     JSONMap    `json:"embedding" gorm:"column:embedding;type:json;not null"`
	ModelName     string     `json:"model_name" gorm:"column:modelname;type:varchar(100);default:'text-embedding-ada-002'"`

	// Standard IAC audit fields (must be at end)
	Active          bool         `json:"active" gorm:"column:active;default:true"`
	ReferenceID     string       `json:"reference_id" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string       `json:"created_by" gorm:"column:createdby;type:varchar(45)"`
	CreatedOn       sql.NullTime `json:"created_at" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string       `json:"modified_by" gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn      sql.NullTime `json:"updated_at" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int          `json:"row_version_stamp" gorm:"column:rowversionstamp;default:1"`
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
	ID                 string             `json:"id" gorm:"column:id;primaryKey;type:varchar(36);default:(UUID())"`
	DatabaseAlias      string             `json:"database_alias" gorm:"column:databasealias;type:varchar(100);not null"`
	EntityName         string             `json:"entity_name" gorm:"column:entityname;type:varchar(255);not null"`
	EntityType         BusinessEntityType `json:"entity_type" gorm:"column:entitytype;type:enum('entity','metric','dimension');not null"`
	Description        string             `json:"description" gorm:"column:description;type:text"`
	TableMappings      JSONMap            `json:"table_mappings" gorm:"column:tablemappings;type:json"`
	ColumnMappings     JSONMap            `json:"column_mappings" gorm:"column:columnmappings;type:json"`
	CalculationFormula string             `json:"calculation_formula" gorm:"column:calculationformula;type:text"`
	Synonyms           json.RawMessage    `json:"synonyms" gorm:"column:synonyms;type:json"`
	Examples           json.RawMessage    `json:"examples" gorm:"column:examples;type:json"`

	// Vector embedding fields for semantic search
	Embedding            VectorArray  `json:"embedding,omitempty" gorm:"column:embedding;type:json"`
	EmbeddingModel       string       `json:"embedding_model,omitempty" gorm:"column:embedding_model;type:varchar(100)"`
	EmbeddingGeneratedAt sql.NullTime `json:"embedding_generated_at,omitempty" gorm:"column:embedding_generated_at"`

	// Standard IAC audit fields (must be at end)
	Active          bool         `json:"active" gorm:"column:active;default:true"`
	ReferenceID     string       `json:"reference_id" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string       `json:"created_by" gorm:"column:createdby;type:varchar(45)"`
	CreatedOn       sql.NullTime `json:"created_at" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string       `json:"modified_by" gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn      sql.NullTime `json:"updated_at" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int          `json:"row_version_stamp" gorm:"column:rowversionstamp;default:1"`
}

// TableName specifies the table name
func (BusinessEntity) TableName() string {
	return "businessentities"
}

// QueryTemplate represents reusable query patterns
type QueryTemplate struct {
	ID                     string          `json:"id" gorm:"column:id;primaryKey;type:varchar(36);default:(UUID())"`
	DatabaseAlias          string          `json:"databasealias" gorm:"column:databasealias;type:varchar(100);not null"`
	TemplateName           string          `json:"templatename" gorm:"column:templatename;type:varchar(255);not null"`
	Description            string          `json:"description" gorm:"column:description;type:text"`
	NaturalLanguagePattern string          `json:"naturallanguagepattern" gorm:"column:naturallanguagepattern;type:text"`
	SQLTemplate            string          `json:"sqltemplate" gorm:"column:sqltemplate;type:text;not null"`
	ExampleQuestions       json.RawMessage `json:"examplequestions" gorm:"column:examplequestions;type:json"`
	Parameters             JSONMap         `json:"parameters" gorm:"column:parameters;type:json"`
	UsageCount             int             `json:"usagecount" gorm:"column:usagecount;default:0"`
	SuccessRate            *float64        `json:"successrate" gorm:"column:successrate;type:decimal(3,2)"`

	// Standard IAC audit fields (must be at end)
	Active          bool         `json:"active" gorm:"column:active;default:true"`
	ReferenceID     string       `json:"referenceid" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string       `json:"createdby" gorm:"column:createdby;type:varchar(45)"`
	CreatedOn       sql.NullTime `json:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string       `json:"modifiedby" gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn      sql.NullTime `json:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int          `json:"rowversionstamp" gorm:"column:rowversionstamp;default:1"`
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
	ID              string         `json:"id" gorm:"column:id;primaryKey;type:varchar(36);default:(UUID())"`
	ConversationID  *string        `json:"conversationid" gorm:"column:conversationid;type:varchar(36)"`
	MessageID       *string        `json:"messageid" gorm:"column:messageid;type:varchar(36)"`
	GenerationType  GenerationType `json:"generationtype" gorm:"column:generationtype;type:enum('sql','report','narrative','chart');not null"`
	InputPrompt     string         `json:"inputprompt" gorm:"column:inputprompt;type:text"`
	SystemPrompt    string         `json:"systemprompt" gorm:"column:systemprompt;type:text"`
	AIResponse      string         `json:"airesponse" gorm:"column:airesponse;type:text"`
	ModelName       string         `json:"modelname" gorm:"column:modelname;type:varchar(100)"`
	TokensUsed      *int           `json:"tokensused" gorm:"column:tokensused"`
	LatencyMs       *int           `json:"latencyms" gorm:"column:latencyms"`
	ConfidenceScore *float64       `json:"confidencescore" gorm:"column:confidencescore;type:decimal(3,2)"`
	WasSuccessful   bool           `json:"wassuccessful" gorm:"column:wassuccessful;default:true"`
	ErrorMessage    string         `json:"errormessage" gorm:"column:errormessage;type:text"`

	// Standard IAC audit fields (must be at end)
	Active          bool         `json:"active" gorm:"column:active;default:true"`
	ReferenceID     string       `json:"referenceid" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string       `json:"createdby" gorm:"column:createdby;type:varchar(45)"`
	CreatedOn       sql.NullTime `json:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string       `json:"modifiedby" gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn      sql.NullTime `json:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int          `json:"rowversionstamp" gorm:"column:rowversionstamp;default:1"`
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
	ID           string      `json:"id" gorm:"column:id;primaryKey;type:varchar(36);default:(UUID())"`
	SettingKey   string      `json:"settingkey" gorm:"column:settingkey;type:varchar(100);not null;uniqueIndex"`
	SettingValue string      `json:"settingvalue" gorm:"column:settingvalue;type:text"`
	SettingType  SettingType `json:"settingtype" gorm:"column:settingtype;type:enum('string','number','boolean','json');default:'string'"`
	Description  string      `json:"description" gorm:"column:description;type:text"`
	IsEncrypted  bool        `json:"isencrypted" gorm:"column:isencrypted;default:false"`

	// Standard IAC audit fields (must be at end)
	Active          bool         `json:"active" gorm:"column:active;default:true"`
	ReferenceID     string       `json:"referenceid" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string       `json:"createdby" gorm:"column:createdby;type:varchar(45)"`
	CreatedOn       sql.NullTime `json:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string       `json:"modifiedby" gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn      sql.NullTime `json:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int          `json:"rowversionstamp" gorm:"column:rowversionstamp;default:1"`
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
