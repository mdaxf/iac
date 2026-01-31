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

package cluster

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	clusterservice "github.com/mdaxf/iac/services/cluster"
)

// ClusterController handles cluster-related API endpoints
type ClusterController struct {
	log logger.Log
}

// NewClusterController creates a new cluster controller
func NewClusterController() *ClusterController {
	return &ClusterController{
		log: logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ClusterController"},
	}
}

// RegisterRoutes registers the cluster routes
func (c *ClusterController) RegisterRoutes(router *gin.RouterGroup) {
	cluster := router.Group("/cluster")
	{
		cluster.GET("/status", c.GetClusterStatus)
		cluster.GET("/instances", c.GetInstances)
		cluster.GET("/instances/:id", c.GetInstance)
		cluster.GET("/discovery/:role", c.DiscoverByRole)
		cluster.POST("/heartbeat", c.Heartbeat)
		cluster.GET("/local", c.GetLocalInstance)
		cluster.GET("/config", c.GetClusterConfig)
	}
}

// GetClusterStatus returns the overall cluster status
// @Summary Get cluster status
// @Description Returns the overall status of the cluster including instance counts
// @Tags cluster
// @Produce json
// @Success 200 {object} models.ClusterStatus
// @Failure 500 {object} map[string]string
// @Router /api/cluster/status [get]
func (c *ClusterController) GetClusterStatus(ctx *gin.Context) {
	registryService := clusterservice.GetGlobalRegistryService()
	if registryService == nil {
		// Return standalone mode info when cluster is not enabled
		clusterConfig := config.GetClusterConfig()
		ctx.JSON(http.StatusOK, gin.H{
			"cluster_enabled": false,
			"mode":            "standalone",
			"instance_name":   config.GetInstanceName(),
			"environment":     config.GetEnvironment(),
			"message":         "Cluster mode is not enabled. Add cluster configuration to enable distributed features.",
			"total_instances": 1,
			"healthy":         1,
			"degraded":        0,
			"unhealthy":       0,
			"config_hint": map[string]interface{}{
				"cluster.enabled": clusterConfig != nil && clusterConfig.Enabled,
			},
		})
		return
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status, err := registryService.GetClusterStatus(reqCtx)
	if err != nil {
		c.log.Error("Failed to get cluster status: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get cluster status",
		})
		return
	}

	ctx.JSON(http.StatusOK, status)
}

// GetInstances returns all registered instances
// @Summary Get all instances
// @Description Returns a list of all registered instances in the cluster
// @Tags cluster
// @Produce json
// @Success 200 {array} models.InstanceRegistry
// @Failure 500 {object} map[string]string
// @Router /api/cluster/instances [get]
func (c *ClusterController) GetInstances(ctx *gin.Context) {
	registryService := clusterservice.GetGlobalRegistryService()
	if registryService == nil {
		// Return local instance info in standalone mode
		localInstance := map[string]interface{}{
			"instance_id":   config.GetInstanceName(),
			"instance_name": config.GetInstanceName(),
			"environment":   config.GetEnvironment(),
			"status":        "healthy",
			"roles":         config.GetRolesConfig().GetEnabledRoles(),
			"started_at":    time.Now().Format(time.RFC3339),
		}
		ctx.JSON(http.StatusOK, gin.H{
			"cluster_enabled": false,
			"instances":       []interface{}{localInstance},
			"count":           1,
		})
		return
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	instances, err := registryService.GetAllInstances(reqCtx)
	if err != nil {
		c.log.Error("Failed to get instances: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get instances",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"cluster_enabled": true,
		"instances":       instances,
		"count":           len(instances),
	})
}

// GetInstance returns a specific instance by ID
// @Summary Get instance by ID
// @Description Returns details of a specific instance
// @Tags cluster
// @Produce json
// @Param id path string true "Instance ID"
// @Success 200 {object} models.InstanceRegistry
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cluster/instances/{id} [get]
func (c *ClusterController) GetInstance(ctx *gin.Context) {
	instanceID := ctx.Param("id")
	if instanceID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Instance ID is required",
		})
		return
	}

	registryService := clusterservice.GetGlobalRegistryService()
	if registryService == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Cluster service not available",
		})
		return
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	instance, err := registryService.GetInstanceByID(reqCtx, instanceID)
	if err != nil {
		c.log.Error("Failed to get instance: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get instance",
		})
		return
	}

	if instance == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Instance not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, instance)
}

// DiscoverByRole returns instances with a specific role
// @Summary Discover instances by role
// @Description Returns instances that have a specific role enabled
// @Tags cluster
// @Produce json
// @Param role path string true "Role name (app, integration_hub, signalr, job_executor)"
// @Success 200 {array} models.InstanceRegistry
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cluster/discovery/{role} [get]
func (c *ClusterController) DiscoverByRole(ctx *gin.Context) {
	roleStr := ctx.Param("role")
	if roleStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Role is required",
		})
		return
	}

	// Convert role string to InstanceRole
	var role models.InstanceRole
	switch roleStr {
	case "app":
		role = models.InstanceRoleApp
	case "integration_hub":
		role = models.InstanceRoleIntegrationHub
	case "signalr":
		role = models.InstanceRoleSignalR
	case "job_executor":
		role = models.InstanceRoleJobExecutor
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":       "Invalid role",
			"valid_roles": []string{"app", "integration_hub", "signalr", "job_executor"},
		})
		return
	}

	registryService := clusterservice.GetGlobalRegistryService()
	if registryService == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Cluster service not available",
		})
		return
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	instances, err := registryService.GetInstancesByRole(reqCtx, role)
	if err != nil {
		c.log.Error("Failed to discover instances: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to discover instances",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"role":      roleStr,
		"instances": instances,
		"count":     len(instances),
	})
}

// Heartbeat receives a heartbeat from an instance
// @Summary Receive heartbeat
// @Description Receives and processes a heartbeat from an instance
// @Tags cluster
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/cluster/heartbeat [post]
func (c *ClusterController) Heartbeat(ctx *gin.Context) {
	// This endpoint can be used by external instances to send heartbeats
	// For now, the internal heartbeat mechanism handles this
	ctx.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GetLocalInstance returns information about the local instance
// @Summary Get local instance info
// @Description Returns information about this IAC instance
// @Tags cluster
// @Produce json
// @Success 200 {object} models.InstanceRegistry
// @Router /api/cluster/local [get]
func (c *ClusterController) GetLocalInstance(ctx *gin.Context) {
	registryService := clusterservice.GetGlobalRegistryService()
	if registryService == nil {
		// Return basic info even if cluster is not enabled
		ctx.JSON(http.StatusOK, gin.H{
			"cluster_enabled": false,
			"instance_name":   config.GetInstanceName(),
			"environment":     config.GetEnvironment(),
		})
		return
	}

	localInstance := registryService.GetLocalInstance()
	if localInstance == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Local instance not registered",
		})
		return
	}

	ctx.JSON(http.StatusOK, localInstance)
}

// GetClusterConfig returns the cluster configuration
// @Summary Get cluster configuration
// @Description Returns the current cluster configuration
// @Tags cluster
// @Produce json
// @Success 200 {object} config.ClusterConfig
// @Router /api/cluster/config [get]
func (c *ClusterController) GetClusterConfig(ctx *gin.Context) {
	clusterConfig := config.GetClusterConfig()

	// Remove sensitive information
	safeConfig := map[string]interface{}{
		"enabled":      clusterConfig.Enabled,
		"mode":         clusterConfig.Mode,
		"instance_name": clusterConfig.InstanceName,
		"environment":  clusterConfig.Environment,
	}

	if clusterConfig.Registry != nil {
		safeConfig["registry"] = map[string]interface{}{
			"heartbeat_interval": clusterConfig.Registry.HeartbeatInterval,
			"heartbeat_ttl":      clusterConfig.Registry.HeartbeatTTL,
			"signalr_discovery":  clusterConfig.Registry.SignalRDiscovery,
		}
	}

	if clusterConfig.Session != nil {
		safeConfig["session"] = map[string]interface{}{
			"distributed":      clusterConfig.Session.Distributed,
			"backend":          clusterConfig.Session.Backend,
			"fallback_backend": clusterConfig.Session.FallbackBackend,
			"ttl":              clusterConfig.Session.TTL,
			"auto_extend":      clusterConfig.Session.AutoExtend,
		}
	}

	ctx.JSON(http.StatusOK, safeConfig)
}
