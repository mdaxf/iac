package funcs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OutboundFuncs executes an Integration Hub outbound call
type OutboundFuncs struct{}

// OutboundFuncConfig holds the configuration stored in Function.mapdata
type OutboundFuncConfig struct {
	// Hub endpoint selection (resolved at design time via cascading dropdowns)
	HubID             string `json:"hubId,omitempty"`
	HubName           string `json:"hubName,omitempty"`
	ProtocolGroupID   string `json:"protocolGroupId,omitempty"`
	ProtocolGroupName string `json:"protocolGroupName,omitempty"`
	EndpointID        string `json:"endpointId,omitempty"`
	EndpointName      string `json:"endpointName,omitempty"`
	// Resolved outbound URL (cached from hub endpoint config at design time)
	OutboundURL string `json:"outboundUrl,omitempty"`
	Method      string `json:"method,omitempty"`      // HTTP method, default POST
	ContentType string `json:"contentType,omitempty"` // Content-Type header, default application/json
	// Payload control
	// PayloadInputName: if set, use only this input's value as the request body.
	// If empty, all function inputs are serialised as a JSON object.
	PayloadInputName string `json:"payloadInputName,omitempty"`
	// DynamicUrl: if true, override OutboundURL with the value of the function
	// input named "url" (useful when the target endpoint is determined at runtime).
	DynamicUrl bool `json:"dynamicUrl,omitempty"`
}

// Execute carries out the outbound call via the Integration Hub protocol dispatcher when
// hub configuration is present, or falls back to a direct HTTP call for manual URL usage.
func (o *OutboundFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		f.iLog.PerformanceWithDuration("engine.funcs.OutboundFuncs.Execute", time.Since(startTime))
	}()

	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("Recovered from panic in Outbound execution: %v", r)
			f.iLog.Error(msg)
			f.ErrorMessage = msg
		}
	}()

	f.iLog.Info(fmt.Sprintf("Start Outbound Function: %s", f.Fobj.Name))

	// Extract config from mapdata
	var cfg OutboundFuncConfig
	if f.Fobj.Mapdata != nil {
		b, err := json.Marshal(f.Fobj.Mapdata)
		if err != nil {
			f.iLog.Error(fmt.Sprintf("OutboundFuncs: failed to marshal config: %v", err))
			f.ErrorMessage = err.Error()
			return
		}
		if err = json.Unmarshal(b, &cfg); err != nil {
			f.iLog.Error(fmt.Sprintf("OutboundFuncs: failed to unmarshal config: %v", err))
			f.ErrorMessage = err.Error()
			return
		}
	}

	// Collect function inputs
	namelist, valuelist, _ := f.SetInputs()
	inputMap := make(map[string]interface{}, len(namelist))
	for i, name := range namelist {
		inputMap[name] = valuelist[i]
	}

	// Build request payload
	var bodyBytes []byte
	if cfg.PayloadInputName != "" {
		if raw, ok := inputMap[cfg.PayloadInputName]; ok {
			switch v := raw.(type) {
			case string:
				bodyBytes = []byte(v)
			case []byte:
				bodyBytes = v
			default:
				b, err := json.Marshal(v)
				if err != nil {
					msg := fmt.Sprintf("OutboundFuncs: failed to marshal payload input '%s': %v", cfg.PayloadInputName, err)
					f.iLog.Error(msg)
					f.CancelExecution(msg)
					f.ErrorMessage = msg
					return
				}
				bodyBytes = b
			}
		}
	} else {
		b, err := json.Marshal(inputMap)
		if err != nil {
			msg := fmt.Sprintf("OutboundFuncs: failed to marshal inputs as payload: %v", err)
			f.iLog.Error(msg)
			f.CancelExecution(msg)
			f.ErrorMessage = msg
			return
		}
		bodyBytes = b
	}

	contentType := cfg.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	method := strings.ToUpper(cfg.Method)
	if method == "" {
		method = "POST"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// --- Protocol-aware path: use Integration Hub dispatcher when hub config is present ---
	if cfg.HubID != "" && cfg.ProtocolGroupID != "" && cfg.EndpointID != "" && SendToHubFunc != nil {
		respBody, statusCode, err := SendToHubFunc(ctx, cfg.HubID, cfg.ProtocolGroupID, cfg.EndpointID, bodyBytes, contentType, method)
		if err != nil {
			msg := fmt.Sprintf("OutboundFuncs: hub send failed (hub=%s, group=%s, endpoint=%s): %v",
				cfg.HubID, cfg.ProtocolGroupID, cfg.EndpointID, err)
			f.iLog.Error(msg)
			f.CancelExecution(msg)
			f.ErrorMessage = msg
			return
		}
		f.iLog.Info(fmt.Sprintf("OutboundFuncs: hub send → status %d (hub=%s endpoint=%s)", statusCode, cfg.HubName, cfg.EndpointName))
		f.SetOutputs(map[string]interface{}{
			"StatusCode":   statusCode,
			"ResponseBody": respBody,
			"Success":      statusCode < 400,
		})
		return
	}

	// --- Fallback path: direct HTTP call using the resolved URL ---
	outboundURL := cfg.OutboundURL
	if cfg.DynamicUrl {
		if urlVal, ok := inputMap["url"].(string); ok && urlVal != "" {
			outboundURL = urlVal
			delete(inputMap, "url")
			// Re-build body without the "url" key
			b, err := json.Marshal(inputMap)
			if err == nil {
				bodyBytes = b
			}
		}
	}
	if outboundURL == "" {
		msg := "OutboundFuncs: outbound URL is not configured and no hub endpoint is selected"
		f.iLog.Error(msg)
		f.CancelExecution(msg)
		f.ErrorMessage = msg
		return
	}

	req, err := http.NewRequestWithContext(ctx, method, outboundURL, bytes.NewReader(bodyBytes))
	if err != nil {
		msg := fmt.Sprintf("OutboundFuncs: failed to build request: %v", err)
		f.iLog.Error(msg)
		f.CancelExecution(msg)
		f.ErrorMessage = msg
		return
	}
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		msg := fmt.Sprintf("OutboundFuncs: request failed: %v", err)
		f.iLog.Error(msg)
		f.CancelExecution(msg)
		f.ErrorMessage = msg
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	f.iLog.Info(fmt.Sprintf("OutboundFuncs: %s %s → %d", method, outboundURL, resp.StatusCode))

	f.SetOutputs(map[string]interface{}{
		"StatusCode":   resp.StatusCode,
		"ResponseBody": string(respBody),
		"Success":      resp.StatusCode < 400,
	})
}
