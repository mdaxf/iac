package models

import (
	"time"
)

// =====================================================
// Plant Studio Models
// =====================================================

// PlantModel represents a plant configuration/layout
type PlantModel struct {
	ID              string                 `json:"_id" bson:"_id"`
	Name            string                 `json:"name" bson:"name"`
	Description     string                 `json:"description" bson:"description"`
	Version         string                 `json:"version" bson:"version"`

	// Plant Configuration
	Objects         []map[string]interface{} `json:"objects" bson:"objects"`         // 3D objects in the plant
	TourPoints      []map[string]interface{} `json:"tourPoints" bson:"tourPoints"`   // Camera tour points
	FloorSettings   map[string]interface{}   `json:"floorSettings" bson:"floorSettings"`
	GridSettings    map[string]interface{}   `json:"gridSettings" bson:"gridSettings"`

	// Simulation Settings
	SimulationSettings map[string]interface{} `json:"simulationSettings,omitempty" bson:"simulationSettings,omitempty"`

	// Workspace Settings
	WorkspaceSettings map[string]interface{} `json:"workspaceSettings,omitempty" bson:"workspaceSettings,omitempty"`

	// Metadata
	Tags            []string               `json:"tags" bson:"tags"`
	Category        string                 `json:"category" bson:"category"`
	Thumbnail       string                 `json:"thumbnail,omitempty" bson:"thumbnail,omitempty"`

	// Status
	Active          bool                   `json:"active" bson:"active"`
	Published       bool                   `json:"published" bson:"published"`

	// Audit Fields
	CreatedBy       string                 `json:"createdby" bson:"createdby"`
	CreatedOn       time.Time              `json:"createdon" bson:"createdon"`
	ModifiedBy      string                 `json:"modifiedby" bson:"modifiedby"`
	ModifiedOn      time.Time              `json:"modifiedon" bson:"modifiedon"`
	RowVersionStamp int                    `json:"rowversionstamp" bson:"rowversionstamp"`
}

// PlantModelAsset represents reusable 3D assets
type PlantModelAsset struct {
	ID              string                 `json:"_id" bson:"_id"`
	Name            string                 `json:"name" bson:"name"`
	Description     string                 `json:"description" bson:"description"`
	AssetType       string                 `json:"assetType" bson:"assetType"` // MACHINE_CNC, ROBOT_ARM, CONVEYOR, etc.

	// Asset Geometry
	PrimitiveParts  []map[string]interface{} `json:"primitiveParts" bson:"primitiveParts"` // Primitive shapes that make up the asset
	BoundingBox     map[string]interface{}   `json:"boundingBox,omitempty" bson:"boundingBox,omitempty"`
	DefaultScale    []float64                `json:"defaultScale,omitempty" bson:"defaultScale,omitempty"`

	// Asset Configuration
	MetricConfigs   []map[string]interface{} `json:"metricConfigs,omitempty" bson:"metricConfigs,omitempty"` // Configurable metrics
	ActionConfigs   []map[string]interface{} `json:"actionConfigs,omitempty" bson:"actionConfigs,omitempty"` // Actions the asset can perform
	DefaultColor    string                   `json:"defaultColor,omitempty" bson:"defaultColor,omitempty"`

	// Metadata
	Tags            []string               `json:"tags" bson:"tags"`
	Category        string                 `json:"category" bson:"category"`
	Thumbnail       string                 `json:"thumbnail,omitempty" bson:"thumbnail,omitempty"`

	// Status
	Active          bool                   `json:"active" bson:"active"`
	Published       bool                   `json:"published" bson:"published"`

	// Audit Fields
	CreatedBy       string                 `json:"createdby" bson:"createdby"`
	CreatedOn       time.Time              `json:"createdon" bson:"createdon"`
	ModifiedBy      string                 `json:"modifiedby" bson:"modifiedby"`
	ModifiedOn      time.Time              `json:"modifiedon" bson:"modifiedon"`
	RowVersionStamp int                    `json:"rowversionstamp" bson:"rowversionstamp"`
}

// PlantModelListResponse represents API response for listing plant models
type PlantModelListResponse struct {
	Data  []PlantModel `json:"data"`
	Count int          `json:"count"`
}

// PlantModelAssetListResponse represents API response for listing plant assets
type PlantModelAssetListResponse struct {
	Data  []PlantModelAsset `json:"data"`
	Count int               `json:"count"`
}
