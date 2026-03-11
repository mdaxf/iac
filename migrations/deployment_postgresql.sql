-- IAC Deployment Manager — Full PostgreSQL Migration
-- Run this once to create all deployment-related tables.
-- All statements use CREATE TABLE IF NOT EXISTS so it is safe to re-run.
-- Run date: 2026-03-07

-- ─── iacpackages ─────────────────────────────────────────────────────────────
-- Stores packed packages (database snapshot + document collections).

CREATE TABLE IF NOT EXISTS iacpackages (
    id               VARCHAR(50)  NOT NULL PRIMARY KEY,
    name             VARCHAR(255) NOT NULL,
    version          VARCHAR(50)  NOT NULL,
    packagetype      VARCHAR(20)  NOT NULL,   -- 'database' | 'document'
    description      TEXT,
    metadata         TEXT,                    -- JSON
    packagedata      TEXT         NOT NULL,   -- JSON-serialised package content
    databasetype     VARCHAR(50),
    databasename     VARCHAR(255),
    includeparent    BOOLEAN      NOT NULL DEFAULT FALSE,
    dependencies     TEXT,                    -- JSON array of dependent package IDs
    checksum         VARCHAR(64)  NOT NULL,
    filesize         BIGINT,
    status           VARCHAR(20)  NOT NULL DEFAULT 'active',  -- active | archived | deleted
    tags             TEXT,                    -- JSON array
    environment      VARCHAR(50),
    -- IAC Standard Fields
    active           BOOLEAN      NOT NULL DEFAULT TRUE,
    referenceid      VARCHAR(255),
    createdby        VARCHAR(255),
    createdon        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby       VARCHAR(255),
    modifiedon       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp  INTEGER      NOT NULL DEFAULT 1,
    CONSTRAINT unique_iacpackages_name_version UNIQUE (name, version)
);

CREATE INDEX IF NOT EXISTS idx_iacpackages_name        ON iacpackages (name);
CREATE INDEX IF NOT EXISTS idx_iacpackages_type        ON iacpackages (packagetype);
CREATE INDEX IF NOT EXISTS idx_iacpackages_status      ON iacpackages (status);
CREATE INDEX IF NOT EXISTS idx_iacpackages_createdon   ON iacpackages (createdon);
CREATE INDEX IF NOT EXISTS idx_iacpackages_environment ON iacpackages (environment);
CREATE INDEX IF NOT EXISTS idx_iacpackages_active      ON iacpackages (active);

-- ─── packageactions ──────────────────────────────────────────────────────────
-- Records every pack / deploy / rollback / import operation.

CREATE TABLE IF NOT EXISTS packageactions (
    id                   VARCHAR(50)  NOT NULL PRIMARY KEY,
    packageid            VARCHAR(50)  NOT NULL,
    actiontype           VARCHAR(20)  NOT NULL,   -- pack | deploy | rollback | export | import | validate
    actionstatus         VARCHAR(20)  NOT NULL,   -- pending | in_progress | completed | failed | rolled_back
    targetdatabase       VARCHAR(255),
    targetenvironment    VARCHAR(50),
    sourceenvironment    VARCHAR(50),
    performedat          TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    performedby          VARCHAR(100) NOT NULL,
    startedat            TIMESTAMP,
    completedat          TIMESTAMP,
    durationseconds      INTEGER      NOT NULL DEFAULT 0,
    options              TEXT,                    -- JSON
    resultdata           TEXT,                    -- JSON
    errorlog             TEXT,                    -- JSON / plain text
    warninglog           TEXT,                    -- JSON / plain text
    metadata             TEXT,                    -- JSON
    recordsprocessed     INTEGER      NOT NULL DEFAULT 0,
    recordssucceeded     INTEGER      NOT NULL DEFAULT 0,
    recordsfailed        INTEGER      NOT NULL DEFAULT 0,
    tablesprocessed      INTEGER      NOT NULL DEFAULT 0,
    collectionsprocessed INTEGER      NOT NULL DEFAULT 0,
    -- IAC Standard Fields
    active               BOOLEAN      NOT NULL DEFAULT TRUE,
    referenceid          VARCHAR(255),
    createdby            VARCHAR(255),
    createdon            TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby           VARCHAR(255),
    modifiedon           TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp      INTEGER      NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_packageactions_packageid    ON packageactions (packageid);
CREATE INDEX IF NOT EXISTS idx_packageactions_type         ON packageactions (actiontype);
CREATE INDEX IF NOT EXISTS idx_packageactions_status       ON packageactions (actionstatus);
CREATE INDEX IF NOT EXISTS idx_packageactions_performedat  ON packageactions (performedat);
CREATE INDEX IF NOT EXISTS idx_packageactions_active       ON packageactions (active);

-- ─── packagedeployments ──────────────────────────────────────────────────────
-- Tracks the latest active deployment per package + environment.

CREATE TABLE IF NOT EXISTS packagedeployments (
    id              VARCHAR(50)  NOT NULL PRIMARY KEY,
    packageid       VARCHAR(50)  NOT NULL,
    actionid        VARCHAR(50)  NOT NULL,
    environment     VARCHAR(50)  NOT NULL,
    databasename    VARCHAR(255) NOT NULL,
    deployedat      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deployedby      VARCHAR(100) NOT NULL,
    isactive        BOOLEAN      NOT NULL DEFAULT TRUE,
    rolledbackat    TIMESTAMP,
    rolledbackby    VARCHAR(100),
    -- IAC Standard Fields
    active          BOOLEAN      NOT NULL DEFAULT TRUE,
    referenceid     VARCHAR(255),
    createdby       VARCHAR(255),
    createdon       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby      VARCHAR(255),
    modifiedon      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INTEGER      NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_packagedeployments_packageid ON packagedeployments (packageid);
CREATE INDEX IF NOT EXISTS idx_packagedeployments_env       ON packagedeployments (environment);
CREATE INDEX IF NOT EXISTS idx_packagedeployments_isactive  ON packagedeployments (isactive);
CREATE INDEX IF NOT EXISTS idx_packagedeployments_active    ON packagedeployments (active);

-- ─── iacpackagedefs ──────────────────────────────────────────────────────────
-- Stable SQL record per package definition (one row per definition).
-- Full config is stored in MongoDB (package_definitions collection).

CREATE TABLE IF NOT EXISTS iacpackagedefs (
    id               VARCHAR(36)  NOT NULL PRIMARY KEY,
    name             VARCHAR(255) NOT NULL,
    packagetype      VARCHAR(50)  NOT NULL,
    description      TEXT,
    environment      VARCHAR(50),
    latestversion    INTEGER      NOT NULL DEFAULT 1,
    active           BOOLEAN      NOT NULL DEFAULT TRUE,
    createdby        VARCHAR(255),
    createdon        TIMESTAMP,
    modifiedby       VARCHAR(255),
    modifiedon       TIMESTAMP,
    rowversionstamp  INTEGER      NOT NULL DEFAULT 1
);

-- ─── iacpackagebuilds ────────────────────────────────────────────────────────
-- One row per pack execution; buildnumber auto-increments per definition.

CREATE TABLE IF NOT EXISTS iacpackagebuilds (
    id               VARCHAR(36)  NOT NULL PRIMARY KEY,
    definitionid     VARCHAR(36)  NOT NULL,
    definitionname   VARCHAR(255),
    defversion       INTEGER      NOT NULL DEFAULT 1,
    buildnumber      INTEGER      NOT NULL,
    packageid        VARCHAR(36),
    packagename      VARCHAR(255),
    packageversion   VARCHAR(100),
    status           VARCHAR(50)  NOT NULL DEFAULT 'completed',
    startedat        TIMESTAMP,
    completedat      TIMESTAMP,
    durationseconds  INTEGER      NOT NULL DEFAULT 0,
    triggeredby      VARCHAR(255),
    errorlog         TEXT,
    metadata         TEXT,
    active           BOOLEAN      NOT NULL DEFAULT TRUE,
    createdby        VARCHAR(255),
    createdon        TIMESTAMP,
    modifiedby       VARCHAR(255),
    modifiedon       TIMESTAMP,
    rowversionstamp  INTEGER      NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_iacpackagebuilds_definitionid  ON iacpackagebuilds (definitionid);
CREATE INDEX IF NOT EXISTS idx_iacpackagebuilds_buildnumber   ON iacpackagebuilds (definitionid, buildnumber);
