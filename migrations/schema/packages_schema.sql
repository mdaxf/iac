-- IAC Packages and Deployment Actions Schema
-- Supports MySQL, PostgreSQL, MSSQL, Oracle
-- Uses IAC standard naming convention: no snake_case, 7 standard fields at end

-- =====================================================
-- Table: iacpackages
-- Purpose: Store package content and metadata
-- =====================================================

CREATE TABLE iacpackages (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    packagetype VARCHAR(20) NOT NULL,  -- 'database' or 'document'
    description TEXT,
    metadata JSON,
    packagedata LONGTEXT NOT NULL,  -- JSON serialized package content
    databasetype VARCHAR(50),  -- mysql, postgresql, mssql, oracle, mongodb
    databasename VARCHAR(255),
    includeparent BOOLEAN DEFAULT FALSE,
    dependencies JSON,  -- Array of dependent package IDs
    checksum VARCHAR(64) NOT NULL,  -- SHA-256 hash for integrity
    filesize BIGINT,  -- Size in bytes
    status VARCHAR(20) DEFAULT 'active',  -- active, archived, deleted
    tags JSON,  -- Array of tags for categorization
    environment VARCHAR(50),  -- dev, staging, production
    -- IAC Standard Fields (7 fields)
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,
    INDEX idx_name (name),
    INDEX idx_packagetype (packagetype),
    INDEX idx_status (status),
    INDEX idx_createdon (createdon),
    INDEX idx_createdby (createdby),
    INDEX idx_environment (environment),
    INDEX idx_active (active),
    UNIQUE KEY unique_name_version (name, version)
);

-- =====================================================
-- Table: packageactions
-- Purpose: Record all package operations (pack, deploy, rollback)
-- =====================================================

CREATE TABLE packageactions (
    id VARCHAR(50) PRIMARY KEY,
    packageid VARCHAR(50) NOT NULL,
    actiontype VARCHAR(20) NOT NULL,  -- 'pack', 'deploy', 'rollback', 'export', 'import', 'validate'
    actionstatus VARCHAR(20) NOT NULL,  -- 'pending', 'in_progress', 'completed', 'failed', 'rolled_back'
    targetdatabase VARCHAR(255),  -- Target database for deployment
    targetenvironment VARCHAR(50),  -- dev, staging, production
    sourceenvironment VARCHAR(50),  -- Source environment for pack
    performedat DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    performedby VARCHAR(100) NOT NULL,
    startedat DATETIME,
    completedat DATETIME,
    durationseconds INT,  -- Action duration
    options JSON,  -- Deployment/pack options
    resultdata JSON,  -- PK mappings, ID mappings, statistics
    errorlog JSON,  -- Array of error messages
    warninglog JSON,  -- Array of warnings
    metadata JSON,  -- Additional action metadata
    recordsprocessed INT DEFAULT 0,
    recordssucceeded INT DEFAULT 0,
    recordsfailed INT DEFAULT 0,
    tablesprocessed INT DEFAULT 0,  -- For database packages
    collectionsprocessed INT DEFAULT 0,  -- For document packages
    -- IAC Standard Fields (7 fields)
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,
    FOREIGN KEY (packageid) REFERENCES iacpackages(id) ON DELETE CASCADE,
    INDEX idx_packageid (packageid),
    INDEX idx_actiontype (actiontype),
    INDEX idx_actionstatus (actionstatus),
    INDEX idx_performedat (performedat),
    INDEX idx_performedby (performedby),
    INDEX idx_targetenvironment (targetenvironment),
    INDEX idx_active (active)
);

-- =====================================================
-- Table: packagerelationships
-- Purpose: Track relationships between packages
-- =====================================================

CREATE TABLE packagerelationships (
    id VARCHAR(50) PRIMARY KEY,
    parentpackageid VARCHAR(50) NOT NULL,
    childpackageid VARCHAR(50) NOT NULL,
    relationshiptype VARCHAR(50) NOT NULL,  -- 'depends_on', 'extends', 'includes'
    -- IAC Standard Fields (7 fields)
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,
    FOREIGN KEY (parentpackageid) REFERENCES iacpackages(id) ON DELETE CASCADE,
    FOREIGN KEY (childpackageid) REFERENCES iacpackages(id) ON DELETE CASCADE,
    UNIQUE KEY unique_relationship (parentpackageid, childpackageid, relationshiptype),
    INDEX idx_parent (parentpackageid),
    INDEX idx_child (childpackageid),
    INDEX idx_active (active)
);

-- =====================================================
-- Table: packagedeployments
-- Purpose: Track active deployments per environment
-- =====================================================

CREATE TABLE packagedeployments (
    id VARCHAR(50) PRIMARY KEY,
    packageid VARCHAR(50) NOT NULL,
    actionid VARCHAR(50) NOT NULL,  -- Reference to packageactions
    environment VARCHAR(50) NOT NULL,
    databasename VARCHAR(255) NOT NULL,
    deployedat DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deployedby VARCHAR(100) NOT NULL,
    isactive BOOLEAN DEFAULT TRUE,
    rolledbackat DATETIME,
    rolledbackby VARCHAR(100),
    -- IAC Standard Fields (7 fields)
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,
    FOREIGN KEY (packageid) REFERENCES iacpackages(id) ON DELETE CASCADE,
    FOREIGN KEY (actionid) REFERENCES packageactions(id) ON DELETE CASCADE,
    INDEX idx_packageid (packageid),
    INDEX idx_environment (environment),
    INDEX idx_isactive (isactive),
    INDEX idx_deployedat (deployedat),
    INDEX idx_active (active)
);

-- =====================================================
-- Table: packagetags
-- Purpose: Tag-based categorization and search
-- =====================================================

CREATE TABLE packagetags (
    id VARCHAR(50) PRIMARY KEY,
    packageid VARCHAR(50) NOT NULL,
    tagname VARCHAR(100) NOT NULL,
    -- IAC Standard Fields (7 fields)
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,
    FOREIGN KEY (packageid) REFERENCES iacpackages(id) ON DELETE CASCADE,
    INDEX idx_packageid (packageid),
    INDEX idx_tagname (tagname),
    INDEX idx_active (active),
    UNIQUE KEY unique_package_tag (packageid, tagname)
);

-- =====================================================
-- Sample queries for common operations
-- =====================================================

-- Get all active packages with their latest action
-- SELECT p.*, a.actiontype, a.actionstatus, a.performedat
-- FROM iacpackages p
-- LEFT JOIN packageactions a ON a.id = (
--     SELECT id FROM packageactions
--     WHERE packageid = p.id AND active = TRUE
--     ORDER BY performedat DESC
--     LIMIT 1
-- )
-- WHERE p.active = TRUE
-- ORDER BY p.createdon DESC;

-- Get deployment history for a package
-- SELECT * FROM packageactions
-- WHERE packageid = ? AND actiontype IN ('deploy', 'rollback') AND active = TRUE
-- ORDER BY performedat DESC;

-- Get currently deployed packages by environment
-- SELECT p.*, pd.environment, pd.deployedat
-- FROM iacpackages p
-- INNER JOIN packagedeployments pd ON pd.packageid = p.id
-- WHERE pd.environment = ? AND pd.isactive = TRUE AND pd.active = TRUE AND p.active = TRUE;

-- Get package with dependencies
-- SELECT p.*, pr.childpackageid, cp.name as dep_name, cp.version as dep_version
-- FROM iacpackages p
-- LEFT JOIN packagerelationships pr ON pr.parentpackageid = p.id AND pr.active = TRUE
-- LEFT JOIN iacpackages cp ON cp.id = pr.childpackageid AND cp.active = TRUE
-- WHERE p.id = ? AND p.active = TRUE;
