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

	// Get database type for dialect-specific queries
	dbType := string(db.GetType())

	// Discover tables
	tablesQuery := GetTablesQuery(dbType, schemaName)
	rows, err := db.Query(ctx, tablesQuery)
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
			Where("databasealias = ? AND tablename = ? AND metadatatype = ?",
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
	// Get database type for dialect-specific queries
	dbType := string(db.GetType())

	// Get columns query
	columnsQuery := GetColumnsQuery(dbType, schemaName, tableName)
	rows, err := db.Query(ctx, columnsQuery)
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
		normalizedType := NormalizeDataType(dbType, column.DataType)

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
			Where("databasealias = ? AND tablename = ? AND columnname = ? AND metadatatype = 'column'",
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

	// Get database type for dialect-specific queries
	dbType := string(db.GetType())
	indexesQuery := GetIndexesQuery(dbType, schemaName, tableName)

	rows, err := db.Query(ctx, indexesQuery)
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

	return db.Query(ctx, query, args...)
}

// GetTableMetadata retrieves metadata for all tables in a database
func (s *SchemaMetadataServiceMultiDB) GetTableMetadata(ctx context.Context, databaseAlias string) ([]models.DatabaseSchemaMetadata, error) {
	var metadata []models.DatabaseSchemaMetadata

	if err := s.appDB.WithContext(ctx).
		Where("databasealias = ? AND metadatatype = ?", databaseAlias, models.MetadataTypeTable).
		Order("tablename").
		Find(&metadata).Error; err != nil {
		return nil, fmt.Errorf("failed to get table metadata: %w", err)
	}

	return metadata, nil
}

// GetColumnMetadata retrieves metadata for all columns in a table
func (s *SchemaMetadataServiceMultiDB) GetColumnMetadata(ctx context.Context, databaseAlias, tableName string) ([]models.DatabaseSchemaMetadata, error) {
	var metadata []models.DatabaseSchemaMetadata

	if err := s.appDB.WithContext(ctx).
		Where("databasealias = ? AND tablename = ? AND metadatatype = ?",
			databaseAlias, tableName, models.MetadataTypeColumn).
		Order("columnname").
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

	updates["modifiedon"] = gorm.Expr("CURRENT_TIMESTAMP")

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

	query := s.appDB.WithContext(ctx).Where("databasealias = ?", databaseAlias)

	if len(tableNames) > 0 {
		query = query.Where("tablename IN ?", tableNames)
	}

	if err := query.Order("tablename, metadatatype DESC, columnname").Find(&metadata).Error; err != nil {
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

// GetDatabaseMetadata retrieves complete metadata (tables and columns) for a database
// If no metadata exists, it automatically discovers and populates it from the database schema
func (s *SchemaMetadataServiceMultiDB) GetDatabaseMetadata(ctx context.Context, databaseAlias string) ([]models.DatabaseSchemaMetadata, error) {
	var metadata []models.DatabaseSchemaMetadata

	// Use Find which returns empty slice if no records found (not an error)
	if err := s.appDB.WithContext(ctx).
		Where("databasealias = ?", databaseAlias).
		Order("tablename, metadatatype DESC, columnname").
		Find(&metadata).Error; err != nil {
		// Only return error for actual database errors, not "record not found"
		return nil, fmt.Errorf("failed to get database metadata: %w", err)
	}

	// If no metadata found, automatically discover and populate it
	if len(metadata) == 0 {
		// Get database connection to determine schema name
		db, err := s.dbHelper.GetUserDB(ctx, databaseAlias)
		if err != nil {
			return nil, fmt.Errorf("failed to get database connection: %w", err)
		}
		defer db.Close()

		// Get schema name based on database type
		var schemaName string
		dbType := string(db.GetType())

		switch dbType {
		case "mysql":
			var dbName sql.NullString
			row := db.QueryRow(ctx, "SELECT DATABASE()")
			if err := row.Scan(&dbName); err != nil {
				return nil, fmt.Errorf("failed to get current database name: %w", err)
			}
			if !dbName.Valid || dbName.String == "" {
				// No database selected, return empty result
				return metadata, nil
			}
			schemaName = dbName.String

		case "postgres":
			var schema sql.NullString
			row := db.QueryRow(ctx, "SELECT current_schema()")
			if err := row.Scan(&schema); err != nil {
				return nil, fmt.Errorf("failed to get current schema: %w", err)
			}
			if !schema.Valid || schema.String == "" {
				schemaName = "public" // Default PostgreSQL schema
			} else {
				schemaName = schema.String
			}

		case "mssql":
			var schema sql.NullString
			row := db.QueryRow(ctx, "SELECT SCHEMA_NAME()")
			if err := row.Scan(&schema); err != nil {
				return nil, fmt.Errorf("failed to get current schema: %w", err)
			}
			if !schema.Valid || schema.String == "" {
				schemaName = "dbo" // Default MSSQL schema
			} else {
				schemaName = schema.String
			}

		case "oracle":
			var schema sql.NullString
			row := db.QueryRow(ctx, "SELECT SYS_CONTEXT('USERENV', 'CURRENT_SCHEMA') FROM DUAL")
			if err := row.Scan(&schema); err != nil {
				return nil, fmt.Errorf("failed to get current schema: %w", err)
			}
			if !schema.Valid || schema.String == "" {
				return metadata, nil
			}
			schemaName = schema.String

		default:
			return nil, fmt.Errorf("unsupported database type: %s", dbType)
		}

		// Auto-discover schema for this database
		if err := s.DiscoverSchema(ctx, databaseAlias, schemaName); err != nil {
			return nil, fmt.Errorf("failed to auto-discover schema: %w", err)
		}

		// Retrieve the newly discovered metadata
		if err := s.appDB.WithContext(ctx).
			Where("databasealias = ?", databaseAlias).
			Order("tablename, metadatatype DESC, columnname").
			Find(&metadata).Error; err != nil {
			return nil, fmt.Errorf("failed to get discovered metadata: %w", err)
		}
	}

	return metadata, nil
}

// IndexInfo represents index information
type IndexInfo struct {
	IndexName  string
	ColumnName string
	NonUnique  int
	SeqInIndex int
	IndexType  string
}
