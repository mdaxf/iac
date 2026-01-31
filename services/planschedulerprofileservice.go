// Copyright 2023 IAC. All Rights Reserved.

package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// PlanSchedulerProfileService handles plan scheduler profile operations
type PlanSchedulerProfileService struct {
	db   *sql.DB
	gorm *gorm.DB
	iLog logger.Log
}

// NewPlanSchedulerProfileService creates a new plan scheduler profile service
func NewPlanSchedulerProfileService(db *sql.DB, gormDB *gorm.DB) *PlanSchedulerProfileService {
	return &PlanSchedulerProfileService{
		db:   db,
		gorm: gormDB,
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "PlanSchedulerProfileService",
		},
	}
}

// GetAllProfiles retrieves all active profiles
func (s *PlanSchedulerProfileService) GetAllProfiles(ctx context.Context) ([]*models.PlanSchedulerProfile, error) {
	s.iLog.Info("GetAllProfiles START")

	var profiles []*models.PlanSchedulerProfile
	err := s.gorm.WithContext(ctx).
		Where("active = ?", true).
		Order("is_default DESC, name ASC").
		Find(&profiles).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get profiles: %v", err))
		return nil, fmt.Errorf("failed to get profiles: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("GetAllProfiles COMPLETE - Found %d profiles", len(profiles)))
	return profiles, nil
}

// GetProfileByID retrieves a profile by ID with all related data
func (s *PlanSchedulerProfileService) GetProfileByID(ctx context.Context, id string) (*models.PlanSchedulerProfile, error) {
	s.iLog.Info(fmt.Sprintf("GetProfileByID START - ID: %s", id))

	var profile models.PlanSchedulerProfile
	err := s.gorm.WithContext(ctx).
		Where("id = ? AND active = ?", id, true).
		First(&profile).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("profile not found: %s", id)
		}
		s.iLog.Error(fmt.Sprintf("Failed to get profile: %v", err))
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Load related data
	if err := s.loadProfileRelations(ctx, &profile); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to load profile relations: %v", err))
		return nil, fmt.Errorf("failed to load profile relations: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("GetProfileByID COMPLETE - Profile: %s", profile.Name))
	return &profile, nil
}

// GetDefaultProfile retrieves the default profile
func (s *PlanSchedulerProfileService) GetDefaultProfile(ctx context.Context) (*models.PlanSchedulerProfile, error) {
	s.iLog.Info("GetDefaultProfile START")

	var profile models.PlanSchedulerProfile
	err := s.gorm.WithContext(ctx).
		Where("is_default = ? AND active = ?", true, true).
		First(&profile).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no default profile found")
		}
		s.iLog.Error(fmt.Sprintf("Failed to get default profile: %v", err))
		return nil, fmt.Errorf("failed to get default profile: %w", err)
	}

	// Load related data
	if err := s.loadProfileRelations(ctx, &profile); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to load profile relations: %v", err))
		return nil, fmt.Errorf("failed to load profile relations: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("GetDefaultProfile COMPLETE - Profile: %s", profile.Name))
	return &profile, nil
}

// loadProfileRelations loads datasources, constraints, and settings for a profile
func (s *PlanSchedulerProfileService) loadProfileRelations(ctx context.Context, profile *models.PlanSchedulerProfile) error {
	// Load data sources
	err := s.gorm.WithContext(ctx).
		Where("profile_id = ? AND active = ?", profile.ID, true).
		Order("display_order ASC, name ASC").
		Find(&profile.DataSources).Error
	if err != nil {
		return fmt.Errorf("failed to load datasources: %w", err)
	}

	// Load constraints
	err = s.gorm.WithContext(ctx).
		Where("profile_id = ? AND active = ?", profile.ID, true).
		Order("display_order ASC, name ASC").
		Find(&profile.Constraints).Error
	if err != nil {
		return fmt.Errorf("failed to load constraints: %w", err)
	}

	// Load settings
	err = s.gorm.WithContext(ctx).
		Where("profile_id = ? AND active = ?", profile.ID, true).
		Find(&profile.Settings).Error
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	return nil
}

// CreateProfile creates a new profile
func (s *PlanSchedulerProfileService) CreateProfile(ctx context.Context, profile *models.PlanSchedulerProfile, createdBy string) error {
	s.iLog.Info(fmt.Sprintf("CreateProfile START - Name: %s", profile.Name))

	// Generate ID if not provided
	if profile.ID == "" {
		profile.ID = uuid.New().String()
	}

	// Set audit fields
	now := time.Now()
	profile.Active = true
	profile.CreatedBy = createdBy
	profile.CreatedOn = now
	profile.ModifiedBy = createdBy
	profile.ModifiedOn = now
	profile.RowVersionStamp = 0

	// If this is set as default, unset other defaults
	if profile.IsDefault {
		if err := s.unsetOtherDefaults(ctx, profile.ID); err != nil {
			return fmt.Errorf("failed to unset other defaults: %w", err)
		}
	}

	// Create profile
	err := s.gorm.WithContext(ctx).Create(profile).Error
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to create profile: %v", err))
		return fmt.Errorf("failed to create profile: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("CreateProfile COMPLETE - ID: %s", profile.ID))
	return nil
}

// UpdateProfile updates an existing profile
func (s *PlanSchedulerProfileService) UpdateProfile(ctx context.Context, profile *models.PlanSchedulerProfile, modifiedBy string) error {
	s.iLog.Info(fmt.Sprintf("UpdateProfile START - ID: %s", profile.ID))

	// Set audit fields
	profile.ModifiedBy = modifiedBy
	profile.ModifiedOn = time.Now()
	profile.RowVersionStamp++

	// If this is set as default, unset other defaults
	if profile.IsDefault {
		if err := s.unsetOtherDefaults(ctx, profile.ID); err != nil {
			return fmt.Errorf("failed to unset other defaults: %w", err)
		}
	}

	// Update profile
	err := s.gorm.WithContext(ctx).
		Model(&models.PlanSchedulerProfile{}).
		Where("id = ?", profile.ID).
		Updates(profile).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to update profile: %v", err))
		return fmt.Errorf("failed to update profile: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("UpdateProfile COMPLETE - ID: %s", profile.ID))
	return nil
}

// DeleteProfile soft-deletes a profile (sets active = false)
func (s *PlanSchedulerProfileService) DeleteProfile(ctx context.Context, id string, modifiedBy string) error {
	s.iLog.Info(fmt.Sprintf("DeleteProfile START - ID: %s", id))

	// Check if it's the default profile
	var profile models.PlanSchedulerProfile
	err := s.gorm.WithContext(ctx).Where("id = ?", id).First(&profile).Error
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	if profile.IsDefault {
		return fmt.Errorf("cannot delete default profile")
	}

	// Soft delete (set active = false)
	err = s.gorm.WithContext(ctx).
		Model(&models.PlanSchedulerProfile{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"active":      false,
			"modifiedby":  modifiedBy,
			"modifiedon":  time.Now(),
			"rowversionstamp": gorm.Expr("rowversionstamp + 1"),
		}).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to delete profile: %v", err))
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("DeleteProfile COMPLETE - ID: %s", id))
	return nil
}

// unsetOtherDefaults removes is_default flag from all other profiles
func (s *PlanSchedulerProfileService) unsetOtherDefaults(ctx context.Context, exceptID string) error {
	return s.gorm.WithContext(ctx).
		Model(&models.PlanSchedulerProfile{}).
		Where("id != ? AND is_default = ?", exceptID, true).
		Update("is_default", false).Error
}

// LoadProfileData loads data from all data sources in a profile
func (s *PlanSchedulerProfileService) LoadProfileData(ctx context.Context, profileID string, parameters map[string]interface{}) (*models.PlanSchedulerProfileData, error) {
	s.iLog.Info(fmt.Sprintf("LoadProfileData START - ProfileID: %s", profileID))

	// Get profile
	profile, err := s.GetProfileByID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	data := &models.PlanSchedulerProfileData{
		ProfileID:   profile.ID,
		ProfileName: profile.Name,
		MasterData:  make(map[string][]map[string]interface{}),
		Constraints: make(map[string]interface{}),
		Settings:    make(map[string]interface{}),
	}

	// Load tasks
	tasksData, err := s.loadDataByType(ctx, profile.DataSources, "tasks", parameters)
	if err != nil {
		s.iLog.Warn(fmt.Sprintf("Failed to load tasks: %v", err))
	} else {
		data.Tasks = tasksData
	}

	// Load resources
	resourcesData, err := s.loadDataByType(ctx, profile.DataSources, "resources", parameters)
	if err != nil {
		s.iLog.Warn(fmt.Sprintf("Failed to load resources: %v", err))
	} else {
		data.Resources = resourcesData
	}

	// Load other master data
	for _, ds := range profile.DataSources {
		if ds.DataType != "tasks" && ds.DataType != "resources" {
			masterData, err := s.loadDataSource(ctx, &ds, parameters)
			if err != nil {
				s.iLog.Warn(fmt.Sprintf("Failed to load %s: %v", ds.DataType, err))
				continue
			}
			data.MasterData[ds.DataType] = masterData
		}
	}

	// Load constraints
	for _, constraint := range profile.Constraints {
		constraintData, err := s.loadConstraint(ctx, &constraint, parameters)
		if err != nil {
			s.iLog.Warn(fmt.Sprintf("Failed to load constraint %s: %v", constraint.Name, err))
			continue
		}
		data.Constraints[constraint.ConstraintType] = constraintData
	}

	// Load settings
	for _, setting := range profile.Settings {
		var value interface{}
		if len(setting.SettingValue) > 0 {
			json.Unmarshal(setting.SettingValue, &value)
		}
		data.Settings[setting.SettingKey] = value
	}

	s.iLog.Info(fmt.Sprintf("LoadProfileData COMPLETE - Loaded %d tasks, %d resources", len(data.Tasks), len(data.Resources)))
	return data, nil
}

// loadDataByType loads data sources of a specific type
func (s *PlanSchedulerProfileService) loadDataByType(ctx context.Context, datasources []models.PlanSchedulerProfileDataSource, dataType string, parameters map[string]interface{}) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	for _, ds := range datasources {
		if ds.DataType == dataType {
			data, err := s.loadDataSource(ctx, &ds, parameters)
			if err != nil {
				return nil, err
			}
			results = append(results, data...)
		}
	}

	return results, nil
}

// loadDataSource loads data from a single data source
func (s *PlanSchedulerProfileService) loadDataSource(ctx context.Context, ds *models.PlanSchedulerProfileDataSource, parameters map[string]interface{}) ([]map[string]interface{}, error) {
	if ds.SourceType == "json" {
		// Load from JSON constant
		var jsonData []map[string]interface{}
		if err := json.Unmarshal(ds.SourceJSON, &jsonData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
		}
		return jsonData, nil
	} else if ds.SourceType == "query" {
		// Execute SQL query
		query := ds.SourceQuery

		// Substitute parameters
		if parameters != nil {
			query = s.substituteParameters(query, parameters)
		}

		// Also substitute datasource-specific parameters
		if len(ds.Parameters) > 0 {
			var dsParams map[string]interface{}
			json.Unmarshal(ds.Parameters, &dsParams)
			query = s.substituteParameters(query, dsParams)
		}

		return s.executeQuery(ctx, query)
	}

	return nil, fmt.Errorf("unknown source type: %s", ds.SourceType)
}

// loadConstraint loads constraint data
func (s *PlanSchedulerProfileService) loadConstraint(ctx context.Context, constraint *models.PlanSchedulerProfileConstraint, parameters map[string]interface{}) (interface{}, error) {
	if constraint.SourceType == "json" {
		var jsonData interface{}
		if err := json.Unmarshal(constraint.SourceJSON, &jsonData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON constraint: %w", err)
		}
		return jsonData, nil
	} else if constraint.SourceType == "query" {
		query := constraint.SourceQuery

		// Substitute parameters
		if parameters != nil {
			query = s.substituteParameters(query, parameters)
		}

		results, err := s.executeQuery(ctx, query)
		if err != nil {
			return nil, err
		}

		// Return first row if only one result, otherwise return all
		if len(results) == 1 {
			return results[0], nil
		}
		return results, nil
	}

	return nil, fmt.Errorf("unknown source type: %s", constraint.SourceType)
}

// substituteParameters replaces {{paramName}} placeholders in query
func (s *PlanSchedulerProfileService) substituteParameters(query string, parameters map[string]interface{}) string {
	for key, value := range parameters {
		placeholder := fmt.Sprintf("{{%s}}", key)
		replacement := fmt.Sprintf("%v", value)
		query = strings.ReplaceAll(query, placeholder, replacement)
	}
	return query
}

// executeQuery executes a SQL query and returns results as array of maps
func (s *PlanSchedulerProfileService) executeQuery(ctx context.Context, query string) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold each column value
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create map for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Handle []byte (convert to string or parse JSON)
			if b, ok := val.([]byte); ok {
				// Try to parse as JSON first
				var jsonVal interface{}
				if err := json.Unmarshal(b, &jsonVal); err == nil {
					rowMap[col] = jsonVal
				} else {
					rowMap[col] = string(b)
				}
			} else {
				rowMap[col] = val
			}
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// CreateDataSource creates a new data source for a profile
func (s *PlanSchedulerProfileService) CreateDataSource(ctx context.Context, dataSource *models.PlanSchedulerProfileDataSource, createdBy string) error {
	s.iLog.Info(fmt.Sprintf("CreateDataSource START - ProfileID: %s, Name: %s", dataSource.ProfileID, dataSource.Name))

	// Generate ID if not provided
	if dataSource.ID == "" {
		dataSource.ID = uuid.New().String()
	}

	// Set audit fields
	dataSource.Active = true
	dataSource.CreatedBy = createdBy
	dataSource.CreatedOn = time.Now()
	dataSource.ModifiedBy = createdBy
	dataSource.ModifiedOn = time.Now()
	dataSource.RowVersionStamp = 0

	// Create data source
	err := s.gorm.WithContext(ctx).Create(dataSource).Error
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to create data source: %v", err))
		return fmt.Errorf("failed to create data source: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("CreateDataSource COMPLETE - ID: %s", dataSource.ID))
	return nil
}

// UpdateDataSource updates an existing data source
func (s *PlanSchedulerProfileService) UpdateDataSource(ctx context.Context, dataSource *models.PlanSchedulerProfileDataSource, modifiedBy string) error {
	s.iLog.Info(fmt.Sprintf("UpdateDataSource START - ID: %s", dataSource.ID))

	// Set audit fields
	dataSource.ModifiedBy = modifiedBy
	dataSource.ModifiedOn = time.Now()
	dataSource.RowVersionStamp++

	// Update data source
	err := s.gorm.WithContext(ctx).
		Model(&models.PlanSchedulerProfileDataSource{}).
		Where("id = ? AND profile_id = ?", dataSource.ID, dataSource.ProfileID).
		Updates(dataSource).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to update data source: %v", err))
		return fmt.Errorf("failed to update data source: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("UpdateDataSource COMPLETE - ID: %s", dataSource.ID))
	return nil
}

// DeleteDataSource soft-deletes a data source (sets active = false)
func (s *PlanSchedulerProfileService) DeleteDataSource(ctx context.Context, profileID string, dataSourceID string, modifiedBy string) error {
	s.iLog.Info(fmt.Sprintf("DeleteDataSource START - ProfileID: %s, ID: %s", profileID, dataSourceID))

	err := s.gorm.WithContext(ctx).
		Model(&models.PlanSchedulerProfileDataSource{}).
		Where("id = ? AND profile_id = ?", dataSourceID, profileID).
		Updates(map[string]interface{}{
			"active":             false,
			"modified_by":        modifiedBy,
			"modified_on":        time.Now(),
			"row_version_stamp":  gorm.Expr("row_version_stamp + 1"),
		}).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to delete data source: %v", err))
		return fmt.Errorf("failed to delete data source: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("DeleteDataSource COMPLETE - ID: %s", dataSourceID))
	return nil
}

// CreateConstraint creates a new constraint for a profile
func (s *PlanSchedulerProfileService) CreateConstraint(ctx context.Context, constraint *models.PlanSchedulerProfileConstraint, createdBy string) error {
	s.iLog.Info(fmt.Sprintf("CreateConstraint START - ProfileID: %s, Name: %s", constraint.ProfileID, constraint.Name))

	// Generate ID if not provided
	if constraint.ID == "" {
		constraint.ID = uuid.New().String()
	}

	// Set audit fields
	constraint.Active = true
	constraint.CreatedBy = createdBy
	constraint.CreatedOn = time.Now()
	constraint.ModifiedBy = createdBy
	constraint.ModifiedOn = time.Now()
	constraint.RowVersionStamp = 0

	// Create constraint
	err := s.gorm.WithContext(ctx).Create(constraint).Error
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to create constraint: %v", err))
		return fmt.Errorf("failed to create constraint: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("CreateConstraint COMPLETE - ID: %s", constraint.ID))
	return nil
}

// UpdateConstraint updates an existing constraint
func (s *PlanSchedulerProfileService) UpdateConstraint(ctx context.Context, constraint *models.PlanSchedulerProfileConstraint, modifiedBy string) error {
	s.iLog.Info(fmt.Sprintf("UpdateConstraint START - ID: %s", constraint.ID))

	// Set audit fields
	constraint.ModifiedBy = modifiedBy
	constraint.ModifiedOn = time.Now()
	constraint.RowVersionStamp++

	// Update constraint
	err := s.gorm.WithContext(ctx).
		Model(&models.PlanSchedulerProfileConstraint{}).
		Where("id = ? AND profile_id = ?", constraint.ID, constraint.ProfileID).
		Updates(constraint).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to update constraint: %v", err))
		return fmt.Errorf("failed to update constraint: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("UpdateConstraint COMPLETE - ID: %s", constraint.ID))
	return nil
}

// DeleteConstraint soft-deletes a constraint (sets active = false)
func (s *PlanSchedulerProfileService) DeleteConstraint(ctx context.Context, profileID string, constraintID string, modifiedBy string) error {
	s.iLog.Info(fmt.Sprintf("DeleteConstraint START - ProfileID: %s, ID: %s", profileID, constraintID))

	err := s.gorm.WithContext(ctx).
		Model(&models.PlanSchedulerProfileConstraint{}).
		Where("id = ? AND profile_id = ?", constraintID, profileID).
		Updates(map[string]interface{}{
			"active":             false,
			"modified_by":        modifiedBy,
			"modified_on":        time.Now(),
			"row_version_stamp":  gorm.Expr("row_version_stamp + 1"),
		}).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to delete constraint: %v", err))
		return fmt.Errorf("failed to delete constraint: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("DeleteConstraint COMPLETE - ID: %s", constraintID))
	return nil
}

// CreateSetting creates a new setting for a profile
func (s *PlanSchedulerProfileService) CreateSetting(ctx context.Context, setting *models.PlanSchedulerProfileSetting, createdBy string) error {
	s.iLog.Info(fmt.Sprintf("CreateSetting START - ProfileID: %s, Key: %s", setting.ProfileID, setting.SettingKey))

	// Generate ID if not provided
	if setting.ID == "" {
		setting.ID = uuid.New().String()
	}

	// Set audit fields
	setting.Active = true
	setting.CreatedBy = createdBy
	setting.CreatedOn = time.Now()
	setting.ModifiedBy = createdBy
	setting.ModifiedOn = time.Now()
	setting.RowVersionStamp = 0

	// Create setting
	err := s.gorm.WithContext(ctx).Create(setting).Error
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to create setting: %v", err))
		return fmt.Errorf("failed to create setting: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("CreateSetting COMPLETE - ID: %s", setting.ID))
	return nil
}

// UpdateSetting updates an existing setting
func (s *PlanSchedulerProfileService) UpdateSetting(ctx context.Context, setting *models.PlanSchedulerProfileSetting, modifiedBy string) error {
	s.iLog.Info(fmt.Sprintf("UpdateSetting START - ID: %s", setting.ID))

	// Set audit fields
	setting.ModifiedBy = modifiedBy
	setting.ModifiedOn = time.Now()
	setting.RowVersionStamp++

	// Update setting
	err := s.gorm.WithContext(ctx).
		Model(&models.PlanSchedulerProfileSetting{}).
		Where("id = ? AND profile_id = ?", setting.ID, setting.ProfileID).
		Updates(setting).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to update setting: %v", err))
		return fmt.Errorf("failed to update setting: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("UpdateSetting COMPLETE - ID: %s", setting.ID))
	return nil
}

// DeleteSetting soft-deletes a setting (sets active = false)
func (s *PlanSchedulerProfileService) DeleteSetting(ctx context.Context, profileID string, settingID string, modifiedBy string) error {
	s.iLog.Info(fmt.Sprintf("DeleteSetting START - ProfileID: %s, ID: %s", profileID, settingID))

	err := s.gorm.WithContext(ctx).
		Model(&models.PlanSchedulerProfileSetting{}).
		Where("id = ? AND profile_id = ?", settingID, profileID).
		Updates(map[string]interface{}{
			"active":             false,
			"modified_by":        modifiedBy,
			"modified_on":        time.Now(),
			"row_version_stamp":  gorm.Expr("row_version_stamp + 1"),
		}).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to delete setting: %v", err))
		return fmt.Errorf("failed to delete setting: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("DeleteSetting COMPLETE - ID: %s", settingID))
	return nil
}

// ====================================================================
// Plan Scheduler Session Methods
// ====================================================================

// SaveScheduleAsSession saves the schedule data as a session
func (s *PlanSchedulerProfileService) SaveScheduleAsSession(ctx context.Context, session *models.PlanSchedulerSession, createdBy string) error {
	s.iLog.Info(fmt.Sprintf("SaveScheduleAsSession START - ProfileID: %s, SessionID: %s", session.ProfileID, session.SessionID))

	// Generate ID if not provided
	if session.ID == "" {
		session.ID = uuid.New().String()
	}

	// Set audit fields
	session.Active = true
	session.CreatedBy = createdBy
	session.CreatedOn = time.Now()
	session.ModifiedBy = createdBy
	session.ModifiedOn = time.Now()
	session.RowVersionStamp = 0

	// If this is set as default, unset other defaults for this profile
	if session.IsDefault {
		err := s.gorm.WithContext(ctx).
			Model(&models.PlanSchedulerSession{}).
			Where("profile_id = ? AND is_default = ?", session.ProfileID, true).
			Update("is_default", false).Error

		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to unset other defaults: %v", err))
			return fmt.Errorf("failed to unset other defaults: %w", err)
		}
	}

	// Save session
	err := s.gorm.WithContext(ctx).Create(session).Error
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to save session: %v", err))
		return fmt.Errorf("failed to save session: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("SaveScheduleAsSession COMPLETE - ID: %s", session.ID))
	return nil
}

// UpdateScheduleSession updates an existing session
func (s *PlanSchedulerProfileService) UpdateScheduleSession(ctx context.Context, session *models.PlanSchedulerSession, modifiedBy string) error {
	s.iLog.Info(fmt.Sprintf("UpdateScheduleSession START - ID: %s", session.ID))

	// Set audit fields
	session.ModifiedBy = modifiedBy
	session.ModifiedOn = time.Now()
	session.RowVersionStamp++

	// If this is set as default, unset other defaults for this profile
	if session.IsDefault {
		err := s.gorm.WithContext(ctx).
			Model(&models.PlanSchedulerSession{}).
			Where("profile_id = ? AND is_default = ? AND id != ?", session.ProfileID, true, session.ID).
			Update("is_default", false).Error

		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to unset other defaults: %v", err))
			return fmt.Errorf("failed to unset other defaults: %w", err)
		}
	}

	// Update session
	err := s.gorm.WithContext(ctx).
		Model(&models.PlanSchedulerSession{}).
		Where("id = ?", session.ID).
		Updates(session).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to update session: %v", err))
		return fmt.Errorf("failed to update session: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("UpdateScheduleSession COMPLETE - ID: %s", session.ID))
	return nil
}

// GetSessionsByProfile gets all sessions for a profile
func (s *PlanSchedulerProfileService) GetSessionsByProfile(ctx context.Context, profileID string) ([]models.PlanSchedulerSession, error) {
	s.iLog.Info(fmt.Sprintf("GetSessionsByProfile START - ProfileID: %s", profileID))

	var sessions []models.PlanSchedulerSession
	err := s.gorm.WithContext(ctx).
		Where("profile_id = ? AND active = ?", profileID, true).
		Order("is_default DESC, createdon DESC").
		Find(&sessions).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get sessions: %v", err))
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("GetSessionsByProfile COMPLETE - Found %d sessions", len(sessions)))
	return sessions, nil
}

// GetSession gets a specific session by ID
func (s *PlanSchedulerProfileService) GetSession(ctx context.Context, sessionID string) (*models.PlanSchedulerSession, error) {
	s.iLog.Info(fmt.Sprintf("GetSession START - ID: %s", sessionID))

	var session models.PlanSchedulerSession
	err := s.gorm.WithContext(ctx).
		Where("id = ?", sessionID).
		First(&session).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get session: %v", err))
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("GetSession COMPLETE - ID: %s", sessionID))
	return &session, nil
}

// DeleteSession soft-deletes a session
func (s *PlanSchedulerProfileService) DeleteSession(ctx context.Context, sessionID, deletedBy string) error {
	s.iLog.Info(fmt.Sprintf("DeleteSession START - ID: %s", sessionID))

	err := s.gorm.WithContext(ctx).
		Model(&models.PlanSchedulerSession{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"active":            false,
			"modified_by":       deletedBy,
			"modified_on":       time.Now(),
			"row_version_stamp": gorm.Expr("row_version_stamp + 1"),
		}).Error

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to delete session: %v", err))
		return fmt.Errorf("failed to delete session: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("DeleteSession COMPLETE - ID: %s", sessionID))
	return nil
}
