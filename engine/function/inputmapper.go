package funcs

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

// InputMapper handles comprehensive input mapping and validation
type InputMapper struct {
	SystemSession       map[string]interface{}
	UserSession         map[string]interface{}
	Externalinputs      map[string]interface{}
	FuncCachedVariables map[string]interface{}
	Log                 logger.Log
}

// NewInputMapper creates a new InputMapper instance
func NewInputMapper(systemSession, userSession, externalInputs, funcCachedVariables map[string]interface{}, log logger.Log) *InputMapper {
	return &InputMapper{
		SystemSession:       systemSession,
		UserSession:         userSession,
		Externalinputs:      externalInputs,
		FuncCachedVariables: funcCachedVariables,
		Log:                 log,
	}
}

// MapInput maps a single input from its source to a typed value
func (im *InputMapper) MapInput(input types.Input) (interface{}, error) {
	im.Log.Debug(fmt.Sprintf("Mapping input: %s, Source: %v, DataType: %v, List: %v", input.Name, input.Source, input.Datatype, input.List))

	// Step 1: Get raw value from source
	rawValue, err := im.getValueFromSource(input)
	if err != nil {
		// If source retrieval fails, try to use default value
		if input.Defaultvalue != "" {
			im.Log.Debug(fmt.Sprintf("Using default value for input %s: %s", input.Name, input.Defaultvalue))
			rawValue = input.Defaultvalue
		} else if input.Inivalue != "" {
			im.Log.Debug(fmt.Sprintf("Using initial value for input %s: %s", input.Name, input.Inivalue))
			rawValue = input.Inivalue
		} else {
			// Return error if no fallback is available
			return nil, types.NewValidationError(fmt.Sprintf("Failed to get value for input %s", input.Name), err).
				WithContext(&types.ExecutionContext{
					FunctionName: "InputMapper.MapInput",
				}).
				WithDetail("input_name", input.Name).
				WithDetail("source", input.Source.String())
		}
	}

	// Step 2: Convert to appropriate type
	typedValue, err := im.convertToType(input.Name, rawValue, input.Datatype, input.List)
	if err != nil {
		return nil, err
	}

	return typedValue, nil
}

// MapAllInputs maps all inputs from a function definition
func (im *InputMapper) MapAllInputs(inputs []types.Input) (map[string]interface{}, []string, []string, error) {
	mappedInputs := make(map[string]interface{})
	nameList := make([]string, len(inputs))
	valueList := make([]string, len(inputs))

	for i, input := range inputs {
		value, err := im.MapInput(input)
		if err != nil {
			im.Log.Error(fmt.Sprintf("Error mapping input %s: %s", input.Name, err.Error()))

			// Check if this is a critical error or if we can use default
			if input.Defaultvalue == "" && input.Inivalue == "" {
				return nil, nil, nil, err
			}

			// Use default/initial value
			if input.Defaultvalue != "" {
				value, _ = im.convertToType(input.Name, input.Defaultvalue, input.Datatype, input.List)
			} else if input.Inivalue != "" {
				value, _ = im.convertToType(input.Name, input.Inivalue, input.Datatype, input.List)
			}
		}

		mappedInputs[input.Name] = value
		nameList[i] = input.Name

		// Convert value to string for valueList
		valueList[i] = im.valueToString(value)
	}

	return mappedInputs, nameList, valueList, nil
}

// getValueFromSource retrieves the raw value from the specified input source
func (im *InputMapper) getValueFromSource(input types.Input) (string, error) {
	switch input.Source {
	case types.Constant:
		// Constant values come from the Value field
		if input.Value != "" {
			return input.Value, nil
		}
		if input.Inivalue != "" {
			return input.Inivalue, nil
		}
		return input.Defaultvalue, nil

	case types.Fromsyssession:
		// Get from system session
		if im.SystemSession == nil {
			return "", fmt.Errorf("system session is nil")
		}

		value, exists := im.SystemSession[input.Aliasname]
		if !exists {
			return "", fmt.Errorf("key '%s' not found in system session", input.Aliasname)
		}

		if value == nil {
			return "", fmt.Errorf("value for key '%s' is nil in system session", input.Aliasname)
		}

		// Safe type conversion to string
		strValue, err := im.convertValueToString(value)
		if err != nil {
			return "", fmt.Errorf("failed to convert system session value to string: %w", err)
		}

		return strValue, nil

	case types.Fromusersession:
		// Get from user session
		if im.UserSession == nil {
			return "", fmt.Errorf("user session is nil")
		}

		value, exists := im.UserSession[input.Aliasname]
		if !exists {
			return "", fmt.Errorf("key '%s' not found in user session", input.Aliasname)
		}

		if value == nil {
			return "", fmt.Errorf("value for key '%s' is nil in user session", input.Aliasname)
		}

		// Safe type conversion to string
		strValue, err := im.convertValueToString(value)
		if err != nil {
			return "", fmt.Errorf("failed to convert user session value to string: %w", err)
		}

		return strValue, nil

	case types.Prefunction:
		// Get from previous function's cached variables
		// Format: "FunctionName.VariableName"
		parts := strings.Split(input.Aliasname, ".")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid prefunction alias format: %s (expected 'FunctionName.VariableName')", input.Aliasname)
		}

		funcName := parts[0]
		varName := parts[1]

		if im.FuncCachedVariables == nil {
			return "", fmt.Errorf("function cached variables is nil")
		}

		funcVars, exists := im.FuncCachedVariables[funcName]
		if !exists {
			return "", fmt.Errorf("function '%s' not found in cached variables", funcName)
		}

		if funcVars == nil {
			return "", fmt.Errorf("cached variables for function '%s' is nil", funcName)
		}

		// Safe type assertion to map
		funcVarsMap, err := types.AssertMap(funcVars, fmt.Sprintf("FuncCachedVariables[%s]", funcName))
		if err != nil {
			return "", fmt.Errorf("failed to access function variables: %w", err)
		}

		value, exists := funcVarsMap[varName]
		if !exists {
			return "", fmt.Errorf("variable '%s' not found in function '%s'", varName, funcName)
		}

		if value == nil {
			return "", fmt.Errorf("value for variable '%s' in function '%s' is nil", varName, funcName)
		}

		// Safe type conversion to string
		strValue, err := im.convertValueToString(value)
		if err != nil {
			return "", fmt.Errorf("failed to convert cached variable value to string: %w", err)
		}

		return strValue, nil

	case types.Fromexternal:
		// Get from external inputs
		if im.Externalinputs == nil {
			return "", fmt.Errorf("external inputs is nil")
		}

		value, exists := im.Externalinputs[input.Aliasname]
		if !exists {
			return "", fmt.Errorf("key '%s' not found in external inputs", input.Aliasname)
		}

		if value == nil {
			return "", fmt.Errorf("value for key '%s' is nil in external inputs", input.Aliasname)
		}

		// Safe type conversion to string
		strValue, err := im.convertValueToString(value)
		if err != nil {
			return "", fmt.Errorf("failed to convert external input value to string: %w", err)
		}

		return strValue, nil

	default:
		return "", fmt.Errorf("unsupported input source: %v", input.Source)
	}
}

// convertValueToString safely converts various types to string
func (im *InputMapper) convertValueToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case int64:
		return fmt.Sprintf("%d", v), nil
	case float64:
		return fmt.Sprintf("%f", v), nil
	case float32:
		return fmt.Sprintf("%f", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	case time.Time:
		return v.Format(types.DateTimeFormat), nil
	case []interface{}, map[string]interface{}:
		// For complex types, marshal to JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal complex value: %w", err)
		}
		return string(jsonBytes), nil
	case []string:
		// For string arrays, marshal to JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal string array: %w", err)
		}
		return string(jsonBytes), nil
	default:
		// Try generic conversion
		return fmt.Sprintf("%v", v), nil
	}
}

// convertToType converts a string value to the appropriate type
func (im *InputMapper) convertToType(name string, rawValue string, dataType types.DataType, isList bool) (interface{}, error) {
	im.Log.Debug(fmt.Sprintf("Converting input %s: value=%s, type=%v, isList=%v", name, rawValue, dataType, isList))

	// Handle list types
	if isList {
		return im.convertToListType(name, rawValue, dataType)
	}

	// Handle single value types
	switch dataType {
	case types.String:
		return rawValue, nil

	case types.Integer:
		return im.parseInteger(rawValue)

	case types.Float:
		return im.parseFloat(rawValue)

	case types.Bool:
		return im.parseBool(rawValue)

	case types.DateTime:
		return im.parseDateTime(rawValue)

	case types.Object:
		return im.parseObject(rawValue)

	default:
		return rawValue, nil
	}
}

// convertToListType converts a string value to a list of the appropriate type
func (im *InputMapper) convertToListType(name string, rawValue string, dataType types.DataType) (interface{}, error) {
	im.Log.Debug(fmt.Sprintf("Converting list input %s: value=%s, type=%v", name, rawValue, dataType))

	switch dataType {
	case types.String:
		return im.parseStringList(rawValue)

	case types.Integer:
		return im.parseIntegerList(rawValue)

	case types.Float:
		return im.parseFloatList(rawValue)

	case types.Bool:
		return im.parseBoolList(rawValue)

	case types.DateTime:
		return im.parseDateTimeList(rawValue)

	case types.Object:
		return im.parseObjectList(rawValue)

	default:
		return im.parseStringList(rawValue)
	}
}

// Integer parsing
func (im *InputMapper) parseInteger(value string) (int, error) {
	if value == "" {
		return 0, nil
	}

	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		// Try parsing as float and converting to int
		var floatResult float64
		_, err2 := fmt.Sscanf(value, "%f", &floatResult)
		if err2 != nil {
			return 0, fmt.Errorf("failed to parse integer value '%s': %w", value, err)
		}
		return int(floatResult), nil
	}
	return result, nil
}

func (im *InputMapper) parseIntegerList(value string) ([]int, error) {
	if value == "" {
		return []int{}, nil
	}

	// Try to unmarshal as JSON array first
	var intList []int
	err := json.Unmarshal([]byte(value), &intList)
	if err == nil {
		return intList, nil
	}

	// Try as string array
	var strList []string
	err = json.Unmarshal([]byte(value), &strList)
	if err == nil {
		result := make([]int, len(strList))
		for i, s := range strList {
			result[i], err = im.parseInteger(s)
			if err != nil {
				return nil, fmt.Errorf("failed to parse integer at index %d: %w", i, err)
			}
		}
		return result, nil
	}

	// Try as float array
	var floatList []float64
	err = json.Unmarshal([]byte(value), &floatList)
	if err == nil {
		result := make([]int, len(floatList))
		for i, f := range floatList {
			result[i] = int(f)
		}
		return result, nil
	}

	// Try parsing as comma-separated or bracket-enclosed values
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimPrefix(trimmed, "[")
	trimmed = strings.TrimSuffix(trimmed, "]")

	parts := strings.Split(trimmed, ",")
	result := make([]int, len(parts))
	for i, part := range parts {
		result[i], err = im.parseInteger(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("failed to parse integer at index %d: %w", i, err)
		}
	}

	return result, nil
}

// Float parsing
func (im *InputMapper) parseFloat(value string) (float64, error) {
	if value == "" {
		return 0.0, nil
	}

	var result float64
	_, err := fmt.Sscanf(value, "%f", &result)
	if err != nil {
		return 0.0, fmt.Errorf("failed to parse float value '%s': %w", value, err)
	}
	return result, nil
}

func (im *InputMapper) parseFloatList(value string) ([]float64, error) {
	if value == "" {
		return []float64{}, nil
	}

	// Try to unmarshal as JSON array
	var floatList []float64
	err := json.Unmarshal([]byte(value), &floatList)
	if err == nil {
		return floatList, nil
	}

	// Try as string array
	var strList []string
	err = json.Unmarshal([]byte(value), &strList)
	if err == nil {
		result := make([]float64, len(strList))
		for i, s := range strList {
			result[i], err = im.parseFloat(s)
			if err != nil {
				return nil, fmt.Errorf("failed to parse float at index %d: %w", i, err)
			}
		}
		return result, nil
	}

	// Try parsing as comma-separated values
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimPrefix(trimmed, "[")
	trimmed = strings.TrimSuffix(trimmed, "]")

	parts := strings.Split(trimmed, ",")
	result := make([]float64, len(parts))
	for i, part := range parts {
		result[i], err = im.parseFloat(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("failed to parse float at index %d: %w", i, err)
		}
	}

	return result, nil
}

// Bool parsing
func (im *InputMapper) parseBool(value string) (bool, error) {
	if value == "" {
		return false, nil
	}

	lowerValue := strings.ToLower(strings.TrimSpace(value))
	switch lowerValue {
	case "true", "1", "yes", "y", "on":
		return true, nil
	case "false", "0", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: '%s'", value)
	}
}

func (im *InputMapper) parseBoolList(value string) ([]bool, error) {
	if value == "" {
		return []bool{}, nil
	}

	// Try to unmarshal as JSON array
	var boolList []bool
	err := json.Unmarshal([]byte(value), &boolList)
	if err == nil {
		return boolList, nil
	}

	// Try as string array
	var strList []string
	err = json.Unmarshal([]byte(value), &strList)
	if err == nil {
		result := make([]bool, len(strList))
		for i, s := range strList {
			result[i], err = im.parseBool(s)
			if err != nil {
				return nil, fmt.Errorf("failed to parse bool at index %d: %w", i, err)
			}
		}
		return result, nil
	}

	// Try parsing as comma-separated values
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimPrefix(trimmed, "[")
	trimmed = strings.TrimSuffix(trimmed, "]")

	parts := strings.Split(trimmed, ",")
	result := make([]bool, len(parts))
	for i, part := range parts {
		result[i], err = im.parseBool(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("failed to parse bool at index %d: %w", i, err)
		}
	}

	return result, nil
}

// DateTime parsing
func (im *InputMapper) parseDateTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	// Try standard format first
	result, err := time.Parse(types.DateTimeFormat, value)
	if err == nil {
		return result, nil
	}

	// Try ISO 8601
	result, err = time.Parse(time.RFC3339, value)
	if err == nil {
		return result, nil
	}

	// Try date only
	result, err = time.Parse("2006-01-02", value)
	if err == nil {
		return result, nil
	}

	return time.Time{}, fmt.Errorf("failed to parse datetime value '%s': %w", value, err)
}

func (im *InputMapper) parseDateTimeList(value string) ([]time.Time, error) {
	if value == "" {
		return []time.Time{}, nil
	}

	// Try to unmarshal as JSON array of strings
	var strList []string
	err := json.Unmarshal([]byte(value), &strList)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal datetime list: %w", err)
	}

	result := make([]time.Time, len(strList))
	for i, s := range strList {
		result[i], err = im.parseDateTime(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse datetime at index %d: %w", i, err)
		}
	}

	return result, nil
}

// String parsing
func (im *InputMapper) parseStringList(value string) ([]string, error) {
	if value == "" {
		return []string{}, nil
	}

	// Try to unmarshal as JSON array
	var strList []string
	err := json.Unmarshal([]byte(value), &strList)
	if err == nil {
		return strList, nil
	}

	// If not JSON, treat as single value
	return []string{value}, nil
}

// Object parsing
func (im *InputMapper) parseObject(value string) (map[string]interface{}, error) {
	if value == "" {
		return map[string]interface{}{}, nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(value), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse object (JSON): %w", err)
	}

	return result, nil
}

func (im *InputMapper) parseObjectList(value string) ([]map[string]interface{}, error) {
	if value == "" {
		return []map[string]interface{}{}, nil
	}

	var result []map[string]interface{}
	err := json.Unmarshal([]byte(value), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse object list (JSON array): %w", err)
	}

	return result, nil
}

// valueToString converts any value to string for logging/display
func (im *InputMapper) valueToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case time.Time:
		return v.Format(types.DateTimeFormat)
	case []interface{}, map[string]interface{}, []string, []int, []float64, []bool, []time.Time:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	default:
		return fmt.Sprintf("%v", v)
	}
}
