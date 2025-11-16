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

package mssql

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/logger"
)

// init registers the MSSQL adapter with the global factory
func init() {
	dbconn.GetFactory().RegisterDriver(dbconn.DBTypeMSSQL, func(config *dbconn.DBConfig) (dbconn.RelationalDB, error) {
		return NewMSSQLAdapter(config), nil
	})
}

// MSSQLAdapter is a MSSQL-specific database adapter
type MSSQLAdapter struct {
	config    *dbconn.DBConfig
	db        *sql.DB
	dialect   *dbconn.MSSQLDialect
	connected bool
	monitor   *MSSQLMonitor
}

// NewMSSQLAdapter creates a new MSSQL adapter
func NewMSSQLAdapter(config *dbconn.DBConfig) *MSSQLAdapter {
	return &MSSQLAdapter{
		config:  config,
		dialect: dbconn.NewMSSQLDialect(),
	}
}

// Connect establishes connection to MSSQL database
func (m *MSSQLAdapter) Connect(config *dbconn.DBConfig) error {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MSSQLAdapter"}

	// Build MSSQL connection string
	connStr := buildMSSQLConnString(config)

	// Open database connection
	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return dbconn.NewDatabaseError("connect", err, dbconn.DBTypeMSSQL, "")
	}

	// Configure connection pool
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetMaxOpenConns(config.MaxOpenConns)
	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	} else {
		db.SetConnMaxLifetime(3 * time.Hour) // Default 3 hours
	}
	if config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	} else {
		db.SetConnMaxIdleTime(10 * time.Minute) // Default 10 minutes
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return dbconn.NewDatabaseError("ping", err, dbconn.DBTypeMSSQL, "")
	}

	m.db = db
	m.connected = true

	// Initialize monitor
	m.monitor = NewMSSQLMonitor(db, config)

	iLog.Info(fmt.Sprintf("Connected to MSSQL database at %s:%d/%s",
		config.Host, config.Port, config.Database))

	return nil
}

// ListTables returns all tables in the database
func (m *MSSQLAdapter) ListTables(ctx context.Context, schema string) ([]string, error) {
	if !m.connected || m.db == nil {
		return nil, dbconn.ErrNotConnected
	}

	schemaName := schema
	if schemaName == "" {
		schemaName = m.config.Schema
		if schemaName == "" {
			schemaName = "dbo"
		}
	}

	query := `
		SELECT TABLE_NAME
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = @p1
		AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`

	rows, err := m.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, dbconn.NewDatabaseError("list_tables", err, dbconn.DBTypeMSSQL, query)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}

// TableExists checks if a table exists
func (m *MSSQLAdapter) TableExists(ctx context.Context, schema, tableName string) (bool, error) {
	if !m.connected || m.db == nil {
		return false, dbconn.ErrNotConnected
	}

	schemaName := schema
	if schemaName == "" {
		schemaName = m.config.Schema
		if schemaName == "" {
			schemaName = "dbo"
		}
	}

	query := `
		SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = @p1
		AND TABLE_NAME = @p2
		AND TABLE_TYPE = 'BASE TABLE'
	`

	var count int
	err := m.db.QueryRowContext(ctx, query, schemaName, tableName).Scan(&count)
	if err != nil {
		return false, dbconn.NewDatabaseError("table_exists", err, dbconn.DBTypeMSSQL, query)
	}

	return count > 0, nil
}

// GetTableSchema retrieves the schema for a table
func (m *MSSQLAdapter) GetTableSchema(ctx context.Context, schema, tableName string) (*dbconn.TableSchema, error) {
	if !m.connected || m.db == nil {
		return nil, dbconn.ErrNotConnected
	}

	schemaName := schema
	if schemaName == "" {
		schemaName = m.config.Schema
		if schemaName == "" {
			schemaName = "dbo"
		}
	}

	// Check if table exists
	exists, err := m.TableExists(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, dbconn.ErrTableNotFound
	}

	tableSchema := &dbconn.TableSchema{
		Schema:    schemaName,
		TableName: tableName,
		Columns:   make([]dbconn.ColumnInfo, 0),
		Indexes:   make([]dbconn.IndexInfo, 0),
	}

	// Get columns
	columnQuery := `
		SELECT
			c.COLUMN_NAME,
			c.DATA_TYPE,
			c.IS_NULLABLE,
			c.COLUMN_DEFAULT,
			c.CHARACTER_MAXIMUM_LENGTH,
			c.NUMERIC_PRECISION,
			c.NUMERIC_SCALE
		FROM INFORMATION_SCHEMA.COLUMNS c
		WHERE c.TABLE_SCHEMA = @p1 AND c.TABLE_NAME = @p2
		ORDER BY c.ORDINAL_POSITION
	`

	rows, err := m.db.QueryContext(ctx, columnQuery, schemaName, tableName)
	if err != nil {
		return nil, dbconn.NewDatabaseError("get_table_schema", err, dbconn.DBTypeMSSQL, columnQuery)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			colName    string
			dataType   string
			isNullable string
			defaultVal sql.NullString
			maxLength  sql.NullInt64
			precision  sql.NullInt64
			scale      sql.NullInt64
		)

		err := rows.Scan(&colName, &dataType, &isNullable, &defaultVal, &maxLength, &precision, &scale)
		if err != nil {
			return nil, err
		}

		colInfo := dbconn.ColumnInfo{
			Name:       colName,
			DataType:   dataType,
			IsNullable: isNullable == "YES",
		}

		if defaultVal.Valid {
			colInfo.DefaultValue = &defaultVal.String
		}
		if maxLength.Valid {
			maxLen := int(maxLength.Int64)
			colInfo.MaxLength = &maxLen
		}
		if precision.Valid {
			prec := int(precision.Int64)
			colInfo.Precision = &prec
		}
		if scale.Valid {
			sc := int(scale.Int64)
			colInfo.Scale = &sc
		}

		tableSchema.Columns = append(tableSchema.Columns, colInfo)
	}

	// Get primary keys
	pkQuery := `
		SELECT kcu.COLUMN_NAME
		FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
		JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
			ON tc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
			AND tc.TABLE_SCHEMA = kcu.TABLE_SCHEMA
		WHERE tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
		AND tc.TABLE_SCHEMA = @p1
		AND tc.TABLE_NAME = @p2
		ORDER BY kcu.ORDINAL_POSITION
	`

	pkRows, err := m.db.QueryContext(ctx, pkQuery, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer pkRows.Close()

	primaryKeys := make([]string, 0)
	pkMap := make(map[string]bool)

	for pkRows.Next() {
		var colName string
		if err := pkRows.Scan(&colName); err != nil {
			return nil, err
		}
		primaryKeys = append(primaryKeys, colName)
		pkMap[colName] = true
	}

	tableSchema.PrimaryKeys = primaryKeys

	// Update column info with PK status
	for i := range tableSchema.Columns {
		if pkMap[tableSchema.Columns[i].Name] {
			tableSchema.Columns[i].IsPrimaryKey = true
		}
	}

	// Get indexes using sys.indexes and sys.index_columns
	indexQuery := `
		SELECT
			i.name as index_name,
			c.name as column_name,
			i.is_unique,
			i.type_desc
		FROM sys.indexes i
		JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
		JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		JOIN sys.tables t ON i.object_id = t.object_id
		JOIN sys.schemas s ON t.schema_id = s.schema_id
		WHERE s.name = @p1
		AND t.name = @p2
		AND i.is_primary_key = 0
		ORDER BY i.name, ic.key_ordinal
	`

	indexRows, err := m.db.QueryContext(ctx, indexQuery, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer indexRows.Close()

	indexMap := make(map[string]*dbconn.IndexInfo)

	for indexRows.Next() {
		var (
			indexName  string
			columnName string
			isUnique   bool
			indexType  string
		)

		if err := indexRows.Scan(&indexName, &columnName, &isUnique, &indexType); err != nil {
			return nil, err
		}

		if idx, exists := indexMap[indexName]; exists {
			idx.Columns = append(idx.Columns, columnName)
		} else {
			indexMap[indexName] = &dbconn.IndexInfo{
				Name:     indexName,
				Columns:  []string{columnName},
				IsUnique: isUnique,
				Type:     indexType,
			}
		}
	}

	for _, idx := range indexMap {
		tableSchema.Indexes = append(tableSchema.Indexes, *idx)
	}

	return tableSchema, nil
}

// GetServerVersion returns the MSSQL server version
func (m *MSSQLAdapter) GetServerVersion(ctx context.Context) (string, error) {
	if !m.connected || m.db == nil {
		return "", dbconn.ErrNotConnected
	}

	var version string
	err := m.db.QueryRowContext(ctx, "SELECT @@VERSION").Scan(&version)
	if err != nil {
		return "", err
	}

	return version, nil
}

// UpdateStatistics updates statistics for a table
func (m *MSSQLAdapter) UpdateStatistics(ctx context.Context, tableName string) error {
	if !m.connected || m.db == nil {
		return dbconn.ErrNotConnected
	}

	query := fmt.Sprintf("UPDATE STATISTICS %s", m.dialect.QuoteIdentifier(tableName))
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// RebuildIndexes rebuilds all indexes on a table
func (m *MSSQLAdapter) RebuildIndexes(ctx context.Context, tableName string) error {
	if !m.connected || m.db == nil {
		return dbconn.ErrNotConnected
	}

	query := fmt.Sprintf("ALTER INDEX ALL ON %s REBUILD", m.dialect.QuoteIdentifier(tableName))
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// GetMonitor returns the MSSQL monitor
func (m *MSSQLAdapter) GetMonitor() *MSSQLMonitor {
	return m.monitor
}

// Helper function to build MSSQL connection string
func buildMSSQLConnString(config *dbconn.DBConfig) string {
	params := url.Values{}
	params.Add("database", config.Database)

	// Connection timeout
	if config.ConnTimeout > 0 {
		params.Add("connection timeout", fmt.Sprintf("%d", config.ConnTimeout))
	}

	// Encryption
	if config.SSLMode != "" && config.SSLMode != "disable" {
		params.Add("encrypt", "true")
	} else {
		params.Add("encrypt", "false")
	}

	// Trust server certificate (for development)
	params.Add("TrustServerCertificate", "true")

	// Application name
	params.Add("app name", "IAC")

	// Additional options
	for k, v := range config.Options {
		params.Add(k, v)
	}

	connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%d?%s",
		url.QueryEscape(config.Username),
		url.QueryEscape(config.Password),
		config.Host,
		config.Port,
		params.Encode(),
	)

	return connStr
}

// Close closes the MSSQL connection
func (m *MSSQLAdapter) Close() error {
	if m.monitor != nil {
		m.monitor.Stop()
	}

	if m.db != nil {
		err := m.db.Close()
		m.connected = false
		return err
	}

	return nil
}

// DB returns the underlying sql.DB
func (m *MSSQLAdapter) DB() *sql.DB {
	return m.db
}

// GetDialect returns the MSSQL dialect
func (m *MSSQLAdapter) GetDialect() dbconn.Dialect {
	return m.dialect
}

// GetType returns the database type
func (m *MSSQLAdapter) GetType() dbconn.DBType {
	return dbconn.DBTypeMSSQL
}

// IsConnected checks if connected
func (m *MSSQLAdapter) IsConnected() bool {
	return m.connected && m.db != nil
}

// Ping pings the database
func (m *MSSQLAdapter) Ping() error {
	if m.db == nil {
		return dbconn.ErrNotConnected
	}
	return m.db.Ping()
}

// Stats returns connection statistics
func (m *MSSQLAdapter) Stats() sql.DBStats {
	if m.db == nil {
		return sql.DBStats{}
	}
	return m.db.Stats()
}

// BeginTx starts a transaction
func (m *MSSQLAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if m.db == nil {
		return nil, dbconn.ErrNotConnected
	}
	return m.db.BeginTx(ctx, opts)
}

// Query executes a query
func (m *MSSQLAdapter) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.db == nil {
		return nil, dbconn.ErrNotConnected
	}
	return m.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (m *MSSQLAdapter) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.db == nil {
		return nil
	}
	return m.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query without returning rows
func (m *MSSQLAdapter) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.db == nil {
		return nil, dbconn.ErrNotConnected
	}
	return m.db.ExecContext(ctx, query, args...)
}

// SupportsFeature checks if a feature is supported
func (m *MSSQLAdapter) SupportsFeature(feature dbconn.Feature) bool {
	if m.dialect == nil {
		return false
	}

	switch feature {
	case dbconn.FeatureCTE:
		return m.dialect.SupportsCTE()
	case dbconn.FeatureJSON:
		return m.dialect.SupportsJSON()
	case dbconn.FeatureFullTextSearch:
		return m.dialect.SupportsFullTextSearch()
	case dbconn.FeatureReturning:
		return m.dialect.SupportsReturning()
	case dbconn.FeatureUpsert:
		return m.dialect.SupportsUpsert()
	default:
		return false
	}
}
