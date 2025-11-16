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

package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/logger"
)

// MySQLMonitor monitors MySQL database health and performance
type MySQLMonitor struct {
	db              *sql.DB
	config          *dbconn.DBConfig
	checkInterval   time.Duration
	stopChan        chan struct{}
	running         bool
	mu              sync.RWMutex
	lastCheckTime   time.Time
	consecutiveFails int
	iLog            logger.Log
}

// MySQLStatus represents MySQL server status
type MySQLStatus struct {
	Version         string
	Uptime          int64
	Threads         int
	Questions       int64
	SlowQueries     int64
	Opens           int64
	FlushTables     int64
	OpenTables      int64
	QueriesPerSec   float64
	BytesReceived   int64
	BytesSent       int64
	Connections     int64
	AbortedClients  int64
	AbortedConnects int64
	MaxUsedConns    int
}

// NewMySQLMonitor creates a new MySQL monitor
func NewMySQLMonitor(db *sql.DB, config *dbconn.DBConfig) *MySQLMonitor {
	return &MySQLMonitor{
		db:            db,
		config:        config,
		checkInterval: 30 * time.Second,
		stopChan:      make(chan struct{}),
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MySQLMonitor"},
	}
}

// Start starts the monitoring
func (m *MySQLMonitor) Start(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	m.iLog.Info("Starting MySQL monitoring")

	go func() {
		ticker := time.NewTicker(m.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.performHealthCheck(ctx)
			case <-m.stopChan:
				m.iLog.Info("Stopping MySQL monitoring")
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the monitoring
func (m *MySQLMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		close(m.stopChan)
		m.running = false
	}
}

// performHealthCheck performs a health check
func (m *MySQLMonitor) performHealthCheck(ctx context.Context) {
	m.mu.Lock()
	m.lastCheckTime = time.Now()
	m.mu.Unlock()

	// Ping database
	if err := m.db.Ping(); err != nil {
		m.handleFailedCheck(err)
		return
	}

	// Reset consecutive fails on success
	m.mu.Lock()
	m.consecutiveFails = 0
	m.mu.Unlock()

	// Get status (optional, can be expensive)
	// status, err := m.GetStatus(ctx)
	// if err != nil {
	// 	m.iLog.Warn(fmt.Sprintf("Failed to get MySQL status: %v", err))
	// }
}

// handleFailedCheck handles a failed health check
func (m *MySQLMonitor) handleFailedCheck(err error) {
	m.mu.Lock()
	m.consecutiveFails++
	fails := m.consecutiveFails
	m.mu.Unlock()

	m.iLog.Error(fmt.Sprintf("MySQL health check failed (consecutive: %d): %v", fails, err))

	// Attempt reconnection after 3 consecutive failures
	if fails >= 3 {
		m.iLog.Warn("Attempting to reconnect to MySQL database...")
		if err := m.db.Ping(); err == nil {
			m.iLog.Info("Successfully reconnected to MySQL database")
			m.mu.Lock()
			m.consecutiveFails = 0
			m.mu.Unlock()
		}
	}
}

// GetStatus retrieves MySQL server status
func (m *MySQLMonitor) GetStatus(ctx context.Context) (*MySQLStatus, error) {
	status := &MySQLStatus{}

	// Get version
	err := m.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&status.Version)
	if err != nil {
		return nil, err
	}

	// Get status variables
	rows, err := m.db.QueryContext(ctx, "SHOW GLOBAL STATUS")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statusVars := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		statusVars[key] = value
	}

	// Parse status variables
	fmt.Sscanf(statusVars["Uptime"], "%d", &status.Uptime)
	fmt.Sscanf(statusVars["Threads_running"], "%d", &status.Threads)
	fmt.Sscanf(statusVars["Questions"], "%d", &status.Questions)
	fmt.Sscanf(statusVars["Slow_queries"], "%d", &status.SlowQueries)
	fmt.Sscanf(statusVars["Opens"], "%d", &status.Opens)
	fmt.Sscanf(statusVars["Flush_commands"], "%d", &status.FlushTables)
	fmt.Sscanf(statusVars["Open_tables"], "%d", &status.OpenTables)
	fmt.Sscanf(statusVars["Bytes_received"], "%d", &status.BytesReceived)
	fmt.Sscanf(statusVars["Bytes_sent"], "%d", &status.BytesSent)
	fmt.Sscanf(statusVars["Connections"], "%d", &status.Connections)
	fmt.Sscanf(statusVars["Aborted_clients"], "%d", &status.AbortedClients)
	fmt.Sscanf(statusVars["Aborted_connects"], "%d", &status.AbortedConnects)
	fmt.Sscanf(statusVars["Max_used_connections"], "%d", &status.MaxUsedConns)

	// Calculate queries per second
	if status.Uptime > 0 {
		status.QueriesPerSec = float64(status.Questions) / float64(status.Uptime)
	}

	return status, nil
}

// GetProcessList retrieves the current process list
func (m *MySQLMonitor) GetProcessList(ctx context.Context) ([]ProcessInfo, error) {
	query := "SHOW FULL PROCESSLIST"

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	processes := make([]ProcessInfo, 0)

	for rows.Next() {
		var (
			id      int64
			user    string
			host    string
			db      sql.NullString
			command string
			timeVal int
			state   sql.NullString
			info    sql.NullString
		)

		if err := rows.Scan(&id, &user, &host, &db, &command, &timeVal, &state, &info); err != nil {
			return nil, err
		}

		process := ProcessInfo{
			ID:      id,
			User:    user,
			Host:    host,
			Command: command,
			Time:    timeVal,
		}

		if db.Valid {
			process.Database = db.String
		}
		if state.Valid {
			process.State = state.String
		}
		if info.Valid {
			process.Info = info.String
		}

		processes = append(processes, process)
	}

	return processes, rows.Err()
}

// KillProcess kills a process by ID
func (m *MySQLMonitor) KillProcess(ctx context.Context, processID int64) error {
	query := fmt.Sprintf("KILL %d", processID)
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// GetSlowQueries retrieves slow queries (requires slow query log enabled)
func (m *MySQLMonitor) GetSlowQueries(ctx context.Context, limit int) ([]SlowQuery, error) {
	// This is a placeholder - actual implementation would require slow query log analysis
	return nil, dbconn.ErrFeatureNotSupported
}

// GetTableSizes retrieves table sizes for the current database
func (m *MySQLMonitor) GetTableSizes(ctx context.Context) ([]TableSize, error) {
	query := `
		SELECT
			TABLE_NAME,
			TABLE_ROWS,
			DATA_LENGTH,
			INDEX_LENGTH,
			DATA_LENGTH + INDEX_LENGTH AS total_size
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		ORDER BY total_size DESC
	`

	rows, err := m.db.QueryContext(ctx, query, m.config.Database)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]TableSize, 0)

	for rows.Next() {
		var table TableSize
		if err := rows.Scan(&table.Name, &table.Rows, &table.DataSize, &table.IndexSize, &table.TotalSize); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

// GetLastCheckTime returns the last health check time
func (m *MySQLMonitor) GetLastCheckTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastCheckTime
}

// GetConsecutiveFailures returns the number of consecutive failures
func (m *MySQLMonitor) GetConsecutiveFailures() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.consecutiveFails
}

// IsHealthy checks if the database is healthy
func (m *MySQLMonitor) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.consecutiveFails == 0
}

// Supporting types

// ProcessInfo represents a MySQL process
type ProcessInfo struct {
	ID       int64
	User     string
	Host     string
	Database string
	Command  string
	Time     int
	State    string
	Info     string
}

// SlowQuery represents a slow query
type SlowQuery struct {
	Query        string
	QueryTime    float64
	LockTime     float64
	RowsSent     int64
	RowsExamined int64
	Timestamp    time.Time
}

// TableSize represents table size information
type TableSize struct {
	Name      string
	Rows      int64
	DataSize  int64
	IndexSize int64
	TotalSize int64
}
