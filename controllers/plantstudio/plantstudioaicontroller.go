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

package plantstudio

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/services"
)

// PlantStudioAIController handles AI-powered 3D generation for Plant Studio
type PlantStudioAIController struct {
	aiService *services.PlantStudioAIService
}

// NewPlantStudioAIController creates a new Plant Studio AI controller
func NewPlantStudioAIController() *PlantStudioAIController {
	return &PlantStudioAIController{
		aiService: services.NewPlantStudioAIService(),
	}
}

// GenerateFromText generates 3D models from text description
// POST /api/plant-studio/ai/generate-from-text
func (c *PlantStudioAIController) GenerateFromText(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PlantStudioAIController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("PlantStudioAIController.GenerateFromText", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	// Parse request body
	var req services.TextTo3DRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Description == "" {
		iLog.Error("Description is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}

	iLog.Info(fmt.Sprintf("Generating 3D models from text: %s", req.Description))

	// Generate 3D models
	response, err := c.aiService.GenerateFromText(ctx.Request.Context(), req)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to generate 3D from text: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate 3D models: %v", err)})
		return
	}

	iLog.Info(fmt.Sprintf("Successfully generated %d 3D models from text", len(response.Assets)))
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response.Assets,
		"count":   len(response.Assets),
	})
}

// GenerateFromImage generates a 3D model from an image
// POST /api/plant-studio/ai/generate-from-image
func (c *PlantStudioAIController) GenerateFromImage(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PlantStudioAIController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("PlantStudioAIController.GenerateFromImage", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	// Parse request body
	var req services.ImageTo3DRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Image == "" {
		iLog.Error("Image is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Image (base64 encoded) is required"})
		return
	}

	iLog.Info("Generating 3D model from image")

	// Generate 3D model
	response, err := c.aiService.GenerateFromImage(ctx.Request.Context(), req)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to generate 3D from image: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate 3D model: %v", err)})
		return
	}

	iLog.Info("Successfully generated 3D model from image")
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response.Asset,
	})
}
