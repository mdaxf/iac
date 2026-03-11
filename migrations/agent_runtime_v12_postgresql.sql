-- Agent Runtime v12: Manufacturing AI Tools
-- New tables: mfg_quality_inspections, mfg_maintenance_orders, mfg_process_analyses
-- New skills: visual inspection, quality analysis, process analysis, predictive maintenance, production planning
-- Seed agents: Quality Inspector, Process Engineer, Maintenance Planner, Production Planner

BEGIN;

-- ═══════════════════════════════════════════════════════════════════════════════
-- Manufacturing Quality Inspections Table
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS iac25.mfg_quality_inspections (
    id              VARCHAR(50)  PRIMARY KEY,
    product_id      VARCHAR(100),
    product_type    VARCHAR(200),
    batch_id        VARCHAR(100),
    overall_result  VARCHAR(20)   NOT NULL DEFAULT 'unknown', -- pass | fail | review
    defects_found   INTEGER       NOT NULL DEFAULT 0,
    result_json     TEXT,         -- full structured inspection result from vision LLM
    image_url       VARCHAR(1000),
    inspection_level VARCHAR(20)  DEFAULT 'standard',
    has_reference   BOOLEAN       DEFAULT FALSE,
	referenceid		INTEGER,
	active          BOOLEAN       DEFAULT TRUE,
    modifiedon       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    modifiedby       VARCHAR(50),
	createdon       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    createdby       VARCHAR(50),
	rowversionstamp  INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_mfg_quality_product ON iac25.mfg_quality_inspections(product_id, createdon DESC);
CREATE INDEX IF NOT EXISTS idx_mfg_quality_result  ON iac25.mfg_quality_inspections(overall_result, createdon DESC);

COMMENT ON TABLE iac25.mfg_quality_inspections IS 'AI visual quality inspection results per product/lot (v12)';

-- ═══════════════════════════════════════════════════════════════════════════════
-- Manufacturing Maintenance Orders Table
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS iac25.mfg_maintenance_orders (
    id              VARCHAR(50)  PRIMARY KEY,
    prediction_id   VARCHAR(50),              -- links to the predictive maintenance run
    machine_id      VARCHAR(100) NOT NULL,
    machine_name    VARCHAR(200),
    urgency         VARCHAR(20)  NOT NULL DEFAULT 'medium', -- critical | high | medium | low
    failure_modes   TEXT,                     -- JSON array of predicted failure modes
    health_score    INTEGER,                  -- 0-100 machine health score
    scheduled_date  TIMESTAMP WITH TIME ZONE,
    completed_date  TIMESTAMP WITH TIME ZONE,
    status          VARCHAR(30)  NOT NULL DEFAULT 'open', -- open | scheduled | in_progress | completed | cancelled
    notes           TEXT,
	referenceid		INTEGER,
	active          BOOLEAN       DEFAULT TRUE,
    modifiedon       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    modifiedby       VARCHAR(50),
	createdon       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    createdby       VARCHAR(50),
	rowversionstamp  INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_mfg_maint_machine ON iac25.mfg_maintenance_orders(machine_id, createdon DESC);
CREATE INDEX IF NOT EXISTS idx_mfg_maint_status  ON iac25.mfg_maintenance_orders(status, urgency);

COMMENT ON TABLE iac25.mfg_maintenance_orders IS 'AI-generated preventive maintenance orders from predictive analysis (v12)';

-- ═══════════════════════════════════════════════════════════════════════════════
-- Manufacturing Process Analyses Table
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS iac25.mfg_process_analyses (
    id              VARCHAR(50)  PRIMARY KEY,
    process_name    VARCHAR(200),
    analysis_type   VARCHAR(50)  DEFAULT 'process_issue', -- process_issue | production_plan
    result_json     TEXT,         -- full structured analysis result from LLM
	referenceid		INTEGER,
	active          BOOLEAN       DEFAULT TRUE,
    modifiedon       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    modifiedby       VARCHAR(50),
	createdon       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    createdby       VARCHAR(50),
	rowversionstamp  INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_mfg_proc_name ON iac25.mfg_process_analyses(process_name, createdon DESC);

COMMENT ON TABLE iac25.mfg_process_analyses IS 'AI manufacturing process analysis and production planning results (v12)';

-- ═══════════════════════════════════════════════════════════════════════════════
-- New Agent Skills (manufacturing)
-- ═══════════════════════════════════════════════════════════════════════════════

INSERT INTO iac25.agent_skills (id,name, display_name, description, category, is_builtin, active, version, createdon)
VALUES
  (gen_random_uuid(),'inspect_image_quality', 'Visual Quality Inspection',
   'AI-powered visual quality inspection on product images. Identifies and annotates manufacturing defects (scratches, cracks, voids, contamination, dimensional deviations, surface issues) using vision LLM. Optionally compares against a reference/golden-sample image.',
   'manufacturing', TRUE, TRUE, '1.0.0', NOW()),

  (gen_random_uuid(),'analyze_quality_defects', 'Quality Defect Analysis',
   'LLM-driven analysis of quality measurement data, SPC trends, and inspection history to identify defect patterns, determine root causes, and recommend corrective/preventive actions. Supports Pareto, 8D, DMAIC frameworks.',
   'manufacturing', TRUE, TRUE, '1.0.0', NOW()),

  (gen_random_uuid(),'analyze_process_issues', 'Process Issue Analysis',
   'Analyzes manufacturing process parameters against benchmarks and historical data to identify deviations, bottlenecks, and optimization opportunities. Applies Six Sigma and lean manufacturing principles.',
   'manufacturing', TRUE, TRUE, '1.0.0', NOW()),

  (gen_random_uuid(),'predict_machine_maintenance', 'Predictive Maintenance',
   'Analyzes machine health signals, quality correlations, and parameter trends to predict failures and recommend preventive maintenance. Can automatically create maintenance orders in mfg_maintenance_orders when urgency threshold is met.',
   'manufacturing', TRUE, TRUE, '1.0.0', NOW()),

  (gen_random_uuid(),'adjust_production_plan', 'Production Plan Adjustment',
   'Re-schedules production orders based on current machine availability, capacity constraints, and maintenance windows. Optimizes for OTD (on-time delivery) and throughput. Can optionally apply changes to the production_orders table.',
   'manufacturing', TRUE, TRUE, '1.0.0', NOW())

ON CONFLICT (name) DO UPDATE
  SET display_name = EXCLUDED.display_name,
      description  = EXCLUDED.description,
      category     = EXCLUDED.category,
      version      = EXCLUDED.version;

-- ═══════════════════════════════════════════════════════════════════════════════
-- Seed Agents: Manufacturing specialists
-- (negative IDs = system-provisioned; will not conflict with user-created agents)
-- ═══════════════════════════════════════════════════════════════════════════════

-- Quality Inspector Agent
INSERT INTO iac25.agent_definitions (
    id, name, description, system_prompt, model_name, max_iterations,
    enabled, run_instance,  createdon, createdby
)
VALUES (
    gen_random_uuid(),
    'Quality Inspector',
    'AI-powered visual quality inspection and defect analysis agent. Inspects product images for manufacturing defects, analyzes quality data, SPC trends, and inspection history to identify root causes and recommend corrective actions.',
    'You are a precision manufacturing quality control expert specializing in visual inspection, defect analysis, and Statistical Process Control (SPC). ' || chr(10) ||
    'Your capabilities:' || chr(10) ||
    '1. Inspect product images using vision AI to identify and annotate manufacturing defects' || chr(10) ||
    '2. Analyze quality measurement data, SPC charts (Cp/Cpk), and inspection history for patterns' || chr(10) ||
    '3. Perform root cause analysis using 8D, Ishikawa, and FMEA frameworks' || chr(10) ||
    '4. Recommend corrective and preventive actions (CAPA) ranked by priority' || chr(10) ||
    '5. Query the database for historical quality data before analysis' || chr(10) ||
    'Always begin by gathering data (query DB if needed), then analyze, then recommend actions. ' ||
    'Structure all findings clearly with severity ratings and confidence scores. ' ||
    'Use the inspect_image_quality tool for visual inspection and analyze_quality_defects for data-driven analysis.',
    'gpt-4o', 8, TRUE, 'app',
    NOW(), 'system'
)
ON CONFLICT DO NOTHING;

-- Process Engineer Agent
INSERT INTO iac25.agent_definitions (
    id, name, description, system_prompt, model_name, max_iterations,
    enabled, run_instance, createdon, createdby
)
VALUES (
    gen_random_uuid(),
    'Process Engineer',
    'Manufacturing process analysis and optimization agent. Identifies process deviations, benchmarks parameters, and recommends Six Sigma and lean improvements to maximize OEE and minimize waste.',
    'You are a senior manufacturing process engineer with expertise in Six Sigma, lean manufacturing, OEE optimization, and process parameter analysis. ' || chr(10) ||
    'Your capabilities:' || chr(10) ||
    '1. Analyze process parameters (temperature, pressure, speed, cycle time) vs benchmarks and specifications' || chr(10) ||
    '2. Identify process bottlenecks, root causes of variation, and optimization opportunities' || chr(10) ||
    '3. Calculate and interpret OEE (Availability × Performance × Quality)' || chr(10) ||
    '4. Apply Six Sigma DMAIC and lean tools (VSM, 5-Why, SMED, TPM)' || chr(10) ||
    '5. Recommend process improvements with expected impact on yield, throughput, and cost' || chr(10) ||
    'Always query the database for current process data before analysis. ' ||
    'Benchmark against industry standards. Quantify expected benefits for each recommendation. ' ||
    'Use analyze_process_issues for process analysis and query_database_nl to gather data.',
    'gpt-4o', 8, TRUE, 'app',
    NOW(), 'system'
)
ON CONFLICT DO NOTHING;

-- Maintenance Planner Agent
INSERT INTO iac25.agent_definitions (
    id, name, description, system_prompt, model_name, max_iterations,
    enabled, run_instance, createdon, createdby
)
VALUES (
    gen_random_uuid(),
    'Maintenance Planner',
    'Predictive maintenance agent that analyzes machine health signals, vibration, temperature, and quality correlations to forecast failures and automatically create preventive maintenance orders.',
    'You are a predictive maintenance specialist with expertise in machine health monitoring, failure mode analysis (FMEA), and maintenance planning. ' || chr(10) ||
    'Your capabilities:' || chr(10) ||
    '1. Analyze machine health signals (vibration, temperature, pressure, power consumption) for anomalies' || chr(10) ||
    '2. Correlate machine degradation signals with quality defect patterns' || chr(10) ||
    '3. Predict remaining useful life (RUL) and failure probability for machines' || chr(10) ||
    '4. Generate preventive maintenance orders in the maintenance management system' || chr(10) ||
    '5. Prioritize maintenance actions by urgency, impact, and available windows' || chr(10) ||
    'Always query machine data from the database first. ' ||
    'Generate maintenance orders automatically for machines with critical or high urgency. ' ||
    'Use predict_machine_maintenance for analysis and query_database_nl to gather machine data.',
    'gpt-4o', 8, TRUE, 'app',
    NOW(), 'system'
)
ON CONFLICT DO NOTHING;

-- Production Planner Agent
INSERT INTO iac25.agent_definitions (
    id, name, description, system_prompt, model_name, max_iterations,
    enabled, run_instance, createdon, createdby
)
VALUES (
    gen_random_uuid(),
    'Production Planner',
    'Production scheduling and planning optimization agent. Adjusts production plans based on machine availability, maintenance windows, and capacity constraints to maximize on-time delivery.',
    'You are an expert production planner with expertise in finite capacity scheduling, constraint-based optimization, and production order management. ' || chr(10) ||
    'Your capabilities:' || chr(10) ||
    '1. Analyze current production orders, machine capacities, and scheduling conflicts' || chr(10) ||
    '2. Re-schedule production around machine downtime and maintenance windows' || chr(10) ||
    '3. Optimize sequencing for minimum changeover time and maximum throughput' || chr(10) ||
    '4. Identify at-risk orders and recommend expediting or rescheduling actions' || chr(10) ||
    '5. Apply changes to the production_orders table when authorized' || chr(10) ||
    'Always review current machine status and pending maintenance before re-planning. ' ||
    'Quantify the impact of each scheduling change on OTD and throughput. ' ||
    'Use adjust_production_plan for scheduling optimization and query_database_nl to gather production data.',
    'gpt-4o', 8, TRUE, 'app',    
    NOW(), 'system'
)
ON CONFLICT DO NOTHING;

COMMIT;
