package services

// Tool & Skill management — allow agents to introspect, search, and manage IAC agent skills at runtime.
//
// INTROSPECTION (read registered Go tool handlers):
// - list_available_tools : list all registered tool handlers with name + short description
// - search_tools         : search tools by keyword (name + description)
// - describe_tool        : get the full parameter schema for a named tool
//
// SKILL CATALOG (agent_skills DB table — controls which tools appear in the UI picker):
// - list_agent_skills    : list skill catalog rows with optional category filter
// - create_agent_skill   : add a new skill entry (must match a registered tool handler name)
// - update_agent_skill   : update display_name, description, category, or active flag
//
// SKILL ASSIGNMENT (agent_definition_skills join table):
// - assign_skill_to_agent  : link a skill to an agent definition
// - remove_skill_from_agent: remove a skill from an agent definition
// - list_agent_assigned_skills: list skills assigned to a specific agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *AgentToolRegistryService) registerToolManagementTools() {
	s.registerToolIntrospectionTools()
	s.registerSkillCatalogTools()
	s.registerSkillAssignmentTools()
}

// ═══════════════════════════════════════════════════════════════════════════════
// TOOL INTROSPECTION
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerToolIntrospectionTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "list_available_tools",
			Description: "Lists all registered tool handlers in the IAC agent runtime with their names and descriptions. " +
				"Use this to discover what capabilities are available before choosing which tools to assign to an agent. " +
				"Filter by keyword to narrow results. The returned names are the exact values used in agent_skills.name.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"keyword": {Type: "string", Description: "Optional case-insensitive keyword to filter by name or description"},
					"limit":   {Type: "integer", Description: "Max tools to return (default: 100)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		keyword := strings.ToLower(stringArg(args, "keyword"))
		limit := 100
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}

		type toolSummary struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		var tools []toolSummary
		for name, def := range s.defs {
			desc := def.Function.Description
			if keyword != "" {
				if !strings.Contains(strings.ToLower(name), keyword) &&
					!strings.Contains(strings.ToLower(desc), keyword) {
					continue
				}
			}
			// Trim description to 120 chars
			if len(desc) > 120 {
				desc = desc[:117] + "..."
			}
			tools = append(tools, toolSummary{Name: name, Description: desc})
		}
		sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
		if len(tools) > limit {
			tools = tools[:limit]
		}

		out, _ := json.Marshal(map[string]interface{}{
			"count": len(tools),
			"tools": tools,
		})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "search_tools",
			Description: "Searches registered tool handlers by keyword and returns ranked matches. " +
				"Ranks by: exact name match > name contains keyword > description contains keyword. " +
				"Use this to find the best tool for a specific task before assigning it to an agent.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"keyword": {Type: "string", Description: "Search keyword (required) — searches tool names and descriptions"},
					"limit":   {Type: "integer", Description: "Max results to return (default: 10)"},
				},
				Required: []string{"keyword"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		keyword := strings.ToLower(stringArg(args, "keyword"))
		if keyword == "" {
			return "", fmt.Errorf("keyword is required")
		}
		limit := 10
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}

		type ranked struct {
			rank int
			name string
			desc string
		}
		var results []ranked
		for name, def := range s.defs {
			nameLow := strings.ToLower(name)
			descLow := strings.ToLower(def.Function.Description)
			rank := 0
			if nameLow == keyword {
				rank = 3
			} else if strings.Contains(nameLow, keyword) {
				rank = 2
			} else if strings.Contains(descLow, keyword) {
				rank = 1
			}
			if rank == 0 {
				continue
			}
			desc := def.Function.Description
			if len(desc) > 200 {
				desc = desc[:197] + "..."
			}
			results = append(results, ranked{rank, name, desc})
		}
		sort.Slice(results, func(i, j int) bool {
			if results[i].rank != results[j].rank {
				return results[i].rank > results[j].rank
			}
			return results[i].name < results[j].name
		})
		if len(results) > limit {
			results = results[:limit]
		}

		type result struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Score       int    `json:"relevance_score"`
		}
		out_results := make([]result, len(results))
		for i, r := range results {
			out_results[i] = result{r.name, r.desc, r.rank}
		}
		out, _ := json.Marshal(map[string]interface{}{
			"keyword": keyword,
			"count":   len(out_results),
			"results": out_results,
		})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "describe_tool",
			Description: "Returns the full parameter schema and description for a specific tool handler. " +
				"Use this to understand exactly what parameters a tool requires before calling it or assigning it to an agent.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name": {Type: "string", Description: "Exact tool handler name (required). Use list_available_tools to find names."},
				},
				Required: []string{"name"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		name := stringArg(args, "name")
		if name == "" {
			return "", fmt.Errorf("name is required")
		}
		def, ok := s.defs[name]
		if !ok {
			return "", fmt.Errorf("tool %q is not registered — use list_available_tools to find valid names", name)
		}
		out, _ := json.Marshal(def)
		return string(out), nil
	})
}

// ═══════════════════════════════════════════════════════════════════════════════
// SKILL CATALOG (agent_skills table)
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerSkillCatalogTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "list_agent_skills",
			Description: "Lists the agent skill catalog from the database. " +
				"Skills define which tool handlers are visible and selectable in the agent configuration UI. " +
				"A skill name must exactly match a registered tool handler name to be functional.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"category":    {Type: "string", Description: "Filter by category (e.g. 'system', 'reporting', 'scheduling')"},
					"active_only": {Type: "boolean", Description: "Return only active skills (default: true)"},
					"keyword":     {Type: "string", Description: "Optional substring filter on name or display_name"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		activeOnly := true
		if v, ok := args["active_only"].(bool); ok {
			activeOnly = v
		}
		category := stringArg(args, "category")
		keyword := strings.ToLower(stringArg(args, "keyword"))

		query := `SELECT id, name, display_name, description, category, active, createdon
		          FROM agent_skills WHERE 1=1`
		var qArgs []interface{}
		idx := 1
		if activeOnly {
			query += fmt.Sprintf(" AND active = $%d", idx)
			qArgs = append(qArgs, true)
			idx++
		}
		if category != "" {
			query += fmt.Sprintf(" AND category = $%d", idx)
			qArgs = append(qArgs, category)
			idx++
		}
		query += " ORDER BY category, name"

		rows, err := s.sqlDB.QueryContext(ctx, query, qArgs...)
		if err != nil {
			return "", fmt.Errorf("query error: %w", err)
		}
		defer rows.Close()

		type skillRow struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			DisplayName string `json:"display_name"`
			Description string `json:"description"`
			Category    string `json:"category"`
			Active      bool   `json:"active"`
			Registered  bool   `json:"registered_in_runtime"`
		}
		var skills []skillRow
		for rows.Next() {
			var r skillRow
			var createdon interface{}
			if err := rows.Scan(&r.ID, &r.Name, &r.DisplayName, &r.Description, &r.Category, &r.Active, &createdon); err != nil {
				continue
			}
			if keyword != "" && !strings.Contains(strings.ToLower(r.Name), keyword) &&
				!strings.Contains(strings.ToLower(r.DisplayName), keyword) {
				continue
			}
			_, r.Registered = s.defs[r.Name]
			skills = append(skills, r)
		}

		out, _ := json.Marshal(map[string]interface{}{
			"count":  len(skills),
			"skills": skills,
		})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_agent_skill",
			Description: "Adds a new entry to the agent skill catalog. " +
				"The name MUST exactly match a registered tool handler name (verify with list_available_tools). " +
				"Once added, the skill appears in the agent configuration UI and can be assigned to agents.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":         {Type: "string", Description: "Tool handler name — must match exactly (required)"},
					"display_name": {Type: "string", Description: "Human-readable label shown in the UI (required)"},
					"description":  {Type: "string", Description: "Short description of what this skill does"},
					"category":     {Type: "string", Description: "Category for grouping: system|reporting|scheduling|web|messaging|data|ai|testing|devops (default: 'system')"},
				},
				Required: []string{"name", "display_name"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		name := stringArg(args, "name")
		displayName := stringArg(args, "display_name")
		description := stringArg(args, "description")
		category := stringArg(args, "category")
		if name == "" || displayName == "" {
			return "", fmt.Errorf("name and display_name are required")
		}
		if category == "" {
			category = "system"
		}

		// Warn if tool handler is not registered
		_, registered := s.defs[name]

		id := uuid.New().String()
		now := time.Now().UTC()
		_, err := s.sqlDB.ExecContext(ctx,
			`INSERT INTO agent_skills (id, name, display_name, description, category, active, createdby, createdon, modifiedby, modifiedon)
			 VALUES ($1,$2,$3,$4,$5,true,'agent',$6,'agent',$6)
			 ON CONFLICT (name) DO NOTHING`,
			id, name, displayName, description, category, now)
		if err != nil {
			return "", fmt.Errorf("failed to create agent skill: %w", err)
		}

		out, _ := json.Marshal(map[string]interface{}{
			"id":                  id,
			"name":                name,
			"display_name":        displayName,
			"category":            category,
			"registered_in_runtime": registered,
			"warning": func() string {
				if !registered {
					return fmt.Sprintf("WARNING: no tool handler named %q is registered in the runtime — the skill will appear in the UI but will fail to execute", name)
				}
				return ""
			}(),
		})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "update_agent_skill",
			Description: "Updates an existing agent skill catalog entry. " +
				"Only supplied fields are changed. Deactivating a skill (active: false) hides it from the UI picker but does not remove existing assignments.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":           {Type: "string", Description: "Skill ID to update (required). Use list_agent_skills to find IDs."},
					"display_name": {Type: "string", Description: "New display label"},
					"description":  {Type: "string", Description: "New description"},
					"category":     {Type: "string", Description: "New category"},
					"active":       {Type: "boolean", Description: "Set to false to hide from UI (does not delete)"},
				},
				Required: []string{"id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}
		sets := []string{"modifiedby='agent'", fmt.Sprintf("modifiedon=$1")}
		vals := []interface{}{time.Now().UTC()}
		idx := 2

		addStr := func(col, key string) {
			if v := stringArg(args, key); v != "" {
				sets = append(sets, fmt.Sprintf("%s=$%d", col, idx))
				vals = append(vals, v)
				idx++
			}
		}
		addStr("display_name", "display_name")
		addStr("description", "description")
		addStr("category", "category")
		if v, ok := args["active"].(bool); ok {
			sets = append(sets, fmt.Sprintf("active=$%d", idx))
			vals = append(vals, v)
			idx++
		}

		vals = append(vals, id)
		query := fmt.Sprintf("UPDATE agent_skills SET %s WHERE id=$%d", strings.Join(sets, ","), idx)
		res, err := s.sqlDB.ExecContext(ctx, query, vals...)
		if err != nil {
			return "", fmt.Errorf("failed to update agent skill: %w", err)
		}
		n, _ := res.RowsAffected()
		out, _ := json.Marshal(map[string]interface{}{"id": id, "rows_updated": n})
		return string(out), nil
	})
}

// ═══════════════════════════════════════════════════════════════════════════════
// SKILL ASSIGNMENT (agent_definition_skills join table)
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerSkillAssignmentTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "list_agent_assigned_skills",
			Description: "Lists all skills currently assigned to a specific agent definition.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"agent_id": {Type: "string", Description: "Agent definition ID (required). Use list_database_schemas + query_database to find agent IDs, or search agent_definitions table."},
				},
				Required: []string{"agent_id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		agentID := stringArg(args, "agent_id")
		if agentID == "" {
			return "", fmt.Errorf("agent_id is required")
		}
		rows, err := s.sqlDB.QueryContext(ctx,
			`SELECT s.id, s.name, s.display_name, s.description, s.category, s.active
			 FROM agent_skills s
			 JOIN agent_definition_skills ads ON ads.agent_skill_id = s.id
			 WHERE ads.agent_definition_id = $1
			 ORDER BY s.category, s.name`, agentID)
		if err != nil {
			return "", fmt.Errorf("query error: %w", err)
		}
		defer rows.Close()
		return s.rowsToJSON(rows)
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "assign_skill_to_agent",
			Description: "Assigns a skill from the catalog to an agent definition. " +
				"The skill will be available to that agent on its next run. " +
				"Use list_agent_skills to find skill names and list_agent_assigned_skills to avoid duplicates.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"agent_id":   {Type: "string", Description: "Agent definition ID (required)"},
					"skill_name": {Type: "string", Description: "Skill name to assign (required) — must exist in agent_skills catalog"},
				},
				Required: []string{"agent_id", "skill_name"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		agentID := stringArg(args, "agent_id")
		skillName := stringArg(args, "skill_name")
		if agentID == "" || skillName == "" {
			return "", fmt.Errorf("agent_id and skill_name are required")
		}
		_, err := s.sqlDB.ExecContext(ctx,
			`INSERT INTO agent_definition_skills (agent_definition_id, agent_skill_id)
			 SELECT $1, id FROM agent_skills WHERE name = $2
			 ON CONFLICT DO NOTHING`,
			agentID, skillName)
		if err != nil {
			return "", fmt.Errorf("failed to assign skill: %w", err)
		}
		out, _ := json.Marshal(map[string]string{
			"agent_id": agentID, "skill_name": skillName, "status": "assigned",
		})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "remove_skill_from_agent",
			Description: "Removes a skill assignment from an agent definition. " +
				"The skill remains in the catalog and can be re-assigned later. " +
				"The agent will no longer have access to this tool on its next run.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"agent_id":   {Type: "string", Description: "Agent definition ID (required)"},
					"skill_name": {Type: "string", Description: "Skill name to remove (required)"},
				},
				Required: []string{"agent_id", "skill_name"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		agentID := stringArg(args, "agent_id")
		skillName := stringArg(args, "skill_name")
		if agentID == "" || skillName == "" {
			return "", fmt.Errorf("agent_id and skill_name are required")
		}
		res, err := s.sqlDB.ExecContext(ctx,
			`DELETE FROM agent_definition_skills
			 WHERE agent_definition_id = $1
			   AND agent_skill_id = (SELECT id FROM agent_skills WHERE name = $2)`,
			agentID, skillName)
		if err != nil {
			return "", fmt.Errorf("failed to remove skill: %w", err)
		}
		n, _ := res.RowsAffected()
		out, _ := json.Marshal(map[string]interface{}{
			"agent_id": agentID, "skill_name": skillName, "removed": n > 0,
		})
		return string(out), nil
	})
}
