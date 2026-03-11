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
	"context"
	"database/sql"
	"fmt"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/logger"
)

// PackageDefinitionRepository handles SQL operations for package definitions and builds.
type PackageDefinitionRepository struct {
	dbTx   *sql.Tx
	dbOp   *dbconn.DBOperation
	logger logger.Log
}

// NewPackageDefinitionRepository creates a new repository for definition + build records.
func NewPackageDefinitionRepository(user string, dbTx *sql.Tx) *PackageDefinitionRepository {
	iLog := logger.Log{ModuleName: logger.Framework, User: user, ControllerName: "PackageDefinitionRepository"}
	return &PackageDefinitionRepository{
		dbTx:   dbTx,
		dbOp:   dbconn.NewDBOperation(user, dbTx, logger.Framework),
		logger: iLog,
	}
}

// SaveDefinition inserts a new record into iacpackagedefs.
func (r *PackageDefinitionRepository) SaveDefinition(rec *PackageDefinitionRecord) error {
	now := time.Now()
	if rec.CreatedOn.IsZero() {
		rec.CreatedOn = now
	}
	if rec.ModifiedOn.IsZero() {
		rec.ModifiedOn = now
	}
	if rec.LatestVersion == 0 {
		rec.LatestVersion = 1
	}
	if rec.RowVersionStamp == 0 {
		rec.RowVersionStamp = 1
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (id, name, packagetype, description, environment, latestversion, active, createdby, createdon, modifiedby, modifiedon, rowversionstamp) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)`,
		r.dbOp.QuoteIdentifier("iacpackagedefs"),
		r.dbOp.GetPlaceholder(1), r.dbOp.GetPlaceholder(2), r.dbOp.GetPlaceholder(3),
		r.dbOp.GetPlaceholder(4), r.dbOp.GetPlaceholder(5), r.dbOp.GetPlaceholder(6),
		r.dbOp.GetPlaceholder(7), r.dbOp.GetPlaceholder(8), r.dbOp.GetPlaceholder(9),
		r.dbOp.GetPlaceholder(10), r.dbOp.GetPlaceholder(11), r.dbOp.GetPlaceholder(12),
	)

	_, err := r.dbOp.Execute(query,
		rec.ID, rec.Name, rec.PackageType, rec.Description, rec.Environment,
		rec.LatestVersion, rec.Active, rec.CreatedBy, rec.CreatedOn,
		rec.ModifiedBy, rec.ModifiedOn, rec.RowVersionStamp,
	)
	if err != nil {
		r.logger.Error(fmt.Sprintf("SaveDefinition failed: %v", err))
	}
	return err
}

// UpdateDefinition updates name, packagetype, description, environment, latestversion and audit fields.
func (r *PackageDefinitionRepository) UpdateDefinition(rec *PackageDefinitionRecord) error {
	if rec.ModifiedOn.IsZero() {
		rec.ModifiedOn = time.Now()
	}

	query := fmt.Sprintf(
		`UPDATE %s SET name=%s, packagetype=%s, description=%s, environment=%s, latestversion=%s, modifiedby=%s, modifiedon=%s WHERE id=%s`,
		r.dbOp.QuoteIdentifier("iacpackagedefs"),
		r.dbOp.GetPlaceholder(1), r.dbOp.GetPlaceholder(2), r.dbOp.GetPlaceholder(3),
		r.dbOp.GetPlaceholder(4), r.dbOp.GetPlaceholder(5), r.dbOp.GetPlaceholder(6),
		r.dbOp.GetPlaceholder(7), r.dbOp.GetPlaceholder(8),
	)

	_, err := r.dbOp.Execute(query,
		rec.Name, rec.PackageType, rec.Description, rec.Environment,
		rec.LatestVersion, rec.ModifiedBy, rec.ModifiedOn, rec.ID,
	)
	if err != nil {
		r.logger.Error(fmt.Sprintf("UpdateDefinition failed: %v", err))
	}
	return err
}

// DeleteDefinition soft-deletes a definition row (sets active=false).
func (r *PackageDefinitionRepository) DeleteDefinition(id, user string) error {
	now := time.Now()
	query := fmt.Sprintf(
		`UPDATE %s SET active=%s, modifiedby=%s, modifiedon=%s WHERE id=%s`,
		r.dbOp.QuoteIdentifier("iacpackagedefs"),
		r.dbOp.GetPlaceholder(1), r.dbOp.GetPlaceholder(2),
		r.dbOp.GetPlaceholder(3), r.dbOp.GetPlaceholder(4),
	)
	_, err := r.dbOp.Execute(query, false, user, now, id)
	if err != nil {
		r.logger.Error(fmt.Sprintf("DeleteDefinition failed: %v", err))
	}
	return err
}

// GetNextBuildNumber returns MAX(buildnumber)+1 for the given definitionid.
// Returns 1 if no builds exist yet.
func (r *PackageDefinitionRepository) GetNextBuildNumber(definitionID string) (int, error) {
	// Use dbTx directly — DBOperation.Query() defers rows.Close() before returning,
	// making the rows unusable by callers.
	q := fmt.Sprintf(
		`SELECT COALESCE(MAX(buildnumber), 0) + 1 FROM %s WHERE definitionid = %s`,
		r.dbOp.QuoteIdentifier("iacpackagebuilds"),
		r.dbOp.GetPlaceholder(1),
	)
	rows, err := r.dbTx.QueryContext(context.Background(), q, definitionID)
	if err != nil {
		return 1, err
	}
	defer rows.Close()

	next := 1
	if rows.Next() {
		if scanErr := rows.Scan(&next); scanErr != nil {
			return 1, scanErr
		}
	}
	if next < 1 {
		next = 1
	}
	return next, nil
}

// SaveBuild inserts a new build record into iacpackagebuilds.
func (r *PackageDefinitionRepository) SaveBuild(rec *PackageBuildRecord) error {
	now := time.Now()
	if rec.CreatedOn.IsZero() {
		rec.CreatedOn = now
	}
	if rec.ModifiedOn.IsZero() {
		rec.ModifiedOn = now
	}
	if rec.RowVersionStamp == 0 {
		rec.RowVersionStamp = 1
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (id, definitionid, definitionname, defversion, buildnumber, packageid, packagename, packageversion, status, startedat, completedat, durationseconds, triggeredby, errorlog, metadata, active, createdby, createdon, modifiedby, modifiedon, rowversionstamp) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)`,
		r.dbOp.QuoteIdentifier("iacpackagebuilds"),
		r.dbOp.GetPlaceholder(1), r.dbOp.GetPlaceholder(2), r.dbOp.GetPlaceholder(3),
		r.dbOp.GetPlaceholder(4), r.dbOp.GetPlaceholder(5), r.dbOp.GetPlaceholder(6),
		r.dbOp.GetPlaceholder(7), r.dbOp.GetPlaceholder(8), r.dbOp.GetPlaceholder(9),
		r.dbOp.GetPlaceholder(10), r.dbOp.GetPlaceholder(11), r.dbOp.GetPlaceholder(12),
		r.dbOp.GetPlaceholder(13), r.dbOp.GetPlaceholder(14), r.dbOp.GetPlaceholder(15),
		r.dbOp.GetPlaceholder(16), r.dbOp.GetPlaceholder(17), r.dbOp.GetPlaceholder(18),
		r.dbOp.GetPlaceholder(19), r.dbOp.GetPlaceholder(20), r.dbOp.GetPlaceholder(21),
	)

	_, err := r.dbOp.Execute(query,
		rec.ID, rec.DefinitionID, rec.DefinitionName, rec.DefVersion, rec.BuildNumber,
		rec.PackageID, rec.PackageName, rec.PackageVersion, rec.Status,
		rec.StartedAt, rec.CompletedAt, rec.DurationSeconds, rec.TriggeredBy,
		rec.ErrorLog, rec.Metadata, rec.Active,
		rec.CreatedBy, rec.CreatedOn, rec.ModifiedBy, rec.ModifiedOn, rec.RowVersionStamp,
	)
	if err != nil {
		r.logger.Error(fmt.Sprintf("SaveBuild failed: %v", err))
	}
	return err
}

// ListBuilds returns all build records for a definition, ordered by build number descending.
func (r *PackageDefinitionRepository) ListBuilds(definitionID string) ([]*PackageBuildRecord, error) {
	// Use dbTx directly — DBOperation.Query() defers rows.Close() before returning,
	// making the rows unusable by callers.
	q := fmt.Sprintf(
		`SELECT id, definitionid, definitionname, defversion, buildnumber, packageid, packagename, packageversion, status, startedat, completedat, durationseconds, triggeredby, errorlog, active, createdby, createdon FROM %s WHERE definitionid = %s AND active = %s ORDER BY buildnumber DESC`,
		r.dbOp.QuoteIdentifier("iacpackagebuilds"),
		r.dbOp.GetPlaceholder(1), r.dbOp.GetPlaceholder(2),
	)
	rows, err := r.dbTx.QueryContext(context.Background(), q, definitionID, true)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var builds []*PackageBuildRecord
	for rows.Next() {
		rec := &PackageBuildRecord{Active: true}
		scanErr := rows.Scan(
			&rec.ID, &rec.DefinitionID, &rec.DefinitionName, &rec.DefVersion,
			&rec.BuildNumber, &rec.PackageID, &rec.PackageName, &rec.PackageVersion,
			&rec.Status, &rec.StartedAt, &rec.CompletedAt, &rec.DurationSeconds,
			&rec.TriggeredBy, &rec.ErrorLog, &rec.Active, &rec.CreatedBy, &rec.CreatedOn,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		builds = append(builds, rec)
	}
	if builds == nil {
		builds = []*PackageBuildRecord{}
	}
	return builds, nil
}
