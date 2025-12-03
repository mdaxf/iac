package ai

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/services"
	"gorm.io/gorm"
)

// SchemaContextController handles API endpoints for schema context retrieval
type SchemaContextController struct {
	DB                    *gorm.DB
	SchemaContextService  *services.SchemaContextService
	iLog                  logger.Log
}

// NewSchemaContextController creates a new schema context controller
func NewSchemaContextController(db *gorm.DB, openAIKey string) *SchemaContextController {
	return &SchemaContextController{
		DB:                   db,
		SchemaContextService: services.NewSchemaContextService(db, openAIKey),
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "SchemaContextController",
		},
	}
}

// GetSchemaContextRequest represents the request to get schema context
type GetSchemaContextRequest struct {
	DatabaseAlias string `json:"databasealias" binding:"required"`
	Query         string `json:"query" binding:"required"`
	MaxTables     int    `json:"max_tables,omitempty"`
}

// GetSchemaContext retrieves relevant schema context using natural language query
// POST /api/ai/schema-context
func (c *SchemaContextController) GetSchemaContext(ctx *gin.Context) {
	c.iLog.Info("GetSchemaContext API called")

	var req GetSchemaContextRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Set default max_tables if not provided
	if req.MaxTables <= 0 {
		req.MaxTables = 10
	}

	c.iLog.Info(fmt.Sprintf("Searching schema context for database '%s' with query: %s (max %d tables)",
		req.DatabaseAlias, req.Query, req.MaxTables))

	// Get schema context using vector search
	bgCtx := context.Background()
	schemaContext, err := c.SchemaContextService.GetSchemaContextByQuery(bgCtx, req.DatabaseAlias, req.Query, req.MaxTables)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to get schema context: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get schema context: " + err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("Schema context retrieved: %d tables, %d entities, %d templates",
		schemaContext.TotalTables, schemaContext.TotalEntities, schemaContext.TotalTemplates))

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    schemaContext,
	})
}

// GetSchemaContextFormatted retrieves schema context formatted as text for AI prompts
// POST /api/ai/schema-context/formatted
func (c *SchemaContextController) GetSchemaContextFormatted(ctx *gin.Context) {
	c.iLog.Info("GetSchemaContextFormatted API called")

	var req GetSchemaContextRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Set default max_tables if not provided
	if req.MaxTables <= 0 {
		req.MaxTables = 10
	}

	c.iLog.Info(fmt.Sprintf("Getting formatted schema context for database '%s' with query: %s",
		req.DatabaseAlias, req.Query))

	// Get schema context using vector search
	bgCtx := context.Background()
	schemaContext, err := c.SchemaContextService.GetSchemaContextByQuery(bgCtx, req.DatabaseAlias, req.Query, req.MaxTables)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to get schema context: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get schema context: " + err.Error(),
		})
		return
	}

	// Format for AI prompt
	formattedContext := c.SchemaContextService.FormatSchemaContextForAI(schemaContext)

	c.iLog.Info(fmt.Sprintf("Formatted schema context generated (%d characters)", len(formattedContext)))

	ctx.JSON(http.StatusOK, gin.H{
		"success":         true,
		"formatted_context": formattedContext,
		"metadata": gin.H{
			"total_tables":    schemaContext.TotalTables,
			"total_entities":  schemaContext.TotalEntities,
			"total_templates": schemaContext.TotalTemplates,
		},
	})
}

// RegisterRoutes registers the schema context routes
func (c *SchemaContextController) RegisterRoutes(router *gin.RouterGroup) {
	schemaContext := router.Group("/schema-context")
	{
		schemaContext.POST("", c.GetSchemaContext)
		schemaContext.POST("/formatted", c.GetSchemaContextFormatted)
	}
	c.iLog.Info("Schema context routes registered")
}
