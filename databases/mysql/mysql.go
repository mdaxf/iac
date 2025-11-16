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

package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/logger"
)

// init registers the MySQL adapter with the global factory
func init() {
	dbconn.GetFactory().RegisterDriver(dbconn.DBTypeMySQL, func(config *dbconn.DBConfig) (dbconn.RelationalDB, error) {
		return NewMySQLAdapter(config), nil
	})
}

// MySQLAdapter is a MySQL-specific database adapter
type MySQLAdapter struct {
	*dbconn.GenericSQLDB
	config    *dbconn.DBConfig
	db        *sql.DB
	dialect   *dbconn.MySQLDialect
	connected bool
	monitor   *MySQLMonitor
}

// NewMySQLAdapter creates a new MySQL adapter
func NewMySQLAdapter(config *dbconn.DBConfig) *MySQLAdapter {
	return &MySQLAdapter{
		config:  config,
		dialect: dbconn.NewMySQLDialect(),
	}
}

// Connect establishes connection to MySQL database
func (m *MySQLAdapter) Connect(config *dbconn.DBConfig) error {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MySQLAdapter"}

	// Build MySQL connection string
	connStr := buildMySQLConnString(config)

	// Open database connection
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return dbconn.NewDatabaseError("connect", err, dbconn.DBTypeMySQL, "")
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
		return dbconn.NewDatabaseError("ping", err, dbconn.DBTypeMySQL, "")
	}

	m.db = db
	m.connected = true

	// Initialize monitor
	m.monitor = NewMySQLMonitor(db, config)

	iLog.Info(fmt.Sprintf("Connected to MySQL database at %s:%d/%s",
		config.Host, config.Port, config.Database))

	return nil
}

// ListTables returns all tables in the database
func (m *MySQLAdapter) ListTables(ctx context.Context, schema string) ([]string, error) {
	if !m.connected || m.db == nil {
		return nil, dbconn.ErrNotConnected
	}

	dbName := schema
	if dbName == "" {
		dbName = m.config.Database
	}

	query := `
		SELECT TABLE_NAME
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`

	rows, err := m.db.QueryContext(ctx, query, dbName)
	if err != nil {
		return nil, dbconn.NewDatabaseError("list_tables", err, dbconn.DBTypeMySQL, query)
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
func (m *MySQLAdapter) TableExists(ctx context.Context, schema, tableName string) (bool, error) {
	if !m.connected || m.db == nil {
		return false, dbconn.ErrNotConnected
	}

	dbName := schema
	if dbName == "" {
		dbName = m.config.Database
	}

	query := `
		SELECT COUNT(*)
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		AND TABLE_NAME = ?
		AND TABLE_TYPE = 'BASE TABLE'
	`

	var count int
	err := m.db.QueryRowContext(ctx, query, dbName, tableName).Scan(&count)
	if err != nil {
		return false, dbconn.NewDatabaseError("table_exists", err, dbconn.DBTypeMySQL, query)
	}

	return count > 0, nil
}

// GetTableSchema retrieves the schema for a table
func (m *MySQLAdapter) GetTableSchema(ctx context.Context, schema, tableName string) (*dbconn.TableSchema, error) {
	if !m.connected || m.db == nil {
		return nil, dbconn.ErrNotConnected
	}

	dbName := schema
	if dbName == "" {
		dbName = m.config.Database
	}

	// Check if table exists
	exists, err := m.TableExists(ctx, dbName, tableName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, dbconn.ErrTableNotFound
	}

	tableSchema := &dbconn.TableSchema{
		Schema:    dbName,
		TableName: tableName,
		Columns:   make([]dbconn.ColumnInfo, 0),
		Indexes:   make([]dbconn.IndexInfo, 0),
	}

	// Get columns
	columnQuery := `
		SELECT
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			CHARACTER_MAXIMUM_LENGTH,
			NUMERIC_PRECISION,
			NUMERIC_SCALE,
			COLUMN_KEY,
			COLUMN_COMMENT
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := m.db.QueryContext(ctx, columnQuery, dbName, tableName)
	if err != nil {
		return nil, dbconn.NewDatabaseError("get_table_schema", err, dbconn.DBTypeMySQL, columnQuery)
	}
	defer rows.Close()

	primaryKeys := make([]string, 0)

	for rows.Next() {
		var (
			colName     string
			dataType    string
			isNullable  string
			defaultVal  sql.NullString
			maxLength   sql.NullInt64
			precision   sql.NullInt64
			scale       sql.NullInt64
			columnKey   string
			comment     string
		)

		err := rows.Scan(&colName, &dataType, &isNullable, &defaultVal, &maxLength, &precision, &scale, &columnKey, &comment)
		if err != nil {
			return nil, err
		}

		colInfo := dbconn.ColumnInfo{
			Name:         colName,
			DataType:     dataType,
			IsNullable:   isNullable == "YES",
			IsPrimaryKey: columnKey == "PRI",
			IsForeignKey: columnKey == "MUL",
			IsUnique:     columnKey == "UNI",
			Comment:      comment,
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

		if colInfo.IsPrimaryKey {
			primaryKeys = append(primaryKeys, colName)
		}
	}

	tableSchema.PrimaryKeys = primaryKeys

	// Get indexes
	indexQuery := `
		SELECT
			INDEX_NAME,
			COLUMN_NAME,
			NON_UNIQUE,
			INDEX_TYPE
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`

	indexRows, err := m.db.QueryContext(ctx, indexQuery, dbName, tableName)
	if err != nil {
		return nil, dbconn.NewDatabaseError("get_indexes", err, dbconn.DBTypeMySQL, indexQuery)
	}
	defer indexRows.Close()

	indexMap := make(map[string]*dbconn.IndexInfo)

	for indexRows.Next() {
		var (
			indexName  string
			columnName string
			nonUnique  int
			indexType  string
		)

		if err := indexRows.Scan(&indexName, &columnName, &nonUnique, &indexType); err != nil {
			return nil, err
		}

		if idx, exists := indexMap[indexName]; exists {
			idx.Columns = append(idx.Columns, columnName)
		} else {
			indexMap[indexName] = &dbconn.IndexInfo{
				Name:     indexName,
				Columns:  []string{columnName},
				IsUnique: nonUnique == 0,
				Type:     indexType,
			}
		}
	}

	for _, idx := range indexMap {
		tableSchema.Indexes = append(tableSchema.Indexes, *idx)
	}

	return tableSchema, nil
}

// GetServerVersion returns the MySQL server version
func (m *MySQLAdapter) GetServerVersion(ctx context.Context) (string, error) {
	if !m.connected || m.db == nil {
		return "", dbconn.ErrNotConnected
	}

	var version string
	err := m.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
	if err != nil {
		return "", err
	}

	return version, nil
}

// Optimize optimizes a table
func (m *MySQLAdapter) Optimize(ctx context.Context, tableName string) error {
	if !m.connected || m.db == nil {
		return dbconn.ErrNotConnected
	}

	query := fmt.Sprintf("OPTIMIZE TABLE %s", m.dialect.QuoteIdentifier(tableName))
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// Analyze analyzes a table
func (m *MySQLAdapter) Analyze(ctx context.Context, tableName string) error {
	if !m.connected || m.db == nil {
		return dbconn.ErrNotConnected
	}

	query := fmt.Sprintf("ANALYZE TABLE %s", m.dialect.QuoteIdentifier(tableName))
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// GetMonitor returns the MySQL monitor
func (m *MySQLAdapter) GetMonitor() *MySQLMonitor {
	return m.monitor
}

// Helper function to build MySQL connection string
func buildMySQLConnString(config *dbconn.DBConfig) string {
	params := []string{
		"charset=utf8mb4",
		"parseTime=True",
		"loc=Local",
	}

	// SSL/TLS configuration
	if config.SSLMode != "" && config.SSLMode != "disable" {
		params = append(params, fmt.Sprintf("tls=%s", config.SSLMode))
	}

	// Additional options
	for k, v := range config.Options {
		params = append(params, fmt.Sprintf("%s=%s", k, v))
	}

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	if len(params) > 0 {
		connStr += "?" + strings.Join(params, "&")
	}

	return connStr
}

// Close closes the MySQL connection
func (m *MySQLAdapter) Close() error {
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
func (m *MySQLAdapter) DB() *sql.DB {
	return m.db
}

// GetDialect returns the MySQL dialect
func (m *MySQLAdapter) GetDialect() dbconn.Dialect {
	return m.dialect
}

// GetType returns the database type
func (m *MySQLAdapter) GetType() dbconn.DBType {
	return dbconn.DBTypeMySQL
}

// IsConnected checks if connected
func (m *MySQLAdapter) IsConnected() bool {
	return m.connected && m.db != nil
}

// Ping pings the database
func (m *MySQLAdapter) Ping() error {
	if m.db == nil {
		return dbconn.ErrNotConnected
	}
	return m.db.Ping()
}

// Stats returns connection statistics
func (m *MySQLAdapter) Stats() sql.DBStats {
	if m.db == nil {
		return sql.DBStats{}
	}
	return m.db.Stats()
}

// BeginTx starts a transaction
func (m *MySQLAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if m.db == nil {
		return nil, dbconn.ErrNotConnected
	}
	return m.db.BeginTx(ctx, opts)
}

// Query executes a query
func (m *MySQLAdapter) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.db == nil {
		return nil, dbconn.ErrNotConnected
	}
	return m.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (m *MySQLAdapter) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.db == nil {
		return nil
	}
	return m.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query without returning rows
func (m *MySQLAdapter) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.db == nil {
		return nil, dbconn.ErrNotConnected
	}
	return m.db.ExecContext(ctx, query, args...)
}

// SupportsFeature checks if a feature is supported
func (m *MySQLAdapter) SupportsFeature(feature dbconn.Feature) bool {
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
