-- IAC Reports Schema
-- Supports MySQL, PostgreSQL, MSSQL
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
    layoutconfig JSON,  -- Layout configuration
    pagesettings JSON,  -- Page settings
    aiprompt TEXT,  -- AI generation prompt
    aianalysis JSON,  -- AI analysis results
    templatesourceid VARCHAR(36),  -- Source template if created from template
    tags JSON,  -- Tags array
    version INT NOT NULL DEFAULT 1,
    lastexecutedon DATETIME,

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    INDEX idx_reporttype (reporttype),
    INDEX idx_ispublic (ispublic),
    INDEX idx_istemplate (istemplate),
    INDEX idx_createdby (createdby),
    INDEX idx_createdon (createdon)
);

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
    selectedtables JSON,  -- Array of selected tables
    selectedfields JSON,  -- Array of selected fields with aliases
    joins JSON,  -- Array of join definitions
    filters JSON,  -- Array of filter conditions
    sorting JSON,  -- Array of sort definitions
    grouping JSON,  -- Array of group by fields
    parameters JSON,  -- Query parameters

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (reportid) REFERENCES reports(id) ON DELETE CASCADE,
    INDEX idx_reportid (reportid),
    INDEX idx_alias (alias),
    INDEX idx_querytype (querytype)
);

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
    dataconfig JSON,
    componentconfig JSON,
    styleconfig JSON,
    charttype VARCHAR(20),  -- Chart types if applicable
    chartconfig JSON,
    barcodetype VARCHAR(20),  -- Barcode types if applicable
    barcodeconfig JSON,
    drilldownconfig JSON,
    conditionalformatting JSON,
    isvisible BOOLEAN NOT NULL DEFAULT TRUE,

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (reportid) REFERENCES reports(id) ON DELETE CASCADE,
    INDEX idx_reportid (reportid),
    INDEX idx_componenttype (componenttype),
    INDEX idx_zindex (zindex)
);

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
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (reportid) REFERENCES reports(id) ON DELETE CASCADE,
    INDEX idx_reportid (reportid),
    INDEX idx_sortorder (sortorder)
);

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
    parameters JSON,
    outputformat VARCHAR(20),
    outputsizebytes BIGINT,
    outputpath VARCHAR(500),
    rowcount INT,

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (reportid) REFERENCES reports(id) ON DELETE CASCADE,
    INDEX idx_reportid (reportid),
    INDEX idx_executionstatus (executionstatus),
    INDEX idx_executedby (executedby),
    INDEX idx_createdon (createdon)
);

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
    expiresat DATETIME,

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (reportid) REFERENCES reports(id) ON DELETE CASCADE,
    INDEX idx_reportid (reportid),
    INDEX idx_sharedwith (sharedwith),
    INDEX idx_sharetoken (sharetoken)
);

-- =====================================================
-- Table: reporttemplates
-- Purpose: Store pre-built report templates
-- =====================================================

CREATE TABLE reporttemplates (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    templateconfig JSON,
    previewimage VARCHAR(500),
    usagecount INT DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0.00,
    aicompatible BOOLEAN NOT NULL DEFAULT FALSE,
    aitags JSON,
    suggestedusecases JSON,
    ispublic BOOLEAN NOT NULL DEFAULT TRUE,
    issystem BOOLEAN NOT NULL DEFAULT FALSE,

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    INDEX idx_category (category),
    INDEX idx_ispublic (ispublic),
    INDEX idx_usagecount (usagecount),
    INDEX idx_rating (rating)
);

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
    deliveryconfig JSON,
    parameters JSON,
    lastrunat DATETIME,
    nextrunat DATETIME,

    -- Standard IAC audit fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (reportid) REFERENCES reports(id) ON DELETE CASCADE,
    INDEX idx_reportid (reportid),
    INDEX idx_nextrunat (nextrunat),
    INDEX idx_active (active)
);
