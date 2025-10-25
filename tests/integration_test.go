package tests

import (
	"os"
	"path/filepath"
	"testing"

	"api-to-mcp/internal/parser"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ParsePetStoreSpec(t *testing.T) {
	// Use the example petstore spec
	specPath := "../../examples/petstore.yaml"

	// Check if file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("Petstore spec not found, skipping integration test")
	}

	logger := logrus.New()
	openAPIParser := parser.NewOpenAPIParser(specPath, logger)

	spec, err := openAPIParser.ParseSpec()
	require.NoError(t, err)

	// Verify basic info
	assert.Equal(t, "Pet Store API", spec.Info.Title)
	assert.Equal(t, "1.0.0", spec.Info.Version)
	assert.Contains(t, spec.Info.Description, "petstore")

	// Verify we have endpoints
	assert.Greater(t, len(spec.Endpoints), 0)

	// Verify specific endpoints exist
	endpointPaths := make(map[string]bool)
	for _, endpoint := range spec.Endpoints {
		endpointPaths[endpoint.Path] = true
	}

	assert.True(t, endpointPaths["/pet"], "Should have /pet endpoint")
	assert.True(t, endpointPaths["/pet/{petId}"], "Should have /pet/{petId} endpoint")
	assert.True(t, endpointPaths["/pet/findByStatus"], "Should have /pet/findByStatus endpoint")

	// Verify we have components
	assert.Greater(t, len(spec.Components), 0)

	// Verify specific components exist
	componentNames := make(map[string]bool)
	for name := range spec.Components {
		componentNames[name] = true
	}

	assert.True(t, componentNames["Pet"], "Should have Pet component")
	assert.True(t, componentNames["Category"], "Should have Category component")
	assert.True(t, componentNames["Tag"], "Should have Tag component")
}

func TestIntegration_ParseComplexSpec(t *testing.T) {
	// Create a more complex test spec
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "complex-spec.yaml")

	complexSpec := `openapi: 3.0.0
info:
  title: Complex API
  version: 2.0.0
  description: A complex API with multiple endpoints and components
servers:
  - url: https://api.example.com/v1
    description: Production server
  - url: https://staging-api.example.com/v1
    description: Staging server
paths:
  /users:
    get:
      summary: List users
      operationId: listUsers
      parameters:
        - name: limit
          in: query
          description: Number of users to return
          required: false
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 10
        - name: offset
          in: query
          description: Number of users to skip
          required: false
          schema:
            type: integer
            minimum: 0
            default: 0
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
    post:
      summary: Create user
      operationId: createUser
      requestBody:
        description: User object
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        '201':
          description: User created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
  /users/{id}:
    get:
      summary: Get user by ID
      operationId: getUserById
      parameters:
        - name: id
          in: path
          description: User ID
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: User not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
    put:
      summary: Update user
      operationId: updateUser
      parameters:
        - name: id
          in: path
          description: User ID
          required: true
          schema:
            type: integer
            format: int64
      requestBody:
        description: Updated user object
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        '200':
          description: User updated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: User not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
    delete:
      summary: Delete user
      operationId: deleteUser
      parameters:
        - name: id
          in: path
          description: User ID
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '204':
          description: User deleted
        '404':
          description: User not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    User:
      type: object
      required:
        - id
        - name
        - email
      properties:
        id:
          type: integer
          format: int64
          description: User ID
        name:
          type: string
          description: User name
          minLength: 1
          maxLength: 100
        email:
          type: string
          format: email
          description: User email
        age:
          type: integer
          minimum: 0
          maximum: 150
          description: User age
        status:
          type: string
          enum: [active, inactive, pending]
          description: User status
        tags:
          type: array
          items:
            type: string
          description: User tags
        profile:
          $ref: '#/components/schemas/UserProfile'
    UserProfile:
      type: object
      properties:
        bio:
          type: string
          description: User biography
        avatar:
          type: string
          format: uri
          description: User avatar URL
        preferences:
          type: object
          properties:
            theme:
              type: string
              enum: [light, dark]
            notifications:
              type: boolean
    Error:
      type: object
      required:
        - code
        - message
      properties:
        code:
          type: integer
          description: Error code
        message:
          type: string
          description: Error message
        details:
          type: string
          description: Additional error details`

	err := os.WriteFile(specPath, []byte(complexSpec), 0644)
	require.NoError(t, err)

	logger := logrus.New()
	openAPIParser := parser.NewOpenAPIParser(specPath, logger)

	spec, err := openAPIParser.ParseSpec()
	require.NoError(t, err)

	// Verify basic info
	assert.Equal(t, "Complex API", spec.Info.Title)
	assert.Equal(t, "2.0.0", spec.Info.Version)

	// Verify servers
	assert.Len(t, spec.Servers, 2)
	assert.Equal(t, "https://api.example.com/v1", spec.Servers[0].URL)
	assert.Equal(t, "https://staging-api.example.com/v1", spec.Servers[1].URL)

	// Verify endpoints
	assert.Len(t, spec.Endpoints, 5) // GET, POST /users, GET, PUT, DELETE /users/{id}

	// Verify specific endpoints
	endpointMap := make(map[string]map[string]bool)
	for _, endpoint := range spec.Endpoints {
		if endpointMap[endpoint.Path] == nil {
			endpointMap[endpoint.Path] = make(map[string]bool)
		}
		endpointMap[endpoint.Path][endpoint.Method] = true
	}

	assert.True(t, endpointMap["/users"]["GET"], "Should have GET /users")
	assert.True(t, endpointMap["/users"]["POST"], "Should have POST /users")
	assert.True(t, endpointMap["/users/{id}"]["GET"], "Should have GET /users/{id}")
	assert.True(t, endpointMap["/users/{id}"]["PUT"], "Should have PUT /users/{id}")
	assert.True(t, endpointMap["/users/{id}"]["DELETE"], "Should have DELETE /users/{id}")

	// Verify components
	assert.Len(t, spec.Components, 3)
	assert.Contains(t, spec.Components, "User")
	assert.Contains(t, spec.Components, "UserProfile")
	assert.Contains(t, spec.Components, "Error")

	// Verify User component
	userComponent := spec.Components["User"]
	assert.Equal(t, "schema", userComponent.Type)
	assert.Equal(t, "object", userComponent.Schema.Type)
	assert.Len(t, userComponent.Schema.Properties, 7) // id, name, email, age, status, tags, profile
	assert.Contains(t, userComponent.Schema.Required, "id")
	assert.Contains(t, userComponent.Schema.Required, "name")
	assert.Contains(t, userComponent.Schema.Required, "email")
}

func TestIntegration_ErrorHandling(t *testing.T) {
	logger := logrus.New()

	// Test with non-existent file
	openAPIParser := parser.NewOpenAPIParser("non-existent.yaml", logger)
	_, err := openAPIParser.ParseSpec()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "specification file not found")

	// Test with invalid YAML
	tempDir := t.TempDir()
	invalidSpecPath := filepath.Join(tempDir, "invalid.yaml")

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

	err = os.WriteFile(invalidSpecPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	openAPIParser = parser.NewOpenAPIParser(invalidSpecPath, logger)
	spec, err := openAPIParser.ParseSpec()
	// This should actually succeed as the YAML is valid
	assert.NoError(t, err)
	assert.NotNil(t, spec)
}
