package logger

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelCritical
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Fields represents structured log fields
type Fields map[string]interface{}

// StructuredLogger provides optimized structured logging
type StructuredLogger struct {
	moduleName     string
	user           string
	controllerName string
	minLevel       LogLevel
	enableDebug    bool
	fields         Fields
	mu             sync.RWMutex
	correlationID  string
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(moduleName, user, controllerName string) *StructuredLogger {
	return &StructuredLogger{
		moduleName:     moduleName,
		user:           user,
		controllerName: controllerName,
		minLevel:       LevelInfo,
		enableDebug:    false,
		fields:         make(Fields),
		correlationID:  generateCorrelationID(),
	}
}

// generateCorrelationID generates a unique correlation ID for request tracing
func generateCorrelationID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// WithFields creates a new logger with additional fields
func (sl *StructuredLogger) WithFields(fields Fields) *StructuredLogger {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	newLogger := &StructuredLogger{
		moduleName:     sl.moduleName,
		user:           sl.user,
		controllerName: sl.controllerName,
		minLevel:       sl.minLevel,
		enableDebug:    sl.enableDebug,
		fields:         make(Fields),
		correlationID:  sl.correlationID,
	}

	// Copy existing fields
	for k, v := range sl.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithField creates a new logger with an additional field
func (sl *StructuredLogger) WithField(key string, value interface{}) *StructuredLogger {
	return sl.WithFields(Fields{key: value})
}

// WithCorrelationID sets the correlation ID for request tracing
func (sl *StructuredLogger) WithCorrelationID(correlationID string) *StructuredLogger {
	newLogger := sl.clone()
	newLogger.correlationID = correlationID
	return newLogger
}

// clone creates a copy of the logger
func (sl *StructuredLogger) clone() *StructuredLogger {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	newLogger := &StructuredLogger{
		moduleName:     sl.moduleName,
		user:           sl.user,
		controllerName: sl.controllerName,
		minLevel:       sl.minLevel,
		enableDebug:    sl.enableDebug,
		fields:         make(Fields),
		correlationID:  sl.correlationID,
	}

	for k, v := range sl.fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// SetMinLevel sets the minimum log level
func (sl *StructuredLogger) SetMinLevel(level LogLevel) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.minLevel = level
}

// EnableDebug enables debug logging
func (sl *StructuredLogger) EnableDebug(enable bool) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.enableDebug = enable
}

// IsDebugEnabled checks if debug logging is enabled
func (sl *StructuredLogger) IsDebugEnabled() bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.enableDebug && sl.minLevel <= LevelDebug
}

// shouldLog checks if a message at the given level should be logged
func (sl *StructuredLogger) shouldLog(level LogLevel) bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	if level == LevelDebug && !sl.enableDebug {
		return false
	}

	return level >= sl.minLevel
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp     time.Time              `json:"timestamp"`
	Level         string                 `json:"level"`
	Message       string                 `json:"message"`
	Module        string                 `json:"module"`
	User          string                 `json:"user"`
	Controller    string                 `json:"controller"`
	CorrelationID string                 `json:"correlation_id"`
	Fields        map[string]interface{} `json:"fields,omitempty"`
}

// buildEntry builds a log entry with all metadata
func (sl *StructuredLogger) buildEntry(level LogLevel, message string) LogEntry {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	fields := make(map[string]interface{})
	for k, v := range sl.fields {
		fields[k] = v
	}

	return LogEntry{
		Timestamp:     time.Now(),
		Level:         level.String(),
		Message:       message,
		Module:        sl.moduleName,
		User:          sl.user,
		Controller:    sl.controllerName,
		CorrelationID: sl.correlationID,
		Fields:        fields,
	}
}

// Debug logs a debug message (lazy evaluation)
func (sl *StructuredLogger) Debug(messageFunc func() string) {
	if !sl.shouldLog(LevelDebug) {
		return // Skip message generation if debug is disabled
	}

	message := messageFunc()
	entry := sl.buildEntry(LevelDebug, message)
	logEntry(entry)
}

// DebugString logs a debug message (direct string - use when message is cheap to create)
func (sl *StructuredLogger) DebugString(message string) {
	if !sl.shouldLog(LevelDebug) {
		return
	}

	entry := sl.buildEntry(LevelDebug, message)
	logEntry(entry)
}

// Info logs an info message
func (sl *StructuredLogger) Info(message string) {
	if !sl.shouldLog(LevelInfo) {
		return
	}

	entry := sl.buildEntry(LevelInfo, message)
	logEntry(entry)
}

// Warning logs a warning message
func (sl *StructuredLogger) Warning(message string) {
	if !sl.shouldLog(LevelWarning) {
		return
	}

	entry := sl.buildEntry(LevelWarning, message)
	logEntry(entry)
}

// Error logs an error message
func (sl *StructuredLogger) Error(message string) {
	if !sl.shouldLog(LevelError) {
		return
	}

	entry := sl.buildEntry(LevelError, message)
	logEntry(entry)
}

// Critical logs a critical error message
func (sl *StructuredLogger) Critical(message string) {
	if !sl.shouldLog(LevelCritical) {
		return
	}

	entry := sl.buildEntry(LevelCritical, message)
	logEntry(entry)
}

// Performance logs a performance metric with duration
func (sl *StructuredLogger) Performance(operation string, duration time.Duration) {
	sl.WithFields(Fields{
		"operation":     operation,
		"duration_ms":   duration.Milliseconds(),
		"duration_μs":   duration.Microseconds(),
		"duration_text": duration.String(),
	}).Info(fmt.Sprintf("Performance: %s completed in %v", operation, duration))
}

// PerformanceWithFields logs a performance metric with additional fields
func (sl *StructuredLogger) PerformanceWithFields(operation string, duration time.Duration, fields Fields) {
	mergedFields := Fields{
		"operation":     operation,
		"duration_ms":   duration.Milliseconds(),
		"duration_μs":   duration.Microseconds(),
		"duration_text": duration.String(),
	}

	for k, v := range fields {
		mergedFields[k] = v
	}

	sl.WithFields(mergedFields).Info(fmt.Sprintf("Performance: %s completed in %v", operation, duration))
}

// logEntry outputs the log entry (can be replaced with external logger)
func logEntry(entry LogEntry) {
	// For now, use the existing Log infrastructure
	// This can be replaced with structured logging to file, ELK, etc.

	// Format: [TIMESTAMP] [LEVEL] [MODULE] [CORRELATION_ID] message {fields}
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")

	fieldsStr := ""
	if len(entry.Fields) > 0 {
		fieldsStr = fmt.Sprintf(" %v", entry.Fields)
	}

	logMessage := fmt.Sprintf("[%s] [%s] [%s] [%s] %s%s",
		timestamp,
		entry.Level,
		entry.Module,
		entry.CorrelationID,
		entry.Message,
		fieldsStr,
	)

	// Use existing logger infrastructure
	log := Log{
		ModuleName:     entry.Module,
		User:           entry.User,
		ControllerName: entry.Controller,
	}

	switch entry.Level {
	case "DEBUG":
		log.Debug(logMessage)
	case "INFO":
		log.Info(logMessage)
	case "WARNING":
		log.Warning(logMessage)
	case "ERROR":
		log.Error(logMessage)
	case "CRITICAL":
		log.Error(logMessage) // Use Error for critical as well
	}
}

// LogContext holds logging context for request tracing
type LogContext struct {
	CorrelationID string
	RequestID     string
	UserID        string
	TransactionID string
}

// NewLogContext creates a new log context
func NewLogContext() *LogContext {
	return &LogContext{
		CorrelationID: generateCorrelationID(),
	}
}

// WithContext creates a logger with context information
func (sl *StructuredLogger) WithContext(ctx context.Context) *StructuredLogger {
	// Extract context values if they exist
	if ctx == nil {
		return sl
	}

	fields := make(Fields)

	if correlationID, ok := ctx.Value("correlation_id").(string); ok {
		fields["correlation_id"] = correlationID
	}

	if requestID, ok := ctx.Value("request_id").(string); ok {
		fields["request_id"] = requestID
	}

	if userID, ok := ctx.Value("user_id").(string); ok {
		fields["user_id"] = userID
	}

	if transactionID, ok := ctx.Value("transaction_id").(string); ok {
		fields["transaction_id"] = transactionID
	}

	if len(fields) > 0 {
		return sl.WithFields(fields)
	}

	return sl
}

// LoggerPool manages a pool of reusable loggers
type LoggerPool struct {
	pool sync.Pool
}

// NewLoggerPool creates a new logger pool
func NewLoggerPool() *LoggerPool {
	return &LoggerPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &StructuredLogger{
					fields: make(Fields),
				}
			},
		},
	}
}

// Get retrieves a logger from the pool
func (lp *LoggerPool) Get(moduleName, user, controllerName string) *StructuredLogger {
	logger := lp.pool.Get().(*StructuredLogger)
	logger.moduleName = moduleName
	logger.user = user
	logger.controllerName = controllerName
	logger.correlationID = generateCorrelationID()
	return logger
}

// Put returns a logger to the pool
func (lp *LoggerPool) Put(logger *StructuredLogger) {
	logger.fields = make(Fields)
	lp.pool.Put(logger)
}

// Global logger pool for reuse
var globalLoggerPool = NewLoggerPool()

// GetPooledLogger gets a logger from the global pool
func GetPooledLogger(moduleName, user, controllerName string) *StructuredLogger {
	return globalLoggerPool.Get(moduleName, user, controllerName)
}

// ReleaseLogger returns a logger to the global pool
func ReleaseLogger(logger *StructuredLogger) {
	globalLoggerPool.Put(logger)
}
