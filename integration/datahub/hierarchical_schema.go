// Package datahub provides hierarchical integration hub schema definitions
package datahub

import "time"

// =====================================================
// Hierarchical Integration Hub Schemas
// =====================================================

// HubInstance represents a hub instance assignment
// Links to system InstanceID (com.InstanceID) for managing integration configurations
type HubInstance struct {
	ID              string                 `json:"id" db:"id"`
	InstanceID      string                 `json:"instance_id" db:"instance_id"`
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	Enabled         bool                   `json:"enabled" db:"enabled"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	Active          bool                   `json:"active" db:"active"`
	ReferenceID     string                 `json:"referenceid" db:"referenceid"`
	CreatedBy       string                 `json:"createdby" db:"createdby"`
	CreatedOn       time.Time              `json:"createdon" db:"createdon"`
	ModifiedBy      string                 `json:"modifiedby" db:"modifiedby"`
	ModifiedOn      time.Time              `json:"modifiedon" db:"modifiedon"`
	RowVersionStamp int                    `json:"rowversionstamp" db:"rowversionstamp"`
}

// HubProtocolGroup represents protocol-level configuration group
// Provides base configuration that child endpoints inherit
type HubProtocolGroup struct {
	ID              string                 `json:"id" db:"id"`
	InstanceID      string                 `json:"instance_id" db:"instance_id"`
	Direction       string                 `json:"direction" db:"direction"` // "inbound" or "outbound"
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	Protocol        string                 `json:"protocol" db:"protocol"` // REST, SOAP, MQTT, Kafka, etc.

	// Base configuration inherited by endpoints
	BaseConfig      map[string]interface{} `json:"base_config" db:"base_config"`

	// Protocol-specific defaults
	MessageType     string                 `json:"message_type" db:"message_type"`
	Timeout         int                    `json:"timeout" db:"timeout"`
	RetryAttempts   int                    `json:"retry_attempts" db:"retry_attempts"`
	RetryInterval   int                    `json:"retry_interval" db:"retry_interval"`

	// Authentication defaults
	AuthType        string                 `json:"auth_type" db:"auth_type"`
	AuthConfig      map[string]interface{} `json:"auth_config" db:"auth_config"`

	Enabled         bool                   `json:"enabled" db:"enabled"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	Active          bool                   `json:"active" db:"active"`
	ReferenceID     string                 `json:"referenceid" db:"referenceid"`
	CreatedBy       string                 `json:"createdby" db:"createdby"`
	CreatedOn       time.Time              `json:"createdon" db:"createdon"`
	ModifiedBy      string                 `json:"modifiedby" db:"modifiedby"`
	ModifiedOn      time.Time              `json:"modifiedon" db:"modifiedon"`
	RowVersionStamp int                    `json:"rowversionstamp" db:"rowversionstamp"`
}

// HubEndpoint represents individual endpoint configuration
// Can override protocol group settings
type HubEndpoint struct {
	ID                string                 `json:"id" db:"id"`
	ProtocolGroupID   string                 `json:"protocol_group_id" db:"protocol_group_id"`
	Name              string                 `json:"name" db:"name"`
	Description       string                 `json:"description" db:"description"`

	// Endpoint-specific configuration
	EndpointURL       string                 `json:"endpoint_url" db:"endpoint_url"`
	Port              int                    `json:"port" db:"port"`
	Path              string                 `json:"path" db:"path"`
	Method            string                 `json:"method" db:"method"` // GET, POST, PUT, DELETE, etc.

	// Override settings (NULL/zero means inherit from protocol_group)
	OverrideConfig    map[string]interface{} `json:"override_config" db:"override_config"`
	MessageType       string                 `json:"message_type" db:"message_type"`
	Timeout           int                    `json:"timeout" db:"timeout"`
	RetryAttempts     int                    `json:"retry_attempts" db:"retry_attempts"`
	RetryInterval     int                    `json:"retry_interval" db:"retry_interval"`

	// Authentication overrides
	AuthType          string                 `json:"auth_type" db:"auth_type"`
	AuthConfig        map[string]interface{} `json:"auth_config" db:"auth_config"`

	// Protocol-specific configs
	QueueConfig       map[string]interface{} `json:"queue_config" db:"queue_config"` // For MQTT, AMQP, Kafka, RabbitMQ
	FileConfig        map[string]interface{} `json:"file_config" db:"file_config"`   // For FTP, SFTP, File

	// Schema validation
	ValidateSchema    bool                   `json:"validate_schema" db:"validate_schema"`
	SchemaDefinition  string                 `json:"schema_definition" db:"schema_definition"`

	Enabled           bool                   `json:"enabled" db:"enabled"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	Active            bool                   `json:"active" db:"active"`
	ReferenceID       string                 `json:"referenceid" db:"referenceid"`
	CreatedBy         string                 `json:"createdby" db:"createdby"`
	CreatedOn         time.Time              `json:"createdon" db:"createdon"`
	ModifiedBy        string                 `json:"modifiedby" db:"modifiedby"`
	ModifiedOn        time.Time              `json:"modifiedon" db:"modifiedon"`
	RowVersionStamp   int                    `json:"rowversionstamp" db:"rowversionstamp"`
}

// HubRoute represents message routing configuration between endpoints
type HubRoute struct {
	ID                    string           `json:"id" db:"id"`
	SourceEndpointID      string           `json:"source_endpoint_id" db:"source_endpoint_id"`
	DestinationEndpointID string           `json:"destination_endpoint_id" db:"destination_endpoint_id"`
	Name                  string           `json:"name" db:"name"`
	Description           string           `json:"description" db:"description"`

	// Routing logic
	SourceFilter          string           `json:"source_filter" db:"source_filter"` // JSONPath or XPath
	Conditions            []RouteCondition `json:"conditions" db:"conditions"`

	// Transformation
	TransformationType    string           `json:"transformation_type" db:"transformation_type"` // None, XSLT, JavaScript, JSONata, Custom
	Transformation        string           `json:"transformation" db:"transformation"`
	FieldMappings         []FieldMapping   `json:"field_mappings" db:"field_mappings"`

	// Execution
	Priority              int              `json:"priority" db:"priority"`
	AsyncMode             bool             `json:"async_mode" db:"async_mode"`

	// Error handling
	OnError               string           `json:"on_error" db:"on_error"` // Retry, Skip, SendToDeadLetter
	DeadLetterQueue       string           `json:"dead_letter_queue" db:"dead_letter_queue"`
	MaxRetries            int              `json:"max_retries" db:"max_retries"`

	Enabled               bool                   `json:"enabled" db:"enabled"`
	Metadata              map[string]interface{} `json:"metadata" db:"metadata"`
	Active                bool                   `json:"active" db:"active"`
	ReferenceID           string                 `json:"referenceid" db:"referenceid"`
	CreatedBy             string                 `json:"createdby" db:"createdby"`
	CreatedOn             time.Time              `json:"createdon" db:"createdon"`
	ModifiedBy            string                 `json:"modifiedby" db:"modifiedby"`
	ModifiedOn            time.Time              `json:"modifiedon" db:"modifiedon"`
	RowVersionStamp       int                    `json:"rowversionstamp" db:"rowversionstamp"`
}

// RouteCondition defines a routing condition
type RouteCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, contains, regex, etc.
	Value    interface{} `json:"value"`
}

// HubEndpointJob links jobs to specific endpoints
type HubEndpointJob struct {
	ID                  string                 `json:"id" db:"id"`
	EndpointID          string                 `json:"endpoint_id" db:"endpoint_id"`
	JobID               string                 `json:"job_id" db:"job_id"`
	JobType             string                 `json:"job_type" db:"job_type"` // Manual, Scheduled, Triggered

	// Schedule configuration
	CronExpression      string                 `json:"cron_expression" db:"cron_expression"`
	IntervalSeconds     int                    `json:"interval_seconds" db:"interval_seconds"`

	// Trigger configuration
	TriggerEventType    string                 `json:"trigger_event_type" db:"trigger_event_type"`
	TriggerEventSource  string                 `json:"trigger_event_source" db:"trigger_event_source"`
	TriggerEventFilter  string                 `json:"trigger_event_filter" db:"trigger_event_filter"`

	// Execution
	Timeout             int                    `json:"timeout" db:"timeout"`
	MaxConcurrent       int                    `json:"max_concurrent" db:"max_concurrent"`
	RetryOnFailure      bool                   `json:"retry_on_failure" db:"retry_on_failure"`
	MaxRetries          int                    `json:"max_retries" db:"max_retries"`

	Parameters          map[string]interface{} `json:"parameters" db:"parameters"`

	Enabled             bool                   `json:"enabled" db:"enabled"`
	Metadata            map[string]interface{} `json:"metadata" db:"metadata"`
	Active              bool                   `json:"active" db:"active"`
	ReferenceID         string                 `json:"referenceid" db:"referenceid"`
	CreatedBy           string                 `json:"createdby" db:"createdby"`
	CreatedOn           time.Time              `json:"createdon" db:"createdon"`
	ModifiedBy          string                 `json:"modifiedby" db:"modifiedby"`
	ModifiedOn          time.Time              `json:"modifiedon" db:"modifiedon"`
	RowVersionStamp     int                    `json:"rowversionstamp" db:"rowversionstamp"`
}

// ResolvedEndpointConfig represents endpoint configuration after inheritance resolution
// Used to get final effective configuration combining protocol group defaults and endpoint overrides
type ResolvedEndpointConfig struct {
	Endpoint      *HubEndpoint       `json:"endpoint"`
	ProtocolGroup *HubProtocolGroup  `json:"protocol_group"`
	Instance      *HubInstance       `json:"instance"`

	// Computed fields after inheritance
	FinalTimeout      int                    `json:"final_timeout"`
	FinalRetryAttempts int                   `json:"final_retry_attempts"`
	FinalRetryInterval int                   `json:"final_retry_interval"`
	FinalMessageType  string                 `json:"final_message_type"`
	FinalAuthType     string                 `json:"final_auth_type"`
	FinalAuthConfig   map[string]interface{} `json:"final_auth_config"`
	FinalConfig       map[string]interface{} `json:"final_config"`
}

// HubMigrationLog tracks migration from old flat structure to hierarchical
type HubMigrationLog struct {
	ID               string                 `json:"id" db:"id"`
	OldHubID         string                 `json:"old_hub_id" db:"old_hub_id"`
	NewInstanceID    string                 `json:"new_instance_id" db:"new_instance_id"`
	MigrationStatus  string                 `json:"migration_status" db:"migration_status"` // pending, completed, failed
	MigrationDetails map[string]interface{} `json:"migration_details" db:"migration_details"`
	ErrorMessage     string                 `json:"error_message" db:"error_message"`
	MigratedAt       time.Time              `json:"migrated_at" db:"migrated_at"`
	CreatedBy        string                 `json:"createdby" db:"createdby"`
	CreatedOn        time.Time              `json:"createdon" db:"createdon"`
}

// =====================================================
// Helper Types for Tree Representation
// =====================================================

// HierarchicalTreeNode represents a node in the hierarchical tree
// Used for UI tree representation
type HierarchicalTreeNode struct {
	ID       string                  `json:"id"`
	Type     string                  `json:"type"` // instance, direction, protocol_group, endpoint
	Label    string                  `json:"label"`
	Icon     string                  `json:"icon"`
	Data     interface{}             `json:"data"` // Actual data (HubInstance, HubProtocolGroup, HubEndpoint, or DirectionNode)
	Children []*HierarchicalTreeNode `json:"children,omitempty"`
}

// DirectionNode represents a direction node in the tree (inbound/outbound)
type DirectionNode struct {
	Direction  string `json:"direction"` // "inbound" or "outbound"
	InstanceID string `json:"instance_id"`
}

// =====================================================
// Constants for Protocol Types
// =====================================================

const (
	ProtocolREST      = "REST"
	ProtocolSOAP      = "SOAP"
	ProtocolMQTT      = "MQTT"
	ProtocolKafka     = "Kafka"
	ProtocolRabbitMQ  = "RabbitMQ"
	ProtocolActiveMQ  = "ActiveMQ"
	ProtocolAMQP      = "AMQP"
	ProtocolTCP       = "TCP"
	ProtocolGraphQL   = "GraphQL"
	ProtocolWebSocket = "WebSocket"
	ProtocolgRPC      = "gRPC"
	ProtocolFTP       = "FTP"
	ProtocolSFTP      = "SFTP"
	ProtocolFile      = "File"
	ProtocolDatabase  = "Database"
	ProtocolEmail     = "Email"
	ProtocolSMS       = "SMS"
)

// =====================================================
// Constants for Directions
// =====================================================

const (
	DirectionInbound  = "inbound"
	DirectionOutbound = "outbound"
)

// =====================================================
// Constants for Job Types
// =====================================================

const (
	JobTypeManual    = "Manual"
	JobTypeScheduled = "Scheduled"
	JobTypeTriggered = "Triggered"
)

// =====================================================
// Constants for Error Handling
// =====================================================

const (
	OnErrorRetry           = "Retry"
	OnErrorSkip            = "Skip"
	OnErrorSendToDeadLetter = "SendToDeadLetter"
)

// =====================================================
// Constants for Transformation Types
// =====================================================

const (
	TransformationNone      = "None"
	TransformationXSLT      = "XSLT"
	TransformationJavaScript = "JavaScript"
	TransformationJSONata   = "JSONata"
	TransformationCustom    = "Custom"
)

// =====================================================
// Constants for Migration Status
// =====================================================

const (
	MigrationStatusPending   = "pending"
	MigrationStatusCompleted = "completed"
	MigrationStatusFailed    = "failed"
)
