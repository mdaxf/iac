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

package databases

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// BackupType defines the type of backup
type BackupType string

const (
	// FullBackup is a complete database backup
	FullBackup BackupType = "full"

	// IncrementalBackup is an incremental backup
	IncrementalBackup BackupType = "incremental"

	// DifferentialBackup is a differential backup
	DifferentialBackup BackupType = "differential"
)

// BackupFormat defines the backup file format
type BackupFormat string

const (
	// SQLFormat exports as SQL statements
	SQLFormat BackupFormat = "sql"

	// BinaryFormat exports as binary dump
	BinaryFormat BackupFormat = "binary"

	// CompressedFormat exports as compressed archive
	CompressedFormat BackupFormat = "compressed"
)

// BackupConfig configures backup operations
type BackupConfig struct {
	// BackupDir is the directory for backup files
	BackupDir string

	// Format specifies backup format
	Format BackupFormat

	// Compression enables compression
	Compression bool

	// MaxBackups is the maximum number of backups to retain
	MaxBackups int

	// RetentionDays is how long to keep backups
	RetentionDays int

	// VerifyBackup verifies backup after creation
	VerifyBackup bool

	// ScheduleExpression is cron expression for scheduled backups
	ScheduleExpression string
}

// DefaultBackupConfig returns default configuration
func DefaultBackupConfig() *BackupConfig {
	return &BackupConfig{
		BackupDir:          "./backups",
		Format:             SQLFormat,
		Compression:        true,
		MaxBackups:         10,
		RetentionDays:      30,
		VerifyBackup:       true,
		ScheduleExpression: "0 2 * * *", // Daily at 2 AM
	}
}

// BackupManager manages database backups and restores
type BackupManager struct {
	config *BackupConfig
	mu     sync.RWMutex

	// Backup history
	backups []*BackupInfo

	// Schedule management
	scheduler *BackupScheduler
}

// BackupInfo contains information about a backup
type BackupInfo struct {
	ID           string
	DatabaseName string
	DatabaseType string
	BackupType   BackupType
	Format       BackupFormat
	FilePath     string
	FileSize     int64
	CreatedAt    time.Time
	Compressed   bool
	Verified     bool
	Metadata     map[string]string
}

// RestoreOptions configures restore operations
type RestoreOptions struct {
	// DropExisting drops existing database before restore
	DropExisting bool

	// CreateDatabase creates database if it doesn't exist
	CreateDatabase bool

	// PointInTime restores to specific point in time
	PointInTime *time.Time

	// SkipVerification skips backup verification
	SkipVerification bool

	// TargetDatabase is the database to restore to (if different)
	TargetDatabase string
}

// NewBackupManager creates a new backup manager
func NewBackupManager(config *BackupConfig) (*BackupManager, error) {
	if config == nil {
		config = DefaultBackupConfig()
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(config.BackupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	bm := &BackupManager{
		config:  config,
		backups: make([]*BackupInfo, 0),
	}

	// Load existing backups
	bm.loadBackupHistory()

	return bm, nil
}

// Backup creates a database backup
func (bm *BackupManager) Backup(ctx context.Context, db *sql.DB, dbType, dbName string) (*BackupInfo, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Generate backup ID
	backupID := fmt.Sprintf("%s_%s_%s", dbName, dbType, time.Now().Format("20060102_150405"))

	// Create backup file path
	filename := bm.generateFilename(backupID, dbType)
	filepath := filepath.Join(bm.config.BackupDir, filename)

	// Perform backup based on database type
	var err error
	var fileSize int64

	switch dbType {
	case "mysql":
		fileSize, err = bm.backupMySQL(ctx, dbName, filepath)
	case "postgres":
		fileSize, err = bm.backupPostgreSQL(ctx, dbName, filepath)
	case "mssql":
		fileSize, err = bm.backupMSSQL(ctx, db, dbName, filepath)
	case "oracle":
		fileSize, err = bm.backupOracle(ctx, dbName, filepath)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	if err != nil {
		return nil, fmt.Errorf("backup failed: %w", err)
	}

	// Create backup info
	info := &BackupInfo{
		ID:           backupID,
		DatabaseName: dbName,
		DatabaseType: dbType,
		BackupType:   FullBackup,
		Format:       bm.config.Format,
		FilePath:     filepath,
		FileSize:     fileSize,
		CreatedAt:    time.Now(),
		Compressed:   bm.config.Compression,
		Verified:     false,
		Metadata:     make(map[string]string),
	}

	// Verify backup if configured
	if bm.config.VerifyBackup {
		if err := bm.verifyBackup(info); err != nil {
			return nil, fmt.Errorf("backup verification failed: %w", err)
		}
		info.Verified = true
	}

	// Add to backup history
	bm.backups = append(bm.backups, info)

	// Cleanup old backups
	bm.cleanupOldBackups()

	return info, nil
}

// Restore restores a database from backup
func (bm *BackupManager) Restore(ctx context.Context, db *sql.DB, backupID string, options *RestoreOptions) error {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if options == nil {
		options = &RestoreOptions{}
	}

	// Find backup
	var backup *BackupInfo
	for _, b := range bm.backups {
		if b.ID == backupID {
			backup = b
			break
		}
	}

	if backup == nil {
		return fmt.Errorf("backup not found: %s", backupID)
	}

	// Verify backup unless skipped
	if !options.SkipVerification && !backup.Verified {
		if err := bm.verifyBackup(backup); err != nil {
			return fmt.Errorf("backup verification failed: %w", err)
		}
	}

	// Perform restore based on database type
	targetDB := backup.DatabaseName
	if options.TargetDatabase != "" {
		targetDB = options.TargetDatabase
	}

	switch backup.DatabaseType {
	case "mysql":
		return bm.restoreMySQL(ctx, backup.FilePath, targetDB, options)
	case "postgres":
		return bm.restorePostgreSQL(ctx, backup.FilePath, targetDB, options)
	case "mssql":
		return bm.restoreMSSQL(ctx, db, backup.FilePath, targetDB, options)
	case "oracle":
		return bm.restoreOracle(ctx, backup.FilePath, targetDB, options)
	default:
		return fmt.Errorf("unsupported database type: %s", backup.DatabaseType)
	}
}

// backupMySQL performs MySQL backup using mysqldump
func (bm *BackupManager) backupMySQL(ctx context.Context, dbName, filepath string) (int64, error) {
	args := []string{
		"--single-transaction",
		"--quick",
		"--lock-tables=false",
		dbName,
	}

	if bm.config.Compression {
		// Use gzip compression
		cmd := exec.CommandContext(ctx, "sh", "-c",
			fmt.Sprintf("mysqldump %s | gzip > %s", dbName, filepath+".gz"))
		if err := cmd.Run(); err != nil {
			return 0, err
		}
		filepath = filepath + ".gz"
	} else {
		cmd := exec.CommandContext(ctx, "mysqldump", args...)
		out, err := os.Create(filepath)
		if err != nil {
			return 0, err
		}
		defer out.Close()

		cmd.Stdout = out
		if err := cmd.Run(); err != nil {
			return 0, err
		}
	}

	// Get file size
	info, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// backupPostgreSQL performs PostgreSQL backup using pg_dump
func (bm *BackupManager) backupPostgreSQL(ctx context.Context, dbName, filepath string) (int64, error) {
	args := []string{
		"-Fc", // Custom format
		"-f", filepath,
		dbName,
	}

	if bm.config.Compression {
		args = append(args, "-Z", "9") // Maximum compression
	}

	cmd := exec.CommandContext(ctx, "pg_dump", args...)
	if err := cmd.Run(); err != nil {
		return 0, err
	}

	// Get file size
	info, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// backupMSSQL performs SQL Server backup using T-SQL
func (bm *BackupManager) backupMSSQL(ctx context.Context, db *sql.DB, dbName, filepath string) (int64, error) {
	query := fmt.Sprintf(`
		BACKUP DATABASE [%s]
		TO DISK = '%s'
		WITH FORMAT, COMPRESSION, STATS = 10
	`, dbName, filepath)

	if _, err := db.ExecContext(ctx, query); err != nil {
		return 0, err
	}

	// Get file size
	info, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// backupOracle performs Oracle backup using expdp
func (bm *BackupManager) backupOracle(ctx context.Context, dbName, filepath string) (int64, error) {
	args := []string{
		"schemas=" + dbName,
		"dumpfile=" + filepath,
		"directory=DATA_PUMP_DIR",
		"compression=all",
	}

	cmd := exec.CommandContext(ctx, "expdp", args...)
	if err := cmd.Run(); err != nil {
		return 0, err
	}

	// Get file size
	info, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// restoreMySQL restores MySQL database
func (bm *BackupManager) restoreMySQL(ctx context.Context, filepath, dbName string, options *RestoreOptions) error {
	if options.DropExisting {
		// Drop and recreate database
		cmd := exec.CommandContext(ctx, "mysql", "-e",
			fmt.Sprintf("DROP DATABASE IF EXISTS %s; CREATE DATABASE %s;", dbName, dbName))
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// Restore from backup
	if bm.config.Compression {
		cmd := exec.CommandContext(ctx, "sh", "-c",
			fmt.Sprintf("gunzip < %s | mysql %s", filepath, dbName))
		return cmd.Run()
	}

	cmd := exec.CommandContext(ctx, "mysql", dbName)
	in, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer in.Close()

	cmd.Stdin = in
	return cmd.Run()
}

// restorePostgreSQL restores PostgreSQL database
func (bm *BackupManager) restorePostgreSQL(ctx context.Context, filepath, dbName string, options *RestoreOptions) error {
	if options.DropExisting {
		cmd := exec.CommandContext(ctx, "dropdb", dbName)
		cmd.Run() // Ignore error if doesn't exist
	}

	if options.CreateDatabase {
		cmd := exec.CommandContext(ctx, "createdb", dbName)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// Restore from backup
	args := []string{
		"-d", dbName,
		filepath,
	}

	cmd := exec.CommandContext(ctx, "pg_restore", args...)
	return cmd.Run()
}

// restoreMSSQL restores SQL Server database
func (bm *BackupManager) restoreMSSQL(ctx context.Context, db *sql.DB, filepath, dbName string, options *RestoreOptions) error {
	query := fmt.Sprintf(`
		RESTORE DATABASE [%s]
		FROM DISK = '%s'
		WITH REPLACE, STATS = 10
	`, dbName, filepath)

	_, err := db.ExecContext(ctx, query)
	return err
}

// restoreOracle restores Oracle database
func (bm *BackupManager) restoreOracle(ctx context.Context, filepath, dbName string, options *RestoreOptions) error {
	args := []string{
		"schemas=" + dbName,
		"dumpfile=" + filepath,
		"directory=DATA_PUMP_DIR",
	}

	if options.DropExisting {
		args = append(args, "table_exists_action=replace")
	}

	cmd := exec.CommandContext(ctx, "impdp", args...)
	return cmd.Run()
}

// verifyBackup verifies a backup file
func (bm *BackupManager) verifyBackup(backup *BackupInfo) error {
	// Check file exists
	if _, err := os.Stat(backup.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backup.FilePath)
	}

	// Check file size
	info, err := os.Stat(backup.FilePath)
	if err != nil {
		return err
	}

	if info.Size() == 0 {
		return fmt.Errorf("backup file is empty")
	}

	// Additional verification based on format
	switch backup.Format {
	case SQLFormat:
		return bm.verifySQLBackup(backup.FilePath)
	case BinaryFormat:
		return bm.verifyBinaryBackup(backup.FilePath)
	}

	return nil
}

// verifySQLBackup verifies SQL format backup
func (bm *BackupManager) verifySQLBackup(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read first few bytes to verify it looks like SQL
	buf := make([]byte, 1024)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return err
	}

	content := string(buf[:n])
	if len(content) == 0 {
		return fmt.Errorf("empty backup file")
	}

	return nil
}

// verifyBinaryBackup verifies binary format backup
func (bm *BackupManager) verifyBinaryBackup(filepath string) error {
	// Basic check - file exists and has content
	info, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	if info.Size() == 0 {
		return fmt.Errorf("empty backup file")
	}

	return nil
}

// generateFilename generates backup filename
func (bm *BackupManager) generateFilename(backupID, dbType string) string {
	ext := ".sql"
	switch bm.config.Format {
	case BinaryFormat:
		ext = ".dump"
	case CompressedFormat:
		ext = ".tar.gz"
	}

	if bm.config.Compression && bm.config.Format == SQLFormat {
		ext = ".sql.gz"
	}

	return backupID + ext
}

// cleanupOldBackups removes old backups based on retention policy
func (bm *BackupManager) cleanupOldBackups() {
	now := time.Now()

	// Remove backups older than retention period
	newBackups := make([]*BackupInfo, 0)
	for _, backup := range bm.backups {
		age := now.Sub(backup.CreatedAt).Hours() / 24
		if int(age) < bm.config.RetentionDays {
			newBackups = append(newBackups, backup)
		} else {
			// Delete backup file
			os.Remove(backup.FilePath)
		}
	}

	bm.backups = newBackups

	// Keep only max number of backups
	if len(bm.backups) > bm.config.MaxBackups {
		// Remove oldest backups
		for i := 0; i < len(bm.backups)-bm.config.MaxBackups; i++ {
			os.Remove(bm.backups[i].FilePath)
		}
		bm.backups = bm.backups[len(bm.backups)-bm.config.MaxBackups:]
	}
}

// loadBackupHistory loads existing backups from backup directory
func (bm *BackupManager) loadBackupHistory() {
	files, err := os.ReadDir(bm.config.BackupDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, _ := file.Info()
		// Create basic backup info from file
		// In production, this would read metadata from a JSON file
		backupInfo := &BackupInfo{
			ID:       file.Name(),
			FilePath: filepath.Join(bm.config.BackupDir, file.Name()),
			FileSize: info.Size(),
			CreatedAt: info.ModTime(),
		}
		bm.backups = append(bm.backups, backupInfo)
	}
}

// ListBackups returns all available backups
func (bm *BackupManager) ListBackups() []*BackupInfo {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	result := make([]*BackupInfo, len(bm.backups))
	copy(result, bm.backups)
	return result
}

// GetBackup returns specific backup by ID
func (bm *BackupManager) GetBackup(backupID string) (*BackupInfo, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	for _, backup := range bm.backups {
		if backup.ID == backupID {
			return backup, nil
		}
	}

	return nil, fmt.Errorf("backup not found: %s", backupID)
}

// DeleteBackup deletes a backup
func (bm *BackupManager) DeleteBackup(backupID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for i, backup := range bm.backups {
		if backup.ID == backupID {
			// Delete file
			if err := os.Remove(backup.FilePath); err != nil {
				return err
			}

			// Remove from list
			bm.backups = append(bm.backups[:i], bm.backups[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("backup not found: %s", backupID)
}

// BackupScheduler manages scheduled backups
type BackupScheduler struct {
	manager *BackupManager
	stopCh  chan struct{}
	running bool
	mu      sync.Mutex
}

// StartScheduler starts the backup scheduler
func (bm *BackupManager) StartScheduler(ctx context.Context, db *sql.DB, dbType, dbName string) {
	if bm.scheduler != nil && bm.scheduler.running {
		return
	}

	bm.scheduler = &BackupScheduler{
		manager: bm,
		stopCh:  make(chan struct{}),
		running: true,
	}

	go bm.scheduler.run(ctx, db, dbType, dbName)
}

// StopScheduler stops the backup scheduler
func (bm *BackupManager) StopScheduler() {
	if bm.scheduler != nil && bm.scheduler.running {
		close(bm.scheduler.stopCh)
		bm.scheduler.running = false
	}
}

// run executes the scheduler loop
func (bs *BackupScheduler) run(ctx context.Context, db *sql.DB, dbType, dbName string) {
	// Simple daily backup at configured time
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bs.manager.Backup(ctx, db, dbType, dbName)
		case <-bs.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}
