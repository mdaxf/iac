package agentruntime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/services"
)

// AgentRuntimeController handles all agent runtime REST endpoints
type AgentRuntimeController struct{}

// ---- Agent Definition endpoints ----

// ListAgents GET /agentruntime/agents
func (ac *AgentRuntimeController) ListAgents(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.ListAgents", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	svc := services.GetGlobalAgentDefinitionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent service not initialized"})
		return
	}

	agents, err := svc.ListAgents()
	if err != nil {
		iLog.Error(fmt.Sprintf("ListAgents error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": agents})
}

// GetAgent GET /agentruntime/agents/:id
func (ac *AgentRuntimeController) GetAgent(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.GetAgent", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalAgentDefinitionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent service not initialized"})
		return
	}

	agent, err := svc.GetAgent(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": agent})
}

// CreateAgent POST /agentruntime/agents
func (ac *AgentRuntimeController) CreateAgent(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.CreateAgent", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	var req services.CreateAgentRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.CreatedBy = user

	svc := services.GetGlobalAgentDefinitionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent service not initialized"})
		return
	}

	agent, err := svc.CreateAgent(req)
	if err != nil {
		iLog.Error(fmt.Sprintf("CreateAgent error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": agent})
}

// UpdateAgent PUT /agentruntime/agents/:id
func (ac *AgentRuntimeController) UpdateAgent(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.UpdateAgent", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	var req services.UpdateAgentRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	svc := services.GetGlobalAgentDefinitionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent service not initialized"})
		return
	}

	agent, err := svc.UpdateAgent(id, req, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("UpdateAgent error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": agent})
}

// DeleteAgent DELETE /agentruntime/agents/:id
func (ac *AgentRuntimeController) DeleteAgent(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.DeleteAgent", time.Since(startTime)) }()

	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalAgentDefinitionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent service not initialized"})
		return
	}

	if err := svc.DeleteAgent(id, user); err != nil {
		iLog.Error(fmt.Sprintf("DeleteAgent error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "agent deleted"})
}

// ---- Skill catalog endpoints ----

// ListSkills GET /agentruntime/skills
func (ac *AgentRuntimeController) ListSkills(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.ListSkills", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	svc := services.GetGlobalAgentDefinitionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent service not initialized"})
		return
	}

	skills, err := svc.ListSkills()
	if err != nil {
		iLog.Error(fmt.Sprintf("ListSkills error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": skills})
}

// CreateSkill POST /agentruntime/skills
func (ac *AgentRuntimeController) CreateSkill(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.CreateSkill", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	var req services.CreateSkillRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.CreatedBy = user

	svc := services.GetGlobalAgentDefinitionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent service not initialized"})
		return
	}

	skill, err := svc.CreateSkill(req)
	if err != nil {
		iLog.Error(fmt.Sprintf("CreateSkill error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": skill})
}

// UpdateSkill PUT /agentruntime/skills/:id
func (ac *AgentRuntimeController) UpdateSkill(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.UpdateSkill", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	var req services.UpdateSkillRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	svc := services.GetGlobalAgentDefinitionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent service not initialized"})
		return
	}

	skill, err := svc.UpdateSkill(id, req, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("UpdateSkill error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": skill})
}

// DeleteSkill DELETE /agentruntime/skills/:id
func (ac *AgentRuntimeController) DeleteSkill(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.DeleteSkill", time.Since(startTime)) }()

	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalAgentDefinitionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent service not initialized"})
		return
	}

	if err := svc.DeleteSkill(id, user); err != nil {
		iLog.Error(fmt.Sprintf("DeleteSkill error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "skill deleted"})
}

// InstallSkill POST /agentruntime/skills/install
func (ac *AgentRuntimeController) InstallSkill(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.InstallSkill", time.Since(startTime)) }()

	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	var req services.InstallSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}
	req.InstalledBy = user

	svc := services.GetGlobalAgentDefinitionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent service not initialized"})
		return
	}

	skill, err := svc.InstallSkillFromURL(req)
	if err != nil {
		iLog.Error(fmt.Sprintf("InstallSkill error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": skill, "message": fmt.Sprintf("Skill '%s' installed successfully", skill.Name)})
}

// UninstallSkill DELETE /agentruntime/skills/:id/uninstall
func (ac *AgentRuntimeController) UninstallSkill(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.UninstallSkill", time.Since(startTime)) }()

	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalAgentDefinitionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent service not initialized"})
		return
	}

	if err := svc.UninstallSkill(id, user); err != nil {
		iLog.Error(fmt.Sprintf("UninstallSkill error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "skill uninstalled"})
}

// ListTools GET /agentruntime/tools
func (ac *AgentRuntimeController) ListTools(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.ListTools", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	registry := services.GetGlobalToolRegistry()
	if registry == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "tool registry not initialized"})
		return
	}

	tools := registry.GetToolDefinitions(nil)
	c.JSON(http.StatusOK, gin.H{"data": tools})
}

// ---- Chat endpoints ----

// CreateConversation POST /agentruntime/conversations
func (ac *AgentRuntimeController) CreateConversation(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.CreateConversation", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	var req struct {
		AgentDefinitionID string `json:"agentdefinitionid"`
		Title             string `json:"title"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	chatSvc := services.GetGlobalAgentChatService()
	if chatSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "chat service not initialized"})
		return
	}

	// user is the username string; we use it as both userID and createdBy for simplicity
	conv, err := chatSvc.CreateAgentConversation(req.AgentDefinitionID, user, req.Title, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("CreateConversation error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": conv})
}

// GetConversation GET /agentruntime/conversations/:id
func (ac *AgentRuntimeController) GetConversation(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.GetConversation", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	chatSvc := services.GetGlobalAgentChatService()
	if chatSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "chat service not initialized"})
		return
	}

	conv, err := chatSvc.GetConversation(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": conv})
}

// ListConversations GET /agentruntime/conversations?agentid=&userid=
func (ac *AgentRuntimeController) ListConversations(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.ListConversations", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	agentID := c.Query("agentid")
	userID := c.Query("userid")

	chatSvc := services.GetGlobalAgentChatService()
	if chatSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "chat service not initialized"})
		return
	}

	convs, err := chatSvc.ListAgentConversations(agentID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": convs})
}

// DeleteConversation DELETE /agentruntime/conversations/:id
func (ac *AgentRuntimeController) DeleteConversation(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.DeleteConversation", time.Since(startTime)) }()

	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	chatSvc := services.GetGlobalAgentChatService()
	if chatSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "chat service not initialized"})
		return
	}

	if err := chatSvc.DeleteConversation(id, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "conversation deleted"})
}

// ---- Run endpoints ----

// RunAgent POST /agentruntime/run
func (ac *AgentRuntimeController) RunAgent(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.RunAgent", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	var req struct {
		AgentDefinitionID string `json:"agentdefinitionid"`
		ConversationID    string `json:"conversationid"`
		TaskPrompt        string `json:"taskprompt"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	runner := services.GetGlobalAgentRunnerService()
	if runner == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent runner not initialized"})
		return
	}

	runID, err := runner.RunAsync(services.AgentRunRequest{
		AgentDefinitionID: req.AgentDefinitionID,
		ConversationID:    req.ConversationID,
		TaskPrompt:        req.TaskPrompt,
		RequestedBy:       user,
	})
	if err != nil {
		iLog.Error(fmt.Sprintf("RunAgent error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"runid": runID, "status": "accepted"})
}

// GetAgentsStatus GET /agentruntime/agents/status
func (ac *AgentRuntimeController) GetAgentsStatus(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.GetAgentsStatus", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	runner := services.GetGlobalAgentRunnerService()
	if runner == nil {
		c.JSON(http.StatusOK, gin.H{"data": map[string]string{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": runner.GetActiveRunsStatus()})
}

// ToggleAgent POST /agentruntime/agents/:id/toggle
func (ac *AgentRuntimeController) ToggleAgent(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.ToggleAgent", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	svc := services.GetGlobalAgentDefinitionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent service not initialized"})
		return
	}

	if err := svc.ToggleAgent(id, req.Enabled, user); err != nil {
		iLog.Error(fmt.Sprintf("ToggleAgent error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "agent updated"})
}

// AbortRun DELETE /agentruntime/run/:runid
func (ac *AgentRuntimeController) AbortRun(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.AbortRun", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	runID := c.Param("runid")
	runner := services.GetGlobalAgentRunnerService()
	if runner != nil {
		runner.AbortRun(runID)
	}
	c.JSON(http.StatusOK, gin.H{"message": "abort requested"})
}

// ---- Schedule endpoints ----

// ListSchedules GET /agentruntime/schedules?agentid=
func (ac *AgentRuntimeController) ListSchedules(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.ListSchedules", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	agentID := c.Query("agentid")
	runner := services.GetGlobalAgentRunnerService()
	if runner == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent runner not initialized"})
		return
	}

	schedules, err := runner.ListSchedules(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": schedules})
}

// CreateSchedule POST /agentruntime/schedules
func (ac *AgentRuntimeController) CreateSchedule(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.CreateSchedule", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	var req services.CreateScheduleRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.CreatedBy = user

	runner := services.GetGlobalAgentRunnerService()
	if runner == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent runner not initialized"})
		return
	}

	sched, err := runner.CreateSchedule(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": sched})
}

// ToggleSchedule PUT /agentruntime/schedules/:id/toggle
func (ac *AgentRuntimeController) ToggleSchedule(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.ToggleSchedule", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	runner := services.GetGlobalAgentRunnerService()
	if runner == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent runner not initialized"})
		return
	}

	if err := runner.ToggleSchedule(id, req.Enabled, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "schedule updated"})
}

// DeleteSchedule DELETE /agentruntime/schedules/:id
func (ac *AgentRuntimeController) DeleteSchedule(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.DeleteSchedule", time.Since(startTime)) }()

	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	runner := services.GetGlobalAgentRunnerService()
	if runner == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent runner not initialized"})
		return
	}

	if err := runner.DeleteSchedule(id, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "schedule deleted"})
}

// ---- MCP Server endpoints ----

// ListMCPServers GET /agentruntime/mcp
func (ac *AgentRuntimeController) ListMCPServers(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.ListMCPServers", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	svc := services.GetGlobalMCPServerService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MCP service not initialized"})
		return
	}

	servers, err := svc.ListMCPServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": servers})
}

// CreateMCPServer POST /agentruntime/mcp
func (ac *AgentRuntimeController) CreateMCPServer(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.CreateMCPServer", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	var req services.CreateMCPServerRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.CreatedBy = user

	svc := services.GetGlobalMCPServerService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MCP service not initialized"})
		return
	}

	srv, err := svc.CreateMCPServer(req)
	if err != nil {
		iLog.Error(fmt.Sprintf("CreateMCPServer error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": srv})
}

// GetMCPServer GET /agentruntime/mcp/:id
func (ac *AgentRuntimeController) GetMCPServer(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.GetMCPServer", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalMCPServerService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MCP service not initialized"})
		return
	}

	srv, err := svc.GetMCPServer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "MCP server not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": srv})
}

// UpdateMCPServer PUT /agentruntime/mcp/:id
func (ac *AgentRuntimeController) UpdateMCPServer(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.UpdateMCPServer", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	var req services.UpdateMCPServerRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	svc := services.GetGlobalMCPServerService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MCP service not initialized"})
		return
	}

	srv, err := svc.UpdateMCPServer(id, req, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("UpdateMCPServer error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": srv})
}

// DeleteMCPServer DELETE /agentruntime/mcp/:id
func (ac *AgentRuntimeController) DeleteMCPServer(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.DeleteMCPServer", time.Since(startTime)) }()

	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalMCPServerService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MCP service not initialized"})
		return
	}

	if err := svc.DeleteMCPServer(id, user); err != nil {
		iLog.Error(fmt.Sprintf("DeleteMCPServer error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "MCP server deleted"})
}

// ---- Memory endpoints ----

// ListMemory GET /agentruntime/memory?agentid=&layer=&priority=&archived=&sortby=
func (ac *AgentRuntimeController) ListMemory(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.ListMemory", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	agentID := c.Query("agentid")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentid is required"})
		return
	}

	svc := services.GetGlobalAgentMemoryService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "memory service not initialized"})
		return
	}

	opts := services.ListMemoryOptions{
		Layer:        c.Query("layer"),
		Priority:     c.Query("priority"),
		ShowArchived: c.Query("archived") == "true",
		SortBy:       c.Query("sortby"),
	}
	items, err := svc.ListMemory(agentID, opts)
	if err != nil {
		iLog.Error(fmt.Sprintf("ListMemory error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateMemory POST /agentruntime/memory
func (ac *AgentRuntimeController) CreateMemory(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.CreateMemory", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	var req struct {
		AgentDefinitionID string `json:"agentdefinitionid"`
		services.CreateMemoryRequest
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.CreatedBy = user

	svc := services.GetGlobalAgentMemoryService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "memory service not initialized"})
		return
	}

	mem, err := svc.CreateMemory(req.AgentDefinitionID, req.CreateMemoryRequest)
	if err != nil {
		iLog.Error(fmt.Sprintf("CreateMemory error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": mem})
}

// GetMemory GET /agentruntime/memory/:id
func (ac *AgentRuntimeController) GetMemory(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.GetMemory", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalAgentMemoryService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "memory service not initialized"})
		return
	}

	mem, err := svc.GetMemory(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "memory item not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": mem})
}

// UpdateMemory PUT /agentruntime/memory/:id
func (ac *AgentRuntimeController) UpdateMemory(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.UpdateMemory", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	var req services.UpdateMemoryRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	svc := services.GetGlobalAgentMemoryService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "memory service not initialized"})
		return
	}

	mem, err := svc.UpdateMemory(id, req, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("UpdateMemory error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": mem})
}

// DeleteMemory DELETE /agentruntime/memory/:id
func (ac *AgentRuntimeController) DeleteMemory(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.DeleteMemory", time.Since(startTime)) }()

	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalAgentMemoryService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "memory service not initialized"})
		return
	}

	if err := svc.DeleteMemory(id, user); err != nil {
		iLog.Error(fmt.Sprintf("DeleteMemory error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "memory item deleted"})
}

// ArchiveMemory POST /agentruntime/memory/:id/archive
func (ac *AgentRuntimeController) ArchiveMemory(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.ArchiveMemory", time.Since(startTime)) }()

	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalAgentMemoryService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "memory service not initialized"})
		return
	}

	if err := svc.ArchiveMemory(id, user); err != nil {
		iLog.Error(fmt.Sprintf("ArchiveMemory error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "memory item archived"})
}

// CleanMemory POST /agentruntime/memory/clean
func (ac *AgentRuntimeController) CleanMemory(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.CleanMemory", time.Since(startTime)) }()

	body, clientid, _, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	var req struct {
		AgentDefinitionID string `json:"agentdefinitionid"`
		Priority          string `json:"priority"`
		Layer             string `json:"layer"`
		ArchivedOnly      bool   `json:"archived_only"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.AgentDefinitionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentdefinitionid is required"})
		return
	}

	svc := services.GetGlobalAgentMemoryService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "memory service not initialized"})
		return
	}

	count, err := svc.CleanMemory(req.AgentDefinitionID, services.CleanOptions{
		Priority:     req.Priority,
		Layer:        req.Layer,
		ArchivedOnly: req.ArchivedOnly,
	})
	if err != nil {
		iLog.Error(fmt.Sprintf("CleanMemory error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": count})
}

// ---- Sandbox endpoints ----

// ListSandboxes GET /agentruntime/sandboxes
func (ac *AgentRuntimeController) ListSandboxes(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.ListSandboxes", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	mgr := services.GetGlobalAgentSandboxManager()
	if mgr == nil {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": mgr.ListActive()})
}

// SendToSandbox POST /agentruntime/sandboxes/:runid/send
func (ac *AgentRuntimeController) SendToSandbox(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentRuntime"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentRuntime.SendToSandbox", time.Since(startTime)) }()

	body, clientid, _, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	runID := c.Param("runid")
	var req struct {
		Topic   string `json:"topic"`
		Payload string `json:"payload"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	mgr := services.GetGlobalAgentSandboxManager()
	if mgr == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sandbox manager not initialized"})
		return
	}

	if err := mgr.Send(runID, services.SandboxMessage{
		Topic:   req.Topic,
		Payload: req.Payload,
	}); err != nil {
		iLog.Error(fmt.Sprintf("SendToSandbox error: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "message sent"})
}
