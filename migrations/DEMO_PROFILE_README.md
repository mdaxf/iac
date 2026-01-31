# Demo Plan Scheduler Profile

## Overview

This demo profile provides a complete manufacturing schedule example for testing and demonstration purposes.

**Profile ID**: `demo-profile-001`
**Profile Name**: Manufacturing Demo - 2026 Q1
**Start Date**: January 1, 2026
**Status**: All tasks are `todo` (new/unscheduled)

## Data Summary

### Work Orders & Tasks
- **5 Work Orders**: WO-101 through WO-105
- **30 Tasks Total**:
  - 5 parent tasks (one per work order)
  - 25 operation tasks (5 operations per work order)
- **Task Types**:
  - Op 01: Processing (Work Center)
  - Op 02: Fabrication (Machine)
  - Op 03: Manual Assembly (Human)
  - Op 04: Fabrication (Machine)
  - Op 05: Manual Assembly (Human)

### Resources (25 Total)
- **5 Work Centers**: WC-1 to WC-5 (Assembly role, standard hours)
- **10 Machines**: M-1 to M-10 (Fabrication role, 24/7 availability)
- **10 Operators**: Person 1 to Person 10 (Operator role, standard hours)

### Schedule Timeline
- **WO-101**: Jan 1-2, 2026
- **WO-102**: Jan 2-3, 2026
- **WO-103**: Jan 3-4, 2026 (spans weekend)
- **WO-104**: Jan 6-7, 2026
- **WO-105**: Jan 7-8, 2026

## Profile Components

### 1. Data Sources (2)
- **Tasks Data Source**: JSON constant with all 30 tasks
- **Resources Data Source**: JSON constant with all 25 resources

### 2. Constraints (3)
1. **Time Constraint**: Standard working hours
   - Monday-Friday, 9:00-17:00
   - Lunch break: 12:00-13:00

2. **Resource Constraint**: Capacity limits
   - Max 3 concurrent tasks per resource
   - No overallocation allowed

3. **Dependency Constraint**: Task dependencies
   - Strict finish-to-start dependencies
   - No negative lag allowed

### 3. Settings (5)
- `optimization_level`: "balanced"
- `max_iterations`: 1000
- `minimize_makespan`: true
- `resource_leveling`: true
- `time_buffer_percent`: 10

## Installation

### For PostgreSQL

```bash
psql -U your_username -d your_database -f demo_plan_profile_postgresql.sql
```

Or in psql:
```sql
\i /path/to/demo_plan_profile_postgresql.sql
```

### For MySQL

```bash
mysql -u your_username -p your_database < demo_plan_profile_mysql.sql
```

Or in mysql:
```sql
source /path/to/demo_plan_profile_mysql.sql;
```

## Verification

After running the script, verify the installation:

```sql
-- Count all components
SELECT
    'Profile' as entity,
    COUNT(*) as count
FROM plan_scheduler_profiles
WHERE id = 'demo-profile-001'

UNION ALL

SELECT 'Data Sources', COUNT(*)
FROM plan_scheduler_profile_data_sources
WHERE profile_id = 'demo-profile-001'

UNION ALL

SELECT 'Constraints', COUNT(*)
FROM plan_scheduler_profile_constraints
WHERE profile_id = 'demo-profile-001'

UNION ALL

SELECT 'Settings', COUNT(*)
FROM plan_scheduler_profile_settings
WHERE profile_id = 'demo-profile-001';
```

Expected output:
```
entity          | count
----------------|------
Profile         | 1
Data Sources    | 2
Constraints     | 3
Settings        | 5
```

## Usage in Application

### 1. Via Profile Management UI

1. Navigate to `/plan-profiles`
2. Click on "Manufacturing Demo - 2026 Q1"
3. Click "Use" to load in scheduler
4. Or click "Edit" to view/modify profile configuration

### 2. Direct URL Access

```
http://localhost:3000/plan-scheduler?profileId=demo-profile-001
```

### 3. API Endpoint

```bash
curl http://localhost:8080/api/plan-scheduler/profiles/demo-profile-001/data
```

## Data Structure

### Task JSON Format
```json
{
  "id": "t-1-1",
  "parentId": "P-1",
  "jobId": "WO-101",
  "name": "Op 01: Processing",
  "resourceIds": ["wc-1"],
  "start": "2026-01-01T08:00:00Z",
  "end": "2026-01-01T15:00:00Z",
  "progress": 0,
  "status": "todo",
  "allocation": 100,
  "setupMinutes": 30,
  "description": "Step 1 assigned to Work Center 1",
  "dependencies": [
    {"predecessorId": "previous-task-id", "type": "FS", "lag": 0}
  ]
}
```

### Resource JSON Format
```json
{
  "id": "wc-1",
  "name": "Work Center 1",
  "role": "Assembly",
  "type": "work_center",
  "color": "#2563eb",
  "availability": "standard"
}
```

## Features Demonstrated

### 1. Task Hierarchies
- Parent tasks (work orders) contain child tasks (operations)
- Parent bounds automatically calculated from children

### 2. Dependencies
- Sequential operations with finish-to-start dependencies
- Each operation waits for previous to complete

### 3. Resource Types
- **Work Centers**: Standard hours (9-5)
- **Machines**: 24/7 availability
- **Humans**: Standard hours (9-5)

### 4. Scheduling Patterns
- Setup times on first operations (30 minutes)
- Varied durations (4-7 hours per operation)
- Weekend awareness (Saturday/Sunday skipped)

### 5. Optimization Opportunities
- Resource leveling across timeline
- Makespan minimization
- Time buffer application
- Dependency enforcement

## Cleanup

To remove the demo profile:

```sql
-- Delete in correct order due to foreign keys
DELETE FROM plan_scheduler_profile_settings
WHERE profile_id = 'demo-profile-001';

DELETE FROM plan_scheduler_profile_constraints
WHERE profile_id = 'demo-profile-001';

DELETE FROM plan_scheduler_profile_data_sources
WHERE profile_id = 'demo-profile-001';

DELETE FROM plan_scheduler_profiles
WHERE id = 'demo-profile-001';
```

## Customization

### Change Start Date

Edit the JSON data sources and update all date strings:
```sql
UPDATE plan_scheduler_profile_data_sources
SET json_data = REPLACE(json_data, '2026-01', '2027-01')
WHERE profile_id = 'demo-profile-001';
```

### Add More Work Orders

1. Copy the pattern from WO-105
2. Increment IDs (WO-106, P-6, t-6-1, etc.)
3. Adjust start dates to continue timeline
4. Update the JSON data source

### Modify Working Hours

Update the time constraint configuration:
```sql
UPDATE plan_scheduler_profile_constraints
SET configuration = '{"workStartHour":8,"workEndHour":18,"breakStartHour":12,"breakEndHour":13,"workDays":[1,2,3,4,5,6],"nonWorkColor":"rgba(241,245,249,0.6)"}'
WHERE id = 'const-demo-time-001';
```

## Notes

- All dates are in ISO 8601 format (UTC)
- Task durations include non-working time for machines (24/7)
- Human and work center tasks respect working hours when scheduled
- Dependencies ensure operation sequence is maintained
- Profile is NOT set as default (can be changed in UI)

## Source

Generated from: `iac-portal/src/features/plan-scheduler/services/shortmockData.ts`

Last updated: 2025-12-11
