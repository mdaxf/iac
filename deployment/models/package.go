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

package models

import (
	"time"
)

// Package represents a deployable package containing database or document data
type Package struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Version       string                 `json:"version"`
	PackageType   string                 `json:"package_type"` // "database" or "document"
	CreatedAt     time.Time              `json:"created_at"`
	CreatedBy     string                 `json:"created_by"`
	Metadata      map[string]interface{} `json:"metadata"`
	DatabaseData  *DatabasePackage       `json:"database_data,omitempty"`
	DocumentData  *DocumentPackage       `json:"document_data,omitempty"`
	Dependencies  []string               `json:"dependencies,omitempty"`  // Package IDs that must be deployed first
	IncludeParent bool                   `json:"include_parent"`          // Auto-include parent records
}

// DatabasePackage contains relational database data with PK mapping
type DatabasePackage struct {
	Tables         []TableData            `json:"tables"`
	PKMappings     map[string]PKMapping   `json:"pk_mappings"`      // Table -> PK mapping config
	Relationships  []Relationship         `json:"relationships"`    // FK relationships
	SequenceInfo   map[string]int64       `json:"sequence_info"`    // Next sequence values per table
	DatabaseType   string                 `json:"database_type"`    // mysql, postgresql, mssql, oracle
	SchemaVersion  string                 `json:"schema_version"`
}

// DocumentPackage contains document database data with ID handling
type DocumentPackage struct {
	Collections    []CollectionData       `json:"collections"`
	IDMappings     map[string]IDMapping   `json:"id_mappings"`      // Collection -> ID mapping config
	References     []DocumentReference    `json:"references"`       // Document references
	SkipIDs        bool                   `json:"skip_ids"`         // Skip ID fields during pack
	DatabaseType   string                 `json:"database_type"`    // mongodb, etc.
	DatabaseName   string                 `json:"database_name"`
}

// TableData represents data from a single table
type TableData struct {
	TableName     string                   `json:"table_name"`
	Schema        string                   `json:"schema,omitempty"`
	Columns       []ColumnInfo             `json:"columns"`
	Rows          []map[string]interface{} `json:"rows"`
	RowCount      int                      `json:"row_count"`
	PKColumns     []string                 `json:"pk_columns"`
	FKColumns     []ForeignKeyInfo         `json:"fk_columns,omitempty"`
}

// CollectionData represents data from a MongoDB collection
type CollectionData struct {
	CollectionName string                   `json:"collection_name"`
	Documents      []map[string]interface{} `json:"documents"`
	DocumentCount  int                      `json:"document_count"`
	IDField        string                   `json:"id_field"`          // Usually "_id"
	IndexInfo      []IndexInfo              `json:"index_info,omitempty"`
}

// ColumnInfo describes a database column
type ColumnInfo struct {
	Name         string `json:"name"`
	DataType     string `json:"data_type"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	IsForeignKey bool   `json:"is_foreign_key"`
	IsNullable   bool   `json:"is_nullable"`
	MaxLength    int    `json:"max_length,omitempty"`
}

// ForeignKeyInfo describes a foreign key relationship
type ForeignKeyInfo struct {
	ColumnName        string `json:"column_name"`
	ReferencedTable   string `json:"referenced_table"`
	ReferencedColumn  string `json:"referenced_column"`
	ConstraintName    string `json:"constraint_name"`
}

// IndexInfo describes a MongoDB index
type IndexInfo struct {
	Name   string                 `json:"name"`
	Keys   map[string]interface{} `json:"keys"`
	Unique bool                   `json:"unique"`
}

// PKMapping defines how primary keys should be handled
type PKMapping struct {
	TableName       string   `json:"table_name"`
	PKColumns       []string `json:"pk_columns"`
	IsAutoIncrement bool     `json:"is_auto_increment"`
	SequenceName    string   `json:"sequence_name,omitempty"`
	Strategy        string   `json:"strategy"` // "auto_increment", "sequence", "uuid", "preserve"
}

// IDMapping defines how document IDs should be handled
type IDMapping struct {
	CollectionName  string `json:"collection_name"`
	IDField         string `json:"id_field"`
	IDType          string `json:"id_type"`     // "objectid", "uuid", "string", "int"
	Strategy        string `json:"strategy"`    // "regenerate", "preserve", "skip"
}

// Relationship tracks foreign key relationships for rebuilding
type Relationship struct {
	ID                string `json:"id"`
	SourceTable       string `json:"source_table"`
	SourceColumn      string `json:"source_column"`
	TargetTable       string `json:"target_table"`
	TargetColumn      string `json:"target_column"`
	ConstraintName    string `json:"constraint_name"`
	OnDelete          string `json:"on_delete,omitempty"`     // CASCADE, SET NULL, etc.
	OnUpdate          string `json:"on_update,omitempty"`
}

// DocumentReference tracks references between documents
type DocumentReference struct {
	ID                string `json:"id"`
	SourceCollection  string `json:"source_collection"`
	SourceField       string `json:"source_field"`
	TargetCollection  string `json:"target_collection"`
	TargetIDField     string `json:"target_id_field"`
	ReferenceType     string `json:"reference_type"`    // "single", "array"
}

// DeploymentRecord tracks package deployments
type DeploymentRecord struct {
	ID              string                 `json:"id"`
	PackageID       string                 `json:"package_id"`
	PackageName     string                 `json:"package_name"`
	PackageVersion  string                 `json:"package_version"`
	TargetDatabase  string                 `json:"target_database"`
	DeployedAt      time.Time              `json:"deployed_at"`
	DeployedBy      string                 `json:"deployed_by"`
	Status          string                 `json:"status"`          // "pending", "in_progress", "completed", "failed", "rolled_back"
	PKMappingResult map[string]map[interface{}]interface{} `json:"pk_mapping_result"` // Table -> OldPK -> NewPK
	IDMappingResult map[string]map[interface{}]interface{} `json:"id_mapping_result"` // Collection -> OldID -> NewID
	ErrorLog        []string               `json:"error_log,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PackageFilter defines criteria for package selection
type PackageFilter struct {
	Tables           []string               `json:"tables,omitempty"`
	Collections      []string               `json:"collections,omitempty"`
	WhereClause      map[string]string      `json:"where_clause,omitempty"`      // Table/Collection -> WHERE condition
	IncludeRelated   bool                   `json:"include_related"`             // Auto-include related records
	MaxDepth         int                    `json:"max_depth"`                   // Max depth for relationship traversal
	ExcludeColumns   map[string][]string    `json:"exclude_columns,omitempty"`   // Table -> columns to exclude
	ExcludeFields    map[string][]string    `json:"exclude_fields,omitempty"`    // Collection -> fields to exclude
}

// DeploymentOptions configures deployment behavior
type DeploymentOptions struct {
	SkipExisting        bool                   `json:"skip_existing"`         // Skip records that already exist
	UpdateExisting      bool                   `json:"update_existing"`       // Update existing records
	ValidateReferences  bool                   `json:"validate_references"`   // Validate FK/references before deploy
	CreateMissing       bool                   `json:"create_missing"`        // Create missing parent records
	RebuildIndexes      bool                   `json:"rebuild_indexes"`       // Rebuild indexes after deployment
	BatchSize           int                    `json:"batch_size"`            // Number of records per batch
	TransactionSize     int                    `json:"transaction_size"`      // Records per transaction
	ContinueOnError     bool                   `json:"continue_on_error"`     // Continue deployment on errors
	DryRun              bool                   `json:"dry_run"`               // Validate but don't deploy
	Mappings            map[string]interface{} `json:"mappings,omitempty"`    // Custom field/column mappings
}
