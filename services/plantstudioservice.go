package services

import (
	"context"
	"fmt"
	"time"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/models"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	CollectionPlantModels      = "Plant_Models"
	CollectionPlantModelAssets = "Plant_Model_Assets"
)

// PlantStudioService handles plant studio operations
type PlantStudioService struct {
	docDB documents.DocumentDB
}

// NewPlantStudioService creates a new plant studio service
func NewPlantStudioService(docDB documents.DocumentDB) *PlantStudioService {
	return &PlantStudioService{
		docDB: docDB,
	}
}

// =====================================================
// Plant Model Operations
// =====================================================

// ListPlantModels retrieves all plant models
func (ps *PlantStudioService) ListPlantModels(ctx context.Context) ([]models.PlantModel, error) {
	filter := bson.M{"active": true}
	results, err := ps.docDB.FindMany(ctx, CollectionPlantModels, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list plant models: %w", err)
	}

	var plantModels []models.PlantModel
	for _, result := range results {
		// result is already map[string]interface{}, convert to bson.M
		doc := bson.M(result)
		var model models.PlantModel
		// Convert bson.M to struct using bson marshal/unmarshal
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &model); err != nil {
			continue
		}
		plantModels = append(plantModels, model)
	}

	return plantModels, nil
}

// GetPlantModel retrieves a plant model by ID
func (ps *PlantStudioService) GetPlantModel(ctx context.Context, id string) (*models.PlantModel, error) {
	filter := bson.M{"_id": id}
	result, err := ps.docDB.FindOne(ctx, CollectionPlantModels, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get plant model: %w", err)
	}

	// result is already map[string]interface{}, convert to bson.M
	doc := bson.M(result)
	var model models.PlantModel
	data, err := bson.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plant model: %w", err)
	}
	if err := bson.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plant model: %w", err)
	}
	return &model, nil
}

// SavePlantModel creates or updates a plant model
func (ps *PlantStudioService) SavePlantModel(ctx context.Context, model *models.PlantModel, user string) error {
	fmt.Printf("[PlantStudioService] SavePlantModel called - isNew: %v, ID: %s, Name: %s\n", model.ID == "", model.ID, model.Name)

	now := time.Now()
	isNew := model.ID == ""

	if isNew {
		// New model - generate ID
		model.ID = fmt.Sprintf("plant-model-%d", now.UnixNano())
		model.CreatedBy = user
		model.CreatedOn = now
		model.Active = true
		model.RowVersionStamp = 1
		fmt.Printf("[PlantStudioService] Generated new model ID: %s\n", model.ID)
	} else {
		// Existing model - increment version
		model.RowVersionStamp++
		fmt.Printf("[PlantStudioService] Updating existing model ID: %s\n", model.ID)
	}

	model.ModifiedBy = user
	model.ModifiedOn = now

	// Convert model to bson.M
	data, err := bson.Marshal(model)
	if err != nil {
		fmt.Printf("[PlantStudioService] Marshal error: %v\n", err)
		return fmt.Errorf("failed to marshal plant model: %w", err)
	}

	var doc bson.M
	if err := bson.Unmarshal(data, &doc); err != nil {
		fmt.Printf("[PlantStudioService] Unmarshal error: %v\n", err)
		return fmt.Errorf("failed to unmarshal to bson.M: %w", err)
	}

	fmt.Printf("[PlantStudioService] Document to save: %+v\n", doc)

	if isNew {
		// Insert new document
		fmt.Printf("[PlantStudioService] Calling InsertOne for collection: %s\n", CollectionPlantModels)
		result, err := ps.docDB.InsertOne(ctx, CollectionPlantModels, doc)
		if err != nil {
			fmt.Printf("[PlantStudioService] InsertOne error: %v\n", err)
			return fmt.Errorf("failed to insert plant model: %w", err)
		}
		fmt.Printf("[PlantStudioService] InsertOne success, result: %+v\n", result)
	} else {
		// Update existing document
		filter := bson.M{"_id": model.ID}
		update := bson.M{"$set": doc}
		fmt.Printf("[PlantStudioService] Calling UpdateOne for collection: %s\n", CollectionPlantModels)
		if err := ps.docDB.UpdateOne(ctx, CollectionPlantModels, filter, update); err != nil {
			fmt.Printf("[PlantStudioService] UpdateOne error: %v\n", err)
			return fmt.Errorf("failed to update plant model: %w", err)
		}
		fmt.Printf("[PlantStudioService] UpdateOne success\n")
	}

	fmt.Printf("[PlantStudioService] SavePlantModel completed successfully\n")
	return nil
}

// DeletePlantModel soft deletes a plant model
func (ps *PlantStudioService) DeletePlantModel(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"active":      false,
			"modifiedon":  time.Now(),
		},
	}

	if err := ps.docDB.UpdateOne(ctx, CollectionPlantModels, filter, update); err != nil {
		return fmt.Errorf("failed to delete plant model: %w", err)
	}

	return nil
}

// =====================================================
// Plant Model Asset Operations
// =====================================================

// ListPlantAssets retrieves all plant assets
func (ps *PlantStudioService) ListPlantAssets(ctx context.Context) ([]models.PlantModelAsset, error) {
	filter := bson.M{"active": true}
	results, err := ps.docDB.FindMany(ctx, CollectionPlantModelAssets, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list plant assets: %w", err)
	}

	var assets []models.PlantModelAsset
	for _, result := range results {
		// result is already map[string]interface{}, convert to bson.M
		doc := bson.M(result)
		var asset models.PlantModelAsset
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &asset); err != nil {
			continue
		}
		assets = append(assets, asset)
	}

	return assets, nil
}

// GetPlantAsset retrieves a plant asset by ID
func (ps *PlantStudioService) GetPlantAsset(ctx context.Context, id string) (*models.PlantModelAsset, error) {
	filter := bson.M{"_id": id}
	result, err := ps.docDB.FindOne(ctx, CollectionPlantModelAssets, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get plant asset: %w", err)
	}

	// result is already map[string]interface{}, convert to bson.M
	doc := bson.M(result)
	var asset models.PlantModelAsset
	data, err := bson.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plant asset: %w", err)
	}
	if err := bson.Unmarshal(data, &asset); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plant asset: %w", err)
	}
	return &asset, nil
}

// SavePlantAsset creates or updates a plant asset
func (ps *PlantStudioService) SavePlantAsset(ctx context.Context, asset *models.PlantModelAsset, user string) error {
	now := time.Now()
	isNew := asset.ID == ""

	if isNew {
		// New asset - generate ID
		asset.ID = fmt.Sprintf("plant-asset-%d", now.UnixNano())
		asset.CreatedBy = user
		asset.CreatedOn = now
		asset.Active = true
		asset.RowVersionStamp = 1
	} else {
		// Existing asset - increment version
		asset.RowVersionStamp++
	}

	asset.ModifiedBy = user
	asset.ModifiedOn = now

	// Convert asset to bson.M
	data, err := bson.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal plant asset: %w", err)
	}

	var doc bson.M
	if err := bson.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("failed to unmarshal to bson.M: %w", err)
	}

	if isNew {
		// Insert new document
		if _, err := ps.docDB.InsertOne(ctx, CollectionPlantModelAssets, doc); err != nil {
			return fmt.Errorf("failed to insert plant asset: %w", err)
		}
	} else {
		// Update existing document
		filter := bson.M{"_id": asset.ID}
		update := bson.M{"$set": doc}
		if err := ps.docDB.UpdateOne(ctx, CollectionPlantModelAssets, filter, update); err != nil {
			return fmt.Errorf("failed to update plant asset: %w", err)
		}
	}

	return nil
}

// DeletePlantAsset soft deletes a plant asset
func (ps *PlantStudioService) DeletePlantAsset(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"active":      false,
			"modifiedon":  time.Now(),
		},
	}

	if err := ps.docDB.UpdateOne(ctx, CollectionPlantModelAssets, filter, update); err != nil {
		return fmt.Errorf("failed to delete plant asset: %w", err)
	}

	return nil
}
