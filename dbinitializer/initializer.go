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

package dbinitializer

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"

	// Import database adapters to register them
	_ "github.com/mdaxf/iac/databases/mssql"
	_ "github.com/mdaxf/iac/databases/mysql"
	_ "github.com/mdaxf/iac/databases/oracle"
	_ "github.com/mdaxf/iac/databases/postgres"

	// Import document database adapters
	_ "github.com/mdaxf/iac/documents/mongodb"
	_ "github.com/mdaxf/iac/documents/postgres"
)

// DatabaseInitializer manages database initialization
type DatabaseInitializer struct {
	RelationalDBs map[string]dbconn.RelationalDB
	DocumentDBs   map[string]documents.DocumentDB
	PoolManager   *dbconn.PoolManager
	DocManager    *documents.DocDBManager
	iLog          logger.Log
}

// NewDatabaseInitializer creates a new database initializer
func NewDatabaseInitializer() *DatabaseInitializer {
	return &DatabaseInitializer{
		RelationalDBs: make(map[string]dbconn.RelationalDB),
		DocumentDBs:   make(map[string]documents.DocumentDB),
		PoolManager:   dbconn.GetPoolManager(),
		DocManager:    documents.GetDocDBManager(),
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DatabaseInitializer"},
	}
}

// InitializeFromEnvironment initializes databases from environment variables
func (di *DatabaseInitializer) InitializeFromEnvironment() error {
	di.iLog.Info("Initializing databases from environment variables")

	// Initialize relational databases
	if err := di.initRelationalDatabases(); err != nil {
		return fmt.Errorf("failed to initialize relational databases: %w", err)
	}

	// Initialize document databases
	if err := di.initDocumentDatabases(); err != nil {
		return fmt.Errorf("failed to initialize document databases: %w", err)
	}

	// Start health checks
	di.startHealthMonitoring()

	di.iLog.Info("Database initialization completed successfully")
	return nil
}

// initRelationalDatabases initializes relational databases from environment
func (di *DatabaseInitializer) initRelationalDatabases() error {
	// Check for primary database configuration
	primaryType := os.Getenv("DB_TYPE")
	if primaryType == "" {
		primaryType = "mysql" // Default to MySQL
	}

	primaryConfig := di.loadRelationalDBConfig("DB", dbconn.DBType(primaryType))

	// Create and connect primary database
	primaryDB, err := dbconn.NewRelationalDB(primaryConfig)
	if err != nil {
		return fmt.Errorf("failed to create primary database: %w", err)
	}

	if err := primaryDB.Connect(primaryConfig); err != nil {
		return fmt.Errorf("failed to connect to primary database: %w", err)
	}

	// Set primary in pool manager
	if err := di.PoolManager.SetPrimary(primaryConfig); err != nil {
		return fmt.Errorf("failed to set primary database: %w", err)
	}

	di.RelationalDBs["primary"] = primaryDB
	di.iLog.Info(fmt.Sprintf("Primary database initialized: %s", primaryType))

	// Check for replica databases
	replicaCount := di.getEnvInt("DB_REPLICA_COUNT", 0)
	for i := 1; i <= replicaCount; i++ {
		prefix := fmt.Sprintf("DB_REPLICA_%d", i)
		replicaType := os.Getenv(prefix + "_TYPE")
		if replicaType == "" {
			replicaType = primaryType
		}

		replicaConfig := di.loadRelationalDBConfig(prefix, dbconn.DBType(replicaType))
		replicaName := fmt.Sprintf("replica_%d", i)

		// Add replica to pool manager
		if err := di.PoolManager.AddReplica(replicaName, replicaConfig, 1); err != nil {
			di.iLog.Error(fmt.Sprintf("Failed to add replica %s: %v", replicaName, err))
			continue
		}

		di.iLog.Info(fmt.Sprintf("Replica database %d initialized: %s", i, replicaType))
	}

	return nil
}

// initDocumentDatabases initializes document databases from environment
func (di *DatabaseInitializer) initDocumentDatabases() error {
	// Check for document database configuration
	docDBType := os.Getenv("DOCDB_TYPE")
	if docDBType == "" {
		docDBType = "mongodb" // Default to MongoDB
	}

	docConfig := di.loadDocumentDBConfig("DOCDB", documents.DocDBType(docDBType))

	// Create and connect document database
	docDB, err := documents.NewDocumentDB(docConfig)
	if err != nil {
		return fmt.Errorf("failed to create document database: %w", err)
	}

	if err := docDB.Connect(docConfig); err != nil {
		return fmt.Errorf("failed to connect to document database: %w", err)
	}

	di.DocumentDBs["primary"] = docDB
	di.iLog.Info(fmt.Sprintf("Document database initialized: %s", docDBType))

	return nil
}

// loadRelationalDBConfig loads relational database config from environment
func (di *DatabaseInitializer) loadRelationalDBConfig(prefix string, dbType dbconn.DBType) *dbconn.DBConfig {
	config := &dbconn.DBConfig{
		Type:     dbType,
		Host:     di.getEnvString(prefix+"_HOST", "localhost"),
		Port:     di.getEnvInt(prefix+"_PORT", di.getDefaultPort(dbType)),
		Database: di.getEnvString(prefix+"_DATABASE", "iac"),
		Username: di.getEnvString(prefix+"_USERNAME", "root"),
		Password: di.getEnvString(prefix+"_PASSWORD", ""),
		SSLMode:  di.getEnvString(prefix+"_SSL_MODE", ""),

		MaxIdleConns:    di.getEnvInt(prefix+"_MAX_IDLE_CONNS", 5),
		MaxOpenConns:    di.getEnvInt(prefix+"_MAX_OPEN_CONNS", 10),
		ConnMaxLifetime: time.Duration(di.getEnvInt(prefix+"_CONN_MAX_LIFETIME", 3600)) * time.Second,
		ConnMaxIdleTime: time.Duration(di.getEnvInt(prefix+"_CONN_MAX_IDLE_TIME", 600)) * time.Second,
		ConnTimeout:     di.getEnvInt(prefix+"_CONN_TIMEOUT", 30),

		Options: make(map[string]string),
	}

	return config
}

// loadDocumentDBConfig loads document database config from environment
func (di *DatabaseInitializer) loadDocumentDBConfig(prefix string, dbType documents.DocDBType) *documents.DocDBConfig {
	config := &documents.DocDBConfig{
		Type:         dbType,
		Host:         di.getEnvString(prefix+"_HOST", "localhost"),
		Port:         di.getEnvInt(prefix+"_PORT", di.getDefaultDocPort(dbType)),
		Database:     di.getEnvString(prefix+"_DATABASE", "iac"),
		Username:     di.getEnvString(prefix+"_USERNAME", ""),
		Password:     di.getEnvString(prefix+"_PASSWORD", ""),
		SSLMode:      di.getEnvString(prefix+"_SSL_MODE", ""),
		AuthSource:   di.getEnvString(prefix+"_AUTH_SOURCE", "admin"),
		ReplicaSet:   di.getEnvString(prefix+"_REPLICA_SET", ""),
		MaxPoolSize:  di.getEnvInt(prefix+"_MAX_POOL_SIZE", 100),
		MinPoolSize:  di.getEnvInt(prefix+"_MIN_POOL_SIZE", 10),
		ConnTimeout:  di.getEnvInt(prefix+"_CONN_TIMEOUT", 30),
		Options:      make(map[string]string),
	}

	return config
}

// getDefaultPort returns the default port for a database type
func (di *DatabaseInitializer) getDefaultPort(dbType dbconn.DBType) int {
	switch dbType {
	case dbconn.DBTypeMySQL:
		return 3306
	case dbconn.DBTypePostgreSQL:
		return 5432
	case dbconn.DBTypeMSSQL:
		return 1433
	case dbconn.DBTypeOracle:
		return 1521
	default:
		return 3306
	}
}

// getDefaultDocPort returns the default port for a document database type
func (di *DatabaseInitializer) getDefaultDocPort(dbType documents.DocDBType) int {
	switch dbType {
	case documents.DocDBTypeMongoDB:
		return 27017
	case documents.DocDBTypePostgres:
		return 5432
	default:
		return 27017
	}
}

// Helper functions for environment variables
func (di *DatabaseInitializer) getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (di *DatabaseInitializer) getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// startHealthMonitoring starts health monitoring for all databases
func (di *DatabaseInitializer) startHealthMonitoring() {
	ctx := context.Background()

	// Start pool manager health checks
	di.PoolManager.StartHealthCheck(ctx)
	di.iLog.Info("Started health monitoring for relational databases")

	// Could add document database health monitoring here
}

// Shutdown gracefully shuts down all database connections
func (di *DatabaseInitializer) Shutdown() error {
	di.iLog.Info("Shutting down database connections")

	// Stop health checks
	di.PoolManager.StopHealthCheck()

	// Close all relational databases
	if err := di.PoolManager.CloseAll(); err != nil {
		di.iLog.Error(fmt.Sprintf("Error closing relational databases: %v", err))
	}

	// Close all document databases
	for name, db := range di.DocumentDBs {
		if err := db.Close(); err != nil {
			di.iLog.Error(fmt.Sprintf("Error closing document database %s: %v", name, err))
		}
	}

	di.iLog.Info("All database connections closed")
	return nil
}

// GetPrimaryDB returns the primary relational database
func (di *DatabaseInitializer) GetPrimaryDB() (dbconn.RelationalDB, error) {
	return di.PoolManager.GetPrimary()
}

// GetReplicaDB returns a replica database for read operations
func (di *DatabaseInitializer) GetReplicaDB() (dbconn.RelationalDB, error) {
	return di.PoolManager.GetReplica()
}

// GetDocumentDB returns the primary document database
func (di *DatabaseInitializer) GetDocumentDB() (documents.DocumentDB, error) {
	if db, exists := di.DocumentDBs["primary"]; exists {
		return db, nil
	}
	return nil, fmt.Errorf("primary document database not initialized")
}

// GetStats returns statistics for all databases
func (di *DatabaseInitializer) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Relational database stats
	relationalStats := di.PoolManager.GetStats()
	stats["relational"] = relationalStats

	// Document database stats
	docStats := make(map[string]interface{})
	for name, db := range di.DocumentDBs {
		if dbStats, err := db.Stats(context.Background()); err == nil {
			docStats[name] = dbStats
		}
	}
	stats["document"] = docStats

	return stats
}

// InitializeWithConfig initializes databases with explicit configuration
func (di *DatabaseInitializer) InitializeWithConfig(config *DatabaseConfig) error {
	di.iLog.Info("Initializing databases with provided configuration")

	// Initialize relational databases
	if config.Primary != nil {
		primaryDB, err := dbconn.NewRelationalDB(config.Primary)
		if err != nil {
			return fmt.Errorf("failed to create primary database: %w", err)
		}

		if err := primaryDB.Connect(config.Primary); err != nil {
			return fmt.Errorf("failed to connect to primary database: %w", err)
		}

		if err := di.PoolManager.SetPrimary(config.Primary); err != nil {
			return fmt.Errorf("failed to set primary database: %w", err)
		}

		di.RelationalDBs["primary"] = primaryDB
		di.iLog.Info(fmt.Sprintf("Primary database initialized: %s", config.Primary.Type))
	}

	// Initialize replicas
	for i, replicaConfig := range config.Replicas {
		replicaName := fmt.Sprintf("replica_%d", i+1)
		if err := di.PoolManager.AddReplica(replicaName, replicaConfig, 1); err != nil {
			di.iLog.Error(fmt.Sprintf("Failed to add replica %s: %v", replicaName, err))
		}
	}

	// Initialize document database
	if config.DocumentDB != nil {
		docDB, err := documents.NewDocumentDB(config.DocumentDB)
		if err != nil {
			return fmt.Errorf("failed to create document database: %w", err)
		}

		if err := docDB.Connect(config.DocumentDB); err != nil {
			return fmt.Errorf("failed to connect to document database: %w", err)
		}

		di.DocumentDBs["primary"] = docDB
		di.iLog.Info(fmt.Sprintf("Document database initialized: %s", config.DocumentDB.Type))
	}

	// Start health monitoring
	di.startHealthMonitoring()

	return nil
}

// DatabaseConfig represents complete database configuration
type DatabaseConfig struct {
	Primary    *dbconn.DBConfig
	Replicas   []*dbconn.DBConfig
	DocumentDB *documents.DocDBConfig
}

// LoadConfigFromFile loads database configuration from a file
func LoadConfigFromFile(filename string) (*DatabaseConfig, error) {
	// This would load from JSON/YAML file
	// Placeholder for now
	return nil, fmt.Errorf("not implemented")
}

// Example environment variables:
//
// Relational Database (Primary):
// DB_TYPE=mysql
// DB_HOST=localhost
// DB_PORT=3306
// DB_DATABASE=iac
// DB_USERNAME=root
// DB_PASSWORD=secret
// DB_MAX_IDLE_CONNS=5
// DB_MAX_OPEN_CONNS=10
//
// Relational Database (Replicas):
// DB_REPLICA_COUNT=2
// DB_REPLICA_1_HOST=replica1.example.com
// DB_REPLICA_1_PORT=3306
// DB_REPLICA_2_HOST=replica2.example.com
// DB_REPLICA_2_PORT=3306
//
// Document Database:
// DOCDB_TYPE=mongodb
// DOCDB_HOST=localhost
// DOCDB_PORT=27017
// DOCDB_DATABASE=iac
// DOCDB_USERNAME=admin
// DOCDB_PASSWORD=secret
// DOCDB_AUTH_SOURCE=admin

// Global database initializer instance
var GlobalInitializer *DatabaseInitializer

// InitializeGlobalDatabases initializes the global database connections
func InitializeGlobalDatabases() error {
	GlobalInitializer = NewDatabaseInitializer()
	return GlobalInitializer.InitializeFromEnvironment()
}

// ShutdownGlobalDatabases shuts down all global database connections
func ShutdownGlobalDatabases() error {
	if GlobalInitializer != nil {
		return GlobalInitializer.Shutdown()
	}
	return nil
}

// PrintDatabaseInfo prints information about initialized databases
func (di *DatabaseInitializer) PrintDatabaseInfo() {
	di.iLog.Info("=== Database Configuration ===")

	// Relational databases
	poolInfo := di.PoolManager.GetPoolInfo()
	di.iLog.Info(fmt.Sprintf("Relational Databases: %d", len(poolInfo)))
	for _, info := range poolInfo {
		di.iLog.Info(fmt.Sprintf("  - %s (%s): %s:%s@%s (Active: %v)",
			info.Name, info.Type, info.DBType, info.Database, info.Host, info.Active))
	}

	// Document databases
	di.iLog.Info(fmt.Sprintf("Document Databases: %d", len(di.DocumentDBs)))
	for name, db := range di.DocumentDBs {
		di.iLog.Info(fmt.Sprintf("  - %s (%s): Connected=%v",
			name, db.GetType(), db.IsConnected()))
	}

	di.iLog.Info("=============================")
}
