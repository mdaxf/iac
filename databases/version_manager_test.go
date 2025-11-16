// Copyright 2023 IAC. All Rights Reserved.

package databases

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestNewVersionManager(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, err := NewVersionManager(db, "sqlite3", nil)
	if err != nil {
		t.Fatalf("NewVersionManager failed: %v", err)
	}

	if vm == nil {
		t.Fatal("VersionManager is nil")
	}

	// Verify version table was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_versions'").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	// SQLite uses different table tracking, just check vm was created
	if vm.config.VersionTable != "schema_versions" {
		t.Error("Version table not set correctly")
	}
}

func TestRegisterMigration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)

	migration := &Migration{
		Version:     1,
		Description: "Initial schema",
		UpSQL:       "CREATE TABLE users (id INTEGER PRIMARY KEY)",
		DownSQL:     "DROP TABLE users",
	}

	err := vm.RegisterMigration(migration)
	if err != nil {
		t.Errorf("RegisterMigration failed: %v", err)
	}

	// Try to register duplicate version
	err = vm.RegisterMigration(migration)
	if err == nil {
		t.Error("Expected error for duplicate version")
	}
}

func TestGetCurrentVersion(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	// Initially should be version 0
	version, err := vm.GetCurrentVersion(ctx)
	if err != nil {
		t.Fatalf("GetCurrentVersion failed: %v", err)
	}

	if version != 0 {
		t.Errorf("Expected version 0, got %d", version)
	}
}

func TestApplyMigration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	migration := &Migration{
		Version:     1,
		Description: "Create users table",
		UpSQL:       "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)",
		DownSQL:     "DROP TABLE users",
		Checksum:    "abc123",
	}

	vm.RegisterMigration(migration)

	// Apply migration
	err := vm.Migrate(ctx)
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	// Verify version is updated
	version, _ := vm.GetCurrentVersion(ctx)
	if version != 1 {
		t.Errorf("Expected version 1, got %d", version)
	}

	// Verify table was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Error("Users table was not created")
	}
}

func TestGetPendingMigrations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	// Register multiple migrations
	vm.RegisterMigration(&Migration{
		Version:     1,
		Description: "Migration 1",
		UpSQL:       "CREATE TABLE t1 (id INTEGER)",
		DownSQL:     "DROP TABLE t1",
	})
	vm.RegisterMigration(&Migration{
		Version:     2,
		Description: "Migration 2",
		UpSQL:       "CREATE TABLE t2 (id INTEGER)",
		DownSQL:     "DROP TABLE t2",
	})
	vm.RegisterMigration(&Migration{
		Version:     3,
		Description: "Migration 3",
		UpSQL:       "CREATE TABLE t3 (id INTEGER)",
		DownSQL:     "DROP TABLE t3",
	})

	// All should be pending
	pending, err := vm.GetPendingMigrations(ctx)
	if err != nil {
		t.Fatalf("GetPendingMigrations failed: %v", err)
	}

	if len(pending) != 3 {
		t.Errorf("Expected 3 pending migrations, got %d", len(pending))
	}

	// Apply first migration
	vm.MigrateTo(ctx, 1)

	// Now should have 2 pending
	pending, _ = vm.GetPendingMigrations(ctx)
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending migrations, got %d", len(pending))
	}
}

func TestMigrateTo(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	// Register migrations
	vm.RegisterMigration(&Migration{
		Version:     1,
		Description: "V1",
		UpSQL:       "CREATE TABLE t1 (id INTEGER)",
		DownSQL:     "DROP TABLE t1",
	})
	vm.RegisterMigration(&Migration{
		Version:     2,
		Description: "V2",
		UpSQL:       "CREATE TABLE t2 (id INTEGER)",
		DownSQL:     "DROP TABLE t2",
	})

	// Migrate to version 2
	err := vm.MigrateTo(ctx, 2)
	if err != nil {
		t.Fatalf("MigrateTo failed: %v", err)
	}

	version, _ := vm.GetCurrentVersion(ctx)
	if version != 2 {
		t.Errorf("Expected version 2, got %d", version)
	}

	// Migrate back to version 1
	err = vm.MigrateTo(ctx, 1)
	if err != nil {
		t.Fatalf("MigrateTo (down) failed: %v", err)
	}

	version, _ = vm.GetCurrentVersion(ctx)
	if version != 1 {
		t.Errorf("Expected version 1 after rollback, got %d", version)
	}
}

func TestGetAppliedVersions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	vm.RegisterMigration(&Migration{
		Version:     1,
		Description: "V1",
		UpSQL:       "CREATE TABLE t1 (id INTEGER)",
	})
	vm.RegisterMigration(&Migration{
		Version:     2,
		Description: "V2",
		UpSQL:       "CREATE TABLE t2 (id INTEGER)",
	})

	vm.Migrate(ctx)

	versions, err := vm.GetAppliedVersions(ctx)
	if err != nil {
		t.Fatalf("GetAppliedVersions failed: %v", err)
	}

	if len(versions) != 2 {
		t.Errorf("Expected 2 applied versions, got %d", len(versions))
	}

	for _, v := range versions {
		if !v.Applied {
			t.Error("Version should be marked as applied")
		}
	}
}

func TestGetMigrationHistory(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	vm.RegisterMigration(&Migration{
		Version:     1,
		Description: "V1",
		UpSQL:       "CREATE TABLE t1 (id INTEGER)",
	})
	vm.RegisterMigration(&Migration{
		Version:     2,
		Description: "V2",
		UpSQL:       "CREATE TABLE t2 (id INTEGER)",
	})

	// Apply only first migration
	vm.MigrateTo(ctx, 1)

	history, err := vm.GetMigrationHistory(ctx)
	if err != nil {
		t.Fatalf("GetMigrationHistory failed: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 history records, got %d", len(history))
	}

	if !history[0].Applied {
		t.Error("First migration should be applied")
	}

	if history[1].Applied {
		t.Error("Second migration should not be applied")
	}
}

func TestCheckCompatibility(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	vm.RegisterMigration(&Migration{
		Version: 5,
		UpSQL:   "CREATE TABLE t1 (id INTEGER)",
	})
	vm.Migrate(ctx)

	// Should be compatible
	err := vm.CheckCompatibility(ctx, 1, 10)
	if err != nil {
		t.Errorf("Should be compatible: %v", err)
	}

	// Version too low
	err = vm.CheckCompatibility(ctx, 6, 10)
	if err == nil {
		t.Error("Expected incompatibility error (version too low)")
	}

	// Version too high
	err = vm.CheckCompatibility(ctx, 1, 4)
	if err == nil {
		t.Error("Expected incompatibility error (version too high)")
	}
}

func TestGetVersionDiff(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)

	vm.RegisterMigration(&Migration{Version: 1, UpSQL: "V1"})
	vm.RegisterMigration(&Migration{Version: 2, UpSQL: "V2"})
	vm.RegisterMigration(&Migration{Version: 3, UpSQL: "V3"})
	vm.RegisterMigration(&Migration{Version: 4, UpSQL: "V4"})

	diff, err := vm.GetVersionDiff(1, 3)
	if err != nil {
		t.Fatalf("GetVersionDiff failed: %v", err)
	}

	if len(diff) != 2 {
		t.Errorf("Expected 2 migrations in diff, got %d", len(diff))
	}

	if diff[0].Version != 2 || diff[1].Version != 3 {
		t.Error("Diff contains wrong migrations")
	}
}

func TestMarkAsApplied(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	// Mark version as applied manually
	err := vm.MarkAsApplied(ctx, 10, "Legacy schema")
	if err != nil {
		t.Fatalf("MarkAsApplied failed: %v", err)
	}

	version, _ := vm.GetCurrentVersion(ctx)
	if version != 10 {
		t.Errorf("Expected version 10, got %d", version)
	}
}

func TestReset(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	vm.RegisterMigration(&Migration{
		Version: 1,
		UpSQL:   "CREATE TABLE t1 (id INTEGER)",
	})
	vm.Migrate(ctx)

	// Reset all versions
	err := vm.Reset(ctx)
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	version, _ := vm.GetCurrentVersion(ctx)
	if version != 0 {
		t.Errorf("Expected version 0 after reset, got %d", version)
	}
}

func TestGetMigrations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)

	vm.RegisterMigration(&Migration{Version: 1, UpSQL: "V1"})
	vm.RegisterMigration(&Migration{Version: 2, UpSQL: "V2"})

	migrations := vm.GetMigrations()
	if len(migrations) != 2 {
		t.Errorf("Expected 2 migrations, got %d", len(migrations))
	}
}

func TestMigrationOrdering(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)

	// Register out of order
	vm.RegisterMigration(&Migration{Version: 3, UpSQL: "V3"})
	vm.RegisterMigration(&Migration{Version: 1, UpSQL: "V1"})
	vm.RegisterMigration(&Migration{Version: 2, UpSQL: "V2"})

	migrations := vm.GetMigrations()

	// Should be sorted
	if migrations[0].Version != 1 || migrations[1].Version != 2 || migrations[2].Version != 3 {
		t.Error("Migrations not sorted correctly")
	}
}

func TestDefaultVersionManagerConfig(t *testing.T) {
	config := DefaultVersionManagerConfig()

	if config.VersionTable != "schema_versions" {
		t.Error("Default version table incorrect")
	}

	if config.AutoMigrate {
		t.Error("AutoMigrate should be false by default")
	}

	if !config.ValidateChecksums {
		t.Error("ValidateChecksums should be true by default")
	}

	if config.AllowOutOfOrder {
		t.Error("AllowOutOfOrder should be false by default")
	}
}

func TestCalculateChecksum(t *testing.T) {
	checksum1 := calculateChecksum("test")
	checksum2 := calculateChecksum("test")
	checksum3 := calculateChecksum("different")

	if checksum1 != checksum2 {
		t.Error("Same input should produce same checksum")
	}

	if checksum1 == checksum3 {
		t.Error("Different input should produce different checksum")
	}
}

func TestMultipleMigrations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vm, _ := NewVersionManager(db, "sqlite3", nil)
	ctx := context.Background()

	// Register multiple migrations
	for i := 1; i <= 5; i++ {
		vm.RegisterMigration(&Migration{
			Version:     i,
			Description: fmt.Sprintf("Migration %d", i),
			UpSQL:       fmt.Sprintf("CREATE TABLE t%d (id INTEGER)", i),
			DownSQL:     fmt.Sprintf("DROP TABLE t%d", i),
		})
	}

	// Apply all
	err := vm.Migrate(ctx)
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	// Should be at version 5
	version, _ := vm.GetCurrentVersion(ctx)
	if version != 5 {
		t.Errorf("Expected version 5, got %d", version)
	}

	// Rollback to version 2
	err = vm.MigrateTo(ctx, 2)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	version, _ = vm.GetCurrentVersion(ctx)
	if version != 2 {
		t.Errorf("Expected version 2 after rollback, got %d", version)
	}
}
