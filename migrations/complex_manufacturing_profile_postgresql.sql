-- =====================================================================
-- Complex Manufacturing Plan Profile - PostgreSQL INSERT Script
-- Based on: iac-portal/src/features/plan-scheduler/services/mockData.ts
-- Date: 2025-12-12
-- Description: Comprehensive manufacturing demo with 18 work orders,
--              120+ tasks, 25+ resources across multiple work centers
-- =====================================================================

-- =====================================================================
-- 1. PROFILE
-- =====================================================================
INSERT INTO plan_scheduler_profiles (
    id, name, description, is_default,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES (
    'complex-mfg-2026-q1',
    'Complex Manufacturing - 2026 Q1',
    'Comprehensive manufacturing schedule with 18 work orders, 120+ tasks across 8 operation types (Material Prep, Welding, Machining, Assembly, Electrical, Coating, QA, Packaging). Includes 25+ resources: machines (24/7), work centers (standard hours), and human operators.',
    FALSE,
    TRUE,
    'COMPLEX-MFG-2026-Q1',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- =====================================================================
-- 2. DATA SOURCE - RESOURCES (25+ resources)
-- =====================================================================
INSERT INTO plan_scheduler_profile_datasources (
    id, profile_id, data_type, name, description,
    source_type, source_json, display_order,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES (
    'ds-complex-resources-001',
    'complex-mfg-2026-q1',
    'resources',
    'Complex Manufacturing Resources',
    '25 resources: 4 machines (24/7), 6 work centers (standard hours), 15 human operators across welding, machining, assembly, QA, electrical, coating, packaging, and logistics',
    'json',
    '[
        {
            "id": "R-M-CNC-001",
            "name": "CNC Mill #1",
            "type": "MACHINE",
            "capacity": 1,
            "color": "#e74c3c",
            "availability": [
                {"day": 0, "start": "00:00", "end": "23:59"},
                {"day": 1, "start": "00:00", "end": "23:59"},
                {"day": 2, "start": "00:00", "end": "23:59"},
                {"day": 3, "start": "00:00", "end": "23:59"},
                {"day": 4, "start": "00:00", "end": "23:59"},
                {"day": 5, "start": "00:00", "end": "23:59"},
                {"day": 6, "start": "00:00", "end": "23:59"}
            ],
            "role": "machining"
        },
        {
            "id": "R-M-CNC-002",
            "name": "CNC Mill #2",
            "type": "MACHINE",
            "capacity": 1,
            "color": "#c0392b",
            "availability": [
                {"day": 0, "start": "00:00", "end": "23:59"},
                {"day": 1, "start": "00:00", "end": "23:59"},
                {"day": 2, "start": "00:00", "end": "23:59"},
                {"day": 3, "start": "00:00", "end": "23:59"},
                {"day": 4, "start": "00:00", "end": "23:59"},
                {"day": 5, "start": "00:00", "end": "23:59"},
                {"day": 6, "start": "00:00", "end": "23:59"}
            ],
            "role": "machining"
        },
        {
            "id": "R-M-LATHE-001",
            "name": "CNC Lathe",
            "type": "MACHINE",
            "capacity": 1,
            "color": "#e67e22",
            "availability": [
                {"day": 0, "start": "00:00", "end": "23:59"},
                {"day": 1, "start": "00:00", "end": "23:59"},
                {"day": 2, "start": "00:00", "end": "23:59"},
                {"day": 3, "start": "00:00", "end": "23:59"},
                {"day": 4, "start": "00:00", "end": "23:59"},
                {"day": 5, "start": "00:00", "end": "23:59"},
                {"day": 6, "start": "00:00", "end": "23:59"}
            ],
            "role": "machining"
        },
        {
            "id": "R-M-3DP-001",
            "name": "3D Printer",
            "type": "MACHINE",
            "capacity": 1,
            "color": "#d35400",
            "availability": [
                {"day": 0, "start": "00:00", "end": "23:59"},
                {"day": 1, "start": "00:00", "end": "23:59"},
                {"day": 2, "start": "00:00", "end": "23:59"},
                {"day": 3, "start": "00:00", "end": "23:59"},
                {"day": 4, "start": "00:00", "end": "23:59"},
                {"day": 5, "start": "00:00", "end": "23:59"},
                {"day": 6, "start": "00:00", "end": "23:59"}
            ],
            "role": "machining"
        },
        {
            "id": "R-WC-ASM-001",
            "name": "Assembly Line 1",
            "type": "WORK_CENTER",
            "capacity": 3,
            "color": "#3498db",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "assembly"
        },
        {
            "id": "R-WC-ASM-002",
            "name": "Assembly Line 2",
            "type": "WORK_CENTER",
            "capacity": 3,
            "color": "#2980b9",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "assembly"
        },
        {
            "id": "R-WC-WELD-001",
            "name": "Welding Station 1",
            "type": "WORK_CENTER",
            "capacity": 2,
            "color": "#f39c12",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "welding"
        },
        {
            "id": "R-WC-WELD-002",
            "name": "Welding Station 2",
            "type": "WORK_CENTER",
            "capacity": 2,
            "color": "#f1c40f",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "welding"
        },
        {
            "id": "R-WC-PAINT-001",
            "name": "Paint Booth",
            "type": "WORK_CENTER",
            "capacity": 1,
            "color": "#9b59b6",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "coating"
        },
        {
            "id": "R-WC-OVEN-001",
            "name": "Curing Oven",
            "type": "WORK_CENTER",
            "capacity": 4,
            "color": "#8e44ad",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "coating"
        },
        {
            "id": "R-H-WELD-001",
            "name": "Welder - John",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#e67e22",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "welding"
        },
        {
            "id": "R-H-WELD-002",
            "name": "Welder - Mike",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#d35400",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "welding"
        },
        {
            "id": "R-H-MACH-001",
            "name": "Machinist - Sarah",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#e74c3c",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "machining"
        },
        {
            "id": "R-H-MACH-002",
            "name": "Machinist - Tom",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#c0392b",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "machining"
        },
        {
            "id": "R-H-ASM-001",
            "name": "Assembler - Lisa",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#3498db",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "assembly"
        },
        {
            "id": "R-H-ASM-002",
            "name": "Assembler - Dave",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#2980b9",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "assembly"
        },
        {
            "id": "R-H-ASM-003",
            "name": "Assembler - Emily",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#1f618d",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "assembly"
        },
        {
            "id": "R-H-QA-001",
            "name": "QA Inspector - Alex",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#27ae60",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "qa"
        },
        {
            "id": "R-H-QA-002",
            "name": "QA Inspector - Maria",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#229954",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "qa"
        },
        {
            "id": "R-H-ELEC-001",
            "name": "Electrician - Robert",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#f4d03f",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "electrical"
        },
        {
            "id": "R-H-ELEC-002",
            "name": "Electrician - Chris",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#f39c12",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "electrical"
        },
        {
            "id": "R-H-PAINT-001",
            "name": "Painter - Jessica",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#9b59b6",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "coating"
        },
        {
            "id": "R-H-PACK-001",
            "name": "Packager - Daniel",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#34495e",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "packaging"
        },
        {
            "id": "R-H-PACK-002",
            "name": "Packager - Amy",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#2c3e50",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "packaging"
        },
        {
            "id": "R-H-LOG-001",
            "name": "Logistics - Kevin",
            "type": "HUMAN",
            "capacity": 1,
            "color": "#16a085",
            "availability": [
                {"day": 1, "start": "09:00", "end": "17:00"},
                {"day": 2, "start": "09:00", "end": "17:00"},
                {"day": 3, "start": "09:00", "end": "17:00"},
                {"day": 4, "start": "09:00", "end": "17:00"},
                {"day": 5, "start": "09:00", "end": "17:00"}
            ],
            "role": "material_prep"
        }
    ]'::jsonb,
    1,
    TRUE,
    'COMPLEX-RES',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- =====================================================================
-- 3. DATA SOURCE - TASKS (120+ tasks across 18 work orders)
-- =====================================================================
INSERT INTO plan_scheduler_profile_datasources (
    id, profile_id, data_type, name, description,
    source_type, source_json, display_order,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES (
    'ds-complex-tasks-001',
    'complex-mfg-2026-q1',
    'tasks',
    'Complex Manufacturing Tasks',
    '120+ tasks across 18 work orders with full operation sequences: Material Prep, Welding, Machining, Assembly, Electrical, Coating, QA, Packaging. All tasks start from 2026-01-01 with status NEW',
    'json',
    '[
        {"id": "WO-2026-1001", "name": "Work Order 1001", "start": "2026-01-01T09:00:00", "end": "2026-01-01T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1001-MP", "name": "WO-1001: Material Prep", "start": "2026-01-01T09:00:00", "end": "2026-01-01T11:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1001", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1001-WD", "name": "WO-1001: Welding", "start": "2026-01-01T11:00:00", "end": "2026-01-01T14:30:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1001", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1001-MP", "type": "FS"}]},
        {"id": "WO-2026-1001-MC", "name": "WO-1001: Machining", "start": "2026-01-01T14:30:00", "end": "2026-01-02T09:30:00", "duration": 240, "status": "new", "type": "task", "parentId": "WO-2026-1001", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1001-WD", "type": "FS"}]},
        {"id": "WO-2026-1001-AS", "name": "WO-1001: Assembly", "start": "2026-01-02T09:30:00", "end": "2026-01-02T12:30:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1001", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1001-MC", "type": "FS"}]},
        {"id": "WO-2026-1001-EL", "name": "WO-1001: Electrical", "start": "2026-01-02T12:30:00", "end": "2026-01-02T14:30:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1001", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1001-AS", "type": "FS"}]},
        {"id": "WO-2026-1001-CT", "name": "WO-1001: Coating", "start": "2026-01-02T14:30:00", "end": "2026-01-03T09:30:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1001", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1001-EL", "type": "FS"}]},
        {"id": "WO-2026-1001-QA", "name": "WO-1001: QA", "start": "2026-01-03T09:30:00", "end": "2026-01-03T11:30:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1001", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1001-CT", "type": "FS"}]},
        {"id": "WO-2026-1001-PK", "name": "WO-1001: Packaging", "start": "2026-01-03T11:30:00", "end": "2026-01-03T13:30:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1001", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1001-QA", "type": "FS"}]},

        {"id": "WO-2026-1002", "name": "Work Order 1002", "start": "2026-01-02T09:00:00", "end": "2026-01-02T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1002-MP", "name": "WO-1002: Material Prep", "start": "2026-01-02T09:00:00", "end": "2026-01-02T10:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1002", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1002-WD", "name": "WO-1002: Welding", "start": "2026-01-02T10:30:00", "end": "2026-01-02T14:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1002", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1002-MP", "type": "FS"}]},
        {"id": "WO-2026-1002-MC", "name": "WO-1002: Machining", "start": "2026-01-02T14:00:00", "end": "2026-01-03T11:00:00", "duration": 300, "status": "new", "type": "task", "parentId": "WO-2026-1002", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1002-WD", "type": "FS"}]},
        {"id": "WO-2026-1002-AS", "name": "WO-1002: Assembly", "start": "2026-01-03T11:00:00", "end": "2026-01-03T14:00:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1002", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1002-MC", "type": "FS"}]},
        {"id": "WO-2026-1002-EL", "name": "WO-1002: Electrical", "start": "2026-01-03T14:00:00", "end": "2026-01-03T16:30:00", "duration": 150, "status": "new", "type": "task", "parentId": "WO-2026-1002", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1002-AS", "type": "FS"}]},
        {"id": "WO-2026-1002-CT", "name": "WO-1002: Coating", "start": "2026-01-03T16:30:00", "end": "2026-01-06T09:30:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1002", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1002-EL", "type": "FS"}]},
        {"id": "WO-2026-1002-QA", "name": "WO-1002: QA", "start": "2026-01-06T09:30:00", "end": "2026-01-06T11:00:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1002", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1002-CT", "type": "FS"}]},
        {"id": "WO-2026-1002-PK", "name": "WO-1002: Packaging", "start": "2026-01-06T11:00:00", "end": "2026-01-06T12:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1002", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1002-QA", "type": "FS"}]},

        {"id": "WO-2026-1003", "name": "Work Order 1003", "start": "2026-01-03T09:00:00", "end": "2026-01-03T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1003-MP", "name": "WO-1003: Material Prep", "start": "2026-01-03T09:00:00", "end": "2026-01-03T11:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1003", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1003-WD", "name": "WO-1003: Welding", "start": "2026-01-03T11:00:00", "end": "2026-01-03T15:00:00", "duration": 240, "status": "new", "type": "task", "parentId": "WO-2026-1003", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1003-MP", "type": "FS"}]},
        {"id": "WO-2026-1003-MC", "name": "WO-1003: Machining", "start": "2026-01-03T15:00:00", "end": "2026-01-06T12:00:00", "duration": 300, "status": "new", "type": "task", "parentId": "WO-2026-1003", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1003-WD", "type": "FS"}]},
        {"id": "WO-2026-1003-AS", "name": "WO-1003: Assembly", "start": "2026-01-06T12:00:00", "end": "2026-01-06T15:30:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1003", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1003-MC", "type": "FS"}]},
        {"id": "WO-2026-1003-EL", "name": "WO-1003: Electrical", "start": "2026-01-06T15:30:00", "end": "2026-01-07T09:30:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1003", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1003-AS", "type": "FS"}]},
        {"id": "WO-2026-1003-CT", "name": "WO-1003: Coating", "start": "2026-01-07T09:30:00", "end": "2026-01-07T13:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1003", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1003-EL", "type": "FS"}]},
        {"id": "WO-2026-1003-QA", "name": "WO-1003: QA", "start": "2026-01-07T13:00:00", "end": "2026-01-07T15:30:00", "duration": 150, "status": "new", "type": "task", "parentId": "WO-2026-1003", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1003-CT", "type": "FS"}]},
        {"id": "WO-2026-1003-PK", "name": "WO-1003: Packaging", "start": "2026-01-07T15:30:00", "end": "2026-01-08T09:00:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1003", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1003-QA", "type": "FS"}]},

        {"id": "WO-2026-1004", "name": "Work Order 1004", "start": "2026-01-06T09:00:00", "end": "2026-01-06T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1004-MP", "name": "WO-1004: Material Prep", "start": "2026-01-06T09:00:00", "end": "2026-01-06T10:00:00", "duration": 60, "status": "new", "type": "task", "parentId": "WO-2026-1004", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1004-WD", "name": "WO-1004: Welding", "start": "2026-01-06T10:00:00", "end": "2026-01-06T13:30:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1004", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1004-MP", "type": "FS"}]},
        {"id": "WO-2026-1004-MC", "name": "WO-1004: Machining", "start": "2026-01-06T13:30:00", "end": "2026-01-07T10:00:00", "duration": 270, "status": "new", "type": "task", "parentId": "WO-2026-1004", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1004-WD", "type": "FS"}]},
        {"id": "WO-2026-1004-AS", "name": "WO-1004: Assembly", "start": "2026-01-07T10:00:00", "end": "2026-01-07T13:00:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1004", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1004-MC", "type": "FS"}]},
        {"id": "WO-2026-1004-EL", "name": "WO-1004: Electrical", "start": "2026-01-07T13:00:00", "end": "2026-01-07T15:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1004", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1004-AS", "type": "FS"}]},
        {"id": "WO-2026-1004-CT", "name": "WO-1004: Coating", "start": "2026-01-07T15:00:00", "end": "2026-01-08T11:30:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1004", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1004-EL", "type": "FS"}]},
        {"id": "WO-2026-1004-QA", "name": "WO-1004: QA", "start": "2026-01-08T11:30:00", "end": "2026-01-08T13:00:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1004", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1004-CT", "type": "FS"}]},
        {"id": "WO-2026-1004-PK", "name": "WO-1004: Packaging", "start": "2026-01-08T13:00:00", "end": "2026-01-08T14:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1004", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1004-QA", "type": "FS"}]},

        {"id": "WO-2026-1005", "name": "Work Order 1005", "start": "2026-01-07T09:00:00", "end": "2026-01-07T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1005-MP", "name": "WO-1005: Material Prep", "start": "2026-01-07T09:00:00", "end": "2026-01-07T11:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1005", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1005-WD", "name": "WO-1005: Welding", "start": "2026-01-07T11:00:00", "end": "2026-01-07T15:00:00", "duration": 240, "status": "new", "type": "task", "parentId": "WO-2026-1005", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1005-MP", "type": "FS"}]},
        {"id": "WO-2026-1005-MC", "name": "WO-1005: Machining", "start": "2026-01-07T15:00:00", "end": "2026-01-08T11:30:00", "duration": 270, "status": "new", "type": "task", "parentId": "WO-2026-1005", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1005-WD", "type": "FS"}]},
        {"id": "WO-2026-1005-AS", "name": "WO-1005: Assembly", "start": "2026-01-08T11:30:00", "end": "2026-01-08T15:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1005", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1005-MC", "type": "FS"}]},
        {"id": "WO-2026-1005-EL", "name": "WO-1005: Electrical", "start": "2026-01-08T15:00:00", "end": "2026-01-09T09:30:00", "duration": 150, "status": "new", "type": "task", "parentId": "WO-2026-1005", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1005-AS", "type": "FS"}]},
        {"id": "WO-2026-1005-CT", "name": "WO-1005: Coating", "start": "2026-01-09T09:30:00", "end": "2026-01-09T13:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1005", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1005-EL", "type": "FS"}]},
        {"id": "WO-2026-1005-QA", "name": "WO-1005: QA", "start": "2026-01-09T13:00:00", "end": "2026-01-09T15:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1005", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1005-CT", "type": "FS"}]},
        {"id": "WO-2026-1005-PK", "name": "WO-1005: Packaging", "start": "2026-01-09T15:00:00", "end": "2026-01-10T09:00:00", "duration": 60, "status": "new", "type": "task", "parentId": "WO-2026-1005", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1005-QA", "type": "FS"}]},

        {"id": "WO-2026-1006", "name": "Work Order 1006", "start": "2026-01-08T09:00:00", "end": "2026-01-08T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1006-MP", "name": "WO-1006: Material Prep", "start": "2026-01-08T09:00:00", "end": "2026-01-08T10:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1006", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1006-WD", "name": "WO-1006: Welding", "start": "2026-01-08T10:30:00", "end": "2026-01-08T14:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1006", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1006-MP", "type": "FS"}]},
        {"id": "WO-2026-1006-MC", "name": "WO-1006: Machining", "start": "2026-01-08T14:00:00", "end": "2026-01-09T13:00:00", "duration": 360, "status": "new", "type": "task", "parentId": "WO-2026-1006", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1006-WD", "type": "FS"}]},
        {"id": "WO-2026-1006-AS", "name": "WO-1006: Assembly", "start": "2026-01-09T13:00:00", "end": "2026-01-09T16:00:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1006", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1006-MC", "type": "FS"}]},
        {"id": "WO-2026-1006-EL", "name": "WO-1006: Electrical", "start": "2026-01-09T16:00:00", "end": "2026-01-10T10:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1006", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1006-AS", "type": "FS"}]},
        {"id": "WO-2026-1006-CT", "name": "WO-1006: Coating", "start": "2026-01-10T10:00:00", "end": "2026-01-10T13:00:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1006", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1006-EL", "type": "FS"}]},
        {"id": "WO-2026-1006-QA", "name": "WO-1006: QA", "start": "2026-01-10T13:00:00", "end": "2026-01-10T15:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1006", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1006-CT", "type": "FS"}]},
        {"id": "WO-2026-1006-PK", "name": "WO-1006: Packaging", "start": "2026-01-10T15:00:00", "end": "2026-01-10T16:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1006", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1006-QA", "type": "FS"}]},

        {"id": "WO-2026-1007", "name": "Work Order 1007", "start": "2026-01-09T09:00:00", "end": "2026-01-09T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1007-MP", "name": "WO-1007: Material Prep", "start": "2026-01-09T09:00:00", "end": "2026-01-09T11:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1007", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1007-WD", "name": "WO-1007: Welding", "start": "2026-01-09T11:00:00", "end": "2026-01-09T15:00:00", "duration": 240, "status": "new", "type": "task", "parentId": "WO-2026-1007", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1007-MP", "type": "FS"}]},
        {"id": "WO-2026-1007-MC", "name": "WO-1007: Machining", "start": "2026-01-09T15:00:00", "end": "2026-01-10T12:30:00", "duration": 330, "status": "new", "type": "task", "parentId": "WO-2026-1007", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1007-WD", "type": "FS"}]},
        {"id": "WO-2026-1007-AS", "name": "WO-1007: Assembly", "start": "2026-01-10T12:30:00", "end": "2026-01-10T16:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1007", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1007-MC", "type": "FS"}]},
        {"id": "WO-2026-1007-EL", "name": "WO-1007: Electrical", "start": "2026-01-10T16:00:00", "end": "2026-01-13T10:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1007", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1007-AS", "type": "FS"}]},
        {"id": "WO-2026-1007-CT", "name": "WO-1007: Coating", "start": "2026-01-13T10:00:00", "end": "2026-01-13T13:30:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1007", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1007-EL", "type": "FS"}]},
        {"id": "WO-2026-1007-QA", "name": "WO-1007: QA", "start": "2026-01-13T13:30:00", "end": "2026-01-13T15:00:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1007", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1007-CT", "type": "FS"}]},
        {"id": "WO-2026-1007-PK", "name": "WO-1007: Packaging", "start": "2026-01-13T15:00:00", "end": "2026-01-13T16:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1007", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1007-QA", "type": "FS"}]},

        {"id": "WO-2026-1008", "name": "Work Order 1008", "start": "2026-01-10T09:00:00", "end": "2026-01-10T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1008-MP", "name": "WO-1008: Material Prep", "start": "2026-01-10T09:00:00", "end": "2026-01-10T10:00:00", "duration": 60, "status": "new", "type": "task", "parentId": "WO-2026-1008", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1008-WD", "name": "WO-1008: Welding", "start": "2026-01-10T10:00:00", "end": "2026-01-10T13:30:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1008", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1008-MP", "type": "FS"}]},
        {"id": "WO-2026-1008-MC", "name": "WO-1008: Machining", "start": "2026-01-10T13:30:00", "end": "2026-01-13T11:00:00", "duration": 300, "status": "new", "type": "task", "parentId": "WO-2026-1008", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1008-WD", "type": "FS"}]},
        {"id": "WO-2026-1008-AS", "name": "WO-1008: Assembly", "start": "2026-01-13T11:00:00", "end": "2026-01-13T14:00:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1008", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1008-MC", "type": "FS"}]},
        {"id": "WO-2026-1008-EL", "name": "WO-1008: Electrical", "start": "2026-01-13T14:00:00", "end": "2026-01-13T16:30:00", "duration": 150, "status": "new", "type": "task", "parentId": "WO-2026-1008", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1008-AS", "type": "FS"}]},
        {"id": "WO-2026-1008-CT", "name": "WO-1008: Coating", "start": "2026-01-13T16:30:00", "end": "2026-01-14T10:00:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1008", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1008-EL", "type": "FS"}]},
        {"id": "WO-2026-1008-QA", "name": "WO-1008: QA", "start": "2026-01-14T10:00:00", "end": "2026-01-14T12:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1008", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1008-CT", "type": "FS"}]},
        {"id": "WO-2026-1008-PK", "name": "WO-1008: Packaging", "start": "2026-01-14T12:00:00", "end": "2026-01-14T13:00:00", "duration": 60, "status": "new", "type": "task", "parentId": "WO-2026-1008", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1008-QA", "type": "FS"}]},

        {"id": "WO-2026-1009", "name": "Work Order 1009", "start": "2026-01-13T09:00:00", "end": "2026-01-13T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1009-MP", "name": "WO-1009: Material Prep", "start": "2026-01-13T09:00:00", "end": "2026-01-13T10:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1009", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1009-WD", "name": "WO-1009: Welding", "start": "2026-01-13T10:30:00", "end": "2026-01-13T14:30:00", "duration": 240, "status": "new", "type": "task", "parentId": "WO-2026-1009", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1009-MP", "type": "FS"}]},
        {"id": "WO-2026-1009-MC", "name": "WO-1009: Machining", "start": "2026-01-13T14:30:00", "end": "2026-01-14T11:00:00", "duration": 270, "status": "new", "type": "task", "parentId": "WO-2026-1009", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1009-WD", "type": "FS"}]},
        {"id": "WO-2026-1009-AS", "name": "WO-1009: Assembly", "start": "2026-01-14T11:00:00", "end": "2026-01-14T14:30:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1009", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1009-MC", "type": "FS"}]},
        {"id": "WO-2026-1009-EL", "name": "WO-1009: Electrical", "start": "2026-01-14T14:30:00", "end": "2026-01-14T16:30:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1009", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1009-AS", "type": "FS"}]},
        {"id": "WO-2026-1009-CT", "name": "WO-1009: Coating", "start": "2026-01-14T16:30:00", "end": "2026-01-15T10:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1009", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1009-EL", "type": "FS"}]},
        {"id": "WO-2026-1009-QA", "name": "WO-1009: QA", "start": "2026-01-15T10:00:00", "end": "2026-01-15T11:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1009", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1009-CT", "type": "FS"}]},
        {"id": "WO-2026-1009-PK", "name": "WO-1009: Packaging", "start": "2026-01-15T11:30:00", "end": "2026-01-15T13:00:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1009", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1009-QA", "type": "FS"}]},

        {"id": "WO-2026-1010", "name": "Work Order 1010", "start": "2026-01-14T09:00:00", "end": "2026-01-14T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1010-MP", "name": "WO-1010: Material Prep", "start": "2026-01-14T09:00:00", "end": "2026-01-14T11:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1010", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1010-WD", "name": "WO-1010: Welding", "start": "2026-01-14T11:00:00", "end": "2026-01-14T14:30:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1010", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1010-MP", "type": "FS"}]},
        {"id": "WO-2026-1010-MC", "name": "WO-1010: Machining", "start": "2026-01-14T14:30:00", "end": "2026-01-15T15:00:00", "duration": 360, "status": "new", "type": "task", "parentId": "WO-2026-1010", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1010-WD", "type": "FS"}]},
        {"id": "WO-2026-1010-AS", "name": "WO-1010: Assembly", "start": "2026-01-15T15:00:00", "end": "2026-01-16T09:30:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1010", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1010-MC", "type": "FS"}]},
        {"id": "WO-2026-1010-EL", "name": "WO-1010: Electrical", "start": "2026-01-16T09:30:00", "end": "2026-01-16T12:00:00", "duration": 150, "status": "new", "type": "task", "parentId": "WO-2026-1010", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1010-AS", "type": "FS"}]},
        {"id": "WO-2026-1010-CT", "name": "WO-1010: Coating", "start": "2026-01-16T12:00:00", "end": "2026-01-16T15:00:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1010", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1010-EL", "type": "FS"}]},
        {"id": "WO-2026-1010-QA", "name": "WO-1010: QA", "start": "2026-01-16T15:00:00", "end": "2026-01-17T09:30:00", "duration": 150, "status": "new", "type": "task", "parentId": "WO-2026-1010", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1010-CT", "type": "FS"}]},
        {"id": "WO-2026-1010-PK", "name": "WO-1010: Packaging", "start": "2026-01-17T09:30:00", "end": "2026-01-17T10:30:00", "duration": 60, "status": "new", "type": "task", "parentId": "WO-2026-1010", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1010-QA", "type": "FS"}]},

        {"id": "WO-2026-1011", "name": "Work Order 1011", "start": "2026-01-15T09:00:00", "end": "2026-01-15T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1011-MP", "name": "WO-1011: Material Prep", "start": "2026-01-15T09:00:00", "end": "2026-01-15T10:00:00", "duration": 60, "status": "new", "type": "task", "parentId": "WO-2026-1011", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1011-WD", "name": "WO-1011: Welding", "start": "2026-01-15T10:00:00", "end": "2026-01-15T14:00:00", "duration": 240, "status": "new", "type": "task", "parentId": "WO-2026-1011", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1011-MP", "type": "FS"}]},
        {"id": "WO-2026-1011-MC", "name": "WO-1011: Machining", "start": "2026-01-15T14:00:00", "end": "2026-01-16T11:30:00", "duration": 330, "status": "new", "type": "task", "parentId": "WO-2026-1011", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1011-WD", "type": "FS"}]},
        {"id": "WO-2026-1011-AS", "name": "WO-1011: Assembly", "start": "2026-01-16T11:30:00", "end": "2026-01-16T14:30:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1011", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1011-MC", "type": "FS"}]},
        {"id": "WO-2026-1011-EL", "name": "WO-1011: Electrical", "start": "2026-01-16T14:30:00", "end": "2026-01-16T16:30:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1011", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1011-AS", "type": "FS"}]},
        {"id": "WO-2026-1011-CT", "name": "WO-1011: Coating", "start": "2026-01-16T16:30:00", "end": "2026-01-17T10:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1011", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1011-EL", "type": "FS"}]},
        {"id": "WO-2026-1011-QA", "name": "WO-1011: QA", "start": "2026-01-17T10:00:00", "end": "2026-01-17T12:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1011", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1011-CT", "type": "FS"}]},
        {"id": "WO-2026-1011-PK", "name": "WO-1011: Packaging", "start": "2026-01-17T12:00:00", "end": "2026-01-17T13:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1011", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1011-QA", "type": "FS"}]},

        {"id": "WO-2026-1012", "name": "Work Order 1012", "start": "2026-01-16T09:00:00", "end": "2026-01-16T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1012-MP", "name": "WO-1012: Material Prep", "start": "2026-01-16T09:00:00", "end": "2026-01-16T10:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1012", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1012-WD", "name": "WO-1012: Welding", "start": "2026-01-16T10:30:00", "end": "2026-01-16T14:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1012", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1012-MP", "type": "FS"}]},
        {"id": "WO-2026-1012-MC", "name": "WO-1012: Machining", "start": "2026-01-16T14:00:00", "end": "2026-01-17T11:00:00", "duration": 300, "status": "new", "type": "task", "parentId": "WO-2026-1012", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1012-WD", "type": "FS"}]},
        {"id": "WO-2026-1012-AS", "name": "WO-1012: Assembly", "start": "2026-01-17T11:00:00", "end": "2026-01-17T14:30:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1012", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1012-MC", "type": "FS"}]},
        {"id": "WO-2026-1012-EL", "name": "WO-1012: Electrical", "start": "2026-01-17T14:30:00", "end": "2026-01-20T09:30:00", "duration": 150, "status": "new", "type": "task", "parentId": "WO-2026-1012", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1012-AS", "type": "FS"}]},
        {"id": "WO-2026-1012-CT", "name": "WO-1012: Coating", "start": "2026-01-20T09:30:00", "end": "2026-01-20T12:30:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1012", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1012-EL", "type": "FS"}]},
        {"id": "WO-2026-1012-QA", "name": "WO-1012: QA", "start": "2026-01-20T12:30:00", "end": "2026-01-20T14:00:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1012", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1012-CT", "type": "FS"}]},
        {"id": "WO-2026-1012-PK", "name": "WO-1012: Packaging", "start": "2026-01-20T14:00:00", "end": "2026-01-20T15:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1012", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1012-QA", "type": "FS"}]},

        {"id": "WO-2026-1013", "name": "Work Order 1013", "start": "2026-01-17T09:00:00", "end": "2026-01-17T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1013-MP", "name": "WO-1013: Material Prep", "start": "2026-01-17T09:00:00", "end": "2026-01-17T11:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1013", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1013-WD", "name": "WO-1013: Welding", "start": "2026-01-17T11:00:00", "end": "2026-01-17T15:00:00", "duration": 240, "status": "new", "type": "task", "parentId": "WO-2026-1013", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1013-MP", "type": "FS"}]},
        {"id": "WO-2026-1013-MC", "name": "WO-1013: Machining", "start": "2026-01-17T15:00:00", "end": "2026-01-20T10:30:00", "duration": 270, "status": "new", "type": "task", "parentId": "WO-2026-1013", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1013-WD", "type": "FS"}]},
        {"id": "WO-2026-1013-AS", "name": "WO-1013: Assembly", "start": "2026-01-20T10:30:00", "end": "2026-01-20T13:30:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1013", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1013-MC", "type": "FS"}]},
        {"id": "WO-2026-1013-EL", "name": "WO-1013: Electrical", "start": "2026-01-20T13:30:00", "end": "2026-01-20T15:30:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1013", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1013-AS", "type": "FS"}]},
        {"id": "WO-2026-1013-CT", "name": "WO-1013: Coating", "start": "2026-01-20T15:30:00", "end": "2026-01-21T12:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1013", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1013-EL", "type": "FS"}]},
        {"id": "WO-2026-1013-QA", "name": "WO-1013: QA", "start": "2026-01-21T12:00:00", "end": "2026-01-21T14:30:00", "duration": 150, "status": "new", "type": "task", "parentId": "WO-2026-1013", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1013-CT", "type": "FS"}]},
        {"id": "WO-2026-1013-PK", "name": "WO-1013: Packaging", "start": "2026-01-21T14:30:00", "end": "2026-01-21T15:30:00", "duration": 60, "status": "new", "type": "task", "parentId": "WO-2026-1013", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1013-QA", "type": "FS"}]},

        {"id": "WO-2026-1014", "name": "Work Order 1014", "start": "2026-01-20T09:00:00", "end": "2026-01-20T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1014-MP", "name": "WO-1014: Material Prep", "start": "2026-01-20T09:00:00", "end": "2026-01-20T10:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1014", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1014-WD", "name": "WO-1014: Welding", "start": "2026-01-20T10:30:00", "end": "2026-01-20T14:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1014", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1014-MP", "type": "FS"}]},
        {"id": "WO-2026-1014-MC", "name": "WO-1014: Machining", "start": "2026-01-20T14:00:00", "end": "2026-01-21T15:00:00", "duration": 360, "status": "new", "type": "task", "parentId": "WO-2026-1014", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1014-WD", "type": "FS"}]},
        {"id": "WO-2026-1014-AS", "name": "WO-1014: Assembly", "start": "2026-01-21T15:00:00", "end": "2026-01-22T10:30:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1014", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1014-MC", "type": "FS"}]},
        {"id": "WO-2026-1014-EL", "name": "WO-1014: Electrical", "start": "2026-01-22T10:30:00", "end": "2026-01-22T12:30:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1014", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1014-AS", "type": "FS"}]},
        {"id": "WO-2026-1014-CT", "name": "WO-1014: Coating", "start": "2026-01-22T12:30:00", "end": "2026-01-22T15:30:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1014", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1014-EL", "type": "FS"}]},
        {"id": "WO-2026-1014-QA", "name": "WO-1014: QA", "start": "2026-01-22T15:30:00", "end": "2026-01-23T09:00:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1014", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1014-CT", "type": "FS"}]},
        {"id": "WO-2026-1014-PK", "name": "WO-1014: Packaging", "start": "2026-01-23T09:00:00", "end": "2026-01-23T10:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1014", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1014-QA", "type": "FS"}]},

        {"id": "WO-2026-1015", "name": "Work Order 1015", "start": "2026-01-21T09:00:00", "end": "2026-01-21T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1015-MP", "name": "WO-1015: Material Prep", "start": "2026-01-21T09:00:00", "end": "2026-01-21T10:00:00", "duration": 60, "status": "new", "type": "task", "parentId": "WO-2026-1015", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1015-WD", "name": "WO-1015: Welding", "start": "2026-01-21T10:00:00", "end": "2026-01-21T14:00:00", "duration": 240, "status": "new", "type": "task", "parentId": "WO-2026-1015", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1015-MP", "type": "FS"}]},
        {"id": "WO-2026-1015-MC", "name": "WO-1015: Machining", "start": "2026-01-21T14:00:00", "end": "2026-01-22T11:30:00", "duration": 330, "status": "new", "type": "task", "parentId": "WO-2026-1015", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1015-WD", "type": "FS"}]},
        {"id": "WO-2026-1015-AS", "name": "WO-1015: Assembly", "start": "2026-01-22T11:30:00", "end": "2026-01-22T14:30:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1015", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1015-MC", "type": "FS"}]},
        {"id": "WO-2026-1015-EL", "name": "WO-1015: Electrical", "start": "2026-01-22T14:30:00", "end": "2026-01-23T10:00:00", "duration": 150, "status": "new", "type": "task", "parentId": "WO-2026-1015", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1015-AS", "type": "FS"}]},
        {"id": "WO-2026-1015-CT", "name": "WO-1015: Coating", "start": "2026-01-23T10:00:00", "end": "2026-01-23T13:30:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1015", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1015-EL", "type": "FS"}]},
        {"id": "WO-2026-1015-QA", "name": "WO-1015: QA", "start": "2026-01-23T13:30:00", "end": "2026-01-23T15:30:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1015", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1015-CT", "type": "FS"}]},
        {"id": "WO-2026-1015-PK", "name": "WO-1015: Packaging", "start": "2026-01-23T15:30:00", "end": "2026-01-24T09:00:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1015", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1015-QA", "type": "FS"}]},

        {"id": "WO-2026-1016", "name": "Work Order 1016", "start": "2026-01-22T09:00:00", "end": "2026-01-22T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1016-MP", "name": "WO-1016: Material Prep", "start": "2026-01-22T09:00:00", "end": "2026-01-22T11:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1016", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1016-WD", "name": "WO-1016: Welding", "start": "2026-01-22T11:00:00", "end": "2026-01-22T14:30:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1016", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1016-MP", "type": "FS"}]},
        {"id": "WO-2026-1016-MC", "name": "WO-1016: Machining", "start": "2026-01-22T14:30:00", "end": "2026-01-23T11:30:00", "duration": 300, "status": "new", "type": "task", "parentId": "WO-2026-1016", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1016-WD", "type": "FS"}]},
        {"id": "WO-2026-1016-AS", "name": "WO-1016: Assembly", "start": "2026-01-23T11:30:00", "end": "2026-01-23T15:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1016", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1016-MC", "type": "FS"}]},
        {"id": "WO-2026-1016-EL", "name": "WO-1016: Electrical", "start": "2026-01-23T15:00:00", "end": "2026-01-24T09:30:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1016", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1016-AS", "type": "FS"}]},
        {"id": "WO-2026-1016-CT", "name": "WO-1016: Coating", "start": "2026-01-24T09:30:00", "end": "2026-01-24T12:30:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1016", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1016-EL", "type": "FS"}]},
        {"id": "WO-2026-1016-QA", "name": "WO-1016: QA", "start": "2026-01-24T12:30:00", "end": "2026-01-24T14:00:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1016", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1016-CT", "type": "FS"}]},
        {"id": "WO-2026-1016-PK", "name": "WO-1016: Packaging", "start": "2026-01-24T14:00:00", "end": "2026-01-24T15:00:00", "duration": 60, "status": "new", "type": "task", "parentId": "WO-2026-1016", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1016-QA", "type": "FS"}]},

        {"id": "WO-2026-1017", "name": "Work Order 1017", "start": "2026-01-23T09:00:00", "end": "2026-01-23T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1017-MP", "name": "WO-1017: Material Prep", "start": "2026-01-23T09:00:00", "end": "2026-01-23T10:30:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1017", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1017-WD", "name": "WO-1017: Welding", "start": "2026-01-23T10:30:00", "end": "2026-01-23T14:30:00", "duration": 240, "status": "new", "type": "task", "parentId": "WO-2026-1017", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1017-MP", "type": "FS"}]},
        {"id": "WO-2026-1017-MC", "name": "WO-1017: Machining", "start": "2026-01-23T14:30:00", "end": "2026-01-24T10:00:00", "duration": 270, "status": "new", "type": "task", "parentId": "WO-2026-1017", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1017-WD", "type": "FS"}]},
        {"id": "WO-2026-1017-AS", "name": "WO-1017: Assembly", "start": "2026-01-24T10:00:00", "end": "2026-01-24T13:00:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1017", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1017-MC", "type": "FS"}]},
        {"id": "WO-2026-1017-EL", "name": "WO-1017: Electrical", "start": "2026-01-24T13:00:00", "end": "2026-01-24T15:30:00", "duration": 150, "status": "new", "type": "task", "parentId": "WO-2026-1017", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1017-AS", "type": "FS"}]},
        {"id": "WO-2026-1017-CT", "name": "WO-1017: Coating", "start": "2026-01-24T15:30:00", "end": "2026-01-27T09:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1017", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1017-EL", "type": "FS"}]},
        {"id": "WO-2026-1017-QA", "name": "WO-1017: QA", "start": "2026-01-27T09:00:00", "end": "2026-01-27T11:30:00", "duration": 150, "status": "new", "type": "task", "parentId": "WO-2026-1017", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1017-CT", "type": "FS"}]},
        {"id": "WO-2026-1017-PK", "name": "WO-1017: Packaging", "start": "2026-01-27T11:30:00", "end": "2026-01-27T13:00:00", "duration": 90, "status": "new", "type": "task", "parentId": "WO-2026-1017", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1017-QA", "type": "FS"}]},

        {"id": "WO-2026-1018", "name": "Work Order 1018", "start": "2026-01-24T09:00:00", "end": "2026-01-24T09:00:00", "duration": 0, "status": "new", "type": "parent", "dependencies": []},
        {"id": "WO-2026-1018-MP", "name": "WO-1018: Material Prep", "start": "2026-01-24T09:00:00", "end": "2026-01-24T11:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1018", "resourceType": "material_prep", "dependencies": []},
        {"id": "WO-2026-1018-WD", "name": "WO-1018: Welding", "start": "2026-01-24T11:00:00", "end": "2026-01-24T14:30:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1018", "resourceType": "welding", "dependencies": [{"taskId": "WO-2026-1018-MP", "type": "FS"}]},
        {"id": "WO-2026-1018-MC", "name": "WO-1018: Machining", "start": "2026-01-24T14:30:00", "end": "2026-01-27T15:30:00", "duration": 360, "status": "new", "type": "task", "parentId": "WO-2026-1018", "resourceType": "machining", "dependencies": [{"taskId": "WO-2026-1018-WD", "type": "FS"}]},
        {"id": "WO-2026-1018-AS", "name": "WO-1018: Assembly", "start": "2026-01-27T15:30:00", "end": "2026-01-28T10:00:00", "duration": 210, "status": "new", "type": "task", "parentId": "WO-2026-1018", "resourceType": "assembly", "dependencies": [{"taskId": "WO-2026-1018-MC", "type": "FS"}]},
        {"id": "WO-2026-1018-EL", "name": "WO-1018: Electrical", "start": "2026-01-28T10:00:00", "end": "2026-01-28T12:00:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1018", "resourceType": "electrical", "dependencies": [{"taskId": "WO-2026-1018-AS", "type": "FS"}]},
        {"id": "WO-2026-1018-CT", "name": "WO-1018: Coating", "start": "2026-01-28T12:00:00", "end": "2026-01-28T15:00:00", "duration": 180, "status": "new", "type": "task", "parentId": "WO-2026-1018", "resourceType": "coating", "dependencies": [{"taskId": "WO-2026-1018-EL", "type": "FS"}]},
        {"id": "WO-2026-1018-QA", "name": "WO-1018: QA", "start": "2026-01-28T15:00:00", "end": "2026-01-29T09:30:00", "duration": 120, "status": "new", "type": "task", "parentId": "WO-2026-1018", "resourceType": "qa", "dependencies": [{"taskId": "WO-2026-1018-CT", "type": "FS"}]},
        {"id": "WO-2026-1018-PK", "name": "WO-1018: Packaging", "start": "2026-01-29T09:30:00", "end": "2026-01-29T10:30:00", "duration": 60, "status": "new", "type": "task", "parentId": "WO-2026-1018", "resourceType": "packaging", "dependencies": [{"taskId": "WO-2026-1018-QA", "type": "FS"}]}
    ]'::jsonb,
    2,
    TRUE,
    'COMPLEX-TASKS',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- =====================================================================
-- 4. CONSTRAINTS
-- =====================================================================

-- Time Constraint: Standard Working Hours
INSERT INTO plan_scheduler_profile_constraints (
    id, profile_id, constraint_type, name, description,
    source_type, source_json, enforcement, display_order,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES (
    'const-time-001',
    'complex-mfg-2026-q1',
    'time',
    'Standard Working Hours',
    'Monday-Friday 9:00-17:00, Lunch break 12:00-13:00',
    'json',
    '{
        "workingDays": [1, 2, 3, 4, 5],
        "workingHours": {
            "start": "09:00",
            "end": "17:00"
        },
        "breaks": [
            {"start": "12:00", "end": "13:00", "name": "Lunch"}
        ],
        "holidays": []
    }'::jsonb,
    'hard',
    1,
    TRUE,
    'TIME-CONST-001',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- Resource Constraint: Capacity Limits
INSERT INTO plan_scheduler_profile_constraints (
    id, profile_id, constraint_type, name, description,
    source_type, source_json, enforcement, display_order,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES (
    'const-resource-001',
    'complex-mfg-2026-q1',
    'resource',
    'Resource Capacity Limits',
    'Enforce resource capacity constraints (machines, work centers, human operators)',
    'json',
    '{
        "respectCapacity": true,
        "allowOverallocation": false,
        "overallocationThreshold": 1.0
    }'::jsonb,
    'hard',
    2,
    TRUE,
    'RES-CONST-001',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- Dependency Constraint: Task Dependencies
INSERT INTO plan_scheduler_profile_constraints (
    id, profile_id, constraint_type, name, description,
    source_type, source_json, enforcement, display_order,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES (
    'const-dep-001',
    'complex-mfg-2026-q1',
    'dependency',
    'Strict Task Dependencies',
    'Enforce FS (Finish-to-Start) dependencies between operations',
    'json',
    '{
        "respectDependencies": true,
        "dependencyTypes": ["FS", "SS"],
        "allowParallel": true
    }'::jsonb,
    'hard',
    3,
    TRUE,
    'DEP-CONST-001',
    'system',
    NOW(),
    'system',
    NOW(),
    1
);

-- =====================================================================
-- 5. SETTINGS (AI Optimization Parameters)
-- =====================================================================

INSERT INTO plan_scheduler_profile_settings (
    id, profile_id, setting_key, setting_value, description,
    active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
) VALUES
    (
        'setting-opt-001',
        'complex-mfg-2026-q1',
        'optimization.objective',
        '"minimize_makespan"'::jsonb,
        'Primary optimization objective: minimize total project duration',
        TRUE,
        'OPT-OBJ',
        'system',
        NOW(),
        'system',
        NOW(),
        1
    ),
    (
        'setting-opt-002',
        'complex-mfg-2026-q1',
        'optimization.algorithm',
        '"genetic_algorithm"'::jsonb,
        'Optimization algorithm: genetic algorithm for complex scheduling',
        TRUE,
        'OPT-ALG',
        'system',
        NOW(),
        'system',
        NOW(),
        1
    ),
    (
        'setting-opt-003',
        'complex-mfg-2026-q1',
        'optimization.max_iterations',
        '1000'::jsonb,
        'Maximum iterations for AI optimization',
        TRUE,
        'OPT-ITER',
        'system',
        NOW(),
        'system',
        NOW(),
        1
    ),
    (
        'setting-opt-004',
        'complex-mfg-2026-q1',
        'optimization.population_size',
        '100'::jsonb,
        'Population size for genetic algorithm',
        TRUE,
        'OPT-POP',
        'system',
        NOW(),
        'system',
        NOW(),
        1
    ),
    (
        'setting-opt-005',
        'complex-mfg-2026-q1',
        'optimization.mutation_rate',
        '0.1'::jsonb,
        'Mutation rate for genetic algorithm (10%)',
        TRUE,
        'OPT-MUT',
        'system',
        NOW(),
        'system',
        NOW(),
        1
    ),
    (
        'setting-opt-006',
        'complex-mfg-2026-q1',
        'optimization.crossover_rate',
        '0.8'::jsonb,
        'Crossover rate for genetic algorithm (80%)',
        TRUE,
        'OPT-CROSS',
        'system',
        NOW(),
        'system',
        NOW(),
        1
    ),
    (
        'setting-disp-001',
        'complex-mfg-2026-q1',
        'display.default_view',
        '"gantt"'::jsonb,
        'Default view mode: Gantt chart',
        TRUE,
        'DISP-VIEW',
        'system',
        NOW(),
        'system',
        NOW(),
        1
    ),
    (
        'setting-disp-002',
        'complex-mfg-2026-q1',
        'display.default_zoom',
        '"day"'::jsonb,
        'Default zoom level: day view',
        TRUE,
        'DISP-ZOOM',
        'system',
        NOW(),
        'system',
        NOW(),
        1
    ),
    (
        'setting-disp-003',
        'complex-mfg-2026-q1',
        'display.color_scheme',
        '"by_resource_type"'::jsonb,
        'Color tasks by resource type',
        TRUE,
        'DISP-COLOR',
        'system',
        NOW(),
        'system',
        NOW(),
        1
    );

-- =====================================================================
-- 6. VERIFICATION QUERIES
-- =====================================================================

-- Verify profile created
SELECT id, name, description, is_default
FROM plan_scheduler_profiles
WHERE id = 'complex-mfg-2026-q1';

-- Verify data sources created
SELECT id, data_type, name,
       jsonb_array_length(source_json) as item_count
FROM plan_scheduler_profile_datasources
WHERE profile_id = 'complex-mfg-2026-q1'
ORDER BY display_order;

-- Verify constraints created
SELECT id, constraint_type, name, enforcement, display_order
FROM plan_scheduler_profile_constraints
WHERE profile_id = 'complex-mfg-2026-q1'
ORDER BY display_order;

-- Verify settings created
SELECT id, setting_key, setting_value, description
FROM plan_scheduler_profile_settings
WHERE profile_id = 'complex-mfg-2026-q1'
ORDER BY setting_key;

-- Count tasks by type
SELECT
    CASE
        WHEN task->>'type' = 'parent' THEN 'Work Orders'
        ELSE 'Operations'
    END as task_type,
    COUNT(*) as count
FROM plan_scheduler_profile_datasources,
     jsonb_array_elements(source_json) as task
WHERE profile_id = 'complex-mfg-2026-q1'
  AND data_type = 'tasks'
GROUP BY task_type;

-- Count resources by type
SELECT
    resource->>'type' as resource_type,
    COUNT(*) as count
FROM plan_scheduler_profile_datasources,
     jsonb_array_elements(source_json) as resource
WHERE profile_id = 'complex-mfg-2026-q1'
  AND data_type = 'resources'
GROUP BY resource_type;

-- =====================================================================
-- END OF SCRIPT
-- =====================================================================
