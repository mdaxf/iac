// Copyright 2023 IAC. All Rights Reserved.

package databases

import (
	"context"
	"testing"
	"time"
)

func TestNewQueryCache(t *testing.T) {
	backend := NewMemoryCache(1024 * 1024)
	qc := NewQueryCache(nil, backend)

	if qc == nil {
		t.Fatal("NewQueryCache returned nil")
	}

	if !qc.config.Enabled {
		t.Error("Cache should be enabled by default")
	}
}

func TestQueryCache_SetAndGet(t *testing.T) {
	backend := NewMemoryCache(1024 * 1024)
	qc := NewQueryCache(nil, backend)
	ctx := context.Background()

	query := "SELECT * FROM users WHERE id = ?"
	result := []byte(`{"id":1,"name":"Alice"}`)
	args := []interface{}{1}

	// Set cache
	err := qc.Set(ctx, query, result, args...)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get from cache
	cached, err := qc.Get(ctx, query, args...)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(cached) != string(result) {
		t.Errorf("Expected %s, got %s", string(result), string(cached))
	}
}

func TestQueryCache_CacheMiss(t *testing.T) {
	backend := NewMemoryCache(1024 * 1024)
	qc := NewQueryCache(nil, backend)
	ctx := context.Background()

	query := "SELECT * FROM users WHERE id = ?"
	args := []interface{}{999}

	_, err := qc.Get(ctx, query, args...)
	if err != ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss, got %v", err)
	}
}

func TestQueryCache_GetOrSet(t *testing.T) {
	backend := NewMemoryCache(1024 * 1024)
	qc := NewQueryCache(nil, backend)
	ctx := context.Background()

	query := "SELECT * FROM users WHERE id = ?"
	result := []byte(`{"id":2,"name":"Bob"}`)
	args := []interface{}{2}

	fetched := false
	fetcher := func() ([]byte, error) {
		fetched = true
		return result, nil
	}

	// First call should fetch
	data, err := qc.GetOrSet(ctx, query, fetcher, args...)
	if err != nil {
		t.Fatalf("GetOrSet failed: %v", err)
	}

	if !fetched {
		t.Error("Fetcher should have been called on cache miss")
	}

	if string(data) != string(result) {
		t.Errorf("Expected %s, got %s", string(result), string(data))
	}

	// Second call should use cache
	fetched = false
	data, err = qc.GetOrSet(ctx, query, fetcher, args...)
	if err != nil {
		t.Fatalf("GetOrSet failed: %v", err)
	}

	if fetched {
		t.Error("Fetcher should not have been called on cache hit")
	}
}

func TestQueryCache_Invalidate(t *testing.T) {
	backend := NewMemoryCache(1024 * 1024)
	qc := NewQueryCache(nil, backend)
	ctx := context.Background()

	query := "SELECT * FROM users WHERE id = ?"
	result := []byte(`{"id":3,"name":"Charlie"}`)
	args := []interface{}{3}

	// Set cache
	qc.Set(ctx, query, result, args...)

	// Verify it's cached
	_, err := qc.Get(ctx, query, args...)
	if err != nil {
		t.Error("Expected cache hit")
	}

	// Invalidate
	err = qc.Invalidate(ctx, query, args...)
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	// Verify it's gone
	_, err = qc.Get(ctx, query, args...)
	if err != ErrCacheMiss {
		t.Error("Expected cache miss after invalidation")
	}
}

func TestQueryCache_InvalidatePattern(t *testing.T) {
	backend := NewMemoryCache(1024 * 1024)
	qc := NewQueryCache(nil, backend)
	ctx := context.Background()

	// Set multiple cache entries
	qc.Set(ctx, "SELECT * FROM users WHERE id = 1", []byte("user1"), 1)
	qc.Set(ctx, "SELECT * FROM users WHERE id = 2", []byte("user2"), 2)
	qc.Set(ctx, "SELECT * FROM orders WHERE id = 1", []byte("order1"), 1)

	// Invalidate all user queries
	err := qc.InvalidatePattern(ctx, "*users*")
	if err != nil {
		t.Fatalf("InvalidatePattern failed: %v", err)
	}

	// User queries should be gone
	_, err = qc.Get(ctx, "SELECT * FROM users WHERE id = 1", 1)
	if err != ErrCacheMiss {
		t.Error("Expected user query to be invalidated")
	}

	// Order query should still exist
	_, err = qc.Get(ctx, "SELECT * FROM orders WHERE id = 1", 1)
	if err != nil {
		t.Error("Order query should not be invalidated")
	}
}

func TestQueryCache_InvalidateTable(t *testing.T) {
	backend := NewMemoryCache(1024 * 1024)
	qc := NewQueryCache(nil, backend)
	ctx := context.Background()

	// Set cache entries for different tables
	qc.Set(ctx, "SELECT * FROM users", []byte("users"), nil)
	qc.Set(ctx, "SELECT * FROM orders", []byte("orders"), nil)

	// Invalidate users table
	err := qc.InvalidateTable(ctx, "users")
	if err != nil {
		t.Fatalf("InvalidateTable failed: %v", err)
	}

	// Users query should be gone
	_, err = qc.Get(ctx, "SELECT * FROM users")
	if err != ErrCacheMiss {
		t.Error("Expected users query to be invalidated")
	}

	// Orders query should still exist
	_, err = qc.Get(ctx, "SELECT * FROM orders")
	if err != nil {
		t.Error("Orders query should not be invalidated")
	}
}

func TestQueryCache_Clear(t *testing.T) {
	backend := NewMemoryCache(1024 * 1024)
	qc := NewQueryCache(nil, backend)
	ctx := context.Background()

	// Set multiple cache entries
	qc.Set(ctx, "query1", []byte("result1"))
	qc.Set(ctx, "query2", []byte("result2"))
	qc.Set(ctx, "query3", []byte("result3"))

	// Clear cache
	err := qc.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// All entries should be gone
	_, err = qc.Get(ctx, "query1")
	if err != ErrCacheMiss {
		t.Error("Expected cache miss after clear")
	}

	_, err = qc.Get(ctx, "query2")
	if err != ErrCacheMiss {
		t.Error("Expected cache miss after clear")
	}
}

func TestQueryCache_SizeLimit(t *testing.T) {
	backend := NewMemoryCache(1024 * 1024)
	config := DefaultQueryCacheConfig()
	config.MaxCacheSize = 100 // Small limit
	qc := NewQueryCache(config, backend)
	ctx := context.Background()

	largeResult := make([]byte, 200)
	err := qc.Set(ctx, "large query", largeResult)

	if err != ErrCacheSizeLimitExceeded {
		t.Errorf("Expected ErrCacheSizeLimitExceeded, got %v", err)
	}
}

func TestQueryCache_Metrics(t *testing.T) {
	backend := NewMemoryCache(1024 * 1024)
	config := DefaultQueryCacheConfig()
	config.EnableMetrics = true
	qc := NewQueryCache(config, backend)
	ctx := context.Background()

	// Generate some cache activity
	qc.Set(ctx, "query1", []byte("result1"))
	qc.Get(ctx, "query1")                 // hit
	qc.Get(ctx, "query2")                 // miss
	qc.Get(ctx, "query1")                 // hit
	qc.Invalidate(ctx, "query1")

	metrics := qc.GetMetrics()

	if metrics.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", metrics.Hits)
	}

	if metrics.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", metrics.Misses)
	}

	if metrics.Sets != 1 {
		t.Errorf("Expected 1 set, got %d", metrics.Sets)
	}

	if metrics.Invalidations != 1 {
		t.Errorf("Expected 1 invalidation, got %d", metrics.Invalidations)
	}

	expectedHitRate := 2.0 / 3.0
	if metrics.HitRate != expectedHitRate {
		t.Errorf("Expected hit rate %.2f, got %.2f", expectedHitRate, metrics.HitRate)
	}
}

func TestQueryCache_ResetMetrics(t *testing.T) {
	backend := NewMemoryCache(1024 * 1024)
	qc := NewQueryCache(nil, backend)
	ctx := context.Background()

	// Generate some activity
	qc.Set(ctx, "query", []byte("result"))
	qc.Get(ctx, "query")

	// Reset metrics
	qc.ResetMetrics()

	metrics := qc.GetMetrics()
	if metrics.Hits != 0 || metrics.Misses != 0 || metrics.Sets != 0 {
		t.Error("Metrics should be reset to zero")
	}
}

func TestQueryCache_Disabled(t *testing.T) {
	backend := NewMemoryCache(1024 * 1024)
	config := DefaultQueryCacheConfig()
	config.Enabled = false
	qc := NewQueryCache(config, backend)
	ctx := context.Background()

	// Try to set
	err := qc.Set(ctx, "query", []byte("result"))
	if err != nil {
		t.Error("Set should not error when disabled (just no-op)")
	}

	// Try to get
	_, err = qc.Get(ctx, "query")
	if err != ErrCacheDisabled {
		t.Errorf("Expected ErrCacheDisabled, got %v", err)
	}
}

func TestQueryCache_TTL(t *testing.T) {
	backend := NewMemoryCache(1024 * 1024)
	config := DefaultQueryCacheConfig()
	config.DefaultTTL = 100 * time.Millisecond
	qc := NewQueryCache(config, backend)
	ctx := context.Background()

	query := "SELECT * FROM users"
	result := []byte("users data")

	// Set cache
	qc.Set(ctx, query, result)

	// Should be cached
	_, err := qc.Get(ctx, query)
	if err != nil {
		t.Error("Expected cache hit")
	}

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, err = qc.Get(ctx, query)
	if err != ErrCacheMiss {
		t.Error("Expected cache miss after TTL expiration")
	}
}

func TestMemoryCache_Eviction(t *testing.T) {
	maxSize := int64(100)
	mc := NewMemoryCache(maxSize)
	ctx := context.Background()

	// Fill cache to capacity
	mc.Set(ctx, "key1", make([]byte, 40), time.Hour)
	mc.Set(ctx, "key2", make([]byte, 40), time.Hour)

	// Verify both are cached
	exists, _ := mc.Exists(ctx, "key1")
	if !exists {
		t.Error("key1 should exist")
	}

	exists, _ = mc.Exists(ctx, "key2")
	if !exists {
		t.Error("key2 should exist")
	}

	// Add entry that requires eviction
	mc.Set(ctx, "key3", make([]byte, 50), time.Hour)

	// At least one old entry should be evicted
	if mc.currentSize > maxSize {
		t.Errorf("Cache size %d exceeds max %d", mc.currentSize, maxSize)
	}
}

func TestMemoryCache_ExpiredCleanup(t *testing.T) {
	mc := NewMemoryCache(1024)
	ctx := context.Background()

	// Set entry with short TTL
	mc.Set(ctx, "short", []byte("data"), 50*time.Millisecond)

	// Verify it exists
	exists, _ := mc.Exists(ctx, "short")
	if !exists {
		t.Error("Entry should exist before expiration")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be gone
	exists, _ = mc.Exists(ctx, "short")
	if exists {
		t.Error("Entry should be expired")
	}
}

func TestGenerateKey_Consistency(t *testing.T) {
	backend := NewMemoryCache(1024)
	qc := NewQueryCache(nil, backend)

	query := "SELECT * FROM users WHERE id = ?"

	// Same query and args should produce same key
	key1 := qc.generateKey(query, 1)
	key2 := qc.generateKey(query, 1)

	if key1 != key2 {
		t.Error("Same query and args should produce same key")
	}

	// Different args should produce different key
	key3 := qc.generateKey(query, 2)
	if key1 == key3 {
		t.Error("Different args should produce different key")
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		s       string
		pattern string
		want    bool
	}{
		{"hello", "hello", true},
		{"hello", "*", true},
		{"hello", "h*", true},
		{"hello", "*o", true},
		{"hello", "h*o", true},
		{"hello world", "*world", true},
		{"hello world", "hello*", true},
		{"hello", "goodbye", false},
	}

	for _, tt := range tests {
		got := matchesPattern(tt.s, tt.pattern)
		if got != tt.want {
			t.Errorf("matchesPattern(%q, %q) = %v, want %v",
				tt.s, tt.pattern, got, tt.want)
		}
	}
}

func BenchmarkQueryCache_Set(b *testing.B) {
	backend := NewMemoryCache(1024 * 1024 * 10)
	qc := NewQueryCache(nil, backend)
	ctx := context.Background()

	query := "SELECT * FROM users WHERE id = ?"
	result := []byte(`{"id":1,"name":"Test User"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qc.Set(ctx, query, result, i)
	}
}

func BenchmarkQueryCache_Get(b *testing.B) {
	backend := NewMemoryCache(1024 * 1024 * 10)
	qc := NewQueryCache(nil, backend)
	ctx := context.Background()

	query := "SELECT * FROM users WHERE id = ?"
	result := []byte(`{"id":1,"name":"Test User"}`)

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		qc.Set(ctx, query, result, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qc.Get(ctx, query, i%100)
	}
}

func BenchmarkQueryCache_GetOrSet(b *testing.B) {
	backend := NewMemoryCache(1024 * 1024 * 10)
	qc := NewQueryCache(nil, backend)
	ctx := context.Background()

	query := "SELECT * FROM users WHERE id = ?"
	result := []byte(`{"id":1,"name":"Test User"}`)

	fetcher := func() ([]byte, error) {
		return result, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qc.GetOrSet(ctx, query, fetcher, i%100)
	}
}
