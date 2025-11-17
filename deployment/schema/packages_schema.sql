-- IAC Packages and Deployment Actions Schema
-- Supports MySQL, PostgreSQL, MSSQL, Oracle

-- =====================================================
-- Table: iacpackages
-- Purpose: Store package content and metadata
-- =====================================================

CREATE TABLE iacpackages (
    id VARCHAR(50) PRIMARY KEY,  -- UUID format
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    package_type VARCHAR(20) NOT NULL,  -- 'database' or 'document'
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL,
    metadata JSON,  -- Additional metadata
    package_data LONGTEXT NOT NULL,  -- JSON serialized package content
    database_type VARCHAR(50),  -- mysql, postgresql, mssql, oracle, mongodb
    database_name VARCHAR(255),
    include_parent BOOLEAN DEFAULT FALSE,
    dependencies JSON,  -- Array of dependent package IDs
    checksum VARCHAR(64) NOT NULL,  -- SHA-256 hash for integrity
    file_size BIGINT,  -- Size in bytes
    status VARCHAR(20) DEFAULT 'active',  -- active, archived, deleted
    tags JSON,  -- Array of tags for categorization
    environment VARCHAR(50),  -- dev, staging, production
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_package_type (package_type),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_created_by (created_by),
    INDEX idx_environment (environment),
    UNIQUE KEY unique_name_version (name, version)
);

-- =====================================================
-- Table: package_actions
-- Purpose: Record all package operations (pack, deploy, rollback)
-- =====================================================

CREATE TABLE package_actions (
    id VARCHAR(50) PRIMARY KEY,  -- UUID format
    package_id VARCHAR(50) NOT NULL,
    action_type VARCHAR(20) NOT NULL,  -- 'pack', 'deploy', 'rollback', 'export', 'import', 'validate'
    action_status VARCHAR(20) NOT NULL,  -- 'pending', 'in_progress', 'completed', 'failed', 'rolled_back'
    target_database VARCHAR(255),  -- Target database for deployment
    target_environment VARCHAR(50),  -- dev, staging, production
    source_environment VARCHAR(50),  -- Source environment for pack
    performed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    performed_by VARCHAR(100) NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_seconds INT,  -- Action duration
    options JSON,  -- Deployment/pack options
    result_data JSON,  -- PK mappings, ID mappings, statistics
    error_log JSON,  -- Array of error messages
    warning_log JSON,  -- Array of warnings
    metadata JSON,  -- Additional action metadata
    records_processed INT DEFAULT 0,
    records_succeeded INT DEFAULT 0,
    records_failed INT DEFAULT 0,
    tables_processed INT DEFAULT 0,  -- For database packages
    collections_processed INT DEFAULT 0,  -- For document packages
    FOREIGN KEY (package_id) REFERENCES iacpackages(id) ON DELETE CASCADE,
    INDEX idx_package_id (package_id),
    INDEX idx_action_type (action_type),
    INDEX idx_action_status (action_status),
    INDEX idx_performed_at (performed_at),
    INDEX idx_performed_by (performed_by),
    INDEX idx_target_environment (target_environment)
);

-- =====================================================
-- Table: package_relationships
-- Purpose: Track relationships between packages
-- =====================================================

CREATE TABLE package_relationships (
    id VARCHAR(50) PRIMARY KEY,
    parent_package_id VARCHAR(50) NOT NULL,
    child_package_id VARCHAR(50) NOT NULL,
    relationship_type VARCHAR(50) NOT NULL,  -- 'depends_on', 'extends', 'includes'
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_package_id) REFERENCES iacpackages(id) ON DELETE CASCADE,
    FOREIGN KEY (child_package_id) REFERENCES iacpackages(id) ON DELETE CASCADE,
    UNIQUE KEY unique_relationship (parent_package_id, child_package_id, relationship_type),
    INDEX idx_parent (parent_package_id),
    INDEX idx_child (child_package_id)
);

-- =====================================================
-- Table: package_deployments
-- Purpose: Track active deployments per environment
-- =====================================================

CREATE TABLE package_deployments (
    id VARCHAR(50) PRIMARY KEY,
    package_id VARCHAR(50) NOT NULL,
    action_id VARCHAR(50) NOT NULL,  -- Reference to package_actions
    environment VARCHAR(50) NOT NULL,
    database_name VARCHAR(255) NOT NULL,
    deployed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deployed_by VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    rolled_back_at TIMESTAMP,
    rolled_back_by VARCHAR(100),
    FOREIGN KEY (package_id) REFERENCES iacpackages(id) ON DELETE CASCADE,
    FOREIGN KEY (action_id) REFERENCES package_actions(id) ON DELETE CASCADE,
    INDEX idx_package_id (package_id),
    INDEX idx_environment (environment),
    INDEX idx_is_active (is_active),
    INDEX idx_deployed_at (deployed_at)
);

-- =====================================================
-- Table: package_tags
-- Purpose: Tag-based categorization and search
-- =====================================================

CREATE TABLE package_tags (
    id VARCHAR(50) PRIMARY KEY,
    package_id VARCHAR(50) NOT NULL,
    tag_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    FOREIGN KEY (package_id) REFERENCES iacpackages(id) ON DELETE CASCADE,
    INDEX idx_package_id (package_id),
    INDEX idx_tag_name (tag_name),
    UNIQUE KEY unique_package_tag (package_id, tag_name)
);

-- =====================================================
-- Sample queries for common operations
-- =====================================================

-- Get all packages with their latest action
-- SELECT p.*, a.action_type, a.action_status, a.performed_at
-- FROM iacpackages p
-- LEFT JOIN package_actions a ON a.id = (
--     SELECT id FROM package_actions
--     WHERE package_id = p.id
--     ORDER BY performed_at DESC
--     LIMIT 1
-- )
-- ORDER BY p.created_at DESC;

-- Get deployment history for a package
-- SELECT * FROM package_actions
-- WHERE package_id = ? AND action_type IN ('deploy', 'rollback')
-- ORDER BY performed_at DESC;

-- Get currently deployed packages by environment
-- SELECT p.*, pd.environment, pd.deployed_at
-- FROM iacpackages p
-- INNER JOIN package_deployments pd ON pd.package_id = p.id
-- WHERE pd.environment = ? AND pd.is_active = TRUE;

-- Get package with dependencies
-- SELECT p.*, pr.child_package_id, cp.name as dep_name, cp.version as dep_version
-- FROM iacpackages p
-- LEFT JOIN package_relationships pr ON pr.parent_package_id = p.id
-- LEFT JOIN iacpackages cp ON cp.id = pr.child_package_id
-- WHERE p.id = ?;
