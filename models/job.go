package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// JobStatus represents the status of a job
type JobStatus int

const (
	JobStatusPending JobStatus = iota
	JobStatusQueued
	JobStatusProcessing
	JobStatusCompleted
	JobStatusFailed
	JobStatusRetrying
	JobStatusCancelled
	JobStatusScheduled
)

// JobType represents the type of job
type JobType int

const (
	JobTypeIntegration JobType = iota
	JobTypeScheduled
	JobTypeManual
	JobTypeSystem
)

// JobDirection represents the direction of message flow
type JobDirection string

const (
	JobDirectionInbound  JobDirection = "inbound"
	JobDirectionOutbound JobDirection = "outbound"
	JobDirectionInternal JobDirection = "internal"
)

// JobMetadata stores flexible metadata for jobs
type JobMetadata map[string]interface{}

// Scan implements sql.Scanner interface
func (j *JobMetadata) Scan(value interface{}) error {
	if value == nil {
		*j = make(JobMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	if len(bytes) == 0 {
		*j = make(JobMetadata)
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// Value implements driver.Valuer interface
func (j JobMetadata) Value() (driver.Value, error) {
	if j == nil {
		return json.Marshal(make(map[string]interface{}))
	}
	return json.Marshal(j)
}

// QueueJob represents a job in the queue to be executed
type QueueJob struct {
	ID              string       `json:"id" db:"id"`
	TypeID          int          `json:"typeid" db:"typeid"`
	Method          string       `json:"method" db:"method"`
	Protocol        string       `json:"protocol" db:"protocol"`
	Direction       JobDirection `json:"direction" db:"direction"`
	Handler         string       `json:"handler" db:"handler"`          // Transaction code or command to execute
	Metadata        JobMetadata  `json:"metadata" db:"metadata"`        // Flexible metadata storage
	Payload         string       `json:"payload" db:"payload"`          // Job payload/data
	Result          string       `json:"result" db:"result"`            // Execution result
	StatusID        int          `json:"statusid" db:"statusid"`        // Job status
	Priority        int          `json:"priority" db:"priority"`        // Higher number = higher priority
	MaxRetries      int          `json:"maxretries" db:"maxretries"`    // Maximum retry attempts
	RetryCount      int          `json:"retrycount" db:"retrycount"`    // Current retry count
	ScheduledAt     *time.Time   `json:"scheduledat" db:"scheduledat"`  // When to execute (null = immediate)
	StartedAt       *time.Time   `json:"startedat" db:"startedat"`      // When execution started
	CompletedAt     *time.Time   `json:"completedat" db:"completedat"`  // When execution completed
	LastError       string       `json:"lasterror" db:"lasterror"`      // Last error message
	ParentJobID     string       `json:"parentjobid" db:"parentjobid"`  // Parent job for chained jobs
	Active          bool         `json:"active" db:"active"`
	ReferenceID     string       `json:"referenceid" db:"referenceid"`
	CreatedBy       string       `json:"createdby" db:"createdby"`
	CreatedOn       time.Time    `json:"createdon" db:"createdon"`
	ModifiedBy      string       `json:"modifiedby" db:"modifiedby"`
	ModifiedOn      time.Time    `json:"modifiedon" db:"modifiedon"`
	RowVersionStamp int          `json:"rowversionstamp" db:"rowversionstamp"`
}

// JobHistory represents the execution history of jobs
type JobHistory struct {
	ID              string      `json:"id" db:"id"`
	JobID           string      `json:"jobid" db:"jobid"`               // Reference to QueueJob
	ExecutionID     string      `json:"executionid" db:"executionid"`   // Unique execution identifier
	StatusID        int         `json:"statusid" db:"statusid"`         // Execution status
	StartedAt       time.Time   `json:"startedat" db:"startedat"`       // Execution start time
	CompletedAt     *time.Time  `json:"completedat" db:"completedat"`   // Execution completion time
	Duration        int64       `json:"duration" db:"duration"`         // Duration in milliseconds
	Result          string      `json:"result" db:"result"`             // Execution result
	ErrorMessage    string      `json:"errormessage" db:"errormessage"` // Error details if failed
	RetryAttempt    int         `json:"retryattempt" db:"retryattempt"` // Which retry attempt this was
	ExecutedBy      string      `json:"executedby" db:"executedby"`     // Worker/instance that executed
	InputData       string      `json:"inputdata" db:"inputdata"`       // Input data snapshot
	OutputData      string      `json:"outputdata" db:"outputdata"`     // Output data
	Metadata        JobMetadata `json:"metadata" db:"metadata"`         // Execution metadata
	Active          bool        `json:"active" db:"active"`
	ReferenceID     string      `json:"referenceid" db:"referenceid"`
	CreatedBy       string      `json:"createdby" db:"createdby"`
	CreatedOn       time.Time   `json:"createdon" db:"createdon"`
	ModifiedBy      string      `json:"modifiedby" db:"modifiedby"`
	ModifiedOn      time.Time   `json:"modifiedon" db:"modifiedon"`
	RowVersionStamp int         `json:"rowversionstamp" db:"rowversionstamp"`
}

// Job represents a scheduled/interval job configuration
type Job struct {
	ID              string      `json:"id" db:"id"`
	Name            string      `json:"name" db:"name"`                   // Job name
	Description     string      `json:"description" db:"description"`     // Job description
	TypeID          int         `json:"typeid" db:"typeid"`               // Job type
	Handler         string      `json:"handler" db:"handler"`             // Transaction code or command to execute
	CronExpression  string      `json:"cronexpression" db:"cronexpression"` // Cron expression for scheduling
	IntervalSeconds int         `json:"intervalseconds" db:"intervalseconds"` // Interval in seconds (alternative to cron)
	StartAt         *time.Time  `json:"startat" db:"startat"`             // When to start execution
	EndAt           *time.Time  `json:"endat" db:"endat"`                 // When to stop execution
	MaxExecutions   int         `json:"maxexecutions" db:"maxexecutions"` // Maximum number of executions (0 = unlimited)
	ExecutionCount  int         `json:"executioncount" db:"executioncount"` // Current execution count
	Enabled         bool        `json:"enabled" db:"enabled"`             // Whether job is enabled
	Condition       string      `json:"condition" db:"condition"`         // SQL or expression to evaluate before execution
	Priority        int         `json:"priority" db:"priority"`           // Job priority
	MaxRetries      int         `json:"maxretries" db:"maxretries"`       // Maximum retry attempts for generated jobs
	Timeout         int         `json:"timeout" db:"timeout"`             // Timeout in seconds
	Metadata        JobMetadata `json:"metadata" db:"metadata"`           // Additional configuration
	LastRunAt       *time.Time  `json:"lastRunAt" db:"lastRunAt"`         // Last execution time
	NextRunAt       *time.Time  `json:"nextRunAt" db:"nextRunAt"`         // Next scheduled execution time
	Active          bool        `json:"active" db:"active"`
	ReferenceID     string      `json:"referenceid" db:"referenceid"`
	CreatedBy       string      `json:"createdby" db:"createdby"`
	CreatedOn       time.Time   `json:"createdon" db:"createdon"`
	ModifiedBy      string      `json:"modifiedby" db:"modifiedby"`
	ModifiedOn      time.Time   `json:"modifiedon" db:"modifiedon"`
	RowVersionStamp int         `json:"rowversionstamp" db:"rowversionstamp"`
}

// JobLock represents a distributed lock for job processing
type JobLock struct {
	JobID      string    `json:"jobid"`
	InstanceID string    `json:"instanceid"`
	LockedAt   time.Time `json:"lockedat"`
	ExpiresAt  time.Time `json:"expiresat"`
}
