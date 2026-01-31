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

package inthubhistory

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

// IntHubHistoryService manages integration hub history tracking
type IntHubHistoryService struct {
	db            *sql.DB
	log           logger.Log
	historyTable  string
	actionTable   string
	buffer        []*models.IntHubHistory
	bufferMutex   sync.Mutex
	bufferSize    int
	flushInterval time.Duration
	stopChan      chan struct{}
}

// Global instance
var (
	globalIntHubHistoryService *IntHubHistoryService
	globalServiceMutex         sync.RWMutex
)

// GetGlobalIntHubHistoryService returns the global service instance
func GetGlobalIntHubHistoryService() *IntHubHistoryService {
	globalServiceMutex.RLock()
	defer globalServiceMutex.RUnlock()
	return globalIntHubHistoryService
}

// SetGlobalIntHubHistoryService sets the global service instance
func SetGlobalIntHubHistoryService(service *IntHubHistoryService) {
	globalServiceMutex.Lock()
	defer globalServiceMutex.Unlock()
	globalIntHubHistoryService = service
}

// ServiceConfig holds configuration for the service
type ServiceConfig struct {
	HistoryTable  string
	ActionTable   string
	BufferSize    int
	FlushInterval int // seconds
}

// DefaultServiceConfig returns default configuration
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		HistoryTable:  "iac_int_hub_history",
		ActionTable:   "iac_int_hub_action_history",
		BufferSize:    100,
		FlushInterval: 5,
	}
}

// NewIntHubHistoryService creates a new integration hub history service
func NewIntHubHistoryService(cfg *ServiceConfig) (*IntHubHistoryService, error) {
	if cfg == nil {
		cfg = DefaultServiceConfig()
	}

	// Get SQL database connection
	if dbconn.DB == nil {
		return nil, fmt.Errorf("SQL database connection not available")
	}

	service := &IntHubHistoryService{
		db:            dbconn.DB,
		log:           logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "IntHubHistoryService"},
		historyTable:  cfg.HistoryTable,
		actionTable:   cfg.ActionTable,
		buffer:        make([]*models.IntHubHistory, 0, cfg.BufferSize),
		bufferSize:    cfg.BufferSize,
		flushInterval: time.Duration(cfg.FlushInterval) * time.Second,
		stopChan:      make(chan struct{}),
	}

	// Start background tasks
	go service.backgroundTasks()

	service.log.Info("Integration Hub History Service initialized")
	return service, nil
}

// backgroundTasks handles periodic buffer flushing
func (s *IntHubHistoryService) backgroundTasks() {
	flushTicker := time.NewTicker(s.flushInterval)
	defer flushTicker.Stop()

	for {
		select {
		case <-flushTicker.C:
			s.flushBuffer()
		case <-s.stopChan:
			s.flushBuffer() // Final flush before stopping
			return
		}
	}
}

// Stop stops the background flush goroutine
func (s *IntHubHistoryService) Stop() {
	close(s.stopChan)
}

// detectDatabaseType tries to detect the database type
func (s *IntHubHistoryService) detectDatabaseType() string {
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

// getPlaceholder returns the appropriate placeholder for the database type
func (s *IntHubHistoryService) getPlaceholder(dbType string, index int) string {
	if dbType == "postgres" {
		return fmt.Sprintf("$%d", index)
	}
	return "?"
}

// RecordTransaction creates a new history record and returns its ID
func (s *IntHubHistoryService) RecordTransaction(ctx context.Context, history *models.IntHubHistory) (string, error) {
	startTime := time.Now()
	defer func() {
		s.log.Debug(fmt.Sprintf("RecordTransaction completed in %v", time.Since(startTime)))
	}()

	if history.ID == "" {
		history.ID = uuid.New().String()
	}
	s.log.Info(fmt.Sprintf("IntHubHistoryService: Recording transaction %s for hub %s, endpoint %s", history.ID, history.HubName, history.EndpointName))

	if s.db == nil {
		s.log.Error("IntHubHistoryService: Database connection is nil")
		return "", fmt.Errorf("database connection is nil")
	}
	if history.StartTime.IsZero() {
		history.StartTime = time.Now()
	}
	if history.Status == "" {
		history.Status = models.IntHubHistoryStatusPending
	}
	history.Active = true
	history.CreatedOn = time.Now()
	history.ModifiedOn = time.Now()
	history.RowVersionStamp = 1

	dbType := s.detectDatabaseType()
	query := fmt.Sprintf(`
		INSERT INTO %s (id, hub_id, hub_name, direction, protocol, protocol_group_id, protocol_group_name,
			endpoint_id, endpoint_name, topic, path, method, source, payload, payload_size, message_type, action, status,
			error_message, mapped_data, response, response_status, start_time, end_time, duration_ms,
			instance_id, instance_name, user_id, client_id, metadata, active, referenceid, createdby,
			createdon, modifiedby, modifiedon, rowversionstamp)
		VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
	`, s.historyTable,
		s.getPlaceholder(dbType, 1), s.getPlaceholder(dbType, 2), s.getPlaceholder(dbType, 3),
		s.getPlaceholder(dbType, 4), s.getPlaceholder(dbType, 5), s.getPlaceholder(dbType, 6),
		s.getPlaceholder(dbType, 7), s.getPlaceholder(dbType, 8), s.getPlaceholder(dbType, 9),
		s.getPlaceholder(dbType, 10), s.getPlaceholder(dbType, 11), s.getPlaceholder(dbType, 12),
		s.getPlaceholder(dbType, 13), s.getPlaceholder(dbType, 14), s.getPlaceholder(dbType, 15),
		s.getPlaceholder(dbType, 16), s.getPlaceholder(dbType, 17), s.getPlaceholder(dbType, 18),
		s.getPlaceholder(dbType, 19), s.getPlaceholder(dbType, 20), s.getPlaceholder(dbType, 21),
		s.getPlaceholder(dbType, 22), s.getPlaceholder(dbType, 23), s.getPlaceholder(dbType, 24),
		s.getPlaceholder(dbType, 25), s.getPlaceholder(dbType, 26), s.getPlaceholder(dbType, 27),
		s.getPlaceholder(dbType, 28), s.getPlaceholder(dbType, 29), s.getPlaceholder(dbType, 30),
		s.getPlaceholder(dbType, 31), s.getPlaceholder(dbType, 32), s.getPlaceholder(dbType, 33),
		s.getPlaceholder(dbType, 34), s.getPlaceholder(dbType, 35), s.getPlaceholder(dbType, 36), s.getPlaceholder(dbType, 37))

	_, err := s.db.ExecContext(ctx, query,
		history.ID,
		history.HubID,
		history.HubName,
		history.Direction,
		history.Protocol,
		history.ProtocolGroupID,
		history.ProtocolGroupName,
		history.EndpointID,
		history.EndpointName,
		history.Topic,
		history.Path,
		history.Method,
		history.Source,
		history.Payload,
		history.PayloadSize,
		history.MessageType,
		history.Action,
		history.Status,
		history.ErrorMessage,
		history.MappedData,
		history.Response,
		history.ResponseStatus,
		history.StartTime,
		history.EndTime,
		history.DurationMs,
		history.InstanceID,
		history.InstanceName,
		history.UserID,
		history.ClientID,
		toJSON(history.Metadata),
		history.Active,
		history.ReferenceID,
		history.CreatedBy,
		history.CreatedOn,
		history.ModifiedBy,
		history.ModifiedOn,
		history.RowVersionStamp,
	)

	if err != nil {
		s.log.Error(fmt.Sprintf("IntHubHistoryService: Failed to insert history record: %v", err))
		return "", err
	}

	return history.ID, nil
}

// RecordAction records an action within a transaction
func (s *IntHubHistoryService) RecordAction(ctx context.Context, action *models.IntHubActionHistory) error {
	startTime := time.Now()
	defer func() {
		s.log.Debug(fmt.Sprintf("RecordAction completed in %v", time.Since(startTime)))
	}()

	if action.ID == "" {
		action.ID = uuid.New().String()
	}
	if action.StartTime.IsZero() {
		action.StartTime = time.Now()
	}
	if action.Status == "" {
		action.Status = models.IntHubActionStatusPending
	}
	action.Active = true
	action.CreatedOn = time.Now()
	action.ModifiedOn = time.Now()
	action.RowVersionStamp = 1

	dbType := s.detectDatabaseType()
	query := fmt.Sprintf(`
		INSERT INTO %s (id, int_hub_his_id, action_sequence, action_type, action_name, action_config,
			input_data, output_data, status, error_message, start_time, end_time, duration_ms,
			metadata, active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp)
		VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
	`, s.actionTable,
		s.getPlaceholder(dbType, 1), s.getPlaceholder(dbType, 2), s.getPlaceholder(dbType, 3),
		s.getPlaceholder(dbType, 4), s.getPlaceholder(dbType, 5), s.getPlaceholder(dbType, 6),
		s.getPlaceholder(dbType, 7), s.getPlaceholder(dbType, 8), s.getPlaceholder(dbType, 9),
		s.getPlaceholder(dbType, 10), s.getPlaceholder(dbType, 11), s.getPlaceholder(dbType, 12),
		s.getPlaceholder(dbType, 13), s.getPlaceholder(dbType, 14), s.getPlaceholder(dbType, 15),
		s.getPlaceholder(dbType, 16), s.getPlaceholder(dbType, 17), s.getPlaceholder(dbType, 18),
		s.getPlaceholder(dbType, 19), s.getPlaceholder(dbType, 20), s.getPlaceholder(dbType, 21))

	_, err := s.db.ExecContext(ctx, query,
		action.ID,
		action.IntHubHisID,
		action.ActionSequence,
		action.ActionType,
		action.ActionName,
		action.ActionConfig,
		action.InputData,
		action.OutputData,
		action.Status,
		action.ErrorMessage,
		action.StartTime,
		action.EndTime,
		action.DurationMs,
		toJSON(action.Metadata),
		action.Active,
		action.ReferenceID,
		action.CreatedBy,
		action.CreatedOn,
		action.ModifiedBy,
		action.ModifiedOn,
		action.RowVersionStamp,
	)

	if err != nil {
		s.log.Error(fmt.Sprintf("Failed to record action: %v", err))
		return err
	}

	return nil
}

// CompleteTransaction marks a transaction as completed
func (s *IntHubHistoryService) CompleteTransaction(ctx context.Context, id string, response string, mappedData string, responseStatus int) error {
	now := time.Now()
	dbType := s.detectDatabaseType()

	// First get the start time to calculate duration
	var startTime time.Time
	selectQuery := fmt.Sprintf("SELECT start_time FROM %s WHERE id = %s", s.historyTable, s.getPlaceholder(dbType, 1))
	err := s.db.QueryRowContext(ctx, selectQuery, id).Scan(&startTime)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	durationMs := now.Sub(startTime).Milliseconds()

	query := fmt.Sprintf(`
		UPDATE %s SET status = %s, response = %s, mapped_data = %s, response_status = %s,
			end_time = %s, duration_ms = %s, modifiedon = %s, rowversionstamp = rowversionstamp + 1
		WHERE id = %s
	`, s.historyTable,
		s.getPlaceholder(dbType, 1), s.getPlaceholder(dbType, 2), s.getPlaceholder(dbType, 3),
		s.getPlaceholder(dbType, 4), s.getPlaceholder(dbType, 5), s.getPlaceholder(dbType, 6),
		s.getPlaceholder(dbType, 7), s.getPlaceholder(dbType, 8))

	_, err = s.db.ExecContext(ctx, query,
		models.IntHubHistoryStatusCompleted,
		response,
		mappedData,
		responseStatus,
		now,
		durationMs,
		now,
		id,
	)

	if err != nil {
		s.log.Error(fmt.Sprintf("Failed to complete transaction: %v", err))
		return err
	}

	return nil
}

// FailTransaction marks a transaction as failed
func (s *IntHubHistoryService) FailTransaction(ctx context.Context, id string, errorMessage string) error {
	now := time.Now()
	dbType := s.detectDatabaseType()

	// First get the start time to calculate duration
	var startTime time.Time
	selectQuery := fmt.Sprintf("SELECT start_time FROM %s WHERE id = %s", s.historyTable, s.getPlaceholder(dbType, 1))
	err := s.db.QueryRowContext(ctx, selectQuery, id).Scan(&startTime)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	durationMs := now.Sub(startTime).Milliseconds()

	query := fmt.Sprintf(`
		UPDATE %s SET status = %s, error_message = %s, end_time = %s, duration_ms = %s,
			modifiedon = %s, rowversionstamp = rowversionstamp + 1
		WHERE id = %s
	`, s.historyTable,
		s.getPlaceholder(dbType, 1), s.getPlaceholder(dbType, 2), s.getPlaceholder(dbType, 3),
		s.getPlaceholder(dbType, 4), s.getPlaceholder(dbType, 5), s.getPlaceholder(dbType, 6))

	_, err = s.db.ExecContext(ctx, query,
		models.IntHubHistoryStatusFailed,
		errorMessage,
		now,
		durationMs,
		now,
		id,
	)

	if err != nil {
		s.log.Error(fmt.Sprintf("Failed to fail transaction: %v", err))
		return err
	}

	return nil
}

// ListHistory returns paginated history records with optional filters
func (s *IntHubHistoryService) ListHistory(ctx context.Context, params models.ListIntHubHistoryParams) (*models.ListIntHubHistoryResponse, error) {
	startTime := time.Now()
	defer func() {
		s.log.Debug(fmt.Sprintf("ListHistory completed in %v", time.Since(startTime)))
	}()

	dbType := s.detectDatabaseType()
	var conditions []string
	var args []interface{}
	placeholderIndex := 1

	// Always filter by active
	conditions = append(conditions, "active = true")

	if params.HubID != "" {
		conditions = append(conditions, fmt.Sprintf("hub_id = %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.HubID)
		placeholderIndex++
	}
	if params.Direction != "" {
		conditions = append(conditions, fmt.Sprintf("direction = %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.Direction)
		placeholderIndex++
	}
	if params.Protocol != "" {
		conditions = append(conditions, fmt.Sprintf("protocol = %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.Protocol)
		placeholderIndex++
	}
	if params.EndpointID != "" {
		conditions = append(conditions, fmt.Sprintf("endpoint_id = %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.EndpointID)
		placeholderIndex++
	}
	if params.Topic != "" {
		conditions = append(conditions, fmt.Sprintf("topic = %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.Topic)
		placeholderIndex++
	}
	if params.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.Status)
		placeholderIndex++
	}
	if !params.StartTime.IsZero() {
		conditions = append(conditions, fmt.Sprintf("start_time >= %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.StartTime)
		placeholderIndex++
	}
	if !params.EndTime.IsZero() {
		conditions = append(conditions, fmt.Sprintf("start_time <= %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, params.EndTime)
		placeholderIndex++
	}
	if params.OnlyErrors {
		conditions = append(conditions, "status = 'failed'")
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", s.historyTable, whereClause)
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Pagination
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	offset := (params.Page - 1) * params.PageSize

	// Query for list items
	selectQuery := fmt.Sprintf(`
		SELECT id, hub_id, hub_name, direction, protocol, endpoint_name, topic, status, duration_ms, start_time
		FROM %s %s
		ORDER BY start_time DESC
		LIMIT %s OFFSET %s
	`, s.historyTable, whereClause, s.getPlaceholder(dbType, placeholderIndex), s.getPlaceholder(dbType, placeholderIndex+1))

	queryArgs := append(args, params.PageSize, offset)
	rows, err := s.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.IntHubHistoryListItem
	for rows.Next() {
		var item models.IntHubHistoryListItem
		var topic, endpointName sql.NullString
		var durationMs sql.NullInt64

		if err := rows.Scan(&item.ID, &item.HubID, &item.HubName, &item.Direction, &item.Protocol,
			&endpointName, &topic, &item.Status, &durationMs, &item.StartTime); err != nil {
			return nil, err
		}

		item.EndpointName = endpointName.String
		item.Topic = topic.String
		item.DurationMs = durationMs.Int64
		history = append(history, item)
	}

	pages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))

	return &models.ListIntHubHistoryResponse{
		History:  history,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    pages,
	}, nil
}

// GetHistory returns a single history record by ID
func (s *IntHubHistoryService) GetHistory(ctx context.Context, id string) (*models.IntHubHistory, error) {
	startTime := time.Now()
	defer func() {
		s.log.Debug(fmt.Sprintf("GetHistory completed in %v", time.Since(startTime)))
	}()

	dbType := s.detectDatabaseType()
	query := fmt.Sprintf(`
		SELECT id, hub_id, hub_name, direction, protocol, protocol_group_id, protocol_group_name,
			endpoint_id, endpoint_name, topic, path, method, source, payload, payload_size, message_type, action,
			status, error_message, mapped_data, response, response_status, start_time, end_time,
			duration_ms, instance_id, instance_name, user_id, client_id, metadata, active,
			referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM %s WHERE id = %s AND active = true
	`, s.historyTable, s.getPlaceholder(dbType, 1))

	row := s.db.QueryRowContext(ctx, query, id)

	var history models.IntHubHistory
	var endTime sql.NullTime
	var metadata sql.NullString

	err := row.Scan(
		&history.ID,
		&history.HubID,
		&history.HubName,
		&history.Direction,
		&history.Protocol,
		&history.ProtocolGroupID,
		&history.ProtocolGroupName,
		&history.EndpointID,
		&history.EndpointName,
		&history.Topic,
		&history.Path,
		&history.Method,
		&history.Source,
		&history.Payload,
		&history.PayloadSize,
		&history.MessageType,
		&history.Action,
		&history.Status,
		&history.ErrorMessage,
		&history.MappedData,
		&history.Response,
		&history.ResponseStatus,
		&history.StartTime,
		&endTime,
		&history.DurationMs,
		&history.InstanceID,
		&history.InstanceName,
		&history.UserID,
		&history.ClientID,
		&metadata,
		&history.Active,
		&history.ReferenceID,
		&history.CreatedBy,
		&history.CreatedOn,
		&history.ModifiedBy,
		&history.ModifiedOn,
		&history.RowVersionStamp,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("record not found")
	}
	if err != nil {
		return nil, err
	}

	if endTime.Valid {
		history.EndTime = &endTime.Time
	}
	if metadata.String != "" {
		json.Unmarshal([]byte(metadata.String), &history.Metadata)
	}

	return &history, nil
}

// GetActionsByTransaction returns all actions for a transaction
func (s *IntHubHistoryService) GetActionsByTransaction(ctx context.Context, transactionID string) ([]models.IntHubActionHistory, error) {
	startTime := time.Now()
	defer func() {
		s.log.Debug(fmt.Sprintf("GetActionsByTransaction completed in %v", time.Since(startTime)))
	}()

	dbType := s.detectDatabaseType()
	query := fmt.Sprintf(`
		SELECT id, int_hub_his_id, action_sequence, action_type, action_name, action_config,
			input_data, output_data, status, error_message, start_time, end_time, duration_ms,
			metadata, active, referenceid, createdby, createdon, modifiedby, modifiedon, rowversionstamp
		FROM %s WHERE int_hub_his_id = %s AND active = true
		ORDER BY action_sequence ASC
	`, s.actionTable, s.getPlaceholder(dbType, 1))

	rows, err := s.db.QueryContext(ctx, query, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []models.IntHubActionHistory
	for rows.Next() {
		var action models.IntHubActionHistory
		var endTime sql.NullTime
		var metadata sql.NullString

		if err := rows.Scan(
			&action.ID,
			&action.IntHubHisID,
			&action.ActionSequence,
			&action.ActionType,
			&action.ActionName,
			&action.ActionConfig,
			&action.InputData,
			&action.OutputData,
			&action.Status,
			&action.ErrorMessage,
			&action.StartTime,
			&endTime,
			&action.DurationMs,
			&metadata,
			&action.Active,
			&action.ReferenceID,
			&action.CreatedBy,
			&action.CreatedOn,
			&action.ModifiedBy,
			&action.ModifiedOn,
			&action.RowVersionStamp,
		); err != nil {
			return nil, err
		}

		if endTime.Valid {
			action.EndTime = &endTime.Time
		}
		if metadata.String != "" {
			json.Unmarshal([]byte(metadata.String), &action.Metadata)
		}

		actions = append(actions, action)
	}

	return actions, nil
}

// DeleteHistory soft-deletes a history record
func (s *IntHubHistoryService) DeleteHistory(ctx context.Context, id string) error {
	dbType := s.detectDatabaseType()
	now := time.Now()

	query := fmt.Sprintf(`
		UPDATE %s SET active = false, modifiedon = %s, rowversionstamp = rowversionstamp + 1
		WHERE id = %s
	`, s.historyTable, s.getPlaceholder(dbType, 1), s.getPlaceholder(dbType, 2))

	result, err := s.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("record not found")
	}

	return nil
}

// GetStatistics returns statistics for integration hub history
func (s *IntHubHistoryService) GetStatistics(ctx context.Context, hubID string, startTime, endTime time.Time) (*models.IntHubHistoryStats, error) {
	dbType := s.detectDatabaseType()
	stats := &models.IntHubHistoryStats{
		ByProtocol:   make(map[string]int64),
		ByDirection:  make(map[string]int64),
		ByStatus:     make(map[string]int64),
		TopEndpoints: make([]models.IntHubEndpointStat, 0),
		TopTopics:    make([]models.IntHubTopicStat, 0),
	}

	// Build base WHERE clause
	conditions := []string{"active = true"}
	args := []interface{}{}
	placeholderIndex := 1

	if hubID != "" {
		conditions = append(conditions, fmt.Sprintf("hub_id = %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, hubID)
		placeholderIndex++
	}
	if !startTime.IsZero() {
		conditions = append(conditions, fmt.Sprintf("start_time >= %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, startTime)
		placeholderIndex++
	}
	if !endTime.IsZero() {
		conditions = append(conditions, fmt.Sprintf("start_time <= %s", s.getPlaceholder(dbType, placeholderIndex)))
		args = append(args, endTime)
		placeholderIndex++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", s.historyTable, whereClause)
	s.db.QueryRowContext(ctx, countQuery, args...).Scan(&stats.TotalTransactions)

	// Count by status
	statusQuery := fmt.Sprintf(`
		SELECT status, COUNT(*) FROM %s %s GROUP BY status
	`, s.historyTable, whereClause)
	rows, err := s.db.QueryContext(ctx, statusQuery, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int64
			rows.Scan(&status, &count)
			stats.ByStatus[status] = count
			switch status {
			case "completed":
				stats.CompletedCount = count
			case "failed":
				stats.FailedCount = count
			case "pending":
				stats.PendingCount = count
			case "processing":
				stats.ProcessingCount = count
			}
		}
	}

	// Count by protocol
	protocolQuery := fmt.Sprintf(`
		SELECT protocol, COUNT(*) FROM %s %s GROUP BY protocol
	`, s.historyTable, whereClause)
	rows, err = s.db.QueryContext(ctx, protocolQuery, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var protocol string
			var count int64
			rows.Scan(&protocol, &count)
			stats.ByProtocol[protocol] = count
		}
	}

	// Count by direction
	directionQuery := fmt.Sprintf(`
		SELECT direction, COUNT(*) FROM %s %s GROUP BY direction
	`, s.historyTable, whereClause)
	rows, err = s.db.QueryContext(ctx, directionQuery, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var direction string
			var count int64
			rows.Scan(&direction, &count)
			stats.ByDirection[direction] = count
		}
	}

	// Duration statistics
	durationQuery := fmt.Sprintf(`
		SELECT COALESCE(AVG(duration_ms), 0), COALESCE(MIN(duration_ms), 0), COALESCE(MAX(duration_ms), 0)
		FROM %s %s AND status = 'completed'
	`, s.historyTable, whereClause)
	s.db.QueryRowContext(ctx, durationQuery, args...).Scan(&stats.AverageDurationMs, &stats.MinDurationMs, &stats.MaxDurationMs)

	// Top endpoints
	endpointQuery := fmt.Sprintf(`
		SELECT endpoint_id, endpoint_name, protocol, COUNT(*) as total,
			COALESCE(AVG(duration_ms), 0) as avg_duration,
			COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 0) as error_rate
		FROM %s %s AND endpoint_id IS NOT NULL AND endpoint_id != ''
		GROUP BY endpoint_id, endpoint_name, protocol
		ORDER BY total DESC
		LIMIT 10
	`, s.historyTable, whereClause)
	rows, err = s.db.QueryContext(ctx, endpointQuery, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat models.IntHubEndpointStat
			rows.Scan(&stat.EndpointID, &stat.EndpointName, &stat.Protocol, &stat.TotalCount, &stat.AvgDuration, &stat.ErrorRate)
			stats.TopEndpoints = append(stats.TopEndpoints, stat)
		}
	}

	// Top topics
	topicQuery := fmt.Sprintf(`
		SELECT topic, protocol, COUNT(*) as total,
			COALESCE(AVG(duration_ms), 0) as avg_duration,
			COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 0) as error_rate
		FROM %s %s AND topic IS NOT NULL AND topic != ''
		GROUP BY topic, protocol
		ORDER BY total DESC
		LIMIT 10
	`, s.historyTable, whereClause)
	rows, err = s.db.QueryContext(ctx, topicQuery, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat models.IntHubTopicStat
			rows.Scan(&stat.Topic, &stat.Protocol, &stat.TotalCount, &stat.AvgDuration, &stat.ErrorRate)
			stats.TopTopics = append(stats.TopTopics, stat)
		}
	}

	// Time range
	if !startTime.IsZero() || !endTime.IsZero() {
		stats.TimeRange = &models.IntHubTimeRangeStats{
			StartTime: startTime,
			EndTime:   endTime,
		}
	}

	return stats, nil
}

// Cleanup deletes records older than the specified date
func (s *IntHubHistoryService) Cleanup(ctx context.Context, before time.Time) (int64, error) {
	dbType := s.detectDatabaseType()

	// First delete actions for old history records
	actionDeleteQuery := fmt.Sprintf(`
		DELETE FROM %s WHERE int_hub_his_id IN (
			SELECT id FROM %s WHERE start_time < %s
		)
	`, s.actionTable, s.historyTable, s.getPlaceholder(dbType, 1))
	s.db.ExecContext(ctx, actionDeleteQuery, before)

	// Then delete the history records
	historyDeleteQuery := fmt.Sprintf("DELETE FROM %s WHERE start_time < %s", s.historyTable, s.getPlaceholder(dbType, 1))
	result, err := s.db.ExecContext(ctx, historyDeleteQuery, before)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// flushBuffer saves all buffered records to database
func (s *IntHubHistoryService) flushBuffer() {
	s.bufferMutex.Lock()
	if len(s.buffer) == 0 {
		s.bufferMutex.Unlock()
		return
	}

	records := s.buffer
	s.buffer = make([]*models.IntHubHistory, 0, s.bufferSize)
	s.bufferMutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, record := range records {
		if _, err := s.RecordTransaction(ctx, record); err != nil {
			s.log.Error(fmt.Sprintf("Failed to save integration hub history: %v", err))
		}
	}
}

// Helper function for JSON serialization
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
