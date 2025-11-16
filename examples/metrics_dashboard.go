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
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/dbinitializer"
	"github.com/mdaxf/iac/metrics"
)

func main() {
	fmt.Println("IAC Database Metrics Dashboard Example")
	fmt.Println("======================================")

	// Initialize databases from environment
	dbInit := dbinitializer.NewDatabaseInitializer()
	if err := dbInit.InitializeFromEnvironment(); err != nil {
		log.Fatalf("Failed to initialize databases: %v", err)
	}

	poolManager := dbInit.GetPoolManager()

	// Create metrics collector
	collector := databases.NewMetricsCollector()

	// Set custom slow query threshold (default is 1 second)
	collector.SetSlowQueryThreshold(500 * time.Millisecond)

	fmt.Println("\n1. Starting metrics dashboard...")
	fmt.Println("   Dashboard URL: http://localhost:8080")
	fmt.Println("   API Endpoint:  http://localhost:8080/api/metrics")

	// Start metrics dashboard in background
	dashboard := metrics.NewDashboard(collector)
	go func() {
		if err := dashboard.StartServer(":8080"); err != nil {
			log.Printf("Dashboard error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	fmt.Println("\n2. Collecting metrics from databases...")

	// Start periodic metrics collection
	go collectMetricsPeriodically(poolManager, collector)

	// Simulate some database operations
	fmt.Println("\n3. Simulating database operations...")
	simulateDatabaseOperations(poolManager, collector)

	fmt.Println("\n4. Dashboard is running!")
	fmt.Println("   Open http://localhost:8080 in your browser")
	fmt.Println("   Press Ctrl+C to stop")
	fmt.Println("")

	// Keep the program running
	select {}
}

// collectMetricsPeriodically collects metrics from all databases every 5 seconds
func collectMetricsPeriodically(poolManager *databases.PoolManager, collector *databases.MetricsCollector) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		dbTypes := poolManager.GetAllDatabases()

		for _, dbType := range dbTypes {
			db, err := poolManager.GetPrimary(dbType)
			if err != nil {
				log.Printf("Failed to get database %s: %v", dbType, err)
				continue
			}

			// Ping database and record metrics
			start := time.Now()
			err = db.Ping()
			duration := time.Since(start)

			collector.RecordQuery(dbType, "PING", duration, err)

			// Simulate getting connection pool stats
			// In a real application, you would get actual stats from the database
			collector.UpdateConnectionPool(dbType, 5, 10, 15, 10)

			db.Close()
		}
	}
}

// simulateDatabaseOperations simulates various database operations
func simulateDatabaseOperations(poolManager *databases.PoolManager, collector *databases.MetricsCollector) {
	dbTypes := poolManager.GetAllDatabases()

	// Simulate operations for each database type
	for _, dbType := range dbTypes {
		go simulateOperationsForDB(poolManager, collector, dbType)
	}
}

// simulateOperationsForDB simulates operations for a specific database
func simulateOperationsForDB(poolManager *databases.PoolManager, collector *databases.MetricsCollector, dbType string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	operations := []struct {
		queryType string
		minDuration time.Duration
		maxDuration time.Duration
		errorRate float64
	}{
		{"SELECT", 10 * time.Millisecond, 50 * time.Millisecond, 0.01},
		{"INSERT", 20 * time.Millisecond, 100 * time.Millisecond, 0.02},
		{"UPDATE", 30 * time.Millisecond, 150 * time.Millisecond, 0.02},
		{"DELETE", 15 * time.Millisecond, 80 * time.Millisecond, 0.01},
	}

	for range ticker.C {
		for _, op := range operations {
			// Random duration between min and max
			duration := op.minDuration + time.Duration(float64(op.maxDuration-op.minDuration)*randomFloat())

			// Random error based on error rate
			var err error
			if randomFloat() < op.errorRate {
				err = fmt.Errorf("simulated error")
			}

			// Record the query
			collector.RecordQuery(dbType, op.queryType, duration, err)
		}

		// Update connection pool stats (simulated)
		active := 3 + int(randomFloat()*7)  // 3-10
		idle := 5 + int(randomFloat()*10)   // 5-15
		collector.UpdateConnectionPool(dbType, active, idle, 20, 15)
	}
}

// randomFloat returns a random float between 0 and 1
func randomFloat() float64 {
	return float64(time.Now().UnixNano()%1000) / 1000.0
}
