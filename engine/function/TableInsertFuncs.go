package funcs

import (
	"fmt"

	dbconn "github.com/mdaxf/iac/databases"
)

type TableInsertFuncs struct {
}

func (cf *TableInsertFuncs) Execute(f *Funcs) {
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

func (cf *TableInsertFuncs) Validate(f *Funcs) (bool, error) {

	return true, nil
}

func (cf *TableInsertFuncs) Testfunction(f *Funcs) (bool, error) {

	return true, nil
}
