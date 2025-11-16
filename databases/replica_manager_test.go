// Copyright 2023 IAC. All Rights Reserved.

package databases

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestNewReplicaManager(t *testing.T) {
	rm := NewReplicaManager(nil)
	if rm == nil {
		t.Fatal("NewReplicaManager returned nil")
	}

	if rm.config == nil {
		t.Error("Config is nil")
	}

	if rm.config.Strategy != WeightedRoundRobin {
		t.Errorf("Expected default strategy %s, got %s", WeightedRoundRobin, rm.config.Strategy)
	}
}

func TestRegisterReplica(t *testing.T) {
	rm := NewReplicaManager(nil)

	// Register replicas with different weights
	rm.RegisterReplica("replica1", 10)
	rm.RegisterReplica("replica2", 5)
	rm.RegisterReplica("replica3", 1)

	rm.mu.RLock()
	if len(rm.replicaHealth) != 3 {
		t.Errorf("Expected 3 replicas, got %d", len(rm.replicaHealth))
	}

	if rm.replicaHealth["replica1"].Weight != 10 {
		t.Errorf("Expected weight 10, got %d", rm.replicaHealth["replica1"].Weight)
	}

	// Check weighted pool
	if len(rm.weightedPool) != 16 { // 10 + 5 + 1
		t.Errorf("Expected weighted pool size 16, got %d", len(rm.weightedPool))
	}
	rm.mu.RUnlock()
}

func TestUnregisterReplica(t *testing.T) {
	rm := NewReplicaManager(nil)

	rm.RegisterReplica("replica1", 5)
	rm.RegisterReplica("replica2", 5)

	rm.UnregisterReplica("replica1")

	rm.mu.RLock()
	if len(rm.replicaHealth) != 1 {
		t.Errorf("Expected 1 replica after unregister, got %d", len(rm.replicaHealth))
	}

	if _, exists := rm.replicaHealth["replica1"]; exists {
		t.Error("replica1 should not exist after unregister")
	}
	rm.mu.RUnlock()
}

func TestSelectReplica_RoundRobin(t *testing.T) {
	config := DefaultReplicaManagerConfig()
	config.Strategy = RoundRobin
	rm := NewReplicaManager(config)

	rm.RegisterReplica("replica1", 1)
	rm.RegisterReplica("replica2", 1)
	rm.RegisterReplica("replica3", 1)

	// Test round-robin distribution
	selections := make(map[string]int)
	for i := 0; i < 9; i++ {
		selected, err := rm.SelectReplica()
		if err != nil {
			t.Fatalf("SelectReplica failed: %v", err)
		}
		selections[selected]++
	}

	// Each replica should be selected 3 times
	for replica, count := range selections {
		if count != 3 {
			t.Errorf("Replica %s selected %d times, expected 3", replica, count)
		}
	}
}

func TestSelectReplica_WeightedRoundRobin(t *testing.T) {
	config := DefaultReplicaManagerConfig()
	config.Strategy = WeightedRoundRobin
	rm := NewReplicaManager(config)

	rm.RegisterReplica("replica1", 3)
	rm.RegisterReplica("replica2", 1)

	// Test weighted distribution
	selections := make(map[string]int)
	for i := 0; i < 8; i++ {
		selected, err := rm.SelectReplica()
		if err != nil {
			t.Fatalf("SelectReplica failed: %v", err)
		}
		selections[selected]++
	}

	// replica1 should be selected ~3x more than replica2
	if selections["replica1"] == 0 || selections["replica2"] == 0 {
		t.Error("Both replicas should be selected at least once")
	}

	ratio := float64(selections["replica1"]) / float64(selections["replica2"])
	if ratio < 2.0 || ratio > 4.0 {
		t.Errorf("Expected ratio ~3.0, got %.2f (replica1=%d, replica2=%d)",
			ratio, selections["replica1"], selections["replica2"])
	}
}

func TestSelectReplica_Random(t *testing.T) {
	config := DefaultReplicaManagerConfig()
	config.Strategy = Random
	rm := NewReplicaManager(config)

	rm.RegisterReplica("replica1", 1)
	rm.RegisterReplica("replica2", 1)
	rm.RegisterReplica("replica3", 1)

	// Test random distribution
	selections := make(map[string]int)
	for i := 0; i < 100; i++ {
		selected, err := rm.SelectReplica()
		if err != nil {
			t.Fatalf("SelectReplica failed: %v", err)
		}
		selections[selected]++
	}

	// All replicas should be selected at least once
	for i := 1; i <= 3; i++ {
		replica := "replica" + string(rune('0'+i))
		if selections[replica] == 0 {
			t.Errorf("Replica %s was never selected", replica)
		}
	}
}

func TestSelectReplica_NoHealthyReplicas(t *testing.T) {
	rm := NewReplicaManager(nil)

	// Don't register any replicas
	_, err := rm.SelectReplica()
	if err == nil {
		t.Error("Expected error when no replicas available")
	}
}

func TestRecordSuccess(t *testing.T) {
	rm := NewReplicaManager(nil)
	rm.RegisterReplica("replica1", 5)

	// Simulate some failures first
	rm.RecordFailure("replica1", nil)
	rm.RecordFailure("replica1", nil)

	rm.mu.RLock()
	health := rm.replicaHealth["replica1"]
	failCount := health.ConsecutiveFails
	rm.mu.RUnlock()

	if failCount != 2 {
		t.Errorf("Expected 2 consecutive fails, got %d", failCount)
	}

	// Record success
	rm.RecordSuccess("replica1", 10*time.Millisecond)

	rm.mu.RLock()
	health = rm.replicaHealth["replica1"]
	rm.mu.RUnlock()

	if health.ConsecutiveFails != 0 {
		t.Errorf("Expected consecutive fails to reset to 0, got %d", health.ConsecutiveFails)
	}

	if health.ResponseTime != 10*time.Millisecond {
		t.Errorf("Expected response time 10ms, got %v", health.ResponseTime)
	}
}

func TestRecordFailure(t *testing.T) {
	config := DefaultReplicaManagerConfig()
	config.FailoverThreshold = 3
	rm := NewReplicaManager(config)

	rm.RegisterReplica("replica1", 5)

	// Replica should be active initially
	rm.mu.RLock()
	if !rm.replicaHealth["replica1"].Active {
		t.Error("Replica should be active initially")
	}
	rm.mu.RUnlock()

	// Record failures below threshold
	rm.RecordFailure("replica1", nil)
	rm.RecordFailure("replica1", nil)

	rm.mu.RLock()
	if !rm.replicaHealth["replica1"].Active {
		t.Error("Replica should still be active")
	}
	rm.mu.RUnlock()

	// Record one more failure to exceed threshold
	rm.RecordFailure("replica1", nil)

	rm.mu.RLock()
	if rm.replicaHealth["replica1"].Active {
		t.Error("Replica should be inactive after exceeding threshold")
	}
	if rm.replicaHealth["replica1"].ConsecutiveFails != 3 {
		t.Errorf("Expected 3 consecutive fails, got %d", rm.replicaHealth["replica1"].ConsecutiveFails)
	}
	rm.mu.RUnlock()
}

func TestUpdateReplicaLag(t *testing.T) {
	config := DefaultReplicaManagerConfig()
	config.MaxReplicaLag = 10.0
	rm := NewReplicaManager(config)

	rm.RegisterReplica("replica1", 1)

	// Update with healthy lag
	rm.UpdateReplicaLag("replica1", 5.0, nil)

	rm.mu.RLock()
	health := rm.replicaHealth["replica1"]
	rm.mu.RUnlock()

	if health.Lag == nil {
		t.Fatal("Lag should be set")
	}

	if health.Lag.LagSeconds != 5.0 {
		t.Errorf("Expected lag 5.0, got %.2f", health.Lag.LagSeconds)
	}

	if !health.Lag.IsHealthy {
		t.Error("Lag should be healthy (below threshold)")
	}

	// Update with unhealthy lag
	rm.UpdateReplicaLag("replica1", 15.0, nil)

	rm.mu.RLock()
	health = rm.replicaHealth["replica1"]
	rm.mu.RUnlock()

	if health.Lag.IsHealthy {
		t.Error("Lag should be unhealthy (above threshold)")
	}
}

func TestSelectReplica_WithLag(t *testing.T) {
	config := DefaultReplicaManagerConfig()
	config.Strategy = LeastLag
	config.MaxReplicaLag = 10.0
	rm := NewReplicaManager(config)

	rm.RegisterReplica("replica1", 1)
	rm.RegisterReplica("replica2", 1)
	rm.RegisterReplica("replica3", 1)

	// Set different lag values
	rm.UpdateReplicaLag("replica1", 2.0, nil)
	rm.UpdateReplicaLag("replica2", 5.0, nil)
	rm.UpdateReplicaLag("replica3", 1.0, nil)

	// Should select replica3 (lowest lag)
	selected, err := rm.SelectReplica()
	if err != nil {
		t.Fatalf("SelectReplica failed: %v", err)
	}

	if selected != "replica3" {
		t.Errorf("Expected replica3 (lowest lag), got %s", selected)
	}
}

func TestSelectReplica_ExceedsMaxLag(t *testing.T) {
	config := DefaultReplicaManagerConfig()
	config.MaxReplicaLag = 10.0
	rm := NewReplicaManager(config)

	rm.RegisterReplica("replica1", 1)
	rm.RegisterReplica("replica2", 1)

	// Set lag above threshold for replica1
	rm.UpdateReplicaLag("replica1", 15.0, nil)
	rm.UpdateReplicaLag("replica2", 5.0, nil)

	// Should only select replica2
	selections := make(map[string]int)
	for i := 0; i < 10; i++ {
		selected, err := rm.SelectReplica()
		if err != nil {
			t.Fatalf("SelectReplica failed: %v", err)
		}
		selections[selected]++
	}

	if selections["replica1"] > 0 {
		t.Error("replica1 should not be selected (lag too high)")
	}

	if selections["replica2"] != 10 {
		t.Errorf("replica2 should be selected 10 times, got %d", selections["replica2"])
	}
}

func TestGetReplicaHealth(t *testing.T) {
	rm := NewReplicaManager(nil)

	rm.RegisterReplica("replica1", 5)
	rm.RegisterReplica("replica2", 3)

	rm.UpdateReplicaLag("replica1", 2.0, nil)
	rm.RecordSuccess("replica1", 10*time.Millisecond)

	health := rm.GetReplicaHealth()

	if len(health) != 2 {
		t.Errorf("Expected 2 replica health records, got %d", len(health))
	}

	if health["replica1"].Weight != 5 {
		t.Errorf("Expected weight 5, got %d", health["replica1"].Weight)
	}

	if health["replica1"].Lag == nil {
		t.Error("Expected lag info for replica1")
	}

	if health["replica1"].Lag.LagSeconds != 2.0 {
		t.Errorf("Expected lag 2.0, got %.2f", health["replica1"].Lag.LagSeconds)
	}
}

func TestGetHealthyReplicaCount(t *testing.T) {
	rm := NewReplicaManager(nil)

	rm.RegisterReplica("replica1", 1)
	rm.RegisterReplica("replica2", 1)
	rm.RegisterReplica("replica3", 1)

	count := rm.GetHealthyReplicaCount()
	if count != 3 {
		t.Errorf("Expected 3 healthy replicas, got %d", count)
	}

	// Mark one as inactive
	rm.mu.Lock()
	rm.replicaHealth["replica1"].Active = false
	rm.mu.Unlock()

	count = rm.GetHealthyReplicaCount()
	if count != 2 {
		t.Errorf("Expected 2 healthy replicas after marking one inactive, got %d", count)
	}
}

func TestGetStats(t *testing.T) {
	rm := NewReplicaManager(nil)

	rm.RegisterReplica("replica1", 1)
	rm.RegisterReplica("replica2", 1)

	stats := rm.GetStats()

	if stats.TotalReplicas != 2 {
		t.Errorf("Expected 2 total replicas, got %d", stats.TotalReplicas)
	}

	if stats.HealthyReplicas != 2 {
		t.Errorf("Expected 2 healthy replicas, got %d", stats.HealthyReplicas)
	}

	if stats.Strategy != string(WeightedRoundRobin) {
		t.Errorf("Expected strategy %s, got %s", WeightedRoundRobin, stats.Strategy)
	}
}

func TestCheckReplicationLag_MySQL(t *testing.T) {
	// This is a unit test structure - would need actual MySQL connection for integration test
	rm := NewReplicaManager(nil)

	// Test that the function handles nil DB gracefully
	_, err := rm.CheckReplicationLag(context.Background(), nil, "mysql")
	if err == nil {
		t.Error("Expected error with nil DB")
	}
}

func TestCheckReplicationLag_PostgreSQL(t *testing.T) {
	// This is a unit test structure - would need actual PostgreSQL connection for integration test
	rm := NewReplicaManager(nil)

	// Test that the function handles nil DB gracefully
	_, err := rm.CheckReplicationLag(context.Background(), nil, "postgres")
	if err == nil {
		t.Error("Expected error with nil DB")
	}
}

func TestCheckReplicationLag_UnsupportedDB(t *testing.T) {
	rm := NewReplicaManager(nil)

	_, err := rm.CheckReplicationLag(context.Background(), &sql.DB{}, "unsupported")
	if err == nil {
		t.Error("Expected error for unsupported database type")
	}
}

func TestStopMonitoring(t *testing.T) {
	rm := NewReplicaManager(nil)

	// Start monitoring
	ctx := context.Background()
	dbGetter := func(name string) (*sql.DB, string, error) {
		return nil, "", nil
	}

	rm.StartMonitoring(ctx, dbGetter)

	if !rm.monitoringActive {
		t.Error("Monitoring should be active after start")
	}

	// Stop monitoring
	rm.StopMonitoring()

	rm.mu.RLock()
	active := rm.monitoringActive
	rm.mu.RUnlock()

	if active {
		t.Error("Monitoring should be inactive after stop")
	}
}

func TestLoadBalancingStrategies(t *testing.T) {
	strategies := []LoadBalancingStrategy{
		RoundRobin,
		WeightedRoundRobin,
		Random,
		LeastLag,
	}

	for _, strategy := range strategies {
		config := DefaultReplicaManagerConfig()
		config.Strategy = strategy
		rm := NewReplicaManager(config)

		rm.RegisterReplica("replica1", 1)
		rm.RegisterReplica("replica2", 1)

		// Should be able to select with any strategy
		_, err := rm.SelectReplica()
		if err != nil {
			t.Errorf("Strategy %s failed to select replica: %v", strategy, err)
		}
	}
}

func TestWeightAdjustment(t *testing.T) {
	rm := NewReplicaManager(nil)
	rm.RegisterReplica("replica1", 5)

	rm.mu.RLock()
	initialWeight := rm.replicaHealth["replica1"].EffectiveWeight
	rm.mu.RUnlock()

	if initialWeight != 5 {
		t.Errorf("Expected initial effective weight 5, got %d", initialWeight)
	}

	// Record failures
	for i := 0; i < 3; i++ {
		rm.RecordFailure("replica1", nil)
	}

	rm.mu.RLock()
	reducedWeight := rm.replicaHealth["replica1"].EffectiveWeight
	rm.mu.RUnlock()

	if reducedWeight >= initialWeight {
		t.Error("Effective weight should be reduced after failures")
	}

	// Record success to recover
	rm.RecordSuccess("replica1", 10*time.Millisecond)

	rm.mu.RLock()
	recoveredWeight := rm.replicaHealth["replica1"].EffectiveWeight
	rm.mu.RUnlock()

	if recoveredWeight <= reducedWeight {
		t.Error("Effective weight should increase after success")
	}
}
