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

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("IAC Query Caching Example")
	fmt.Println("==========================")

	// Example 1: Basic caching
	basicCachingExample()

	// Example 2: Cache invalidation
	cacheInvalidationExample()

	// Example 3: Cache metrics
	cacheMetricsExample()

	// Example 4: GetOrSet pattern
	getOrSetExample()

	// Example 5: TTL configuration
	ttlConfigExample()
}

func basicCachingExample() {
	fmt.Println("\n1. Basic Query Caching")
	fmt.Println("-----------------------")

	// Create in-memory cache backend
	backend := dbconn.NewMemoryCache(10 * 1024 * 1024) // 10MB

	// Create query cache
	config := dbconn.DefaultQueryCacheConfig()
	config.DefaultTTL = 5 * time.Minute
	qc := dbconn.NewQueryCache(config, backend)

	ctx := context.Background()

	// Simulate query results
	query := "SELECT * FROM users WHERE id = ?"
	userID := 12345
	userData := map[string]interface{}{
		"id":    12345,
		"name":  "Alice",
		"email": "alice@example.com",
	}

	userJSON, _ := json.Marshal(userData)

	// Cache the result
	err := qc.Set(ctx, query, userJSON, userID)
	if err != nil {
		log.Printf("Failed to cache: %v", err)
	}
	fmt.Println("Cached query result")

	// Retrieve from cache
	cached, err := qc.Get(ctx, query, userID)
	if err != nil {
		log.Printf("Cache miss: %v", err)
	} else {
		var user map[string]interface{}
		json.Unmarshal(cached, &user)
		fmt.Printf("Retrieved from cache: %v\n", user)
	}

	// Try to get non-existent entry
	_, err = qc.Get(ctx, query, 99999)
	if err == dbconn.ErrCacheMiss {
		fmt.Println("Cache miss for non-existent key (expected)")
	}
}

func cacheInvalidationExample() {
	fmt.Println("\n2. Cache Invalidation")
	fmt.Println("----------------------")

	backend := dbconn.NewMemoryCache(10 * 1024 * 1024)
	qc := dbconn.NewQueryCache(nil, backend)
	ctx := context.Background()

	// Cache some user queries
	for i := 1; i <= 3; i++ {
		query := "SELECT * FROM users WHERE id = ?"
		userData := map[string]interface{}{"id": i, "name": fmt.Sprintf("User%d", i)}
		userJSON, _ := json.Marshal(userData)
		qc.Set(ctx, query, userJSON, i)
	}

	// Cache an order query
	orderQuery := "SELECT * FROM orders WHERE id = ?"
	orderData := map[string]interface{}{"id": 1, "total": 99.99}
	orderJSON, _ := json.Marshal(orderData)
	qc.Set(ctx, orderQuery, orderJSON, 1)

	fmt.Println("Cached 3 user queries and 1 order query")

	// Invalidate specific user query
	qc.Invalidate(ctx, "SELECT * FROM users WHERE id = ?", 1)
	fmt.Println("Invalidated user 1")

	// Check if invalidated
	_, err := qc.Get(ctx, "SELECT * FROM users WHERE id = ?", 1)
	if err == dbconn.ErrCacheMiss {
		fmt.Println("  User 1 is invalidated (cache miss)")
	}

	// User 2 should still be cached
	_, err = qc.Get(ctx, "SELECT * FROM users WHERE id = ?", 2)
	if err == nil {
		fmt.Println("  User 2 is still cached")
	}

	// Invalidate all users table queries
	qc.InvalidateTable(ctx, "users")
	fmt.Println("\nInvalidated all 'users' table queries")

	// All user queries should be gone
	_, err = qc.Get(ctx, "SELECT * FROM users WHERE id = ?", 2)
	if err == dbconn.ErrCacheMiss {
		fmt.Println("  All user queries invalidated")
	}

	// Order query should still exist
	_, err = qc.Get(ctx, orderQuery, 1)
	if err == nil {
		fmt.Println("  Order query still cached")
	}
}

func cacheMetricsExample() {
	fmt.Println("\n3. Cache Metrics")
	fmt.Println("-----------------")

	backend := dbconn.NewMemoryCache(10 * 1024 * 1024)
	config := dbconn.DefaultQueryCacheConfig()
	config.EnableMetrics = true
	qc := dbconn.NewQueryCache(config, backend)
	ctx := context.Background()

	// Simulate some cache activity
	query := "SELECT * FROM products WHERE id = ?"

	// Cache some products
	for i := 1; i <= 5; i++ {
		product := map[string]interface{}{"id": i, "name": fmt.Sprintf("Product%d", i)}
		productJSON, _ := json.Marshal(product)
		qc.Set(ctx, query, productJSON, i)
	}

	// Generate hits and misses
	qc.Get(ctx, query, 1) // hit
	qc.Get(ctx, query, 2) // hit
	qc.Get(ctx, query, 3) // hit
	qc.Get(ctx, query, 99) // miss
	qc.Get(ctx, query, 98) // miss
	qc.Get(ctx, query, 1) // hit

	// Get metrics
	metrics := qc.GetMetrics()

	fmt.Printf("Cache Metrics:\n")
	fmt.Printf("  Hits: %d\n", metrics.Hits)
	fmt.Printf("  Misses: %d\n", metrics.Misses)
	fmt.Printf("  Sets: %d\n", metrics.Sets)
	fmt.Printf("  Hit Rate: %.2f%%\n", metrics.HitRate*100)
	fmt.Printf("  Invalidations: %d\n", metrics.Invalidations)
}

func getOrSetExample() {
	fmt.Println("\n4. GetOrSet Pattern")
	fmt.Println("--------------------")

	backend := dbconn.NewMemoryCache(10 * 1024 * 1024)
	qc := dbconn.NewQueryCache(nil, backend)
	ctx := context.Background()

	// Simulate database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create and populate table
	db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)")
	db.Exec("INSERT INTO users VALUES (1, 'Alice', 'alice@example.com')")
	db.Exec("INSERT INTO users VALUES (2, 'Bob', 'bob@example.com')")

	// Function to get user with caching
	getUser := func(userID int) (map[string]interface{}, error) {
		query := "SELECT id, name, email FROM users WHERE id = ?"

		// Fetcher function - called on cache miss
		fetcher := func() ([]byte, error) {
			fmt.Printf("  Querying database for user %d...\n", userID)

			var id int
			var name, email string
			err := db.QueryRow(query, userID).Scan(&id, &name, &email)
			if err != nil {
				return nil, err
			}

			user := map[string]interface{}{
				"id":    id,
				"name":  name,
				"email": email,
			}

			return json.Marshal(user)
		}

		// Get from cache or fetch
		data, err := qc.GetOrSet(ctx, query, fetcher, userID)
		if err != nil {
			return nil, err
		}

		var user map[string]interface{}
		json.Unmarshal(data, &user)
		return user, nil
	}

	// First call - cache miss, will query database
	fmt.Println("First call (cache miss):")
	user, _ := getUser(1)
	fmt.Printf("  Result: %v\n", user)

	// Second call - cache hit, won't query database
	fmt.Println("\nSecond call (cache hit):")
	user, _ = getUser(1)
	fmt.Printf("  Result: %v\n", user)

	// Different user - cache miss
	fmt.Println("\nDifferent user (cache miss):")
	user, _ = getUser(2)
	fmt.Printf("  Result: %v\n", user)
}

func ttlConfigExample() {
	fmt.Println("\n5. TTL Configuration")
	fmt.Println("---------------------")

	backend := dbconn.NewMemoryCache(10 * 1024 * 1024)
	config := dbconn.DefaultQueryCacheConfig()

	// Short TTL for frequently changing data
	config.DefaultTTL = 200 * time.Millisecond
	qc := dbconn.NewQueryCache(config, backend)
	ctx := context.Background()

	query := "SELECT COUNT(*) FROM active_sessions"
	result := []byte(`{"count":42}`)

	// Cache result
	qc.Set(ctx, query, result)
	fmt.Println("Cached result with 200ms TTL")

	// Should be cached
	_, err := qc.Get(ctx, query)
	if err == nil {
		fmt.Println("  Immediately: Cache hit")
	}

	// Wait for TTL to expire
	time.Sleep(250 * time.Millisecond)

	// Should be expired
	_, err = qc.Get(ctx, query)
	if err == dbconn.ErrCacheMiss {
		fmt.Println("  After 250ms: Cache miss (expired)")
	}
}

// Example: Complete caching integration with database
func completeCachingExample() {
	fmt.Println("\n6. Complete Caching Integration")
	fmt.Println("---------------------------------")

	// Setup database
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		name TEXT,
		email TEXT,
		created_at TIMESTAMP
	)`)

	// Insert test data
	db.Exec("INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', datetime('now'))")
	db.Exec("INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', datetime('now'))")

	// Setup cache
	backend := dbconn.NewMemoryCache(10 * 1024 * 1024)
	config := &dbconn.QueryCacheConfig{
		Enabled:        true,
		DefaultTTL:     5 * time.Minute,
		MaxCacheSize:   1024 * 1024,
		CacheKeyPrefix: "app:users:",
		EnableMetrics:  true,
	}
	qc := dbconn.NewQueryCache(config, backend)
	ctx := context.Background()

	// Service layer function
	getUserByID := func(userID int) (map[string]interface{}, error) {
		query := "SELECT id, name, email FROM users WHERE id = ?"

		// Try cache first
		data, err := qc.Get(ctx, query, userID)
		if err == nil {
			// Cache hit
			var user map[string]interface{}
			json.Unmarshal(data, &user)
			return user, nil
		}

		// Cache miss - query database
		var id int
		var name, email string
		err = db.QueryRow(query, userID).Scan(&id, &name, &email)
		if err != nil {
			return nil, err
		}

		user := map[string]interface{}{
			"id":    id,
			"name":  name,
			"email": email,
		}

		// Cache the result
		userJSON, _ := json.Marshal(user)
		qc.Set(ctx, query, userJSON, userID)

		return user, nil
	}

	// Update user function (invalidates cache)
	updateUser := func(userID int, name string) error {
		query := "UPDATE users SET name = ? WHERE id = ?"
		_, err := db.Exec(query, name, userID)
		if err != nil {
			return err
		}

		// Invalidate cache for this user
		qc.Invalidate(ctx, "SELECT id, name, email FROM users WHERE id = ?", userID)
		return nil
	}

	// Demo usage
	fmt.Println("Getting user 1 (cache miss):")
	user, _ := getUserByID(1)
	fmt.Printf("  %v\n", user)

	fmt.Println("\nGetting user 1 again (cache hit):")
	user, _ = getUserByID(1)
	fmt.Printf("  %v\n", user)

	fmt.Println("\nUpdating user 1:")
	updateUser(1, "Alice Smith")
	fmt.Println("  Cache invalidated")

	fmt.Println("\nGetting user 1 after update (cache miss):")
	user, _ = getUserByID(1)
	fmt.Printf("  %v\n", user)

	// Show final metrics
	metrics := qc.GetMetrics()
	fmt.Printf("\nFinal Metrics:\n")
	fmt.Printf("  Hits: %d, Misses: %d\n", metrics.Hits, metrics.Misses)
	fmt.Printf("  Hit Rate: %.2f%%\n", metrics.HitRate*100)
}
