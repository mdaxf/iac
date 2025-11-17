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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	deploymgr "github.com/mdaxf/iac/deployment/deploy"
	"github.com/mdaxf/iac/deployment/devops"
	"github.com/mdaxf/iac/deployment/models"
	packagemgr "github.com/mdaxf/iac/deployment/package"
	"github.com/mdaxf/iac/deployment/repository"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
)

// PackageDatabaseRequest represents a database packaging request
type PackageDatabaseRequest struct {
	Name          string                `json:"name"`
	Version       string                `json:"version"`
	Description   string                `json:"description"`
	Environment   string                `json:"environment"`   // dev, staging, production
	Filter        models.PackageFilter  `json:"filter"`
}

// PackageDocumentRequest represents a document packaging request
type PackageDocumentRequest struct {
	Name          string                `json:"name"`
	Version       string                `json:"version"`
	Description   string                `json:"description"`
	Environment   string                `json:"environment"`   // dev, staging, production
	Filter        models.PackageFilter  `json:"filter"`
}

// DeployPackageRequest represents a deployment request
type DeployPackageRequest struct {
	PackageData []byte                     `json:"package_data"`
	Options     models.DeploymentOptions   `json:"options"`
}

// VersionControlRequest represents a version control request
type VersionControlRequest struct {
	Action        string                 `json:"action"` // "commit", "branch", "merge", "tag", "revert"
	ObjectType    string                 `json:"object_type"`
	ObjectID      string                 `json:"object_id"`
	Content       map[string]interface{} `json:"content,omitempty"`
	CommitMessage string                 `json:"commit_message,omitempty"`
	Branch        string                 `json:"branch,omitempty"`
	SourceBranch  string                 `json:"source_branch,omitempty"`
	TargetBranch  string                 `json:"target_branch,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	VersionID     string                 `json:"version_id,omitempty"`
}

// PackageDatabase packages database tables
func PackageDatabase(w http.ResponseWriter, r *http.Request) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "DeploymentController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("deployment.PackageDatabase", elapsed)
	}()

	// Parse request
	var req PackageDatabaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Get user from context (simplified - in practice would use auth)
	user := "System"

	// Begin database transaction
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to begin transaction: %v", err), http.StatusInternalServerError)
		return
	}
	defer dbTx.Rollback()

	// Create packager
	packager := packagemgr.NewDatabasePackager(user, dbTx, dbconn.DatabaseType)

	// Package tables
	pkg, err := packager.PackageTables(req.Name, req.Version, user, req.Filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to package database: %v", err), http.StatusInternalServerError)
		return
	}

	pkg.Description = req.Description

	// Save package to database
	repo := repository.NewPackageRepository(user, dbTx)
	packageRecord, err := repo.SavePackage(pkg, req.Environment, &req.Filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save package: %v", err), http.StatusInternalServerError)
		return
	}

	// Record pack action
	now := time.Now()
	packAction := &repository.PackageActionRecord{
		PackageID:         pkg.ID,
		ActionType:        repository.ActionTypePack,
		ActionStatus:      repository.ActionStatusCompleted,
		SourceEnvironment: req.Environment,
		PerformedBy:       user,
		PerformedAt:       now,
		StartedAt:         &startTime,
		CompletedAt:       &now,
		TablesProcessed:   len(pkg.DatabaseData.Tables),
	}

	// Calculate records processed
	for _, table := range pkg.DatabaseData.Tables {
		packAction.RecordsProcessed += table.RowCount
		packAction.RecordsSucceeded += table.RowCount
	}

	if err := repo.SaveAction(packAction); err != nil {
		iLog.Warn(fmt.Sprintf("Failed to save pack action: %v", err))
	}

	// Export package
	data, err := packager.ExportPackage(pkg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to export package: %v", err), http.StatusInternalServerError)
		return
	}

	dbTx.Commit()

	// Return package with metadata
	response := map[string]interface{}{
		"package_id":   packageRecord.ID,
		"name":         packageRecord.Name,
		"version":      packageRecord.Version,
		"checksum":     packageRecord.Checksum,
		"file_size":    packageRecord.FileSize,
		"tables":       len(pkg.DatabaseData.Tables),
		"records":      packAction.RecordsProcessed,
		"package_data": data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	iLog.Info(fmt.Sprintf("Database packaged and saved: %s v%s (ID: %s)", pkg.Name, pkg.Version, pkg.ID))
}

// PackageDocuments packages document collections
func PackageDocuments(w http.ResponseWriter, r *http.Request) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "DeploymentController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("deployment.PackageDocuments", elapsed)
	}()

	// Parse request
	var req PackageDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Get user from context
	user := "System"

	// Create packager
	packager := packagemgr.NewDocumentPackager(documents.DocDBCon, user)

	// Package collections
	pkg, err := packager.PackageCollections(req.Name, req.Version, user, req.Filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to package documents: %v", err), http.StatusInternalServerError)
		return
	}

	pkg.Description = req.Description

	// Save package to database
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to begin transaction: %v", err), http.StatusInternalServerError)
		return
	}
	defer dbTx.Rollback()

	repo := repository.NewPackageRepository(user, dbTx)
	packageRecord, err := repo.SavePackage(pkg, req.Environment, &req.Filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save package: %v", err), http.StatusInternalServerError)
		return
	}

	// Record pack action
	now := time.Now()
	packAction := &repository.PackageActionRecord{
		PackageID:            pkg.ID,
		ActionType:           repository.ActionTypePack,
		ActionStatus:         repository.ActionStatusCompleted,
		SourceEnvironment:    req.Environment,
		PerformedBy:          user,
		PerformedAt:          now,
		StartedAt:            &startTime,
		CompletedAt:          &now,
		CollectionsProcessed: len(pkg.DocumentData.Collections),
	}

	// Calculate documents processed
	for _, collection := range pkg.DocumentData.Collections {
		packAction.RecordsProcessed += collection.DocumentCount
		packAction.RecordsSucceeded += collection.DocumentCount
	}

	if err := repo.SaveAction(packAction); err != nil {
		iLog.Warn(fmt.Sprintf("Failed to save pack action: %v", err))
	}

	// Export package
	data, err := packager.ExportPackage(pkg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to export package: %v", err), http.StatusInternalServerError)
		return
	}

	dbTx.Commit()

	// Return package with metadata
	response := map[string]interface{}{
		"package_id":   packageRecord.ID,
		"name":         packageRecord.Name,
		"version":      packageRecord.Version,
		"checksum":     packageRecord.Checksum,
		"file_size":    packageRecord.FileSize,
		"collections":  len(pkg.DocumentData.Collections),
		"documents":    packAction.RecordsProcessed,
		"package_data": data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	iLog.Info(fmt.Sprintf("Documents packaged and saved: %s v%s (ID: %s)", pkg.Name, pkg.Version, pkg.ID))
}

// DeployDatabasePackage deploys a database package
func DeployDatabasePackage(w http.ResponseWriter, r *http.Request) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "DeploymentController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("deployment.DeployDatabasePackage", elapsed)
	}()

	// Parse request
	var req DeployPackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Get user from context
	user := "System"

	// Import package
	packager := packagemgr.NewDatabasePackager(user, nil, dbconn.DatabaseType)
	pkg, err := packager.ImportPackage(req.PackageData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to import package: %v", err), http.StatusBadRequest)
		return
	}

	// Begin database transaction
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to begin transaction: %v", err), http.StatusInternalServerError)
		return
	}
	defer dbTx.Rollback()

	// Create deployer
	deployer := deploymgr.NewDatabaseDeployer(user, dbTx, dbconn.DatabaseType)

	// Deploy package
	record, err := deployer.Deploy(pkg, req.Options)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to deploy package: %v", err), http.StatusInternalServerError)
		return
	}

	// Commit if successful
	if record.Status == "completed" {
		dbTx.Commit()
	}

	// Return deployment record
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)

	iLog.Info(fmt.Sprintf("Database package deployed: %s", record.ID))
}

// DeployDocumentPackage deploys a document package
func DeployDocumentPackage(w http.ResponseWriter, r *http.Request) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "DeploymentController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("deployment.DeployDocumentPackage", elapsed)
	}()

	// Parse request
	var req DeployPackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Get user from context
	user := "System"

	// Import package
	packager := packagemgr.NewDocumentPackager(documents.DocDBCon, user)
	pkg, err := packager.ImportPackage(req.PackageData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to import package: %v", err), http.StatusBadRequest)
		return
	}

	// Create deployer
	deployer := deploymgr.NewDocumentDeployer(documents.DocDBCon, user)

	// Deploy package
	record, err := deployer.Deploy(pkg, req.Options)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to deploy package: %v", err), http.StatusInternalServerError)
		return
	}

	// Return deployment record
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)

	iLog.Info(fmt.Sprintf("Document package deployed: %s", record.ID))
}

// VersionControl handles version control operations
func VersionControl(w http.ResponseWriter, r *http.Request) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "VersionControl"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("deployment.VersionControl", elapsed)
	}()

	// Parse request
	var req VersionControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Get user from context
	user := "System"

	// Create version control manager
	vc := devops.NewVersionControl(documents.DocDBCon, user)

	var result interface{}
	var err error

	// Handle action
	switch req.Action {
	case "commit":
		result, err = vc.Commit(
			devops.ObjectType(req.ObjectType),
			req.ObjectID,
			req.Content,
			req.CommitMessage,
			req.Branch,
			user,
		)

	case "branch":
		result, err = vc.CreateBranch(
			devops.ObjectType(req.ObjectType),
			req.ObjectID,
			req.Branch,
			req.VersionID,
			user,
		)

	case "merge":
		result, err = vc.MergeBranch(
			devops.ObjectType(req.ObjectType),
			req.ObjectID,
			req.SourceBranch,
			req.TargetBranch,
			user,
		)

	case "tag":
		err = vc.TagVersion(req.VersionID, req.Tags)
		result = map[string]interface{}{
			"status": "tagged",
			"version_id": req.VersionID,
			"tags": req.Tags,
		}

	case "revert":
		result, err = vc.Revert(
			devops.ObjectType(req.ObjectType),
			req.ObjectID,
			req.VersionID,
			user,
		)

	case "get_latest":
		result, err = vc.GetLatestVersion(
			devops.ObjectType(req.ObjectType),
			req.ObjectID,
			req.Branch,
		)

	case "list_versions":
		result, err = vc.ListVersions(
			devops.ObjectType(req.ObjectType),
			req.ObjectID,
			req.Branch,
		)

	case "get_changelog":
		result, err = vc.GetChangelog(
			devops.ObjectType(req.ObjectType),
			req.ObjectID,
			50,
		)

	case "diff":
		// Expecting req.VersionID and req.TargetBranch (repurposed for version2)
		result, err = vc.Diff(req.VersionID, req.TargetBranch)

	case "auto_commit":
		result, err = vc.AutoCommit(
			devops.ObjectType(req.ObjectType),
			req.ObjectID,
			req.CommitMessage,
			user,
		)

	default:
		http.Error(w, fmt.Sprintf("Unknown action: %s", req.Action), http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Operation failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

	iLog.Info(fmt.Sprintf("Version control operation completed: %s", req.Action))
}

// PackageDevOpsObjects packages specific DevOps objects
func PackageDevOpsObjects(w http.ResponseWriter, r *http.Request) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "DeploymentController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("deployment.PackageDevOpsObjects", elapsed)
	}()

	// Parse request
	var req struct {
		ObjectType string   `json:"object_type"`
		ObjectIDs  []string `json:"object_ids"`
		Name       string   `json:"name"`
		Version    string   `json:"version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Get user from context
	user := "System"

	// Create packager
	packager := packagemgr.NewDocumentPackager(documents.DocDBCon, user)

	// Package DevOps objects
	pkg, err := packager.PackageDevOpsObjects(req.ObjectType, req.ObjectIDs, req.Name, req.Version, user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to package objects: %v", err), http.StatusInternalServerError)
		return
	}

	// Export package
	data, err := packager.ExportPackage(pkg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to export package: %v", err), http.StatusInternalServerError)
		return
	}

	// Return package
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-v%s.json", pkg.Name, pkg.Version))
	w.Write(data)

	iLog.Info(fmt.Sprintf("DevOps objects packaged: %s v%s", pkg.Name, pkg.Version))
}

// UploadPackage handles package file uploads
func UploadPackage(w http.ResponseWriter, r *http.Request) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "DeploymentController"}

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("package")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file content
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse package
	var pkg models.Package
	if err := json.Unmarshal(data, &pkg); err != nil {
		http.Error(w, fmt.Sprintf("Invalid package format: %v", err), http.StatusBadRequest)
		return
	}

	// Return package info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           pkg.ID,
		"name":         pkg.Name,
		"version":      pkg.Version,
		"package_type": pkg.PackageType,
		"created_at":   pkg.CreatedAt,
		"created_by":   pkg.CreatedBy,
	})

	iLog.Info(fmt.Sprintf("Package uploaded: %s v%s", pkg.Name, pkg.Version))
}
