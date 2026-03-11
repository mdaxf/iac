-- Agent Runtime v11: OS Management, File System, Tool Management, and Testing Skills
-- Adds skills for:
--   • OS service control (list, status, start, stop, restart)
--   • System diagnostics and process inspection
--   • File system operations with automatic backup
--   • Tool/skill catalog introspection and management
--   • Unit, system, and UI test execution
--   • Ad-hoc shell command execution
-- Provisions seed agents:
--   • "System Engineer"  — service control + diagnostics + file ops + notifications
--   • "DevOps Agent"     — full deployment lifecycle (services, files, tests, reports)
--   • "Test Runner"      — automated testing + result reporting
-- Run AFTER agent_runtime_v10_postgresql.sql.

-- ─────────────────────────────────────────────────────────────────────────────
-- 1. OS Service Management skills
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    (gen_random_uuid()::TEXT, 'list_services',       'List Services',        'Enumerate OS services (Windows SC / Linux systemd) with optional name and state filters',                     'devops', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'get_service_status',  'Get Service Status',   'Get current status, state, and details of a named OS service',                                               'devops', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'start_service',       'Start Service',        'Start a stopped OS service by name',                                                                          'devops', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'stop_service',        'Stop Service',         'Stop a running OS service by name',                                                                           'devops', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'restart_service',     'Restart Service',      'Restart an OS service (stop + start on Windows; systemctl restart on Linux)',                                 'devops', 'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 2. System Diagnostics skills
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    (gen_random_uuid()::TEXT, 'diagnose_system',     'Diagnose System',      'Comprehensive health check: OS info, CPU, disk, memory, DB connectivity, running services — returns structured JSON report',  'devops', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'get_process_list',    'Get Process List',     'List running processes filtered by name — uses ps aux (Linux) or tasklist (Windows)',                        'devops', 'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 3. File System skills
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    (gen_random_uuid()::TEXT, 'read_file',           'Read File',            'Read text file content (up to 1 MB; truncates with warning)',                                                 'devops', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_directory',      'List Directory',       'List files and subdirectories with optional recursive traversal and glob filter',                             'devops', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'write_file',          'Write File',           'Create or overwrite a file — always backs up the existing file before writing',                              'devops', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'append_to_file',      'Append to File',       'Append content to the end of a file — backs up the existing file before appending',                         'devops', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'delete_file',         'Delete File',          'Soft-delete a file by renaming to .deleted.{timestamp} — recoverable by renaming back',                     'devops', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'backup_file',         'Backup File',          'Create an explicit backup copy of a file at {path}.bak.{timestamp}',                                        'devops', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'find_files',          'Find Files',           'Search for files by name glob pattern and/or content substring — skips build artifacts and hidden dirs',    'devops', 'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 4. Tool & Skill Management skills
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    (gen_random_uuid()::TEXT, 'list_available_tools',     'List Available Tools',      'List all registered tool handlers in the IAC agent runtime with names and descriptions',         'system',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'search_tools',             'Search Tools',              'Search registered tools by keyword (ranked: exact name > name contains > description contains)', 'system',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'describe_tool',            'Describe Tool',             'Get the full parameter schema and description for a specific tool handler',                       'system',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_agent_skills',        'List Agent Skills',         'List the agent skill catalog (agent_skills table) with optional category and keyword filters',   'system',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'create_agent_skill',       'Create Agent Skill',        'Add a new entry to the agent skill catalog — name must match a registered tool handler',         'system',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'update_agent_skill',       'Update Agent Skill',        'Update display_name, description, category, or active flag of an existing skill catalog entry',  'system',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_agent_assigned_skills','List Agent Assigned Skills','List all skills currently assigned to a specific agent definition',                               'system',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'assign_skill_to_agent',    'Assign Skill to Agent',     'Assign a skill from the catalog to an agent definition',                                          'system',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'remove_skill_from_agent',  'Remove Skill from Agent',   'Remove a skill assignment from an agent definition',                                              'system',  'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 5. Testing skills
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    (gen_random_uuid()::TEXT, 'run_unit_tests',    'Run Unit Tests',    'Run Go unit tests (go test ./...) in a working directory with optional pattern filter',              'testing', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'run_system_tests',  'Run System Tests',  'HTTP health checks against API endpoints — verifies status codes, latency, and response content',  'testing', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'run_ui_tests',      'Run UI Tests',      'Run Playwright UI test suite (npx playwright test) with browser and reporter options',              'testing', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'run_shell_command', 'Run Shell Command', 'Execute an OS command with args (no shell interpolation) — for npm, git, make, pip, etc.',          'devops',  'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 6. Seed "System Engineer" agent
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_definitions
    (id, name, description, model_provider, model_name, system_prompt,
     max_iterations, temperature, is_default, run_instance, active,
     createdby, createdon, modifiedby, modifiedon, rowversionstamp)
VALUES (
    'a1000012-0000-0000-0000-000000000012',
    'System Engineer',
    'Manages OS services, reads/modifies files (with auto-backup), diagnoses system health, and sends reports via email or Slack.',
    'openai', 'gpt-4o',
    E'You are an IAC System Engineer agent. You manage the IAC host system: services, files, diagnostics, and health reporting.\n\n## Capabilities\n- List, start, stop, and restart OS services (Windows SC or Linux systemd)\n- Read and modify configuration files (every write auto-creates a .bak backup)\n- Run comprehensive system diagnostics and send reports\n- Search for files by name pattern or content\n- Notify stakeholders via email, Slack, or Telegram\n\n## Service management workflow\n1. Use list_services or get_service_status to find the exact service name and current state\n2. Confirm the action with the user before stopping or restarting a production service\n3. After start/stop/restart, call get_service_status to verify the new state\n4. If a service fails to start, check logs with run_shell_command (journalctl -u {name} --no-pager -n 50 on Linux)\n\n## File modification workflow\n1. Use read_file to inspect the current content\n2. Call backup_file for extra safety before any critical change\n3. Use write_file or append_to_file — both auto-backup the existing file\n4. Confirm the change was applied correctly by reading the file again\n\n## Diagnostics workflow\n1. Call diagnose_system to get OS/disk/memory/DB health\n2. If degraded: investigate services and logs\n3. Generate an HTML report with generate_html_report\n4. Send via send_email (content_type: html) or summarize via send_slack\n\n## Safety rules\n- NEVER delete files permanently — use delete_file (which renames to .deleted.{timestamp})\n- NEVER modify /etc/passwd, /etc/shadow, system32 paths, or any blocked system path\n- Always confirm before stopping a running service\n- Keep the user informed of every action taken',
    20, 0.25, false, 'app', true,
    'system', NOW(), 'system', NOW(), 1
) ON CONFLICT (id) DO UPDATE SET
    description   = EXCLUDED.description,
    system_prompt = EXCLUDED.system_prompt,
    modifiedon    = NOW();

INSERT INTO agent_definition_skills (agent_definition_id, agent_skill_id)
SELECT 'a1000012-0000-0000-0000-000000000012', s.id
FROM agent_skills s
WHERE s.name IN (
    -- Service control
    'list_services', 'get_service_status', 'start_service', 'stop_service', 'restart_service',
    -- Diagnostics
    'diagnose_system', 'get_process_list',
    -- File ops
    'read_file', 'list_directory', 'write_file', 'append_to_file', 'backup_file', 'find_files',
    -- Notifications
    'send_email', 'send_slack', 'send_telegram', 'send_webhook',
    -- Reporting
    'generate_html_report',
    -- Utility
    'get_current_time', 'run_shell_command'
)
ON CONFLICT DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 7. Seed "DevOps Agent"
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_definitions
    (id, name, description, model_provider, model_name, system_prompt,
     max_iterations, temperature, is_default, run_instance, active,
     createdby, createdon, modifiedby, modifiedon, rowversionstamp)
VALUES (
    'a1000013-0000-0000-0000-000000000013',
    'DevOps Agent',
    'Manages the full IAC deployment lifecycle: service control, configuration updates, test execution, diagnostics, and deployment notifications.',
    'openai', 'gpt-4o',
    E'You are an IAC DevOps agent. Your role is to manage the complete deployment lifecycle for IAC services.\n\n## Capabilities\n- OS service management (list, start, stop, restart)\n- Configuration file management (read, backup, write)\n- System and database health diagnostics\n- Unit, system, and UI test execution\n- Shell command execution for build tools (npm, go, make, git)\n- Agent skill and tool catalog management\n- Deployment status reporting via email/Slack\n\n## Deployment workflow\n1. **Pre-check**: diagnose_system to confirm baseline health\n2. **Stop services**: stop_service for each affected IAC service\n3. **Update config**: read_file → write_file (auto-backup created)\n4. **Start services**: start_service → verify with get_service_status\n5. **Run tests**: run_unit_tests (backend) → run_system_tests (API health) → run_ui_tests (optional)\n6. **Report**: generate_html_report with test results + system metrics → send_email\n\n## Config update best practice\n1. Always read_file first to understand the current configuration\n2. Call backup_file for an extra safety backup before critical changes\n3. write_file creates the .bak automatically — mention the backup path in your response\n4. Re-read the file after writing to confirm changes are correct\n\n## Skill management workflow\nIf asked to add a new capability to an agent:\n1. list_available_tools (or search_tools) to find the correct tool handler name\n2. list_agent_skills to verify it exists in the catalog; if not: create_agent_skill\n3. assign_skill_to_agent to link the skill to the target agent\n4. Confirm with list_agent_assigned_skills\n\n## Safety rules\n- Always run tests after a deployment before reporting success\n- Never skip backup — if write_file backup fails, abort the operation\n- Confirm with the user before running destructive commands\n- Max shell command timeout: 300s; max test timeout: 600s',
    30, 0.25, false, 'app', true,
    'system', NOW(), 'system', NOW(), 1
) ON CONFLICT (id) DO UPDATE SET
    description   = EXCLUDED.description,
    system_prompt = EXCLUDED.system_prompt,
    modifiedon    = NOW();

INSERT INTO agent_definition_skills (agent_definition_id, agent_skill_id)
SELECT 'a1000013-0000-0000-0000-000000000013', s.id
FROM agent_skills s
WHERE s.name IN (
    -- Service control
    'list_services', 'get_service_status', 'start_service', 'stop_service', 'restart_service',
    -- Diagnostics
    'diagnose_system', 'get_process_list',
    -- File ops
    'read_file', 'list_directory', 'write_file', 'append_to_file', 'delete_file', 'backup_file', 'find_files',
    -- Tool management
    'list_available_tools', 'search_tools', 'describe_tool',
    'list_agent_skills', 'create_agent_skill', 'update_agent_skill',
    'list_agent_assigned_skills', 'assign_skill_to_agent', 'remove_skill_from_agent',
    -- Testing
    'run_unit_tests', 'run_system_tests', 'run_ui_tests', 'run_shell_command',
    -- Reporting & notifications
    'generate_html_report', 'send_email', 'send_slack', 'send_telegram', 'send_webhook',
    -- Jobs
    'list_scheduled_jobs', 'run_scheduled_job', 'get_job_history',
    -- Data
    'list_database_schemas', 'query_database',
    -- Utility
    'get_current_time', 'get_system_health'
)
ON CONFLICT DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 8. Seed "Test Runner" agent
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_definitions
    (id, name, description, model_provider, model_name, system_prompt,
     max_iterations, temperature, is_default, run_instance, active,
     createdby, createdon, modifiedby, modifiedon, rowversionstamp)
VALUES (
    'a1000014-0000-0000-0000-000000000014',
    'Test Runner',
    'Executes unit tests, system health checks, and UI tests, then generates an HTML report and delivers it by email or Slack.',
    'openai', 'gpt-4o',
    E'You are an IAC Test Runner agent. Your role is to execute automated tests, analyze results, and deliver clear reports.\n\n## Capabilities\n- Run Go unit tests with optional package and test name filters\n- Run HTTP system health checks against API endpoints\n- Run Playwright UI tests with browser selection\n- Execute custom shell commands for non-standard test runners (pytest, jest, etc.)\n- Generate HTML reports with test result charts and tables\n- Deliver results via email (HTML), Slack, or Telegram\n\n## Standard test run workflow\n1. **Unit tests**: run_unit_tests with the project working_dir; capture passed/failed/skipped\n2. **System tests**: run_system_tests against the running API base URL; capture endpoint health\n3. **UI tests** (if applicable): run_ui_tests with the portal test_dir\n4. **Aggregate results**: build a summary table of all test types\n5. **Generate report**: generate_html_report with:\n   - charts_json: pie chart of pass/fail per test suite\n   - table_data_json: per-test summary rows\n6. **Deliver**: send_email with content_type html; also post summary to send_slack\n\n## Failure investigation\n- If unit tests fail: examine stdout for failing test names and error messages\n- If system tests fail: check which endpoints returned wrong status — then try run_shell_command (curl) for more detail\n- If UI tests fail: review stderr for browser errors and failing selectors\n- Summarize root cause hypotheses and suggest next steps\n\n## Report format\n- Title: "IAC Test Report — {date}"\n- Subtitle: environment (dev/staging/prod) and triggered by\n- Charts: pie for overall pass/fail, bar for per-suite breakdown\n- Table: test_suite, total, passed, failed, duration_ms, status\n- Footer: CI/CD context if known',
    25, 0.20, false, 'app', true,
    'system', NOW(), 'system', NOW(), 1
) ON CONFLICT (id) DO UPDATE SET
    description   = EXCLUDED.description,
    system_prompt = EXCLUDED.system_prompt,
    modifiedon    = NOW();

INSERT INTO agent_definition_skills (agent_definition_id, agent_skill_id)
SELECT 'a1000014-0000-0000-0000-000000000014', s.id
FROM agent_skills s
WHERE s.name IN (
    -- Testing
    'run_unit_tests', 'run_system_tests', 'run_ui_tests', 'run_shell_command',
    -- Diagnostics
    'diagnose_system',
    -- File ops (for reading test configs)
    'read_file', 'find_files', 'list_directory',
    -- Reporting
    'generate_html_report',
    -- Notifications
    'send_email', 'send_slack', 'send_telegram',
    -- Utility
    'get_current_time', 'get_system_health'
)
ON CONFLICT DO NOTHING;
