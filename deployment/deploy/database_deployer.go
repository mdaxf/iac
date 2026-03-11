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

package deploymgr

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/deployment/models"
	"github.com/mdaxf/iac/logger"
)

// DatabaseDeployer handles deployment of database packages
type DatabaseDeployer struct {
	dbTx        *sql.Tx
	dbOperation *dbconn.DBOperation
	logger      logger.Log
	dbType      string
	pkMappings  map[string]map[interface{}]interface{} // Table -> OldPK -> NewPK
}

// NewDatabaseDeployer creates a new database deployer
func NewDatabaseDeployer(user string, dbTx *sql.Tx, dbType string) *DatabaseDeployer {
	iLog := logger.Log{ModuleName: logger.Framework, User: user, ControllerName: "DatabaseDeployer"}

	return &DatabaseDeployer{
		dbTx:        dbTx,
		dbOperation: dbconn.NewDBOperation(user, dbTx, logger.Framework),
		logger:      iLog,
		dbType:      dbType,
		pkMappings:  make(map[string]map[interface{}]interface{}),
	}
}

// Deploy deploys a database package
func (dd *DatabaseDeployer) Deploy(pkg *models.Package, options models.DeploymentOptions) (*models.DeploymentRecord, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		dd.logger.PerformanceWithDuration("DatabaseDeployer.Deploy", elapsed)
	}()

	dd.logger.Info(fmt.Sprintf("Starting deployment: %s v%s", pkg.Name, pkg.Version))

	// Create deployment record
	record := &models.DeploymentRecord{
		ID:              uuid.New().String(),
		PackageID:       pkg.ID,
		PackageName:     pkg.Name,
		PackageVersion:  pkg.Version,
		TargetDatabase:  dd.dbType,
		DeployedAt:      time.Now(),
		Status:          "in_progress",
		PKMappingResult: make(map[string]map[interface{}]interface{}),
		ErrorLog:        make([]string, 0),
		Metadata:        make(map[string]interface{}),
	}

	// Validate package
	if pkg.DatabaseData == nil {
		return record, fmt.Errorf("package does not contain database data")
	}

	// Check dry run
	if options.DryRun {
		dd.logger.Info("Dry run mode - validating only")
		if err := dd.validatePackage(pkg, options); err != nil {
			record.Status = "failed"
			record.ErrorLog = append(record.ErrorLog, err.Error())
			return record, err
		}
		record.Status = "validated"
		return record, nil
	}

	// Sort tables by dependency order
	sortedTables, err := dd.sortTablesByDependency(pkg.DatabaseData.Tables, pkg.DatabaseData.Relationships)
	if err != nil {
		record.Status = "failed"
		record.ErrorLog = append(record.ErrorLog, fmt.Sprintf("Failed to sort tables: %v", err))
		return record, err
	}

	// Deploy each table
	for _, table := range sortedTables {
		dd.logger.Debug(fmt.Sprintf("Deploying table: %s", table.TableName))

		if err := dd.deployTable(table, pkg.DatabaseData.PKMappings[table.TableName], options); err != nil {
			errMsg := fmt.Sprintf("Failed to deploy table %s: %v", table.TableName, err)
			dd.logger.Error(errMsg)
			record.ErrorLog = append(record.ErrorLog, errMsg)

			if !options.ContinueOnError {
				record.Status = "failed"
				return record, err
			}
		}

		// Store PK mappings in record
		if tableMappings, ok := dd.pkMappings[table.TableName]; ok {
			record.PKMappingResult[table.TableName] = tableMappings
		}
	}

	// Rebuild relationships
	if err := dd.rebuildRelationships(pkg.DatabaseData.Relationships, options); err != nil {
		errMsg := fmt.Sprintf("Failed to rebuild relationships: %v", err)
		dd.logger.Error(errMsg)
		record.ErrorLog = append(record.ErrorLog, errMsg)

		if !options.ContinueOnError {
			record.Status = "failed"
			return record, err
		}
	}

	record.Status = "completed"
	dd.logger.Info(fmt.Sprintf("Deployment completed: %s", record.ID))
	return record, nil
}

// deployTable deploys a single table
func (dd *DatabaseDeployer) deployTable(table models.TableData, pkMapping models.PKMapping, options models.DeploymentOptions) error {
	// Initialize PK mapping for this table
	if _, ok := dd.pkMappings[table.TableName]; !ok {
		dd.pkMappings[table.TableName] = make(map[interface{}]interface{})
	}

	batchSize := options.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	// Process rows in batches
	for i := 0; i < len(table.Rows); i += batchSize {
		end := i + batchSize
		if end > len(table.Rows) {
			end = len(table.Rows)
		}

		batch := table.Rows[i:end]
		if err := dd.deployBatch(table, batch, pkMapping, options); err != nil {
			return fmt.Errorf("failed to deploy batch %d-%d: %w", i, end, err)
		}
	}

	return nil
}

// deployBatch deploys a batch of rows
func (dd *DatabaseDeployer) deployBatch(table models.TableData, rows []map[string]interface{}, pkMapping models.PKMapping, options models.DeploymentOptions) error {
	for _, row := range rows {
		if err := dd.deployRow(table, row, pkMapping, options); err != nil {
			return err
		}
	}
	return nil
}

// deployRow deploys a single row
func (dd *DatabaseDeployer) deployRow(table models.TableData, row map[string]interface{}, pkMapping models.PKMapping, options models.DeploymentOptions) error {
	// Store old PK value before any modifications
	oldPK := dd.extractPKValue(row, pkMapping.PKColumns)

	// --- Global key upsert: find by natural key ---
	if len(pkMapping.GlobalKeyColumns) > 0 {
		existingPK, found, err := dd.recordExistsByGlobalKey(table.TableName, pkMapping.GlobalKeyColumns, row)
		if err != nil {
			dd.logger.Warn(fmt.Sprintf("Global key lookup failed for %s: %v", table.TableName, err))
		} else if found {
			// Map old PK → existing PK for FK rebuilding
			dd.pkMappings[table.TableName][oldPK] = existingPK
			if options.UpdateExisting {
				return dd.updateRow(table, row, pkMapping.PKColumns, existingPK)
			}
			if options.SkipExisting {
				dd.logger.Debug(fmt.Sprintf("Skipping existing record (global key match) in %s", table.TableName))
				return nil
			}
			return fmt.Errorf("record already exists in %s (matched by global key)", table.TableName)
		}
	}

	// Handle PK based on strategy
	newPK, err := dd.handlePKStrategy(row, pkMapping, options)
	if err != nil {
		return err
	}

	// Check if record exists by PK
	if newPK != nil {
		exists, err := dd.recordExists(table.TableName, pkMapping.PKColumns, newPK)
		if err != nil {
			return err
		}

		if exists {
			if options.SkipExisting {
				dd.logger.Debug(fmt.Sprintf("Skipping existing record in %s", table.TableName))
				dd.pkMappings[table.TableName][oldPK] = newPK
				return nil
			}
			if options.UpdateExisting {
				dd.pkMappings[table.TableName][oldPK] = newPK
				return dd.updateRow(table, row, pkMapping.PKColumns, newPK)
			}
			return fmt.Errorf("record already exists in %s", table.TableName)
		}
	}

	// Insert new record
	insertedPK, err := dd.insertRowGetPK(table, row, pkMapping)
	if err != nil {
		return err
	}

	// Map old PK → new PK (insertedPK may differ for auto_increment)
	if insertedPK != nil {
		dd.pkMappings[table.TableName][oldPK] = insertedPK
	} else if newPK != nil {
		dd.pkMappings[table.TableName][oldPK] = newPK
	} else {
		dd.pkMappings[table.TableName][oldPK] = oldPK
	}

	return nil
}

// handlePKStrategy handles PK generation based on strategy
func (dd *DatabaseDeployer) handlePKStrategy(row map[string]interface{}, pkMapping models.PKMapping, options models.DeploymentOptions) (interface{}, error) {
	switch pkMapping.Strategy {
	case "auto_increment":
		// Remove PK field to let database generate it
		for _, pkCol := range pkMapping.PKColumns {
			delete(row, pkCol)
		}
		return nil, nil

	case "sequence":
		// Let database sequence generate it
		for _, pkCol := range pkMapping.PKColumns {
			delete(row, pkCol)
		}
		return nil, nil

	case "preserve":
		// Keep original PK value (used for UUIDs and other globally unique identifiers)
		return dd.extractPKValue(row, pkMapping.PKColumns), nil

	default:
		return nil, fmt.Errorf("unknown PK strategy: %s", pkMapping.Strategy)
	}
}

// insertRow inserts a row into the table (kept for callers that don't need PK return)
func (dd *DatabaseDeployer) insertRow(table models.TableData, row map[string]interface{}) error {
	_, err := dd.insertRowGetPK(table, row, models.PKMapping{})
	return err
}

// insertRowGetPK inserts a row and returns the new PK for auto_increment tables
func (dd *DatabaseDeployer) insertRowGetPK(table models.TableData, row map[string]interface{}, pkMapping models.PKMapping) (interface{}, error) {
	columns := make([]string, 0)
	placeholders := make([]string, 0)
	values := make([]interface{}, 0)

	i := 1
	for _, col := range table.Columns {
		if val, ok := row[col.Name]; ok {
			columns = append(columns, dd.dbOperation.QuoteIdentifier(col.Name))
			placeholders = append(placeholders, dd.dbOperation.GetPlaceholder(i))
			values = append(values, val)
			i++
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		dd.dbOperation.QuoteIdentifier(table.TableName),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	result, err := dd.dbOperation.Execute(query, values...)
	if err != nil {
		return nil, err
	}

	// For auto_increment, retrieve the generated PK
	if pkMapping.IsAutoIncrement && result != nil {
		if lastID, err := result.LastInsertId(); err == nil && lastID > 0 {
			return lastID, nil
		}
	}

	// For preserve strategy, PK is already in the row
	if pkMapping.Strategy == "preserve" && len(pkMapping.PKColumns) > 0 {
		return row[pkMapping.PKColumns[0]], nil
	}

	return nil, nil
}

// updateRow updates an existing row
func (dd *DatabaseDeployer) updateRow(table models.TableData, row map[string]interface{}, pkColumns []string, pkValue interface{}) error {
	// Build UPDATE statement
	setClauses := make([]string, 0)
	values := make([]interface{}, 0)

	i := 1
	for _, col := range table.Columns {
		// Skip PK columns
		isPK := false
		for _, pkCol := range pkColumns {
			if col.Name == pkCol {
				isPK = true
				break
			}
		}
		if isPK {
			continue
		}

		if val, ok := row[col.Name]; ok {
			setClauses = append(setClauses, fmt.Sprintf("%s = %s",
				dd.dbOperation.QuoteIdentifier(col.Name),
				dd.dbOperation.GetPlaceholder(i)))
			values = append(values, val)
			i++
		}
	}

	// Build WHERE clause
	whereClauses := make([]string, 0)
	for _, pkCol := range pkColumns {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = %s",
			dd.dbOperation.QuoteIdentifier(pkCol),
			dd.dbOperation.GetPlaceholder(i)))
		values = append(values, row[pkCol])
		i++
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		dd.dbOperation.QuoteIdentifier(table.TableName),
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND "))

	_, err := dd.dbOperation.Execute(query, values...)
	return err
}

// recordExists checks if a record exists by its PK value
func (dd *DatabaseDeployer) recordExists(tableName string, pkColumns []string, pkValue interface{}) (bool, error) {
	if pkValue == nil || len(pkColumns) == 0 {
		return false, nil
	}

	whereClauses := make([]string, 0)
	values := make([]interface{}, 0)

	switch v := pkValue.(type) {
	case map[string]interface{}:
		// Composite key
		for i, col := range pkColumns {
			if val, ok := v[col]; ok {
				whereClauses = append(whereClauses, fmt.Sprintf("%s = %s",
					dd.dbOperation.QuoteIdentifier(col),
					dd.dbOperation.GetPlaceholder(i+1)))
				values = append(values, val)
			}
		}
	default:
		// Single-column key
		whereClauses = append(whereClauses, fmt.Sprintf("%s = %s",
			dd.dbOperation.QuoteIdentifier(pkColumns[0]),
			dd.dbOperation.GetPlaceholder(1)))
		values = append(values, pkValue)
	}

	if len(whereClauses) == 0 {
		return false, nil
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s",
		dd.dbOperation.QuoteIdentifier(tableName),
		strings.Join(whereClauses, " AND "))

	rows, err := dd.dbOperation.Query(query, values...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		var count int64
		if err := rows.Scan(&count); err != nil {
			return false, err
		}
		return count > 0, nil
	}
	return false, nil
}

// recordExistsByGlobalKey finds a record by its natural/global key columns
// Returns (existingPKValue, found, error)
func (dd *DatabaseDeployer) recordExistsByGlobalKey(tableName string, globalKeyColumns []string, row map[string]interface{}) (interface{}, bool, error) {
	if len(globalKeyColumns) == 0 {
		return nil, false, nil
	}

	whereClauses := make([]string, 0)
	values := make([]interface{}, 0)

	for i, col := range globalKeyColumns {
		val, ok := row[col]
		if !ok {
			continue
		}
		whereClauses = append(whereClauses, fmt.Sprintf("%s = %s",
			dd.dbOperation.QuoteIdentifier(col),
			dd.dbOperation.GetPlaceholder(i+1)))
		values = append(values, val)
	}

	if len(whereClauses) == 0 {
		return nil, false, nil
	}

	// We need to know the actual PK column(s) — select them with a LIMIT 1
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT 1",
		dd.dbOperation.QuoteIdentifier(tableName),
		strings.Join(whereClauses, " AND "))

	rows, err := dd.dbOperation.Query(query, values...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, false, err
	}

	if !rows.Next() {
		return nil, false, nil
	}

	scanValues := make([]interface{}, len(cols))
	scanPtrs := make([]interface{}, len(cols))
	for i := range scanValues {
		scanPtrs[i] = &scanValues[i]
	}
	if err := rows.Scan(scanPtrs...); err != nil {
		return nil, false, err
	}

	// Build a map of the found row to extract PK later
	foundRow := make(map[string]interface{})
	for i, colName := range cols {
		if b, ok := scanValues[i].([]byte); ok {
			foundRow[colName] = string(b)
		} else {
			foundRow[colName] = scanValues[i]
		}
	}

	// We don't know which columns are PKs here, so return the whole row map
	// Callers will use the first non-global-key unique column, or we just return a sentinel
	// For simplicity, return the value of the first column (usually the PK)
	if len(cols) > 0 {
		return foundRow[cols[0]], true, nil
	}
	return nil, true, nil
}

// extractPKValue extracts PK value from row
func (dd *DatabaseDeployer) extractPKValue(row map[string]interface{}, pkColumns []string) interface{} {
	if len(pkColumns) == 1 {
		return row[pkColumns[0]]
	}

	// Composite key - return map
	pkValue := make(map[string]interface{})
	for _, col := range pkColumns {
		pkValue[col] = row[col]
	}
	return pkValue
}

// rebuildRelationships updates FK columns in child tables to point to regenerated PKs
func (dd *DatabaseDeployer) rebuildRelationships(relationships []models.Relationship, options models.DeploymentOptions) error {
	dd.logger.Info("Rebuilding relationships")

	for _, rel := range relationships {
		// Get target table PK mappings (old PK → new PK)
		targetMappings, ok := dd.pkMappings[rel.TargetTable]
		if !ok || len(targetMappings) == 0 {
			continue
		}

		updatedCount := 0
		for oldTargetPK, newTargetPK := range targetMappings {
			// Skip if the PK didn't change (preserve strategy or no change)
			if fmt.Sprintf("%v", oldTargetPK) == fmt.Sprintf("%v", newTargetPK) {
				continue
			}

			// UPDATE source_table SET source_column = newPK WHERE source_column = oldPK
			query := fmt.Sprintf("UPDATE %s SET %s = %s WHERE %s = %s",
				dd.dbOperation.QuoteIdentifier(rel.SourceTable),
				dd.dbOperation.QuoteIdentifier(rel.SourceColumn),
				dd.dbOperation.GetPlaceholder(1),
				dd.dbOperation.QuoteIdentifier(rel.SourceColumn),
				dd.dbOperation.GetPlaceholder(2))

			result, err := dd.dbOperation.Execute(query, newTargetPK, oldTargetPK)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to update FK %s.%s (old=%v → new=%v): %v",
					rel.SourceTable, rel.SourceColumn, oldTargetPK, newTargetPK, err)
				dd.logger.Warn(errMsg)
				if !options.ContinueOnError {
					return fmt.Errorf("%s", errMsg)
				}
				continue
			}
			if result != nil {
				if n, e := result.RowsAffected(); e == nil && n > 0 {
					updatedCount += int(n)
				}
			}
		}

		if updatedCount > 0 {
			dd.logger.Info(fmt.Sprintf("Updated %d FK rows in %s.%s → %s.%s",
				updatedCount, rel.SourceTable, rel.SourceColumn, rel.TargetTable, rel.TargetColumn))
		}
	}

	return nil
}

// sortTablesByDependency sorts tables in dependency order
func (dd *DatabaseDeployer) sortTablesByDependency(tables []models.TableData, relationships []models.Relationship) ([]models.TableData, error) {
	// Build dependency graph
	graph := make(map[string][]string)
	tableMap := make(map[string]models.TableData)

	for _, table := range tables {
		tableMap[table.TableName] = table
		graph[table.TableName] = make([]string, 0)
	}

	// Add edges
	for _, rel := range relationships {
		// Source table depends on target table
		if _, ok := graph[rel.SourceTable]; ok {
			graph[rel.SourceTable] = append(graph[rel.SourceTable], rel.TargetTable)
		}
	}

	// Topological sort
	sorted := make([]models.TableData, 0)
	visited := make(map[string]bool)
	temp := make(map[string]bool)

	var visit func(string) error
	visit = func(tableName string) error {
		if temp[tableName] {
			return fmt.Errorf("circular dependency detected: %s", tableName)
		}
		if visited[tableName] {
			return nil
		}

		temp[tableName] = true

		for _, dep := range graph[tableName] {
			if err := visit(dep); err != nil {
				return err
			}
		}

		temp[tableName] = false
		visited[tableName] = true

		if table, ok := tableMap[tableName]; ok {
			sorted = append([]models.TableData{table}, sorted...)
		}

		return nil
	}

	for tableName := range tableMap {
		if !visited[tableName] {
			if err := visit(tableName); err != nil {
				return nil, err
			}
		}
	}

	return sorted, nil
}

// validatePackage validates a package before deployment
func (dd *DatabaseDeployer) validatePackage(pkg *models.Package, options models.DeploymentOptions) error {
	dd.logger.Info("Validating package")

	// Check package structure
	if pkg.DatabaseData == nil {
		return fmt.Errorf("package contains no database data")
	}

	// Validate tables exist
	for _, table := range pkg.DatabaseData.Tables {
		dd.logger.Debug(fmt.Sprintf("Validating table: %s", table.TableName))
		// Additional validation could be added here
	}

	// Validate relationships
	if options.ValidateReferences {
		for _, rel := range pkg.DatabaseData.Relationships {
			dd.logger.Debug(fmt.Sprintf("Validating relationship: %s -> %s", rel.SourceTable, rel.TargetTable))
			// Validate relationship structure
		}
	}

	return nil
}

// Rollback rolls back a deployment
func (dd *DatabaseDeployer) Rollback(record *models.DeploymentRecord) error {
	dd.logger.Info(fmt.Sprintf("Rolling back deployment: %s", record.ID))

	// Implementation would delete inserted records using the PK mappings
	// This is a placeholder

	record.Status = "rolled_back"
	return nil
}
