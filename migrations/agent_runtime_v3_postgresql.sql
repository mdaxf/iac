-- Agent Runtime V3 Migration for PostgreSQL
-- Adds: enabled/run_instance to agent_definitions, mcp_servers catalog, agent_mcp_servers join table

-- Feature 2: enabled flag and run_instance on agents
ALTER TABLE agent_definitions
  ADD COLUMN IF NOT EXISTS enabled BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS run_instance VARCHAR(50) NOT NULL DEFAULT 'main';

-- Feature 4: MCP server catalog
CREATE TABLE IF NOT EXISTS mcp_servers (
  id              VARCHAR(36)  PRIMARY KEY,
  name            VARCHAR(100) NOT NULL,
  description     TEXT,
  transport_type  VARCHAR(20)  NOT NULL DEFAULT 'http',
  url             VARCHAR(500),
  command         VARCHAR(500),
  args            JSONB        NOT NULL DEFAULT '[]',
  headers         JSONB        NOT NULL DEFAULT '{}',
  enabled         BOOLEAN      NOT NULL DEFAULT TRUE,
  active          BOOLEAN      NOT NULL DEFAULT TRUE,
  createdby       VARCHAR(45),
  createdon       TIMESTAMP    DEFAULT NOW(),
  modifiedby      VARCHAR(45),
  modifiedon      TIMESTAMP    DEFAULT NOW(),
  rowversionstamp INT          NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_mcp_servers_active ON mcp_servers(active) WHERE active = TRUE;

-- Feature 4: many-to-many agent <-> MCP server
CREATE TABLE IF NOT EXISTS agent_mcp_servers (
  agent_definition_id VARCHAR(36) NOT NULL REFERENCES agent_definitions(id) ON DELETE CASCADE,
  mcp_server_id       VARCHAR(36) NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
  PRIMARY KEY (agent_definition_id, mcp_server_id)
);

CREATE INDEX IF NOT EXISTS idx_agent_mcp_servers_agent ON agent_mcp_servers(agent_definition_id);
CREATE INDEX IF NOT EXISTS idx_agent_mcp_servers_mcp   ON agent_mcp_servers(mcp_server_id);
