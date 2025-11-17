#!/bin/bash

# IAC Database Integration Test Runner
# This script runs integration tests for all supported databases

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "=========================================="
echo "IAC Database Integration Tests"
echo "=========================================="
echo ""

# Change to project root
cd "$PROJECT_ROOT"

# Default test configuration (can be overridden by environment variables)
export TEST_MYSQL_HOST="${TEST_MYSQL_HOST:-localhost}"
export TEST_MYSQL_PORT="${TEST_MYSQL_PORT:-3306}"
export TEST_MYSQL_DATABASE="${TEST_MYSQL_DATABASE:-iac}"
export TEST_MYSQL_USERNAME="${TEST_MYSQL_USERNAME:-iac_user}"
export TEST_MYSQL_PASSWORD="${TEST_MYSQL_PASSWORD:-iac_pass}"

export TEST_POSTGRES_HOST="${TEST_POSTGRES_HOST:-localhost}"
export TEST_POSTGRES_PORT="${TEST_POSTGRES_PORT:-5432}"
export TEST_POSTGRES_DATABASE="${TEST_POSTGRES_DATABASE:-iac}"
export TEST_POSTGRES_USERNAME="${TEST_POSTGRES_USERNAME:-iac_user}"
export TEST_POSTGRES_PASSWORD="${TEST_POSTGRES_PASSWORD:-iac_pass}"

export TEST_MSSQL_HOST="${TEST_MSSQL_HOST:-localhost}"
export TEST_MSSQL_PORT="${TEST_MSSQL_PORT:-1433}"
export TEST_MSSQL_DATABASE="${TEST_MSSQL_DATABASE:-iac}"
export TEST_MSSQL_USERNAME="${TEST_MSSQL_USERNAME:-sa}"
export TEST_MSSQL_PASSWORD="${TEST_MSSQL_PASSWORD:-MsSql_Pass123!}"

export TEST_ORACLE_HOST="${TEST_ORACLE_HOST:-localhost}"
export TEST_ORACLE_PORT="${TEST_ORACLE_PORT:-1521}"
export TEST_ORACLE_DATABASE="${TEST_ORACLE_DATABASE:-iac}"
export TEST_ORACLE_USERNAME="${TEST_ORACLE_USERNAME:-iac_user}"
export TEST_ORACLE_PASSWORD="${TEST_ORACLE_PASSWORD:-iac_pass}"

export TEST_MONGODB_HOST="${TEST_MONGODB_HOST:-localhost}"
export TEST_MONGODB_PORT="${TEST_MONGODB_PORT:-27017}"
export TEST_MONGODB_DATABASE="${TEST_MONGODB_DATABASE:-iac_test}"
export TEST_MONGODB_USERNAME="${TEST_MONGODB_USERNAME:-iac_user}"
export TEST_MONGODB_PASSWORD="${TEST_MONGODB_PASSWORD:-iac_pass}"

# Parse command line arguments
TEST_TYPE="${1:-all}"
TEST_VERBOSE="${2:-}"

# Function to check if database is available
check_database() {
    local db_type=$1
    local host=$2
    local port=$3

    echo -n "Checking $db_type ($host:$port)... "

    case $db_type in
        mysql)
            if command -v mysql &> /dev/null; then
                if mysql -h "$host" -P "$port" -u "$TEST_MYSQL_USERNAME" -p"$TEST_MYSQL_PASSWORD" -e "SELECT 1" &> /dev/null; then
                    echo "✓ Available"
                    return 0
                fi
            fi
            ;;
        postgres)
            if command -v psql &> /dev/null; then
                if PGPASSWORD="$TEST_POSTGRES_PASSWORD" psql -h "$host" -p "$port" -U "$TEST_POSTGRES_USERNAME" -d "$TEST_POSTGRES_DATABASE" -c "SELECT 1" &> /dev/null; then
                    echo "✓ Available"
                    return 0
                fi
            fi
            ;;
        mongodb)
            if command -v mongosh &> /dev/null || command -v mongo &> /dev/null; then
                # Try to connect using docker exec if client not available
                if docker exec iac-mongodb mongosh --eval "db.adminCommand('ping')" &> /dev/null; then
                    echo "✓ Available"
                    return 0
                fi
            fi
            ;;
    esac

    echo "✗ Not Available (will skip)"
    return 1
}

# Check database availability and set skip flags
echo "Checking database availability..."
echo ""

if ! check_database "mysql" "$TEST_MYSQL_HOST" "$TEST_MYSQL_PORT"; then
    export SKIP_mysql_TESTS=true
fi

if ! check_database "postgres" "$TEST_POSTGRES_HOST" "$TEST_POSTGRES_PORT"; then
    export SKIP_postgres_TESTS=true
    export SKIP_postgres_jsonb_TESTS=true
fi

if ! check_database "mongodb" "$TEST_MONGODB_HOST" "$TEST_MONGODB_PORT"; then
    export SKIP_mongodb_TESTS=true
fi

# MSSQL and Oracle often not available in dev environments
echo -n "Checking mssql ($TEST_MSSQL_HOST:$TEST_MSSQL_PORT)... "
if [ -z "${SKIP_mssql_TESTS}" ]; then
    echo "Not checked (set SKIP_mssql_TESTS=false to enable)"
    export SKIP_mssql_TESTS=true
else
    echo "Skipped"
fi

echo -n "Checking oracle ($TEST_ORACLE_HOST:$TEST_ORACLE_PORT)... "
if [ -z "${SKIP_oracle_TESTS}" ]; then
    echo "Not checked (set SKIP_oracle_TESTS=false to enable)"
    export SKIP_oracle_TESTS=true
else
    echo "Skipped"
fi

echo ""
echo "=========================================="
echo "Running Tests"
echo "=========================================="
echo ""

# Determine verbosity flag
VERBOSE_FLAG=""
if [ "$TEST_VERBOSE" = "-v" ] || [ "$TEST_VERBOSE" = "verbose" ]; then
    VERBOSE_FLAG="-v"
fi

# Run tests based on type
case $TEST_TYPE in
    all)
        echo "Running all database integration tests..."
        go test $VERBOSE_FLAG ./databases/integration_test.go
        go test $VERBOSE_FLAG ./documents/integration_test.go
        ;;
    relational|databases)
        echo "Running relational database integration tests..."
        go test $VERBOSE_FLAG ./databases/integration_test.go
        ;;
    document|documents)
        echo "Running document database integration tests..."
        go test $VERBOSE_FLAG ./documents/integration_test.go
        ;;
    mysql)
        echo "Running MySQL integration tests..."
        export SKIP_postgres_TESTS=true
        export SKIP_mssql_TESTS=true
        export SKIP_oracle_TESTS=true
        go test $VERBOSE_FLAG ./databases/integration_test.go -run TestDatabaseConnection/mysql
        go test $VERBOSE_FLAG ./databases/integration_test.go -run TestDatabaseBasicOperations/mysql
        go test $VERBOSE_FLAG ./databases/integration_test.go -run TestDatabaseTransactions/mysql
        ;;
    postgres|postgresql)
        echo "Running PostgreSQL integration tests..."
        export SKIP_mysql_TESTS=true
        export SKIP_mssql_TESTS=true
        export SKIP_oracle_TESTS=true
        go test $VERBOSE_FLAG ./databases/integration_test.go -run TestDatabaseConnection/postgres
        go test $VERBOSE_FLAG ./databases/integration_test.go -run TestDatabaseBasicOperations/postgres
        go test $VERBOSE_FLAG ./databases/integration_test.go -run TestDatabaseTransactions/postgres
        go test $VERBOSE_FLAG ./documents/integration_test.go -run TestDocumentDatabaseConnection/postgres_jsonb
        go test $VERBOSE_FLAG ./documents/integration_test.go -run TestDocumentCRUDOperations/postgres_jsonb
        ;;
    mongodb|mongo)
        echo "Running MongoDB integration tests..."
        go test $VERBOSE_FLAG ./documents/integration_test.go -run TestDocumentDatabaseConnection/mongodb
        go test $VERBOSE_FLAG ./documents/integration_test.go -run TestDocumentCRUDOperations/mongodb
        go test $VERBOSE_FLAG ./documents/integration_test.go -run TestDocumentIndexOperations/mongodb
        go test $VERBOSE_FLAG ./documents/integration_test.go -run TestDocumentAggregation/mongodb
        ;;
    *)
        echo "Unknown test type: $TEST_TYPE"
        echo ""
        echo "Usage: $0 [test_type] [-v]"
        echo ""
        echo "Test types:"
        echo "  all            - Run all integration tests (default)"
        echo "  relational     - Run relational database tests only"
        echo "  document       - Run document database tests only"
        echo "  mysql          - Run MySQL tests only"
        echo "  postgres       - Run PostgreSQL tests only"
        echo "  mongodb        - Run MongoDB tests only"
        echo ""
        echo "Examples:"
        echo "  $0                    # Run all tests"
        echo "  $0 all -v             # Run all tests with verbose output"
        echo "  $0 mysql              # Run MySQL tests only"
        echo "  $0 postgres -v        # Run PostgreSQL tests with verbose output"
        exit 1
        ;;
esac

echo ""
echo "=========================================="
echo "Tests completed"
echo "=========================================="
