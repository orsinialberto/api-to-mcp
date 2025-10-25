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

	tools := make([]mcp.Tool, 0)

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
			g.logger.WithError(err).WithFields(logrus.Fields{
				"path":   endpoint.Path,
				"method": endpoint.Method,
			}).Error("Failed to generate tool for endpoint")
			continue
		}

		tools = append(tools, *tool)
	}

	g.logger.WithField("tool_count", len(tools)).Info("Generated MCP tools")
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
		// For now, we'll add a simple "body" parameter
		// TODO: Parse request body schema properly
		schema.Properties["body"] = mcp.Property{
			Type:        "object",
			Description: endpoint.RequestBody.Description,
		}
		if endpoint.RequestBody.Required {
			schema.Required = append(schema.Required, "body")
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
