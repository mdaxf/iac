package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"

	"github.com/google/uuid"
)

// JobService provides methods for managing jobs
type JobService struct {
	db   *sql.DB
	iLog logger.Log
}

// NewJobService creates a new job service instance
func NewJobService(db *sql.DB) *JobService {
	return &JobService{
		db:   db,
		iLog: logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "JobService"},
	}
}

// CreateQueueJob creates a new job in the queue
func (js *JobService) CreateQueueJob(ctx context.Context, job *models.QueueJob) error {
	startTime := time.Now()
	defer func() {
		js.iLog.Debug(fmt.Sprintf("CreateQueueJob completed in %v", time.Since(startTime)))
	}()

	if job.ID == "" {
		job.ID = uuid.New().String()
	}

	if job.CreatedOn.IsZero() {
		job.CreatedOn = time.Now()
	}
	job.ModifiedOn = time.Now()
	job.RowVersionStamp = 1
	job.Active = true

	if job.StatusID == 0 {
		if job.ScheduledAt != nil && job.ScheduledAt.After(time.Now()) {
			job.StatusID = int(models.JobStatusScheduled)
		} else {
			job.StatusID = int(models.JobStatusPending)
		}
	}

	metadataJSON, err := json.Marshal(job.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = databases.TableInsert(
		"queue_jobs",
		[]string{
			"id", "typeid", "method", "protocol", "direction", "handler", "metadata", "payload",
			"result", "statusid", "priority", "maxretries", "retrycount", "scheduledat",
			"startedat", "completedat", "lasterror", "parentjobid", "active", "referenceid",
			"createdby", "createdon", "modifiedby", "modifiedon", "rowversionstamp",
		},
		[]interface{}{
			job.ID, job.TypeID, job.Method, job.Protocol, job.Direction, job.Handler, string(metadataJSON), job.Payload,
			job.Result, job.StatusID, job.Priority, job.MaxRetries, job.RetryCount, job.ScheduledAt,
			job.StartedAt, job.CompletedAt, job.LastError, job.ParentJobID, job.Active, job.ReferenceID,
			job.CreatedBy, job.CreatedOn, job.ModifiedBy, job.ModifiedOn, job.RowVersionStamp,
		},
		js.db,
		ctx,
	)

	if err != nil {
		js.iLog.Error(fmt.Sprintf("Failed to create queue job: %v", err))
		return fmt.Errorf("failed to create queue job: %w", err)
	}

	js.iLog.Info(fmt.Sprintf("Created queue job: %s (Handler: %s, Priority: %d)", job.ID, job.Handler, job.Priority))
	return nil
}

// UpdateQueueJobStatus updates the status of a queue job
func (js *JobService) UpdateQueueJobStatus(ctx context.Context, jobID string, statusID int, result string, errorMsg string) error {
	updateData := map[string]interface{}{
		"statusid":   statusID,
		"modifiedon": time.Now(),
	}

	if result != "" {
		updateData["result"] = result
	}

	if errorMsg != "" {
		updateData["lasterror"] = errorMsg
	}

	if statusID == int(models.JobStatusProcessing) {
		updateData["startedat"] = time.Now()
	} else if statusID == int(models.JobStatusCompleted) || statusID == int(models.JobStatusFailed) || statusID == int(models.JobStatusCancelled) {
		updateData["completedat"] = time.Now()
	}

	_, err := databases.TableUpdate(
		"queue_jobs",
		updateData,
		map[string]interface{}{"id": jobID},
		js.db,
		ctx,
	)

	if err != nil {
		js.iLog.Error(fmt.Sprintf("Failed to update job status: %v", err))
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// IncrementRetryCount increments the retry count for a job
func (js *JobService) IncrementRetryCount(ctx context.Context, jobID string) error {
	query := `UPDATE queue_jobs SET retrycount = retrycount + 1, modifiedon = ? WHERE id = ?`

	_, err := js.db.ExecContext(ctx, query, time.Now(), jobID)
	if err != nil {
		js.iLog.Error(fmt.Sprintf("Failed to increment retry count: %v", err))
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	return nil
}

// GetNextPendingJob retrieves the next pending job from the queue
func (js *JobService) GetNextPendingJob(ctx context.Context) (*models.QueueJob, error) {
	query := `
		SELECT id, typeid, method, protocol, direction, handler, metadata, payload,
		       result, statusid, priority, maxretries, retrycount, scheduledat,
		       startedat, completedat, lasterror, parentjobid, active, referenceid,
		       createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM queue_jobs
		WHERE active = ?
		  AND statusid IN (?, ?)
		  AND (scheduledat IS NULL OR scheduledat <= ?)
		ORDER BY priority DESC, createdon ASC
		LIMIT 1
	`

	row := js.db.QueryRowContext(
		ctx,
		query,
		true,
		int(models.JobStatusPending),
		int(models.JobStatusQueued),
		time.Now(),
	)

	job := &models.QueueJob{}
	var metadataJSON string

	err := row.Scan(
		&job.ID, &job.TypeID, &job.Method, &job.Protocol, &job.Direction, &job.Handler, &metadataJSON, &job.Payload,
		&job.Result, &job.StatusID, &job.Priority, &job.MaxRetries, &job.RetryCount, &job.ScheduledAt,
		&job.StartedAt, &job.CompletedAt, &job.LastError, &job.ParentJobID, &job.Active, &job.ReferenceID,
		&job.CreatedBy, &job.CreatedOn, &job.ModifiedBy, &job.ModifiedOn, &job.RowVersionStamp,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		js.iLog.Error(fmt.Sprintf("Failed to get next pending job: %v", err))
		return nil, fmt.Errorf("failed to get next pending job: %w", err)
	}

	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &job.Metadata); err != nil {
			js.iLog.Warning(fmt.Sprintf("Failed to unmarshal metadata for job %s: %v", job.ID, err))
		}
	}

	return job, nil
}

// CreateJobHistory creates a job execution history record
func (js *JobService) CreateJobHistory(ctx context.Context, history *models.JobHistory) error {
	if history.ID == "" {
		history.ID = uuid.New().String()
	}

	if history.ExecutionID == "" {
		history.ExecutionID = uuid.New().String()
	}

	if history.CreatedOn.IsZero() {
		history.CreatedOn = time.Now()
	}
	history.ModifiedOn = time.Now()
	history.RowVersionStamp = 1
	history.Active = true

	if !history.StartedAt.IsZero() && history.CompletedAt != nil {
		history.Duration = history.CompletedAt.Sub(history.StartedAt).Milliseconds()
	}

	metadataJSON, err := json.Marshal(history.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = databases.TableInsert(
		"job_histories",
		[]string{
			"id", "jobid", "executionid", "statusid", "startedat", "completedat",
			"duration", "result", "errormessage", "retryattempt", "executedby",
			"inputdata", "outputdata", "metadata", "active", "referenceid",
			"createdby", "createdon", "modifiedby", "modifiedon", "rowversionstamp",
		},
		[]interface{}{
			history.ID, history.JobID, history.ExecutionID, history.StatusID, history.StartedAt, history.CompletedAt,
			history.Duration, history.Result, history.ErrorMessage, history.RetryAttempt, history.ExecutedBy,
			history.InputData, history.OutputData, string(metadataJSON), history.Active, history.ReferenceID,
			history.CreatedBy, history.CreatedOn, history.ModifiedBy, history.ModifiedOn, history.RowVersionStamp,
		},
		js.db,
		ctx,
	)

	if err != nil {
		js.iLog.Error(fmt.Sprintf("Failed to create job history: %v", err))
		return fmt.Errorf("failed to create job history: %w", err)
	}

	return nil
}

// GetJobByID retrieves a job by its ID
func (js *JobService) GetJobByID(ctx context.Context, jobID string) (*models.QueueJob, error) {
	query := `
		SELECT id, typeid, method, protocol, direction, handler, metadata, payload,
		       result, statusid, priority, maxretries, retrycount, scheduledat,
		       startedat, completedat, lasterror, parentjobid, active, referenceid,
		       createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM queue_jobs
		WHERE id = ?
	`

	row := js.db.QueryRowContext(ctx, query, jobID)

	job := &models.QueueJob{}
	var metadataJSON string

	err := row.Scan(
		&job.ID, &job.TypeID, &job.Method, &job.Protocol, &job.Direction, &job.Handler, &metadataJSON, &job.Payload,
		&job.Result, &job.StatusID, &job.Priority, &job.MaxRetries, &job.RetryCount, &job.ScheduledAt,
		&job.StartedAt, &job.CompletedAt, &job.LastError, &job.ParentJobID, &job.Active, &job.ReferenceID,
		&job.CreatedBy, &job.CreatedOn, &job.ModifiedBy, &job.ModifiedOn, &job.RowVersionStamp,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	if err != nil {
		js.iLog.Error(fmt.Sprintf("Failed to get job by ID: %v", err))
		return nil, fmt.Errorf("failed to get job by ID: %w", err)
	}

	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &job.Metadata); err != nil {
			js.iLog.Warning(fmt.Sprintf("Failed to unmarshal metadata for job %s: %v", job.ID, err))
		}
	}

	return job, nil
}

// GetScheduledJobs retrieves all enabled scheduled jobs that should run
func (js *JobService) GetScheduledJobs(ctx context.Context) ([]*models.Job, error) {
	query := `
		SELECT id, name, description, typeid, handler, cronexpression, intervalseconds,
		       startat, endat, maxexecutions, executioncount, enabled, ` + "`condition`" + `,
		       priority, maxretries, timeout, metadata, lastrunat, nextrunat,
		       active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM jobs
		WHERE active = ? AND enabled = ?
		  AND (nextrunat IS NULL OR nextrunat <= ?)
		  AND (endat IS NULL OR endat >= ?)
		  AND (maxexecutions = 0 OR executioncount < maxexecutions)
		ORDER BY priority DESC, nextrunat ASC
	`

	rows, err := js.db.QueryContext(ctx, query, true, true, time.Now(), time.Now())
	if err != nil {
		js.iLog.Error(fmt.Sprintf("Failed to get scheduled jobs: %v", err))
		return nil, fmt.Errorf("failed to get scheduled jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.Job
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
			js.iLog.Error(fmt.Sprintf("Failed to scan scheduled job: %v", err))
			continue
		}

		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &job.Metadata); err != nil {
				js.iLog.Warning(fmt.Sprintf("Failed to unmarshal metadata for job %s: %v", job.ID, err))
			}
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// UpdateScheduledJobNextRun updates the next run time for a scheduled job
func (js *JobService) UpdateScheduledJobNextRun(ctx context.Context, jobID string, nextRunAt time.Time) error {
	updateData := map[string]interface{}{
		"nextrunat":      nextRunAt,
		"lastrunat":      time.Now(),
		"executioncount": "executioncount + 1", // This will be handled specially
		"modifiedon":     time.Now(),
	}

	query := `
		UPDATE jobs
		SET nextrunat = ?, lastrunat = ?, executioncount = executioncount + 1, modifiedon = ?
		WHERE id = ?
	`

	_, err := js.db.ExecContext(ctx, query, nextRunAt, time.Now(), time.Now(), jobID)
	if err != nil {
		js.iLog.Error(fmt.Sprintf("Failed to update scheduled job next run: %v", err))
		return fmt.Errorf("failed to update scheduled job next run: %w", err)
	}

	return nil
}

// CreateJobFromIntegrationMessage creates a queue job from an integration message
func (js *JobService) CreateJobFromIntegrationMessage(ctx context.Context, method, protocol, direction, handler, payload string, metadata models.JobMetadata) (*models.QueueJob, error) {
	job := &models.QueueJob{
		TypeID:     int(models.JobTypeIntegration),
		Method:     method,
		Protocol:   protocol,
		Direction:  models.JobDirection(direction),
		Handler:    handler,
		Payload:    payload,
		Metadata:   metadata,
		Priority:   5, // Default priority for integration jobs
		MaxRetries: config.GlobalConfiguration.JobsConfig.MaxRetries,
		CreatedBy:  "integration",
	}

	err := js.CreateQueueJob(ctx, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// GetJobStatistics retrieves job statistics
func (js *JobService) GetJobStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count jobs by status
	query := `
		SELECT statusid, COUNT(*) as count
		FROM queue_jobs
		WHERE active = ?
		GROUP BY statusid
	`

	rows, err := js.db.QueryContext(ctx, query, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get job statistics: %w", err)
	}
	defer rows.Close()

	statusCounts := make(map[int]int)
	for rows.Next() {
		var statusID, count int
		if err := rows.Scan(&statusID, &count); err != nil {
			continue
		}
		statusCounts[statusID] = count
	}

	stats["status_counts"] = statusCounts
	stats["timestamp"] = time.Now()

	return stats, nil
}
