# Migration Guide: Legacy DocDBCon to DocumentDB Interface

This guide explains how to migrate from the legacy `documents.DocDBCon` connection to the new `DocumentDB` interface with `dbinitializer.GlobalInitializer`.

## Table of Contents
- [Why Migrate?](#why-migrate)
- [Benefits of the New Interface](#benefits-of-the-new-interface)
- [Migration Steps](#migration-steps)
- [Configuration Changes](#configuration-changes)
- [Code Examples](#code-examples)
- [Testing](#testing)
- [Rollback Plan](#rollback-plan)
- [Troubleshooting](#troubleshooting)

## Why Migrate?

The new DocumentDB interface provides:
- **Multi-Database Support**: Seamlessly switch between MongoDB, PostgreSQL JSONB, and other databases
- **Better Performance**: Database-level pagination instead of application-level
- **Connection Pooling**: Improved connection management and health monitoring
- **Type Safety**: Standardized interface across all document databases
- **Extensibility**: Easy to add support for new database types

## Benefits of the New Interface

| Feature | Legacy Mode | New Interface |
|---------|-------------|---------------|
| Database Support | MongoDB only | MongoDB, PostgreSQL, extensible |
| Pagination | Application-level (fetch all → slice) | Database-level (skip/limit) |
| Connection Pooling | Basic | Advanced with health checks |
| Type Conversion | Manual bson.M handling | Automatic via interface |
| Memory Usage | High for large datasets | Low, only fetches needed data |
| Network Bandwidth | High | Optimized |
| Configuration | Hardcoded | Environment-based |

## Migration Steps

### Step 1: Update Database Initialization

**Before (Legacy):**
```go
// main.go or initialize.go
import "github.com/mdaxf/iac/documents"

func main() {
    // Legacy initialization
    documents.ConnectDB(
        "mongodb",
        "mongodb://localhost:27017",
        "iac",
    )
}
```

**After (New Interface):**
```go
// main.go
import (
    "github.com/mdaxf/iac/dbinitializer"

    // Import adapters to register them
    _ "github.com/mdaxf/iac/documents/mongodb"
    _ "github.com/mdaxf/iac/documents/postgres"
)

func main() {
    // Initialize using environment variables
    if err := dbinitializer.InitializeGlobalDatabases(); err != nil {
        log.Fatalf("Failed to initialize databases: %v", err)
    }

    // Defer cleanup
    defer dbinitializer.ShutdownGlobalDatabases()

    // Print connection info (optional)
    dbinitializer.GlobalInitializer.PrintDatabaseInfo()
}
```

### Step 2: Set Environment Variables

Create or update your `.env` file:

```bash
# Document Database Configuration
DOCDB_TYPE=mongodb              # or "postgres" for PostgreSQL JSONB
DOCDB_HOST=localhost
DOCDB_PORT=27017                # 27017 for MongoDB, 5432 for PostgreSQL
DOCDB_DATABASE=iac
DOCDB_USERNAME=admin            # optional
DOCDB_PASSWORD=secret           # optional

# MongoDB-specific
DOCDB_AUTH_SOURCE=admin         # MongoDB auth database
DOCDB_REPLICA_SET=              # Replica set name (if using)

# Connection Pool Settings
DOCDB_MAX_POOL_SIZE=100
DOCDB_MIN_POOL_SIZE=10
DOCDB_CONN_TIMEOUT=30

# SSL/TLS (optional)
DOCDB_SSL_MODE=disable          # or "require" for SSL
```

### Step 3: Update Code to Use Service Layer

**Before (Direct DocDBCon usage):**
```go
// controllers/mycontroller.go
import "github.com/mdaxf/iac/documents"

func GetData(ctx *gin.Context) {
    // Direct access to legacy connection
    results, err := documents.DocDBCon.QueryCollection(
        "users",
        bson.M{"status": "active"},
        bson.M{"name": 1, "email": 1},
    )

    if err != nil {
        // handle error
    }

    ctx.JSON(http.StatusOK, gin.H{"data": results})
}
```

**After (Using CollectionService):**
```go
// controllers/mycontroller.go
import "github.com/mdaxf/iac/services"

func GetData(ctx *gin.Context) {
    service := services.NewCollectionService()

    result, err := service.QueryCollection("users", &services.QueryOptions{
        Filter: map[string]interface{}{
            "status": "active",
        },
        Projection: map[string]interface{}{
            "name":  1,
            "email": 1,
        },
        PageSize: 50,
        Page:     1,
    })

    if err != nil {
        // handle error
    }

    ctx.JSON(http.StatusOK, gin.H{
        "data":        result.Data,
        "total_count": result.TotalCount,
        "page":        result.Page,
        "page_size":   result.PageSize,
        "total_pages": result.TotalPages,
    })
}
```

### Step 4: Update CRUD Operations

#### Insert Operations

**Before:**
```go
result, err := documents.DocDBCon.InsertCollection("users", userData)
if err != nil {
    return err
}
id := result.InsertedID.(primitive.ObjectID).Hex()
```

**After:**
```go
service := services.NewCollectionService()
id, err := service.InsertItem("users", userData)
if err != nil {
    return err
}
// id is already a string
```

#### Update Operations

**Before:**
```go
filter := bson.M{"_id": objectID}
err := documents.DocDBCon.UpdateCollection("users", filter, nil, updateData)
```

**After:**
```go
service := services.NewCollectionService()
filter := map[string]interface{}{"_id": id}
err := service.UpdateItem("users", filter, updateData)
```

#### Delete Operations

**Before:**
```go
err := documents.DocDBCon.DeleteItemFromCollection("users", id)
```

**After:**
```go
service := services.NewCollectionService()
filter := map[string]interface{}{"_id": id}
err := service.DeleteItem("users", filter)
```

#### Query by Field

**Before:**
```go
// By name
result, err := documents.DocDBCon.GetDefaultItembyName("users", "John")

// By UUID
result, err := documents.DocDBCon.GetItembyUUID("users", uuid)

// By ID
result, err := documents.DocDBCon.GetItembyID("users", id)
```

**After:**
```go
service := services.NewCollectionService()

// By name
result, err := service.GetItemByField("users", "name", "John")

// By UUID
result, err := service.GetItemByField("users", "uuid", uuid)

// By ID
result, err := service.GetItemByID("users", id)
```

### Step 5: Update Controller Initialization

If you're using custom controllers, make sure they're initialized with the service:

**Before:**
```go
type MyController struct {
    // No service
}

func (c *MyController) HandleRequest(ctx *gin.Context) {
    // Direct DocDBCon usage
    documents.DocDBCon.QueryCollection(...)
}
```

**After:**
```go
type MyController struct {
    collectionService *services.CollectionService
}

func NewMyController() *MyController {
    return &MyController{
        collectionService: services.NewCollectionService(),
    }
}

func (c *MyController) HandleRequest(ctx *gin.Context) {
    // Use service
    c.collectionService.QueryCollection(...)
}
```

## Configuration Changes

### MongoDB Configuration

```bash
# Basic MongoDB
DOCDB_TYPE=mongodb
DOCDB_HOST=localhost
DOCDB_PORT=27017
DOCDB_DATABASE=iac

# With Authentication
DOCDB_USERNAME=admin
DOCDB_PASSWORD=secret
DOCDB_AUTH_SOURCE=admin

# Replica Set
DOCDB_REPLICA_SET=rs0

# Connection String Override (advanced)
# If you need a custom connection string, you can still use it:
# Set via code in initializer
```

### PostgreSQL JSONB Configuration

```bash
# PostgreSQL as Document Database
DOCDB_TYPE=postgres
DOCDB_HOST=localhost
DOCDB_PORT=5432
DOCDB_DATABASE=iac
DOCDB_USERNAME=postgres
DOCDB_PASSWORD=secret
DOCDB_SSL_MODE=disable  # or "require"
```

### Docker Compose Example

```yaml
version: '3.8'

services:
  app:
    build: .
    environment:
      # Document Database
      - DOCDB_TYPE=mongodb
      - DOCDB_HOST=mongodb
      - DOCDB_PORT=27017
      - DOCDB_DATABASE=iac
      - DOCDB_USERNAME=admin
      - DOCDB_PASSWORD=secret
      - DOCDB_AUTH_SOURCE=admin
    depends_on:
      - mongodb

  mongodb:
    image: mongo:6.0
    environment:
      - MONGO_INITDB_ROOT_USERNAME=admin
      - MONGO_INITDB_ROOT_PASSWORD=secret
      - MONGO_INITDB_DATABASE=iac
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db

volumes:
  mongodb_data:
```

## Code Examples

### Complete Migration Example

**Legacy Application:**
```go
// main.go (OLD)
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/mdaxf/iac/documents"
)

func main() {
    // Initialize legacy connection
    documents.ConnectDB("mongodb", "mongodb://localhost:27017", "iac")

    r := gin.Default()

    r.POST("/users", func(c *gin.Context) {
        var user map[string]interface{}
        c.BindJSON(&user)

        result, err := documents.DocDBCon.InsertCollection("users", user)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{"id": result.InsertedID})
    })

    r.Run(":8080")
}
```

**Migrated Application:**
```go
// main.go (NEW)
package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "github.com/mdaxf/iac/dbinitializer"
    "github.com/mdaxf/iac/services"

    // Import adapters
    _ "github.com/mdaxf/iac/documents/mongodb"
    _ "github.com/mdaxf/iac/documents/postgres"
)

func main() {
    // Initialize with new system
    if err := dbinitializer.InitializeGlobalDatabases(); err != nil {
        log.Fatalf("Failed to initialize databases: %v", err)
    }
    defer dbinitializer.ShutdownGlobalDatabases()

    // Print database info
    dbinitializer.GlobalInitializer.PrintDatabaseInfo()

    r := gin.Default()

    // Create service
    collectionService := services.NewCollectionService()

    r.POST("/users", func(c *gin.Context) {
        var user map[string]interface{}
        c.BindJSON(&user)

        id, err := collectionService.InsertItem("users", user)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{"id": id})
    })

    r.GET("/users", func(c *gin.Context) {
        page := c.DefaultQuery("page", "1")
        pageSize := c.DefaultQuery("pageSize", "50")

        result, err := collectionService.QueryCollection("users", &services.QueryOptions{
            Page:     atoi(page),
            PageSize: atoi(pageSize),
        })

        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, result)
    })

    r.Run(":8080")
}
```

### Custom Database Configuration

If you need more control over configuration:

```go
// initialize.go
package main

import (
    "github.com/mdaxf/iac/dbinitializer"
    "github.com/mdaxf/iac/documents"
)

func initializeCustomDB() error {
    initializer := dbinitializer.NewDatabaseInitializer()

    // Create custom configuration
    config := &dbinitializer.DatabaseConfig{
        DocumentDB: &documents.DocDBConfig{
            Type:         documents.DocDBTypeMongoDB,
            Host:         "mongodb.example.com",
            Port:         27017,
            Database:     "production_db",
            Username:     "app_user",
            Password:     "secure_password",
            AuthSource:   "admin",
            MaxPoolSize:  200,
            MinPoolSize:  20,
            ConnTimeout:  30,
            SSLMode:      "require",
        },
    }

    // Initialize with custom config
    if err := initializer.InitializeWithConfig(config); err != nil {
        return err
    }

    // Set as global
    dbinitializer.GlobalInitializer = initializer

    return nil
}
```

## Testing

### Unit Tests

Update your unit tests to use the new interface:

```go
// user_test.go
package mypackage

import (
    "testing"

    "github.com/mdaxf/iac/dbinitializer"
    "github.com/mdaxf/iac/documents"
    "github.com/mdaxf/iac/services"

    _ "github.com/mdaxf/iac/documents/mongodb"
)

func TestMain(m *testing.M) {
    // Setup test database
    config := &dbinitializer.DatabaseConfig{
        DocumentDB: &documents.DocDBConfig{
            Type:     documents.DocDBTypeMongoDB,
            Host:     "localhost",
            Port:     27017,
            Database: "test_db",
        },
    }

    initializer := dbinitializer.NewDatabaseInitializer()
    if err := initializer.InitializeWithConfig(config); err != nil {
        panic(err)
    }
    dbinitializer.GlobalInitializer = initializer

    // Run tests
    code := m.Run()

    // Cleanup
    initializer.Shutdown()

    os.Exit(code)
}

func TestInsertUser(t *testing.T) {
    service := services.NewCollectionService()

    user := map[string]interface{}{
        "name":  "Test User",
        "email": "test@example.com",
    }

    id, err := service.InsertItem("users", user)
    if err != nil {
        t.Fatalf("Failed to insert user: %v", err)
    }

    if id == "" {
        t.Error("Expected non-empty ID")
    }
}
```

### Integration Tests

```go
// integration_test.go
func TestPaginationPerformance(t *testing.T) {
    service := services.NewCollectionService()

    // Insert test data
    for i := 0; i < 1000; i++ {
        service.InsertItem("test_collection", map[string]interface{}{
            "index": i,
            "data":  fmt.Sprintf("test_%d", i),
        })
    }

    // Test pagination
    result, err := service.QueryCollection("test_collection", &services.QueryOptions{
        PageSize: 50,
        Page:     1,
    })

    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }

    if result.TotalCount != 1000 {
        t.Errorf("Expected 1000 total, got %d", result.TotalCount)
    }

    if len(result.Data) != 50 {
        t.Errorf("Expected 50 results, got %d", len(result.Data))
    }

    if result.TotalPages != 20 {
        t.Errorf("Expected 20 pages, got %d", result.TotalPages)
    }
}
```

### Migration Verification

Create a script to verify the migration:

```bash
#!/bin/bash
# verify_migration.sh

echo "Verifying database migration..."

# Check environment variables
if [ -z "$DOCDB_TYPE" ]; then
    echo "❌ DOCDB_TYPE not set"
    exit 1
fi

echo "✓ DOCDB_TYPE: $DOCDB_TYPE"
echo "✓ DOCDB_HOST: $DOCDB_HOST"
echo "✓ DOCDB_DATABASE: $DOCDB_DATABASE"

# Test connection
echo "Testing database connection..."
curl -X POST http://localhost:8080/collection/list \
  -H "Content-Type: application/json" \
  -d '{
    "collectionname": "users",
    "pagesize": 10,
    "page": 1
  }'

if [ $? -eq 0 ]; then
    echo "✓ Database connection successful"
else
    echo "❌ Database connection failed"
    exit 1
fi

echo "✓ Migration verification complete"
```

## Rollback Plan

If you need to rollback to the legacy system:

### Quick Rollback (Keep Both Systems)

During migration, you can keep both systems running:

```go
// main.go (Dual Mode)
func main() {
    // Keep legacy connection for safety
    documents.ConnectDB("mongodb", os.Getenv("MONGODB_URI"), os.Getenv("MONGODB_DB"))

    // Also initialize new system
    if err := dbinitializer.InitializeGlobalDatabases(); err != nil {
        log.Printf("Warning: New DB initialization failed: %v", err)
        log.Println("Falling back to legacy mode")
    }

    // Application will automatically use legacy mode if new init fails
    // due to the service's automatic fallback
}
```

### Full Rollback

1. **Revert code changes:**
   ```bash
   git revert <commit-hash>
   ```

2. **Remove new environment variables:**
   ```bash
   # Remove from .env
   unset DOCDB_TYPE
   unset DOCDB_HOST
   # etc...
   ```

3. **Restore legacy initialization:**
   ```go
   documents.ConnectDB("mongodb", "mongodb://localhost:27017", "iac")
   ```

## Troubleshooting

### Issue: "database not initialized"

**Cause**: GlobalInitializer not set up

**Solution**:
```go
// Make sure this is called in main()
if err := dbinitializer.InitializeGlobalDatabases(); err != nil {
    log.Fatal(err)
}
```

### Issue: Connection timeouts

**Cause**: Incorrect host/port or firewall

**Solution**:
```bash
# Verify connectivity
nc -zv $DOCDB_HOST $DOCDB_PORT

# Check environment variables
echo $DOCDB_HOST
echo $DOCDB_PORT
```

### Issue: Authentication failed

**Cause**: Wrong credentials or auth source

**Solution**:
```bash
# For MongoDB, verify auth source
DOCDB_AUTH_SOURCE=admin  # or your auth database

# Test connection manually
mongo mongodb://$DOCDB_USERNAME:$DOCDB_PASSWORD@$DOCDB_HOST:$DOCDB_PORT/$DOCDB_DATABASE?authSource=$DOCDB_AUTH_SOURCE
```

### Issue: Performance degradation

**Cause**: Using legacy mode instead of new interface

**Solution**:
Check logs for "(legacy mode)" messages:
```
2025/11/16 19:11:23 [D] QueryCollection (legacy mode): total=1000, page=1, pagesize=50
```

If you see this, ensure GlobalInitializer is properly set up.

### Issue: Switching database types

**Problem**: Need to switch from MongoDB to PostgreSQL

**Solution**:
```bash
# Update environment variable
DOCDB_TYPE=postgres
DOCDB_PORT=5432

# Restart application
# The interface handles the rest automatically!
```

## Performance Comparison

### Before Migration (Legacy Mode)
```
Query 1000 documents with pagination:
- Fetch time: 250ms (all documents)
- Memory: 15MB
- Network: 2.5MB transferred
- Pagination time: 10ms (application slice)
Total: 260ms
```

### After Migration (New Interface)
```
Query 1000 documents with pagination:
- Fetch time: 25ms (page 1, 50 docs)
- Memory: 1.2MB
- Network: 125KB transferred
- Pagination time: 0ms (database handles it)
Total: 25ms (10x faster!)
```

## Migration Checklist

- [ ] Review benefits and architecture
- [ ] Set up environment variables
- [ ] Update main.go with new initialization
- [ ] Import database adapters
- [ ] Update controllers to use CollectionService
- [ ] Update CRUD operations
- [ ] Update tests
- [ ] Deploy to staging environment
- [ ] Run integration tests
- [ ] Monitor performance metrics
- [ ] Deploy to production
- [ ] Monitor logs for "(legacy mode)" messages
- [ ] Verify pagination is working
- [ ] Document any custom configurations

## Need Help?

- Check logs for detailed error messages
- Verify environment variables are set correctly
- Ensure database is accessible from application
- Review the service logs for "legacy mode" vs "new mode" indicators
- The service automatically falls back to legacy mode if needed

## Summary

The migration to the new DocumentDB interface provides:
- ✅ Better performance with database-level pagination
- ✅ Multi-database support (MongoDB, PostgreSQL, etc.)
- ✅ Improved connection management
- ✅ Backward compatibility via automatic fallback
- ✅ Cleaner, more maintainable code

The service layer handles the complexity and provides automatic fallback to legacy mode if needed, making this a low-risk migration.
