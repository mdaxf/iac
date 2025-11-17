-- MySQL Migration Script for Job System Tables
-- Execute this script to create the job management tables

-- Table: queue_jobs
-- Stores all jobs to be executed (from integrations or scheduled jobs)
CREATE TABLE IF NOT EXISTS queue_jobs (
    id VARCHAR(255) PRIMARY KEY,
    typeid INT NOT NULL DEFAULT 0,
    method VARCHAR(100),
    protocol VARCHAR(50),
    direction VARCHAR(20),
    handler VARCHAR(255) NOT NULL,
    metadata JSON,
    payload LONGTEXT,
    result LONGTEXT,
    statusid INT NOT NULL DEFAULT 0,
    priority INT NOT NULL DEFAULT 0,
    maxretries INT NOT NULL DEFAULT 3,
    retrycount INT NOT NULL DEFAULT 0,
    scheduledat DATETIME NULL,
    startedat DATETIME NULL,
    completedat DATETIME NULL,
    lasterror TEXT,
    parentjobid VARCHAR(255),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,
    INDEX idx_statusid (statusid),
    INDEX idx_scheduledat (scheduledat),
    INDEX idx_priority (priority DESC),
    INDEX idx_createdon (createdon),
    INDEX idx_direction (direction),
    INDEX idx_parentjobid (parentjobid),
    INDEX idx_active_status (active, statusid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: job_histories
-- Stores execution history for all job runs
CREATE TABLE IF NOT EXISTS job_histories (
    id VARCHAR(255) PRIMARY KEY,
    jobid VARCHAR(255) NOT NULL,
    executionid VARCHAR(255) NOT NULL,
    statusid INT NOT NULL,
    startedat DATETIME NOT NULL,
    completedat DATETIME NULL,
    duration BIGINT,
    result LONGTEXT,
    errormessage TEXT,
    retryattempt INT NOT NULL DEFAULT 0,
    executedby VARCHAR(255),
    inputdata LONGTEXT,
    outputdata LONGTEXT,
    metadata JSON,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,
    INDEX idx_jobid (jobid),
    INDEX idx_executionid (executionid),
    INDEX idx_statusid (statusid),
    INDEX idx_startedat (startedat),
    INDEX idx_createdon (createdon),
    FOREIGN KEY (jobid) REFERENCES queue_jobs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: jobs
-- Stores scheduled/interval job configurations
CREATE TABLE IF NOT EXISTS jobs (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    typeid INT NOT NULL DEFAULT 0,
    handler VARCHAR(255) NOT NULL,
    cronexpression VARCHAR(100),
    intervalseconds INT,
    startat DATETIME NULL,
    endat DATETIME NULL,
    maxexecutions INT NOT NULL DEFAULT 0,
    executioncount INT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    `condition` TEXT,
    priority INT NOT NULL DEFAULT 0,
    maxretries INT NOT NULL DEFAULT 3,
    timeout INT NOT NULL DEFAULT 300,
    metadata JSON,
    lastrunat DATETIME NULL,
    nextrunat DATETIME NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,
    INDEX idx_enabled (enabled),
    INDEX idx_nextrunat (nextrunat),
    INDEX idx_name (name),
    INDEX idx_active_enabled (active, enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample scheduled jobs
INSERT INTO jobs (id, name, description, typeid, handler, cronexpression, enabled, priority, createdby, createdon)
VALUES
    ('job-cleanup-old-histories', 'Cleanup Old Job Histories', 'Remove job histories older than 90 days', 3, 'system.cleanup.histories', '0 2 * * *', TRUE, 1, 'system', NOW()),
    ('job-retry-failed-jobs', 'Retry Failed Jobs', 'Retry jobs that failed with retriable errors', 3, 'system.retry.failed', '*/15 * * * *', TRUE, 5, 'system', NOW())
ON DUPLICATE KEY UPDATE modifiedon = NOW();
