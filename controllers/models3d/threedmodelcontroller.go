// Copyright 2025 IAC. All Rights Reserved.

package models3d

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

var (
	threeDModelService      *services.ThreeDModelService
	threeDModelAssetService *services.ThreeDModelAssetService
)

// SetThreeDServices sets the service instances for 3D models
func SetThreeDServices(modelService *services.ThreeDModelService, assetService *services.ThreeDModelAssetService) {
	threeDModelService = modelService
	threeDModelAssetService = assetService
}

// ThreeDModelController handles versioned 3D model management
type ThreeDModelController struct{}

// =============================================================================
// 3D MODEL CRUD OPERATIONS
// =============================================================================

// CreateThreeDModel creates a new 3D model with version control
// POST /api/3dstudio/models
func (c *ThreeDModelController) CreateThreeDModel(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ThreeDModel"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("threedmodel.CreateThreeDModel", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var req models.ThreeDModelCreateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model, err := threeDModelService.CreateModel(ctx.Request.Context(), &req, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error creating 3D model: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Created 3D model: %s", model.ID))
	ctx.JSON(http.StatusOK, gin.H{"data": model})
}

// GetThreeDModel retrieves a 3D model by ID
// GET /api/3dstudio/models/:id
func (c *ThreeDModelController) GetThreeDModel(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ThreeDModel"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("threedmodel.GetThreeDModel", elapsed)
	}()

	_, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	modelID := ctx.Param("id")
	if modelID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Model ID is required"})
		return
	}

	model, err := threeDModelService.GetModel(ctx.Request.Context(), modelID)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting 3D model: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": model})
}

// ListThreeDModels retrieves all 3D models
// POST /api/3dstudio/models/list
func (c *ThreeDModelController) ListThreeDModels(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ThreeDModel"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("threedmodel.ListThreeDModels", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	// Parse optional filters
	var filters struct {
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
	}
	if len(body) > 0 {
		json.Unmarshal(body, &filters)
	}

	models, err := threeDModelService.ListModels(ctx.Request.Context(), filters.Category, filters.Tags)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error listing 3D models: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  models,
		"count": len(models),
	})
}

// UpdateThreeDModel updates a 3D model (creates new version if modelData provided)
// PUT /api/3dstudio/models/:id
func (c *ThreeDModelController) UpdateThreeDModel(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ThreeDModel"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("threedmodel.UpdateThreeDModel", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	modelID := ctx.Param("id")
	if modelID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Model ID is required"})
		return
	}

	var req models.ThreeDModelUpdateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model, err := threeDModelService.UpdateModel(ctx.Request.Context(), modelID, &req, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error updating 3D model: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Updated 3D model: %s", modelID))
	ctx.JSON(http.StatusOK, gin.H{"data": model})
}

// DeleteThreeDModel deletes a 3D model
// DELETE /api/3dstudio/models/:id
func (c *ThreeDModelController) DeleteThreeDModel(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ThreeDModel"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("threedmodel.DeleteThreeDModel", elapsed)
	}()

	_, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	modelID := ctx.Param("id")
	if modelID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Model ID is required"})
		return
	}

	if err := threeDModelService.DeleteModel(ctx.Request.Context(), modelID); err != nil {
		iLog.Error(fmt.Sprintf("Error deleting 3D model: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Deleted 3D model: %s", modelID))
	ctx.JSON(http.StatusOK, gin.H{"message": "Model deleted successfully"})
}

// SearchThreeDModels searches 3D models
// POST /api/3dstudio/models/search
func (c *ThreeDModelController) SearchThreeDModels(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ThreeDModel"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("threedmodel.SearchThreeDModels", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var req struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models, err := threeDModelService.SearchModels(ctx.Request.Context(), req.Query)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error searching 3D models: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  models,
		"count": len(models),
	})
}

// GetModelVersions retrieves all versions of a model
// GET /api/3dstudio/models/:id/versions
func (c *ThreeDModelController) GetModelVersions(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ThreeDModel"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("threedmodel.GetModelVersions", elapsed)
	}()

	_, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	modelID := ctx.Param("id")
	if modelID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Model ID is required"})
		return
	}

	versions, err := threeDModelService.ListModelVersions(ctx.Request.Context(), modelID)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting model versions: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": versions})
}

// =============================================================================
// 3D MODEL ASSET LIBRARY OPERATIONS
// =============================================================================

// CreateAsset creates a new 3D model asset
// POST /api/3dstudio/assets
func (c *ThreeDModelController) CreateAsset(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ThreeDModelAsset"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("threedmodel.CreateAsset", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var req models.ThreeDModelAssetCreateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	asset, err := threeDModelAssetService.CreateAsset(ctx.Request.Context(), &req, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error creating asset: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Created 3D asset: %s", asset.ID))
	ctx.JSON(http.StatusOK, gin.H{"data": asset})
}

// ListAssets retrieves 3D model assets
// POST /api/3dstudio/assets/list
func (c *ThreeDModelController) ListAssets(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ThreeDModelAsset"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("threedmodel.ListAssets", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var filters struct {
		AssetType  models.ThreeDModelAssetType `json:"assetType"`
		Category   string                      `json:"category"`
		Tags       []string                    `json:"tags"`
		PublicOnly bool                        `json:"publicOnly"`
	}
	if len(body) > 0 {
		json.Unmarshal(body, &filters)
	}

	assets, err := threeDModelAssetService.ListAssets(ctx.Request.Context(), filters.AssetType, filters.Category, filters.Tags, filters.PublicOnly, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error listing assets: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  assets,
		"count": len(assets),
	})
}

// GetAsset retrieves an asset by ID
// GET /api/3dstudio/assets/:id
func (c *ThreeDModelController) GetAsset(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ThreeDModelAsset"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("threedmodel.GetAsset", elapsed)
	}()

	_, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	assetID := ctx.Param("id")
	if assetID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Asset ID is required"})
		return
	}

	asset, err := threeDModelAssetService.GetAsset(ctx.Request.Context(), assetID)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting asset: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": asset})
}

// UpdateAsset updates an asset
// PUT /api/3dstudio/assets/:id
func (c *ThreeDModelController) UpdateAsset(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ThreeDModelAsset"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("threedmodel.UpdateAsset", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	assetID := ctx.Param("id")
	if assetID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Asset ID is required"})
		return
	}

	var req models.ThreeDModelAssetUpdateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling request: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	asset, err := threeDModelAssetService.UpdateAsset(ctx.Request.Context(), assetID, &req, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error updating asset: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Updated 3D asset: %s", assetID))
	ctx.JSON(http.StatusOK, gin.H{"data": asset})
}

// DeleteAsset deletes an asset
// DELETE /api/3dstudio/assets/:id
func (c *ThreeDModelController) DeleteAsset(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ThreeDModelAsset"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("threedmodel.DeleteAsset", elapsed)
	}()

	_, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	assetID := ctx.Param("id")
	if assetID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Asset ID is required"})
		return
	}

	if err := threeDModelAssetService.DeleteAsset(ctx.Request.Context(), assetID, user); err != nil {
		iLog.Error(fmt.Sprintf("Error deleting asset: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Deleted 3D asset: %s", assetID))
	ctx.JSON(http.StatusOK, gin.H{"message": "Asset deleted successfully"})
}
