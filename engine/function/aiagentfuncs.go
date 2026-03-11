package funcs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// AIAgentFuncs handles AI Agent and Skills execution
type AIAgentFuncs struct{}

// AIAgentFuncConfig represents the configuration for an AI Agent/Skills call
type AIAgentFuncConfig struct {
	// ExecutionMode controls which execution path is taken:
	//   "agent" — look up the agent by AgentID/AgentName and run it via the agent runner (ReAct loop)
	//   "skill" — call each skill in Skills directly via the tool registry
	//   ""      — legacy: build prompts and call the LLM directly
	ExecutionMode string `json:"executionMode,omitempty"`
	// AgentID is the UUID of the agent definition (preferred over AgentName)
	AgentID      string              `json:"agentId,omitempty"`
	AgentName    string              `json:"agentName,omitempty"`
	Skills       []string            `json:"skills,omitempty"`
	AIModel      string              `json:"aiModel,omitempty"`
	AIVendor     string              `json:"aiVendor,omitempty"`
	SystemPrompt string              `json:"systemPrompt,omitempty"`
	UserPrompt   string              `json:"userPrompt,omitempty"`
	Temperature  float64             `json:"temperature,omitempty"`
	MaxTokens    int                 `json:"maxTokens,omitempty"`
	Timeout      int                 `json:"timeout,omitempty"`
	InputData    map[string]interface{} `json:"inputData,omitempty"`
}

// Execute executes an AI Agent/Skills function
// It reads the AI agent configuration from function.mapdata,
// collects input values from function inputs,
// calls the AI agent/skills service, and sets the outputs
func (a *AIAgentFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("AIAgentFuncs.Execute", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("Recovered from panic in AI Agent execution: %v", r)
			f.iLog.Error(errMsg)
			f.ErrorMessage = errMsg
		}
	}()

	f.iLog.Info(fmt.Sprintf("Start AI Agent Function: %s", f.Fobj.Name))

	// Extract AI agent configuration from function mapdata
	var agentConfig AIAgentFuncConfig
	if f.Fobj.Mapdata != nil {
		configBytes, err := json.Marshal(f.Fobj.Mapdata)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to marshal AI agent config: %v", err)
			f.iLog.Error(errMsg)
			f.ErrorMessage = errMsg
			return
		}

		err = json.Unmarshal(configBytes, &agentConfig)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to unmarshal AI agent config: %v", err)
			f.iLog.Error(errMsg)
			f.ErrorMessage = errMsg
			return
		}
	}

	// Collect input values
	inputValues := make(map[string]interface{})
	for _, input := range f.Fobj.Inputs {
		var value interface{}

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
		f.iLog.Debug(fmt.Sprintf("AI agent input '%s' = %v", input.Name, value))
	}

	agentConfig.InputData = inputValues

	f.iLog.Debug(fmt.Sprintf("AI Agent Config: %+v", agentConfig))

	// Choose execution path based on ExecutionMode
	var response string
	var execErr error

	switch agentConfig.ExecutionMode {
	case "agent":
		// Run a named agent from the catalog via the AgentRunnerService (full ReAct loop)
		f.iLog.Info(fmt.Sprintf("AI Agent node: running catalog agent id=%s name=%s", agentConfig.AgentID, agentConfig.AgentName))
		response, execErr = a.executeViaAgentRuntime(f, agentConfig, inputValues)
	case "skill":
		// Call individual skills from the tool registry directly
		f.iLog.Info(fmt.Sprintf("AI Agent node: executing skills %v", agentConfig.Skills))
		response, execErr = a.executeViaSkills(f, agentConfig, inputValues)
	default:
		// Legacy path: build prompts and call LLM directly
		response, execErr = a.executeAIAgent(f, agentConfig)
	}

	if execErr != nil {
		errMsg := fmt.Sprintf("AI Agent execution failed: %v", execErr)
		f.iLog.Error(errMsg)
		f.ErrorMessage = errMsg
		return
	}

	f.iLog.Info("AI Agent execution completed successfully")

	// Set outputs
	outputs := make(map[string]interface{})
	outputs["_ai_agent_response"] = response

	for i, output := range f.Fobj.Outputs {
		if i == 0 {
			// First output gets the full response
			output.Value = response
			outputs[output.Name] = response
		} else {
			// Try to parse response as JSON and extract specific fields
			var responseData map[string]interface{}
			if err := json.Unmarshal([]byte(response), &responseData); err == nil {
				if value, exists := responseData[output.Name]; exists {
					strValue := fmt.Sprintf("%v", value)
					output.Value = strValue
					outputs[output.Name] = value
				}
			}
		}
		f.iLog.Debug(fmt.Sprintf("AI agent output '%s' = %v", output.Name, output.Value))
	}

	// Append outputs to FunctionOutputs slice
	f.FunctionOutputs = append(f.FunctionOutputs, outputs)

	f.iLog.Info(fmt.Sprintf("End AI Agent Function: %s", f.Fobj.Name))
}

// executeViaAgentRuntime uses the injected RunAgentSyncFunc to run a catalog agent end-to-end.
// The agent's own system prompt, skills, and MCP servers defined in Agent Studio are used;
// the user prompt is built from agentConfig.UserPrompt with {{field}} substitution.
func (a *AIAgentFuncs) executeViaAgentRuntime(f *Funcs, agentConfig AIAgentFuncConfig, inputValues map[string]interface{}) (string, error) {
	if RunAgentSyncFunc == nil {
		return "", fmt.Errorf("agent runtime not initialized (RunAgentSyncFunc is nil)")
	}

	// Build the task prompt
	prompt := agentConfig.UserPrompt
	if prompt == "" {
		inputJSON, _ := json.MarshalIndent(inputValues, "", "  ")
		prompt = fmt.Sprintf("Process the following data:\n\n%s", string(inputJSON))
	} else {
		for key, val := range inputValues {
			prompt = strings.ReplaceAll(prompt, fmt.Sprintf("{{%s}}", key), fmt.Sprintf("%v", val))
		}
	}

	ctx := context.Background()
	return RunAgentSyncFunc(ctx, agentConfig.AgentID, agentConfig.AgentName, prompt, agentConfig.Timeout)
}

// executeViaSkills calls each skill in agentConfig.Skills sequentially via the injected
// ExecuteSkillFunc and concatenates the results. Input values are passed as JSON arguments.
func (a *AIAgentFuncs) executeViaSkills(f *Funcs, agentConfig AIAgentFuncConfig, inputValues map[string]interface{}) (string, error) {
	if ExecuteSkillFunc == nil {
		return "", fmt.Errorf("skill executor not initialized (ExecuteSkillFunc is nil)")
	}
	if len(agentConfig.Skills) == 0 {
		return "", fmt.Errorf("no skills configured for skill execution mode")
	}

	argsJSON, _ := json.Marshal(inputValues)

	ctx := context.Background()
	if agentConfig.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(agentConfig.Timeout)*time.Second)
		defer cancel()
	}

	results := make([]string, 0, len(agentConfig.Skills))
	for _, skillName := range agentConfig.Skills {
		result, err := ExecuteSkillFunc(ctx, skillName, string(argsJSON))
		if err != nil {
			f.iLog.Error(fmt.Sprintf("Skill '%s' error: %v", skillName, err))
			results = append(results, fmt.Sprintf("Skill %s error: %v", skillName, err))
		} else {
			results = append(results, result)
		}
	}
	return strings.Join(results, "\n---\n"), nil
}

// executeAIAgent executes the AI agent/skills call
func (a *AIAgentFuncs) executeAIAgent(f *Funcs, agentConfig AIAgentFuncConfig) (string, error) {
	// Get AI configuration
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

	// Determine which AI vendor to use
	vendor := agentConfig.AIVendor
	if vendor == "" {
		vendor = "openai" // default vendor
	}

	vendorConfig, exists := aiConfigData.AIVendors[vendor]
	if !exists || !vendorConfig.Enabled {
		return "", fmt.Errorf("AI vendor '%s' not configured or disabled", vendor)
	}

	// Determine which model to use
	aiModel := agentConfig.AIModel
	if aiModel == "" {
		// Try to get default model from vendor config
		if defaultModel, ok := vendorConfig.Models["default"]; ok {
			aiModel = defaultModel
		} else {
			// Fallback to a common default
			aiModel = "gpt-3.5-turbo"
		}
	}

	f.iLog.Info(fmt.Sprintf("Using AI vendor: %s, model: %s", vendor, aiModel))

	// Build system prompt
	systemPrompt := agentConfig.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant"
	}

	// Add agent name and skills to system prompt
	if agentConfig.AgentName != "" {
		systemPrompt = fmt.Sprintf("You are an AI agent named '%s'. %s", agentConfig.AgentName, systemPrompt)
	}

	if len(agentConfig.Skills) > 0 {
		skillsList := strings.Join(agentConfig.Skills, ", ")
		systemPrompt = fmt.Sprintf("%s\n\nYou have access to the following skills: %s", systemPrompt, skillsList)
	}

	// Build user prompt from template and input data
	userPrompt := agentConfig.UserPrompt
	if userPrompt == "" {
		// Create a default prompt from input data
		inputJSON, _ := json.MarshalIndent(agentConfig.InputData, "", "  ")
		userPrompt = fmt.Sprintf("Process the following input data:\n\n%s", string(inputJSON))
	} else {
		// Replace {{variableName}} placeholders with input values
		for key, value := range agentConfig.InputData {
			placeholder := fmt.Sprintf("{{%s}}", key)
			valueStr := fmt.Sprintf("%v", value)
			userPrompt = strings.ReplaceAll(userPrompt, placeholder, valueStr)
		}
	}

	f.iLog.Debug(fmt.Sprintf("System prompt: %s", systemPrompt))
	f.iLog.Debug(fmt.Sprintf("User prompt: %s", userPrompt))

	// Call AI service based on vendor
	var response string
	if vendor == "openai" || vendor == "azure_openai" {
		response, err = a.callOpenAI(f, systemPrompt, userPrompt, agentConfig, vendorConfig, aiModel)
	} else {
		return "", fmt.Errorf("AI vendor '%s' not supported yet", vendor)
	}

	if err != nil {
		return "", err
	}

	return response, nil
}

// callOpenAI calls the OpenAI API
func (a *AIAgentFuncs) callOpenAI(f *Funcs, systemPrompt, userPrompt string, agentConfig AIAgentFuncConfig, vendorConfig AIVendor, modelName string) (string, error) {
	apiKey := vendorConfig.APIKey
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// Apply timeout if specified
	if agentConfig.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(agentConfig.Timeout)*time.Second)
		defer cancel()
	}

	// Prepare messages
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

	// Set temperature
	temperature := float32(0.7)
	if agentConfig.Temperature > 0 {
		temperature = float32(agentConfig.Temperature)
	}

	// Set max tokens
	maxTokens := 3000
	if agentConfig.MaxTokens > 0 {
		maxTokens = agentConfig.MaxTokens
	}

	f.iLog.Debug(fmt.Sprintf("Calling OpenAI with model: %s, temperature: %.2f, maxTokens: %d", modelName, temperature, maxTokens))

	// Create completion request
	req := openai.ChatCompletionRequest{
		Model:       modelName,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	// Call OpenAI API
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	response := resp.Choices[0].Message.Content

	f.iLog.Info(fmt.Sprintf("OpenAI API call successful, tokens used: %d", resp.Usage.TotalTokens))

	return response, nil
}
