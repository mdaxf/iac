// Copyright 2023 IAC. All Rights Reserved.

package dbconn

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewBackupManager(t *testing.T) {
	config := DefaultBackupConfig()
	config.BackupDir = filepath.Join(os.TempDir(), "test_backups")
	defer os.RemoveAll(config.BackupDir)

	bm, err := NewBackupManager(config)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	if bm == nil {
		t.Fatal("BackupManager is nil")
	}

	// Verify backup directory was created
	if _, err := os.Stat(config.BackupDir); os.IsNotExist(err) {
		t.Error("Backup directory was not created")
	}
}

func TestBackupManager_GenerateFilename(t *testing.T) {
	config := DefaultBackupConfig()
	config.Format = SQLFormat
	config.Compression = true

	bm, _ := NewBackupManager(config)

	filename := bm.generateFilename("test_backup_001", "mysql")
	if filename != "test_backup_001.sql.gz" {
		t.Errorf("Expected test_backup_001.sql.gz, got %s", filename)
	}

	// Test binary format
	config.Format = BinaryFormat
	config.Compression = false
	bm, _ = NewBackupManager(config)

	filename = bm.generateFilename("test_backup_002", "postgres")
	if filename != "test_backup_002.dump" {
		t.Errorf("Expected test_backup_002.dump, got %s", filename)
	}
}

func TestBackupManager_ListBackups(t *testing.T) {
	config := DefaultBackupConfig()
	config.BackupDir = filepath.Join(os.TempDir(), "test_backups_list")
	defer os.RemoveAll(config.BackupDir)

	bm, _ := NewBackupManager(config)

	// Initially should be empty
	backups := bm.ListBackups()
	if len(backups) != 0 {
		t.Errorf("Expected 0 backups, got %d", len(backups))
	}

	// Add some test backup info
	bm.mu.Lock()
	bm.backups = append(bm.backups, &BackupInfo{
		ID:           "backup1",
		DatabaseName: "testdb",
		CreatedAt:    time.Now(),
	})
	bm.backups = append(bm.backups, &BackupInfo{
		ID:           "backup2",
		DatabaseName: "testdb",
		CreatedAt:    time.Now(),
	})
	bm.mu.Unlock()

	backups = bm.ListBackups()
	if len(backups) != 2 {
		t.Errorf("Expected 2 backups, got %d", len(backups))
	}
}

func TestBackupManager_GetBackup(t *testing.T) {
	config := DefaultBackupConfig()
	config.BackupDir = filepath.Join(os.TempDir(), "test_backups_get")
	defer os.RemoveAll(config.BackupDir)

	bm, _ := NewBackupManager(config)

	// Add test backup
	testBackup := &BackupInfo{
		ID:           "backup123",
		DatabaseName: "testdb",
		CreatedAt:    time.Now(),
	}
	bm.mu.Lock()
	bm.backups = append(bm.backups, testBackup)
	bm.mu.Unlock()

	// Get existing backup
	backup, err := bm.GetBackup("backup123")
	if err != nil {
		t.Fatalf("GetBackup failed: %v", err)
	}

	if backup.ID != "backup123" {
		t.Errorf("Expected backup123, got %s", backup.ID)
	}

	// Get non-existent backup
	_, err = bm.GetBackup("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent backup")
	}
}

func TestBackupManager_DeleteBackup(t *testing.T) {
	config := DefaultBackupConfig()
	config.BackupDir = filepath.Join(os.TempDir(), "test_backups_delete")
	defer os.RemoveAll(config.BackupDir)

	bm, _ := NewBackupManager(config)

	// Create a test backup file
	testFile := filepath.Join(config.BackupDir, "test.sql")
	os.WriteFile(testFile, []byte("test backup"), 0644)

	// Add backup info
	bm.mu.Lock()
	bm.backups = append(bm.backups, &BackupInfo{
		ID:       "test_backup",
		FilePath: testFile,
	})
	bm.mu.Unlock()

	// Delete backup
	err := bm.DeleteBackup("test_backup")
	if err != nil {
		t.Fatalf("DeleteBackup failed: %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("Backup file should be deleted")
	}

	// Verify removed from list
	backups := bm.ListBackups()
	if len(backups) != 0 {
		t.Error("Backup should be removed from list")
	}
}

func TestBackupManager_CleanupOldBackups(t *testing.T) {
	config := DefaultBackupConfig()
	config.BackupDir = filepath.Join(os.TempDir(), "test_backups_cleanup")
	config.RetentionDays = 7
	config.MaxBackups = 3
	defer os.RemoveAll(config.BackupDir)

	bm, _ := NewBackupManager(config)

	// Add old backups
	for i := 0; i < 5; i++ {
		filename := filepath.Join(config.BackupDir, "backup"+string(rune('0'+i))+".sql")
		os.WriteFile(filename, []byte("test"), 0644)

		bm.mu.Lock()
		bm.backups = append(bm.backups, &BackupInfo{
			ID:        "backup" + string(rune('0'+i)),
			FilePath:  filename,
			CreatedAt: time.Now().Add(-time.Duration(10-i) * 24 * time.Hour),
		})
		bm.mu.Unlock()
	}

	// Cleanup
	bm.mu.Lock()
	bm.cleanupOldBackups()
	bm.mu.Unlock()

	// Should keep only MaxBackups (3)
	if len(bm.backups) > config.MaxBackups {
		t.Errorf("Expected max %d backups, got %d", config.MaxBackups, len(bm.backups))
	}
}

func TestBackupManager_VerifyBackup(t *testing.T) {
	config := DefaultBackupConfig()
	config.BackupDir = filepath.Join(os.TempDir(), "test_backups_verify")
	defer os.RemoveAll(config.BackupDir)

	bm, _ := NewBackupManager(config)

	// Create test SQL backup file
	testFile := filepath.Join(config.BackupDir, "test.sql")
	os.WriteFile(testFile, []byte("CREATE TABLE test (id INT);"), 0644)

	backup := &BackupInfo{
		ID:       "test",
		FilePath: testFile,
		Format:   SQLFormat,
	}

	// Verify backup
	err := bm.verifyBackup(backup)
	if err != nil {
		t.Errorf("Verification failed: %v", err)
	}

	// Test with non-existent file
	backup.FilePath = "/nonexistent/file.sql"
	err = bm.verifyBackup(backup)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test with empty file
	emptyFile := filepath.Join(config.BackupDir, "empty.sql")
	os.WriteFile(emptyFile, []byte(""), 0644)
	backup.FilePath = emptyFile
	err = bm.verifyBackup(backup)
	if err == nil {
		t.Error("Expected error for empty file")
	}
}

func TestBackupTypes(t *testing.T) {
	types := []BackupType{FullBackup, IncrementalBackup, DifferentialBackup}

	for _, bt := range types {
		if string(bt) == "" {
			t.Errorf("Backup type should not be empty: %v", bt)
		}
	}
}

func TestBackupFormats(t *testing.T) {
	formats := []BackupFormat{SQLFormat, BinaryFormat, CompressedFormat}

	for _, f := range formats {
		if string(f) == "" {
			t.Errorf("Backup format should not be empty: %v", f)
		}
	}
}

func TestDefaultBackupConfig(t *testing.T) {
	config := DefaultBackupConfig()

	if config.BackupDir == "" {
		t.Error("BackupDir should not be empty")
	}

	if config.Format == "" {
		t.Error("Format should not be empty")
	}

	if config.MaxBackups <= 0 {
		t.Error("MaxBackups should be positive")
	}

	if config.RetentionDays <= 0 {
		t.Error("RetentionDays should be positive")
	}
}

func TestRestoreOptions(t *testing.T) {
	options := &RestoreOptions{
		DropExisting:     true,
		CreateDatabase:   true,
		SkipVerification: false,
		TargetDatabase:   "restored_db",
	}

	if !options.DropExisting {
		t.Error("DropExisting should be true")
	}

	if options.TargetDatabase != "restored_db" {
		t.Errorf("Expected 'restored_db', got %s", options.TargetDatabase)
	}
}

func TestBackupInfo(t *testing.T) {
	info := &BackupInfo{
		ID:           "test123",
		DatabaseName: "mydb",
		DatabaseType: "mysql",
		BackupType:   FullBackup,
		Format:       SQLFormat,
		FilePath:     "/path/to/backup.sql",
		FileSize:     1024,
		CreatedAt:    time.Now(),
		Compressed:   true,
		Verified:     true,
		Metadata:     make(map[string]string),
	}

	if info.ID != "test123" {
		t.Error("ID not set correctly")
	}

	if info.DatabaseName != "mydb" {
		t.Error("DatabaseName not set correctly")
	}

	if !info.Verified {
		t.Error("Verified should be true")
	}

	// Test metadata
	info.Metadata["key"] = "value"
	if info.Metadata["key"] != "value" {
		t.Error("Metadata not working correctly")
	}
}

func TestBackupScheduler_StartStop(t *testing.T) {
	config := DefaultBackupConfig()
	config.BackupDir = filepath.Join(os.TempDir(), "test_backups_scheduler")
	defer os.RemoveAll(config.BackupDir)

	bm, _ := NewBackupManager(config)

	// Create test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Start scheduler
	bm.StartScheduler(ctx, db, "sqlite3", "test")

	if bm.scheduler == nil {
		t.Error("Scheduler should be initialized")
	}

	if !bm.scheduler.running {
		t.Error("Scheduler should be running")
	}

	// Stop scheduler
	bm.StopScheduler()

	// Give it a moment to stop
	time.Sleep(100 * time.Millisecond)

	if bm.scheduler.running {
		t.Error("Scheduler should be stopped")
	}
}

func TestBackupManager_LoadBackupHistory(t *testing.T) {
	config := DefaultBackupConfig()
	config.BackupDir = filepath.Join(os.TempDir(), "test_backups_history")
	defer os.RemoveAll(config.BackupDir)

	// Create backup directory
	os.MkdirAll(config.BackupDir, 0755)

	// Create some test backup files
	files := []string{"backup1.sql", "backup2.sql", "backup3.sql"}
	for _, f := range files {
		filepath := filepath.Join(config.BackupDir, f)
		os.WriteFile(filepath, []byte("test backup"), 0644)
	}

	// Create backup manager (should load existing backups)
	bm, _ := NewBackupManager(config)

	// Should have loaded existing backups
	backups := bm.ListBackups()
	if len(backups) != len(files) {
		t.Errorf("Expected %d backups loaded, got %d", len(files), len(backups))
	}
}

func TestBackupManager_Integration(t *testing.T) {
	config := DefaultBackupConfig()
	config.BackupDir = filepath.Join(os.TempDir(), "test_backups_integration")
	config.VerifyBackup = true
	defer os.RemoveAll(config.BackupDir)

	bm, err := NewBackupManager(config)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	// Initially no backups
	if len(bm.ListBackups()) != 0 {
		t.Error("Should start with no backups")
	}

	// Create a dummy backup file for testing
	backupID := "test_backup_integration"
	filename := bm.generateFilename(backupID, "mysql")
	filepath := filepath.Join(config.BackupDir, filename)
	os.WriteFile(filepath, []byte("-- MySQL dump\nCREATE TABLE test (id INT);"), 0644)

	info := &BackupInfo{
		ID:           backupID,
		DatabaseName: "testdb",
		DatabaseType: "mysql",
		BackupType:   FullBackup,
		Format:       SQLFormat,
		FilePath:     filepath,
		FileSize:     100,
		CreatedAt:    time.Now(),
		Verified:     false,
	}

	// Verify the backup
	err = bm.verifyBackup(info)
	if err != nil {
		t.Errorf("Backup verification failed: %v", err)
	}

	// Add to backups list
	bm.mu.Lock()
	bm.backups = append(bm.backups, info)
	bm.mu.Unlock()

	// Should have one backup
	backups := bm.ListBackups()
	if len(backups) != 1 {
		t.Errorf("Expected 1 backup, got %d", len(backups))
	}

	// Get the backup
	backup, err := bm.GetBackup(backupID)
	if err != nil {
		t.Errorf("Failed to get backup: %v", err)
	}

	if backup.ID != backupID {
		t.Errorf("Expected backup ID %s, got %s", backupID, backup.ID)
	}

	// Delete the backup
	err = bm.DeleteBackup(backupID)
	if err != nil {
		t.Errorf("Failed to delete backup: %v", err)
	}

	// Should have no backups
	if len(bm.ListBackups()) != 0 {
		t.Error("Should have no backups after deletion")
	}
}
