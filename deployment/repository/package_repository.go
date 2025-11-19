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

package repository

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/deployment/models"
	"github.com/mdaxf/iac/logger"
)

// PackageRepository handles database operations for packages
type PackageRepository struct {
	dbTx   *sql.Tx
	dbOp   *dbconn.DBOperation
	logger logger.Log
}

// NewPackageRepository creates a new package repository
func NewPackageRepository(user string, dbTx *sql.Tx) *PackageRepository {
	iLog := logger.Log{ModuleName: logger.Framework, User: user, ControllerName: "PackageRepository"}

	return &PackageRepository{
		dbTx:   dbTx,
		dbOp:   dbconn.NewDBOperation(user, dbTx, logger.Framework),
		logger: iLog,
	}
}

// SavePackage saves a package to the database
func (pr *PackageRepository) SavePackage(pkg *models.Package, environment string, filter *models.PackageFilter) (*PackageRecord, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		pr.logger.PerformanceWithDuration("PackageRepository.SavePackage", elapsed)
	}()

	pr.logger.Info(fmt.Sprintf("Saving package: %s v%s", pkg.Name, pkg.Version))

	// Serialize package data
	packageData, err := json.Marshal(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize package: %w", err)
	}

	// Calculate checksum
	hash := sha256.Sum256(packageData)
	checksum := hex.EncodeToString(hash[:])

	// Build enhanced metadata with packaged entities and selection criteria
	enhancedMetadata := pr.buildPackageMetadata(pkg, filter)

	// Merge with existing metadata
	if pkg.Metadata != nil {
		for k, v := range pkg.Metadata {
			enhancedMetadata[k] = v
		}
	}

	// Serialize metadata
	metadataJSON, _ := json.Marshal(enhancedMetadata)
	depsJSON, _ := json.Marshal(pkg.Dependencies)

	record := &PackageRecord{
		ID:            pkg.ID,
		Name:          pkg.Name,
		Version:       pkg.Version,
		PackageType:   pkg.PackageType,
		Description:   pkg.Description,
		Metadata:      string(metadataJSON),
		PackageData:   string(packageData),
		Checksum:      checksum,
		FileSize:      int64(len(packageData)),
		Status:        PackageStatusActive,
		Environment:   environment,
		IncludeParent: pkg.IncludeParent,
		Dependencies:  string(depsJSON),
		// IAC Standard Fields
		Active:          true,
		CreatedBy:       pkg.CreatedBy,
		CreatedOn:       pkg.CreatedAt,
		ModifiedBy:      pkg.CreatedBy,
		ModifiedOn:      pkg.CreatedAt,
		RowVersionStamp: 1,
	}

	// Set database-specific fields
	if pkg.DatabaseData != nil {
		record.DatabaseType = pkg.DatabaseData.DatabaseType
		// Database name might not be in DatabaseData, get from metadata if available
		if dbName, ok := pkg.Metadata["database_name"].(string); ok {
			record.DatabaseName = dbName
		}
	} else if pkg.DocumentData != nil {
		record.DatabaseType = pkg.DocumentData.DatabaseType
		record.DatabaseName = pkg.DocumentData.DatabaseName
	}

	// Insert package record
	query := pr.buildInsertPackageQuery()
	_, err = pr.dbOp.Exec(query,
		record.ID,
		record.Name,
		record.Version,
		record.PackageType,
		record.Description,
		record.Metadata,
		record.PackageData,
		record.DatabaseType,
		record.DatabaseName,
		record.IncludeParent,
		record.Dependencies,
		record.Checksum,
		record.FileSize,
		record.Status,
		"", // tags - empty for now
		record.Environment,
		// IAC Standard Fields
		record.Active,
		record.ReferenceID,
		record.CreatedBy,
		record.CreatedOn,
		record.ModifiedBy,
		record.ModifiedOn,
		record.RowVersionStamp,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert package: %w", err)
	}

	pr.logger.Info(fmt.Sprintf("Package saved: %s (checksum: %s, size: %d bytes)", record.ID, checksum[:8], record.FileSize))
	return record, nil
}

// GetPackage retrieves a package by ID
func (pr *PackageRepository) GetPackage(packageID string) (*models.Package, error) {
	query := fmt.Sprintf(`
		SELECT packagedata
		FROM %s
		WHERE id = %s AND status != %s AND active = %s`,
		pr.dbOp.QuoteIdentifier("iacpackages"),
		pr.dbOp.GetPlaceholder(1),
		pr.dbOp.GetPlaceholder(2),
		pr.dbOp.GetPlaceholder(3))

	rows, err := pr.dbOp.Query(query, packageID, PackageStatusDeleted, true)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("package not found: %s", packageID)
	}

	var packageData string
	if err := rows.Scan(&packageData); err != nil {
		return nil, err
	}

	var pkg models.Package
	if err := json.Unmarshal([]byte(packageData), &pkg); err != nil {
		return nil, fmt.Errorf("failed to deserialize package: %w", err)
	}

	return &pkg, nil
}

// GetPackageByNameVersion retrieves a package by name and version
func (pr *PackageRepository) GetPackageByNameVersion(name, version string) (*models.Package, error) {
	query := fmt.Sprintf(`
		SELECT packagedata
		FROM %s
		WHERE name = %s AND version = %s AND status != %s AND active = %s`,
		pr.dbOp.QuoteIdentifier("iacpackages"),
		pr.dbOp.GetPlaceholder(1),
		pr.dbOp.GetPlaceholder(2),
		pr.dbOp.GetPlaceholder(3),
		pr.dbOp.GetPlaceholder(4))

	rows, err := pr.dbOp.Query(query, name, version, PackageStatusDeleted, true)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("package not found: %s v%s", name, version)
	}

	var packageData string
	if err := rows.Scan(&packageData); err != nil {
		return nil, err
	}

	var pkg models.Package
	if err := json.Unmarshal([]byte(packageData), &pkg); err != nil {
		return nil, fmt.Errorf("failed to deserialize package: %w", err)
	}

	return &pkg, nil
}

// ListPackages lists packages with optional filters
func (pr *PackageRepository) ListPackages(packageType, environment, status string, limit, offset int) ([]PackageRecord, error) {
	query := fmt.Sprintf(`
		SELECT id, name, version, packagetype, description, createdby,
		       databasetype, databasename, checksum, filesize, status, environment,
		       active, referenceid, createdon, modifiedby, modifiedon, rowversionstamp
		FROM %s
		WHERE active = %s`,
		pr.dbOp.QuoteIdentifier("iacpackages"),
		pr.dbOp.GetPlaceholder(1))

	args := make([]interface{}, 0)
	args = append(args, true)
	paramIdx := 2

	if packageType != "" {
		query += fmt.Sprintf(" AND packagetype = %s", pr.dbOp.GetPlaceholder(paramIdx))
		args = append(args, packageType)
		paramIdx++
	}

	if environment != "" {
		query += fmt.Sprintf(" AND environment = %s", pr.dbOp.GetPlaceholder(paramIdx))
		args = append(args, environment)
		paramIdx++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = %s", pr.dbOp.GetPlaceholder(paramIdx))
		args = append(args, status)
		paramIdx++
	} else {
		query += fmt.Sprintf(" AND status != %s", pr.dbOp.GetPlaceholder(paramIdx))
		args = append(args, PackageStatusDeleted)
		paramIdx++
	}

	query += " ORDER BY createdon DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %s", pr.dbOp.GetPlaceholder(paramIdx))
		args = append(args, limit)
		paramIdx++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %s", pr.dbOp.GetPlaceholder(paramIdx))
		args = append(args, offset)
	}

	rows, err := pr.dbOp.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	packages := make([]PackageRecord, 0)
	for rows.Next() {
		var pkg PackageRecord
		err := rows.Scan(
			&pkg.ID,
			&pkg.Name,
			&pkg.Version,
			&pkg.PackageType,
			&pkg.Description,
			&pkg.CreatedBy,
			&pkg.DatabaseType,
			&pkg.DatabaseName,
			&pkg.Checksum,
			&pkg.FileSize,
			&pkg.Status,
			&pkg.Environment,
			&pkg.Active,
			&pkg.ReferenceID,
			&pkg.CreatedOn,
			&pkg.ModifiedBy,
			&pkg.ModifiedOn,
			&pkg.RowVersionStamp,
		)
		if err != nil {
			pr.logger.Warn(fmt.Sprintf("Error scanning package: %v", err))
			continue
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}

// SaveAction saves a package action record
func (pr *PackageRepository) SaveAction(action *PackageActionRecord) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		pr.logger.PerformanceWithDuration("PackageRepository.SaveAction", elapsed)
	}()

	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	if action.PerformedAt.IsZero() {
		action.PerformedAt = time.Now()
	}

	// Set IAC standard fields
	action.Active = true
	if action.CreatedOn.IsZero() {
		action.CreatedOn = time.Now()
	}
	if action.ModifiedOn.IsZero() {
		action.ModifiedOn = time.Now()
	}
	if action.RowVersionStamp == 0 {
		action.RowVersionStamp = 1
	}

	query := pr.buildInsertActionQuery()
	_, err := pr.dbOp.Exec(query,
		action.ID,
		action.PackageID,
		action.ActionType,
		action.ActionStatus,
		action.TargetDatabase,
		action.TargetEnvironment,
		action.SourceEnvironment,
		action.PerformedAt,
		action.PerformedBy,
		action.StartedAt,
		action.CompletedAt,
		action.DurationSeconds,
		action.Options,
		action.ResultData,
		action.ErrorLog,
		action.WarningLog,
		action.Metadata,
		action.RecordsProcessed,
		action.RecordsSucceeded,
		action.RecordsFailed,
		action.TablesProcessed,
		action.CollectionsProcessed,
		// IAC Standard Fields
		action.Active,
		action.ReferenceID,
		action.CreatedBy,
		action.CreatedOn,
		action.ModifiedBy,
		action.ModifiedOn,
		action.RowVersionStamp,
	)

	if err != nil {
		return fmt.Errorf("failed to insert action: %w", err)
	}

	pr.logger.Info(fmt.Sprintf("Action saved: %s (type: %s, status: %s)", action.ID, action.ActionType, action.ActionStatus))
	return nil
}

// UpdateActionStatus updates the status of an action
func (pr *PackageRepository) UpdateActionStatus(actionID, status string, completedAt *time.Time, errorLog []string) error {
	errLogJSON, _ := json.Marshal(errorLog)

	query := fmt.Sprintf(`
		UPDATE %s
		SET actionstatus = %s, completedat = %s, errorlog = %s
		WHERE id = %s`,
		pr.dbOp.QuoteIdentifier("packageactions"),
		pr.dbOp.GetPlaceholder(1),
		pr.dbOp.GetPlaceholder(2),
		pr.dbOp.GetPlaceholder(3),
		pr.dbOp.GetPlaceholder(4))

	_, err := pr.dbOp.Exec(query, status, completedAt, string(errLogJSON), actionID)
	return err
}

// GetActionsByPackage retrieves all actions for a package
func (pr *PackageRepository) GetActionsByPackage(packageID string, limit int) ([]PackageActionRecord, error) {
	query := fmt.Sprintf(`
		SELECT id, packageid, actiontype, actionstatus, targetdatabase, targetenvironment,
		       performedat, performedby, startedat, completedat, durationseconds,
		       recordsprocessed, recordssucceeded, recordsfailed,
		       tablesprocessed, collectionsprocessed,
		       active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM %s
		WHERE packageid = %s AND active = %s
		ORDER BY performedat DESC`,
		pr.dbOp.QuoteIdentifier("packageactions"),
		pr.dbOp.GetPlaceholder(1),
		pr.dbOp.GetPlaceholder(2))

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %s", pr.dbOp.GetPlaceholder(3))
	}

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = pr.dbOp.Query(query, packageID, true, limit)
	} else {
		rows, err = pr.dbOp.Query(query, packageID, true)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actions := make([]PackageActionRecord, 0)
	for rows.Next() {
		var action PackageActionRecord
		err := rows.Scan(
			&action.ID,
			&action.PackageID,
			&action.ActionType,
			&action.ActionStatus,
			&action.TargetDatabase,
			&action.TargetEnvironment,
			&action.PerformedAt,
			&action.PerformedBy,
			&action.StartedAt,
			&action.CompletedAt,
			&action.DurationSeconds,
			&action.RecordsProcessed,
			&action.RecordsSucceeded,
			&action.RecordsFailed,
			&action.TablesProcessed,
			&action.CollectionsProcessed,
			&action.Active,
			&action.ReferenceID,
			&action.CreatedBy,
			&action.CreatedOn,
			&action.ModifiedBy,
			&action.ModifiedOn,
			&action.RowVersionStamp,
		)
		if err != nil {
			pr.logger.Warn(fmt.Sprintf("Error scanning action: %v", err))
			continue
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// SaveDeployment records a successful deployment
func (pr *PackageRepository) SaveDeployment(deployment *PackageDeployment) error {
	if deployment.ID == "" {
		deployment.ID = uuid.New().String()
	}

	if deployment.DeployedAt.IsZero() {
		deployment.DeployedAt = time.Now()
	}

	if deployment.CreatedOn.IsZero() {
		deployment.CreatedOn = time.Now()
	}

	if deployment.ModifiedOn.IsZero() {
		deployment.ModifiedOn = time.Now()
	}

	deployment.Active = true
	if deployment.RowVersionStamp == 0 {
		deployment.RowVersionStamp = 1
	}

	query := fmt.Sprintf(`
		INSERT INTO %s
		(id, packageid, actionid, environment, databasename, deployedat, deployedby, isactive,
		 active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp)
		VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)`,
		pr.dbOp.QuoteIdentifier("packagedeployments"),
		pr.dbOp.GetPlaceholder(1),
		pr.dbOp.GetPlaceholder(2),
		pr.dbOp.GetPlaceholder(3),
		pr.dbOp.GetPlaceholder(4),
		pr.dbOp.GetPlaceholder(5),
		pr.dbOp.GetPlaceholder(6),
		pr.dbOp.GetPlaceholder(7),
		pr.dbOp.GetPlaceholder(8),
		pr.dbOp.GetPlaceholder(9),
		pr.dbOp.GetPlaceholder(10),
		pr.dbOp.GetPlaceholder(11),
		pr.dbOp.GetPlaceholder(12),
		pr.dbOp.GetPlaceholder(13),
		pr.dbOp.GetPlaceholder(14),
		pr.dbOp.GetPlaceholder(15))

	_, err := pr.dbOp.Exec(query,
		deployment.ID,
		deployment.PackageID,
		deployment.ActionID,
		deployment.Environment,
		deployment.DatabaseName,
		deployment.DeployedAt,
		deployment.DeployedBy,
		deployment.IsActive,
		deployment.Active,
		deployment.ReferenceID,
		deployment.CreatedBy,
		deployment.CreatedOn,
		deployment.ModifiedBy,
		deployment.ModifiedOn,
		deployment.RowVersionStamp,
	)

	return err
}

// buildPackageMetadata builds comprehensive metadata about packaged entities and selection criteria
func (pr *PackageRepository) buildPackageMetadata(pkg *models.Package, filter *models.PackageFilter) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Package type specific metadata
	if pkg.PackageType == "database" && pkg.DatabaseData != nil {
		entities := make([]map[string]interface{}, 0)
		totalRecords := 0

		for _, table := range pkg.DatabaseData.Tables {
			entity := map[string]interface{}{
				"name":           table.TableName,
				"type":           "table",
				"row_count":      table.RowCount,
				"column_count":   len(table.Columns),
				"pk_columns":     table.PKColumns,
				"fk_count":       len(table.FKColumns),
				"pk_strategy":    pkg.DatabaseData.PKMappings[table.TableName].Strategy,
			}

			// Add WHERE clause if present
			if filter != nil && filter.WhereClause != nil {
				if whereClause, ok := filter.WhereClause[table.TableName]; ok {
					entity["where_clause"] = whereClause
				}
			}

			// Add excluded columns if present
			if filter != nil && filter.ExcludeColumns != nil {
				if excludedCols, ok := filter.ExcludeColumns[table.TableName]; ok && len(excludedCols) > 0 {
					entity["excluded_columns"] = excludedCols
				}
			}

			entities = append(entities, entity)
			totalRecords += table.RowCount
		}

		metadata["packaged_entities"] = entities
		metadata["entity_count"] = len(entities)
		metadata["total_records"] = totalRecords
		metadata["total_relationships"] = len(pkg.DatabaseData.Relationships)

		// Selection criteria
		if filter != nil {
			selectionCriteria := make(map[string]interface{})
			if len(filter.Tables) > 0 {
				selectionCriteria["tables"] = filter.Tables
			}
			if filter.IncludeRelated {
				selectionCriteria["include_related"] = true
				selectionCriteria["max_depth"] = filter.MaxDepth
			}
			if len(filter.WhereClause) > 0 {
				selectionCriteria["where_clauses"] = filter.WhereClause
			}
			if len(filter.ExcludeColumns) > 0 {
				selectionCriteria["excluded_columns"] = filter.ExcludeColumns
			}
			metadata["selection_criteria"] = selectionCriteria
		}

	} else if pkg.PackageType == "document" && pkg.DocumentData != nil {
		entities := make([]map[string]interface{}, 0)
		totalDocuments := 0

		for _, collection := range pkg.DocumentData.Collections {
			entity := map[string]interface{}{
				"name":           collection.CollectionName,
				"type":           "collection",
				"document_count": collection.DocumentCount,
				"index_count":    len(collection.IndexInfo),
				"id_field":       collection.IDField,
				"id_strategy":    pkg.DocumentData.IDMappings[collection.CollectionName].Strategy,
			}

			// Add query filter if present
			if filter != nil && filter.WhereClause != nil {
				if queryFilter, ok := filter.WhereClause[collection.CollectionName]; ok {
					entity["query_filter"] = queryFilter
				}
			}

			// Add excluded fields if present
			if filter != nil && filter.ExcludeFields != nil {
				if excludedFields, ok := filter.ExcludeFields[collection.CollectionName]; ok && len(excludedFields) > 0 {
					entity["excluded_fields"] = excludedFields
				}
			}

			entities = append(entities, entity)
			totalDocuments += collection.DocumentCount
		}

		metadata["packaged_entities"] = entities
		metadata["entity_count"] = len(entities)
		metadata["total_documents"] = totalDocuments
		metadata["total_references"] = len(pkg.DocumentData.References)

		// Selection criteria
		if filter != nil {
			selectionCriteria := make(map[string]interface{})
			if len(filter.Collections) > 0 {
				selectionCriteria["collections"] = filter.Collections
			}
			if len(filter.WhereClause) > 0 {
				selectionCriteria["query_filters"] = filter.WhereClause
			}
			if len(filter.ExcludeFields) > 0 {
				selectionCriteria["excluded_fields"] = filter.ExcludeFields
			}
			metadata["selection_criteria"] = selectionCriteria
		}
	}

	// Common metadata
	metadata["include_parent_data"] = pkg.IncludeParent
	if len(pkg.Dependencies) > 0 {
		metadata["has_dependencies"] = true
		metadata["dependency_count"] = len(pkg.Dependencies)
	}

	return metadata
}

// Helper functions

func (pr *PackageRepository) buildInsertPackageQuery() string {
	return fmt.Sprintf(`
		INSERT INTO %s
		(id, name, version, packagetype, description, metadata,
		 packagedata, databasetype, databasename, includeparent, dependencies,
		 checksum, filesize, status, tags, environment,
		 active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp)
		VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)`,
		pr.dbOp.QuoteIdentifier("iacpackages"),
		pr.dbOp.GetPlaceholder(1),
		pr.dbOp.GetPlaceholder(2),
		pr.dbOp.GetPlaceholder(3),
		pr.dbOp.GetPlaceholder(4),
		pr.dbOp.GetPlaceholder(5),
		pr.dbOp.GetPlaceholder(6),
		pr.dbOp.GetPlaceholder(7),
		pr.dbOp.GetPlaceholder(8),
		pr.dbOp.GetPlaceholder(9),
		pr.dbOp.GetPlaceholder(10),
		pr.dbOp.GetPlaceholder(11),
		pr.dbOp.GetPlaceholder(12),
		pr.dbOp.GetPlaceholder(13),
		pr.dbOp.GetPlaceholder(14),
		pr.dbOp.GetPlaceholder(15),
		pr.dbOp.GetPlaceholder(16),
		pr.dbOp.GetPlaceholder(17),
		pr.dbOp.GetPlaceholder(18),
		pr.dbOp.GetPlaceholder(19),
		pr.dbOp.GetPlaceholder(20),
		pr.dbOp.GetPlaceholder(21),
		pr.dbOp.GetPlaceholder(22),
		pr.dbOp.GetPlaceholder(23))
}

func (pr *PackageRepository) buildInsertActionQuery() string {
	return fmt.Sprintf(`
		INSERT INTO %s
		(id, packageid, actiontype, actionstatus, targetdatabase, targetenvironment,
		 sourceenvironment, performedat, performedby, startedat, completedat, durationseconds,
		 options, resultdata, errorlog, warninglog, metadata, recordsprocessed,
		 recordssucceeded, recordsfailed, tablesprocessed, collectionsprocessed,
		 active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp)
		VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)`,
		pr.dbOp.QuoteIdentifier("packageactions"),
		pr.dbOp.GetPlaceholder(1),
		pr.dbOp.GetPlaceholder(2),
		pr.dbOp.GetPlaceholder(3),
		pr.dbOp.GetPlaceholder(4),
		pr.dbOp.GetPlaceholder(5),
		pr.dbOp.GetPlaceholder(6),
		pr.dbOp.GetPlaceholder(7),
		pr.dbOp.GetPlaceholder(8),
		pr.dbOp.GetPlaceholder(9),
		pr.dbOp.GetPlaceholder(10),
		pr.dbOp.GetPlaceholder(11),
		pr.dbOp.GetPlaceholder(12),
		pr.dbOp.GetPlaceholder(13),
		pr.dbOp.GetPlaceholder(14),
		pr.dbOp.GetPlaceholder(15),
		pr.dbOp.GetPlaceholder(16),
		pr.dbOp.GetPlaceholder(17),
		pr.dbOp.GetPlaceholder(18),
		pr.dbOp.GetPlaceholder(19),
		pr.dbOp.GetPlaceholder(20),
		pr.dbOp.GetPlaceholder(21),
		pr.dbOp.GetPlaceholder(22),
		pr.dbOp.GetPlaceholder(23),
		pr.dbOp.GetPlaceholder(24),
		pr.dbOp.GetPlaceholder(25),
		pr.dbOp.GetPlaceholder(26),
		pr.dbOp.GetPlaceholder(27),
		pr.dbOp.GetPlaceholder(28),
		pr.dbOp.GetPlaceholder(29))
}
