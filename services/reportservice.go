package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/databases/orm"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// ReportService handles report business logic
type ReportService struct {
	DB *gorm.DB
}

// NewReportService creates a new report service
func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{DB: db}
}

// CreateReport creates a new report
func (s *ReportService) CreateReport(report *models.Report) error {
	// Generate UUID if not provided
	if report.ID == "" {
		report.ID = uuid.New().String()
	}

	// Set defaults
	if report.Version == 0 {
		report.Version = 1
	}

	return s.DB.Create(report).Error
}

// GetReportByID retrieves a report by ID with all relationships
func (s *ReportService) GetReportByID(id string) (*models.Report, error) {
	var report models.Report
	err := s.DB.Preload("Datasources").
		Preload("Components").
		Preload("Parameters").
		Preload("Executions").
		Preload("Shares").
		First(&report, "id = ?", id).Error

	if err != nil {
		return nil, err
	}

	return &report, nil
}

// ListReports retrieves reports with pagination and filtering
func (s *ReportService) ListReports(userID string, isPublic bool, reportType string, page, pageSize int) ([]models.Report, int64, error) {
	var reports []models.Report
	var total int64

	query := s.DB.Model(&models.Report{})

	// Apply filters
	if !isPublic {
		query = query.Where("createdby = ? OR ispublic = ?", userID, true)
	} else {
		query = query.Where("ispublic = ?", true)
	}

	if reportType != "" {
		query = query.Where("reporttype = ?", reportType)
	}

	query = query.Where("active = ?", true)

	// Get total count
	query.Count(&total)

	// Apply pagination
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("modifiedon DESC").
		Find(&reports).Error

	if err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// UpdateReport updates an existing report
func (s *ReportService) UpdateReport(id string, updates map[string]interface{}) error {
	// Filter out relationship fields - these should be updated through their dedicated endpoints
	// to avoid GORM trying to pass complex nested structures as SQL parameters
	relationshipFields := []string{"datasources", "components", "parameters", "executions", "shares"}
	filteredUpdates := make(map[string]interface{})

	for key, value := range updates {
		// Skip relationship fields
		isRelationship := false
		for _, relField := range relationshipFields {
			if key == relField {
				isRelationship = true
				break
			}
		}
		if !isRelationship {
			filteredUpdates[key] = value
		}
	}

	// Add updated_at timestamp
	filteredUpdates["modifiedon"] = time.Now()

	return s.DB.Model(&models.Report{}).Where("id = ?", id).Updates(filteredUpdates).Error
}

// DeleteReport soft deletes a report
func (s *ReportService) DeleteReport(id string) error {
	return s.DB.Model(&models.Report{}).Where("id = ?", id).Update("active", false).Error
}

// HardDeleteReport permanently deletes a report
func (s *ReportService) HardDeleteReport(id string) error {
	return s.DB.Where("id = ?", id).Delete(&models.Report{}).Error
}

// AddDatasource adds a datasource to a report
func (s *ReportService) AddDatasource(datasource *models.ReportDatasource) error {
	if datasource.ID == "" {
		datasource.ID = uuid.New().String()
	}

	return s.DB.Create(datasource).Error
}

// GetDatasources retrieves all datasources for a report
func (s *ReportService) GetDatasources(reportID string) ([]models.ReportDatasource, error) {
	var datasources []models.ReportDatasource
	err := s.DB.Where("reportid = ?", reportID).Find(&datasources).Error
	return datasources, err
}

// UpdateDatasource updates a datasource
func (s *ReportService) UpdateDatasource(id string, updates map[string]interface{}) error {
	// JSON fields that need special handling for serialization
	jsonFields := []string{"selectedtables", "selectedfields", "joins", "filters", "sorting", "grouping", "parameters"}

	processedUpdates := make(map[string]interface{})

	for key, value := range updates {
		// Check if this is a JSON field
		isJSONField := false
		for _, jf := range jsonFields {
			if key == jf {
				isJSONField = true
				break
			}
		}

		if isJSONField && value != nil {
			// Serialize JSON fields to []byte for GORM
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("error serializing %s: %v", key, err)
			}
			processedUpdates[key] = jsonBytes
		} else {
			// Non-JSON fields pass through as-is
			processedUpdates[key] = value
		}
	}

	processedUpdates["modifiedon"] = time.Now()
	return s.DB.Model(&models.ReportDatasource{}).Where("id = ?", id).Updates(processedUpdates).Error
}

// DeleteDatasource deletes a datasource
func (s *ReportService) DeleteDatasource(id string) error {
	return s.DB.Where("id = ?", id).Delete(&models.ReportDatasource{}).Error
}

// AddComponent adds a component to a report
func (s *ReportService) AddComponent(component *models.ReportComponent) error {
	if component.ID == "" {
		component.ID = uuid.New().String()
	}

	return s.DB.Create(component).Error
}

// GetComponents retrieves all visible components for a report
func (s *ReportService) GetComponents(reportID string) ([]models.ReportComponent, error) {
	var components []models.ReportComponent
	err := s.DB.Where("reportid = ? AND isvisible = ?", reportID, true).
		Order("z_index ASC").
		Find(&components).Error
	return components, err
}

// GetAllComponents retrieves all components for a report (including invisible)
func (s *ReportService) GetAllComponents(reportID string) ([]models.ReportComponent, error) {
	var components []models.ReportComponent
	err := s.DB.Where("reportid = ?", reportID).
		Order("z_index ASC").
		Find(&components).Error
	return components, err
}

// UpdateComponent updates a component
func (s *ReportService) UpdateComponent(id string, updates map[string]interface{}) error {
	// List of JSON fields that need serialization
	jsonFields := []string{"dataconfig", "componentconfig", "styleconfig", "chartconfig", "barcodeconfig", "drilldownconfig", "conditionalformatting"}

	// Create a new map with serialized JSON fields
	processedUpdates := make(map[string]interface{})

	for key, value := range updates {
		// Check if this is a JSON field
		isJSONField := false
		for _, jf := range jsonFields {
			if key == jf {
				isJSONField = true
				break
			}
		}

		// If it's a JSON field and not nil, serialize it to []byte
		if isJSONField && value != nil {
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("error serializing %s: %v", key, err)
			}
			processedUpdates[key] = jsonBytes
		} else {
			// For non-JSON fields, pass through as-is
			processedUpdates[key] = value
		}
	}

	processedUpdates["modifiedon"] = time.Now()
	return s.DB.Model(&models.ReportComponent{}).Where("id = ?", id).Updates(processedUpdates).Error
}

// DeleteComponent deletes a component
func (s *ReportService) DeleteComponent(id string) error {
	return s.DB.Where("id = ?", id).Delete(&models.ReportComponent{}).Error
}

// AddParameter adds a parameter to a report
func (s *ReportService) AddParameter(parameter *models.ReportParameter) error {
	if parameter.ID == "" {
		parameter.ID = uuid.New().String()
	}

	return s.DB.Create(parameter).Error
}

// GetParameters retrieves all parameters for a report
func (s *ReportService) GetParameters(reportID string) ([]models.ReportParameter, error) {
	var parameters []models.ReportParameter
	err := s.DB.Where("reportid = ? AND isenabled = ?", reportID, true).
		Order("sort_order ASC").
		Find(&parameters).Error
	return parameters, err
}

// UpdateParameter updates a parameter
func (s *ReportService) UpdateParameter(id string, updates map[string]interface{}) error {
	updates["modifiedon"] = time.Now()
	return s.DB.Model(&models.ReportParameter{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteParameter deletes a parameter
func (s *ReportService) DeleteParameter(id string) error {
	return s.DB.Where("id = ?", id).Delete(&models.ReportParameter{}).Error
}

// CreateExecution creates a new execution record
func (s *ReportService) CreateExecution(execution *models.ReportExecution) error {
	if execution.ID == "" {
		execution.ID = uuid.New().String()
	}

	return s.DB.Create(execution).Error
}

// UpdateExecution updates an execution record
func (s *ReportService) UpdateExecution(id string, updates map[string]interface{}) error {
	return s.DB.Model(&models.ReportExecution{}).Where("id = ?", id).Updates(updates).Error
}

// GetExecutionHistory retrieves execution history for a report
func (s *ReportService) GetExecutionHistory(reportID string, limit int) ([]models.ReportExecution, error) {
	var executions []models.ReportExecution
	err := s.DB.Where("reportid = ?", reportID).
		Order("createdob DESC").
		Limit(limit).
		Find(&executions).Error
	return executions, err
}

// UpdateLastExecutedAt updates the last execution timestamp
func (s *ReportService) UpdateLastExecutedAt(reportID string) error {
	now := time.Now()
	return s.DB.Model(&models.Report{}).Where("id = ?", reportID).Update("lastexecutedon", now).Error
}

// ShareReport creates a share record
func (s *ReportService) ShareReport(share *models.ReportShare) error {
	if share.ID == "" {
		share.ID = uuid.New().String()
	}

	// Generate share token if not provided
	if share.ShareToken == "" {
		share.ShareToken = uuid.New().String()
	}

	return s.DB.Create(share).Error
}

// GetShares retrieves all shares for a report
func (s *ReportService) GetShares(reportID string) ([]models.ReportShare, error) {
	var shares []models.ReportShare
	err := s.DB.Where("reportid = ? AND active = ?", reportID, true).Find(&shares).Error
	return shares, err
}

// RevokeShare deactivates a share
func (s *ReportService) RevokeShare(id string) error {
	return s.DB.Model(&models.ReportShare{}).Where("id = ?", id).Update("active", false).Error
}

// GetShareByToken retrieves a share by token
func (s *ReportService) GetShareByToken(token string) (*models.ReportShare, error) {
	var share models.ReportShare
	err := s.DB.Where("share_token = ? AND active = ?", token, true).First(&share).Error
	if err != nil {
		return nil, err
	}

	// Check expiration
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("share link has expired")
	}

	return &share, nil
}

// ListTemplates retrieves report templates
func (s *ReportService) ListTemplates(category string, isPublic bool) ([]models.ReportTemplate, error) {
	var templates []models.ReportTemplate
	query := s.DB.Model(&models.ReportTemplate{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if isPublic {
		query = query.Where("ispublic = ?", true)
	}

	err := query.Order("usage_count DESC, rating DESC").Find(&templates).Error
	return templates, err
}

// GetTemplateByID retrieves a template by ID
func (s *ReportService) GetTemplateByID(id string) (*models.ReportTemplate, error) {
	var template models.ReportTemplate
	err := s.DB.First(&template, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return &template, nil
}

// CreateFromTemplate creates a report from a template
func (s *ReportService) CreateFromTemplate(templateID, userID string) (*models.Report, error) {
	// Get template
	template, err := s.GetTemplateByID(templateID)
	if err != nil {
		return nil, err
	}

	// Create report from template
	report := &models.Report{
		ID:               uuid.New().String(),
		Name:             template.Name + " (Copy)",
		Description:      template.Description,
		ReportType:       models.ReportTypeTemplate,
		CreatedBy:        userID,
		IsPublic:         false,
		TemplateSourceID: templateID,
		LayoutConfig:     template.TemplateConfig,
		Version:          1,
		Active:           true,
	}

	err = s.CreateReport(report)
	if err != nil {
		return nil, err
	}

	// Increment template usage count
	s.DB.Model(&models.ReportTemplate{}).Where("id = ?", templateID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1"))

	return report, nil
}

// DuplicateReport creates a copy of an existing report
func (s *ReportService) DuplicateReport(reportID, userID string) (*models.Report, error) {
	// Get original report with all relationships
	original, err := s.GetReportByID(reportID)
	if err != nil {
		return nil, err
	}

	// Create new report
	newReport := &models.Report{
		ID:           uuid.New().String(),
		Name:         original.Name + " (Copy)",
		Description:  original.Description,
		ReportType:   original.ReportType,
		CreatedBy:    userID,
		IsPublic:     false,
		LayoutConfig: original.LayoutConfig,
		PageSettings: original.PageSettings,
		Version:      1,
		Active:       true,
	}

	err = s.CreateReport(newReport)
	if err != nil {
		return nil, err
	}

	// Copy datasources
	for _, ds := range original.Datasources {
		newDS := ds
		newDS.ID = uuid.New().String()
		newDS.ReportID = newReport.ID
		s.AddDatasource(&newDS)
	}

	// Copy components
	for _, comp := range original.Components {
		newComp := comp
		newComp.ID = uuid.New().String()
		newComp.ReportID = newReport.ID
		s.AddComponent(&newComp)
	}

	// Copy parameters
	for _, param := range original.Parameters {
		newParam := param
		newParam.ID = uuid.New().String()
		newParam.ReportID = newReport.ID
		s.AddParameter(&newParam)
	}

	return newReport, nil
}

// SearchReports searches reports by name or description
func (s *ReportService) SearchReports(keyword string, userID string, limit int) ([]models.Report, error) {
	var reports []models.Report
	searchTerm := "%" + keyword + "%"

	err := s.DB.Where("(createdby = ? OR ispublic = ?) AND active = ? AND (name LIKE ? OR description LIKE ?)",
		userID, true, true, searchTerm, searchTerm).
		Limit(limit).
		Order("modifiedon DESC").
		Find(&reports).Error

	return reports, err
}

// ExecuteReportQuery executes all datasource queries for a report
func (s *ReportService) ExecuteReportQuery(reportID string, parameters map[string]interface{}) (map[string]interface{}, error) {
	// Get report with datasources
	report, err := s.GetReportByID(reportID)
	if err != nil {
		return nil, fmt.Errorf("error fetching report: %v", err)
	}

	// Get datasources
	datasources, err := s.GetDatasources(reportID)
	if err != nil {
		return nil, fmt.Errorf("error fetching datasources: %v", err)
	}

	if len(datasources) == 0 {
		return nil, fmt.Errorf("report has no datasources configured")
	}

	// Execute each datasource and collect results
	datasourceResults := make(map[string]interface{})

	for _, ds := range datasources {
		var sqlQuery string
		var queryParams []interface{}

		// Check if datasource has custom SQL or uses visual query builder
		if ds.CustomSQL != "" && strings.TrimSpace(ds.CustomSQL) != "" {
			// Use custom SQL
			sqlQuery = ds.CustomSQL

			// Replace parameter placeholders (@paramName) with values
			for paramName, paramValue := range parameters {
				placeholder := fmt.Sprintf("@%s", paramName)
				valueStr := fmt.Sprintf("%v", paramValue)
				sqlQuery = strings.ReplaceAll(sqlQuery, placeholder, valueStr)
			}
		} else {
			// Visual query builder - build SQL from fields
			sqlQuery, queryParams, err = s.buildSQLFromVisualQuery(&ds, parameters)
			if err != nil {
				return nil, fmt.Errorf("error building SQL for datasource '%s': %v", ds.Alias, err)
			}
		}

		// Execute the query
		result, err := s.executeDatasourceQuery(ds.DatabaseAlias, sqlQuery, queryParams)
		if err != nil {
			return nil, fmt.Errorf("error executing datasource '%s': %v", ds.Alias, err)
		}

		// Store result with datasource alias as key
		datasourceResults[ds.Alias] = result
	}

	return map[string]interface{}{
		"reportId":    reportID,
		"reportName":  report.Name,
		"datasources": datasourceResults,
		"executedAt":  time.Now(),
	}, nil
}

// executeDatasourceQuery executes a SQL query on the specified database alias
func (s *ReportService) executeDatasourceQuery(databaseAlias string, sqlQuery string, params []interface{}) (map[string]interface{}, error) {
	// Get database connection
	db, err := orm.GetDB(databaseAlias)
	if err != nil {
		return nil, fmt.Errorf("error getting database for alias '%s': %v", databaseAlias, err)
	}

	// Safety check - only allow SELECT queries
	normalizedSQL := strings.ToUpper(strings.TrimSpace(sqlQuery))
	if !strings.HasPrefix(normalizedSQL, "SELECT") {
		return nil, fmt.Errorf("only SELECT queries are allowed")
	}

	// Check for dangerous operations
	dangerousOps := []string{"DROP", "DELETE", "TRUNCATE", "INSERT", "UPDATE", "ALTER", "CREATE", "EXEC", "EXECUTE"}
	for _, op := range dangerousOps {
		if strings.Contains(normalizedSQL, op) {
			return nil, fmt.Errorf("query contains forbidden operation: %s", op)
		}
	}

	// Execute query
	rows, err := db.Query(sqlQuery, params...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns: %v", err)
	}

	// Get column types
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("error getting column types: %v", err)
	}

	// Build fields metadata
	fields := make([]map[string]interface{}, 0, len(columns))
	for i, col := range columns {
		fieldInfo := map[string]interface{}{
			"name":     col,
			"dataType": columnTypes[i].DatabaseTypeName(),
		}
		fields = append(fields, fieldInfo)
	}

	// Read all rows
	var resultRows []map[string]interface{}
	for rows.Next() {
		// Create slice to hold column values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan row
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		// Create map for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Convert []byte to string for better JSON serialization
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		resultRows = append(resultRows, rowMap)
	}

	return map[string]interface{}{
		"fields":    fields,
		"rows":      resultRows,
		"totalRows": len(resultRows),
	}, nil
}

// buildSQLFromVisualQuery builds SQL from visual query builder fields
func (s *ReportService) buildSQLFromVisualQuery(ds *models.ReportDatasource, parameters map[string]interface{}) (string, []interface{}, error) {
	var sqlParts []string
	var queryParams []interface{}

	// 1. Build SELECT clause
	selectClause := "SELECT "
	if ds.SelectedFields.Data != nil {
		fields, err := s.extractSelectedFields(ds.SelectedFields.Data)
		if err != nil {
			return "", nil, fmt.Errorf("error parsing selected fields: %v", err)
		}
		if len(fields) == 0 {
			selectClause += "*"
		} else {
			selectClause += strings.Join(fields, ", ")
		}
	} else {
		selectClause += "*"
	}
	sqlParts = append(sqlParts, selectClause)

	// 2. Build FROM clause
	if ds.SelectedTables.Data == nil {
		return "", nil, fmt.Errorf("no tables selected in datasource '%s'", ds.Alias)
	}

	tables, err := s.extractSelectedTables(ds.SelectedTables.Data)
	if err != nil {
		return "", nil, fmt.Errorf("error parsing selected tables: %v", err)
	}
	if len(tables) == 0 {
		return "", nil, fmt.Errorf("no tables selected in datasource '%s'", ds.Alias)
	}

	fromClause := "FROM " + tables[0]
	sqlParts = append(sqlParts, fromClause)

	// 3. Build JOIN clauses
	if ds.Joins.Data != nil {
		joinClauses, err := s.extractJoins(ds.Joins.Data)
		if err != nil {
			return "", nil, fmt.Errorf("error parsing joins: %v", err)
		}
		for _, joinClause := range joinClauses {
			sqlParts = append(sqlParts, joinClause)
		}
	}

	// 4. Build WHERE clause from filters
	if ds.Filters.Data != nil {
		whereClause, filterParams, err := s.extractFilters(ds.Filters.Data, parameters)
		if err != nil {
			return "", nil, fmt.Errorf("error parsing filters: %v", err)
		}
		if whereClause != "" {
			sqlParts = append(sqlParts, "WHERE "+whereClause)
			queryParams = append(queryParams, filterParams...)
		}
	}

	// 5. Build GROUP BY clause
	if ds.Grouping.Data != nil {
		groupByClause, err := s.extractGrouping(ds.Grouping.Data)
		if err != nil {
			return "", nil, fmt.Errorf("error parsing grouping: %v", err)
		}
		if groupByClause != "" {
			sqlParts = append(sqlParts, "GROUP BY "+groupByClause)
		}
	}

	// 6. Build ORDER BY clause
	if ds.Sorting.Data != nil {
		orderByClause, err := s.extractSorting(ds.Sorting.Data)
		if err != nil {
			return "", nil, fmt.Errorf("error parsing sorting: %v", err)
		}
		if orderByClause != "" {
			sqlParts = append(sqlParts, "ORDER BY "+orderByClause)
		}
	}

	sql := strings.Join(sqlParts, " ")
	return sql, queryParams, nil
}

// quoteColumnIdentifier quotes a SQL column identifier with backticks if needed
func (s *ReportService) quoteColumnIdentifier(identifier string) string {
	identifier = strings.TrimSpace(identifier)

	// Don't quote functions or expressions (contain parentheses)
	if strings.Contains(identifier, "(") && strings.Contains(identifier, ")") {
		return identifier
	}

	// If identifier contains spaces or special chars, quote it
	if strings.Contains(identifier, " ") || strings.Contains(identifier, "-") {
		return fmt.Sprintf("`%s`", identifier)
	}

	return identifier
}

// quoteAlias quotes a SQL alias with double quotes if it contains spaces
func (s *ReportService) quoteAlias(alias string) string {
	alias = strings.TrimSpace(alias)

	// If alias contains spaces, quote with double quotes (SQL standard)
	if strings.Contains(alias, " ") {
		return fmt.Sprintf("\"%s\"", alias)
	}

	return alias
}

// extractSelectedFields extracts field names from selectedfields JSON
func (s *ReportService) extractSelectedFields(data interface{}) ([]string, error) {
	var fields []string

	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			if fieldStr, ok := item.(string); ok {
				// Clean and validate field name
				fieldStr = strings.TrimSpace(fieldStr)
				if fieldStr != "" {
					// Quote column identifier if needed (backticks for columns)
					quotedField := s.quoteColumnIdentifier(fieldStr)
					fields = append(fields, quotedField)
				}
			} else if fieldMap, ok := item.(map[string]interface{}); ok {
				// Handle object format: {field: "COUNT(menu.id)", alias: "Count of Menu"}
				if field, ok := fieldMap["field"].(string); ok {
					// Don't quote the field if it's a function
					quotedField := s.quoteColumnIdentifier(field)
					if alias, hasAlias := fieldMap["alias"].(string); hasAlias && alias != "" {
						// Quote alias with double quotes if it has spaces
						quotedAlias := s.quoteAlias(alias)
						fields = append(fields, fmt.Sprintf("%s AS %s", quotedField, quotedAlias))
					} else {
						fields = append(fields, quotedField)
					}
				}
			}
		}
	case string:
		// Single field as string
		if v != "" {
			quotedField := s.quoteColumnIdentifier(v)
			fields = append(fields, quotedField)
		}
	}

	return fields, nil
}

// extractSelectedTables extracts table names from selectedtables JSON
func (s *ReportService) extractSelectedTables(data interface{}) ([]string, error) {
	var tables []string

	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			if tableStr, ok := item.(string); ok {
				tableStr = strings.TrimSpace(tableStr)
				if tableStr != "" {
					tables = append(tables, tableStr)
				}
			} else if tableMap, ok := item.(map[string]interface{}); ok {
				// Handle object format: {table: "tablename", alias: "t"}
				if table, ok := tableMap["table"].(string); ok {
					if alias, hasAlias := tableMap["alias"].(string); hasAlias && alias != "" {
						tables = append(tables, fmt.Sprintf("%s AS %s", table, alias))
					} else {
						tables = append(tables, table)
					}
				} else if name, ok := tableMap["name"].(string); ok {
					tables = append(tables, name)
				}
			}
		}
	case string:
		if v != "" {
			tables = append(tables, v)
		}
	}

	return tables, nil
}

// extractJoins extracts JOIN clauses from joins JSON
func (s *ReportService) extractJoins(data interface{}) ([]string, error) {
	var joins []string

	joinArray, ok := data.([]interface{})
	if !ok {
		return joins, nil
	}

	for _, item := range joinArray {
		joinMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		joinType := "INNER JOIN"
		if jt, ok := joinMap["type"].(string); ok {
			joinType = strings.ToUpper(jt) + " JOIN"
		}

		table, ok := joinMap["table"].(string)
		if !ok {
			continue
		}

		var alias string
		if a, ok := joinMap["alias"].(string); ok && a != "" {
			alias = a
			table = fmt.Sprintf("%s AS %s", table, alias)
		}

		onCondition := ""
		if on, ok := joinMap["on"].(string); ok {
			onCondition = on
		} else if on, ok := joinMap["condition"].(string); ok {
			onCondition = on
		}

		if onCondition == "" {
			return nil, fmt.Errorf("join on table '%s' missing ON condition", table)
		}

		joinClause := fmt.Sprintf("%s %s ON %s", joinType, table, onCondition)
		joins = append(joins, joinClause)
	}

	return joins, nil
}

// extractFilters extracts WHERE conditions from filters JSON
func (s *ReportService) extractFilters(data interface{}, parameters map[string]interface{}) (string, []interface{}, error) {
	var conditions []string
	var params []interface{}

	filterArray, ok := data.([]interface{})
	if !ok || len(filterArray) == 0 {
		return "", nil, nil
	}

	for _, item := range filterArray {
		filterMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		field, ok := filterMap["field"].(string)
		if !ok || field == "" {
			continue
		}

		operator := "="
		if op, ok := filterMap["operator"].(string); ok {
			operator = op
		}

		value := filterMap["value"]

		// Handle parameter references
		if valueStr, ok := value.(string); ok && strings.HasPrefix(valueStr, "@") {
			paramName := strings.TrimPrefix(valueStr, "@")
			if paramValue, exists := parameters[paramName]; exists {
				value = paramValue
			}
		}

		// Build condition based on operator
		var condition string
		switch strings.ToUpper(operator) {
		case "IS NULL":
			condition = fmt.Sprintf("%s IS NULL", field)
		case "IS NOT NULL":
			condition = fmt.Sprintf("%s IS NOT NULL", field)
		case "IN":
			// Handle IN operator with array values
			if valueArray, ok := value.([]interface{}); ok {
				placeholders := make([]string, len(valueArray))
				for i, v := range valueArray {
					placeholders[i] = "?"
					params = append(params, v)
				}
				condition = fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", "))
			}
		case "BETWEEN":
			// Handle BETWEEN operator
			if valueArray, ok := value.([]interface{}); ok && len(valueArray) == 2 {
				condition = fmt.Sprintf("%s BETWEEN ? AND ?", field)
				params = append(params, valueArray[0], valueArray[1])
			}
		case "LIKE", "NOT LIKE":
			condition = fmt.Sprintf("%s %s ?", field, operator)
			params = append(params, value)
		default:
			// Standard comparison operators: =, !=, <, >, <=, >=
			condition = fmt.Sprintf("%s %s ?", field, operator)
			params = append(params, value)
		}

		if condition != "" {
			conditions = append(conditions, condition)
		}
	}

	if len(conditions) == 0 {
		return "", nil, nil
	}

	// Combine conditions with AND (could be enhanced to support OR groups)
	whereClause := strings.Join(conditions, " AND ")
	return whereClause, params, nil
}

// extractGrouping extracts GROUP BY fields from grouping JSON
func (s *ReportService) extractGrouping(data interface{}) (string, error) {
	var groupByFields []string

	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			if fieldStr, ok := item.(string); ok {
				fieldStr = strings.TrimSpace(fieldStr)
				if fieldStr != "" {
					groupByFields = append(groupByFields, fieldStr)
				}
			} else if fieldMap, ok := item.(map[string]interface{}); ok {
				if field, ok := fieldMap["field"].(string); ok {
					groupByFields = append(groupByFields, field)
				}
			}
		}
	case string:
		if v != "" {
			groupByFields = append(groupByFields, v)
		}
	}

	if len(groupByFields) == 0 {
		return "", nil
	}

	return strings.Join(groupByFields, ", "), nil
}

// extractSorting extracts ORDER BY clause from sorting JSON
func (s *ReportService) extractSorting(data interface{}) (string, error) {
	var orderByFields []string

	sortArray, ok := data.([]interface{})
	if !ok {
		return "", nil
	}

	for _, item := range sortArray {
		sortMap, ok := item.(map[string]interface{})
		if !ok {
			// Handle string format
			if sortStr, ok := item.(string); ok && sortStr != "" {
				orderByFields = append(orderByFields, sortStr)
			}
			continue
		}

		field, ok := sortMap["field"].(string)
		if !ok || field == "" {
			continue
		}

		direction := "ASC"
		if dir, ok := sortMap["direction"].(string); ok {
			direction = strings.ToUpper(dir)
		} else if dir, ok := sortMap["order"].(string); ok {
			direction = strings.ToUpper(dir)
		}

		orderByFields = append(orderByFields, fmt.Sprintf("%s %s", field, direction))
	}

	if len(orderByFields) == 0 {
		return "", nil
	}

	return strings.Join(orderByFields, ", "), nil
}
