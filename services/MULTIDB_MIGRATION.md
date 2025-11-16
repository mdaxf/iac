# Service Layer Multi-Database Migration Guide

This guide explains how to update the service layer to support multiple database types (MySQL, PostgreSQL, MSSQL, Oracle).

## Overview

The service layer has been enhanced to support multiple database types through:

1. **DatabaseHelper** - Provides dialect-aware operations and database selection
2. **Schema Query Builders** - Dialect-specific queries for schema discovery
3. **Service Factory** - Centralized service creation with multi-DB support
4. **SchemaMetadataServiceMultiDB** - Enhanced schema discovery for all database types

## Architecture

```
┌─────────────────────────────────────────┐
│         Service Factory                  │
│  (Creates services with multi-DB)       │
└──────────────┬──────────────────────────┘
               │
               ├──► DatabaseHelper ──────► DatabaseSelector ──────► PoolManager
               │                                                         │
               └──► Services                                            │
                    ├── BusinessEntityService (uses appDB)              │
                    ├── QueryTemplateService (uses appDB)               │
                    ├── SchemaMetadataServiceMultiDB ◄──────────────────┘
                    │   (uses both appDB and user databases)
                    └── Other services...
```

## Key Components

### 1. DatabaseHelper (`services/dbhelper.go`)

Provides multi-database support for services:

```go
type DatabaseHelper struct {
    selector *databases.DatabaseSelector
    gormDB   *gorm.DB  // IAC application's own database
}
```

**Features:**
- Get user database connections via selector
- Dialect-aware query execution
- Helper functions for common database operations

**Example Usage:**
```go
// Get user database for reading
db, err := dbHelper.GetUserDB(ctx, "customer_db")

// Get user database for writing
db, err := dbHelper.GetUserDBForWrite(ctx, "customer_db")

// Execute dialect-aware query
rows, err := dbHelper.ExecuteDialectQuery(ctx, db, func(dialect string) string {
    switch dialect {
    case "mysql":
        return "SELECT * FROM users LIMIT 10"
    case "postgres":
        return "SELECT * FROM users LIMIT 10"
    case "mssql":
        return "SELECT TOP 10 * FROM users"
    case "oracle":
        return "SELECT * FROM users FETCH FIRST 10 ROWS ONLY"
    default:
        return "SELECT * FROM users LIMIT 10"
    }
})
```

### 2. Dialect Helper Functions (`services/dbhelper.go`)

Utility functions for dialect-specific SQL:

```go
// Get current timestamp expression
expr := CurrentTimestampExpr(dialect)  // "CURRENT_TIMESTAMP", "GETDATE()", "SYSTIMESTAMP"

// Get LIKE operator (case-insensitive)
op := LikeOperator(dialect, false)  // "LIKE", "ILIKE"

// Build pagination clause
clause := LimitOffsetClause(dialect, 10, 20)  // "LIMIT 10 OFFSET 20", "OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY"

// String concatenation
concat := StringConcatExpr(dialect, "'Hello'", "' '", "'World'")  // "CONCAT(...)", "... + ...", "... || ..."

// JSON extraction
json := JSONExtractExpr(dialect, "data", "name")  // "data->>'name'", "JSON_VALUE(data, '$.name')"
```

### 3. Schema Discovery Queries (`services/schema_queries.go`)

Dialect-specific queries for schema discovery:

```go
// Get tables in a schema
query := GetTablesQuery(dialect, schemaName)

// Get columns for a table
query := GetColumnsQuery(dialect, schemaName, tableName)

// Get indexes for a table
query := GetIndexesQuery(dialect, schemaName, tableName)

// List all databases/schemas
query := GetDatabaseListQuery(dialect)

// Normalize data types across databases
normalizedType := NormalizeDataType(dialect, "VARCHAR2")  // Returns "varchar"
```

### 4. SchemaMetadataServiceMultiDB (`services/schemametadataservice_multidb.go`)

Enhanced service for multi-database schema discovery:

```go
// Create service
svc := NewSchemaMetadataServiceMultiDB(dbHelper, appDB)

// Discover schema (works with all database types)
err := svc.DiscoverSchema(ctx, "customer_db", "public")

// Discover indexes
indexes, err := svc.DiscoverIndexes(ctx, "customer_db", "public", "users")

// Execute custom query
rows, err := svc.ExecuteQuery(ctx, "customer_db", "SELECT * FROM users WHERE id = ?", 123)
```

## Migration Steps

### Step 1: Initialize Service Factory

In your application initialization (`main.go` or similar):

```go
import (
    "github.com/mdaxf/iac/databases"
    "github.com/mdaxf/iac/dbinitializer"
    "github.com/mdaxf/iac/services"
)

func initializeServices() (*services.ServiceFactory, error) {
    // Initialize database layer from environment
    dbInit := dbinitializer.NewDatabaseInitializer()
    if err := dbInit.InitializeFromEnvironment(); err != nil {
        return nil, fmt.Errorf("failed to initialize databases: %w", err)
    }

    // Get pool manager
    poolManager := dbInit.GetPoolManager()

    // Get application database (GORM instance)
    appDB := dbInit.GetGORMDB()

    // Create service factory
    serviceFactory, err := services.NewServiceFactory(poolManager, appDB)
    if err != nil {
        return nil, fmt.Errorf("failed to create service factory: %w", err)
    }

    return serviceFactory, nil
}
```

### Step 2: Use Services

```go
// Get services from factory
businessEntitySvc := serviceFactory.GetBusinessEntityService()
schemaMetadataSvc := serviceFactory.GetSchemaMetadataServiceMultiDB()

// Use services normally
entities, err := businessEntitySvc.ListEntities(ctx, "customer_db")

// Discover schema for any database type
err = schemaMetadataSvc.DiscoverSchema(ctx, "customer_db", "public")
```

### Step 3: Update Existing Services (If Needed)

For services that need to query user databases directly:

**Before:**
```go
type MyService struct {
    db *gorm.DB
}

func (s *MyService) QueryUserDB(ctx context.Context, dbName string) error {
    // Hardcoded MySQL query
    rows, err := s.db.Raw("SELECT * FROM information_schema.TABLES WHERE TABLE_SCHEMA = ?", dbName).Rows()
    // ...
}
```

**After:**
```go
type MyService struct {
    dbHelper *services.DatabaseHelper
    appDB    *gorm.DB
}

func (s *MyService) QueryUserDB(ctx context.Context, databaseAlias string, schemaName string) error {
    // Get user database
    db, err := s.dbHelper.GetUserDB(ctx, databaseAlias)
    if err != nil {
        return err
    }
    defer db.Close()

    // Use dialect-aware query
    dialect := db.GetDialect()
    query := services.GetTablesQuery(dialect, schemaName)
    rows, err := db.Query(query)
    // ...
}
```

## Best Practices

### 1. Separation of Concerns

- **IAC App Database (appDB)**: Use GORM directly for IAC's own tables (business entities, templates, metadata)
- **User Databases**: Use DatabaseHelper and dialect-aware queries

### 2. Always Use Dialect-Aware Queries

**Don't:**
```go
// Hardcoded MySQL query
query := "SELECT * FROM users LIMIT 10"
```

**Do:**
```go
// Dialect-aware query
dialect := db.GetDialect()
query := fmt.Sprintf("SELECT * FROM users %s", services.LimitOffsetClause(dialect, 10, 0))
```

### 3. Use Schema Query Builders

**Don't:**
```go
// Hardcoded information_schema query (MySQL specific)
query := "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = ?"
```

**Do:**
```go
// Use the schema query builder
dialect := db.GetDialect()
query := services.GetTablesQuery(dialect, schemaName)
```

### 4. Close Connections

Always close database connections obtained from DatabaseHelper:

```go
db, err := dbHelper.GetUserDB(ctx, databaseAlias)
if err != nil {
    return err
}
defer db.Close()  // Important!

// Use db...
```

### 5. Use Service Factory

Always create services through the ServiceFactory:

**Don't:**
```go
// Direct instantiation
svc := services.NewBusinessEntityService(appDB)
```

**Do:**
```go
// Use factory
svc := serviceFactory.GetBusinessEntityService()
```

## Common Patterns

### Pattern 1: Schema Discovery

```go
// Discover all tables and columns in a PostgreSQL database
svc := serviceFactory.GetSchemaMetadataServiceMultiDB()
err := svc.DiscoverSchema(ctx, "postgres_db", "public")

// Discover all tables and columns in a MySQL database
err = svc.DiscoverSchema(ctx, "mysql_db", "myschema")

// Discover all tables and columns in an Oracle database
err = svc.DiscoverSchema(ctx, "oracle_db", "MYUSER")
```

### Pattern 2: Dynamic Query Execution

```go
func executeQuery(ctx context.Context, dbHelper *services.DatabaseHelper, databaseAlias string, baseQuery string, limit int) error {
    db, err := dbHelper.GetUserDB(ctx, databaseAlias)
    if err != nil {
        return err
    }
    defer db.Close()

    dialect := db.GetDialect()

    // Build dialect-aware query
    fullQuery := fmt.Sprintf("%s %s", baseQuery, services.LimitOffsetClause(dialect, limit, 0))

    rows, err := db.Query(fullQuery)
    if err != nil {
        return err
    }
    defer rows.Close()

    // Process rows...
    return nil
}
```

### Pattern 3: Type Normalization

```go
// Get column metadata
columnMeta, err := svc.GetColumnMetadata(ctx, databaseAlias, tableName)

// Normalize data types for display
for _, col := range columnMeta {
    // col.DataType is already normalized (e.g., VARCHAR2 -> varchar, NUMBER -> integer)
    fmt.Printf("Column: %s, Type: %s\n", col.Column, col.DataType)
}
```

## Testing

When writing tests for services with multi-database support:

```go
func TestServiceWithMultiDB(t *testing.T) {
    // Setup test databases
    poolManager := setupTestPoolManager(t)
    appDB := setupTestGORMDB(t)

    // Create service factory
    factory, err := services.NewServiceFactory(poolManager, appDB)
    require.NoError(t, err)

    // Get service
    svc := factory.GetSchemaMetadataServiceMultiDB()

    // Test with different database types
    t.Run("MySQL", func(t *testing.T) {
        err := svc.DiscoverSchema(ctx, "test_mysql", "testdb")
        assert.NoError(t, err)
    })

    t.Run("PostgreSQL", func(t *testing.T) {
        err := svc.DiscoverSchema(ctx, "test_postgres", "public")
        assert.NoError(t, err)
    })
}
```

## Backward Compatibility

The original services (e.g., `SchemaMetadataService`) remain unchanged for backward compatibility. New code should use the enhanced multi-DB versions:

- `SchemaMetadataService` → Use `SchemaMetadataServiceMultiDB`
- Direct `*gorm.DB` usage → Use `DatabaseHelper`

## Configuration

Ensure your environment variables are set up correctly:

```bash
# Relational databases
DB_TYPE=mysql
DB_HOST=localhost
DB_PORT=3306
DB_DATABASE=iac
DB_USERNAME=iac_user
DB_PASSWORD=iac_pass

# Replicas (optional)
DB_REPLICA_HOSTS=replica1:3306,replica2:3306

# Document databases (optional)
DOCDB_TYPE=mongodb
DOCDB_HOST=localhost
DOCDB_PORT=27017
```

## Troubleshooting

### Issue: "Database connection failed"

**Solution:** Verify database configuration in environment variables and ensure the database is running.

### Issue: "Unsupported dialect"

**Solution:** Ensure the database type is one of: mysql, postgres, mssql, oracle

### Issue: "information_schema query failed on Oracle"

**Solution:** Make sure you're using `SchemaMetadataServiceMultiDB` instead of the legacy `SchemaMetadataService`

### Issue: "SQL syntax error"

**Solution:** Check that you're using dialect-aware queries. Use the helper functions in `dbhelper.go` and `schema_queries.go`

## Summary

The multi-database service layer provides:

✅ Support for MySQL, PostgreSQL, MSSQL, and Oracle
✅ Dialect-aware SQL query generation
✅ Centralized service creation via ServiceFactory
✅ Backward compatibility with existing services
✅ Comprehensive schema discovery for all database types
✅ Connection pooling and selection strategies
✅ Helper functions for common database operations

For more information, see:
- `/databases/` - Core database abstraction layer
- `/dbinitializer/` - Database initialization
- `/services/` - Service layer implementation
