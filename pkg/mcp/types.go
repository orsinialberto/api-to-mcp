package mcp

// Tool represents an MCP tool
type Tool struct {
	Name        string                                                   `json:"name"`
	Description string                                                   `json:"description"`
	InputSchema *InputSchema                                             `json:"inputSchema"`
	Handler     func(params map[string]interface{}) (interface{}, error) `json:"-"`
}

// InputSchema defines the input schema for a tool
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

// Property defines a property in the input schema
type Property struct {
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Format      string      `json:"format,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Minimum     *float64    `json:"minimum,omitempty"`
	Maximum     *float64    `json:"maximum,omitempty"`
	MinLength   *int        `json:"minLength,omitempty"`
	MaxLength   *int        `json:"maxLength,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
}

// Request represents a JSON-RPC request
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

// Response represents a JSON-RPC response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

// Error represents a JSON-RPC error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ListToolsRequest represents a request to list available tools
type ListToolsRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      string `json:"id"`
}

// ListToolsResponse represents the response to list tools
type ListToolsResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  struct {
		Tools []Tool `json:"tools"`
	} `json:"result"`
	ID string `json:"id"`
}

// CallToolRequest represents a request to call a tool
type CallToolRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	Method  string         `json:"method"`
	Params  CallToolParams `json:"params"`
	ID      string         `json:"id"`
}

// CallToolParams represents the parameters for calling a tool
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// CallToolResponse represents the response to a tool call
type CallToolResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	ID      string      `json:"id"`
}

// ServerInfo represents information about the MCP server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// NewError creates a new JSON-RPC error
func NewError(code int, message string, data interface{}) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// Standard JSON-RPC error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// MCP method names
const (
	MethodListTools = "tools/list"
	MethodCallTool  = "tools/call"
)
