package funcs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

// EnhancedCSharpExecutor provides enhanced C# code execution with safety and performance features
type EnhancedCSharpExecutor struct {
	Config     *ScriptExecutionConfig
	Log        logger.Log
	DotnetPath string // Path to dotnet executable
}

// NewEnhancedCSharpExecutor creates a new enhanced C# executor
func NewEnhancedCSharpExecutor(config *ScriptExecutionConfig, log logger.Log) *EnhancedCSharpExecutor {
	if config == nil {
		config = DefaultScriptConfig()
	}
	return &EnhancedCSharpExecutor{
		Config:     config,
		Log:        log,
		DotnetPath: "dotnet", // Can be configured
	}
}

// GetType returns the executor type
func (e *EnhancedCSharpExecutor) GetType() string {
	return "CSharp"
}

// Execute executes C# code with timeout and error handling
func (e *EnhancedCSharpExecutor) Execute(
	ctx context.Context,
	script string,
	inputs map[string]interface{},
	outputs []string,
) (map[string]interface{}, error) {

	startTime := time.Now()
	e.Log.Debug(fmt.Sprintf("Executing C# code (length: %d bytes)", len(script)))

	// Validate script
	if err := ValidateScriptSafety(script, "CSharp"); err != nil {
		return nil, err
	}

	// Build command arguments
	cmdArgs := []string{"-c", script}

	// Add input parameters
	for key, value := range inputs {
		// Convert value to string representation
		valueStr := e.convertValueToString(value)
		cmdArgs = append(cmdArgs, fmt.Sprintf("-p:%s=%s", key, valueStr))
	}

	e.Log.Debug(fmt.Sprintf("C# command args: %v", cmdArgs))

	// Create command with context for timeout support
	cmd := exec.CommandContext(ctx, e.DotnetPath, cmdArgs...)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()

	// Get execution time
	executionTime := time.Since(startTime)

	// Check for errors
	if err != nil {
		// Check if it's a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return nil, types.NewTimeoutError(
				"C# execution",
				e.Config.Timeout,
			).WithContext(&types.ExecutionContext{
				FunctionType:  "CSharp",
				ExecutionTime: startTime,
			})
		}

		// Regular execution error
		errorMsg := stderr.String()
		if errorMsg == "" {
			errorMsg = err.Error()
		}

		return nil, types.NewScriptError(
			"C#",
			"C# execution failed",
			err,
		).WithContext(&types.ExecutionContext{
			FunctionType:  "CSharp",
			ExecutionTime: startTime,
		}).WithDetail("stderr", errorMsg).
			WithDetail("exit_code", fmt.Sprintf("%v", cmd.ProcessState.ExitCode()))
	}

	// Parse output
	outputBytes := stdout.Bytes()
	e.Log.Debug(fmt.Sprintf("C# output: %s", string(outputBytes)))

	// If output is empty, return empty map
	if len(outputBytes) == 0 {
		e.Log.Debug("C# execution produced no output")
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
		e.Log.Debug(fmt.Sprintf("Failed to parse C# output as JSON: %s", err.Error()))

		// If we have exactly one output, use the raw output
		if len(outputs) == 1 {
			result := make(map[string]interface{})
			result[outputs[0]] = string(outputBytes)
			return result, nil
		}

		// Otherwise, return error
		return nil, types.NewScriptError(
			"C#",
			"Failed to parse C# output as JSON",
			err,
		).WithDetail("output", string(outputBytes))
	}

	// Extract expected outputs
	result := make(map[string]interface{})
	for _, outputName := range outputs {
		if value, exists := outputMap[outputName]; exists {
			result[outputName] = value
		} else {
			e.Log.Debug(fmt.Sprintf("Expected output '%s' not found in C# result", outputName))
			result[outputName] = nil
		}
	}

	e.Log.Info(fmt.Sprintf("C# code executed in %v", executionTime))

	return result, nil
}

// convertValueToString converts a value to string for command line parameter
func (e *EnhancedCSharpExecutor) convertValueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int64, int32, float32, float64, bool:
		return fmt.Sprintf("%v", v)
	case time.Time:
		return v.Format(types.DateTimeFormat)
	default:
		// For complex types, try to marshal as JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	}
}

// Validate validates C# code syntax
// Note: This is a basic check. Full validation requires dotnet compilation
func (e *EnhancedCSharpExecutor) Validate(script string) error {
	if script == "" {
		return types.NewValidationError("C# code is empty", nil)
	}

	// Could add more sophisticated validation here
	// For now, just check that dotnet is available
	cmd := exec.Command(e.DotnetPath, "--version")
	err := cmd.Run()
	if err != nil {
		return types.NewValidationError(
			"dotnet command not available",
			err,
		).WithDetail("dotnet_path", e.DotnetPath)
	}

	return nil
}

// EnhancedCSharpFuncs is the enhanced version of CSharpFuncs with better error handling
type EnhancedCSharpFuncs struct {
	executor *EnhancedCSharpExecutor
}

// Execute executes the enhanced C# function
func (cf *EnhancedCSharpFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.EnhancedCSharpFuncs.Execute", elapsed)
	}()

	// ROLLBACK DESIGN: Catch panics and create structured errors
	defer func() {
		if err := recover(); err != nil {
			errMsg := fmt.Sprintf("Panic in EnhancedCSharpFuncs: %v", err)
			f.iLog.Error(errMsg)

			// Create structured error
			execContext := &types.ExecutionContext{
				FunctionName:  f.Fobj.Name,
				FunctionType:  "CSharp",
				ExecutionTime: startTime,
			}

			structuredErr := types.NewScriptError("C#", errMsg, nil).
				WithContext(execContext).
				WithRollbackReason("C# code execution failed")

			f.iLog.Error(structuredErr.GetFormattedError())
			f.CancelExecution(errMsg)
			f.ErrorMessage = errMsg

			// Re-panic to trigger rollback
			panic(structuredErr)
		}
	}()

	f.iLog.Info(fmt.Sprintf("Start processing C# code for function %s", f.Fobj.Name))

	// Get inputs using the new comprehensive mapper
	namelist, _, inputs := f.SetInputs()
	f.iLog.Debug(fmt.Sprintf("C# inputs: %v", namelist))

	// Create executor with config
	config := &ScriptExecutionConfig{
		Timeout:        60 * time.Second, // C# may need more time for compilation
		MaxMemoryMB:    256,
		SandboxEnabled: false, // Dotnet has its own sandboxing
	}

	executor := NewEnhancedCSharpExecutor(config, f.iLog)

	// Extract output names
	outputNames := make([]string, len(f.Fobj.Outputs))
	for i, output := range f.Fobj.Outputs {
		outputNames[i] = output.Name
	}

	// Execute with timeout
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
		// Handle error
		scriptErr := ScriptErrorHandler(
			err,
			"CSharp",
			f.Fobj.Content,
			f.Fobj.Name,
			startTime,
			f.iLog,
		)

		f.ErrorMessage = scriptErr.Error()
		f.iLog.Error(scriptErr.Error())

		// Trigger rollback if it's a critical error
		if bpmErr, ok := scriptErr.(*types.BPMError); ok {
			if bpmErr.Severity == types.ErrorSeverityCritical || bpmErr.Severity == types.ErrorSeverityError {
				panic(bpmErr)
			}
		}

		return
	}

	// Set outputs
	if result.Outputs != nil {
		f.SetOutputs(result.Outputs)
		f.iLog.Info(fmt.Sprintf("C# code executed successfully in %v", result.ExecutionTime))
	}
}

// Validate validates the C# code
func (cf *EnhancedCSharpFuncs) Validate(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.EnhancedCSharpFuncs.Validate", elapsed)
	}()

	config := DefaultScriptConfig()
	executor := NewEnhancedCSharpExecutor(config, f.iLog)

	err := executor.Validate(f.Fobj.Content)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Testfunction tests the C# code with provided inputs
func (cf *EnhancedCSharpFuncs) Testfunction(
	content string,
	inputs interface{},
	outputs []string,
) (map[string]interface{}, error) {

	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "EnhancedCSharp Function"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("engine.funcs.EnhancedCSharpFuncs.Testfunction", elapsed)
	}()

	iLog.Debug("Test C# code execution")
	iLog.Debug(fmt.Sprintf("C# code length: %d", len(content)))

	// Convert inputs to map
	inputMap, ok := inputs.(map[string]interface{})
	if !ok {
		return nil, types.NewValidationError("Inputs must be a map", nil).
			WithDetail("input_type", fmt.Sprintf("%T", inputs))
	}

	// Create executor
	config := &ScriptExecutionConfig{
		Timeout:        60 * time.Second,
		MaxMemoryMB:    256,
		SandboxEnabled: false,
	}
	executor := NewEnhancedCSharpExecutor(config, iLog)

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	result, err := executor.Execute(ctx, content, inputMap, outputs)
	if err != nil {
		iLog.Error(fmt.Sprintf("Test execution failed: %s", err.Error()))
		return nil, err
	}

	iLog.Debug(fmt.Sprintf("Test result: %v", result))
	return result, nil
}
