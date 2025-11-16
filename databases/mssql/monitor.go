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

package mssql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/logger"
)

// MSSQLMonitor monitors MSSQL database health and performance
type MSSQLMonitor struct {
	db               *sql.DB
	config           *dbconn.DBConfig
	checkInterval    time.Duration
	stopChan         chan struct{}
	running          bool
	mu               sync.RWMutex
	lastCheckTime    time.Time
	consecutiveFails int
	iLog             logger.Log
}

// MSSQLServerInfo represents SQL Server information
type MSSQLServerInfo struct {
	Version         string
	Edition         string
	ProductLevel    string
	ProductVersion  string
	MachineName     string
	ServerName      string
	InstanceName    string
	IsClustered     bool
	IsHadrEnabled   bool
}

// MSSQLPerformanceStats represents performance statistics
type MSSQLPerformanceStats struct {
	CPUTime              int64
	TotalElapsedTime     int64
	ConnectionCount      int
	ActiveSessions       int
	BatchRequestsPerSec  float64
	PageLifeExpectancy   int64
	BufferCacheHitRatio  float64
	LockWaitsPerSec      float64
	PageReadsPerSec      float64
	PageWritesPerSec     float64
}

// NewMSSQLMonitor creates a new MSSQL monitor
func NewMSSQLMonitor(db *sql.DB, config *dbconn.DBConfig) *MSSQLMonitor {
	return &MSSQLMonitor{
		db:            db,
		config:        config,
		checkInterval: 30 * time.Second,
		stopChan:      make(chan struct{}),
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MSSQLMonitor"},
	}
}

// Start starts the monitoring
func (m *MSSQLMonitor) Start(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	m.iLog.Info("Starting MSSQL monitoring")

	go func() {
		ticker := time.NewTicker(m.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.performHealthCheck(ctx)
			case <-m.stopChan:
				m.iLog.Info("Stopping MSSQL monitoring")
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the monitoring
func (m *MSSQLMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		close(m.stopChan)
		m.running = false
	}
}

// performHealthCheck performs a health check
func (m *MSSQLMonitor) performHealthCheck(ctx context.Context) {
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
}

// handleFailedCheck handles a failed health check
func (m *MSSQLMonitor) handleFailedCheck(err error) {
	m.mu.Lock()
	m.consecutiveFails++
	fails := m.consecutiveFails
	m.mu.Unlock()

	m.iLog.Error(fmt.Sprintf("MSSQL health check failed (consecutive: %d): %v", fails, err))

	// Attempt reconnection after 3 consecutive failures
	if fails >= 3 {
		m.iLog.Warn("Attempting to reconnect to MSSQL database...")
		if err := m.db.Ping(); err == nil {
			m.iLog.Info("Successfully reconnected to MSSQL database")
			m.mu.Lock()
			m.consecutiveFails = 0
			m.mu.Unlock()
		}
	}
}

// GetServerInfo retrieves SQL Server information
func (m *MSSQLMonitor) GetServerInfo(ctx context.Context) (*MSSQLServerInfo, error) {
	info := &MSSQLServerInfo{}

	// Get version and edition
	query := `
		SELECT
			@@VERSION as Version,
			SERVERPROPERTY('Edition') as Edition,
			SERVERPROPERTY('ProductLevel') as ProductLevel,
			SERVERPROPERTY('ProductVersion') as ProductVersion,
			SERVERPROPERTY('MachineName') as MachineName,
			SERVERPROPERTY('ServerName') as ServerName,
			ISNULL(SERVERPROPERTY('InstanceName'), 'DEFAULT') as InstanceName,
			SERVERPROPERTY('IsClustered') as IsClustered,
			SERVERPROPERTY('IsHadrEnabled') as IsHadrEnabled
	`

	var (
		edition      sql.NullString
		productLevel sql.NullString
		productVer   sql.NullString
		machineName  sql.NullString
		serverName   sql.NullString
		instanceName sql.NullString
		isClustered  sql.NullInt64
		isHadrEnabled sql.NullInt64
	)

	err := m.db.QueryRowContext(ctx, query).Scan(
		&info.Version,
		&edition,
		&productLevel,
		&productVer,
		&machineName,
		&serverName,
		&instanceName,
		&isClustered,
		&isHadrEnabled,
	)

	if err != nil {
		return nil, err
	}

	if edition.Valid {
		info.Edition = edition.String
	}
	if productLevel.Valid {
		info.ProductLevel = productLevel.String
	}
	if productVer.Valid {
		info.ProductVersion = productVer.String
	}
	if machineName.Valid {
		info.MachineName = machineName.String
	}
	if serverName.Valid {
		info.ServerName = serverName.String
	}
	if instanceName.Valid {
		info.InstanceName = instanceName.String
	}
	if isClustered.Valid {
		info.IsClustered = isClustered.Int64 == 1
	}
	if isHadrEnabled.Valid {
		info.IsHadrEnabled = isHadrEnabled.Int64 == 1
	}

	return info, nil
}

// GetPerformanceStats retrieves performance statistics
func (m *MSSQLMonitor) GetPerformanceStats(ctx context.Context) (*MSSQLPerformanceStats, error) {
	stats := &MSSQLPerformanceStats{}

	// Get CPU and elapsed time
	query := `
		SELECT
			total_worker_time as CPUTime,
			total_elapsed_time as TotalElapsedTime
		FROM sys.dm_exec_query_stats
		ORDER BY total_worker_time DESC
		OFFSET 0 ROWS FETCH NEXT 1 ROWS ONLY
	`

	err := m.db.QueryRowContext(ctx, query).Scan(&stats.CPUTime, &stats.TotalElapsedTime)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Get connection count and active sessions
	sessionQuery := `
		SELECT
			COUNT(*) as ConnectionCount,
			SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as ActiveSessions
		FROM sys.dm_exec_sessions
		WHERE is_user_process = 1
	`

	err = m.db.QueryRowContext(ctx, sessionQuery).Scan(&stats.ConnectionCount, &stats.ActiveSessions)
	if err != nil {
		return nil, err
	}

	// Get performance counters
	counterQuery := `
		SELECT
			counter_name,
			cntr_value
		FROM sys.dm_os_performance_counters
		WHERE object_name LIKE '%:Buffer Manager%'
		OR object_name LIKE '%:SQL Statistics%'
		OR object_name LIKE '%:Locks%'
	`

	rows, err := m.db.QueryContext(ctx, counterQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counters := make(map[string]int64)
	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		counters[name] = value
	}

	// Extract relevant counters
	if val, ok := counters["Page life expectancy"]; ok {
		stats.PageLifeExpectancy = val
	}
	if val, ok := counters["Buffer cache hit ratio"]; ok {
		stats.BufferCacheHitRatio = float64(val)
	}
	if val, ok := counters["Lock Waits/sec"]; ok {
		stats.LockWaitsPerSec = float64(val)
	}
	if val, ok := counters["Page reads/sec"]; ok {
		stats.PageReadsPerSec = float64(val)
	}
	if val, ok := counters["Page writes/sec"]; ok {
		stats.PageWritesPerSec = float64(val)
	}
	if val, ok := counters["Batch Requests/sec"]; ok {
		stats.BatchRequestsPerSec = float64(val)
	}

	return stats, nil
}

// GetActiveSessions retrieves currently active sessions
func (m *MSSQLMonitor) GetActiveSessions(ctx context.Context) ([]SessionInfo, error) {
	query := `
		SELECT
			s.session_id,
			s.login_name,
			s.host_name,
			s.program_name,
			s.status,
			r.command,
			r.wait_type,
			r.wait_time,
			r.cpu_time,
			r.total_elapsed_time,
			t.text as query_text
		FROM sys.dm_exec_sessions s
		LEFT JOIN sys.dm_exec_requests r ON s.session_id = r.session_id
		OUTER APPLY sys.dm_exec_sql_text(r.sql_handle) t
		WHERE s.is_user_process = 1
		AND s.session_id <> @@SPID
		ORDER BY s.session_id
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]SessionInfo, 0)

	for rows.Next() {
		var (
			sessionID    int
			loginName    string
			hostName     sql.NullString
			programName  sql.NullString
			status       string
			command      sql.NullString
			waitType     sql.NullString
			waitTime     sql.NullInt64
			cpuTime      sql.NullInt64
			elapsedTime  sql.NullInt64
			queryText    sql.NullString
		)

		if err := rows.Scan(&sessionID, &loginName, &hostName, &programName, &status,
			&command, &waitType, &waitTime, &cpuTime, &elapsedTime, &queryText); err != nil {
			return nil, err
		}

		session := SessionInfo{
			SessionID: sessionID,
			LoginName: loginName,
			Status:    status,
		}

		if hostName.Valid {
			session.HostName = hostName.String
		}
		if programName.Valid {
			session.ProgramName = programName.String
		}
		if command.Valid {
			session.Command = command.String
		}
		if waitType.Valid {
			session.WaitType = waitType.String
		}
		if waitTime.Valid {
			session.WaitTime = int(waitTime.Int64)
		}
		if cpuTime.Valid {
			session.CPUTime = cpuTime.Int64
		}
		if elapsedTime.Valid {
			session.ElapsedTime = elapsedTime.Int64
		}
		if queryText.Valid {
			session.QueryText = queryText.String
		}

		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// KillSession terminates a session by ID
func (m *MSSQLMonitor) KillSession(ctx context.Context, sessionID int) error {
	query := fmt.Sprintf("KILL %d", sessionID)
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// GetDatabaseSizes retrieves database sizes
func (m *MSSQLMonitor) GetDatabaseSizes(ctx context.Context) ([]DatabaseSize, error) {
	query := `
		SELECT
			DB_NAME(database_id) as database_name,
			SUM(CASE WHEN type = 0 THEN size END) * 8 / 1024 as data_size_mb,
			SUM(CASE WHEN type = 1 THEN size END) * 8 / 1024 as log_size_mb,
			SUM(size) * 8 / 1024 as total_size_mb
		FROM sys.master_files
		WHERE database_id > 4
		GROUP BY database_id
		ORDER BY total_size_mb DESC
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	databases := make([]DatabaseSize, 0)

	for rows.Next() {
		var db DatabaseSize
		if err := rows.Scan(&db.DatabaseName, &db.DataSizeMB, &db.LogSizeMB, &db.TotalSizeMB); err != nil {
			return nil, err
		}
		databases = append(databases, db)
	}

	return databases, rows.Err()
}

// GetTableSizes retrieves table sizes for the current database
func (m *MSSQLMonitor) GetTableSizes(ctx context.Context) ([]TableSize, error) {
	query := `
		SELECT
			s.Name AS SchemaName,
			t.NAME AS TableName,
			p.rows AS RowCounts,
			SUM(a.total_pages) * 8 / 1024 AS TotalSpaceMB,
			SUM(a.used_pages) * 8 / 1024 AS UsedSpaceMB,
			(SUM(a.total_pages) - SUM(a.used_pages)) * 8 / 1024 AS UnusedSpaceMB
		FROM sys.tables t
		INNER JOIN sys.indexes i ON t.OBJECT_ID = i.object_id
		INNER JOIN sys.partitions p ON i.object_id = p.OBJECT_ID AND i.index_id = p.index_id
		INNER JOIN sys.allocation_units a ON p.partition_id = a.container_id
		LEFT OUTER JOIN sys.schemas s ON t.schema_id = s.schema_id
		WHERE t.is_ms_shipped = 0
		AND i.OBJECT_ID > 255
		GROUP BY s.Name, t.Name, p.Rows
		ORDER BY TotalSpaceMB DESC
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]TableSize, 0)

	for rows.Next() {
		var table TableSize
		if err := rows.Scan(&table.SchemaName, &table.TableName, &table.RowCount,
			&table.TotalSpaceMB, &table.UsedSpaceMB, &table.UnusedSpaceMB); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

// GetLastCheckTime returns the last health check time
func (m *MSSQLMonitor) GetLastCheckTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastCheckTime
}

// GetConsecutiveFailures returns the number of consecutive failures
func (m *MSSQLMonitor) GetConsecutiveFailures() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.consecutiveFails
}

// IsHealthy checks if the database is healthy
func (m *MSSQLMonitor) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.consecutiveFails == 0
}

// Supporting types

// SessionInfo represents an active session
type SessionInfo struct {
	SessionID   int
	LoginName   string
	HostName    string
	ProgramName string
	Status      string
	Command     string
	WaitType    string
	WaitTime    int
	CPUTime     int64
	ElapsedTime int64
	QueryText   string
}

// DatabaseSize represents database size information
type DatabaseSize struct {
	DatabaseName string
	DataSizeMB   int64
	LogSizeMB    int64
	TotalSizeMB  int64
}

// TableSize represents table size information
type TableSize struct {
	SchemaName    string
	TableName     string
	RowCount      int64
	TotalSpaceMB  int64
	UsedSpaceMB   int64
	UnusedSpaceMB int64
}
