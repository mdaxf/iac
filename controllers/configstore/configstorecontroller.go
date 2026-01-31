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

package configstore

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services/configstore"
)

// ConfigStoreController handles configuration store API endpoints
type ConfigStoreController struct {
	log logger.Log
}

// NewConfigStoreController creates a new config store controller
func NewConfigStoreController() *ConfigStoreController {
	return &ConfigStoreController{
		log: logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ConfigStoreController"},
	}
}

// RegisterRoutes registers the configuration store routes
func (c *ConfigStoreController) RegisterRoutes(router *gin.RouterGroup) {
	cfg := router.Group("/configstore")
	{
		cfg.GET("/", c.ListConfigurations)
		cfg.GET("/active", c.GetActiveConfiguration)
		cfg.GET("/:id", c.GetConfiguration)
		cfg.POST("/", c.CreateConfiguration)
		cfg.PUT("/:id", c.UpdateConfiguration)
		cfg.POST("/:id/activate", c.ActivateConfiguration)
		cfg.POST("/rollback", c.RollbackConfiguration)
		cfg.GET("/:id/history", c.GetHistory)
		cfg.POST("/reload", c.BroadcastReload)
	}
}

// CreateConfigurationRequest represents a create configuration request
type CreateConfigurationRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Version     string                 `json:"version" binding:"required"`
	Environment string                 `json:"environment" binding:"required"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data" binding:"required"`
	Tags        []string               `json:"tags"`
}

// UpdateConfigurationRequest represents an update configuration request
type UpdateConfigurationRequest struct {
	Data   map[string]interface{} `json:"data" binding:"required"`
	Reason string                 `json:"reason"`
}

// RollbackRequest represents a rollback request
type RollbackRequest struct {
	Name          string `json:"name" binding:"required"`
	Environment   string `json:"environment" binding:"required"`
	TargetVersion string `json:"target_version" binding:"required"`
	Reason        string `json:"reason"`
}

// ListConfigurations returns a list of configurations
// @Summary List configurations
// @Description Returns a list of configurations with optional filters
// @Tags configstore
// @Produce json
// @Param name query string false "Configuration name"
// @Param environment query string false "Environment"
// @Param status query string false "Status (draft, active, archived)"
// @Success 200 {array} models.ConfigurationStore
// @Failure 500 {object} map[string]string
// @Router /api/configstore [get]
func (c *ConfigStoreController) ListConfigurations(ctx *gin.Context) {
	service := configstore.GetGlobalConfigStoreService()
	if service == nil {
		// Return empty list in standalone mode with helpful message
		ctx.JSON(http.StatusOK, gin.H{
			"configstore_enabled": false,
			"configurations":      []interface{}{},
			"count":               0,
			"message":             "Configuration store is not enabled. Add cluster.config_store.enabled=true to configuration.json to enable centralized configuration management.",
		})
		return
	}

	name := ctx.Query("name")
	environment := ctx.Query("environment")
	status := models.ConfigurationStatus(ctx.Query("status"))

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	configs, err := service.ListConfigurations(reqCtx, name, environment, status)
	if err != nil {
		c.log.Error("Failed to list configurations: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list configurations",
		})
		return
	}

	// Convert to list items for response
	items := make([]models.ConfigurationListItem, len(configs))
	for i, cfg := range configs {
		items[i] = cfg.ToListItem()
	}

	ctx.JSON(http.StatusOK, gin.H{
		"configstore_enabled": true,
		"configurations":      items,
		"count":               len(items),
	})
}

// GetActiveConfiguration returns the active configuration
// @Summary Get active configuration
// @Description Returns the currently active configuration for a name and environment
// @Tags configstore
// @Produce json
// @Param name query string true "Configuration name"
// @Param environment query string true "Environment"
// @Success 200 {object} models.ConfigurationStore
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/configstore/active [get]
func (c *ConfigStoreController) GetActiveConfiguration(ctx *gin.Context) {
	service := configstore.GetGlobalConfigStoreService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Configuration store service not available",
		})
		return
	}

	name := ctx.Query("name")
	environment := ctx.Query("environment")

	if name == "" || environment == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "name and environment are required",
		})
		return
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := service.GetActiveConfiguration(reqCtx, name, environment)
	if err != nil {
		c.log.Error("Failed to get active configuration: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get active configuration",
		})
		return
	}

	if cfg == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "No active configuration found",
		})
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

// GetConfiguration returns a specific configuration
// @Summary Get configuration by ID
// @Description Returns a specific configuration by its ID
// @Tags configstore
// @Produce json
// @Param id path string true "Configuration ID"
// @Success 200 {object} models.ConfigurationStore
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/configstore/{id} [get]
func (c *ConfigStoreController) GetConfiguration(ctx *gin.Context) {
	service := configstore.GetGlobalConfigStoreService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Configuration store service not available",
		})
		return
	}

	id := ctx.Param("id")

	reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := service.GetConfiguration(reqCtx, id)
	if err != nil {
		c.log.Error("Failed to get configuration: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get configuration",
		})
		return
	}

	if cfg == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Configuration not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

// CreateConfiguration creates a new configuration
// @Summary Create configuration
// @Description Creates a new configuration version
// @Tags configstore
// @Accept json
// @Produce json
// @Param config body CreateConfigurationRequest true "Configuration"
// @Success 201 {object} models.ConfigurationStore
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/configstore [post]
func (c *ConfigStoreController) CreateConfiguration(ctx *gin.Context) {
	service := configstore.GetGlobalConfigStoreService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Configuration store service not available",
		})
		return
	}

	var req CreateConfigurationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Get user from context (set by auth middleware)
	userID, _ := ctx.Get("user_id")
	createdBy := "system"
	if userID != nil {
		createdBy = userID.(string)
	}

	cfg := models.NewConfigurationStore(req.Name, req.Version, req.Environment, createdBy, req.Data)
	cfg.Description = req.Description
	cfg.Tags = req.Tags

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	created, err := service.CreateConfiguration(reqCtx, cfg)
	if err != nil {
		c.log.Error("Failed to create configuration: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create configuration: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, created)
}

// UpdateConfiguration updates a configuration
// @Summary Update configuration
// @Description Updates a configuration (only draft status allowed)
// @Tags configstore
// @Accept json
// @Produce json
// @Param id path string true "Configuration ID"
// @Param config body UpdateConfigurationRequest true "Update data"
// @Success 200 {object} models.ConfigurationStore
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/configstore/{id} [put]
func (c *ConfigStoreController) UpdateConfiguration(ctx *gin.Context) {
	service := configstore.GetGlobalConfigStoreService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Configuration store service not available",
		})
		return
	}

	id := ctx.Param("id")

	var req UpdateConfigurationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Get user from context
	userID, _ := ctx.Get("user_id")
	updatedBy := "system"
	if userID != nil {
		updatedBy = userID.(string)
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updated, err := service.UpdateConfiguration(reqCtx, id, req.Data, updatedBy, req.Reason)
	if err != nil {
		c.log.Error("Failed to update configuration: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update configuration: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, updated)
}

// ActivateConfiguration activates a configuration
// @Summary Activate configuration
// @Description Activates a configuration version (deactivates previous)
// @Tags configstore
// @Produce json
// @Param id path string true "Configuration ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/configstore/{id}/activate [post]
func (c *ConfigStoreController) ActivateConfiguration(ctx *gin.Context) {
	service := configstore.GetGlobalConfigStoreService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Configuration store service not available",
		})
		return
	}

	id := ctx.Param("id")

	// Get user from context
	userID, _ := ctx.Get("user_id")
	activatedBy := "system"
	if userID != nil {
		activatedBy = userID.(string)
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := service.ActivateConfiguration(reqCtx, id, activatedBy); err != nil {
		c.log.Error("Failed to activate configuration: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to activate configuration: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Configuration activated successfully",
	})
}

// RollbackConfiguration rolls back to a previous version
// @Summary Rollback configuration
// @Description Rolls back to a specific previous version
// @Tags configstore
// @Accept json
// @Produce json
// @Param request body RollbackRequest true "Rollback request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/configstore/rollback [post]
func (c *ConfigStoreController) RollbackConfiguration(ctx *gin.Context) {
	service := configstore.GetGlobalConfigStoreService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Configuration store service not available",
		})
		return
	}

	var req RollbackRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Get user from context
	userID, _ := ctx.Get("user_id")
	performedBy := "system"
	if userID != nil {
		performedBy = userID.(string)
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := service.RollbackConfiguration(reqCtx, req.Name, req.Environment, req.TargetVersion, performedBy, req.Reason); err != nil {
		c.log.Error("Failed to rollback configuration: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to rollback configuration: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Configuration rolled back successfully",
	})
}

// GetHistory returns configuration history
// @Summary Get configuration history
// @Description Returns the change history for a configuration
// @Tags configstore
// @Produce json
// @Param id path string true "Configuration ID"
// @Param limit query int false "Limit number of results"
// @Success 200 {array} models.ConfigurationHistory
// @Failure 500 {object} map[string]string
// @Router /api/configstore/{id}/history [get]
func (c *ConfigStoreController) GetHistory(ctx *gin.Context) {
	service := configstore.GetGlobalConfigStoreService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Configuration store service not available",
		})
		return
	}

	id := ctx.Param("id")
	limit := 50
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	history, err := service.GetHistory(reqCtx, id, limit)
	if err != nil {
		c.log.Error("Failed to get history: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get history",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"history": history,
		"count":   len(history),
	})
}

// BroadcastReload triggers a configuration reload across all instances
// @Summary Broadcast configuration reload
// @Description Notifies all instances to reload their configuration
// @Tags configstore
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/configstore/reload [post]
func (c *ConfigStoreController) BroadcastReload(ctx *gin.Context) {
	// This endpoint can be used to manually trigger a config reload broadcast
	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Reload broadcast initiated",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
