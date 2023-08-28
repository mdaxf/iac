package callback

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

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

func ExecuteTranCode(key string, tcode string, inputs map[string]interface{}, ctx context.Context, ctxcancel context.CancelFunc, dbTx *sql.Tx) []interface{} {
	log := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "CallbackExecution"}
	log.Debug(fmt.Sprintf("CallBackFunc: %s with %s %s %s %s %s", key, tcode, inputs, ctx, ctxcancel, dbTx))
	log.Debug(fmt.Sprintf("callBackMap: %s", callBackMap))
	if callBack, ok := callBackMap[key]; ok {

		log.Debug(fmt.Sprintf("callBack: %s", callBack))

		in := make([]reflect.Value, 5)
		in[0] = reflect.ValueOf(tcode)
		if inputs == nil {
			inputs = map[string]interface{}{}
		}

		in[1] = reflect.ValueOf(inputs)

		if ctx == nil {
			ctx = context.Background()
		}
		in[2] = reflect.ValueOf(ctx)

		if ctxcancel == nil {
			ctxcancel = func() {}
		}
		in[3] = reflect.ValueOf(ctxcancel)

		in[4] = reflect.ValueOf(dbTx)

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