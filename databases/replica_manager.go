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

package dbconn

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ReplicaLag represents replication lag information for a replica
type ReplicaLag struct {
	ReplicaName string
	LagSeconds  float64
	LastChecked time.Time
	IsHealthy   bool
	Error       error
}

// ReplicaHealth represents the health status of a replica
type ReplicaHealth struct {
	Name              string
	Active            bool
	Lag               *ReplicaLag
	ResponseTime      time.Duration
	ErrorCount        int
	LastError         error
	LastSuccessTime   time.Time
	ConsecutiveFails  int
	Weight            int
	EffectiveWeight   int // Adjusted based on health
}

// LoadBalancingStrategy defines the strategy for selecting replicas
type LoadBalancingStrategy string

const (
	// RoundRobin selects replicas in round-robin fashion
	RoundRobin LoadBalancingStrategy = "round_robin"

	// WeightedRoundRobin selects replicas based on their weight
	WeightedRoundRobin LoadBalancingStrategy = "weighted_round_robin"

	// LeastConnections selects replica with least active connections
	LeastConnections LoadBalancingStrategy = "least_connections"

	// Random selects a random healthy replica
	Random LoadBalancingStrategy = "random"

	// LeastLag selects replica with minimum replication lag
	LeastLag LoadBalancingStrategy = "least_lag"
)

// ReplicaManagerConfig configures the replica manager
type ReplicaManagerConfig struct {
	// Strategy for load balancing across replicas
	Strategy LoadBalancingStrategy

	// MaxReplicaLag is the maximum acceptable replication lag in seconds
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

// DefaultReplicaManagerConfig returns default configuration
func DefaultReplicaManagerConfig() *ReplicaManagerConfig {
	return &ReplicaManagerConfig{
		Strategy:              WeightedRoundRobin,
		MaxReplicaLag:         10.0, // 10 seconds
		LagCheckInterval:      5 * time.Second,
		FailoverThreshold:     3,
		RecoveryCheckInterval: 30 * time.Second,
		EnableAutoRecovery:    true,
		PreferLocalReplica:    false,
		LocalRegion:           "",
	}
}

// ReplicaManager manages read replica connections with advanced features
type ReplicaManager struct {
	config          *ReplicaManagerConfig
	replicaHealth   map[string]*ReplicaHealth
	mu              sync.RWMutex

	// For round-robin
	roundRobinIndex int

	// For weighted selection
	weightedPool    []string // Replica names repeated by weight

	// Background monitoring
	stopMonitoring  chan struct{}
	monitoringActive bool

	// Metrics
	totalRequests   int64
	failedRequests  int64
	lagChecks       int64
}

// NewReplicaManager creates a new replica manager
func NewReplicaManager(config *ReplicaManagerConfig) *ReplicaManager {
	if config == nil {
		config = DefaultReplicaManagerConfig()
	}

	return &ReplicaManager{
		config:         config,
		replicaHealth:  make(map[string]*ReplicaHealth),
		weightedPool:   make([]string, 0),
		stopMonitoring: make(chan struct{}),
	}
}

// RegisterReplica registers a replica for monitoring
func (rm *ReplicaManager) RegisterReplica(name string, weight int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if weight <= 0 {
		weight = 1
	}

	rm.replicaHealth[name] = &ReplicaHealth{
		Name:            name,
		Active:          true,
		Weight:          weight,
		EffectiveWeight: weight,
		LastSuccessTime: time.Now(),
	}

	// Rebuild weighted pool
	rm.rebuildWeightedPool()
}

// UnregisterReplica removes a replica from monitoring
func (rm *ReplicaManager) UnregisterReplica(name string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.replicaHealth, name)
	rm.rebuildWeightedPool()
}

// rebuildWeightedPool rebuilds the weighted replica pool
func (rm *ReplicaManager) rebuildWeightedPool() {
	rm.weightedPool = make([]string, 0)

	for name, health := range rm.replicaHealth {
		if health.Active && health.EffectiveWeight > 0 {
			// Add replica multiple times based on effective weight
			for i := 0; i < health.EffectiveWeight; i++ {
				rm.weightedPool = append(rm.weightedPool, name)
			}
		}
	}
}

// SelectReplica selects a replica based on the configured strategy
func (rm *ReplicaManager) SelectReplica() (string, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Get healthy replicas
	healthyReplicas := rm.getHealthyReplicas()

	if len(healthyReplicas) == 0 {
		return "", fmt.Errorf("no healthy replicas available")
	}

	var selected string

	switch rm.config.Strategy {
	case RoundRobin:
		selected = rm.selectRoundRobin(healthyReplicas)

	case WeightedRoundRobin:
		selected = rm.selectWeightedRoundRobin()

	case Random:
		selected = healthyReplicas[rand.Intn(len(healthyReplicas))]

	case LeastLag:
		selected = rm.selectLeastLag(healthyReplicas)

	default:
		selected = rm.selectWeightedRoundRobin()
	}

	if selected == "" {
		return "", fmt.Errorf("failed to select replica")
	}

	return selected, nil
}

// getHealthyReplicas returns list of healthy replica names
func (rm *ReplicaManager) getHealthyReplicas() []string {
	healthy := make([]string, 0)

	for name, health := range rm.replicaHealth {
		if !health.Active {
			continue
		}

		// Check replication lag if available
		if health.Lag != nil && rm.config.MaxReplicaLag > 0 {
			if health.Lag.LagSeconds > rm.config.MaxReplicaLag {
				continue // Skip replicas with too much lag
			}
		}

		healthy = append(healthy, name)
	}

	return healthy
}

// selectRoundRobin selects replica using round-robin
func (rm *ReplicaManager) selectRoundRobin(replicas []string) string {
	if len(replicas) == 0 {
		return ""
	}

	selected := replicas[rm.roundRobinIndex%len(replicas)]
	rm.roundRobinIndex++

	return selected
}

// selectWeightedRoundRobin selects replica using weighted round-robin
func (rm *ReplicaManager) selectWeightedRoundRobin() string {
	if len(rm.weightedPool) == 0 {
		return ""
	}

	selected := rm.weightedPool[rm.roundRobinIndex%len(rm.weightedPool)]
	rm.roundRobinIndex++

	return selected
}

// selectLeastLag selects replica with minimum replication lag
func (rm *ReplicaManager) selectLeastLag(replicas []string) string {
	var bestReplica string
	var minLag float64 = -1

	for _, name := range replicas {
		health := rm.replicaHealth[name]
		if health.Lag != nil {
			if minLag < 0 || health.Lag.LagSeconds < minLag {
				minLag = health.Lag.LagSeconds
				bestReplica = name
			}
		}
	}

	if bestReplica == "" && len(replicas) > 0 {
		return replicas[0]
	}

	return bestReplica
}

// RecordSuccess records a successful operation on a replica
func (rm *ReplicaManager) RecordSuccess(replicaName string, responseTime time.Duration) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	health, exists := rm.replicaHealth[replicaName]
	if !exists {
		return
	}

	health.ResponseTime = responseTime
	health.LastSuccessTime = time.Now()
	health.ConsecutiveFails = 0
	health.LastError = nil

	// Increase effective weight if it was reduced
	if health.EffectiveWeight < health.Weight {
		health.EffectiveWeight++
		rm.rebuildWeightedPool()
	}
}

// RecordFailure records a failed operation on a replica
func (rm *ReplicaManager) RecordFailure(replicaName string, err error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	health, exists := rm.replicaHealth[replicaName]
	if !exists {
		return
	}

	health.ErrorCount++
	health.LastError = err
	health.ConsecutiveFails++

	// Reduce effective weight
	if health.EffectiveWeight > 0 {
		health.EffectiveWeight--
		rm.rebuildWeightedPool()
	}

	// Mark as inactive if threshold exceeded
	if health.ConsecutiveFails >= rm.config.FailoverThreshold {
		health.Active = false
		health.EffectiveWeight = 0
		rm.rebuildWeightedPool()
	}
}

// UpdateReplicaLag updates the replication lag for a replica
func (rm *ReplicaManager) UpdateReplicaLag(replicaName string, lagSeconds float64, err error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	health, exists := rm.replicaHealth[replicaName]
	if !exists {
		return
	}

	health.Lag = &ReplicaLag{
		ReplicaName: replicaName,
		LagSeconds:  lagSeconds,
		LastChecked: time.Now(),
		IsHealthy:   lagSeconds <= rm.config.MaxReplicaLag,
		Error:       err,
	}

	rm.lagChecks++
}

// CheckReplicationLag checks replication lag for a specific database type
func (rm *ReplicaManager) CheckReplicationLag(ctx context.Context, db *sql.DB, dbType string) (float64, error) {
	var query string
	var lag float64

	switch dbType {
	case "mysql":
		// MySQL replication lag query
		query = "SHOW SLAVE STATUS"
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return 0, fmt.Errorf("failed to query slave status: %w", err)
		}
		defer rows.Close()

		if rows.Next() {
			// Get column names
			columns, err := rows.Columns()
			if err != nil {
				return 0, err
			}

			// Create scan targets
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return 0, err
			}

			// Find Seconds_Behind_Master column
			for i, col := range columns {
				if col == "Seconds_Behind_Master" {
					if values[i] != nil {
						if v, ok := values[i].(int64); ok {
							lag = float64(v)
						}
					}
					break
				}
			}
		}

	case "postgres":
		// PostgreSQL replication lag query
		query = "SELECT EXTRACT(EPOCH FROM (NOW() - pg_last_xact_replay_timestamp()))::FLOAT"
		err := db.QueryRowContext(ctx, query).Scan(&lag)
		if err != nil {
			// If this is not a standby, lag is 0
			if err == sql.ErrNoRows {
				return 0, nil
			}
			return 0, fmt.Errorf("failed to query replication lag: %w", err)
		}

	default:
		return 0, fmt.Errorf("replication lag monitoring not supported for database type: %s", dbType)
	}

	return lag, nil
}

// StartMonitoring starts background monitoring of replicas
func (rm *ReplicaManager) StartMonitoring(ctx context.Context, dbGetter func(string) (*sql.DB, string, error)) {
	rm.mu.Lock()
	if rm.monitoringActive {
		rm.mu.Unlock()
		return
	}
	rm.monitoringActive = true
	rm.mu.Unlock()

	// Lag monitoring goroutine
	go func() {
		ticker := time.NewTicker(rm.config.LagCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rm.checkAllReplicaLags(ctx, dbGetter)
			case <-rm.stopMonitoring:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	// Recovery monitoring goroutine
	if rm.config.EnableAutoRecovery {
		go func() {
			ticker := time.NewTicker(rm.config.RecoveryCheckInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					rm.checkReplicaRecovery(ctx, dbGetter)
				case <-rm.stopMonitoring:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

// StopMonitoring stops background monitoring
func (rm *ReplicaManager) StopMonitoring() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.monitoringActive {
		close(rm.stopMonitoring)
		rm.monitoringActive = false
		rm.stopMonitoring = make(chan struct{}) // Reset for potential restart
	}
}

// checkAllReplicaLags checks replication lag for all registered replicas
func (rm *ReplicaManager) checkAllReplicaLags(ctx context.Context, dbGetter func(string) (*sql.DB, string, error)) {
	rm.mu.RLock()
	replicas := make([]string, 0, len(rm.replicaHealth))
	for name := range rm.replicaHealth {
		replicas = append(replicas, name)
	}
	rm.mu.RUnlock()

	for _, replicaName := range replicas {
		db, dbType, err := dbGetter(replicaName)
		if err != nil {
			rm.UpdateReplicaLag(replicaName, 0, err)
			continue
		}

		lag, err := rm.CheckReplicationLag(ctx, db, dbType)
		rm.UpdateReplicaLag(replicaName, lag, err)
	}
}

// checkReplicaRecovery checks if failed replicas have recovered
func (rm *ReplicaManager) checkReplicaRecovery(ctx context.Context, dbGetter func(string) (*sql.DB, string, error)) {
	rm.mu.RLock()
	failedReplicas := make([]string, 0)
	for name, health := range rm.replicaHealth {
		if !health.Active {
			failedReplicas = append(failedReplicas, name)
		}
	}
	rm.mu.RUnlock()

	for _, replicaName := range failedReplicas {
		db, _, err := dbGetter(replicaName)
		if err != nil {
			continue
		}

		// Try to ping the database
		if err := db.PingContext(ctx); err == nil {
			// Replica recovered
			rm.mu.Lock()
			if health, exists := rm.replicaHealth[replicaName]; exists {
				health.Active = true
				health.ConsecutiveFails = 0
				health.EffectiveWeight = health.Weight
				rm.rebuildWeightedPool()
			}
			rm.mu.Unlock()
		}
	}
}

// GetReplicaHealth returns health information for all replicas
func (rm *ReplicaManager) GetReplicaHealth() map[string]*ReplicaHealth {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make(map[string]*ReplicaHealth)
	for name, health := range rm.replicaHealth {
		// Create a copy
		healthCopy := *health
		if health.Lag != nil {
			lagCopy := *health.Lag
			healthCopy.Lag = &lagCopy
		}
		result[name] = &healthCopy
	}

	return result
}

// GetHealthyReplicaCount returns the number of healthy replicas
func (rm *ReplicaManager) GetHealthyReplicaCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return len(rm.getHealthyReplicas())
}

// GetStats returns statistics about replica manager
func (rm *ReplicaManager) GetStats() ReplicaManagerStats {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return ReplicaManagerStats{
		TotalReplicas:   len(rm.replicaHealth),
		HealthyReplicas: len(rm.getHealthyReplicas()),
		TotalRequests:   rm.totalRequests,
		FailedRequests:  rm.failedRequests,
		LagChecks:       rm.lagChecks,
		Strategy:        string(rm.config.Strategy),
	}
}

// ReplicaManagerStats contains statistics about the replica manager
type ReplicaManagerStats struct {
	TotalReplicas   int
	HealthyReplicas int
	TotalRequests   int64
	FailedRequests  int64
	LagChecks       int64
	Strategy        string
}
