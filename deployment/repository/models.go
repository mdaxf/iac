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
	"time"
)

// PackageRecord represents a package stored in the database
type PackageRecord struct {
	ID             string                 `db:"id" json:"id"`
	Name           string                 `db:"name" json:"name"`
	Version        string                 `db:"version" json:"version"`
	PackageType    string                 `db:"packagetype" json:"packagetype"`
	Description    string                 `db:"description" json:"description"`
	Metadata       string                 `db:"metadata" json:"metadata"` // JSON string
	PackageData    string                 `db:"packagedata" json:"packagedata"` // JSON string
	DatabaseType   string                 `db:"databasetype" json:"databasetype"`
	DatabaseName   string                 `db:"databasename" json:"databasename"`
	IncludeParent  bool                   `db:"includeparent" json:"includeparent"`
	Dependencies   string                 `db:"dependencies" json:"dependencies"` // JSON string
	Checksum       string                 `db:"checksum" json:"checksum"`
	FileSize       int64                  `db:"filesize" json:"filesize"`
	Status         string                 `db:"status" json:"status"`
	Tags           string                 `db:"tags" json:"tags"` // JSON string
	Environment    string                 `db:"environment" json:"environment"`
	// IAC Standard Fields
	Active          bool      `db:"active" json:"active"`
	ReferenceID     string    `db:"referenceid" json:"referenceid"`
	CreatedBy       string    `db:"createdby" json:"createdby"`
	CreatedOn       time.Time `db:"createdon" json:"createdon"`
	ModifiedBy      string    `db:"modifiedby" json:"modifiedby"`
	ModifiedOn      time.Time `db:"modifiedon" json:"modifiedon"`
	RowVersionStamp int       `db:"rowversionstamp" json:"rowversionstamp"`
}

// PackageActionRecord represents a package action (pack/deploy/rollback)
type PackageActionRecord struct {
	ID                   string    `db:"id" json:"id"`
	PackageID            string    `db:"packageid" json:"packageid"`
	ActionType           string    `db:"actiontype" json:"actiontype"`
	ActionStatus         string    `db:"actionstatus" json:"actionstatus"`
	TargetDatabase       string    `db:"targetdatabase" json:"targetdatabase"`
	TargetEnvironment    string    `db:"targetenvironment" json:"targetenvironment"`
	SourceEnvironment    string    `db:"sourceenvironment" json:"sourceenvironment"`
	PerformedAt          time.Time `db:"performedat" json:"performedat"`
	PerformedBy          string    `db:"performedby" json:"performedby"`
	StartedAt            *time.Time `db:"startedat" json:"startedat"`
	CompletedAt          *time.Time `db:"completedat" json:"completedat"`
	DurationSeconds      int       `db:"durationseconds" json:"durationseconds"`
	Options              string    `db:"options" json:"options"` // JSON string
	ResultData           string    `db:"resultdata" json:"resultdata"` // JSON string
	ErrorLog             string    `db:"errorlog" json:"errorlog"` // JSON string
	WarningLog           string    `db:"warninglog" json:"warninglog"` // JSON string
	Metadata             string    `db:"metadata" json:"metadata"` // JSON string
	RecordsProcessed     int       `db:"recordsprocessed" json:"recordsprocessed"`
	RecordsSucceeded     int       `db:"recordssucceeded" json:"recordssucceeded"`
	RecordsFailed        int       `db:"recordsfailed" json:"recordsfailed"`
	TablesProcessed      int       `db:"tablesprocessed" json:"tablesprocessed"`
	CollectionsProcessed int       `db:"collectionsprocessed" json:"collectionsprocessed"`
	// IAC Standard Fields
	Active          bool      `db:"active" json:"active"`
	ReferenceID     string    `db:"referenceid" json:"referenceid"`
	CreatedBy       string    `db:"createdby" json:"createdby"`
	CreatedOn       time.Time `db:"createdon" json:"createdon"`
	ModifiedBy      string    `db:"modifiedby" json:"modifiedby"`
	ModifiedOn      time.Time `db:"modifiedon" json:"modifiedon"`
	RowVersionStamp int       `db:"rowversionstamp" json:"rowversionstamp"`
}

// PackageRelationship represents relationships between packages
type PackageRelationship struct {
	ID               string    `db:"id" json:"id"`
	ParentPackageID  string    `db:"parentpackageid" json:"parentpackageid"`
	ChildPackageID   string    `db:"childpackageid" json:"childpackageid"`
	RelationshipType string    `db:"relationshiptype" json:"relationshiptype"`
	// IAC Standard Fields
	Active          bool      `db:"active" json:"active"`
	ReferenceID     string    `db:"referenceid" json:"referenceid"`
	CreatedBy       string    `db:"createdby" json:"createdby"`
	CreatedOn       time.Time `db:"createdon" json:"createdon"`
	ModifiedBy      string    `db:"modifiedby" json:"modifiedby"`
	ModifiedOn      time.Time `db:"modifiedon" json:"modifiedon"`
	RowVersionStamp int       `db:"rowversionstamp" json:"rowversionstamp"`
}

// PackageDeployment tracks active deployments
type PackageDeployment struct {
	ID             string     `db:"id" json:"id"`
	PackageID      string     `db:"packageid" json:"packageid"`
	ActionID       string     `db:"actionid" json:"actionid"`
	Environment    string     `db:"environment" json:"environment"`
	DatabaseName   string     `db:"databasename" json:"databasename"`
	DeployedAt     time.Time  `db:"deployedat" json:"deployedat"`
	DeployedBy     string     `db:"deployedby" json:"deployedby"`
	IsActive       bool       `db:"isactive" json:"isactive"`
	RolledBackAt   *time.Time `db:"rolledbackat" json:"rolledbackat"`
	RolledBackBy   string     `db:"rolledbackby" json:"rolledbackby"`
	// IAC Standard Fields
	Active          bool      `db:"active" json:"active"`
	ReferenceID     string    `db:"referenceid" json:"referenceid"`
	CreatedBy       string    `db:"createdby" json:"createdby"`
	CreatedOn       time.Time `db:"createdon" json:"createdon"`
	ModifiedBy      string    `db:"modifiedby" json:"modifiedby"`
	ModifiedOn      time.Time `db:"modifiedon" json:"modifiedon"`
	RowVersionStamp int       `db:"rowversionstamp" json:"rowversionstamp"`
}

// PackageTag represents package tags
type PackageTag struct {
	ID        string    `db:"id" json:"id"`
	PackageID string    `db:"packageid" json:"packageid"`
	TagName   string    `db:"tagname" json:"tagname"`
	// IAC Standard Fields
	Active          bool      `db:"active" json:"active"`
	ReferenceID     string    `db:"referenceid" json:"referenceid"`
	CreatedBy       string    `db:"createdby" json:"createdby"`
	CreatedOn       time.Time `db:"createdon" json:"createdon"`
	ModifiedBy      string    `db:"modifiedby" json:"modifiedby"`
	ModifiedOn      time.Time `db:"modifiedon" json:"modifiedon"`
	RowVersionStamp int       `db:"rowversionstamp" json:"rowversionstamp"`
}

// PackageWithActions combines package with its recent actions
type PackageWithActions struct {
	PackageRecord
	RecentActions []PackageActionRecord `json:"recent_actions"`
}

// ActionType constants
const (
	ActionTypePack     = "pack"
	ActionTypeDeploy   = "deploy"
	ActionTypeRollback = "rollback"
	ActionTypeExport   = "export"
	ActionTypeImport   = "import"
	ActionTypeValidate = "validate"
)

// ActionStatus constants
const (
	ActionStatusPending    = "pending"
	ActionStatusInProgress = "in_progress"
	ActionStatusCompleted  = "completed"
	ActionStatusFailed     = "failed"
	ActionStatusRolledBack = "rolled_back"
)

// PackageStatus constants
const (
	PackageStatusActive   = "active"
	PackageStatusArchived = "archived"
	PackageStatusDeleted  = "deleted"
)

// RelationshipType constants
const (
	RelationshipDependsOn = "depends_on"
	RelationshipExtends   = "extends"
	RelationshipIncludes  = "includes"
)
