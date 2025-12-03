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

package report

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/documents/schema"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/services"
)

// ReportDocController handles HTTP requests for report documents in MongoDB
type ReportDocController struct {
	reportDocService      *services.ReportDocService
	reportExecutionService *services.ReportExecutionService
	iLog                  logger.Log
}

// NewReportDocController creates a new report document controller
func NewReportDocController(docDB documents.DocumentDB, schemaMetadataService *services.SchemaMetadataService) *ReportDocController {
	return &ReportDocController{
		reportDocService:       services.NewReportDocService(docDB),
		reportExecutionService: services.NewReportExecutionService(schemaMetadataService),
		iLog:                   logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "ReportDocController"},
	}
}

// CreateReport handles POST /api/reportdocs
// @Summary Create a new report document
// @Description Creates a new report configuration document in MongoDB
// @Tags ReportDocuments
// @Accept json
// @Produce json
// @Param report body schema.ReportDocument true "Report document"
// @Success 201 {object} map[string]interface{} "Report created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs [post]
func (c *ReportDocController) CreateReport(ctx *gin.Context) {
	startTime := time.Now()
	c.iLog.Debug("CreateReport endpoint called")

	var report schema.ReportDocument
	if err := ctx.ShouldBindJSON(&report); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to bind JSON: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get user from context (assuming authentication middleware sets this)
	userID := ctx.GetString("userID")
	if userID == "" {
		userID = "system" // Default if not set
	}

	// Create the report
	id, err := c.reportDocService.CreateReport(ctx.Request.Context(), &report, userID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to create report: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create report", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("Report created successfully with ID: %s in %v", id, time.Since(startTime)))

	// Fetch the created report to return complete document with ID
	createdReport, err := c.reportDocService.GetReportByID(ctx.Request.Context(), id)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to fetch created report: %v", err))
		// Still return success with ID if we can't fetch
		ctx.JSON(http.StatusCreated, gin.H{
			"message": "Report created successfully",
			"id":      id,
		})
		return
	}

	ctx.JSON(http.StatusCreated, createdReport)
}

// GetReportByID handles GET /api/reportdocs/:id
// @Summary Get report by ID
// @Description Retrieves a report document by its ID
// @Tags ReportDocuments
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} schema.ReportDocument
// @Failure 404 {object} map[string]interface{} "Report not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/{id} [get]
func (c *ReportDocController) GetReportByID(ctx *gin.Context) {
	startTime := time.Now()
	id := ctx.Param("id")
	c.iLog.Debug(fmt.Sprintf("GetReportByID endpoint called for ID: %s", id))

	report, err := c.reportDocService.GetReportByID(ctx.Request.Context(), id)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to get report: %v", err))
		if err == documents.ErrDocumentNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve report", "details": err.Error()})
		}
		return
	}

	c.iLog.Debug(fmt.Sprintf("GetReportByID completed in %v", time.Since(startTime)))
	ctx.JSON(http.StatusOK, report)
}

// GetReportByName handles GET /api/reportdocs/name/:name
// @Summary Get report by name
// @Description Retrieves a report document by its name
// @Tags ReportDocuments
// @Produce json
// @Param name path string true "Report name"
// @Success 200 {object} schema.ReportDocument
// @Failure 404 {object} map[string]interface{} "Report not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/name/{name} [get]
func (c *ReportDocController) GetReportByName(ctx *gin.Context) {
	startTime := time.Now()
	name := ctx.Param("name")
	c.iLog.Debug(fmt.Sprintf("GetReportByName endpoint called for name: %s", name))

	report, err := c.reportDocService.GetReportByName(ctx.Request.Context(), name)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to get report: %v", err))
		if err == documents.ErrDocumentNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve report", "details": err.Error()})
		}
		return
	}

	c.iLog.Debug(fmt.Sprintf("GetReportByName completed in %v", time.Since(startTime)))
	ctx.JSON(http.StatusOK, report)
}

// ListReports handles GET /api/reportdocs
// @Summary List reports
// @Description Retrieves a list of reports with filtering and pagination
// @Tags ReportDocuments
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param pagesize query int false "Page size (default: 20)"
// @Param ispublic query bool false "Filter by public reports"
// @Param category query string false "Filter by category"
// @Param reporttype query string false "Filter by report type"
// @Success 200 {object} map[string]interface{} "List of reports"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs [get]
func (c *ReportDocController) ListReports(ctx *gin.Context) {
	startTime := time.Now()
	c.iLog.Debug("ListReports endpoint called")

	// Parse query parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pagesize", "20"))
	isPublic, _ := strconv.ParseBool(ctx.DefaultQuery("ispublic", "false"))
	category := ctx.Query("category")
	reportType := ctx.Query("reporttype")

	// Get user from context
	userID := ctx.GetString("userID")
	if userID == "" {
		userID = "system"
	}

	// List reports
	reports, total, err := c.reportDocService.ListReports(ctx.Request.Context(), userID, isPublic, category, reportType, page, pageSize)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to list reports: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list reports", "details": err.Error()})
		return
	}

	c.iLog.Debug(fmt.Sprintf("ListReports completed in %v, found %d reports", time.Since(startTime), len(reports)))
	ctx.JSON(http.StatusOK, gin.H{
		"reports":  reports,
		"total":    total,
		"page":     page,
		"pagesize": pageSize,
	})
}

// UpdateReport handles PUT /api/reportdocs/:id
// @Summary Update a report
// @Description Updates an existing report document
// @Tags ReportDocuments
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param updates body map[string]interface{} true "Report updates"
// @Success 200 {object} map[string]interface{} "Report updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/{id} [put]
func (c *ReportDocController) UpdateReport(ctx *gin.Context) {
	startTime := time.Now()
	id := ctx.Param("id")
	c.iLog.Debug(fmt.Sprintf("UpdateReport endpoint called for ID: %s", id))

	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to bind JSON: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get user from context
	userID := ctx.GetString("userID")
	if userID == "" {
		userID = "system"
	}

	// Update the report
	err := c.reportDocService.UpdateReport(ctx.Request.Context(), id, updates, userID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to update report: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update report", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("Report updated successfully: %s in %v", id, time.Since(startTime)))

	// Fetch the updated report to return complete document
	updatedReport, err := c.reportDocService.GetReportByID(ctx.Request.Context(), id)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to fetch updated report: %v", err))
		// Still return success message if we can't fetch
		ctx.JSON(http.StatusOK, gin.H{"message": "Report updated successfully"})
		return
	}

	ctx.JSON(http.StatusOK, updatedReport)
}

// DeleteReport handles DELETE /api/reportdocs/:id
// @Summary Delete a report
// @Description Soft deletes a report document
// @Tags ReportDocuments
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} map[string]interface{} "Report deleted successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/{id} [delete]
func (c *ReportDocController) DeleteReport(ctx *gin.Context) {
	startTime := time.Now()
	id := ctx.Param("id")
	c.iLog.Debug(fmt.Sprintf("DeleteReport endpoint called for ID: %s", id))

	// Get user from context
	userID := ctx.GetString("userID")
	if userID == "" {
		userID = "system"
	}

	// Delete the report
	err := c.reportDocService.DeleteReport(ctx.Request.Context(), id, userID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to delete report: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete report", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("Report deleted successfully: %s in %v", id, time.Since(startTime)))
	ctx.JSON(http.StatusOK, gin.H{"message": "Report deleted successfully"})
}

// AddDatasource handles POST /api/reportdocs/:id/datasources
// @Summary Add a datasource to a report
// @Description Adds a new datasource to an existing report
// @Tags ReportDocuments
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param datasource body schema.ReportDatasourceDoc true "Datasource configuration"
// @Success 200 {object} map[string]interface{} "Datasource added successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/{id}/datasources [post]
func (c *ReportDocController) AddDatasource(ctx *gin.Context) {
	startTime := time.Now()
	reportID := ctx.Param("id")
	c.iLog.Debug(fmt.Sprintf("AddDatasource endpoint called for report ID: %s", reportID))

	var datasource schema.ReportDatasourceDoc
	if err := ctx.ShouldBindJSON(&datasource); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to bind JSON: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Generate ID if not provided
	if datasource.ID == "" {
		datasource.ID = uuid.New().String()
	}

	// Get user from context
	userID := ctx.GetString("userID")
	if userID == "" {
		userID = "system"
	}

	// Add datasource
	err := c.reportDocService.AddDatasource(ctx.Request.Context(), reportID, datasource, userID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to add datasource: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add datasource", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("Datasource added successfully to report: %s in %v", reportID, time.Since(startTime)))
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Datasource added successfully",
		"id":      datasource.ID,
	})
}

// AddComponent handles POST /api/reportdocs/:id/components
// @Summary Add a component to a report
// @Description Adds a new component to an existing report
// @Tags ReportDocuments
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param component body schema.ReportComponentDoc true "Component configuration"
// @Success 200 {object} map[string]interface{} "Component added successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/{id}/components [post]
func (c *ReportDocController) AddComponent(ctx *gin.Context) {
	startTime := time.Now()
	reportID := ctx.Param("id")
	c.iLog.Debug(fmt.Sprintf("AddComponent endpoint called for report ID: %s", reportID))

	var component schema.ReportComponentDoc
	if err := ctx.ShouldBindJSON(&component); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to bind JSON: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Generate ID if not provided
	if component.ID == "" {
		component.ID = uuid.New().String()
	}

	// Get user from context
	userID := ctx.GetString("userID")
	if userID == "" {
		userID = "system"
	}

	// Add component
	err := c.reportDocService.AddComponent(ctx.Request.Context(), reportID, component, userID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to add component: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add component", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("Component added successfully to report: %s in %v", reportID, time.Since(startTime)))
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Component added successfully",
		"id":      component.ID,
	})
}

// AddParameter handles POST /api/reportdocs/:id/parameters
// @Summary Add a parameter to a report
// @Description Adds a new parameter to an existing report
// @Tags ReportDocuments
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param parameter body schema.ReportParameterDoc true "Parameter configuration"
// @Success 200 {object} map[string]interface{} "Parameter added successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/{id}/parameters [post]
func (c *ReportDocController) AddParameter(ctx *gin.Context) {
	startTime := time.Now()
	reportID := ctx.Param("id")
	c.iLog.Debug(fmt.Sprintf("AddParameter endpoint called for report ID: %s", reportID))

	var parameter schema.ReportParameterDoc
	if err := ctx.ShouldBindJSON(&parameter); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to bind JSON: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Generate ID if not provided
	if parameter.ID == "" {
		parameter.ID = uuid.New().String()
	}

	// Get user from context
	userID := ctx.GetString("userID")
	if userID == "" {
		userID = "system"
	}

	// Add parameter
	err := c.reportDocService.AddParameter(ctx.Request.Context(), reportID, parameter, userID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to add parameter: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add parameter", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("Parameter added successfully to report: %s in %v", reportID, time.Since(startTime)))
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Parameter added successfully",
		"id":      parameter.ID,
	})
}

// AddExecution handles POST /api/reportdocs/:id/executions
// @Summary Record a report execution
// @Description Records a report execution in the report's history
// @Tags ReportDocuments
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param execution body schema.ReportExecutionDoc true "Execution details"
// @Success 200 {object} map[string]interface{} "Execution recorded successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/{id}/executions [post]
func (c *ReportDocController) AddExecution(ctx *gin.Context) {
	startTime := time.Now()
	reportID := ctx.Param("id")
	c.iLog.Debug(fmt.Sprintf("AddExecution endpoint called for report ID: %s", reportID))

	var execution schema.ReportExecutionDoc
	if err := ctx.ShouldBindJSON(&execution); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to bind JSON: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Generate ID if not provided
	if execution.ID == "" {
		execution.ID = uuid.New().String()
	}

	// Add execution
	err := c.reportDocService.AddExecution(ctx.Request.Context(), reportID, execution)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to add execution: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record execution", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("Execution recorded successfully for report: %s in %v", reportID, time.Since(startTime)))
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Execution recorded successfully",
		"id":      execution.ID,
	})
}

// SearchReports handles GET /api/reportdocs/search
// @Summary Search reports
// @Description Searches reports by name or description
// @Tags ReportDocuments
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number (default: 1)"
// @Param pagesize query int false "Page size (default: 20)"
// @Success 200 {object} map[string]interface{} "Search results"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/search [get]
func (c *ReportDocController) SearchReports(ctx *gin.Context) {
	startTime := time.Now()
	searchText := ctx.Query("q")
	c.iLog.Debug(fmt.Sprintf("SearchReports endpoint called with query: %s", searchText))

	if searchText == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pagesize", "20"))

	// Search reports
	reports, total, err := c.reportDocService.SearchReports(ctx.Request.Context(), searchText, page, pageSize)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to search reports: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search reports", "details": err.Error()})
		return
	}

	c.iLog.Debug(fmt.Sprintf("SearchReports completed in %v, found %d reports", time.Since(startTime), len(reports)))
	ctx.JSON(http.StatusOK, gin.H{
		"reports":  reports,
		"total":    total,
		"page":     page,
		"pagesize": pageSize,
		"query":    searchText,
	})
}

// GetReportsByCategory handles GET /api/reportdocs/category/:category
// @Summary Get reports by category
// @Description Retrieves reports filtered by category
// @Tags ReportDocuments
// @Produce json
// @Param category path string true "Category name"
// @Param page query int false "Page number (default: 1)"
// @Param pagesize query int false "Page size (default: 20)"
// @Success 200 {object} map[string]interface{} "List of reports"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/category/{category} [get]
func (c *ReportDocController) GetReportsByCategory(ctx *gin.Context) {
	startTime := time.Now()
	category := ctx.Param("category")
	c.iLog.Debug(fmt.Sprintf("GetReportsByCategory endpoint called for category: %s", category))

	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pagesize", "20"))

	// Get reports by category
	reports, total, err := c.reportDocService.GetReportsByCategory(ctx.Request.Context(), category, page, pageSize)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to get reports by category: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reports by category", "details": err.Error()})
		return
	}

	c.iLog.Debug(fmt.Sprintf("GetReportsByCategory completed in %v, found %d reports", time.Since(startTime), len(reports)))
	ctx.JSON(http.StatusOK, gin.H{
		"reports":  reports,
		"total":    total,
		"page":     page,
		"pagesize": pageSize,
		"category": category,
	})
}

// GetDefaultReport handles GET /api/reportdocs/default
// @Summary Get default report
// @Description Retrieves the default report configuration
// @Tags ReportDocuments
// @Produce json
// @Success 200 {object} schema.ReportDocument
// @Failure 404 {object} map[string]interface{} "Default report not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/default [get]
func (c *ReportDocController) GetDefaultReport(ctx *gin.Context) {
	startTime := time.Now()
	c.iLog.Debug("GetDefaultReport endpoint called")

	report, err := c.reportDocService.GetDefaultReport(ctx.Request.Context())
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to get default report: %v", err))
		if err == documents.ErrDocumentNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Default report not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve default report", "details": err.Error()})
		}
		return
	}

	c.iLog.Debug(fmt.Sprintf("GetDefaultReport completed in %v", time.Since(startTime)))
	ctx.JSON(http.StatusOK, report)
}

// SetDefaultReport handles PUT /api/reportdocs/:id/default
// @Summary Set default report
// @Description Sets a report as the default report
// @Tags ReportDocuments
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} map[string]interface{} "Default report set successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/{id}/default [put]
func (c *ReportDocController) SetDefaultReport(ctx *gin.Context) {
	startTime := time.Now()
	id := ctx.Param("id")
	c.iLog.Debug(fmt.Sprintf("SetDefaultReport endpoint called for ID: %s", id))

	// Get user from context
	userID := ctx.GetString("userID")
	if userID == "" {
		userID = "system"
	}

	// Set default report
	err := c.reportDocService.SetDefaultReport(ctx.Request.Context(), id, userID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to set default report: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set default report", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("Default report set successfully: %s in %v", id, time.Since(startTime)))
	ctx.JSON(http.StatusOK, gin.H{"message": "Default report set successfully"})
}

// InitializeIndexes handles POST /api/reportdocs/indexes/init
// @Summary Initialize indexes
// @Description Creates required indexes for the Reports collection
// @Tags ReportDocuments
// @Produce json
// @Success 200 {object} map[string]interface{} "Indexes initialized successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/indexes/init [post]
func (c *ReportDocController) InitializeIndexes(ctx *gin.Context) {
	startTime := time.Now()
	c.iLog.Debug("InitializeIndexes endpoint called")

	err := c.reportDocService.InitializeIndexes(ctx.Request.Context())
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to initialize indexes: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize indexes", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("Indexes initialized successfully in %v", time.Since(startTime)))
	ctx.JSON(http.StatusOK, gin.H{"message": "Indexes initialized successfully"})
}

// ExecuteReport handles POST /api/reportdocs/:id/execute
// @Summary Execute report
// @Description Execute a report from MongoDB with parameters and return results with datasource data
// @Tags ReportDocuments
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param body body services.ExecutionRequest true "Execution parameters"
// @Success 200 {object} services.ExecutionResult "Execution results with datasource data"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Report not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reportdocs/{id}/execute [post]
func (c *ReportDocController) ExecuteReport(ctx *gin.Context) {
	id := ctx.Param("id")
	c.iLog.Info(fmt.Sprintf("ExecuteReport endpoint called for ID: %s", id))

	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	// Get user from context
	userID := ctx.GetString("userID")
	if userID == "" {
		userID = "system"
	}

	// Parse request body
	var execRequest services.ExecutionRequest
	if err := ctx.ShouldBindJSON(&execRequest); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Set default output format if not specified
	if execRequest.OutputFormat == "" {
		execRequest.OutputFormat = "json"
	}

	// Initialize parameters map if nil
	if execRequest.Parameters == nil {
		execRequest.Parameters = make(map[string]interface{})
	}

	// Get report from MongoDB
	report, err := c.reportDocService.GetReportByID(ctx.Request.Context(), id)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to get report: %v", err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Report not found", "details": err.Error()})
		return
	}

	c.iLog.Debug(fmt.Sprintf("Executing report %s with %d datasources", report.Name, len(report.Datasources)))

	// Log datasource details for debugging
	if len(report.Datasources) == 0 {
		c.iLog.Warn(fmt.Sprintf("Report %s (%s) has NO datasources configured in MongoDB", report.Name, report.ID))
	} else {
		for i, ds := range report.Datasources {
			c.iLog.Debug(fmt.Sprintf("  DS[%d]: Alias='%s', Active=%v, Type='%s', DB='%s', HasSQL=%v",
				i, ds.Alias, ds.Active, ds.QueryType, ds.DatabaseAlias, len(ds.CustomSQL) > 0))
		}
	}

	// Execute the report using the execution service
	result, err := c.reportExecutionService.ExecuteReport(
		ctx.Request.Context(),
		report,
		execRequest,
		userID,
	)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to execute report: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to execute report",
			"details": err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("Report %s executed successfully in %dms (Status: %s, Datasources: %d)",
		id, result.ExecutionTimeMs, result.Status, len(result.Data.Datasources)))

	ctx.JSON(http.StatusOK, result)
}

// RegisterRoutes registers all report document routes
func (c *ReportDocController) RegisterRoutes(router *gin.RouterGroup) {
	reportDocs := router.Group("/reportdocs")
	{
		// Collection operations
		reportDocs.POST("", c.CreateReport)
		reportDocs.GET("", c.ListReports)
		reportDocs.GET("/search", c.SearchReports)
		reportDocs.GET("/default", c.GetDefaultReport)
		reportDocs.POST("/indexes/init", c.InitializeIndexes)

		// Category operations
		reportDocs.GET("/category/:category", c.GetReportsByCategory)

		// Name-based lookup
		reportDocs.GET("/name/:name", c.GetReportByName)

		// Document operations
		reportDocs.GET("/:id", c.GetReportByID)
		reportDocs.PUT("/:id", c.UpdateReport)
		reportDocs.DELETE("/:id", c.DeleteReport)
		reportDocs.PUT("/:id/default", c.SetDefaultReport)

		// Sub-document operations
		reportDocs.POST("/:id/datasources", c.AddDatasource)
		reportDocs.POST("/:id/components", c.AddComponent)
		reportDocs.POST("/:id/parameters", c.AddParameter)
		reportDocs.POST("/:id/executions", c.AddExecution)
	}

	c.iLog.Info("Report document routes registered successfully")
}
