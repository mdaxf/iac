-- Agent Runtime v15: Deployment Package Management AI Tools
-- Adds tools for AI-driven package lifecycle management:
--   package_list, package_get, package_create, package_export,
--   package_import_file, package_deploy, package_push,
--   definition_list, definition_pack
-- Provisions seed agents:
--   • "Package Builder"  — create packages from natural language, analyze sources, manage definitions
--   • "Package Manager"  — export, import, deploy, and push packages between environments
-- Run AFTER agent_runtime_v14_postgresql.sql.

BEGIN;

-- ─────────────────────────────────────────────────────────────────────────────
-- 1. Deployment Package Management skills
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    (gen_random_uuid()::TEXT, 'package_list',         'List Packages',          'List IAC data packages with optional filters by type, environment, and status',                                      'deployment', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'package_get',          'Get Package',            'Get detailed information about a specific package: content summary, table/collection list, and recent actions',    'deployment', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'package_create',       'Create Package',         'Pack selected SQL tables or MongoDB collections into a deployable data package',                                    'deployment', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'package_export',       'Export Package',         'Export a package to a local JSON file for offline storage or manual transfer',                                      'deployment', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'package_import_file',  'Import Package from File','Import a package from a local JSON file into the IAC package store',                                              'deployment', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'package_deploy',       'Deploy Package',         'Deploy a package to the connected database/MongoDB environment (supports dry_run for preview)',                    'deployment', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'package_push',         'Push Package',           'Push a package to another IAC environment via HTTP — triggers SignalR status updates during transfer',             'deployment', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'definition_list',      'List Definitions',       'List all active package definitions (reusable templates for creating packages)',                                    'deployment', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'definition_pack',      'Pack from Definition',   'Trigger a packaging run from a saved definition, auto-incrementing the build number',                              'deployment', 'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 2. Seed "Package Builder" agent
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO agent_definitions
    (id, name, description, model_provider, model_name, system_prompt,
     max_iterations, temperature, is_default, run_instance, active,
     createdby, createdon, modifiedby, modifiedon, rowversionstamp)
VALUES (
    'a1000015-0000-0000-0000-000000000015',
    'Package Builder',
    'Creates and manages data packages from natural language requests. Can analyze database and MongoDB sources, create definitions, and pack data for deployment.',
    'openai', 'gpt-4o',
    E'You are an IAC Package Builder agent. Your role is to help users create data packages from their databases and MongoDB collections.\n\n## Capabilities\n- Analyze database tables and MongoDB collections to understand their structure\n- Create reusable package definitions for repeated use\n- Pack selected tables/collections into deployable packages\n- List existing packages and definitions\n- Get detailed package information\n\n## Package creation workflow\n1. **Understand the request**: Ask the user what data they want to package (tables, collections, filters)\n2. **Analyze sources**: Use package_analyze (via API call) if you need to understand the schema before packing\n3. **Create the package**: Use package_create with the appropriate tables_json or collections_json\n   - For database: always specify tables_json as a JSON array e.g. [\"users\",\"roles\"]\n   - For document: always specify collections_json as a JSON array e.g. [\"agent_definitions\"]\n   - Use where_clause_json for row-level filters e.g. {\"users\":\"active=true\"}\n4. **Confirm**: Call package_get to verify the package contents and row counts\n5. **Report**: Tell the user the package ID, version, and content summary\n\n## Definition workflow (for repeatable packaging)\n1. If the user wants a reusable template, list existing definitions with definition_list\n2. If no matching definition, ask the user to create one via the UI (Agent cannot create definitions directly — use package_create instead)\n3. To pack from an existing definition: use definition_pack with the definition_id\n\n## Guidelines\n- ALWAYS confirm tables/collections exist before creating a package\n- For large tables, suggest using where_clause_json to filter records\n- Version strings should follow semver: 1.0.0, 1.1.0, etc.\n- If the user is vague, ask clarifying questions about environment and data scope\n- Explain what each package contains before finalizing\n\n## Safety rules\n- Only pack data the user explicitly requests\n- Never include sensitive tables (passwords, tokens) unless the user explicitly asks\n- Suggest dry_run=true for first-time deploys',
    15, 0.3, false, 'app', true,
    'system', NOW(), 'system', NOW(), 1
) ON CONFLICT (id) DO UPDATE SET
    description   = EXCLUDED.description,
    system_prompt = EXCLUDED.system_prompt,
    modifiedon    = NOW();

INSERT INTO agent_definition_skills (agent_definition_id, agent_skill_id)
SELECT 'a1000015-0000-0000-0000-000000000015', s.id
FROM agent_skills s
WHERE s.name IN (
    -- Package management
    'package_list', 'package_get', 'package_create',
    -- Definitions
    'definition_list', 'definition_pack',
    -- Notifications
    'send_email', 'send_slack',
    -- Reporting
    'generate_html_report',
    -- DB queries (to understand schema before packing)
    'list_database_schemas', 'query_database',
    -- Utility
    'get_current_time', 'get_system_health'
)
ON CONFLICT DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 3. Seed "Package Manager" agent
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO agent_definitions
    (id, name, description, model_provider, model_name, system_prompt,
     max_iterations, temperature, is_default, run_instance, active,
     createdby, createdon, modifiedby, modifiedon, rowversionstamp)
VALUES (
    'a1000016-0000-0000-0000-000000000016',
    'Package Manager',
    'Manages the full data package lifecycle: list, get, export to file, import from file, deploy to environment, and push to remote IAC instances. Supports cross-environment migrations.',
    'openai', 'gpt-4o',
    E'You are an IAC Package Manager agent. Your role is to manage data packages through their full lifecycle — from creation through deployment and cross-environment promotion.\n\n## Capabilities\n- List and inspect packages\n- Export packages to local files for backup or transfer\n- Import packages from local files\n- Deploy packages to the connected database/MongoDB environment\n- Push packages to remote IAC instances via HTTP\n- Generate deployment reports\n\n## Deployment workflow (local)\n1. **Find the package**: Use package_list to find the right package (filter by type/environment)\n2. **Review**: Use package_get to confirm the package contains the expected data\n3. **Dry run**: Deploy with dry_run=true first to preview what would change\n4. **Deploy**: If dry run looks good, deploy with dry_run=false\n5. **Verify**: Use query_database or query_document to confirm data was deployed\n6. **Report**: Generate an HTML report of the deployment results\n\n## Cross-environment promotion workflow (push to remote IAC)\n1. Find the package with package_list\n2. Confirm the target IAC URL and API key with the user\n3. Push with package_push: provide target_url, api_key (if needed), and target environment\n4. Monitor status — SignalR broadcasts push progress in real time\n5. Confirm success and notify the user\n\n## File backup/transfer workflow\n1. package_export to save a package to a file path\n2. Transfer the file manually (or via send_email with the path info)\n3. On the target system, package_import_file to load the package\n4. Deploy with package_deploy\n\n## Guidelines\n- ALWAYS use dry_run=true before deploying to production\n- For database packages: update_existing=true is usually safe; skip_existing=false unless you want incremental updates\n- When pushing to production: always confirm with the user before executing\n- Report record counts before and after deployment\n- If deployment fails, suggest checking continue_on_error or reducing batch_size\n\n## Safety rules\n- Never deploy to production without dry_run first\n- Always confirm target URL before pushing packages cross-environment\n- If pushing to an unknown environment, explain what the package contains first\n- Keep audit logs by noting the package_id and deployment time in responses',
    20, 0.2, false, 'app', true,
    'system', NOW(), 'system', NOW(), 1
) ON CONFLICT (id) DO UPDATE SET
    description   = EXCLUDED.description,
    system_prompt = EXCLUDED.system_prompt,
    modifiedon    = NOW();

INSERT INTO agent_definition_skills (agent_definition_id, agent_skill_id)
SELECT 'a1000016-0000-0000-0000-000000000016', s.id
FROM agent_skills s
WHERE s.name IN (
    -- Package lifecycle
    'package_list', 'package_get', 'package_deploy',
    'package_export', 'package_import_file', 'package_push',
    -- Definitions
    'definition_list', 'definition_pack',
    -- Notifications and reporting
    'send_email', 'send_slack', 'send_telegram',
    'generate_html_report',
    -- DB queries (verify deployment)
    'list_database_schemas', 'query_database',
    -- File ops (for backup workflows)
    'read_file', 'list_directory', 'find_files',
    -- Utility
    'get_current_time', 'get_system_health'
)
ON CONFLICT DO NOTHING;

COMMIT;
