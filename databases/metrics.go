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

package dbconn

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/mdaxf/iac/logger"
)

// QueryMetrics tracks query execution metrics
type QueryMetrics struct {
	TotalQueries     int64
	TotalExecs       int64
	TotalErrors      int64
	SlowQueries      int64
	TotalDuration    time.Duration
	SlowQueryThreshold time.Duration

	// Query type breakdown
	SelectQueries    int64
	InsertQueries    int64
	UpdateQueries    int64
	DeleteQueries    int64
	OtherQueries     int64

	mu sync.RWMutex
}

// QueryLog represents a single query execution log
type QueryLog struct {
	Timestamp    time.Time
	Query        string
	Args         []interface{}
	Duration     time.Duration
	Error        error
	DBType       DBType
	Operation    string
	RowsAffected int64
}

// MetricsCollector collects and aggregates database metrics
type MetricsCollector struct {
	metrics         map[DBType]*QueryMetrics
	queryLogs       []QueryLog
	maxLogSize      int
	slowQueryThreshold time.Duration
	enableLogging   bool
	enableMetrics   bool
	mu              sync.RWMutex
	iLog            logger.Log
}

var (
	globalMetricsCollector *MetricsCollector
	metricsOnce            sync.Once
)

// GetMetricsCollector returns the global metrics collector
func GetMetricsCollector() *MetricsCollector {
	metricsOnce.Do(func() {
		globalMetricsCollector = NewMetricsCollector()
	})
	return globalMetricsCollector
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:            make(map[DBType]*QueryMetrics),
		queryLogs:          make([]QueryLog, 0, 1000),
		maxLogSize:         1000,
		slowQueryThreshold: 1 * time.Second,
		enableLogging:      true,
		enableMetrics:      true,
		iLog:               logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MetricsCollector"},
	}
}

// SetSlowQueryThreshold sets the threshold for slow query detection
func (mc *MetricsCollector) SetSlowQueryThreshold(threshold time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.slowQueryThreshold = threshold
}

// SetMaxLogSize sets the maximum number of query logs to keep
func (mc *MetricsCollector) SetMaxLogSize(size int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.maxLogSize = size
}

// EnableLogging enables or disables query logging
func (mc *MetricsCollector) EnableLogging(enable bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.enableLogging = enable
}

// EnableMetrics enables or disables metrics collection
func (mc *MetricsCollector) EnableMetrics(enable bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.enableMetrics = enable
}

// getOrCreateMetrics gets or creates metrics for a database type
func (mc *MetricsCollector) getOrCreateMetrics(dbType DBType) *QueryMetrics {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.metrics[dbType]; !exists {
		mc.metrics[dbType] = &QueryMetrics{
			SlowQueryThreshold: mc.slowQueryThreshold,
		}
	}
	return mc.metrics[dbType]
}

// RecordQuery records a query execution
func (mc *MetricsCollector) RecordQuery(dbType DBType, query string, args []interface{}, duration time.Duration, err error, rowsAffected int64) {
	if !mc.enableMetrics && !mc.enableLogging {
		return
	}

	operation := detectOperation(query)

	// Record metrics
	if mc.enableMetrics {
		metrics := mc.getOrCreateMetrics(dbType)
		metrics.mu.Lock()
		metrics.TotalQueries++
		metrics.TotalDuration += duration

		if err != nil {
			metrics.TotalErrors++
		}

		if duration >= mc.slowQueryThreshold {
			metrics.SlowQueries++
		}

		// Track query types
		switch operation {
		case "SELECT":
			metrics.SelectQueries++
		case "INSERT":
			metrics.InsertQueries++
		case "UPDATE":
			metrics.UpdateQueries++
		case "DELETE":
			metrics.DeleteQueries++
		default:
			metrics.OtherQueries++
		}

		metrics.mu.Unlock()
	}

	// Log query
	if mc.enableLogging {
		queryLog := QueryLog{
			Timestamp:    time.Now(),
			Query:        query,
			Args:         args,
			Duration:     duration,
			Error:        err,
			DBType:       dbType,
			Operation:    operation,
			RowsAffected: rowsAffected,
		}

		mc.addQueryLog(queryLog)

		// Log slow queries
		if duration >= mc.slowQueryThreshold {
			mc.iLog.Warn(fmt.Sprintf("Slow query detected [%s] Duration: %v, Query: %s",
				dbType, duration, truncateQuery(query, 200)))
		}

		// Log errors
		if err != nil {
			mc.iLog.Error(fmt.Sprintf("Query error [%s]: %v, Query: %s",
				dbType, err, truncateQuery(query, 200)))
		}
	}
}

// RecordExec records an exec operation
func (mc *MetricsCollector) RecordExec(dbType DBType, query string, args []interface{}, duration time.Duration, err error, rowsAffected int64) {
	if !mc.enableMetrics && !mc.enableLogging {
		return
	}

	operation := detectOperation(query)

	// Record metrics
	if mc.enableMetrics {
		metrics := mc.getOrCreateMetrics(dbType)
		metrics.mu.Lock()
		metrics.TotalExecs++
		metrics.TotalDuration += duration

		if err != nil {
			metrics.TotalErrors++
		}

		if duration >= mc.slowQueryThreshold {
			metrics.SlowQueries++
		}

		// Track query types
		switch operation {
		case "INSERT":
			metrics.InsertQueries++
		case "UPDATE":
			metrics.UpdateQueries++
		case "DELETE":
			metrics.DeleteQueries++
		default:
			metrics.OtherQueries++
		}

		metrics.mu.Unlock()
	}

	// Log exec
	if mc.enableLogging {
		queryLog := QueryLog{
			Timestamp:    time.Now(),
			Query:        query,
			Args:         args,
			Duration:     duration,
			Error:        err,
			DBType:       dbType,
			Operation:    operation,
			RowsAffected: rowsAffected,
		}

		mc.addQueryLog(queryLog)

		// Log slow operations
		if duration >= mc.slowQueryThreshold {
			mc.iLog.Warn(fmt.Sprintf("Slow exec detected [%s] Duration: %v, Query: %s",
				dbType, duration, truncateQuery(query, 200)))
		}

		// Log errors
		if err != nil {
			mc.iLog.Error(fmt.Sprintf("Exec error [%s]: %v, Query: %s",
				dbType, err, truncateQuery(query, 200)))
		}
	}
}

// addQueryLog adds a query log entry
func (mc *MetricsCollector) addQueryLog(log QueryLog) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.queryLogs = append(mc.queryLogs, log)

	// Keep only the most recent logs
	if len(mc.queryLogs) > mc.maxLogSize {
		mc.queryLogs = mc.queryLogs[len(mc.queryLogs)-mc.maxLogSize:]
	}
}

// GetMetrics returns metrics for a specific database type
func (mc *MetricsCollector) GetMetrics(dbType DBType) *QueryMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if metrics, exists := mc.metrics[dbType]; exists {
		// Return a copy to avoid race conditions
		metrics.mu.RLock()
		defer metrics.mu.RUnlock()

		return &QueryMetrics{
			TotalQueries:       metrics.TotalQueries,
			TotalExecs:         metrics.TotalExecs,
			TotalErrors:        metrics.TotalErrors,
			SlowQueries:        metrics.SlowQueries,
			TotalDuration:      metrics.TotalDuration,
			SlowQueryThreshold: metrics.SlowQueryThreshold,
			SelectQueries:      metrics.SelectQueries,
			InsertQueries:      metrics.InsertQueries,
			UpdateQueries:      metrics.UpdateQueries,
			DeleteQueries:      metrics.DeleteQueries,
			OtherQueries:       metrics.OtherQueries,
		}
	}

	return &QueryMetrics{}
}

// GetAllMetrics returns metrics for all database types
func (mc *MetricsCollector) GetAllMetrics() map[DBType]*QueryMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[DBType]*QueryMetrics)
	for dbType, metrics := range mc.metrics {
		metrics.mu.RLock()
		result[dbType] = &QueryMetrics{
			TotalQueries:       metrics.TotalQueries,
			TotalExecs:         metrics.TotalExecs,
			TotalErrors:        metrics.TotalErrors,
			SlowQueries:        metrics.SlowQueries,
			TotalDuration:      metrics.TotalDuration,
			SlowQueryThreshold: metrics.SlowQueryThreshold,
			SelectQueries:      metrics.SelectQueries,
			InsertQueries:      metrics.InsertQueries,
			UpdateQueries:      metrics.UpdateQueries,
			DeleteQueries:      metrics.DeleteQueries,
			OtherQueries:       metrics.OtherQueries,
		}
		metrics.mu.RUnlock()
	}

	return result
}

// GetRecentLogs returns the most recent query logs
func (mc *MetricsCollector) GetRecentLogs(limit int) []QueryLog {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	start := 0
	if len(mc.queryLogs) > limit {
		start = len(mc.queryLogs) - limit
	}

	logs := make([]QueryLog, len(mc.queryLogs)-start)
	copy(logs, mc.queryLogs[start:])

	return logs
}

// GetSlowQueries returns recent slow queries
func (mc *MetricsCollector) GetSlowQueries(limit int) []QueryLog {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	slowQueries := make([]QueryLog, 0)

	for i := len(mc.queryLogs) - 1; i >= 0 && len(slowQueries) < limit; i-- {
		if mc.queryLogs[i].Duration >= mc.slowQueryThreshold {
			slowQueries = append(slowQueries, mc.queryLogs[i])
		}
	}

	return slowQueries
}

// ResetMetrics resets all metrics
func (mc *MetricsCollector) ResetMetrics() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics = make(map[DBType]*QueryMetrics)
	mc.queryLogs = make([]QueryLog, 0, mc.maxLogSize)

	mc.iLog.Info("Database metrics reset")
}

// detectOperation detects the SQL operation type from a query
func detectOperation(query string) string {
	// Simple detection based on first keyword
	trimmed := trimLeadingWhitespace(query)
	if len(trimmed) < 6 {
		return "UNKNOWN"
	}

	prefix := toUpperFirst6(trimmed)

	if startsWith(prefix, "SELECT") {
		return "SELECT"
	} else if startsWith(prefix, "INSERT") {
		return "INSERT"
	} else if startsWith(prefix, "UPDATE") {
		return "UPDATE"
	} else if startsWith(prefix, "DELETE") {
		return "DELETE"
	} else if startsWith(prefix, "CREATE") {
		return "CREATE"
	} else if startsWith(prefix, "ALTER") {
		return "ALTER"
	} else if startsWith(prefix, "DROP") {
		return "DROP"
	}

	return "OTHER"
}

// Helper functions
func trimLeadingWhitespace(s string) string {
	for i, c := range s {
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			return s[i:]
		}
	}
	return s
}

func toUpperFirst6(s string) string {
	if len(s) > 6 {
		s = s[:6]
	}

	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'a' && s[i] <= 'z' {
			result[i] = s[i] - 32
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

func startsWith(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}

func truncateQuery(query string, maxLen int) string {
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "..."
}

// TrackedDB wraps a RelationalDB with query tracking
type TrackedDB struct {
	RelationalDB
	dbType DBType
}

// NewTrackedDB creates a new tracked database wrapper
func NewTrackedDB(db RelationalDB, dbType DBType) *TrackedDB {
	return &TrackedDB{
		RelationalDB: db,
		dbType:       dbType,
	}
}

// Query executes a query with tracking
func (t *TrackedDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := t.RelationalDB.Query(ctx, query, args...)
	duration := time.Since(start)

	GetMetricsCollector().RecordQuery(t.dbType, query, args, duration, err, 0)

	return rows, err
}

// Exec executes a statement with tracking
func (t *TrackedDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := t.RelationalDB.Exec(ctx, query, args...)
	duration := time.Since(start)

	var rowsAffected int64
	if result != nil {
		rowsAffected, _ = result.RowsAffected()
	}

	GetMetricsCollector().RecordExec(t.dbType, query, args, duration, err, rowsAffected)

	return result, err
}

// GetAverageQueryTime returns the average query execution time
func (qm *QueryMetrics) GetAverageQueryTime() time.Duration {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	total := qm.TotalQueries + qm.TotalExecs
	if total == 0 {
		return 0
	}

	return time.Duration(int64(qm.TotalDuration) / total)
}

// GetErrorRate returns the error rate as a percentage
func (qm *QueryMetrics) GetErrorRate() float64 {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	total := qm.TotalQueries + qm.TotalExecs
	if total == 0 {
		return 0
	}

	return float64(qm.TotalErrors) / float64(total) * 100
}

// GetSlowQueryRate returns the slow query rate as a percentage
func (qm *QueryMetrics) GetSlowQueryRate() float64 {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	total := qm.TotalQueries + qm.TotalExecs
	if total == 0 {
		return 0
	}

	return float64(qm.SlowQueries) / float64(total) * 100
}
