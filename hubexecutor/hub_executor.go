package hubexecutor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// HubExecutor manages the execution of an integration hub configuration
type HubExecutor struct {
	mu              sync.RWMutex
	iLog            logger.Log
	hubName         string
	hub             *models.IntegrationHub
	state           *HubExecutorState
	destExecutor    DestinationExecutor
	handlers        []ProtocolHandler
	historyRecorder HistoryRecorder
	onMessage       func(payload *MessagePayload)
	onError         func(err error, source string)
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewHubExecutor creates a new hub executor for a specific hub configuration
func NewHubExecutor(config HubExecutorConfig) *HubExecutor {
	ctx, cancel := context.WithCancel(context.Background())

	executor := &HubExecutor{
		iLog:            logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "HubExecutor"},
		hubName:         config.HubName,
		hub:             config.Hub,
		state:           NewHubExecutorState(),
		destExecutor:    config.DestinationExec,
		handlers:        make([]ProtocolHandler, 0),
		historyRecorder: config.HistoryRecorder,
		onMessage:       config.OnMessageReceive,
		onError:         config.OnError,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Use default destination executor if none provided
	if executor.destExecutor == nil {
		executor.destExecutor = NewDestinationExecutor()
	}

	return executor
}

// Start starts the hub executor and all configured handlers
func (e *HubExecutor) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	fmt.Printf("HubExecutor %s: Start() called\n", e.hubName)

	if e.state.GetStatus() == StatusRunning {
		return fmt.Errorf("hub executor %s is already running", e.hubName)
	}

	e.state.SetStatus(StatusStarting)
	e.iLog.Info(fmt.Sprintf("Starting HubExecutor for hub: %s (ID: %s, Version: %d)", e.hubName, e.hub.ID, e.hub.Version))

	// Log configuration summary
	if e.hub.Inbound != nil {
		e.iLog.Info(fmt.Sprintf("HubExecutor %s: Inbound direction found (Enabled: %v, ProtocolGroups: %d)", e.hubName, e.hub.Inbound.Enabled, len(e.hub.Inbound.ProtocolGroups)))
		fmt.Printf("HubExecutor %s: Inbound struct: %+v\n", e.hubName, e.hub.Inbound)
	} else {
		e.iLog.Warn(fmt.Sprintf("HubExecutor %s: Inbound configuration is nil", e.hubName))
		fmt.Printf("HubExecutor %s: Inbound is NIL\n", e.hubName)
	}

	if e.hub.Outbound != nil {
		e.iLog.Info(fmt.Sprintf("HubExecutor %s: Outbound direction found (Enabled: %v, ProtocolGroups: %d)", e.hubName, e.hub.Outbound.Enabled, len(e.hub.Outbound.ProtocolGroups)))
	} else {
		e.iLog.Warn(fmt.Sprintf("HubExecutor %s: Outbound configuration is nil", e.hubName))
	}

	// Initialize handlers based on hub configuration
	if err := e.initializeHandlers(); err != nil {
		e.state.SetStatus(StatusError)
		return fmt.Errorf("failed to initialize handlers: %v", err)
	}

	// Start all handlers
	for _, handler := range e.handlers {
		if err := handler.Start(e.ctx); err != nil {
			e.iLog.Error(fmt.Sprintf("Failed to start handler %s: %v", handler.Name(), err))
			// Continue starting other handlers
		} else {
			e.state.mu.Lock()
			e.state.Handlers[handler.Name()] = handler
			e.state.ActiveHandlers++
			e.state.mu.Unlock()
		}
	}

	now := time.Now()
	e.state.mu.Lock()
	e.state.StartTime = &now
	e.state.mu.Unlock()

	e.state.SetStatus(StatusRunning)
	e.iLog.Info(fmt.Sprintf("HubExecutor %s started with %d active handlers", e.hubName, e.state.ActiveHandlers))

	return nil
}

// Stop stops the hub executor and all handlers
func (e *HubExecutor) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state.GetStatus() == StatusStopped {
		return nil
	}

	e.state.SetStatus(StatusStopping)
	e.iLog.Info(fmt.Sprintf("Stopping HubExecutor: %s", e.hubName))

	// Cancel context to signal all handlers to stop
	e.cancel()

	// Stop all handlers
	for _, handler := range e.handlers {
		if err := handler.Stop(e.ctx); err != nil {
			e.iLog.Error(fmt.Sprintf("Error stopping handler %s: %v", handler.Name(), err))
		}
	}

	e.state.mu.Lock()
	e.state.Handlers = make(map[string]ProtocolHandler)
	e.state.ActiveHandlers = 0
	e.state.mu.Unlock()

	e.state.SetStatus(StatusStopped)
	e.iLog.Info(fmt.Sprintf("HubExecutor %s stopped", e.hubName))

	return nil
}

// Status returns the current executor status
func (e *HubExecutor) Status() ExecutorStatus {
	return e.state.GetStatus()
}

// GetInfo returns information about the executor
func (e *HubExecutor) GetInfo() ExecutorInfo {
	e.state.mu.RLock()
	defer e.state.mu.RUnlock()

	handlers := make([]HandlerInfo, 0, len(e.state.Handlers))
	for _, handler := range e.state.Handlers {
		handlers = append(handlers, HandlerInfo{
			Name:     handler.Name(),
			Type:     handler.Type(),
			Status:   handler.Status(),
			Protocol: string(handler.Type()),
		})
	}

	return ExecutorInfo{
		HubName:        e.hubName,
		HubID:          e.hub.ID,
		Version:        e.hub.Version,
		Status:         e.state.Status,
		StartTime:      e.state.StartTime,
		LastActivity:   e.state.LastActivity,
		MessageCount:   e.state.MessageCount,
		ErrorCount:     e.state.ErrorCount,
		ActiveHandlers: e.state.ActiveHandlers,
		Handlers:       handlers,
	}
}

// initializeHandlers creates handlers based on hub configuration
func (e *HubExecutor) initializeHandlers() error {
	e.handlers = make([]ProtocolHandler, 0)

	// Process inbound protocol groups
	if e.hub.Inbound != nil {
		e.iLog.Info(fmt.Sprintf("HubExecutor %s: processing %d inbound protocol groups", e.hubName, len(e.hub.Inbound.ProtocolGroups)))
		for i := range e.hub.Inbound.ProtocolGroups {
			pg := &e.hub.Inbound.ProtocolGroups[i]
			if !pg.Enabled {
				e.iLog.Info(fmt.Sprintf("HubExecutor %s: skipping disabled inbound protocol group %s", e.hubName, pg.Name))
				continue
			}

			e.iLog.Info(fmt.Sprintf("HubExecutor %s: loading inbound protocol group '%s' (Protocol: %s) with %d endpoints", e.hubName, pg.Name, pg.Protocol, len(pg.Endpoints)))
			handler, err := e.createProtocolHandler(pg, "inbound")
			if err != nil {
				e.iLog.Error(fmt.Sprintf("Failed to create handler for protocol group %s: %v", pg.Name, err))
				continue
			}
			if handler != nil {
				e.handlers = append(e.handlers, handler)
			}
		}
	}

	// Process outbound protocol groups (for monitoring/client connections)
	if e.hub.Outbound != nil {
		e.iLog.Info(fmt.Sprintf("HubExecutor %s: processing %d outbound protocol groups", e.hubName, len(e.hub.Outbound.ProtocolGroups)))
		for i := range e.hub.Outbound.ProtocolGroups {
			pg := &e.hub.Outbound.ProtocolGroups[i]
			if !pg.Enabled {
				e.iLog.Info(fmt.Sprintf("HubExecutor %s: skipping disabled outbound protocol group %s", e.hubName, pg.Name))
				continue
			}

			e.iLog.Info(fmt.Sprintf("HubExecutor %s: loading outbound protocol group '%s' (Protocol: %s) with %d endpoints", e.hubName, pg.Name, pg.Protocol, len(pg.Endpoints)))
			handler, err := e.createProtocolHandler(pg, "outbound")
			if err != nil {
				e.iLog.Error(fmt.Sprintf("Failed to create handler for protocol group %s: %v", pg.Name, err))
				continue
			}
			if handler != nil {
				e.handlers = append(e.handlers, handler)
			}
		}
	}

	e.iLog.Debug(fmt.Sprintf("Initialized %d handlers for hub %s", len(e.handlers), e.hub.Name))
	return nil
}

// createProtocolHandler creates a handler based on the protocol type
func (e *HubExecutor) createProtocolHandler(pg *models.HubSimpleProtocolGroup, direction string) (ProtocolHandler, error) {
	protocolType := e.mapProtocol(pg.Protocol)

	e.iLog.Debug(fmt.Sprintf("Creating %s handler for protocol: %s (%s)", direction, pg.Protocol, pg.Name))

	switch protocolType {
	case ProtocolREST, ProtocolSOAP, ProtocolGraphQL:
		return e.createWebServerHandler(pg, protocolType, direction)
	case ProtocolMQTT:
		return e.createMQTTHandler(pg, direction)
	case ProtocolKafka:
		return e.createKafkaHandler(pg, direction)
	case ProtocolActiveMQ:
		return e.createActiveMQHandler(pg, direction)
	case ProtocolOPCUA:
		return e.createOPCUAHandler(pg, direction)
	case ProtocolTCP:
		return e.createTCPHandler(pg, direction)
	case ProtocolSignalR:
		return e.createSignalRHandler(pg, direction)
	case ProtocolMCP:
		return e.createMCPHandler(pg, direction)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", pg.Protocol)
	}
}

// mapProtocol maps protocol string to ProtocolType
func (e *HubExecutor) mapProtocol(protocol string) ProtocolType {
	switch protocol {
	case "REST API", "REST", "HTTP", "HTTPS":
		return ProtocolREST
	case "SOAP API", "SOAP":
		return ProtocolSOAP
	case "GraphQL":
		return ProtocolGraphQL
	case "MQTT Client", "MQTT":
		return ProtocolMQTT
	case "Kafka":
		return ProtocolKafka
	case "ActiveMQ", "AMQP":
		return ProtocolActiveMQ
	case "OPC UA", "OPCUA":
		return ProtocolOPCUA
	case "TCP":
		return ProtocolTCP
	case "SignalR":
		return ProtocolSignalR
	case "MCP":
		return ProtocolMCP
	default:
		return ProtocolType(protocol)
	}
}

// createWebServerHandler creates a web server handler for REST/SOAP/GraphQL
func (e *HubExecutor) createWebServerHandler(pg *models.HubSimpleProtocolGroup, protocolType ProtocolType, direction string) (ProtocolHandler, error) {
	// Only create web server for inbound direction
	if direction != "inbound" {
		e.iLog.Debug(fmt.Sprintf("Skipping web server handler for outbound protocol group: %s", pg.Name))
		return nil, nil
	}

	// Get port: check endpoints first, then base_config
	port := 8080
	foundPort := false
	for _, ep := range pg.Endpoints {
		if ep.Port > 0 {
			port = ep.Port
			foundPort = true
			break
		}
	}

	if !foundPort && pg.BaseConfig != nil {
		if p, ok := pg.BaseConfig["port"]; ok {
			switch v := p.(type) {
			case float64:
				port = int(v)
			case int:
				port = v
			case string:
				fmt.Sscanf(v, "%d", &port)
			}
		}
	}

	e.iLog.Debug(fmt.Sprintf("Creating web server for %s on port %d", pg.Name, port))

	// Create the web server handler using our handlers package
	handler := NewWebServerHandlerFromProtocolGroup(WebServerFromProtocolGroupConfig{
		ProtocolGroup: pg,
		Protocol:      protocolType,
		Port:          port,
		DestExecutor:  e.destExecutor,
		OnMessage:     e.handleMessage,
		OnComplete:    e.CompleteTransaction,
		OnFail:        e.FailTransaction,
		OnError:       e.handleError,
		State:         e.state,
	})

	return handler, nil
}

// createMQTTHandler creates an MQTT handler
func (e *HubExecutor) createMQTTHandler(pg *models.HubSimpleProtocolGroup, direction string) (ProtocolHandler, error) {
	if pg.BrokerConfig == nil {
		return nil, fmt.Errorf("broker config required for MQTT protocol")
	}

	handler := NewMQTTHandler(MQTTHandlerConfig{
		Name:          pg.Name,
		ProtocolGroup: pg,
		BrokerConfig:  pg.BrokerConfig,
		DestExecutor:  e.destExecutor,
		OnMessage:     e.handleMessage,
		OnComplete:    e.CompleteTransaction,
		OnFail:        e.FailTransaction,
		OnError:       e.handleError,
		State:         e.state,
	})

	return handler, nil
}

// createKafkaHandler creates a Kafka handler
func (e *HubExecutor) createKafkaHandler(pg *models.HubSimpleProtocolGroup, direction string) (ProtocolHandler, error) {
	if pg.BrokerConfig == nil {
		return nil, fmt.Errorf("broker config required for Kafka protocol")
	}

	handler := NewKafkaHandler(KafkaHandlerConfig{
		Name:          pg.Name,
		ProtocolGroup: pg,
		BrokerConfig:  pg.BrokerConfig,
		DestExecutor:  e.destExecutor,
		OnMessage:     e.handleMessage,
		OnComplete:    e.CompleteTransaction,
		OnFail:        e.FailTransaction,
		OnError:       e.handleError,
		State:         e.state,
	})

	return handler, nil
}

// createActiveMQHandler creates an ActiveMQ handler
func (e *HubExecutor) createActiveMQHandler(pg *models.HubSimpleProtocolGroup, direction string) (ProtocolHandler, error) {
	if pg.BrokerConfig == nil {
		return nil, fmt.Errorf("broker config required for ActiveMQ protocol")
	}

	handler := NewActiveMQHandler(ActiveMQHandlerConfig{
		Name:          pg.Name,
		ProtocolGroup: pg,
		BrokerConfig:  pg.BrokerConfig,
		DestExecutor:  e.destExecutor,
		OnMessage:     e.handleMessage,
		OnError:       e.handleError,
		State:         e.state,
	})

	return handler, nil
}

// createOPCUAHandler creates an OPC UA handler
func (e *HubExecutor) createOPCUAHandler(pg *models.HubSimpleProtocolGroup, direction string) (ProtocolHandler, error) {
	handler := NewOPCUAHandler(OPCUAHandlerConfig{
		Name:          pg.Name,
		ProtocolGroup: pg,
		DestExecutor:  e.destExecutor,
		OnMessage:     e.handleMessage,
		OnError:       e.handleError,
		State:         e.state,
	})

	return handler, nil
}

// createTCPHandler creates a TCP handler
func (e *HubExecutor) createTCPHandler(pg *models.HubSimpleProtocolGroup, direction string) (ProtocolHandler, error) {
	handler := NewTCPHandler(TCPHandlerConfig{
		Name:          pg.Name,
		ProtocolGroup: pg,
		Direction:     direction,
		DestExecutor:  e.destExecutor,
		OnMessage:     e.handleMessage,
		OnError:       e.handleError,
		State:         e.state,
	})

	return handler, nil
}

// createSignalRHandler creates a SignalR handler
func (e *HubExecutor) createSignalRHandler(pg *models.HubSimpleProtocolGroup, direction string) (ProtocolHandler, error) {
	handler := NewSignalRHandler(SignalRHandlerConfig{
		Name:          pg.Name,
		ProtocolGroup: pg,
		DestExecutor:  e.destExecutor,
		OnMessage:     e.handleMessage,
		OnError:       e.handleError,
		State:         e.state,
	})

	return handler, nil
}

// createMCPHandler creates an MCP client handler for outbound MCP tool calls.
func (e *HubExecutor) createMCPHandler(pg *models.HubSimpleProtocolGroup, direction string) (ProtocolHandler, error) {
	if pg.MCPConfig == nil {
		return nil, fmt.Errorf("mcp_config is required for MCP protocol group '%s'", pg.Name)
	}

	handler := NewMCPClientHandler(MCPClientHandlerConfig{
		Name:          pg.Name,
		ProtocolGroup: pg,
		MCPConfig:     pg.MCPConfig,
		Direction:     direction,
		DestExecutor:  e.destExecutor,
		OnMessage:     e.handleMessage,
		OnError:       e.handleError,
		State:         e.state,
	})
	return handler, nil
}

// handleMessage handles incoming messages from handlers
func (e *HubExecutor) handleMessage(payload *MessagePayload) {
	e.state.IncrementMessageCount()

	// Handle history recording
	if e.historyRecorder != nil {
		e.iLog.Info(fmt.Sprintf("HubExecutor %s: Received message for endpoint ID: '%s'", e.hubName, payload.EndpointID))
		endpoint := e.findEndpoint(payload.EndpointID)
		e.iLog.Info(fmt.Sprintf("HubExecutor %s: found endpoint %v for ID %s", e.hubName, endpoint != nil, payload.EndpointID))
		if endpoint != nil && endpoint.ShouldTrackHistory() {
			e.iLog.Info(fmt.Sprintf("HubExecutor %s: recording history for endpoint %s", e.hubName, endpoint.Name))

			// Use direction from payload, default to "inbound" if not specified
			direction := payload.Direction
			if direction == "" {
				direction = "inbound"
			}

			history := &models.IntHubHistory{
				HubID:        e.hub.ID,
				HubName:      e.hub.Name,
				Direction:    direction,
				Protocol:     string(payload.Protocol),
				EndpointID:   endpoint.ID,
				EndpointName: endpoint.Name,
				Path:         payload.Path,
				Method:       payload.Method,
				Payload:      string(payload.Body),
				PayloadSize:  len(payload.Body),
				StartTime:    payload.Timestamp,
				Status:       models.IntHubHistoryStatusProcessing,
				Metadata:     payload.Metadata,
			}

			// Capture protocol group info if possible
			pg := e.findProtocolGroupForEndpoint(endpoint.ID)
			if pg != nil {
				history.ProtocolGroupID = pg.ID
				history.ProtocolGroupName = pg.Name
			}

			id, err := e.historyRecorder.RecordTransaction(context.Background(), history)
			if err == nil {
				payload.TransactionID = id
			} else {
				e.iLog.Error(fmt.Sprintf("Failed to record history: %v", err))
			}
		}
	}

	if e.onMessage != nil {
		e.onMessage(payload)
	}
}

// CompleteTransaction marks a transaction as completed
func (e *HubExecutor) CompleteTransaction(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int) {
	if payload.TransactionID != "" && e.historyRecorder != nil {
		err := e.historyRecorder.CompleteTransaction(ctx, payload.TransactionID, response, mappedData, responseStatus)
		if err != nil {
			e.iLog.Error(fmt.Sprintf("Failed to complete history transaction %s: %v", payload.TransactionID, err))
		}
	}
}

// FailTransaction marks a transaction as failed
func (e *HubExecutor) FailTransaction(ctx context.Context, payload *MessagePayload, errorMessage string) {
	if payload.TransactionID != "" && e.historyRecorder != nil {
		err := e.historyRecorder.FailTransaction(ctx, payload.TransactionID, errorMessage)
		if err != nil {
			e.iLog.Error(fmt.Sprintf("Failed to fail history transaction %s: %v", payload.TransactionID, err))
		}
	}
}

// findEndpoint finds an endpoint by ID in the hub config
func (e *HubExecutor) findEndpoint(endpointID string) *models.HubSimpleEndpoint {
	if e.hub.Inbound != nil {
		for _, pg := range e.hub.Inbound.ProtocolGroups {
			for i := range pg.Endpoints {
				if pg.Endpoints[i].ID == endpointID {
					return &pg.Endpoints[i]
				}
			}
		}
	}
	if e.hub.Outbound != nil {
		for _, pg := range e.hub.Outbound.ProtocolGroups {
			for i := range pg.Endpoints {
				if pg.Endpoints[i].ID == endpointID {
					return &pg.Endpoints[i]
				}
			}
		}
	}
	return nil
}

// findProtocolGroupForEndpoint finds the protocol group that contains the given endpoint
func (e *HubExecutor) findProtocolGroupForEndpoint(endpointID string) *models.HubSimpleProtocolGroup {
	if e.hub.Inbound != nil {
		for i := range e.hub.Inbound.ProtocolGroups {
			for _, ep := range e.hub.Inbound.ProtocolGroups[i].Endpoints {
				if ep.ID == endpointID {
					return &e.hub.Inbound.ProtocolGroups[i]
				}
			}
		}
	}
	if e.hub.Outbound != nil {
		for i := range e.hub.Outbound.ProtocolGroups {
			for _, ep := range e.hub.Outbound.ProtocolGroups[i].Endpoints {
				if ep.ID == endpointID {
					return &e.hub.Outbound.ProtocolGroups[i]
				}
			}
		}
	}
	return nil
}

// handleError handles errors from handlers
func (e *HubExecutor) handleError(err error, source string) {
	e.state.IncrementErrorCount()
	e.iLog.Error(fmt.Sprintf("Error from %s: %v", source, err))

	if e.onError != nil {
		e.onError(err, source)
	}
}

// GetHub returns the hub configuration
func (e *HubExecutor) GetHub() *models.IntegrationHub {
	return e.hub
}

// GetHubName returns the hub name
func (e *HubExecutor) GetHubName() string {
	return e.hubName
}
