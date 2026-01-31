package models

import (
	"time"
)

// InstanceStatus represents the current status of an IAC instance
type InstanceStatus string

const (
	InstanceStatusStarting    InstanceStatus = "starting"
	InstanceStatusRunning     InstanceStatus = "running"
	InstanceStatusDegraded    InstanceStatus = "degraded"
	InstanceStatusMaintenance InstanceStatus = "maintenance"
	InstanceStatusStopping    InstanceStatus = "stopping"
	InstanceStatusStopped     InstanceStatus = "stopped"
	InstanceStatusUnknown     InstanceStatus = "unknown"
)

// InstanceRole represents the role of an IAC instance
type InstanceRole string

const (
	InstanceRoleApp            InstanceRole = "app"
	InstanceRoleIntegrationHub InstanceRole = "integration_hub"
	InstanceRoleSignalR        InstanceRole = "signalr"
	InstanceRoleJobExecutor    InstanceRole = "job_executor"
)

// InstanceRegistry represents an IAC instance in the cluster registry
type InstanceRegistry struct {
	// Instance identification
	InstanceID   string `json:"instance_id" bson:"instance_id"`
	InstanceName string `json:"instance_name" bson:"instance_name"`
	Hostname     string `json:"hostname" bson:"hostname"`
	IPAddress    string `json:"ip_address" bson:"ip_address"`
	Port         int    `json:"port" bson:"port"`

	// Instance roles and capabilities
	Roles       []InstanceRole `json:"roles" bson:"roles"`
	Capabilities []string      `json:"capabilities,omitempty" bson:"capabilities,omitempty"`

	// Status and health
	Status       InstanceStatus `json:"status" bson:"status"`
	HealthStatus *HealthStatus  `json:"health_status,omitempty" bson:"health_status,omitempty"`

	// Version information
	Version     string `json:"version" bson:"version"`
	BuildTime   string `json:"build_time,omitempty" bson:"build_time,omitempty"`
	GitCommit   string `json:"git_commit,omitempty" bson:"git_commit,omitempty"`

	// Environment
	Environment string `json:"environment" bson:"environment"`
	Region      string `json:"region,omitempty" bson:"region,omitempty"`
	Zone        string `json:"zone,omitempty" bson:"zone,omitempty"`

	// Container/Kubernetes metadata
	Container *ContainerInfo `json:"container,omitempty" bson:"container,omitempty"`

	// Capacity and metrics
	Capacity *InstanceCapacity `json:"capacity,omitempty" bson:"capacity,omitempty"`
	Metrics  *InstanceMetrics  `json:"metrics,omitempty" bson:"metrics,omitempty"`

	// Configuration
	ConfigVersion string                 `json:"config_version,omitempty" bson:"config_version,omitempty"`
	Settings      map[string]interface{} `json:"settings,omitempty" bson:"settings,omitempty"`

	// Timing
	StartTime       time.Time `json:"start_time" bson:"start_time"`
	RegisteredAt    time.Time `json:"registered_at" bson:"registered_at"`
	LastHeartbeat   time.Time `json:"last_heartbeat" bson:"last_heartbeat"`
	LastStatusChange time.Time `json:"last_status_change,omitempty" bson:"last_status_change,omitempty"`

	// SignalR connectivity
	SignalRConnected bool   `json:"signalr_connected" bson:"signalr_connected"`
	SignalRHubURL    string `json:"signalr_hub_url,omitempty" bson:"signalr_hub_url,omitempty"`
}

// HealthStatus represents the health of an instance
type HealthStatus struct {
	Overall    string                    `json:"overall" bson:"overall"` // healthy, unhealthy, degraded
	Components map[string]ComponentHealth `json:"components,omitempty" bson:"components,omitempty"`
	LastCheck  time.Time                 `json:"last_check" bson:"last_check"`
}

// ComponentHealth represents the health of a specific component
type ComponentHealth struct {
	Status  string `json:"status" bson:"status"` // healthy, unhealthy, degraded
	Message string `json:"message,omitempty" bson:"message,omitempty"`
	Latency int64  `json:"latency_ms,omitempty" bson:"latency_ms,omitempty"`
}

// ContainerInfo holds container/Kubernetes specific information
type ContainerInfo struct {
	ContainerID   string `json:"container_id,omitempty" bson:"container_id,omitempty"`
	ContainerName string `json:"container_name,omitempty" bson:"container_name,omitempty"`
	ImageName     string `json:"image_name,omitempty" bson:"image_name,omitempty"`
	ImageTag      string `json:"image_tag,omitempty" bson:"image_tag,omitempty"`

	// Kubernetes specific
	PodName      string            `json:"pod_name,omitempty" bson:"pod_name,omitempty"`
	PodNamespace string            `json:"pod_namespace,omitempty" bson:"pod_namespace,omitempty"`
	NodeName     string            `json:"node_name,omitempty" bson:"node_name,omitempty"`
	ServiceName  string            `json:"service_name,omitempty" bson:"service_name,omitempty"`
	Labels       map[string]string `json:"labels,omitempty" bson:"labels,omitempty"`
}

// InstanceCapacity represents the capacity configuration of an instance
type InstanceCapacity struct {
	MaxConnections     int `json:"max_connections" bson:"max_connections"`
	MaxConcurrentJobs  int `json:"max_concurrent_jobs,omitempty" bson:"max_concurrent_jobs,omitempty"`
	MaxSessionsPerUser int `json:"max_sessions_per_user,omitempty" bson:"max_sessions_per_user,omitempty"`
	MemoryLimitMB      int `json:"memory_limit_mb,omitempty" bson:"memory_limit_mb,omitempty"`
	CPUCores           int `json:"cpu_cores,omitempty" bson:"cpu_cores,omitempty"`
}

// InstanceMetrics represents real-time metrics of an instance
type InstanceMetrics struct {
	ActiveConnections int     `json:"active_connections" bson:"active_connections"`
	ActiveSessions    int     `json:"active_sessions" bson:"active_sessions"`
	ActiveJobs        int     `json:"active_jobs,omitempty" bson:"active_jobs,omitempty"`
	RequestsPerSecond float64 `json:"requests_per_second,omitempty" bson:"requests_per_second,omitempty"`
	AvgResponseTimeMs float64 `json:"avg_response_time_ms,omitempty" bson:"avg_response_time_ms,omitempty"`
	MemoryUsedMB      int     `json:"memory_used_mb,omitempty" bson:"memory_used_mb,omitempty"`
	CPUUsagePercent   float64 `json:"cpu_usage_percent,omitempty" bson:"cpu_usage_percent,omitempty"`
	UptimeSeconds     int64   `json:"uptime_seconds" bson:"uptime_seconds"`
	CollectedAt       time.Time `json:"collected_at" bson:"collected_at"`
}

// InstanceAnnouncement is a lightweight message for SignalR broadcasts
type InstanceAnnouncement struct {
	InstanceID   string           `json:"instance_id"`
	InstanceName string           `json:"instance_name"`
	Hostname     string           `json:"hostname"`
	IPAddress    string           `json:"ip_address"`
	Port         int              `json:"port"`
	Roles        []InstanceRole   `json:"roles"`
	Status       InstanceStatus   `json:"status"`
	Version      string           `json:"version"`
	Environment  string           `json:"environment"`
	StartTime    time.Time        `json:"start_time"`
	Timestamp    time.Time        `json:"timestamp"`
	Capacity     *InstanceCapacity `json:"capacity,omitempty"`
}

// ClusterStatus represents the overall cluster status
type ClusterStatus struct {
	TotalInstances    int                       `json:"total_instances"`
	HealthyInstances  int                       `json:"healthy_instances"`
	DegradedInstances int                       `json:"degraded_instances"`
	UnhealthyInstances int                      `json:"unhealthy_instances"`
	InstancesByRole   map[InstanceRole]int      `json:"instances_by_role"`
	InstancesByStatus map[InstanceStatus]int    `json:"instances_by_status"`
	Instances         []*InstanceRegistry       `json:"instances,omitempty"`
	LastUpdated       time.Time                 `json:"last_updated"`
}

// NewInstanceRegistry creates a new instance registry entry
func NewInstanceRegistry(instanceID, instanceName, hostname, ipAddress string, port int) *InstanceRegistry {
	now := time.Now()
	return &InstanceRegistry{
		InstanceID:   instanceID,
		InstanceName: instanceName,
		Hostname:     hostname,
		IPAddress:    ipAddress,
		Port:         port,
		Status:       InstanceStatusStarting,
		StartTime:    now,
		RegisteredAt: now,
		LastHeartbeat: now,
		Roles:        []InstanceRole{},
		Capacity: &InstanceCapacity{
			MaxConnections: 1000,
		},
		Metrics: &InstanceMetrics{
			CollectedAt: now,
		},
	}
}

// ToAnnouncement converts the registry entry to a lightweight announcement
func (r *InstanceRegistry) ToAnnouncement() *InstanceAnnouncement {
	return &InstanceAnnouncement{
		InstanceID:   r.InstanceID,
		InstanceName: r.InstanceName,
		Hostname:     r.Hostname,
		IPAddress:    r.IPAddress,
		Port:         r.Port,
		Roles:        r.Roles,
		Status:       r.Status,
		Version:      r.Version,
		Environment:  r.Environment,
		StartTime:    r.StartTime,
		Timestamp:    time.Now(),
		Capacity:     r.Capacity,
	}
}

// UpdateHeartbeat updates the heartbeat timestamp
func (r *InstanceRegistry) UpdateHeartbeat() {
	r.LastHeartbeat = time.Now()
}

// SetStatus updates the instance status
func (r *InstanceRegistry) SetStatus(status InstanceStatus) {
	if r.Status != status {
		r.Status = status
		r.LastStatusChange = time.Now()
	}
}

// HasRole checks if the instance has a specific role
func (r *InstanceRegistry) HasRole(role InstanceRole) bool {
	for _, r := range r.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// AddRole adds a role to the instance
func (r *InstanceRegistry) AddRole(role InstanceRole) {
	if !r.HasRole(role) {
		r.Roles = append(r.Roles, role)
	}
}

// IsHealthy checks if the instance is considered healthy
func (r *InstanceRegistry) IsHealthy() bool {
	return r.Status == InstanceStatusRunning
}

// IsStale checks if the instance heartbeat is stale (no heartbeat within threshold)
func (r *InstanceRegistry) IsStale(threshold time.Duration) bool {
	return time.Since(r.LastHeartbeat) > threshold
}

// GetEndpoint returns the HTTP endpoint URL for this instance
func (r *InstanceRegistry) GetEndpoint() string {
	if r.IPAddress != "" {
		return "http://" + r.IPAddress + ":" + string(rune(r.Port))
	}
	return "http://" + r.Hostname + ":" + string(rune(r.Port))
}

// UpdateMetrics updates the instance metrics
func (r *InstanceRegistry) UpdateMetrics(metrics *InstanceMetrics) {
	if metrics != nil {
		metrics.CollectedAt = time.Now()
		metrics.UptimeSeconds = int64(time.Since(r.StartTime).Seconds())
		r.Metrics = metrics
	}
}

// UpdateHealth updates the health status
func (r *InstanceRegistry) UpdateHealth(health *HealthStatus) {
	if health != nil {
		health.LastCheck = time.Now()
		r.HealthStatus = health

		// Update instance status based on health
		switch health.Overall {
		case "healthy":
			r.SetStatus(InstanceStatusRunning)
		case "degraded":
			r.SetStatus(InstanceStatusDegraded)
		case "unhealthy":
			r.SetStatus(InstanceStatusUnknown)
		}
	}
}
