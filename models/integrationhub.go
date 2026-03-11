package models

import (
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// =====================================================
// Simplified Single-Document Integration Hub Model
// =====================================================

// IntegrationHubStatus represents the status of an integration hub configuration
type IntegrationHubStatus string

const (
	IntegrationHubStatusDev        IntegrationHubStatus = "dev"
	IntegrationHubStatusTest       IntegrationHubStatus = "test"
	IntegrationHubStatusProduction IntegrationHubStatus = "production"
	IntegrationHubStatusDeprecated IntegrationHubStatus = "deprecated"
)

// IntegrationHub represents a complete integration hub configuration
// stored as a single document in the Integration_Hub collection
type IntegrationHub struct {
	ID              string                 `json:"_id" bson:"_id"`
	Name            string                 `json:"name" bson:"name"` // Unique name for grouping versions
	Description     string                 `json:"description" bson:"description"`
	InstanceID      string                 `json:"instance_id" bson:"instance_id"`
	Version         int                    `json:"version" bson:"version"`       // Version number for this configuration
	IsDefault       bool                   `json:"is_default" bson:"is_default"` // Only one default per instance_name
	Status          IntegrationHubStatus   `json:"status" bson:"status"`         // dev, test, production, deprecated
	Enabled         bool                   `json:"enabled" bson:"enabled"`
	Inbound         *HubDirection          `json:"inbound" bson:"inbound"`
	Outbound        *HubDirection          `json:"outbound" bson:"outbound"`
	Routes          []HubSimpleRoute       `json:"routes" bson:"routes"`
	Jobs            []HubSimpleJob         `json:"jobs" bson:"jobs"`
	Metadata        map[string]interface{} `json:"metadata" bson:"metadata"`
	Active          bool                   `json:"active" bson:"active"`
	CreatedBy       string                 `json:"createdby" bson:"createdby"`
	CreatedOn       time.Time              `json:"createdon" bson:"createdon"`
	ModifiedBy      string                 `json:"modifiedby" bson:"modifiedby"`
	ModifiedOn      time.Time              `json:"modifiedon" bson:"modifiedon"`
	RowVersionStamp int                    `json:"rowversionstamp" bson:"rowversionstamp"`
}

// HubDirection represents inbound or outbound configuration
type HubDirection struct {
	Enabled        bool                     `json:"enabled" bson:"enabled"`
	ProtocolGroups []HubSimpleProtocolGroup `json:"protocol_groups" bson:"protocol_groups"`
}

// MCPProtocolConfig holds MCP server connection settings for Integration Hub protocol groups.
// Used when Protocol == "MCP" to route outbound messages to an MCP tool via JSON-RPC 2.0.
type MCPProtocolConfig struct {
	ServerURL  string            `json:"server_url" bson:"server_url"`   // MCP server base URL (e.g. http://localhost:3000)
	MCPPath    string            `json:"mcp_path" bson:"mcp_path"`       // JSON-RPC endpoint path (default "/mcp")
	ToolName   string            `json:"tool_name" bson:"tool_name"`     // MCP tool to invoke
	Headers    map[string]string `json:"headers" bson:"headers"`         // Extra HTTP headers
	TimeoutSec int               `json:"timeout_sec" bson:"timeout_sec"` // Per-call timeout in seconds (default 30)
}

// HubSimpleProtocolGroup represents a protocol group with its endpoints
type HubSimpleProtocolGroup struct {
	ID            string                 `json:"id" bson:"id"`
	Name          string                 `json:"name" bson:"name"`
	Description   string                 `json:"description" bson:"description"`
	Protocol      string                 `json:"protocol" bson:"protocol"` // HTTP, HTTPS, REST, SOAP, MQTT, AMQP, Kafka, MCP, etc.
	MessageType   string                 `json:"message_type" bson:"message_type"`
	Timeout       int                    `json:"timeout" bson:"timeout"`
	RetryAttempts int                    `json:"retry_attempts" bson:"retry_attempts"`
	RetryInterval int                    `json:"retry_interval" bson:"retry_interval"`
	AuthType      string                 `json:"auth_type" bson:"auth_type"`
	AuthConfig    map[string]interface{} `json:"auth_config" bson:"auth_config"`
	BaseConfig    map[string]interface{} `json:"base_config" bson:"base_config"`

	// Message Bus Configuration (for MQTT, AMQP, Kafka, etc.)
	BrokerConfig *MessageBrokerConfig `json:"broker_config,omitempty" bson:"broker_config,omitempty"`

	// MCP Configuration (for Protocol == "MCP")
	MCPConfig *MCPProtocolConfig `json:"mcp_config,omitempty" bson:"mcp_config,omitempty"`

	Endpoints []HubSimpleEndpoint    `json:"endpoints" bson:"endpoints"`
	Enabled   bool                   `json:"enabled" bson:"enabled"`
	Metadata  map[string]interface{} `json:"metadata" bson:"metadata"`
}

// BrokerTopic represents a message broker topic with tracking settings
type BrokerTopic struct {
	Topic        string `json:"topic" bson:"topic"`                 // Topic/queue name
	TrackHistory *bool  `json:"track_history" bson:"track_history"` // Enable/disable history tracking (defaults to true if nil)
}

// UnmarshalJSON handles both string (legacy) and object formats for BrokerTopic
func (bt *BrokerTopic) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		bt.Topic = s
		bt.TrackHistory = nil
		return nil
	}

	// Try to unmarshal as a struct
	type alias BrokerTopic
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*bt = BrokerTopic(a)
	return nil
}

// UnmarshalBSONValue handles both string (legacy) and object formats for BrokerTopic in BSON
func (bt *BrokerTopic) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t == bsontype.String {
		var s string
		if err := bson.UnmarshalValue(t, data, &s); err != nil {
			return err
		}
		bt.Topic = s
		bt.TrackHistory = nil
		return nil
	}

	if t == bsontype.EmbeddedDocument {
		type alias BrokerTopic
		var a alias
		if err := bson.Unmarshal(data, &a); err != nil {
			return err
		}
		*bt = BrokerTopic(a)
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into BrokerTopic", t)
}

// MessageBrokerConfig represents message broker connection settings
type MessageBrokerConfig struct {
	BrokerURL        string                 `json:"broker_url" bson:"broker_url"`   // e.g., mqtt://broker.example.com:1883
	BrokerHost       string                 `json:"broker_host" bson:"broker_host"` // e.g., broker.example.com
	BrokerPort       int                    `json:"broker_port" bson:"broker_port"` // e.g., 1883, 5672, 9092
	Topics           []BrokerTopic          `json:"topics" bson:"topics"`           // Topics/queues to subscribe to (inbound) or publish to (outbound)
	Username         string                 `json:"username,omitempty" bson:"username,omitempty"`
	Password         string                 `json:"password,omitempty" bson:"password,omitempty"`
	ClientID         string                 `json:"client_id,omitempty" bson:"client_id,omitempty"`
	UseTLS           bool                   `json:"use_tls" bson:"use_tls"`
	TLSConfig        map[string]interface{} `json:"tls_config,omitempty" bson:"tls_config,omitempty"`         // Certificate paths, etc.
	QoS              int                    `json:"qos,omitempty" bson:"qos,omitempty"`                       // Quality of Service (MQTT)
	ConsumerGroup    string                 `json:"consumer_group,omitempty" bson:"consumer_group,omitempty"` // Kafka consumer group
	AutoCommit       bool                   `json:"auto_commit" bson:"auto_commit"`                           // Kafka auto-commit
	MaxRetries       int                    `json:"max_retries" bson:"max_retries"`
	RetryInterval    int                    `json:"retry_interval" bson:"retry_interval"`             // Milliseconds
	KeepAlive        int                    `json:"keep_alive,omitempty" bson:"keep_alive,omitempty"` // Keep-alive interval in seconds
	AdditionalConfig map[string]interface{} `json:"additional_config,omitempty" bson:"additional_config,omitempty"`
}

// ShouldTrackHistory returns true if history tracking is enabled for a BrokerTopic
// Defaults to true if TrackHistory is nil
func (bt *BrokerTopic) ShouldTrackHistory() bool {
	if bt.TrackHistory == nil {
		return true
	}
	return *bt.TrackHistory
}

// HubSimpleEndpoint represents an individual endpoint
type HubSimpleEndpoint struct {
	ID          string `json:"id" bson:"id"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`

	// Endpoint Configuration
	EndpointURL string `json:"endpoint_url" bson:"endpoint_url"` // For outbound: external URL to call. For inbound webservice: leave empty (uses Path)
	Port        int    `json:"port" bson:"port"`
	Path        string `json:"path" bson:"path"`     // For inbound webservice: /api/receive/data. For outbound: appended to EndpointURL
	Method      string `json:"method" bson:"method"` // GET, POST, PUT, DELETE, etc.

	// Message Configuration
	MessageType      string `json:"message_type" bson:"message_type"` // JSON, XML, CSV, etc.
	ValidateSchema   bool   `json:"validate_schema" bson:"validate_schema"`
	SchemaDefinition string `json:"schema_definition" bson:"schema_definition"`

	// Connection Settings
	Timeout       int `json:"timeout" bson:"timeout"`
	RetryAttempts int `json:"retry_attempts" bson:"retry_attempts"`
	RetryInterval int `json:"retry_interval" bson:"retry_interval"`

	// Authentication
	AuthType   string                 `json:"auth_type" bson:"auth_type"`
	AuthConfig map[string]interface{} `json:"auth_config" bson:"auth_config"`

	// Additional Configuration
	OverrideConfig map[string]interface{} `json:"override_config" bson:"override_config"`
	QueueConfig    map[string]interface{} `json:"queue_config" bson:"queue_config"`
	FileConfig     map[string]interface{} `json:"file_config" bson:"file_config"`

	// Inbound Message Handler (for inbound endpoints only)
	Handler *MessageHandler `json:"handler,omitempty" bson:"handler,omitempty"`

	Enabled      bool                   `json:"enabled" bson:"enabled"`
	TrackHistory *bool                  `json:"track_history" bson:"track_history"` // Enable/disable history tracking (defaults to true if nil)
	Metadata     map[string]interface{} `json:"metadata" bson:"metadata"`
}

// UnmarshalJSON handles both string (legacy/ID only) and object formats for HubSimpleEndpoint
func (e *HubSimpleEndpoint) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		e.ID = s
		return nil
	}

	// Try to unmarshal as a struct
	type alias HubSimpleEndpoint
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*e = HubSimpleEndpoint(a)
	return nil
}

// UnmarshalBSONValue handles both string (legacy/ID only) and object formats for HubSimpleEndpoint in BSON
func (e *HubSimpleEndpoint) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t == bsontype.String {
		var s string
		if err := bson.UnmarshalValue(t, data, &s); err != nil {
			return err
		}
		e.ID = s
		return nil
	}

	if t == bsontype.EmbeddedDocument {
		type alias HubSimpleEndpoint
		var a alias
		if err := bson.Unmarshal(data, &a); err != nil {
			return err
		}
		*e = HubSimpleEndpoint(a)
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into HubSimpleEndpoint", t)
}

// ShouldTrackHistory returns true if history tracking is enabled for an endpoint
// Defaults to true if TrackHistory is nil
func (e *HubSimpleEndpoint) ShouldTrackHistory() bool {
	if e.TrackHistory == nil {
		return true
	}
	return *e.TrackHistory
}

// MessageHandler defines how to process received messages (for inbound endpoints)
type MessageHandler struct {
	HandlerType     string                 `json:"handler_type" bson:"handler_type"`                             // route, store, transform, script, webhook
	RouteIDs        []string               `json:"route_ids,omitempty" bson:"route_ids,omitempty"`               // References to HubSimpleRoute IDs
	StorageTarget   string                 `json:"storage_target,omitempty" bson:"storage_target,omitempty"`     // Database table/collection name
	TransformScript string                 `json:"transform_script,omitempty" bson:"transform_script,omitempty"` // Transformation logic
	WebhookURL      string                 `json:"webhook_url,omitempty" bson:"webhook_url,omitempty"`           // Webhook to call after processing
	OnSuccess       string                 `json:"on_success,omitempty" bson:"on_success,omitempty"`             // Action on success: ack, forward, store
	OnFailure       string                 `json:"on_failure,omitempty" bson:"on_failure,omitempty"`             // Action on failure: retry, dlq, ignore
	MaxRetries      int                    `json:"max_retries" bson:"max_retries"`
	DeadLetterQueue string                 `json:"dead_letter_queue,omitempty" bson:"dead_letter_queue,omitempty"`
	Parameters      map[string]interface{} `json:"parameters,omitempty" bson:"parameters,omitempty"` // Additional handler-specific params
}

// HubSimpleRoute represents a message route
type HubSimpleRoute struct {
	ID                 string                   `json:"id" bson:"id"`
	Name               string                   `json:"name" bson:"name"`
	Description        string                   `json:"description" bson:"description"`
	Source             string                   `json:"source" bson:"source"`
	SourceFilter       string                   `json:"source_filter" bson:"source_filter"`
	Destination        string                   `json:"destination" bson:"destination"`
	Conditions         []map[string]interface{} `json:"conditions" bson:"conditions"`
	TransformationType string                   `json:"transformation_type" bson:"transformation_type"`
	Transformation     string                   `json:"transformation" bson:"transformation"`
	FieldMappings      []map[string]interface{} `json:"field_mappings" bson:"field_mappings"`
	Priority           int                      `json:"priority" bson:"priority"`
	AsyncMode          bool                     `json:"async_mode" bson:"async_mode"`
	OnError            string                   `json:"on_error" bson:"on_error"`
	DeadLetterQueue    string                   `json:"dead_letter_queue" bson:"dead_letter_queue"`
	MaxRetries         int                      `json:"max_retries" bson:"max_retries"`
	Enabled            bool                     `json:"enabled" bson:"enabled"`
	Metadata           map[string]interface{}   `json:"metadata" bson:"metadata"`
}

// HubSimpleJob represents a scheduled job
type HubSimpleJob struct {
	ID                 string                 `json:"id" bson:"id"`
	Name               string                 `json:"name" bson:"name"`
	Description        string                 `json:"description" bson:"description"`
	JobType            string                 `json:"job_type" bson:"job_type"`
	EndpointRef        string                 `json:"endpoint_ref" bson:"endpoint_ref"`
	CronExpression     string                 `json:"cron_expression" bson:"cron_expression"`
	IntervalSeconds    int                    `json:"interval_seconds" bson:"interval_seconds"`
	TriggerEventType   string                 `json:"trigger_event_type" bson:"trigger_event_type"`
	TriggerEventSource string                 `json:"trigger_event_source" bson:"trigger_event_source"`
	TriggerEventFilter string                 `json:"trigger_event_filter" bson:"trigger_event_filter"`
	Timeout            int                    `json:"timeout" bson:"timeout"`
	MaxConcurrent      int                    `json:"max_concurrent" bson:"max_concurrent"`
	RetryOnFailure     bool                   `json:"retry_on_failure" bson:"retry_on_failure"`
	MaxRetries         int                    `json:"max_retries" bson:"max_retries"`
	Parameters         map[string]interface{} `json:"parameters" bson:"parameters"`
	Enabled            bool                   `json:"enabled" bson:"enabled"`
	Metadata           map[string]interface{} `json:"metadata" bson:"metadata"`
}
