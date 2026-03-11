package agentgateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// AgentGatewayController handles all agent gateway REST endpoints
// including IAC-native CRUD/execute routes and Google A2A protocol routes.
type AgentGatewayController struct{}

// ─── IAC REST: CRUD ────────────────────────────────────────────────────────────

// List GET /api/agentgateways
func (gc *AgentGatewayController) List(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentGateway"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentGateway.List", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	svc := services.GetGlobalAgentGatewayService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent gateway service not initialized"})
		return
	}

	gateways, err := svc.GetAll()
	if err != nil {
		iLog.Error(fmt.Sprintf("List error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gateways})
}

// Create POST /api/agentgateways
func (gc *AgentGatewayController) Create(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentGateway"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentGateway.Create", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	var doc models.AgentGatewayDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	doc.CreatedBy = user
	doc.ModifiedBy = user

	svc := services.GetGlobalAgentGatewayService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent gateway service not initialized"})
		return
	}

	if err := svc.Create(&doc); err != nil {
		iLog.Error(fmt.Sprintf("Create error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": doc})
}

// GetByID GET /api/agentgateways/:id
func (gc *AgentGatewayController) GetByID(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentGateway"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentGateway.GetByID", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalAgentGatewayService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent gateway service not initialized"})
		return
	}

	gateway, err := svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "gateway not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gateway})
}

// Update PUT /api/agentgateways/:id
func (gc *AgentGatewayController) Update(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentGateway"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentGateway.Update", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	var doc models.AgentGatewayDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	doc.ID = id
	doc.ModifiedBy = user

	svc := services.GetGlobalAgentGatewayService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent gateway service not initialized"})
		return
	}

	if err := svc.Update(id, &doc); err != nil {
		iLog.Error(fmt.Sprintf("Update error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": doc})
}

// Delete DELETE /api/agentgateways/:id
func (gc *AgentGatewayController) Delete(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentGateway"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentGateway.Delete", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalAgentGatewayService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent gateway service not initialized"})
		return
	}

	if err := svc.Delete(id); err != nil {
		iLog.Error(fmt.Sprintf("Delete error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "gateway deleted"})
}

// ─── IAC REST: Agent Card & Execute ───────────────────────────────────────────

// GetAgentCard GET /api/agentgateways/:id/agent-card
// Returns the Google A2A agent card for a specific gateway.
func (gc *AgentGatewayController) GetAgentCard(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentGateway"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentGateway.GetAgentCard", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalAgentGatewayService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent gateway service not initialized"})
		return
	}

	gateway, err := svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "gateway not found"})
		return
	}
	c.JSON(http.StatusOK, buildA2ACard(gateway))
}

// Execute POST /api/agentgateways/:id/execute
// Synchronous IAC-native call: runs the gateway and returns the response immediately.
func (gc *AgentGatewayController) Execute(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentGateway"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentGateway.Execute", time.Since(startTime)) }()

	body, clientid, _, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	var req struct {
		Input string `json:"input"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.Input == "" {
		// Try raw body as input
		req.Input = string(body)
	}

	svc := services.GetGlobalAgentGatewayService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent gateway service not initialized"})
		return
	}

	response, err := svc.ExecuteSync(c.Request.Context(), id, req.Input)
	if err != nil {
		iLog.Error(fmt.Sprintf("Execute error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": response})
}

// ─── Google A2A Protocol ───────────────────────────────────────────────────────

// GlobalAgentCard GET /.well-known/agent.json
// Returns an A2A-compliant agent card listing all enabled gateways as skills.
func (gc *AgentGatewayController) GlobalAgentCard(c *gin.Context) {
	svc := services.GetGlobalAgentGatewayService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent gateway service not initialized"})
		return
	}

	gateways, err := svc.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	skills := make([]map[string]interface{}, 0)
	for _, gw := range gateways {
		if !gw.Enabled {
			continue
		}
		skill := map[string]interface{}{
			"id":          gw.ID,
			"name":        gw.Name,
			"description": gw.Description,
		}
		skills = append(skills, skill)
	}

	card := map[string]interface{}{
		"name":        "IAC Agent Gateway",
		"description": "IAC Agent Gateway — wraps agent workflows behind the Google A2A protocol",
		"url":         "/api/a2a",
		"version":     "1.0.0",
		"skills":      skills,
	}
	c.JSON(http.StatusOK, card)
}

// A2ASubmitTask POST /api/a2a/:gatewayId
// Accepts a Google A2A task request, creates a task, and starts async execution.
func (gc *AgentGatewayController) A2ASubmitTask(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentGateway"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentGateway.A2ASubmitTask", time.Since(startTime)) }()

	gatewayID := c.Param("gatewayId")

	// Parse Google A2A request body:
	// { "id": "client-id", "message": { "role": "user", "parts": [{ "text": "..." }] } }
	var req struct {
		ID      string `json:"id"`
		Message struct {
			Role  string `json:"role"`
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid A2A request body"})
		return
	}

	// Extract text from first part
	input := ""
	if len(req.Message.Parts) > 0 {
		input = req.Message.Parts[0].Text
	}

	svc := services.GetGlobalAgentGatewayService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent gateway service not initialized"})
		return
	}

	taskID, err := svc.ExecuteAsync(c.Request.Context(), gatewayID, input)
	if err != nil {
		iLog.Error(fmt.Sprintf("A2ASubmitTask error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Google A2A response shape
	c.JSON(http.StatusOK, map[string]interface{}{
		"id":     taskID,
		"status": map[string]interface{}{"state": "submitted"},
	})
}

// A2AGetTask GET /api/a2a/:gatewayId/tasks/:taskId
// Polls the task status; on completion returns the response in A2A artifact format.
func (gc *AgentGatewayController) A2AGetTask(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentGateway"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentGateway.A2AGetTask", time.Since(startTime)) }()

	gatewayID := c.Param("gatewayId")
	taskID := c.Param("taskId")

	svc := services.GetGlobalAgentGatewayService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent gateway service not initialized"})
		return
	}

	task, err := svc.GetTask(c.Request.Context(), gatewayID, taskID)
	if err != nil {
		iLog.Error(fmt.Sprintf("A2AGetTask error: %v", err))
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	resp := map[string]interface{}{
		"id": task.ID,
		"status": map[string]interface{}{
			"state": task.Status,
		},
	}

	// Add response message and artifacts when completed
	if task.Status == "completed" && task.Output != "" {
		resp["status"] = map[string]interface{}{
			"state": "completed",
			"message": map[string]interface{}{
				"role":  "agent",
				"parts": []map[string]interface{}{{"text": task.Output}},
			},
		}
		resp["artifacts"] = []map[string]interface{}{
			{"parts": []map[string]interface{}{{"text": task.Output}}},
		}
	}

	if task.Status == "failed" && task.Error != "" {
		resp["status"].(map[string]interface{})["message"] = map[string]interface{}{
			"role":  "agent",
			"parts": []map[string]interface{}{{"text": fmt.Sprintf("Error: %s", task.Error)}},
		}
	}

	c.JSON(http.StatusOK, resp)
}

// A2ACancelTask POST /api/a2a/:gatewayId/tasks/:taskId/cancel
func (gc *AgentGatewayController) A2ACancelTask(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentGateway"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentGateway.A2ACancelTask", time.Since(startTime)) }()

	gatewayID := c.Param("gatewayId")
	taskID := c.Param("taskId")

	svc := services.GetGlobalAgentGatewayService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent gateway service not initialized"})
		return
	}

	if err := svc.CancelTask(c.Request.Context(), gatewayID, taskID); err != nil {
		iLog.Error(fmt.Sprintf("A2ACancelTask error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id":     taskID,
		"status": map[string]interface{}{"state": "canceled"},
	})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// buildA2ACard converts a gateway document to an A2A agent card JSON object.
func buildA2ACard(gw *models.AgentGatewayDoc) map[string]interface{} {
	skills := make([]map[string]interface{}, 0, len(gw.AgentCard.Skills))
	for _, s := range gw.AgentCard.Skills {
		skills = append(skills, map[string]interface{}{
			"id":          s.ID,
			"name":        s.Name,
			"description": s.Description,
		})
	}
	return map[string]interface{}{
		"name":        gw.AgentCard.DisplayName,
		"description": gw.AgentCard.Description,
		"provider": map[string]interface{}{
			"organization": gw.AgentCard.Provider,
			"url":          gw.AgentCard.ProviderURL,
		},
		"url":     fmt.Sprintf("/api/a2a/%s", gw.ID),
		"version": "1.0.0",
		"skills":  skills,
	}
}
