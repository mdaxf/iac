// Database Sharding Guide

**Version:** 1.0
**Status:** ✅ Production Ready
**Component:** Phase 5 - Advanced Features

## Overview

The Shard Manager provides horizontal database sharding for scalability, allowing you to distribute data across multiple database instances based on configurable strategies.

## Features

- **Multiple Sharding Strategies**: Hash modulo, consistent hashing, range-based, and lookup-based
- **Flexible Hash Functions**: CRC32, FNV-1, FNV-1a
- **Cross-Shard Queries**: Execute queries across all shards with parallelization
- **Consistent Hashing**: Minimize data movement when adding/removing shards
- **Range Sharding**: Perfect for time-series or alphabetically-sorted data
- **Lookup Sharding**: Ideal for multi-tenant applications
- **Statistics & Monitoring**: Track query counts and errors per shard

## Sharding Strategies

### 1. Hash Modulo (Default)

Simple hash-based sharding using modulo operation.

```go
config := databases.DefaultShardManagerConfig()
config.Strategy = databases.HashModulo
sm := databases.NewShardManager(config)
```

**Pros:**
- Simple and fast
- Even distribution
- Deterministic routing

**Cons:**
- Adding/removing shards requires re-sharding all data
- Not suitable for dynamic shard management

**Use Cases:**
- Fixed number of shards
- Simple user/product ID-based sharding
- When resharding is rare

### 2. Consistent Hashing

Uses consistent hashing with virtual nodes for better distribution.

```go
config := databases.DefaultShardManagerConfig()
config.Strategy = databases.ConsistentHash
config.VirtualNodes = 150  // More nodes = better distribution
sm := databases.NewShardManager(config)
```

**Pros:**
- Minimal data movement when adding/removing shards
- Good distribution with virtual nodes
- Scales well

**Cons:**
- Slightly more complex
- Requires more memory for hash ring

**Use Cases:**
- Dynamic shard management
- Frequently adding/removing shards
- Large-scale systems

### 3. Range Sharding

Distributes data based on value ranges.

```go
config := databases.DefaultShardManagerConfig()
config.Strategy = databases.RangeSharding
sm := databases.NewShardManager(config)

// Define ranges
sm.AddRangeMapping(0, "2020-01-01", "2021-01-01")
sm.AddRangeMapping(1, "2021-01-01", "2022-01-01")
sm.AddRangeMapping(2, "2022-01-01", "2023-01-01")
```

**Pros:**
- Efficient range queries
- Natural data organization
- Easy to archive old data

**Cons:**
- Can lead to hotspots
- Requires predefined ranges
- Uneven distribution possible

**Use Cases:**
- Time-series data (logs, events, metrics)
- Alphabetical distribution (A-H, I-P, Q-Z)
- Geographic distribution

### 4. Lookup Sharding

Uses explicit mapping table for shard assignment.

```go
config := databases.DefaultShardManagerConfig()
config.Strategy = databases.LookupSharding
sm := databases.NewShardManager(config)

// Map specific keys to shards
sm.AddLookupMapping("tenant:acme", 0)
sm.AddLookupMapping("tenant:globex", 1)
```

**Pros:**
- Complete control over data placement
- Can implement custom logic
- Easy tenant isolation

**Cons:**
- Requires lookup table maintenance
- Not suitable for large key spaces
- Additional storage overhead

**Use Cases:**
- Multi-tenant SaaS applications
- VIP customer isolation
- Compliance requirements

## Quick Start

### Basic Setup

```go
package main

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/mdaxf/iac/databases"
)

func main() {
    // Create shard manager
    sm := databases.NewShardManager(nil)

    // Add shards (in production, these would be separate DB servers)
    for i := 0; i < 3; i++ {
        db, err := sql.Open("postgres",
            fmt.Sprintf("host=shard%d.example.com ...", i))
        if err != nil {
            panic(err)
        }

        shard := &databases.Shard{
            ID:     i,
            Name:   fmt.Sprintf("shard-%d", i),
            DB:     db,
            DBType: "postgres",
            Active: true,
        }

        sm.AddShard(shard)
    }

    // Use sharding
    ctx := context.Background()

    // Insert data - routed to appropriate shard
    userID := "user:12345"
    rows, err := sm.ExecuteOnShard(ctx, userID,
        "INSERT INTO users (id, name) VALUES ($1, $2)",
        12345, "Alice")

    // Query data - routes to same shard
    rows, err = sm.ExecuteOnShard(ctx, userID,
        "SELECT * FROM users WHERE id = $1", 12345)
}
```

### Consistent Hashing Setup

```go
config := &databases.ShardManagerConfig{
    Strategy:             databases.ConsistentHash,
    HashFunc:             databases.CRC32Hash,
    VirtualNodes:         150,
    EnableCrossShard:     true,
    MaxCrossShardQueries: 10,
    CrossShardTimeout:    30,
}

sm := databases.NewShardManager(config)

// Add shards
for i := 0; i < 4; i++ {
    db, _ := sql.Open("postgres", ...)
    shard := &databases.Shard{
        ID:     i,
        Name:   fmt.Sprintf("shard-%d", i),
        DB:     db,
        Active: true,
    }
    sm.AddShard(shard)
}

// Query routing
shardKey := "user:67890"
shard, err := sm.GetShard(shardKey)
if err != nil {
    // Handle error
}

// Use the selected shard
rows, err := shard.DB.Query("SELECT * FROM users WHERE ...")
```

## API Reference

### Core Methods

#### NewShardManager

```go
func NewShardManager(config *ShardManagerConfig) *ShardManager
```

Creates a new shard manager with the specified configuration.

#### AddShard

```go
func (sm *ShardManager) AddShard(shard *Shard) error
```

Adds a shard to the manager. For consistent hashing, this updates the hash ring.

#### RemoveShard

```go
func (sm *ShardManager) RemoveShard(shardID int) error
```

Removes a shard from the manager.

#### GetShard

```go
func (sm *ShardManager) GetShard(shardKey string) (*Shard, error)
```

Returns the appropriate shard for a given shard key based on the configured strategy.

#### GetAllShards

```go
func (sm *ShardManager) GetAllShards() []*Shard
```

Returns all active shards.

### Query Methods

#### ExecuteOnShard

```go
func (sm *ShardManager) ExecuteOnShard(
    ctx context.Context,
    shardKey string,
    query string,
    args ...interface{},
) (*sql.Rows, error)
```

Executes a query on the appropriate shard for the given key.

#### ExecuteOnAllShards

```go
func (sm *ShardManager) ExecuteOnAllShards(
    ctx context.Context,
    query string,
    args ...interface{},
) ([]*ShardQueryResult, error)
```

Executes a query on all shards in parallel (cross-shard query).

#### ExecuteOnShardTx

```go
func (sm *ShardManager) ExecuteOnShardTx(
    ctx context.Context,
    shardKey string,
    fn func(*sql.Tx) error,
) error
```

Executes operations within a transaction on the appropriate shard.

### Configuration Methods

#### AddRangeMapping

```go
func (sm *ShardManager) AddRangeMapping(
    shardID int,
    minValue, maxValue interface{},
)
```

Adds a range mapping for range-based sharding.

#### AddLookupMapping

```go
func (sm *ShardManager) AddLookupMapping(key string, shardID int)
```

Adds a lookup mapping for lookup-based sharding.

### Statistics Methods

#### GetShardStats

```go
func (sm *ShardManager) GetShardStats() []ShardStats
```

Returns statistics for all shards.

#### ShardDistribution

```go
func (sm *ShardManager) ShardDistribution(keys []string) map[int]int
```

Calculates the distribution of keys across shards.

## Configuration

### ShardManagerConfig

```go
type ShardManagerConfig struct {
    // Strategy for sharding
    Strategy ShardingStrategy

    // HashFunc for hash-based strategies
    HashFunc HashFunction

    // VirtualNodes for consistent hashing (default: 150)
    VirtualNodes int

    // EnableCrossShard allows cross-shard queries
    EnableCrossShard bool

    // MaxCrossShardQueries limits concurrent cross-shard queries
    MaxCrossShardQueries int

    // CrossShardTimeout for cross-shard query timeout (seconds)
    CrossShardTimeout int
}
```

### Shard

```go
type Shard struct {
    ID       int           // Unique shard ID
    Name     string        // Shard name
    DB       *sql.DB       // Database connection
    DBType   string        // Database type (postgres, mysql, etc.)
    Active   bool          // Is shard active?
    Weight   int           // Shard weight (for future use)
    Region   string        // Geographic region

    // For range sharding
    MinRange interface{}
    MaxRange interface{}

    // Statistics
    QueryCount int64
    ErrorCount int64
}
```

## Advanced Usage

### Cross-Shard Queries

Execute a query across all shards and aggregate results:

```go
ctx := context.Background()
results, err := sm.ExecuteOnAllShards(ctx,
    "SELECT COUNT(*) as count FROM users WHERE active = $1", true)

if err != nil {
    return err
}

totalCount := 0
for _, result := range results {
    if result.Error != nil {
        log.Printf("Shard %s error: %v", result.ShardName, result.Error)
        continue
    }

    for result.Rows.Next() {
        var count int
        result.Rows.Scan(&count)
        totalCount += count
    }
    result.Rows.Close()
}

fmt.Printf("Total active users: %d\n", totalCount)
```

### Transaction on Shard

```go
userID := "user:12345"
err := sm.ExecuteOnShardTx(ctx, userID, func(tx *sql.Tx) error {
    // All operations happen on the same shard in a transaction
    _, err := tx.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2",
        100, 12345)
    if err != nil {
        return err
    }

    _, err = tx.Exec("INSERT INTO transactions (user_id, amount) VALUES ($1, $2)",
        12345, -100)
    if err != nil {
        return err
    }

    return nil
})
```

### Shard Key Selection

Choosing the right shard key is critical:

```go
// Good shard keys (high cardinality, even distribution)
userID := "user:12345"           // User ID
orderID := "order:67890"         // Order ID
sessionID := "session:abc123"    // Session ID

// Poor shard keys (low cardinality, uneven distribution)
status := "active"               // Few values
region := "us-east-1"            // Geographic clustering
```

### Monitoring Shard Health

```go
stats := sm.GetShardStats()
for _, stat := range stats {
    if !stat.Active {
        log.Printf("WARNING: Shard %s is inactive", stat.ShardName)
    }

    errorRate := float64(stat.ErrorCount) / float64(stat.QueryCount)
    if errorRate > 0.01 {  // > 1% error rate
        log.Printf("WARNING: Shard %s has high error rate: %.2f%%",
            stat.ShardName, errorRate*100)
    }
}
```

## Best Practices

### 1. Shard Key Selection

- **High Cardinality**: Choose keys with many unique values
- **Even Distribution**: Avoid keys that cluster on few values
- **Immutable**: Don't use keys that change (changing requires data movement)
- **Predictable**: Key should be derivable from application context

```go
// Good examples
"user:12345"           // User ID - high cardinality, immutable
"order:67890"          // Order ID - high cardinality, immutable
"product:ABC123"       // Product SKU - high cardinality, immutable

// Bad examples
"status:active"        // Low cardinality
"created_at:2023"      // Can cluster
"user:email@domain"    // Email can change
```

### 2. Number of Shards

Start with a reasonable number and plan for growth:

```go
// Small application (< 1M records)
initialShards := 4

// Medium application (1M - 100M records)
initialShards := 16

// Large application (> 100M records)
initialShards := 64

// Use consistent hashing for easier scaling
config.Strategy = databases.ConsistentHash
config.VirtualNodes = 150
```

### 3. Cross-Shard Query Limits

Limit concurrent cross-shard queries to avoid resource exhaustion:

```go
config := &databases.ShardManagerConfig{
    EnableCrossShard:     true,
    MaxCrossShardQueries: 10,  // Max 10 concurrent shard queries
    CrossShardTimeout:    30,  // 30 second timeout
}
```

### 4. Error Handling

Always handle shard-level errors:

```go
results, err := sm.ExecuteOnAllShards(ctx, query)
if err != nil {
    return fmt.Errorf("cross-shard query failed: %w", err)
}

successCount := 0
for _, result := range results {
    if result.Error != nil {
        // Log but continue - partial results may be acceptable
        log.Printf("Shard %s failed: %v", result.ShardName, result.Error)
        continue
    }
    successCount++
    // Process result.Rows
}

if successCount == 0 {
    return fmt.Errorf("all shards failed")
}
```

### 5. Monitoring

Monitor shard distribution and performance:

```go
// Check distribution
keys := getAllUserIDs()  // Your function to get all keys
distribution := sm.ShardDistribution(keys)

for shardID, count := range distribution {
    percentage := float64(count) / float64(len(keys)) * 100
    fmt.Printf("Shard %d: %d keys (%.1f%%)\n", shardID, count, percentage)

    // Alert if distribution is very uneven
    if percentage > 40 || percentage < 10 {
        log.Printf("WARNING: Uneven distribution on shard %d", shardID)
    }
}
```

## Common Patterns

### Multi-Tenant SaaS

Use lookup sharding for tenant isolation:

```go
config := &databases.ShardManagerConfig{
    Strategy: databases.LookupSharding,
}
sm := databases.NewShardManager(config)

// Map tenants to shards
sm.AddLookupMapping("tenant:acme", 0)
sm.AddLookupMapping("tenant:globex", 0)
sm.AddLookupMapping("tenant:stark", 1)

// Query always routes to correct shard
tenantID := "tenant:acme"
shard, _ := sm.GetShard(tenantID)
rows, _ := shard.DB.Query("SELECT * FROM users WHERE tenant_id = $1", tenantID)
```

### Time-Series Data

Use range sharding for time-based data:

```go
config := &databases.ShardManagerConfig{
    Strategy: databases.RangeSharding,
}
sm := databases.NewShardManager(config)

// Shard by year
sm.AddRangeMapping(0, "2020-01-01", "2021-01-01")
sm.AddRangeMapping(1, "2021-01-01", "2022-01-01")
sm.AddRangeMapping(2, "2022-01-01", "2023-01-01")

// Queries automatically route to correct shard
logDate := "2021-06-15"
shard, _ := sm.GetShard(logDate)
rows, _ := shard.DB.Query("SELECT * FROM logs WHERE date = $1", logDate)
```

### User/Product Sharding

Use consistent hashing for user or product data:

```go
config := &databases.ShardManagerConfig{
    Strategy:     databases.ConsistentHash,
    VirtualNodes: 150,
}
sm := databases.NewShardManager(config)

// Add shards
for i := 0; i < 16; i++ {
    shard := createShard(i)  // Your shard creation function
    sm.AddShard(shard)
}

// User data automatically distributed
userID := "user:12345"
shard, _ := sm.GetShard(userID)
shard.DB.Query("SELECT * FROM users WHERE id = $1", 12345)
```

## Performance Considerations

### Hash Function Selection

Choose appropriate hash function for your use case:

```go
// CRC32 - Fast, good distribution
config.HashFunc = databases.CRC32Hash  // Recommended for most cases

// FNV-1a - Alternative hash, slightly slower
config.HashFunc = databases.FNV1aHash
```

### Virtual Nodes

More virtual nodes = better distribution but more memory:

```go
// Small number of shards (< 10)
config.VirtualNodes = 100

// Medium number of shards (10-50)
config.VirtualNodes = 150  // Default

// Large number of shards (> 50)
config.VirtualNodes = 200
```

### Cross-Shard Query Optimization

Limit and parallelize cross-shard queries:

```go
config := &databases.ShardManagerConfig{
    MaxCrossShardQueries: 10,  // Max concurrent queries per cross-shard operation
    CrossShardTimeout:    30,   // Timeout per shard
}
```

## Troubleshooting

### Uneven Distribution

**Problem:** Some shards have much more data than others

**Solutions:**
1. Use consistent hashing instead of hash modulo
2. Increase virtual nodes for consistent hashing
3. Check if shard key has good cardinality
4. Consider using different shard key

```go
// Check distribution
distribution := sm.ShardDistribution(allKeys)
for shardID, count := range distribution {
    percentage := float64(count) / float64(len(allKeys)) * 100
    if percentage > 35 || percentage < 15 {
        log.Printf("Shard %d has uneven distribution: %.1f%%",
            shardID, percentage)
    }
}
```

### Cross-Shard Query Timeout

**Problem:** Cross-shard queries timing out

**Solutions:**
1. Increase timeout
2. Reduce number of concurrent shard queries
3. Optimize queries on each shard
4. Consider if cross-shard query is necessary

```go
config.CrossShardTimeout = 60  // Increase to 60 seconds
config.MaxCrossShardQueries = 5  // Reduce concurrency
```

### Hotspot Shards

**Problem:** One shard receiving disproportionate traffic

**Solutions:**
1. Check shard key for clustering
2. Use consistent hashing
3. Split hot shard into multiple shards
4. Consider different sharding strategy

```go
stats := sm.GetShardStats()
for _, stat := range stats {
    qps := float64(stat.QueryCount) / elapsedSeconds
    if qps > averageQPS * 2 {
        log.Printf("Hotspot detected on shard %s: %.1f QPS",
            stat.ShardName, qps)
    }
}
```

## Testing

Run shard manager tests:

```bash
cd databases
go test -v -run TestShard
```

Run all tests:

```bash
go test -v ./databases/...
```

## Examples

Complete examples available in:
- `/examples/shard_manager_example.go` - Comprehensive usage examples
- `/databases/shard_manager_test.go` - Unit tests with examples

## License

Copyright 2023 IAC. All Rights Reserved.

Licensed under the Apache License, Version 2.0.

---

**Version:** 1.0
**Last Updated:** 2025-11-16
**Maintained By:** IAC Development Team
