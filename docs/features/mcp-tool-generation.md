# MCP Tool Generation

## Overview

Phase 3 of the API-to-MCP project focuses on generating MCP (Model Context Protocol) tools from OpenAPI specifications. This phase converts OpenAPI endpoints into MCP tools that can be used by AI assistants and other MCP clients.

## Features Implemented

### 1. MCP Tool Structure

The generator creates MCP tools with the following structure:

```go
type Tool struct {
    Name        string                                                   `json:"name"`
    Description string                                                   `json:"description"`
    InputSchema *InputSchema                                             `json:"inputSchema"`
    Handler     func(params map[string]interface{}) (interface{}, error) `json:"-"`
}
```

### 2. Endpoint to Tool Mapping

- **Tool Naming**: Automatically generates tool names from OpenAPI operation IDs or endpoint paths
- **Description Generation**: Uses OpenAPI summary, description, or fallback to method + path
- **Input Schema**: Converts OpenAPI parameters and request body schemas to MCP input schemas

### 3. Type Conversion

The generator maps OpenAPI types to MCP types:

| OpenAPI Type | MCP Type | Notes |
|--------------|----------|-------|
| string       | string   | Direct mapping |
| integer      | integer  | Direct mapping |
| number       | number   | Direct mapping |
| boolean      | boolean  | Direct mapping |
| array        | array    | With item type information |
| object       | object   | With property information |

### 4. Parameter Handling

#### Path Parameters
- Automatically detected from endpoint paths (e.g., `/users/{id}`)
- Marked as required in the input schema
- Properly typed based on OpenAPI schema

#### Query Parameters
- Optional parameters with default values
- Type constraints (minimum, maximum, pattern, etc.)
- Enum values for string parameters

#### Request Body Parameters
- Full schema parsing for JSON request bodies
- Nested object support
- Array parameter handling
- Required field validation

### 5. Advanced Schema Support

#### Nested Objects
- Recursive schema parsing for complex objects
- Property information in descriptions
- Proper type mapping for nested structures

#### Arrays
- Item type detection and description
- Support for arrays of primitives and objects
- Proper validation constraints

#### Constraints and Validation
- String length constraints (minLength, maxLength)
- Numeric range constraints (minimum, maximum)
- Pattern matching for strings
- Enum value support
- Default value handling

### 6. Error Handling and Validation

#### Input Validation
- Validates generator inputs (spec, config, logger)
- Checks for required fields and proper configuration
- Validates base URL and other critical settings

#### Tool Validation
- Ensures generated tools have all required fields
- Validates input schemas for consistency
- Checks property constraints for logical consistency

#### Error Recovery
- Continues processing even if some tools fail to generate
- Provides detailed error logging
- Returns partial results with error information

### 7. Filtering Support

#### Path Filtering
- Include/exclude specific paths
- Pattern-based filtering
- Supports multiple patterns

#### Method Filtering
- Include/exclude specific HTTP methods
- Case-insensitive matching
- Supports multiple methods

## Usage Examples

### Basic Tool Generation

```go
// Parse OpenAPI specification
parser := parser.NewOpenAPIParser("spec.yaml", logger)
spec, err := parser.ParseSpec()

// Create generator
config := &config.Config{
    OpenAPI: config.OpenAPIConfig{
        BaseURL: "https://api.example.com",
    },
    Filters: config.FilterConfig{},
}

generator := generator.NewMCPToolGenerator(spec, config, logger)

// Generate tools
tools, err := generator.GenerateTools()
```

### Filtered Tool Generation

```go
config := &config.Config{
    OpenAPI: config.OpenAPIConfig{
        BaseURL: "https://api.example.com",
    },
    Filters: config.FilterConfig{
        IncludePaths:   []string{"/api/v1"},
        ExcludePaths:   []string{"/admin"},
        IncludeMethods: []string{"GET", "POST"},
    },
}
```

## Testing

### Unit Tests
- Comprehensive test coverage for all generator functions
- Edge case testing for error conditions
- Validation testing for input schemas and properties

### Integration Tests
- Real OpenAPI specification testing
- Petstore API integration test
- Complex schema testing with nested objects

### Test Coverage
- Generator creation and configuration
- Tool generation from various endpoint types
- Parameter and request body handling
- Error handling and validation
- Filtering functionality

## Performance Considerations

- Efficient schema parsing with minimal memory allocation
- Lazy evaluation of complex schemas
- Proper error handling without performance impact
- Logging at appropriate levels to avoid overhead

## Future Enhancements

- Schema reference resolution ($ref support)
- More sophisticated nested object handling
- Custom tool naming strategies
- Advanced filtering options
- Schema validation improvements

## Related Documentation

- [OpenAPI Parsing](../features/openapi-parsing.md)
- [Architecture Overview](../architecture/overview.md)
- [Development Best Practices](../development/best-practices.md)
