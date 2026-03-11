package hubexecutor

import (
	"context"
	"sync"
	"time"

	"github.com/mdaxf/iac/models"
)

// ExecutorStatus represents the status of a hub executor
type ExecutorStatus string

const (
	StatusStopped  ExecutorStatus = "stopped"
	StatusStarting ExecutorStatus = "starting"
	StatusRunning  ExecutorStatus = "running"
	StatusStopping ExecutorStatus = "stopping"
	StatusError    ExecutorStatus = "error"
)

// ProtocolType represents supported protocol types
type ProtocolType string

const (
	ProtocolREST     ProtocolType = "REST API"
	ProtocolSOAP     ProtocolType = "SOAP API"
	ProtocolGraphQL  ProtocolType = "GraphQL"
	ProtocolMQTT     ProtocolType = "MQTT Client"
	ProtocolKafka    ProtocolType = "Kafka"
	ProtocolActiveMQ ProtocolType = "ActiveMQ"
	ProtocolOPCUA    ProtocolType = "OPC UA"
	ProtocolTCP      ProtocolType = "TCP"
	ProtocolSignalR  ProtocolType = "SignalR"
	// ProtocolMCP is the Model Context Protocol (JSON-RPC 2.0 over HTTP / stdio)
	ProtocolMCP ProtocolType = "MCP"
)

// TriggerMethod represents action trigger methods
type TriggerMethod string

const (
	TriggerInterval   TriggerMethod = "interval"
	TriggerStartup    TriggerMethod = "startup"
	TriggerShutdown   TriggerMethod = "shutdown"
	TriggerDataChange TriggerMethod = "datachange"
)

// OPCTag represents an OPC UA tag configuration
type OPCTag struct {
	ID             string  `json:"id"`
	TagName        string  `json:"tagName"`
	Address        string  `json:"address"`
	DataType       string  `json:"dataType"`
	IsArray        bool    `json:"isArray"`
	ArraySize      int     `json:"arraySize,omitempty"`
	Permission     string  `json:"permission"`
	SampleInterval int     `json:"sampleInterval"` // milliseconds
	Deadband       float64 `json:"deadband,omitempty"`
	Description    string  `json:"description,omitempty"`
	Unit           string  `json:"unit,omitempty"`
	ScaleFactor    float64 `json:"scaleFactor,omitempty"`
	Offset         float64 `json:"offset,omitempty"`
	Enabled        bool    `json:"enabled"`
}

// OPCTagValue represents a current OPC tag value
type OPCTagValue struct {
	TagID     string
	TagName   string
	Value     interface{}
	Quality   int
	Timestamp time.Time
}

// ActionTrigger represents an action trigger configuration
type ActionTrigger struct {
	Method          TriggerMethod `json:"method"`
	IntervalMs      int           `json:"intervalMs,omitempty"`
	MonitoredTags   []string      `json:"monitoredTags,omitempty"`
	ConditionScript string        `json:"conditionScript,omitempty"`
}

// Destination represents a destination configuration
type Destination struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // Job, Execute Transcode, Route to Outbound, Custom Script
	Target    string                 `json:"target,omitempty"`
	MappingID string                 `json:"mappingId,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

// MessagePayload represents an incoming message payload
type MessagePayload struct {
	TransactionID string
	Protocol      ProtocolType
	Direction     string // "inbound" or "outbound"
	EndpointID    string
	Path          string
	Method        string
	Headers       map[string]string
	Body          []byte
	ContentType   string
	Metadata      map[string]interface{}
	Timestamp     time.Time
}

// ExecutionResult represents the result of destination execution
type ExecutionResult struct {
	Success   bool
	Message   string
	Data      interface{}
	Error     error
	Duration  time.Duration
	Timestamp time.Time
}

// ProtocolHandler is the interface for all protocol handlers
type ProtocolHandler interface {
	// Start starts the protocol handler
	Start(ctx context.Context) error

	// Stop stops the protocol handler
	Stop(ctx context.Context) error

	// Status returns the current status
	Status() ExecutorStatus

	// Name returns the handler name/identifier
	Name() string

	// Type returns the protocol type
	Type() ProtocolType
}

// DestinationExecutor is the interface for executing destinations
type DestinationExecutor interface {
	// Execute executes a destination with the given payload
	Execute(ctx context.Context, dest Destination, payload *MessagePayload) (*ExecutionResult, error)
}

// HistoryRecorder is the interface for recording integration hub history
type HistoryRecorder interface {
	RecordTransaction(ctx context.Context, history *models.IntHubHistory) (string, error)
	CompleteTransaction(ctx context.Context, id string, response string, mappedData string, responseStatus int) error
	FailTransaction(ctx context.Context, id string, errorMessage string) error
	RecordAction(ctx context.Context, action *models.IntHubActionHistory) error
}

// HubExecutorConfig holds configuration for a hub executor
type HubExecutorConfig struct {
	Hub              *models.IntegrationHub
	HubName          string
	DestinationExec  DestinationExecutor
	HistoryRecorder  HistoryRecorder
	OnMessageReceive func(payload *MessagePayload)
	OnError          func(err error, source string)
}

// HubExecutorState holds the runtime state of a hub executor
type HubExecutorState struct {
	mu             sync.RWMutex
	Status         ExecutorStatus
	StartTime      *time.Time
	LastActivity   *time.Time
	MessageCount   int64
	ErrorCount     int64
	ActiveHandlers int
	Handlers       map[string]ProtocolHandler
	OPCTagValues   map[string]*OPCTagValue // tag ID -> current value
}

// NewHubExecutorState creates a new executor state
func NewHubExecutorState() *HubExecutorState {
	return &HubExecutorState{
		Status:       StatusStopped,
		Handlers:     make(map[string]ProtocolHandler),
		OPCTagValues: make(map[string]*OPCTagValue),
	}
}

// SetStatus sets the executor status
func (s *HubExecutorState) SetStatus(status ExecutorStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
}

// GetStatus gets the executor status
func (s *HubExecutorState) GetStatus() ExecutorStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status
}

// IncrementMessageCount increments the message counter
func (s *HubExecutorState) IncrementMessageCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MessageCount++
	now := time.Now()
	s.LastActivity = &now
}

// IncrementErrorCount increments the error counter
func (s *HubExecutorState) IncrementErrorCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ErrorCount++
}

// UpdateOPCTagValue updates an OPC tag value
func (s *HubExecutorState) UpdateOPCTagValue(tagID string, value *OPCTagValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.OPCTagValues[tagID] = value
}

// GetOPCTagValue gets an OPC tag value
func (s *HubExecutorState) GetOPCTagValue(tagID string) *OPCTagValue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.OPCTagValues[tagID]
}

// ExecutorInfo represents information about a running executor
type ExecutorInfo struct {
	HubName        string         `json:"hub_name"`
	HubID          string         `json:"hub_id"`
	Version        int            `json:"version"`
	Status         ExecutorStatus `json:"status"`
	StartTime      *time.Time     `json:"start_time,omitempty"`
	LastActivity   *time.Time     `json:"last_activity,omitempty"`
	MessageCount   int64          `json:"message_count"`
	ErrorCount     int64          `json:"error_count"`
	ActiveHandlers int            `json:"active_handlers"`
	Handlers       []HandlerInfo  `json:"handlers"`
}

// HandlerInfo represents information about a protocol handler
type HandlerInfo struct {
	Name     string         `json:"name"`
	Type     ProtocolType   `json:"type"`
	Status   ExecutorStatus `json:"status"`
	Protocol string         `json:"protocol"`
	Address  string         `json:"address,omitempty"`
}
