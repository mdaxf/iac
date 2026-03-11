package services

// Codegen tools — allow agents to generate IAC entities from natural language:
// BPM (TranCode), Workflow, Page, View, Whiteboard.
// Each tool is a thin adapter over the corresponding codegen package function.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mdaxf/iac/codegen"
)

// registerCodegenTools registers all AI code generation tools
func (s *AgentToolRegistryService) registerCodegenTools() {
	s.registerBPMCodegenTool()
	s.registerWorkflowCodegenTool()
	s.registerPageCodegenTool()
	s.registerViewCodegenTool()
	s.registerWhiteboardCodegenTool()
}

// ─── BPM ──────────────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerBPMCodegenTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "generate_bpm",
			Description: "Generates a BPM transaction code (TranCode) structure from a natural language description using multi-step AI. " +
				"Produces a complete function graph with inputs, outputs, and data-flow bindings ready for the IAC BPM editor. " +
				"Describe the business process step by step — the AI handles function type selection, routing, and linkage.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"description": {
						Type:        "string",
						Description: "Natural language description of the business process or transaction to generate (required). " +
							"Example: 'Create a purchase order approval flow: validate PO data, check budget, send approval email, update order status'.",
					},
					"current_trancode_json": {
						Type:        "string",
						Description: "Optional. Existing TranCode JSON to enhance or extend. The AI will merge new functions with existing ones.",
					},
				},
				Required: []string{"description"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.openAIKey == "" {
			return "", fmt.Errorf("OpenAI API key is not configured; generate_bpm requires AI capabilities")
		}
		description := stringArg(args, "description")
		if description == "" {
			return "", fmt.Errorf("description is required")
		}

		var currentTranCode map[string]interface{}
		if v := stringArg(args, "current_trancode_json"); v != "" {
			json.Unmarshal([]byte(v), &currentTranCode) //nolint:errcheck
		}

		result, err := codegen.GenerateFlowMultiStep(ctx, description, s.openAIKey, s.openAIModel, currentTranCode)
		if err != nil {
			return "", fmt.Errorf("BPM generation failed: %w", err)
		}

		b, _ := json.Marshal(result)
		return string(b), nil
	})
}

// ─── WORKFLOW ─────────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerWorkflowCodegenTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "generate_workflow",
			Description: "Generates a workflow diagram (start, task, gateway, subflow, end nodes) from a natural language description. " +
				"Returns React Flow compatible nodes and edges for IAC's Workflow editor. " +
				"Tasks can reference pages and trancodes; gateways handle conditional branching.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"description": {
						Type:        "string",
						Description: "Natural language description of the workflow (required). " +
							"Example: 'Leave request approval: employee submits → manager reviews → if approved notify HR, if rejected notify employee'.",
					},
					"current_workflow_json": {
						Type:        "string",
						Description: "Optional. Existing workflow JSON (nodes + edges) to extend or modify.",
					},
				},
				Required: []string{"description"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.openAIKey == "" {
			return "", fmt.Errorf("OpenAI API key is not configured; generate_workflow requires AI capabilities")
		}
		description := stringArg(args, "description")
		if description == "" {
			return "", fmt.Errorf("description is required")
		}

		var currentWorkflow map[string]interface{}
		if v := stringArg(args, "current_workflow_json"); v != "" {
			json.Unmarshal([]byte(v), &currentWorkflow) //nolint:errcheck
		}

		result, err := codegen.GenerateWorkflowFromDescription(description, s.openAIKey, s.openAIModel, currentWorkflow)
		if err != nil {
			return "", fmt.Errorf("workflow generation failed: %w", err)
		}

		b, _ := json.Marshal(result)
		return string(b), nil
	})
}

// ─── PAGE ─────────────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerPageCodegenTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "generate_page",
			Description: "Generates a complete IAC page structure (panels, views, trancodes, actions) from a natural language description using multi-step AI. " +
				"Returns a page JSON with all required views and backend trancodes identified, ready for the IAC Page designer.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"description": {
						Type:        "string",
						Description: "Natural language description of the page to generate (required). " +
							"Example: 'An inventory management page with a searchable product list, a details panel, and buttons to add/edit/delete products'.",
					},
					"current_page_json": {
						Type:        "string",
						Description: "Optional. Existing page JSON to enhance or extend.",
					},
				},
				Required: []string{"description"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.openAIKey == "" {
			return "", fmt.Errorf("OpenAI API key is not configured; generate_page requires AI capabilities")
		}
		description := stringArg(args, "description")
		if description == "" {
			return "", fmt.Errorf("description is required")
		}

		var currentPage map[string]interface{}
		if v := stringArg(args, "current_page_json"); v != "" {
			json.Unmarshal([]byte(v), &currentPage) //nolint:errcheck
		}

		result, err := codegen.GeneratePageMultiStep(ctx, description, s.openAIKey, s.openAIModel, currentPage)
		if err != nil {
			return "", fmt.Errorf("page generation failed: %w", err)
		}

		b, _ := json.Marshal(result)
		return string(b), nil
	})
}

// ─── VIEW ─────────────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerViewCodegenTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "generate_view",
			Description: "Generates an IAC view structure (fields, actions, data source bindings, HTML/JS code) from a natural language description. " +
				"Supports list views, form views, detail views, and custom layouts. " +
				"Returns view JSON ready for the IAC View designer.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"description": {
						Type:        "string",
						Description: "Natural language description of the view to generate (required). " +
							"Example: 'A form view for creating a new customer with name, email, phone, address fields and save/cancel buttons'.",
					},
					"current_view_json": {
						Type:        "string",
						Description: "Optional. Existing view JSON to enhance or modify.",
					},
				},
				Required: []string{"description"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.openAIKey == "" {
			return "", fmt.Errorf("OpenAI API key is not configured; generate_view requires AI capabilities")
		}
		description := stringArg(args, "description")
		if description == "" {
			return "", fmt.Errorf("description is required")
		}

		var currentView map[string]interface{}
		if v := stringArg(args, "current_view_json"); v != "" {
			json.Unmarshal([]byte(v), &currentView) //nolint:errcheck
		}

		result, err := codegen.GenerateViewFromDescription(description, s.openAIKey, s.openAIModel, currentView)
		if err != nil {
			return "", fmt.Errorf("view generation failed: %w", err)
		}

		b, _ := json.Marshal(result)
		return string(b), nil
	})
}

// ─── WHITEBOARD ───────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerWhiteboardCodegenTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "generate_whiteboard",
			Description: "Generates Excalidraw-compatible diagram elements from a natural language description. " +
				"Supports flowcharts, org charts, mind maps, UML class diagrams, sequence diagrams, and free-form diagrams. " +
				"Returns elements JSON for IAC's Whiteboard editor.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"description": {
						Type:        "string",
						Description: "Natural language description of the diagram to draw (required). " +
							"Examples: 'order processing flowchart', 'org chart for engineering team with 3 teams', " +
							"'UML class diagram for e-commerce with Product, Order, Customer classes'.",
					},
					"current_whiteboard_json": {
						Type:        "string",
						Description: "Optional. Existing whiteboard elements JSON to add to or modify.",
					},
				},
				Required: []string{"description"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.openAIKey == "" {
			return "", fmt.Errorf("OpenAI API key is not configured; generate_whiteboard requires AI capabilities")
		}
		description := stringArg(args, "description")
		if description == "" {
			return "", fmt.Errorf("description is required")
		}

		var currentWhiteboard map[string]interface{}
		if v := stringArg(args, "current_whiteboard_json"); v != "" {
			json.Unmarshal([]byte(v), &currentWhiteboard) //nolint:errcheck
		}

		result, err := codegen.GenerateWhiteboardFromDescription(description, s.openAIKey, s.openAIModel, currentWhiteboard)
		if err != nil {
			return "", fmt.Errorf("whiteboard generation failed: %w", err)
		}

		b, _ := json.Marshal(result)
		return string(b), nil
	})
}
