package datahub

import (
	"encoding/json"
	"time"
)

// MessageEnvelope represents a universal message container for all protocols
type MessageEnvelope struct {
	ID            string                 `json:"id"`
	Protocol      string                 `json:"protocol"` // REST, SOAP, GraphQL, TCP, MQTT, Kafka, etc.
	Source        string                 `json:"source"`
	Destination   string                 `json:"destination"`
	Timestamp     time.Time              `json:"timestamp"`
	ContentType   string                 `json:"content_type"` // application/json, application/xml, text/plain, etc.
	Headers       map[string]interface{} `json:"headers,omitempty"`
	Body          interface{}            `json:"body"`
	OriginalBody  []byte                 `json:"original_body,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	TransformPath []string               `json:"transform_path,omitempty"` // Track transformation history
}

// MappingDefinition defines how to transform messages between schemas
type MappingDefinition struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	SourceProtocol  string              `json:"source_protocol"`
	SourceSchema    string              `json:"source_schema"`
	TargetProtocol  string              `json:"target_protocol"`
	TargetSchema    string              `json:"target_schema"`
	Mappings        []FieldMapping      `json:"mappings"`
	Transformations []TransformRule     `json:"transformations,omitempty"`
	Conditions      []MappingCondition  `json:"conditions,omitempty"`
	Priority        int                 `json:"priority,omitempty"`
	Active          bool                `json:"active"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

// FieldMapping defines a single field transformation
type FieldMapping struct {
	SourcePath      string      `json:"source_path"`      // JSONPath or XPath expression
	TargetPath      string      `json:"target_path"`      // JSONPath or XPath expression
	DataType        string      `json:"data_type"`        // string, int, float, bool, date, array, object
	DefaultValue    interface{} `json:"default_value,omitempty"`
	Required        bool        `json:"required"`
	TransformFunc   string      `json:"transform_func,omitempty"` // Built-in function name
	CustomScript    string      `json:"custom_script,omitempty"`  // JavaScript/Lua script for complex transforms
}

// TransformRule defines a transformation to apply to the message
type TransformRule struct {
	Type        string                 `json:"type"` // filter, enrich, aggregate, split, merge
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Order       int                    `json:"order"`
}

// MappingCondition defines when a mapping should be applied
type MappingCondition struct {
	Field    string      `json:"field"`    // Field to check
	Operator string      `json:"operator"` // eq, ne, gt, lt, contains, regex, exists
	Value    interface{} `json:"value"`
	LogicOp  string      `json:"logic_op,omitempty"` // and, or (for chaining conditions)
}

// ProtocolAdapter interface for all protocol implementations
type ProtocolAdapter interface {
	// Send sends a message through this protocol
	Send(envelope *MessageEnvelope) error

	// Receive receives a message from this protocol (blocking or with timeout)
	Receive(timeout time.Duration) (*MessageEnvelope, error)

	// GetProtocolName returns the protocol name
	GetProtocolName() string

	// Initialize initializes the adapter with configuration
	Initialize(config map[string]interface{}) error

	// Close closes the adapter and releases resources
	Close() error

	// Health checks the health of the adapter
	Health() error
}

// DataHub is the central message transformation and routing hub
type DataHub struct {
	adapters         map[string]ProtocolAdapter
	mappings         map[string]*MappingDefinition
	transformEngine  *TransformEngine
	routingRules     []*RoutingRule
	messageHistory   *MessageHistory
	enabled          bool
}

// RoutingRule defines how messages should be routed
type RoutingRule struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Source      string              `json:"source"`      // Source protocol/endpoint
	Destination string              `json:"destination"` // Destination protocol/endpoint
	Conditions  []MappingCondition  `json:"conditions,omitempty"`
	MappingID   string              `json:"mapping_id,omitempty"` // Optional transformation
	Priority    int                 `json:"priority"`
	Active      bool                `json:"active"`
}

// TransformEngine handles message transformation logic
type TransformEngine struct {
	builtInFunctions map[string]TransformFunction
	customScripts    map[string]string
}

// TransformFunction is a built-in transformation function
type TransformFunction func(input interface{}, params map[string]interface{}) (interface{}, error)

// MessageHistory tracks message transformations for audit
type MessageHistory struct {
	maxEntries int
	entries    []MessageHistoryEntry
}

// MessageHistoryEntry represents a single transformation event
type MessageHistoryEntry struct {
	MessageID     string                 `json:"message_id"`
	Timestamp     time.Time              `json:"timestamp"`
	SourceProto   string                 `json:"source_protocol"`
	TargetProto   string                 `json:"target_protocol"`
	MappingID     string                 `json:"mapping_id,omitempty"`
	Success       bool                   `json:"success"`
	Error         string                 `json:"error,omitempty"`
	Duration      time.Duration          `json:"duration"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// MarshalJSON custom marshaller for MessageEnvelope to handle binary data
func (m *MessageEnvelope) MarshalJSON() ([]byte, error) {
	type Alias MessageEnvelope
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	})
}
