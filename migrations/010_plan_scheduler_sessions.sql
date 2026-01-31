-- =====================================================================
-- Plan Scheduler Sessions Table
-- Purpose: Store saved plan/schedule sessions for profiles
-- Use case: When datasource is query-based (not static JSON), we need
--           to save different planned schedules as separate sessions
-- =====================================================================

-- Drop table if exists (for re-running migration)
DROP TABLE IF EXISTS plan_scheduler_sessions CASCADE;

-- Create plan_scheduler_sessions table
CREATE TABLE plan_scheduler_sessions (
    id VARCHAR(36) PRIMARY KEY,
    profile_id VARCHAR(36) NOT NULL,
    session_id VARCHAR(100) NOT NULL,
    session_name VARCHAR(200),
    description TEXT,

    -- Saved schedule data
    tasks_data JSONB,
    resources_data JSONB,
    metadata JSONB,

    -- Session status
    status VARCHAR(20) DEFAULT 'draft', -- 'draft', 'active', 'archived'
    is_default BOOLEAN DEFAULT FALSE,

    -- Standard IAC 7 moderate fields
    active BOOLEAN DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(100),
    createdon TIMESTAMP DEFAULT NOW(),
    modifiedby VARCHAR(100),
    modifiedon TIMESTAMP DEFAULT NOW(),
    rowversionstamp BIGINT DEFAULT 0,

    -- Foreign key constraint
    CONSTRAINT fk_session_profile FOREIGN KEY (profile_id)
        REFERENCES plan_scheduler_profiles(id) ON DELETE CASCADE,

    -- Unique constraint: one session_id per profile
    CONSTRAINT uq_profile_session UNIQUE (profile_id, session_id)
);

-- Indexes for performance
CREATE INDEX idx_sessions_profile_id ON plan_scheduler_sessions(profile_id);
CREATE INDEX idx_sessions_session_id ON plan_scheduler_sessions(session_id);
CREATE INDEX idx_sessions_status ON plan_scheduler_sessions(status);
CREATE INDEX idx_sessions_is_default ON plan_scheduler_sessions(profile_id, is_default);
CREATE INDEX idx_sessions_createdon ON plan_scheduler_sessions(createdon DESC);

-- Comments
COMMENT ON TABLE plan_scheduler_sessions IS 'Stores saved plan/schedule sessions for profiles. Used when datasource is query-based (not static JSON).';
COMMENT ON COLUMN plan_scheduler_sessions.session_id IS 'User-defined session identifier (e.g., "2024-Q1-Plan", "Optimized-v2")';
COMMENT ON COLUMN plan_scheduler_sessions.tasks_data IS 'Saved task schedule data as JSONB';
COMMENT ON COLUMN plan_scheduler_sessions.resources_data IS 'Saved resource allocation data as JSONB';
COMMENT ON COLUMN plan_scheduler_sessions.metadata IS 'Additional metadata (optimization params, KPIs, etc.)';
COMMENT ON COLUMN plan_scheduler_sessions.status IS 'Session status: draft, active, archived';
COMMENT ON COLUMN plan_scheduler_sessions.is_default IS 'Whether this is the default session for the profile';

-- =====================================================================
-- Sample Data (Optional - for testing)
-- =====================================================================

-- Insert a sample session
INSERT INTO plan_scheduler_sessions (
    id, profile_id, session_id, session_name, description,
    tasks_data, resources_data, metadata, status, is_default,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES (
    'session-sample-001',
    'demo-profile-001',  -- Assumes demo profile exists
    'baseline-2024-q1',
    'Baseline Plan - Q1 2024',
    'Initial baseline schedule for Q1 2024 production',
    '[]'::jsonb,  -- Empty tasks array
    '[]'::jsonb,  -- Empty resources array
    '{"optimizationLevel": "balanced", "totalMakespan": 0}'::jsonb,
    'draft',
    TRUE,
    TRUE,
    'SESSION-BASELINE',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- =====================================================================
-- Verification Query
-- =====================================================================

SELECT
    id,
    profile_id,
    session_id,
    session_name,
    status,
    is_default,
    createdon
FROM plan_scheduler_sessions
ORDER BY createdon DESC;

-- =====================================================================
-- END OF MIGRATION
-- =====================================================================
