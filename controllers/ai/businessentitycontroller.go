package ai

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/gormdb"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// BusinessEntityController handles business entity HTTP requests
type BusinessEntityController struct {
	service *services.BusinessEntityService
}

// NewBusinessEntityController creates a new business entity controller
func NewBusinessEntityController() *BusinessEntityController {
	return &BusinessEntityController{
		service: services.NewBusinessEntityService(gormdb.DB),
	}
}

// CreateEntity creates a new business entity
// POST /api/business-entities
func (c *BusinessEntityController) CreateEntity(ctx *gin.Context) {
	var entity models.BusinessEntity

	if err := ctx.ShouldBindJSON(&entity); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.CreateEntity(ctx.Request.Context(), &entity); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    entity,
		"message": "Business entity created successfully",
	})
}

// GetEntity retrieves a business entity by ID
// GET /api/business-entities/:id
func (c *BusinessEntityController) GetEntity(ctx *gin.Context) {
	id := ctx.Param("id")

	entity, err := c.service.GetEntity(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    entity,
	})
}

// ListEntities retrieves all business entities
// GET /api/business-entities?database_alias=xxx
func (c *BusinessEntityController) ListEntities(ctx *gin.Context) {
	databaseAlias := ctx.Query("database_alias")

	entities, err := c.service.ListEntities(ctx.Request.Context(), databaseAlias)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    entities,
		"count":   len(entities),
	})
}

// UpdateEntity updates a business entity
// PUT /api/business-entities/:id
func (c *BusinessEntityController) UpdateEntity(ctx *gin.Context) {
	id := ctx.Param("id")

	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Remove fields that shouldn't be updated directly
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")

	if err := c.service.UpdateEntity(ctx.Request.Context(), id, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Business entity updated successfully",
	})
}

// DeleteEntity deletes a business entity
// DELETE /api/business-entities/:id
func (c *BusinessEntityController) DeleteEntity(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.service.DeleteEntity(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Business entity deleted successfully",
	})
}

// SearchEntities searches business entities
// GET /api/business-entities/search?database_alias=xxx&keyword=xxx
func (c *BusinessEntityController) SearchEntities(ctx *gin.Context) {
	databaseAlias := ctx.Query("database_alias")
	keyword := ctx.Query("keyword")

	if keyword == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "keyword is required"})
		return
	}

	entities, err := c.service.SearchEntities(ctx.Request.Context(), databaseAlias, keyword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    entities,
		"count":   len(entities),
	})
}

// GetEntitiesByTable retrieves business entities for a table
// GET /api/business-entities/by-table?database_alias=xxx&table_name=xxx
func (c *BusinessEntityController) GetEntitiesByTable(ctx *gin.Context) {
	databaseAlias := ctx.Query("database_alias")
	tableName := ctx.Query("table_name")

	if databaseAlias == "" || tableName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "database_alias and table_name are required"})
		return
	}

	entities, err := c.service.GetEntitiesByTable(ctx.Request.Context(), databaseAlias, tableName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    entities,
		"count":   len(entities),
	})
}
