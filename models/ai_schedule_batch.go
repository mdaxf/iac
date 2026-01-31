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

// AIScheduleBatchJobStatus represents the status of a batch job
type AIScheduleBatchJobStatus string

const (
	BatchStatusPending    AIScheduleBatchJobStatus = "pending"
	BatchStatusProcessing AIScheduleBatchJobStatus = "processing"
	BatchStatusCompleted  AIScheduleBatchJobStatus = "completed"
	BatchStatusFailed     AIScheduleBatchJobStatus = "failed"
	BatchStatusPartial    AIScheduleBatchJobStatus = "partial" // Some batches succeeded, some failed
)

// AIScheduleBatchJob represents a batch AI schedule optimization job
type AIScheduleBatchJob struct {
	ID               string                   `json:"id" db:"id" gorm:"column:id;type:varchar(100);primaryKey"`
	ParentJobID      string                   `json:"parent_job_id" db:"parent_job_id" gorm:"column:parent_job_id;type:varchar(100)"`
	TotalBatches     int                      `json:"total_batches" db:"total_batches" gorm:"column:total_batches;type:int;default:0"`
	CompletedBatches int                      `json:"completed_batches" db:"completed_batches" gorm:"column:completed_batches;type:int;default:0"`
	FailedBatches    int                      `json:"failed_batches" db:"failed_batches" gorm:"column:failed_batches;type:int;default:0"`
	Status           AIScheduleBatchJobStatus `json:"status" db:"status" gorm:"column:status;type:varchar(20);default:'pending'"`
	Objective        string                   `json:"objective" db:"objective" gorm:"column:objective;type:text"`
	TotalTasks       int                      `json:"total_tasks" db:"total_tasks" gorm:"column:total_tasks;type:int;default:0"`
	ProcessedTasks   int                      `json:"processed_tasks" db:"processed_tasks" gorm:"column:processed_tasks;type:int;default:0"`
	BatchSize        int                      `json:"batch_size" db:"batch_size" gorm:"column:batch_size;type:int;default:30"`
	Progress         int                      `json:"progress" db:"progress" gorm:"column:progress;type:int;default:0"`
	ErrorMessage     string                   `json:"error_message" db:"error_message" gorm:"column:error_message;type:text"`
	StartedAt        *time.Time               `json:"started_at" db:"started_at" gorm:"column:started_at;type:timestamp"`
	CompletedAt      *time.Time               `json:"completed_at" db:"completed_at" gorm:"column:completed_at;type:timestamp"`
	ReferenceID      int                      `json:"referenceid" db:"referenceid" gorm:"column:referenceid;type:int"`
	ModifiedOn       time.Time                `json:"modifiedon" db:"modifiedon" gorm:"column:modifiedon;type:datetime;autoUpdateTime"`
	ModifiedBy       string                   `json:"modifiedby" db:"modifiedby" gorm:"column:modifiedby;type:varchar(50)"`
	CreatedOn        time.Time                `json:"createdon" db:"createdon" gorm:"column:createdon;type:datetime;autoCreateTime"`
	CreatedBy        string                   `json:"createdby" db:"createdby" gorm:"column:createdby;type:varchar(50)"`
	Active           bool                     `json:"active" db:"active" gorm:"column:active;type:boolean;default:1"`
	RowVersionStamp  int                      `json:"rowversionstamp" db:"rowversionstamp" gorm:"column:rowversionstamp;type:int;default:1"`
}

// TableName specifies the table name for GORM
func (AIScheduleBatchJob) TableName() string {
	return "ai_schedule_batch_jobs"
}

// AIScheduleBatchResult represents the result of a single batch
type AIScheduleBatchResult struct {
	ID              string     `json:"id" db:"id" gorm:"column:id;type:varchar(100);primaryKey"`
	BatchJobID      string     `json:"batch_job_id" db:"batch_job_id" gorm:"column:batch_job_id;type:varchar(100);not null"`
	BatchNumber     int        `json:"batch_number" db:"batch_number" gorm:"column:batch_number;type:int;not null"`
	TasksJSON       string     `json:"tasks_json" db:"tasks_json" gorm:"column:tasks_json;type:longtext"`
	ChangesJSON     string     `json:"changes_json" db:"changes_json" gorm:"column:changes_json;type:longtext"`
	StartIndex      int        `json:"start_index" db:"start_index" gorm:"column:start_index;type:int;not null"`
	EndIndex        int        `json:"end_index" db:"end_index" gorm:"column:end_index;type:int;not null"`
	Status          string     `json:"status" db:"status" gorm:"column:status;type:varchar(20);default:'pending'"`
	ErrorMessage    string     `json:"error_message" db:"error_message" gorm:"column:error_message;type:text"`
	ProcessedAt     *time.Time `json:"processed_at" db:"processed_at" gorm:"column:processed_at;type:timestamp"`
	ReferenceID     int        `json:"referenceid" db:"referenceid" gorm:"column:referenceid;type:int"`
	ModifiedOn      time.Time  `json:"modifiedon" db:"modifiedon" gorm:"column:modifiedon;type:datetime;autoUpdateTime"`
	ModifiedBy      string     `json:"modifiedby" db:"modifiedby" gorm:"column:modifiedby;type:varchar(50)"`
	CreatedOn       time.Time  `json:"createdon" db:"createdon" gorm:"column:createdon;type:datetime;autoCreateTime"`
	CreatedBy       string     `json:"createdby" db:"createdby" gorm:"column:createdby;type:varchar(50)"`
	Active          bool       `json:"active" db:"active" gorm:"column:active;type:boolean;default:1"`
	RowVersionStamp int        `json:"rowversionstamp" db:"rowversionstamp" gorm:"column:rowversionstamp;type:int;default:1"`
}

// TableName specifies the table name for GORM
func (AIScheduleBatchResult) TableName() string {
	return "ai_schedule_batch_results"
}
