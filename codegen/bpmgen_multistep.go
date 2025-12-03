package codegen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdaxf/iac/logger"
)

// Multi-step BPM generation workflow
// Step 1: Understand intent (question vs code generation)
// Step 2: Identify code type and requirements
// Step 3: Plan tasks (routing, function count, function groups)
// Step 4: Generate functions one by one
// Step 5: Combine generated functions into final BPM structure
// Step 6: Return for frontend merge

// GenerationPlan represents the plan for generating BPM code
type GenerationPlan struct {
	HasRouting      bool     `json:"hasRouting"`
	FunctionCount   int      `json:"functionCount"`
	FunctionGroups  int      `json:"functionGroups"`
	FunctionsPerGroup []int  `json:"functionsPerGroup"`
	Functions       []FunctionSpec `json:"functions"`
}

// FunctionSpec represents a single function to be generated
type FunctionSpec struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Purpose     string `json:"purpose"`
	Index       int    `json:"index"`
	GroupIndex  int    `json:"groupIndex"`
}

// GeneratedFunction represents a generated function with full details
type GeneratedFunction struct {
	Name        string                   `json:"name"`
	Type        string                   `json:"type"`
	Description string                   `json:"description"`
	Inputs      []map[string]interface{} `json:"inputs"`
	Outputs     []map[string]interface{} `json:"outputs"`
	Content     string                   `json:"content,omitempty"`     // Stores script, query, or other function-specific content
	MapData     map[string]interface{}   `json:"mapdata,omitempty"`     // For input/output mappings
}

// Step1_ClassifyIntent determines if user wants to generate code or ask a question
func Step1_ClassifyIntent(question string, apiKey string, model string) (isCodeGeneration bool, confidence float64, err error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "BPMGenMultiStep"}

	prompt := fmt.Sprintf(`Analyze this user request and determine if they want to GENERATE CODE or ask a QUESTION.

User request: "%s"

Respond with ONLY this format:
TYPE: code_generation OR question
CONFIDENCE: 0.0-1.0

Rules:
- If user says "generate", "create", "build", "make", "add functions" -> code_generation
- If user asks "how to", "what is", "explain", "help" -> question
- Default to question if unclear`, question)

	response, err := CallOpenAI(apiKey, model, []map[string]interface{}{
		{"role": "system", "content": "You are an intent classifier. Be concise."},
		{"role": "user", "content": prompt},
	}, 0.3)

	if err != nil {
		return false, 0, err
	}

	// Parse response
	isCodeGeneration = false
	confidence = 0.5

	if contains(response, "TYPE: code_generation") {
		isCodeGeneration = true
	}

	// Extract confidence
	confidence = extractConfidence(response)

	iLog.Debug(fmt.Sprintf("Step 1 - Intent: %v, Confidence: %.2f", isCodeGeneration, confidence))

	return isCodeGeneration, confidence, nil
}

// Step2_AnalyzeRequirements analyzes what needs to be generated
func Step2_AnalyzeRequirements(question string, apiKey string, model string, currentTranCode map[string]interface{}) (*GenerationPlan, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "BPMGenMultiStep"}

	contextInfo := buildContextInfo(currentTranCode)

	prompt := fmt.Sprintf(`Analyze this BPM generation request and create a generation plan.

User request: "%s"

%s

Analyze and respond with JSON:
{
  "hasRouting": false,
  "functionCount": 5,
  "functionGroups": 1,
  "functionsPerGroup": [5],
  "functions": [
    {
      "name": "FunctionName",
      "type": "javascript|query|tableinsert|etc",
      "description": "What this function does",
      "purpose": "Specific purpose",
      "index": 0,
      "groupIndex": 0
    }
  ]
}

Rules:
- hasRouting: true only if user explicitly needs branching/conditional logic
- functionGroups: 1 unless hasRouting=true OR functionCount > 8
- If functionCount > 8, split into groups of max 8 functions
- List ALL functions needed with clear purpose for each
- Function types: inputmap, goexpr, javascript, query, storeprocedure, subtrancode, tableinsert, tableupdate, tabledelete, collectioninsert, collectionupdate, collectiondelete, throwerror, sendmessage, sendemail, webservicecall

Return ONLY valid JSON.`, question, contextInfo)

	response, err := CallOpenAI(apiKey, model, []map[string]interface{}{
		{"role": "system", "content": "You are a BPM task analyzer. Generate detailed plans."},
		{"role": "user", "content": prompt},
	}, 0.5)

	if err != nil {
		return nil, err
	}

	// Parse JSON response
	response = extractJSONFromMarkdown(response)

	var plan GenerationPlan
	err = json.Unmarshal([]byte(response), &plan)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse generation plan: %v", err))
		return nil, fmt.Errorf("failed to parse generation plan: %v", err)
	}

	iLog.Info(fmt.Sprintf("Step 2 - Plan: %d functions, %d groups, routing: %v",
		plan.FunctionCount, plan.FunctionGroups, plan.HasRouting))

	return &plan, nil
}

// Step3_GenerateFunctionByFunction generates each function one by one with full context
func Step3_GenerateFunctionByFunction(plan *GenerationPlan, apiKey string, model string) ([]GeneratedFunction, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "BPMGenMultiStep"}

	generatedFunctions := make([]GeneratedFunction, 0, len(plan.Functions))

	for _, funcSpec := range plan.Functions {
		iLog.Debug(fmt.Sprintf("Generating function %d/%d: %s", funcSpec.Index+1, len(plan.Functions), funcSpec.Name))

		// Generate this function with context of previous functions
		generatedFunc, err := generateSingleFunction(funcSpec, generatedFunctions, apiKey, model)
		if err != nil {
			return nil, fmt.Errorf("failed to generate function %s: %v", funcSpec.Name, err)
		}

		generatedFunctions = append(generatedFunctions, *generatedFunc)

		// Small delay to avoid rate limiting
		time.Sleep(200 * time.Millisecond)
	}

	iLog.Info(fmt.Sprintf("Step 3 - Generated %d functions", len(generatedFunctions)))

	return generatedFunctions, nil
}

// generateSingleFunction generates a single function with full details
func generateSingleFunction(spec FunctionSpec, previousFunctions []GeneratedFunction, apiKey string, model string) (*GeneratedFunction, error) {
	// Build context from previous functions
	previousContext := ""
	if len(previousFunctions) > 0 {
		previousContext = "\n\nPrevious functions generated (USE THESE OUTPUTS as inputs where logical):\n"
		for _, pf := range previousFunctions {
			previousContext += fmt.Sprintf("- Function: %s (%s)\n", pf.Name, pf.Type)
			previousContext += fmt.Sprintf("  Description: %s\n", pf.Description)
			if len(pf.Outputs) > 0 {
				previousContext += "  Available Outputs (use as inputs with source:1 and aliasname:\"FunctionName.OutputName\"):\n"
				for _, out := range pf.Outputs {
					outName := out["name"]
					outType := out["datatype"]
					outDesc := out["description"]
					previousContext += fmt.Sprintf("    * %s.%s (%v) - %v\n", pf.Name, outName, outType, outDesc)
				}
			}
		}
		previousContext += "\nIMPORTANT: Connect these outputs to your inputs using source:1 and aliasname where it makes sense!\n"
	}

	prompt := fmt.Sprintf(`Generate a complete function definition for a BPM flow.

Function to generate:
- Name: %s
- Type: %s
- Description: %s
- Purpose: %s
%s

Generate JSON with this EXACT structure:
{
  "name": "%s",
  "type": "%s",
  "description": "%s",
  "inputs": [
    {
      "name": "inputName",
      "datatype": "string|integer|float|bool|datetime|object",
      "source": 0,
      "value": "",
      "aliasname": "",
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
  "content": "Function-specific content (script/query/etc)",
  "mapdata": {}
}

Rules for "content" field (CRITICAL - stores all function-specific data):
- For "javascript" type: Store JavaScript code as a string in "content"
- For "query" type: Store SQL query as a string in "content"
- For "storeprocedure" type: Store procedure call as a string in "content"
- For "goexpr" type: Store Go expression as a string in "content"
- For "tableinsert/tableupdate/tabledelete": Store table name or operation details as string
- For "inputmap" type: Use "mapdata" field for mapping configuration
- For "webservicecall": Store endpoint URL or API details as string

Rules for Input/Output Linkage (CRITICAL - establishes data flow between functions):
- Input "source" values:
  * 0 = Direct value (use "value" field)
  * 1 = From previous function output (MUST set "aliasname" to "PreviousFunctionName.outputName")
  * 2 = From session variable (use "value" field for session key)
  * 3 = From external input (use "value" field for external key)
- When input comes from previous function (source: 1):
  * MUST set "aliasname" to format: "FunctionName.OutputName"
  * Example: "aliasname": "ValidateUser.userId" means get "userId" output from "ValidateUser" function
  * Leave "value" empty when using aliasname
- Always create logical data flow by connecting function outputs to next function inputs
- Example flow: Function1 outputs "userId" -> Function2 input has source:1, aliasname:"Function1.userId"

Other Rules:
- Define realistic inputs and outputs based on function purpose
- ALWAYS use outputs from previous functions as inputs when logical (set source:1 and aliasname)
- Ensure output names are unique and descriptive
- Create complete data flow chain through all functions

Return ONLY valid JSON.`,
		spec.Name, spec.Type, spec.Description, spec.Purpose, previousContext,
		spec.Name, spec.Type, spec.Description)

	response, err := CallOpenAI(apiKey, model, []map[string]interface{}{
		{"role": "system", "content": "You are a BPM function generator. Generate complete, working function definitions."},
		{"role": "user", "content": prompt},
	}, 0.7)

	if err != nil {
		return nil, err
	}

	// Parse JSON response
	response = extractJSONFromMarkdown(response)

	var generatedFunc GeneratedFunction
	err = json.Unmarshal([]byte(response), &generatedFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated function: %v", err)
	}

	return &generatedFunc, nil
}

// Step4_CombineFunctions combines all generated functions into final BPM structure
func Step4_CombineFunctions(plan *GenerationPlan, functions []GeneratedFunction) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "BPMGenMultiStep"}

	functionGroups := make([]map[string]interface{}, plan.FunctionGroups)

	// Distribute functions into groups according to plan
	currentFuncIndex := 0
	for groupIdx := 0; groupIdx < plan.FunctionGroups; groupIdx++ {
		functionsInThisGroup := plan.FunctionsPerGroup[groupIdx]

		groupFunctions := make([]map[string]interface{}, 0, functionsInThisGroup)

		for i := 0; i < functionsInThisGroup && currentFuncIndex < len(functions); i++ {
			fn := functions[currentFuncIndex]

			// Convert to map - keep type as string for frontend processing
			funcMap := map[string]interface{}{
				"id":          fmt.Sprintf("fn-%d-%d", groupIdx, i),
				"name":        fn.Name,
				"functionname": fn.Name,
				"type":        fn.Type, // Keep as string type (javascript, query, etc.) - frontend will convert
				"description": fn.Description,
				"inputs":      fn.Inputs,
				"outputs":     fn.Outputs,
			}

			// Add content if present (stores scripts, queries, etc.)
			if fn.Content != "" {
				funcMap["content"] = fn.Content
			}

			// Add mapdata if present (for input/output mappings)
			if fn.MapData != nil && len(fn.MapData) > 0 {
				funcMap["mapdata"] = fn.MapData
			}

			groupFunctions = append(groupFunctions, funcMap)
			currentFuncIndex++
		}

		functionGroups[groupIdx] = map[string]interface{}{
			"name":        fmt.Sprintf("FunctionGroup_%d", groupIdx+1),
			"description": fmt.Sprintf("Function group %d", groupIdx+1),
			"routing":     plan.HasRouting && groupIdx < plan.FunctionGroups-1,
			"functions":   groupFunctions,
		}
	}

	result := map[string]interface{}{
		"functionGroups": functionGroups,
	}

	iLog.Info(fmt.Sprintf("Step 4 - Combined %d functions into %d groups", len(functions), plan.FunctionGroups))

	return result, nil
}

// GenerateFlowMultiStep is the main entry point for multi-step generation
func GenerateFlowMultiStep(ctx context.Context, description string, apiKey string, openaiModel string, currentTranCode map[string]interface{}) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "BPMGenMultiStep"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("codegen.GenerateFlowMultiStep", elapsed)
	}()

	// Step 1: Classify intent
	isCodeGen, confidence, err := Step1_ClassifyIntent(description, apiKey, openaiModel)
	if err != nil {
		return nil, fmt.Errorf("step 1 failed: %v", err)
	}

	if !isCodeGen || confidence < 0.7 {
		return nil, fmt.Errorf("request does not appear to be for code generation (confidence: %.2f)", confidence)
	}

	// Step 2: Analyze requirements and create plan
	plan, err := Step2_AnalyzeRequirements(description, apiKey, openaiModel, currentTranCode)
	if err != nil {
		return nil, fmt.Errorf("step 2 failed: %v", err)
	}

	// Step 3: Generate functions one by one
	functions, err := Step3_GenerateFunctionByFunction(plan, apiKey, openaiModel)
	if err != nil {
		return nil, fmt.Errorf("step 3 failed: %v", err)
	}

	// Step 4: Combine into final structure
	result, err := Step4_CombineFunctions(plan, functions)
	if err != nil {
		return nil, fmt.Errorf("step 4 failed: %v", err)
	}

	iLog.Info("Multi-step BPM generation completed successfully")

	return map[string]interface{}{
		"flow": result,
		"plan": plan,
	}, nil
}

// Helper functions

func buildContextInfo(currentTranCode map[string]interface{}) string {
	if currentTranCode == nil {
		return ""
	}

	contextInfo := ""
	if name, ok := currentTranCode["name"].(string); ok && name != "" {
		contextInfo += fmt.Sprintf("\nCurrent TranCode name: %s", name)
	}
	if functionGroups, ok := currentTranCode["functiongroups"].([]interface{}); ok && len(functionGroups) > 0 {
		contextInfo += fmt.Sprintf("\nExisting function groups: %d", len(functionGroups))
	}

	return contextInfo
}

func contains(text string, substr string) bool {
	return len(text) >= len(substr) && findSubstring(text, substr) >= 0
}

func findSubstring(text string, substr string) int {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func extractConfidence(response string) float64 {
	// Try to extract confidence value from response
	if idx := findSubstring(response, "CONFIDENCE:"); idx >= 0 {
		start := idx + len("CONFIDENCE:")
		end := start
		for end < len(response) && (response[end] >= '0' && response[end] <= '9' || response[end] == '.') {
			end++
		}
		if end > start {
			var conf float64
			fmt.Sscanf(response[start:end], "%f", &conf)
			return conf
		}
	}
	return 0.7 // Default
}

// getFuncTypeFromString converts function type string to integer (matching FunctionType enum)
func getFuncTypeFromString(typeStr string) int {
	typeMap := map[string]int{
		"inputmap":            0,
		"goexpr":              1,
		"javascript":          2,
		"query":               3,
		"storeprocedure":      4,
		"subtrancode":         5,
		"tableinsert":         6,
		"tableupdate":         7,
		"tabledelete":         8,
		"collectioninsert":    9,
		"collectionupdate":    10,
		"collectiondelete":    11,
		"throwerror":          12,
		"sendmessage":         13,
		"sendemail":           14,
		"explodeworkflow":     15,
		"startworkflowtask":   16,
		"completeworkflowtask": 17,
		"sendmessagebykafka":  18,
		"sendmessagebymqtt":   19,
		"sendmessagebyaqmp":   20,
		"webservicecall":      21,
	}

	// Convert to lowercase for case-insensitive matching
	lowerTypeStr := ""
	for _, c := range typeStr {
		if c >= 'A' && c <= 'Z' {
			lowerTypeStr += string(c + 32) // Convert to lowercase
		} else {
			lowerTypeStr += string(c)
		}
	}

	if val, ok := typeMap[lowerTypeStr]; ok {
		return val
	}

	return 2 // Default to javascript
}
