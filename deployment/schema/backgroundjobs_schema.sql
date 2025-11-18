-- IAC Background Jobs Schema
-- Supports MySQL, PostgreSQL, MSSQL, Oracle
-- Uses IAC standard naming convention: no snake_case, 7 standard fields at end

-- =====================================================
-- Table: backgroundjobs
-- Purpose: Store background job definitions and status
-- =====================================================

CREATE TABLE backgroundjobs (
    id VARCHAR(50) PRIMARY KEY,
    jobtype VARCHAR(50) NOT NULL,  -- 'package_deployment', 'data_migration', etc.
    jobdata JSON NOT NULL,  -- Job-specific data
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- 'pending', 'running', 'completed', 'failed'
    scheduledat DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    startedat DATETIME,
    completedat DATETIME,
    durationseconds INT,
    errorlog JSON,  -- Array of error messages
    resultdata JSON,  -- Job result data
    retrycount INT DEFAULT 0,
    maxretries INT DEFAULT 3,
    priority INT DEFAULT 0,  -- Higher number = higher priority
    -- IAC Standard Fields (7 fields)
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,
    INDEX idx_jobtype (jobtype),
    INDEX idx_status (status),
    INDEX idx_scheduledat (scheduledat),
    INDEX idx_createdby (createdby),
    INDEX idx_active (active),
    INDEX idx_priority_status (priority, status)
);

-- Sample queries

-- Get pending jobs ready to execute
-- SELECT * FROM backgroundjobs
-- WHERE status = 'pending'
--   AND scheduledat <= NOW()
--   AND active = TRUE
-- ORDER BY priority DESC, scheduledat ASC
-- LIMIT 10;

-- Get failed jobs for retry
-- SELECT * FROM backgroundjobs
-- WHERE status = 'failed'
--   AND retrycount < maxretries
--   AND active = TRUE
-- ORDER BY scheduledat ASC
-- LIMIT 10;

-- Get job statistics
-- SELECT jobtype, status, COUNT(*) as count
-- FROM backgroundjobs
-- WHERE active = TRUE
-- GROUP BY jobtype, status;
