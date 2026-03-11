-- Agent Runtime v8: Schedule & Planning Skills
-- Adds skills for:
--   • AI schedule generation and optimization (AIScheduleService)
--   • Async batch optimization for large task lists (AIScheduleBatchService)
--   • Plan scheduler profile discovery and data loading (PlanSchedulerProfileService)
--   • Missing job tools from v6b: run_scheduled_job, get_job_history
-- Provisions a "Production Scheduler" seed agent.
-- Run AFTER agent_runtime_v7_postgresql.sql.

-- ─────────────────────────────────────────────────────────────────────────────
-- 1. Missing job skills (registered in Go but not seeded in any prior migration)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    (gen_random_uuid()::TEXT, 'run_scheduled_job', 'Run Scheduled Job', 'Immediately queue an existing scheduled job for execution by its ID', 'system', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'get_job_history',   'Get Job History',   'Query recent queue job execution history with optional handler and status filters', 'system', 'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 2. Planning skills (registered in Go but not seeded in any prior migration)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    (gen_random_uuid()::TEXT, 'list_plan_profiles', 'List Plan Profiles', 'List all plan scheduler profiles to discover available profiles and their IDs', 'scheduling', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'plan_optimize',       'Plan Optimize',      'Load a plan scheduler profile full data (tasks, resources, constraints) for AI-driven optimization analysis', 'scheduling', 'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 3. Schedule tools (new in this release)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    (gen_random_uuid()::TEXT, 'generate_schedule',         'Generate Schedule',           'Generate a production schedule from a natural language prompt and available resources',          'scheduling', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'optimize_schedule',         'Optimize Schedule',           'Synchronously optimize a schedule (≤ 30 tasks). Objectives: compress, balance, cost, priority', 'scheduling', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'optimize_schedule_batch',   'Optimize Schedule (Batch)',   'Async batch optimization for large schedules (30+ tasks). Returns a job_id to poll',            'scheduling', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'get_batch_job_status',      'Get Batch Job Status',        'Poll the status and progress of an async batch schedule optimization job',                      'scheduling', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'get_batch_job_result',      'Get Batch Job Result',        'Retrieve the aggregated optimized tasks and changes from a completed batch job',                'scheduling', 'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'schedule_chat',             'Schedule Chat',               'Ask the scheduling AI a conversational question about an existing schedule',                    'scheduling', 'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- 4. Seed a "Production Scheduler" agent
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_definitions
    (id, name, description, model_provider, model_name, system_prompt, max_iterations, temperature, is_default, active, createdby, createdon, modifiedby, modifiedon, rowversionstamp)
VALUES (
    'a1000009-0000-0000-0000-000000000009',
    'Production Scheduler',
    'Generates and optimizes production schedules using AI. Handles both small inline schedules and large async batch jobs. Also queries plan scheduler profiles and answers conversational schedule questions.',
    'openai', 'gpt-4o',
    E'You are an IAC Production Scheduler agent. Your role is to generate, optimize, and explain manufacturing production schedules.\n\n## Capabilities\n- Generate schedules from natural language descriptions\n- Optimize schedules by compressing timelines, balancing resource load, minimizing cost, or reordering by priority\n- Handle large schedules (30+ tasks) using async batch optimization\n- Load and analyze plan scheduler profiles with predefined data sources and constraints\n- Answer conversational questions about schedule decisions, conflicts, and critical paths\n\n## Workflow\n\n### Generating a new schedule\n1. Clarify requirements: product, quantity, deadline, available resources, shift patterns\n2. Call generate_schedule with a detailed prompt and resources_json if known\n3. Review the result and call optimize_schedule if further refinement is needed\n4. Present the final schedule with a summary of key dates and resource assignments\n\n### Optimizing an existing schedule\n1. If task count ≤ 30: call optimize_schedule directly\n2. If task count > 30: call optimize_schedule_batch, then poll get_batch_job_status every 10–15 seconds until status = ''completed'' or ''partial'', then call get_batch_job_result\n3. Summarize the changes made and explain the improvements\n\n### Working with plan profiles\n1. Call list_plan_profiles to discover available profiles\n2. Call plan_optimize with the chosen profile_id to load tasks, resources, constraints, and settings\n3. Analyze the data for conflicts, bottlenecks, and optimization opportunities\n4. Use schedule_chat to answer specific questions or use optimize_schedule to apply changes\n\n## Guidelines\n- Always describe what each schedule change achieves and why\n- Respect task dependencies (FS = Finish-to-Start, SS = Start-to-Start, FF = Finish-to-Finish)\n- Highlight over-allocated resources and suggest realistic alternatives\n- When batch optimization is partial (some batches failed), clearly state which tasks were not optimized\n- Provide ISO date strings (YYYY-MM-DD or full ISO) for all start/end fields',
    20, 0.40, false, true, 'system', NOW(), 'system', NOW(), 1
) ON CONFLICT (id) DO UPDATE SET
    description   = EXCLUDED.description,
    system_prompt = EXCLUDED.system_prompt,
    max_iterations = EXCLUDED.max_iterations,
    modifiedon    = NOW();

-- ─────────────────────────────────────────────────────────────────────────────
-- 5. Assign scheduling + planning + data skills to Production Scheduler
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO agent_definition_skills (agent_definition_id, agent_skill_id)
SELECT 'a1000009-0000-0000-0000-000000000009', s.id
FROM agent_skills s
WHERE s.name IN (
    -- Schedule generation & optimization
    'generate_schedule',
    'optimize_schedule',
    'optimize_schedule_batch',
    'get_batch_job_status',
    'get_batch_job_result',
    'schedule_chat',
    -- Plan profiles
    'list_plan_profiles',
    'plan_optimize',
    -- Job management (run and review)
    'run_scheduled_job',
    'get_job_history',
    'list_scheduled_jobs',
    'list_queue_jobs',
    -- Data utilities (for querying schedule-related data)
    'list_database_schemas',
    'query_database',
    'query_database_nl',
    -- Utility
    'get_current_time',
    'web_search'
)
ON CONFLICT DO NOTHING;
