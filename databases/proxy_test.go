// Copyright 2023 IAC. All Rights Reserved.

package dbconn

import (
	"context"
	"testing"
	"time"
)

func TestNewDatabaseProxy(t *testing.T) {
	proxy := NewDatabaseProxy(nil)
	if proxy == nil {
		t.Fatal("NewDatabaseProxy returned nil")
	}

	if proxy.config == nil {
		t.Error("Config is nil")
	}

	if proxy.rewriter == nil {
		t.Error("Rewriter is nil")
	}
}

func TestProxyConfig(t *testing.T) {
	config := DefaultProxyConfig()

	if config.MaxConnections <= 0 {
		t.Error("MaxConnections should be positive")
	}

	if config.IdleTimeout <= 0 {
		t.Error("IdleTimeout should be positive")
	}

	if config.QueryTimeout <= 0 {
		t.Error("QueryTimeout should be positive")
	}
}

func TestIsSelectQuery(t *testing.T) {
	tests := []struct {
		query  string
		want   bool
	}{
		{"SELECT * FROM users", true},
		{"select id from orders", true},
		{"  SELECT name FROM products", true},
		{"SHOW TABLES", true},
		{"DESCRIBE users", true},
		{"INSERT INTO users VALUES (1)", false},
		{"UPDATE users SET name='test'", false},
		{"DELETE FROM users", false},
	}

	for _, tt := range tests {
		got := isSelectQuery(tt.query)
		if got != tt.want {
			t.Errorf("isSelectQuery(%q) = %v, want %v", tt.query, got, tt.want)
		}
	}
}

func TestQueryRewriter(t *testing.T) {
	qr := NewQueryRewriter()

	// Add a test rule
	qr.AddRule(RewriteRule{
		Name:    "test-rule",
		Pattern: "old_table",
		Replace: "new_table",
		Enabled: true,
	})

	query := "SELECT * FROM old_table"
	rewritten := qr.Rewrite(query)

	expected := "SELECT * FROM new_table"
	if rewritten != expected {
		t.Errorf("Expected %q, got %q", expected, rewritten)
	}
}

func TestQueryRewriter_DisabledRule(t *testing.T) {
	qr := NewQueryRewriter()

	qr.AddRule(RewriteRule{
		Name:    "disabled-rule",
		Pattern: "foo",
		Replace: "bar",
		Enabled: false,
	})

	query := "SELECT * FROM foo"
	rewritten := qr.Rewrite(query)

	if rewritten != query {
		t.Error("Disabled rule should not rewrite query")
	}
}

func TestProxyMetrics(t *testing.T) {
	proxy := NewDatabaseProxy(nil)

	// Simulate some queries
	for i := 0; i < 5; i++ {
		proxy.totalQueries++
		proxy.totalDuration += 1000 // 1ms each
	}

	proxy.slowQueries = 1
	proxy.failedQueries = 1

	metrics := proxy.GetMetrics()

	if metrics.TotalQueries != 5 {
		t.Errorf("Expected 5 total queries, got %d", metrics.TotalQueries)
	}

	if metrics.SlowQueries != 1 {
		t.Errorf("Expected 1 slow query, got %d", metrics.SlowQueries)
	}

	if metrics.FailedQueries != 1 {
		t.Errorf("Expected 1 failed query, got %d", metrics.FailedQueries)
	}
}

func TestLoadBalancer(t *testing.T) {
	lb := NewLoadBalancer(RoundRobin)

	// No databases initially
	_, err := lb.SelectDatabase()
	if err == nil {
		t.Error("Expected error when no databases available")
	}
}

func TestQueryLogger(t *testing.T) {
	logger := NewQueryLogger(10)

	// Log some queries
	logger.Log("SELECT 1", 10*time.Millisecond, nil)
	logger.Log("SELECT 2", 20*time.Millisecond, nil)
	logger.Log("SELECT 3", 30*time.Millisecond, nil)

	recent := logger.GetRecentQueries(2)
	if len(recent) != 2 {
		t.Errorf("Expected 2 recent queries, got %d", len(recent))
	}

	// Most recent should be SELECT 3
	if recent[1].Query != "SELECT 3" {
		t.Error("Wrong order for recent queries")
	}
}

func TestQueryLogger_Clear(t *testing.T) {
	logger := NewQueryLogger(10)

	logger.Log("SELECT 1", 10*time.Millisecond, nil)
	logger.Log("SELECT 2", 20*time.Millisecond, nil)

	logger.Clear()

	recent := logger.GetRecentQueries(10)
	if len(recent) != 0 {
		t.Error("Query log should be empty after clear")
	}
}

func TestQueryLogger_MaxLogs(t *testing.T) {
	logger := NewQueryLogger(5)

	// Log more than max
	for i := 0; i < 10; i++ {
		logger.Log("SELECT "+string(rune('0'+i)), time.Millisecond, nil)
	}

	logger.mu.RLock()
	count := len(logger.queries)
	logger.mu.RUnlock()

	if count > 5 {
		t.Errorf("Should keep only 5 logs, got %d", count)
	}
}

func TestConnectionPoolProxy(t *testing.T) {
	// Note: Requires actual DB connection for full test
	// This is a structure test
	cpp := NewConnectionPoolProxy(nil, 10)

	if cpp == nil {
		t.Fatal("NewConnectionPoolProxy returned nil")
	}

	if cpp.maxConns != 10 {
		t.Error("MaxConns not set correctly")
	}
}

func TestProxySetters(t *testing.T) {
	proxy := NewDatabaseProxy(nil)

	// Test setters don't panic
	proxy.SetPoolManager(nil)
	proxy.SetReplicaManager(nil)
	proxy.SetQueryCache(nil)

	// Verify they were set (even if nil)
	if proxy.poolManager != nil {
		t.Error("Expected nil pool manager")
	}
}

func TestRewriteRule(t *testing.T) {
	rule := RewriteRule{
		Name:    "test",
		Pattern: "foo",
		Replace: "bar",
		Enabled: true,
	}

	if rule.Name != "test" {
		t.Error("Rule name not set")
	}

	if !rule.Enabled {
		t.Error("Rule should be enabled")
	}
}

func TestProxyMetrics_AverageDuration(t *testing.T) {
	proxy := NewDatabaseProxy(nil)

	// No queries
	metrics := proxy.GetMetrics()
	if metrics.AverageDuration != 0 {
		t.Error("Average duration should be 0 when no queries")
	}

	// Add some query stats
	proxy.totalQueries = 4
	proxy.totalDuration = 4000 // 4ms total

	metrics = proxy.GetMetrics()
	expected := 1 * time.Millisecond
	if metrics.AverageDuration != expected {
		t.Errorf("Expected average %v, got %v", expected, metrics.AverageDuration)
	}
}

func TestLoadBalancer_Strategies(t *testing.T) {
	strategies := []LoadBalancingStrategy{
		RoundRobin,
		Random,
	}

	for _, strategy := range strategies {
		lb := NewLoadBalancer(strategy)
		if lb == nil {
			t.Errorf("Failed to create load balancer with strategy %s", strategy)
		}
	}
}
