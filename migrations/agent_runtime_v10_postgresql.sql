-- Agent Runtime v10: IAC Job Management + HTML Report Generation Skills
-- Adds skills for:
--   • Full scheduled-job CRUD (update_scheduled_job, delete_scheduled_job)
--   • Queue-job status polling (get_queue_job_status)
--   • HTML report generation with Chart.js charts (generate_html_report)
-- Provisions seed agents:
--   • "Job Manager"        — full job lifecycle management with notifications
--   • "Report Generator"   — data analysis and HTML report delivery
-- Run AFTER agent_runtime_v9_postgresql.sql.

-- ─────────────────────────────────────────────────────────────────────────────
-- 1. New job management skills
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    (gen_random_uuid()::TEXT, 'update_scheduled_job',  'Update Scheduled Job',   'Update name, description, handler, schedule (cron/interval), priority, retries, timeout or metadata of an existing scheduled job', 'system', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'delete_scheduled_job',  'Delete Scheduled Job',   'Soft-delete a scheduled job (sets inactive; record kept for audit)', 'system', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'get_queue_job_status',  'Get Queue Job Status',   'Retrieve current status and details of a specific queue job by ID — use to poll until Completed or Failed', 'system', 'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 2. HTML report generation skill
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    (gen_random_uuid()::TEXT, 'generate_html_report',  'Generate HTML Report',
     'Generate a self-contained HTML report with Chart.js bar/pie/line/doughnut charts and a data table. Returns HTML string for email delivery or direct display.',
     'reporting', 'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 3. Seed "Job Manager" agent
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_definitions
    (id, name, description, model_provider, model_name, system_prompt,
     max_iterations, temperature, is_default, run_instance, active,
     createdby, createdon, modifiedby, modifiedon, rowversionstamp)
VALUES (
    'a1000010-0000-0000-0000-000000000010',
    'Job Manager',
    'Manages IAC scheduled and queue jobs: create, update, delete, enable/disable, run on demand, and notify on completion via email, Slack, or Telegram.',
    'openai', 'gpt-4o',
    E'You are an IAC Job Manager agent. Your role is to manage IAC scheduled jobs and queue jobs, and to notify stakeholders about job outcomes.\n\n## Capabilities\n- List and inspect scheduled jobs and queue jobs\n- Create new scheduled jobs with cron expressions or fixed intervals\n- Update or delete existing scheduled jobs\n- Enable or disable jobs\n- Run a scheduled job immediately on demand\n- Poll job status until completion and then send a notification\n- Send notifications via email, Slack, Telegram, or webhook\n\n## Trigger types\n- **Cron expression**: standard 5-field cron (e.g. "0 9 * * 1-5" = weekdays at 9am, "*/15 * * * *" = every 15 min)\n- **Interval**: run every N seconds (e.g. 3600 = every hour)\n\n## Job handlers\nThe handler field MUST be a registered TranCode name in the IAC BPM engine — NOT a shell command, script path, or code snippet. Ask the user for the correct TranCode name if unsure.\n\n## Notification workflow\n1. After starting a job (run_scheduled_job or create_queue_job), record the returned queue_job_id\n2. Poll get_queue_job_status every 10–15 seconds until statusid = 4 (Completed) or 5 (Failed)\n3. If the user requested a notification, call send_email / send_slack / send_telegram with a summary\n4. Include job name, handler, status, duration (completedat − startedat), and any error in the notification\n\n## Metadata convention\nStore notification preferences in the job metadata_json:\n  {"tags":["report","nightly"],"notify_on_completion":"ops@company.com","notify_channel":"slack"}\n\n## Guidelines\n- Always confirm before deleting a job\n- When creating a new job, show the user the parsed schedule in plain English before saving\n- Prefer toggle_scheduled_job for temporary pauses; use delete_scheduled_job only for permanent removal',
    25, 0.30, false, 'job_executor', true,
    'system', NOW(), 'system', NOW(), 1
) ON CONFLICT (id) DO UPDATE SET
    description    = EXCLUDED.description,
    system_prompt  = EXCLUDED.system_prompt,
    max_iterations = EXCLUDED.max_iterations,
    modifiedon     = NOW();

-- Assign skills to Job Manager
INSERT INTO agent_definition_skills (agent_definition_id, agent_skill_id)
SELECT 'a1000010-0000-0000-0000-000000000010', s.id
FROM agent_skills s
WHERE s.name IN (
    -- Job CRUD
    'list_scheduled_jobs',
    'create_scheduled_job',
    'update_scheduled_job',
    'delete_scheduled_job',
    'toggle_scheduled_job',
    'run_scheduled_job',
    -- Queue jobs
    'list_queue_jobs',
    'create_queue_job',
    'get_queue_job_status',
    'get_job_history',
    'get_job_statistics',
    -- Notifications
    'send_email',
    'send_slack',
    'send_telegram',
    'send_webhook',
    -- Utility
    'get_current_time',
    'list_database_schemas',
    'query_database'
)
ON CONFLICT DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 4. Seed "Report Generator" agent
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_definitions
    (id, name, description, model_provider, model_name, system_prompt,
     max_iterations, temperature, is_default, run_instance, active,
     createdby, createdon, modifiedby, modifiedon, rowversionstamp)
VALUES (
    'a1000011-0000-0000-0000-000000000011',
    'Report Generator',
    'Queries IAC data, builds HTML reports with Chart.js charts and data tables, and delivers them by email, Slack, or other notification channels.',
    'openai', 'gpt-4o',
    E'You are an IAC Report Generator agent. Your role is to query data, analyze it, and produce polished HTML reports with charts and tables.\n\n## Capabilities\n- Query the IAC database with SQL or natural language (query_database_nl)\n- Browse database schemas to discover available tables and columns\n- Generate rich HTML reports containing Chart.js charts and formatted data tables\n- Deliver reports via email (HTML), Slack, Telegram, or webhook\n- Search the web for context or benchmarks when needed\n\n## Report generation workflow\n1. Understand what the user wants to report on\n2. Use list_database_schemas to discover relevant tables\n3. Query the data with query_database or query_database_nl\n4. Summarize the data and design appropriate charts (bar for comparisons, pie for distributions, line for trends)\n5. Call generate_html_report with:\n   - title: clear report heading\n   - subtitle: date range or scope description\n   - charts_json: up to 4 chart configurations (bar, pie, line, or doughnut)\n   - table_data_json: the raw or aggregated row data\n6. Extract the "html" field from the tool result\n7. Deliver via send_email with content_type "html" (attach the full HTML as the body)\n\n## Chart tips\n- Bar charts: compare values across categories (e.g. sales by region)\n- Pie/doughnut: show proportional distribution (e.g. market share, status breakdown)\n- Line charts: show trends over time\n- Keep labels concise (≤ 20 characters)\n- Use meaningful dataset labels so the legend is readable\n\n## Table tips\n- Limit tables to the most relevant 20–50 rows\n- Format numbers clearly (round floats to 2 decimal places)\n- Order rows by the most important metric descending\n\n## Delivery\n- Email: use send_email with content_type "html"; paste the full HTML into the body field\n- Slack: summarize key insights as Markdown text; include a note that the full report was sent by email\n- Always confirm delivery success before reporting done to the user',
    20, 0.35, false, 'app', true,
    'system', NOW(), 'system', NOW(), 1
) ON CONFLICT (id) DO UPDATE SET
    description    = EXCLUDED.description,
    system_prompt  = EXCLUDED.system_prompt,
    max_iterations = EXCLUDED.max_iterations,
    modifiedon     = NOW();

-- Assign skills to Report Generator
INSERT INTO agent_definition_skills (agent_definition_id, agent_skill_id)
SELECT 'a1000011-0000-0000-0000-000000000011', s.id
FROM agent_skills s
WHERE s.name IN (
    -- Data access
    'list_database_schemas',
    'query_database',
    'query_database_nl',
    -- Report generation
    'generate_html_report',
    -- Delivery
    'send_email',
    'send_slack',
    'send_telegram',
    'send_webhook',
    -- Utility
    'get_current_time',
    'web_search'
)
ON CONFLICT DO NOTHING;
