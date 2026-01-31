// Copyright 2023 IAC. All Rights Reserved.

package planscheduler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/gormdb"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// PlanSchedulerProfileController handles plan scheduler profile HTTP requests
type PlanSchedulerProfileController struct {
	iLog logger.Log
}

// NewPlanSchedulerProfileController creates a new controller instance
func NewPlanSchedulerProfileController() *PlanSchedulerProfileController {
	return &PlanSchedulerProfileController{
		iLog: logger.Log{
			ModuleName:     logger.API,
			User:           "System",
			ControllerName: "PlanSchedulerProfileController",
		},
	}
}

// GetAllProfiles handles GET /api/plan-scheduler/profiles
func (c *PlanSchedulerProfileController) GetAllProfiles(ctx *gin.Context) {
	c.iLog.Info("GetAllProfiles START")

	// Get user from context
	user := c.getUserFromContext(ctx)

	// Create service
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	// Get all profiles
	profiles, err := service.GetAllProfiles(ctx.Request.Context())
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to get profiles: %v, %s", err, user))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get profiles",
			"details": err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("GetAllProfiles COMPLETE - Found %d profiles", len(profiles)))
	ctx.JSON(http.StatusOK, gin.H{
		"profiles": profiles,
		"count":    len(profiles),
	})
}

// GetProfile handles GET /api/plan-scheduler/profiles/:id
func (c *PlanSchedulerProfileController) GetProfile(ctx *gin.Context) {
	id := ctx.Param("id")
	c.iLog.Info(fmt.Sprintf("GetProfile START - ID: %s", id))

	// Create service
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	// Get profile
	profile, err := service.GetProfileByID(ctx.Request.Context(), id)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to get profile: %v", err))
		ctx.JSON(http.StatusNotFound, gin.H{
			"error":   "Profile not found",
			"details": err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("GetProfile COMPLETE - Profile: %s", profile.Name))
	ctx.JSON(http.StatusOK, profile)
}

// GetDefaultProfile handles GET /api/plan-scheduler/profiles/default
func (c *PlanSchedulerProfileController) GetDefaultProfile(ctx *gin.Context) {
	c.iLog.Info("GetDefaultProfile START")

	// Create service
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	// Get default profile
	profile, err := service.GetDefaultProfile(ctx.Request.Context())
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to get default profile: %v", err))
		ctx.JSON(http.StatusNotFound, gin.H{
			"error":   "Default profile not found",
			"details": err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("GetDefaultProfile COMPLETE - Profile: %s", profile.Name))
	ctx.JSON(http.StatusOK, profile)
}

// CreateProfile handles POST /api/plan-scheduler/profiles
func (c *PlanSchedulerProfileController) CreateProfile(ctx *gin.Context) {
	c.iLog.Info("CreateProfile START")

	// Get user from context
	user := c.getUserFromContext(ctx)

	// Parse request body
	var profile models.PlanSchedulerProfile
	if err := ctx.ShouldBindJSON(&profile); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate
	if profile.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Profile name is required",
		})
		return
	}

	// Create service
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	// Create profile
	if err := service.CreateProfile(ctx.Request.Context(), &profile, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to create profile: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create profile",
			"details": err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("CreateProfile COMPLETE - ID: %s", profile.ID))
	ctx.JSON(http.StatusCreated, profile)
}

// UpdateProfile handles PUT /api/plan-scheduler/profiles/:id
func (c *PlanSchedulerProfileController) UpdateProfile(ctx *gin.Context) {
	id := ctx.Param("id")
	c.iLog.Info(fmt.Sprintf("UpdateProfile START - ID: %s", id))

	// Get user from context
	user := c.getUserFromContext(ctx)

	// Parse request body
	var profile models.PlanSchedulerProfile
	if err := ctx.ShouldBindJSON(&profile); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Ensure ID matches
	profile.ID = id

	// Create service
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	// Update profile
	if err := service.UpdateProfile(ctx.Request.Context(), &profile, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to update profile: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update profile",
			"details": err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("UpdateProfile COMPLETE - ID: %s", id))
	ctx.JSON(http.StatusOK, profile)
}

// DeleteProfile handles DELETE /api/plan-scheduler/profiles/:id
func (c *PlanSchedulerProfileController) DeleteProfile(ctx *gin.Context) {
	id := ctx.Param("id")
	c.iLog.Info(fmt.Sprintf("DeleteProfile START - ID: %s", id))

	// Get user from context
	user := c.getUserFromContext(ctx)

	// Create service
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	// Delete profile
	if err := service.DeleteProfile(ctx.Request.Context(), id, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to delete profile: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete profile",
			"details": err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("DeleteProfile COMPLETE - ID: %s", id))
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Profile deleted successfully",
		"id":      id,
	})
}

// LoadProfileData handles GET /api/plan-scheduler/profiles/:id/data
func (c *PlanSchedulerProfileController) LoadProfileData(ctx *gin.Context) {
	id := ctx.Param("id")
	c.iLog.Info(fmt.Sprintf("LoadProfileData START - ID: %s", id))

	// Parse query parameters as profile data query parameters
	parameters := make(map[string]interface{})
	for key, values := range ctx.Request.URL.Query() {
		if len(values) > 0 {
			parameters[key] = values[0]
		}
	}

	// Create service
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	// Load profile data
	data, err := service.LoadProfileData(ctx.Request.Context(), id, parameters)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to load profile data: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load profile data",
			"details": err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("LoadProfileData COMPLETE - Loaded %d tasks, %d resources", len(data.Tasks), len(data.Resources)))
	ctx.JSON(http.StatusOK, data)
}

// SaveScheduleData handles POST /api/plan-scheduler/profiles/:id/save
// Smart save logic:
// - If datasource is static JSON (source_type = 'json'): Update source_json field
// - If datasource is query-based (source_type = 'query'): Save as session
func (c *PlanSchedulerProfileController) SaveScheduleData(ctx *gin.Context) {
	profileID := ctx.Param("id")
	c.iLog.Info(fmt.Sprintf("SaveScheduleData START - ProfileID: %s", profileID))

	user := c.getUserFromContext(ctx)

	// Parse request body
	var saveRequest struct {
		Tasks       []map[string]interface{} `json:"tasks"`
		Resources   []map[string]interface{} `json:"resources,omitempty"`
		SessionID   string                   `json:"sessionId,omitempty"`   // For session-based save
		SessionName string                   `json:"sessionName,omitempty"` // For session-based save
		Description string                   `json:"description,omitempty"` // For session-based save
		IsDefault   bool                     `json:"isDefault,omitempty"`   // Whether to set as default session
	}

	if err := ctx.ShouldBindJSON(&saveRequest); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	// Get profile with data sources
	profile, err := service.GetProfileByID(ctx.Request.Context(), profileID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to get profile: %v", err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Profile not found", "details": err.Error()})
		return
	}

	// Determine save strategy based on data sources
	hasStaticJSON := false
	hasQueryBased := false

	for _, ds := range profile.DataSources {
		if ds.SourceType == "json" && ds.DataType == "tasks" {
			hasStaticJSON = true
		} else if ds.SourceType == "query" && ds.DataType == "tasks" {
			hasQueryBased = true
		}
	}

	c.iLog.Info(fmt.Sprintf("Save strategy - Static JSON: %v, Query-based: %v", hasStaticJSON, hasQueryBased))

	// Strategy 1: Update static JSON datasource
	if hasStaticJSON && !hasQueryBased {
		c.iLog.Info("Using Strategy 1: Update static JSON datasource")

		// Find the tasks datasource
		var tasksDS *models.PlanSchedulerProfileDataSource
		for i := range profile.DataSources {
			if profile.DataSources[i].SourceType == "json" && profile.DataSources[i].DataType == "tasks" {
				tasksDS = &profile.DataSources[i]
				break
			}
		}

		if tasksDS == nil {
			c.iLog.Error("No JSON tasks datasource found")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "No static JSON tasks datasource found"})
			return
		}

		// Convert tasks to JSON
		tasksJSON, err := json.Marshal(saveRequest.Tasks)
		if err != nil {
			c.iLog.Error(fmt.Sprintf("Failed to marshal tasks: %v", err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal tasks", "details": err.Error()})
			return
		}

		tasksDS.SourceJSON = json.RawMessage(tasksJSON)

		// Update datasource
		if err := service.UpdateDataSource(ctx.Request.Context(), tasksDS, user); err != nil {
			c.iLog.Error(fmt.Sprintf("Failed to update datasource: %v", err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update datasource", "details": err.Error()})
			return
		}

		c.iLog.Info(fmt.Sprintf("SaveScheduleData COMPLETE - Updated JSON datasource with %d tasks", len(saveRequest.Tasks)))
		ctx.JSON(http.StatusOK, gin.H{
			"message":      "Schedule saved to static JSON datasource",
			"saved":        true,
			"saveType":     "json_datasource",
			"datasourceId": tasksDS.ID,
			"taskCount":    len(saveRequest.Tasks),
		})
		return
	}

	// Strategy 2: Save as session (for query-based or mixed datasources)
	c.iLog.Info("Using Strategy 2: Save as session")

	// Generate session ID if not provided
	sessionID := saveRequest.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%s", time.Now().Format("20060102-150405"))
	}

	// Marshal tasks and resources
	tasksJSON, err := json.Marshal(saveRequest.Tasks)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to marshal tasks: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal tasks", "details": err.Error()})
		return
	}

	var resourcesJSON json.RawMessage
	if saveRequest.Resources != nil {
		resourcesJSON, err = json.Marshal(saveRequest.Resources)
		if err != nil {
			c.iLog.Error(fmt.Sprintf("Failed to marshal resources: %v", err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal resources", "details": err.Error()})
			return
		}
	}

	// Create session object
	session := &models.PlanSchedulerSession{
		ProfileID:     profileID,
		SessionID:     sessionID,
		SessionName:   saveRequest.SessionName,
		Description:   saveRequest.Description,
		TasksData:     json.RawMessage(tasksJSON),
		ResourcesData: resourcesJSON,
		Status:        "active",
		IsDefault:     saveRequest.IsDefault,
	}

	// Check if session already exists
	existingSession, err := service.GetSessionsByProfile(ctx.Request.Context(), profileID)
	if err == nil {
		// Find if this session ID already exists
		for _, s := range existingSession {
			if s.SessionID == sessionID {
				// Update existing session
				session.ID = s.ID
				if err := service.UpdateScheduleSession(ctx.Request.Context(), session, user); err != nil {
					c.iLog.Error(fmt.Sprintf("Failed to update session: %v", err))
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session", "details": err.Error()})
					return
				}

				c.iLog.Info(fmt.Sprintf("SaveScheduleData COMPLETE - Updated session %s", sessionID))
				ctx.JSON(http.StatusOK, gin.H{
					"message":   "Schedule saved to existing session",
					"saved":     true,
					"saveType":  "session_update",
					"sessionId": sessionID,
					"taskCount": len(saveRequest.Tasks),
					"isDefault": saveRequest.IsDefault,
				})
				return
			}
		}
	}

	// Create new session
	if err := service.SaveScheduleAsSession(ctx.Request.Context(), session, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to save session: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("SaveScheduleData COMPLETE - Created new session %s", sessionID))
	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Schedule saved as new session",
		"saved":     true,
		"saveType":  "session_create",
		"sessionId": sessionID,
		"taskCount": len(saveRequest.Tasks),
		"isDefault": saveRequest.IsDefault,
	})
}

// CreateDataSource handles POST /api/plan-scheduler/profiles/:id/datasources
func (c *PlanSchedulerProfileController) CreateDataSource(ctx *gin.Context) {
	profileID := ctx.Param("id")
	c.iLog.Info(fmt.Sprintf("CreateDataSource START - ProfileID: %s", profileID))

	user := c.getUserFromContext(ctx)

	var dataSource models.PlanSchedulerProfileDataSource
	if err := ctx.ShouldBindJSON(&dataSource); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	dataSource.ProfileID = profileID
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	if err := service.CreateDataSource(ctx.Request.Context(), &dataSource, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to create data source: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create data source", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("CreateDataSource COMPLETE - ID: %s", dataSource.ID))
	ctx.JSON(http.StatusCreated, dataSource)
}

// UpdateDataSource handles PUT /api/plan-scheduler/profiles/:id/datasources/:dsid
func (c *PlanSchedulerProfileController) UpdateDataSource(ctx *gin.Context) {
	profileID := ctx.Param("id")
	dsID := ctx.Param("dsid")
	c.iLog.Info(fmt.Sprintf("UpdateDataSource START - ProfileID: %s, DataSourceID: %s", profileID, dsID))

	user := c.getUserFromContext(ctx)

	var dataSource models.PlanSchedulerProfileDataSource
	if err := ctx.ShouldBindJSON(&dataSource); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	dataSource.ID = dsID
	dataSource.ProfileID = profileID
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	if err := service.UpdateDataSource(ctx.Request.Context(), &dataSource, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to update data source: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data source", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("UpdateDataSource COMPLETE - ID: %s", dsID))
	ctx.JSON(http.StatusOK, dataSource)
}

// DeleteDataSource handles DELETE /api/plan-scheduler/profiles/:id/datasources/:dsid
func (c *PlanSchedulerProfileController) DeleteDataSource(ctx *gin.Context) {
	profileID := ctx.Param("id")
	dsID := ctx.Param("dsid")
	c.iLog.Info(fmt.Sprintf("DeleteDataSource START - ProfileID: %s, DataSourceID: %s", profileID, dsID))

	user := c.getUserFromContext(ctx)
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	if err := service.DeleteDataSource(ctx.Request.Context(), profileID, dsID, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to delete data source: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data source", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("DeleteDataSource COMPLETE - ID: %s", dsID))
	ctx.JSON(http.StatusOK, gin.H{"message": "Data source deleted successfully", "id": dsID})
}

// CreateConstraint handles POST /api/plan-scheduler/profiles/:id/constraints
func (c *PlanSchedulerProfileController) CreateConstraint(ctx *gin.Context) {
	profileID := ctx.Param("id")
	c.iLog.Info(fmt.Sprintf("CreateConstraint START - ProfileID: %s", profileID))

	user := c.getUserFromContext(ctx)

	var constraint models.PlanSchedulerProfileConstraint
	if err := ctx.ShouldBindJSON(&constraint); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	constraint.ProfileID = profileID
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	if err := service.CreateConstraint(ctx.Request.Context(), &constraint, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to create constraint: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create constraint", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("CreateConstraint COMPLETE - ID: %s", constraint.ID))
	ctx.JSON(http.StatusCreated, constraint)
}

// UpdateConstraint handles PUT /api/plan-scheduler/profiles/:id/constraints/:cid
func (c *PlanSchedulerProfileController) UpdateConstraint(ctx *gin.Context) {
	profileID := ctx.Param("id")
	cID := ctx.Param("cid")
	c.iLog.Info(fmt.Sprintf("UpdateConstraint START - ProfileID: %s, ConstraintID: %s", profileID, cID))

	user := c.getUserFromContext(ctx)

	var constraint models.PlanSchedulerProfileConstraint
	if err := ctx.ShouldBindJSON(&constraint); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	constraint.ID = cID
	constraint.ProfileID = profileID
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	if err := service.UpdateConstraint(ctx.Request.Context(), &constraint, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to update constraint: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update constraint", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("UpdateConstraint COMPLETE - ID: %s", cID))
	ctx.JSON(http.StatusOK, constraint)
}

// DeleteConstraint handles DELETE /api/plan-scheduler/profiles/:id/constraints/:cid
func (c *PlanSchedulerProfileController) DeleteConstraint(ctx *gin.Context) {
	profileID := ctx.Param("id")
	cID := ctx.Param("cid")
	c.iLog.Info(fmt.Sprintf("DeleteConstraint START - ProfileID: %s, ConstraintID: %s", profileID, cID))

	user := c.getUserFromContext(ctx)
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	if err := service.DeleteConstraint(ctx.Request.Context(), profileID, cID, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to delete constraint: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete constraint", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("DeleteConstraint COMPLETE - ID: %s", cID))
	ctx.JSON(http.StatusOK, gin.H{"message": "Constraint deleted successfully", "id": cID})
}

// CreateSetting handles POST /api/plan-scheduler/profiles/:id/settings
func (c *PlanSchedulerProfileController) CreateSetting(ctx *gin.Context) {
	profileID := ctx.Param("id")
	c.iLog.Info(fmt.Sprintf("CreateSetting START - ProfileID: %s", profileID))

	user := c.getUserFromContext(ctx)

	var setting models.PlanSchedulerProfileSetting
	if err := ctx.ShouldBindJSON(&setting); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	setting.ProfileID = profileID
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	if err := service.CreateSetting(ctx.Request.Context(), &setting, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to create setting: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create setting", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("CreateSetting COMPLETE - ID: %s", setting.ID))
	ctx.JSON(http.StatusCreated, setting)
}

// UpdateSetting handles PUT /api/plan-scheduler/profiles/:id/settings/:sid
func (c *PlanSchedulerProfileController) UpdateSetting(ctx *gin.Context) {
	profileID := ctx.Param("id")
	sID := ctx.Param("sid")
	c.iLog.Info(fmt.Sprintf("UpdateSetting START - ProfileID: %s, SettingID: %s", profileID, sID))

	user := c.getUserFromContext(ctx)

	var setting models.PlanSchedulerProfileSetting
	if err := ctx.ShouldBindJSON(&setting); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	setting.ID = sID
	setting.ProfileID = profileID
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	if err := service.UpdateSetting(ctx.Request.Context(), &setting, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to update setting: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("UpdateSetting COMPLETE - ID: %s", sID))
	ctx.JSON(http.StatusOK, setting)
}

// DeleteSetting handles DELETE /api/plan-scheduler/profiles/:id/settings/:sid
func (c *PlanSchedulerProfileController) DeleteSetting(ctx *gin.Context) {
	profileID := ctx.Param("id")
	sID := ctx.Param("sid")
	c.iLog.Info(fmt.Sprintf("DeleteSetting START - ProfileID: %s, SettingID: %s", profileID, sID))

	user := c.getUserFromContext(ctx)
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	if err := service.DeleteSetting(ctx.Request.Context(), profileID, sID, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to delete setting: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete setting", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("DeleteSetting COMPLETE - ID: %s", sID))
	ctx.JSON(http.StatusOK, gin.H{"message": "Setting deleted successfully", "id": sID})
}

// ====================================================================
// Plan Scheduler Session Endpoints
// ====================================================================

// GetSessions handles GET /api/plan-scheduler/profiles/:id/sessions
func (c *PlanSchedulerProfileController) GetSessions(ctx *gin.Context) {
	profileID := ctx.Param("id")
	c.iLog.Info(fmt.Sprintf("GetSessions START - ProfileID: %s", profileID))

	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	sessions, err := service.GetSessionsByProfile(ctx.Request.Context(), profileID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to get sessions: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sessions", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("GetSessions COMPLETE - Found %d sessions", len(sessions)))
	ctx.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// GetSession handles GET /api/plan-scheduler/profiles/:id/sessions/:sid
func (c *PlanSchedulerProfileController) GetSession(ctx *gin.Context) {
	profileID := ctx.Param("id")
	sessionID := ctx.Param("sid")
	c.iLog.Info(fmt.Sprintf("GetSession START - ProfileID: %s, SessionID: %s", profileID, sessionID))

	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	session, err := service.GetSession(ctx.Request.Context(), sessionID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to get session: %v", err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Session not found", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("GetSession COMPLETE - SessionID: %s", sessionID))
	ctx.JSON(http.StatusOK, session)
}

// DeleteSession handles DELETE /api/plan-scheduler/profiles/:id/sessions/:sid
func (c *PlanSchedulerProfileController) DeleteSession(ctx *gin.Context) {
	profileID := ctx.Param("id")
	sessionID := ctx.Param("sid")
	c.iLog.Info(fmt.Sprintf("DeleteSession START - ProfileID: %s, SessionID: %s", profileID, sessionID))

	user := c.getUserFromContext(ctx)
	service := services.NewPlanSchedulerProfileService(dbconn.DB, gormdb.DB)

	if err := service.DeleteSession(ctx.Request.Context(), sessionID, user); err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to delete session: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session", "details": err.Error()})
		return
	}

	c.iLog.Info(fmt.Sprintf("DeleteSession COMPLETE - SessionID: %s", sessionID))
	ctx.JSON(http.StatusOK, gin.H{"message": "Session deleted successfully", "sessionId": sessionID})
}

// getUserFromContext extracts username from gin context
func (c *PlanSchedulerProfileController) getUserFromContext(ctx *gin.Context) string {
	// Try to get from session or token
	if user, exists := ctx.Get("user"); exists {
		if userStr, ok := user.(string); ok {
			return userStr
		}
	}

	// Try from header
	user := ctx.GetHeader("X-User-Id")
	if user == "" {
		user = ctx.GetHeader("X-Username")
	}

	// Default
	if user == "" {
		user = "system"
	}

	return user
}
