package plantstudio

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// PlantStudioController handles plant studio API requests
type PlantStudioController struct {
	plantStudioService *services.PlantStudioService
}

// NewPlantStudioController creates a new plant studio controller
func NewPlantStudioController(docDB documents.DocumentDB) *PlantStudioController {
	return &PlantStudioController{
		plantStudioService: services.NewPlantStudioService(docDB),
	}
}

// =====================================================
// Plant Model Endpoints
// =====================================================

// ListPlantModels handles GET /api/plant-studio/models
func (pc *PlantStudioController) ListPlantModels(ctx *gin.Context) {
	plantModels, err := pc.plantStudioService.ListPlantModels(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  plantModels,
		"count": len(plantModels),
	})
}

// GetPlantModel handles GET /api/plant-studio/models/:id
func (pc *PlantStudioController) GetPlantModel(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "model ID is required"})
		return
	}

	plantModel, err := pc.plantStudioService.GetPlantModel(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": plantModel})
}

// SavePlantModel handles POST/PUT /api/plant-studio/models
func (pc *PlantStudioController) SavePlantModel(ctx *gin.Context) {
	var plantModel models.PlantModel
	if err := ctx.ShouldBindJSON(&plantModel); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user from authentication context
	user := "system"

	if err := pc.plantStudioService.SavePlantModel(ctx.Request.Context(), &plantModel, user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Plant model saved successfully",
		"data":    plantModel,
	})
}

// DeletePlantModel handles DELETE /api/plant-studio/models/:id
func (pc *PlantStudioController) DeletePlantModel(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "model ID is required"})
		return
	}

	if err := pc.plantStudioService.DeletePlantModel(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Plant model deleted successfully"})
}

// =====================================================
// Plant Asset Endpoints
// =====================================================

// ListPlantAssets handles GET /api/plant-studio/assets
func (pc *PlantStudioController) ListPlantAssets(ctx *gin.Context) {
	assets, err := pc.plantStudioService.ListPlantAssets(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  assets,
		"count": len(assets),
	})
}

// GetPlantAsset handles GET /api/plant-studio/assets/:id
func (pc *PlantStudioController) GetPlantAsset(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "asset ID is required"})
		return
	}

	asset, err := pc.plantStudioService.GetPlantAsset(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": asset})
}

// SavePlantAsset handles POST/PUT /api/plant-studio/assets
func (pc *PlantStudioController) SavePlantAsset(ctx *gin.Context) {
	var asset models.PlantModelAsset
	if err := ctx.ShouldBindJSON(&asset); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user from authentication context
	user := "system"

	if err := pc.plantStudioService.SavePlantAsset(ctx.Request.Context(), &asset, user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Plant asset saved successfully",
		"data":    asset,
	})
}

// DeletePlantAsset handles DELETE /api/plant-studio/assets/:id
func (pc *PlantStudioController) DeletePlantAsset(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "asset ID is required"})
		return
	}

	if err := pc.plantStudioService.DeletePlantAsset(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Plant asset deleted successfully"})
}
