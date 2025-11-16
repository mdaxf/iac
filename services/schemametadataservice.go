package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// SchemaMetadataService handles database schema discovery and metadata management
type SchemaMetadataService struct {
	db *gorm.DB
}

// NewSchemaMetadataService creates a new schema metadata service
func NewSchemaMetadataService(db *gorm.DB) *SchemaMetadataService {
	return &SchemaMetadataService{db: db}
}

// DiscoverDatabaseSchema discovers all tables and columns in a database
func (s *SchemaMetadataService) DiscoverDatabaseSchema(ctx context.Context, databaseAlias, dbName string) error {
	// Get database connection for the specified alias
	// For now, we'll use the main DB connection
	// In production, you would get the connection from a connection pool based on alias

	// Get all tables
	var tables []struct {
		TableName    string
		TableComment string
	}

	query := `
		SELECT
			TABLE_NAME as tablename,
			COALESCE(TABLE_COMMENT, '') as table_comment
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`

	if err := s.db.Raw(query, dbName).Scan(&tables).Error; err != nil {
		return fmt.Errorf("failed to discover tables: %w", err)
	}

	// Process each table
	for _, table := range tables {
		// Create or update table metadata
		tableMeta := &models.DatabaseSchemaMetadata{
			DatabaseAlias: databaseAlias,
			Table:         table.TableName,
			MetadataType:  models.MetadataTypeTable,
			Description:   table.TableComment,
		}

		if err := s.db.WithContext(ctx).
			Where("databasealias = ? AND tablename = ? AND metadatatype = ?", databaseAlias, table.TableName, models.MetadataTypeTable).
			Assign(tableMeta).
			FirstOrCreate(tableMeta).Error; err != nil {
			return fmt.Errorf("failed to save table metadata for %s: %w", table.TableName, err)
		}

		// Get columns for this table
		var columns []struct {
			ColumnName    string
			DataType      string
			IsNullable    string
			ColumnKey     string
			ColumnComment string
		}

		columnQuery := `
			SELECT
				COLUMN_NAME as columnname,
				DATA_TYPE as data_type,
				IS_NULLABLE as is_nullable,
				COLUMN_KEY as column_key,
				COALESCE(COLUMN_COMMENT, '') as column_comment
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = ?
			AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION
		`

		if err := s.db.Raw(columnQuery, dbName, table.TableName).Scan(&columns).Error; err != nil {
			return fmt.Errorf("failed to discover columns for %s: %w", table.TableName, err)
		}

		// Process each column
		for _, column := range columns {
			isNullable := column.IsNullable == "YES"

			columnMeta := &models.DatabaseSchemaMetadata{
				DatabaseAlias: databaseAlias,
				Table:         table.TableName,
				Column:        column.ColumnName,
				MetadataType:  models.MetadataTypeColumn,
				DataType:      column.DataType,
				IsNullable:    &isNullable,
				ColumnComment: column.ColumnComment,
			}

			if err := s.db.WithContext(ctx).
				Where("databasealias = ? AND tablename = ? AND columnname = ? AND metadatatype = 'column'",
					databaseAlias, table.TableName, column.ColumnName).
				Assign(columnMeta).
				FirstOrCreate(columnMeta).Error; err != nil {
				return fmt.Errorf("failed to save column metadata for %s.%s: %w", table.TableName, column.ColumnName, err)
			}
		}
	}

	return nil
}

// GetTableMetadata retrieves metadata for all tables in a database
func (s *SchemaMetadataService) GetTableMetadata(ctx context.Context, databaseAlias string) ([]models.DatabaseSchemaMetadata, error) {
	var metadata []models.DatabaseSchemaMetadata

	if err := s.db.WithContext(ctx).
		Where("databasealias = ? AND metadatatype = ?", databaseAlias, models.MetadataTypeTable).
		Order("tablename").
		Find(&metadata).Error; err != nil {
		return nil, fmt.Errorf("failed to get table metadata: %w", err)
	}

	return metadata, nil
}

// GetColumnMetadata retrieves metadata for all columns in a table
func (s *SchemaMetadataService) GetColumnMetadata(ctx context.Context, databaseAlias, tableName string) ([]models.DatabaseSchemaMetadata, error) {
	var metadata []models.DatabaseSchemaMetadata

	if err := s.db.WithContext(ctx).
		Where("databasealias = ? AND tablename = ? AND metadatatype = ?", databaseAlias, tableName, models.MetadataTypeColumn).
		Order("columnname").
		Find(&metadata).Error; err != nil {
		return nil, fmt.Errorf("failed to get column metadata: %w", err)
	}

	return metadata, nil
}

// UpdateMetadata updates the description for a metadata entry
func (s *SchemaMetadataService) UpdateMetadata(ctx context.Context, id string, description *string) error {
	updates := make(map[string]interface{})

	if description != nil {
		updates["description"] = *description
	}

	if len(updates) == 0 {
		return nil
	}

	updates["modifiedon"] = gorm.Expr("CURRENT_TIMESTAMP")

	if err := s.db.WithContext(ctx).
		Model(&models.DatabaseSchemaMetadata{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}

// DeleteMetadata deletes a metadata entry
func (s *SchemaMetadataService) DeleteMetadata(ctx context.Context, id string) error {
	if err := s.db.WithContext(ctx).
		Delete(&models.DatabaseSchemaMetadata{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}

// GetSchemaContext builds a comprehensive schema context for AI
func (s *SchemaMetadataService) GetSchemaContext(ctx context.Context, databaseAlias string, tableNames []string) (string, error) {
	var metadata []models.DatabaseSchemaMetadata

	query := s.db.WithContext(ctx).Where("databasealias = ?", databaseAlias)

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

// GetAllDatabases returns a list of all unique database aliases
func (s *SchemaMetadataService) GetAllDatabases(ctx context.Context) ([]string, error) {
	var databases []string

	if err := s.db.WithContext(ctx).
		Model(&models.DatabaseSchemaMetadata{}).
		Distinct("databasealias").
		Pluck("databasealias", &databases).Error; err != nil {
		return nil, fmt.Errorf("failed to get databases: %w", err)
	}

	return databases, nil
}

// SearchMetadata searches metadata by keyword
func (s *SchemaMetadataService) SearchMetadata(ctx context.Context, databaseAlias, keyword string) ([]models.DatabaseSchemaMetadata, error) {
	var metadata []models.DatabaseSchemaMetadata

	searchPattern := "%" + keyword + "%"

	if err := s.db.WithContext(ctx).
		Where("databasealias = ?", databaseAlias).
		Where("tablename LIKE ? OR COALESCE(columnname, '') LIKE ? OR COALESCE(description, '') LIKE ? OR COALESCE(businessterms, '') LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern).
		Order("tablename, metadatatype DESC, columnname").
		Find(&metadata).Error; err != nil {
		return nil, fmt.Errorf("failed to search metadata: %w", err)
	}

	return metadata, nil
}
