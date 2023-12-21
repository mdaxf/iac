package funcs

import (
	"fmt"
	"time"
)

type ThrowErrorFuncs struct {
}

// Execute executes the ThrowErrorFuncs function.
// It measures the execution time and logs the performance.
// It recovers from any panics and logs the error.
// It rolls back the database transaction, cancels the execution context, and sets the error message.
func (cf *ThrowErrorFuncs) Execute(f *Funcs) {
	// function execution start time
	startTime := time.Now()
	defer func() {
		// calculate elapsed time
		elapsed := time.Since(startTime)
		// log performance with duration
		f.iLog.PerformanceWithDuration("engine.funcs.ThrowErrorFuncs.Execute", elapsed)
	}()

	defer func() {
		// recover from any panics
		if err := recover(); err != nil {
			// log the error
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.ThrowErrorFuncs.Execute with error: %s", err))
			// cancel execution and set error message
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.ThrowErrorFuncs.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.ThrowErrorFuncs.Execute with error: %s", err)
		}
	}()

	// log debug information
	f.iLog.Debug(fmt.Sprintf("ThrowErrorFuncs Execute: %v", f))
	// rollback database transaction
	f.DBTx.Rollback()
	// cancel execution context
	f.CtxCancel()
	// mark execution as done
	f.Ctx.Done()
	// set error message
	f.ErrorMessage = "ThrowErrorFuncs Execute"
}
