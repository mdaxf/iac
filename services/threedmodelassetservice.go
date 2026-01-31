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
	Collection3DModelAssets = "3D_Model_Assets"
)

// ThreeDModelAssetService handles 3D model asset library management
type ThreeDModelAssetService struct {
	docDB documents.DocumentDB
}

// NewThreeDModelAssetService creates a new 3D model asset service
func NewThreeDModelAssetService(docDB documents.DocumentDB) *ThreeDModelAssetService {
	return &ThreeDModelAssetService{
		docDB: docDB,
	}
}

// CreateAsset creates a new 3D model asset
func (s *ThreeDModelAssetService) CreateAsset(ctx context.Context, req *models.ThreeDModelAssetCreateRequest, user string) (*models.ThreeDModelAsset, error) {
	assetID := uuid.New().String()
	now := time.Now()

	// Set default format if not provided
	format := req.Format
	if format == "" {
		format = models.ThreeDModelFormatJSON
	}

	asset := &models.ThreeDModelAsset{
		ID:              assetID,
		Name:            req.Name,
		Type:            req.Type,
		Description:     req.Description,
		Tags:            req.Tags,
		Category:        req.Category,
		AssetData:       req.AssetData,
		Format:          format,
		Thumbnail:       req.Thumbnail,
		IsPublic:        req.IsPublic,
		UsageCount:      0,
		IsAIGenerated:   req.IsAIGenerated,
		AIPrompt:        req.AIPrompt,
		DocumentID:      req.DocumentID,
		CreatedBy:       user,
		CreatedOn:       now,
		ModifiedBy:      user,
		ModifiedOn:      now,
		RowVersionStamp: 1,
	}

	// Save to database
	data, err := bson.Marshal(asset)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal 3D asset: %w", err)
	}

	var assetMap map[string]interface{}
	if err := bson.Unmarshal(data, &assetMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	if _, err := s.docDB.InsertOne(ctx, Collection3DModelAssets, assetMap); err != nil {
		return nil, fmt.Errorf("failed to insert 3D asset: %w", err)
	}

	return asset, nil
}

// GetAsset retrieves a 3D model asset by ID
func (s *ThreeDModelAssetService) GetAsset(ctx context.Context, id string) (*models.ThreeDModelAsset, error) {
	filter := bson.M{"_id": id}
	result, err := s.docDB.FindOne(ctx, Collection3DModelAssets, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get 3D asset: %w", err)
	}

	doc := bson.M(result)
	var asset models.ThreeDModelAsset
	data, err := bson.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal 3D asset: %w", err)
	}
	if err := bson.Unmarshal(data, &asset); err != nil {
		return nil, fmt.Errorf("failed to unmarshal 3D asset: %w", err)
	}

	return &asset, nil
}

// ListAssets retrieves all 3D model assets with optional filters
func (s *ThreeDModelAssetService) ListAssets(ctx context.Context, assetType models.ThreeDModelAssetType, category string, tags []string, publicOnly bool, user string) ([]models.ThreeDModelAsset, error) {
	filter := bson.M{}

	// Add type filter if provided
	if assetType != "" {
		filter["type"] = assetType
	}

	// Add category filter if provided
	if category != "" {
		filter["category"] = category
	}

	// Add tags filter if provided
	if len(tags) > 0 {
		filter["tags"] = bson.M{"$in": tags}
	}

	// Add public/ownership filter
	if publicOnly {
		filter["isPublic"] = true
	} else if user != "" {
		// Show public assets and user's own assets
		filter["$or"] = []bson.M{
			{"isPublic": true},
			{"createdby": user},
		}
	}

	results, err := s.docDB.FindMany(ctx, Collection3DModelAssets, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list 3D assets: %w", err)
	}

	var assets []models.ThreeDModelAsset
	for _, result := range results {
		doc := bson.M(result)
		var a models.ThreeDModelAsset
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &a); err != nil {
			continue
		}
		assets = append(assets, a)
	}

	return assets, nil
}

// UpdateAsset updates a 3D model asset
func (s *ThreeDModelAssetService) UpdateAsset(ctx context.Context, id string, req *models.ThreeDModelAssetUpdateRequest, user string) (*models.ThreeDModelAsset, error) {
	// Get existing asset
	existing, err := s.GetAsset(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check ownership (only creator can update unless it's public)
	if existing.CreatedBy != user && !existing.IsPublic {
		return nil, fmt.Errorf("unauthorized: only asset owner can update")
	}

	now := time.Now()

	// Update fields
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
	}
	if req.Category != "" {
		existing.Category = req.Category
	}
	if req.AssetData != nil {
		existing.AssetData = req.AssetData
	}
	if req.Thumbnail != "" {
		existing.Thumbnail = req.Thumbnail
	}
	existing.IsPublic = req.IsPublic

	existing.ModifiedBy = user
	existing.ModifiedOn = now
	existing.RowVersionStamp++

	// Save to database
	data, err := bson.Marshal(existing)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal 3D asset: %w", err)
	}

	var assetMap map[string]interface{}
	if err := bson.Unmarshal(data, &assetMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	filter := bson.M{"_id": id}
	if err := s.docDB.UpdateOne(ctx, Collection3DModelAssets, filter, bson.M{"$set": assetMap}); err != nil {
		return nil, fmt.Errorf("failed to update 3D asset: %w", err)
	}

	return existing, nil
}

// DeleteAsset deletes a 3D model asset
func (s *ThreeDModelAssetService) DeleteAsset(ctx context.Context, id string, user string) error {
	// Get asset to check ownership
	asset, err := s.GetAsset(ctx, id)
	if err != nil {
		return err
	}

	// Check ownership
	if asset.CreatedBy != user && !asset.IsPublic {
		return fmt.Errorf("unauthorized: only asset owner can delete")
	}

	filter := bson.M{"_id": id}
	if err := s.docDB.DeleteOne(ctx, Collection3DModelAssets, filter); err != nil {
		return fmt.Errorf("failed to delete 3D asset: %w", err)
	}
	return nil
}

// IncrementUsageCount increments the usage count when an asset is used
func (s *ThreeDModelAssetService) IncrementUsageCount(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$inc": bson.M{"usageCount": 1}}
	if err := s.docDB.UpdateOne(ctx, Collection3DModelAssets, filter, update); err != nil {
		return fmt.Errorf("failed to increment usage count: %w", err)
	}
	return nil
}

// SearchAssets searches 3D model assets by name, description, or tags
func (s *ThreeDModelAssetService) SearchAssets(ctx context.Context, query string, user string) ([]models.ThreeDModelAsset, error) {
	filter := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"name": bson.M{"$regex": query, "$options": "i"}},
					{"description": bson.M{"$regex": query, "$options": "i"}},
					{"tags": bson.M{"$regex": query, "$options": "i"}},
					{"category": bson.M{"$regex": query, "$options": "i"}},
				},
			},
			{
				"$or": []bson.M{
					{"isPublic": true},
					{"createdby": user},
				},
			},
		},
	}

	results, err := s.docDB.FindMany(ctx, Collection3DModelAssets, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search 3D assets: %w", err)
	}

	var assets []models.ThreeDModelAsset
	for _, result := range results {
		doc := bson.M(result)
		var a models.ThreeDModelAsset
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &a); err != nil {
			continue
		}
		assets = append(assets, a)
	}

	return assets, nil
}

// GetPopularAssets retrieves the most used assets
func (s *ThreeDModelAssetService) GetPopularAssets(ctx context.Context, limit int, user string) ([]models.ThreeDModelAsset, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"isPublic": true},
			{"createdby": user},
		},
	}

	// Sort by usage count descending
	options := &documents.FindOptions{
		Sort:  map[string]int{"usageCount": -1},
		Limit: int64(limit),
	}

	results, err := s.docDB.FindMany(ctx, Collection3DModelAssets, filter, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular assets: %w", err)
	}

	var assets []models.ThreeDModelAsset
	for _, result := range results {
		doc := bson.M(result)
		var a models.ThreeDModelAsset
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &a); err != nil {
			continue
		}
		assets = append(assets, a)
	}

	return assets, nil
}
