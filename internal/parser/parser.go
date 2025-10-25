package parser

import (
	"fmt"
	"os"

	"api-to-mcp/pkg/openapi"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/sirupsen/logrus"
)

// OpenAPIParser parses OpenAPI specifications
type OpenAPIParser struct {
	specPath string
	logger   *logrus.Logger
}

// NewOpenAPIParser creates a new OpenAPI parser
func NewOpenAPIParser(specPath string, logger *logrus.Logger) *OpenAPIParser {
	return &OpenAPIParser{
		specPath: specPath,
		logger:   logger,
	}
}

// ParseSpec parses the OpenAPI specification
func (p *OpenAPIParser) ParseSpec() (*openapi.ParsedSpec, error) {
	p.logger.WithField("spec_path", p.specPath).Info("Parsing OpenAPI specification")

	// Check if file exists
	if _, err := os.Stat(p.specPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("specification file not found: %s", p.specPath)
	}

	// Load the OpenAPI document
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(p.specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	// Validate the document
	if err := doc.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI specification: %w", err)
	}

	// Convert to our internal representation
	parsedSpec := p.convertToParsedSpec(doc)

	p.logger.WithFields(logrus.Fields{
		"title":      parsedSpec.Info.Title,
		"version":    parsedSpec.Info.Version,
		"endpoints":  len(parsedSpec.Endpoints),
		"components": len(parsedSpec.Components),
	}).Info("Successfully parsed OpenAPI specification")

	return parsedSpec, nil
}

// convertToParsedSpec converts OpenAPI3 document to our internal representation
func (p *OpenAPIParser) convertToParsedSpec(doc *openapi3.T) *openapi.ParsedSpec {
	spec := &openapi.ParsedSpec{
		Info: openapi.Info{
			Title:       doc.Info.Title,
			Version:     doc.Info.Version,
			Description: doc.Info.Description,
		},
		Servers:    make([]openapi.Server, 0),
		Endpoints:  make([]openapi.Endpoint, 0),
		Components: make(map[string]openapi.Component),
	}

	// Convert servers
	for _, server := range doc.Servers {
		spec.Servers = append(spec.Servers, openapi.Server{
			URL:         server.URL,
			Description: server.Description,
		})
	}

	// Convert paths and operations
	for path, pathItem := range doc.Paths.Map() {
		p.convertPathItem(path, pathItem, spec)
	}

	// Convert components
	if doc.Components != nil {
		p.convertComponents(doc.Components, spec)
	}

	return spec
}

// convertPathItem converts a path item to endpoints
func (p *OpenAPIParser) convertPathItem(path string, pathItem *openapi3.PathItem, spec *openapi.ParsedSpec) {
	operations := map[string]*openapi3.Operation{
		"GET":     pathItem.Get,
		"POST":    pathItem.Post,
		"PUT":     pathItem.Put,
		"DELETE":  pathItem.Delete,
		"PATCH":   pathItem.Patch,
		"HEAD":    pathItem.Head,
		"OPTIONS": pathItem.Options,
	}

	for method, operation := range operations {
		if operation == nil {
			continue
		}

		endpoint := openapi.Endpoint{
			Path:        path,
			Method:      method,
			OperationID: operation.OperationID,
			Summary:     operation.Summary,
			Description: operation.Description,
			Parameters:  make([]openapi.Parameter, 0),
			RequestBody: nil,
			Responses:   make(map[string]openapi.Response),
		}

		// Convert parameters
		for _, param := range operation.Parameters {
			endpoint.Parameters = append(endpoint.Parameters, p.convertParameter(param))
		}

		// Convert request body
		if operation.RequestBody != nil {
			endpoint.RequestBody = p.convertRequestBody(operation.RequestBody)
		}

		// Convert responses
		for statusCode, response := range operation.Responses.Map() {
			endpoint.Responses[statusCode] = p.convertResponse(response)
		}

		spec.Endpoints = append(spec.Endpoints, endpoint)
	}
}

// convertParameter converts an OpenAPI3 parameter to our internal representation
func (p *OpenAPIParser) convertParameter(param *openapi3.ParameterRef) openapi.Parameter {
	if param.Value == nil {
		return openapi.Parameter{}
	}

	return openapi.Parameter{
		Name:        param.Value.Name,
		In:          param.Value.In,
		Description: param.Value.Description,
		Required:    param.Value.Required,
		Schema:      p.convertSchema(param.Value.Schema),
	}
}

// convertRequestBody converts an OpenAPI3 request body to our internal representation
func (p *OpenAPIParser) convertRequestBody(body *openapi3.RequestBodyRef) *openapi.RequestBody {
	if body.Value == nil {
		return nil
	}

	return &openapi.RequestBody{
		Description: body.Value.Description,
		Required:    body.Value.Required,
		Content:     p.convertContent(body.Value.Content),
	}
}

// convertResponse converts an OpenAPI3 response to our internal representation
func (p *OpenAPIParser) convertResponse(response *openapi3.ResponseRef) openapi.Response {
	if response.Value == nil {
		return openapi.Response{}
	}

	description := ""
	if response.Value.Description != nil {
		description = *response.Value.Description
	}

	return openapi.Response{
		Description: description,
		Content:     p.convertContent(response.Value.Content),
	}
}

// convertContent converts OpenAPI3 content to our internal representation
func (p *OpenAPIParser) convertContent(content openapi3.Content) map[string]openapi.MediaType {
	result := make(map[string]openapi.MediaType)
	for mediaType, mediaTypeObj := range content {
		result[mediaType] = openapi.MediaType{
			Schema: p.convertSchema(mediaTypeObj.Schema),
		}
	}
	return result
}

// convertSchema converts an OpenAPI3 schema to our internal representation
func (p *OpenAPIParser) convertSchema(schema *openapi3.SchemaRef) openapi.Schema {
	if schema == nil || schema.Value == nil {
		return openapi.Schema{}
	}

	return openapi.Schema{
		Type:        schema.Value.Type,
		Format:      schema.Value.Format,
		Description: schema.Value.Description,
		Properties:  p.convertSchemaProperties(schema.Value.Properties),
		Required:    schema.Value.Required,
		Items: func() *openapi.Schema {
			if schema.Value.Items != nil {
				items := p.convertSchema(schema.Value.Items)
				return &items
			}
			return nil
		}(),
		Enum:    schema.Value.Enum,
		Default: schema.Value.Default,
		Minimum: schema.Value.Min,
		Maximum: schema.Value.Max,
		MinLength: func() *int {
			if schema.Value.MinLength > 0 {
				val := int(schema.Value.MinLength)
				return &val
			}
			return nil
		}(),
		MaxLength: func() *int {
			if schema.Value.MaxLength != nil && *schema.Value.MaxLength > 0 {
				val := int(*schema.Value.MaxLength)
				return &val
			}
			return nil
		}(),
		Pattern: schema.Value.Pattern,
	}
}

// convertSchemaProperties converts schema properties
func (p *OpenAPIParser) convertSchemaProperties(properties openapi3.Schemas) map[string]openapi.Schema {
	result := make(map[string]openapi.Schema)
	for name, schema := range properties {
		result[name] = p.convertSchema(schema)
	}
	return result
}

// convertComponents converts OpenAPI3 components to our internal representation
func (p *OpenAPIParser) convertComponents(components *openapi3.Components, spec *openapi.ParsedSpec) {
	// Convert schemas
	for name, schema := range components.Schemas {
		spec.Components[name] = openapi.Component{
			Type:   "schema",
			Schema: p.convertSchema(schema),
		}
	}
}
