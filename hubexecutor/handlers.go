package hubexecutor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// ============================================================================
// Web Server Handler (REST, SOAP, GraphQL)
// ============================================================================

// WebServerFromProtocolGroupConfig holds config for creating WebServerHandler from protocol group
type WebServerFromProtocolGroupConfig struct {
	ProtocolGroup *models.HubSimpleProtocolGroup
	Protocol      ProtocolType
	Port          int
	DestExecutor  DestinationExecutor
	OnMessage     func(payload *MessagePayload)
	OnComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	OnFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	OnError       func(err error, source string)
	State         *HubExecutorState
}

// WebServerHandler handles REST, SOAP, and GraphQL protocols by creating HTTP servers
type WebServerHandler struct {
	mu            sync.RWMutex
	iLog          logger.Log
	name          string
	protocol      ProtocolType
	protocolGroup *models.HubSimpleProtocolGroup
	status        ExecutorStatus
	server        *http.Server
	router        *gin.Engine
	port          int
	destExecutor  DestinationExecutor
	onMessage     func(payload *MessagePayload)
	onComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	onFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	onError       func(err error, source string)
	state         *HubExecutorState
}

// NewWebServerHandlerFromProtocolGroup creates a web server handler from protocol group config
func NewWebServerHandlerFromProtocolGroup(config WebServerFromProtocolGroupConfig) *WebServerHandler {
	name := config.ProtocolGroup.Name
	if name == "" {
		name = fmt.Sprintf("webserver-%s", config.ProtocolGroup.ID)
	}

	handler := &WebServerHandler{
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "WebServerHandler"},
		name:          name,
		protocol:      config.Protocol,
		protocolGroup: config.ProtocolGroup,
		status:        StatusStopped,
		port:          config.Port,
		destExecutor:  config.DestExecutor,
		onMessage:     config.OnMessage,
		onComplete:    config.OnComplete,
		onFail:        config.OnFail,
		onError:       config.OnError,
		state:         config.State,
	}

	handler.iLog.Info(fmt.Sprintf("WebServerHandler %s initialized for protocol %s on port %d with %d endpoints",
		name, config.Protocol, config.Port, len(config.ProtocolGroup.Endpoints)))

	return handler
}

// Start starts the web server
func (h *WebServerHandler) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusRunning {
		return nil
	}

	h.status = StatusStarting
	h.iLog.Debug(fmt.Sprintf("Starting WebServerHandler: %s on port %d", h.name, h.port))

	// Create Gin router in release mode
	gin.SetMode(gin.ReleaseMode)
	h.router = gin.New()
	h.router.Use(gin.Recovery())

	// Enable CORS for the web server
	h.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Token, session")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Register endpoints
	if h.protocolGroup != nil {
		for _, endpoint := range h.protocolGroup.Endpoints {
			if !endpoint.Enabled {
				continue
			}
			h.registerEndpoint(endpoint)
		}
	}

	// Create HTTP server
	h.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", h.port),
		Handler:      h.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		h.iLog.Info(fmt.Sprintf("WebServerHandler %s starting HTTP server on %s", h.name, h.server.Addr))
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.iLog.Error(fmt.Sprintf("WebServerHandler %s failed: %v", h.name, err))
			h.mu.Lock()
			h.status = StatusError
			h.mu.Unlock()
			if h.onError != nil {
				h.onError(err, h.name)
			}
		}
	}()

	h.status = StatusRunning
	return nil
}

// Stop stops the web server
func (h *WebServerHandler) Stop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusStopped {
		return nil
	}

	h.status = StatusStopping
	h.iLog.Debug(fmt.Sprintf("Stopping WebServerHandler: %s", h.name))

	if h.server != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := h.server.Shutdown(shutdownCtx); err != nil {
			h.iLog.Error(fmt.Sprintf("Error shutting down WebServerHandler %s: %v", h.name, err))
			return err
		}
	}

	h.status = StatusStopped
	return nil
}

// Status returns the current status
func (h *WebServerHandler) Status() ExecutorStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// Name returns the handler name
func (h *WebServerHandler) Name() string {
	return h.name
}

// Type returns the protocol type
func (h *WebServerHandler) Type() ProtocolType {
	return h.protocol
}

// registerEndpoint registers a single endpoint on the router
func (h *WebServerHandler) registerEndpoint(endpoint models.HubSimpleEndpoint) {
	// Protocol level base path
	basePath := ""
	if h.protocolGroup != nil && h.protocolGroup.BaseConfig != nil {
		basePath = getStringFromMap(h.protocolGroup.BaseConfig, "serverUrl")
		if basePath == "" {
			basePath = getStringFromMap(h.protocolGroup.BaseConfig, "base_path")
		}
		if basePath == "" {
			basePath = getStringFromMap(h.protocolGroup.BaseConfig, "path")
		}
	}

	// Clean basePath if it's a full URL
	if idx := strings.Index(basePath, "://"); idx != -1 {
		remaining := basePath[idx+3:]
		if pathIdx := strings.Index(remaining, "/"); pathIdx != -1 {
			basePath = remaining[pathIdx:]
		} else {
			basePath = ""
		}
	}

	// Endpoint subpath
	subPath := endpoint.Path
	if subPath == "" {
		subPath = "/" + endpoint.ID
	}

	// Combine paths
	fullPath := subPath
	if basePath != "" {
		b := strings.TrimSuffix(basePath, "/")
		e := strings.TrimPrefix(subPath, "/")
		if b == "" {
			fullPath = "/" + e
		} else {
			if !strings.HasPrefix(b, "/") {
				b = "/" + b
			}
			fullPath = b + "/" + e
		}
	}

	// Ensure fullPath starts with /
	if !strings.HasPrefix(fullPath, "/") {
		fullPath = "/" + fullPath
	}

	method := endpoint.Method
	if method == "" {
		method = "POST"
	}

	h.iLog.Info(fmt.Sprintf("WebServerHandler %s registering endpoint: %s http://[host]:%d%s", h.name, method, h.port, fullPath))

	handler := h.createEndpointHandler(endpoint)

	switch method {
	case "GET":
		h.router.GET(fullPath, handler)
	case "POST":
		h.router.POST(fullPath, handler)
	case "PUT":
		h.router.PUT(fullPath, handler)
	case "DELETE":
		h.router.DELETE(fullPath, handler)
	case "PATCH":
		h.router.PATCH(fullPath, handler)
	default:
		h.router.Any(fullPath, handler)
	}
}

// createEndpointHandler creates a Gin handler for an endpoint
func (h *WebServerHandler) createEndpointHandler(endpoint models.HubSimpleEndpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Read request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			h.iLog.Error(fmt.Sprintf("Error reading request body: %v", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Build headers map
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		// Create message payload
		payload := &MessagePayload{
			Protocol:    h.protocol,
			Direction:   "inbound",
			EndpointID:  endpoint.ID,
			Path:        c.Request.URL.Path,
			Method:      c.Request.Method,
			Headers:     headers,
			Body:        body,
			ContentType: c.ContentType(),
			Metadata: map[string]interface{}{
				"query_params": c.Request.URL.Query(),
				"remote_addr":  c.ClientIP(),
				"endpoint":     endpoint.Name,
			},
			Timestamp: time.Now(),
		}

		// Call message callback if set
		if h.onMessage != nil {
			h.iLog.Info(fmt.Sprintf("WebServerHandler %s: Calling onMessage callback for endpoint %s", h.name, endpoint.Name))
			h.onMessage(payload)
		} else {
			h.iLog.Warn(fmt.Sprintf("WebServerHandler %s: No onMessage callback configured", h.name))
		}

		// Update state
		if h.state != nil {
			h.state.IncrementMessageCount()
		}

		// Get destinations from endpoint config
		destinations := h.getEndpointDestinations(endpoint)

		// Execute destinations
		var lastResult *ExecutionResult
		allSuccess := true

		for _, dest := range destinations {
			if h.destExecutor != nil {
				result, execErr := h.destExecutor.Execute(c.Request.Context(), dest, payload)
				if execErr != nil {
					h.iLog.Error(fmt.Sprintf("Destination execution error: %v", execErr))
					allSuccess = false
					if h.state != nil {
						h.state.IncrementErrorCount()
					}
					if h.onError != nil {
						h.onError(execErr, fmt.Sprintf("%s/%s", h.name, endpoint.ID))
					}
				}
				lastResult = result
			}
		}

		// Build response
		duration := time.Since(startTime)

		if allSuccess {
			respData := ""
			if lastResult != nil && lastResult.Data != nil {
				if b, err := json.Marshal(lastResult.Data); err == nil {
					respData = string(b)
				}
			}

			if h.onComplete != nil {
				h.onComplete(c.Request.Context(), payload, "Success", respData, http.StatusOK)
			}

			response := gin.H{
				"success":  true,
				"message":  "Message processed successfully",
				"duration": duration.Milliseconds(),
			}
			if lastResult != nil && lastResult.Data != nil {
				response["data"] = lastResult.Data
			}
			c.JSON(http.StatusOK, response)
		} else {
			if h.onFail != nil {
				h.onFail(c.Request.Context(), payload, "Message processing failed")
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"success":  false,
				"message":  "Message processing failed",
				"duration": duration.Milliseconds(),
			})
		}
	}
}

// getEndpointDestinations extracts destinations from endpoint config
func (h *WebServerHandler) getEndpointDestinations(endpoint models.HubSimpleEndpoint) []Destination {
	var destinations []Destination

	// Check for destinations in override_config
	if destList, ok := endpoint.OverrideConfig["destinations"].([]interface{}); ok {
		for _, d := range destList {
			if destMap, ok := d.(map[string]interface{}); ok {
				dest := Destination{
					ID:     getStringFromMap(destMap, "id"),
					Type:   getStringFromMap(destMap, "type"),
					Target: getStringFromMap(destMap, "target"),
				}
				if configMap, ok := destMap["config"].(map[string]interface{}); ok {
					dest.Config = configMap
				}
				destinations = append(destinations, dest)
			}
		}
	}

	// Check handler configuration
	if endpoint.Handler != nil {
		if endpoint.Handler.HandlerType == "route" {
			for _, routeID := range endpoint.Handler.RouteIDs {
				destinations = append(destinations, Destination{
					ID:     routeID,
					Type:   "Job",
					Target: routeID,
				})
			}
		}
	}

	return destinations
}

// ============================================================================
// MQTT Handler
// ============================================================================

// MQTTHandlerConfig holds configuration for MQTT handler
type MQTTHandlerConfig struct {
	Name          string
	ProtocolGroup *models.HubSimpleProtocolGroup
	BrokerConfig  *models.MessageBrokerConfig
	DestExecutor  DestinationExecutor
	OnMessage     func(payload *MessagePayload)
	OnComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	OnFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	OnError       func(err error, source string)
	State         *HubExecutorState
}

// MQTTHandler handles MQTT protocol
type MQTTHandler struct {
	mu            sync.RWMutex
	iLog          logger.Log
	name          string
	protocolGroup *models.HubSimpleProtocolGroup
	brokerConfig  *models.MessageBrokerConfig
	status        ExecutorStatus
	client        mqtt.Client
	destExecutor  DestinationExecutor
	onMessage     func(payload *MessagePayload)
	onComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	onFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	onError       func(err error, source string)
	state         *HubExecutorState
	stopChan      chan struct{}
}

// NewMQTTHandler creates a new MQTT handler
func NewMQTTHandler(config MQTTHandlerConfig) *MQTTHandler {
	return &MQTTHandler{
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MQTTHandler"},
		name:          config.Name,
		protocolGroup: config.ProtocolGroup,
		brokerConfig:  config.BrokerConfig,
		status:        StatusStopped,
		destExecutor:  config.DestExecutor,
		onMessage:     config.OnMessage,
		onComplete:    config.OnComplete,
		onFail:        config.OnFail,
		onError:       config.OnError,
		state:         config.State,
		stopChan:      make(chan struct{}),
	}
}

// Start starts the MQTT handler
func (h *MQTTHandler) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusRunning {
		return nil
	}

	h.status = StatusStarting
	h.iLog.Debug(fmt.Sprintf("Starting MQTTHandler: %s", h.name))

	// Build broker URL
	brokerURL := h.brokerConfig.BrokerURL
	if brokerURL == "" {
		protocol := "tcp"
		if h.brokerConfig.UseTLS {
			protocol = "ssl"
		}
		brokerURL = fmt.Sprintf("%s://%s:%d", protocol, h.brokerConfig.BrokerHost, h.brokerConfig.BrokerPort)
	}

	// Create MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(h.brokerConfig.ClientID)
	if h.brokerConfig.ClientID == "" {
		opts.SetClientID(fmt.Sprintf("hub-%s-%s", h.name, uuid.New().String()[:8]))
	}
	opts.SetUsername(h.brokerConfig.Username)
	opts.SetPassword(h.brokerConfig.Password)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)

	opts.OnConnect = func(client mqtt.Client) {
		h.iLog.Info(fmt.Sprintf("MQTTHandler %s: SUCCESSFULLY CONNECTED to broker", h.name))
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		h.iLog.Error(fmt.Sprintf("MQTTHandler %s: CONNECTION LOST: %v", h.name, err))
	}

	if h.brokerConfig.KeepAlive > 0 {
		opts.SetKeepAlive(time.Duration(h.brokerConfig.KeepAlive) * time.Second)
	}

	// Create and connect client
	h.iLog.Debug(fmt.Sprintf("MQTTHandler %s connecting to broker: %s with client ID: %s", h.name, brokerURL, opts.ClientID))
	h.client = mqtt.NewClient(opts)
	if token := h.client.Connect(); token.Wait() && token.Error() != nil {
		h.status = StatusError
		h.iLog.Error(fmt.Sprintf("MQTTHandler %s connection failed: %v", h.name, token.Error()))
		return fmt.Errorf("failed to connect to MQTT broker %s: %v", brokerURL, token.Error())
	}

	// Subscribe to topics
	topics := h.brokerConfig.Topics
	if len(topics) == 0 {
		// Fallback: collect topics from endpoints
		for _, ep := range h.protocolGroup.Endpoints {
			if ep.Enabled && ep.Path != "" {
				topics = append(topics, models.BrokerTopic{Topic: ep.Path})
			}
		}
	}

	if len(topics) == 0 {
		h.iLog.Warn(fmt.Sprintf("MQTTHandler %s: no topics configured or found in endpoints", h.name))
	}

	for _, brokerTopic := range topics {
		qos := byte(h.brokerConfig.QoS)
		topicName := brokerTopic.Topic
		if token := h.client.Subscribe(topicName, qos, h.createMessageHandler(topicName)); token.Wait() && token.Error() != nil {
			h.iLog.Error(fmt.Sprintf("MQTTHandler %s failed to subscribe to topic %s: %v", h.name, topicName, token.Error()))
		} else {
			h.iLog.Info(fmt.Sprintf("MQTTHandler %s subscribed to topic: %s", h.name, topicName))
		}
	}

	h.status = StatusRunning
	h.iLog.Info(fmt.Sprintf("MQTTHandler %s successfully connected and initialized", h.name))

	return nil
}

// createMessageHandler creates a message handler for a topic
func (h *MQTTHandler) createMessageHandler(topic string) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		h.iLog.Debug(fmt.Sprintf("Received MQTT message on topic %s", msg.Topic()))

		// Create payload
		payload := &MessagePayload{
			Protocol:   ProtocolMQTT,
			Direction:  "inbound",
			EndpointID: topic,
			Path:       msg.Topic(),
			Body:       msg.Payload(),
			Metadata: map[string]interface{}{
				"topic":      msg.Topic(),
				"message_id": msg.MessageID(),
				"qos":        msg.Qos(),
				"retained":   msg.Retained(),
			},
			Timestamp: time.Now(),
		}

		// Update state
		if h.state != nil {
			h.state.IncrementMessageCount()
		}

		// Execute destinations
		h.executeDestinations(payload)
	}
}

// executeDestinations executes configured destinations for MQTT
func (h *MQTTHandler) executeDestinations(payload *MessagePayload) {
	if h.protocolGroup == nil || h.destExecutor == nil {
		return
	}

	// Find matching endpoint for the topic
	for _, endpoint := range h.protocolGroup.Endpoints {
		if !endpoint.Enabled {
			continue
		}

		// Check if endpoint handles this topic
		// Endpoint Path is the topic it matches
		if endpoint.Path != "" && endpoint.Path != payload.Path {
			continue
		}

		// Create a specific payload for this endpoint (to have correct ID in history)
		epPayload := *payload
		epPayload.EndpointID = endpoint.ID

		// Trigger history recording
		if h.onMessage != nil {
			h.onMessage(&epPayload)
		}

		// Execute destinations
		allSuccess := true
		if endpoint.Handler != nil {
			for _, routeID := range endpoint.Handler.RouteIDs {
				dest := Destination{
					ID:     routeID,
					Type:   "Job",
					Target: routeID,
				}

				if _, err := h.destExecutor.Execute(context.Background(), dest, &epPayload); err != nil {
					h.iLog.Error(fmt.Sprintf("Failed to execute destination %s: %v", routeID, err))
					allSuccess = false
					if h.state != nil {
						h.state.IncrementErrorCount()
					}
					if h.onError != nil {
						h.onError(err, h.name)
					}
				}
			}
		}

		if allSuccess {
			if h.onComplete != nil {
				h.onComplete(context.Background(), &epPayload, "Success", "", 0)
			}
		} else {
			if h.onFail != nil {
				h.onFail(context.Background(), &epPayload, "One or more destinations failed")
			}
		}
	}
}

// Stop stops the MQTT handler
func (h *MQTTHandler) Stop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusStopped {
		return nil
	}

	h.status = StatusStopping
	h.iLog.Debug(fmt.Sprintf("Stopping MQTTHandler: %s", h.name))

	close(h.stopChan)

	if h.client != nil && h.client.IsConnected() {
		h.client.Disconnect(250)
	}

	h.status = StatusStopped
	return nil
}

// Status returns the current status
func (h *MQTTHandler) Status() ExecutorStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// Name returns the handler name
func (h *MQTTHandler) Name() string {
	return h.name
}

// Type returns the protocol type
func (h *MQTTHandler) Type() ProtocolType {
	return ProtocolMQTT
}

// ============================================================================
// Kafka Handler
// ============================================================================

// KafkaHandlerConfig holds configuration for Kafka handler
type KafkaHandlerConfig struct {
	Name          string
	ProtocolGroup *models.HubSimpleProtocolGroup
	BrokerConfig  *models.MessageBrokerConfig
	DestExecutor  DestinationExecutor
	OnMessage     func(payload *MessagePayload)
	OnComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	OnFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	OnError       func(err error, source string)
	State         *HubExecutorState
}

// KafkaHandler handles Kafka protocol
type KafkaHandler struct {
	mu            sync.RWMutex
	iLog          logger.Log
	name          string
	protocolGroup *models.HubSimpleProtocolGroup
	brokerConfig  *models.MessageBrokerConfig
	status        ExecutorStatus
	consumer      sarama.Consumer
	destExecutor  DestinationExecutor
	onMessage     func(payload *MessagePayload)
	onComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	onFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	onError       func(err error, source string)
	state         *HubExecutorState
	stopChan      chan struct{}
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewKafkaHandler creates a new Kafka handler
func NewKafkaHandler(config KafkaHandlerConfig) *KafkaHandler {
	return &KafkaHandler{
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "KafkaHandler"},
		name:          config.Name,
		protocolGroup: config.ProtocolGroup,
		brokerConfig:  config.BrokerConfig,
		status:        StatusStopped,
		destExecutor:  config.DestExecutor,
		onMessage:     config.OnMessage,
		onComplete:    config.OnComplete,
		onFail:        config.OnFail,
		onError:       config.OnError,
		state:         config.State,
		stopChan:      make(chan struct{}),
	}
}

// Start starts the Kafka handler
func (h *KafkaHandler) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusRunning {
		return nil
	}

	h.status = StatusStarting
	h.iLog.Debug(fmt.Sprintf("Starting KafkaHandler: %s", h.name))

	h.ctx, h.cancel = context.WithCancel(ctx)

	// Build bootstrap servers
	servers := []string{fmt.Sprintf("%s:%d", h.brokerConfig.BrokerHost, h.brokerConfig.BrokerPort)}
	if h.brokerConfig.BrokerURL != "" {
		servers = []string{h.brokerConfig.BrokerURL}
	}

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Create consumer
	consumer, err := sarama.NewConsumer(servers, config)
	if err != nil {
		h.status = StatusError
		return fmt.Errorf("failed to create Kafka consumer: %v", err)
	}
	h.consumer = consumer

	// Subscribe to topics
	topics := h.brokerConfig.Topics
	if len(topics) == 0 {
		for _, ep := range h.protocolGroup.Endpoints {
			if ep.Enabled && ep.Path != "" {
				topics = append(topics, models.BrokerTopic{Topic: ep.Path})
			}
		}
	}

	for _, brokerTopic := range topics {
		go h.consumeTopic(brokerTopic.Topic)
	}

	h.status = StatusRunning
	h.iLog.Info(fmt.Sprintf("KafkaHandler %s connected to %v", h.name, servers))
	return nil
}

// consumeTopic consumes messages from a specific Kafka topic
func (h *KafkaHandler) consumeTopic(topic string) {
	// For simplicity, we consume all partitions.
	// In a real scenario, you'd use a Consumer Group for distributed consumption.
	partitions, err := h.consumer.Partitions(topic)
	if err != nil {
		h.iLog.Error(fmt.Sprintf("Failed to get partitions for topic %s: %v", topic, err))
		return
	}

	for _, partition := range partitions {
		pc, err := h.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			h.iLog.Error(fmt.Sprintf("Failed to consume partition %d for topic %s: %v", partition, topic, err))
			continue
		}

		go func(p sarama.PartitionConsumer, t string) {
			defer p.Close()
			for {
				select {
				case msg := <-p.Messages():
					h.handleKafkaMessage(msg)
				case err := <-p.Errors():
					h.iLog.Error(fmt.Sprintf("Kafka consumer error on topic %s: %v", t, err))
				case <-h.ctx.Done():
					return
				}
			}
		}(pc, topic)
	}
}

// handleKafkaMessage processes an individual Kafka message
func (h *KafkaHandler) handleKafkaMessage(msg *sarama.ConsumerMessage) {
	h.iLog.Debug(fmt.Sprintf("Received Kafka message on topic %s", msg.Topic))

	payload := &MessagePayload{
		Protocol:   ProtocolKafka,
		Direction:  "inbound",
		EndpointID: msg.Topic,
		Path:       msg.Topic,
		Body:       msg.Value,
		Metadata: map[string]interface{}{
			"topic":     msg.Topic,
			"partition": msg.Partition,
			"offset":    msg.Offset,
			"key":       string(msg.Key),
		},
		Timestamp: time.Now(),
	}

	if h.state != nil {
		h.state.IncrementMessageCount()
	}

	h.executeDestinations(payload)
}

// executeDestinations executes configured destinations for Kafka
func (h *KafkaHandler) executeDestinations(payload *MessagePayload) {
	if h.protocolGroup == nil || h.destExecutor == nil {
		return
	}

	for _, endpoint := range h.protocolGroup.Endpoints {
		if !endpoint.Enabled {
			continue
		}

		if endpoint.Path != "" && endpoint.Path != payload.Path {
			continue
		}

		// Create a specific payload for this endpoint (to have correct ID in history)
		epPayload := *payload
		epPayload.EndpointID = endpoint.ID

		// Trigger history recording
		if h.onMessage != nil {
			h.onMessage(&epPayload)
		}

		allSuccess := true
		if endpoint.Handler != nil {
			for _, routeID := range endpoint.Handler.RouteIDs {
				dest := Destination{
					ID:     routeID,
					Type:   "Job",
					Target: routeID,
				}

				if _, err := h.destExecutor.Execute(context.Background(), dest, &epPayload); err != nil {
					h.iLog.Error(fmt.Sprintf("Kafka destination execution error: %v", err))
					allSuccess = false
					if h.state != nil {
						h.state.IncrementErrorCount()
					}
					if h.onError != nil {
						h.onError(err, h.name)
					}
				}
			}
		}

		if allSuccess {
			if h.onComplete != nil {
				h.onComplete(context.Background(), &epPayload, "Success", "", 0)
			}
		} else {
			if h.onFail != nil {
				h.onFail(context.Background(), &epPayload, "One or more destinations failed")
			}
		}
	}
}

// Stop stops the Kafka handler
func (h *KafkaHandler) Stop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusStopped {
		return nil
	}

	h.status = StatusStopping
	h.iLog.Debug(fmt.Sprintf("Stopping KafkaHandler: %s", h.name))

	if h.cancel != nil {
		h.cancel()
	}

	if h.consumer != nil {
		if err := h.consumer.Close(); err != nil {
			h.iLog.Error(fmt.Sprintf("Error closing Kafka consumer: %v", err))
		}
	}

	h.status = StatusStopped
	return nil
}

// Status returns the current status
func (h *KafkaHandler) Status() ExecutorStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// Name returns the handler name
func (h *KafkaHandler) Name() string {
	return h.name
}

// Type returns the protocol type
func (h *KafkaHandler) Type() ProtocolType {
	return ProtocolKafka
}

// ============================================================================
// ActiveMQ Handler
// ============================================================================

// ActiveMQHandlerConfig holds configuration for ActiveMQ handler
type ActiveMQHandlerConfig struct {
	Name          string
	ProtocolGroup *models.HubSimpleProtocolGroup
	BrokerConfig  *models.MessageBrokerConfig
	DestExecutor  DestinationExecutor
	OnMessage     func(payload *MessagePayload)
	OnComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	OnFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	OnError       func(err error, source string)
	State         *HubExecutorState
}

// ActiveMQHandler handles ActiveMQ protocol
type ActiveMQHandler struct {
	mu            sync.RWMutex
	iLog          logger.Log
	name          string
	protocolGroup *models.HubSimpleProtocolGroup
	brokerConfig  *models.MessageBrokerConfig
	status        ExecutorStatus
	destExecutor  DestinationExecutor
	onMessage     func(payload *MessagePayload)
	onComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	onFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	onError       func(err error, source string)
	state         *HubExecutorState
	stopChan      chan struct{}
}

// NewActiveMQHandler creates a new ActiveMQ handler
func NewActiveMQHandler(config ActiveMQHandlerConfig) *ActiveMQHandler {
	return &ActiveMQHandler{
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "ActiveMQHandler"},
		name:          config.Name,
		protocolGroup: config.ProtocolGroup,
		brokerConfig:  config.BrokerConfig,
		status:        StatusStopped,
		destExecutor:  config.DestExecutor,
		onMessage:     config.OnMessage,
		onComplete:    config.OnComplete,
		onFail:        config.OnFail,
		onError:       config.OnError,
		state:         config.State,
		stopChan:      make(chan struct{}),
	}
}

// Start starts the ActiveMQ handler
func (h *ActiveMQHandler) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusRunning {
		return nil
	}

	h.status = StatusStarting
	h.iLog.Debug(fmt.Sprintf("Starting ActiveMQHandler: %s", h.name))

	// TODO: Implement ActiveMQ consumer using existing integration/activemq package

	h.iLog.Warn(fmt.Sprintf("ActiveMQHandler %s: ActiveMQ consumer integration pending - connect to %s:%d",
		h.name, h.brokerConfig.BrokerHost, h.brokerConfig.BrokerPort))

	h.status = StatusRunning
	return nil
}

// Stop stops the ActiveMQ handler
func (h *ActiveMQHandler) Stop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusStopped {
		return nil
	}

	h.status = StatusStopping
	h.iLog.Debug(fmt.Sprintf("Stopping ActiveMQHandler: %s", h.name))

	close(h.stopChan)

	h.status = StatusStopped
	return nil
}

// Status returns the current status
func (h *ActiveMQHandler) Status() ExecutorStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// Name returns the handler name
func (h *ActiveMQHandler) Name() string {
	return h.name
}

// Type returns the protocol type
func (h *ActiveMQHandler) Type() ProtocolType {
	return ProtocolActiveMQ
}

// ============================================================================
// OPC UA Handler
// ============================================================================

// OPCUAHandlerConfig holds configuration for OPC UA handler
type OPCUAHandlerConfig struct {
	Name          string
	ProtocolGroup *models.HubSimpleProtocolGroup
	DestExecutor  DestinationExecutor
	OnMessage     func(payload *MessagePayload)
	OnComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	OnFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	OnError       func(err error, source string)
	State         *HubExecutorState
}

// OPCUAHandler handles OPC UA protocol
type OPCUAHandler struct {
	mu            sync.RWMutex
	iLog          logger.Log
	name          string
	protocolGroup *models.HubSimpleProtocolGroup
	status        ExecutorStatus
	destExecutor  DestinationExecutor
	onMessage     func(payload *MessagePayload)
	onComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	onFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	onError       func(err error, source string)
	state         *HubExecutorState
	stopChan      chan struct{}
	tags          []OPCTag
	actions       []OPCAction
}

// OPCAction represents an action triggered by OPC UA data changes
type OPCAction struct {
	ID              string
	Name            string
	TriggerMethod   TriggerMethod
	MonitoredTags   []string
	ConditionScript string
	Destinations    []Destination
	IntervalMs      int
}

// NewOPCUAHandler creates a new OPC UA handler
func NewOPCUAHandler(config OPCUAHandlerConfig) *OPCUAHandler {
	handler := &OPCUAHandler{
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "OPCUAHandler"},
		name:          config.Name,
		protocolGroup: config.ProtocolGroup,
		status:        StatusStopped,
		destExecutor:  config.DestExecutor,
		onMessage:     config.OnMessage,
		onComplete:    config.OnComplete,
		onFail:        config.OnFail,
		onError:       config.OnError,
		state:         config.State,
		stopChan:      make(chan struct{}),
		tags:          make([]OPCTag, 0),
		actions:       make([]OPCAction, 0),
	}

	// Parse tags and actions from protocol group configuration
	handler.parseConfiguration()

	return handler
}

// parseConfiguration parses OPC UA tags and actions from the protocol group
func (h *OPCUAHandler) parseConfiguration() {
	if h.protocolGroup == nil {
		return
	}

	// Parse OPC tags from endpoints
	for _, endpoint := range h.protocolGroup.Endpoints {
		if !endpoint.Enabled {
			continue
		}

		// Check for OPC tags in override_config
		if tagsData, ok := endpoint.OverrideConfig["opc_tags"].([]interface{}); ok {
			for _, t := range tagsData {
				if tagMap, ok := t.(map[string]interface{}); ok {
					tag := OPCTag{
						ID:             getStringFromMap(tagMap, "id"),
						TagName:        getStringFromMap(tagMap, "tagName"),
						Address:        getStringFromMap(tagMap, "address"),
						DataType:       getStringFromMap(tagMap, "dataType"),
						IsArray:        getBoolFromMap(tagMap, "isArray"),
						ArraySize:      getIntFromMap(tagMap, "arraySize"),
						Permission:     getStringFromMap(tagMap, "permission"),
						SampleInterval: getIntFromMap(tagMap, "sampleInterval"),
						Deadband:       getFloatFromMap(tagMap, "deadband"),
						Description:    getStringFromMap(tagMap, "description"),
						Unit:           getStringFromMap(tagMap, "unit"),
						ScaleFactor:    getFloatFromMap(tagMap, "scaleFactor"),
						Offset:         getFloatFromMap(tagMap, "offset"),
						Enabled:        getBoolFromMap(tagMap, "enabled"),
					}
					h.tags = append(h.tags, tag)
				}
			}
		}

		// Check for action trigger configuration
		if triggerData, ok := endpoint.OverrideConfig["action_trigger"].(map[string]interface{}); ok {
			action := OPCAction{
				ID:              endpoint.ID,
				Name:            endpoint.Name,
				TriggerMethod:   TriggerMethod(getStringFromMap(triggerData, "method")),
				ConditionScript: getStringFromMap(triggerData, "conditionScript"),
				IntervalMs:      getIntFromMap(triggerData, "intervalMs"),
			}

			// Parse monitored tags
			if tags, ok := triggerData["monitoredTags"].([]interface{}); ok {
				for _, t := range tags {
					if tagName, ok := t.(string); ok {
						action.MonitoredTags = append(action.MonitoredTags, tagName)
					}
				}
			}

			// Parse destinations
			if destData, ok := endpoint.OverrideConfig["destinations"].([]interface{}); ok {
				for _, d := range destData {
					if destMap, ok := d.(map[string]interface{}); ok {
						dest := Destination{
							ID:     getStringFromMap(destMap, "id"),
							Type:   getStringFromMap(destMap, "type"),
							Target: getStringFromMap(destMap, "target"),
						}
						if configMap, ok := destMap["config"].(map[string]interface{}); ok {
							dest.Config = configMap
						}
						action.Destinations = append(action.Destinations, dest)
					}
				}
			}

			h.actions = append(h.actions, action)
		}
	}

	h.iLog.Debug(fmt.Sprintf("Parsed %d OPC tags and %d actions for handler %s", len(h.tags), len(h.actions), h.name))
}

// Start starts the OPC UA handler
func (h *OPCUAHandler) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusRunning {
		return nil
	}

	h.status = StatusStarting
	h.iLog.Debug(fmt.Sprintf("Starting OPCUAHandler: %s with %d tags", h.name, len(h.tags)))

	// Get OPC UA server configuration from protocol group
	serverURL := ""
	if h.protocolGroup != nil && h.protocolGroup.BaseConfig != nil {
		serverURL = getStringFromMap(h.protocolGroup.BaseConfig, "server_url")
	}

	if serverURL == "" {
		h.iLog.Warn(fmt.Sprintf("OPCUAHandler %s: No OPC UA server URL configured", h.name))
	}

	// TODO: Connect to OPC UA server using existing integration/opcclient package
	// For now, start interval-based and startup actions

	// Start startup actions
	h.executeStartupActions()

	// Start interval-based actions
	go h.runIntervalActions()

	h.status = StatusRunning
	h.iLog.Info(fmt.Sprintf("OPCUAHandler %s started", h.name))

	return nil
}

// executeStartupActions executes actions triggered at startup
func (h *OPCUAHandler) executeStartupActions() {
	for _, action := range h.actions {
		if action.TriggerMethod == TriggerStartup {
			h.iLog.Debug(fmt.Sprintf("Executing startup action: %s", action.Name))
			h.executeAction(action, nil, nil)
		}
	}
}

// runIntervalActions runs interval-based actions
func (h *OPCUAHandler) runIntervalActions() {
	for _, action := range h.actions {
		if action.TriggerMethod == TriggerInterval && action.IntervalMs > 0 {
			go h.runIntervalAction(action)
		}
	}
}

// runIntervalAction runs a single interval-based action
func (h *OPCUAHandler) runIntervalAction(action OPCAction) {
	ticker := time.NewTicker(time.Duration(action.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopChan:
			return
		case <-ticker.C:
			h.executeAction(action, nil, nil)
		}
	}
}

// executeAction executes an OPC action
func (h *OPCUAHandler) executeAction(action OPCAction, newValue, oldValue interface{}) {
	// Evaluate condition if specified
	if action.ConditionScript != "" {
		result, err := EvaluateCondition(action.ConditionScript, newValue, oldValue, "")
		if err != nil {
			h.iLog.Error(fmt.Sprintf("Error evaluating condition for action %s: %v", action.Name, err))
			return
		}
		if !result {
			return
		}
	}

	// Create payload
	payload := &MessagePayload{
		Protocol:   ProtocolOPCUA,
		Direction:  "inbound", // OPC UA actions are typically triggered by data changes (inbound)
		EndpointID: action.ID,
		Metadata: map[string]interface{}{
			"action_name":    action.Name,
			"trigger_method": string(action.TriggerMethod),
			"new_value":      newValue,
			"old_value":      oldValue,
		},
		Timestamp: time.Now(),
	}

	// Serialize values to body
	if newValue != nil {
		bodyData, _ := json.Marshal(map[string]interface{}{
			"new_value": newValue,
			"old_value": oldValue,
		})
		payload.Body = bodyData
	}

	// Call message callback
	if h.onMessage != nil {
		h.onMessage(payload)
	}

	// Update state
	if h.state != nil {
		h.state.IncrementMessageCount()
	}

	// Execute destinations
	allSuccess := true
	for _, dest := range action.Destinations {
		if h.destExecutor != nil {
			if _, err := h.destExecutor.Execute(context.Background(), dest, payload); err != nil {
				h.iLog.Error(fmt.Sprintf("Failed to execute destination %s: %v", dest.ID, err))
				allSuccess = false
				if h.state != nil {
					h.state.IncrementErrorCount()
				}
				if h.onError != nil {
					h.onError(err, h.name)
				}
			}
		}
	}

	if allSuccess {
		if h.onComplete != nil {
			h.onComplete(context.Background(), payload, "Success", "", 0)
		}
	} else {
		if h.onFail != nil {
			h.onFail(context.Background(), payload, "One or more destinations failed")
		}
	}
}

// Stop stops the OPC UA handler
func (h *OPCUAHandler) Stop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusStopped {
		return nil
	}

	h.status = StatusStopping
	h.iLog.Debug(fmt.Sprintf("Stopping OPCUAHandler: %s", h.name))

	// Execute shutdown actions
	for _, action := range h.actions {
		if action.TriggerMethod == TriggerShutdown {
			h.iLog.Debug(fmt.Sprintf("Executing shutdown action: %s", action.Name))
			h.executeAction(action, nil, nil)
		}
	}

	close(h.stopChan)

	h.status = StatusStopped
	return nil
}

// Status returns the current status
func (h *OPCUAHandler) Status() ExecutorStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// Name returns the handler name
func (h *OPCUAHandler) Name() string {
	return h.name
}

// Type returns the protocol type
func (h *OPCUAHandler) Type() ProtocolType {
	return ProtocolOPCUA
}

// ============================================================================
// TCP Handler
// ============================================================================

// TCPHandlerConfig holds configuration for TCP handler
type TCPHandlerConfig struct {
	Name          string
	ProtocolGroup *models.HubSimpleProtocolGroup
	Direction     string
	DestExecutor  DestinationExecutor
	OnMessage     func(payload *MessagePayload)
	OnComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	OnFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	OnError       func(err error, source string)
	State         *HubExecutorState
}

// TCPHandler handles TCP protocol
type TCPHandler struct {
	mu            sync.RWMutex
	iLog          logger.Log
	name          string
	protocolGroup *models.HubSimpleProtocolGroup
	direction     string
	status        ExecutorStatus
	destExecutor  DestinationExecutor
	onMessage     func(payload *MessagePayload)
	onComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	onFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	onError       func(err error, source string)
	state         *HubExecutorState
	stopChan      chan struct{}
}

// NewTCPHandler creates a new TCP handler
func NewTCPHandler(config TCPHandlerConfig) *TCPHandler {
	return &TCPHandler{
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "TCPHandler"},
		name:          config.Name,
		protocolGroup: config.ProtocolGroup,
		direction:     config.Direction,
		status:        StatusStopped,
		destExecutor:  config.DestExecutor,
		onMessage:     config.OnMessage,
		onComplete:    config.OnComplete,
		onFail:        config.OnFail,
		onError:       config.OnError,
		state:         config.State,
		stopChan:      make(chan struct{}),
	}
}

// Start starts the TCP handler
func (h *TCPHandler) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusRunning {
		return nil
	}

	h.status = StatusStarting
	h.iLog.Debug(fmt.Sprintf("Starting TCPHandler: %s (%s)", h.name, h.direction))

	// TODO: Implement TCP server/client using existing integration/tcp package

	h.iLog.Warn(fmt.Sprintf("TCPHandler %s: TCP integration pending", h.name))

	h.status = StatusRunning
	return nil
}

// Stop stops the TCP handler
func (h *TCPHandler) Stop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusStopped {
		return nil
	}

	h.status = StatusStopping
	close(h.stopChan)
	h.status = StatusStopped
	return nil
}

// Status returns the current status
func (h *TCPHandler) Status() ExecutorStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// Name returns the handler name
func (h *TCPHandler) Name() string {
	return h.name
}

// Type returns the protocol type
func (h *TCPHandler) Type() ProtocolType {
	return ProtocolTCP
}

// ============================================================================
// SignalR Handler
// ============================================================================

// SignalRHandlerConfig holds configuration for SignalR handler
type SignalRHandlerConfig struct {
	Name          string
	ProtocolGroup *models.HubSimpleProtocolGroup
	DestExecutor  DestinationExecutor
	OnMessage     func(payload *MessagePayload)
	OnComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	OnFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	OnError       func(err error, source string)
	State         *HubExecutorState
}

// SignalRHandler handles SignalR protocol
type SignalRHandler struct {
	mu            sync.RWMutex
	iLog          logger.Log
	name          string
	protocolGroup *models.HubSimpleProtocolGroup
	status        ExecutorStatus
	destExecutor  DestinationExecutor
	onMessage     func(payload *MessagePayload)
	onComplete    func(ctx context.Context, payload *MessagePayload, response string, mappedData string, responseStatus int)
	onFail        func(ctx context.Context, payload *MessagePayload, errorMessage string)
	onError       func(err error, source string)
	state         *HubExecutorState
	stopChan      chan struct{}
}

// NewSignalRHandler creates a new SignalR handler
func NewSignalRHandler(config SignalRHandlerConfig) *SignalRHandler {
	return &SignalRHandler{
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "SignalRHandler"},
		name:          config.Name,
		protocolGroup: config.ProtocolGroup,
		status:        StatusStopped,
		destExecutor:  config.DestExecutor,
		onMessage:     config.OnMessage,
		onComplete:    config.OnComplete,
		onFail:        config.OnFail,
		onError:       config.OnError,
		state:         config.State,
		stopChan:      make(chan struct{}),
	}
}

// Start starts the SignalR handler
func (h *SignalRHandler) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusRunning {
		return nil
	}

	h.status = StatusStarting
	h.iLog.Debug(fmt.Sprintf("Starting SignalRHandler: %s", h.name))

	// TODO: Implement SignalR client

	h.iLog.Warn(fmt.Sprintf("SignalRHandler %s: SignalR integration pending", h.name))

	h.status = StatusRunning
	return nil
}

// Stop stops the SignalR handler
func (h *SignalRHandler) Stop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusStopped {
		return nil
	}

	h.status = StatusStopping
	close(h.stopChan)
	h.status = StatusStopped
	return nil
}

// Status returns the current status
func (h *SignalRHandler) Status() ExecutorStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// Name returns the handler name
func (h *SignalRHandler) Name() string {
	return h.name
}

// Type returns the protocol type
func (h *SignalRHandler) Type() ProtocolType {
	return ProtocolSignalR
}

// ============================================================================
// Helper functions
// ============================================================================

func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getIntFromMap(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return 0
}

func getFloatFromMap(m map[string]interface{}, key string) float64 {
	switch v := m[key].(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	}
	return 0
}

func getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

// ============================================================================
// MCP Client Handler (Model Context Protocol – JSON-RPC 2.0 over HTTP)
// ============================================================================

// MCPClientHandlerConfig holds configuration for the MCP client handler.
type MCPClientHandlerConfig struct {
	Name          string
	ProtocolGroup *models.HubSimpleProtocolGroup
	MCPConfig     *models.MCPProtocolConfig
	Direction     string
	DestExecutor  DestinationExecutor
	OnMessage     func(payload *MessagePayload)
	OnError       func(err error, source string)
	State         *HubExecutorState
}

// MCPClientHandler connects to an MCP server (JSON-RPC 2.0 over HTTP) and
// can invoke named tools as part of Integration Hub outbound message routing.
type MCPClientHandler struct {
	mu            sync.RWMutex
	iLog          logger.Log
	name          string
	protocolGroup *models.HubSimpleProtocolGroup
	mcpConfig     *models.MCPProtocolConfig
	direction     string
	status        ExecutorStatus
	sessionID     string // Mcp-Session-Id obtained on initialize
	destExecutor  DestinationExecutor
	onMessage     func(payload *MessagePayload)
	onError       func(err error, source string)
	state         *HubExecutorState
}

// NewMCPClientHandler creates a new MCP client handler.
func NewMCPClientHandler(config MCPClientHandlerConfig) *MCPClientHandler {
	name := config.Name
	if name == "" {
		name = fmt.Sprintf("mcp-%s", config.ProtocolGroup.ID)
	}
	return &MCPClientHandler{
		iLog:          logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MCPClientHandler"},
		name:          name,
		protocolGroup: config.ProtocolGroup,
		mcpConfig:     config.MCPConfig,
		direction:     config.Direction,
		status:        StatusStopped,
		destExecutor:  config.DestExecutor,
		onMessage:     config.OnMessage,
		onError:       config.OnError,
		state:         config.State,
	}
}

// Start initializes the MCP session by sending the JSON-RPC 2.0 "initialize" request.
func (h *MCPClientHandler) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusRunning {
		return nil
	}
	h.status = StatusStarting

	cfg := h.mcpConfig
	if cfg == nil || cfg.ServerURL == "" {
		h.status = StatusError
		return fmt.Errorf("MCPClientHandler %s: server_url is required", h.name)
	}

	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	initCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build initialize request
	mcpPath := cfg.MCPPath
	if mcpPath == "" {
		mcpPath = "/mcp"
	}
	endpoint := cfg.ServerURL + mcpPath

	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "iac-integration-hub",
				"version": "1.0.0",
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequestWithContext(initCtx, http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		h.status = StatusError
		return fmt.Errorf("MCPClientHandler %s: build initialize request: %w", h.name, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range cfg.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		// Non-fatal: server may not require initialization
		h.iLog.Warn(fmt.Sprintf("MCPClientHandler %s: initialize failed (non-fatal): %v", h.name, err))
		h.status = StatusRunning
		return nil
	}
	defer resp.Body.Close()

	// Capture session ID if the server provides one
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		h.sessionID = sid
		h.iLog.Info(fmt.Sprintf("MCPClientHandler %s: MCP session established: %s", h.name, sid))
	}
	io.ReadAll(resp.Body) // drain

	h.status = StatusRunning
	h.iLog.Info(fmt.Sprintf("MCPClientHandler %s started (server: %s)", h.name, cfg.ServerURL))
	return nil
}

// Stop sends an MCP shutdown notification and marks the handler as stopped.
func (h *MCPClientHandler) Stop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == StatusStopped {
		return nil
	}

	cfg := h.mcpConfig
	if cfg != nil && cfg.ServerURL != "" && h.sessionID != "" {
		// Best-effort: send shutdown notification
		mcpPath := cfg.MCPPath
		if mcpPath == "" {
			mcpPath = "/mcp"
		}
		endpoint := cfg.ServerURL + mcpPath
		notif := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "notifications/cancelled",
		}
		body, _ := json.Marshal(notif)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			if h.sessionID != "" {
				req.Header.Set("Mcp-Session-Id", h.sessionID)
			}
			for k, v := range cfg.Headers {
				req.Header.Set(k, v)
			}
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				resp.Body.Close()
			}
		}
	}

	h.status = StatusStopped
	h.iLog.Info(fmt.Sprintf("MCPClientHandler %s stopped", h.name))
	return nil
}

// Status returns the current handler status.
func (h *MCPClientHandler) Status() ExecutorStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// Name returns the handler name.
func (h *MCPClientHandler) Name() string { return h.name }

// Type returns the protocol type.
func (h *MCPClientHandler) Type() ProtocolType { return ProtocolMCP }

// CallTool invokes a named tool on the MCP server with the given JSON arguments string.
// It uses JSON-RPC 2.0 over HTTP and returns the concatenated text content from the response.
func (h *MCPClientHandler) CallTool(ctx context.Context, toolName, argsJSON string) (string, error) {
	h.mu.RLock()
	cfg := h.mcpConfig
	sessionID := h.sessionID
	h.mu.RUnlock()

	if cfg == nil || cfg.ServerURL == "" {
		return "", fmt.Errorf("MCPClientHandler %s: no server configured", h.name)
	}

	mcpPath := cfg.MCPPath
	if mcpPath == "" {
		mcpPath = "/mcp"
	}
	endpoint := cfg.ServerURL + mcpPath

	var arguments interface{}
	if argsJSON != "" && argsJSON != "null" {
		if err := json.Unmarshal([]byte(argsJSON), &arguments); err != nil {
			arguments = map[string]interface{}{}
		}
	} else {
		arguments = map[string]interface{}{}
	}

	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}
	body, _ := json.Marshal(reqBody)

	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(callCtx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("MCPClientHandler build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", sessionID)
	}
	for k, v := range cfg.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("MCPClientHandler CallTool request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("MCPClientHandler CallTool HTTP %d: %s", resp.StatusCode, string(b))
	}

	var rpcResp struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return "", fmt.Errorf("MCPClientHandler decode response: %w", err)
	}
	if rpcResp.Error != nil {
		return "", fmt.Errorf("MCP error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	if rpcResp.Result.IsError {
		if len(rpcResp.Result.Content) > 0 {
			return "", fmt.Errorf("MCP tool error: %s", rpcResp.Result.Content[0].Text)
		}
		return "", fmt.Errorf("MCP tool returned error")
	}

	out := ""
	for _, c := range rpcResp.Result.Content {
		if c.Type == "text" {
			out += c.Text
		}
	}
	return out, nil
}
