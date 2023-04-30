package funcs

import (
	"fmt"

	"github.com/robertkrimen/otto"
)

type JSFuncs struct {
}

func (cf *JSFuncs) Execute(f *Funcs) {
	namelist, valuelist, _ := f.SetInputs()

	vm := otto.New()
	for i := 0; i < len(namelist); i++ {
		vm.Set(namelist[i], valuelist[i])
	}

	vm.Run(f.Fobj.Content)

	outputs := make(map[string]interface{})

	for i := 0; i < len(f.Fobj.Outputs); i++ {
		if value, err := vm.Get(f.Fobj.Outputs[i].Name); err == nil {
			outputs[f.Fobj.Outputs[i].Name] = value.String()
		} else {
			fmt.Println(err)
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

func (cf *JSFuncs) Testfunction(f *Funcs) (map[string]interface{}, error) {

	pass, err := cf.Validate(f)

	if !pass {
		return nil, err
	}

	namelist, valuelist, _ := f.SetInputs()

	vm := otto.New()
	for i := 0; i < len(namelist); i++ {
		vm.Set(namelist[i], valuelist[i])
	}

	vm.Run(f.Fobj.Content)

	outputs := make(map[string]interface{})

	for i := 0; i < len(f.Fobj.Outputs); i++ {
		if value, err := vm.Get(f.Fobj.Outputs[i].Name); err == nil {
			outputs[f.Fobj.Outputs[i].Name] = value.String()
		} else {
			fmt.Println(err)
		}
	}
	return outputs, nil
}
