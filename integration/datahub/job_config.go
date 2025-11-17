package datahub

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

// JobConfig defines configuration for automated job creation
type JobConfig struct {
	Jobs []JobDefinition `json:"jobs"`
}

// JobDefinition defines a single job configuration
type JobDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Type        string                 `json:"type"` // transform, receive, send, route

	// Trigger configuration
	Trigger     JobTrigger             `json:"trigger"`

	// Job parameters
	Protocol    string                 `json:"protocol"`
	Source      string                 `json:"source,omitempty"`
	Destination string                 `json:"destination,omitempty"`
	MappingID   string                 `json:"mapping_id,omitempty"`
	RoutingRule string                 `json:"routing_rule,omitempty"`

	// Job execution settings
	Priority    int                    `json:"priority"`
	MaxRetries  int                    `json:"max_retries"`
	Timeout     int                    `json:"timeout"` // seconds
	AutoRoute   bool                   `json:"auto_route"`

	// Metadata and options
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// JobTrigger defines when a job should be created
type JobTrigger struct {
	Type      string                 `json:"type"` // on_receive, on_schedule, on_event, manual
	Protocol  string                 `json:"protocol,omitempty"` // For on_receive triggers
	Topic     string                 `json:"topic,omitempty"`
	Schedule  string                 `json:"schedule,omitempty"` // Cron expression
	Event     string                 `json:"event,omitempty"`
	Condition *MappingCondition      `json:"condition,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

// JobConfigManager manages job configurations
type JobConfigManager struct {
	config     *JobConfig
	hub        *DataHub
	logger     *logrus.Logger
	enabled    bool
	configFile string
}

// NewJobConfigManager creates a new job config manager
func NewJobConfigManager(hub *DataHub, logger *logrus.Logger) *JobConfigManager {
	if logger == nil {
		logger = logrus.New()
	}

	return &JobConfigManager{
		config:  &JobConfig{Jobs: make([]JobDefinition, 0)},
		hub:     hub,
		logger:  logger,
		enabled: true,
	}
}

// LoadFromFile loads job configurations from a JSON file
func (jcm *JobConfigManager) LoadFromFile(filePath string) error {
	jcm.logger.Infof("Loading job configurations from %s", filePath)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config JobConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	jcm.config = &config
	jcm.configFile = filePath

	jcm.logger.Infof("Loaded %d job configurations", len(config.Jobs))

	return nil
}

// GetJobDefinition gets a job definition by ID
func (jcm *JobConfigManager) GetJobDefinition(jobID string) (*JobDefinition, error) {
	for i := range jcm.config.Jobs {
		if jcm.config.Jobs[i].ID == jobID {
			return &jcm.config.Jobs[i], nil
		}
	}
	return nil, fmt.Errorf("job definition %s not found", jobID)
}

// GetJobsForTrigger gets all jobs that match a trigger type
func (jcm *JobConfigManager) GetJobsForTrigger(triggerType string) []JobDefinition {
	jobs := make([]JobDefinition, 0)

	for _, job := range jcm.config.Jobs {
		if job.Enabled && job.Trigger.Type == triggerType {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

// GetJobsForProtocol gets all jobs for a specific protocol
func (jcm *JobConfigManager) GetJobsForProtocol(protocol string) []JobDefinition {
	jobs := make([]JobDefinition, 0)

	for _, job := range jcm.config.Jobs {
		if job.Enabled && (job.Protocol == protocol || job.Trigger.Protocol == protocol) {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

// CreateJobPayload creates a job payload from a job definition and message data
func (jcm *JobConfigManager) CreateJobPayload(jobDef *JobDefinition, messageData map[string]interface{}) map[string]interface{} {
	payload := map[string]interface{}{
		"job_definition_id": jobDef.ID,
		"type":              jobDef.Type,
		"protocol":          jobDef.Protocol,
		"source":            jobDef.Source,
		"destination":       jobDef.Destination,
		"mapping_id":        jobDef.MappingID,
		"routing_rule_id":   jobDef.RoutingRule,
		"priority":          jobDef.Priority,
		"max_retries":       jobDef.MaxRetries,
		"timeout":           jobDef.Timeout,
		"auto_route":        jobDef.AutoRoute,
	}

	// Merge metadata
	if jobDef.Metadata != nil {
		payload["metadata"] = jobDef.Metadata
	}

	// Add message data
	if messageData != nil {
		for k, v := range messageData {
			if _, exists := payload[k]; !exists {
				payload[k] = v
			}
		}
		payload["body"] = messageData["body"]
		payload["content_type"] = messageData["content_type"]
	}

	return payload
}

// EvaluateTriggerCondition evaluates whether a trigger condition is met
func (jcm *JobConfigManager) EvaluateTriggerCondition(condition *MappingCondition, data map[string]interface{}) bool {
	if condition == nil {
		return true
	}

	// Get field value
	fieldValue, exists := data[condition.Field]
	if !exists {
		if condition.Operator == "exists" {
			return condition.Value == false || condition.Value == "false"
		}
		return false
	}

	// Evaluate based on operator
	switch condition.Operator {
	case "eq":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", condition.Value)
	case "ne":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", condition.Value)
	case "contains":
		fieldStr := fmt.Sprintf("%v", fieldValue)
		valueStr := fmt.Sprintf("%v", condition.Value)
		return contains(fieldStr, valueStr)
	case "exists":
		return condition.Value == true || condition.Value == "true"
	default:
		return false
	}
}

// FindMatchingJobs finds all jobs that match the given message
func (jcm *JobConfigManager) FindMatchingJobs(protocol string, topic string, messageData map[string]interface{}) []JobDefinition {
	matchedJobs := make([]JobDefinition, 0)

	for _, job := range jcm.config.Jobs {
		if !job.Enabled {
			continue
		}

		// Check trigger type
		if job.Trigger.Type != "on_receive" {
			continue
		}

		// Check protocol
		if job.Trigger.Protocol != "" && job.Trigger.Protocol != protocol {
			continue
		}

		// Check topic
		if job.Trigger.Topic != "" && job.Trigger.Topic != topic {
			continue
		}

		// Check trigger condition
		if job.Trigger.Condition != nil {
			if !jcm.EvaluateTriggerCondition(job.Trigger.Condition, messageData) {
				continue
			}
		}

		matchedJobs = append(matchedJobs, job)
	}

	jcm.logger.Infof("Found %d matching jobs for protocol=%s, topic=%s", len(matchedJobs), protocol, topic)

	return matchedJobs
}

// Enable enables the job config manager
func (jcm *JobConfigManager) Enable() {
	jcm.enabled = true
	jcm.logger.Info("Job config manager enabled")
}

// Disable disables the job config manager
func (jcm *JobConfigManager) Disable() {
	jcm.enabled = false
	jcm.logger.Info("Job config manager disabled")
}

// IsEnabled returns whether the job config manager is enabled
func (jcm *JobConfigManager) IsEnabled() bool {
	return jcm.enabled
}

// GetStats returns statistics about configured jobs
func (jcm *JobConfigManager) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":    jcm.enabled,
		"total_jobs": len(jcm.config.Jobs),
	}

	// Count by type
	typeCount := make(map[string]int)
	triggerCount := make(map[string]int)
	protocolCount := make(map[string]int)
	enabledCount := 0

	for _, job := range jcm.config.Jobs {
		typeCount[job.Type]++
		triggerCount[job.Trigger.Type]++
		protocolCount[job.Protocol]++

		if job.Enabled {
			enabledCount++
		}
	}

	stats["enabled_jobs"] = enabledCount
	stats["by_type"] = typeCount
	stats["by_trigger"] = triggerCount
	stats["by_protocol"] = protocolCount

	return stats
}

// Reload reloads the configuration from file
func (jcm *JobConfigManager) Reload() error {
	if jcm.configFile == "" {
		return fmt.Errorf("no config file loaded")
	}

	return jcm.LoadFromFile(jcm.configFile)
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}
