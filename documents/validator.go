// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package documents

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mdaxf/iac/logger"
)

// ValidationError represents a document validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no validation errors"
	}

	messages := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

func (e *ValidationErrors) Add(field, message string, value interface{}) {
	e.Errors = append(e.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// Schema represents a document validation schema
type Schema struct {
	Type       string                 // "object", "string", "number", "boolean", "array"
	Properties map[string]*Schema     // For object types
	Items      *Schema                // For array types
	Required   []string               // Required field names
	Enum       []interface{}          // Allowed values
	Pattern    string                 // Regex pattern for strings
	MinLength  *int                   // Minimum string length
	MaxLength  *int                   // Maximum string length
	Minimum    *float64               // Minimum number value
	Maximum    *float64               // Maximum number value
	Format     string                 // email, date-time, uri, etc.
	Default    interface{}            // Default value
	Custom     func(interface{}) error // Custom validation function
}

// Validator validates documents against schemas
type Validator struct {
	schemas map[string]*Schema
	mu      sync.RWMutex
	iLog    logger.Log
}

// NewValidator creates a new document validator
func NewValidator() *Validator {
	return &Validator{
		schemas: make(map[string]*Schema),
		iLog:    logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DocumentValidator"},
	}
}

// RegisterSchema registers a schema for a collection
func (v *Validator) RegisterSchema(collection string, schema *Schema) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.schemas[collection] = schema
	v.iLog.Info(fmt.Sprintf("Registered validation schema for collection: %s", collection))
}

// GetSchema retrieves a schema for a collection
func (v *Validator) GetSchema(collection string) (*Schema, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	schema, exists := v.schemas[collection]
	return schema, exists
}

// Validate validates a document against a schema
func (v *Validator) Validate(collection string, document interface{}) error {
	schema, exists := v.GetSchema(collection)
	if !exists {
		// No schema registered, skip validation
		return nil
	}

	errors := &ValidationErrors{}
	v.validateValue("", document, schema, errors)

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// validateValue validates a value against a schema
func (v *Validator) validateValue(path string, value interface{}, schema *Schema, errors *ValidationErrors) {
	if value == nil {
		if schema.Required != nil && len(schema.Required) > 0 {
			errors.Add(path, "value is required but null", value)
		}
		return
	}

	// Check type
	if !v.checkType(value, schema.Type) {
		errors.Add(path, fmt.Sprintf("expected type %s", schema.Type), value)
		return
	}

	// Validate based on type
	switch schema.Type {
	case "object":
		v.validateObject(path, value, schema, errors)
	case "array":
		v.validateArray(path, value, schema, errors)
	case "string":
		v.validateString(path, value, schema, errors)
	case "number", "integer":
		v.validateNumber(path, value, schema, errors)
	case "boolean":
		v.validateBoolean(path, value, schema, errors)
	}

	// Check enum
	if schema.Enum != nil && len(schema.Enum) > 0 {
		if !v.inEnum(value, schema.Enum) {
			errors.Add(path, fmt.Sprintf("value must be one of %v", schema.Enum), value)
		}
	}

	// Custom validation
	if schema.Custom != nil {
		if err := schema.Custom(value); err != nil {
			errors.Add(path, err.Error(), value)
		}
	}
}

// validateObject validates an object
func (v *Validator) validateObject(path string, value interface{}, schema *Schema, errors *ValidationErrors) {
	objMap, ok := value.(map[string]interface{})
	if !ok {
		errors.Add(path, "expected object/map", value)
		return
	}

	// Check required fields
	for _, requiredField := range schema.Required {
		if _, exists := objMap[requiredField]; !exists {
			errors.Add(v.joinPath(path, requiredField), "required field missing", nil)
		}
	}

	// Validate properties
	if schema.Properties != nil {
		for fieldName, fieldValue := range objMap {
			if fieldSchema, exists := schema.Properties[fieldName]; exists {
				v.validateValue(v.joinPath(path, fieldName), fieldValue, fieldSchema, errors)
			}
			// Unknown fields are allowed unless explicitly forbidden
		}
	}
}

// validateArray validates an array
func (v *Validator) validateArray(path string, value interface{}, schema *Schema, errors *ValidationErrors) {
	arr := reflect.ValueOf(value)
	if arr.Kind() != reflect.Slice && arr.Kind() != reflect.Array {
		errors.Add(path, "expected array", value)
		return
	}

	// Validate items
	if schema.Items != nil {
		for i := 0; i < arr.Len(); i++ {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			v.validateValue(itemPath, arr.Index(i).Interface(), schema.Items, errors)
		}
	}
}

// validateString validates a string
func (v *Validator) validateString(path string, value interface{}, schema *Schema, errors *ValidationErrors) {
	str, ok := value.(string)
	if !ok {
		errors.Add(path, "expected string", value)
		return
	}

	// Check length
	if schema.MinLength != nil && len(str) < *schema.MinLength {
		errors.Add(path, fmt.Sprintf("string length must be at least %d", *schema.MinLength), value)
	}

	if schema.MaxLength != nil && len(str) > *schema.MaxLength {
		errors.Add(path, fmt.Sprintf("string length must be at most %d", *schema.MaxLength), value)
	}

	// Check pattern
	if schema.Pattern != "" {
		matched, err := regexp.MatchString(schema.Pattern, str)
		if err != nil {
			errors.Add(path, fmt.Sprintf("invalid regex pattern: %s", err), value)
		} else if !matched {
			errors.Add(path, fmt.Sprintf("string does not match pattern: %s", schema.Pattern), value)
		}
	}

	// Check format
	if schema.Format != "" {
		if !v.validateFormat(str, schema.Format) {
			errors.Add(path, fmt.Sprintf("invalid format: %s", schema.Format), value)
		}
	}
}

// validateNumber validates a number
func (v *Validator) validateNumber(path string, value interface{}, schema *Schema, errors *ValidationErrors) {
	var num float64

	switch v := value.(type) {
	case int:
		num = float64(v)
	case int32:
		num = float64(v)
	case int64:
		num = float64(v)
	case float32:
		num = float64(v)
	case float64:
		num = v
	default:
		errors.Add(path, "expected number", value)
		return
	}

	// Check minimum
	if schema.Minimum != nil && num < *schema.Minimum {
		errors.Add(path, fmt.Sprintf("number must be at least %f", *schema.Minimum), value)
	}

	// Check maximum
	if schema.Maximum != nil && num > *schema.Maximum {
		errors.Add(path, fmt.Sprintf("number must be at most %f", *schema.Maximum), value)
	}
}

// validateBoolean validates a boolean
func (v *Validator) validateBoolean(path string, value interface{}, schema *Schema, errors *ValidationErrors) {
	if _, ok := value.(bool); !ok {
		errors.Add(path, "expected boolean", value)
	}
}

// checkType checks if a value matches a type
func (v *Validator) checkType(value interface{}, schemaType string) bool {
	if schemaType == "" {
		return true // No type constraint
	}

	switch schemaType {
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	case "array":
		rt := reflect.TypeOf(value)
		return rt != nil && (rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array)
	case "string":
		_, ok := value.(string)
		return ok
	case "number", "integer":
		switch value.(type) {
		case int, int32, int64, float32, float64:
			return true
		}
		return false
	case "boolean":
		_, ok := value.(bool)
		return ok
	default:
		return true
	}
}

// validateFormat validates string formats
func (v *Validator) validateFormat(value, format string) bool {
	switch format {
	case "email":
		emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		matched, _ := regexp.MatchString(emailRegex, value)
		return matched

	case "uri", "url":
		urlRegex := `^https?://[^\s/$.?#].[^\s]*$`
		matched, _ := regexp.MatchString(urlRegex, value)
		return matched

	case "date":
		_, err := time.Parse("2006-01-02", value)
		return err == nil

	case "date-time":
		_, err := time.Parse(time.RFC3339, value)
		return err == nil

	case "uuid":
		uuidRegex := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
		matched, _ := regexp.MatchString(uuidRegex, value)
		return matched

	case "ipv4":
		ipv4Regex := `^(\d{1,3}\.){3}\d{1,3}$`
		matched, _ := regexp.MatchString(ipv4Regex, value)
		return matched

	case "ipv6":
		ipv6Regex := `^([0-9a-f]{1,4}:){7}[0-9a-f]{1,4}$`
		matched, _ := regexp.MatchString(ipv6Regex, value)
		return matched

	default:
		return true // Unknown format, skip validation
	}
}

// inEnum checks if a value is in an enum
func (v *Validator) inEnum(value interface{}, enum []interface{}) bool {
	for _, e := range enum {
		if reflect.DeepEqual(value, e) {
			return true
		}
	}
	return false
}

// joinPath joins path segments
func (v *Validator) joinPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

// SchemaBuilder helps build schemas fluently
type SchemaBuilder struct {
	schema *Schema
}

// NewSchemaBuilder creates a new schema builder
func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		schema: &Schema{
			Properties: make(map[string]*Schema),
			Required:   make([]string, 0),
		},
	}
}

// Type sets the schema type
func (sb *SchemaBuilder) Type(t string) *SchemaBuilder {
	sb.schema.Type = t
	return sb
}

// Property adds a property schema
func (sb *SchemaBuilder) Property(name string, propSchema *Schema) *SchemaBuilder {
	sb.schema.Properties[name] = propSchema
	return sb
}

// Required marks fields as required
func (sb *SchemaBuilder) Required(fields ...string) *SchemaBuilder {
	sb.schema.Required = append(sb.schema.Required, fields...)
	return sb
}

// Build returns the schema
func (sb *SchemaBuilder) Build() *Schema {
	return sb.schema
}

// Common schema helpers

// StringSchema creates a string schema
func StringSchema() *Schema {
	return &Schema{Type: "string"}
}

// NumberSchema creates a number schema
func NumberSchema() *Schema {
	return &Schema{Type: "number"}
}

// IntegerSchema creates an integer schema
func IntegerSchema() *Schema {
	return &Schema{Type: "integer"}
}

// BooleanSchema creates a boolean schema
func BooleanSchema() *Schema {
	return &Schema{Type: "boolean"}
}

// ArraySchema creates an array schema
func ArraySchema(items *Schema) *Schema {
	return &Schema{
		Type:  "array",
		Items: items,
	}
}

// ObjectSchema creates an object schema
func ObjectSchema(properties map[string]*Schema, required []string) *Schema {
	return &Schema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
}

// EmailSchema creates an email validation schema
func EmailSchema() *Schema {
	return &Schema{
		Type:   "string",
		Format: "email",
	}
}

// URLSchema creates a URL validation schema
func URLSchema() *Schema {
	return &Schema{
		Type:   "string",
		Format: "url",
	}
}

// DateTimeSchema creates a date-time validation schema
func DateTimeSchema() *Schema {
	return &Schema{
		Type:   "string",
		Format: "date-time",
	}
}

// UUIDSchema creates a UUID validation schema
func UUIDSchema() *Schema {
	return &Schema{
		Type:   "string",
		Format: "uuid",
	}
}

// EnumSchema creates an enum schema
func EnumSchema(values ...interface{}) *Schema {
	return &Schema{
		Enum: values,
	}
}

// Example usage
func ExampleValidatorUsage() {
	validator := NewValidator()

	// Define a user schema
	userSchema := NewSchemaBuilder().
		Type("object").
		Property("name", &Schema{
			Type:      "string",
			MinLength: intPtr(1),
			MaxLength: intPtr(100),
		}).
		Property("email", EmailSchema()).
		Property("age", &Schema{
			Type:    "integer",
			Minimum: float64Ptr(0),
			Maximum: float64Ptr(150),
		}).
		Property("status", EnumSchema("active", "inactive", "pending")).
		Property("tags", ArraySchema(StringSchema())).
		Required("name", "email").
		Build()

	validator.RegisterSchema("users", userSchema)

	// Validate a document
	user := map[string]interface{}{
		"name":   "John Doe",
		"email":  "john@example.com",
		"age":    30,
		"status": "active",
		"tags":   []interface{}{"admin", "user"},
	}

	err := validator.Validate("users", user)
	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
	} else {
		fmt.Println("Document is valid")
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

// ValidatedDocumentDB wraps a DocumentDB with validation
type ValidatedDocumentDB struct {
	DocumentDB
	validator *Validator
	iLog      logger.Log
}

// NewValidatedDocumentDB creates a validated document database wrapper
func NewValidatedDocumentDB(db DocumentDB, validator *Validator) *ValidatedDocumentDB {
	return &ValidatedDocumentDB{
		DocumentDB: db,
		validator:  validator,
		iLog:       logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "ValidatedDocumentDB"},
	}
}

// InsertOne validates before inserting
func (vdb *ValidatedDocumentDB) InsertOne(ctx interface{}, collection string, document interface{}) (string, error) {
	if err := vdb.validator.Validate(collection, document); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	return vdb.DocumentDB.InsertOne(ctx.(interface{ context.Context }), collection, document)
}

// UpdateOne validates update operations
func (vdb *ValidatedDocumentDB) UpdateOne(ctx interface{}, collection string, filter interface{}, update interface{}) error {
	// For updates, we might want to validate the update document
	// This is simplified - in practice, you'd validate the final state
	return vdb.DocumentDB.UpdateOne(ctx.(interface{ context.Context }), collection, filter, update)
}

// Global validator instance
var globalValidator *Validator
var validatorOnce sync.Once

// GetGlobalValidator returns the global validator
func GetGlobalValidator() *Validator {
	validatorOnce.Do(func() {
		globalValidator = NewValidator()
	})
	return globalValidator
}
