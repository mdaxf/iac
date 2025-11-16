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
	"fmt"
	"log"

	"github.com/mdaxf/iac/databases"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("IAC Database Sharding Example")
	fmt.Println("==============================")

	// Example 1: Hash-based sharding
	hashShardingExample()

	// Example 2: Consistent hashing
	consistentHashingExample()

	// Example 3: Range-based sharding
	rangeShardingExample()

	// Example 4: Lookup-based sharding
	lookupShardingExample()

	// Example 5: Cross-shard queries
	crossShardQueryExample()

	// Example 6: Shard statistics
	shardStatisticsExample()
}

func hashShardingExample() {
	fmt.Println("\n1. Hash-Based Sharding (Modulo)")
	fmt.Println("--------------------------------")

	// Create shard manager with hash modulo strategy
	config := databases.DefaultShardManagerConfig()
	config.Strategy = databases.HashModulo
	sm := databases.NewShardManager(config)

	// Create 3 shards (in production, these would be separate databases)
	for i := 0; i < 3; i++ {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Create users table in each shard
		_, err = db.Exec(`
			CREATE TABLE users (
				id INTEGER PRIMARY KEY,
				name TEXT,
				email TEXT
			)
		`)
		if err != nil {
			log.Fatal(err)
		}

		shard := &databases.Shard{
			ID:     i,
			Name:   fmt.Sprintf("shard-%d", i),
			DB:     db,
			DBType: "sqlite3",
			Active: true,
			Region: "us-east-1",
		}

		sm.AddShard(shard)
		fmt.Printf("Added %s\n", shard.Name)
	}

	// Insert users - they'll be distributed across shards by user ID
	ctx := context.Background()
	users := []struct {
		id    int
		name  string
		email string
	}{
		{1, "Alice", "alice@example.com"},
		{2, "Bob", "bob@example.com"},
		{3, "Charlie", "charlie@example.com"},
		{4, "Dave", "dave@example.com"},
		{5, "Eve", "eve@example.com"},
	}

	for _, user := range users {
		shardKey := fmt.Sprintf("user:%d", user.id)
		shard, err := sm.GetShard(shardKey)
		if err != nil {
			log.Printf("Error getting shard: %v", err)
			continue
		}

		_, err = shard.DB.ExecContext(ctx,
			"INSERT INTO users (id, name, email) VALUES (?, ?, ?)",
			user.id, user.name, user.email)
		if err != nil {
			log.Printf("Error inserting user: %v", err)
			continue
		}

		fmt.Printf("  User %d (%s) -> %s\n", user.id, user.name, shard.Name)
	}

	// Show distribution
	keys := make([]string, len(users))
	for i, user := range users {
		keys[i] = fmt.Sprintf("user:%d", user.id)
	}

	distribution := sm.ShardDistribution(keys)
	fmt.Println("\nDistribution:")
	for shardID, count := range distribution {
		fmt.Printf("  Shard %d: %d users\n", shardID, count)
	}
}

func consistentHashingExample() {
	fmt.Println("\n2. Consistent Hashing")
	fmt.Println("----------------------")

	// Consistent hashing provides better distribution when adding/removing shards
	config := databases.DefaultShardManagerConfig()
	config.Strategy = databases.ConsistentHash
	config.VirtualNodes = 150 // More virtual nodes = better distribution
	sm := databases.NewShardManager(config)

	// Create 4 shards
	for i := 0; i < 4; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &databases.Shard{
			ID:     i,
			Name:   fmt.Sprintf("shard-%d", i),
			DB:     db,
			Active: true,
		}

		sm.AddShard(shard)
	}

	fmt.Printf("Created %d shards with consistent hashing\n", sm.GetShardCount())

	// Test distribution with many keys
	testKeys := 100
	distribution := make(map[int]int)

	for i := 0; i < testKeys; i++ {
		key := fmt.Sprintf("key:%d", i)
		shard, err := sm.GetShard(key)
		if err != nil {
			continue
		}
		distribution[shard.ID]++
	}

	fmt.Printf("\nDistribution of %d keys:\n", testKeys)
	for shardID, count := range distribution {
		percentage := float64(count) / float64(testKeys) * 100
		fmt.Printf("  Shard %d: %d keys (%.1f%%)\n", shardID, count, percentage)
	}

	// Demonstrate adding a new shard
	fmt.Println("\nAdding shard-4...")
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	newShard := &databases.Shard{
		ID:     4,
		Name:   "shard-4",
		DB:     db,
		Active: true,
	}
	sm.AddShard(newShard)

	// Recheck distribution
	distribution = make(map[int]int)
	for i := 0; i < testKeys; i++ {
		key := fmt.Sprintf("key:%d", i)
		shard, err := sm.GetShard(key)
		if err != nil {
			continue
		}
		distribution[shard.ID]++
	}

	fmt.Printf("\nDistribution after adding shard-4:\n")
	for shardID, count := range distribution {
		percentage := float64(count) / float64(testKeys) * 100
		fmt.Printf("  Shard %d: %d keys (%.1f%%)\n", shardID, count, percentage)
	}
}

func rangeShardingExample() {
	fmt.Println("\n3. Range-Based Sharding")
	fmt.Println("------------------------")

	// Range sharding is useful for time-series data or alphabetical distribution
	config := databases.DefaultShardManagerConfig()
	config.Strategy = databases.RangeSharding
	sm := databases.NewShardManager(config)

	// Create 3 shards for different date ranges
	shardRanges := []struct {
		id       int
		name     string
		minRange string
		maxRange string
	}{
		{0, "shard-2020", "2020-01-01", "2021-01-01"},
		{1, "shard-2021", "2021-01-01", "2022-01-01"},
		{2, "shard-2022", "2022-01-01", "2023-01-01"},
	}

	for _, sr := range shardRanges {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &databases.Shard{
			ID:     sr.id,
			Name:   sr.name,
			DB:     db,
			Active: true,
		}

		sm.AddShard(shard)
		sm.AddRangeMapping(sr.id, sr.minRange, sr.maxRange)

		fmt.Printf("Added %s (range: %s to %s)\n",
			sr.name, sr.minRange, sr.maxRange)
	}

	// Route queries by date
	dates := []string{
		"2020-05-15",
		"2021-08-20",
		"2022-03-10",
	}

	fmt.Println("\nRouting queries by date:")
	for _, date := range dates {
		shard, err := sm.GetShard(date)
		if err != nil {
			fmt.Printf("  %s -> Error: %v\n", date, err)
			continue
		}
		fmt.Printf("  %s -> %s\n", date, shard.Name)
	}
}

func lookupShardingExample() {
	fmt.Println("\n4. Lookup-Based Sharding")
	fmt.Println("-------------------------")

	// Lookup sharding is useful for multi-tenant applications
	config := databases.DefaultShardManagerConfig()
	config.Strategy = databases.LookupSharding
	sm := databases.NewShardManager(config)

	// Create 3 shards
	for i := 0; i < 3; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &databases.Shard{
			ID:     i,
			Name:   fmt.Sprintf("shard-%d", i),
			DB:     db,
			Active: true,
		}

		sm.AddShard(shard)
	}

	// Map tenants to specific shards
	tenantMappings := map[string]int{
		"tenant-acme":   0,
		"tenant-globex": 0,
		"tenant-stark":  1,
		"tenant-wayne":  1,
		"tenant-umbrella": 2,
	}

	fmt.Println("Tenant to shard mappings:")
	for tenant, shardID := range tenantMappings {
		sm.AddLookupMapping(tenant, shardID)
		fmt.Printf("  %s -> shard-%d\n", tenant, shardID)
	}

	// Route tenant queries
	fmt.Println("\nRouting tenant queries:")
	for tenant := range tenantMappings {
		shard, err := sm.GetShard(tenant)
		if err != nil {
			fmt.Printf("  %s -> Error: %v\n", tenant, err)
			continue
		}
		fmt.Printf("  %s -> %s\n", tenant, shard.Name)
	}
}

func crossShardQueryExample() {
	fmt.Println("\n5. Cross-Shard Queries")
	fmt.Println("-----------------------")

	config := databases.DefaultShardManagerConfig()
	config.EnableCrossShard = true
	config.MaxCrossShardQueries = 5
	sm := databases.NewShardManager(config)

	// Create 3 shards with data
	for i := 0; i < 3; i++ {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Create and populate table
		_, err = db.Exec(`
			CREATE TABLE orders (
				id INTEGER PRIMARY KEY,
				customer_id INTEGER,
				amount REAL
			)
		`)
		if err != nil {
			log.Fatal(err)
		}

		// Insert sample data
		_, err = db.Exec(
			"INSERT INTO orders (id, customer_id, amount) VALUES (?, ?, ?)",
			i*100, i, 100.00+float64(i)*50)
		if err != nil {
			log.Fatal(err)
		}

		shard := &databases.Shard{
			ID:     i,
			Name:   fmt.Sprintf("shard-%d", i),
			DB:     db,
			Active: true,
		}

		sm.AddShard(shard)
	}

	// Execute cross-shard query
	ctx := context.Background()
	results, err := sm.ExecuteOnAllShards(ctx, "SELECT id, customer_id, amount FROM orders")
	if err != nil {
		log.Printf("Cross-shard query failed: %v", err)
		return
	}

	fmt.Println("Results from all shards:")
	totalAmount := 0.0
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("  %s: Error - %v\n", result.ShardName, result.Error)
			continue
		}

		for result.Rows.Next() {
			var id, customerID int
			var amount float64
			if err := result.Rows.Scan(&id, &customerID, &amount); err != nil {
				continue
			}

			fmt.Printf("  %s: Order %d, Customer %d, Amount $%.2f\n",
				result.ShardName, id, customerID, amount)
			totalAmount += amount
		}
		result.Rows.Close()
	}

	fmt.Printf("\nTotal amount across all shards: $%.2f\n", totalAmount)
}

func shardStatisticsExample() {
	fmt.Println("\n6. Shard Statistics")
	fmt.Println("--------------------")

	sm := databases.NewShardManager(nil)

	// Create shards
	for i := 0; i < 3; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		db.Exec("CREATE TABLE test (id INTEGER)")

		shard := &databases.Shard{
			ID:     i,
			Name:   fmt.Sprintf("shard-%d", i),
			DB:     db,
			Active: true,
			Region: "us-east-1",
		}

		sm.AddShard(shard)
	}

	// Execute some queries
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key:%d", i)
		sm.ExecuteOnShard(ctx, key, "SELECT 1")
	}

	// Get and display statistics
	stats := sm.GetShardStats()
	fmt.Println("Shard Statistics:")
	for _, stat := range stats {
		fmt.Printf("  %s:\n", stat.ShardName)
		fmt.Printf("    Active: %v\n", stat.Active)
		fmt.Printf("    Queries: %d\n", stat.QueryCount)
		fmt.Printf("    Errors: %d\n", stat.ErrorCount)
		fmt.Printf("    Region: %s\n", stat.Region)
	}

	fmt.Printf("\nTotal shards: %d\n", sm.GetShardCount())
	fmt.Printf("Active shards: %d\n", sm.GetActiveShardCount())
}

// Example: Production-ready sharding setup
func productionShardingExample() {
	fmt.Println("\n7. Production Sharding Setup")
	fmt.Println("-----------------------------")

	// Configure shard manager for production
	config := &databases.ShardManagerConfig{
		Strategy:             databases.ConsistentHash,
		HashFunc:             databases.CRC32Hash,
		VirtualNodes:         150,
		EnableCrossShard:     true,
		MaxCrossShardQueries: 10,
		CrossShardTimeout:    30,
	}

	sm := databases.NewShardManager(config)

	// In production, connect to actual database servers
	shardConfigs := []struct {
		id     int
		name   string
		host   string
		region string
	}{
		{0, "shard-east-1", "db1.east.example.com", "us-east-1"},
		{1, "shard-east-2", "db2.east.example.com", "us-east-1"},
		{2, "shard-west-1", "db1.west.example.com", "us-west-1"},
		{3, "shard-west-2", "db2.west.example.com", "us-west-1"},
	}

	for _, sc := range shardConfigs {
		// In production:
		// db, err := sql.Open("postgres", fmt.Sprintf(
		//     "host=%s user=app password=secret dbname=mydb sslmode=require",
		//     sc.host))

		// For this example, using memory
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &databases.Shard{
			ID:     sc.id,
			Name:   sc.name,
			DB:     db,
			DBType: "postgres",
			Active: true,
			Region: sc.region,
		}

		if err := sm.AddShard(shard); err != nil {
			log.Printf("Failed to add shard %s: %v", sc.name, err)
			continue
		}

		fmt.Printf("Configured: %s in %s\n", sc.name, sc.region)
	}

	fmt.Printf("\nSharding configuration complete:\n")
	fmt.Printf("  Strategy: %s\n", config.Strategy)
	fmt.Printf("  Virtual Nodes: %d\n", config.VirtualNodes)
	fmt.Printf("  Cross-Shard Enabled: %v\n", config.EnableCrossShard)
	fmt.Printf("  Total Shards: %d\n", sm.GetShardCount())
}
