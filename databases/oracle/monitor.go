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

package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/logger"
)

// OracleMonitor monitors Oracle database health and performance
type OracleMonitor struct {
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

// OracleInstanceInfo represents Oracle instance information
type OracleInstanceInfo struct {
	InstanceName string
	HostName     string
	Version      string
	Status       string
	StartTime    time.Time
	Uptime       int64
	DatabaseRole string
	OpenMode     string
	LogMode      string
}

// OraclePerformanceStats represents performance statistics
type OraclePerformanceStats struct {
	SessionCount       int
	ActiveSessions     int
	PhysicalReads      int64
	PhysicalWrites     int64
	LogicalReads       int64
	BufferCacheHitRate float64
	LibraryCacheHitRate float64
	CPUUsage           float64
	MemoryUsage        int64
	PGAUsage           int64
	SGAUsage           int64
}

// NewOracleMonitor creates a new Oracle monitor
func NewOracleMonitor(db *sql.DB, config *dbconn.DBConfig) *OracleMonitor {
	return &OracleMonitor{
		db:            db,
		config:        config,
		checkInterval: 30 * time.Second,
		stopChan:      make(chan struct{}),
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "OracleMonitor"},
	}
}

// Start starts the monitoring
func (m *OracleMonitor) Start(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	m.iLog.Info("Starting Oracle monitoring")

	go func() {
		ticker := time.NewTicker(m.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.performHealthCheck(ctx)
			case <-m.stopChan:
				m.iLog.Info("Stopping Oracle monitoring")
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the monitoring
func (m *OracleMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		close(m.stopChan)
		m.running = false
	}
}

// performHealthCheck performs a health check
func (m *OracleMonitor) performHealthCheck(ctx context.Context) {
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
func (m *OracleMonitor) handleFailedCheck(err error) {
	m.mu.Lock()
	m.consecutiveFails++
	fails := m.consecutiveFails
	m.mu.Unlock()

	m.iLog.Error(fmt.Sprintf("Oracle health check failed (consecutive: %d): %v", fails, err))

	// Attempt reconnection after 3 consecutive failures
	if fails >= 3 {
		m.iLog.Warn("Attempting to reconnect to Oracle database...")
		if err := m.db.Ping(); err == nil {
			m.iLog.Info("Successfully reconnected to Oracle database")
			m.mu.Lock()
			m.consecutiveFails = 0
			m.mu.Unlock()
		}
	}
}

// GetInstanceInfo retrieves Oracle instance information
func (m *OracleMonitor) GetInstanceInfo(ctx context.Context) (*OracleInstanceInfo, error) {
	info := &OracleInstanceInfo{}

	query := `
		SELECT
			i.instance_name,
			i.host_name,
			i.version,
			i.status,
			i.startup_time,
			d.database_role,
			d.open_mode,
			d.log_mode
		FROM v$instance i, v$database d
		WHERE rownum = 1
	`

	var startupTime time.Time
	err := m.db.QueryRowContext(ctx, query).Scan(
		&info.InstanceName,
		&info.HostName,
		&info.Version,
		&info.Status,
		&startupTime,
		&info.DatabaseRole,
		&info.OpenMode,
		&info.LogMode,
	)

	if err != nil {
		return nil, err
	}

	info.StartTime = startupTime
	info.Uptime = int64(time.Since(startupTime).Seconds())

	return info, nil
}

// GetPerformanceStats retrieves performance statistics
func (m *OracleMonitor) GetPerformanceStats(ctx context.Context) (*OraclePerformanceStats, error) {
	stats := &OraclePerformanceStats{}

	// Get session count
	sessionQuery := `
		SELECT
			COUNT(*) as total_sessions,
			SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END) as active_sessions
		FROM v$session
		WHERE type = 'USER'
	`

	err := m.db.QueryRowContext(ctx, sessionQuery).Scan(&stats.SessionCount, &stats.ActiveSessions)
	if err != nil {
		return nil, err
	}

	// Get I/O statistics
	ioQuery := `
		SELECT
			SUM(CASE WHEN name = 'physical reads' THEN value ELSE 0 END) as physical_reads,
			SUM(CASE WHEN name = 'physical writes' THEN value ELSE 0 END) as physical_writes,
			SUM(CASE WHEN name = 'db block gets' THEN value ELSE 0 END) +
			SUM(CASE WHEN name = 'consistent gets' THEN value ELSE 0 END) as logical_reads
		FROM v$sysstat
		WHERE name IN ('physical reads', 'physical writes', 'db block gets', 'consistent gets')
	`

	err = m.db.QueryRowContext(ctx, ioQuery).Scan(
		&stats.PhysicalReads,
		&stats.PhysicalWrites,
		&stats.LogicalReads,
	)
	if err != nil {
		return nil, err
	}

	// Calculate buffer cache hit rate
	if stats.LogicalReads > 0 {
		stats.BufferCacheHitRate = (1.0 - float64(stats.PhysicalReads)/float64(stats.LogicalReads)) * 100
	}

	// Get library cache hit rate
	libraryCacheQuery := `
		SELECT
			SUM(pins) as pins,
			SUM(pinhits) as pinhits
		FROM v$librarycache
	`

	var pins, pinhits int64
	err = m.db.QueryRowContext(ctx, libraryCacheQuery).Scan(&pins, &pinhits)
	if err == nil && pins > 0 {
		stats.LibraryCacheHitRate = (float64(pinhits) / float64(pins)) * 100
	}

	// Get memory usage (SGA and PGA)
	memoryQuery := `
		SELECT
			(SELECT value FROM v$sga WHERE name = 'Total System Global Area') as sga_size,
			(SELECT value FROM v$pgastat WHERE name = 'total PGA allocated') as pga_size
		FROM dual
	`

	err = m.db.QueryRowContext(ctx, memoryQuery).Scan(&stats.SGAUsage, &stats.PGAUsage)
	if err == nil {
		stats.MemoryUsage = stats.SGAUsage + stats.PGAUsage
	}

	return stats, nil
}

// GetActiveSessions retrieves currently active sessions
func (m *OracleMonitor) GetActiveSessions(ctx context.Context) ([]OracleSessionInfo, error) {
	query := `
		SELECT
			s.sid,
			s.serial#,
			s.username,
			s.osuser,
			s.machine,
			s.program,
			s.status,
			s.logon_time,
			sq.sql_text,
			s.last_call_et
		FROM v$session s
		LEFT JOIN v$sql sq ON s.sql_id = sq.sql_id
		WHERE s.type = 'USER'
		AND s.username IS NOT NULL
		ORDER BY s.logon_time DESC
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]OracleSessionInfo, 0)

	for rows.Next() {
		var (
			sid        int
			serial     int
			username   sql.NullString
			osuser     sql.NullString
			machine    sql.NullString
			program    sql.NullString
			status     string
			logonTime  time.Time
			sqlText    sql.NullString
			lastCallET int
		)

		if err := rows.Scan(&sid, &serial, &username, &osuser, &machine, &program,
			&status, &logonTime, &sqlText, &lastCallET); err != nil {
			return nil, err
		}

		session := OracleSessionInfo{
			SID:        sid,
			Serial:     serial,
			Status:     status,
			LogonTime:  logonTime,
			LastCallET: lastCallET,
		}

		if username.Valid {
			session.Username = username.String
		}
		if osuser.Valid {
			session.OSUser = osuser.String
		}
		if machine.Valid {
			session.Machine = machine.String
		}
		if program.Valid {
			session.Program = program.String
		}
		if sqlText.Valid {
			session.SQLText = sqlText.String
		}

		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// KillSession terminates a session
func (m *OracleMonitor) KillSession(ctx context.Context, sid, serial int) error {
	query := fmt.Sprintf("ALTER SYSTEM KILL SESSION '%d,%d'", sid, serial)
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// GetTablespaceSizes retrieves tablespace sizes
func (m *OracleMonitor) GetTablespaceSizes(ctx context.Context) ([]TablespaceSize, error) {
	query := `
		SELECT
			df.tablespace_name,
			NVL(SUM(df.bytes) / 1024 / 1024, 0) as total_mb,
			NVL(SUM(df.bytes) / 1024 / 1024, 0) - NVL(SUM(fs.bytes) / 1024 / 1024, 0) as used_mb,
			NVL(SUM(fs.bytes) / 1024 / 1024, 0) as free_mb,
			ROUND((1 - (NVL(SUM(fs.bytes), 0) / SUM(df.bytes))) * 100, 2) as pct_used
		FROM dba_data_files df
		LEFT JOIN (
			SELECT tablespace_name, SUM(bytes) as bytes
			FROM dba_free_space
			GROUP BY tablespace_name
		) fs ON df.tablespace_name = fs.tablespace_name
		GROUP BY df.tablespace_name
		ORDER BY pct_used DESC
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tablespaces := make([]TablespaceSize, 0)

	for rows.Next() {
		var ts TablespaceSize
		if err := rows.Scan(&ts.TablespaceName, &ts.TotalMB, &ts.UsedMB, &ts.FreeMB, &ts.PctUsed); err != nil {
			return nil, err
		}
		tablespaces = append(tablespaces, ts)
	}

	return tablespaces, rows.Err()
}

// GetTopQueries retrieves top queries by execution time
func (m *OracleMonitor) GetTopQueries(ctx context.Context, limit int) ([]TopQuery, error) {
	query := fmt.Sprintf(`
		SELECT * FROM (
			SELECT
				sql_id,
				sql_text,
				executions,
				elapsed_time / 1000000 as elapsed_seconds,
				cpu_time / 1000000 as cpu_seconds,
				disk_reads,
				buffer_gets,
				rows_processed
			FROM v$sql
			WHERE executions > 0
			ORDER BY elapsed_time DESC
		) WHERE rownum <= %d
	`, limit)

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	queries := make([]TopQuery, 0)

	for rows.Next() {
		var q TopQuery
		if err := rows.Scan(&q.SQLID, &q.SQLText, &q.Executions, &q.ElapsedSeconds,
			&q.CPUSeconds, &q.DiskReads, &q.BufferGets, &q.RowsProcessed); err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}

	return queries, rows.Err()
}

// GetLastCheckTime returns the last health check time
func (m *OracleMonitor) GetLastCheckTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastCheckTime
}

// GetConsecutiveFailures returns the number of consecutive failures
func (m *OracleMonitor) GetConsecutiveFailures() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.consecutiveFails
}

// IsHealthy checks if the database is healthy
func (m *OracleMonitor) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.consecutiveFails == 0
}

// Supporting types

// OracleSessionInfo represents an active Oracle session
type OracleSessionInfo struct {
	SID        int
	Serial     int
	Username   string
	OSUser     string
	Machine    string
	Program    string
	Status     string
	LogonTime  time.Time
	SQLText    string
	LastCallET int
}

// TablespaceSize represents tablespace size information
type TablespaceSize struct {
	TablespaceName string
	TotalMB        float64
	UsedMB         float64
	FreeMB         float64
	PctUsed        float64
}

// TopQuery represents a top query by execution time
type TopQuery struct {
	SQLID          string
	SQLText        string
	Executions     int64
	ElapsedSeconds float64
	CPUSeconds     float64
	DiskReads      int64
	BufferGets     int64
	RowsProcessed  int64
}
