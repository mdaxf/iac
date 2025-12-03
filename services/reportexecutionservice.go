// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mdaxf/iac/documents/schema"
	"github.com/mdaxf/iac/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportExecutionService handles report execution logic
type ReportExecutionService struct {
	schemaMetadataService *SchemaMetadataService
	iLog                  logger.Log
}

// NewReportExecutionService creates a new report execution service
func NewReportExecutionService(schemaMetadataService *SchemaMetadataService) *ReportExecutionService {
	return &ReportExecutionService{
		schemaMetadataService: schemaMetadataService,
		iLog:                  logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "ReportExecutionService"},
	}
}

// ExecutionRequest represents a report execution request
type ExecutionRequest struct {
	Parameters   map[string]interface{} `json:"parameters"`
	OutputFormat string                 `json:"output_format"`
}

// ExecutionResult represents the result of a report execution
type ExecutionResult struct {
	ReportID        string                 `json:"report_id"`
	ReportName      string                 `json:"report_name"`
	ExecutedBy      string                 `json:"executed_by"`
	ExecutedAt      time.Time              `json:"executed_at"`
	Parameters      map[string]interface{} `json:"parameters"`
	OutputFormat    string                 `json:"output_format"`
	Status          string                 `json:"status"`
	ExecutionTimeMs int64                  `json:"execution_time_ms"`
	Data            ExecutionData          `json:"data"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
}

// ExecutionData wraps the datasource results to match frontend expectations
type ExecutionData struct {
	Datasources map[string]DatasourceExecutionResult `json:"datasources"`
}

// DatasourceExecutionResult represents the execution result of a single datasource
type DatasourceExecutionResult struct {
	Alias           string                   `json:"alias"`
	Columns         []string                 `json:"columns"`
	Rows            []map[string]interface{} `json:"rows"`            // Primary field for frontend TableRenderer
	Data            []map[string]interface{} `json:"data"`            // Backward compatibility
	TotalRows       int                      `json:"total_rows"`
	ExecutionTimeMs int64                    `json:"execution_time_ms"`
	SQL             string                   `json:"sql,omitempty"`
	Error           string                   `json:"error,omitempty"`
}

// ExecuteReport executes a report and returns the results
func (s *ReportExecutionService) ExecuteReport(
	ctx context.Context,
	report *schema.ReportDocument,
	request ExecutionRequest,
	executedBy string,
) (*ExecutionResult, error) {
	startTime := time.Now()
	s.iLog.Info(fmt.Sprintf("Starting execution of report: %s (ID: %s)", report.Name, report.ID))
	s.iLog.Debug(fmt.Sprintf("Report has %d datasources configured", len(report.Datasources)))

	result := &ExecutionResult{
		ReportID:     report.ID,
		ReportName:   report.Name,
		ExecutedBy:   executedBy,
		ExecutedAt:   time.Now(),
		Parameters:   request.Parameters,
		OutputFormat: request.OutputFormat,
		Status:       "running",
		Data: ExecutionData{
			Datasources: make(map[string]DatasourceExecutionResult),
		},
	}

	// Execute each datasource
	if len(report.Datasources) == 0 {
		s.iLog.Warn(fmt.Sprintf("Report %s has no datasources", report.ID))
		result.Status = "success"
		result.ExecutionTimeMs = time.Since(startTime).Milliseconds()
		return result, nil
	}

	s.iLog.Debug(fmt.Sprintf("Executing %d datasources for report %s", len(report.Datasources), report.ID))

	// Execute datasources
	hasError := false
	for _, datasource := range report.Datasources {
		s.iLog.Debug(fmt.Sprintf("Executing datasource: %s (Type: %s)", datasource.Alias, datasource.QueryType))

		datasourceResult, err := s.executeDatasource(ctx, &datasource, request.Parameters)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to execute datasource %s: %v", datasource.Alias, err))
			datasourceResult.Error = err.Error()
			hasError = true
		}

		result.Data.Datasources[datasource.Alias] = *datasourceResult
	}

	// Set final status
	if hasError {
		result.Status = "partial_success"
	} else {
		result.Status = "success"
	}

	result.ExecutionTimeMs = time.Since(startTime).Milliseconds()
	s.iLog.Info(fmt.Sprintf("Report execution completed: %s in %dms (Status: %s)",
		report.ID, result.ExecutionTimeMs, result.Status))

	return result, nil
}

// executeDatasource executes a single datasource and returns its data
func (s *ReportExecutionService) executeDatasource(
	ctx context.Context,
	datasource *schema.ReportDatasourceDoc,
	parameters map[string]interface{},
) (*DatasourceExecutionResult, error) {
	startTime := time.Now()

	result := &DatasourceExecutionResult{
		Alias: datasource.Alias,
	}

	// Validate database alias
	if datasource.DatabaseAlias == "" {
		return result, fmt.Errorf("database alias is required for datasource %s", datasource.Alias)
	}

	var sqlQuery string
	var err error

	// Determine SQL based on query type
	s.iLog.Info(fmt.Sprintf(">>> Datasource %s: QueryType='%s', HasCustomSQL=%v",
		datasource.Alias, datasource.QueryType, datasource.CustomSQL != ""))

	switch datasource.QueryType {
	case "custom", "custom_sql":
		// Use custom SQL
		s.iLog.Info(fmt.Sprintf(">>> PATH 1: Using CUSTOM SQL for datasource %s", datasource.Alias))
		sqlQuery = datasource.CustomSQL
		if sqlQuery == "" {
			return result, fmt.Errorf("custom SQL is empty for datasource %s", datasource.Alias)
		}

		// Apply parameter substitution
		sqlQuery = s.substituteParameters(sqlQuery, parameters)
		result.SQL = sqlQuery

	case "visual":
		// Check if custom SQL is available as fallback
		if datasource.CustomSQL != "" {
			s.iLog.Info(fmt.Sprintf(">>> PATH 2: Visual query with customSQL FALLBACK for datasource %s", datasource.Alias))
			sqlQuery = datasource.CustomSQL
			// Apply parameter substitution
			sqlQuery = s.substituteParameters(sqlQuery, parameters)
			result.SQL = sqlQuery
		} else {
			s.iLog.Info(fmt.Sprintf(">>> PATH 3: Building SQL from VISUAL QUERY for datasource %s", datasource.Alias))
			// Build SQL from visual query
			sqlQuery, err = s.buildVisualQuery(datasource)
			if err != nil {
				return result, fmt.Errorf("failed to build visual query: %w", err)
			}

			// Apply parameter substitution
			sqlQuery = s.substituteParameters(sqlQuery, parameters)
			result.SQL = sqlQuery
		}

	default:
		return result, fmt.Errorf("unsupported query type: %s", datasource.QueryType)
	}

	s.iLog.Info(fmt.Sprintf(">>> Generated SQL for datasource %s: %s", datasource.Alias, sqlQuery))

	s.iLog.Debug(fmt.Sprintf("Executing SQL for datasource %s on database '%s': %s",
		datasource.Alias, datasource.DatabaseAlias, sqlQuery))

	// Execute the query using SchemaMetadataService
	queryResult, err := s.schemaMetadataService.ExecuteCustomSQL(
		ctx,
		datasource.DatabaseAlias,
		sqlQuery,
	)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Query execution failed for datasource %s on database '%s': %v",
			datasource.Alias, datasource.DatabaseAlias, err))
		return result, fmt.Errorf("query execution failed: %w", err)
	}

	s.iLog.Debug(fmt.Sprintf("Query executed successfully, parsing results..."))
	s.iLog.Debug(fmt.Sprintf("Query result keys: %v", getKeys(queryResult)))

	// Parse the result - SchemaMetadataService returns data at root level
	// Expected structure: {"columns": [...], "rows": [...], "count": N, "query": "..."}

	// Extract columns - handle both []interface{} and []map[string]interface{} types
	columnsRaw := queryResult["columns"]
	s.iLog.Debug(fmt.Sprintf("Columns raw type: %T", columnsRaw))

	// Try []interface{} first (most common)
	if columns, ok := columnsRaw.([]interface{}); ok {
		result.Columns = make([]string, len(columns))
		for i, col := range columns {
			// Handle both string columns and metadata objects
			if colStr, ok := col.(string); ok {
				result.Columns[i] = colStr
			} else if colMap, ok := col.(map[string]interface{}); ok {
				// Column metadata format: {"name": "col_name", "type": "VARCHAR"}
				if name, ok := colMap["name"].(string); ok {
					result.Columns[i] = name
				}
			}
		}
		s.iLog.Debug(fmt.Sprintf("Extracted %d columns: %v", len(result.Columns), result.Columns))
	} else if columnMaps, ok := columnsRaw.([]map[string]interface{}); ok {
		// Handle native Go slice of maps
		result.Columns = make([]string, len(columnMaps))
		for i, colMap := range columnMaps {
			if name, ok := colMap["name"].(string); ok {
				result.Columns[i] = name
			}
		}
		s.iLog.Debug(fmt.Sprintf("Extracted %d columns (from []map): %v", len(result.Columns), result.Columns))
	} else {
		s.iLog.Warn(fmt.Sprintf("Could not extract columns, unexpected type: %T", columnsRaw))
	}

	// Extract rows - handle both []interface{} and []map[string]interface{} types
	rowsRaw := queryResult["rows"]
	s.iLog.Debug(fmt.Sprintf("Rows raw type: %T", rowsRaw))

	// Try []interface{} first (most common)
	if rows, ok := rowsRaw.([]interface{}); ok {
		rowData := make([]map[string]interface{}, len(rows))
		for i, row := range rows {
			if rowMap, ok := row.(map[string]interface{}); ok {
				rowData[i] = rowMap
			}
		}
		result.Rows = rowData      // Primary field for frontend
		result.Data = rowData       // Backward compatibility
		result.TotalRows = len(rows)
		s.iLog.Info(fmt.Sprintf("Successfully extracted %d rows from query result", len(rows)))
		if len(rows) > 0 {
			s.iLog.Debug(fmt.Sprintf("Sample row (first): %+v", rows[0]))
		}
	} else if rowMaps, ok := rowsRaw.([]map[string]interface{}); ok {
		// Handle native Go slice of maps (direct assignment)
		result.Rows = rowMaps      // Primary field for frontend
		result.Data = rowMaps       // Backward compatibility
		result.TotalRows = len(rowMaps)
		s.iLog.Info(fmt.Sprintf("Successfully extracted %d rows (from []map) from query result", len(rowMaps)))
		if len(rowMaps) > 0 {
			s.iLog.Debug(fmt.Sprintf("Sample row (first): %+v", rowMaps[0]))
		}
	} else {
		s.iLog.Warn(fmt.Sprintf("Could not extract rows, unexpected type: %T", rowsRaw))
	}

	// Extract count if available
	if count, ok := queryResult["count"].(int); ok {
		if result.TotalRows == 0 {
			result.TotalRows = count
		}
	}

	if result.ExecutionTimeMs == 0 {
		result.ExecutionTimeMs = time.Since(startTime).Milliseconds()
	}

	s.iLog.Debug(fmt.Sprintf("Datasource %s executed successfully: %d rows in %dms",
		datasource.Alias, result.TotalRows, result.ExecutionTimeMs))

	return result, nil
}

// substituteParameters replaces parameter placeholders in SQL with actual values
// Supports formats: {{paramName}}, :paramName, @paramName
func (s *ReportExecutionService) substituteParameters(sql string, parameters map[string]interface{}) string {
	if parameters == nil || len(parameters) == 0 {
		return sql
	}

	result := sql

	for paramName, paramValue := range parameters {
		// Convert parameter value to string
		valueStr := fmt.Sprintf("%v", paramValue)

		// Handle different parameter formats
		placeholders := []string{
			fmt.Sprintf("{{%s}}", paramName),  // Mustache style
			fmt.Sprintf(":%s", paramName),      // Colon style
			fmt.Sprintf("@%s", paramName),      // At sign style
			fmt.Sprintf("${%s}", paramName),    // Dollar brace style
		}

		for _, placeholder := range placeholders {
			// Quote string values
			replacement := valueStr
			if _, ok := paramValue.(string); ok {
				replacement = fmt.Sprintf("'%s'", strings.ReplaceAll(valueStr, "'", "''"))
			}

			result = strings.ReplaceAll(result, placeholder, replacement)
		}
	}

	return result
}

// getKeys returns the keys from a map[string]interface{} for debugging
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// buildVisualQuery builds SQL from visual query components
func (s *ReportExecutionService) buildVisualQuery(datasource *schema.ReportDatasourceDoc) (string, error) {
	s.iLog.Info(fmt.Sprintf("========================================"))
	s.iLog.Info(fmt.Sprintf(">>> ENTERING buildVisualQuery() for datasource: %s", datasource.Alias))
	s.iLog.Info(fmt.Sprintf("========================================"))

	// Debug log the datasource structure
	s.iLog.Debug(fmt.Sprintf("Datasource SelectedTables type: %T, value: %v", datasource.SelectedTables, datasource.SelectedTables))
	s.iLog.Debug(fmt.Sprintf("Datasource SelectedFields type: %T, value: %v", datasource.SelectedFields, datasource.SelectedFields))
	s.iLog.Debug(fmt.Sprintf("Datasource Joins type: %T, value: %v", datasource.Joins, datasource.Joins))

	// Parse visual query components
	var tables []map[string]string // Changed to support aliases: [{name: "menus", alias: "m1"}, {name: "menus", alias: "m2"}]
	var fields []map[string]interface{}
	var joins []map[string]interface{}
	var filters []map[string]interface{}
	var sorting []map[string]interface{}
	var grouping []string

	// Extract tables with support for aliases
	if datasource.SelectedTables != nil {
		// Convert primitive.A to []interface{}
		var tablesSlice []interface{}

		// Try primitive.A first (BSON array from MongoDB)
		if primA, ok := datasource.SelectedTables.(primitive.A); ok {
			tablesSlice = []interface{}(primA)
			s.iLog.Debug(fmt.Sprintf("Converted primitive.A to []interface{}, length: %d", len(tablesSlice)))
		} else if sliceIface, ok := datasource.SelectedTables.([]interface{}); ok {
			tablesSlice = sliceIface
			s.iLog.Debug(fmt.Sprintf("Using []interface{} directly, length: %d", len(tablesSlice)))
		} else {
			s.iLog.Warn(fmt.Sprintf("SelectedTables is unexpected type: %T", datasource.SelectedTables))
		}

		for _, t := range tablesSlice {
			// Handle nested array from MongoDB: [[{name: "menus"}]]
			if nestedArray, ok := t.([]interface{}); ok {
				for _, nested := range nestedArray {
					// Try primitive.D first
					if primDoc, ok := nested.(primitive.D); ok {
						tableMap := primDoc.Map()
						tableName, _ := tableMap["name"].(string)
						tableAlias, _ := tableMap["alias"].(string)
						if tableName != "" {
							tables = append(tables, map[string]string{"name": tableName, "alias": tableAlias})
							s.iLog.Debug(fmt.Sprintf("Extracted table: name='%s', alias='%s'", tableName, tableAlias))
						}
					} else if tableMap, ok := nested.(map[string]interface{}); ok {
						tableName, _ := tableMap["name"].(string)
						tableAlias, _ := tableMap["alias"].(string)
						if tableName != "" {
							tables = append(tables, map[string]string{"name": tableName, "alias": tableAlias})
							s.iLog.Debug(fmt.Sprintf("Extracted table: name='%s', alias='%s'", tableName, tableAlias))
						}
					}
				}
			} else if primDoc, ok := t.(primitive.D); ok {
				// Direct primitive.D document
				tableMap := primDoc.Map()
				tableName, _ := tableMap["name"].(string)
				tableAlias, _ := tableMap["alias"].(string)
				if tableName != "" {
					tables = append(tables, map[string]string{"name": tableName, "alias": tableAlias})
					s.iLog.Debug(fmt.Sprintf("Extracted table: name='%s', alias='%s'", tableName, tableAlias))
				}
			} else if tableStr, ok := t.(string); ok {
				// Simple string table name - no alias
				tables = append(tables, map[string]string{"name": tableStr, "alias": ""})
				s.iLog.Debug(fmt.Sprintf("Extracted table: name='%s', alias=''", tableStr))
			} else if tableMap, ok := t.(map[string]interface{}); ok {
				// Direct map format: {name: "tablename", alias: "t1"}
				tableName, _ := tableMap["name"].(string)
				tableAlias, _ := tableMap["alias"].(string)
				if tableName != "" {
					tables = append(tables, map[string]string{"name": tableName, "alias": tableAlias})
					s.iLog.Debug(fmt.Sprintf("Extracted table: name='%s', alias='%s'", tableName, tableAlias))
				}
			} else {
				s.iLog.Warn(fmt.Sprintf("Unknown table type: %T, value: %v", t, t))
			}
		}
	}

	// Extract fields
	if datasource.SelectedFields != nil {
		var fieldsSlice []interface{}

		// Try primitive.A first (BSON array from MongoDB)
		if primA, ok := datasource.SelectedFields.(primitive.A); ok {
			fieldsSlice = []interface{}(primA)
		} else if sliceIface, ok := datasource.SelectedFields.([]interface{}); ok {
			fieldsSlice = sliceIface
		}

		if len(fieldsSlice) > 0 {
			for _, f := range fieldsSlice {
				// Handle nested array from MongoDB: [[{table: "menus", field: "id"}]]
				if nestedArray, ok := f.([]interface{}); ok {
					for _, nested := range nestedArray {
						if primDoc, ok := nested.(primitive.D); ok {
							fields = append(fields, primDoc.Map())
						} else if fieldMap, ok := nested.(map[string]interface{}); ok {
							fields = append(fields, fieldMap)
						}
					}
				} else if primDoc, ok := f.(primitive.D); ok {
					fields = append(fields, primDoc.Map())
				} else if fieldMap, ok := f.(map[string]interface{}); ok {
					// Direct map format
					fields = append(fields, fieldMap)
				}
			}
		}
	}

	// Extract joins
	if datasource.Joins != nil {
		var joinsSlice []interface{}

		// Try primitive.A first (BSON array from MongoDB)
		if primA, ok := datasource.Joins.(primitive.A); ok {
			joinsSlice = []interface{}(primA)
		} else if sliceIface, ok := datasource.Joins.([]interface{}); ok {
			joinsSlice = sliceIface
		}

		if len(joinsSlice) > 0 {
			for _, j := range joinsSlice {
				// Handle nested array from MongoDB
				if nestedArray, ok := j.([]interface{}); ok {
					for _, nested := range nestedArray {
						if primDoc, ok := nested.(primitive.D); ok {
							joins = append(joins, primDoc.Map())
						} else if joinMap, ok := nested.(map[string]interface{}); ok {
							joins = append(joins, joinMap)
						}
					}
				} else if primDoc, ok := j.(primitive.D); ok {
					joins = append(joins, primDoc.Map())
				} else if joinMap, ok := j.(map[string]interface{}); ok {
					joins = append(joins, joinMap)
				}
			}
		}
	}

	// Extract filters
	if datasource.Filters != nil {
		var filtersSlice []interface{}

		if primA, ok := datasource.Filters.(primitive.A); ok {
			filtersSlice = []interface{}(primA)
		} else if sliceIface, ok := datasource.Filters.([]interface{}); ok {
			filtersSlice = sliceIface
		}

		if len(filtersSlice) > 0 {
			for _, f := range filtersSlice {
				// Handle nested array from MongoDB
				if nestedArray, ok := f.([]interface{}); ok {
					for _, nested := range nestedArray {
						if primDoc, ok := nested.(primitive.D); ok {
							filters = append(filters, primDoc.Map())
						} else if filterMap, ok := nested.(map[string]interface{}); ok {
							filters = append(filters, filterMap)
						}
					}
				} else if primDoc, ok := f.(primitive.D); ok {
					filters = append(filters, primDoc.Map())
				} else if filterMap, ok := f.(map[string]interface{}); ok {
					filters = append(filters, filterMap)
				}
			}
		}
	}

	// Extract sorting
	if datasource.Sorting != nil {
		var sortingSlice []interface{}

		if primA, ok := datasource.Sorting.(primitive.A); ok {
			sortingSlice = []interface{}(primA)
		} else if sliceIface, ok := datasource.Sorting.([]interface{}); ok {
			sortingSlice = sliceIface
		}

		if len(sortingSlice) > 0 {
			for _, s := range sortingSlice {
				// Handle nested array from MongoDB
				if nestedArray, ok := s.([]interface{}); ok {
					for _, nested := range nestedArray {
						if primDoc, ok := nested.(primitive.D); ok {
							sorting = append(sorting, primDoc.Map())
						} else if sortMap, ok := nested.(map[string]interface{}); ok {
							sorting = append(sorting, sortMap)
						}
					}
				} else if primDoc, ok := s.(primitive.D); ok {
					sorting = append(sorting, primDoc.Map())
				} else if sortMap, ok := s.(map[string]interface{}); ok {
					sorting = append(sorting, sortMap)
				}
			}
		}
	}

	// Extract grouping
	if datasource.Grouping != nil {
		var groupingSlice []interface{}

		if primA, ok := datasource.Grouping.(primitive.A); ok {
			groupingSlice = []interface{}(primA)
		} else if sliceIface, ok := datasource.Grouping.([]interface{}); ok {
			groupingSlice = sliceIface
		}

		if len(groupingSlice) > 0 {
			for _, g := range groupingSlice {
				// Handle nested array from MongoDB
				if nestedArray, ok := g.([]interface{}); ok {
					for _, nested := range nestedArray {
						if groupStr, ok := nested.(string); ok {
							grouping = append(grouping, groupStr)
						} else if groupMap, ok := nested.(map[string]interface{}); ok {
							if field, ok := groupMap["field"].(string); ok {
								grouping = append(grouping, field)
							}
						}
					}
				} else if groupStr, ok := g.(string); ok {
					grouping = append(grouping, groupStr)
				} else if groupMap, ok := g.(map[string]interface{}); ok {
					if field, ok := groupMap["field"].(string); ok {
						grouping = append(grouping, field)
					}
				}
			}
		}
	}

	// Log extraction results BEFORE validation
	s.iLog.Info(fmt.Sprintf(">>> EXTRACTION COMPLETE: Tables=%d, Fields=%d, Joins=%d", len(tables), len(fields), len(joins)))
	if len(tables) > 0 {
		s.iLog.Info(fmt.Sprintf(">>> Tables extracted: %v", tables))
	}
	if len(joins) > 0 {
		s.iLog.Info(fmt.Sprintf(">>> Joins extracted: %v", joins))
	}

	// Validate minimum requirements
	if len(tables) == 0 {
		s.iLog.Error(">>> ERROR: No tables extracted!")
		return "", fmt.Errorf("no tables selected in visual query")
	}
	if len(fields) == 0 {
		s.iLog.Error(">>> ERROR: No fields extracted!")
		return "", fmt.Errorf("no fields selected in visual query")
	}

	s.iLog.Debug(fmt.Sprintf("Visual query components - Tables: %d, Fields: %d, Joins: %d", len(tables), len(fields), len(joins)))
	s.iLog.Debug(fmt.Sprintf("Extracted tables: %v", tables))
	s.iLog.Debug(fmt.Sprintf("Extracted fields: %v", fields))
	s.iLog.Debug(fmt.Sprintf("Extracted joins: %v", joins))

	// Build SELECT clause
	var selectFields []string
	var hasAggregation bool
	var nonAggregatedFields []string

	for _, field := range fields {
		table, _ := field["table"].(string)

		// Support both "column" and "field" keys
		column, _ := field["column"].(string)
		if column == "" {
			column, _ = field["field"].(string)
		}

		alias, _ := field["alias"].(string)
		aggregation, _ := field["aggregation"].(string)

		if table == "" || column == "" {
			s.iLog.Debug(fmt.Sprintf("Skipping field - table: '%s', column: '%s'", table, column))
			continue
		}

		fieldExpr := fmt.Sprintf("%s.%s", table, column)

		if aggregation != "" {
			hasAggregation = true
			fieldExpr = fmt.Sprintf("%s(%s)", aggregation, fieldExpr)
		} else {
			nonAggregatedFields = append(nonAggregatedFields, fmt.Sprintf("%s.%s", table, column))
		}

		if alias != "" {
			fieldExpr = fmt.Sprintf("%s AS %s", fieldExpr, alias)
		}

		selectFields = append(selectFields, fieldExpr)
	}

	if len(selectFields) == 0 {
		return "", fmt.Errorf("no valid fields to select")
	}

	sql := fmt.Sprintf("SELECT %s", strings.Join(selectFields, ", "))

	// Build FROM clause with alias support
	if len(tables) == 0 {
		return "", fmt.Errorf("no tables specified for query")
	}

	// Use first table as FROM clause
	baseTableName := tables[0]["name"]
	baseTableAlias := tables[0]["alias"]

	if baseTableAlias != "" {
		sql += fmt.Sprintf("\nFROM %s AS %s", baseTableName, baseTableAlias)
	} else {
		sql += fmt.Sprintf("\nFROM %s", baseTableName)
	}

	// Track which tables are already in the query
	tablesInQuery := make(map[string]bool)
	tablesInQuery[baseTableName] = true
	if baseTableAlias != "" {
		tablesInQuery[baseTableAlias] = true
	}

	// Build JOIN clauses with alias support
	for i, join := range joins {
		// Support both camelCase and underscore keys from MongoDB
		leftTable, _ := join["leftTable"].(string)
		if leftTable == "" {
			leftTable, _ = join["left_table"].(string)
		}

		leftColumn, _ := join["leftColumn"].(string)
		if leftColumn == "" {
			leftColumn, _ = join["left_field"].(string)
		}

		rightTable, _ := join["rightTable"].(string)
		if rightTable == "" {
			rightTable, _ = join["right_table"].(string)
		}

		rightColumn, _ := join["rightColumn"].(string)
		if rightColumn == "" {
			rightColumn, _ = join["right_field"].(string)
		}

		joinType, _ := join["joinType"].(string)
		if joinType == "" {
			joinType, _ = join["join_type"].(string)
		}

		s.iLog.Debug(fmt.Sprintf("Processing JOIN[%d]: left='%s.%s', right='%s.%s', type='%s'",
			i, leftTable, leftColumn, rightTable, rightColumn, joinType))

		if leftTable == "" || leftColumn == "" || rightTable == "" || rightColumn == "" {
			s.iLog.Warn(fmt.Sprintf("Skipping incomplete JOIN[%d]: missing required fields", i))
			continue
		}

		if joinType == "" {
			joinType = "INNER"
		}

		// Determine which table to join (the one not already in the query)
		var joinTable string
		var joinTableAlias string
		var onClause string

		// Check if tables are already in query
		isLeftInQuery := tablesInQuery[leftTable]
		isRightInQuery := tablesInQuery[rightTable]

		s.iLog.Debug(fmt.Sprintf("  isLeftInQuery(%s)=%v, isRightInQuery(%s)=%v", leftTable, isLeftInQuery, rightTable, isRightInQuery))

		if !isLeftInQuery && isRightInQuery {
			// Join the leftTable (rightTable already in query)
			joinTable = leftTable
			// Find alias for leftTable
			for _, t := range tables {
				if t["name"] == leftTable && t["alias"] != "" {
					joinTableAlias = t["alias"]
					break
				}
			}
			onClause = fmt.Sprintf("%s.%s = %s.%s", leftTable, leftColumn, rightTable, rightColumn)
			s.iLog.Debug(fmt.Sprintf("  Decision: JOIN leftTable '%s'", joinTable))
		} else if isLeftInQuery && !isRightInQuery {
			// Join the rightTable (leftTable already in query)
			joinTable = rightTable
			// Find alias for rightTable
			for _, t := range tables {
				if t["name"] == rightTable && t["alias"] != "" {
					joinTableAlias = t["alias"]
					break
				}
			}
			onClause = fmt.Sprintf("%s.%s = %s.%s", leftTable, leftColumn, rightTable, rightColumn)
			s.iLog.Debug(fmt.Sprintf("  Decision: JOIN rightTable '%s'", joinTable))
		} else if !isLeftInQuery && !isRightInQuery {
			// Neither table in query - this shouldn't happen in well-formed joins
			// Default to joining rightTable
			joinTable = rightTable
			for _, t := range tables {
				if t["name"] == rightTable && t["alias"] != "" {
					joinTableAlias = t["alias"]
					break
				}
			}
			onClause = fmt.Sprintf("%s.%s = %s.%s", leftTable, leftColumn, rightTable, rightColumn)
			s.iLog.Warn(fmt.Sprintf("  Warning: Neither table in query, defaulting to JOIN rightTable '%s'", joinTable))
		} else {
			// Both tables already in query - this is a filter condition, not a join
			s.iLog.Warn(fmt.Sprintf("  Warning: Both tables already in query - skipping JOIN"))
			continue
		}

		// Build JOIN SQL
		if joinTableAlias != "" {
			sql += fmt.Sprintf("\n%s JOIN %s AS %s ON %s", joinType, joinTable, joinTableAlias, onClause)
		} else {
			sql += fmt.Sprintf("\n%s JOIN %s ON %s", joinType, joinTable, onClause)
		}

		// Add joined table to tracking map
		tablesInQuery[joinTable] = true
		if joinTableAlias != "" {
			tablesInQuery[joinTableAlias] = true
		}

		s.iLog.Debug(fmt.Sprintf("  Added to query: table='%s', alias='%s'", joinTable, joinTableAlias))
	}

	// Build WHERE clause
	if len(filters) > 0 {
		var whereConditions []string
		for _, filter := range filters {
			field, _ := filter["field"].(string)
			operator, _ := filter["operator"].(string)
			value := filter["value"]

			if field == "" || operator == "" {
				continue
			}

			var condition string
			if operator == "IS NULL" || operator == "IS NOT NULL" {
				condition = fmt.Sprintf("%s %s", field, operator)
			} else if operator == "IN" || operator == "NOT IN" {
				// Handle array values
				condition = fmt.Sprintf("%s %s (%v)", field, operator, value)
			} else {
				// Handle single value
				if valueStr, ok := value.(string); ok {
					condition = fmt.Sprintf("%s %s '%s'", field, operator, valueStr)
				} else {
					condition = fmt.Sprintf("%s %s %v", field, operator, value)
				}
			}

			whereConditions = append(whereConditions, condition)
		}

		if len(whereConditions) > 0 {
			sql += fmt.Sprintf("\nWHERE %s", strings.Join(whereConditions, " AND "))
		}
	}

	// Build GROUP BY clause
	// If we have aggregations, we need to group by non-aggregated fields
	if hasAggregation {
		if len(grouping) > 0 {
			sql += fmt.Sprintf("\nGROUP BY %s", strings.Join(grouping, ", "))
		} else if len(nonAggregatedFields) > 0 {
			// Auto-group by non-aggregated fields
			sql += fmt.Sprintf("\nGROUP BY %s", strings.Join(nonAggregatedFields, ", "))
			s.iLog.Debug(fmt.Sprintf("Auto-added GROUP BY for aggregation: %s", strings.Join(nonAggregatedFields, ", ")))
		}
	} else if len(grouping) > 0 {
		sql += fmt.Sprintf("\nGROUP BY %s", strings.Join(grouping, ", "))
	}

	// Build ORDER BY clause
	if len(sorting) > 0 {
		var orderBy []string
		for _, sort := range sorting {
			field, _ := sort["field"].(string)
			direction, _ := sort["direction"].(string)

			if field == "" {
				continue
			}

			if direction == "" {
				direction = "ASC"
			}

			orderBy = append(orderBy, fmt.Sprintf("%s %s", field, direction))
		}

		if len(orderBy) > 0 {
			sql += fmt.Sprintf("\nORDER BY %s", strings.Join(orderBy, ", "))
		}
	}

	// Add LIMIT if specified
	if datasource.Parameters != nil {
		if paramsMap, ok := datasource.Parameters.(map[string]interface{}); ok {
			if limit, ok := paramsMap["limit"].(float64); ok && limit > 0 {
				sql += fmt.Sprintf("\nLIMIT %d", int(limit))
			}
		}
	}

	s.iLog.Info(fmt.Sprintf("========================================"))
	s.iLog.Info(fmt.Sprintf(">>> EXITING buildVisualQuery() - Generated SQL:"))
	s.iLog.Info(fmt.Sprintf("%s", sql))
	s.iLog.Info(fmt.Sprintf("========================================"))

	return sql, nil
}
