# Database Backup & Restore

**Version:** 1.0
**Status:** ✅ Production Ready
**Component:** Phase 5 - Advanced Features

## Overview

The Backup & Restore system provides automated database backup and recovery capabilities for all supported database types with verification, scheduling, and retention management.

## Features

- **Multi-Database Support**: MySQL, PostgreSQL, SQL Server, Oracle
- **Multiple Backup Types**: Full, incremental, differential
- **Flexible Formats**: SQL, binary, compressed
- **Automatic Verification**: Validates backups after creation
- **Retention Management**: Automatic cleanup of old backups
- **Scheduled Backups**: Cron-style scheduling
- **Point-in-Time Recovery**: Restore to specific time (where supported)
- **Compression**: Optional gzip compression

## Quick Start

### Basic Backup

```go
package main

import (
    "context"
    "database/sql"
    "github.com/mdaxf/iac/databases"
)

func main() {
    // Create backup manager
    bm, err := databases.NewBackupManager(nil)
    if err != nil {
        panic(err)
    }

    // Connect to database
    db, _ := sql.Open("mysql", "user:pass@tcp(localhost:3306)/mydb")
    defer db.Close()

    // Create backup
    ctx := context.Background()
    backup, err := bm.Backup(ctx, db, "mysql", "mydb")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Backup created: %s\n", backup.ID)
}
```

### Basic Restore

```go
// Restore from backup
options := &databases.RestoreOptions{
    DropExisting:   true,
    CreateDatabase: false,
}

err := bm.Restore(ctx, db, backup.ID, options)
if err != nil {
    panic(err)
}
```

## Configuration

### BackupConfig

```go
config := &databases.BackupConfig{
    // Backup directory
    BackupDir: "/var/backups/db",

    // Backup format
    Format: databases.SQLFormat,

    // Enable compression
    Compression: true,

    // Maximum backups to retain
    MaxBackups: 30,

    // Days to keep backups
    RetentionDays: 90,

    // Verify backup after creation
    VerifyBackup: true,

    // Cron expression for scheduled backups
    ScheduleExpression: "0 2 * * *", // 2 AM daily
}

bm, err := databases.NewBackupManager(config)
```

### Default Configuration

```go
config := databases.DefaultBackupConfig()

// Defaults:
// - BackupDir: "./backups"
// - Format: SQLFormat
// - Compression: true
// - MaxBackups: 10
// - RetentionDays: 30
// - VerifyBackup: true
// - ScheduleExpression: "0 2 * * *"
```

## Backup Types

### Full Backup

Complete database backup:

```go
backup, err := bm.Backup(ctx, db, "mysql", "production_db")
// Creates complete backup of entire database
```

### Incremental Backup (Future)

Only changes since last backup:

```go
// TODO: Incremental backup support
```

### Differential Backup (Future)

Changes since last full backup:

```go
// TODO: Differential backup support
```

## Backup Formats

### SQL Format

Text-based SQL statements:

```go
config.Format = databases.SQLFormat
config.Compression = true // Creates .sql.gz

// MySQL: Uses mysqldump
// PostgreSQL: Uses pg_dump --format=plain
```

### Binary Format

Database-specific binary format:

```go
config.Format = databases.BinaryFormat

// MySQL: Uses mysqlpump
// PostgreSQL: Uses pg_dump --format=custom
// SQL Server: Uses native BACKUP DATABASE
```

### Compressed Format

Compressed archive:

```go
config.Format = databases.CompressedFormat

// Creates .tar.gz archive
```

## Database-Specific Features

### MySQL

```go
// Backup uses mysqldump
backup, err := bm.Backup(ctx, db, "mysql", "mydb")

// Features:
// - Single transaction for consistency
// - Quick mode for large tables
// - No table locks
// - Optional compression

// Restore uses mysql client
err = bm.Restore(ctx, db, backup.ID, options)
```

### PostgreSQL

```go
// Backup uses pg_dump
backup, err := bm.Backup(ctx, db, "postgres", "mydb")

// Features:
// - Custom format (-Fc)
// - Built-in compression
// - Parallel dump (future)
// - Schema-only or data-only options (future)

// Restore uses pg_restore
err = bm.Restore(ctx, db, backup.ID, options)
```

### SQL Server

```go
// Backup uses T-SQL BACKUP DATABASE
backup, err := bm.Backup(ctx, db, "mssql", "mydb")

// Features:
// - Native compression
// - Differential backups (future)
// - Transaction log backups (future)

// Restore uses T-SQL RESTORE DATABASE
err = bm.Restore(ctx, db, backup.ID, options)
```

### Oracle

```go
// Backup uses Data Pump (expdp)
backup, err := bm.Backup(ctx, db, "oracle", "mydb")

// Features:
// - Schema-level export
// - Parallel processing
// - Compression

// Restore uses Data Pump (impdp)
err = bm.Restore(ctx, db, backup.ID, options)
```

## API Reference

### Core Methods

#### NewBackupManager

```go
func NewBackupManager(config *BackupConfig) (*BackupManager, error)
```

Creates a new backup manager. Automatically creates backup directory if it doesn't exist.

#### Backup

```go
func (bm *BackupManager) Backup(
    ctx context.Context,
    db *sql.DB,
    dbType, dbName string,
) (*BackupInfo, error)
```

Creates a database backup and returns backup information.

#### Restore

```go
func (bm *BackupManager) Restore(
    ctx context.Context,
    db *sql.DB,
    backupID string,
    options *RestoreOptions,
) error
```

Restores database from backup.

### Management Methods

#### ListBackups

```go
func (bm *BackupManager) ListBackups() []*BackupInfo
```

Returns all available backups.

#### GetBackup

```go
func (bm *BackupManager) GetBackup(backupID string) (*BackupInfo, error)
```

Returns specific backup by ID.

#### DeleteBackup

```go
func (bm *BackupManager) DeleteBackup(backupID string) error
```

Deletes a backup file and removes from history.

### Scheduling Methods

#### StartScheduler

```go
func (bm *BackupManager) StartScheduler(
    ctx context.Context,
    db *sql.DB,
    dbType, dbName string,
)
```

Starts automated backup scheduler.

#### StopScheduler

```go
func (bm *BackupManager) StopScheduler()
```

Stops backup scheduler.

## Restore Options

### RestoreOptions

```go
type RestoreOptions struct {
    // Drop existing database before restore
    DropExisting bool

    // Create database if it doesn't exist
    CreateDatabase bool

    // Restore to specific point in time
    PointInTime *time.Time

    // Skip backup verification
    SkipVerification bool

    // Target database name (if different)
    TargetDatabase string
}
```

### Usage Examples

```go
// Drop and replace database
options := &databases.RestoreOptions{
    DropExisting: true,
}

// Restore to different database
options := &databases.RestoreOptions{
    CreateDatabase: true,
    TargetDatabase: "mydb_restored",
}

// Point-in-time recovery
targetTime := time.Now().Add(-24 * time.Hour)
options := &databases.RestoreOptions{
    PointInTime: &targetTime,
}
```

## Backup Information

### BackupInfo

```go
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
```

## Best Practices

### 1. Regular Backups

```go
// Daily backups at low-traffic time
config.ScheduleExpression = "0 2 * * *" // 2 AM

// Weekly full backups
config.ScheduleExpression = "0 2 * * 0" // Sunday 2 AM

// Hourly during business hours
config.ScheduleExpression = "0 9-17 * * 1-5" // 9 AM-5 PM weekdays
```

### 2. Retention Policy

```go
// Production: Long retention
config.MaxBackups = 60
config.RetentionDays = 180

// Development: Short retention
config.MaxBackups = 7
config.RetentionDays = 14

// Critical data: Very long retention
config.RetentionDays = 365
```

### 3. Verification

```go
// Always verify in production
config.VerifyBackup = true

// Can skip in development for speed
config.VerifyBackup = false
```

### 4. Compression

```go
// Enable for large databases
config.Compression = true

// May skip for small databases or fast networks
config.Compression = false
```

### 5. Off-Site Storage

```go
// After backup, copy to off-site location
backup, err := bm.Backup(ctx, db, "mysql", "mydb")
if err == nil {
    copyToS3(backup.FilePath)
    // or
    copyToGCS(backup.FilePath)
    // or
    copyToAzure(backup.FilePath)
}
```

## Backup Strategies

### Strategy 1: Full Daily Backups

```go
config := &databases.BackupConfig{
    BackupDir:          "/backups/daily",
    ScheduleExpression: "0 2 * * *",
    MaxBackups:         30,
    RetentionDays:      90,
}

// Pros: Simple, complete backups
// Cons: Storage intensive, slower
```

### Strategy 2: Weekly Full + Daily Incremental (Future)

```go
// Sunday: Full backup
// Monday-Saturday: Incremental

// Pros: Less storage, faster daily backups
// Cons: Complex restore process
```

### Strategy 3: Monthly Full + Weekly Differential (Future)

```go
// 1st of month: Full backup
// Weekly: Differential

// Pros: Balanced storage and speed
// Cons: Moderate complexity
```

## Monitoring

### Track Backup Status

```go
// List recent backups
backups := bm.ListBackups()
for _, backup := range backups {
    fmt.Printf("Backup: %s\n", backup.ID)
    fmt.Printf("  Created: %s\n", backup.CreatedAt)
    fmt.Printf("  Size: %d bytes\n", backup.FileSize)
    fmt.Printf("  Verified: %v\n", backup.Verified)
}
```

### Alert on Failures

```go
backup, err := bm.Backup(ctx, db, "mysql", "production")
if err != nil {
    // Send alert
    sendAlert("Backup failed: " + err.Error())
    logError(err)
}

if !backup.Verified {
    sendAlert("Backup verification failed")
}
```

### Monitor Disk Usage

```go
backups := bm.ListBackups()
var totalSize int64
for _, b := range backups {
    totalSize += b.FileSize
}

if totalSize > maxAllowedSize {
    sendAlert("Backup storage exceeds limit")
}
```

## Disaster Recovery

### Recovery Procedure

1. **Identify Backup**
   ```go
   backups := bm.ListBackups()
   // Select appropriate backup
   ```

2. **Verify Backup**
   ```go
   backup, _ := bm.GetBackup(backupID)
   if !backup.Verified {
       err := bm.verifyBackup(backup)
   }
   ```

3. **Prepare Target**
   ```go
   options := &databases.RestoreOptions{
       DropExisting:   true,
       CreateDatabase: true,
   }
   ```

4. **Restore**
   ```go
   err := bm.Restore(ctx, db, backupID, options)
   ```

5. **Verify Data**
   ```go
   // Run consistency checks
   // Verify critical data
   ```

### Recovery Time Objective (RTO)

```go
// Minimize downtime
- Keep backups on fast storage
- Use parallel restore where available
- Test restore procedures regularly
```

### Recovery Point Objective (RPO)

```go
// Minimize data loss
- Frequent backups (hourly for critical data)
- Transaction log backups
- Point-in-time recovery
```

## Testing

### Test Backup Process

```go
// Regularly test backup creation
backup, err := bm.Backup(ctx, db, "mysql", "test_db")
if err != nil {
    t.Errorf("Backup failed: %v", err)
}

// Verify backup file exists
if _, err := os.Stat(backup.FilePath); os.IsNotExist(err) {
    t.Error("Backup file not created")
}
```

### Test Restore Process

```go
// Test restore to verify backups are usable
testDB := "test_restore_db"
options := &databases.RestoreOptions{
    CreateDatabase: true,
    TargetDatabase: testDB,
}

err := bm.Restore(ctx, db, backup.ID, options)
if err != nil {
    t.Errorf("Restore failed: %v", err)
}

// Verify data
// Clean up test database
```

## Security

### Encrypt Backups

```go
// After backup, encrypt file
backup, _ := bm.Backup(ctx, db, "mysql", "mydb")
encryptFile(backup.FilePath, encryptionKey)
```

### Secure Storage

```go
// Set restrictive permissions
os.Chmod(backup.FilePath, 0600) // Owner read/write only

// Store in encrypted filesystem
// Use encrypted cloud storage
```

### Access Control

```go
// Restrict backup directory access
os.Mkdir(config.BackupDir, 0700)

// Use separate backup user with minimal privileges
```

## Performance

### Backup Performance

```go
// Use parallel dump (PostgreSQL future)
// Adjust buffer sizes
// Use fast compression levels
```

### Restore Performance

```go
// Disable indexes during restore
// Bulk load data
// Use parallel restore
```

## Troubleshooting

### Backup Fails

```go
// Check disk space
// Verify database connectivity
// Check permissions
// Review error logs
```

### Restore Fails

```go
// Verify backup file integrity
// Check target database state
// Ensure sufficient disk space
// Review compatibility
```

### Large Backup Files

```go
// Increase compression
config.Compression = true

// Use binary format
config.Format = databases.BinaryFormat

// Implement incremental backups (future)
```

## Examples

Complete examples available in:
- `/examples/backup_restore_example.go` - Comprehensive usage examples
- `/databases/backup_manager_test.go` - Unit tests with examples

## License

Copyright 2023 IAC. All Rights Reserved.

Licensed under the Apache License, Version 2.0.

---

**Version:** 1.0
**Last Updated:** 2025-11-16
**Maintained By:** IAC Development Team
