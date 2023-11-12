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

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/signalrsrv/signalr"
)

type Funcs struct {
	Fobj                 types.Function
	DBTx                 *sql.Tx
	Ctx                  context.Context
	CtxCancel            context.CancelFunc
	SystemSession        map[string]interface{} // {sessionanme: value}
	UserSession          map[string]interface{} // {sessionanme: value}
	Externalinputs       map[string]interface{} // {sessionanme: value}
	Externaloutputs      map[string]interface{} // {sessionanme: value}
	FuncCachedVariables  map[string]interface{}
	iLog                 logger.Log
	FunctionInputs       []map[string]interface{}
	FunctionOutputs      []map[string]interface{}
	ExecutionNumber      int
	ExecutionCount       int
	FunctionMappedInputs map[string]interface{}
	DocDBCon             *documents.DocDB
	SignalRClient        signalr.Client
	ErrorMessage         string
}

func NewFuncs(DocDBCon *documents.DocDB, SignalRClient signalr.Client, dbTx *sql.Tx, fobj types.Function, systemSession, userSession, externalinputs, externaloutputs, funcCachedVariables map[string]interface{}, ctx context.Context, ctxcancel context.CancelFunc) *Funcs {
	log := logger.Log{}
	log.ModuleName = logger.TranCode
	log.ControllerName = "Function"
	if systemSession["User"] != nil {
		log.User = systemSession["User"].(string)
	} else {
		log.User = "System"
	}
	var newdata []map[string]interface{}

	systemSession["UTCTime"] = time.Now().UTC()
	systemSession["LocalTime"] = time.Now()
	if systemSession["UserNo"] == nil {
		systemSession["UserNo"] = "System"
	}
	if systemSession["UserID"] == nil {
		systemSession["UserID"] = 0
	}

	if systemSession["WorkSpace"] == nil {
		systemSession["WorkSpace"] = ""
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
		FunctionInputs:      newdata,
		FunctionOutputs:     newdata,
		ExecutionNumber:     1,
		ExecutionCount:      0,
		DocDBCon:            DocDBCon,
		SignalRClient:       SignalRClient,
		ErrorMessage:        "",
	}
}

func (f *Funcs) HandleInputs() ([]string, []string, map[string]interface{}, error) {
	f.iLog.Debug(fmt.Sprintf("Start process %s", reflect.ValueOf(f.HandleInputs).Kind().String()))
	f.iLog.Debug(fmt.Sprintf("function inputs: %s", logger.ConvertJson(f.Fobj.Inputs)))

	newinputs := make(map[string]interface{})
	namelist := make([]string, len(f.Fobj.Inputs))
	valuelist := make([]string, len(f.Fobj.Inputs))

	inputs := f.Fobj.Inputs
	//	newinputs = f.FunctionMappedInputs

	for i, _ := range inputs {
		//	f.iLog.Debug(fmt.Sprintf("function input: %s, Source: %s", logger.ConvertJson(inputs[i]), inputs[i].Source))
		switch inputs[i].Source {
		case types.Fromsyssession:

			if f.SystemSession[inputs[i].Aliasname] != nil {
				inputs[i].Value = f.SystemSession[inputs[i].Aliasname].(string)
			} else {

				f.iLog.Error(fmt.Sprintf("Error in SetInputs: %s SystemSession[%s]", "System Session not found", inputs[i].Aliasname))
				//	f.DBTx.Rollback()
			}
		case types.Fromusersession:
			if f.UserSession[inputs[i].Aliasname] != nil {
				inputs[i].Value = f.UserSession[inputs[i].Aliasname].(string)
			} else {
				f.iLog.Error(fmt.Sprintf("Error in SetInputs: %s UserSession[%s]", "User Session not found", inputs[i].Aliasname))
				//	f.DBTx.Rollback()
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
					if inputs[i].Datatype == types.DateTime {
						if inputs[i].Defaultvalue == "" {
							inputs[i].Value = time.Now().Format(types.DateTimeFormat)
						} else {
							inputs[i].Value = inputs[i].Defaultvalue
						}

					} else if inputs[i].Datatype == types.Integer {
						if inputs[i].Defaultvalue == "" {
							inputs[i].Value = "0"
						} else {
							inputs[i].Value = inputs[i].Defaultvalue
						}
					} else if inputs[i].Datatype == types.Float {
						if inputs[i].Defaultvalue == "" {
							inputs[i].Value = "0.0"
						} else {
							inputs[i].Value = inputs[i].Defaultvalue
						}

					} else if inputs[i].Datatype == types.Bool {
						if inputs[i].Defaultvalue == "" {
							inputs[i].Value = "false"
						} else {
							inputs[i].Value = inputs[i].Defaultvalue
						}

					} else {

						inputs[i].Value = inputs[i].Defaultvalue
					}
				}
			} else {
				f.iLog.Error(fmt.Sprintf("Error in SetInputs: %s Prefunction[%s]", "Prefunction not found", inputs[i].Aliasname))
				//	f.DBTx.Rollback()
				if inputs[i].Datatype == types.DateTime {
					if inputs[i].Defaultvalue == "" {
						inputs[i].Value = time.Now().Format(types.DateTimeFormat)
					} else {
						inputs[i].Value = inputs[i].Defaultvalue
					}

				} else if inputs[i].Datatype == types.Integer {
					if inputs[i].Defaultvalue == "" {
						inputs[i].Value = "0"
					} else {
						inputs[i].Value = inputs[i].Defaultvalue
					}
				} else if inputs[i].Datatype == types.Float {
					if inputs[i].Defaultvalue == "" {
						inputs[i].Value = "0.0"
					} else {
						inputs[i].Value = inputs[i].Defaultvalue
					}

				} else if inputs[i].Datatype == types.Bool {
					if inputs[i].Defaultvalue == "" {
						inputs[i].Value = "false"
					} else {
						inputs[i].Value = inputs[i].Defaultvalue
					}

				} else {

					inputs[i].Value = inputs[i].Defaultvalue
				}
			}

		case types.Fromexternal:
			f.iLog.Debug(fmt.Sprintf("Externalinputs: %s", logger.ConvertJson(f.Externalinputs)))
			value, _ := f.checkinputvalue(inputs[i].Aliasname, f.Externalinputs)
			inputs[i].Value = value

		}

		f.iLog.Debug(fmt.Sprintf("function input: %s, Source: %v, value: %s type: %d", logger.ConvertJson(inputs[i]), inputs[i].Source, inputs[i].Value, inputs[i].Datatype))
		switch inputs[i].Datatype {

		case types.Integer:

			if inputs[i].List == false {

				temp := f.ConverttoInt(inputs[i].Value)
				newinputs[inputs[i].Name] = temp

			}
			if inputs[i].List == true {

				f.iLog.Debug(fmt.Sprintf("check value %s if it isArray: %s", inputs[i].Value, reflect.ValueOf(inputs[i].Value).Kind().String()))

				var tempstr []string
				err := json.Unmarshal([]byte(inputs[i].Value), &tempstr)
				f.iLog.Debug(fmt.Sprintf("Unmarshal %s result: %s length:%d", inputs[i].Value, logger.ConvertJson(tempstr), len(tempstr)))

				if err != nil {
					//f.iLog.Error(fmt.Sprintf("Unmarshal %s error: %s", inputs[i].Value, err.Error()))

					var tempint []int
					err := json.Unmarshal([]byte(inputs[i].Value), &tempint)
					if err != nil {
						//f.iLog.Error(fmt.Sprintf("Unmarshal %s error: %s", inputs[i].Value, err.Error()))

						var tempfloat []float64
						err := json.Unmarshal([]byte(inputs[i].Value), &tempfloat)
						if err != nil {
							//f.iLog.Error(fmt.Sprintf("Unmarshal %s error: %s", inputs[i].Value, err.Error()))
							//	f.DBTx.Rollback()

							str := inputs[i].Value
							str = strings.TrimPrefix(str, "[")
							str = strings.TrimSuffix(str, "]")
							strList := strings.Split(str, ",")

							intList := make([]int, len(strList))
							for index, str := range strList {
								intList[index] = f.ConverttoInt(str)
							}
							newinputs[inputs[i].Name] = intList

						} else {
							temp := make([]int, 0)
							for _, str := range tempfloat {
								temp = append(temp, int(str))
							}
							newinputs[inputs[i].Name] = temp
						}

					} else {
						newinputs[inputs[i].Name] = tempint
					}
				} else {
					temp := make([]int, 0)
					for _, str := range tempstr {
						temp = append(temp, f.ConverttoInt(str))
					}
					newinputs[inputs[i].Name] = temp

				}
				/*
					if isArray(inputs[i].Value) {



					} else {
						temp := make([]int, 0)
						temp = append(temp, f.ConverttoInt(inputs[i].Value))
						newinputs[inputs[i].Name] = temp
					} */

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
				var datetimeStrings []string
				err := json.Unmarshal([]byte(inputs[i].Value), &datetimeStrings)
				if err != nil {
					f.iLog.Error(fmt.Sprintf("Unmarshal %s error: %s", inputs[i].Value, err.Error()))
					//	f.DBTx.Rollback()
				}
				for _, dtStr := range datetimeStrings {
					dt, err := time.Parse("2006-01-02 15:04:05", dtStr)
					if err != nil {
						f.iLog.Error(fmt.Sprintf("Parse %s error: %s", dtStr, err.Error()))
						//	f.DBTx.Rollback()
					}
					temp = append(temp, dt)
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
		namelist[i] = inputs[i].Name
		valuelist[i] = inputs[i].Value
	}

	f.FunctionMappedInputs = newinputs

	return namelist, valuelist, newinputs, nil
}

func (f *Funcs) SetInputs() ([]string, []string, map[string]interface{}) {

	f.iLog.Debug(fmt.Sprintf("Start process %s", reflect.ValueOf(f.SetInputs).Kind().String()))

	namelist, valuelist, newinputs, err := f.HandleInputs()

	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in SetInputs: %s", err.Error()))
	}
	inputs := f.Fobj.Inputs

	if f.ExecutionNumber > 1 {
		f.iLog.Debug(fmt.Sprintf("function inputs: %s execution count: %d/%d", logger.ConvertJson(inputs), f.ExecutionCount, f.ExecutionNumber))

		f.iLog.Debug(fmt.Sprintf("function mapped inputs: %s", logger.ConvertJson(f.FunctionMappedInputs)))
		for i := 0; i < len(inputs); i++ {

			//	newinputs[inputs[i].Name] = inputs[i].Value
			namelist[i] = inputs[i].Name
			f.iLog.Debug(fmt.Sprintf("function input: %s, Source: %d", logger.ConvertJson(inputs[i]), inputs[i].Source))
			if f.ExecutionNumber > 1 && inputs[i].Repeat && inputs[i].List {

				//valuelist[i] = (f.FunctionMappedInputs[inputs[i].Name]).([]interface{})[f.ExecutionCount].(string)
				if inputs[i].Datatype == types.DateTime {
					temp := (f.FunctionMappedInputs[inputs[i].Name]).([]time.Time)[f.ExecutionCount]
					newinputs[inputs[i].Name] = temp.Format("2006-01-02 15:04:05")
					valuelist[i] = temp.Format("2006-01-02 15:04:05")
				} else if inputs[i].Datatype == types.Integer {
					temp := (f.FunctionMappedInputs[inputs[i].Name]).([]int)[f.ExecutionCount]
					newinputs[inputs[i].Name] = strconv.Itoa(temp)
					valuelist[i] = strconv.Itoa(temp)
				} else if inputs[i].Datatype == types.Float {
					temp := (f.FunctionMappedInputs[inputs[i].Name]).([]float64)[f.ExecutionCount]
					newinputs[inputs[i].Name] = strconv.FormatFloat(temp, 'f', -1, 64)
					valuelist[i] = strconv.FormatFloat(temp, 'f', -1, 64)
				} else if inputs[i].Datatype == types.Bool {
					temp := (f.FunctionMappedInputs[inputs[i].Name]).([]bool)[f.ExecutionCount]
					newinputs[inputs[i].Name] = strconv.FormatBool(temp)
					valuelist[i] = strconv.FormatBool(temp)
				} else {
					valuelist[i] = (f.FunctionMappedInputs[inputs[i].Name]).([]string)[f.ExecutionCount]
					newinputs[inputs[i].Name] = (f.FunctionMappedInputs[inputs[i].Name]).([]string)[f.ExecutionCount]
				}

			} else {
				newinputs[inputs[i].Name] = f.FunctionMappedInputs[inputs[i].Name]
				valuelist[i] = inputs[i].Value
			}

			//inputs[i].Value
		}

		f.iLog.Debug(fmt.Sprintf("function mapped inputs: %s for execution: %d/%d", logger.ConvertJson(newinputs), f.ExecutionCount, f.ExecutionNumber))
	}
	//f.Fobj.Inputs = inputs

	return namelist, valuelist, newinputs
}

func isArray(value interface{}) bool {
	// Use reflection to check if the value's kind is an array

	val := reflect.ValueOf(value)
	return val.Kind() == reflect.Array || val.Kind() == reflect.Slice
}

func (f *Funcs) checkifRepeatExecution() (int, error) {

	inputs := f.Fobj.Inputs
	lastcount := -2

	_, _, newinputs, err := f.HandleInputs()

	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in SetInputs: %s", err.Error()))
	}

	f.Fobj.Inputs = inputs
	f.iLog.Debug(fmt.Sprintf("parse inputs result: %s", logger.ConvertJson(newinputs)))
	for _, input := range inputs {

		f.iLog.Debug(fmt.Sprintf("checkifRepeatExecution input: %s", logger.ConvertJson(input)))

		count := -1
		if input.Repeat && input.List {
			inputvalue := newinputs[input.Name]
			if isArray(inputvalue) {
				if input.Datatype == types.DateTime {
					count = len(inputvalue.([]time.Time))
				} else if input.Datatype == types.Integer {
					count = len(inputvalue.([]int))
				} else if input.Datatype == types.Float {
					count = len(inputvalue.([]float64))
				} else if input.Datatype == types.Bool {
					count = len(inputvalue.([]bool))
				} else {
					count = len(inputvalue.([]string))
				}

			} else {
				count = 1
			}
			f.iLog.Debug(fmt.Sprintf("checkifRepeatExecution input.Value: %s, count: %d, lastcount %d", input.Value, count, lastcount))
			if lastcount == -2 {
				lastcount = count
			} else if lastcount != count {
				return -1, nil
			}

		}
	}

	if lastcount > 0 {
		return lastcount, nil
	} else {
		return 1, nil
	}
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
					//	f.DBTx.Rollback()
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
			Result = string(Result)
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

	f.FunctionOutputs = append(f.FunctionOutputs, outputs)

}

func (f *Funcs) SetfuncOutputs() {
	newoutputs := make(map[string]interface{})
	if f.ExecutionNumber > 1 {
		for index, outputs := range f.FunctionOutputs {
			for key, value := range outputs {
				if newoutputs[key] == nil {
					newoutputs[key] = make([]interface{}, f.ExecutionNumber)
				}
				newoutputs[key].([]interface{})[index] = value
			}
		}

		f.SetfuncSingleOutputs(newoutputs)
	} else {
		f.SetfuncSingleOutputs(f.FunctionOutputs[0])
	}

}

func (f *Funcs) SetfuncSingleOutputs(outputs map[string]interface{}) {
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
		f.iLog.Error(fmt.Sprintf("error: %v", err.Error()))
	}
	return temobj
}

func (f *Funcs) Execute() {

	defer func() {
		if r := recover(); r != nil {
			f.iLog.Error(fmt.Sprintf("Panic in Execute: %s", r))
			f.CancelExecution(fmt.Sprintf("Panic in Execute: %s", r))
			f.ErrorMessage = fmt.Sprintf("Panic in Execute: %s", r)
			return
		}
	}()

	f.iLog.Debug(fmt.Sprintf("Execute function: %s", f.Fobj.Name))
	//f.iLog.Debug(fmt.Sprintf("Start process %s", reflect.ValueOf(f.Execute).Kind().String()))
	number, err := f.checkifRepeatExecution()

	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in Execute: %s", err.Error()))
	}
	if number < 1 {
		f.iLog.Debug(fmt.Sprintf("No repeat execution"))
		outputs := make(map[string]interface{})
		f.SetOutputs(outputs)
		return
	}
	f.iLog.Debug(fmt.Sprintf("Execute the function: %s, inputs: %s mapped inputs: %s", f.Fobj.Name, logger.ConvertJson(f.Fobj.Inputs), logger.ConvertJson(f.FunctionMappedInputs)))
	f.iLog.Debug(fmt.Sprintf("Repeat execution: %d", number))

	f.ExecutionNumber = number
	f.ExecutionCount = 0

	for i := 0; i < f.ExecutionNumber; i++ {
		f.iLog.Debug(fmt.Sprintf("Execute the function: %s, execution count: %d / %d", f.Fobj.Name, i+1, f.ExecutionNumber))
		switch f.Fobj.Functype {
		case types.InputMap:
			inputmapfuncs := InputMapFuncs{}
			inputmapfuncs.Execute(f)

		case types.GoExpr:
			goexprfuncs := GoExprFuncs{}
			goexprfuncs.Execute(f)
		case types.Javascript:
			jsfuncs := JSFuncs{}
			jsfuncs.Execute(f)

		case types.Query:
			qfuncs := QueryFuncs{}
			qfuncs.Execute(f)

		case types.SubTranCode:
			stcfuncs := SubTranCodeFuncs{}
			//stcfuncs := NewSubTran()
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

		f.iLog.Debug(fmt.Sprintf("executed function %s with outputs: %s", f.Fobj.Name, logger.ConvertJson(f.FunctionOutputs)))

		f.ExecutionCount = i + 1
	}
	f.SetfuncOutputs()
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
