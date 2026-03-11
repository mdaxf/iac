package hubexecutor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/robertkrimen/otto"
)

// DefaultDestinationExecutor implements the DestinationExecutor interface
type DefaultDestinationExecutor struct {
	iLog                 logger.Log
	httpClient           *http.Client
	jobExecutor          func(ctx context.Context, jobName string, params map[string]interface{}) error
	agentGatewayExecutor func(ctx context.Context, gatewayID, input string) (string, error)
	historyRecorder      HistoryRecorder
	hubID                string
	hubName              string
}

// NewDestinationExecutor creates a new destination executor
func NewDestinationExecutor() *DefaultDestinationExecutor {
	return &DefaultDestinationExecutor{
		iLog: logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DestinationExecutor"},
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetJobExecutor sets the job executor function
func (d *DefaultDestinationExecutor) SetJobExecutor(executor func(ctx context.Context, jobName string, params map[string]interface{}) error) {
	d.jobExecutor = executor
}

// SetAgentGatewayExecutor sets the agent gateway execution callback.
// Call this from contollershandlers.go after AgentGatewayService is initialized.
func (d *DefaultDestinationExecutor) SetAgentGatewayExecutor(fn func(ctx context.Context, gatewayID, input string) (string, error)) {
	d.agentGatewayExecutor = fn
}

// SetHistoryRecorder sets the history recorder for tracking outbound operations
func (d *DefaultDestinationExecutor) SetHistoryRecorder(recorder HistoryRecorder) {
	d.historyRecorder = recorder
}

// SetHubInfo sets the hub information for history recording
func (d *DefaultDestinationExecutor) SetHubInfo(hubID, hubName string) {
	d.hubID = hubID
	d.hubName = hubName
}

// Execute executes a destination with the given payload
func (d *DefaultDestinationExecutor) Execute(ctx context.Context, dest Destination, payload *MessagePayload) (*ExecutionResult, error) {
	startTime := time.Now()
	d.iLog.Debug(fmt.Sprintf("Executing destination: %s (type: %s)", dest.ID, dest.Type))

	var result *ExecutionResult
	var err error

	switch dest.Type {
	case "Job":
		result, err = d.executeJob(ctx, dest, payload)
	case "Execute Transcode":
		result, err = d.executeTranscode(ctx, dest, payload)
	case "Route to Outbound":
		result, err = d.executeRouteToOutbound(ctx, dest, payload)
	case "Custom Script":
		result, err = d.executeCustomScript(ctx, dest, payload)
	case "MCP Tool":
		result, err = d.executeMCPTool(ctx, dest, payload)
	case "Agent Gateway":
		result, err = d.executeAgentGateway(ctx, dest, payload)
	default:
		err = fmt.Errorf("unknown destination type: %s", dest.Type)
		result = &ExecutionResult{
			Success:   false,
			Error:     err,
			Message:   err.Error(),
			Timestamp: time.Now(),
		}
	}

	if result != nil {
		result.Duration = time.Since(startTime)
	}

	return result, err
}

// executeJob executes a job destination
func (d *DefaultDestinationExecutor) executeJob(ctx context.Context, dest Destination, payload *MessagePayload) (*ExecutionResult, error) {
	d.iLog.Debug(fmt.Sprintf("Executing job: %s", dest.Target))

	// Build job parameters from payload
	params := make(map[string]interface{})

	// Add payload data to parameters
	if payload != nil {
		params["endpoint_id"] = payload.EndpointID
		params["protocol"] = string(payload.Protocol)
		params["path"] = payload.Path
		params["method"] = payload.Method
		params["headers"] = payload.Headers
		params["content_type"] = payload.ContentType
		params["timestamp"] = payload.Timestamp

		// Try to parse body as JSON
		if len(payload.Body) > 0 {
			var bodyData interface{}
			if err := json.Unmarshal(payload.Body, &bodyData); err == nil {
				params["body"] = bodyData
			} else {
				params["body"] = string(payload.Body)
			}
		}

		// Add metadata
		for k, v := range payload.Metadata {
			params[k] = v
		}
	}

	// Add destination config parameters
	for k, v := range dest.Config {
		params[k] = v
	}

	// Record history for job execution (as outbound action)
	var transactionID string
	if d.historyRecorder != nil {
		direction := "outbound"
		if payload != nil && payload.Direction != "" {
			direction = payload.Direction
		}
		history := &models.IntHubHistory{
			HubID:        d.hubID,
			HubName:      d.hubName,
			Direction:    direction,
			Protocol:     "Job",
			EndpointID:   dest.ID,
			EndpointName: dest.Target,
			Path:         dest.Target,
			Payload:      string(payload.Body),
			PayloadSize:  len(payload.Body),
			StartTime:    time.Now(),
			Status:       models.IntHubHistoryStatusProcessing,
			Action:       "Execute Job",
			Metadata:     payload.Metadata,
		}
		var err error
		transactionID, err = d.historyRecorder.RecordTransaction(ctx, history)
		if err != nil {
			d.iLog.Error(fmt.Sprintf("Failed to record job history: %v", err))
		}
	}

	// Execute the job
	if d.jobExecutor != nil {
		if err := d.jobExecutor(ctx, dest.Target, params); err != nil {
			if transactionID != "" && d.historyRecorder != nil {
				d.historyRecorder.FailTransaction(ctx, transactionID, fmt.Sprintf("Job execution failed: %v", err))
			}
			return &ExecutionResult{
				Success:   false,
				Error:     err,
				Message:   fmt.Sprintf("Job execution failed: %v", err),
				Timestamp: time.Now(),
			}, err
		}
	} else {
		d.iLog.Warn(fmt.Sprintf("Job executor not configured, skipping job: %s", dest.Target))
	}

	// Complete history transaction
	if transactionID != "" && d.historyRecorder != nil {
		d.historyRecorder.CompleteTransaction(ctx, transactionID, "Job executed successfully", "", 0)
	}

	return &ExecutionResult{
		Success:   true,
		Message:   fmt.Sprintf("Job %s executed successfully", dest.Target),
		Timestamp: time.Now(),
	}, nil
}

// executeTranscode executes a transcode destination
func (d *DefaultDestinationExecutor) executeTranscode(ctx context.Context, dest Destination, payload *MessagePayload) (*ExecutionResult, error) {
	d.iLog.Debug(fmt.Sprintf("Executing transcode: %s", dest.Target))

	// Get mapping configuration
	mappingID := dest.MappingID
	if mappingID == "" {
		d.iLog.Warn("No mapping ID specified for transcode, passing data through")
		return &ExecutionResult{
			Success:   true,
			Message:   "No transcode mapping specified",
			Data:      payload.Body,
			Timestamp: time.Now(),
		}, nil
	}

	// TODO: Implement actual transcode logic using mapping definitions
	// For now, pass through the data
	return &ExecutionResult{
		Success:   true,
		Message:   fmt.Sprintf("Transcode %s executed (mapping: %s)", dest.Target, mappingID),
		Data:      payload.Body,
		Timestamp: time.Now(),
	}, nil
}

// executeRouteToOutbound routes the message to an outbound endpoint
func (d *DefaultDestinationExecutor) executeRouteToOutbound(ctx context.Context, dest Destination, payload *MessagePayload) (*ExecutionResult, error) {
	d.iLog.Debug(fmt.Sprintf("Routing to outbound: %s", dest.Target))

	// Get outbound endpoint URL from config
	outboundURL, _ := dest.Config["outbound_url"].(string)
	if outboundURL == "" {
		outboundURL = dest.Target
	}

	if outboundURL == "" {
		return &ExecutionResult{
			Success:   false,
			Error:     fmt.Errorf("no outbound URL specified"),
			Message:   "No outbound URL specified",
			Timestamp: time.Now(),
		}, fmt.Errorf("no outbound URL specified")
	}

	// Get HTTP method
	method, _ := dest.Config["method"].(string)
	if method == "" {
		method = "POST"
	}

	// Record history for outbound operation
	var transactionID string
	if d.historyRecorder != nil {
		history := &models.IntHubHistory{
			HubID:        d.hubID,
			HubName:      d.hubName,
			Direction:    "outbound",
			Protocol:     string(payload.Protocol),
			EndpointID:   dest.ID,
			EndpointName: dest.Target,
			Path:         outboundURL,
			Method:       method,
			Payload:      string(payload.Body),
			PayloadSize:  len(payload.Body),
			StartTime:    time.Now(),
			Status:       models.IntHubHistoryStatusProcessing,
			Action:       "Route to Outbound",
			Metadata:     payload.Metadata,
		}
		var err error
		transactionID, err = d.historyRecorder.RecordTransaction(ctx, history)
		if err != nil {
			d.iLog.Error(fmt.Sprintf("Failed to record outbound history: %v", err))
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, outboundURL, bytes.NewReader(payload.Body))
	if err != nil {
		if transactionID != "" && d.historyRecorder != nil {
			d.historyRecorder.FailTransaction(ctx, transactionID, fmt.Sprintf("Failed to create request: %v", err))
		}
		return &ExecutionResult{
			Success:   false,
			Error:     err,
			Message:   fmt.Sprintf("Failed to create request: %v", err),
			Timestamp: time.Now(),
		}, err
	}

	// Set content type
	if payload.ContentType != "" {
		req.Header.Set("Content-Type", payload.ContentType)
	} else {
		req.Header.Set("Content-Type", "application/json")
	}

	// Copy headers from config
	if headers, ok := dest.Config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if strVal, ok := v.(string); ok {
				req.Header.Set(k, strVal)
			}
		}
	}

	// Execute request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		if transactionID != "" && d.historyRecorder != nil {
			d.historyRecorder.FailTransaction(ctx, transactionID, fmt.Sprintf("Outbound request failed: %v", err))
		}
		return &ExecutionResult{
			Success:   false,
			Error:     err,
			Message:   fmt.Sprintf("Outbound request failed: %v", err),
			Timestamp: time.Now(),
		}, err
	}
	defer resp.Body.Close()

	// Read response body for history
	respBody, _ := io.ReadAll(resp.Body)
	respBodyStr := string(respBody)

	if resp.StatusCode >= 400 {
		if transactionID != "" && d.historyRecorder != nil {
			d.historyRecorder.FailTransaction(ctx, transactionID, fmt.Sprintf("Outbound request returned status %d: %s", resp.StatusCode, respBodyStr))
		}
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("Outbound request returned status %d", resp.StatusCode),
			Timestamp: time.Now(),
		}, fmt.Errorf("outbound request returned status %d", resp.StatusCode)
	}

	// Complete history transaction
	if transactionID != "" && d.historyRecorder != nil {
		d.historyRecorder.CompleteTransaction(ctx, transactionID, respBodyStr, "", resp.StatusCode)
	}

	return &ExecutionResult{
		Success:   true,
		Message:   fmt.Sprintf("Routed to outbound %s (status: %d)", outboundURL, resp.StatusCode),
		Data:      respBodyStr,
		Timestamp: time.Now(),
	}, nil
}

// executeCustomScript executes a custom JavaScript script
func (d *DefaultDestinationExecutor) executeCustomScript(ctx context.Context, dest Destination, payload *MessagePayload) (*ExecutionResult, error) {
	d.iLog.Debug(fmt.Sprintf("Executing custom script: %s", dest.Target))

	// Get script from config
	script, _ := dest.Config["script"].(string)
	if script == "" {
		return &ExecutionResult{
			Success:   false,
			Error:     fmt.Errorf("no script specified"),
			Message:   "No script specified",
			Timestamp: time.Now(),
		}, fmt.Errorf("no script specified")
	}

	// Create JavaScript VM
	vm := otto.New()

	// Set payload data in VM
	if payload != nil {
		payloadMap := map[string]interface{}{
			"endpoint_id":  payload.EndpointID,
			"protocol":     string(payload.Protocol),
			"path":         payload.Path,
			"method":       payload.Method,
			"headers":      payload.Headers,
			"content_type": payload.ContentType,
			"timestamp":    payload.Timestamp.Format(time.RFC3339),
		}

		// Try to parse body as JSON
		if len(payload.Body) > 0 {
			var bodyData interface{}
			if err := json.Unmarshal(payload.Body, &bodyData); err == nil {
				payloadMap["body"] = bodyData
			} else {
				payloadMap["body"] = string(payload.Body)
			}
		}

		vm.Set("payload", payloadMap)
		vm.Set("metadata", payload.Metadata)
	}

	// Set destination config
	vm.Set("config", dest.Config)

	// Set console.log function
	vm.Set("console", map[string]interface{}{
		"log": func(call otto.FunctionCall) otto.Value {
			args := make([]interface{}, len(call.ArgumentList))
			for i, arg := range call.ArgumentList {
				args[i], _ = arg.Export()
			}
			d.iLog.Debug(fmt.Sprintf("Script console.log: %v", args))
			return otto.UndefinedValue()
		},
	})

	// Execute script with timeout
	done := make(chan struct{})
	var scriptResult interface{}
	var scriptErr error

	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				scriptErr = fmt.Errorf("script panic: %v", r)
			}
		}()

		result, err := vm.Run(script)
		if err != nil {
			scriptErr = err
			return
		}

		scriptResult, _ = result.Export()
	}()

	// Wait for script with timeout
	select {
	case <-done:
		if scriptErr != nil {
			return &ExecutionResult{
				Success:   false,
				Error:     scriptErr,
				Message:   fmt.Sprintf("Script execution failed: %v", scriptErr),
				Timestamp: time.Now(),
			}, scriptErr
		}
	case <-ctx.Done():
		return &ExecutionResult{
			Success:   false,
			Error:     ctx.Err(),
			Message:   "Script execution timed out",
			Timestamp: time.Now(),
		}, ctx.Err()
	case <-time.After(30 * time.Second):
		return &ExecutionResult{
			Success:   false,
			Error:     fmt.Errorf("script execution timed out"),
			Message:   "Script execution timed out",
			Timestamp: time.Now(),
		}, fmt.Errorf("script execution timed out")
	}

	return &ExecutionResult{
		Success:   true,
		Message:   "Custom script executed successfully",
		Data:      scriptResult,
		Timestamp: time.Now(),
	}, nil
}

// EvaluateCondition evaluates a JavaScript condition script
func EvaluateCondition(script string, newValue, oldValue interface{}, tagName string) (bool, error) {
	if script == "" || script == "return true;" {
		return true, nil
	}

	vm := otto.New()
	vm.Set("newValue", newValue)
	vm.Set("oldValue", oldValue)
	vm.Set("tagName", tagName)

	result, err := vm.Run(script)
	if err != nil {
		return false, fmt.Errorf("condition script error: %v", err)
	}

	boolResult, err := result.ToBoolean()
	if err != nil {
		return false, fmt.Errorf("condition script did not return boolean: %v", err)
	}

	return boolResult, nil
}

// executeMCPTool calls a named tool on an MCP server via JSON-RPC 2.0.
//
// Destination config keys:
//
//	server_url  – MCP server base URL (required)
//	mcp_path    – JSON-RPC endpoint path (default "/mcp")
//	tool_name   – MCP tool to invoke (required)
//	timeout_sec – per-call HTTP timeout (default 30)
func (d *DefaultDestinationExecutor) executeMCPTool(ctx context.Context, dest Destination, payload *MessagePayload) (*ExecutionResult, error) {
	startTime := time.Now()

	serverURL, _ := dest.Config["server_url"].(string)
	mcpPath, _ := dest.Config["mcp_path"].(string)
	toolName, _ := dest.Config["tool_name"].(string)

	if serverURL == "" {
		err := fmt.Errorf("MCP Tool destination '%s': server_url is required", dest.ID)
		return &ExecutionResult{Success: false, Error: err, Message: err.Error(), Timestamp: startTime}, err
	}
	if toolName == "" {
		err := fmt.Errorf("MCP Tool destination '%s': tool_name is required", dest.ID)
		return &ExecutionResult{Success: false, Error: err, Message: err.Error(), Timestamp: startTime}, err
	}
	if mcpPath == "" {
		mcpPath = "/mcp"
	}

	timeoutSec := 30
	if ts, ok := dest.Config["timeout_sec"].(float64); ok && ts > 0 {
		timeoutSec = int(ts)
	}

	// Build arguments from the message payload
	arguments := map[string]interface{}{}
	if payload != nil && len(payload.Body) > 0 {
		var bodyData interface{}
		if err := json.Unmarshal(payload.Body, &bodyData); err == nil {
			arguments["payload"] = bodyData
		} else {
			arguments["payload"] = string(payload.Body)
		}
		if payload.Path != "" {
			arguments["path"] = payload.Path
		}
		if payload.Method != "" {
			arguments["method"] = payload.Method
		}
	}
	// Allow additional arguments from destination config
	if extra, ok := dest.Config["arguments"].(map[string]interface{}); ok {
		for k, v := range extra {
			arguments[k] = v
		}
	}

	// Build JSON-RPC 2.0 request
	rpcReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      time.Now().UnixNano(),
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}
	reqBody, _ := json.Marshal(rpcReq)

	endpoint := serverURL + mcpPath
	callCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(callCtx, http.MethodPost, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return &ExecutionResult{Success: false, Error: err, Message: err.Error(), Timestamp: startTime}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	// Extra headers from destination config
	if headers, ok := dest.Config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if sv, ok := v.(string); ok {
				httpReq.Header.Set(k, sv)
			}
		}
	}

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return &ExecutionResult{Success: false, Error: err, Message: err.Error(), Timestamp: startTime, Duration: time.Since(startTime)}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("MCP Tool HTTP %d: %s", resp.StatusCode, string(respBody))
		return &ExecutionResult{Success: false, Error: err, Message: err.Error(), Timestamp: startTime, Duration: time.Since(startTime)}, err
	}

	// Parse JSON-RPC 2.0 response
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
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return &ExecutionResult{Success: false, Error: err, Message: err.Error(), Timestamp: startTime, Duration: time.Since(startTime)}, err
	}
	if rpcResp.Error != nil {
		err = fmt.Errorf("MCP error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
		return &ExecutionResult{Success: false, Error: err, Message: err.Error(), Timestamp: startTime, Duration: time.Since(startTime)}, err
	}
	if rpcResp.Result.IsError {
		msg := "MCP tool returned error"
		if len(rpcResp.Result.Content) > 0 {
			msg = rpcResp.Result.Content[0].Text
		}
		err = fmt.Errorf("%s", msg)
		return &ExecutionResult{Success: false, Error: err, Message: msg, Timestamp: startTime, Duration: time.Since(startTime)}, err
	}

	// Concatenate text content
	out := ""
	for _, c := range rpcResp.Result.Content {
		if c.Type == "text" {
			out += c.Text
		}
	}

	d.iLog.Info(fmt.Sprintf("MCP Tool '%s' executed successfully on %s", toolName, serverURL))
	return &ExecutionResult{
		Success:   true,
		Message:   out,
		Timestamp: startTime,
		Duration:  time.Since(startTime),
	}, nil
}

// executeAgentGateway runs an Agent Gateway synchronously via the injected callback.
//
// Destination config keys for "Agent Gateway" type:
//
//	gateway_id – UUID of the AgentGateway to trigger (required; falls back to dest.Target)
func (d *DefaultDestinationExecutor) executeAgentGateway(ctx context.Context, dest Destination, payload *MessagePayload) (*ExecutionResult, error) {
	startTime := time.Now()

	if d.agentGatewayExecutor == nil {
		err := fmt.Errorf("Agent Gateway executor not configured for destination '%s'", dest.ID)
		return &ExecutionResult{Success: false, Error: err, Message: err.Error(), Timestamp: startTime}, err
	}

	gatewayID, _ := dest.Config["gateway_id"].(string)
	if gatewayID == "" {
		gatewayID = dest.Target
	}
	if gatewayID == "" {
		err := fmt.Errorf("Agent Gateway destination '%s': gateway_id is required", dest.ID)
		return &ExecutionResult{Success: false, Error: err, Message: err.Error(), Timestamp: startTime}, err
	}

	// Build input from payload body
	input := ""
	if payload != nil && len(payload.Body) > 0 {
		input = string(payload.Body)
	}

	response, err := d.agentGatewayExecutor(ctx, gatewayID, input)
	if err != nil {
		d.iLog.Error(fmt.Sprintf("Agent Gateway '%s' execution failed: %v", gatewayID, err))
		return &ExecutionResult{
			Success:   false,
			Error:     err,
			Message:   fmt.Sprintf("Agent Gateway execution failed: %v", err),
			Timestamp: startTime,
			Duration:  time.Since(startTime),
		}, err
	}

	d.iLog.Info(fmt.Sprintf("Agent Gateway '%s' executed successfully", gatewayID))
	return &ExecutionResult{
		Success:   true,
		Message:   response,
		Data:      response,
		Timestamp: startTime,
		Duration:  time.Since(startTime),
	}, nil
}
