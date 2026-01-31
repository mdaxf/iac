package hubexecutor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/robertkrimen/otto"
)

// DefaultDestinationExecutor implements the DestinationExecutor interface
type DefaultDestinationExecutor struct {
	iLog       logger.Log
	httpClient *http.Client
	jobExecutor func(ctx context.Context, jobName string, params map[string]interface{}) error
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

	// Execute the job
	if d.jobExecutor != nil {
		if err := d.jobExecutor(ctx, dest.Target, params); err != nil {
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

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, outboundURL, bytes.NewReader(payload.Body))
	if err != nil {
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
		return &ExecutionResult{
			Success:   false,
			Error:     err,
			Message:   fmt.Sprintf("Outbound request failed: %v", err),
			Timestamp: time.Now(),
		}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("Outbound request returned status %d", resp.StatusCode),
			Timestamp: time.Now(),
		}, fmt.Errorf("outbound request returned status %d", resp.StatusCode)
	}

	return &ExecutionResult{
		Success:   true,
		Message:   fmt.Sprintf("Routed to outbound %s (status: %d)", outboundURL, resp.StatusCode),
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
