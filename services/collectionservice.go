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

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/mdaxf/iac/dbinitializer"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"go.mongodb.org/mongo-driver/bson"
)

// CollectionService provides collection operations that work across multiple document DB types
type CollectionService struct {
	iLog logger.Log
}

// NewCollectionService creates a new collection service instance
func NewCollectionService() *CollectionService {
	return &CollectionService{
		iLog: logger.Log{ModuleName: logger.API, User: "System", ControllerName: "CollectionService"},
	}
}

// QueryOptions represents options for querying collections
type QueryOptions struct {
	Filter     map[string]interface{} // Root element filters
	Projection map[string]interface{} // Fields to include/exclude
	PageSize   int                    // Number of items per page (limit)
	Page       int                    // Page number (starts at 1)
	Sort       map[string]int         // Sort fields: 1 for asc, -1 for desc
}

// QueryResult represents the result of a collection query with pagination
type QueryResult struct {
	Data       []map[string]interface{} `json:"data"`
	TotalCount int64                    `json:"total_count"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
}

// getDocumentDB retrieves the document database instance
func (s *CollectionService) getDocumentDB() (documents.DocumentDB, error) {
	if dbinitializer.GlobalInitializer == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return dbinitializer.GlobalInitializer.GetDocumentDB()
}

// QueryCollection queries a collection with pagination and filtering
func (s *CollectionService) QueryCollection(collectionName string, opts *QueryOptions) (*QueryResult, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		s.iLog.PerformanceWithDuration("CollectionService.QueryCollection", elapsed)
	}()

	if opts == nil {
		opts = &QueryOptions{}
	}

	// Set default page size if not specified
	if opts.PageSize <= 0 {
		opts.PageSize = 100 // Default page size
	}

	// Set default page if not specified
	if opts.Page <= 0 {
		opts.Page = 1
	}

	// Get document database
	docDB, err := s.getDocumentDB()
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get document database: %v", err))
		return nil, err
	}

	ctx := context.Background()

	// Count total documents matching the filter
	totalCount, err := docDB.CountDocuments(ctx, collectionName, opts.Filter)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to count documents: %v", err))
		return nil, err
	}

	// Calculate offset based on page number
	offset := int64((opts.Page - 1) * opts.PageSize)

	// Build find options
	findOpts := &documents.FindOptions{
		Limit:      int64(opts.PageSize),
		Skip:       offset,
		Projection: convertToIntMap(opts.Projection),
		Sort:       opts.Sort,
	}

	// Query documents
	results, err := docDB.FindMany(ctx, collectionName, opts.Filter, findOpts)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to query collection: %v", err))
		return nil, err
	}

	// Calculate total pages
	totalPages := int(totalCount) / opts.PageSize
	if int(totalCount)%opts.PageSize > 0 {
		totalPages++
	}

	return &QueryResult{
		Data:       results,
		TotalCount: totalCount,
		Page:       opts.Page,
		PageSize:   opts.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetItemByID retrieves a single item by its ID
func (s *CollectionService) GetItemByID(collectionName string, id string) (map[string]interface{}, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		s.iLog.PerformanceWithDuration("CollectionService.GetItemByID", elapsed)
	}()

	docDB, err := s.getDocumentDB()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	filter := map[string]interface{}{"_id": id}

	result, err := docDB.FindOne(ctx, collectionName, filter)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get item by ID: %v", err))
		return nil, err
	}

	return result, nil
}

// GetItemByField retrieves a single item by a specific field
func (s *CollectionService) GetItemByField(collectionName string, field string, value interface{}) (map[string]interface{}, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		s.iLog.PerformanceWithDuration("CollectionService.GetItemByField", elapsed)
	}()

	docDB, err := s.getDocumentDB()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	filter := map[string]interface{}{field: value}

	result, err := docDB.FindOne(ctx, collectionName, filter)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get item by field %s: %v", field, err))
		return nil, err
	}

	return result, nil
}

// InsertItem inserts a new item into the collection
func (s *CollectionService) InsertItem(collectionName string, data map[string]interface{}) (string, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		s.iLog.PerformanceWithDuration("CollectionService.InsertItem", elapsed)
	}()

	docDB, err := s.getDocumentDB()
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	id, err := docDB.InsertOne(ctx, collectionName, data)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to insert item: %v", err))
		return "", err
	}

	return id, nil
}

// UpdateItem updates an item in the collection
func (s *CollectionService) UpdateItem(collectionName string, filter map[string]interface{}, update map[string]interface{}) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		s.iLog.PerformanceWithDuration("CollectionService.UpdateItem", elapsed)
	}()

	docDB, err := s.getDocumentDB()
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = docDB.UpdateOne(ctx, collectionName, filter, update)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to update item: %v", err))
		return err
	}

	return nil
}

// DeleteItem deletes an item from the collection
func (s *CollectionService) DeleteItem(collectionName string, filter map[string]interface{}) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		s.iLog.PerformanceWithDuration("CollectionService.DeleteItem", elapsed)
	}()

	docDB, err := s.getDocumentDB()
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = docDB.DeleteOne(ctx, collectionName, filter)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to delete item: %v", err))
		return err
	}

	return nil
}

// convertToIntMap converts a map[string]interface{} to map[string]int for projections
func convertToIntMap(m map[string]interface{}) map[string]int {
	if m == nil {
		return nil
	}

	result := make(map[string]int)
	for k, v := range m {
		switch val := v.(type) {
		case int:
			result[k] = val
		case int64:
			result[k] = int(val)
		case float64:
			result[k] = int(val)
		case bool:
			if val {
				result[k] = 1
			} else {
				result[k] = 0
			}
		default:
			result[k] = 1 // Default to include
		}
	}
	return result
}

// Legacy compatibility functions for backward compatibility with old DocDB interface

// QueryCollectionLegacy provides backward compatibility with the old DocDB.QueryCollection method
func (s *CollectionService) QueryCollectionLegacy(collectionName string, filter bson.M, projection bson.M) ([]bson.M, error) {
	opts := &QueryOptions{
		Filter:     convertBsonMToMap(filter),
		Projection: convertBsonMToMap(projection),
		PageSize:   0, // No pagination for legacy calls
	}

	// For legacy calls without pagination, we need to fetch all documents
	docDB, err := s.getDocumentDB()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	findOpts := &documents.FindOptions{
		Projection: convertToIntMap(opts.Projection),
	}

	results, err := docDB.FindMany(ctx, collectionName, opts.Filter, findOpts)
	if err != nil {
		return nil, err
	}

	// Convert back to bson.M for compatibility
	bsonResults := make([]bson.M, len(results))
	for i, result := range results {
		bsonResults[i] = bson.M(result)
	}

	return bsonResults, nil
}

// convertBsonMToMap converts bson.M to map[string]interface{}
func convertBsonMToMap(m bson.M) map[string]interface{} {
	if m == nil {
		return nil
	}
	return map[string]interface{}(m)
}
