-- Agent Runtime v5: agent_memory table (3-layer memory system)
-- Layers: L0 = index/title+brief, L1 = summary, L2 = full content
-- Priority: P0 = critical/permanent, P1 = normal, P2 = low/disposable
-- Retention: permanent | interval | temporary

CREATE TABLE IF NOT EXISTS agent_memory (
  id                  VARCHAR(36)  PRIMARY KEY,
  agent_definition_id VARCHAR(36)  NOT NULL REFERENCES agent_definitions(id) ON DELETE CASCADE,
  title               VARCHAR(200) NOT NULL,
  summary             TEXT,                                        -- L0/L1 summary
  content             TEXT,                                        -- L2 full content
  content_type        VARCHAR(50)  NOT NULL DEFAULT 'text',        -- text | json | markdown
  tags                JSONB        NOT NULL DEFAULT '[]',
  layer               VARCHAR(2)   NOT NULL DEFAULT 'L1',          -- L0 | L1 | L2
  priority            VARCHAR(2)   NOT NULL DEFAULT 'P1',          -- P0 | P1 | P2
  retention_type      VARCHAR(20)  NOT NULL DEFAULT 'permanent',   -- permanent | interval | temporary
  retention_days      INT,
  expires_at          TIMESTAMP,
  access_count        INT          NOT NULL DEFAULT 0,
  last_accessed_on    TIMESTAMP,
  active              BOOLEAN      NOT NULL DEFAULT TRUE,
  archived            BOOLEAN      NOT NULL DEFAULT FALSE,
  createdby           VARCHAR(45),
  createdon           TIMESTAMP    DEFAULT NOW(),
  modifiedby          VARCHAR(45),
  modifiedon          TIMESTAMP    DEFAULT NOW(),
  rowversionstamp     INT          NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_agent_memory_agent    ON agent_memory(agent_definition_id, active, archived);
CREATE INDEX IF NOT EXISTS idx_agent_memory_priority ON agent_memory(agent_definition_id, priority, layer, active);
CREATE INDEX IF NOT EXISTS idx_agent_memory_expires  ON agent_memory(expires_at) WHERE expires_at IS NOT NULL AND archived = FALSE;
