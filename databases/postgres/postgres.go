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

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/logger"
)

// init registers the PostgreSQL adapter with the global factory
func init() {
	dbconn.GetFactory().RegisterDriver(dbconn.DBTypePostgreSQL, func(config *dbconn.DBConfig) (dbconn.RelationalDB, error) {
		return NewPostgreSQLAdapter(config), nil
	})
}

// PostgreSQLAdapter is a PostgreSQL-specific database adapter
type PostgreSQLAdapter struct {
	config    *dbconn.DBConfig
	db        *sql.DB
	dialect   *dbconn.PostgreSQLDialect
	connected bool
	monitor   *PostgreSQLMonitor
}

// NewPostgreSQLAdapter creates a new PostgreSQL adapter
func NewPostgreSQLAdapter(config *dbconn.DBConfig) *PostgreSQLAdapter {
	return &PostgreSQLAdapter{
		config:  config,
		dialect: dbconn.NewPostgreSQLDialect(),
	}
}

// Connect establishes connection to PostgreSQL database
func (p *PostgreSQLAdapter) Connect(config *dbconn.DBConfig) error {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "PostgreSQLAdapter"}

	// Build PostgreSQL connection string
	connStr := buildPostgreSQLConnString(config)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return dbconn.NewDatabaseError("connect", err, dbconn.DBTypePostgreSQL, "")
	}

	// Configure connection pool
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetMaxOpenConns(config.MaxOpenConns)
	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	} else {
		db.SetConnMaxLifetime(3 * time.Hour)
	}
	if config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	} else {
		db.SetConnMaxIdleTime(10 * time.Minute)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return dbconn.NewDatabaseError("ping", err, dbconn.DBTypePostgreSQL, "")
	}

	p.db = db
	p.connected = true

	// Initialize monitor
	p.monitor = NewPostgreSQLMonitor(db, config)

	iLog.Info(fmt.Sprintf("Connected to PostgreSQL database at %s:%d/%s",
		config.Host, config.Port, config.Database))

	return nil
}

// ListTables returns all tables in the schema
func (p *PostgreSQLAdapter) ListTables(ctx context.Context, schema string) ([]string, error) {
	if !p.connected || p.db == nil {
		return nil, dbconn.ErrNotConnected
	}

	schemaName := schema
	if schemaName == "" {
		schemaName = p.config.Schema
		if schemaName == "" {
			schemaName = "public"
		}
	}

	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := p.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, dbconn.NewDatabaseError("list_tables", err, dbconn.DBTypePostgreSQL, query)
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
func (p *PostgreSQLAdapter) TableExists(ctx context.Context, schema, tableName string) (bool, error) {
	if !p.connected || p.db == nil {
		return false, dbconn.ErrNotConnected
	}

	schemaName := schema
	if schemaName == "" {
		schemaName = p.config.Schema
		if schemaName == "" {
			schemaName = "public"
		}
	}

	query := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = $1
		AND table_name = $2
		AND table_type = 'BASE TABLE'
	`

	var count int
	err := p.db.QueryRowContext(ctx, query, schemaName, tableName).Scan(&count)
	if err != nil {
		return false, dbconn.NewDatabaseError("table_exists", err, dbconn.DBTypePostgreSQL, query)
	}

	return count > 0, nil
}

// GetTableSchema retrieves the schema for a table
func (p *PostgreSQLAdapter) GetTableSchema(ctx context.Context, schema, tableName string) (*dbconn.TableSchema, error) {
	if !p.connected || p.db == nil {
		return nil, dbconn.ErrNotConnected
	}

	schemaName := schema
	if schemaName == "" {
		schemaName = p.config.Schema
		if schemaName == "" {
			schemaName = "public"
		}
	}

	// Check if table exists
	exists, err := p.TableExists(ctx, schemaName, tableName)
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
			column_name,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length,
			numeric_precision,
			numeric_scale
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`

	rows, err := p.db.QueryContext(ctx, columnQuery, schemaName, tableName)
	if err != nil {
		return nil, dbconn.NewDatabaseError("get_table_schema", err, dbconn.DBTypePostgreSQL, columnQuery)
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
		SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.constraint_type = 'PRIMARY KEY'
		AND tc.table_schema = $1
		AND tc.table_name = $2
		ORDER BY kcu.ordinal_position
	`

	pkRows, err := p.db.QueryContext(ctx, pkQuery, schemaName, tableName)
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

	// Get indexes
	indexQuery := `
		SELECT
			i.relname as index_name,
			a.attname as column_name,
			ix.indisunique as is_unique
		FROM pg_class t
		JOIN pg_index ix ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		JOIN pg_namespace n ON n.oid = t.relnamespace
		WHERE n.nspname = $1
		AND t.relname = $2
		AND t.relkind = 'r'
		ORDER BY i.relname, a.attnum
	`

	indexRows, err := p.db.QueryContext(ctx, indexQuery, schemaName, tableName)
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
		)

		if err := indexRows.Scan(&indexName, &columnName, &isUnique); err != nil {
			return nil, err
		}

		if idx, exists := indexMap[indexName]; exists {
			idx.Columns = append(idx.Columns, columnName)
		} else {
			indexMap[indexName] = &dbconn.IndexInfo{
				Name:     indexName,
				Columns:  []string{columnName},
				IsUnique: isUnique,
				Type:     "btree", // PostgreSQL default
			}
		}
	}

	for _, idx := range indexMap {
		tableSchema.Indexes = append(tableSchema.Indexes, *idx)
	}

	return tableSchema, nil
}

// GetServerVersion returns the PostgreSQL server version
func (p *PostgreSQLAdapter) GetServerVersion(ctx context.Context) (string, error) {
	if !p.connected || p.db == nil {
		return "", dbconn.ErrNotConnected
	}

	var version string
	err := p.db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	return version, err
}

// Vacuum runs VACUUM on a table
func (p *PostgreSQLAdapter) Vacuum(ctx context.Context, tableName string, full bool) error {
	if !p.connected || p.db == nil {
		return dbconn.ErrNotConnected
	}

	query := "VACUUM"
	if full {
		query += " FULL"
	}
	query += " " + p.dialect.QuoteIdentifier(tableName)

	_, err := p.db.ExecContext(ctx, query)
	return err
}

// Analyze analyzes a table
func (p *PostgreSQLAdapter) Analyze(ctx context.Context, tableName string) error {
	if !p.connected || p.db == nil {
		return dbconn.ErrNotConnected
	}

	query := fmt.Sprintf("ANALYZE %s", p.dialect.QuoteIdentifier(tableName))
	_, err := p.db.ExecContext(ctx, query)
	return err
}

// GetMonitor returns the PostgreSQL monitor
func (p *PostgreSQLAdapter) GetMonitor() *PostgreSQLMonitor {
	return p.monitor
}

// Helper function to build PostgreSQL connection string
func buildPostgreSQLConnString(config *dbconn.DBConfig) string {
	parts := []string{
		fmt.Sprintf("host=%s", config.Host),
		fmt.Sprintf("port=%d", config.Port),
		fmt.Sprintf("user=%s", config.Username),
		fmt.Sprintf("password=%s", config.Password),
		fmt.Sprintf("dbname=%s", config.Database),
	}

	// Schema (search_path)
	if config.Schema != "" {
		parts = append(parts, fmt.Sprintf("search_path=%s", config.Schema))
	}

	// SSL Mode
	sslMode := config.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	parts = append(parts, fmt.Sprintf("sslmode=%s", sslMode))

	// SSL certificates
	if config.SSLCert != "" {
		parts = append(parts, fmt.Sprintf("sslcert=%s", config.SSLCert))
	}
	if config.SSLKey != "" {
		parts = append(parts, fmt.Sprintf("sslkey=%s", config.SSLKey))
	}
	if config.SSLRootCert != "" {
		parts = append(parts, fmt.Sprintf("sslrootcert=%s", config.SSLRootCert))
	}

	// Connection timeout
	parts = append(parts, fmt.Sprintf("connect_timeout=%d", config.ConnTimeout))

	// Additional options
	for k, v := range config.Options {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(parts, " ")
}

// Interface implementation

func (p *PostgreSQLAdapter) Close() error {
	if p.monitor != nil {
		p.monitor.Stop()
	}
	if p.db != nil {
		err := p.db.Close()
		p.connected = false
		return err
	}
	return nil
}

func (p *PostgreSQLAdapter) DB() *sql.DB {
	return p.db
}

func (p *PostgreSQLAdapter) GetDialect() dbconn.Dialect {
	return p.dialect
}

func (p *PostgreSQLAdapter) GetType() dbconn.DBType {
	return dbconn.DBTypePostgreSQL
}

func (p *PostgreSQLAdapter) IsConnected() bool {
	return p.connected && p.db != nil
}

func (p *PostgreSQLAdapter) Ping() error {
	if p.db == nil {
		return dbconn.ErrNotConnected
	}
	return p.db.Ping()
}

func (p *PostgreSQLAdapter) Stats() sql.DBStats {
	if p.db == nil {
		return sql.DBStats{}
	}
	return p.db.Stats()
}

func (p *PostgreSQLAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if p.db == nil {
		return nil, dbconn.ErrNotConnected
	}
	return p.db.BeginTx(ctx, opts)
}

func (p *PostgreSQLAdapter) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if p.db == nil {
		return nil, dbconn.ErrNotConnected
	}
	return p.db.QueryContext(ctx, query, args...)
}

func (p *PostgreSQLAdapter) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if p.db == nil {
		return nil
	}
	return p.db.QueryRowContext(ctx, query, args...)
}

func (p *PostgreSQLAdapter) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if p.db == nil {
		return nil, dbconn.ErrNotConnected
	}
	return p.db.ExecContext(ctx, query, args...)
}

func (p *PostgreSQLAdapter) SupportsFeature(feature dbconn.Feature) bool {
	if p.dialect == nil {
		return false
	}

	switch feature {
	case dbconn.FeatureCTE:
		return p.dialect.SupportsCTE()
	case dbconn.FeatureJSON:
		return p.dialect.SupportsJSON()
	case dbconn.FeatureFullTextSearch:
		return p.dialect.SupportsFullTextSearch()
	case dbconn.FeatureReturning:
		return p.dialect.SupportsReturning()
	case dbconn.FeatureUpsert:
		return p.dialect.SupportsUpsert()
	default:
		return false
	}
}
