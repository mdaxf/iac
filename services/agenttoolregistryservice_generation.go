package services

// Entity generation tools — allow agents to create and update IAC content:
// Report (BI), 3D Model, Plant Model, Work Instruction, Dataset (View/Page).
// Each tool is a thin adapter: the agent generates the content/structure via its
// LLM reasoning loop, then calls the tool to persist it through the existing services.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	docschema "github.com/mdaxf/iac/documents/schema"
	iacmodels "github.com/mdaxf/iac/models"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func stringArg(args map[string]interface{}, key string) string {
	v, _ := args[key].(string)
	return v
}

func boolArg(args map[string]interface{}, key string) bool {
	v, _ := args[key].(bool)
	return v
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func docNotAvailable() (string, error) {
	return "", fmt.Errorf("document database (MongoDB) is not available — this tool requires a document DB connection")
}

func newID() string { return uuid.New().String() }

// ─── registration ─────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerGenerationTools() {
	s.registerReportTools()
	s.registerThreeDModelTools()
	s.registerPlantModelTools()
	s.registerWorkInstructionTools()
	s.registerDatasetTools()
	s.registerEntityTools()
}

// ═══════════════════════════════════════════════════════════════════════════════
// REPORT — BI / analytics
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerReportTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_report",
			Description: "Creates a BI report document with an optional SQL datasource and visual components. " +
				"Use query_database first to verify the SQL returns the expected data, then call this tool to persist the report. " +
				"Returns the new report ID on success.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":             {Type: "string", Description: "Report name (required)"},
					"description":      {Type: "string", Description: "Report description"},
					"category":         {Type: "string", Description: "Report category (e.g. 'Operations', 'Finance')"},
					"datasource_alias": {Type: "string", Description: "Alias for the main datasource (e.g. 'ds_main'). Leave empty to skip datasource."},
					"datasource_sql":   {Type: "string", Description: "SELECT SQL query for the main datasource"},
					"components_json": {
						Type: "string",
						Description: `JSON array of component definitions. Each component must have 'componenttype' (table|chart|text|image), 'name', and optionally 'charttype' (bar|line|pie|area), 'datasourcealias'. ` +
							`Example: [{"componenttype":"table","name":"Sales Table","datasourcealias":"ds_main","width":800,"height":400}]`,
					},
					"tags":      {Type: "string", Description: "Comma-separated tags"},
					"is_public": {Type: "string", Description: "Set to 'true' to make the report public"},
				},
				Required: []string{"name"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		name := stringArg(args, "name")
		if name == "" {
			return "", fmt.Errorf("name is required")
		}

		report := &docschema.ReportDocument{
			Name:        name,
			Description: stringArg(args, "description"),
			Category:    stringArg(args, "category"),
			ReportType:  "ai_generated",
			IsPublic:    stringArg(args, "is_public") == "true",
			Tags:        splitCSV(stringArg(args, "tags")),
			AIPrompt:    fmt.Sprintf("AI-generated report: %s", name),
		}

		// Datasource
		dsAlias := stringArg(args, "datasource_alias")
		dsSQL := stringArg(args, "datasource_sql")
		if dsSQL != "" {
			if dsAlias == "" {
				dsAlias = "ds_main"
			}
			report.Datasources = []docschema.ReportDatasourceDoc{{
				ID:        newID(),
				Alias:     dsAlias,
				QueryType: "custom",
				CustomSQL: dsSQL,
				Active:    true,
			}}
		}

		// Components
		if compJSON := stringArg(args, "components_json"); compJSON != "" {
			var comps []docschema.ReportComponentDoc
			if err := json.Unmarshal([]byte(compJSON), &comps); err != nil {
				// Fallback: accept raw maps and convert to component docs
				var raw []map[string]interface{}
				if json.Unmarshal([]byte(compJSON), &raw) == nil {
					for i, c := range raw {
						comp := docschema.ReportComponentDoc{
							ID:        newID(),
							IsVisible: true,
							Active:    true,
							ZIndex:    i + 1,
						}
						if v, ok := c["componenttype"].(string); ok {
							comp.ComponentType = v
						}
						if v, ok := c["name"].(string); ok {
							comp.Name = v
						}
						if v, ok := c["charttype"].(string); ok {
							comp.ChartType = v
						}
						if v, ok := c["datasourcealias"].(string); ok {
							comp.DatasourceAlias = v
						}
						if v, ok := c["width"].(float64); ok {
							comp.Width = v
						}
						if v, ok := c["height"].(float64); ok {
							comp.Height = v
						}
						report.Components = append(report.Components, comp)
					}
				}
			} else {
				// Ensure IDs are set on parsed components
				for i := range comps {
					if comps[i].ID == "" {
						comps[i].ID = newID()
					}
					comps[i].Active = true
					comps[i].IsVisible = true
				}
				report.Components = comps
			}
		}

		svc := NewReportDocService(s.docDB)
		id, err := svc.CreateReport(ctx, report, "agent")
		if err != nil {
			return "", fmt.Errorf("failed to create report: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": id, "name": name, "status": "created"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "update_report",
			Description: "Updates metadata or components of an existing BI report by ID. " +
				"Provide only the fields you want to change — unspecified fields are left unchanged.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":              {Type: "string", Description: "Report ID to update (required)"},
					"name":            {Type: "string", Description: "New report name"},
					"description":     {Type: "string", Description: "New description"},
					"category":        {Type: "string", Description: "New category"},
					"components_json": {Type: "string", Description: "JSON array of updated ReportComponentDoc objects (replaces all components)"},
					"tags":            {Type: "string", Description: "Comma-separated tags (replaces existing tags)"},
				},
				Required: []string{"id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}

		updates := map[string]interface{}{
			"modifiedby": "agent",
			"modifiedon": time.Now(),
		}
		if v := stringArg(args, "name"); v != "" {
			updates["name"] = v
		}
		if v := stringArg(args, "description"); v != "" {
			updates["description"] = v
		}
		if v := stringArg(args, "category"); v != "" {
			updates["category"] = v
		}
		if v := splitCSV(stringArg(args, "tags")); len(v) > 0 {
			updates["tags"] = v
		}
		if compJSON := stringArg(args, "components_json"); compJSON != "" {
			var comps []docschema.ReportComponentDoc
			if err := json.Unmarshal([]byte(compJSON), &comps); err == nil {
				updates["components"] = comps
			}
		}

		svc := NewReportDocService(s.docDB)
		if err := svc.UpdateReport(ctx, id, updates, "agent"); err != nil {
			return "", fmt.Errorf("failed to update report: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": id, "status": "updated"})
		return string(out), nil
	})
}

// ═══════════════════════════════════════════════════════════════════════════════
// 3D MODEL
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerThreeDModelTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_3d_model",
			Description: "Creates a new 3D model in IAC 3D Studio. " +
				"Provide the model content as a JSON array of Three.js-compatible objects in 'objects_json'. " +
				"Each object should have at minimum: type (Mesh|Group|Light), name, position {x,y,z}, and optionally geometry and material. " +
				"Returns the new model ID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":         {Type: "string", Description: "Model name (required)"},
					"description":  {Type: "string", Description: "Model description"},
					"category":     {Type: "string", Description: "Category (e.g. 'machinery', 'architecture', 'plant')"},
					"tags":         {Type: "string", Description: "Comma-separated tags"},
					"objects_json": {Type: "string", Description: `JSON array of Three.js scene objects. Example: [{"type":"Mesh","name":"Box1","geometry":{"type":"BoxGeometry","width":1,"height":1,"depth":1},"position":{"x":0,"y":0,"z":0},"material":{"type":"MeshStandardMaterial","color":"#4488ff"}}]`},
					"ai_prompt":    {Type: "string", Description: "The original prompt or intent used to generate this model"},
				},
				Required: []string{"name", "objects_json"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		name := stringArg(args, "name")
		objectsJSON := stringArg(args, "objects_json")
		if name == "" || objectsJSON == "" {
			return "", fmt.Errorf("name and objects_json are required")
		}

		var objects []map[string]interface{}
		if err := json.Unmarshal([]byte(objectsJSON), &objects); err != nil {
			return "", fmt.Errorf("invalid objects_json: %w", err)
		}

		req := &iacmodels.ThreeDModelCreateRequest{
			Name:          name,
			Description:   stringArg(args, "description"),
			Category:      stringArg(args, "category"),
			Tags:          splitCSV(stringArg(args, "tags")),
			Format:        iacmodels.ThreeDModelFormatJSON,
			IsAIGenerated: true,
			AIPrompt:      stringArg(args, "ai_prompt"),
			ModelData: map[string]interface{}{
				"objects": objects,
			},
		}

		svc := NewThreeDModelService(s.docDB)
		model, err := svc.CreateModel(ctx, req, "agent")
		if err != nil {
			return "", fmt.Errorf("failed to create 3D model: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": model.ID, "name": model.Name, "status": "created"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "update_3d_model",
			Description: "Updates an existing 3D model by creating a new version or updating its metadata. " +
				"Providing 'objects_json' creates a new version; omit it to only update metadata.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":                  {Type: "string", Description: "3D model ID (required)"},
					"objects_json":        {Type: "string", Description: "JSON array of updated Three.js objects — triggers new version creation"},
					"version_description": {Type: "string", Description: "Description of what changed in this version"},
					"name":                {Type: "string", Description: "New model name"},
					"description":         {Type: "string", Description: "New model description"},
					"category":            {Type: "string", Description: "New category"},
					"tags":                {Type: "string", Description: "Comma-separated tags"},
					"status":              {Type: "string", Description: "New status: DRAFT | IN_PROGRESS | REVIEW | PUBLISHED | ARCHIVED"},
				},
				Required: []string{"id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}

		req := &iacmodels.ThreeDModelUpdateRequest{
			Name:               stringArg(args, "name"),
			Description:        stringArg(args, "description"),
			Category:           stringArg(args, "category"),
			Tags:               splitCSV(stringArg(args, "tags")),
			VersionDescription: stringArg(args, "version_description"),
			Status:             iacmodels.ThreeDModelStatus(stringArg(args, "status")),
			SetAsDefault:       true,
		}

		if objectsJSON := stringArg(args, "objects_json"); objectsJSON != "" {
			var objects []map[string]interface{}
			if err := json.Unmarshal([]byte(objectsJSON), &objects); err != nil {
				return "", fmt.Errorf("invalid objects_json: %w", err)
			}
			req.ModelData = map[string]interface{}{"objects": objects}
		}

		svc := NewThreeDModelService(s.docDB)
		model, err := svc.UpdateModel(ctx, id, req, "agent")
		if err != nil {
			return "", fmt.Errorf("failed to update 3D model: %w", err)
		}
		out, _ := json.Marshal(map[string]interface{}{
			"id":      model.ID,
			"name":    model.Name,
			"version": model.CurrentVersion,
			"status":  "updated",
		})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_3d_models",
			Description: "Lists 3D models in IAC 3D Studio. Use this to discover existing models before creating or updating.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"limit": {Type: "integer", Description: "Maximum number of models to return (default 20)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		svc := NewThreeDModelService(s.docDB)
		modelList, err := svc.ListModels(ctx, "", nil)
		if err != nil {
			return "", fmt.Errorf("failed to list 3D models: %w", err)
		}
		type item struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
			Status   string `json:"status"`
		}
		var items []item
		for _, m := range modelList {
			items = append(items, item{ID: m.ID, Name: m.Name, Category: m.Category, Status: string(m.Status)})
		}
		if items == nil {
			items = []item{}
		}
		out, _ := json.Marshal(items)
		return string(out), nil
	})
}

// ═══════════════════════════════════════════════════════════════════════════════
// PLANT MODEL — plant layout / factory floor
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerPlantModelTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_plant_model",
			Description: "Creates a new plant layout in IAC Plant Studio. " +
				"Provide the plant objects (machines, robots, conveyors, etc.) as a JSON array in 'objects_json'. " +
				"Returns the new plant model ID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":         {Type: "string", Description: "Plant model name (required)"},
					"description":  {Type: "string", Description: "Plant description"},
					"category":     {Type: "string", Description: "Category (e.g. 'Assembly', 'Welding', 'Warehouse')"},
					"tags":         {Type: "string", Description: "Comma-separated tags"},
					"version":      {Type: "string", Description: "Version string (e.g. '1.0')"},
					"objects_json": {Type: "string", Description: `JSON array of plant objects. Each object should have: assetType (MACHINE_CNC|ROBOT_ARM|CONVEYOR|SHELF|SENSOR|WORKSTATION), name, position {x,y,z}, rotation {x,y,z}, scale {x,y,z}. Example: [{"assetType":"ROBOT_ARM","name":"Robot1","position":{"x":2,"y":0,"z":3},"rotation":{"x":0,"y":0,"z":0},"scale":{"x":1,"y":1,"z":1}}]`},
					"floor_settings_json": {Type: "string", Description: `Optional floor settings JSON: {"width":20,"depth":20,"color":"#cccccc"}`},
				},
				Required: []string{"name"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		name := stringArg(args, "name")
		if name == "" {
			return "", fmt.Errorf("name is required")
		}

		model := &iacmodels.PlantModel{
			ID:          newID(),
			Name:        name,
			Description: stringArg(args, "description"),
			Category:    stringArg(args, "category"),
			Tags:        splitCSV(stringArg(args, "tags")),
			Version:     stringArg(args, "version"),
			Active:      true,
		}
		if model.Version == "" {
			model.Version = "1.0"
		}

		if objectsJSON := stringArg(args, "objects_json"); objectsJSON != "" {
			var objects []map[string]interface{}
			if err := json.Unmarshal([]byte(objectsJSON), &objects); err != nil {
				return "", fmt.Errorf("invalid objects_json: %w", err)
			}
			model.Objects = objects
		}

		if floorJSON := stringArg(args, "floor_settings_json"); floorJSON != "" {
			var fs map[string]interface{}
			if json.Unmarshal([]byte(floorJSON), &fs) == nil {
				model.FloorSettings = fs
			}
		}

		svc := NewPlantStudioService(s.docDB)
		if err := svc.SavePlantModel(ctx, model, "agent"); err != nil {
			return "", fmt.Errorf("failed to create plant model: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": model.ID, "name": model.Name, "status": "created"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "update_plant_model",
			Description: "Updates an existing plant layout by ID. Fetches the current model and applies changes. " +
				"Provide only the fields to change — unspecified fields retain their current values.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":           {Type: "string", Description: "Plant model ID (required)"},
					"name":         {Type: "string", Description: "New name"},
					"description":  {Type: "string", Description: "New description"},
					"objects_json": {Type: "string", Description: "JSON array of updated plant objects (replaces all objects)"},
					"tags":         {Type: "string", Description: "Comma-separated tags (replaces existing)"},
				},
				Required: []string{"id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}

		svc := NewPlantStudioService(s.docDB)
		model, err := svc.GetPlantModel(ctx, id)
		if err != nil {
			return "", fmt.Errorf("plant model not found: %w", err)
		}

		if v := stringArg(args, "name"); v != "" {
			model.Name = v
		}
		if v := stringArg(args, "description"); v != "" {
			model.Description = v
		}
		if v := splitCSV(stringArg(args, "tags")); len(v) > 0 {
			model.Tags = v
		}
		if objectsJSON := stringArg(args, "objects_json"); objectsJSON != "" {
			var objects []map[string]interface{}
			if err := json.Unmarshal([]byte(objectsJSON), &objects); err != nil {
				return "", fmt.Errorf("invalid objects_json: %w", err)
			}
			model.Objects = objects
		}

		if err := svc.SavePlantModel(ctx, model, "agent"); err != nil {
			return "", fmt.Errorf("failed to update plant model: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": model.ID, "name": model.Name, "status": "updated"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_plant_models",
			Description: "Lists all plant models in IAC Plant Studio.",
			Parameters:  ToolParameterSchema{Type: "object", Properties: map[string]ToolPropertySchema{}},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		svc := NewPlantStudioService(s.docDB)
		list, err := svc.ListPlantModels(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to list plant models: %w", err)
		}
		type item struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
			Version  string `json:"version"`
		}
		var items []item
		for _, m := range list {
			items = append(items, item{ID: m.ID, Name: m.Name, Category: m.Category, Version: m.Version})
		}
		if items == nil {
			items = []item{}
		}
		out, _ := json.Marshal(items)
		return string(out), nil
	})
}

// ═══════════════════════════════════════════════════════════════════════════════
// WORK INSTRUCTION
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerWorkInstructionTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_work_instruction",
			Description: "Creates a work instruction document with step-by-step instructions. " +
				"Generate the steps content first, then call this tool to persist it. " +
				"Returns the new work instruction ID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"title":       {Type: "string", Description: "Work instruction title (required)"},
					"part_number": {Type: "string", Description: "Part or product number this instruction applies to (required)"},
					"version":     {Type: "string", Description: "Version string, e.g. '1.0' (required)"},
					"author":      {Type: "string", Description: "Author name"},
					"category":    {Type: "string", Description: "Category (e.g. 'Assembly', 'Maintenance', 'Quality')"},
					"description": {Type: "string", Description: "Overall description of the work instruction"},
					"tags":        {Type: "string", Description: "Comma-separated tags"},
					"steps_json": {
						Type: "string",
						Description: `JSON array of steps. Each step: {"order":1,"title":"Step title","description":"Detailed instructions in HTML or plain text"}. ` +
							`Example: [{"order":1,"title":"Prepare workspace","description":"Clear the work area and gather required tools."},{"order":2,"title":"Assemble part A","description":"Insert part A into slot B and tighten with M8 bolt."}]`,
					},
				},
				Required: []string{"title", "part_number", "version"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		title := stringArg(args, "title")
		partNumber := stringArg(args, "part_number")
		version := stringArg(args, "version")
		if title == "" || partNumber == "" || version == "" {
			return "", fmt.Errorf("title, part_number, and version are required")
		}

		req := &iacmodels.WorkInstructionCreateRequest{
			Title:       title,
			PartNumber:  partNumber,
			Version:     version,
			Author:      stringArg(args, "author"),
			Category:    stringArg(args, "category"),
			Description: stringArg(args, "description"),
			Tags:        splitCSV(stringArg(args, "tags")),
			Status:      iacmodels.StatusDraft,
		}
		if req.Author == "" {
			req.Author = "agent"
		}

		if stepsJSON := stringArg(args, "steps_json"); stepsJSON != "" {
			type simpleStep struct {
				Order       int    `json:"order"`
				Title       string `json:"title"`
				Description string `json:"description"`
			}
			var simple []simpleStep
			if err := json.Unmarshal([]byte(stepsJSON), &simple); err != nil {
				return "", fmt.Errorf("invalid steps_json: %w", err)
			}
			for _, ss := range simple {
				req.Steps = append(req.Steps, iacmodels.Step{
					ID:          newID(),
					Order:       ss.Order,
					Title:       ss.Title,
					Description: ss.Description,
				})
			}
		}

		svc := NewWorkInstructionService(s.docDB, "")
		wi, err := svc.CreateWorkInstruction(ctx, req, "agent")
		if err != nil {
			return "", fmt.Errorf("failed to create work instruction: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": wi.ID, "title": wi.Title, "status": "created"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "update_work_instruction",
			Description: "Updates an existing work instruction by ID. " +
				"Use status='RELEASED' to publish, 'OBSOLETE' to retire. " +
				"Providing steps_json replaces all existing steps.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":          {Type: "string", Description: "Work instruction ID (required)"},
					"title":       {Type: "string", Description: "New title"},
					"description": {Type: "string", Description: "New description"},
					"version":     {Type: "string", Description: "New version string"},
					"status":      {Type: "string", Description: "New status: DRAFT | REVIEW | RELEASED | OBSOLETE"},
					"tags":        {Type: "string", Description: "Comma-separated tags (replaces existing)"},
					"steps_json":  {Type: "string", Description: `JSON array of steps: [{"order":1,"title":"...","description":"..."}]`},
				},
				Required: []string{"id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}

		req := &iacmodels.WorkInstructionUpdateRequest{
			Title:       stringArg(args, "title"),
			Description: stringArg(args, "description"),
			Version:     stringArg(args, "version"),
			Status:      iacmodels.InstructionStatus(stringArg(args, "status")),
			Tags:        splitCSV(stringArg(args, "tags")),
		}

		if stepsJSON := stringArg(args, "steps_json"); stepsJSON != "" {
			type simpleStep struct {
				Order       int    `json:"order"`
				Title       string `json:"title"`
				Description string `json:"description"`
			}
			var simple []simpleStep
			if err := json.Unmarshal([]byte(stepsJSON), &simple); err != nil {
				return "", fmt.Errorf("invalid steps_json: %w", err)
			}
			for _, ss := range simple {
				req.Steps = append(req.Steps, iacmodels.Step{
					ID:          newID(),
					Order:       ss.Order,
					Title:       ss.Title,
					Description: ss.Description,
				})
			}
		}

		svc := NewWorkInstructionService(s.docDB, "")
		wi, err := svc.UpdateWorkInstruction(ctx, id, req, "agent")
		if err != nil {
			return "", fmt.Errorf("failed to update work instruction: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": wi.ID, "title": wi.Title, "status": "updated"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_work_instructions",
			Description: "Lists work instructions in IAC. Use this to discover existing instructions before creating or updating.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"status": {Type: "string", Description: "Filter by status: DRAFT | REVIEW | RELEASED | OBSOLETE (omit for all)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		svc := NewWorkInstructionService(s.docDB, "")
		status := stringArg(args, "status")
		var list []iacmodels.WorkInstruction
		var err error
		if status != "" {
			list, err = svc.GetByStatus(ctx, iacmodels.InstructionStatus(status))
		} else {
			list, err = svc.ListWorkInstructions(ctx)
		}
		if err != nil {
			return "", fmt.Errorf("failed to list work instructions: %w", err)
		}
		type item struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			PartNumber string `json:"part_number"`
			Version    string `json:"version"`
			Status     string `json:"status"`
		}
		var items []item
		for i := range list {
			items = append(items, item{
				ID:         list[i].ID,
				Title:      list[i].Title,
				PartNumber: list[i].PartNumber,
				Version:    list[i].Version,
				Status:     string(list[i].Status),
			})
		}
		if items == nil {
			items = []item{}
		}
		out, _ := json.Marshal(items)
		return string(out), nil
	})
}

// ═══════════════════════════════════════════════════════════════════════════════
// DATASET — view / page schema
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerDatasetTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_dataset",
			Description: "Creates a dataset schema in IAC. Datasets define the data source and field layout for views and pages. " +
				"Use query_database first to verify your SQL, then create the dataset. " +
				"Returns the new dataset ID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":             {Type: "string", Description: "Dataset name (required, must be unique)"},
					"datasource_type":  {Type: "string", Description: "Database type: postgresql | mysql | mssql | oracle (required)"},
					"datasource":       {Type: "string", Description: "Connection alias or connection name (required)"},
					"key_field":        {Type: "string", Description: "Primary key field name (required, e.g. 'id')"},
					"query":            {Type: "string", Description: "SQL SELECT query for the dataset"},
					"list_fields":      {Type: "string", Description: "Comma-separated field names to display in list views"},
					"hidden_fields":    {Type: "string", Description: "Comma-separated field names to hide"},
					"description":      {Type: "string", Description: "Dataset description"},
					"tags":             {Type: "string", Description: "Comma-separated tags"},
					"version":          {Type: "string", Description: "Version string (default '1.0')"},
					"detail_page_json": {Type: "string", Description: `Optional detail page config JSON: {"page":"detail","fields":["id","name","status"]}`},
				},
				Required: []string{"name", "datasource_type", "datasource", "key_field"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		name := stringArg(args, "name")
		dsType := stringArg(args, "datasource_type")
		ds := stringArg(args, "datasource")
		keyField := stringArg(args, "key_field")
		if name == "" || dsType == "" || ds == "" || keyField == "" {
			return "", fmt.Errorf("name, datasource_type, datasource, and key_field are required")
		}

		version := stringArg(args, "version")
		if version == "" {
			version = "1.0"
		}

		req := &iacmodels.DataSetCreateRequest{
			Schema:         "http://json-schema.org/draft-07/schema#",
			Ref:            fmt.Sprintf("#/definitions/%s", name),
			Name:           name,
			Version:        version,
			DataSourceType: dsType,
			DataSource:     ds,
			KeyField:       keyField,
			Query:          stringArg(args, "query"),
			ListFields:     splitCSV(stringArg(args, "list_fields")),
			HiddenFields:   splitCSV(stringArg(args, "hidden_fields")),
			Description:    stringArg(args, "description"),
			Tags:           splitCSV(stringArg(args, "tags")),
		}

		if detailJSON := stringArg(args, "detail_page_json"); detailJSON != "" {
			var dp map[string]interface{}
			if json.Unmarshal([]byte(detailJSON), &dp) == nil {
				req.DetailPage = dp
			}
		}

		svc := NewDataSetService(s.docDB)
		dataset, err := svc.CreateDataSet(ctx, req, "agent")
		if err != nil {
			return "", fmt.Errorf("failed to create dataset: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": dataset.ID, "name": dataset.Name, "status": "created"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "update_dataset",
			Description: "Updates an existing dataset schema by ID. Provide only the fields to change.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":           {Type: "string", Description: "Dataset ID (required)"},
					"query":        {Type: "string", Description: "New SQL query"},
					"list_fields":  {Type: "string", Description: "Comma-separated field names for list views (replaces existing)"},
					"hidden_fields": {Type: "string", Description: "Comma-separated hidden field names (replaces existing)"},
					"description":  {Type: "string", Description: "New description"},
					"tags":         {Type: "string", Description: "Comma-separated tags (replaces existing)"},
					"version":      {Type: "string", Description: "New version string"},
				},
				Required: []string{"id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}

		req := &iacmodels.DataSetUpdateRequest{
			Query:       stringArg(args, "query"),
			ListFields:  splitCSV(stringArg(args, "list_fields")),
			HiddenFields: splitCSV(stringArg(args, "hidden_fields")),
			Description: stringArg(args, "description"),
			Tags:        splitCSV(stringArg(args, "tags")),
			Version:     stringArg(args, "version"),
		}

		svc := NewDataSetService(s.docDB)
		dataset, err := svc.UpdateDataSet(ctx, id, req, "agent")
		if err != nil {
			return "", fmt.Errorf("failed to update dataset: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": dataset.ID, "name": dataset.Name, "status": "updated"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_datasets",
			Description: "Lists all dataset schemas in IAC. Use this to discover existing datasets before creating or updating.",
			Parameters:  ToolParameterSchema{Type: "object", Properties: map[string]ToolPropertySchema{}},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		svc := NewDataSetService(s.docDB)
		list, err := svc.ListDataSets(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to list datasets: %w", err)
		}
		type item struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Version string `json:"version"`
		}
		var items []item
		for _, d := range list {
			items = append(items, item{ID: d.ID, Name: d.Name, Version: d.Version})
		}
		if items == nil {
			items = []item{}
		}
		out, _ := json.Marshal(items)
		return string(out), nil
	})
}
