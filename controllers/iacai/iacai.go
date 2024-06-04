package iacai

import (
	//"log"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

func (f *IACAIController) UserStorytoMockup(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.UserStorytoMockup", elapsed)
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

	iLog.Debug(fmt.Sprintf("UserStory to Mockup: %v", body))

	isStreamingRequest := strings.Contains(c.Request.Header.Get("Accept"), "stream")

	var data RequestBody

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("UserStory to Mockup call header: %v", c.Request.Header)) //application/json, text/event-stream
	iLog.Debug(fmt.Sprintf("UserStory to Mockup: %v", data))

	if isStreamingRequest {
		f.UserStorytoMockupbyStream(c, data)
	} else {
		result, err := codegen.GetMockupFromUserStory(config.OpenAiKey, config.OpenAiModel, data.Text)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": result})
	}

	//c.JSON(http.StatusOK, gin.H{"data": result})

}

func (f *IACAIController) UserStorytoMockupbyStream(c *gin.Context, data RequestBody) {
	iLog := logger.Log{ModuleName: logger.API, User: "System",
		ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.UserStorytoMockupbyStream", elapsed)
	}()

	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	stream := make(chan string)

	// Start a goroutine to generate the mockup and send it in chunks
	go func() {
		defer close(stream)

		result, err := codegen.GetMockupFromUserStory(config.OpenAiKey, config.OpenAiModel, data.Text)
		if err != nil {
			stream <- fmt.Sprintf("event: error\ndata: %s\n\n", err.Error())
			return
		}

		// Split the result into chunks and send them one by one
		chunks := common.ChunkString(fmt.Sprintf("%v", result), 4096) // Adjust the chunk size as needed
		for _, chunk := range chunks {
			stream <- fmt.Sprintf("data: %s\n\n", chunk)
		}
	}()

	// Listen for connection close and close the channel
	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-stream; ok {
			c.Writer.Write([]byte(msg))
			return true
		}
		return false
	})
}
