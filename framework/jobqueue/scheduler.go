package jobqueue

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"

	"github.com/robfig/cron/v3"
)

// JobScheduler manages scheduled and interval jobs
type JobScheduler struct {
	jobService      *services.JobService
	queueManager    *DistributedQueueManager
	db              *sql.DB
	logger          logger.Log
	cron            *cron.Cron
	running         bool
	mu              sync.RWMutex
	checkInterval   time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
	scheduledJobs   map[string]cron.EntryID // jobID -> cronEntryID
}

// NewJobScheduler creates a new job scheduler
func NewJobScheduler(db *sql.DB, queueManager *DistributedQueueManager) *JobScheduler {
	checkInterval := time.Duration(config.GlobalConfiguration.JobsConfig.SchedulerCheckInterval) * time.Second
	if checkInterval == 0 {
		checkInterval = 60 * time.Second // Default to 1 minute
	}

	return &JobScheduler{
		jobService:    services.NewJobService(db),
		queueManager:  queueManager,
		db:            db,
		logger:        logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "JobScheduler"},
		cron:          cron.New(cron.WithSeconds()),
		checkInterval: checkInterval,
		scheduledJobs: make(map[string]cron.EntryID),
	}
}

// Start starts the job scheduler
func (js *JobScheduler) Start(ctx context.Context) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	if js.running {
		return fmt.Errorf("scheduler already running")
	}

	js.ctx, js.cancel = context.WithCancel(ctx)
	js.running = true

	js.logger.Info("Starting job scheduler...")

	// Load and schedule jobs
	if err := js.loadScheduledJobs(); err != nil {
		js.logger.Error(fmt.Sprintf("Failed to load scheduled jobs: %v", err))
	}

	// Start cron scheduler
	js.cron.Start()

	// Start periodic check for new/updated scheduled jobs
	go js.runPeriodicCheck()

	js.logger.Info("Job scheduler started successfully")
	return nil
}

// Stop stops the job scheduler
func (js *JobScheduler) Stop() error {
	js.mu.Lock()
	defer js.mu.Unlock()

	if !js.running {
		return fmt.Errorf("scheduler not running")
	}

	js.logger.Info("Stopping job scheduler...")

	// Stop cron
	js.cron.Stop()

	// Cancel context
	js.cancel()

	js.running = false
	js.logger.Info("Job scheduler stopped")
	return nil
}

// loadScheduledJobs loads all scheduled jobs from database and schedules them
func (js *JobScheduler) loadScheduledJobs() error {
	ctx := context.Background()

	// Get all active scheduled jobs
	query := `
		SELECT id, name, description, typeid, handler, cronexpression, intervalseconds,
		       startat, endat, maxexecutions, executioncount, enabled, ` + "`condition`" + `,
		       priority, maxretries, timeout, metadata, lastrunat, nextrunat,
		       active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM jobs
		WHERE active = ? AND enabled = ?
		ORDER BY priority DESC
	`

	rows, err := js.db.QueryContext(ctx, query, true, true)
	if err != nil {
		return fmt.Errorf("failed to load scheduled jobs: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		job := &models.Job{}
		var metadataJSON string

		err := rows.Scan(
			&job.ID, &job.Name, &job.Description, &job.TypeID, &job.Handler, &job.CronExpression, &job.IntervalSeconds,
			&job.StartAt, &job.EndAt, &job.MaxExecutions, &job.ExecutionCount, &job.Enabled, &job.Condition,
			&job.Priority, &job.MaxRetries, &job.Timeout, &metadataJSON, &job.LastRunAt, &job.NextRunAt,
			&job.Active, &job.ReferenceID, &job.CreatedBy, &job.CreatedOn, &job.ModifiedBy, &job.ModifiedOn, &job.RowVersionStamp,
		)

		if err != nil {
			js.logger.Error(fmt.Sprintf("Failed to scan scheduled job: %v", err))
			continue
		}

		// Schedule the job
		if err := js.scheduleJob(job); err != nil {
			js.logger.Error(fmt.Sprintf("Failed to schedule job %s: %v", job.Name, err))
			continue
		}

		count++
	}

	js.logger.Info(fmt.Sprintf("Loaded and scheduled %d jobs", count))
	return nil
}

// scheduleJob schedules a single job based on its configuration
func (js *JobScheduler) scheduleJob(job *models.Job) error {
	// Check if job should start now
	if job.StartAt != nil && job.StartAt.After(time.Now()) {
		js.logger.Info(fmt.Sprintf("Job %s not yet started (starts at %v)", job.Name, job.StartAt))
		return nil
	}

	// Check if job has ended
	if job.EndAt != nil && job.EndAt.Before(time.Now()) {
		js.logger.Info(fmt.Sprintf("Job %s has ended (ended at %v)", job.Name, job.EndAt))
		return nil
	}

	// Check if max executions reached
	if job.MaxExecutions > 0 && job.ExecutionCount >= job.MaxExecutions {
		js.logger.Info(fmt.Sprintf("Job %s has reached max executions (%d)", job.Name, job.MaxExecutions))
		return nil
	}

	// Unschedule if already scheduled
	if entryID, exists := js.scheduledJobs[job.ID]; exists {
		js.cron.Remove(entryID)
		delete(js.scheduledJobs, job.ID)
	}

	// Schedule based on cron expression or interval
	var entryID cron.EntryID
	var err error

	if job.CronExpression != "" {
		// Use cron expression
		entryID, err = js.cron.AddFunc(job.CronExpression, func() {
			js.executeScheduledJob(job)
		})
		if err != nil {
			return fmt.Errorf("failed to add cron job: %w", err)
		}
		js.logger.Info(fmt.Sprintf("Scheduled job %s with cron: %s", job.Name, job.CronExpression))
	} else if job.IntervalSeconds > 0 {
		// Use interval
		interval := time.Duration(job.IntervalSeconds) * time.Second
		entryID, err = js.cron.AddFunc(fmt.Sprintf("@every %s", interval), func() {
			js.executeScheduledJob(job)
		})
		if err != nil {
			return fmt.Errorf("failed to add interval job: %w", err)
		}
		js.logger.Info(fmt.Sprintf("Scheduled job %s with interval: %v", job.Name, interval))
	} else {
		return fmt.Errorf("job %s has no cron expression or interval", job.Name)
	}

	js.scheduledJobs[job.ID] = entryID
	return nil
}

// executeScheduledJob executes a scheduled job by creating a queue job
func (js *JobScheduler) executeScheduledJob(job *models.Job) {
	ctx := context.Background()

	js.logger.Info(fmt.Sprintf("Executing scheduled job: %s", job.Name))

	// Check condition if specified
	if job.Condition != "" {
		shouldRun, err := js.evaluateCondition(ctx, job.Condition)
		if err != nil {
			js.logger.Error(fmt.Sprintf("Failed to evaluate condition for job %s: %v", job.Name, err))
			return
		}
		if !shouldRun {
			js.logger.Info(fmt.Sprintf("Job %s condition not met, skipping execution", job.Name))
			return
		}
	}

	// Check if job has ended
	if job.EndAt != nil && job.EndAt.Before(time.Now()) {
		js.logger.Info(fmt.Sprintf("Job %s has ended, unscheduling", job.Name))
		if entryID, exists := js.scheduledJobs[job.ID]; exists {
			js.cron.Remove(entryID)
			delete(js.scheduledJobs, job.ID)
		}
		return
	}

	// Check if max executions reached
	if job.MaxExecutions > 0 && job.ExecutionCount >= job.MaxExecutions {
		js.logger.Info(fmt.Sprintf("Job %s has reached max executions, unscheduling", job.Name))
		if entryID, exists := js.scheduledJobs[job.ID]; exists {
			js.cron.Remove(entryID)
			delete(js.scheduledJobs, job.ID)
		}
		return
	}

	// Create queue job
	queueJob := &models.QueueJob{
		TypeID:     int(models.JobTypeScheduled),
		Handler:    job.Handler,
		Priority:   job.Priority,
		MaxRetries: job.MaxRetries,
		Metadata: models.JobMetadata{
			"scheduled_job_id":   job.ID,
			"scheduled_job_name": job.Name,
			"execution_count":    job.ExecutionCount + 1,
		},
		CreatedBy: "scheduler",
	}

	// Create the job
	if err := js.jobService.CreateQueueJob(ctx, queueJob); err != nil {
		js.logger.Error(fmt.Sprintf("Failed to create queue job for scheduled job %s: %v", job.Name, err))
		return
	}

	// Enqueue in cache (if queue manager is available)
	if js.queueManager != nil {
		if err := js.queueManager.EnqueueJob(ctx, queueJob.ID, queueJob.Priority); err != nil {
			js.logger.Error(fmt.Sprintf("Failed to enqueue job %s: %v", queueJob.ID, err))
		}
	}

	// Calculate next run time
	nextRunAt := js.calculateNextRunTime(job)

	// Update scheduled job
	if err := js.jobService.UpdateScheduledJobNextRun(ctx, job.ID, nextRunAt); err != nil {
		js.logger.Error(fmt.Sprintf("Failed to update scheduled job next run: %v", err))
	}

	js.logger.Info(fmt.Sprintf("Created queue job %s for scheduled job %s (next run: %v)", queueJob.ID, job.Name, nextRunAt))
}

// calculateNextRunTime calculates the next run time for a scheduled job
func (js *JobScheduler) calculateNextRunTime(job *models.Job) time.Time {
	if job.CronExpression != "" {
		// Parse cron expression and get next run time
		schedule, err := cron.ParseStandard(job.CronExpression)
		if err != nil {
			js.logger.Error(fmt.Sprintf("Failed to parse cron expression: %v", err))
			return time.Now().Add(1 * time.Hour) // Default fallback
		}
		return schedule.Next(time.Now())
	} else if job.IntervalSeconds > 0 {
		return time.Now().Add(time.Duration(job.IntervalSeconds) * time.Second)
	}

	return time.Now().Add(1 * time.Hour) // Default fallback
}

// evaluateCondition evaluates a SQL condition
func (js *JobScheduler) evaluateCondition(ctx context.Context, condition string) (bool, error) {
	query := fmt.Sprintf("SELECT CASE WHEN (%s) THEN 1 ELSE 0 END", condition)

	var result int
	err := js.db.QueryRowContext(ctx, query).Scan(&result)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate condition: %w", err)
	}

	return result == 1, nil
}

// runPeriodicCheck periodically checks for new or updated scheduled jobs
func (js *JobScheduler) runPeriodicCheck() {
	ticker := time.NewTicker(js.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-js.ctx.Done():
			return

		case <-ticker.C:
			if err := js.loadScheduledJobs(); err != nil {
				js.logger.Error(fmt.Sprintf("Failed to reload scheduled jobs: %v", err))
			}
		}
	}
}

// AddJob adds a new scheduled job
func (js *JobScheduler) AddJob(job *models.Job) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	return js.scheduleJob(job)
}

// RemoveJob removes a scheduled job
func (js *JobScheduler) RemoveJob(jobID string) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	if entryID, exists := js.scheduledJobs[jobID]; exists {
		js.cron.Remove(entryID)
		delete(js.scheduledJobs, jobID)
		js.logger.Info(fmt.Sprintf("Removed scheduled job: %s", jobID))
		return nil
	}

	return fmt.Errorf("job not found: %s", jobID)
}

// GetScheduledJobCount returns the number of currently scheduled jobs
func (js *JobScheduler) GetScheduledJobCount() int {
	js.mu.RLock()
	defer js.mu.RUnlock()

	return len(js.scheduledJobs)
}

// IsRunning returns whether the scheduler is running
func (js *JobScheduler) IsRunning() bool {
	js.mu.RLock()
	defer js.mu.RUnlock()

	return js.running
}

// GetStatus returns the current status of the scheduler
func (js *JobScheduler) GetStatus() map[string]interface{} {
	js.mu.RLock()
	defer js.mu.RUnlock()

	return map[string]interface{}{
		"running":        js.running,
		"scheduled_jobs": len(js.scheduledJobs),
		"check_interval": js.checkInterval.String(),
	}
}
