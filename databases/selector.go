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

	"github.com/mdaxf/iac/logger"
)

// OperationType represents the type of database operation
type OperationType string

const (
	OperationRead  OperationType = "read"
	OperationWrite OperationType = "write"
	OperationBulk  OperationType = "bulk"
	OperationAny   OperationType = "any"
)

// DatabaseContext represents context for database selection
type DatabaseContext struct {
	Operation   OperationType
	TenantID    string
	UserID      string
	Priority    int
	Metadata    map[string]interface{}
}

// SelectionStrategy represents the strategy for selecting a database
type SelectionStrategy string

const (
	StrategyRoundRobin    SelectionStrategy = "round_robin"
	StrategyLeastConn     SelectionStrategy = "least_conn"
	StrategyPriority      SelectionStrategy = "priority"
	StrategyRandom        SelectionStrategy = "random"
	StrategyConsistentHash SelectionStrategy = "consistent_hash"
)

// DatabaseSelector selects appropriate database based on context
type DatabaseSelector struct {
	poolManager *PoolManager
	strategy    SelectionStrategy
	rules       []SelectionRule
	mu          sync.RWMutex
	iLog        logger.Log
}

// SelectionRule represents a rule for database selection
type SelectionRule struct {
	Name      string
	Priority  int
	Condition func(*DatabaseContext) bool
	Action    func(*DatabaseContext) (string, error)
}

// NewDatabaseSelector creates a new database selector
func NewDatabaseSelector(poolManager *PoolManager) *DatabaseSelector {
	return &DatabaseSelector{
		poolManager: poolManager,
		strategy:    StrategyRoundRobin,
		rules:       make([]SelectionRule, 0),
		iLog:        logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DatabaseSelector"},
	}
}

// SetStrategy sets the selection strategy
func (ds *DatabaseSelector) SetStrategy(strategy SelectionStrategy) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.strategy = strategy
}

// AddRule adds a selection rule
func (ds *DatabaseSelector) AddRule(rule SelectionRule) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.rules = append(ds.rules, rule)
}

// SelectDatabase selects a database based on context
func (ds *DatabaseSelector) SelectDatabase(ctx *DatabaseContext) (RelationalDB, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	// Apply rules first
	for _, rule := range ds.rules {
		if rule.Condition(ctx) {
			poolName, err := rule.Action(ctx)
			if err != nil {
				continue
			}
			if db, err := ds.poolManager.GetByName(poolName); err == nil {
				ds.iLog.Debug(fmt.Sprintf("Selected database '%s' by rule '%s'", poolName, rule.Name))
				return db, nil
			}
		}
	}

	// Fall back to strategy-based selection
	return ds.selectByStrategy(ctx)
}

// selectByStrategy selects database based on configured strategy
func (ds *DatabaseSelector) selectByStrategy(ctx *DatabaseContext) (RelationalDB, error) {
	switch ctx.Operation {
	case OperationRead:
		return ds.poolManager.GetForRead()
	case OperationWrite, OperationBulk:
		return ds.poolManager.GetForWrite()
	default:
		return ds.poolManager.GetPrimary()
	}
}

// SelectForRead selects a database for read operations
func (ds *DatabaseSelector) SelectForRead() (RelationalDB, error) {
	return ds.SelectDatabase(&DatabaseContext{
		Operation: OperationRead,
	})
}

// SelectForWrite selects a database for write operations
func (ds *DatabaseSelector) SelectForWrite() (RelationalDB, error) {
	return ds.SelectDatabase(&DatabaseContext{
		Operation: OperationWrite,
	})
}

// ContextKey is the type for context keys
type ContextKey string

const (
	// DatabaseContextKey is the context key for database selection
	DatabaseContextKey ContextKey = "db_context"
	// SelectedDatabaseKey is the context key for the selected database
	SelectedDatabaseKey ContextKey = "selected_db"
)

// WithDatabaseContext adds database context to context.Context
func WithDatabaseContext(ctx context.Context, dbCtx *DatabaseContext) context.Context {
	return context.WithValue(ctx, DatabaseContextKey, dbCtx)
}

// GetDatabaseContext retrieves database context from context.Context
func GetDatabaseContext(ctx context.Context) (*DatabaseContext, bool) {
	dbCtx, ok := ctx.Value(DatabaseContextKey).(*DatabaseContext)
	return dbCtx, ok
}

// WithDatabase adds a database to context.Context
func WithDatabase(ctx context.Context, db RelationalDB) context.Context {
	return context.WithValue(ctx, SelectedDatabaseKey, db)
}

// GetDatabase retrieves database from context.Context
func GetDatabase(ctx context.Context) (RelationalDB, bool) {
	db, ok := ctx.Value(SelectedDatabaseKey).(RelationalDB)
	return db, ok
}

// DatabaseRouter routes requests to appropriate databases
type DatabaseRouter struct {
	selector      *DatabaseSelector
	tenantDBs     map[string]string // tenant ID -> database pool name
	routingRules  []RoutingRule
	mu            sync.RWMutex
	iLog          logger.Log
}

// RoutingRule represents a routing rule
type RoutingRule struct {
	Name      string
	Match     func(context.Context) bool
	Route     string
	Priority  int
}

// NewDatabaseRouter creates a new database router
func NewDatabaseRouter(selector *DatabaseSelector) *DatabaseRouter {
	return &DatabaseRouter{
		selector:     selector,
		tenantDBs:    make(map[string]string),
		routingRules: make([]RoutingRule, 0),
		iLog:         logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DatabaseRouter"},
	}
}

// RegisterTenant registers a tenant with a specific database pool
func (dr *DatabaseRouter) RegisterTenant(tenantID, poolName string) {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	dr.tenantDBs[tenantID] = poolName
	dr.iLog.Info(fmt.Sprintf("Registered tenant '%s' to pool '%s'", tenantID, poolName))
}

// AddRoutingRule adds a routing rule
func (dr *DatabaseRouter) AddRoutingRule(rule RoutingRule) {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	dr.routingRules = append(dr.routingRules, rule)
}

// Route routes a request to appropriate database
func (dr *DatabaseRouter) Route(ctx context.Context) (RelationalDB, error) {
	// Get database context
	dbCtx, ok := GetDatabaseContext(ctx)
	if !ok {
		// Create default context
		dbCtx = &DatabaseContext{
			Operation: OperationAny,
		}
	}

	// Check routing rules
	dr.mu.RLock()
	for _, rule := range dr.routingRules {
		if rule.Match(ctx) {
			dr.mu.RUnlock()
			return dr.selector.poolManager.GetByName(rule.Route)
		}
	}
	dr.mu.RUnlock()

	// Check tenant-specific routing
	if dbCtx.TenantID != "" {
		dr.mu.RLock()
		if poolName, ok := dr.tenantDBs[dbCtx.TenantID]; ok {
			dr.mu.RUnlock()
			return dr.selector.poolManager.GetByName(poolName)
		}
		dr.mu.RUnlock()
	}

	// Use selector
	return dr.selector.SelectDatabase(dbCtx)
}

// Middleware creates a middleware for automatic database selection
func (dr *DatabaseRouter) Middleware(next func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		// Select database
		db, err := dr.Route(ctx)
		if err != nil {
			return fmt.Errorf("failed to route database: %w", err)
		}

		// Add to context
		ctx = WithDatabase(ctx, db)

		// Call next
		return next(ctx)
	}
}

// ReadWriteSplitter splits read and write operations
type ReadWriteSplitter struct {
	selector *DatabaseSelector
	iLog     logger.Log
}

// NewReadWriteSplitter creates a new read/write splitter
func NewReadWriteSplitter(selector *DatabaseSelector) *ReadWriteSplitter {
	return &ReadWriteSplitter{
		selector: selector,
		iLog:     logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "ReadWriteSplitter"},
	}
}

// ForRead returns a database connection for read operations
func (rws *ReadWriteSplitter) ForRead(ctx context.Context) (RelationalDB, error) {
	dbCtx := &DatabaseContext{
		Operation: OperationRead,
	}

	// Merge with existing context if available
	if existingCtx, ok := GetDatabaseContext(ctx); ok {
		dbCtx.TenantID = existingCtx.TenantID
		dbCtx.UserID = existingCtx.UserID
		dbCtx.Priority = existingCtx.Priority
		dbCtx.Metadata = existingCtx.Metadata
	}

	return rws.selector.SelectDatabase(dbCtx)
}

// ForWrite returns a database connection for write operations
func (rws *ReadWriteSplitter) ForWrite(ctx context.Context) (RelationalDB, error) {
	dbCtx := &DatabaseContext{
		Operation: OperationWrite,
	}

	// Merge with existing context if available
	if existingCtx, ok := GetDatabaseContext(ctx); ok {
		dbCtx.TenantID = existingCtx.TenantID
		dbCtx.UserID = existingCtx.UserID
		dbCtx.Priority = existingCtx.Priority
		dbCtx.Metadata = existingCtx.Metadata
	}

	return rws.selector.SelectDatabase(dbCtx)
}

// Global instances
var (
	globalSelector *DatabaseSelector
	globalRouter   *DatabaseRouter
	globalSplitter *ReadWriteSplitter
	selectorOnce   sync.Once
)

// GetGlobalSelector returns the global database selector
func GetGlobalSelector() *DatabaseSelector {
	selectorOnce.Do(func() {
		globalSelector = NewDatabaseSelector(GetPoolManager())
		globalRouter = NewDatabaseRouter(globalSelector)
		globalSplitter = NewReadWriteSplitter(globalSelector)
	})
	return globalSelector
}

// GetGlobalRouter returns the global database router
func GetGlobalRouter() *DatabaseRouter {
	GetGlobalSelector() // Ensure initialized
	return globalRouter
}

// GetGlobalSplitter returns the global read/write splitter
func GetGlobalSplitter() *ReadWriteSplitter {
	GetGlobalSelector() // Ensure initialized
	return globalSplitter
}

// Example usage:
func ExampleDatabaseSelection() {
	// Get selector
	selector := GetGlobalSelector()

	// Add custom rule
	selector.AddRule(SelectionRule{
		Name:     "Admin to Primary",
		Priority: 10,
		Condition: func(ctx *DatabaseContext) bool {
			if ctx.Metadata == nil {
				return false
			}
			role, ok := ctx.Metadata["role"].(string)
			return ok && role == "admin"
		},
		Action: func(ctx *DatabaseContext) (string, error) {
			return "primary", nil
		},
	})

	// Use selector
	ctx := &DatabaseContext{
		Operation: OperationRead,
		UserID:    "user123",
		Metadata: map[string]interface{}{
			"role": "user",
		},
	}

	db, err := selector.SelectDatabase(ctx)
	if err != nil {
		fmt.Printf("Selection error: %v\n", err)
		return
	}

	// Use database
	_ = db

	// Using router
	router := GetGlobalRouter()
	router.RegisterTenant("tenant1", "replica_1")

	bgCtx := context.Background()
	dbCtx := &DatabaseContext{
		TenantID: "tenant1",
	}
	bgCtx = WithDatabaseContext(bgCtx, dbCtx)

	db, err = router.Route(bgCtx)
	if err != nil {
		fmt.Printf("Routing error: %v\n", err)
		return
	}

	// Using read/write splitter
	splitter := GetGlobalSplitter()

	// For read
	readDB, _ := splitter.ForRead(bgCtx)
	_ = readDB

	// For write
	writeDB, _ := splitter.ForWrite(bgCtx)
	_ = writeDB
}
