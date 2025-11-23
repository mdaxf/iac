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
	"context"
	"database/sql"
	"time"
)

// RelationalDB defines the interface for relational database operations
// This interface abstracts database-specific operations to support multiple database types
type RelationalDB interface {
	// Connection Management
	Connect(config *DBConfig) error
	Close() error
	Ping() error
	DB() *sql.DB

	// Transaction Management
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	// Query Operations
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Database Information
	GetDialect() Dialect
	GetType() DBType
	SupportsFeature(feature Feature) bool

	// Health and Monitoring
	Stats() sql.DBStats
	IsConnected() bool

	// Schema Operations
	ListTables(ctx context.Context, schema string) ([]string, error)
	TableExists(ctx context.Context, schema, tableName string) (bool, error)
	GetTableSchema(ctx context.Context, schema, tableName string) (*TableSchema, error)
}

// Dialect defines the interface for database-specific SQL dialect operations
type Dialect interface {
	// Query Building
	QuoteIdentifier(name string) string
	Placeholder(n int) string
	LimitOffset(limit, offset int) string

	// Type Conversion
	DataTypeMapping(genericType string) string
	ConvertValue(value interface{}, targetType string) (interface{}, error)

	// Feature Detection
	SupportsReturning() bool
	SupportsUpsert() bool
	SupportsCTE() bool
	SupportsJSON() bool
	SupportsFullTextSearch() bool

	// Query Translation
	TranslatePagination(query string, limit, offset int) string
	TranslateUpsert(table string, columns []string, conflictColumns []string) string
	ConvertJSONQuery(query string) string // Convert MySQL JSON_TABLE to database-specific syntax

	// DDL Generation - Schema Management
	// These methods enable database-agnostic table schema definitions
	CreateTableDDL(schema *TableSchema) string
	AddColumnDDL(tableName string, column *ColumnInfo) string
	DropColumnDDL(tableName, columnName string) string
	AlterColumnDDL(tableName string, column *ColumnInfo) string
	CreateIndexDDL(tableName string, index *IndexInfo) string
	DropIndexDDL(tableName, indexName string) string
}

// DBConfig represents database connection configuration
type DBConfig struct {
	// Connection Details
	Type         DBType            `json:"type"`
	Host         string            `json:"host"`
	Port         int               `json:"port"`
	Database     string            `json:"database"`
	Schema       string            `json:"schema,omitempty"`
	Username     string            `json:"username"`
	Password     string            `json:"password"`

	// Connection Options
	SSLMode      string            `json:"ssl_mode,omitempty"`
	SSLCert      string            `json:"ssl_cert,omitempty"`
	SSLKey       string            `json:"ssl_key,omitempty"`
	SSLRootCert  string            `json:"ssl_root_cert,omitempty"`

	// Pool Configuration
	MaxIdleConns int               `json:"max_idle_conns"`
	MaxOpenConns int               `json:"max_open_conns"`
	ConnMaxLifetime time.Duration  `json:"conn_max_lifetime,omitempty"`
	ConnMaxIdleTime time.Duration  `json:"conn_max_idle_time,omitempty"`
	ConnTimeout  int               `json:"conn_timeout"`

	// Database-Specific Options
	Options      map[string]string `json:"options,omitempty"`
}

// DBType represents supported database types
type DBType string

const (
	DBTypeMySQL      DBType = "mysql"
	DBTypePostgreSQL DBType = "postgres"
	DBTypeMSSQL      DBType = "mssql"
	DBTypeOracle     DBType = "oracle"
)

// Feature represents database feature flags
type Feature string

const (
	FeatureCTE              Feature = "cte"
	FeatureWindowFunctions  Feature = "window_functions"
	FeatureJSON             Feature = "json"
	FeatureFullTextSearch   Feature = "fulltext_search"
	FeatureReturning        Feature = "returning"
	FeatureUpsert           Feature = "upsert"
	FeatureArrays           Feature = "arrays"
	FeaturePartitioning     Feature = "partitioning"
	FeatureStoredProcedures Feature = "stored_procedures"
)

// TableSchema represents database table schema information
type TableSchema struct {
	Schema      string
	TableName   string
	Columns     []ColumnInfo
	PrimaryKeys []string
	Indexes     []IndexInfo
}

// ColumnInfo represents column metadata
type ColumnInfo struct {
	Name         string
	DataType     string
	IsNullable   bool
	DefaultValue *string
	MaxLength    *int
	Precision    *int
	Scale        *int
	IsPrimaryKey bool
	IsForeignKey bool
	IsUnique     bool
	Comment      string
}

// IndexInfo represents index metadata
type IndexInfo struct {
	Name     string
	Columns  []string
	IsUnique bool
	Type     string
}

// ConnectionStats represents connection pool statistics
type ConnectionStats struct {
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
	WaitCount          int64
	WaitDuration       time.Duration
	MaxIdleClosed      int64
	MaxLifetimeClosed  int64
}

// BuildConnectionString builds database-specific connection string
func (c *DBConfig) BuildConnectionString() string {
	// This will be implemented by database-specific builders
	return ""
}

// Validate validates the database configuration
func (c *DBConfig) Validate() error {
	if c.Type == "" {
		return ErrInvalidDatabaseType
	}
	if c.Host == "" {
		return ErrMissingHost
	}
	if c.Database == "" {
		return ErrMissingDatabase
	}
	if c.Username == "" {
		return ErrMissingUsername
	}

	// Set defaults
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 5
	}
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 10
	}
	if c.ConnTimeout == 0 {
		c.ConnTimeout = 30
	}

	return nil
}

// GetStats converts sql.DBStats to ConnectionStats
func GetStats(stats sql.DBStats) ConnectionStats {
	return ConnectionStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}
}
