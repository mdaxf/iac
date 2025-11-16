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

	"github.com/mdaxf/iac/logger"
)

// SchemaManager provides database-agnostic schema management operations
// It uses the dialect system to generate appropriate DDL for different databases
type SchemaManager struct {
	db      *sql.DB
	dialect Dialect
	iLog    logger.Log
}

// NewSchemaManager creates a new SchemaManager instance
// It automatically detects the database type and uses the appropriate dialect
func NewSchemaManager(db *sql.DB, user string) (*SchemaManager, error) {
	dialect, err := GetFactory().GetDialect(DBType(DatabaseType))
	if err != nil {
		return nil, fmt.Errorf("failed to get dialect: %w", err)
	}

	return &SchemaManager{
		db:      db,
		dialect: dialect,
		iLog:    logger.Log{ModuleName: logger.Database, User: user},
	}, nil
}

// CreateTable creates a new table based on the schema definition
// Works across all supported database types (MySQL, PostgreSQL, MSSQL, Oracle)
func (sm *SchemaManager) CreateTable(ctx context.Context, schema *TableSchema) error {
	ddl := sm.dialect.CreateTableDDL(schema)
	sm.iLog.Debug(fmt.Sprintf("CreateTable DDL: %s", ddl))

	_, err := sm.db.ExecContext(ctx, ddl)
	if err != nil {
		sm.iLog.Error(fmt.Sprintf("Failed to create table %s: %v", schema.TableName, err))
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Create indexes if defined
	for _, index := range schema.Indexes {
		if err := sm.CreateIndex(ctx, schema.TableName, &index); err != nil {
			sm.iLog.Error(fmt.Sprintf("Failed to create index %s: %v", index.Name, err))
			// Continue with other indexes even if one fails
		}
	}

	sm.iLog.Info(fmt.Sprintf("Successfully created table: %s", schema.TableName))
	return nil
}

// AddColumn adds a new column to an existing table
func (sm *SchemaManager) AddColumn(ctx context.Context, tableName string, column *ColumnInfo) error {
	ddl := sm.dialect.AddColumnDDL(tableName, column)
	sm.iLog.Debug(fmt.Sprintf("AddColumn DDL: %s", ddl))

	_, err := sm.db.ExecContext(ctx, ddl)
	if err != nil {
		sm.iLog.Error(fmt.Sprintf("Failed to add column %s to table %s: %v", column.Name, tableName, err))
		return fmt.Errorf("failed to add column: %w", err)
	}

	sm.iLog.Info(fmt.Sprintf("Successfully added column %s to table %s", column.Name, tableName))
	return nil
}

// DropColumn removes a column from a table
func (sm *SchemaManager) DropColumn(ctx context.Context, tableName, columnName string) error {
	ddl := sm.dialect.DropColumnDDL(tableName, columnName)
	sm.iLog.Debug(fmt.Sprintf("DropColumn DDL: %s", ddl))

	_, err := sm.db.ExecContext(ctx, ddl)
	if err != nil {
		sm.iLog.Error(fmt.Sprintf("Failed to drop column %s from table %s: %v", columnName, tableName, err))
		return fmt.Errorf("failed to drop column: %w", err)
	}

	sm.iLog.Info(fmt.Sprintf("Successfully dropped column %s from table %s", columnName, tableName))
	return nil
}

// AlterColumn modifies an existing column's definition
func (sm *SchemaManager) AlterColumn(ctx context.Context, tableName string, column *ColumnInfo) error {
	ddl := sm.dialect.AlterColumnDDL(tableName, column)
	sm.iLog.Debug(fmt.Sprintf("AlterColumn DDL: %s", ddl))

	_, err := sm.db.ExecContext(ctx, ddl)
	if err != nil {
		sm.iLog.Error(fmt.Sprintf("Failed to alter column %s in table %s: %v", column.Name, tableName, err))
		return fmt.Errorf("failed to alter column: %w", err)
	}

	sm.iLog.Info(fmt.Sprintf("Successfully altered column %s in table %s", column.Name, tableName))
	return nil
}

// CreateIndex creates a new index on a table
func (sm *SchemaManager) CreateIndex(ctx context.Context, tableName string, index *IndexInfo) error {
	ddl := sm.dialect.CreateIndexDDL(tableName, index)
	sm.iLog.Debug(fmt.Sprintf("CreateIndex DDL: %s", ddl))

	_, err := sm.db.ExecContext(ctx, ddl)
	if err != nil {
		sm.iLog.Error(fmt.Sprintf("Failed to create index %s on table %s: %v", index.Name, tableName, err))
		return fmt.Errorf("failed to create index: %w", err)
	}

	sm.iLog.Info(fmt.Sprintf("Successfully created index %s on table %s", index.Name, tableName))
	return nil
}

// DropIndex removes an index from a table
func (sm *SchemaManager) DropIndex(ctx context.Context, tableName, indexName string) error {
	ddl := sm.dialect.DropIndexDDL(tableName, indexName)
	sm.iLog.Debug(fmt.Sprintf("DropIndex DDL: %s", ddl))

	_, err := sm.db.ExecContext(ctx, ddl)
	if err != nil {
		sm.iLog.Error(fmt.Sprintf("Failed to drop index %s from table %s: %v", indexName, tableName, err))
		return fmt.Errorf("failed to drop index: %w", err)
	}

	sm.iLog.Info(fmt.Sprintf("Successfully dropped index %s from table %s", indexName, tableName))
	return nil
}

// DropTable removes a table from the database
func (sm *SchemaManager) DropTable(ctx context.Context, tableName string) error {
	ddl := fmt.Sprintf("DROP TABLE %s", sm.dialect.QuoteIdentifier(tableName))
	sm.iLog.Debug(fmt.Sprintf("DropTable DDL: %s", ddl))

	_, err := sm.db.ExecContext(ctx, ddl)
	if err != nil {
		sm.iLog.Error(fmt.Sprintf("Failed to drop table %s: %v", tableName, err))
		return fmt.Errorf("failed to drop table: %w", err)
	}

	sm.iLog.Info(fmt.Sprintf("Successfully dropped table: %s", tableName))
	return nil
}

// TableExists checks if a table exists in the database
func (sm *SchemaManager) TableExists(ctx context.Context, tableName string) (bool, error) {
	var query string
	switch DBType(DatabaseType) {
	case DBTypeMySQL:
		query = fmt.Sprintf("SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = '%s'", tableName)
	case DBTypePostgreSQL:
		query = fmt.Sprintf("SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '%s'", tableName)
	case DBTypeMSSQL:
		query = fmt.Sprintf("SELECT 1 FROM information_schema.tables WHERE table_name = '%s'", tableName)
	case DBTypeOracle:
		query = fmt.Sprintf("SELECT 1 FROM user_tables WHERE table_name = '%s'", tableName)
	default:
		return false, fmt.Errorf("unsupported database type: %s", DatabaseType)
	}

	var exists int
	err := sm.db.QueryRowContext(ctx, query).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// MigrateSchema compares current schema with desired schema and applies changes
// This is useful for schema evolution without manual DDL management
func (sm *SchemaManager) MigrateSchema(ctx context.Context, desiredSchema *TableSchema) error {
	exists, err := sm.TableExists(ctx, desiredSchema.TableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !exists {
		// Table doesn't exist, create it
		return sm.CreateTable(ctx, desiredSchema)
	}

	// TODO: Implement schema comparison and migration logic
	// This would involve:
	// 1. Getting current table schema
	// 2. Comparing with desired schema
	// 3. Generating and executing ALTER statements for differences
	sm.iLog.Info(fmt.Sprintf("Table %s exists. Schema migration logic to be implemented.", desiredSchema.TableName))
	return nil
}
