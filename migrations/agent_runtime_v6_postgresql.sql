-- Agent Runtime v6: Entity Generation Skills
-- Adds skills that allow agents to create and update IAC content entities:
-- Report (BI), 3D Model, Plant Model, Work Instruction, Dataset,
-- Workflow, BPM/TranCode, UI View, UI Page, Whiteboard.
-- Run AFTER agent_runtime_v2_postgresql.sql (which creates the agent_skills table).

-- 1. Seed new skill definitions
INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    -- Document-backed content (MongoDB via service layer)
    (gen_random_uuid()::TEXT, 'create_report',           'Create Report',           'Create a BI report with SQL datasource and visual components',       'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'update_report',           'Update Report',           'Update an existing BI report metadata or components',                'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'create_3d_model',         'Create 3D Model',         'Create a 3D model with Three.js scene objects',                      'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'update_3d_model',         'Update 3D Model',         'Update a 3D model or create a new version',                          'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_3d_models',          'List 3D Models',          'Discover existing 3D models in IAC 3D Studio',                       'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'create_plant_model',      'Create Plant Model',      'Create a plant layout with machines, robots, and equipment',          'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'update_plant_model',      'Update Plant Model',      'Update an existing plant layout',                                    'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_plant_models',       'List Plant Models',       'Discover existing plant models in IAC Plant Studio',                  'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'create_work_instruction', 'Create Work Instruction', 'Create a step-by-step work instruction document',                    'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'update_work_instruction', 'Update Work Instruction', 'Update an existing work instruction (steps, status, version)',         'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_work_instructions',  'List Work Instructions',  'Discover existing work instructions in IAC',                         'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'create_dataset',          'Create Dataset',          'Create a dataset schema that powers views and pages',                 'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'update_dataset',          'Update Dataset',          'Update an existing dataset schema (query, fields, version)',           'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_datasets',           'List Datasets',           'Discover existing datasets in IAC',                                  'content', 'system', NOW(), 'system', NOW()),
    -- Direct MongoDB entity tools
    (gen_random_uuid()::TEXT, 'create_workflow',         'Create Workflow',         'Create a workflow with React Flow nodes and edges',                   'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'update_workflow',         'Update Workflow',         'Update workflow nodes, edges, or metadata',                           'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_workflows',          'List Workflows',          'Discover existing workflows in IAC',                                  'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'create_bpm',              'Create BPM',              'Create a BPM/TranCode with function groups and logic',                'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'update_bpm',              'Update BPM',              'Update an existing BPM/TranCode definition',                          'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_bpms',               'List BPMs',               'Discover existing BPM/TranCode definitions in IAC',                  'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'create_view',             'Create View',             'Create a UI View with fields and actions',                            'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'update_view',             'Update View',             'Update a UI View fields, actions, or datasource',                     'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_views',              'List Views',              'Discover existing UI Views in IAC',                                   'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'create_page',             'Create Page',             'Create a UI Page with panel layout referencing views',                'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'update_page',             'Update Page',             'Update a UI Page panels or actions',                                  'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_pages',              'List Pages',              'Discover existing UI Pages in IAC',                                   'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'create_whiteboard',       'Create Whiteboard',       'Create a whiteboard with Excalidraw-compatible diagram elements',     'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'update_whiteboard',       'Update Whiteboard',       'Update whiteboard elements or metadata',                              'content', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_whiteboards',        'List Whiteboards',        'Discover existing whiteboards in IAC',                                'content', 'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- 2. Seed a Content Generator agent that has all generation skills
INSERT INTO agent_definitions
    (id, name, description, model_provider, model_name, system_prompt, max_iterations, temperature, is_default, active, createdby, createdon, modifiedby, modifiedon, rowversionstamp)
VALUES (
    'a1000007-0000-0000-0000-000000000007',
    'Content Generator',
    'Generates and updates IAC content: BI reports, 3D models, plant layouts, work instructions, datasets, workflows, BPM/TranCodes, UI views, UI pages, and whiteboards.',
    'openai', 'gpt-4o',
    E'You are an IAC Content Generator agent. Your role is to generate and update structured content in IAC systems.\n\n## Capabilities\n- Create and update BI reports with SQL datasources and visual components\n- Create and update 3D models with Three.js scene objects\n- Create and update plant layouts with equipment objects\n- Create and update step-by-step work instructions\n- Create and update dataset schemas for views and pages\n- Create and update workflows with React Flow nodes and edges\n- Create and update BPM/TranCode logic with function groups and functions\n- Create and update UI Views with field definitions and actions\n- Create and update UI Pages with panel layouts referencing views\n- Create and update whiteboards with Excalidraw-compatible diagram elements\n\n## Instructions\nWhen generating content:\n1. Use list tools first to discover existing entities (list_workflows, list_bpms, list_views, list_pages, list_whiteboards, list_3d_models, list_plant_models, list_work_instructions, list_datasets, list_reports)\n2. Use query_database with connection=''iac'' to understand the available data before writing SQL\n3. Use list_database_schemas to discover available tables when needed\n4. Generate well-structured, complete content before calling the creation tool\n5. For reports: verify the SQL returns expected data with query_database before creating\n6. For 3D models: generate proper Three.js-compatible objects with correct geometry and material fields\n7. For plant models: use standard assetType values: MACHINE_CNC, ROBOT_ARM, CONVEYOR, SHELF, SENSOR, WORKSTATION\n8. For work instructions: write clear, numbered steps with actionable descriptions\n9. For datasets: ensure the SQL query and field names are correct\n10. For workflows: use React Flow format with proper node types (start, task, gateway, subflow, end) and edge connections\n11. For BPM/TranCode: design function groups sequentially; use query functype for SQL, goexpr for expressions, javascript for scripts\n12. For UI Views: define all required fields with correct types and datasource reference\n13. For UI Pages: reference existing view names in panel definitions; call list_views first\n14. For whiteboards: generate unique element IDs; position elements logically for the diagram type\n15. Always confirm what was created and provide the entity ID in your response',
    25, 0.30, false, true, 'system', NOW(), 'system', NOW(), 1
) ON CONFLICT (id) DO UPDATE SET
    description = EXCLUDED.description,
    system_prompt = EXCLUDED.system_prompt,
    max_iterations = EXCLUDED.max_iterations,
    modifiedon = NOW();

-- 3. Assign all generation + data skills to the Content Generator agent
INSERT INTO agent_definition_skills (agent_definition_id, agent_skill_id)
SELECT 'a1000007-0000-0000-0000-000000000007', s.id
FROM agent_skills s
WHERE s.name IN (
    -- Document content
    'create_report', 'update_report',
    'create_3d_model', 'update_3d_model', 'list_3d_models',
    'create_plant_model', 'update_plant_model', 'list_plant_models',
    'create_work_instruction', 'update_work_instruction', 'list_work_instructions',
    'create_dataset', 'update_dataset', 'list_datasets',
    -- Entity content
    'create_workflow', 'update_workflow', 'list_workflows',
    'create_bpm', 'update_bpm', 'list_bpms',
    'create_view', 'update_view', 'list_views',
    'create_page', 'update_page', 'list_pages',
    'create_whiteboard', 'update_whiteboard', 'list_whiteboards',
    -- Data / web utilities
    'list_database_schemas', 'query_database', 'list_reports',
    'get_current_time', 'web_search', 'web_fetch'
)
ON CONFLICT DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- v6b: System Management + Notification Skills
-- ─────────────────────────────────────────────────────────────────────────────

-- 4. Seed system management and notification skill definitions
INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    -- Configuration
    (gen_random_uuid()::TEXT, 'list_configs',                  'List Configs',                 'List IAC system configurations in ConfigStore',                                 'system',       'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'get_config',                    'Get Config',                   'Get active system configuration by name and environment',                       'system',       'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'update_config',                 'Update Config',                'Update a draft system configuration data',                                      'system',       'system', NOW(), 'system', NOW()),
    -- Integration Hub
    (gen_random_uuid()::TEXT, 'list_integration_hubs',         'List Integration Hubs',        'List integration hub configurations in IAC',                                    'system',       'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'get_integration_hub',           'Get Integration Hub',          'Get full integration hub configuration by ID',                                  'system',       'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'create_integration_hub',        'Create Integration Hub',       'Create a new integration hub configuration',                                    'system',       'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'update_integration_hub_status', 'Update Hub Status',            'Update integration hub status or enabled state',                                'system',       'system', NOW(), 'system', NOW()),
    -- Jobs
    (gen_random_uuid()::TEXT, 'list_scheduled_jobs',           'List Scheduled Jobs',          'List scheduled job definitions with cron and next run time',                    'system',       'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'create_scheduled_job',          'Create Scheduled Job',         'Create a new scheduled job with cron expression or interval',                   'system',       'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'toggle_scheduled_job',          'Toggle Scheduled Job',         'Enable or disable a scheduled job',                                             'system',       'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_queue_jobs',               'List Queue Jobs',              'List execution queue jobs with optional status filter',                          'system',       'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'create_queue_job',              'Create Queue Job',             'Enqueue a TranCode or system handler for immediate execution',                  'system',       'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'get_job_statistics',            'Get Job Statistics',           'Get aggregate statistics for the IAC job system',                               'system',       'system', NOW(), 'system', NOW()),
    -- Notifications
    (gen_random_uuid()::TEXT, 'send_email',                    'Send Email',                   'Send an email via SMTP to report agent results or summaries',                   'notification', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'send_telegram',                 'Send Telegram',                'Send a message to a Telegram chat via Bot API',                                 'notification', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'send_webhook',                  'Send Webhook',                 'POST a JSON payload to any HTTP webhook (Teams, Discord, custom)',              'notification', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'send_slack',                    'Send Slack',                   'Send a message to a Slack channel via Incoming Webhook',                        'notification', 'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- 5. Seed a System Manager agent
INSERT INTO agent_definitions
    (id, name, description, model_provider, model_name, system_prompt, max_iterations, temperature, is_default, active, createdby, createdon, modifiedby, modifiedon, rowversionstamp)
VALUES (
    'a1000008-0000-0000-0000-000000000008',
    'System Manager',
    'Manages IAC system resources: configurations, integration hubs, and jobs. Can send notifications via email, Telegram, Slack, or webhooks.',
    'openai', 'gpt-4o',
    E'You are an IAC System Manager agent. Your role is to monitor, configure, and manage IAC system resources, and to notify stakeholders of important events.\n\n## Capabilities\n- Read and update versioned system configurations (ConfigStore)\n- List and manage Integration Hub connections\n- Create and manage scheduled jobs and queue jobs\n- Send notifications via email, Telegram, Slack, or HTTP webhooks\n- Query IAC data to build status reports\n\n## Instructions\n1. Always use list tools first to discover existing resources before modifying them\n2. For configurations: use get_config to read before update_config to modify; only draft configs can be updated\n3. For integration hubs: use list_integration_hubs to discover, get_integration_hub for details before changes\n4. For jobs: use list_scheduled_jobs to review existing jobs; always validate handler (TranCode) names with list_bpms before creating jobs\n5. For notifications: retrieve credentials from system config using get_config if stored there, or use values provided in the conversation\n6. Format notification messages clearly: what happened, key results/metrics, timestamp, and next steps\n7. For Telegram messages use Markdown formatting; for Slack use mrkdwn; for email prefer HTML for rich formatting\n8. Always confirm what action was taken and provide IDs/details in your response\n9. Never delete or disable resources without explicit confirmation in the conversation',
    15, 0.20, false, true, 'system', NOW(), 'system', NOW(), 1
) ON CONFLICT (id) DO NOTHING;

-- 6. Assign system + notification + data skills to System Manager
INSERT INTO agent_definition_skills (agent_definition_id, agent_skill_id)
SELECT 'a1000008-0000-0000-0000-000000000008', s.id
FROM agent_skills s
WHERE s.name IN (
    'list_configs', 'get_config', 'update_config',
    'list_integration_hubs', 'get_integration_hub', 'create_integration_hub', 'update_integration_hub_status',
    'list_scheduled_jobs', 'create_scheduled_job', 'toggle_scheduled_job',
    'list_queue_jobs', 'create_queue_job', 'get_job_statistics',
    'send_email', 'send_telegram', 'send_webhook', 'send_slack',
    'list_database_schemas', 'query_database', 'list_agents', 'list_reports',
    'get_current_time', 'get_system_health',
    'list_conversations', 'send_message',
    'web_search', 'web_fetch'
)
ON CONFLICT DO NOTHING;
