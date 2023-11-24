package funcs

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/mdaxf/iac/logger"
	"github.com/robertkrimen/otto"
)

type JSFuncs struct {
}

func (cf *JSFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.JSFuncs.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.JSFuncs.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.JSFuncs.Execute with error: %s", err)
			return
		}
	}()
	f.iLog.Debug(fmt.Sprintf("Start process %s : %s", "JSFuncs.Execute", f.Fobj.Name))

	namelist, _, inputs := f.SetInputs()

	vm := goja.New()

	for i := 0; i < len(namelist); i++ {
		vm.Set(namelist[i], inputs[namelist[i]])
	}
	f.iLog.Debug(fmt.Sprintf("Fucntion %s script: %s", f.Fobj.Name, f.Fobj.Content))

	value, err := vm.RunString(f.Fobj.Content)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in JSFuncs.Execute: %s", err.Error()))
		return
	}

	result := value.Export()
	f.iLog.Debug(fmt.Sprintf("Fucntion %s result: %s", f.Fobj.Name, result))

	outputs := make(map[string]interface{})

	for i := 0; i < len(f.Fobj.Outputs); i++ {
		value := vm.Get(f.Fobj.Outputs[i].Name)
		outputs[f.Fobj.Outputs[i].Name] = value.String()
	}
	f.iLog.Debug(fmt.Sprintf("Fucntion %s outputs: %s", f.Fobj.Name, outputs))
	f.SetOutputs(outputs)
}

func (cf *JSFuncs) Validate(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.JSFuncs.Validate", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.JSFuncs.Validate with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.JSFuncs.Validate with error: %s", err)
			return
		}
	}()

	vm := goja.New()
	_, err := vm.RunString(f.Fobj.Content)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (cf *JSFuncs) Testfunction(content string, inputs interface{}, outputs []string) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.TranCode, User: "System", ControllerName: "JSFuncs"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("engine.funcs.JSFuncs.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to engine.funcs.JSFuncs.Execute with error: %s", err))
			//ErrorMessage = fmt.Sprintf("There is error to engine.funcs.JSFuncs.Execute with error: %s", err)
			return
		}
	}()

	namelist := make([]string, 0)
	valuelist := make([]interface{}, 0)

	for key, value := range inputs.(map[string]interface{}) {
		namelist = append(namelist, key)
		valuelist = append(valuelist, value)
	}

	vm := goja.New()

	for i := 0; i < len(namelist); i++ {
		vm.Set(namelist[i], valuelist[i])
	}

	value, err := vm.RunString(content)

	if err != nil {
		println(err.Error())
		return nil, err
	}

	result := value.Export()
	functionoutputs := make(map[string]interface{})

	println(result)

	for i := 0; i < len(outputs); i++ {
		value := vm.Get(outputs[i])
		functionoutputs[outputs[i]] = value
	}

	return functionoutputs, nil
}

func (cf *JSFuncs) Execute_otto(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.JSFuncs.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.JSFuncs.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.JSFuncs.Execute with error: %s", err)
			return
		}
	}()
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

func (cf *JSFuncs) Validate_otto(f *Funcs) (bool, error) {
	vm := otto.New()
	_, err := vm.Compile("", f.Fobj.Content)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (cf *JSFuncs) Testfunction_otto(content string, inputs interface{}, outputs []string) (map[string]interface{}, error) {

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
