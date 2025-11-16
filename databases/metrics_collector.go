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
	"sync"
	"time"
)

// MetricsCollector collects and aggregates database metrics for the dashboard
type MetricsCollector struct {
	metrics            map[string]*QueryMetrics
	slowQueryThreshold time.Duration
	mu                 sync.RWMutex
	startTime          time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:            make(map[string]*QueryMetrics),
		slowQueryThreshold: 1 * time.Second,
		startTime:          time.Now(),
	}
}

// RecordQuery records a query execution
func (mc *MetricsCollector) RecordQuery(dbType string, queryType string, duration time.Duration, err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	metrics := mc.getOrCreateMetrics(dbType)

	metrics.TotalQueries++
	metrics.TotalDuration += duration

	if err != nil {
		metrics.TotalErrors++
	}

	if duration > mc.slowQueryThreshold {
		metrics.SlowQueries++
	}

	// Update query type counters
	switch queryType {
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

	// Update average duration
	if metrics.TotalQueries > 0 {
		metrics.AvgQueryDuration = float64(metrics.TotalDuration.Milliseconds()) / float64(metrics.TotalQueries)
	}

	// Update queries per second
	elapsed := time.Since(mc.startTime).Seconds()
	if elapsed > 0 {
		metrics.QueriesPerSecond = float64(metrics.TotalQueries) / elapsed
	}
}

// UpdateConnectionPool updates connection pool metrics
func (mc *MetricsCollector) UpdateConnectionPool(dbType string, active, idle, maxOpen, maxIdle int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	metrics := mc.getOrCreateMetrics(dbType)
	metrics.ActiveConnections = active
	metrics.IdleConnections = idle
	metrics.MaxOpenConnections = maxOpen
	metrics.MaxIdleConnections = maxIdle
}

// GetMetrics returns metrics for a specific database
func (mc *MetricsCollector) GetMetrics(dbType string) *QueryMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics := mc.metrics[dbType]
	if metrics == nil {
		return &QueryMetrics{}
	}

	// Return a copy to prevent race conditions
	copy := *metrics
	return &copy
}

// GetAllMetrics returns metrics for all databases
func (mc *MetricsCollector) GetAllMetrics() map[string]*QueryMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]*QueryMetrics)
	for dbType, metrics := range mc.metrics {
		copy := *metrics
		result[dbType] = &copy
	}

	return result
}

// ResetMetrics resets metrics for a specific database
func (mc *MetricsCollector) ResetMetrics(dbType string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if metrics, exists := mc.metrics[dbType]; exists {
		metrics.TotalQueries = 0
		metrics.TotalErrors = 0
		metrics.SlowQueries = 0
		metrics.TotalDuration = 0
		metrics.SelectQueries = 0
		metrics.InsertQueries = 0
		metrics.UpdateQueries = 0
		metrics.DeleteQueries = 0
		metrics.OtherQueries = 0
		metrics.AvgQueryDuration = 0
		metrics.QueriesPerSecond = 0
	}
}

// SetSlowQueryThreshold sets the threshold for slow query detection
func (mc *MetricsCollector) SetSlowQueryThreshold(threshold time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.slowQueryThreshold = threshold
}

// getOrCreateMetrics gets or creates metrics for a database type
func (mc *MetricsCollector) getOrCreateMetrics(dbType string) *QueryMetrics {
	metrics, exists := mc.metrics[dbType]
	if !exists {
		metrics = &QueryMetrics{
			SlowQueryThreshold: mc.slowQueryThreshold,
		}
		mc.metrics[dbType] = metrics
	}
	return metrics
}

// QueryMetrics tracks query execution metrics (extended for dashboard)
type QueryMetrics struct {
	TotalQueries     int64
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

	// Connection pool stats
	ActiveConnections  int
	IdleConnections    int
	MaxOpenConnections int
	MaxIdleConnections int

	// Performance metrics
	AvgQueryDuration float64 // in milliseconds
	QueriesPerSecond float64
}
