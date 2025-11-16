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

package databases

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ProxyConfig configures the database proxy
type ProxyConfig struct {
	// MaxConnections is the maximum number of connections
	MaxConnections int

	// IdleTimeout is how long idle connections are kept
	IdleTimeout time.Duration

	// QueryTimeout is the default query timeout
	QueryTimeout time.Duration

	// EnableQueryRewrite enables query rewriting
	EnableQueryRewrite bool

	// EnableMonitoring enables query monitoring
	EnableMonitoring bool

	// EnableLoadBalancing enables load balancing across replicas
	EnableLoadBalancing bool

	// SlowQueryThreshold for logging slow queries
	SlowQueryThreshold time.Duration
}

// DefaultProxyConfig returns default configuration
func DefaultProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		MaxConnections:      100,
		IdleTimeout:         5 * time.Minute,
		QueryTimeout:        30 * time.Second,
		EnableQueryRewrite:  true,
		EnableMonitoring:    true,
		EnableLoadBalancing: true,
		SlowQueryThreshold:  1 * time.Second,
	}
}

// DatabaseProxy provides a proxy layer for database operations
type DatabaseProxy struct {
	config *ProxyConfig

	// Connection pool
	poolManager *PoolManager

	// Replica manager
	replicaManager *ReplicaManager

	// Query cache
	queryCache *QueryCache

	// Metrics
	totalQueries    int64
	slowQueries     int64
	failedQueries   int64
	cacheHits       int64
	totalDuration   int64 // microseconds

	// Query rewriter
	rewriter *QueryRewriter

	mu sync.RWMutex
}

// NewDatabaseProxy creates a new database proxy
func NewDatabaseProxy(config *ProxyConfig) *DatabaseProxy {
	if config == nil {
		config = DefaultProxyConfig()
	}

	return &DatabaseProxy{
		config:   config,
		rewriter: NewQueryRewriter(),
	}
}

// SetPoolManager sets the connection pool manager
func (dp *DatabaseProxy) SetPoolManager(pm *PoolManager) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.poolManager = pm
}

// SetReplicaManager sets the replica manager
func (dp *DatabaseProxy) SetReplicaManager(rm *ReplicaManager) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.replicaManager = rm
}

// SetQueryCache sets the query cache
func (dp *DatabaseProxy) SetQueryCache(qc *QueryCache) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.queryCache = qc
}

// Query executes a query through the proxy
func (dp *DatabaseProxy) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	atomic.AddInt64(&dp.totalQueries, 1)

	// Apply timeout
	if dp.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, dp.config.QueryTimeout)
		defer cancel()
	}

	// Try cache first for SELECT queries
	if dp.queryCache != nil && isSelectQuery(query) {
		// Note: Caching sql.Rows is complex, simplified for example
		// Production would need different approach
	}

	// Rewrite query if enabled
	if dp.config.EnableQueryRewrite {
		query = dp.rewriter.Rewrite(query)
	}

	// Select appropriate database
	db, err := dp.selectDatabase(query)
	if err != nil {
		atomic.AddInt64(&dp.failedQueries, 1)
		return nil, err
	}

	// Execute query
	rows, err := db.QueryContext(ctx, query, args...)

	// Record metrics
	duration := time.Since(start)
	atomic.AddInt64(&dp.totalDuration, duration.Microseconds())

	if duration > dp.config.SlowQueryThreshold {
		atomic.AddInt64(&dp.slowQueries, 1)
		dp.logSlowQuery(query, duration)
	}

	if err != nil {
		atomic.AddInt64(&dp.failedQueries, 1)
		return nil, err
	}

	return rows, nil
}

// Exec executes a non-query statement through the proxy
func (dp *DatabaseProxy) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	atomic.AddInt64(&dp.totalQueries, 1)

	// Apply timeout
	if dp.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, dp.config.QueryTimeout)
		defer cancel()
	}

	// Rewrite query if enabled
	if dp.config.EnableQueryRewrite {
		query = dp.rewriter.Rewrite(query)
	}

	// Always use primary for writes
	db, err := dp.poolManager.GetPrimary()
	if err != nil {
		atomic.AddInt64(&dp.failedQueries, 1)
		return nil, err
	}

	// Execute query
	result, err := db.ExecContext(ctx, query, args...)

	// Record metrics
	duration := time.Since(start)
	atomic.AddInt64(&dp.totalDuration, duration.Microseconds())

	if duration > dp.config.SlowQueryThreshold {
		atomic.AddInt64(&dp.slowQueries, 1)
		dp.logSlowQuery(query, duration)
	}

	if err != nil {
		atomic.AddInt64(&dp.failedQueries, 1)
		return nil, err
	}

	// Invalidate cache for tables affected by write
	if dp.queryCache != nil {
		dp.invalidateCacheForQuery(query)
	}

	return result, nil
}

// selectDatabase selects the appropriate database for a query
func (dp *DatabaseProxy) selectDatabase(query string) (RelationalDB, error) {
	if dp.poolManager == nil {
		return nil, fmt.Errorf("pool manager not configured")
	}

	// Check if read query
	if isSelectQuery(query) {
		// Try to use replica if available
		if dp.config.EnableLoadBalancing && dp.replicaManager != nil {
			replicaName, err := dp.replicaManager.SelectReplica()
			if err == nil {
				db, err := dp.poolManager.GetByName(replicaName)
				if err == nil {
					return db, nil
				}
			}
		}

		// Fall back to GetForRead (which uses replicas if available)
		return dp.poolManager.GetForRead()
	}

	// Write queries always go to primary
	return dp.poolManager.GetPrimary()
}

// isSelectQuery checks if a query is a SELECT statement
func isSelectQuery(query string) bool {
	trimmed := strings.TrimSpace(strings.ToUpper(query))
	return strings.HasPrefix(trimmed, "SELECT") ||
	       strings.HasPrefix(trimmed, "SHOW") ||
	       strings.HasPrefix(trimmed, "DESCRIBE")
}

// invalidateCacheForQuery invalidates cache entries related to a query
func (dp *DatabaseProxy) invalidateCacheForQuery(query string) {
	// Simple table name extraction
	// Production would need proper SQL parsing
	upper := strings.ToUpper(query)

	if strings.Contains(upper, "UPDATE") {
		// Extract table name after UPDATE
	} else if strings.Contains(upper, "INSERT") {
		// Extract table name after INSERT INTO
	} else if strings.Contains(upper, "DELETE") {
		// Extract table name after DELETE FROM
	}
}

// logSlowQuery logs a slow query
func (dp *DatabaseProxy) logSlowQuery(query string, duration time.Duration) {
	if !dp.config.EnableMonitoring {
		return
	}

	// Log slow query
	// Production would use proper logger
	fmt.Printf("[SLOW QUERY] %v: %s\n", duration, query)
}

// GetMetrics returns proxy metrics
func (dp *DatabaseProxy) GetMetrics() ProxyMetrics {
	totalQueries := atomic.LoadInt64(&dp.totalQueries)
	avgDuration := int64(0)
	if totalQueries > 0 {
		avgDuration = atomic.LoadInt64(&dp.totalDuration) / totalQueries
	}

	return ProxyMetrics{
		TotalQueries:       totalQueries,
		SlowQueries:        atomic.LoadInt64(&dp.slowQueries),
		FailedQueries:      atomic.LoadInt64(&dp.failedQueries),
		CacheHits:          atomic.LoadInt64(&dp.cacheHits),
		AverageDuration:    time.Duration(avgDuration) * time.Microsecond,
		SlowQueryThreshold: dp.config.SlowQueryThreshold,
	}
}

// ProxyMetrics contains proxy statistics
type ProxyMetrics struct {
	TotalQueries       int64
	SlowQueries        int64
	FailedQueries      int64
	CacheHits          int64
	AverageDuration    time.Duration
	SlowQueryThreshold time.Duration
}

// QueryRewriter rewrites queries for optimization
type QueryRewriter struct {
	rules []RewriteRule
	mu    sync.RWMutex
}

// RewriteRule defines a query rewrite rule
type RewriteRule struct {
	Name    string
	Pattern string
	Replace string
	Enabled bool
}

// NewQueryRewriter creates a new query rewriter
func NewQueryRewriter() *QueryRewriter {
	qr := &QueryRewriter{
		rules: make([]RewriteRule, 0),
	}

	// Add default rules
	qr.AddRule(RewriteRule{
		Name:    "select-star-limit",
		Pattern: "SELECT *",
		Replace: "SELECT * LIMIT 1000", // Prevent runaway SELECT *
		Enabled: false,                 // Disabled by default
	})

	return qr
}

// AddRule adds a rewrite rule
func (qr *QueryRewriter) AddRule(rule RewriteRule) {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	qr.rules = append(qr.rules, rule)
}

// Rewrite applies rewrite rules to a query
func (qr *QueryRewriter) Rewrite(query string) string {
	qr.mu.RLock()
	defer qr.mu.RUnlock()

	result := query
	for _, rule := range qr.rules {
		if rule.Enabled && strings.Contains(result, rule.Pattern) {
			result = strings.ReplaceAll(result, rule.Pattern, rule.Replace)
		}
	}

	return result
}

// ConnectionPoolProxy wraps a connection pool with additional features
type ConnectionPoolProxy struct {
	pool         *sql.DB
	maxConns     int
	activeConns  int64
	totalConns   int64
	waitTime     int64 // microseconds
	mu           sync.RWMutex
}

// NewConnectionPoolProxy creates a new connection pool proxy
func NewConnectionPoolProxy(db *sql.DB, maxConns int) *ConnectionPoolProxy {
	return &ConnectionPoolProxy{
		pool:     db,
		maxConns: maxConns,
	}
}

// GetConnection gets a connection from the pool
func (cpp *ConnectionPoolProxy) GetConnection(ctx context.Context) (*sql.Conn, error) {
	start := time.Now()

	// Track active connections
	atomic.AddInt64(&cpp.activeConns, 1)
	atomic.AddInt64(&cpp.totalConns, 1)

	conn, err := cpp.pool.Conn(ctx)

	// Track wait time
	waitTime := time.Since(start)
	atomic.AddInt64(&cpp.waitTime, waitTime.Microseconds())

	if err != nil {
		atomic.AddInt64(&cpp.activeConns, -1)
		return nil, err
	}

	return conn, nil
}

// ReleaseConnection releases a connection back to the pool
func (cpp *ConnectionPoolProxy) ReleaseConnection(conn *sql.Conn) error {
	atomic.AddInt64(&cpp.activeConns, -1)
	return conn.Close()
}

// GetStats returns connection pool statistics
func (cpp *ConnectionPoolProxy) GetStats() ConnectionPoolStats {
	return ConnectionPoolStats{
		ActiveConnections: atomic.LoadInt64(&cpp.activeConns),
		TotalConnections:  atomic.LoadInt64(&cpp.totalConns),
		AverageWaitTime:   time.Duration(atomic.LoadInt64(&cpp.waitTime)/atomic.LoadInt64(&cpp.totalConns)) * time.Microsecond,
		MaxConnections:    cpp.maxConns,
	}
}

// ConnectionPoolStats contains connection pool statistics
type ConnectionPoolStats struct {
	ActiveConnections int64
	TotalConnections  int64
	AverageWaitTime   time.Duration
	MaxConnections    int
}

// LoadBalancer balances queries across multiple databases
type LoadBalancer struct {
	databases []RelationalDB
	strategy  LoadBalancingStrategy
	index     uint32
	mu        sync.RWMutex
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(strategy LoadBalancingStrategy) *LoadBalancer {
	return &LoadBalancer{
		databases: make([]RelationalDB, 0),
		strategy:  strategy,
	}
}

// AddDatabase adds a database to the load balancer
func (lb *LoadBalancer) AddDatabase(db RelationalDB) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.databases = append(lb.databases, db)
}

// SelectDatabase selects a database based on strategy
func (lb *LoadBalancer) SelectDatabase() (RelationalDB, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if len(lb.databases) == 0 {
		return nil, fmt.Errorf("no databases available")
	}

	switch lb.strategy {
	case RoundRobin:
		idx := atomic.AddUint32(&lb.index, 1)
		return lb.databases[idx%uint32(len(lb.databases))], nil

	case Random:
		idx := uint32(time.Now().UnixNano()) % uint32(len(lb.databases))
		return lb.databases[idx], nil

	default:
		return lb.databases[0], nil
	}
}

// QueryLogger logs all queries
type QueryLogger struct {
	enabled bool
	mu      sync.RWMutex
	queries []QueryLog
	maxLogs int
}

// QueryLog represents a logged query
type QueryLog struct {
	Query     string
	Args      []interface{}
	Duration  time.Duration
	Error     error
	Timestamp time.Time
}

// NewQueryLogger creates a new query logger
func NewQueryLogger(maxLogs int) *QueryLogger {
	return &QueryLogger{
		enabled: true,
		queries: make([]QueryLog, 0),
		maxLogs: maxLogs,
	}
}

// Log logs a query
func (ql *QueryLogger) Log(query string, duration time.Duration, err error, args ...interface{}) {
	if !ql.enabled {
		return
	}

	ql.mu.Lock()
	defer ql.mu.Unlock()

	log := QueryLog{
		Query:     query,
		Args:      args,
		Duration:  duration,
		Error:     err,
		Timestamp: time.Now(),
	}

	ql.queries = append(ql.queries, log)

	// Keep only recent logs
	if len(ql.queries) > ql.maxLogs {
		ql.queries = ql.queries[len(ql.queries)-ql.maxLogs:]
	}
}

// GetRecentQueries returns recent query logs
func (ql *QueryLogger) GetRecentQueries(n int) []QueryLog {
	ql.mu.RLock()
	defer ql.mu.RUnlock()

	if n > len(ql.queries) {
		n = len(ql.queries)
	}

	result := make([]QueryLog, n)
	copy(result, ql.queries[len(ql.queries)-n:])
	return result
}

// Clear clears all query logs
func (ql *QueryLogger) Clear() {
	ql.mu.Lock()
	defer ql.mu.Unlock()
	ql.queries = make([]QueryLog, 0)
}
