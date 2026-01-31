package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// =====================================================
// Hierarchical Integration Hub Models
// =====================================================

// HubMetadata stores flexible metadata for hub entities
type HubMetadata map[string]interface{}

// Scan implements sql.Scanner interface for JSONB
func (h *HubMetadata) Scan(value interface{}) error {
	if value == nil {
		*h = make(HubMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	if len(bytes) == 0 {
		*h = make(HubMetadata)
		return nil
	}

	return json.Unmarshal(bytes, h)
}

// Value implements driver.Valuer interface for JSONB
func (h HubMetadata) Value() (driver.Value, error) {
	if h == nil {
		return json.Marshal(make(map[string]interface{}))
	}
	return json.Marshal(h)
}

// HubConfig stores configuration data in JSONB
type HubConfig map[string]interface{}

// Scan implements sql.Scanner interface for JSONB
func (h *HubConfig) Scan(value interface{}) error {
	if value == nil {
		*h = make(HubConfig)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	if len(bytes) == 0 {
		*h = make(HubConfig)
		return nil
	}

	return json.Unmarshal(bytes, h)
}

// Value implements driver.Valuer interface for JSONB
func (h HubConfig) Value() (driver.Value, error) {
	if h == nil {
		return json.Marshal(make(map[string]interface{}))
	}
	return json.Marshal(h)
}

// RouteConditions stores array of routing conditions in JSONB
type RouteConditions []RouteCondition

// Scan implements sql.Scanner interface for JSONB
func (r *RouteConditions) Scan(value interface{}) error {
	if value == nil {
		*r = make(RouteConditions, 0)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	if len(bytes) == 0 {
		*r = make(RouteConditions, 0)
		return nil
	}

	return json.Unmarshal(bytes, r)
}

// Value implements driver.Valuer interface for JSONB
func (r RouteConditions) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]RouteCondition{})
	}
	return json.Marshal(r)
}

// FieldMappings stores array of field mappings in JSONB
type FieldMappings []FieldMapping

// Scan implements sql.Scanner interface for JSONB
func (f *FieldMappings) Scan(value interface{}) error {
	if value == nil {
		*f = make(FieldMappings, 0)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	if len(bytes) == 0 {
		*f = make(FieldMappings, 0)
		return nil
	}

	return json.Unmarshal(bytes, f)
}

// Value implements driver.Valuer interface for JSONB
func (f FieldMappings) Value() (driver.Value, error) {
	if f == nil {
		return json.Marshal([]FieldMapping{})
	}
	return json.Marshal(f)
}

// HubInstance represents a hub instance assignment
type HubInstance struct {
	ID              string      `json:"id" db:"id"`
	InstanceID      string      `json:"instance_id" db:"instance_id"`
	Name            string      `json:"name" db:"name"`
	Description     string      `json:"description" db:"description"`
	Enabled         bool        `json:"enabled" db:"enabled"`
	Metadata        HubMetadata `json:"metadata" db:"metadata"`
	Active          bool        `json:"active" db:"active"`
	ReferenceID     string      `json:"referenceid" db:"referenceid"`
	CreatedBy       string      `json:"createdby" db:"createdby"`
	CreatedOn       time.Time   `json:"createdon" db:"createdon"`
	ModifiedBy      string      `json:"modifiedby" db:"modifiedby"`
	ModifiedOn      time.Time   `json:"modifiedon" db:"modifiedon"`
	RowVersionStamp int         `json:"rowversionstamp" db:"rowversionstamp"`
}

// HubProtocolGroup represents protocol-level configuration group
type HubProtocolGroup struct {
	ID              string      `json:"id" db:"id"`
	InstanceID      string      `json:"instance_id" db:"instance_id"`
	Direction       string      `json:"direction" db:"direction"`
	Name            string      `json:"name" db:"name"`
	Description     string      `json:"description" db:"description"`
	Protocol        string      `json:"protocol" db:"protocol"`
	BaseConfig      HubConfig   `json:"base_config" db:"base_config"`
	MessageType     string      `json:"message_type" db:"message_type"`
	Timeout         int         `json:"timeout" db:"timeout"`
	RetryAttempts   int         `json:"retry_attempts" db:"retry_attempts"`
	RetryInterval   int         `json:"retry_interval" db:"retry_interval"`
	AuthType        string      `json:"auth_type" db:"auth_type"`
	AuthConfig      HubConfig   `json:"auth_config" db:"auth_config"`
	Enabled         bool        `json:"enabled" db:"enabled"`
	Metadata        HubMetadata `json:"metadata" db:"metadata"`
	Active          bool        `json:"active" db:"active"`
	ReferenceID     string      `json:"referenceid" db:"referenceid"`
	CreatedBy       string      `json:"createdby" db:"createdby"`
	CreatedOn       time.Time   `json:"createdon" db:"createdon"`
	ModifiedBy      string      `json:"modifiedby" db:"modifiedby"`
	ModifiedOn      time.Time   `json:"modifiedon" db:"modifiedon"`
	RowVersionStamp int         `json:"rowversionstamp" db:"rowversionstamp"`
}

// HubEndpoint represents individual endpoint configuration
type HubEndpoint struct {
	ID               string      `json:"id" db:"id"`
	ProtocolGroupID  string      `json:"protocol_group_id" db:"protocol_group_id"`
	Name             string      `json:"name" db:"name"`
	Description      string      `json:"description" db:"description"`
	EndpointURL      string      `json:"endpoint_url" db:"endpoint_url"`
	Port             int         `json:"port" db:"port"`
	Path             string      `json:"path" db:"path"`
	Method           string      `json:"method" db:"method"`
	OverrideConfig   HubConfig   `json:"override_config" db:"override_config"`
	MessageType      string      `json:"message_type" db:"message_type"`
	Timeout          int         `json:"timeout" db:"timeout"`
	RetryAttempts    int         `json:"retry_attempts" db:"retry_attempts"`
	RetryInterval    int         `json:"retry_interval" db:"retry_interval"`
	AuthType         string      `json:"auth_type" db:"auth_type"`
	AuthConfig       HubConfig   `json:"auth_config" db:"auth_config"`
	QueueConfig      HubConfig   `json:"queue_config" db:"queue_config"`
	FileConfig       HubConfig   `json:"file_config" db:"file_config"`
	ValidateSchema   bool        `json:"validate_schema" db:"validate_schema"`
	SchemaDefinition string      `json:"schema_definition" db:"schema_definition"`
	Enabled          bool        `json:"enabled" db:"enabled"`
	Metadata         HubMetadata `json:"metadata" db:"metadata"`
	Active           bool        `json:"active" db:"active"`
	ReferenceID      string      `json:"referenceid" db:"referenceid"`
	CreatedBy        string      `json:"createdby" db:"createdby"`
	CreatedOn        time.Time   `json:"createdon" db:"createdon"`
	ModifiedBy       string      `json:"modifiedby" db:"modifiedby"`
	ModifiedOn       time.Time   `json:"modifiedon" db:"modifiedon"`
	RowVersionStamp  int         `json:"rowversionstamp" db:"rowversionstamp"`
}

// HubRoute represents message routing configuration
type HubRoute struct {
	ID                    string           `json:"id" db:"id"`
	SourceEndpointID      string           `json:"source_endpoint_id" db:"source_endpoint_id"`
	DestinationEndpointID string           `json:"destination_endpoint_id" db:"destination_endpoint_id"`
	Name                  string           `json:"name" db:"name"`
	Description           string           `json:"description" db:"description"`
	SourceFilter          string           `json:"source_filter" db:"source_filter"`
	Conditions            RouteConditions  `json:"conditions" db:"conditions"`
	TransformationType    string           `json:"transformation_type" db:"transformation_type"`
	Transformation        string           `json:"transformation" db:"transformation"`
	FieldMappings         FieldMappings    `json:"field_mappings" db:"field_mappings"`
	Priority              int              `json:"priority" db:"priority"`
	AsyncMode             bool             `json:"async_mode" db:"async_mode"`
	OnError               string           `json:"on_error" db:"on_error"`
	DeadLetterQueue       string           `json:"dead_letter_queue" db:"dead_letter_queue"`
	MaxRetries            int              `json:"max_retries" db:"max_retries"`
	Enabled               bool             `json:"enabled" db:"enabled"`
	Metadata              HubMetadata      `json:"metadata" db:"metadata"`
	Active                bool             `json:"active" db:"active"`
	ReferenceID           string           `json:"referenceid" db:"referenceid"`
	CreatedBy             string           `json:"createdby" db:"createdby"`
	CreatedOn             time.Time        `json:"createdon" db:"createdon"`
	ModifiedBy            string           `json:"modifiedby" db:"modifiedby"`
	ModifiedOn            time.Time        `json:"modifiedon" db:"modifiedon"`
	RowVersionStamp       int              `json:"rowversionstamp" db:"rowversionstamp"`
}

// RouteCondition defines a routing condition
type RouteCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// FieldMapping represents a field mapping for transformations
// Note: Using existing FieldMapping from datahub schema
type FieldMapping struct {
	SourcePath    string      `json:"source_path"`
	TargetPath    string      `json:"target_path"`
	DataType      string      `json:"data_type"`
	DefaultValue  interface{} `json:"default_value"`
	Required      bool        `json:"required"`
	TransformFunc string      `json:"transform_func"`
}

// HubEndpointJob links jobs to specific endpoints
type HubEndpointJob struct {
	ID                 string      `json:"id" db:"id"`
	EndpointID         string      `json:"endpoint_id" db:"endpoint_id"`
	JobID              string      `json:"job_id" db:"job_id"`
	JobType            string      `json:"job_type" db:"job_type"`
	CronExpression     string      `json:"cron_expression" db:"cron_expression"`
	IntervalSeconds    int         `json:"interval_seconds" db:"interval_seconds"`
	TriggerEventType   string      `json:"trigger_event_type" db:"trigger_event_type"`
	TriggerEventSource string      `json:"trigger_event_source" db:"trigger_event_source"`
	TriggerEventFilter string      `json:"trigger_event_filter" db:"trigger_event_filter"`
	Timeout            int         `json:"timeout" db:"timeout"`
	MaxConcurrent      int         `json:"max_concurrent" db:"max_concurrent"`
	RetryOnFailure     bool        `json:"retry_on_failure" db:"retry_on_failure"`
	MaxRetries         int         `json:"max_retries" db:"max_retries"`
	Parameters         HubConfig   `json:"parameters" db:"parameters"`
	Enabled            bool        `json:"enabled" db:"enabled"`
	Metadata           HubMetadata `json:"metadata" db:"metadata"`
	Active             bool        `json:"active" db:"active"`
	ReferenceID        string      `json:"referenceid" db:"referenceid"`
	CreatedBy          string      `json:"createdby" db:"createdby"`
	CreatedOn          time.Time   `json:"createdon" db:"createdon"`
	ModifiedBy         string      `json:"modifiedby" db:"modifiedby"`
	ModifiedOn         time.Time   `json:"modifiedon" db:"modifiedon"`
	RowVersionStamp    int         `json:"rowversionstamp" db:"rowversionstamp"`
}

// ResolvedEndpointConfig represents endpoint configuration after inheritance resolution
type ResolvedEndpointConfig struct {
	Endpoint           *HubEndpoint       `json:"endpoint"`
	ProtocolGroup      *HubProtocolGroup  `json:"protocol_group"`
	Instance           *HubInstance       `json:"instance"`
	FinalTimeout       int                `json:"final_timeout"`
	FinalRetryAttempts int                `json:"final_retry_attempts"`
	FinalRetryInterval int                `json:"final_retry_interval"`
	FinalMessageType   string             `json:"final_message_type"`
	FinalAuthType      string             `json:"final_auth_type"`
	FinalAuthConfig    map[string]interface{} `json:"final_auth_config"`
	FinalConfig        map[string]interface{} `json:"final_config"`
}

// HubMigrationLog tracks migration from old flat structure to hierarchical
type HubMigrationLog struct {
	ID               string      `json:"id" db:"id"`
	OldHubID         string      `json:"old_hub_id" db:"old_hub_id"`
	NewInstanceID    string      `json:"new_instance_id" db:"new_instance_id"`
	MigrationStatus  string      `json:"migration_status" db:"migration_status"`
	MigrationDetails HubMetadata `json:"migration_details" db:"migration_details"`
	ErrorMessage     string      `json:"error_message" db:"error_message"`
	MigratedAt       time.Time   `json:"migrated_at" db:"migrated_at"`
	CreatedBy        string      `json:"createdby" db:"createdby"`
	CreatedOn        time.Time   `json:"createdon" db:"createdon"`
}

// HierarchicalTreeNode represents a node in the hierarchical tree
type HierarchicalTreeNode struct {
	ID       string                  `json:"id"`
	Type     string                  `json:"type"`
	Label    string                  `json:"label"`
	Icon     string                  `json:"icon"`
	Data     interface{}             `json:"data"`
	Children []*HierarchicalTreeNode `json:"children,omitempty"`
}

// DirectionNode represents a direction node in the tree
type DirectionNode struct {
	Direction  string `json:"direction"`
	InstanceID string `json:"instance_id"`
}
