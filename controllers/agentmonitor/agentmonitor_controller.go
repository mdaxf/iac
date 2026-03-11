package agentmonitor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/services"
)

// AgentMonitorController exposes read-only observability endpoints for the
// Agent Monitor service.  All endpoints are read-only; the only mutation is
// ResetSession which ends a channel-user conversation so the next message
// starts a fresh one.
//
// Routes (registered in apiconfig.json under path "api/agentmonitor"):
//
//	GET  /api/agentmonitor/status                            → GetStatus
//	GET  /api/agentmonitor/monitors                          → ListMonitors
//	GET  /api/agentmonitor/sessions                          → ListSessions
//	GET  /api/agentmonitor/sessions/:channelId               → ListSessionsByChannel
//	POST /api/agentmonitor/sessions/:channelId/:userId/reset → ResetSession
type AgentMonitorController struct{}

func monitorSvc() *services.AgentMonitorService {
	return services.GetGlobalAgentMonitorService()
}

func channelSvc() *services.AgentChannelService {
	return services.GetGlobalAgentChannelService()
}

// GetStatus returns the top-level AgentMonitor status summary.
//
// GET /api/agentmonitor/status
func (mc *AgentMonitorController) GetStatus(c *gin.Context) {
	svc := monitorSvc()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AgentMonitorService not initialized"})
		return
	}
	c.JSON(http.StatusOK, svc.GetStatus())
}

// ListMonitors returns all active channel_monitor (agent, channel) bindings.
// This shows which agents are "running" as channel monitors and which channels
// they are listening to — analogous to listing running integration hub executors.
//
// GET /api/agentmonitor/monitors
func (mc *AgentMonitorController) ListMonitors(c *gin.Context) {
	svc := monitorSvc()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AgentMonitorService not initialized"})
		return
	}
	c.JSON(http.StatusOK, svc.GetActiveMonitorInfos())
}

// ListSessions returns all currently tracked channel-user sessions.
//
// GET /api/agentmonitor/sessions
func (mc *AgentMonitorController) ListSessions(c *gin.Context) {
	svc := monitorSvc()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AgentMonitorService not initialized"})
		return
	}
	c.JSON(http.StatusOK, svc.GetActiveSessions())
}

// ListSessionsByChannel returns sessions for a specific channel binding.
//
// GET /api/agentmonitor/sessions/:channelId
func (mc *AgentMonitorController) ListSessionsByChannel(c *gin.Context) {
	svc := monitorSvc()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AgentMonitorService not initialized"})
		return
	}
	channelID := c.Param("channelId")
	c.JSON(http.StatusOK, svc.GetSessionsByChannel(channelID))
}

// ResetSession ends the tracked conversation for a channel user.
// The next inbound message from that user will start a fresh conversation.
//
// POST /api/agentmonitor/sessions/:channelId/:userId/reset
func (mc *AgentMonitorController) ResetSession(c *gin.Context) {
	svc := monitorSvc()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AgentMonitorService not initialized"})
		return
	}
	channelID := c.Param("channelId")
	userID := c.Param("userId")
	if channelID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channelId and userId are required"})
		return
	}
	if err := svc.ResetSession(channelSvc(), channelID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "session reset"})
}
