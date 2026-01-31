package configstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services/cluster"
)

// ConfigStoreService manages centralized configuration storage
type ConfigStoreService struct {
	configCollection  *mongo.Collection
	historyCollection *mongo.Collection
	overrideCollection *mongo.Collection
	config            *config.ConfigStoreConfig
	log               logger.Log
	mutex             sync.RWMutex
	activeConfig      *models.ConfigurationStore
	onReloadCallbacks []func(*models.ConfigReloadEvent)
	callbackMutex     sync.RWMutex
}

// Global instance
var (
	globalConfigStoreService *ConfigStoreService
	globalConfigStoreMutex   sync.RWMutex
)

// GetGlobalConfigStoreService returns the global config store service
func GetGlobalConfigStoreService() *ConfigStoreService {
	globalConfigStoreMutex.RLock()
	defer globalConfigStoreMutex.RUnlock()
	return globalConfigStoreService
}

// SetGlobalConfigStoreService sets the global config store service
func SetGlobalConfigStoreService(service *ConfigStoreService) {
	globalConfigStoreMutex.Lock()
	defer globalConfigStoreMutex.Unlock()
	globalConfigStoreService = service
}

// NewConfigStoreService creates a new configuration store service
func NewConfigStoreService(cfg *config.ConfigStoreConfig) (*ConfigStoreService, error) {
	if cfg == nil {
		cfg = config.DefaultConfigStoreConfig()
	}

	// Get MongoDB client from global connection
	if com.IACDocDBConn.MongoDBClient == nil {
		return nil, fmt.Errorf("MongoDB client not available")
	}

	db := com.IACDocDBConn.MongoDBClient.Database(com.IACDocDBConn.DatabaseName)
	configCollection := db.Collection(cfg.Collection)
	historyCollection := db.Collection(cfg.Collection + "_history")
	overrideCollection := db.Collection(cfg.Collection + "_overrides")

	service := &ConfigStoreService{
		configCollection:   configCollection,
		historyCollection:  historyCollection,
		overrideCollection: overrideCollection,
		config:             cfg,
		log:                logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "ConfigStoreService"},
		onReloadCallbacks:  make([]func(*models.ConfigReloadEvent), 0),
	}

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Config collection indexes
	configCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}, {Key: "version", Value: 1}, {Key: "environment", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "name", Value: 1}, {Key: "environment", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
	})

	// History collection indexes
	historyCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "config_id", Value: 1}, {Key: "performed_at", Value: -1}}},
		{Keys: bson.D{{Key: "config_name", Value: 1}, {Key: "performed_at", Value: -1}}},
	})

	// Override collection indexes
	overrideCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "config_id", Value: 1}, {Key: "instance_pattern", Value: 1}}, Options: options.Index().SetUnique(true)},
	})

	service.log.Info("Configuration store service initialized")
	return service, nil
}

// CreateConfiguration creates a new configuration version
func (s *ConfigStoreService) CreateConfiguration(ctx context.Context, cfg *models.ConfigurationStore) (*models.ConfigurationStore, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Generate checksum
	cfg.Checksum = s.generateChecksum(cfg.Data)
	cfg.CreatedAt = time.Now()
	cfg.UpdatedAt = cfg.CreatedAt

	// Find previous active version
	var prevConfig models.ConfigurationStore
	err := s.configCollection.FindOne(ctx, bson.M{
		"name":        cfg.Name,
		"environment": cfg.Environment,
		"status":      models.ConfigStatusActive,
	}).Decode(&prevConfig)
	if err == nil {
		cfg.PreviousVersion = prevConfig.Version
	}

	result, err := s.configCollection.InsertOne(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration: %w", err)
	}

	cfg.ID = result.InsertedID.(primitive.ObjectID).Hex()

	// Record history
	s.recordHistory(ctx, cfg.ID, cfg.Name, cfg.Version, cfg.Environment, "created", cfg.CreatedBy, nil, cfg.Data, "")

	s.log.Info(fmt.Sprintf("Created configuration: %s v%s (%s)", cfg.Name, cfg.Version, cfg.Environment))
	return cfg, nil
}

// GetConfiguration retrieves a configuration by ID
func (s *ConfigStoreService) GetConfiguration(ctx context.Context, id string) (*models.ConfigurationStore, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration ID")
	}

	var cfg models.ConfigurationStore
	err = s.configCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&cfg)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	cfg.ID = id
	return &cfg, nil
}

// GetActiveConfiguration retrieves the active configuration for a name and environment
func (s *ConfigStoreService) GetActiveConfiguration(ctx context.Context, name, environment string) (*models.ConfigurationStore, error) {
	var cfg models.ConfigurationStore
	err := s.configCollection.FindOne(ctx, bson.M{
		"name":        name,
		"environment": environment,
		"status":      models.ConfigStatusActive,
	}).Decode(&cfg)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active configuration: %w", err)
	}

	return &cfg, nil
}

// ListConfigurations lists configurations with optional filters
func (s *ConfigStoreService) ListConfigurations(ctx context.Context, name, environment string, status models.ConfigurationStatus) ([]*models.ConfigurationStore, error) {
	filter := bson.M{}
	if name != "" {
		filter["name"] = name
	}
	if environment != "" {
		filter["environment"] = environment
	}
	if status != "" {
		filter["status"] = status
	}

	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	cursor, err := s.configCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list configurations: %w", err)
	}
	defer cursor.Close(ctx)

	var configs []*models.ConfigurationStore
	for cursor.Next(ctx) {
		var cfg models.ConfigurationStore
		if err := cursor.Decode(&cfg); err != nil {
			continue
		}
		configs = append(configs, &cfg)
	}

	return configs, nil
}

// UpdateConfiguration updates a configuration
func (s *ConfigStoreService) UpdateConfiguration(ctx context.Context, id string, data map[string]interface{}, updatedBy, reason string) (*models.ConfigurationStore, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cfg, err := s.GetConfiguration(ctx, id)
	if err != nil || cfg == nil {
		return nil, fmt.Errorf("configuration not found")
	}

	if cfg.Status == models.ConfigStatusActive {
		return nil, fmt.Errorf("cannot modify active configuration, create a new version instead")
	}

	oldData := cfg.Data
	cfg.Data = data
	cfg.Checksum = s.generateChecksum(data)
	cfg.UpdatedBy = updatedBy
	cfg.UpdatedAt = time.Now()

	objID, _ := primitive.ObjectIDFromHex(id)
	_, err = s.configCollection.ReplaceOne(ctx, bson.M{"_id": objID}, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to update configuration: %w", err)
	}

	// Record history
	s.recordHistory(ctx, id, cfg.Name, cfg.Version, cfg.Environment, "updated", updatedBy, oldData, data, reason)

	return cfg, nil
}

// ActivateConfiguration activates a configuration version
func (s *ConfigStoreService) ActivateConfiguration(ctx context.Context, id, activatedBy string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cfg, err := s.GetConfiguration(ctx, id)
	if err != nil || cfg == nil {
		return fmt.Errorf("configuration not found")
	}

	// Deactivate current active configuration
	_, err = s.configCollection.UpdateMany(ctx, bson.M{
		"name":        cfg.Name,
		"environment": cfg.Environment,
		"status":      models.ConfigStatusActive,
	}, bson.M{
		"$set": bson.M{
			"status":     models.ConfigStatusArchived,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to deactivate previous configuration: %w", err)
	}

	// Activate new configuration
	cfg.Activate(activatedBy)
	objID, _ := primitive.ObjectIDFromHex(id)
	_, err = s.configCollection.ReplaceOne(ctx, bson.M{"_id": objID}, cfg)
	if err != nil {
		return fmt.Errorf("failed to activate configuration: %w", err)
	}

	// Record history
	s.recordHistory(ctx, id, cfg.Name, cfg.Version, cfg.Environment, "activated", activatedBy, nil, nil, "")

	// Cache active config
	s.activeConfig = cfg

	// Broadcast reload event
	s.broadcastReload(cfg, "activated")

	s.log.Info(fmt.Sprintf("Activated configuration: %s v%s (%s)", cfg.Name, cfg.Version, cfg.Environment))
	return nil
}

// RollbackConfiguration rolls back to a previous version
func (s *ConfigStoreService) RollbackConfiguration(ctx context.Context, name, environment, targetVersion, performedBy, reason string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Find target version
	var targetCfg models.ConfigurationStore
	err := s.configCollection.FindOne(ctx, bson.M{
		"name":        name,
		"environment": environment,
		"version":     targetVersion,
	}).Decode(&targetCfg)
	if err != nil {
		return fmt.Errorf("target version not found")
	}

	// Deactivate current active
	_, err = s.configCollection.UpdateMany(ctx, bson.M{
		"name":        name,
		"environment": environment,
		"status":      models.ConfigStatusActive,
	}, bson.M{
		"$set": bson.M{
			"status":     models.ConfigStatusArchived,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to deactivate current configuration: %w", err)
	}

	// Activate target version
	targetCfg.Activate(performedBy)
	objID, _ := primitive.ObjectIDFromHex(targetCfg.ID)
	_, err = s.configCollection.ReplaceOne(ctx, bson.M{"_id": objID}, &targetCfg)
	if err != nil {
		return fmt.Errorf("failed to activate target version: %w", err)
	}

	// Record history
	s.recordHistory(ctx, targetCfg.ID, name, targetVersion, environment, "rolled_back", performedBy, nil, nil, reason)

	// Broadcast reload
	s.broadcastReload(&targetCfg, "rolled_back")

	s.log.Info(fmt.Sprintf("Rolled back configuration: %s to v%s (%s)", name, targetVersion, environment))
	return nil
}

// GetHistory retrieves configuration history
func (s *ConfigStoreService) GetHistory(ctx context.Context, configID string, limit int) ([]*models.ConfigurationHistory, error) {
	filter := bson.M{}
	if configID != "" {
		filter["config_id"] = configID
	}

	opts := options.Find().SetSort(bson.D{{Key: "performed_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := s.historyCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	defer cursor.Close(ctx)

	var history []*models.ConfigurationHistory
	for cursor.Next(ctx) {
		var h models.ConfigurationHistory
		if err := cursor.Decode(&h); err != nil {
			continue
		}
		history = append(history, &h)
	}

	return history, nil
}

// Helper functions

func (s *ConfigStoreService) generateChecksum(data map[string]interface{}) string {
	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

func (s *ConfigStoreService) recordHistory(ctx context.Context, configID, configName, version, environment, action, performedBy string, oldData, newData map[string]interface{}, reason string) {
	history := models.NewConfigurationHistory(configID, configName, version, environment, action, performedBy)
	history.Reason = reason
	history.PreviousData = oldData
	history.NewData = newData

	s.historyCollection.InsertOne(ctx, history)
}

func (s *ConfigStoreService) broadcastReload(cfg *models.ConfigurationStore, action string) {
	event := &models.ConfigReloadEvent{
		ConfigID:       cfg.ID,
		ConfigName:     cfg.Name,
		Version:        cfg.Version,
		Environment:    cfg.Environment,
		Action:         action,
		SourceInstance: config.GetInstanceName(),
		Timestamp:      time.Now(),
	}

	// Notify local callbacks
	s.callbackMutex.RLock()
	for _, cb := range s.onReloadCallbacks {
		go cb(event)
	}
	s.callbackMutex.RUnlock()

	// Broadcast via SignalR
	discoveryService := cluster.GetGlobalDiscoveryService()
	if discoveryService != nil {
		discoveryService.BroadcastConfigReload(cfg.Version)
	}
}

// OnReload registers a callback for configuration reload events
func (s *ConfigStoreService) OnReload(callback func(*models.ConfigReloadEvent)) {
	s.callbackMutex.Lock()
	defer s.callbackMutex.Unlock()
	s.onReloadCallbacks = append(s.onReloadCallbacks, callback)
}

// GetActiveConfig returns the cached active configuration
func (s *ConfigStoreService) GetActiveConfig() *models.ConfigurationStore {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.activeConfig
}
