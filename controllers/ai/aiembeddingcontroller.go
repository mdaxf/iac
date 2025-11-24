package ai

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

type AIEmbeddingController struct {
	service *services.AIEmbeddingService
}

func NewAIEmbeddingController() *AIEmbeddingController {
	return &AIEmbeddingController{
		service: &services.AIEmbeddingService{},
	}
}

// GetEmbeddingConfigurations returns all embedding configurations
func (c *AIEmbeddingController) GetEmbeddingConfigurations(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.GetConfigurations"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.ai.GetEmbeddingConfigurations", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	configs, err := c.service.GetEmbeddingConfigurations(&iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting embedding configurations: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, configs)
}

// GetEmbeddingConfigStats returns embedding configuration statistics
func (c *AIEmbeddingController) GetEmbeddingConfigStats(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.GetConfigStats"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.ai.GetEmbeddingConfigStats", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	stats, err := c.service.GetEmbeddingConfigStats(&iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting embedding config stats: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

// GetDatabasesWithEmbeddings returns list of databases that have embeddings
func (c *AIEmbeddingController) GetDatabasesWithEmbeddings(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.GetDatabasesWithEmbeddings"}

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	configID := ctx.Query("config_id")
	if configID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "config_id is required"})
		return
	}

	var configIDInt int
	fmt.Sscanf(configID, "%d", &configIDInt)

	databases, err := c.service.GetDatabasesWithEmbeddings(configIDInt, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting databases with embeddings: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, databases)
}

// GetSchemaMetadata returns schema metadata for a database
func (c *AIEmbeddingController) GetSchemaMetadata(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.GetSchemaMetadata"}

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	configID := ctx.Query("config_id")
	databaseAlias := ctx.Query("database_alias")

	if configID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "config_id is required"})
		return
	}

	var configIDInt int
	fmt.Sscanf(configID, "%d", &configIDInt)

	embeddings, err := c.service.GetDatabaseSchemaMetadata(configIDInt, databaseAlias, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting schema metadata: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, embeddings)
}

// GenerateSchemaEmbeddings generates embeddings for database schema
func (c *AIEmbeddingController) GenerateSchemaEmbeddings(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.GenerateSchemaEmbeddings"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.ai.GenerateSchemaEmbeddings", elapsed)
	}()

	_, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var req models.SchemaMetadataRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		iLog.Error(fmt.Sprintf("Error binding request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Generating schema embeddings for database: %s", req.DatabaseAlias))

	// TODO: Get embedding function from AI service
	embeddingFunc := func(text string) ([]float32, error) {
		// Placeholder - implement actual embedding generation
		return make([]float32, 1536), nil
	}

	go c.service.GenerateSchemaEmbeddings(context.Background(), req, embeddingFunc, &iLog)

	ctx.JSON(http.StatusAccepted, gin.H{
		"message": "Schema embedding generation started",
		"database": req.DatabaseAlias,
	})
}

// GetBusinessEntities returns all business entities
func (c *AIEmbeddingController) GetBusinessEntities(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.GetBusinessEntities"}

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	configID := ctx.Query("config_id")
	if configID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "config_id is required"})
		return
	}

	var configIDInt int
	fmt.Sscanf(configID, "%d", &configIDInt)

	entities, err := c.service.GetBusinessEntities(configIDInt, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting business entities: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, entities)
}

// CreateBusinessEntity creates a new business entity
func (c *AIEmbeddingController) CreateBusinessEntity(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.CreateBusinessEntity"}

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var req models.BusinessEntityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		iLog.Error(fmt.Sprintf("Error binding request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get embedding function from AI service
	embeddingFunc := func(text string) ([]float32, error) {
		return make([]float32, 1536), nil
	}

	entity, err := c.service.CreateBusinessEntity(context.Background(), req, embeddingFunc, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error creating business entity: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, entity)
}

// GetQueryTemplates returns all query templates
func (c *AIEmbeddingController) GetQueryTemplates(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.GetQueryTemplates"}

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	configID := ctx.Query("config_id")
	if configID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "config_id is required"})
		return
	}

	var configIDInt int
	fmt.Sscanf(configID, "%d", &configIDInt)

	templates, err := c.service.GetQueryTemplates(configIDInt, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting query templates: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, templates)
}

// CreateQueryTemplate creates a new query template
func (c *AIEmbeddingController) CreateQueryTemplate(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.CreateQueryTemplate"}

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var req models.QueryTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		iLog.Error(fmt.Sprintf("Error binding request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get embedding function from AI service
	embeddingFunc := func(text string) ([]float32, error) {
		return make([]float32, 1536), nil
	}

	template, err := c.service.CreateQueryTemplate(context.Background(), req, embeddingFunc, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error creating query template: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, template)
}

// SearchEmbeddings performs vector similarity search
func (c *AIEmbeddingController) SearchEmbeddings(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.SearchEmbeddings"}

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var req models.SearchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		iLog.Error(fmt.Sprintf("Error binding request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get embedding function from AI service
	embeddingFunc := func(text string) ([]float32, error) {
		return make([]float32, 1536), nil
	}

	results, err := c.service.SearchSchema(context.Background(), req, embeddingFunc, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error searching embeddings: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, results)
}

// UpdateBusinessEntity updates an existing business entity
func (c *AIEmbeddingController) UpdateBusinessEntity(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.UpdateBusinessEntity"}

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	id := ctx.Param("id")
	var idInt int
	fmt.Sscanf(id, "%d", &idInt)

	var req models.BusinessEntityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		iLog.Error(fmt.Sprintf("Error binding request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get embedding function from AI service
	embeddingFunc := func(text string) ([]float32, error) {
		return make([]float32, 1536), nil
	}

	entity, err := c.service.UpdateBusinessEntity(context.Background(), idInt, req, embeddingFunc, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error updating business entity: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, entity)
}

// DeleteBusinessEntity deletes a business entity
func (c *AIEmbeddingController) DeleteBusinessEntity(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.DeleteBusinessEntity"}

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	id := ctx.Param("id")
	var idInt int
	fmt.Sscanf(id, "%d", &idInt)

	if err := c.service.DeleteBusinessEntity(idInt, &iLog); err != nil {
		iLog.Error(fmt.Sprintf("Error deleting business entity: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Business entity deleted successfully"})
}

// UpdateQueryTemplate updates an existing query template
func (c *AIEmbeddingController) UpdateQueryTemplate(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.UpdateQueryTemplate"}

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	id := ctx.Param("id")
	var idInt int
	fmt.Sscanf(id, "%d", &idInt)

	var req models.QueryTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		iLog.Error(fmt.Sprintf("Error binding request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get embedding function from AI service
	embeddingFunc := func(text string) ([]float32, error) {
		return make([]float32, 1536), nil
	}

	template, err := c.service.UpdateQueryTemplate(context.Background(), idInt, req, embeddingFunc, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error updating query template: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, template)
}

// DeleteQueryTemplate deletes a query template
func (c *AIEmbeddingController) DeleteQueryTemplate(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIEmbedding.DeleteQueryTemplate"}

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	id := ctx.Param("id")
	var idInt int
	fmt.Sscanf(id, "%d", &idInt)

	if err := c.service.DeleteQueryTemplate(idInt, &iLog); err != nil {
		iLog.Error(fmt.Sprintf("Error deleting query template: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Query template deleted successfully"})
}

// RegisterRoutes registers all routes for AI embeddings
func (c *AIEmbeddingController) RegisterRoutes(router *gin.RouterGroup) {
	embeddings := router.Group("/embeddings")
	{
		// Configurations
		embeddings.GET("/configurations", c.GetEmbeddingConfigurations)
		embeddings.GET("/configurations/stats", c.GetEmbeddingConfigStats)
		embeddings.GET("/databases", c.GetDatabasesWithEmbeddings)

		// Schema Metadata
		schema := embeddings.Group("/schema")
		{
			schema.GET("/metadata", c.GetSchemaMetadata)
			schema.POST("/generate", c.GenerateSchemaEmbeddings)
		}

		// Business Entities
		entities := embeddings.Group("/entities")
		{
			entities.GET("", c.GetBusinessEntities)
			entities.POST("", c.CreateBusinessEntity)
			entities.PUT("/:id", c.UpdateBusinessEntity)
			entities.DELETE("/:id", c.DeleteBusinessEntity)
		}

		// Query Templates
		templates := embeddings.Group("/query-templates")
		{
			templates.GET("", c.GetQueryTemplates)
			templates.POST("", c.CreateQueryTemplate)
			templates.PUT("/:id", c.UpdateQueryTemplate)
			templates.DELETE("/:id", c.DeleteQueryTemplate)
		}

		// Search
		embeddings.POST("/search", c.SearchEmbeddings)
	}
}
