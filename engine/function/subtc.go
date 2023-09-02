package funcs

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mdaxf/iac/engine/callback"
)

type TranFlow interface {
	Execute(string, map[string]interface{}, context.Context, context.CancelFunc, *sql.Tx) (map[string]interface{}, error)
}

type SubTranCodeFuncs struct {
	TranFlowstr TranFlow
}

func NewSubTran(tci TranFlow) *SubTranCodeFuncs {
	return &SubTranCodeFuncs{
		TranFlowstr: tci,
	}

}

type SubTranCode struct {
}

func (cf *SubTranCodeFuncs) Execute(f *Funcs) {
	tcode := ""
	f.iLog.Debug(fmt.Sprintf("Executing subtran function"))
	namelist, valuelist, mappedinputs := f.SetInputs()

	for i, name := range namelist {
		if name == "TranCode" {
			tcode = valuelist[i]
		}
	}
	if tcode == "" {
		f.iLog.Error(fmt.Sprintf("Error executing transaction code: ", "No trancode input"))
		return
	}

	f.iLog.Debug(fmt.Sprintf("Executing subtran function to call transaction code: %v with inputs %s", tcode, mappedinputs))

	outputs := callback.ExecuteTranCode("TranFlowstr_Execute", tcode, mappedinputs, nil, nil, f.DBTx)
	//outputs, err := cf.TranFlowstr.Execute(tcode, mappedinputs, f.Ctx, f.CtxCancel, f.DBTx)
	/*if err != nil {
		f.iLog.Error(fmt.Sprintf("Error executing transaction code: %v", err))
		return
	}  */
	f.SetOutputs(convertSliceToMap(outputs))
}

func (cf *SubTranCodeFuncs) Validate(f *Funcs) (bool, error) {

	return true, nil
}

func convertSliceToMap(slice []interface{}) map[string]interface{} {
	resultMap := make(map[string]interface{})

	for i := 0; i < len(slice); i += 2 {
		key, ok1 := slice[i].(string)
		value := slice[i+1]

		if ok1 {
			resultMap[key] = value
		}
	}

	return resultMap
}
