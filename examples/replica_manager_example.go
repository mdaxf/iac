//go:build ignore
// +build ignore
Copyright 2023 IAC. All Rights Reserved.
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
	"database/sql"
	"fmt"
	"log"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
)

func main() {
	fmt.Println("IAC Read Replica Manager Example")
	fmt.Println("==================================")

	// Example 1: Basic replica management
	basicReplicaManagement()

	// Example 2: Weighted load balancing
	weightedLoadBalancing()

	// Example 3: Replica lag monitoring
	replicaLagMonitoring()

	// Example 4: Automatic failover
	automaticFailover()
}

func basicReplicaManagement() {
	fmt.Println("
1. Basic Replica Management")
	fmt.Println("----------------------------")

	// Create replica manager with default config
	rm := dbconn.NewReplicaManager(nil)

	// Register replicas
	rm.RegisterReplica("replica-1", 1)
	rm.RegisterReplica("replica-2", 1)
	rm.RegisterReplica("replica-3", 1)

	fmt.Printf("Registered 3 replicas
")

	// Select replicas using round-robin
	for i := 0; i < 5; i++ {
		replica, err := rm.SelectReplica()
		if err != nil {
			log.Printf("Error selecting replica: %v", err)
			continue
		}
		fmt.Printf("  Request %d -> %s
", i+1, replica)
	}

	// Get statistics
	stats := rm.GetStats()
	fmt.Printf("
Stats: %d total replicas, %d healthy
",
		stats.TotalReplicas, stats.HealthyReplicas)
}

func weightedLoadBalancing() {
	fmt.Println("
2. Weighted Load Balancing")
	fmt.Println("---------------------------")

	// Create config with weighted round-robin strategy
	config := dbconn.DefaultReplicaManagerConfig()
	config.Strategy = dbconn.WeightedRoundRobin

	rm := dbconn.NewReplicaManager(config)

	// Register replicas with different weights
	// Higher weight = more traffic
	rm.RegisterReplica("high-capacity-replica", 5)   // Gets 5x traffic
	rm.RegisterReplica("medium-capacity-replica", 2) // Gets 2x traffic
	rm.RegisterReplica("low-capacity-replica", 1)    // Gets 1x traffic

	fmt.Printf("Registered replicas with weights: 5, 2, 1
")

	// Distribute 16 requests
	distribution := make(map[string]int)
	for i := 0; i < 16; i++ {
		replica, err := rm.SelectReplica()
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		distribution[replica]++
	}

	fmt.Println("
Request distribution:")
	for replica, count := range distribution {
		fmt.Printf("  %s: %d requests
", replica, count)
	}
}

func replicaLagMonitoring() {
	fmt.Println("
3. Replica Lag Monitoring")
	fmt.Println("--------------------------")

	config := dbconn.DefaultReplicaManagerConfig()
	config.Strategy = dbconn.LeastLag
	config.MaxReplicaLag = 10.0 // 10 seconds max lag

	rm := dbconn.NewReplicaManager(config)

	// Register replicas
	rm.RegisterReplica("replica-east-1", 1)
	rm.RegisterReplica("replica-west-1", 1)
	rm.RegisterReplica("replica-central-1", 1)

	// Simulate lag information
	rm.UpdateReplicaLag("replica-east-1", 2.5, nil)    // 2.5s lag
	rm.UpdateReplicaLag("replica-west-1", 8.0, nil)    // 8s lag
	rm.UpdateReplicaLag("replica-central-1", 1.0, nil) // 1s lag (best)

	fmt.Println("Updated replication lag for all replicas")

	// Select replica (should prefer lowest lag)
	replica, err := rm.SelectReplica()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Selected replica: %s (has lowest lag)
", replica)

	// Display health information
	fmt.Println("
Replica Health:")
	health := rm.GetReplicaHealth()
	for name, h := range health {
		if h.Lag != nil {
			fmt.Printf("  %s: %.2fs lag, healthy=%v
",
				name, h.Lag.LagSeconds, h.Lag.IsHealthy)
		}
	}

	// Simulate replica with excessive lag
	rm.UpdateReplicaLag("replica-east-1", 15.0, nil) // Exceeds threshold
	fmt.Println("
After updating replica-east-1 lag to 15s:")

	health = rm.GetReplicaHealth()
	for name, h := range health {
		if h.Lag != nil {
			fmt.Printf("  %s: %.2fs lag, healthy=%v
",
				name, h.Lag.LagSeconds, h.Lag.IsHealthy)
		}
	}
}

func automaticFailover() {
	fmt.Println("
4. Automatic Failover")
	fmt.Println("----------------------")

	config := dbconn.DefaultReplicaManagerConfig()
	config.FailoverThreshold = 3 // Mark unhealthy after 3 failures

	rm := dbconn.NewReplicaManager(config)

	// Register replicas
	rm.RegisterReplica("replica-1", 1)
	rm.RegisterReplica("replica-2", 1)

	fmt.Printf("Registered 2 replicas with failover threshold = %d
",
		config.FailoverThreshold)

	// Simulate failures on replica-1
	fmt.Println("
Simulating failures on replica-1:")
	for i := 0; i < 3; i++ {
		rm.RecordFailure("replica-1", fmt.Errorf("connection timeout"))
		fmt.Printf("  Failure %d recorded
", i+1)

		health := rm.GetReplicaHealth()
		fmt.Printf("  replica-1 active: %v, consecutive fails: %d
",
			health["replica-1"].Active,
			health["replica-1"].ConsecutiveFails)
	}

	// Try to select replica
	fmt.Println("
Selecting replicas after failover:")
	for i := 0; i < 5; i++ {
		replica, err := rm.SelectReplica()
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		fmt.Printf("  Request %d -> %s
", i+1, replica)
	}

	stats := rm.GetStats()
	fmt.Printf("
Stats: %d healthy replicas (down from 2)
",
		stats.HealthyReplicas)

	// Simulate recovery
	fmt.Println("
Simulating recovery of replica-1:")
	rm.RecordSuccess("replica-1", 10*time.Millisecond)

	health := rm.GetReplicaHealth()
	fmt.Printf("  replica-1 consecutive fails: %d
",
		health["replica-1"].ConsecutiveFails)
	fmt.Printf("  replica-1 response time: %v
",
		health["replica-1"].ResponseTime)
}

// Example integration with actual database connections
func exampleWithRealDatabase() {
	fmt.Println("
5. Integration with Real Database")
	fmt.Println("----------------------------------")

	// Create replica manager
	config := dbconn.DefaultReplicaManagerConfig()
	config.Strategy = dbconn.WeightedRoundRobin
	config.MaxReplicaLag = 10.0
	config.LagCheckInterval = 5 * time.Second

	rm := dbconn.NewReplicaManager(config)

	// Register replicas
	rm.RegisterReplica("postgres-replica-1", 5)
	rm.RegisterReplica("postgres-replica-2", 3)

	// Database getter function (provides DB connection by name)
	dbGetter := func(name string) (*sql.DB, string, error) {
		// In real application, this would return actual database connections
		// For example, from a connection pool
		switch name {
		case "postgres-replica-1":
			// return db1, "postgres", nil
			return nil, "postgres", fmt.Errorf("example only")
		case "postgres-replica-2":
			// return db2, "postgres", nil
			return nil, "postgres", fmt.Errorf("example only")
		default:
			return nil, "", fmt.Errorf("unknown replica: %s", name)
		}
	}

	// Start background monitoring
	ctx := context.Background()
	rm.StartMonitoring(ctx, dbGetter)
	defer rm.StopMonitoring()

	fmt.Println("Started background monitoring:")
	fmt.Printf("  - Lag checks every %v
", config.LagCheckInterval)
	fmt.Printf("  - Max acceptable lag: %.1fs
", config.MaxReplicaLag)
	fmt.Printf("  - Auto-recovery: %v
", config.EnableAutoRecovery)

	// In a real application, you would now use the replica manager
	// to select replicas for read operations:
	/*
		for {
			replica, err := rm.SelectReplica()
			if err != nil {
				log.Printf("No healthy replicas: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// Get database connection
			db, _, err := dbGetter(replica)
			if err != nil {
				rm.RecordFailure(replica, err)
				continue
			}

			// Execute read query
			start := time.Now()
			err = executeReadQuery(db)
			duration := time.Since(start)

			if err != nil {
				rm.RecordFailure(replica, err)
			} else {
				rm.RecordSuccess(replica, duration)
			}
		}
	*/
}

// Example: Complete replica management setup
func exampleCompleteSetup() {
	fmt.Println("
6. Complete Setup Example")
	fmt.Println("--------------------------")

	// Step 1: Configure replica manager
	config := &dbconn.ReplicaManagerConfig{
		Strategy:              dbconn.WeightedRoundRobin,
		MaxReplicaLag:         5.0,
		LagCheckInterval:      10 * time.Second,
		FailoverThreshold:     3,
		RecoveryCheckInterval: 30 * time.Second,
		EnableAutoRecovery:    true,
		PreferLocalReplica:    true,
		LocalRegion:           "us-east-1",
	}

	rm := dbconn.NewReplicaManager(config)

	// Step 2: Register all replicas with weights
	replicas := []struct {
		name   string
		weight int
	}{
		{"postgres-primary-replica", 5},
		{"postgres-secondary-replica", 3},
		{"postgres-backup-replica", 1},
	}

	for _, r := range replicas {
		rm.RegisterReplica(r.name, r.weight)
		fmt.Printf("Registered: %s (weight: %d)
", r.name, r.weight)
	}

	// Step 3: Display configuration
	fmt.Println("
Configuration:")
	fmt.Printf("  Strategy: %s
", config.Strategy)
	fmt.Printf("  Max Lag: %.1fs
", config.MaxReplicaLag)
	fmt.Printf("  Failover Threshold: %d
", config.FailoverThreshold)
	fmt.Printf("  Auto Recovery: %v
", config.EnableAutoRecovery)

	// Step 4: Show initial stats
	stats := rm.GetStats()
	fmt.Println("
Initial Stats:")
	fmt.Printf("  Total Replicas: %d
", stats.TotalReplicas)
	fmt.Printf("  Healthy Replicas: %d
", stats.HealthyReplicas)

	fmt.Println("
Replica manager is ready for production use!")
}
