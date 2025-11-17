-- IAC Packages and Deployment Actions Schema - PostgreSQL
-- Uses JSONB for better performance

-- =====================================================
-- Table: iacpackages
-- Purpose: Store package content and metadata
-- =====================================================

CREATE TABLE iacpackages (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    package_type VARCHAR(20) NOT NULL CHECK (package_type IN ('database', 'document')),
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL,
    metadata JSONB,
    package_data JSONB NOT NULL,  -- Store as JSONB for better indexing
    database_type VARCHAR(50),
    database_name VARCHAR(255),
    include_parent BOOLEAN DEFAULT FALSE,
    dependencies JSONB,
    checksum VARCHAR(64) NOT NULL,
    file_size BIGINT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'archived', 'deleted')),
    tags JSONB,
    environment VARCHAR(50),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_name_version UNIQUE (name, version)
);

CREATE INDEX idx_packages_name ON iacpackages(name);
CREATE INDEX idx_packages_type ON iacpackages(package_type);
CREATE INDEX idx_packages_status ON iacpackages(status);
CREATE INDEX idx_packages_created ON iacpackages(created_at);
CREATE INDEX idx_packages_creator ON iacpackages(created_by);
CREATE INDEX idx_packages_env ON iacpackages(environment);
CREATE INDEX idx_packages_tags ON iacpackages USING GIN(tags);
CREATE INDEX idx_packages_metadata ON iacpackages USING GIN(metadata);

-- =====================================================
-- Table: package_actions
-- Purpose: Record all package operations
-- =====================================================

CREATE TABLE package_actions (
    id VARCHAR(50) PRIMARY KEY,
    package_id VARCHAR(50) NOT NULL,
    action_type VARCHAR(20) NOT NULL CHECK (action_type IN ('pack', 'deploy', 'rollback', 'export', 'import', 'validate')),
    action_status VARCHAR(20) NOT NULL CHECK (action_status IN ('pending', 'in_progress', 'completed', 'failed', 'rolled_back')),
    target_database VARCHAR(255),
    target_environment VARCHAR(50),
    source_environment VARCHAR(50),
    performed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    performed_by VARCHAR(100) NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_seconds INTEGER,
    options JSONB,
    result_data JSONB,
    error_log JSONB,
    warning_log JSONB,
    metadata JSONB,
    records_processed INTEGER DEFAULT 0,
    records_succeeded INTEGER DEFAULT 0,
    records_failed INTEGER DEFAULT 0,
    tables_processed INTEGER DEFAULT 0,
    collections_processed INTEGER DEFAULT 0,
    FOREIGN KEY (package_id) REFERENCES iacpackages(id) ON DELETE CASCADE
);

CREATE INDEX idx_actions_package ON package_actions(package_id);
CREATE INDEX idx_actions_type ON package_actions(action_type);
CREATE INDEX idx_actions_status ON package_actions(action_status);
CREATE INDEX idx_actions_performed ON package_actions(performed_at);
CREATE INDEX idx_actions_performer ON package_actions(performed_by);
CREATE INDEX idx_actions_target_env ON package_actions(target_environment);
CREATE INDEX idx_actions_result ON package_actions USING GIN(result_data);

-- =====================================================
-- Table: package_relationships
-- =====================================================

CREATE TABLE package_relationships (
    id VARCHAR(50) PRIMARY KEY,
    parent_package_id VARCHAR(50) NOT NULL,
    child_package_id VARCHAR(50) NOT NULL,
    relationship_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_package_id) REFERENCES iacpackages(id) ON DELETE CASCADE,
    FOREIGN KEY (child_package_id) REFERENCES iacpackages(id) ON DELETE CASCADE,
    CONSTRAINT unique_relationship UNIQUE (parent_package_id, child_package_id, relationship_type)
);

CREATE INDEX idx_relationships_parent ON package_relationships(parent_package_id);
CREATE INDEX idx_relationships_child ON package_relationships(child_package_id);

-- =====================================================
-- Table: package_deployments
-- =====================================================

CREATE TABLE package_deployments (
    id VARCHAR(50) PRIMARY KEY,
    package_id VARCHAR(50) NOT NULL,
    action_id VARCHAR(50) NOT NULL,
    environment VARCHAR(50) NOT NULL,
    database_name VARCHAR(255) NOT NULL,
    deployed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deployed_by VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    rolled_back_at TIMESTAMP,
    rolled_back_by VARCHAR(100),
    FOREIGN KEY (package_id) REFERENCES iacpackages(id) ON DELETE CASCADE,
    FOREIGN KEY (action_id) REFERENCES package_actions(id) ON DELETE CASCADE
);

CREATE INDEX idx_deployments_package ON package_deployments(package_id);
CREATE INDEX idx_deployments_env ON package_deployments(environment);
CREATE INDEX idx_deployments_active ON package_deployments(is_active);
CREATE INDEX idx_deployments_deployed ON package_deployments(deployed_at);

-- =====================================================
-- Table: package_tags
-- =====================================================

CREATE TABLE package_tags (
    id VARCHAR(50) PRIMARY KEY,
    package_id VARCHAR(50) NOT NULL,
    tag_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    FOREIGN KEY (package_id) REFERENCES iacpackages(id) ON DELETE CASCADE,
    CONSTRAINT unique_package_tag UNIQUE (package_id, tag_name)
);

CREATE INDEX idx_tags_package ON package_tags(package_id);
CREATE INDEX idx_tags_name ON package_tags(tag_name);

-- =====================================================
-- Functions and Triggers
-- =====================================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_iacpackages_updated_at
    BEFORE UPDATE ON iacpackages
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Calculate action duration on completion
CREATE OR REPLACE FUNCTION calculate_action_duration()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.action_status IN ('completed', 'failed', 'rolled_back') AND NEW.started_at IS NOT NULL THEN
        NEW.duration_seconds = EXTRACT(EPOCH FROM (NEW.completed_at - NEW.started_at))::INTEGER;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER calculate_package_action_duration
    BEFORE UPDATE ON package_actions
    FOR EACH ROW
    EXECUTE FUNCTION calculate_action_duration();
