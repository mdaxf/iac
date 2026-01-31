package models

import (
	"time"
)

// ConfigurationStatus represents the status of a configuration version
type ConfigurationStatus string

const (
	ConfigStatusDraft    ConfigurationStatus = "draft"
	ConfigStatusActive   ConfigurationStatus = "active"
	ConfigStatusArchived ConfigurationStatus = "archived"
)

// ConfigurationStore represents a versioned configuration entry
type ConfigurationStore struct {
	// Unique identifier
	ID string `json:"id" bson:"_id,omitempty"`

	// Version identifier (semantic versioning recommended)
	Version string `json:"version" bson:"version"`

	// Environment this configuration applies to
	Environment string `json:"environment" bson:"environment"`

	// Status of this configuration version
	Status ConfigurationStatus `json:"status" bson:"status"`

	// Configuration name/type
	Name string `json:"name" bson:"name"`

	// Description of this version
	Description string `json:"description,omitempty" bson:"description,omitempty"`

	// The actual configuration data
	Data map[string]interface{} `json:"data" bson:"data"`

	// Schema version for migration support
	SchemaVersion int `json:"schema_version" bson:"schema_version"`

	// Instance pattern for targeted configurations (e.g., "app-*", "hub-1")
	InstancePattern string `json:"instance_pattern,omitempty" bson:"instance_pattern,omitempty"`

	// Tags for categorization
	Tags []string `json:"tags,omitempty" bson:"tags,omitempty"`

	// Audit information
	CreatedBy   string    `json:"created_by" bson:"created_by"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedBy   string    `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	ActivatedBy string    `json:"activated_by,omitempty" bson:"activated_by,omitempty"`
	ActivatedAt *time.Time `json:"activated_at,omitempty" bson:"activated_at,omitempty"`

	// Previous version reference for rollback
	PreviousVersion string `json:"previous_version,omitempty" bson:"previous_version,omitempty"`

	// Checksum for integrity verification
	Checksum string `json:"checksum,omitempty" bson:"checksum,omitempty"`
}

// ConfigurationHistory represents a change history entry
type ConfigurationHistory struct {
	ID            string                 `json:"id" bson:"_id,omitempty"`
	ConfigID      string                 `json:"config_id" bson:"config_id"`
	ConfigName    string                 `json:"config_name" bson:"config_name"`
	Version       string                 `json:"version" bson:"version"`
	Environment   string                 `json:"environment" bson:"environment"`
	Action        string                 `json:"action" bson:"action"` // created, updated, activated, archived, rolled_back
	Changes       []ConfigChange         `json:"changes,omitempty" bson:"changes,omitempty"`
	PreviousData  map[string]interface{} `json:"previous_data,omitempty" bson:"previous_data,omitempty"`
	NewData       map[string]interface{} `json:"new_data,omitempty" bson:"new_data,omitempty"`
	PerformedBy   string                 `json:"performed_by" bson:"performed_by"`
	PerformedAt   time.Time              `json:"performed_at" bson:"performed_at"`
	Reason        string                 `json:"reason,omitempty" bson:"reason,omitempty"`
	InstanceID    string                 `json:"instance_id,omitempty" bson:"instance_id,omitempty"`
}

// ConfigChange represents a single configuration change
type ConfigChange struct {
	Path      string      `json:"path" bson:"path"`
	OldValue  interface{} `json:"old_value,omitempty" bson:"old_value,omitempty"`
	NewValue  interface{} `json:"new_value,omitempty" bson:"new_value,omitempty"`
	Operation string      `json:"operation" bson:"operation"` // add, update, delete
}

// ConfigurationOverride represents instance-specific configuration overrides
type ConfigurationOverride struct {
	ID              string                 `json:"id" bson:"_id,omitempty"`
	ConfigID        string                 `json:"config_id" bson:"config_id"`
	InstancePattern string                 `json:"instance_pattern" bson:"instance_pattern"`
	Environment     string                 `json:"environment" bson:"environment"`
	Overrides       map[string]interface{} `json:"overrides" bson:"overrides"`
	Description     string                 `json:"description,omitempty" bson:"description,omitempty"`
	Enabled         bool                   `json:"enabled" bson:"enabled"`
	CreatedBy       string                 `json:"created_by" bson:"created_by"`
	CreatedAt       time.Time              `json:"created_at" bson:"created_at"`
	UpdatedBy       string                 `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	UpdatedAt       time.Time              `json:"updated_at" bson:"updated_at"`
}

// ConfigReloadEvent represents a configuration reload event for SignalR notification
type ConfigReloadEvent struct {
	ConfigID      string    `json:"config_id"`
	ConfigName    string    `json:"config_name"`
	Version       string    `json:"version"`
	Environment   string    `json:"environment"`
	Action        string    `json:"action"` // activated, updated, rolled_back
	SourceInstance string   `json:"source_instance"`
	Timestamp     time.Time `json:"timestamp"`
}

// NewConfigurationStore creates a new configuration store entry
func NewConfigurationStore(name, version, environment, createdBy string, data map[string]interface{}) *ConfigurationStore {
	now := time.Now()
	return &ConfigurationStore{
		Name:          name,
		Version:       version,
		Environment:   environment,
		Status:        ConfigStatusDraft,
		Data:          data,
		SchemaVersion: 1,
		CreatedBy:     createdBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Activate marks the configuration as active
func (c *ConfigurationStore) Activate(activatedBy string) {
	now := time.Now()
	c.Status = ConfigStatusActive
	c.ActivatedBy = activatedBy
	c.ActivatedAt = &now
	c.UpdatedAt = now
}

// Archive marks the configuration as archived
func (c *ConfigurationStore) Archive(updatedBy string) {
	c.Status = ConfigStatusArchived
	c.UpdatedBy = updatedBy
	c.UpdatedAt = time.Now()
}

// IsActive returns true if the configuration is active
func (c *ConfigurationStore) IsActive() bool {
	return c.Status == ConfigStatusActive
}

// GetValue retrieves a value from the configuration data by key path
func (c *ConfigurationStore) GetValue(key string) (interface{}, bool) {
	if c.Data == nil {
		return nil, false
	}
	value, ok := c.Data[key]
	return value, ok
}

// SetValue sets a value in the configuration data
func (c *ConfigurationStore) SetValue(key string, value interface{}) {
	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	c.Data[key] = value
}

// NewConfigurationHistory creates a new history entry
func NewConfigurationHistory(configID, configName, version, environment, action, performedBy string) *ConfigurationHistory {
	return &ConfigurationHistory{
		ConfigID:    configID,
		ConfigName:  configName,
		Version:     version,
		Environment: environment,
		Action:      action,
		PerformedBy: performedBy,
		PerformedAt: time.Now(),
	}
}

// AddChange adds a change record to the history
func (h *ConfigurationHistory) AddChange(path, operation string, oldValue, newValue interface{}) {
	h.Changes = append(h.Changes, ConfigChange{
		Path:      path,
		OldValue:  oldValue,
		NewValue:  newValue,
		Operation: operation,
	})
}

// ConfigurationListItem represents a summary of a configuration for listing
type ConfigurationListItem struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Version     string              `json:"version"`
	Environment string              `json:"environment"`
	Status      ConfigurationStatus `json:"status"`
	Description string              `json:"description,omitempty"`
	UpdatedAt   time.Time           `json:"updated_at"`
	UpdatedBy   string              `json:"updated_by,omitempty"`
}

// ToListItem converts a ConfigurationStore to a list item
func (c *ConfigurationStore) ToListItem() ConfigurationListItem {
	return ConfigurationListItem{
		ID:          c.ID,
		Name:        c.Name,
		Version:     c.Version,
		Environment: c.Environment,
		Status:      c.Status,
		Description: c.Description,
		UpdatedAt:   c.UpdatedAt,
		UpdatedBy:   c.UpdatedBy,
	}
}
