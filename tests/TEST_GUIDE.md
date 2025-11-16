# Database Integration Test Guide

This guide explains how to run integration tests for all supported database types in the IAC system.

## Overview

The IAC system includes comprehensive integration tests for:

**Relational Databases:**
- MySQL
- PostgreSQL
- Microsoft SQL Server (MSSQL)
- Oracle

**Document Databases:**
- MongoDB
- PostgreSQL JSONB

## Prerequisites

### 1. Database Servers

You need access to running instances of the databases you want to test. The easiest way is to use Docker:

```bash
# Start all databases using Docker Compose
cd /path/to/iac
docker-compose -f docker-compose.databases.yml up -d

# Or use the quick start script
./scripts/start-databases.sh
```

### 2. Environment Configuration

Configure test database connections using environment variables:

#### MySQL
```bash
export TEST_MYSQL_HOST=localhost
export TEST_MYSQL_PORT=3306
export TEST_MYSQL_DATABASE=iac
export TEST_MYSQL_USERNAME=iac_user
export TEST_MYSQL_PASSWORD=iac_pass
```

#### PostgreSQL
```bash
export TEST_POSTGRES_HOST=localhost
export TEST_POSTGRES_PORT=5432
export TEST_POSTGRES_DATABASE=iac
export TEST_POSTGRES_USERNAME=iac_user
export TEST_POSTGRES_PASSWORD=iac_pass
```

#### Microsoft SQL Server
```bash
export TEST_MSSQL_HOST=localhost
export TEST_MSSQL_PORT=1433
export TEST_MSSQL_DATABASE=iac
export TEST_MSSQL_USERNAME=sa
export TEST_MSSQL_PASSWORD=MsSql_Pass123!
```

#### Oracle
```bash
export TEST_ORACLE_HOST=localhost
export TEST_ORACLE_PORT=1521
export TEST_ORACLE_DATABASE=iac
export TEST_ORACLE_USERNAME=iac_user
export TEST_ORACLE_PASSWORD=iac_pass
```

#### MongoDB
```bash
export TEST_MONGODB_HOST=localhost
export TEST_MONGODB_PORT=27017
export TEST_MONGODB_DATABASE=iac_test
export TEST_MONGODB_USERNAME=iac_user
export TEST_MONGODB_PASSWORD=iac_pass
```

### 3. Skip Unavailable Databases

If a database is not available, you can skip its tests:

```bash
export SKIP_mysql_TESTS=true
export SKIP_postgres_TESTS=true
export SKIP_mssql_TESTS=true
export SKIP_oracle_TESTS=true
export SKIP_mongodb_TESTS=true
export SKIP_postgres_jsonb_TESTS=true
```

## Running Tests

### Run All Tests

```bash
# Run all database integration tests
go test -v ./databases/... ./documents/... -tags=integration
```

### Run Tests for Specific Database Package

```bash
# Run only relational database tests
go test -v ./databases/integration_test.go

# Run only document database tests
go test -v ./documents/integration_test.go
```

### Run Specific Test Cases

```bash
# Run only connection tests
go test -v ./databases/integration_test.go -run TestDatabaseConnection

# Run only CRUD operation tests
go test -v ./databases/integration_test.go -run TestDatabaseBasicOperations

# Run only transaction tests
go test -v ./databases/integration_test.go -run TestDatabaseTransactions

# Run document CRUD tests
go test -v ./documents/integration_test.go -run TestDocumentCRUDOperations
```

### Run Tests for Single Database Type

```bash
# Test MySQL only
go test -v ./databases/integration_test.go -run TestDatabaseConnection/mysql

# Test PostgreSQL only
go test -v ./databases/integration_test.go -run TestDatabaseConnection/postgres

# Test MongoDB only
go test -v ./documents/integration_test.go -run TestDocumentDatabaseConnection/mongodb
```

## Test Coverage

### Relational Database Tests (`databases/integration_test.go`)

#### 1. Connection Tests
- **TestDatabaseConnection**: Verifies basic connectivity to each database
  - Creates database instance
  - Tests connection
  - Tests ping
  - Verifies correct dialect

#### 2. CRUD Operation Tests
- **TestDatabaseBasicOperations**: Tests basic SQL operations
  - CREATE TABLE
  - INSERT (with parameterized queries)
  - SELECT (with WHERE clause)
  - UPDATE (with affected rows)
  - DELETE (with LIKE operator)

#### 3. Transaction Tests
- **TestDatabaseTransactions**: Tests transaction support
  - BEGIN/COMMIT transaction
  - BEGIN/ROLLBACK transaction
  - Data consistency verification

#### 4. Feature Detection Tests
- **TestDatabaseFeatureDetection**: Tests database capability detection
  - Transactions support
  - JSON/JSONB support
  - CTE (Common Table Expressions)
  - Window functions
  - Auto-increment/sequences
  - Arrays (PostgreSQL)

### Document Database Tests (`documents/integration_test.go`)

#### 1. Connection Tests
- **TestDocumentDatabaseConnection**: Verifies document database connectivity
  - Creates database instance
  - Tests ping

#### 2. CRUD Operation Tests
- **TestDocumentCRUDOperations**: Tests document operations
  - InsertOne
  - InsertMany
  - FindOne
  - Find (with query operators)
  - UpdateOne
  - UpdateMany
  - DeleteOne
  - DeleteMany
  - Count

#### 3. Index Operation Tests
- **TestDocumentIndexOperations**: Tests index management
  - Create single field index
  - Create unique index
  - Create compound index
  - List indexes
  - Drop index

#### 4. Aggregation Tests
- **TestDocumentAggregation**: Tests aggregation pipeline
  - $group operator
  - $sum aggregation
  - $sort operator

## Expected Test Output

### Successful Run

```
=== RUN   TestDatabaseConnection
=== RUN   TestDatabaseConnection/mysql
    integration_test.go:89: ✓ mysql connection successful (dialect: mysql)
=== RUN   TestDatabaseConnection/postgres
    integration_test.go:89: ✓ postgres connection successful (dialect: postgres)
=== RUN   TestDatabaseConnection/mssql
    integration_test.go:89: ✓ mssql connection successful (dialect: mssql)
=== RUN   TestDatabaseConnection/oracle
    integration_test.go:89: ✓ oracle connection successful (dialect: oracle)
--- PASS: TestDatabaseConnection (0.50s)
    --- PASS: TestDatabaseConnection/mysql (0.12s)
    --- PASS: TestDatabaseConnection/postgres (0.13s)
    --- PASS: TestDatabaseConnection/mssql (0.12s)
    --- PASS: TestDatabaseConnection/oracle (0.13s)
```

### Skipped Tests

```
=== RUN   TestDatabaseConnection/oracle
    integration_test.go:62: Skipping oracle tests (SKIP_oracle_TESTS=true)
--- SKIP: TestDatabaseConnection/oracle (0.00s)
```

### Failed Tests

```
=== RUN   TestDatabaseConnection/mysql
    integration_test.go:79: Failed to connect to mysql: dial tcp 127.0.0.1:3306: connect: connection refused
--- FAIL: TestDatabaseConnection/mysql (0.01s)
```

## Troubleshooting

### Connection Refused Errors

**Problem:** `dial tcp ... connect: connection refused`

**Solutions:**
1. Verify database is running: `docker ps` or check your database service
2. Check database host and port configuration
3. Verify firewall rules allow connections

### Authentication Errors

**Problem:** `Access denied for user ...` or `Login failed for user ...`

**Solutions:**
1. Verify username and password are correct
2. Check that the user has been created in the database
3. Verify the user has necessary permissions

### Database Not Found Errors

**Problem:** `database "iac" does not exist`

**Solutions:**
1. Create the database manually or use initialization scripts
2. Run: `docker-compose -f docker-compose.databases.yml down -v` then `up -d` to recreate
3. Check database name in environment variables

### SSL/TLS Errors (PostgreSQL)

**Problem:** `SSL is required` or `SSL connection error`

**Solutions:**
1. For local testing, set `sslmode=disable` in connection string
2. For production, configure proper SSL certificates

### Timeout Errors

**Problem:** `context deadline exceeded` or `timeout`

**Solutions:**
1. Increase connection timeout in test configuration
2. Check network connectivity
3. Verify database is not under heavy load

## Continuous Integration

### GitHub Actions Example

```yaml
name: Database Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: mysql_root_pass
          MYSQL_DATABASE: iac
          MYSQL_USER: iac_user
          MYSQL_PASSWORD: iac_pass
        ports:
          - 3306:3306

      postgres:
        image: postgres:14
        env:
          POSTGRES_DB: iac
          POSTGRES_USER: iac_user
          POSTGRES_PASSWORD: iac_pass
        ports:
          - 5432:5432

      mongodb:
        image: mongo:6
        env:
          MONGO_INITDB_ROOT_USERNAME: admin
          MONGO_INITDB_ROOT_PASSWORD: mongo_pass
        ports:
          - 27017:27017

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.20'

      - name: Run tests
        env:
          TEST_MYSQL_HOST: localhost
          TEST_MYSQL_PORT: 3306
          TEST_POSTGRES_HOST: localhost
          TEST_POSTGRES_PORT: 5432
          TEST_MONGODB_HOST: localhost
          TEST_MONGODB_PORT: 27017
          SKIP_mssql_TESTS: true
          SKIP_oracle_TESTS: true
        run: |
          go test -v ./databases/integration_test.go
          go test -v ./documents/integration_test.go
```

## Test Data Cleanup

All tests automatically clean up their test data:
- Relational tests: Drop test tables after each test
- Document tests: Drop test collections after each test

If a test is interrupted, you may need to manually clean up:

```sql
-- MySQL/PostgreSQL
DROP TABLE IF EXISTS test_mysql_1234567890;

-- MongoDB
use iac_test;
db.test_mongodb_1234567890.drop();
```

## Best Practices

1. **Always run tests in a test database**, never in production
2. **Use Docker** for consistent test environments
3. **Set appropriate timeouts** for slow databases (Oracle)
4. **Run tests in parallel** when possible: `go test -v -parallel=4 ./...`
5. **Monitor test coverage**: `go test -cover ./databases/... ./documents/...`
6. **Include tests in PR reviews** to catch database compatibility issues early

## Performance Benchmarks

See `/databases/benchmark_test.go` for performance benchmarking tests across all database types.

Run benchmarks:
```bash
go test -bench=. -benchmem ./databases/benchmark_test.go
```

## Additional Resources

- [Database Setup Scripts](/scripts/README.md)
- [Multi-Database Migration Guide](/services/MULTIDB_MIGRATION.md)
- [Docker Compose Configuration](/docker-compose.databases.yml)
- [Database Improvement Plan](/db_improvement.md)
