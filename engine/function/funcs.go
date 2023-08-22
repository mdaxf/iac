package funcs

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

type Funcs struct {
	Fobj                types.Function
	DBTx                *sql.Tx
	Ctx                 context.Context
	CtxCancel           context.CancelFunc
	SystemSession       map[string]interface{} // {sessionanme: value}
	UserSession         map[string]interface{} // {sessionanme: value}
	Externalinputs      map[string]interface{} // {sessionanme: value}
	Externaloutputs     map[string]interface{} // {sessionanme: value}
	FuncCachedVariables map[string]interface{}
	iLog                logger.Log
}

func NewFuncs(dbTx *sql.Tx, fobj types.Function, systemSession, userSession, externalinputs, externaloutputs, funcCachedVariables map[string]interface{}, ctx context.Context, ctxcancel context.CancelFunc) *Funcs {
	log := logger.Log{}
	log.ModuleName = logger.TranCode
	log.ControllerName = "Function"
	if systemSession["User"] != nil {
		log.User = systemSession["User"].(string)
	} else {
		log.User = "System"
	}

	return &Funcs{
		Fobj:                fobj,
		DBTx:                dbTx,
		Ctx:                 ctx,
		CtxCancel:           ctxcancel,
		SystemSession:       systemSession,
		UserSession:         userSession,
		Externalinputs:      externalinputs,
		Externaloutputs:     externaloutputs,
		FuncCachedVariables: funcCachedVariables,
		iLog:                log,
	}
}

func (f *Funcs) SetInputs() ([]string, []string, map[string]interface{}) {

	f.iLog.Debug(fmt.Sprintf("Start process %s", reflect.ValueOf(f.SetInputs).Kind().String()))

	newinputs := make(map[string]interface{})
	namelist := make([]string, len(f.Fobj.Inputs))
	valuelist := make([]string, len(f.Fobj.Inputs))

	inputs := f.Fobj.Inputs

	f.iLog.Debug(fmt.Sprintf("function inputs: %s", logger.ConvertJson(inputs)))

	for i := 0; i < len(inputs); i++ {
		f.iLog.Debug(fmt.Sprintf("function input: %s, Source: %s", logger.ConvertJson(inputs[i]), inputs[i].Source))
		switch inputs[i].Source {
		case types.Fromsyssession:

			if f.SystemSession[inputs[i].Aliasname] != nil {
				inputs[i].Value = f.SystemSession[inputs[i].Aliasname].(string)
			} else {

				f.iLog.Error(fmt.Sprintf("Error in SetInputs: %s SystemSession[%s]", "System Session not found", inputs[i].Aliasname))
				f.DBTx.Rollback()
			}
		case types.Fromusersession:
			if f.UserSession[inputs[i].Aliasname] != nil {
				inputs[i].Value = f.UserSession[inputs[i].Aliasname].(string)
			} else {
				f.iLog.Error(fmt.Sprintf("Error in SetInputs: %s UserSession[%s]", "User Session not found", inputs[i].Aliasname))
				f.DBTx.Rollback()
			}
		case types.Prefunction:
			arr := strings.Split(inputs[i].Aliasname, ".")
			f.iLog.Debug(fmt.Sprintf("Prefunction: %s", logger.ConvertJson(arr)))
			if len(arr) == 2 {
				f.iLog.Debug(fmt.Sprintf("Prefunction variables: %s", logger.ConvertJson(f.FuncCachedVariables)))
				if f.FuncCachedVariables[arr[0]] != nil {
					value, _ := f.checkinputvalue(arr[1], f.FuncCachedVariables[arr[0]].(map[string]interface{}))
					inputs[i].Value = value
				} else {
					f.iLog.Error(fmt.Sprintf("Error in SetInputs: %s Prefunction[%s]", "Prefunction not found", inputs[i].Aliasname))
					//	f.DBTx.Rollback()
					inputs[i].Value = ""
				}
			} else {
				f.iLog.Error(fmt.Sprintf("Error in SetInputs: %s Prefunction[%s]", "Prefunction not found", inputs[i].Aliasname))
				//	f.DBTx.Rollback()
				inputs[i].Value = ""
			}

		case types.Fromexternal:
			f.iLog.Debug(fmt.Sprintf("Externalinputs: %s", logger.ConvertJson(f.Externalinputs)))
			value, _ := f.checkinputvalue(inputs[i].Aliasname, f.Externalinputs)
			inputs[i].Value = value

		}

		switch inputs[i].Datatype {

		case types.Integer:
			if inputs[i].List == false {

				temp := f.ConverttoInt(inputs[i].Value)
				newinputs[inputs[i].Name] = temp

			}
			if inputs[i].List == true {

				if isArray(inputs[i].Value) {
					var temp []int
					err := json.Unmarshal([]byte(inputs[i].Value), &temp)
					if err != nil {
						f.iLog.Error(fmt.Sprintf("Unmarshal %s error: %s", inputs[i].Value, err.Error()))
						//	f.DBTx.Rollback()
					}
					newinputs[inputs[i].Name] = temp
				} else {
					temp := make([]int, 0)
					temp = append(temp, f.ConverttoInt(inputs[i].Value))
					newinputs[inputs[i].Name] = temp
				}

			}
		case types.Float:
			if inputs[i].List == false {

				temp := f.ConverttoFloat(inputs[i].Value)
				newinputs[inputs[i].Name] = temp

			}
			if inputs[i].List == true {

				var temp []float64
				err := json.Unmarshal([]byte(inputs[i].Value), &temp)
				if err != nil {
					f.iLog.Error(fmt.Sprintf("Unmarshal %s error: %s", inputs[i].Value, err.Error()))
					//	f.DBTx.Rollback()
				}
				newinputs[inputs[i].Name] = temp

			}
		case types.Bool:
			if inputs[i].List == false {
				temp := f.ConverttoBool(inputs[i].Value)
				newinputs[inputs[i].Name] = temp
			}
			if inputs[i].List == true {

				var temp []bool
				err := json.Unmarshal([]byte(inputs[i].Value), &temp)
				if err != nil {
					f.iLog.Error(fmt.Sprintf("Unmarshal %s error: %s", inputs[i].Value, err.Error()))
					//	f.DBTx.Rollback()
				}
				newinputs[inputs[i].Name] = temp

			}
		case types.DateTime:
			if inputs[i].List == false {

				temp := f.ConverttoDateTime(inputs[i].Value)
				newinputs[inputs[i].Name] = temp

			}
			if inputs[i].List == true {

				var temp []time.Time
				err := json.Unmarshal([]byte(inputs[i].Value), &temp)
				if err != nil {
					f.iLog.Error(fmt.Sprintf("Unmarshal %s error: %s", inputs[i].Value, err.Error()))
					//	f.DBTx.Rollback()
				}
				newinputs[inputs[i].Name] = temp

			}
		default:
			if inputs[i].List == false {
				newinputs[inputs[i].Name] = inputs[i].Value
			}
			if inputs[i].List == true {

				var temp []string
				err := json.Unmarshal([]byte(inputs[i].Value), &temp)
				if err != nil {
					f.iLog.Error(fmt.Sprintf("Unmarshal %s error: %s", inputs[i].Value, err.Error()))
					//	f.DBTx.Rollback()
				}
				newinputs[inputs[i].Name] = temp

			}

		}
		//	newinputs[inputs[i].Name] = inputs[i].Value
		namelist[i] = inputs[i].Name
		valuelist[i] = inputs[i].Value
	}

	f.iLog.Debug(fmt.Sprintf("function mapped inputs: %s", logger.ConvertJson(newinputs)))

	f.Fobj.Inputs = inputs

	return namelist, valuelist, newinputs
}

func isArray(value interface{}) bool {
	// Use reflection to check if the value's kind is an array
	val := reflect.ValueOf(value)
	return val.Kind() == reflect.Array || val.Kind() == reflect.Slice
}

func (f *Funcs) checkinputvalue(Aliasname string, variables map[string]interface{}) (string, error) {
	var err error
	err = nil
	var Result string
	if variables[Aliasname] != nil {
		if resultMap, ok := variables[Aliasname].([]interface{}); ok {
			Result = "["
			for index, value := range resultMap {
				_value, err := json.Marshal(value)
				if err != nil {
					f.iLog.Error(fmt.Sprintf("Marshal %s error: %s", Aliasname, err.Error()))
					f.DBTx.Rollback()
				}

				if index == 0 {
					Result += string(_value)
				} else {
					Result = Result + "," + string(_value)
				}
			}
			Result += "]"
		} else if resultMap, ok := variables[Aliasname].(map[string]interface{}); ok {
			value, err := json.Marshal(resultMap)
			if err != nil {
				f.iLog.Error(fmt.Sprintf("Marshal %s error: %s", Aliasname, err.Error()))
				//	f.DBTx.Rollback()
			} else {
				Result = string(value)
			}

			Result = string(value)
		} else if resultMap, ok := variables[Aliasname].([]string); ok {
			Result = ""
			for index, value := range resultMap {
				if index == 0 {
					Result = value
				} else {
					Result = Result + "," + value
				}
			}
		} else {

			/*temp, err := json.Marshal(variables[Aliasname])
			if err != nil {
				f.iLog.Error(fmt.Sprintf("Marshal %s error: %s", Aliasname, err.Error()))
			}
			f.iLog.Debug(fmt.Sprintf("checkinputvalue %s temp: %s", variables[Aliasname], temp))  */
			switch variables[Aliasname].(type) {
			case int:
				Result = strconv.Itoa(variables[Aliasname].(int))
			case int64:
				value := int(variables[Aliasname].(int64))
				Result = strconv.Itoa(value)
			case float64:
				Result = strconv.FormatFloat(variables[Aliasname].(float64), 'f', -1, 64)
			case bool:
				Result = strconv.FormatBool(variables[Aliasname].(bool))
			case string:
				Result = variables[Aliasname].(string)
			case time.Time:
				Result = variables[Aliasname].(time.Time).Format(types.DateTimeFormat)
			case nil:
				Result = ""
			default:
				Result = fmt.Sprint(variables[Aliasname])

			}
		}
	} else {
		f.iLog.Error(fmt.Sprintf("Error in SetInputs: %s Externalinputs[%s]", "Externalinputs not found", Aliasname))
		//	f.DBTx.Rollback()

	}
	f.iLog.Debug(fmt.Sprintf("checkinputvalue %s result: %s", variables[Aliasname], Result))
	return Result, err
}

func customMarshal(v interface{}) string {
	var jsonStr string
	switch reflect.TypeOf(v).Kind() {
	case reflect.Struct:
		t := reflect.TypeOf(v)
		v := reflect.ValueOf(v)
		jsonStr += "{"
		for i := 0; i < t.NumField(); i++ {
			fieldName := t.Field(i).Tag.Get("json")
			fieldValue := fmt.Sprintf("%v", v.Field(i))
			jsonStr += fmt.Sprintf(`%s:%s,`, fieldName, fieldValue)
		}
		jsonStr = jsonStr[:len(jsonStr)-1] // Remove trailing comma
		jsonStr += "}"
	}
	return jsonStr
}

func (f *Funcs) ConverttoInt(str string) int {
	temp, err := strconv.Atoi(str)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Convert %s to int error: %s", str, err.Error()))
	}

	return temp
}

func (f *Funcs) ConverttoFloat(str string) float64 {
	temp, err := strconv.ParseFloat(str, 64)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Convert %s to float error: %s", str, err.Error()))
	}

	return temp
}

func (f *Funcs) ConverttoBool(str string) bool {
	temp, err := strconv.ParseBool(str)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Convert %s to bool error: %s", str, err.Error()))
	}

	return temp
}

func (f *Funcs) ConverttoDateTime(str string) time.Time {
	temp, err := time.Parse(types.DateTimeFormat, str)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Convert %s to time error: %s", str, err.Error()))
	}

	return temp
}

func (f *Funcs) SetOutputs(outputs map[string]interface{}) {

	f.iLog.Debug(fmt.Sprintf("function's ouputs: %s", logger.ConvertJson(outputs)))

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
	f.iLog.Debug(fmt.Sprintf("UserSession after function: %s", logger.ConvertJson(f.UserSession)))
	f.iLog.Debug(fmt.Sprintf("Externaloutputs after function: %s", logger.ConvertJson(f.Externaloutputs)))
	f.iLog.Debug(fmt.Sprintf("FuncCachedVariables after function: %s", logger.ConvertJson(f.FuncCachedVariables[f.Fobj.Name])))
}

func (f *Funcs) ConvertfromBytes(bytesbuffer []byte) map[string]interface{} {

	f.iLog.Debug(fmt.Sprintf("Start process %s", reflect.ValueOf(f.ConvertfromBytes).Kind().String()))

	temobj := make(map[string]interface{})
	for i := 0; i < len(f.Fobj.Outputs); i++ {
		temobj[f.Fobj.Outputs[i].Name] = f.Fobj.Outputs[i].Defaultvalue
	}

	err := json.Unmarshal(bytesbuffer, &temobj)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("error:", err.Error()))
	}
	return temobj
}

func (f *Funcs) Execute() {

	f.iLog.Debug(fmt.Sprintf("Execute function: %s", f.Fobj.Name))
	//f.iLog.Debug(fmt.Sprintf("Start process %s", reflect.ValueOf(f.Execute).Kind().String()))

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
		spfuncs := StoreProcFuncs{}
		spfuncs.Execute(f)

	case types.TableInsert:
		ti := TableInsertFuncs{}
		ti.Execute(f)

	case types.TableUpdate:
		tu := TableUpdateFuncs{}
		tu.Execute(f)

	case types.TableDelete:
		td := TableDeleteFuncs{}
		td.Execute(f)

	case types.CollectionInsert:
		ci := CollectionInsertFuncs{}
		ci.Execute(f)

	case types.CollectionUpdate:
		cu := CollectionUpdateFuncs{}
		cu.Execute(f)

	case types.CollectionDelete:
		cd := CollectionDeleteFuncs{}
		cd.Execute(f)

	case types.ThrowError:
		te := ThrowErrorFuncs{}
		te.Execute(f)

	case types.SendMessage:
		sm := SendMessageFuncs{}
		sm.Execute(f)

	case types.SendEmail:
		se := EmailFuncs{}
		se.Execute(f)
	}
}

func (f *Funcs) Validate() (bool, error) {

	return true, nil
}

func (f *Funcs) CancelExecution(errormessage string) {

	f.iLog.Error(fmt.Sprintf("There is error during functoin %s execution: %s", f.Fobj.Name, errormessage))
	f.DBTx.Rollback()
	f.CtxCancel()
	f.Ctx.Done()
	return
}

func (f *Funcs) convertMap(m map[string][]interface{}) map[string]interface{} {

	f.iLog.Debug(fmt.Sprintf("Convert Map to Json Objects: %s", m))

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

	f.iLog.Debug(fmt.Sprintf("Convert Map to Json Objects result: %s", result))
	return result
}

func (f *Funcs) revertMap(m map[string]interface{}) map[string][]interface{} {

	f.iLog.Debug(fmt.Sprintf("Revert Json Objects to array: %s", m))
	reverted := make(map[string][]interface{})

	for k, v := range m {
		switch t := v.(type) {
		case []interface{}:
			reverted[k] = t
		default:
			reverted[k] = []interface{}{v}
		}
	}
	f.iLog.Debug(fmt.Sprintf("Revert Json Objects to array result: %s", reverted))

	return reverted
}
