// Copyright 2023 IAC. All Rights Reserved.
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

package apicallhistory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// APICallHistoryService manages API call history tracking
type APICallHistoryService struct {
	db            *sql.DB
	log           logger.Log
	config        *models.APICallHistoryConfig
	configMutex   sync.RWMutex
	buffer        []*models.APICallHistory
	bufferMutex   sync.Mutex
	bufferSize    int
	flushInterval time.Duration
	stopChan      chan struct{}
	historyTable  string
	configTable   string
}

// Global instance
var (
	globalAPICallHistoryService *APICallHistoryService
	globalServiceMutex          sync.RWMutex
)

// GetGlobalAPICallHistoryService returns the global service instance
func GetGlobalAPICallHistoryService() *APICallHistoryService {
	globalServiceMutex.RLock()
	defer globalServiceMutex.RUnlock()
	return globalAPICallHistoryService
}

// SetGlobalAPICallHistoryService sets the global service instance
func SetGlobalAPICallHistoryService(service *APICallHistoryService) {
	globalServiceMutex.Lock()
	defer globalServiceMutex.Unlock()
	globalAPICallHistoryService = service
}

// ServiceConfig holds configuration for the service
type ServiceConfig struct {
	HistoryTable  string
	ConfigTable   string
	BufferSize    int
	FlushInterval int // seconds
}

// DefaultServiceConfig returns default configuration
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		HistoryTable:  "iac_api_call_history",
		ConfigTable:   "iac_api_call_history_config",
		BufferSize:    100,
		FlushInterval: 5,
	}
}

// NewAPICallHistoryService creates a new API call history service
func NewAPICallHistoryService(cfg *ServiceConfig) (*APICallHistoryService, error) {
	if cfg == nil {
		cfg = DefaultServiceConfig()
	}

	// Get SQL database connection
	if dbconn.DB == nil {
		return nil, fmt.Errorf("SQL database connection not available")
	}

	service := &APICallHistoryService{
		db:            dbconn.DB,
		log:           logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "APICallHistoryService"},
		buffer:        make([]*models.APICallHistory, 0, cfg.BufferSize),
		bufferSize:    cfg.BufferSize,
		flushInterval: time.Duration(cfg.FlushInterval) * time.Second,
		stopChan:      make(chan struct{}),
		historyTable:  cfg.HistoryTable,
		configTable:   cfg.ConfigTable,
		config:        models.NewAPICallHistoryConfig(),
	}

	// Create tables if they don't exist
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := service.ensureTables(ctx); err != nil {
		service.log.Error(fmt.Sprintf("Failed to create tables: %v", err))
		return nil, err
	}

	// Load or create default configuration
	if err := service.loadOrCreateConfig(ctx); err != nil {
		service.log.Error(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Start background tasks
	go service.backgroundTasks()

	service.log.Info("API Call History Service initialized")
	return service, nil
}

// ensureTables creates the necessary tables if they don't exist
func (s *APICallHistoryService) ensureTables(ctx context.Context) error {
	// Detect database type by trying a simple query
	dbType := s.detectDatabaseType()
	s.log.Info(fmt.Sprintf("Detected database type: %s", dbType))

	// Use database-specific datetime type
	datetimeType := "DATETIME"
	defaultTimestamp := "CURRENT_TIMESTAMP"
	if dbType == "postgres" {
		datetimeType = "TIMESTAMP"
		defaultTimestamp = "CURRENT_TIMESTAMP"
	} else if dbType == "mysql" {
		datetimeType = "DATETIME(6)"
		defaultTimestamp = "CURRENT_TIMESTAMP(6)"
	}

	// Create history table
	historyTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(36) PRIMARY KEY,
			method VARCHAR(10) NOT NULL,
			endpoint VARCHAR(500) NOT NULL,
			full_path VARCHAR(2000),
			request_headers TEXT,
			request_body TEXT,
			query_params TEXT,
			status_code INT NOT NULL,
			response_body TEXT,
			response_headers TEXT,
			source_ip VARCHAR(45) NOT NULL,
			source_machine VARCHAR(255),
			user_agent VARCHAR(1000),
			user_id VARCHAR(100),
			user_name VARCHAR(255),
			client_id VARCHAR(100),
			auth_type VARCHAR(50),
			start_time %s NOT NULL,
			end_time %s NOT NULL,
			duration_ms BIGINT NOT NULL,
			instance_id VARCHAR(100),
			instance_name VARCHAR(255),
			error_message TEXT,
			tags TEXT,
			metadata TEXT,
			created_at %s DEFAULT %s
		)
	`, s.historyTable, datetimeType, datetimeType, datetimeType, defaultTimestamp)

	if _, err := s.db.ExecContext(ctx, historyTableSQL); err != nil {
		return fmt.Errorf("failed to create history table: %w", err)
	}

	// Create config table
	configTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(36) PRIMARY KEY,
			enabled BOOLEAN NOT NULL DEFAULT FALSE,
			include_endpoints TEXT,
			exclude_endpoints TEXT,
			include_methods TEXT,
			exclude_methods TEXT,
			include_source_ips TEXT,
			exclude_source_ips TEXT,
			include_users TEXT,
			exclude_users TEXT,
			include_status_codes TEXT,
			exclude_status_codes TEXT,
			only_errors BOOLEAN NOT NULL DEFAULT FALSE,
			capture_request_body BOOLEAN NOT NULL DEFAULT TRUE,
			capture_response_body BOOLEAN NOT NULL DEFAULT TRUE,
			capture_headers BOOLEAN NOT NULL DEFAULT FALSE,
			max_body_size INT NOT NULL DEFAULT 10240,
			mask_sensitive_fields TEXT,
			retention_days INT NOT NULL DEFAULT 30,
			sampling_rate DECIMAL(3,2) NOT NULL DEFAULT 1.00,
			refresh_interval INT NOT NULL DEFAULT 5,
			updated_by VARCHAR(100),
			updated_at %s DEFAULT %s,
			created_at %s DEFAULT %s
		)
	`, s.configTable, datetimeType, defaultTimestamp, datetimeType, defaultTimestamp)

	if _, err := s.db.ExecContext(ctx, configTableSQL); err != nil {
		return fmt.Errorf("failed to create config table: %w", err)
	}

	// Create indexes (ignore errors as syntax varies between databases)
	s.createIndexes(ctx, dbType)

	return nil
}

// backgroundTasks handles periodic tasks like buffer flushing and config refresh
func (s *APICallHistoryService) backgroundTasks() {
	flushTicker := time.NewTicker(s.flushInterval)
	defer flushTicker.Stop()

	// Initial check wait
	checkConfigTicker := time.NewTicker(30 * time.Second)
	defer checkConfigTicker.Stop()

	var lastConfigRefresh time.Time = time.Now()

	for {
		select {
		case <-flushTicker.C:
			s.flushBuffer()
		case <-checkConfigTicker.C:
			// Check if we need to refresh config based on dynamic interval
			s.configMutex.RLock()
			intervalMinutes := s.config.RefreshInterval
			if intervalMinutes <= 0 {
				intervalMinutes = 5 // Default
			}
			s.configMutex.RUnlock()

			if time.Since(lastConfigRefresh) >= time.Duration(intervalMinutes)*time.Minute {
				s.refreshConfig()
				lastConfigRefresh = time.Now()
			}
		case <-s.stopChan:
			s.flushBuffer() // Final flush before stopping
			return
		}
	}
}

// refreshConfig reloads the configuration from the database
func (s *APICallHistoryService) refreshConfig() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.loadOrCreateConfig(ctx); err != nil {
		s.log.Error(fmt.Sprintf("Failed to refresh config: %v", err))
	} else {
		// s.log.Debug("Configuration refreshed from database")
	}
}

// detectDatabaseType tries to detect the database type
func (s *APICallHistoryService) detectDatabaseType() string {
	// Check global configuration first
	if dbconn.DatabaseType != "" {
		if strings.Contains(strings.ToLower(dbconn.DatabaseType), "postgres") {
			return "postgres"
		}
		if strings.Contains(strings.ToLower(dbconn.DatabaseType), "mysql") || strings.Contains(strings.ToLower(dbconn.DatabaseType), "mariadb") {
			return "mysql"
		}
	}

	// Try PostgreSQL-specific query
	var version string
	err := s.db.QueryRow("SELECT version()").Scan(&version)
	if err == nil {
		if strings.Contains(strings.ToLower(version), "postgres") {
			return "postgres"
		}
		if strings.Contains(strings.ToLower(version), "mysql") || strings.Contains(strings.ToLower(version), "mariadb") {
			return "mysql"
		}
	}

	// Default to MySQL syntax as it's more common
	return "mysql"
}

// createIndexes creates indexes based on database type
func (s *APICallHistoryService) createIndexes(ctx context.Context, dbType string) {
	var indexes []string

	if dbType == "postgres" {
		indexes = []string{
			fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_start_time ON %s(start_time)", s.historyTable, s.historyTable),
			fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_endpoint ON %s(endpoint)", s.historyTable, s.historyTable),
			fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_method ON %s(method)", s.historyTable, s.historyTable),
			fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_status_code ON %s(status_code)", s.historyTable, s.historyTable),
			fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_user_id ON %s(user_id)", s.historyTable, s.historyTable),
		}
	} else {
		// MySQL syntax with prefix length for TEXT columns
		indexes = []string{
			fmt.Sprintf("CREATE INDEX idx_%s_start_time ON %s(start_time)", s.historyTable, s.historyTable),
			fmt.Sprintf("CREATE INDEX idx_%s_endpoint ON %s(endpoint(255))", s.historyTable, s.historyTable),
			fmt.Sprintf("CREATE INDEX idx_%s_method ON %s(method)", s.historyTable, s.historyTable),
			fmt.Sprintf("CREATE INDEX idx_%s_status_code ON %s(status_code)", s.historyTable, s.historyTable),
			fmt.Sprintf("CREATE INDEX idx_%s_user_id ON %s(user_id)", s.historyTable, s.historyTable),
		}
	}

	for _, idx := range indexes {
		// Ignore errors for indexes as they may already exist
		s.db.ExecContext(ctx, idx)
	}
}

// loadOrCreateConfig loads configuration from database or creates default
func (s *APICallHistoryService) loadOrCreateConfig(ctx context.Context) error {
	// Order by updated_at DESC to ensure we get the latest config if multiple exist
	query := fmt.Sprintf("SELECT id, enabled, include_endpoints, exclude_endpoints, include_methods, exclude_methods, include_source_ips, exclude_source_ips, include_users, exclude_users, include_status_codes, exclude_status_codes, only_errors, capture_request_body, capture_response_body, capture_headers, max_body_size, mask_sensitive_fields, retention_days, sampling_rate, refresh_interval, updated_by, updated_at FROM %s ORDER BY updated_at DESC LIMIT 1", s.configTable)

	row := s.db.QueryRowContext(ctx, query)

	var config models.APICallHistoryConfig
	var includeEndpoints, excludeEndpoints, includeMethods, excludeMethods sql.NullString
	var includeSourceIPs, excludeSourceIPs, includeUsers, excludeUsers sql.NullString
	var includeStatusCodes, excludeStatusCodes, maskSensitiveFields sql.NullString
	var updatedAt sql.NullTime
	// Helper variable for scanning nullable int
	var refreshInterval sql.NullInt64

	err := row.Scan(
		&config.ID,
		&config.Enabled,
		&includeEndpoints,
		&excludeEndpoints,
		&includeMethods,
		&excludeMethods,
		&includeSourceIPs,
		&excludeSourceIPs,
		&includeUsers,
		&excludeUsers,
		&includeStatusCodes,
		&excludeStatusCodes,
		&config.OnlyErrors,
		&config.CaptureRequestBody,
		&config.CaptureResponseBody,
		&config.CaptureHeaders,
		&config.MaxBodySize,
		&maskSensitiveFields,
		&config.RetentionDays,
		&config.SamplingRate,
		&refreshInterval,
		&config.UpdatedBy,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		// Create default configuration
		config = *models.NewAPICallHistoryConfig()
		config.ID = uuid.New().String()

		if err := s.insertConfig(ctx, &config); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
		s.log.Info("Created default API call history configuration")
	} else if err != nil {
		// Column might be missing if upgrading from older schema
		if strings.Contains(err.Error(), "no such column") || strings.Contains(err.Error(), "does not exist") {
			s.log.Warn("refresh_interval column missing, attempting migration...")
			if migrateErr := s.migrateConfigTable(ctx); migrateErr != nil {
				s.log.Error(fmt.Sprintf("Migration failed: %v", migrateErr))
			}
			// Retry load after migration (recursion guarded by finite retries would be better but simple retry here)
			// For now, just return this error and next startup/refresh will handle it
			return fmt.Errorf("schema mismatch, migration attempted: %w", err)
		}
		return fmt.Errorf("failed to load config: %w", err)
	} else {
		// Parse JSON arrays
		config.IncludeEndpoints = parseJSONArray(includeEndpoints.String)
		config.ExcludeEndpoints = parseJSONArray(excludeEndpoints.String)
		config.IncludeMethods = parseJSONArray(includeMethods.String)
		config.ExcludeMethods = parseJSONArray(excludeMethods.String)
		config.IncludeSourceIPs = parseJSONArray(includeSourceIPs.String)
		config.ExcludeSourceIPs = parseJSONArray(excludeSourceIPs.String)
		config.IncludeUsers = parseJSONArray(includeUsers.String)
		config.ExcludeUsers = parseJSONArray(excludeUsers.String)
		config.IncludeStatusCodes = parseJSONIntArray(includeStatusCodes.String)
		config.ExcludeStatusCodes = parseJSONIntArray(excludeStatusCodes.String)
		config.MaskSensitiveFields = parseJSONArray(maskSensitiveFields.String)
		config.RefreshInterval = int(refreshInterval.Int64)
		if updatedAt.Valid {
			config.UpdatedAt = updatedAt.Time
		}
	}

	s.configMutex.Lock()
	s.config = &config
	s.configMutex.Unlock()

	return nil
}

// migrateConfigTable adds missing columns to config table
func (s *APICallHistoryService) migrateConfigTable(ctx context.Context) error {
	alterSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN refresh_interval INT NOT NULL DEFAULT 5", s.configTable)
	_, err := s.db.ExecContext(ctx, alterSQL)
	return err
}

// insertConfig inserts a new configuration record
func (s *APICallHistoryService) insertConfig(ctx context.Context, config *models.APICallHistoryConfig) error {
	dbType := s.detectDatabaseType()

	query := fmt.Sprintf(`
		INSERT INTO %s (id, enabled, include_endpoints, exclude_endpoints, include_methods, exclude_methods,
			include_source_ips, exclude_source_ips, include_users, exclude_users, include_status_codes,
			exclude_status_codes, only_errors, capture_request_body, capture_response_body, capture_headers,
			max_body_size, mask_sensitive_fields, retention_days, sampling_rate, refresh_interval, updated_by, updated_at, created_at)
		VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
	`, s.configTable,
		s.getPlaceholder(dbType, 1), s.getPlaceholder(dbType, 2), s.getPlaceholder(dbType, 3),
		s.getPlaceholder(dbType, 4), s.getPlaceholder(dbType, 5), s.getPlaceholder(dbType, 6),
		s.getPlaceholder(dbType, 7), s.getPlaceholder(dbType, 8), s.getPlaceholder(dbType, 9),
		s.getPlaceholder(dbType, 10), s.getPlaceholder(dbType, 11), s.getPlaceholder(dbType, 12),
		s.getPlaceholder(dbType, 13), s.getPlaceholder(dbType, 14), s.getPlaceholder(dbType, 15),
		s.getPlaceholder(dbType, 16), s.getPlaceholder(dbType, 17), s.getPlaceholder(dbType, 18),
		s.getPlaceholder(dbType, 19), s.getPlaceholder(dbType, 20), s.getPlaceholder(dbType, 21),
		s.getPlaceholder(dbType, 22), s.getPlaceholder(dbType, 23), s.getPlaceholder(dbType, 24))

	s.log.Info(fmt.Sprintf("insertConfig: inserting new config with ID=%s", config.ID))

	_, err := s.db.ExecContext(ctx, query,
		config.ID,
		config.Enabled,
		toJSONArray(config.IncludeEndpoints),
		toJSONArray(config.ExcludeEndpoints),
		toJSONArray(config.IncludeMethods),
		toJSONArray(config.ExcludeMethods),
		toJSONArray(config.IncludeSourceIPs),
		toJSONArray(config.ExcludeSourceIPs),
		toJSONArray(config.IncludeUsers),
		toJSONArray(config.ExcludeUsers),
		toJSONIntArray(config.IncludeStatusCodes),
		toJSONIntArray(config.ExcludeStatusCodes),
		config.OnlyErrors,
		config.CaptureRequestBody,
		config.CaptureResponseBody,
		config.CaptureHeaders,
		config.MaxBodySize,
		toJSONArray(config.MaskSensitiveFields),
		config.RetentionDays,
		config.SamplingRate,
		config.RefreshInterval,
		config.UpdatedBy,
		time.Now(),
		time.Now(),
	)

	return err
}

// GetConfig returns the current configuration
func (s *APICallHistoryService) GetConfig() *models.APICallHistoryConfig {
	s.configMutex.RLock()
	defer s.configMutex.RUnlock()

	if s.config == nil {
		// Return a default config if not initialized (requests won't be tracked, but UI won't break)
		// We return a copy to avoid concurrency issues if we were to set it
		defaultConfig := models.NewAPICallHistoryConfig()
		defaultConfig.Enabled = false
		return defaultConfig
	}
	return s.config
}

// UpdateConfig updates the configuration
func (s *APICallHistoryService) UpdateConfig(ctx context.Context, config *models.APICallHistoryConfig, updatedBy string) error {
	config.UpdatedBy = updatedBy
	config.UpdatedAt = time.Now()

	// If ID is missing, generate one
	if config.ID == "" {
		config.ID = uuid.New().String()
	}

	dbType := s.detectDatabaseType()

	// Try to update efficiently
	query := fmt.Sprintf(`
		UPDATE %s SET
			enabled = %s, include_endpoints = %s, exclude_endpoints = %s, include_methods = %s, exclude_methods = %s,
			include_source_ips = %s, exclude_source_ips = %s, include_users = %s, exclude_users = %s,
			include_status_codes = %s, exclude_status_codes = %s, only_errors = %s, capture_request_body = %s,
			capture_response_body = %s, capture_headers = %s, max_body_size = %s, mask_sensitive_fields = %s,
			retention_days = %s, sampling_rate = %s, refresh_interval = %s, updated_by = %s, updated_at = %s
		WHERE id = %s
	`, s.configTable,
		s.getPlaceholder(dbType, 1), s.getPlaceholder(dbType, 2), s.getPlaceholder(dbType, 3),
		s.getPlaceholder(dbType, 4), s.getPlaceholder(dbType, 5), s.getPlaceholder(dbType, 6),
		s.getPlaceholder(dbType, 7), s.getPlaceholder(dbType, 8), s.getPlaceholder(dbType, 9),
		s.getPlaceholder(dbType, 10), s.getPlaceholder(dbType, 11), s.getPlaceholder(dbType, 12),
		s.getPlaceholder(dbType, 13), s.getPlaceholder(dbType, 14), s.getPlaceholder(dbType, 15),
		s.getPlaceholder(dbType, 16), s.getPlaceholder(dbType, 17), s.getPlaceholder(dbType, 18),
		s.getPlaceholder(dbType, 19), s.getPlaceholder(dbType, 20), s.getPlaceholder(dbType, 21),
		s.getPlaceholder(dbType, 22), s.getPlaceholder(dbType, 23))

	s.log.Info(fmt.Sprintf("UpdateConfig: dbType=%s", dbType))
	s.log.Info(fmt.Sprintf("UpdateConfig: query=%s", query))

	result, err := s.db.ExecContext(ctx, query,
		config.Enabled,
		toJSONArray(config.IncludeEndpoints),
		toJSONArray(config.ExcludeEndpoints),
		toJSONArray(config.IncludeMethods),
		toJSONArray(config.ExcludeMethods),
		toJSONArray(config.IncludeSourceIPs),
		toJSONArray(config.ExcludeSourceIPs),
		toJSONArray(config.IncludeUsers),
		toJSONArray(config.ExcludeUsers),
		toJSONIntArray(config.IncludeStatusCodes),
		toJSONIntArray(config.ExcludeStatusCodes),
		config.OnlyErrors,
		config.CaptureRequestBody,
		config.CaptureResponseBody,
		config.CaptureHeaders,
		config.MaxBodySize,
		toJSONArray(config.MaskSensitiveFields),
		config.RetentionDays,
		config.SamplingRate,
		config.RefreshInterval,
		config.UpdatedBy,
		config.UpdatedAt,
		config.ID,
	)

	if err != nil {
		// Handle missing column on update
		if strings.Contains(err.Error(), "no such column") || strings.Contains(err.Error(), "does not exist") {
			s.log.Warn("refresh_interval column missing during update, attempting migration...")
			if migrateErr := s.migrateConfigTable(ctx); migrateErr != nil {
				return fmt.Errorf("migration failed during update: %w", migrateErr)
			}
			// Retry update recursively (one level deep ideally, but this is simple)
			return s.UpdateConfig(ctx, config, updatedBy)
		}
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		// Record does not exist, insert it
		return s.insertConfig(ctx, config)
	}

	s.configMutex.Lock()
	s.config = config
	s.configMutex.Unlock()

	return nil
}

// ResetConfig resets configuration to defaults
func (s *APICallHistoryService) ResetConfig(ctx context.Context, updatedBy string) error {
	s.configMutex.RLock()
	var configID string
	if s.config != nil {
		configID = s.config.ID
	}
	s.configMutex.RUnlock()

	newConfig := models.NewAPICallHistoryConfig()
	if configID != "" {
		newConfig.ID = configID
	} else {
		newConfig.ID = uuid.New().String()
	}
	return s.UpdateConfig(ctx, newConfig, updatedBy)
}

// RecordCall adds an API call to the buffer for async saving
func (s *APICallHistoryService) RecordCall(history *models.APICallHistory) {
	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()

	if history.ID == "" {
		history.ID = uuid.New().String()
	}

	s.buffer = append(s.buffer, history)

	if len(s.buffer) >= s.bufferSize {
		go s.flushBuffer()
	}
}

// flushBuffer saves all buffered records to database
func (s *APICallHistoryService) flushBuffer() {
	s.bufferMutex.Lock()
	if len(s.buffer) == 0 {
		s.bufferMutex.Unlock()
		return
	}

	records := s.buffer
	s.buffer = make([]*models.APICallHistory, 0, s.bufferSize)
	s.bufferMutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, record := range records {
		if err := s.insertHistory(ctx, record); err != nil {
			s.log.Error(fmt.Sprintf("Failed to save API call history: %v", err))
		}
	}
}

// insertHistory inserts a single history record
func (s *APICallHistoryService) insertHistory(ctx context.Context, history *models.APICallHistory) error {
	dbType := s.detectDatabaseType()

	query := fmt.Sprintf(`
		INSERT INTO %s (id, method, endpoint, full_path, request_headers, request_body, query_params,
			status_code, response_body, response_headers, source_ip, source_machine, user_agent,
			user_id, user_name, client_id, auth_type, start_time, end_time, duration_ms,
			instance_id, instance_name, error_message, tags, metadata, created_at)
		VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
	`, s.historyTable,
		s.getPlaceholder(dbType, 1), s.getPlaceholder(dbType, 2), s.getPlaceholder(dbType, 3),
		s.getPlaceholder(dbType, 4), s.getPlaceholder(dbType, 5), s.getPlaceholder(dbType, 6),
		s.getPlaceholder(dbType, 7), s.getPlaceholder(dbType, 8), s.getPlaceholder(dbType, 9),
		s.getPlaceholder(dbType, 10), s.getPlaceholder(dbType, 11), s.getPlaceholder(dbType, 12),
		s.getPlaceholder(dbType, 13), s.getPlaceholder(dbType, 14), s.getPlaceholder(dbType, 15),
		s.getPlaceholder(dbType, 16), s.getPlaceholder(dbType, 17), s.getPlaceholder(dbType, 18),
		s.getPlaceholder(dbType, 19), s.getPlaceholder(dbType, 20), s.getPlaceholder(dbType, 21),
		s.getPlaceholder(dbType, 22), s.getPlaceholder(dbType, 23), s.getPlaceholder(dbType, 24),
		s.getPlaceholder(dbType, 25), s.getPlaceholder(dbType, 26))

	_, err := s.db.ExecContext(ctx, query,
		history.ID,
		history.Method,
		history.Endpoint,
		history.FullPath,
		toJSON(history.RequestHeaders),
		toJSON(history.RequestBody),
		toJSON(history.QueryParams),
		history.StatusCode,
		toJSON(history.ResponseBody),
		toJSON(history.ResponseHeaders),
		history.SourceIP,
		history.SourceMachine,
		history.UserAgent,
		history.UserID,
		history.UserName,
		history.ClientID,
		history.AuthType,
		history.StartTime,
		history.EndTime,
		history.DurationMs,
		history.InstanceID,
		history.InstanceName,
		history.ErrorMessage,
		toJSONArray(history.Tags),
		toJSON(history.Metadata),
		time.Now(),
	)

	return err
}

// ListHistory returns paginated history records with optional filters
func (s *APICallHistoryService) ListHistory(ctx context.Context, params models.ListHistoryParams) (*models.ListHistoryResponse, error) {
	// Detect database type
	dbType := s.detectDatabaseType()

	var conditions []string
	var args []interface{}

	// Build WHERE clause with placeholders
	placeholderIndex := 1

	if params.Method != "" {
		conditions = append(conditions, fmt.Sprintf("method = %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.Method)
		placeholderIndex++
	}
	if params.Endpoint != "" {
		conditions = append(conditions, fmt.Sprintf("endpoint LIKE %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, "%"+params.Endpoint+"%")
		placeholderIndex++
	}
	if params.StatusCode > 0 {
		conditions = append(conditions, fmt.Sprintf("status_code = %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.StatusCode)
		placeholderIndex++
	}
	if params.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.UserID)
		placeholderIndex++
	}
	if params.SourceIP != "" {
		conditions = append(conditions, fmt.Sprintf("source_ip = %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.SourceIP)
		placeholderIndex++
	}
	if !params.StartDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("start_time >= %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.StartDate)
		placeholderIndex++
	}
	if !params.EndDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("start_time <= %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.EndDate)
		placeholderIndex++
	}
	if params.OnlyErrors {
		conditions = append(conditions, "status_code >= 400")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count (reuse same args, placeholders are already set correctly)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", s.historyTable, whereClause)
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Get paginated results
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	offset := (params.Page - 1) * params.PageSize

	// Build pagination clause based on database type
	var selectQuery string
	var queryArgs []interface{} = args // Start with existing args

	switch dbType {
	case "postgres", "mysql":
		// PostgreSQL and MySQL support LIMIT/OFFSET
		selectQuery = fmt.Sprintf(`
			SELECT id, method, endpoint, status_code, duration_ms, user_id, user_name, source_ip, start_time
			FROM %s %s
			ORDER BY start_time DESC
			LIMIT %s OFFSET %s
		`, s.historyTable, whereClause, s.getPlaceholder(dbType, placeholderIndex), s.getPlaceholder(dbType, placeholderIndex+1))
		queryArgs = append(queryArgs, params.PageSize, offset)

	case "mssql", "oracle":
		// SQL Server and Oracle use OFFSET/FETCH (requires ORDER BY)
		selectQuery = fmt.Sprintf(`
			SELECT id, method, endpoint, status_code, duration_ms, user_id, user_name, source_ip, start_time
			FROM %s %s
			ORDER BY start_time DESC
			OFFSET %s ROWS FETCH NEXT %s ROWS ONLY
		`, s.historyTable, whereClause, s.getPlaceholder(dbType, placeholderIndex), s.getPlaceholder(dbType, placeholderIndex+1))
		queryArgs = append(queryArgs, offset, params.PageSize)

	default:
		// Default to LIMIT/OFFSET
		selectQuery = fmt.Sprintf(`
			SELECT id, method, endpoint, status_code, duration_ms, user_id, user_name, source_ip, start_time
			FROM %s %s
			ORDER BY start_time DESC
			LIMIT %s OFFSET %s
		`, s.historyTable, whereClause, s.getPlaceholder(dbType, placeholderIndex), s.getPlaceholder(dbType, placeholderIndex+1))
		queryArgs = append(queryArgs, params.PageSize, offset)
	}

	rows, err := s.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.APICallHistoryListItem
	for rows.Next() {
		var item models.APICallHistoryListItem
		var userName sql.NullString
		if err := rows.Scan(&item.ID, &item.Method, &item.Endpoint, &item.StatusCode, &item.DurationMs, &item.UserID, &userName, &item.SourceIP, &item.StartTime); err != nil {
			return nil, err
		}
		item.UserName = userName.String
		history = append(history, item)
	}

	return &models.ListHistoryResponse{
		History: history,
		Total:   total,
		Page:    params.Page,
		Pages:   (total + int64(params.PageSize) - 1) / int64(params.PageSize),
	}, nil
}

// GetHistory returns a single history record by ID
func (s *APICallHistoryService) GetHistory(ctx context.Context, id string) (*models.APICallHistory, error) {
	query := fmt.Sprintf(`
		SELECT id, method, endpoint, full_path, request_headers, request_body, query_params,
			status_code, response_body, response_headers, source_ip, source_machine, user_agent,
			user_id, user_name, client_id, auth_type, start_time, end_time, duration_ms,
			instance_id, instance_name, error_message, tags, metadata
		FROM %s WHERE id = ?
	`, s.historyTable)

	row := s.db.QueryRowContext(ctx, query, id)

	var history models.APICallHistory
	var requestHeaders, requestBody, queryParams, responseBody, responseHeaders sql.NullString
	var sourceMachine, userAgent, userID, userName, clientID, authType sql.NullString
	var instanceID, instanceName, errorMessage, tags, metadata sql.NullString

	err := row.Scan(
		&history.ID,
		&history.Method,
		&history.Endpoint,
		&history.FullPath,
		&requestHeaders,
		&requestBody,
		&queryParams,
		&history.StatusCode,
		&responseBody,
		&responseHeaders,
		&history.SourceIP,
		&sourceMachine,
		&userAgent,
		&userID,
		&userName,
		&clientID,
		&authType,
		&history.StartTime,
		&history.EndTime,
		&history.DurationMs,
		&instanceID,
		&instanceName,
		&errorMessage,
		&tags,
		&metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("record not found")
	}
	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	history.RequestHeaders = parseJSONMap(requestHeaders.String)
	history.RequestBody = parseJSONInterface(requestBody.String)
	history.QueryParams = parseJSONMap(queryParams.String)
	history.ResponseBody = parseJSONInterface(responseBody.String)
	history.ResponseHeaders = parseJSONMap(responseHeaders.String)
	history.SourceMachine = sourceMachine.String
	history.UserAgent = userAgent.String
	history.UserID = userID.String
	history.UserName = userName.String
	history.ClientID = clientID.String
	history.AuthType = authType.String
	history.InstanceID = instanceID.String
	history.InstanceName = instanceName.String
	history.ErrorMessage = errorMessage.String
	history.Tags = parseJSONArray(tags.String)
	history.Metadata = parseJSONMapInterface(metadata.String)

	return &history, nil
}

// DeleteHistory deletes a history record by ID
func (s *APICallHistoryService) DeleteHistory(ctx context.Context, id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", s.historyTable)
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("record not found")
	}

	return nil
}

// Cleanup deletes records older than the specified date
func (s *APICallHistoryService) Cleanup(ctx context.Context, before time.Time) (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s WHERE start_time < ?", s.historyTable)
	result, err := s.db.ExecContext(ctx, query, before)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// GetStats returns statistics about API calls
func (s *APICallHistoryService) GetStats(ctx context.Context) (*models.APICallHistoryStats, error) {
	stats := &models.APICallHistoryStats{
		EndpointStats: make([]models.EndpointStat, 0),
		UserStats:     make([]models.UserStat, 0),
	}

	// Total calls
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", s.historyTable)
	s.db.QueryRowContext(ctx, countQuery).Scan(&stats.TotalCalls)

	// Average duration
	avgQuery := fmt.Sprintf("SELECT COALESCE(AVG(duration_ms), 0) FROM %s", s.historyTable)
	s.db.QueryRowContext(ctx, avgQuery).Scan(&stats.AvgDurationMs)

	// Error count
	errorQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE status_code >= 400", s.historyTable)
	s.db.QueryRowContext(ctx, errorQuery).Scan(&stats.ErrorCount)

	// Calculate error rate
	if stats.TotalCalls > 0 {
		stats.ErrorRate = float64(stats.ErrorCount) / float64(stats.TotalCalls) * 100
	}

	// Endpoint stats (top 10)
	endpointQuery := fmt.Sprintf(`
		SELECT endpoint, COUNT(*) as count, AVG(duration_ms) as avg_duration
		FROM %s
		GROUP BY endpoint
		ORDER BY count DESC
		LIMIT 10
	`, s.historyTable)

	rows, err := s.db.QueryContext(ctx, endpointQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat models.EndpointStat
			rows.Scan(&stat.Endpoint, &stat.Count, &stat.AvgDurationMs)
			stats.EndpointStats = append(stats.EndpointStats, stat)
		}
	}

	// User stats (top 10)
	userQuery := fmt.Sprintf(`
		SELECT COALESCE(user_id, 'anonymous') as user_id, COUNT(*) as count
		FROM %s
		GROUP BY user_id
		ORDER BY count DESC
		LIMIT 10
	`, s.historyTable)

	rows, err = s.db.QueryContext(ctx, userQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat models.UserStat
			rows.Scan(&stat.UserID, &stat.Count)
			stats.UserStats = append(stats.UserStats, stat)
		}
	}

	return stats, nil
}

// Stop stops the background flush goroutine
func (s *APICallHistoryService) Stop() {
	close(s.stopChan)
}

// Helper functions for JSON serialization
func toJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func toJSONArray(arr []string) string {
	if arr == nil || len(arr) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(arr)
	return string(b)
}

func toJSONIntArray(arr []int) string {
	if arr == nil || len(arr) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(arr)
	return string(b)
}

func parseJSONArray(s string) []string {
	if s == "" || s == "[]" {
		return []string{}
	}
	var arr []string
	json.Unmarshal([]byte(s), &arr)
	return arr
}

func parseJSONIntArray(s string) []int {
	if s == "" || s == "[]" {
		return []int{}
	}
	var arr []int
	json.Unmarshal([]byte(s), &arr)
	return arr
}

func parseJSONMap(s string) map[string]string {
	if s == "" {
		return nil
	}
	var m map[string]string
	json.Unmarshal([]byte(s), &m)
	return m
}

func parseJSONMapInterface(s string) map[string]interface{} {
	if s == "" {
		return nil
	}
	var m map[string]interface{}
	json.Unmarshal([]byte(s), &m)
	return m
}

func parseJSONInterface(s string) interface{} {
	if s == "" {
		return nil
	}
	var v interface{}
	json.Unmarshal([]byte(s), &v)
	return v
}

// getPlaceholder returns the appropriate placeholder for the database type
// PostgreSQL uses $1, $2, etc., while others use ?
func (s *APICallHistoryService) getPlaceholder(dbType string, index int) string {
	if dbType == "postgres" {
		return fmt.Sprintf("$%d", index)
	}
	return "?"
}
