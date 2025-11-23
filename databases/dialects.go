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

package dbconn

import (
	"fmt"
	"strings"
)

// MySQLDialect implements MySQL-specific SQL operations
type MySQLDialect struct{}

func NewMySQLDialect() *MySQLDialect {
	return &MySQLDialect{}
}

func (d *MySQLDialect) QuoteIdentifier(name string) string {
	return "`" + strings.Replace(name, "`", "``", -1) + "`"
}

func (d *MySQLDialect) Placeholder(n int) string {
	return "?"
}

func (d *MySQLDialect) LimitOffset(limit, offset int) string {
	if offset > 0 {
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}
	return fmt.Sprintf("LIMIT %d", limit)
}

func (d *MySQLDialect) DataTypeMapping(genericType string) string {
	mappings := map[string]string{
		"string":    "VARCHAR(255)",
		"text":      "TEXT",
		"int":       "INT",
		"bigint":    "BIGINT",
		"float":     "FLOAT",
		"double":    "DOUBLE",
		"decimal":   "DECIMAL",
		"bool":      "BOOLEAN",
		"date":      "DATE",
		"datetime":  "DATETIME",
		"timestamp": "TIMESTAMP",
		"json":      "JSON",
		"blob":      "BLOB",
	}
	if mapped, ok := mappings[genericType]; ok {
		return mapped
	}
	return genericType
}

func (d *MySQLDialect) ConvertValue(value interface{}, targetType string) (interface{}, error) {
	return value, nil
}

func (d *MySQLDialect) SupportsReturning() bool {
	return false
}

func (d *MySQLDialect) SupportsUpsert() bool {
	return true // ON DUPLICATE KEY UPDATE
}

func (d *MySQLDialect) SupportsCTE() bool {
	return true // MySQL 8.0+
}

func (d *MySQLDialect) SupportsJSON() bool {
	return true // MySQL 5.7+
}

func (d *MySQLDialect) SupportsFullTextSearch() bool {
	return true
}

func (d *MySQLDialect) TranslatePagination(query string, limit, offset int) string {
	return query + " " + d.LimitOffset(limit, offset)
}

func (d *MySQLDialect) ConvertJSONQuery(query string) string {
	// MySQL queries don't need conversion - already in MySQL format
	return query
}

func (d *MySQLDialect) TranslateUpsert(table string, columns []string, conflictColumns []string) string {
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = d.QuoteIdentifier(col)
	}

	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = "?"
	}

	updates := make([]string, len(columns))
	for i, col := range columns {
		quoted := d.QuoteIdentifier(col)
		updates[i] = fmt.Sprintf("%s=VALUES(%s)", quoted, quoted)
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s",
		d.QuoteIdentifier(table),
		strings.Join(quotedColumns, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(updates, ", "),
	)
}

// DDL Generation Methods for MySQL

func (d *MySQLDialect) CreateTableDDL(schema *TableSchema) string {
	var ddl strings.Builder
	ddl.WriteString("CREATE TABLE ")
	ddl.WriteString(d.QuoteIdentifier(schema.TableName))
	ddl.WriteString(" (\n")

	// Add columns
	for i, col := range schema.Columns {
		if i > 0 {
			ddl.WriteString(",\n")
		}
		ddl.WriteString("  ")
		ddl.WriteString(d.QuoteIdentifier(col.Name))
		ddl.WriteString(" ")
		ddl.WriteString(d.DataTypeMapping(col.DataType))

		if col.MaxLength != nil && (col.DataType == "string" || col.DataType == "VARCHAR") {
			ddl.WriteString(fmt.Sprintf("(%d)", *col.MaxLength))
		}
		if col.Precision != nil && col.Scale != nil {
			ddl.WriteString(fmt.Sprintf("(%d,%d)", *col.Precision, *col.Scale))
		}

		if !col.IsNullable {
			ddl.WriteString(" NOT NULL")
		}
		if col.DefaultValue != nil {
			ddl.WriteString(" DEFAULT ")
			ddl.WriteString(*col.DefaultValue)
		}
		if col.Comment != "" {
			ddl.WriteString(" COMMENT '")
			ddl.WriteString(strings.ReplaceAll(col.Comment, "'", "''"))
			ddl.WriteString("'")
		}
	}

	// Add primary key
	if len(schema.PrimaryKeys) > 0 {
		ddl.WriteString(",\n  PRIMARY KEY (")
		for i, pk := range schema.PrimaryKeys {
			if i > 0 {
				ddl.WriteString(", ")
			}
			ddl.WriteString(d.QuoteIdentifier(pk))
		}
		ddl.WriteString(")")
	}

	ddl.WriteString("\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")
	return ddl.String()
}

func (d *MySQLDialect) AddColumnDDL(tableName string, column *ColumnInfo) string {
	var ddl strings.Builder
	ddl.WriteString("ALTER TABLE ")
	ddl.WriteString(d.QuoteIdentifier(tableName))
	ddl.WriteString(" ADD COLUMN ")
	ddl.WriteString(d.QuoteIdentifier(column.Name))
	ddl.WriteString(" ")
	ddl.WriteString(d.DataTypeMapping(column.DataType))

	if column.MaxLength != nil {
		ddl.WriteString(fmt.Sprintf("(%d)", *column.MaxLength))
	}
	if !column.IsNullable {
		ddl.WriteString(" NOT NULL")
	}
	if column.DefaultValue != nil {
		ddl.WriteString(" DEFAULT ")
		ddl.WriteString(*column.DefaultValue)
	}
	return ddl.String()
}

func (d *MySQLDialect) DropColumnDDL(tableName, columnName string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s",
		d.QuoteIdentifier(tableName),
		d.QuoteIdentifier(columnName))
}

func (d *MySQLDialect) AlterColumnDDL(tableName string, column *ColumnInfo) string {
	var ddl strings.Builder
	ddl.WriteString("ALTER TABLE ")
	ddl.WriteString(d.QuoteIdentifier(tableName))
	ddl.WriteString(" MODIFY COLUMN ")
	ddl.WriteString(d.QuoteIdentifier(column.Name))
	ddl.WriteString(" ")
	ddl.WriteString(d.DataTypeMapping(column.DataType))

	if column.MaxLength != nil {
		ddl.WriteString(fmt.Sprintf("(%d)", *column.MaxLength))
	}
	if !column.IsNullable {
		ddl.WriteString(" NOT NULL")
	}
	if column.DefaultValue != nil {
		ddl.WriteString(" DEFAULT ")
		ddl.WriteString(*column.DefaultValue)
	}
	return ddl.String()
}

func (d *MySQLDialect) CreateIndexDDL(tableName string, index *IndexInfo) string {
	var ddl strings.Builder
	if index.IsUnique {
		ddl.WriteString("CREATE UNIQUE INDEX ")
	} else {
		ddl.WriteString("CREATE INDEX ")
	}
	ddl.WriteString(d.QuoteIdentifier(index.Name))
	ddl.WriteString(" ON ")
	ddl.WriteString(d.QuoteIdentifier(tableName))
	ddl.WriteString(" (")
	for i, col := range index.Columns {
		if i > 0 {
			ddl.WriteString(", ")
		}
		ddl.WriteString(d.QuoteIdentifier(col))
	}
	ddl.WriteString(")")
	return ddl.String()
}

func (d *MySQLDialect) DropIndexDDL(tableName, indexName string) string {
	return fmt.Sprintf("DROP INDEX %s ON %s",
		d.QuoteIdentifier(indexName),
		d.QuoteIdentifier(tableName))
}

// PostgreSQLDialect implements PostgreSQL-specific SQL operations
type PostgreSQLDialect struct{}

func NewPostgreSQLDialect() *PostgreSQLDialect {
	return &PostgreSQLDialect{}
}

func (d *PostgreSQLDialect) QuoteIdentifier(name string) string {
	return `"` + strings.Replace(name, `"`, `""`, -1) + `"`
}

func (d *PostgreSQLDialect) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n)
}

func (d *PostgreSQLDialect) LimitOffset(limit, offset int) string {
	if offset > 0 {
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}
	return fmt.Sprintf("LIMIT %d", limit)
}

func (d *PostgreSQLDialect) DataTypeMapping(genericType string) string {
	mappings := map[string]string{
		"string":    "VARCHAR(255)",
		"text":      "TEXT",
		"int":       "INTEGER",
		"bigint":    "BIGINT",
		"float":     "REAL",
		"double":    "DOUBLE PRECISION",
		"decimal":   "DECIMAL",
		"bool":      "BOOLEAN",
		"date":      "DATE",
		"datetime":  "TIMESTAMP",
		"timestamp": "TIMESTAMP WITH TIME ZONE",
		"json":      "JSONB",
		"blob":      "BYTEA",
	}
	if mapped, ok := mappings[genericType]; ok {
		return mapped
	}
	return genericType
}

func (d *PostgreSQLDialect) ConvertValue(value interface{}, targetType string) (interface{}, error) {
	return value, nil
}

func (d *PostgreSQLDialect) SupportsReturning() bool {
	return true
}

func (d *PostgreSQLDialect) SupportsUpsert() bool {
	return true // ON CONFLICT
}

func (d *PostgreSQLDialect) SupportsCTE() bool {
	return true
}

func (d *PostgreSQLDialect) SupportsJSON() bool {
	return true // JSONB
}

func (d *PostgreSQLDialect) SupportsFullTextSearch() bool {
	return true
}

func (d *PostgreSQLDialect) TranslatePagination(query string, limit, offset int) string {
	return query + " " + d.LimitOffset(limit, offset)
}

func (d *PostgreSQLDialect) ConvertJSONQuery(query string) string {
	// Convert MySQL JSON_TABLE syntax to PostgreSQL jsonb_array_elements
	// Example conversion:
	// FROM: JSON_TABLE(we.workflow, '$.nodes[*]' COLUMNS (id VARCHAR(100) PATH '$.id', name VARCHAR(255) PATH '$.name')) AS jt
	// TO: LATERAL (SELECT elem->>'id' AS id, elem->>'name' AS name FROM jsonb_array_elements(we.workflow->'nodes') AS elem) AS jt

	// This is a simplified conversion - for complex JSON_TABLE queries, you may need a more sophisticated parser
	if !strings.Contains(query, "JSON_TABLE") {
		return query // No JSON_TABLE to convert
	}

	// For now, return query as-is with a warning comment
	// TODO: Implement full JSON_TABLE to jsonb_array_elements conversion
	// Users should use database-specific queries or store them in a table with database type flag
	return query
}

func (d *PostgreSQLDialect) TranslateUpsert(table string, columns []string, conflictColumns []string) string {
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = d.QuoteIdentifier(col)
	}

	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = d.Placeholder(i + 1)
	}

	updates := make([]string, 0)
	for _, col := range columns {
		// Skip conflict columns in updates
		isConflict := false
		for _, conflictCol := range conflictColumns {
			if col == conflictCol {
				isConflict = true
				break
			}
		}
		if !isConflict {
			quoted := d.QuoteIdentifier(col)
			updates = append(updates, fmt.Sprintf("%s=EXCLUDED.%s", quoted, quoted))
		}
	}

	quotedConflict := make([]string, len(conflictColumns))
	for i, col := range conflictColumns {
		quotedConflict[i] = d.QuoteIdentifier(col)
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
		d.QuoteIdentifier(table),
		strings.Join(quotedColumns, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(quotedConflict, ", "),
		strings.Join(updates, ", "),
	)
}

// DDL Generation Methods for PostgreSQL

func (d *PostgreSQLDialect) CreateTableDDL(schema *TableSchema) string {
	var ddl strings.Builder
	ddl.WriteString("CREATE TABLE ")
	ddl.WriteString(d.QuoteIdentifier(schema.TableName))
	ddl.WriteString(" (\n")

	// Add columns
	for i, col := range schema.Columns {
		if i > 0 {
			ddl.WriteString(",\n")
		}
		ddl.WriteString("  ")
		ddl.WriteString(d.QuoteIdentifier(col.Name))
		ddl.WriteString(" ")
		ddl.WriteString(d.DataTypeMapping(col.DataType))

		if col.MaxLength != nil && (col.DataType == "string" || col.DataType == "VARCHAR") {
			ddl.WriteString(fmt.Sprintf("(%d)", *col.MaxLength))
		}
		if col.Precision != nil && col.Scale != nil {
			ddl.WriteString(fmt.Sprintf("(%d,%d)", *col.Precision, *col.Scale))
		}

		if !col.IsNullable {
			ddl.WriteString(" NOT NULL")
		}
		if col.DefaultValue != nil {
			ddl.WriteString(" DEFAULT ")
			ddl.WriteString(*col.DefaultValue)
		}
	}

	// Add primary key
	if len(schema.PrimaryKeys) > 0 {
		ddl.WriteString(",\n  PRIMARY KEY (")
		for i, pk := range schema.PrimaryKeys {
			if i > 0 {
				ddl.WriteString(", ")
			}
			ddl.WriteString(d.QuoteIdentifier(pk))
		}
		ddl.WriteString(")")
	}

	ddl.WriteString("\n)")
	return ddl.String()
}

func (d *PostgreSQLDialect) AddColumnDDL(tableName string, column *ColumnInfo) string {
	var ddl strings.Builder
	ddl.WriteString("ALTER TABLE ")
	ddl.WriteString(d.QuoteIdentifier(tableName))
	ddl.WriteString(" ADD COLUMN ")
	ddl.WriteString(d.QuoteIdentifier(column.Name))
	ddl.WriteString(" ")
	ddl.WriteString(d.DataTypeMapping(column.DataType))

	if column.MaxLength != nil {
		ddl.WriteString(fmt.Sprintf("(%d)", *column.MaxLength))
	}
	if !column.IsNullable {
		ddl.WriteString(" NOT NULL")
	}
	if column.DefaultValue != nil {
		ddl.WriteString(" DEFAULT ")
		ddl.WriteString(*column.DefaultValue)
	}
	return ddl.String()
}

func (d *PostgreSQLDialect) DropColumnDDL(tableName, columnName string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s",
		d.QuoteIdentifier(tableName),
		d.QuoteIdentifier(columnName))
}

func (d *PostgreSQLDialect) AlterColumnDDL(tableName string, column *ColumnInfo) string {
	// PostgreSQL uses ALTER COLUMN for type changes
	var ddl strings.Builder
	ddl.WriteString("ALTER TABLE ")
	ddl.WriteString(d.QuoteIdentifier(tableName))
	ddl.WriteString(" ALTER COLUMN ")
	ddl.WriteString(d.QuoteIdentifier(column.Name))
	ddl.WriteString(" TYPE ")
	ddl.WriteString(d.DataTypeMapping(column.DataType))

	if column.MaxLength != nil {
		ddl.WriteString(fmt.Sprintf("(%d)", *column.MaxLength))
	}
	return ddl.String()
}

func (d *PostgreSQLDialect) CreateIndexDDL(tableName string, index *IndexInfo) string {
	var ddl strings.Builder
	if index.IsUnique {
		ddl.WriteString("CREATE UNIQUE INDEX ")
	} else {
		ddl.WriteString("CREATE INDEX ")
	}
	ddl.WriteString(d.QuoteIdentifier(index.Name))
	ddl.WriteString(" ON ")
	ddl.WriteString(d.QuoteIdentifier(tableName))
	ddl.WriteString(" (")
	for i, col := range index.Columns {
		if i > 0 {
			ddl.WriteString(", ")
		}
		ddl.WriteString(d.QuoteIdentifier(col))
	}
	ddl.WriteString(")")
	return ddl.String()
}

func (d *PostgreSQLDialect) DropIndexDDL(tableName, indexName string) string {
	// PostgreSQL doesn't need table name for DROP INDEX
	return fmt.Sprintf("DROP INDEX %s", d.QuoteIdentifier(indexName))
}

// MSSQLDialect implements MSSQL/SQL Server-specific SQL operations
type MSSQLDialect struct{}

func NewMSSQLDialect() *MSSQLDialect {
	return &MSSQLDialect{}
}

func (d *MSSQLDialect) QuoteIdentifier(name string) string {
	return "[" + strings.Replace(name, "]", "]]", -1) + "]"
}

func (d *MSSQLDialect) Placeholder(n int) string {
	return fmt.Sprintf("@p%d", n)
}

func (d *MSSQLDialect) LimitOffset(limit, offset int) string {
	if offset > 0 {
		return fmt.Sprintf("OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
	}
	return fmt.Sprintf("OFFSET 0 ROWS FETCH NEXT %d ROWS ONLY", limit)
}

func (d *MSSQLDialect) DataTypeMapping(genericType string) string {
	mappings := map[string]string{
		"string":    "NVARCHAR(255)",
		"text":      "NVARCHAR(MAX)",
		"int":       "INT",
		"bigint":    "BIGINT",
		"float":     "FLOAT",
		"double":    "FLOAT(53)",
		"decimal":   "DECIMAL",
		"bool":      "BIT",
		"date":      "DATE",
		"datetime":  "DATETIME2",
		"timestamp": "DATETIME2",
		"json":      "NVARCHAR(MAX)", // JSON functions available in SQL Server 2016+
		"blob":      "VARBINARY(MAX)",
	}
	if mapped, ok := mappings[genericType]; ok {
		return mapped
	}
	return genericType
}

func (d *MSSQLDialect) ConvertValue(value interface{}, targetType string) (interface{}, error) {
	return value, nil
}

func (d *MSSQLDialect) SupportsReturning() bool {
	return true // OUTPUT clause
}

func (d *MSSQLDialect) SupportsUpsert() bool {
	return true // MERGE
}

func (d *MSSQLDialect) SupportsCTE() bool {
	return true
}

func (d *MSSQLDialect) SupportsJSON() bool {
	return true // SQL Server 2016+
}

func (d *MSSQLDialect) SupportsFullTextSearch() bool {
	return true
}

func (d *MSSQLDialect) TranslatePagination(query string, limit, offset int) string {
	// MSSQL requires ORDER BY for OFFSET/FETCH
	return query + " " + d.LimitOffset(limit, offset)
}

func (d *MSSQLDialect) ConvertJSONQuery(query string) string {
	// MSSQL uses OPENJSON for JSON queries
	// MySQL JSON_TABLE would need to be converted to OPENJSON syntax
	return query // TODO: Implement MySQL JSON_TABLE to MSSQL OPENJSON conversion
}

func (d *MSSQLDialect) TranslateUpsert(table string, columns []string, conflictColumns []string) string {
	// MSSQL uses MERGE statement
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = d.QuoteIdentifier(col)
	}

	// This is a simplified MERGE statement
	return fmt.Sprintf("MERGE %s AS target USING (VALUES (...)) AS source (...) ON ... WHEN MATCHED THEN UPDATE ... WHEN NOT MATCHED THEN INSERT ...",
		d.QuoteIdentifier(table))
}

// OracleDialect implements Oracle-specific SQL operations
type OracleDialect struct{}

func NewOracleDialect() *OracleDialect {
	return &OracleDialect{}
}

func (d *OracleDialect) QuoteIdentifier(name string) string {
	return `"` + strings.Replace(name, `"`, `""`, -1) + `"`
}

func (d *OracleDialect) Placeholder(n int) string {
	return fmt.Sprintf(":%d", n)
}

func (d *OracleDialect) LimitOffset(limit, offset int) string {
	// Oracle uses ROWNUM or OFFSET/FETCH (12c+)
	if offset > 0 {
		return fmt.Sprintf("OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
	}
	return fmt.Sprintf("FETCH NEXT %d ROWS ONLY", limit)
}

func (d *OracleDialect) DataTypeMapping(genericType string) string {
	mappings := map[string]string{
		"string":    "VARCHAR2(255)",
		"text":      "CLOB",
		"int":       "NUMBER(10)",
		"bigint":    "NUMBER(19)",
		"float":     "BINARY_FLOAT",
		"double":    "BINARY_DOUBLE",
		"decimal":   "NUMBER",
		"bool":      "NUMBER(1)",
		"date":      "DATE",
		"datetime":  "TIMESTAMP",
		"timestamp": "TIMESTAMP",
		"json":      "CLOB", // Or JSON type in 21c+
		"blob":      "BLOB",
	}
	if mapped, ok := mappings[genericType]; ok {
		return mapped
	}
	return genericType
}

func (d *OracleDialect) ConvertValue(value interface{}, targetType string) (interface{}, error) {
	return value, nil
}

func (d *OracleDialect) SupportsReturning() bool {
	return true // RETURNING INTO
}

func (d *OracleDialect) SupportsUpsert() bool {
	return true // MERGE
}

func (d *OracleDialect) SupportsCTE() bool {
	return true
}

func (d *OracleDialect) SupportsJSON() bool {
	return true // Oracle 12c+
}

func (d *OracleDialect) SupportsFullTextSearch() bool {
	return true // Oracle Text
}

func (d *OracleDialect) TranslatePagination(query string, limit, offset int) string {
	return query + " " + d.LimitOffset(limit, offset)
}

func (d *OracleDialect) ConvertJSONQuery(query string) string {
	// Oracle uses JSON_TABLE starting from 12c
	// MySQL JSON_TABLE syntax is similar but may need minor adjustments
	return query // TODO: Implement MySQL JSON_TABLE to Oracle JSON_TABLE conversion if needed
}

func (d *OracleDialect) TranslateUpsert(table string, columns []string, conflictColumns []string) string {
	// Oracle uses MERGE statement
	return fmt.Sprintf("MERGE INTO %s ...", d.QuoteIdentifier(table))
}

// DDL Generation Methods for MSSQL

func (d *MSSQLDialect) CreateTableDDL(schema *TableSchema) string {
	var ddl strings.Builder
	ddl.WriteString("CREATE TABLE ")
	ddl.WriteString(d.QuoteIdentifier(schema.TableName))
	ddl.WriteString(" (\n")

	// Add columns
	for i, col := range schema.Columns {
		if i > 0 {
			ddl.WriteString(",\n")
		}
		ddl.WriteString("  ")
		ddl.WriteString(d.QuoteIdentifier(col.Name))
		ddl.WriteString(" ")
		ddl.WriteString(d.DataTypeMapping(col.DataType))

		if col.MaxLength != nil && (col.DataType == "string" || col.DataType == "NVARCHAR") {
			ddl.WriteString(fmt.Sprintf("(%d)", *col.MaxLength))
		}
		if col.Precision != nil && col.Scale != nil {
			ddl.WriteString(fmt.Sprintf("(%d,%d)", *col.Precision, *col.Scale))
		}

		if !col.IsNullable {
			ddl.WriteString(" NOT NULL")
		}
		if col.DefaultValue != nil {
			ddl.WriteString(" DEFAULT ")
			ddl.WriteString(*col.DefaultValue)
		}
	}

	// Add primary key
	if len(schema.PrimaryKeys) > 0 {
		ddl.WriteString(",\n  PRIMARY KEY (")
		for i, pk := range schema.PrimaryKeys {
			if i > 0 {
				ddl.WriteString(", ")
			}
			ddl.WriteString(d.QuoteIdentifier(pk))
		}
		ddl.WriteString(")")
	}

	ddl.WriteString("\n)")
	return ddl.String()
}

func (d *MSSQLDialect) AddColumnDDL(tableName string, column *ColumnInfo) string {
	var ddl strings.Builder
	ddl.WriteString("ALTER TABLE ")
	ddl.WriteString(d.QuoteIdentifier(tableName))
	ddl.WriteString(" ADD ")
	ddl.WriteString(d.QuoteIdentifier(column.Name))
	ddl.WriteString(" ")
	ddl.WriteString(d.DataTypeMapping(column.DataType))

	if column.MaxLength != nil {
		ddl.WriteString(fmt.Sprintf("(%d)", *column.MaxLength))
	}
	if !column.IsNullable {
		ddl.WriteString(" NOT NULL")
	}
	if column.DefaultValue != nil {
		ddl.WriteString(" DEFAULT ")
		ddl.WriteString(*column.DefaultValue)
	}
	return ddl.String()
}

func (d *MSSQLDialect) DropColumnDDL(tableName, columnName string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s",
		d.QuoteIdentifier(tableName),
		d.QuoteIdentifier(columnName))
}

func (d *MSSQLDialect) AlterColumnDDL(tableName string, column *ColumnInfo) string {
	var ddl strings.Builder
	ddl.WriteString("ALTER TABLE ")
	ddl.WriteString(d.QuoteIdentifier(tableName))
	ddl.WriteString(" ALTER COLUMN ")
	ddl.WriteString(d.QuoteIdentifier(column.Name))
	ddl.WriteString(" ")
	ddl.WriteString(d.DataTypeMapping(column.DataType))

	if column.MaxLength != nil {
		ddl.WriteString(fmt.Sprintf("(%d)", *column.MaxLength))
	}
	if !column.IsNullable {
		ddl.WriteString(" NOT NULL")
	}
	return ddl.String()
}

func (d *MSSQLDialect) CreateIndexDDL(tableName string, index *IndexInfo) string {
	var ddl strings.Builder
	if index.IsUnique {
		ddl.WriteString("CREATE UNIQUE INDEX ")
	} else {
		ddl.WriteString("CREATE INDEX ")
	}
	ddl.WriteString(d.QuoteIdentifier(index.Name))
	ddl.WriteString(" ON ")
	ddl.WriteString(d.QuoteIdentifier(tableName))
	ddl.WriteString(" (")
	for i, col := range index.Columns {
		if i > 0 {
			ddl.WriteString(", ")
		}
		ddl.WriteString(d.QuoteIdentifier(col))
	}
	ddl.WriteString(")")
	return ddl.String()
}

func (d *MSSQLDialect) DropIndexDDL(tableName, indexName string) string {
	return fmt.Sprintf("DROP INDEX %s ON %s",
		d.QuoteIdentifier(indexName),
		d.QuoteIdentifier(tableName))
}

// DDL Generation Methods for Oracle

func (d *OracleDialect) CreateTableDDL(schema *TableSchema) string {
	var ddl strings.Builder
	ddl.WriteString("CREATE TABLE ")
	ddl.WriteString(d.QuoteIdentifier(schema.TableName))
	ddl.WriteString(" (\n")

	// Add columns
	for i, col := range schema.Columns {
		if i > 0 {
			ddl.WriteString(",\n")
		}
		ddl.WriteString("  ")
		ddl.WriteString(d.QuoteIdentifier(col.Name))
		ddl.WriteString(" ")
		ddl.WriteString(d.DataTypeMapping(col.DataType))

		if col.MaxLength != nil && (col.DataType == "string" || col.DataType == "VARCHAR2") {
			ddl.WriteString(fmt.Sprintf("(%d)", *col.MaxLength))
		}
		if col.Precision != nil && col.Scale != nil {
			ddl.WriteString(fmt.Sprintf("(%d,%d)", *col.Precision, *col.Scale))
		}

		if !col.IsNullable {
			ddl.WriteString(" NOT NULL")
		}
		if col.DefaultValue != nil {
			ddl.WriteString(" DEFAULT ")
			ddl.WriteString(*col.DefaultValue)
		}
	}

	// Add primary key
	if len(schema.PrimaryKeys) > 0 {
		ddl.WriteString(",\n  PRIMARY KEY (")
		for i, pk := range schema.PrimaryKeys {
			if i > 0 {
				ddl.WriteString(", ")
			}
			ddl.WriteString(d.QuoteIdentifier(pk))
		}
		ddl.WriteString(")")
	}

	ddl.WriteString("\n)")
	return ddl.String()
}

func (d *OracleDialect) AddColumnDDL(tableName string, column *ColumnInfo) string {
	var ddl strings.Builder
	ddl.WriteString("ALTER TABLE ")
	ddl.WriteString(d.QuoteIdentifier(tableName))
	ddl.WriteString(" ADD ")
	ddl.WriteString(d.QuoteIdentifier(column.Name))
	ddl.WriteString(" ")
	ddl.WriteString(d.DataTypeMapping(column.DataType))

	if column.MaxLength != nil {
		ddl.WriteString(fmt.Sprintf("(%d)", *column.MaxLength))
	}
	if !column.IsNullable {
		ddl.WriteString(" NOT NULL")
	}
	if column.DefaultValue != nil {
		ddl.WriteString(" DEFAULT ")
		ddl.WriteString(*column.DefaultValue)
	}
	return ddl.String()
}

func (d *OracleDialect) DropColumnDDL(tableName, columnName string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s",
		d.QuoteIdentifier(tableName),
		d.QuoteIdentifier(columnName))
}

func (d *OracleDialect) AlterColumnDDL(tableName string, column *ColumnInfo) string {
	var ddl strings.Builder
	ddl.WriteString("ALTER TABLE ")
	ddl.WriteString(d.QuoteIdentifier(tableName))
	ddl.WriteString(" MODIFY ")
	ddl.WriteString(d.QuoteIdentifier(column.Name))
	ddl.WriteString(" ")
	ddl.WriteString(d.DataTypeMapping(column.DataType))

	if column.MaxLength != nil {
		ddl.WriteString(fmt.Sprintf("(%d)", *column.MaxLength))
	}
	if !column.IsNullable {
		ddl.WriteString(" NOT NULL")
	}
	return ddl.String()
}

func (d *OracleDialect) CreateIndexDDL(tableName string, index *IndexInfo) string {
	var ddl strings.Builder
	if index.IsUnique {
		ddl.WriteString("CREATE UNIQUE INDEX ")
	} else {
		ddl.WriteString("CREATE INDEX ")
	}
	ddl.WriteString(d.QuoteIdentifier(index.Name))
	ddl.WriteString(" ON ")
	ddl.WriteString(d.QuoteIdentifier(tableName))
	ddl.WriteString(" (")
	for i, col := range index.Columns {
		if i > 0 {
			ddl.WriteString(", ")
		}
		ddl.WriteString(d.QuoteIdentifier(col))
	}
	ddl.WriteString(")")
	return ddl.String()
}

func (d *OracleDialect) DropIndexDDL(tableName, indexName string) string {
	// Oracle doesn't need table name for DROP INDEX
	return fmt.Sprintf("DROP INDEX %s", d.QuoteIdentifier(indexName))
}
