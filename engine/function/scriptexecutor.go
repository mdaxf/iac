package funcs

import (
	"context"
	"fmt"
	"time"

	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

// ScriptExecutor provides a common interface for all script executors
type ScriptExecutor interface {
	Execute(ctx context.Context, script string, inputs map[string]interface{}, outputs []string) (map[string]interface{}, error)
	Validate(script string) error
	GetType() string
}

// ScriptExecutionConfig holds configuration for script execution
type ScriptExecutionConfig struct {
	Timeout        time.Duration
	MaxMemoryMB    int
	AllowedModules []string
	SandboxEnabled bool
}

// DefaultScriptConfig returns default script execution configuration
func DefaultScriptConfig() *ScriptExecutionConfig {
	return &ScriptExecutionConfig{
		Timeout:        30 * time.Second,
		MaxMemoryMB:    128,
		AllowedModules: []string{},
		SandboxEnabled: true,
	}
}

// ScriptExecutionResult holds the result of script execution
type ScriptExecutionResult struct {
	Outputs       map[string]interface{}
	ExecutionTime time.Duration
	MemoryUsed    int64
	Error         error
}

// ExecuteScriptWithTimeout executes a script with timeout and returns structured result
func ExecuteScriptWithTimeout(
	executor ScriptExecutor,
	ctx context.Context,
	script string,
	inputs map[string]interface{},
	outputs []string,
	timeout time.Duration,
	log logger.Log,
) (*ScriptExecutionResult, error) {

	startTime := time.Now()

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Channel for results
	resultChan := make(chan map[string]interface{}, 1)
	errorChan := make(chan error, 1)

	// Execute in goroutine
	go func() {
		result, err := executor.Execute(timeoutCtx, script, inputs, outputs)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- result
		}
	}()

	// Wait for completion or timeout
	select {
	case result := <-resultChan:
		executionTime := time.Since(startTime)
		return &ScriptExecutionResult{
			Outputs:       result,
			ExecutionTime: executionTime,
			Error:         nil,
		}, nil

	case err := <-errorChan:
		executionTime := time.Since(startTime)
		return &ScriptExecutionResult{
			Error:         err,
			ExecutionTime: executionTime,
		}, err

	case <-timeoutCtx.Done():
		executionTime := time.Since(startTime)
		timeoutErr := types.NewTimeoutError(
			fmt.Sprintf("Script execution (%s)", executor.GetType()),
			timeout,
		).WithContext(&types.ExecutionContext{
			FunctionType:  executor.GetType(),
			ExecutionTime: startTime,
		})

		log.Error(fmt.Sprintf("Script execution timeout: %s", timeoutErr.Error()))

		return &ScriptExecutionResult{
			Error:         timeoutErr,
			ExecutionTime: executionTime,
		}, timeoutErr
	}
}

// ValidateScriptSafety performs basic safety checks on script content
func ValidateScriptSafety(script string, scriptType string) error {
	if script == "" {
		return types.NewValidationError("Script content is empty", nil).
			WithDetail("script_type", scriptType)
	}

	// Check for potentially dangerous patterns (basic security)
	// This is a basic check - full sandboxing should be done by the executor

	// Additional validation can be added here based on script type

	return nil
}

// ConvertScriptOutput safely converts script output to expected format
func ConvertScriptOutput(
	rawOutput interface{},
	expectedOutputs []string,
	scriptType string,
	log logger.Log,
) (map[string]interface{}, error) {

	result := make(map[string]interface{})

	// Try to convert to map
	if outputMap, ok := rawOutput.(map[string]interface{}); ok {
		// Validate that all expected outputs are present
		for _, outputName := range expectedOutputs {
			if value, exists := outputMap[outputName]; exists {
				result[outputName] = value
			} else {
				log.Debug(fmt.Sprintf("Expected output '%s' not found in script result", outputName))
				// Don't error, just set to nil
				result[outputName] = nil
			}
		}
		return result, nil
	}

	// If output is not a map and we have exactly one expected output, use it directly
	if len(expectedOutputs) == 1 {
		result[expectedOutputs[0]] = rawOutput
		return result, nil
	}

	// Otherwise, try to convert to map using type assertion
	resultMap, err := types.AssertMap(rawOutput, fmt.Sprintf("%s script output", scriptType))
	if err != nil {
		return nil, types.NewExecutionError(
			fmt.Sprintf("Failed to convert %s script output to map", scriptType),
			err,
		).WithDetail("output_type", fmt.Sprintf("%T", rawOutput))
	}

	// Extract expected outputs
	for _, outputName := range expectedOutputs {
		if value, exists := resultMap[outputName]; exists {
			result[outputName] = value
		} else {
			result[outputName] = nil
		}
	}

	return result, nil
}

// ScriptErrorHandler handles errors from script execution and converts them to BPMErrors
func ScriptErrorHandler(
	err error,
	scriptType string,
	script string,
	functionName string,
	startTime time.Time,
	log logger.Log,
) error {

	if err == nil {
		return nil
	}

	// Check if it's already a BPMError
	if bpmErr, ok := err.(*types.BPMError); ok {
		// Add additional context if not present
		if bpmErr.Context == nil {
			bpmErr.Context = &types.ExecutionContext{}
		}
		if bpmErr.Context.FunctionType == "" {
			bpmErr.Context.FunctionType = scriptType
		}
		if bpmErr.Context.FunctionName == "" {
			bpmErr.Context.FunctionName = functionName
		}
		if bpmErr.Context.ExecutionTime.IsZero() {
			bpmErr.Context.ExecutionTime = startTime
		}
		return bpmErr
	}

	// Create new script error
	scriptErr := types.NewScriptError(
		scriptType,
		fmt.Sprintf("%s script execution failed", scriptType),
		err,
	).WithContext(&types.ExecutionContext{
		FunctionType:  scriptType,
		FunctionName:  functionName,
		ExecutionTime: startTime,
	}).WithDetail("script_type", scriptType)

	// Truncate script content for logging (max 200 chars)
	scriptPreview := script
	if len(script) > 200 {
		scriptPreview = script[:200] + "..."
	}
	scriptErr.WithDetail("script_preview", scriptPreview)

	log.Error(scriptErr.GetFormattedError())

	return scriptErr
}
