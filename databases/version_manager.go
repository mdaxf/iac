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
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Version represents a database schema version
type Version struct {
	Number      int
	Description string
	AppliedAt   time.Time
	Applied     bool
	Checksum    string
}

// Migration represents a schema migration
type Migration struct {
	Version     int
	Description string
	UpSQL       string
	DownSQL     string
	Checksum    string
}

// VersionManagerConfig configures version management
type VersionManagerConfig struct {
	// VersionTable is the name of the version tracking table
	VersionTable string

	// AutoMigrate enables automatic migration on startup
	AutoMigrate bool

	// ValidateChecksums verifies migration checksums
	ValidateChecksums bool

	// AllowOutOfOrder allows out-of-order migrations
	AllowOutOfOrder bool
}

// DefaultVersionManagerConfig returns default configuration
func DefaultVersionManagerConfig() *VersionManagerConfig {
	return &VersionManagerConfig{
		VersionTable:      "schema_versions",
		AutoMigrate:       false,
		ValidateChecksums: true,
		AllowOutOfOrder:   false,
	}
}

// VersionManager manages database schema versions
type VersionManager struct {
	config     *VersionManagerConfig
	db         *sql.DB
	dbType     string
	migrations []*Migration
	mu         sync.RWMutex
}

// NewVersionManager creates a new version manager
func NewVersionManager(db *sql.DB, dbType string, config *VersionManagerConfig) (*VersionManager, error) {
	if config == nil {
		config = DefaultVersionManagerConfig()
	}

	vm := &VersionManager{
		config:     config,
		db:         db,
		dbType:     dbType,
		migrations: make([]*Migration, 0),
	}

	// Initialize version table
	if err := vm.initVersionTable(); err != nil {
		return nil, fmt.Errorf("failed to initialize version table: %w", err)
	}

	return vm, nil
}

// initVersionTable creates the version tracking table if it doesn't exist
func (vm *VersionManager) initVersionTable() error {
	var createTableSQL string

	switch vm.dbType {
	case "mysql":
		createTableSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				version INT PRIMARY KEY,
				description VARCHAR(255),
				checksum VARCHAR(64),
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				INDEX idx_version (version),
				INDEX idx_applied_at (applied_at)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
		`, vm.config.VersionTable)

	case "postgres":
		createTableSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				version INTEGER PRIMARY KEY,
				description VARCHAR(255),
				checksum VARCHAR(64),
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_version ON %s(version);
			CREATE INDEX IF NOT EXISTS idx_applied_at ON %s(applied_at);
		`, vm.config.VersionTable, vm.config.VersionTable, vm.config.VersionTable)

	case "mssql":
		createTableSQL = fmt.Sprintf(`
			IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = '%s')
			BEGIN
				CREATE TABLE %s (
					version INT PRIMARY KEY,
					description NVARCHAR(255),
					checksum VARCHAR(64),
					applied_at DATETIME DEFAULT GETDATE()
				);
				CREATE INDEX idx_version ON %s(version);
				CREATE INDEX idx_applied_at ON %s(applied_at);
			END
		`, vm.config.VersionTable, vm.config.VersionTable, vm.config.VersionTable, vm.config.VersionTable)

	case "oracle":
		createTableSQL = fmt.Sprintf(`
			BEGIN
				EXECUTE IMMEDIATE 'CREATE TABLE %s (
					version NUMBER PRIMARY KEY,
					description VARCHAR2(255),
					checksum VARCHAR2(64),
					applied_at TIMESTAMP DEFAULT SYSTIMESTAMP
				)';
			EXCEPTION
				WHEN OTHERS THEN
					IF SQLCODE != -955 THEN RAISE; END IF;
			END;
		`, vm.config.VersionTable)

	default:
		return fmt.Errorf("unsupported database type: %s", vm.dbType)
	}

	_, err := vm.db.Exec(createTableSQL)
	return err
}

// RegisterMigration registers a migration
func (vm *VersionManager) RegisterMigration(migration *Migration) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// Check for duplicate version
	for _, m := range vm.migrations {
		if m.Version == migration.Version {
			return fmt.Errorf("migration version %d already registered", migration.Version)
		}
	}

	// Calculate checksum if not provided
	if migration.Checksum == "" {
		migration.Checksum = calculateChecksum(migration.UpSQL + migration.DownSQL)
	}

	vm.migrations = append(vm.migrations, migration)

	// Sort migrations by version
	sort.Slice(vm.migrations, func(i, j int) bool {
		return vm.migrations[i].Version < vm.migrations[j].Version
	})

	return nil
}

// GetCurrentVersion returns the current database version
func (vm *VersionManager) GetCurrentVersion(ctx context.Context) (int, error) {
	query := fmt.Sprintf("SELECT MAX(version) FROM %s", vm.config.VersionTable)

	var version sql.NullInt64
	err := vm.db.QueryRowContext(ctx, query).Scan(&version)
	if err != nil {
		return 0, err
	}

	if !version.Valid {
		return 0, nil
	}

	return int(version.Int64), nil
}

// GetAppliedVersions returns all applied versions
func (vm *VersionManager) GetAppliedVersions(ctx context.Context) ([]*Version, error) {
	query := fmt.Sprintf(`
		SELECT version, description, checksum, applied_at
		FROM %s
		ORDER BY version
	`, vm.config.VersionTable)

	rows, err := vm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make([]*Version, 0)
	for rows.Next() {
		v := &Version{Applied: true}
		if err := rows.Scan(&v.Number, &v.Description, &v.Checksum, &v.AppliedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}

	return versions, nil
}

// GetPendingMigrations returns migrations not yet applied
func (vm *VersionManager) GetPendingMigrations(ctx context.Context) ([]*Migration, error) {
	currentVersion, err := vm.GetCurrentVersion(ctx)
	if err != nil {
		return nil, err
	}

	vm.mu.RLock()
	defer vm.mu.RUnlock()

	pending := make([]*Migration, 0)
	for _, m := range vm.migrations {
		if m.Version > currentVersion {
			pending = append(pending, m)
		}
	}

	return pending, nil
}

// Migrate applies all pending migrations
func (vm *VersionManager) Migrate(ctx context.Context) error {
	pending, err := vm.GetPendingMigrations(ctx)
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		return nil // No migrations to apply
	}

	for _, migration := range pending {
		if err := vm.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", migration.Version, err)
		}
	}

	return nil
}

// MigrateTo migrates to a specific version
func (vm *VersionManager) MigrateTo(ctx context.Context, targetVersion int) error {
	currentVersion, err := vm.GetCurrentVersion(ctx)
	if err != nil {
		return err
	}

	if targetVersion == currentVersion {
		return nil // Already at target version
	}

	if targetVersion > currentVersion {
		// Migrate up
		return vm.migrateUp(ctx, currentVersion, targetVersion)
	}

	// Migrate down
	return vm.migrateDown(ctx, currentVersion, targetVersion)
}

// migrateUp applies migrations from current to target version
func (vm *VersionManager) migrateUp(ctx context.Context, current, target int) error {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	for _, m := range vm.migrations {
		if m.Version > current && m.Version <= target {
			if err := vm.applyMigration(ctx, m); err != nil {
				return err
			}
		}
	}

	return nil
}

// migrateDown rolls back migrations from current to target version
func (vm *VersionManager) migrateDown(ctx context.Context, current, target int) error {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	// Get migrations to rollback in reverse order
	toRollback := make([]*Migration, 0)
	for _, m := range vm.migrations {
		if m.Version > target && m.Version <= current {
			toRollback = append(toRollback, m)
		}
	}

	// Reverse order for rollback
	for i := len(toRollback) - 1; i >= 0; i-- {
		if err := vm.rollbackMigration(ctx, toRollback[i]); err != nil {
			return err
		}
	}

	return nil
}

// applyMigration applies a single migration
func (vm *VersionManager) applyMigration(ctx context.Context, migration *Migration) error {
	// Validate checksum if enabled
	if vm.config.ValidateChecksums {
		if err := vm.validateChecksum(ctx, migration); err != nil {
			return err
		}
	}

	// Start transaction
	tx, err := vm.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration
	if _, err := tx.ExecContext(ctx, migration.UpSQL); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record version
	recordSQL := fmt.Sprintf(`
		INSERT INTO %s (version, description, checksum)
		VALUES (?, ?, ?)
	`, vm.config.VersionTable)

	if _, err := tx.ExecContext(ctx, recordSQL, migration.Version, migration.Description, migration.Checksum); err != nil {
		return fmt.Errorf("failed to record version: %w", err)
	}

	// Commit transaction
	return tx.Commit()
}

// rollbackMigration rolls back a single migration
func (vm *VersionManager) rollbackMigration(ctx context.Context, migration *Migration) error {
	// Start transaction
	tx, err := vm.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute rollback
	if migration.DownSQL != "" {
		if _, err := tx.ExecContext(ctx, migration.DownSQL); err != nil {
			return fmt.Errorf("failed to execute rollback: %w", err)
		}
	}

	// Remove version record
	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE version = ?", vm.config.VersionTable)
	if _, err := tx.ExecContext(ctx, deleteSQL, migration.Version); err != nil {
		return fmt.Errorf("failed to remove version record: %w", err)
	}

	// Commit transaction
	return tx.Commit()
}

// validateChecksum validates migration checksum
func (vm *VersionManager) validateChecksum(ctx context.Context, migration *Migration) error {
	query := fmt.Sprintf("SELECT checksum FROM %s WHERE version = ?", vm.config.VersionTable)

	var storedChecksum string
	err := vm.db.QueryRowContext(ctx, query, migration.Version).Scan(&storedChecksum)

	if err == sql.ErrNoRows {
		return nil // Migration not applied yet
	}

	if err != nil {
		return err
	}

	if storedChecksum != migration.Checksum {
		return fmt.Errorf("checksum mismatch for version %d", migration.Version)
	}

	return nil
}

// GetMigrationHistory returns full migration history
func (vm *VersionManager) GetMigrationHistory(ctx context.Context) ([]MigrationRecord, error) {
	applied, err := vm.GetAppliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	vm.mu.RLock()
	defer vm.mu.RUnlock()

	appliedMap := make(map[int]*Version)
	for _, v := range applied {
		appliedMap[v.Number] = v
	}

	history := make([]MigrationRecord, 0)
	for _, m := range vm.migrations {
		record := MigrationRecord{
			Version:     m.Version,
			Description: m.Description,
			Applied:     false,
		}

		if v, exists := appliedMap[m.Version]; exists {
			record.Applied = true
			record.AppliedAt = v.AppliedAt
		}

		history = append(history, record)
	}

	return history, nil
}

// CheckCompatibility checks if application is compatible with database version
func (vm *VersionManager) CheckCompatibility(ctx context.Context, minVersion, maxVersion int) error {
	currentVersion, err := vm.GetCurrentVersion(ctx)
	if err != nil {
		return err
	}

	if currentVersion < minVersion {
		return fmt.Errorf("database version %d is below minimum required version %d", currentVersion, minVersion)
	}

	if maxVersion > 0 && currentVersion > maxVersion {
		return fmt.Errorf("database version %d is above maximum supported version %d", currentVersion, maxVersion)
	}

	return nil
}

// GetVersionDiff returns the difference between two versions
func (vm *VersionManager) GetVersionDiff(version1, version2 int) ([]*Migration, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	minVersion := version1
	maxVersion := version2
	if version1 > version2 {
		minVersion = version2
		maxVersion = version1
	}

	diff := make([]*Migration, 0)
	for _, m := range vm.migrations {
		if m.Version > minVersion && m.Version <= maxVersion {
			diff = append(diff, m)
		}
	}

	return diff, nil
}

// MigrationRecord represents a migration history record
type MigrationRecord struct {
	Version     int
	Description string
	Applied     bool
	AppliedAt   time.Time
}

// calculateChecksum calculates a checksum for a string
func calculateChecksum(s string) string {
	// Simple checksum - production would use crypto/sha256
	sum := 0
	for _, c := range s {
		sum += int(c)
	}
	return fmt.Sprintf("%x", sum)
}

// Reset removes all version records (dangerous!)
func (vm *VersionManager) Reset(ctx context.Context) error {
	query := fmt.Sprintf("DELETE FROM %s", vm.config.VersionTable)
	_, err := vm.db.ExecContext(ctx, query)
	return err
}

// GetMigrations returns all registered migrations
func (vm *VersionManager) GetMigrations() []*Migration {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	result := make([]*Migration, len(vm.migrations))
	copy(result, vm.migrations)
	return result
}

// MarkAsApplied manually marks a version as applied (for legacy databases)
func (vm *VersionManager) MarkAsApplied(ctx context.Context, version int, description string) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (version, description, checksum)
		VALUES (?, ?, ?)
	`, vm.config.VersionTable)

	_, err := vm.db.ExecContext(ctx, query, version, description, "manual")
	return err
}
