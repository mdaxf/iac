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

type FlowGenerationRequest struct {
	Description      string                 `json:"description"`
	CurrentTranCode  map[string]interface{} `json:"currentTranCode"`
}

type PageGenerationRequest struct {
	Description  string                 `json:"description"`
	CurrentPage  map[string]interface{} `json:"currentPage"`
}

type ViewGenerationRequest struct {
	Description  string                 `json:"description"`
	CurrentView  map[string]interface{} `json:"currentView"`
}

type WorkflowGenerationRequest struct {
	Description      string                 `json:"description"`
	CurrentWorkflow  map[string]interface{} `json:"currentWorkflow"`
}

type WhiteboardGenerationRequest struct {
	Description        string                 `json:"description"`
	CurrentWhiteboard  map[string]interface{} `json:"currentWhiteboard"`
}

type AssistantRequest struct {
	Question             string                   `json:"question"`
	PageContext          map[string]interface{}   `json:"pageContext"`
	ConversationHistory  []map[string]interface{} `json:"conversationHistory"`
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

// GenerateFlow handles BPM flow generation from text description using AI
func (f *IACAIController) GenerateFlow(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.GenerateFlow", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("Generate Flow request body: %v", string(body)))

	var data FlowGenerationRequest

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Generate Flow request data: %v", data))

	if data.Description == "" {
		iLog.Error("Description is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}

	result, err := codegen.GenerateFlowFromDescription(
		data.Description,
		config.OpenAiKey,
		config.OpenAiModel,
		data.CurrentTranCode,
	)

	if err != nil {
		iLog.Error(fmt.Sprintf("Error generating flow: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Successfully generated flow for user: %s", user))

	c.JSON(http.StatusOK, result)
}

// GeneratePage handles page structure generation from text description using AI
func (f *IACAIController) GeneratePage(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.GeneratePage", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("Generate Page request body: %v", string(body)))

	var data PageGenerationRequest

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Generate Page request data: %v", data))

	if data.Description == "" {
		iLog.Error("Description is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}

	result, err := codegen.GeneratePageFromDescription(
		data.Description,
		config.OpenAiKey,
		config.OpenAiModel,
		data.CurrentPage,
	)

	if err != nil {
		iLog.Error(fmt.Sprintf("Error generating page: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Successfully generated page for user: %s", user))

	c.JSON(http.StatusOK, result)
}

// GenerateView handles view structure generation from text description using AI
func (f *IACAIController) GenerateView(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.GenerateView", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("Generate View request body: %v", string(body)))

	var data ViewGenerationRequest

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Generate View request data: %v", data))

	if data.Description == "" {
		iLog.Error("Description is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}

	result, err := codegen.GenerateViewFromDescription(
		data.Description,
		config.OpenAiKey,
		config.OpenAiModel,
		data.CurrentView,
	)

	if err != nil {
		iLog.Error(fmt.Sprintf("Error generating view: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Successfully generated view for user: %s", user))

	c.JSON(http.StatusOK, result)
}

// GenerateWorkflow handles workflow structure generation from text description using AI
func (f *IACAIController) GenerateWorkflow(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.GenerateWorkflow", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("Generate Workflow request body: %v", string(body)))

	var data WorkflowGenerationRequest

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Generate Workflow request data: %v", data))

	if data.Description == "" {
		iLog.Error("Description is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}

	result, err := codegen.GenerateWorkflowFromDescription(
		data.Description,
		config.OpenAiKey,
		config.OpenAiModel,
		data.CurrentWorkflow,
	)

	if err != nil {
		iLog.Error(fmt.Sprintf("Error generating workflow: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Successfully generated workflow for user: %s", user))

	c.JSON(http.StatusOK, result)
}

// GenerateWhiteboard handles whiteboard structure generation from text description using AI
func (f *IACAIController) GenerateWhiteboard(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.GenerateWhiteboard", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("Generate Whiteboard request body: %v", string(body)))

	var data WhiteboardGenerationRequest

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Generate Whiteboard request data: %v", data))

	if data.Description == "" {
		iLog.Error("Description is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}

	result, err := codegen.GenerateWhiteboardFromDescription(
		data.Description,
		config.OpenAiKey,
		config.OpenAiModel,
		data.CurrentWhiteboard,
	)

	if err != nil {
		iLog.Error(fmt.Sprintf("Error generating whiteboard: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Successfully generated whiteboard for user: %s", user))

	c.JSON(http.StatusOK, result)
}

// AskAssistant handles AI assistant questions with page context
func (f *IACAIController) AskAssistant(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.AskAssistant", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("Ask Assistant request body: %v", string(body)))

	var data AssistantRequest

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Ask Assistant request data: %v", data))

	if data.Question == "" {
		iLog.Error("Question is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question is required"})
		return
	}

	answer, err := codegen.GenerateAssistantResponse(
		data.Question,
		config.OpenAiKey,
		config.OpenAiModel,
		data.PageContext,
		data.ConversationHistory,
	)

	if err != nil {
		iLog.Error(fmt.Sprintf("Error generating assistant response: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Successfully generated assistant response for user: %s", user))

	c.JSON(http.StatusOK, gin.H{"answer": answer})
}
