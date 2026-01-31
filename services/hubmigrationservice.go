package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// HubMigrationService handles migration from flat to hierarchical structure
type HubMigrationService struct {
	db          *sql.DB
	docDB       documents.DocumentDB
	hubService  *HubService
	iLog        logger.Log
}

// NewHubMigrationService creates new migration service
func NewHubMigrationService(db *sql.DB, docDB documents.DocumentDB) *HubMigrationService {
	return &HubMigrationService{
		db:         db,
		docDB:      docDB,
		hubService: NewHubService(docDB),
		iLog:       logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "HubMigrationService"},
	}
}

// OldIntegrationHub represents the old flat structure from MongoDB
type OldIntegrationHub struct {
	ID          string         `json:"_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Enabled     bool           `json:"enabled"`
	Inbound     InboundConfig  `json:"inbound"`
	Outbound    OutboundConfig `json:"outbound"`
	Routes      []MessageRoute `json:"routes"`
	Jobs        []JobConfig    `json:"jobs"`
	CreatedBy   string         `json:"createdBy"`
}

// InboundConfig from old structure
type InboundConfig struct {
	Enabled        bool              `json:"enabled"`
	Protocol       string            `json:"protocol"`
	MessageType    string            `json:"messageType"`
	Endpoint       string            `json:"endpoint"`
	Port           int               `json:"port"`
	Path           string            `json:"path"`
	Method         string            `json:"method"`
	Timeout        int               `json:"timeout"`
	RetryAttempts  int               `json:"retryAttempts"`
	RetryInterval  int               `json:"retryInterval"`
	Authentication AuthConfig        `json:"authentication"`
	QueueConfig    map[string]interface{} `json:"queueConfig"`
	FileConfig     map[string]interface{} `json:"fileConfig"`
}

// OutboundConfig from old structure
type OutboundConfig struct {
	Enabled       bool              `json:"enabled"`
	Destination   string            `json:"destination"`
	Protocol      string            `json:"protocol"`
	MessageType   string            `json:"messageType"`
	Method        string            `json:"method"`
	Timeout       int               `json:"timeout"`
	RetryAttempts int               `json:"retryAttempts"`
	RetryInterval int               `json:"retryInterval"`
	Authorization AuthConfig        `json:"authorization"`
	QueueConfig   map[string]interface{} `json:"queueConfig"`
	FileConfig    map[string]interface{} `json:"fileConfig"`
}

// AuthConfig for authentication/authorization
type AuthConfig struct {
	Type        string                 `json:"type"`
	Token       string                 `json:"token"`
	APIKey      string                 `json:"apiKey"`
	Username    string                 `json:"username"`
	Password    string                 `json:"password"`
	OAuth2      map[string]interface{} `json:"oauth2Config"`
	Credentials map[string]string      `json:"credentials"`
}

// MessageRoute from old structure
type MessageRoute struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	Enabled             bool                   `json:"enabled"`
	Source              string                 `json:"source"`
	SourceFilter        string                 `json:"sourceFilter"`
	Destination         string                 `json:"destination"`
	TransformationType  string                 `json:"transformationType"`
	Transformation      string                 `json:"transformation"`
	Mapping             []map[string]interface{} `json:"mapping"`
	Conditions          []map[string]interface{} `json:"conditions"`
	Priority            int                    `json:"priority"`
	Async               bool                   `json:"async"`
	ErrorHandling       map[string]interface{} `json:"errorHandling"`
}

// JobConfig from old structure
type JobConfig struct {
	Enabled         bool                   `json:"enabled"`
	JobType         string                 `json:"jobType"`
	Schedule        map[string]interface{} `json:"schedule"`
	Trigger         map[string]interface{} `json:"trigger"`
	ExecutionConfig map[string]interface{} `json:"executionConfig"`
	Parameters      map[string]interface{} `json:"parameters"`
}

// MigrateFromFlat migrates old Integration_Hub to new hierarchical structure
func (hms *HubMigrationService) MigrateFromFlat(ctx context.Context, oldHub OldIntegrationHub, targetInstanceID string) error {
	startTime := time.Now()
	defer func() {
		hms.iLog.Info(fmt.Sprintf("Migration completed in %v", time.Since(startTime)))
	}()

	hms.iLog.Info(fmt.Sprintf("Starting migration for hub: %s", oldHub.Name))

	// Start transaction for atomic migration
	tx, err := hms.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	migrationLog := &models.HubMigrationLog{
		ID:            uuid.New().String(),
		OldHubID:      oldHub.ID,
		NewInstanceID: targetInstanceID,
		MigrationStatus: "pending",
		MigrationDetails: make(models.HubMetadata),
		CreatedBy:     oldHub.CreatedBy,
		CreatedOn:     time.Now(),
	}

	// 1. Create or get hub instance
	instance, err := hms.createHubInstance(ctx, oldHub, targetInstanceID)
	if err != nil {
		hms.logMigrationError(ctx, migrationLog, err)
		return fmt.Errorf("failed to create hub instance: %w", err)
	}
	migrationLog.NewInstanceID = instance.ID
	hms.iLog.Info(fmt.Sprintf("Created hub instance: %s", instance.ID))

	// 2. Migrate inbound configuration
	if oldHub.Inbound.Enabled {
		inboundGroupID, inboundEndpointID, err := hms.migrateInbound(ctx, instance.ID, oldHub.Inbound, oldHub.CreatedBy)
		if err != nil {
			hms.logMigrationError(ctx, migrationLog, err)
			return fmt.Errorf("failed to migrate inbound: %w", err)
		}
		migrationLog.MigrationDetails["inbound_group_id"] = inboundGroupID
		migrationLog.MigrationDetails["inbound_endpoint_id"] = inboundEndpointID
		hms.iLog.Info(fmt.Sprintf("Migrated inbound configuration: group=%s, endpoint=%s", inboundGroupID, inboundEndpointID))
	}

	// 3. Migrate outbound configuration
	if oldHub.Outbound.Enabled {
		outboundGroupID, outboundEndpointID, err := hms.migrateOutbound(ctx, instance.ID, oldHub.Outbound, oldHub.CreatedBy)
		if err != nil {
			hms.logMigrationError(ctx, migrationLog, err)
			return fmt.Errorf("failed to migrate outbound: %w", err)
		}
		migrationLog.MigrationDetails["outbound_group_id"] = outboundGroupID
		migrationLog.MigrationDetails["outbound_endpoint_id"] = outboundEndpointID
		hms.iLog.Info(fmt.Sprintf("Migrated outbound configuration: group=%s, endpoint=%s", outboundGroupID, outboundEndpointID))
	}

	// 4. Migrate routes (if needed - routes reference endpoints)
	// Note: Routes will need endpoint IDs from inbound/outbound
	// This is a placeholder for route migration logic

	// 5. Commit transaction
	if err := tx.Commit(); err != nil {
		hms.logMigrationError(ctx, migrationLog, err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 6. Log successful migration
	migrationLog.MigrationStatus = "completed"
	migrationLog.MigratedAt = time.Now()
	hms.saveMigrationLog(ctx, migrationLog)

	hms.iLog.Info(fmt.Sprintf("Successfully migrated hub %s to instance %s", oldHub.ID, instance.ID))
	return nil
}

// createHubInstance creates a hub instance for the old hub
func (hms *HubMigrationService) createHubInstance(ctx context.Context, oldHub OldIntegrationHub, targetInstanceID string) (*models.HubInstance, error) {
	// Use provided instance ID or current system instance
	if targetInstanceID == "" {
		targetInstanceID = com.InstanceID
	}

	instance := &models.HubInstance{
		ID:          uuid.New().String(),
		InstanceID:  targetInstanceID,
		Name:        oldHub.Name,
		Description: oldHub.Description,
		Enabled:     oldHub.Enabled,
		Metadata: models.HubMetadata{
			"migrated_from":    oldHub.ID,
			"migration_date":   time.Now().Format(time.RFC3339),
			"original_hub_name": oldHub.Name,
		},
		CreatedBy: oldHub.CreatedBy,
	}

	err := hms.hubService.CreateInstance(ctx, instance)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

// migrateInbound migrates inbound configuration to protocol group and endpoint
func (hms *HubMigrationService) migrateInbound(ctx context.Context, instanceID string, inbound InboundConfig, createdBy string) (string, string, error) {
	// 1. Create protocol group for inbound
	protocolGroup := &models.HubProtocolGroup{
		ID:          uuid.New().String(),
		InstanceID:  instanceID,
		Direction:   "inbound",
		Name:        fmt.Sprintf("%s Inbound", inbound.Protocol),
		Description: fmt.Sprintf("Migrated inbound %s configuration", inbound.Protocol),
		Protocol:    inbound.Protocol,
		BaseConfig:  models.HubConfig{},
		MessageType: inbound.MessageType,
		Timeout:     inbound.Timeout,
		RetryAttempts: inbound.RetryAttempts,
		RetryInterval: inbound.RetryInterval,
		Enabled:     inbound.Enabled,
		CreatedBy:   createdBy,
	}

	// Set authentication from inbound
	if inbound.Authentication.Type != "" {
		protocolGroup.AuthType = inbound.Authentication.Type
		protocolGroup.AuthConfig = hms.convertAuthConfig(inbound.Authentication)
	}

	err := hms.hubService.CreateProtocolGroup(ctx, protocolGroup)
	if err != nil {
		return "", "", err
	}

	// 2. Create endpoint
	endpoint := &models.HubEndpoint{
		ID:              uuid.New().String(),
		ProtocolGroupID: protocolGroup.ID,
		Name:            "Main Endpoint",
		Description:     "Migrated inbound endpoint",
		EndpointURL:     inbound.Endpoint,
		Port:            inbound.Port,
		Path:            inbound.Path,
		Method:          inbound.Method,
		OverrideConfig:  models.HubConfig{},
		QueueConfig:     models.HubConfig(inbound.QueueConfig),
		FileConfig:      models.HubConfig(inbound.FileConfig),
		Enabled:         inbound.Enabled,
		CreatedBy:       createdBy,
	}

	err = hms.hubService.CreateEndpoint(ctx, endpoint)
	if err != nil {
		return protocolGroup.ID, "", err
	}

	return protocolGroup.ID, endpoint.ID, nil
}

// migrateOutbound migrates outbound configuration to protocol group and endpoint
func (hms *HubMigrationService) migrateOutbound(ctx context.Context, instanceID string, outbound OutboundConfig, createdBy string) (string, string, error) {
	// 1. Create protocol group for outbound
	protocolGroup := &models.HubProtocolGroup{
		ID:          uuid.New().String(),
		InstanceID:  instanceID,
		Direction:   "outbound",
		Name:        fmt.Sprintf("%s Outbound", outbound.Protocol),
		Description: fmt.Sprintf("Migrated outbound %s configuration", outbound.Protocol),
		Protocol:    outbound.Protocol,
		BaseConfig:  models.HubConfig{},
		MessageType: outbound.MessageType,
		Timeout:     outbound.Timeout,
		RetryAttempts: outbound.RetryAttempts,
		RetryInterval: outbound.RetryInterval,
		Enabled:     outbound.Enabled,
		CreatedBy:   createdBy,
	}

	// Set authorization from outbound
	if outbound.Authorization.Type != "" {
		protocolGroup.AuthType = outbound.Authorization.Type
		protocolGroup.AuthConfig = hms.convertAuthConfig(outbound.Authorization)
	}

	err := hms.hubService.CreateProtocolGroup(ctx, protocolGroup)
	if err != nil {
		return "", "", err
	}

	// 2. Create endpoint
	endpoint := &models.HubEndpoint{
		ID:              uuid.New().String(),
		ProtocolGroupID: protocolGroup.ID,
		Name:            "Main Destination",
		Description:     "Migrated outbound endpoint",
		EndpointURL:     outbound.Destination,
		Method:          outbound.Method,
		OverrideConfig:  models.HubConfig{},
		QueueConfig:     models.HubConfig(outbound.QueueConfig),
		FileConfig:      models.HubConfig(outbound.FileConfig),
		Enabled:         outbound.Enabled,
		CreatedBy:       createdBy,
	}

	err = hms.hubService.CreateEndpoint(ctx, endpoint)
	if err != nil {
		return protocolGroup.ID, "", err
	}

	return protocolGroup.ID, endpoint.ID, nil
}

// convertAuthConfig converts old auth config to new format
func (hms *HubMigrationService) convertAuthConfig(auth AuthConfig) models.HubConfig {
	config := make(models.HubConfig)

	if auth.Token != "" {
		config["token"] = auth.Token
	}
	if auth.APIKey != "" {
		config["api_key"] = auth.APIKey
	}
	if auth.Username != "" {
		config["username"] = auth.Username
	}
	if auth.Password != "" {
		config["password"] = auth.Password
	}
	if auth.OAuth2 != nil {
		config["oauth2"] = auth.OAuth2
	}
	if auth.Credentials != nil {
		for k, v := range auth.Credentials {
			config[k] = v
		}
	}

	return config
}

// logMigrationError logs migration error
func (hms *HubMigrationService) logMigrationError(ctx context.Context, migrationLog *models.HubMigrationLog, err error) {
	migrationLog.MigrationStatus = "failed"
	migrationLog.ErrorMessage = err.Error()
	hms.saveMigrationLog(ctx, migrationLog)
	hms.iLog.Error(fmt.Sprintf("Migration failed: %v", err))
}

// saveMigrationLog saves migration log to database
func (hms *HubMigrationService) saveMigrationLog(ctx context.Context, log *models.HubMigrationLog) error {
	detailsJSON, _ := json.Marshal(log.MigrationDetails)

	query := `
		INSERT INTO hub_migration_log (
			id, old_hub_id, new_instance_id, migration_status, migration_details,
			error_message, migrated_at, createdby, createdon
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Convert placeholders for the database type
	// For SQL Server, use numbered parameters @p1, @p2, etc.
	// For PostgreSQL, use $1, $2, etc.
	// For MySQL/SQLite, use ? as is
	// This is a simplified version - you may need to detect the database type
	_, err := hms.db.ExecContext(ctx, query,
		log.ID, log.OldHubID, log.NewInstanceID, log.MigrationStatus,
		string(detailsJSON), log.ErrorMessage, log.MigratedAt,
		log.CreatedBy, log.CreatedOn,
	)

	return err
}

// MigrateAll migrates all hubs from a provided list
func (hms *HubMigrationService) MigrateAll(ctx context.Context, oldHubs []OldIntegrationHub, targetInstanceID string) (int, int, error) {
	successCount := 0
	failCount := 0

	for _, oldHub := range oldHubs {
		err := hms.MigrateFromFlat(ctx, oldHub, targetInstanceID)
		if err != nil {
			hms.iLog.Error(fmt.Sprintf("Failed to migrate hub %s: %v", oldHub.Name, err))
			failCount++
			continue
		}
		successCount++
	}

	hms.iLog.Info(fmt.Sprintf("Migration complete: %d succeeded, %d failed", successCount, failCount))
	return successCount, failCount, nil
}
