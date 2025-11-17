# Database Schema Management

This document explains how to use the database-agnostic schema management features to create and modify tables across different database types (MySQL, PostgreSQL, MSSQL, Oracle).

## Overview

The schema management system uses the dialect pattern to automatically generate appropriate DDL for your target database. You define your schema once, and it works across all supported databases.

## Quick Start

### 1. Define Your Table Schema

```go
package main

import (
    dbconn "github.com/mdaxf/iac/databases"
)

func defineUserSchema() *dbconn.TableSchema {
    maxLen100 := 100
    maxLen255 := 255
    defaultActive := "1"

    return &dbconn.TableSchema{
        TableName: "users",
        Columns: []dbconn.ColumnInfo{
            {
                Name:       "id",
                DataType:   "bigint",
                IsNullable: false,
                IsPrimaryKey: true,
                Comment:    "User ID (auto-increment)",
            },
            {
                Name:         "username",
                DataType:     "string",
                MaxLength:    &maxLen100,
                IsNullable:   false,
                IsUnique:     true,
                Comment:      "Unique username",
            },
            {
                Name:         "email",
                DataType:     "string",
                MaxLength:    &maxLen255,
                IsNullable:   false,
                Comment:      "User email address",
            },
            {
                Name:         "is_active",
                DataType:     "bool",
                IsNullable:   false,
                DefaultValue: &defaultActive,
                Comment:      "Whether user account is active",
            },
            {
                Name:       "created_at",
                DataType:   "datetime",
                IsNullable: false,
                Comment:    "Account creation timestamp",
            },
        },
        PrimaryKeys: []string{"id"},
        Indexes: []dbconn.IndexInfo{
            {
                Name:     "idx_username",
                Columns:  []string{"username"},
                IsUnique: true,
            },
            {
                Name:     "idx_email",
                Columns:  []string{"email"},
                IsUnique: false,
            },
        },
    }
}
```

### 2. Create the Table

```go
func createUserTable(db *sql.DB) error {
    ctx := context.Background()

    // Create schema manager
    schemaManager, err := dbconn.NewSchemaManager(db, "admin")
    if err != nil {
        return fmt.Errorf("failed to create schema manager: %w", err)
    }

    // Define schema
    schema := defineUserSchema()

    // Create table - automatically uses correct DDL for your database type
    err = schemaManager.CreateTable(ctx, schema)
    if err != nil {
        return fmt.Errorf("failed to create table: %w", err)
    }

    fmt.Println("Table created successfully!")
    return nil
}
```

### 3. Add a Column Dynamically

```go
func addColumnExample(db *sql.DB) error {
    ctx := context.Background()
    schemaManager, _ := dbconn.NewSchemaManager(db, "admin")

    maxLen50 := 50
    newColumn := &dbconn.ColumnInfo{
        Name:       "phone_number",
        DataType:   "string",
        MaxLength:  &maxLen50,
        IsNullable: true,
        Comment:    "User phone number",
    }

    return schemaManager.AddColumn(ctx, "users", newColumn)
}
```

### 4. Create an Index

```go
func addIndexExample(db *sql.DB) error {
    ctx := context.Background()
    schemaManager, _ := dbconn.NewSchemaManager(db, "admin")

    index := &dbconn.IndexInfo{
        Name:     "idx_created_at",
        Columns:  []string{"created_at"},
        IsUnique: false,
    }

    return schemaManager.CreateIndex(ctx, "users", index)
}
```

## Generated DDL Examples

The same schema definition generates appropriate DDL for each database type:

### MySQL
```sql
CREATE TABLE `users` (
  `id` BIGINT NOT NULL COMMENT 'User ID (auto-increment)',
  `username` VARCHAR(100) NOT NULL COMMENT 'Unique username',
  `email` VARCHAR(255) NOT NULL COMMENT 'User email address',
  `is_active` BOOLEAN NOT NULL DEFAULT 1 COMMENT 'Whether user account is active',
  `created_at` DATETIME NOT NULL COMMENT 'Account creation timestamp',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```

### PostgreSQL
```sql
CREATE TABLE "users" (
  "id" BIGINT NOT NULL,
  "username" VARCHAR(100) NOT NULL,
  "email" VARCHAR(255) NOT NULL,
  "is_active" BOOLEAN NOT NULL DEFAULT 1,
  "created_at" TIMESTAMP NOT NULL,
  PRIMARY KEY ("id")
)
```

### MSSQL
```sql
CREATE TABLE [users] (
  [id] BIGINT NOT NULL,
  [username] NVARCHAR(100) NOT NULL,
  [email] NVARCHAR(255) NOT NULL,
  [is_active] BIT NOT NULL DEFAULT 1,
  [created_at] DATETIME2 NOT NULL,
  PRIMARY KEY ([id])
)
```

### Oracle
```sql
CREATE TABLE "users" (
  "id" NUMBER(19) NOT NULL,
  "username" VARCHAR2(100) NOT NULL,
  "email" VARCHAR2(255) NOT NULL,
  "is_active" NUMBER(1) NOT NULL DEFAULT 1,
  "created_at" TIMESTAMP NOT NULL,
  PRIMARY KEY ("id")
)
```

## Supported Data Types

Generic types are automatically mapped to database-specific types:

| Generic Type | MySQL | PostgreSQL | MSSQL | Oracle |
|-------------|-------|------------|-------|--------|
| string | VARCHAR | VARCHAR | NVARCHAR | VARCHAR2 |
| text | TEXT | TEXT | NVARCHAR(MAX) | CLOB |
| int | INT | INTEGER | INT | NUMBER(10) |
| bigint | BIGINT | BIGINT | BIGINT | NUMBER(19) |
| float | FLOAT | REAL | FLOAT | BINARY_FLOAT |
| double | DOUBLE | DOUBLE PRECISION | FLOAT(53) | BINARY_DOUBLE |
| bool | BOOLEAN | BOOLEAN | BIT | NUMBER(1) |
| datetime | DATETIME | TIMESTAMP | DATETIME2 | TIMESTAMP |
| json | JSON | JSONB | NVARCHAR(MAX) | CLOB |
| blob | BLOB | BYTEA | VARBINARY(MAX) | BLOB |

## Schema Manager API Reference

### Core Methods

- **CreateTable(ctx, schema)** - Creates a new table with indexes
- **AddColumn(ctx, tableName, column)** - Adds a column to existing table
- **DropColumn(ctx, tableName, columnName)** - Removes a column
- **AlterColumn(ctx, tableName, column)** - Modifies column definition
- **CreateIndex(ctx, tableName, index)** - Creates an index
- **DropIndex(ctx, tableName, indexName)** - Removes an index
- **DropTable(ctx, tableName)** - Drops a table
- **TableExists(ctx, tableName)** - Checks if table exists
- **MigrateSchema(ctx, schema)** - Creates table or migrates existing schema

## Best Practices

1. **Always use generic data types** - Let the dialect system handle database-specific mapping
2. **Define constraints in schema** - Primary keys, uniqueness, nullability
3. **Use context for timeouts** - Pass context with appropriate timeout for DDL operations
4. **Check for existence** - Use TableExists() before creating tables to avoid errors
5. **Transaction support** - Wrap multiple schema changes in a transaction for atomicity

## Integration with Application Code

```go
// In your initialization code
func InitializeDatabase() error {
    // Connect to database (uses existing connection logic)
    db := dbconn.DB

    // Create schema manager
    schemaManager, err := dbconn.NewSchemaManager(db, "system")
    if err != nil {
        return err
    }

    // Define all your application schemas
    schemas := []*dbconn.TableSchema{
        defineUserSchema(),
        defineProductSchema(),
        defineOrderSchema(),
        // ... more schemas
    }

    // Create or migrate each table
    ctx := context.Background()
    for _, schema := range schemas {
        if err := schemaManager.MigrateSchema(ctx, schema); err != nil {
            return fmt.Errorf("failed to migrate %s: %w", schema.TableName, err)
        }
    }

    return nil
}
```

## Migration Strategy

For existing applications, you can gradually migrate to schema definitions:

1. **Document existing tables** as TableSchema structs
2. **Use MigrateSchema()** which only creates if table doesn't exist
3. **Plan future changes** as schema modifications in code
4. **Version control** your schema definitions

This approach provides database portability and eliminates manual DDL management while maintaining backwards compatibility.
