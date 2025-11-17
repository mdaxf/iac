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

package metrics

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
)

// Dashboard provides HTTP endpoints for database metrics
type Dashboard struct {
	collector *dbconn.MetricsCollector
	mu        sync.RWMutex
}

// NewDashboard creates a new metrics dashboard
func NewDashboard(collector *dbconn.MetricsCollector) *Dashboard {
	return &Dashboard{
		collector: collector,
	}
}

// ServeHTTP implements http.Handler for the dashboard
func (d *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set security headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")

	switch r.URL.Path {
	case "/api/metrics":
		d.handleMetrics(w, r)
	case "/api/metrics/summary":
		d.handleSummary(w, r)
	case "/api/metrics/database":
		d.handleDatabaseMetrics(w, r)
	case "/api/health":
		d.handleHealth(w, r)
	default:
		http.NotFound(w, r)
	}
}

// MetricsResponse represents the metrics API response
type MetricsResponse struct {
	Timestamp string                       `json:"timestamp"`
	Databases map[string]*DatabaseMetrics  `json:"databases"`
}

// DatabaseMetrics represents metrics for a single database
type DatabaseMetrics struct {
	Type            string              `json:"type"`
	Status          string              `json:"status"`
	ConnectionPool  ConnectionPoolStats `json:"connection_pool"`
	QueryStats      QueryStats          `json:"query_stats"`
	Performance     PerformanceStats    `json:"performance"`
	ErrorRate       float64             `json:"error_rate"`
}

// ConnectionPoolStats represents connection pool statistics
type ConnectionPoolStats struct {
	Active      int `json:"active"`
	Idle        int `json:"idle"`
	MaxOpen     int `json:"max_open"`
	MaxIdle     int `json:"max_idle"`
	WaitCount   int `json:"wait_count"`
	WaitDuration int64 `json:"wait_duration_ms"`
}

// QueryStats represents query statistics
type QueryStats struct {
	Total       int64   `json:"total"`
	Success     int64   `json:"success"`
	Errors      int64   `json:"errors"`
	Slow        int64   `json:"slow"`
	AvgDuration float64 `json:"avg_duration_ms"`
	MinDuration float64 `json:"min_duration_ms"`
	MaxDuration float64 `json:"max_duration_ms"`
}

// PerformanceStats represents performance statistics
type PerformanceStats struct {
	QueriesPerSecond float64            `json:"queries_per_second"`
	AvgLatency       float64            `json:"avg_latency_ms"`
	P50Latency       float64            `json:"p50_latency_ms"`
	P95Latency       float64            `json:"p95_latency_ms"`
	P99Latency       float64            `json:"p99_latency_ms"`
	ByType           map[string]int64   `json:"by_type"`
}

// handleMetrics returns all metrics
func (d *Dashboard) handleMetrics(w http.ResponseWriter, r *http.Request) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	metrics := d.collectAllMetrics()

	response := MetricsResponse{
		Timestamp: time.Now().Format(time.RFC3339),
		Databases: metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSummary returns a summary of metrics
func (d *Dashboard) handleSummary(w http.ResponseWriter, r *http.Request) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	metrics := d.collectAllMetrics()

	summary := map[string]interface{}{
		"timestamp":        time.Now().Format(time.RFC3339),
		"total_databases":  len(metrics),
		"healthy_count":    d.countHealthy(metrics),
		"total_queries":    d.sumTotalQueries(metrics),
		"total_errors":     d.sumTotalErrors(metrics),
		"avg_error_rate":   d.calculateAvgErrorRate(metrics),
		"databases":        metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// handleDatabaseMetrics returns metrics for a specific database
func (d *Dashboard) handleDatabaseMetrics(w http.ResponseWriter, r *http.Request) {
	dbType := r.URL.Query().Get("type")
	if dbType == "" {
		http.Error(w, "database type required", http.StatusBadRequest)
		return
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	metrics := d.collectDatabaseMetrics(dbType)
	if metrics == nil {
		http.Error(w, "database not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"database":  dbType,
		"metrics":   metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth returns health status
func (d *Dashboard) handleHealth(w http.ResponseWriter, r *http.Request) {
	metrics := d.collectAllMetrics()

	health := map[string]interface{}{
		"status":          "healthy",
		"timestamp":       time.Now().Format(time.RFC3339),
		"databases":       len(metrics),
		"healthy_count":   d.countHealthy(metrics),
		"unhealthy_count": len(metrics) - d.countHealthy(metrics),
	}

	// Set overall status
	if d.countHealthy(metrics) < len(metrics) {
		health["status"] = "degraded"
	}
	if d.countHealthy(metrics) == 0 {
		health["status"] = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// collectAllMetrics collects metrics from all databases
func (d *Dashboard) collectAllMetrics() map[string]*DatabaseMetrics {
	if d.collector == nil {
		return make(map[string]*DatabaseMetrics)
	}

	allMetrics := d.collector.GetAllMetrics()
	result := make(map[string]*DatabaseMetrics)

	for dbType, m := range allMetrics {
		result[dbType] = &DatabaseMetrics{
			Type:   dbType,
			Status: d.getStatus(m),
			ConnectionPool: ConnectionPoolStats{
				Active:   m.ActiveConnections,
				Idle:     m.IdleConnections,
				MaxOpen:  m.MaxOpenConnections,
				MaxIdle:  m.MaxIdleConnections,
			},
			QueryStats: QueryStats{
				Total:       m.TotalQueries,
				Success:     m.TotalQueries - m.TotalErrors,
				Errors:      m.TotalErrors,
				Slow:        m.SlowQueries,
				AvgDuration: m.AvgQueryDuration,
			},
			Performance: PerformanceStats{
				QueriesPerSecond: m.QueriesPerSecond,
				AvgLatency:       m.AvgQueryDuration,
				ByType: map[string]int64{
					"select": m.SelectQueries,
					"insert": m.InsertQueries,
					"update": m.UpdateQueries,
					"delete": m.DeleteQueries,
				},
			},
			ErrorRate: d.calculateErrorRate(m),
		}
	}

	return result
}

// collectDatabaseMetrics collects metrics for a specific database
func (d *Dashboard) collectDatabaseMetrics(dbType string) *DatabaseMetrics {
	allMetrics := d.collectAllMetrics()
	return allMetrics[dbType]
}

// getStatus determines the status of a database
func (d *Dashboard) getStatus(m *dbconn.QueryMetrics) string {
	if m.TotalErrors > 0 && float64(m.TotalErrors)/float64(m.TotalQueries) > 0.1 {
		return "unhealthy"
	}
	if m.SlowQueries > 0 && float64(m.SlowQueries)/float64(m.TotalQueries) > 0.2 {
		return "degraded"
	}
	return "healthy"
}

// calculateErrorRate calculates the error rate
func (d *Dashboard) calculateErrorRate(m *dbconn.QueryMetrics) float64 {
	if m.TotalQueries == 0 {
		return 0
	}
	return float64(m.TotalErrors) / float64(m.TotalQueries) * 100
}

// countHealthy counts healthy databases
func (d *Dashboard) countHealthy(metrics map[string]*DatabaseMetrics) int {
	count := 0
	for _, m := range metrics {
		if m.Status == "healthy" {
			count++
		}
	}
	return count
}

// sumTotalQueries sums total queries across all databases
func (d *Dashboard) sumTotalQueries(metrics map[string]*DatabaseMetrics) int64 {
	total := int64(0)
	for _, m := range metrics {
		total += m.QueryStats.Total
	}
	return total
}

// sumTotalErrors sums total errors across all databases
func (d *Dashboard) sumTotalErrors(metrics map[string]*DatabaseMetrics) int64 {
	total := int64(0)
	for _, m := range metrics {
		total += m.QueryStats.Errors
	}
	return total
}

// calculateAvgErrorRate calculates average error rate
func (d *Dashboard) calculateAvgErrorRate(metrics map[string]*DatabaseMetrics) float64 {
	if len(metrics) == 0 {
		return 0
	}

	totalRate := 0.0
	for _, m := range metrics {
		totalRate += m.ErrorRate
	}

	return totalRate / float64(len(metrics))
}

// RegisterRoutes registers dashboard routes with an HTTP mux
func (d *Dashboard) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/api/metrics", d)
	mux.Handle("/api/metrics/summary", d)
	mux.Handle("/api/metrics/database", d)
	mux.Handle("/api/health", d)
}

// StartServer starts the metrics dashboard server
func (d *Dashboard) StartServer(addr string) error {
	mux := http.NewServeMux()
	d.RegisterRoutes(mux)

	// Add index page
	mux.HandleFunc("/", d.serveIndex)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server.ListenAndServe()
}

// serveIndex serves the dashboard HTML page
func (d *Dashboard) serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML))
}

// dashboardHTML is the embedded HTML for the metrics dashboard
const dashboardHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>IAC Database Metrics Dashboard</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif; background: #f5f7fa; padding: 20px; }
        .container { max-width: 1400px; margin: 0 auto; }
        h1 { color: #2c3e50; margin-bottom: 30px; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .card h3 { color: #7f8c8d; font-size: 14px; margin-bottom: 10px; text-transform: uppercase; }
        .card .value { font-size: 32px; font-weight: bold; color: #2c3e50; }
        .card .label { font-size: 12px; color: #95a5a6; margin-top: 5px; }
        .databases { display: grid; grid-template-columns: repeat(auto-fit, minmax(400px, 1fr)); gap: 20px; }
        .db-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .db-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px; }
        .db-name { font-size: 18px; font-weight: bold; color: #2c3e50; }
        .status { padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: 600; }
        .status.healthy { background: #d4edda; color: #155724; }
        .status.degraded { background: #fff3cd; color: #856404; }
        .status.unhealthy { background: #f8d7da; color: #721c24; }
        .metric-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #ecf0f1; }
        .metric-label { color: #7f8c8d; font-size: 14px; }
        .metric-value { color: #2c3e50; font-weight: 600; font-size: 14px; }
        .refresh { position: fixed; bottom: 20px; right: 20px; background: #3498db; color: white; border: none; padding: 12px 24px; border-radius: 24px; cursor: pointer; box-shadow: 0 4px 6px rgba(0,0,0,0.1); font-size: 14px; font-weight: 600; }
        .refresh:hover { background: #2980b9; }
        .loading { text-align: center; padding: 40px; color: #95a5a6; }
        .error { background: #f8d7da; color: #721c24; padding: 15px; border-radius: 8px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>IAC Database Metrics Dashboard</h1>
        <div id="error"></div>
        <div id="summary" class="summary"></div>
        <div id="databases" class="databases"></div>
        <div id="loading" class="loading">Loading...</div>
    </div>
    <button class="refresh" onclick="loadMetrics()">Refresh</button>

    <script>
        function formatNumber(num) {
            if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
            if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
            return num.toString();
        }

        function formatDuration(ms) {
            if (ms < 1) return (ms * 1000).toFixed(1) + 'µs';
            if (ms < 1000) return ms.toFixed(1) + 'ms';
            return (ms / 1000).toFixed(2) + 's';
        }

        async function loadMetrics() {
            try {
                document.getElementById('error').innerHTML = '';
                const response = await fetch('/api/metrics/summary');
                if (!response.ok) throw new Error('Failed to load metrics');

                const data = await response.json();
                document.getElementById('loading').style.display = 'none';

                // Render summary
                const summary = document.getElementById('summary');
                summary.innerHTML = \`
                    <div class="card">
                        <h3>Total Databases</h3>
                        <div class="value">\${data.total_databases}</div>
                        <div class="label">\${data.healthy_count} healthy</div>
                    </div>
                    <div class="card">
                        <h3>Total Queries</h3>
                        <div class="value">\${formatNumber(data.total_queries)}</div>
                        <div class="label">All databases</div>
                    </div>
                    <div class="card">
                        <h3>Total Errors</h3>
                        <div class="value">\${formatNumber(data.total_errors)}</div>
                        <div class="label">\${data.avg_error_rate.toFixed(2)}% error rate</div>
                    </div>
                    <div class="card">
                        <h3>Last Updated</h3>
                        <div class="value" style="font-size: 16px;">\${new Date(data.timestamp).toLocaleTimeString()}</div>
                        <div class="label">\${new Date(data.timestamp).toLocaleDateString()}</div>
                    </div>
                \`;

                // Render databases
                const databases = document.getElementById('databases');
                dbconn.innerHTML = Object.entries(data.databases).map(([type, db]) => \`
                    <div class="db-card">
                        <div class="db-header">
                            <div class="db-name">\${type.toUpperCase()}</div>
                            <div class="status \${db.status}">\${db.status}</div>
                        </div>
                        <div class="metric-row">
                            <span class="metric-label">Active Connections</span>
                            <span class="metric-value">\${db.connection_pool.active} / \${db.connection_pool.max_open}</span>
                        </div>
                        <div class="metric-row">
                            <span class="metric-label">Total Queries</span>
                            <span class="metric-value">\${formatNumber(db.query_stats.total)}</span>
                        </div>
                        <div class="metric-row">
                            <span class="metric-label">Errors</span>
                            <span class="metric-value">\${formatNumber(db.query_stats.errors)} (\${db.error_rate.toFixed(2)}%)</span>
                        </div>
                        <div class="metric-row">
                            <span class="metric-label">Slow Queries</span>
                            <span class="metric-value">\${formatNumber(db.query_stats.slow)}</span>
                        </div>
                        <div class="metric-row">
                            <span class="metric-label">Avg Duration</span>
                            <span class="metric-value">\${formatDuration(db.query_stats.avg_duration_ms)}</span>
                        </div>
                        <div class="metric-row">
                            <span class="metric-label">QPS</span>
                            <span class="metric-value">\${db.performance.queries_per_second.toFixed(2)}</span>
                        </div>
                    </div>
                \`).join('');

            } catch (error) {
                document.getElementById('error').innerHTML = \`
                    <div class="error">Error loading metrics: \${error.message}</div>
                \`;
            }
        }

        // Load metrics on page load
        loadMetrics();

        // Auto-refresh every 5 seconds
        setInterval(loadMetrics, 5000);
    </script>
</body>
</html>
`
