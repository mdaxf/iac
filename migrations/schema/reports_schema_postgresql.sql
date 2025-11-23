-- IAC Reports Schema for PostgreSQL
-- Follows IAC naming convention: lowercase, no underscores

-- =====================================================
-- Table: reports
-- Purpose: Store report definitions
-- =====================================================

CREATE TABLE reports (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    reporttype VARCHAR(20) NOT NULL DEFAULT 'manual',  -- 'manual', 'ai_generated', 'template'
    ispublic BOOLEAN NOT NULL DEFAULT FALSE,
    istemplate BOOLEAN NOT NULL DEFAULT FALSE,
    layoutconfig JSONB,  -- Layout configuration
    pagesettings JSONB,  -- Page settings
    aiprompt TEXT,  -- AI generation prompt
    aianalysis JSONB,  -- AI analysis results
    templatesourceid VARCHAR(36),  -- Source template if created from template
    tags JSONB,  -- Tags array
    version INT NOT NULL DEFAULT 1,
    lastexecutedon TIMESTAMP,

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1
);

CREATE INDEX idx_reports_reporttype ON reports(reporttype);
CREATE INDEX idx_reports_ispublic ON reports(ispublic);
CREATE INDEX idx_reports_istemplate ON reports(istemplate);
CREATE INDEX idx_reports_createdby ON reports(createdby);
CREATE INDEX idx_reports_createdon ON reports(createdon);

-- Trigger to update modifiedon
CREATE OR REPLACE FUNCTION update_modifiedon_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modifiedon = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_reports_modifiedon BEFORE UPDATE ON reports
    FOR EACH ROW EXECUTE FUNCTION update_modifiedon_column();

-- =====================================================
-- Table: reportdatasources
-- Purpose: Store report data sources
-- =====================================================

CREATE TABLE reportdatasources (
    id VARCHAR(36) PRIMARY KEY,
    reportid VARCHAR(36) NOT NULL,
    alias VARCHAR(100) NOT NULL,
    databasealias VARCHAR(100),
    querytype VARCHAR(20) DEFAULT 'visual',  -- 'visual' or 'custom'
    customsql TEXT,
    selectedtables JSONB,  -- Array of selected tables
    selectedfields JSONB,  -- Array of selected fields with aliases
    joins JSONB,  -- Array of join definitions
    filters JSONB,  -- Array of filter conditions
    sorting JSONB,  -- Array of sort definitions
    grouping JSONB,  -- Array of group by fields
    parameters JSONB,  -- Query parameters

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (reportid) REFERENCES reports(id) ON DELETE CASCADE
);

CREATE INDEX idx_reportdatasources_reportid ON reportdatasources(reportid);
CREATE INDEX idx_reportdatasources_alias ON reportdatasources(alias);
CREATE INDEX idx_reportdatasources_querytype ON reportdatasources(querytype);

CREATE TRIGGER update_reportdatasources_modifiedon BEFORE UPDATE ON reportdatasources
    FOR EACH ROW EXECUTE FUNCTION update_modifiedon_column();

-- =====================================================
-- Table: reportcomponents
-- Purpose: Store report visual components
-- =====================================================

CREATE TABLE reportcomponents (
    id VARCHAR(36) PRIMARY KEY,
    reportid VARCHAR(36) NOT NULL,
    componenttype VARCHAR(20) NOT NULL,  -- 'table', 'chart', 'barcode', 'sub_report', 'text', 'image', 'drill_down'
    name VARCHAR(255) NOT NULL,
    x DECIMAL(10,2) DEFAULT 0,
    y DECIMAL(10,2) DEFAULT 0,
    width DECIMAL(10,2) DEFAULT 200,
    height DECIMAL(10,2) DEFAULT 100,
    zindex INT DEFAULT 0,
    datasourcealias VARCHAR(100),
    dataconfig JSONB,
    componentconfig JSONB,
    styleconfig JSONB,
    charttype VARCHAR(20),  -- Chart types if applicable
    chartconfig JSONB,
    barcodetype VARCHAR(20),  -- Barcode types if applicable
    barcodeconfig JSONB,
    drilldownconfig JSONB,
    conditionalformatting JSONB,
    isvisible BOOLEAN NOT NULL DEFAULT TRUE,

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (reportid) REFERENCES reports(id) ON DELETE CASCADE
);

CREATE INDEX idx_reportcomponents_reportid ON reportcomponents(reportid);
CREATE INDEX idx_reportcomponents_componenttype ON reportcomponents(componenttype);
CREATE INDEX idx_reportcomponents_zindex ON reportcomponents(zindex);

CREATE TRIGGER update_reportcomponents_modifiedon BEFORE UPDATE ON reportcomponents
    FOR EACH ROW EXECUTE FUNCTION update_modifiedon_column();

-- =====================================================
-- Table: reportparameters
-- Purpose: Store report input parameters
-- =====================================================

CREATE TABLE reportparameters (
    id VARCHAR(36) PRIMARY KEY,
    reportid VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    displayname VARCHAR(100),
    parametertype VARCHAR(20) DEFAULT 'text',  -- 'text', 'number', 'date', 'datetime', 'select', 'multi_select', 'boolean'
    defaultvalue TEXT,
    isrequired BOOLEAN NOT NULL DEFAULT FALSE,
    isenabled BOOLEAN NOT NULL DEFAULT TRUE,
    validationrules TEXT,
    options TEXT,
    description TEXT,
    sortorder INT DEFAULT 0,

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (reportid) REFERENCES reports(id) ON DELETE CASCADE
);

CREATE INDEX idx_reportparameters_reportid ON reportparameters(reportid);
CREATE INDEX idx_reportparameters_sortorder ON reportparameters(sortorder);

CREATE TRIGGER update_reportparameters_modifiedon BEFORE UPDATE ON reportparameters
    FOR EACH ROW EXECUTE FUNCTION update_modifiedon_column();

-- =====================================================
-- Table: reportexecutions
-- Purpose: Store report execution history
-- =====================================================

CREATE TABLE reportexecutions (
    id VARCHAR(36) PRIMARY KEY,
    reportid VARCHAR(36) NOT NULL,
    executedby VARCHAR(36),
    executionstatus VARCHAR(20) DEFAULT 'pending',  -- 'pending', 'running', 'success', 'failed'
    executiontimems INT,
    errormessage TEXT,
    parameters JSONB,
    outputformat VARCHAR(20),
    outputsizebytes BIGINT,
    outputpath VARCHAR(500),
    rowcount INT,

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (reportid) REFERENCES reports(id) ON DELETE CASCADE
);

CREATE INDEX idx_reportexecutions_reportid ON reportexecutions(reportid);
CREATE INDEX idx_reportexecutions_status ON reportexecutions(executionstatus);
CREATE INDEX idx_reportexecutions_executedby ON reportexecutions(executedby);
CREATE INDEX idx_reportexecutions_createdon ON reportexecutions(createdon);

CREATE TRIGGER update_reportexecutions_modifiedon BEFORE UPDATE ON reportexecutions
    FOR EACH ROW EXECUTE FUNCTION update_modifiedon_column();

-- =====================================================
-- Table: reportshares
-- Purpose: Store report sharing permissions
-- =====================================================

CREATE TABLE reportshares (
    id VARCHAR(36) PRIMARY KEY,
    reportid VARCHAR(36) NOT NULL,
    sharedby VARCHAR(36),
    sharedwith VARCHAR(36),
    canview BOOLEAN NOT NULL DEFAULT TRUE,
    canedit BOOLEAN NOT NULL DEFAULT FALSE,
    canexecute BOOLEAN NOT NULL DEFAULT TRUE,
    canshare BOOLEAN NOT NULL DEFAULT FALSE,
    sharetoken VARCHAR(255) UNIQUE,
    expiresat TIMESTAMP,

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (reportid) REFERENCES reports(id) ON DELETE CASCADE
);

CREATE INDEX idx_reportshares_reportid ON reportshares(reportid);
CREATE INDEX idx_reportshares_sharedwith ON reportshares(sharedwith);
CREATE INDEX idx_reportshares_sharetoken ON reportshares(sharetoken);

CREATE TRIGGER update_reportshares_modifiedon BEFORE UPDATE ON reportshares
    FOR EACH ROW EXECUTE FUNCTION update_modifiedon_column();

-- =====================================================
-- Table: reporttemplates
-- Purpose: Store pre-built report templates
-- =====================================================

CREATE TABLE reporttemplates (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    templateconfig JSONB,
    previewimage VARCHAR(500),
    usagecount INT DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0.00,
    aicompatible BOOLEAN NOT NULL DEFAULT FALSE,
    aitags JSONB,
    suggestedusecases JSONB,
    ispublic BOOLEAN NOT NULL DEFAULT TRUE,
    issystem BOOLEAN NOT NULL DEFAULT FALSE,

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1
);

CREATE INDEX idx_reporttemplates_category ON reporttemplates(category);
CREATE INDEX idx_reporttemplates_ispublic ON reporttemplates(ispublic);
CREATE INDEX idx_reporttemplates_usagecount ON reporttemplates(usagecount);
CREATE INDEX idx_reporttemplates_rating ON reporttemplates(rating);

CREATE TRIGGER update_reporttemplates_modifiedon BEFORE UPDATE ON reporttemplates
    FOR EACH ROW EXECUTE FUNCTION update_modifiedon_column();

-- =====================================================
-- Table: reportschedules
-- Purpose: Store scheduled report executions
-- =====================================================

CREATE TABLE reportschedules (
    id VARCHAR(36) PRIMARY KEY,
    reportid VARCHAR(36) NOT NULL,
    schedulename VARCHAR(255),
    cronexpression VARCHAR(100) NOT NULL,
    timezone VARCHAR(50) DEFAULT 'UTC',
    outputformat VARCHAR(20) DEFAULT 'pdf',
    deliverymethod VARCHAR(20) DEFAULT 'email',
    deliveryconfig JSONB,
    parameters JSONB,
    lastrunat TIMESTAMP,
    nextrunat TIMESTAMP,

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (reportid) REFERENCES reports(id) ON DELETE CASCADE
);

CREATE INDEX idx_reportschedules_reportid ON reportschedules(reportid);
CREATE INDEX idx_reportschedules_nextrunat ON reportschedules(nextrunat);
CREATE INDEX idx_reportschedules_active ON reportschedules(active);

CREATE TRIGGER update_reportschedules_modifiedon BEFORE UPDATE ON reportschedules
    FOR EACH ROW EXECUTE FUNCTION update_modifiedon_column();
