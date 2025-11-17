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

package documents

import (
	"context"
	"time"
)

// DocumentDB defines the interface for document database operations
// This interface abstracts document database operations to support MongoDB, PostgreSQL JSONB, etc.
type DocumentDB interface {
	// Connection Management
	Connect(config *DocDBConfig) error
	Close() error
	Ping(ctx context.Context) error

	// Collection/Table Operations
	CreateCollection(ctx context.Context, name string) error
	DropCollection(ctx context.Context, name string) error
	ListCollections(ctx context.Context) ([]string, error)
	CollectionExists(ctx context.Context, name string) (bool, error)

	// Document CRUD Operations
	InsertOne(ctx context.Context, collection string, document interface{}) (string, error)
	InsertMany(ctx context.Context, collection string, documents []interface{}) ([]string, error)
	FindOne(ctx context.Context, collection string, filter interface{}) (map[string]interface{}, error)
	FindMany(ctx context.Context, collection string, filter interface{}, opts *FindOptions) ([]map[string]interface{}, error)
	UpdateOne(ctx context.Context, collection string, filter interface{}, update interface{}) error
	UpdateMany(ctx context.Context, collection string, filter interface{}, update interface{}) (int64, error)
	DeleteOne(ctx context.Context, collection string, filter interface{}) error
	DeleteMany(ctx context.Context, collection string, filter interface{}) (int64, error)
	CountDocuments(ctx context.Context, collection string, filter interface{}) (int64, error)

	// Index Operations
	CreateIndex(ctx context.Context, collection string, keys map[string]int, options *IndexOptions) error
	DropIndex(ctx context.Context, collection string, name string) error
	ListIndexes(ctx context.Context, collection string) ([]IndexInfo, error)

	// Aggregation
	Aggregate(ctx context.Context, collection string, pipeline []map[string]interface{}) ([]map[string]interface{}, error)

	// Database Information
	GetType() DocDBType
	IsConnected() bool
	Stats(ctx context.Context) (*DBStats, error)
}

// DocDBConfig represents document database connection configuration
type DocDBConfig struct {
	// Connection Details
	Type         DocDBType         `json:"type"`
	Host         string            `json:"host"`
	Port         int               `json:"port"`
	Database     string            `json:"database"`
	Username     string            `json:"username"`
	Password     string            `json:"password"`

	// Connection Options
	SSLMode      string            `json:"ssl_mode,omitempty"`
	AuthSource   string            `json:"auth_source,omitempty"`
	ReplicaSet   string            `json:"replica_set,omitempty"`

	// Pool Configuration
	MaxPoolSize  int               `json:"max_pool_size"`
	MinPoolSize  int               `json:"min_pool_size"`
	ConnTimeout  int               `json:"conn_timeout"`

	// Database-Specific Options
	Options      map[string]string `json:"options,omitempty"`
}

// DocDBType represents supported document database types
type DocDBType string

const (
	DocDBTypeMongoDB   DocDBType = "mongodb"
	DocDBTypePostgres  DocDBType = "postgres" // Using JSONB
)

// FindOptions represents options for find operations
type FindOptions struct {
	Sort       map[string]int // field -> 1 (asc) or -1 (desc)
	Limit      int64
	Skip       int64
	Projection map[string]int // field -> 1 (include) or 0 (exclude)
}

// IndexOptions represents options for index creation
type IndexOptions struct {
	Name       string
	Unique     bool
	Background bool
	Sparse     bool
}

// IndexInfo represents index metadata
type IndexInfo struct {
	Name    string
	Keys    map[string]int
	Unique  bool
	Sparse  bool
}

// DBStats represents database statistics
type DBStats struct {
	Collections int64
	Documents   int64
	DataSize    int64
	IndexSize   int64
	StorageSize int64
}

// Validate validates the document database configuration
func (c *DocDBConfig) Validate() error {
	if c.Type == "" {
		return ErrInvalidDocDBType
	}
	if c.Host == "" {
		return ErrMissingHost
	}
	if c.Database == "" {
		return ErrMissingDatabase
	}

	// Set defaults
	if c.MaxPoolSize == 0 {
		c.MaxPoolSize = 100
	}
	if c.MinPoolSize == 0 {
		c.MinPoolSize = 10
	}
	if c.ConnTimeout == 0 {
		c.ConnTimeout = 30
	}
	if c.Port == 0 {
		switch c.Type {
		case DocDBTypeMongoDB:
			c.Port = 27017
		case DocDBTypePostgres:
			c.Port = 5432
		}
	}

	return nil
}

// BuildConnectionString builds database-specific connection string
func (c *DocDBConfig) BuildConnectionString() string {
	// This will be implemented by database-specific builders
	return ""
}

// Document-specific errors
var (
	ErrInvalidDocDBType    = NewDocError("invalid document database type")
	ErrMissingHost         = NewDocError("database host is required")
	ErrMissingDatabase     = NewDocError("database name is required")
	ErrCollectionNotFound  = NewDocError("collection not found")
	ErrDocumentNotFound    = NewDocError("document not found")
	ErrInvalidFilter       = NewDocError("invalid filter")
	ErrInvalidUpdate       = NewDocError("invalid update operation")
	ErrConnectionFailed    = NewDocError("failed to connect to document database")
)

// DocError represents a document database error
type DocError struct {
	Message string
}

func (e *DocError) Error() string {
	return e.Message
}

// NewDocError creates a new document database error
func NewDocError(message string) *DocError {
	return &DocError{Message: message}
}

// QueryBuilder helps build database-agnostic queries
type QueryBuilder interface {
	// Filter Operations
	Equals(field string, value interface{}) QueryBuilder
	NotEquals(field string, value interface{}) QueryBuilder
	GreaterThan(field string, value interface{}) QueryBuilder
	LessThan(field string, value interface{}) QueryBuilder
	In(field string, values []interface{}) QueryBuilder
	Contains(field string, value string) QueryBuilder

	// Logical Operations
	And(queries ...QueryBuilder) QueryBuilder
	Or(queries ...QueryBuilder) QueryBuilder
	Not(query QueryBuilder) QueryBuilder

	// Build final filter
	Build() interface{}
}

// UpdateBuilder helps build database-agnostic update operations
type UpdateBuilder interface {
	// Update Operations
	Set(field string, value interface{}) UpdateBuilder
	Unset(field string) UpdateBuilder
	Increment(field string, value interface{}) UpdateBuilder
	Push(field string, value interface{}) UpdateBuilder
	Pull(field string, value interface{}) UpdateBuilder

	// Build final update
	Build() interface{}
}

// AggregationPipeline represents an aggregation pipeline
type AggregationPipeline struct {
	Stages []map[string]interface{}
}

// NewAggregationPipeline creates a new aggregation pipeline
func NewAggregationPipeline() *AggregationPipeline {
	return &AggregationPipeline{
		Stages: make([]map[string]interface{}, 0),
	}
}

// Match adds a $match stage
func (p *AggregationPipeline) Match(filter interface{}) *AggregationPipeline {
	p.Stages = append(p.Stages, map[string]interface{}{"$match": filter})
	return p
}

// Group adds a $group stage
func (p *AggregationPipeline) Group(groupBy interface{}, fields map[string]interface{}) *AggregationPipeline {
	stage := map[string]interface{}{
		"_id": groupBy,
	}
	for k, v := range fields {
		stage[k] = v
	}
	p.Stages = append(p.Stages, map[string]interface{}{"$group": stage})
	return p
}

// Sort adds a $sort stage
func (p *AggregationPipeline) Sort(sort map[string]int) *AggregationPipeline {
	p.Stages = append(p.Stages, map[string]interface{}{"$sort": sort})
	return p
}

// Limit adds a $limit stage
func (p *AggregationPipeline) Limit(limit int64) *AggregationPipeline {
	p.Stages = append(p.Stages, map[string]interface{}{"$limit": limit})
	return p
}

// Skip adds a $skip stage
func (p *AggregationPipeline) Skip(skip int64) *AggregationPipeline {
	p.Stages = append(p.Stages, map[string]interface{}{"$skip": skip})
	return p
}

// Project adds a $project stage
func (p *AggregationPipeline) Project(projection map[string]interface{}) *AggregationPipeline {
	p.Stages = append(p.Stages, map[string]interface{}{"$project": projection})
	return p
}

// Build returns the pipeline stages
func (p *AggregationPipeline) Build() []map[string]interface{} {
	return p.Stages
}

// ConnectionInfo represents document database connection information
type ConnectionInfo struct {
	Type        DocDBType
	Host        string
	Port        int
	Database    string
	IsConnected bool
	ConnectedAt time.Time
}
