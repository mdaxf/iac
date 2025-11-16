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

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/logger"
)

// PostgreSQLMonitor monitors PostgreSQL database health and performance
type PostgreSQLMonitor struct {
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

// NewPostgreSQLMonitor creates a new PostgreSQL monitor
func NewPostgreSQLMonitor(db *sql.DB, config *dbconn.DBConfig) *PostgreSQLMonitor {
	return &PostgreSQLMonitor{
		db:            db,
		config:        config,
		checkInterval: 30 * time.Second,
		stopChan:      make(chan struct{}),
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "PostgreSQLMonitor"},
	}
}

// Start starts the monitoring
func (m *PostgreSQLMonitor) Start(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	m.iLog.Info("Starting PostgreSQL monitoring")

	go func() {
		ticker := time.NewTicker(m.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.performHealthCheck(ctx)
			case <-m.stopChan:
				m.iLog.Info("Stopping PostgreSQL monitoring")
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the monitoring
func (m *PostgreSQLMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		close(m.stopChan)
		m.running = false
	}
}

// performHealthCheck performs a health check
func (m *PostgreSQLMonitor) performHealthCheck(ctx context.Context) {
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
func (m *PostgreSQLMonitor) handleFailedCheck(err error) {
	m.mu.Lock()
	m.consecutiveFails++
	fails := m.consecutiveFails
	m.mu.Unlock()

	m.iLog.Error(fmt.Sprintf("PostgreSQL health check failed (consecutive: %d): %v", fails, err))

	// Attempt reconnection after 3 consecutive failures
	if fails >= 3 {
		m.iLog.Warn("Attempting to reconnect to PostgreSQL database...")
		if err := m.db.Ping(); err == nil {
			m.iLog.Info("Successfully reconnected to PostgreSQL database")
			m.mu.Lock()
			m.consecutiveFails = 0
			m.mu.Unlock()
		}
	}
}

// GetDatabaseStats retrieves PostgreSQL database statistics
func (m *PostgreSQLMonitor) GetDatabaseStats(ctx context.Context) (*DatabaseStats, error) {
	query := `
		SELECT
			numbackends,
			xact_commit,
			xact_rollback,
			blks_read,
			blks_hit,
			tup_returned,
			tup_fetched,
			tup_inserted,
			tup_updated,
			tup_deleted
		FROM pg_stat_database
		WHERE datname = current_database()
	`

	stats := &DatabaseStats{}
	err := m.db.QueryRowContext(ctx, query).Scan(
		&stats.ActiveConnections,
		&stats.TransactionsCommitted,
		&stats.TransactionsRolledBack,
		&stats.BlocksRead,
		&stats.BlocksHit,
		&stats.TuplesReturned,
		&stats.TuplesFetched,
		&stats.TuplesInserted,
		&stats.TuplesUpdated,
		&stats.TuplesDeleted,
	)

	if err != nil {
		return nil, err
	}

	// Calculate cache hit ratio
	totalBlocks := stats.BlocksRead + stats.BlocksHit
	if totalBlocks > 0 {
		stats.CacheHitRatio = float64(stats.BlocksHit) / float64(totalBlocks) * 100
	}

	return stats, nil
}

// GetActiveQueries retrieves currently running queries
func (m *PostgreSQLMonitor) GetActiveQueries(ctx context.Context) ([]ActiveQuery, error) {
	query := `
		SELECT
			pid,
			usename,
			application_name,
			client_addr,
			state,
			query,
			EXTRACT(EPOCH FROM (now() - query_start)) as duration
		FROM pg_stat_activity
		WHERE state != 'idle'
		AND pid != pg_backend_pid()
		ORDER BY query_start
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	queries := make([]ActiveQuery, 0)

	for rows.Next() {
		var (
			pid         int
			username    string
			appName     sql.NullString
			clientAddr  sql.NullString
			state       string
			queryText   string
			duration    float64
		)

		if err := rows.Scan(&pid, &username, &appName, &clientAddr, &state, &queryText, &duration); err != nil {
			return nil, err
		}

		query := ActiveQuery{
			PID:      pid,
			Username: username,
			State:    state,
			Query:    queryText,
			Duration: duration,
		}

		if appName.Valid {
			query.AppName = appName.String
		}
		if clientAddr.Valid {
			query.ClientAddr = clientAddr.String
		}

		queries = append(queries, query)
	}

	return queries, rows.Err()
}

// TerminateQuery terminates a query by PID
func (m *PostgreSQLMonitor) TerminateQuery(ctx context.Context, pid int) error {
	query := fmt.Sprintf("SELECT pg_terminate_backend(%d)", pid)
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// GetTableStats retrieves table statistics
func (m *PostgreSQLMonitor) GetTableStats(ctx context.Context) ([]TableStats, error) {
	query := `
		SELECT
			schemaname,
			tablename,
			n_live_tup,
			n_dead_tup,
			n_tup_ins,
			n_tup_upd,
			n_tup_del,
			last_vacuum,
			last_autovacuum,
			last_analyze,
			last_autoanalyze
		FROM pg_stat_user_tables
		ORDER BY n_live_tup DESC
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]TableStats, 0)

	for rows.Next() {
		var (
			schema          string
			table           string
			liveTuples      int64
			deadTuples      int64
			inserted        int64
			updated         int64
			deleted         int64
			lastVacuum      sql.NullTime
			lastAutovacuum  sql.NullTime
			lastAnalyze     sql.NullTime
			lastAutoanalyze sql.NullTime
		)

		if err := rows.Scan(&schema, &table, &liveTuples, &deadTuples, &inserted, &updated, &deleted,
			&lastVacuum, &lastAutovacuum, &lastAnalyze, &lastAutoanalyze); err != nil {
			return nil, err
		}

		stats := TableStats{
			Schema:      schema,
			Table:       table,
			LiveTuples:  liveTuples,
			DeadTuples:  deadTuples,
			Inserted:    inserted,
			Updated:     updated,
			Deleted:     deleted,
		}

		if lastVacuum.Valid {
			stats.LastVacuum = &lastVacuum.Time
		}
		if lastAutovacuum.Valid {
			stats.LastAutovacuum = &lastAutovacuum.Time
		}
		if lastAnalyze.Valid {
			stats.LastAnalyze = &lastAnalyze.Time
		}
		if lastAutoanalyze.Valid {
			stats.LastAutoanalyze = &lastAutoanalyze.Time
		}

		tables = append(tables, stats)
	}

	return tables, rows.Err()
}

// GetIndexStats retrieves index statistics
func (m *PostgreSQLMonitor) GetIndexStats(ctx context.Context) ([]IndexStats, error) {
	query := `
		SELECT
			schemaname,
			tablename,
			indexname,
			idx_scan,
			idx_tup_read,
			idx_tup_fetch
		FROM pg_stat_user_indexes
		ORDER BY idx_scan DESC
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make([]IndexStats, 0)

	for rows.Next() {
		var stats IndexStats
		if err := rows.Scan(&stats.Schema, &stats.Table, &stats.Index,
			&stats.Scans, &stats.TuplesRead, &stats.TuplesFetched); err != nil {
			return nil, err
		}
		indexes = append(indexes, stats)
	}

	return indexes, rows.Err()
}

// GetLastCheckTime returns the last health check time
func (m *PostgreSQLMonitor) GetLastCheckTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastCheckTime
}

// GetConsecutiveFailures returns the number of consecutive failures
func (m *PostgreSQLMonitor) GetConsecutiveFailures() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.consecutiveFails
}

// IsHealthy checks if the database is healthy
func (m *PostgreSQLMonitor) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.consecutiveFails == 0
}

// Supporting types

// DatabaseStats represents PostgreSQL database statistics
type DatabaseStats struct {
	ActiveConnections       int
	TransactionsCommitted   int64
	TransactionsRolledBack  int64
	BlocksRead              int64
	BlocksHit               int64
	TuplesReturned          int64
	TuplesFetched           int64
	TuplesInserted          int64
	TuplesUpdated           int64
	TuplesDeleted           int64
	CacheHitRatio           float64
}

// ActiveQuery represents an active query
type ActiveQuery struct {
	PID        int
	Username   string
	AppName    string
	ClientAddr string
	State      string
	Query      string
	Duration   float64
}

// TableStats represents table statistics
type TableStats struct {
	Schema          string
	Table           string
	LiveTuples      int64
	DeadTuples      int64
	Inserted        int64
	Updated         int64
	Deleted         int64
	LastVacuum      *time.Time
	LastAutovacuum  *time.Time
	LastAnalyze     *time.Time
	LastAutoanalyze *time.Time
}

// IndexStats represents index statistics
type IndexStats struct {
	Schema        string
	Table         string
	Index         string
	Scans         int64
	TuplesRead    int64
	TuplesFetched int64
}
