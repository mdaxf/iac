package iacai

import (
	//"log"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/codegen"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/logger"
)

type IACAIController struct {
}

type RequestBody struct {
	Image         string                   `json:"image"`
	Text          string                   `json:"text"`
	Grid          string                   `json:"grid"`
	Theme         string                   `json:"theme"`
	PreviouseObjs []map[string]interface{} `json:"previouseObjs"`
}

// CreateAI handles the creation of a new AI.
// It retrieves the request body and user information from the context,
func (f *IACAIController) ImagetoHTML(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.ImagetoHTML", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("IACAI.ImageToHTML error: %s", err))
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("Image to HTML: %v", body))

	var data RequestBody

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Image to HTML: %v", data))

	result, err := codegen.GetHtmlCodeFromImage(data.Image, config.OpenAiKey, config.OpenAiModel, data.Text, data.Grid, data.Theme, data.PreviouseObjs)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})

}

func (f *IACAIController) ImagetoBPMFlow(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.ImagetoBPMFlow", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("IACAI.ImageToHTML error: %s", err))
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("Image to BPM logic: %v", body))

	var data RequestBody

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Image to BPM logic: %v", data))

	result, err := codegen.GetBPMLogicFromImage(data.Image, config.OpenAiKey, config.OpenAiModel, data.Text, data.Grid, data.Theme, data.PreviouseObjs)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})

}

func (f *IACAIController) ImagetoWorkFlow(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.ImagetoWorkFlow", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("IACAI.ImageToHTML error: %s", err))
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("Image to work Flow: %v", body))

	var data RequestBody

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Image to work Flow: %v", data))

	result, err := codegen.GetWorkFlowFromImage(data.Image, config.OpenAiKey, config.OpenAiModel, data.Text, data.Grid, data.Theme, data.PreviouseObjs)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})

}
