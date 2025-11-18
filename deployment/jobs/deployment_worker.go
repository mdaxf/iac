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
	"encoding/json"
	"fmt"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	deploymgr "github.com/mdaxf/iac/deployment/deploy"
	"github.com/mdaxf/iac/deployment/models"
	"github.com/mdaxf/iac/deployment/repository"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
)

// DeploymentWorker handles background deployment jobs
type DeploymentWorker struct {
	logger logger.Log
}

// NewDeploymentWorker creates a new deployment worker
func NewDeploymentWorker() *DeploymentWorker {
	return &DeploymentWorker{
		logger: logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DeploymentWorker"},
	}
}

// Start starts the deployment worker
func (dw *DeploymentWorker) Start(pollInterval time.Duration) {
	dw.logger.Info("Deployment worker started")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		dw.processPendingJobs()
	}
}

// processPendingJobs processes all pending deployment jobs
func (dw *DeploymentWorker) processPendingJobs() {
	jobs, err := dw.getPendingJobs()
	if err != nil {
		dw.logger.Error(fmt.Sprintf("Failed to get pending jobs: %v", err))
		return
	}

	for _, job := range jobs {
		dw.logger.Info(fmt.Sprintf("Processing deployment job: %s", job.ID))
		dw.processJob(job)
	}
}

// DeploymentJob represents a deployment job
type DeploymentJob struct {
	ID          string
	JobType     string
	JobData     string
	ScheduledAt time.Time
}

// getPendingJobs retrieves all pending jobs that are ready to execute
func (dw *DeploymentWorker) getPendingJobs() ([]DeploymentJob, error) {
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer dbTx.Rollback()

	dbOp := dbconn.NewDBOperation("System", dbTx, logger.Framework)

	query := fmt.Sprintf(`
		SELECT id, jobtype, jobdata, scheduledat
		FROM %s
		WHERE status = %s
		  AND scheduledat <= %s
		  AND active = %s
		  AND jobtype = %s
		ORDER BY scheduledat ASC
		LIMIT 10`,
		dbOp.QuoteIdentifier("backgroundjobs"),
		dbOp.GetPlaceholder(1),
		dbOp.GetPlaceholder(2),
		dbOp.GetPlaceholder(3),
		dbOp.GetPlaceholder(4))

	rows, err := dbOp.Query(query, "pending", time.Now(), true, "package_deployment")
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]DeploymentJob, 0)
	for rows.Next() {
		var job DeploymentJob
		if err := rows.Scan(&job.ID, &job.JobType, &job.JobData, &job.ScheduledAt); err != nil {
			dw.logger.Warn(fmt.Sprintf("Failed to scan job: %v", err))
			continue
		}
		jobs = append(jobs, job)
	}

	dbTx.Commit()

	return jobs, nil
}

// processJob processes a single deployment job
func (dw *DeploymentWorker) processJob(job DeploymentJob) {
	// Parse job data
	var jobData map[string]interface{}
	if err := json.Unmarshal([]byte(job.JobData), &jobData); err != nil {
		dw.logger.Error(fmt.Sprintf("Failed to parse job data: %v", err))
		dw.updateJobStatus(job.ID, "failed", fmt.Sprintf("Invalid job data: %v", err))
		return
	}

	packageID, ok := jobData["package_id"].(string)
	if !ok {
		dw.logger.Error("Missing package_id in job data")
		dw.updateJobStatus(job.ID, "failed", "Missing package_id")
		return
	}

	environment, _ := jobData["environment"].(string)
	userName, _ := jobData["user"].(string)
	if userName == "" {
		userName = "System"
	}

	// Update job status to running
	dw.updateJobStatus(job.ID, "running", "")
	startTime := time.Now()

	// Execute deployment
	err := dw.executeDeployment(packageID, environment, userName, jobData)
	if err != nil {
		dw.logger.Error(fmt.Sprintf("Deployment job failed: %v", err))
		dw.updateJobStatus(job.ID, "failed", err.Error())
		return
	}

	elapsed := time.Since(startTime)
	dw.logger.Info(fmt.Sprintf("Deployment job completed in %v: %s", elapsed, job.ID))
	dw.updateJobStatus(job.ID, "completed", "")
}

// executeDeployment executes the deployment
func (dw *DeploymentWorker) executeDeployment(packageID, environment, userName string, jobData map[string]interface{}) error {
	startTime := time.Now()

	// Begin database transaction
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer dbTx.Rollback()

	// Get package
	repo := repository.NewPackageRepository(userName, dbTx)
	pkg, err := repo.GetPackage(packageID)
	if err != nil {
		return fmt.Errorf("package not found: %w", err)
	}

	// Parse deployment options
	var options models.DeploymentOptions
	if optionsData, ok := jobData["options"]; ok {
		optionsJSON, _ := json.Marshal(optionsData)
		if err := json.Unmarshal(optionsJSON, &options); err != nil {
			dw.logger.Warn(fmt.Sprintf("Failed to parse deployment options: %v", err))
			// Use default options
			options = models.DeploymentOptions{
				SkipExisting:       true,
				ValidateReferences: true,
				BatchSize:          100,
			}
		}
	}

	// Create deployment action record
	deployAction := &repository.PackageActionRecord{
		PackageID:         packageID,
		ActionType:        repository.ActionTypeDeploy,
		ActionStatus:      repository.ActionStatusInProgress,
		TargetEnvironment: environment,
		PerformedBy:       userName,
		PerformedAt:       time.Now(),
		StartedAt:         &startTime,
	}

	if err := repo.SaveAction(deployAction); err != nil {
		dw.logger.Warn(fmt.Sprintf("Failed to save deploy action: %v", err))
	}

	// Deploy based on package type
	var deploymentRecord *models.DeploymentRecord

	if pkg.PackageType == "database" {
		deployer := deploymgr.NewDatabaseDeployer(userName, dbTx, dbconn.DatabaseType)
		deploymentRecord, err = deployer.Deploy(pkg, options)
	} else if pkg.PackageType == "document" {
		deployer := deploymgr.NewDocumentDeployer(documents.DocDBCon, userName)
		deploymentRecord, err = deployer.Deploy(pkg, options)
	} else {
		return fmt.Errorf("unsupported package type: %s", pkg.PackageType)
	}

	if err != nil {
		// Update action status to failed
		now := time.Now()
		deployAction.ActionStatus = repository.ActionStatusFailed
		deployAction.CompletedAt = &now
		errorLog := []string{err.Error()}
		repo.UpdateActionStatus(deployAction.ID, repository.ActionStatusFailed, &now, errorLog)

		return fmt.Errorf("deployment failed: %w", err)
	}

	// Update action status to completed
	now := time.Now()
	deployAction.ActionStatus = repository.ActionStatusCompleted
	deployAction.CompletedAt = &now

	resultDataJSON, _ := json.Marshal(deploymentRecord)
	deployAction.ResultData = string(resultDataJSON)

	repo.UpdateActionStatus(deployAction.ID, repository.ActionStatusCompleted, &now, nil)

	// Save deployment record
	deployment := &repository.PackageDeployment{
		PackageID:    packageID,
		ActionID:     deployAction.ID,
		Environment:  environment,
		DatabaseName: dbconn.DatabaseName,
		DeployedAt:   time.Now(),
		DeployedBy:   userName,
		IsActive:     true,
		CreatedBy:    userName,
	}

	if err := repo.SaveDeployment(deployment); err != nil {
		dw.logger.Warn(fmt.Sprintf("Failed to save deployment record: %v", err))
	}

	// Commit transaction
	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit deployment: %w", err)
	}

	dw.logger.Info(fmt.Sprintf("Package deployed successfully: %s", packageID))

	return nil
}

// updateJobStatus updates the status of a job
func (dw *DeploymentWorker) updateJobStatus(jobID, status, errorMsg string) {
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		dw.logger.Error(fmt.Sprintf("Failed to begin transaction: %v", err))
		return
	}
	defer dbTx.Rollback()

	dbOp := dbconn.NewDBOperation("System", dbTx, logger.Framework)

	now := time.Now()

	var query string
	var args []interface{}

	if status == "running" {
		query = fmt.Sprintf(`
			UPDATE %s
			SET status = %s, startedat = %s, modifiedon = %s
			WHERE id = %s`,
			dbOp.QuoteIdentifier("backgroundjobs"),
			dbOp.GetPlaceholder(1),
			dbOp.GetPlaceholder(2),
			dbOp.GetPlaceholder(3),
			dbOp.GetPlaceholder(4))
		args = []interface{}{status, now, now, jobID}

	} else if status == "completed" || status == "failed" {
		errorLogJSON := ""
		if errorMsg != "" {
			errorLog := []string{errorMsg}
			errorLogBytes, _ := json.Marshal(errorLog)
			errorLogJSON = string(errorLogBytes)
		}

		query = fmt.Sprintf(`
			UPDATE %s
			SET status = %s, completedat = %s, errorlog = %s, modifiedon = %s
			WHERE id = %s`,
			dbOp.QuoteIdentifier("backgroundjobs"),
			dbOp.GetPlaceholder(1),
			dbOp.GetPlaceholder(2),
			dbOp.GetPlaceholder(3),
			dbOp.GetPlaceholder(4),
			dbOp.GetPlaceholder(5))
		args = []interface{}{status, now, errorLogJSON, now, jobID}
	}

	_, err = dbOp.Exec(query, args...)
	if err != nil {
		dw.logger.Error(fmt.Sprintf("Failed to update job status: %v", err))
		return
	}

	dbTx.Commit()
}
