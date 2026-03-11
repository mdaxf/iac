package services

// OS management tools — allow agents to control OS services and inspect system health.
//
// - list_services        : enumerate OS services (Windows SC / Linux systemd)
// - get_service_status   : get current status of a named service
// - start_service        : start a stopped service
// - stop_service         : stop a running service
// - restart_service      : restart a service (stop → start on Windows; systemctl restart on Linux)
// - diagnose_system      : comprehensive health check (OS, CPU, disk, memory, DB, services)
// - get_process_list     : list running processes, optionally filtered by name

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// validServiceName matches legal service/unit names on both Windows and Linux.
var validServiceName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_\-\.@]*$`)

func validateServiceName(name string) error {
	if name == "" {
		return fmt.Errorf("service name is required")
	}
	if len(name) > 256 {
		return fmt.Errorf("service name is too long (max 256 chars)")
	}
	if !validServiceName.MatchString(name) {
		return fmt.Errorf("invalid service name %q: only alphanumeric, underscore, hyphen, dot, @ allowed", name)
	}
	return nil
}

// runCmd executes a command with a deadline and returns stdout, stderr, exit code.
func runCmd(ctx context.Context, timeoutSec int, name string, args ...string) (stdout, stderr string, exitCode int, err error) {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			err = nil // non-zero exit is a valid result, not a Go error
		} else {
			err = runErr
		}
	}
	return
}

func (s *AgentToolRegistryService) registerOSTools() {
	s.registerServiceTools()
	s.registerDiagnosticTools()
}

// ═══════════════════════════════════════════════════════════════════════════════
// SERVICE CONTROL
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerServiceTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "list_services",
			Description: "Lists OS services on the host. " +
				"On Windows, uses the Service Control Manager (sc query). " +
				"On Linux, uses systemd (systemctl list-units). " +
				"Filter by state: 'running', 'stopped', or 'all' (default). " +
				"Use this before start_service or stop_service to find the correct service name.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"filter": {Type: "string", Description: "Optional name filter (case-insensitive substring match)"},
					"state":  {Type: "string", Description: "State filter: 'running' | 'stopped' | 'all' (default: 'all')"},
					"limit":  {Type: "integer", Description: "Max services to return (default: 50)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		filter := strings.ToLower(stringArg(args, "filter"))
		state := strings.ToLower(stringArg(args, "state"))
		limit := 50
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}

		var services []map[string]string
		var err error
		if runtime.GOOS == "windows" {
			services, err = listServicesWindows(ctx, state)
		} else {
			services, err = listServicesLinux(ctx, state)
		}
		if err != nil {
			return "", fmt.Errorf("failed to list services: %w", err)
		}

		// Apply name filter
		if filter != "" {
			var filtered []map[string]string
			for _, svc := range services {
				if strings.Contains(strings.ToLower(svc["name"]), filter) ||
					strings.Contains(strings.ToLower(svc["display_name"]), filter) {
					filtered = append(filtered, svc)
				}
			}
			services = filtered
		}

		// Apply limit
		if len(services) > limit {
			services = services[:limit]
		}

		out, _ := json.Marshal(map[string]interface{}{
			"count":    len(services),
			"os":       runtime.GOOS,
			"services": services,
		})
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "get_service_status",
			Description: "Returns the current status and details of a named OS service. " +
				"Use list_services to discover the correct service name first.",
			Parameters: ToolParameterSchema{
				Type:     "object",
				Properties: map[string]ToolPropertySchema{
					"name": {Type: "string", Description: "Service name (required). On Linux this is the systemd unit name (e.g. 'nginx.service' or 'nginx')."},
				},
				Required: []string{"name"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		name := stringArg(args, "name")
		if err := validateServiceName(name); err != nil {
			return "", err
		}
		return getServiceStatus(ctx, name)
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "start_service",
			Description: "Starts a stopped OS service. " +
				"Requires the exact service name — use list_services or get_service_status to verify. " +
				"Returns the new service status.",
			Parameters: ToolParameterSchema{
				Type:     "object",
				Properties: map[string]ToolPropertySchema{
					"name": {Type: "string", Description: "Service name to start (required)"},
				},
				Required: []string{"name"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		name := stringArg(args, "name")
		if err := validateServiceName(name); err != nil {
			return "", err
		}
		return controlService(ctx, "start", name)
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "stop_service",
			Description: "Stops a running OS service. " +
				"Requires the exact service name — use list_services to verify. " +
				"Returns the new service status.",
			Parameters: ToolParameterSchema{
				Type:     "object",
				Properties: map[string]ToolPropertySchema{
					"name": {Type: "string", Description: "Service name to stop (required)"},
				},
				Required: []string{"name"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		name := stringArg(args, "name")
		if err := validateServiceName(name); err != nil {
			return "", err
		}
		return controlService(ctx, "stop", name)
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "restart_service",
			Description: "Restarts an OS service (stop then start, or systemctl restart on Linux). " +
				"Use this to apply configuration changes without a full system reboot.",
			Parameters: ToolParameterSchema{
				Type:     "object",
				Properties: map[string]ToolPropertySchema{
					"name": {Type: "string", Description: "Service name to restart (required)"},
				},
				Required: []string{"name"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		name := stringArg(args, "name")
		if err := validateServiceName(name); err != nil {
			return "", err
		}
		return controlService(ctx, "restart", name)
	})
}

// ─── service helpers ──────────────────────────────────────────────────────────

func getServiceStatus(ctx context.Context, name string) (string, error) {
	var stdout, stderr string
	var exitCode int
	var err error

	if runtime.GOOS == "windows" {
		stdout, stderr, exitCode, err = runCmd(ctx, 15, "sc", "query", name)
	} else {
		stdout, stderr, exitCode, err = runCmd(ctx, 15, "systemctl", "status", name, "--no-pager", "-l")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get service status: %w", err)
	}

	out, _ := json.Marshal(map[string]interface{}{
		"service":   name,
		"os":        runtime.GOOS,
		"exit_code": exitCode,
		"output":    strings.TrimSpace(stdout),
		"error":     strings.TrimSpace(stderr),
	})
	return string(out), nil
}

func controlService(ctx context.Context, action, name string) (string, error) {
	var stdout, stderr string
	var exitCode int
	var err error

	start := time.Now()
	if runtime.GOOS == "windows" {
		switch action {
		case "start":
			stdout, stderr, exitCode, err = runCmd(ctx, 60, "sc", "start", name)
		case "stop":
			stdout, stderr, exitCode, err = runCmd(ctx, 60, "sc", "stop", name)
		case "restart":
			// Windows SC has no native restart; stop then start
			runCmd(ctx, 60, "sc", "stop", name) //nolint:errcheck
			time.Sleep(2 * time.Second)
			stdout, stderr, exitCode, err = runCmd(ctx, 60, "sc", "start", name)
		}
	} else {
		stdout, stderr, exitCode, err = runCmd(ctx, 60, "systemctl", action, name)
	}
	if err != nil {
		return "", fmt.Errorf("failed to %s service %q: %w", action, name, err)
	}

	success := exitCode == 0
	out, _ := json.Marshal(map[string]interface{}{
		"service":     name,
		"action":      action,
		"success":     success,
		"exit_code":   exitCode,
		"output":      strings.TrimSpace(stdout),
		"error":       strings.TrimSpace(stderr),
		"duration_ms": time.Since(start).Milliseconds(),
	})
	return string(out), nil
}

func listServicesWindows(ctx context.Context, state string) ([]map[string]string, error) {
	args := []string{"query", "type=", "all"}
	switch state {
	case "running":
		args = append(args, "state=", "running")
	case "stopped":
		args = append(args, "state=", "inactive")
	default:
		args = append(args, "state=", "all")
	}
	stdout, _, _, err := runCmd(ctx, 30, "sc", args...)
	if err != nil {
		return nil, err
	}

	// Parse sc query output — blocks separated by blank lines
	var services []map[string]string
	var current map[string]string
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			if current != nil {
				services = append(services, current)
				current = nil
			}
			continue
		}
		if strings.HasPrefix(line, "SERVICE_NAME:") {
			current = map[string]string{
				"name": strings.TrimSpace(strings.TrimPrefix(line, "SERVICE_NAME:")),
			}
		} else if current != nil {
			if strings.Contains(line, "DISPLAY_NAME:") {
				current["display_name"] = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			} else if strings.Contains(line, "STATE") && strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					stateVal := strings.TrimSpace(parts[1])
					// STATE : 4  RUNNING  →  extract word after digit
					fields := strings.Fields(stateVal)
					if len(fields) >= 2 {
						current["state"] = fields[1]
					}
				}
			}
		}
	}
	if current != nil {
		services = append(services, current)
	}
	return services, nil
}

func listServicesLinux(ctx context.Context, state string) ([]map[string]string, error) {
	args := []string{"list-units", "--type=service", "--no-pager", "--plain", "--no-legend"}
	switch state {
	case "running":
		args = append(args, "--state=running")
	case "stopped":
		args = append(args, "--state=dead,failed")
	default:
		args = append(args, "--all")
	}
	stdout, _, _, err := runCmd(ctx, 30, "systemctl", args...)
	if err != nil {
		return nil, err
	}

	var services []map[string]string
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		desc := ""
		if len(fields) > 4 {
			desc = strings.Join(fields[4:], " ")
		}
		services = append(services, map[string]string{
			"name":         fields[0],
			"display_name": desc,
			"load":         fields[1],
			"active":       fields[2],
			"state":        fields[3],
		})
	}
	return services, nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// SYSTEM DIAGNOSTICS
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerDiagnosticTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "diagnose_system",
			Description: "Runs a comprehensive system health check and returns a structured report. " +
				"Checks: OS info, CPU, disk, memory, database connectivity, and optionally IAC service status. " +
				"The returned JSON report can be passed to generate_html_report and sent via send_email. " +
				"Sections (comma-separated): 'os', 'disk', 'memory', 'database', 'services', 'all' (default: all).",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"sections":       {Type: "string", Description: "Comma-separated sections to collect: os|disk|memory|database|services|all (default: all)"},
					"service_filter": {Type: "string", Description: "Optional service name filter for the services section (e.g. 'iac', 'nginx')"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		sectionsArg := strings.ToLower(stringArg(args, "sections"))
		if sectionsArg == "" {
			sectionsArg = "all"
		}
		wantAll := sectionsArg == "all"
		want := func(s string) bool { return wantAll || strings.Contains(sectionsArg, s) }

		report := map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"os":        runtime.GOOS,
			"arch":      runtime.GOARCH,
		}

		// OS + CPU + Go runtime
		if want("os") {
			hostname, _ := os.Hostname()
			report["hostname"] = hostname
			report["cpus"] = runtime.NumCPU()
			report["go_version"] = runtime.Version()
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			report["process_heap_mb"] = ms.HeapAlloc / 1024 / 1024
			report["goroutines"] = runtime.NumGoroutine()
		}

		// Disk space
		if want("disk") {
			report["disk"] = collectDiskInfo(ctx)
		}

		// Memory
		if want("memory") {
			report["memory"] = collectMemoryInfo(ctx)
		}

		// Database
		if want("database") {
			if s.sqlDB != nil {
				dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				if err := s.sqlDB.PingContext(dbCtx); err != nil {
					report["database"] = map[string]string{"status": "error", "error": err.Error()}
				} else {
					var dbStats = s.sqlDB.Stats()
					report["database"] = map[string]interface{}{
						"status":           "connected",
						"open_connections": dbStats.OpenConnections,
						"in_use":           dbStats.InUse,
						"idle":             dbStats.Idle,
					}
				}
			} else {
				report["database"] = map[string]string{"status": "not_configured"}
			}
		}

		// Services
		if want("services") {
			filter := stringArg(args, "service_filter")
			var services []map[string]string
			var err error
			if runtime.GOOS == "windows" {
				services, err = listServicesWindows(ctx, "running")
			} else {
				services, err = listServicesLinux(ctx, "running")
			}
			if err == nil && filter != "" {
				filter = strings.ToLower(filter)
				var filtered []map[string]string
				for _, svc := range services {
					if strings.Contains(strings.ToLower(svc["name"]), filter) {
						filtered = append(filtered, svc)
					}
				}
				services = filtered
			}
			report["running_services"] = services
			if err != nil {
				report["services_error"] = err.Error()
			}
		}

		// Overall status heuristic
		status := "healthy"
		if db, ok := report["database"].(map[string]interface{}); ok {
			if db["status"] != "connected" {
				status = "degraded"
			}
		}
		if db, ok := report["database"].(map[string]string); ok {
			if db["status"] == "error" {
				status = "degraded"
			}
		}
		report["overall_status"] = status

		out, _ := json.Marshal(report)
		return string(out), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "get_process_list",
			Description: "Lists running processes on the host. " +
				"On Linux uses 'ps aux'; on Windows uses 'tasklist'. " +
				"Use the filter parameter to narrow results by process name.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"filter": {Type: "string", Description: "Optional case-insensitive substring filter on process name"},
					"limit":  {Type: "integer", Description: "Max processes to return (default: 30)"},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		filter := strings.ToLower(stringArg(args, "filter"))
		limit := 30
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}

		var stdout string
		var err error
		if runtime.GOOS == "windows" {
			stdout, _, _, err = runCmd(ctx, 20, "tasklist", "/FO", "CSV", "/NH")
		} else {
			stdout, _, _, err = runCmd(ctx, 20, "ps", "aux", "--no-header")
		}
		if err != nil {
			return "", fmt.Errorf("failed to list processes: %w", err)
		}

		lines := strings.Split(strings.TrimSpace(stdout), "\n")
		var result []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if filter == "" || strings.Contains(strings.ToLower(line), filter) {
				result = append(result, line)
				if len(result) >= limit {
					break
				}
			}
		}

		out, _ := json.Marshal(map[string]interface{}{
			"os":        runtime.GOOS,
			"count":     len(result),
			"processes": result,
		})
		return string(out), nil
	})
}

// ─── system info helpers ──────────────────────────────────────────────────────

func collectDiskInfo(ctx context.Context) interface{} {
	var stdout string
	var err error
	if runtime.GOOS == "windows" {
		stdout, _, _, err = runCmd(ctx, 15, "wmic", "logicaldisk", "get",
			"DeviceID,FreeSpace,Size", "/format:csv")
	} else {
		stdout, _, _, err = runCmd(ctx, 15, "df", "-h")
	}
	if err != nil {
		return map[string]string{"error": err.Error()}
	}
	return strings.TrimSpace(stdout)
}

func collectMemoryInfo(ctx context.Context) interface{} {
	var stdout string
	var err error
	if runtime.GOOS == "windows" {
		stdout, _, _, err = runCmd(ctx, 15, "wmic", "OS", "get",
			"FreePhysicalMemory,TotalVisibleMemorySize", "/format:csv")
	} else {
		stdout, _, _, err = runCmd(ctx, 15, "free", "-m")
	}
	if err != nil {
		return map[string]string{"error": err.Error()}
	}
	return strings.TrimSpace(stdout)
}
