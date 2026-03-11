package services

// Entity tools for Workflow, BPM/TranCode, UI View, UI Page, and Whiteboard.
// These entities are stored directly in MongoDB (no dedicated service layer exists yet).
// The agent generates the structure via its LLM reasoning, then calls these tools to persist.

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdaxf/iac/documents"
)

// MongoDB collection names
const (
	collWorkflow   = "WorkFlow"
	collTranCode   = "Transaction_Code"
	collUIView     = "UI_View"
	collUIPage     = "UI_Page"
	collWhiteboard = "Whiteboard"
)

// ─── registration ─────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerEntityTools() {
	s.registerWorkflowTools()
	s.registerBPMTools()
	s.registerViewTools()
	s.registerPageTools()
	s.registerWhiteboardTools()
}

// ─── shared helpers ───────────────────────────────────────────────────────────

// entityInsert inserts a document and returns its uuid.
func (s *AgentToolRegistryService) entityInsert(ctx context.Context, collection string, doc map[string]interface{}) (string, error) {
	id := newID()
	doc["uuid"] = id
	now := time.Now().UTC().Format(time.RFC3339)
	if _, ok := doc["createdon"]; !ok {
		doc["createdon"] = now
	}
	if _, ok := doc["createdby"]; !ok {
		doc["createdby"] = "agent"
	}
	doc["modifiedon"] = now
	doc["modifiedby"] = "agent"
	doc["active"] = true

	if _, err := s.docDB.InsertOne(ctx, collection, doc); err != nil {
		return "", err
	}
	return id, nil
}

// entityFetch retrieves one document by uuid.
func (s *AgentToolRegistryService) entityFetch(ctx context.Context, collection, id string) (map[string]interface{}, error) {
	filter := map[string]interface{}{"uuid": id}
	doc, err := s.docDB.FindOne(ctx, collection, filter)
	if err != nil {
		return nil, fmt.Errorf("document not found (uuid=%s): %w", id, err)
	}
	return doc, nil
}

// entityUpdate applies a partial update to a document identified by uuid.
func (s *AgentToolRegistryService) entityUpdate(ctx context.Context, collection, id string, fields map[string]interface{}) error {
	fields["modifiedon"] = time.Now().UTC().Format(time.RFC3339)
	fields["modifiedby"] = "agent"
	filter := map[string]interface{}{"uuid": id}
	update := map[string]interface{}{"$set": fields}
	return s.docDB.UpdateOne(ctx, collection, filter, update)
}

// entityList returns up to limit documents from a collection.
func (s *AgentToolRegistryService) entityList(ctx context.Context, collection string, limit int) ([]map[string]interface{}, error) {
	opts := &documents.FindOptions{
		Sort:  map[string]int{"createdon": -1},
		Limit: int64(limit),
	}
	return s.docDB.FindMany(ctx, collection, map[string]interface{}{}, opts)
}

// ═══════════════════════════════════════════════════════════════════════════════
// WORKFLOW — React Flow node/edge diagram stored in WorkFlow collection
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerWorkflowTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_workflow",
			Description: "Creates a new workflow in IAC by saving a node/edge definition to the WorkFlow collection. " +
				"Provide the workflow nodes as a JSON array in 'nodes_json' and the connections as a JSON array in 'edges_json'. " +
				"Nodes follow React Flow format: {id, type (start|task|gateway|subflow|end), position {x,y}, data {name,description,...}}. " +
				"Edges follow React Flow format: {id, source, target, type, data {label}}. " +
				"Returns the new workflow UUID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":        {Type: "string", Description: "Workflow name (required)"},
					"description": {Type: "string", Description: "Workflow description"},
					"version":     {Type: "string", Description: "Version string, e.g. '1.0'"},
					"type":        {Type: "string", Description: "Workflow type: approval | task | process | custom"},
					"nodes_json": {
						Type: "string",
						Description: `JSON array of React Flow nodes. Required node fields: id, type (start|task|gateway|subflow|end), position {x,y}, data {name, description, label}. ` +
							`Task nodes may also have data fields: page, trancode, users (array), roles (array), processdata (object). ` +
							`Gateway nodes must have data.routingtables: [{sequence,data,value,target,isDefault}]. ` +
							`Example: [{"id":"n1","type":"start","position":{"x":250,"y":50},"data":{"name":"Start","label":"Start","width":24,"height":24}},` +
							`{"id":"n2","type":"task","position":{"x":200,"y":200},"data":{"name":"Review","label":"Review","description":"Manager reviews request"}}]`,
					},
					"edges_json": {
						Type: "string",
						Description: `JSON array of React Flow edges connecting nodes. Fields: id, source (node id), target (node id), type ("default"|"smoothstep"), data {label}. ` +
							`Example: [{"id":"e1-2","source":"n1","target":"n2","type":"default","data":{"label":""}}]`,
					},
				},
				Required: []string{"name", "nodes_json", "edges_json"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		name := stringArg(args, "name")
		nodesJSON := stringArg(args, "nodes_json")
		edgesJSON := stringArg(args, "edges_json")
		if name == "" || nodesJSON == "" || edgesJSON == "" {
			return "", fmt.Errorf("name, nodes_json, and edges_json are required")
		}

		var nodes, edges []map[string]interface{}
		if err := json.Unmarshal([]byte(nodesJSON), &nodes); err != nil {
			return "", fmt.Errorf("invalid nodes_json: %w", err)
		}
		if err := json.Unmarshal([]byte(edgesJSON), &edges); err != nil {
			return "", fmt.Errorf("invalid edges_json: %w", err)
		}

		version := stringArg(args, "version")
		if version == "" {
			version = "1.0"
		}
		doc := map[string]interface{}{
			"name":        name,
			"description": stringArg(args, "description"),
			"version":     version,
			"type":        stringArg(args, "type"),
			"isdefault":   true,
			"nodes":       nodes,
			"edges":       edges,
		}

		id, err := s.entityInsert(ctx, collWorkflow, doc)
		if err != nil {
			return "", fmt.Errorf("failed to create workflow: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": id, "name": name, "status": "created"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "update_workflow",
			Description: "Updates an existing workflow by UUID. Provide only the fields to change. Providing nodes_json and/or edges_json replaces those arrays entirely.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":          {Type: "string", Description: "Workflow UUID (required)"},
					"name":        {Type: "string", Description: "New name"},
					"description": {Type: "string", Description: "New description"},
					"version":     {Type: "string", Description: "New version"},
					"nodes_json":  {Type: "string", Description: "JSON array of updated React Flow nodes (replaces all nodes)"},
					"edges_json":  {Type: "string", Description: "JSON array of updated React Flow edges (replaces all edges)"},
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

		fields := map[string]interface{}{}
		if v := stringArg(args, "name"); v != "" {
			fields["name"] = v
		}
		if v := stringArg(args, "description"); v != "" {
			fields["description"] = v
		}
		if v := stringArg(args, "version"); v != "" {
			fields["version"] = v
		}
		if nodesJSON := stringArg(args, "nodes_json"); nodesJSON != "" {
			var nodes []map[string]interface{}
			if err := json.Unmarshal([]byte(nodesJSON), &nodes); err != nil {
				return "", fmt.Errorf("invalid nodes_json: %w", err)
			}
			fields["nodes"] = nodes
		}
		if edgesJSON := stringArg(args, "edges_json"); edgesJSON != "" {
			var edges []map[string]interface{}
			if err := json.Unmarshal([]byte(edgesJSON), &edges); err != nil {
				return "", fmt.Errorf("invalid edges_json: %w", err)
			}
			fields["edges"] = edges
		}

		if err := s.entityUpdate(ctx, collWorkflow, id, fields); err != nil {
			return "", fmt.Errorf("failed to update workflow: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": id, "status": "updated"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_workflows",
			Description: "Lists workflows in IAC. Use to discover existing workflows before creating or updating.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"limit": {Type: "integer", Description: "Maximum number of workflows to return (default 20)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		limit := 20
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}
		docs, err := s.entityList(ctx, collWorkflow, limit)
		if err != nil {
			return "", fmt.Errorf("failed to list workflows: %w", err)
		}
		type item struct {
			ID          interface{} `json:"id"`
			Name        interface{} `json:"name"`
			Version     interface{} `json:"version"`
			Description interface{} `json:"description"`
		}
		var items []item
		for _, d := range docs {
			items = append(items, item{
				ID:          d["uuid"],
				Name:        d["name"],
				Version:     d["version"],
				Description: d["description"],
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
// BPM / TRANCODE — stored in Transaction_Code collection
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerBPMTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_bpm",
			Description: "Creates a new BPM (TranCode / Transaction Code) in IAC. " +
				"A BPM is a function-group based execution unit: it has inputs, outputs, and one or more function groups, each containing functions. " +
				"Provide the function groups as a JSON array in 'functiongroups_json'. " +
				"Use query_database first to understand the data model before generating SQL inside query functions. " +
				"Returns the new TranCode UUID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":        {Type: "string", Description: "TranCode name, used as trancodename (required, unique)"},
					"description": {Type: "string", Description: "BPM description"},
					"version":     {Type: "string", Description: "Version string, e.g. '1.0'"},
					"inputs_json": {
						Type: "string",
						Description: `JSON array of input parameters. Each: {"name":"inputName","datatype":"string|integer|float|bool|datetime|object","description":"...","inivalue":"","defaultvalue":""}. ` +
							`Example: [{"name":"customerID","datatype":"string","description":"Customer identifier"}]`,
					},
					"outputs_json": {
						Type: "string",
						Description: `JSON array of output parameters. Each: {"name":"outputName","datatype":"string|integer|float|bool|datetime|object","description":"..."}. ` +
							`Example: [{"name":"result","datatype":"object","description":"Processing result"}]`,
					},
					"functiongroups_json": {
						Type: "string",
						Description: `JSON array of function groups. Each group: {"id":"fg1","name":"GroupName","description":"...","executionsequence":"sequential","functions":[...]}. ` +
							`Each function in a group: {"id":"f1","name":"FuncName","functype":"query|goexpr|javascript|inputmap|tableinsert|tableupdate|tabledelete","description":"...","content":"SQL or script","inputs":[...],"outputs":[...]}. ` +
							`Common functypes: "query" (SELECT SQL), "tableinsert" (INSERT), "tableupdate" (UPDATE), "goexpr" (Go expression), "javascript" (JS code), "inputmap" (map inputs to outputs). ` +
							`Example: [{"id":"fg1","name":"FetchData","executionsequence":"sequential","functions":[{"id":"f1","name":"GetCustomer","functype":"query","content":"SELECT * FROM customers WHERE id = {{customerID}} LIMIT 1","inputs":[{"name":"customerID","source":"input","datatype":"string"}],"outputs":[{"name":"customer","datatype":"object"}]}]}]`,
					},
					"first_funcgroup": {Type: "string", Description: "ID of the first function group to execute (defaults to first in the array)"},
				},
				Required: []string{"name", "functiongroups_json"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		name := stringArg(args, "name")
		fgJSON := stringArg(args, "functiongroups_json")
		if name == "" || fgJSON == "" {
			return "", fmt.Errorf("name and functiongroups_json are required")
		}

		var functionGroups []map[string]interface{}
		if err := json.Unmarshal([]byte(fgJSON), &functionGroups); err != nil {
			return "", fmt.Errorf("invalid functiongroups_json: %w", err)
		}

		// Determine firstfuncgroup
		firstFG := stringArg(args, "first_funcgroup")
		if firstFG == "" && len(functionGroups) > 0 {
			if id, ok := functionGroups[0]["id"].(string); ok {
				firstFG = id
			}
		}

		// Parse optional inputs/outputs
		var inputs, outputs []map[string]interface{}
		if inJSON := stringArg(args, "inputs_json"); inJSON != "" {
			_ = json.Unmarshal([]byte(inJSON), &inputs)
		}
		if outJSON := stringArg(args, "outputs_json"); outJSON != "" {
			_ = json.Unmarshal([]byte(outJSON), &outputs)
		}

		version := stringArg(args, "version")
		if version == "" {
			version = "1.0"
		}

		doc := map[string]interface{}{
			"trancodename":   name,
			"name":           name,
			"description":    stringArg(args, "description"),
			"version":        version,
			"isdefault":      true,
			"status":         0,
			"functiongroups": functionGroups,
			"firstfuncgroup": firstFG,
		}
		if inputs != nil {
			doc["inputs"] = inputs
		}
		if outputs != nil {
			doc["outputs"] = outputs
		}

		id, err := s.entityInsert(ctx, collTranCode, doc)
		if err != nil {
			return "", fmt.Errorf("failed to create BPM: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": id, "name": name, "status": "created"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "update_bpm",
			Description: "Updates an existing BPM/TranCode by UUID. Provide only the fields to change.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":                  {Type: "string", Description: "BPM UUID (required)"},
					"description":         {Type: "string", Description: "New description"},
					"version":             {Type: "string", Description: "New version"},
					"inputs_json":         {Type: "string", Description: "JSON array of updated inputs (replaces existing)"},
					"outputs_json":        {Type: "string", Description: "JSON array of updated outputs (replaces existing)"},
					"functiongroups_json": {Type: "string", Description: "JSON array of updated function groups (replaces all)"},
					"first_funcgroup":     {Type: "string", Description: "ID of the first function group to execute"},
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

		fields := map[string]interface{}{}
		if v := stringArg(args, "description"); v != "" {
			fields["description"] = v
		}
		if v := stringArg(args, "version"); v != "" {
			fields["version"] = v
		}
		if v := stringArg(args, "first_funcgroup"); v != "" {
			fields["firstfuncgroup"] = v
		}
		if fgJSON := stringArg(args, "functiongroups_json"); fgJSON != "" {
			var fgs []map[string]interface{}
			if err := json.Unmarshal([]byte(fgJSON), &fgs); err != nil {
				return "", fmt.Errorf("invalid functiongroups_json: %w", err)
			}
			fields["functiongroups"] = fgs
		}
		if inJSON := stringArg(args, "inputs_json"); inJSON != "" {
			var ins []map[string]interface{}
			if err := json.Unmarshal([]byte(inJSON), &ins); err == nil {
				fields["inputs"] = ins
			}
		}
		if outJSON := stringArg(args, "outputs_json"); outJSON != "" {
			var outs []map[string]interface{}
			if err := json.Unmarshal([]byte(outJSON), &outs); err == nil {
				fields["outputs"] = outs
			}
		}

		if err := s.entityUpdate(ctx, collTranCode, id, fields); err != nil {
			return "", fmt.Errorf("failed to update BPM: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": id, "status": "updated"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_bpms",
			Description: "Lists BPM/TranCode definitions in IAC. Use to discover existing transaction codes before creating or updating.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"limit": {Type: "integer", Description: "Maximum number of results to return (default 20)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		limit := 20
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}
		docs, err := s.entityList(ctx, collTranCode, limit)
		if err != nil {
			return "", fmt.Errorf("failed to list BPMs: %w", err)
		}
		type item struct {
			ID          interface{} `json:"id"`
			Name        interface{} `json:"name"`
			Version     interface{} `json:"version"`
			Description interface{} `json:"description"`
		}
		var items []item
		for _, d := range docs {
			items = append(items, item{
				ID:          d["uuid"],
				Name:        d["trancodename"],
				Version:     d["version"],
				Description: d["description"],
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
// UI VIEW — stored in UI_View collection
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerViewTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_view",
			Description: "Creates a new UI View in IAC. A view defines the data fields, layout, and actions for a UI component. " +
				"Views are used by Pages as panels. Provide the fields as a JSON array. " +
				"Use list_datasets or query_database to discover existing data sources before specifying datasource. " +
				"Returns the new view UUID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":        {Type: "string", Description: "View name (required, used as identifier)"},
					"title":       {Type: "string", Description: "Display title"},
					"type":        {Type: "string", Description: "View type: list | form | detail | custom (required)"},
					"description": {Type: "string", Description: "View description"},
					"datasource":  {Type: "string", Description: "Data source: table name, dataset name, or TranCode name"},
					"datasource_type": {Type: "string", Description: "Data source type: table | dataset | trancode | api"},
					"is_form":     {Type: "string", Description: "'true' if this is a form view that submits data"},
					"fields_json": {
						Type: "string",
						Description: `JSON array of field definitions. Each field: {"name":"fieldName","label":"Display Label","type":"text|number|date|select|checkbox|textarea","datatype":"string|integer|float|bool|datetime","required":false,"readonly":false,"editable":true,"visible":true,"defaultvalue":""}. ` +
							`Example: [{"name":"id","label":"ID","type":"text","datatype":"string","readonly":true,"visible":true},` +
							`{"name":"name","label":"Name","type":"text","datatype":"string","required":true,"editable":true}]`,
					},
					"actions_json": {
						Type: "string",
						Description: `JSON array of action definitions. Each: {"name":"ActionName","type":"button|link","label":"Label","target":"ViewOrTranCode","targetType":"view|trancode|url","event":"click|submit","position":"toolbar|inline"}. ` +
							`Example: [{"name":"btnSave","type":"button","label":"Save","target":"SaveTranCode","targetType":"trancode","event":"click","position":"toolbar"}]`,
					},
				},
				Required: []string{"name", "type", "fields_json"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		name := stringArg(args, "name")
		viewType := stringArg(args, "type")
		fieldsJSON := stringArg(args, "fields_json")
		if name == "" || viewType == "" || fieldsJSON == "" {
			return "", fmt.Errorf("name, type, and fields_json are required")
		}

		var fields []map[string]interface{}
		if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
			return "", fmt.Errorf("invalid fields_json: %w", err)
		}

		doc := map[string]interface{}{
			"name":           name,
			"title":          stringArg(args, "title"),
			"type":           viewType,
			"description":    stringArg(args, "description"),
			"datasource":     stringArg(args, "datasource"),
			"datasourcetype": stringArg(args, "datasource_type"),
			"isform":         boolArg(args, "is_form"),
			"fields":         fields,
		}
		if actJSON := stringArg(args, "actions_json"); actJSON != "" {
			var actions []map[string]interface{}
			if json.Unmarshal([]byte(actJSON), &actions) == nil {
				doc["actions"] = actions
			}
		}

		id, err := s.entityInsert(ctx, collUIView, doc)
		if err != nil {
			return "", fmt.Errorf("failed to create view: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": id, "name": name, "status": "created"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "update_view",
			Description: "Updates an existing UI View by UUID. Provide only the fields to change.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":           {Type: "string", Description: "View UUID (required)"},
					"title":        {Type: "string", Description: "New display title"},
					"description":  {Type: "string", Description: "New description"},
					"datasource":   {Type: "string", Description: "New data source"},
					"fields_json":  {Type: "string", Description: "JSON array of updated fields (replaces all fields)"},
					"actions_json": {Type: "string", Description: "JSON array of updated actions (replaces all actions)"},
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

		fields := map[string]interface{}{}
		if v := stringArg(args, "title"); v != "" {
			fields["title"] = v
		}
		if v := stringArg(args, "description"); v != "" {
			fields["description"] = v
		}
		if v := stringArg(args, "datasource"); v != "" {
			fields["datasource"] = v
		}
		if fJSON := stringArg(args, "fields_json"); fJSON != "" {
			var flds []map[string]interface{}
			if err := json.Unmarshal([]byte(fJSON), &flds); err != nil {
				return "", fmt.Errorf("invalid fields_json: %w", err)
			}
			fields["fields"] = flds
		}
		if aJSON := stringArg(args, "actions_json"); aJSON != "" {
			var acts []map[string]interface{}
			if json.Unmarshal([]byte(aJSON), &acts) == nil {
				fields["actions"] = acts
			}
		}

		if err := s.entityUpdate(ctx, collUIView, id, fields); err != nil {
			return "", fmt.Errorf("failed to update view: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": id, "status": "updated"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_views",
			Description: "Lists UI Views in IAC. Use to discover existing views before creating or updating.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"limit": {Type: "integer", Description: "Maximum number of views to return (default 20)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		limit := 20
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}
		docs, err := s.entityList(ctx, collUIView, limit)
		if err != nil {
			return "", fmt.Errorf("failed to list views: %w", err)
		}
		type item struct {
			ID          interface{} `json:"id"`
			Name        interface{} `json:"name"`
			Type        interface{} `json:"type"`
			Description interface{} `json:"description"`
		}
		var items []item
		for _, d := range docs {
			items = append(items, item{
				ID:          d["uuid"],
				Name:        d["name"],
				Type:        d["type"],
				Description: d["description"],
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
// UI PAGE — stored in UI_Page collection
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerPageTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_page",
			Description: "Creates a new UI Page in IAC. A page is a layout container made of panels, where each panel references a view. " +
				"Use list_views first to discover existing views to reference in panels. " +
				"Returns the new page UUID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":        {Type: "string", Description: "Page name (required, used as identifier)"},
					"title":       {Type: "string", Description: "Display title"},
					"description": {Type: "string", Description: "Page description"},
					"version":     {Type: "string", Description: "Version string (default '1.0')"},
					"orientation": {Type: "string", Description: "Layout orientation: 1 (vertical) | 2 (horizontal) | 3 (grid)"},
					"panels_json": {
						Type: "string",
						Description: `JSON array of panel definitions. Each panel: {"id":"panel1","type":"grid|form|detail|custom","title":"Panel Title","view":"ViewName","position":{"x":0,"y":0,"width":12,"height":6},"properties":{"showHeader":true,"collapsible":false}}. ` +
							`The "view" field references the name of an existing view. ` +
							`Example: [{"id":"p1","type":"list","title":"Customer List","view":"CustomerListView","position":{"x":0,"y":0,"width":12,"height":8}}]`,
					},
					"actions_json": {
						Type: "string",
						Description: `JSON array of page-level actions: [{"name":"btnNew","type":"button","label":"New","target":"CustomerFormView","targetType":"view","position":"toolbar"}]`,
					},
				},
				Required: []string{"name", "panels_json"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		name := stringArg(args, "name")
		panelsJSON := stringArg(args, "panels_json")
		if name == "" || panelsJSON == "" {
			return "", fmt.Errorf("name and panels_json are required")
		}

		var panels []map[string]interface{}
		if err := json.Unmarshal([]byte(panelsJSON), &panels); err != nil {
			return "", fmt.Errorf("invalid panels_json: %w", err)
		}

		version := stringArg(args, "version")
		if version == "" {
			version = "1.0"
		}
		orientation := 1
		if v := stringArg(args, "orientation"); v != "" {
			fmt.Sscanf(v, "%d", &orientation)
		}

		doc := map[string]interface{}{
			"name":        name,
			"title":       stringArg(args, "title"),
			"description": stringArg(args, "description"),
			"version":     version,
			"isdefault":   false,
			"status":      0,
			"orientation": orientation,
			"panels":      panels,
		}
		if actJSON := stringArg(args, "actions_json"); actJSON != "" {
			var actions []map[string]interface{}
			if json.Unmarshal([]byte(actJSON), &actions) == nil {
				doc["actions"] = actions
			}
		}

		id, err := s.entityInsert(ctx, collUIPage, doc)
		if err != nil {
			return "", fmt.Errorf("failed to create page: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": id, "name": name, "status": "created"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "update_page",
			Description: "Updates an existing UI Page by UUID. Provide only the fields to change.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":           {Type: "string", Description: "Page UUID (required)"},
					"title":        {Type: "string", Description: "New display title"},
					"description":  {Type: "string", Description: "New description"},
					"panels_json":  {Type: "string", Description: "JSON array of updated panels (replaces all panels)"},
					"actions_json": {Type: "string", Description: "JSON array of updated page actions"},
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

		fields := map[string]interface{}{}
		if v := stringArg(args, "title"); v != "" {
			fields["title"] = v
		}
		if v := stringArg(args, "description"); v != "" {
			fields["description"] = v
		}
		if pJSON := stringArg(args, "panels_json"); pJSON != "" {
			var panels []map[string]interface{}
			if err := json.Unmarshal([]byte(pJSON), &panels); err != nil {
				return "", fmt.Errorf("invalid panels_json: %w", err)
			}
			fields["panels"] = panels
		}
		if aJSON := stringArg(args, "actions_json"); aJSON != "" {
			var acts []map[string]interface{}
			if json.Unmarshal([]byte(aJSON), &acts) == nil {
				fields["actions"] = acts
			}
		}

		if err := s.entityUpdate(ctx, collUIPage, id, fields); err != nil {
			return "", fmt.Errorf("failed to update page: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": id, "status": "updated"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_pages",
			Description: "Lists UI Pages in IAC. Use to discover existing pages before creating or updating.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"limit": {Type: "integer", Description: "Maximum number of pages to return (default 20)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		limit := 20
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}
		docs, err := s.entityList(ctx, collUIPage, limit)
		if err != nil {
			return "", fmt.Errorf("failed to list pages: %w", err)
		}
		type item struct {
			ID          interface{} `json:"id"`
			Name        interface{} `json:"name"`
			Title       interface{} `json:"title"`
			Description interface{} `json:"description"`
		}
		var items []item
		for _, d := range docs {
			items = append(items, item{
				ID:          d["uuid"],
				Name:        d["name"],
				Title:       d["title"],
				Description: d["description"],
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
// WHITEBOARD — Excalidraw-compatible elements stored in Whiteboard collection
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerWhiteboardTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_whiteboard",
			Description: "Creates a new whiteboard in IAC using Excalidraw-compatible elements. " +
				"Generate the diagram elements as a JSON array in 'elements_json'. " +
				"Element types: rectangle, ellipse, diamond, arrow, line, text, freedraw. " +
				"Returns the new whiteboard UUID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":        {Type: "string", Description: "Whiteboard name (required)"},
					"description": {Type: "string", Description: "Whiteboard description"},
					"elements_json": {
						Type: "string",
						Description: `JSON array of Excalidraw-compatible elements. Each element requires: id (unique string), type (rectangle|ellipse|diamond|arrow|line|text|freedraw), x, y, width, height, angle (0), strokeColor ("#000000"), backgroundColor ("transparent" or hex), fillStyle ("hachure"|"solid"|"cross-hatch"), strokeWidth (1|2|4), opacity (100), seed (random integer), version (1), isDeleted (false), groupIds ([]), boundElements (null). ` +
							`Text elements additionally need: text (string), fontSize (16|20|28), fontFamily (1), textAlign ("left"|"center"), verticalAlign ("top"|"middle"). ` +
							`Arrow/Line elements additionally need: points ([[x1,y1],[x2,y2],...]), startArrowhead (null|"arrow"|"bar"), endArrowhead ("arrow"|null). ` +
							`Example: [{"id":"r1","type":"rectangle","x":100,"y":100,"width":200,"height":100,"angle":0,"strokeColor":"#000000","backgroundColor":"#4299e1","fillStyle":"solid","strokeWidth":2,"opacity":100,"seed":12345,"version":1,"isDeleted":false,"groupIds":[],"boundElements":null},` +
							`{"id":"t1","type":"text","x":120,"y":140,"width":160,"height":24,"text":"Hello IAC","fontSize":20,"fontFamily":1,"textAlign":"center","verticalAlign":"middle","angle":0,"strokeColor":"#ffffff","backgroundColor":"transparent","fillStyle":"hachure","strokeWidth":1,"opacity":100,"seed":99999,"version":1,"isDeleted":false,"groupIds":[],"boundElements":null}]`,
					},
					"app_state_json": {
						Type: "string",
						Description: `Optional Excalidraw appState JSON: {"viewBackgroundColor":"#ffffff","gridSize":null}`,
					},
				},
				Required: []string{"name", "elements_json"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		name := stringArg(args, "name")
		elemJSON := stringArg(args, "elements_json")
		if name == "" || elemJSON == "" {
			return "", fmt.Errorf("name and elements_json are required")
		}

		var elements []map[string]interface{}
		if err := json.Unmarshal([]byte(elemJSON), &elements); err != nil {
			return "", fmt.Errorf("invalid elements_json: %w", err)
		}

		doc := map[string]interface{}{
			"name":        name,
			"description": stringArg(args, "description"),
			"elements":    elements,
		}
		if appStateJSON := stringArg(args, "app_state_json"); appStateJSON != "" {
			var appState map[string]interface{}
			if json.Unmarshal([]byte(appStateJSON), &appState) == nil {
				doc["appState"] = appState
			}
		} else {
			doc["appState"] = map[string]interface{}{
				"viewBackgroundColor": "#ffffff",
				"gridSize":            nil,
			}
		}

		id, err := s.entityInsert(ctx, collWhiteboard, doc)
		if err != nil {
			return "", fmt.Errorf("failed to create whiteboard: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": id, "name": name, "status": "created"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "update_whiteboard",
			Description: "Updates an existing whiteboard by UUID. Provide elements_json to replace all elements, or app_state_json to update app state.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":             {Type: "string", Description: "Whiteboard UUID (required)"},
					"name":           {Type: "string", Description: "New name"},
					"description":    {Type: "string", Description: "New description"},
					"elements_json":  {Type: "string", Description: "JSON array of updated Excalidraw elements (replaces all elements)"},
					"app_state_json": {Type: "string", Description: "Updated Excalidraw appState JSON"},
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

		fields := map[string]interface{}{}
		if v := stringArg(args, "name"); v != "" {
			fields["name"] = v
		}
		if v := stringArg(args, "description"); v != "" {
			fields["description"] = v
		}
		if eJSON := stringArg(args, "elements_json"); eJSON != "" {
			var elements []map[string]interface{}
			if err := json.Unmarshal([]byte(eJSON), &elements); err != nil {
				return "", fmt.Errorf("invalid elements_json: %w", err)
			}
			fields["elements"] = elements
		}
		if asJSON := stringArg(args, "app_state_json"); asJSON != "" {
			var appState map[string]interface{}
			if json.Unmarshal([]byte(asJSON), &appState) == nil {
				fields["appState"] = appState
			}
		}

		if err := s.entityUpdate(ctx, collWhiteboard, id, fields); err != nil {
			return "", fmt.Errorf("failed to update whiteboard: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": id, "status": "updated"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_whiteboards",
			Description: "Lists whiteboards in IAC. Use to discover existing whiteboards before creating or updating.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"limit": {Type: "integer", Description: "Maximum number of whiteboards to return (default 20)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		limit := 20
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}
		docs, err := s.entityList(ctx, collWhiteboard, limit)
		if err != nil {
			return "", fmt.Errorf("failed to list whiteboards: %w", err)
		}
		type item struct {
			ID          interface{} `json:"id"`
			Name        interface{} `json:"name"`
			Description interface{} `json:"description"`
		}
		var items []item
		for _, d := range docs {
			items = append(items, item{
				ID:          d["uuid"],
				Name:        d["name"],
				Description: d["description"],
			})
		}
		if items == nil {
			items = []item{}
		}
		out, _ := json.Marshal(items)
		return string(out), nil
	})
}
