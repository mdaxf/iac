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

package databases_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mdaxf/iac/databases"
)

// BenchmarkDatabaseConnection benchmarks connection establishment
func BenchmarkDatabaseConnection(b *testing.B) {
	configs := getTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipTest(dbType) {
				b.Skip("Database not available")
			}

			dbConfig := databases.DBConfig{
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

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				db, err := databases.NewRelationalDB(dbConfig)
				if err != nil {
					b.Fatalf("Failed to create database: %v", err)
				}

				if err := db.Connect(dbConfig); err != nil {
					b.Fatalf("Failed to connect: %v", err)
				}

				db.Close()
			}
		})
	}
}

// BenchmarkDatabasePing benchmarks ping operations
func BenchmarkDatabasePing(b *testing.B) {
	configs := getTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipTest(dbType) {
				b.Skip("Database not available")
			}

			db := setupDatabase(b, config)
			defer db.Close()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := db.Ping(); err != nil {
					b.Fatalf("Ping failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkDatabaseInsert benchmarks INSERT operations
func BenchmarkDatabaseInsert(b *testing.B) {
	configs := getTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipTest(dbType) {
				b.Skip("Database not available")
			}

			db := setupDatabase(b, config)
			defer db.Close()

			tableName := fmt.Sprintf("bench_insert_%s_%d", dbType, time.Now().Unix())
			createTableSQL := getCreateTableSQL(dbType, tableName)
			_, err := db.Exec(createTableSQL)
			if err != nil {
				b.Fatalf("Failed to create table: %v", err)
			}
			defer cleanupTable(db, tableName)

			insertSQL := getInsertSQL(dbType, tableName)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				name := fmt.Sprintf("User%d", i)
				email := fmt.Sprintf("user%d@example.com", i)
				_, err := db.Exec(insertSQL, name, email)
				if err != nil {
					b.Fatalf("Insert failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkDatabaseSelect benchmarks SELECT operations
func BenchmarkDatabaseSelect(b *testing.B) {
	configs := getTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipTest(dbType) {
				b.Skip("Database not available")
			}

			db := setupDatabase(b, config)
			defer db.Close()

			tableName := fmt.Sprintf("bench_select_%s_%d", dbType, time.Now().Unix())
			createTableSQL := getCreateTableSQL(dbType, tableName)
			_, err := db.Exec(createTableSQL)
			if err != nil {
				b.Fatalf("Failed to create table: %v", err)
			}
			defer cleanupTable(db, tableName)

			// Insert test data
			insertSQL := getInsertSQL(dbType, tableName)
			for i := 0; i < 1000; i++ {
				name := fmt.Sprintf("User%d", i)
				email := fmt.Sprintf("user%d@example.com", i)
				db.Exec(insertSQL, name, email)
			}

			selectSQL := getSelectSQL(dbType, tableName)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				name := fmt.Sprintf("User%d", i%1000)
				rows, err := db.Query(selectSQL, name)
				if err != nil {
					b.Fatalf("Select failed: %v", err)
				}
				for rows.Next() {
					var id int
					var n, e string
					rows.Scan(&id, &n, &e)
				}
				rows.Close()
			}
		})
	}
}

// BenchmarkDatabaseUpdate benchmarks UPDATE operations
func BenchmarkDatabaseUpdate(b *testing.B) {
	configs := getTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipTest(dbType) {
				b.Skip("Database not available")
			}

			db := setupDatabase(b, config)
			defer db.Close()

			tableName := fmt.Sprintf("bench_update_%s_%d", dbType, time.Now().Unix())
			createTableSQL := getCreateTableSQL(dbType, tableName)
			_, err := db.Exec(createTableSQL)
			if err != nil {
				b.Fatalf("Failed to create table: %v", err)
			}
			defer cleanupTable(db, tableName)

			// Insert test data
			insertSQL := getInsertSQL(dbType, tableName)
			for i := 0; i < 100; i++ {
				name := fmt.Sprintf("User%d", i)
				email := fmt.Sprintf("user%d@example.com", i)
				db.Exec(insertSQL, name, email)
			}

			updateSQL := getUpdateSQL(dbType, tableName)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				name := fmt.Sprintf("User%d", i%100)
				email := fmt.Sprintf("updated%d@example.com", i)
				_, err := db.Exec(updateSQL, email, name)
				if err != nil {
					b.Fatalf("Update failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkDatabaseTransaction benchmarks transaction operations
func BenchmarkDatabaseTransaction(b *testing.B) {
	configs := getTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipTest(dbType) {
				b.Skip("Database not available")
			}

			db := setupDatabase(b, config)
			defer db.Close()

			if !db.SupportsFeature("transactions") {
				b.Skip("Database does not support transactions")
			}

			tableName := fmt.Sprintf("bench_tx_%s_%d", dbType, time.Now().Unix())
			createTableSQL := getCreateTableSQL(dbType, tableName)
			_, err := db.Exec(createTableSQL)
			if err != nil {
				b.Fatalf("Failed to create table: %v", err)
			}
			defer cleanupTable(db, tableName)

			insertSQL := getInsertSQL(dbType, tableName)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tx, err := db.BeginTx(context.Background(), nil)
				if err != nil {
					b.Fatalf("Failed to begin transaction: %v", err)
				}

				name := fmt.Sprintf("User%d", i)
				email := fmt.Sprintf("user%d@example.com", i)
				_, err = tx.Exec(insertSQL, name, email)
				if err != nil {
					tx.Rollback()
					b.Fatalf("Insert failed: %v", err)
				}

				if err := tx.Commit(); err != nil {
					b.Fatalf("Commit failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkDatabaseBulkInsert benchmarks bulk insert operations
func BenchmarkDatabaseBulkInsert(b *testing.B) {
	configs := getTestConfigs()
	batchSize := 100

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipTest(dbType) {
				b.Skip("Database not available")
			}

			db := setupDatabase(b, config)
			defer db.Close()

			tableName := fmt.Sprintf("bench_bulk_%s_%d", dbType, time.Now().Unix())
			createTableSQL := getCreateTableSQL(dbType, tableName)
			_, err := db.Exec(createTableSQL)
			if err != nil {
				b.Fatalf("Failed to create table: %v", err)
			}
			defer cleanupTable(db, tableName)

			insertSQL := getInsertSQL(dbType, tableName)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tx, _ := db.BeginTx(context.Background(), nil)

				for j := 0; j < batchSize; j++ {
					name := fmt.Sprintf("User%d", i*batchSize+j)
					email := fmt.Sprintf("user%d@example.com", i*batchSize+j)
					tx.Exec(insertSQL, name, email)
				}

				tx.Commit()
			}
		})
	}
}

// BenchmarkDatabaseConcurrentReads benchmarks concurrent read operations
func BenchmarkDatabaseConcurrentReads(b *testing.B) {
	configs := getTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipTest(dbType) {
				b.Skip("Database not available")
			}

			db := setupDatabase(b, config)
			defer db.Close()

			tableName := fmt.Sprintf("bench_concurrent_%s_%d", dbType, time.Now().Unix())
			createTableSQL := getCreateTableSQL(dbType, tableName)
			_, err := db.Exec(createTableSQL)
			if err != nil {
				b.Fatalf("Failed to create table: %v", err)
			}
			defer cleanupTable(db, tableName)

			// Insert test data
			insertSQL := getInsertSQL(dbType, tableName)
			for i := 0; i < 100; i++ {
				name := fmt.Sprintf("User%d", i)
				email := fmt.Sprintf("user%d@example.com", i)
				db.Exec(insertSQL, name, email)
			}

			selectSQL := getSelectSQL(dbType, tableName)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					name := fmt.Sprintf("User%d", time.Now().UnixNano()%100)
					rows, err := db.Query(selectSQL, name)
					if err != nil {
						b.Fatalf("Select failed: %v", err)
					}
					for rows.Next() {
						var id int
						var n, e string
						rows.Scan(&id, &n, &e)
					}
					rows.Close()
				}
			})
		})
	}
}

// Helper function to check if test should be skipped
func skipTest(dbType string) bool {
	skipEnv := fmt.Sprintf("SKIP_%s_TESTS", dbType)
	return getEnv(skipEnv, "false") == "true"
}

// Helper to convert testing.B to testing.T interface
func setupDatabase(tb testing.TB, config TestConfig) databases.RelationalDB {
	dbConfig := databases.DBConfig{
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

	db, err := databases.NewRelationalDB(dbConfig)
	if err != nil {
		tb.Fatalf("Failed to create database: %v", err)
	}

	if err := db.Connect(dbConfig); err != nil {
		tb.Fatalf("Failed to connect: %v", err)
	}

	return db
}
