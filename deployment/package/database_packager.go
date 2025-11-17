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
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/deployment/models"
	"github.com/mdaxf/iac/logger"
)

// DatabasePackager handles packaging of relational database data
type DatabasePackager struct {
	dbTx         *sql.Tx
	dbOperation  *dbconn.DBOperation
	logger       logger.Log
	databaseType string
}

// NewDatabasePackager creates a new database packager
func NewDatabasePackager(user string, dbTx *sql.Tx, databaseType string) *DatabasePackager {
	iLog := logger.Log{ModuleName: logger.Framework, User: user, ControllerName: "DatabasePackager"}

	return &DatabasePackager{
		dbTx:         dbTx,
		dbOperation:  dbconn.NewDBOperation(user, dbTx, logger.Framework),
		logger:       iLog,
		databaseType: databaseType,
	}
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
	rows, err := dp.dbOperation.Query(query)
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
				if !processedTables[fk.ReferencedTable] {
					*relatedTables = append(*relatedTables, fk.ReferencedTable)
				}
			}
		}
	}

	tableData.RowCount = len(tableData.Rows)
	pkg.DatabaseData.Tables = append(pkg.DatabaseData.Tables, tableData)

	// Create PK mapping
	pkMapping := models.PKMapping{
		TableName: tableName,
		PKColumns: pkColumns,
		Strategy:  dp.determinePKStrategy(columns),
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
	case "postgresql":
		query = fmt.Sprintf(`
			SELECT
				column_name,
				data_type,
				CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_pk,
				is_nullable = 'YES' as is_nullable,
				character_maximum_length
			FROM information_schema.columns c
			LEFT JOIN (
				SELECT ku.column_name
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage ku
					ON tc.constraint_name = ku.constraint_name
				WHERE tc.constraint_type = 'PRIMARY KEY'
				AND tc.table_name = '%s'
			) pk ON c.column_name = pk.column_name
			WHERE c.table_name = '%s'
			ORDER BY ordinal_position`, tableName, tableName)
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

	rows, err := dp.dbOperation.Query(query)
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
	case "postgresql":
		query = fmt.Sprintf(`
			SELECT
				kcu.column_name,
				ccu.table_name AS referenced_table,
				ccu.column_name AS referenced_column,
				tc.constraint_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
			JOIN information_schema.constraint_column_usage ccu
				ON ccu.constraint_name = tc.constraint_name
			WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_name = '%s'`, tableName)
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

	rows, err := dp.dbOperation.Query(query)
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
			case "postgresql":
				// PostgreSQL uses sequences
				query = fmt.Sprintf("SELECT last_value FROM %s_id_seq", tableName)
			case "mssql":
				query = fmt.Sprintf("SELECT IDENT_CURRENT('%s')", tableName)
			default:
				continue
			}

			rows, err := dp.dbOperation.Query(query)
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
