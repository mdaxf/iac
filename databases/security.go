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
	"os"
	"regexp"
	"strings"
)

// ValidateIdentifier validates database identifiers (schema, table, column names)
// to prevent SQL injection via identifier manipulation
func ValidateIdentifier(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("identifier cannot be empty")
	}

	// Maximum length (MySQL limit is 64, PostgreSQL is 63)
	if len(identifier) > 63 {
		return fmt.Errorf("identifier too long: maximum 63 characters")
	}

	// Must start with letter or underscore, contain only alphanumeric and underscores
	// This prevents SQL keywords and special characters
	pattern := `^[a-zA-Z_][a-zA-Z0-9_]*$`
	matched, err := regexp.MatchString(pattern, identifier)
	if err != nil {
		return fmt.Errorf("failed to validate identifier: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid identifier: must start with letter or underscore and contain only alphanumeric characters and underscores")
	}

	// Check for SQL keywords that could be dangerous even if quoted
	dangerousKeywords := []string{
		"DROP", "DELETE", "TRUNCATE", "ALTER", "CREATE",
		"GRANT", "REVOKE", "EXEC", "EXECUTE", "SCRIPT",
	}

	upperIdentifier := strings.ToUpper(identifier)
	for _, keyword := range dangerousKeywords {
		if upperIdentifier == keyword {
			return fmt.Errorf("identifier cannot be SQL keyword: %s", keyword)
		}
	}

	return nil
}

// ValidateSSLConfig validates SSL configuration and enforces SSL in production
func ValidateSSLConfig(config DBConfig) error {
	env := strings.ToLower(os.Getenv("ENV"))
	isProduction := env == "production" || env == "prod"

	sslMode, hasSslMode := config.Options["sslmode"]

	// For PostgreSQL, enforce SSL in production
	if config.Type == "postgres" && isProduction {
		if !hasSslMode || sslMode == "disable" {
			return fmt.Errorf("SSL must be enabled in production for PostgreSQL (current: %s)", sslMode)
		}

		// Recommend verify-full for production
		if sslMode != "verify-full" && sslMode != "verify-ca" {
			fmt.Fprintf(os.Stderr, "WARNING: SSL mode '%s' does not verify certificates. Consider 'verify-full' for production.\n", sslMode)
		}
	}

	// For MySQL, check SSL configuration
	if config.Type == "mysql" && isProduction {
		sslEnabled, hasSSL := config.Options["tls"]
		if !hasSSL || sslEnabled == "false" || sslEnabled == "skip-verify" {
			fmt.Fprintf(os.Stderr, "WARNING: SSL not properly configured for MySQL in production.\n")
		}
	}

	return nil
}

// ValidateConnectionConfig validates database connection configuration for security
func ValidateConnectionConfig(config DBConfig) error {
	// Validate identifier fields
	if config.Database != "" {
		if err := ValidateIdentifier(config.Database); err != nil {
			return fmt.Errorf("invalid database name: %w", err)
		}
	}

	// Validate username (less restrictive than identifiers)
	if config.Username == "" {
		return fmt.Errorf("username is required")
	}

	if len(config.Username) > 128 {
		return fmt.Errorf("username too long: maximum 128 characters")
	}

	// Check for obviously weak passwords (basic check)
	if len(config.Password) < 8 {
		fmt.Fprintf(os.Stderr, "WARNING: Password is very short. Minimum 12 characters recommended.\n")
	}

	// Validate timeout
	if config.ConnTimeout <= 0 {
		return fmt.Errorf("connection timeout must be positive")
	}

	if config.ConnTimeout > 300 {
		fmt.Fprintf(os.Stderr, "WARNING: Connection timeout is very high (%d seconds).\n", config.ConnTimeout)
	}

	// Validate connection pool settings
	if config.MaxOpenConns < config.MaxIdleConns {
		return fmt.Errorf("max open connections (%d) must be >= max idle connections (%d)",
			config.MaxOpenConns, config.MaxIdleConns)
	}

	if config.MaxOpenConns > 1000 {
		fmt.Fprintf(os.Stderr, "WARNING: Very high max open connections (%d). This may cause resource exhaustion.\n",
			config.MaxOpenConns)
	}

	// Validate SSL configuration
	if err := ValidateSSLConfig(config); err != nil {
		return err
	}

	return nil
}

// SanitizeErrorMessage removes sensitive information from error messages
func SanitizeErrorMessage(err error, config DBConfig) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	// Remove password from error message
	if config.Password != "" {
		errMsg = strings.ReplaceAll(errMsg, config.Password, "***")
	}

	// Remove full connection strings
	patterns := []string{
		config.Username + ":" + config.Password,
		config.Password + "@",
	}

	for _, pattern := range patterns {
		if pattern != "" && pattern != ":" && pattern != "@" {
			errMsg = strings.ReplaceAll(errMsg, pattern, "***")
		}
	}

	// Remove IP addresses in production (keep for debugging in dev)
	if os.Getenv("ENV") == "production" {
		ipPattern := regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`)
		errMsg = ipPattern.ReplaceAllString(errMsg, "xxx.xxx.xxx.xxx")
	}

	return errMsg
}

// AuditLog represents a database access audit log entry
type AuditLog struct {
	Timestamp   string `json:"timestamp"`
	User        string `json:"user"`
	Database    string `json:"database"`
	Operation   string `json:"operation"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	IPAddress   string `json:"ip_address,omitempty"`
	Duration    int64  `json:"duration_ms,omitempty"`
}

// LogDatabaseAccess logs database access for audit purposes
// This is a placeholder for actual audit logging implementation
func LogDatabaseAccess(log AuditLog) {
	// TODO: Implement actual audit logging
	// For now, just log to stderr in non-production
	if os.Getenv("ENV") != "production" {
		fmt.Fprintf(os.Stderr, "AUDIT: %s - User: %s, DB: %s, Op: %s, Success: %v\n",
			log.Timestamp, log.User, log.Database, log.Operation, log.Success)
	}
}

// RateLimiter provides basic rate limiting for database operations
type RateLimiter struct {
	maxQueriesPerSecond int
	// TODO: Implement actual rate limiting logic
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxQueriesPerSecond int) *RateLimiter {
	return &RateLimiter{
		maxQueriesPerSecond: maxQueriesPerSecond,
	}
}

// AllowQuery checks if a query is allowed based on rate limits
func (rl *RateLimiter) AllowQuery() bool {
	// TODO: Implement actual rate limiting
	// For now, always allow
	return true
}

// SecurityConfig holds security configuration options
type SecurityConfig struct {
	EnforceSSL          bool   // Enforce SSL/TLS connections
	ValidateIdentifiers bool   // Validate schema/table/column names
	LogQueries          bool   // Log all queries (be careful with sensitive data)
	LogParameters       bool   // Log query parameters (disable in production)
	MaxQueryDuration    int    // Maximum query duration in seconds
	EnableAuditLog      bool   // Enable audit logging
	RateLimitQPS        int    // Queries per second limit (0 = unlimited)
	Environment         string // Environment (production, staging, development)
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	isProduction := env == "production" || env == "prod"

	return &SecurityConfig{
		EnforceSSL:          isProduction,
		ValidateIdentifiers: true,
		LogQueries:          !isProduction,
		LogParameters:       false, // Never log parameters by default
		MaxQueryDuration:    30,
		EnableAuditLog:      isProduction,
		RateLimitQPS:        0, // Unlimited by default
		Environment:         env,
	}
}

// ValidateSecurityConfig validates the security configuration
func ValidateSecurityConfig(config *SecurityConfig) error {
	if config == nil {
		return fmt.Errorf("security config cannot be nil")
	}

	if config.MaxQueryDuration < 1 || config.MaxQueryDuration > 3600 {
		return fmt.Errorf("max query duration must be between 1 and 3600 seconds")
	}

	if config.RateLimitQPS < 0 {
		return fmt.Errorf("rate limit QPS cannot be negative")
	}

	// Warn about risky configurations
	if config.LogParameters && config.Environment == "production" {
		fmt.Fprintf(os.Stderr, "WARNING: Query parameter logging is enabled in production. This may log sensitive data.\n")
	}

	if !config.EnforceSSL && config.Environment == "production" {
		fmt.Fprintf(os.Stderr, "WARNING: SSL enforcement is disabled in production. This is a security risk.\n")
	}

	return nil
}
