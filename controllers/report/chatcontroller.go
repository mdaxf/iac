package report

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/gormdb"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// ChatController handles chat and AI conversation endpoints
type ChatController struct {
	service *services.ChatService
}

// NewChatController creates a new chat controller
func NewChatController() *ChatController {
	// Check if GORM DB is initialized
	if gormdb.DB == nil {
		iLog := logger.Log{
			ModuleName:     logger.API,
			User:           "System",
			ControllerName: "chat",
		}
		iLog.Error("Failed to create ChatController: gormdb.DB is nil. GORM database may not be initialized properly.")
		return &ChatController{
			service: nil, // Service will be nil, endpoints should check this
		}
	}

	// Initialize SchemaMetadataService for auto-discovery fallback
	schemaService := services.NewSchemaMetadataService(gormdb.DB)

	return &ChatController{
		service: services.NewChatService(gormdb.DB, config.OpenAiKey, config.OpenAiModel, schemaService),
	}
}

// CreateConversation handles POST / - Create new conversation
func (cc *ChatController) CreateConversation(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "chat"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("chat.CreateConversation", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var request struct {
		Title            string `json:"title"`
		DatabaseAlias    string `json:"database_alias"`
		AutoExecuteQuery bool   `json:"auto_execute_query"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if service is initialized
	if cc.service == nil {
		iLog.Error("ChatService is not initialized - database may not be connected")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat service is not available. Database may not be initialized."})
		return
	}

	conversation, err := cc.service.CreateConversation(user, request.DatabaseAlias, request.Title, request.AutoExecuteQuery)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error creating conversation: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		return
	}

	iLog.Info(fmt.Sprintf("Conversation created: %s", conversation.ID))
	c.JSON(http.StatusCreated, conversation)
}

// GetConversation handles GET /:id - Get conversation by ID
func (cc *ChatController) GetConversation(c *gin.Context) {
	conversationID := c.Param("id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}

	// Check if service is initialized
	if cc.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat service is not available. Database may not be initialized."})
		return
	}

	conversation, err := cc.service.GetConversation(conversationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conversation)
}

// ListConversations handles GET / - List user conversations
func (cc *ChatController) ListConversations(c *gin.Context) {
	_, _, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if service is initialized
	if cc.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat service is not available. Database may not be initialized."})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	conversations, err := cc.service.ListConversations(user, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list conversations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversations": conversations,
		"count":         len(conversations),
	})
}

// DeleteConversation handles DELETE /:id - Delete conversation
func (cc *ChatController) DeleteConversation(c *gin.Context) {
	conversationID := c.Param("id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}

	// Check if service is initialized
	if cc.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat service is not available. Database may not be initialized."})
		return
	}

	if err := cc.service.DeleteConversation(conversationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted successfully"})
}

// SendMessage handles POST /:id/message - Send message and get AI response
func (cc *ChatController) SendMessage(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "chat"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("chat.SendMessage", elapsed)
	}()

	conversationID := c.Param("id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var request struct {
		Message          string `json:"message"`
		DatabaseAlias    string `json:"database_alias"`
		AutoExecuteQuery bool   `json:"auto_execute_query"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if request.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
		return
	}

	// Check if service is initialized
	if cc.service == nil {
		iLog.Error("ChatService is not initialized - database may not be connected")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat service is not available. Database may not be initialized."})
		return
	}

	// Process message and generate AI response
	response, err := cc.service.ProcessMessage(
		c.Request.Context(),
		conversationID,
		request.Message,
		request.DatabaseAlias,
		request.AutoExecuteQuery,
	)

	if err != nil {
		iLog.Error(fmt.Sprintf("Error processing message: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process message"})
		return
	}

	iLog.Info(fmt.Sprintf("Message processed successfully for conversation %s", conversationID))
	c.JSON(http.StatusOK, response)
}

// CreateSchemaEmbedding handles POST /schema/embedding - Create embedding for schema element
func (cc *ChatController) CreateSchemaEmbedding(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "chat"}

	body, _, _, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var request struct {
		DatabaseAlias string            `json:"database_alias"`
		EntityType    models.EntityType `json:"entity_type"`
		EntityID      string            `json:"entity_id"`
		EntityText    string            `json:"entity_text"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if service is initialized
	if cc.service == nil {
		iLog.Error("ChatService is not initialized - database may not be connected")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat service is not available. Database may not be initialized."})
		return
	}

	// Create embedding
	embedding, err := cc.service.CreateEmbedding(c.Request.Context(), request.EntityText)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error creating embedding: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create embedding"})
		return
	}

	// Save to database
	if err := cc.service.SaveSchemaEmbedding(request.DatabaseAlias, request.EntityType, request.EntityID, request.EntityText, embedding); err != nil {
		iLog.Error(fmt.Sprintf("Error saving embedding: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save embedding"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Embedding created successfully",
		"dimensions": len(embedding),
	})
}

// SearchSchema handles POST /schema/search - Search schema using semantic search
func (cc *ChatController) SearchSchema(c *gin.Context) {
	body, _, _, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var request struct {
		DatabaseAlias string `json:"database_alias"`
		Query         string `json:"query"`
		Limit         int    `json:"limit"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if service is initialized
	if cc.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat service is not available. Database may not be initialized."})
		return
	}

	if request.Limit == 0 {
		request.Limit = 10
	}

	// Search similar schema elements
	results, err := cc.service.SearchSimilarSchemaElements(c.Request.Context(), request.DatabaseAlias, request.Query, request.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search schema"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}
