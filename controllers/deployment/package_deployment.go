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

package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	dbconn "github.com/mdaxf/iac/databases"
	deploymgr "github.com/mdaxf/iac/deployment/deploy"
	deploymodels "github.com/mdaxf/iac/deployment/models"
	"github.com/mdaxf/iac/deployment/repository"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// DeployPackage godoc
// @Summary Deploy a package
// @Description Deploy a package to the target environment
// @Tags packages
// @Accept json
// @Produce json
// @Param id path string true "Package ID"
// @Param request body DeployPackageRequest true "Deployment request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages/{id}/deploy [post]
func (pc *PackageController) DeployPackage(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.DeployPackage"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("PackageController.DeployPackage", elapsed)
	}()

	packageID := c.Param("id")

	// Parse request
	var req DeployPackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Get user from context
	user, _ := c.Get("user")
	userName := fmt.Sprintf("%v", user)
	if userName == "" {
		userName = "System"
	}

	// If scheduled or background job requested, create job entry
	if req.ScheduleAt != nil || req.RunAsBackgroundJob {
		jobID, err := pc.createDeploymentJob(packageID, userName, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create deployment job: %v", err)})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"message": "Deployment job created",
			"job_id":  jobID,
			"status":  "scheduled",
		})
		return
	}

	// Execute deployment immediately
	deploymentRecord, err := pc.executeDeployment(packageID, userName, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Deployment failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Package deployed successfully",
		"deployment": deploymentRecord,
	})
}

// executeDeployment performs the actual deployment
func (pc *PackageController) executeDeployment(packageID, userName string, req DeployPackageRequest) (*deploymodels.DeploymentRecord, error) {
	iLog := logger.Log{ModuleName: logger.API, User: userName, ControllerName: "PackageController.executeDeployment"}
	startTime := time.Now()

	// Begin database transaction
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer dbTx.Rollback()

	// Get package
	repo := repository.NewPackageRepository(userName, dbTx)
	pkg, err := repo.GetPackage(packageID)
	if err != nil {
		return nil, fmt.Errorf("package not found: %w", err)
	}

	// Create deployment action record
	deployAction := &repository.PackageActionRecord{
		PackageID:         packageID,
		ActionType:        repository.ActionTypeDeploy,
		ActionStatus:      repository.ActionStatusInProgress,
		TargetEnvironment: req.Environment,
		PerformedBy:       userName,
		PerformedAt:       time.Now(),
		StartedAt:         &startTime,
	}

	optionsJSON, _ := json.Marshal(req.Options)
	deployAction.Options = string(optionsJSON)

	if err := repo.SaveAction(deployAction); err != nil {
		iLog.Warn(fmt.Sprintf("Failed to save deploy action: %v", err))
	}

	// Deploy based on package type
	var deploymentRecord *deploymodels.DeploymentRecord

	if pkg.PackageType == "database" {
		deployer := deploymgr.NewDatabaseDeployer(userName, dbTx, dbconn.DatabaseType)
		deploymentRecord, err = deployer.Deploy(pkg, req.Options)
	} else if pkg.PackageType == "document" {
		deployer := deploymgr.NewDocumentDeployer(documents.DocDBCon, userName)
		deploymentRecord, err = deployer.Deploy(pkg, req.Options)
	} else {
		return nil, fmt.Errorf("unsupported package type: %s", pkg.PackageType)
	}

	if err != nil {
		// Update action status to failed
		now := time.Now()
		deployAction.ActionStatus = repository.ActionStatusFailed
		deployAction.CompletedAt = &now
		errorLog := []string{err.Error()}
		errorLogJSON, _ := json.Marshal(errorLog)
		deployAction.ErrorLog = string(errorLogJSON)
		repo.UpdateActionStatus(deployAction.ID, repository.ActionStatusFailed, &now, errorLog)

		return nil, fmt.Errorf("deployment failed: %w", err)
	}

	// Update action status to completed
	now := time.Now()
	deployAction.ActionStatus = repository.ActionStatusCompleted
	deployAction.CompletedAt = &now

	resultDataJSON, _ := json.Marshal(deploymentRecord)
	deployAction.ResultData = string(resultDataJSON)
	deployAction.RecordsProcessed = len(deploymentRecord.PKMappingResult) + len(deploymentRecord.IDMappingResult)

	repo.UpdateActionStatus(deployAction.ID, repository.ActionStatusCompleted, &now, nil)

	// Save deployment record
	deployment := &repository.PackageDeployment{
		PackageID:    packageID,
		ActionID:     deployAction.ID,
		Environment:  req.Environment,
		DatabaseName: dbconn.DatabaseName,
		DeployedAt:   time.Now(),
		DeployedBy:   userName,
		IsActive:     true,
		CreatedBy:    userName,
	}

	if err := repo.SaveDeployment(deployment); err != nil {
		iLog.Warn(fmt.Sprintf("Failed to save deployment record: %v", err))
	}

	// Commit transaction
	if err := dbTx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit deployment: %w", err)
	}

	iLog.Info(fmt.Sprintf("Package deployed successfully: %s", packageID))

	return deploymentRecord, nil
}

// createDeploymentJob creates a background job for deployment using the existing job framework
func (pc *PackageController) createDeploymentJob(packageID, userName string, req DeployPackageRequest) (string, error) {
	// Create job service
	jobService := services.NewJobService(dbconn.DB)

	// Prepare job payload
	jobData := map[string]interface{}{
		"package_id":  packageID,
		"environment": req.Environment,
		"options":     req.Options,
		"user":        userName,
	}
	payloadJSON, _ := json.Marshal(jobData)

	// Prepare job metadata
	metadata := models.JobMetadata{
		"package_id":     packageID,
		"deployment_env": req.Environment,
		"created_by":     userName,
	}

	// Create queue job
	job := &models.QueueJob{
		TypeID:      int(models.JobTypeManual),
		Method:      "POST",
		Protocol:    "internal",
		Direction:   models.JobDirectionInternal,
		Handler:     "PACKAGE_DEPLOYMENT", // This will be handled by the job worker
		Metadata:    metadata,
		Payload:     string(payloadJSON),
		Priority:    5, // Medium priority
		MaxRetries:  3,
		RetryCount:  0,
		ScheduledAt: req.ScheduleAt,
		CreatedBy:   userName,
	}

	// Create the job
	ctx := context.Background()
	if err := jobService.CreateQueueJob(ctx, job); err != nil {
		return "", fmt.Errorf("failed to create deployment job: %w", err)
	}

	return job.ID, nil
}

// GetPackageActions godoc
// @Summary Get package action history
// @Description Get all actions (pack, deploy, rollback) for a package
// @Tags packages
// @Accept json
// @Produce json
// @Param id path string true "Package ID"
// @Param limit query int false "Limit results" default(50)
// @Success 200 {array} repository.PackageActionRecord
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages/{id}/actions [get]
func (pc *PackageController) GetPackageActions(c *gin.Context) {
	packageID := c.Param("id")
	limit := c.GetInt("limit")
	if limit == 0 {
		limit = 50
	}

	// Get user from context
	user, _ := c.Get("user")
	userName := fmt.Sprintf("%v", user)
	if userName == "" {
		userName = "System"
	}

	// Begin database transaction
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to begin transaction: %v", err)})
		return
	}
	defer dbTx.Rollback()

	// Get actions
	repo := repository.NewPackageRepository(userName, dbTx)
	actions, err := repo.GetActionsByPackage(packageID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get actions: %v", err)})
		return
	}

	dbTx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"actions": actions,
		"count":   len(actions),
	})
}

// ListDeployments godoc
// @Summary List all deployments
// @Description Get a list of all package deployments with optional filters
// @Tags deployments
// @Accept json
// @Produce json
// @Param environment query string false "Environment filter"
// @Param is_active query bool false "Active status filter"
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} repository.PackageDeployment
// @Failure 500 {object} map[string]interface{}
// @Router /api/deployments [get]
func (pc *PackageController) ListDeployments(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.ListDeployments"}

	// Get query parameters
	environment := c.Query("environment")
	isActiveStr := c.Query("is_active")
	limit := c.GetInt("limit")
	offset := c.GetInt("offset")

	if limit == 0 {
		limit = 50
	}

	// Get user from context
	user, _ := c.Get("user")
	userName := fmt.Sprintf("%v", user)
	if userName == "" {
		userName = "System"
	}

	// Begin database transaction
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to begin transaction: %v", err)})
		return
	}
	defer dbTx.Rollback()

	// Build query
	dbOp := dbconn.NewDBOperation(userName, dbTx, logger.Framework)

	query := fmt.Sprintf(`
		SELECT id, packageid, actionid, environment, databasename, deployedat, deployedby,
		       isactive, rolledbackat, rolledbackby,
		       active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM %s
		WHERE active = %s`,
		dbOp.QuoteIdentifier("packagedeployments"),
		dbOp.GetPlaceholder(1))

	args := []interface{}{true}
	paramIdx := 2

	if environment != "" {
		query += fmt.Sprintf(" AND environment = %s", dbOp.GetPlaceholder(paramIdx))
		args = append(args, environment)
		paramIdx++
	}

	if isActiveStr != "" {
		isActive := isActiveStr == "true"
		query += fmt.Sprintf(" AND isactive = %s", dbOp.GetPlaceholder(paramIdx))
		args = append(args, isActive)
		paramIdx++
	}

	query += " ORDER BY deployedat DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %s", dbOp.GetPlaceholder(paramIdx))
		args = append(args, limit)
		paramIdx++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %s", dbOp.GetPlaceholder(paramIdx))
		args = append(args, offset)
	}

	rows, err := dbOp.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to query deployments: %v", err)})
		return
	}
	defer rows.Close()

	deployments := make([]repository.PackageDeployment, 0)
	for rows.Next() {
		var deployment repository.PackageDeployment
		err := rows.Scan(
			&deployment.ID,
			&deployment.PackageID,
			&deployment.ActionID,
			&deployment.Environment,
			&deployment.DatabaseName,
			&deployment.DeployedAt,
			&deployment.DeployedBy,
			&deployment.IsActive,
			&deployment.RolledBackAt,
			&deployment.RolledBackBy,
			&deployment.Active,
			&deployment.ReferenceID,
			&deployment.CreatedBy,
			&deployment.CreatedOn,
			&deployment.ModifiedBy,
			&deployment.ModifiedOn,
			&deployment.RowVersionStamp,
		)
		if err != nil {
			iLog.Warn(fmt.Sprintf("Error scanning deployment: %v", err))
			continue
		}
		deployments = append(deployments, deployment)
	}

	dbTx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"deployments": deployments,
		"count":       len(deployments),
		"limit":       limit,
		"offset":      offset,
	})
}

// GetDeployment godoc
// @Summary Get deployment details
// @Description Get detailed information about a specific deployment
// @Tags deployments
// @Accept json
// @Produce json
// @Param id path string true "Deployment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/deployments/{id} [get]
func (pc *PackageController) GetDeployment(c *gin.Context) {
	deploymentID := c.Param("id")

	// Get user from context
	user, _ := c.Get("user")
	userName := fmt.Sprintf("%v", user)
	if userName == "" {
		userName = "System"
	}

	// Begin database transaction
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to begin transaction: %v", err)})
		return
	}
	defer dbTx.Rollback()

	// Build query
	dbOp := dbconn.NewDBOperation(userName, dbTx, logger.Framework)

	query := fmt.Sprintf(`
		SELECT id, packageid, actionid, environment, databasename, deployedat, deployedby,
		       isactive, rolledbackat, rolledbackby,
		       active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM %s
		WHERE id = %s AND active = %s`,
		dbOp.QuoteIdentifier("packagedeployments"),
		dbOp.GetPlaceholder(1),
		dbOp.GetPlaceholder(2))

	rows, err := dbOp.Query(query, deploymentID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to query deployment: %v", err)})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}

	var deployment repository.PackageDeployment
	err = rows.Scan(
		&deployment.ID,
		&deployment.PackageID,
		&deployment.ActionID,
		&deployment.Environment,
		&deployment.DatabaseName,
		&deployment.DeployedAt,
		&deployment.DeployedBy,
		&deployment.IsActive,
		&deployment.RolledBackAt,
		&deployment.RolledBackBy,
		&deployment.Active,
		&deployment.ReferenceID,
		&deployment.CreatedBy,
		&deployment.CreatedOn,
		&deployment.ModifiedBy,
		&deployment.ModifiedOn,
		&deployment.RowVersionStamp,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to scan deployment: %v", err)})
		return
	}

	dbTx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"deployment": deployment,
	})
}

// GetJobStatus godoc
// @Summary Get background job status
// @Description Get status and details of a background deployment job
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/jobs/{id} [get]
func (pc *PackageController) GetJobStatus(c *gin.Context) {
	jobID := c.Param("id")

	// Get user from context
	user, _ := c.Get("user")
	userName := fmt.Sprintf("%v", user)
	if userName == "" {
		userName = "System"
	}

	// Create job service
	jobService := services.NewJobService(dbconn.DB)

	// Get job details
	ctx := context.Background()
	job, err := jobService.GetQueueJob(ctx, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Job not found: %v", err)})
		return
	}

	// Map job status to human-readable string
	statusMap := map[int]string{
		int(models.JobStatusPending):    "pending",
		int(models.JobStatusQueued):     "queued",
		int(models.JobStatusProcessing): "processing",
		int(models.JobStatusCompleted):  "completed",
		int(models.JobStatusFailed):     "failed",
		int(models.JobStatusRetrying):   "retrying",
		int(models.JobStatusCancelled):  "cancelled",
		int(models.JobStatusScheduled):  "scheduled",
	}

	response := gin.H{
		"job_id":       job.ID,
		"handler":      job.Handler,
		"status":       statusMap[job.StatusID],
		"status_id":    job.StatusID,
		"priority":     job.Priority,
		"scheduled_at": job.ScheduledAt,
		"started_at":   job.StartedAt,
		"completed_at": job.CompletedAt,
		"retry_count":  job.RetryCount,
		"max_retries":  job.MaxRetries,
		"created_by":   job.CreatedBy,
		"created_on":   job.CreatedOn,
	}

	if job.LastError != "" {
		response["last_error"] = job.LastError
	}

	if job.Result != "" {
		response["result"] = job.Result
	}

	// Parse payload
	var payloadData map[string]interface{}
	if err := json.Unmarshal([]byte(job.Payload), &payloadData); err == nil {
		response["job_data"] = payloadData
	}

	// Include metadata
	if job.Metadata != nil && len(job.Metadata) > 0 {
		response["metadata"] = job.Metadata
	}

	c.JSON(http.StatusOK, response)
}
