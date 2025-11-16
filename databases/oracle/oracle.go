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

package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/sijms/go-ora/v2"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/logger"
)

// init registers the Oracle adapter with the global factory
func init() {
	dbconn.GetFactory().RegisterDriver(dbconn.DBTypeOracle, func(config *dbconn.DBConfig) (dbconn.RelationalDB, error) {
		return NewOracleAdapter(config), nil
	})
}

// OracleAdapter is an Oracle-specific database adapter
type OracleAdapter struct {
	config    *dbconn.DBConfig
	db        *sql.DB
	dialect   *dbconn.OracleDialect
	connected bool
	monitor   *OracleMonitor
}

// NewOracleAdapter creates a new Oracle adapter
func NewOracleAdapter(config *dbconn.DBConfig) *OracleAdapter {
	return &OracleAdapter{
		config:  config,
		dialect: dbconn.NewOracleDialect(),
	}
}

// Connect establishes connection to Oracle database
func (o *OracleAdapter) Connect(config *dbconn.DBConfig) error {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "OracleAdapter"}

	// Build Oracle connection string
	connStr := buildOracleConnString(config)

	// Open database connection
	db, err := sql.Open("oracle", connStr)
	if err != nil {
		return dbconn.NewDatabaseError("connect", err, dbconn.DBTypeOracle, "")
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
		return dbconn.NewDatabaseError("ping", err, dbconn.DBTypeOracle, "")
	}

	o.db = db
	o.connected = true

	// Initialize monitor
	o.monitor = NewOracleMonitor(db, config)

	iLog.Info(fmt.Sprintf("Connected to Oracle database at %s:%d/%s",
		config.Host, config.Port, config.Database))

	return nil
}

// ListTables returns all tables in the schema
func (o *OracleAdapter) ListTables(ctx context.Context, schema string) ([]string, error) {
	if !o.connected || o.db == nil {
		return nil, dbconn.ErrNotConnected
	}

	schemaName := schema
	if schemaName == "" {
		schemaName = o.config.Schema
		if schemaName == "" {
			schemaName = strings.ToUpper(o.config.Username)
		}
	}

	query := `
		SELECT table_name
		FROM all_tables
		WHERE owner = :1
		ORDER BY table_name
	`

	rows, err := o.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, dbconn.NewDatabaseError("list_tables", err, dbconn.DBTypeOracle, query)
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
func (o *OracleAdapter) TableExists(ctx context.Context, schema, tableName string) (bool, error) {
	if !o.connected || o.db == nil {
		return false, dbconn.ErrNotConnected
	}

	schemaName := schema
	if schemaName == "" {
		schemaName = o.config.Schema
		if schemaName == "" {
			schemaName = strings.ToUpper(o.config.Username)
		}
	}

	query := `
		SELECT COUNT(*)
		FROM all_tables
		WHERE owner = :1
		AND table_name = :2
	`

	var count int
	err := o.db.QueryRowContext(ctx, query, schemaName, strings.ToUpper(tableName)).Scan(&count)
	if err != nil {
		return false, dbconn.NewDatabaseError("table_exists", err, dbconn.DBTypeOracle, query)
	}

	return count > 0, nil
}

// GetTableSchema retrieves the schema for a table
func (o *OracleAdapter) GetTableSchema(ctx context.Context, schema, tableName string) (*dbconn.TableSchema, error) {
	if !o.connected || o.db == nil {
		return nil, dbconn.ErrNotConnected
	}

	schemaName := schema
	if schemaName == "" {
		schemaName = o.config.Schema
		if schemaName == "" {
			schemaName = strings.ToUpper(o.config.Username)
		}
	}

	// Check if table exists
	exists, err := o.TableExists(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, dbconn.ErrTableNotFound
	}

	tableSchema := &dbconn.TableSchema{
		Schema:    schemaName,
		TableName: strings.ToUpper(tableName),
		Columns:   make([]dbconn.ColumnInfo, 0),
		Indexes:   make([]dbconn.IndexInfo, 0),
	}

	// Get columns
	columnQuery := `
		SELECT
			column_name,
			data_type,
			nullable,
			data_default,
			data_length,
			data_precision,
			data_scale
		FROM all_tab_columns
		WHERE owner = :1 AND table_name = :2
		ORDER BY column_id
	`

	rows, err := o.db.QueryContext(ctx, columnQuery, schemaName, strings.ToUpper(tableName))
	if err != nil {
		return nil, dbconn.NewDatabaseError("get_table_schema", err, dbconn.DBTypeOracle, columnQuery)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			colName    string
			dataType   string
			nullable   string
			defaultVal sql.NullString
			length     sql.NullInt64
			precision  sql.NullInt64
			scale      sql.NullInt64
		)

		err := rows.Scan(&colName, &dataType, &nullable, &defaultVal, &length, &precision, &scale)
		if err != nil {
			return nil, err
		}

		colInfo := dbconn.ColumnInfo{
			Name:       colName,
			DataType:   dataType,
			IsNullable: nullable == "Y",
		}

		if defaultVal.Valid {
			colInfo.DefaultValue = &defaultVal.String
		}
		if length.Valid {
			maxLen := int(length.Int64)
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
		SELECT cols.column_name
		FROM all_constraints cons
		JOIN all_cons_columns cols
			ON cons.constraint_name = cols.constraint_name
			AND cons.owner = cols.owner
		WHERE cons.constraint_type = 'P'
		AND cons.owner = :1
		AND cons.table_name = :2
		ORDER BY cols.position
	`

	pkRows, err := o.db.QueryContext(ctx, pkQuery, schemaName, strings.ToUpper(tableName))
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
			i.index_name,
			ic.column_name,
			i.uniqueness
		FROM all_indexes i
		JOIN all_ind_columns ic
			ON i.index_name = ic.index_name
			AND i.owner = ic.index_owner
		WHERE i.owner = :1
		AND i.table_name = :2
		AND i.index_type != 'LOB'
		ORDER BY i.index_name, ic.column_position
	`

	indexRows, err := o.db.QueryContext(ctx, indexQuery, schemaName, strings.ToUpper(tableName))
	if err != nil {
		return nil, err
	}
	defer indexRows.Close()

	indexMap := make(map[string]*dbconn.IndexInfo)

	for indexRows.Next() {
		var (
			indexName  string
			columnName string
			uniqueness string
		)

		if err := indexRows.Scan(&indexName, &columnName, &uniqueness); err != nil {
			return nil, err
		}

		if idx, exists := indexMap[indexName]; exists {
			idx.Columns = append(idx.Columns, columnName)
		} else {
			indexMap[indexName] = &dbconn.IndexInfo{
				Name:     indexName,
				Columns:  []string{columnName},
				IsUnique: uniqueness == "UNIQUE",
				Type:     "BTREE", // Oracle default
			}
		}
	}

	for _, idx := range indexMap {
		tableSchema.Indexes = append(tableSchema.Indexes, *idx)
	}

	return tableSchema, nil
}

// GetServerVersion returns the Oracle server version
func (o *OracleAdapter) GetServerVersion(ctx context.Context) (string, error) {
	if !o.connected || o.db == nil {
		return "", dbconn.ErrNotConnected
	}

	var version string
	err := o.db.QueryRowContext(ctx, "SELECT banner FROM v$version WHERE rownum = 1").Scan(&version)
	if err != nil {
		return "", err
	}

	return version, nil
}

// GatherStats gathers statistics for a table
func (o *OracleAdapter) GatherStats(ctx context.Context, schema, tableName string) error {
	if !o.connected || o.db == nil {
		return dbconn.ErrNotConnected
	}

	schemaName := schema
	if schemaName == "" {
		schemaName = o.config.Schema
		if schemaName == "" {
			schemaName = strings.ToUpper(o.config.Username)
		}
	}

	query := `BEGIN DBMS_STATS.GATHER_TABLE_STATS(:1, :2); END;`
	_, err := o.db.ExecContext(ctx, query, schemaName, strings.ToUpper(tableName))
	return err
}

// RebuildIndex rebuilds an index
func (o *OracleAdapter) RebuildIndex(ctx context.Context, indexName string) error {
	if !o.connected || o.db == nil {
		return dbconn.ErrNotConnected
	}

	query := fmt.Sprintf("ALTER INDEX %s REBUILD", o.dialect.QuoteIdentifier(indexName))
	_, err := o.db.ExecContext(ctx, query)
	return err
}

// GetMonitor returns the Oracle monitor
func (o *OracleAdapter) GetMonitor() *OracleMonitor {
	return o.monitor
}

// Helper function to build Oracle connection string
func buildOracleConnString(config *dbconn.DBConfig) string {
	params := url.Values{}

	// Connection timeout
	if config.ConnTimeout > 0 {
		params.Add("TIMEOUT", fmt.Sprintf("%d", config.ConnTimeout))
	}

	// SSL/TLS configuration
	if config.SSLMode != "" && config.SSLMode != "disable" {
		params.Add("SSL", "true")
		if config.SSLCert != "" {
			params.Add("SSL Cert", config.SSLCert)
		}
		if config.SSLKey != "" {
			params.Add("SSL Key", config.SSLKey)
		}
		if config.SSLRootCert != "" {
			params.Add("SSL CA", config.SSLRootCert)
		}
	}

	// Additional options
	for k, v := range config.Options {
		params.Add(k, v)
	}

	// Service name or SID
	serviceName := config.Database
	if serviceName == "" {
		serviceName = "ORCL" // Default Oracle SID
	}

	connStr := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		url.QueryEscape(config.Username),
		url.QueryEscape(config.Password),
		config.Host,
		config.Port,
		serviceName,
	)

	if len(params) > 0 {
		connStr += "?" + params.Encode()
	}

	return connStr
}

// Close closes the Oracle connection
func (o *OracleAdapter) Close() error {
	if o.monitor != nil {
		o.monitor.Stop()
	}

	if o.db != nil {
		err := o.db.Close()
		o.connected = false
		return err
	}

	return nil
}

// DB returns the underlying sql.DB
func (o *OracleAdapter) DB() *sql.DB {
	return o.db
}

// GetDialect returns the Oracle dialect
func (o *OracleAdapter) GetDialect() dbconn.Dialect {
	return o.dialect
}

// GetType returns the database type
func (o *OracleAdapter) GetType() dbconn.DBType {
	return dbconn.DBTypeOracle
}

// IsConnected checks if connected
func (o *OracleAdapter) IsConnected() bool {
	return o.connected && o.db != nil
}

// Ping pings the database
func (o *OracleAdapter) Ping() error {
	if o.db == nil {
		return dbconn.ErrNotConnected
	}
	return o.db.Ping()
}

// Stats returns connection statistics
func (o *OracleAdapter) Stats() sql.DBStats {
	if o.db == nil {
		return sql.DBStats{}
	}
	return o.db.Stats()
}

// BeginTx starts a transaction
func (o *OracleAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if o.db == nil {
		return nil, dbconn.ErrNotConnected
	}
	return o.db.BeginTx(ctx, opts)
}

// Query executes a query
func (o *OracleAdapter) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if o.db == nil {
		return nil, dbconn.ErrNotConnected
	}
	return o.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (o *OracleAdapter) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if o.db == nil {
		return nil
	}
	return o.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query without returning rows
func (o *OracleAdapter) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if o.db == nil {
		return nil, dbconn.ErrNotConnected
	}
	return o.db.ExecContext(ctx, query, args...)
}

// SupportsFeature checks if a feature is supported
func (o *OracleAdapter) SupportsFeature(feature dbconn.Feature) bool {
	if o.dialect == nil {
		return false
	}

	switch feature {
	case dbconn.FeatureCTE:
		return o.dialect.SupportsCTE()
	case dbconn.FeatureJSON:
		return o.dialect.SupportsJSON()
	case dbconn.FeatureFullTextSearch:
		return o.dialect.SupportsFullTextSearch()
	case dbconn.FeatureReturning:
		return o.dialect.SupportsReturning()
	case dbconn.FeatureUpsert:
		return o.dialect.SupportsUpsert()
	default:
		return false
	}
}
