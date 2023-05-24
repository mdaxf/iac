package funcs

import (
	"fmt"
	"strings"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/engine/types"
)

type TableDeleteFuncs struct {
}

func (cf *TableDeleteFuncs) Execute(f *Funcs) {
	f.iLog.Debug(fmt.Sprintf("Start process %s : %s", "TableDeleteFuncs.Execute", f.Fobj.Name))

	namelist, valuelist, _ := f.SetInputs()

	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs content: %s", f.Fobj.Content))

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
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs columnList: %s", columnList))
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs columnvalueList: %s", columnvalueList))
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs columndatatypeList: %s", columndatatypeList))
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs keycolumnList: %s", keycolumnList))
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs keycolumnvalueList: %s", keycolumnvalueList))
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs keycolumndatatypeList: %s", keycolumndatatypeList))
	f.iLog.Debug(fmt.Sprintf("TableDeleteFuncs TableName: %s", TableName))

	if TableName == "" {
		f.iLog.Error(fmt.Sprintf("Error in TableDeleteFuncs.Execute: %s", "TableName is empty"))
		return
	}

	if len(keycolumnList) == 0 {
		f.iLog.Error(fmt.Sprintf("Error in TableDeleteFuncs.Execute: %s", "keycolumnList is empty"))
		return
	}

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

	var user string

	if f.SystemSession["User"] != nil {
		user = f.SystemSession["User"].(string)
	} else {
		user = "System"
	}

	dboperation := dbconn.NewDBOperation(user, f.DBTx, "TableDete Function")

	output, err := dboperation.TableDelete(TableName, Where)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in TableDete Execute: %s", err.Error()))
		return
	}
	f.iLog.Debug(fmt.Sprintf("TableDete Execution Result: %s", output))

	outputs := make(map[string]interface{})
	outputs["RowCount"] = output
	f.SetOutputs(outputs)
}

func (cf *TableDeleteFuncs) Validate(f *Funcs) (bool, error) {

	return true, nil
}

func (cf *TableDeleteFuncs) Testfunction(f *Funcs) (bool, error) {

	return true, nil
}
