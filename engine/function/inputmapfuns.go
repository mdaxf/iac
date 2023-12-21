package funcs

import (
	"fmt"
	"time"

	"github.com/mdaxf/iac/engine/types"
)

type InputMapFuncs struct{}

// Execute executes the input mapping functions.
// It retrieves the input values from the Funcs object, maps them to the corresponding keys in the map data,
// and sets the output values in the Funcs object.
// If there is an error during execution, it logs the error, cancels the execution, and sets the error message.
// It also logs the performance of the function.

func (cf *InputMapFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.InputMapFuncs.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			errormsg := fmt.Sprintf("There is error to engine.funcs.InputMapFuncs.Execute with error: %s", err)
			f.iLog.Error(errormsg)
			f.CancelExecution(errormsg)
			f.ErrorMessage = errormsg
			return
		}
	}()
	namelist, valuelist, _ := f.SetInputs()

	data := f.Fobj.Mapdata

	//var data map[string]interface{}

	// Convert the string to JSON format
	/*err := json.Unmarshal([]byte(mapstr), &data)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error: %v", err))
	} */
	outputs := make(map[string]interface{})

	for key := range data {
		mapedinput := data[key]
		for i := range namelist {
			if mapedinput == namelist[i] {
				outputs[key] = valuelist[i]
			}
		}
	}

	f.SetOutputs(outputs)
}

// Validate validates the input map functions.
// It checks if the inputs and outputs in the map are valid and compatible.
// If any error is encountered, it logs the error and returns false.
// Otherwise, it returns true.
// It also logs the performance of the function.

func (cf *InputMapFuncs) Validate(f *Funcs) bool {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.InputMapFuncs.Validate", elapsed)
	}()
	/*	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.InputMapFuncs.Validate with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.InputMapFuncs.Validate with error: %s", err)
			return
		}
	}() */

	data := f.Fobj.Mapdata

	//var data map[string]interface{}

	// Convert the string to JSON format
	/*err := json.Unmarshal([]byte(mapstr), &data)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error: %v", err))
		return false
	} */
	//	outputs := make(map[string]interface{})

	for key := range data {
		mapedinput := data[key]
		mapedoutput := key

		var input types.Input
		var output types.Output
		foundinput := false
		foundoutput := false
		for i := range f.Fobj.Inputs {
			if mapedinput == f.Fobj.Inputs[i].Name {
				input = f.Fobj.Inputs[i]
				foundinput = true
				break
			}
		}

		for i := range f.Fobj.Outputs {
			if mapedoutput == f.Fobj.Outputs[i].Name {
				output = f.Fobj.Outputs[i]
				foundoutput = true
				break
			}
		}

		if foundinput && foundoutput {
			if output.Datatype == types.String {
				continue
			}
			if output.Datatype == types.Integer && input.Datatype != types.Integer {
				f.iLog.Error(fmt.Sprintf("Output %s is Integer and input %s is not Integer", output.Name, input.Name))
				return false
			}
			if output.Datatype == types.Float && (input.Datatype != types.Float && input.Datatype != types.Integer) {
				f.iLog.Error(fmt.Sprintf("Output %s is float and input %s is not Integer or float", output.Name, input.Name))
				return false
			}

			if output.Datatype == types.Bool && input.Datatype != types.Bool {
				f.iLog.Error(fmt.Sprintf("Output %s is boolean and input %s is not boolean", output.Name, input.Name))
				return false
			}

			if output.Datatype == types.DateTime && input.Datatype != types.DateTime {
				f.iLog.Error(fmt.Sprintf("Output %s is date and input %s is not date", output.Name, input.Name))
				return false
			}

			if output.List != input.List {
				f.iLog.Error(fmt.Sprintf("Output %s and input %s are not same list flag", output.Name, input.Name))
				return false
			}

		} else {
			f.iLog.Error(fmt.Sprintf("Output %s or input %s are not found", mapedoutput, mapedinput))
			return false
		}

	}

	return true
}
