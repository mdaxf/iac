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
	fmt.Println("\n1. Basic Replica Management")
	fmt.Println("----------------------------")

	// Create replica manager with default config
	rm := dbconn.NewReplicaManager(nil)

	// Register replicas
	rm.RegisterReplica("replica-1", 1)
	rm.RegisterReplica("replica-2", 1)
	rm.RegisterReplica("replica-3", 1)

	fmt.Printf("Registered 3 replicas\n")

	// Select replicas using round-robin
	for i := 0; i < 5; i++ {
		replica, err := rm.SelectReplica()
		if err != nil {
			log.Printf("Error selecting replica: %v", err)
			continue
		}
		fmt.Printf("  Request %d -> %s\n", i+1, replica)
	}

	// Get statistics
	stats := rm.GetStats()
	fmt.Printf("\nStats: %d total replicas, %d healthy\n",
		stats.TotalReplicas, stats.HealthyReplicas)
}

func weightedLoadBalancing() {
	fmt.Println("\n2. Weighted Load Balancing")
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

	fmt.Printf("Registered replicas with weights: 5, 2, 1\n")

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

	fmt.Println("\nRequest distribution:")
	for replica, count := range distribution {
		fmt.Printf("  %s: %d requests\n", replica, count)
	}
}

func replicaLagMonitoring() {
	fmt.Println("\n3. Replica Lag Monitoring")
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
	fmt.Printf("Selected replica: %s (has lowest lag)\n", replica)

	// Display health information
	fmt.Println("\nReplica Health:")
	health := rm.GetReplicaHealth()
	for name, h := range health {
		if h.Lag != nil {
			fmt.Printf("  %s: %.2fs lag, healthy=%v\n",
				name, h.Lag.LagSeconds, h.Lag.IsHealthy)
		}
	}

	// Simulate replica with excessive lag
	rm.UpdateReplicaLag("replica-east-1", 15.0, nil) // Exceeds threshold
	fmt.Println("\nAfter updating replica-east-1 lag to 15s:")

	health = rm.GetReplicaHealth()
	for name, h := range health {
		if h.Lag != nil {
			fmt.Printf("  %s: %.2fs lag, healthy=%v\n",
				name, h.Lag.LagSeconds, h.Lag.IsHealthy)
		}
	}
}

func automaticFailover() {
	fmt.Println("\n4. Automatic Failover")
	fmt.Println("----------------------")

	config := dbconn.DefaultReplicaManagerConfig()
	config.FailoverThreshold = 3 // Mark unhealthy after 3 failures

	rm := dbconn.NewReplicaManager(config)

	// Register replicas
	rm.RegisterReplica("replica-1", 1)
	rm.RegisterReplica("replica-2", 1)

	fmt.Printf("Registered 2 replicas with failover threshold = %d\n",
		config.FailoverThreshold)

	// Simulate failures on replica-1
	fmt.Println("\nSimulating failures on replica-1:")
	for i := 0; i < 3; i++ {
		rm.RecordFailure("replica-1", fmt.Errorf("connection timeout"))
		fmt.Printf("  Failure %d recorded\n", i+1)

		health := rm.GetReplicaHealth()
		fmt.Printf("  replica-1 active: %v, consecutive fails: %d\n",
			health["replica-1"].Active,
			health["replica-1"].ConsecutiveFails)
	}

	// Try to select replica
	fmt.Println("\nSelecting replicas after failover:")
	for i := 0; i < 5; i++ {
		replica, err := rm.SelectReplica()
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		fmt.Printf("  Request %d -> %s\n", i+1, replica)
	}

	stats := rm.GetStats()
	fmt.Printf("\nStats: %d healthy replicas (down from 2)\n",
		stats.HealthyReplicas)

	// Simulate recovery
	fmt.Println("\nSimulating recovery of replica-1:")
	rm.RecordSuccess("replica-1", 10*time.Millisecond)

	health := rm.GetReplicaHealth()
	fmt.Printf("  replica-1 consecutive fails: %d\n",
		health["replica-1"].ConsecutiveFails)
	fmt.Printf("  replica-1 response time: %v\n",
		health["replica-1"].ResponseTime)
}

// Example integration with actual database connections
func exampleWithRealDatabase() {
	fmt.Println("\n5. Integration with Real Database")
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
	fmt.Printf("  - Lag checks every %v\n", config.LagCheckInterval)
	fmt.Printf("  - Max acceptable lag: %.1fs\n", config.MaxReplicaLag)
	fmt.Printf("  - Auto-recovery: %v\n", config.EnableAutoRecovery)

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
	fmt.Println("\n6. Complete Setup Example")
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
		fmt.Printf("Registered: %s (weight: %d)\n", r.name, r.weight)
	}

	// Step 3: Display configuration
	fmt.Println("\nConfiguration:")
	fmt.Printf("  Strategy: %s\n", config.Strategy)
	fmt.Printf("  Max Lag: %.1fs\n", config.MaxReplicaLag)
	fmt.Printf("  Failover Threshold: %d\n", config.FailoverThreshold)
	fmt.Printf("  Auto Recovery: %v\n", config.EnableAutoRecovery)

	// Step 4: Show initial stats
	stats := rm.GetStats()
	fmt.Println("\nInitial Stats:")
	fmt.Printf("  Total Replicas: %d\n", stats.TotalReplicas)
	fmt.Printf("  Healthy Replicas: %d\n", stats.HealthyReplicas)

	fmt.Println("\nReplica manager is ready for production use!")
}
