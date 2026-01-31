package dataset

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// DataSetController handles dataset API requests
type DataSetController struct {
	datasetService *services.DataSetService
}

// NewDataSetController creates a new dataset controller
func NewDataSetController(docDB documents.DocumentDB) *DataSetController {
	return &DataSetController{
		datasetService: services.NewDataSetService(docDB),
	}
}

// CreateDataSet handles POST /api/datasets
func (dc *DataSetController) CreateDataSet(ctx *gin.Context) {
	var req models.DataSetCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user from authentication context
	user := "system"

	dataset, err := dc.datasetService.CreateDataSet(ctx.Request.Context(), &req, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Dataset created successfully",
		"data":    dataset,
	})
}

// ListDataSets handles GET /api/datasets
func (dc *DataSetController) ListDataSets(ctx *gin.Context) {
	datasets, err := dc.datasetService.ListDataSets(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  datasets,
		"count": len(datasets),
	})
}

// GetDataSet handles GET /api/datasets/:id
func (dc *DataSetController) GetDataSet(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "dataset ID is required"})
		return
	}

	dataset, err := dc.datasetService.GetDataSet(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": dataset})
}

// GetDataSetByName handles GET /api/datasets/name/:name
func (dc *DataSetController) GetDataSetByName(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "dataset name is required"})
		return
	}

	dataset, err := dc.datasetService.GetDataSetByName(ctx.Request.Context(), name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": dataset})
}

// UpdateDataSet handles PUT /api/datasets/:id
func (dc *DataSetController) UpdateDataSet(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "dataset ID is required"})
		return
	}

	var req models.DataSetUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user from authentication context
	user := "system"

	dataset, err := dc.datasetService.UpdateDataSet(ctx.Request.Context(), id, &req, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Dataset updated successfully",
		"data":    dataset,
	})
}

// DeleteDataSet handles DELETE /api/datasets/:id
func (dc *DataSetController) DeleteDataSet(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "dataset ID is required"})
		return
	}

	if err := dc.datasetService.DeleteDataSet(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Dataset deleted successfully"})
}

// SearchDataSets handles GET /api/datasets/search?q=query
func (dc *DataSetController) SearchDataSets(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	datasets, err := dc.datasetService.SearchDataSets(ctx.Request.Context(), query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  datasets,
		"count": len(datasets),
	})
}
