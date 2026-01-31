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

package services

import (
	"context"
	"fmt"

	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
	openai "github.com/sashabaranov/go-openai"
)

// GlobalMapAIService handles AI-powered analysis for Global Map
type GlobalMapAIService struct {
	iLog logger.Log
}

// PlantKPIs represents the key performance indicators for a plant
type PlantKPIs struct {
	Efficiency struct {
		Value  float64 `json:"value"`
		Unit   string  `json:"unit"`
		Target float64 `json:"target"`
	} `json:"efficiency"`
	Output struct {
		Value  float64 `json:"value"`
		Unit   string  `json:"unit"`
		Target float64 `json:"target"`
	} `json:"output"`
	Safety struct {
		Value  float64 `json:"value"`
		Unit   string  `json:"unit"`
		Target float64 `json:"target"`
	} `json:"safety"`
	Energy struct {
		Value  float64 `json:"value"`
		Unit   string  `json:"unit"`
		Target float64 `json:"target"`
	} `json:"energy"`
}

// PlantAnalysisRequest represents a request to analyze a single plant
type PlantAnalysisRequest struct {
	Name   string    `json:"name"`
	Type   string    `json:"type"`
	Status string    `json:"status"`
	KPIs   PlantKPIs `json:"kpis"`
}

// PlantAnalysisResponse represents the AI analysis response for a single plant
type PlantAnalysisResponse struct {
	Insight string `json:"insight"`
}

// FleetPlant represents a simplified plant data for fleet analysis
type FleetPlant struct {
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	Efficiency float64 `json:"efficiency"`
}

// FleetAnalysisRequest represents a request to analyze the global fleet
type FleetAnalysisRequest struct {
	Plants []FleetPlant `json:"plants"`
}

// FleetAnalysisResponse represents the AI analysis response for the fleet
type FleetAnalysisResponse struct {
	Insight string `json:"insight"`
}

// AnalyzePlant generates an AI-powered analysis for a single plant
func (s *GlobalMapAIService) AnalyzePlant(request PlantAnalysisRequest) (PlantAnalysisResponse, error) {
	s.iLog = logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "GlobalMapAIService"}
	s.iLog.Info("Starting plant analysis for: " + request.Name)

	// Get AI configuration
	aiConfig := config.GetAIConfig()
	useCase, exists := aiConfig.UseCases["global_map_plant_analysis"]
	if !exists {
		err := fmt.Errorf("AI use case 'global_map_plant_analysis' not configured")
		s.iLog.Error(err.Error())
		return PlantAnalysisResponse{}, err
	}

	// Get vendor config
	vendor := useCase.Vendor
	if vendor == "" {
		vendor = "openai" // default fallback
	}

	vendorConfig, exists := aiConfig.AIVendors[vendor]
	if !exists || !vendorConfig.Enabled {
		err := fmt.Errorf("AI vendor '%s' not configured or disabled", vendor)
		s.iLog.Error(err.Error())
		return PlantAnalysisResponse{}, err
	}

	// Build the prompt
	prompt := fmt.Sprintf(`
Analyze the following industrial plant KPI data and provide a concise, executive summary of its performance.
Highlight any critical issues (like low efficiency or safety incidents) or successes.
Keep it under 100 words.

Plant Name: %s
Type: %s
Status: %s
KPIs:
- Efficiency: %.1f%s (Target: %.1f)
- Output: %.0f %s (Target: %.0f)
- Safety: %.0f days incident free (Target: %.0f)
- Energy: %.1f%s (Target: %.1f)
`,
		request.Name,
		request.Type,
		request.Status,
		request.KPIs.Efficiency.Value, request.KPIs.Efficiency.Unit, request.KPIs.Efficiency.Target,
		request.KPIs.Output.Value, request.KPIs.Output.Unit, request.KPIs.Output.Target,
		request.KPIs.Safety.Value, request.KPIs.Safety.Target,
		request.KPIs.Energy.Value, request.KPIs.Energy.Unit, request.KPIs.Energy.Target,
	)

	s.iLog.Debug("Generated prompt for plant analysis")

	// Call OpenAI API (or configured vendor)
	if vendor == "openai" {
		return s.callOpenAI(prompt, useCase, vendorConfig)
	}

	return PlantAnalysisResponse{}, fmt.Errorf("vendor %s not yet supported for global map", vendor)
}

// AnalyzeFleet generates an AI-powered analysis for the global fleet
func (s *GlobalMapAIService) AnalyzeFleet(request FleetAnalysisRequest) (FleetAnalysisResponse, error) {
	s.iLog = logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "GlobalMapAIService"}
	s.iLog.Info(fmt.Sprintf("Starting fleet analysis for %d plants", len(request.Plants)))

	// Get AI configuration
	aiConfig := config.GetAIConfig()
	useCase, exists := aiConfig.UseCases["global_map_fleet_analysis"]
	if !exists {
		err := fmt.Errorf("AI use case 'global_map_fleet_analysis' not configured")
		s.iLog.Error(err.Error())
		return FleetAnalysisResponse{}, err
	}

	// Get vendor config
	vendor := useCase.Vendor
	if vendor == "" {
		vendor = "openai" // default fallback
	}

	vendorConfig, exists := aiConfig.AIVendors[vendor]
	if !exists || !vendorConfig.Enabled {
		err := fmt.Errorf("AI vendor '%s' not configured or disabled", vendor)
		s.iLog.Error(err.Error())
		return FleetAnalysisResponse{}, err
	}

	// Build plant data string
	plantsData := ""
	for _, plant := range request.Plants {
		plantsData += fmt.Sprintf("%s: Status=%s, Eff=%.0f%%\n", plant.Name, plant.Status, plant.Efficiency)
	}

	// Build the prompt
	prompt := fmt.Sprintf(`
You are a Global Operations Director. Provide a 2-sentence high-level summary of the global fleet status based on these plants.
Identify the highest performing and lowest performing sites.

Plants Data:
%s
`, plantsData)

	s.iLog.Debug("Generated prompt for fleet analysis")

	// Call OpenAI API (or configured vendor)
	if vendor == "openai" {
		return s.callOpenAIFleet(prompt, useCase, vendorConfig)
	}

	return FleetAnalysisResponse{}, fmt.Errorf("vendor %s not yet supported for global map", vendor)
}

// callOpenAI makes the actual API call to OpenAI for plant analysis
func (s *GlobalMapAIService) callOpenAI(prompt string, useCase config.UseCaseConfig, vendorConfig config.AIVendorConfig) (PlantAnalysisResponse, error) {
	apiKey := vendorConfig.APIKey
	if apiKey == "" {
		err := fmt.Errorf("OpenAI API key not configured")
		s.iLog.Error(err.Error())
		return PlantAnalysisResponse{}, err
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// Prepare messages
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: useCase.SystemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	// Get temperature and max tokens from use case or vendor config
	temperature := useCase.Temperature
	if temperature == 0 {
		if temp, ok := vendorConfig.Parameters["temperature"].(float64); ok {
			temperature = temp
		} else {
			temperature = 0.7 // Default
		}
	}

	maxTokens := useCase.MaxTokens
	if maxTokens == 0 {
		maxTokens = 300 // Default for this use case
	}

	// Get model
	model := useCase.ModelOverride
	if model == "" {
		if m, ok := vendorConfig.Models["chatbot"]; ok {
			model = m
		}
	}
	if model == "" {
		model = openai.GPT4o // Default fallback
	}

	s.iLog.Debug(fmt.Sprintf("Calling OpenAI with model: %s, temp: %.2f, max_tokens: %d", model, temperature, maxTokens))

	// Make API call
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: float32(temperature),
			MaxTokens:   maxTokens,
		},
	)

	if err != nil {
		s.iLog.Error(fmt.Sprintf("OpenAI API error: %v", err))
		return PlantAnalysisResponse{}, fmt.Errorf("failed to get AI response: %w", err)
	}

	if len(resp.Choices) == 0 {
		err := fmt.Errorf("no response from AI")
		s.iLog.Error(err.Error())
		return PlantAnalysisResponse{}, err
	}

	insight := resp.Choices[0].Message.Content
	s.iLog.Info("Successfully generated plant analysis")

	return PlantAnalysisResponse{
		Insight: insight,
	}, nil
}

// callOpenAIFleet makes the actual API call to OpenAI for fleet analysis
func (s *GlobalMapAIService) callOpenAIFleet(prompt string, useCase config.UseCaseConfig, vendorConfig config.AIVendorConfig) (FleetAnalysisResponse, error) {
	apiKey := vendorConfig.APIKey
	if apiKey == "" {
		err := fmt.Errorf("OpenAI API key not configured")
		s.iLog.Error(err.Error())
		return FleetAnalysisResponse{}, err
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// Prepare messages
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: useCase.SystemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	// Get temperature and max tokens from use case or vendor config
	temperature := useCase.Temperature
	if temperature == 0 {
		if temp, ok := vendorConfig.Parameters["temperature"].(float64); ok {
			temperature = temp
		} else {
			temperature = 0.7 // Default
		}
	}

	maxTokens := useCase.MaxTokens
	if maxTokens == 0 {
		maxTokens = 200 // Default for this use case
	}

	// Get model
	model := useCase.ModelOverride
	if model == "" {
		if m, ok := vendorConfig.Models["chatbot"]; ok {
			model = m
		}
	}
	if model == "" {
		model = openai.GPT4o // Default fallback
	}

	s.iLog.Debug(fmt.Sprintf("Calling OpenAI with model: %s, temp: %.2f, max_tokens: %d", model, temperature, maxTokens))

	// Make API call
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: float32(temperature),
			MaxTokens:   maxTokens,
		},
	)

	if err != nil {
		s.iLog.Error(fmt.Sprintf("OpenAI API error: %v", err))
		return FleetAnalysisResponse{}, fmt.Errorf("failed to get AI response: %w", err)
	}

	if len(resp.Choices) == 0 {
		err := fmt.Errorf("no response from AI")
		s.iLog.Error(err.Error())
		return FleetAnalysisResponse{}, err
	}

	insight := resp.Choices[0].Message.Content
	s.iLog.Info("Successfully generated fleet analysis")

	return FleetAnalysisResponse{
		Insight: insight,
	}, nil
}
