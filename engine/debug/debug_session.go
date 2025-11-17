package debug

import (
	"fmt"
	"sync"
	"time"
)

// DebugSession represents an active debug session
type DebugSession struct {
	SessionID    string
	TranCodeName string
	UserID       string
	Description  string
	Status       string // "running", "completed", "failed"
	StartTime    time.Time
	EndTime      time.Time
	EventCount   int
	Events       []*DebugEvent
	mu           sync.RWMutex
	enabled      bool
}

// NewDebugSession creates a new debug session
func NewDebugSession(sessionID, tranCodeName, userID string) *DebugSession {
	return &DebugSession{
		SessionID:    sessionID,
		TranCodeName: tranCodeName,
		UserID:       userID,
		Status:       "initialized",
		Events:       make([]*DebugEvent, 0),
		enabled:      false,
	}
}

// Start starts the debug session
func (ds *DebugSession) Start() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.Status = "running"
	ds.StartTime = time.Now()
	ds.enabled = true
}

// Stop stops the debug session
func (ds *DebugSession) Stop() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.EndTime = time.Now()
	if ds.Status == "running" {
		ds.Status = "completed"
	}
	ds.enabled = false
}

// Fail marks the debug session as failed
func (ds *DebugSession) Fail() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.EndTime = time.Now()
	ds.Status = "failed"
	ds.enabled = false
}

// IsEnabled checks if debug is enabled for this session
func (ds *DebugSession) IsEnabled() bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.enabled
}

// AddEvent adds a debug event to the session
func (ds *DebugSession) AddEvent(event *DebugEvent) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.Events = append(ds.Events, event)
	ds.EventCount++
}

// GetEvents returns all events for this session
func (ds *DebugSession) GetEvents() []*DebugEvent {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	// Return a copy to prevent concurrent modification
	events := make([]*DebugEvent, len(ds.Events))
	copy(events, ds.Events)
	return events
}

// GetExecutionTrace returns the execution trace for this session
func (ds *DebugSession) GetExecutionTrace() *ExecutionTrace {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	events := make([]*DebugEvent, len(ds.Events))
	copy(events, ds.Events)

	return &ExecutionTrace{
		SessionID:    ds.SessionID,
		TranCodeName: ds.TranCodeName,
		StartTime:    ds.StartTime,
		EndTime:      ds.EndTime,
		Status:       ds.Status,
		Events:       events,
	}
}

// GetSummary returns a summary of the debug session
func (ds *DebugSession) GetSummary() map[string]interface{} {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	summary := map[string]interface{}{
		"sessionID":    ds.SessionID,
		"tranCodeName": ds.TranCodeName,
		"userID":       ds.UserID,
		"description":  ds.Description,
		"status":       ds.Status,
		"startTime":    ds.StartTime,
		"eventCount":   ds.EventCount,
	}

	if !ds.EndTime.IsZero() {
		summary["endTime"] = ds.EndTime
		summary["duration"] = ds.EndTime.Sub(ds.StartTime).String()
	}

	return summary
}

// DebugSessionManager manages multiple debug sessions
type DebugSessionManager struct {
	sessions   map[string]*DebugSession
	messageBus *MessageBus
	mu         sync.RWMutex
}

// NewDebugSessionManager creates a new debug session manager
func NewDebugSessionManager(messageBus *MessageBus) *DebugSessionManager {
	return &DebugSessionManager{
		sessions:   make(map[string]*DebugSession),
		messageBus: messageBus,
	}
}

// CreateSession creates a new debug session
func (dsm *DebugSessionManager) CreateSession(sessionID, tranCodeName, userID string) *DebugSession {
	dsm.mu.Lock()
	defer dsm.mu.Unlock()

	session := NewDebugSession(sessionID, tranCodeName, userID)
	dsm.sessions[sessionID] = session

	return session
}

// GetSession retrieves a debug session by ID
func (dsm *DebugSessionManager) GetSession(sessionID string) (*DebugSession, bool) {
	dsm.mu.RLock()
	defer dsm.mu.RUnlock()

	session, exists := dsm.sessions[sessionID]
	return session, exists
}

// RemoveSession removes a debug session
func (dsm *DebugSessionManager) RemoveSession(sessionID string) {
	dsm.mu.Lock()
	defer dsm.mu.Unlock()

	delete(dsm.sessions, sessionID)
}

// EmitEvent emits a debug event to the message bus and stores it in the session
func (dsm *DebugSessionManager) EmitEvent(event *DebugEvent) {
	// Add to session if it exists
	if session, exists := dsm.GetSession(event.SessionID); exists && session.IsEnabled() {
		session.AddEvent(event)
	}

	// Publish to message bus for real-time streaming
	dsm.messageBus.Publish(event)
}

// IsDebugEnabled checks if debug is enabled for a session
func (dsm *DebugSessionManager) IsDebugEnabled(sessionID string) bool {
	if session, exists := dsm.GetSession(sessionID); exists {
		return session.IsEnabled()
	}
	return false
}

// Global debug session manager instance
var globalDebugSessionManager *DebugSessionManager
var globalDebugSessionManagerOnce sync.Once

// GetGlobalDebugSessionManager returns the global debug session manager
func GetGlobalDebugSessionManager() *DebugSessionManager {
	globalDebugSessionManagerOnce.Do(func() {
		globalDebugSessionManager = NewDebugSessionManager(GetGlobalMessageBus())
	})
	return globalDebugSessionManager
}

// DebugHelper provides convenience functions for emitting debug events
type DebugHelper struct {
	sessionID       string
	tranCodeName    string
	tranCodeVersion string
	manager         *DebugSessionManager
	stepCounter     int
	mu              sync.Mutex
}

// NewDebugHelper creates a new debug helper
func NewDebugHelper(sessionID, tranCodeName, tranCodeVersion string) *DebugHelper {
	return &DebugHelper{
		sessionID:       sessionID,
		tranCodeName:    tranCodeName,
		tranCodeVersion: tranCodeVersion,
		manager:         GetGlobalDebugSessionManager(),
		stepCounter:     0,
	}
}

// IsEnabled checks if debug is enabled
func (dh *DebugHelper) IsEnabled() bool {
	return dh.manager.IsDebugEnabled(dh.sessionID)
}

// nextStep increments and returns the next step number
func (dh *DebugHelper) nextStep() int {
	dh.mu.Lock()
	defer dh.mu.Unlock()
	dh.stepCounter++
	return dh.stepCounter
}

// EmitTranCodeStart emits a trancode start event
func (dh *DebugHelper) EmitTranCodeStart() {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventTranCodeStart).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithStep(dh.nextStep()).
		WithMessage("TranCode execution started")

	dh.manager.EmitEvent(event)
}

// EmitTranCodeComplete emits a trancode complete event
func (dh *DebugHelper) EmitTranCodeComplete(duration time.Duration, outputs map[string]interface{}) {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventTranCodeComplete).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithStep(dh.nextStep()).
		WithOutputs(outputs).
		WithMessage("TranCode execution completed").
		WithMetadata("duration_ms", duration.Milliseconds())

	dh.manager.EmitEvent(event)
}

// EmitFuncGroupStart emits a funcgroup start event
func (dh *DebugHelper) EmitFuncGroupStart(funcGroupName string) {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventFuncGroupStart).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithFuncGroup(funcGroupName).
		WithStep(dh.nextStep()).
		WithMessage(fmt.Sprintf("FuncGroup '%s' execution started", funcGroupName))

	dh.manager.EmitEvent(event)
}

// EmitFuncGroupComplete emits a funcgroup complete event
func (dh *DebugHelper) EmitFuncGroupComplete(funcGroupName string, duration time.Duration) {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventFuncGroupComplete).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithFuncGroup(funcGroupName).
		WithStep(dh.nextStep()).
		WithMessage(fmt.Sprintf("FuncGroup '%s' execution completed", funcGroupName)).
		WithMetadata("duration_ms", duration.Milliseconds())

	dh.manager.EmitEvent(event)
}

// EmitFuncGroupRouting emits a funcgroup routing event
func (dh *DebugHelper) EmitFuncGroupRouting(funcGroupName string, routingValue interface{}, routingPath string) {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventFuncGroupRouting).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithFuncGroup(funcGroupName).
		WithStep(dh.nextStep()).
		WithRouting(routingValue, routingPath).
		WithMessage(fmt.Sprintf("FuncGroup routing to: %s", routingPath))

	dh.manager.EmitEvent(event)
}

// EmitFunctionStart emits a function start event
func (dh *DebugHelper) EmitFunctionStart(funcGroupName, functionName, functionType string, inputs map[string]interface{}) {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventFunctionStart).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithFuncGroup(funcGroupName).
		WithFunction(functionName, functionType).
		WithStep(dh.nextStep()).
		WithInputs(inputs).
		WithMessage(fmt.Sprintf("Function '%s' (%s) execution started", functionName, functionType))

	dh.manager.EmitEvent(event)
}

// EmitFunctionComplete emits a function complete event
func (dh *DebugHelper) EmitFunctionComplete(funcGroupName, functionName, functionType string,
	outputs map[string]interface{}, startTime time.Time, endTime time.Time) {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventFunctionComplete).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithFuncGroup(funcGroupName).
		WithFunction(functionName, functionType).
		WithStep(dh.nextStep()).
		WithOutputs(outputs).
		WithTiming(startTime, endTime).
		WithMessage(fmt.Sprintf("Function '%s' (%s) execution completed in %v",
			functionName, functionType, endTime.Sub(startTime)))

	dh.manager.EmitEvent(event)
}

// EmitInputMapping emits an input mapping event
func (dh *DebugHelper) EmitInputMapping(funcGroupName, functionName string, inputs map[string]interface{}) {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventInputMapping).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithFuncGroup(funcGroupName).
		WithFunction(functionName, "").
		WithStep(dh.nextStep()).
		WithInputs(inputs).
		WithMessage(fmt.Sprintf("Mapped %d inputs for function '%s'", len(inputs), functionName))

	dh.manager.EmitEvent(event)
}

// EmitOutputMapping emits an output mapping event
func (dh *DebugHelper) EmitOutputMapping(funcGroupName, functionName string, outputs map[string]interface{}) {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventOutputMapping).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithFuncGroup(funcGroupName).
		WithFunction(functionName, "").
		WithStep(dh.nextStep()).
		WithOutputs(outputs).
		WithMessage(fmt.Sprintf("Mapped %d outputs from function '%s'", len(outputs), functionName))

	dh.manager.EmitEvent(event)
}

// EmitDatabaseQuery emits a database query event
func (dh *DebugHelper) EmitDatabaseQuery(funcGroupName, functionName, query string, duration time.Duration) {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventDatabaseQuery).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithFuncGroup(funcGroupName).
		WithFunction(functionName, "").
		WithStep(dh.nextStep()).
		WithMessage(fmt.Sprintf("Database query executed in %v", duration)).
		WithMetadata("query", query).
		WithMetadata("duration_ms", duration.Milliseconds())

	dh.manager.EmitEvent(event)
}

// EmitScriptExecution emits a script execution event
func (dh *DebugHelper) EmitScriptExecution(funcGroupName, functionName, scriptType, script string, duration time.Duration) {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventScriptExecution).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithFuncGroup(funcGroupName).
		WithFunction(functionName, scriptType).
		WithStep(dh.nextStep()).
		WithMessage(fmt.Sprintf("%s script executed in %v", scriptType, duration)).
		WithMetadata("script_type", scriptType).
		WithMetadata("script", script).
		WithMetadata("duration_ms", duration.Milliseconds())

	dh.manager.EmitEvent(event)
}

// EmitError emits an error event
func (dh *DebugHelper) EmitError(funcGroupName, functionName string, err error) {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventFunctionComplete).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithFuncGroup(funcGroupName).
		WithFunction(functionName, "").
		WithStep(dh.nextStep()).
		WithError(err).
		WithLevel("ERROR")

	dh.manager.EmitEvent(event)
}

// EmitTransactionBegin emits a transaction begin event
func (dh *DebugHelper) EmitTransactionBegin() {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventTransactionBegin).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithStep(dh.nextStep()).
		WithMessage("Database transaction started")

	dh.manager.EmitEvent(event)
}

// EmitTransactionCommit emits a transaction commit event
func (dh *DebugHelper) EmitTransactionCommit() {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventTransactionCommit).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithStep(dh.nextStep()).
		WithMessage("Database transaction committed")

	dh.manager.EmitEvent(event)
}

// EmitTransactionRollback emits a transaction rollback event
func (dh *DebugHelper) EmitTransactionRollback(reason string) {
	if !dh.IsEnabled() {
		return
	}

	event := NewDebugEvent(dh.sessionID, EventTransactionRollback).
		WithTranCode(dh.tranCodeName, dh.tranCodeVersion).
		WithStep(dh.nextStep()).
		WithMessage(fmt.Sprintf("Database transaction rolled back: %s", reason)).
		WithMetadata("rollback_reason", reason)

	dh.manager.EmitEvent(event)
}
