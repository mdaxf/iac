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
	Image string `json:"image"`
	Text  string `json:"text"`
	Grid  string `json:"grid"`
	Theme string `json:"theme"`
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

	html, err := codegen.GetHtmlCodeFromImage(data.Image, config.OpenAiKey, data.Text, data.Grid, data.Theme)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": html})

}
