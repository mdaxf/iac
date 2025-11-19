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
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	dbconn "github.com/mdaxf/iac/databases"
	deploymgr "github.com/mdaxf/iac/deployment/deploy"
	"github.com/mdaxf/iac/deployment/models"
	packagemgr "github.com/mdaxf/iac/deployment/package"
	"github.com/mdaxf/iac/deployment/repository"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
)

// PackageController handles package management operations
type PackageController struct{}

// CreatePackageRequest represents a request to create a package definition
type CreatePackageRequest struct {
	Name          string                 `json:"name" binding:"required"`
	Version       string                 `json:"version" binding:"required"`
	Description   string                 `json:"description"`
	PackageType   string                 `json:"package_type" binding:"required"` // "database" or "document"
	Environment   string                 `json:"environment"`                     // dev, staging, production
	Filter        models.PackageFilter   `json:"filter"`
	Metadata      map[string]interface{} `json:"metadata"`
	IncludeParent bool                   `json:"include_parent"`
}

// UpdatePackageRequest represents a request to update package metadata
type UpdatePackageRequest struct {
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// GeneratePackageRequest represents a request to generate a package file
type GeneratePackageRequest struct {
	Tables      []string          `json:"tables"`       // For database packages
	Collections []string          `json:"collections"`  // For document packages
	WhereClause map[string]string `json:"where_clause"` // Table/Collection -> WHERE condition
	Format      string            `json:"format"`       // "json" or "zip"
}

// DeployPackageRequest represents a deployment request
type DeployPackageRequest struct {
	PackageID         string                     `json:"package_id"`
	Environment       string                     `json:"environment"`
	Options           models.DeploymentOptions   `json:"options"`
	ScheduleAt        *time.Time                 `json:"schedule_at"`        // Optional: Schedule for later
	RunAsBackgroundJob bool                      `json:"run_as_background"` // Run as background job
}

// ImportPackageRequest represents an import request
type ImportPackageRequest struct {
	PackageData json.RawMessage `json:"package_data" binding:"required"`
	Environment string          `json:"environment"`
}

// ListPackages godoc
// @Summary List all packages
// @Description Get a list of all packages with optional filters
// @Tags packages
// @Accept json
// @Produce json
// @Param package_type query string false "Package type filter (database/document)"
// @Param environment query string false "Environment filter (dev/staging/production)"
// @Param status query string false "Status filter (active/archived/deleted)"
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} repository.PackageRecord
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages [get]
func (pc *PackageController) ListPackages(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.ListPackages"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("PackageController.ListPackages", elapsed)
	}()

	// Get query parameters
	packageType := c.Query("package_type")
	environment := c.Query("environment")
	status := c.Query("status")
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

	// Create repository
	repo := repository.NewPackageRepository(userName, dbTx)

	// List packages
	packages, err := repo.ListPackages(packageType, environment, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list packages: %v", err)})
		return
	}

	dbTx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"packages": packages,
		"count":    len(packages),
		"limit":    limit,
		"offset":   offset,
	})
}

// GetPackage godoc
// @Summary Get package by ID
// @Description Get detailed information about a specific package
// @Tags packages
// @Accept json
// @Produce json
// @Param id path string true "Package ID"
// @Success 200 {object} models.Package
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages/{id} [get]
func (pc *PackageController) GetPackage(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.GetPackage"}
	packageID := c.Param("id")

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

	// Create repository
	repo := repository.NewPackageRepository(userName, dbTx)

	// Get package
	pkg, err := repo.GetPackage(packageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Package not found: %v", err)})
		return
	}

	// Get recent actions
	actions, err := repo.GetActionsByPackage(packageID, 10)
	if err != nil {
		iLog.Warn(fmt.Sprintf("Failed to get actions: %v", err))
	}

	dbTx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"package":        pkg,
		"recent_actions": actions,
	})
}

// CreatePackage godoc
// @Summary Create a new package definition
// @Description Create a new package definition with specified tables/collections
// @Tags packages
// @Accept json
// @Produce json
// @Param request body CreatePackageRequest true "Package creation request"
// @Success 201 {object} repository.PackageRecord
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages [post]
func (pc *PackageController) CreatePackage(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.CreatePackage"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("PackageController.CreatePackage", elapsed)
	}()

	// Parse request
	var req CreatePackageRequest
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

	// Begin database transaction
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to begin transaction: %v", err)})
		return
	}
	defer dbTx.Rollback()

	var pkg *models.Package
	var packageRecord *repository.PackageRecord

	// Create package based on type
	if req.PackageType == "database" {
		// Create database packager
		packager := packagemgr.NewDatabasePackager(userName, dbTx, dbconn.DatabaseType)

		// Package tables
		pkg, err = packager.PackageTables(req.Name, req.Version, userName, req.Filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to package database: %v", err)})
			return
		}

	} else if req.PackageType == "document" {
		// Create document packager
		packager := packagemgr.NewDocumentPackager(documents.DocDBCon, userName)

		// Package collections
		pkg, err = packager.PackageCollections(req.Name, req.Version, userName, req.Filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to package documents: %v", err)})
			return
		}

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package_type. Must be 'database' or 'document'"})
		return
	}

	// Set package properties
	pkg.Description = req.Description
	pkg.IncludeParent = req.IncludeParent
	if req.Metadata != nil {
		pkg.Metadata = req.Metadata
	}

	// Save package to database
	repo := repository.NewPackageRepository(userName, dbTx)
	packageRecord, err = repo.SavePackage(pkg, req.Environment, &req.Filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save package: %v", err)})
		return
	}

	// Record pack action
	now := time.Now()
	packAction := &repository.PackageActionRecord{
		PackageID:         pkg.ID,
		ActionType:        repository.ActionTypePack,
		ActionStatus:      repository.ActionStatusCompleted,
		SourceEnvironment: req.Environment,
		PerformedBy:       userName,
		PerformedAt:       now,
		StartedAt:         &startTime,
		CompletedAt:       &now,
	}

	// Calculate statistics
	if pkg.DatabaseData != nil {
		packAction.TablesProcessed = len(pkg.DatabaseData.Tables)
		for _, table := range pkg.DatabaseData.Tables {
			packAction.RecordsProcessed += table.RowCount
			packAction.RecordsSucceeded += table.RowCount
		}
	} else if pkg.DocumentData != nil {
		packAction.CollectionsProcessed = len(pkg.DocumentData.Collections)
		for _, collection := range pkg.DocumentData.Collections {
			packAction.RecordsProcessed += collection.DocumentCount
			packAction.RecordsSucceeded += collection.DocumentCount
		}
	}

	if err := repo.SaveAction(packAction); err != nil {
		iLog.Warn(fmt.Sprintf("Failed to save pack action: %v", err))
	}

	dbTx.Commit()

	iLog.Info(fmt.Sprintf("Package created: %s v%s (ID: %s)", pkg.Name, pkg.Version, pkg.ID))

	c.JSON(http.StatusCreated, gin.H{
		"package": packageRecord,
		"message": "Package created successfully",
	})
}

// UpdatePackage godoc
// @Summary Update package metadata
// @Description Update package description, status, tags, and metadata
// @Tags packages
// @Accept json
// @Produce json
// @Param id path string true "Package ID"
// @Param request body UpdatePackageRequest true "Package update request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages/{id} [put]
func (pc *PackageController) UpdatePackage(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.UpdatePackage"}
	packageID := c.Param("id")

	// Parse request
	var req UpdatePackageRequest
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

	// Begin database transaction
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to begin transaction: %v", err)})
		return
	}
	defer dbTx.Rollback()

	// Build update query
	dbOp := dbconn.NewDBOperation(userName, dbTx, logger.Framework)

	// Update package metadata
	tagsJSON, _ := json.Marshal(req.Tags)
	metadataJSON, _ := json.Marshal(req.Metadata)

	query := fmt.Sprintf(`
		UPDATE %s
		SET description = %s, status = %s, tags = %s, metadata = %s, modifiedby = %s, modifiedon = %s
		WHERE id = %s AND active = %s`,
		dbOp.QuoteIdentifier("iacpackages"),
		dbOp.GetPlaceholder(1),
		dbOp.GetPlaceholder(2),
		dbOp.GetPlaceholder(3),
		dbOp.GetPlaceholder(4),
		dbOp.GetPlaceholder(5),
		dbOp.GetPlaceholder(6),
		dbOp.GetPlaceholder(7),
		dbOp.GetPlaceholder(8))

	_, err = dbOp.Exec(query, req.Description, req.Status, string(tagsJSON), string(metadataJSON), userName, time.Now(), packageID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update package: %v", err)})
		return
	}

	dbTx.Commit()

	iLog.Info(fmt.Sprintf("Package updated: %s", packageID))

	c.JSON(http.StatusOK, gin.H{
		"message":    "Package updated successfully",
		"package_id": packageID,
	})
}

// DeletePackage godoc
// @Summary Delete package (soft delete)
// @Description Soft delete a package by setting active = false
// @Tags packages
// @Accept json
// @Produce json
// @Param id path string true "Package ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages/{id} [delete]
func (pc *PackageController) DeletePackage(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.DeletePackage"}
	packageID := c.Param("id")

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

	// Build update query for soft delete
	dbOp := dbconn.NewDBOperation(userName, dbTx, logger.Framework)

	query := fmt.Sprintf(`
		UPDATE %s
		SET active = %s, status = %s, modifiedby = %s, modifiedon = %s
		WHERE id = %s`,
		dbOp.QuoteIdentifier("iacpackages"),
		dbOp.GetPlaceholder(1),
		dbOp.GetPlaceholder(2),
		dbOp.GetPlaceholder(3),
		dbOp.GetPlaceholder(4),
		dbOp.GetPlaceholder(5))

	_, err = dbOp.Exec(query, false, repository.PackageStatusDeleted, userName, time.Now(), packageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete package: %v", err)})
		return
	}

	dbTx.Commit()

	iLog.Info(fmt.Sprintf("Package deleted: %s", packageID))

	c.JSON(http.StatusOK, gin.H{
		"message":    "Package deleted successfully",
		"package_id": packageID,
	})
}

// GeneratePackageFile godoc
// @Summary Generate package file
// @Description Generate a package file (JSON or ZIP) with selected data
// @Tags packages
// @Accept json
// @Produce application/zip
// @Param id path string true "Package ID"
// @Param request body GeneratePackageRequest true "Package generation request"
// @Success 200 {file} application/zip
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages/{id}/generate [post]
func (pc *PackageController) GeneratePackageFile(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.GeneratePackageFile"}
	packageID := c.Param("id")

	// Parse request
	var req GeneratePackageRequest
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

	// Begin database transaction
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to begin transaction: %v", err)})
		return
	}
	defer dbTx.Rollback()

	// Get package
	repo := repository.NewPackageRepository(userName, dbTx)
	pkg, err := repo.GetPackage(packageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Package not found: %v", err)})
		return
	}

	// Export package based on type
	var packageData []byte
	if pkg.PackageType == "database" {
		packager := packagemgr.NewDatabasePackager(userName, dbTx, dbconn.DatabaseType)
		packageData, err = packager.ExportPackage(pkg)
	} else {
		packager := packagemgr.NewDocumentPackager(documents.DocDBCon, userName)
		packageData, err = packager.ExportPackage(pkg)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to export package: %v", err)})
		return
	}

	dbTx.Commit()

	// Generate ZIP if requested
	if req.Format == "zip" {
		zipBuffer := new(bytes.Buffer)
		zipWriter := zip.NewWriter(zipBuffer)

		// Add package JSON to ZIP
		packageFile, err := zipWriter.Create(fmt.Sprintf("%s-v%s.json", pkg.Name, pkg.Version))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create ZIP: %v", err)})
			return
		}

		_, err = packageFile.Write(packageData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to write to ZIP: %v", err)})
			return
		}

		// Add metadata file
		metadata := map[string]interface{}{
			"package_id":   pkg.ID,
			"name":         pkg.Name,
			"version":      pkg.Version,
			"package_type": pkg.PackageType,
			"created_at":   pkg.CreatedAt,
			"created_by":   pkg.CreatedBy,
			"description":  pkg.Description,
		}
		metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
		metadataFile, _ := zipWriter.Create("metadata.json")
		metadataFile.Write(metadataJSON)

		zipWriter.Close()

		// Return ZIP file
		c.Header("Content-Type", "application/zip")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s-v%s.zip", pkg.Name, pkg.Version))
		c.Data(http.StatusOK, "application/zip", zipBuffer.Bytes())

		iLog.Info(fmt.Sprintf("Package ZIP generated: %s v%s", pkg.Name, pkg.Version))
	} else {
		// Return JSON
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s-v%s.json", pkg.Name, pkg.Version))
		c.Data(http.StatusOK, "application/json", packageData)

		iLog.Info(fmt.Sprintf("Package JSON generated: %s v%s", pkg.Name, pkg.Version))
	}
}

// ImportPackage godoc
// @Summary Import a package
// @Description Import a package from JSON data and save it to the database
// @Tags packages
// @Accept json
// @Produce json
// @Param request body ImportPackageRequest true "Package import request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages/import [post]
func (pc *PackageController) ImportPackage(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.ImportPackage"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("PackageController.ImportPackage", elapsed)
	}()

	// Parse request
	var req ImportPackageRequest
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

	// Parse package
	var pkg models.Package
	if err := json.Unmarshal(req.PackageData, &pkg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid package format: %v", err)})
		return
	}

	// Begin database transaction
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to begin transaction: %v", err)})
		return
	}
	defer dbTx.Rollback()

	// Save package to database
	repo := repository.NewPackageRepository(userName, dbTx)
	packageRecord, err := repo.SavePackage(&pkg, req.Environment, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save package: %v", err)})
		return
	}

	// Record import action
	now := time.Now()
	importAction := &repository.PackageActionRecord{
		PackageID:         pkg.ID,
		ActionType:        repository.ActionTypeImport,
		ActionStatus:      repository.ActionStatusCompleted,
		TargetEnvironment: req.Environment,
		PerformedBy:       userName,
		PerformedAt:       now,
		StartedAt:         &startTime,
		CompletedAt:       &now,
	}

	if err := repo.SaveAction(importAction); err != nil {
		iLog.Warn(fmt.Sprintf("Failed to save import action: %v", err))
	}

	dbTx.Commit()

	iLog.Info(fmt.Sprintf("Package imported: %s v%s (ID: %s)", pkg.Name, pkg.Version, pkg.ID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Package imported successfully",
		"package": packageRecord,
	})
}

// To be continued in next part...
