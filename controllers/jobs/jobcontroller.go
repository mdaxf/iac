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

package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdaxf/iac/controllers/common"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/framework/jobqueue"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// JobController handles job management API endpoints
type JobController struct {
	jobService *services.JobService
	iLog       logger.Log
}

// NewJobController creates a new job controller instance
func NewJobController() *JobController {
	return &JobController{
		jobService: services.NewJobService(dbconn.DB),
		iLog:       logger.Log{ModuleName: logger.API, User: "System", ControllerName: "JobController"},
	}
}

// convertPlaceholders converts ? placeholders to $1, $2, etc. for PostgreSQL
func convertPlaceholders(query string) string {
	dbType := strings.ToLower(dbconn.DatabaseType)
	if dbType == "postgres" || dbType == "postgresql" {
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

// ============== Queue Jobs API ==============

// ListQueueJobs returns a paginated list of queue jobs
// GET /jobs/queue
func (jc *JobController) ListQueueJobs(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	status := c.Query("status")
	direction := c.Query("direction")
	handler := c.Query("handler")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Build query
	query := `
		SELECT id, typeid, method, protocol, direction, handler, metadata, payload,
		       result, statusid, priority, maxretries, retrycount, scheduledat,
		       startedat, completedat, lasterror, parentjobid, active, referenceid,
		       createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM queue_jobs
		WHERE active = ?
	`
	countQuery := `SELECT COUNT(*) FROM queue_jobs WHERE active = ?`
	args := []interface{}{true}

	if status != "" {
		statusID, _ := strconv.Atoi(status)
		query += ` AND statusid = ?`
		countQuery += ` AND statusid = ?`
		args = append(args, statusID)
	}

	if direction != "" {
		query += ` AND direction = ?`
		countQuery += ` AND direction = ?`
		args = append(args, direction)
	}

	if handler != "" {
		query += ` AND handler LIKE ?`
		countQuery += ` AND handler LIKE ?`
		args = append(args, "%"+handler+"%")
	}

	query += ` ORDER BY createdon DESC LIMIT ? OFFSET ?`

	query = convertPlaceholders(query)
	countQuery = convertPlaceholders(countQuery)

	// Get total count
	var total int
	err = dbconn.DB.QueryRowContext(c, countQuery, args...).Scan(&total)
	if err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to count queue jobs: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count jobs"})
		return
	}

	// Get jobs
	args = append(args, pageSize, offset)
	rows, err := dbconn.DB.QueryContext(c, query, args...)
	if err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to list queue jobs: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list jobs"})
		return
	}
	defer rows.Close()

	var jobs []models.QueueJob
	for rows.Next() {
		job := models.QueueJob{}
		var metadataJSON string
		err := rows.Scan(
			&job.ID, &job.TypeID, &job.Method, &job.Protocol, &job.Direction, &job.Handler, &metadataJSON, &job.Payload,
			&job.Result, &job.StatusID, &job.Priority, &job.MaxRetries, &job.RetryCount, &job.ScheduledAt,
			&job.StartedAt, &job.CompletedAt, &job.LastError, &job.ParentJobID, &job.Active, &job.ReferenceID,
			&job.CreatedBy, &job.CreatedOn, &job.ModifiedBy, &job.ModifiedOn, &job.RowVersionStamp,
		)
		if err != nil {
			jc.iLog.Error(fmt.Sprintf("Failed to scan job: %v", err))
			continue
		}
		if metadataJSON != "" {
			json.Unmarshal([]byte(metadataJSON), &job.Metadata)
		}
		jobs = append(jobs, job)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     jobs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetQueueJob returns a single queue job by ID
// GET /jobs/queue/:id
func (jc *JobController) GetQueueJob(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	job, err := jc.jobService.GetJobByID(c, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": job})
}

// CreateQueueJob creates a new queue job
// POST /jobs/queue
func (jc *JobController) CreateQueueJob(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	var job models.QueueJob
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job.CreatedBy = user
	job.ModifiedBy = user

	if err := jc.jobService.CreateQueueJob(c, &job); err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to create queue job: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job"})
		return
	}

	// Enqueue the job if queue manager is available
	if jobqueue.GlobalQueueManager != nil {
		jobqueue.GlobalQueueManager.EnqueueJob(c, job.ID, job.Priority)
	}

	c.JSON(http.StatusCreated, gin.H{"data": job})
}

// UpdateQueueJobStatus updates a queue job's status
// PUT /jobs/queue/:id/status
func (jc *JobController) UpdateQueueJobStatus(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	var request struct {
		StatusID int    `json:"statusid"`
		Result   string `json:"result"`
		Error    string `json:"error"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := jc.jobService.UpdateQueueJobStatus(c, jobID, request.StatusID, request.Result, request.Error); err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to update job status: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job status updated"})
}

// CancelQueueJob cancels a pending or scheduled job
// POST /jobs/queue/:id/cancel
func (jc *JobController) CancelQueueJob(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	// Check if job exists and is cancellable
	job, err := jc.jobService.GetJobByID(c, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	if job.StatusID != int(models.JobStatusPending) && job.StatusID != int(models.JobStatusScheduled) && job.StatusID != int(models.JobStatusQueued) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending, queued, or scheduled jobs can be cancelled"})
		return
	}

	if err := jc.jobService.UpdateQueueJobStatus(c, jobID, int(models.JobStatusCancelled), "", "Cancelled by "+user); err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to cancel job: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel job"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job cancelled"})
}

// RetryQueueJob retries a failed job
// POST /jobs/queue/:id/retry
func (jc *JobController) RetryQueueJob(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	// Check if job exists and is retryable
	job, err := jc.jobService.GetJobByID(c, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	if job.StatusID != int(models.JobStatusFailed) && job.StatusID != int(models.JobStatusCancelled) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only failed or cancelled jobs can be retried"})
		return
	}

	// Reset job status to pending
	if err := jc.jobService.UpdateQueueJobStatus(c, jobID, int(models.JobStatusPending), "", ""); err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to retry job: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retry job"})
		return
	}

	// Enqueue the job
	if jobqueue.GlobalQueueManager != nil {
		jobqueue.GlobalQueueManager.EnqueueJob(c, job.ID, job.Priority)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job queued for retry"})
}

// ============== Scheduled Jobs API ==============

// ListScheduledJobs returns a list of scheduled job definitions
// GET /jobs/scheduled
func (jc *JobController) ListScheduledJobs(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	// Parse query parameters
	enabledOnly := c.Query("enabledOnly") == "true"

	// Build query with proper identifier quoting for different databases
	conditionColumn := "condition"
	dbType := strings.ToLower(dbconn.DatabaseType)
	if dbType == "mysql" {
		conditionColumn = "`condition`"
	} else if dbType == "postgres" || dbType == "postgresql" {
		conditionColumn = "\"condition\""
	}

	query := fmt.Sprintf(`
		SELECT id, name, description, typeid, handler, cronexpression, intervalseconds,
		       startat, endat, maxexecutions, executioncount, enabled, %s,
		       priority, maxretries, timeout, metadata, lastrunat, nextrunat,
		       active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM jobs
		WHERE active = ?
	`, conditionColumn)

	args := []interface{}{true}

	if enabledOnly {
		query += ` AND enabled = ?`
		args = append(args, true)
	}

	query += ` ORDER BY name ASC`
	query = convertPlaceholders(query)

	rows, err := dbconn.DB.QueryContext(c, query, args...)
	if err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to list scheduled jobs: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list scheduled jobs"})
		return
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		job := models.Job{}
		var metadataJSON sql.NullString
		err := rows.Scan(
			&job.ID, &job.Name, &job.Description, &job.TypeID, &job.Handler, &job.CronExpression, &job.IntervalSeconds,
			&job.StartAt, &job.EndAt, &job.MaxExecutions, &job.ExecutionCount, &job.Enabled, &job.Condition,
			&job.Priority, &job.MaxRetries, &job.Timeout, &metadataJSON, &job.LastRunAt, &job.NextRunAt,
			&job.Active, &job.ReferenceID, &job.CreatedBy, &job.CreatedOn, &job.ModifiedBy, &job.ModifiedOn, &job.RowVersionStamp,
		)
		if err != nil {
			jc.iLog.Error(fmt.Sprintf("Failed to scan scheduled job: %v", err))
			continue
		}
		if metadataJSON.Valid && metadataJSON.String != "" {
			json.Unmarshal([]byte(metadataJSON.String), &job.Metadata)
		}
		jobs = append(jobs, job)
	}

	c.JSON(http.StatusOK, gin.H{"data": jobs})
}

// GetScheduledJob returns a single scheduled job by ID
// GET /jobs/scheduled/:id
func (jc *JobController) GetScheduledJob(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	// Build query with proper identifier quoting
	conditionColumn := "condition"
	dbType := strings.ToLower(dbconn.DatabaseType)
	if dbType == "mysql" {
		conditionColumn = "`condition`"
	} else if dbType == "postgres" || dbType == "postgresql" {
		conditionColumn = "\"condition\""
	}

	query := fmt.Sprintf(`
		SELECT id, name, description, typeid, handler, cronexpression, intervalseconds,
		       startat, endat, maxexecutions, executioncount, enabled, %s,
		       priority, maxretries, timeout, metadata, lastrunat, nextrunat,
		       active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM jobs
		WHERE id = ?
	`, conditionColumn)
	query = convertPlaceholders(query)

	job := models.Job{}
	var metadataJSON sql.NullString
	err = dbconn.DB.QueryRowContext(c, query, jobID).Scan(
		&job.ID, &job.Name, &job.Description, &job.TypeID, &job.Handler, &job.CronExpression, &job.IntervalSeconds,
		&job.StartAt, &job.EndAt, &job.MaxExecutions, &job.ExecutionCount, &job.Enabled, &job.Condition,
		&job.Priority, &job.MaxRetries, &job.Timeout, &metadataJSON, &job.LastRunAt, &job.NextRunAt,
		&job.Active, &job.ReferenceID, &job.CreatedBy, &job.CreatedOn, &job.ModifiedBy, &job.ModifiedOn, &job.RowVersionStamp,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scheduled job not found"})
		return
	}
	if err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to get scheduled job: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scheduled job"})
		return
	}

	if metadataJSON.Valid && metadataJSON.String != "" {
		json.Unmarshal([]byte(metadataJSON.String), &job.Metadata)
	}

	c.JSON(http.StatusOK, gin.H{"data": job})
}

// CreateScheduledJob creates a new scheduled job definition
// POST /jobs/scheduled
func (jc *JobController) CreateScheduledJob(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	var job models.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if job.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job name is required"})
		return
	}
	if job.Handler == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job handler is required"})
		return
	}
	if job.CronExpression == nil && job.IntervalSeconds == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either cron expression or interval is required"})
		return
	}

	job.ID = uuid.New().String()
	job.CreatedBy = &user
	job.ModifiedBy = &user
	job.CreatedOn = time.Now()
	job.ModifiedOn = time.Now()
	job.Active = true
	job.RowVersionStamp = 1

	if job.TypeID == 0 {
		job.TypeID = int(models.JobTypeScheduled)
	}

	metadataJSON, _ := json.Marshal(job.Metadata)

	// Build query with proper identifier quoting
	conditionColumn := "condition"
	dbType := strings.ToLower(dbconn.DatabaseType)
	if dbType == "mysql" {
		conditionColumn = "`condition`"
	} else if dbType == "postgres" || dbType == "postgresql" {
		conditionColumn = "\"condition\""
	}

	query := fmt.Sprintf(`
		INSERT INTO jobs (
			id, name, description, typeid, handler, cronexpression, intervalseconds,
			startat, endat, maxexecutions, executioncount, enabled, %s,
			priority, maxretries, timeout, metadata, lastrunat, nextrunat,
			active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, conditionColumn)
	query = convertPlaceholders(query)

	_, err = dbconn.DB.ExecContext(c, query,
		job.ID, job.Name, job.Description, job.TypeID, job.Handler, job.CronExpression, job.IntervalSeconds,
		job.StartAt, job.EndAt, job.MaxExecutions, job.ExecutionCount, job.Enabled, job.Condition,
		job.Priority, job.MaxRetries, job.Timeout, string(metadataJSON), job.LastRunAt, job.NextRunAt,
		job.Active, job.ReferenceID, job.CreatedBy, job.CreatedOn, job.ModifiedBy, job.ModifiedOn, job.RowVersionStamp,
	)

	if err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to create scheduled job: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scheduled job"})
		return
	}

	// Add to scheduler if enabled
	if job.Enabled && jobqueue.GlobalJobScheduler != nil {
		jobqueue.GlobalJobScheduler.AddJob(&job)
	}

	c.JSON(http.StatusCreated, gin.H{"data": job})
}

// UpdateScheduledJob updates a scheduled job definition
// PUT /jobs/scheduled/:id
func (jc *JobController) UpdateScheduledJob(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	var job models.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job.ID = jobID
	job.ModifiedBy = &user
	job.ModifiedOn = time.Now()

	metadataJSON, _ := json.Marshal(job.Metadata)

	// Build query with proper identifier quoting
	conditionColumn := "condition"
	dbType := strings.ToLower(dbconn.DatabaseType)
	if dbType == "mysql" {
		conditionColumn = "`condition`"
	} else if dbType == "postgres" || dbType == "postgresql" {
		conditionColumn = "\"condition\""
	}

	query := fmt.Sprintf(`
		UPDATE jobs SET
			name = ?, description = ?, typeid = ?, handler = ?, cronexpression = ?, intervalseconds = ?,
			startat = ?, endat = ?, maxexecutions = ?, enabled = ?, %s = ?,
			priority = ?, maxretries = ?, timeout = ?, metadata = ?,
			modifiedby = ?, modifiedon = ?, rowversionstamp = rowversionstamp + 1
		WHERE id = ?
	`, conditionColumn)
	query = convertPlaceholders(query)

	_, err = dbconn.DB.ExecContext(c, query,
		job.Name, job.Description, job.TypeID, job.Handler, job.CronExpression, job.IntervalSeconds,
		job.StartAt, job.EndAt, job.MaxExecutions, job.Enabled, job.Condition,
		job.Priority, job.MaxRetries, job.Timeout, string(metadataJSON),
		job.ModifiedBy, job.ModifiedOn, job.ID,
	)

	if err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to update scheduled job: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update scheduled job"})
		return
	}

	// Update scheduler
	if jobqueue.GlobalJobScheduler != nil {
		jobqueue.GlobalJobScheduler.RemoveJob(jobID)
		if job.Enabled {
			jobqueue.GlobalJobScheduler.AddJob(&job)
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": job})
}

// DeleteScheduledJob deletes a scheduled job (soft delete)
// DELETE /jobs/scheduled/:id
func (jc *JobController) DeleteScheduledJob(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	query := `UPDATE jobs SET active = ?, modifiedby = ?, modifiedon = ? WHERE id = ?`
	query = convertPlaceholders(query)

	_, err = dbconn.DB.ExecContext(c, query, false, user, time.Now(), jobID)
	if err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to delete scheduled job: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete scheduled job"})
		return
	}

	// Remove from scheduler
	if jobqueue.GlobalJobScheduler != nil {
		jobqueue.GlobalJobScheduler.RemoveJob(jobID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled job deleted"})
}

// ToggleScheduledJob enables or disables a scheduled job
// POST /jobs/scheduled/:id/toggle
func (jc *JobController) ToggleScheduledJob(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	var request struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `UPDATE jobs SET enabled = ?, modifiedby = ?, modifiedon = ? WHERE id = ?`
	query = convertPlaceholders(query)

	_, err = dbconn.DB.ExecContext(c, query, request.Enabled, user, time.Now(), jobID)
	if err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to toggle scheduled job: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle scheduled job"})
		return
	}

	// Update scheduler
	if jobqueue.GlobalJobScheduler != nil {
		jobqueue.GlobalJobScheduler.RemoveJob(jobID)
		if request.Enabled {
			// Re-fetch job to add to scheduler
			// Note: In a real implementation, you'd fetch and add the job here
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled job toggled", "enabled": request.Enabled})
}

// RunScheduledJobNow triggers immediate execution of a scheduled job
// POST /jobs/scheduled/:id/run
func (jc *JobController) RunScheduledJobNow(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	// Get the scheduled job
	conditionColumn := "condition"
	dbType := strings.ToLower(dbconn.DatabaseType)
	if dbType == "mysql" {
		conditionColumn = "`condition`"
	} else if dbType == "postgres" || dbType == "postgresql" {
		conditionColumn = "\"condition\""
	}

	query := fmt.Sprintf(`
		SELECT id, name, description, typeid, handler, cronexpression, intervalseconds,
		       startat, endat, maxexecutions, executioncount, enabled, %s,
		       priority, maxretries, timeout, metadata, lastrunat, nextrunat,
		       active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM jobs
		WHERE id = ? AND active = ?
	`, conditionColumn)
	query = convertPlaceholders(query)

	scheduledJob := models.Job{}
	var metadataJSON sql.NullString
	err = dbconn.DB.QueryRowContext(c, query, jobID, true).Scan(
		&scheduledJob.ID, &scheduledJob.Name, &scheduledJob.Description, &scheduledJob.TypeID, &scheduledJob.Handler,
		&scheduledJob.CronExpression, &scheduledJob.IntervalSeconds, &scheduledJob.StartAt, &scheduledJob.EndAt,
		&scheduledJob.MaxExecutions, &scheduledJob.ExecutionCount, &scheduledJob.Enabled, &scheduledJob.Condition,
		&scheduledJob.Priority, &scheduledJob.MaxRetries, &scheduledJob.Timeout, &metadataJSON,
		&scheduledJob.LastRunAt, &scheduledJob.NextRunAt, &scheduledJob.Active, &scheduledJob.ReferenceID,
		&scheduledJob.CreatedBy, &scheduledJob.CreatedOn, &scheduledJob.ModifiedBy, &scheduledJob.ModifiedOn, &scheduledJob.RowVersionStamp,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scheduled job not found"})
		return
	}
	if err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to get scheduled job: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scheduled job"})
		return
	}

	// Create a queue job from the scheduled job
	queueJob := &models.QueueJob{
		TypeID:     int(models.JobTypeManual),
		Handler:    scheduledJob.Handler,
		Priority:   scheduledJob.Priority,
		MaxRetries: scheduledJob.MaxRetries,
		CreatedBy:  user,
		Metadata: models.JobMetadata{
			"scheduled_job_id":   scheduledJob.ID,
			"scheduled_job_name": scheduledJob.Name,
			"triggered_by":       user,
			"trigger_type":       "manual",
		},
	}

	if metadataJSON.Valid && metadataJSON.String != "" {
		var metadata models.JobMetadata
		json.Unmarshal([]byte(metadataJSON.String), &metadata)
		for k, v := range metadata {
			queueJob.Metadata[k] = v
		}
	}

	if err := jc.jobService.CreateQueueJob(c, queueJob); err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to create queue job: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create queue job"})
		return
	}

	// Enqueue the job
	if jobqueue.GlobalQueueManager != nil {
		jobqueue.GlobalQueueManager.EnqueueJob(c, queueJob.ID, queueJob.Priority)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job queued for immediate execution", "queueJobId": queueJob.ID})
}

// ============== Job History API ==============

// ListJobHistory returns job execution history
// GET /jobs/history
func (jc *JobController) ListJobHistory(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	jobID := c.Query("jobId")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := `
		SELECT id, jobid, executionid, statusid, startedat, completedat,
		       duration, result, errormessage, retryattempt, executedby,
		       inputdata, outputdata, metadata, active, referenceid,
		       createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM job_histories
		WHERE active = ?
	`
	countQuery := `SELECT COUNT(*) FROM job_histories WHERE active = ?`
	args := []interface{}{true}

	if jobID != "" {
		query += ` AND jobid = ?`
		countQuery += ` AND jobid = ?`
		args = append(args, jobID)
	}

	query += ` ORDER BY startedat DESC LIMIT ? OFFSET ?`

	query = convertPlaceholders(query)
	countQuery = convertPlaceholders(countQuery)

	// Get total count
	var total int
	err = dbconn.DB.QueryRowContext(c, countQuery, args...).Scan(&total)
	if err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to count job history: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count job history"})
		return
	}

	// Get history records
	args = append(args, pageSize, offset)
	rows, err := dbconn.DB.QueryContext(c, query, args...)
	if err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to list job history: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list job history"})
		return
	}
	defer rows.Close()

	var histories []models.JobHistory
	for rows.Next() {
		history := models.JobHistory{}
		var metadataJSON sql.NullString
		err := rows.Scan(
			&history.ID, &history.JobID, &history.ExecutionID, &history.StatusID, &history.StartedAt, &history.CompletedAt,
			&history.Duration, &history.Result, &history.ErrorMessage, &history.RetryAttempt, &history.ExecutedBy,
			&history.InputData, &history.OutputData, &metadataJSON, &history.Active, &history.ReferenceID,
			&history.CreatedBy, &history.CreatedOn, &history.ModifiedBy, &history.ModifiedOn, &history.RowVersionStamp,
		)
		if err != nil {
			jc.iLog.Error(fmt.Sprintf("Failed to scan job history: %v", err))
			continue
		}
		if metadataJSON.Valid && metadataJSON.String != "" {
			json.Unmarshal([]byte(metadataJSON.String), &history.Metadata)
		}
		histories = append(histories, history)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     histories,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// ============== Statistics API ==============

// GetJobStatistics returns job system statistics
// GET /jobs/statistics
func (jc *JobController) GetJobStatistics(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	ctx := context.Background()
	stats, err := jc.jobService.GetJobStatistics(ctx)
	if err != nil {
		jc.iLog.Error(fmt.Sprintf("Failed to get job statistics: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get job statistics"})
		return
	}

	// Add worker status if available
	if jobqueue.GlobalJobWorker != nil {
		stats["worker"] = jobqueue.GlobalJobWorker.GetStatus()
	}

	// Add scheduler status if available
	if jobqueue.GlobalJobScheduler != nil {
		stats["scheduler"] = jobqueue.GlobalJobScheduler.GetStatus()
	}

	// Count scheduled jobs
	query := `SELECT COUNT(*) FROM jobs WHERE active = ? AND enabled = ?`
	query = convertPlaceholders(query)
	var scheduledCount int
	dbconn.DB.QueryRowContext(c, query, true, true).Scan(&scheduledCount)
	stats["scheduled_jobs_count"] = scheduledCount

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetExecutorInstances returns available job executor instances
// GET /jobs/executors
func (jc *JobController) GetExecutorInstances(c *gin.Context) {
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jc.iLog.ClientID = clientid
	jc.iLog.User = user

	// Return current instance info and any known instances
	instances := []map[string]interface{}{}

	if jobqueue.GlobalQueueManager != nil {
		instances = append(instances, map[string]interface{}{
			"id":       jobqueue.GlobalQueueManager.GetInstanceID(),
			"type":     "local",
			"status":   "running",
			"isCurrent": true,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": instances})
}
