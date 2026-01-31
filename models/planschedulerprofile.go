// Copyright 2023 IAC. All Rights Reserved.

package models

import (
	"encoding/json"
	"time"
)

// PlanSchedulerProfile represents a plan scheduler profile
type PlanSchedulerProfile struct {
	ID          string    `json:"id" db:"id" gorm:"column:id;primaryKey;type:varchar(36)"`
	Name        string    `json:"name" db:"name" gorm:"column:name;type:varchar(200);not null"`
	Description string    `json:"description,omitempty" db:"description" gorm:"column:description;type:text"`
	IsDefault   bool      `json:"isDefault" db:"is_default" gorm:"column:is_default;default:false"`

	// Standard IAC 7 moderate fields
	Active          bool      `json:"active" db:"active" gorm:"column:active;default:true"`
	ReferenceID     string    `json:"referenceId,omitempty" db:"referenceid" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string    `json:"createdBy,omitempty" db:"createdby" gorm:"column:createdby;type:varchar(100)"`
	CreatedOn       time.Time `json:"createdOn" db:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string    `json:"modifiedBy,omitempty" db:"modifiedby" gorm:"column:modifiedby;type:varchar(100)"`
	ModifiedOn      time.Time `json:"modifiedOn" db:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int64     `json:"rowVersionStamp" db:"rowversionstamp" gorm:"column:rowversionstamp;default:0"`

	// Related data (loaded separately)
	DataSources []PlanSchedulerProfileDataSource  `json:"dataSources,omitempty" gorm:"foreignKey:ProfileID;references:ID"`
	Constraints []PlanSchedulerProfileConstraint  `json:"constraints,omitempty" gorm:"foreignKey:ProfileID;references:ID"`
	Settings    []PlanSchedulerProfileSetting     `json:"settings,omitempty" gorm:"foreignKey:ProfileID;references:ID"`
}

// TableName specifies the table name for GORM
func (PlanSchedulerProfile) TableName() string {
	return "plan_scheduler_profiles"
}

// PlanSchedulerProfileDataSource represents a data source configuration
type PlanSchedulerProfileDataSource struct {
	ID          string          `json:"id" db:"id" gorm:"column:id;primaryKey;type:varchar(36)"`
	ProfileID   string          `json:"profileId" db:"profile_id" gorm:"column:profile_id;type:varchar(36);not null"`
	DataType    string          `json:"dataType" db:"data_type" gorm:"column:data_type;type:varchar(50);not null"` // 'tasks', 'resources', 'materials', 'equipment', 'custom'
	Name        string          `json:"name" db:"name" gorm:"column:name;type:varchar(200);not null"`
	Description string          `json:"description,omitempty" db:"description" gorm:"column:description;type:text"`
	SourceType  string          `json:"sourceType" db:"source_type" gorm:"column:source_type;type:varchar(20);not null"` // 'query' or 'json'
	SourceQuery string          `json:"sourceQuery,omitempty" db:"source_query" gorm:"column:source_query;type:text"`
	SourceJSON  json.RawMessage `json:"sourceJson,omitempty" db:"source_json" gorm:"column:source_json;type:jsonb"`
	Parameters  json.RawMessage `json:"parameters,omitempty" db:"parameters" gorm:"column:parameters;type:jsonb"`
	DisplayOrder int            `json:"displayOrder" db:"display_order" gorm:"column:display_order;default:0"`

	// Standard IAC 7 moderate fields
	Active          bool      `json:"active" db:"active" gorm:"column:active;default:true"`
	ReferenceID     string    `json:"referenceId,omitempty" db:"referenceid" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string    `json:"createdBy,omitempty" db:"createdby" gorm:"column:createdby;type:varchar(100)"`
	CreatedOn       time.Time `json:"createdOn" db:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string    `json:"modifiedBy,omitempty" db:"modifiedby" gorm:"column:modifiedby;type:varchar(100)"`
	ModifiedOn      time.Time `json:"modifiedOn" db:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int64     `json:"rowVersionStamp" db:"rowversionstamp" gorm:"column:rowversionstamp;default:0"`
}

// TableName specifies the table name for GORM
func (PlanSchedulerProfileDataSource) TableName() string {
	return "plan_scheduler_profile_datasources"
}

// PlanSchedulerProfileConstraint represents a scheduling constraint
type PlanSchedulerProfileConstraint struct {
	ID             string          `json:"id" db:"id" gorm:"column:id;primaryKey;type:varchar(36)"`
	ProfileID      string          `json:"profileId" db:"profile_id" gorm:"column:profile_id;type:varchar(36);not null"`
	ConstraintType string          `json:"constraintType" db:"constraint_type" gorm:"column:constraint_type;type:varchar(50);not null"` // 'time', 'resource', 'dependency', 'custom'
	Name           string          `json:"name" db:"name" gorm:"column:name;type:varchar(200);not null"`
	Description    string          `json:"description,omitempty" db:"description" gorm:"column:description;type:text"`
	SourceType     string          `json:"sourceType" db:"source_type" gorm:"column:source_type;type:varchar(20);not null"` // 'query' or 'json'
	SourceQuery    string          `json:"sourceQuery,omitempty" db:"source_query" gorm:"column:source_query;type:text"`
	SourceJSON     json.RawMessage `json:"sourceJson,omitempty" db:"source_json" gorm:"column:source_json;type:jsonb"`
	Enforcement    string          `json:"enforcement" db:"enforcement" gorm:"column:enforcement;type:varchar(20);default:'hard'"` // 'hard', 'soft', 'advisory'
	DisplayOrder   int             `json:"displayOrder" db:"display_order" gorm:"column:display_order;default:0"`

	// Standard IAC 7 moderate fields
	Active          bool      `json:"active" db:"active" gorm:"column:active;default:true"`
	ReferenceID     string    `json:"referenceId,omitempty" db:"referenceid" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string    `json:"createdBy,omitempty" db:"createdby" gorm:"column:createdby;type:varchar(100)"`
	CreatedOn       time.Time `json:"createdOn" db:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string    `json:"modifiedBy,omitempty" db:"modifiedby" gorm:"column:modifiedby;type:varchar(100)"`
	ModifiedOn      time.Time `json:"modifiedOn" db:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int64     `json:"rowVersionStamp" db:"rowversionstamp" gorm:"column:rowversionstamp;default:0"`
}

// TableName specifies the table name for GORM
func (PlanSchedulerProfileConstraint) TableName() string {
	return "plan_scheduler_profile_constraints"
}

// PlanSchedulerProfileSetting represents a profile setting
type PlanSchedulerProfileSetting struct {
	ID           string          `json:"id" db:"id" gorm:"column:id;primaryKey;type:varchar(36)"`
	ProfileID    string          `json:"profileId" db:"profile_id" gorm:"column:profile_id;type:varchar(36);not null"`
	SettingKey   string          `json:"settingKey" db:"setting_key" gorm:"column:setting_key;type:varchar(100);not null"`
	SettingValue json.RawMessage `json:"settingValue,omitempty" db:"setting_value" gorm:"column:setting_value;type:jsonb"`
	Description  string          `json:"description,omitempty" db:"description" gorm:"column:description;type:text"`

	// Standard IAC 7 moderate fields
	Active          bool      `json:"active" db:"active" gorm:"column:active;default:true"`
	ReferenceID     string    `json:"referenceId,omitempty" db:"referenceid" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string    `json:"createdBy,omitempty" db:"createdby" gorm:"column:createdby;type:varchar(100)"`
	CreatedOn       time.Time `json:"createdOn" db:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string    `json:"modifiedBy,omitempty" db:"modifiedby" gorm:"column:modifiedby;type:varchar(100)"`
	ModifiedOn      time.Time `json:"modifiedOn" db:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int64     `json:"rowVersionStamp" db:"rowversionstamp" gorm:"column:rowversionstamp;default:0"`
}

// TableName specifies the table name for GORM
func (PlanSchedulerProfileSetting) TableName() string {
	return "plan_scheduler_profile_settings"
}

// PlanSchedulerProfileData represents the loaded data from a profile
type PlanSchedulerProfileData struct {
	ProfileID   string                 `json:"profileId"`
	ProfileName string                 `json:"profileName"`
	Tasks       []map[string]interface{} `json:"tasks,omitempty"`
	Resources   []map[string]interface{} `json:"resources,omitempty"`
	MasterData  map[string][]map[string]interface{} `json:"masterData,omitempty"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// PlanSchedulerSession represents a saved plan/schedule session
type PlanSchedulerSession struct {
	ID          string          `json:"id" db:"id" gorm:"column:id;primaryKey;type:varchar(36)"`
	ProfileID   string          `json:"profileId" db:"profile_id" gorm:"column:profile_id;type:varchar(36);not null"`
	SessionID   string          `json:"sessionId" db:"session_id" gorm:"column:session_id;type:varchar(100);not null"`
	SessionName string          `json:"sessionName,omitempty" db:"session_name" gorm:"column:session_name;type:varchar(200)"`
	Description string          `json:"description,omitempty" db:"description" gorm:"column:description;type:text"`

	// Saved schedule data
	TasksData     json.RawMessage `json:"tasksData,omitempty" db:"tasks_data" gorm:"column:tasks_data;type:jsonb"`
	ResourcesData json.RawMessage `json:"resourcesData,omitempty" db:"resources_data" gorm:"column:resources_data;type:jsonb"`
	Metadata      json.RawMessage `json:"metadata,omitempty" db:"metadata" gorm:"column:metadata;type:jsonb"`

	// Session status
	Status    string `json:"status" db:"status" gorm:"column:status;type:varchar(20);default:'draft'"` // 'draft', 'active', 'archived'
	IsDefault bool   `json:"isDefault" db:"is_default" gorm:"column:is_default;default:false"`

	// Standard IAC 7 moderate fields
	Active          bool      `json:"active" db:"active" gorm:"column:active;default:true"`
	ReferenceID     string    `json:"referenceId,omitempty" db:"referenceid" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string    `json:"createdBy,omitempty" db:"createdby" gorm:"column:createdby;type:varchar(100)"`
	CreatedOn       time.Time `json:"createdOn" db:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string    `json:"modifiedBy,omitempty" db:"modifiedby" gorm:"column:modifiedby;type:varchar(100)"`
	ModifiedOn      time.Time `json:"modifiedOn" db:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int64     `json:"rowVersionStamp" db:"rowversionstamp" gorm:"column:rowversionstamp;default:0"`
}

// TableName specifies the table name for GORM
func (PlanSchedulerSession) TableName() string {
	return "plan_scheduler_sessions"
}
