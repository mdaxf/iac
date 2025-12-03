package codegen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdaxf/iac/logger"
)

// Multi-step Page generation workflow
// Step 1: Understand intent (question vs code generation)
// Step 2: Analyze page requirements (layout, components, interactions)
// Step 3: Plan component structure (sections, widgets, forms)
// Step 4: Generate components one by one with proper props and bindings
// Step 5: Combine generated components into final page structure
// Step 6: Return for frontend merge

// PageGenerationPlan represents the plan for generating page code
type PageGenerationPlan struct {
	PageType          string           `json:"pageType"`          // dashboard, form, list, detail, custom
	LayoutType        string           `json:"layoutType"`        // grid, flex, tabs, sections
	ComponentCount    int              `json:"componentCount"`
	SectionCount      int              `json:"sectionCount"`
	Components        []ComponentSpec  `json:"components"`
}

// ComponentSpec represents a single component to be generated
type ComponentSpec struct {
	Name        string `json:"name"`
	Type        string `json:"type"`  // table, form, chart, card, button, input, etc.
	Description string `json:"description"`
	Purpose     string `json:"purpose"`
	Index       int    `json:"index"`
	SectionIndex int   `json:"sectionIndex"`
}

// GeneratedComponent represents a generated component with full details
type GeneratedComponent struct {
	Name         string                   `json:"name"`
	Type         string                   `json:"type"`
	Description  string                   `json:"description"`
	SectionIndex int                      `json:"sectionIndex"`
	Props        map[string]interface{}   `json:"props"`
	DataBinding  map[string]interface{}   `json:"dataBinding,omitempty"`
	Events       []map[string]interface{} `json:"events,omitempty"`
	Validation   map[string]interface{}   `json:"validation,omitempty"`
	Style        map[string]interface{}   `json:"style,omitempty"`
}

// Step1_ClassifyPageIntent determines if user wants to generate page code or ask a question
func Step1_ClassifyPageIntent(question string, apiKey string, model string) (isCodeGeneration bool, confidence float64, err error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PageGenMultiStep"}

	prompt := fmt.Sprintf(`Analyze this user request and determine if they want to GENERATE PAGE CODE or ask a QUESTION.

User request: "%s"

Respond with ONLY this format:
TYPE: code_generation OR question
CONFIDENCE: 0.0-1.0

Rules:
- If user says "generate", "create", "build", "make", "add page", "add components" -> code_generation
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

// Step2_AnalyzePageRequirements analyzes what page components need to be generated
func Step2_AnalyzePageRequirements(question string, apiKey string, model string, currentPage map[string]interface{}) (*PageGenerationPlan, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PageGenMultiStep"}

	contextInfo := buildPageContextInfo(currentPage)

	prompt := fmt.Sprintf(`Analyze this page generation request and create a generation plan.

User request: "%s"

%s

Analyze and respond with JSON:
{
  "pageType": "dashboard|form|list|detail|custom",
  "layoutType": "grid|flex|tabs|sections",
  "componentCount": 5,
  "sectionCount": 1,
  "components": [
    {
      "name": "ComponentName",
      "type": "table|form|chart|card|button|input|select|etc",
      "description": "What this component does",
      "purpose": "Specific purpose",
      "index": 0,
      "sectionIndex": 0
    }
  ]
}

Component types available: table, form, chart, card, button, input, select, textarea, checkbox, radio, datepicker, upload, tabs, panel, grid, list

Return ONLY valid JSON.`, question, contextInfo)

	response, err := CallOpenAI(apiKey, model, []map[string]interface{}{
		{"role": "system", "content": "You are a page structure analyzer. Generate detailed plans."},
		{"role": "user", "content": prompt},
	}, 0.5)

	if err != nil {
		return nil, err
	}

	// Parse JSON response
	response = extractJSONFromMarkdown(response)

	var plan PageGenerationPlan
	err = json.Unmarshal([]byte(response), &plan)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse generation plan: %v", err))
		return nil, fmt.Errorf("failed to parse generation plan: %v", err)
	}

	iLog.Info(fmt.Sprintf("Step 2 - Plan: %d components, %d sections, type: %s",
		plan.ComponentCount, plan.SectionCount, plan.PageType))

	return &plan, nil
}

// Step3_GenerateComponentByComponent generates each component one by one
func Step3_GeneratePageComponentByComponent(plan *PageGenerationPlan, apiKey string, model string) ([]GeneratedComponent, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PageGenMultiStep"}

	generatedComponents := make([]GeneratedComponent, 0, len(plan.Components))

	for _, compSpec := range plan.Components {
		iLog.Debug(fmt.Sprintf("Generating component %d/%d: %s", compSpec.Index+1, len(plan.Components), compSpec.Name))

		// Generate this component with context of previous components
		generatedComp, err := generateSinglePageComponent(compSpec, generatedComponents, apiKey, model)
		if err != nil {
			return nil, fmt.Errorf("failed to generate component %s: %v", compSpec.Name, err)
		}

		generatedComponents = append(generatedComponents, *generatedComp)

		// Small delay to avoid rate limiting
		time.Sleep(200 * time.Millisecond)
	}

	iLog.Info(fmt.Sprintf("Step 3 - Generated %d components", len(generatedComponents)))

	return generatedComponents, nil
}

// generateSinglePageComponent generates a single page component
func generateSinglePageComponent(spec ComponentSpec, previousComponents []GeneratedComponent, apiKey string, model string) (*GeneratedComponent, error) {
	// Build context from previous components
	previousContext := ""
	if len(previousComponents) > 0 {
		previousContext = "\n\nPrevious components generated:\n"
		for _, pc := range previousComponents {
			previousContext += fmt.Sprintf("- %s (%s): %s\n", pc.Name, pc.Type, pc.Description)
			if pc.DataBinding != nil && len(pc.DataBinding) > 0 {
				previousContext += fmt.Sprintf("  Data bindings: %v\n", pc.DataBinding)
			}
		}
	}

	prompt := fmt.Sprintf(`Generate a complete component definition for a page.

Component to generate:
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
  "props": {
    "label": "Component Label",
    "placeholder": "Placeholder text",
    "required": true,
    "disabled": false
  },
  "dataBinding": {
    "source": "api|state|props|form",
    "path": "data.fieldName",
    "bindingType": "value|options|disabled|visible"
  },
  "events": [
    {
      "eventType": "onClick|onChange|onSubmit|onLoad",
      "action": "navigate|submit|validate|loadData",
      "target": "ComponentName|ApiEndpoint",
      "params": {}
    }
  ],
  "validation": {
    "rules": ["required", "email", "minLength:5"],
    "messages": {"required": "This field is required"}
  },
  "style": {
    "width": "100%%",
    "className": "custom-class"
  }
}

Rules:
- Generate realistic props based on component type
- Set up data bindings to APIs or state management
- Define event handlers for user interactions
- Add validation rules for input components
- Use previous components' data where logical

Return ONLY valid JSON.`,
		spec.Name, spec.Type, spec.Description, spec.Purpose, previousContext,
		spec.Name, spec.Type, spec.Description)

	response, err := CallOpenAI(apiKey, model, []map[string]interface{}{
		{"role": "system", "content": "You are a UI component generator. Generate complete, working component definitions."},
		{"role": "user", "content": prompt},
	}, 0.7)

	if err != nil {
		return nil, err
	}

	// Parse JSON response
	response = extractJSONFromMarkdown(response)

	var generatedComp GeneratedComponent
	err = json.Unmarshal([]byte(response), &generatedComp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated component: %v", err)
	}

	// Set section index from spec
	generatedComp.SectionIndex = spec.SectionIndex

	return &generatedComp, nil
}

// Step4_CombinePageComponents combines all generated components into final page structure
func Step4_CombinePageComponents(plan *PageGenerationPlan, components []GeneratedComponent) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PageGenMultiStep"}

	sections := make([]map[string]interface{}, plan.SectionCount)

	// Distribute components into sections
	currentCompIndex := 0
	for sectionIdx := 0; sectionIdx < plan.SectionCount; sectionIdx++ {
		sectionComponents := make([]map[string]interface{}, 0)

		for i := 0; i < len(components); i++ {
			if components[i].SectionIndex == sectionIdx || plan.SectionCount == 1 {
				comp := components[currentCompIndex]

				componentMap := map[string]interface{}{
					"id":          fmt.Sprintf("comp-%d-%d", sectionIdx, i),
					"name":        comp.Name,
					"type":        comp.Type,
					"description": comp.Description,
					"props":       comp.Props,
					"dataBinding": comp.DataBinding,
					"events":      comp.Events,
					"validation":  comp.Validation,
					"style":       comp.Style,
				}

				sectionComponents = append(sectionComponents, componentMap)
				currentCompIndex++
				if currentCompIndex >= len(components) {
					break
				}
			}
		}

		sections[sectionIdx] = map[string]interface{}{
			"name":       fmt.Sprintf("Section_%d", sectionIdx+1),
			"components": sectionComponents,
		}

		if currentCompIndex >= len(components) {
			break
		}
	}

	result := map[string]interface{}{
		"pageType":    plan.PageType,
		"layoutType":  plan.LayoutType,
		"sections":    sections,
	}

	iLog.Info(fmt.Sprintf("Step 4 - Combined %d components into %d sections", len(components), plan.SectionCount))

	return result, nil
}

// GeneratePageMultiStep is the main entry point for multi-step page generation
func GeneratePageMultiStep(ctx context.Context, description string, apiKey string, openaiModel string, currentPage map[string]interface{}) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PageGenMultiStep"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("codegen.GeneratePageMultiStep", elapsed)
	}()

	// Step 1: Classify intent
	isCodeGen, confidence, err := Step1_ClassifyPageIntent(description, apiKey, openaiModel)
	if err != nil {
		return nil, fmt.Errorf("step 1 failed: %v", err)
	}

	if !isCodeGen || confidence < 0.7 {
		return nil, fmt.Errorf("request does not appear to be for code generation (confidence: %.2f)", confidence)
	}

	// Step 2: Analyze requirements and create plan
	plan, err := Step2_AnalyzePageRequirements(description, apiKey, openaiModel, currentPage)
	if err != nil {
		return nil, fmt.Errorf("step 2 failed: %v", err)
	}

	// Step 3: Generate components one by one
	components, err := Step3_GeneratePageComponentByComponent(plan, apiKey, openaiModel)
	if err != nil {
		return nil, fmt.Errorf("step 3 failed: %v", err)
	}

	// Step 4: Combine into final structure
	result, err := Step4_CombinePageComponents(plan, components)
	if err != nil {
		return nil, fmt.Errorf("step 4 failed: %v", err)
	}

	iLog.Info("Multi-step page generation completed successfully")

	return map[string]interface{}{
		"page": result,
		"plan": plan,
	}, nil
}

// Helper function
func buildPageContextInfo(currentPage map[string]interface{}) string {
	if currentPage == nil {
		return ""
	}

	contextInfo := ""
	if name, ok := currentPage["name"].(string); ok && name != "" {
		contextInfo += fmt.Sprintf("\nCurrent Page name: %s", name)
	}
	if sections, ok := currentPage["sections"].([]interface{}); ok && len(sections) > 0 {
		contextInfo += fmt.Sprintf("\nExisting sections: %d", len(sections))
	}

	return contextInfo
}
