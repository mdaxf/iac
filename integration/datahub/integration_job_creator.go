package datahub

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// IntegrationJobCreator creates jobs automatically when messages are received
// This bridges the gap between protocol adapters and the job system
type IntegrationJobCreator struct {
	jobConfigManager *JobConfigManager
	db               *sql.DB
	logger           *logrus.Logger
	enabled          bool
}

// NewIntegrationJobCreator creates a new integration job creator
func NewIntegrationJobCreator(jobConfigManager *JobConfigManager, db *sql.DB, logger *logrus.Logger) *IntegrationJobCreator {
	if logger == nil {
		logger = logrus.New()
	}

	return &IntegrationJobCreator{
		jobConfigManager: jobConfigManager,
		db:               db,
		logger:           logger,
		enabled:          true,
	}
}

// OnMessageReceived is called when a message is received from any protocol
// It automatically creates jobs based on configuration
func (ijc *IntegrationJobCreator) OnMessageReceived(
	ctx context.Context,
	protocol string,
	topic string,
	payload interface{},
	metadata map[string]interface{},
) ([]string, error) {

	if !ijc.enabled {
		return nil, fmt.Errorf("integration job creator is disabled")
	}

	ijc.logger.Infof("Message received: protocol=%s, topic=%s", protocol, topic)

	// Prepare message data
	messageData := map[string]interface{}{
		"protocol":     protocol,
		"topic":        topic,
		"body":         payload,
		"received_at":  time.Now(),
		"content_type": "application/json",
	}

	// Add metadata
	if metadata != nil {
		for k, v := range metadata {
			messageData[k] = v
		}
	}

	// Find matching job configurations
	matchingJobs := ijc.jobConfigManager.FindMatchingJobs(protocol, topic, messageData)

	if len(matchingJobs) == 0 {
		ijc.logger.Debugf("No matching jobs found for protocol=%s, topic=%s", protocol, topic)
		return nil, nil
	}

	createdJobIDs := make([]string, 0)

	// Create jobs for each matching configuration
	for _, jobDef := range matchingJobs {
		jobID, err := ijc.createJob(ctx, &jobDef, messageData)
		if err != nil {
			ijc.logger.Errorf("Failed to create job for %s: %v", jobDef.Name, err)
			continue
		}

		createdJobIDs = append(createdJobIDs, jobID)
		ijc.logger.Infof("Created job %s for %s", jobID, jobDef.Name)
	}

	return createdJobIDs, nil
}

// createJob creates a single job in the queue
func (ijc *IntegrationJobCreator) createJob(ctx context.Context, jobDef *JobDefinition, messageData map[string]interface{}) (string, error) {
	// Create job payload
	payload := ijc.jobConfigManager.CreateJobPayload(jobDef, messageData)

	// Convert payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Determine handler based on job type
	handler := ijc.getHandlerForJobType(jobDef.Type)

	// Create job metadata
	metadata := map[string]interface{}{
		"source":             "datahub",
		"protocol":           protocol,
		"topic":              jobDef.Trigger.Topic,
		"job_definition_id":  jobDef.ID,
		"job_definition_name": jobDef.Name,
	}

	// Merge job definition metadata
	if jobDef.Metadata != nil {
		for k, v := range jobDef.Metadata {
			metadata[k] = v
		}
	}

	// Create job record
	// Note: This is a simplified version - in production you'd use the proper job service
	jobID := fmt.Sprintf("dh-%s-%d", jobDef.ID, time.Now().UnixNano())

	// Prepare SQL insert (simplified - adjust to your actual schema)
	query := `
		INSERT INTO queue_jobs (
			id, type_id, method, protocol, direction, handler,
			payload, metadata, priority, max_retries, status_id, created_by, created_on
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	metadataJSON, _ := json.Marshal(metadata)

	_, err = ijc.db.ExecContext(ctx, query,
		jobID,
		2,                    // JobTypeIntegration
		jobDef.Type,
		jobDef.Protocol,
		"inbound",            // JobDirectionInbound
		handler,
		string(payloadJSON),
		string(metadataJSON),
		jobDef.Priority,
		jobDef.MaxRetries,
		1,                    // JobStatusPending
		"datahub",
		time.Now(),
	)

	if err != nil {
		return "", fmt.Errorf("failed to create job record: %w", err)
	}

	return jobID, nil
}

// getHandlerForJobType returns the appropriate handler name for a job type
func (ijc *IntegrationJobCreator) getHandlerForJobType(jobType string) string {
	switch jobType {
	case "transform":
		return "DataHub_Transform"
	case "receive":
		return "DataHub_Receive"
	case "send":
		return "DataHub_Send"
	case "route":
		return "DataHub_Route"
	default:
		return "DataHub_Transform"
	}
}

// OnMessageSent is called before a message is sent
// It can create jobs for outbound processing
func (ijc *IntegrationJobCreator) OnMessageSent(
	ctx context.Context,
	protocol string,
	destination string,
	payload interface{},
	metadata map[string]interface{},
) (string, error) {

	if !ijc.enabled {
		return "", fmt.Errorf("integration job creator is disabled")
	}

	ijc.logger.Infof("Message sending: protocol=%s, destination=%s", protocol, destination)

	// Prepare message data
	messageData := map[string]interface{}{
		"protocol":    protocol,
		"destination": destination,
		"body":        payload,
		"sent_at":     time.Now(),
	}

	// Add metadata
	if metadata != nil {
		for k, v := range metadata {
			messageData[k] = v
		}
	}

	// Find matching outbound jobs
	jobs := ijc.jobConfigManager.GetJobsForTrigger("on_send")

	for _, jobDef := range jobs {
		if jobDef.Protocol == protocol || jobDef.Protocol == "" {
			jobID, err := ijc.createJob(ctx, &jobDef, messageData)
			if err != nil {
				ijc.logger.Errorf("Failed to create outbound job: %v", err)
				continue
			}
			return jobID, nil
		}
	}

	return "", nil
}

// CreateManualJob creates a job manually (not triggered by message)
func (ijc *IntegrationJobCreator) CreateManualJob(
	ctx context.Context,
	jobDefinitionID string,
	customData map[string]interface{},
) (string, error) {

	if !ijc.enabled {
		return "", fmt.Errorf("integration job creator is disabled")
	}

	// Get job definition
	jobDef, err := ijc.jobConfigManager.GetJobDefinition(jobDefinitionID)
	if err != nil {
		return "", fmt.Errorf("failed to get job definition: %w", err)
	}

	if !jobDef.Enabled {
		return "", fmt.Errorf("job definition %s is disabled", jobDefinitionID)
	}

	// Create job
	jobID, err := ijc.createJob(ctx, jobDef, customData)
	if err != nil {
		return "", fmt.Errorf("failed to create manual job: %w", err)
	}

	ijc.logger.Infof("Created manual job %s for %s", jobID, jobDef.Name)
	return jobID, nil
}

// Enable enables the integration job creator
func (ijc *IntegrationJobCreator) Enable() {
	ijc.enabled = true
	ijc.logger.Info("Integration job creator enabled")
}

// Disable disables the integration job creator
func (ijc *IntegrationJobCreator) Disable() {
	ijc.enabled = false
	ijc.logger.Info("Integration job creator disabled")
}

// IsEnabled returns whether the integration job creator is enabled
func (ijc *IntegrationJobCreator) IsEnabled() bool {
	return ijc.enabled
}

// GetStats returns statistics
func (ijc *IntegrationJobCreator) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled": ijc.enabled,
	}

	if ijc.jobConfigManager != nil {
		stats["config_stats"] = ijc.jobConfigManager.GetStats()
	}

	return stats
}

// Helper function to get protocol name
func getProtocol(messageData map[string]interface{}) string {
	if protocol, ok := messageData["protocol"].(string); ok {
		return protocol
	}
	return "UNKNOWN"
}
