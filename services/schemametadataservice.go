package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// SchemaMetadataService handles database schema discovery and metadata management
type SchemaMetadataService struct {
	db   *gorm.DB
	iLog logger.Log
}

// NewSchemaMetadataService creates a new schema metadata service
func NewSchemaMetadataService(db *gorm.DB) *SchemaMetadataService {
	return &SchemaMetadataService{
		db: db,
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "SchemaMetadataService",
		},
	}
}

// DiscoverDatabaseSchema discovers all tables and columns in a database
func (s *SchemaMetadataService) DiscoverDatabaseSchema(ctx context.Context, databaseAlias, dbName string) error {
	s.iLog.Info(fmt.Sprintf("Discovering schema for database '%s' with alias '%s'", dbName, databaseAlias))

	// Get all tables
	var tables []struct {
		TableName    string
		TableComment string
	}

	query := `
		SELECT
			TABLE_NAME as tablename,
			COALESCE(TABLE_COMMENT, '') as table_comment
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`

	if err := s.db.Raw(query, dbName).Scan(&tables).Error; err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to query INFORMATION_SCHEMA.TABLES for database '%s': %v", dbName, err))
		return fmt.Errorf("failed to discover tables: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Found %d tables in database '%s'", len(tables), dbName))

	if len(tables) == 0 {
		s.iLog.Warn(fmt.Sprintf("No tables found in database '%s' - database may be empty or TABLE_SCHEMA filter may be incorrect", dbName))
		return nil
	}

	// Process each table
	for _, table := range tables {
		// Create or update table metadata
		tableMeta := &models.DatabaseSchemaMetadata{
			DatabaseAlias: databaseAlias,
			Table:         table.TableName,
			MetadataType:  models.MetadataTypeTable,
			Description:   table.TableComment,
		}

		if err := s.db.WithContext(ctx).
			Where("databasealias = ? AND tablename = ? AND metadatatype = ?", databaseAlias, table.TableName, models.MetadataTypeTable).
			Assign(tableMeta).
			FirstOrCreate(tableMeta).Error; err != nil {
			return fmt.Errorf("failed to save table metadata for %s: %w", table.TableName, err)
		}

		// Get columns for this table
		var columns []struct {
			ColumnName    string
			DataType      string
			IsNullable    string
			ColumnKey     string
			ColumnComment string
		}

		columnQuery := `
			SELECT
				COLUMN_NAME as columnname,
				DATA_TYPE as data_type,
				IS_NULLABLE as is_nullable,
				COLUMN_KEY as column_key,
				COALESCE(COLUMN_COMMENT, '') as column_comment
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = ?
			AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION
		`

		if err := s.db.Raw(columnQuery, dbName, table.TableName).Scan(&columns).Error; err != nil {
			return fmt.Errorf("failed to discover columns for %s: %w", table.TableName, err)
		}

		// Process each column
		for _, column := range columns {
			isNullable := column.IsNullable == "YES"

			columnMeta := &models.DatabaseSchemaMetadata{
				DatabaseAlias: databaseAlias,
				Table:         table.TableName,
				Column:        column.ColumnName,
				MetadataType:  models.MetadataTypeColumn,
				DataType:      column.DataType,
				IsNullable:    &isNullable,
				ColumnComment: column.ColumnComment,
			}

			if err := s.db.WithContext(ctx).
				Where("databasealias = ? AND tablename = ? AND columnname = ? AND metadatatype = 'column'",
					databaseAlias, table.TableName, column.ColumnName).
				Assign(columnMeta).
				FirstOrCreate(columnMeta).Error; err != nil {
				return fmt.Errorf("failed to save column metadata for %s.%s: %w", table.TableName, column.ColumnName, err)
			}
		}
	}

	return nil
}

// GetTableMetadata retrieves metadata for all tables in a database
func (s *SchemaMetadataService) GetTableMetadata(ctx context.Context, databaseAlias string) ([]models.DatabaseSchemaMetadata, error) {
	var metadata []models.DatabaseSchemaMetadata

	if err := s.db.WithContext(ctx).
		Where("databasealias = ? AND metadatatype = ?", databaseAlias, models.MetadataTypeTable).
		Order("tablename").
		Find(&metadata).Error; err != nil {
		return nil, fmt.Errorf("failed to get table metadata: %w", err)
	}

	return metadata, nil
}

// GetColumnMetadata retrieves metadata for all columns in a table
func (s *SchemaMetadataService) GetColumnMetadata(ctx context.Context, databaseAlias, tableName string) ([]models.DatabaseSchemaMetadata, error) {
	var metadata []models.DatabaseSchemaMetadata

	if err := s.db.WithContext(ctx).
		Where("databasealias = ? AND tablename = ? AND metadatatype = ?", databaseAlias, tableName, models.MetadataTypeColumn).
		Order("columnname").
		Find(&metadata).Error; err != nil {
		return nil, fmt.Errorf("failed to get column metadata: %w", err)
	}

	return metadata, nil
}

// UpdateMetadata updates the description for a metadata entry
func (s *SchemaMetadataService) UpdateMetadata(ctx context.Context, id string, description *string) error {
	updates := make(map[string]interface{})

	if description != nil {
		updates["description"] = *description
	}

	if len(updates) == 0 {
		return nil
	}

	updates["modifiedon"] = gorm.Expr("CURRENT_TIMESTAMP")

	if err := s.db.WithContext(ctx).
		Model(&models.DatabaseSchemaMetadata{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}

// DeleteMetadata deletes a metadata entry
func (s *SchemaMetadataService) DeleteMetadata(ctx context.Context, id string) error {
	if err := s.db.WithContext(ctx).
		Delete(&models.DatabaseSchemaMetadata{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}

// GetSchemaContext builds a comprehensive schema context for AI
func (s *SchemaMetadataService) GetSchemaContext(ctx context.Context, databaseAlias string, tableNames []string) (string, error) {
	var metadata []models.DatabaseSchemaMetadata

	query := s.db.WithContext(ctx).Where("databasealias = ?", databaseAlias)

	if len(tableNames) > 0 {
		query = query.Where("tablename IN ?", tableNames)
	}

	if err := query.Order("tablename, metadatatype DESC, columnname").Find(&metadata).Error; err != nil {
		return "", fmt.Errorf("failed to get schema context: %w", err)
	}

	// Build context string
	var context strings.Builder
	context.WriteString("Database Schema Information:\n\n")

	currentTable := ""
	for _, meta := range metadata {
		if meta.MetadataType == models.MetadataTypeTable {
			currentTable = meta.Table
			context.WriteString(fmt.Sprintf("## Table: %s\n", meta.Table))

			if meta.Description != "" {
				context.WriteString(fmt.Sprintf("Description: %s\n", meta.Description))
			}

			context.WriteString("Columns:\n")
		} else if meta.MetadataType == models.MetadataTypeColumn && meta.Table == currentTable {
			columnInfo := fmt.Sprintf("  - %s", meta.Column)

			if meta.DataType != "" {
				columnInfo += fmt.Sprintf(" (%s)", meta.DataType)
			}

			if meta.IsNullable != nil && !*meta.IsNullable {
				columnInfo += " [NOT NULL]"
			}

			if meta.ColumnComment != "" {
				columnInfo += fmt.Sprintf(" - %s", meta.ColumnComment)
			}

			context.WriteString(columnInfo + "\n")
		}
	}

	return context.String(), nil
}

// GetAllDatabases returns a list of all unique database aliases
func (s *SchemaMetadataService) GetAllDatabases(ctx context.Context) ([]string, error) {
	var databases []string

	if err := s.db.WithContext(ctx).
		Model(&models.DatabaseSchemaMetadata{}).
		Distinct("databasealias").
		Pluck("databasealias", &databases).Error; err != nil {
		return nil, fmt.Errorf("failed to get databases: %w", err)
	}

	return databases, nil
}

// SearchMetadata searches metadata by keyword
func (s *SchemaMetadataService) SearchMetadata(ctx context.Context, databaseAlias, keyword string) ([]models.DatabaseSchemaMetadata, error) {
	var metadata []models.DatabaseSchemaMetadata

	searchPattern := "%" + keyword + "%"

	if err := s.db.WithContext(ctx).
		Where("databasealias = ?", databaseAlias).
		Where("tablename LIKE ? OR COALESCE(columnname, '') LIKE ? OR COALESCE(description, '') LIKE ? OR COALESCE(businessterms, '') LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern).
		Order("tablename, metadatatype DESC, columnname").
		Find(&metadata).Error; err != nil {
		return nil, fmt.Errorf("failed to search metadata: %w", err)
	}

	return metadata, nil
}

// GetDatabaseMetadata retrieves complete metadata (tables and columns) for a database
// If no metadata exists or existing metadata is incomplete, it automatically discovers and populates it from information_schema
func (s *SchemaMetadataService) GetDatabaseMetadata(ctx context.Context, databaseAlias string) ([]models.DatabaseSchemaMetadata, error) {
	var metadata []models.DatabaseSchemaMetadata

	// Use Find which returns empty slice if no records found (not an error)
	if err := s.db.WithContext(ctx).
		Where("databasealias = ?", databaseAlias).
		Order("tablename, metadatatype DESC, columnname").
		Find(&metadata).Error; err != nil {
		// Only return error for actual database errors, not "record not found"
		return nil, fmt.Errorf("failed to get database metadata: %w", err)
	}

	// Check if existing metadata is useful (has both tables and columns)
	needsDiscovery := false
	discoveryReason := ""

	if len(metadata) == 0 {
		needsDiscovery = true
		discoveryReason = "no metadata found"
		s.iLog.Info(fmt.Sprintf("No metadata found for alias '%s', attempting auto-discovery", databaseAlias))
	} else {
		// Count tables and columns
		tableCount := 0
		columnCount := 0
		for _, meta := range metadata {
			if meta.MetadataType == models.MetadataTypeTable {
				tableCount++
			} else if meta.MetadataType == models.MetadataTypeColumn {
				columnCount++
			}
		}

		// Check if metadata is incomplete
		if columnCount == 0 {
			needsDiscovery = true
			discoveryReason = fmt.Sprintf("found %d tables but no columns", tableCount)
		} else if tableCount < 5 {
			needsDiscovery = true
			discoveryReason = fmt.Sprintf("found only %d tables (need at least 5)", tableCount)
		} else if len(metadata) < 10 {
			needsDiscovery = true
			discoveryReason = fmt.Sprintf("found only %d total entries (need at least 10)", len(metadata))
		} else if tableCount > 0 && columnCount < (tableCount*2) {
			needsDiscovery = true
			discoveryReason = fmt.Sprintf("found %d tables but only %d columns (need at least 2 per table)", tableCount, columnCount)
		}

		if needsDiscovery {
			s.iLog.Warn(fmt.Sprintf("Existing metadata for alias '%s' is incomplete (%s), triggering auto-discovery", databaseAlias, discoveryReason))
			// Delete existing incomplete metadata before rediscovering
			if err := s.db.WithContext(ctx).Where("databasealias = ?", databaseAlias).Delete(&models.DatabaseSchemaMetadata{}).Error; err != nil {
				s.iLog.Error(fmt.Sprintf("Failed to delete incomplete metadata: %v", err))
			} else {
				s.iLog.Info("Deleted incomplete metadata, will perform fresh discovery")
			}
		}
	}

	// If metadata is missing or incomplete, automatically discover and populate it
	if needsDiscovery {

		// Get the current database name from the connection
		var dbName string
		if err := s.db.WithContext(ctx).Raw("SELECT DATABASE()").Scan(&dbName).Error; err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to execute SELECT DATABASE(): %v", err))
			return nil, fmt.Errorf("failed to get current database name: %w", err)
		}

		s.iLog.Debug(fmt.Sprintf("SELECT DATABASE() returned: '%s'", dbName))

		if dbName == "" {
			s.iLog.Warn("SELECT DATABASE() returned empty string, attempting to find database via SHOW DATABASES")

			// Query information_schema without database filter
			var databases []string
			if err := s.db.Raw("SHOW DATABASES").Scan(&databases).Error; err == nil && len(databases) > 0 {
				s.iLog.Debug(fmt.Sprintf("Found %d databases via SHOW DATABASES", len(databases)))
				// Try to use the first non-system database
				for _, db := range databases {
					if db != "information_schema" && db != "mysql" && db != "performance_schema" && db != "sys" {
						dbName = db
						s.iLog.Info(fmt.Sprintf("Selected database '%s' from available databases", dbName))
						break
					}
				}
			}

			// If still no database name, return error instead of empty metadata
			if dbName == "" {
				s.iLog.Error(fmt.Sprintf("Unable to determine database name for alias '%s'", databaseAlias))
				return nil, fmt.Errorf("no database selected and unable to determine database name for alias '%s' - check GORM connection string includes database name", databaseAlias)
			}
		}

		s.iLog.Info(fmt.Sprintf("Starting schema discovery for database '%s' with alias '%s'", dbName, databaseAlias))

		// Auto-discover schema for this database
		if err := s.DiscoverDatabaseSchema(ctx, databaseAlias, dbName); err != nil {
			s.iLog.Error(fmt.Sprintf("Schema discovery failed for database '%s': %v", dbName, err))
			return nil, fmt.Errorf("failed to auto-discover schema for database '%s': %w", dbName, err)
		}

		s.iLog.Info(fmt.Sprintf("Schema discovery completed for database '%s', retrieving metadata", dbName))

		// Retrieve the newly discovered metadata
		if err := s.db.WithContext(ctx).
			Where("databasealias = ?", databaseAlias).
			Order("tablename, metadatatype DESC, columnname").
			Find(&metadata).Error; err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to retrieve discovered metadata: %v", err))
			return nil, fmt.Errorf("failed to get discovered metadata: %w", err)
		}

		// If still no metadata after discovery, there might be no tables
		if len(metadata) == 0 {
			s.iLog.Warn(fmt.Sprintf("No tables discovered in database '%s' for alias '%s'", dbName, databaseAlias))
			return nil, fmt.Errorf("no tables found in database '%s' for alias '%s' - database may be empty", dbName, databaseAlias)
		}

		s.iLog.Info(fmt.Sprintf("Successfully discovered %d metadata entries for alias '%s'", len(metadata), databaseAlias))
	}

	return metadata, nil
}

// GetTableDetail retrieves detailed information about a specific table including its columns
func (s *SchemaMetadataService) GetTableDetail(ctx context.Context, databaseAlias, tableName, schemaName string) (map[string]interface{}, error) {
	var metadata []models.DatabaseSchemaMetadata

	query := s.db.WithContext(ctx).
		Where("databasealias = ? AND tablename = ?", databaseAlias, tableName)

	if err := query.Order("metadatatype DESC, columnname").Find(&metadata).Error; err != nil {
		return nil, fmt.Errorf("failed to get table detail: %w", err)
	}

	if len(metadata) == 0 {
		return nil, fmt.Errorf("table not found: %s", tableName)
	}

	// Build response structure
	result := map[string]interface{}{
		"table_name": tableName,
		"schema":     schemaName,
		"fields":     []map[string]interface{}{},
	}

	fields := []map[string]interface{}{}
	for _, meta := range metadata {
		if meta.MetadataType == models.MetadataTypeColumn {
			field := map[string]interface{}{
				"name":      meta.Column,
				"data_type": meta.DataType,
			}
			if meta.IsNullable != nil {
				field["is_nullable"] = *meta.IsNullable
			}
			if meta.ColumnComment != "" {
				field["comment"] = meta.ColumnComment
			}
			fields = append(fields, field)
		}
	}
	result["fields"] = fields

	return result, nil
}

// ExecuteVisualQuery converts a visual query structure to SQL and executes it
func (s *SchemaMetadataService) ExecuteVisualQuery(ctx context.Context, databaseAlias string, visualQuery map[string]interface{}) (map[string]interface{}, error) {
	// Parse the visual query structure
	vq, err := ParseVisualQuery(visualQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to parse visual query: %w", err)
	}

	// Generate SQL from visual query
	sqlQuery, args, err := vq.GenerateSQL()
	if err != nil {
		return nil, fmt.Errorf("failed to generate SQL: %w", err)
	}

	// Execute the generated SQL
	return s.executeQueryWithResults(ctx, databaseAlias, sqlQuery, args...)
}

// ExecuteCustomSQL executes a custom SQL query
func (s *SchemaMetadataService) ExecuteCustomSQL(ctx context.Context, databaseAlias string, sqlQuery string) (map[string]interface{}, error) {
	// Validate the SQL first
	validation, err := s.ValidateSQL(ctx, databaseAlias, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to validate SQL: %w", err)
	}

	// Check if validation passed
	if valid, ok := validation["valid"].(bool); !ok || !valid {
		if errMsg, ok := validation["error"].(string); ok {
			return nil, fmt.Errorf("SQL validation failed: %s", errMsg)
		}
		return nil, fmt.Errorf("SQL validation failed")
	}

	// Execute the query
	return s.executeQueryWithResults(ctx, databaseAlias, sqlQuery)
}

// ValidateSQL validates SQL syntax without executing it
func (s *SchemaMetadataService) ValidateSQL(ctx context.Context, databaseAlias string, sqlQuery string) (map[string]interface{}, error) {
	// Basic validation checks
	if strings.TrimSpace(sqlQuery) == "" {
		return map[string]interface{}{
			"valid": false,
			"error": "SQL query cannot be empty",
		}, nil
	}

	// Check for dangerous operations
	dangerousKeywords := []string{"DROP", "DELETE", "TRUNCATE", "ALTER", "CREATE"}
	upperSQL := strings.ToUpper(sqlQuery)
	for _, keyword := range dangerousKeywords {
		if strings.Contains(upperSQL, keyword) {
			return map[string]interface{}{
				"valid":   false,
				"error":   fmt.Sprintf("Query contains potentially dangerous keyword: %s", keyword),
				"warning": "Only SELECT queries are allowed in the query builder",
			}, nil
		}
	}

	// Check if it's a SELECT query
	if !strings.HasPrefix(strings.TrimSpace(upperSQL), "SELECT") {
		return map[string]interface{}{
			"valid": false,
			"error": "Only SELECT queries are supported in the query builder",
		}, nil
	}

	// In production, you would also:
	// 1. Use database-specific EXPLAIN to validate syntax
	// 2. Check table/column existence
	// 3. Validate permissions

	return map[string]interface{}{
		"valid":   true,
		"message": "SQL query is valid",
	}, nil
}

// executeQueryWithResults executes a SQL query and returns formatted results
func (s *SchemaMetadataService) executeQueryWithResults(ctx context.Context, databaseAlias string, sqlQuery string, args ...interface{}) (map[string]interface{}, error) {
	// Use GORM's Raw query to execute the SQL
	// NOTE: For now, this executes against the application's database
	// In a multi-database setup, you would get the connection from a pool based on databaseAlias

	var results []map[string]interface{}

	// Execute the query using GORM
	rows, err := s.db.WithContext(ctx).Raw(sqlQuery, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Get column types
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("failed to get column types: %w", err)
	}

	// Build metadata about columns
	columnMetadata := make([]map[string]interface{}, len(columns))
	for i, col := range columns {
		columnMetadata[i] = map[string]interface{}{
			"name": col,
			"type": columnTypes[i].DatabaseTypeName(),
		}
	}

	// Fetch all rows
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			row[col] = v
		}
		results = append(results, row)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Return formatted response
	return map[string]interface{}{
		"columns": columnMetadata,
		"rows":    results,
		"count":   len(results),
		"query":   sqlQuery,
	}, nil
}
