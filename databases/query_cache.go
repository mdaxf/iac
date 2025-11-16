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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// CacheBackend defines the interface for cache storage backends
type CacheBackend interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
}

// QueryCacheConfig configures the query cache
type QueryCacheConfig struct {
	// Enabled enables/disables caching
	Enabled bool

	// DefaultTTL is the default time-to-live for cached queries
	DefaultTTL time.Duration

	// MaxCacheSize is the maximum size of cached values (in bytes)
	MaxCacheSize int

	// CacheKeyPrefix is prepended to all cache keys
	CacheKeyPrefix string

	// EnableMetrics enables cache hit/miss metrics
	EnableMetrics bool

	// InvalidationRules defines cache invalidation rules
	InvalidationRules []InvalidationRule
}

// DefaultQueryCacheConfig returns default configuration
func DefaultQueryCacheConfig() *QueryCacheConfig {
	return &QueryCacheConfig{
		Enabled:        true,
		DefaultTTL:     5 * time.Minute,
		MaxCacheSize:   1024 * 1024, // 1MB
		CacheKeyPrefix: "iac:query:",
		EnableMetrics:  true,
		InvalidationRules: []InvalidationRule{},
	}
}

// InvalidationRule defines when to invalidate cache
type InvalidationRule struct {
	// Pattern matches cache keys to invalidate
	Pattern string

	// OnWrite invalidates on write operations to specified tables
	OnWrite []string

	// TTL custom TTL for this rule
	TTL time.Duration
}

// QueryCache provides query result caching
type QueryCache struct {
	config  *QueryCacheConfig
	backend CacheBackend
	mu      sync.RWMutex

	// Metrics
	hits          int64
	misses        int64
	sets          int64
	errors        int64
	invalidations int64
}

// NewQueryCache creates a new query cache
func NewQueryCache(config *QueryCacheConfig, backend CacheBackend) *QueryCache {
	if config == nil {
		config = DefaultQueryCacheConfig()
	}

	return &QueryCache{
		config:  config,
		backend: backend,
	}
}

// Get retrieves a cached query result
func (qc *QueryCache) Get(ctx context.Context, query string, args ...interface{}) ([]byte, error) {
	if !qc.config.Enabled {
		return nil, ErrCacheDisabled
	}

	key := qc.generateKey(query, args...)

	data, err := qc.backend.Get(ctx, key)
	if err != nil {
		atomic.AddInt64(&qc.misses, 1)
		return nil, err
	}

	if data == nil {
		atomic.AddInt64(&qc.misses, 1)
		return nil, ErrCacheMiss
	}

	atomic.AddInt64(&qc.hits, 1)
	return data, nil
}

// Set stores a query result in cache
func (qc *QueryCache) Set(ctx context.Context, query string, result []byte, args ...interface{}) error {
	if !qc.config.Enabled {
		return nil
	}

	// Check size limit
	if len(result) > qc.config.MaxCacheSize {
		return ErrCacheSizeLimitExceeded
	}

	key := qc.generateKey(query, args...)
	ttl := qc.getTTL(query)

	err := qc.backend.Set(ctx, key, result, ttl)
	if err != nil {
		atomic.AddInt64(&qc.errors, 1)
		return err
	}

	atomic.AddInt64(&qc.sets, 1)
	return nil
}

// GetOrSet retrieves from cache or executes query and caches result
func (qc *QueryCache) GetOrSet(
	ctx context.Context,
	query string,
	fetcher func() ([]byte, error),
	args ...interface{},
) ([]byte, error) {
	// Try to get from cache first
	data, err := qc.Get(ctx, query, args...)
	if err == nil {
		return data, nil
	}

	// Cache miss or error - fetch data
	data, err = fetcher()
	if err != nil {
		return nil, err
	}

	// Store in cache (don't fail if caching fails)
	_ = qc.Set(ctx, query, data, args...)

	return data, nil
}

// Invalidate removes a specific cache entry
func (qc *QueryCache) Invalidate(ctx context.Context, query string, args ...interface{}) error {
	if !qc.config.Enabled {
		return nil
	}

	key := qc.generateKey(query, args...)
	err := qc.backend.Delete(ctx, key)
	if err != nil {
		return err
	}

	atomic.AddInt64(&qc.invalidations, 1)
	return nil
}

// InvalidatePattern removes cache entries matching a pattern
func (qc *QueryCache) InvalidatePattern(ctx context.Context, pattern string) error {
	if !qc.config.Enabled {
		return nil
	}

	fullPattern := qc.config.CacheKeyPrefix + pattern
	err := qc.backend.DeletePattern(ctx, fullPattern)
	if err != nil {
		return err
	}

	atomic.AddInt64(&qc.invalidations, 1)
	return nil
}

// InvalidateTable invalidates all queries related to a table
func (qc *QueryCache) InvalidateTable(ctx context.Context, tableName string) error {
	pattern := fmt.Sprintf("*%s*", tableName)
	return qc.InvalidatePattern(ctx, pattern)
}

// Clear removes all cached entries
func (qc *QueryCache) Clear(ctx context.Context) error {
	if !qc.config.Enabled {
		return nil
	}

	return qc.backend.Clear(ctx)
}

// generateKey generates a cache key from query and arguments
func (qc *QueryCache) generateKey(query string, args ...interface{}) string {
	// Create a deterministic key from query and args
	hasher := sha256.New()
	hasher.Write([]byte(query))

	// Hash arguments
	for _, arg := range args {
		argBytes, _ := json.Marshal(arg)
		hasher.Write(argBytes)
	}

	hash := hex.EncodeToString(hasher.Sum(nil))
	return qc.config.CacheKeyPrefix + hash[:16] // Use first 16 chars of hash
}

// getTTL returns the TTL for a query based on invalidation rules
func (qc *QueryCache) getTTL(query string) time.Duration {
	for _, rule := range qc.config.InvalidationRules {
		if matchesPattern(query, rule.Pattern) && rule.TTL > 0 {
			return rule.TTL
		}
	}

	return qc.config.DefaultTTL
}

// GetMetrics returns cache metrics
func (qc *QueryCache) GetMetrics() CacheMetrics {
	hits := atomic.LoadInt64(&qc.hits)
	misses := atomic.LoadInt64(&qc.misses)
	total := hits + misses

	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	return CacheMetrics{
		Hits:          hits,
		Misses:        misses,
		Sets:          atomic.LoadInt64(&qc.sets),
		Errors:        atomic.LoadInt64(&qc.errors),
		Invalidations: atomic.LoadInt64(&qc.invalidations),
		HitRate:       hitRate,
	}
}

// ResetMetrics resets all metrics counters
func (qc *QueryCache) ResetMetrics() {
	atomic.StoreInt64(&qc.hits, 0)
	atomic.StoreInt64(&qc.misses, 0)
	atomic.StoreInt64(&qc.sets, 0)
	atomic.StoreInt64(&qc.errors, 0)
	atomic.StoreInt64(&qc.invalidations, 0)
}

// CacheMetrics contains cache statistics
type CacheMetrics struct {
	Hits          int64
	Misses        int64
	Sets          int64
	Errors        int64
	Invalidations int64
	HitRate       float64
}

// MemoryCache is a simple in-memory cache backend
type MemoryCache struct {
	data   map[string]*cacheEntry
	mu     sync.RWMutex
	maxSize int64
	currentSize int64
}

type cacheEntry struct {
	value      []byte
	expiration time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(maxSize int64) *MemoryCache {
	mc := &MemoryCache{
		data:    make(map[string]*cacheEntry),
		maxSize: maxSize,
	}

	// Start cleanup goroutine
	go mc.cleanupExpired()

	return mc
}

// Get retrieves a value from memory cache
func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	entry, exists := mc.data[key]
	if !exists {
		return nil, ErrCacheMiss
	}

	// Check expiration
	if time.Now().After(entry.expiration) {
		return nil, ErrCacheMiss
	}

	return entry.value, nil
}

// Set stores a value in memory cache
func (mc *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check size limit
	newSize := mc.currentSize + int64(len(value))
	if mc.maxSize > 0 && newSize > mc.maxSize {
		// Evict oldest entries until there's space
		mc.evictOldest(int64(len(value)))
	}

	mc.data[key] = &cacheEntry{
		value:      value,
		expiration: time.Now().Add(ttl),
	}

	mc.currentSize += int64(len(value))
	return nil
}

// Delete removes a value from memory cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if entry, exists := mc.data[key]; exists {
		mc.currentSize -= int64(len(entry.value))
		delete(mc.data, key)
	}

	return nil
}

// DeletePattern removes all entries matching pattern
func (mc *MemoryCache) DeletePattern(ctx context.Context, pattern string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	for key, entry := range mc.data {
		if matchesPattern(key, pattern) {
			mc.currentSize -= int64(len(entry.value))
			delete(mc.data, key)
		}
	}

	return nil
}

// Exists checks if a key exists
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	entry, exists := mc.data[key]
	if !exists {
		return false, nil
	}

	// Check expiration
	if time.Now().After(entry.expiration) {
		return false, nil
	}

	return true, nil
}

// Clear removes all entries
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.data = make(map[string]*cacheEntry)
	mc.currentSize = 0
	return nil
}

// evictOldest evicts oldest entries to make space
func (mc *MemoryCache) evictOldest(needed int64) {
	// Simple LRU-like eviction
	for key, entry := range mc.data {
		if mc.currentSize+needed <= mc.maxSize {
			break
		}

		mc.currentSize -= int64(len(entry.value))
		delete(mc.data, key)
	}
}

// cleanupExpired periodically removes expired entries
func (mc *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		mc.mu.Lock()
		now := time.Now()

		for key, entry := range mc.data {
			if now.After(entry.expiration) {
				mc.currentSize -= int64(len(entry.value))
				delete(mc.data, key)
			}
		}
		mc.mu.Unlock()
	}
}

// matchesPattern checks if a string matches a simple glob pattern
func matchesPattern(s, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Simple wildcard matching - production would use more sophisticated matching
	if len(pattern) == 0 {
		return len(s) == 0
	}

	// Check for * wildcard
	if pattern[0] == '*' {
		// Try matching rest of pattern at any position
		for i := 0; i <= len(s); i++ {
			if matchesPattern(s[i:], pattern[1:]) {
				return true
			}
		}
		return false
	}

	// Check if first characters match
	if len(s) > 0 && pattern[0] == s[0] {
		return matchesPattern(s[1:], pattern[1:])
	}

	return false
}

// Cache errors
var (
	ErrCacheDisabled          = fmt.Errorf("cache is disabled")
	ErrCacheMiss              = fmt.Errorf("cache miss")
	ErrCacheSizeLimitExceeded = fmt.Errorf("cache size limit exceeded")
)
