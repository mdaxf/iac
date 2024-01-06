package callback_mgr

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mdaxf/iac/logger"
)

var CallBackMap map[string]interface{}

func init() {
	CallBackMap = make(map[string]interface{})
}

func RegisterCallBack(key string, callBack interface{}) {
	log := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "CallbackRegister"}
	log.Debug(fmt.Sprintf("RegisterCallBack: %s with %v", key, callBack))
	CallBackMap[key] = callBack

	log.Debug(fmt.Sprintf("callBackMap: %s", CallBackMap))
}

func CallBackFunc(key string, args ...interface{}) ([]interface{}, error) {
	log := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "CallbackExecution"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("CallbackExecution", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			log.Error(fmt.Sprintf("CallbackExecution error: %v", r))
			return

		}
	}()

	log.Debug(fmt.Sprintf("CallBackFunc: %s with %v", key, args))
	log.Debug(fmt.Sprintf("callBackMap: %s", CallBackMap))
	if callBack, ok := CallBackMap[key]; ok {

		log.Debug(fmt.Sprintf("callBack: %s", callBack))

		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			if arg == nil {
				arg = reflect.ValueOf("")
			}

			in[i] = reflect.ValueOf(arg)
		}
		log.Debug(fmt.Sprintf("in: %s", in))

		funcValue := reflect.ValueOf(callBack)
		log.Debug(fmt.Sprintf("funcValue: %s", funcValue))

		if funcValue.Kind() != reflect.Func {

			log.Error(fmt.Sprintf("callBack(%s) is not a function", key))
			return nil, fmt.Errorf(fmt.Sprintf("callBack(%s) is not a function", key))
		}

		outList := funcValue.Call(in)
		result := make([]interface{}, len(outList))

		log.Debug(fmt.Sprintf("outList: %s", logger.ConvertJson(outList)))

		for i, out := range outList {
			result[i] = out.Interface()
		}
		log.Debug(fmt.Sprintf("result: %s", logger.ConvertJson(result)))
		return result, nil
	} else {
		log.Error(fmt.Sprintf("callBack(%s) not found", key))
		return nil, fmt.Errorf(fmt.Sprintf("callBack(%s) not found", key))
	}
}

// convertSliceToMap converts a slice of interfaces into a map[string]interface{}.
// It iterates over the slice, treating every even-indexed element as the key and the following odd-indexed element as the value.
// If the key is a string, it adds the key-value pair to the resulting map.
// The function returns the resulting map.
func ConvertSliceToMap(slice []interface{}) map[string]interface{} {
	resultMap := make(map[string]interface{})

	for i := 0; i < len(slice); i += 2 {
		key, ok1 := slice[i].(string)
		value := slice[i+1]

		if ok1 {
			resultMap[key] = value
		}
	}

	return resultMap
}
