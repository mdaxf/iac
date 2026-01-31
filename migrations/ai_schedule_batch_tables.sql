-- AI Schedule Batch Processing Tables
-- Copyright 2023 IAC. All Rights Reserved.

-- Table for tracking batch AI schedule optimization jobs
CREATE TABLE IF NOT EXISTS ai_schedule_batch_jobs (
    id VARCHAR(100) PRIMARY KEY COMMENT 'Unique identifier for the batch job',
    parent_job_id VARCHAR(100) NULL COMMENT 'Parent job ID for hierarchical jobs',
    total_batches INT NOT NULL DEFAULT 0 COMMENT 'Total number of batches in this job',
    completed_batches INT NOT NULL DEFAULT 0 COMMENT 'Number of batches completed',
    failed_batches INT NOT NULL DEFAULT 0 COMMENT 'Number of batches that failed',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'Job status: pending, processing, completed, partial, failed',
    objective TEXT COMMENT 'Optimization objective description',
    total_tasks INT NOT NULL DEFAULT 0 COMMENT 'Total number of tasks to process',
    processed_tasks INT NOT NULL DEFAULT 0 COMMENT 'Number of tasks processed',
    batch_size INT NOT NULL DEFAULT 30 COMMENT 'Number of tasks per batch',
    progress INT NOT NULL DEFAULT 0 COMMENT 'Progress percentage (0-100)',
    error_message TEXT COMMENT 'Error message if job failed',
    started_at TIMESTAMP NULL COMMENT 'When the job started processing',
    completed_at TIMESTAMP NULL COMMENT 'When the job completed',
    referenceid INT NULL COMMENT 'Reserved',
    modifiedon DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'The date when the record was last updated',
    modifiedby VARCHAR(50) NULL COMMENT 'The user which last updated the record',
    createdon DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'The date the record was created',
    createdby VARCHAR(50) NULL COMMENT 'The user who created the record',
    active BOOLEAN NOT NULL DEFAULT 1 COMMENT 'Indicates if the record is active or deleted',
    rowversionstamp INT DEFAULT 1 COMMENT 'Current version identifier for the row, for detection of concurrency violations',

    INDEX idx_status (status),
    INDEX idx_parent_job (parent_job_id),
    INDEX idx_createdon (createdon),
    INDEX idx_active_status (active, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Tracks batch AI schedule optimization jobs';

-- Table for storing individual batch results
CREATE TABLE IF NOT EXISTS ai_schedule_batch_results (
    id VARCHAR(100) PRIMARY KEY COMMENT 'Unique identifier for the batch result',
    batch_job_id VARCHAR(100) NOT NULL COMMENT 'Reference to parent batch job',
    batch_number INT NOT NULL COMMENT 'Batch sequence number',
    tasks_json LONGTEXT COMMENT 'JSON array of optimized tasks for this batch',
    changes_json LONGTEXT COMMENT 'JSON array of changes made in this batch',
    start_index INT NOT NULL COMMENT 'Starting index in original task list',
    end_index INT NOT NULL COMMENT 'Ending index in original task list',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'Batch status: pending, processing, completed, failed',
    error_message TEXT COMMENT 'Error message if batch failed',
    processed_at TIMESTAMP NULL COMMENT 'When the batch was processed',
    referenceid INT NULL COMMENT 'Reserved',
    modifiedon DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'The date when the record was last updated',
    modifiedby VARCHAR(50) NULL COMMENT 'The user which last updated the record',
    createdon DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'The date the record was created',
    createdby VARCHAR(50) NULL COMMENT 'The user who created the record',
    active BOOLEAN NOT NULL DEFAULT 1 COMMENT 'Indicates if the record is active or deleted',
    rowversionstamp INT DEFAULT 1 COMMENT 'Current version identifier for the row, for detection of concurrency violations',

    FOREIGN KEY (batch_job_id) REFERENCES ai_schedule_batch_jobs(id) ON DELETE CASCADE,
    INDEX idx_batch_job (batch_job_id),
    INDEX idx_batch_number (batch_job_id, batch_number),
    INDEX idx_status (status),
    UNIQUE KEY uk_batch_job_number (batch_job_id, batch_number)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Stores individual batch results for AI schedule optimization';
