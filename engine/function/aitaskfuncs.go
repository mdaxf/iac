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

package funcs

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// Global AI config accessor function - to be set during initialization
var GetAIConfigFunc func() interface{}

// AIConfig represents the AI configuration structure
type AIConfig struct {
	UseCases  map[string]AIUseCase  `json:"use_cases"`
	AIVendors map[string]AIVendor   `json:"ai_vendors"`
}

type AIUseCase struct {
	Vendor      string                 `json:"vendor"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type AIVendor struct {
	Enabled    bool                   `json:"enabled"`
	APIKey     string                 `json:"api_key"`
	Models     map[string]string      `json:"models"`
	Parameters map[string]interface{} `json:"parameters"`
}

// AITaskFuncs handles the execution of AI tasks in workflows
type AITaskFuncs struct {
}

// AITaskConfig represents the configuration for an AI task
type AITaskConfig struct {
	UseCaseName   string              `json:"useCaseName,omitempty"`
	AIVendor      string              `json:"aiVendor,omitempty"`
	Agent         string              `json:"agent,omitempty"`
	Skills        []string            `json:"skills,omitempty"`
	MCPServers    []string            `json:"mcpServers,omitempty"`
	SystemPrompt  string              `json:"systemPrompt,omitempty"`
	UserPrompt    string              `json:"userPrompt,omitempty"`
	InputMapping  []AIInputMapping    `json:"inputMapping,omitempty"`
	OutputMapping []AIOutputMapping   `json:"outputMapping,omitempty"`
	Temperature   float64             `json:"temperature,omitempty"`
	MaxTokens     int                 `json:"maxTokens,omitempty"`
	Timeout       int                 `json:"timeout,omitempty"`
	RetryAttempts int                 `json:"retryAttempts,omitempty"`
}

// AIInputMapping represents how to map workflow inputs to AI prompt variables
type AIInputMapping struct {
	ID               string `json:"id"`
	InputName        string `json:"inputName"`
	PromptVariable   string `json:"promptVariable"`
	Transformation   string `json:"transformation,omitempty"`
}

// AIOutputMapping represents how to extract outputs from AI response
type AIOutputMapping struct {
	ID             string `json:"id"`
	OutputName     string `json:"outputName"`
	AIResponsePath string `json:"aiResponsePath"`
	Transformation string `json:"transformation,omitempty"`
	DefaultValue   string `json:"defaultValue,omitempty"`
}

// Execute executes an AI task function
// It reads the AI task configuration from function.mapdata,
// collects input values from function inputs,
// calls the AI service, and sets the outputs
func (a *AITaskFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.function.AITaskFuncs.Execute", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			errMsg := fmt.Sprintf("Error executing AI task function '%s': %v", f.Fobj.Name, err)
			f.iLog.Error(errMsg)
			f.ErrorMessage = errMsg
			f.CancelExecution(errMsg)
		}
	}()

	f.iLog.Info(fmt.Sprintf("Executing AI task function: %s", f.Fobj.Name))

	// Parse AI task configuration from mapdata
	var aiConfig AITaskConfig
	if f.Fobj.Mapdata != nil {
		// Try to extract from mapdata
		configBytes, err := json.Marshal(f.Fobj.Mapdata)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to marshal AI task config: %v", err)
			f.iLog.Error(errMsg)
			f.ErrorMessage = errMsg
			return
		}

		err = json.Unmarshal(configBytes, &aiConfig)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to unmarshal AI task config: %v", err)
			f.iLog.Error(errMsg)
			f.ErrorMessage = errMsg
			return
		}
	} else {
		errMsg := "AI task configuration is required in mapdata"
		f.iLog.Error(errMsg)
		f.ErrorMessage = errMsg
		return
	}

	// Collect input values
	inputValues := make(map[string]interface{})
	for _, input := range f.Fobj.Inputs {
		// Get the actual value from the input
		var value interface{}

		// Check if value is already set in input.Value
		if input.Value != "" {
			value = input.Value
		} else if input.Inivalue != "" {
			value = input.Inivalue
		} else if input.Defaultvalue != "" {
			value = input.Defaultvalue
		}

		// Try to parse as JSON if it's a string that looks like JSON
		if strValue, ok := value.(string); ok {
			if len(strValue) > 0 && (strValue[0] == '{' || strValue[0] == '[') {
				var jsonValue interface{}
				if err := json.Unmarshal([]byte(strValue), &jsonValue); err == nil {
					value = jsonValue
				}
			}
		}

		inputValues[input.Name] = value
		f.iLog.Debug(fmt.Sprintf("AI task input '%s' = %v", input.Name, value))
	}

	// Execute AI task directly
	response, err := a.executeAITask(f, aiConfig, inputValues)
	if err != nil {
		errMsg := fmt.Sprintf("AI task execution failed: %v", err)
		f.iLog.Error(errMsg)
		f.ErrorMessage = errMsg
		return
	}

	f.iLog.Info("AI task executed successfully")

	// Extract outputs from AI response
	outputValues, err := a.extractOutputs(f, aiConfig, response)
	if err != nil {
		f.iLog.Warn(fmt.Sprintf("Failed to extract outputs: %v", err))
	}

	// Create output map for this execution
	outputs := make(map[string]interface{})
	outputs["_ai_raw_response"] = response

	// Set output values
	for _, output := range f.Fobj.Outputs {
		if value, exists := outputValues[output.Name]; exists {
			// Convert value to string for storage in output.Value
			var strValue string
			switch v := value.(type) {
			case string:
				strValue = v
			case nil:
				strValue = ""
			default:
				// Convert complex types to JSON string
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					f.iLog.Warn(fmt.Sprintf("Failed to marshal output '%s': %v", output.Name, err))
					strValue = fmt.Sprintf("%v", v)
				} else {
					strValue = string(jsonBytes)
				}
			}

			output.Value = strValue
			outputs[output.Name] = value
			f.iLog.Debug(fmt.Sprintf("AI task output '%s' = %v", output.Name, value))
		} else {
			// Use default value if output not provided by AI
			if output.Defaultvalue != "" {
				output.Value = output.Defaultvalue
				outputs[output.Name] = output.Defaultvalue
			} else {
				output.Value = ""
				outputs[output.Name] = nil
			}
			f.iLog.Warn(fmt.Sprintf("AI task did not provide output '%s', using default", output.Name))
		}
	}

	// Append outputs to FunctionOutputs slice
	f.FunctionOutputs = append(f.FunctionOutputs, outputs)

	f.iLog.Debug(fmt.Sprintf("AI task function completed with outputs: %v", outputs))
}

// executeAITask executes the AI task with retry logic
func (a *AITaskFuncs) executeAITask(f *Funcs, aiConfig AITaskConfig, inputValues map[string]interface{}) (string, error) {
	// Set default retry attempts if not specified
	retryAttempts := aiConfig.RetryAttempts
	if retryAttempts <= 0 {
		retryAttempts = 1
	}

	var lastError error
	var response string

	// Retry loop
	for attempt := 1; attempt <= retryAttempts; attempt++ {
		if attempt > 1 {
			f.iLog.Info(fmt.Sprintf("Retry attempt %d of %d", attempt, retryAttempts))
		}

		response, lastError = a.executeAITaskOnce(f, aiConfig, inputValues)
		if lastError == nil {
			return response, nil
		}

		if attempt < retryAttempts {
			// Wait before retry (exponential backoff)
			waitTime := time.Duration(attempt) * 2 * time.Second
			f.iLog.Info(fmt.Sprintf("Waiting %v before retry", waitTime))
			time.Sleep(waitTime)
		}
	}

	return "", fmt.Errorf("failed after %d attempts: %v", retryAttempts, lastError)
}

// executeAITaskOnce executes the AI task once (without retry logic)
func (a *AITaskFuncs) executeAITaskOnce(f *Funcs, aiConfig AITaskConfig, inputValues map[string]interface{}) (string, error) {
	// Get AI configuration using the global accessor
	if GetAIConfigFunc == nil {
		return "", fmt.Errorf("AI config accessor not initialized")
	}

	rawConfig := GetAIConfigFunc()
	configBytes, err := json.Marshal(rawConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal AI config: %v", err)
	}

	var aiConfigData AIConfig
	if err := json.Unmarshal(configBytes, &aiConfigData); err != nil {
		return "", fmt.Errorf("failed to unmarshal AI config: %v", err)
	}

	// Determine which use case to use
	useCaseName := aiConfig.UseCaseName
	if useCaseName == "" {
		useCaseName = "workflow_ai_task" // default use case
	}

	useCase, exists := aiConfigData.UseCases[useCaseName]
	if !exists {
		return "", fmt.Errorf("AI use case '%s' not configured", useCaseName)
	}

	// Determine vendor (allow config override)
	vendor := aiConfig.AIVendor
	if vendor == "" {
		vendor = useCase.Vendor
	}
	if vendor == "" {
		vendor = "openai" // fallback default
	}

	vendorConfig, exists := aiConfigData.AIVendors[vendor]
	if !exists || !vendorConfig.Enabled {
		return "", fmt.Errorf("AI vendor '%s' not configured or disabled", vendor)
	}

	// Build the prompt with input mappings
	userPrompt, err := a.buildPrompt(f, aiConfig, inputValues)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt: %v", err)
	}

	// Get system prompt (config override or use case default)
	systemPrompt := aiConfig.SystemPrompt
	if systemPrompt == "" {
		if sp, ok := useCase.Parameters["system_prompt"].(string); ok {
			systemPrompt = sp
		}
	}

	f.iLog.Debug("Calling AI vendor: " + vendor)

	// Call the appropriate vendor
	var rawResponse string
	if vendor == "openai" {
		rawResponse, err = a.callOpenAI(f, systemPrompt, userPrompt, aiConfig, useCase, vendorConfig)
	} else {
		err = fmt.Errorf("vendor %s not yet supported for AI tasks", vendor)
	}

	if err != nil {
		return "", fmt.Errorf("AI vendor call failed: %v", err)
	}

	return rawResponse, nil
}

// buildPrompt builds the user prompt by replacing variables with input values
func (a *AITaskFuncs) buildPrompt(f *Funcs, aiConfig AITaskConfig, inputValues map[string]interface{}) (string, error) {
	prompt := aiConfig.UserPrompt
	if prompt == "" {
		return "", fmt.Errorf("user prompt is required")
	}

	// Create a map of prompt variables to their values
	variables := make(map[string]string)

	for _, mapping := range aiConfig.InputMapping {
		// Get the input value
		value, exists := inputValues[mapping.InputName]
		if !exists {
			f.iLog.Warn(fmt.Sprintf("Input value '%s' not found, using empty string", mapping.InputName))
			value = ""
		}

		// Apply transformation if specified
		var valueStr string
		if mapping.Transformation != "" {
			// For now, just convert to JSON string
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				valueStr = fmt.Sprintf("%v", value)
			} else {
				valueStr = string(jsonBytes)
			}
		} else {
			valueStr = fmt.Sprintf("%v", value)
		}

		variables[mapping.PromptVariable] = valueStr
	}

	// Replace {{variableName}} patterns in the prompt
	for varName, varValue := range variables {
		pattern := fmt.Sprintf("{{%s}}", varName)
		prompt = strings.ReplaceAll(prompt, pattern, varValue)
	}

	// Check for any remaining unreplaced variables and warn
	re := regexp.MustCompile(`{{(\w+)}}`)
	unreplaced := re.FindAllStringSubmatch(prompt, -1)
	if len(unreplaced) > 0 {
		varNames := make([]string, len(unreplaced))
		for i, match := range unreplaced {
			varNames[i] = match[1]
		}
		f.iLog.Warn(fmt.Sprintf("Unreplaced variables in prompt: %v", varNames))
	}

	return prompt, nil
}

// callOpenAI calls the OpenAI API to get a completion
func (a *AITaskFuncs) callOpenAI(f *Funcs, systemPrompt, userPrompt string, taskConfig AITaskConfig, useCase AIUseCase, vendorConfig AIVendor) (string, error) {
	apiKey := vendorConfig.APIKey
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// Apply timeout if specified
	if taskConfig.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(taskConfig.Timeout)*time.Second)
		defer cancel()
	}

	// Determine model
	model := ""
	if m, ok := vendorConfig.Models["chatbot"]; ok {
		model = m
	}
	if model == "" {
		model = "gpt-4o" // fallback default
	}

	// Determine temperature
	temperature := taskConfig.Temperature
	if temperature == 0 {
		if temp, ok := vendorConfig.Parameters["temperature"].(float64); ok {
			temperature = temp
		} else {
			temperature = 0.7 // fallback default
		}
	}

	// Determine max tokens
	maxTokens := taskConfig.MaxTokens
	if maxTokens == 0 {
		if tokens, ok := useCase.Parameters["max_tokens"].(float64); ok {
			maxTokens = int(tokens)
		} else if tokens, ok := useCase.Parameters["max_tokens"].(int); ok {
			maxTokens = tokens
		} else {
			maxTokens = 3000 // fallback default
		}
	}

	// Build messages
	messages := []openai.ChatCompletionMessage{}

	if systemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userPrompt,
	})

	f.iLog.Debug(fmt.Sprintf("Calling OpenAI with model: %s, temperature: %.2f, max_tokens: %d", model, temperature, maxTokens))

	req := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: float32(temperature),
		MaxTokens:   maxTokens,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	content := resp.Choices[0].Message.Content
	f.iLog.Debug(fmt.Sprintf("Received response from OpenAI (%d tokens)", resp.Usage.CompletionTokens))

	return content, nil
}

// extractOutputs extracts output values from the AI response using the output mappings
func (a *AITaskFuncs) extractOutputs(f *Funcs, aiConfig AITaskConfig, rawResponse string) (map[string]interface{}, error) {
	outputs := make(map[string]interface{})

	// If no output mappings, return the raw response as a single "result" output
	if len(aiConfig.OutputMapping) == 0 {
		outputs["result"] = rawResponse
		return outputs, nil
	}

	// Try to parse response as JSON
	var responseData map[string]interface{}
	err := json.Unmarshal([]byte(rawResponse), &responseData)
	if err != nil {
		// If response is not JSON, use it as plain text for all outputs
		f.iLog.Warn("AI response is not valid JSON, using as plain text")
		for _, mapping := range aiConfig.OutputMapping {
			if mapping.DefaultValue != "" {
				outputs[mapping.OutputName] = mapping.DefaultValue
			} else {
				outputs[mapping.OutputName] = rawResponse
			}
		}
		return outputs, nil
	}

	// Extract each output using the specified path
	for _, mapping := range aiConfig.OutputMapping {
		value, err := a.extractValueFromJSON(responseData, mapping.AIResponsePath)
		if err != nil || value == nil {
			// Use default value if extraction fails
			if mapping.DefaultValue != "" {
				outputs[mapping.OutputName] = mapping.DefaultValue
			} else {
				f.iLog.Warn(fmt.Sprintf("Failed to extract output '%s' from path '%s'", mapping.OutputName, mapping.AIResponsePath))
				outputs[mapping.OutputName] = nil
			}
		} else {
			outputs[mapping.OutputName] = value
		}
	}

	return outputs, nil
}

// extractValueFromJSON extracts a value from JSON data using a simple path notation
func (a *AITaskFuncs) extractValueFromJSON(data map[string]interface{}, path string) (interface{}, error) {
	// For simple paths without dots or brackets, direct lookup
	if !strings.Contains(path, ".") && !strings.Contains(path, "[") {
		// Remove leading $ if present (JSONPath notation)
		path = strings.TrimPrefix(path, "$.")
		path = strings.TrimPrefix(path, "$")

		if value, exists := data[path]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("path '%s' not found in response", path)
	}

	// For complex paths, parse step by step
	path = strings.TrimPrefix(path, "$.")
	path = strings.TrimPrefix(path, "$")

	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		// Navigate to the next level
		if currentMap, ok := current.(map[string]interface{}); ok {
			if value, exists := currentMap[part]; exists {
				current = value
			} else {
				return nil, fmt.Errorf("path '%s' not found at part '%s'", path, part)
			}
		} else {
			return nil, fmt.Errorf("cannot navigate path '%s' at part '%s', not a map", path, part)
		}
	}

	return current, nil
}
