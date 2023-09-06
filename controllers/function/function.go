package function

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"

	funcs "github.com/mdaxf/iac/engine/function"
	"github.com/mdaxf/iac/logger"
)

type FunctionController struct {
}

type FuncData struct {
	Content string      `json:"content"`
	Inputs  interface{} `json:"inputs"`
	Outputs []string    `json:"outputs"`
	Type    int         `json:"type"`
}

func (f *FunctionController) TestExecFunction(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Function"}
	iLog.Debug(fmt.Sprintf("Test Exec Function"))

	body, err := GetRequestBody(c)

	if err != nil {
		iLog.Error(fmt.Sprintf("Test Exec Function error: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var funcdata FuncData
	err = json.Unmarshal(body, &funcdata)
	if err != nil {
		iLog.Error(fmt.Sprintf("Test Exec Function get the message body error: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	/*
		var inputs []map[string]interface{}
		err = json.Unmarshal([]byte(funcdata.Inputs.(string)), &inputs)
		if err != nil {
			iLog.Error(fmt.Sprintf("Test Exec Function get the inputs error: %s", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}  */
	/*
		if funcobj.Functype == types.Csharp {
			//TODO
		} else
	*/
	if funcdata.Type == 2 {
		jsfuncs := funcs.JSFuncs{}
		outputs, err := jsfuncs.Testfunction(funcdata.Content, funcdata.Inputs, funcdata.Outputs)

		if err != nil {
			iLog.Error(fmt.Sprintf("Test Exec Function error: %s", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": outputs})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "not supported type"})
	return
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
