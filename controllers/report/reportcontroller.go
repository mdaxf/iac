package report

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/gormdb"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// ReportController handles report-related HTTP requests
type ReportController struct {
	service   *services.ReportService
	aiService *services.AIReportService
}

// NewReportController creates a new report controller
func NewReportController() *ReportController {
	return &ReportController{
		service:   services.NewReportService(gormdb.DB),
		aiService: services.NewAIReportService(config.OpenAiKey, config.OpenAiModel),
	}
}

// CreateReport handles POST / - Create new report
func (rc *ReportController) CreateReport(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("report.CreateReport", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var report models.Report
	if err := json.Unmarshal(body, &report); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling report: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set created by
	report.CreatedBy = user

	if err := rc.service.CreateReport(&report); err != nil {
		iLog.Error(fmt.Sprintf("Error creating report: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create report"})
		return
	}

	iLog.Info(fmt.Sprintf("Report created successfully: %s", report.ID))
	c.JSON(http.StatusCreated, report)
}

// GetReport handles GET /:id - Get report by ID
func (rc *ReportController) GetReport(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("report.GetReport", elapsed)
	}()

	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	report, err := rc.service.GetReportByID(reportID)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error fetching report: %v", err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// ListReports handles GET / - List reports with pagination
func (rc *ReportController) ListReports(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("report.ListReports", elapsed)
	}()

	_, _, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	reportType := c.Query("type")
	isPublic := c.Query("public") == "true"

	reports, total, err := rc.service.ListReports(user, isPublic, reportType, page, pageSize)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error listing reports: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reports"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  reports,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// UpdateReport handles PUT /:id - Update report
func (rc *ReportController) UpdateReport(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("report.UpdateReport", elapsed)
	}()

	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	body, _, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updates map[string]interface{}
	if err := json.Unmarshal(body, &updates); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling updates: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verify ownership or public access
	report, err := rc.service.GetReportByID(reportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	if report.CreatedBy != user {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this report"})
		return
	}

	if err := rc.service.UpdateReport(reportID, updates); err != nil {
		iLog.Error(fmt.Sprintf("Error updating report: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update report"})
		return
	}

	// Fetch updated report
	updatedReport, _ := rc.service.GetReportByID(reportID)
	c.JSON(http.StatusOK, updatedReport)
}

// DeleteReport handles DELETE /:id - Delete report
func (rc *ReportController) DeleteReport(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("report.DeleteReport", elapsed)
	}()

	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	_, _, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify ownership
	report, err := rc.service.GetReportByID(reportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	if report.CreatedBy != user {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this report"})
		return
	}

	if err := rc.service.DeleteReport(reportID); err != nil {
		iLog.Error(fmt.Sprintf("Error deleting report: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Report deleted successfully"})
}

// AddDatasource handles POST /:id/datasources - Add datasource to report
func (rc *ReportController) AddDatasource(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}

	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	body, _, _, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var datasource models.ReportDatasource
	if err := json.Unmarshal(body, &datasource); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling datasource: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	datasource.ReportID = reportID

	if err := rc.service.AddDatasource(&datasource); err != nil {
		iLog.Error(fmt.Sprintf("Error adding datasource: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add datasource"})
		return
	}

	c.JSON(http.StatusCreated, datasource)
}

// GetDatasources handles GET /:id/datasources - Get report datasources
func (rc *ReportController) GetDatasources(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	datasources, err := rc.service.GetDatasources(reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch datasources"})
		return
	}

	c.JSON(http.StatusOK, datasources)
}

// AddComponent handles POST /:id/components - Add component to report
func (rc *ReportController) AddComponent(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}

	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	body, _, _, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var component models.ReportComponent
	if err := json.Unmarshal(body, &component); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling component: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	component.ReportID = reportID

	if err := rc.service.AddComponent(&component); err != nil {
		iLog.Error(fmt.Sprintf("Error adding component: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add component"})
		return
	}

	c.JSON(http.StatusCreated, component)
}

// GetComponents handles GET /:id/components - Get report components
func (rc *ReportController) GetComponents(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	components, err := rc.service.GetComponents(reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch components"})
		return
	}

	c.JSON(http.StatusOK, components)
}

// ExecuteReport handles POST /:id/execute - Execute report with parameters
func (rc *ReportController) ExecuteReport(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}
	startTime := time.Now()

	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	body, _, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var execRequest struct {
		Parameters map[string]interface{} `json:"parameters"`
	}
	if err := json.Unmarshal(body, &execRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create execution record
	execution := &models.ReportExecution{
		ReportID:        reportID,
		ExecutedBy:      user,
		ExecutionStatus: "running",
		Parameters:      execRequest.Parameters,
	}

	if err := rc.service.CreateExecution(execution); err != nil {
		iLog.Error(fmt.Sprintf("Error creating execution: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create execution"})
		return
	}

	// TODO: Implement actual report execution logic here
	// For now, just mark as success
	elapsed := time.Since(startTime)
	rc.service.UpdateExecution(execution.ID, map[string]interface{}{
		"execution_status":  "success",
		"execution_time_ms": int(elapsed.Milliseconds()),
	})

	rc.service.UpdateLastExecutedAt(reportID)

	c.JSON(http.StatusOK, gin.H{
		"execution_id": execution.ID,
		"status":       "success",
		"message":      "Report executed successfully",
	})
}

// ShareReport handles POST /:id/share - Share report with users
func (rc *ReportController) ShareReport(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}

	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	body, _, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var share models.ReportShare
	if err := json.Unmarshal(body, &share); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling share: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	share.ReportID = reportID
	share.SharedBy = user

	if err := rc.service.ShareReport(&share); err != nil {
		iLog.Error(fmt.Sprintf("Error sharing report: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share report"})
		return
	}

	c.JSON(http.StatusCreated, share)
}

// ListTemplates handles GET /templates - List report templates
func (rc *ReportController) ListTemplates(c *gin.Context) {
	category := c.Query("category")
	isPublic := c.Query("public") != "false"

	templates, err := rc.service.ListTemplates(category, isPublic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch templates"})
		return
	}

	c.JSON(http.StatusOK, templates)
}

// CreateFromTemplate handles POST /from-template/:templateId - Create report from template
func (rc *ReportController) CreateFromTemplate(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}

	templateID := c.Param("templateId")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
		return
	}

	_, _, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := rc.service.CreateFromTemplate(templateID, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error creating from template: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create report from template"})
		return
	}

	c.JSON(http.StatusCreated, report)
}

// DuplicateReport handles POST /:id/duplicate - Duplicate an existing report
func (rc *ReportController) DuplicateReport(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}

	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	_, _, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := rc.service.DuplicateReport(reportID, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error duplicating report: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to duplicate report"})
		return
	}

	c.JSON(http.StatusCreated, report)
}

// SearchReports handles GET /search - Search reports by keyword
func (rc *ReportController) SearchReports(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search keyword is required"})
		return
	}

	_, _, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	reports, err := rc.service.SearchReports(keyword, user, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search reports"})
		return
	}

	c.JSON(http.StatusOK, reports)
}

// Text2SQL handles POST /ai/text2sql - Convert natural language to SQL
func (rc *ReportController) Text2SQL(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("report.Text2SQL", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var request services.Text2SQLRequest
	if err := json.Unmarshal(body, &request); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: Get actual schema information from the database
	// For now, use a placeholder
	schemaInfo := "Schema information will be retrieved from database metadata"

	response, err := rc.aiService.GenerateSQL(c.Request.Context(), request, schemaInfo)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error generating SQL: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate SQL"})
		return
	}

	iLog.Info(fmt.Sprintf("SQL generated successfully with confidence %.2f", response.Confidence))
	c.JSON(http.StatusOK, response)
}

// GenerateReportFromData handles POST /ai/generate-report - Generate report structure from data
func (rc *ReportController) GenerateReportFromData(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("report.GenerateReportFromData", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var request services.ReportGenerationRequest
	if err := json.Unmarshal(body, &request); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	response, err := rc.aiService.GenerateReport(c.Request.Context(), request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error generating report: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report"})
		return
	}

	iLog.Info(fmt.Sprintf("Report structure generated successfully: %s", response.Title))
	c.JSON(http.StatusOK, response)
}
