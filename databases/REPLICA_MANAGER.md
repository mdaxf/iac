# Database Replica Manager

**Version:** 1.0
**Status:** ✅ Production Ready
**Component:** Phase 5 - Advanced Features

## Overview

The Replica Manager provides advanced read replica support for the IAC database layer, including:

- **Intelligent Load Balancing**: Multiple strategies (round-robin, weighted, least-lag)
- **Replication Lag Monitoring**: Real-time monitoring of replica lag
- **Automatic Failover**: Automatic detection and failover for unhealthy replicas
- **Auto-Recovery**: Automatic re-activation of recovered replicas
- **Health Tracking**: Comprehensive health metrics for each replica

## Features

### 1. Load Balancing Strategies

#### Round Robin
Distributes requests evenly across all healthy replicas.

```go
config := databases.DefaultReplicaManagerConfig()
config.Strategy = databases.RoundRobin
rm := databases.NewReplicaManager(config)
```

#### Weighted Round Robin (Default)
Distributes requests based on replica capacity/weight.

```go
rm := databases.NewReplicaManager(nil)
rm.RegisterReplica("high-capacity", 5)  // Gets 5x traffic
rm.RegisterReplica("low-capacity", 1)   // Gets 1x traffic
```

#### Least Lag
Selects replica with minimum replication lag.

```go
config := databases.DefaultReplicaManagerConfig()
config.Strategy = databases.LeastLag
config.MaxReplicaLag = 10.0  // 10 seconds max
rm := databases.NewReplicaManager(config)
```

#### Random
Randomly selects from healthy replicas.

```go
config := databases.DefaultReplicaManagerConfig()
config.Strategy = databases.Random
rm := databases.NewReplicaManager(config)
```

### 2. Replication Lag Monitoring

Automatically monitors replication lag for MySQL and PostgreSQL:

#### MySQL
Uses `SHOW SLAVE STATUS` to get `Seconds_Behind_Master`.

#### PostgreSQL
Uses `pg_last_xact_replay_timestamp()` to calculate lag.

```go
config := databases.DefaultReplicaManagerConfig()
config.MaxReplicaLag = 5.0              // 5 seconds max
config.LagCheckInterval = 10 * time.Second
rm := databases.NewReplicaManager(config)

// Replicas with lag > MaxReplicaLag are excluded from selection
```

### 3. Automatic Failover

Automatically marks replicas as unhealthy after consecutive failures:

```go
config := databases.DefaultReplicaManagerConfig()
config.FailoverThreshold = 3  // 3 consecutive failures
rm := databases.NewReplicaManager(config)

// Simulate failures
rm.RecordFailure("replica-1", err)
rm.RecordFailure("replica-1", err)
rm.RecordFailure("replica-1", err)
// replica-1 is now marked inactive
```

### 4. Auto-Recovery

Automatically checks and reactivates recovered replicas:

```go
config := databases.DefaultReplicaManagerConfig()
config.EnableAutoRecovery = true
config.RecoveryCheckInterval = 30 * time.Second
rm := databases.NewReplicaManager(config)

// Failed replicas are automatically checked and reactivated
```

### 5. Health Tracking

Track comprehensive health metrics for each replica:

```go
health := rm.GetReplicaHealth()
for name, h := range health {
    fmt.Printf("Replica: %s\n", name)
    fmt.Printf("  Active: %v\n", h.Active)
    fmt.Printf("  Weight: %d (Effective: %d)\n", h.Weight, h.EffectiveWeight)
    fmt.Printf("  Error Count: %d\n", h.ErrorCount)
    fmt.Printf("  Consecutive Fails: %d\n", h.ConsecutiveFails)
    fmt.Printf("  Response Time: %v\n", h.ResponseTime)

    if h.Lag != nil {
        fmt.Printf("  Replication Lag: %.2fs\n", h.Lag.LagSeconds)
        fmt.Printf("  Lag Healthy: %v\n", h.Lag.IsHealthy)
    }
}
```

## Quick Start

### Basic Setup

```go
package main

import (
    "github.com/mdaxf/iac/databases"
)

func main() {
    // Create replica manager
    rm := databases.NewReplicaManager(nil)

    // Register replicas
    rm.RegisterReplica("replica-1", 1)
    rm.RegisterReplica("replica-2", 1)

    // Select replica for read operation
    replica, err := rm.SelectReplica()
    if err != nil {
        panic(err)
    }

    // Use the selected replica...
}
```

### Advanced Setup with Monitoring

```go
package main

import (
    "context"
    "database/sql"
    "time"

    "github.com/mdaxf/iac/databases"
)

func main() {
    // Configure replica manager
    config := &databases.ReplicaManagerConfig{
        Strategy:              databases.WeightedRoundRobin,
        MaxReplicaLag:         5.0,
        LagCheckInterval:      10 * time.Second,
        FailoverThreshold:     3,
        RecoveryCheckInterval: 30 * time.Second,
        EnableAutoRecovery:    true,
    }

    rm := databases.NewReplicaManager(config)

    // Register replicas with weights
    rm.RegisterReplica("primary-replica", 5)
    rm.RegisterReplica("secondary-replica", 2)

    // Database getter function
    dbGetter := func(name string) (*sql.DB, string, error) {
        // Return database connection by name
        // This would typically come from your connection pool
        return getDBConnection(name)
    }

    // Start background monitoring
    ctx := context.Background()
    rm.StartMonitoring(ctx, dbGetter)
    defer rm.StopMonitoring()

    // Use replica manager for read operations
    for {
        replica, err := rm.SelectReplica()
        if err != nil {
            // No healthy replicas available
            continue
        }

        // Execute read query
        db, _, _ := dbGetter(replica)
        start := time.Now()
        err = executeQuery(db)
        duration := time.Since(start)

        if err != nil {
            rm.RecordFailure(replica, err)
        } else {
            rm.RecordSuccess(replica, duration)
        }
    }
}
```

## Configuration Options

### ReplicaManagerConfig

```go
type ReplicaManagerConfig struct {
    // Strategy for load balancing across replicas
    Strategy LoadBalancingStrategy

    // MaxReplicaLag is the maximum acceptable replication lag in seconds
    // Replicas exceeding this lag are excluded from selection
    MaxReplicaLag float64

    // LagCheckInterval is how often to check replication lag
    LagCheckInterval time.Duration

    // FailoverThreshold is how many consecutive failures before marking unhealthy
    FailoverThreshold int

    // RecoveryCheckInterval is how often to check if failed replicas recovered
    RecoveryCheckInterval time.Duration

    // EnableAutoRecovery automatically marks replicas as active when they recover
    EnableAutoRecovery bool

    // PreferLocalReplica prefers replicas in the same region/zone
    PreferLocalReplica bool

    // LocalRegion is the local region/zone identifier
    LocalRegion string
}
```

### Default Configuration

```go
config := databases.DefaultReplicaManagerConfig()

// Defaults:
// - Strategy: WeightedRoundRobin
// - MaxReplicaLag: 10.0 seconds
// - LagCheckInterval: 5 seconds
// - FailoverThreshold: 3
// - RecoveryCheckInterval: 30 seconds
// - EnableAutoRecovery: true
```

## API Reference

### Core Methods

#### NewReplicaManager

```go
func NewReplicaManager(config *ReplicaManagerConfig) *ReplicaManager
```

Creates a new replica manager with the specified configuration.

#### RegisterReplica

```go
func (rm *ReplicaManager) RegisterReplica(name string, weight int)
```

Registers a replica for load balancing. Weight determines how much traffic the replica receives (higher = more traffic).

#### SelectReplica

```go
func (rm *ReplicaManager) SelectReplica() (string, error)
```

Selects a replica based on the configured strategy. Returns an error if no healthy replicas are available.

#### RecordSuccess

```go
func (rm *ReplicaManager) RecordSuccess(replicaName string, responseTime time.Duration)
```

Records a successful operation on a replica. Increases effective weight and resets failure count.

#### RecordFailure

```go
func (rm *ReplicaManager) RecordFailure(replicaName string, err error)
```

Records a failed operation on a replica. Decreases effective weight and marks as inactive if threshold is exceeded.

### Monitoring Methods

#### StartMonitoring

```go
func (rm *ReplicaManager) StartMonitoring(
    ctx context.Context,
    dbGetter func(string) (*sql.DB, string, error),
)
```

Starts background monitoring of replica health and replication lag.

#### StopMonitoring

```go
func (rm *ReplicaManager) StopMonitoring()
```

Stops background monitoring.

#### UpdateReplicaLag

```go
func (rm *ReplicaManager) UpdateReplicaLag(
    replicaName string,
    lagSeconds float64,
    err error,
)
```

Manually updates replication lag for a replica.

#### CheckReplicationLag

```go
func (rm *ReplicaManager) CheckReplicationLag(
    ctx context.Context,
    db *sql.DB,
    dbType string,
) (float64, error)
```

Checks replication lag for a specific database. Supports MySQL and PostgreSQL.

### Information Methods

#### GetReplicaHealth

```go
func (rm *ReplicaManager) GetReplicaHealth() map[string]*ReplicaHealth
```

Returns health information for all registered replicas.

#### GetHealthyReplicaCount

```go
func (rm *ReplicaManager) GetHealthyReplicaCount() int
```

Returns the number of currently healthy replicas.

#### GetStats

```go
func (rm *ReplicaManager) GetStats() ReplicaManagerStats
```

Returns statistics about the replica manager.

## Integration Examples

### With Existing PoolManager

```go
// Get database from pool manager
poolManager := dbconn.GetPoolManager()
primaryDB, _ := poolManager.GetPrimary()

// Create replica manager
rm := databases.NewReplicaManager(nil)
rm.RegisterReplica("replica-1", 1)
rm.RegisterReplica("replica-2", 1)

// For write operations: use primary
_, err := primaryDB.Exec("INSERT INTO users ...")

// For read operations: use replica
replicaName, _ := rm.SelectReplica()
replicaDB, _ := poolManager.GetByName(replicaName)
rows, err := replicaDB.Query("SELECT * FROM users ...")
```

### With Service Layer

```go
type UserService struct {
    replicaManager *databases.ReplicaManager
    poolManager    *dbconn.PoolManager
}

func (s *UserService) GetUser(id int) (*User, error) {
    // Select replica for read
    replica, err := s.replicaManager.SelectReplica()
    if err != nil {
        return nil, err
    }

    // Get database connection
    db, err := s.poolManager.GetByName(replica)
    if err != nil {
        s.replicaManager.RecordFailure(replica, err)
        return nil, err
    }

    // Execute query
    start := time.Now()
    var user User
    err = db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user)
    duration := time.Since(start)

    if err != nil {
        s.replicaManager.RecordFailure(replica, err)
        return nil, err
    }

    s.replicaManager.RecordSuccess(replica, duration)
    return &user, nil
}
```

## Best Practices

### 1. Weight Assignment

Assign weights based on replica capacity:

```go
rm.RegisterReplica("high-spec-replica", 10)   // 32GB RAM, 8 cores
rm.RegisterReplica("medium-spec-replica", 5)  // 16GB RAM, 4 cores
rm.RegisterReplica("low-spec-replica", 1)     // 8GB RAM, 2 cores
```

### 2. Lag Thresholds

Set appropriate lag thresholds for your application:

```go
// Real-time analytics: strict lag requirement
config.MaxReplicaLag = 1.0  // 1 second

// Reporting queries: relaxed lag requirement
config.MaxReplicaLag = 30.0  // 30 seconds
```

### 3. Failover Threshold

Balance between sensitivity and stability:

```go
// Aggressive failover (less stable)
config.FailoverThreshold = 1

// Balanced failover (recommended)
config.FailoverThreshold = 3

// Conservative failover (more stable)
config.FailoverThreshold = 5
```

### 4. Monitoring Intervals

Adjust based on your requirements:

```go
// Critical applications: frequent checks
config.LagCheckInterval = 5 * time.Second
config.RecoveryCheckInterval = 10 * time.Second

// Normal applications: balanced
config.LagCheckInterval = 10 * time.Second
config.RecoveryCheckInterval = 30 * time.Second

// Low-priority applications: infrequent checks
config.LagCheckInterval = 30 * time.Second
config.RecoveryCheckInterval = 60 * time.Second
```

## Performance Considerations

### Memory Usage

- Replica manager maintains health state for each replica
- Memory usage: ~1KB per replica
- Suitable for hundreds of replicas

### CPU Usage

- Selection operations: O(n) where n = number of healthy replicas
- Weighted selection: O(w) where w = total weight
- Lag monitoring: One query per replica per interval

### Lock Contention

- Read-heavy operations use read locks (RLock)
- Write operations (registration, health updates) use write locks
- Minimal contention in typical scenarios

## Troubleshooting

### All Replicas Marked Unhealthy

**Symptom:** `no healthy replicas available` error

**Causes:**
1. All replicas actually failed
2. Replication lag exceeds threshold
3. Failover threshold too aggressive

**Solutions:**
```go
// Check replica health
health := rm.GetReplicaHealth()
for name, h := range health {
    fmt.Printf("%s: active=%v, fails=%d, lag=%.2f\n",
        name, h.Active, h.ConsecutiveFails, h.Lag.LagSeconds)
}

// Increase thresholds if needed
config.FailoverThreshold = 5
config.MaxReplicaLag = 20.0
```

### Uneven Load Distribution

**Symptom:** Some replicas receive no traffic

**Causes:**
1. Weight too low
2. Replica marked unhealthy
3. Lag exceeds threshold

**Solutions:**
```go
// Check weights
health := rm.GetReplicaHealth()
for name, h := range health {
    fmt.Printf("%s: weight=%d, effective=%d\n",
        name, h.Weight, h.EffectiveWeight)
}

// Increase weight
rm.UnregisterReplica("replica-1")
rm.RegisterReplica("replica-1", 10)  // Higher weight
```

### High Lag Not Detected

**Symptom:** Replicas with high lag still receiving traffic

**Causes:**
1. Monitoring not started
2. Lag check interval too high
3. MaxReplicaLag not set

**Solutions:**
```go
// Start monitoring
dbGetter := func(name string) (*sql.DB, string, error) {
    // Return DB connection
}
rm.StartMonitoring(ctx, dbGetter)

// Set stricter lag limit
config.MaxReplicaLag = 5.0
config.LagCheckInterval = 5 * time.Second
```

## Testing

Run the replica manager tests:

```bash
cd databases
go test -v -run TestReplicaManager
```

Run all replica-related tests:

```bash
go test -v -run Replica
```

## Examples

Complete examples are available in:
- `/examples/replica_manager_example.go` - Basic usage examples
- `/databases/replica_manager_test.go` - Unit tests with examples

## License

Copyright 2023 IAC. All Rights Reserved.

Licensed under the Apache License, Version 2.0.

---

**Version:** 1.0
**Last Updated:** 2025-11-16
**Maintained By:** IAC Development Team
