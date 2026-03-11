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
	"github.com/lib/pq"
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Load configuration from existing table (table must exist via SQL migration)
	if err := service.loadOrCreateConfig(ctx); err != nil {
		service.log.Error(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Start background tasks
	go service.backgroundTasks()

	service.log.Info("API Call History Service initialized")
	return service, nil
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

// detectDatabaseType returns the detected database type string.
func (s *APICallHistoryService) detectDatabaseType() string {
	if dbconn.DatabaseType != "" {
		t := strings.ToLower(dbconn.DatabaseType)
		if strings.Contains(t, "postgres") {
			return "postgres"
		}
		if strings.Contains(t, "mysql") || strings.Contains(t, "mariadb") {
			return "mysql"
		}
	}
	var version string
	if err := s.db.QueryRow("SELECT version()").Scan(&version); err == nil {
		v := strings.ToLower(version)
		if strings.Contains(v, "postgres") {
			return "postgres"
		}
		if strings.Contains(v, "mysql") || strings.Contains(v, "mariadb") {
			return "mysql"
		}
	}
	return "mysql"
}

// loadOrCreateConfig loads configuration from database or creates default
func (s *APICallHistoryService) loadOrCreateConfig(ctx context.Context) error {
	// Order by updated_at DESC to ensure we get the latest config if multiple exist
	query := fmt.Sprintf("SELECT id, enabled, include_endpoints, exclude_endpoints, include_methods, exclude_methods, include_source_ips, exclude_source_ips, include_users, exclude_users, include_status_codes, exclude_status_codes, only_errors, capture_request_body, capture_response_body, capture_headers, max_body_size, mask_sensitive_fields, retention_days, sampling_rate, refresh_interval, updated_by, updated_at FROM %s ORDER BY updated_at DESC LIMIT 1", s.configTable)

	row := s.db.QueryRowContext(ctx, query)

	var config models.APICallHistoryConfig
	var updatedAt sql.NullTime
	var refreshInterval sql.NullInt64

	err := row.Scan(
		&config.ID,
		&config.Enabled,
		pq.Array(&config.IncludeEndpoints),
		pq.Array(&config.ExcludeEndpoints),
		pq.Array(&config.IncludeMethods),
		pq.Array(&config.ExcludeMethods),
		pq.Array(&config.IncludeSourceIPs),
		pq.Array(&config.ExcludeSourceIPs),
		pq.Array(&config.IncludeUsers),
		pq.Array(&config.ExcludeUsers),
		pq.Array(&config.IncludeStatusCodes),
		pq.Array(&config.ExcludeStatusCodes),
		&config.OnlyErrors,
		&config.CaptureRequestBody,
		&config.CaptureResponseBody,
		&config.CaptureHeaders,
		&config.MaxBodySize,
		pq.Array(&config.MaskSensitiveFields),
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
		// Arrays are already populated directly into config fields via pq.Array() scanner.
		// Ensure nil slices become empty slices for consistent JSON serialisation.
		if config.IncludeEndpoints == nil {
			config.IncludeEndpoints = []string{}
		}
		if config.ExcludeEndpoints == nil {
			config.ExcludeEndpoints = []string{}
		}
		if config.IncludeMethods == nil {
			config.IncludeMethods = []string{}
		}
		if config.ExcludeMethods == nil {
			config.ExcludeMethods = []string{}
		}
		if config.IncludeSourceIPs == nil {
			config.IncludeSourceIPs = []string{}
		}
		if config.ExcludeSourceIPs == nil {
			config.ExcludeSourceIPs = []string{}
		}
		if config.IncludeUsers == nil {
			config.IncludeUsers = []string{}
		}
		if config.ExcludeUsers == nil {
			config.ExcludeUsers = []string{}
		}
		if config.IncludeStatusCodes == nil {
			config.IncludeStatusCodes = []int{}
		}
		if config.ExcludeStatusCodes == nil {
			config.ExcludeStatusCodes = []int{}
		}
		if config.MaskSensitiveFields == nil {
			config.MaskSensitiveFields = []string{}
		}
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
		pq.Array(config.IncludeEndpoints),
		pq.Array(config.ExcludeEndpoints),
		pq.Array(config.IncludeMethods),
		pq.Array(config.ExcludeMethods),
		pq.Array(config.IncludeSourceIPs),
		pq.Array(config.ExcludeSourceIPs),
		pq.Array(config.IncludeUsers),
		pq.Array(config.ExcludeUsers),
		pq.Array(config.IncludeStatusCodes),
		pq.Array(config.ExcludeStatusCodes),
		config.OnlyErrors,
		config.CaptureRequestBody,
		config.CaptureResponseBody,
		config.CaptureHeaders,
		config.MaxBodySize,
		pq.Array(config.MaskSensitiveFields),
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
		pq.Array(config.IncludeEndpoints),
		pq.Array(config.ExcludeEndpoints),
		pq.Array(config.IncludeMethods),
		pq.Array(config.ExcludeMethods),
		pq.Array(config.IncludeSourceIPs),
		pq.Array(config.ExcludeSourceIPs),
		pq.Array(config.IncludeUsers),
		pq.Array(config.ExcludeUsers),
		pq.Array(config.IncludeStatusCodes),
		pq.Array(config.ExcludeStatusCodes),
		config.OnlyErrors,
		config.CaptureRequestBody,
		config.CaptureResponseBody,
		config.CaptureHeaders,
		config.MaxBodySize,
		pq.Array(config.MaskSensitiveFields),
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
		pq.Array(history.Tags),
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
	dbType := s.detectDatabaseType()
	query := fmt.Sprintf(`
		SELECT id, method, endpoint, full_path, request_headers, request_body, query_params,
			status_code, response_body, response_headers, source_ip, source_machine, user_agent,
			user_id, user_name, client_id, auth_type, start_time, end_time, duration_ms,
			instance_id, instance_name, error_message, tags, metadata
		FROM %s WHERE id = %s
	`, s.historyTable, s.getPlaceholder(dbType, 1))

	row := s.db.QueryRowContext(ctx, query, id)

	var history models.APICallHistory
	var requestHeaders, requestBody, queryParams, responseBody, responseHeaders sql.NullString
	var sourceMachine, userAgent, userID, userName, clientID, authType sql.NullString
	var instanceID, instanceName, errorMessage, metadata sql.NullString

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
		pq.Array(&history.Tags),
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
	history.Metadata = parseJSONMapInterface(metadata.String)

	return &history, nil
}

// DeleteHistory deletes a history record by ID
func (s *APICallHistoryService) DeleteHistory(ctx context.Context, id string) error {
	dbType := s.detectDatabaseType()
	query := fmt.Sprintf("DELETE FROM %s WHERE id = %s", s.historyTable, s.getPlaceholder(dbType, 1))
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
	dbType := s.detectDatabaseType()
	query := fmt.Sprintf("DELETE FROM %s WHERE start_time < %s", s.historyTable, s.getPlaceholder(dbType, 1))
	result, err := s.db.ExecContext(ctx, query, before)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// GetStats returns statistics about API calls
func (s *APICallHistoryService) GetStats(ctx context.Context) (*models.APICallHistoryStats, error) {
	stats := &models.APICallHistoryStats{
		TopEndpoints:  make([]models.EndpointStat, 0),
		TopUsers:      make([]models.UserStat, 0),
		CallsByMethod: make(map[string]int64),
		CallsByStatus: make(map[string]int64),
	}

	// 1. Summary: total, successful, failed, avg/min/max duration in one pass
	summaryQuery := fmt.Sprintf(`
		SELECT
			COUNT(*) as total_calls,
			COUNT(CASE WHEN status_code < 400 THEN 1 END) as successful_calls,
			COUNT(CASE WHEN status_code >= 400 THEN 1 END) as failed_calls,
			COALESCE(AVG(duration_ms), 0) as avg_duration_ms,
			COALESCE(MIN(duration_ms), 0) as min_duration_ms,
			COALESCE(MAX(duration_ms), 0) as max_duration_ms
		FROM %s`, s.historyTable)

	var avgDur, minDur, maxDur float64
	if err := s.db.QueryRowContext(ctx, summaryQuery).Scan(
		&stats.TotalCalls, &stats.SuccessfulCalls, &stats.FailedCalls,
		&avgDur, &minDur, &maxDur,
	); err != nil {
		s.log.Error(fmt.Sprintf("GetStats summary query failed: %v", err))
	}
	stats.AverageDurationMs = avgDur
	stats.MinDurationMs = int64(minDur)
	stats.MaxDurationMs = int64(maxDur)
	if stats.TotalCalls > 0 {
		stats.ErrorRate = float64(stats.FailedCalls) / float64(stats.TotalCalls) * 100
	}

	// 2. Calls by HTTP method
	methodQuery := fmt.Sprintf(`SELECT method, COUNT(*) FROM %s GROUP BY method ORDER BY COUNT(*) DESC`, s.historyTable)
	if rows, err := s.db.QueryContext(ctx, methodQuery); err == nil {
		defer rows.Close()
		for rows.Next() {
			var method string
			var count int64
			if err := rows.Scan(&method, &count); err == nil {
				stats.CallsByMethod[method] = count
			}
		}
	}

	// 3. Calls by status code
	statusQuery := fmt.Sprintf(`SELECT status_code, COUNT(*) FROM %s GROUP BY status_code ORDER BY status_code`, s.historyTable)
	if rows, err := s.db.QueryContext(ctx, statusQuery); err == nil {
		defer rows.Close()
		for rows.Next() {
			var code int
			var count int64
			if err := rows.Scan(&code, &count); err == nil {
				stats.CallsByStatus[fmt.Sprintf("%d", code)] = count
			}
		}
	}

	// 4. Top 10 endpoints by call count (with avg duration and error rate)
	endpointQuery := fmt.Sprintf(`
		SELECT endpoint,
			COUNT(*) as total_calls,
			COALESCE(AVG(duration_ms), 0) as avg_duration_ms,
			COUNT(CASE WHEN status_code >= 400 THEN 1 END) * 100.0 / COUNT(*) as error_rate
		FROM %s
		GROUP BY endpoint
		ORDER BY total_calls DESC
		LIMIT 10`, s.historyTable)

	if rows, err := s.db.QueryContext(ctx, endpointQuery); err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat models.EndpointStat
			if err := rows.Scan(&stat.Endpoint, &stat.TotalCalls, &stat.AvgDurationMs, &stat.ErrorRate); err == nil {
				stats.TopEndpoints = append(stats.TopEndpoints, stat)
			}
		}
	}

	// 5. Top 10 users by call count
	userQuery := fmt.Sprintf(`
		SELECT COALESCE(NULLIF(user_id, ''), 'anonymous') as user_id, COUNT(*) as total_calls
		FROM %s
		GROUP BY user_id
		ORDER BY total_calls DESC
		LIMIT 10`, s.historyTable)

	if rows, err := s.db.QueryContext(ctx, userQuery); err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat models.UserStat
			if err := rows.Scan(&stat.UserID, &stat.TotalCalls); err == nil {
				stats.TopUsers = append(stats.TopUsers, stat)
			}
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
