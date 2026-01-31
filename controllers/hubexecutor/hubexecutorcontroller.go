package hubexecutor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/hubexecutor"
	"github.com/mdaxf/iac/logger"
)

// HubExecutorController handles API requests for hub executor management
type HubExecutorController struct {
	iLog    logger.Log
	manager *hubexecutor.HubExecutorManager
}

// NewHubExecutorController creates a new hub executor controller
func NewHubExecutorController(manager *hubexecutor.HubExecutorManager) *HubExecutorController {
	return &HubExecutorController{
		iLog:    logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "HubExecutorController"},
		manager: manager,
	}
}

// RegisterRoutes registers the hub executor API routes
func (c *HubExecutorController) RegisterRoutes(router *gin.RouterGroup) {
	executor := router.Group("/hubexecutor")
	{
		executor.POST("/start/:instanceName", c.StartExecutor)
		executor.POST("/stop/:instanceName", c.StopExecutor)
		executor.POST("/restart/:instanceName", c.RestartExecutor)
		executor.GET("/status/:instanceName", c.GetStatus)
		executor.GET("/info/:instanceName", c.GetInfo)
		executor.GET("/list", c.ListExecutors)
		executor.POST("/start-all", c.StartAll)
		executor.POST("/stop-all", c.StopAll)
		executor.POST("/start-by-id/:hubId", c.StartExecutorByHubID)
	}
}

// StartExecutor handles POST /hubexecutor/start/:instanceName
// @Summary Start a hub executor by instance name
// @Description Starts a hub executor for the specified instance name, loading the default configuration
// @Tags HubExecutor
// @Accept json
// @Produce json
// @Param instanceName path string true "Instance Name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /hubexecutor/start/{instanceName} [post]
func (c *HubExecutorController) StartExecutor(ctx *gin.Context) {
	instanceName := ctx.Param("instanceName")
	if instanceName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "instance name is required",
		})
		return
	}

	c.iLog.Debug("Starting executor for instance: " + instanceName)

	if err := c.manager.StartExecutor(instanceName); err != nil {
		c.iLog.Error("Failed to start executor: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Executor started successfully",
		"instance_name": instanceName,
	})
}

// StartExecutorByHubID handles POST /hubexecutor/start-by-id/:hubId
// @Summary Start a hub executor by hub ID
// @Description Starts a hub executor using a specific hub configuration ID
// @Tags HubExecutor
// @Accept json
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

	c.iLog.Debug("Starting executor for hub ID: " + hubID)

	if err := c.manager.StartExecutorByHubID(hubID); err != nil {
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
		"hub_id":  hubID,
	})
}

// StopExecutor handles POST /hubexecutor/stop/:instanceName
// @Summary Stop a hub executor by instance name
// @Description Stops the running hub executor for the specified instance name
// @Tags HubExecutor
// @Accept json
// @Produce json
// @Param instanceName path string true "Instance Name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /hubexecutor/stop/{instanceName} [post]
func (c *HubExecutorController) StopExecutor(ctx *gin.Context) {
	instanceName := ctx.Param("instanceName")
	if instanceName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "instance name is required",
		})
		return
	}

	c.iLog.Debug("Stopping executor for instance: " + instanceName)

	if err := c.manager.StopExecutor(instanceName); err != nil {
		c.iLog.Error("Failed to stop executor: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Executor stopped successfully",
		"instance_name": instanceName,
	})
}

// RestartExecutor handles POST /hubexecutor/restart/:instanceName
// @Summary Restart a hub executor by instance name
// @Description Restarts the hub executor for the specified instance name
// @Tags HubExecutor
// @Accept json
// @Produce json
// @Param instanceName path string true "Instance Name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /hubexecutor/restart/{instanceName} [post]
func (c *HubExecutorController) RestartExecutor(ctx *gin.Context) {
	instanceName := ctx.Param("instanceName")
	if instanceName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "instance name is required",
		})
		return
	}

	c.iLog.Debug("Restarting executor for instance: " + instanceName)

	if err := c.manager.RestartExecutor(instanceName); err != nil {
		c.iLog.Error("Failed to restart executor: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Executor restarted successfully",
		"instance_name": instanceName,
	})
}

// GetStatus handles GET /hubexecutor/status/:instanceName
// @Summary Get executor status by instance name
// @Description Gets the current status of the hub executor for the specified instance name
// @Tags HubExecutor
// @Produce json
// @Param instanceName path string true "Instance Name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /hubexecutor/status/{instanceName} [get]
func (c *HubExecutorController) GetStatus(ctx *gin.Context) {
	instanceName := ctx.Param("instanceName")
	if instanceName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "instance name is required",
		})
		return
	}

	status, err := c.manager.GetExecutorStatus(instanceName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":       true,
		"instance_name": instanceName,
		"status":        status,
		"is_running":    c.manager.IsRunning(instanceName),
	})
}

// GetInfo handles GET /hubexecutor/info/:instanceName
// @Summary Get executor info by instance name
// @Description Gets detailed information about the hub executor for the specified instance name
// @Tags HubExecutor
// @Produce json
// @Param instanceName path string true "Instance Name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /hubexecutor/info/{instanceName} [get]
func (c *HubExecutorController) GetInfo(ctx *gin.Context) {
	instanceName := ctx.Param("instanceName")
	if instanceName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "instance name is required",
		})
		return
	}

	info, err := c.manager.GetExecutorInfo(instanceName)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// ListExecutors handles GET /hubexecutor/list
// @Summary List all running executors
// @Description Gets a list of all running hub executors with their status
// @Tags HubExecutor
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /hubexecutor/list [get]
func (c *HubExecutorController) ListExecutors(ctx *gin.Context) {
	executors := c.manager.ListExecutors()

	ctx.JSON(http.StatusOK, gin.H{
		"success":       true,
		"data":          executors,
		"total":         len(executors),
		"running_count": c.manager.GetRunningCount(),
	})
}

// StartAll handles POST /hubexecutor/start-all
// @Summary Start all enabled hub executors
// @Description Starts executors for all enabled hub configurations
// @Tags HubExecutor
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /hubexecutor/start-all [post]
func (c *HubExecutorController) StartAll(ctx *gin.Context) {
	c.iLog.Debug("Starting all hub executors")

	if err := c.manager.StartAll(); err != nil {
		c.iLog.Error("Failed to start all executors: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "All executors started",
		"running_count": c.manager.GetRunningCount(),
	})
}

// StopAll handles POST /hubexecutor/stop-all
// @Summary Stop all running executors
// @Description Stops all running hub executors
// @Tags HubExecutor
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /hubexecutor/stop-all [post]
func (c *HubExecutorController) StopAll(ctx *gin.Context) {
	c.iLog.Debug("Stopping all hub executors")

	if err := c.manager.StopAll(); err != nil {
		c.iLog.Error("Failed to stop all executors: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All executors stopped",
	})
}

// SetManager sets the hub executor manager
func (c *HubExecutorController) SetManager(manager *hubexecutor.HubExecutorManager) {
	c.manager = manager
}

// GetManager returns the hub executor manager
func (c *HubExecutorController) GetManager() *hubexecutor.HubExecutorManager {
	return c.manager
}
