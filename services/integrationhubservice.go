package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"go.mongodb.org/mongo-driver/bson"
)

const CollectionIntegrationHub = "Integration_Hub"

// IntegrationHubService provides simplified single-document integration hub operations
type IntegrationHubService struct {
	docDB documents.DocumentDB
	iLog  logger.Log
}

// NewIntegrationHubService creates new integration hub service with document database
func NewIntegrationHubService(docDB documents.DocumentDB) *IntegrationHubService {
	return &IntegrationHubService{
		docDB: docDB,
		iLog:  logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "IntegrationHubService"},
	}
}

// ListHubs retrieves all integration hub configurations
func (s *IntegrationHubService) ListHubs(ctx context.Context) ([]models.IntegrationHub, error) {
	startTime := time.Now()
	defer func() {
		s.iLog.Debug(fmt.Sprintf("ListHubs completed in %v", time.Since(startTime)))
	}()

	filter := bson.M{"active": true}

	results, err := s.docDB.FindMany(ctx, CollectionIntegrationHub, filter, nil)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to list hubs: %v", err))
		return nil, fmt.Errorf("failed to list integration hubs: %w", err)
	}

	hubs := make([]models.IntegrationHub, 0, len(results))
	for _, result := range results {
		var hub models.IntegrationHub
		bsonBytes, err := bson.Marshal(result)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to marshal result: %v", err))
			continue
		}

		err = bson.Unmarshal(bsonBytes, &hub)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to unmarshal hub: %v", err))
			continue
		}

		// Map _id to ID
		if objID, ok := result["_id"].(string); ok {
			hub.ID = objID
		}

		hubs = append(hubs, hub)
	}

	s.iLog.Info(fmt.Sprintf("Retrieved %d integration hubs", len(hubs)))
	return hubs, nil
}

// GetHub retrieves a single integration hub configuration by ID
func (s *IntegrationHubService) GetHub(ctx context.Context, id string) (*models.IntegrationHub, error) {
	startTime := time.Now()
	defer func() {
		s.iLog.Debug(fmt.Sprintf("GetHub completed in %v", time.Since(startTime)))
	}()

	filter := bson.M{"_id": id, "active": true}

	result, err := s.docDB.FindOne(ctx, CollectionIntegrationHub, filter)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get hub: %v", err))
		return nil, fmt.Errorf("failed to get integration hub: %w", err)
	}

	var hub models.IntegrationHub
	bsonBytes, err := bson.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	err = bson.Unmarshal(bsonBytes, &hub)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal hub: %w", err)
	}

	// Map _id to ID
	if objID, ok := result["_id"].(string); ok {
		hub.ID = objID
	}

	s.iLog.Info(fmt.Sprintf("Retrieved integration hub: %s", id))
	return &hub, nil
}

// SaveHub creates or updates an integration hub configuration (whole document)
func (s *IntegrationHubService) SaveHub(ctx context.Context, hub *models.IntegrationHub, user string) error {
	startTime := time.Now()
	defer func() {
		s.iLog.Debug(fmt.Sprintf("SaveHub completed in %v", time.Since(startTime)))
	}()

	// Check if hub exists
	isUpdate := false
	var existingHub *models.IntegrationHub
	if hub.ID != "" {
		existing, err := s.GetHub(ctx, hub.ID)
		if err == nil && existing != nil {
			isUpdate = true
			existingHub = existing
		}
	}

	// Set metadata
	if !isUpdate {
		// Creating new hub
		if hub.ID == "" {
			hub.ID = uuid.New().String()
		}
		hub.CreatedBy = user
		hub.CreatedOn = time.Now()
		hub.RowVersionStamp = 1
		hub.Active = true

		// Set default version if not specified
		if hub.Version == 0 {
			nextVersion, _ := s.GetNextVersion(ctx, hub.Name)
			hub.Version = nextVersion
		}

		// Set default status if not specified
		if hub.Status == "" {
			hub.Status = models.IntegrationHubStatusDev
		}

		// Handle IsDefault for new hubs
		if hub.IsDefault && hub.Name != "" {
			// Clear existing defaults
			clearFilter := bson.M{"name": hub.Name, "is_default": true, "active": true}
			clearUpdate := bson.M{
				"$set": bson.M{
					"is_default": false,
					"modifiedby": user,
					"modifiedon": time.Now(),
				},
			}
			_, _ = s.docDB.UpdateMany(ctx, CollectionIntegrationHub, clearFilter, clearUpdate)
		}
	} else {
		// Updating existing hub
		hub.RowVersionStamp = existingHub.RowVersionStamp + 1

		// Handle IsDefault change
		if hub.IsDefault && !existingHub.IsDefault && hub.Name != "" {
			// Clear existing defaults
			clearFilter := bson.M{"name": hub.Name, "is_default": true, "active": true, "_id": bson.M{"$ne": hub.ID}}
			clearUpdate := bson.M{
				"$set": bson.M{
					"is_default": false,
					"modifiedby": user,
					"modifiedon": time.Now(),
				},
			}
			_, _ = s.docDB.UpdateMany(ctx, CollectionIntegrationHub, clearFilter, clearUpdate)
		}
	}

	hub.ModifiedBy = user
	hub.ModifiedOn = time.Now()

	// Initialize nested structures if nil
	if hub.Inbound == nil {
		hub.Inbound = &models.HubDirection{
			Enabled:        false,
			ProtocolGroups: []models.HubSimpleProtocolGroup{},
		}
	}
	if hub.Outbound == nil {
		hub.Outbound = &models.HubDirection{
			Enabled:        false,
			ProtocolGroups: []models.HubSimpleProtocolGroup{},
		}
	}
	if hub.Routes == nil {
		hub.Routes = []models.HubSimpleRoute{}
	}
	if hub.Jobs == nil {
		hub.Jobs = []models.HubSimpleJob{}
	}
	if hub.Metadata == nil {
		hub.Metadata = make(map[string]interface{})
	}

	// Convert to BSON document
	doc, err := bson.Marshal(hub)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to marshal hub: %v", err))
		return fmt.Errorf("failed to marshal hub: %w", err)
	}

	var bsonDoc bson.M
	err = bson.Unmarshal(doc, &bsonDoc)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to unmarshal to bson.M: %v", err))
		return fmt.Errorf("failed to unmarshal to bson.M: %w", err)
	}

	if !isUpdate {
		// Insert new document
		_, err = s.docDB.InsertOne(ctx, CollectionIntegrationHub, bsonDoc)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to create hub: %v", err))
			return fmt.Errorf("failed to create hub: %w", err)
		}
		s.iLog.Info(fmt.Sprintf("Created integration hub: %s", hub.ID))
	} else {
		// Replace entire document
		filter := bson.M{"_id": hub.ID}
		update := bson.M{"$set": bsonDoc}

		err = s.docDB.UpdateOne(ctx, CollectionIntegrationHub, filter, update)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to update hub: %v", err))
			return fmt.Errorf("failed to update hub: %w", err)
		}
		s.iLog.Info(fmt.Sprintf("Updated integration hub: %s", hub.ID))
	}

	return nil
}

// DeleteHub soft-deletes an integration hub configuration
func (s *IntegrationHubService) DeleteHub(ctx context.Context, id string) error {
	startTime := time.Now()
	defer func() {
		s.iLog.Debug(fmt.Sprintf("DeleteHub completed in %v", time.Since(startTime)))
	}()

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"active":     false,
			"modifiedon": time.Now(),
		},
	}

	err := s.docDB.UpdateOne(ctx, CollectionIntegrationHub, filter, update)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to delete hub: %v", err))
		return fmt.Errorf("failed to delete hub: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Deleted integration hub: %s", id))
	return nil
}

// ListHubsByName retrieves all configurations for a specific name
func (s *IntegrationHubService) ListHubsByName(ctx context.Context, name string) ([]models.IntegrationHub, error) {
	startTime := time.Now()
	defer func() {
		s.iLog.Debug(fmt.Sprintf("ListHubsByInstanceName completed in %v", time.Since(startTime)))
	}()

	filter := bson.M{"name": name, "active": true}

	results, err := s.docDB.FindMany(ctx, CollectionIntegrationHub, filter, nil)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to list hubs by instance name: %v", err))
		return nil, fmt.Errorf("failed to list hubs by instance name: %w", err)
	}

	hubs := make([]models.IntegrationHub, 0, len(results))
	for _, result := range results {
		var hub models.IntegrationHub
		bsonBytes, err := bson.Marshal(result)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to marshal result: %v", err))
			continue
		}

		err = bson.Unmarshal(bsonBytes, &hub)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to unmarshal hub: %v", err))
			continue
		}

		if objID, ok := result["_id"].(string); ok {
			hub.ID = objID
		}

		hubs = append(hubs, hub)
	}

	s.iLog.Info(fmt.Sprintf("Retrieved %d integration hubs for name: %s", len(hubs), name))
	return hubs, nil
}

// GetNextVersion gets the next version number for a hub name
func (s *IntegrationHubService) GetNextVersion(ctx context.Context, name string) (int, error) {
	startTime := time.Now()
	defer func() {
		s.iLog.Debug(fmt.Sprintf("GetNextVersion completed in %v", time.Since(startTime)))
	}()

	hubs, err := s.ListHubsByName(ctx, name)
	if err != nil {
		return 1, nil // First version if no existing hubs
	}

	maxVersion := 0
	for _, hub := range hubs {
		if hub.Version > maxVersion {
			maxVersion = hub.Version
		}
	}

	return maxVersion + 1, nil
}

// SetAsDefault sets a hub as the default for its name, clearing other defaults
func (s *IntegrationHubService) SetAsDefault(ctx context.Context, id string, user string) error {
	startTime := time.Now()
	defer func() {
		s.iLog.Debug(fmt.Sprintf("SetAsDefault completed in %v", time.Since(startTime)))
	}()

	// Get the hub to find its name
	hub, err := s.GetHub(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get hub: %w", err)
	}

	// Clear existing defaults for this name
	clearFilter := bson.M{"name": hub.Name, "is_default": true, "active": true}
	clearUpdate := bson.M{
		"$set": bson.M{
			"is_default": false,
			"modifiedby": user,
			"modifiedon": time.Now(),
		},
	}

	_, err = s.docDB.UpdateMany(ctx, CollectionIntegrationHub, clearFilter, clearUpdate)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to clear existing defaults: %v", err))
		return fmt.Errorf("failed to clear existing defaults: %w", err)
	}

	// Set this hub as default
	setFilter := bson.M{"_id": id}
	setUpdate := bson.M{
		"$set": bson.M{
			"is_default": true,
			"modifiedby": user,
			"modifiedon": time.Now(),
		},
	}

	err = s.docDB.UpdateOne(ctx, CollectionIntegrationHub, setFilter, setUpdate)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to set as default: %v", err))
		return fmt.Errorf("failed to set as default: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Set integration hub %s as default for name: %s", id, hub.Name))
	return nil
}

// CreateRevision creates a new version based on an existing hub configuration
func (s *IntegrationHubService) CreateRevision(ctx context.Context, sourceId string, user string) (*models.IntegrationHub, error) {
	startTime := time.Now()
	defer func() {
		s.iLog.Debug(fmt.Sprintf("CreateRevision completed in %v", time.Since(startTime)))
	}()

	// Get the source hub
	sourceHub, err := s.GetHub(ctx, sourceId)
	if err != nil {
		return nil, fmt.Errorf("failed to get source hub: %w", err)
	}

	// Get next version
	nextVersion, err := s.GetNextVersion(ctx, sourceHub.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get next version: %w", err)
	}

	// Create new hub based on source
	newHub := *sourceHub
	newHub.ID = uuid.New().String()
	newHub.Version = nextVersion
	newHub.IsDefault = false
	newHub.Status = models.IntegrationHubStatusDev
	newHub.CreatedBy = user
	newHub.CreatedOn = time.Now()
	newHub.ModifiedBy = user
	newHub.ModifiedOn = time.Now()
	newHub.RowVersionStamp = 1

	// Convert to BSON document
	doc, err := bson.Marshal(newHub)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to marshal hub: %v", err))
		return nil, fmt.Errorf("failed to marshal hub: %w", err)
	}

	var bsonDoc bson.M
	err = bson.Unmarshal(doc, &bsonDoc)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to unmarshal to bson.M: %v", err))
		return nil, fmt.Errorf("failed to unmarshal to bson.M: %w", err)
	}

	// Insert new document
	_, err = s.docDB.InsertOne(ctx, CollectionIntegrationHub, bsonDoc)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to create revision: %v", err))
		return nil, fmt.Errorf("failed to create revision: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Created revision %d for name: %s", nextVersion, sourceHub.Name))
	return &newHub, nil
}

// UpdateStatus updates the status of a hub configuration
func (s *IntegrationHubService) UpdateStatus(ctx context.Context, id string, status models.IntegrationHubStatus, user string) error {
	startTime := time.Now()
	defer func() {
		s.iLog.Debug(fmt.Sprintf("UpdateStatus completed in %v", time.Since(startTime)))
	}()

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"modifiedby": user,
			"modifiedon": time.Now(),
		},
	}

	err := s.docDB.UpdateOne(ctx, CollectionIntegrationHub, filter, update)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to update status: %v", err))
		return fmt.Errorf("failed to update status: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Updated status of hub %s to: %s", id, status))
	return nil
}

// GetDefaultHub retrieves the default hub configuration for a name
func (s *IntegrationHubService) GetDefaultHub(ctx context.Context, name string) (*models.IntegrationHub, error) {
	startTime := time.Now()
	defer func() {
		s.iLog.Debug(fmt.Sprintf("GetDefaultHub completed in %v", time.Since(startTime)))
	}()

	// Try to find the one marked as default
	filter := bson.M{"name": name, "is_default": true, "active": true}

	result, err := s.docDB.FindOne(ctx, CollectionIntegrationHub, filter)
	if err != nil {
		s.iLog.Warn(fmt.Sprintf("No default hub found for name %s explicitly, checking for any active version", name))

		// Try to find any active documents for this name
		filter = bson.M{"name": name, "active": true}
		results, err2 := s.docDB.FindMany(ctx, CollectionIntegrationHub, filter, nil)

		if err2 != nil || len(results) == 0 {
			s.iLog.Error(fmt.Sprintf("Failed to get default hub for %s: %v", name, err))
			return nil, fmt.Errorf("failed to get default hub: %w", err)
		}

		// Pick the one with the highest version
		var latestResult map[string]interface{}
		maxVersion := -1

		for _, item := range results {
			version := 0
			// Handle different possible types for version field
			if v, ok := item["version"].(int32); ok {
				version = int(v)
			} else if v, ok := item["version"].(int64); ok {
				version = int(v)
			} else if v, ok := item["version"].(int); ok {
				version = v
			} else if v, ok := item["version"].(float64); ok {
				version = int(v)
			}

			if version > maxVersion {
				maxVersion = version
				latestResult = item
			}
		}

		if latestResult == nil {
			return nil, fmt.Errorf("failed to get default hub: %w", err)
		}
		result = latestResult
		s.iLog.Info(fmt.Sprintf("Using latest version %d for name %s as default", maxVersion, name))
	}

	var hub models.IntegrationHub
	bsonBytes, err := bson.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	err = bson.Unmarshal(bsonBytes, &hub)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal hub: %w", err)
	}

	if objID, ok := result["_id"].(string); ok {
		hub.ID = objID
	}

	s.iLog.Info(fmt.Sprintf("Retrieved default hub for name: %s", name))
	return &hub, nil
}

// GetAvailableStatuses returns all available status options
func (s *IntegrationHubService) GetAvailableStatuses() []models.IntegrationHubStatus {
	return []models.IntegrationHubStatus{
		models.IntegrationHubStatusDev,
		models.IntegrationHubStatusTest,
		models.IntegrationHubStatusProduction,
		models.IntegrationHubStatusDeprecated,
	}
}

// GetAllEnabledHubs retrieves all enabled hub configurations (for hub executor)
func (s *IntegrationHubService) GetAllEnabledHubs(ctx context.Context) ([]*models.IntegrationHub, error) {
	startTime := time.Now()
	defer func() {
		s.iLog.Debug(fmt.Sprintf("GetAllEnabledHubs completed in %v", time.Since(startTime)))
	}()

	filter := bson.M{"enabled": true, "active": true}

	results, err := s.docDB.FindMany(ctx, CollectionIntegrationHub, filter, nil)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get enabled hubs: %v", err))
		return nil, fmt.Errorf("failed to get enabled hubs: %w", err)
	}

	hubs := make([]*models.IntegrationHub, 0, len(results))
	for _, result := range results {
		var hub models.IntegrationHub
		bsonBytes, err := bson.Marshal(result)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to marshal result: %v", err))
			continue
		}

		err = bson.Unmarshal(bsonBytes, &hub)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to unmarshal hub: %v", err))
			continue
		}

		if objID, ok := result["_id"].(string); ok {
			hub.ID = objID
		}

		hubs = append(hubs, &hub)
	}

	s.iLog.Info(fmt.Sprintf("Retrieved %d enabled integration hubs", len(hubs)))
	return hubs, nil
}

// GetDefaultEnabledHubs retrieves all default hub configurations that are enabled
func (s *IntegrationHubService) GetDefaultEnabledHubs(ctx context.Context) ([]*models.IntegrationHub, error) {
	startTime := time.Now()
	defer func() {
		s.iLog.Debug(fmt.Sprintf("GetDefaultEnabledHubs completed in %v", time.Since(startTime)))
	}()

	filter := bson.M{"enabled": true, "is_default": true, "active": true}

	results, err := s.docDB.FindMany(ctx, CollectionIntegrationHub, filter, nil)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to get default enabled hubs: %v", err))
		return nil, fmt.Errorf("failed to get default enabled hubs: %w", err)
	}

	hubs := make([]*models.IntegrationHub, 0, len(results))
	for _, result := range results {
		var hub models.IntegrationHub
		bsonBytes, err := bson.Marshal(result)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to marshal result: %v", err))
			continue
		}

		err = bson.Unmarshal(bsonBytes, &hub)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to unmarshal hub: %v", err))
			continue
		}

		if objID, ok := result["_id"].(string); ok {
			hub.ID = objID
		}

		hubs = append(hubs, &hub)
	}

	s.iLog.Info(fmt.Sprintf("Retrieved %d default enabled integration hubs", len(hubs)))
	return hubs, nil
}

// ============================================================================
// Package-level functions for hub executor integration
// ============================================================================

// getDefaultIntegrationHubService returns a default IntegrationHubService using the global document DB
func getDefaultIntegrationHubService() *IntegrationHubService {
	factory := documents.GetDocFactory()
	if factory == nil {
		return nil
	}

	db, err := factory.GetDB("default")
	if err != nil {
		return nil
	}

	return NewIntegrationHubService(db)
}

// GetDefaultHub retrieves the default hub configuration for a name (package-level function)
func GetDefaultHub(name string) (*models.IntegrationHub, error) {
	service := getDefaultIntegrationHubService()
	if service == nil {
		return nil, fmt.Errorf("integration hub service not available")
	}

	return service.GetDefaultHub(context.Background(), name)
}

// GetHub retrieves a hub configuration by ID (package-level function)
func GetHub(hubID string) (*models.IntegrationHub, error) {
	service := getDefaultIntegrationHubService()
	if service == nil {
		return nil, fmt.Errorf("integration hub service not available")
	}

	return service.GetHub(context.Background(), hubID)
}

// GetDefaultEnabledHubs retrieves all default enabled hub configurations (package-level function)
func GetDefaultEnabledHubs() ([]*models.IntegrationHub, error) {
	service := getDefaultIntegrationHubService()
	if service == nil {
		return nil, fmt.Errorf("integration hub service not available")
	}

	return service.GetDefaultEnabledHubs(context.Background())
}
