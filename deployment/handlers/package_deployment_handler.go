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

package handlers

import (
	"database/sql"
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

// PackageDeploymentHandler handles background deployment jobs
// This function is called by the job worker when processing PACKAGE_DEPLOYMENT jobs
//
// Input data should contain:
//   - package_id: string - ID of the package to deploy
//   - environment: string - Target environment
//   - options: DeploymentOptions - Deployment configuration
//   - user: string - User who initiated the deployment
//
// Returns deployment result as JSON
func PackageDeploymentHandler(inputs map[string]interface{}, tx *sql.Tx, docDB *documents.DocDB) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "PackageDeploymentHandler"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("PackageDeploymentHandler", elapsed)
	}()

	// Extract parameters
	packageID, ok := inputs["package_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid package_id")
	}

	environment, _ := inputs["environment"].(string)
	if environment == "" {
		environment = "production"
	}

	userName, _ := inputs["user"].(string)
	if userName == "" {
		userName = "System"
	}

	// Parse deployment options
	var options models.DeploymentOptions
	if optionsData, ok := inputs["options"]; ok {
		// Convert options to JSON and back to struct
		optionsJSON, err := json.Marshal(optionsData)
		if err == nil {
			if err := json.Unmarshal(optionsJSON, &options); err != nil {
				iLog.Warn(fmt.Sprintf("Failed to parse deployment options: %v, using defaults", err))
				options = getDefaultDeploymentOptions()
			}
		} else {
			options = getDefaultDeploymentOptions()
		}
	} else {
		options = getDefaultDeploymentOptions()
	}

	iLog.Info(fmt.Sprintf("Starting package deployment: %s to %s", packageID, environment))

	// Get package from repository
	repo := repository.NewPackageRepository(userName, tx)
	pkg, err := repo.GetPackage(packageID)
	if err != nil {
		return nil, fmt.Errorf("package not found: %w", err)
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

	optionsJSON, _ := json.Marshal(options)
	deployAction.Options = string(optionsJSON)

	if err := repo.SaveAction(deployAction); err != nil {
		iLog.Warn(fmt.Sprintf("Failed to save deploy action: %v", err))
	}

	// Deploy based on package type
	var deploymentRecord *models.DeploymentRecord

	if pkg.PackageType == "database" {
		// Use the provided transaction for database deployments
		deployer := deploymgr.NewDatabaseDeployer(userName, tx, dbconn.DatabaseType)
		deploymentRecord, err = deployer.Deploy(pkg, options)
	} else if pkg.PackageType == "document" {
		// For document deployments, use docDB
		if docDB == nil {
			docDB = documents.DocDBCon
		}
		deployer := deploymgr.NewDocumentDeployer(docDB, userName)
		deploymentRecord, err = deployer.Deploy(pkg, options)
	} else {
		return nil, fmt.Errorf("unsupported package type: %s", pkg.PackageType)
	}

	if err != nil {
		// Update action status to failed
		now := time.Now()
		deployAction.ActionStatus = repository.ActionStatusFailed
		deployAction.CompletedAt = &now
		errorLog := []string{err.Error()}
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
		Environment:  environment,
		DatabaseName: dbconn.DatabaseName,
		DeployedAt:   time.Now(),
		DeployedBy:   userName,
		IsActive:     true,
		CreatedBy:    userName,
	}

	if err := repo.SaveDeployment(deployment); err != nil {
		iLog.Warn(fmt.Sprintf("Failed to save deployment record: %v", err))
	}

	iLog.Info(fmt.Sprintf("Package deployment completed successfully: %s", packageID))

	// Return deployment results
	result := map[string]interface{}{
		"status":           "completed",
		"deployment_id":    deployment.ID,
		"package_id":       packageID,
		"environment":      environment,
		"deployed_at":      deployment.DeployedAt,
		"records_deployed": deployAction.RecordsProcessed,
	}

	// Include PK/ID mappings if available
	if deploymentRecord.PKMappingResult != nil && len(deploymentRecord.PKMappingResult) > 0 {
		result["pk_mappings_count"] = len(deploymentRecord.PKMappingResult)
	}
	if deploymentRecord.IDMappingResult != nil && len(deploymentRecord.IDMappingResult) > 0 {
		result["id_mappings_count"] = len(deploymentRecord.IDMappingResult)
	}

	return result, nil
}

// getDefaultDeploymentOptions returns default deployment options
func getDefaultDeploymentOptions() models.DeploymentOptions {
	return models.DeploymentOptions{
		SkipExisting:       true,
		UpdateExisting:     false,
		ValidateReferences: true,
		CreateMissing:      false,
		RebuildIndexes:     false,
		BatchSize:          100,
		TransactionSize:    1000,
		ContinueOnError:    false,
		DryRun:             false,
	}
}
