package services

// Testing tools — allow agents to run tests and execute shell commands.
//
// - run_unit_tests    : run Go unit tests (go test ./...) in a working directory
// - run_system_tests  : HTTP health checks against a list of endpoints
// - run_ui_tests      : run Playwright UI test suite (npx playwright test)
// - run_shell_command : execute an arbitrary command with args (no shell interpolation)
//
// Security: run_shell_command uses exec.Command (not exec.Command("sh","-c",...))
// so no shell injection is possible. The command binary must exist in PATH.
// All executions are logged and subject to a configurable timeout (max 600s).

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (s *AgentToolRegistryService) registerTestingTools() {
	s.registerUnitTestsTool()
	s.registerSystemTestsTool()
	s.registerUITestsTool()
	s.registerShellCommandTool()
}

// ─── run_unit_tests ───────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerUnitTestsTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "run_unit_tests",
			Description: "Runs Go unit tests using 'go test' in the specified working directory. " +
				"Returns exit code, pass/fail counts extracted from output, and full test output. " +
				"Use pattern to run specific tests (e.g. 'TestAgent.*'). " +
				"Use packages to target specific packages (default: './...' = all packages).",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"working_dir":     {Type: "string", Description: "Directory containing go.mod to run tests from (required)"},
					"packages":        {Type: "string", Description: "Package path(s) to test (default: './...'). Example: './services/...' or './services/agentrunnerservice.go'"},
					"pattern":         {Type: "string", Description: "Test name filter regex (passed to -run). Example: 'TestAgent' runs all tests starting with TestAgent"},
					"verbose":         {Type: "boolean", Description: "Include per-test output (default: true)"},
					"timeout_seconds": {Type: "integer", Description: "Max test run time in seconds (default: 120, max: 600)"},
					"count":           {Type: "integer", Description: "Run each test N times (default: 1). Use count=1 to disable test caching."},
				},
				Required: []string{"working_dir"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		workDir := stringArg(args, "working_dir")
		if err := validateFSPath(workDir); err != nil {
			return "", err
		}
		if _, err := os.Stat(workDir); err != nil {
			return "", fmt.Errorf("working_dir not found: %w", err)
		}

		pkgs := stringArg(args, "packages")
		if pkgs == "" {
			pkgs = "./..."
		}
		verbose := true
		if v, ok := args["verbose"].(bool); ok {
			verbose = v
		}
		timeout := 120
		if v, ok := args["timeout_seconds"].(float64); ok && v > 0 {
			timeout = int(v)
		}
		if timeout > 600 {
			timeout = 600
		}
		count := 1
		if v, ok := args["count"].(float64); ok && v > 0 {
			count = int(v)
		}

		goArgs := []string{"test", fmt.Sprintf("-timeout=%ds", timeout)}
		if verbose {
			goArgs = append(goArgs, "-v")
		}
		if pattern := stringArg(args, "pattern"); pattern != "" {
			goArgs = append(goArgs, "-run", pattern)
		}
		goArgs = append(goArgs, fmt.Sprintf("-count=%d", count))
		goArgs = append(goArgs, pkgs)

		return s.runTestCommand(ctx, workDir, timeout+10, "go", goArgs...)
	})
}

// ─── run_system_tests ─────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerSystemTestsTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "run_system_tests",
			Description: "Runs HTTP health checks against a list of endpoints and reports status codes, latency, and response previews. " +
				"Use this to verify API availability after a deployment or service restart. " +
				"endpoints_json is an array of endpoint objects; base_url is prepended to each path.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"base_url": {Type: "string", Description: "Base URL to prepend (e.g. 'http://localhost:8080'). Required."},
					"endpoints_json": {
						Type: "string",
						Description: `JSON array of endpoint objects. Each: {"path":"/api/health","method":"GET","expect_status":200,"body":"{}"}. ` +
							`If omitted, checks ["/health","/api/health","/api/status"].`,
					},
					"timeout_seconds": {Type: "integer", Description: "Per-request timeout in seconds (default: 10)"},
					"headers_json":    {Type: "string", Description: `Optional headers for all requests: {"Authorization":"Bearer token"}`},
				},
				Required: []string{"base_url"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		baseURL := strings.TrimRight(stringArg(args, "base_url"), "/")
		if baseURL == "" {
			return "", fmt.Errorf("base_url is required")
		}

		type endpoint struct {
			Path         string `json:"path"`
			Method       string `json:"method"`
			ExpectStatus int    `json:"expect_status"`
			Body         string `json:"body"`
		}

		var endpoints []endpoint
		if ej := stringArg(args, "endpoints_json"); ej != "" {
			if err := json.Unmarshal([]byte(ej), &endpoints); err != nil {
				return "", fmt.Errorf("invalid endpoints_json: %w", err)
			}
		} else {
			// Default health check paths
			for _, p := range []string{"/health", "/api/health", "/api/status"} {
				endpoints = append(endpoints, endpoint{Path: p, Method: "GET", ExpectStatus: 200})
			}
		}

		timeoutSec := 10
		if v, ok := args["timeout_seconds"].(float64); ok && v > 0 {
			timeoutSec = int(v)
		}

		var globalHeaders map[string]string
		if hj := stringArg(args, "headers_json"); hj != "" {
			json.Unmarshal([]byte(hj), &globalHeaders) //nolint:errcheck
		}

		client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}

		type result struct {
			URL          string `json:"url"`
			Method       string `json:"method"`
			Status       int    `json:"status_code"`
			LatencyMs    int64  `json:"latency_ms"`
			Success      bool   `json:"success"`
			Error        string `json:"error,omitempty"`
			ResponseBody string `json:"response_preview,omitempty"`
		}
		var results []result
		passed, failed := 0, 0

		for _, ep := range endpoints {
			method := ep.Method
			if method == "" {
				method = "GET"
			}
			url := baseURL + ep.Path
			expectStatus := ep.ExpectStatus
			if expectStatus == 0 {
				expectStatus = 200
			}

			var bodyReader io.Reader
			if ep.Body != "" && method != "GET" {
				bodyReader = strings.NewReader(ep.Body)
			}

			req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
			if err != nil {
				results = append(results, result{URL: url, Method: method, Error: err.Error()})
				failed++
				continue
			}
			for k, v := range globalHeaders {
				req.Header.Set(k, v)
			}
			if ep.Body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			start := time.Now()
			resp, err := client.Do(req)
			latency := time.Since(start).Milliseconds()
			if err != nil {
				results = append(results, result{URL: url, Method: method, LatencyMs: latency, Error: err.Error()})
				failed++
				continue
			}
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			resp.Body.Close()

			success := resp.StatusCode == expectStatus
			if success {
				passed++
			} else {
				failed++
			}
			preview := strings.TrimSpace(string(body))
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			results = append(results, result{
				URL:          url,
				Method:       method,
				Status:       resp.StatusCode,
				LatencyMs:    latency,
				Success:      success,
				ResponseBody: preview,
			})
		}

		out, _ := json.Marshal(map[string]interface{}{
			"base_url": baseURL,
			"total":    len(results),
			"passed":   passed,
			"failed":   failed,
			"results":  results,
		})
		return string(out), nil
	})
}

// ─── run_ui_tests ─────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerUITestsTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "run_ui_tests",
			Description: "Runs Playwright UI tests using 'npx playwright test'. " +
				"Requires Node.js and Playwright to be installed in the test directory. " +
				"Returns exit code, test summary, and output. " +
				"Specify test_file to run a single spec file, or omit to run all tests.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"test_dir":        {Type: "string", Description: "Directory containing playwright.config.js/ts (required)"},
					"test_file":       {Type: "string", Description: "Specific test file to run (relative to test_dir, e.g. 'tests/login.spec.ts')"},
					"browser":         {Type: "string", Description: "Browser to use: chromium (default) | firefox | webkit"},
					"reporter":        {Type: "string", Description: "Playwright reporter: list (default) | dot | json | html"},
					"timeout_seconds": {Type: "integer", Description: "Max run time in seconds (default: 300, max: 600)"},
					"headed":          {Type: "boolean", Description: "Run in headed mode (shows browser window) — default: false (headless)"},
				},
				Required: []string{"test_dir"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		testDir := stringArg(args, "test_dir")
		if err := validateFSPath(testDir); err != nil {
			return "", err
		}
		if _, err := os.Stat(testDir); err != nil {
			return "", fmt.Errorf("test_dir not found: %w", err)
		}

		timeout := 300
		if v, ok := args["timeout_seconds"].(float64); ok && v > 0 {
			timeout = int(v)
		}
		if timeout > 600 {
			timeout = 600
		}

		browser := stringArg(args, "browser")
		if browser == "" {
			browser = "chromium"
		}
		reporter := stringArg(args, "reporter")
		if reporter == "" {
			reporter = "list"
		}

		pwArgs := []string{"playwright", "test",
			"--browser=" + browser,
			"--reporter=" + reporter,
		}
		if boolArg(args, "headed") {
			pwArgs = append(pwArgs, "--headed")
		}
		if testFile := stringArg(args, "test_file"); testFile != "" {
			pwArgs = append(pwArgs, testFile)
		}

		return s.runTestCommand(ctx, testDir, timeout+10, "npx", pwArgs...)
	})
}

// ─── run_shell_command ────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerShellCommandTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "run_shell_command",
			Description: "Executes an OS command directly (no shell interpolation). " +
				"The command binary must exist in PATH. Arguments are passed as an array — no shell quoting needed. " +
				"Returns stdout, stderr, exit code, and duration. " +
				"Use this for operations not covered by other tools: npm install, git pull, make, pip install, etc. " +
				"IMPORTANT: Never pass user-supplied values as shell arguments without review. Always prefer dedicated tools when available.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"command":         {Type: "string", Description: "Command binary to execute (required), e.g. 'go', 'npm', 'git', 'make', 'python3'"},
					"args_json":       {Type: "string", Description: `JSON array of arguments (default: []). Example: ["test","./..."] for 'go test ./...'`},
					"working_dir":     {Type: "string", Description: "Working directory for the command (default: current process directory)"},
					"env_json":        {Type: "string", Description: `Additional environment variables as JSON object: {"GO_ENV":"production","PORT":"8080"}. Merged with current process env.`},
					"timeout_seconds": {Type: "integer", Description: "Max execution time in seconds (default: 60, max: 600)"},
				},
				Required: []string{"command"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		command := stringArg(args, "command")
		if command == "" {
			return "", fmt.Errorf("command is required")
		}
		// Prevent absolute paths with traversal
		if strings.Contains(command, "..") {
			return "", fmt.Errorf("command path traversal (..) is not allowed")
		}

		var cmdArgs []string
		if aj := stringArg(args, "args_json"); aj != "" {
			if err := json.Unmarshal([]byte(aj), &cmdArgs); err != nil {
				return "", fmt.Errorf("invalid args_json: %w", err)
			}
		}

		workDir := stringArg(args, "working_dir")
		if workDir != "" {
			if err := validateFSPath(workDir); err != nil {
				return "", err
			}
		}

		timeout := 60
		if v, ok := args["timeout_seconds"].(float64); ok && v > 0 {
			timeout = int(v)
		}
		if timeout > 600 {
			timeout = 600
		}

		cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

		cmd := exec.CommandContext(cmdCtx, command, cmdArgs...)
		if workDir != "" {
			cmd.Dir = workDir
		}

		// Merge extra env vars with current process environment
		if ej := stringArg(args, "env_json"); ej != "" {
			var extra map[string]string
			if err := json.Unmarshal([]byte(ej), &extra); err != nil {
				return "", fmt.Errorf("invalid env_json: %w", err)
			}
			env := os.Environ()
			for k, v := range extra {
				env = append(env, fmt.Sprintf("%s=%s", k, v))
			}
			cmd.Env = env
		}

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		start := time.Now()
		runErr := cmd.Run()
		duration := time.Since(start)

		exitCode := 0
		errMsg := ""
		if runErr != nil {
			if exitErr, ok := runErr.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				errMsg = runErr.Error()
			}
		}

		// Truncate large outputs
		const maxOut = 50 * 1024
		stdoutStr := stdout.String()
		stderrStr := stderr.String()
		if len(stdoutStr) > maxOut {
			stdoutStr = stdoutStr[:maxOut] + "\n... (output truncated)"
		}
		if len(stderrStr) > maxOut {
			stderrStr = stderrStr[:maxOut] + "\n... (stderr truncated)"
		}

		out, _ := json.Marshal(map[string]interface{}{
			"command":     command,
			"args":        cmdArgs,
			"working_dir": workDir,
			"exit_code":   exitCode,
			"success":     exitCode == 0 && errMsg == "",
			"stdout":      stdoutStr,
			"stderr":      stderrStr,
			"duration_ms": duration.Milliseconds(),
			"error":       errMsg,
		})
		return string(out), nil
	})
}

// ─── shared test runner helper ─────────────────────────────────────────────────

// runTestCommand executes a test command and parses the result into a structured summary.
func (s *AgentToolRegistryService) runTestCommand(ctx context.Context, workDir string, timeoutSec int, binary string, args ...string) (string, error) {
	// Resolve absolute path of binary
	binaryPath, err := exec.LookPath(binary)
	if err != nil {
		return "", fmt.Errorf("binary %q not found in PATH: %w", binary, err)
	}

	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, binaryPath, args...)
	cmd.Dir = workDir

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	start := time.Now()
	runErr := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	stdout := outBuf.String()
	stderr := errBuf.String()

	// Truncate large outputs
	const maxOut = 100 * 1024
	if len(stdout) > maxOut {
		stdout = stdout[:maxOut] + "\n... (output truncated at 100KB)"
	}
	if len(stderr) > maxOut {
		stderr = stderr[:maxOut] + "\n... (stderr truncated at 100KB)"
	}

	// Parse Go test counts from output (PASS/FAIL lines)
	passed, failed, skipped := parseGoTestCounts(stdout + stderr)

	out, _ := json.Marshal(map[string]interface{}{
		"binary":      filepath.Base(binary),
		"args":        args,
		"working_dir": workDir,
		"exit_code":   exitCode,
		"success":     exitCode == 0,
		"passed":      passed,
		"failed":      failed,
		"skipped":     skipped,
		"duration_ms": duration.Milliseconds(),
		"stdout":      stdout,
		"stderr":      stderr,
	})
	return string(out), nil
}

// parseGoTestCounts counts --- PASS, --- FAIL, --- SKIP lines in go test -v output.
func parseGoTestCounts(output string) (passed, failed, skipped int) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "--- PASS") {
			passed++
		} else if strings.HasPrefix(line, "--- FAIL") {
			failed++
		} else if strings.HasPrefix(line, "--- SKIP") {
			skipped++
		}
	}
	return
}
