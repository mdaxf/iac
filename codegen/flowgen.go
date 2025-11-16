package codegen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mdaxf/iac/logger"
)

var s_OPENAI_FLOW_SYSTEM_PROMPT = `You are an expert in business process modeling and workflow design. Your role is to analyze business requirements and generate structured BPM (Business Process Management) flows.

When given a description of business logic requirements, you should:
1. Break down the requirements into logical function groups
2. Identify the sequence and routing logic between function groups
3. Determine the functions needed within each function group
4. Specify inputs and outputs for each function
5. Define any conditional routing logic

For each business requirement, generate a JSON response with the following structure:
{
  "functionGroups": [
    {
      "name": "FunctionGroupName",
      "description": "What this function group does",
      "routing": false,
      "functions": [
        {
          "name": "FunctionName",
          "type": "javascript|query|tableinsert|tableupdate|tabledelete",
          "description": "What this function does",
          "inputs": [
            {
              "name": "inputName",
              "datatype": "string|integer|float|bool|datetime|object",
              "source": 0,
              "value": "",
              "description": "Input description"
            }
          ],
          "outputs": [
            {
              "name": "outputName",
              "datatype": "string|integer|float|bool|datetime|object",
              "description": "Output description"
            }
          ],
          "content": "SQL query or table name if applicable",
          "script": "JavaScript code if type is javascript"
        }
      ],
      "routerdef": {
        "variable": "variableName",
        "vartype": "string|integer|bool",
        "values": ["value1", "value2"],
        "nextfuncgroups": ["NextFG1", "NextFG2"],
        "defaultfuncgroup": "DefaultFG"
      }
    }
  ]
}

Available function types:
- "inputmap": Input mapping
- "goexpr": Go expression
- "javascript": JavaScript execution
- "query": Database query (SELECT)
- "storeprocedure": Stored procedure call
- "subtrancode": Call another TranCode
- "tableinsert": Insert into database table
- "tableupdate": Update database table
- "tabledelete": Delete from database table
- "collectioninsert": Insert into NoSQL collection
- "collectionupdate": Update NoSQL collection
- "collectiondelete": Delete from NoSQL collection
- "throwerror": Throw an error
- "sendmessage": Send a message
- "sendemail": Send an email
- "webservicecall": Call a web service

Data types:
- "string": String value
- "integer": Integer number
- "float": Floating point number
- "bool": Boolean true/false
- "datetime": Date and time
- "object": Complex object/JSON

Routing logic:
- Set "routing": true if the function group needs conditional routing
- In "routerdef", specify the variable to check, possible values, and next function groups
- Always specify a "defaultfuncgroup" for cases not matching any value

Best practices:
- Use descriptive names for function groups and functions
- Keep function groups focused on a single responsibility
- Use appropriate function types for the task
- Define clear inputs and outputs
- Include validation steps when needed
- Consider error handling paths

Return ONLY valid JSON. Do not include any explanations or markdown formatting.`

// GenerateFlowFromDescription generates a BPM flow structure from a text description using OpenAI
func GenerateFlowFromDescription(description string, apiKey string, openaiModel string, currentTranCode map[string]interface{}) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "FlowGenAI"}
	result := make(map[string]interface{})

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("codegen.GenerateFlowFromDescription", elapsed)
	}()

	if apiKey == "" {
		iLog.Error("API key is required")
		return result, errors.New("API key is required")
	}

	if openaiModel == "" {
		iLog.Error("OpenAI model is required")
		return result, errors.New("OpenAI model is required")
	}

	if description == "" {
		iLog.Error("Description is required")
		return result, errors.New("Description is required")
	}

	// Build user prompt with context
	userPrompt := fmt.Sprintf("Generate a BPM flow for the following business requirement:\n\n%s", description)

	// Add context from current TranCode if provided
	if currentTranCode != nil {
		contextInfo := ""
		if name, ok := currentTranCode["name"].(string); ok && name != "" {
			contextInfo += fmt.Sprintf("\nCurrent TranCode name: %s", name)
		}
		if inputs, ok := currentTranCode["inputs"].([]interface{}); ok && len(inputs) > 0 {
			contextInfo += fmt.Sprintf("\nExisting inputs: %d defined", len(inputs))
		}
		if outputs, ok := currentTranCode["outputs"].([]interface{}); ok && len(outputs) > 0 {
			contextInfo += fmt.Sprintf("\nExisting outputs: %d defined", len(outputs))
		}
		if sessionVars, ok := currentTranCode["sessionVariables"].([]interface{}); ok && len(sessionVars) > 0 {
			contextInfo += fmt.Sprintf("\nExisting session variables: %d defined", len(sessionVars))
		}

		if contextInfo != "" {
			userPrompt += fmt.Sprintf("\n\nContext from current TranCode:%s", contextInfo)
		}
	}

	userPrompt += "\n\nGenerate the flow structure in JSON format as specified in the system prompt."

	// Create messages for OpenAI API
	messages := []map[string]interface{}{
		{
			"role":    "system",
			"content": s_OPENAI_FLOW_SYSTEM_PROMPT,
		},
		{
			"role":    "user",
			"content": userPrompt,
		},
	}

	// Create request
	requestBody := GPT4VCompletionRequest{
		Model:       openaiModel,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   4096,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error marshaling request: %v", err))
		return result, err
	}

	iLog.Debug(fmt.Sprintf("Request to OpenAI: %s", string(jsonBody)))

	// Make HTTP request to OpenAI
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		iLog.Error(fmt.Sprintf("Error creating request: %v", err))
		return result, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error making request: %v", err))
		return result, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading response: %v", err))
		return result, err
	}

	iLog.Debug(fmt.Sprintf("Response from OpenAI: %s", string(body)))

	if resp.StatusCode != http.StatusOK {
		iLog.Error(fmt.Sprintf("OpenAI API error: %s", string(body)))
		return result, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	// Parse OpenAI response
	var openAIResp GPT4VCompletionResponse
	err = json.Unmarshal(body, &openAIResp)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling response: %v", err))
		return result, err
	}

	if len(openAIResp.Choices) == 0 {
		iLog.Error("No choices returned from OpenAI")
		return result, errors.New("No choices returned from OpenAI")
	}

	// Extract the generated flow JSON from the response
	content := openAIResp.Choices[0].Message.Content

	iLog.Debug(fmt.Sprintf("Generated flow content: %s", content))

	// Parse the generated JSON
	var flowData map[string]interface{}
	err = json.Unmarshal([]byte(content), &flowData)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error parsing generated flow JSON: %v", err))
		// Try to extract JSON from markdown code blocks
		content = extractJSONFromMarkdown(content)
		err = json.Unmarshal([]byte(content), &flowData)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error parsing extracted JSON: %v", err))
			return result, fmt.Errorf("Error parsing generated flow: %v", err)
		}
	}

	result["flow"] = flowData

	iLog.Info("Successfully generated BPM flow from description")

	return result, nil
}

// extractJSONFromMarkdown attempts to extract JSON from markdown code blocks
func extractJSONFromMarkdown(content string) string {
	// Try to find JSON in ```json ... ``` or ``` ... ``` blocks
	start := -1
	end := -1

	// Look for ```json
	jsonStart := "```json"
	codeStart := "```"
	codeEnd := "```"

	if idx := find(content, jsonStart); idx != -1 {
		start = idx + len(jsonStart)
	} else if idx := find(content, codeStart); idx != -1 {
		start = idx + len(codeStart)
	}

	if start != -1 {
		// Find the closing ```
		if idx := findAfter(content, codeEnd, start); idx != -1 {
			end = idx
			return content[start:end]
		}
	}

	// If no code blocks found, return as is
	return content
}

func find(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func findAfter(s, substr string, after int) int {
	for i := after; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
