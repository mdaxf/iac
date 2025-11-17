#!/bin/bash

# IAC Database Quick Start Script
# This script starts all database containers and waits for them to be ready

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "=========================================="
echo "IAC Database Setup"
echo "=========================================="

# Change to project root
cd "$PROJECT_ROOT"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running. Please start Docker first."
    exit 1
fi

# Start databases
echo ""
echo "Starting database containers..."
docker-compose -f docker-compose.databases.yml up -d

# Wait for databases to be ready
echo ""
echo "Waiting for databases to be ready..."

# Wait for MySQL
echo -n "MySQL: "
until docker exec iac-mysql-primary mysqladmin ping -h localhost -u root -pmysql_root_pass --silent > /dev/null 2>&1; do
    echo -n "."
    sleep 2
done
echo " Ready ✓"

# Wait for PostgreSQL
echo -n "PostgreSQL: "
until docker exec iac-postgres pg_isready -U iac_user -d iac > /dev/null 2>&1; do
    echo -n "."
    sleep 2
done
echo " Ready ✓"

# Wait for MongoDB
echo -n "MongoDB: "
until docker exec iac-mongodb mongosh --eval "db.adminCommand('ping')" > /dev/null 2>&1; do
    echo -n "."
    sleep 2
done
echo " Ready ✓"

# Wait for MSSQL (may take longer)
echo -n "MSSQL: "
for i in {1..30}; do
    if docker exec iac-mssql /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P MsSql_Pass123! -Q "SELECT 1" > /dev/null 2>&1; then
        echo " Ready ✓"
        break
    fi
    echo -n "."
    sleep 3
done

# Wait for Oracle (may take longer)
echo -n "Oracle: "
for i in {1..40}; do
    if docker exec iac-oracle healthcheck.sh > /dev/null 2>&1; then
        echo " Ready ✓"
        break
    fi
    echo -n "."
    sleep 3
done

echo ""
echo "=========================================="
echo "All databases are ready!"
echo "=========================================="
echo ""
echo "Connection Information:"
echo "----------------------"
echo "MySQL:      localhost:3306  (user: iac_user, pass: iac_pass)"
echo "PostgreSQL: localhost:5432  (user: iac_user, pass: iac_pass)"
echo "MSSQL:      localhost:1433  (user: sa, pass: MsSql_Pass123!)"
echo "Oracle:     localhost:1521  (user: iac_user, pass: iac_pass)"
echo "MongoDB:    localhost:27017 (user: iac_user, pass: iac_pass)"
echo ""
echo "Admin Tools:"
echo "------------"
echo "phpMyAdmin:    http://localhost:8080"
echo "pgAdmin:       http://localhost:8081"
echo "Mongo Express: http://localhost:8082"
echo ""
echo "To stop: docker-compose -f docker-compose.databases.yml down"
echo "=========================================="
