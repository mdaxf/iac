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

// NewFuncs creates a new instance of the Funcs struct.
// It initializes the Funcs struct with the provided parameters and returns a pointer to the created instance.
// The Funcs struct represents a collection of functions and their associated data for execution.
// Parameters:
// - DocDBCon: A pointer to the DocDB connection object.
// - SignalRClient: The SignalR client object.
// - dbTx: A pointer to the SQL transaction object.
// - fobj: The function object.
// - systemSession: A map containing system session data.
// - userSession: A map containing user session data.
// - externalinputs: A map containing external input data.
// - externaloutputs: A map containing external output data.
// - funcCachedVariables: A map containing cached variables for the function.
// - ctx: The context object.
// - ctxcancel: The cancel function for the context.
// Returns:
// - A pointer to the created Funcs instance.

func NewFuncs(DocDBCon *documents.DocDB, SignalRClient signalr.Client, dbTx *sql.Tx, fobj types.Function, systemSession, userSession, externalinputs, externaloutputs, funcCachedVariables map[string]interface{}, ctx context.Context, ctxcancel context.CancelFunc) *Funcs {
	log := logger.Log{}
	log.ModuleName = logger.TranCode
	log.ControllerName = "Function"
	if systemSession["User"] != nil {
		log.User = systemSession["User"].(string)
	} else {
		log.User = "System"
	}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("engine.funcs.NewFuncs", elapsed)
	}()

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

// HandleInputs handles the inputs for the Funcs struct.
// It retrieves the inputs from various sources such as system session, user session, pre-function, and external inputs.
// The retrieved inputs are then converted to the appropriate data types and stored in the newinputs map.
// Finally, it returns the list of input names, input values, and the newinputs map.
// Returns:
// - A list of input names.
// - A list of input values.
// - A map containing the new inputs.
// - An error if there was an error in the process.

func (f *Funcs) HandleInputs() ([]string, []string, map[string]interface{}, error) {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.HandleInputs", elapsed)
		}()
	*/defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.HandleInputs with error: %s", err))
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.HandleInputs with error: %s", err))
			return
		}
	}()

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

// SetInputs sets the inputs for the Funcs object.
// It handles the mapping of inputs and prepares them for execution.
// It returns the list of input names, the list of input values, and a map of new inputs.

func (f *Funcs) SetInputs() ([]string, []string, map[string]interface{}) {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.SetInputs", elapsed)
		}()
	*/defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SetInputs with error: %s", err))
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.HandleInputs with error: %s", err))
			return
		}
	}()

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

// isArray checks if the given value is an array or a slice.
// It uses reflection to determine the kind of the value.
// Returns true if the value is an array or a slice, false otherwise.
func isArray(value interface{}) bool {
	// Use reflection to check if the value's kind is an array

	val := reflect.ValueOf(value)
	return val.Kind() == reflect.Array || val.Kind() == reflect.Slice
}

// checkifRepeatExecution checks if the function should be repeated execution based on the inputs.
// It returns the count of repetitions and an error if any.
// Returns:
// - The count of repetitions.
// - An error if there was an error in the process.

func (f *Funcs) checkifRepeatExecution() (int, error) {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.checkifRepeatExecution", elapsed)
		}()
		defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.checkifRepeatExecution with error: %s", err))
				return
			}
		}()
	*/
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

// checkinputvalue is a function that checks the input value for a given alias name and variables.
// It returns the result as a string and an error if any.
// Parameters:
// - Aliasname: The alias name of the input.
// - variables: A map containing the variables.
// Returns:
// - The result as a string.
// - An error if there was an error in the process.
func (f *Funcs) checkinputvalue(Aliasname string, variables map[string]interface{}) (string, error) {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.checkinputvalue", elapsed)
		}()
		defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.checkinputvalue with error: %s", err))
				return
			}
		}()  */

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

// customMarshal is a function that takes an interface{} as input and returns a string representation of the input object in JSON format.
// The function uses reflection to iterate over the fields of the input object and constructs a JSON string by concatenating the field names and values.
// The field names are obtained from the "json" tag of the struct fields.
// If the input object is not a struct, the function returns an empty string.
// Parameters:
// - v: The input object.
// Returns:
// - A string representation of the input object in JSON format.

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

// ConverttoInt converts a string to an integer.
// It measures the performance of the conversion and logs any errors that occur.
// If an error occurs during the conversion, it returns 0.
// Parameters:
// - str: The string to be converted.
// Returns:
// - The converted integer value.

func (f *Funcs) ConverttoInt(str string) int {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.ConverttoInt", elapsed)
		}()
		defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.ConverttoInt with error: %s", err))
				return
			}
		}()
	*/
	temp, err := strconv.Atoi(str)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Convert %s to int error: %s", str, err.Error()))
	}

	return temp
}

// ConverttoFloat converts a string to a float64 value.
// It uses the strconv.ParseFloat function to perform the conversion.
// If the conversion fails, an error is logged and the function returns 0.
// The performance of this function is logged using the provided iLog instance.
// Parameters:
// - str: The string to be converted.
// Returns:
// - The converted float64 value.

func (f *Funcs) ConverttoFloat(str string) float64 {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.ConverttoFloat", elapsed)
		}()
		defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.ConverttoFloat with error: %s", err))
				return
			}
		}()
	*/
	temp, err := strconv.ParseFloat(str, 64)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Convert %s to float error: %s", str, err.Error()))
	}

	return temp
}

// ConverttoBool converts a string to a boolean value.
// It uses the strconv.ParseBool function to parse the string.
// If the string cannot be parsed, an error is logged and false is returned.
// The performance of this function is logged using the iLog.PerformanceWithDuration method.
// If a panic occurs during the execution of this function, the panic is recovered and logged as an error.
// Parameters:
// - str: The string to be converted to a boolean value.
// Returns:
// - bool: The boolean value parsed from the string.

func (f *Funcs) ConverttoBool(str string) bool {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.ConverttoBool", elapsed)
		}()
		defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.ConverttoBool with error: %s", err))
				return
			}
		}()
	*/
	temp, err := strconv.ParseBool(str)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Convert %s to bool error: %s", str, err.Error()))
	}

	return temp
}

// ConverttoDateTime converts a string to a time.Time value.
// It uses the specified DateTimeFormat to parse the string.
// If the string cannot be parsed, an error is logged and the zero time value is returned.
// The function also logs the performance duration of the conversion.
// If a panic occurs during the execution of this function, the panic is recovered and logged as an error.
// Parameters:
// - str: The string to be converted to a time.Time value.
// Returns:
// - time.Time: The time.Time value parsed from the string.

func (f *Funcs) ConverttoDateTime(str string) time.Time {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.ConverttoDatTime", elapsed)
		}()
		defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.ConverttoDateTime with error: %s", err))
				return
			}
		}()
	*/
	temp, err := time.Parse(types.DateTimeFormat, str)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Convert %s to time error: %s", str, err.Error()))
	}

	return temp
}

// SetOutputs sets the outputs of the function.
// It takes a map of string keys and interface{} values representing the outputs.
// The function logs the performance duration and any errors that occur during execution.
// It also appends the outputs to the FunctionOutputs slice.
// Parameters:
// - outputs: A map of string keys and interface{} values representing the outputs.

func (f *Funcs) SetOutputs(outputs map[string]interface{}) {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.SetOutputs", elapsed)
		}()
		defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SetOutputs with error: %s", err))
				return
			}
		}()
	*/
	f.iLog.Debug(fmt.Sprintf("function's ouputs: %s", logger.ConvertJson(outputs)))

	f.FunctionOutputs = append(f.FunctionOutputs, outputs)

}

// SetfuncOutputs sets the function outputs for the Funcs struct.
// It combines the outputs from multiple function executions into a single output map.
// If the execution number is greater than 1, it iterates over the function outputs and combines them.
// Otherwise, it sets the single function output.
// The function also measures the performance duration and logs any errors that occur during execution.
// If a panic occurs during the execution of this function, the panic is recovered and logged as an error.
// Parameters:
// - outputs: A map of string keys and interface{} values representing the outputs.

func (f *Funcs) SetfuncOutputs() {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.SetfuncOutputs", elapsed)
		}()
		defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SetfuncOutputs with error: %s", err))
				return
			}
		}()
	*/
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

// SetfuncSingleOutputs sets the single outputs of the Funcs object based on the provided map of outputs.
// It iterates through the outputs of the Funcs object and assigns the corresponding values from the map to the appropriate destinations.
// The UserSession, Externaloutputs, and FuncCachedVariables are updated accordingly.
// If an error occurs during the process, it is recovered and logged.
// The performance duration of the function is also logged.
// Parameters:
// - outputs: A map of string keys and interface{} values representing the outputs.
// Returns:
// - An error if there was an error in the process.
func (f *Funcs) SetfuncSingleOutputs(outputs map[string]interface{}) {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.SetfuncSingleOutputs", elapsed)
		}()
		defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SetfuncSingleOutputs with error: %s", err))
				return
			}
		}()
	*/
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

// ConvertfromBytes converts a byte buffer into a map[string]interface{}.
// It takes a byte buffer as input and returns a map containing the converted data.
// The function also logs the performance duration and handles any panics that occur during execution.
// If there is an error during the JSON unmarshaling process, it is logged as well.

func (f *Funcs) ConvertfromBytes(bytesbuffer []byte) map[string]interface{} {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.ConvertfromBytes", elapsed)
		}()
		defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.ConvertfromBytes with error: %s", err))
				return
			}
		}()
	*/
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

// Execute executes the function.
// It measures the execution time, handles panics, and executes the appropriate function based on the function type.
// It also sets the function outputs and updates the execution count.
func (f *Funcs) Execute() {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.Execute", elapsed)
	}()

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
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.Validate", elapsed)
	}()
	/*	defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.Validate with error: %s", err))
				return
			}
		}()
	*/
	return true, nil
}

// CancelExecution cancels the execution of the function and rolls back any database transaction.
// It takes an error message as input and logs the error message along with the function name.
// It also cancels the context and returns.
func (f *Funcs) CancelExecution(errormessage string) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.CancelExecution", elapsed)
	}()
	/*	defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.CancelExecution with error: %s", err))
				return
			}
		}()
	*/
	f.iLog.Error(fmt.Sprintf("There is error during functoin %s execution: %s", f.Fobj.Name, errormessage))
	f.DBTx.Rollback()
	f.CtxCancel()
	f.Ctx.Done()
	return
}

// convertMap converts a map[string][]interface{} to a map[string]interface{}.
// It iterates over the input map and assigns the first value of each key's slice
// to the corresponding key in the output map. If a key's slice has more than one
// value, the entire slice is assigned to the key in the output map.
// The function also logs the performance duration and debug information.
// It returns the converted map[string]interface{}.

func (f *Funcs) convertMap(m map[string][]interface{}) map[string]interface{} {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.convertMap", elapsed)
		}()
		defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.convertMap with error: %s", err))
				return
			}
		}()
	*/
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

// revertMap reverts a map[string]interface{} to a map[string][]interface{}.
// It converts each value in the input map to an array, preserving the original key-value pairs.
// If the value is already an array, it is preserved as is. Otherwise, it is converted to a single-element array.
// The reverted map is returned as the result.
func (f *Funcs) revertMap(m map[string]interface{}) map[string][]interface{} {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			f.iLog.PerformanceWithDuration("engine.funcs.revertMap", elapsed)
		}()
		defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.revertMap with error: %s", err))
				return
			}
		}()
	*/
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
