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
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/documents/schema"
	"github.com/mdaxf/iac/logger"
)

// ReportModel handles MongoDB operations for report documents
type ReportModel struct {
	docDB documents.DocumentDB
	iLog  logger.Log
}

// NewReportModel creates a new report model instance
func NewReportModel(docDB documents.DocumentDB) *ReportModel {
	return &ReportModel{
		docDB: docDB,
		iLog:  logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "ReportModel"},
	}
}

// CreateReport creates a new report document in MongoDB
func (m *ReportModel) CreateReport(ctx context.Context, report *schema.ReportDocument) (string, error) {
	startTime := time.Now()
	m.iLog.Debug(fmt.Sprintf("CreateReport started for report: %s", report.Name))

	// Set default values
	now := time.Now()
	if report.CreatedOn.IsZero() {
		report.CreatedOn = now
	}
	if report.ModifiedOn.IsZero() {
		report.ModifiedOn = now
	}
	if report.RowVersionStamp == 0 {
		report.RowVersionStamp = 1
	}
	if report.Revision == 0 {
		report.Revision = 1
	}
	report.Active = true

	// Insert the document
	id, err := m.docDB.InsertOne(ctx, schema.ReportsCollection, report)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to create report: %v", err))
		return "", fmt.Errorf("failed to create report: %w", err)
	}

	m.iLog.Debug(fmt.Sprintf("CreateReport completed in %v, ID: %s", time.Since(startTime), id))
	return id, nil
}

// GetReportByID retrieves a report document by its ID
func (m *ReportModel) GetReportByID(ctx context.Context, id string) (*schema.ReportDocument, error) {
	startTime := time.Now()
	m.iLog.Debug(fmt.Sprintf("GetReportByID started for ID: %s", id))

	// Convert string ID to ObjectID if needed
	var filter bson.M
	if oid, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": oid}
	} else {
		filter = bson.M{"_id": id}
	}

	result, err := m.docDB.FindOne(ctx, schema.ReportsCollection, filter)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to get report by ID: %v", err))
		return nil, fmt.Errorf("failed to get report by ID: %w", err)
	}

	// Convert map to ReportDocument
	report, err := m.mapToReportDocument(result)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to convert map to report document: %v", err))
		return nil, fmt.Errorf("failed to convert result to report document: %w", err)
	}

	m.iLog.Debug(fmt.Sprintf("GetReportByID completed in %v", time.Since(startTime)))
	return report, nil
}

// GetReportByName retrieves a report document by name
func (m *ReportModel) GetReportByName(ctx context.Context, name string) (*schema.ReportDocument, error) {
	startTime := time.Now()
	m.iLog.Debug(fmt.Sprintf("GetReportByName started for name: %s", name))

	filter := bson.M{"name": name, "active": true}

	result, err := m.docDB.FindOne(ctx, schema.ReportsCollection, filter)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to get report by name: %v", err))
		return nil, fmt.Errorf("failed to get report by name: %w", err)
	}

	report, err := m.mapToReportDocument(result)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to convert map to report document: %v", err))
		return nil, fmt.Errorf("failed to convert result to report document: %w", err)
	}

	m.iLog.Debug(fmt.Sprintf("GetReportByName completed in %v", time.Since(startTime)))
	return report, nil
}

// ListReports retrieves reports with filtering, sorting, and pagination
func (m *ReportModel) ListReports(ctx context.Context, filter bson.M, opts *documents.FindOptions) ([]*schema.ReportDocument, error) {
	startTime := time.Now()
	m.iLog.Debug("ListReports started")

	// Ensure active filter is applied
	if filter == nil {
		filter = bson.M{}
	}
	if _, exists := filter["active"]; !exists {
		filter["active"] = true
	}

	results, err := m.docDB.FindMany(ctx, schema.ReportsCollection, filter, opts)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to list reports: %v", err))
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}

	reports := make([]*schema.ReportDocument, 0, len(results))
	for _, result := range results {
		report, err := m.mapToReportDocument(result)
		if err != nil {
			m.iLog.Error(fmt.Sprintf("Failed to convert map to report document: %v", err))
			continue
		}
		reports = append(reports, report)
	}

	m.iLog.Debug(fmt.Sprintf("ListReports completed in %v, found %d reports", time.Since(startTime), len(reports)))
	return reports, nil
}

// UpdateReport updates a report document
func (m *ReportModel) UpdateReport(ctx context.Context, id string, update bson.M) error {
	startTime := time.Now()
	m.iLog.Debug(fmt.Sprintf("UpdateReport started for ID: %s", id))

	// Convert string ID to ObjectID if needed
	var filter bson.M
	if oid, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": oid}
	} else {
		filter = bson.M{"_id": id}
	}

	// Ensure modifiedon is updated
	if update["$set"] == nil {
		update["$set"] = bson.M{}
	}
	updateSet := update["$set"].(bson.M)
	updateSet["modifiedon"] = time.Now()

	// Increment row version stamp
	if update["$inc"] == nil {
		update["$inc"] = bson.M{}
	}
	updateInc := update["$inc"].(bson.M)
	updateInc["rowversionstamp"] = 1

	err := m.docDB.UpdateOne(ctx, schema.ReportsCollection, filter, update)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to update report: %v", err))
		return fmt.Errorf("failed to update report: %w", err)
	}

	m.iLog.Debug(fmt.Sprintf("UpdateReport completed in %v", time.Since(startTime)))
	return nil
}

// UpdateManyReports updates multiple report documents matching the filter
func (m *ReportModel) UpdateManyReports(ctx context.Context, filter bson.M, update bson.M) (int64, error) {
	startTime := time.Now()
	m.iLog.Debug("UpdateManyReports started")

	count, err := m.docDB.UpdateMany(ctx, schema.ReportsCollection, filter, update)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to update multiple reports: %v", err))
		return 0, fmt.Errorf("failed to update multiple reports: %w", err)
	}

	m.iLog.Debug(fmt.Sprintf("UpdateManyReports completed in %v, updated %d documents", time.Since(startTime), count))
	return count, nil
}

// DeleteReport soft deletes a report document (sets active to false)
func (m *ReportModel) DeleteReport(ctx context.Context, id string) error {
	startTime := time.Now()
	m.iLog.Debug(fmt.Sprintf("DeleteReport started for ID: %s", id))

	update := bson.M{
		"$set": bson.M{
			"active":     false,
			"modifiedon": time.Now(),
		},
		"$inc": bson.M{
			"rowversionstamp": 1,
		},
	}

	err := m.UpdateReport(ctx, id, update)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to delete report: %v", err))
		return fmt.Errorf("failed to delete report: %w", err)
	}

	m.iLog.Debug(fmt.Sprintf("DeleteReport completed in %v", time.Since(startTime)))
	return nil
}

// HardDeleteReport permanently deletes a report document
func (m *ReportModel) HardDeleteReport(ctx context.Context, id string) error {
	startTime := time.Now()
	m.iLog.Debug(fmt.Sprintf("HardDeleteReport started for ID: %s", id))

	// Convert string ID to ObjectID if needed
	var filter bson.M
	if oid, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": oid}
	} else {
		filter = bson.M{"_id": id}
	}

	err := m.docDB.DeleteOne(ctx, schema.ReportsCollection, filter)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to hard delete report: %v", err))
		return fmt.Errorf("failed to hard delete report: %w", err)
	}

	m.iLog.Debug(fmt.Sprintf("HardDeleteReport completed in %v", time.Since(startTime)))
	return nil
}

// CountReports returns the count of reports matching the filter
func (m *ReportModel) CountReports(ctx context.Context, filter bson.M) (int64, error) {
	startTime := time.Now()
	m.iLog.Debug("CountReports started")

	// Ensure active filter is applied
	if filter == nil {
		filter = bson.M{}
	}
	if _, exists := filter["active"]; !exists {
		filter["active"] = true
	}

	count, err := m.docDB.CountDocuments(ctx, schema.ReportsCollection, filter)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to count reports: %v", err))
		return 0, fmt.Errorf("failed to count reports: %w", err)
	}

	m.iLog.Debug(fmt.Sprintf("CountReports completed in %v, count: %d", time.Since(startTime), count))
	return count, nil
}

// UpdateReportRevision creates a new revision of the report
func (m *ReportModel) UpdateReportRevision(ctx context.Context, id string, modifiedBy string) error {
	startTime := time.Now()
	m.iLog.Debug(fmt.Sprintf("UpdateReportRevision started for ID: %s", id))

	update := bson.M{
		"$set": bson.M{
			"modifiedby": modifiedBy,
			"modifiedon": time.Now(),
		},
		"$inc": bson.M{
			"revision":        1,
			"rowversionstamp": 1,
		},
	}

	err := m.UpdateReport(ctx, id, update)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to update report revision: %v", err))
		return fmt.Errorf("failed to update report revision: %w", err)
	}

	m.iLog.Debug(fmt.Sprintf("UpdateReportRevision completed in %v", time.Since(startTime)))
	return nil
}

// AddReportExecution adds an execution record to the report's recent executions
func (m *ReportModel) AddReportExecution(ctx context.Context, id string, execution schema.ReportExecutionDoc) error {
	startTime := time.Now()
	m.iLog.Debug(fmt.Sprintf("AddReportExecution started for report ID: %s", id))

	// Convert string ID to ObjectID if needed
	var filter bson.M
	if oid, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": oid}
	} else {
		filter = bson.M{"_id": id}
	}

	// Push the new execution and keep only the last 10 executions
	update := bson.M{
		"$push": bson.M{
			"recentexecutions": bson.M{
				"$each":  []interface{}{execution},
				"$slice": -10, // Keep only the last 10 executions
			},
		},
		"$set": bson.M{
			"lastexecutedon": execution.ExecutedOn,
			"modifiedon":     time.Now(),
		},
		"$inc": bson.M{
			"rowversionstamp": 1,
		},
	}

	err := m.docDB.UpdateOne(ctx, schema.ReportsCollection, filter, update)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to add report execution: %v", err))
		return fmt.Errorf("failed to add report execution: %w", err)
	}

	m.iLog.Debug(fmt.Sprintf("AddReportExecution completed in %v", time.Since(startTime)))
	return nil
}

// GetReportsByCategory retrieves reports by category
func (m *ReportModel) GetReportsByCategory(ctx context.Context, category string, opts *documents.FindOptions) ([]*schema.ReportDocument, error) {
	startTime := time.Now()
	m.iLog.Debug(fmt.Sprintf("GetReportsByCategory started for category: %s", category))

	filter := bson.M{
		"category": category,
		"active":   true,
	}

	reports, err := m.ListReports(ctx, filter, opts)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to get reports by category: %v", err))
		return nil, fmt.Errorf("failed to get reports by category: %w", err)
	}

	m.iLog.Debug(fmt.Sprintf("GetReportsByCategory completed in %v, found %d reports", time.Since(startTime), len(reports)))
	return reports, nil
}

// GetDefaultReport retrieves the default report
func (m *ReportModel) GetDefaultReport(ctx context.Context) (*schema.ReportDocument, error) {
	startTime := time.Now()
	m.iLog.Debug("GetDefaultReport started")

	filter := bson.M{
		"isdefault": true,
		"active":    true,
	}

	result, err := m.docDB.FindOne(ctx, schema.ReportsCollection, filter)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to get default report: %v", err))
		return nil, fmt.Errorf("failed to get default report: %w", err)
	}

	report, err := m.mapToReportDocument(result)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to convert map to report document: %v", err))
		return nil, fmt.Errorf("failed to convert result to report document: %w", err)
	}

	m.iLog.Debug(fmt.Sprintf("GetDefaultReport completed in %v", time.Since(startTime)))
	return report, nil
}

// SearchReports performs a text search on report names and descriptions
func (m *ReportModel) SearchReports(ctx context.Context, searchText string, opts *documents.FindOptions) ([]*schema.ReportDocument, error) {
	startTime := time.Now()
	m.iLog.Debug(fmt.Sprintf("SearchReports started with text: %s", searchText))

	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": searchText, "$options": "i"}},
			{"description": bson.M{"$regex": searchText, "$options": "i"}},
		},
		"active": true,
	}

	reports, err := m.ListReports(ctx, filter, opts)
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to search reports: %v", err))
		return nil, fmt.Errorf("failed to search reports: %w", err)
	}

	m.iLog.Debug(fmt.Sprintf("SearchReports completed in %v, found %d reports", time.Since(startTime), len(reports)))
	return reports, nil
}

// CreateIndexes creates the required indexes for the Reports collection
func (m *ReportModel) CreateIndexes(ctx context.Context) error {
	startTime := time.Now()
	m.iLog.Debug("CreateIndexes started")

	indexDefs := schema.GetReportIndexDefinitions()

	for _, indexDef := range indexDefs {
		keys := indexDef["keys"].(map[string]int)
		opts := &documents.IndexOptions{}

		if options, ok := indexDef["options"].(map[string]interface{}); ok {
			if name, ok := options["name"].(string); ok {
				opts.Name = name
			}
			if unique, ok := options["unique"].(bool); ok {
				opts.Unique = unique
			}
			if background, ok := options["background"].(bool); ok {
				opts.Background = background
			}
			if sparse, ok := options["sparse"].(bool); ok {
				opts.Sparse = sparse
			}
		}

		err := m.docDB.CreateIndex(ctx, schema.ReportsCollection, keys, opts)
		if err != nil {
			m.iLog.Error(fmt.Sprintf("Failed to create index: %v", err))
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	m.iLog.Debug(fmt.Sprintf("CreateIndexes completed in %v", time.Since(startTime)))
	return nil
}

// Helper function to convert map to ReportDocument
func (m *ReportModel) mapToReportDocument(data map[string]interface{}) (*schema.ReportDocument, error) {
	// Convert map to BSON bytes
	bsonBytes, err := bson.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map to BSON: %w", err)
	}

	// Unmarshal BSON bytes to ReportDocument
	var report schema.ReportDocument
	if err := bson.Unmarshal(bsonBytes, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal BSON to ReportDocument: %w", err)
	}

	// Handle _id conversion
	if idVal, ok := data["_id"]; ok {
		if idStr, ok := idVal.(string); ok {
			report.ID = idStr
		}
	}

	return &report, nil
}
