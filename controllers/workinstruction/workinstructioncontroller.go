package workinstruction

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// WorkInstructionController handles work instruction API requests
type WorkInstructionController struct {
	workInstructionService *services.WorkInstructionService
}

// NewWorkInstructionController creates a new work instruction controller
func NewWorkInstructionController(docDB documents.DocumentDB, serverURL string) *WorkInstructionController {
	return &WorkInstructionController{
		workInstructionService: services.NewWorkInstructionService(docDB, serverURL),
	}
}

// =====================================================
// Work Instruction Endpoints
// =====================================================

// ListWorkInstructions handles GET /api/wis
func (wic *WorkInstructionController) ListWorkInstructions(ctx *gin.Context) {
	workInstructions, err := wic.workInstructionService.ListWorkInstructions(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  workInstructions,
		"count": len(workInstructions),
	})
}

// GetWorkInstruction handles GET /api/wis/:id
func (wic *WorkInstructionController) GetWorkInstruction(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "work instruction ID is required"})
		return
	}

	workInstruction, err := wic.workInstructionService.GetWorkInstruction(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": workInstruction})
}

// CreateWorkInstruction handles POST /api/wis
func (wic *WorkInstructionController) CreateWorkInstruction(ctx *gin.Context) {
	var req models.WorkInstructionCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user from authentication context
	user := "system"

	workInstruction, err := wic.workInstructionService.CreateWorkInstruction(ctx.Request.Context(), &req, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Work instruction created successfully",
		"data":    workInstruction,
	})
}

// UpdateWorkInstruction handles PUT /api/wis/:id
func (wic *WorkInstructionController) UpdateWorkInstruction(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "work instruction ID is required"})
		return
	}

	var req models.WorkInstructionUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user from authentication context
	user := "system"

	workInstruction, err := wic.workInstructionService.UpdateWorkInstruction(ctx.Request.Context(), id, &req, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Work instruction updated successfully",
		"data":    workInstruction,
	})
}

// DeleteWorkInstruction handles DELETE /api/wis/:id
func (wic *WorkInstructionController) DeleteWorkInstruction(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "work instruction ID is required"})
		return
	}

	if err := wic.workInstructionService.DeleteWorkInstruction(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Work instruction deleted successfully"})
}

// DuplicateWorkInstruction handles POST /api/wis/:id/duplicate
func (wic *WorkInstructionController) DuplicateWorkInstruction(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "work instruction ID is required"})
		return
	}

	// TODO: Get user from authentication context
	user := "system"

	workInstruction, err := wic.workInstructionService.DuplicateWorkInstruction(ctx.Request.Context(), id, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Work instruction duplicated successfully",
		"data":    workInstruction,
	})
}

// ChangeStatus handles PATCH /api/wis/:id/status
func (wic *WorkInstructionController) ChangeStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "work instruction ID is required"})
		return
	}

	var req struct {
		Status models.InstructionStatus `json:"status" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user from authentication context
	user := "system"

	if err := wic.workInstructionService.ChangeStatus(ctx.Request.Context(), id, req.Status, user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Work instruction status updated successfully"})
}

// GetByStatus handles GET /api/wis/status/:status
func (wic *WorkInstructionController) GetByStatus(ctx *gin.Context) {
	status := models.InstructionStatus(ctx.Param("status"))

	workInstructions, err := wic.workInstructionService.GetByStatus(ctx.Request.Context(), status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  workInstructions,
		"count": len(workInstructions),
	})
}

// SearchWorkInstructions handles GET /api/wis/search?q=:query
func (wic *WorkInstructionController) SearchWorkInstructions(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	workInstructions, err := wic.workInstructionService.SearchWorkInstructions(ctx.Request.Context(), query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  workInstructions,
		"count": len(workInstructions),
	})
}
