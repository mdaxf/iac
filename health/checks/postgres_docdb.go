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

package checks

import (
	"context"
	"fmt"
	"time"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
)

// PostgresDocDBHealthCheck performs health check for PostgreSQL document database
type PostgresDocDBHealthCheck struct {
	db       documents.DocumentDB
	name     string
	timeout  time.Duration
	iLog     logger.Log
}

// NewPostgresDocDBHealthCheck creates a new PostgreSQL document database health check
func NewPostgresDocDBHealthCheck(db documents.DocumentDB, name string) *PostgresDocDBHealthCheck {
	return &PostgresDocDBHealthCheck{
		db:      db,
		name:    name,
		timeout: 5 * time.Second,
		iLog:    logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "PostgresDocDBHealthCheck"},
	}
}

// SetTimeout sets the health check timeout
func (hc *PostgresDocDBHealthCheck) SetTimeout(timeout time.Duration) {
	hc.timeout = timeout
}

// Check performs the health check
func (hc *PostgresDocDBHealthCheck) Check() HealthCheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	result := HealthCheckResult{
		Name:      hc.name,
		Type:      "postgres_document_db",
		Status:    StatusHealthy,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// Check connection
	if !hc.db.IsConnected() {
		result.Status = StatusUnhealthy
		result.Message = "Database not connected"
		return result
	}

	// Ping the database
	startPing := time.Now()
	if err := hc.db.Ping(ctx); err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Ping failed: %v", err)
		result.Details["error"] = err.Error()
		return result
	}
	pingDuration := time.Since(startPing)

	result.Details["ping_duration_ms"] = pingDuration.Milliseconds()
	result.Details["db_type"] = hc.db.GetType()

	// Get database statistics
	stats, err := hc.db.Stats(ctx)
	if err != nil {
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("Stats retrieval failed: %v", err)
		result.Details["stats_error"] = err.Error()
	} else {
		result.Details["collections"] = stats.Collections
		result.Details["documents"] = stats.Documents
		result.Details["data_size"] = stats.DataSize
		result.Details["index_size"] = stats.IndexSize
		result.Details["storage_size"] = stats.StorageSize
	}

	// List collections (as a basic operation test)
	collections, err := hc.db.ListCollections(ctx)
	if err != nil {
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("List collections failed: %v", err)
		result.Details["collections_error"] = err.Error()
	} else {
		result.Details["collection_count"] = len(collections)
		result.Details["collection_names"] = collections
	}

	// Set success message if still healthy
	if result.Status == StatusHealthy {
		result.Message = "PostgreSQL document database is healthy"
	}

	result.Duration = time.Since(result.Timestamp)

	return result
}

// DocumentDBHealthMonitor monitors document database health
type DocumentDBHealthMonitor struct {
	db               documents.DocumentDB
	name             string
	checkInterval    time.Duration
	unhealthyCount   int
	maxUnhealthy     int
	stopChan         chan struct{}
	resultChan       chan HealthCheckResult
	running          bool
	iLog             logger.Log
}

// NewDocumentDBHealthMonitor creates a new document database health monitor
func NewDocumentDBHealthMonitor(db documents.DocumentDB, name string) *DocumentDBHealthMonitor {
	return &DocumentDBHealthMonitor{
		db:            db,
		name:          name,
		checkInterval: 30 * time.Second,
		maxUnhealthy:  3,
		stopChan:      make(chan struct{}),
		resultChan:    make(chan HealthCheckResult, 10),
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DocumentDBHealthMonitor"},
	}
}

// SetCheckInterval sets the interval between health checks
func (m *DocumentDBHealthMonitor) SetCheckInterval(interval time.Duration) {
	m.checkInterval = interval
}

// SetMaxUnhealthy sets the maximum number of consecutive unhealthy checks
func (m *DocumentDBHealthMonitor) SetMaxUnhealthy(max int) {
	m.maxUnhealthy = max
}

// Start starts the health monitoring
func (m *DocumentDBHealthMonitor) Start() {
	if m.running {
		return
	}

	m.running = true
	m.iLog.Info(fmt.Sprintf("Starting health monitoring for document database: %s", m.name))

	go m.monitorLoop()
}

// Stop stops the health monitoring
func (m *DocumentDBHealthMonitor) Stop() {
	if !m.running {
		return
	}

	m.iLog.Info(fmt.Sprintf("Stopping health monitoring for document database: %s", m.name))
	close(m.stopChan)
	m.running = false
}

// GetResults returns the results channel
func (m *DocumentDBHealthMonitor) GetResults() <-chan HealthCheckResult {
	return m.resultChan
}

// monitorLoop runs the monitoring loop
func (m *DocumentDBHealthMonitor) monitorLoop() {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	// Perform initial check
	m.performCheck()

	for {
		select {
		case <-ticker.C:
			m.performCheck()
		case <-m.stopChan:
			return
		}
	}
}

// performCheck performs a single health check
func (m *DocumentDBHealthMonitor) performCheck() {
	checker := NewPostgresDocDBHealthCheck(m.db, m.name)
	result := checker.Check()

	// Track consecutive unhealthy checks
	if result.Status == StatusUnhealthy {
		m.unhealthyCount++
		m.iLog.Warn(fmt.Sprintf("Document database '%s' is unhealthy (count: %d/%d): %s",
			m.name, m.unhealthyCount, m.maxUnhealthy, result.Message))

		if m.unhealthyCount >= m.maxUnhealthy {
			m.iLog.Error(fmt.Sprintf("Document database '%s' has been unhealthy for %d consecutive checks",
				m.name, m.unhealthyCount))
			// Could trigger alerts or automated recovery here
		}
	} else {
		if m.unhealthyCount > 0 {
			m.iLog.Info(fmt.Sprintf("Document database '%s' has recovered", m.name))
		}
		m.unhealthyCount = 0
	}

	// Send result to channel
	select {
	case m.resultChan <- result:
	default:
		// Channel full, drop oldest result
		<-m.resultChan
		m.resultChan <- result
	}
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Name      string
	Type      string
	Status    HealthStatus
	Message   string
	Timestamp time.Time
	Duration  time.Duration
	Details   map[string]interface{}
}

// HealthStatus represents the health status
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

// IsHealthy returns true if the status is healthy
func (r HealthCheckResult) IsHealthy() bool {
	return r.Status == StatusHealthy
}

// IsDegraded returns true if the status is degraded
func (r HealthCheckResult) IsDegraded() bool {
	return r.Status == StatusDegraded
}

// IsUnhealthy returns true if the status is unhealthy
func (r HealthCheckResult) IsUnhealthy() bool {
	return r.Status == StatusUnhealthy
}

// DocumentDBHealthCheckManager manages health checks for multiple document databases
type DocumentDBHealthCheckManager struct {
	monitors map[string]*DocumentDBHealthMonitor
	iLog     logger.Log
}

// NewDocumentDBHealthCheckManager creates a new health check manager
func NewDocumentDBHealthCheckManager() *DocumentDBHealthCheckManager {
	return &DocumentDBHealthCheckManager{
		monitors: make(map[string]*DocumentDBHealthMonitor),
		iLog:     logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DocumentDBHealthCheckManager"},
	}
}

// RegisterDatabase registers a database for health monitoring
func (m *DocumentDBHealthCheckManager) RegisterDatabase(name string, db documents.DocumentDB) {
	if _, exists := m.monitors[name]; exists {
		m.iLog.Warn(fmt.Sprintf("Database '%s' is already registered for health monitoring", name))
		return
	}

	monitor := NewDocumentDBHealthMonitor(db, name)
	m.monitors[name] = monitor

	m.iLog.Info(fmt.Sprintf("Registered document database '%s' for health monitoring", name))
}

// UnregisterDatabase unregisters a database from health monitoring
func (m *DocumentDBHealthCheckManager) UnregisterDatabase(name string) {
	if monitor, exists := m.monitors[name]; exists {
		monitor.Stop()
		delete(m.monitors, name)
		m.iLog.Info(fmt.Sprintf("Unregistered document database '%s' from health monitoring", name))
	}
}

// StartAll starts health monitoring for all registered databases
func (m *DocumentDBHealthCheckManager) StartAll() {
	for name, monitor := range m.monitors {
		monitor.Start()
		m.iLog.Info(fmt.Sprintf("Started health monitoring for database: %s", name))
	}
}

// StopAll stops health monitoring for all registered databases
func (m *DocumentDBHealthCheckManager) StopAll() {
	for name, monitor := range m.monitors {
		monitor.Stop()
		m.iLog.Info(fmt.Sprintf("Stopped health monitoring for database: %s", name))
	}
}

// GetMonitor returns a health monitor by name
func (m *DocumentDBHealthCheckManager) GetMonitor(name string) (*DocumentDBHealthMonitor, bool) {
	monitor, exists := m.monitors[name]
	return monitor, exists
}

// CheckAll performs health checks on all registered databases
func (m *DocumentDBHealthCheckManager) CheckAll() map[string]HealthCheckResult {
	results := make(map[string]HealthCheckResult)

	for name, monitor := range m.monitors {
		checker := NewPostgresDocDBHealthCheck(monitor.db, name)
		results[name] = checker.Check()
	}

	return results
}

// GetHealthStatus returns the overall health status
func (m *DocumentDBHealthCheckManager) GetHealthStatus() HealthStatus {
	results := m.CheckAll()

	unhealthyCount := 0
	degradedCount := 0

	for _, result := range results {
		if result.IsUnhealthy() {
			unhealthyCount++
		} else if result.IsDegraded() {
			degradedCount++
		}
	}

	if unhealthyCount > 0 {
		return StatusUnhealthy
	}

	if degradedCount > 0 {
		return StatusDegraded
	}

	return StatusHealthy
}

// Example usage
func ExampleDocumentDBHealthCheck() {
	// Assuming we have a document database instance
	// db := documents.GetDocDBInstance("my-db")

	// Create and perform a single health check
	// checker := NewPostgresDocDBHealthCheck(db, "my-db")
	// result := checker.Check()
	// fmt.Printf("Health check result: %+v\n", result)

	// Create a health monitor for continuous monitoring
	// monitor := NewDocumentDBHealthMonitor(db, "my-db")
	// monitor.SetCheckInterval(30 * time.Second)
	// monitor.Start()

	// // Listen for health check results
	// go func() {
	// 	for result := range monitor.GetResults() {
	// 		fmt.Printf("Health status: %s - %s\n", result.Status, result.Message)
	// 		if result.IsUnhealthy() {
	// 			// Take action...
	// 		}
	// 	}
	// }()

	// Use the manager for multiple databases
	// manager := NewDocumentDBHealthCheckManager()
	// manager.RegisterDatabase("mongodb-primary", mongoDB)
	// manager.RegisterDatabase("postgres-jsonb", postgresDB)
	// manager.StartAll()

	// // Check overall health
	// status := manager.GetHealthStatus()
	// fmt.Printf("Overall health: %s\n", status)
}
