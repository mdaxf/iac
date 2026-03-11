-- Agent Runtime V2 Migration for PostgreSQL
-- Replaces system_user_id and tools_enabled JSONB with proper agent_skills catalog
-- and agent_definition_skills many-to-many join table.

-- 1. Agent skills catalog (one row per available tool/skill)
CREATE TABLE IF NOT EXISTS agent_skills (
    id              VARCHAR(36)     NOT NULL PRIMARY KEY,
    name            VARCHAR(100)    NOT NULL UNIQUE,   -- tool identifier used by the runner
    display_name    VARCHAR(100)    NOT NULL,
    description     TEXT            NULL,
    category        VARCHAR(50)     NULL,              -- e.g. 'search', 'data', 'system'
    active          BOOLEAN         NOT NULL DEFAULT TRUE,
    createdby       VARCHAR(45)     NULL,
    createdon       TIMESTAMP       DEFAULT CURRENT_TIMESTAMP,
    modifiedby      VARCHAR(45)     NULL,
    modifiedon      TIMESTAMP       DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT             NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_agent_skills_name     ON agent_skills(name);
CREATE INDEX IF NOT EXISTS idx_agent_skills_category ON agent_skills(category) WHERE active = TRUE;

-- 2. Many-to-many: agent definitions <-> skills
CREATE TABLE IF NOT EXISTS agent_definition_skills (
    agent_definition_id VARCHAR(36) NOT NULL REFERENCES agent_definitions(id) ON DELETE CASCADE,
    agent_skill_id      VARCHAR(36) NOT NULL REFERENCES agent_skills(id)      ON DELETE CASCADE,
    PRIMARY KEY (agent_definition_id, agent_skill_id)
);

CREATE INDEX IF NOT EXISTS idx_agent_def_skills_agent ON agent_definition_skills(agent_definition_id);
CREATE INDEX IF NOT EXISTS idx_agent_def_skills_skill ON agent_definition_skills(agent_skill_id);

-- 3. Migrate existing tools_enabled JSONB data into the new tables
--    Insert distinct skill names from all agent_definitions into agent_skills catalog
INSERT INTO agent_skills (id, name, display_name, category, createdby, createdon, modifiedby, modifiedon)
SELECT
    gen_random_uuid()::TEXT,
    skill_name,
    skill_name,   -- display_name defaults to name; update manually later
    'general',
    'system',
    NOW(),
    'system',
    NOW()
FROM (
    SELECT DISTINCT jsonb_array_elements_text(tools_enabled) AS skill_name
    FROM agent_definitions
    WHERE tools_enabled IS NOT NULL AND jsonb_array_length(tools_enabled) > 0
) t
ON CONFLICT (name) DO NOTHING;

--    Populate the join table from existing tools_enabled data
INSERT INTO agent_definition_skills (agent_definition_id, agent_skill_id)
SELECT
    ad.id,
    ags.id
FROM agent_definitions ad,
     jsonb_array_elements_text(ad.tools_enabled) AS tool_name
JOIN agent_skills ags ON ags.name = tool_name
WHERE ad.tools_enabled IS NOT NULL AND jsonb_array_length(ad.tools_enabled) > 0
ON CONFLICT DO NOTHING;

-- 4. Drop superseded columns from agent_definitions
ALTER TABLE agent_definitions DROP COLUMN IF EXISTS system_user_id;
ALTER TABLE agent_definitions DROP COLUMN IF EXISTS tools_enabled;

-- 5. Remove the system-user infrastructure that is no longer needed
DROP INDEX IF EXISTS idx_agent_definitions_system_user_id;
