package function

import (
	"encoding/json"
	"fmt"

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

func (f *FunctionController) TestExecFunction(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Function"}
	iLog.Debug(fmt.Sprintf("Test Exec Function"))

	body, err := common.GetRequestBody(c)

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

	iLog.Debug(fmt.Sprintf("Test Exec Function funcdata: %s", funcdata))

	if funcdata.Type == int(types.Csharp) {
		csfuncs := funcs.CSharpFuncs{}
		outputs, err := csfuncs.Testfunction(funcdata.Content, funcdata.Inputs, funcdata.Outputs)

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