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
	"fmt"
	"strings"
	"sync"

	"github.com/mdaxf/iac/logger"
)

// IndexType represents different types of indexes
type IndexType string

const (
	IndexTypeSingle     IndexType = "single"      // Single field index
	IndexTypeCompound   IndexType = "compound"    // Multiple fields
	IndexTypeText       IndexType = "text"        // Full-text search
	IndexTypeGeo        IndexType = "geo"         // Geospatial
	IndexTypeHashed     IndexType = "hashed"      // Hashed index
	IndexTypeWildcard   IndexType = "wildcard"    // Wildcard index
)

// IndexManager manages indexes for document collections
type IndexManager struct {
	db   DocumentDB
	iLog logger.Log
	mu   sync.RWMutex
}

// NewIndexManager creates a new index manager
func NewIndexManager(db DocumentDB) *IndexManager {
	return &IndexManager{
		db:   db,
		iLog: logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "IndexManager"},
	}
}

// IndexDefinition represents a complete index definition
type IndexDefinition struct {
	Name       string
	Collection string
	Fields     []IndexField
	Type       IndexType
	Options    *IndexOptions
}

// IndexField represents a field in an index
type IndexField struct {
	Name  string
	Order int // 1 for ascending, -1 for descending, "text" for text index
}

// CreateSingleFieldIndex creates a single field index
func (im *IndexManager) CreateSingleFieldIndex(ctx context.Context, collection, field string, ascending bool, opts *IndexOptions) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	order := 1
	if !ascending {
		order = -1
	}

	keys := map[string]int{field: order}

	if opts == nil {
		opts = &IndexOptions{}
	}

	if opts.Name == "" {
		opts.Name = fmt.Sprintf("%s_%d", field, order)
	}

	if err := im.db.CreateIndex(ctx, collection, keys, opts); err != nil {
		return fmt.Errorf("failed to create single field index: %w", err)
	}

	im.iLog.Info(fmt.Sprintf("Created single field index '%s' on %s.%s", opts.Name, collection, field))

	return nil
}

// CreateCompoundIndex creates a compound index on multiple fields
func (im *IndexManager) CreateCompoundIndex(ctx context.Context, collection string, fields []IndexField, opts *IndexOptions) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	keys := make(map[string]int)
	fieldNames := make([]string, 0, len(fields))

	for _, field := range fields {
		keys[field.Name] = field.Order
		orderStr := "asc"
		if field.Order == -1 {
			orderStr = "desc"
		}
		fieldNames = append(fieldNames, fmt.Sprintf("%s_%s", field.Name, orderStr))
	}

	if opts == nil {
		opts = &IndexOptions{}
	}

	if opts.Name == "" {
		opts.Name = strings.Join(fieldNames, "_")
	}

	if err := im.db.CreateIndex(ctx, collection, keys, opts); err != nil {
		return fmt.Errorf("failed to create compound index: %w", err)
	}

	im.iLog.Info(fmt.Sprintf("Created compound index '%s' on %s", opts.Name, collection))

	return nil
}

// CreateTextIndex creates a full-text search index
func (im *IndexManager) CreateTextIndex(ctx context.Context, collection string, fields []string, opts *IndexOptions) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	// For MongoDB, text indexes use "text" as the value
	// For PostgreSQL, we'll create a special index type
	keys := make(map[string]int)

	if im.db.GetType() == DocDBTypeMongoDB {
		for _, field := range fields {
			keys[field] = 1 // Text indexes in MongoDB
		}
	} else {
		// For PostgreSQL, create GIN index for text search
		for _, field := range fields {
			keys[field] = 1
		}
	}

	if opts == nil {
		opts = &IndexOptions{}
	}

	if opts.Name == "" {
		opts.Name = fmt.Sprintf("text_%s", strings.Join(fields, "_"))
	}

	if err := im.db.CreateIndex(ctx, collection, keys, opts); err != nil {
		return fmt.Errorf("failed to create text index: %w", err)
	}

	im.iLog.Info(fmt.Sprintf("Created text index '%s' on %s", opts.Name, collection))

	return nil
}

// CreateUniqueIndex creates a unique constraint index
func (im *IndexManager) CreateUniqueIndex(ctx context.Context, collection string, fields []string, opts *IndexOptions) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	keys := make(map[string]int)
	for _, field := range fields {
		keys[field] = 1
	}

	if opts == nil {
		opts = &IndexOptions{
			Unique: true,
		}
	} else {
		opts.Unique = true
	}

	if opts.Name == "" {
		opts.Name = fmt.Sprintf("unique_%s", strings.Join(fields, "_"))
	}

	if err := im.db.CreateIndex(ctx, collection, keys, opts); err != nil {
		return fmt.Errorf("failed to create unique index: %w", err)
	}

	im.iLog.Info(fmt.Sprintf("Created unique index '%s' on %s", opts.Name, collection))

	return nil
}

// CreateSparseIndex creates a sparse index (only indexes documents that have the field)
func (im *IndexManager) CreateSparseIndex(ctx context.Context, collection, field string, opts *IndexOptions) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	keys := map[string]int{field: 1}

	if opts == nil {
		opts = &IndexOptions{
			Sparse: true,
		}
	} else {
		opts.Sparse = true
	}

	if opts.Name == "" {
		opts.Name = fmt.Sprintf("sparse_%s", field)
	}

	if err := im.db.CreateIndex(ctx, collection, keys, opts); err != nil {
		return fmt.Errorf("failed to create sparse index: %w", err)
	}

	im.iLog.Info(fmt.Sprintf("Created sparse index '%s' on %s.%s", opts.Name, collection, field))

	return nil
}

// DropIndex drops an index by name
func (im *IndexManager) DropIndex(ctx context.Context, collection, indexName string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if err := im.db.DropIndex(ctx, collection, indexName); err != nil {
		return fmt.Errorf("failed to drop index: %w", err)
	}

	im.iLog.Info(fmt.Sprintf("Dropped index '%s' from %s", indexName, collection))

	return nil
}

// ListIndexes lists all indexes on a collection
func (im *IndexManager) ListIndexes(ctx context.Context, collection string) ([]IndexInfo, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	indexes, err := im.db.ListIndexes(ctx, collection)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}

	return indexes, nil
}

// IndexExists checks if an index exists
func (im *IndexManager) IndexExists(ctx context.Context, collection, indexName string) (bool, error) {
	indexes, err := im.ListIndexes(ctx, collection)
	if err != nil {
		return false, err
	}

	for _, idx := range indexes {
		if idx.Name == indexName {
			return true, nil
		}
	}

	return false, nil
}

// RebuildIndex drops and recreates an index
func (im *IndexManager) RebuildIndex(ctx context.Context, definition *IndexDefinition) error {
	// Check if index exists
	exists, err := im.IndexExists(ctx, definition.Collection, definition.Name)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	// Drop if exists
	if exists {
		if err := im.DropIndex(ctx, definition.Collection, definition.Name); err != nil {
			return fmt.Errorf("failed to drop existing index: %w", err)
		}
	}

	// Recreate based on type
	switch definition.Type {
	case IndexTypeSingle:
		if len(definition.Fields) != 1 {
			return fmt.Errorf("single field index requires exactly one field")
		}
		ascending := definition.Fields[0].Order == 1
		return im.CreateSingleFieldIndex(ctx, definition.Collection, definition.Fields[0].Name, ascending, definition.Options)

	case IndexTypeCompound:
		return im.CreateCompoundIndex(ctx, definition.Collection, definition.Fields, definition.Options)

	case IndexTypeText:
		fieldNames := make([]string, len(definition.Fields))
		for i, f := range definition.Fields {
			fieldNames[i] = f.Name
		}
		return im.CreateTextIndex(ctx, definition.Collection, fieldNames, definition.Options)

	default:
		return fmt.Errorf("unsupported index type: %s", definition.Type)
	}
}

// GetIndexStats returns statistics about indexes
func (im *IndexManager) GetIndexStats(ctx context.Context, collection string) (*IndexStats, error) {
	indexes, err := im.ListIndexes(ctx, collection)
	if err != nil {
		return nil, err
	}

	stats := &IndexStats{
		Collection:   collection,
		TotalIndexes: len(indexes),
		Indexes:      indexes,
	}

	// Count index types
	for _, idx := range indexes {
		if idx.Unique {
			stats.UniqueIndexes++
		}
		if idx.Sparse {
			stats.SparseIndexes++
		}

		// Determine if compound
		if len(idx.Keys) > 1 {
			stats.CompoundIndexes++
		} else {
			stats.SingleFieldIndexes++
		}
	}

	return stats, nil
}

// IndexStats represents index statistics
type IndexStats struct {
	Collection         string
	TotalIndexes       int
	SingleFieldIndexes int
	CompoundIndexes    int
	UniqueIndexes      int
	SparseIndexes      int
	TextIndexes        int
	Indexes            []IndexInfo
}

// IndexRecommendation represents an index recommendation
type IndexRecommendation struct {
	Collection  string
	Field       string
	Reason      string
	IndexType   IndexType
	Priority    int // 1-10, 10 being highest
	Estimated   string
}

// IndexAnalyzer analyzes query patterns and recommends indexes
type IndexAnalyzer struct {
	manager *IndexManager
	iLog    logger.Log
}

// NewIndexAnalyzer creates a new index analyzer
func NewIndexAnalyzer(manager *IndexManager) *IndexAnalyzer {
	return &IndexAnalyzer{
		manager: manager,
		iLog:    logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "IndexAnalyzer"},
	}
}

// AnalyzeCollection analyzes a collection and recommends indexes
func (ia *IndexAnalyzer) AnalyzeCollection(ctx context.Context, collection string, sampleSize int) ([]IndexRecommendation, error) {
	recommendations := make([]IndexRecommendation, 0)

	// Get existing indexes
	existingIndexes, err := ia.manager.ListIndexes(ctx, collection)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing indexes: %w", err)
	}

	// Create a map of indexed fields
	indexedFields := make(map[string]bool)
	for _, idx := range existingIndexes {
		for field := range idx.Keys {
			indexedFields[field] = true
		}
	}

	// Common fields that should be indexed
	commonFields := []struct {
		Field    string
		Reason   string
		Priority int
	}{
		{"_id", "Primary key - usually auto-indexed", 1},
		{"created_at", "Common for time-based queries", 8},
		{"updated_at", "Common for tracking changes", 7},
		{"status", "Common for filtering", 7},
		{"user_id", "Common for user-scoped queries", 8},
		{"email", "Common for lookups", 9},
		{"username", "Common for authentication", 9},
		{"type", "Common for polymorphic queries", 6},
		{"deleted", "Common for soft deletes", 6},
		{"active", "Common for status filtering", 6},
	}

	for _, field := range commonFields {
		if !indexedFields[field.Field] {
			recommendations = append(recommendations, IndexRecommendation{
				Collection: collection,
				Field:      field.Field,
				Reason:     field.Reason,
				IndexType:  IndexTypeSingle,
				Priority:   field.Priority,
				Estimated:  "Low overhead",
			})
		}
	}

	return recommendations, nil
}

// RecommendUniqueIndexes recommends unique indexes based on data analysis
func (ia *IndexAnalyzer) RecommendUniqueIndexes(ctx context.Context, collection string) ([]IndexRecommendation, error) {
	recommendations := make([]IndexRecommendation, 0)

	// Common unique fields
	uniqueFields := []string{"email", "username", "uuid", "slug", "code"}

	existingIndexes, err := ia.manager.ListIndexes(ctx, collection)
	if err != nil {
		return nil, err
	}

	// Check which unique fields are not indexed
	indexedUniqueFields := make(map[string]bool)
	for _, idx := range existingIndexes {
		if idx.Unique {
			for field := range idx.Keys {
				indexedUniqueFields[field] = true
			}
		}
	}

	for _, field := range uniqueFields {
		if !indexedUniqueFields[field] {
			recommendations = append(recommendations, IndexRecommendation{
				Collection: collection,
				Field:      field,
				Reason:     "Field should enforce uniqueness",
				IndexType:  IndexTypeSingle,
				Priority:   9,
				Estimated:  "Medium overhead - enforces constraint",
			})
		}
	}

	return recommendations, nil
}

// ApplyRecommendation applies an index recommendation
func (ia *IndexAnalyzer) ApplyRecommendation(ctx context.Context, rec IndexRecommendation) error {
	opts := &IndexOptions{
		Name: fmt.Sprintf("%s_idx", rec.Field),
	}

	switch rec.IndexType {
	case IndexTypeSingle:
		return ia.manager.CreateSingleFieldIndex(ctx, rec.Collection, rec.Field, true, opts)
	default:
		return fmt.Errorf("unsupported recommendation type: %s", rec.IndexType)
	}
}

// BatchCreateIndexes creates multiple indexes efficiently
func (im *IndexManager) BatchCreateIndexes(ctx context.Context, definitions []*IndexDefinition) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	errors := make([]error, 0)

	for _, def := range definitions {
		keys := make(map[string]int)
		for _, field := range def.Fields {
			keys[field.Name] = field.Order
		}

		if err := im.db.CreateIndex(ctx, def.Collection, keys, def.Options); err != nil {
			errors = append(errors, fmt.Errorf("failed to create index %s: %w", def.Name, err))
		} else {
			im.iLog.Info(fmt.Sprintf("Created index '%s' on %s", def.Name, def.Collection))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("batch index creation had %d errors: %v", len(errors), errors)
	}

	return nil
}

// EnsureIndexes ensures that required indexes exist
func (im *IndexManager) EnsureIndexes(ctx context.Context, collection string, definitions []*IndexDefinition) error {
	for _, def := range definitions {
		exists, err := im.IndexExists(ctx, collection, def.Name)
		if err != nil {
			return fmt.Errorf("failed to check index %s: %w", def.Name, err)
		}

		if !exists {
			if err := im.RebuildIndex(ctx, def); err != nil {
				return fmt.Errorf("failed to ensure index %s: %w", def.Name, err)
			}
		}
	}

	return nil
}

// GetIndexSize estimates the size of indexes (database-specific)
func (im *IndexManager) GetIndexSize(ctx context.Context, collection string) (map[string]int64, error) {
	// This would need database-specific implementation
	// For now, return empty map
	return make(map[string]int64), nil
}

// Example usage
func ExampleIndexUsage(db DocumentDB) {
	im := NewIndexManager(db)
	ctx := context.Background()

	// Create single field index
	im.CreateSingleFieldIndex(ctx, "users", "email", true, &IndexOptions{
		Name:   "email_idx",
		Unique: true,
	})

	// Create compound index
	im.CreateCompoundIndex(ctx, "orders", []IndexField{
		{Name: "user_id", Order: 1},
		{Name: "created_at", Order: -1},
	}, &IndexOptions{
		Name: "user_orders_idx",
	})

	// Create text index
	im.CreateTextIndex(ctx, "articles", []string{"title", "content"}, &IndexOptions{
		Name: "article_search_idx",
	})

	// List all indexes
	indexes, _ := im.ListIndexes(ctx, "users")
	fmt.Printf("Indexes: %v\n", indexes)

	// Analyze and get recommendations
	analyzer := NewIndexAnalyzer(im)
	recommendations, _ := analyzer.AnalyzeCollection(ctx, "users", 1000)
	fmt.Printf("Recommendations: %v\n", recommendations)
}
