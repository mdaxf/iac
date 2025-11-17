# IAC Database Administration Tool (dbadmin)

A comprehensive command-line tool for managing IAC databases across multiple database types.

## Features

- **Connection Testing** - Verify database connectivity and credentials
- **Health Checks** - Monitor database health status
- **Schema Discovery** - Explore database schemas, tables, and columns
- **Migration Management** - Execute database schema migrations
- **Backup & Restore** - Database backup and restore operations
- **Performance Metrics** - View database performance statistics
- **Multi-Database Support** - Works with MySQL, PostgreSQL, MSSQL, and Oracle

## Installation

### Build from Source

```bash
cd cmd/dbadmin
go build -o dbadmin
```

### Install Globally

```bash
go install github.com/mdaxf/iac/cmd/dbadmin@latest
```

### Using Makefile

```bash
make build-dbadmin
```

## Quick Start

### 1. Test Database Connection

```bash
# MySQL
./dbadmin connect --type mysql --host localhost --port 3306 \
  --database iac --username iac_user --password iac_pass

# PostgreSQL
./dbadmin connect --type postgres --host localhost --port 5432 \
  --database iac --username iac_user --password iac_pass \
  --ssl-mode disable

# MSSQL
./dbadmin connect --type mssql --host localhost --port 1433 \
  --database iac --username sa --password YourPassword

# Oracle
./dbadmin connect --type oracle --host localhost --port 1521 \
  --database iac --username iac_user --password iac_pass
```

### 2. Check Database Health

```bash
# Check all configured databases
./dbadmin health

# Verbose output
./dbadmin health --verbose
```

### 3. Discover Database Schema

```bash
# Discover all tables and columns
./dbadmin schema discover --type mysql --host localhost \
  --database iac --username iac_user --password iac_pass

# List all databases
./dbadmin schema list --type mysql --host localhost \
  --username root --password root_pass
```

### 4. View Metrics

```bash
# Display database metrics
./dbadmin metrics

# JSON output
./dbadmin metrics --format json
```

### 5. List Configured Databases

```bash
# List all databases from environment
./dbadmin list

# Detailed information
./dbadmin list --verbose
```

## Commands

### connect

Test connection to a database and verify credentials.

**Usage:**
```bash
dbadmin connect [flags]
```

**Flags:**
- `-t, --type string` - Database type (mysql, postgres, mssql, oracle)
- `-H, --host string` - Database host (default: localhost)
- `-p, --port int` - Database port (default: 3306)
- `-d, --database string` - Database name (required)
- `-u, --username string` - Database username (required)
- `-P, --password string` - Database password
- `--ssl-mode string` - SSL mode (disable, require, verify-ca, verify-full)

**Example:**
```bash
./dbadmin connect -t postgres -H localhost -p 5432 \
  -d mydb -u myuser -P mypass --ssl-mode disable
```

**Output:**
```
Connecting to postgres database at localhost:5432...
✓ Connected successfully in 45.2ms
✓ Ping successful in 12.1ms

Database Information:
  Type:     postgres
  Dialect:  postgres
  Host:     localhost:5432
  Database: mydb

Supported Features:
  ✓ transactions
  ✓ jsonb
  ✓ cte
  ✓ window_functions
  ✗ fulltext
  ✓ arrays
```

### health

Check health status of all configured databases.

**Usage:**
```bash
dbadmin health [flags]
```

**Flags:**
- `-v, --verbose` - Show detailed information
- `-j, --json` - Output in JSON format

**Example:**
```bash
./dbadmin health --verbose
```

**Output:**
```
Checking health of 3 database(s)...

✓ mysql: Healthy (ping: 15.3ms)
    Dialect: mysql
✓ postgres: Healthy (ping: 18.7ms)
    Dialect: postgres
✗ oracle: Unhealthy - connection refused

Summary:
  Healthy:   2
  Unhealthy: 1
  Total:     3
```

### schema

Schema discovery and management operations.

#### schema discover

Discover all tables and columns in a database.

**Usage:**
```bash
dbadmin schema discover [flags]
```

**Flags:**
- `-t, --type string` - Database type
- `-H, --host string` - Database host
- `-p, --port int` - Database port
- `-d, --database string` - Database name (required)
- `-s, --schema string` - Schema name (default: same as database)
- `-u, --username string` - Database username (required)
- `-P, --password string` - Database password

**Example:**
```bash
./dbadmin schema discover -t postgres -H localhost \
  -d iac -u iac_user -P iac_pass -s public
```

**Output:**
```
Discovering schema for postgres database...

Tables in schema 'public':

1. users - User accounts table
   - id (integer) NOT NULL [PRIMARY KEY]
   - username (varchar) NOT NULL
   - email (varchar) NOT NULL
   - created_at (timestamp) NOT NULL

2. sessions - User session tracking
   - id (integer) NOT NULL [PRIMARY KEY]
   - user_id (integer) NOT NULL
   - token (varchar) NOT NULL
   - expires_at (timestamp) NOT NULL

Total tables: 2
```

#### schema list

List all available databases or schemas.

**Usage:**
```bash
dbadmin schema list [flags]
```

**Example:**
```bash
./dbadmin schema list -t mysql -H localhost -u root -P root_pass
```

### migrate

Execute database schema migrations.

**Usage:**
```bash
dbadmin migrate [flags]
```

**Flags:**
- `-t, --type string` - Database type
- `-d, --direction string` - Migration direction (up/down)
- `-s, --steps int` - Number of steps (0 = all)
- `--dry-run` - Dry run (don't apply changes)

**Example:**
```bash
./dbadmin migrate -t mysql -d up -s 5
```

**Note:** Migration feature shows syntax and will be fully implemented in future versions.

### backup

Create a backup of the specified database.

**Usage:**
```bash
dbadmin backup [flags]
```

**Flags:**
- `-t, --type string` - Database type
- `-H, --host string` - Database host
- `-p, --port int` - Database port
- `-d, --database string` - Database name (required)
- `-u, --username string` - Database username (required)
- `-P, --password string` - Database password
- `-o, --output string` - Output file (auto-generated if not specified)
- `-c, --compress` - Compress backup

**Example:**
```bash
./dbadmin backup -t mysql -H localhost -d iac \
  -u iac_user -P iac_pass -o backup.sql -c
```

### restore

Restore a database from backup.

**Usage:**
```bash
dbadmin restore [flags]
```

**Flags:**
- `-t, --type string` - Database type
- `-H, --host string` - Database host
- `-p, --port int` - Database port
- `-d, --database string` - Database name (required)
- `-u, --username string` - Database username (required)
- `-P, --password string` - Database password
- `-i, --input string` - Input file (required)
- `-c, --compressed` - Input is compressed

**Example:**
```bash
./dbadmin restore -t mysql -H localhost -d iac \
  -u iac_user -P iac_pass -i backup.sql.gz -c
```

### metrics

Display performance metrics for all configured databases.

**Usage:**
```bash
dbadmin metrics [flags]
```

**Flags:**
- `-f, --format string` - Output format (text, json)
- `-w, --watch` - Watch mode (update every second)

**Example:**
```bash
./dbadmin metrics --format json
```

**Output:**
```
Database Metrics
================

Database: mysql
  Dialect:  mysql
  Status:   Healthy

  Connection Pool:
    Active:     5
    Idle:       10
    Max:        15

  Query Statistics:
    Total:      1234
    Errors:     0
    Slow:       5
    Avg Time:   45.30ms
```

### list

List all databases configured in the environment.

**Usage:**
```bash
dbadmin list [flags]
```

**Flags:**
- `-v, --verbose` - Show detailed information

**Example:**
```bash
./dbadmin list --verbose
```

**Output:**
```
Configured Databases (3):

1. mysql
   Status:  Healthy
   Dialect: mysql
   Replicas: 2

2. postgres
   Status:  Healthy
   Dialect: postgres

3. mongodb
   Status:  Healthy
   Dialect: mongodb
   Replicas: 1
```

## Environment Variables

The `health`, `metrics`, and `list` commands use databases configured via environment variables:

```bash
# Primary database
export DB_TYPE=mysql
export DB_HOST=localhost
export DB_PORT=3306
export DB_DATABASE=iac
export DB_USERNAME=iac_user
export DB_PASSWORD=iac_pass

# Replicas (optional)
export DB_REPLICA_HOSTS=replica1:3306,replica2:3306

# Document database (optional)
export DOCDB_TYPE=mongodb
export DOCDB_HOST=localhost
export DOCDB_PORT=27017
export DOCDB_DATABASE=iac_docs
export DOCDB_USERNAME=iac_user
export DOCDB_PASSWORD=iac_pass
```

## Common Workflows

### Database Health Monitoring

```bash
# Quick health check
./dbadmin health

# Detailed health information
./dbadmin health --verbose

# Continuous monitoring (manually)
while true; do clear; ./dbadmin health; sleep 5; done
```

### Schema Exploration

```bash
# List all databases
./dbadmin schema list -t mysql -H localhost -u root -P root_pass

# Discover schema for specific database
./dbadmin schema discover -t mysql -H localhost \
  -d iac -u iac_user -P iac_pass

# Discover PostgreSQL schema
./dbadmin schema discover -t postgres -H localhost \
  -d iac -u iac_user -P iac_pass -s public
```

### Backup Workflow

```bash
# 1. Create compressed backup
./dbadmin backup -t mysql -H localhost -d iac \
  -u iac_user -P iac_pass -o iac_backup.sql -c

# 2. Verify backup file exists
ls -lh iac_backup.sql.gz

# 3. Restore to test database (if needed)
./dbadmin restore -t mysql -H localhost -d iac_test \
  -u iac_user -P iac_pass -i iac_backup.sql.gz -c
```

### Connection Testing

```bash
# Test all database types
for db in mysql postgres mssql oracle; do
  echo "Testing $db..."
  ./dbadmin connect -t $db -H localhost \
    -d testdb -u testuser -P testpass || echo "Failed: $db"
done
```

## Troubleshooting

### Connection Refused

**Problem:** `connection refused` error

**Solutions:**
- Verify database is running
- Check host and port are correct
- Verify firewall allows connections

### Authentication Failed

**Problem:** `Access denied` or `authentication failed`

**Solutions:**
- Verify username and password
- Check user has necessary permissions
- Verify user can connect from your host

### SSL/TLS Errors

**Problem:** `SSL connection error`

**Solutions:**
- For local development, use `--ssl-mode disable`
- For production, configure proper SSL certificates
- Verify SSL certificate validity

### No Databases Configured

**Problem:** `No databases configured` message

**Solutions:**
- Set environment variables (DB_TYPE, DB_HOST, etc.)
- Verify environment variables are exported
- Check syntax of environment variables

## Development

### Building

```bash
# Build for current platform
go build -o dbadmin

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o dbadmin-linux
GOOS=windows GOARCH=amd64 go build -o dbadmin.exe
GOOS=darwin GOARCH=amd64 go build -o dbadmin-mac
```

### Testing

```bash
# Run tests
go test ./...

# Test specific command
go run main.go connect --help
```

## License

Copyright 2023 IAC. All Rights Reserved.

Licensed under the Apache License, Version 2.0.
