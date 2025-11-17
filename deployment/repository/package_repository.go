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
func (pr *PackageRepository) SavePackage(pkg *models.Package, environment string) (*PackageRecord, error) {
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

	// Serialize metadata
	metadataJSON, _ := json.Marshal(pkg.Metadata)
	depsJSON, _ := json.Marshal(pkg.Dependencies)

	record := &PackageRecord{
		ID:            pkg.ID,
		Name:          pkg.Name,
		Version:       pkg.Version,
		PackageType:   pkg.PackageType,
		Description:   pkg.Description,
		CreatedAt:     pkg.CreatedAt,
		CreatedBy:     pkg.CreatedBy,
		Metadata:      string(metadataJSON),
		PackageData:   string(packageData),
		Checksum:      checksum,
		FileSize:      int64(len(packageData)),
		Status:        PackageStatusActive,
		Environment:   environment,
		IncludeParent: pkg.IncludeParent,
		Dependencies:  string(depsJSON),
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
		record.CreatedAt,
		record.CreatedBy,
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
		SELECT package_data
		FROM %s
		WHERE id = %s AND status != %s`,
		pr.dbOp.QuoteIdentifier("iacpackages"),
		pr.dbOp.GetPlaceholder(1),
		pr.dbOp.GetPlaceholder(2))

	rows, err := pr.dbOp.Query(query, packageID, PackageStatusDeleted)
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
		SELECT package_data
		FROM %s
		WHERE name = %s AND version = %s AND status != %s`,
		pr.dbOp.QuoteIdentifier("iacpackages"),
		pr.dbOp.GetPlaceholder(1),
		pr.dbOp.GetPlaceholder(2),
		pr.dbOp.GetPlaceholder(3))

	rows, err := pr.dbOp.Query(query, name, version, PackageStatusDeleted)
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
		SELECT id, name, version, package_type, description, created_at, created_by,
		       database_type, database_name, checksum, file_size, status, environment
		FROM %s
		WHERE 1=1`,
		pr.dbOp.QuoteIdentifier("iacpackages"))

	args := make([]interface{}, 0)
	paramIdx := 1

	if packageType != "" {
		query += fmt.Sprintf(" AND package_type = %s", pr.dbOp.GetPlaceholder(paramIdx))
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

	query += " ORDER BY created_at DESC"

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
			&pkg.CreatedAt,
			&pkg.CreatedBy,
			&pkg.DatabaseType,
			&pkg.DatabaseName,
			&pkg.Checksum,
			&pkg.FileSize,
			&pkg.Status,
			&pkg.Environment,
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
		SET action_status = %s, completed_at = %s, error_log = %s
		WHERE id = %s`,
		pr.dbOp.QuoteIdentifier("package_actions"),
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
		SELECT id, package_id, action_type, action_status, target_database, target_environment,
		       performed_at, performed_by, started_at, completed_at, duration_seconds,
		       records_processed, records_succeeded, records_failed,
		       tables_processed, collections_processed
		FROM %s
		WHERE package_id = %s
		ORDER BY performed_at DESC`,
		pr.dbOp.QuoteIdentifier("package_actions"),
		pr.dbOp.GetPlaceholder(1))

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %s", pr.dbOp.GetPlaceholder(2))
	}

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = pr.dbOp.Query(query, packageID, limit)
	} else {
		rows, err = pr.dbOp.Query(query, packageID)
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

	query := fmt.Sprintf(`
		INSERT INTO %s
		(id, package_id, action_id, environment, database_name, deployed_at, deployed_by, is_active)
		VALUES (%s, %s, %s, %s, %s, %s, %s, %s)`,
		pr.dbOp.QuoteIdentifier("package_deployments"),
		pr.dbOp.GetPlaceholder(1),
		pr.dbOp.GetPlaceholder(2),
		pr.dbOp.GetPlaceholder(3),
		pr.dbOp.GetPlaceholder(4),
		pr.dbOp.GetPlaceholder(5),
		pr.dbOp.GetPlaceholder(6),
		pr.dbOp.GetPlaceholder(7),
		pr.dbOp.GetPlaceholder(8))

	_, err := pr.dbOp.Exec(query,
		deployment.ID,
		deployment.PackageID,
		deployment.ActionID,
		deployment.Environment,
		deployment.DatabaseName,
		deployment.DeployedAt,
		deployment.DeployedBy,
		deployment.IsActive,
	)

	return err
}

// Helper functions

func (pr *PackageRepository) buildInsertPackageQuery() string {
	return fmt.Sprintf(`
		INSERT INTO %s
		(id, name, version, package_type, description, created_at, created_by, metadata,
		 package_data, database_type, database_name, include_parent, dependencies,
		 checksum, file_size, status, tags, environment)
		VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)`,
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
		pr.dbOp.GetPlaceholder(18))
}

func (pr *PackageRepository) buildInsertActionQuery() string {
	return fmt.Sprintf(`
		INSERT INTO %s
		(id, package_id, action_type, action_status, target_database, target_environment,
		 source_environment, performed_at, performed_by, started_at, completed_at, duration_seconds,
		 options, result_data, error_log, warning_log, metadata, records_processed,
		 records_succeeded, records_failed, tables_processed, collections_processed)
		VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)`,
		pr.dbOp.QuoteIdentifier("package_actions"),
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
		pr.dbOp.GetPlaceholder(22))
}
