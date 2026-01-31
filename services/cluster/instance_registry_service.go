package cluster

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// InstanceRegistryService manages instance registration and discovery
type InstanceRegistryService struct {
	collection      *mongo.Collection
	localInstance   *models.InstanceRegistry
	config          *config.RegistryConfig
	log             logger.Log
	heartbeatTicker *time.Ticker
	cleanupTicker   *time.Ticker
	stopCh          chan struct{}
	mutex           sync.RWMutex
	started         bool
}

// Global instance
var (
	globalRegistryService *InstanceRegistryService
	globalRegistryMutex   sync.RWMutex
)

// GetGlobalRegistryService returns the global registry service instance
func GetGlobalRegistryService() *InstanceRegistryService {
	globalRegistryMutex.RLock()
	defer globalRegistryMutex.RUnlock()
	return globalRegistryService
}

// SetGlobalRegistryService sets the global registry service instance
func SetGlobalRegistryService(service *InstanceRegistryService) {
	globalRegistryMutex.Lock()
	defer globalRegistryMutex.Unlock()
	globalRegistryService = service
}

// NewInstanceRegistryService creates a new instance registry service
func NewInstanceRegistryService(cfg *config.RegistryConfig) (*InstanceRegistryService, error) {
	if cfg == nil {
		cfg = config.DefaultRegistryConfig()
	}

	// Get MongoDB client from global connection
	if com.IACDocDBConn.MongoDBClient == nil {
		return nil, fmt.Errorf("MongoDB client not available")
	}

	db := com.IACDocDBConn.MongoDBClient.Database(com.IACDocDBConn.DatabaseName)
	collection := db.Collection(cfg.Collection)

	service := &InstanceRegistryService{
		collection: collection,
		config:     cfg,
		log:        logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "InstanceRegistryService"},
		stopCh:     make(chan struct{}),
	}

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Unique index on instance_id
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "instance_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		service.log.Error(fmt.Sprintf("Failed to create instance_id index: %v", err))
	}

	// Index on status
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}},
	})
	if err != nil {
		service.log.Error(fmt.Sprintf("Failed to create status index: %v", err))
	}

	// Index on roles
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "roles", Value: 1}},
	})
	if err != nil {
		service.log.Error(fmt.Sprintf("Failed to create roles index: %v", err))
	}

	// TTL index for automatic cleanup of stale instances
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "last_heartbeat", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(int32(cfg.HeartbeatTTL * 2)),
	})
	if err != nil {
		service.log.Error(fmt.Sprintf("Failed to create TTL index: %v", err))
	}

	service.log.Info("Instance registry service initialized")
	return service, nil
}

// Start begins the heartbeat and cleanup processes
func (s *InstanceRegistryService) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.started {
		return fmt.Errorf("registry service already started")
	}

	// Create local instance entry
	s.localInstance = s.createLocalInstance()

	// Register this instance
	if err := s.register(ctx); err != nil {
		return fmt.Errorf("failed to register instance: %w", err)
	}

	// Start heartbeat
	s.heartbeatTicker = time.NewTicker(time.Duration(s.config.HeartbeatInterval) * time.Second)
	go s.heartbeatLoop()

	// Start cleanup
	s.cleanupTicker = time.NewTicker(time.Duration(s.config.CleanupInterval) * time.Second)
	go s.cleanupLoop()

	s.started = true
	s.log.Info(fmt.Sprintf("Instance registry started - ID: %s, Name: %s", s.localInstance.InstanceID, s.localInstance.InstanceName))

	return nil
}

// Stop stops the heartbeat and deregisters the instance
func (s *InstanceRegistryService) Stop(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.started {
		return nil
	}

	// Signal stop
	close(s.stopCh)

	// Stop tickers
	if s.heartbeatTicker != nil {
		s.heartbeatTicker.Stop()
	}
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}

	// Update status to stopping
	if s.localInstance != nil {
		s.localInstance.SetStatus(models.InstanceStatusStopping)
		s.updateInstance(ctx, s.localInstance)
	}

	s.started = false
	s.log.Info("Instance registry stopped")

	return nil
}

// createLocalInstance creates the local instance registry entry
func (s *InstanceRegistryService) createLocalInstance() *models.InstanceRegistry {
	clusterConfig := config.GetClusterConfig()

	instanceID := com.InstanceID
	instanceName := config.GetInstanceName()
	hostname, _ := os.Hostname()
	ipAddress := getLocalIP()
	port := getConfiguredPort()

	instance := models.NewInstanceRegistry(instanceID, instanceName, hostname, ipAddress, port)
	instance.Environment = config.GetEnvironment()
	instance.Version = getVersion()

	// Add roles based on configuration
	rolesConfig := config.GetRolesConfig()
	if rolesConfig != nil {
		for _, role := range rolesConfig.GetEnabledRoles() {
			switch role {
			case config.RoleApp:
				instance.AddRole(models.InstanceRoleApp)
			case config.RoleIntegrationHub:
				instance.AddRole(models.InstanceRoleIntegrationHub)
			case config.RoleSignalR:
				instance.AddRole(models.InstanceRoleSignalR)
			case config.RoleJobExecutor:
				instance.AddRole(models.InstanceRoleJobExecutor)
			}
		}
	}

	// Add container info if available
	instance.Container = detectContainerInfo()

	// Add capacity
	instance.Capacity = &models.InstanceCapacity{
		MaxConnections: 1000,
		CPUCores:       runtime.NumCPU(),
	}

	// Check SignalR connectivity
	if clusterConfig != nil && clusterConfig.Registry != nil && clusterConfig.Registry.SignalRDiscovery {
		instance.SignalRConnected = com.IACMessageBusClient != nil
	}

	return instance
}

// register registers this instance in the registry
func (s *InstanceRegistryService) register(ctx context.Context) error {
	s.localInstance.RegisteredAt = time.Now()
	s.localInstance.LastHeartbeat = time.Now()
	s.localInstance.SetStatus(models.InstanceStatusRunning)

	filter := bson.M{"instance_id": s.localInstance.InstanceID}
	update := bson.M{"$set": s.localInstance}
	opts := options.Update().SetUpsert(true)

	_, err := s.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to register instance: %w", err)
	}

	s.log.Info(fmt.Sprintf("Instance registered: %s (%s)", s.localInstance.InstanceName, s.localInstance.InstanceID))
	return nil
}

// heartbeatLoop sends periodic heartbeats
func (s *InstanceRegistryService) heartbeatLoop() {
	for {
		select {
		case <-s.heartbeatTicker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			s.sendHeartbeat(ctx)
			cancel()
		case <-s.stopCh:
			return
		}
	}
}

// sendHeartbeat sends a heartbeat update
func (s *InstanceRegistryService) sendHeartbeat(ctx context.Context) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.localInstance == nil {
		return
	}

	s.localInstance.UpdateHeartbeat()
	s.localInstance.UpdateMetrics(s.collectMetrics())

	if err := s.updateInstance(ctx, s.localInstance); err != nil {
		s.log.Error(fmt.Sprintf("Failed to send heartbeat: %v", err))
	}
}

// updateInstance updates an instance in the registry
func (s *InstanceRegistryService) updateInstance(ctx context.Context, instance *models.InstanceRegistry) error {
	filter := bson.M{"instance_id": instance.InstanceID}
	update := bson.M{"$set": instance}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

// collectMetrics collects current instance metrics
func (s *InstanceRegistryService) collectMetrics() *models.InstanceMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &models.InstanceMetrics{
		MemoryUsedMB:    int(memStats.Alloc / 1024 / 1024),
		UptimeSeconds:   int64(time.Since(s.localInstance.StartTime).Seconds()),
		CollectedAt:     time.Now(),
	}
}

// cleanupLoop periodically removes stale instances
func (s *InstanceRegistryService) cleanupLoop() {
	for {
		select {
		case <-s.cleanupTicker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			s.cleanupStaleInstances(ctx)
			cancel()
		case <-s.stopCh:
			return
		}
	}
}

// cleanupStaleInstances removes instances that haven't sent heartbeats
func (s *InstanceRegistryService) cleanupStaleInstances(ctx context.Context) {
	threshold := time.Now().Add(-time.Duration(s.config.HeartbeatTTL) * time.Second)

	filter := bson.M{
		"last_heartbeat": bson.M{"$lt": threshold},
		"instance_id":    bson.M{"$ne": s.localInstance.InstanceID},
	}

	result, err := s.collection.DeleteMany(ctx, filter)
	if err != nil {
		s.log.Error(fmt.Sprintf("Failed to cleanup stale instances: %v", err))
		return
	}

	if result.DeletedCount > 0 {
		s.log.Info(fmt.Sprintf("Cleaned up %d stale instances", result.DeletedCount))
	}
}

// GetAllInstances returns all registered instances
func (s *InstanceRegistryService) GetAllInstances(ctx context.Context) ([]*models.InstanceRegistry, error) {
	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}
	defer cursor.Close(ctx)

	var instances []*models.InstanceRegistry
	for cursor.Next(ctx) {
		var instance models.InstanceRegistry
		if err := cursor.Decode(&instance); err != nil {
			s.log.Error(fmt.Sprintf("Failed to decode instance: %v", err))
			continue
		}
		instances = append(instances, &instance)
	}

	return instances, nil
}

// GetInstancesByRole returns instances with a specific role
func (s *InstanceRegistryService) GetInstancesByRole(ctx context.Context, role models.InstanceRole) ([]*models.InstanceRegistry, error) {
	filter := bson.M{
		"roles":  role,
		"status": models.InstanceStatusRunning,
	}

	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get instances by role: %w", err)
	}
	defer cursor.Close(ctx)

	var instances []*models.InstanceRegistry
	for cursor.Next(ctx) {
		var instance models.InstanceRegistry
		if err := cursor.Decode(&instance); err != nil {
			continue
		}
		instances = append(instances, &instance)
	}

	return instances, nil
}

// GetInstanceByID returns a specific instance
func (s *InstanceRegistryService) GetInstanceByID(ctx context.Context, instanceID string) (*models.InstanceRegistry, error) {
	var instance models.InstanceRegistry
	err := s.collection.FindOne(ctx, bson.M{"instance_id": instanceID}).Decode(&instance)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	return &instance, nil
}

// GetClusterStatus returns the overall cluster status
func (s *InstanceRegistryService) GetClusterStatus(ctx context.Context) (*models.ClusterStatus, error) {
	instances, err := s.GetAllInstances(ctx)
	if err != nil {
		return nil, err
	}

	status := &models.ClusterStatus{
		InstancesByRole:   make(map[models.InstanceRole]int),
		InstancesByStatus: make(map[models.InstanceStatus]int),
		Instances:         instances,
		LastUpdated:       time.Now(),
	}

	threshold := time.Duration(s.config.HeartbeatTTL) * time.Second

	for _, inst := range instances {
		status.TotalInstances++

		// Count by status
		if inst.IsStale(threshold) {
			status.UnhealthyInstances++
			status.InstancesByStatus[models.InstanceStatusUnknown]++
		} else {
			switch inst.Status {
			case models.InstanceStatusRunning:
				status.HealthyInstances++
			case models.InstanceStatusDegraded:
				status.DegradedInstances++
			default:
				status.UnhealthyInstances++
			}
			status.InstancesByStatus[inst.Status]++
		}

		// Count by role
		for _, role := range inst.Roles {
			status.InstancesByRole[role]++
		}
	}

	return status, nil
}

// GetLocalInstance returns the local instance registry entry
func (s *InstanceRegistryService) GetLocalInstance() *models.InstanceRegistry {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.localInstance
}

// UpdateLocalStatus updates the local instance status
func (s *InstanceRegistryService) UpdateLocalStatus(status models.InstanceStatus) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.localInstance != nil {
		s.localInstance.SetStatus(status)
	}
}

// Helper functions

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func getConfiguredPort() int {
	// Try to get from global configuration
	if config.GlobalConfiguration != nil {
		if webConfig := config.GlobalConfiguration.WebServerConfig; webConfig != nil {
			if port, ok := webConfig["port"].(float64); ok {
				return int(port)
			}
		}
	}
	return 8080 // Default port
}

func getVersion() string {
	// This could be injected at build time
	return "1.0.0"
}

func detectContainerInfo() *models.ContainerInfo {
	info := &models.ContainerInfo{}

	// Check for Kubernetes environment variables
	if podName := os.Getenv("HOSTNAME"); podName != "" {
		info.PodName = podName
	}
	if podNamespace := os.Getenv("POD_NAMESPACE"); podNamespace != "" {
		info.PodNamespace = podNamespace
	}
	if nodeName := os.Getenv("NODE_NAME"); nodeName != "" {
		info.NodeName = nodeName
	}
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		info.ServiceName = serviceName
	}

	// Check for container ID
	if containerID := os.Getenv("CONTAINER_ID"); containerID != "" {
		info.ContainerID = containerID
	}

	// Return nil if no container info detected
	if info.PodName == "" && info.ContainerID == "" {
		return nil
	}

	return info
}
