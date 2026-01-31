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

package aischedule

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/services"
)

// AIScheduleController handles AI schedule generation and optimization
type AIScheduleController struct {
	service      *services.AIScheduleService
	batchService *services.AIScheduleBatchService
}

// NewAIScheduleController creates a new AI schedule controller
func NewAIScheduleController() *AIScheduleController {
	return &AIScheduleController{
		service:      services.NewAIScheduleService(),
		batchService: nil, // Initialized on demand
	}
}

// SetBatchService sets the batch service (called after initialization with DB access)
func (c *AIScheduleController) SetBatchService(batchService *services.AIScheduleBatchService) {
	c.batchService = batchService
}

// GenerateSchedule generates a production schedule from natural language
// POST /api/ai/schedule/generate
func (c *AIScheduleController) GenerateSchedule(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIScheduleController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("AIScheduleController.GenerateSchedule", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	// Parse request
	var req services.GenerateScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	iLog.Info(fmt.Sprintf("Generate schedule request - Prompt: %s, Resources: %d", req.Prompt, len(req.AvailableResources)))

	// Generate schedule
	response, err := c.service.GenerateSchedule(ctx.Request.Context(), req)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to generate schedule: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate schedule: %v", err)})
		return
	}

	iLog.Info(fmt.Sprintf("Schedule generated successfully - %d tasks created", len(response.Tasks)))
	ctx.JSON(http.StatusOK, response)
}

// OptimizeSchedule optimizes an existing schedule
// POST /api/ai/schedule/optimize
func (c *AIScheduleController) OptimizeSchedule(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIScheduleController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("AIScheduleController.OptimizeSchedule", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	// Parse request
	var req services.OptimizeScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	iLog.Info(fmt.Sprintf("Optimize schedule request - Objective: %s, Tasks: %d", req.Objective, len(req.CurrentTasks)))

	// Optimize schedule
	response, err := c.service.OptimizeSchedule(ctx.Request.Context(), req)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to optimize schedule: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to optimize schedule: %v", err)})
		return
	}

	iLog.Info(fmt.Sprintf("Schedule optimized successfully - %d tasks, %d changes", len(response.Tasks), len(response.Changes)))
	ctx.JSON(http.StatusOK, response)
}

// Chat handles conversational questions about schedules
// POST /api/ai/schedule/chat
func (c *AIScheduleController) Chat(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIScheduleController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("AIScheduleController.Chat", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	// Parse request
	var req services.ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	iLog.Info(fmt.Sprintf("Chat request - Question: %s", req.Question))

	// Process chat
	response, err := c.service.Chat(ctx.Request.Context(), req)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to process chat: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process chat: %v", err)})
		return
	}

	iLog.Info(fmt.Sprintf("Chat processed successfully"))
	ctx.JSON(http.StatusOK, response)
}

// OptimizeScheduleBatch starts an async batch optimization job
// POST /api/ai/schedule/optimize-batch
func (c *AIScheduleController) OptimizeScheduleBatch(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIScheduleController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("AIScheduleController.OptimizeScheduleBatch", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	if c.batchService == nil {
		iLog.Error("Batch service not initialized")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Batch service not available"})
		return
	}

	// Parse request
	var req services.OptimizeScheduleBatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	iLog.Info(fmt.Sprintf("Optimize schedule batch request - Objective: %s, Tasks: %d, BatchSize: %d",
		req.Objective, len(req.CurrentTasks), req.BatchSize))

	// Start batch optimization
	response, err := c.batchService.OptimizeScheduleBatch(ctx.Request.Context(), req, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to start batch optimization: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to start batch optimization: %v", err)})
		return
	}

	iLog.Info(fmt.Sprintf("Batch optimization started - JobID: %s, Batches: %d", response.JobID, response.TotalBatches))
	ctx.JSON(http.StatusAccepted, response)
}

// GetBatchJobStatus retrieves the status of a batch job
// GET /api/ai/schedule/batch/:jobId/status
func (c *AIScheduleController) GetBatchJobStatus(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIScheduleController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("AIScheduleController.GetBatchJobStatus", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	if c.batchService == nil {
		iLog.Error("Batch service not initialized")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Batch service not available"})
		return
	}

	jobID := ctx.Param("jobId")
	if jobID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	iLog.Info(fmt.Sprintf("Get batch job status - JobID: %s", jobID))

	// Get status
	status, err := c.batchService.GetBatchJobStatus(ctx.Request.Context(), jobID)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get batch job status: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get job status: %v", err)})
		return
	}

	iLog.Info(fmt.Sprintf("Batch job status - JobID: %s, Status: %s, Progress: %d%%", jobID, status.Status, status.Progress))
	ctx.JSON(http.StatusOK, status)
}

// GetBatchJobResult retrieves the final result of a completed batch job
// GET /api/ai/schedule/batch/:jobId/result
func (c *AIScheduleController) GetBatchJobResult(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIScheduleController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("AIScheduleController.GetBatchJobResult", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	if c.batchService == nil {
		iLog.Error("Batch service not initialized")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Batch service not available"})
		return
	}

	jobID := ctx.Param("jobId")
	if jobID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	iLog.Info(fmt.Sprintf("Get batch job result - JobID: %s", jobID))

	// Get result
	result, err := c.batchService.GetBatchJobResult(ctx.Request.Context(), jobID)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get batch job result: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get job result: %v", err)})
		return
	}

	iLog.Info(fmt.Sprintf("Batch job result retrieved - JobID: %s, Tasks: %d, Status: %s", jobID, len(result.Tasks), result.Status))
	ctx.JSON(http.StatusOK, result)
}
