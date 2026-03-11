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

package packagemgr

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/deployment/models"
	"github.com/mdaxf/iac/logger"
)

// DatabasePackager handles packaging of relational database data
type DatabasePackager struct {
	db           *sql.DB  // for schema introspection (ListTables, GetTableSourceInfo, etc.)
	dbTx         *sql.Tx  // for transactional data reads during actual packaging
	dbOperation  *dbconn.DBOperation
	logger       logger.Log
	databaseType string
	// tableSchema caches the schema name for each table, populated by ListTables()
	tableSchema  map[string]string
}

// NewDatabasePackager creates a packager for actual data packaging (uses transaction).
func NewDatabasePackager(user string, dbTx *sql.Tx, databaseType string) *DatabasePackager {
	iLog := logger.Log{ModuleName: logger.Framework, User: user, ControllerName: "DatabasePackager"}
	return &DatabasePackager{
		db:           dbconn.DB,
		dbTx:         dbTx,
		dbOperation:  dbconn.NewDBOperation(user, dbTx, logger.Framework),
		logger:       iLog,
		databaseType: strings.ToLower(databaseType),
	}
}

// NewDatabasePackagerFromDB creates a packager for schema-only operations (no transaction needed).
func NewDatabasePackagerFromDB(user string, db *sql.DB, databaseType string) *DatabasePackager {
	iLog := logger.Log{ModuleName: logger.Framework, User: user, ControllerName: "DatabasePackager"}
	return &DatabasePackager{
		db:           db,
		dbTx:         nil,
		dbOperation:  dbconn.NewDBOperation(user, nil, logger.Framework), // for QuoteIdentifier only
		logger:       iLog,
		databaseType: strings.ToLower(databaseType),
	}
}

// getCurrentSchema detects the active schema/database name at runtime.
// MySQL → SELECT DATABASE(), PostgreSQL → SELECT current_schema()
func (dp *DatabasePackager) getCurrentSchema() string {
	if dp.db == nil {
		return "public"
	}
	var schema string
	// MySQL
	if err := dp.db.QueryRow("SELECT DATABASE()").Scan(&schema); err == nil && schema != "" {
		return schema
	}
	// PostgreSQL / others
	if err := dp.db.QueryRow("SELECT current_schema()").Scan(&schema); err == nil && schema != "" {
		return schema
	}
	return "public"
}

// convertPlaceholders converts ? to database-specific placeholders ($1 for postgres, etc.)
func convertPlaceholders(query string, dbType string) string {
	switch dbType {
	case "mysql", "mariadb":
		return query // MySQL uses ?
	}
	count := 0
	return regexp.MustCompile(`\?`).ReplaceAllStringFunc(query, func(_ string) string {
		count++
		switch dbType {
		case "postgres", "postgresql":
			return fmt.Sprintf("$%d", count)
		case "mssql", "sqlserver":
			return fmt.Sprintf("@p%d", count)
		case "oracle":
			return fmt.Sprintf(":%d", count)
		}
		return "?"
	})
}

// execQuery runs a query on dp.db (schema queries) or dp.dbTx (data queries).
// Placeholder conversion is applied automatically.
func (dp *DatabasePackager) execQuery(query string, args ...interface{}) (*sql.Rows, error) {
	q := convertPlaceholders(query, dp.databaseType)
	if dp.db != nil {
		return dp.db.Query(q, args...)
	}
	if dp.dbTx != nil {
		return dp.dbTx.QueryContext(context.Background(), q, args...)
	}
	return nil, fmt.Errorf("no database connection available")
}

// rawQuery is kept for backward compatibility inside PackageTables data extraction.
func (dp *DatabasePackager) rawQuery(query string, args ...interface{}) (*sql.Rows, error) {
	return dp.execQuery(query, args...)
}

// PackageTables packages specified tables into a deployable package
func (dp *DatabasePackager) PackageTables(packageName, version, createdBy string, filter models.PackageFilter) (*models.Package, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		dp.logger.PerformanceWithDuration("DatabasePackager.PackageTables", elapsed)
	}()

	dp.logger.Info(fmt.Sprintf("Starting database packaging: %s v%s", packageName, version))

	pkg := &models.Package{
		ID:          uuid.New().String(),
		Name:        packageName,
		Version:     version,
		PackageType: "database",
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
		Metadata:    make(map[string]interface{}),
		DatabaseData: &models.DatabasePackage{
			Tables:        make([]models.TableData, 0),
			PKMappings:    make(map[string]models.PKMapping),
			Relationships: make([]models.Relationship, 0),
			SequenceInfo:  make(map[string]int64),
			DatabaseType:  dp.databaseType,
		},
		IncludeParent: filter.IncludeRelated,
	}

	// Track processed tables to avoid duplicates
	processedTables := make(map[string]bool)
	relatedTables := make([]string, 0)

	// Process primary tables
	for _, tableName := range filter.Tables {
		if err := dp.packageTable(pkg, tableName, filter, processedTables, &relatedTables, 0, filter.MaxDepth); err != nil {
			dp.logger.Error(fmt.Sprintf("Error packaging table %s: %v", tableName, err))
			return nil, err
		}
	}

	// Build relationship graph
	if err := dp.buildRelationships(pkg); err != nil {
		dp.logger.Error(fmt.Sprintf("Error building relationships: %v", err))
		return nil, err
	}

	// Get sequence information for auto-increment fields
	if err := dp.getSequenceInfo(pkg); err != nil {
		dp.logger.Error(fmt.Sprintf("Error getting sequence info: %v", err))
		// Non-fatal, continue
	}

	dp.logger.Info(fmt.Sprintf("Package created: %s with %d tables", pkg.ID, len(pkg.DatabaseData.Tables)))
	return pkg, nil
}

// packageTable packages a single table's data
func (dp *DatabasePackager) packageTable(pkg *models.Package, tableName string, filter models.PackageFilter,
	processedTables map[string]bool, relatedTables *[]string, depth, maxDepth int) error {

	// Check if already processed
	if processedTables[tableName] {
		return nil
	}

	// Check depth limit
	if maxDepth > 0 && depth > maxDepth {
		dp.logger.Debug(fmt.Sprintf("Skipping table %s: max depth %d reached", tableName, maxDepth))
		return nil
	}

	processedTables[tableName] = true
	dp.logger.Debug(fmt.Sprintf("Packaging table: %s (depth: %d)", tableName, depth))

	tableData := models.TableData{
		TableName: tableName,
		Columns:   make([]models.ColumnInfo, 0),
		Rows:      make([]map[string]interface{}, 0),
	}

	// Get table schema
	columns, err := dp.getTableSchema(tableName)
	if err != nil {
		return fmt.Errorf("failed to get schema for table %s: %w", tableName, err)
	}
	tableData.Columns = columns

	// Extract PK columns
	pkColumns := make([]string, 0)
	for _, col := range columns {
		if col.IsPrimaryKey {
			pkColumns = append(pkColumns, col.Name)
		}
	}
	tableData.PKColumns = pkColumns

	// Get FK information
	fkInfo, err := dp.getForeignKeys(tableName)
	if err != nil {
		dp.logger.Warn(fmt.Sprintf("Failed to get FK info for %s: %v", tableName, err))
	} else {
		tableData.FKColumns = fkInfo
	}

	// Build query with optional WHERE clause
	query := dp.buildSelectQuery(tableName, columns, filter)

	// Execute query and get data
	rows, err := dp.rawQuery(query)
	if err != nil {
		return fmt.Errorf("failed to query table %s: %w", tableName, err)
	}
	defer rows.Close()

	// Process rows
	columnNames := make([]string, len(columns))
	for i, col := range columns {
		columnNames[i] = col.Name
	}

	for rows.Next() {
		rowData := make(map[string]interface{})
		values := make([]interface{}, len(columnNames))
		valuePtrs := make([]interface{}, len(columnNames))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			dp.logger.Error(fmt.Sprintf("Error scanning row: %v", err))
			continue
		}

		for i, colName := range columnNames {
			// Convert []byte to string for text fields
			if b, ok := values[i].([]byte); ok {
				rowData[colName] = string(b)
			} else {
				rowData[colName] = values[i]
			}
		}

		tableData.Rows = append(tableData.Rows, rowData)

		// If include related is enabled, find related records
		if filter.IncludeRelated && len(fkInfo) > 0 {
			for _, fk := range fkInfo {
				if processedTables[fk.ReferencedTable] {
					continue
				}
				// Honor SelectRelations: if specified for this table, only follow listed FK columns
				if selectCols, ok := filter.SelectRelations[tableName]; ok && len(selectCols) > 0 {
					followed := false
					for _, sc := range selectCols {
						if sc == fk.ColumnName {
							followed = true
							break
						}
					}
					if !followed {
						continue
					}
				}
				// Honor ExcludeRelations: skip FK columns listed for this table
				if excludeCols, ok := filter.ExcludeRelations[tableName]; ok {
					skip := false
					for _, ec := range excludeCols {
						if ec == fk.ColumnName {
							skip = true
							break
						}
					}
					if skip {
						continue
					}
				}
				*relatedTables = append(*relatedTables, fk.ReferencedTable)
			}
		}
	}

	tableData.RowCount = len(tableData.Rows)
	pkg.DatabaseData.Tables = append(pkg.DatabaseData.Tables, tableData)

	// Create PK mapping
	pkMapping := models.PKMapping{
		TableName:        tableName,
		PKColumns:        pkColumns,
		Strategy:         dp.determinePKStrategy(columns),
		GlobalKeyColumns: filter.GlobalKeyColumns[tableName], // from filter
	}

	// Check if auto-increment
	for _, col := range columns {
		if col.IsPrimaryKey && strings.Contains(strings.ToLower(col.DataType), "auto_increment") {
			pkMapping.IsAutoIncrement = true
			break
		}
	}

	pkg.DatabaseData.PKMappings[tableName] = pkMapping

	// Process related tables
	if filter.IncludeRelated {
		for _, relatedTable := range *relatedTables {
			if err := dp.packageTable(pkg, relatedTable, filter, processedTables, relatedTables, depth+1, maxDepth); err != nil {
				dp.logger.Warn(fmt.Sprintf("Failed to package related table %s: %v", relatedTable, err))
			}
		}
	}

	return nil
}

// getTableSchema retrieves schema information for a table
func (dp *DatabasePackager) getTableSchema(tableName string) ([]models.ColumnInfo, error) {
	var query string

	switch strings.ToLower(dp.databaseType) {
	case "mysql":
		query = fmt.Sprintf(`
			SELECT
				COLUMN_NAME,
				DATA_TYPE,
				COLUMN_KEY = 'PRI' as IS_PK,
				IS_NULLABLE = 'YES' as IS_NULLABLE,
				CHARACTER_MAXIMUM_LENGTH
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_NAME = '%s'
			AND TABLE_SCHEMA = DATABASE()
			ORDER BY ORDINAL_POSITION`, tableName)
	case "postgresql", "postgres":
		// Use pg_catalog — works across all schemas without information_schema permission quirks
		dp.lookupTableSchema(tableName)
		schemaFilter := "n.nspname NOT IN ('pg_catalog','information_schema','pg_toast')"
		if dp.tableSchema != nil {
			if schema, ok := dp.tableSchema[tableName]; ok && schema != "" {
				schemaFilter = fmt.Sprintf("n.nspname = '%s'", schema)
			}
		}
		query = fmt.Sprintf(`
			SELECT
				a.attname AS column_name,
				pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type,
				EXISTS (
					SELECT 1 FROM pg_index i
					WHERE i.indrelid = a.attrelid AND i.indisprimary
					AND a.attnum = ANY(i.indkey)
				) AS is_pk,
				NOT a.attnotnull AS is_nullable,
				NULL::integer AS character_maximum_length
			FROM pg_catalog.pg_attribute a
			JOIN pg_catalog.pg_class c ON c.oid = a.attrelid
			JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
			WHERE c.relname = '%s'
			AND %s
			AND a.attnum > 0
			AND NOT a.attisdropped
			ORDER BY a.attnum`, tableName, schemaFilter)
	case "mssql":
		query = fmt.Sprintf(`
			SELECT
				c.COLUMN_NAME,
				c.DATA_TYPE,
				CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END as IS_PK,
				CASE WHEN c.IS_NULLABLE = 'YES' THEN 1 ELSE 0 END as IS_NULLABLE,
				c.CHARACTER_MAXIMUM_LENGTH
			FROM INFORMATION_SCHEMA.COLUMNS c
			LEFT JOIN (
				SELECT ku.COLUMN_NAME
				FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
				JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE ku
					ON tc.CONSTRAINT_NAME = ku.CONSTRAINT_NAME
				WHERE tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
				AND tc.TABLE_NAME = '%s'
			) pk ON c.COLUMN_NAME = pk.COLUMN_NAME
			WHERE c.TABLE_NAME = '%s'
			ORDER BY c.ORDINAL_POSITION`, tableName, tableName)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dp.databaseType)
	}

	rows, err := dp.rawQuery(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make([]models.ColumnInfo, 0)
	for rows.Next() {
		var col models.ColumnInfo
		var maxLength sql.NullInt64
		var isPK interface{}
		var isNullable interface{}

		if err := rows.Scan(&col.Name, &col.DataType, &isPK, &isNullable, &maxLength); err != nil {
			return nil, err
		}

		// Handle boolean conversion
		switch v := isPK.(type) {
		case bool:
			col.IsPrimaryKey = v
		case int64:
			col.IsPrimaryKey = v != 0
		case []byte:
			col.IsPrimaryKey = string(v) == "1" || strings.ToLower(string(v)) == "true"
		}

		switch v := isNullable.(type) {
		case bool:
			col.IsNullable = v
		case int64:
			col.IsNullable = v != 0
		case []byte:
			col.IsNullable = string(v) == "1" || strings.ToLower(string(v)) == "true"
		}

		if maxLength.Valid {
			col.MaxLength = int(maxLength.Int64)
		}

		columns = append(columns, col)
	}

	return columns, nil
}

// getForeignKeys retrieves foreign key information for a table
func (dp *DatabasePackager) getForeignKeys(tableName string) ([]models.ForeignKeyInfo, error) {
	var query string

	switch strings.ToLower(dp.databaseType) {
	case "mysql":
		query = fmt.Sprintf(`
			SELECT
				COLUMN_NAME,
				REFERENCED_TABLE_NAME,
				REFERENCED_COLUMN_NAME,
				CONSTRAINT_NAME
			FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
			WHERE TABLE_NAME = '%s'
			AND TABLE_SCHEMA = DATABASE()
			AND REFERENCED_TABLE_NAME IS NOT NULL`, tableName)
	case "postgresql", "postgres":
		// Use pg_catalog for FK info — reliable across all schemas
		query = fmt.Sprintf(`
			SELECT
				a.attname AS column_name,
				c2.relname AS referenced_table,
				a2.attname AS referenced_column,
				con.conname AS constraint_name
			FROM pg_constraint con
			JOIN pg_class c ON c.oid = con.conrelid
			JOIN pg_class c2 ON c2.oid = con.confrelid
			JOIN pg_namespace n ON n.oid = c.relnamespace
			JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = con.conkey[1]
			JOIN pg_attribute a2 ON a2.attrelid = c2.oid AND a2.attnum = con.confkey[1]
			WHERE con.contype = 'f'
			AND c.relname = '%s'`, tableName)
	case "mssql":
		query = fmt.Sprintf(`
			SELECT
				COL_NAME(fc.parent_object_id, fc.parent_column_id) AS column_name,
				OBJECT_NAME(fc.referenced_object_id) AS referenced_table,
				COL_NAME(fc.referenced_object_id, fc.referenced_column_id) AS referenced_column,
				fk.name AS constraint_name
			FROM sys.foreign_keys AS fk
			INNER JOIN sys.foreign_key_columns AS fc
				ON fk.object_id = fc.constraint_object_id
			WHERE OBJECT_NAME(fc.parent_object_id) = '%s'`, tableName)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dp.databaseType)
	}

	rows, err := dp.rawQuery(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fkInfo := make([]models.ForeignKeyInfo, 0)
	for rows.Next() {
		var fk models.ForeignKeyInfo
		if err := rows.Scan(&fk.ColumnName, &fk.ReferencedTable, &fk.ReferencedColumn, &fk.ConstraintName); err != nil {
			return nil, err
		}
		fkInfo = append(fkInfo, fk)
	}

	return fkInfo, nil
}

// buildSelectQuery constructs a SELECT query with optional WHERE clause
func (dp *DatabasePackager) buildSelectQuery(tableName string, columns []models.ColumnInfo, filter models.PackageFilter) string {
	columnNames := make([]string, 0)
	excludeColumns := filter.ExcludeColumns[tableName]

	for _, col := range columns {
		// Check if column should be excluded
		excluded := false
		for _, excCol := range excludeColumns {
			if col.Name == excCol {
				excluded = true
				break
			}
		}
		if !excluded {
			columnNames = append(columnNames, dp.dbOperation.QuoteIdentifier(col.Name))
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(columnNames, ", "),
		dp.dbOperation.QuoteIdentifier(tableName))

	// Add WHERE clause if specified
	if whereClause, ok := filter.WhereClause[tableName]; ok && whereClause != "" {
		query += " WHERE " + whereClause
	}

	return query
}

// buildRelationships builds the relationship graph
func (dp *DatabasePackager) buildRelationships(pkg *models.Package) error {
	for _, table := range pkg.DatabaseData.Tables {
		for _, fk := range table.FKColumns {
			rel := models.Relationship{
				ID:             uuid.New().String(),
				SourceTable:    table.TableName,
				SourceColumn:   fk.ColumnName,
				TargetTable:    fk.ReferencedTable,
				TargetColumn:   fk.ReferencedColumn,
				ConstraintName: fk.ConstraintName,
			}
			pkg.DatabaseData.Relationships = append(pkg.DatabaseData.Relationships, rel)
		}
	}
	return nil
}

// getSequenceInfo retrieves current sequence values
func (dp *DatabasePackager) getSequenceInfo(pkg *models.Package) error {
	for tableName, pkMapping := range pkg.DatabaseData.PKMappings {
		if pkMapping.IsAutoIncrement {
			var query string

			switch strings.ToLower(dp.databaseType) {
			case "mysql":
				query = fmt.Sprintf("SELECT AUTO_INCREMENT FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = '%s' AND TABLE_SCHEMA = DATABASE()", tableName)
			case "postgresql", "postgres":
				// PostgreSQL uses sequences
				query = fmt.Sprintf("SELECT last_value FROM %s_id_seq", tableName)
			case "mssql":
				query = fmt.Sprintf("SELECT IDENT_CURRENT('%s')", tableName)
			default:
				continue
			}

			rows, err := dp.rawQuery(query)
			if err != nil {
				dp.logger.Warn(fmt.Sprintf("Failed to get sequence info for %s: %v", tableName, err))
				continue
			}
			defer rows.Close()

			if rows.Next() {
				var seqValue int64
				if err := rows.Scan(&seqValue); err == nil {
					pkg.DatabaseData.SequenceInfo[tableName] = seqValue
				}
			}
		}
	}
	return nil
}

// determinePKStrategy determines the best PK generation strategy
func (dp *DatabasePackager) determinePKStrategy(columns []models.ColumnInfo) string {
	for _, col := range columns {
		if !col.IsPrimaryKey {
			continue
		}

		dataType := strings.ToLower(col.DataType)

		// Check for auto-increment
		if strings.Contains(dataType, "auto_increment") || strings.Contains(dataType, "serial") {
			return "auto_increment"
		}

		// Check for sequence (PostgreSQL)
		if strings.Contains(dataType, "nextval") {
			return "sequence"
		}

		// UUID/GUID should be preserved as they are globally unique
		if strings.Contains(dataType, "uuid") || strings.Contains(dataType, "uniqueidentifier") {
			return "preserve"
		}
	}

	// Default to preserve if no special strategy detected
	return "preserve"
}

// ExportPackage exports package to JSON
func (dp *DatabasePackager) ExportPackage(pkg *models.Package) ([]byte, error) {
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package: %w", err)
	}
	return data, nil
}

// ImportPackage imports package from JSON
func (dp *DatabasePackager) ImportPackage(data []byte) (*models.Package, error) {
	var pkg models.Package
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package: %w", err)
	}
	return &pkg, nil
}

// lookupTableSchema finds which schema a table belongs to (for PostgreSQL).
// Result is cached in dp.tableSchema. Safe to call multiple times.
func (dp *DatabasePackager) lookupTableSchema(tableName string) {
	if dp.tableSchema == nil {
		dp.tableSchema = make(map[string]string)
	}
	if _, ok := dp.tableSchema[tableName]; ok {
		return // already cached
	}
	switch dp.databaseType {
	case "postgres", "postgresql":
		var schema string
		q := convertPlaceholders(`SELECT schemaname FROM pg_tables WHERE tablename = ? AND schemaname NOT IN ('pg_catalog','information_schema','pg_toast') LIMIT 1`, dp.databaseType)
		if err := dp.db.QueryRow(q, tableName).Scan(&schema); err == nil && schema != "" {
			dp.tableSchema[tableName] = schema
		}
	}
}

// qualifiedName returns a schema-qualified table reference for PostgreSQL when the table
// was discovered via ListTables() or lookupTableSchema().
func (dp *DatabasePackager) qualifiedName(tableName string) string {
	dp.lookupTableSchema(tableName)
	if dp.tableSchema != nil {
		if schema, ok := dp.tableSchema[tableName]; ok && schema != "" {
			return fmt.Sprintf("%q.%q", schema, tableName)
		}
	}
	return dp.dbOperation.QuoteIdentifier(tableName)
}

// CountTableRows returns the total number of rows in a table.
func (dp *DatabasePackager) CountTableRows(tableName string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", dp.qualifiedName(tableName))
	rows, err := dp.rawQuery(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var count int
	if rows.Next() {
		rows.Scan(&count) //nolint:errcheck
	}
	return count, nil
}

// QueryTableRows returns paginated rows from a table for the record browser.
// Returns (columnNames, pkColumns, rows, error).
func (dp *DatabasePackager) QueryTableRows(tableName string, limit, offset int) ([]string, []string, []map[string]interface{}, error) {
	columns, err := dp.getTableSchema(tableName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get schema for %s: %w", tableName, err)
	}

	colNames := make([]string, len(columns))
	pkCols := make([]string, 0)
	for i, c := range columns {
		colNames[i] = c.Name
		if c.IsPrimaryKey {
			pkCols = append(pkCols, c.Name)
		}
	}

	quotedCols := make([]string, len(colNames))
	for i, n := range colNames {
		quotedCols[i] = dp.dbOperation.QuoteIdentifier(n)
	}

	var query string
	switch dp.databaseType {
	case "mssql":
		query = fmt.Sprintf(
			"SELECT %s FROM %s ORDER BY (SELECT NULL) OFFSET %d ROWS FETCH NEXT %d ROWS ONLY",
			strings.Join(quotedCols, ", "),
			dp.qualifiedName(tableName),
			offset, limit,
		)
	default:
		query = fmt.Sprintf(
			"SELECT %s FROM %s LIMIT %d OFFSET %d",
			strings.Join(quotedCols, ", "),
			dp.qualifiedName(tableName),
			limit, offset,
		)
	}

	rows, err := dp.rawQuery(query)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to query rows from %s: %w", tableName, err)
	}
	defer rows.Close()

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(colNames))
		valuePtrs := make([]interface{}, len(colNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, name := range colNames {
			if b, ok := values[i].([]byte); ok {
				row[name] = string(b)
			} else {
				row[name] = values[i]
			}
		}
		result = append(result, row)
	}

	return colNames, pkCols, result, nil
}

// ListTables returns all user table names in the connected database.
// For PostgreSQL it also populates dp.tableSchema so data queries can be schema-qualified.
func (dp *DatabasePackager) ListTables() ([]string, error) {
	dp.logger.Info(fmt.Sprintf("ListTables: dbType=%s", dp.databaseType))

	dp.tableSchema = make(map[string]string)

	switch dp.databaseType {
	case "postgres", "postgresql":
		// Select schemaname too so we can qualify data queries later
		rows, err := dp.execQuery(`SELECT schemaname, tablename FROM pg_tables WHERE schemaname NOT IN ('pg_catalog','information_schema','pg_toast') ORDER BY tablename`)
		if err != nil {
			return nil, fmt.Errorf("failed to list tables (type=%s): %w", dp.databaseType, err)
		}
		defer rows.Close()
		names := make([]string, 0)
		for rows.Next() {
			var schema, tbl string
			if err := rows.Scan(&schema, &tbl); err == nil {
				names = append(names, tbl)
				dp.tableSchema[tbl] = schema
			}
		}
		dp.logger.Info(fmt.Sprintf("ListTables: found %d tables", len(names)))
		return names, nil

	case "mysql", "mariadb":
		rows, err := dp.execQuery(`SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_NAME`)
		if err != nil {
			return nil, fmt.Errorf("failed to list tables (type=%s): %w", dp.databaseType, err)
		}
		defer rows.Close()
		names := make([]string, 0)
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err == nil {
				names = append(names, name)
			}
		}
		dp.logger.Info(fmt.Sprintf("ListTables: found %d tables", len(names)))
		return names, nil

	case "mssql", "sqlserver":
		rows, err := dp.execQuery(`SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_NAME`)
		if err != nil {
			return nil, fmt.Errorf("failed to list tables (type=%s): %w", dp.databaseType, err)
		}
		defer rows.Close()
		names := make([]string, 0)
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err == nil {
				names = append(names, name)
			}
		}
		dp.logger.Info(fmt.Sprintf("ListTables: found %d tables", len(names)))
		return names, nil

	default:
		schema := dp.getCurrentSchema()
		rows, err := dp.execQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = ? AND table_type = 'BASE TABLE' ORDER BY table_name`, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to list tables (type=%s): %w", dp.databaseType, err)
		}
		defer rows.Close()
		names := make([]string, 0)
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err == nil {
				names = append(names, name)
			}
		}
		dp.logger.Info(fmt.Sprintf("ListTables: found %d tables", len(names)))
		return names, nil
	}
}

// GetTableSourceInfo returns schema + row count for a single table (for the sources browser)
func (dp *DatabasePackager) GetTableSourceInfo(tableName string) (*models.TableSource, error) {
	columns, err := dp.getTableSchema(tableName)
	if err != nil {
		return nil, err
	}

	pkCols := make([]string, 0)
	for _, c := range columns {
		if c.IsPrimaryKey {
			pkCols = append(pkCols, c.Name)
		}
	}

	fks, _ := dp.getForeignKeys(tableName) // non-fatal

	// Row count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", dp.qualifiedName(tableName))
	countRows, err := dp.rawQuery(countQuery)
	rowCount := 0
	if err == nil {
		defer countRows.Close()
		if countRows.Next() {
			countRows.Scan(&rowCount) //nolint:errcheck
		}
	}

	return &models.TableSource{
		Name:      tableName,
		RowCount:  rowCount,
		Columns:   columns,
		PKColumns: pkCols,
		FKColumns: fks,
	}, nil
}
