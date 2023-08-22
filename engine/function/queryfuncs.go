package funcs

import (
	"fmt"

	dbconn "github.com/mdaxf/iac/databases"
)

type QueryFuncs struct {
}

func (cf *QueryFuncs) Execute(f *Funcs) {
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

func (cf *QueryFuncs) Validate(f *Funcs) (bool, error) {

	return true, nil
}

func (cf *QueryFuncs) Testfunction(f *Funcs) (bool, error) {

	return true, nil
}
