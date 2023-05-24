package funcs

import (
	"fmt"
	"strings"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/engine/types"
)

type TableUpdateFuncs struct {
}

func (cf *TableUpdateFuncs) Execute(f *Funcs) {
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
	f.iLog.Debug(fmt.Sprintf("TableUpdateFuncs columndatatypeList: %s", columndatatypeList))
	f.iLog.Debug(fmt.Sprintf("TableUpdateFuncs keycolumnList: %s", keycolumnList))
	f.iLog.Debug(fmt.Sprintf("TableUpdateFuncs keycolumnvalueList: %s", keycolumnvalueList))
	f.iLog.Debug(fmt.Sprintf("TableUpdateFuncs keycolumndatatypeList: %s", keycolumndatatypeList))
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

	dboperation := dbconn.NewDBOperation(user, f.DBTx, "TableUpdate Function")

	output, err := dboperation.TableUpdate(TableName, columnList, columnvalueList, columndatatypeList, Where)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in TableUpdate Execute: %s", err.Error()))
		return
	}
	f.iLog.Debug(fmt.Sprintf("TableUpdate Execution Result: %s", output))

	outputs := make(map[string]interface{})
	outputs["RowCount"] = output
	f.SetOutputs(outputs)
}

func (cf *TableUpdateFuncs) Validate(f *Funcs) (bool, error) {

	return true, nil
}

func (cf *TableUpdateFuncs) Testfunction(f *Funcs) (bool, error) {

	return true, nil
}
