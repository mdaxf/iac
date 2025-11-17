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
	fmt.Println("IAC Database Backup & Restore Example")
	fmt.Println("======================================")

	// Example 1: Basic backup
	basicBackupExample()

	// Example 2: Backup with custom configuration
	customConfigExample()

	// Example 3: List and manage backups
	manageBackupsExample()

	// Example 4: Restore from backup
	restoreExample()

	// Example 5: Scheduled backups
	scheduledBackupExample()
}

func basicBackupExample() {
	fmt.Println("\n1. Basic Backup")
	fmt.Println("----------------")

	// Create backup manager with default config
	bm, err := dbconn.NewBackupManager(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Create test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create some test data
	db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	db.Exec("INSERT INTO users VALUES (1, 'Alice'), (2, 'Bob')")

	ctx := context.Background()

	// Note: For SQLite in-memory databases, actual backup would need
	// special handling. This example shows the API usage.
	fmt.Println("Creating backup...")
	fmt.Println("  Database: testdb")
	fmt.Println("  Type: sqlite3")
	fmt.Println("  Format: SQL")

	// In production with real databases:
	// backup, err := bm.Backup(ctx, db, "mysql", "production_db")
	// For this example, we'll simulate
	fmt.Println("  Status: Backup created successfully (simulated)")
	fmt.Println("  File: ./backups/testdb_sqlite3_20250116_120000.sql")
}

func customConfigExample() {
	fmt.Println("\n2. Custom Backup Configuration")
	fmt.Println("--------------------------------")

	// Configure backup manager
	config := &dbconn.BackupConfig{
		BackupDir:          "./my_backups",
		Format:             dbconn.CompressedFormat,
		Compression:        true,
		MaxBackups:         20,
		RetentionDays:      60,
		VerifyBackup:       true,
		ScheduleExpression: "0 3 * * *", // 3 AM daily
	}

	bm, err := dbconn.NewBackupManager(config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Backup Configuration:")
	fmt.Printf("  Backup Directory: %s\n", config.BackupDir)
	fmt.Printf("  Format: %s\n", config.Format)
	fmt.Printf("  Compression: %v\n", config.Compression)
	fmt.Printf("  Max Backups: %d\n", config.MaxBackups)
	fmt.Printf("  Retention: %d days\n", config.RetentionDays)
	fmt.Printf("  Verification: %v\n", config.VerifyBackup)
	fmt.Printf("  Schedule: %s\n", config.ScheduleExpression)

	_ = bm // Use bm to avoid unused variable error
}

func manageBackupsExample() {
	fmt.Println("\n3. Manage Backups")
	fmt.Println("------------------")

	bm, _ := dbconn.NewBackupManager(nil)

	// List all backups
	backups := bm.ListBackups()
	fmt.Printf("Total backups: %d\n", len(backups))

	if len(backups) > 0 {
		fmt.Println("\nBackup List:")
		for _, backup := range backups {
			fmt.Printf("  ID: %s\n", backup.ID)
			fmt.Printf("    Database: %s (%s)\n", backup.DatabaseName, backup.DatabaseType)
			fmt.Printf("    Created: %s\n", backup.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("    Size: %d bytes\n", backup.FileSize)
			fmt.Printf("    Verified: %v\n", backup.Verified)
			fmt.Println()
		}

		// Get specific backup
		if len(backups) > 0 {
			firstBackup := backups[0]
			backup, err := bm.GetBackup(firstBackup.ID)
			if err == nil {
				fmt.Printf("Retrieved backup: %s\n", backup.ID)
				fmt.Printf("  Path: %s\n", backup.FilePath)
			}
		}
	} else {
		fmt.Println("  No backups found")
	}
}

func restoreExample() {
	fmt.Println("\n4. Restore from Backup")
	fmt.Println("-----------------------")

	bm, _ := dbconn.NewBackupManager(nil)

	// Open database connection
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Restore options
	options := &dbconn.RestoreOptions{
		DropExisting:     true,
		CreateDatabase:   true,
		SkipVerification: false,
		TargetDatabase:   "restored_db",
	}

	fmt.Println("Restore Options:")
	fmt.Printf("  Drop Existing: %v\n", options.DropExisting)
	fmt.Printf("  Create Database: %v\n", options.CreateDatabase)
	fmt.Printf("  Skip Verification: %v\n", options.SkipVerification)
	fmt.Printf("  Target Database: %s\n", options.TargetDatabase)

	// In production:
	// err = bm.Restore(ctx, db, "backup_id_here", options)
	// For this example:
	fmt.Println("\nRestoring backup...")
	fmt.Println("  Status: Restore completed successfully (simulated)")
	fmt.Println("  Records restored: 1,234,567")
	fmt.Println("  Duration: 45 seconds")

	_ = ctx
}

func scheduledBackupExample() {
	fmt.Println("\n5. Scheduled Backups")
	fmt.Println("---------------------")

	config := dbconn.DefaultBackupConfig()
	config.ScheduleExpression = "0 2 * * *" // 2 AM daily

	bm, _ := dbconn.NewBackupManager(config)

	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	ctx := context.Background()

	// Start scheduler
	fmt.Println("Starting backup scheduler...")
	fmt.Printf("  Schedule: %s (2 AM daily)\n", config.ScheduleExpression)

	// In production:
	// bm.StartScheduler(ctx, db, "mysql", "production_db")
	// defer bm.StopScheduler()

	fmt.Println("  Status: Scheduler started")
	fmt.Println("  Next backup: Tomorrow at 2:00 AM")

	_ = ctx
}

// Complete example with MySQL
func completeMySQLExample() {
	fmt.Println("\n6. Complete MySQL Backup Example")
	fmt.Println("----------------------------------")

	// Configure for MySQL
	config := &dbconn.BackupConfig{
		BackupDir:     "/var/backups/mysql",
		Format:        dbconn.SQLFormat,
		Compression:   true,
		MaxBackups:    30,
		RetentionDays: 90,
		VerifyBackup:  true,
	}

	bm, err := dbconn.NewBackupManager(config)
	if err != nil {
		log.Fatalf("Failed to create backup manager: %v", err)
	}

	// Connect to MySQL
	// db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/mydb")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()

	fmt.Println("MySQL Backup Configuration:")
	fmt.Printf("  Backup Directory: %s\n", config.BackupDir)
	fmt.Printf("  Format: %s (with compression)\n", config.Format)
	fmt.Printf("  Retention: Keep %d backups for %d days\n",
		config.MaxBackups, config.RetentionDays)

	// Create backup
	ctx := context.Background()
	// backup, err := bm.Backup(ctx, db, "mysql", "mydb")
	// if err != nil {
	// 	log.Fatalf("Backup failed: %v", err)
	// }

	fmt.Println("\nBackup Process:")
	fmt.Println("  1. Connecting to database...")
	fmt.Println("  2. Locking tables for consistent snapshot...")
	fmt.Println("  3. Dumping database schema...")
	fmt.Println("  4. Dumping table data...")
	fmt.Println("  5. Compressing backup file...")
	fmt.Println("  6. Verifying backup integrity...")
	fmt.Println("  7. Cleaning up old backups...")

	// fmt.Printf("\nBackup completed:\n")
	// fmt.Printf("  Backup ID: %s\n", backup.ID)
	// fmt.Printf("  File: %s\n", backup.FilePath)
	// fmt.Printf("  Size: %d bytes\n", backup.FileSize)
	// fmt.Printf("  Duration: %v\n", time.Since(backup.CreatedAt))

	_ = ctx
}

// Complete example with PostgreSQL
func completePostgreSQLExample() {
	fmt.Println("\n7. Complete PostgreSQL Backup Example")
	fmt.Println("---------------------------------------")

	config := &dbconn.BackupConfig{
		BackupDir:     "/var/backups/postgres",
		Format:        dbconn.BinaryFormat,
		Compression:   true,
		MaxBackups:    14,
		RetentionDays: 30,
		VerifyBackup:  true,
	}

	bm, err := dbconn.NewBackupManager(config)
	if err != nil {
		log.Fatalf("Failed to create backup manager: %v", err)
	}

	fmt.Println("PostgreSQL Backup Configuration:")
	fmt.Printf("  Backup Directory: %s\n", config.BackupDir)
	fmt.Printf("  Format: %s (pg_dump custom format)\n", config.Format)
	fmt.Printf("  Retention: Keep last %d backups\n", config.MaxBackups)

	// Connect to PostgreSQL
	// db, err := sql.Open("postgres",
	// 	"host=localhost port=5432 user=postgres password=secret dbname=mydb sslmode=disable")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()

	ctx := context.Background()

	// Create backup
	// backup, err := bm.Backup(ctx, db, "postgres", "mydb")

	fmt.Println("\nBackup Process:")
	fmt.Println("  Using pg_dump with custom format...")
	fmt.Println("  Compression level: 9 (maximum)")
	fmt.Println("  Status: Backup completed successfully (simulated)")

	// Restore example
	fmt.Println("\nRestore Process:")
	options := &dbconn.RestoreOptions{
		DropExisting:   false,
		CreateDatabase: false,
		TargetDatabase: "mydb_restored",
	}

	// err = bm.Restore(ctx, db, backup.ID, options)
	fmt.Printf("  Restoring to: %s\n", options.TargetDatabase)
	fmt.Println("  Using pg_restore...")
	fmt.Println("  Status: Restore completed (simulated)")

	_ = ctx
}

// Point-in-time recovery example
func pointInTimeRecoveryExample() {
	fmt.Println("\n8. Point-in-Time Recovery")
	fmt.Println("--------------------------")

	bm, _ := dbconn.NewBackupManager(nil)

	// Find backup closest to target time
	// targetTime := time.Now().Add(-24 * time.Hour)

	fmt.Println("Point-in-Time Recovery Process:")
	fmt.Println("  1. Identify target recovery time")
	fmt.Println("  2. Find appropriate backup")
	fmt.Println("  3. Restore base backup")
	fmt.Println("  4. Apply transaction logs up to target time")
	fmt.Println("  5. Verify data consistency")

	options := &dbconn.RestoreOptions{
		DropExisting: true,
		// PointInTime:  &targetTime,
	}

	fmt.Printf("\nRestore Options:\n")
	fmt.Printf("  Drop Existing: %v\n", options.DropExisting)
	fmt.Println("  Target Time: 2025-01-15 14:30:00")
	fmt.Println("  Status: Recovery completed (simulated)")

	_ = bm
}
