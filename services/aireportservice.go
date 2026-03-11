package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mdaxf/iac/llm"
)

// AIReportService handles AI-powered report generation
type AIReportService struct {
	OpenAIKey   string
	OpenAIModel string
}

// NewAIReportService creates a new AI report service
func NewAIReportService(openAIKey, openAIModel string) *AIReportService {
	return &AIReportService{
		OpenAIKey:   openAIKey,
		OpenAIModel: openAIModel,
	}
}

// Text2SQLRequest represents a natural language question
type Text2SQLRequest struct {
	Question    string `json:"question"`
	DatabaseID  string `json:"database_id"`
	AutoExecute bool   `json:"auto_execute"`
	ThreadID    string `json:"thread_id,omitempty"`
}

// Text2SQLResponse represents the AI-generated SQL
type Text2SQLResponse struct {
	SQL         string                   `json:"sql"`
	Explanation string                   `json:"explanation"`
	Confidence  float64                  `json:"confidence"`
	TablesUsed  []string                 `json:"tables_used"`
	ColumnsUsed []string                 `json:"columns_used"`
	Reasoning   string                   `json:"reasoning"`
	QueryType   string                   `json:"query_type"`
	Data        []map[string]interface{} `json:"data,omitempty"`
	RowCount    int                      `json:"row_count,omitempty"`
}

// ComponentRecommendation represents an AI-recommended report component
type ComponentRecommendation struct {
	ComponentType string                 `json:"component_type"`
	Name          string                 `json:"name"`
	X             float64                `json:"x"`
	Y             float64                `json:"y"`
	Width         float64                `json:"width"`
	Height        float64                `json:"height"`
	DataConfig    map[string]interface{} `json:"data_config"`
	ChartType     string                 `json:"chart_type,omitempty"`
	ChartConfig   map[string]interface{} `json:"chart_config,omitempty"`
	StyleConfig   map[string]interface{} `json:"style_config,omitempty"`
}

// ReportGenerationRequest represents a request to generate a report from data
type ReportGenerationRequest struct {
	Question   string                   `json:"question"`
	SQL        string                   `json:"sql"`
	Data       []map[string]interface{} `json:"data"`
	ReportName string                   `json:"report_name,omitempty"`
}

// ReportGenerationResponse represents the AI-generated report structure
type ReportGenerationResponse struct {
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	Components  []ComponentRecommendation `json:"components"`
	Insights    []string                  `json:"insights"`
}


// GenerateSQL generates SQL from natural language using AI
func (s *AIReportService) GenerateSQL(ctx context.Context, request Text2SQLRequest, schemaInfo string) (*Text2SQLResponse, error) {
	// Build system prompt
	systemPrompt := `You are an expert SQL query generator that converts natural language questions into accurate SQL queries.

CORE PRINCIPLES:
1. Generate syntactically correct SQL for the specific database system
2. Use proper table and column names from the provided schema
3. Apply appropriate filters, joins, and aggregations
4. Ensure queries are efficient and follow best practices
5. Provide clear explanations for your reasoning

RESPONSE FORMAT (JSON):
{
  "sql": "The generated SQL query",
  "explanation": "Clear explanation of what the query does",
  "confidence": 0.95,
  "tables_used": ["table1", "table2"],
  "columns_used": ["column1", "column2"],
  "reasoning": "Step-by-step reasoning process",
  "query_type": "SELECT"
}

CRITICAL REQUIREMENTS:
- **ONLY** generate SELECT queries (read-only data retrieval)
- **NEVER** generate INSERT, UPDATE, DELETE, DROP, ALTER, CREATE, TRUNCATE, or any data modification operations
- If the user asks to modify data, explain that only data retrieval is allowed and suggest a SELECT query alternative
- Add LIMIT clause if not specified (default: 100)
- The query will be validated and rejected if it contains any write operations`

	// Build user prompt
	userPrompt := fmt.Sprintf(`### DATABASE SCHEMA ###
%s

### QUESTION ###
User's Question: %s

### INSTRUCTIONS ###
1. Analyze the question to understand what data is being requested
2. Identify the relevant tables and columns from the schema
3. Determine the appropriate joins, filters, and aggregations
4. Generate the SQL query following best practices
5. Provide reasoning for your choices

Respond with JSON only, no additional text.`, schemaInfo, request.Question)

	// Call OpenAI API
	response, err := s.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	// Parse response
	var result Text2SQLResponse
	cleanedResponse := cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w (content: %s)", err, cleanedResponse)
	}

	// Validate SQL is read-only
	if !s.isReadOnlySQL(result.SQL) {
		fmt.Printf("[WARN] AI generated non-read-only SQL: %s\n", result.SQL)
		return nil, fmt.Errorf("the generated query contains write operations (INSERT/UPDATE/DELETE). Only read-only SELECT queries are allowed. Please rephrase your question to request data retrieval instead of data modification")
	}

	return &result, nil
}

// GenerateReport generates report structure from query results
func (s *AIReportService) GenerateReport(ctx context.Context, request ReportGenerationRequest) (*ReportGenerationResponse, error) {
	// Analyze data structure
	columns := []string{}
	numericColumns := []string{}
	dateColumns := []string{}
	textColumns := []string{}

	if len(request.Data) > 0 {
		firstRow := request.Data[0]
		for col, val := range firstRow {
			columns = append(columns, col)

			switch val.(type) {
			case int, int64, float64:
				numericColumns = append(numericColumns, col)
			case string:
				// Check if it's a date string
				if s.isDateString(val.(string)) {
					dateColumns = append(dateColumns, col)
				} else {
					textColumns = append(textColumns, col)
				}
			}
		}
	}

	systemPrompt := `You are an expert data analyst and report designer. You analyze query results and recommend appropriate visualizations.

COMPONENT TYPES:
- table: Data table
- chart: Chart (line, bar, pie, area, scatter)
- text: Metric card or text content

RESPONSE FORMAT (JSON):
{
  "title": "Report Title",
  "description": "Report description",
  "components": [
    {
      "component_type": "table",
      "name": "Data Table",
      "x": 50,
      "y": 50,
      "width": 800,
      "height": 400,
      "data_config": {"query": "SQL here", "fields": ["col1", "col2"]}
    }
  ],
  "insights": ["Insight 1", "Insight 2"]
}`

	dataInfo := fmt.Sprintf(`Rows: %d
Columns: %d
Numeric columns: %v
Date columns: %v
Text columns: %v`, len(request.Data), len(columns), numericColumns, dateColumns, textColumns)

	userPrompt := fmt.Sprintf(`### QUERY ###
%s

### DATA STRUCTURE ###
%s

### QUESTION ###
%s

### INSTRUCTIONS ###
1. Analyze the data structure
2. Recommend appropriate components (table, charts, metrics)
3. Suggest chart types based on data types
4. Provide insights about the data

Respond with JSON only.`, request.SQL, dataInfo, request.Question)

	// Call OpenAI API
	response, err := s.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	// Parse response
	var result ReportGenerationResponse
	cleanedResponse := cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w (content: %s)", err, cleanedResponse)
	}

	return &result, nil
}

// callOpenAI calls the LLM for this service, routing through the unified llm
// package which reads aiconfig.json (text2sql use case). Falls back to
// OpenAI using s.OpenAIKey / s.OpenAIModel when config is unavailable.
func (s *AIReportService) callOpenAI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	messages := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": userPrompt},
	}
	temp := 0.1 // low temperature for consistent SQL generation
	return llm.CallLLM(ctx, "text2sql", s.OpenAIKey, s.OpenAIModel, messages, temp)
}

// isReadOnlySQL checks if SQL is a read-only SELECT query
func (s *AIReportService) isReadOnlySQL(sql string) bool {
	sqlLower := strings.ToLower(strings.TrimSpace(sql))

	// Must start with SELECT
	if !strings.HasPrefix(sqlLower, "select") {
		return false
	}

	// Block dangerous operations - check for whole words only
	// Use word boundaries to avoid false positives (e.g., "DATE" shouldn't match "UPDATE")
	dangerous := []string{
		"insert ", "insert(", " insert ",
		"update ", "update(", " update ",
		"delete ", "delete(", " delete ",
		"drop ", "drop(", " drop ",
		"alter ", "alter(", " alter ",
		"create ", "create(", " create ",
		"truncate ", "truncate(", " truncate ",
		"exec ", "exec(", " exec ",
		"execute ", "execute(", " execute ",
	}
	for _, keyword := range dangerous {
		if strings.Contains(sqlLower, keyword) {
			return false
		}
	}

	return true
}

// isDateString checks if a string represents a date
func (s *AIReportService) isDateString(str string) bool {
	// Try common date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"01-02-2006",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, str); err == nil {
			return true
		}
	}

	return false
}
