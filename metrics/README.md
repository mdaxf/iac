# IAC Database Metrics Dashboard

A real-time web-based dashboard for monitoring database performance metrics across all supported database types.

## Features

- **Real-time Metrics** - Auto-refreshing dashboard with 5-second updates
- **Multi-Database Support** - Monitor MySQL, PostgreSQL, MSSQL, Oracle, MongoDB simultaneously
- **REST API** - JSON API for programmatic access
- **Health Monitoring** - Overall system health and individual database status
- **Performance Metrics** - Queries per second, latency, error rates
- **Connection Pool Stats** - Active/idle connections, pool utilization

## Quick Start

### 1. Start the Dashboard

```go
package main

import (
    "log"

    "github.com/mdaxf/iac/databases"
    "github.com/mdaxf/iac/metrics"
)

func main() {
    // Create metrics collector
    collector := databases.NewMetricsCollector()

    // Create dashboard
    dashboard := metrics.NewDashboard(collector)

    // Start server on port 8080
    log.Println("Starting metrics dashboard on http://localhost:8080")
    if err := dashboard.StartServer(":8080"); err != nil {
        log.Fatal(err)
    }
}
```

### 2. Access the Dashboard

Open your browser to `http://localhost:8080`

The dashboard will display:
- Total number of databases
- Total queries across all databases
- Total errors and error rates
- Real-time metrics for each database

## API Endpoints

### GET /api/metrics

Returns all metrics for all databases.

**Response:**
```json
{
  "timestamp": "2025-11-16T12:34:56Z",
  "databases": {
    "mysql": {
      "type": "mysql",
      "status": "healthy",
      "connection_pool": {
        "active": 5,
        "idle": 10,
        "max_open": 15,
        "max_idle": 10
      },
      "query_stats": {
        "total": 12345,
        "success": 12300,
        "errors": 45,
        "slow": 23,
        "avg_duration_ms": 45.3
      },
      "performance": {
        "queries_per_second": 123.45,
        "avg_latency_ms": 45.3,
        "by_type": {
          "select": 10000,
          "insert": 1500,
          "update": 500,
          "delete": 345
        }
      },
      "error_rate": 0.36
    }
  }
}
```

### GET /api/metrics/summary

Returns a summary of all metrics.

**Response:**
```json
{
  "timestamp": "2025-11-16T12:34:56Z",
  "total_databases": 3,
  "healthy_count": 3,
  "total_queries": 45678,
  "total_errors": 123,
  "avg_error_rate": 0.27,
  "databases": { ... }
}
```

### GET /api/metrics/database?type=mysql

Returns metrics for a specific database.

**Parameters:**
- `type` - Database type (mysql, postgres, mssql, oracle, mongodb)

**Response:**
```json
{
  "timestamp": "2025-11-16T12:34:56Z",
  "database": "mysql",
  "metrics": {
    "type": "mysql",
    "status": "healthy",
    ...
  }
}
```

### GET /api/health

Returns overall system health status.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-16T12:34:56Z",
  "databases": 3,
  "healthy_count": 3,
  "unhealthy_count": 0
}
```

**Status Codes:**
- `200 OK` - System is healthy
- `503 Service Unavailable` - All databases are unhealthy

## Database Status

Each database can have one of three statuses:

- **healthy** - Error rate < 10%, slow query rate < 20%
- **degraded** - Slow query rate > 20%
- **unhealthy** - Error rate > 10%

## Integration

### With Database Pool Manager

```go
import (
    "github.com/mdaxf/iac/databases"
    "github.com/mdaxf/iac/dbinitializer"
    "github.com/mdaxf/iac/metrics"
)

func setupMetrics() error {
    // Initialize databases
    dbInit := dbinitializer.NewDatabaseInitializer()
    if err := dbInit.InitializeFromEnvironment(); err != nil {
        return err
    }

    poolManager := dbInit.GetPoolManager()

    // Create metrics collector
    collector := databases.NewMetricsCollector()

    // Record metrics for database operations
    db, _ := poolManager.GetPrimary("mysql")

    start := time.Now()
    rows, err := db.Query("SELECT * FROM users")
    duration := time.Since(start)

    // Record the query
    collector.RecordQuery("mysql", "SELECT", duration, err)

    // Update connection pool stats
    stats := db.Stats() // assuming db has a Stats() method
    collector.UpdateConnectionPool("mysql",
        stats.InUse,
        stats.Idle,
        stats.MaxOpenConnections,
        stats.MaxIdleConnections)

    // Start dashboard
    dashboard := metrics.NewDashboard(collector)
    go dashboard.StartServer(":8080")

    return nil
}
```

### Recording Custom Metrics

```go
// Record a successful query
collector.RecordQuery("postgres", "INSERT", 50*time.Millisecond, nil)

// Record a failed query
collector.RecordQuery("postgres", "SELECT", 100*time.Millisecond, errors.New("connection lost"))

// Update connection pool
collector.UpdateConnectionPool("postgres", 8, 7, 15, 10)
```

### Resetting Metrics

```go
// Reset metrics for a specific database
collector.ResetMetrics("mysql")

// Set custom slow query threshold (default: 1 second)
collector.SetSlowQueryThreshold(500 * time.Millisecond)
```

## Configuration

### Slow Query Threshold

Queries exceeding this threshold are counted as "slow queries":

```go
collector.SetSlowQueryThreshold(500 * time.Millisecond)
```

### Server Timeouts

The dashboard server has the following default timeouts:
- Read Timeout: 10 seconds
- Write Timeout: 10 seconds
- Idle Timeout: 60 seconds

### Security Headers

The dashboard automatically sets security headers:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Content-Security-Policy: default-src 'self'`

## Monitoring Best Practices

### 1. Set Appropriate Thresholds

```go
// For fast databases (local/LAN)
collector.SetSlowQueryThreshold(100 * time.Millisecond)

// For remote databases (WAN)
collector.SetSlowQueryThreshold(1 * time.Second)
```

### 2. Monitor Error Rates

```go
metrics := collector.GetMetrics("mysql")
errorRate := float64(metrics.TotalErrors) / float64(metrics.TotalQueries) * 100

if errorRate > 5.0 {
    // Alert: High error rate
}
```

### 3. Track Slow Queries

```go
metrics := collector.GetMetrics("postgres")
slowRate := float64(metrics.SlowQueries) / float64(metrics.TotalQueries) * 100

if slowRate > 10.0 {
    // Alert: Too many slow queries
}
```

### 4. Monitor Connection Pool

```go
metrics := collector.GetMetrics("mysql")
utilization := float64(metrics.ActiveConnections) / float64(metrics.MaxOpenConnections) * 100

if utilization > 80.0 {
    // Warning: High pool utilization
    // Consider increasing max connections
}
```

## Example: Complete Setup

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/mdaxf/iac/databases"
    "github.com/mdaxf/iac/dbinitializer"
    "github.com/mdaxf/iac/metrics"
)

func main() {
    // Initialize databases
    dbInit := dbinitializer.NewDatabaseInitializer()
    if err := dbInit.InitializeFromEnvironment(); err != nil {
        log.Fatal(err)
    }

    poolManager := dbInit.GetPoolManager()

    // Create metrics collector
    collector := databases.NewMetricsCollector()
    collector.SetSlowQueryThreshold(500 * time.Millisecond)

    // Start metrics dashboard
    dashboard := metrics.NewDashboard(collector)
    go func() {
        log.Println("Starting metrics dashboard on :8080")
        if err := dashboard.StartServer(":8080"); err != nil {
            log.Printf("Dashboard error: %v", err)
        }
    }()

    // Start periodic metrics collection
    go collectMetricsPeriodically(poolManager, collector)

    // Your application logic here
    select {}
}

func collectMetricsPeriodically(poolManager *databases.PoolManager, collector *databases.MetricsCollector) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        databases := poolManager.GetAllDatabases()

        for _, dbType := range databases {
            db, err := poolManager.GetPrimary(dbType)
            if err != nil {
                continue
            }

            // Ping and record
            start := time.Now()
            err = db.Ping()
            duration := time.Since(start)

            collector.RecordQuery(dbType, "PING", duration, err)

            db.Close()
        }
    }
}
```

## Troubleshooting

### Dashboard Not Loading

**Problem:** Dashboard shows blank page or loading indefinitely

**Solutions:**
1. Check that metrics collector is properly initialized
2. Verify at least one database is configured
3. Check browser console for errors
4. Ensure API endpoints are accessible

### No Metrics Showing

**Problem:** Dashboard loads but shows no data

**Solutions:**
1. Verify `RecordQuery()` is being called
2. Check that database queries are actually executing
3. Ensure metrics collector is passed to dashboard
4. Check for JavaScript errors in browser console

### High Memory Usage

**Problem:** Dashboard consuming too much memory

**Solutions:**
1. Reduce auto-refresh interval (modify JavaScript in dashboard.go)
2. Limit number of concurrent connections to dashboard
3. Reset metrics periodically for databases not in active use

## Performance

The metrics dashboard is designed for low overhead:
- Metrics collection: < 1µs per query
- Memory per database: ~200 bytes
- Dashboard rendering: Client-side (no server load)
- API response time: < 10ms typically

## License

Copyright 2023 IAC. All Rights Reserved.

Licensed under the Apache License, Version 2.0.
