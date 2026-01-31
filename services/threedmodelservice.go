package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/models"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	Collection3DModels = "3D_Models"
)

// ThreeDModelService handles 3D model storage and version control
type ThreeDModelService struct {
	docDB documents.DocumentDB
}

// NewThreeDModelService creates a new 3D model service
func NewThreeDModelService(docDB documents.DocumentDB) *ThreeDModelService {
	return &ThreeDModelService{
		docDB: docDB,
	}
}

// CreateModel creates a new 3D model with simplified document structure
func (s *ThreeDModelService) CreateModel(ctx context.Context, req *models.ThreeDModelCreateRequest, user string) (*models.ThreeDModel, error) {
	modelID := uuid.New().String()
	now := time.Now()

	// Set default format if not provided
	format := req.Format
	if format == "" {
		format = models.ThreeDModelFormatJSON
	}

	// Extract objects array from modelData
	var objects []map[string]interface{}
	var metadata map[string]interface{}

	if req.ModelData != nil {
		// If modelData has objects array, extract it
		if objArray, ok := req.ModelData["objects"].([]interface{}); ok {
			for _, obj := range objArray {
				if objMap, ok := obj.(map[string]interface{}); ok {
					objects = append(objects, objMap)
				}
			}
		}
		// Store other modelData as metadata
		metadata = make(map[string]interface{})
		for k, v := range req.ModelData {
			if k != "objects" {
				metadata[k] = v
			}
		}
	}

	// Calculate object count and extract materials
	objectCount := len(objects)
	materials := extractMaterialsFromObjects(objects)

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = models.ThreeDModelStatusDraft
	}

	model := &models.ThreeDModel{
		ID:              modelID,
		Name:            req.Name,
		Description:     req.Description,
		Tags:            req.Tags,
		Category:        req.Category,
		Thumbnail:       req.Thumbnail,
		Status:          status,
		CurrentVersion:  1,
		Versions:        nil, // No longer using versions array
		ExportedFormats: make(map[models.ThreeDModelFormat]string),
		ObjectCount:     objectCount,
		Materials:       materials,
		CreatedBy:       user,
		CreatedOn:       now,
		ModifiedBy:      user,
		ModifiedOn:      now,
		RowVersionStamp: 1,
	}

	// Save to database
	data, err := bson.Marshal(model)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal 3D model: %w", err)
	}

	var modelMap map[string]interface{}
	if err := bson.Unmarshal(data, &modelMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// Add the simplified structure fields directly
	modelMap["objects"] = objects
	modelMap["metadata"] = metadata
	modelMap["format"] = format
	modelMap["status"] = status
	modelMap["isdefault"] = true
	modelMap["version"] = 1
	modelMap["isAIGenerated"] = req.IsAIGenerated
	modelMap["aiPrompt"] = req.AIPrompt
	modelMap["generationType"] = "manual"
	if req.IsAIGenerated {
		modelMap["generationType"] = "text-to-3d"
	}
	modelMap["isActive"] = true

	if _, err := s.docDB.InsertOne(ctx, Collection3DModels, modelMap); err != nil {
		return nil, fmt.Errorf("failed to insert 3D model: %w", err)
	}

	return model, nil
}

// Helper function to extract materials from objects array
func extractMaterialsFromObjects(objects []map[string]interface{}) []string {
	materialsMap := make(map[string]bool)
	var materials []string

	for _, obj := range objects {
		if material, ok := obj["material"].(string); ok && material != "" {
			if !materialsMap[material] {
				materialsMap[material] = true
				materials = append(materials, material)
			}
		}
		// Also check for material object
		if materialObj, ok := obj["material"].(map[string]interface{}); ok {
			if name, ok := materialObj["name"].(string); ok && name != "" {
				if !materialsMap[name] {
					materialsMap[name] = true
					materials = append(materials, name)
				}
			}
		}
	}

	return materials
}

// GetModel retrieves a 3D model by ID with simplified structure
func (s *ThreeDModelService) GetModel(ctx context.Context, id string) (*models.ThreeDModel, error) {
	filter := bson.M{"_id": id, "isActive": bson.M{"$ne": false}}
	result, err := s.docDB.FindOne(ctx, Collection3DModels, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get 3D model: %w", err)
	}

	doc := bson.M(result)
	var model models.ThreeDModel
	data, err := bson.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal 3D model: %w", err)
	}
	if err := bson.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to unmarshal 3D model: %w", err)
	}

	// For backward compatibility, reconstruct modelData from objects and metadata
	if objects, ok := doc["objects"]; ok {
		modelData := make(map[string]interface{})
		modelData["objects"] = objects

		// Add metadata if exists
		if metadata, ok := doc["metadata"].(map[string]interface{}); ok {
			for k, v := range metadata {
				modelData[k] = v
			}
		}

		// Create a version structure for backward compatibility
		// Get format from document
		format := models.ThreeDModelFormatJSON
		if fmt, ok := doc["format"].(string); ok {
			format = models.ThreeDModelFormat(fmt)
		}

		version := models.ThreeDModelVersion{
			Version:       1,
			IsDefault:     true,
			ModelData:     modelData,
			Format:        format,
			IsAIGenerated: false,
			CreatedBy:     model.CreatedBy,
			CreatedOn:     model.CreatedOn,
		}

		if isAI, ok := doc["isAIGenerated"].(bool); ok {
			version.IsAIGenerated = isAI
		}
		if prompt, ok := doc["aiPrompt"].(string); ok {
			version.AIPrompt = prompt
		}

		model.Versions = []models.ThreeDModelVersion{version}
		model.CurrentVersion = 1
	}

	return &model, nil
}

// ListModels retrieves all active 3D models with simplified structure
func (s *ThreeDModelService) ListModels(ctx context.Context, category string, tags []string) ([]models.ThreeDModel, error) {
	filter := bson.M{"isActive": bson.M{"$ne": false}}

	// Add category filter if provided
	if category != "" {
		filter["category"] = category
	}

	// Add tags filter if provided
	if len(tags) > 0 {
		filter["tags"] = bson.M{"$in": tags}
	}

	results, err := s.docDB.FindMany(ctx, Collection3DModels, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list 3D models: %w", err)
	}

	var modelList []models.ThreeDModel
	for _, result := range results {
		doc := bson.M(result)
		var m models.ThreeDModel
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &m); err != nil {
			continue
		}

		// Reconstruct modelData for backward compatibility
		if objects, ok := doc["objects"]; ok {
			modelData := make(map[string]interface{})
			modelData["objects"] = objects

			if metadata, ok := doc["metadata"].(map[string]interface{}); ok {
				for k, v := range metadata {
					modelData[k] = v
				}
			}

			// Get format from document
			format := models.ThreeDModelFormatJSON
			if fmt, ok := doc["format"].(string); ok {
				format = models.ThreeDModelFormat(fmt)
			}

			version := models.ThreeDModelVersion{
				Version:       1,
				IsDefault:     true,
				ModelData:     modelData,
				Format:        format,
				IsAIGenerated: false,
				CreatedBy:     m.CreatedBy,
				CreatedOn:     m.CreatedOn,
			}

			if isAI, ok := doc["isAIGenerated"].(bool); ok {
				version.IsAIGenerated = isAI
			}
			if prompt, ok := doc["aiPrompt"].(string); ok {
				version.AIPrompt = prompt
			}

			m.Versions = []models.ThreeDModelVersion{version}
			m.CurrentVersion = 1
		}

		modelList = append(modelList, m)
	}

	return modelList, nil
}

// UpdateModel updates a 3D model with simplified structure
func (s *ThreeDModelService) UpdateModel(ctx context.Context, id string, req *models.ThreeDModelUpdateRequest, user string) (*models.ThreeDModel, error) {
	// Get existing model
	existing, err := s.GetModel(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	updateFields := bson.M{}

	// Update metadata fields
	if req.Name != "" {
		existing.Name = req.Name
		updateFields["name"] = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
		updateFields["description"] = req.Description
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
		updateFields["tags"] = req.Tags
	}
	if req.Category != "" {
		existing.Category = req.Category
		updateFields["category"] = req.Category
	}
	if req.Status != "" {
		existing.Status = req.Status
		updateFields["status"] = req.Status
	}
	if req.Thumbnail != "" {
		existing.Thumbnail = req.Thumbnail
		updateFields["thumbnail"] = req.Thumbnail
	}

	// Update objects array if model data is provided
	if req.ModelData != nil {
		var objects []map[string]interface{}
		var metadata map[string]interface{}

		// Extract objects array from modelData
		if objArray, ok := req.ModelData["objects"].([]interface{}); ok {
			for _, obj := range objArray {
				if objMap, ok := obj.(map[string]interface{}); ok {
					objects = append(objects, objMap)
				}
			}
		}

		// Store other modelData as metadata
		metadata = make(map[string]interface{})
		for k, v := range req.ModelData {
			if k != "objects" {
				metadata[k] = v
			}
		}

		// Calculate object count and extract materials
		objectCount := len(objects)
		materials := extractMaterialsFromObjects(objects)

		updateFields["objects"] = objects
		updateFields["metadata"] = metadata
		updateFields["objectCount"] = objectCount
		updateFields["materials"] = materials

		existing.ObjectCount = objectCount
		existing.Materials = materials
	}

	// Update audit fields
	updateFields["modifiedby"] = user
	updateFields["modifiedon"] = now
	updateFields["rowversionstamp"] = existing.RowVersionStamp + 1

	existing.ModifiedBy = user
	existing.ModifiedOn = now
	existing.RowVersionStamp++

	// Save to database
	filter := bson.M{"_id": id}
	if err := s.docDB.UpdateOne(ctx, Collection3DModels, filter, bson.M{"$set": updateFields}); err != nil {
		return nil, fmt.Errorf("failed to update 3D model: %w", err)
	}

	// Reconstruct the version for backward compatibility
	if len(existing.Versions) > 0 {
		existing.Versions[0].ModelData = req.ModelData
	}

	return existing, nil
}

// DeleteModel soft deletes a 3D model by setting isActive to false
func (s *ThreeDModelService) DeleteModel(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"isActive": false}}
	if err := s.docDB.UpdateOne(ctx, Collection3DModels, filter, update); err != nil {
		return fmt.Errorf("failed to delete 3D model: %w", err)
	}
	return nil
}

// GetModelVersion retrieves a specific version of a 3D model
func (s *ThreeDModelService) GetModelVersion(ctx context.Context, id string, version int) (*models.ThreeDModelVersion, error) {
	model, err := s.GetModel(ctx, id)
	if err != nil {
		return nil, err
	}

	// Find the requested version
	for _, v := range model.Versions {
		if v.Version == version {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("version %d not found for model %s", version, id)
}

// ListModelVersions retrieves all versions of a 3D model
func (s *ThreeDModelService) ListModelVersions(ctx context.Context, id string) (*models.ThreeDModelVersionListResponse, error) {
	model, err := s.GetModel(ctx, id)
	if err != nil {
		return nil, err
	}

	return &models.ThreeDModelVersionListResponse{
		ModelID:  model.ID,
		Name:     model.Name,
		Versions: model.Versions,
	}, nil
}

// SearchModels searches active 3D models by name, description, or tags
func (s *ThreeDModelService) SearchModels(ctx context.Context, query string) ([]models.ThreeDModel, error) {
	filter := bson.M{
		"isActive": bson.M{"$ne": false},
		"$or": []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$regex": query, "$options": "i"}},
			{"category": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	results, err := s.docDB.FindMany(ctx, Collection3DModels, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search 3D models: %w", err)
	}

	var modelList []models.ThreeDModel
	for _, result := range results {
		doc := bson.M(result)
		var m models.ThreeDModel
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &m); err != nil {
			continue
		}

		// Reconstruct modelData for backward compatibility
		if objects, ok := doc["objects"]; ok {
			modelData := make(map[string]interface{})
			modelData["objects"] = objects

			if metadata, ok := doc["metadata"].(map[string]interface{}); ok {
				for k, v := range metadata {
					modelData[k] = v
				}
			}

			// Get format from document
			format := models.ThreeDModelFormatJSON
			if fmt, ok := doc["format"].(string); ok {
				format = models.ThreeDModelFormat(fmt)
			}

			version := models.ThreeDModelVersion{
				Version:       1,
				IsDefault:     true,
				ModelData:     modelData,
				Format:        format,
				IsAIGenerated: false,
				CreatedBy:     m.CreatedBy,
				CreatedOn:     m.CreatedOn,
			}

			if isAI, ok := doc["isAIGenerated"].(bool); ok {
				version.IsAIGenerated = isAI
			}
			if prompt, ok := doc["aiPrompt"].(string); ok {
				version.AIPrompt = prompt
			}

			m.Versions = []models.ThreeDModelVersion{version}
			m.CurrentVersion = 1
		}

		modelList = append(modelList, m)
	}

	return modelList, nil
}

// AddExportedFormat adds a reference to an exported format document
func (s *ThreeDModelService) AddExportedFormat(ctx context.Context, modelID string, format models.ThreeDModelFormat, documentID string) error {
	model, err := s.GetModel(ctx, modelID)
	if err != nil {
		return err
	}

	if model.ExportedFormats == nil {
		model.ExportedFormats = make(map[models.ThreeDModelFormat]string)
	}

	model.ExportedFormats[format] = documentID

	// Update in database
	filter := bson.M{"_id": modelID}
	update := bson.M{"$set": bson.M{"exportedFormats": model.ExportedFormats}}
	if err := s.docDB.UpdateOne(ctx, Collection3DModels, filter, update); err != nil {
		return fmt.Errorf("failed to update exported formats: %w", err)
	}

	return nil
}
