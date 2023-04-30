package funcs

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"database/sql"

	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

type Funcs struct {
	Fobj                types.Function
	DBTx                *sql.Tx
	SystemSession       map[string]interface{} // {sessionanme: value}
	UserSession         map[string]interface{} // {sessionanme: value}
	Externalinputs      map[string]interface{} // {sessionanme: value}
	Externaloutputs     map[string]interface{} // {sessionanme: value}
	FuncCachedVariables map[string]interface{}
}

func NewFuncs(dbTx *sql.Tx, fobj types.Function, systemSession, userSession, externalinputs, externaloutputs, funcCachedVariables map[string]interface{}) *Funcs {
	return &Funcs{
		Fobj:                fobj,
		DBTx:                dbTx,
		SystemSession:       systemSession,
		UserSession:         userSession,
		Externalinputs:      externalinputs,
		Externaloutputs:     externaloutputs,
		FuncCachedVariables: funcCachedVariables,
	}
}

func (f *Funcs) SetInputs() ([]string, []string, map[string]interface{}) {

	logger.Debug(fmt.Sprintf("Start process %s", reflect.ValueOf(f.SetInputs).Kind().String()))

	newinputs := make(map[string]interface{})
	namelist := make([]string, len(f.Fobj.Inputs))
	valuelist := make([]string, len(f.Fobj.Inputs))

	inputs := f.Fobj.Inputs

	logger.Debug(fmt.Sprintf("function inputs: %s", logger.ConvertJson(inputs)))

	for i := 0; i < len(inputs); i++ {
		switch inputs[i].Source {
		case types.Fromsyssession:
			if f.SystemSession[inputs[i].Aliasname] != nil {
				inputs[i].Value = f.SystemSession[inputs[i].Aliasname].(string)
			} else {
				inputs[i].Value = inputs[i].Defaultvalue
			}
		case types.Fromusersession:
			if f.UserSession[inputs[i].Aliasname] != nil {
				inputs[i].Value = f.SystemSession[inputs[i].Aliasname].(string)
			} else {
				inputs[i].Value = inputs[i].Defaultvalue
			}
		case types.Prefunction:
			arr := strings.Split(inputs[i].Aliasname, ".")
			if len(arr) == 2 {

				if f.FuncCachedVariables[arr[0]] != nil {
					tempobj := f.FuncCachedVariables[arr[0]].(map[string]interface{})
					if tempobj[arr[1]] != nil {
						inputs[i].Value = tempobj[arr[1]].(string)
					}
				}
			} else {
				inputs[i].Value = inputs[i].Defaultvalue
			}

		case types.Fromexternal:
			if f.Externalinputs[inputs[i].Aliasname] != nil {
				inputs[i].Value = f.Externalinputs[inputs[i].Aliasname].(string)
			} else {
				inputs[i].Value = inputs[i].Defaultvalue
			}
		}
		newinputs[inputs[i].Name] = inputs[i].Value
		namelist[i] = inputs[i].Name
		valuelist[i] = inputs[i].Value
	}

	logger.Debug(fmt.Sprintf("function mapped inputs: %s", logger.ConvertJson(inputs)))

	f.Fobj.Inputs = inputs

	return namelist, valuelist, newinputs
}

func (f *Funcs) SetOutputs(outputs map[string]interface{}) {
	logger.Debug(fmt.Sprintf("function's ouputs: %s", logger.ConvertJson(outputs)))

	for i := 0; i < len(f.Fobj.Outputs); i++ {
		if outputs[f.Fobj.Outputs[i].Name] == nil {
			continue
		}

		for j := 0; j < len(f.Fobj.Outputs[i].Outputdest); j++ {

			switch f.Fobj.Outputs[i].Outputdest[j] {
			case types.Tosession:
				f.UserSession[f.Fobj.Outputs[i].Aliasname[j]] = outputs[f.Fobj.Outputs[i].Name]

			case types.Toexternal:
				f.Externaloutputs[f.Fobj.Outputs[i].Aliasname[j]] = outputs[f.Fobj.Outputs[i].Name]

			}
		}

		f.FuncCachedVariables[f.Fobj.Name] = outputs
	}
	logger.Debug(fmt.Sprintf("UserSession after function: %s", logger.ConvertJson(f.UserSession)))
	logger.Debug(fmt.Sprintf("Externaloutputs after function: %s", logger.ConvertJson(f.Externaloutputs)))
	logger.Debug(fmt.Sprintf("FuncCachedVariables after function: %s", logger.ConvertJson(f.FuncCachedVariables[f.Fobj.Name])))
}

func (f *Funcs) ConvertfromBytes(bytesbuffer []byte) map[string]interface{} {
	temobj := make(map[string]interface{})
	for i := 0; i < len(f.Fobj.Outputs); i++ {
		temobj[f.Fobj.Outputs[i].Name] = f.Fobj.Outputs[i].Defaultvalue
	}

	err := json.Unmarshal(bytesbuffer, &temobj)
	if err != nil {
		log.Println("error:", err)
	}
	return temobj
}

func (f *Funcs) Execute() {

	switch f.Fobj.Functype {
	case types.InputMap:
		inputmapfuncs := InputMapFuncs{}
		inputmapfuncs.Execute(f)

	case types.Csharp:
		csharpfuncs := CSharpFuncs{}
		csharpfuncs.Execute(f)
	case types.Javascript:
		jsfuncs := JSFuncs{}
		jsfuncs.Execute(f)

	case types.Query:
		qfuncs := QueryFuncs{}
		qfuncs.Execute(f)

	case types.SubTranCode:
		stcfuncs := SubTranCodeFuncs{}
		stcfuncs.Execute(f)

	case types.StoreProcedure:

	}
}

func (f *Funcs) convertMap(m map[string][]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range m {
		var interfaceValue interface{}
		if len(value) == 1 {
			interfaceValue = value[0]
		} else {
			interfaceValue = value
		}
		result[key] = interfaceValue
	}
	return result
}

func (f *Funcs) revertMap(m map[string]interface{}) map[string][]interface{} {
	reverted := make(map[string][]interface{})

	for k, v := range m {
		switch t := v.(type) {
		case []interface{}:
			reverted[k] = t
		default:
			reverted[k] = []interface{}{v}
		}
	}

	return reverted
}
