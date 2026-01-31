-- =====================================================
-- Demo Plan Scheduler Profile with Sample Data
-- Generated from shortmockData.ts
-- Start Date: 2026-01-01
-- All tasks: todo/new status
-- =====================================================

-- 1. Create Demo Profile
INSERT INTO plan_scheduler_profiles (
    id, name, description, is_default,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES (
    'demo-profile-001',
    'Manufacturing Demo - 2026 Q1',
    'Sample manufacturing schedule with 5 work orders, 25 tasks, and 25 resources (5 work centers, 10 machines, 10 operators). Starting January 1, 2026.',
    FALSE,
    TRUE,
    'DEMO-2026-Q1',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- 2. Data Source: Tasks (JSON Constant)
INSERT INTO plan_scheduler_profile_datasources (
    id, profile_id, data_type, name, description,
    source_type, source_json, display_order,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES (
    'ds-demo-tasks-001',
    'demo-profile-001',
    'tasks',
    'Demo Manufacturing Tasks',
    '25 tasks across 5 work orders (WO-101 to WO-105), each with 5 operations',
    'json',
    '[
  {"id":"P-1","jobId":"WO-101","name":"WO-101 - Production Run","resourceIds":[],"start":"2026-01-01T08:00:00Z","end":"2026-01-02T00:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Main production run for order WO-101"},
  {"id":"t-1-1","parentId":"P-1","jobId":"WO-101","name":"Op 01: Processing","resourceIds":["wc-1"],"start":"2026-01-01T08:00:00Z","end":"2026-01-01T15:00:00Z","progress":0,"status":"todo","allocation":100,"setupMinutes":30,"description":"Step 1 assigned to Work Center 1"},
  {"id":"t-1-2","parentId":"P-1","jobId":"WO-101","name":"Op 02: Fabrication","resourceIds":["m-3"],"start":"2026-01-01T15:00:00Z","end":"2026-01-01T21:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 2 assigned to Machine 3","dependencies":[{"predecessorId":"t-1-1","type":"FS","lag":0}]},
  {"id":"t-1-3","parentId":"P-1","jobId":"WO-101","name":"Op 03: Manual Assembly","resourceIds":["h-2"],"start":"2026-01-01T21:00:00Z","end":"2026-01-02T02:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 3 assigned to Person 2","dependencies":[{"predecessorId":"t-1-2","type":"FS","lag":0}]},
  {"id":"t-1-4","parentId":"P-1","jobId":"WO-101","name":"Op 04: Fabrication","resourceIds":["m-7"],"start":"2026-01-02T02:00:00Z","end":"2026-01-02T09:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 4 assigned to Machine 7","dependencies":[{"predecessorId":"t-1-3","type":"FS","lag":0}]},
  {"id":"t-1-5","parentId":"P-1","jobId":"WO-101","name":"Op 05: Manual Assembly","resourceIds":["h-5"],"start":"2026-01-02T09:00:00Z","end":"2026-01-02T16:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 5 assigned to Person 5","dependencies":[{"predecessorId":"t-1-4","type":"FS","lag":0}]},
  {"id":"P-2","jobId":"WO-102","name":"WO-102 - Production Run","resourceIds":[],"start":"2026-01-02T08:00:00Z","end":"2026-01-03T02:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Main production run for order WO-102"},
  {"id":"t-2-1","parentId":"P-2","jobId":"WO-102","name":"Op 01: Processing","resourceIds":["wc-2"],"start":"2026-01-02T08:00:00Z","end":"2026-01-02T14:00:00Z","progress":0,"status":"todo","allocation":100,"setupMinutes":30,"description":"Step 1 assigned to Work Center 2"},
  {"id":"t-2-2","parentId":"P-2","jobId":"WO-102","name":"Op 02: Fabrication","resourceIds":["m-1"],"start":"2026-01-02T14:00:00Z","end":"2026-01-02T21:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 2 assigned to Machine 1","dependencies":[{"predecessorId":"t-2-1","type":"FS","lag":0}]},
  {"id":"t-2-3","parentId":"P-2","jobId":"WO-102","name":"Op 03: Manual Assembly","resourceIds":["h-8"],"start":"2026-01-02T21:00:00Z","end":"2026-01-03T04:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 3 assigned to Person 8","dependencies":[{"predecessorId":"t-2-2","type":"FS","lag":0}]},
  {"id":"t-2-4","parentId":"P-2","jobId":"WO-102","name":"Op 04: Fabrication","resourceIds":["m-4"],"start":"2026-01-03T04:00:00Z","end":"2026-01-03T10:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 4 assigned to Machine 4","dependencies":[{"predecessorId":"t-2-3","type":"FS","lag":0}]},
  {"id":"t-2-5","parentId":"P-2","jobId":"WO-102","name":"Op 05: Manual Assembly","resourceIds":["h-1"],"start":"2026-01-03T10:00:00Z","end":"2026-01-03T15:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 5 assigned to Person 1","dependencies":[{"predecessorId":"t-2-4","type":"FS","lag":0}]},
  {"id":"P-3","jobId":"WO-103","name":"WO-103 - Production Run","resourceIds":[],"start":"2026-01-03T08:00:00Z","end":"2026-01-04T07:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Main production run for order WO-103"},
  {"id":"t-3-1","parentId":"P-3","jobId":"WO-103","name":"Op 01: Processing","resourceIds":["wc-4"],"start":"2026-01-03T08:00:00Z","end":"2026-01-03T13:00:00Z","progress":0,"status":"todo","allocation":100,"setupMinutes":30,"description":"Step 1 assigned to Work Center 4"},
  {"id":"t-3-2","parentId":"P-3","jobId":"WO-103","name":"Op 02: Fabrication","resourceIds":["m-9"],"start":"2026-01-03T13:00:00Z","end":"2026-01-03T20:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 2 assigned to Machine 9","dependencies":[{"predecessorId":"t-3-1","type":"FS","lag":0}]},
  {"id":"t-3-3","parentId":"P-3","jobId":"WO-103","name":"Op 03: Manual Assembly","resourceIds":["h-6"],"start":"2026-01-03T20:00:00Z","end":"2026-01-04T02:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 3 assigned to Person 6","dependencies":[{"predecessorId":"t-3-2","type":"FS","lag":0}]},
  {"id":"t-3-4","parentId":"P-3","jobId":"WO-103","name":"Op 04: Fabrication","resourceIds":["m-2"],"start":"2026-01-04T02:00:00Z","end":"2026-01-04T08:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 4 assigned to Machine 2","dependencies":[{"predecessorId":"t-3-3","type":"FS","lag":0}]},
  {"id":"t-3-5","parentId":"P-3","jobId":"WO-103","name":"Op 05: Manual Assembly","resourceIds":["h-10"],"start":"2026-01-04T08:00:00Z","end":"2026-01-04T15:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 5 assigned to Person 10","dependencies":[{"predecessorId":"t-3-4","type":"FS","lag":0}]},
  {"id":"P-4","jobId":"WO-104","name":"WO-104 - Production Run","resourceIds":[],"start":"2026-01-06T08:00:00Z","end":"2026-01-07T05:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Main production run for order WO-104"},
  {"id":"t-4-1","parentId":"P-4","jobId":"WO-104","name":"Op 01: Processing","resourceIds":["wc-3"],"start":"2026-01-06T08:00:00Z","end":"2026-01-06T15:00:00Z","progress":0,"status":"todo","allocation":100,"setupMinutes":30,"description":"Step 1 assigned to Work Center 3"},
  {"id":"t-4-2","parentId":"P-4","jobId":"WO-104","name":"Op 02: Fabrication","resourceIds":["m-5"],"start":"2026-01-06T15:00:00Z","end":"2026-01-06T19:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 2 assigned to Machine 5","dependencies":[{"predecessorId":"t-4-1","type":"FS","lag":0}]},
  {"id":"t-4-3","parentId":"P-4","jobId":"WO-104","name":"Op 03: Manual Assembly","resourceIds":["h-3"],"start":"2026-01-06T19:00:00Z","end":"2026-01-07T00:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 3 assigned to Person 3","dependencies":[{"predecessorId":"t-4-2","type":"FS","lag":0}]},
  {"id":"t-4-4","parentId":"P-4","jobId":"WO-104","name":"Op 04: Fabrication","resourceIds":["m-10"],"start":"2026-01-07T00:00:00Z","end":"2026-01-07T06:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 4 assigned to Machine 10","dependencies":[{"predecessorId":"t-4-3","type":"FS","lag":0}]},
  {"id":"t-4-5","parentId":"P-4","jobId":"WO-104","name":"Op 05: Manual Assembly","resourceIds":["h-7"],"start":"2026-01-07T06:00:00Z","end":"2026-01-07T13:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 5 assigned to Person 7","dependencies":[{"predecessorId":"t-4-4","type":"FS","lag":0}]},
  {"id":"P-5","jobId":"WO-105","name":"WO-105 - Production Run","resourceIds":[],"start":"2026-01-07T08:00:00Z","end":"2026-01-08T01:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Main production run for order WO-105"},
  {"id":"t-5-1","parentId":"P-5","jobId":"WO-105","name":"Op 01: Processing","resourceIds":["wc-5"],"start":"2026-01-07T08:00:00Z","end":"2026-01-07T12:00:00Z","progress":0,"status":"todo","allocation":100,"setupMinutes":30,"description":"Step 1 assigned to Work Center 5"},
  {"id":"t-5-2","parentId":"P-5","jobId":"WO-105","name":"Op 02: Fabrication","resourceIds":["m-6"],"start":"2026-01-07T12:00:00Z","end":"2026-01-07T19:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 2 assigned to Machine 6","dependencies":[{"predecessorId":"t-5-1","type":"FS","lag":0}]},
  {"id":"t-5-3","parentId":"P-5","jobId":"WO-105","name":"Op 03: Manual Assembly","resourceIds":["h-9"],"start":"2026-01-07T19:00:00Z","end":"2026-01-08T01:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 3 assigned to Person 9","dependencies":[{"predecessorId":"t-5-2","type":"FS","lag":0}]},
  {"id":"t-5-4","parentId":"P-5","jobId":"WO-105","name":"Op 04: Fabrication","resourceIds":["m-8"],"start":"2026-01-08T01:00:00Z","end":"2026-01-08T06:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 4 assigned to Machine 8","dependencies":[{"predecessorId":"t-5-3","type":"FS","lag":0}]},
  {"id":"t-5-5","parentId":"P-5","jobId":"WO-105","name":"Op 05: Manual Assembly","resourceIds":["h-4"],"start":"2026-01-08T06:00:00Z","end":"2026-01-08T13:00:00Z","progress":0,"status":"todo","allocation":100,"description":"Step 5 assigned to Person 4","dependencies":[{"predecessorId":"t-5-4","type":"FS","lag":0}]}
]'::jsonb,
    1,
    TRUE,
    'DEMO-TASKS',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- 3. Data Source: Resources (JSON Constant)
INSERT INTO plan_scheduler_profile_datasources (
    id, profile_id, data_type, name, description,
    source_type, source_json, display_order,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES (
    'ds-demo-resources-001',
    'demo-profile-001',
    'resources',
    'Demo Manufacturing Resources',
    '25 resources: 5 work centers, 10 machines (24/7 availability), 10 operators (standard hours)',
    'json',
    '[
  {"id":"wc-1","name":"Work Center 1","role":"Assembly","type":"work_center","color":"#2563eb","availability":"standard"},
  {"id":"wc-2","name":"Work Center 2","role":"Assembly","type":"work_center","color":"#3b82f6","availability":"standard"},
  {"id":"wc-3","name":"Work Center 3","role":"Assembly","type":"work_center","color":"#60a5fa","availability":"standard"},
  {"id":"wc-4","name":"Work Center 4","role":"Assembly","type":"work_center","color":"#93c5fd","availability":"standard"},
  {"id":"wc-5","name":"Work Center 5","role":"Assembly","type":"work_center","color":"#bfdbfe","availability":"standard"},
  {"id":"m-1","name":"Machine 1","role":"Fabrication","type":"machine","color":"#b45309","availability":"always"},
  {"id":"m-2","name":"Machine 2","role":"Fabrication","type":"machine","color":"#d97706","availability":"always"},
  {"id":"m-3","name":"Machine 3","role":"Fabrication","type":"machine","color":"#f59e0b","availability":"always"},
  {"id":"m-4","name":"Machine 4","role":"Fabrication","type":"machine","color":"#fbbf24","availability":"always"},
  {"id":"m-5","name":"Machine 5","role":"Fabrication","type":"machine","color":"#fcd34d","availability":"always"},
  {"id":"m-6","name":"Machine 6","role":"Fabrication","type":"machine","color":"#7f1d1d","availability":"always"},
  {"id":"m-7","name":"Machine 7","role":"Fabrication","type":"machine","color":"#991b1b","availability":"always"},
  {"id":"m-8","name":"Machine 8","role":"Fabrication","type":"machine","color":"#b91c1c","availability":"always"},
  {"id":"m-9","name":"Machine 9","role":"Fabrication","type":"machine","color":"#dc2626","availability":"always"},
  {"id":"m-10","name":"Machine 10","role":"Fabrication","type":"machine","color":"#ef4444","availability":"always"},
  {"id":"h-1","name":"Person 1","role":"Operator","type":"human","color":"#064e3b","availability":"standard"},
  {"id":"h-2","name":"Person 2","role":"Operator","type":"human","color":"#065f46","availability":"standard"},
  {"id":"h-3","name":"Person 3","role":"Operator","type":"human","color":"#047857","availability":"standard"},
  {"id":"h-4","name":"Person 4","role":"Operator","type":"human","color":"#059669","availability":"standard"},
  {"id":"h-5","name":"Person 5","role":"Operator","type":"human","color":"#10b981","availability":"standard"},
  {"id":"h-6","name":"Person 6","role":"Operator","type":"human","color":"#34d399","availability":"standard"},
  {"id":"h-7","name":"Person 7","role":"Operator","type":"human","color":"#6ee7b7","availability":"standard"},
  {"id":"h-8","name":"Person 8","role":"Operator","type":"human","color":"#4c1d95","availability":"standard"},
  {"id":"h-9","name":"Person 9","role":"Operator","type":"human","color":"#6d28d9","availability":"standard"},
  {"id":"h-10","name":"Person 10","role":"Operator","type":"human","color":"#8b5cf6","availability":"standard"}
]'::jsonb,
    2,
    TRUE,
    'DEMO-RESOURCES',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- 4. Constraints: Time Constraint (Standard Working Hours)
INSERT INTO plan_scheduler_profile_constraints (
    id, profile_id, constraint_type, name, description,
    source_type, source_json, enforcement, display_order,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES (
    'const-demo-time-001',
    'demo-profile-001',
    'time',
    'Standard Working Hours',
    'Mon-Fri 9:00-17:00 with 12:00-13:00 lunch break',
    'json',
    '{"workStartHour":9,"workEndHour":17,"breakStartHour":12,"breakEndHour":13,"workDays":[1,2,3,4,5],"nonWorkColor":"rgba(241,245,249,0.6)"}'::jsonb,
    'hard',
    1,
    TRUE,
    'TIME-STD',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- 5. Constraints: Resource Constraint
INSERT INTO plan_scheduler_profile_constraints (
    id, profile_id, constraint_type, name, description,
    source_type, source_json, enforcement, display_order,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES (
    'const-demo-resource-001',
    'demo-profile-001',
    'resource',
    'Resource Capacity Limits',
    'Each resource can handle max 3 concurrent tasks',
    'json',
    '{"maxConcurrentTasks":3,"allowOverallocation":false}'::jsonb,
    'soft',
    2,
    TRUE,
    'RES-CAP',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- 6. Constraints: Dependency Constraint
INSERT INTO plan_scheduler_profile_constraints (
    id, profile_id, constraint_type, name, description,
    source_type, source_json, enforcement, display_order,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES (
    'const-demo-dep-001',
    'demo-profile-001',
    'dependency',
    'Task Dependencies',
    'Enforce finish-to-start dependencies between operations',
    'json',
    '{"strictDependencies":true,"allowNegativeLag":false}'::jsonb,
    'hard',
    3,
    TRUE,
    'DEP-FS',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- 7. Settings: AI Optimization Parameters
INSERT INTO plan_scheduler_profile_settings (
    id, profile_id, setting_key, setting_value, description,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES
(
    'set-demo-opt-001',
    'demo-profile-001',
    'optimization_level',
    '"balanced"'::jsonb,
    'AI optimization level: fast, balanced, or thorough',
    TRUE,
    'OPT-LEVEL',
    'system',
    NOW(),
    'system',
    NOW(),
    1
),
(
    'set-demo-opt-002',
    'demo-profile-001',
    'max_iterations',
    '1000'::jsonb,
    'Maximum AI optimization iterations',
    TRUE,
    'OPT-MAXITER',
    'system',
    NOW(),
    'system',
    NOW(),
    1
),
(
    'set-demo-opt-003',
    'demo-profile-001',
    'minimize_makespan',
    'true'::jsonb,
    'Minimize total schedule duration',
    TRUE,
    'OPT-MAKESPAN',
    'system',
    NOW(),
    'system',
    NOW(),
    1
),
(
    'set-demo-opt-004',
    'demo-profile-001',
    'resource_leveling',
    'true'::jsonb,
    'Balance resource utilization across timeline',
    TRUE,
    'OPT-LEVELING',
    'system',
    NOW(),
    'system',
    NOW(),
    1
),
(
    'set-demo-opt-005',
    'demo-profile-001',
    'time_buffer_percent',
    '10'::jsonb,
    'Safety buffer as percentage of task duration',
    TRUE,
    'OPT-BUFFER',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- =====================================================
-- VERIFICATION QUERIES
-- =====================================================

-- Count profile data
SELECT
    'Profile' as entity,
    COUNT(*) as count
FROM plan_scheduler_profiles
WHERE id = 'demo-profile-001'

UNION ALL

SELECT
    'Data Sources' as entity,
    COUNT(*) as count
FROM plan_scheduler_profile_datasources
WHERE profile_id = 'demo-profile-001'

UNION ALL

SELECT
    'Constraints' as entity,
    COUNT(*) as count
FROM plan_scheduler_profile_constraints
WHERE profile_id = 'demo-profile-001'

UNION ALL

SELECT
    'Settings' as entity,
    COUNT(*) as count
FROM plan_scheduler_profile_settings
WHERE profile_id = 'demo-profile-001';

-- Show profile summary
SELECT
    p.name,
    p.description,
    (SELECT COUNT(*) FROM plan_scheduler_profile_datasources WHERE profile_id = p.id) as data_sources,
    (SELECT COUNT(*) FROM plan_scheduler_profile_constraints WHERE profile_id = p.id) as constraints,
    (SELECT COUNT(*) FROM plan_scheduler_profile_settings WHERE profile_id = p.id) as settings,
    p.createdon
FROM plan_scheduler_profiles p
WHERE p.id = 'demo-profile-001';
