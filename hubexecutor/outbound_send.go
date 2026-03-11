package hubexecutor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/mdaxf/iac/models"
)

// OutboundSendResult is the result of a protocol-aware outbound send.
type OutboundSendResult struct {
	ResponseBody string
	StatusCode   int
	Success      bool
}

// SendToEndpoint sends payload to a specific hub endpoint using the protocol configured for that group.
// It loads the hub configuration by ID, resolves the protocol group and endpoint, then dispatches
// to the appropriate protocol implementation (HTTP/REST, MQTT, Kafka, MCP, etc.).
func (m *HubExecutorManager) SendToEndpoint(
	ctx context.Context,
	hubID string,
	protocolGroupID string,
	endpointID string,
	payload []byte,
	contentType string,
	method string,
) (*OutboundSendResult, error) {
	if m.hubLoader == nil {
		return nil, fmt.Errorf("hubexecutor: hub loader not configured")
	}

	hub, err := m.hubLoader.LoadByID(hubID)
	if err != nil {
		return nil, fmt.Errorf("hubexecutor: failed to load hub %s: %v", hubID, err)
	}
	if hub == nil {
		return nil, fmt.Errorf("hubexecutor: hub %s not found", hubID)
	}
	if hub.Outbound == nil {
		return nil, fmt.Errorf("hubexecutor: hub %s has no outbound configuration", hubID)
	}

	// Find protocol group
	var pg *models.HubSimpleProtocolGroup
	for i := range hub.Outbound.ProtocolGroups {
		if hub.Outbound.ProtocolGroups[i].ID == protocolGroupID {
			pg = &hub.Outbound.ProtocolGroups[i]
			break
		}
	}
	if pg == nil {
		return nil, fmt.Errorf("hubexecutor: protocol group %s not found in hub %s", protocolGroupID, hubID)
	}

	// Find endpoint
	var ep *models.HubSimpleEndpoint
	for i := range pg.Endpoints {
		if pg.Endpoints[i].ID == endpointID {
			ep = &pg.Endpoints[i]
			break
		}
	}
	if ep == nil {
		return nil, fmt.Errorf("hubexecutor: endpoint %s not found in protocol group %s", endpointID, protocolGroupID)
	}

	// Apply defaults
	if method == "" {
		method = ep.Method
		if method == "" {
			method = "POST"
		}
	}
	if contentType == "" {
		contentType = "application/json"
	}

	// Dispatch by protocol
	switch normalizeProtocol(pg.Protocol) {
	case "HTTP":
		return m.sendHTTP(ctx, pg, ep, payload, contentType, method)
	case "MQTT":
		return m.sendMQTT(ctx, pg, ep, payload)
	case "Kafka":
		return m.sendKafka(ctx, pg, ep, payload)
	case "MCP":
		return m.sendMCP(ctx, pg, ep, payload)
	default:
		// Graceful fallback: attempt HTTP if the endpoint has a URL
		if ep.EndpointURL != "" || (ep.Path != "" && strings.HasPrefix(ep.Path, "http")) {
			m.iLog.Warn(fmt.Sprintf("SendToEndpoint: protocol '%s' has no dedicated sender; falling back to HTTP", pg.Protocol))
			return m.sendHTTP(ctx, pg, ep, payload, contentType, method)
		}
		return nil, fmt.Errorf("hubexecutor: protocol '%s' is not supported for outbound sends", pg.Protocol)
	}
}

// normalizeProtocol maps a protocol string to a simple category key
func normalizeProtocol(protocol string) string {
	switch strings.ToUpper(strings.TrimSpace(protocol)) {
	case "REST API", "REST", "HTTP", "HTTPS", "SOAP", "SOAP API", "GRAPHQL":
		return "HTTP"
	case "MQTT CLIENT", "MQTT":
		return "MQTT"
	case "KAFKA":
		return "Kafka"
	case "ACTIVEMQ", "AMQP":
		return "AMQP"
	case "MCP":
		return "MCP"
	default:
		return protocol
	}
}

// sendHTTP sends payload via HTTP/REST to the configured endpoint URL.
func (m *HubExecutorManager) sendHTTP(
	ctx context.Context,
	pg *models.HubSimpleProtocolGroup,
	ep *models.HubSimpleEndpoint,
	payload []byte,
	contentType string,
	method string,
) (*OutboundSendResult, error) {
	url := strings.TrimRight(ep.EndpointURL, "/")
	if ep.Path != "" {
		path := strings.TrimLeft(ep.Path, "/")
		if url != "" {
			url = url + "/" + path
		} else {
			url = path
		}
	}
	if url == "" {
		return nil, fmt.Errorf("sendHTTP: endpoint '%s' has no URL configured", ep.Name)
	}

	timeout := time.Duration(60) * time.Second
	if ep.Timeout > 0 {
		timeout = time.Duration(ep.Timeout) * time.Second
	} else if pg.Timeout > 0 {
		timeout = time.Duration(pg.Timeout) * time.Second
	}

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("sendHTTP: failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)

	// Apply protocol group auth (endpoint-level overrides below)
	applyAuth(req, pg.AuthType, pg.AuthConfig)
	// Endpoint-level auth overrides group auth
	if ep.AuthType != "" {
		applyAuth(req, ep.AuthType, ep.AuthConfig)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sendHTTP: request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return &OutboundSendResult{
		ResponseBody: string(body),
		StatusCode:   resp.StatusCode,
		Success:      resp.StatusCode < 400,
	}, nil
}

// applyAuth sets Authorization or API-Key headers based on auth type configuration
func applyAuth(req *http.Request, authType string, authConfig map[string]interface{}) {
	if authType == "" || authConfig == nil {
		return
	}
	switch strings.ToLower(authType) {
	case "bearer":
		if token, ok := authConfig["token"].(string); ok && token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case "basic":
		username, _ := authConfig["username"].(string)
		password, _ := authConfig["password"].(string)
		req.SetBasicAuth(username, password)
	case "api-key", "apikey":
		headerName, _ := authConfig["header"].(string)
		key, _ := authConfig["key"].(string)
		if headerName == "" {
			headerName = "X-API-Key"
		}
		if key != "" {
			req.Header.Set(headerName, key)
		}
	}
}

// sendMQTT creates a transient MQTT client, publishes the payload to the topic defined in the endpoint path, then disconnects.
func (m *HubExecutorManager) sendMQTT(
	ctx context.Context,
	pg *models.HubSimpleProtocolGroup,
	ep *models.HubSimpleEndpoint,
	payload []byte,
) (*OutboundSendResult, error) {
	if pg.BrokerConfig == nil {
		return nil, fmt.Errorf("sendMQTT: broker_config is required for MQTT protocol group '%s'", pg.Name)
	}
	bc := pg.BrokerConfig

	brokerURL := bc.BrokerURL
	if brokerURL == "" {
		scheme := "tcp"
		if bc.UseTLS {
			scheme = "ssl"
		}
		brokerURL = fmt.Sprintf("%s://%s:%d", scheme, bc.BrokerHost, bc.BrokerPort)
	}

	// Topic: endpoint path takes priority, fall back to first broker topic
	topic := strings.TrimLeft(ep.Path, "/")
	if topic == "" && len(bc.Topics) > 0 {
		topic = bc.Topics[0].Topic
	}
	if topic == "" {
		return nil, fmt.Errorf("sendMQTT: no topic configured for endpoint '%s'", ep.Name)
	}

	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(fmt.Sprintf("iac-outbound-%s", uuid.New().String()[:8])).
		SetUsername(bc.Username).
		SetPassword(bc.Password).
		SetAutoReconnect(false).
		SetConnectTimeout(10 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.WaitTimeout(10*time.Second) && token.Error() != nil {
		return nil, fmt.Errorf("sendMQTT: connect to '%s' failed: %v", brokerURL, token.Error())
	}
	defer client.Disconnect(250)

	qos := byte(bc.QoS)
	token := client.Publish(topic, qos, false, payload)
	if token.WaitTimeout(15*time.Second) && token.Error() != nil {
		return nil, fmt.Errorf("sendMQTT: publish to topic '%s' failed: %v", topic, token.Error())
	}

	return &OutboundSendResult{
		ResponseBody: fmt.Sprintf(`{"published":true,"topic":"%s","broker":"%s"}`, topic, brokerURL),
		StatusCode:   200,
		Success:      true,
	}, nil
}

// sendKafka creates a transient Sarama sync producer, sends the payload to the topic defined in the endpoint path, then closes.
func (m *HubExecutorManager) sendKafka(
	ctx context.Context,
	pg *models.HubSimpleProtocolGroup,
	ep *models.HubSimpleEndpoint,
	payload []byte,
) (*OutboundSendResult, error) {
	if pg.BrokerConfig == nil {
		return nil, fmt.Errorf("sendKafka: broker_config is required for Kafka protocol group '%s'", pg.Name)
	}
	bc := pg.BrokerConfig

	servers := []string{fmt.Sprintf("%s:%d", bc.BrokerHost, bc.BrokerPort)}
	if bc.BrokerURL != "" {
		servers = []string{bc.BrokerURL}
	}

	// Topic: endpoint path takes priority, fall back to first broker topic
	topic := strings.TrimLeft(ep.Path, "/")
	if topic == "" && len(bc.Topics) > 0 {
		topic = bc.Topics[0].Topic
	}
	if topic == "" {
		return nil, fmt.Errorf("sendKafka: no topic configured for endpoint '%s'", ep.Name)
	}

	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Return.Successes = true
	cfg.Producer.Retry.Max = 3

	producer, err := sarama.NewSyncProducer(servers, cfg)
	if err != nil {
		return nil, fmt.Errorf("sendKafka: failed to create producer for %v: %v", servers, err)
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(payload),
	}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return nil, fmt.Errorf("sendKafka: failed to send to topic '%s': %v", topic, err)
	}

	return &OutboundSendResult{
		ResponseBody: fmt.Sprintf(`{"published":true,"topic":"%s","partition":%d,"offset":%d}`, topic, partition, offset),
		StatusCode:   200,
		Success:      true,
	}, nil
}

// sendMCP calls a named tool on an MCP server using JSON-RPC 2.0.
func (m *HubExecutorManager) sendMCP(
	ctx context.Context,
	pg *models.HubSimpleProtocolGroup,
	ep *models.HubSimpleEndpoint,
	payload []byte,
) (*OutboundSendResult, error) {
	if pg.MCPConfig == nil {
		return nil, fmt.Errorf("sendMCP: mcp_config is required for MCP protocol group '%s'", pg.Name)
	}
	mc := pg.MCPConfig

	toolName := mc.ToolName
	if toolName == "" {
		return nil, fmt.Errorf("sendMCP: tool_name is required in mcp_config for protocol group '%s'", pg.Name)
	}
	mcpPath := mc.MCPPath
	if mcpPath == "" {
		mcpPath = "/mcp"
	}
	timeoutSec := mc.TimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = 30
	}

	// Build arguments from payload (try JSON object, fall back to string payload)
	var arguments map[string]interface{}
	if err := json.Unmarshal(payload, &arguments); err != nil {
		arguments = map[string]interface{}{"payload": string(payload)}
	}

	rpcReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      time.Now().UnixNano(),
		"method":  "tools/call",
		"params":  map[string]interface{}{"name": toolName, "arguments": arguments},
	}
	reqBody, _ := json.Marshal(rpcReq)

	callCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	endpoint := strings.TrimRight(mc.ServerURL, "/") + mcpPath
	httpReq, err := http.NewRequestWithContext(callCtx, http.MethodPost, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("sendMCP: failed to build request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	if mc.Headers != nil {
		for k, v := range mc.Headers {
			httpReq.Header.Set(k, v)
		}
	}

	httpClient := &http.Client{Timeout: time.Duration(timeoutSec+5) * time.Second}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sendMCP: HTTP call failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return &OutboundSendResult{
			StatusCode:   resp.StatusCode,
			ResponseBody: string(body),
			Success:      false,
		}, fmt.Errorf("sendMCP: HTTP %d from MCP server", resp.StatusCode)
	}

	// Parse JSON-RPC 2.0 response
	var rpcResp struct {
		Result *struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result,omitempty"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		// Return raw body if not parseable as JSON-RPC
		return &OutboundSendResult{StatusCode: 200, ResponseBody: string(body), Success: true}, nil
	}
	if rpcResp.Error != nil {
		return &OutboundSendResult{
			StatusCode:   500,
			ResponseBody: rpcResp.Error.Message,
			Success:      false,
		}, fmt.Errorf("sendMCP: RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	out := ""
	if rpcResp.Result != nil {
		for _, c := range rpcResp.Result.Content {
			if c.Type == "text" {
				out += c.Text
			}
		}
		if rpcResp.Result.IsError {
			return &OutboundSendResult{StatusCode: 500, ResponseBody: out, Success: false},
				fmt.Errorf("sendMCP: tool returned error: %s", out)
		}
	}

	return &OutboundSendResult{StatusCode: 200, ResponseBody: out, Success: true}, nil
}
