package funcs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

// EnhancedPythonExecutor provides Python code execution with safety and performance features
type EnhancedPythonExecutor struct {
	Config     *ScriptExecutionConfig
	Log        logger.Log
	PythonPath string // Path to python executable (python3)
	IsExpr     bool   // true for expression evaluation, false for full script
}

// NewEnhancedPythonExecutor creates a new enhanced Python executor
func NewEnhancedPythonExecutor(config *ScriptExecutionConfig, isExpr bool, log logger.Log) *EnhancedPythonExecutor {
	if config == nil {
		config = DefaultScriptConfig()
	}
	return &EnhancedPythonExecutor{
		Config:     config,
		Log:        log,
		PythonPath: "python3", // Can be configured
		IsExpr:     isExpr,
	}
}

// GetType returns the executor type
func (e *EnhancedPythonExecutor) GetType() string {
	if e.IsExpr {
		return "PythonExpr"
	}
	return "PythonScript"
}

// Execute executes Python code/expression with timeout and error handling
func (e *EnhancedPythonExecutor) Execute(
	ctx context.Context,
	script string,
	inputs map[string]interface{},
	outputs []string,
) (map[string]interface{}, error) {

	startTime := time.Now()
	scriptType := e.GetType()
	e.Log.Debug(fmt.Sprintf("Executing %s (length: %d bytes)", scriptType, len(script)))

	// Validate script
	if err := ValidateScriptSafety(script, scriptType); err != nil {
		return nil, err
	}

	// Build Python wrapper script
	wrapperScript, err := e.buildWrapperScript(script, inputs, outputs)
	if err != nil {
		return nil, err
	}

	e.Log.Debug(fmt.Sprintf("Python wrapper script:\n%s", wrapperScript))

	// Create command with context for timeout support
	cmd := exec.CommandContext(ctx, e.PythonPath, "-c", wrapperScript)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err = cmd.Run()

	// Get execution time
	executionTime := time.Since(startTime)

	// Check for errors
	if err != nil {
		// Check if it's a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return nil, types.NewTimeoutError(
				fmt.Sprintf("%s execution", scriptType),
				e.Config.Timeout,
			).WithContext(&types.ExecutionContext{
				FunctionType:  scriptType,
				ExecutionTime: startTime,
			})
		}

		// Regular execution error
		errorMsg := stderr.String()
		if errorMsg == "" {
			errorMsg = err.Error()
		}

		return nil, types.NewScriptError(
			scriptType,
			fmt.Sprintf("%s failed", scriptType),
			err,
		).WithContext(&types.ExecutionContext{
			FunctionType:  scriptType,
			ExecutionTime: startTime,
		}).WithDetail("stderr", errorMsg).
			WithDetail("exit_code", fmt.Sprintf("%v", cmd.ProcessState.ExitCode()))
	}

	// Parse output
	outputBytes := stdout.Bytes()
	e.Log.Debug(fmt.Sprintf("Python output: %s", string(outputBytes)))

	// Check stderr for warnings
	if stderr.Len() > 0 {
		e.Log.Debug(fmt.Sprintf("Python warnings: %s", stderr.String()))
	}

	// If output is empty, return empty map
	if len(outputBytes) == 0 {
		e.Log.Debug("Python execution produced no output")
		result := make(map[string]interface{})
		for _, outputName := range outputs {
			result[outputName] = nil
		}
		return result, nil
	}

	// Try to parse as JSON
	var outputMap map[string]interface{}
	err = json.Unmarshal(outputBytes, &outputMap)
	if err != nil {
		// If JSON parsing fails, treat entire output as string
		e.Log.Debug(fmt.Sprintf("Failed to parse Python output as JSON: %s", err.Error()))

		// If we have exactly one output, use the raw output
		if len(outputs) == 1 {
			result := make(map[string]interface{})
			result[outputs[0]] = strings.TrimSpace(string(outputBytes))
			return result, nil
		}

		// Otherwise, return error
		return nil, types.NewScriptError(
			scriptType,
			"Failed to parse Python output as JSON",
			err,
		).WithDetail("output", string(outputBytes))
	}

	// Extract expected outputs
	result := make(map[string]interface{})
	for _, outputName := range outputs {
		if value, exists := outputMap[outputName]; exists {
			result[outputName] = value
		} else {
			e.Log.Debug(fmt.Sprintf("Expected output '%s' not found in Python result", outputName))
			result[outputName] = nil
		}
	}

	e.Log.Info(fmt.Sprintf("%s executed in %v", scriptType, executionTime))

	return result, nil
}

// buildWrapperScript builds a Python wrapper script that handles I/O
func (e *EnhancedPythonExecutor) buildWrapperScript(
	userScript string,
	inputs map[string]interface{},
	outputs []string,
) (string, error) {

	var script strings.Builder

	// Import required modules
	script.WriteString("import json\n")
	script.WriteString("import sys\n")
	script.WriteString("from datetime import datetime\n\n")

	// Define input variables
	script.WriteString("# Input variables\n")
	for key, value := range inputs {
		// Convert Go value to Python representation
		pyValue, err := e.convertToPythonValue(value)
		if err != nil {
			return "", types.NewScriptError(
				"Python",
				fmt.Sprintf("Failed to convert input '%s' to Python value", key),
				err,
			)
		}
		script.WriteString(fmt.Sprintf("%s = %s\n", key, pyValue))
	}
	script.WriteString("\n")

	// Add user script
	if e.IsExpr {
		// For expressions, we need to evaluate and capture result
		script.WriteString("# Evaluate expression\n")
		script.WriteString("try:\n")
		script.WriteString(fmt.Sprintf("    __result__ = %s\n", userScript))
		script.WriteString("except Exception as e:\n")
		script.WriteString("    print(json.dumps({'error': str(e)}), file=sys.stderr)\n")
		script.WriteString("    sys.exit(1)\n\n")

		// For single expression, create output map
		script.WriteString("# Build output\n")
		script.WriteString("__outputs__ = {}\n")
		if len(outputs) == 1 {
			// If single output, use expression result
			script.WriteString(fmt.Sprintf("__outputs__['%s'] = __result__\n", outputs[0]))
		} else {
			// If multiple outputs, result should be a dict
			script.WriteString("if isinstance(__result__, dict):\n")
			script.WriteString("    __outputs__ = __result__\n")
			script.WriteString("else:\n")
			for _, outputName := range outputs {
				script.WriteString(fmt.Sprintf("    __outputs__['%s'] = globals().get('%s')\n", outputName, outputName))
			}
		}
	} else {
		// For full scripts, execute and extract variables
		script.WriteString("# Execute user script\n")
		script.WriteString("try:\n")
		// Indent user script
		for _, line := range strings.Split(userScript, "\n") {
			if strings.TrimSpace(line) != "" {
				script.WriteString("    " + line + "\n")
			}
		}
		script.WriteString("except Exception as e:\n")
		script.WriteString("    print(json.dumps({'error': str(e)}), file=sys.stderr)\n")
		script.WriteString("    sys.exit(1)\n\n")

		// Extract output variables
		script.WriteString("# Extract outputs\n")
		script.WriteString("__outputs__ = {}\n")
		for _, outputName := range outputs {
			script.WriteString(fmt.Sprintf("if '%s' in globals():\n", outputName))
			script.WriteString(fmt.Sprintf("    __outputs__['%s'] = %s\n", outputName, outputName))
		}
	}

	// Output results as JSON
	script.WriteString("\n# Output results as JSON\n")
	script.WriteString("def json_serializer(obj):\n")
	script.WriteString("    if isinstance(obj, datetime):\n")
	script.WriteString("        return obj.isoformat()\n")
	script.WriteString("    raise TypeError(f'Object of type {type(obj)} is not JSON serializable')\n\n")
	script.WriteString("print(json.dumps(__outputs__, default=json_serializer))\n")

	return script.String(), nil
}

// convertToPythonValue converts a Go value to Python representation
func (e *EnhancedPythonExecutor) convertToPythonValue(value interface{}) (string, error) {
	switch v := value.(type) {
	case nil:
		return "None", nil
	case string:
		// Escape string for Python
		escaped := strings.ReplaceAll(v, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "'", "\\'")
		escaped = strings.ReplaceAll(escaped, "\n", "\\n")
		return fmt.Sprintf("'%s'", escaped), nil
	case int, int64, int32:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%f", v), nil
	case bool:
		if v {
			return "True", nil
		}
		return "False", nil
	case time.Time:
		return fmt.Sprintf("datetime.fromisoformat('%s')", v.Format(time.RFC3339)), nil
	case []interface{}, map[string]interface{}:
		// For complex types, use JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(jsonBytes), nil
	default:
		// Try JSON marshaling
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(jsonBytes), nil
	}
}

// Validate validates Python code syntax
func (e *EnhancedPythonExecutor) Validate(script string) error {
	if script == "" {
		return types.NewValidationError("Python code is empty", nil)
	}

	// Check that python is available
	cmd := exec.Command(e.PythonPath, "--version")
	err := cmd.Run()
	if err != nil {
		return types.NewValidationError(
			"python3 command not available",
			err,
		).WithDetail("python_path", e.PythonPath)
	}

	// Try to compile the script
	compileScript := fmt.Sprintf("import py_compile; py_compile.compile('-', doraise=True)")
	cmd = exec.Command(e.PythonPath, "-c", compileScript)

	var stdin bytes.Buffer
	stdin.WriteString(script)
	cmd.Stdin = &stdin

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return types.NewValidationError(
			"Invalid Python syntax",
			err,
		).WithDetail("compile_error", stderr.String())
	}

	return nil
}

// PythonExprFuncs handles Python expression evaluation
type PythonExprFuncs struct {
	executor *EnhancedPythonExecutor
}

// Execute executes the Python expression
func (cf *PythonExprFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.PythonExprFuncs.Execute", elapsed)
	}()

	// ROLLBACK DESIGN: Catch panics and create structured errors
	defer func() {
		if err := recover(); err != nil {
			errMsg := fmt.Sprintf("Panic in PythonExprFuncs: %v", err)
			f.iLog.Error(errMsg)

			execContext := &types.ExecutionContext{
				FunctionName:  f.Fobj.Name,
				FunctionType:  "PythonExpr",
				ExecutionTime: startTime,
			}

			structuredErr := types.NewScriptError("PythonExpr", errMsg, nil).
				WithContext(execContext).
				WithRollbackReason("Python expression execution failed")

			f.iLog.Error(structuredErr.GetFormattedError())
			f.CancelExecution(errMsg)
			f.ErrorMessage = errMsg
			panic(structuredErr)
		}
	}()

	f.iLog.Info(fmt.Sprintf("Start processing Python expression for function %s", f.Fobj.Name))

	namelist, _, inputs := f.SetInputs()
	f.iLog.Debug(fmt.Sprintf("Python expression inputs: %v", namelist))

	config := &ScriptExecutionConfig{
		Timeout:        30 * time.Second,
		MaxMemoryMB:    128,
		SandboxEnabled: true,
	}

	executor := NewEnhancedPythonExecutor(config, true, f.iLog)

	outputNames := make([]string, len(f.Fobj.Outputs))
	for i, output := range f.Fobj.Outputs {
		outputNames[i] = output.Name
	}

	result, err := ExecuteScriptWithTimeout(
		executor,
		f.Ctx,
		f.Fobj.Content,
		inputs,
		outputNames,
		config.Timeout,
		f.iLog,
	)

	if err != nil {
		scriptErr := ScriptErrorHandler(err, "PythonExpr", f.Fobj.Content, f.Fobj.Name, startTime, f.iLog)
		f.ErrorMessage = scriptErr.Error()
		f.iLog.Error(scriptErr.Error())

		if bpmErr, ok := scriptErr.(*types.BPMError); ok {
			if bpmErr.Severity == types.SeverityCritical || bpmErr.Severity == types.SeverityError {
				panic(bpmErr)
			}
		}
		return
	}

	if result.Outputs != nil {
		f.SetOutputs(result.Outputs)
		f.iLog.Info(fmt.Sprintf("Python expression executed successfully in %v", result.ExecutionTime))
	}
}

// Validate validates the Python expression
func (cf *PythonExprFuncs) Validate(f *Funcs) (bool, error) {
	config := DefaultScriptConfig()
	executor := NewEnhancedPythonExecutor(config, true, f.iLog)
	err := executor.Validate(f.Fobj.Content)
	return err == nil, err
}

// Testfunction tests the Python expression
func (cf *PythonExprFuncs) Testfunction(content string, inputs interface{}, outputs []string) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PythonExpr Function"}

	inputMap, ok := inputs.(map[string]interface{})
	if !ok {
		return nil, types.NewValidationError("Inputs must be a map", nil)
	}

	config := DefaultScriptConfig()
	executor := NewEnhancedPythonExecutor(config, true, iLog)

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	return executor.Execute(ctx, content, inputMap, outputs)
}

// PythonScriptFuncs handles Python script execution
type PythonScriptFuncs struct {
	executor *EnhancedPythonExecutor
}

// Execute executes the Python script
func (cf *PythonScriptFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.PythonScriptFuncs.Execute", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			errMsg := fmt.Sprintf("Panic in PythonScriptFuncs: %v", err)
			f.iLog.Error(errMsg)

			execContext := &types.ExecutionContext{
				FunctionName:  f.Fobj.Name,
				FunctionType:  "PythonScript",
				ExecutionTime: startTime,
			}

			structuredErr := types.NewScriptError("PythonScript", errMsg, nil).
				WithContext(execContext).
				WithRollbackReason("Python script execution failed")

			f.iLog.Error(structuredErr.GetFormattedError())
			f.CancelExecution(errMsg)
			f.ErrorMessage = errMsg
			panic(structuredErr)
		}
	}()

	f.iLog.Info(fmt.Sprintf("Start processing Python script for function %s", f.Fobj.Name))

	namelist, _, inputs := f.SetInputs()
	f.iLog.Debug(fmt.Sprintf("Python script inputs: %v", namelist))

	config := &ScriptExecutionConfig{
		Timeout:        60 * time.Second,
		MaxMemoryMB:    256,
		SandboxEnabled: true,
	}

	executor := NewEnhancedPythonExecutor(config, false, f.iLog)

	outputNames := make([]string, len(f.Fobj.Outputs))
	for i, output := range f.Fobj.Outputs {
		outputNames[i] = output.Name
	}

	result, err := ExecuteScriptWithTimeout(
		executor,
		f.Ctx,
		f.Fobj.Content,
		inputs,
		outputNames,
		config.Timeout,
		f.iLog,
	)

	if err != nil {
		scriptErr := ScriptErrorHandler(err, "PythonScript", f.Fobj.Content, f.Fobj.Name, startTime, f.iLog)
		f.ErrorMessage = scriptErr.Error()
		f.iLog.Error(scriptErr.Error())

		if bpmErr, ok := scriptErr.(*types.BPMError); ok {
			if bpmErr.Severity == types.SeverityCritical || bpmErr.Severity == types.SeverityError {
				panic(bpmErr)
			}
		}
		return
	}

	if result.Outputs != nil {
		f.SetOutputs(result.Outputs)
		f.iLog.Info(fmt.Sprintf("Python script executed successfully in %v", result.ExecutionTime))
	}
}

// Validate validates the Python script
func (cf *PythonScriptFuncs) Validate(f *Funcs) (bool, error) {
	config := DefaultScriptConfig()
	executor := NewEnhancedPythonExecutor(config, false, f.iLog)
	err := executor.Validate(f.Fobj.Content)
	return err == nil, err
}

// Testfunction tests the Python script
func (cf *PythonScriptFuncs) Testfunction(content string, inputs interface{}, outputs []string) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PythonScript Function"}

	inputMap, ok := inputs.(map[string]interface{})
	if !ok {
		return nil, types.NewValidationError("Inputs must be a map", nil)
	}

	config := &ScriptExecutionConfig{
		Timeout:        60 * time.Second,
		MaxMemoryMB:    256,
		SandboxEnabled: true,
	}
	executor := NewEnhancedPythonExecutor(config, false, iLog)

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	return executor.Execute(ctx, content, inputMap, outputs)
}
