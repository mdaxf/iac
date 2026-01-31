package integrationhub

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// IntegrationHubController handles simplified single-document integration hub HTTP requests
type IntegrationHubController struct {
}

// getHubService creates a new IntegrationHubService with MongoDB connection
func (c *IntegrationHubController) getHubService() (*services.IntegrationHubService, error) {
	factory := documents.GetDocFactory()
	if factory == nil {
		return nil, fmt.Errorf("document factory not available")
	}

	docDB, err := factory.GetDB("default")
	if err != nil || docDB == nil {
		return nil, fmt.Errorf("document database not available: %w", err)
	}

	return services.NewIntegrationHubService(docDB), nil
}

// =====================================================
// Integration Hub CRUD Endpoints
// =====================================================

// ListHubs handles GET /api/integration-hub
func (c *IntegrationHubController) ListHubs(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntegrationHubController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("integrationhub.ListHubs", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	hubService, err := c.getHubService()
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get hub service: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Service not available"})
		return
	}

	hubs, err := hubService.ListHubs(ctx.Request.Context())
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to list hubs: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": hubs, "count": len(hubs)})
}

// GetHub handles GET /api/integration-hub/:id
func (c *IntegrationHubController) GetHub(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntegrationHubController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("integrationhub.GetHub", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	hubID := ctx.Param("id")
	if hubID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "hub ID is required"})
		return
	}

	hubService, err := c.getHubService()
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get hub service: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Service not available"})
		return
	}

	hub, err := hubService.GetHub(ctx.Request.Context(), hubID)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get hub: %v", err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": hub})
}

// SaveHub handles POST /api/integration-hub and PUT /api/integration-hub/:id
// Saves the entire hub configuration (create or update)
func (c *IntegrationHubController) SaveHub(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntegrationHubController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("integrationhub.SaveHub", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	// Get ID from URL if present (for PUT requests)
	hubID := ctx.Param("id")

	var hub models.IntegrationHub
	if err := ctx.ShouldBindJSON(&hub); err != nil {
		iLog.Error(fmt.Sprintf("Failed to bind JSON: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If ID is in URL, use it (for PUT)
	if hubID != "" {
		hub.ID = hubID
	}

	hubService, err := c.getHubService()
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get hub service: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Service not available"})
		return
	}

	err = hubService.SaveHub(ctx.Request.Context(), &hub, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to save hub: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Saved integration hub: %s", hub.ID))
	ctx.JSON(http.StatusOK, gin.H{"data": hub, "message": "Hub saved successfully"})
}

// DeleteHub handles DELETE /api/integration-hub/:id
func (c *IntegrationHubController) DeleteHub(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntegrationHubController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("integrationhub.DeleteHub", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	hubID := ctx.Param("id")
	if hubID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "hub ID is required"})
		return
	}

	hubService, err := c.getHubService()
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get hub service: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Service not available"})
		return
	}

	err = hubService.DeleteHub(ctx.Request.Context(), hubID)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to delete hub: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Deleted integration hub: %s", hubID))
	ctx.JSON(http.StatusOK, gin.H{"message": "Hub deleted successfully"})
}

// =====================================================
// Utility Endpoints
// =====================================================

// GetAvailableProtocols handles GET /api/integration/protocols
func (c *IntegrationHubController) GetAvailableProtocols(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntegrationHubController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("integrationhub.GetAvailableProtocols", elapsed)
	}()

	protocols := []string{
		"HTTP",
		"HTTPS",
		"REST",
		"SOAP",
		"MQTT",
		"AMQP",
		"Kafka",
		"WebSocket",
		"gRPC",
		"File",
		"FTP",
		"SFTP",
		"Database",
	}

	ctx.JSON(http.StatusOK, gin.H{"data": protocols})
}

// GetAvailableAuthTypes handles GET /api/integration/auth-types
func (c *IntegrationHubController) GetAvailableAuthTypes(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntegrationHubController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("integrationhub.GetAvailableAuthTypes", elapsed)
	}()

	authTypes := []string{
		"None",
		"Basic",
		"Bearer",
		"OAuth2",
		"API-Key",
		"Certificate",
		"JWT",
		"SAML",
	}

	ctx.JSON(http.StatusOK, gin.H{"data": authTypes})
}

// GetAvailableMessageTypes handles GET /api/integration/message-types
func (c *IntegrationHubController) GetAvailableMessageTypes(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntegrationHubController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("integrationhub.GetAvailableMessageTypes", elapsed)
	}()

	messageTypes := []string{
		"JSON",
		"XML",
		"CSV",
		"Binary",
		"Plain-Text",
		"HL7",
		"EDI",
		"Avro",
		"Protobuf",
	}

	ctx.JSON(http.StatusOK, gin.H{"data": messageTypes})
}

// GetAvailableStatuses handles GET /api/integration/statuses
func (c *IntegrationHubController) GetAvailableStatuses(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntegrationHubController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("integrationhub.GetAvailableStatuses", elapsed)
	}()

	statuses := []string{
		string(models.IntegrationHubStatusDev),
		string(models.IntegrationHubStatusTest),
		string(models.IntegrationHubStatusProduction),
		string(models.IntegrationHubStatusDeprecated),
	}

	ctx.JSON(http.StatusOK, gin.H{"data": statuses})
}

// ListHubsByName handles GET /api/integration-hub/name/:name
func (c *IntegrationHubController) ListHubsByName(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntegrationHubController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("integrationhub.ListHubsByName", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	hubService, err := c.getHubService()
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get hub service: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Service not available"})
		return
	}

	hubs, err := hubService.ListHubsByName(ctx.Request.Context(), name)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to list hubs: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": hubs, "count": len(hubs)})
}

// GetDefaultHub handles GET /api/integration-hub/name/:name/default
func (c *IntegrationHubController) GetDefaultHub(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntegrationHubController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("integrationhub.GetDefaultHub", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	hubService, err := c.getHubService()
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get hub service: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Service not available"})
		return
	}

	hub, err := hubService.GetDefaultHub(ctx.Request.Context(), name)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get default hub: %v", err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": hub})
}

// SetAsDefault handles PUT /api/integration-hub/:id/default
func (c *IntegrationHubController) SetAsDefault(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntegrationHubController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("integrationhub.SetAsDefault", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	hubID := ctx.Param("id")
	if hubID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "hub ID is required"})
		return
	}

	hubService, err := c.getHubService()
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get hub service: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Service not available"})
		return
	}

	err = hubService.SetAsDefault(ctx.Request.Context(), hubID, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to set as default: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Hub set as default successfully"})
}

// CreateRevision handles POST /api/integration-hub/:id/revision
func (c *IntegrationHubController) CreateRevision(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntegrationHubController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("integrationhub.CreateRevision", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	hubID := ctx.Param("id")
	if hubID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "hub ID is required"})
		return
	}

	hubService, err := c.getHubService()
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get hub service: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Service not available"})
		return
	}

	newHub, err := hubService.CreateRevision(ctx.Request.Context(), hubID, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to create revision: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Created revision for hub: %s, new ID: %s", hubID, newHub.ID))
	ctx.JSON(http.StatusOK, gin.H{"data": newHub, "message": "Revision created successfully"})
}

// UpdateStatus handles PUT /api/integration-hub/:id/status
func (c *IntegrationHubController) UpdateStatus(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "IntegrationHubController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("integrationhub.UpdateStatus", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	hubID := ctx.Param("id")
	if hubID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "hub ID is required"})
		return
	}

	var request struct {
		Status string `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		iLog.Error(fmt.Sprintf("Failed to bind JSON: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hubService, err := c.getHubService()
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get hub service: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Service not available"})
		return
	}

	status := models.IntegrationHubStatus(request.Status)
	err = hubService.UpdateStatus(ctx.Request.Context(), hubID, status, user)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to update status: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}
