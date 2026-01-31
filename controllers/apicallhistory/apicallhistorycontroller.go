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

package apicallhistory

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/framework/auth"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services/apicallhistory"
)

// APICallHistoryController handles API call history endpoints
type APICallHistoryController struct {
	log logger.Log
}

// NewAPICallHistoryController creates a new controller
func NewAPICallHistoryController() *APICallHistoryController {
	return &APICallHistoryController{
		log: logger.Log{ModuleName: logger.API, User: "System", ControllerName: "APICallHistoryController"},
	}
}

// GetConfig returns the current configuration
// @Summary Get API call history configuration
// @Description Returns the current API call history tracking configuration
// @Tags apicallhistory
// @Produce json
// @Success 200 {object} models.APICallHistoryConfig
// @Failure 503 {object} map[string]string
// @Router /api/apicallhistory/config [get]
func (c *APICallHistoryController) GetConfig(ctx *gin.Context) {
	service := apicallhistory.GetGlobalAPICallHistoryService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "API call history service not available",
			"message": "Service may not be initialized",
		})
		return
	}

	config := service.GetConfig()
	ctx.JSON(http.StatusOK, config)
}

// UpdateConfig updates the configuration
// @Summary Update API call history configuration
// @Description Updates the API call history tracking configuration
// @Tags apicallhistory
// @Accept json
// @Produce json
// @Param config body models.APICallHistoryConfig true "Configuration"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/apicallhistory/config [put]
func (c *APICallHistoryController) UpdateConfig(ctx *gin.Context) {
	service := apicallhistory.GetGlobalAPICallHistoryService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "API call history service not available",
		})
		return
	}

	var config models.APICallHistoryConfig
	if err := ctx.ShouldBindJSON(&config); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Get user info
	userID, _, _, _ := auth.GetUserInformation(ctx)
	if userID == "" {
		userID = "system"
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := service.UpdateConfig(reqCtx, &config, userID); err != nil {
		c.log.Error("Failed to update config: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update configuration",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
	})
}

// ResetConfig resets configuration to defaults
// @Summary Reset API call history configuration
// @Description Resets the configuration to default values
// @Tags apicallhistory
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/apicallhistory/config/reset [post]
func (c *APICallHistoryController) ResetConfig(ctx *gin.Context) {
	service := apicallhistory.GetGlobalAPICallHistoryService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "API call history service not available",
		})
		return
	}

	userID, _, _, _ := auth.GetUserInformation(ctx)
	if userID == "" {
		userID = "system"
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := service.ResetConfig(reqCtx, userID); err != nil {
		c.log.Error("Failed to reset config: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reset configuration",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Configuration reset to defaults",
	})
}

// ListHistory returns paginated API call history
// @Summary List API call history
// @Description Returns paginated list of API call history with filters
// @Tags apicallhistory
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param method query string false "HTTP method filter"
// @Param endpoint query string false "Endpoint filter (supports partial match)"
// @Param statusCode query int false "Status code filter"
// @Param userId query string false "User ID filter"
// @Param sourceIp query string false "Source IP filter"
// @Param startDate query string false "Start time filter (RFC3339)"
// @Param endDate query string false "End time filter (RFC3339)"
// @Param onlyErrors query bool false "Only show errors (4xx/5xx)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/apicallhistory [get]
func (c *APICallHistoryController) ListHistory(ctx *gin.Context) {
	service := apicallhistory.GetGlobalAPICallHistoryService()
	if service == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"enabled":  false,
			"history":  []interface{}{},
			"total":    0,
			"page":     1,
			"pageSize": 20,
			"message":  "API call history service not enabled",
		})
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Parse filters
	params := models.ListHistoryParams{
		Method:     ctx.Query("method"),
		Endpoint:   ctx.Query("endpoint"),
		UserID:     ctx.Query("userId"),
		SourceIP:   ctx.Query("sourceIp"),
		OnlyErrors: ctx.Query("onlyErrors") == "true",
		Page:       page,
		PageSize:   pageSize,
	}

	if statusCode := ctx.Query("statusCode"); statusCode != "" {
		code, _ := strconv.Atoi(statusCode)
		params.StatusCode = code
	}

	if startDate := ctx.Query("startDate"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			params.StartDate = t
		}
	}

	if endDate := ctx.Query("endDate"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			params.EndDate = t
		}
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := service.ListHistory(reqCtx, params)
	if err != nil {
		c.log.Error("Failed to list history: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list history",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"enabled":  true,
		"history":  response.History,
		"total":    response.Total,
		"page":     response.Page,
		"pages":    response.Pages,
		"pageSize": pageSize,
	})
}

// GetHistory returns a single history record
// @Summary Get API call history record
// @Description Returns detailed information about a single API call
// @Tags apicallhistory
// @Produce json
// @Param id path string true "History record ID"
// @Success 200 {object} models.APICallHistory
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/apicallhistory/{id} [get]
func (c *APICallHistoryController) GetHistory(ctx *gin.Context) {
	service := apicallhistory.GetGlobalAPICallHistoryService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "API call history service not available",
		})
		return
	}

	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "ID is required",
		})
		return
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	history, err := service.GetHistory(reqCtx, id)
	if err != nil {
		if err.Error() == "record not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "History record not found",
			})
			return
		}
		c.log.Error("Failed to get history: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get history",
		})
		return
	}

	ctx.JSON(http.StatusOK, history)
}

// DeleteHistory deletes a history record
// @Summary Delete API call history record
// @Description Deletes a single API call history record
// @Tags apicallhistory
// @Param id path string true "History record ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/apicallhistory/{id} [delete]
func (c *APICallHistoryController) DeleteHistory(ctx *gin.Context) {
	service := apicallhistory.GetGlobalAPICallHistoryService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "API call history service not available",
		})
		return
	}

	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "ID is required",
		})
		return
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := service.DeleteHistory(reqCtx, id); err != nil {
		if err.Error() == "record not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "History record not found",
			})
			return
		}
		c.log.Error("Failed to delete history: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete history",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "History record deleted",
	})
}

// CleanupRequest represents a cleanup request
type CleanupRequest struct {
	OlderThanDays int `json:"older_than_days" binding:"required,min=1"`
}

// Cleanup deletes old history records
// @Summary Cleanup old API call history
// @Description Deletes API call history records older than specified days
// @Tags apicallhistory
// @Accept json
// @Produce json
// @Param request body CleanupRequest true "Cleanup request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/apicallhistory/cleanup [post]
func (c *APICallHistoryController) Cleanup(ctx *gin.Context) {
	service := apicallhistory.GetGlobalAPICallHistoryService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "API call history service not available",
		})
		return
	}

	var req CleanupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	cutoff := time.Now().AddDate(0, 0, -req.OlderThanDays)

	reqCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	deleted, err := service.Cleanup(reqCtx, cutoff)
	if err != nil {
		c.log.Error("Failed to cleanup: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cleanup history",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Cleanup completed",
		"deleted": deleted,
		"cutoff":  cutoff.Format(time.RFC3339),
	})
}

// GetStats returns API call statistics
// @Summary Get API call statistics
// @Description Returns statistics about API calls
// @Tags apicallhistory
// @Produce json
// @Success 200 {object} models.APICallHistoryStats
// @Failure 500 {object} map[string]string
// @Router /api/apicallhistory/stats [get]
func (c *APICallHistoryController) GetStats(ctx *gin.Context) {
	service := apicallhistory.GetGlobalAPICallHistoryService()
	if service == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"enabled":     false,
			"total_calls": 0,
			"message":     "API call history service not enabled",
		})
		return
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stats, err := service.GetStats(reqCtx)
	if err != nil {
		c.log.Error("Failed to get stats: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get statistics",
		})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}
