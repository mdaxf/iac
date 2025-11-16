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

package documents

import (
	"fmt"
	"sync"

	"github.com/mdaxf/iac/logger"
)

// DocumentDBFactory manages document database instance creation
type DocumentDBFactory struct {
	drivers      map[DocDBType]DocDBConstructor
	instances    map[string]DocumentDB
	mu           sync.RWMutex
	instancesMu  sync.RWMutex
	iLog         logger.Log
}

// DocDBConstructor is a function that creates a document database driver instance
type DocDBConstructor func(*DocDBConfig) (DocumentDB, error)

var (
	// globalDocFactory is the singleton factory instance
	globalDocFactory *DocumentDBFactory
	docFactoryOnce   sync.Once
)

// GetDocFactory returns the global document database factory instance
func GetDocFactory() *DocumentDBFactory {
	docFactoryOnce.Do(func() {
		globalDocFactory = &DocumentDBFactory{
			drivers:   make(map[DocDBType]DocDBConstructor),
			instances: make(map[string]DocumentDB),
			iLog:      logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DocumentDBFactory"},
		}
		// Register built-in drivers
		globalDocFactory.registerBuiltinDrivers()
	})
	return globalDocFactory
}

// RegisterDriver registers a document database driver constructor
func (f *DocumentDBFactory) RegisterDriver(dbType DocDBType, constructor DocDBConstructor) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.drivers[dbType] = constructor
	f.iLog.Info(fmt.Sprintf("Registered document database driver: %s", dbType))
}

// NewDocumentDB creates a new document database instance
func (f *DocumentDBFactory) NewDocumentDB(config *DocDBConfig) (DocumentDB, error) {
	if config == nil {
		return nil, fmt.Errorf("document database configuration is required")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document database configuration: %w", err)
	}

	f.mu.RLock()
	constructor, exists := f.drivers[config.Type]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("document database driver not found for type: %s", config.Type)
	}

	return constructor(config)
}

// GetOrCreateDB gets an existing document database instance or creates a new one
func (f *DocumentDBFactory) GetOrCreateDB(name string, config *DocDBConfig) (DocumentDB, error) {
	// Check if instance exists
	f.instancesMu.RLock()
	if db, exists := f.instances[name]; exists {
		f.instancesMu.RUnlock()
		return db, nil
	}
	f.instancesMu.RUnlock()

	// Create new instance
	db, err := f.NewDocumentDB(config)
	if err != nil {
		return nil, err
	}

	// Connect to database
	if err := db.Connect(config); err != nil {
		return nil, fmt.Errorf("failed to connect to document database: %w", err)
	}

	// Store instance
	f.instancesMu.Lock()
	f.instances[name] = db
	f.instancesMu.Unlock()

	f.iLog.Info(fmt.Sprintf("Created and connected to document database instance '%s' of type %s", name, config.Type))

	return db, nil
}

// GetDB retrieves an existing document database instance
func (f *DocumentDBFactory) GetDB(name string) (DocumentDB, error) {
	f.instancesMu.RLock()
	defer f.instancesMu.RUnlock()

	db, exists := f.instances[name]
	if !exists {
		return nil, fmt.Errorf("document database instance not found: %s", name)
	}

	return db, nil
}

// CloseDB closes and removes a document database instance
func (f *DocumentDBFactory) CloseDB(name string) error {
	f.instancesMu.Lock()
	defer f.instancesMu.Unlock()

	db, exists := f.instances[name]
	if !exists {
		return fmt.Errorf("document database instance not found: %s", name)
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close document database: %w", err)
	}

	delete(f.instances, name)
	f.iLog.Info(fmt.Sprintf("Closed document database instance '%s'", name))

	return nil
}

// CloseAll closes all document database instances
func (f *DocumentDBFactory) CloseAll() error {
	f.instancesMu.Lock()
	defer f.instancesMu.Unlock()

	var errors []error
	for name, db := range f.instances {
		if err := db.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close %s: %w", name, err))
		}
	}

	f.instances = make(map[string]DocumentDB)

	if len(errors) > 0 {
		return fmt.Errorf("errors closing document databases: %v", errors)
	}

	f.iLog.Info("Closed all document database instances")

	return nil
}

// ListInstances returns the names of all active document database instances
func (f *DocumentDBFactory) ListInstances() []string {
	f.instancesMu.RLock()
	defer f.instancesMu.RUnlock()

	names := make([]string, 0, len(f.instances))
	for name := range f.instances {
		names = append(names, name)
	}

	return names
}

// GetInstanceInfo returns information about a document database instance
func (f *DocumentDBFactory) GetInstanceInfo(name string) (*ConnectionInfo, error) {
	f.instancesMu.RLock()
	defer f.instancesMu.RUnlock()

	db, exists := f.instances[name]
	if !exists {
		return nil, fmt.Errorf("document database instance not found: %s", name)
	}

	// This would need to be extended to get actual connection info
	info := &ConnectionInfo{
		Type:        db.GetType(),
		IsConnected: db.IsConnected(),
	}

	return info, nil
}

// registerBuiltinDrivers registers the built-in document database drivers
// Individual adapter packages should override these using their init() functions
func (f *DocumentDBFactory) registerBuiltinDrivers() {
	// MongoDB driver (will be overridden by mongodb package init)
	f.RegisterDriver(DocDBTypeMongoDB, func(config *DocDBConfig) (DocumentDB, error) {
		return nil, fmt.Errorf("MongoDB driver not registered - import github.com/mdaxf/iac/documents/mongodb")
	})

	// PostgreSQL JSONB driver (will be overridden by postgres package init)
	f.RegisterDriver(DocDBTypePostgres, func(config *DocDBConfig) (DocumentDB, error) {
		return nil, fmt.Errorf("PostgreSQL JSONB driver not registered - import github.com/mdaxf/iac/documents/postgres")
	})
}

// Convenience functions

// NewDocumentDB creates a new document database instance using the global factory
func NewDocumentDB(config *DocDBConfig) (DocumentDB, error) {
	return GetDocFactory().NewDocumentDB(config)
}

// GetOrCreateDocDB gets or creates a document database instance using the global factory
func GetOrCreateDocDB(name string, config *DocDBConfig) (DocumentDB, error) {
	return GetDocFactory().GetOrCreateDB(name, config)
}

// GetDocDBInstance retrieves a document database instance using the global factory
func GetDocDBInstance(name string) (DocumentDB, error) {
	return GetDocFactory().GetDB(name)
}

// CloseDocDBInstance closes a document database instance using the global factory
func CloseDocDBInstance(name string) error {
	return GetDocFactory().CloseDB(name)
}

// CloseAllDocDBs closes all document database instances using the global factory
func CloseAllDocDBs() error {
	return GetDocFactory().CloseAll()
}

// ListDocDBInstances returns the names of all active document database instances
func ListDocDBInstances() []string {
	return GetDocFactory().ListInstances()
}

// DocDBManager provides a high-level interface for managing document databases
type DocDBManager struct {
	factory *DocumentDBFactory
}

// NewDocDBManager creates a new document database manager
func NewDocDBManager() *DocDBManager {
	return &DocDBManager{
		factory: GetDocFactory(),
	}
}

// Connect connects to a document database with the given configuration
func (m *DocDBManager) Connect(name string, config *DocDBConfig) (DocumentDB, error) {
	return m.factory.GetOrCreateDB(name, config)
}

// Get retrieves an existing document database instance
func (m *DocDBManager) Get(name string) (DocumentDB, error) {
	return m.factory.GetDB(name)
}

// Disconnect closes a document database connection
func (m *DocDBManager) Disconnect(name string) error {
	return m.factory.CloseDB(name)
}

// DisconnectAll closes all document database connections
func (m *DocDBManager) DisconnectAll() error {
	return m.factory.CloseAll()
}

// List returns the names of all active connections
func (m *DocDBManager) List() []string {
	return m.factory.ListInstances()
}

// HealthCheck checks the health of a document database instance
func (m *DocDBManager) HealthCheck(name string) error {
	db, err := m.Get(name)
	if err != nil {
		return err
	}

	if !db.IsConnected() {
		return fmt.Errorf("document database '%s' is not connected", name)
	}

	return nil
}

// HealthCheckAll checks the health of all document database instances
func (m *DocDBManager) HealthCheckAll() map[string]error {
	instances := m.List()
	results := make(map[string]error)

	for _, name := range instances {
		results[name] = m.HealthCheck(name)
	}

	return results
}

// Global manager instance
var globalDocDBManager *DocDBManager
var docDBManagerOnce sync.Once

// GetDocDBManager returns the global document database manager
func GetDocDBManager() *DocDBManager {
	docDBManagerOnce.Do(func() {
		globalDocDBManager = NewDocDBManager()
	})
	return globalDocDBManager
}
