package datahub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	globalDataHub     *DataHub
	globalDataHubOnce sync.Once
	dhLogger          *logrus.Logger
)

// NewDataHub creates a new DataHub instance
func NewDataHub(logger *logrus.Logger) *DataHub {
	if logger == nil {
		logger = logrus.New()
	}
	dhLogger = logger

	return &DataHub{
		adapters:        make(map[string]ProtocolAdapter),
		mappings:        make(map[string]*MappingDefinition),
		transformEngine: NewTransformEngine(),
		routingRules:    make([]*RoutingRule, 0),
		messageHistory:  NewMessageHistory(1000),
		enabled:         true,
	}
}

// GetGlobalDataHub returns the global DataHub instance (singleton)
func GetGlobalDataHub() *DataHub {
	globalDataHubOnce.Do(func() {
		globalDataHub = NewDataHub(nil)
	})
	return globalDataHub
}

// RegisterAdapter registers a protocol adapter
func (dh *DataHub) RegisterAdapter(name string, adapter ProtocolAdapter) error {
	if _, exists := dh.adapters[name]; exists {
		return fmt.Errorf("adapter %s already registered", name)
	}
	dh.adapters[name] = adapter
	dhLogger.Infof("Registered protocol adapter: %s", name)
	return nil
}

// GetAdapter retrieves a protocol adapter by name
func (dh *DataHub) GetAdapter(name string) (ProtocolAdapter, error) {
	adapter, exists := dh.adapters[name]
	if !exists {
		return nil, fmt.Errorf("adapter %s not found", name)
	}
	return adapter, nil
}

// LoadMappingsFromFile loads mapping definitions from a JSON file
func (dh *DataHub) LoadMappingsFromFile(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read mapping file: %w", err)
	}

	var config struct {
		Mappings     []MappingDefinition `json:"mappings"`
		RoutingRules []RoutingRule       `json:"routing_rules"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse mapping file: %w", err)
	}

	// Load mappings
	for i := range config.Mappings {
		mapping := &config.Mappings[i]
		dh.mappings[mapping.ID] = mapping
		dhLogger.Infof("Loaded mapping: %s (%s -> %s)", mapping.Name, mapping.SourceProtocol, mapping.TargetProtocol)
	}

	// Load routing rules
	for i := range config.RoutingRules {
		rule := &config.RoutingRules[i]
		dh.routingRules = append(dh.routingRules, rule)
		dhLogger.Infof("Loaded routing rule: %s (%s -> %s)", rule.Name, rule.Source, rule.Destination)
	}

	return nil
}

// AddMapping adds a mapping definition
func (dh *DataHub) AddMapping(mapping *MappingDefinition) {
	dh.mappings[mapping.ID] = mapping
	dhLogger.Infof("Added mapping: %s", mapping.ID)
}

// GetMapping retrieves a mapping definition by ID
func (dh *DataHub) GetMapping(id string) (*MappingDefinition, error) {
	mapping, exists := dh.mappings[id]
	if !exists {
		return nil, fmt.Errorf("mapping %s not found", id)
	}
	return mapping, nil
}

// AddRoutingRule adds a routing rule
func (dh *DataHub) AddRoutingRule(rule *RoutingRule) {
	dh.routingRules = append(dh.routingRules, rule)
	dhLogger.Infof("Added routing rule: %s", rule.ID)
}

// RouteMessage routes a message based on routing rules
func (dh *DataHub) RouteMessage(envelope *MessageEnvelope) error {
	if !dh.enabled {
		return fmt.Errorf("datahub is disabled")
	}

	startTime := time.Now()

	// Find matching routing rules
	matchedRules := dh.findMatchingRules(envelope)
	if len(matchedRules) == 0 {
		dhLogger.Warnf("No routing rules matched for message %s from %s", envelope.ID, envelope.Source)
		return fmt.Errorf("no routing rules matched")
	}

	// Process each matched rule
	for _, rule := range matchedRules {
		if err := dh.processRoutingRule(envelope, rule, startTime); err != nil {
			dhLogger.Errorf("Failed to process routing rule %s: %v", rule.ID, err)
			return err
		}
	}

	return nil
}

// findMatchingRules finds routing rules that match the message
func (dh *DataHub) findMatchingRules(envelope *MessageEnvelope) []*RoutingRule {
	matched := make([]*RoutingRule, 0)

	for _, rule := range dh.routingRules {
		if !rule.Active {
			continue
		}

		// Check if source matches
		if !dh.sourceMatches(envelope, rule.Source) {
			continue
		}

		// Check conditions
		if len(rule.Conditions) > 0 {
			if !dh.evaluateConditions(envelope, rule.Conditions) {
				continue
			}
		}

		matched = append(matched, rule)
	}

	// Sort by priority (higher priority first)
	for i := 0; i < len(matched)-1; i++ {
		for j := i + 1; j < len(matched); j++ {
			if matched[j].Priority > matched[i].Priority {
				matched[i], matched[j] = matched[j], matched[i]
			}
		}
	}

	return matched
}

// processRoutingRule processes a single routing rule
func (dh *DataHub) processRoutingRule(envelope *MessageEnvelope, rule *RoutingRule, startTime time.Time) error {
	dhLogger.Infof("Processing routing rule %s for message %s", rule.Name, envelope.ID)

	// Apply transformation if mapping is specified
	var transformedEnvelope *MessageEnvelope
	var err error

	if rule.MappingID != "" {
		mapping, err := dh.GetMapping(rule.MappingID)
		if err != nil {
			return fmt.Errorf("failed to get mapping %s: %w", rule.MappingID, err)
		}

		transformedEnvelope, err = dh.transformEngine.Transform(envelope, mapping)
		if err != nil {
			dh.recordHistory(envelope, rule, false, err, time.Since(startTime))
			return fmt.Errorf("failed to transform message: %w", err)
		}
	} else {
		transformedEnvelope = envelope
	}

	// Send to destination
	destAdapter, err := dh.getDestinationAdapter(rule.Destination)
	if err != nil {
		dh.recordHistory(envelope, rule, false, err, time.Since(startTime))
		return fmt.Errorf("failed to get destination adapter: %w", err)
	}

	if err := destAdapter.Send(transformedEnvelope); err != nil {
		dh.recordHistory(envelope, rule, false, err, time.Since(startTime))
		return fmt.Errorf("failed to send message: %w", err)
	}

	dh.recordHistory(envelope, rule, true, nil, time.Since(startTime))
	dhLogger.Infof("Successfully routed message %s via rule %s", envelope.ID, rule.Name)

	return nil
}

// sourceMatches checks if the envelope source matches the rule source pattern
func (dh *DataHub) sourceMatches(envelope *MessageEnvelope, ruleSource string) bool {
	// Simple implementation - can be enhanced with regex/glob patterns
	return envelope.Source == ruleSource || envelope.Protocol+":"+envelope.Source == ruleSource
}

// getDestinationAdapter extracts the protocol adapter from destination string
func (dh *DataHub) getDestinationAdapter(destination string) (ProtocolAdapter, error) {
	// Parse destination format: "PROTOCOL:endpoint"
	// e.g., "REST:http://api.example.com", "SOAP:http://soap.example.com"

	// Simple parsing - can be enhanced
	for protocol, adapter := range dh.adapters {
		if len(destination) > len(protocol) && destination[:len(protocol)] == protocol {
			return adapter, nil
		}
	}

	return nil, fmt.Errorf("no adapter found for destination: %s", destination)
}

// evaluateConditions evaluates routing conditions
func (dh *DataHub) evaluateConditions(envelope *MessageEnvelope, conditions []MappingCondition) bool {
	// Simple implementation - always return true for now
	// TODO: Implement full condition evaluation with JSONPath
	return true
}

// recordHistory records a message transformation event
func (dh *DataHub) recordHistory(envelope *MessageEnvelope, rule *RoutingRule, success bool, err error, duration time.Duration) {
	entry := MessageHistoryEntry{
		MessageID:   envelope.ID,
		Timestamp:   time.Now(),
		SourceProto: envelope.Protocol,
		TargetProto: rule.Destination,
		MappingID:   rule.MappingID,
		Success:     success,
		Duration:    duration,
		Metadata: map[string]interface{}{
			"rule_id":   rule.ID,
			"rule_name": rule.Name,
		},
	}

	if err != nil {
		entry.Error = err.Error()
	}

	dh.messageHistory.Add(entry)
}

// CreateEnvelope creates a new message envelope
func CreateEnvelope(protocol, source, destination string, contentType string, body interface{}) *MessageEnvelope {
	return &MessageEnvelope{
		ID:          uuid.New().String(),
		Protocol:    protocol,
		Source:      source,
		Destination: destination,
		Timestamp:   time.Now(),
		ContentType: contentType,
		Headers:     make(map[string]interface{}),
		Body:        body,
		Metadata:    make(map[string]interface{}),
		TransformPath: make([]string, 0),
	}
}

// Enable enables the DataHub
func (dh *DataHub) Enable() {
	dh.enabled = true
	dhLogger.Info("DataHub enabled")
}

// Disable disables the DataHub
func (dh *DataHub) Disable() {
	dh.enabled = false
	dhLogger.Info("DataHub disabled")
}

// IsEnabled returns whether the DataHub is enabled
func (dh *DataHub) IsEnabled() bool {
	return dh.enabled
}

// GetMessageHistory returns message history entries
func (dh *DataHub) GetMessageHistory(limit int) []MessageHistoryEntry {
	return dh.messageHistory.GetRecent(limit)
}

// Close closes all adapters and cleans up resources
func (dh *DataHub) Close() error {
	dhLogger.Info("Closing DataHub and all adapters")

	for name, adapter := range dh.adapters {
		if err := adapter.Close(); err != nil {
			dhLogger.Errorf("Failed to close adapter %s: %v", name, err)
		}
	}

	return nil
}
