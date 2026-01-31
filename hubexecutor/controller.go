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

package hubexecutor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/logger"
)

// HubExecutorController handles HTTP requests for hub executor management
type HubExecutorController struct {
	manager *HubExecutorManager
	iLog    logger.Log
}

// NewHubExecutorController creates a new hub executor controller
func NewHubExecutorController(manager *HubExecutorManager) *HubExecutorController {
	return &HubExecutorController{
		manager: manager,
		iLog:    logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "HubExecutorController"},
	}
}

// RegisterRoutes registers the hub executor API routes
func (c *HubExecutorController) RegisterRoutes(router *gin.RouterGroup) {
	hubexecutor := router.Group("/hubexecutor")
	{
		hubexecutor.POST("/start/:name", c.StartExecutor)
		hubexecutor.POST("/start-by-id/:hubId", c.StartExecutorByHubID)
		hubexecutor.POST("/stop/:name", c.StopExecutor)
		hubexecutor.POST("/restart/:name", c.RestartExecutor)
		hubexecutor.GET("/status/:name", c.GetExecutorStatus)
		hubexecutor.GET("/info/:name", c.GetExecutorInfo)
		hubexecutor.GET("/list", c.ListExecutors)
		hubexecutor.POST("/start-all", c.StartAllExecutors)
		hubexecutor.POST("/stop-all", c.StopAllExecutors)
	}
}

// StartExecutor handles POST /hubexecutor/start/:instanceName
// @Summary Start executor by instance name
// @Description Starts a hub executor for the specified instance name
// @Tags HubExecutor
// @Produce json
// @Param name path string true "Hub Name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /hubexecutor/start/{name} [post]
func (c *HubExecutorController) StartExecutor(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "name is required",
		})
		return
	}

	err := c.manager.StartExecutor(name)
	if err != nil {
		c.iLog.Error("Failed to start executor: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Executor started successfully",
		"name":    name,
	})
}

// StartExecutorByHubID handles POST /hubexecutor/start-by-id/:hubId
// @Summary Start executor by hub ID
// @Description Starts a hub executor using a specific hub configuration ID
// @Tags HubExecutor
// @Produce json
// @Param hubId path string true "Hub ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /hubexecutor/start-by-id/{hubId} [post]
func (c *HubExecutorController) StartExecutorByHubID(ctx *gin.Context) {
	hubID := ctx.Param("hubId")
	if hubID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "hub ID is required",
		})
		return
	}

	err := c.manager.StartExecutorByHubID(hubID)
	if err != nil {
		c.iLog.Error("Failed to start executor by hub ID: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Executor started successfully",
		"hubId":   hubID,
	})
}

// StopExecutor handles POST /hubexecutor/stop/:instanceName
// @Summary Stop executor
// @Description Stops a hub executor for the specified instance name
// @Tags HubExecutor
// @Produce json
// @Param name path string true "Hub Name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /hubexecutor/stop/{name} [post]
func (c *HubExecutorController) StopExecutor(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "name is required",
		})
		return
	}

	err := c.manager.StopExecutor(name)
	if err != nil {
		c.iLog.Error("Failed to stop executor: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Executor stopped successfully",
		"name":    name,
	})
}

// RestartExecutor handles POST /hubexecutor/restart/:instanceName
// @Summary Restart executor
// @Description Restarts a hub executor for the specified instance name
// @Tags HubExecutor
// @Produce json
// @Param name path string true "Hub Name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /hubexecutor/restart/{name} [post]
func (c *HubExecutorController) RestartExecutor(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "name is required",
		})
		return
	}

	err := c.manager.RestartExecutor(name)
	if err != nil {
		c.iLog.Error("Failed to restart executor: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Executor restarted successfully",
		"name":    name,
	})
}

// GetExecutorStatus handles GET /hubexecutor/status/:instanceName
// @Summary Get executor status
// @Description Gets the status of a hub executor
// @Tags HubExecutor
// @Produce json
// @Param name path string true "Hub Name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /hubexecutor/status/{name} [get]
func (c *HubExecutorController) GetExecutorStatus(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "name is required",
		})
		return
	}

	status, err := c.manager.GetExecutorStatus(name)
	if err != nil {
		c.iLog.Error("Failed to get executor status: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"name":    name,
		"status":  status,
	})
}

// GetExecutorInfo handles GET /hubexecutor/info/:instanceName
// @Summary Get executor detailed info
// @Description Gets detailed information about a hub executor
// @Tags HubExecutor
// @Produce json
// @Param name path string true "Hub Name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /hubexecutor/info/{name} [get]
func (c *HubExecutorController) GetExecutorInfo(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "name is required",
		})
		return
	}

	info, err := c.manager.GetExecutorInfo(name)
	if err != nil {
		c.iLog.Error("Failed to get executor info: " + err.Error())
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"info":    info,
	})
}

// ListExecutors handles GET /hubexecutor/list
// @Summary List all running executors
// @Description Returns information about all running hub executors
// @Tags HubExecutor
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /hubexecutor/list [get]
func (c *HubExecutorController) ListExecutors(ctx *gin.Context) {
	executors := c.manager.ListExecutors()

	ctx.JSON(http.StatusOK, gin.H{
		"success":   true,
		"executors": executors,
		"count":     len(executors),
	})
}

// StartAllExecutors handles POST /hubexecutor/start-all
// @Summary Start all enabled hub executors
// @Description Starts executors for all enabled hub configurations
// @Tags HubExecutor
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /hubexecutor/start-all [post]
func (c *HubExecutorController) StartAllExecutors(ctx *gin.Context) {
	results := c.manager.StartAllExecutors()

	// Count successes and failures
	successes := 0
	failures := 0
	errorDetails := make(map[string]string)

	for name, err := range results {
		if err != nil {
			failures++
			errorDetails[name] = err.Error()
		} else {
			successes++
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": failures == 0,
		"started": successes,
		"failed":  failures,
		"errors":  errorDetails,
		"message": "Start all executors completed",
	})
}

// StopAllExecutors handles POST /hubexecutor/stop-all
// @Summary Stop all running executors
// @Description Stops all currently running hub executors
// @Tags HubExecutor
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /hubexecutor/stop-all [post]
func (c *HubExecutorController) StopAllExecutors(ctx *gin.Context) {
	err := c.manager.StopAll()
	if err != nil {
		c.iLog.Error("Failed to stop all executors: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All executors stopped successfully",
	})
}
