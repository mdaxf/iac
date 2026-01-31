package document

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// DocumentController handles document API requests
type DocumentController struct {
	documentService *services.DocumentService
}

// NewDocumentController creates a new document controller
func NewDocumentController(docDB documents.DocumentDB, storagePath string, baseURL string, serverURL string) *DocumentController {
	return &DocumentController{
		documentService: services.NewDocumentService(docDB, storagePath, baseURL, serverURL),
	}
}

// UploadDocument handles POST /api/documents/upload
func (dc *DocumentController) UploadDocument(ctx *gin.Context) {
	println("=== DocumentController.UploadDocument called ===")

	// Get file from form
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		println("ERROR: Failed to get file from form:", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	println("Received file:", header.Filename, "Size:", header.Size)

	// Get optional metadata
	description := ctx.PostForm("description")
	tags := ctx.PostFormArray("tags")

	// TODO: Get user from authentication context
	user := "system"

	// Upload document
	doc, err := dc.documentService.UploadDocument(ctx.Request.Context(), file, header, user, description, tags)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Document uploaded successfully",
		"data":    doc,
	})
}

// ListDocuments handles GET /api/documents
func (dc *DocumentController) ListDocuments(ctx *gin.Context) {
	docs, err := dc.documentService.ListDocuments(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  docs,
		"count": len(docs),
	})
}

// GetDocument handles GET /api/documents/:id
func (dc *DocumentController) GetDocument(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "document ID is required"})
		return
	}

	doc, err := dc.documentService.GetDocument(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": doc})
}

// UpdateDocument handles PUT /api/documents/:id
func (dc *DocumentController) UpdateDocument(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "document ID is required"})
		return
	}

	var req models.DocumentUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user from authentication context
	user := "system"

	doc, err := dc.documentService.UpdateDocument(ctx.Request.Context(), id, &req, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Document updated successfully",
		"data":    doc,
	})
}

// DeleteDocument handles DELETE /api/documents/:id
func (dc *DocumentController) DeleteDocument(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "document ID is required"})
		return
	}

	if err := dc.documentService.DeleteDocument(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}

// SearchDocuments handles GET /api/documents/search?q=query
func (dc *DocumentController) SearchDocuments(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	docs, err := dc.documentService.SearchDocuments(ctx.Request.Context(), query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  docs,
		"count": len(docs),
	})
}

// GetByType handles GET /api/documents/type/:type
func (dc *DocumentController) GetByType(ctx *gin.Context) {
	docType := models.DocumentType(ctx.Param("type"))

	docs, err := dc.documentService.GetDocumentsByType(ctx.Request.Context(), docType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  docs,
		"count": len(docs),
	})
}
