#!/bin/bash

# IAC Database Performance Benchmark Runner
# This script runs performance benchmarks for all supported databases

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "=========================================="
echo "IAC Database Performance Benchmarks"
echo "=========================================="
echo ""

# Change to project root
cd "$PROJECT_ROOT"

# Default benchmark configuration
export BENCHMARK_TIME="${BENCHMARK_TIME:-5s}"
export BENCHMARK_ITERATIONS="${BENCHMARK_ITERATIONS:-}"
export BENCHMARK_MEMORY="${BENCHMARK_MEMORY:-true}"

# Default test configuration
export TEST_MYSQL_HOST="${TEST_MYSQL_HOST:-localhost}"
export TEST_MYSQL_PORT="${TEST_MYSQL_PORT:-3306}"
export TEST_POSTGRES_HOST="${TEST_POSTGRES_HOST:-localhost}"
export TEST_POSTGRES_PORT="${TEST_POSTGRES_PORT:-5432}"
export TEST_MONGODB_HOST="${TEST_MONGODB_HOST:-localhost}"
export TEST_MONGODB_PORT="${TEST_MONGODB_PORT:-27017}"

# Parse command line arguments
BENCHMARK_TYPE="${1:-all}"
OUTPUT_FILE="${2:-benchmark_results.txt}"

# Benchmark flags
BENCH_FLAGS="-bench=. -benchmem"

if [ -n "$BENCHMARK_TIME" ]; then
    BENCH_FLAGS="$BENCH_FLAGS -benchtime=$BENCHMARK_TIME"
fi

if [ -n "$BENCHMARK_ITERATIONS" ]; then
    BENCH_FLAGS="$BENCH_FLAGS -benchtime=${BENCHMARK_ITERATIONS}x"
fi

echo "Benchmark configuration:"
echo "  Type: $BENCHMARK_TYPE"
echo "  Time: $BENCHMARK_TIME"
echo "  Output: $OUTPUT_FILE"
echo ""

# Function to run benchmarks
run_benchmarks() {
    local package=$1
    local name=$2

    echo "----------------------------------------"
    echo "Running $name benchmarks..."
    echo "----------------------------------------"
    echo ""

    go test $BENCH_FLAGS "$package" | tee -a "$OUTPUT_FILE"

    echo ""
}

# Clear previous results
> "$OUTPUT_FILE"

echo "Starting benchmarks at $(date)" | tee -a "$OUTPUT_FILE"
echo "" | tee -a "$OUTPUT_FILE"

# Run benchmarks based on type
case $BENCHMARK_TYPE in
    all)
        echo "Running all database performance benchmarks..."
        run_benchmarks "./databases/benchmark_test.go" "Relational Database"
        run_benchmarks "./documents/benchmark_test.go" "Document Database"
        ;;
    relational|databases)
        echo "Running relational database benchmarks..."
        run_benchmarks "./databases/benchmark_test.go" "Relational Database"
        ;;
    document|documents)
        echo "Running document database benchmarks..."
        run_benchmarks "./documents/benchmark_test.go" "Document Database"
        ;;
    connection)
        echo "Running connection benchmarks..."
        go test $BENCH_FLAGS ./databases/benchmark_test.go -run=^$ -bench=BenchmarkDatabaseConnection | tee -a "$OUTPUT_FILE"
        ;;
    insert)
        echo "Running insert benchmarks..."
        go test $BENCH_FLAGS ./databases/benchmark_test.go -run=^$ -bench=BenchmarkDatabaseInsert | tee -a "$OUTPUT_FILE"
        go test $BENCH_FLAGS ./documents/benchmark_test.go -run=^$ -bench=BenchmarkDocumentInsert | tee -a "$OUTPUT_FILE"
        ;;
    select|query)
        echo "Running query benchmarks..."
        go test $BENCH_FLAGS ./databases/benchmark_test.go -run=^$ -bench=BenchmarkDatabaseSelect | tee -a "$OUTPUT_FILE"
        go test $BENCH_FLAGS ./documents/benchmark_test.go -run=^$ -bench=BenchmarkDocumentFind | tee -a "$OUTPUT_FILE"
        ;;
    update)
        echo "Running update benchmarks..."
        go test $BENCH_FLAGS ./databases/benchmark_test.go -run=^$ -bench=BenchmarkDatabaseUpdate | tee -a "$OUTPUT_FILE"
        go test $BENCH_FLAGS ./documents/benchmark_test.go -run=^$ -bench=BenchmarkDocumentUpdate | tee -a "$OUTPUT_FILE"
        ;;
    transaction)
        echo "Running transaction benchmarks..."
        go test $BENCH_FLAGS ./databases/benchmark_test.go -run=^$ -bench=BenchmarkDatabaseTransaction | tee -a "$OUTPUT_FILE"
        ;;
    bulk)
        echo "Running bulk operation benchmarks..."
        go test $BENCH_FLAGS ./databases/benchmark_test.go -run=^$ -bench=BenchmarkDatabaseBulkInsert | tee -a "$OUTPUT_FILE"
        go test $BENCH_FLAGS ./documents/benchmark_test.go -run=^$ -bench=BenchmarkDocumentInsertMany | tee -a "$OUTPUT_FILE"
        ;;
    concurrent)
        echo "Running concurrent operation benchmarks..."
        go test $BENCH_FLAGS ./databases/benchmark_test.go -run=^$ -bench=BenchmarkDatabaseConcurrentReads | tee -a "$OUTPUT_FILE"
        go test $BENCH_FLAGS ./documents/benchmark_test.go -run=^$ -bench=BenchmarkDocumentConcurrentReads | tee -a "$OUTPUT_FILE"
        ;;
    *)
        echo "Unknown benchmark type: $BENCHMARK_TYPE"
        echo ""
        echo "Usage: $0 [benchmark_type] [output_file]"
        echo ""
        echo "Benchmark types:"
        echo "  all            - Run all benchmarks (default)"
        echo "  relational     - Run relational database benchmarks only"
        echo "  document       - Run document database benchmarks only"
        echo "  connection     - Run connection benchmarks"
        echo "  insert         - Run insert benchmarks"
        echo "  select         - Run select/query benchmarks"
        echo "  update         - Run update benchmarks"
        echo "  transaction    - Run transaction benchmarks"
        echo "  bulk           - Run bulk operation benchmarks"
        echo "  concurrent     - Run concurrent operation benchmarks"
        echo ""
        echo "Examples:"
        echo "  $0                           # Run all benchmarks"
        echo "  $0 all results.txt           # Run all benchmarks, save to results.txt"
        echo "  $0 insert                    # Run insert benchmarks only"
        echo "  BENCHMARK_TIME=10s $0 all    # Run all benchmarks for 10 seconds each"
        exit 1
        ;;
esac

echo ""
echo "=========================================="
echo "Benchmarks completed at $(date)"
echo "Results saved to: $OUTPUT_FILE"
echo "=========================================="
echo ""

# Generate summary if benchstat is available
if command -v benchstat &> /dev/null; then
    echo "Generating benchmark statistics..."
    benchstat "$OUTPUT_FILE" > "${OUTPUT_FILE%.txt}_stats.txt"
    echo "Statistics saved to: ${OUTPUT_FILE%.txt}_stats.txt"
else
    echo "Tip: Install benchstat for better analysis:"
    echo "  go install golang.org/x/perf/cmd/benchstat@latest"
fi

echo ""
echo "To compare results:"
echo "  benchstat old_results.txt new_results.txt"
echo ""
echo "To visualize results, you can use:"
echo "  go-torch -b benchmark_results.txt"
