package services

// Manufacturing Process tools — AI-powered process analysis, predictive maintenance, and production planning.
//
// - analyze_process_issues      : identifies process parameter deviations, bottlenecks, and
//                                  optimization opportunities vs benchmarks.
// - predict_machine_maintenance : analyzes machine health signals, quality correlations, and
//                                  parameter trends to predict failures and create preventive orders.
// - adjust_production_plan      : re-schedules production orders based on current machine availability,
//                                  capacity constraints, and maintenance windows.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/llm"
)

func (s *AgentToolRegistryService) registerMfgProcessTools() {
	s.registerProcessIssuesTool()
	s.registerPredictiveMaintenanceTool()
	s.registerProductionPlanAdjustTool()
}

// ═══════════════════════════════════════════════════════════════════════════════
// PROCESS ISSUE ANALYSIS & OPTIMIZATION
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerProcessIssuesTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "analyze_process_issues",
			Description: "Analyzes manufacturing process parameters against benchmarks and historical data " +
				"to identify deviations, bottlenecks, and improvement opportunities. " +
				"Applies Six Sigma and lean manufacturing principles to recommend optimizations. " +
				"The agent can pre-populate parameters by querying the database, then pass results here. " +
				"Returns a prioritized optimization plan with expected impact for each improvement.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"process_name": {
						Type:        "string",
						Description: "Name of the manufacturing process (e.g. 'CNC Milling Station A', 'Injection Molding Line 3', 'Welding Cell B').",
					},
					"process_parameters_json": {
						Type: "string",
						Description: `Current process parameter values as JSON array. ` +
							`Example: [{"parameter":"feed_rate_mm_min","current":450,"optimal":500,"unit":"mm/min"},{"parameter":"temperature_C","current":218,"optimal":210,"unit":"°C"}]`,
					},
					"benchmark_json": {
						Type: "string",
						Description: `Industry or internal benchmark targets as JSON. ` +
							`Example: {"oee_target_pct":85,"cycle_time_target_sec":120,"scrap_rate_target_pct":0.5,"throughput_target_pph":240}`,
					},
					"actual_performance_json": {
						Type: "string",
						Description: `Actual performance KPIs as JSON. ` +
							`Example: {"oee_pct":72,"cycle_time_sec":145,"scrap_rate_pct":1.8,"throughput_pph":198,"downtime_min_last_shift":47}`,
					},
					"historical_trend_json": {
						Type: "string",
						Description: `Historical parameter or KPI trend data (last N shifts/days). ` +
							`Example: [{"date":"2026-03-01","oee_pct":75,"scrap_rate_pct":1.2},{"date":"2026-02-28","oee_pct":70,"scrap_rate_pct":2.1}]`,
					},
					"quality_correlation_json": {
						Type: "string",
						Description: `Known correlations between process parameters and quality outcomes. ` +
							`Example: [{"parameter":"temperature_C","quality_impact":"high","direction":"high temp → more warping defects"}]`,
					},
					"constraints_json": {
						Type: "string",
						Description: `Process constraints that cannot be changed. ` +
							`Example: {"max_feed_rate":600,"material_type":"ABS","operator_skill_level":"intermediate"}`,
					},
					"optimization_goals": {
						Type:        "string",
						Description: "Primary optimization objectives (comma-separated): 'reduce_scrap', 'increase_throughput', 'reduce_cycle_time', 'improve_oee', 'reduce_energy', 'reduce_cost' (default: all).",
					},
					"save_result": {
						Type:        "boolean",
						Description: "If true, saves the analysis result to mfg_process_analyses table (default: false).",
					},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		processName := stringArg(args, "process_name")
		optimGoals := stringArg(args, "optimization_goals")
		if optimGoals == "" {
			optimGoals = "reduce_scrap, increase_throughput, improve_oee, reduce_cost"
		}

		systemPrompt := `You are a manufacturing process engineer with 20+ years of experience in Six Sigma (Black Belt), lean manufacturing, and process optimization.
Analyze process data to identify root causes of inefficiency, quantify improvement opportunities, and provide a prioritized action plan.
Respond ONLY with valid JSON — no markdown, no additional text.`

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("## Manufacturing Process Analysis\nProcess: %s\nOptimization Goals: %s\n\n", processName, optimGoals))

		sections := map[string]string{
			"Current Process Parameters": "process_parameters_json",
			"Performance Benchmarks":     "benchmark_json",
			"Actual Performance KPIs":    "actual_performance_json",
			"Historical Trend":           "historical_trend_json",
			"Quality Correlations":       "quality_correlation_json",
			"Process Constraints":        "constraints_json",
		}
		for title, key := range sections {
			if v := stringArg(args, key); v != "" {
				sb.WriteString(fmt.Sprintf("### %s\n%s\n\n", title, v))
			}
		}

		sb.WriteString(`Respond with this JSON structure:
{
  "process_health_score": <0-100>,
  "oee_breakdown": {
    "availability_pct": <float>, "performance_pct": <float>, "quality_pct": <float>, "oee_pct": <float>,
    "losses": ["<top loss 1>", "<top loss 2>"]
  },
  "parameter_deviations": [
    {
      "parameter": "<name>",
      "current": <value>,
      "optimal": <value>,
      "deviation_pct": <float>,
      "impact": "high|medium|low",
      "affected_kpi": "<scrap_rate|throughput|cycle_time|quality>",
      "recommendation": "<specific adjustment>"
    }
  ],
  "bottlenecks": [
    {
      "location": "<station or step>",
      "type": "capacity|quality|downtime|changeover|material",
      "impact_description": "<how it constrains the process>",
      "estimated_throughput_loss_pct": <float>
    }
  ],
  "optimization_opportunities": [
    {
      "priority": 1,
      "opportunity": "<description>",
      "category": "parameter_adjustment|process_redesign|maintenance|training|material|equipment_upgrade",
      "estimated_improvement": "<e.g. +12% OEE, -0.8% scrap rate>",
      "implementation_effort": "low|medium|high",
      "payback_period": "<e.g. 2 weeks, 3 months>",
      "action_steps": ["<step 1>", "<step 2>"]
    }
  ],
  "benchmarking_gaps": {
    "oee_gap_pct": <float>,
    "scrap_gap_pct": <float>,
    "cycle_time_gap_sec": <float>,
    "summary": "<gap analysis summary>"
  },
  "immediate_actions": ["<action within today>"],
  "kpi_targets_30d": {
    "oee_pct": <float>,
    "scrap_rate_pct": <float>,
    "throughput_pph": <float>
  },
  "summary": "<executive summary>"
}`)

		messages := []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": sb.String()},
		}
		apiKey, model := s.openAIKey, "gpt-4o"
		resultText, err := llm.CallLLM(ctx, "process_analysis", apiKey, model, messages, 0.2)
		if err != nil {
			return "", fmt.Errorf("process analysis LLM call failed: %w", err)
		}
		resultText = cleanLLMJSON(resultText)

		var analysis map[string]interface{}
		json.Unmarshal([]byte(resultText), &analysis) //nolint:errcheck

		analysisID := uuid.New().String()
		if boolArg(args, "save_result") && s.sqlDB != nil {
			resultJSON, _ := json.Marshal(analysis)
			s.sqlDB.ExecContext(ctx, //nolint:errcheck
				`INSERT INTO mfg_process_analyses (id, process_name, result_json, createdon)
				 VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`,
				analysisID, processName, string(resultJSON), time.Now().UTC())
		}

		out, _ := json.Marshal(map[string]interface{}{
			"analysis_id":  analysisID,
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
			"process_name": processName,
			"analysis":     analysis,
		})
		return string(out), nil
	})
}

// ═══════════════════════════════════════════════════════════════════════════════
// PREDICTIVE MACHINE MAINTENANCE
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerPredictiveMaintenanceTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "predict_machine_maintenance",
			Description: "Analyzes machine health signals, operating parameters, quality trends, and maintenance history " +
				"to predict potential failures and recommend preventive maintenance actions. " +
				"Assigns a Machine Health Score (0–100) and estimates time-to-failure for each failure mode. " +
				"If create_order is true, automatically creates a preventive maintenance order in the mfg_maintenance_orders table. " +
				"Returns predicted failure modes with probability, urgency, and recommended maintenance tasks.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"machine_id": {
						Type:        "string",
						Description: "Machine or equipment ID from the system.",
					},
					"machine_name": {
						Type:        "string",
						Description: "Human-readable machine name (e.g. 'CNC Mill #3', 'Press Line A', 'Robot Arm Cell 2').",
					},
					"machine_type": {
						Type:        "string",
						Description: "Machine type for domain-specific analysis (e.g. 'CNC Machine', 'Injection Molder', 'Welding Robot', 'Conveyor', 'Hydraulic Press').",
					},
					"status_data_json": {
						Type: "string",
						Description: `Current machine status and sensor readings as JSON. ` +
							`Example: {"power_kw":18.5,"spindle_load_pct":73,"vibration_mm_s":4.2,"temperature_C":68,"coolant_pressure_bar":3.1,"cycle_count":48250,"uptime_hours_since_maintenance":1240}`,
					},
					"parameter_history_json": {
						Type: "string",
						Description: `Time-series of machine parameters (last N readings or shifts). ` +
							`Example: [{"timestamp":"2026-03-02T06:00:00Z","vibration_mm_s":3.1,"temperature_C":65,"spindle_load_pct":68}]`,
					},
					"quality_trend_json": {
						Type: "string",
						Description: `Quality metrics correlated to this machine (defect rates, dimensional deviations). ` +
							`Example: [{"date":"2026-03-01","defect_rate_pct":0.8,"dimension_cpk":1.2},{"date":"2026-03-02","defect_rate_pct":1.4,"dimension_cpk":0.95}]`,
					},
					"maintenance_history_json": {
						Type: "string",
						Description: `Recent maintenance log. Example: [{"date":"2026-01-15","type":"PM","work_done":"Replace spindle bearing, lubrication"},{"date":"2026-02-20","type":"CM","work_done":"Emergency coolant pump replacement"}]`,
					},
					"alarm_log_json": {
						Type: "string",
						Description: `Recent alarm/fault log from the machine controller. Example: [{"timestamp":"2026-03-01T14:22:00Z","code":"E103","description":"Overtemperature warning"},{"timestamp":"2026-03-02T07:45:00Z","code":"E103","description":"Overtemperature warning"}]`,
					},
					"operating_schedule_json": {
						Type: "string",
						Description: `Planned production schedule for this machine. Example: {"shifts_per_day":2,"production_hours_remaining_this_week":60,"next_scheduled_maintenance":"2026-03-15"}`,
					},
					"create_order": {
						Type:        "boolean",
						Description: "If true and failures are predicted, creates a preventive maintenance order in the mfg_maintenance_orders table (default: false).",
					},
					"urgency_threshold": {
						Type:        "string",
						Description: "Only create a maintenance order if predicted failure is at or above this urgency: 'critical|high|medium|low' (default: 'medium').",
					},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		machineID := stringArg(args, "machine_id")
		machineName := stringArg(args, "machine_name")
		machineType := stringArg(args, "machine_type")
		urgencyThreshold := stringArg(args, "urgency_threshold")
		if urgencyThreshold == "" {
			urgencyThreshold = "medium"
		}

		systemPrompt := `You are a predictive maintenance engineer with expertise in industrial equipment reliability, Failure Mode and Effects Analysis (FMEA), and condition-based monitoring.
Analyze machine health data to predict failure modes, assess risk, and recommend proactive maintenance interventions.
Respond ONLY with valid JSON — no markdown, no additional text.`

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("## Predictive Maintenance Analysis\nMachine: %s (ID: %s) | Type: %s\n\n", machineName, machineID, machineType))

		sections := map[string]string{
			"Current Machine Status & Sensor Data": "status_data_json",
			"Parameter History (Time-Series)":      "parameter_history_json",
			"Quality Trend Correlation":            "quality_trend_json",
			"Maintenance History":                  "maintenance_history_json",
			"Alarm / Fault Log":                    "alarm_log_json",
			"Operating Schedule":                   "operating_schedule_json",
		}
		for title, key := range sections {
			if v := stringArg(args, key); v != "" {
				sb.WriteString(fmt.Sprintf("### %s\n%s\n\n", title, v))
			}
		}

		sb.WriteString(`Respond with this JSON structure:
{
  "machine_health_score": <0-100>,
  "health_trend": "improving|stable|degrading|critical",
  "predicted_failures": [
    {
      "failure_mode": "<description>",
      "component": "<affected component>",
      "probability_pct": <0-100>,
      "estimated_time_to_failure": "<e.g. 3-5 days, 2-3 weeks, 1-2 months>",
      "urgency": "critical|high|medium|low",
      "evidence": ["<sensor reading>", "<alarm code>", "<trend observation>"],
      "quality_impact": "<how this failure affects product quality>",
      "production_impact": "<estimated downtime if failure occurs>"
    }
  ],
  "key_indicators": [
    {"indicator": "<parameter name>", "current_value": "<value>", "threshold": "<warning threshold>", "status": "normal|warning|critical", "trend": "increasing|stable|decreasing"}
  ],
  "recommended_maintenance": [
    {
      "task": "<maintenance task description>",
      "priority": "immediate|this-week|next-pm",
      "estimated_duration_hours": <float>,
      "required_parts": ["<part name>"],
      "required_skills": "<technician skill level>",
      "addresses_failure_mode": "<which predicted failure this prevents>"
    }
  ],
  "optimal_maintenance_window": "<suggested best time to perform maintenance, e.g. 'weekend shutdown', 'between shifts Thursday'>",
  "estimated_maintenance_cost": "<rough estimate>",
  "cost_of_unplanned_failure": "<estimate of downtime + repair cost if failure occurs>",
  "quality_risk_summary": "<how current machine health is affecting or will affect product quality>",
  "summary": "<executive summary for maintenance manager>"
}`)

		messages := []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": sb.String()},
		}
		apiKey, model := s.openAIKey, "gpt-4o"
		resultText, err := llm.CallLLM(ctx, "process_analysis", apiKey, model, messages, 0.2)
		if err != nil {
			return "", fmt.Errorf("predictive maintenance LLM call failed: %w", err)
		}
		resultText = cleanLLMJSON(resultText)

		var prediction map[string]interface{}
		json.Unmarshal([]byte(resultText), &prediction) //nolint:errcheck

		predictionID := uuid.New().String()
		var maintenanceOrderID string

		// Optionally create a preventive maintenance order
		if boolArg(args, "create_order") && s.sqlDB != nil && prediction != nil {
			maintenanceOrderID = maybeCreateMaintenanceOrder(ctx, s, predictionID, machineID, machineName, urgencyThreshold, prediction)
		}

		result := map[string]interface{}{
			"prediction_id": predictionID,
			"timestamp":     time.Now().UTC().Format(time.RFC3339),
			"machine_id":    machineID,
			"machine_name":  machineName,
			"prediction":    prediction,
		}
		if maintenanceOrderID != "" {
			result["maintenance_order_id"] = maintenanceOrderID
			result["maintenance_order_created"] = true
		}

		out, _ := json.Marshal(result)
		return string(out), nil
	})
}

// maybeCreateMaintenanceOrder creates a preventive maintenance order if urgency meets threshold.
func maybeCreateMaintenanceOrder(ctx context.Context, s *AgentToolRegistryService,
	predictionID, machineID, machineName, urgencyThreshold string, prediction map[string]interface{}) string {

	urgencyRank := map[string]int{"low": 1, "medium": 2, "high": 3, "critical": 4}
	thresholdRank := urgencyRank[urgencyThreshold]

	// Find highest urgency predicted failure
	maxUrgency := ""
	if failures, ok := prediction["predicted_failures"].([]interface{}); ok {
		for _, f := range failures {
			if fm, ok := f.(map[string]interface{}); ok {
				u, _ := fm["urgency"].(string)
				if urgencyRank[u] > urgencyRank[maxUrgency] {
					maxUrgency = u
				}
			}
		}
	}

	if urgencyRank[maxUrgency] < thresholdRank {
		return "" // urgency below threshold, don't create order
	}

	orderID := uuid.New().String()
	summary, _ := prediction["summary"].(string)
	healthScore := 0
	if hs, ok := prediction["machine_health_score"].(float64); ok {
		healthScore = int(hs)
	}
	predJSON, _ := json.Marshal(prediction)

	_, err := s.sqlDB.ExecContext(ctx,
		`INSERT INTO mfg_maintenance_orders
		 (id, machine_id, machine_name, order_type, urgency, status,
		  health_score_at_creation, prediction_id, summary, prediction_json,
		  created_by, createdon)
		 VALUES ($1,$2,$3,'preventive',$4,'open',$5,$6,$7,$8,'agent',$9)
		 ON CONFLICT DO NOTHING`,
		orderID, machineID, machineName, maxUrgency, healthScore,
		predictionID, summary, string(predJSON), time.Now().UTC())
	if err != nil {
		return ""
	}
	return orderID
}

// ═══════════════════════════════════════════════════════════════════════════════
// PRODUCTION PLAN ADJUSTMENT
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerProductionPlanAdjustTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "adjust_production_plan",
			Description: "Re-optimizes production orders based on current machine availability, " +
				"maintenance windows, capacity constraints, and order priorities. " +
				"Accounts for machines that are down for maintenance or running at reduced capacity. " +
				"Returns a recommended adjusted schedule with reasoning for each change, " +
				"impact on delivery dates, and capacity utilization metrics. " +
				"The agent can pass pre-queried production order and machine data directly.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"production_orders_json": {
						Type: "string",
						Description: `Current production orders as JSON array. ` +
							`Example: [{"order_id":"PO-001","product":"Part-A","quantity":500,"due_date":"2026-03-05","priority":"high","required_machine":"CNC Mill #3","estimated_cycle_time_hrs":8},{"order_id":"PO-002","product":"Part-B","quantity":200,"due_date":"2026-03-07","priority":"medium","required_machine":"CNC Mill #3","estimated_cycle_time_hrs":4}]`,
					},
					"machine_status_json": {
						Type: "string",
						Description: `Current status of all machines as JSON array. ` +
							`Example: [{"machine_id":"CNC-03","name":"CNC Mill #3","status":"maintenance_required","available_from":"2026-03-04T06:00:00Z","capacity_pct":0},{"machine_id":"CNC-01","name":"CNC Mill #1","status":"running","capacity_pct":100}]`,
					},
					"maintenance_windows_json": {
						Type: "string",
						Description: `Planned maintenance windows affecting machine availability. ` +
							`Example: [{"machine_id":"CNC-03","start":"2026-03-03T08:00:00Z","end":"2026-03-04T06:00:00Z","type":"preventive"}]`,
					},
					"alternative_machines_json": {
						Type: "string",
						Description: `Alternative machines that can run specific products (for routing flexibility). ` +
							`Example: [{"product":"Part-A","alternative_machines":["CNC Mill #1","CNC Mill #2"],"efficiency_pct":90}]`,
					},
					"shift_calendar_json": {
						Type: "string",
						Description: `Shift schedule and capacity hours per machine. ` +
							`Example: {"shifts_per_day":2,"hours_per_shift":8,"working_days_this_week":["2026-03-03","2026-03-04","2026-03-05","2026-03-06"]}`,
					},
					"constraints_json": {
						Type: "string",
						Description: `Planning constraints. ` +
							`Example: {"max_overtime_hours_per_machine":4,"minimum_changeover_minutes":30,"priority_rules":"customer_critical_first","frozen_zone_hours":24}`,
					},
					"optimization_objective": {
						Type:        "string",
						Description: "Primary scheduling objective: 'on_time_delivery' (default) | 'maximize_throughput' | 'minimize_makespan' | 'balance_load'.",
					},
					"apply_changes": {
						Type:        "boolean",
						Description: "If true and production_orders table exists, applies the recommended schedule changes to the database (default: false — returns recommendations only).",
					},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		objective := stringArg(args, "optimization_objective")
		if objective == "" {
			objective = "on_time_delivery"
		}

		systemPrompt := fmt.Sprintf(`You are a manufacturing production planner with expertise in constraint-based scheduling, capacity planning, and manufacturing execution systems (MES).
Your objective is to optimize the production schedule with the primary goal of: %s.
Minimize disruption to on-time delivery while accounting for machine availability and capacity constraints.
Respond ONLY with valid JSON — no markdown, no additional text.`, objective)

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("## Production Plan Adjustment Request\nObjective: %s\n\n", objective))

		sections := map[string]string{
			"Current Production Orders":       "production_orders_json",
			"Machine Status":                  "machine_status_json",
			"Maintenance Windows":             "maintenance_windows_json",
			"Alternative Machine Routings":    "alternative_machines_json",
			"Shift Calendar":                  "shift_calendar_json",
			"Planning Constraints":            "constraints_json",
		}
		for title, key := range sections {
			if v := stringArg(args, key); v != "" {
				sb.WriteString(fmt.Sprintf("### %s\n%s\n\n", title, v))
			}
		}

		sb.WriteString(`Respond with this JSON structure:
{
  "schedule_feasible": <bool>,
  "capacity_summary": {
    "total_required_hours": <float>,
    "total_available_hours": <float>,
    "utilization_pct": <float>,
    "overloaded_machines": ["<machine name>"],
    "underutilized_machines": ["<machine name>"]
  },
  "adjusted_orders": [
    {
      "order_id": "<id>",
      "product": "<product>",
      "quantity": <int>,
      "original_machine": "<machine>",
      "assigned_machine": "<machine>",
      "original_start": "<datetime or null>",
      "recommended_start": "<datetime>",
      "recommended_end": "<datetime>",
      "original_due_date": "<date>",
      "on_time": <bool>,
      "delay_days": <int or 0>,
      "priority": "<critical|high|medium|low>",
      "change_reason": "<why this order was moved or kept>",
      "action": "no_change|rescheduled|rerouted|split|expedite"
    }
  ],
  "at_risk_orders": [
    {
      "order_id": "<id>",
      "risk": "<description of delivery risk>",
      "mitigation_options": ["<option 1>", "<option 2>"]
    }
  ],
  "recommendations": [
    {
      "type": "overtime|subcontract|expedite_material|split_batch|reprioritize",
      "description": "<specific recommendation>",
      "affected_orders": ["<order_id>"],
      "expected_outcome": "<result if implemented>"
    }
  ],
  "kpi_impact": {
    "orders_on_time": <int>,
    "orders_delayed": <int>,
    "avg_delay_days": <float>,
    "capacity_utilization_pct": <float>,
    "vs_current_plan": "<summary of improvement or trade-offs>"
  },
  "summary": "<executive summary for production manager>"
}`)

		messages := []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": sb.String()},
		}
		apiKey, model := s.openAIKey, "gpt-4o"
		resultText, err := llm.CallLLM(ctx, "process_analysis", apiKey, model, messages, 0.2)
		if err != nil {
			return "", fmt.Errorf("production planning LLM call failed: %w", err)
		}
		resultText = cleanLLMJSON(resultText)

		var planResult map[string]interface{}
		json.Unmarshal([]byte(resultText), &planResult) //nolint:errcheck

		planID := uuid.New().String()
		appliedChanges := 0

		// Optionally apply changes to the DB
		if boolArg(args, "apply_changes") && s.sqlDB != nil && planResult != nil {
			appliedChanges = applyProductionPlanChanges(ctx, s, planResult)
		}

		result := map[string]interface{}{
			"plan_id":        planID,
			"timestamp":      time.Now().UTC().Format(time.RFC3339),
			"objective":      objective,
			"adjusted_plan":  planResult,
		}
		if boolArg(args, "apply_changes") {
			result["changes_applied"] = appliedChanges
		}

		out, _ := json.Marshal(result)
		return string(out), nil
	})
}

// applyProductionPlanChanges attempts to update production_orders table with recommended schedule.
// Returns the number of orders updated.
func applyProductionPlanChanges(ctx context.Context, s *AgentToolRegistryService, plan map[string]interface{}) int {
	orders, ok := plan["adjusted_orders"].([]interface{})
	if !ok {
		return 0
	}
	updated := 0
	for _, o := range orders {
		order, ok := o.(map[string]interface{})
		if !ok {
			continue
		}
		action, _ := order["action"].(string)
		if action == "no_change" {
			continue
		}
		orderID, _ := order["order_id"].(string)
		assignedMachine, _ := order["assigned_machine"].(string)
		recStart, _ := order["recommended_start"].(string)
		recEnd, _ := order["recommended_end"].(string)
		if orderID == "" {
			continue
		}

		// Try to update production_orders table (ignore if table doesn't exist)
		_, err := s.sqlDB.ExecContext(ctx,
			`UPDATE production_orders
			 SET assigned_machine = $1, planned_start = $2, planned_end = $3,
			     modifiedby = 'agent', modifiedon = $4
			 WHERE id = $5 OR order_number = $5`,
			assignedMachine, recStart, recEnd, time.Now().UTC(), orderID)
		if err == nil {
			updated++
		}
	}
	return updated
}
