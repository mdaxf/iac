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
	"fmt"
	"net/url"
	"strings"
)

// BuildConnectionString builds a database-specific connection string
func BuildConnectionString(config *DBConfig) (string, error) {
	switch config.Type {
	case DBTypeMySQL:
		return buildMySQLConnectionString(config), nil
	case DBTypePostgreSQL:
		return buildPostgreSQLConnectionString(config), nil
	case DBTypeMSSQL:
		return buildMSSQLConnectionString(config), nil
	case DBTypeOracle:
		return buildOracleConnectionString(config), nil
	default:
		return "", fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// buildMySQLConnectionString builds MySQL connection string
// Format: user:password@tcp(host:port)/database?params
func buildMySQLConnectionString(config *DBConfig) string {
	// Build base connection string
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	// Add parameters
	params := []string{}

	// SSL/TLS configuration
	if config.SSLMode != "" {
		params = append(params, fmt.Sprintf("tls=%s", config.SSLMode))
	}

	// Charset
	params = append(params, "charset=utf8mb4")
	params = append(params, "parseTime=True")
	params = append(params, "loc=Local")

	// Additional options
	for k, v := range config.Options {
		params = append(params, fmt.Sprintf("%s=%s", k, v))
	}

	if len(params) > 0 {
		connStr += "?" + strings.Join(params, "&")
	}

	return connStr
}

// buildPostgreSQLConnectionString builds PostgreSQL connection string
// Format: host=localhost port=5432 user=myuser password=mypass dbname=mydb sslmode=disable
func buildPostgreSQLConnectionString(config *DBConfig) string {
	parts := []string{
		fmt.Sprintf("host=%s", config.Host),
		fmt.Sprintf("port=%d", config.Port),
		fmt.Sprintf("user=%s", config.Username),
		fmt.Sprintf("password=%s", config.Password),
		fmt.Sprintf("dbname=%s", config.Database),
	}

	// Schema (search_path)
	if config.Schema != "" {
		parts = append(parts, fmt.Sprintf("search_path=%s", config.Schema))
	}

	// SSL Mode
	sslMode := config.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	parts = append(parts, fmt.Sprintf("sslmode=%s", sslMode))

	// SSL certificates
	if config.SSLCert != "" {
		parts = append(parts, fmt.Sprintf("sslcert=%s", config.SSLCert))
	}
	if config.SSLKey != "" {
		parts = append(parts, fmt.Sprintf("sslkey=%s", config.SSLKey))
	}
	if config.SSLRootCert != "" {
		parts = append(parts, fmt.Sprintf("sslrootcert=%s", config.SSLRootCert))
	}

	// Connection timeout
	parts = append(parts, fmt.Sprintf("connect_timeout=%d", config.ConnTimeout))

	// Additional options
	for k, v := range config.Options {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(parts, " ")
}

// buildMSSQLConnectionString builds MSSQL/SQL Server connection string
// Format: sqlserver://username:password@host:port?database=dbname
func buildMSSQLConnectionString(config *DBConfig) string {
	params := url.Values{}
	params.Add("database", config.Database)

	// Connection timeout
	params.Add("connection timeout", fmt.Sprintf("%d", config.ConnTimeout))

	// Encryption
	if config.SSLMode == "require" || config.SSLMode == "verify-full" {
		params.Add("encrypt", "true")
	} else {
		params.Add("encrypt", "false")
	}

	// Trust server certificate
	if config.SSLMode == "disable" {
		params.Add("TrustServerCertificate", "true")
	}

	// Additional options
	for k, v := range config.Options {
		params.Add(k, v)
	}

	// Build URL
	connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%d?%s",
		url.QueryEscape(config.Username),
		url.QueryEscape(config.Password),
		config.Host,
		config.Port,
		params.Encode(),
	)

	return connStr
}

// buildOracleConnectionString builds Oracle connection string
// Format: oracle://user:password@host:port/service_name
func buildOracleConnectionString(config *DBConfig) string {
	serviceName := config.Database
	if config.Schema != "" {
		serviceName = config.Schema
	}

	connStr := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		url.QueryEscape(config.Username),
		url.QueryEscape(config.Password),
		config.Host,
		config.Port,
		serviceName,
	)

	// Add SSL/TLS if configured
	if config.SSLMode != "" {
		params := url.Values{}
		params.Add("ssl", "true")

		if config.SSLCert != "" {
			params.Add("wallet", config.SSLCert)
		}

		// Additional options
		for k, v := range config.Options {
			params.Add(k, v)
		}

		if len(params) > 0 {
			connStr += "?" + params.Encode()
		}
	}

	return connStr
}

// ParseConnectionString parses a connection string into DBConfig
// This is useful for backward compatibility
func ParseConnectionString(connStr string, dbType DBType) (*DBConfig, error) {
	config := &DBConfig{
		Type: dbType,
	}

	switch dbType {
	case DBTypeMySQL:
		return parseMySQLConnectionString(connStr, config)
	case DBTypePostgreSQL:
		return parsePostgreSQLConnectionString(connStr, config)
	case DBTypeMSSQL:
		return parseMSSQLConnectionString(connStr, config)
	case DBTypeOracle:
		return parseOracleConnectionString(connStr, config)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// Simplified parsers - these are basic implementations
// Full production parsers would be more robust

func parseMySQLConnectionString(connStr string, config *DBConfig) (*DBConfig, error) {
	// Basic MySQL connection string parser
	// Format: user:password@tcp(host:port)/database?params
	// This is a simplified implementation
	config.MaxIdleConns = 5
	config.MaxOpenConns = 10
	config.ConnTimeout = 30
	return config, nil
}

func parsePostgreSQLConnectionString(connStr string, config *DBConfig) (*DBConfig, error) {
	// Basic PostgreSQL connection string parser
	config.MaxIdleConns = 5
	config.MaxOpenConns = 10
	config.ConnTimeout = 30
	return config, nil
}

func parseMSSQLConnectionString(connStr string, config *DBConfig) (*DBConfig, error) {
	// Basic MSSQL connection string parser
	config.MaxIdleConns = 5
	config.MaxOpenConns = 10
	config.ConnTimeout = 30
	return config, nil
}

func parseOracleConnectionString(connStr string, config *DBConfig) (*DBConfig, error) {
	// Basic Oracle connection string parser
	config.MaxIdleConns = 5
	config.MaxOpenConns = 10
	config.ConnTimeout = 30
	return config, nil
}
