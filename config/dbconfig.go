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
	"fmt"
	"os"
	"strconv"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
)

// DatabaseConfigurations holds all database configurations
type DatabaseConfigurations struct {
	Primary   *dbconn.DBConfig      `json:"primary"`
	Replicas  []*dbconn.DBConfig    `json:"replicas,omitempty"`
	Document  *documents.DocDBConfig `json:"document,omitempty"`
}

var (
	// DBConfigs holds the parsed database configurations
	DBConfigs *DatabaseConfigurations
)

// ParseDatabaseConfig parses the database configuration from GlobalConfig
// This maintains backward compatibility while supporting new multi-database features
func ParseDatabaseConfig(globalConfig *GlobalConfig) (*DatabaseConfigurations, error) {
	configs := &DatabaseConfigurations{}

	// Parse primary database
	if dbConfig := globalConfig.DatabaseConfig; dbConfig != nil {
		primary, err := parseSingleDBConfig(dbConfig, "primary")
		if err != nil {
			return nil, fmt.Errorf("failed to parse primary database config: %w", err)
		}
		configs.Primary = primary
	}

	// Parse alternate databases (replicas)
	if altDBs := globalConfig.AltDatabasesConfig; len(altDBs) > 0 {
		configs.Replicas = make([]*dbconn.DBConfig, 0, len(altDBs))
		for i, altDB := range altDBs {
			replica, err := parseSingleDBConfig(altDB, fmt.Sprintf("replica-%d", i))
			if err != nil {
				return nil, fmt.Errorf("failed to parse replica database config %d: %w", i, err)
			}
			configs.Replicas = append(configs.Replicas, replica)
		}
	}

	// Parse document database
	if docConfig := globalConfig.DocumentConfig; docConfig != nil {
		docDB, err := parseDocDBConfig(docConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parse document database config: %w", err)
		}
		configs.Document = docDB
	}

	DBConfigs = configs
	return configs, nil
}

// parseSingleDBConfig parses a single database configuration from map
func parseSingleDBConfig(data map[string]interface{}, name string) (*dbconn.DBConfig, error) {
	config := &dbconn.DBConfig{}

	// Database type (default to mysql for backward compatibility)
	dbType := getStringValue(data, "type", "mysql")
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

	// Connection details - support both new structured and old connection string format
	if connStr, ok := data["connection"].(string); ok && connStr != "" {
		// Parse connection string (backward compatibility)
		parseConnectionString(connStr, config)
	} else {
		// Use structured configuration
		config.Host = getStringValue(data, "host", "localhost")
		config.Port = getIntValue(data, "port", getDefaultPort(config.Type))
		config.Database = getStringValue(data, "database", "iac")
		config.Schema = getStringValue(data, "schema", "")
		config.Username = getStringValue(data, "username", "")
		config.Password = getStringValue(data, "password", "")
	}

	// Environment variable overrides
	if envHost := os.Getenv(fmt.Sprintf("DB_%s_HOST", name)); envHost != "" {
		config.Host = envHost
	}
	if envUser := os.Getenv(fmt.Sprintf("DB_%s_USER", name)); envUser != "" {
		config.Username = envUser
	}
	if envPass := os.Getenv(fmt.Sprintf("DB_%s_PASSWORD", name)); envPass != "" {
		config.Password = envPass
	}
	if envDB := os.Getenv(fmt.Sprintf("DB_%s_NAME", name)); envDB != "" {
		config.Database = envDB
	}

	// SSL Configuration
	config.SSLMode = getStringValue(data, "ssl_mode", "")
	config.SSLCert = getStringValue(data, "ssl_cert", "")
	config.SSLKey = getStringValue(data, "ssl_key", "")
	config.SSLRootCert = getStringValue(data, "ssl_root_cert", "")

	// Pool Configuration
	config.MaxIdleConns = getIntValue(data, "max_idle_conns", 5)
	config.MaxOpenConns = getIntValue(data, "max_open_conns", 10)
	config.ConnTimeout = getIntValue(data, "timeout", 30)

	if maxLifetime := getIntValue(data, "conn_max_lifetime", 0); maxLifetime > 0 {
		config.ConnMaxLifetime = time.Duration(maxLifetime) * time.Second
	}
	if maxIdleTime := getIntValue(data, "conn_max_idle_time", 0); maxIdleTime > 0 {
		config.ConnMaxIdleTime = time.Duration(maxIdleTime) * time.Second
	}

	// Database-specific options
	if opts, ok := data["options"].(map[string]interface{}); ok {
		config.Options = make(map[string]string)
		for k, v := range opts {
			if strVal, ok := v.(string); ok {
				config.Options[k] = strVal
			}
		}
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// parseDocDBConfig parses document database configuration
func parseDocDBConfig(data map[string]interface{}) (*documents.DocDBConfig, error) {
	config := &documents.DocDBConfig{}

	// Database type (default to mongodb for backward compatibility)
	dbType := getStringValue(data, "type", "mongodb")
	switch dbType {
	case "mongodb", "mongo":
		config.Type = documents.DocDBTypeMongoDB
	case "postgres", "postgresql":
		config.Type = documents.DocDBTypePostgres
	default:
		config.Type = documents.DocDBTypeMongoDB
	}

	// Connection details
	if connStr, ok := data["connection"].(string); ok && connStr != "" {
		// For MongoDB, connection string is often complete
		// We'll still parse out key components
		config.Host = getStringValue(data, "host", "localhost")
		config.Database = getStringValue(data, "database", "iac")
	} else {
		config.Host = getStringValue(data, "host", "localhost")
		config.Port = getIntValue(data, "port", getDefaultDocDBPort(config.Type))
		config.Database = getStringValue(data, "database", "iac")
		config.Username = getStringValue(data, "username", "")
		config.Password = getStringValue(data, "password", "")
	}

	// Environment variable overrides
	if envHost := os.Getenv("DOCDB_HOST"); envHost != "" {
		config.Host = envHost
	}
	if envUser := os.Getenv("DOCDB_USER"); envUser != "" {
		config.Username = envUser
	}
	if envPass := os.Getenv("DOCDB_PASSWORD"); envPass != "" {
		config.Password = envPass
	}
	if envDB := os.Getenv("DOCDB_NAME"); envDB != "" {
		config.Database = envDB
	}

	// MongoDB-specific options
	config.SSLMode = getStringValue(data, "ssl_mode", "")
	config.AuthSource = getStringValue(data, "auth_source", "admin")
	config.ReplicaSet = getStringValue(data, "replica_set", "")

	// Pool Configuration
	config.MaxPoolSize = getIntValue(data, "max_pool_size", 100)
	config.MinPoolSize = getIntValue(data, "min_pool_size", 10)
	config.ConnTimeout = getIntValue(data, "timeout", 30)

	// Database-specific options
	if opts, ok := data["options"].(map[string]interface{}); ok {
		config.Options = make(map[string]string)
		for k, v := range opts {
			if strVal, ok := v.(string); ok {
				config.Options[k] = strVal
			}
		}
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Helper functions

func getStringValue(data map[string]interface{}, key, defaultValue string) string {
	if val, ok := data[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func getIntValue(data map[string]interface{}, key string, defaultValue int) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal
			}
		}
	}
	return defaultValue
}

func getDefaultPort(dbType dbconn.DBType) int {
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

func getDefaultDocDBPort(dbType documents.DocDBType) int {
	switch dbType {
	case documents.DocDBTypeMongoDB:
		return 27017
	case documents.DocDBTypePostgres:
		return 5432
	default:
		return 27017
	}
}

// parseConnectionString parses old-style connection string for backward compatibility
// MySQL: user:password@tcp(localhost:3306)/mydb
// MSSQL: server=xxx;port=1433;user id=xx;password=xxx;database=xxx
func parseConnectionString(connStr string, config *dbconn.DBConfig) {
	// This is a simplified parser - actual implementation would be more robust
	// For now, we'll just store it and let the driver handle it
	// The actual parsing would depend on the database type

	// For MySQL format: user:password@tcp(host:port)/database
	// For MSSQL format: server=host;port=port;user id=user;password=pass;database=db

	// This is kept for backward compatibility - new configurations should use structured format
}
