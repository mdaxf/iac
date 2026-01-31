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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
	openai "github.com/sashabaranov/go-openai"
)

// AIScheduleService handles AI-powered production scheduling
type AIScheduleService struct {
	iLog logger.Log
}

// Resource represents a manufacturing resource
type Resource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Role string `json:"role"`
}

// Task represents a production task
type Task struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	ResourceIDs      []string     `json:"resourceIds"`
	Start            time.Time    `json:"start"`
	End              time.Time    `json:"end"`
	Progress         int          `json:"progress"`
	Status           string       `json:"status"`
	Description      string       `json:"description"`
	SetupMinutes     int          `json:"setupMinutes,omitempty"`
	CycleTimeMinutes int          `json:"cycleTimeMinutes,omitempty"`
	Requirements     []string     `json:"requirements,omitempty"`
	Badges           []Badge      `json:"badges,omitempty"`
	Dependencies     []Dependency `json:"dependencies,omitempty"`
}

// Badge represents a task badge
type Badge struct {
	Label string `json:"label"`
	Color string `json:"color"`
}

// Dependency represents a task dependency
type Dependency struct {
	PredecessorID string `json:"predecessorId"`
	Type          string `json:"type"` // FS, SS, FF, SF
	Lag           int    `json:"lag"`
}

// GenerateScheduleRequest represents a schedule generation request
type GenerateScheduleRequest struct {
	Prompt             string                 `json:"prompt"`
	AvailableResources []Resource             `json:"availableResources"`
	CalendarConfig     map[string]interface{} `json:"calendarConfig,omitempty"`
	SystemPrompt       string                 `json:"systemPrompt,omitempty"`
}

// GenerateScheduleResponse represents a schedule generation response
type GenerateScheduleResponse struct {
	Tasks    []Task                 `json:"tasks"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// OptimizeScheduleRequest represents a schedule optimization request
type OptimizeScheduleRequest struct {
	CurrentTasks []Task                 `json:"currentTasks"`
	Resources    []Resource             `json:"resources"`
	Objective    string                 `json:"objective"`
	Constraints  map[string]interface{} `json:"constraints,omitempty"`
	SystemPrompt string                 `json:"systemPrompt,omitempty"`
}

// OptimizeScheduleResponse represents a schedule optimization response
type OptimizeScheduleResponse struct {
	Tasks    []Task                 `json:"tasks"`
	Changes  []ScheduleChange       `json:"changes"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ScheduleChange represents a change made during optimization
type ScheduleChange struct {
	TaskID   string      `json:"taskId"`
	Field    string      `json:"field"`
	OldValue interface{} `json:"oldValue"`
	NewValue interface{} `json:"newValue"`
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Question string                 `json:"question"`
	Context  map[string]interface{} `json:"context"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Answer string `json:"answer"`
}

// NewAIScheduleService creates a new AI schedule service
func NewAIScheduleService() *AIScheduleService {
	return &AIScheduleService{
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "AIScheduleService",
		},
	}
}

// GenerateSchedule generates a production schedule from natural language
func (s *AIScheduleService) GenerateSchedule(ctx context.Context, req GenerateScheduleRequest) (*GenerateScheduleResponse, error) {
	s.iLog.Info(fmt.Sprintf("GenerateSchedule START - Prompt: %s", req.Prompt))

	// Get AI configuration for schedule_generation use case
	aiConfig := config.GetAIConfig()
	if aiConfig == nil {
		return nil, fmt.Errorf("AI configuration not loaded")
	}

	// Get use case config or use chatbot as fallback
	useCaseName := "schedule_generation"
	useCaseConfig, exists := aiConfig.UseCases[useCaseName]
	if !exists {
		s.iLog.Warn(fmt.Sprintf("Use case '%s' not found, falling back to 'chatbot'", useCaseName))
		useCaseConfig = aiConfig.UseCases["chatbot"]
		useCaseName = "chatbot"
	}

	vendor := useCaseConfig.Vendor
	vendorConfig, exists := aiConfig.AIVendors[vendor]
	if !exists || !vendorConfig.Enabled {
		return nil, fmt.Errorf("vendor '%s' not found or disabled", vendor)
	}

	// Get model for this use case
	model, exists := vendorConfig.Models[useCaseName]
	if !exists {
		model = vendorConfig.Models["chatbot"]
		s.iLog.Warn(fmt.Sprintf("Model for use case '%s' not found, using chatbot model: %s", useCaseName, model))
	}

	// Use model override if specified
	if useCaseConfig.ModelOverride != "" {
		model = useCaseConfig.ModelOverride
	}

	s.iLog.Info(fmt.Sprintf("Using vendor: %s, model: %s", vendor, model))

	// Build system prompt
	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = useCaseConfig.SystemPrompt
	}

	// Call appropriate AI vendor
	var tasks []Task
	var err error

	switch vendor {
	case "openai":
		tasks, err = s.generateScheduleOpenAI(ctx, vendorConfig, model, systemPrompt, req)
	case "google":
		tasks, err = s.generateScheduleGoogle(ctx, vendorConfig, model, systemPrompt, req)
	case "anthropic":
		tasks, err = s.generateScheduleAnthropic(ctx, vendorConfig, model, systemPrompt, req)
	default:
		return nil, fmt.Errorf("vendor '%s' not supported for schedule generation", vendor)
	}

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to generate schedule: %v", err))
		return nil, err
	}

	s.iLog.Info(fmt.Sprintf("GenerateSchedule COMPLETE - Generated %d tasks", len(tasks)))

	return &GenerateScheduleResponse{
		Tasks: tasks,
		Metadata: map[string]interface{}{
			"model":      model,
			"vendor":     vendor,
			"confidence": 0.85,
		},
	}, nil
}

// generateScheduleOpenAI generates schedule using OpenAI
func (s *AIScheduleService) generateScheduleOpenAI(ctx context.Context, vendorConfig config.AIVendorConfig, model, systemPrompt string, req GenerateScheduleRequest) ([]Task, error) {
	client := openai.NewClient(vendorConfig.APIKey)

	// Build user prompt
	userPrompt := fmt.Sprintf("User Request: %s\n\nAvailable Resources: %s",
		req.Prompt,
		s.formatResources(req.AvailableResources))

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	}

	// Get temperature from vendor config or default
	temperature := 0.7
	if temp, ok := vendorConfig.Parameters["temperature"].(float64); ok {
		temperature = temp
	}

	// Create chat completion request
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: float32(temperature),
			MaxTokens:   3000,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse response
	content := resp.Choices[0].Message.Content
	s.iLog.Debug(fmt.Sprintf("OpenAI response: %s", content))

	// Extract JSON from markdown code blocks if present
	content = s.extractJSON(content)

	var result struct {
		Tasks []Task `json:"tasks"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// Try parsing as direct array
		var tasks []Task
		if err2 := json.Unmarshal([]byte(content), &tasks); err2 != nil {
			return nil, fmt.Errorf("failed to parse OpenAI response: %w (original: %v)", err2, err)
		}
		return tasks, nil
	}

	return result.Tasks, nil
}

// generateScheduleGoogle generates schedule using Google Gemini
func (s *AIScheduleService) generateScheduleGoogle(ctx context.Context, vendorConfig config.AIVendorConfig, model, systemPrompt string, req GenerateScheduleRequest) ([]Task, error) {
	// TODO: Implement Google Gemini integration
	// For now, return error indicating not implemented
	return nil, fmt.Errorf("Google Gemini integration not yet implemented - please use OpenAI vendor")
}

// generateScheduleAnthropic generates schedule using Anthropic Claude
func (s *AIScheduleService) generateScheduleAnthropic(ctx context.Context, vendorConfig config.AIVendorConfig, model, systemPrompt string, req GenerateScheduleRequest) ([]Task, error) {
	// TODO: Implement Anthropic Claude integration
	// For now, return error indicating not implemented
	return nil, fmt.Errorf("Anthropic Claude integration not yet implemented - please use OpenAI vendor")
}

// OptimizeSchedule optimizes an existing schedule
func (s *AIScheduleService) OptimizeSchedule(ctx context.Context, req OptimizeScheduleRequest) (*OptimizeScheduleResponse, error) {
	s.iLog.Info(fmt.Sprintf("OptimizeSchedule START - Objective: %s", req.Objective))

	// Get AI configuration
	aiConfig := config.GetAIConfig()
	if aiConfig == nil {
		return nil, fmt.Errorf("AI configuration not loaded")
	}

	// Get use case config or use chatbot as fallback
	useCaseName := "schedule_optimization"
	useCaseConfig, exists := aiConfig.UseCases[useCaseName]
	if !exists {
		s.iLog.Warn(fmt.Sprintf("Use case '%s' not found, falling back to 'chatbot'", useCaseName))
		useCaseConfig = aiConfig.UseCases["chatbot"]
		useCaseName = "chatbot"
	}

	vendor := useCaseConfig.Vendor
	vendorConfig := aiConfig.AIVendors[vendor]
	model := vendorConfig.Models[useCaseName]
	if model == "" {
		model = vendorConfig.Models["chatbot"]
	}

	// Use model override if specified
	if useCaseConfig.ModelOverride != "" {
		model = useCaseConfig.ModelOverride
	}

	// Build system prompt
	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = useCaseConfig.SystemPrompt
	}

	// Call appropriate AI vendor
	var tasks []Task
	var changes []ScheduleChange
	var err error

	switch vendor {
	case "openai":
		tasks, changes, err = s.optimizeScheduleOpenAI(ctx, vendorConfig, model, systemPrompt, req)
	case "google":
		tasks, changes, err = s.optimizeScheduleGoogle(ctx, vendorConfig, model, systemPrompt, req)
	case "anthropic":
		tasks, changes, err = s.optimizeScheduleAnthropic(ctx, vendorConfig, model, systemPrompt, req)
	default:
		return nil, fmt.Errorf("vendor '%s' not supported for schedule optimization", vendor)
	}

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to optimize schedule: %v", err))
		return nil, err
	}

	s.iLog.Info(fmt.Sprintf("OptimizeSchedule COMPLETE - Optimized %d tasks with %d changes", len(tasks), len(changes)))

	return &OptimizeScheduleResponse{
		Tasks:   tasks,
		Changes: changes,
		Metadata: map[string]interface{}{
			"model":             model,
			"vendor":            vendor,
			"optimizationScore": 0.85,
		},
	}, nil
}

// optimizeScheduleOpenAI optimizes schedule using OpenAI
func (s *AIScheduleService) optimizeScheduleOpenAI(ctx context.Context, vendorConfig config.AIVendorConfig, model, systemPrompt string, req OptimizeScheduleRequest) ([]Task, []ScheduleChange, error) {
	client := openai.NewClient(vendorConfig.APIKey)

	// Build user prompt
	tasksJSON, _ := json.Marshal(req.CurrentTasks)
	resourcesJSON, _ := json.Marshal(req.Resources)

	userPrompt := fmt.Sprintf("Optimization Objective: %s\n\nCurrent Schedule:\n%s\n\nAvailable Resources:\n%s",
		req.Objective,
		string(tasksJSON),
		string(resourcesJSON))

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	}

	// Get temperature from vendor config or default
	temperature := 0.3 // Lower for optimization
	if temp, ok := vendorConfig.Parameters["temperature"].(float64); ok {
		temperature = temp
	}

	// Calculate appropriate max tokens based on model
	maxTokens := 16000 // Default for modern models (gpt-4o, gpt-4-turbo)

	// Adjust for specific models
	if strings.Contains(model, "gpt-3.5") {
		maxTokens = 4000
	} else if strings.Contains(model, "gpt-4o") || strings.Contains(model, "gpt-4-turbo") {
		maxTokens = 16000 // These models support up to 128k context
	}

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
		return nil, nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse response
	content := resp.Choices[0].Message.Content
	s.iLog.Debug(fmt.Sprintf("OpenAI optimization response (length: %d chars): %s", len(content), content))

	// Check if response was truncated
	finishReason := resp.Choices[0].FinishReason
	if finishReason == "length" {
		s.iLog.Warn("OpenAI response was truncated due to max_tokens limit - response may be incomplete")
		return nil, nil, fmt.Errorf("response truncated: AI response exceeded max_tokens limit. Please use batch processing for large task lists (30+ tasks)")
	}

	// Extract JSON from markdown code blocks if present
	content = s.extractJSON(content)

	var result struct {
		Tasks    []Task           `json:"tasks"`
		Changes  []ScheduleChange `json:"changes"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to parse JSON response (length: %d). Content preview: %s", len(content), content[:min(500, len(content))]))
		return nil, nil, fmt.Errorf("failed to parse OpenAI response: %w. Hint: Response may be truncated. Try batch processing for large task lists", err)
	}

	return result.Tasks, result.Changes, nil
}

// optimizeScheduleGoogle optimizes schedule using Google Gemini
func (s *AIScheduleService) optimizeScheduleGoogle(ctx context.Context, vendorConfig config.AIVendorConfig, model, systemPrompt string, req OptimizeScheduleRequest) ([]Task, []ScheduleChange, error) {
	// TODO: Implement Google Gemini integration
	return nil, nil, fmt.Errorf("Google Gemini integration not yet implemented - please use OpenAI vendor")
}

// optimizeScheduleAnthropic optimizes schedule using Anthropic Claude
func (s *AIScheduleService) optimizeScheduleAnthropic(ctx context.Context, vendorConfig config.AIVendorConfig, model, systemPrompt string, req OptimizeScheduleRequest) ([]Task, []ScheduleChange, error) {
	// TODO: Implement Anthropic Claude integration
	return nil, nil, fmt.Errorf("Anthropic Claude integration not yet implemented - please use OpenAI vendor")
}

// Chat provides conversational AI for schedule questions
func (s *AIScheduleService) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	s.iLog.Info(fmt.Sprintf("Chat START - Question: %s", req.Question))

	// Get AI configuration
	aiConfig := config.GetAIConfig()
	if aiConfig == nil {
		return nil, fmt.Errorf("AI configuration not loaded")
	}

	useCaseConfig := aiConfig.UseCases["chatbot"]
	vendor := useCaseConfig.Vendor
	vendorConfig := aiConfig.AIVendors[vendor]
	model := vendorConfig.Models["chatbot"]

	// Use model override if specified
	if useCaseConfig.ModelOverride != "" {
		model = useCaseConfig.ModelOverride
	}

	// Build context
	contextJSON, _ := json.Marshal(req.Context)
	systemPrompt := fmt.Sprintf("You are a production scheduling assistant. Answer questions about the schedule.\n\nContext: %s", string(contextJSON))

	client := openai.NewClient(vendorConfig.APIKey)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: req.Question,
		},
	}

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: 0.7,
			MaxTokens:   1000,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	answer := resp.Choices[0].Message.Content
	s.iLog.Info(fmt.Sprintf("Chat COMPLETE - Answer: %s", answer))

	return &ChatResponse{
		Answer: answer,
	}, nil
}

// Helper functions

func (s *AIScheduleService) formatResources(resources []Resource) string {
	result := ""
	for i, r := range resources {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s (%s, %s)", r.Name, r.Type, r.Role)
	}
	return result
}

func (s *AIScheduleService) extractJSON(content string) string {
	// Remove markdown code blocks if present
	if idx := strings.Index(content, "```json"); idx >= 0 {
		content = content[idx+7:]
		if idx := strings.Index(content, "```"); idx >= 0 {
			content = content[:idx]
		}
	} else if idx := strings.Index(content, "```"); idx >= 0 {
		content = content[idx+3:]
		if idx := strings.Index(content, "```"); idx >= 0 {
			content = content[:idx]
		}
	}
	return strings.TrimSpace(content)
}
