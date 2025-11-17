package funcs

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mdaxf/iac/engine/types"
)

// TypeConverter provides common type conversion operations
type TypeConverter struct{}

// ConvertToInt safely converts a value to int with error handling
func (tc *TypeConverter) ConvertToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case int32:
		return int(v), nil
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	case string:
		if v == "" {
			return 0, nil
		}
		result, err := strconv.Atoi(v)
		if err != nil {
			// Try parsing as float first
			floatVal, err2 := strconv.ParseFloat(v, 64)
			if err2 != nil {
				return 0, fmt.Errorf("failed to convert '%s' to int: %w", v, err)
			}
			return int(floatVal), nil
		}
		return result, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("unsupported type for int conversion: %T", value)
	}
}

// ConvertToFloat safely converts a value to float64 with error handling
func (tc *TypeConverter) ConvertToFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case string:
		if v == "" {
			return 0.0, nil
		}
		result, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0.0, fmt.Errorf("failed to convert '%s' to float: %w", v, err)
		}
		return result, nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0.0, fmt.Errorf("unsupported type for float conversion: %T", value)
	}
}

// ConvertToBool safely converts a value to bool with error handling
func (tc *TypeConverter) ConvertToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int, int64, int32:
		return v != 0, nil
	case float64, float32:
		return v != 0, nil
	case string:
		lowerValue := strings.ToLower(strings.TrimSpace(v))
		switch lowerValue {
		case "true", "1", "yes", "y", "on":
			return true, nil
		case "false", "0", "no", "n", "off", "":
			return false, nil
		default:
			return false, fmt.Errorf("invalid boolean value: '%s'", v)
		}
	default:
		return false, fmt.Errorf("unsupported type for bool conversion: %T", value)
	}
}

// ConvertToDateTime safely converts a value to time.Time with error handling
func (tc *TypeConverter) ConvertToDateTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		if v == "" {
			return time.Time{}, nil
		}

		// Try standard format
		result, err := time.Parse(types.DateTimeFormat, v)
		if err == nil {
			return result, nil
		}

		// Try ISO 8601
		result, err = time.Parse(time.RFC3339, v)
		if err == nil {
			return result, nil
		}

		// Try date only
		result, err = time.Parse("2006-01-02", v)
		if err == nil {
			return result, nil
		}

		return time.Time{}, fmt.Errorf("failed to parse datetime '%s': %w", v, err)
	case int64:
		// Assume Unix timestamp
		return time.Unix(v, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported type for datetime conversion: %T", value)
	}
}

// ConvertToString safely converts any value to string
func (tc *TypeConverter) ConvertToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int64, int32, int16, int8:
		return fmt.Sprintf("%d", v)
	case uint, uint64, uint32, uint16, uint8:
		return fmt.Sprintf("%d", v)
	case float64, float32:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case time.Time:
		return v.Format(types.DateTimeFormat)
	case []interface{}, map[string]interface{}:
		// For complex types, marshal as JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Global type converter instance
var TypeConv = &TypeConverter{}

// OutputBuilder helps build output maps with validation
type OutputBuilder struct {
	outputs map[string]interface{}
	errors  []error
}

// NewOutputBuilder creates a new OutputBuilder
func NewOutputBuilder() *OutputBuilder {
	return &OutputBuilder{
		outputs: make(map[string]interface{}),
		errors:  make([]error, 0),
	}
}

// Set sets an output value with optional type validation
func (ob *OutputBuilder) Set(name string, value interface{}) *OutputBuilder {
	ob.outputs[name] = value
	return ob
}

// SetString sets a string output with validation
func (ob *OutputBuilder) SetString(name string, value interface{}) *OutputBuilder {
	strValue := TypeConv.ConvertToString(value)
	ob.outputs[name] = strValue
	return ob
}

// SetInt sets an int output with validation
func (ob *OutputBuilder) SetInt(name string, value interface{}) *OutputBuilder {
	intValue, err := TypeConv.ConvertToInt(value)
	if err != nil {
		ob.errors = append(ob.errors, fmt.Errorf("output '%s': %w", name, err))
		ob.outputs[name] = 0
	} else {
		ob.outputs[name] = intValue
	}
	return ob
}

// SetFloat sets a float output with validation
func (ob *OutputBuilder) SetFloat(name string, value interface{}) *OutputBuilder {
	floatValue, err := TypeConv.ConvertToFloat(value)
	if err != nil {
		ob.errors = append(ob.errors, fmt.Errorf("output '%s': %w", name, err))
		ob.outputs[name] = 0.0
	} else {
		ob.outputs[name] = floatValue
	}
	return ob
}

// SetBool sets a bool output with validation
func (ob *OutputBuilder) SetBool(name string, value interface{}) *OutputBuilder {
	boolValue, err := TypeConv.ConvertToBool(value)
	if err != nil {
		ob.errors = append(ob.errors, fmt.Errorf("output '%s': %w", name, err))
		ob.outputs[name] = false
	} else {
		ob.outputs[name] = boolValue
	}
	return ob
}

// SetDateTime sets a datetime output with validation
func (ob *OutputBuilder) SetDateTime(name string, value interface{}) *OutputBuilder {
	dateValue, err := TypeConv.ConvertToDateTime(value)
	if err != nil {
		ob.errors = append(ob.errors, fmt.Errorf("output '%s': %w", name, err))
		ob.outputs[name] = time.Time{}
	} else {
		ob.outputs[name] = dateValue
	}
	return ob
}

// Build returns the output map and any errors encountered
func (ob *OutputBuilder) Build() (map[string]interface{}, error) {
	if len(ob.errors) > 0 {
		// Combine all errors into one
		errMsgs := make([]string, len(ob.errors))
		for i, err := range ob.errors {
			errMsgs[i] = err.Error()
		}
		return ob.outputs, fmt.Errorf("output validation errors: %s", strings.Join(errMsgs, "; "))
	}
	return ob.outputs, nil
}

// GetOutputs returns the outputs map (even if there were errors)
func (ob *OutputBuilder) GetOutputs() map[string]interface{} {
	return ob.outputs
}

// SessionHelper provides convenient session operations
type SessionHelper struct {
	systemSession map[string]interface{}
	userSession   map[string]interface{}
}

// NewSessionHelper creates a new SessionHelper
func NewSessionHelper(systemSession, userSession map[string]interface{}) *SessionHelper {
	return &SessionHelper{
		systemSession: systemSession,
		userSession:   userSession,
	}
}

// GetSystemString gets a string from system session with default
func (sh *SessionHelper) GetSystemString(key, defaultValue string) string {
	if sh.systemSession == nil {
		return defaultValue
	}
	value, exists := sh.systemSession[key]
	if !exists || value == nil {
		return defaultValue
	}
	return TypeConv.ConvertToString(value)
}

// GetUserString gets a string from user session with default
func (sh *SessionHelper) GetUserString(key, defaultValue string) string {
	if sh.userSession == nil {
		return defaultValue
	}
	value, exists := sh.userSession[key]
	if !exists || value == nil {
		return defaultValue
	}
	return TypeConv.ConvertToString(value)
}

// GetSystemInt gets an int from system session with default
func (sh *SessionHelper) GetSystemInt(key string, defaultValue int) int {
	if sh.systemSession == nil {
		return defaultValue
	}
	value, exists := sh.systemSession[key]
	if !exists || value == nil {
		return defaultValue
	}
	intValue, err := TypeConv.ConvertToInt(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetUserInt gets an int from user session with default
func (sh *SessionHelper) GetUserInt(key string, defaultValue int) int {
	if sh.userSession == nil {
		return defaultValue
	}
	value, exists := sh.userSession[key]
	if !exists || value == nil {
		return defaultValue
	}
	intValue, err := TypeConv.ConvertToInt(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// SetSystemValue sets a value in system session
func (sh *SessionHelper) SetSystemValue(key string, value interface{}) {
	if sh.systemSession != nil {
		sh.systemSession[key] = value
	}
}

// SetUserValue sets a value in user session
func (sh *SessionHelper) SetUserValue(key string, value interface{}) {
	if sh.userSession != nil {
		sh.userSession[key] = value
	}
}

// JSONHelper provides JSON marshaling/unmarshaling utilities
type JSONHelper struct{}

// Marshal safely marshals a value to JSON string
func (jh *JSONHelper) Marshal(value interface{}) (string, error) {
	if value == nil {
		return "{}", nil
	}

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// Unmarshal safely unmarshals JSON string to map
func (jh *JSONHelper) Unmarshal(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return result, nil
}

// UnmarshalToSlice safely unmarshals JSON string to slice
func (jh *JSONHelper) UnmarshalToSlice(jsonStr string) ([]interface{}, error) {
	if jsonStr == "" {
		return make([]interface{}, 0), nil
	}

	var result []interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON array: %w", err)
	}

	return result, nil
}

// MarshalIndent marshals with indentation for readability
func (jh *JSONHelper) MarshalIndent(value interface{}) (string, error) {
	if value == nil {
		return "{}", nil
	}

	jsonBytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// Global JSON helper instance
var JSON = &JSONHelper{}

// SliceHelper provides slice manipulation utilities
type SliceHelper struct{}

// ContainsString checks if a string slice contains a value
func (sh *SliceHelper) ContainsString(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// ContainsInt checks if an int slice contains a value
func (sh *SliceHelper) ContainsInt(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// UniqueStrings returns unique strings from a slice
func (sh *SliceHelper) UniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// FilterStrings filters a string slice by a predicate function
func (sh *SliceHelper) FilterStrings(slice []string, predicate func(string) bool) []string {
	result := make([]string, 0)
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Global slice helper instance
var Slice = &SliceHelper{}

// StringHelper provides string manipulation utilities
type StringHelper struct{}

// IsEmpty checks if a string is empty or whitespace only
func (sh *StringHelper) IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// DefaultIfEmpty returns defaultValue if s is empty
func (sh *StringHelper) DefaultIfEmpty(s, defaultValue string) string {
	if sh.IsEmpty(s) {
		return defaultValue
	}
	return s
}

// Truncate truncates a string to maxLength
func (sh *StringHelper) Truncate(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// SplitAndTrim splits a string and trims each part
func (sh *StringHelper) SplitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// Global string helper instance
var Str = &StringHelper{}
