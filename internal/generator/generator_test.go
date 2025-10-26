package generator

import (
	"os"
	"testing"

	"api-to-mcp/internal/config"
	"api-to-mcp/internal/parser"
	"api-to-mcp/pkg/mcp"
	"api-to-mcp/pkg/openapi"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMCPToolGenerator(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}

	generator := NewMCPToolGenerator(spec, config, logger)

	assert.NotNil(t, generator)
	assert.Equal(t, spec, generator.spec)
	assert.Equal(t, config, generator.config)
	assert.Equal(t, logger, generator.logger)
}

func TestGenerateTools_SimpleSpec(t *testing.T) {
	// Create a simple test spec
	spec := &openapi.ParsedSpec{
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Endpoints: []openapi.Endpoint{
			{
				Path:        "/users",
				Method:      "GET",
				OperationID: "getUsers",
				Summary:     "Get all users",
				Parameters:  []openapi.Parameter{},
				Responses:   make(map[string]openapi.Response),
			},
			{
				Path:        "/users/{id}",
				Method:      "GET",
				OperationID: "getUserById",
				Summary:     "Get user by ID",
				Parameters: []openapi.Parameter{
					{
						Name:        "id",
						In:          "path",
						Description: "User ID",
						Required:    true,
						Schema: openapi.Schema{
							Type: "integer",
						},
					},
				},
				Responses: make(map[string]openapi.Response),
			},
		},
	}

	config := &config.Config{
		OpenAPI: config.OpenAPIConfig{
			BaseURL: "https://api.example.com",
		},
		Filters: config.FilterConfig{},
	}

	logger := logrus.New()
	generator := NewMCPToolGenerator(spec, config, logger)

	tools, err := generator.GenerateTools()
	require.NoError(t, err)
	assert.Len(t, tools, 2)

	// Check first tool
	tool1 := tools[0]
	assert.Equal(t, "getusers", tool1.Name)
	assert.Equal(t, "Get all users", tool1.Description)
	assert.NotNil(t, tool1.InputSchema)
	assert.NotNil(t, tool1.Handler)

	// Check second tool
	tool2 := tools[1]
	assert.Equal(t, "getuserbyid", tool2.Name)
	assert.Equal(t, "Get user by ID", tool2.Description)
	assert.NotNil(t, tool2.InputSchema)
	assert.NotNil(t, tool2.Handler)
	assert.Contains(t, tool2.InputSchema.Properties, "id")
	assert.Contains(t, tool2.InputSchema.Required, "id")
}

func TestGenerateTools_WithQueryParameters(t *testing.T) {
	spec := &openapi.ParsedSpec{
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Endpoints: []openapi.Endpoint{
			{
				Path:        "/users",
				Method:      "GET",
				OperationID: "searchUsers",
				Summary:     "Search users",
				Parameters: []openapi.Parameter{
					{
						Name:        "limit",
						In:          "query",
						Description: "Number of users to return",
						Required:    false,
						Schema: openapi.Schema{
							Type:    "integer",
							Minimum: func() *float64 { v := float64(1); return &v }(),
							Maximum: func() *float64 { v := float64(100); return &v }(),
							Default: 10,
						},
					},
					{
						Name:        "offset",
						In:          "query",
						Description: "Number of users to skip",
						Required:    false,
						Schema: openapi.Schema{
							Type:    "integer",
							Minimum: func() *float64 { v := float64(0); return &v }(),
							Default: 0,
						},
					},
					{
						Name:        "status",
						In:          "query",
						Description: "User status",
						Required:    true,
						Schema: openapi.Schema{
							Type: "string",
							Enum: []interface{}{"active", "inactive", "pending"},
						},
					},
				},
				Responses: make(map[string]openapi.Response),
			},
		},
	}

	config := &config.Config{
		OpenAPI: config.OpenAPIConfig{
			BaseURL: "https://api.example.com",
		},
		Filters: config.FilterConfig{},
	}

	logger := logrus.New()
	generator := NewMCPToolGenerator(spec, config, logger)

	tools, err := generator.GenerateTools()
	require.NoError(t, err)
	assert.Len(t, tools, 1)

	tool := tools[0]
	assert.Equal(t, "searchusers", tool.Name)
	assert.Equal(t, "Search users", tool.Description)

	// Check input schema
	schema := tool.InputSchema
	assert.Equal(t, "object", schema.Type)
	assert.Len(t, schema.Properties, 3)
	assert.Len(t, schema.Required, 1)

	// Check limit parameter
	limitProp := schema.Properties["limit"]
	assert.Equal(t, "integer", limitProp.Type)
	assert.Equal(t, "Number of users to return", limitProp.Description)
	assert.NotNil(t, limitProp.Minimum)
	assert.Equal(t, float64(1), *limitProp.Minimum)
	assert.NotNil(t, limitProp.Maximum)
	assert.Equal(t, float64(100), *limitProp.Maximum)
	assert.Equal(t, 10, limitProp.Default)

	// Check offset parameter
	offsetProp := schema.Properties["offset"]
	assert.Equal(t, "integer", offsetProp.Type)
	assert.Equal(t, "Number of users to skip", offsetProp.Description)
	assert.NotNil(t, offsetProp.Minimum)
	assert.Equal(t, float64(0), *offsetProp.Minimum)
	assert.Equal(t, 0, offsetProp.Default)

	// Check status parameter
	statusProp := schema.Properties["status"]
	assert.Equal(t, "string", statusProp.Type)
	assert.Equal(t, "User status", statusProp.Description)
	assert.Len(t, statusProp.Enum, 3)
	assert.Contains(t, statusProp.Enum, "active")
	assert.Contains(t, statusProp.Enum, "inactive")
	assert.Contains(t, statusProp.Enum, "pending")

	// Check required fields
	assert.Contains(t, schema.Required, "status")
	assert.NotContains(t, schema.Required, "limit")
	assert.NotContains(t, schema.Required, "offset")
}

func TestGenerateTools_WithRequestBody(t *testing.T) {
	spec := &openapi.ParsedSpec{
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Endpoints: []openapi.Endpoint{
			{
				Path:        "/users",
				Method:      "POST",
				OperationID: "createUser",
				Summary:     "Create a new user",
				Parameters:  []openapi.Parameter{},
				RequestBody: &openapi.RequestBody{
					Description: "User object",
					Required:    true,
					Content: map[string]openapi.MediaType{
						"application/json": {
							Schema: openapi.Schema{
								Type: "object",
								Properties: map[string]openapi.Schema{
									"name": {
										Type:        "string",
										Description: "User name",
									},
									"email": {
										Type:        "string",
										Format:      "email",
										Description: "User email",
									},
								},
								Required: []string{"name", "email"},
							},
						},
					},
				},
				Responses: make(map[string]openapi.Response),
			},
		},
	}

	config := &config.Config{
		OpenAPI: config.OpenAPIConfig{
			BaseURL: "https://api.example.com",
		},
		Filters: config.FilterConfig{},
	}

	logger := logrus.New()
	generator := NewMCPToolGenerator(spec, config, logger)

	tools, err := generator.GenerateTools()
	require.NoError(t, err)
	assert.Len(t, tools, 1)

	tool := tools[0]
	assert.Equal(t, "createuser", tool.Name)
	assert.Equal(t, "Create a new user", tool.Description)

	// Check input schema - should now have parsed request body properties
	schema := tool.InputSchema
	assert.Equal(t, "object", schema.Type)

	// Should have the request body properties directly in the schema
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "email")
	assert.Contains(t, schema.Required, "name")
	assert.Contains(t, schema.Required, "email")

	// Check name property
	nameProp := schema.Properties["name"]
	assert.Equal(t, "string", nameProp.Type)
	assert.Equal(t, "User name", nameProp.Description)

	// Check email property
	emailProp := schema.Properties["email"]
	assert.Equal(t, "string", emailProp.Type)
	assert.Equal(t, "email", emailProp.Format)
	assert.Equal(t, "User email", emailProp.Description)
}

func TestGenerateTools_WithFilters(t *testing.T) {
	spec := &openapi.ParsedSpec{
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Endpoints: []openapi.Endpoint{
			{
				Path:        "/users",
				Method:      "GET",
				OperationID: "getUsers",
				Summary:     "Get all users",
				Parameters:  []openapi.Parameter{},
				Responses:   make(map[string]openapi.Response),
			},
			{
				Path:        "/admin/users",
				Method:      "GET",
				OperationID: "getAdminUsers",
				Summary:     "Get admin users",
				Parameters:  []openapi.Parameter{},
				Responses:   make(map[string]openapi.Response),
			},
			{
				Path:        "/posts",
				Method:      "GET",
				OperationID: "getPosts",
				Summary:     "Get all posts",
				Parameters:  []openapi.Parameter{},
				Responses:   make(map[string]openapi.Response),
			},
		},
	}

	config := &config.Config{
		OpenAPI: config.OpenAPIConfig{
			BaseURL: "https://api.example.com",
		},
		Filters: config.FilterConfig{
			IncludePaths: []string{"/users"},
			ExcludePaths: []string{"/admin"},
		},
	}

	logger := logrus.New()
	generator := NewMCPToolGenerator(spec, config, logger)

	tools, err := generator.GenerateTools()
	require.NoError(t, err)
	assert.Len(t, tools, 1) // Only /users should be included

	tool := tools[0]
	assert.Equal(t, "getusers", tool.Name)
}

func TestGenerateTools_WithMethodFilters(t *testing.T) {
	spec := &openapi.ParsedSpec{
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Endpoints: []openapi.Endpoint{
			{
				Path:        "/users",
				Method:      "GET",
				OperationID: "getUsers",
				Summary:     "Get all users",
				Parameters:  []openapi.Parameter{},
				Responses:   make(map[string]openapi.Response),
			},
			{
				Path:        "/users",
				Method:      "POST",
				OperationID: "createUser",
				Summary:     "Create user",
				Parameters:  []openapi.Parameter{},
				Responses:   make(map[string]openapi.Response),
			},
			{
				Path:        "/users",
				Method:      "DELETE",
				OperationID: "deleteUser",
				Summary:     "Delete user",
				Parameters:  []openapi.Parameter{},
				Responses:   make(map[string]openapi.Response),
			},
		},
	}

	config := &config.Config{
		OpenAPI: config.OpenAPIConfig{
			BaseURL: "https://api.example.com",
		},
		Filters: config.FilterConfig{
			IncludeMethods: []string{"GET", "POST"},
		},
	}

	logger := logrus.New()
	generator := NewMCPToolGenerator(spec, config, logger)

	tools, err := generator.GenerateTools()
	require.NoError(t, err)
	assert.Len(t, tools, 2) // Only GET and POST should be included

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	assert.True(t, toolNames["getusers"])
	assert.True(t, toolNames["createuser"])
	assert.False(t, toolNames["deleteuser"])
}

func TestGenerateTools_WithComplexRequestBody(t *testing.T) {
	spec := &openapi.ParsedSpec{
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Endpoints: []openapi.Endpoint{
			{
				Path:        "/users",
				Method:      "POST",
				OperationID: "createUser",
				Summary:     "Create a new user",
				Parameters:  []openapi.Parameter{},
				RequestBody: &openapi.RequestBody{
					Description: "User object with complex schema",
					Required:    true,
					Content: map[string]openapi.MediaType{
						"application/json": {
							Schema: openapi.Schema{
								Type: "object",
								Properties: map[string]openapi.Schema{
									"name": {
										Type:        "string",
										Description: "User name",
										MinLength:   func() *int { v := 1; return &v }(),
										MaxLength:   func() *int { v := 100; return &v }(),
									},
									"email": {
										Type:        "string",
										Format:      "email",
										Description: "User email",
										Pattern:     "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
									},
									"age": {
										Type:        "integer",
										Description: "User age",
										Minimum:     func() *float64 { v := float64(0); return &v }(),
										Maximum:     func() *float64 { v := float64(150); return &v }(),
									},
									"status": {
										Type:        "string",
										Description: "User status",
										Enum:        []interface{}{"active", "inactive", "pending"},
										Default:     "pending",
									},
									"tags": {
										Type:        "array",
										Description: "User tags",
										Items: &openapi.Schema{
											Type: "string",
										},
									},
								},
								Required: []string{"name", "email"},
							},
						},
					},
				},
				Responses: make(map[string]openapi.Response),
			},
		},
	}

	config := &config.Config{
		OpenAPI: config.OpenAPIConfig{
			BaseURL: "https://api.example.com",
		},
		Filters: config.FilterConfig{},
	}

	logger := logrus.New()
	generator := NewMCPToolGenerator(spec, config, logger)

	tools, err := generator.GenerateTools()
	require.NoError(t, err)
	assert.Len(t, tools, 1)

	tool := tools[0]
	schema := tool.InputSchema

	// Check all properties are present
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "email")
	assert.Contains(t, schema.Properties, "age")
	assert.Contains(t, schema.Properties, "status")
	assert.Contains(t, schema.Properties, "tags")

	// Check required fields
	assert.Contains(t, schema.Required, "name")
	assert.Contains(t, schema.Required, "email")
	assert.NotContains(t, schema.Required, "age")
	assert.NotContains(t, schema.Required, "status")
	assert.NotContains(t, schema.Required, "tags")

	// Check name property constraints
	nameProp := schema.Properties["name"]
	assert.Equal(t, "string", nameProp.Type)
	assert.NotNil(t, nameProp.MinLength)
	assert.Equal(t, 1, *nameProp.MinLength)
	assert.NotNil(t, nameProp.MaxLength)
	assert.Equal(t, 100, *nameProp.MaxLength)

	// Check email property constraints
	emailProp := schema.Properties["email"]
	assert.Equal(t, "string", emailProp.Type)
	assert.Equal(t, "email", emailProp.Format)
	assert.Equal(t, "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", emailProp.Pattern)

	// Check age property constraints
	ageProp := schema.Properties["age"]
	assert.Equal(t, "integer", ageProp.Type)
	assert.NotNil(t, ageProp.Minimum)
	assert.Equal(t, float64(0), *ageProp.Minimum)
	assert.NotNil(t, ageProp.Maximum)
	assert.Equal(t, float64(150), *ageProp.Maximum)

	// Check status property constraints
	statusProp := schema.Properties["status"]
	assert.Equal(t, "string", statusProp.Type)
	assert.Len(t, statusProp.Enum, 3)
	assert.Contains(t, statusProp.Enum, "active")
	assert.Contains(t, statusProp.Enum, "inactive")
	assert.Contains(t, statusProp.Enum, "pending")
	assert.Equal(t, "pending", statusProp.Default)

	// Check tags property (array)
	tagsProp := schema.Properties["tags"]
	assert.Equal(t, "array", tagsProp.Type)
	assert.Contains(t, tagsProp.Description, "array of string")
}

func TestParseRequestBodySchema(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	// Test with valid request body
	requestBody := &openapi.RequestBody{
		Description: "Test request body",
		Required:    true,
		Content: map[string]openapi.MediaType{
			"application/json": {
				Schema: openapi.Schema{
					Type: "object",
					Properties: map[string]openapi.Schema{
						"test": {
							Type:        "string",
							Description: "Test field",
						},
					},
					Required: []string{"test"},
				},
			},
		},
	}

	schema, err := generator.parseRequestBodySchema(requestBody)
	require.NoError(t, err)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "test")
	assert.Contains(t, schema.Required, "test")

	// Test with nil request body
	_, err = generator.parseRequestBodySchema(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request body is nil")

	// Test with unsupported content type
	unsupportedBody := &openapi.RequestBody{
		Content: map[string]openapi.MediaType{
			"text/plain": {
				Schema: openapi.Schema{Type: "string"},
			},
		},
	}
	_, err = generator.parseRequestBodySchema(unsupportedBody)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no supported content type found")
}

func TestConvertSchemaToProperty(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	// Test with simple string schema
	schema := openapi.Schema{
		Type:        "string",
		Description: "Test string",
		Format:      "email",
		Pattern:     "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
		MinLength:   func() *int { v := 5; return &v }(),
		MaxLength:   func() *int { v := 100; return &v }(),
		Enum:        []interface{}{"option1", "option2"},
		Default:     "default",
	}

	property, err := generator.convertSchemaToProperty(schema)
	require.NoError(t, err)

	assert.Equal(t, "string", property.Type)
	assert.Equal(t, "Test string", property.Description)
	assert.Equal(t, "email", property.Format)
	assert.Equal(t, "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", property.Pattern)
	assert.NotNil(t, property.MinLength)
	assert.Equal(t, 5, *property.MinLength)
	assert.NotNil(t, property.MaxLength)
	assert.Equal(t, 100, *property.MaxLength)
	assert.Len(t, property.Enum, 2)
	assert.Contains(t, property.Enum, "option1")
	assert.Contains(t, property.Enum, "option2")
	assert.Equal(t, "default", property.Default)

	// Test with array schema
	arraySchema := openapi.Schema{
		Type:        "array",
		Description: "Test array",
		Items: &openapi.Schema{
			Type: "string",
		},
	}

	arrayProperty, err := generator.convertSchemaToProperty(arraySchema)
	require.NoError(t, err)

	assert.Equal(t, "array", arrayProperty.Type)
	assert.Contains(t, arrayProperty.Description, "array of string")
}

func TestGenerateTools_WithNestedObjectSchema(t *testing.T) {
	spec := &openapi.ParsedSpec{
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Endpoints: []openapi.Endpoint{
			{
				Path:        "/users",
				Method:      "POST",
				OperationID: "createUser",
				Summary:     "Create a new user with nested profile",
				Parameters:  []openapi.Parameter{},
				RequestBody: &openapi.RequestBody{
					Description: "User object with nested profile",
					Required:    true,
					Content: map[string]openapi.MediaType{
						"application/json": {
							Schema: openapi.Schema{
								Type: "object",
								Properties: map[string]openapi.Schema{
									"name": {
										Type:        "string",
										Description: "User name",
									},
									"email": {
										Type:        "string",
										Format:      "email",
										Description: "User email",
									},
									"profile": {
										Type:        "object",
										Description: "User profile information",
										Properties: map[string]openapi.Schema{
											"bio": {
												Type:        "string",
												Description: "User biography",
												MaxLength:   func() *int { v := 500; return &v }(),
											},
											"avatar": {
												Type:        "string",
												Format:      "uri",
												Description: "User avatar URL",
											},
											"preferences": {
												Type:        "object",
												Description: "User preferences",
												Properties: map[string]openapi.Schema{
													"theme": {
														Type:        "string",
														Description: "UI theme",
														Enum:        []interface{}{"light", "dark"},
														Default:     "light",
													},
													"notifications": {
														Type:        "boolean",
														Description: "Enable notifications",
														Default:     true,
													},
												},
												Required: []string{"theme"},
											},
										},
										Required: []string{"bio"},
									},
									"tags": {
										Type:        "array",
										Description: "User tags",
										Items: &openapi.Schema{
											Type: "string",
										},
									},
								},
								Required: []string{"name", "email", "profile"},
							},
						},
					},
				},
				Responses: make(map[string]openapi.Response),
			},
		},
	}

	config := &config.Config{
		OpenAPI: config.OpenAPIConfig{
			BaseURL: "https://api.example.com",
		},
		Filters: config.FilterConfig{},
	}

	logger := logrus.New()
	generator := NewMCPToolGenerator(spec, config, logger)

	tools, err := generator.GenerateTools()
	require.NoError(t, err)
	assert.Len(t, tools, 1)

	tool := tools[0]
	schema := tool.InputSchema

	// Check top-level properties
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "email")
	assert.Contains(t, schema.Properties, "profile")
	assert.Contains(t, schema.Properties, "tags")

	// Check required fields
	assert.Contains(t, schema.Required, "name")
	assert.Contains(t, schema.Required, "email")
	assert.Contains(t, schema.Required, "profile")

	// Check profile property (nested object)
	profileProp := schema.Properties["profile"]
	assert.Equal(t, "object", profileProp.Type)
	assert.Contains(t, profileProp.Description, "object with")
	assert.Contains(t, profileProp.Description, "bio")
	assert.Contains(t, profileProp.Description, "avatar")
	assert.Contains(t, profileProp.Description, "preferences")

	// Check tags property (array)
	tagsProp := schema.Properties["tags"]
	assert.Equal(t, "array", tagsProp.Type)
	assert.Contains(t, tagsProp.Description, "array of string")
}

func TestConvertSchemaToPropertyWithReferences(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	// Test with nested object schema
	nestedSchema := openapi.Schema{
		Type:        "object",
		Description: "Nested object",
		Properties: map[string]openapi.Schema{
			"field1": {
				Type:        "string",
				Description: "First field",
			},
			"field2": {
				Type:        "integer",
				Description: "Second field",
			},
		},
		Required: []string{"field1"},
	}

	property, err := generator.convertSchemaToPropertyWithReferences(nestedSchema)
	require.NoError(t, err)

	assert.Equal(t, "object", property.Type)
	assert.Contains(t, property.Description, "Nested object")
	assert.Contains(t, property.Description, "object with 2 properties")
	assert.Contains(t, property.Description, "field1")
	assert.Contains(t, property.Description, "field2")

	// Test with array of objects
	arraySchema := openapi.Schema{
		Type:        "array",
		Description: "Array of objects",
		Items:       &nestedSchema,
	}

	arrayProperty, err := generator.convertSchemaToPropertyWithReferences(arraySchema)
	require.NoError(t, err)

	assert.Equal(t, "array", arrayProperty.Type)
	assert.Contains(t, arrayProperty.Description, "array of object")
}

func TestResolveSchemaReference(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	// Test with simple schema
	schema := openapi.Schema{
		Type:        "string",
		Description: "Test schema",
	}

	resolved, err := generator.resolveSchemaReference(schema)
	require.NoError(t, err)
	assert.Equal(t, schema, resolved)
}

func TestValidateInput(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{
		OpenAPI: config.OpenAPIConfig{
			BaseURL: "https://api.example.com",
		},
	}
	spec := &openapi.ParsedSpec{
		Endpoints: []openapi.Endpoint{
			{Path: "/test", Method: "GET"},
		},
	}

	generator := NewMCPToolGenerator(spec, config, logger)

	// Test valid input
	err := generator.validateInput()
	assert.NoError(t, err)

	// Test nil spec
	generator.spec = nil
	err = generator.validateInput()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "specification is nil")

	// Test nil config
	generator.spec = spec
	generator.config = nil
	err = generator.validateInput()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration is nil")

	// Test nil logger
	generator.config = config
	generator.logger = nil
	err = generator.validateInput()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logger is nil")

	// Test empty endpoints
	generator.logger = logger
	generator.spec.Endpoints = []openapi.Endpoint{}
	err = generator.validateInput()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no endpoints found")

	// Test empty base URL
	generator.spec.Endpoints = []openapi.Endpoint{{Path: "/test", Method: "GET"}}
	generator.config.OpenAPI.BaseURL = ""
	err = generator.validateInput()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base URL is required")
}

func TestValidateTool(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	// Test valid tool
	validTool := &mcp.Tool{
		Name:        "test_tool",
		Description: "Test tool",
		InputSchema: &mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"test": {
					Type: "string",
				},
			},
			Required: []string{"test"},
		},
		Handler: func(params map[string]interface{}) (interface{}, error) {
			return "test", nil
		},
	}

	err := generator.validateTool(validTool)
	assert.NoError(t, err)

	// Test nil tool
	err = generator.validateTool(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool is nil")

	// Test empty name
	tool := *validTool
	tool.Name = ""
	err = generator.validateTool(&tool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool name is empty")

	// Test empty description
	tool = *validTool
	tool.Description = ""
	err = generator.validateTool(&tool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool description is empty")

	// Test nil input schema
	tool = *validTool
	tool.InputSchema = nil
	err = generator.validateTool(&tool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool input schema is nil")

	// Test nil handler
	tool = *validTool
	tool.Handler = nil
	err = generator.validateTool(&tool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool handler is nil")
}

func TestValidateInputSchema(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	// Test valid schema
	validSchema := &mcp.InputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"test": {
				Type: "string",
			},
		},
		Required: []string{"test"},
	}

	err := generator.validateInputSchema(validSchema)
	assert.NoError(t, err)

	// Test nil schema
	err = generator.validateInputSchema(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema is nil")

	// Test empty type
	schema := *validSchema
	schema.Type = ""
	err = generator.validateInputSchema(&schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema type is empty")

	// Test unsupported type
	schema = *validSchema
	schema.Type = "array"
	err = generator.validateInputSchema(&schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported schema type")

	// Test empty property name
	schema = *validSchema
	schema.Properties[""] = mcp.Property{Type: "string"}
	err = generator.validateInputSchema(&schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "property name is empty")

	// Test required field not in properties
	schema = *validSchema
	schema.Properties = map[string]mcp.Property{
		"test": {Type: "string"},
	}
	schema.Required = []string{"missing"}
	err = generator.validateInputSchema(&schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required field 'missing' not found in properties")
}

func TestValidateProperty(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	// Test valid property
	validProperty := mcp.Property{
		Type: "string",
	}

	err := generator.validateProperty(validProperty)
	assert.NoError(t, err)

	// Test empty type
	property := validProperty
	property.Type = ""
	err = generator.validateProperty(property)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "property type is empty")

	// Test invalid string constraints
	property = validProperty
	minLength := 10
	maxLength := 5
	property.MinLength = &minLength
	property.MaxLength = &maxLength
	err = generator.validateProperty(property)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minLength (10) cannot be greater than maxLength (5)")

	// Test invalid numeric constraints
	property = mcp.Property{Type: "integer"}
	minimum := 10.0
	maximum := 5.0
	property.Minimum = &minimum
	property.Maximum = &maximum
	err = generator.validateProperty(property)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minimum (10.000000) cannot be greater than maximum (5.000000)")

	// Test enum with non-string type
	property = mcp.Property{
		Type: "integer",
		Enum: []string{"1", "2", "3"},
	}
	err = generator.validateProperty(property)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "enum can only be used with string type")
}

func TestGenerateToolName(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	testCases := []struct {
		endpoint     openapi.Endpoint
		expectedName string
	}{
		{
			endpoint: openapi.Endpoint{
				Path:        "/users",
				Method:      "GET",
				OperationID: "getUsers",
			},
			expectedName: "getusers",
		},
		{
			endpoint: openapi.Endpoint{
				Path:        "/users/{id}",
				Method:      "GET",
				OperationID: "getUserById",
			},
			expectedName: "getuserbyid",
		},
		{
			endpoint: openapi.Endpoint{
				Path:        "/users/{id}/posts",
				Method:      "GET",
				OperationID: "getUserPosts",
			},
			expectedName: "getuserposts",
		},
		{
			endpoint: openapi.Endpoint{
				Path:   "/users",
				Method: "GET",
				// No OperationID
			},
			expectedName: "get_users",
		},
		{
			endpoint: openapi.Endpoint{
				Path:   "/users/{userId}/posts/{postId}",
				Method: "GET",
				// No OperationID
			},
			expectedName: "get_users_userId_posts_postId",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedName, func(t *testing.T) {
			name := generator.generateToolName(tc.endpoint)
			assert.Equal(t, tc.expectedName, name)
		})
	}
}

func TestGenerateToolDescription(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	testCases := []struct {
		endpoint            openapi.Endpoint
		expectedDescription string
	}{
		{
			endpoint: openapi.Endpoint{
				Summary: "Get all users",
			},
			expectedDescription: "Get all users",
		},
		{
			endpoint: openapi.Endpoint{
				Description: "Retrieve a list of all users in the system",
			},
			expectedDescription: "Retrieve a list of all users in the system",
		},
		{
			endpoint: openapi.Endpoint{
				Path:   "/users",
				Method: "GET",
			},
			expectedDescription: "GET /users",
		},
		{
			endpoint: openapi.Endpoint{
				Summary:     "Get users",
				Description: "Retrieve users with detailed information",
			},
			expectedDescription: "Get users", // Summary takes precedence
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedDescription, func(t *testing.T) {
			description := generator.generateToolDescription(tc.endpoint)
			assert.Equal(t, tc.expectedDescription, description)
		})
	}
}

func TestConvertParameterToProperty(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	param := openapi.Parameter{
		Name:        "userId",
		Description: "User ID",
		Schema: openapi.Schema{
			Type:        "integer",
			Format:      "int64",
			Description: "User identifier",
			Minimum:     func() *float64 { v := float64(1); return &v }(),
			Maximum:     func() *float64 { v := float64(1000000); return &v }(),
			Default:     1,
		},
	}

	property := generator.convertParameterToProperty(param)

	assert.Equal(t, "integer", property.Type)
	assert.Equal(t, "User ID", property.Description)
	assert.Equal(t, "int64", property.Format)
	assert.NotNil(t, property.Minimum)
	assert.Equal(t, float64(1), *property.Minimum)
	assert.NotNil(t, property.Maximum)
	assert.Equal(t, float64(1000000), *property.Maximum)
	assert.Equal(t, 1, property.Default)
}

func TestConvertParameterToProperty_WithEnum(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	param := openapi.Parameter{
		Name:        "status",
		Description: "User status",
		Schema: openapi.Schema{
			Type: "string",
			Enum: []interface{}{"active", "inactive", "pending"},
		},
	}

	property := generator.convertParameterToProperty(param)

	assert.Equal(t, "string", property.Type)
	assert.Equal(t, "User status", property.Description)
	assert.Len(t, property.Enum, 3)
	assert.Contains(t, property.Enum, "active")
	assert.Contains(t, property.Enum, "inactive")
	assert.Contains(t, property.Enum, "pending")
}

func TestMapOpenAPITypeToMCPType(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	testCases := []struct {
		openAPIType     string
		expectedMCPType string
	}{
		{"string", "string"},
		{"integer", "integer"},
		{"number", "number"},
		{"boolean", "boolean"},
		{"array", "array"},
		{"object", "object"},
		{"unknown", "string"}, // Default fallback
	}

	for _, tc := range testCases {
		t.Run(tc.openAPIType, func(t *testing.T) {
			mcpType := generator.mapOpenAPITypeToMCPType(tc.openAPIType)
			assert.Equal(t, tc.expectedMCPType, mcpType)
		})
	}
}

func TestBuildURL(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	testCases := []struct {
		path     string
		params   map[string]interface{}
		expected string
	}{
		{
			path:     "/users",
			params:   map[string]interface{}{},
			expected: "/users",
		},
		{
			path:     "/users/{id}",
			params:   map[string]interface{}{"id": 123},
			expected: "/users/123",
		},
		{
			path:     "/users/{userId}/posts/{postId}",
			params:   map[string]interface{}{"userId": 456, "postId": 789},
			expected: "/users/456/posts/789",
		},
		{
			path:     "/users/{id}",
			params:   map[string]interface{}{"id": "test", "other": "ignored"},
			expected: "/users/test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			result := generator.buildURL(tc.path, tc.params)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestShouldIncludeEndpoint(t *testing.T) {
	logger := logrus.New()
	config := &config.Config{
		Filters: config.FilterConfig{},
	}
	spec := &openapi.ParsedSpec{}
	generator := NewMCPToolGenerator(spec, config, logger)

	// Test with no filters (should include all)
	endpoint := openapi.Endpoint{
		Path:   "/users",
		Method: "GET",
	}
	assert.True(t, generator.shouldIncludeEndpoint(endpoint))

	// Test with include path filter
	config.Filters.IncludePaths = []string{"/users"}
	assert.True(t, generator.shouldIncludeEndpoint(endpoint))

	config.Filters.IncludePaths = []string{"/admin"}
	assert.False(t, generator.shouldIncludeEndpoint(endpoint))

	// Test with exclude path filter
	config.Filters.IncludePaths = []string{}
	config.Filters.ExcludePaths = []string{"/users"}
	assert.False(t, generator.shouldIncludeEndpoint(endpoint))

	config.Filters.ExcludePaths = []string{"/admin"}
	assert.True(t, generator.shouldIncludeEndpoint(endpoint))

	// Test with include method filter
	config.Filters.IncludePaths = []string{}
	config.Filters.ExcludePaths = []string{}
	config.Filters.IncludeMethods = []string{"GET"}
	assert.True(t, generator.shouldIncludeEndpoint(endpoint))

	config.Filters.IncludeMethods = []string{"POST"}
	assert.False(t, generator.shouldIncludeEndpoint(endpoint))

	// Test with exclude method filter
	config.Filters.IncludeMethods = []string{}
	config.Filters.ExcludeMethods = []string{"GET"}
	assert.False(t, generator.shouldIncludeEndpoint(endpoint))

	config.Filters.ExcludeMethods = []string{"POST"}
	assert.True(t, generator.shouldIncludeEndpoint(endpoint))
}

func TestGenerateTools_IntegrationWithRealSpec(t *testing.T) {
	// Use the example petstore spec if it exists
	specPath := "../../examples/petstore.yaml"
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("Petstore spec not found, skipping integration test")
	}

	// Parse the spec
	logger := logrus.New()
	openAPIParser := parser.NewOpenAPIParser(specPath, logger)
	spec, err := openAPIParser.ParseSpec()
	require.NoError(t, err)

	// Create generator
	config := &config.Config{
		OpenAPI: config.OpenAPIConfig{
			BaseURL: "https://petstore3.swagger.io/api/v3",
		},
		Filters: config.FilterConfig{},
	}

	generator := NewMCPToolGenerator(spec, config, logger)

	// Generate tools
	tools, err := generator.GenerateTools()
	require.NoError(t, err)
	assert.Greater(t, len(tools), 0)

	// Verify we have some expected tools
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
		assert.NotEmpty(t, tool.Name)
		assert.NotEmpty(t, tool.Description)
		assert.NotNil(t, tool.InputSchema)
		assert.NotNil(t, tool.Handler)
	}

	// Check for some expected petstore endpoints
	expectedTools := []string{"getpet", "addpet", "updatepet", "findpetsbystatus"}
	foundExpected := 0
	for _, expected := range expectedTools {
		if toolNames[expected] {
			foundExpected++
		}
	}
	assert.Greater(t, foundExpected, 0, "Should have found some expected petstore tools")
}
