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
	"database/sql"
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"sort"
	"sync"
)

// ShardingStrategy defines the strategy for shard key selection
type ShardingStrategy string

const (
	// HashModulo uses hash(key) % shard_count
	HashModulo ShardingStrategy = "hash_modulo"

	// ConsistentHash uses consistent hashing
	ConsistentHash ShardingStrategy = "consistent_hash"

	// RangeSharding uses range-based sharding
	RangeSharding ShardingStrategy = "range"

	// LookupSharding uses a lookup table
	LookupSharding ShardingStrategy = "lookup"
)

// HashFunction defines hash functions for shard key
type HashFunction string

const (
	// CRC32Hash uses CRC32 hash
	CRC32Hash HashFunction = "crc32"

	// FNV1Hash uses FNV-1 hash
	FNV1Hash HashFunction = "fnv1"

	// FNV1aHash uses FNV-1a hash
	FNV1aHash HashFunction = "fnv1a"
)

// Shard represents a single database shard
type Shard struct {
	ID       int
	Name     string
	DB       *sql.DB
	DBType   string
	Config   *DBConfig
	Active   bool
	Weight   int
	Region   string

	// For range sharding
	MinRange interface{}
	MaxRange interface{}

	// Statistics
	QueryCount int64
	ErrorCount int64

	mu sync.RWMutex
}

// ShardManagerConfig configures the shard manager
type ShardManagerConfig struct {
	// Strategy for sharding
	Strategy ShardingStrategy

	// HashFunc for hash-based strategies
	HashFunc HashFunction

	// VirtualNodes for consistent hashing
	VirtualNodes int

	// EnableCrossShard allows cross-shard queries
	EnableCrossShard bool

	// MaxCrossShardQueries limits concurrent cross-shard queries
	MaxCrossShardQueries int

	// CrossShardTimeout for cross-shard query timeout
	CrossShardTimeout int // seconds
}

// DefaultShardManagerConfig returns default configuration
func DefaultShardManagerConfig() *ShardManagerConfig {
	return &ShardManagerConfig{
		Strategy:             HashModulo,
		HashFunc:             CRC32Hash,
		VirtualNodes:         150,
		EnableCrossShard:     true,
		MaxCrossShardQueries: 10,
		CrossShardTimeout:    30,
	}
}

// ShardManager manages database sharding
type ShardManager struct {
	config *ShardManagerConfig
	shards []*Shard
	mu     sync.RWMutex

	// For consistent hashing
	ring       map[uint32]int // hash -> shard ID
	sortedKeys []uint32

	// For range sharding
	ranges []*ShardRange

	// For lookup sharding
	lookupTable map[string]int // shard key -> shard ID
}

// ShardRange represents a range for range-based sharding
type ShardRange struct {
	ShardID  int
	MinValue interface{}
	MaxValue interface{}
}

// NewShardManager creates a new shard manager
func NewShardManager(config *ShardManagerConfig) *ShardManager {
	if config == nil {
		config = DefaultShardManagerConfig()
	}

	return &ShardManager{
		config:      config,
		shards:      make([]*Shard, 0),
		ring:        make(map[uint32]int),
		sortedKeys:  make([]uint32, 0),
		ranges:      make([]*ShardRange, 0),
		lookupTable: make(map[string]int),
	}
}

// AddShard adds a shard to the manager
func (sm *ShardManager) AddShard(shard *Shard) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Validate shard
	if shard == nil {
		return fmt.Errorf("shard cannot be nil")
	}

	if shard.DB == nil {
		return fmt.Errorf("shard DB cannot be nil")
	}

	// Add shard
	sm.shards = append(sm.shards, shard)

	// Update consistent hash ring if needed
	if sm.config.Strategy == ConsistentHash {
		sm.addToRing(shard)
	}

	return nil
}

// RemoveShard removes a shard from the manager
func (sm *ShardManager) RemoveShard(shardID int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, shard := range sm.shards {
		if shard.ID == shardID {
			// Remove from slice
			sm.shards = append(sm.shards[:i], sm.shards[i+1:]...)

			// Remove from ring if needed
			if sm.config.Strategy == ConsistentHash {
				sm.removeFromRing(shard)
			}

			return nil
		}
	}

	return fmt.Errorf("shard not found: %d", shardID)
}

// GetShard returns the shard for a given shard key
func (sm *ShardManager) GetShard(shardKey string) (*Shard, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.shards) == 0 {
		return nil, fmt.Errorf("no shards available")
	}

	var shardID int

	switch sm.config.Strategy {
	case HashModulo:
		shardID = sm.hashModulo(shardKey)

	case ConsistentHash:
		shardID = sm.consistentHash(shardKey)

	case RangeSharding:
		var err error
		shardID, err = sm.rangeShard(shardKey)
		if err != nil {
			return nil, err
		}

	case LookupSharding:
		var err error
		shardID, err = sm.lookupShard(shardKey)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown sharding strategy: %s", sm.config.Strategy)
	}

	// Find shard by ID
	for _, shard := range sm.shards {
		if shard.ID == shardID && shard.Active {
			return shard, nil
		}
	}

	return nil, fmt.Errorf("shard not found or inactive: %d", shardID)
}

// GetAllShards returns all active shards
func (sm *ShardManager) GetAllShards() []*Shard {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	active := make([]*Shard, 0)
	for _, shard := range sm.shards {
		if shard.Active {
			active = append(active, shard)
		}
	}

	return active
}

// hashModulo uses hash(key) % shard_count
func (sm *ShardManager) hashModulo(key string) int {
	hash := sm.hash(key)
	return int(hash % uint32(len(sm.shards)))
}

// consistentHash uses consistent hashing
func (sm *ShardManager) consistentHash(key string) int {
	if len(sm.sortedKeys) == 0 {
		return 0
	}

	hash := sm.hash(key)

	// Binary search for the appropriate hash
	idx := sort.Search(len(sm.sortedKeys), func(i int) bool {
		return sm.sortedKeys[i] >= hash
	})

	// Wrap around if necessary
	if idx == len(sm.sortedKeys) {
		idx = 0
	}

	return sm.ring[sm.sortedKeys[idx]]
}

// rangeShard uses range-based sharding
func (sm *ShardManager) rangeShard(key string) (int, error) {
	// This would need type conversion based on actual range values
	// For simplicity, we'll use string comparison
	for _, r := range sm.ranges {
		// Simplified range check - production would need proper type handling
		if key >= fmt.Sprintf("%v", r.MinValue) && key < fmt.Sprintf("%v", r.MaxValue) {
			return r.ShardID, nil
		}
	}

	return 0, fmt.Errorf("no shard found for key: %s", key)
}

// lookupShard uses lookup table
func (sm *ShardManager) lookupShard(key string) (int, error) {
	shardID, exists := sm.lookupTable[key]
	if !exists {
		return 0, fmt.Errorf("no shard mapping found for key: %s", key)
	}

	return shardID, nil
}

// hash computes hash of a string using configured hash function
func (sm *ShardManager) hash(s string) uint32 {
	switch sm.config.HashFunc {
	case CRC32Hash:
		return crc32.ChecksumIEEE([]byte(s))

	case FNV1Hash, FNV1aHash:
		h := fnv.New32a()
		h.Write([]byte(s))
		return h.Sum32()

	default:
		return crc32.ChecksumIEEE([]byte(s))
	}
}

// addToRing adds a shard to the consistent hash ring
func (sm *ShardManager) addToRing(shard *Shard) {
	for i := 0; i < sm.config.VirtualNodes; i++ {
		virtualKey := fmt.Sprintf("%s-%d", shard.Name, i)
		hash := sm.hash(virtualKey)
		sm.ring[hash] = shard.ID
		sm.sortedKeys = append(sm.sortedKeys, hash)
	}

	sort.Slice(sm.sortedKeys, func(i, j int) bool {
		return sm.sortedKeys[i] < sm.sortedKeys[j]
	})
}

// removeFromRing removes a shard from the consistent hash ring
func (sm *ShardManager) removeFromRing(shard *Shard) {
	for i := 0; i < sm.config.VirtualNodes; i++ {
		virtualKey := fmt.Sprintf("%s-%d", shard.Name, i)
		hash := sm.hash(virtualKey)
		delete(sm.ring, hash)

		// Remove from sorted keys
		for idx, key := range sm.sortedKeys {
			if key == hash {
				sm.sortedKeys = append(sm.sortedKeys[:idx], sm.sortedKeys[idx+1:]...)
				break
			}
		}
	}
}

// ExecuteOnShard executes a query on a specific shard
func (sm *ShardManager) ExecuteOnShard(ctx context.Context, shardKey string, query string, args ...interface{}) (*sql.Rows, error) {
	shard, err := sm.GetShard(shardKey)
	if err != nil {
		return nil, err
	}

	shard.mu.Lock()
	shard.QueryCount++
	shard.mu.Unlock()

	rows, err := shard.DB.QueryContext(ctx, query, args...)

	if err != nil {
		shard.mu.Lock()
		shard.ErrorCount++
		shard.mu.Unlock()
		return nil, err
	}

	return rows, nil
}

// ExecuteOnAllShards executes a query on all shards (cross-shard query)
func (sm *ShardManager) ExecuteOnAllShards(ctx context.Context, query string, args ...interface{}) ([]*ShardQueryResult, error) {
	if !sm.config.EnableCrossShard {
		return nil, fmt.Errorf("cross-shard queries are disabled")
	}

	shards := sm.GetAllShards()
	if len(shards) == 0 {
		return nil, fmt.Errorf("no active shards")
	}

	// Use semaphore to limit concurrent queries
	sem := make(chan struct{}, sm.config.MaxCrossShardQueries)
	results := make([]*ShardQueryResult, len(shards))
	var wg sync.WaitGroup

	for i, shard := range shards {
		wg.Add(1)
		go func(idx int, s *Shard) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			result := &ShardQueryResult{
				ShardID:   s.ID,
				ShardName: s.Name,
			}

			s.mu.Lock()
			s.QueryCount++
			s.mu.Unlock()

			rows, err := s.DB.QueryContext(ctx, query, args...)
			if err != nil {
				result.Error = err
				s.mu.Lock()
				s.ErrorCount++
				s.mu.Unlock()
			} else {
				result.Rows = rows
			}

			results[idx] = result
		}(i, shard)
	}

	wg.Wait()

	return results, nil
}

// ExecuteOnShardTx executes a query within a transaction on a shard
func (sm *ShardManager) ExecuteOnShardTx(ctx context.Context, shardKey string, fn func(*sql.Tx) error) error {
	shard, err := sm.GetShard(shardKey)
	if err != nil {
		return err
	}

	tx, err := shard.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// AddRangeMapping adds a range mapping for range-based sharding
func (sm *ShardManager) AddRangeMapping(shardID int, minValue, maxValue interface{}) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.ranges = append(sm.ranges, &ShardRange{
		ShardID:  shardID,
		MinValue: minValue,
		MaxValue: maxValue,
	})

	// Sort ranges by min value for efficient lookup
	sort.Slice(sm.ranges, func(i, j int) bool {
		return fmt.Sprintf("%v", sm.ranges[i].MinValue) < fmt.Sprintf("%v", sm.ranges[j].MinValue)
	})
}

// AddLookupMapping adds a lookup mapping for lookup-based sharding
func (sm *ShardManager) AddLookupMapping(key string, shardID int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.lookupTable[key] = shardID
}

// GetShardStats returns statistics for all shards
func (sm *ShardManager) GetShardStats() []ShardStats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := make([]ShardStats, len(sm.shards))
	for i, shard := range sm.shards {
		shard.mu.RLock()
		stats[i] = ShardStats{
			ShardID:    shard.ID,
			ShardName:  shard.Name,
			Active:     shard.Active,
			QueryCount: shard.QueryCount,
			ErrorCount: shard.ErrorCount,
			Region:     shard.Region,
		}
		shard.mu.RUnlock()
	}

	return stats
}

// RebalanceShards rebalances data across shards (placeholder for complex logic)
func (sm *ShardManager) RebalanceShards(ctx context.Context) error {
	// This is a complex operation that would involve:
	// 1. Determining optimal data distribution
	// 2. Moving data between shards
	// 3. Updating consistent hash ring
	// 4. Handling in-flight queries
	// For now, this is a placeholder

	return fmt.Errorf("rebalancing not yet implemented")
}

// MigrateShard migrates a shard key from one shard to another
func (sm *ShardManager) MigrateShard(ctx context.Context, key string, targetShardID int) error {
	// This would involve:
	// 1. Reading data from source shard
	// 2. Writing to target shard
	// 3. Updating lookup table if using lookup sharding
	// 4. Deleting from source shard
	// For now, this is a placeholder

	return fmt.Errorf("shard migration not yet implemented")
}

// GetShardCount returns the number of shards
func (sm *ShardManager) GetShardCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.shards)
}

// GetActiveShardCount returns the number of active shards
func (sm *ShardManager) GetActiveShardCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := 0
	for _, shard := range sm.shards {
		if shard.Active {
			count++
		}
	}

	return count
}

// ShardQueryResult represents the result of a query on a shard
type ShardQueryResult struct {
	ShardID   int
	ShardName string
	Rows      *sql.Rows
	Error     error
}

// ShardStats represents statistics for a shard
type ShardStats struct {
	ShardID    int
	ShardName  string
	Active     bool
	QueryCount int64
	ErrorCount int64
	Region     string
}

// ShardDistribution calculates the distribution of keys across shards
func (sm *ShardManager) ShardDistribution(keys []string) map[int]int {
	distribution := make(map[int]int)

	for _, key := range keys {
		shard, err := sm.GetShard(key)
		if err != nil {
			continue
		}

		distribution[shard.ID]++
	}

	return distribution
}

// ValidateShardKey validates a shard key
func ValidateShardKey(key string) error {
	if key == "" {
		return fmt.Errorf("shard key cannot be empty")
	}

	if len(key) > 255 {
		return fmt.Errorf("shard key too long (max 255 characters)")
	}

	return nil
}
