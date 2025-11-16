package funcs

import (
	"fmt"
	"strings"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
)

type TableUpdateFuncs struct {
}

// Execute executes the TableUpdateFuncs function.
// It retrieves the inputs, sets up the necessary variables, and performs the table update operation.
// If any errors occur during the execution, it logs the error and returns.
// Finally, it sets the output values and returns.

func (cf *TableUpdateFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.TableUpdateFuncs.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.TableUpdateFuncs.Execute with error: %s", err))
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.TableUpdateFuncs.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.TableUpdateFuncs.Execute with error: %s", err)
			return
		}
	}()

	f.iLog.Debug(fmt.Sprintf("Start process %s : %s", "TableUpdateFuncs.Execute", f.Fobj.Name))

	namelist, valuelist, _ := f.SetInputs()

	f.iLog.Debug(fmt.Sprintf("TableUpdateFuncs content: %s", f.Fobj.Content))

	columnList := []string{}
	columnvalueList := []string{}
	columndatatypeList := []int{}
	keycolumnList := []string{}
	keycolumnvalueList := []string{}
	keycolumndatatypeList := []int{}
	TableName := ""

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
	f.iLog.Debug(fmt.Sprintf("TableUpdateFuncs columnList: %s", columnList))
	f.iLog.Debug(fmt.Sprintf("TableUpdateFuncs columnvalueList: %s", columnvalueList))
	f.iLog.Debug(fmt.Sprintf("TableUpdateFuncs columndatatypeList: %v", columndatatypeList))
	f.iLog.Debug(fmt.Sprintf("TableUpdateFuncs keycolumnList: %s", keycolumnList))
	f.iLog.Debug(fmt.Sprintf("TableUpdateFuncs keycolumnvalueList: %s", keycolumnvalueList))
	f.iLog.Debug(fmt.Sprintf("TableUpdateFuncs keycolumndatatypeList: %v", keycolumndatatypeList))
	f.iLog.Debug(fmt.Sprintf("TableUpdateFuncs TableName: %s", TableName))

	if TableName == "" {
		f.iLog.Error(fmt.Sprintf("Error in TableInsertFuncs.Execute: %s", "TableName is empty"))
		return
	}

	if len(columnList) == 0 {
		f.iLog.Error(fmt.Sprintf("Error in TableInsertFuncs.Execute: %s", "columnList is empty"))
		return
	}

	if len(keycolumnList) == 0 {
		f.iLog.Error(fmt.Sprintf("Error in TableInsertFuncs.Execute: %s", "keycolumnList is empty"))
		return
	}

	// Get database dialect for proper identifier quoting and WHERE clause building
	var user string
	if f.SystemSession["User"] != nil {
		user = f.SystemSession["User"].(string)
	} else {
		user = "System"
	}

	dboperation := dbconn.NewDBOperation(user, f.DBTx, "TableUpdate Function")

	// Build WHERE clause with database-specific quoted identifiers
	Where := ""
	for i, column := range keycolumnList {
		if Where != "" {
			Where = fmt.Sprintf("%s AND ", Where)
		}
		value := strings.Replace(keycolumnvalueList[i], "'", "", -1)

		// Use dialect-specific identifier quoting for better database portability
		quotedColumn := dboperation.QuoteIdentifier(column)
		Where = fmt.Sprintf("%s %s ='%s'", Where, quotedColumn, value)
	}
	f.iLog.Debug(fmt.Sprintf("TableUpdateFuncs Where: %s", Where))

	output, err := dboperation.TableUpdate(TableName, columnList, columnvalueList, columndatatypeList, Where)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in TableUpdate Execute: %s", err.Error()))
		return
	}
	f.iLog.Debug(fmt.Sprintf("TableUpdate Execution Result: %v", output))

	outputs := make(map[string]interface{})
	outputs["RowCount"] = output
	f.SetOutputs(outputs)
}

// Validate is a method of the TableUpdateFuncs struct that validates the function.
// It measures the performance of the function and logs the duration.
// It returns a boolean value indicating the validation result and an error if any.

func (cf *TableUpdateFuncs) Validate(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.TableUpdateFuncs.Validate", elapsed)
	}()
	/*defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.TableUpdateFuncs.Validate with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.TableUpdateFuncs.Validate with error: %s", err)
			return
		}
	}() */

	return true, nil
}

// Testfunction is a function that performs a test operation.
// It measures the performance of the function and handles any errors that occur.
// It returns a boolean value indicating the success of the test and an error if any.

func (cf *TableUpdateFuncs) Testfunction(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.TableUpdateFuncs.Testfunction", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.TableUpdateFuncs.Testfunction with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.TableUpdateFuncs.Testfunction with error: %s", err)
			return
		}
	}()

	return true, nil
}
