package datahub

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// NewTransformEngine creates a new transform engine with built-in functions
func NewTransformEngine() *TransformEngine {
	te := &TransformEngine{
		builtInFunctions: make(map[string]TransformFunction),
		customScripts:    make(map[string]string),
	}

	// Register built-in transformation functions
	te.registerBuiltInFunctions()

	return te
}

// registerBuiltInFunctions registers all built-in transformation functions
func (te *TransformEngine) registerBuiltInFunctions() {
	// String transformations
	te.builtInFunctions["to_upper"] = toUpper
	te.builtInFunctions["to_lower"] = toLower
	te.builtInFunctions["trim"] = trim
	te.builtInFunctions["substring"] = substring
	te.builtInFunctions["concat"] = concat
	te.builtInFunctions["replace"] = replace

	// Number transformations
	te.builtInFunctions["round_to_2_decimals"] = roundTo2Decimals
	te.builtInFunctions["round_to_n_decimals"] = roundToNDecimals
	te.builtInFunctions["to_int"] = toInt
	te.builtInFunctions["to_float"] = toFloat

	// Date/Time transformations
	te.builtInFunctions["iso8601_to_soap_datetime"] = iso8601ToSOAPDateTime
	te.builtInFunctions["unix_timestamp_to_iso8601"] = unixTimestampToISO8601
	te.builtInFunctions["current_timestamp_iso8601"] = currentTimestampISO8601
	te.builtInFunctions["format_date"] = formatDate

	// Binary transformations
	te.builtInFunctions["bytes_to_hex"] = bytesToHex
	te.builtInFunctions["bytes_to_float32"] = bytesToFloat32
	te.builtInFunctions["bytes_to_string"] = bytesToString

	// Protocol-specific transformations
	te.builtInFunctions["mqtt_topic_to_channel"] = mqttTopicToChannel
	te.builtInFunctions["soap_envelope_wrap"] = soapEnvelopeWrap
	te.builtInFunctions["rest_to_graphql_query"] = restToGraphQLQuery

	// Array/Object transformations
	te.builtInFunctions["array_join"] = arrayJoin
	te.builtInFunctions["array_filter"] = arrayFilter
	te.builtInFunctions["object_merge"] = objectMerge
}

// Transform transforms a message envelope using a mapping definition
func (te *TransformEngine) Transform(source *MessageEnvelope, mapping *MappingDefinition) (*MessageEnvelope, error) {
	if !mapping.Active {
		return nil, fmt.Errorf("mapping %s is not active", mapping.ID)
	}

	dhLogger.Infof("Transforming message %s using mapping %s", source.ID, mapping.Name)

	// Create target envelope
	target := &MessageEnvelope{
		ID:            source.ID,
		Protocol:      mapping.TargetProtocol,
		Source:        source.Source,
		Destination:   source.Destination,
		Timestamp:     time.Now(),
		ContentType:   te.determineTargetContentType(mapping.TargetProtocol),
		Headers:       make(map[string]interface{}),
		Metadata:      make(map[string]interface{}),
		TransformPath: append(source.TransformPath, mapping.ID),
	}

	// Convert source body to JSON for processing
	sourceJSON, err := te.toJSON(source.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to convert source body to JSON: %w", err)
	}

	// Initialize target body
	targetBody := make(map[string]interface{})

	// Apply field mappings
	for _, fieldMapping := range mapping.Mappings {
		if err := te.applyFieldMapping(sourceJSON, targetBody, &fieldMapping); err != nil {
			if fieldMapping.Required {
				return nil, fmt.Errorf("failed to apply required field mapping %s -> %s: %w",
					fieldMapping.SourcePath, fieldMapping.TargetPath, err)
			}
			dhLogger.Warnf("Failed to apply optional field mapping %s -> %s: %v",
				fieldMapping.SourcePath, fieldMapping.TargetPath, err)
		}
	}

	// Apply transformations (filters, enrichments, etc.)
	if err := te.applyTransformations(targetBody, mapping.Transformations); err != nil {
		return nil, fmt.Errorf("failed to apply transformations: %w", err)
	}

	target.Body = targetBody

	// Convert body to appropriate format based on target protocol
	if err := te.convertBodyFormat(target); err != nil {
		return nil, fmt.Errorf("failed to convert body format: %w", err)
	}

	dhLogger.Infof("Successfully transformed message %s", source.ID)
	return target, nil
}

// applyFieldMapping applies a single field mapping
func (te *TransformEngine) applyFieldMapping(sourceJSON string, targetBody map[string]interface{}, mapping *FieldMapping) error {
	var value interface{}

	// Get source value
	if mapping.SourcePath == "" {
		// Use default value if no source path
		value = mapping.DefaultValue
	} else {
		result := gjson.Get(sourceJSON, mapping.SourcePath)
		if !result.Exists() {
			if mapping.Required && mapping.DefaultValue == nil {
				return fmt.Errorf("source path %s not found", mapping.SourcePath)
			}
			value = mapping.DefaultValue
		} else {
			value = result.Value()
		}
	}

	// Apply transformation function if specified
	if mapping.TransformFunc != "" {
		transformFunc, exists := te.builtInFunctions[mapping.TransformFunc]
		if !exists {
			return fmt.Errorf("transform function %s not found", mapping.TransformFunc)
		}

		var err error
		value, err = transformFunc(value, nil)
		if err != nil {
			return fmt.Errorf("transform function %s failed: %w", mapping.TransformFunc, err)
		}
	}

	// Convert to target data type
	value, err := te.convertDataType(value, mapping.DataType)
	if err != nil {
		return fmt.Errorf("failed to convert data type: %w", err)
	}

	// Set target value using sjson
	targetJSON, err := json.Marshal(targetBody)
	if err != nil {
		return fmt.Errorf("failed to marshal target body: %w", err)
	}

	targetPath := mapping.TargetPath
	// Convert XPath-like syntax to JSONPath if needed
	if strings.HasPrefix(targetPath, "//") {
		targetPath = te.xpathToJSONPath(targetPath)
	}

	updatedJSON, err := sjson.Set(string(targetJSON), targetPath, value)
	if err != nil {
		return fmt.Errorf("failed to set target path %s: %w", targetPath, err)
	}

	// Update target body
	if err := json.Unmarshal([]byte(updatedJSON), &targetBody); err != nil {
		return fmt.Errorf("failed to unmarshal updated target body: %w", err)
	}

	return nil
}

// applyTransformations applies transformation rules
func (te *TransformEngine) applyTransformations(targetBody map[string]interface{}, transformations []TransformRule) error {
	// Sort by order
	sortedTransforms := make([]TransformRule, len(transformations))
	copy(sortedTransforms, transformations)

	for i := 0; i < len(sortedTransforms)-1; i++ {
		for j := i + 1; j < len(sortedTransforms); j++ {
			if sortedTransforms[j].Order < sortedTransforms[i].Order {
				sortedTransforms[i], sortedTransforms[j] = sortedTransforms[j], sortedTransforms[i]
			}
		}
	}

	// Apply each transformation
	for _, transform := range sortedTransforms {
		switch transform.Type {
		case "enrich":
			if err := te.applyEnrichment(targetBody, transform.Config); err != nil {
				return fmt.Errorf("failed to apply enrichment: %w", err)
			}
		case "filter":
			if err := te.applyFilter(targetBody, transform.Config); err != nil {
				return fmt.Errorf("failed to apply filter: %w", err)
			}
		// Add more transformation types as needed
		default:
			dhLogger.Warnf("Unknown transformation type: %s", transform.Type)
		}
	}

	return nil
}

// applyEnrichment applies enrichment transformation
func (te *TransformEngine) applyEnrichment(targetBody map[string]interface{}, config map[string]interface{}) error {
	// Simple implementation - add fields from config
	if fields, ok := config["fields"].(map[string]interface{}); ok {
		for key, value := range fields {
			targetBody[key] = value
		}
	}

	if targetPath, ok := config["target_path"].(string); ok {
		if value, ok := config["value"]; ok {
			// Handle template values like {{current_timestamp}}
			if strValue, ok := value.(string); ok {
				if strings.HasPrefix(strValue, "{{") && strings.HasSuffix(strValue, "}}") {
					funcName := strings.TrimSpace(strValue[2 : len(strValue)-2])
					if funcName == "current_timestamp" {
						value = time.Now().Format(time.RFC3339)
					}
				}
			}
			targetBody[targetPath] = value
		}
	}

	return nil
}

// applyFilter applies filter transformation
func (te *TransformEngine) applyFilter(targetBody map[string]interface{}, config map[string]interface{}) error {
	// Simple implementation - can be enhanced
	return nil
}

// convertDataType converts a value to the specified data type
func (te *TransformEngine) convertDataType(value interface{}, dataType string) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch dataType {
	case "string":
		return fmt.Sprintf("%v", value), nil
	case "int":
		switch v := value.(type) {
		case int:
			return v, nil
		case int64:
			return int(v), nil
		case float64:
			return int(v), nil
		case string:
			return strconv.Atoi(v)
		default:
			return nil, fmt.Errorf("cannot convert %T to int", value)
		}
	case "float":
		switch v := value.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case string:
			return strconv.ParseFloat(v, 64)
		default:
			return nil, fmt.Errorf("cannot convert %T to float", value)
		}
	case "bool":
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			return strconv.ParseBool(v)
		default:
			return nil, fmt.Errorf("cannot convert %T to bool", value)
		}
	case "date":
		// Return as ISO8601 string
		switch v := value.(type) {
		case string:
			return v, nil
		case time.Time:
			return v.Format(time.RFC3339), nil
		default:
			return fmt.Sprintf("%v", value), nil
		}
	case "array", "object":
		return value, nil
	default:
		return value, nil
	}
}

// toJSON converts any value to JSON string
func (te *TransformEngine) toJSON(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
}

// xpathToJSONPath converts simple XPath-like syntax to JSONPath
func (te *TransformEngine) xpathToJSONPath(xpath string) string {
	// Simple conversion: //OrderRequest/Customer/EmailAddress -> OrderRequest.Customer.EmailAddress
	path := strings.TrimPrefix(xpath, "//")
	path = strings.ReplaceAll(path, "/", ".")
	return path
}

// determineTargetContentType determines content type based on protocol
func (te *TransformEngine) determineTargetContentType(protocol string) string {
	switch protocol {
	case "SOAP":
		return "application/soap+xml"
	case "REST", "GraphQL":
		return "application/json"
	case "TCP":
		return "application/octet-stream"
	default:
		return "application/json"
	}
}

// convertBodyFormat converts the body to the appropriate format for the target protocol
func (te *TransformEngine) convertBodyFormat(envelope *MessageEnvelope) error {
	switch envelope.Protocol {
	case "SOAP":
		// Wrap in SOAP envelope if needed
		// This is a simplified version
		return nil
	case "REST", "GraphQL", "Kafka", "MQTT":
		// Ensure JSON format
		if _, ok := envelope.Body.(map[string]interface{}); !ok {
			data, err := json.Marshal(envelope.Body)
			if err != nil {
				return err
			}
			var jsonBody map[string]interface{}
			if err := json.Unmarshal(data, &jsonBody); err != nil {
				return err
			}
			envelope.Body = jsonBody
		}
	}
	return nil
}

// NewMessageHistory creates a new message history
func NewMessageHistory(maxEntries int) *MessageHistory {
	return &MessageHistory{
		maxEntries: maxEntries,
		entries:    make([]MessageHistoryEntry, 0, maxEntries),
	}
}

// Add adds a history entry
func (mh *MessageHistory) Add(entry MessageHistoryEntry) {
	mh.entries = append(mh.entries, entry)
	if len(mh.entries) > mh.maxEntries {
		mh.entries = mh.entries[1:]
	}
}

// GetRecent returns recent history entries
func (mh *MessageHistory) GetRecent(limit int) []MessageHistoryEntry {
	if limit <= 0 || limit > len(mh.entries) {
		limit = len(mh.entries)
	}
	return mh.entries[len(mh.entries)-limit:]
}

// Built-in transformation functions

func toUpper(input interface{}, params map[string]interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", input)
	return strings.ToUpper(str), nil
}

func toLower(input interface{}, params map[string]interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", input)
	return strings.ToLower(str), nil
}

func trim(input interface{}, params map[string]interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", input)
	return strings.TrimSpace(str), nil
}

func substring(input interface{}, params map[string]interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", input)
	start := params["start"].(int)
	end := params["end"].(int)
	if start < 0 || end > len(str) || start > end {
		return "", fmt.Errorf("invalid substring parameters")
	}
	return str[start:end], nil
}

func concat(input interface{}, params map[string]interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", input)
	suffix := fmt.Sprintf("%v", params["suffix"])
	return str + suffix, nil
}

func replace(input interface{}, params map[string]interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", input)
	old := fmt.Sprintf("%v", params["old"])
	new := fmt.Sprintf("%v", params["new"])
	return strings.ReplaceAll(str, old, new), nil
}

func roundTo2Decimals(input interface{}, params map[string]interface{}) (interface{}, error) {
	var val float64
	switch v := input.(type) {
	case float64:
		val = v
	case int:
		val = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		val = parsed
	default:
		return nil, fmt.Errorf("cannot convert %T to float", input)
	}
	return math.Round(val*100) / 100, nil
}

func roundToNDecimals(input interface{}, params map[string]interface{}) (interface{}, error) {
	decimals := params["decimals"].(int)
	multiplier := math.Pow(10, float64(decimals))

	var val float64
	switch v := input.(type) {
	case float64:
		val = v
	case int:
		val = float64(v)
	default:
		return nil, fmt.Errorf("cannot convert %T to float", input)
	}
	return math.Round(val*multiplier) / multiplier, nil
}

func toInt(input interface{}, params map[string]interface{}) (interface{}, error) {
	switch v := input.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return nil, fmt.Errorf("cannot convert %T to int", input)
	}
}

func toFloat(input interface{}, params map[string]interface{}) (interface{}, error) {
	switch v := input.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return nil, fmt.Errorf("cannot convert %T to float", input)
	}
}

func iso8601ToSOAPDateTime(input interface{}, params map[string]interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", input)
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return nil, err
	}
	// SOAP datetime format: 2006-01-02T15:04:05
	return t.Format("2006-01-02T15:04:05"), nil
}

func unixTimestampToISO8601(input interface{}, params map[string]interface{}) (interface{}, error) {
	var timestamp int64
	switch v := input.(type) {
	case int64:
		timestamp = v
	case int:
		timestamp = int64(v)
	case float64:
		timestamp = int64(v)
	default:
		return nil, fmt.Errorf("cannot convert %T to timestamp", input)
	}
	t := time.Unix(timestamp, 0)
	return t.Format(time.RFC3339), nil
}

func currentTimestampISO8601(input interface{}, params map[string]interface{}) (interface{}, error) {
	return time.Now().Format(time.RFC3339), nil
}

func formatDate(input interface{}, params map[string]interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", input)
	format := params["format"].(string)
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return nil, err
	}
	return t.Format(format), nil
}

func bytesToHex(input interface{}, params map[string]interface{}) (interface{}, error) {
	bytes, ok := input.([]byte)
	if !ok {
		return nil, fmt.Errorf("input is not []byte")
	}
	return fmt.Sprintf("%x", bytes), nil
}

func bytesToFloat32(input interface{}, params map[string]interface{}) (interface{}, error) {
	// Simple implementation - would need proper binary parsing
	return 0.0, nil
}

func bytesToString(input interface{}, params map[string]interface{}) (interface{}, error) {
	bytes, ok := input.([]byte)
	if !ok {
		return nil, fmt.Errorf("input is not []byte")
	}
	return string(bytes), nil
}

func mqttTopicToChannel(input interface{}, params map[string]interface{}) (interface{}, error) {
	topic := fmt.Sprintf("%v", input)
	// Convert MQTT topic format to channel format
	// e.g., "devices/sensor1/temperature" -> "devices.sensor1.temperature"
	return strings.ReplaceAll(topic, "/", "."), nil
}

func soapEnvelopeWrap(input interface{}, params map[string]interface{}) (interface{}, error) {
	// Wrap content in SOAP envelope
	return input, nil
}

func restToGraphQLQuery(input interface{}, params map[string]interface{}) (interface{}, error) {
	// Convert REST parameters to GraphQL query
	return input, nil
}

func arrayJoin(input interface{}, params map[string]interface{}) (interface{}, error) {
	arr, ok := input.([]interface{})
	if !ok {
		return nil, fmt.Errorf("input is not an array")
	}
	separator := params["separator"].(string)
	strArr := make([]string, len(arr))
	for i, v := range arr {
		strArr[i] = fmt.Sprintf("%v", v)
	}
	return strings.Join(strArr, separator), nil
}

func arrayFilter(input interface{}, params map[string]interface{}) (interface{}, error) {
	// Simple array filtering - can be enhanced
	return input, nil
}

func objectMerge(input interface{}, params map[string]interface{}) (interface{}, error) {
	// Merge objects - can be enhanced
	return input, nil
}
