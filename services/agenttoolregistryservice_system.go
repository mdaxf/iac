package services

// System management tools — allow agents to inspect and configure IAC system resources:
// - System Configuration (ConfigStore — versioned MongoDB-backed config)
// - Integration Hubs (list, get, create, update status/enabled)
// - Jobs (list scheduled, create, toggle, list queue, statistics, enqueue)

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	iacmodels "github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services/configstore"
)

func (s *AgentToolRegistryService) registerSystemTools() {
	s.registerConfigTools()
	s.registerIntegrationHubTools()
	s.registerJobTools()
}

// ═══════════════════════════════════════════════════════════════════════════════
// SYSTEM CONFIGURATION — versioned ConfigStore (MongoDB)
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerConfigTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "list_configs",
			Description: "Lists system configurations in IAC ConfigStore. " +
				"Configurations are versioned, environment-aware settings objects. " +
				"Optionally filter by name or environment.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":        {Type: "string", Description: "Filter by configuration name (optional)"},
					"environment": {Type: "string", Description: "Filter by environment: production | staging | development (optional)"},
					"status":      {Type: "string", Description: "Filter by status: draft | active | archived (optional, default: active)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		svc := configstore.GetGlobalConfigStoreService()
		if svc == nil {
			return "", fmt.Errorf("ConfigStore service is not available")
		}
		name := stringArg(args, "name")
		env := stringArg(args, "environment")
		statusStr := stringArg(args, "status")
		if statusStr == "" {
			statusStr = "active"
		}
		var status iacmodels.ConfigurationStatus
		switch statusStr {
		case "draft":
			status = iacmodels.ConfigStatusDraft
		case "archived":
			status = iacmodels.ConfigStatusArchived
		default:
			status = iacmodels.ConfigStatusActive
		}

		configs, err := svc.ListConfigurations(ctx, name, env, status)
		if err != nil {
			return "", fmt.Errorf("failed to list configurations: %w", err)
		}

		type item struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Version     string `json:"version"`
			Environment string `json:"environment"`
			Status      string `json:"status"`
			UpdatedAt   string `json:"updated_at,omitempty"`
		}
		var items []item
		for _, c := range configs {
			items = append(items, item{
				ID:          c.ID,
				Name:        c.Name,
				Version:     c.Version,
				Environment: c.Environment,
				Status:      string(c.Status),
				UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
			})
		}
		if items == nil {
			items = []item{}
		}
		out, _ := json.Marshal(items)
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "get_config",
			Description: "Retrieves the active configuration by name and environment. " +
				"Returns the full configuration data object. " +
				"Use list_configs first to find available configuration names.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":        {Type: "string", Description: "Configuration name (required)"},
					"environment": {Type: "string", Description: "Environment: production | staging | development (required)"},
				},
				Required: []string{"name", "environment"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		svc := configstore.GetGlobalConfigStoreService()
		if svc == nil {
			return "", fmt.Errorf("ConfigStore service is not available")
		}
		name := stringArg(args, "name")
		env := stringArg(args, "environment")
		if name == "" || env == "" {
			return "", fmt.Errorf("name and environment are required")
		}
		cfg, err := svc.GetActiveConfiguration(ctx, name, env)
		if err != nil {
			return "", fmt.Errorf("configuration not found: %w", err)
		}
		out, _ := json.Marshal(map[string]interface{}{
			"id":          cfg.ID,
			"name":        cfg.Name,
			"version":     cfg.Version,
			"environment": cfg.Environment,
			"status":      string(cfg.Status),
			"data":        cfg.Data,
			"updated_at":  cfg.UpdatedAt.Format(time.RFC3339),
		})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "update_config",
			Description: "Updates the data fields of a draft system configuration. " +
				"Only configurations in 'draft' status can be updated — use get_config to retrieve the ID first. " +
				"Provide the new data as a JSON object. Existing keys not in data_json are preserved.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":      {Type: "string", Description: "Configuration ID (required, must be a draft)"},
					"data_json": {
						Type:        "string",
						Description: `JSON object with the new configuration data. Example: {"smtp_host":"mail.example.com","smtp_port":587,"max_agents":10}`,
					},
					"reason": {Type: "string", Description: "Reason for the change (for audit trail)"},
				},
				Required: []string{"id", "data_json"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		svc := configstore.GetGlobalConfigStoreService()
		if svc == nil {
			return "", fmt.Errorf("ConfigStore service is not available")
		}
		id := stringArg(args, "id")
		dataJSON := stringArg(args, "data_json")
		if id == "" || dataJSON == "" {
			return "", fmt.Errorf("id and data_json are required")
		}
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
			return "", fmt.Errorf("invalid data_json: %w", err)
		}
		reason := stringArg(args, "reason")
		if reason == "" {
			reason = "Updated by agent"
		}
		cfg, err := svc.UpdateConfiguration(ctx, id, data, "agent", reason)
		if err != nil {
			return "", fmt.Errorf("failed to update configuration: %w", err)
		}
		out, _ := json.Marshal(map[string]string{
			"id":      cfg.ID,
			"name":    cfg.Name,
			"version": cfg.Version,
			"status":  string(cfg.Status),
		})
		return string(out), nil
	})
}

// ═══════════════════════════════════════════════════════════════════════════════
// INTEGRATION HUB — MongoDB Integration_Hub collection
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerIntegrationHubTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "list_integration_hubs",
			Description: "Lists Integration Hub configurations in IAC. " +
				"Shows name, version, status, enabled state, and supported protocols.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"enabled_only": {Type: "boolean", Description: "If true, return only enabled hubs (default: false)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		svc := NewIntegrationHubService(s.docDB)
		var hubs []iacmodels.IntegrationHub
		var err error
		if boolArg(args, "enabled_only") {
			enabledHubs, listErr := svc.GetAllEnabledHubs(ctx)
			if listErr != nil {
				return "", fmt.Errorf("failed to list integration hubs: %w", listErr)
			}
			for _, h := range enabledHubs {
				if h != nil {
					hubs = append(hubs, *h)
				}
			}
		} else {
			hubs, err = svc.ListHubs(ctx)
			if err != nil {
				return "", fmt.Errorf("failed to list integration hubs: %w", err)
			}
		}

		type item struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Version     int    `json:"version"`
			Status      string `json:"status"`
			Enabled     bool   `json:"enabled"`
			Description string `json:"description"`
		}
		var items []item
		for _, h := range hubs {
			items = append(items, item{
				ID:          h.ID,
				Name:        h.Name,
				Version:     h.Version,
				Status:      string(h.Status),
				Enabled:     h.Enabled,
				Description: h.Description,
			})
		}
		if items == nil {
			items = []item{}
		}
		out, _ := json.Marshal(items)
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "get_integration_hub",
			Description: "Retrieves the full configuration of an Integration Hub by ID. " +
				"Use list_integration_hubs to find available hub IDs.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id": {Type: "string", Description: "Integration Hub ID (required)"},
				},
				Required: []string{"id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}
		svc := NewIntegrationHubService(s.docDB)
		hub, err := svc.GetHub(ctx, id)
		if err != nil {
			return "", fmt.Errorf("integration hub not found: %w", err)
		}
		out, _ := json.Marshal(hub)
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_integration_hub",
			Description: "Creates a new Integration Hub configuration. " +
				"An integration hub defines inbound/outbound protocol connections. " +
				"For simple hub creation, provide name and description; add protocol details as metadata_json. " +
				"Returns the new hub ID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":        {Type: "string", Description: "Hub name (required)"},
					"description": {Type: "string", Description: "Hub description"},
					"status":      {Type: "string", Description: "Initial status: dev | test | production (default: dev)"},
					"enabled":     {Type: "boolean", Description: "Whether hub is enabled (default: false)"},
					"metadata_json": {
						Type: "string",
						Description: `Optional metadata JSON for additional hub configuration. Example: {"tags":["api","webhook"],"notes":"Created by agent"}`,
					},
				},
				Required: []string{"name"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		name := stringArg(args, "name")
		if name == "" {
			return "", fmt.Errorf("name is required")
		}
		statusStr := stringArg(args, "status")
		if statusStr == "" {
			statusStr = "dev"
		}

		hub := &iacmodels.IntegrationHub{
			ID:          uuid.New().String(),
			Name:        name,
			Description: stringArg(args, "description"),
			Version:     1,
			IsDefault:   true,
			Enabled:     boolArg(args, "enabled"),
			Active:      true,
			Status:      iacmodels.IntegrationHubStatus(statusStr),
		}
		if metaJSON := stringArg(args, "metadata_json"); metaJSON != "" {
			var meta map[string]interface{}
			if json.Unmarshal([]byte(metaJSON), &meta) == nil {
				hub.Metadata = meta
			}
		}

		svc := NewIntegrationHubService(s.docDB)
		if err := svc.SaveHub(ctx, hub, "agent"); err != nil {
			return "", fmt.Errorf("failed to create integration hub: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": hub.ID, "name": hub.Name, "status": "created"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "update_integration_hub_status",
			Description: "Updates the status or enabled state of an Integration Hub. " +
				"Use this to promote a hub from dev → test → production, or to enable/disable it.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":      {Type: "string", Description: "Integration Hub ID (required)"},
					"status":  {Type: "string", Description: "New status: dev | test | production | deprecated"},
					"enabled": {Type: "string", Description: "Set to 'true' to enable or 'false' to disable (optional)"},
				},
				Required: []string{"id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}
		svc := NewIntegrationHubService(s.docDB)

		if statusStr := stringArg(args, "status"); statusStr != "" {
			if err := svc.UpdateStatus(ctx, id, iacmodels.IntegrationHubStatus(statusStr), "agent"); err != nil {
				return "", fmt.Errorf("failed to update hub status: %w", err)
			}
		}

		// Toggle enabled if specified
		if enabledStr := stringArg(args, "enabled"); enabledStr != "" {
			hub, err := svc.GetHub(ctx, id)
			if err != nil {
				return "", fmt.Errorf("hub not found: %w", err)
			}
			hub.Enabled = enabledStr == "true"
			if err := svc.SaveHub(ctx, hub, "agent"); err != nil {
				return "", fmt.Errorf("failed to update hub enabled state: %w", err)
			}
		}

		out, _ := json.Marshal(map[string]string{"id": id, "status": "updated"})
		return string(out), nil
	})
}

// ═══════════════════════════════════════════════════════════════════════════════
// JOBS — Scheduled Jobs + Queue Jobs (PostgreSQL)
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerJobTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "list_scheduled_jobs",
			Description: "Lists scheduled job definitions in IAC with their handler, cron expression, and next run time. " +
				"Optionally filter by enabled state.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"enabled_only": {Type: "boolean", Description: "If true, return only enabled jobs (default: false)"},
					"limit":        {Type: "integer", Description: "Maximum number of jobs to return (default 30)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		limit := 30
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}
		query := `SELECT id, name, description, handler, cronexpression, intervalseconds,
		               enabled, priority, nextrunat, lastrunat
		          FROM jobs WHERE active = true`
		if boolArg(args, "enabled_only") {
			query += ` AND enabled = true`
		}
		query += fmt.Sprintf(` ORDER BY name ASC LIMIT %d`, limit)

		rows, err := s.sqlDB.QueryContext(ctx, query)
		if err != nil {
			if isTableNotFoundErr(err) {
				out, _ := json.Marshal(map[string]interface{}{"rows": []interface{}{}, "note": "jobs table not configured"})
				return string(out), nil
			}
			return "", fmt.Errorf("query error: %w", err)
		}
		defer rows.Close()
		return s.rowsToJSON(rows)
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_scheduled_job",
			Description: "Creates a new scheduled cron/interval job in IAC. " +
				"IMPORTANT: The handler MUST be a TranCode name (e.g. 'ReportDaily', 'SyncInventory') — " +
				"NOT a script path, shell command, or Python/Go code. " +
				"TranCodes are Go-based business logic units registered in the IAC BPM engine. " +
				"Provide either a cron expression OR an interval in seconds — not both. " +
				"Use run_scheduled_job to execute an existing scheduled job immediately. " +
				"Returns the new job ID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":             {Type: "string", Description: "Job name (required, unique)"},
					"description":      {Type: "string", Description: "Job description"},
					"handler":          {Type: "string", Description: "TranCode name to execute on schedule (required). Must be a registered TranCode name, NOT a script or command."},
					"cron_expression":  {Type: "string", Description: `Cron expression for scheduling. Examples: "0 9 * * 1-5" (weekdays 9am), "*/15 * * * *" (every 15 min). Either this or interval_seconds is required.`},
					"interval_seconds": {Type: "integer", Description: "Run every N seconds (alternative to cron). Either this or cron_expression is required."},
					"enabled":          {Type: "boolean", Description: "Whether the job starts enabled (default: true)"},
					"priority":         {Type: "integer", Description: "Job priority, higher = runs first (default: 5)"},
					"max_retries":      {Type: "integer", Description: "Maximum retry attempts on failure (default: 3)"},
					"timeout":          {Type: "integer", Description: "Execution timeout in seconds (default: 300)"},
					"metadata_json":    {Type: "string", Description: `Optional metadata JSON: {"tags":["report","nightly"],"notify_on_failure":"admin@example.com"}`},
				},
				Required: []string{"name", "handler"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		name := stringArg(args, "name")
		handler := stringArg(args, "handler")
		if name == "" || handler == "" {
			return "", fmt.Errorf("name and handler are required")
		}
		cronExpr := stringArg(args, "cron_expression")
		intervalSec := 0
		if v, ok := args["interval_seconds"].(float64); ok {
			intervalSec = int(v)
		}
		if cronExpr == "" && intervalSec == 0 {
			return "", fmt.Errorf("either cron_expression or interval_seconds is required")
		}

		enabled := true
		if v, ok := args["enabled"].(bool); ok {
			enabled = v
		}
		priority := 5
		if v, ok := args["priority"].(float64); ok && v > 0 {
			priority = int(v)
		}
		maxRetries := 3
		if v, ok := args["max_retries"].(float64); ok && v > 0 {
			maxRetries = int(v)
		}
		timeout := 300
		if v, ok := args["timeout"].(float64); ok && v > 0 {
			timeout = int(v)
		}

		metaJSON := `{}`
		if metaStr := stringArg(args, "metadata_json"); metaStr != "" {
			metaJSON = metaStr
		}

		jobID := uuid.New().String()
		now := time.Now().UTC()

		var cronExprPtr interface{}
		if cronExpr != "" {
			cronExprPtr = cronExpr
		}
		var intervalSecPtr interface{}
		if intervalSec > 0 {
			intervalSecPtr = intervalSec
		}

		_, err := s.sqlDB.ExecContext(ctx, `
			INSERT INTO jobs (
				id, name, description, typeid, handler, cronexpression, intervalseconds,
				executioncount, enabled, "condition", priority, maxretries, timeout, metadata,
				active, createdby, createdon, modifiedby, modifiedon, rowversionstamp
			) VALUES ($1,$2,$3,$4,$5,$6,$7,0,$8,NULL,$9,$10,$11,$12,true,'agent',$13,'agent',$13,1)`,
			jobID, name, stringArg(args, "description"), int(iacmodels.JobTypeScheduled),
			handler, cronExprPtr, intervalSecPtr,
			enabled, priority, maxRetries, timeout, metaJSON, now,
		)
		if err != nil {
			return "", fmt.Errorf("failed to create scheduled job: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": jobID, "name": name, "status": "created"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "toggle_scheduled_job",
			Description: "Enables or disables a scheduled job by ID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":      {Type: "string", Description: "Scheduled job ID (required)"},
					"enabled": {Type: "boolean", Description: "true to enable, false to disable (required)"},
				},
				Required: []string{"id", "enabled"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}
		enabled, ok := args["enabled"].(bool)
		if !ok {
			return "", fmt.Errorf("enabled (boolean) is required")
		}
		_, err := s.sqlDB.ExecContext(ctx,
			`UPDATE jobs SET enabled = $1, modifiedby = 'agent', modifiedon = $2, rowversionstamp = rowversionstamp + 1 WHERE id = $3`,
			enabled, time.Now().UTC(), id)
		if err != nil {
			return "", fmt.Errorf("failed to toggle job: %w", err)
		}
		state := "enabled"
		if !enabled {
			state = "disabled"
		}
		out, _ := json.Marshal(map[string]string{"id": id, "state": state})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "list_queue_jobs",
			Description: "Lists jobs in the IAC execution queue. " +
				"Status: 1=Pending, 2=Queued, 3=Processing, 4=Completed, 5=Failed, 6=Retrying, 7=Cancelled.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"status": {Type: "string", Description: "Filter by status ID: 1 (Pending) | 2 (Queued) | 3 (Processing) | 4 (Completed) | 5 (Failed) | 6 (Retrying) | 7 (Cancelled)"},
					"limit":  {Type: "integer", Description: "Maximum number of jobs to return (default 20)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		limit := 20
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}
		query := `SELECT id, typeid, method, protocol, direction, handler, statusid, priority,
		               retrycount, maxretries, scheduledat, startedat, completedat, lasterror, createdon
		          FROM queue_jobs WHERE active = true`
		var queryArgs []interface{}
		if statusStr := stringArg(args, "status"); statusStr != "" {
			query += ` AND statusid = $1`
			queryArgs = append(queryArgs, statusStr)
		}
		query += fmt.Sprintf(` ORDER BY createdon DESC LIMIT %d`, limit)

		rows, err := s.sqlDB.QueryContext(ctx, query, queryArgs...)
		if err != nil {
			if isTableNotFoundErr(err) {
				out, _ := json.Marshal(map[string]interface{}{"rows": []interface{}{}, "note": "queue_jobs table not configured"})
				return string(out), nil
			}
			return "", fmt.Errorf("query error: %w", err)
		}
		defer rows.Close()
		return s.rowsToJSON(rows)
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "create_queue_job",
			Description: "Enqueues a one-time job for immediate execution by the IAC job worker. " +
				"IMPORTANT: The handler MUST be a TranCode name (e.g. 'ProcessOrder', 'GenerateReport') — " +
				"NOT a script path, shell command, or Python/Go code. " +
				"The worker executes TranCodes using the IAC BPM engine. " +
				"Returns the new queue job ID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"handler":       {Type: "string", Description: "TranCode name to execute (required). Must be a registered TranCode, NOT a script or command."},
					"method":        {Type: "string", Description: "Execution method (default: 'trancode')"},
					"payload":       {Type: "string", Description: "JSON payload/input data passed to the TranCode handler"},
					"priority":      {Type: "integer", Description: "Priority level — higher number = higher priority (default: 5)"},
					"max_retries":   {Type: "integer", Description: "Maximum retry attempts (default: 3)"},
					"metadata_json": {Type: "string", Description: `Optional metadata JSON: {"source":"agent","reason":"on-demand"}`},
				},
				Required: []string{"handler"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		handler := stringArg(args, "handler")
		if handler == "" {
			return "", fmt.Errorf("handler is required")
		}

		method := stringArg(args, "method")
		if method == "" {
			method = "trancode"
		}
		priority := 5
		if v, ok := args["priority"].(float64); ok && v > 0 {
			priority = int(v)
		}
		maxRetries := 3
		if v, ok := args["max_retries"].(float64); ok && v > 0 {
			maxRetries = int(v)
		}
		payload := stringArg(args, "payload")
		if payload == "" {
			payload = "{}"
		}
		metaJSON := `{}`
		if metaStr := stringArg(args, "metadata_json"); metaStr != "" {
			metaJSON = metaStr
		}

		jobID := uuid.New().String()
		now := time.Now().UTC()

		_, err := s.sqlDB.ExecContext(ctx, `
			INSERT INTO queue_jobs (
				id, typeid, method, protocol, direction, handler, payload, metadata,
				statusid, priority, maxretries, retrycount,
				active, createdby, createdon, modifiedby, modifiedon, rowversionstamp
			) VALUES ($1,$2,$3,'internal','internal',$4,$5,$6,$7,$8,$9,0,true,'agent',$10,'agent',$10,1)`,
			jobID, int(iacmodels.JobTypeManual), method,
			handler, payload, metaJSON,
			int(iacmodels.JobStatusPending), priority, maxRetries, now,
		)
		if err != nil {
			return "", fmt.Errorf("failed to create queue job: %w", err)
		}
		out, _ := json.Marshal(map[string]string{"id": jobID, "handler": handler, "status": "queued"})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "get_job_statistics",
			Description: "Returns job system statistics: total, pending, queued, processing, completed, failed, and cancelled job counts.",
			Parameters:  ToolParameterSchema{Type: "object", Properties: map[string]ToolPropertySchema{}},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		svc := NewJobService(s.sqlDB)
		stats, err := svc.GetJobStatistics(ctx)
		if err != nil {
			if isTableNotFoundErr(err) {
				out, _ := json.Marshal(map[string]interface{}{"note": "jobs tables not yet configured"})
				return string(out), nil
			}
			return "", fmt.Errorf("failed to get job statistics: %w", err)
		}
		out, _ := json.Marshal(stats)
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "run_scheduled_job",
			Description: "Immediately executes a scheduled job by placing it in the execution queue. " +
				"Reads the job definition (handler, priority, max_retries) from the scheduled job and " +
				"creates a queue_jobs entry with Pending status so the worker picks it up right away. " +
				"Use list_scheduled_jobs to find job IDs.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":      {Type: "string", Description: "Scheduled job ID to execute now (required)"},
					"payload": {Type: "string", Description: "Optional JSON payload to pass to the handler (overrides job default)"},
				},
				Required: []string{"id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}

		// Load the scheduled job definition
		var name, handler, description string
		var priority, maxRetries int
		err := s.sqlDB.QueryRowContext(ctx,
			`SELECT name, handler, description, priority, maxretries FROM jobs WHERE id = $1 AND active = true`, id,
		).Scan(&name, &handler, &description, &priority, &maxRetries)
		if err != nil {
			return "", fmt.Errorf("scheduled job not found (id=%s): %w", id, err)
		}

		payload := stringArg(args, "payload")
		if payload == "" {
			payload = "{}"
		}

		queueJobID := uuid.New().String()
		now := time.Now().UTC()
		_, err = s.sqlDB.ExecContext(ctx, `
			INSERT INTO queue_jobs (
				id, typeid, method, protocol, direction, handler, payload, metadata,
				statusid, priority, maxretries, retrycount,
				active, createdby, createdon, modifiedby, modifiedon, rowversionstamp
			) VALUES ($1,$2,'trancode','internal','internal',$3,$4,'{"source":"agent_run_now"}',
			          $5,$6,$7,0,true,'agent',$8,'agent',$8,1)`,
			queueJobID, int(iacmodels.JobTypeManual),
			handler, payload,
			int(iacmodels.JobStatusPending), priority, maxRetries, now,
		)
		if err != nil {
			return "", fmt.Errorf("failed to queue job for immediate execution: %w", err)
		}

		out, _ := json.Marshal(map[string]string{
			"queue_job_id":   queueJobID,
			"scheduled_job":  name,
			"handler":        handler,
			"status":         "queued_for_immediate_execution",
		})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "get_job_history",
			Description: "Returns recent execution history for queue jobs, optionally filtered by scheduled job handler. " +
				"Shows status, start time, completion time, and any errors.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"handler": {Type: "string", Description: "Filter by handler name (TranCode name). Omit to see all recent jobs."},
					"status":  {Type: "string", Description: "Filter by status: 4 (Completed) | 5 (Failed) | 3 (Processing)"},
					"limit":   {Type: "integer", Description: "Maximum number of records to return (default: 20)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		limit := 20
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}

		query := `SELECT id, handler, statusid, startedat, completedat, lasterror, createdon
		          FROM queue_jobs WHERE active = true`
		var queryArgs []interface{}
		argIdx := 1

		if handler := stringArg(args, "handler"); handler != "" {
			query += fmt.Sprintf(` AND handler = $%d`, argIdx)
			queryArgs = append(queryArgs, handler)
			argIdx++
		}
		if status := stringArg(args, "status"); status != "" {
			query += fmt.Sprintf(` AND statusid = $%d`, argIdx)
			queryArgs = append(queryArgs, status)
			argIdx++
		}
		query += fmt.Sprintf(` ORDER BY createdon DESC LIMIT %d`, limit)

		rows, err := s.sqlDB.QueryContext(ctx, query, queryArgs...)
		if err != nil {
			if isTableNotFoundErr(err) {
				out, _ := json.Marshal(map[string]interface{}{"rows": []interface{}{}, "note": "queue_jobs table not configured"})
				return string(out), nil
			}
			return "", fmt.Errorf("query error: %w", err)
		}
		defer rows.Close()
		return s.rowsToJSON(rows)
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "get_queue_job_status",
			Description: "Returns the current status and details of a specific queue job by ID. " +
				"Use this to poll a job after creation until it reaches Completed or Failed. " +
				"Status codes: 1=Pending, 2=Queued, 3=Processing, 4=Completed, 5=Failed, 6=Retrying, 7=Cancelled.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id": {Type: "string", Description: "Queue job ID (required)"},
				},
				Required: []string{"id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}
		rows, err := s.sqlDB.QueryContext(ctx,
			`SELECT id, handler, statusid, priority, retrycount, maxretries,
			        scheduledat, startedat, completedat, lasterror, payload, createdon
			 FROM queue_jobs WHERE id = $1`, id)
		if err != nil {
			return "", fmt.Errorf("query error: %w", err)
		}
		defer rows.Close()
		return s.rowsToJSON(rows)
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "update_scheduled_job",
			Description: "Updates an existing scheduled job's settings. " +
				"Only fields explicitly provided will be changed; omitted fields keep their current values. " +
				"Supply either cron_expression or interval_seconds to change the schedule — not both.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":               {Type: "string", Description: "Scheduled job ID to update (required)"},
					"name":             {Type: "string", Description: "New job name"},
					"description":      {Type: "string", Description: "New description"},
					"handler":          {Type: "string", Description: "New TranCode handler name"},
					"cron_expression":  {Type: "string", Description: "New cron expression (clears interval_seconds if set)"},
					"interval_seconds": {Type: "integer", Description: "New interval in seconds (clears cron_expression if set)"},
					"enabled":          {Type: "boolean", Description: "Enable or disable the job"},
					"priority":         {Type: "integer", Description: "New priority level"},
					"max_retries":      {Type: "integer", Description: "New maximum retry count"},
					"timeout":          {Type: "integer", Description: "New execution timeout in seconds"},
					"metadata_json":    {Type: "string", Description: `Replacement metadata JSON object, e.g. {"tags":["report"],"notify_on_failure":"admin@example.com"}`},
				},
				Required: []string{"id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}

		// Build dynamic SET clause only for provided fields
		sets := []string{"modifiedby = 'agent'", "modifiedon = $1", "rowversionstamp = rowversionstamp + 1"}
		vals := []interface{}{time.Now().UTC()}
		idx := 2

		addStr := func(col, key string) {
			if v := stringArg(args, key); v != "" {
				sets = append(sets, fmt.Sprintf("%s = $%d", col, idx))
				vals = append(vals, v)
				idx++
			}
		}
		addStr("name", "name")
		addStr("description", "description")
		addStr("handler", "handler")

		if cron := stringArg(args, "cron_expression"); cron != "" {
			sets = append(sets, fmt.Sprintf("cronexpression = $%d, intervalseconds = NULL", idx))
			vals = append(vals, cron)
			idx++
		} else if v, ok := args["interval_seconds"].(float64); ok && v > 0 {
			sets = append(sets, fmt.Sprintf("intervalseconds = $%d, cronexpression = NULL", idx))
			vals = append(vals, int(v))
			idx++
		}

		if v, ok := args["enabled"].(bool); ok {
			sets = append(sets, fmt.Sprintf("enabled = $%d", idx))
			vals = append(vals, v)
			idx++
		}
		if v, ok := args["priority"].(float64); ok {
			sets = append(sets, fmt.Sprintf("priority = $%d", idx))
			vals = append(vals, int(v))
			idx++
		}
		if v, ok := args["max_retries"].(float64); ok {
			sets = append(sets, fmt.Sprintf("maxretries = $%d", idx))
			vals = append(vals, int(v))
			idx++
		}
		if v, ok := args["timeout"].(float64); ok {
			sets = append(sets, fmt.Sprintf("timeout = $%d", idx))
			vals = append(vals, int(v))
			idx++
		}
		if metaStr := stringArg(args, "metadata_json"); metaStr != "" {
			sets = append(sets, fmt.Sprintf("metadata = $%d", idx))
			vals = append(vals, metaStr)
			idx++
		}

		vals = append(vals, id)
		query := fmt.Sprintf("UPDATE jobs SET %s WHERE id = $%d AND active = true",
			strings.Join(sets, ", "), idx)
		res, err := s.sqlDB.ExecContext(ctx, query, vals...)
		if err != nil {
			return "", fmt.Errorf("failed to update scheduled job: %w", err)
		}
		n, _ := res.RowsAffected()
		out, _ := json.Marshal(map[string]interface{}{"id": id, "rows_updated": n})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "delete_scheduled_job",
			Description: "Soft-deletes a scheduled job by setting it inactive. " +
				"The job record is retained for audit purposes but will no longer be executed. " +
				"Use toggle_scheduled_job to temporarily disable without deleting.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id": {Type: "string", Description: "Scheduled job ID to delete (required)"},
				},
				Required: []string{"id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}
		now := time.Now().UTC()
		res, err := s.sqlDB.ExecContext(ctx,
			`UPDATE jobs SET active = false, enabled = false, modifiedby = 'agent', modifiedon = $1,
			 rowversionstamp = rowversionstamp + 1 WHERE id = $2 AND active = true`, now, id)
		if err != nil {
			return "", fmt.Errorf("failed to delete scheduled job: %w", err)
		}
		n, _ := res.RowsAffected()
		out, _ := json.Marshal(map[string]interface{}{"id": id, "deleted": n > 0})
		return string(out), nil
	})
}

// sqlNullString wraps an optional string as a sql.NullString
func sqlNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
