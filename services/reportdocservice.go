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

	"go.mongodb.org/mongo-driver/bson"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/documents/models"
	"github.com/mdaxf/iac/documents/schema"
	"github.com/mdaxf/iac/logger"
)

// ReportDocService handles business logic for report documents in MongoDB
type ReportDocService struct {
	reportModel *models.ReportModel
	iLog        logger.Log
}

// NewReportDocService creates a new report document service
func NewReportDocService(docDB documents.DocumentDB) *ReportDocService {
	return &ReportDocService{
		reportModel: models.NewReportModel(docDB),
		iLog:        logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "ReportDocService"},
	}
}

// CreateReport creates a new report document
func (s *ReportDocService) CreateReport(ctx context.Context, report *schema.ReportDocument, createdBy string) (string, error) {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("CreateReport started for report: %s by user: %s", report.Name, createdBy))

	// Validate report
	if report.Name == "" {
		return "", fmt.Errorf("report name is required")
	}

	// Set audit fields
	now := time.Now()
	report.CreatedBy = createdBy
	report.CreatedOn = now
	report.ModifiedBy = createdBy
	report.ModifiedOn = now
	report.Active = true
	report.RowVersionStamp = 1

	if report.Revision == 0 {
		report.Revision = 1
	}

	// Set default values
	if report.ReportType == "" {
		report.ReportType = "manual"
	}

	// Set audit fields for embedded datasources
	for i := range report.Datasources {
		report.Datasources[i].CreatedBy = createdBy
		report.Datasources[i].CreatedOn = now
		report.Datasources[i].ModifiedBy = createdBy
		report.Datasources[i].ModifiedOn = now
		report.Datasources[i].Active = true
		report.Datasources[i].RowVersionStamp = 1
		s.iLog.Debug(fmt.Sprintf("Set audit fields for datasource: %s", report.Datasources[i].Alias))
	}

	// Set audit fields for embedded components
	for i := range report.Components {
		report.Components[i].CreatedBy = createdBy
		report.Components[i].CreatedOn = now
		report.Components[i].ModifiedBy = createdBy
		report.Components[i].ModifiedOn = now
		report.Components[i].Active = true
		report.Components[i].RowVersionStamp = 1
		s.iLog.Debug(fmt.Sprintf("Set audit fields for component: %s (type: %s)", report.Components[i].Name, report.Components[i].ComponentType))
	}

	s.iLog.Info(fmt.Sprintf("Creating report with %d datasources and %d components", len(report.Datasources), len(report.Components)))

	// Create the report
	id, err := s.reportModel.CreateReport(ctx, report)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to create report: %v", err))
		return "", fmt.Errorf("failed to create report: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Report created successfully with ID: %s in %v", id, time.Since(startTime)))
	return id, nil
}

// GetReportByID retrieves a report by its ID
func (s *ReportDocService) GetReportByID(ctx context.Context, id string) (*schema.ReportDocument, error) {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("GetReportByID started for ID: %s", id))

	report, err := s.reportModel.GetReportByID(ctx, id)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get report by ID: %v", err))
		return nil, fmt.Errorf("failed to get report by ID: %w", err)
	}

	s.iLog.Debug(fmt.Sprintf("GetReportByID completed in %v", time.Since(startTime)))
	return report, nil
}

// GetReportByName retrieves a report by its name
func (s *ReportDocService) GetReportByName(ctx context.Context, name string) (*schema.ReportDocument, error) {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("GetReportByName started for name: %s", name))

	report, err := s.reportModel.GetReportByName(ctx, name)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get report by name: %v", err))
		return nil, fmt.Errorf("failed to get report by name: %w", err)
	}

	s.iLog.Debug(fmt.Sprintf("GetReportByName completed in %v", time.Since(startTime)))
	return report, nil
}

// ListReports retrieves reports with filtering, sorting, and pagination
func (s *ReportDocService) ListReports(ctx context.Context, userID string, isPublic bool, category string, reportType string, page, pageSize int) ([]*schema.ReportDocument, int64, error) {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("ListReports started - userID: %s, isPublic: %v, category: %s, reportType: %s, page: %d, pageSize: %d",
		userID, isPublic, category, reportType, page, pageSize))

	// Build filter
	filter := bson.M{"active": true}

	// Access control filter
	if !isPublic {
		filter["$or"] = []bson.M{
			{"createdby": userID},
			{"ispublic": true},
		}
	} else {
		filter["ispublic"] = true
	}

	// Category filter
	if category != "" {
		filter["category"] = category
	}

	// Report type filter
	if reportType != "" {
		filter["reporttype"] = reportType
	}

	// Build find options
	opts := &documents.FindOptions{
		Sort:  map[string]int{"modifiedon": -1},
		Limit: int64(pageSize),
		Skip:  int64((page - 1) * pageSize),
	}

	// Get reports
	reports, err := s.reportModel.ListReports(ctx, filter, opts)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to list reports: %v", err))
		return nil, 0, fmt.Errorf("failed to list reports: %w", err)
	}

	// Get total count
	total, err := s.reportModel.CountReports(ctx, filter)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to count reports: %v", err))
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	s.iLog.Debug(fmt.Sprintf("ListReports completed in %v, found %d reports (total: %d)", time.Since(startTime), len(reports), total))
	return reports, total, nil
}

// UpdateReport updates a report document
func (s *ReportDocService) UpdateReport(ctx context.Context, id string, updates map[string]interface{}, modifiedBy string) error {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("UpdateReport started for ID: %s by user: %s", id, modifiedBy))

	// Validate ID
	if id == "" {
		return fmt.Errorf("report ID is required")
	}

	// Build update document
	update := bson.M{
		"$set": bson.M{
			"modifiedby": modifiedBy,
			"modifiedon": time.Now(),
		},
	}

	// Add user updates to $set
	updateSet := update["$set"].(bson.M)
	for key, value := range updates {
		// Skip audit fields that shouldn't be directly updated
		if key == "createdby" || key == "createdon" || key == "rowversionstamp" {
			continue
		}
		updateSet[key] = value
	}

	// Update the report
	err := s.reportModel.UpdateReport(ctx, id, update)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to update report: %v", err))
		return fmt.Errorf("failed to update report: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Report updated successfully: %s in %v", id, time.Since(startTime)))
	return nil
}

// DeleteReport soft deletes a report
func (s *ReportDocService) DeleteReport(ctx context.Context, id string, deletedBy string) error {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("DeleteReport started for ID: %s by user: %s", id, deletedBy))

	if id == "" {
		return fmt.Errorf("report ID is required")
	}

	// Perform soft delete
	err := s.reportModel.DeleteReport(ctx, id)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to delete report: %v", err))
		return fmt.Errorf("failed to delete report: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Report deleted successfully: %s in %v", id, time.Since(startTime)))
	return nil
}

// UpdateReportRevision increments the revision number
func (s *ReportDocService) UpdateReportRevision(ctx context.Context, id string, modifiedBy string) error {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("UpdateReportRevision started for ID: %s by user: %s", id, modifiedBy))

	if id == "" {
		return fmt.Errorf("report ID is required")
	}

	err := s.reportModel.UpdateReportRevision(ctx, id, modifiedBy)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to update report revision: %v", err))
		return fmt.Errorf("failed to update report revision: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Report revision updated successfully: %s in %v", id, time.Since(startTime)))
	return nil
}

// AddDatasource adds a data source to a report
func (s *ReportDocService) AddDatasource(ctx context.Context, reportID string, datasource schema.ReportDatasourceDoc, modifiedBy string) error {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("AddDatasource started for report ID: %s", reportID))

	if reportID == "" {
		return fmt.Errorf("report ID is required")
	}

	// Set audit fields for datasource
	now := time.Now()
	datasource.Active = true
	datasource.CreatedBy = modifiedBy
	datasource.CreatedOn = now
	datasource.ModifiedBy = modifiedBy
	datasource.ModifiedOn = now
	datasource.RowVersionStamp = 1

	update := bson.M{
		"$push": bson.M{
			"datasources": datasource,
		},
		"$set": bson.M{
			"modifiedby": modifiedBy,
			"modifiedon": now,
		},
		"$inc": bson.M{
			"rowversionstamp": 1,
		},
	}

	err := s.reportModel.UpdateReport(ctx, reportID, update)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to add datasource: %v", err))
		return fmt.Errorf("failed to add datasource: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Datasource added successfully to report: %s in %v", reportID, time.Since(startTime)))
	return nil
}

// AddComponent adds a component to a report
func (s *ReportDocService) AddComponent(ctx context.Context, reportID string, component schema.ReportComponentDoc, modifiedBy string) error {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("AddComponent started for report ID: %s", reportID))

	if reportID == "" {
		return fmt.Errorf("report ID is required")
	}

	// Set audit fields for component
	now := time.Now()
	component.Active = true
	component.CreatedBy = modifiedBy
	component.CreatedOn = now
	component.ModifiedBy = modifiedBy
	component.ModifiedOn = now
	component.RowVersionStamp = 1

	update := bson.M{
		"$push": bson.M{
			"components": component,
		},
		"$set": bson.M{
			"modifiedby": modifiedBy,
			"modifiedon": now,
		},
		"$inc": bson.M{
			"rowversionstamp": 1,
		},
	}

	err := s.reportModel.UpdateReport(ctx, reportID, update)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to add component: %v", err))
		return fmt.Errorf("failed to add component: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Component added successfully to report: %s in %v", reportID, time.Since(startTime)))
	return nil
}

// AddParameter adds a parameter to a report
func (s *ReportDocService) AddParameter(ctx context.Context, reportID string, parameter schema.ReportParameterDoc, modifiedBy string) error {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("AddParameter started for report ID: %s", reportID))

	if reportID == "" {
		return fmt.Errorf("report ID is required")
	}

	// Set audit fields for parameter
	now := time.Now()
	parameter.Active = true
	parameter.CreatedBy = modifiedBy
	parameter.CreatedOn = now
	parameter.ModifiedBy = modifiedBy
	parameter.ModifiedOn = now
	parameter.RowVersionStamp = 1

	update := bson.M{
		"$push": bson.M{
			"parameters": parameter,
		},
		"$set": bson.M{
			"modifiedby": modifiedBy,
			"modifiedon": now,
		},
		"$inc": bson.M{
			"rowversionstamp": 1,
		},
	}

	err := s.reportModel.UpdateReport(ctx, reportID, update)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to add parameter: %v", err))
		return fmt.Errorf("failed to add parameter: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Parameter added successfully to report: %s in %v", reportID, time.Since(startTime)))
	return nil
}

// AddExecution records a report execution
func (s *ReportDocService) AddExecution(ctx context.Context, reportID string, execution schema.ReportExecutionDoc) error {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("AddExecution started for report ID: %s", reportID))

	if reportID == "" {
		return fmt.Errorf("report ID is required")
	}

	// Set execution timestamp
	if execution.ExecutedOn.IsZero() {
		execution.ExecutedOn = time.Now()
	}

	err := s.reportModel.AddReportExecution(ctx, reportID, execution)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to add execution: %v", err))
		return fmt.Errorf("failed to add execution: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Execution added successfully to report: %s in %v", reportID, time.Since(startTime)))
	return nil
}

// SearchReports performs text search on reports
func (s *ReportDocService) SearchReports(ctx context.Context, searchText string, page, pageSize int) ([]*schema.ReportDocument, int64, error) {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("SearchReports started with text: %s, page: %d, pageSize: %d", searchText, page, pageSize))

	if searchText == "" {
		return nil, 0, fmt.Errorf("search text is required")
	}

	// Build find options
	opts := &documents.FindOptions{
		Sort:  map[string]int{"modifiedon": -1},
		Limit: int64(pageSize),
		Skip:  int64((page - 1) * pageSize),
	}

	// Search reports
	reports, err := s.reportModel.SearchReports(ctx, searchText, opts)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to search reports: %v", err))
		return nil, 0, fmt.Errorf("failed to search reports: %w", err)
	}

	// Get total count for search
	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": searchText, "$options": "i"}},
			{"description": bson.M{"$regex": searchText, "$options": "i"}},
		},
		"active": true,
	}
	total, err := s.reportModel.CountReports(ctx, filter)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to count search results: %v", err))
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	s.iLog.Debug(fmt.Sprintf("SearchReports completed in %v, found %d reports (total: %d)", time.Since(startTime), len(reports), total))
	return reports, total, nil
}

// GetReportsByCategory retrieves reports by category
func (s *ReportDocService) GetReportsByCategory(ctx context.Context, category string, page, pageSize int) ([]*schema.ReportDocument, int64, error) {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("GetReportsByCategory started for category: %s, page: %d, pageSize: %d", category, page, pageSize))

	if category == "" {
		return nil, 0, fmt.Errorf("category is required")
	}

	// Build find options
	opts := &documents.FindOptions{
		Sort:  map[string]int{"modifiedon": -1},
		Limit: int64(pageSize),
		Skip:  int64((page - 1) * pageSize),
	}

	// Get reports by category
	reports, err := s.reportModel.GetReportsByCategory(ctx, category, opts)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get reports by category: %v", err))
		return nil, 0, fmt.Errorf("failed to get reports by category: %w", err)
	}

	// Get total count
	filter := bson.M{"category": category, "active": true}
	total, err := s.reportModel.CountReports(ctx, filter)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to count reports by category: %v", err))
		return nil, 0, fmt.Errorf("failed to count reports by category: %w", err)
	}

	s.iLog.Debug(fmt.Sprintf("GetReportsByCategory completed in %v, found %d reports (total: %d)", time.Since(startTime), len(reports), total))
	return reports, total, nil
}

// GetDefaultReport retrieves the default report
func (s *ReportDocService) GetDefaultReport(ctx context.Context) (*schema.ReportDocument, error) {
	startTime := time.Now()
	s.iLog.Debug("GetDefaultReport started")

	report, err := s.reportModel.GetDefaultReport(ctx)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get default report: %v", err))
		return nil, fmt.Errorf("failed to get default report: %w", err)
	}

	s.iLog.Debug(fmt.Sprintf("GetDefaultReport completed in %v", time.Since(startTime)))
	return report, nil
}

// SetDefaultReport sets a report as the default
func (s *ReportDocService) SetDefaultReport(ctx context.Context, reportID string, modifiedBy string) error {
	startTime := time.Now()
	s.iLog.Debug(fmt.Sprintf("SetDefaultReport started for ID: %s", reportID))

	if reportID == "" {
		return fmt.Errorf("report ID is required")
	}

	// First, unset any existing default reports
	unsetFilter := bson.M{"isdefault": true, "active": true}
	unsetUpdate := bson.M{
		"$set": bson.M{
			"isdefault":  false,
			"modifiedby": modifiedBy,
			"modifiedon": time.Now(),
		},
		"$inc": bson.M{
			"rowversionstamp": 1,
		},
	}

	_, err := s.reportModel.UpdateManyReports(ctx, unsetFilter, unsetUpdate)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to unset existing default reports: %v", err))
		return fmt.Errorf("failed to unset existing default reports: %w", err)
	}

	// Set the new default report
	updates := map[string]interface{}{
		"isdefault": true,
	}

	err = s.UpdateReport(ctx, reportID, updates, modifiedBy)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to set default report: %v", err))
		return fmt.Errorf("failed to set default report: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Default report set successfully: %s in %v", reportID, time.Since(startTime)))
	return nil
}

// InitializeIndexes creates the required indexes for the Reports collection
func (s *ReportDocService) InitializeIndexes(ctx context.Context) error {
	startTime := time.Now()
	s.iLog.Debug("InitializeIndexes started")

	err := s.reportModel.CreateIndexes(ctx)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to initialize indexes: %v", err))
		return fmt.Errorf("failed to initialize indexes: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Indexes initialized successfully in %v", time.Since(startTime)))
	return nil
}
