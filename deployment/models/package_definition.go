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

// PackageDefinitionDoc represents a reusable package configuration stored in MongoDB.
// A definition captures which datasource, tables/collections, record filters, and
// global-key settings to use — so packaging can be re-run without re-configuring.
type PackageDefinitionDoc struct {
	ID          string        `bson:"_id"        json:"id"`
	Name        string        `bson:"name"       json:"name"`
	Version     string        `bson:"version"    json:"version"`
	Description string        `bson:"description" json:"description"`
	PackageType string        `bson:"package_type" json:"package_type"` // "database" | "document"
	Environment string        `bson:"environment"  json:"environment"`  // dev/staging/production

	// The full filter config (tables/collections, WHERE clauses, global keys, relations)
	Filter      PackageFilter `bson:"filter"     json:"filter"`

	// Entity-level metadata set by the user (display purposes + per-entity overrides)
	Entities    []EntityDefinition `bson:"entities" json:"entities"`

	// Deployment defaults to use when packing from this definition
	AutoVersion    bool   `bson:"auto_version"    json:"auto_version"`    // append timestamp to version
	IncludeParent  bool   `bson:"include_parent"  json:"include_parent"`

	// Version control — incremented on each save; MongoDB _id stays stable (= SQL iacpackagedefs.id)
	VersionNumber int    `bson:"version_number" json:"version_number"` // 1, 2, 3 …

	Active     bool   `bson:"active"      json:"active"`
	CreatedBy  string `bson:"createdby"   json:"createdby"`
	CreatedOn  string `bson:"createdon"   json:"createdon"`
	ModifiedBy string `bson:"modifiedby"  json:"modifiedby"`
	ModifiedOn string `bson:"modifiedon"  json:"modifiedon"`
}

// EntityDefinition holds user-configured options per table or collection
type EntityDefinition struct {
	Name            string   `bson:"name"             json:"name"`
	EntityType      string   `bson:"entity_type"      json:"entity_type"`    // "table" | "collection"
	WhereClause     string   `bson:"where_clause"     json:"where_clause"`    // SQL WHERE / MongoDB BSON JSON
	ExcludeColumns  []string `bson:"exclude_columns"  json:"exclude_columns"` // columns/fields to exclude
	GlobalKeyFields []string `bson:"global_key_fields" json:"global_key_fields"`
	Selected        bool     `bson:"selected"         json:"selected"`        // included in package
}

// TableSource describes a table available for packaging
type TableSource struct {
	Name      string      `json:"name"`
	RowCount  int         `json:"row_count"`
	Columns   []ColumnInfo `json:"columns"`
	PKColumns []string    `json:"pk_columns"`
	FKColumns []ForeignKeyInfo `json:"fk_columns,omitempty"`
}

// CollectionSource describes a MongoDB collection available for packaging
type CollectionSource struct {
	Name          string `json:"name"`
	DocumentCount int64  `json:"document_count"`
}

// SourcesResponse is returned by GET /api/packages/sources
type SourcesResponse struct {
	DatabaseType string             `json:"database_type"`
	DatabaseName string             `json:"database_name"`
	Tables       []TableSource      `json:"tables"`
	DocDatabaseName string          `json:"doc_database_name"`
	Collections  []CollectionSource `json:"collections"`
}

// TableRecordsResponse is returned by GET /api/packages/sources/:table/records
type TableRecordsResponse struct {
	Table   string                   `json:"table"`
	Total   int                      `json:"total"`
	Limit   int                      `json:"limit"`
	Offset  int                      `json:"offset"`
	Columns []string                 `json:"columns"`
	PKCols  []string                 `json:"pk_cols"`
	Rows    []map[string]interface{} `json:"rows"`
}

// CollectionDocumentsResponse is returned by GET /api/packages/sources/:collection/documents
type CollectionDocumentsResponse struct {
	Collection string                   `json:"collection"`
	Total      int                      `json:"total"`
	Limit      int                      `json:"limit"`
	Offset     int                      `json:"offset"`
	Columns    []string                 `json:"columns"`
	IDField    string                   `json:"id_field"`
	Rows       []map[string]interface{} `json:"rows"`
}
