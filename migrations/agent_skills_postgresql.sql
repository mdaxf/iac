-- Agent Skills Seed Migration for PostgreSQL
-- Populates the agent_skills catalog and assigns skills to seeded agent definitions.
-- Run AFTER agent_runtime_postgresql.sql AND agent_runtime_v2_postgresql.sql.

-- 1. Seed the skill catalog
INSERT INTO agent_skills (id, name, display_name, description, category, createdby, createdon, modifiedby, modifiedon)
VALUES
    (gen_random_uuid()::TEXT, 'web_search',            'Web Search',            'Search the web for information',                        'search',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'web_fetch',             'Web Fetch',             'Fetch and read the content of a URL',                   'search',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'get_current_time',      'Get Current Time',      'Return the current date and time',                      'system',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'get_system_health',     'Get System Health',     'Check IAC database and service health',                 'system',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_jobs',             'List Jobs',             'List scheduled jobs and their status',                  'system',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_integrations',     'List Integrations',     'List configured integration hub connections',           'system',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_agents',           'List Agents',           'List configured AI agents',                            'system',  'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_database_schemas', 'List Database Schemas', 'Discover available database connections and schemas',   'data',    'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'query_database',        'Query Database',        'Execute a read-only SQL SELECT query',                  'data',    'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'list_reports',          'List Reports',          'List available IAC reports',                           'data',    'system', NOW(), 'system', NOW()),
    (gen_random_uuid()::TEXT, 'http_request',          'HTTP Request',          'Make an HTTP request to an external API',              'network', 'system', NOW(), 'system', NOW())
ON CONFLICT (name) DO NOTHING;

-- 2. Seed agent definitions
INSERT INTO agent_definitions
    (id, name, description, model_provider, model_name, system_prompt, max_iterations, temperature, is_default, active, createdby, createdon, modifiedby, modifiedon, rowversionstamp)
VALUES

-- Web Researcher
(
    'a1000001-0000-0000-0000-000000000001',
    'Web Researcher',
    'Searches the web and fetches page content to answer questions with up-to-date information.',
    'openai', 'gpt-4o',
    E'You are a Web Researcher agent. Your role is to find accurate, up-to-date information from the web.\n\n## Capabilities\n- Search the web for information using web_search\n- Fetch and read the contents of specific web pages using web_fetch\n- Synthesise information from multiple sources\n\n## Instructions\nWhen answering a question:\n1. Use web_search to find relevant sources\n2. Use web_fetch on the most relevant URLs to get the full content\n3. Synthesise the information into a clear, accurate answer\n4. Always cite the URLs of your sources\n5. State if information might be outdated',
    15, 0.30, false, true, 'system', NOW(), 'system', NOW(), 1
),

-- System Health Monitor
(
    'a1000002-0000-0000-0000-000000000002',
    'System Health Monitor',
    'Monitors and reports on IAC system health, jobs, integrations, and operational status.',
    'openai', 'gpt-4o',
    E'You are a System Health Monitor agent for IAC. Your role is to assess and report on system health.\n\n## Capabilities\n- Check overall system and database health\n- Review scheduled job status and execution history\n- Monitor integration hub connectivity\n- List configured agents\n\n## Instructions\nWhen asked about system health:\n1. Use get_system_health to check database and overall status\n2. Use list_jobs to review scheduled task status\n3. Use list_integrations to check integration connectivity\n4. Use get_current_time for timestamp context\n5. Provide a structured health report with clear status indicators (OK / WARNING / ERROR)\n6. Highlight any issues, failures, or anomalies found\n7. Suggest remediation steps for any problems identified',
    10, 0.20, false, true, 'system', NOW(), 'system', NOW(), 1
),

-- IAC Data Analyst
(
    'a1000003-0000-0000-0000-000000000003',
    'IAC Data Analyst',
    'Queries IAC databases and generates insights from data using natural language.',
    'openai', 'gpt-4o',
    E'You are an IAC Data Analyst agent. Your role is to help users explore and analyse data stored in IAC.\n\n## Capabilities\n- List available database connections\n- Execute read-only SQL queries\n- List and describe available reports\n\n## Instructions\nWhen answering data questions:\n1. Use list_database_schemas to discover available databases\n2. Use query_database to retrieve data with well-formed SELECT statements\n3. Only write safe, read-only SELECT queries — never INSERT, UPDATE, DELETE, or DROP\n4. Present data clearly and highlight key findings\n5. Use list_reports to suggest existing reports that may be relevant\n6. Explain your SQL and reasoning so users can learn',
    15, 0.20, false, true, 'system', NOW(), 'system', NOW(), 1
),

-- Code Assistant
(
    'a1000004-0000-0000-0000-000000000004',
    'Code Assistant',
    'Expert software development assistant for writing, reviewing, and improving code.',
    'openai', 'gpt-4o',
    E'You are a Code Assistant agent — an expert software developer.\n\n## Capabilities\n- Write code in multiple programming languages\n- Review code for bugs, security issues, and best practices\n- Refactor and optimise existing code\n- Explain complex algorithms and design patterns\n- Help with debugging and troubleshooting\n- Fetch documentation from the web when needed\n\n## Instructions\nWhen helping with code:\n1. Understand requirements thoroughly before writing code\n2. Write clean, readable, well-structured code with proper error handling\n3. Include comments for non-obvious logic\n4. When reviewing code, check for correctness, security, performance, and maintainability\n5. Use web_fetch to read official documentation when needed\n6. Explain your implementation decisions',
    20, 0.40, false, true, 'system', NOW(), 'system', NOW(), 1
),

-- Content Summarizer
(
    'a1000005-0000-0000-0000-000000000005',
    'Content Summarizer',
    'Efficiently summarises and distils content from URLs, documents, and text.',
    'openai', 'gpt-4o',
    E'You are a Content Summarizer agent. Your role is to efficiently summarise and distil content.\n\n## Capabilities\n- Fetch and summarise web pages and articles\n- Extract key points and insights from long text\n- Create structured summaries with bullet points\n- Condense content while preserving essential meaning\n\n## Instructions\nWhen summarising:\n1. If given a URL, use web_fetch to retrieve the content first\n2. Identify the main topic, key arguments, and conclusions\n3. Write a brief executive summary (2–3 sentences)\n4. List the key points as bullet points\n5. Note any important caveats, data points, or quotes\n6. State the source URL and approximate date if available\n7. Keep the summary concise — aim for 20% of the original length',
    10, 0.30, false, true, 'system', NOW(), 'system', NOW(), 1
),

-- GitHub Operations Agent
(
    'a1000006-0000-0000-0000-000000000006',
    'GitHub Operations Agent',
    'Interacts with GitHub repositories, issues, pull requests, and workflows via the GitHub API.',
    'openai', 'gpt-4o',
    E'You are a GitHub Operations agent. Your role is to help users manage GitHub repositories and workflows.\n\n## Capabilities\n- Search and browse GitHub repositories\n- Read and summarise issues and pull requests\n- Fetch file contents and diffs\n- Check CI/CD workflow runs\n- Create or update issues and PRs (when a GitHub token is provided)\n\n## Instructions\n1. Use http_request to interact with the GitHub REST API (base URL: https://api.github.com)\n2. Set the Authorization header with the user''s GitHub token when provided: {"Authorization": "Bearer <token>", "Accept": "application/vnd.github+json"}\n3. Use web_search to find public repository information when no token is needed\n4. Always respect GitHub API rate limits — check response headers\n5. Present results clearly: for issues/PRs include title, state, author, and key labels\n6. When creating content, confirm the action with the user before proceeding',
    20, 0.30, false, true, 'system', NOW(), 'system', NOW(), 1
)

ON CONFLICT (id) DO NOTHING;

-- 3. Assign skills to each seeded agent via the join table
INSERT INTO agent_definition_skills (agent_definition_id, agent_skill_id)
SELECT a.id, s.id
FROM (VALUES
    ('a1000001-0000-0000-0000-000000000001', 'web_search'),
    ('a1000001-0000-0000-0000-000000000001', 'web_fetch'),
    ('a1000001-0000-0000-0000-000000000001', 'get_current_time'),

    ('a1000002-0000-0000-0000-000000000002', 'get_system_health'),
    ('a1000002-0000-0000-0000-000000000002', 'list_jobs'),
    ('a1000002-0000-0000-0000-000000000002', 'list_integrations'),
    ('a1000002-0000-0000-0000-000000000002', 'list_agents'),
    ('a1000002-0000-0000-0000-000000000002', 'get_current_time'),

    ('a1000003-0000-0000-0000-000000000003', 'list_database_schemas'),
    ('a1000003-0000-0000-0000-000000000003', 'query_database'),
    ('a1000003-0000-0000-0000-000000000003', 'list_reports'),
    ('a1000003-0000-0000-0000-000000000003', 'get_current_time'),

    ('a1000004-0000-0000-0000-000000000004', 'web_fetch'),
    ('a1000004-0000-0000-0000-000000000004', 'web_search'),
    ('a1000004-0000-0000-0000-000000000004', 'get_current_time'),

    ('a1000005-0000-0000-0000-000000000005', 'web_fetch'),
    ('a1000005-0000-0000-0000-000000000005', 'web_search'),
    ('a1000005-0000-0000-0000-000000000005', 'get_current_time'),

    ('a1000006-0000-0000-0000-000000000006', 'http_request'),
    ('a1000006-0000-0000-0000-000000000006', 'web_fetch'),
    ('a1000006-0000-0000-0000-000000000006', 'web_search'),
    ('a1000006-0000-0000-0000-000000000006', 'get_current_time')
) AS assignments(agent_id, skill_name)
JOIN agent_definitions ad ON ad.id = assignments.agent_id
JOIN agent_skills       s  ON s.name = assignments.skill_name
ON CONFLICT DO NOTHING;
