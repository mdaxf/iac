-- Agent Runtime v9: Align run_instance values with system InstanceRole constants
-- Renames legacy values:
--   'main' → 'app'          (InstanceRoleApp)
--   'jobs' → 'job_executor' (InstanceRoleJobExecutor)
-- 'integration_hub' is already correct (InstanceRoleIntegrationHub).
-- Run AFTER agent_runtime_v8_postgresql.sql.

-- ─────────────────────────────────────────────────────────────────────────────
-- 1. Migrate existing rows
-- ─────────────────────────────────────────────────────────────────────────────
UPDATE agent_definitions SET run_instance = 'app'          WHERE run_instance = 'main';
UPDATE agent_definitions SET run_instance = 'job_executor' WHERE run_instance = 'jobs';

-- ─────────────────────────────────────────────────────────────────────────────
-- 2. Update column default so new rows created outside the API also get 'app'
-- ─────────────────────────────────────────────────────────────────────────────
ALTER TABLE agent_definitions ALTER COLUMN run_instance SET DEFAULT 'app';
