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
	PackageType    string                 `db:"package_type" json:"package_type"`
	Description    string                 `db:"description" json:"description"`
	CreatedAt      time.Time              `db:"created_at" json:"created_at"`
	CreatedBy      string                 `db:"created_by" json:"created_by"`
	Metadata       string                 `db:"metadata" json:"metadata"` // JSON string
	PackageData    string                 `db:"package_data" json:"package_data"` // JSON string
	DatabaseType   string                 `db:"database_type" json:"database_type"`
	DatabaseName   string                 `db:"database_name" json:"database_name"`
	IncludeParent  bool                   `db:"include_parent" json:"include_parent"`
	Dependencies   string                 `db:"dependencies" json:"dependencies"` // JSON string
	Checksum       string                 `db:"checksum" json:"checksum"`
	FileSize       int64                  `db:"file_size" json:"file_size"`
	Status         string                 `db:"status" json:"status"`
	Tags           string                 `db:"tags" json:"tags"` // JSON string
	Environment    string                 `db:"environment" json:"environment"`
	UpdatedAt      time.Time              `db:"updated_at" json:"updated_at"`
}

// PackageActionRecord represents a package action (pack/deploy/rollback)
type PackageActionRecord struct {
	ID                   string    `db:"id" json:"id"`
	PackageID            string    `db:"package_id" json:"package_id"`
	ActionType           string    `db:"action_type" json:"action_type"`
	ActionStatus         string    `db:"action_status" json:"action_status"`
	TargetDatabase       string    `db:"target_database" json:"target_database"`
	TargetEnvironment    string    `db:"target_environment" json:"target_environment"`
	SourceEnvironment    string    `db:"source_environment" json:"source_environment"`
	PerformedAt          time.Time `db:"performed_at" json:"performed_at"`
	PerformedBy          string    `db:"performed_by" json:"performed_by"`
	StartedAt            *time.Time `db:"started_at" json:"started_at"`
	CompletedAt          *time.Time `db:"completed_at" json:"completed_at"`
	DurationSeconds      int       `db:"duration_seconds" json:"duration_seconds"`
	Options              string    `db:"options" json:"options"` // JSON string
	ResultData           string    `db:"result_data" json:"result_data"` // JSON string
	ErrorLog             string    `db:"error_log" json:"error_log"` // JSON string
	WarningLog           string    `db:"warning_log" json:"warning_log"` // JSON string
	Metadata             string    `db:"metadata" json:"metadata"` // JSON string
	RecordsProcessed     int       `db:"records_processed" json:"records_processed"`
	RecordsSucceeded     int       `db:"records_succeeded" json:"records_succeeded"`
	RecordsFailed        int       `db:"records_failed" json:"records_failed"`
	TablesProcessed      int       `db:"tables_processed" json:"tables_processed"`
	CollectionsProcessed int       `db:"collections_processed" json:"collections_processed"`
}

// PackageRelationship represents relationships between packages
type PackageRelationship struct {
	ID               string    `db:"id" json:"id"`
	ParentPackageID  string    `db:"parent_package_id" json:"parent_package_id"`
	ChildPackageID   string    `db:"child_package_id" json:"child_package_id"`
	RelationshipType string    `db:"relationship_type" json:"relationship_type"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

// PackageDeployment tracks active deployments
type PackageDeployment struct {
	ID             string     `db:"id" json:"id"`
	PackageID      string     `db:"package_id" json:"package_id"`
	ActionID       string     `db:"action_id" json:"action_id"`
	Environment    string     `db:"environment" json:"environment"`
	DatabaseName   string     `db:"database_name" json:"database_name"`
	DeployedAt     time.Time  `db:"deployed_at" json:"deployed_at"`
	DeployedBy     string     `db:"deployed_by" json:"deployed_by"`
	IsActive       bool       `db:"is_active" json:"is_active"`
	RolledBackAt   *time.Time `db:"rolled_back_at" json:"rolled_back_at"`
	RolledBackBy   string     `db:"rolled_back_by" json:"rolled_back_by"`
}

// PackageTag represents package tags
type PackageTag struct {
	ID        string    `db:"id" json:"id"`
	PackageID string    `db:"package_id" json:"package_id"`
	TagName   string    `db:"tag_name" json:"tag_name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	CreatedBy string    `db:"created_by" json:"created_by"`
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
