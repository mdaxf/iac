package workflow

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	wftype "github.com/mdaxf/iac/workflow/types"
)

// ExecuteOutbound executes an Integration Hub outbound call for a workflow outbound node.
// When the node is configured with a hub endpoint (HubID + ProtocolGroupID + EndpointID), it
// dispatches via the protocol-aware SendToHubFunc (HTTP, MQTT, Kafka, MCP, …).
// Otherwise it falls back to a direct HTTP call using the resolved OutboundURL.
func ExecuteOutbound(workflowtaskid int64, NodeData wftype.Node, idbTx *sql.Tx, DocDBCon *documents.DocDB, UserName string) error {
	iLog := logger.Log{ModuleName: logger.Framework, ControllerName: "workflow outbound execution", User: UserName}
	startTime := time.Now()
	defer func() {
		iLog.PerformanceWithDuration("ExecuteOutbound", time.Since(startTime))
	}()

	defer func() {
		if r := recover(); r != nil {
			iLog.Error(fmt.Sprintf("Recovered in ExecuteOutbound: %v", r))
		}
	}()

	iLog.Debug(fmt.Sprintf("ExecuteOutbound for workflowtaskid: %d", workflowtaskid))

	if NodeData.OutboundConfig == nil {
		err := fmt.Errorf("outbound configuration is missing for outbound node")
		iLog.Error(err.Error())
		return err
	}

	cfg := NodeData.OutboundConfig

	internaltransaction := false
	var err error

	if idbTx == nil {
		idbTx, err = dbconn.DB.Begin()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error creating DB connection: %v", err))
			return err
		}
		internaltransaction = true
		defer idbTx.Rollback()
	}

	if DocDBCon == nil {
		DocDBCon = documents.DocDBCon
	}

	dbop := dbconn.NewDBOperation(UserName, idbTx, logger.Framework)

	// Load current ProcessData
	rows, err := dbop.Query_Json(fmt.Sprintf(
		"select workflowentityid, workflownodeid, processdata from workflow_tasks where id = %d", workflowtaskid))
	if err != nil {
		iLog.Error(fmt.Sprintf("Error loading workflow task: %v", err))
		return err
	}
	if len(rows) == 0 {
		return fmt.Errorf("workflow task %d not found", workflowtaskid)
	}

	processData := map[string]interface{}{}
	if rows[0]["processdata"] != nil {
		if err = json.Unmarshal([]byte(rows[0]["processdata"].(string)), &processData); err != nil {
			iLog.Error(fmt.Sprintf("Error parsing process data: %v", err))
			return err
		}
	}

	// Merge node-level processdata
	for k, v := range NodeData.ProcessData {
		processData[k] = v
	}

	// Build request payload
	var bodyBytes []byte
	if cfg.PayloadDataKey != "" {
		if raw, ok := processData[cfg.PayloadDataKey]; ok {
			switch v := raw.(type) {
			case string:
				bodyBytes = []byte(v)
			case []byte:
				bodyBytes = v
			default:
				bodyBytes, err = json.Marshal(v)
				if err != nil {
					iLog.Error(fmt.Sprintf("Error marshaling payload from ProcessData[%s]: %v", cfg.PayloadDataKey, err))
					return err
				}
			}
		}
	} else {
		bodyBytes, err = json.Marshal(processData)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error marshaling ProcessData as payload: %v", err))
			return err
		}
	}

	contentType := cfg.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	method := strings.ToUpper(cfg.Method)
	if method == "" {
		method = "POST"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var respBodyStr string
	var statusCode int

	// --- Protocol-aware path: use Integration Hub dispatcher when hub config is present ---
	if cfg.HubID != "" && cfg.ProtocolGroupID != "" && cfg.EndpointID != "" && SendToHubFunc != nil {
		respBody, code, sendErr := SendToHubFunc(ctx, cfg.HubID, cfg.ProtocolGroupID, cfg.EndpointID, bodyBytes, contentType, method)
		if sendErr != nil {
			iLog.Error(fmt.Sprintf("ExecuteOutbound: hub send failed (hub=%s, group=%s, endpoint=%s): %v",
				cfg.HubID, cfg.ProtocolGroupID, cfg.EndpointID, sendErr))
			return sendErr
		}
		respBodyStr = respBody
		statusCode = code
		iLog.Info(fmt.Sprintf("ExecuteOutbound: hub send → status %d (hub=%s endpoint=%s)", statusCode, cfg.HubName, cfg.EndpointName))
	} else {
		// --- Fallback path: direct HTTP call using the resolved URL ---
		outboundURL := cfg.OutboundURL
		if cfg.DynamicUrl {
			if dynURL, ok := processData["outbound_url"].(string); ok && dynURL != "" {
				outboundURL = dynURL
			}
		}
		if outboundURL == "" {
			err = fmt.Errorf("outbound URL is not configured for node %s and no hub endpoint is selected", NodeData.Name)
			iLog.Error(err.Error())
			return err
		}

		req, err := http.NewRequestWithContext(ctx, method, outboundURL, bytes.NewReader(bodyBytes))
		if err != nil {
			iLog.Error(fmt.Sprintf("Error building outbound request: %v", err))
			return err
		}
		req.Header.Set("Content-Type", contentType)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			iLog.Error(fmt.Sprintf("Outbound request failed: %v", err))
			return err
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		respBodyStr = string(respBody)
		statusCode = resp.StatusCode
		iLog.Info(fmt.Sprintf("ExecuteOutbound: %s %s → %d", method, outboundURL, statusCode))
	}

	// Store result in ProcessData
	responseKey := cfg.ResponseDataKey
	if responseKey == "" {
		responseKey = "outbound_response"
	}
	processData[responseKey] = respBodyStr
	processData["outbound_status_code"] = statusCode
	processData["outbound_success"] = statusCode < 400
	processData["outbound_endpoint"] = cfg.EndpointName
	processData["outbound_timestamp"] = time.Now().UTC().Format("2006-01-02 15:04:05")

	// Try to parse response as JSON and merge individual fields into processData
	var respJSON map[string]interface{}
	if json.Unmarshal([]byte(respBodyStr), &respJSON) == nil {
		for k, v := range respJSON {
			processData[fmt.Sprintf("outbound_%s", k)] = v
		}
	}

	// Persist updated ProcessData
	jsonData, err := json.Marshal(processData)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error marshaling updated ProcessData: %v", err))
		return err
	}

	idCol := dbop.QuoteIdentifier("id")
	_, err = dbop.TableUpdate("workflow_tasks",
		[]string{"processdata"},
		[]string{string(jsonData)},
		[]int{0},
		fmt.Sprintf("%s = %d", idCol, workflowtaskid))
	if err != nil {
		iLog.Error(fmt.Sprintf("Error updating workflow task ProcessData: %v", err))
		return err
	}

	if internaltransaction {
		idbTx.Commit()
	}

	iLog.Info(fmt.Sprintf("ExecuteOutbound completed for workflowtaskid: %d", workflowtaskid))
	return nil
}
