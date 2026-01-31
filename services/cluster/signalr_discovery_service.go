package cluster

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// SignalR topics for cluster communication
const (
	TopicInstanceJoined    = "iac:cluster:instances:joined"
	TopicInstanceLeaving   = "iac:cluster:instances:leaving"
	TopicInstanceHeartbeat = "iac:cluster:instances:heartbeat"
	TopicInstanceStatus    = "iac:cluster:instances:status"
	TopicInstanceDiscovery = "iac:cluster:instances:discovery"
	TopicConfigReload      = "iac:cluster:config:reload"
	TopicSessionInvalidate = "iac:cluster:session:invalidate"
)

// SignalRDiscoveryService provides real-time instance discovery via SignalR
type SignalRDiscoveryService struct {
	localInstance    *models.InstanceAnnouncement
	knownInstances   map[string]*models.InstanceAnnouncement
	log              logger.Log
	heartbeatTicker  *time.Ticker
	stopCh           chan struct{}
	mutex            sync.RWMutex
	started          bool

	// Callbacks
	onJoinCallbacks   []func(*models.InstanceAnnouncement)
	onLeaveCallbacks  []func(string)
	onStatusCallbacks []func(*models.InstanceAnnouncement)
	callbackMutex     sync.RWMutex
}

// Global instance
var (
	globalDiscoveryService *SignalRDiscoveryService
	globalDiscoveryMutex   sync.RWMutex
)

// GetGlobalDiscoveryService returns the global discovery service instance
func GetGlobalDiscoveryService() *SignalRDiscoveryService {
	globalDiscoveryMutex.RLock()
	defer globalDiscoveryMutex.RUnlock()
	return globalDiscoveryService
}

// SetGlobalDiscoveryService sets the global discovery service instance
func SetGlobalDiscoveryService(service *SignalRDiscoveryService) {
	globalDiscoveryMutex.Lock()
	defer globalDiscoveryMutex.Unlock()
	globalDiscoveryService = service
}

// NewSignalRDiscoveryService creates a new SignalR discovery service
func NewSignalRDiscoveryService() *SignalRDiscoveryService {
	return &SignalRDiscoveryService{
		knownInstances:    make(map[string]*models.InstanceAnnouncement),
		log:               logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "SignalRDiscoveryService"},
		stopCh:            make(chan struct{}),
		onJoinCallbacks:   make([]func(*models.InstanceAnnouncement), 0),
		onLeaveCallbacks:  make([]func(string), 0),
		onStatusCallbacks: make([]func(*models.InstanceAnnouncement), 0),
	}
}

// Start begins the discovery service
func (s *SignalRDiscoveryService) Start(localInstance *models.InstanceRegistry) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.started {
		return fmt.Errorf("discovery service already started")
	}

	// Check if SignalR client is available
	if com.IACMessageBusClient == nil {
		s.log.Warn("SignalR client not available - discovery service will not broadcast")
	}

	// Set local instance
	s.localInstance = localInstance.ToAnnouncement()

	// Register message handlers
	s.registerHandlers()

	// Broadcast join announcement
	s.broadcastJoin()

	// Start heartbeat
	clusterConfig := config.GetClusterConfig()
	interval := 30
	if clusterConfig != nil && clusterConfig.Registry != nil {
		interval = clusterConfig.Registry.HeartbeatInterval
	}
	s.heartbeatTicker = time.NewTicker(time.Duration(interval) * time.Second)
	go s.heartbeatLoop()

	// Request discovery from other instances
	s.requestDiscovery()

	s.started = true
	s.log.Info("SignalR discovery service started")

	return nil
}

// Stop stops the discovery service
func (s *SignalRDiscoveryService) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.started {
		return
	}

	// Signal stop
	close(s.stopCh)

	// Stop heartbeat
	if s.heartbeatTicker != nil {
		s.heartbeatTicker.Stop()
	}

	// Broadcast leaving
	s.broadcastLeaving()

	s.started = false
	s.log.Info("SignalR discovery service stopped")
}

// registerHandlers registers SignalR message handlers
// Note: The SignalR client uses server-side handlers. For now, we rely on the
// embedded SignalR server to route messages. Instance discovery primarily works
// through MongoDB registry with SignalR as a notification mechanism.
func (s *SignalRDiscoveryService) registerHandlers() {
	// SignalR message handlers are typically registered on the server side
	// The client primarily sends messages via Invoke
	// For full bidirectional communication, consider implementing a custom handler
	// on the SignalR server that forwards cluster messages to clients
	s.log.Debug("SignalR discovery handlers registration - using invoke-based messaging")
}

// broadcastJoin announces this instance joining the cluster
func (s *SignalRDiscoveryService) broadcastJoin() {
	if com.IACMessageBusClient == nil || s.localInstance == nil {
		return
	}

	s.localInstance.Timestamp = time.Now()
	data, err := json.Marshal(s.localInstance)
	if err != nil {
		s.log.Error(fmt.Sprintf("Failed to marshal join announcement: %v", err))
		return
	}

	// Use Invoke to send message through SignalR
	com.IACMessageBusClient.Invoke("send", TopicInstanceJoined, string(data), "")
	s.log.Info("Broadcast instance joined announcement")
}

// broadcastLeaving announces this instance leaving the cluster
func (s *SignalRDiscoveryService) broadcastLeaving() {
	if com.IACMessageBusClient == nil || s.localInstance == nil {
		return
	}

	data := map[string]string{
		"instance_id": s.localInstance.InstanceID,
		"timestamp":   time.Now().Format(time.RFC3339),
	}
	jsonData, _ := json.Marshal(data)

	com.IACMessageBusClient.Invoke("send", TopicInstanceLeaving, string(jsonData), "")
	s.log.Info("Broadcast instance leaving announcement")
}

// BroadcastHeartbeat sends a heartbeat to all instances
func (s *SignalRDiscoveryService) BroadcastHeartbeat() {
	if com.IACMessageBusClient == nil || s.localInstance == nil {
		return
	}

	s.localInstance.Timestamp = time.Now()
	data, err := json.Marshal(s.localInstance)
	if err != nil {
		return
	}

	com.IACMessageBusClient.Invoke("send", TopicInstanceHeartbeat, string(data), "")
}

// BroadcastStatusChange announces a status change
func (s *SignalRDiscoveryService) BroadcastStatusChange(status models.InstanceStatus) {
	if com.IACMessageBusClient == nil || s.localInstance == nil {
		return
	}

	s.mutex.Lock()
	s.localInstance.Status = status
	s.localInstance.Timestamp = time.Now()
	s.mutex.Unlock()

	data, err := json.Marshal(s.localInstance)
	if err != nil {
		return
	}

	com.IACMessageBusClient.Invoke("send", TopicInstanceStatus, string(data), "")
}

// requestDiscovery requests all instances to respond with their info
func (s *SignalRDiscoveryService) requestDiscovery() {
	if com.IACMessageBusClient == nil || s.localInstance == nil {
		return
	}

	data := map[string]string{
		"requester_id": s.localInstance.InstanceID,
		"timestamp":    time.Now().Format(time.RFC3339),
	}
	jsonData, _ := json.Marshal(data)

	com.IACMessageBusClient.Invoke("send", TopicInstanceDiscovery, string(jsonData), "")
}

// heartbeatLoop sends periodic heartbeats
func (s *SignalRDiscoveryService) heartbeatLoop() {
	for {
		select {
		case <-s.heartbeatTicker.C:
			s.BroadcastHeartbeat()
			s.cleanupStaleInstances()
		case <-s.stopCh:
			return
		}
	}
}

// cleanupStaleInstances removes instances that haven't sent heartbeats
func (s *SignalRDiscoveryService) cleanupStaleInstances() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	threshold := time.Now().Add(-90 * time.Second)
	for id, instance := range s.knownInstances {
		if instance.Timestamp.Before(threshold) {
			delete(s.knownInstances, id)
			s.log.Debug(fmt.Sprintf("Removed stale instance from discovery cache: %s", id))

			// Notify callbacks
			go s.notifyLeaveCallbacks(id)
		}
	}
}

// Handler functions

func (s *SignalRDiscoveryService) handleInstanceJoined(data interface{}) {
	var announcement models.InstanceAnnouncement
	if err := s.parseAnnouncement(data, &announcement); err != nil {
		s.log.Error(fmt.Sprintf("Failed to parse join announcement: %v", err))
		return
	}

	// Ignore our own announcements
	if s.localInstance != nil && announcement.InstanceID == s.localInstance.InstanceID {
		return
	}

	s.mutex.Lock()
	s.knownInstances[announcement.InstanceID] = &announcement
	s.mutex.Unlock()

	s.log.Info(fmt.Sprintf("Instance joined: %s (%s)", announcement.InstanceName, announcement.InstanceID))

	// Respond with our own info
	s.broadcastJoin()

	// Notify callbacks
	s.notifyJoinCallbacks(&announcement)
}

func (s *SignalRDiscoveryService) handleInstanceLeaving(data interface{}) {
	var msg map[string]string
	if err := s.parseMap(data, &msg); err != nil {
		return
	}

	instanceID := msg["instance_id"]
	if instanceID == "" {
		return
	}

	s.mutex.Lock()
	delete(s.knownInstances, instanceID)
	s.mutex.Unlock()

	s.log.Info(fmt.Sprintf("Instance leaving: %s", instanceID))

	// Notify callbacks
	s.notifyLeaveCallbacks(instanceID)
}

func (s *SignalRDiscoveryService) handleHeartbeat(data interface{}) {
	var announcement models.InstanceAnnouncement
	if err := s.parseAnnouncement(data, &announcement); err != nil {
		return
	}

	// Ignore our own heartbeats
	if s.localInstance != nil && announcement.InstanceID == s.localInstance.InstanceID {
		return
	}

	s.mutex.Lock()
	s.knownInstances[announcement.InstanceID] = &announcement
	s.mutex.Unlock()
}

func (s *SignalRDiscoveryService) handleStatusChange(data interface{}) {
	var announcement models.InstanceAnnouncement
	if err := s.parseAnnouncement(data, &announcement); err != nil {
		return
	}

	// Ignore our own status changes
	if s.localInstance != nil && announcement.InstanceID == s.localInstance.InstanceID {
		return
	}

	s.mutex.Lock()
	s.knownInstances[announcement.InstanceID] = &announcement
	s.mutex.Unlock()

	s.log.Info(fmt.Sprintf("Instance status changed: %s -> %s", announcement.InstanceName, announcement.Status))

	// Notify callbacks
	s.notifyStatusCallbacks(&announcement)
}

func (s *SignalRDiscoveryService) handleDiscoveryRequest(data interface{}) {
	var msg map[string]string
	if err := s.parseMap(data, &msg); err != nil {
		return
	}

	// Don't respond to our own requests
	if s.localInstance != nil && msg["requester_id"] == s.localInstance.InstanceID {
		return
	}

	// Respond with our info
	s.broadcastJoin()
}

// Parsing helpers

func (s *SignalRDiscoveryService) parseAnnouncement(data interface{}, announcement *models.InstanceAnnouncement) error {
	var jsonStr string
	switch v := data.(type) {
	case string:
		jsonStr = v
	case []byte:
		jsonStr = string(v)
	default:
		bytes, err := json.Marshal(data)
		if err != nil {
			return err
		}
		jsonStr = string(bytes)
	}

	return json.Unmarshal([]byte(jsonStr), announcement)
}

func (s *SignalRDiscoveryService) parseMap(data interface{}, result *map[string]string) error {
	var jsonStr string
	switch v := data.(type) {
	case string:
		jsonStr = v
	case []byte:
		jsonStr = string(v)
	default:
		bytes, err := json.Marshal(data)
		if err != nil {
			return err
		}
		jsonStr = string(bytes)
	}

	return json.Unmarshal([]byte(jsonStr), result)
}

// Callback management

// OnInstanceJoined registers a callback for when an instance joins
func (s *SignalRDiscoveryService) OnInstanceJoined(callback func(*models.InstanceAnnouncement)) {
	s.callbackMutex.Lock()
	defer s.callbackMutex.Unlock()
	s.onJoinCallbacks = append(s.onJoinCallbacks, callback)
}

// OnInstanceLeft registers a callback for when an instance leaves
func (s *SignalRDiscoveryService) OnInstanceLeft(callback func(string)) {
	s.callbackMutex.Lock()
	defer s.callbackMutex.Unlock()
	s.onLeaveCallbacks = append(s.onLeaveCallbacks, callback)
}

// OnStatusChange registers a callback for status changes
func (s *SignalRDiscoveryService) OnStatusChange(callback func(*models.InstanceAnnouncement)) {
	s.callbackMutex.Lock()
	defer s.callbackMutex.Unlock()
	s.onStatusCallbacks = append(s.onStatusCallbacks, callback)
}

func (s *SignalRDiscoveryService) notifyJoinCallbacks(announcement *models.InstanceAnnouncement) {
	s.callbackMutex.RLock()
	defer s.callbackMutex.RUnlock()
	for _, cb := range s.onJoinCallbacks {
		go cb(announcement)
	}
}

func (s *SignalRDiscoveryService) notifyLeaveCallbacks(instanceID string) {
	s.callbackMutex.RLock()
	defer s.callbackMutex.RUnlock()
	for _, cb := range s.onLeaveCallbacks {
		go cb(instanceID)
	}
}

func (s *SignalRDiscoveryService) notifyStatusCallbacks(announcement *models.InstanceAnnouncement) {
	s.callbackMutex.RLock()
	defer s.callbackMutex.RUnlock()
	for _, cb := range s.onStatusCallbacks {
		go cb(announcement)
	}
}

// Query methods

// GetKnownInstances returns all known instances from SignalR discovery
func (s *SignalRDiscoveryService) GetKnownInstances() []*models.InstanceAnnouncement {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	instances := make([]*models.InstanceAnnouncement, 0, len(s.knownInstances))
	for _, inst := range s.knownInstances {
		instances = append(instances, inst)
	}
	return instances
}

// GetKnownInstanceByID returns a specific known instance
func (s *SignalRDiscoveryService) GetKnownInstanceByID(instanceID string) *models.InstanceAnnouncement {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.knownInstances[instanceID]
}

// GetKnownInstancesByRole returns instances with a specific role
func (s *SignalRDiscoveryService) GetKnownInstancesByRole(role models.InstanceRole) []*models.InstanceAnnouncement {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var instances []*models.InstanceAnnouncement
	for _, inst := range s.knownInstances {
		for _, r := range inst.Roles {
			if r == role {
				instances = append(instances, inst)
				break
			}
		}
	}
	return instances
}

// BroadcastConfigReload notifies all instances to reload configuration
func (s *SignalRDiscoveryService) BroadcastConfigReload(configVersion string) error {
	if com.IACMessageBusClient == nil {
		return fmt.Errorf("SignalR client not available")
	}

	data := map[string]string{
		"config_version": configVersion,
		"timestamp":      time.Now().Format(time.RFC3339),
		"source_id":      s.localInstance.InstanceID,
	}
	jsonData, _ := json.Marshal(data)

	com.IACMessageBusClient.Invoke("send", TopicConfigReload, string(jsonData), "")
	return nil
}

// BroadcastSessionInvalidate notifies all instances to invalidate a session
func (s *SignalRDiscoveryService) BroadcastSessionInvalidate(sessionID string) error {
	if com.IACMessageBusClient == nil {
		return fmt.Errorf("SignalR client not available")
	}

	data := map[string]string{
		"session_id": sessionID,
		"timestamp":  time.Now().Format(time.RFC3339),
		"source_id":  s.localInstance.InstanceID,
	}
	jsonData, _ := json.Marshal(data)

	com.IACMessageBusClient.Invoke("send", TopicSessionInvalidate, string(jsonData), "")
	return nil
}
