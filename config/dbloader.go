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

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
)

// LoadDatabaseConfigurations loads all database configurations from various sources
// Priority: Environment Variables > Config File > Defaults
func LoadDatabaseConfigurations() (*DatabaseConfigurations, error) {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DatabaseLoader"}

	var configs *DatabaseConfigurations
	var err error

	// Try to load from dedicated database config file first
	configs, err = loadFromDatabaseConfigFile()
	if err != nil {
		iLog.Debug(fmt.Sprintf("No separate database config file found, using global config: %v", err))

		// Fall back to global configuration
		if GlobalConfiguration != nil {
			configs, err = ParseDatabaseConfig(GlobalConfiguration)
			if err != nil {
				return nil, fmt.Errorf("failed to parse database config from global config: %w", err)
			}
		} else {
			// No configuration file found, use environment variables only
			configs = &DatabaseConfigurations{}
		}
	}

	// Override with environment variables
	if err := loadFromEnvironment(configs); err != nil {
		iLog.Warn(fmt.Sprintf("Failed to load some environment variables: %v", err))
	}

	// Validate all configurations
	if err := validateConfigurations(configs); err != nil {
		return nil, fmt.Errorf("database configuration validation failed: %w", err)
	}

	iLog.Info(fmt.Sprintf("Database configurations loaded successfully: Primary=%s, Replicas=%d, Document=%v",
		configs.Primary.Type, len(configs.Replicas), configs.Document != nil))

	DBConfigs = configs
	return configs, nil
}

// loadFromDatabaseConfigFile loads configuration from databases.json
func loadFromDatabaseConfigFile() (*DatabaseConfigurations, error) {
	configFile := "databases.json"

	// Check if file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("database config file not found: %s", configFile)
	}

	// Read file
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read database config file: %w", err)
	}

	// Parse JSON
	var configs DatabaseConfigurations
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse database config file: %w", err)
	}

	return &configs, nil
}

// loadFromEnvironment loads database configuration from environment variables
func loadFromEnvironment(configs *DatabaseConfigurations) error {
	// Load primary database from environment
	if primaryConfig := loadPrimaryDBFromEnv(); primaryConfig != nil {
		if configs.Primary == nil {
			configs.Primary = primaryConfig
		} else {
			// Merge with existing config (env vars take precedence)
			mergePrimaryConfig(configs.Primary, primaryConfig)
		}
	}

	// Load document database from environment
	if docConfig := loadDocDBFromEnv(); docConfig != nil {
		if configs.Document == nil {
			configs.Document = docConfig
		} else {
			// Merge with existing config (env vars take precedence)
			mergeDocConfig(configs.Document, docConfig)
		}
	}

	return nil
}

// loadPrimaryDBFromEnv loads primary database configuration from environment variables
func loadPrimaryDBFromEnv() *dbconn.DBConfig {
	// Check if any primary DB env vars are set
	if os.Getenv("DB_TYPE") == "" && os.Getenv("DB_HOST") == "" {
		return nil
	}

	config := &dbconn.DBConfig{}

	// Database type
	dbType := getEnv("DB_TYPE", "mysql")
	switch dbType {
	case "mysql":
		config.Type = dbconn.DBTypeMySQL
	case "postgres", "postgresql":
		config.Type = dbconn.DBTypePostgreSQL
	case "mssql", "sqlserver":
		config.Type = dbconn.DBTypeMSSQL
	case "oracle":
		config.Type = dbconn.DBTypeOracle
	default:
		config.Type = dbconn.DBTypeMySQL
	}

	// Connection details
	config.Host = getEnv("DB_HOST", "localhost")
	config.Port = getEnvInt("DB_PORT", getDefaultPort(config.Type))
	config.Database = getEnv("DB_NAME", "iac")
	config.Schema = getEnv("DB_SCHEMA", "")
	config.Username = getEnv("DB_USER", "")
	config.Password = getEnv("DB_PASSWORD", "")

	// SSL Configuration
	config.SSLMode = getEnv("DB_SSL_MODE", "")
	config.SSLCert = getEnv("DB_SSL_CERT", "")
	config.SSLKey = getEnv("DB_SSL_KEY", "")
	config.SSLRootCert = getEnv("DB_SSL_ROOT_CERT", "")

	// Pool Configuration
	config.MaxIdleConns = getEnvInt("DB_MAX_IDLE_CONNS", 5)
	config.MaxOpenConns = getEnvInt("DB_MAX_OPEN_CONNS", 10)
	config.ConnTimeout = getEnvInt("DB_TIMEOUT", 30)

	return config
}

// loadDocDBFromEnv loads document database configuration from environment variables
func loadDocDBFromEnv() *documents.DocDBConfig {
	// Check if any document DB env vars are set
	if os.Getenv("DOCDB_TYPE") == "" && os.Getenv("DOCDB_HOST") == "" {
		return nil
	}

	config := &documents.DocDBConfig{}

	// Database type
	dbType := getEnv("DOCDB_TYPE", "mongodb")
	switch dbType {
	case "mongodb", "mongo":
		config.Type = documents.DocDBTypeMongoDB
	case "postgres", "postgresql":
		config.Type = documents.DocDBTypePostgres
	default:
		config.Type = documents.DocDBTypeMongoDB
	}

	// Connection details
	config.Host = getEnv("DOCDB_HOST", "localhost")
	config.Port = getEnvInt("DOCDB_PORT", getDefaultDocDBPort(config.Type))
	config.Database = getEnv("DOCDB_NAME", "iac")
	config.Username = getEnv("DOCDB_USER", "")
	config.Password = getEnv("DOCDB_PASSWORD", "")

	// MongoDB-specific
	config.AuthSource = getEnv("DOCDB_AUTH_SOURCE", "admin")
	config.ReplicaSet = getEnv("DOCDB_REPLICA_SET", "")
	config.SSLMode = getEnv("DOCDB_SSL_MODE", "")

	// Pool Configuration
	config.MaxPoolSize = getEnvInt("DOCDB_MAX_POOL_SIZE", 100)
	config.MinPoolSize = getEnvInt("DOCDB_MIN_POOL_SIZE", 10)
	config.ConnTimeout = getEnvInt("DOCDB_TIMEOUT", 30)

	return config
}

// mergePrimaryConfig merges environment config into existing config
func mergePrimaryConfig(existing, envConfig *dbconn.DBConfig) {
	if envConfig.Type != "" {
		existing.Type = envConfig.Type
	}
	if envConfig.Host != "" {
		existing.Host = envConfig.Host
	}
	if envConfig.Port > 0 {
		existing.Port = envConfig.Port
	}
	if envConfig.Database != "" {
		existing.Database = envConfig.Database
	}
	if envConfig.Schema != "" {
		existing.Schema = envConfig.Schema
	}
	if envConfig.Username != "" {
		existing.Username = envConfig.Username
	}
	if envConfig.Password != "" {
		existing.Password = envConfig.Password
	}
	if envConfig.SSLMode != "" {
		existing.SSLMode = envConfig.SSLMode
	}
	if envConfig.MaxIdleConns > 0 {
		existing.MaxIdleConns = envConfig.MaxIdleConns
	}
	if envConfig.MaxOpenConns > 0 {
		existing.MaxOpenConns = envConfig.MaxOpenConns
	}
}

// mergeDocConfig merges environment config into existing document DB config
func mergeDocConfig(existing, envConfig *documents.DocDBConfig) {
	if envConfig.Type != "" {
		existing.Type = envConfig.Type
	}
	if envConfig.Host != "" {
		existing.Host = envConfig.Host
	}
	if envConfig.Port > 0 {
		existing.Port = envConfig.Port
	}
	if envConfig.Database != "" {
		existing.Database = envConfig.Database
	}
	if envConfig.Username != "" {
		existing.Username = envConfig.Username
	}
	if envConfig.Password != "" {
		existing.Password = envConfig.Password
	}
	if envConfig.AuthSource != "" {
		existing.AuthSource = envConfig.AuthSource
	}
	if envConfig.ReplicaSet != "" {
		existing.ReplicaSet = envConfig.ReplicaSet
	}
}

// validateConfigurations validates all database configurations
func validateConfigurations(configs *DatabaseConfigurations) error {
	// Validate primary database (required)
	if configs.Primary == nil {
		return fmt.Errorf("primary database configuration is required")
	}
	if err := configs.Primary.Validate(); err != nil {
		return fmt.Errorf("primary database validation failed: %w", err)
	}

	// Validate replica databases (optional)
	for i, replica := range configs.Replicas {
		if err := replica.Validate(); err != nil {
			return fmt.Errorf("replica database %d validation failed: %w", i, err)
		}
	}

	// Validate document database (optional)
	if configs.Document != nil {
		if err := configs.Document.Validate(); err != nil {
			return fmt.Errorf("document database validation failed: %w", err)
		}
	}

	return nil
}

// GetDatabaseConfig returns the current database configurations
func GetDatabaseConfig() *DatabaseConfigurations {
	return DBConfigs
}

// GetPrimaryDBConfig returns the primary database configuration
func GetPrimaryDBConfig() *dbconn.DBConfig {
	if DBConfigs != nil {
		return DBConfigs.Primary
	}
	return nil
}

// GetDocumentDBConfig returns the document database configuration
func GetDocumentDBConfig() *documents.DocDBConfig {
	if DBConfigs != nil {
		return DBConfigs.Document
	}
	return nil
}

// GetReplicaDBConfigs returns all replica database configurations
func GetReplicaDBConfigs() []*dbconn.DBConfig {
	if DBConfigs != nil {
		return DBConfigs.Replicas
	}
	return nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := parseInt(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
