package parser

import (
	"fmt"
	"strings"

	"api-to-mcp/pkg/openapi"

	"github.com/sirupsen/logrus"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// Validator validates OpenAPI specifications
type Validator struct {
	logger *logrus.Logger
}

// NewValidator creates a new validator
func NewValidator(logger *logrus.Logger) *Validator {
	return &Validator{
		logger: logger,
	}
}

// ValidateSpec validates a parsed OpenAPI specification
func (v *Validator) ValidateSpec(spec *openapi.ParsedSpec) error {
	var errors []error

	// Validate basic info
	if err := v.validateInfo(spec.Info); err != nil {
		errors = append(errors, err)
	}

	// Validate endpoints
	if err := v.validateEndpoints(spec.Endpoints); err != nil {
		errors = append(errors, err)
	}

	// Validate components
	if err := v.validateComponents(spec.Components); err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	v.logger.Info("OpenAPI specification validation passed")
	return nil
}

// validateInfo validates the API info section
func (v *Validator) validateInfo(info openapi.Info) error {
	if info.Title == "" {
		return &ValidationError{
			Field:   "info.title",
			Message: "title is required",
		}
	}

	if info.Version == "" {
		return &ValidationError{
			Field:   "info.version",
			Message: "version is required",
		}
	}

	return nil
}

// validateEndpoints validates the API endpoints
func (v *Validator) validateEndpoints(endpoints []openapi.Endpoint) error {
	if len(endpoints) == 0 {
		return &ValidationError{
			Field:   "paths",
			Message: "at least one endpoint is required",
		}
	}

	for i, endpoint := range endpoints {
		if err := v.validateEndpoint(endpoint, i); err != nil {
			return err
		}
	}

	return nil
}

// validateEndpoint validates a single endpoint
func (v *Validator) validateEndpoint(endpoint openapi.Endpoint, index int) error {
	// Validate path
	if endpoint.Path == "" {
		return &ValidationError{
			Field:   fmt.Sprintf("paths[%d].path", index),
			Message: "path is required",
		}
	}

	// Validate method
	if endpoint.Method == "" {
		return &ValidationError{
			Field:   fmt.Sprintf("paths[%d].method", index),
			Message: "method is required",
		}
	}

	// Validate method is supported
	supportedMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	if !v.isValidMethod(endpoint.Method, supportedMethods) {
		return &ValidationError{
			Field:   fmt.Sprintf("paths[%d].method", index),
			Message: fmt.Sprintf("unsupported method: %s", endpoint.Method),
		}
	}

	// Validate parameters
	for j, param := range endpoint.Parameters {
		if err := v.validateParameter(param, index, j); err != nil {
			return err
		}
	}

	// Validate responses
	if len(endpoint.Responses) == 0 {
		return &ValidationError{
			Field:   fmt.Sprintf("paths[%d].responses", index),
			Message: "at least one response is required",
		}
	}

	return nil
}

// validateParameter validates a parameter
func (v *Validator) validateParameter(param openapi.Parameter, endpointIndex, paramIndex int) error {
	if param.Name == "" {
		return &ValidationError{
			Field:   fmt.Sprintf("paths[%d].parameters[%d].name", endpointIndex, paramIndex),
			Message: "parameter name is required",
		}
	}

	// Validate parameter location
	validLocations := []string{"path", "query", "header", "cookie"}
	if !v.isValidLocation(param.In, validLocations) {
		return &ValidationError{
			Field:   fmt.Sprintf("paths[%d].parameters[%d].in", endpointIndex, paramIndex),
			Message: fmt.Sprintf("invalid parameter location: %s", param.In),
		}
	}

	// Validate schema
	if err := v.validateSchema(param.Schema, fmt.Sprintf("paths[%d].parameters[%d].schema", endpointIndex, paramIndex)); err != nil {
		return err
	}

	return nil
}

// validateSchema validates a schema
func (v *Validator) validateSchema(schema openapi.Schema, fieldPath string) error {
	if schema.Type == "" {
		return &ValidationError{
			Field:   fieldPath,
			Message: "schema type is required",
		}
	}

	// Validate type
	validTypes := []string{"string", "integer", "number", "boolean", "array", "object"}
	if !v.isValidType(schema.Type, validTypes) {
		return &ValidationError{
			Field:   fieldPath,
			Message: fmt.Sprintf("invalid schema type: %s", schema.Type),
		}
	}

	// Validate constraints
	if err := v.validateConstraints(schema, fieldPath); err != nil {
		return err
	}

	// Validate properties for object types
	if schema.Type == "object" && len(schema.Properties) > 0 {
		for propName, propSchema := range schema.Properties {
			if err := v.validateSchema(propSchema, fmt.Sprintf("%s.properties.%s", fieldPath, propName)); err != nil {
				return err
			}
		}
	}

	// Validate items for array types
	if schema.Type == "array" && schema.Items != nil {
		if err := v.validateSchema(*schema.Items, fmt.Sprintf("%s.items", fieldPath)); err != nil {
			return err
		}
	}

	return nil
}

// validateConstraints validates schema constraints
func (v *Validator) validateConstraints(schema openapi.Schema, fieldPath string) error {
	// Validate minimum/maximum for numeric types
	if schema.Type == "integer" || schema.Type == "number" {
		if schema.Minimum != nil && schema.Maximum != nil && *schema.Minimum > *schema.Maximum {
			return &ValidationError{
				Field:   fieldPath,
				Message: "minimum cannot be greater than maximum",
			}
		}
	}

	// Validate minLength/maxLength for string types
	if schema.Type == "string" {
		if schema.MinLength != nil && schema.MaxLength != nil && *schema.MinLength > *schema.MaxLength {
			return &ValidationError{
				Field:   fieldPath,
				Message: "minLength cannot be greater than maxLength",
			}
		}
	}

	return nil
}

// validateComponents validates the components section
func (v *Validator) validateComponents(components map[string]openapi.Component) error {
	for name, component := range components {
		if err := v.validateComponent(component, name); err != nil {
			return err
		}
	}

	return nil
}

// validateComponent validates a single component
func (v *Validator) validateComponent(component openapi.Component, name string) error {
	if component.Type == "" {
		return &ValidationError{
			Field:   fmt.Sprintf("components.%s.type", name),
			Message: "component type is required",
		}
	}

	// Validate component type
	validComponentTypes := []string{"schema", "response", "parameter", "example", "requestBody", "header", "securityScheme", "link", "callback"}
	if !v.isValidType(component.Type, validComponentTypes) {
		return &ValidationError{
			Field:   fmt.Sprintf("components.%s.type", name),
			Message: fmt.Sprintf("invalid component type: %s", component.Type),
		}
	}

	// Validate schema for schema components
	if component.Type == "schema" {
		if err := v.validateSchema(component.Schema, fmt.Sprintf("components.%s.schema", name)); err != nil {
			return err
		}
	}

	return nil
}

// Helper methods for validation

func (v *Validator) isValidMethod(method string, validMethods []string) bool {
	for _, validMethod := range validMethods {
		if strings.EqualFold(method, validMethod) {
			return true
		}
	}
	return false
}

func (v *Validator) isValidLocation(location string, validLocations []string) bool {
	for _, validLocation := range validLocations {
		if strings.EqualFold(location, validLocation) {
			return true
		}
	}
	return false
}

func (v *Validator) isValidType(typeStr string, validTypes []string) bool {
	for _, validType := range validTypes {
		if strings.EqualFold(typeStr, validType) {
			return true
		}
	}
	return false
}
