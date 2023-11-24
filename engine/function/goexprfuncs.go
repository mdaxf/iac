package funcs

import (
	"fmt"
	"time"

	"github.com/antonmedv/expr"
	//	"reflect"
	"github.com/mdaxf/iac/logger"
)

type GoExprFuncs struct {
}

func (cf *GoExprFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.GoExprFuncs.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.GoExprFuncs.Execute with error: %s", err))
			return
		}
	}()

	namelist, _, inputs := f.SetInputs()

	env := make(map[string]interface{}, 0)
	for i := range namelist {
		env[namelist[i]] = inputs[namelist[i]]
	}

	program, err := expr.Compile(f.Fobj.Content, expr.Env(env))
	if err != nil {
		f.iLog.Error(fmt.Sprintf("failed to compile expression: %v", err))
		return
	}

	output, err := expr.Run(program, env)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("failed to run expression: %v", err))
		return
	}

	f.SetOutputs(output.(map[string]interface{}))
}

func (cf *GoExprFuncs) Validate(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.GoExprFuncs.Validate", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.GoExprFuncs.Validate with error: %s", err))
			return
		}
	}()

	return true, nil
}

func (cf *GoExprFuncs) Testfunction(content string, inputs interface{}, outputs []string) (map[string]interface{}, error) {

	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "CSharp Function"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("engine.funcs.GoExprFuncs.Testfunction", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to engine.funcs.GoExprFuncs.Testfunction with error: %s", err))
			return
		}
	}()

	iLog.Debug(fmt.Sprintf("Test Exec Function"))

	iLog.Debug(fmt.Sprintf("Test Exec Function content: %s", content))
	env := inputs.(map[string]interface{})

	program, err := expr.Compile(content, expr.Env(env))
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to compile expression: %v", err))
		return nil, err
	}

	output, err := expr.Run(program, env)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to run expression: %v", err))
		return nil, err
	}

	return output.(map[string]interface{}), nil
}
