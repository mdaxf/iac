package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
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

// OpenAIMessage represents a message in the OpenAI API
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents a request to the OpenAI API
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
}

// OpenAIResponse represents a response from the OpenAI API
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
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

IMPORTANT:
- Only generate SELECT queries (read-only)
- Never generate INSERT, UPDATE, DELETE, DROP, ALTER, or other dangerous operations
- Add LIMIT clause if not specified (default: 100)`

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
		return nil, fmt.Errorf("generated SQL is not read-only")
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

// callOpenAI makes a request to the OpenAI API
func (s *AIReportService) callOpenAI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	reqBody := OpenAIRequest{
		Model: s.OpenAIModel,
		Messages: []OpenAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.1, // Low temperature for consistent, factual responses
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.OpenAIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error: status %d", resp.StatusCode)
	}

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// isReadOnlySQL checks if SQL is a read-only SELECT query
func (s *AIReportService) isReadOnlySQL(sql string) bool {
	sqlLower := strings.ToLower(strings.TrimSpace(sql))

	// Must start with SELECT
	if !strings.HasPrefix(sqlLower, "select") {
		return false
	}

	// Block dangerous operations
	dangerous := []string{"insert", "update", "delete", "drop", "alter", "create", "truncate", "exec", "execute"}
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
