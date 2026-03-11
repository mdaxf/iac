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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/deployment/models"
	packagemgr "github.com/mdaxf/iac/deployment/package"
	"github.com/mdaxf/iac/deployment/repository"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/signalrserver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoopts "go.mongodb.org/mongo-driver/mongo/options"
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

	if dbconn.DB == nil {
		c.JSON(http.StatusOK, gin.H{"packages": []interface{}{}, "count": 0, "limit": limit, "offset": offset})
		return
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
		iLog.Error(fmt.Sprintf("Failed to list packages: %v", err))
		// Return empty list instead of 500 when table doesn't exist yet
		c.JSON(http.StatusOK, gin.H{"packages": []interface{}{}, "count": 0, "limit": limit, "offset": offset})
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

	if err := dbTx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to commit package: %v", err)})
		return
	}

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

	_, err = dbOp.Execute(query, req.Description, req.Status, string(tagsJSON), string(metadataJSON), userName, time.Now(), packageID, true)
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

	_, err = dbOp.Execute(query, false, repository.PackageStatusDeleted, userName, time.Now(), packageID)
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

	if err := dbTx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to commit package import: %v", err)})
		return
	}

	iLog.Info(fmt.Sprintf("Package imported: %s v%s (ID: %s)", pkg.Name, pkg.Version, pkg.ID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Package imported successfully",
		"package": packageRecord,
	})
}

// ImportPackageFile godoc
// @Summary Import a package from an uploaded file (JSON or ZIP)
// @Description Accepts a multipart upload of a package JSON or ZIP file and imports it.
// @Tags packages
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Package file (.json or .zip)"
// @Param environment formData string false "Target environment"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages/import-file [post]
func (pc *PackageController) ImportPackageFile(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.ImportPackageFile"}
	startTime := time.Now()
	defer func() {
		iLog.PerformanceWithDuration("PackageController.ImportPackageFile", time.Since(startTime))
	}()

	user, _ := c.Get("user")
	userName := fmt.Sprintf("%v", user)
	if userName == "" {
		userName = "System"
	}
	environment := c.PostForm("environment")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("No file uploaded: %v", err)})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to open file: %v", err)})
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to read file: %v", err)})
		return
	}

	// Determine file type and extract package JSON
	var packageJSON []byte
	filename := strings.ToLower(fileHeader.Filename)
	if strings.HasSuffix(filename, ".zip") {
		// Extract the package JSON from the ZIP archive (first .json file that isn't metadata.json)
		zr, zerr := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
		if zerr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid ZIP file: %v", zerr)})
			return
		}
		for _, f := range zr.File {
			if strings.HasSuffix(f.Name, ".json") && f.Name != "metadata.json" {
				rc, oerr := f.Open()
				if oerr != nil {
					continue
				}
				packageJSON, _ = io.ReadAll(rc)
				rc.Close()
				break
			}
		}
		if len(packageJSON) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No package JSON found in ZIP"})
			return
		}
	} else {
		packageJSON = fileBytes
	}

	// Parse and import
	var pkg models.Package
	if err := json.Unmarshal(packageJSON, &pkg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid package format: %v", err)})
		return
	}

	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to begin transaction: %v", err)})
		return
	}
	defer dbTx.Rollback()

	repo := repository.NewPackageRepository(userName, dbTx)
	packageRecord, err := repo.SavePackage(&pkg, environment, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save package: %v", err)})
		return
	}

	now := time.Now()
	_ = repo.SaveAction(&repository.PackageActionRecord{
		PackageID:         pkg.ID,
		ActionType:        repository.ActionTypeImport,
		ActionStatus:      repository.ActionStatusCompleted,
		TargetEnvironment: environment,
		PerformedBy:       userName,
		PerformedAt:       now,
		StartedAt:         &startTime,
		CompletedAt:       &now,
	})

	if err := dbTx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to commit import: %v", err)})
		return
	}

	iLog.Info(fmt.Sprintf("Package file imported: %s v%s (ID: %s) from %s", pkg.Name, pkg.Version, pkg.ID, fileHeader.Filename))
	c.JSON(http.StatusOK, gin.H{
		"message": "Package imported successfully",
		"package": packageRecord,
	})
}

// AnalyzeRequest contains tables/collections to analyze before packaging
type AnalyzeRequest struct {
	PackageType string   `json:"package_type" binding:"required"` // "database" or "document"
	Tables      []string `json:"tables,omitempty"`
	Collections []string `json:"collections,omitempty"`
}

// AnalyzeSource godoc
// @Summary Analyze tables or collections before packaging
// @Description Returns schema, FK relationships, and global-key candidates to help configure a package
// @Tags packages
// @Accept json
// @Produce json
// @Param request body AnalyzeRequest true "Analyze request"
// @Success 200 {object} models.AnalysisResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages/analyze [post]
func (pc *PackageController) AnalyzeSource(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.AnalyzeSource"}

	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	user, _ := c.Get("user")
	userName := fmt.Sprintf("%v", user)
	if userName == "" {
		userName = "System"
	}

	result := models.AnalysisResult{
		PackageType:   req.PackageType,
		Entities:      make([]models.EntityAnalysis, 0),
		Relationships: make([]models.RelationshipAnalysis, 0),
	}

	if req.PackageType == "database" {
		dbTx, err := dbconn.DB.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to begin transaction: %v", err)})
			return
		}
		defer dbTx.Rollback()

		packager := packagemgr.NewDatabasePackager(userName, dbTx, dbconn.DatabaseType)

		// Build a temp package to extract schema + FK info per table
		filter := models.PackageFilter{Tables: req.Tables, IncludeRelated: false}
		pkg, err := packager.PackageTables("__analyze__", "0", userName, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to analyze tables: %v", err)})
			return
		}

		for _, table := range pkg.DatabaseData.Tables {
			entity := models.EntityAnalysis{
				Name:                table.TableName,
				EntityType:          "table",
				RowCount:            table.RowCount,
				Columns:             table.Columns,
				PKColumns:           table.PKColumns,
				GlobalKeyCandidates: suggestGlobalKeyCandidates(table.Columns, table.PKColumns),
			}
			result.Entities = append(result.Entities, entity)
		}

		for _, rel := range pkg.DatabaseData.Relationships {
			result.Relationships = append(result.Relationships, models.RelationshipAnalysis{
				SourceEntity:     rel.SourceTable,
				SourceColumn:     rel.SourceColumn,
				TargetEntity:     rel.TargetTable,
				TargetColumn:     rel.TargetColumn,
				ConstraintName:   rel.ConstraintName,
				IncludeByDefault: true,
			})
		}

		dbTx.Commit()

	} else if req.PackageType == "document" {
		packager := packagemgr.NewDocumentPackager(documents.DocDBCon, userName)

		filter := models.PackageFilter{Collections: req.Collections}
		pkg, err := packager.PackageCollections("__analyze__", "0", userName, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to analyze collections: %v", err)})
			return
		}

		for _, coll := range pkg.DocumentData.Collections {
			idMapping := pkg.DocumentData.IDMappings[coll.CollectionName]
			entity := models.EntityAnalysis{
				Name:                coll.CollectionName,
				EntityType:          "collection",
				RowCount:            coll.DocumentCount,
				GlobalKeyCandidates: suggestDocGlobalKeyCandidates(idMapping, coll.Documents),
			}
			result.Entities = append(result.Entities, entity)
		}

		for _, ref := range pkg.DocumentData.References {
			result.Relationships = append(result.Relationships, models.RelationshipAnalysis{
				SourceEntity:     ref.SourceCollection,
				SourceColumn:     ref.SourceField,
				TargetEntity:     ref.TargetCollection,
				TargetColumn:     ref.TargetIDField,
				IncludeByDefault: true,
			})
		}

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package_type must be 'database' or 'document'"})
		return
	}

	iLog.Info(fmt.Sprintf("Analysis completed: %d entities, %d relationships", len(result.Entities), len(result.Relationships)))
	c.JSON(http.StatusOK, result)
}

// suggestGlobalKeyCandidates suggests natural key columns from a table's schema
func suggestGlobalKeyCandidates(columns []models.ColumnInfo, pkColumns []string) [][]string {
	candidates := make([][]string, 0)

	// Single candidate: "name" or "code" column alone
	for _, col := range columns {
		lower := strings.ToLower(col.Name)
		if lower == "name" || lower == "code" || lower == "key" || lower == "slug" {
			candidates = append(candidates, []string{col.Name})
		}
	}

	// Compound candidate: "name" + "version"/"revision"
	var nameCol, versionCol string
	for _, col := range columns {
		lower := strings.ToLower(col.Name)
		if lower == "name" {
			nameCol = col.Name
		}
		if lower == "version" || lower == "revision" || lower == "rev" {
			versionCol = col.Name
		}
	}
	if nameCol != "" && versionCol != "" {
		candidates = append(candidates, []string{nameCol, versionCol})
	}

	return candidates
}

// suggestDocGlobalKeyCandidates suggests global key fields for a document collection
func suggestDocGlobalKeyCandidates(idMapping models.IDMapping, docs []map[string]interface{}) [][]string {
	// If _id is UUID → no global key needed (UUID itself is the global key)
	if idMapping.IDType == "uuid" {
		return nil
	}

	// Inspect first document for common field names
	if len(docs) == 0 {
		return [][]string{{"name", "version"}}
	}

	doc := docs[0]
	candidates := make([][]string, 0)

	var nameKey, versionKey string
	for k := range doc {
		lower := strings.ToLower(k)
		if lower == "name" {
			nameKey = k
		}
		if lower == "version" || lower == "revision" || lower == "rev" {
			versionKey = k
		}
	}

	if nameKey != "" && versionKey != "" {
		candidates = append(candidates, []string{nameKey, versionKey})
	} else if nameKey != "" {
		candidates = append(candidates, []string{nameKey})
	}

	return candidates
}

// GetSources godoc
// @Summary List available tables and MongoDB collections
// @Description Returns all tables (with schema) from the connected SQL database and all
//
//	collections from MongoDB, so the user can browse and select what to package.
//
// @Tags packages
// @Produce json
// @Success 200 {object} models.SourcesResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages/sources [get]
func (pc *PackageController) GetSources(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.GetSources"}

	resp := models.SourcesResponse{
		DatabaseType:    dbconn.DatabaseType,
		DatabaseName:    extractDBName(dbconn.DatabaseConnection),
		Tables:          make([]models.TableSource, 0),
		Collections:     make([]models.CollectionSource, 0),
		DocDatabaseName: "",
	}

	// ── SQL tables ───────────────────────────────────────────────────────────
	if dbconn.DB != nil {
		packager := packagemgr.NewDatabasePackagerFromDB("System", dbconn.DB, dbconn.DatabaseType)
		tableNames, err := packager.ListTables()
		if err != nil {
			iLog.Warn(fmt.Sprintf("Failed to list SQL tables: %v", err))
		} else {
			for _, tbl := range tableNames {
				src := models.TableSource{Name: tbl}
				if info, err := packager.GetTableSourceInfo(tbl); err == nil {
					src.Columns = info.Columns
					src.PKColumns = info.PKColumns
					src.FKColumns = info.FKColumns
					src.RowCount = info.RowCount
				}
				resp.Tables = append(resp.Tables, src)
			}
		}
	}

	// ── MongoDB collections ───────────────────────────────────────────────────
	if documents.DocDBCon != nil && documents.DocDBCon.MongoDBDatabase != nil {
		resp.DocDatabaseName = documents.DocDBCon.DatabaseName
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		collNames, err := documents.DocDBCon.MongoDBDatabase.ListCollectionNames(ctx, bson.M{})
		if err != nil {
			iLog.Warn(fmt.Sprintf("Failed to list MongoDB collections: %v", err))
		} else {
			for _, name := range collNames {
				if strings.HasPrefix(name, "system.") {
					continue
				}
				cnt, _ := documents.DocDBCon.MongoDBDatabase.Collection(name).CountDocuments(ctx, bson.M{})
				resp.Collections = append(resp.Collections, models.CollectionSource{
					Name:          name,
					DocumentCount: cnt,
				})
			}
		}
	}

	c.JSON(http.StatusOK, resp)
}

// GetTableRecords godoc
// @Summary Browse rows in a SQL table
// @Description Returns paginated rows from the specified table — used by the record picker
// @Tags packages
// @Produce json
// @Param table path string true "Table name"
// @Param limit query int false "Rows per page" default(50)
// @Param offset query int false "Row offset" default(0)
// @Success 200 {object} models.TableRecordsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages/sources/{table}/records [get]
func (pc *PackageController) GetTableRecords(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.GetTableRecords"}

	tableName := c.Param("source")
	if tableName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table name is required"})
		return
	}

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := fmt.Sscanf(l, "%d", &limit); v != 1 || err != nil || limit <= 0 {
			limit = 50
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := fmt.Sscanf(o, "%d", &offset); v != 1 || err != nil || offset < 0 {
			offset = 0
		}
	}
	if limit > 200 {
		limit = 200
	}

	if dbconn.DB == nil {
		c.JSON(http.StatusOK, &models.TableRecordsResponse{
			Table: tableName, Total: 0, Limit: limit, Offset: offset,
			Columns: []string{}, PKCols: []string{}, Rows: []map[string]interface{}{},
		})
		return
	}

	packager := packagemgr.NewDatabasePackagerFromDB("System", dbconn.DB, dbconn.DatabaseType)

	total, err := packager.CountTableRows(tableName)
	if err != nil {
		iLog.Warn(fmt.Sprintf("Failed to count rows for %s: %v", tableName, err))
	}

	colNames, pkCols, rows, err := packager.QueryTableRows(tableName, limit, offset)
	if err != nil {
		iLog.Error(fmt.Sprintf("QueryTableRows error for %s: %v", tableName, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to query table %s: %v", tableName, err)})
		return
	}

	c.JSON(http.StatusOK, &models.TableRecordsResponse{
		Table:   tableName,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		Columns: colNames,
		PKCols:  pkCols,
		Rows:    rows,
	})
}

// GetCollectionDocuments godoc
// @Summary Browse documents in a MongoDB collection
// @Description Returns paginated documents from the specified collection — used by the record picker
// @Tags packages
// @Produce json
// @Param collection path string true "Collection name"
// @Param limit query int false "Documents per page" default(50)
// @Param offset query int false "Document offset" default(0)
// @Success 200 {object} models.CollectionDocumentsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages/sources/{collection}/documents [get]
func (pc *PackageController) GetCollectionDocuments(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.GetCollectionDocuments"}

	collectionName := c.Param("source")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collection name is required"})
		return
	}

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := fmt.Sscanf(l, "%d", &limit); v != 1 || err != nil || limit <= 0 {
			limit = 50
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := fmt.Sscanf(o, "%d", &offset); v != 1 || err != nil || offset < 0 {
			offset = 0
		}
	}
	if limit > 200 {
		limit = 200
	}

	empty := &models.CollectionDocumentsResponse{
		Collection: collectionName, Total: 0, Limit: limit, Offset: offset,
		Columns: []string{}, IDField: "_id", Rows: []map[string]interface{}{},
	}

	if documents.DocDBCon == nil || documents.DocDBCon.MongoDBDatabase == nil {
		c.JSON(http.StatusOK, empty)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	coll := documents.DocDBCon.MongoDBDatabase.Collection(collectionName)

	total, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		iLog.Warn(fmt.Sprintf("Failed to count docs in %s: %v", collectionName, err))
	}

	findOpts := mongoopts.Find().SetLimit(int64(limit)).SetSkip(int64(offset))
	cursor, err := coll.Find(ctx, bson.M{}, findOpts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to query collection %s: %v", collectionName, err)})
		return
	}
	defer cursor.Close(ctx)

	rows := make([]map[string]interface{}, 0)
	for cursor.Next(ctx) {
		var raw bson.M
		if err := cursor.Decode(&raw); err != nil {
			continue
		}
		row := make(map[string]interface{}, len(raw))
		for k, v := range raw {
			row[k] = bsonToJSON(v)
		}
		rows = append(rows, row)
	}

	// Collect ordered column names: _id first, rest sorted
	colSet := map[string]bool{}
	for _, row := range rows {
		for k := range row {
			colSet[k] = true
		}
	}
	columns := []string{}
	if colSet["_id"] {
		columns = append(columns, "_id")
		delete(colSet, "_id")
	}
	rest := make([]string, 0, len(colSet))
	for k := range colSet {
		rest = append(rest, k)
	}
	sort.Strings(rest)
	columns = append(columns, rest...)

	c.JSON(http.StatusOK, &models.CollectionDocumentsResponse{
		Collection: collectionName,
		Total:      int(total),
		Limit:      limit,
		Offset:     offset,
		Columns:    columns,
		IDField:    "_id",
		Rows:       rows,
	})
}

// bsonToJSON converts BSON-specific types (ObjectID, DateTime, etc.) to JSON-safe values.
func bsonToJSON(v interface{}) interface{} {
	switch t := v.(type) {
	case primitive.ObjectID:
		return t.Hex()
	case primitive.DateTime:
		return t.Time().UTC().Format(time.RFC3339)
	case primitive.Timestamp:
		return fmt.Sprintf("%d", t.T)
	case primitive.Binary:
		return fmt.Sprintf("<binary len=%d>", len(t.Data))
	case bson.M:
		out := make(map[string]interface{}, len(t))
		for k, val := range t {
			out[k] = bsonToJSON(val)
		}
		return out
	case bson.A:
		out := make([]interface{}, len(t))
		for i, val := range t {
			out[i] = bsonToJSON(val)
		}
		return out
	case nil:
		return nil
	default:
		return v
	}
}

// extractDBName parses the database name from a connection string.
// Handles MySQL DSN (user:pass@tcp(host)/dbname) and PostgreSQL DSN (host=x dbname=x).
func extractDBName(conn string) string {
	// MySQL: ...@tcp(host:port)/dbname?params
	if idx := strings.LastIndex(conn, "/"); idx >= 0 {
		rest := conn[idx+1:]
		if q := strings.Index(rest, "?"); q >= 0 {
			rest = rest[:q]
		}
		if rest != "" {
			return rest
		}
	}
	// PostgreSQL DSN: dbname=xxx
	for _, part := range strings.Fields(conn) {
		if strings.HasPrefix(part, "dbname=") {
			return strings.TrimPrefix(part, "dbname=")
		}
	}
	return ""
}

// ─── PushPackage ──────────────────────────────────────────────────────────────

// PushPackageRequest is the body for POST /api/packages/:id/push
type PushPackageRequest struct {
	TargetURL   string `json:"target_url" binding:"required"` // e.g. https://remote.iac.example.com
	APIKey      string `json:"api_key"`                       // Bearer token for target IAC
	Environment string `json:"environment"`                   // override environment on import
}

// PushPackage godoc
// @Summary Push a package to another IAC environment
// @Description Exports the package as JSON and imports it into the target IAC instance via its REST API
// @Tags packages
// @Accept json
// @Produce json
// @Param id path string true "Package ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/packages/{id}/push [post]
func (pc *PackageController) PushPackage(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageController.PushPackage"}

	packageID := c.Param("id")
	var req PushPackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	targetURL := strings.TrimRight(req.TargetURL, "/")

	pushBroadcast(packageID, "starting", targetURL, "")

	if dbconn.DB == nil {
		pushBroadcast(packageID, "error", targetURL, "Database not available")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	// Load package from SQL
	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		pushBroadcast(packageID, "error", targetURL, "Internal error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to begin transaction: %v", err)})
		return
	}
	defer dbTx.Rollback()

	repo := repository.NewPackageRepository("System", dbTx)
	pkg, err := repo.GetPackage(packageID)
	if err != nil {
		pushBroadcast(packageID, "error", targetURL, "Package not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Package not found"})
		return
	}
	dbTx.Commit()

	pushBroadcast(packageID, "pushing", targetURL, "")

	// Serialize package
	pkgBytes, err := json.Marshal(pkg)
	if err != nil {
		pushBroadcast(packageID, "error", targetURL, "Serialization failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to marshal package: %v", err)})
		return
	}

	payload := map[string]interface{}{
		"package_data": json.RawMessage(pkgBytes),
		"environment":  req.Environment,
	}
	payloadBytes, _ := json.Marshal(payload)

	// POST to target IAC import endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()

	importURL := targetURL + "/api/packages/import"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, importURL, bytes.NewReader(payloadBytes))
	if err != nil {
		pushBroadcast(packageID, "error", targetURL, "Failed to build request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to build request: %v", err)})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if req.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
	}

	client := &http.Client{Timeout: 55 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		iLog.Error(fmt.Sprintf("Push to %s failed: %v", importURL, err))
		pushBroadcast(packageID, "error", targetURL, "Failed to reach target: "+err.Error())
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Failed to reach target: %v", err)})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result) //nolint:errcheck

	if resp.StatusCode >= 300 {
		msg := "Target IAC returned an error"
		if m, ok := result["error"].(string); ok {
			msg = m
		}
		pushBroadcast(packageID, "error", targetURL, msg)
		c.JSON(http.StatusBadGateway, gin.H{"error": msg, "target_status": resp.StatusCode})
		return
	}

	pushBroadcast(packageID, "success", targetURL, "")
	iLog.Info(fmt.Sprintf("Package %s pushed to %s successfully", packageID, targetURL))
	c.JSON(http.StatusOK, gin.H{
		"message":       "Package pushed successfully",
		"target_url":    targetURL,
		"target_result": result,
	})
}

// pushBroadcast sends a SignalR status event for a push operation.
func pushBroadcast(packageID, status, targetURL, errMsg string) {
	srv := signalrserver.GetGlobalSignalRServer()
	if srv == nil || !srv.IsRunning() {
		return
	}
	payload := map[string]string{"package_id": packageID, "status": status}
	if targetURL != "" {
		payload["target_url"] = targetURL
	}
	if errMsg != "" {
		payload["error"] = errMsg
	}
	data, _ := json.Marshal(payload)
	srv.BroadcastToHub("package.push.status."+packageID, string(data))
}
