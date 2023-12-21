package function

import (
	"encoding/json"
	"fmt"
	"time"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mdaxf/iac/controllers/common"

	funcs "github.com/mdaxf/iac/engine/function"
	"github.com/mdaxf/iac/engine/types"
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

// TestExecFunction is a controller function that tests the execution of a function.
// It receives a request context and returns the test results as JSON.
// The function logs performance metrics and any errors encountered during execution.
// It supports two types of functions: Go expressions and JavaScript functions.
// For Go expressions, it uses the GoExprFuncs package to execute the function.
// For JavaScript functions, it uses the JSFuncs package to execute the function.
// If the function type is not supported, it returns an error response.

func (f *FunctionController) TestExecFunction(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Function"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.function.TestExecFunction", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("DeleteDataFromTable error: %s", err))
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}() */
	/*	_, user, clientid, err := common.GetRequestUser(c)
		if err != nil {
			iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		iLog.ClientID = clientid
		iLog.User = user

		iLog.Debug(fmt.Sprintf("Test Exec Function"))

		body, err := common.GetRequestBody(c)

		if err != nil {
			iLog.Error(fmt.Sprintf("Test Exec Function error: %s", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} */
	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var funcdata FuncData
	err = json.Unmarshal(body, &funcdata)
	if err != nil {
		iLog.Error(fmt.Sprintf("Test Exec Function get the message body error: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Test Exec Function funcdata: %v", funcdata))

	if funcdata.Type == int(types.GoExpr) {
		gofuncs := funcs.GoExprFuncs{}
		outputs, err := gofuncs.Testfunction(funcdata.Content, funcdata.Inputs, funcdata.Outputs)

		if err != nil {
			iLog.Error(fmt.Sprintf("Test Exec Function error: %s", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": outputs})
		return
	} else if funcdata.Type == int(types.Javascript) {
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
