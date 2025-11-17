package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// SchemaMetadataServiceMultiDB provides multi-database schema discovery
// This is an enhanced version of SchemaMetadataService that supports all database types
type SchemaMetadataServiceMultiDB struct {
	dbHelper *DatabaseHelper
	appDB    *gorm.DB
}

// NewSchemaMetadataServiceMultiDB creates a new multi-database schema metadata service
func NewSchemaMetadataServiceMultiDB(dbHelper *DatabaseHelper, appDB *gorm.DB) *SchemaMetadataServiceMultiDB {
	return &SchemaMetadataServiceMultiDB{
		dbHelper: dbHelper,
		appDB:    appDB,
	}
}

// DiscoverSchema discovers all tables and columns in a user database
// This method supports all database types (MySQL, PostgreSQL, MSSQL, Oracle)
func (s *SchemaMetadataServiceMultiDB) DiscoverSchema(ctx context.Context, databaseAlias, schemaName string) error {
	// Get database connection for the specified alias
	db, err := s.dbHelper.GetUserDB(ctx, databaseAlias)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	defer db.Close()

	dialect := db.GetDialect()

	// Discover tables
	tablesQuery := GetTablesQuery(dialect, schemaName)
	rows, err := db.Query(tablesQuery)
	if err != nil {
		return fmt.Errorf("failed to discover tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.TableName, &table.TableComment, &table.TableSchema); err != nil {
			return fmt.Errorf("failed to scan table row: %w", err)
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating table rows: %w", err)
	}

	// Process each table
	for _, table := range tables {
		// Create or update table metadata in IAC database
		tableMeta := &models.DatabaseSchemaMetadata{
			DatabaseAlias: databaseAlias,
			Table:         table.TableName,
			MetadataType:  models.MetadataTypeTable,
			Description:   table.TableComment,
		}

		if err := s.appDB.WithContext(ctx).
			Where("database_alias = ? AND table_name = ? AND metadata_type = ?",
				databaseAlias, table.TableName, models.MetadataTypeTable).
			Assign(tableMeta).
			FirstOrCreate(tableMeta).Error; err != nil {
			return fmt.Errorf("failed to save table metadata for %s: %w", table.TableName, err)
		}

		// Discover columns for this table
		if err := s.discoverTableColumns(ctx, db, databaseAlias, schemaName, table.TableName); err != nil {
			return fmt.Errorf("failed to discover columns for %s: %w", table.TableName, err)
		}
	}

	return nil
}

// discoverTableColumns discovers all columns for a specific table
func (s *SchemaMetadataServiceMultiDB) discoverTableColumns(ctx context.Context, db dbconn.RelationalDB, databaseAlias, schemaName, tableName string) error {
	dialect := db.GetDialect()

	// Get columns query
	columnsQuery := GetColumnsQuery(dialect, schemaName, tableName)
	rows, err := db.Query(columnsQuery)
	if err != nil {
		return fmt.Errorf("failed to discover columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var column ColumnInfo
		if err := rows.Scan(&column.ColumnName, &column.DataType, &column.IsNullable,
			&column.ColumnKey, &column.ColumnComment, &column.OrdinalPos); err != nil {
			return fmt.Errorf("failed to scan column row: %w", err)
		}
		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating column rows: %w", err)
	}

	// Process each column
	for _, column := range columns {
		isNullable := column.IsNullable == "YES"

		// Normalize data type to common format
		normalizedType := NormalizeDataType(dialect, column.DataType)

		columnMeta := &models.DatabaseSchemaMetadata{
			DatabaseAlias: databaseAlias,
			Table:         tableName,
			Column:        column.ColumnName,
			MetadataType:  models.MetadataTypeColumn,
			DataType:      normalizedType,
			IsNullable:    &isNullable,
			ColumnComment: column.ColumnComment,
		}

		if err := s.appDB.WithContext(ctx).
			Where("database_alias = ? AND table_name = ? AND column_name = ? AND metadata_type = 'column'",
				databaseAlias, tableName, column.ColumnName).
			Assign(columnMeta).
			FirstOrCreate(columnMeta).Error; err != nil {
			return fmt.Errorf("failed to save column metadata for %s.%s: %w", tableName, column.ColumnName, err)
		}
	}

	return nil
}

// DiscoverIndexes discovers all indexes for a specific table
func (s *SchemaMetadataServiceMultiDB) DiscoverIndexes(ctx context.Context, databaseAlias, schemaName, tableName string) ([]IndexInfo, error) {
	db, err := s.dbHelper.GetUserDB(ctx, databaseAlias)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}
	defer db.Close()

	dialect := db.GetDialect()
	indexesQuery := GetIndexesQuery(dialect, schemaName, tableName)

	rows, err := db.Query(indexesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to discover indexes: %w", err)
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var index IndexInfo
		if err := rows.Scan(&index.IndexName, &index.ColumnName, &index.NonUnique,
			&index.SeqInIndex, &index.IndexType); err != nil {
			return nil, fmt.Errorf("failed to scan index row: %w", err)
		}
		indexes = append(indexes, index)
	}

	return indexes, rows.Err()
}

// ExecuteQuery executes a custom SQL query against a user database
// This is useful for ad-hoc queries with dialect-aware execution
func (s *SchemaMetadataServiceMultiDB) ExecuteQuery(ctx context.Context, databaseAlias, query string, args ...interface{}) (*sql.Rows, error) {
	db, err := s.dbHelper.GetUserDB(ctx, databaseAlias)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}
	// Note: Caller must close db after using rows

	return db.Query(query, args...)
}

// GetTableMetadata retrieves metadata for all tables in a database
func (s *SchemaMetadataServiceMultiDB) GetTableMetadata(ctx context.Context, databaseAlias string) ([]models.DatabaseSchemaMetadata, error) {
	var metadata []models.DatabaseSchemaMetadata

	if err := s.appDB.WithContext(ctx).
		Where("database_alias = ? AND metadata_type = ?", databaseAlias, models.MetadataTypeTable).
		Order("table_name").
		Find(&metadata).Error; err != nil {
		return nil, fmt.Errorf("failed to get table metadata: %w", err)
	}

	return metadata, nil
}

// GetColumnMetadata retrieves metadata for all columns in a table
func (s *SchemaMetadataServiceMultiDB) GetColumnMetadata(ctx context.Context, databaseAlias, tableName string) ([]models.DatabaseSchemaMetadata, error) {
	var metadata []models.DatabaseSchemaMetadata

	if err := s.appDB.WithContext(ctx).
		Where("database_alias = ? AND table_name = ? AND metadata_type = ?",
			databaseAlias, tableName, models.MetadataTypeColumn).
		Order("column_name").
		Find(&metadata).Error; err != nil {
		return nil, fmt.Errorf("failed to get column metadata: %w", err)
	}

	return metadata, nil
}

// UpdateMetadata updates the description for a metadata entry
func (s *SchemaMetadataServiceMultiDB) UpdateMetadata(ctx context.Context, id string, description *string) error {
	updates := make(map[string]interface{})

	if description != nil {
		updates["description"] = *description
	}

	if len(updates) == 0 {
		return nil
	}

	updates["updated_at"] = gorm.Expr("CURRENT_TIMESTAMP")

	if err := s.appDB.WithContext(ctx).
		Model(&models.DatabaseSchemaMetadata{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}

// GetSchemaContext builds a comprehensive schema context for AI
func (s *SchemaMetadataServiceMultiDB) GetSchemaContext(ctx context.Context, databaseAlias string, tableNames []string) (string, error) {
	var metadata []models.DatabaseSchemaMetadata

	query := s.appDB.WithContext(ctx).Where("database_alias = ?", databaseAlias)

	if len(tableNames) > 0 {
		query = query.Where("table_name IN ?", tableNames)
	}

	if err := query.Order("table_name, metadata_type DESC, column_name").Find(&metadata).Error; err != nil {
		return "", fmt.Errorf("failed to get schema context: %w", err)
	}

	// Build context string
	var context strings.Builder
	context.WriteString("Database Schema Information:\n\n")

	currentTable := ""
	for _, meta := range metadata {
		if meta.MetadataType == models.MetadataTypeTable {
			currentTable = meta.Table
			context.WriteString(fmt.Sprintf("## Table: %s\n", meta.Table))

			if meta.Description != "" {
				context.WriteString(fmt.Sprintf("Description: %s\n", meta.Description))
			}

			context.WriteString("Columns:\n")
		} else if meta.MetadataType == models.MetadataTypeColumn && meta.Table == currentTable {
			columnInfo := fmt.Sprintf("  - %s", meta.Column)

			if meta.DataType != "" {
				columnInfo += fmt.Sprintf(" (%s)", meta.DataType)
			}

			if meta.IsNullable != nil && !*meta.IsNullable {
				columnInfo += " [NOT NULL]"
			}

			if meta.ColumnComment != "" {
				columnInfo += fmt.Sprintf(" - %s", meta.ColumnComment)
			}

			context.WriteString(columnInfo + "\n")
		}
	}

	return context.String(), nil
}

// IndexInfo represents index information
type IndexInfo struct {
	IndexName  string
	ColumnName string
	NonUnique  int
	SeqInIndex int
	IndexType  string
}
