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
	"time"
)

// =====================================================
// Integration Hub History Models
// =====================================================

// IntHubHistoryStatus represents the status of an integration hub transaction
type IntHubHistoryStatus string

const (
	IntHubHistoryStatusPending    IntHubHistoryStatus = "pending"
	IntHubHistoryStatusProcessing IntHubHistoryStatus = "processing"
	IntHubHistoryStatusCompleted  IntHubHistoryStatus = "completed"
	IntHubHistoryStatusFailed     IntHubHistoryStatus = "failed"
)

// IntHubActionStatus represents the status of an action within a transaction
type IntHubActionStatus string

const (
	IntHubActionStatusPending    IntHubActionStatus = "pending"
	IntHubActionStatusProcessing IntHubActionStatus = "processing"
	IntHubActionStatusCompleted  IntHubActionStatus = "completed"
	IntHubActionStatusFailed     IntHubActionStatus = "failed"
	IntHubActionStatusSkipped    IntHubActionStatus = "skipped"
)

// IntHubHistory represents a transaction history record for integration hub
type IntHubHistory struct {
	ID                string                 `json:"id" bson:"id" gorm:"column:id;primaryKey"`
	HubID             string                 `json:"hub_id" bson:"hub_id" gorm:"column:hub_id"`
	HubName           string                 `json:"hub_name" bson:"hub_name" gorm:"column:hub_name"`
	Direction         string                 `json:"direction" bson:"direction" gorm:"column:direction"`
	Protocol          string                 `json:"protocol" bson:"protocol" gorm:"column:protocol"`
	ProtocolGroupID   string                 `json:"protocol_group_id" bson:"protocol_group_id" gorm:"column:protocol_group_id"`
	ProtocolGroupName string                 `json:"protocol_group_name" bson:"protocol_group_name" gorm:"column:protocol_group_name"`
	EndpointID        string                 `json:"endpoint_id" bson:"endpoint_id" gorm:"column:endpoint_id"`
	EndpointName      string                 `json:"endpoint_name" bson:"endpoint_name" gorm:"column:endpoint_name"`
	Topic             string                 `json:"topic" bson:"topic" gorm:"column:topic"`
	Path              string                 `json:"path" bson:"path" gorm:"column:path"`
	Method            string                 `json:"method" bson:"method" gorm:"column:method"`
	Source            string                 `json:"source" bson:"source" gorm:"column:source"`
	Payload           string                 `json:"payload" bson:"payload" gorm:"column:payload"`
	PayloadSize       int                    `json:"payload_size" bson:"payload_size" gorm:"column:payload_size"`
	MessageType       string                 `json:"message_type" bson:"message_type" gorm:"column:message_type"`
	Action            string                 `json:"action" bson:"action" gorm:"column:action"`
	Status            IntHubHistoryStatus    `json:"status" bson:"status" gorm:"column:status"`
	ErrorMessage      string                 `json:"error_message" bson:"error_message" gorm:"column:error_message"`
	MappedData        string                 `json:"mapped_data" bson:"mapped_data" gorm:"column:mapped_data"`
	Response          string                 `json:"response" bson:"response" gorm:"column:response"`
	ResponseStatus    int                    `json:"response_status" bson:"response_status" gorm:"column:response_status"`
	StartTime         time.Time              `json:"start_time" bson:"start_time" gorm:"column:start_time"`
	EndTime           *time.Time             `json:"end_time" bson:"end_time" gorm:"column:end_time"`
	DurationMs        int64                  `json:"duration_ms" bson:"duration_ms" gorm:"column:duration_ms"`
	InstanceID        string                 `json:"instance_id" bson:"instance_id" gorm:"column:instance_id"`
	InstanceName      string                 `json:"instance_name" bson:"instance_name" gorm:"column:instance_name"`
	UserID            string                 `json:"user_id" bson:"user_id" gorm:"column:user_id"`
	ClientID          string                 `json:"client_id" bson:"client_id" gorm:"column:client_id"`
	Metadata          map[string]interface{} `json:"metadata" bson:"metadata" gorm:"column:metadata;serializer:json"`

	// IAC Standard Fields
	Active          bool      `json:"active" bson:"active" gorm:"column:active;default:true"`
	ReferenceID     string    `json:"referenceid" bson:"referenceid" gorm:"column:referenceid"`
	CreatedBy       string    `json:"createdby" bson:"createdby" gorm:"column:createdby"`
	CreatedOn       time.Time `json:"createdon" bson:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" bson:"modifiedby" gorm:"column:modifiedby"`
	ModifiedOn      time.Time `json:"modifiedon" bson:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" bson:"rowversionstamp" gorm:"column:rowversionstamp;default:1"`

	// Related actions (not stored in DB, populated on demand)
	Actions []IntHubActionHistory `json:"actions,omitempty" bson:"-" gorm:"-"`
}

// TableName returns the table name for GORM
func (IntHubHistory) TableName() string {
	return "iac_int_hub_history"
}

// IntHubActionHistory represents an individual action within a transaction
type IntHubActionHistory struct {
	ID             string                 `json:"id" bson:"id" gorm:"column:id;primaryKey"`
	IntHubHisID    string                 `json:"int_hub_his_id" bson:"int_hub_his_id" gorm:"column:int_hub_his_id"`
	ActionSequence int                    `json:"action_sequence" bson:"action_sequence" gorm:"column:action_sequence"`
	ActionType     string                 `json:"action_type" bson:"action_type" gorm:"column:action_type"`
	ActionName     string                 `json:"action_name" bson:"action_name" gorm:"column:action_name"`
	ActionConfig   string                 `json:"action_config" bson:"action_config" gorm:"column:action_config"`
	InputData      string                 `json:"input_data" bson:"input_data" gorm:"column:input_data"`
	OutputData     string                 `json:"output_data" bson:"output_data" gorm:"column:output_data"`
	Status         IntHubActionStatus     `json:"status" bson:"status" gorm:"column:status"`
	ErrorMessage   string                 `json:"error_message" bson:"error_message" gorm:"column:error_message"`
	StartTime      time.Time              `json:"start_time" bson:"start_time" gorm:"column:start_time"`
	EndTime        *time.Time             `json:"end_time" bson:"end_time" gorm:"column:end_time"`
	DurationMs     int64                  `json:"duration_ms" bson:"duration_ms" gorm:"column:duration_ms"`
	Metadata       map[string]interface{} `json:"metadata" bson:"metadata" gorm:"column:metadata;serializer:json"`

	// IAC Standard Fields
	Active          bool      `json:"active" bson:"active" gorm:"column:active;default:true"`
	ReferenceID     string    `json:"referenceid" bson:"referenceid" gorm:"column:referenceid"`
	CreatedBy       string    `json:"createdby" bson:"createdby" gorm:"column:createdby"`
	CreatedOn       time.Time `json:"createdon" bson:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" bson:"modifiedby" gorm:"column:modifiedby"`
	ModifiedOn      time.Time `json:"modifiedon" bson:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" bson:"rowversionstamp" gorm:"column:rowversionstamp;default:1"`
}

// TableName returns the table name for GORM
func (IntHubActionHistory) TableName() string {
	return "iac_int_hub_action_history"
}

// IntHubHistoryListItem represents a simplified history item for list views
type IntHubHistoryListItem struct {
	ID           string              `json:"id"`
	HubID        string              `json:"hub_id"`
	HubName      string              `json:"hub_name"`
	Direction    string              `json:"direction"`
	Protocol     string              `json:"protocol"`
	EndpointName string              `json:"endpoint_name"`
	Topic        string              `json:"topic"`
	Status       IntHubHistoryStatus `json:"status"`
	DurationMs   int64               `json:"duration_ms"`
	StartTime    time.Time           `json:"start_time"`
}

// IntHubHistoryStats represents statistics for integration hub history
type IntHubHistoryStats struct {
	TotalTransactions int64                 `json:"total_transactions"`
	CompletedCount    int64                 `json:"completed_count"`
	FailedCount       int64                 `json:"failed_count"`
	PendingCount      int64                 `json:"pending_count"`
	ProcessingCount   int64                 `json:"processing_count"`
	AverageDurationMs float64               `json:"average_duration_ms"`
	MinDurationMs     int64                 `json:"min_duration_ms"`
	MaxDurationMs     int64                 `json:"max_duration_ms"`
	ByProtocol        map[string]int64      `json:"by_protocol"`
	ByDirection       map[string]int64      `json:"by_direction"`
	ByStatus          map[string]int64      `json:"by_status"`
	TopEndpoints      []IntHubEndpointStat  `json:"top_endpoints"`
	TopTopics         []IntHubTopicStat     `json:"top_topics"`
	TimeRange         *IntHubTimeRangeStats `json:"time_range"`
}

// IntHubEndpointStat represents statistics for a specific endpoint in integration hub
type IntHubEndpointStat struct {
	EndpointID   string  `json:"endpoint_id"`
	EndpointName string  `json:"endpoint_name"`
	Protocol     string  `json:"protocol"`
	TotalCount   int64   `json:"total_count"`
	AvgDuration  float64 `json:"avg_duration"`
	ErrorRate    float64 `json:"error_rate"`
}

// IntHubTopicStat represents statistics for a specific topic in integration hub
type IntHubTopicStat struct {
	Topic       string  `json:"topic"`
	Protocol    string  `json:"protocol"`
	TotalCount  int64   `json:"total_count"`
	AvgDuration float64 `json:"avg_duration"`
	ErrorRate   float64 `json:"error_rate"`
}

// IntHubTimeRangeStats represents time range for integration hub statistics
type IntHubTimeRangeStats struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// ListIntHubHistoryParams represents parameters for listing history
type ListIntHubHistoryParams struct {
	HubID      string              `json:"hub_id" form:"hub_id"`
	Direction  string              `json:"direction" form:"direction"`
	Protocol   string              `json:"protocol" form:"protocol"`
	EndpointID string              `json:"endpoint_id" form:"endpoint_id"`
	Topic      string              `json:"topic" form:"topic"`
	Status     IntHubHistoryStatus `json:"status" form:"status"`
	StartTime  time.Time           `json:"start_time" form:"start_time"`
	EndTime    time.Time           `json:"end_time" form:"end_time"`
	OnlyErrors bool                `json:"only_errors" form:"only_errors"`
	Page       int                 `json:"page" form:"page"`
	PageSize   int                 `json:"page_size" form:"page_size"`
}

// ListIntHubHistoryResponse represents the response for listing history
type ListIntHubHistoryResponse struct {
	History  []IntHubHistoryListItem `json:"history"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
	Pages    int                     `json:"pages"`
}

// CleanupIntHubHistoryRequest represents a cleanup request
type CleanupIntHubHistoryRequest struct {
	OlderThanDays int `json:"older_than_days" binding:"required,min=1"`
}

// CleanupIntHubHistoryResponse represents a cleanup response
type CleanupIntHubHistoryResponse struct {
	Message string    `json:"message"`
	Deleted int64     `json:"deleted"`
	Cutoff  time.Time `json:"cutoff"`
}
