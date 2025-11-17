package jobqueue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/engine/trancode"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac-signalr/signalr"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// JobWorker processes jobs from the queue
type JobWorker struct {
	id              string
	jobService      *services.JobService
	queueManager    *DistributedQueueManager
	db              *sql.DB
	docDB           *documents.DocDB
	signalRClient   signalr.Client
	logger          logger.Log
	running         bool
	mu              sync.RWMutex
	workerCount     int
	pollInterval    time.Duration
	maxRetries      int
	shutdownTimeout time.Duration
	workers         []*Worker
	ctx             context.Context
	cancel          context.CancelFunc
}

// Worker represents a single worker goroutine
type Worker struct {
	id     int
	logger logger.Log
}

// NewJobWorker creates a new job worker
func NewJobWorker(
	id string,
	db *sql.DB,
	docDB *documents.DocDB,
	signalRClient signalr.Client,
	queueManager *DistributedQueueManager,
) *JobWorker {
	jobService := services.NewJobService(db)

	// Get configuration
	workerCount := config.GlobalConfiguration.JobsConfig.Workers
	if workerCount == 0 {
		workerCount = 5
	}

	pollInterval := time.Duration(config.GlobalConfiguration.JobsConfig.PollInterval) * time.Second
	if pollInterval == 0 {
		pollInterval = 5 * time.Second
	}

	maxRetries := config.GlobalConfiguration.JobsConfig.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	return &JobWorker{
		id:              id,
		jobService:      jobService,
		queueManager:    queueManager,
		db:              db,
		docDB:           docDB,
		signalRClient:   signalRClient,
		logger:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "JobWorker"},
		workerCount:     workerCount,
		pollInterval:    pollInterval,
		maxRetries:      maxRetries,
		shutdownTimeout: 30 * time.Second,
	}
}

// Start starts the job worker
func (jw *JobWorker) Start(ctx context.Context) error {
	jw.mu.Lock()
	defer jw.mu.Unlock()

	if jw.running {
		return fmt.Errorf("worker already running")
	}

	jw.ctx, jw.cancel = context.WithCancel(ctx)
	jw.running = true

	jw.logger.Info(fmt.Sprintf("Starting job worker %s with %d workers", jw.id, jw.workerCount))

	// Start worker goroutines
	jw.workers = make([]*Worker, jw.workerCount)
	for i := 0; i < jw.workerCount; i++ {
		worker := &Worker{
			id:     i + 1,
			logger: logger.Log{ModuleName: fmt.Sprintf("Worker-%d", i+1)},
		}
		jw.workers[i] = worker
		go jw.runWorker(worker)
	}

	jw.logger.Info(fmt.Sprintf("Job worker %s started successfully", jw.id))
	return nil
}

// Stop stops the job worker
func (jw *JobWorker) Stop() error {
	jw.mu.Lock()
	defer jw.mu.Unlock()

	if !jw.running {
		return fmt.Errorf("worker not running")
	}

	jw.logger.Info(fmt.Sprintf("Stopping job worker %s...", jw.id))

	// Cancel context to signal workers to stop
	jw.cancel()

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		// Give workers time to finish current jobs
		time.Sleep(jw.shutdownTimeout)
		close(done)
	}()

	<-done

	jw.running = false
	jw.logger.Info(fmt.Sprintf("Job worker %s stopped", jw.id))
	return nil
}

// runWorker is the main worker loop
func (jw *JobWorker) runWorker(worker *Worker) {
	worker.logger.Info(fmt.Sprintf("Worker %d started", worker.id))

	ticker := time.NewTicker(jw.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-jw.ctx.Done():
			worker.logger.Info(fmt.Sprintf("Worker %d shutting down", worker.id))
			return

		case <-ticker.C:
			jw.processNextJob(worker)
		}
	}
}

// processNextJob processes the next available job
func (jw *JobWorker) processNextJob(worker *Worker) {
	ctx := context.Background()

	// Get next pending job from database
	job, err := jw.jobService.GetNextPendingJob(ctx)
	if err != nil {
		worker.logger.Error(fmt.Sprintf("Failed to get next pending job: %v", err))
		return
	}

	if job == nil {
		// No jobs available
		return
	}

	// Try to acquire distributed lock (if queue manager is available)
	if jw.queueManager != nil {
		locked, err := jw.queueManager.AcquireLock(ctx, job.ID, 5*time.Minute)
		if err != nil {
			worker.logger.Error(fmt.Sprintf("Failed to acquire lock for job %s: %v", job.ID, err))
			return
		}

		if !locked {
			worker.logger.Debug(fmt.Sprintf("Could not acquire lock for job %s (already processing)", job.ID))
			return
		}

		// Ensure lock is released when done
		defer func() {
			if err := jw.queueManager.ReleaseLock(ctx, job.ID); err != nil {
				worker.logger.Error(fmt.Sprintf("Failed to release lock for job %s: %v", job.ID, err))
			}
		}()
	}

	// Update job status to processing
	err = jw.jobService.UpdateQueueJobStatus(ctx, job.ID, int(models.JobStatusProcessing), "", "")
	if err != nil {
		worker.logger.Error(fmt.Sprintf("Failed to update job status: %v", err))
		return
	}

	// Update cache status (if queue manager is available)
	if jw.queueManager != nil {
		jw.queueManager.SetJobStatus(ctx, job.ID, int(models.JobStatusProcessing))
	}

	worker.logger.Info(fmt.Sprintf("Processing job %s (Handler: %s, Priority: %d)", job.ID, job.Handler, job.Priority))

	// Execute the job
	jw.executeJob(ctx, worker, job)
}

// executeJob executes a single job
func (jw *JobWorker) executeJob(ctx context.Context, worker *Worker, job *models.QueueJob) {
	startTime := time.Now()

	// Create job history record
	history := &models.JobHistory{
		JobID:        job.ID,
		StatusID:     int(models.JobStatusProcessing),
		StartedAt:    startTime,
		RetryAttempt: job.RetryCount,
		ExecutedBy:   fmt.Sprintf("%s-worker-%d", jw.id, worker.id),
		InputData:    job.Payload,
		Metadata:     job.Metadata,
		CreatedBy:    "system",
	}

	// Execute the job handler
	result, err := jw.executeJobHandler(ctx, job)

	// Calculate duration
	endTime := time.Now()
	history.CompletedAt = &endTime
	history.Duration = endTime.Sub(startTime).Milliseconds()

	// Update job based on result
	if err != nil {
		worker.logger.Error(fmt.Sprintf("Job %s failed: %v", job.ID, err))

		history.StatusID = int(models.JobStatusFailed)
		history.ErrorMessage = err.Error()
		history.Result = fmt.Sprintf("Error: %v", err)

		// Check if should retry
		if job.RetryCount < job.MaxRetries {
			worker.logger.Info(fmt.Sprintf("Retrying job %s (attempt %d/%d)", job.ID, job.RetryCount+1, job.MaxRetries))

			// Increment retry count
			jw.jobService.IncrementRetryCount(ctx, job.ID)

			// Update status to retrying
			jw.jobService.UpdateQueueJobStatus(ctx, job.ID, int(models.JobStatusRetrying), "", err.Error())

			// Update cache status (if queue manager is available)
			if jw.queueManager != nil {
				jw.queueManager.SetJobStatus(ctx, job.ID, int(models.JobStatusRetrying))
				// Re-enqueue with lower priority
				jw.queueManager.EnqueueJob(ctx, job.ID, job.Priority-1)
			}
		} else {
			worker.logger.Error(fmt.Sprintf("Job %s failed permanently after %d retries", job.ID, job.RetryCount))

			// Update status to failed
			jw.jobService.UpdateQueueJobStatus(ctx, job.ID, int(models.JobStatusFailed), "", err.Error())

			// Update cache status (if queue manager is available)
			if jw.queueManager != nil {
				jw.queueManager.SetJobStatus(ctx, job.ID, int(models.JobStatusFailed))
			}
		}
	} else {
		worker.logger.Info(fmt.Sprintf("Job %s completed successfully in %v", job.ID, time.Since(startTime)))

		history.StatusID = int(models.JobStatusCompleted)
		history.OutputData = result
		history.Result = "Success"

		// Update status to completed
		jw.jobService.UpdateQueueJobStatus(ctx, job.ID, int(models.JobStatusCompleted), result, "")

		// Update cache status (if queue manager is available)
		if jw.queueManager != nil {
			jw.queueManager.SetJobStatus(ctx, job.ID, int(models.JobStatusCompleted))
		}
	}

	// Save job history
	if err := jw.jobService.CreateJobHistory(ctx, history); err != nil {
		worker.logger.Error(fmt.Sprintf("Failed to create job history: %v", err))
	}

	// Save to DocumentDB Job_History collection for backward compatibility
	if jw.docDB != nil {
		jw.saveToLegacyJobHistory(ctx, job, history, result, err)
	}

	// Clear cached data for completed job (if queue manager is available)
	if history.StatusID == int(models.JobStatusCompleted) && jw.queueManager != nil {
		jw.queueManager.ClearJobData(ctx, job.ID)
	}
}

// executeJobHandler executes the job handler (transaction code or command)
func (jw *JobWorker) executeJobHandler(ctx context.Context, job *models.QueueJob) (string, error) {
	// Parse payload
	var payloadData map[string]interface{}
	if job.Payload != "" {
		if err := json.Unmarshal([]byte(job.Payload), &payloadData); err != nil {
			payloadData = map[string]interface{}{"raw": job.Payload}
		}
	}

	// Begin transaction
	tx, err := jw.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			jw.logger.Error(fmt.Sprintf("Job %s panicked: %v", job.ID, r))
		}
	}()

	// Execute the handler
	outputs, err := trancode.ExecutebyExternal(job.Handler, payloadData, tx, jw.docDB, jw.signalRClient)

	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("handler execution failed: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Convert outputs to JSON
	outputJSON, err := json.Marshal(outputs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal outputs: %w", err)
	}

	return string(outputJSON), nil
}

// IsRunning returns whether the worker is running
func (jw *JobWorker) IsRunning() bool {
	jw.mu.RLock()
	defer jw.mu.RUnlock()
	return jw.running
}

// GetStatus returns the current status of the worker
func (jw *JobWorker) GetStatus() map[string]interface{} {
	jw.mu.RLock()
	defer jw.mu.RUnlock()

	return map[string]interface{}{
		"id":           jw.id,
		"running":      jw.running,
		"worker_count": jw.workerCount,
		"poll_interval": jw.pollInterval.String(),
		"max_retries":  jw.maxRetries,
	}
}

// saveToLegacyJobHistory saves job execution to DocumentDB Job_History collection
// for backward compatibility with framework/queue system
func (jw *JobWorker) saveToLegacyJobHistory(ctx context.Context, job *models.QueueJob, history *models.JobHistory, result string, jobErr error) {
	// Create legacy message structure
	message := map[string]interface{}{
		"Id":          job.ID,
		"UUID":        job.Metadata["uuid"],
		"Retry":       job.MaxRetries,
		"Execute":     job.RetryCount,
		"Topic":       job.Metadata["topic"],
		"PayLoad":     job.Payload,
		"Handler":     job.Handler,
		"CreatedOn":   job.CreatedOn,
		"ExecutedOn":  history.StartedAt,
		"CompletedOn": history.CompletedAt,
	}

	status := "Success"
	errorMessage := ""
	if jobErr != nil {
		status = "Failed"
		errorMessage = jobErr.Error()
	}

	// Create legacy job history document
	legacyHistory := map[string]interface{}{
		"message":      message,
		"executedon":   history.StartedAt,
		"executedby":   history.ExecutedBy,
		"status":       status,
		"errormessage": errorMessage,
		"messagequeue": job.Metadata["queue_name"],
		"outputs":      result,
		"jobid":        job.ID,
		"executionid":  history.ExecutionID,
		"duration":     history.Duration,
		"retryattempt": history.RetryAttempt,
	}

	// Save to Job_History collection
	_, err := jw.docDB.InsertCollection("Job_History", legacyHistory)
	if err != nil {
		jw.logger.Info(fmt.Sprintf("Failed to save to legacy Job_History collection: %v", err))
	} else {
		jw.logger.Debug(fmt.Sprintf("Saved job %s to legacy Job_History collection", job.ID))
	}
}
