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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RoleWorker represents a worker running a specific role
type RoleWorker struct {
	RoleType   InstanceRoleType
	Running    bool
	StartedAt  time.Time
	Error      error
	cancelFunc context.CancelFunc
	done       chan struct{}
}

// RoleInitializer handles initialization of services based on configured roles
type RoleInitializer struct {
	config      *InstanceRolesConfig
	mu          sync.RWMutex
	initialized bool
	workers     map[InstanceRoleType]*RoleWorker
	wg          sync.WaitGroup

	// Callbacks for role-specific initialization
	onAppInit            func(ctx context.Context, config *AppRoleConfig, router *gin.Engine) error
	onIntegrationHubInit func(ctx context.Context, config *IntegrationHubRoleConfig, router *gin.Engine) error
	onSignalRInit        func(ctx context.Context, config *SignalRRoleConfig) error
	onJobExecutorInit    func(ctx context.Context, config *JobExecutorRoleConfig) error

	// Callbacks for role-specific shutdown
	onAppShutdown            func(ctx context.Context) error
	onIntegrationHubShutdown func(ctx context.Context) error
	onSignalRShutdown        func(ctx context.Context) error
	onJobExecutorShutdown    func(ctx context.Context) error
}

// NewRoleInitializer creates a new role initializer
func NewRoleInitializer(config *InstanceRolesConfig) *RoleInitializer {
	return &RoleInitializer{
		config:      config,
		initialized: false,
		workers:     make(map[InstanceRoleType]*RoleWorker),
	}
}

// SetAppInitializer sets the callback for app role initialization
func (ri *RoleInitializer) SetAppInitializer(init func(ctx context.Context, config *AppRoleConfig, router *gin.Engine) error, shutdown func(ctx context.Context) error) {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	ri.onAppInit = init
	ri.onAppShutdown = shutdown
}

// SetIntegrationHubInitializer sets the callback for integration hub role initialization
func (ri *RoleInitializer) SetIntegrationHubInitializer(init func(ctx context.Context, config *IntegrationHubRoleConfig, router *gin.Engine) error, shutdown func(ctx context.Context) error) {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	ri.onIntegrationHubInit = init
	ri.onIntegrationHubShutdown = shutdown
}

// SetSignalRInitializer sets the callback for SignalR role initialization
func (ri *RoleInitializer) SetSignalRInitializer(init func(ctx context.Context, config *SignalRRoleConfig) error, shutdown func(ctx context.Context) error) {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	ri.onSignalRInit = init
	ri.onSignalRShutdown = shutdown
}

// SetJobExecutorInitializer sets the callback for job executor role initialization
func (ri *RoleInitializer) SetJobExecutorInitializer(init func(ctx context.Context, config *JobExecutorRoleConfig) error, shutdown func(ctx context.Context) error) {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	ri.onJobExecutorInit = init
	ri.onJobExecutorShutdown = shutdown
}

// Initialize initializes all enabled roles in separate goroutines
func (ri *RoleInitializer) Initialize(ctx context.Context, router *gin.Engine) error {
	ri.mu.Lock()
	if ri.initialized {
		ri.mu.Unlock()
		return fmt.Errorf("role initializer already initialized")
	}

	if ri.config == nil {
		// Use default configuration if none provided
		ri.config = DefaultRolesConfig()
	}

	// Validate configuration
	if err := ri.config.Validate(); err != nil {
		ri.mu.Unlock()
		return fmt.Errorf("roles configuration validation failed: %v", err)
	}

	enabledRoles := ri.config.GetEnabledRoles()
	roles := ri.config.Roles
	ri.mu.Unlock() // Release lock before starting goroutines

	fmt.Printf("Initializing instance with roles: %v\n", enabledRoles)

	// Channel to collect initialization errors
	errChan := make(chan error, len(roles))
	var initWg sync.WaitGroup

	// Initialize each enabled role in a separate goroutine
	for _, role := range roles {
		if !role.Enabled {
			continue
		}

		roleCopy := role // Create a copy for the goroutine
		initWg.Add(1)

		go func(role RoleConfig) {
			defer initWg.Done()

			fmt.Printf("  - Starting role worker: %s\n", role.Type)

			// Create cancellable context for this role
			roleCtx, cancel := context.WithCancel(ctx)

			// Create role worker
			worker := &RoleWorker{
				RoleType:   role.Type,
				Running:    false,
				StartedAt:  time.Now(),
				cancelFunc: cancel,
				done:       make(chan struct{}),
			}

			ri.mu.Lock()
			ri.workers[role.Type] = worker
			ri.mu.Unlock()

			var initErr error

			switch role.Type {
			case RoleApp:
				fmt.Printf("    Processing app role, callback registered: %v\n", ri.onAppInit != nil)
				if ri.onAppInit != nil {
					appConfig := role.App
					if appConfig == nil {
						fmt.Println("    App config is nil, using defaults")
						appConfig = &AppRoleConfig{
							EnablePortal:      true,
							EnableStaticFiles: true,
						}
					} else {
						fmt.Printf("    App config: portal=%v, static=%v\n", appConfig.EnablePortal, appConfig.EnableStaticFiles)
					}
					initErr = ri.onAppInit(roleCtx, appConfig, router)
					fmt.Printf("    App role init completed, error: %v\n", initErr)
				} else {
					fmt.Println("    WARNING: App role callback not registered!")
				}

			case RoleIntegrationHub:
				if ri.onIntegrationHubInit != nil && role.IntegrationHub != nil {
					initErr = ri.onIntegrationHubInit(roleCtx, role.IntegrationHub, router)
				}

			case RoleSignalR:
				if ri.onSignalRInit != nil && role.SignalR != nil {
					initErr = ri.onSignalRInit(roleCtx, role.SignalR)
				}

			case RoleJobExecutor:
				fmt.Printf("    Processing job_executor role, callback registered: %v\n", ri.onJobExecutorInit != nil)
				if ri.onJobExecutorInit != nil && role.JobExecutor != nil {
					initErr = ri.onJobExecutorInit(roleCtx, role.JobExecutor)
				}

			default:
				fmt.Printf("    WARNING: Unknown role type: '%s' (len=%d)\n", role.Type, len(role.Type))
			}

			if initErr != nil {
				worker.Error = initErr
				errChan <- fmt.Errorf("failed to initialize %s role: %v", role.Type, initErr)
				fmt.Printf("    Role %s initialization failed: %v\n", role.Type, initErr)
				return
			}

			worker.Running = true
			fmt.Printf("    Role %s worker started successfully\n", role.Type)

			// Start the role's main loop in a goroutine
			ri.wg.Add(1)
			go func() {
				defer ri.wg.Done()
				defer close(worker.done)

				// Wait for context cancellation (shutdown signal)
				<-roleCtx.Done()

				fmt.Printf("    Role %s worker received shutdown signal\n", role.Type)
				worker.Running = false
			}()
		}(roleCopy)
	}

	// Wait for all role initializations to complete
	initWg.Wait()
	close(errChan)

	// Collect any errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("role initialization errors: %v", errors)
	}

	ri.mu.Lock()
	ri.initialized = true
	ri.mu.Unlock()
	fmt.Println("All role workers initialized and running")

	return nil
}

// Shutdown shuts down all initialized roles
func (ri *RoleInitializer) Shutdown(ctx context.Context) error {
	ri.mu.Lock()

	if !ri.initialized {
		ri.mu.Unlock()
		return nil
	}

	fmt.Println("Shutting down role workers...")

	var errors []error

	// Shutdown in reverse order of initialization
	for i := len(ri.config.Roles) - 1; i >= 0; i-- {
		role := ri.config.Roles[i]
		if !role.Enabled {
			continue
		}

		worker, exists := ri.workers[role.Type]
		if !exists {
			continue
		}

		fmt.Printf("  - Shutting down role worker: %s\n", role.Type)

		// Cancel the role's context
		if worker.cancelFunc != nil {
			worker.cancelFunc()
		}

		// Call role-specific shutdown callback
		var err error
		switch role.Type {
		case RoleApp:
			if ri.onAppShutdown != nil {
				err = ri.onAppShutdown(ctx)
			}

		case RoleIntegrationHub:
			if ri.onIntegrationHubShutdown != nil {
				err = ri.onIntegrationHubShutdown(ctx)
			}

		case RoleSignalR:
			if ri.onSignalRShutdown != nil {
				err = ri.onSignalRShutdown(ctx)
			}

		case RoleJobExecutor:
			if ri.onJobExecutorShutdown != nil {
				err = ri.onJobExecutorShutdown(ctx)
			}
		}

		if err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown %s: %v", role.Type, err))
		}

		// Wait for the worker's done channel with timeout
		if worker.done != nil {
			select {
			case <-worker.done:
				fmt.Printf("    Role %s worker stopped\n", role.Type)
			case <-time.After(5 * time.Second):
				fmt.Printf("    Role %s worker shutdown timeout\n", role.Type)
			}
		}
	}

	ri.initialized = false
	ri.mu.Unlock()

	// Wait for all role goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		ri.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("All role workers shut down successfully")
	case <-time.After(10 * time.Second):
		fmt.Println("Timeout waiting for role workers to shut down")
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	return nil
}

// IsInitialized returns whether the roles have been initialized
func (ri *RoleInitializer) IsInitialized() bool {
	ri.mu.RLock()
	defer ri.mu.RUnlock()
	return ri.initialized
}

// HasRole checks if a specific role is enabled
func (ri *RoleInitializer) HasRole(roleType InstanceRoleType) bool {
	if ri.config == nil {
		return false
	}
	return ri.config.HasRole(roleType)
}

// GetConfig returns the roles configuration
func (ri *RoleInitializer) GetConfig() *InstanceRolesConfig {
	return ri.config
}

// IsRoleRunning checks if a specific role worker is running
func (ri *RoleInitializer) IsRoleRunning(roleType InstanceRoleType) bool {
	ri.mu.RLock()
	defer ri.mu.RUnlock()

	worker, exists := ri.workers[roleType]
	if !exists {
		return false
	}
	return worker.Running
}

// GetRoleWorkerStatus returns the status of a specific role worker
func (ri *RoleInitializer) GetRoleWorkerStatus(roleType InstanceRoleType) *RoleWorker {
	ri.mu.RLock()
	defer ri.mu.RUnlock()

	worker, exists := ri.workers[roleType]
	if !exists {
		return nil
	}
	return worker
}

// GetRoleStatus returns the status of all roles
func (ri *RoleInitializer) GetRoleStatus() map[string]interface{} {
	ri.mu.RLock()
	defer ri.mu.RUnlock()

	status := make(map[string]interface{})
	status["initialized"] = ri.initialized

	roles := make(map[string]interface{})
	if ri.config != nil {
		for _, role := range ri.config.Roles {
			roleStatus := map[string]interface{}{
				"enabled": role.Enabled,
			}

			// Add worker status if available
			if worker, exists := ri.workers[role.Type]; exists {
				roleStatus["running"] = worker.Running
				roleStatus["started_at"] = worker.StartedAt
				if worker.Error != nil {
					roleStatus["error"] = worker.Error.Error()
				}
			}

			switch role.Type {
			case RoleApp:
				if role.App != nil {
					roleStatus["port"] = role.App.Port
					roleStatus["enable_portal"] = role.App.EnablePortal
				}
			case RoleIntegrationHub:
				if role.IntegrationHub != nil {
					roleStatus["hub_names"] = role.IntegrationHub.HubNames
					roleStatus["auto_start"] = role.IntegrationHub.AutoStart
				}
			case RoleSignalR:
				if role.SignalR != nil {
					roleStatus["port"] = role.SignalR.Port
					roleStatus["hub"] = role.SignalR.Hub
				}
			case RoleJobExecutor:
				if role.JobExecutor != nil {
					roleStatus["workers"] = role.JobExecutor.Workers
					roleStatus["enable_scheduler"] = role.JobExecutor.EnableScheduler
				}
			}

			roles[string(role.Type)] = roleStatus
		}
	}
	status["roles"] = roles

	return status
}

// GetRunningRoles returns a list of currently running role types
func (ri *RoleInitializer) GetRunningRoles() []InstanceRoleType {
	ri.mu.RLock()
	defer ri.mu.RUnlock()

	var running []InstanceRoleType
	for roleType, worker := range ri.workers {
		if worker.Running {
			running = append(running, roleType)
		}
	}
	return running
}

// Global role initializer instance
var (
	globalRoleInitializer *RoleInitializer
	roleInitMu            sync.RWMutex
)

// GetGlobalRoleInitializer returns the global role initializer
func GetGlobalRoleInitializer() *RoleInitializer {
	roleInitMu.RLock()
	defer roleInitMu.RUnlock()
	return globalRoleInitializer
}

// SetGlobalRoleInitializer sets the global role initializer
func SetGlobalRoleInitializer(initializer *RoleInitializer) {
	roleInitMu.Lock()
	defer roleInitMu.Unlock()
	globalRoleInitializer = initializer
}

// InitGlobalRoleInitializer initializes and sets the global role initializer
func InitGlobalRoleInitializer(config *InstanceRolesConfig) *RoleInitializer {
	initializer := NewRoleInitializer(config)
	SetGlobalRoleInitializer(initializer)
	return initializer
}
