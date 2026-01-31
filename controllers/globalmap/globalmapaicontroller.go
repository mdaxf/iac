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

package globalmap

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/services"
)

// GlobalMapAIController handles AI-powered analysis endpoints for Global Map
type GlobalMapAIController struct{}

// NewGlobalMapAIController creates a new instance of GlobalMapAIController
func NewGlobalMapAIController() *GlobalMapAIController {
	return &GlobalMapAIController{}
}

// AnalyzePlant handles POST /api/global-map/ai/analyze-plant
// Analyzes a single plant's KPIs and returns AI-generated insights
func (c *GlobalMapAIController) AnalyzePlant(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GlobalMapAIController.AnalyzePlant"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("globalmap.AnalyzePlant", elapsed)
	}()

	iLog.Info("Received plant analysis request")

	// Parse request body
	var request services.PlantAnalysisRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		iLog.Error(fmt.Sprintf("Invalid request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"message": err.Error(),
		})
		return
	}

	iLog.Debug(fmt.Sprintf("Analyzing plant: %s", request.Name))

	// Create service and call analysis
	aiService := services.GlobalMapAIService{}
	response, err := aiService.AnalyzePlant(request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to analyze plant: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "AI analysis failed",
			"message": err.Error(),
		})
		return
	}

	iLog.Info("Successfully generated plant analysis")

	// Return success response
	ctx.JSON(http.StatusOK, response)
}

// AnalyzeFleet handles POST /api/global-map/ai/analyze-fleet
// Analyzes the global fleet and returns AI-generated insights
func (c *GlobalMapAIController) AnalyzeFleet(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GlobalMapAIController.AnalyzeFleet"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("globalmap.AnalyzeFleet", elapsed)
	}()

	iLog.Info("Received fleet analysis request")

	// Parse request body
	var request services.FleetAnalysisRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		iLog.Error(fmt.Sprintf("Invalid request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"message": err.Error(),
		})
		return
	}

	iLog.Debug(fmt.Sprintf("Analyzing fleet with %d plants", len(request.Plants)))

	// Create service and call analysis
	aiService := services.GlobalMapAIService{}
	response, err := aiService.AnalyzeFleet(request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to analyze fleet: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "AI analysis failed",
			"message": err.Error(),
		})
		return
	}

	iLog.Info("Successfully generated fleet analysis")

	// Return success response
	ctx.JSON(http.StatusOK, response)
}
