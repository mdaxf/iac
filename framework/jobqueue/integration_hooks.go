package jobqueue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// IntegrationJobCreator handles creating jobs from integration messages
type IntegrationJobCreator struct {
	jobService   *services.JobService
	queueManager *DistributedQueueManager
	logger       logger.Log
}

// NewIntegrationJobCreator creates a new integration job creator
func NewIntegrationJobCreator(db *sql.DB, queueManager *DistributedQueueManager) *IntegrationJobCreator {
	return &IntegrationJobCreator{
		jobService:   services.NewJobService(db),
		queueManager: queueManager,
		logger:       logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "IntegrationJobCreator"},
	}
}

// CreateJobFromMessage creates a queue job from an integration message
// This should be called when a message is received through any integration channel
func (ijc *IntegrationJobCreator) CreateJobFromMessage(
	ctx context.Context,
	topic string,
	payload interface{},
	handler string,
	method string,
	protocol string,
	direction models.JobDirection,
	priority int,
) (*models.QueueJob, error) {

	ijc.logger.Debug(fmt.Sprintf("Creating job from integration message: topic=%s, handler=%s, method=%s, protocol=%s, direction=%s",
		topic, handler, method, protocol, direction))

	// Convert payload to JSON string and create payload structure compatible with legacy system
	var payloadStr string
	var rawPayload string

	switch v := payload.(type) {
	case string:
		rawPayload = v
	case []byte:
		rawPayload = string(v)
	default:
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		rawPayload = string(payloadJSON)
	}

	// Create payload in legacy-compatible format (Topic/Payload structure)
	legacyPayload := map[string]interface{}{
		"Topic":   topic,
		"Payload": rawPayload,
	}

	legacyPayloadJSON, err := json.Marshal(legacyPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal legacy payload: %w", err)
	}
	payloadStr = string(legacyPayloadJSON)

	// Create metadata with legacy fields for backward compatibility
	metadata := models.JobMetadata{
		"source":   "integration",
		"topic":    topic,
		"protocol": protocol,
		"method":   method,
		"uuid":     fmt.Sprintf("%s-%d", handler, time.Now().UnixNano()),
	}

	// Create queue job
	job := &models.QueueJob{
		TypeID:     int(models.JobTypeIntegration),
		Method:     method,
		Protocol:   protocol,
		Direction:  direction,
		Handler:    handler,
		Payload:    payloadStr,
		Metadata:   metadata,
		Priority:   priority,
		MaxRetries: 3, // Default retry count
		StatusID:   int(models.JobStatusPending),
		CreatedBy:  "integration",
	}

	// Save to database
	err = ijc.jobService.CreateQueueJob(ctx, job)
	if err != nil {
		ijc.logger.Error(fmt.Sprintf("Failed to create job from message: %v", err))
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Enqueue in cache for distributed processing (if queue manager is available)
	if ijc.queueManager != nil {
		err = ijc.queueManager.EnqueueJob(ctx, job.ID, job.Priority)
		if err != nil {
			ijc.logger.Error(fmt.Sprintf("Failed to enqueue job %s: %v", job.ID, err))
			// Don't fail the operation if cache enqueueing fails
			// The job is still in the database and will be picked up by polling
		}
	}

	ijc.logger.Info(fmt.Sprintf("Created integration job %s for topic %s (handler: %s, direction: %s)",
		job.ID, topic, handler, direction))

	return job, nil
}

// CreateInboundJob creates a job for inbound integration messages
func (ijc *IntegrationJobCreator) CreateInboundJob(
	ctx context.Context,
	topic string,
	payload interface{},
	handler string,
	protocol string,
	method string,
) (*models.QueueJob, error) {

	return ijc.CreateJobFromMessage(
		ctx,
		topic,
		payload,
		handler,
		method,
		protocol,
		models.JobDirectionInbound,
		5, // Default priority for inbound
	)
}

// CreateOutboundJob creates a job for outbound integration messages
func (ijc *IntegrationJobCreator) CreateOutboundJob(
	ctx context.Context,
	topic string,
	payload interface{},
	handler string,
	protocol string,
	method string,
) (*models.QueueJob, error) {

	return ijc.CreateJobFromMessage(
		ctx,
		topic,
		payload,
		handler,
		method,
		protocol,
		models.JobDirectionOutbound,
		5, // Default priority for outbound
	)
}

// CreateJobFromSignalRMessage creates a job from a SignalR message
func (ijc *IntegrationJobCreator) CreateJobFromSignalRMessage(
	ctx context.Context,
	topic string,
	payload interface{},
	handler string,
) (*models.QueueJob, error) {

	return ijc.CreateInboundJob(ctx, topic, payload, handler, "signalr", "websocket")
}

// CreateJobFromKafkaMessage creates a job from a Kafka message
func (ijc *IntegrationJobCreator) CreateJobFromKafkaMessage(
	ctx context.Context,
	topic string,
	payload interface{},
	handler string,
	direction models.JobDirection,
) (*models.QueueJob, error) {

	return ijc.CreateJobFromMessage(ctx, topic, payload, handler, "publish", "kafka", direction, 5)
}

// CreateJobFromMQTTMessage creates a job from an MQTT message
func (ijc *IntegrationJobCreator) CreateJobFromMQTTMessage(
	ctx context.Context,
	topic string,
	payload interface{},
	handler string,
	direction models.JobDirection,
) (*models.QueueJob, error) {

	return ijc.CreateJobFromMessage(ctx, topic, payload, handler, "publish", "mqtt", direction, 5)
}

// CreateJobFromActiveMQMessage creates a job from an ActiveMQ message
func (ijc *IntegrationJobCreator) CreateJobFromActiveMQMessage(
	ctx context.Context,
	topic string,
	payload interface{},
	handler string,
	direction models.JobDirection,
) (*models.QueueJob, error) {

	return ijc.CreateJobFromMessage(ctx, topic, payload, handler, "send", "activemq", direction, 5)
}

// CreateJobFromHTTPRequest creates a job from an HTTP request
func (ijc *IntegrationJobCreator) CreateJobFromHTTPRequest(
	ctx context.Context,
	endpoint string,
	payload interface{},
	handler string,
	method string,
	direction models.JobDirection,
) (*models.QueueJob, error) {

	return ijc.CreateJobFromMessage(ctx, endpoint, payload, handler, method, "http", direction, 7)
}

// BatchCreateJobs creates multiple jobs in a batch
func (ijc *IntegrationJobCreator) BatchCreateJobs(
	ctx context.Context,
	jobs []*models.QueueJob,
) ([]string, error) {

	createdIDs := make([]string, 0, len(jobs))

	for _, job := range jobs {
		err := ijc.jobService.CreateQueueJob(ctx, job)
		if err != nil {
			ijc.logger.Error(fmt.Sprintf("Failed to create job in batch: %v", err))
			continue
		}

		// Enqueue in cache (if queue manager is available)
		if ijc.queueManager != nil {
			ijc.queueManager.EnqueueJob(ctx, job.ID, job.Priority)
		}

		createdIDs = append(createdIDs, job.ID)
	}

	ijc.logger.Info(fmt.Sprintf("Batch created %d jobs", len(createdIDs)))
	return createdIDs, nil
}

// UpdateJobProgress updates the progress metadata for a job
func (ijc *IntegrationJobCreator) UpdateJobProgress(
	ctx context.Context,
	jobID string,
	progress int,
	message string,
) error {

	job, err := ijc.jobService.GetJobByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job.Metadata == nil {
		job.Metadata = make(models.JobMetadata)
	}

	job.Metadata["progress"] = progress
	job.Metadata["progress_message"] = message
	job.Metadata["last_updated"] = fmt.Sprintf("%v", ctx.Value("timestamp"))

	ijc.logger.Debug(fmt.Sprintf("Updated progress for job %s: %d%% - %s", jobID, progress, message))
	return nil
}
