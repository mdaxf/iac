package funcs

import (
	"fmt"

	"github.com/mdaxf/iac/engine/types"
)

type InputMapFuncs struct{}

func (cf *InputMapFuncs) Execute(f *Funcs) {
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

func (cf *InputMapFuncs) Validate(f *Funcs) bool {
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
