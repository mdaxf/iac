package funcs

import (
	"fmt"
	"time"

	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

// ExecutionContext provides common context for function execution
type ExecutionContext struct {
	FunctionName string
	FunctionType string
	StartTime    time.Time
	Log          logger.Log
}

// ExecuteWithRecovery executes a function with standard panic recovery and logging
// This eliminates the duplication of defer/recover patterns across all Execute functions
func ExecuteWithRecovery(
	ctx *ExecutionContext,
	executeFunc func() error,
	onError func(error),
) {
	// Performance tracking
	defer func() {
		elapsed := time.Since(ctx.StartTime)
		ctx.Log.PerformanceWithDuration(
			fmt.Sprintf("engine.funcs.%s.Execute", ctx.FunctionType),
			elapsed,
		)
	}()

	// ROLLBACK DESIGN: Standard panic recovery with structured errors
	defer func() {
		if err := recover(); err != nil {
			errMsg := fmt.Sprintf("Panic in %s: %v", ctx.FunctionType, err)
			ctx.Log.Error(errMsg)

			// Create or enhance structured error
			var structuredErr *types.BPMError

			if bpmErr, ok := err.(*types.BPMError); ok {
				// Already a BPMError, enhance context if needed
				structuredErr = bpmErr
				if structuredErr.Context == nil {
					structuredErr.Context = &types.ExecutionContext{}
				}
			} else {
				// Create new BPMError
				execContext := &types.ExecutionContext{
					FunctionName:  ctx.FunctionName,
					FunctionType:  ctx.FunctionType,
					ExecutionTime: ctx.StartTime,
				}

				structuredErr = types.NewExecutionError(errMsg, nil).
					WithContext(execContext).
					WithRollbackReason(fmt.Sprintf("%s execution failed", ctx.FunctionType))
			}

			// Ensure context is complete
			if structuredErr.Context.FunctionName == "" {
				structuredErr.Context.FunctionName = ctx.FunctionName
			}
			if structuredErr.Context.FunctionType == "" {
				structuredErr.Context.FunctionType = ctx.FunctionType
			}
			if structuredErr.Context.ExecutionTime.IsZero() {
				structuredErr.Context.ExecutionTime = ctx.StartTime
			}

			ctx.Log.Error(structuredErr.GetFormattedError())

			// Call error handler
			if onError != nil {
				onError(structuredErr)
			}

			// Re-panic with structured error to trigger rollback
			panic(structuredErr)
		}
	}()

	// Execute the function
	ctx.Log.Info(fmt.Sprintf("Start processing %s for function %s", ctx.FunctionType, ctx.FunctionName))

	err := executeFunc()
	if err != nil {
		// Handle non-panic errors
		ctx.Log.Error(fmt.Sprintf("Error in %s: %s", ctx.FunctionType, err.Error()))

		// Create structured error if needed
		if bpmErr, ok := err.(*types.BPMError); ok {
			if bpmErr.Severity == types.SeverityCritical || bpmErr.Severity == types.SeverityError {
				panic(bpmErr)
			}
		} else {
			// Convert to BPMError and panic if critical
			structuredErr := types.NewExecutionError(err.Error(), err).
				WithContext(&types.ExecutionContext{
					FunctionName:  ctx.FunctionName,
					FunctionType:  ctx.FunctionType,
					ExecutionTime: ctx.StartTime,
				})

			panic(structuredErr)
		}
	}

	ctx.Log.Info(fmt.Sprintf("Completed processing %s for function %s", ctx.FunctionType, ctx.FunctionName))
}

// ValidateWithRecovery executes validation with standard panic recovery
func ValidateWithRecovery(
	functionType string,
	content string,
	log logger.Log,
	validateFunc func() error,
) (bool, error) {
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration(
			fmt.Sprintf("engine.funcs.%s.Validate", functionType),
			elapsed,
		)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("Panic in %s validation: %v", functionType, err))
		}
	}()

	err := validateFunc()
	if err != nil {
		return false, err
	}

	return true, nil
}

// TestFunctionWithRecovery executes test function with standard panic recovery
func TestFunctionWithRecovery(
	functionType string,
	content string,
	inputs interface{},
	outputs []string,
	log logger.Log,
	testFunc func() (map[string]interface{}, error),
) (map[string]interface{}, error) {
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration(
			fmt.Sprintf("engine.funcs.%s.Testfunction", functionType),
			elapsed,
		)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("Panic in %s test: %v", functionType, err))
		}
	}()

	log.Debug(fmt.Sprintf("Test %s execution", functionType))
	log.Debug(fmt.Sprintf("Content length: %d", len(content)))

	result, err := testFunc()
	if err != nil {
		log.Error(fmt.Sprintf("Test execution failed: %s", err.Error()))
		return nil, err
	}

	log.Debug(fmt.Sprintf("Test result: %v", result))
	return result, nil
}

// BaseExecutor provides common functionality for all function executors
type BaseExecutor struct {
	FunctionType string
	Log          logger.Log
}

// NewBaseExecutor creates a new BaseExecutor
func NewBaseExecutor(functionType string, log logger.Log) *BaseExecutor {
	return &BaseExecutor{
		FunctionType: functionType,
		Log:          log,
	}
}

// ExecuteFunction is a template method for function execution
func (be *BaseExecutor) ExecuteFunction(
	f *Funcs,
	executionLogic func(inputs map[string]interface{}) (map[string]interface{}, error),
) {
	execContext := &ExecutionContext{
		FunctionName: f.Fobj.Name,
		FunctionType: be.FunctionType,
		StartTime:    time.Now(),
		Log:          f.iLog,
	}

	ExecuteWithRecovery(execContext, func() error {
		// Get inputs
		_, _, inputs := f.SetInputs()
		f.iLog.Debug(fmt.Sprintf("%s inputs: %v", be.FunctionType, inputs))

		// Execute the specific logic
		outputs, err := executionLogic(inputs)
		if err != nil {
			f.ErrorMessage = err.Error()
			return err
		}

		// Set outputs
		if outputs != nil {
			f.SetOutputs(outputs)
			f.iLog.Info(fmt.Sprintf("%s executed successfully", be.FunctionType))
		}

		return nil
	}, func(err error) {
		f.ErrorMessage = err.Error()
		f.CancelExecution(err.Error())
	})
}

// OutputProcessor handles common output processing logic
type OutputProcessor struct {
	Log logger.Log
}

// NewOutputProcessor creates a new OutputProcessor
func NewOutputProcessor(log logger.Log) *OutputProcessor {
	return &OutputProcessor{Log: log}
}

// ProcessOutputs extracts outputs from a function based on output definitions
func (op *OutputProcessor) ProcessOutputs(
	rawOutputs interface{},
	outputDefs []types.Output,
) (map[string]interface{}, error) {

	result := make(map[string]interface{})

	// Try to convert to map
	if outputMap, ok := rawOutputs.(map[string]interface{}); ok {
		for _, outputDef := range outputDefs {
			if value, exists := outputMap[outputDef.Name]; exists {
				// Convert value to appropriate type based on output definition
				convertedValue, err := op.convertOutputValue(value, outputDef.Datatype)
				if err != nil {
					op.Log.Debug(fmt.Sprintf("Warning: Failed to convert output '%s': %s", outputDef.Name, err.Error()))
					result[outputDef.Name] = value // Use raw value
				} else {
					result[outputDef.Name] = convertedValue
				}
			} else {
				op.Log.Debug(fmt.Sprintf("Output '%s' not found in results", outputDef.Name))
				result[outputDef.Name] = nil
			}
		}
		return result, nil
	}

	// If single value and single output, use it directly
	if len(outputDefs) == 1 {
		convertedValue, err := op.convertOutputValue(rawOutputs, outputDefs[0].Datatype)
		if err != nil {
			result[outputDefs[0].Name] = rawOutputs
		} else {
			result[outputDefs[0].Name] = convertedValue
		}
		return result, nil
	}

	return nil, fmt.Errorf("cannot process outputs: expected map or single value")
}

// convertOutputValue converts a value to the appropriate type
func (op *OutputProcessor) convertOutputValue(value interface{}, dataType types.DataType) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch dataType {
	case types.Integer:
		return TypeConv.ConvertToInt(value)
	case types.Float:
		return TypeConv.ConvertToFloat(value)
	case types.Bool:
		return TypeConv.ConvertToBool(value)
	case types.DateTime:
		return TypeConv.ConvertToDateTime(value)
	case types.String:
		return TypeConv.ConvertToString(value), nil
	case types.Object:
		// Keep as-is for objects
		return value, nil
	default:
		return TypeConv.ConvertToString(value), nil
	}
}

// InputProcessor handles common input processing logic
type InputProcessor struct {
	Log logger.Log
}

// NewInputProcessor creates a new InputProcessor
func NewInputProcessor(log logger.Log) *InputProcessor {
	return &InputProcessor{Log: log}
}

// ValidateInputs validates that all required inputs are present
func (ip *InputProcessor) ValidateInputs(
	inputs map[string]interface{},
	requiredInputs []string,
) error {
	missing := make([]string, 0)

	for _, required := range requiredInputs {
		if value, exists := inputs[required]; !exists || value == nil {
			missing = append(missing, required)
		}
	}

	if len(missing) > 0 {
		return types.NewValidationError(
			fmt.Sprintf("Missing required inputs: %v", missing),
			nil,
		)
	}

	return nil
}

// GetInputWithDefault gets an input value with a default
func (ip *InputProcessor) GetInputWithDefault(
	inputs map[string]interface{},
	key string,
	defaultValue interface{},
) interface{} {
	if value, exists := inputs[key]; exists && value != nil {
		return value
	}
	return defaultValue
}

// GetStringInput safely gets a string input
func (ip *InputProcessor) GetStringInput(
	inputs map[string]interface{},
	key string,
	defaultValue string,
) string {
	value := ip.GetInputWithDefault(inputs, key, defaultValue)
	return TypeConv.ConvertToString(value)
}

// GetIntInput safely gets an int input
func (ip *InputProcessor) GetIntInput(
	inputs map[string]interface{},
	key string,
	defaultValue int,
) (int, error) {
	value := ip.GetInputWithDefault(inputs, key, defaultValue)
	return TypeConv.ConvertToInt(value)
}

// GetBoolInput safely gets a bool input
func (ip *InputProcessor) GetBoolInput(
	inputs map[string]interface{},
	key string,
	defaultValue bool,
) (bool, error) {
	value := ip.GetInputWithDefault(inputs, key, defaultValue)
	return TypeConv.ConvertToBool(value)
}
