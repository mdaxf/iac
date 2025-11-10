// Copyright 2025 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models3d

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Models3DController struct {
}

// RequestBody for text-to-3D and image-to-3D generation
type GenerateRequest struct {
	Type        string                 `json:"type"`        // "text" or "image"
	Prompt      string                 `json:"prompt"`      // Text description
	ImageData   string                 `json:"imageData"`   // Base64 encoded image
	Parameters  map[string]interface{} `json:"parameters"`  // Additional parameters
	GeneratedBy string                 `json:"generatedBy"` // User who generated
}

// Model3D represents a 3D model document in MongoDB collection: 3D_Models
type Model3D struct {
	ID          primitive.ObjectID     `json:"_id" bson:"_id,omitempty"`
	Name        string                 `json:"name" bson:"name"`
	Type        string                 `json:"type" bson:"type"` // "text-to-3d", "image-to-3d", "manual"
	Prompt      string                 `json:"prompt" bson:"prompt"`
	ImageData   string                 `json:"imageData" bson:"imageData,omitempty"`
	Status      string                 `json:"status" bson:"status"` // "pending", "processing", "completed", "failed"
	Progress    int                    `json:"progress" bson:"progress"`
	ModelURL    string                 `json:"modelUrl" bson:"modelUrl,omitempty"`
	ThumbnailURL string                `json:"thumbnailUrl" bson:"thumbnailUrl,omitempty"`
	Format      string                 `json:"format" bson:"format"` // "glb", "gltf", "obj", etc.
	FileSize    int64                  `json:"fileSize" bson:"fileSize"`
	Error       string                 `json:"error" bson:"error,omitempty"`
	Parameters  map[string]interface{} `json:"parameters" bson:"parameters,omitempty"`
	GeneratedBy string                 `json:"generatedBy" bson:"generatedBy"`
	CreatedOn   time.Time              `json:"createdOn" bson:"createdOn"`
	ModifiedOn  time.Time              `json:"modifiedOn" bson:"modifiedOn"`
	CompletedOn *time.Time             `json:"completedOn" bson:"completedOn,omitempty"`
}

const COLLECTION_NAME = "3D_Models"

// ListModels retrieves list of all 3D models
// GET /api/3dmodels/list
func (c *Models3DController) ListModels(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Models3D"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("models3d.ListModels", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("List 3D models request body: %s", body))

	// Query all models, sorted by creation date descending
	filter := bson.M{}
	options := bson.M{"createdOn": -1}

	models, err := documents.DocDBCon.QueryCollection(COLLECTION_NAME, filter, options)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error querying 3D models: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Found %d 3D models", len(models)))
	ctx.JSON(http.StatusOK, gin.H{"data": models})
}

// GetModelByID retrieves a specific 3D model by ID
// GET /api/3dmodels/:id
func (c *Models3DController) GetModelByID(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Models3D"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("models3d.GetModelByID", elapsed)
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

	iLog.Debug(fmt.Sprintf("Get 3D model by ID: %s", modelID))

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(modelID)
	if err != nil {
		iLog.Error(fmt.Sprintf("Invalid model ID: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model ID"})
		return
	}

	filter := bson.M{"_id": objectID}
	models, err := documents.DocDBCon.QueryCollection(COLLECTION_NAME, filter, nil)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error querying 3D model: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(models) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}

	iLog.Debug(fmt.Sprintf("Found 3D model: %s", logger.ConvertJson(models[0])))
	ctx.JSON(http.StatusOK, gin.H{"data": models[0]})
}

// GenerateTextTo3D starts text-to-3D model generation
// POST /api/3dmodels/generate/text
func (c *Models3DController) GenerateTextTo3D(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Models3D"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("models3d.GenerateTextTo3D", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("Generate text-to-3D request body: %s", body))

	var request GenerateRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Prompt == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Prompt is required"})
		return
	}

	// Create new model document
	now := time.Now()
	model := Model3D{
		ID:          primitive.NewObjectID(),
		Name:        fmt.Sprintf("Text-to-3D: %s", request.Prompt[:min(30, len(request.Prompt))]),
		Type:        "text-to-3d",
		Prompt:      request.Prompt,
		Status:      "pending",
		Progress:    0,
		Format:      "glb",
		Parameters:  request.Parameters,
		GeneratedBy: user,
		CreatedOn:   now,
		ModifiedOn:  now,
	}

	// Insert into MongoDB
	modelMap := bson.M{
		"_id":         model.ID,
		"name":        model.Name,
		"type":        model.Type,
		"prompt":      model.Prompt,
		"status":      model.Status,
		"progress":    model.Progress,
		"format":      model.Format,
		"parameters":  model.Parameters,
		"generatedBy": model.GeneratedBy,
		"createdOn":   model.CreatedOn,
		"modifiedOn":  model.ModifiedOn,
	}

	_, err = documents.DocDBCon.InsertCollection(COLLECTION_NAME, modelMap)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error inserting 3D model: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Created text-to-3D generation job: %s", model.ID.Hex()))

	// Start async generation process in background goroutine
	go c.processTextTo3DGeneration(model.ID.Hex(), request.Prompt, iLog)

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":      model.ID.Hex(),
			"status":  model.Status,
			"message": "Generation job created successfully",
		},
	})
}

// GenerateImageTo3D starts image-to-3D model generation
// POST /api/3dmodels/generate/image
func (c *Models3DController) GenerateImageTo3D(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Models3D"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("models3d.GenerateImageTo3D", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("Generate image-to-3D request body length: %d", len(body)))

	var request GenerateRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.ImageData == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Image data is required"})
		return
	}

	// Create new model document
	now := time.Now()
	modelName := "Image-to-3D"
	if request.Prompt != "" {
		modelName = fmt.Sprintf("Image-to-3D: %s", request.Prompt[:min(30, len(request.Prompt))])
	}

	model := Model3D{
		ID:          primitive.NewObjectID(),
		Name:        modelName,
		Type:        "image-to-3d",
		Prompt:      request.Prompt,
		ImageData:   request.ImageData,
		Status:      "pending",
		Progress:    0,
		Format:      "glb",
		Parameters:  request.Parameters,
		GeneratedBy: user,
		CreatedOn:   now,
		ModifiedOn:  now,
	}

	// Insert into MongoDB
	modelMap := bson.M{
		"_id":         model.ID,
		"name":        model.Name,
		"type":        model.Type,
		"prompt":      model.Prompt,
		"imageData":   model.ImageData,
		"status":      model.Status,
		"progress":    model.Progress,
		"format":      model.Format,
		"parameters":  model.Parameters,
		"generatedBy": model.GeneratedBy,
		"createdOn":   model.CreatedOn,
		"modifiedOn":  model.ModifiedOn,
	}

	_, err = documents.DocDBCon.InsertCollection(COLLECTION_NAME, modelMap)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error inserting 3D model: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Created image-to-3D generation job: %s", model.ID.Hex()))

	// Start async generation process in background goroutine
	go c.processImageTo3DGeneration(model.ID.Hex(), request.ImageData, request.Prompt, iLog)

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":      model.ID.Hex(),
			"status":  model.Status,
			"message": "Generation job created successfully",
		},
	})
}

// UpdateModelStatus updates the status of a 3D model generation job
// PUT /api/3dmodels/:id/status
func (c *Models3DController) UpdateModelStatus(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Models3D"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("models3d.UpdateModelStatus", elapsed)
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

	iLog.Debug(fmt.Sprintf("Update model status: %s, body: %s", modelID, body))

	var updateData map[string]interface{}
	err = json.Unmarshal(body, &updateData)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(modelID)
	if err != nil {
		iLog.Error(fmt.Sprintf("Invalid model ID: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model ID"})
		return
	}

	// Add modifiedOn timestamp
	updateData["modifiedOn"] = time.Now()

	// If status is completed, add completedOn
	if status, ok := updateData["status"].(string); ok && status == "completed" {
		now := time.Now()
		updateData["completedOn"] = now
	}

	// Update in MongoDB
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": updateData}

	err = documents.DocDBCon.UpdateCollection(COLLECTION_NAME, filter, update, nil)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error updating 3D model: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Updated 3D model status: %s", modelID))
	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":      modelID,
			"message": "Model updated successfully",
		},
	})
}

// DeleteModel deletes a 3D model
// DELETE /api/3dmodels/:id
func (c *Models3DController) DeleteModel(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Models3D"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("models3d.DeleteModel", elapsed)
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

	iLog.Debug(fmt.Sprintf("Delete 3D model: %s", modelID))


	// Delete from MongoDB using modelID string directly
	err = documents.DocDBCon.DeleteItemFromCollection(COLLECTION_NAME, modelID)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error deleting 3D model: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Deleted 3D model: %s", modelID))
	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":      modelID,
			"message": "Model deleted successfully",
		},
	})
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// ASYNC GENERATION PROCESSING FUNCTIONS
// =============================================================================

// processTextTo3DGeneration handles async text-to-3D model generation
func (c *Models3DController) processTextTo3DGeneration(modelID string, prompt string, iLog logger.Log) {
	iLog.Info(fmt.Sprintf("[Text-to-3D] Starting generation for model: %s", modelID))

	// Update status to processing
	c.updateModelProgress(modelID, "processing", 10, "", iLog)

	// Simulate generation process (replace with actual AI service call)
	// Example AI services:
	// - Meshy AI: https://api.meshy.ai
	// - Zoo ML API: https://zoo.dev
	// - OpenAI DALL-E 3D (when available)
	// - Kaedim3D: https://www.kaedim3d.com

	// For demonstration, simulate a generation process with progress updates
	result, err := c.generateWithAIService("text-to-3d", map[string]interface{}{
		"prompt": prompt,
	}, modelID, iLog)

	if err != nil {
		iLog.Error(fmt.Sprintf("[Text-to-3D] Generation failed for model %s: %v", modelID, err))
		c.updateModelProgress(modelID, "failed", 0, err.Error(), iLog)
		return
	}

	// Save generated model file
	modelURL, fileSize, err := c.saveGeneratedModel(modelID, result, iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("[Text-to-3D] Failed to save model %s: %v", modelID, err))
		c.updateModelProgress(modelID, "failed", 100, err.Error(), iLog)
		return
	}

	// Update model with final status
	c.updateModelComplete(modelID, modelURL, fileSize, iLog)
	iLog.Info(fmt.Sprintf("[Text-to-3D] Generation completed for model: %s, URL: %s", modelID, modelURL))
}

// processImageTo3DGeneration handles async image-to-3D model generation
func (c *Models3DController) processImageTo3DGeneration(modelID string, imageData string, prompt string, iLog logger.Log) {
	iLog.Info(fmt.Sprintf("[Image-to-3D] Starting generation for model: %s", modelID))

	// Update status to processing
	c.updateModelProgress(modelID, "processing", 10, "", iLog)

	// Simulate generation process (replace with actual AI service call)
	// Example AI services for image-to-3D:
	// - Meshy AI (supports image-to-3D)
	// - CSM AI: https://csm.ai
	// - Luma AI: https://lumalabs.ai
	// - Tripo AI: https://www.tripo3d.ai

	result, err := c.generateWithAIService("image-to-3d", map[string]interface{}{
		"imageData": imageData,
		"prompt":    prompt,
	}, modelID, iLog)

	if err != nil {
		iLog.Error(fmt.Sprintf("[Image-to-3D] Generation failed for model %s: %v", modelID, err))
		c.updateModelProgress(modelID, "failed", 0, err.Error(), iLog)
		return
	}

	// Save generated model file
	modelURL, fileSize, err := c.saveGeneratedModel(modelID, result, iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("[Image-to-3D] Failed to save model %s: %v", modelID, err))
		c.updateModelProgress(modelID, "failed", 100, err.Error(), iLog)
		return
	}

	// Update model with final status
	c.updateModelComplete(modelID, modelURL, fileSize, iLog)
	iLog.Info(fmt.Sprintf("[Image-to-3D] Generation completed for model: %s, URL: %s", modelID, modelURL))
}

// generateWithAIService simulates calling an external AI service for 3D generation
// In production, replace this with actual API calls to:
// - Meshy AI, Zoo ML, Kaedim3D, CSM AI, Luma AI, Tripo AI, etc.
func (c *Models3DController) generateWithAIService(genType string, params map[string]interface{}, modelID string, iLog logger.Log) ([]byte, error) {
	iLog.Info(fmt.Sprintf("[AI Service] Starting %s generation for model: %s", genType, modelID))

	// TODO: Replace with actual AI service integration
	// Example for Meshy AI:
	// 1. POST to https://api.meshy.ai/v1/text-to-3d with API key
	// 2. Get task ID from response
	// 3. Poll GET https://api.meshy.ai/v1/text-to-3d/{task_id} for status
	// 4. Download model from result URL when status is "SUCCEEDED"

	// Example for Zoo ML API:
	// 1. POST to https://zoo.dev/api/text-to-cad with API key
	// 2. Poll for completion
	// 3. Download model file

	// Simulate progress updates
	progressSteps := []int{20, 40, 60, 80, 95}
	for _, progress := range progressSteps {
		time.Sleep(2 * time.Second) // Simulate processing time
		c.updateModelProgress(modelID, "processing", progress, "", iLog)
		iLog.Info(fmt.Sprintf("[AI Service] Progress: %d%% for model: %s", progress, modelID))
	}

	// Simulate generating a simple GLB file (replace with actual model data)
	// In production, this would be the downloaded model file from the AI service
	modelData := c.generateSimpleGLBModel(params, iLog)

	return modelData, nil
}

// generateSimpleGLBModel creates a simple GLB model for demonstration
// In production, this would be replaced by downloading from AI service
func (c *Models3DController) generateSimpleGLBModel(params map[string]interface{}, iLog logger.Log) []byte {
	// This is a minimal valid GLB file (placeholder)
	// In production, this would be actual model data from AI service

	// GLB file structure (binary glTF):
	// 12 bytes: Header (magic, version, length)
	// N bytes: JSON chunk
	// M bytes: Binary chunk (geometry data)

	// For now, return a minimal GLB file structure
	// You would replace this with actual model data from the AI service

	iLog.Info("[GLB Generator] Creating placeholder GLB model")

	// Minimal GLB header + basic glTF JSON
	glbHeader := []byte{
		0x67, 0x6C, 0x54, 0x46, // magic: "glTF"
		0x02, 0x00, 0x00, 0x00, // version: 2
		0x00, 0x00, 0x00, 0x00, // length (placeholder, will be updated)
	}

	// Minimal glTF JSON content
	jsonContent := `{
		"asset": {"version": "2.0", "generator": "IAC 3D Generator"},
		"scene": 0,
		"scenes": [{"nodes": [0]}],
		"nodes": [{"mesh": 0}],
		"meshes": [{"primitives": [{"attributes": {"POSITION": 0}, "indices": 1}]}],
		"accessors": [
			{"bufferView": 0, "componentType": 5126, "count": 8, "type": "VEC3", "max": [1,1,1], "min": [-1,-1,-1]},
			{"bufferView": 1, "componentType": 5123, "count": 36, "type": "SCALAR"}
		],
		"bufferViews": [
			{"buffer": 0, "byteOffset": 0, "byteLength": 96, "target": 34962},
			{"buffer": 0, "byteOffset": 96, "byteLength": 72, "target": 34963}
		],
		"buffers": [{"byteLength": 168}]
	}`

	// Pad JSON to 4-byte alignment
	jsonBytes := []byte(jsonContent)
	jsonLength := len(jsonBytes)
	jsonPadding := (4 - (jsonLength % 4)) % 4
	for i := 0; i < jsonPadding; i++ {
		jsonBytes = append(jsonBytes, 0x20) // space padding
	}

	// JSON chunk header
	jsonChunkHeader := make([]byte, 8)
	binary.LittleEndian.PutUint32(jsonChunkHeader[0:4], uint32(len(jsonBytes)))
	copy(jsonChunkHeader[4:8], []byte{0x4A, 0x53, 0x4F, 0x4E}) // "JSON"

	// Simple cube geometry data (binary buffer)
	bufferData := make([]byte, 168) // positions (96) + indices (72)

	// Calculate total length
	totalLength := 12 + 8 + len(jsonBytes) + 8 + len(bufferData)
	binary.LittleEndian.PutUint32(glbHeader[8:12], uint32(totalLength))

	// Binary chunk header
	binChunkHeader := make([]byte, 8)
	binary.LittleEndian.PutUint32(binChunkHeader[0:4], uint32(len(bufferData)))
	copy(binChunkHeader[4:8], []byte{0x42, 0x49, 0x4E, 0x00}) // "BIN\0"

	// Combine all parts
	result := append(glbHeader, jsonChunkHeader...)
	result = append(result, jsonBytes...)
	result = append(result, binChunkHeader...)
	result = append(result, bufferData...)

	return result
}

// saveGeneratedModel saves the generated 3D model file to storage
func (c *Models3DController) saveGeneratedModel(modelID string, modelData []byte, iLog logger.Log) (string, int64, error) {
	// Create storage directory if it doesn't exist
	storageDir := "./storage/3d_models"
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create storage directory: %v", err)
	}

	// Generate filename
	filename := fmt.Sprintf("%s.glb", modelID)
	filepath := fmt.Sprintf("%s/%s", storageDir, filename)

	// Write file to disk
	if err := os.WriteFile(filepath, modelData, 0644); err != nil {
		return "", 0, fmt.Errorf("failed to write model file: %v", err)
	}

	fileSize := int64(len(modelData))

	// Generate URL (adjust based on your server configuration)
	// In production, this might be a CDN URL or cloud storage URL
	modelURL := fmt.Sprintf("/storage/3d_models/%s", filename)

	iLog.Info(fmt.Sprintf("[Storage] Saved model file: %s (size: %d bytes)", filepath, fileSize))

	return modelURL, fileSize, nil
}

// updateModelProgress updates the generation progress in MongoDB
func (c *Models3DController) updateModelProgress(modelID string, status string, progress int, errorMsg string, iLog logger.Log) error {
	filter := bson.M{"_id": modelID}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"progress":   progress,
			"modifiedOn": time.Now(),
		},
	}

	if errorMsg != "" {
		update["$set"].(bson.M)["error"] = errorMsg
	}

	err := documents.DocDBCon.UpdateCollection(COLLECTION_NAME, filter, update, nil)
	if err != nil {
		iLog.Error(fmt.Sprintf("[Progress Update] Failed to update model %s: %v", modelID, err))
		return err
	}

	iLog.Debug(fmt.Sprintf("[Progress Update] Model %s: status=%s, progress=%d%%", modelID, status, progress))
	return nil
}

// updateModelComplete marks the model as completed with final data
func (c *Models3DController) updateModelComplete(modelID string, modelURL string, fileSize int64, iLog logger.Log) error {
	filter := bson.M{"_id": modelID}
	now := time.Now()

	update := bson.M{
		"$set": bson.M{
			"status":      "completed",
			"progress":    100,
			"modelUrl":    modelURL,
			"fileSize":    fileSize,
			"modifiedOn":  now,
			"completedOn": now,
		},
	}

	err := documents.DocDBCon.UpdateCollection(COLLECTION_NAME, filter, update, nil)
	if err != nil {
		iLog.Error(fmt.Sprintf("[Completion Update] Failed to mark model %s as complete: %v", modelID, err))
		return err
	}

	iLog.Info(fmt.Sprintf("[Completion Update] Model %s marked as completed", modelID))
	return nil
}
