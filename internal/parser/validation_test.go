package parser

import (
	"testing"

	"api-to-mcp/pkg/openapi"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	logger := logrus.New()
	validator := NewValidator(logger)

	assert.NotNil(t, validator)
	assert.Equal(t, logger, validator.logger)
}

func TestValidateSpec_ValidSpec(t *testing.T) {
	logger := logrus.New()
	validator := NewValidator(logger)

	spec := &openapi.ParsedSpec{
		Info: openapi.Info{
			Title:       "Test API",
			Version:     "1.0.0",
			Description: "A test API",
		},
		Endpoints: []openapi.Endpoint{
			{
				Path:   "/users",
				Method: "GET",
				Responses: map[string]openapi.Response{
					"200": {
						Description: "Success",
					},
				},
			},
		},
		Components: make(map[string]openapi.Component),
	}

	err := validator.ValidateSpec(spec)
	assert.NoError(t, err)
}

func TestValidateSpec_InvalidInfo(t *testing.T) {
	logger := logrus.New()
	validator := NewValidator(logger)

	spec := &openapi.ParsedSpec{
		Info: openapi.Info{
			Title:   "", // Missing title
			Version: "1.0.0",
		},
		Endpoints: []openapi.Endpoint{
			{
				Path:   "/users",
				Method: "GET",
				Responses: map[string]openapi.Response{
					"200": {
						Description: "Success",
					},
				},
			},
		},
		Components: make(map[string]openapi.Component),
	}

	err := validator.ValidateSpec(spec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title is required")
}

func TestValidateSpec_NoEndpoints(t *testing.T) {
	logger := logrus.New()
	validator := NewValidator(logger)

	spec := &openapi.ParsedSpec{
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Endpoints:  []openapi.Endpoint{}, // No endpoints
		Components: make(map[string]openapi.Component),
	}

	err := validator.ValidateSpec(spec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one endpoint is required")
}

func TestValidateEndpoint_InvalidMethod(t *testing.T) {
	logger := logrus.New()
	validator := NewValidator(logger)

	endpoints := []openapi.Endpoint{
		{
			Path:   "/users",
			Method: "INVALID", // Invalid method
			Responses: map[string]openapi.Response{
				"200": {
					Description: "Success",
				},
			},
		},
	}

	err := validator.validateEndpoints(endpoints)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported method")
}

func TestValidateParameter_InvalidLocation(t *testing.T) {
	logger := logrus.New()
	validator := NewValidator(logger)

	endpoint := openapi.Endpoint{
		Path:   "/users",
		Method: "GET",
		Parameters: []openapi.Parameter{
			{
				Name: "id",
				In:   "invalid", // Invalid location
				Schema: openapi.Schema{
					Type: "integer",
				},
			},
		},
		Responses: map[string]openapi.Response{
			"200": {
				Description: "Success",
			},
		},
	}

	err := validator.validateEndpoint(endpoint, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parameter location")
}

func TestValidateSchema_InvalidType(t *testing.T) {
	logger := logrus.New()
	validator := NewValidator(logger)

	schema := openapi.Schema{
		Type: "invalid", // Invalid type
	}

	err := validator.validateSchema(schema, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid schema type")
}

func TestValidateSchema_InvalidConstraints(t *testing.T) {
	logger := logrus.New()
	validator := NewValidator(logger)

	schema := openapi.Schema{
		Type:    "integer",
		Minimum: func() *float64 { v := 10.0; return &v }(),
		Maximum: func() *float64 { v := 5.0; return &v }(), // Maximum < Minimum
	}

	err := validator.validateSchema(schema, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minimum cannot be greater than maximum")
}

func TestValidateSchema_StringConstraints(t *testing.T) {
	logger := logrus.New()
	validator := NewValidator(logger)

	schema := openapi.Schema{
		Type:      "string",
		MinLength: func() *int { v := 10; return &v }(),
		MaxLength: func() *int { v := 5; return &v }(), // MaxLength < MinLength
	}

	err := validator.validateSchema(schema, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minLength cannot be greater than maxLength")
}

func TestValidateComponent_InvalidType(t *testing.T) {
	logger := logrus.New()
	validator := NewValidator(logger)

	components := map[string]openapi.Component{
		"User": {
			Type: "invalid", // Invalid component type
		},
	}

	err := validator.validateComponents(components)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid component type")
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "test.field",
		Message: "test message",
	}

	expected := "validation error in field 'test.field': test message"
	assert.Equal(t, expected, err.Error())
}
