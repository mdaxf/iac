package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// ToolDefinition is the JSON schema passed to the LLM for function calling
type ToolDefinition struct {
	Type     string              `json:"type"` // always "function"
	Function ToolFunctionDef     `json:"function"`
}

type ToolFunctionDef struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Parameters  ToolParameterSchema `json:"parameters"`
}

type ToolParameterSchema struct {
	Type       string                        `json:"type"` // always "object"
	Properties map[string]ToolPropertySchema `json:"properties"`
	Required   []string                      `json:"required,omitempty"`
}

type ToolPropertySchema struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ToolHandler is a function that executes a tool with JSON args and returns a string result
type ToolHandler func(ctx context.Context, args map[string]interface{}) (string, error)

// AgentToolRegistryService manages available tools for agents
type AgentToolRegistryService struct {
	db          *gorm.DB
	sqlDB       *sql.DB
	docDB       documents.DocumentDB // MongoDB — nil when document DB is not configured
	openAIKey   string               // OpenAI API key for AI-powered tools
	openAIModel string               // OpenAI model for AI-powered tools
	handlers    map[string]ToolHandler
	defs        map[string]ToolDefinition
	iLog        logger.Log
}

var (
	globalToolRegistry *AgentToolRegistryService
)

func GetGlobalToolRegistry() *AgentToolRegistryService {
	return globalToolRegistry
}

func SetGlobalToolRegistry(s *AgentToolRegistryService) {
	globalToolRegistry = s
}

// NewAgentToolRegistryService creates and registers all built-in IAC tools.
// docDB is optional (pass nil when MongoDB is not configured).
// openAIKey and openAIModel enable AI-powered tools (query_database_nl, generate_bpm, etc.).
func NewAgentToolRegistryService(db *gorm.DB, sqlDB *sql.DB, docDB documents.DocumentDB, openAIKey, openAIModel string) *AgentToolRegistryService {
	s := &AgentToolRegistryService{
		db:          db,
		sqlDB:       sqlDB,
		docDB:       docDB,
		openAIKey:   openAIKey,
		openAIModel: openAIModel,
		handlers:    make(map[string]ToolHandler),
		defs:        make(map[string]ToolDefinition),
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "AgentToolRegistryService",
		},
	}
	s.registerBuiltins()
	return s
}

// registerBuiltins registers all built-in IAC tools
func (s *AgentToolRegistryService) registerBuiltins() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "get_current_time",
			Description: "Returns the current server date and time in ISO 8601 format.",
			Parameters:  ToolParameterSchema{Type: "object", Properties: map[string]ToolPropertySchema{}},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		return time.Now().UTC().Format(time.RFC3339), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "get_system_health",
			Description: "Returns the current health status of the IAC system, including database connectivity.",
			Parameters:  ToolParameterSchema{Type: "object", Properties: map[string]ToolPropertySchema{}},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		result := map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		if s.sqlDB != nil {
			if err := s.sqlDB.PingContext(ctx); err != nil {
				result["database"] = "error: " + err.Error()
			} else {
				result["database"] = "connected"
			}
		}
		b, _ := json.Marshal(result)
		return string(b), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "list_database_schemas",
			Description: "Lists all tables and their schemas available in the IAC database (default) or a named connection. " +
				"Always call this first before writing SQL queries so you know which tables exist. " +
				"Returns table names grouped by schema. Use connection='iac' (or omit) for the IAC system database.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"connection": {
						Type: "string",
						Description: "Optional. Database connection name. Omit or use 'iac' to query the default IAC database. " +
							"Only specify a different name if the user has explicitly mentioned another database.",
					},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		// For now, always query the IAC database via information_schema.
		// Cross-DB routing can be wired in when a dbconnections registry is added.
		rows, err := s.sqlDB.QueryContext(ctx, `
			SELECT table_schema, table_name, table_type
			FROM information_schema.tables
			WHERE table_schema NOT IN ('information_schema', 'pg_catalog')
			ORDER BY table_schema, table_name`)
		if err != nil {
			return "", fmt.Errorf("schema discovery error: %w", err)
		}
		defer rows.Close()

		type TableInfo struct {
			Schema string `json:"schema"`
			Table  string `json:"table"`
			Type   string `json:"type"`
		}
		var tables []TableInfo
		for rows.Next() {
			var t TableInfo
			if err := rows.Scan(&t.Schema, &t.Table, &t.Type); err == nil {
				tables = append(tables, t)
			}
		}
		if tables == nil {
			tables = []TableInfo{}
		}
		result := map[string]interface{}{
			"connection": "iac",
			"tables":     tables,
			"hint":       "Use query_database with connection='iac' (or omit connection) to query any of these tables",
		}
		b, _ := json.Marshal(result)
		return string(b), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "query_database",
			Description: "Executes a read-only SQL SELECT query and returns up to 200 rows as JSON. " +
				"Only SELECT statements are allowed — never INSERT, UPDATE, DELETE, or DROP. " +
				"Use connection='iac' (or omit) to query the IAC system database. " +
				"Call list_database_schemas first if you need to discover available tables. " +
				"Always include a LIMIT clause to avoid fetching too many rows.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"connection": {
						Type:        "string",
						Description: "Optional. Database connection name. Omit or use 'iac' for the IAC system database.",
					},
					"sql": {
						Type: "string",
						Description: "A valid SQL SELECT statement. Must begin with SELECT. " +
							"Include a LIMIT clause (e.g. LIMIT 50) to control result size.",
					},
				},
				Required: []string{"sql"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		query, _ := args["sql"].(string)
		connection, _ := args["connection"].(string)
		if query == "" {
			return "", fmt.Errorf("sql is required")
		}
		if connection == "" {
			connection = "iac"
		}
		return s.executeReadOnlyQuery(ctx, connection, query)
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_jobs",
			Description: "Lists scheduled jobs in IAC with their status, handler, and next run time.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"limit": {Type: "integer", Description: "Maximum number of jobs to return (default 20)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		limit := 20
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}
		rows, err := s.sqlDB.QueryContext(ctx,
			`SELECT id, name, handler, enabled, last_run_at, next_run_at, cron_expression
			 FROM jobs WHERE active = true ORDER BY name LIMIT $1`, limit)
		if err != nil {
			if isTableNotFoundErr(err) {
				b, _ := json.Marshal(map[string]interface{}{"rows": []interface{}{}, "note": "jobs table not yet configured"})
				return string(b), nil
			}
			return "", fmt.Errorf("query error: %w", err)
		}
		defer rows.Close()
		return s.rowsToJSON(rows)
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_agents",
			Description: "Lists all configured AI agents in IAC.",
			Parameters:  ToolParameterSchema{Type: "object", Properties: map[string]ToolPropertySchema{}},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		rows, err := s.sqlDB.QueryContext(ctx,
			`SELECT id, name, description, model_name, is_default FROM agent_definitions WHERE active = true ORDER BY name`)
		if err != nil {
			return "", fmt.Errorf("query error: %w", err)
		}
		defer rows.Close()
		return s.rowsToJSON(rows)
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_reports",
			Description: "Lists available reports in IAC.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"limit": {Type: "integer", Description: "Maximum number of reports to return (default 20)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		limit := 20
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}
		rows, err := s.sqlDB.QueryContext(ctx,
			`SELECT id, name, description FROM reports WHERE active = true ORDER BY name LIMIT $1`, limit)
		if err != nil {
			if isTableNotFoundErr(err) {
				b, _ := json.Marshal(map[string]interface{}{"rows": []interface{}{}, "note": "reports table not yet configured"})
				return string(b), nil
			}
			return "", fmt.Errorf("query error: %w", err)
		}
		defer rows.Close()
		return s.rowsToJSON(rows)
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_integrations",
			Description: "Lists integration hub connections configured in IAC.",
			Parameters:  ToolParameterSchema{Type: "object", Properties: map[string]ToolPropertySchema{}},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		rows, err := s.sqlDB.QueryContext(ctx,
			`SELECT id, name, description, enabled FROM integrationhubs WHERE active = true ORDER BY name`)
		if err != nil {
			if isTableNotFoundErr(err) {
				b, _ := json.Marshal(map[string]interface{}{"rows": []interface{}{}, "note": "integrationhubs table not yet configured"})
				return string(b), nil
			}
			return "", fmt.Errorf("query error: %w", err)
		}
		defer rows.Close()
		return s.rowsToJSON(rows)
	})

	s.registerWebTools()
	s.registerMessagingTools()
	s.registerSystemTools()
	s.registerNotificationTools()
	s.registerGenerationTools()
	s.registerAIDatabaseTools()
	s.registerCodegenTools()
	s.registerPlanningTools()
	s.registerScheduleTools()
	s.registerReportingTools()
	s.registerOSTools()
	s.registerFileSystemTools()
	s.registerToolManagementTools()
	s.registerTestingTools()
	s.registerMfgQualityTools()
	s.registerMfgProcessTools()
	s.registerDeploymentTools()
}

// register adds a tool definition and its handler
func (s *AgentToolRegistryService) register(def ToolDefinition, handler ToolHandler) {
	s.defs[def.Function.Name] = def
	s.handlers[def.Function.Name] = handler
}

// GetToolDefinitions returns LLM-ready tool definitions for the requested tool names.
// If enabledTools is empty, all tools are returned.
func (s *AgentToolRegistryService) GetToolDefinitions(enabledTools []string) []ToolDefinition {
	if len(enabledTools) == 0 {
		result := make([]ToolDefinition, 0, len(s.defs))
		for _, d := range s.defs {
			result = append(result, d)
		}
		return result
	}
	result := make([]ToolDefinition, 0, len(enabledTools))
	for _, name := range enabledTools {
		if d, ok := s.defs[name]; ok {
			result = append(result, d)
		}
	}
	return result
}

// GetToolDefinitionsForSkills returns LLM-ready tool definitions for the given AgentSkill slice.
// When a skill carries a non-empty Content field (stored in MongoDB), that content is appended
// to the tool's description so the LLM sees the full schema/instructions from the skill catalog
// rather than only the hardcoded Go description.
func (s *AgentToolRegistryService) GetToolDefinitionsForSkills(skills []models.AgentSkill) []ToolDefinition {
	result := make([]ToolDefinition, 0, len(skills))
	for _, skill := range skills {
		d, ok := s.defs[skill.Name]
		if !ok {
			continue
		}
		if skill.Content != "" {
			// Clone and enrich — do not mutate the registry entry
			enriched := d
			enriched.Function.Description = d.Function.Description + "\n\n" + skill.Content
			result = append(result, enriched)
		} else {
			result = append(result, d)
		}
	}
	return result
}

// ExecuteTool executes the named tool with JSON-decoded args
func (s *AgentToolRegistryService) ExecuteTool(ctx context.Context, name string, argsJSON string) (string, error) {
	handler, ok := s.handlers[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	var args map[string]interface{}
	if argsJSON != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return "", fmt.Errorf("invalid tool args JSON: %w", err)
		}
	}
	result, err := handler(ctx, args)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Tool %s execution failed: %v", name, err))
		return "", err
	}
	return result, nil
}

// executeReadOnlyQuery executes a SELECT on the named connection.
// Currently only the "iac" (default) connection is supported; other aliases
// will be routed to the same IAC database until a dbconnections registry is added.
func (s *AgentToolRegistryService) executeReadOnlyQuery(ctx context.Context, connection, query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("sql query is required")
	}
	// Safety: only allow SELECT statements
	trimmed := strings.TrimSpace(query)
	upper := strings.ToUpper(trimmed)
	if !strings.HasPrefix(upper, "SELECT") {
		return "", fmt.Errorf("only SELECT queries are allowed; received statement starting with: %.30s", trimmed)
	}

	const maxRows = 200
	rows, err := s.sqlDB.QueryContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("failed to read columns: %w", err)
	}

	var result []map[string]interface{}
	rowCount := 0
	for rows.Next() && rowCount < maxRows {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = values[i]
		}
		result = append(result, row)
		rowCount++
	}
	if result == nil {
		result = []map[string]interface{}{}
	}

	truncated := false
	if rowCount == maxRows {
		// Check if there are more rows
		truncated = rows.Next()
	}

	out := map[string]interface{}{
		"connection": connection,
		"row_count":  len(result),
		"rows":       result,
	}
	if truncated {
		out["warning"] = fmt.Sprintf("Results truncated at %d rows. Add a LIMIT clause to your query to fetch fewer rows.", maxRows)
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}

// isTableNotFoundErr returns true when a SQL error indicates the queried table does not exist.
// Handles PostgreSQL ("relation ... does not exist") and MySQL ("Table ... doesn't exist").
func isTableNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "does not exist") || strings.Contains(msg, "doesn't exist") || strings.Contains(msg, "no such table")
}

// rowsToJSON converts sql.Rows to a JSON array string
func (s *AgentToolRegistryService) rowsToJSON(rows *sql.Rows) (string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}
	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = values[i]
		}
		result = append(result, row)
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	b, err := json.Marshal(result)
	return string(b), err
}
