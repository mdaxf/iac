package funcs

import (
	"fmt"

	"github.com/robertkrimen/otto"
)

type JSFuncs struct {
}

func (cf *JSFuncs) Execute(f *Funcs) {
	f.iLog.Debug(fmt.Sprintf("Start process %s : %s", "JSFuncs.Execute", f.Fobj.Name))

	namelist, _, inputs := f.SetInputs()

	vm := otto.New()
	for i := 0; i < len(namelist); i++ {
		vm.Set(namelist[i], inputs[namelist[i]])
	}
	f.iLog.Debug(fmt.Sprintf("Fucntion %s script: %s", f.Fobj.Name, f.Fobj.Content))

	vm.Run(f.Fobj.Content)

	outputs := make(map[string]interface{})

	for i := 0; i < len(f.Fobj.Outputs); i++ {
		if value, err := vm.Get(f.Fobj.Outputs[i].Name); err == nil {
			outputs[f.Fobj.Outputs[i].Name] = value.String()
		} else {
			f.iLog.Error(fmt.Sprintf("Error in JSFuncs.Execute: %s", err.Error()))
			return
		}
	}

	f.SetOutputs(outputs)
}

func (cf *JSFuncs) Validate(f *Funcs) (bool, error) {
	vm := otto.New()
	_, err := vm.Compile("", f.Fobj.Content)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (cf *JSFuncs) Testfunction(content string, inputs interface{}, outputs []string) (map[string]interface{}, error) {

	/* pass, err := cf.Validate(f)

	if !pass {
		return nil, err
	}
	*/
	namelist := make([]string, 0)
	valuelist := make([]interface{}, 0)

	for key, value := range inputs.(map[string]interface{}) {
		namelist = append(namelist, key)
		valuelist = append(valuelist, value)
	}

	vm := otto.New()
	for i := 0; i < len(namelist); i++ {
		vm.Set(namelist[i], valuelist[i])
	}

	vm.Run(content)

	functionoutputs := make(map[string]interface{})

	for i := 0; i < len(outputs); i++ {
		if value, err := vm.Get(outputs[i]); err == nil {
			functionoutputs[outputs[i]] = value
		} else {
			fmt.Println(err)
		}
	}
	return functionoutputs, nil
}
