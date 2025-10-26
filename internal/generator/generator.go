package generator

import (
	"fmt"
	"strings"

	"api-to-mcp/internal/config"
	"api-to-mcp/internal/utils"
	"api-to-mcp/pkg/mcp"
	"api-to-mcp/pkg/openapi"

	"github.com/sirupsen/logrus"
)

// MCPToolGenerator generates MCP tools from OpenAPI specifications
type MCPToolGenerator struct {
	spec   *openapi.ParsedSpec
	config *config.Config
	logger *logrus.Logger
}

// NewMCPToolGenerator creates a new MCP tool generator
func NewMCPToolGenerator(spec *openapi.ParsedSpec, cfg *config.Config, logger *logrus.Logger) *MCPToolGenerator {
	return &MCPToolGenerator{
		spec:   spec,
		config: cfg,
		logger: logger,
	}
}

// GenerateTools generates MCP tools from the OpenAPI specification
func (g *MCPToolGenerator) GenerateTools() ([]mcp.Tool, error) {
	g.logger.Info("Generating MCP tools from OpenAPI specification")

	// Validate input
	if err := g.validateInput(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	tools := make([]mcp.Tool, 0)
	errors := make([]error, 0)

	for _, endpoint := range g.spec.Endpoints {
		// Apply filters
		if !g.shouldIncludeEndpoint(endpoint) {
			g.logger.WithFields(logrus.Fields{
				"path":   endpoint.Path,
				"method": endpoint.Method,
			}).Debug("Skipping filtered endpoint")
			continue
		}

		// Generate tool for this endpoint
		tool, err := g.generateToolForEndpoint(endpoint)
		if err != nil {
			errorMsg := fmt.Errorf("failed to generate tool for endpoint %s %s: %w", endpoint.Method, endpoint.Path, err)
			errors = append(errors, errorMsg)
			g.logger.WithError(err).WithFields(logrus.Fields{
				"path":   endpoint.Path,
				"method": endpoint.Method,
			}).Error("Failed to generate tool for endpoint")
			continue
		}

		// Validate generated tool
		if err := g.validateTool(tool); err != nil {
			errorMsg := fmt.Errorf("generated tool validation failed for %s %s: %w", endpoint.Method, endpoint.Path, err)
			errors = append(errors, errorMsg)
			g.logger.WithError(err).WithFields(logrus.Fields{
				"path":   endpoint.Path,
				"method": endpoint.Method,
				"tool":   tool.Name,
			}).Error("Generated tool failed validation")
			continue
		}

		tools = append(tools, *tool)
	}

	// Log summary
	g.logger.WithFields(logrus.Fields{
		"tool_count":      len(tools),
		"error_count":     len(errors),
		"total_endpoints": len(g.spec.Endpoints),
	}).Info("Generated MCP tools")

	// If we have errors but some tools were generated, log warnings
	if len(errors) > 0 {
		g.logger.WithField("error_count", len(errors)).Warn("Some tools failed to generate")
		for _, err := range errors {
			g.logger.WithError(err).Warn("Tool generation error")
		}
	}

	// If no tools were generated, return an error
	if len(tools) == 0 {
		if len(errors) > 0 {
			return nil, fmt.Errorf("no tools could be generated: %d errors occurred", len(errors))
		}
		return nil, fmt.Errorf("no tools could be generated: all endpoints were filtered out")
	}

	return tools, nil
}

// generateToolForEndpoint generates a single MCP tool for an endpoint
func (g *MCPToolGenerator) generateToolForEndpoint(endpoint openapi.Endpoint) (*mcp.Tool, error) {
	// Generate tool name
	toolName := g.generateToolName(endpoint)

	// Generate tool description
	description := g.generateToolDescription(endpoint)

	// Generate input schema
	inputSchema, err := g.generateInputSchema(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to generate input schema: %w", err)
	}

	// Create HTTP client for this tool
	httpClient := utils.NewHTTPClient(g.config.OpenAPI.BaseURL, g.logger)

	// Create tool handler
	handler := g.createToolHandler(endpoint, httpClient)

	tool := &mcp.Tool{
		Name:        toolName,
		Description: description,
		InputSchema: inputSchema,
		Handler:     handler,
	}

	g.logger.WithFields(logrus.Fields{
		"tool_name": toolName,
		"path":      endpoint.Path,
		"method":    endpoint.Method,
	}).Debug("Generated tool for endpoint")

	return tool, nil
}

// generateToolName generates a tool name from an endpoint
func (g *MCPToolGenerator) generateToolName(endpoint openapi.Endpoint) string {
	// Use operation ID if available
	if endpoint.OperationID != "" {
		return strings.ToLower(endpoint.OperationID)
	}

	// Generate from path and method
	path := strings.TrimPrefix(endpoint.Path, "/")
	path = strings.ReplaceAll(path, "/", "_")
	path = strings.ReplaceAll(path, "{", "")
	path = strings.ReplaceAll(path, "}", "")

	method := strings.ToLower(endpoint.Method)

	return fmt.Sprintf("%s_%s", method, path)
}

// generateToolDescription generates a tool description from an endpoint
func (g *MCPToolGenerator) generateToolDescription(endpoint openapi.Endpoint) string {
	if endpoint.Summary != "" {
		return endpoint.Summary
	}

	if endpoint.Description != "" {
		return endpoint.Description
	}

	return fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path)
}

// generateInputSchema generates the input schema for a tool
func (g *MCPToolGenerator) generateInputSchema(endpoint openapi.Endpoint) (*mcp.InputSchema, error) {
	schema := &mcp.InputSchema{
		Type:       "object",
		Properties: make(map[string]mcp.Property),
		Required:   make([]string, 0),
	}

	// Add path parameters
	for _, param := range endpoint.Parameters {
		if param.In == "path" {
			property := g.convertParameterToProperty(param)
			schema.Properties[param.Name] = property
			if param.Required {
				schema.Required = append(schema.Required, param.Name)
			}
		}
	}

	// Add query parameters
	for _, param := range endpoint.Parameters {
		if param.In == "query" {
			property := g.convertParameterToProperty(param)
			schema.Properties[param.Name] = property
			if param.Required {
				schema.Required = append(schema.Required, param.Name)
			}
		}
	}

	// Add request body parameters
	if endpoint.RequestBody != nil {
		// Parse request body schema properly
		bodySchema, err := g.parseRequestBodySchema(endpoint.RequestBody)
		if err != nil {
			g.logger.WithError(err).Warn("Failed to parse request body schema, using fallback")
			// Fallback to simple body parameter
			schema.Properties["body"] = mcp.Property{
				Type:        "object",
				Description: endpoint.RequestBody.Description,
			}
		} else {
			// Merge body schema properties into main schema
			for key, property := range bodySchema.Properties {
				schema.Properties[key] = property
			}
			// Add body schema required fields
			for _, required := range bodySchema.Required {
				schema.Required = append(schema.Required, required)
			}
		}
	}

	return schema, nil
}

// convertParameterToProperty converts an OpenAPI parameter to an MCP property
func (g *MCPToolGenerator) convertParameterToProperty(param openapi.Parameter) mcp.Property {
	property := mcp.Property{
		Type:        g.mapOpenAPITypeToMCPType(param.Schema.Type),
		Description: param.Description,
	}

	// Add format if available
	if param.Schema.Format != "" {
		property.Format = param.Schema.Format
	}

	// Add enum if available
	if len(param.Schema.Enum) > 0 {
		enum := make([]string, len(param.Schema.Enum))
		for i, v := range param.Schema.Enum {
			enum[i] = fmt.Sprintf("%v", v)
		}
		property.Enum = enum
	}

	// Add default if available
	if param.Schema.Default != nil {
		property.Default = param.Schema.Default
	}

	// Add constraints
	if param.Schema.Minimum != nil {
		property.Minimum = param.Schema.Minimum
	}
	if param.Schema.Maximum != nil {
		property.Maximum = param.Schema.Maximum
	}
	if param.Schema.MinLength != nil {
		property.MinLength = param.Schema.MinLength
	}
	if param.Schema.MaxLength != nil {
		property.MaxLength = param.Schema.MaxLength
	}
	if param.Schema.Pattern != "" {
		property.Pattern = param.Schema.Pattern
	}

	return property
}

// mapOpenAPITypeToMCPType maps OpenAPI types to MCP types
func (g *MCPToolGenerator) mapOpenAPITypeToMCPType(openAPIType string) string {
	switch openAPIType {
	case "string":
		return "string"
	case "integer":
		return "integer"
	case "number":
		return "number"
	case "boolean":
		return "boolean"
	case "array":
		return "array"
	case "object":
		return "object"
	default:
		return "string" // Default to string
	}
}

// createToolHandler creates a handler function for a tool
func (g *MCPToolGenerator) createToolHandler(endpoint openapi.Endpoint, httpClient *utils.HTTPClient) func(map[string]interface{}) (interface{}, error) {
	return func(params map[string]interface{}) (interface{}, error) {
		// Build URL with path parameters
		url := g.buildURL(endpoint.Path, params)

		// Make HTTP request
		response, err := httpClient.MakeRequest(endpoint.Method, url, params)
		if err != nil {
			return nil, fmt.Errorf("HTTP request failed: %w", err)
		}

		return response, nil
	}
}

// buildURL builds the URL for an endpoint with path parameters
func (g *MCPToolGenerator) buildURL(path string, params map[string]interface{}) string {
	url := path

	// Replace path parameters
	for key, value := range params {
		placeholder := fmt.Sprintf("{%s}", key)
		if strings.Contains(url, placeholder) {
			url = strings.ReplaceAll(url, placeholder, fmt.Sprintf("%v", value))
		}
	}

	return url
}

// shouldIncludeEndpoint checks if an endpoint should be included based on filters
func (g *MCPToolGenerator) shouldIncludeEndpoint(endpoint openapi.Endpoint) bool {
	// Check path filters
	if len(g.config.Filters.IncludePaths) > 0 {
		include := false
		for _, includePath := range g.config.Filters.IncludePaths {
			if strings.HasPrefix(endpoint.Path, includePath) {
				include = true
				break
			}
		}
		if !include {
			return false
		}
	}

	if len(g.config.Filters.ExcludePaths) > 0 {
		for _, excludePath := range g.config.Filters.ExcludePaths {
			if strings.HasPrefix(endpoint.Path, excludePath) {
				return false
			}
		}
	}

	// Check method filters
	if len(g.config.Filters.IncludeMethods) > 0 {
		include := false
		for _, includeMethod := range g.config.Filters.IncludeMethods {
			if strings.EqualFold(endpoint.Method, includeMethod) {
				include = true
				break
			}
		}
		if !include {
			return false
		}
	}

	if len(g.config.Filters.ExcludeMethods) > 0 {
		for _, excludeMethod := range g.config.Filters.ExcludeMethods {
			if strings.EqualFold(endpoint.Method, excludeMethod) {
				return false
			}
		}
	}

	return true
}

// parseRequestBodySchema parses the request body schema and converts it to MCP input schema
func (g *MCPToolGenerator) parseRequestBodySchema(requestBody *openapi.RequestBody) (*mcp.InputSchema, error) {
	if requestBody == nil {
		return nil, fmt.Errorf("request body is nil")
	}

	// Look for JSON content type
	jsonContent, exists := requestBody.Content["application/json"]
	if !exists {
		// Fallback to any content type
		for contentType, content := range requestBody.Content {
			if contentType == "application/json" || contentType == "application/*" || contentType == "*/*" {
				jsonContent = content
				exists = true
				break
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("no supported content type found in request body")
	}

	// Convert the schema to MCP input schema
	return g.convertSchemaToInputSchema(jsonContent.Schema)
}

// convertSchemaToInputSchema converts an OpenAPI schema to MCP input schema
func (g *MCPToolGenerator) convertSchemaToInputSchema(schema openapi.Schema) (*mcp.InputSchema, error) {
	inputSchema := &mcp.InputSchema{
		Type:       "object",
		Properties: make(map[string]mcp.Property),
		Required:   make([]string, 0),
	}

	// Handle object type
	if schema.Type == "object" {
		// Add properties
		for name, propSchema := range schema.Properties {
			property, err := g.convertSchemaToProperty(propSchema)
			if err != nil {
				g.logger.WithError(err).WithField("property", name).Warn("Failed to convert property schema")
				continue
			}
			inputSchema.Properties[name] = property
		}

		// Add required fields
		inputSchema.Required = append(inputSchema.Required, schema.Required...)
	} else {
		// Handle non-object types (array, primitive)
		property, err := g.convertSchemaToProperty(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert schema to property: %w", err)
		}
		inputSchema.Properties["value"] = property
		if schema.Required != nil && len(schema.Required) > 0 {
			inputSchema.Required = append(inputSchema.Required, "value")
		}
	}

	return inputSchema, nil
}

// convertSchemaToProperty converts an OpenAPI schema to MCP property
func (g *MCPToolGenerator) convertSchemaToProperty(schema openapi.Schema) (mcp.Property, error) {
	property := mcp.Property{
		Type:        g.mapOpenAPITypeToMCPType(schema.Type),
		Description: schema.Description,
		Format:      schema.Format,
		Default:     schema.Default,
	}

	// Add constraints
	if schema.Minimum != nil {
		property.Minimum = schema.Minimum
	}
	if schema.Maximum != nil {
		property.Maximum = schema.Maximum
	}
	if schema.MinLength != nil {
		property.MinLength = schema.MinLength
	}
	if schema.MaxLength != nil {
		property.MaxLength = schema.MaxLength
	}
	if schema.Pattern != "" {
		property.Pattern = schema.Pattern
	}

	// Add enum
	if len(schema.Enum) > 0 {
		enum := make([]string, len(schema.Enum))
		for i, v := range schema.Enum {
			enum[i] = fmt.Sprintf("%v", v)
		}
		property.Enum = enum
	}

	// Handle array items
	if schema.Type == "array" && schema.Items != nil {
		itemsProperty, err := g.convertSchemaToProperty(*schema.Items)
		if err != nil {
			return property, fmt.Errorf("failed to convert array items: %w", err)
		}
		// For arrays, we'll store the items schema in a custom field
		// This is a simplified approach - in a full implementation,
		// you might want to handle nested schemas more comprehensively
		property.Description = fmt.Sprintf("%s (array of %s)", property.Description, itemsProperty.Type)
	}

	// Handle object properties for nested objects
	if schema.Type == "object" && len(schema.Properties) > 0 {
		// For nested objects, we'll create a simplified representation
		// In a full implementation, you might want to flatten or handle nested objects differently
		property.Description = fmt.Sprintf("%s (object with %d properties)", property.Description, len(schema.Properties))

		// Add a note about the object structure
		propertyNames := make([]string, 0, len(schema.Properties))
		for name := range schema.Properties {
			propertyNames = append(propertyNames, name)
		}
		if len(propertyNames) > 0 {
			property.Description = fmt.Sprintf("%s - properties: %s", property.Description, strings.Join(propertyNames, ", "))
		}
	}

	return property, nil
}

// convertSchemaToInputSchemaWithReferences converts an OpenAPI schema to MCP input schema with reference support
func (g *MCPToolGenerator) convertSchemaToInputSchemaWithReferences(schema openapi.Schema) (*mcp.InputSchema, error) {
	inputSchema := &mcp.InputSchema{
		Type:       "object",
		Properties: make(map[string]mcp.Property),
		Required:   make([]string, 0),
	}

	// Handle object type
	if schema.Type == "object" {
		// Add properties
		for name, propSchema := range schema.Properties {
			property, err := g.convertSchemaToPropertyWithReferences(propSchema)
			if err != nil {
				g.logger.WithError(err).WithField("property", name).Warn("Failed to convert property schema")
				continue
			}
			inputSchema.Properties[name] = property
		}

		// Add required fields
		inputSchema.Required = append(inputSchema.Required, schema.Required...)
	} else {
		// Handle non-object types (array, primitive)
		property, err := g.convertSchemaToPropertyWithReferences(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert schema to property: %w", err)
		}
		inputSchema.Properties["value"] = property
		if schema.Required != nil && len(schema.Required) > 0 {
			inputSchema.Required = append(inputSchema.Required, "value")
		}
	}

	return inputSchema, nil
}

// convertSchemaToPropertyWithReferences converts an OpenAPI schema to MCP property with reference support
func (g *MCPToolGenerator) convertSchemaToPropertyWithReferences(schema openapi.Schema) (mcp.Property, error) {
	property := mcp.Property{
		Type:        g.mapOpenAPITypeToMCPType(schema.Type),
		Description: schema.Description,
		Format:      schema.Format,
		Default:     schema.Default,
	}

	// Add constraints
	if schema.Minimum != nil {
		property.Minimum = schema.Minimum
	}
	if schema.Maximum != nil {
		property.Maximum = schema.Maximum
	}
	if schema.MinLength != nil {
		property.MinLength = schema.MinLength
	}
	if schema.MaxLength != nil {
		property.MaxLength = schema.MaxLength
	}
	if schema.Pattern != "" {
		property.Pattern = schema.Pattern
	}

	// Add enum
	if len(schema.Enum) > 0 {
		enum := make([]string, len(schema.Enum))
		for i, v := range schema.Enum {
			enum[i] = fmt.Sprintf("%v", v)
		}
		property.Enum = enum
	}

	// Handle array items
	if schema.Type == "array" && schema.Items != nil {
		itemsProperty, err := g.convertSchemaToPropertyWithReferences(*schema.Items)
		if err != nil {
			return property, fmt.Errorf("failed to convert array items: %w", err)
		}
		property.Description = fmt.Sprintf("%s (array of %s)", property.Description, itemsProperty.Type)
	}

	// Handle object properties for nested objects
	if schema.Type == "object" && len(schema.Properties) > 0 {
		property.Description = fmt.Sprintf("%s (object with %d properties)", property.Description, len(schema.Properties))

		// Add a note about the object structure
		propertyNames := make([]string, 0, len(schema.Properties))
		for name := range schema.Properties {
			propertyNames = append(propertyNames, name)
		}
		if len(propertyNames) > 0 {
			property.Description = fmt.Sprintf("%s - properties: %s", property.Description, strings.Join(propertyNames, ", "))
		}
	}

	return property, nil
}

// resolveSchemaReference resolves a schema reference if it exists in the components
func (g *MCPToolGenerator) resolveSchemaReference(schema openapi.Schema) (openapi.Schema, error) {
	// This is a placeholder for schema reference resolution
	// In a full implementation, you would resolve $ref references to components
	// For now, we'll return the schema as-is
	return schema, nil
}

// validateInput validates the input to the generator
func (g *MCPToolGenerator) validateInput() error {
	if g.spec == nil {
		return fmt.Errorf("specification is nil")
	}

	if g.config == nil {
		return fmt.Errorf("configuration is nil")
	}

	if g.logger == nil {
		return fmt.Errorf("logger is nil")
	}

	if len(g.spec.Endpoints) == 0 {
		return fmt.Errorf("no endpoints found in specification")
	}

	// Validate configuration
	if g.config.OpenAPI.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	return nil
}

// validateTool validates a generated tool
func (g *MCPToolGenerator) validateTool(tool *mcp.Tool) error {
	if tool == nil {
		return fmt.Errorf("tool is nil")
	}

	if tool.Name == "" {
		return fmt.Errorf("tool name is empty")
	}

	if tool.Description == "" {
		return fmt.Errorf("tool description is empty")
	}

	if tool.InputSchema == nil {
		return fmt.Errorf("tool input schema is nil")
	}

	if tool.Handler == nil {
		return fmt.Errorf("tool handler is nil")
	}

	// Validate input schema
	if err := g.validateInputSchema(tool.InputSchema); err != nil {
		return fmt.Errorf("input schema validation failed: %w", err)
	}

	return nil
}

// validateInputSchema validates an input schema
func (g *MCPToolGenerator) validateInputSchema(schema *mcp.InputSchema) error {
	if schema == nil {
		return fmt.Errorf("schema is nil")
	}

	if schema.Type == "" {
		return fmt.Errorf("schema type is empty")
	}

	if schema.Type != "object" {
		return fmt.Errorf("unsupported schema type: %s", schema.Type)
	}

	// Validate properties
	for name, property := range schema.Properties {
		if name == "" {
			return fmt.Errorf("property name is empty")
		}

		if err := g.validateProperty(property); err != nil {
			return fmt.Errorf("property '%s' validation failed: %w", name, err)
		}
	}

	// Validate required fields
	for _, required := range schema.Required {
		if required == "" {
			return fmt.Errorf("required field name is empty")
		}

		if _, exists := schema.Properties[required]; !exists {
			return fmt.Errorf("required field '%s' not found in properties", required)
		}
	}

	return nil
}

// validateProperty validates a property
func (g *MCPToolGenerator) validateProperty(property mcp.Property) error {
	if property.Type == "" {
		return fmt.Errorf("property type is empty")
	}

	// Validate type-specific constraints
	switch property.Type {
	case "string":
		if property.MinLength != nil && property.MaxLength != nil {
			if *property.MinLength > *property.MaxLength {
				return fmt.Errorf("minLength (%d) cannot be greater than maxLength (%d)", *property.MinLength, *property.MaxLength)
			}
		}
	case "integer", "number":
		if property.Minimum != nil && property.Maximum != nil {
			if *property.Minimum > *property.Maximum {
				return fmt.Errorf("minimum (%f) cannot be greater than maximum (%f)", *property.Minimum, *property.Maximum)
			}
		}
	}

	// Validate enum values
	if len(property.Enum) > 0 {
		if property.Type != "string" {
			return fmt.Errorf("enum can only be used with string type, got %s", property.Type)
		}
	}

	return nil
}
