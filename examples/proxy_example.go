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
	"fmt"
	"time"

	"github.com/mdaxf/iac/databases"
)

func main() {
	fmt.Println("IAC Database Proxy Layer Example")
	fmt.Println("=================================")

	basicProxyExample()
	queryRewriterExample()
	loadBalancingExample()
	queryLoggingExample()
	metricsExample()
}

func basicProxyExample() {
	fmt.Println("\n1. Basic Database Proxy")
	fmt.Println("------------------------")

	// Create proxy with default config
	proxy := databases.NewDatabaseProxy(nil)

	fmt.Println("Database proxy initialized")
	fmt.Printf("  Max Connections: %d\n", proxy.config.MaxConnections)
	fmt.Printf("  Query Timeout: %v\n", proxy.config.QueryTimeout)
	fmt.Printf("  Slow Query Threshold: %v\n", proxy.config.SlowQueryThreshold)
	fmt.Printf("  Query Rewrite: %v\n", proxy.config.EnableQueryRewrite)
	fmt.Printf("  Monitoring: %v\n", proxy.config.EnableMonitoring)
	fmt.Printf("  Load Balancing: %v\n", proxy.config.EnableLoadBalancing)
}

func queryRewriterExample() {
	fmt.Println("\n2. Query Rewriting")
	fmt.Println("-------------------")

	rewriter := databases.NewQueryRewriter()

	// Add custom rewrite rules
	rules := []databases.RewriteRule{
		{
			Name:    "optimize-select-star",
			Pattern: "SELECT * FROM users",
			Replace: "SELECT id, name, email FROM users LIMIT 100",
			Enabled: true,
		},
		{
			Name:    "add-index-hint",
			Pattern: "FROM orders WHERE",
			Replace: "FROM orders USE INDEX (idx_created_at) WHERE",
			Enabled: true,
		},
	}

	for _, rule := range rules {
		rewriter.AddRule(rule)
		fmt.Printf("Added rule: %s\n", rule.Name)
	}

	// Test query rewriting
	fmt.Println("\nQuery Rewriting Examples:")

	originalQuery1 := "SELECT * FROM users"
	rewritten1 := rewriter.Rewrite(originalQuery1)
	fmt.Printf("  Original:  %s\n", originalQuery1)
	fmt.Printf("  Rewritten: %s\n", rewritten1)

	originalQuery2 := "SELECT * FROM orders WHERE status = 'pending'"
	rewritten2 := rewriter.Rewrite(originalQuery2)
	fmt.Printf("\n  Original:  %s\n", originalQuery2)
	fmt.Printf("  Rewritten: %s\n", rewritten2)
}

func loadBalancingExample() {
	fmt.Println("\n3. Load Balancing")
	fmt.Println("------------------")

	// Create load balancer
	lb := databases.NewLoadBalancer(databases.RoundRobin)

	fmt.Println("Load balancer initialized with Round Robin strategy")
	fmt.Println("  Strategy: Round Robin")
	fmt.Println("  Databases: (simulated)")

	// In production, would add actual database connections
	// lb.AddDatabase(db1)
	// lb.AddDatabase(db2)
	// lb.AddDatabase(db3)

	fmt.Println("\nQuery distribution simulation:")
	for i := 1; i <= 6; i++ {
		dbNum := (i % 3) + 1 // Simulate round-robin
		fmt.Printf("  Query %d -> Database %d\n", i, dbNum)
	}
}

func queryLoggingExample() {
	fmt.Println("\n4. Query Logging")
	fmt.Println("-----------------")

	logger := databases.NewQueryLogger(100)

	// Simulate some queries
	queries := []struct {
		sql      string
		duration time.Duration
	}{
		{"SELECT * FROM users WHERE id = 1", 5 * time.Millisecond},
		{"SELECT * FROM orders WHERE user_id = 1", 15 * time.Millisecond},
		{"UPDATE users SET last_login = NOW() WHERE id = 1", 8 * time.Millisecond},
		{"INSERT INTO audit_log VALUES (...)", 3 * time.Millisecond},
	}

	for _, q := range queries {
		logger.Log(q.sql, q.duration, nil)
	}

	fmt.Printf("Logged %d queries\n", len(queries))

	// Get recent queries
	recent := logger.GetRecentQueries(3)
	fmt.Println("\nMost recent queries:")
	for i, log := range recent {
		fmt.Printf("  %d. %s (%.2fms)\n", i+1, log.Query, float64(log.Duration.Microseconds())/1000.0)
	}
}

func metricsExample() {
	fmt.Println("\n5. Proxy Metrics")
	fmt.Println("-----------------")

	proxy := databases.NewDatabaseProxy(nil)

	// Simulate some query activity
	proxy.totalQueries = 1000
	proxy.slowQueries = 15
	proxy.failedQueries = 3
	proxy.cacheHits = 450
	proxy.totalDuration = 2500000 // 2.5 seconds total

	metrics := proxy.GetMetrics()

	fmt.Println("Proxy Performance Metrics:")
	fmt.Printf("  Total Queries: %d\n", metrics.TotalQueries)
	fmt.Printf("  Slow Queries: %d (%.1f%%)\n",
		metrics.SlowQueries,
		float64(metrics.SlowQueries)/float64(metrics.TotalQueries)*100)
	fmt.Printf("  Failed Queries: %d (%.1f%%)\n",
		metrics.FailedQueries,
		float64(metrics.FailedQueries)/float64(metrics.TotalQueries)*100)
	fmt.Printf("  Cache Hits: %d (%.1f%%)\n",
		metrics.CacheHits,
		float64(metrics.CacheHits)/float64(metrics.TotalQueries)*100)
	fmt.Printf("  Average Duration: %.2fms\n",
		float64(metrics.AverageDuration.Microseconds())/1000.0)
	fmt.Printf("  Slow Query Threshold: %v\n", metrics.SlowQueryThreshold)
}

func completeProxyExample() {
	fmt.Println("\n6. Complete Proxy Setup")
	fmt.Println("------------------------")

	// Configure proxy
	config := &databases.ProxyConfig{
		MaxConnections:      200,
		IdleTimeout:         10 * time.Minute,
		QueryTimeout:        60 * time.Second,
		EnableQueryRewrite:  true,
		EnableMonitoring:    true,
		EnableLoadBalancing: true,
		SlowQueryThreshold:  500 * time.Millisecond,
	}

	proxy := databases.NewDatabaseProxy(config)

	fmt.Println("Production proxy configuration:")
	fmt.Printf("  Max Connections: %d\n", config.MaxConnections)
	fmt.Printf("  Idle Timeout: %v\n", config.IdleTimeout)
	fmt.Printf("  Query Timeout: %v\n", config.QueryTimeout)
	fmt.Printf("  Query Rewriting: %v\n", config.EnableQueryRewrite)
	fmt.Printf("  Monitoring: %v\n", config.EnableMonitoring)
	fmt.Printf("  Load Balancing: %v\n", config.EnableLoadBalancing)
	fmt.Printf("  Slow Query Alert: >%v\n", config.SlowQueryThreshold)

	// In production, would set up components:
	// proxy.SetPoolManager(poolManager)
	// proxy.SetReplicaManager(replicaManager)
	// proxy.SetQueryCache(queryCache)

	fmt.Println("\nProxy ready for production use!")
	fmt.Println("Features enabled:")
	fmt.Println("  ✓ Connection pooling")
	fmt.Println("  ✓ Query rewriting")
	fmt.Println("  ✓ Load balancing")
	fmt.Println("  ✓ Query monitoring")
	fmt.Println("  ✓ Slow query detection")
	fmt.Println("  ✓ Query logging")

	_ = proxy
}
