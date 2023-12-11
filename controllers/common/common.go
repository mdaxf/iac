package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/framework/auth"
	"github.com/mdaxf/iac/logger"
)

func GetRequestUser(ctx *gin.Context) (string, string, string, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetRequestBody"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.common.GetRequestUser", elapsed)
	}()
	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("controllers.common.GetRequestUser error: %s", err))
		}
	}()

	iLog.Debug(fmt.Sprintf("GetRequestUser"))

	userid, user, clientid, err := auth.GetUserInformation(ctx)

	if err != nil {
		iLog.Error(fmt.Sprintf("GetRequestUser error: %s", err.Error()))
		return "", "", "", err
	}
	iLog.Debug(fmt.Sprintf("GetRequestUser userid: %s user: %s clientid: %s", userid, user, clientid))

	return userid, user, clientid, nil
}

func GetRequestBody(ctx *gin.Context) ([]byte, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetRequestBody"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.common.GetRequestBody", elapsed)
	}()

	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("controllers.common.GetRequestBody error: %s", err))
		}
	}()

	_, user, clientid, err := GetRequestUser(ctx)

	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("GetRequestBody"))

	body, err := ioutil.ReadAll(ctx.Request.Body)

	if err != nil {
		iLog.Error(fmt.Sprintf("GetRequestBody error: %s", err.Error()))
		return nil, err
	}
	iLog.Debug(fmt.Sprintf("GetRequestBody body: %s", body))
	return body, nil
}

func GetRequestBodyandUser(ctx *gin.Context) ([]byte, string, string, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetRequestBody"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.common.GetRequestBody", elapsed)
	}()

	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("controllers.common.GetRequestBody error: %s", err))
		}
	}()

	_, user, clientid, err := GetRequestUser(ctx)

	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("GetRequestBody"))

	body, err := ioutil.ReadAll(ctx.Request.Body)

	if err != nil {
		iLog.Error(fmt.Sprintf("GetRequestBody error: %s", err.Error()))
		return nil, clientid, user, err
	}
	iLog.Debug(fmt.Sprintf("GetRequestBody body: %s", body))
	return body, clientid, user, nil
}

func GetRequestBodyandUserbyJson(ctx *gin.Context) (map[string]interface{}, string, string, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetRequestBodyandUserbyJson"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.common.GetRequestBodyandUserbyJson", elapsed)
	}()
	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("controllers.common.GetRequestBodyandUserbyJson error: %s", err))
		}
	}()
	_, user, clientid, err := GetRequestUser(ctx)

	iLog.ClientID = clientid
	iLog.User = user
	iLog.Debug(fmt.Sprintf("GetRequestBodyandUserbyJson"))

	var request map[string]interface{}
	err = json.NewDecoder(ctx.Request.Body).Decode(&request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to decode request body: %v", err))
		return nil, clientid, user, err
	}
	iLog.Debug(fmt.Sprintf("GetRequestBodybyJson request: %s", request))
	return request, clientid, user, nil
}

func GetRequestBodybyJson(ctx *gin.Context) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetRequestBodybyJson"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.common.GetRequestBodybyJson", elapsed)
	}()
	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("controllers.common.GetRequestBodybyJson error: %s", err))
		}
	}()
	_, user, clientid, err := GetRequestUser(ctx)

	iLog.ClientID = clientid
	iLog.User = user
	iLog.Debug(fmt.Sprintf("GetRequestBodybyJson"))

	var request map[string]interface{}
	err = json.NewDecoder(ctx.Request.Body).Decode(&request)
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
