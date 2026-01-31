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

package inthubhistory

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services/inthubhistory"
)

// IntHubHistoryController handles integration hub history endpoints
type IntHubHistoryController struct {
	log logger.Log
}

// NewIntHubHistoryController creates a new controller
func NewIntHubHistoryController() *IntHubHistoryController {
	return &IntHubHistoryController{
		log: logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntHubHistoryController"},
	}
}

// ListHistory returns paginated integration hub history
// @Summary List integration hub history
// @Description Returns paginated list of integration hub history with filters
// @Tags inthubhistory
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param hub_id query string false "Hub ID filter"
// @Param direction query string false "Direction filter (inbound/outbound)"
// @Param protocol query string false "Protocol filter"
// @Param endpoint_id query string false "Endpoint ID filter"
// @Param topic query string false "Topic filter"
// @Param status query string false "Status filter"
// @Param start_time query string false "Start time filter (RFC3339)"
// @Param end_time query string false "End time filter (RFC3339)"
// @Param only_errors query bool false "Only show errors"
// @Success 200 {object} models.ListIntHubHistoryResponse
// @Failure 500 {object} map[string]string
// @Router /api/integration-hub/history [get]
func (c *IntHubHistoryController) ListHistory(ctx *gin.Context) {
	startTime := time.Now()
	defer func() {
		c.log.PerformanceWithDuration("IntHubHistoryController.ListHistory", time.Since(startTime))
	}()

	service := inthubhistory.GetGlobalIntHubHistoryService()
	if service == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"enabled":   false,
			"history":   []interface{}{},
			"total":     0,
			"page":      1,
			"page_size": 20,
			"pages":     0,
			"message":   "Integration hub history service not enabled",
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
	params := models.ListIntHubHistoryParams{
		HubID:      ctx.Query("hub_id"),
		Direction:  ctx.Query("direction"),
		Protocol:   ctx.Query("protocol"),
		EndpointID: ctx.Query("endpoint_id"),
		Topic:      ctx.Query("topic"),
		OnlyErrors: ctx.Query("only_errors") == "true",
		Page:       page,
		PageSize:   pageSize,
	}

	if status := ctx.Query("status"); status != "" {
		params.Status = models.IntHubHistoryStatus(status)
	}

	if startTimeStr := ctx.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			params.StartTime = t
		}
	}

	if endTimeStr := ctx.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			params.EndTime = t
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
		"enabled":   true,
		"history":   response.History,
		"total":     response.Total,
		"page":      response.Page,
		"page_size": response.PageSize,
		"pages":     response.Pages,
	})
}

// GetHistory returns a single history record with its actions
// @Summary Get integration hub history record
// @Description Returns detailed information about a single integration hub transaction
// @Tags inthubhistory
// @Produce json
// @Param id path string true "History record ID"
// @Success 200 {object} models.IntHubHistory
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/integration-hub/history/{id} [get]
func (c *IntHubHistoryController) GetHistory(ctx *gin.Context) {
	startTime := time.Now()
	defer func() {
		c.log.PerformanceWithDuration("IntHubHistoryController.GetHistory", time.Since(startTime))
	}()

	service := inthubhistory.GetGlobalIntHubHistoryService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Integration hub history service not available",
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

	// Also get actions for this transaction
	actions, err := service.GetActionsByTransaction(reqCtx, id)
	if err == nil {
		history.Actions = actions
	}

	ctx.JSON(http.StatusOK, history)
}

// GetActions returns actions for a specific transaction
// @Summary Get actions for a transaction
// @Description Returns all actions for a specific integration hub transaction
// @Tags inthubhistory
// @Produce json
// @Param id path string true "History record ID"
// @Success 200 {array} models.IntHubActionHistory
// @Failure 500 {object} map[string]string
// @Router /api/integration-hub/history/{id}/actions [get]
func (c *IntHubHistoryController) GetActions(ctx *gin.Context) {
	startTime := time.Now()
	defer func() {
		c.log.PerformanceWithDuration("IntHubHistoryController.GetActions", time.Since(startTime))
	}()

	service := inthubhistory.GetGlobalIntHubHistoryService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Integration hub history service not available",
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

	actions, err := service.GetActionsByTransaction(reqCtx, id)
	if err != nil {
		c.log.Error("Failed to get actions: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get actions",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"actions": actions,
		"count":   len(actions),
	})
}

// DeleteHistory deletes a history record
// @Summary Delete integration hub history record
// @Description Soft-deletes a single integration hub history record
// @Tags inthubhistory
// @Param id path string true "History record ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/integration-hub/history/{id} [delete]
func (c *IntHubHistoryController) DeleteHistory(ctx *gin.Context) {
	startTime := time.Now()
	defer func() {
		c.log.PerformanceWithDuration("IntHubHistoryController.DeleteHistory", time.Since(startTime))
	}()

	service := inthubhistory.GetGlobalIntHubHistoryService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Integration hub history service not available",
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

// GetStats returns integration hub history statistics
// @Summary Get integration hub history statistics
// @Description Returns statistics about integration hub transactions
// @Tags inthubhistory
// @Produce json
// @Param hub_id query string false "Hub ID filter"
// @Param start_time query string false "Start time filter (RFC3339)"
// @Param end_time query string false "End time filter (RFC3339)"
// @Success 200 {object} models.IntHubHistoryStats
// @Failure 500 {object} map[string]string
// @Router /api/integration-hub/history/stats [get]
func (c *IntHubHistoryController) GetStats(ctx *gin.Context) {
	startTime := time.Now()
	defer func() {
		c.log.PerformanceWithDuration("IntHubHistoryController.GetStats", time.Since(startTime))
	}()

	service := inthubhistory.GetGlobalIntHubHistoryService()
	if service == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"enabled":            false,
			"total_transactions": 0,
			"message":            "Integration hub history service not enabled",
		})
		return
	}

	hubID := ctx.Query("hub_id")
	var filterStartTime, filterEndTime time.Time

	if startTimeStr := ctx.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filterStartTime = t
		}
	}

	if endTimeStr := ctx.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filterEndTime = t
		}
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stats, err := service.GetStatistics(reqCtx, hubID, filterStartTime, filterEndTime)
	if err != nil {
		c.log.Error("Failed to get stats: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get statistics",
		})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

// Cleanup deletes old history records
// @Summary Cleanup old integration hub history
// @Description Deletes integration hub history records older than specified days
// @Tags inthubhistory
// @Accept json
// @Produce json
// @Param request body models.CleanupIntHubHistoryRequest true "Cleanup request"
// @Success 200 {object} models.CleanupIntHubHistoryResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/integration-hub/history/cleanup [post]
func (c *IntHubHistoryController) Cleanup(ctx *gin.Context) {
	startTime := time.Now()
	defer func() {
		c.log.PerformanceWithDuration("IntHubHistoryController.Cleanup", time.Since(startTime))
	}()

	service := inthubhistory.GetGlobalIntHubHistoryService()
	if service == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Integration hub history service not available",
		})
		return
	}

	var req models.CleanupIntHubHistoryRequest
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
