// Copyright 2023 IAC. All Rights Reserved.

package dbconn

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewShardManager(t *testing.T) {
	sm := NewShardManager(nil)
	if sm == nil {
		t.Fatal("NewShardManager returned nil")
	}

	if sm.config.Strategy != HashModulo {
		t.Errorf("Expected default strategy %s, got %s", HashModulo, sm.config.Strategy)
	}
}

func TestAddShard(t *testing.T) {
	sm := NewShardManager(nil)

	// Create test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	shard := &Shard{
		ID:     1,
		Name:   "shard-1",
		DB:     db,
		DBType: "sqlite3",
		Active: true,
	}

	err = sm.AddShard(shard)
	if err != nil {
		t.Errorf("AddShard failed: %v", err)
	}

	if sm.GetShardCount() != 1 {
		t.Errorf("Expected 1 shard, got %d", sm.GetShardCount())
	}
}

func TestAddShard_Nil(t *testing.T) {
	sm := NewShardManager(nil)

	err := sm.AddShard(nil)
	if err == nil {
		t.Error("Expected error when adding nil shard")
	}
}

func TestRemoveShard(t *testing.T) {
	sm := NewShardManager(nil)

	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	shard := &Shard{
		ID:     1,
		Name:   "shard-1",
		DB:     db,
		Active: true,
	}

	sm.AddShard(shard)

	err := sm.RemoveShard(1)
	if err != nil {
		t.Errorf("RemoveShard failed: %v", err)
	}

	if sm.GetShardCount() != 0 {
		t.Errorf("Expected 0 shards after removal, got %d", sm.GetShardCount())
	}
}

func TestGetShard_HashModulo(t *testing.T) {
	config := DefaultShardManagerConfig()
	config.Strategy = HashModulo
	sm := NewShardManager(config)

	// Create 3 shards
	for i := 0; i < 3; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &Shard{
			ID:     i,
			Name:   "shard-" + string(rune('0'+i)),
			DB:     db,
			Active: true,
		}
		sm.AddShard(shard)
	}

	// Test that same key always returns same shard
	key := "user:12345"
	shard1, err := sm.GetShard(key)
	if err != nil {
		t.Fatalf("GetShard failed: %v", err)
	}

	shard2, err := sm.GetShard(key)
	if err != nil {
		t.Fatalf("GetShard failed: %v", err)
	}

	if shard1.ID != shard2.ID {
		t.Error("Same key should return same shard")
	}
}

func TestGetShard_ConsistentHash(t *testing.T) {
	config := DefaultShardManagerConfig()
	config.Strategy = ConsistentHash
	config.VirtualNodes = 150
	sm := NewShardManager(config)

	// Create shards
	for i := 0; i < 3; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &Shard{
			ID:     i,
			Name:   "shard-" + string(rune('0'+i)),
			DB:     db,
			Active: true,
		}
		sm.AddShard(shard)
	}

	// Test consistency
	key := "user:67890"
	shard1, err := sm.GetShard(key)
	if err != nil {
		t.Fatalf("GetShard failed: %v", err)
	}

	shard2, err := sm.GetShard(key)
	if err != nil {
		t.Fatalf("GetShard failed: %v", err)
	}

	if shard1.ID != shard2.ID {
		t.Error("Consistent hash should return same shard for same key")
	}
}

func TestGetShard_RangeSharding(t *testing.T) {
	config := DefaultShardManagerConfig()
	config.Strategy = RangeSharding
	sm := NewShardManager(config)

	// Create shards
	for i := 0; i < 3; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &Shard{
			ID:     i,
			Name:   "shard-" + string(rune('0'+i)),
			DB:     db,
			Active: true,
		}
		sm.AddShard(shard)
	}

	// Add range mappings
	sm.AddRangeMapping(0, "A", "H")
	sm.AddRangeMapping(1, "H", "P")
	sm.AddRangeMapping(2, "P", "Z")

	// Test range selection
	shard, err := sm.GetShard("Alice")
	if err != nil {
		t.Fatalf("GetShard failed: %v", err)
	}

	if shard.ID != 0 {
		t.Errorf("Expected shard 0 for 'Alice', got %d", shard.ID)
	}
}

func TestGetShard_LookupSharding(t *testing.T) {
	config := DefaultShardManagerConfig()
	config.Strategy = LookupSharding
	sm := NewShardManager(config)

	// Create shards
	for i := 0; i < 2; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &Shard{
			ID:     i,
			Name:   "shard-" + string(rune('0'+i)),
			DB:     db,
			Active: true,
		}
		sm.AddShard(shard)
	}

	// Add lookup mappings
	sm.AddLookupMapping("customer:1", 0)
	sm.AddLookupMapping("customer:2", 1)

	// Test lookup
	shard, err := sm.GetShard("customer:1")
	if err != nil {
		t.Fatalf("GetShard failed: %v", err)
	}

	if shard.ID != 0 {
		t.Errorf("Expected shard 0, got %d", shard.ID)
	}
}

func TestGetShard_NoShards(t *testing.T) {
	sm := NewShardManager(nil)

	_, err := sm.GetShard("test")
	if err == nil {
		t.Error("Expected error when no shards available")
	}
}

func TestGetAllShards(t *testing.T) {
	sm := NewShardManager(nil)

	// Create 3 shards
	for i := 0; i < 3; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &Shard{
			ID:     i,
			Name:   "shard-" + string(rune('0'+i)),
			DB:     db,
			Active: true,
		}
		sm.AddShard(shard)
	}

	// Mark one as inactive
	sm.mu.Lock()
	sm.shards[1].Active = false
	sm.mu.Unlock()

	shards := sm.GetAllShards()
	if len(shards) != 2 {
		t.Errorf("Expected 2 active shards, got %d", len(shards))
	}
}

func TestHashFunctions(t *testing.T) {
	hashFuncs := []HashFunction{CRC32Hash, FNV1Hash, FNV1aHash}

	for _, hashFunc := range hashFuncs {
		config := DefaultShardManagerConfig()
		config.HashFunc = hashFunc
		sm := NewShardManager(config)

		hash1 := sm.hash("test-key")
		hash2 := sm.hash("test-key")

		if hash1 != hash2 {
			t.Errorf("Hash function %s not consistent", hashFunc)
		}

		// Different keys should (likely) produce different hashes
		hash3 := sm.hash("different-key")
		if hash1 == hash3 {
			// This could happen but is unlikely
			t.Logf("Warning: Same hash for different keys with %s", hashFunc)
		}
	}
}

func TestExecuteOnShard(t *testing.T) {
	sm := NewShardManager(nil)

	// Create shard with test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create test table
	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO users (id, name) VALUES (1, 'Alice'), (2, 'Bob')")
	if err != nil {
		t.Fatal(err)
	}

	shard := &Shard{
		ID:     0,
		Name:   "shard-0",
		DB:     db,
		Active: true,
	}
	sm.AddShard(shard)

	// Execute query
	ctx := context.Background()
	rows, err := sm.ExecuteOnShard(ctx, "test-key", "SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("ExecuteOnShard failed: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	if count != 2 {
		t.Errorf("Expected 2 rows, got %d", count)
	}

	// Check query count
	if shard.QueryCount != 1 {
		t.Errorf("Expected query count 1, got %d", shard.QueryCount)
	}
}

func TestExecuteOnAllShards(t *testing.T) {
	config := DefaultShardManagerConfig()
	config.EnableCrossShard = true
	config.MaxCrossShardQueries = 5
	sm := NewShardManager(config)

	// Create 3 shards
	for i := 0; i < 3; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		// Create table
		db.Exec("CREATE TABLE users (id INTEGER, name TEXT)")
		db.Exec("INSERT INTO users VALUES (?, ?)", i, "User"+string(rune('A'+i)))

		shard := &Shard{
			ID:     i,
			Name:   "shard-" + string(rune('0'+i)),
			DB:     db,
			Active: true,
		}
		sm.AddShard(shard)
	}

	// Execute on all shards
	ctx := context.Background()
	results, err := sm.ExecuteOnAllShards(ctx, "SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("ExecuteOnAllShards failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Check each result
	for _, result := range results {
		if result.Error != nil {
			t.Errorf("Shard %d returned error: %v", result.ShardID, result.Error)
		}

		if result.Rows == nil {
			t.Errorf("Shard %d returned nil rows", result.ShardID)
			continue
		}

		// Count rows
		count := 0
		for result.Rows.Next() {
			count++
		}
		result.Rows.Close()

		if count != 1 {
			t.Errorf("Expected 1 row from shard %d, got %d", result.ShardID, count)
		}
	}
}

func TestExecuteOnAllShards_Disabled(t *testing.T) {
	config := DefaultShardManagerConfig()
	config.EnableCrossShard = false
	sm := NewShardManager(config)

	ctx := context.Background()
	_, err := sm.ExecuteOnAllShards(ctx, "SELECT 1")
	if err == nil {
		t.Error("Expected error when cross-shard queries disabled")
	}
}

func TestGetShardStats(t *testing.T) {
	sm := NewShardManager(nil)

	// Create 2 shards
	for i := 0; i < 2; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &Shard{
			ID:     i,
			Name:   "shard-" + string(rune('0'+i)),
			DB:     db,
			Active: true,
			Region: "us-east-1",
		}
		sm.AddShard(shard)
	}

	// Execute some queries
	ctx := context.Background()
	sm.mu.RLock()
	sm.shards[0].DB.Exec("CREATE TABLE test (id INTEGER)")
	sm.mu.RUnlock()

	sm.ExecuteOnShard(ctx, "key1", "SELECT 1")
	sm.ExecuteOnShard(ctx, "key2", "SELECT 1")

	stats := sm.GetShardStats()
	if len(stats) != 2 {
		t.Errorf("Expected 2 stat entries, got %d", len(stats))
	}

	totalQueries := int64(0)
	for _, stat := range stats {
		totalQueries += stat.QueryCount
	}

	if totalQueries != 2 {
		t.Errorf("Expected 2 total queries, got %d", totalQueries)
	}
}

func TestShardDistribution(t *testing.T) {
	sm := NewShardManager(nil)

	// Create 3 shards
	for i := 0; i < 3; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &Shard{
			ID:     i,
			Name:   "shard-" + string(rune('0'+i)),
			DB:     db,
			Active: true,
		}
		sm.AddShard(shard)
	}

	// Test distribution with sample keys
	keys := []string{
		"user:1", "user:2", "user:3", "user:4", "user:5",
		"user:6", "user:7", "user:8", "user:9", "user:10",
	}

	distribution := sm.ShardDistribution(keys)

	// Check that all keys are distributed
	total := 0
	for _, count := range distribution {
		total += count
	}

	if total != len(keys) {
		t.Errorf("Expected %d keys distributed, got %d", len(keys), total)
	}
}

func TestValidateShardKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"Valid key", "user:12345", false},
		{"Empty key", "", true},
		{"Long key", string(make([]byte, 300)), true},
		{"Normal key", "product:abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShardKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateShardKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetShardCount(t *testing.T) {
	sm := NewShardManager(nil)

	if sm.GetShardCount() != 0 {
		t.Errorf("Expected 0 shards initially, got %d", sm.GetShardCount())
	}

	// Add shards
	for i := 0; i < 5; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &Shard{
			ID:     i,
			Name:   "shard-" + string(rune('0'+i)),
			DB:     db,
			Active: true,
		}
		sm.AddShard(shard)
	}

	if sm.GetShardCount() != 5 {
		t.Errorf("Expected 5 shards, got %d", sm.GetShardCount())
	}
}

func TestGetActiveShardCount(t *testing.T) {
	sm := NewShardManager(nil)

	// Add shards
	for i := 0; i < 5; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &Shard{
			ID:     i,
			Name:   "shard-" + string(rune('0'+i)),
			DB:     db,
			Active: i < 3, // Only first 3 are active
		}
		sm.AddShard(shard)
	}

	count := sm.GetActiveShardCount()
	if count != 3 {
		t.Errorf("Expected 3 active shards, got %d", count)
	}
}

func TestConsistentHashDistribution(t *testing.T) {
	config := DefaultShardManagerConfig()
	config.Strategy = ConsistentHash
	config.VirtualNodes = 150
	sm := NewShardManager(config)

	// Create 4 shards
	for i := 0; i < 4; i++ {
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		shard := &Shard{
			ID:     i,
			Name:   "shard-" + string(rune('0'+i)),
			DB:     db,
			Active: true,
		}
		sm.AddShard(shard)
	}

	// Generate many keys and check distribution
	keys := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		keys[i] = "key:" + string(rune('0'+i))
	}

	distribution := sm.ShardDistribution(keys)

	// Each shard should get roughly 25% of keys (allowing for variance)
	for shardID, count := range distribution {
		percentage := float64(count) / float64(len(keys)) * 100

		if percentage < 10 || percentage > 40 {
			t.Errorf("Shard %d has poor distribution: %.1f%% (count: %d)",
				shardID, percentage, count)
		}
	}
}

func TestAddRangeMapping(t *testing.T) {
	sm := NewShardManager(nil)

	sm.AddRangeMapping(0, 0, 1000)
	sm.AddRangeMapping(1, 1000, 2000)
	sm.AddRangeMapping(2, 2000, 3000)

	if len(sm.ranges) != 3 {
		t.Errorf("Expected 3 range mappings, got %d", len(sm.ranges))
	}
}

func TestAddLookupMapping(t *testing.T) {
	sm := NewShardManager(nil)

	sm.AddLookupMapping("tenant:A", 0)
	sm.AddLookupMapping("tenant:B", 1)
	sm.AddLookupMapping("tenant:C", 0)

	if len(sm.lookupTable) != 3 {
		t.Errorf("Expected 3 lookup mappings, got %d", len(sm.lookupTable))
	}

	shardID, exists := sm.lookupTable["tenant:A"]
	if !exists || shardID != 0 {
		t.Error("Lookup mapping not found or incorrect")
	}
}
