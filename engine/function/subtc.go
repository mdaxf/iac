package funcs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	//	"github.com/mdaxf/iac/engine/callback"
	"github.com/mdaxf/iac/framework/callback_mgr"
)

type TranFlow interface {
	Execute(string, map[string]interface{}, context.Context, context.CancelFunc, *sql.Tx) (map[string]interface{}, error)
}

type SubTranCodeFuncs struct {
	TranFlowstr TranFlow
}

// NewSubTran creates a new instance of SubTranCodeFuncs with the provided TranFlow.
// It returns a pointer to the newly created SubTranCodeFuncs.
func NewSubTran(tci TranFlow) *SubTranCodeFuncs {
	return &SubTranCodeFuncs{
		TranFlowstr: tci,
	}

}

type SubTranCode struct {
}

// Execute executes the subtran function.
// It sets the inputs, retrieves the transaction code, and calls the callback function to execute the transaction code.
// The outputs are then converted and set as the function outputs.
// If there is an error during execution, it logs the error and sets the error message.
// It also logs the performance of the function.
func (cf *SubTranCodeFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.SubTransCode.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SubTransCode.Execute with error: %s", err))
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.SubTransCode.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.SubTransCode.Execute with error: %s", err)
			return
		}
	}()

	tcode := ""
	f.iLog.Debug(fmt.Sprintf("Executing subtran function"))
	namelist, valuelist, mappedinputs := f.SetInputs()

	for i, name := range namelist {
		if name == "TranCode" {
			tcode = valuelist[i]
		}
	}
	if tcode == "" {
		f.iLog.Error(fmt.Sprintf("Error executing transaction code: %s", "No trancode input"))
		return
	}

	f.iLog.Debug(fmt.Sprintf("Executing subtran function to call transaction code: %v with inputs %s", tcode, mappedinputs))

	outputs, err := callback_mgr.CallBackFunc("TranCode_Execute", tcode, mappedinputs, nil, nil, f.DBTx, f.DocDBCon, f.SignalRClient)
	//outputs := callback.ExecuteTranCode("TranFlowstr_Execute", tcode, mappedinputs, nil, nil, f.DBTx, f.DocDBCon, f.SignalRClient)
	//outputs, err := cf.TranFlowstr.Execute(tcode, mappedinputs, f.Ctx, f.CtxCancel, f.DBTx)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error executing transaction code: %v", err))
		return
	}
	f.SetOutputs(convertSliceToMap(outputs))
}

// Validate is a method of the SubTranCodeFuncs struct that validates the function.
// It measures the performance of the function and logs the duration.
// It returns a boolean value indicating the success of the validation and an error if any.
// It also logs the performance of the function.
func (cf *SubTranCodeFuncs) Validate(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.SubTransCode.Validate", elapsed)
	}()
	/*	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SubTransCode.Validate with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.SubTransCode.Validate with error: %s", err)
			return
		}
	}() */

	return true, nil
}

// convertSliceToMap converts a slice of interfaces into a map[string]interface{}.
// It iterates over the slice, treating every even-indexed element as the key and the following odd-indexed element as the value.
// If the key is a string, it adds the key-value pair to the resulting map.
// The function returns the resulting map.
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
