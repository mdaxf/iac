# Database Initialization Troubleshooting

## Issue: "database not initialized" Error

If you see this error:
```
[E] System CollectionService Failed to get document database: database not initialized
[E] GetListofCollectionData failed to retrieve the list from collection: database not initialized
```

This means both the new and legacy database systems are not properly initialized.

## Quick Fix - Restart Required

The most common cause is that the application is running with old code. After updating the code:

### 1. Stop the Application
```bash
# Stop your running application
pkill -f iac  # or however you stop your app
```

### 2. Rebuild the Application
```bash
# Rebuild to include the new initialization code
go build -o iac .
```

### 3. Restart the Application
```bash
# Start with the new build
./iac
```

## What to Look For in Logs

After restart, you should see these log messages indicating successful initialization:

### Successful Initialization:
```
[I] Initializing legacy document database: type=mongodb, connection=mongodb://..., database=IAC_CFG
[I] Legacy document database connection initialized successfully
[I] New document database system initialized successfully
=== Database Configuration ===
Document Databases: 1
  - primary (mongodb): Connected=true
=============================
```

### Collection Query (should see one of these):
```
[D] QueryCollection: Using legacy database mode
[D] Using legacy database mode for collection query
```
OR
```
[D] QueryCollection: Using new database interface mode
```

## Verification Steps

### 1. Check if Database Config is Present

Verify your configuration file has the document database settings:
```json
{
  "documentdb": {
    "type": "mongodb",
    "connection": "mongodb://localhost:27017",
    "database": "IAC_CFG"
  }
}
```

### 2. Test MongoDB Connection

Verify MongoDB is running and accessible:
```bash
# Test MongoDB connection
mongo mongodb://localhost:27017/IAC_CFG

# Or with mongosh (newer versions)
mongosh mongodb://localhost:27017/IAC_CFG
```

### 3. Check Logs for Initialization

Look for these log messages in the startup sequence:
```bash
# Grep for database initialization logs
grep -i "initialize document" /path/to/logs

# Should show:
# [I] Initializing legacy document database: ...
# [I] Legacy document database connection initialized successfully
# [I] New document database system initialized successfully
```

## Common Issues and Solutions

### Issue 1: MongoDB Not Running
**Error**: Connection timeout or refused
**Solution**:
```bash
# Start MongoDB
sudo systemctl start mongod

# Or with Docker
docker start mongodb
```

### Issue 2: Wrong MongoDB Connection String
**Error**: Connection failed
**Solution**: Check configuration file has correct connection string
```json
{
  "documentdb": {
    "connection": "mongodb://localhost:27017"  // ← Check this
  }
}
```

### Issue 3: Database Name Missing
**Error**: "DatabaseName is missing"
**Solution**: Ensure database name is in config:
```json
{
  "documentdb": {
    "database": "IAC_CFG"  // ← Must be present
  }
}
```

### Issue 4: Old Binary Running
**Error**: Still seeing "database not initialized" after code update
**Solution**:
1. Kill old process: `pkill -f iac`
2. Rebuild: `go build -o iac .`
3. Start new: `./iac`

## Debug Mode

To enable more detailed logging, set log level to DEBUG in your config:
```json
{
  "log": {
    "level": "DEBUG"
  }
}
```

Then restart and check logs for detailed initialization sequence.

## Manual Verification

You can verify database connectivity manually:

### Test Legacy Connection:
```go
// In your code, after startup
if documents.DocDBCon != nil {
    log.Println("Legacy DB: Connected")
} else {
    log.Println("Legacy DB: NOT connected")
}
```

### Test New Connection:
```go
// In your code, after startup
if dbinitializer.GlobalInitializer != nil {
    db, err := dbinitializer.GlobalInitializer.GetDocumentDB()
    if err == nil && db != nil {
        log.Println("New DB: Connected")
    } else {
        log.Printf("New DB: Error - %v", err)
    }
} else {
    log.Println("New DB: GlobalInitializer is nil")
}
```

## Still Not Working?

If you've tried all the above and still see errors:

1. **Check application startup logs** for any errors during initialization
2. **Verify MongoDB is accessible** from the application server
3. **Check firewall rules** if MongoDB is on a different server
4. **Verify credentials** if using authentication
5. **Check the initialize.go** file is being executed during startup

## Expected Behavior

After successful initialization, the service will:
1. ✅ Try new database interface first (if GlobalInitializer exists)
2. ✅ Fall back to legacy mode (if DocDBCon exists)
3. ✅ Return "database not initialized" only if BOTH are nil

## Need More Help?

Enable debug logging and capture the full startup sequence:
```bash
# Redirect logs to file
./iac 2>&1 | tee startup.log

# Search for initialization
grep -A 5 "initialize Document" startup.log
```

Then review the initialization sequence to see where it's failing.
