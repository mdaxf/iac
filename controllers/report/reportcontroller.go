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
		"reports":   reports,
		"total":     total,
		"page":      page,
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

	// Handle datasources separately if included in the request
	if datasourcesRaw, hasDatasources := updates["datasources"]; hasDatasources {
		if datasourcesArray, ok := datasourcesRaw.([]interface{}); ok {
			// Collect IDs of datasources in the incoming request
			incomingDatasourceIDs := make(map[string]bool)

			for _, dsRaw := range datasourcesArray {
				if dsMap, ok := dsRaw.(map[string]interface{}); ok {
					// Extract datasource ID to determine if this is update or create
					dsID, _ := dsMap["id"].(string)

					// Track this datasource ID
					if dsID != "" {
						incomingDatasourceIDs[dsID] = true
					}

					if dsID != "" {
						// Update existing datasource - work directly with the map
						dsUpdates := make(map[string]interface{})

						// Copy scalar fields
						if v, ok := dsMap["alias"]; ok {
							dsUpdates["alias"] = v
						}
						if v, ok := dsMap["databasealias"]; ok {
							dsUpdates["databasealias"] = v
						}
						if v, ok := dsMap["querytype"]; ok {
							dsUpdates["querytype"] = v
						}
						if v, ok := dsMap["customsql"]; ok {
							dsUpdates["customsql"] = v
						}

						// Copy JSON fields (arrays/objects) as-is - they'll be serialized by GORM
						if v, ok := dsMap["selectedtables"]; ok {
							dsUpdates["selectedtables"] = v
						}
						if v, ok := dsMap["selectedfields"]; ok {
							dsUpdates["selectedfields"] = v
						}
						if v, ok := dsMap["joins"]; ok {
							dsUpdates["joins"] = v
						}
						if v, ok := dsMap["filters"]; ok {
							dsUpdates["filters"] = v
						}
						if v, ok := dsMap["sorting"]; ok {
							dsUpdates["sorting"] = v
						}
						if v, ok := dsMap["grouping"]; ok {
							dsUpdates["grouping"] = v
						}
						if v, ok := dsMap["parameters"]; ok {
							dsUpdates["parameters"] = v
						}

						dsUpdates["modifiedby"] = user
						dsUpdates["reportid"] = reportID

						if err := rc.service.UpdateDatasource(dsID, dsUpdates); err != nil {
							iLog.Error(fmt.Sprintf("Error updating datasource %s: %v", dsID, err))
						}
					} else {
						// Create new datasource
						// For create, we build a minimal struct and let GORM handle the JSON fields
						datasource := &models.ReportDatasource{
							ReportID:   reportID,
							CreatedBy:  user,
							ModifiedBy: user,
						}

						// Set scalar fields if present
						if v, ok := dsMap["alias"].(string); ok {
							datasource.Alias = v
						}
						if v, ok := dsMap["databasealias"].(string); ok {
							datasource.DatabaseAlias = v
						}
						if v, ok := dsMap["querytype"].(string); ok {
							datasource.QueryType = v
						}
						if v, ok := dsMap["customsql"].(string); ok {
							datasource.CustomSQL = v
						}

						if err := rc.service.AddDatasource(datasource); err != nil {
							iLog.Error(fmt.Sprintf("Error creating datasource: %v", err))
							continue
						}

						// After creation, update the complex JSON fields using the update path
						if datasource.ID != "" {
							dsUpdates := make(map[string]interface{})

							if v, ok := dsMap["selectedtables"]; ok {
								dsUpdates["selectedtables"] = v
							}
							if v, ok := dsMap["selectedfields"]; ok {
								dsUpdates["selectedfields"] = v
							}
							if v, ok := dsMap["joins"]; ok {
								dsUpdates["joins"] = v
							}
							if v, ok := dsMap["filters"]; ok {
								dsUpdates["filters"] = v
							}
							if v, ok := dsMap["sorting"]; ok {
								dsUpdates["sorting"] = v
							}
							if v, ok := dsMap["grouping"]; ok {
								dsUpdates["grouping"] = v
							}
							if v, ok := dsMap["parameters"]; ok {
								dsUpdates["parameters"] = v
							}

							if len(dsUpdates) > 0 {
								if err := rc.service.UpdateDatasource(datasource.ID, dsUpdates); err != nil {
									iLog.Error(fmt.Sprintf("Error updating new datasource JSON fields %s: %v", datasource.ID, err))
								}
							}
						}
					}
				}
			}

			// Delete datasources that are in the database but not in the incoming request
			existingDatasources, err := rc.service.GetDatasources(reportID)
			if err != nil {
				iLog.Error(fmt.Sprintf("Error getting existing datasources: %v", err))
			} else {
				for _, existingDs := range existingDatasources {
					// If this datasource ID is not in the incoming request, delete it
					if !incomingDatasourceIDs[existingDs.ID] {
						if err := rc.service.DeleteDatasource(existingDs.ID); err != nil {
							iLog.Error(fmt.Sprintf("Error deleting datasource %s: %v", existingDs.ID, err))
						} else {
							iLog.Debug(fmt.Sprintf("Deleted datasource %s that was removed from report", existingDs.ID))
						}
					}
				}
			}
		}
	}

	// Handle components separately if included in the request
	if componentsRaw, hasComponents := updates["components"]; hasComponents {
		if componentsArray, ok := componentsRaw.([]interface{}); ok {
			// Collect IDs of components in the incoming request
			incomingComponentIDs := make(map[string]bool)

			for _, compRaw := range componentsArray {
				if compMap, ok := compRaw.(map[string]interface{}); ok {
					// Extract component ID to determine if this is update or create
					compID, _ := compMap["id"].(string)
					componentType, _ := compMap["componenttype"].(string)

					// Track this component ID
					if compID != "" {
						incomingComponentIDs[compID] = true
					}

					if compID != "" {
						// Update existing component - work directly with the map
						compUpdates := make(map[string]interface{})

						// Copy common scalar fields
						if v, ok := compMap["componenttype"]; ok {
							compUpdates["componenttype"] = v
						}
						if v, ok := compMap["name"]; ok {
							compUpdates["name"] = v
						}
						if v, ok := compMap["x"]; ok {
							compUpdates["x"] = v
						}
						if v, ok := compMap["y"]; ok {
							compUpdates["y"] = v
						}
						if v, ok := compMap["width"]; ok {
							compUpdates["width"] = v
						}
						if v, ok := compMap["height"]; ok {
							compUpdates["height"] = v
						}
						if v, ok := compMap["zindex"]; ok {
							compUpdates["zindex"] = v
						}
						if v, ok := compMap["datasourcealias"]; ok {
							compUpdates["datasourcealias"] = v
						}
						if v, ok := compMap["isvisible"]; ok {
							compUpdates["isvisible"] = v
						}

						// Copy type-specific fields based on componenttype
						if componentType == "chart" {
							if v, ok := compMap["charttype"]; ok {
								compUpdates["charttype"] = v
							}
							if v, ok := compMap["chartconfig"]; ok {
								compUpdates["chartconfig"] = v
							}
						}
						if componentType == "barcode" {
							if v, ok := compMap["barcodetype"]; ok {
								compUpdates["barcodetype"] = v
							}
							if v, ok := compMap["barcodeconfig"]; ok {
								compUpdates["barcodeconfig"] = v
							}
						}
						if componentType == "drill_down" {
							if v, ok := compMap["drilldownconfig"]; ok {
								compUpdates["drilldownconfig"] = v
							}
						}

						// Copy common JSON fields (arrays/objects) as-is - they'll be serialized by UpdateComponent
						if v, ok := compMap["dataconfig"]; ok {
							compUpdates["dataconfig"] = v
						}
						if v, ok := compMap["componentconfig"]; ok {
							compUpdates["componentconfig"] = v
						}
						if v, ok := compMap["styleconfig"]; ok {
							compUpdates["styleconfig"] = v
						}
						if v, ok := compMap["conditionalformatting"]; ok {
							compUpdates["conditionalformatting"] = v
						}

						compUpdates["modifiedby"] = user
						compUpdates["reportid"] = reportID

						if err := rc.service.UpdateComponent(compID, compUpdates); err != nil {
							iLog.Error(fmt.Sprintf("Error updating component %s: %v", compID, err))
						}
					} else {
						// Create new component
						// For create, we build a minimal struct and let GORM handle the JSON fields
						component := &models.ReportComponent{
							ReportID:   reportID,
							CreatedBy:  user,
							ModifiedBy: user,
						}

						// Set scalar fields if present
						if v, ok := compMap["componenttype"].(string); ok {
							component.ComponentType = models.ComponentType(v)
						}
						if v, ok := compMap["name"].(string); ok {
							component.Name = v
						}
						if v, ok := compMap["x"].(float64); ok {
							component.X = v
						}
						if v, ok := compMap["y"].(float64); ok {
							component.Y = v
						}
						if v, ok := compMap["width"].(float64); ok {
							component.Width = v
						}
						if v, ok := compMap["height"].(float64); ok {
							component.Height = v
						}
						if v, ok := compMap["zindex"].(float64); ok {
							component.ZIndex = int(v)
						}
						if v, ok := compMap["datasourcealias"].(string); ok {
							component.DatasourceAlias = v
						}
						if v, ok := compMap["isvisible"].(bool); ok {
							component.IsVisible = v
						}

						if err := rc.service.AddComponent(component); err != nil {
							iLog.Error(fmt.Sprintf("Error creating component: %v", err))
							continue
						}

						// After creation, update the complex JSON fields using the update path
						if component.ID != "" {
							compUpdates := make(map[string]interface{})

							// Common JSON fields for all component types
							if v, ok := compMap["dataconfig"]; ok {
								compUpdates["dataconfig"] = v
							}
							if v, ok := compMap["componentconfig"]; ok {
								compUpdates["componentconfig"] = v
							}
							if v, ok := compMap["styleconfig"]; ok {
								compUpdates["styleconfig"] = v
							}
							if v, ok := compMap["conditionalformatting"]; ok {
								compUpdates["conditionalformatting"] = v
							}

							// Type-specific fields based on componenttype
							if componentType == "chart" {
								if v, ok := compMap["chartconfig"]; ok {
									compUpdates["chartconfig"] = v
								}
							}
							if componentType == "barcode" {
								if v, ok := compMap["barcodeconfig"]; ok {
									compUpdates["barcodeconfig"] = v
								}
							}
							if componentType == "drill_down" {
								if v, ok := compMap["drilldownconfig"]; ok {
									compUpdates["drilldownconfig"] = v
								}
							}

							if len(compUpdates) > 0 {
								if err := rc.service.UpdateComponent(component.ID, compUpdates); err != nil {
									iLog.Error(fmt.Sprintf("Error updating new component JSON fields %s: %v", component.ID, err))
								}
							}
						}
					}
				}
			}

			// Delete components that are in the database but not in the incoming request
			existingComponents, err := rc.service.GetComponents(reportID)
			if err != nil {
				iLog.Error(fmt.Sprintf("Error getting existing components: %v", err))
			} else {
				for _, existingComp := range existingComponents {
					// If this component ID is not in the incoming request, delete it
					if !incomingComponentIDs[existingComp.ID] {
						if err := rc.service.DeleteComponent(existingComp.ID); err != nil {
							iLog.Error(fmt.Sprintf("Error deleting component %s: %v", existingComp.ID, err))
						} else {
							iLog.Debug(fmt.Sprintf("Deleted component %s that was removed from report", existingComp.ID))
						}
					}
				}
			}
		}
	}

	// Update the report (UpdateReport will filter out relationship fields)
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

// UpdateDatasourceEndpoint handles PUT /:id/datasources/:datasourceId - Update a datasource
func (rc *ReportController) UpdateDatasourceEndpoint(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("report.UpdateDatasourceEndpoint", elapsed)
	}()

	reportID := c.Param("id")
	datasourceID := c.Param("datasourceId")

	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}
	if datasourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datasource ID is required"})
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

	// Verify report ownership
	report, err := rc.service.GetReportByID(reportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	if report.CreatedBy != user {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this report's datasources"})
		return
	}

	// Ensure reportid and modifiedby are set correctly
	updates["reportid"] = reportID
	updates["modifiedby"] = user

	if err := rc.service.UpdateDatasource(datasourceID, updates); err != nil {
		iLog.Error(fmt.Sprintf("Error updating datasource: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update datasource"})
		return
	}

	// Fetch updated datasources
	datasources, _ := rc.service.GetDatasources(reportID)
	c.JSON(http.StatusOK, gin.H{
		"message":     "Datasource updated successfully",
		"datasources": datasources,
	})
}

// DeleteDatasourceEndpoint handles DELETE /:id/datasources/:datasourceId - Delete a datasource
func (rc *ReportController) DeleteDatasourceEndpoint(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "report"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("report.DeleteDatasourceEndpoint", elapsed)
	}()

	reportID := c.Param("id")
	datasourceID := c.Param("datasourceId")

	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}
	if datasourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datasource ID is required"})
		return
	}

	_, _, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify report ownership
	report, err := rc.service.GetReportByID(reportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	if report.CreatedBy != user {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this report's datasources"})
		return
	}

	if err := rc.service.DeleteDatasource(datasourceID); err != nil {
		iLog.Error(fmt.Sprintf("Error deleting datasource: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete datasource"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Datasource deleted successfully"})
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

	// Execute the report query
	result, err := rc.service.ExecuteReportQuery(reportID, execRequest.Parameters)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error executing report: %v", err))

		// Update execution record with error
		elapsed := time.Since(startTime)
		rc.service.UpdateExecution(execution.ID, map[string]interface{}{
			"executionstatus": "failed",
			"executiontimems": int(elapsed.Milliseconds()),
			"errormessage":    err.Error(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{
			"execution_id": execution.ID,
			"status":       "failed",
			"error":        err.Error(),
		})
		return
	}

	// Calculate result size and row count
	totalRows := 0
	if datasources, ok := result["datasources"].(map[string]interface{}); ok {
		for _, dsResult := range datasources {
			if dsMap, ok := dsResult.(map[string]interface{}); ok {
				if rows, ok := dsMap["totalRows"].(int); ok {
					totalRows += rows
				}
			}
		}
	}

	// Update execution record with success
	elapsed := time.Since(startTime)
	rc.service.UpdateExecution(execution.ID, map[string]interface{}{
		"executionstatus": "success",
		"executiontimems": int(elapsed.Milliseconds()),
		"rowcount":        totalRows,
	})

	rc.service.UpdateLastExecutedAt(reportID)

	// Return execution result with data
	c.JSON(http.StatusOK, gin.H{
		"execution_id": execution.ID,
		"status":       "success",
		"message":      "Report executed successfully",
		"data":         result,
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
