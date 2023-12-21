package funcs

import (
	"fmt"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
)

type TableInsertFuncs struct {
}

// Execute executes the TableInsertFuncs function.
// It retrieves the input values, sets up the necessary variables,
// and performs the table insert operation using the provided database connection.
// The execution can be canceled if an error occurs during the process.
// The function also logs debug and error messages for troubleshooting purposes.
// Finally, it sets the output values for further processing.

func (cf *TableInsertFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.TableInsertFuncs.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.TableInsertFuncs.Execute with error: %s", err))
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.TableInsertFuncs.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.TableInsertFuncs.Execute with error: %s", err)
			return
		}
	}()

	f.iLog.Debug(fmt.Sprintf("Start process %s : %s", "TableInsertFuncs.Execute", f.Fobj.Name))

	namelist, valuelist, _ := f.SetInputs()

	columnList := []string{}
	columnvalueList := []string{}
	columndatatypeList := []int{}

	var execution bool
	execution = true

	TableName := ""
	for i, name := range namelist {
		if name == "TableName" {
			TableName = valuelist[i]
		} else if name == "Execution" {
			if valuelist[i] == "false" {
				execution = false
			}
		} else {
			columnList = append(columnList, name)
			columnvalueList = append(columnvalueList, valuelist[i])
			columndatatypeList = append(columndatatypeList, int(f.Fobj.Inputs[i].Datatype))
		}

	}

	f.iLog.Debug(fmt.Sprintf("TableInsertFuncs columnList: %s", columnList))
	f.iLog.Debug(fmt.Sprintf("TableInsertFuncs columnvalueList: %s", columnvalueList))
	f.iLog.Debug(fmt.Sprintf("TableInsertFuncs columndatatypeList: %v", columndatatypeList))
	f.iLog.Debug(fmt.Sprintf("TableInsertFuncs TableName: %s", TableName))

	if execution == false {
		f.iLog.Debug(fmt.Sprintf("TableInsertFuncs Execution is false, skip execution"))
		return
	}

	if TableName == "" {
		f.iLog.Error(fmt.Sprintf("Error in TableInsertFuncs.Execute: %s", "TableName is empty"))
		return
	}

	if len(columnList) == 0 {
		f.iLog.Error(fmt.Sprintf("Error in TableInsertFuncs.Execute: %s", "columnList is empty"))
		return
	}

	var user string

	if f.SystemSession["User"] != nil {
		user = f.SystemSession["User"].(string)
	} else {
		user = "System"
	}

	// Create SELECT clause with aliases
	dboperation := dbconn.NewDBOperation(user, f.DBTx, "TableInsert Function")

	output, err := dboperation.TableInsert(TableName, columnList, columnvalueList)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in TableInsert Execute: %s", err.Error()))
		return
	}
	f.iLog.Debug(fmt.Sprintf("TableInsert Execution Result: %v", output))

	outputs := make(map[string]interface{})
	outputs["Identify"] = output
	f.SetOutputs(outputs)
}

// Validate validates the TableInsertFuncs.
// It measures the performance of the function and logs the duration.
// It returns true if the validation is successful, otherwise it returns an error.
// It also logs the performance of the function.
func (cf *TableInsertFuncs) Validate(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.TableInsertFuncs.Validate", elapsed)
	}()
	/*defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.TableInsertFuncs.Validate with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.TableInsertFuncs.Validate with error: %s", err)
			return
		}
	}() */

	return true, nil
}

// Testfunction is a function that performs a test operation.
// It measures the performance of the function and logs the duration.
// It returns a boolean value indicating the success of the test and an error if any.

func (cf *TableInsertFuncs) Testfunction(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.TableInsertFuncs.Testfunction", elapsed)
	}()
	/*defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.TableInsertFuncs.Testfunction with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.TableInsertFuncs.Testfunction with error: %s", err)
			return
		}
	}() */

	return true, nil
}
