package services

// AI-powered database tools — natural language to SQL query execution
// using AIReportService (text-to-SQL) and schema introspection.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// registerAIDatabaseTools registers AI-powered database query tools
func (s *AgentToolRegistryService) registerAIDatabaseTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "query_database_nl",
			Description: "Converts a natural language question into a SQL SELECT query and executes it. " +
				"Uses AI to generate the appropriate query based on the live database schema — no SQL knowledge required. " +
				"Returns the generated SQL, an explanation, confidence score, and the query results. " +
				"Use query_database when you already know the exact SQL. Use this tool when you only know what data you need.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"question": {
						Type:        "string",
						Description: "Natural language question about the data (e.g. 'How many agents are active?', 'List all jobs scheduled for today').",
					},
					"database": {
						Type:        "string",
						Description: "Optional. Database connection alias. Omit or use 'iac' for the IAC system database.",
					},
				},
				Required: []string{"question"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.openAIKey == "" {
			return "", fmt.Errorf("OpenAI API key is not configured; query_database_nl requires AI capabilities")
		}

		question, _ := args["question"].(string)
		if question == "" {
			return "", fmt.Errorf("question is required")
		}

		database, _ := args["database"].(string)
		if database == "" {
			database = "iac"
		}

		// Gather schema info for AI context
		schemaInfo, _ := s.getSchemaInfo(ctx)

		// Generate SQL via AIReportService
		svc := NewAIReportService(s.openAIKey, s.openAIModel)
		resp, err := svc.GenerateSQL(ctx, Text2SQLRequest{
			Question:    question,
			DatabaseID:  database,
			AutoExecute: false,
		}, schemaInfo)
		if err != nil {
			return "", fmt.Errorf("SQL generation failed: %w", err)
		}

		// Execute the generated SQL
		queryResult, execErr := s.executeReadOnlyQuery(ctx, database, resp.SQL)

		out := map[string]interface{}{
			"question":    question,
			"sql":         resp.SQL,
			"explanation": resp.Explanation,
			"confidence":  resp.Confidence,
			"tables_used": resp.TablesUsed,
			"reasoning":   resp.Reasoning,
		}
		if execErr != nil {
			out["execution_error"] = execErr.Error()
			out["hint"] = "The generated SQL failed to execute. Try refining your question or use query_database with a hand-written SELECT."
		} else {
			var data map[string]interface{}
			if json.Unmarshal([]byte(queryResult), &data) == nil {
				out["data"] = data
			}
		}

		b, _ := json.Marshal(out)
		return string(b), nil
	})
}

// getSchemaInfo fetches a compact schema description from information_schema for AI SQL generation context.
// Returns a multi-line string: "schema.table (col1 type, col2 type, ...)\n..."
func (s *AgentToolRegistryService) getSchemaInfo(ctx context.Context) (string, error) {
	rows, err := s.sqlDB.QueryContext(ctx, `
		SELECT t.table_schema, t.table_name, c.column_name, c.data_type
		FROM information_schema.tables t
		JOIN information_schema.columns c
		  ON t.table_schema = c.table_schema AND t.table_name = c.table_name
		WHERE t.table_schema NOT IN ('information_schema', 'pg_catalog')
		  AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_schema, t.table_name, c.ordinal_position
		LIMIT 2000`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type colEntry struct {
		col   string
		dtype string
	}
	tableMap := make(map[string][]colEntry)
	tableOrder := []string{}
	seen := map[string]bool{}

	for rows.Next() {
		var schema, table, col, dtype string
		if err := rows.Scan(&schema, &table, &col, &dtype); err != nil {
			continue
		}
		key := schema + "." + table
		if !seen[key] {
			seen[key] = true
			tableOrder = append(tableOrder, key)
		}
		tableMap[key] = append(tableMap[key], colEntry{col, dtype})
	}

	var sb strings.Builder
	for _, key := range tableOrder {
		cols := tableMap[key]
		sb.WriteString(key + " (")
		for i, c := range cols {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(c.col + " " + c.dtype)
		}
		sb.WriteString(")\n")
	}
	return sb.String(), nil
}
