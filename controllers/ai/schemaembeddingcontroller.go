package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/services"
	"gorm.io/gorm"
)

// SchemaEmbeddingController handles API endpoints for vector embedding generation
type SchemaEmbeddingController struct {
	DB                     *gorm.DB
	SchemaEmbeddingService *services.SchemaEmbeddingService
	iLog                   logger.Log
}

// NewSchemaEmbeddingController creates a new schema embedding controller
func NewSchemaEmbeddingController(db *gorm.DB, openAIKey string) *SchemaEmbeddingController {
	return &SchemaEmbeddingController{
		DB:                     db,
		SchemaEmbeddingService: services.NewSchemaEmbeddingService(db, openAIKey),
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "SchemaEmbeddingController",
		},
	}
}

// GenerateEmbeddingsRequest represents the request to generate embeddings
type GenerateEmbeddingsRequest struct {
	DatabaseAlias string `json:"databasealias" binding:"required"`
}

// GenerateEmbeddingsResponse represents the response from embedding generation
type GenerateEmbeddingsResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	TablesCount   int    `json:"tablescount"`
	ColumnsCount  int    `json:"columnscount"`
	EntitiesCount int    `json:"entitiescount"`
}

// SearchRequest represents the request for semantic search
type SearchRequest struct {
	DatabaseAlias string `json:"databasealias" binding:"required"`
	Query         string `json:"query" binding:"required"`
	Limit         int    `json:"limit"`
}

// GenerateEmbeddings generates embeddings for all schema elements in a database
// POST /api/ai/embeddings/generate
func (c *SchemaEmbeddingController) GenerateEmbeddings(ctx *gin.Context) {
	c.iLog.Info("GenerateEmbeddings API called")

	var req GenerateEmbeddingsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("Generating embeddings for database: %s", req.DatabaseAlias))

	// Generate embeddings in background context
	bgCtx := context.Background()
	err := c.SchemaEmbeddingService.GenerateEmbeddingsForDatabase(bgCtx, req.DatabaseAlias)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to generate embeddings: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate embeddings: " + err.Error(),
		})
		return
	}

	// Count generated embeddings
	var tablesCount, columnsCount, entitiesCount int64
	c.DB.Table("databaseschemametadata").
		Where("databasealias = ? AND metadatatype = ? AND embedding IS NOT NULL", req.DatabaseAlias, "table").
		Count(&tablesCount)
	c.DB.Table("databaseschemametadata").
		Where("databasealias = ? AND metadatatype = ? AND embedding IS NOT NULL", req.DatabaseAlias, "column").
		Count(&columnsCount)
	c.DB.Table("businessentities").
		Where("databasealias = ? AND embedding IS NOT NULL", req.DatabaseAlias).
		Count(&entitiesCount)

	c.iLog.Info(fmt.Sprintf("Successfully generated embeddings: %d tables, %d columns, %d entities",
		tablesCount, columnsCount, entitiesCount))

	ctx.JSON(http.StatusOK, GenerateEmbeddingsResponse{
		Success:       true,
		Message:       "Embeddings generated successfully",
		TablesCount:   int(tablesCount),
		ColumnsCount:  int(columnsCount),
		EntitiesCount: int(entitiesCount),
	})
}

// SearchTables searches for tables using semantic search
// POST /api/ai/embeddings/search/tables
func (c *SchemaEmbeddingController) SearchTables(ctx *gin.Context) {
	c.iLog.Info("SearchTables API called")

	var req SearchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	c.iLog.Info(fmt.Sprintf("Searching tables in %s for query: %s", req.DatabaseAlias, req.Query))

	bgCtx := context.Background()
	results, err := c.SchemaEmbeddingService.SearchSimilarTables(bgCtx, req.DatabaseAlias, req.Query, req.Limit)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Search failed: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Search failed: " + err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("Found %d tables", len(results)))

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   len(results),
		"results": results,
	})
}

// SearchColumns searches for columns using semantic search
// POST /api/ai/embeddings/search/columns
func (c *SchemaEmbeddingController) SearchColumns(ctx *gin.Context) {
	c.iLog.Info("SearchColumns API called")

	var req SearchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	c.iLog.Info(fmt.Sprintf("Searching columns in %s for query: %s", req.DatabaseAlias, req.Query))

	bgCtx := context.Background()
	results, err := c.SchemaEmbeddingService.SearchSimilarColumns(bgCtx, req.DatabaseAlias, req.Query, req.Limit)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Search failed: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Search failed: " + err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("Found %d columns", len(results)))

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   len(results),
		"results": results,
	})
}

// SearchBusinessEntities searches for business entities using semantic search
// POST /api/ai/embeddings/search/entities
func (c *SchemaEmbeddingController) SearchBusinessEntities(ctx *gin.Context) {
	c.iLog.Info("SearchBusinessEntities API called")

	var req SearchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	c.iLog.Info(fmt.Sprintf("Searching business entities in %s for query: %s", req.DatabaseAlias, req.Query))

	bgCtx := context.Background()
	results, err := c.SchemaEmbeddingService.SearchSimilarBusinessEntities(bgCtx, req.DatabaseAlias, req.Query, req.Limit)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Search failed: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Search failed: " + err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("Found %d business entities", len(results)))

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   len(results),
		"results": results,
	})
}

// GetEmbeddingStatus returns the status of embeddings for a database
// GET /api/ai/embeddings/status/:databasealias
func (c *SchemaEmbeddingController) GetEmbeddingStatus(ctx *gin.Context) {
	databaseAlias := ctx.Param("databasealias")
	c.iLog.Info(fmt.Sprintf("GetEmbeddingStatus API called for database: %s", databaseAlias))

	if databaseAlias == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Database alias is required",
		})
		return
	}

	// Count total and embedded items
	var totalTables, embeddedTables int64
	var totalColumns, embeddedColumns int64
	var totalEntities, embeddedEntities int64

	c.DB.Table("databaseschemametadata").
		Where("databasealias = ? AND metadatatype = ?", databaseAlias, "table").
		Count(&totalTables)
	c.DB.Table("databaseschemametadata").
		Where("databasealias = ? AND metadatatype = ? AND embedding IS NOT NULL", databaseAlias, "table").
		Count(&embeddedTables)

	c.DB.Table("databaseschemametadata").
		Where("databasealias = ? AND metadatatype = ?", databaseAlias, "column").
		Count(&totalColumns)
	c.DB.Table("databaseschemametadata").
		Where("databasealias = ? AND metadatatype = ? AND embedding IS NOT NULL", databaseAlias, "column").
		Count(&embeddedColumns)

	c.DB.Table("businessentities").
		Where("databasealias = ?", databaseAlias).
		Count(&totalEntities)
	c.DB.Table("businessentities").
		Where("databasealias = ? AND embedding IS NOT NULL", databaseAlias).
		Count(&embeddedEntities)

	tablesCoverage := float64(0)
	if totalTables > 0 {
		tablesCoverage = float64(embeddedTables) / float64(totalTables) * 100
	}

	columnsCoverage := float64(0)
	if totalColumns > 0 {
		columnsCoverage = float64(embeddedColumns) / float64(totalColumns) * 100
	}

	entitiesCoverage := float64(0)
	if totalEntities > 0 {
		entitiesCoverage = float64(embeddedEntities) / float64(totalEntities) * 100
	}

	c.iLog.Info(fmt.Sprintf("Embedding status: Tables %.1f%%, Columns %.1f%%, Entities %.1f%%",
		tablesCoverage, columnsCoverage, entitiesCoverage))

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"tables": gin.H{
			"total":    totalTables,
			"embedded": embeddedTables,
			"coverage": fmt.Sprintf("%.1f%%", tablesCoverage),
		},
		"columns": gin.H{
			"total":    totalColumns,
			"embedded": embeddedColumns,
			"coverage": fmt.Sprintf("%.1f%%", columnsCoverage),
		},
		"entities": gin.H{
			"total":    totalEntities,
			"embedded": embeddedEntities,
			"coverage": fmt.Sprintf("%.1f%%", entitiesCoverage),
		},
	})
}

// DeleteEmbeddings deletes all embeddings for a database
// DELETE /api/ai/embeddings/:databasealias
func (c *SchemaEmbeddingController) DeleteEmbeddings(ctx *gin.Context) {
	databaseAlias := ctx.Param("databasealias")
	c.iLog.Info(fmt.Sprintf("DeleteEmbeddings API called for database: %s", databaseAlias))

	if databaseAlias == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Database alias is required",
		})
		return
	}

	// Set embeddings to NULL for all schema metadata
	err := c.DB.Exec(`
		UPDATE databaseschemametadata
		SET embedding = NULL, embedding_model = NULL, embedding_generated_at = NULL
		WHERE databasealias = ?
	`, databaseAlias).Error

	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to delete schema metadata embeddings: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete embeddings: " + err.Error(),
		})
		return
	}

	// Set embeddings to NULL for all business entities
	err = c.DB.Exec(`
		UPDATE businessentities
		SET embedding = NULL, embedding_model = NULL, embedding_generated_at = NULL
		WHERE databasealias = ?
	`, databaseAlias).Error

	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to delete business entity embeddings: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete embeddings: " + err.Error(),
		})
		return
	}

	c.iLog.Info(fmt.Sprintf("Successfully deleted all embeddings for database: %s", databaseAlias))

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Embeddings deleted successfully",
	})
}

// RegisterRoutes registers the schema embedding routes
func (c *SchemaEmbeddingController) RegisterRoutes(router *gin.RouterGroup) {
	embeddings := router.Group("/embeddings")
	{
		embeddings.POST("/generate", c.GenerateEmbeddings)
		embeddings.GET("/status/:databasealias", c.GetEmbeddingStatus)
		embeddings.DELETE("/:databasealias", c.DeleteEmbeddings)

		search := embeddings.Group("/search")
		{
			search.POST("/tables", c.SearchTables)
			search.POST("/columns", c.SearchColumns)
			search.POST("/entities", c.SearchBusinessEntities)
		}
	}
	c.iLog.Info("Schema embedding routes registered")
}

// Helper function to convert JSON to string
func jsonToString(data interface{}) string {
	bytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(bytes)
}
