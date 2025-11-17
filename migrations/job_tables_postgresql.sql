-- PostgreSQL Migration Script for Job System Tables
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
    metadata JSONB,
    payload TEXT,
    result TEXT,
    statusid INT NOT NULL DEFAULT 0,
    priority INT NOT NULL DEFAULT 0,
    maxretries INT NOT NULL DEFAULT 3,
    retrycount INT NOT NULL DEFAULT 0,
    scheduledat TIMESTAMP NULL,
    startedat TIMESTAMP NULL,
    completedat TIMESTAMP NULL,
    lasterror TEXT,
    parentjobid VARCHAR(255),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1
);

-- Indexes for queue_jobs
CREATE INDEX IF NOT EXISTS idx_queue_jobs_statusid ON queue_jobs(statusid);
CREATE INDEX IF NOT EXISTS idx_queue_jobs_scheduledat ON queue_jobs(scheduledat);
CREATE INDEX IF NOT EXISTS idx_queue_jobs_priority ON queue_jobs(priority DESC);
CREATE INDEX IF NOT EXISTS idx_queue_jobs_createdon ON queue_jobs(createdon);
CREATE INDEX IF NOT EXISTS idx_queue_jobs_direction ON queue_jobs(direction);
CREATE INDEX IF NOT EXISTS idx_queue_jobs_parentjobid ON queue_jobs(parentjobid);
CREATE INDEX IF NOT EXISTS idx_queue_jobs_active_status ON queue_jobs(active, statusid);
CREATE INDEX IF NOT EXISTS idx_queue_jobs_metadata ON queue_jobs USING GIN(metadata);

-- Table: job_histories
-- Stores execution history for all job runs
CREATE TABLE IF NOT EXISTS job_histories (
    id VARCHAR(255) PRIMARY KEY,
    jobid VARCHAR(255) NOT NULL,
    executionid VARCHAR(255) NOT NULL,
    statusid INT NOT NULL,
    startedat TIMESTAMP NOT NULL,
    completedat TIMESTAMP NULL,
    duration BIGINT,
    result TEXT,
    errormessage TEXT,
    retryattempt INT NOT NULL DEFAULT 0,
    executedby VARCHAR(255),
    inputdata TEXT,
    outputdata TEXT,
    metadata JSONB,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,
    FOREIGN KEY (jobid) REFERENCES queue_jobs(id) ON DELETE CASCADE
);

-- Indexes for job_histories
CREATE INDEX IF NOT EXISTS idx_job_histories_jobid ON job_histories(jobid);
CREATE INDEX IF NOT EXISTS idx_job_histories_executionid ON job_histories(executionid);
CREATE INDEX IF NOT EXISTS idx_job_histories_statusid ON job_histories(statusid);
CREATE INDEX IF NOT EXISTS idx_job_histories_startedat ON job_histories(startedat);
CREATE INDEX IF NOT EXISTS idx_job_histories_createdon ON job_histories(createdon);
CREATE INDEX IF NOT EXISTS idx_job_histories_metadata ON job_histories USING GIN(metadata);

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
    startat TIMESTAMP NULL,
    endat TIMESTAMP NULL,
    maxexecutions INT NOT NULL DEFAULT 0,
    executioncount INT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    condition TEXT,
    priority INT NOT NULL DEFAULT 0,
    maxretries INT NOT NULL DEFAULT 3,
    timeout INT NOT NULL DEFAULT 300,
    metadata JSONB,
    lastrunat TIMESTAMP NULL,
    nextrunat TIMESTAMP NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1
);

-- Indexes for jobs
CREATE INDEX IF NOT EXISTS idx_jobs_enabled ON jobs(enabled);
CREATE INDEX IF NOT EXISTS idx_jobs_nextrunat ON jobs(nextrunat);
CREATE INDEX IF NOT EXISTS idx_jobs_name ON jobs(name);
CREATE INDEX IF NOT EXISTS idx_jobs_active_enabled ON jobs(active, enabled);
CREATE INDEX IF NOT EXISTS idx_jobs_metadata ON jobs USING GIN(metadata);

-- Trigger to update modifiedon timestamp
CREATE OR REPLACE FUNCTION update_modifiedon_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modifiedon = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_queue_jobs_modifiedon BEFORE UPDATE ON queue_jobs
    FOR EACH ROW EXECUTE FUNCTION update_modifiedon_column();

CREATE TRIGGER update_job_histories_modifiedon BEFORE UPDATE ON job_histories
    FOR EACH ROW EXECUTE FUNCTION update_modifiedon_column();

CREATE TRIGGER update_jobs_modifiedon BEFORE UPDATE ON jobs
    FOR EACH ROW EXECUTE FUNCTION update_modifiedon_column();

-- Insert sample scheduled jobs
INSERT INTO jobs (id, name, description, typeid, handler, cronexpression, enabled, priority, createdby, createdon)
VALUES
    ('job-cleanup-old-histories', 'Cleanup Old Job Histories', 'Remove job histories older than 90 days', 3, 'system.cleanup.histories', '0 2 * * *', TRUE, 1, 'system', CURRENT_TIMESTAMP),
    ('job-retry-failed-jobs', 'Retry Failed Jobs', 'Retry jobs that failed with retriable errors', 3, 'system.retry.failed', '*/15 * * * *', TRUE, 5, 'system', CURRENT_TIMESTAMP)
ON CONFLICT (id) DO UPDATE SET modifiedon = CURRENT_TIMESTAMP;
