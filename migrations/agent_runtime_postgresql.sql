-- Agent Runtime Migration for PostgreSQL
-- Extends IAC with native agent/gateway capability

-- 1. User types lookup table (identifies agent vs human users)
CREATE TABLE IF NOT EXISTS user_types (
    id          SERIAL          PRIMARY KEY,
    name        VARCHAR(50)     NOT NULL UNIQUE,
    description TEXT            NULL,
    active      BOOLEAN         NOT NULL DEFAULT TRUE
);

-- Seed the agent user type
INSERT INTO user_types (name, description)
VALUES ('agent', 'AI agent system user')
ON CONFLICT (name) DO NOTHING;

-- Extend users table: typeid FK to user_types, externallogin stores agent reference
ALTER TABLE users ADD COLUMN IF NOT EXISTS typeid       INT  NULL REFERENCES user_types(id);
ALTER TABLE users ADD COLUMN IF NOT EXISTS externallogin VARCHAR(255) NULL;

CREATE INDEX IF NOT EXISTS idx_users_typeid        ON users(typeid);
CREATE INDEX IF NOT EXISTS idx_users_externallogin ON users(externallogin);

-- 2. Extend conversations table with agent association
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS agent_definition_id VARCHAR(36) NULL;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS conversation_type VARCHAR(20) NOT NULL DEFAULT 'database';

CREATE INDEX IF NOT EXISTS idx_conversations_agent_definition_id ON conversations(agent_definition_id);
CREATE INDEX IF NOT EXISTS idx_conversations_conversation_type ON conversations(conversation_type);

-- 3. Agent definitions table
CREATE TABLE IF NOT EXISTS agent_definitions (
    id                  VARCHAR(36)     NOT NULL PRIMARY KEY,
    name                VARCHAR(100)    NOT NULL,
    description         TEXT            NULL,
    system_user_id      VARCHAR(36)     NOT NULL,
    avatar              VARCHAR(500)    NULL,
    model_provider      VARCHAR(50)     NOT NULL DEFAULT 'openai',
    model_name          VARCHAR(100)    NOT NULL DEFAULT 'gpt-4o',
    system_prompt       TEXT            NOT NULL,
    tools_enabled       JSONB           NOT NULL DEFAULT '[]',
    max_iterations      INT             NOT NULL DEFAULT 10,
    temperature         DECIMAL(3,2)    NOT NULL DEFAULT 0.70,
    is_default          BOOLEAN         NOT NULL DEFAULT FALSE,
    active              BOOLEAN         NOT NULL DEFAULT TRUE,
    referenceid         VARCHAR(36)     NULL,
    createdby           VARCHAR(45)     NULL,
    createdon           TIMESTAMP       DEFAULT CURRENT_TIMESTAMP,
    modifiedby          VARCHAR(45)     NULL,
    modifiedon          TIMESTAMP       DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp     INT             NOT NULL DEFAULT 1
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_definitions_name ON agent_definitions(name) WHERE active = TRUE;
CREATE INDEX IF NOT EXISTS idx_agent_definitions_system_user_id ON agent_definitions(system_user_id);

-- 4. Agent schedules table
CREATE TABLE IF NOT EXISTS agent_schedules (
    id                  VARCHAR(36)     NOT NULL PRIMARY KEY,
    agent_definition_id VARCHAR(36)     NOT NULL REFERENCES agent_definitions(id),
    name                VARCHAR(100)    NOT NULL,
    description         TEXT            NULL,
    task_prompt         TEXT            NOT NULL,
    cron_expression     VARCHAR(100)    NULL,
    run_once_at         TIMESTAMP       NULL,
    last_run_at         TIMESTAMP       NULL,
    last_run_status     VARCHAR(20)     NULL,
    last_run_output     TEXT            NULL,
    job_id              VARCHAR(36)     NULL,
    enabled             BOOLEAN         NOT NULL DEFAULT TRUE,
    active              BOOLEAN         NOT NULL DEFAULT TRUE,
    referenceid         VARCHAR(36)     NULL,
    createdby           VARCHAR(45)     NULL,
    createdon           TIMESTAMP       DEFAULT CURRENT_TIMESTAMP,
    modifiedby          VARCHAR(45)     NULL,
    modifiedon          TIMESTAMP       DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp     INT             NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_agent_schedules_agent_definition_id ON agent_schedules(agent_definition_id);
CREATE INDEX IF NOT EXISTS idx_agent_schedules_enabled ON agent_schedules(enabled) WHERE active = TRUE;
