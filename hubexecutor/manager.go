package hubexecutor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// HubExecutorManager manages multiple hub executors
type HubExecutorManager struct {
	mu              sync.RWMutex
	iLog            logger.Log
	executors       map[string]*HubExecutor // key: hubName
	hubLoader       HubLoader
	destExecutor    DestinationExecutor
	historyRecorder HistoryRecorder
	onMessage       func(payload *MessagePayload)
	onError         func(err error, source string)
}

// HubLoader is an interface for loading hub configurations
type HubLoader interface {
	// LoadByHubName loads the default hub configuration for a hub name
	LoadByHubName(hubName string) (*models.IntegrationHub, error)

	// LoadByID loads a hub configuration by its ID
	LoadByID(hubID string) (*models.IntegrationHub, error)

	// LoadAll loads all enabled hub configurations
	LoadAll() ([]*models.IntegrationHub, error)
}

// ManagerConfig holds configuration for the hub executor manager
type ManagerConfig struct {
	HubLoader       HubLoader
	DestExecutor    DestinationExecutor
	HistoryRecorder HistoryRecorder
	OnMessage       func(payload *MessagePayload)
	OnError         func(err error, source string)
}

// NewHubExecutorManager creates a new hub executor manager
func NewHubExecutorManager(config ManagerConfig) *HubExecutorManager {
	manager := &HubExecutorManager{
		iLog:            logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "HubExecutorManager"},
		executors:       make(map[string]*HubExecutor),
		hubLoader:       config.HubLoader,
		destExecutor:    config.DestExecutor,
		historyRecorder: config.HistoryRecorder,
		onMessage:       config.OnMessage,
		onError:         config.OnError,
	}

	// Use default destination executor if none provided
	if manager.destExecutor == nil {
		manager.destExecutor = NewDestinationExecutor()
	}

	return manager
}

// StartExecutor starts a hub executor for the given hub name
func (m *HubExecutorManager) StartExecutor(hubName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if executor already exists and is running
	if executor, exists := m.executors[hubName]; exists {
		if executor.Status() == StatusRunning {
			return fmt.Errorf("executor for hub %s is already running", hubName)
		}
	}

	// Load hub configuration
	if m.hubLoader == nil {
		return fmt.Errorf("hub loader not configured")
	}

	hub, err := m.hubLoader.LoadByHubName(hubName)
	if err != nil {
		return fmt.Errorf("failed to load hub for %s: %v", hubName, err)
	}

	if hub == nil {
		return fmt.Errorf("no hub configuration found for %s", hubName)
	}

	// Check if hub is enabled
	if !hub.Enabled {
		return fmt.Errorf("hub %s is not enabled", hub.Name)
	}

	// Create and start executor
	executor := NewHubExecutor(HubExecutorConfig{
		Hub:              hub,
		HubName:          hubName,
		DestinationExec:  m.destExecutor,
		HistoryRecorder:  m.historyRecorder,
		OnMessageReceive: m.handleMessage,
		OnError:          m.handleError,
	})

	m.iLog.Info(fmt.Sprintf("HubExecutorManager: Attempting to start executor for hub %s (ID: %s)", hubName, hub.ID))
	fmt.Printf("HubExecutorManager: Starting hub %s\n", hubName)

	if err := executor.Start(); err != nil {
		return fmt.Errorf("failed to start executor for hub %s: %v", hubName, err)
	}

	m.executors[hubName] = executor
	m.iLog.Info(fmt.Sprintf("Started executor for hub: %s (version: %d)", hubName, hub.Version))

	return nil
}

// StartExecutorByHubID starts a hub executor using a specific hub configuration ID
func (m *HubExecutorManager) StartExecutorByHubID(hubID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load hub configuration by ID
	if m.hubLoader == nil {
		return fmt.Errorf("hub loader not configured")
	}

	hub, err := m.hubLoader.LoadByID(hubID)
	if err != nil {
		return fmt.Errorf("failed to load hub %s: %v", hubID, err)
	}

	if hub == nil {
		return fmt.Errorf("no hub configuration found for ID %s", hubID)
	}

	hubName := hub.Name
	if hubName == "" {
		hubName = hub.ID
	}

	// Check if executor already exists and is running
	if executor, exists := m.executors[hubName]; exists {
		if executor.Status() == StatusRunning {
			return fmt.Errorf("executor for hub %s is already running", hubName)
		}
	}

	// Check if hub is enabled
	if !hub.Enabled {
		return fmt.Errorf("hub %s is not enabled", hub.Name)
	}

	// Create and start executor
	executor := NewHubExecutor(HubExecutorConfig{
		Hub:              hub,
		HubName:          hubName,
		DestinationExec:  m.destExecutor,
		HistoryRecorder:  m.historyRecorder,
		OnMessageReceive: m.handleMessage,
		OnError:          m.handleError,
	})

	if err := executor.Start(); err != nil {
		return fmt.Errorf("failed to start executor for hub %s: %v", hubID, err)
	}

	m.executors[hubName] = executor
	m.iLog.Info(fmt.Sprintf("Started executor for hub ID: %s (name: %s)", hubID, hubName))

	return nil
}

// StopExecutor stops a hub executor by hub name
func (m *HubExecutorManager) StopExecutor(hubName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	executor, exists := m.executors[hubName]
	if !exists {
		return fmt.Errorf("no executor found for hub %s", hubName)
	}

	if err := executor.Stop(); err != nil {
		return fmt.Errorf("failed to stop executor for hub %s: %v", hubName, err)
	}

	delete(m.executors, hubName)
	m.iLog.Info(fmt.Sprintf("Stopped executor for hub: %s", hubName))

	return nil
}

// StopAll stops all running executors
func (m *HubExecutorManager) StopAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.iLog.Info("Stopping all hub executors...")

	var errors []error
	for hubName, executor := range m.executors {
		if err := executor.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop executor %s: %v", hubName, err))
		}
	}

	m.executors = make(map[string]*HubExecutor)

	if len(errors) > 0 {
		return fmt.Errorf("errors stopping executors: %v", errors)
	}

	m.iLog.Info("All hub executors stopped")
	return nil
}

// RestartExecutor restarts a hub executor
func (m *HubExecutorManager) RestartExecutor(hubName string) error {
	// Stop if running
	m.mu.RLock()
	_, exists := m.executors[hubName]
	m.mu.RUnlock()

	if exists {
		if err := m.StopExecutor(hubName); err != nil {
			return fmt.Errorf("failed to stop executor during restart: %v", err)
		}
	}

	// Wait a bit before restarting
	time.Sleep(500 * time.Millisecond)

	// Start
	return m.StartExecutor(hubName)
}

// GetExecutorStatus gets the status of a specific executor
func (m *HubExecutorManager) GetExecutorStatus(hubName string) (ExecutorStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	executor, exists := m.executors[hubName]
	if !exists {
		return StatusStopped, nil
	}

	return executor.Status(), nil
}

// GetExecutorInfo gets detailed info for a specific executor
func (m *HubExecutorManager) GetExecutorInfo(hubName string) (*ExecutorInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	executor, exists := m.executors[hubName]
	if !exists {
		return nil, fmt.Errorf("no executor found for %s", hubName)
	}

	info := executor.GetInfo()
	return &info, nil
}

// ListExecutors returns information about all running executors
func (m *HubExecutorManager) ListExecutors() []ExecutorInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]ExecutorInfo, 0, len(m.executors))
	for _, executor := range m.executors {
		infos = append(infos, executor.GetInfo())
	}

	return infos
}

// IsRunning checks if an executor is running for the given name
func (m *HubExecutorManager) IsRunning(hubName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	executor, exists := m.executors[hubName]
	if !exists {
		return false
	}

	return executor.Status() == StatusRunning
}

// GetRunningCount returns the number of running executors
func (m *HubExecutorManager) GetRunningCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, executor := range m.executors {
		if executor.Status() == StatusRunning {
			count++
		}
	}
	return count
}

// StartAll starts executors for all enabled hubs
func (m *HubExecutorManager) StartAll() error {
	if m.hubLoader == nil {
		return fmt.Errorf("hub loader not configured")
	}

	hubs, err := m.hubLoader.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load hubs: %v", err)
	}

	m.iLog.Info(fmt.Sprintf("Starting executors for %d hubs...", len(hubs)))

	var errors []error
	for _, hub := range hubs {
		if !hub.Enabled {
			continue
		}

		hubName := hub.Name
		if hubName == "" {
			hubName = hub.ID
		}

		if err := m.StartExecutor(hubName); err != nil {
			errors = append(errors, err)
			m.iLog.Error(fmt.Sprintf("Failed to start executor for %s: %v", hubName, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors starting executors: %v", errors)
	}

	return nil
}

// StartAllExecutors starts executors for all enabled hubs and returns results map
func (m *HubExecutorManager) StartAllExecutors() map[string]error {
	results := make(map[string]error)

	if m.hubLoader == nil {
		return results
	}

	hubs, err := m.hubLoader.LoadAll()
	if err != nil {
		m.iLog.Error(fmt.Sprintf("Failed to load hubs: %v", err))
		return results
	}

	m.iLog.Info(fmt.Sprintf("Starting executors for %d hubs...", len(hubs)))

	for _, hub := range hubs {
		if !hub.Enabled {
			continue
		}

		hubName := hub.Name
		if hubName == "" {
			hubName = hub.ID
		}

		err := m.StartExecutor(hubName)
		results[hubName] = err
	}

	return results
}

// handleMessage handles incoming messages from executors
func (m *HubExecutorManager) handleMessage(payload *MessagePayload) {
	if m.onMessage != nil {
		m.onMessage(payload)
	}
}

// handleError handles errors from executors
func (m *HubExecutorManager) handleError(err error, source string) {
	m.iLog.Error(fmt.Sprintf("Error from %s: %v", source, err))
	if m.onError != nil {
		m.onError(err, source)
	}
}

// SetHubLoader sets the hub loader
func (m *HubExecutorManager) SetHubLoader(loader HubLoader) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hubLoader = loader
}

// SetDestinationExecutor sets the destination executor
func (m *HubExecutorManager) SetDestinationExecutor(executor DestinationExecutor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.destExecutor = executor
}

// ============================================================================
// Default Hub Loader Implementation
// ============================================================================

// MongoHubLoader loads hub configurations from MongoDB
type MongoHubLoader struct {
	iLog       logger.Log
	getHub     func(name string) (*models.IntegrationHub, error)
	getHubByID func(hubID string) (*models.IntegrationHub, error)
	getAllHubs func() ([]*models.IntegrationHub, error)
}

// MongoHubLoaderConfig holds configuration for MongoHubLoader
type MongoHubLoaderConfig struct {
	GetHubByHubName func(name string) (*models.IntegrationHub, error)
	GetHubByID      func(hubID string) (*models.IntegrationHub, error)
	GetAllHubs      func() ([]*models.IntegrationHub, error)
}

// NewMongoHubLoader creates a new MongoDB-based hub loader
func NewMongoHubLoader(config MongoHubLoaderConfig) *MongoHubLoader {
	return &MongoHubLoader{
		iLog:       logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MongoHubLoader"},
		getHub:     config.GetHubByHubName,
		getHubByID: config.GetHubByID,
		getAllHubs: config.GetAllHubs,
	}
}

// LoadByHubName loads the default hub configuration for a hub name
func (l *MongoHubLoader) LoadByHubName(hubName string) (*models.IntegrationHub, error) {
	if l.getHub == nil {
		return nil, fmt.Errorf("getHub function not configured")
	}
	return l.getHub(hubName)
}

// LoadByID loads a hub configuration by its ID
func (l *MongoHubLoader) LoadByID(hubID string) (*models.IntegrationHub, error) {
	if l.getHubByID == nil {
		return nil, fmt.Errorf("getHubByID function not configured")
	}
	return l.getHubByID(hubID)
}

// LoadAll loads all enabled hub configurations
func (l *MongoHubLoader) LoadAll() ([]*models.IntegrationHub, error) {
	if l.getAllHubs == nil {
		return nil, fmt.Errorf("getAllHubs function not configured")
	}
	return l.getAllHubs()
}

// ============================================================================
// Global Manager Instance
// ============================================================================

var (
	globalManager *HubExecutorManager
	globalMu      sync.RWMutex
)

// GetGlobalManager returns the global hub executor manager
func GetGlobalManager() *HubExecutorManager {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalManager
}

// SetGlobalManager sets the global hub executor manager
func SetGlobalManager(manager *HubExecutorManager) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalManager = manager
}

// InitGlobalManager initializes and sets the global manager
func InitGlobalManager(config ManagerConfig) *HubExecutorManager {
	manager := NewHubExecutorManager(config)
	SetGlobalManager(manager)
	return manager
}

// StartGlobalExecutor starts an executor using the global manager
func StartGlobalExecutor(hubName string) error {
	manager := GetGlobalManager()
	if manager == nil {
		return fmt.Errorf("global hub executor manager not initialized")
	}
	return manager.StartExecutor(hubName)
}

// StopGlobalExecutor stops an executor using the global manager
func StopGlobalExecutor(hubName string) error {
	manager := GetGlobalManager()
	if manager == nil {
		return fmt.Errorf("global hub executor manager not initialized")
	}
	return manager.StopExecutor(hubName)
}

// ============================================================================
// Job Executor Integration
// ============================================================================

// JobExecutorFunc is a function type for executing jobs
type JobExecutorFunc func(ctx context.Context, jobName string, params map[string]interface{}) error

// SetJobExecutor sets the job executor function on the destination executor
func (m *HubExecutorManager) SetJobExecutor(executor JobExecutorFunc) {
	if destExec, ok := m.destExecutor.(*DefaultDestinationExecutor); ok {
		destExec.SetJobExecutor(executor)
	}
}
