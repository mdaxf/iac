package funcs

import (
	"fmt"
	"strings"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/engine/types"
)

type TableDeleteFuncs struct {
}

// Execute executes the TableDeleteFuncs function.
// It retrieves the input values, constructs the WHERE clause based on the key columns,
// and performs a table delete operation using the provided table name and WHERE clause.
// The result is stored in the "RowCount" output.
// If any error occurs during the execution, it is logged and returned.
func (cf *TableDeleteFuncs) Execute(f *Funcs) {
	// function execution start time
	startTime := time.Now()

	// defer function to log the performance duration
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.TableDeleteFuncs.Execute", elapsed)
	}()

	// defer function to recover from panics and log the error
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.TableDeleteFuncs.Execute with error: %s", err))
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.TableDeleteFuncs.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.TableDeleteFuncs.Execute with error: %s", err)
			return
		}
	}()

	// log the start of the process
	f.iLog.Debug(fmt.Sprintf("Start process %s : %s", "TableDeleteFuncs.Execute", f.Fobj.Name))

	// retrieve the input values
	namelist, valuelist, _ := f.SetInputs()

	// log the content of TableDeleteFuncs
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs content: %s", f.Fobj.Content))

	// initialize lists for columns, column values, column datatypes, key columns, key column values, and key column datatypes
	columnList := []string{}
	columnvalueList := []string{}
	columndatatypeList := []int{}
	keycolumnList := []string{}
	keycolumnvalueList := []string{}
	keycolumndatatypeList := []int{}
	TableName := ""

	// iterate over the input names and values
	for i, name := range namelist {
		if name == "TableName" {
			TableName = valuelist[i]
		} else if strings.HasSuffix(name, "KEY") {
			name = strings.Replace(name+"_|", "KEY_|", "", -1)
			keycolumnList = append(keycolumnList, name)
			keycolumnvalueList = append(keycolumnvalueList, valuelist[i])
			keycolumndatatypeList = append(keycolumndatatypeList, int(f.Fobj.Inputs[i].Datatype))
		} else {
			columnList = append(columnList, name)
			columnvalueList = append(columnvalueList, valuelist[i])
			columndatatypeList = append(columndatatypeList, int(f.Fobj.Inputs[i].Datatype))
		}
	}

	// log the column lists and TableName
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs columnList: %s", columnList))
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs columnvalueList: %s", columnvalueList))
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs columndatatypeList: %v", columndatatypeList))
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs keycolumnList: %s", keycolumnList))
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs keycolumnvalueList: %s", keycolumnvalueList))
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs keycolumndatatypeList: %v", keycolumndatatypeList))
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs TableName: %s", TableName))

	// check if TableName is empty
	if TableName == "" {
		f.iLog.Error(fmt.Sprintf("Error in TableDeleteFuncs.Execute: %s", "TableName is empty"))
		return
	}

	// check if keycolumnList is empty
	if len(keycolumnList) == 0 {
		f.iLog.Error(fmt.Sprintf("Error in TableDeleteFuncs.Execute: %s", "keycolumnList is empty"))
		return
	}

	// construct the WHERE clause based on the key columns
	Where := ""
	for i, column := range keycolumnList {
		switch keycolumndatatypeList[i] {
		case int(types.String):
		case int(types.DateTime):
			Where += column + " = '" + keycolumnvalueList[i] + "'"
		default:
			Where += column + " = " + keycolumnvalueList[i]
		}
		if i < len(keycolumnList)-1 {
			Where += " AND "
		}
	}

	// get the user from the system session or set it to "System" if not available
	var user string
	if f.SystemSession["User"] != nil {
		user = f.SystemSession["User"].(string)
	} else {
		user = "System"
	}

	// create a new DBOperation instance
	dboperation := dbconn.NewDBOperation(user, f.DBTx, "TableDete Function")

	// perform the table delete operation
	output, err := dboperation.TableDelete(TableName, Where)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in TableDete Execute: %s", err.Error()))
		return
	}

	// log the execution result
	f.iLog.Debug(fmt.Sprintf("TableDete Execution Result: %v", output))

	// store the result in the "RowCount" output
	outputs := make(map[string]interface{})
	outputs["RowCount"] = output
	f.SetOutputs(outputs)
}

// Validate validates the TableDeleteFuncs function.
// It measures the performance of the function and logs any errors that occur.
// Returns true if the validation is successful, otherwise returns an error.

func (cf *TableDeleteFuncs) Validate(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.TableDeleteFuncs.Validate", elapsed)
	}()
	/*	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.TableDeleteFuncs.Validate with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.TableDeleteFuncs.Validate with error: %s", err)
			return
		}
	}() */

	return true, nil
}

// Testfunction is a function that performs a test operation.
// It measures the performance of the function and logs the duration.
// It returns a boolean value indicating the success of the test and an error if any.

func (cf *TableDeleteFuncs) Testfunction(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.TableDeleteFuncs.Testfunction", elapsed)
	}()
	/*	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.TableDeleteFuncs.Testfunction with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.TableDeleteFuncs.Testfunction with error: %s", err)
			return
		}
	}() */

	return true, nil
}
