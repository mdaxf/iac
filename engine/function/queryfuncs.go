package funcs

import (
	"fmt"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
)

type QueryFuncs struct {
}

// Execute executes the query function.
// It measures the execution time and logs performance metrics.
// If there is an error during execution, it logs the error and sets the error message.
// It sets the user session and creates a SELECT clause with aliases.
// It performs the query operation using the provided database connection.
// It logs the outputs, column count, and row count.
// Finally, it sets the outputs of the function.
func (cf *QueryFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.QueryFuncs.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.QueryFuncs.Execute with error: %s", err))
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.QueryFuncs.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.QueryFuncs.Execute with error: %s", err)
			return
		}
	}()
	f.iLog.Debug(fmt.Sprintf("Start process %s : %s", "QueryFuncs.Execute", f.Fobj.Name))

	namelist, _, inputs := f.SetInputs()

	var user string

	if f.SystemSession["User"] != nil {
		user = f.SystemSession["User"].(string)
	} else {
		user = "System"
	}

	// Create SELECT clause with aliases
	dboperation := dbconn.NewDBOperation(user, f.DBTx, "Execute Query Function")

	outputs, ColumnCount, RowCount, err := dboperation.QuerybyList(f.Fobj.Content, namelist, inputs, f.Fobj.Inputs)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in QueryFuncs.Execute: %s", err.Error()))
		return
	}

	f.iLog.Debug(fmt.Sprintf("QueryFuncs outputs: %s", outputs))
	f.iLog.Debug(fmt.Sprintf("QueryFuncs ColumnCount: %d", ColumnCount))
	f.iLog.Debug(fmt.Sprintf("QueryFuncs RowCount: %d", RowCount))

	outputs["ColumnCount"] = []interface{}{ColumnCount}
	outputs["RowCount"] = []interface{}{RowCount}
	f.SetOutputs(f.convertMap(outputs))

}

// Validate is a method of the QueryFuncs struct that validates the function.
// It measures the performance of the function and logs any errors that occur.
// It returns a boolean value indicating whether the validation was successful and an error if any.

func (cf *QueryFuncs) Validate(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.QueryFuncs.Validate", elapsed)
	}()
	/*	defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.QueryFuncs.Validate with error: %s", err))
				f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.QueryFuncs.Validate with error: %s", err)
				return
			}
		}()
	*/
	return true, nil
}

// Testfunction is a function that performs a test operation.
// It measures the performance of the function and logs the duration.
// It returns a boolean value indicating the success of the test and an error if any.

func (cf *QueryFuncs) Testfunction(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.QueryFuncs.Testfunction", elapsed)
	}()
	/*	defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.QueryFuncs.Testfunction with error: %s", err))
				f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.QueryFuncs.Testfunction with error: %s", err)
				return
			}
		}()
	*/
	return true, nil
}
