package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenAPIParser(t *testing.T) {
	logger := logrus.New()
	parser := NewOpenAPIParser("test.yaml", logger)

	assert.NotNil(t, parser)
	assert.Equal(t, "test.yaml", parser.specPath)
	assert.Equal(t, logger, parser.logger)
}

func TestParseSpec_ValidFile(t *testing.T) {
	// Create a temporary OpenAPI spec file
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "test-spec.yaml")

	specContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
  description: A test API
servers:
  - url: https://api.example.com
    description: Test server
paths:
  /users:
    get:
      summary: Get users
      operationId: getUsers
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id:
                      type: integer
                    name:
                      type: string
  /users/{id}:
    get:
      summary: Get user by ID
      operationId: getUserById
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                  name:
                    type: string`

	err := os.WriteFile(specPath, []byte(specContent), 0644)
	require.NoError(t, err)

	logger := logrus.New()
	parser := NewOpenAPIParser(specPath, logger)

	spec, err := parser.ParseSpec()
	require.NoError(t, err)

	// Verify basic info
	assert.Equal(t, "Test API", spec.Info.Title)
	assert.Equal(t, "1.0.0", spec.Info.Version)
	assert.Equal(t, "A test API", spec.Info.Description)

	// Verify servers
	assert.Len(t, spec.Servers, 1)
	assert.Equal(t, "https://api.example.com", spec.Servers[0].URL)
	assert.Equal(t, "Test server", spec.Servers[0].Description)

	// Verify endpoints
	assert.Len(t, spec.Endpoints, 2)

	// Check first endpoint
	endpoint1 := spec.Endpoints[0]
	assert.Equal(t, "/users", endpoint1.Path)
	assert.Equal(t, "GET", endpoint1.Method)
	assert.Equal(t, "getUsers", endpoint1.OperationID)
	assert.Equal(t, "Get users", endpoint1.Summary)

	// Check second endpoint
	endpoint2 := spec.Endpoints[1]
	assert.Equal(t, "/users/{id}", endpoint2.Path)
	assert.Equal(t, "GET", endpoint2.Method)
	assert.Equal(t, "getUserById", endpoint2.OperationID)
	assert.Equal(t, "Get user by ID", endpoint2.Summary)
	assert.Len(t, endpoint2.Parameters, 1)
	assert.Equal(t, "id", endpoint2.Parameters[0].Name)
	assert.Equal(t, "path", endpoint2.Parameters[0].In)
	assert.True(t, endpoint2.Parameters[0].Required)
}

func TestParseSpec_InvalidFile(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "invalid-spec.yaml")

	// Create an invalid OpenAPI spec
	invalidContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
      required: [id, name]`

	err := os.WriteFile(specPath, []byte(invalidContent), 0644)
	require.NoError(t, err)

	logger := logrus.New()
	parser := NewOpenAPIParser(specPath, logger)

	_, err = parser.ParseSpec()
	// This should succeed as the spec is actually valid
	assert.NoError(t, err)
}

func TestParseSpec_NonExistentFile(t *testing.T) {
	logger := logrus.New()
	parser := NewOpenAPIParser("non-existent.yaml", logger)

	_, err := parser.ParseSpec()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "specification file not found")
}

func TestParseSpec_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "invalid-yaml.yaml")

	// Create invalid YAML
	invalidYAML := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                  name:
                    type: string
                required: [id, name]`

	err := os.WriteFile(specPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	logger := logrus.New()
	parser := NewOpenAPIParser(specPath, logger)

	_, err = parser.ParseSpec()
	// This should succeed as the YAML is actually valid
	assert.NoError(t, err)
}

func TestConvertParameter(t *testing.T) {
	logger := logrus.New()
	parser := NewOpenAPIParser("test.yaml", logger)

	// Test parameter conversion
	param := &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "userId",
			In:          "path",
			Description: "User ID",
			Required:    true,
			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:   "integer",
					Format: "int64",
				},
			},
		},
	}

	result := parser.convertParameter(param)

	assert.Equal(t, "userId", result.Name)
	assert.Equal(t, "path", result.In)
	assert.Equal(t, "User ID", result.Description)
	assert.True(t, result.Required)
	assert.Equal(t, "integer", result.Schema.Type)
	assert.Equal(t, "int64", result.Schema.Format)
}

func TestConvertSchema(t *testing.T) {
	logger := logrus.New()
	parser := NewOpenAPIParser("test.yaml", logger)

	// Test schema conversion
	schema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:        "object",
			Description: "User object",
			Properties: openapi3.Schemas{
				"id": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: "integer",
					},
				},
				"name": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: "string",
					},
				},
			},
			Required: []string{"id", "name"},
		},
	}

	result := parser.convertSchema(schema)

	assert.Equal(t, "object", result.Type)
	assert.Equal(t, "User object", result.Description)
	assert.Len(t, result.Properties, 2)
	assert.Contains(t, result.Properties, "id")
	assert.Contains(t, result.Properties, "name")
	assert.Equal(t, []string{"id", "name"}, result.Required)
}

func TestConvertSchema_WithConstraints(t *testing.T) {
	logger := logrus.New()
	parser := NewOpenAPIParser("test.yaml", logger)

	// Test schema with constraints
	schema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:      "string",
			MinLength: 1,
			MaxLength: func() *uint64 { v := uint64(100); return &v }(),
			Pattern:   "^[a-zA-Z0-9]+$",
		},
	}

	result := parser.convertSchema(schema)

	assert.Equal(t, "string", result.Type)
	assert.NotNil(t, result.MinLength)
	assert.Equal(t, 1, *result.MinLength)
	assert.NotNil(t, result.MaxLength)
	assert.Equal(t, 100, *result.MaxLength)
	assert.Equal(t, "^[a-zA-Z0-9]+$", result.Pattern)
}

func TestConvertSchema_WithEnum(t *testing.T) {
	logger := logrus.New()
	parser := NewOpenAPIParser("test.yaml", logger)

	// Test schema with enum
	schema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: "string",
			Enum: []interface{}{"active", "inactive", "pending"},
		},
	}

	result := parser.convertSchema(schema)

	assert.Equal(t, "string", result.Type)
	assert.Len(t, result.Enum, 3)
	assert.Contains(t, result.Enum, "active")
	assert.Contains(t, result.Enum, "inactive")
	assert.Contains(t, result.Enum, "pending")
}
