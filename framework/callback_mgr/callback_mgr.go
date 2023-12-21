package callback_mgr

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mdaxf/iac/logger"
)

var callBackMap map[string]interface{}

func init() {
	callBackMap = make(map[string]interface{})
}

func RegisterCallBack(key string, callBack interface{}) {
	log := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "CallbackRegister"}
	log.Debug(fmt.Sprintf("RegisterCallBack: %s with %v", key, callBack))
	callBackMap[key] = callBack

	log.Debug(fmt.Sprintf("callBackMap: %s", callBackMap))
}

func CallBackFunc(key string, args ...interface{}) []interface{} {
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
	log.Debug(fmt.Sprintf("callBackMap: %s", callBackMap))
	if callBack, ok := callBackMap[key]; ok {

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
			return nil
		}

		outList := funcValue.Call(in)
		result := make([]interface{}, len(outList))

		log.Debug(fmt.Sprintf("outList: %s", logger.ConvertJson(outList)))

		for i, out := range outList {
			result[i] = out.Interface()
		}
		log.Debug(fmt.Sprintf("result: %s", logger.ConvertJson(result)))
		return result
	} else {
		log.Error(fmt.Sprintf("callBack(%s) not found", key))
		return nil
	}
}
