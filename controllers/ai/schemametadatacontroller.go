package ai

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/gormdb"
	"github.com/mdaxf/iac/services"
)

// SchemaMetadataController handles schema metadata HTTP requests
type SchemaMetadataController struct {
	service *services.SchemaMetadataService
}

// NewSchemaMetadataController creates a new schema metadata controller
func NewSchemaMetadataController() *SchemaMetadataController {
	return &SchemaMetadataController{
		service: services.NewSchemaMetadataService(gormdb.DB),
	}
}

// DiscoverSchema discovers database schema
// POST /api/schema/discover
func (c *SchemaMetadataController) DiscoverSchema(ctx *gin.Context) {
	var request struct {
		DatabaseAlias string `json:"database_alias" binding:"required"`
		DatabaseName  string `json:"database_name" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.DiscoverDatabaseSchema(ctx.Request.Context(), request.DatabaseAlias, request.DatabaseName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Schema discovered successfully",
	})
}

// GetTables retrieves all table metadata
// GET /api/schema/tables?database_alias=xxx
func (c *SchemaMetadataController) GetTables(ctx *gin.Context) {
	databaseAlias := ctx.Query("database_alias")

	if databaseAlias == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "database_alias is required"})
		return
	}

	tables, err := c.service.GetTableMetadata(ctx.Request.Context(), databaseAlias)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tables,
		"count":   len(tables),
	})
}

// GetColumns retrieves all column metadata for a table
// GET /api/schema/columns?database_alias=xxx&table_name=xxx
func (c *SchemaMetadataController) GetColumns(ctx *gin.Context) {
	databaseAlias := ctx.Query("database_alias")
	tableName := ctx.Query("table_name")

	if databaseAlias == "" || tableName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "database_alias and table_name are required"})
		return
	}

	columns, err := c.service.GetColumnMetadata(ctx.Request.Context(), databaseAlias, tableName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    columns,
		"count":   len(columns),
	})
}

// UpdateMetadata updates metadata description and business name
// PUT /api/schema/:id
func (c *SchemaMetadataController) UpdateMetadata(ctx *gin.Context) {
	id := ctx.Param("id")

	var request struct {
		Description *string `json:"description"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.UpdateMetadata(ctx.Request.Context(), id, request.Description); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Metadata updated successfully",
	})
}

// DeleteMetadata deletes metadata entry
// DELETE /api/schema/:id
func (c *SchemaMetadataController) DeleteMetadata(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.service.DeleteMetadata(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Metadata deleted successfully",
	})
}

// SearchMetadata searches metadata by keyword
// GET /api/schema/search?database_alias=xxx&keyword=xxx
func (c *SchemaMetadataController) SearchMetadata(ctx *gin.Context) {
	databaseAlias := ctx.Query("database_alias")
	keyword := ctx.Query("keyword")

	if databaseAlias == "" || keyword == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "database_alias and keyword are required"})
		return
	}

	results, err := c.service.SearchMetadata(ctx.Request.Context(), databaseAlias, keyword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"count":   len(results),
	})
}

// GetDatabases retrieves all database aliases
// GET /api/schema/databases
func (c *SchemaMetadataController) GetDatabases(ctx *gin.Context) {
	databases, err := c.service.GetAllDatabases(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    databases,
		"count":   len(databases),
	})
}

// GetSchemaContext retrieves schema context for AI
// POST /api/schema/context
func (c *SchemaMetadataController) GetSchemaContext(ctx *gin.Context) {
	var request struct {
		DatabaseAlias string   `json:"database_alias" binding:"required"`
		TableNames    []string `json:"table_names"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	context, err := c.service.GetSchemaContext(ctx.Request.Context(), request.DatabaseAlias, request.TableNames)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"context": context,
	})
}

// GetDatabaseMetadata retrieves complete metadata (tables and columns) for a database
// GET /api/schema-metadata/databases/:dbName/metadata
func (c *SchemaMetadataController) GetDatabaseMetadata(ctx *gin.Context) {
	databaseAlias := ctx.Param("dbName")

	if databaseAlias == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "database alias is required"})
		return
	}

	metadata, err := c.service.GetDatabaseMetadata(ctx.Request.Context(), databaseAlias)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metadata,
		"count":   len(metadata),
	})
}
