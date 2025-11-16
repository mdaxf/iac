package types

import (
	"fmt"
	"time"
)

// TypeSafeSession provides type-safe access to session data
type TypeSafeSession struct {
	data map[string]interface{}
}

// NewTypeSafeSession creates a new type-safe session wrapper
func NewTypeSafeSession(data map[string]interface{}) *TypeSafeSession {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &TypeSafeSession{data: data}
}

// GetString safely gets a string value from the session
func (ts *TypeSafeSession) GetString(key string) (string, error) {
	val, exists := ts.data[key]
	if !exists {
		return "", fmt.Errorf("key '%s' not found in session", key)
	}

	strVal, ok := val.(string)
	if !ok {
		return "", NewTypeAssertionError("string", fmt.Sprintf("%T", val), key)
	}
	return strVal, nil
}

// GetStringOr safely gets a string value or returns a default
func (ts *TypeSafeSession) GetStringOr(key, defaultVal string) string {
	val, err := ts.GetString(key)
	if err != nil {
		return defaultVal
	}
	return val
}

// GetInt safely gets an int value from the session
func (ts *TypeSafeSession) GetInt(key string) (int, error) {
	val, exists := ts.data[key]
	if !exists {
		return 0, fmt.Errorf("key '%s' not found in session", key)
	}

	// Try int first
	if intVal, ok := val.(int); ok {
		return intVal, nil
	}

	// Try int64
	if int64Val, ok := val.(int64); ok {
		return int(int64Val), nil
	}

	// Try float64 (JSON unmarshaling sometimes creates float64)
	if float64Val, ok := val.(float64); ok {
		return int(float64Val), nil
	}

	return 0, NewTypeAssertionError("int", fmt.Sprintf("%T", val), key)
}

// GetIntOr safely gets an int value or returns a default
func (ts *TypeSafeSession) GetIntOr(key string, defaultVal int) int {
	val, err := ts.GetInt(key)
	if err != nil {
		return defaultVal
	}
	return val
}

// GetBool safely gets a bool value from the session
func (ts *TypeSafeSession) GetBool(key string) (bool, error) {
	val, exists := ts.data[key]
	if !exists {
		return false, fmt.Errorf("key '%s' not found in session", key)
	}

	boolVal, ok := val.(bool)
	if !ok {
		return false, NewTypeAssertionError("bool", fmt.Sprintf("%T", val), key)
	}
	return boolVal, nil
}

// GetBoolOr safely gets a bool value or returns a default
func (ts *TypeSafeSession) GetBoolOr(key string, defaultVal bool) bool {
	val, err := ts.GetBool(key)
	if err != nil {
		return defaultVal
	}
	return val
}

// GetTime safely gets a time.Time value from the session
func (ts *TypeSafeSession) GetTime(key string) (time.Time, error) {
	val, exists := ts.data[key]
	if !exists {
		return time.Time{}, fmt.Errorf("key '%s' not found in session", key)
	}

	timeVal, ok := val.(time.Time)
	if !ok {
		return time.Time{}, NewTypeAssertionError("time.Time", fmt.Sprintf("%T", val), key)
	}
	return timeVal, nil
}

// GetTimeOr safely gets a time.Time value or returns a default
func (ts *TypeSafeSession) GetTimeOr(key string, defaultVal time.Time) time.Time {
	val, err := ts.GetTime(key)
	if err != nil {
		return defaultVal
	}
	return val
}

// GetMap safely gets a map value from the session
func (ts *TypeSafeSession) GetMap(key string) (map[string]interface{}, error) {
	val, exists := ts.data[key]
	if !exists {
		return nil, fmt.Errorf("key '%s' not found in session", key)
	}

	mapVal, ok := val.(map[string]interface{})
	if !ok {
		return nil, NewTypeAssertionError("map[string]interface{}", fmt.Sprintf("%T", val), key)
	}
	return mapVal, nil
}

// GetMapOr safely gets a map value or returns a default
func (ts *TypeSafeSession) GetMapOr(key string, defaultVal map[string]interface{}) map[string]interface{} {
	val, err := ts.GetMap(key)
	if err != nil {
		return defaultVal
	}
	return val
}

// Set sets a value in the session
func (ts *TypeSafeSession) Set(key string, value interface{}) {
	ts.data[key] = value
}

// Has checks if a key exists in the session
func (ts *TypeSafeSession) Has(key string) bool {
	_, exists := ts.data[key]
	return exists
}

// GetRaw gets the raw value without type checking
func (ts *TypeSafeSession) GetRaw(key string) (interface{}, bool) {
	val, exists := ts.data[key]
	return val, exists
}

// GetData returns the underlying map (use with caution)
func (ts *TypeSafeSession) GetData() map[string]interface{} {
	return ts.data
}

// TypeSafeGetter provides type-safe access to map values
type TypeSafeGetter struct{}

// SafeGetString safely gets a string from a map
func (tsg TypeSafeGetter) SafeGetString(data map[string]interface{}, key string) (string, error) {
	val, exists := data[key]
	if !exists {
		return "", fmt.Errorf("key '%s' not found", key)
	}

	strVal, ok := val.(string)
	if !ok {
		return "", NewTypeAssertionError("string", fmt.Sprintf("%T", val), key)
	}
	return strVal, nil
}

// SafeGetStringOr safely gets a string or returns a default
func (tsg TypeSafeGetter) SafeGetStringOr(data map[string]interface{}, key, defaultVal string) string {
	val, err := tsg.SafeGetString(data, key)
	if err != nil {
		return defaultVal
	}
	return val
}

// SafeGetInt safely gets an int from a map
func (tsg TypeSafeGetter) SafeGetInt(data map[string]interface{}, key string) (int, error) {
	val, exists := data[key]
	if !exists {
		return 0, fmt.Errorf("key '%s' not found", key)
	}

	// Try int first
	if intVal, ok := val.(int); ok {
		return intVal, nil
	}

	// Try int64
	if int64Val, ok := val.(int64); ok {
		return int(int64Val), nil
	}

	// Try float64
	if float64Val, ok := val.(float64); ok {
		return int(float64Val), nil
	}

	return 0, NewTypeAssertionError("int", fmt.Sprintf("%T", val), key)
}

// SafeGetIntOr safely gets an int or returns a default
func (tsg TypeSafeGetter) SafeGetIntOr(data map[string]interface{}, key string, defaultVal int) int {
	val, err := tsg.SafeGetInt(data, key)
	if err != nil {
		return defaultVal
	}
	return val
}

// SafeGetBool safely gets a bool from a map
func (tsg TypeSafeGetter) SafeGetBool(data map[string]interface{}, key string) (bool, error) {
	val, exists := data[key]
	if !exists {
		return false, fmt.Errorf("key '%s' not found", key)
	}

	boolVal, ok := val.(bool)
	if !ok {
		return false, NewTypeAssertionError("bool", fmt.Sprintf("%T", val), key)
	}
	return boolVal, nil
}

// SafeGetBoolOr safely gets a bool or returns a default
func (tsg TypeSafeGetter) SafeGetBoolOr(data map[string]interface{}, key string, defaultVal bool) bool {
	val, err := tsg.SafeGetBool(data, key)
	if err != nil {
		return defaultVal
	}
	return val
}

// SafeGetMap safely gets a map from a map
func (tsg TypeSafeGetter) SafeGetMap(data map[string]interface{}, key string) (map[string]interface{}, error) {
	val, exists := data[key]
	if !exists {
		return nil, fmt.Errorf("key '%s' not found", key)
	}

	if val == nil {
		return nil, nil
	}

	mapVal, ok := val.(map[string]interface{})
	if !ok {
		return nil, NewTypeAssertionError("map[string]interface{}", fmt.Sprintf("%T", val), key)
	}
	return mapVal, nil
}

// SafeGetMapOr safely gets a map or returns a default
func (tsg TypeSafeGetter) SafeGetMapOr(data map[string]interface{}, key string, defaultVal map[string]interface{}) map[string]interface{} {
	val, err := tsg.SafeGetMap(data, key)
	if err != nil {
		return defaultVal
	}
	return val
}

// Global type-safe getter instance
var SafeGetter = TypeSafeGetter{}

// AssertString safely asserts a value to string
func AssertString(value interface{}, context string) (string, error) {
	if value == nil {
		return "", fmt.Errorf("nil value in context: %s", context)
	}

	strVal, ok := value.(string)
	if !ok {
		return "", NewTypeAssertionError("string", fmt.Sprintf("%T", value), context)
	}
	return strVal, nil
}

// AssertInt safely asserts a value to int
func AssertInt(value interface{}, context string) (int, error) {
	if value == nil {
		return 0, fmt.Errorf("nil value in context: %s", context)
	}

	// Try int first
	if intVal, ok := value.(int); ok {
		return intVal, nil
	}

	// Try int64
	if int64Val, ok := value.(int64); ok {
		return int(int64Val), nil
	}

	// Try float64
	if float64Val, ok := value.(float64); ok {
		return int(float64Val), nil
	}

	return 0, NewTypeAssertionError("int", fmt.Sprintf("%T", value), context)
}

// AssertBool safely asserts a value to bool
func AssertBool(value interface{}, context string) (bool, error) {
	if value == nil {
		return false, fmt.Errorf("nil value in context: %s", context)
	}

	boolVal, ok := value.(bool)
	if !ok {
		return false, NewTypeAssertionError("bool", fmt.Sprintf("%T", value), context)
	}
	return boolVal, nil
}

// AssertMap safely asserts a value to map[string]interface{}
func AssertMap(value interface{}, context string) (map[string]interface{}, error) {
	if value == nil {
		return nil, fmt.Errorf("nil value in context: %s", context)
	}

	mapVal, ok := value.(map[string]interface{})
	if !ok {
		return nil, NewTypeAssertionError("map[string]interface{}", fmt.Sprintf("%T", value), context)
	}
	return mapVal, nil
}

// AssertStringSlice safely asserts a value to []string
func AssertStringSlice(value interface{}, context string) ([]string, error) {
	if value == nil {
		return nil, fmt.Errorf("nil value in context: %s", context)
	}

	sliceVal, ok := value.([]string)
	if !ok {
		return nil, NewTypeAssertionError("[]string", fmt.Sprintf("%T", value), context)
	}
	return sliceVal, nil
}

// AssertIntSlice safely asserts a value to []int
func AssertIntSlice(value interface{}, context string) ([]int, error) {
	if value == nil {
		return nil, fmt.Errorf("nil value in context: %s", context)
	}

	sliceVal, ok := value.([]int)
	if !ok {
		return nil, NewTypeAssertionError("[]int", fmt.Sprintf("%T", value), context)
	}
	return sliceVal, nil
}
