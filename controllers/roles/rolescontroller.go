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

package roles

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
)

// RolesController handles API requests for instance role management
type RolesController struct {
	iLog logger.Log
}

// NewRolesController creates a new roles controller
func NewRolesController() *RolesController {
	return &RolesController{
		iLog: logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "RolesController"},
	}
}

// RegisterRoutes registers the roles API routes
func (c *RolesController) RegisterRoutes(router *gin.RouterGroup) {
	roles := router.Group("/roles")
	{
		roles.GET("/status", c.GetRoleStatus)
		roles.GET("/config", c.GetRolesConfig)
		roles.GET("/enabled", c.GetEnabledRoles)
	}
}

// GetRoleStatus handles GET /roles/status
// @Summary Get instance role status
// @Description Returns the current status of all configured roles
// @Tags Roles
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /roles/status [get]
func (c *RolesController) GetRoleStatus(ctx *gin.Context) {
	initializer := config.GetGlobalRoleInitializer()
	if initializer == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"success":     true,
			"initialized": false,
			"message":     "Role initializer not configured",
			"roles":       config.GetRolesConfig().GetEnabledRoles(),
		})
		return
	}

	status := initializer.GetRoleStatus()
	status["success"] = true

	ctx.JSON(http.StatusOK, status)
}

// GetRolesConfig handles GET /roles/config
// @Summary Get roles configuration
// @Description Returns the full roles configuration
// @Tags Roles
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /roles/config [get]
func (c *RolesController) GetRolesConfig(ctx *gin.Context) {
	rolesConfig := config.GetRolesConfig()

	response := gin.H{
		"success": true,
		"config":  rolesConfig,
	}

	ctx.JSON(http.StatusOK, response)
}

// GetEnabledRoles handles GET /roles/enabled
// @Summary Get enabled roles
// @Description Returns a list of enabled role types with worker status
// @Tags Roles
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /roles/enabled [get]
func (c *RolesController) GetEnabledRoles(ctx *gin.Context) {
	rolesConfig := config.GetRolesConfig()
	enabledRoles := rolesConfig.GetEnabledRoles()
	initializer := config.GetGlobalRoleInitializer()

	roleDetails := make([]map[string]interface{}, 0)
	for _, roleType := range enabledRoles {
		detail := map[string]interface{}{
			"type": string(roleType),
		}

		// Add worker status if initializer is available
		if initializer != nil {
			detail["running"] = initializer.IsRoleRunning(roleType)
			if worker := initializer.GetRoleWorkerStatus(roleType); worker != nil {
				detail["started_at"] = worker.StartedAt
				if worker.Error != nil {
					detail["error"] = worker.Error.Error()
				}
			}
		}

		switch roleType {
		case config.RoleApp:
			appConfig := config.GetAppRoleConfig()
			if appConfig != nil {
				detail["port"] = appConfig.Port
				detail["enable_portal"] = appConfig.EnablePortal
				detail["enable_static_files"] = appConfig.EnableStaticFiles
			}

		case config.RoleIntegrationHub:
			hubConfig := config.GetIntegrationHubRoleConfig()
			if hubConfig != nil {
				detail["hub_names"] = hubConfig.HubNames
				detail["auto_start"] = hubConfig.AutoStart
				detail["enable_api"] = hubConfig.EnableAPI
			}

		case config.RoleSignalR:
			signalrConfig := config.GetSignalRRoleConfig()
			if signalrConfig != nil {
				detail["port"] = signalrConfig.Port
				detail["hub"] = signalrConfig.Hub
			}

		case config.RoleJobExecutor:
			jobConfig := config.GetJobExecutorRoleConfig()
			if jobConfig != nil {
				detail["workers"] = jobConfig.Workers
				detail["enable_scheduler"] = jobConfig.EnableScheduler
			}
		}

		roleDetails = append(roleDetails, detail)
	}

	// Get running roles count
	runningCount := 0
	if initializer != nil {
		runningCount = len(initializer.GetRunningRoles())
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":       true,
		"roles":         roleDetails,
		"count":         len(enabledRoles),
		"running_count": runningCount,
	})
}
