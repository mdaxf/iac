-- IAC Background Jobs Schema - PostgreSQL
-- Uses JSONB for better performance
-- Uses IAC standard naming convention: no snake_case, 7 standard fields at end

-- =====================================================
-- Table: backgroundjobs
-- Purpose: Store background job definitions and status
-- =====================================================

CREATE TABLE backgroundjobs (
    id VARCHAR(50) PRIMARY KEY,
    jobtype VARCHAR(50) NOT NULL CHECK (jobtype IN ('package_deployment', 'data_migration', 'scheduled_report', 'cleanup')),
    jobdata JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    scheduledat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    startedat TIMESTAMP,
    completedat TIMESTAMP,
    durationseconds INTEGER,
    errorlog JSONB,
    resultdata JSONB,
    retrycount INTEGER DEFAULT 0,
    maxretries INTEGER DEFAULT 3,
    priority INTEGER DEFAULT 0,
    -- IAC Standard Fields (7 fields)
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_jobs_type ON backgroundjobs(jobtype);
CREATE INDEX idx_jobs_status ON backgroundjobs(status);
CREATE INDEX idx_jobs_scheduled ON backgroundjobs(scheduledat);
CREATE INDEX idx_jobs_createdby ON backgroundjobs(createdby);
CREATE INDEX idx_jobs_active ON backgroundjobs(active);
CREATE INDEX idx_jobs_priority_status ON backgroundjobs(priority DESC, status);
CREATE INDEX idx_jobs_data ON backgroundjobs USING GIN(jobdata);

-- =====================================================
-- Functions and Triggers
-- =====================================================

-- Auto-update modifiedon timestamp and increment rowversionstamp
CREATE OR REPLACE FUNCTION update_backgroundjobs_modifiedon()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modifiedon = CURRENT_TIMESTAMP;
    NEW.rowversionstamp = OLD.rowversionstamp + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_backgroundjobs_timestamp
    BEFORE UPDATE ON backgroundjobs
    FOR EACH ROW
    EXECUTE FUNCTION update_backgroundjobs_modifiedon();

-- Calculate job duration on completion
CREATE OR REPLACE FUNCTION calculate_job_duration()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status IN ('completed', 'failed', 'cancelled') AND NEW.startedat IS NOT NULL AND NEW.completedat IS NOT NULL THEN
        NEW.durationseconds = EXTRACT(EPOCH FROM (NEW.completedat - NEW.startedat))::INTEGER;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER calculate_backgroundjob_duration
    BEFORE UPDATE ON backgroundjobs
    FOR EACH ROW
    EXECUTE FUNCTION calculate_job_duration();
