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

package dbconn_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/sijms/go-ora/v2"
)

// TestConfig holds test database configuration
type TestConfig struct {
	Type     string
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
}

// getTestConfigs returns test configurations for all databases
// These are read from environment variables or use defaults
func getTestConfigs() map[string]TestConfig {
	return map[string]TestConfig{
		"mysql": {
			Type:     "mysql",
			Host:     getEnv("TEST_MYSQL_HOST", "localhost"),
			Port:     getEnvInt("TEST_MYSQL_PORT", 3306),
			Database: getEnv("TEST_MYSQL_DATABASE", "iac"),
			Username: getEnv("TEST_MYSQL_USERNAME", "iac_user"),
			Password: getEnv("TEST_MYSQL_PASSWORD", "iac_pass"),
		},
		"postgres": {
			Type:     "postgres",
			Host:     getEnv("TEST_POSTGRES_HOST", "localhost"),
			Port:     getEnvInt("TEST_POSTGRES_PORT", 5432),
			Database: getEnv("TEST_POSTGRES_DATABASE", "iac"),
			Username: getEnv("TEST_POSTGRES_USERNAME", "iac_user"),
			Password: getEnv("TEST_POSTGRES_PASSWORD", "iac_pass"),
			SSLMode:  "disable",
		},
		"mssql": {
			Type:     "mssql",
			Host:     getEnv("TEST_MSSQL_HOST", "localhost"),
			Port:     getEnvInt("TEST_MSSQL_PORT", 1433),
			Database: getEnv("TEST_MSSQL_DATABASE", "iac"),
			Username: getEnv("TEST_MSSQL_USERNAME", "sa"),
			Password: getEnv("TEST_MSSQL_PASSWORD", "MsSql_Pass123!"),
		},
		"oracle": {
			Type:     "oracle",
			Host:     getEnv("TEST_ORACLE_HOST", "localhost"),
			Port:     getEnvInt("TEST_ORACLE_PORT", 1521),
			Database: getEnv("TEST_ORACLE_DATABASE", "iac"),
			Username: getEnv("TEST_ORACLE_USERNAME", "iac_user"),
			Password: getEnv("TEST_ORACLE_PASSWORD", "iac_pass"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		fmt.Sscanf(value, "%d", &intValue)
		return intValue
	}
	return defaultValue
}

// TestDatabaseConnection tests basic connection for all database types
func TestDatabaseConnection(t *testing.T) {
	configs := getTestConfigs()

	for dbType, config := range configs {
		t.Run(dbType, func(t *testing.T) {
			// Skip if database is not available (detected by env var)
			if os.Getenv(fmt.Sprintf("SKIP_%s_TESTS", dbType)) == "true" {
				t.Skipf("Skipping %s tests (SKIP_%s_TESTS=true)", dbType, dbType)
			}

			dbConfig := dbconn.DBConfig{
				Type:         config.Type,
				Host:         config.Host,
				Port:         config.Port,
				Database:     config.Database,
				Username:     config.Username,
				Password:     config.Password,
				MaxIdleConns: 5,
				MaxOpenConns: 10,
				ConnTimeout:  30,
				Options:      make(map[string]string),
			}

			if config.SSLMode != "" {
				dbConfig.Options["sslmode"] = config.SSLMode
			}

			// Create database instance
			db, err := dbconn.NewRelationalDB(dbConfig)
			if err != nil {
				t.Fatalf("Failed to create %s database: %v", dbType, err)
			}
			defer db.Close()

			// Test connection
			if err := db.Connect(dbConfig); err != nil {
				t.Fatalf("Failed to connect to %s: %v", dbType, err)
			}

			// Test ping
			if err := db.Ping(); err != nil {
				t.Fatalf("Failed to ping %s: %v", dbType, err)
			}

			// Verify dialect
			dialect := db.GetDialect()
			if dialect != dbType {
				t.Errorf("Expected dialect %s, got %s", dbType, dialect)
			}

			t.Logf("✓ %s connection successful (dialect: %s)", dbType, dialect)
		})
	}
}

// TestDatabaseBasicOperations tests basic CRUD operations
func TestDatabaseBasicOperations(t *testing.T) {
	configs := getTestConfigs()

	for dbType, config := range configs {
		t.Run(dbType, func(t *testing.T) {
			if os.Getenv(fmt.Sprintf("SKIP_%s_TESTS", dbType)) == "true" {
				t.Skipf("Skipping %s tests", dbType)
			}

			db := setupDatabase(t, config)
			defer db.Close()

			// Create test table
			tableName := fmt.Sprintf("test_%s_%d", dbType, time.Now().Unix())
			createTableSQL := getCreateTableSQL(dbType, tableName)

			_, err := db.Exec(createTableSQL)
			if err != nil {
				t.Fatalf("Failed to create table: %v", err)
			}
			defer cleanupTable(db, tableName)

			// Test INSERT
			insertSQL := getInsertSQL(dbType, tableName)
			result, err := db.Exec(insertSQL, "John Doe", "john@example.com")
			if err != nil {
				t.Fatalf("Failed to insert record: %v", err)
			}

			// Check affected rows (if supported)
			if db.SupportsFeature("last_insert_id") {
				_, err := result.LastInsertId()
				if err != nil {
					t.Logf("Note: LastInsertId not fully supported: %v", err)
				}
			}

			// Test SELECT
			selectSQL := getSelectSQL(dbType, tableName)
			rows, err := db.Query(selectSQL, "John Doe")
			if err != nil {
				t.Fatalf("Failed to query records: %v", err)
			}
			defer rows.Close()

			count := 0
			for rows.Next() {
				var id int
				var name, email string
				if err := rows.Scan(&id, &name, &email); err != nil {
					t.Fatalf("Failed to scan row: %v", err)
				}
				count++
				if name != "John Doe" || email != "john@example.com" {
					t.Errorf("Unexpected values: name=%s, email=%s", name, email)
				}
			}

			if count != 1 {
				t.Errorf("Expected 1 record, got %d", count)
			}

			// Test UPDATE
			updateSQL := getUpdateSQL(dbType, tableName)
			result, err = db.Exec(updateSQL, "jane@example.com", "John Doe")
			if err != nil {
				t.Fatalf("Failed to update record: %v", err)
			}

			affected, _ := result.RowsAffected()
			if affected != 1 {
				t.Errorf("Expected 1 affected row, got %d", affected)
			}

			// Test DELETE
			deleteSQL := getDeleteSQL(dbType, tableName)
			result, err = db.Exec(deleteSQL, "Jane")
			if err != nil {
				t.Fatalf("Failed to delete record: %v", err)
			}

			t.Logf("✓ %s basic operations successful", dbType)
		})
	}
}

// TestDatabaseTransactions tests transaction support
func TestDatabaseTransactions(t *testing.T) {
	configs := getTestConfigs()

	for dbType, config := range configs {
		t.Run(dbType, func(t *testing.T) {
			if os.Getenv(fmt.Sprintf("SKIP_%s_TESTS", dbType)) == "true" {
				t.Skipf("Skipping %s tests", dbType)
			}

			db := setupDatabase(t, config)
			defer db.Close()

			if !db.SupportsFeature("transactions") {
				t.Skip("Database does not support transactions")
			}

			tableName := fmt.Sprintf("test_tx_%s_%d", dbType, time.Now().Unix())
			createTableSQL := getCreateTableSQL(dbType, tableName)
			_, err := db.Exec(createTableSQL)
			if err != nil {
				t.Fatalf("Failed to create table: %v", err)
			}
			defer cleanupTable(db, tableName)

			// Test COMMIT
			t.Run("Commit", func(t *testing.T) {
				tx, err := db.BeginTx(context.Background(), nil)
				if err != nil {
					t.Fatalf("Failed to begin transaction: %v", err)
				}

				insertSQL := getInsertSQL(dbType, tableName)
				_, err = tx.Exec(insertSQL, "User1", "user1@example.com")
				if err != nil {
					tx.Rollback()
					t.Fatalf("Failed to insert in transaction: %v", err)
				}

				if err := tx.Commit(); err != nil {
					t.Fatalf("Failed to commit transaction: %v", err)
				}

				// Verify data is committed
				rows, err := db.Query(getSelectSQL(dbType, tableName), "User1")
				if err != nil {
					t.Fatalf("Failed to query: %v", err)
				}
				defer rows.Close()

				if !rows.Next() {
					t.Error("Expected record not found after commit")
				}
			})

			// Test ROLLBACK
			t.Run("Rollback", func(t *testing.T) {
				tx, err := db.BeginTx(context.Background(), nil)
				if err != nil {
					t.Fatalf("Failed to begin transaction: %v", err)
				}

				insertSQL := getInsertSQL(dbType, tableName)
				_, err = tx.Exec(insertSQL, "User2", "user2@example.com")
				if err != nil {
					tx.Rollback()
					t.Fatalf("Failed to insert in transaction: %v", err)
				}

				if err := tx.Rollback(); err != nil {
					t.Fatalf("Failed to rollback transaction: %v", err)
				}

				// Verify data is not present
				rows, err := db.Query(getSelectSQL(dbType, tableName), "User2")
				if err != nil {
					t.Fatalf("Failed to query: %v", err)
				}
				defer rows.Close()

				if rows.Next() {
					t.Error("Unexpected record found after rollback")
				}
			})

			t.Logf("✓ %s transaction support verified", dbType)
		})
	}
}

// TestDatabaseFeatureDetection tests feature detection
func TestDatabaseFeatureDetection(t *testing.T) {
	configs := getTestConfigs()

	expectedFeatures := map[string][]string{
		"mysql":    {"transactions", "last_insert_id", "auto_increment"},
		"postgres": {"transactions", "jsonb", "cte", "window_functions", "arrays"},
		"mssql":    {"transactions", "cte", "window_functions"},
		"oracle":   {"transactions", "cte", "window_functions"},
	}

	for dbType, config := range configs {
		t.Run(dbType, func(t *testing.T) {
			if os.Getenv(fmt.Sprintf("SKIP_%s_TESTS", dbType)) == "true" {
				t.Skipf("Skipping %s tests", dbType)
			}

			db := setupDatabase(t, config)
			defer db.Close()

			features := expectedFeatures[dbType]
			for _, feature := range features {
				if db.SupportsFeature(feature) {
					t.Logf("✓ %s supports %s", dbType, feature)
				} else {
					t.Logf("✗ %s does not support %s", dbType, feature)
				}
			}
		})
	}
}

// Helper functions

func setupDatabase(t *testing.T, config TestConfig) dbconn.RelationalDB {
	dbConfig := dbconn.DBConfig{
		Type:         config.Type,
		Host:         config.Host,
		Port:         config.Port,
		Database:     config.Database,
		Username:     config.Username,
		Password:     config.Password,
		MaxIdleConns: 5,
		MaxOpenConns: 10,
		ConnTimeout:  30,
		Options:      make(map[string]string),
	}

	if config.SSLMode != "" {
		dbConfig.Options["sslmode"] = config.SSLMode
	}

	db, err := dbconn.NewRelationalDB(dbConfig)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := db.Connect(dbConfig); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	return db
}

func cleanupTable(db dbconn.RelationalDB, tableName string) {
	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	db.Exec(dropSQL)
}

func getCreateTableSQL(dbType, tableName string) string {
	switch dbType {
	case "mysql":
		return fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id INT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(100),
				email VARCHAR(100)
			)`, tableName)
	case "postgres":
		return fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id SERIAL PRIMARY KEY,
				name VARCHAR(100),
				email VARCHAR(100)
			)`, tableName)
	case "mssql":
		return fmt.Sprintf(`
			CREATE TABLE %s (
				id INT IDENTITY(1,1) PRIMARY KEY,
				name NVARCHAR(100),
				email NVARCHAR(100)
			)`, tableName)
	case "oracle":
		return fmt.Sprintf(`
			CREATE TABLE %s (
				id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
				name VARCHAR2(100),
				email VARCHAR2(100)
			)`, tableName)
	default:
		return fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id INT PRIMARY KEY,
				name VARCHAR(100),
				email VARCHAR(100)
			)`, tableName)
	}
}

func getInsertSQL(dbType, tableName string) string {
	return fmt.Sprintf("INSERT INTO %s (name, email) VALUES (?, ?)", tableName)
}

func getSelectSQL(dbType, tableName string) string {
	return fmt.Sprintf("SELECT id, name, email FROM %s WHERE name = ?", tableName)
}

func getUpdateSQL(dbType, tableName string) string {
	return fmt.Sprintf("UPDATE %s SET email = ? WHERE name = ?", tableName)
}

func getDeleteSQL(dbType, tableName string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE name LIKE ?", tableName)
}
