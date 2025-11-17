package datahub

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdaxf/iac/documents"
	"github.com/sirupsen/logrus"
)

// JobHandler handles DataHub transformation and routing jobs
type JobHandler struct {
	hub       *DataHub
	db        *sql.DB
	docDB     *documents.DocDB
	logger    *logrus.Logger
	enabled   bool
}

// NewJobHandler creates a new DataHub job handler
func NewJobHandler(hub *DataHub, db *sql.DB, docDB *documents.DocDB, logger *logrus.Logger) *JobHandler {
	if logger == nil {
		logger = logrus.New()
	}

	return &JobHandler{
		hub:     hub,
		db:      db,
		docDB:   docDB,
		logger:  logger,
		enabled: true,
	}
}

// ExecuteTransformJob executes a transformation job
// This is called by the job worker to process messages
func (jh *JobHandler) ExecuteTransformJob(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error) {
	if !jh.enabled {
		return nil, fmt.Errorf("job handler is disabled")
	}

	jh.logger.Info("Executing DataHub transform job")

	// Extract job parameters
	protocol, _ := payload["protocol"].(string)
	source, _ := payload["source"].(string)
	destination, _ := payload["destination"].(string)
	contentType, _ := payload["content_type"].(string)
	body := payload["body"]
	mappingID, _ := payload["mapping_id"].(string)
	routingRuleID, _ := payload["routing_rule_id"].(string)

	if protocol == "" {
		protocol = "UNKNOWN"
	}
	if contentType == "" {
		contentType = "application/json"
	}

	// Create message envelope
	envelope := CreateEnvelope(protocol, source, destination, contentType, body)

	// Add metadata from payload
	if metadata, ok := payload["metadata"].(map[string]interface{}); ok {
		for k, v := range metadata {
			envelope.Metadata[k] = v
		}
	}

	envelope.Metadata["job_id"] = ctx.Value("job_id")
	envelope.Metadata["routing_rule_id"] = routingRuleID
	envelope.Metadata["mapping_id"] = mappingID

	// Apply transformation if mapping ID is specified
	if mappingID != "" {
		jh.logger.Infof("Applying transformation mapping: %s", mappingID)

		mapping, err := jh.hub.GetMapping(mappingID)
		if err != nil {
			return nil, fmt.Errorf("failed to get mapping: %w", err)
		}

		transformedEnvelope, err := jh.hub.transformEngine.Transform(envelope, mapping)
		if err != nil {
			return nil, fmt.Errorf("transformation failed: %w", err)
		}

		envelope = transformedEnvelope
	}

	// Route message through DataHub
	if routingRuleID != "" || destination != "" {
		jh.logger.Info("Routing message through DataHub")

		if err := jh.hub.RouteMessage(envelope); err != nil {
			return nil, fmt.Errorf("routing failed: %w", err)
		}
	}

	// Prepare result
	result := map[string]interface{}{
		"status":         "success",
		"message_id":     envelope.ID,
		"protocol":       envelope.Protocol,
		"destination":    envelope.Destination,
		"transform_path": envelope.TransformPath,
		"timestamp":      envelope.Timestamp,
	}

	// Include transformed body in result
	bodyJSON, err := json.Marshal(envelope.Body)
	if err != nil {
		result["body"] = fmt.Sprintf("%v", envelope.Body)
	} else {
		result["body"] = string(bodyJSON)
	}

	jh.logger.Info("DataHub transform job completed successfully")
	return result, nil
}

// ExecuteReceiveJob executes a message receive job
// This polls a protocol adapter for new messages
func (jh *JobHandler) ExecuteReceiveJob(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error) {
	if !jh.enabled {
		return nil, fmt.Errorf("job handler is disabled")
	}

	jh.logger.Info("Executing DataHub receive job")

	// Extract parameters
	protocol, _ := payload["protocol"].(string)
	timeout := 30 * time.Second
	if timeoutSec, ok := payload["timeout"].(float64); ok {
		timeout = time.Duration(timeoutSec) * time.Second
	}

	autoRoute := true
	if ar, ok := payload["auto_route"].(bool); ok {
		autoRoute = ar
	}

	if protocol == "" {
		return nil, fmt.Errorf("protocol is required")
	}

	// Get protocol adapter
	adapter, err := jh.hub.GetAdapter(protocol)
	if err != nil {
		return nil, fmt.Errorf("failed to get adapter: %w", err)
	}

	// Receive message
	jh.logger.Infof("Receiving message from %s protocol", protocol)
	envelope, err := adapter.Receive(timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}

	if envelope == nil {
		return map[string]interface{}{
			"status":  "no_message",
			"message": "No messages available",
		}, nil
	}

	envelope.Metadata["received_at"] = time.Now()
	envelope.Metadata["job_id"] = ctx.Value("job_id")

	// Auto-route through DataHub if enabled
	if autoRoute {
		jh.logger.Info("Auto-routing received message")

		if err := jh.hub.RouteMessage(envelope); err != nil {
			jh.logger.Warnf("Auto-routing failed: %v", err)
			// Don't fail the job, message was received successfully
		}
	}

	// Prepare result
	result := map[string]interface{}{
		"status":      "success",
		"message_id":  envelope.ID,
		"protocol":    envelope.Protocol,
		"source":      envelope.Source,
		"timestamp":   envelope.Timestamp,
		"auto_routed": autoRoute,
	}

	bodyJSON, err := json.Marshal(envelope.Body)
	if err != nil {
		result["body"] = fmt.Sprintf("%v", envelope.Body)
	} else {
		result["body"] = string(bodyJSON)
	}

	jh.logger.Info("DataHub receive job completed successfully")
	return result, nil
}

// ExecuteSendJob executes a message send job
// This sends a message through a protocol adapter
func (jh *JobHandler) ExecuteSendJob(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error) {
	if !jh.enabled {
		return nil, fmt.Errorf("job handler is disabled")
	}

	jh.logger.Info("Executing DataHub send job")

	// Extract parameters
	protocol, _ := payload["protocol"].(string)
	destination, _ := payload["destination"].(string)
	contentType, _ := payload["content_type"].(string)
	body := payload["body"]
	mappingID, _ := payload["mapping_id"].(string)

	if protocol == "" {
		return nil, fmt.Errorf("protocol is required")
	}
	if destination == "" {
		return nil, fmt.Errorf("destination is required")
	}

	if contentType == "" {
		contentType = "application/json"
	}

	// Create message envelope
	envelope := CreateEnvelope(protocol, "", destination, contentType, body)

	// Add metadata
	if metadata, ok := payload["metadata"].(map[string]interface{}); ok {
		for k, v := range metadata {
			envelope.Metadata[k] = v
		}
	}

	envelope.Metadata["job_id"] = ctx.Value("job_id")

	// Apply transformation if mapping ID is specified
	if mappingID != "" {
		jh.logger.Infof("Applying transformation mapping: %s", mappingID)

		mapping, err := jh.hub.GetMapping(mappingID)
		if err != nil {
			return nil, fmt.Errorf("failed to get mapping: %w", err)
		}

		transformedEnvelope, err := jh.hub.transformEngine.Transform(envelope, mapping)
		if err != nil {
			return nil, fmt.Errorf("transformation failed: %w", err)
		}

		envelope = transformedEnvelope
	}

	// Get protocol adapter
	adapter, err := jh.hub.GetAdapter(protocol)
	if err != nil {
		return nil, fmt.Errorf("failed to get adapter: %w", err)
	}

	// Send message
	jh.logger.Infof("Sending message via %s protocol to %s", protocol, destination)
	if err := adapter.Send(envelope); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Prepare result
	result := map[string]interface{}{
		"status":         "success",
		"message_id":     envelope.ID,
		"protocol":       envelope.Protocol,
		"destination":    envelope.Destination,
		"transform_path": envelope.TransformPath,
		"timestamp":      envelope.Timestamp,
	}

	jh.logger.Info("DataHub send job completed successfully")
	return result, nil
}

// Enable enables the job handler
func (jh *JobHandler) Enable() {
	jh.enabled = true
	jh.logger.Info("DataHub job handler enabled")
}

// Disable disables the job handler
func (jh *JobHandler) Disable() {
	jh.enabled = false
	jh.logger.Info("DataHub job handler disabled")
}

// IsEnabled returns whether the job handler is enabled
func (jh *JobHandler) IsEnabled() bool {
	return jh.enabled
}

// GetStats returns job handler statistics
func (jh *JobHandler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":      jh.enabled,
		"hub_enabled":  jh.hub.IsEnabled(),
		"adapters":     len(jh.hub.adapters),
		"mappings":     len(jh.hub.mappings),
		"routing_rules": len(jh.hub.routingRules),
	}
}
