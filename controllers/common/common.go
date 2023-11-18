package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/logger"
)

func GetRequestBody(ctx *gin.Context) ([]byte, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetRequestBody"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.Performance(fmt.Sprintf(" %s elapsed time: %v", "controllers.common.GetRequestBody", elapsed))
	}()

	iLog.Debug(fmt.Sprintf("GetRequestBody"))

	body, err := ioutil.ReadAll(ctx.Request.Body)

	if err != nil {
		iLog.Error(fmt.Sprintf("GetRequestBody error: %s", err.Error()))
		return nil, err
	}
	iLog.Debug(fmt.Sprintf("GetRequestBody body: %s", body))
	return body, nil
}

func GetRequestBodybyJson(ctx *gin.Context) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetRequestBodybyJson"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.Performance(fmt.Sprintf(" %s elapsed time: %v", "controllers.common.GetRequestBodybyJson", elapsed))
	}()
	iLog.Debug(fmt.Sprintf("GetRequestBodybyJson"))

	var request map[string]interface{}
	err := json.NewDecoder(ctx.Request.Body).Decode(&request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to decode request body: %v", err))
		return nil, err
	}
	iLog.Debug(fmt.Sprintf("GetRequestBodybyJson request: %s", request))
	return request, nil
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
