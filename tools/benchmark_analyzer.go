// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// BenchmarkResult represents a parsed benchmark result
type BenchmarkResult struct {
	Name       string
	Database   string
	Operation  string
	Iterations int
	NsPerOp    float64
	BytesPerOp int64
	AllocsPerOp int64
}

// DatabaseStats holds aggregated statistics for a database
type DatabaseStats struct {
	Database     string
	TotalOps     int
	AvgNsPerOp   float64
	AvgBytesPerOp float64
	AvgAllocsPerOp float64
	Results      []BenchmarkResult
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: benchmark_analyzer <benchmark_results_file>")
		fmt.Println("")
		fmt.Println("This tool analyzes database benchmark results and generates a comparison report.")
		os.Exit(1)
	}

	filename := os.Args[1]

	results, err := parseBenchmarkFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing benchmark file: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("No benchmark results found in file")
		os.Exit(1)
	}

	// Generate report
	generateReport(results)
}

func parseBenchmarkFile(filename string) ([]BenchmarkResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []BenchmarkResult
	scanner := bufio.NewScanner(file)

	// Regex to parse benchmark lines
	// Example: BenchmarkDatabaseConnection/mysql-8  100  12345678 ns/op  1234 B/op  56 allocs/op
	benchmarkRegex := regexp.MustCompile(`^Benchmark(\w+)/(\w+)-\d+\s+(\d+)\s+([\d.]+)\s+ns/op(?:\s+([\d.]+)\s+B/op)?(?:\s+([\d.]+)\s+allocs/op)?`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := benchmarkRegex.FindStringSubmatch(line)

		if matches != nil && len(matches) >= 5 {
			iterations, _ := strconv.Atoi(matches[3])
			nsPerOp, _ := strconv.ParseFloat(matches[4], 64)

			var bytesPerOp int64
			var allocsPerOp int64

			if len(matches) > 5 && matches[5] != "" {
				bytesPerOp, _ = strconv.ParseInt(matches[5], 10, 64)
			}

			if len(matches) > 6 && matches[6] != "" {
				allocsPerOp, _ = strconv.ParseInt(matches[6], 10, 64)
			}

			result := BenchmarkResult{
				Name:        matches[0],
				Operation:   matches[1],
				Database:    matches[2],
				Iterations:  iterations,
				NsPerOp:     nsPerOp,
				BytesPerOp:  bytesPerOp,
				AllocsPerOp: allocsPerOp,
			}

			results = append(results, result)
		}
	}

	return results, scanner.Err()
}

func generateReport(results []BenchmarkResult) {
	fmt.Println("========================================")
	fmt.Println("Database Benchmark Comparison Report")
	fmt.Println("========================================")
	fmt.Println("")

	// Group results by operation
	operationMap := make(map[string][]BenchmarkResult)
	for _, result := range results {
		operationMap[result.Operation] = append(operationMap[result.Operation], result)
	}

	// Print results by operation
	operations := make([]string, 0, len(operationMap))
	for op := range operationMap {
		operations = append(operations, op)
	}
	sort.Strings(operations)

	for _, operation := range operations {
		opResults := operationMap[operation]
		printOperationComparison(operation, opResults)
		fmt.Println("")
	}

	// Generate database statistics
	dbStats := calculateDatabaseStats(results)
	printDatabaseStats(dbStats)

	// Generate recommendations
	printRecommendations(dbStats, operationMap)
}

func printOperationComparison(operation string, results []BenchmarkResult) {
	fmt.Printf("Operation: %s\n", operation)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-15s %12s %15s %15s %15s\n", "Database", "Iterations", "ns/op", "B/op", "allocs/op")
	fmt.Println(strings.Repeat("-", 80))

	// Sort by performance (ns/op)
	sort.Slice(results, func(i, j int) bool {
		return results[i].NsPerOp < results[j].NsPerOp
	})

	for _, result := range results {
		fmt.Printf("%-15s %12d %15.0f %15d %15d\n",
			result.Database,
			result.Iterations,
			result.NsPerOp,
			result.BytesPerOp,
			result.AllocsPerOp)
	}

	// Print fastest
	if len(results) > 0 {
		fastest := results[0]
		slowest := results[len(results)-1]
		speedup := slowest.NsPerOp / fastest.NsPerOp

		fmt.Println("")
		fmt.Printf("  ⚡ Fastest: %s (%.0f ns/op)\n", fastest.Database, fastest.NsPerOp)
		fmt.Printf("  🐌 Slowest: %s (%.0f ns/op) - %.2fx slower\n", slowest.Database, slowest.NsPerOp, speedup)
	}
}

func calculateDatabaseStats(results []BenchmarkResult) []DatabaseStats {
	statsMap := make(map[string]*DatabaseStats)

	for _, result := range results {
		if _, exists := statsMap[result.Database]; !exists {
			statsMap[result.Database] = &DatabaseStats{
				Database: result.Database,
				Results:  []BenchmarkResult{},
			}
		}

		stats := statsMap[result.Database]
		stats.Results = append(stats.Results, result)
		stats.TotalOps++
		stats.AvgNsPerOp += result.NsPerOp
		stats.AvgBytesPerOp += float64(result.BytesPerOp)
		stats.AvgAllocsPerOp += float64(result.AllocsPerOp)
	}

	// Calculate averages
	var statsList []DatabaseStats
	for _, stats := range statsMap {
		if stats.TotalOps > 0 {
			stats.AvgNsPerOp /= float64(stats.TotalOps)
			stats.AvgBytesPerOp /= float64(stats.TotalOps)
			stats.AvgAllocsPerOp /= float64(stats.TotalOps)
		}
		statsList = append(statsList, *stats)
	}

	// Sort by average performance
	sort.Slice(statsList, func(i, j int) bool {
		return statsList[i].AvgNsPerOp < statsList[j].AvgNsPerOp
	})

	return statsList
}

func printDatabaseStats(stats []DatabaseStats) {
	fmt.Println("========================================")
	fmt.Println("Overall Database Performance")
	fmt.Println("========================================")
	fmt.Println("")
	fmt.Printf("%-15s %10s %15s %15s %15s\n", "Database", "Operations", "Avg ns/op", "Avg B/op", "Avg allocs/op")
	fmt.Println(strings.Repeat("-", 80))

	for _, stat := range stats {
		fmt.Printf("%-15s %10d %15.0f %15.0f %15.0f\n",
			stat.Database,
			stat.TotalOps,
			stat.AvgNsPerOp,
			stat.AvgBytesPerOp,
			stat.AvgAllocsPerOp)
	}

	fmt.Println("")
	if len(stats) > 0 {
		fmt.Printf("🏆 Best Overall Performance: %s\n", stats[0].Database)
	}
}

func printRecommendations(stats []DatabaseStats, operationMap map[string][]BenchmarkResult) {
	fmt.Println("")
	fmt.Println("========================================")
	fmt.Println("Recommendations")
	fmt.Println("========================================")
	fmt.Println("")

	// Find best database for each operation type
	recommendations := make(map[string]string)
	for operation, results := range operationMap {
		if len(results) == 0 {
			continue
		}

		sort.Slice(results, func(i, j int) bool {
			return results[i].NsPerOp < results[j].NsPerOp
		})

		recommendations[operation] = results[0].Database
	}

	fmt.Println("Best database for each operation:")
	for operation, database := range recommendations {
		fmt.Printf("  • %-20s → %s\n", operation, database)
	}

	fmt.Println("")
	fmt.Println("General Recommendations:")

	// Analyze patterns
	if len(stats) >= 2 {
		best := stats[0]
		worst := stats[len(stats)-1]

		fmt.Printf("  1. %s shows the best overall performance\n", best.Database)
		fmt.Printf("  2. Consider optimizing %s queries (%.0f%% slower than %s)\n",
			worst.Database,
			((worst.AvgNsPerOp-best.AvgNsPerOp)/best.AvgNsPerOp)*100,
			best.Database)

		// Check memory usage
		var highMemDB string
		var maxMem float64
		for _, stat := range stats {
			if stat.AvgBytesPerOp > maxMem {
				maxMem = stat.AvgBytesPerOp
				highMemDB = stat.Database
			}
		}

		if maxMem > 0 {
			fmt.Printf("  3. %s has the highest memory usage (%.0f B/op average)\n", highMemDB, maxMem)
		}

		// Check allocations
		var highAllocDB string
		var maxAlloc float64
		for _, stat := range stats {
			if stat.AvgAllocsPerOp > maxAlloc {
				maxAlloc = stat.AvgAllocsPerOp
				highAllocDB = stat.Database
			}
		}

		if maxAlloc > 0 {
			fmt.Printf("  4. %s has the most allocations (%.0f allocs/op average)\n", highAllocDB, maxAlloc)
		}
	}

	fmt.Println("")
	fmt.Println("Performance Optimization Tips:")
	fmt.Println("  • Use connection pooling for better performance")
	fmt.Println("  • Enable query caching where applicable")
	fmt.Println("  • Create indexes on frequently queried columns")
	fmt.Println("  • Use bulk operations instead of single inserts/updates")
	fmt.Println("  • Consider read replicas for read-heavy workloads")
}
