package funcs

import (
	"fmt"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
)

type StoreProcFuncs struct {
}

func (cf *StoreProcFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.StoreProcedure.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.StoreProcedure.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.StoreProcedure.Execute with error: %s", err)
			return
		}
	}()

	f.iLog.Debug(fmt.Sprintf("Start process %s : %s", "StoreProcFuncs.Execute", f.Fobj.Name))

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
	outputs["ColumnCount"] = []interface{}{ColumnCount}
	outputs["RowCount"] = []interface{}{RowCount}
	f.SetOutputs(f.convertMap(outputs))
}

func (cf *StoreProcFuncs) Validate(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.StoreProcedure.Validate", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.StoreProcedure.Validate with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.StoreProcedure.Validate with error: %s", err)
			return
		}
	}()

	return true, nil
}

func (cf *StoreProcFuncs) Testfunction(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.StoreProcedure.Testfunction", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.StoreProcedure.Testfunction with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.StoreProcedure.Testfunction with error: %s", err)
			return
		}
	}()

	return true, nil
}
