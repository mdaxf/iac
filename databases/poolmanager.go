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
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mdaxf/iac/logger"
)

// ConnectionPool represents a named database connection
type ConnectionPool struct {
	Name     string
	Type     PoolType
	DB       RelationalDB
	Config   *DBConfig
	Priority int
	Weight   int
	Active   bool
	mu       sync.RWMutex
}

// PoolType represents the type of connection pool
type PoolType string

const (
	PoolTypePrimary PoolType = "primary"
	PoolTypeReplica PoolType = "replica"
	PoolTypeBackup  PoolType = "backup"
)

// PoolManager manages multiple database connection pools
type PoolManager struct {
	primary  *ConnectionPool
	replicas []*ConnectionPool
	backups  []*ConnectionPool

	// Load balancing
	replicaIndex uint32

	// Monitoring
	healthCheckInterval time.Duration
	stopHealthCheck     chan struct{}
	healthCheckRunning  bool

	mu    sync.RWMutex
	iLog  logger.Log
}

// NewPoolManager creates a new pool manager
func NewPoolManager() *PoolManager {
	return &PoolManager{
		replicas:            make([]*ConnectionPool, 0),
		backups:             make([]*ConnectionPool, 0),
		healthCheckInterval: 30 * time.Second,
		stopHealthCheck:     make(chan struct{}),
		iLog:                logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "PoolManager"},
	}
}

// SetPrimary sets the primary database connection
func (pm *PoolManager) SetPrimary(config *DBConfig) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pool, err := pm.createPool("primary", PoolTypePrimary, config, 100, 1)
	if err != nil {
		return fmt.Errorf("failed to create primary pool: %w", err)
	}

	pm.primary = pool
	pm.iLog.Info("Primary database connection established")

	return nil
}

// AddReplica adds a replica database connection
func (pm *PoolManager) AddReplica(name string, config *DBConfig, weight int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pool, err := pm.createPool(name, PoolTypeReplica, config, 50, weight)
	if err != nil {
		return fmt.Errorf("failed to create replica pool %s: %w", name, err)
	}

	pm.replicas = append(pm.replicas, pool)
	pm.iLog.Info(fmt.Sprintf("Replica database connection '%s' established", name))

	return nil
}

// AddBackup adds a backup database connection
func (pm *PoolManager) AddBackup(name string, config *DBConfig) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pool, err := pm.createPool(name, PoolTypeBackup, config, 10, 1)
	if err != nil {
		return fmt.Errorf("failed to create backup pool %s: %w", name, err)
	}

	pm.backups = append(pm.backups, pool)
	pm.iLog.Info(fmt.Sprintf("Backup database connection '%s' established", name))

	return nil
}

// createPool creates a new connection pool
func (pm *PoolManager) createPool(name string, poolType PoolType, config *DBConfig, priority, weight int) (*ConnectionPool, error) {
	db, err := GetFactory().NewRelationalDB(config)
	if err != nil {
		return nil, err
	}

	if err := db.Connect(config); err != nil {
		return nil, err
	}

	pool := &ConnectionPool{
		Name:     name,
		Type:     poolType,
		DB:       db,
		Config:   config,
		Priority: priority,
		Weight:   weight,
		Active:   true,
	}

	return pool, nil
}

// GetPrimary returns the primary database connection
func (pm *PoolManager) GetPrimary() (RelationalDB, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.primary == nil {
		return nil, fmt.Errorf("primary database not configured")
	}

	if !pm.primary.Active {
		return nil, fmt.Errorf("primary database is not active")
	}

	return pm.primary.DB, nil
}

// GetReplica returns a replica database connection using round-robin load balancing
func (pm *PoolManager) GetReplica() (RelationalDB, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.replicas) == 0 {
		// Fall back to primary if no replicas
		pm.iLog.Debug("No replicas available, using primary for read")
		if pm.primary == nil {
			return nil, fmt.Errorf("no database connections available")
		}
		return pm.primary.DB, nil
	}

	// Get active replicas
	activeReplicas := make([]*ConnectionPool, 0)
	for _, replica := range pm.replicas {
		if replica.Active {
			activeReplicas = append(activeReplicas, replica)
		}
	}

	if len(activeReplicas) == 0 {
		// Fall back to primary if no active replicas
		pm.iLog.Warn("No active replicas available, using primary for read")
		if pm.primary == nil {
			return nil, fmt.Errorf("no database connections available")
		}
		return pm.primary.DB, nil
	}

	// Round-robin load balancing
	index := atomic.AddUint32(&pm.replicaIndex, 1)
	replica := activeReplicas[index%uint32(len(activeReplicas))]

	return replica.DB, nil
}

// GetForRead returns a database connection for read operations (prefers replicas)
func (pm *PoolManager) GetForRead() (RelationalDB, error) {
	return pm.GetReplica()
}

// GetForWrite returns a database connection for write operations (always primary)
func (pm *PoolManager) GetForWrite() (RelationalDB, error) {
	return pm.GetPrimary()
}

// GetByName returns a connection pool by name
func (pm *PoolManager) GetByName(name string) (RelationalDB, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.primary != nil && pm.primary.Name == name {
		return pm.primary.DB, nil
	}

	for _, replica := range pm.replicas {
		if replica.Name == name {
			return replica.DB, nil
		}
	}

	for _, backup := range pm.backups {
		if backup.Name == name {
			return backup.DB, nil
		}
	}

	return nil, fmt.Errorf("connection pool not found: %s", name)
}

// StartHealthCheck starts periodic health checks for all connections
func (pm *PoolManager) StartHealthCheck(ctx context.Context) {
	pm.mu.Lock()
	if pm.healthCheckRunning {
		pm.mu.Unlock()
		return
	}
	pm.healthCheckRunning = true
	pm.mu.Unlock()

	pm.iLog.Info("Starting database health checks")

	go func() {
		ticker := time.NewTicker(pm.healthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				pm.performHealthChecks(ctx)
			case <-pm.stopHealthCheck:
				pm.iLog.Info("Stopping database health checks")
				return
			case <-ctx.Done():
				pm.iLog.Info("Context cancelled, stopping health checks")
				return
			}
		}
	}()
}

// StopHealthCheck stops the health check goroutine
func (pm *PoolManager) StopHealthCheck() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.healthCheckRunning {
		close(pm.stopHealthCheck)
		pm.healthCheckRunning = false
	}
}

// performHealthChecks checks the health of all connections
func (pm *PoolManager) performHealthChecks(ctx context.Context) {
	pm.mu.RLock()
	pools := make([]*ConnectionPool, 0)

	if pm.primary != nil {
		pools = append(pools, pm.primary)
	}
	pools = append(pools, pm.replicas...)
	pools = append(pools, pm.backups...)

	pm.mu.RUnlock()

	for _, pool := range pools {
		go pm.checkPoolHealth(pool)
	}
}

// checkPoolHealth checks the health of a single pool
func (pm *PoolManager) checkPoolHealth(pool *ConnectionPool) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	err := pool.DB.Ping()
	wasActive := pool.Active

	if err != nil {
		pool.Active = false
		if wasActive {
			pm.iLog.Error(fmt.Sprintf("Connection pool '%s' became unhealthy: %v", pool.Name, err))
		}
	} else {
		pool.Active = true
		if !wasActive {
			pm.iLog.Info(fmt.Sprintf("Connection pool '%s' recovered", pool.Name))
		}
	}
}

// GetStats returns statistics for all connection pools
func (pm *PoolManager) GetStats() map[string]ConnectionStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := make(map[string]ConnectionStats)

	if pm.primary != nil {
		stats[pm.primary.Name] = GetStats(pm.primary.DB.Stats())
	}

	for _, replica := range pm.replicas {
		stats[replica.Name] = GetStats(replica.DB.Stats())
	}

	for _, backup := range pm.backups {
		stats[backup.Name] = GetStats(backup.DB.Stats())
	}

	return stats
}

// GetPoolInfo returns information about all pools
func (pm *PoolManager) GetPoolInfo() []PoolInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	info := make([]PoolInfo, 0)

	if pm.primary != nil {
		info = append(info, pm.getPoolInfo(pm.primary))
	}

	for _, replica := range pm.replicas {
		info = append(info, pm.getPoolInfo(replica))
	}

	for _, backup := range pm.backups {
		info = append(info, pm.getPoolInfo(backup))
	}

	return info
}

// getPoolInfo converts a ConnectionPool to PoolInfo
func (pm *PoolManager) getPoolInfo(pool *ConnectionPool) PoolInfo {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return PoolInfo{
		Name:     pool.Name,
		Type:     pool.Type,
		DBType:   pool.Config.Type,
		Host:     pool.Config.Host,
		Database: pool.Config.Database,
		Active:   pool.Active,
		Priority: pool.Priority,
		Weight:   pool.Weight,
	}
}

// CloseAll closes all database connections
func (pm *PoolManager) CloseAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.StopHealthCheck()

	var errors []error

	if pm.primary != nil {
		if err := pm.primary.DB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close primary: %w", err))
		}
	}

	for _, replica := range pm.replicas {
		if err := replica.DB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close replica %s: %w", replica.Name, err))
		}
	}

	for _, backup := range pm.backups {
		if err := backup.DB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close backup %s: %w", backup.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing connections: %v", errors)
	}

	pm.iLog.Info("All database connections closed")
	return nil
}

// PoolInfo contains information about a connection pool
type PoolInfo struct {
	Name     string
	Type     PoolType
	DBType   DBType
	Host     string
	Database string
	Active   bool
	Priority int
	Weight   int
}

// Global pool manager instance
var (
	globalPoolManager *PoolManager
	poolManagerOnce   sync.Once
)

// GetPoolManager returns the global pool manager instance
func GetPoolManager() *PoolManager {
	poolManagerOnce.Do(func() {
		globalPoolManager = NewPoolManager()
	})
	return globalPoolManager
}
