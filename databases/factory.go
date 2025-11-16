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
	"fmt"
	"sync"

	"github.com/mdaxf/iac/logger"
)

// DatabaseFactory manages database instance creation
type DatabaseFactory struct {
	drivers      map[DBType]DriverConstructor
	dialects     map[DBType]Dialect
	mu           sync.RWMutex
	instances    map[string]RelationalDB
	instancesMu  sync.RWMutex
}

// DriverConstructor is a function that creates a database driver instance
type DriverConstructor func(*DBConfig) (RelationalDB, error)

var (
	// globalFactory is the singleton factory instance
	globalFactory *DatabaseFactory
	factoryOnce   sync.Once
)

// GetFactory returns the global database factory instance
func GetFactory() *DatabaseFactory {
	factoryOnce.Do(func() {
		globalFactory = &DatabaseFactory{
			drivers:   make(map[DBType]DriverConstructor),
			dialects:  make(map[DBType]Dialect),
			instances: make(map[string]RelationalDB),
		}
		// Register built-in drivers
		globalFactory.registerBuiltinDrivers()
	})
	return globalFactory
}

// RegisterDriver registers a database driver constructor
func (f *DatabaseFactory) RegisterDriver(dbType DBType, constructor DriverConstructor) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.drivers[dbType] = constructor
}

// RegisterDialect registers a database dialect
func (f *DatabaseFactory) RegisterDialect(dbType DBType, dialect Dialect) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dialects[dbType] = dialect
}

// NewRelationalDB creates a new relational database instance
func (f *DatabaseFactory) NewRelationalDB(config *DBConfig) (RelationalDB, error) {
	if config == nil {
		return nil, ErrInvalidConfig
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}

	f.mu.RLock()
	constructor, exists := f.drivers[config.Type]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("database driver not found for type: %s", config.Type)
	}

	return constructor(config)
}

// GetOrCreateDB gets an existing database instance or creates a new one
func (f *DatabaseFactory) GetOrCreateDB(name string, config *DBConfig) (RelationalDB, error) {
	// Check if instance exists
	f.instancesMu.RLock()
	if db, exists := f.instances[name]; exists {
		f.instancesMu.RUnlock()
		return db, nil
	}
	f.instancesMu.RUnlock()

	// Create new instance
	db, err := f.NewRelationalDB(config)
	if err != nil {
		return nil, err
	}

	// Connect to database
	if err := db.Connect(config); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Store instance
	f.instancesMu.Lock()
	f.instances[name] = db
	f.instancesMu.Unlock()

	return db, nil
}

// GetDB retrieves an existing database instance
func (f *DatabaseFactory) GetDB(name string) (RelationalDB, error) {
	f.instancesMu.RLock()
	defer f.instancesMu.RUnlock()

	db, exists := f.instances[name]
	if !exists {
		return nil, fmt.Errorf("database instance not found: %s", name)
	}

	return db, nil
}

// CloseDB closes and removes a database instance
func (f *DatabaseFactory) CloseDB(name string) error {
	f.instancesMu.Lock()
	defer f.instancesMu.Unlock()

	db, exists := f.instances[name]
	if !exists {
		return fmt.Errorf("database instance not found: %s", name)
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	delete(f.instances, name)
	return nil
}

// CloseAll closes all database instances
func (f *DatabaseFactory) CloseAll() error {
	f.instancesMu.Lock()
	defer f.instancesMu.Unlock()

	var errors []error
	for name, db := range f.instances {
		if err := db.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close %s: %w", name, err))
		}
	}

	f.instances = make(map[string]RelationalDB)

	if len(errors) > 0 {
		return fmt.Errorf("errors closing databases: %v", errors)
	}

	return nil
}

// GetDialect returns the dialect for a database type
func (f *DatabaseFactory) GetDialect(dbType DBType) (Dialect, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	dialect, exists := f.dialects[dbType]
	if !exists {
		return nil, ErrDialectNotFound
	}

	return dialect, nil
}

// ListInstances returns the names of all active database instances
func (f *DatabaseFactory) ListInstances() []string {
	f.instancesMu.RLock()
	defer f.instancesMu.RUnlock()

	names := make([]string, 0, len(f.instances))
	for name := range f.instances {
		names = append(names, name)
	}

	return names
}

// registerBuiltinDrivers registers the built-in database drivers
// Individual adapter packages will override these using their init() functions
func (f *DatabaseFactory) registerBuiltinDrivers() {
	// MySQL driver (will be overridden by mysql package init)
	f.RegisterDriver(DBTypeMySQL, func(config *DBConfig) (RelationalDB, error) {
		return NewGenericSQLDB(config), nil
	})

	// PostgreSQL driver (will be overridden by postgres package init)
	f.RegisterDriver(DBTypePostgreSQL, func(config *DBConfig) (RelationalDB, error) {
		return NewGenericSQLDB(config), nil
	})

	// MSSQL driver (will be overridden by mssql package init)
	f.RegisterDriver(DBTypeMSSQL, func(config *DBConfig) (RelationalDB, error) {
		return NewGenericSQLDB(config), nil
	})

	// Oracle driver (will be overridden by oracle package init)
	f.RegisterDriver(DBTypeOracle, func(config *DBConfig) (RelationalDB, error) {
		return NewGenericSQLDB(config), nil
	})

	// Register dialects
	f.RegisterDialect(DBTypeMySQL, NewMySQLDialect())
	f.RegisterDialect(DBTypePostgreSQL, NewPostgreSQLDialect())
	f.RegisterDialect(DBTypeMSSQL, NewMSSQLDialect())
	f.RegisterDialect(DBTypeOracle, NewOracleDialect())
}

// Convenience functions

// NewRelationalDB creates a new relational database instance using the global factory
func NewRelationalDB(config *DBConfig) (RelationalDB, error) {
	return GetFactory().NewRelationalDB(config)
}

// GetOrCreateDB gets or creates a database instance using the global factory
func GetOrCreateDB(name string, config *DBConfig) (RelationalDB, error) {
	return GetFactory().GetOrCreateDB(name, config)
}

// GetDBInstance retrieves a database instance using the global factory
func GetDBInstance(name string) (RelationalDB, error) {
	return GetFactory().GetDB(name)
}

// CloseDBInstance closes a database instance using the global factory
func CloseDBInstance(name string) error {
	return GetFactory().CloseDB(name)
}

// CloseAllDatabases closes all database instances using the global factory
func CloseAllDatabases() error {
	return GetFactory().CloseAll()
}

// GenericSQLDB is a generic SQL database adapter implementing RelationalDB
type GenericSQLDB struct {
	config    *DBConfig
	db        *sql.DB
	dialect   Dialect
	connected bool
	mu        sync.RWMutex
}

// NewGenericSQLDB creates a new generic SQL database adapter
func NewGenericSQLDB(config *DBConfig) *GenericSQLDB {
	dialect, _ := GetFactory().GetDialect(config.Type)
	return &GenericSQLDB{
		config:  config,
		dialect: dialect,
	}
}

// Connect establishes connection to the database
func (g *GenericSQLDB) Connect(config *DBConfig) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "GenericSQLDB"}

	// Build connection string
	connStr, err := BuildConnectionString(config)
	if err != nil {
		return fmt.Errorf("failed to build connection string: %w", err)
	}

	// Open database connection
	db, err := sql.Open(string(config.Type), connStr)
	if err != nil {
		return NewDatabaseError("connect", err, config.Type, "")
	}

	// Configure connection pool
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetMaxOpenConns(config.MaxOpenConns)
	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	}
	if config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return NewDatabaseError("ping", err, config.Type, "")
	}

	g.db = db
	g.connected = true

	iLog.Info(fmt.Sprintf("Connected to %s database at %s:%d/%s",
		config.Type, config.Host, config.Port, config.Database))

	return nil
}

// Implement RelationalDB interface methods

func (g *GenericSQLDB) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.db != nil {
		err := g.db.Close()
		g.connected = false
		return err
	}
	return nil
}

func (g *GenericSQLDB) Ping() error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.db == nil {
		return ErrNotConnected
	}
	return g.db.Ping()
}

func (g *GenericSQLDB) DB() *sql.DB {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.db
}

func (g *GenericSQLDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.db == nil {
		return nil, ErrNotConnected
	}
	return g.db.BeginTx(ctx, opts)
}

func (g *GenericSQLDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.db == nil {
		return nil, ErrNotConnected
	}
	return g.db.QueryContext(ctx, query, args...)
}

func (g *GenericSQLDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.db == nil {
		return nil
	}
	return g.db.QueryRowContext(ctx, query, args...)
}

func (g *GenericSQLDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.db == nil {
		return nil, ErrNotConnected
	}
	return g.db.ExecContext(ctx, query, args...)
}

func (g *GenericSQLDB) GetDialect() Dialect {
	return g.dialect
}

func (g *GenericSQLDB) GetType() DBType {
	return g.config.Type
}

func (g *GenericSQLDB) SupportsFeature(feature Feature) bool {
	if g.dialect == nil {
		return false
	}

	switch feature {
	case FeatureCTE:
		return g.dialect.SupportsCTE()
	case FeatureJSON:
		return g.dialect.SupportsJSON()
	case FeatureFullTextSearch:
		return g.dialect.SupportsFullTextSearch()
	case FeatureReturning:
		return g.dialect.SupportsReturning()
	case FeatureUpsert:
		return g.dialect.SupportsUpsert()
	default:
		return false
	}
}

func (g *GenericSQLDB) Stats() sql.DBStats {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.db == nil {
		return sql.DBStats{}
	}
	return g.db.Stats()
}

func (g *GenericSQLDB) IsConnected() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.connected && g.db != nil
}

func (g *GenericSQLDB) ListTables(ctx context.Context, schema string) ([]string, error) {
	// Implementation will be database-specific
	return nil, ErrFeatureNotSupported
}

func (g *GenericSQLDB) TableExists(ctx context.Context, schema, tableName string) (bool, error) {
	// Implementation will be database-specific
	return false, ErrFeatureNotSupported
}

func (g *GenericSQLDB) GetTableSchema(ctx context.Context, schema, tableName string) (*TableSchema, error) {
	// Implementation will be database-specific
	return nil, ErrFeatureNotSupported
}
