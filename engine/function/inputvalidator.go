package funcs

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mdaxf/iac/engine/types"
)

// ValidationRule defines a validation rule for an input
type ValidationRule struct {
	Type    ValidationType
	Param   interface{}
	Message string
}

// ValidationType defines the type of validation
type ValidationType int

const (
	ValidationRequired ValidationType = iota
	ValidationMinLength
	ValidationMaxLength
	ValidationMinValue
	ValidationMaxValue
	ValidationRegex
	ValidationCustom
	ValidationEnum
)

// String returns the string representation of ValidationType
func (vt ValidationType) String() string {
	switch vt {
	case ValidationRequired:
		return "Required"
	case ValidationMinLength:
		return "MinLength"
	case ValidationMaxLength:
		return "MaxLength"
	case ValidationMinValue:
		return "MinValue"
	case ValidationMaxValue:
		return "MaxValue"
	case ValidationRegex:
		return "Regex"
	case ValidationCustom:
		return "Custom"
	case ValidationEnum:
		return "Enum"
	default:
		return "Unknown"
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Rule    ValidationType
	Message string
}

// Error returns the error message
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s' [%s]: %s", ve.Field, ve.Rule.String(), ve.Message)
}

// ValidationResult holds the results of validation
type ValidationResult struct {
	IsValid bool
	Errors  []*ValidationError
}

// AddError adds a validation error
func (vr *ValidationResult) AddError(field string, rule ValidationType, message string) {
	vr.IsValid = false
	vr.Errors = append(vr.Errors, &ValidationError{
		Field:   field,
		Rule:    rule,
		Message: message,
	})
}

// GetErrorMessages returns all error messages
func (vr *ValidationResult) GetErrorMessages() []string {
	messages := make([]string, len(vr.Errors))
	for i, err := range vr.Errors {
		messages[i] = err.Error()
	}
	return messages
}

// InputValidator handles comprehensive input validation
type InputValidator struct {
	Rules map[string][]ValidationRule
}

// NewInputValidator creates a new InputValidator
func NewInputValidator() *InputValidator {
	return &InputValidator{
		Rules: make(map[string][]ValidationRule),
	}
}

// AddRule adds a validation rule for a specific input field
func (iv *InputValidator) AddRule(fieldName string, rule ValidationRule) {
	if iv.Rules[fieldName] == nil {
		iv.Rules[fieldName] = []ValidationRule{}
	}
	iv.Rules[fieldName] = append(iv.Rules[fieldName], rule)
}

// AddRequiredRule adds a required validation rule
func (iv *InputValidator) AddRequiredRule(fieldName string) {
	iv.AddRule(fieldName, ValidationRule{
		Type:    ValidationRequired,
		Message: fmt.Sprintf("Field '%s' is required", fieldName),
	})
}

// AddMinLengthRule adds a minimum length validation rule
func (iv *InputValidator) AddMinLengthRule(fieldName string, minLength int) {
	iv.AddRule(fieldName, ValidationRule{
		Type:    ValidationMinLength,
		Param:   minLength,
		Message: fmt.Sprintf("Field '%s' must be at least %d characters", fieldName, minLength),
	})
}

// AddMaxLengthRule adds a maximum length validation rule
func (iv *InputValidator) AddMaxLengthRule(fieldName string, maxLength int) {
	iv.AddRule(fieldName, ValidationRule{
		Type:    ValidationMaxLength,
		Param:   maxLength,
		Message: fmt.Sprintf("Field '%s' must be at most %d characters", fieldName, maxLength),
	})
}

// AddMinValueRule adds a minimum value validation rule (for numbers)
func (iv *InputValidator) AddMinValueRule(fieldName string, minValue float64) {
	iv.AddRule(fieldName, ValidationRule{
		Type:    ValidationMinValue,
		Param:   minValue,
		Message: fmt.Sprintf("Field '%s' must be at least %v", fieldName, minValue),
	})
}

// AddMaxValueRule adds a maximum value validation rule (for numbers)
func (iv *InputValidator) AddMaxValueRule(fieldName string, maxValue float64) {
	iv.AddRule(fieldName, ValidationRule{
		Type:    ValidationMaxValue,
		Param:   maxValue,
		Message: fmt.Sprintf("Field '%s' must be at most %v", fieldName, maxValue),
	})
}

// AddRegexRule adds a regex pattern validation rule
func (iv *InputValidator) AddRegexRule(fieldName string, pattern string, message string) {
	if message == "" {
		message = fmt.Sprintf("Field '%s' does not match required pattern", fieldName)
	}
	iv.AddRule(fieldName, ValidationRule{
		Type:    ValidationRegex,
		Param:   pattern,
		Message: message,
	})
}

// AddEnumRule adds an enum validation rule (value must be in allowed list)
func (iv *InputValidator) AddEnumRule(fieldName string, allowedValues []string) {
	iv.AddRule(fieldName, ValidationRule{
		Type:    ValidationEnum,
		Param:   allowedValues,
		Message: fmt.Sprintf("Field '%s' must be one of: %s", fieldName, strings.Join(allowedValues, ", ")),
	})
}

// AddCustomRule adds a custom validation function
func (iv *InputValidator) AddCustomRule(fieldName string, validatorFunc func(interface{}) error, message string) {
	iv.AddRule(fieldName, ValidationRule{
		Type:    ValidationCustom,
		Param:   validatorFunc,
		Message: message,
	})
}

// Validate validates all inputs against their rules
func (iv *InputValidator) Validate(inputs map[string]interface{}) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []*ValidationError{},
	}

	// Validate each field that has rules
	for fieldName, rules := range iv.Rules {
		value, exists := inputs[fieldName]

		// Check if field exists for required validation
		for _, rule := range rules {
			if rule.Type == ValidationRequired && !exists {
				result.AddError(fieldName, ValidationRequired, rule.Message)
				continue
			}
		}

		// If field doesn't exist and is not required, skip other validations
		if !exists {
			continue
		}

		// Validate against all rules
		for _, rule := range rules {
			err := iv.validateRule(fieldName, value, rule)
			if err != nil {
				result.AddError(fieldName, rule.Type, err.Error())
			}
		}
	}

	return result
}

// validateRule validates a single rule
func (iv *InputValidator) validateRule(fieldName string, value interface{}, rule ValidationRule) error {
	// Skip required check here (already done in Validate)
	if rule.Type == ValidationRequired {
		return nil
	}

	// Handle nil values
	if value == nil {
		return nil // Nil values pass non-required validations
	}

	switch rule.Type {
	case ValidationMinLength:
		return iv.validateMinLength(value, rule)

	case ValidationMaxLength:
		return iv.validateMaxLength(value, rule)

	case ValidationMinValue:
		return iv.validateMinValue(value, rule)

	case ValidationMaxValue:
		return iv.validateMaxValue(value, rule)

	case ValidationRegex:
		return iv.validateRegex(value, rule)

	case ValidationEnum:
		return iv.validateEnum(value, rule)

	case ValidationCustom:
		return iv.validateCustom(value, rule)

	default:
		return nil
	}
}

// validateMinLength validates minimum length for strings
func (iv *InputValidator) validateMinLength(value interface{}, rule ValidationRule) error {
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("value is not a string")
	}

	minLength, ok := rule.Param.(int)
	if !ok {
		return fmt.Errorf("invalid min length parameter")
	}

	if len(strValue) < minLength {
		return fmt.Errorf(rule.Message)
	}

	return nil
}

// validateMaxLength validates maximum length for strings
func (iv *InputValidator) validateMaxLength(value interface{}, rule ValidationRule) error {
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("value is not a string")
	}

	maxLength, ok := rule.Param.(int)
	if !ok {
		return fmt.Errorf("invalid max length parameter")
	}

	if len(strValue) > maxLength {
		return fmt.Errorf(rule.Message)
	}

	return nil
}

// validateMinValue validates minimum value for numbers
func (iv *InputValidator) validateMinValue(value interface{}, rule ValidationRule) error {
	var numValue float64
	var ok bool

	switch v := value.(type) {
	case int:
		numValue = float64(v)
		ok = true
	case int64:
		numValue = float64(v)
		ok = true
	case float64:
		numValue = v
		ok = true
	case float32:
		numValue = float64(v)
		ok = true
	default:
		ok = false
	}

	if !ok {
		return fmt.Errorf("value is not a number")
	}

	minValue, ok := rule.Param.(float64)
	if !ok {
		return fmt.Errorf("invalid min value parameter")
	}

	if numValue < minValue {
		return fmt.Errorf(rule.Message)
	}

	return nil
}

// validateMaxValue validates maximum value for numbers
func (iv *InputValidator) validateMaxValue(value interface{}, rule ValidationRule) error {
	var numValue float64
	var ok bool

	switch v := value.(type) {
	case int:
		numValue = float64(v)
		ok = true
	case int64:
		numValue = float64(v)
		ok = true
	case float64:
		numValue = v
		ok = true
	case float32:
		numValue = float64(v)
		ok = true
	default:
		ok = false
	}

	if !ok {
		return fmt.Errorf("value is not a number")
	}

	maxValue, ok := rule.Param.(float64)
	if !ok {
		return fmt.Errorf("invalid max value parameter")
	}

	if numValue > maxValue {
		return fmt.Errorf(rule.Message)
	}

	return nil
}

// validateRegex validates regex pattern match
func (iv *InputValidator) validateRegex(value interface{}, rule ValidationRule) error {
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("value is not a string")
	}

	pattern, ok := rule.Param.(string)
	if !ok {
		return fmt.Errorf("invalid regex pattern parameter")
	}

	matched, err := regexp.MatchString(pattern, strValue)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	if !matched {
		return fmt.Errorf(rule.Message)
	}

	return nil
}

// validateEnum validates that value is in allowed list
func (iv *InputValidator) validateEnum(value interface{}, rule ValidationRule) error {
	strValue := fmt.Sprintf("%v", value)

	allowedValues, ok := rule.Param.([]string)
	if !ok {
		return fmt.Errorf("invalid enum parameter")
	}

	for _, allowed := range allowedValues {
		if strValue == allowed {
			return nil
		}
	}

	return fmt.Errorf(rule.Message)
}

// validateCustom validates using a custom function
func (iv *InputValidator) validateCustom(value interface{}, rule ValidationRule) error {
	validatorFunc, ok := rule.Param.(func(interface{}) error)
	if !ok {
		return fmt.Errorf("invalid custom validator function")
	}

	err := validatorFunc(value)
	if err != nil {
		if rule.Message != "" {
			return fmt.Errorf(rule.Message)
		}
		return err
	}

	return nil
}

// ValidateInputDefinition validates that an input definition is correctly configured
func ValidateInputDefinition(input types.Input) error {
	// Validate input name
	if input.Name == "" {
		return fmt.Errorf("input name cannot be empty")
	}

	// Validate source-specific requirements
	switch input.Source {
	case types.Fromsyssession, types.Fromusersession, types.Fromexternal:
		if input.Aliasname == "" {
			return fmt.Errorf("input '%s': aliasname is required for source type %v", input.Name, input.Source)
		}

	case types.Prefunction:
		if input.Aliasname == "" {
			return fmt.Errorf("input '%s': aliasname is required for prefunction source", input.Name)
		}
		parts := strings.Split(input.Aliasname, ".")
		if len(parts) != 2 {
			return fmt.Errorf("input '%s': prefunction aliasname must be in format 'FunctionName.VariableName'", input.Name)
		}

	case types.Constant:
		// Constants should have an initial value or default value
		if input.Value == "" && input.Inivalue == "" && input.Defaultvalue == "" {
			return fmt.Errorf("input '%s': constant source requires a value, initial value, or default value", input.Name)
		}
	}

	// Validate datatype compatibility
	if input.List {
		// List inputs should have appropriate default handling
		if input.Datatype == types.Object && input.Defaultvalue != "" {
			// Validate that default value is valid JSON array
			// This would be done during actual mapping
		}
	}

	return nil
}

// Common validation helpers

// IsValidEmail validates email format
func IsValidEmail(email string) error {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(pattern, email)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// IsValidURL validates URL format
func IsValidURL(url string) error {
	pattern := `^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`
	matched, err := regexp.MatchString(pattern, url)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("invalid URL format")
	}
	return nil
}

// IsValidPhoneNumber validates phone number format (simple validation)
func IsValidPhoneNumber(phone string) error {
	pattern := `^\+?[1-9]\d{1,14}$`
	matched, err := regexp.MatchString(pattern, phone)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("invalid phone number format")
	}
	return nil
}

// IsValidDateRange validates that start date is before end date
func IsValidDateRange(startDate, endDate time.Time) error {
	if startDate.After(endDate) {
		return fmt.Errorf("start date must be before end date")
	}
	return nil
}

// IsPositive validates that a number is positive
func IsPositive(value interface{}) error {
	var numValue float64

	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case float64:
		numValue = v
	case float32:
		numValue = float64(v)
	default:
		return fmt.Errorf("value is not a number")
	}

	if numValue <= 0 {
		return fmt.Errorf("value must be positive")
	}

	return nil
}

// IsNonNegative validates that a number is non-negative
func IsNonNegative(value interface{}) error {
	var numValue float64

	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case float64:
		numValue = v
	case float32:
		numValue = float64(v)
	default:
		return fmt.Errorf("value is not a number")
	}

	if numValue < 0 {
		return fmt.Errorf("value must be non-negative")
	}

	return nil
}
