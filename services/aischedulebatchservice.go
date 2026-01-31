// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/com"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// AIScheduleBatchService handles batch AI schedule operations
type AIScheduleBatchService struct {
	iLog            logger.Log
	db              *sql.DB
	docdb           *documents.DocDB
	scheduleService *AIScheduleService
}

// OptimizeScheduleBatchRequest represents an async batch optimization request
type OptimizeScheduleBatchRequest struct {
	CurrentTasks []Task                 `json:"currentTasks"`
	Resources    []Resource             `json:"resources"`
	Objective    string                 `json:"objective"`
	Constraints  map[string]interface{} `json:"constraints,omitempty"`
	SystemPrompt string                 `json:"systemPrompt,omitempty"`
	BatchSize    int                    `json:"batchSize,omitempty"` // Default: 30 tasks per batch
}

// OptimizeScheduleBatchResponse represents the initial batch job response
type OptimizeScheduleBatchResponse struct {
	JobID        string `json:"jobId"`
	TotalBatches int    `json:"totalBatches"`
	TotalTasks   int    `json:"totalTasks"`
	BatchSize    int    `json:"batchSize"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

// BatchJobStatusResponse represents the status of a batch job
type BatchJobStatusResponse struct {
	JobID            string                         `json:"jobId"`
	Status           models.AIScheduleBatchJobStatus `json:"status"`
	Progress         int                            `json:"progress"`
	TotalBatches     int                            `json:"totalBatches"`
	CompletedBatches int                            `json:"completedBatches"`
	FailedBatches    int                            `json:"failedBatches"`
	TotalTasks       int                            `json:"totalTasks"`
	ProcessedTasks   int                            `json:"processedTasks"`
	StartedAt        *time.Time                     `json:"startedAt,omitempty"`
	CompletedAt      *time.Time                     `json:"completedAt,omitempty"`
	ErrorMessage     string                         `json:"errorMessage,omitempty"`
}

// BatchJobResultResponse represents the final aggregated result
type BatchJobResultResponse struct {
	JobID    string                 `json:"jobId"`
	Status   string                 `json:"status"`
	Tasks    []Task                 `json:"tasks"`
	Changes  []ScheduleChange       `json:"changes"`
	Metadata map[string]interface{} `json:"metadata"`
}

// BatchProcessPayload represents the payload for batch processing job
type BatchProcessPayload struct {
	BatchJobID   string                      `json:"batchJobId"`
	BatchNumber  int                         `json:"batchNumber"`
	Tasks        []Task                      `json:"tasks"`
	Resources    []Resource                  `json:"resources"`
	Objective    string                      `json:"objective"`
	Constraints  map[string]interface{}      `json:"constraints,omitempty"`
	SystemPrompt string                      `json:"systemPrompt,omitempty"`
	StartIndex   int                         `json:"startIndex"`
	EndIndex     int                         `json:"endIndex"`
}

// NewAIScheduleBatchService creates a new batch service
func NewAIScheduleBatchService(db *sql.DB, docdb *documents.DocDB) *AIScheduleBatchService {
	return &AIScheduleBatchService{
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "AIScheduleBatchService",
		},
		db:              db,
		docdb:           docdb,
		scheduleService: NewAIScheduleService(),
	}
}

// convertPlaceholders converts ? placeholders to $1, $2, etc. for PostgreSQL
func (s *AIScheduleBatchService) convertPlaceholders(query string) string {
	dbType := strings.ToLower(dbconn.DatabaseType)
	if dbType == "postgres" || dbType == "postgresql" {
		// Convert ? to $1, $2, $3, etc.
		result := query
		count := 1
		for strings.Contains(result, "?") {
			result = strings.Replace(result, "?", fmt.Sprintf("$%d", count), 1)
			count++
		}
		return result
	}
	return query
}

// sendJobStatusNotification sends a SignalR notification about job status change
func (s *AIScheduleBatchService) sendJobStatusNotification(job *models.AIScheduleBatchJob) {
	// Check if SignalR client is available
	if com.IACMessageBusClient == nil {
		s.iLog.Debug("SignalR client not available, skipping notification")
		return
	}

	notification := map[string]interface{}{
		"type":             "AI_JOB_STATUS_UPDATE",
		"jobId":            job.ID,
		"status":           job.Status,
		"progress":         job.Progress,
		"completedBatches": job.CompletedBatches,
		"totalBatches":     job.TotalBatches,
		"processedTasks":   job.ProcessedTasks,
		"totalTasks":       job.TotalTasks,
	}

	if job.CompletedAt != nil {
		notification["completedAt"] = job.CompletedAt
	}

	// Convert notification to JSON string
	jsonData, err := json.Marshal(notification)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to marshal SignalR notification: %v", err))
		return
	}

	// Send notification via SignalR (non-blocking)
	// PERFORMANCE FIX: Improved error handling and graceful failure
	go func() {
		defer func() {
			// Recover from any panic in SignalR client
			if r := recover(); r != nil {
				s.iLog.Warn(fmt.Sprintf("SignalR notification panic recovered: %v", r))
			}
		}()

		// Try to send notification
		resultChan := com.IACMessageBusClient.Invoke("send", "ai_job_status", string(jsonData), "")

		// Wait for result with timeout
		select {
		case result := <-resultChan:
			if result.Error != nil {
				s.iLog.Debug(fmt.Sprintf("SignalR notification error: %s", result.Error.Error()))
			} else {
				s.iLog.Debug(fmt.Sprintf("SignalR notification sent for job %s (status: %s, progress: %d%%)", job.ID, job.Status, job.Progress))
			}
		case <-time.After(2 * time.Second):
			s.iLog.Debug("SignalR notification timeout (client may not be connected)")
		}
	}()
}

// OptimizeScheduleBatch creates batch jobs for async optimization
func (s *AIScheduleBatchService) OptimizeScheduleBatch(ctx context.Context, req OptimizeScheduleBatchRequest, user string) (*OptimizeScheduleBatchResponse, error) {
	s.iLog.Info(fmt.Sprintf("OptimizeScheduleBatch START - Tasks: %d, Objective: %s", len(req.CurrentTasks), req.Objective))

	// Set default batch size
	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 30 // Default: 30 tasks per batch
	}

	// Validate minimum batch size
	if batchSize < 10 {
		batchSize = 10
	}

	// Validate maximum batch size
	if batchSize > 100 {
		batchSize = 100
	}

	totalTasks := len(req.CurrentTasks)
	totalBatches := (totalTasks + batchSize - 1) / batchSize // Ceiling division

	// Create parent batch job record
	jobID := uuid.New().String()
	now := time.Now()

	batchJob := &models.AIScheduleBatchJob{
		ID:               jobID,
		TotalBatches:     totalBatches,
		CompletedBatches: 0,
		FailedBatches:    0,
		Status:           models.BatchStatusPending,
		Objective:        req.Objective,
		TotalTasks:       totalTasks,
		ProcessedTasks:   0,
		BatchSize:        batchSize,
		Progress:         0,
		StartedAt:        &now,
		Active:           true,
		CreatedBy:        user,
		CreatedOn:        now,
		ModifiedBy:       user,
		ModifiedOn:       now,
		RowVersionStamp:  1,
	}

	// Save batch job to database
	if err := s.createBatchJob(ctx, batchJob); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to create batch job: %v", err))
		return nil, fmt.Errorf("failed to create batch job: %w", err)
	}

	// Send initial notification
	s.sendJobStatusNotification(batchJob)

	// Create individual batch processing jobs
	for i := 0; i < totalBatches; i++ {
		startIdx := i * batchSize
		endIdx := startIdx + batchSize
		if endIdx > totalTasks {
			endIdx = totalTasks
		}

		batchTasks := req.CurrentTasks[startIdx:endIdx]

		// Create batch process payload
		payload := BatchProcessPayload{
			BatchJobID:   jobID,
			BatchNumber:  i + 1,
			Tasks:        batchTasks,
			Resources:    req.Resources,
			Objective:    req.Objective,
			Constraints:  req.Constraints,
			SystemPrompt: req.SystemPrompt,
			StartIndex:   startIdx,
			EndIndex:     endIdx,
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to marshal batch payload: %v", err))
			continue
		}

		// Create queue job for this batch
		queueJob := &models.QueueJob{
			ID:         uuid.New().String(),
			TypeID:     int(models.JobTypeSystem),
			Handler:    "ai.schedule.batch.process",
			Payload:    string(payloadJSON),
			Priority:   5,
			MaxRetries: 2,
			Metadata: models.JobMetadata{
				"batch_job_id": jobID,
				"batch_number": i + 1,
				"task_count":   len(batchTasks),
			},
			Active:     true,
			CreatedBy:  user,
			CreatedOn:  now,
			ModifiedBy: user,
			ModifiedOn: now,
		}

		jobService := NewJobService(s.db)
		if err := jobService.CreateQueueJob(ctx, queueJob); err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to create queue job for batch %d: %v", i+1, err))
			// Continue creating other batches even if one fails
		}

		// Create batch result placeholder
		batchResult := &models.AIScheduleBatchResult{
			ID:              uuid.New().String(),
			BatchJobID:      jobID,
			BatchNumber:     i + 1,
			StartIndex:      startIdx,
			EndIndex:        endIdx,
			Status:          "pending",
			Active:          true,
			CreatedBy:       user,
			CreatedOn:       now,
			ModifiedBy:      user,
			ModifiedOn:      now,
			RowVersionStamp: 1,
		}

		if err := s.createBatchResult(ctx, batchResult); err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to create batch result record: %v", err))
		}
	}

	// Update job status to processing
	batchJob.Status = models.BatchStatusProcessing
	if err := s.updateBatchJobStatus(ctx, jobID, models.BatchStatusProcessing, 0); err != nil {
		s.iLog.Warn(fmt.Sprintf("Failed to update batch job status: %v", err))
	} else {
		// Send notification after status update
		s.sendJobStatusNotification(batchJob)
	}

	s.iLog.Info(fmt.Sprintf("OptimizeScheduleBatch COMPLETE - JobID: %s, Batches: %d", jobID, totalBatches))

	return &OptimizeScheduleBatchResponse{
		JobID:        jobID,
		TotalBatches: totalBatches,
		TotalTasks:   totalTasks,
		BatchSize:    batchSize,
		Status:       string(models.BatchStatusProcessing),
		Message:      fmt.Sprintf("Batch optimization started with %d batches", totalBatches),
	}, nil
}

// ProcessBatch processes a single batch (called by job worker)
func (s *AIScheduleBatchService) ProcessBatch(ctx context.Context, payload BatchProcessPayload, user string) error {
	s.iLog.Info(fmt.Sprintf("ProcessBatch START - Job: %s, Batch: %d", payload.BatchJobID, payload.BatchNumber))

	// Create optimize request for this batch
	req := OptimizeScheduleRequest{
		CurrentTasks: payload.Tasks,
		Resources:    payload.Resources,
		Objective:    payload.Objective,
		Constraints:  payload.Constraints,
		SystemPrompt: payload.SystemPrompt,
	}

	// Call the AI optimization service
	response, err := s.scheduleService.OptimizeSchedule(ctx, req)

	now := time.Now()

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Batch %d failed: %v", payload.BatchNumber, err))

		// Update batch result with error
		if err := s.updateBatchResult(ctx, payload.BatchJobID, payload.BatchNumber, nil, nil, "failed", err.Error(), &now); err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to update batch result: %v", err))
		}

		// Increment failed batch count
		if err := s.incrementFailedBatch(ctx, payload.BatchJobID); err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to increment failed batch count: %v", err))
		}

		// Send notification after batch failure
		if status, err := s.GetBatchJobStatus(ctx, payload.BatchJobID); err == nil {
			job := &models.AIScheduleBatchJob{
				ID:               status.JobID,
				Status:           status.Status,
				Progress:         status.Progress,
				TotalBatches:     status.TotalBatches,
				CompletedBatches: status.CompletedBatches,
				FailedBatches:    status.FailedBatches,
				TotalTasks:       status.TotalTasks,
				ProcessedTasks:   status.ProcessedTasks,
				CompletedAt:      status.CompletedAt,
			}
			s.sendJobStatusNotification(job)
		}

		return fmt.Errorf("batch processing failed: %w", err)
	}

	// Convert response to JSON
	tasksJSON, err := json.Marshal(response.Tasks)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to marshal tasks: %v", err))
		return err
	}

	changesJSON, err := json.Marshal(response.Changes)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to marshal changes: %v", err))
		return err
	}

	// Update batch result with success
	if err := s.updateBatchResult(ctx, payload.BatchJobID, payload.BatchNumber, tasksJSON, changesJSON, "completed", "", &now); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to update batch result: %v", err))
		return err
	}

	// Increment completed batch count and update progress
	if err := s.incrementCompletedBatch(ctx, payload.BatchJobID, len(payload.Tasks)); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to increment completed batch count: %v", err))
		return err
	}

	// Send notification after batch completion
	if status, err := s.GetBatchJobStatus(ctx, payload.BatchJobID); err == nil {
		job := &models.AIScheduleBatchJob{
			ID:               status.JobID,
			Status:           status.Status,
			Progress:         status.Progress,
			TotalBatches:     status.TotalBatches,
			CompletedBatches: status.CompletedBatches,
			FailedBatches:    status.FailedBatches,
			TotalTasks:       status.TotalTasks,
			ProcessedTasks:   status.ProcessedTasks,
			CompletedAt:      status.CompletedAt,
		}
		s.sendJobStatusNotification(job)
	}

	// Check if all batches are complete
	if err := s.checkAndFinalizeBatchJob(ctx, payload.BatchJobID); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to finalize batch job: %v", err))
	}

	s.iLog.Info(fmt.Sprintf("ProcessBatch COMPLETE - Batch: %d, Tasks: %d", payload.BatchNumber, len(response.Tasks)))

	return nil
}

// GetBatchJobStatus retrieves the status of a batch job
func (s *AIScheduleBatchService) GetBatchJobStatus(ctx context.Context, jobID string) (*BatchJobStatusResponse, error) {
	query := `SELECT id, status, progress, total_batches, completed_batches, failed_batches,
	          total_tasks, processed_tasks, started_at, completed_at, error_message
	          FROM ai_schedule_batch_jobs WHERE id = ? AND active = TRUE`

	query = s.convertPlaceholders(query)

	var job models.AIScheduleBatchJob
	var errorMessage sql.NullString

	err := s.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID, &job.Status, &job.Progress, &job.TotalBatches, &job.CompletedBatches,
		&job.FailedBatches, &job.TotalTasks, &job.ProcessedTasks, &job.StartedAt,
		&job.CompletedAt, &errorMessage,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("batch job not found: %s", jobID)
		}
		return nil, fmt.Errorf("failed to get batch job status: %w", err)
	}

	// Convert sql.NullString to string
	if errorMessage.Valid {
		job.ErrorMessage = errorMessage.String
	}

	return &BatchJobStatusResponse{
		JobID:            job.ID,
		Status:           job.Status,
		Progress:         job.Progress,
		TotalBatches:     job.TotalBatches,
		CompletedBatches: job.CompletedBatches,
		FailedBatches:    job.FailedBatches,
		TotalTasks:       job.TotalTasks,
		ProcessedTasks:   job.ProcessedTasks,
		StartedAt:        job.StartedAt,
		CompletedAt:      job.CompletedAt,
		ErrorMessage:     job.ErrorMessage,
	}, nil
}

// GetBatchJobResult retrieves the aggregated result of a completed batch job
func (s *AIScheduleBatchService) GetBatchJobResult(ctx context.Context, jobID string) (*BatchJobResultResponse, error) {
	s.iLog.Info(fmt.Sprintf("GetBatchJobResult START - JobID: %s", jobID))

	// Check job status
	status, err := s.GetBatchJobStatus(ctx, jobID)
	if err != nil {
		return nil, err
	}

	if status.Status != models.BatchStatusCompleted && status.Status != models.BatchStatusPartial {
		return &BatchJobResultResponse{
			JobID:  jobID,
			Status: string(status.Status),
			Metadata: map[string]interface{}{
				"progress":         status.Progress,
				"completedBatches": status.CompletedBatches,
				"totalBatches":     status.TotalBatches,
			},
		}, nil
	}

	// Retrieve all batch results
	query := `SELECT batch_number, tasks_json, changes_json, status, start_index, end_index
	          FROM ai_schedule_batch_results
	          WHERE batch_job_id = ? AND active = TRUE
	          ORDER BY batch_number ASC`

	query = s.convertPlaceholders(query)

	rows, err := s.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve batch results: %w", err)
	}
	defer rows.Close()

	var allTasks []Task
	var allChanges []ScheduleChange

	for rows.Next() {
		var batchNum, startIdx, endIdx int
		var tasksJSON, changesJSON sql.NullString
		var batchStatus string

		if err := rows.Scan(&batchNum, &tasksJSON, &changesJSON, &batchStatus, &startIdx, &endIdx); err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to scan batch result: %v", err))
			continue
		}

		if batchStatus != "completed" {
			continue
		}

		// Parse tasks
		var tasks []Task
		if tasksJSON.Valid && tasksJSON.String != "" {
			if err := json.Unmarshal([]byte(tasksJSON.String), &tasks); err != nil {
				s.iLog.Error(fmt.Sprintf("Failed to unmarshal tasks for batch %d: %v", batchNum, err))
				continue
			}
		}

		// Parse changes
		var changes []ScheduleChange
		if changesJSON.Valid && changesJSON.String != "" {
			if err := json.Unmarshal([]byte(changesJSON.String), &changes); err != nil {
				s.iLog.Error(fmt.Sprintf("Failed to unmarshal changes for batch %d: %v", batchNum, err))
				continue
			}
		}

		allTasks = append(allTasks, tasks...)
		allChanges = append(allChanges, changes...)
	}

	s.iLog.Info(fmt.Sprintf("GetBatchJobResult COMPLETE - Tasks: %d, Changes: %d", len(allTasks), len(allChanges)))

	return &BatchJobResultResponse{
		JobID:   jobID,
		Status:  string(status.Status),
		Tasks:   allTasks,
		Changes: allChanges,
		Metadata: map[string]interface{}{
			"totalBatches":     status.TotalBatches,
			"completedBatches": status.CompletedBatches,
			"failedBatches":    status.FailedBatches,
			"totalTasks":       status.TotalTasks,
		},
	}, nil
}

// Database helper methods

func (s *AIScheduleBatchService) createBatchJob(ctx context.Context, job *models.AIScheduleBatchJob) error {
	query := `INSERT INTO ai_schedule_batch_jobs
	          (id, total_batches, completed_batches, failed_batches, status, objective,
	           total_tasks, processed_tasks, batch_size, progress, started_at,
	           modifiedon, modifiedby, createdon, createdby, active, rowversionstamp)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	query = s.convertPlaceholders(query)

	_, err := s.db.ExecContext(ctx, query,
		job.ID, job.TotalBatches, job.CompletedBatches, job.FailedBatches, job.Status,
		job.Objective, job.TotalTasks, job.ProcessedTasks, job.BatchSize, job.Progress,
		job.StartedAt, job.ModifiedOn, job.ModifiedBy, job.CreatedOn, job.CreatedBy,
		job.Active, job.RowVersionStamp,
	)

	return err
}

func (s *AIScheduleBatchService) createBatchResult(ctx context.Context, result *models.AIScheduleBatchResult) error {
	query := `INSERT INTO ai_schedule_batch_results
	          (id, batch_job_id, batch_number, start_index, end_index, status,
	           modifiedon, modifiedby, createdon, createdby, active, rowversionstamp)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	query = s.convertPlaceholders(query)

	_, err := s.db.ExecContext(ctx, query,
		result.ID, result.BatchJobID, result.BatchNumber, result.StartIndex, result.EndIndex,
		result.Status, result.ModifiedOn, result.ModifiedBy, result.CreatedOn, result.CreatedBy,
		result.Active, result.RowVersionStamp,
	)

	return err
}

func (s *AIScheduleBatchService) updateBatchResult(ctx context.Context, batchJobID string, batchNumber int, tasksJSON, changesJSON []byte, status, errorMsg string, processedAt *time.Time) error {
	query := `UPDATE ai_schedule_batch_results
	          SET tasks_json = ?, changes_json = ?, status = ?, error_message = ?,
	              processed_at = ?, modifiedon = ?, rowversionstamp = rowversionstamp + 1
	          WHERE batch_job_id = ? AND batch_number = ?`

	query = s.convertPlaceholders(query)

	_, err := s.db.ExecContext(ctx, query, string(tasksJSON), string(changesJSON), status, errorMsg, processedAt, time.Now(), batchJobID, batchNumber)
	return err
}

func (s *AIScheduleBatchService) updateBatchJobStatus(ctx context.Context, jobID string, status models.AIScheduleBatchJobStatus, progress int) error {
	query := `UPDATE ai_schedule_batch_jobs
	          SET status = ?, progress = ?, modifiedon = ?, rowversionstamp = rowversionstamp + 1
	          WHERE id = ?`

	query = s.convertPlaceholders(query)

	_, err := s.db.ExecContext(ctx, query, status, progress, time.Now(), jobID)
	return err
}

func (s *AIScheduleBatchService) incrementCompletedBatch(ctx context.Context, jobID string, tasksProcessed int) error {
	query := `UPDATE ai_schedule_batch_jobs
	          SET completed_batches = completed_batches + 1,
	              processed_tasks = processed_tasks + ?,
	              modifiedon = ?,
	              rowversionstamp = rowversionstamp + 1
	          WHERE id = ?`

	query = s.convertPlaceholders(query)

	_, err := s.db.ExecContext(ctx, query, tasksProcessed, time.Now(), jobID)
	return err
}

func (s *AIScheduleBatchService) incrementFailedBatch(ctx context.Context, jobID string) error {
	query := `UPDATE ai_schedule_batch_jobs
	          SET failed_batches = failed_batches + 1,
	              modifiedon = ?,
	              rowversionstamp = rowversionstamp + 1
	          WHERE id = ?`

	query = s.convertPlaceholders(query)

	_, err := s.db.ExecContext(ctx, query, time.Now(), jobID)
	return err
}

func (s *AIScheduleBatchService) checkAndFinalizeBatchJob(ctx context.Context, jobID string) error {
	// Get current job status
	var totalBatches, completedBatches, failedBatches int
	query := `SELECT total_batches, completed_batches, failed_batches FROM ai_schedule_batch_jobs WHERE id = ?`
	query = s.convertPlaceholders(query)

	err := s.db.QueryRowContext(ctx, query, jobID).Scan(&totalBatches, &completedBatches, &failedBatches)
	if err != nil {
		return err
	}

	processedBatches := completedBatches + failedBatches

	if processedBatches >= totalBatches {
		// All batches processed
		now := time.Now()
		var finalStatus models.AIScheduleBatchJobStatus

		if failedBatches == 0 {
			finalStatus = models.BatchStatusCompleted
		} else if completedBatches > 0 {
			finalStatus = models.BatchStatusPartial
		} else {
			finalStatus = models.BatchStatusFailed
		}

		progress := 100
		updateQuery := `UPDATE ai_schedule_batch_jobs
		                SET status = ?, progress = ?, completed_at = ?, modifiedon = ?,
		                    rowversionstamp = rowversionstamp + 1
		                WHERE id = ?`
		updateQuery = s.convertPlaceholders(updateQuery)

		_, err = s.db.ExecContext(ctx, updateQuery, finalStatus, progress, now, now, jobID)

		if err == nil {
			s.iLog.Info(fmt.Sprintf("Batch job %s finalized with status: %s", jobID, finalStatus))

			// Send final notification
			if status, err := s.GetBatchJobStatus(ctx, jobID); err == nil {
				job := &models.AIScheduleBatchJob{
					ID:               status.JobID,
					Status:           status.Status,
					Progress:         status.Progress,
					TotalBatches:     status.TotalBatches,
					CompletedBatches: status.CompletedBatches,
					FailedBatches:    status.FailedBatches,
					TotalTasks:       status.TotalTasks,
					ProcessedTasks:   status.ProcessedTasks,
					CompletedAt:      status.CompletedAt,
				}
				s.sendJobStatusNotification(job)
			}
		}

		return err
	}

	// Update progress
	progress := (processedBatches * 100) / totalBatches
	return s.updateBatchJobStatus(ctx, jobID, models.BatchStatusProcessing, progress)
}
