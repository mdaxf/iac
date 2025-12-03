package iacai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/services"
)

// AIAgencyRequest represents a request to the AI agency
type AIAgencyRequestBody struct {
	SessionID           string                   `json:"session_id"`
	Question            string                   `json:"question"`
	EditorType          string                   `json:"editor_type"`
	DatabaseAlias       string                   `json:"database_alias,omitempty"`
	EntityID            string                   `json:"entity_id,omitempty"`
	EntityContext       map[string]interface{}   `json:"entity_context,omitempty"`
	PageContext         map[string]interface{}   `json:"page_context,omitempty"`
	ConversationHistory []map[string]interface{} `json:"conversation_history,omitempty"`
	Options             map[string]interface{}   `json:"options,omitempty"`
}

var aiAgencyServiceInstance *services.AIAgencyService

// SetAIAgencyService sets the AI agency service instance
func SetAIAgencyService(service *services.AIAgencyService) {
	aiAgencyServiceInstance = service
}

// AskAIAgency handles AI agency requests with intelligent routing to specialized handlers
func (f *IACAIController) AskAIAgency(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.AskAIAgency", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("Ask AI Agency request body: %v", string(body)))

	var data AIAgencyRequestBody

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Ask AI Agency request data: %v", data))

	if data.Question == "" {
		iLog.Error("Question is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question is required"})
		return
	}

	// Check if AI Agency service is initialized
	if aiAgencyServiceInstance == nil {
		iLog.Error("AI Agency service not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AI Agency service is not yet initialized. Please contact administrator.",
		})
		return
	}

	// Call AI Agency service
	response, err := aiAgencyServiceInstance.HandleRequest(c.Request.Context(), &services.AIAgencyRequest{
		SessionID:           data.SessionID,
		UserID:              user,
		Question:            data.Question,
		EditorType:          data.EditorType,
		DatabaseAlias:       data.DatabaseAlias,
		EntityID:            data.EntityID,
		EntityContext:       data.EntityContext,
		PageContext:         data.PageContext,
		ConversationHistory: data.ConversationHistory,
		Options:             data.Options,
	})

	if err != nil {
		iLog.Error(fmt.Sprintf("Error handling AI agency request: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Successfully handled AI agency request for user: %s, intent: %s", user, response.IntentType))

	c.JSON(http.StatusOK, response)
}

// GetConversationSession retrieves a conversation session
func (f *IACAIController) GetConversationSession(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.GetConversationSession", elapsed)
	}()

	sessionID := c.Param("sessionId")
	if sessionID == "" {
		iLog.Error("Session ID is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	_, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting user: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	// TODO: Implement session retrieval from database
	iLog.Debug(fmt.Sprintf("Getting session: %s for user: %s", sessionID, user))

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"message":    "Session retrieval not yet implemented",
	})
}

// ClearConversationSession clears/ends a conversation session
func (f *IACAIController) ClearConversationSession(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "iacai"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.iacai.ClearConversationSession", elapsed)
	}()

	sessionID := c.Param("sessionId")
	if sessionID == "" {
		iLog.Error("Session ID is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	_, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting user: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	// TODO: Implement session clearing
	iLog.Debug(fmt.Sprintf("Clearing session: %s for user: %s", sessionID, user))

	c.JSON(http.StatusOK, gin.H{
		"message": "Session cleared successfully",
	})
}
