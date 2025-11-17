package funcs

import (
	"fmt"
	"time"

	"github.com/mdaxf/iac/engine/types"
)

type ThrowErrorFuncs struct {
}

// Execute executes the ThrowErrorFuncs function.
// This function is intentionally designed to trigger transaction rollback.
// When executed with iserror=true, it ensures the entire transaction is rolled back
// to prevent partial data changes, maintaining data integrity.
//
// ROLLBACK DESIGN:
// - Creates a structured BPMError with full execution context
// - Rolls back the database transaction
// - Cancels the execution context
// - Panics with the error to propagate rollback up the call chain
// - Parent transcode will catch this panic and handle rollback
func (cf *ThrowErrorFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.ThrowErrorFuncs.Execute", elapsed)
	}()

	defer func() {
		// Catch any unexpected panics during error creation
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("Unexpected error in ThrowErrorFuncs.Execute: %s", err))
			f.ErrorMessage = fmt.Sprintf("Unexpected error in ThrowErrorFuncs.Execute: %s", err)
			f.CancelExecution(f.ErrorMessage)
		}
	}()

	// Get inputs to determine error details
	_, _, inputs := f.SetInputs()

	// Extract error message and severity from inputs
	errorMessage := "Business error occurred"
	if msg, ok := inputs["message"]; ok {
		errorMessage = fmt.Sprintf("%v", msg)
	}

	// Check if this is an actual error condition
	isError := true
	if isErrorInput, ok := inputs["iserror"]; ok {
		if val, ok := isErrorInput.(bool); ok {
			isError = val
		} else if val, ok := isErrorInput.(string); ok {
			isError = val == "true" || val == "1"
		}
	}

	// If not an error condition, just log and return
	if !isError {
		f.iLog.Info(fmt.Sprintf("ThrowError executed with iserror=false: %s", errorMessage))
		outputs := make(map[string]interface{})
		outputs["result"] = "no_error"
		f.SetOutputs(outputs)
		return
	}

	// Create execution context for the error
	execContext := &types.ExecutionContext{
		FunctionName: f.Fobj.Name,
		FunctionType: f.Fobj.Functype.String(),
		ExecutionTime: startTime,
	}

	// Add user context if available
	if userNo, ok := f.SystemSession["UserNo"].(string); ok {
		execContext.UserNo = userNo
	}
	if clientID, ok := f.SystemSession["ClientID"].(string); ok {
		execContext.ClientID = clientID
	}

	// Create structured business error
	bpmErr := types.NewBusinessError(errorMessage).
		WithContext(execContext).
		WithRollbackReason("Explicit error thrown via ThrowError function").
		WithDetail("function_name", f.Fobj.Name)

	// Add any additional context from inputs
	if category, ok := inputs["category"]; ok {
		bpmErr.WithDetail("error_category", fmt.Sprintf("%v", category))
	}
	if code, ok := inputs["error_code"]; ok {
		bpmErr.WithDetail("error_code", fmt.Sprintf("%v", code))
	}

	// Log the formatted error
	f.iLog.Error(bpmErr.GetFormattedError())

	// Set error message for propagation
	f.ErrorMessage = bpmErr.Error()

	// Rollback the database transaction
	f.iLog.Info("Rolling back transaction due to ThrowError function")
	f.DBTx.Rollback()

	// Cancel execution context
	f.CtxCancel()
	f.Ctx.Done()

	// Panic with the structured error to trigger rollback up the chain
	// This is BY DESIGN - the panic/recover pattern ensures transaction atomicity
	panic(bpmErr)
}
