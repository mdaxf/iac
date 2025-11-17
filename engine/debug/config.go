package debug

import (
	"sync"
	"time"
)

// DebugConfig holds configuration for the debug system
type DebugConfig struct {
	// Global enable/disable for debug system
	Enabled bool

	// Maximum number of events to store per session
	MaxEventsPerSession int

	// Event buffer size for subscribers
	SubscriberBufferSize int

	// Timeout for inactive subscribers
	SubscriberTimeout time.Duration

	// Cleanup interval for inactive subscribers
	CleanupInterval time.Duration

	// Maximum age for completed sessions before cleanup
	MaxSessionAge time.Duration

	// Session cleanup interval
	SessionCleanupInterval time.Duration

	// Maximum number of concurrent debug sessions
	MaxConcurrentSessions int

	// Sanitize sensitive data in events
	SanitizeSensitiveData bool

	// Sensitive field names to sanitize
	SensitiveFields []string

	// Maximum input/output data size (bytes) to include in events
	// If 0, no limit
	MaxDataSize int

	// Event types to exclude from streaming
	ExcludedEventTypes []EventType

	// Minimum log level for events
	MinLogLevel string
}

// DefaultDebugConfig returns default debug configuration
func DefaultDebugConfig() *DebugConfig {
	return &DebugConfig{
		Enabled:                false, // Disabled by default for performance
		MaxEventsPerSession:    10000,
		SubscriberBufferSize:   100,
		SubscriberTimeout:      5 * time.Minute,
		CleanupInterval:        1 * time.Minute,
		MaxSessionAge:          1 * time.Hour,
		SessionCleanupInterval: 10 * time.Minute,
		MaxConcurrentSessions:  100,
		SanitizeSensitiveData:  true,
		SensitiveFields: []string{
			"password", "passwd", "pwd",
			"token", "api_key", "apikey", "access_token",
			"secret", "private_key", "privatekey",
			"credit_card", "creditcard", "ssn",
		},
		MaxDataSize:        10 * 1024 * 1024, // 10MB
		ExcludedEventTypes: []EventType{},
		MinLogLevel:        "DEBUG",
	}
}

// DebugConfigManager manages debug configuration
type DebugConfigManager struct {
	config *DebugConfig
	mu     sync.RWMutex
}

// NewDebugConfigManager creates a new debug config manager
func NewDebugConfigManager(config *DebugConfig) *DebugConfigManager {
	if config == nil {
		config = DefaultDebugConfig()
	}

	return &DebugConfigManager{
		config: config,
	}
}

// GetConfig returns a copy of the current configuration
func (dcm *DebugConfigManager) GetConfig() *DebugConfig {
	dcm.mu.RLock()
	defer dcm.mu.RUnlock()

	// Return a copy to prevent external modification
	configCopy := *dcm.config
	return &configCopy
}

// SetConfig updates the configuration
func (dcm *DebugConfigManager) SetConfig(config *DebugConfig) {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	dcm.config = config
}

// IsEnabled checks if debug is globally enabled
func (dcm *DebugConfigManager) IsEnabled() bool {
	dcm.mu.RLock()
	defer dcm.mu.RUnlock()

	return dcm.config.Enabled
}

// Enable enables debug globally
func (dcm *DebugConfigManager) Enable() {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	dcm.config.Enabled = true
}

// Disable disables debug globally
func (dcm *DebugConfigManager) Disable() {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	dcm.config.Enabled = false
}

// IsEventTypeExcluded checks if an event type is excluded
func (dcm *DebugConfigManager) IsEventTypeExcluded(eventType EventType) bool {
	dcm.mu.RLock()
	defer dcm.mu.RUnlock()

	for _, excludedType := range dcm.config.ExcludedEventTypes {
		if excludedType == eventType {
			return true
		}
	}

	return false
}

// ShouldSanitize checks if sensitive data should be sanitized
func (dcm *DebugConfigManager) ShouldSanitize() bool {
	dcm.mu.RLock()
	defer dcm.mu.RUnlock()

	return dcm.config.SanitizeSensitiveData
}

// GetSensitiveFields returns the list of sensitive field names
func (dcm *DebugConfigManager) GetSensitiveFields() []string {
	dcm.mu.RLock()
	defer dcm.mu.RUnlock()

	// Return a copy
	fields := make([]string, len(dcm.config.SensitiveFields))
	copy(fields, dcm.config.SensitiveFields)
	return fields
}

// GetMaxDataSize returns the maximum data size for events
func (dcm *DebugConfigManager) GetMaxDataSize() int {
	dcm.mu.RLock()
	defer dcm.mu.RUnlock()

	return dcm.config.MaxDataSize
}

// UpdateSensitiveFields updates the list of sensitive field names
func (dcm *DebugConfigManager) UpdateSensitiveFields(fields []string) {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	dcm.config.SensitiveFields = fields
}

// AddSensitiveField adds a field name to the sensitive fields list
func (dcm *DebugConfigManager) AddSensitiveField(field string) {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	// Check if already exists
	for _, existingField := range dcm.config.SensitiveFields {
		if existingField == field {
			return
		}
	}

	dcm.config.SensitiveFields = append(dcm.config.SensitiveFields, field)
}

// Global debug config manager instance
var globalDebugConfigManager *DebugConfigManager
var globalDebugConfigManagerOnce sync.Once

// GetGlobalDebugConfig returns the global debug config manager
func GetGlobalDebugConfig() *DebugConfigManager {
	globalDebugConfigManagerOnce.Do(func() {
		globalDebugConfigManager = NewDebugConfigManager(DefaultDebugConfig())
	})
	return globalDebugConfigManager
}

// SetGlobalDebugConfig sets the global debug configuration
func SetGlobalDebugConfig(config *DebugConfig) {
	GetGlobalDebugConfig().SetConfig(config)
}

// EnableGlobalDebug enables debug globally
func EnableGlobalDebug() {
	GetGlobalDebugConfig().Enable()
}

// DisableGlobalDebug disables debug globally
func DisableGlobalDebug() {
	GetGlobalDebugConfig().Disable()
}

// IsGlobalDebugEnabled checks if debug is globally enabled
func IsGlobalDebugEnabled() bool {
	return GetGlobalDebugConfig().IsEnabled()
}
