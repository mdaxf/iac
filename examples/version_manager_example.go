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

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	dbconn "github.com/mdaxf/iac/databases"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("IAC Database Versioning Example")
	fmt.Println("================================")

	basicVersioningExample()
	migrationExample()
	rollbackExample()
	compatibilityCheckExample()
}

func basicVersioningExample() {
	fmt.Println("\n1. Basic Database Versioning")
	fmt.Println("------------------------------")

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create version manager
	vm, err := dbconn.NewVersionManager(db, "sqlite3", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Version manager initialized")
	fmt.Println("  Version table: schema_versions")

	// Check current version
	ctx := context.Background()
	version, _ := vm.GetCurrentVersion(ctx)
	fmt.Printf("  Current version: %d\n", version)
}

func migrationExample() {
	fmt.Println("\n2. Schema Migrations")
	fmt.Println("---------------------")

	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	vm, _ := dbconn.NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	// Define migrations
	migrations := []*dbconn.Migration{
		{
			Version:     1,
			Description: "Create users table",
			UpSQL: `
				CREATE TABLE users (
					id INTEGER PRIMARY KEY,
					name TEXT NOT NULL,
					email TEXT UNIQUE NOT NULL
				)
			`,
			DownSQL: "DROP TABLE users",
		},
		{
			Version:     2,
			Description: "Create orders table",
			UpSQL: `
				CREATE TABLE orders (
					id INTEGER PRIMARY KEY,
					user_id INTEGER,
					total REAL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (user_id) REFERENCES users(id)
				)
			`,
			DownSQL: "DROP TABLE orders",
		},
		{
			Version:     3,
			Description: "Add users.created_at column",
			UpSQL:       "ALTER TABLE users ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
			DownSQL:     "ALTER TABLE users DROP COLUMN created_at",
		},
	}

	// Register all migrations
	for _, m := range migrations {
		if err := vm.RegisterMigration(m); err != nil {
			log.Printf("Failed to register migration %d: %v", m.Version, err)
			continue
		}
		fmt.Printf("Registered: v%d - %s\n", m.Version, m.Description)
	}

	// Check pending migrations
	pending, _ := vm.GetPendingMigrations(ctx)
	fmt.Printf("\nPending migrations: %d\n", len(pending))

	// Apply all migrations
	fmt.Println("\nApplying migrations...")
	if err := vm.Migrate(ctx); err != nil {
		log.Printf("Migration failed: %v", err)
	} else {
		fmt.Println("  All migrations applied successfully")
	}

	// Check current version
	version, _ := vm.GetCurrentVersion(ctx)
	fmt.Printf("  Current version: %d\n", version)

	// Show migration history
	history, _ := vm.GetMigrationHistory(ctx)
	fmt.Println("\nMigration History:")
	for _, record := range history {
		status := "pending"
		if record.Applied {
			status = fmt.Sprintf("applied at %s", record.AppliedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("  v%d: %s (%s)\n", record.Version, record.Description, status)
	}
}

func rollbackExample() {
	fmt.Println("\n3. Migration Rollback")
	fmt.Println("----------------------")

	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	vm, _ := dbconn.NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	// Register and apply migrations
	vm.RegisterMigration(&dbconn.Migration{
		Version:     1,
		Description: "Create table1",
		UpSQL:       "CREATE TABLE table1 (id INTEGER)",
		DownSQL:     "DROP TABLE table1",
	})
	vm.RegisterMigration(&dbconn.Migration{
		Version:     2,
		Description: "Create table2",
		UpSQL:       "CREATE TABLE table2 (id INTEGER)",
		DownSQL:     "DROP TABLE table2",
	})
	vm.RegisterMigration(&dbconn.Migration{
		Version:     3,
		Description: "Create table3",
		UpSQL:       "CREATE TABLE table3 (id INTEGER)",
		DownSQL:     "DROP TABLE table3",
	})

	vm.Migrate(ctx)
	fmt.Println("Applied migrations to version 3")

	currentVersion, _ := vm.GetCurrentVersion(ctx)
	fmt.Printf("  Current version: %d\n", currentVersion)

	// Rollback to version 1
	fmt.Println("\nRolling back to version 1...")
	if err := vm.MigrateTo(ctx, 1); err != nil {
		log.Printf("Rollback failed: %v", err)
	} else {
		fmt.Println("  Rollback successful")
	}

	currentVersion, _ = vm.GetCurrentVersion(ctx)
	fmt.Printf("  Current version: %d\n", currentVersion)

	// Migrate forward again
	fmt.Println("\nMigrating forward to version 2...")
	vm.MigrateTo(ctx, 2)

	currentVersion, _ = vm.GetCurrentVersion(ctx)
	fmt.Printf("  Current version: %d\n", currentVersion)
}

func compatibilityCheckExample() {
	fmt.Println("\n4. Compatibility Checking")
	fmt.Println("--------------------------")

	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	vm, _ := dbconn.NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	// Apply some migrations
	vm.RegisterMigration(&dbconn.Migration{
		Version: 5,
		UpSQL:   "CREATE TABLE test (id INTEGER)",
	})
	vm.Migrate(ctx)

	currentVersion, _ := vm.GetCurrentVersion(ctx)
	fmt.Printf("Database version: %d\n", currentVersion)

	// Check compatibility with application requirements
	minVersion := 3
	maxVersion := 10

	fmt.Printf("\nApplication requirements:\n")
	fmt.Printf("  Minimum version: %d\n", minVersion)
	fmt.Printf("  Maximum version: %d\n", maxVersion)

	err := vm.CheckCompatibility(ctx, minVersion, maxVersion)
	if err != nil {
		fmt.Printf("  Status: INCOMPATIBLE - %v\n", err)
	} else {
		fmt.Printf("  Status: COMPATIBLE\n")
	}

	// Test with incompatible requirements
	fmt.Println("\nTesting with min version 6...")
	err = vm.CheckCompatibility(ctx, 6, 10)
	if err != nil {
		fmt.Printf("  Status: INCOMPATIBLE - %v\n", err)
	}

	fmt.Println("\nTesting with max version 4...")
	err = vm.CheckCompatibility(ctx, 1, 4)
	if err != nil {
		fmt.Printf("  Status: INCOMPATIBLE - %v\n", err)
	}
}

func productionExample() {
	fmt.Println("\n5. Production Setup")
	fmt.Println("--------------------")

	// Connect to production database
	// db, _ := sql.Open("postgres", "host=localhost dbname=production...")
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	// Configure version manager
	config := &dbconn.VersionManagerConfig{
		VersionTable:      "schema_versions",
		AutoMigrate:       false, // Manual control in production
		ValidateChecksums: true,  // Always validate in production
		AllowOutOfOrder:   false, // Strict ordering
	}

	vm, _ := dbconn.NewVersionManager(db, "postgres", config)
	ctx := context.Background()

	// Define all application migrations
	migrations := []struct {
		version     int
		description string
		upSQL       string
		downSQL     string
	}{
		{1, "Initial schema", "CREATE TABLE ...", "DROP TABLE ..."},
		{2, "Add indexes", "CREATE INDEX ...", "DROP INDEX ..."},
		{3, "Add user roles", "CREATE TABLE roles ...", "DROP TABLE roles"},
		// ... more migrations
	}

	// Register all migrations
	for _, m := range migrations {
		vm.RegisterMigration(&dbconn.Migration{
			Version:     m.version,
			Description: m.description,
			UpSQL:       m.upSQL,
			DownSQL:     m.downSQL,
		})
	}

	// Check current version
	currentVersion, _ := vm.GetCurrentVersion(ctx)
	fmt.Printf("Current database version: %d\n", currentVersion)

	// Check if migrations needed
	pending, _ := vm.GetPendingMigrations(ctx)
	if len(pending) > 0 {
		fmt.Printf("  %d pending migrations\n", len(pending))
		fmt.Println("  Run 'migrate' command to apply")
	} else {
		fmt.Println("  Database is up to date")
	}

	// Check compatibility with this application version
	appMinVersion := 3
	appMaxVersion := 10

	if err := vm.CheckCompatibility(ctx, appMinVersion, appMaxVersion); err != nil {
		log.Fatalf("Database incompatible with application: %v", err)
	}

	fmt.Println("  Database compatible with application")
}

func legacyDatabaseExample() {
	fmt.Println("\n6. Legacy Database Integration")
	fmt.Println("--------------------------------")

	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	vm, _ := dbconn.NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	// Existing database already at version 10
	fmt.Println("Integrating with legacy database at version 10...")

	// Mark current version without running migrations
	if err := vm.MarkAsApplied(ctx, 10, "Legacy schema baseline"); err != nil {
		log.Printf("Failed to mark version: %v", err)
	} else {
		fmt.Println("  Marked version 10 as applied")
	}

	currentVersion, _ := vm.GetCurrentVersion(ctx)
	fmt.Printf("  Current version: %d\n", currentVersion)

	// Now can register new migrations starting from 11
	vm.RegisterMigration(&dbconn.Migration{
		Version:     11,
		Description: "First new migration",
		UpSQL:       "CREATE TABLE new_feature (id INTEGER)",
	})

	pending, _ := vm.GetPendingMigrations(ctx)
	fmt.Printf("  Pending migrations: %d\n", len(pending))
}
