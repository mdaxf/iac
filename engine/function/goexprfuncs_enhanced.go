package funcs

import (
	"context"
	"fmt"
	"time"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

// EnhancedGoExprExecutor provides enhanced Go expression execution with safety and performance features
type EnhancedGoExprExecutor struct {
	Config *ScriptExecutionConfig
	Log    logger.Log
}

// NewEnhancedGoExprExecutor creates a new enhanced Go expression executor
func NewEnhancedGoExprExecutor(config *ScriptExecutionConfig, log logger.Log) *EnhancedGoExprExecutor {
	if config == nil {
		config = DefaultScriptConfig()
	}
	return &EnhancedGoExprExecutor{
		Config: config,
		Log:    log,
	}
}

// GetType returns the executor type
func (e *EnhancedGoExprExecutor) GetType() string {
	return "GoExpression"
}

// Execute executes a Go expression with timeout and error handling
func (e *EnhancedGoExprExecutor) Execute(
	ctx context.Context,
	script string,
	inputs map[string]interface{},
	outputs []string,
) (map[string]interface{}, error) {

	startTime := time.Now()
	e.Log.Debug(fmt.Sprintf("Executing Go expression: %s", script))

	// Validate script
	if err := ValidateScriptSafety(script, "GoExpression"); err != nil {
		return nil, err
	}

	// Prepare environment
	env := make(map[string]interface{})
	for k, v := range inputs {
		env[k] = v
	}

	e.Log.Debug(fmt.Sprintf("Expression environment: %v", env))

	// Compile expression
	program, err := expr.Compile(script, expr.Env(env))
	if err != nil {
		return nil, types.NewScriptError(
			"Failed to compile Go expression",
			err,
		).WithContext(&types.ExecutionContext{
			FunctionType:  "GoExpression",
			ExecutionTime: startTime,
		}).WithDetail("compilation_error", err.Error())
	}

	// Execute with context support
	result, err := e.executeWithContext(ctx, program, env)
	if err != nil {
		return nil, err
	}

	// Convert output
	outputMap, err := ConvertScriptOutput(result, outputs, "GoExpression", e.Log)
	if err != nil {
		return nil, err
	}

	executionTime := time.Since(startTime)
	e.Log.Performance(fmt.Sprintf("Go expression executed in %v", executionTime))

	return outputMap, nil
}

// executeWithContext executes the compiled program with context support
func (e *EnhancedGoExprExecutor) executeWithContext(
	ctx context.Context,
	program *vm.Program,
	env map[string]interface{},
) (interface{}, error) {

	// Channel for results
	resultChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)

	// Execute in goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorChan <- fmt.Errorf("panic during expression execution: %v", r)
			}
		}()

		result, err := expr.Run(program, env)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- result
		}
	}()

	// Wait for completion or context cancellation
	select {
	case result := <-resultChan:
		return result, nil

	case err := <-errorChan:
		return nil, types.NewScriptError(
			"Go expression execution failed",
			err,
		).WithDetail("error_type", "execution_error")

	case <-ctx.Done():
		return nil, types.NewTimeoutError(
			"Go expression execution cancelled or timed out",
			ctx.Err(),
		)
	}
}

// Validate validates a Go expression syntax
func (e *EnhancedGoExprExecutor) Validate(script string) error {
	if script == "" {
		return types.NewValidationError("Go expression is empty", nil)
	}

	// Try to compile with empty environment
	_, err := expr.Compile(script, expr.Env(map[string]interface{}{}))
	if err != nil {
		return types.NewValidationError(
			"Invalid Go expression syntax",
			err,
		).WithDetail("compilation_error", err.Error())
	}

	return nil
}

// EnhancedGoExprFuncs is the enhanced version of GoExprFuncs with better error handling
type EnhancedGoExprFuncs struct {
	executor *EnhancedGoExprExecutor
}

// Execute executes the enhanced Go expression function
func (cf *EnhancedGoExprFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.EnhancedGoExprFuncs.Execute", elapsed)
	}()

	// ROLLBACK DESIGN: Catch panics and create structured errors
	defer func() {
		if err := recover(); err != nil {
			errMsg := fmt.Sprintf("Panic in EnhancedGoExprFuncs: %v", err)
			f.iLog.Error(errMsg)

			// Create structured error
			execContext := &types.ExecutionContext{
				FunctionName:  f.Fobj.Name,
				FunctionType:  "GoExpression",
				ExecutionTime: startTime,
			}

			structuredErr := types.NewScriptError(errMsg, nil).
				WithContext(execContext).
				WithRollbackReason("Go expression execution failed")

			f.iLog.Error(structuredErr.GetFormattedError())
			f.CancelExecution(errMsg)
			f.ErrorMessage = errMsg

			// Re-panic to trigger rollback
			panic(structuredErr)
		}
	}()

	f.iLog.Info(fmt.Sprintf("Start processing Go expression for function %s", f.Fobj.Name))

	// Get inputs using the new comprehensive mapper
	namelist, _, inputs := f.SetInputs()
	f.iLog.Debug(fmt.Sprintf("Expression inputs: %v", namelist))

	// Create executor with config
	config := &ScriptExecutionConfig{
		Timeout:        30 * time.Second,
		MaxMemoryMB:    128,
		SandboxEnabled: true,
	}

	executor := NewEnhancedGoExprExecutor(config, f.iLog)

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
			"GoExpression",
			f.Fobj.Content,
			f.Fobj.Name,
			startTime,
			f.iLog,
		)

		f.ErrorMessage = scriptErr.Error()
		f.iLog.Error(scriptErr.Error())

		// Trigger rollback if it's a critical error
		if bpmErr, ok := scriptErr.(*types.BPMError); ok {
			if bpmErr.Severity == types.SeverityCritical || bpmErr.Severity == types.SeverityError {
				panic(bpmErr)
			}
		}

		return
	}

	// Set outputs
	if result.Outputs != nil {
		f.SetOutputs(result.Outputs)
		f.iLog.Info(fmt.Sprintf("Go expression executed successfully in %v", result.ExecutionTime))
	}
}

// Validate validates the Go expression
func (cf *EnhancedGoExprFuncs) Validate(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.EnhancedGoExprFuncs.Validate", elapsed)
	}()

	config := DefaultScriptConfig()
	executor := NewEnhancedGoExprExecutor(config, f.iLog)

	err := executor.Validate(f.Fobj.Content)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Testfunction tests the Go expression with provided inputs
func (cf *EnhancedGoExprFuncs) Testfunction(
	content string,
	inputs interface{},
	outputs []string,
) (map[string]interface{}, error) {

	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "EnhancedGoExpr Function"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("engine.funcs.EnhancedGoExprFuncs.Testfunction", elapsed)
	}()

	iLog.Debug("Test Go expression execution")
	iLog.Debug(fmt.Sprintf("Expression: %s", content))

	// Convert inputs to map
	inputMap, ok := inputs.(map[string]interface{})
	if !ok {
		return nil, types.NewValidationError("Inputs must be a map", nil).
			WithDetail("input_type", fmt.Sprintf("%T", inputs))
	}

	// Create executor
	config := DefaultScriptConfig()
	executor := NewEnhancedGoExprExecutor(config, iLog)

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
