# Database Performance Benchmark Guide

This guide explains how to run and analyze performance benchmarks for all supported database types in the IAC system.

## Overview

The IAC system includes comprehensive performance benchmarks for:

**Relational Databases:**
- MySQL
- PostgreSQL
- Microsoft SQL Server (MSSQL)
- Oracle

**Document Databases:**
- MongoDB
- PostgreSQL JSONB

## Benchmark Categories

### Relational Database Benchmarks

1. **BenchmarkDatabaseConnection** - Connection establishment performance
2. **BenchmarkDatabasePing** - Ping/health check performance
3. **BenchmarkDatabaseInsert** - Single row INSERT performance
4. **BenchmarkDatabaseSelect** - SELECT query performance
5. **BenchmarkDatabaseUpdate** - UPDATE operation performance
6. **BenchmarkDatabaseTransaction** - Transaction commit/rollback performance
7. **BenchmarkDatabaseBulkInsert** - Bulk INSERT performance (100 rows per batch)
8. **BenchmarkDatabaseConcurrentReads** - Concurrent SELECT performance

### Document Database Benchmarks

1. **BenchmarkDocumentInsertOne** - Single document insert performance
2. **BenchmarkDocumentInsertMany** - Bulk document insert performance (100 docs per batch)
3. **BenchmarkDocumentFindOne** - Single document lookup performance
4. **BenchmarkDocumentFind** - Multi-document query performance
5. **BenchmarkDocumentUpdateOne** - Single document update performance
6. **BenchmarkDocumentUpdateMany** - Bulk document update performance
7. **BenchmarkDocumentAggregate** - Aggregation pipeline performance
8. **BenchmarkDocumentIndexedQuery** - Query performance with indexes
9. **BenchmarkDocumentConcurrentReads** - Concurrent document read performance

## Prerequisites

### 1. Database Setup

Start all databases using Docker Compose:

```bash
# Start all databases
docker-compose -f docker-compose.databases.yml up -d

# Or use the quick start script
./scripts/start-databases.sh

# Wait for databases to be ready (about 30 seconds)
sleep 30
```

### 2. Environment Configuration

Set environment variables for test databases (same as integration tests):

```bash
export TEST_MYSQL_HOST=localhost
export TEST_MYSQL_PORT=3306
export TEST_POSTGRES_HOST=localhost
export TEST_POSTGRES_PORT=5432
export TEST_MONGODB_HOST=localhost
export TEST_MONGODB_PORT=27017
```

### 3. Install Analysis Tools (Optional)

For advanced benchmark analysis:

```bash
# Install benchstat for statistical comparison
go install golang.org/x/perf/cmd/benchstat@latest

# Build the benchmark analyzer tool
cd tools
go build -o benchmark_analyzer benchmark_analyzer.go
```

## Running Benchmarks

### Quick Start

```bash
# Run all benchmarks (takes 5-10 minutes)
./scripts/run-benchmarks.sh all

# Run specific benchmark categories
./scripts/run-benchmarks.sh insert
./scripts/run-benchmarks.sh select
./scripts/run-benchmarks.sh transaction
```

### Using Go Test Directly

```bash
# Run all relational database benchmarks
go test -bench=. -benchmem ./databases/benchmark_test.go

# Run all document database benchmarks
go test -bench=. -benchmem ./documents/benchmark_test.go

# Run specific benchmark
go test -bench=BenchmarkDatabaseInsert -benchmem ./databases/benchmark_test.go

# Run with custom benchmark time (default is 1 second)
go test -bench=. -benchmem -benchtime=10s ./databases/benchmark_test.go

# Run with specific number of iterations
go test -bench=. -benchmem -benchtime=1000x ./databases/benchmark_test.go
```

### Using Makefile

Add to your Makefile:

```makefile
.PHONY: benchmark-all benchmark-quick benchmark-insert benchmark-select

benchmark-all:
	./scripts/run-benchmarks.sh all benchmark_results.txt

benchmark-quick:
	BENCHMARK_TIME=1s ./scripts/run-benchmarks.sh all quick_results.txt

benchmark-insert:
	./scripts/run-benchmarks.sh insert insert_results.txt

benchmark-select:
	./scripts/run-benchmarks.sh select select_results.txt
```

Then run:

```bash
make benchmark-all
```

## Benchmark Configuration

### Benchmark Time

Control how long each benchmark runs:

```bash
# Run each benchmark for 10 seconds
BENCHMARK_TIME=10s ./scripts/run-benchmarks.sh all

# Run each benchmark for exactly 1000 iterations
BENCHMARK_ITERATIONS=1000 ./scripts/run-benchmarks.sh all
```

### Database Selection

Skip specific databases:

```bash
# Skip Oracle and MSSQL
export SKIP_oracle_TESTS=true
export SKIP_mssql_TESTS=true
./scripts/run-benchmarks.sh all
```

### Output Options

```bash
# Save results to custom file
./scripts/run-benchmarks.sh all my_results.txt

# Run benchmarks and save to timestamped file
./scripts/run-benchmarks.sh all "benchmark_$(date +%Y%m%d_%H%M%S).txt"
```

## Analyzing Results

### Using Benchmark Analyzer

The custom benchmark analyzer generates a comprehensive comparison report:

```bash
# Run benchmarks
./scripts/run-benchmarks.sh all results.txt

# Analyze results
./tools/benchmark_analyzer results.txt
```

**Output includes:**
- Operation-by-operation comparison
- Fastest vs slowest database for each operation
- Overall database performance ranking
- Memory and allocation statistics
- Specific recommendations

**Example output:**

```
========================================
Database Benchmark Comparison Report
========================================

Operation: DatabaseInsert
--------------------------------------------------------------------------------
Database         Iterations          ns/op           B/op      allocs/op
--------------------------------------------------------------------------------
postgres             10000         125000           1024             12
mysql                 8000         156000           1536             15
mssql                 5000         245000           2048             20

  ⚡ Fastest: postgres (125000 ns/op)
  🐌 Slowest: mssql (245000 ns/op) - 1.96x slower

========================================
Overall Database Performance
========================================

Database         Operations      Avg ns/op      Avg B/op  Avg allocs/op
--------------------------------------------------------------------------------
postgres                  8         145000          1200             14
mysql                     8         178000          1600             17
mongodb                   6         189000          1800             19
mssql                     8         267000          2100             22

🏆 Best Overall Performance: postgres

========================================
Recommendations
========================================

Best database for each operation:
  • DatabaseInsert      → postgres
  • DatabaseSelect      → mysql
  • DatabaseUpdate      → postgres
  • DocumentInsertMany  → mongodb

General Recommendations:
  1. postgres shows the best overall performance
  2. Consider optimizing mssql queries (84% slower than postgres)
  3. mssql has the highest memory usage (2100 B/op average)
  4. Use bulk operations instead of single inserts/updates
  5. Consider read replicas for read-heavy workloads
```

### Using Benchstat

For statistical comparison between two benchmark runs:

```bash
# Run baseline benchmarks
./scripts/run-benchmarks.sh all baseline.txt

# Make changes to your code...

# Run new benchmarks
./scripts/run-benchmarks.sh all optimized.txt

# Compare results
benchstat baseline.txt optimized.txt
```

**Example output:**

```
name                     old time/op    new time/op    delta
DatabaseInsert/mysql       156µs ± 2%     142µs ± 3%   -8.97%  (p=0.000 n=10+10)
DatabaseInsert/postgres    125µs ± 1%     118µs ± 2%   -5.60%  (p=0.000 n=10+10)

name                     old alloc/op   new alloc/op   delta
DatabaseInsert/mysql      1.54kB ± 0%    1.28kB ± 0%  -16.88%  (p=0.000 n=10+10)
DatabaseInsert/postgres   1.02kB ± 0%    0.96kB ± 0%   -5.88%  (p=0.000 n=10+10)
```

## Understanding Benchmark Results

### Metrics Explained

- **ns/op** - Nanoseconds per operation (lower is better)
  - 1,000 ns = 1 microsecond (µs)
  - 1,000,000 ns = 1 millisecond (ms)
  - 1,000,000,000 ns = 1 second (s)

- **B/op** - Bytes allocated per operation (lower is better)
  - Measures memory allocation
  - Lower values mean less GC pressure

- **allocs/op** - Number of allocations per operation (lower is better)
  - Measures how many times memory is allocated
  - Fewer allocations = better performance

- **Iterations** - Number of times the benchmark ran
  - Higher is better for statistical significance
  - Go automatically adjusts to run for ~1 second

### Performance Targets

**Good Performance:**
- Connection: < 1 ms (1,000,000 ns)
- Single INSERT: < 500 µs (500,000 ns)
- Single SELECT: < 200 µs (200,000 ns)
- Bulk INSERT (100 rows): < 10 ms (10,000,000 ns)
- Transaction: < 1 ms (1,000,000 ns)

**Excellent Performance:**
- Connection: < 500 µs
- Single INSERT: < 200 µs
- Single SELECT: < 100 µs
- Bulk INSERT (100 rows): < 5 ms
- Transaction: < 500 µs

## Best Practices

### 1. Consistent Environment

- Always run benchmarks on the same hardware
- Close other applications to reduce noise
- Use dedicated test databases, not production
- Run multiple times and average results

### 2. Baseline First

```bash
# Establish baseline before making changes
./scripts/run-benchmarks.sh all baseline.txt

# Make your optimizations...

# Compare with baseline
./scripts/run-benchmarks.sh all optimized.txt
benchstat baseline.txt optimized.txt
```

### 3. Focus on Bottlenecks

Run specific benchmarks to test optimizations:

```bash
# If inserts are slow, test only inserts
go test -bench=BenchmarkDatabase.*Insert -benchmem -benchtime=10s ./databases/benchmark_test.go
```

### 4. Use Appropriate Benchmark Time

```bash
# Quick check (1 second per benchmark)
BENCHMARK_TIME=1s ./scripts/run-benchmarks.sh all

# Accurate results (10 seconds per benchmark)
BENCHMARK_TIME=10s ./scripts/run-benchmarks.sh all

# Very accurate (30 seconds per benchmark - use for final validation)
BENCHMARK_TIME=30s ./scripts/run-benchmarks.sh all
```

### 5. Test Under Load

Use concurrent benchmarks to simulate real-world usage:

```bash
go test -bench=Concurrent -benchmem -benchtime=5s ./databases/benchmark_test.go
```

## Common Patterns

### Pattern 1: Before/After Comparison

```bash
# Before optimization
git checkout main
./scripts/run-benchmarks.sh all before.txt

# After optimization
git checkout optimization-branch
./scripts/run-benchmarks.sh all after.txt

# Compare
benchstat before.txt after.txt
```

### Pattern 2: Database Selection

```bash
# Benchmark all databases for a specific operation
go test -bench=BenchmarkDatabaseInsert -benchmem ./databases/benchmark_test.go

# Analyze results to choose best database for your use case
./tools/benchmark_analyzer results.txt
```

### Pattern 3: Continuous Benchmarking

```bash
# Add to CI/CD pipeline
#!/bin/bash
./scripts/run-benchmarks.sh all "benchmarks/$(date +%Y%m%d).txt"

# Compare with previous day
benchstat "benchmarks/$(date -d yesterday +%Y%m%d).txt" "benchmarks/$(date +%Y%m%d).txt"
```

## Troubleshooting

### Benchmarks Take Too Long

**Problem:** Benchmarks running for more than 30 minutes

**Solutions:**
- Use shorter benchmark time: `BENCHMARK_TIME=1s`
- Run specific benchmarks instead of all
- Skip slow databases: `export SKIP_mssql_TESTS=true`

### High Variance in Results

**Problem:** Results vary significantly between runs

**Solutions:**
- Close other applications
- Run for longer time: `BENCHMARK_TIME=10s`
- Run benchmarks multiple times and average
- Check for background processes

### Database Connection Errors

**Problem:** Benchmarks fail with connection errors

**Solutions:**
- Verify databases are running: `docker ps`
- Check environment variables
- Increase connection timeout in test config
- Verify database credentials

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Performance Benchmarks

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  benchmark:
    runs-on: ubuntu-latest

    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: root
          MYSQL_DATABASE: iac
        ports:
          - 3306:3306

      postgres:
        image: postgres:14
        env:
          POSTGRES_DB: iac
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.20'

      - name: Run benchmarks
        env:
          BENCHMARK_TIME: 5s
          SKIP_mssql_TESTS: true
          SKIP_oracle_TESTS: true
        run: |
          ./scripts/run-benchmarks.sh all benchmark_results.txt

      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: benchmark-results
          path: benchmark_results.txt

      - name: Compare with main
        if: github.event_name == 'pull_request'
        run: |
          git fetch origin main
          git checkout origin/main
          ./scripts/run-benchmarks.sh all main_results.txt
          git checkout -
          benchstat main_results.txt benchmark_results.txt
```

## Additional Resources

- [Go Benchmark Documentation](https://golang.org/pkg/testing/#hdr-Benchmarks)
- [Benchstat Tool](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [Database Setup Scripts](/scripts/README.md)
- [Integration Test Guide](/tests/TEST_GUIDE.md)

## Summary

Performance benchmarking is essential for:
- ✅ Choosing the right database for your workload
- ✅ Validating optimizations
- ✅ Preventing performance regressions
- ✅ Understanding database characteristics
- ✅ Making informed architectural decisions

Run benchmarks regularly and compare results to ensure your database layer remains performant!
