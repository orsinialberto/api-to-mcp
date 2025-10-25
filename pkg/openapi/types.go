package openapi

// ParsedSpec represents a parsed OpenAPI specification
type ParsedSpec struct {
	Info       Info                 `json:"info"`
	Servers    []Server             `json:"servers"`
	Endpoints  []Endpoint           `json:"endpoints"`
	Components map[string]Component `json:"components"`
}

// Info represents the API information
type Info struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// Server represents a server configuration
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

// Endpoint represents an API endpoint
type Endpoint struct {
	Path        string              `json:"path"`
	Method      string              `json:"method"`
	OperationID string              `json:"operationId"`
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Parameters  []Parameter         `json:"parameters"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

// Parameter represents a parameter
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
}

// RequestBody represents a request body
type RequestBody struct {
	Description string               `json:"description"`
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

// Response represents a response
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content"`
}

// MediaType represents a media type
type MediaType struct {
	Schema Schema `json:"schema"`
}

// Schema represents a schema
type Schema struct {
	Type        string            `json:"type"`
	Format      string            `json:"format"`
	Description string            `json:"description"`
	Properties  map[string]Schema `json:"properties,omitempty"`
	Required    []string          `json:"required,omitempty"`
	Items       *Schema           `json:"items,omitempty"`
	Enum        []interface{}     `json:"enum,omitempty"`
	Default     interface{}       `json:"default,omitempty"`
	Minimum     *float64          `json:"minimum,omitempty"`
	Maximum     *float64          `json:"maximum,omitempty"`
	MinLength   *int              `json:"minLength,omitempty"`
	MaxLength   *int              `json:"maxLength,omitempty"`
	Pattern     string            `json:"pattern,omitempty"`
}

// Component represents a reusable component
type Component struct {
	Type   string `json:"type"`
	Schema Schema `json:"schema"`
}
