-- Plan Scheduler Profile Tables for PostgreSQL
-- This creates tables to store plan scheduler profiles with data source queries and constraints

-- Main profile table
CREATE TABLE IF NOT EXISTS plan_scheduler_profiles (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT FALSE,

    -- Standard IAC 7 moderate fields
    active BOOLEAN DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(100),
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(100),
    modifiedon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp BIGINT DEFAULT 0
);

-- Profile data sources (tasks, resources, master data)
CREATE TABLE IF NOT EXISTS plan_scheduler_profile_datasources (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    profile_id VARCHAR(36) NOT NULL REFERENCES plan_scheduler_profiles(id) ON DELETE CASCADE,

    data_type VARCHAR(50) NOT NULL, -- 'tasks', 'resources', 'materials', 'equipment', 'custom'
    name VARCHAR(200) NOT NULL, -- Display name for this data source
    description TEXT,

    -- Data source configuration
    source_type VARCHAR(20) NOT NULL, -- 'query' or 'json'
    source_query TEXT, -- SQL query if source_type='query'
    source_json JSONB, -- JSON constant data if source_type='json'

    -- Query parameters (for parameterized queries)
    parameters JSONB, -- { "param1": "value1", "param2": "value2" }

    -- Display order
    display_order INTEGER DEFAULT 0,

    -- Standard IAC 7 moderate fields
    active BOOLEAN DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(100),
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(100),
    modifiedon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp BIGINT DEFAULT 0,

    CONSTRAINT uq_profile_datasource_name UNIQUE (profile_id, data_type, name)
);

-- Profile constraints
CREATE TABLE IF NOT EXISTS plan_scheduler_profile_constraints (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    profile_id VARCHAR(36) NOT NULL REFERENCES plan_scheduler_profiles(id) ON DELETE CASCADE,

    constraint_type VARCHAR(50) NOT NULL, -- 'time', 'resource', 'dependency', 'custom'
    name VARCHAR(200) NOT NULL,
    description TEXT,

    -- Constraint configuration
    source_type VARCHAR(20) NOT NULL, -- 'query' or 'json'
    source_query TEXT, -- SQL query that returns constraint values
    source_json JSONB, -- JSON constant constraint values

    -- Constraint enforcement level
    enforcement VARCHAR(20) DEFAULT 'hard', -- 'hard', 'soft', 'advisory'

    -- Display order
    display_order INTEGER DEFAULT 0,

    -- Standard IAC 7 moderate fields
    active BOOLEAN DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(100),
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(100),
    modifiedon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp BIGINT DEFAULT 0,

    CONSTRAINT uq_profile_constraint_name UNIQUE (profile_id, constraint_type, name)
);

-- Profile settings (additional configuration options)
CREATE TABLE IF NOT EXISTS plan_scheduler_profile_settings (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    profile_id VARCHAR(36) NOT NULL REFERENCES plan_scheduler_profiles(id) ON DELETE CASCADE,

    setting_key VARCHAR(100) NOT NULL,
    setting_value JSONB,
    description TEXT,

    -- Standard IAC 7 moderate fields
    active BOOLEAN DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(100),
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(100),
    modifiedon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp BIGINT DEFAULT 0,

    CONSTRAINT uq_profile_setting UNIQUE (profile_id, setting_key)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_profile_datasources_profile ON plan_scheduler_profile_datasources(profile_id);
CREATE INDEX IF NOT EXISTS idx_profile_datasources_type ON plan_scheduler_profile_datasources(data_type);
CREATE INDEX IF NOT EXISTS idx_profile_constraints_profile ON plan_scheduler_profile_constraints(profile_id);
CREATE INDEX IF NOT EXISTS idx_profile_constraints_type ON plan_scheduler_profile_constraints(constraint_type);
CREATE INDEX IF NOT EXISTS idx_profile_settings_profile ON plan_scheduler_profile_settings(profile_id);
CREATE INDEX IF NOT EXISTS idx_profile_default ON plan_scheduler_profiles(is_default) WHERE is_default = TRUE;

-- Insert sample profile for demonstration
INSERT INTO plan_scheduler_profiles (id, name, description, is_default, createdby)
VALUES
    ('default-profile-001', 'Default Production Profile', 'Default profile for production scheduling', TRUE, 'system')
ON CONFLICT (id) DO NOTHING;

-- Insert sample datasources
INSERT INTO plan_scheduler_profile_datasources (id, profile_id, data_type, name, source_type, source_query, display_order, createdby)
VALUES
    ('ds-tasks-001', 'default-profile-001', 'tasks', 'Production Tasks', 'query',
     'SELECT id, name, description, start_date as start, end_date as end, status, progress, resource_ids as "resourceIds", requirements, dependencies FROM production_tasks WHERE active = true ORDER BY start_date',
     1, 'system'),
    ('ds-resources-001', 'default-profile-001', 'resources', 'Production Resources', 'query',
     'SELECT id, name, type, role, capacity, cost_per_hour as "costPerHour" FROM production_resources WHERE active = true ORDER BY name',
     2, 'system')
ON CONFLICT (id) DO NOTHING;

-- Insert sample constraints
INSERT INTO plan_scheduler_profile_constraints (id, profile_id, constraint_type, name, source_type, source_json, enforcement, display_order, createdby)
VALUES
    ('const-time-001', 'default-profile-001', 'time', 'Working Hours', 'json',
     '{"workingDays": [1,2,3,4,5], "startHour": 8, "endHour": 17, "timezone": "UTC"}',
     'hard', 1, 'system'),
    ('const-resource-001', 'default-profile-001', 'resource', 'Max Resource Utilization', 'json',
     '{"maxUtilization": 0.85, "allowOvertime": false}',
     'soft', 2, 'system')
ON CONFLICT (id) DO NOTHING;

-- Comments for documentation
COMMENT ON TABLE plan_scheduler_profiles IS 'Stores plan scheduler profile definitions';
COMMENT ON TABLE plan_scheduler_profile_datasources IS 'Stores data source configurations (queries or JSON) for tasks, resources, and master data';
COMMENT ON TABLE plan_scheduler_profile_constraints IS 'Stores constraint definitions for scheduling optimization';
COMMENT ON TABLE plan_scheduler_profile_settings IS 'Stores additional profile-specific settings';
