-- IAC Packages and Deployment Actions Schema - PostgreSQL
-- Uses JSONB for better performance
-- Uses IAC standard naming convention: no snake_case, 7 standard fields at end

-- =====================================================
-- Table: iacpackages
-- Purpose: Store package content and metadata
-- =====================================================

CREATE TABLE iacpackages (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    packagetype VARCHAR(20) NOT NULL CHECK (packagetype IN ('database', 'document')),
    description TEXT,
    metadata JSONB,
    packagedata JSONB NOT NULL,  -- Store as JSONB for better indexing
    databasetype VARCHAR(50),
    databasename VARCHAR(255),
    includeparent BOOLEAN DEFAULT FALSE,
    dependencies JSONB,
    checksum VARCHAR(64) NOT NULL,
    filesize BIGINT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'archived', 'deleted')),
    tags JSONB,
    environment VARCHAR(50),
    -- IAC Standard Fields (7 fields)
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT unique_name_version UNIQUE (name, version)
);

CREATE INDEX idx_packages_name ON iacpackages(name);
CREATE INDEX idx_packages_type ON iacpackages(packagetype);
CREATE INDEX idx_packages_status ON iacpackages(status);
CREATE INDEX idx_packages_createdon ON iacpackages(createdon);
CREATE INDEX idx_packages_createdby ON iacpackages(createdby);
CREATE INDEX idx_packages_env ON iacpackages(environment);
CREATE INDEX idx_packages_active ON iacpackages(active);
CREATE INDEX idx_packages_tags ON iacpackages USING GIN(tags);
CREATE INDEX idx_packages_metadata ON iacpackages USING GIN(metadata);

-- =====================================================
-- Table: packageactions
-- Purpose: Record all package operations
-- =====================================================

CREATE TABLE packageactions (
    id VARCHAR(50) PRIMARY KEY,
    packageid VARCHAR(50) NOT NULL,
    actiontype VARCHAR(20) NOT NULL CHECK (actiontype IN ('pack', 'deploy', 'rollback', 'export', 'import', 'validate')),
    actionstatus VARCHAR(20) NOT NULL CHECK (actionstatus IN ('pending', 'in_progress', 'completed', 'failed', 'rolled_back')),
    targetdatabase VARCHAR(255),
    targetenvironment VARCHAR(50),
    sourceenvironment VARCHAR(50),
    performedat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    performedby VARCHAR(100) NOT NULL,
    startedat TIMESTAMP,
    completedat TIMESTAMP,
    durationseconds INTEGER,
    options JSONB,
    resultdata JSONB,
    errorlog JSONB,
    warninglog JSONB,
    metadata JSONB,
    recordsprocessed INTEGER DEFAULT 0,
    recordssucceeded INTEGER DEFAULT 0,
    recordsfailed INTEGER DEFAULT 0,
    tablesprocessed INTEGER DEFAULT 0,
    collectionsprocessed INTEGER DEFAULT 0,
    -- IAC Standard Fields (7 fields)
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (packageid) REFERENCES iacpackages(id) ON DELETE CASCADE
);

CREATE INDEX idx_actions_packageid ON packageactions(packageid);
CREATE INDEX idx_actions_type ON packageactions(actiontype);
CREATE INDEX idx_actions_status ON packageactions(actionstatus);
CREATE INDEX idx_actions_performedat ON packageactions(performedat);
CREATE INDEX idx_actions_performedby ON packageactions(performedby);
CREATE INDEX idx_actions_targetenv ON packageactions(targetenvironment);
CREATE INDEX idx_actions_active ON packageactions(active);
CREATE INDEX idx_actions_result ON packageactions USING GIN(resultdata);

-- =====================================================
-- Table: packagerelationships
-- =====================================================

CREATE TABLE packagerelationships (
    id VARCHAR(50) PRIMARY KEY,
    parentpackageid VARCHAR(50) NOT NULL,
    childpackageid VARCHAR(50) NOT NULL,
    relationshiptype VARCHAR(50) NOT NULL,
    -- IAC Standard Fields (7 fields)
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (parentpackageid) REFERENCES iacpackages(id) ON DELETE CASCADE,
    FOREIGN KEY (childpackageid) REFERENCES iacpackages(id) ON DELETE CASCADE,
    CONSTRAINT unique_relationship UNIQUE (parentpackageid, childpackageid, relationshiptype)
);

CREATE INDEX idx_relationships_parent ON packagerelationships(parentpackageid);
CREATE INDEX idx_relationships_child ON packagerelationships(childpackageid);
CREATE INDEX idx_relationships_active ON packagerelationships(active);

-- =====================================================
-- Table: packagedeployments
-- =====================================================

CREATE TABLE packagedeployments (
    id VARCHAR(50) PRIMARY KEY,
    packageid VARCHAR(50) NOT NULL,
    actionid VARCHAR(50) NOT NULL,
    environment VARCHAR(50) NOT NULL,
    databasename VARCHAR(255) NOT NULL,
    deployedat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deployedby VARCHAR(100) NOT NULL,
    isactive BOOLEAN DEFAULT TRUE,
    rolledbackat TIMESTAMP,
    rolledbackby VARCHAR(100),
    -- IAC Standard Fields (7 fields)
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (packageid) REFERENCES iacpackages(id) ON DELETE CASCADE,
    FOREIGN KEY (actionid) REFERENCES packageactions(id) ON DELETE CASCADE
);

CREATE INDEX idx_deployments_packageid ON packagedeployments(packageid);
CREATE INDEX idx_deployments_env ON packagedeployments(environment);
CREATE INDEX idx_deployments_isactive ON packagedeployments(isactive);
CREATE INDEX idx_deployments_deployedat ON packagedeployments(deployedat);
CREATE INDEX idx_deployments_active ON packagedeployments(active);

-- =====================================================
-- Table: packagetags
-- =====================================================

CREATE TABLE packagetags (
    id VARCHAR(50) PRIMARY KEY,
    packageid VARCHAR(50) NOT NULL,
    tagname VARCHAR(100) NOT NULL,
    -- IAC Standard Fields (7 fields)
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (packageid) REFERENCES iacpackages(id) ON DELETE CASCADE,
    CONSTRAINT unique_package_tag UNIQUE (packageid, tagname)
);

CREATE INDEX idx_tags_packageid ON packagetags(packageid);
CREATE INDEX idx_tags_tagname ON packagetags(tagname);
CREATE INDEX idx_tags_active ON packagetags(active);

-- =====================================================
-- Functions and Triggers
-- =====================================================

-- Auto-update modifiedon timestamp and increment rowversionstamp
CREATE OR REPLACE FUNCTION update_modifiedon_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modifiedon = CURRENT_TIMESTAMP;
    NEW.rowversionstamp = OLD.rowversionstamp + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_iacpackages_modifiedon
    BEFORE UPDATE ON iacpackages
    FOR EACH ROW
    EXECUTE FUNCTION update_modifiedon_column();

CREATE TRIGGER update_packageactions_modifiedon
    BEFORE UPDATE ON packageactions
    FOR EACH ROW
    EXECUTE FUNCTION update_modifiedon_column();

CREATE TRIGGER update_packagerelationships_modifiedon
    BEFORE UPDATE ON packagerelationships
    FOR EACH ROW
    EXECUTE FUNCTION update_modifiedon_column();

CREATE TRIGGER update_packagedeployments_modifiedon
    BEFORE UPDATE ON packagedeployments
    FOR EACH ROW
    EXECUTE FUNCTION update_modifiedon_column();

CREATE TRIGGER update_packagetags_modifiedon
    BEFORE UPDATE ON packagetags
    FOR EACH ROW
    EXECUTE FUNCTION update_modifiedon_column();

-- Calculate action duration on completion
CREATE OR REPLACE FUNCTION calculate_action_duration()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.actionstatus IN ('completed', 'failed', 'rolled_back') AND NEW.startedat IS NOT NULL THEN
        NEW.durationseconds = EXTRACT(EPOCH FROM (NEW.completedat - NEW.startedat))::INTEGER;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER calculate_package_action_duration
    BEFORE UPDATE ON packageactions
    FOR EACH ROW
    EXECUTE FUNCTION calculate_action_duration();
