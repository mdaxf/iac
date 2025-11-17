package jobqueue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac-signalr/signalr"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// LegacyMessage represents a message from the old framework/queue system
type LegacyMessage struct {
	Id          string
	UUID        string
	Retry       int
	Execute     int
	Topic       string
	PayLoad     []byte
	Handler     string
	CreatedOn   time.Time
	ExecutedOn  time.Time
	CompletedOn time.Time
}

// MessageQueueAdapter provides compatibility between old MessageQueue and new JobQueue
type MessageQueueAdapter struct {
	queueID       string
	queueName     string
	jobService    *services.JobService
	queueManager  *DistributedQueueManager
	jobCreator    *IntegrationJobCreator
	logger        logger.Log
	db            *sql.DB
	docDB         *documents.DocDB
	signalRClient signalr.Client
}

// NewMessageQueueAdapter creates an adapter for migrating from old MessageQueue to new JobQueue
func NewMessageQueueAdapter(
	queueID string,
	queueName string,
	db *sql.DB,
	docDB *documents.DocDB,
	signalRClient signalr.Client,
	queueManager *DistributedQueueManager,
) *MessageQueueAdapter {
	return &MessageQueueAdapter{
		queueID:       queueID,
		queueName:     queueName,
		jobService:    services.NewJobService(db),
		queueManager:  queueManager,
		jobCreator:    NewIntegrationJobCreator(db, queueManager),
		logger:        logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MessageQueueAdapter"},
		db:            db,
		docDB:         docDB,
		signalRClient: signalRClient,
	}
}

// Push converts a legacy Message to a QueueJob and enqueues it
func (mqa *MessageQueueAdapter) Push(message LegacyMessage) error {
	ctx := context.Background()

	mqa.logger.Debug(fmt.Sprintf("Adapting legacy message %s to new job system", message.Id))

	// Create payload in legacy-compatible format
	payloadData := map[string]interface{}{
		"Topic":   message.Topic,
		"Payload": string(message.PayLoad),
		"ID":      message.Id,
		"UUID":    message.UUID,
	}

	payloadJSON, err := json.Marshal(payloadData)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create job with legacy metadata
	job := &models.QueueJob{
		ID:         message.Id,
		TypeID:     int(models.JobTypeIntegration),
		Handler:    message.Handler,
		Payload:    string(payloadJSON),
		Priority:   5, // Default priority
		MaxRetries: message.Retry,
		RetryCount: message.Execute,
		Metadata: models.JobMetadata{
			"source":     "legacy_adapter",
			"topic":      message.Topic,
			"uuid":       message.UUID,
			"queue_name": mqa.queueName,
			"queue_id":   mqa.queueID,
		},
		CreatedBy: "legacy_adapter",
		CreatedOn: message.CreatedOn,
	}

	// If message was already executed, set appropriate timestamps
	if !message.ExecutedOn.IsZero() {
		job.StartedAt = &message.ExecutedOn
	}
	if !message.CompletedOn.IsZero() {
		job.CompletedAt = &message.CompletedOn
		job.StatusID = int(models.JobStatusCompleted)
	}

	// Create the job
	err = mqa.jobService.CreateQueueJob(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to create queue job: %w", err)
	}

	// Enqueue in cache (if available)
	if mqa.queueManager != nil {
		err = mqa.queueManager.EnqueueJob(ctx, job.ID, job.Priority)
		if err != nil {
			mqa.logger.Info(fmt.Sprintf("Failed to enqueue job %s in cache: %v", job.ID, err))
			// Don't fail - job is in database and will be picked up by polling
		}
	}

	mqa.logger.Info(fmt.Sprintf("Migrated legacy message %s to job system", message.Id))
	return nil
}

// PushBatch converts multiple legacy Messages to QueueJobs
func (mqa *MessageQueueAdapter) PushBatch(messages []LegacyMessage) error {
	successCount := 0
	errorCount := 0

	for _, message := range messages {
		err := mqa.Push(message)
		if err != nil {
			mqa.logger.Error(fmt.Sprintf("Failed to migrate message %s: %v", message.Id, err))
			errorCount++
		} else {
			successCount++
		}
	}

	mqa.logger.Info(fmt.Sprintf("Batch migration completed: %d succeeded, %d failed", successCount, errorCount))

	if errorCount > 0 {
		return fmt.Errorf("batch migration had %d failures", errorCount)
	}

	return nil
}

// MigrateFromLegacyJobHistory migrates existing Job_History documents to the new system
func (mqa *MessageQueueAdapter) MigrateFromLegacyJobHistory(ctx context.Context, limit int) (int, error) {
	if mqa.docDB == nil {
		return 0, fmt.Errorf("DocumentDB connection not available")
	}

	mqa.logger.Info("Starting migration from legacy Job_History collection...")

	// Query legacy Job_History collection
	filter := map[string]interface{}{
		// Could add filters here, e.g., only recent jobs
	}

	results, err := mqa.docDB.QueryCollection("Job_History", filter, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to query Job_History collection: %w", err)
	}

	migratedCount := 0
	for _, doc := range results {
		// Extract message from legacy document
		messageData, ok := doc["message"].(map[string]interface{})
		if !ok {
			mqa.logger.Info("Skipping document with invalid message structure")
			continue
		}

		// Create legacy message
		legacyMsg := LegacyMessage{
			Id:      getStringField(messageData, "Id"),
			UUID:    getStringField(messageData, "UUID"),
			Topic:   getStringField(messageData, "Topic"),
			Handler: getStringField(messageData, "Handler"),
		}

		// Get payload
		if payloadStr := getStringField(messageData, "PayLoad"); payloadStr != "" {
			legacyMsg.PayLoad = []byte(payloadStr)
		}

		// Get retry/execute counts
		if retry, ok := messageData["Retry"].(float64); ok {
			legacyMsg.Retry = int(retry)
		}
		if execute, ok := messageData["Execute"].(float64); ok {
			legacyMsg.Execute = int(execute)
		}

		// Get timestamps
		if createdOn, ok := messageData["CreatedOn"].(time.Time); ok {
			legacyMsg.CreatedOn = createdOn
		}
		if executedOn, ok := messageData["ExecutedOn"].(time.Time); ok {
			legacyMsg.ExecutedOn = executedOn
		}
		if completedOn, ok := messageData["CompletedOn"].(time.Time); ok {
			legacyMsg.CompletedOn = completedOn
		}

		// Migrate the message
		err := mqa.Push(legacyMsg)
		if err != nil {
			mqa.logger.Error(fmt.Sprintf("Failed to migrate legacy job %s: %v", legacyMsg.Id, err))
			continue
		}

		migratedCount++

		if limit > 0 && migratedCount >= limit {
			break
		}
	}

	mqa.logger.Info(fmt.Sprintf("Migration completed: %d jobs migrated from Job_History", migratedCount))
	return migratedCount, nil
}

// getStringField safely extracts a string field from a map
func getStringField(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetQueueID returns the queue ID
func (mqa *MessageQueueAdapter) GetQueueID() string {
	return mqa.queueID
}

// GetQueueName returns the queue name
func (mqa *MessageQueueAdapter) GetQueueName() string {
	return mqa.queueName
}

// ConvertLegacyMessageToJob converts a legacy Message to a QueueJob (without saving)
func ConvertLegacyMessageToJob(message LegacyMessage, queueName string) *models.QueueJob {
	// Create payload in legacy-compatible format
	payloadData := map[string]interface{}{
		"Topic":   message.Topic,
		"Payload": string(message.PayLoad),
		"ID":      message.Id,
		"UUID":    message.UUID,
	}

	payloadJSON, _ := json.Marshal(payloadData)

	job := &models.QueueJob{
		ID:         message.Id,
		TypeID:     int(models.JobTypeIntegration),
		Handler:    message.Handler,
		Payload:    string(payloadJSON),
		Priority:   5,
		MaxRetries: message.Retry,
		RetryCount: message.Execute,
		Metadata: models.JobMetadata{
			"source":     "legacy_conversion",
			"topic":      message.Topic,
			"uuid":       message.UUID,
			"queue_name": queueName,
		},
		CreatedBy: "legacy_adapter",
		CreatedOn: message.CreatedOn,
	}

	if !message.ExecutedOn.IsZero() {
		job.StartedAt = &message.ExecutedOn
	}
	if !message.CompletedOn.IsZero() {
		job.CompletedAt = &message.CompletedOn
		job.StatusID = int(models.JobStatusCompleted)
	} else {
		job.StatusID = int(models.JobStatusPending)
	}

	return job
}
