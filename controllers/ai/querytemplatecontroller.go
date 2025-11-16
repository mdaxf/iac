package ai

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/gormdb"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// QueryTemplateController handles query template HTTP requests
type QueryTemplateController struct {
	service *services.QueryTemplateService
}

// NewQueryTemplateController creates a new query template controller
func NewQueryTemplateController() *QueryTemplateController {
	return &QueryTemplateController{
		service: services.NewQueryTemplateService(gormdb.DB),
	}
}

// CreateTemplate creates a new query template
// POST /api/query-templates
func (c *QueryTemplateController) CreateTemplate(ctx *gin.Context) {
	var template models.QueryTemplate

	if err := ctx.ShouldBindJSON(&template); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.CreateTemplate(ctx.Request.Context(), &template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    template,
		"message": "Query template created successfully",
	})
}

// GetTemplate retrieves a query template by ID
// GET /api/query-templates/:id
func (c *QueryTemplateController) GetTemplate(ctx *gin.Context) {
	id := ctx.Param("id")

	template, err := c.service.GetTemplate(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    template,
	})
}

// ListTemplates retrieves all query templates
// GET /api/query-templates?database_alias=xxx
func (c *QueryTemplateController) ListTemplates(ctx *gin.Context) {
	databaseAlias := ctx.Query("database_alias")

	templates, err := c.service.ListTemplates(ctx.Request.Context(), databaseAlias)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    templates,
		"count":   len(templates),
	})
}

// UpdateTemplate updates a query template
// PUT /api/query-templates/:id
func (c *QueryTemplateController) UpdateTemplate(ctx *gin.Context) {
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
	delete(updates, "usage_count")
	delete(updates, "last_used_at")

	if err := c.service.UpdateTemplate(ctx.Request.Context(), id, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Query template updated successfully",
	})
}

// DeleteTemplate deletes a query template
// DELETE /api/query-templates/:id
func (c *QueryTemplateController) DeleteTemplate(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.service.DeleteTemplate(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Query template deleted successfully",
	})
}

// SearchTemplates searches query templates
// GET /api/query-templates/search?database_alias=xxx&keyword=xxx
func (c *QueryTemplateController) SearchTemplates(ctx *gin.Context) {
	databaseAlias := ctx.Query("database_alias")
	keyword := ctx.Query("keyword")

	if keyword == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "keyword is required"})
		return
	}

	templates, err := c.service.SearchTemplates(ctx.Request.Context(), databaseAlias, keyword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    templates,
		"count":   len(templates),
	})
}

// GetTemplatesByIntent retrieves templates by natural language intent
// POST /api/query-templates/by-intent
func (c *QueryTemplateController) GetTemplatesByIntent(ctx *gin.Context) {
	var request struct {
		DatabaseAlias string `json:"database_alias" binding:"required"`
		Intent        string `json:"intent" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	templates, err := c.service.GetTemplatesByIntent(ctx.Request.Context(), request.DatabaseAlias, request.Intent)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    templates,
		"count":   len(templates),
	})
}

// IncrementUsageCount increments the usage count of a template
// POST /api/query-templates/:id/use
func (c *QueryTemplateController) IncrementUsageCount(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.service.IncrementUsageCount(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Usage count incremented",
	})
}

// GetTemplateContext retrieves template context for AI
// GET /api/query-templates/context?database_alias=xxx&limit=10
func (c *QueryTemplateController) GetTemplateContext(ctx *gin.Context) {
	databaseAlias := ctx.Query("database_alias")
	limitStr := ctx.Query("limit")

	if databaseAlias == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "database_alias is required"})
		return
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	context, err := c.service.GetTemplateContext(ctx.Request.Context(), databaseAlias, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"context": context,
	})
}
