# Go Development Best Practices

## Code Organization

### Package Structure
- **`cmd/`**: Application entry points
- **`internal/`**: Private application code
- **`pkg/`**: Public library code
- **`docs/`**: Documentation
- **`examples/`**: Example usage
- **`tests/`**: Test files

### Naming Conventions
- **Packages**: Lowercase, single word (e.g., `parser`, `generator`)
- **Files**: Snake_case (e.g., `openapi_parser.go`)
- **Types**: PascalCase (e.g., `MCPTool`, `OpenAPISpec`)
- **Functions**: PascalCase for public, camelCase for private
- **Variables**: camelCase

## Go-Specific Guidelines

### Error Handling
```go
// Good: Explicit error handling
result, err := parseOpenAPI(spec)
if err != nil {
    return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
}

// Bad: Ignoring errors
result, _ := parseOpenAPI(spec)
```

### Interface Design
```go
// Good: Small, focused interfaces
type OpenAPIParser interface {
    ParseSpec(filePath string) (*ParsedSpec, error)
    ValidateSpec(spec *ParsedSpec) error
}

// Bad: Large, unfocused interfaces
type Everything interface {
    ParseSpec(filePath string) (*ParsedSpec, error)
    GenerateTools(spec *ParsedSpec) []MCPTool
    StartServer(port int) error
    // ... many more methods
}
```

### Context Usage
```go
// Good: Use context for cancellation and timeouts
func (s *Server) HandleRequest(ctx context.Context, req *Request) (*Response, error) {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    return s.processRequest(ctx, req)
}
```

## Testing Standards

### Unit Tests
- **Coverage**: Aim for 80%+ test coverage
- **Naming**: `TestFunctionName_Scenario_ExpectedResult`
- **Structure**: Arrange, Act, Assert pattern

```go
func TestOpenAPIParser_ParseSpec_ValidFile_ReturnsParsedSpec(t *testing.T) {
    // Arrange
    parser := NewOpenAPIParser()
    specPath := "testdata/valid-spec.yaml"
    
    // Act
    result, err := parser.ParseSpec(specPath)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "Pet Store API", result.Info.Title)
}
```

### Integration Tests
- Test complete workflows
- Use real OpenAPI specifications
- Mock external dependencies

### Test Data
- Store test files in `testdata/` directory
- Use descriptive names for test cases
- Include both valid and invalid test data

## Documentation Standards

### Code Comments
```go
// ParseSpec parses an OpenAPI specification from the given file path.
// It supports both JSON and YAML formats and validates the structure.
// Returns an error if the file cannot be read or the spec is invalid.
func (p *OpenAPIParser) ParseSpec(filePath string) (*ParsedSpec, error) {
    // Implementation...
}
```

### README Files
- Include setup instructions
- Provide usage examples
- Document configuration options
- Include troubleshooting section

## Performance Guidelines

### Memory Management
- Use `sync.Pool` for frequently allocated objects
- Avoid unnecessary string allocations
- Use `strings.Builder` for string concatenation

### Concurrency
- Use goroutines for I/O operations
- Implement proper synchronization
- Use channels for communication
- Avoid goroutine leaks

### HTTP Client
- Use connection pooling
- Implement retry logic with exponential backoff
- Set appropriate timeouts
- Handle rate limiting

## Security Considerations

### Input Validation
- Validate all user inputs
- Sanitize file paths
- Check file sizes and types
- Use secure parsing libraries

### Configuration
- Never log sensitive information
- Use environment variables for secrets
- Validate configuration on startup
- Implement secure defaults

## Logging Standards

### Log Levels
- **DEBUG**: Detailed information for debugging
- **INFO**: General information about program flow
- **WARN**: Warning messages for recoverable issues
- **ERROR**: Error messages for unrecoverable issues

### Log Format
```go
log.WithFields(log.Fields{
    "component": "parser",
    "file": specPath,
    "error": err,
}).Error("Failed to parse OpenAPI specification")
```

## Configuration Management

### Environment Variables
- Use `API_TO_MCP_` prefix
- Document all configuration options
- Provide sensible defaults
- Validate on startup

### Configuration Files
- Support YAML and JSON formats
- Use hierarchical structure
- Implement hot reload
- Validate configuration schema
