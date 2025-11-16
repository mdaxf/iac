# Query Caching Layer

**Version:** 1.0
**Status:** ✅ Production Ready
**Component:** Phase 5 - Advanced Features

## Overview

The Query Caching Layer provides automatic caching of database query results to reduce database load and improve application performance.

## Features

- **Flexible Cache Backends**: Memory cache (built-in) or Redis (external)
- **Automatic TTL Management**: Configurable time-to-live per query
- **Cache Invalidation**: Manual, pattern-based, or table-based
- **Hit/Miss Metrics**: Track cache performance
- **Size Limits**: Prevent cache overflow
- **GetOrSet Pattern**: Simplify cache-aside pattern
- **Pattern Matching**: Invalidate multiple entries at once

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "github.com/mdaxf/iac/databases"
)

func main() {
    // Create cache backend
    backend := databases.NewMemoryCache(10 * 1024 * 1024) // 10MB

    // Create query cache
    qc := databases.NewQueryCache(nil, backend)

    ctx := context.Background()

    // Cache a query result
    query := "SELECT * FROM users WHERE id = ?"
    result := []byte(`{"id":1,"name":"Alice"}`)
    qc.Set(ctx, query, result, 1)

    // Retrieve from cache
    cached, err := qc.Get(ctx, query, 1)
    if err == databases.ErrCacheMiss {
        // Cache miss - query database
    } else {
        // Use cached result
    }
}
```

### GetOrSet Pattern

```go
// Simplify cache-aside pattern
data, err := qc.GetOrSet(ctx, query, func() ([]byte, error) {
    // This function is only called on cache miss
    return queryDatabase()
}, args...)
```

## Configuration

### QueryCacheConfig

```go
config := &databases.QueryCacheConfig{
    // Enable/disable caching
    Enabled: true,

    // Default TTL for cached entries
    DefaultTTL: 5 * time.Minute,

    // Maximum size of cached values (bytes)
    MaxCacheSize: 1024 * 1024, // 1MB

    // Prefix for cache keys
    CacheKeyPrefix: "iac:query:",

    // Enable metrics tracking
    EnableMetrics: true,

    // Invalidation rules (optional)
    InvalidationRules: []databases.InvalidationRule{
        {
            Pattern: "*users*",
            TTL:     1 * time.Minute, // Shorter TTL for user queries
        },
    },
}

qc := databases.NewQueryCache(config, backend)
```

### Default Configuration

```go
// Use default configuration
qc := databases.NewQueryCache(nil, backend)

// Defaults:
// - Enabled: true
// - DefaultTTL: 5 minutes
// - MaxCacheSize: 1MB
// - CacheKeyPrefix: "iac:query:"
// - EnableMetrics: true
```

## Cache Backends

### Memory Cache

Built-in in-memory cache implementation:

```go
// Create memory cache with max size
backend := databases.NewMemoryCache(10 * 1024 * 1024) // 10MB

// Features:
// - Automatic LRU eviction
// - Automatic expired entry cleanup
// - No external dependencies
```

### Redis Cache (Future)

External Redis cache for distributed caching:

```go
// TODO: Redis backend implementation
// backend := databases.NewRedisCache("localhost:6379")
```

## API Reference

### Core Methods

#### NewQueryCache

```go
func NewQueryCache(
    config *QueryCacheConfig,
    backend CacheBackend,
) *QueryCache
```

Creates a new query cache with the specified configuration and backend.

#### Get

```go
func (qc *QueryCache) Get(
    ctx context.Context,
    query string,
    args ...interface{},
) ([]byte, error)
```

Retrieves a cached query result. Returns `ErrCacheMiss` if not found.

#### Set

```go
func (qc *QueryCache) Set(
    ctx context.Context,
    query string,
    result []byte,
    args ...interface{},
) error
```

Stores a query result in cache with configured TTL.

#### GetOrSet

```go
func (qc *QueryCache) GetOrSet(
    ctx context.Context,
    query string,
    fetcher func() ([]byte, error),
    args ...interface{},
) ([]byte, error)
```

Gets from cache or executes fetcher function on cache miss. Automatically caches the result.

### Invalidation Methods

#### Invalidate

```go
func (qc *QueryCache) Invalidate(
    ctx context.Context,
    query string,
    args ...interface{},
) error
```

Invalidates a specific cache entry.

#### InvalidatePattern

```go
func (qc *QueryCache) InvalidatePattern(
    ctx context.Context,
    pattern string,
) error
```

Invalidates all cache entries matching a pattern.

#### InvalidateTable

```go
func (qc *QueryCache) InvalidateTable(
    ctx context.Context,
    tableName string,
) error
```

Invalidates all queries related to a table.

#### Clear

```go
func (qc *QueryCache) Clear(ctx context.Context) error
```

Removes all cached entries.

### Metrics Methods

#### GetMetrics

```go
func (qc *QueryCache) GetMetrics() CacheMetrics
```

Returns cache statistics including hits, misses, and hit rate.

#### ResetMetrics

```go
func (qc *QueryCache) ResetMetrics()
```

Resets all metrics counters to zero.

## Usage Patterns

### 1. Cache-Aside Pattern

Manual cache management:

```go
func getUser(userID int) (*User, error) {
    query := "SELECT * FROM users WHERE id = ?"

    // Try cache first
    data, err := qc.Get(ctx, query, userID)
    if err == nil {
        var user User
        json.Unmarshal(data, &user)
        return &user, nil
    }

    // Cache miss - query database
    user, err := queryDatabase(userID)
    if err != nil {
        return nil, err
    }

    // Cache the result
    userJSON, _ := json.Marshal(user)
    qc.Set(ctx, query, userJSON, userID)

    return user, nil
}
```

### 2. GetOrSet Pattern (Recommended)

Simplified cache management:

```go
func getUser(userID int) (*User, error) {
    query := "SELECT * FROM users WHERE id = ?"

    data, err := qc.GetOrSet(ctx, query, func() ([]byte, error) {
        // Only executed on cache miss
        user, err := queryDatabase(userID)
        if err != nil {
            return nil, err
        }
        return json.Marshal(user)
    }, userID)

    if err != nil {
        return nil, err
    }

    var user User
    json.Unmarshal(data, &user)
    return &user, nil
}
```

### 3. Write-Through Pattern

Cache on write:

```go
func updateUser(user *User) error {
    query := "UPDATE users SET name = ? WHERE id = ?"

    // Update database
    err := db.Exec(query, user.Name, user.ID)
    if err != nil {
        return err
    }

    // Invalidate cache
    qc.Invalidate(ctx, "SELECT * FROM users WHERE id = ?", user.ID)

    // Or update cache directly
    userJSON, _ := json.Marshal(user)
    qc.Set(ctx, "SELECT * FROM users WHERE id = ?", userJSON, user.ID)

    return nil
}
```

### 4. Write-Behind Pattern

Delayed cache invalidation:

```go
func updateUser(user *User) error {
    // Update database
    err := db.Exec("UPDATE users SET name = ? WHERE id = ?",
        user.Name, user.ID)
    if err != nil {
        return err
    }

    // Schedule cache invalidation (e.g., after transaction commit)
    defer qc.InvalidateTable(ctx, "users")

    return nil
}
```

## Cache Invalidation Strategies

### Manual Invalidation

```go
// Invalidate specific entry
qc.Invalidate(ctx, query, args...)

// Invalidate by pattern
qc.InvalidatePattern(ctx, "*users*")

// Invalidate entire table
qc.InvalidateTable(ctx, "users")
```

### Automatic TTL

```go
// Set short TTL for frequently changing data
config := databases.DefaultQueryCacheConfig()
config.DefaultTTL = 1 * time.Minute

// Or use invalidation rules
config.InvalidationRules = []databases.InvalidationRule{
    {
        Pattern: "*users*",
        TTL:     30 * time.Second,
    },
    {
        Pattern: "*sessions*",
        TTL:     10 * time.Second,
    },
}
```

### Event-Driven Invalidation

```go
// Listen for database changes
func onUserUpdated(userID int) {
    qc.Invalidate(ctx, "SELECT * FROM users WHERE id = ?", userID)
}

// Or invalidate related queries
func onOrderCreated(orderID int) {
    // Invalidate user's orders
    qc.InvalidatePattern(ctx, "*user:*:orders")

    // Invalidate order summaries
    qc.InvalidateTable(ctx, "orders")
}
```

## Best Practices

### 1. Choose Appropriate TTL

```go
// Short TTL for frequently changing data
config.DefaultTTL = 1 * time.Minute     // User sessions, real-time data

// Medium TTL for moderate changes
config.DefaultTTL = 5 * time.Minute     // User profiles, product info

// Long TTL for rarely changing data
config.DefaultTTL = 1 * time.Hour       // Static config, reference data
```

### 2. Cache Size Management

```go
// Limit cache entry size
config.MaxCacheSize = 1024 * 1024 // 1MB per entry

// Limit total cache size (memory backend)
backend := databases.NewMemoryCache(100 * 1024 * 1024) // 100MB total
```

### 3. Selective Caching

Don't cache everything:

```go
func shouldCache(query string) bool {
    // Don't cache writes
    if isWriteQuery(query) {
        return false
    }

    // Don't cache large result sets
    if isLargeQuery(query) {
        return false
    }

    // Don't cache user-specific sensitive data
    if isSensitiveQuery(query) {
        return false
    }

    return true
}
```

### 4. Monitor Cache Performance

```go
// Periodically check metrics
metrics := qc.GetMetrics()

// Alert if hit rate is low
if metrics.HitRate < 0.5 {
    log.Printf("WARNING: Low cache hit rate: %.2f%%", metrics.HitRate*100)
}

// Alert if too many errors
errorRate := float64(metrics.Errors) / float64(metrics.Sets)
if errorRate > 0.01 {
    log.Printf("WARNING: High cache error rate: %.2f%%", errorRate*100)
}
```

### 5. Handle Cache Failures Gracefully

```go
data, err := qc.Get(ctx, query, args...)
if err != nil {
    // Don't fail - just query database
    data, err = queryDatabase()
}

// Similarly for Set
err = qc.Set(ctx, query, data, args...)
if err != nil {
    // Log but don't fail the request
    log.Printf("Failed to cache query: %v", err)
}
```

## Performance Considerations

### Cache Key Generation

Cache keys are generated using SHA-256 hash:

```go
// Efficient key generation
key := qc.generateKey(query, args...)

// Keys are deterministic:
// Same query + args = same key
```

### Memory Usage

```go
// Monitor memory cache size
backend := databases.NewMemoryCache(maxSize)

// Each entry uses:
// - Key size: ~50 bytes
// - Value size: actual data
// - Metadata: ~100 bytes
```

### Concurrency

All cache operations are thread-safe:

```go
// Safe for concurrent use
go qc.Set(ctx, query1, data1)
go qc.Set(ctx, query2, data2)
go qc.Get(ctx, query3)
```

## Troubleshooting

### Low Hit Rate

**Problem:** Cache hit rate < 50%

**Solutions:**
1. Increase TTL
2. Check if queries are cacheable
3. Verify cache invalidation isn't too aggressive
4. Monitor query patterns

```go
metrics := qc.GetMetrics()
if metrics.HitRate < 0.5 {
    // Analyze query patterns
    // Adjust TTL
    // Review invalidation rules
}
```

### High Memory Usage

**Problem:** Cache using too much memory

**Solutions:**
1. Reduce max cache size
2. Lower MaxCacheSize per entry
3. Implement more aggressive eviction
4. Use Redis instead of memory

```go
// Reduce memory usage
config.MaxCacheSize = 512 * 1024 // 512KB per entry
backend := databases.NewMemoryCache(50 * 1024 * 1024) // 50MB total
```

### Cache Stampede

**Problem:** Many requests hit database simultaneously on cache miss

**Solution:** Use single-flight pattern:

```go
// TODO: Implement single-flight to prevent stampede
// Only one request fetches data, others wait
```

## Integration Examples

### With Service Layer

```go
type UserService struct {
    db    *sql.DB
    cache *databases.QueryCache
}

func (s *UserService) GetUser(userID int) (*User, error) {
    query := "SELECT * FROM users WHERE id = ?"

    return s.cache.GetOrSet(ctx, query, func() ([]byte, error) {
        var user User
        err := s.db.QueryRow(query, userID).Scan(&user)
        if err != nil {
            return nil, err
        }
        return json.Marshal(user)
    }, userID)
}

func (s *UserService) UpdateUser(user *User) error {
    err := s.db.Exec("UPDATE users SET ... WHERE id = ?", user.ID)
    if err != nil {
        return err
    }

    // Invalidate cache
    return s.cache.Invalidate(ctx, "SELECT * FROM users WHERE id = ?", user.ID)
}
```

### With ORM (GORM)

```go
func getCachedUser(db *gorm.DB, cache *databases.QueryCache, userID int) (*User, error) {
    query := "users:id:" + strconv.Itoa(userID)

    data, err := cache.GetOrSet(ctx, query, func() ([]byte, error) {
        var user User
        if err := db.First(&user, userID).Error; err != nil {
            return nil, err
        }
        return json.Marshal(user)
    })

    if err != nil {
        return nil, err
    }

    var user User
    json.Unmarshal(data, &user)
    return &user, nil
}
```

## Metrics

### Available Metrics

```go
type CacheMetrics struct {
    Hits          int64   // Number of cache hits
    Misses        int64   // Number of cache misses
    Sets          int64   // Number of cache sets
    Errors        int64   // Number of errors
    Invalidations int64   // Number of invalidations
    HitRate       float64 // Hit rate (hits / (hits + misses))
}
```

### Monitoring

```go
// Get current metrics
metrics := qc.GetMetrics()

// Log metrics periodically
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        m := qc.GetMetrics()
        log.Printf("Cache: Hits=%d, Misses=%d, HitRate=%.2f%%",
            m.Hits, m.Misses, m.HitRate*100)
    }
}()
```

## Testing

Run query cache tests:

```bash
cd databases
go test -v -run TestQueryCache
```

Run benchmarks:

```bash
go test -bench=BenchmarkQueryCache -benchmem
```

## Examples

Complete examples available in:
- `/examples/query_cache_example.go` - Comprehensive usage examples
- `/databases/query_cache_test.go` - Unit tests with examples

## Future Enhancements

- Redis backend implementation
- Single-flight pattern for cache stampede prevention
- Distributed cache synchronization
- Query result compression
- Adaptive TTL based on query patterns
- Cache warming strategies

## License

Copyright 2023 IAC. All Rights Reserved.

Licensed under the Apache License, Version 2.0.

---

**Version:** 1.0
**Last Updated:** 2025-11-16
**Maintained By:** IAC Development Team
