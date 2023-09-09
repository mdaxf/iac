package common

import (
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/logger"
)

func GetRequestBody(ctx *gin.Context) ([]byte, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetRequestBody"}
	iLog.Debug(fmt.Sprintf("GetRequestBody"))

	body, err := ioutil.ReadAll(ctx.Request.Body)

	if err != nil {
		iLog.Error(fmt.Sprintf("GetRequestBody error: %s", err.Error()))
		return nil, err
	}
	iLog.Debug(fmt.Sprintf("GetRequestBody body: %s", body))
	return body, nil
}

func mapToStruct(data map[string]interface{}, outStruct interface{}) error {
	// Get the type of the struct
	structType := reflect.TypeOf(outStruct)

	// Make sure outStruct is a pointer to a struct
	if structType.Kind() != reflect.Ptr || structType.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("outStruct must be a pointer to a struct")
	}

	// Get the value of the struct
	structValue := reflect.ValueOf(outStruct).Elem()

	// Iterate through the map and set struct fields
	for key, value := range data {
		structField := structValue.FieldByName(key)

		if !structField.IsValid() {
			continue
			//return fmt.Errorf("field %s does not exist in the struct", key)
		}

		if !structField.CanSet() {
			continue
			//return fmt.Errorf("field %s cannot be set", key)
		}

		fieldValue := reflect.ValueOf(value)
		if !fieldValue.Type().AssignableTo(structField.Type()) {
			//return fmt.Errorf("field %s type mismatch", key)
			continue
		}

		structField.Set(fieldValue)
	}

	return nil
}
