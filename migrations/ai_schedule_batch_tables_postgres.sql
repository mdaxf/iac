-- AI Schedule Batch Processing Tables - PostgreSQL
-- Copyright 2023 IAC. All Rights Reserved.

-- Table for tracking batch AI schedule optimization jobs
CREATE TABLE IF NOT EXISTS ai_schedule_batch_jobs (
    id VARCHAR(100) PRIMARY KEY,
    parent_job_id VARCHAR(100) NULL,
    total_batches INT NOT NULL DEFAULT 0,
    completed_batches INT NOT NULL DEFAULT 0,
    failed_batches INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    objective TEXT,
    total_tasks INT NOT NULL DEFAULT 0,
    processed_tasks INT NOT NULL DEFAULT 0,
    batch_size INT NOT NULL DEFAULT 30,
    progress INT NOT NULL DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    referenceid INT NULL,
    modifiedon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(50) NULL,
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    createdby VARCHAR(50) NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    rowversionstamp INT DEFAULT 1
);

-- Add comments for ai_schedule_batch_jobs columns
COMMENT ON TABLE ai_schedule_batch_jobs IS 'Tracks batch AI schedule optimization jobs';
COMMENT ON COLUMN ai_schedule_batch_jobs.id IS 'Unique identifier for the batch job';
COMMENT ON COLUMN ai_schedule_batch_jobs.parent_job_id IS 'Parent job ID for hierarchical jobs';
COMMENT ON COLUMN ai_schedule_batch_jobs.total_batches IS 'Total number of batches in this job';
COMMENT ON COLUMN ai_schedule_batch_jobs.completed_batches IS 'Number of batches completed';
COMMENT ON COLUMN ai_schedule_batch_jobs.failed_batches IS 'Number of batches that failed';
COMMENT ON COLUMN ai_schedule_batch_jobs.status IS 'Job status: pending, processing, completed, partial, failed';
COMMENT ON COLUMN ai_schedule_batch_jobs.objective IS 'Optimization objective description';
COMMENT ON COLUMN ai_schedule_batch_jobs.total_tasks IS 'Total number of tasks to process';
COMMENT ON COLUMN ai_schedule_batch_jobs.processed_tasks IS 'Number of tasks processed';
COMMENT ON COLUMN ai_schedule_batch_jobs.batch_size IS 'Number of tasks per batch';
COMMENT ON COLUMN ai_schedule_batch_jobs.progress IS 'Progress percentage (0-100)';
COMMENT ON COLUMN ai_schedule_batch_jobs.error_message IS 'Error message if job failed';
COMMENT ON COLUMN ai_schedule_batch_jobs.started_at IS 'When the job started processing';
COMMENT ON COLUMN ai_schedule_batch_jobs.completed_at IS 'When the job completed';
COMMENT ON COLUMN ai_schedule_batch_jobs.referenceid IS 'Reserved';
COMMENT ON COLUMN ai_schedule_batch_jobs.modifiedon IS 'The date when the record was last updated';
COMMENT ON COLUMN ai_schedule_batch_jobs.modifiedby IS 'The user which last updated the record';
COMMENT ON COLUMN ai_schedule_batch_jobs.createdon IS 'The date the record was created';
COMMENT ON COLUMN ai_schedule_batch_jobs.createdby IS 'The user who created the record';
COMMENT ON COLUMN ai_schedule_batch_jobs.active IS 'Indicates if the record is active or deleted';
COMMENT ON COLUMN ai_schedule_batch_jobs.rowversionstamp IS 'Current version identifier for the row, for detection of concurrency violations';

-- Create indexes for ai_schedule_batch_jobs
CREATE INDEX IF NOT EXISTS idx_batch_jobs_status ON ai_schedule_batch_jobs(status);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_parent_job ON ai_schedule_batch_jobs(parent_job_id);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_createdon ON ai_schedule_batch_jobs(createdon);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_active_status ON ai_schedule_batch_jobs(active, status);

-- Create trigger function to update modifiedon timestamp
CREATE OR REPLACE FUNCTION update_modifiedon_batch_jobs()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modifiedon = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for ai_schedule_batch_jobs
DROP TRIGGER IF EXISTS trigger_update_modifiedon_batch_jobs ON ai_schedule_batch_jobs;
CREATE TRIGGER trigger_update_modifiedon_batch_jobs
    BEFORE UPDATE ON ai_schedule_batch_jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_modifiedon_batch_jobs();

-- Table for storing individual batch results
CREATE TABLE IF NOT EXISTS ai_schedule_batch_results (
    id VARCHAR(100) PRIMARY KEY,
    batch_job_id VARCHAR(100) NOT NULL,
    batch_number INT NOT NULL,
    tasks_json TEXT,
    changes_json TEXT,
    start_index INT NOT NULL,
    end_index INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    processed_at TIMESTAMP NULL,
    referenceid INT NULL,
    modifiedon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(50) NULL,
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    createdby VARCHAR(50) NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    rowversionstamp INT DEFAULT 1,

    CONSTRAINT fk_batch_job FOREIGN KEY (batch_job_id)
        REFERENCES ai_schedule_batch_jobs(id) ON DELETE CASCADE,
    CONSTRAINT uk_batch_job_number UNIQUE (batch_job_id, batch_number)
);

-- Add comments for ai_schedule_batch_results columns
COMMENT ON TABLE ai_schedule_batch_results IS 'Stores individual batch results for AI schedule optimization';
COMMENT ON COLUMN ai_schedule_batch_results.id IS 'Unique identifier for the batch result';
COMMENT ON COLUMN ai_schedule_batch_results.batch_job_id IS 'Reference to parent batch job';
COMMENT ON COLUMN ai_schedule_batch_results.batch_number IS 'Batch sequence number';
COMMENT ON COLUMN ai_schedule_batch_results.tasks_json IS 'JSON array of optimized tasks for this batch';
COMMENT ON COLUMN ai_schedule_batch_results.changes_json IS 'JSON array of changes made in this batch';
COMMENT ON COLUMN ai_schedule_batch_results.start_index IS 'Starting index in original task list';
COMMENT ON COLUMN ai_schedule_batch_results.end_index IS 'Ending index in original task list';
COMMENT ON COLUMN ai_schedule_batch_results.status IS 'Batch status: pending, processing, completed, failed';
COMMENT ON COLUMN ai_schedule_batch_results.error_message IS 'Error message if batch failed';
COMMENT ON COLUMN ai_schedule_batch_results.processed_at IS 'When the batch was processed';
COMMENT ON COLUMN ai_schedule_batch_results.referenceid IS 'Reserved';
COMMENT ON COLUMN ai_schedule_batch_results.modifiedon IS 'The date when the record was last updated';
COMMENT ON COLUMN ai_schedule_batch_results.modifiedby IS 'The user which last updated the record';
COMMENT ON COLUMN ai_schedule_batch_results.createdon IS 'The date the record was created';
COMMENT ON COLUMN ai_schedule_batch_results.createdby IS 'The user who created the record';
COMMENT ON COLUMN ai_schedule_batch_results.active IS 'Indicates if the record is active or deleted';
COMMENT ON COLUMN ai_schedule_batch_results.rowversionstamp IS 'Current version identifier for the row, for detection of concurrency violations';

-- Create indexes for ai_schedule_batch_results
CREATE INDEX IF NOT EXISTS idx_batch_results_batch_job ON ai_schedule_batch_results(batch_job_id);
CREATE INDEX IF NOT EXISTS idx_batch_results_batch_number ON ai_schedule_batch_results(batch_job_id, batch_number);
CREATE INDEX IF NOT EXISTS idx_batch_results_status ON ai_schedule_batch_results(status);

-- Create trigger function to update modifiedon timestamp
CREATE OR REPLACE FUNCTION update_modifiedon_batch_results()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modifiedon = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for ai_schedule_batch_results
DROP TRIGGER IF EXISTS trigger_update_modifiedon_batch_results ON ai_schedule_batch_results;
CREATE TRIGGER trigger_update_modifiedon_batch_results
    BEFORE UPDATE ON ai_schedule_batch_results
    FOR EACH ROW
    EXECUTE FUNCTION update_modifiedon_batch_results();
