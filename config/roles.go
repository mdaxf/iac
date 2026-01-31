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
)

// InstanceRoleType represents the type of role an instance can have
type InstanceRoleType string

const (
	// RoleApp handles the application API endpoint features (controllers, portal, static files)
	RoleApp InstanceRoleType = "app"

	// RoleIntegrationHub handles integration hub execution
	RoleIntegrationHub InstanceRoleType = "integration_hub"

	// RoleSignalR handles SignalR server features
	RoleSignalR InstanceRoleType = "signalr"

	// RoleJobExecutor executes jobs from job queue
	RoleJobExecutor InstanceRoleType = "job_executor"
)

// InstanceRolesConfig holds the configuration for all instance roles
type InstanceRolesConfig struct {
	// Roles is a list of enabled role configurations
	Roles []RoleConfig `json:"roles"`
}

// RoleConfig holds configuration for a single role
type RoleConfig struct {
	// Type is the role type (main, integration_hub, signalr, job_executor)
	Type InstanceRoleType `json:"type"`

	// Enabled indicates if this role is active (default: true if present)
	Enabled bool `json:"enabled"`

	// App role configuration
	App *AppRoleConfig `json:"app,omitempty"`

	// Integration hub role configuration
	IntegrationHub *IntegrationHubRoleConfig `json:"integration_hub,omitempty"`

	// SignalR role configuration
	SignalR *SignalRRoleConfig `json:"signalr,omitempty"`

	// Job executor role configuration
	JobExecutor *JobExecutorRoleConfig `json:"job_executor,omitempty"`
}

// AppRoleConfig holds configuration for the application API role
type AppRoleConfig struct {
	// Port for the API server (overrides apiconfig.json if set)
	Port int `json:"port,omitempty"`

	// EnablePortal enables serving the portal UI
	EnablePortal bool `json:"enable_portal"`

	// EnableStaticFiles enables serving static files
	EnableStaticFiles bool `json:"enable_static_files"`

	// Controllers to load (if empty, loads all from apiconfig.json)
	Controllers []string `json:"controllers,omitempty"`
}

// IntegrationHubRoleConfig holds configuration for the integration hub role
type IntegrationHubRoleConfig struct {
	// HubNames is a list of integration hub names to run
	// If empty, runs all enabled default hubs from the database
	HubNames []string `json:"hub_names,omitempty"`

	// AutoStart starts the hub executors automatically on startup
	AutoStart bool `json:"auto_start"`

	// EnableAPI enables the hub executor management API endpoints
	EnableAPI bool `json:"enable_api"`
}

// SignalRRoleConfig holds configuration for the SignalR server role
type SignalRRoleConfig struct {
	// Port for the SignalR server
	Port int `json:"port"`

	// Hub is the SignalR hub path (e.g., "/iacmessagebus")
	Hub string `json:"hub"`

	// AllowedOrigins for CORS
	AllowedOrigins []string `json:"allowed_origins,omitempty"`

	// MaxConnections limits concurrent connections (0 = unlimited)
	MaxConnections int `json:"max_connections,omitempty"`

	// KeepAliveInterval in seconds
	KeepAliveInterval int `json:"keep_alive_interval,omitempty"`

	// TimeoutInterval in seconds for connection timeout
	TimeoutInterval int `json:"timeout_interval,omitempty"`

	// EnableWebSocket enables WebSocket transport
	EnableWebSocket bool `json:"enable_websocket"`

	// EnableSSE enables Server-Sent Events transport
	EnableSSE bool `json:"enable_sse"`

	// InsecureSkipVerify skips TLS certificate verification (for testing only)
	InsecureSkipVerify bool `json:"insecure_skip_verify,omitempty"`

	// TLS configuration
	TLS *TLSConfig `json:"tls,omitempty"`
}

// TLSConfig holds TLS/SSL configuration
type TLSConfig struct {
	// Enabled indicates if TLS is enabled
	Enabled bool `json:"enabled"`

	// CertFile is the path to the certificate file
	CertFile string `json:"cert_file"`

	// KeyFile is the path to the key file
	KeyFile string `json:"key_file"`
}

// JobExecutorRoleConfig holds configuration for the job executor role
type JobExecutorRoleConfig struct {
	// Workers is the number of concurrent job workers
	Workers int `json:"workers"`

	// PollInterval is the interval in seconds to poll for new jobs
	PollInterval int `json:"poll_interval"`

	// MaxRetries is the maximum number of retries for failed jobs
	MaxRetries int `json:"max_retries"`

	// EnableScheduler enables the cron-based job scheduler
	EnableScheduler bool `json:"enable_scheduler"`

	// SchedulerCheckInterval in seconds
	SchedulerCheckInterval int `json:"scheduler_check_interval"`

	// JobTypes is a list of job types this executor will handle (empty = all)
	JobTypes []string `json:"job_types,omitempty"`

	// UseRedis enables Redis-based distributed job queue
	UseRedis bool `json:"use_redis"`

	// EnableMetrics enables job execution metrics
	EnableMetrics bool `json:"enable_metrics"`
}

// Validate validates the instance roles configuration
func (c *InstanceRolesConfig) Validate() error {
	if len(c.Roles) == 0 {
		return fmt.Errorf("at least one role must be configured")
	}

	enabledRoles := 0
	for _, role := range c.Roles {
		if role.Enabled {
			enabledRoles++

			// Validate role-specific configuration
			if err := role.Validate(); err != nil {
				return fmt.Errorf("role %s validation failed: %v", role.Type, err)
			}
		}
	}

	if enabledRoles == 0 {
		return fmt.Errorf("at least one role must be enabled")
	}

	return nil
}

// Validate validates a single role configuration
func (r *RoleConfig) Validate() error {
	switch r.Type {
	case RoleApp:
		// App role has optional configuration
		return nil

	case RoleIntegrationHub:
		if r.IntegrationHub == nil {
			return fmt.Errorf("integration_hub configuration is required")
		}
		return nil

	case RoleSignalR:
		if r.SignalR == nil {
			return fmt.Errorf("signalr configuration is required")
		}
		if r.SignalR.Port == 0 {
			return fmt.Errorf("signalr port is required")
		}
		if r.SignalR.Hub == "" {
			return fmt.Errorf("signalr hub path is required")
		}
		return nil

	case RoleJobExecutor:
		if r.JobExecutor == nil {
			return fmt.Errorf("job_executor configuration is required")
		}
		if r.JobExecutor.Workers <= 0 {
			r.JobExecutor.Workers = 5 // Default workers
		}
		return nil

	default:
		return fmt.Errorf("unknown role type: %s", r.Type)
	}
}

// HasRole checks if a specific role is enabled
func (c *InstanceRolesConfig) HasRole(roleType InstanceRoleType) bool {
	for _, role := range c.Roles {
		if role.Type == roleType && role.Enabled {
			return true
		}
	}
	return false
}

// GetRole returns the configuration for a specific role
func (c *InstanceRolesConfig) GetRole(roleType InstanceRoleType) *RoleConfig {
	for i := range c.Roles {
		if c.Roles[i].Type == roleType && c.Roles[i].Enabled {
			return &c.Roles[i]
		}
	}
	return nil
}

// GetEnabledRoles returns a list of enabled role types
func (c *InstanceRolesConfig) GetEnabledRoles() []InstanceRoleType {
	roles := make([]InstanceRoleType, 0)
	for _, role := range c.Roles {
		if role.Enabled {
			roles = append(roles, role.Type)
		}
	}
	return roles
}

// DefaultRolesConfig returns a default configuration with the app role enabled
func DefaultRolesConfig() *InstanceRolesConfig {
	return &InstanceRolesConfig{
		Roles: []RoleConfig{
			{
				Type:    RoleApp,
				Enabled: true,
				App: &AppRoleConfig{
					EnablePortal:      true,
					EnableStaticFiles: true,
				},
			},
			{
				Type:    RoleJobExecutor,
				Enabled: true,
				JobExecutor: &JobExecutorRoleConfig{
					Workers:                5,
					PollInterval:           5,
					MaxRetries:             3,
					EnableScheduler:        true,
					SchedulerCheckInterval: 60,
					EnableMetrics:          true,
				},
			},
		},
	}
}
