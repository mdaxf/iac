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

package config

import (
	"fmt"
	"os"
	"time"

	"github.com/mdaxf/iac/services/session"
)

// DefaultClusterConfig returns the default cluster configuration
func DefaultClusterConfig() *ClusterConfig {
	return &ClusterConfig{
		Enabled:      false,
		Mode:         "standalone",
		InstanceName: getDefaultInstanceName(),
		Environment:  "development",
		Registry:     DefaultRegistryConfig(),
		ConfigStore:  DefaultConfigStoreConfig(),
		Session:      DefaultClusterSessionConfig(),
	}
}

// DefaultRegistryConfig returns the default registry configuration
func DefaultRegistryConfig() *RegistryConfig {
	return &RegistryConfig{
		Collection:        "iac_instance_registry",
		HeartbeatInterval: 30,
		HeartbeatTTL:      90,
		CleanupInterval:   300,
		SignalRDiscovery:  true,
	}
}

// DefaultConfigStoreConfig returns the default config store configuration
func DefaultConfigStoreConfig() *ConfigStoreConfig {
	return &ConfigStoreConfig{
		Enabled:      false,
		Collection:   "iac_configurations",
		WatchEnabled: true,
		AutoReload:   true,
	}
}

// DefaultClusterSessionConfig returns the default cluster session configuration
func DefaultClusterSessionConfig() *ClusterSessionConfig {
	return &ClusterSessionConfig{
		Distributed:     false,
		Backend:         "memory",
		FallbackBackend: "memory",
		TTL:             1800, // 30 minutes
		LockTTL:         30,
		AutoExtend:      true,
		Redis: &RedisSessionConfig{
			URL:       "localhost:6379",
			DB:        0,
			Password:  "",
			PoolSize:  10,
			KeyPrefix: "iac:session",
		},
		MongoDB: &MongoSessionConfig{
			Collection: "iac_sessions",
			Database:   "",
		},
	}
}

// getDefaultInstanceName generates a default instance name
func getDefaultInstanceName() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("iac-%s-%d", hostname, os.Getpid())
}

// GetClusterConfig returns the cluster configuration from global configuration
func GetClusterConfig() *ClusterConfig {
	if GlobalConfiguration == nil || GlobalConfiguration.Cluster == nil {
		return DefaultClusterConfig()
	}
	return GlobalConfiguration.Cluster
}

// IsClusterEnabled checks if cluster mode is enabled
func IsClusterEnabled() bool {
	config := GetClusterConfig()
	return config.Enabled
}

// IsDistributedSessionEnabled checks if distributed sessions are enabled
func IsDistributedSessionEnabled() bool {
	config := GetClusterConfig()
	if config.Session == nil {
		return false
	}
	return config.Session.Distributed
}

// GetInstanceName returns the configured instance name
func GetInstanceName() string {
	config := GetClusterConfig()
	if config.InstanceName == "" {
		return getDefaultInstanceName()
	}
	return config.InstanceName
}

// GetEnvironment returns the configured environment
func GetEnvironment() string {
	config := GetClusterConfig()
	if config.Environment == "" {
		return "development"
	}
	return config.Environment
}

// ToDistributedSessionConfig converts cluster session config to service config
func (c *ClusterSessionConfig) ToDistributedSessionConfig(instanceID string) *session.DistributedSessionConfig {
	if c == nil {
		return session.DefaultDistributedSessionConfig()
	}

	config := &session.DistributedSessionConfig{
		Enabled:         c.Distributed,
		BackendType:     session.SessionBackendType(c.Backend),
		FallbackType:    session.SessionBackendType(c.FallbackBackend),
		SessionTTL:      time.Duration(c.TTL) * time.Second,
		TokenTTL:        time.Duration(c.TTL) * time.Second,
		LockTTL:         time.Duration(c.LockTTL) * time.Second,
		AutoExtend:      c.AutoExtend,
		ExtendThreshold: 5 * time.Minute,
		InstanceID:      instanceID,
		BackendConfig:   c.ToBackendConfig(),
	}

	return config
}

// ToBackendConfig converts cluster session config to backend config
func (c *ClusterSessionConfig) ToBackendConfig() *session.SessionBackendConfig {
	if c == nil {
		return session.DefaultSessionBackendConfig()
	}

	config := session.DefaultSessionBackendConfig()
	config.TTL = time.Duration(c.TTL) * time.Second
	config.LockTTL = time.Duration(c.LockTTL) * time.Second

	if c.Redis != nil {
		config.RedisURL = c.Redis.URL
		config.RedisDB = c.Redis.DB
		config.RedisPassword = c.Redis.Password
		config.RedisPoolSize = c.Redis.PoolSize
		config.RedisKeyPrefix = c.Redis.KeyPrefix
	}

	if c.MongoDB != nil {
		config.MongoCollection = c.MongoDB.Collection
		config.MongoDatabase = c.MongoDB.Database
	}

	return config
}

// ValidateClusterConfig validates the cluster configuration
func ValidateClusterConfig(config *ClusterConfig) error {
	if config == nil {
		return nil
	}

	if config.Enabled {
		if config.Mode != "standalone" && config.Mode != "distributed" {
			return fmt.Errorf("invalid cluster mode: %s (must be 'standalone' or 'distributed')", config.Mode)
		}

		if config.InstanceName == "" {
			return fmt.Errorf("instance_name is required when cluster is enabled")
		}

		if config.Session != nil && config.Session.Distributed {
			if config.Session.Backend == "" {
				return fmt.Errorf("session backend is required when distributed sessions are enabled")
			}

			validBackends := map[string]bool{
				"redis":   true,
				"mongodb": true,
				"memory":  true,
			}
			if !validBackends[config.Session.Backend] {
				return fmt.Errorf("invalid session backend: %s", config.Session.Backend)
			}
		}
	}

	return nil
}
