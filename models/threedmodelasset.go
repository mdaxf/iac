package models

import (
	"time"
)

// ThreeDModelAssetType represents the type of 3D asset
type ThreeDModelAssetType string

const (
	ThreeDModelAssetTypePrimitive ThreeDModelAssetType = "PRIMITIVE" // Basic shapes
	ThreeDModelAssetTypeModel     ThreeDModelAssetType = "MODEL"     // Complex models
	ThreeDModelAssetTypeMaterial  ThreeDModelAssetType = "MATERIAL"  // Material presets
	ThreeDModelAssetTypeTexture   ThreeDModelAssetType = "TEXTURE"   // Textures
	ThreeDModelAssetTypeTemplate  ThreeDModelAssetType = "TEMPLATE"  // Scene templates
)

// ThreeDModelAsset represents a reusable 3D asset in the library
type ThreeDModelAsset struct {
	ID          string               `json:"_id" bson:"_id"`
	Name        string               `json:"name" bson:"name"`
	Type        ThreeDModelAssetType `json:"type" bson:"type"`
	Description string               `json:"description,omitempty" bson:"description,omitempty"`
	Tags        []string             `json:"tags,omitempty" bson:"tags,omitempty"`
	Category    string               `json:"category,omitempty" bson:"category,omitempty"`

	// Asset Data
	AssetData      map[string]interface{} `json:"assetData" bson:"assetData"` // Three.js object/material data
	Format         ThreeDModelFormat      `json:"format" bson:"format"`
	Thumbnail      string                 `json:"thumbnail,omitempty" bson:"thumbnail,omitempty"`
	PreviewURL     string                 `json:"previewUrl,omitempty" bson:"previewUrl,omitempty"` // 3D preview

	// File Reference (if asset is from imported file)
	DocumentID     string                 `json:"documentId,omitempty" bson:"documentId,omitempty"`

	// Metadata
	IsPublic       bool                   `json:"isPublic" bson:"isPublic"` // Shared with all users
	UsageCount     int                    `json:"usageCount" bson:"usageCount"` // Track how many times used
	IsAIGenerated  bool                   `json:"isAIGenerated" bson:"isAIGenerated"`
	AIPrompt       string                 `json:"aiPrompt,omitempty" bson:"aiPrompt,omitempty"`

	// Technical Details
	PolyCount      int                    `json:"polyCount,omitempty" bson:"polyCount,omitempty"` // Polygon count
	BoundingBox    map[string]interface{} `json:"boundingBox,omitempty" bson:"boundingBox,omitempty"` // Size info

	// Audit Fields
	CreatedBy       string    `json:"createdby" bson:"createdby"`
	CreatedOn       time.Time `json:"createdon" bson:"createdon"`
	ModifiedBy      string    `json:"modifiedby" bson:"modifiedby"`
	ModifiedOn      time.Time `json:"modifiedon" bson:"modifiedon"`
	RowVersionStamp int       `json:"rowversionstamp" bson:"rowversionstamp"`
}

// ThreeDModelAssetCreateRequest represents the request to create a new asset
type ThreeDModelAssetCreateRequest struct {
	Name          string                 `json:"name" binding:"required"`
	Type          ThreeDModelAssetType   `json:"type" binding:"required"`
	Description   string                 `json:"description,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Category      string                 `json:"category,omitempty"`
	AssetData     map[string]interface{} `json:"assetData" binding:"required"`
	Format        ThreeDModelFormat      `json:"format"`
	Thumbnail     string                 `json:"thumbnail,omitempty"`
	IsPublic      bool                   `json:"isPublic"`
	IsAIGenerated bool                   `json:"isAIGenerated"`
	AIPrompt      string                 `json:"aiPrompt,omitempty"`
	DocumentID    string                 `json:"documentId,omitempty"`
}

// ThreeDModelAssetUpdateRequest represents the request to update an asset
type ThreeDModelAssetUpdateRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Category    string                 `json:"category,omitempty"`
	AssetData   map[string]interface{} `json:"assetData,omitempty"`
	Thumbnail   string                 `json:"thumbnail,omitempty"`
	IsPublic    bool                   `json:"isPublic,omitempty"`
}

// ThreeDModelAssetListResponse represents the response for listing assets
type ThreeDModelAssetListResponse struct {
	Data  []ThreeDModelAsset `json:"data"`
	Count int                `json:"count"`
}
