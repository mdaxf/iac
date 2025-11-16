package debug

import (
	"encoding/json"
	"time"

	"github.com/mdaxf/iac/engine/types"
)

// EventType represents the type of debug event
type EventType string

const (
	EventTranCodeStart      EventType = "trancode.start"
	EventTranCodeComplete   EventType = "trancode.complete"
	EventTranCodeError      EventType = "trancode.error"
	EventFuncGroupStart     EventType = "funcgroup.start"
	EventFuncGroupComplete  EventType = "funcgroup.complete"
	EventFuncGroupRouting   EventType = "funcgroup.routing"
	EventFunctionStart      EventType = "function.start"
	EventFunctionComplete   EventType = "function.complete"
	EventFunctionError      EventType = "function.error"
	EventInputMapping       EventType = "input.mapping"
	EventOutputMapping      EventType = "output.mapping"
	EventDatabaseQuery      EventType = "database.query"
	EventScriptExecution    EventType = "script.execution"
	EventValidation         EventType = "validation"
	EventTransactionBegin   EventType = "transaction.begin"
	EventTransactionCommit  EventType = "transaction.commit"
	EventTransactionRollback EventType = "transaction.rollback"
)

// DebugEvent represents a single debug event during execution
type DebugEvent struct {
	ID            string                 `json:"id"`
	SessionID     string                 `json:"session_id"`
	Timestamp     time.Time              `json:"timestamp"`
	EventType     EventType              `json:"event_type"`
	Level         string                 `json:"level"` // DEBUG, INFO, WARNING, ERROR

	// Execution context
	TranCodeName    string `json:"trancode_name,omitempty"`
	TranCodeVersion string `json:"trancode_version,omitempty"`
	FuncGroupName   string `json:"funcgroup_name,omitempty"`
	FunctionName    string `json:"function_name,omitempty"`
	FunctionType    string `json:"function_type,omitempty"`

	// Execution details
	ExecutionStep   int                    `json:"execution_step"`
	ExecutionTime   time.Duration          `json:"execution_time,omitempty"`
	StartTime       time.Time              `json:"start_time,omitempty"`
	EndTime         time.Time              `json:"end_time,omitempty"`

	// Data
	Inputs          map[string]interface{} `json:"inputs,omitempty"`
	Outputs         map[string]interface{} `json:"outputs,omitempty"`
	RoutingValue    interface{}            `json:"routing_value,omitempty"`
	RoutingPath     string                 `json:"routing_path,omitempty"`

	// Additional context
	Message         string                 `json:"message,omitempty"`
	Error           string                 `json:"error,omitempty"`
	StackTrace      string                 `json:"stack_trace,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`

	// Performance metrics
	MemoryUsage     int64                  `json:"memory_usage,omitempty"`
	CPUTime         time.Duration          `json:"cpu_time,omitempty"`
}

// NewDebugEvent creates a new debug event
func NewDebugEvent(sessionID string, eventType EventType) *DebugEvent {
	return &DebugEvent{
		ID:        generateEventID(),
		SessionID: sessionID,
		Timestamp: time.Now(),
		EventType: eventType,
		Level:     "INFO",
		Metadata:  make(map[string]interface{}),
	}
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return time.Now().Format("20060102150405.000000")
}

// WithTranCode adds trancode context to the event
func (de *DebugEvent) WithTranCode(name, version string) *DebugEvent {
	de.TranCodeName = name
	de.TranCodeVersion = version
	return de
}

// WithFuncGroup adds function group context to the event
func (de *DebugEvent) WithFuncGroup(name string) *DebugEvent {
	de.FuncGroupName = name
	return de
}

// WithFunction adds function context to the event
func (de *DebugEvent) WithFunction(name, funcType string) *DebugEvent {
	de.FunctionName = name
	de.FunctionType = funcType
	return de
}

// WithInputs adds input data to the event
func (de *DebugEvent) WithInputs(inputs map[string]interface{}) *DebugEvent {
	de.Inputs = sanitizeData(inputs)
	return de
}

// WithOutputs adds output data to the event
func (de *DebugEvent) WithOutputs(outputs map[string]interface{}) *DebugEvent {
	de.Outputs = sanitizeData(outputs)
	return de
}

// WithRouting adds routing information to the event
func (de *DebugEvent) WithRouting(value interface{}, path string) *DebugEvent {
	de.RoutingValue = value
	de.RoutingPath = path
	return de
}

// WithTiming adds timing information to the event
func (de *DebugEvent) WithTiming(start, end time.Time) *DebugEvent {
	de.StartTime = start
	de.EndTime = end
	de.ExecutionTime = end.Sub(start)
	return de
}

// WithError adds error information to the event
func (de *DebugEvent) WithError(err error) *DebugEvent {
	if err != nil {
		de.Error = err.Error()
		de.Level = "ERROR"

		// If it's a BPMError, extract additional context
		if bpmErr, ok := err.(*types.BPMError); ok {
			de.StackTrace = bpmErr.StackTrace
			if bpmErr.Details != nil {
				for k, v := range bpmErr.Details {
					de.Metadata[k] = v
				}
			}
		}
	}
	return de
}

// WithMessage adds a descriptive message to the event
func (de *DebugEvent) WithMessage(message string) *DebugEvent {
	de.Message = message
	return de
}

// WithMetadata adds custom metadata to the event
func (de *DebugEvent) WithMetadata(key string, value interface{}) *DebugEvent {
	de.Metadata[key] = value
	return de
}

// SetLevel sets the log level of the event
func (de *DebugEvent) SetLevel(level string) *DebugEvent {
	de.Level = level
	return de
}

// ToJSON converts the event to JSON
func (de *DebugEvent) ToJSON() ([]byte, error) {
	return json.Marshal(de)
}

// ToSSEFormat formats the event for Server-Sent Events
func (de *DebugEvent) ToSSEFormat() string {
	jsonData, _ := de.ToJSON()
	return "data: " + string(jsonData) + "\n\n"
}

// sanitizeData removes sensitive information from data
func sanitizeData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	sanitized := make(map[string]interface{})
	sensitiveKeys := []string{"password", "token", "secret", "key", "credential"}

	for k, v := range data {
		// Check if key contains sensitive information
		isSensitive := false
		for _, sensitive := range sensitiveKeys {
			if contains(k, sensitive) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sanitized[k] = "***REDACTED***"
		} else {
			sanitized[k] = v
		}
	}

	return sanitized
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return indexOf(s, substr) >= 0
}

// toLower converts string to lowercase
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// indexOf finds the index of a substring
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ExecutionTrace represents a complete execution trace
type ExecutionTrace struct {
	SessionID     string        `json:"session_id"`
	TranCodeName  string        `json:"trancode_name"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	TotalDuration time.Duration `json:"total_duration"`
	Events        []*DebugEvent `json:"events"`
	Status        string        `json:"status"` // SUCCESS, FAILED, TIMEOUT
	ErrorMessage  string        `json:"error_message,omitempty"`
}

// NewExecutionTrace creates a new execution trace
func NewExecutionTrace(sessionID, tranCodeName string) *ExecutionTrace {
	return &ExecutionTrace{
		SessionID:    sessionID,
		TranCodeName: tranCodeName,
		StartTime:    time.Now(),
		Events:       make([]*DebugEvent, 0),
		Status:       "RUNNING",
	}
}

// AddEvent adds an event to the trace
func (et *ExecutionTrace) AddEvent(event *DebugEvent) {
	et.Events = append(et.Events, event)
}

// Complete marks the trace as complete
func (et *ExecutionTrace) Complete(status string) {
	et.EndTime = time.Now()
	et.TotalDuration = et.EndTime.Sub(et.StartTime)
	et.Status = status
}

// ToJSON converts the trace to JSON
func (et *ExecutionTrace) ToJSON() ([]byte, error) {
	return json.Marshal(et)
}

// GetSummary returns a summary of the execution
func (et *ExecutionTrace) GetSummary() map[string]interface{} {
	summary := map[string]interface{}{
		"session_id":     et.SessionID,
		"trancode_name":  et.TranCodeName,
		"start_time":     et.StartTime,
		"end_time":       et.EndTime,
		"total_duration": et.TotalDuration,
		"status":         et.Status,
		"total_events":   len(et.Events),
	}

	// Count events by type
	eventCounts := make(map[EventType]int)
	errorCount := 0

	for _, event := range et.Events {
		eventCounts[event.EventType]++
		if event.Level == "ERROR" {
			errorCount++
		}
	}

	summary["event_counts"] = eventCounts
	summary["error_count"] = errorCount

	return summary
}
