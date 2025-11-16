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
	"sync"
	"sync/atomic"
	"time"

	"github.com/mdaxf/iac/logger"
)

// MigrationConfig represents configuration for document migration
type MigrationConfig struct {
	BatchSize       int
	Workers         int
	ContinueOnError bool
	Transform       TransformFunc
	Filter          interface{}
	DryRun          bool
	DropTarget      bool // Drop target collection before migration
}

// TransformFunc is a function that transforms a document during migration
type TransformFunc func(doc map[string]interface{}) (map[string]interface{}, error)

// MigrationProgress represents migration progress
type MigrationProgress struct {
	TotalDocuments     int64
	ProcessedDocuments int64
	SuccessDocuments   int64
	FailedDocuments    int64
	SkippedDocuments   int64
	StartTime          time.Time
	EndTime            time.Time
	Duration           time.Duration
	Errors             []MigrationError
	mu                 sync.RWMutex
}

// MigrationError represents an error during migration
type MigrationError struct {
	DocumentID string
	Error      error
	Timestamp  time.Time
}

// Add adds an error to the progress
func (p *MigrationProgress) AddError(docID string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Errors = append(p.Errors, MigrationError{
		DocumentID: docID,
		Error:      err,
		Timestamp:  time.Now(),
	})
}

// GetProgress returns current progress statistics
func (p *MigrationProgress) GetProgress() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	processed := atomic.LoadInt64(&p.ProcessedDocuments)
	total := atomic.LoadInt64(&p.TotalDocuments)

	percentage := float64(0)
	if total > 0 {
		percentage = float64(processed) / float64(total) * 100
	}

	return map[string]interface{}{
		"total":          total,
		"processed":      processed,
		"success":        atomic.LoadInt64(&p.SuccessDocuments),
		"failed":         atomic.LoadInt64(&p.FailedDocuments),
		"skipped":        atomic.LoadInt64(&p.SkippedDocuments),
		"percentage":     percentage,
		"elapsed":        time.Since(p.StartTime).String(),
		"error_count":    len(p.Errors),
	}
}

// DocumentMigrator migrates documents between databases
type DocumentMigrator struct {
	source   DocumentDB
	target   DocumentDB
	iLog     logger.Log
}

// NewDocumentMigrator creates a new document migrator
func NewDocumentMigrator(source, target DocumentDB) *DocumentMigrator {
	return &DocumentMigrator{
		source: source,
		target: target,
		iLog:   logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DocumentMigrator"},
	}
}

// Migrate migrates documents from source to target
func (m *DocumentMigrator) Migrate(ctx context.Context, sourceCollection, targetCollection string, config *MigrationConfig) (*MigrationProgress, error) {
	if config == nil {
		config = &MigrationConfig{
			BatchSize:       100,
			Workers:         5,
			ContinueOnError: true,
		}
	}

	progress := &MigrationProgress{
		StartTime: time.Now(),
		Errors:    make([]MigrationError, 0),
	}

	m.iLog.Info(fmt.Sprintf("Starting migration from %s to %s (source: %s, target: %s)",
		sourceCollection, targetCollection, m.source.GetType(), m.target.GetType()))

	// Drop target collection if requested
	if config.DropTarget && !config.DryRun {
		m.iLog.Info(fmt.Sprintf("Dropping target collection: %s", targetCollection))
		if err := m.target.DropCollection(ctx, targetCollection); err != nil {
			m.iLog.Warn(fmt.Sprintf("Failed to drop target collection (may not exist): %v", err))
		}
	}

	// Ensure target collection exists
	if !config.DryRun {
		if err := m.target.CreateCollection(ctx, targetCollection); err != nil {
			m.iLog.Warn(fmt.Sprintf("Failed to create target collection (may already exist): %v", err))
		}
	}

	// Get total document count
	filter := config.Filter
	if filter == nil {
		filter = map[string]interface{}{}
	}

	totalCount, err := m.source.CountDocuments(ctx, sourceCollection, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count source documents: %w", err)
	}

	atomic.StoreInt64(&progress.TotalDocuments, totalCount)
	m.iLog.Info(fmt.Sprintf("Total documents to migrate: %d", totalCount))

	// Migrate in batches
	offset := int64(0)
	batchSize := int64(config.BatchSize)

	for offset < totalCount {
		select {
		case <-ctx.Done():
			return progress, ctx.Err()
		default:
		}

		// Fetch batch
		findOpts := &FindOptions{
			Skip:  offset,
			Limit: batchSize,
		}

		documents, err := m.source.FindMany(ctx, sourceCollection, filter, findOpts)
		if err != nil {
			if config.ContinueOnError {
				m.iLog.Error(fmt.Sprintf("Failed to fetch batch at offset %d: %v", offset, err))
				progress.AddError(fmt.Sprintf("batch_%d", offset), err)
				offset += batchSize
				continue
			}
			return progress, fmt.Errorf("failed to fetch batch: %w", err)
		}

		// Process batch
		if err := m.processBatch(ctx, targetCollection, documents, config, progress); err != nil {
			if !config.ContinueOnError {
				return progress, err
			}
		}

		offset += int64(len(documents))

		// Log progress
		if offset%1000 == 0 || offset >= totalCount {
			m.iLog.Info(fmt.Sprintf("Migration progress: %d/%d (%.2f%%)",
				offset, totalCount, float64(offset)/float64(totalCount)*100))
		}
	}

	progress.EndTime = time.Now()
	progress.Duration = progress.EndTime.Sub(progress.StartTime)

	m.iLog.Info(fmt.Sprintf("Migration completed: %d total, %d success, %d failed, %d skipped in %v",
		totalCount, progress.SuccessDocuments, progress.FailedDocuments, progress.SkippedDocuments, progress.Duration))

	return progress, nil
}

// processBatch processes a batch of documents
func (m *DocumentMigrator) processBatch(ctx context.Context, targetCollection string, documents []map[string]interface{}, config *MigrationConfig, progress *MigrationProgress) error {
	for _, doc := range documents {
		atomic.AddInt64(&progress.ProcessedDocuments, 1)

		// Transform document if transform function is provided
		transformedDoc := doc
		if config.Transform != nil {
			var err error
			transformedDoc, err = config.Transform(doc)
			if err != nil {
				atomic.AddInt64(&progress.FailedDocuments, 1)
				docID := m.getDocumentID(doc)
				progress.AddError(docID, fmt.Errorf("transform failed: %w", err))
				if !config.ContinueOnError {
					return err
				}
				continue
			}
		}

		// Skip if transform returned nil (document filtered out)
		if transformedDoc == nil {
			atomic.AddInt64(&progress.SkippedDocuments, 1)
			continue
		}

		// Insert into target (skip in dry run mode)
		if !config.DryRun {
			_, err := m.target.InsertOne(ctx, targetCollection, transformedDoc)
			if err != nil {
				atomic.AddInt64(&progress.FailedDocuments, 1)
				docID := m.getDocumentID(doc)
				progress.AddError(docID, fmt.Errorf("insert failed: %w", err))
				if !config.ContinueOnError {
					return err
				}
				continue
			}
		}

		atomic.AddInt64(&progress.SuccessDocuments, 1)
	}

	return nil
}

// getDocumentID extracts document ID for error reporting
func (m *DocumentMigrator) getDocumentID(doc map[string]interface{}) string {
	if id, ok := doc["_id"]; ok {
		return fmt.Sprintf("%v", id)
	}
	if uuid, ok := doc["_uuid"]; ok {
		return fmt.Sprintf("%v", uuid)
	}
	return "unknown"
}

// MigrateCollection is a convenience method to migrate an entire collection
func (m *DocumentMigrator) MigrateCollection(ctx context.Context, collection string, config *MigrationConfig) (*MigrationProgress, error) {
	return m.Migrate(ctx, collection, collection, config)
}

// MigrateWithTransform migrates documents with a transformation function
func (m *DocumentMigrator) MigrateWithTransform(ctx context.Context, sourceCollection, targetCollection string, transform TransformFunc, config *MigrationConfig) (*MigrationProgress, error) {
	if config == nil {
		config = &MigrationConfig{}
	}
	config.Transform = transform

	return m.Migrate(ctx, sourceCollection, targetCollection, config)
}

// SyncDirection represents the direction of synchronization
type SyncDirection string

const (
	SyncSourceToTarget SyncDirection = "source_to_target"
	SyncTargetToSource SyncDirection = "target_to_source"
	SyncBidirectional  SyncDirection = "bidirectional"
)

// DocumentSynchronizer synchronizes documents between two databases
type DocumentSynchronizer struct {
	migrator *DocumentMigrator
	iLog     logger.Log
}

// NewDocumentSynchronizer creates a new document synchronizer
func NewDocumentSynchronizer(source, target DocumentDB) *DocumentSynchronizer {
	return &DocumentSynchronizer{
		migrator: NewDocumentMigrator(source, target),
		iLog:     logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DocumentSynchronizer"},
	}
}

// Sync synchronizes documents between databases
func (s *DocumentSynchronizer) Sync(ctx context.Context, collection string, direction SyncDirection) error {
	s.iLog.Info(fmt.Sprintf("Starting synchronization for collection: %s (direction: %s)", collection, direction))

	switch direction {
	case SyncSourceToTarget:
		_, err := s.migrator.MigrateCollection(ctx, collection, &MigrationConfig{
			BatchSize:       100,
			ContinueOnError: true,
		})
		return err

	case SyncTargetToSource:
		// Reverse source and target
		reversed := NewDocumentMigrator(s.migrator.target, s.migrator.source)
		_, err := reversed.MigrateCollection(ctx, collection, &MigrationConfig{
			BatchSize:       100,
			ContinueOnError: true,
		})
		return err

	case SyncBidirectional:
		// This is complex - would need conflict resolution
		return fmt.Errorf("bidirectional sync not yet implemented")

	default:
		return fmt.Errorf("unknown sync direction: %s", direction)
	}
}

// Common transformation functions

// RemoveFieldsTransform removes specified fields from documents
func RemoveFieldsTransform(fields ...string) TransformFunc {
	return func(doc map[string]interface{}) (map[string]interface{}, error) {
		for _, field := range fields {
			delete(doc, field)
		}
		return doc, nil
	}
}

// RenameFieldsTransform renames fields in documents
func RenameFieldsTransform(mapping map[string]string) TransformFunc {
	return func(doc map[string]interface{}) (map[string]interface{}, error) {
		for oldName, newName := range mapping {
			if value, exists := doc[oldName]; exists {
				doc[newName] = value
				delete(doc, oldName)
			}
		}
		return doc, nil
	}
}

// FilterTransform filters documents based on a condition
func FilterTransform(condition func(map[string]interface{}) bool) TransformFunc {
	return func(doc map[string]interface{}) (map[string]interface{}, error) {
		if condition(doc) {
			return doc, nil
		}
		return nil, nil // Return nil to skip document
	}
}

// ChainTransforms chains multiple transform functions
func ChainTransforms(transforms ...TransformFunc) TransformFunc {
	return func(doc map[string]interface{}) (map[string]interface{}, error) {
		result := doc
		for _, transform := range transforms {
			var err error
			result, err = transform(result)
			if err != nil {
				return nil, err
			}
			if result == nil {
				return nil, nil // Document filtered out
			}
		}
		return result, nil
	}
}

// AddFieldTransform adds a field to documents
func AddFieldTransform(field string, value interface{}) TransformFunc {
	return func(doc map[string]interface{}) (map[string]interface{}, error) {
		doc[field] = value
		return doc, nil
	}
}

// TimestampTransform adds migration timestamp
func TimestampTransform() TransformFunc {
	return func(doc map[string]interface{}) (map[string]interface{}, error) {
		doc["migrated_at"] = time.Now()
		return doc, nil
	}
}

// Example usage
func ExampleMigrationUsage() {
	// Assuming we have source and target databases
	// sourceDB := documents.GetDocDBInstance("mongodb")
	// targetDB := documents.GetDocDBInstance("postgres-jsonb")

	// migrator := NewDocumentMigrator(sourceDB, targetDB)

	// // Simple migration
	// ctx := context.Background()
	// progress, err := migrator.MigrateCollection(ctx, "users", &MigrationConfig{
	// 	BatchSize:       100,
	// 	Workers:         5,
	// 	ContinueOnError: true,
	// })

	// if err != nil {
	// 	fmt.Printf("Migration failed: %v\n", err)
	// } else {
	// 	fmt.Printf("Migration completed: %+v\n", progress.GetProgress())
	// }

	// // Migration with transformation
	// transform := ChainTransforms(
	// 	RemoveFieldsTransform("_id"), // Remove MongoDB ObjectID
	// 	RenameFieldsTransform(map[string]string{
	// 		"old_field": "new_field",
	// 	}),
	// 	TimestampTransform(),
	// 	FilterTransform(func(doc map[string]interface{}) bool {
	// 		status, _ := doc["status"].(string)
	// 		return status == "active"
	// 	}),
	// )

	// progress, err = migrator.MigrateWithTransform(ctx, "users", "active_users", transform, &MigrationConfig{
	// 	BatchSize:       100,
	// 	ContinueOnError: true,
	// 	DropTarget:      true,
	// })

	// // Dry run to test migration
	// progress, err = migrator.MigrateCollection(ctx, "users", &MigrationConfig{
	// 	BatchSize: 100,
	// 	DryRun:    true,
	// })
	// fmt.Printf("Dry run results: %+v\n", progress.GetProgress())
}

// MigrationPlan represents a multi-step migration plan
type MigrationPlan struct {
	Steps []MigrationStep
	iLog  logger.Log
}

// MigrationStep represents a single migration step
type MigrationStep struct {
	Name               string
	SourceCollection   string
	TargetCollection   string
	Transform          TransformFunc
	Config             *MigrationConfig
}

// NewMigrationPlan creates a new migration plan
func NewMigrationPlan() *MigrationPlan {
	return &MigrationPlan{
		Steps: make([]MigrationStep, 0),
		iLog:  logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MigrationPlan"},
	}
}

// AddStep adds a migration step
func (p *MigrationPlan) AddStep(step MigrationStep) {
	p.Steps = append(p.Steps, step)
}

// Execute executes the migration plan
func (p *MigrationPlan) Execute(ctx context.Context, migrator *DocumentMigrator) map[string]*MigrationProgress {
	results := make(map[string]*MigrationProgress)

	for i, step := range p.Steps {
		p.iLog.Info(fmt.Sprintf("Executing migration step %d/%d: %s", i+1, len(p.Steps), step.Name))

		progress, err := migrator.MigrateWithTransform(ctx, step.SourceCollection, step.TargetCollection, step.Transform, step.Config)
		if err != nil {
			p.iLog.Error(fmt.Sprintf("Migration step '%s' failed: %v", step.Name, err))
		}

		results[step.Name] = progress
	}

	return results
}
