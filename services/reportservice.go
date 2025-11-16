package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// ReportService handles report business logic
type ReportService struct {
	DB *gorm.DB
}

// NewReportService creates a new report service
func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{DB: db}
}

// CreateReport creates a new report
func (s *ReportService) CreateReport(report *models.Report) error {
	// Generate UUID if not provided
	if report.ID == "" {
		report.ID = uuid.New().String()
	}

	// Set defaults
	if report.Version == 0 {
		report.Version = 1
	}

	return s.DB.Create(report).Error
}

// GetReportByID retrieves a report by ID with all relationships
func (s *ReportService) GetReportByID(id string) (*models.Report, error) {
	var report models.Report
	err := s.DB.Preload("Datasources").
		Preload("Components").
		Preload("Parameters").
		Preload("Executions").
		Preload("Shares").
		First(&report, "id = ?", id).Error

	if err != nil {
		return nil, err
	}

	return &report, nil
}

// ListReports retrieves reports with pagination and filtering
func (s *ReportService) ListReports(userID string, isPublic bool, reportType string, page, pageSize int) ([]models.Report, int64, error) {
	var reports []models.Report
	var total int64

	query := s.DB.Model(&models.Report{})

	// Apply filters
	if !isPublic {
		query = query.Where("created_by = ? OR is_public = ?", userID, true)
	} else {
		query = query.Where("is_public = ?", true)
	}

	if reportType != "" {
		query = query.Where("report_type = ?", reportType)
	}

	query = query.Where("is_active = ?", true)

	// Get total count
	query.Count(&total)

	// Apply pagination
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("updated_at DESC").
		Find(&reports).Error

	if err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// UpdateReport updates an existing report
func (s *ReportService) UpdateReport(id string, updates map[string]interface{}) error {
	// Add updated_at timestamp
	updates["updated_at"] = time.Now()

	return s.DB.Model(&models.Report{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteReport soft deletes a report
func (s *ReportService) DeleteReport(id string) error {
	return s.DB.Model(&models.Report{}).Where("id = ?", id).Update("is_active", false).Error
}

// HardDeleteReport permanently deletes a report
func (s *ReportService) HardDeleteReport(id string) error {
	return s.DB.Where("id = ?", id).Delete(&models.Report{}).Error
}

// AddDatasource adds a datasource to a report
func (s *ReportService) AddDatasource(datasource *models.ReportDatasource) error {
	if datasource.ID == "" {
		datasource.ID = uuid.New().String()
	}

	return s.DB.Create(datasource).Error
}

// GetDatasources retrieves all datasources for a report
func (s *ReportService) GetDatasources(reportID string) ([]models.ReportDatasource, error) {
	var datasources []models.ReportDatasource
	err := s.DB.Where("report_id = ?", reportID).Find(&datasources).Error
	return datasources, err
}

// UpdateDatasource updates a datasource
func (s *ReportService) UpdateDatasource(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.DB.Model(&models.ReportDatasource{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteDatasource deletes a datasource
func (s *ReportService) DeleteDatasource(id string) error {
	return s.DB.Where("id = ?", id).Delete(&models.ReportDatasource{}).Error
}

// AddComponent adds a component to a report
func (s *ReportService) AddComponent(component *models.ReportComponent) error {
	if component.ID == "" {
		component.ID = uuid.New().String()
	}

	return s.DB.Create(component).Error
}

// GetComponents retrieves all components for a report
func (s *ReportService) GetComponents(reportID string) ([]models.ReportComponent, error) {
	var components []models.ReportComponent
	err := s.DB.Where("report_id = ? AND is_visible = ?", reportID, true).
		Order("z_index ASC").
		Find(&components).Error
	return components, err
}

// UpdateComponent updates a component
func (s *ReportService) UpdateComponent(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.DB.Model(&models.ReportComponent{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteComponent deletes a component
func (s *ReportService) DeleteComponent(id string) error {
	return s.DB.Where("id = ?", id).Delete(&models.ReportComponent{}).Error
}

// AddParameter adds a parameter to a report
func (s *ReportService) AddParameter(parameter *models.ReportParameter) error {
	if parameter.ID == "" {
		parameter.ID = uuid.New().String()
	}

	return s.DB.Create(parameter).Error
}

// GetParameters retrieves all parameters for a report
func (s *ReportService) GetParameters(reportID string) ([]models.ReportParameter, error) {
	var parameters []models.ReportParameter
	err := s.DB.Where("report_id = ? AND is_enabled = ?", reportID, true).
		Order("sort_order ASC").
		Find(&parameters).Error
	return parameters, err
}

// UpdateParameter updates a parameter
func (s *ReportService) UpdateParameter(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.DB.Model(&models.ReportParameter{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteParameter deletes a parameter
func (s *ReportService) DeleteParameter(id string) error {
	return s.DB.Where("id = ?", id).Delete(&models.ReportParameter{}).Error
}

// CreateExecution creates a new execution record
func (s *ReportService) CreateExecution(execution *models.ReportExecution) error {
	if execution.ID == "" {
		execution.ID = uuid.New().String()
	}

	return s.DB.Create(execution).Error
}

// UpdateExecution updates an execution record
func (s *ReportService) UpdateExecution(id string, updates map[string]interface{}) error {
	return s.DB.Model(&models.ReportExecution{}).Where("id = ?", id).Updates(updates).Error
}

// GetExecutionHistory retrieves execution history for a report
func (s *ReportService) GetExecutionHistory(reportID string, limit int) ([]models.ReportExecution, error) {
	var executions []models.ReportExecution
	err := s.DB.Where("report_id = ?", reportID).
		Order("created_at DESC").
		Limit(limit).
		Find(&executions).Error
	return executions, err
}

// UpdateLastExecutedAt updates the last execution timestamp
func (s *ReportService) UpdateLastExecutedAt(reportID string) error {
	now := time.Now()
	return s.DB.Model(&models.Report{}).Where("id = ?", reportID).Update("last_executed_at", now).Error
}

// ShareReport creates a share record
func (s *ReportService) ShareReport(share *models.ReportShare) error {
	if share.ID == "" {
		share.ID = uuid.New().String()
	}

	// Generate share token if not provided
	if share.ShareToken == "" {
		share.ShareToken = uuid.New().String()
	}

	return s.DB.Create(share).Error
}

// GetShares retrieves all shares for a report
func (s *ReportService) GetShares(reportID string) ([]models.ReportShare, error) {
	var shares []models.ReportShare
	err := s.DB.Where("report_id = ? AND is_active = ?", reportID, true).Find(&shares).Error
	return shares, err
}

// RevokeShare deactivates a share
func (s *ReportService) RevokeShare(id string) error {
	return s.DB.Model(&models.ReportShare{}).Where("id = ?", id).Update("is_active", false).Error
}

// GetShareByToken retrieves a share by token
func (s *ReportService) GetShareByToken(token string) (*models.ReportShare, error) {
	var share models.ReportShare
	err := s.DB.Where("share_token = ? AND is_active = ?", token, true).First(&share).Error
	if err != nil {
		return nil, err
	}

	// Check expiration
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("share link has expired")
	}

	return &share, nil
}

// ListTemplates retrieves report templates
func (s *ReportService) ListTemplates(category string, isPublic bool) ([]models.ReportTemplate, error) {
	var templates []models.ReportTemplate
	query := s.DB.Model(&models.ReportTemplate{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if isPublic {
		query = query.Where("is_public = ?", true)
	}

	err := query.Order("usage_count DESC, rating DESC").Find(&templates).Error
	return templates, err
}

// GetTemplateByID retrieves a template by ID
func (s *ReportService) GetTemplateByID(id string) (*models.ReportTemplate, error) {
	var template models.ReportTemplate
	err := s.DB.First(&template, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return &template, nil
}

// CreateFromTemplate creates a report from a template
func (s *ReportService) CreateFromTemplate(templateID, userID string) (*models.Report, error) {
	// Get template
	template, err := s.GetTemplateByID(templateID)
	if err != nil {
		return nil, err
	}

	// Create report from template
	report := &models.Report{
		ID:               uuid.New().String(),
		Name:             template.Name + " (Copy)",
		Description:      template.Description,
		ReportType:       models.ReportTypeTemplate,
		CreatedBy:        userID,
		IsPublic:         false,
		TemplateSourceID: templateID,
		LayoutConfig:     template.TemplateConfig,
		Version:          1,
		IsActive:         true,
	}

	err = s.CreateReport(report)
	if err != nil {
		return nil, err
	}

	// Increment template usage count
	s.DB.Model(&models.ReportTemplate{}).Where("id = ?", templateID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1"))

	return report, nil
}

// DuplicateReport creates a copy of an existing report
func (s *ReportService) DuplicateReport(reportID, userID string) (*models.Report, error) {
	// Get original report with all relationships
	original, err := s.GetReportByID(reportID)
	if err != nil {
		return nil, err
	}

	// Create new report
	newReport := &models.Report{
		ID:           uuid.New().String(),
		Name:         original.Name + " (Copy)",
		Description:  original.Description,
		ReportType:   original.ReportType,
		CreatedBy:    userID,
		IsPublic:     false,
		LayoutConfig: original.LayoutConfig,
		PageSettings: original.PageSettings,
		Version:      1,
		IsActive:     true,
	}

	err = s.CreateReport(newReport)
	if err != nil {
		return nil, err
	}

	// Copy datasources
	for _, ds := range original.Datasources {
		newDS := ds
		newDS.ID = uuid.New().String()
		newDS.ReportID = newReport.ID
		s.AddDatasource(&newDS)
	}

	// Copy components
	for _, comp := range original.Components {
		newComp := comp
		newComp.ID = uuid.New().String()
		newComp.ReportID = newReport.ID
		s.AddComponent(&newComp)
	}

	// Copy parameters
	for _, param := range original.Parameters {
		newParam := param
		newParam.ID = uuid.New().String()
		newParam.ReportID = newReport.ID
		s.AddParameter(&newParam)
	}

	return newReport, nil
}

// SearchReports searches reports by name or description
func (s *ReportService) SearchReports(keyword string, userID string, limit int) ([]models.Report, error) {
	var reports []models.Report
	searchTerm := "%" + keyword + "%"

	err := s.DB.Where("(created_by = ? OR is_public = ?) AND is_active = ? AND (name LIKE ? OR description LIKE ?)",
		userID, true, true, searchTerm, searchTerm).
		Limit(limit).
		Order("updated_at DESC").
		Find(&reports).Error

	return reports, err
}
