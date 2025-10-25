# OpenAPI Parsing

This document describes the OpenAPI parsing functionality implemented in the API-to-MCP server.

## Overview

The OpenAPI parser is responsible for:
- Loading and parsing OpenAPI 3.0/3.1 and Swagger 2.0 specifications
- Converting OpenAPI schemas to internal representations
- Validating specification structure and content
- Extracting endpoints, parameters, and components

## Architecture

```
OpenAPI Spec File → Parser → Validation → Internal Representation
```

### Components

1. **OpenAPIParser** - Main parser class
2. **Validator** - Specification validation
3. **Type Converters** - OpenAPI to internal type conversion

## Supported Formats

### OpenAPI Versions
- **OpenAPI 3.0** - Full support
- **OpenAPI 3.1** - Full support  
- **Swagger 2.0** - Basic support (via conversion)

### File Formats
- **YAML** - Primary format
- **JSON** - Supported via conversion

## Parser Implementation

### Core Functionality

```go
type OpenAPIParser struct {
    specPath string
    logger   *logrus.Logger
}

func (p *OpenAPIParser) ParseSpec() (*openapi.ParsedSpec, error)
```

### Parsing Process

1. **File Loading**
   - Check file existence
   - Load file content
   - Handle encoding issues

2. **OpenAPI Validation**
   - Use kin-openapi library
   - Validate against OpenAPI schema
   - Report validation errors

3. **Internal Conversion**
   - Convert OpenAPI types to internal types
   - Handle type mappings
   - Preserve metadata

4. **Custom Validation**
   - Validate business rules
   - Check required fields
   - Verify constraints

## Type Conversion

### OpenAPI → Internal Types

| OpenAPI Type | Internal Type | Notes |
|--------------|---------------|-------|
| `string` | `string` | Direct mapping |
| `integer` | `integer` | Direct mapping |
| `number` | `number` | Direct mapping |
| `boolean` | `boolean` | Direct mapping |
| `array` | `array` | With item schema |
| `object` | `object` | With properties |

### Schema Constraints

```go
type Schema struct {
    Type        string             `json:"type"`
    Format      string             `json:"format"`
    Description string             `json:"description"`
    Properties  map[string]Schema  `json:"properties,omitempty"`
    Required    []string           `json:"required,omitempty"`
    Items       *Schema            `json:"items,omitempty"`
    Enum        []interface{}      `json:"enum,omitempty"`
    Default     interface{}        `json:"default,omitempty"`
    Minimum     *float64           `json:"minimum,omitempty"`
    Maximum     *float64           `json:"maximum,omitempty"`
    MinLength   *int               `json:"minLength,omitempty"`
    MaxLength   *int               `json:"maxLength,omitempty"`
    Pattern     string             `json:"pattern,omitempty"`
}
```

## Validation

### Built-in Validation

The parser uses kin-openapi for basic OpenAPI validation:
- Schema compliance
- Reference resolution
- Format validation

### Custom Validation

Additional validation rules:

#### Info Section
- Title is required
- Version is required
- Description is optional

#### Paths Section
- At least one endpoint required
- Valid HTTP methods only
- Path parameters must be valid
- Responses must be defined

#### Parameters
- Name is required
- Location must be valid (path, query, header, cookie)
- Schema must be valid
- Constraints must be consistent

#### Schemas
- Type must be valid
- Constraints must be logical (min ≤ max)
- References must be resolvable

## Error Handling

### Error Types

1. **File Errors**
   - File not found
   - Permission denied
   - Invalid encoding

2. **Parse Errors**
   - Invalid YAML/JSON
   - Malformed structure
   - Missing required fields

3. **Validation Errors**
   - Schema violations
   - Constraint violations
   - Reference errors

### Error Reporting

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}
```

## Usage Examples

### Basic Parsing

```go
logger := logrus.New()
parser := NewOpenAPIParser("spec.yaml", logger)

spec, err := parser.ParseSpec()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("API: %s v%s\n", spec.Info.Title, spec.Info.Version)
fmt.Printf("Endpoints: %d\n", len(spec.Endpoints))
```

### With Validation

```go
spec, err := parser.ParseSpec()
if err != nil {
    // Handle parsing errors
    return err
}

// Additional validation
validator := NewValidator(logger)
if err := validator.ValidateSpec(spec); err != nil {
    // Handle validation errors
    return err
}
```

## Performance Considerations

### Memory Usage
- Large specifications are loaded entirely into memory
- Consider streaming for very large specs
- Component references are resolved eagerly

### Parsing Speed
- YAML parsing is generally faster than JSON
- Complex schemas with many references are slower
- Validation adds overhead but improves reliability

## Testing

### Unit Tests
- Test individual parser functions
- Mock OpenAPI documents
- Verify type conversions

### Integration Tests
- Test with real OpenAPI specifications
- Verify end-to-end parsing
- Test error scenarios

### Test Data
- Use `testdata/` directory for test files
- Include both valid and invalid examples
- Cover edge cases and error conditions

## Configuration

### Parser Options

```go
type ParserConfig struct {
    StrictMode    bool   // Enable strict validation
    ResolveRefs   bool   // Resolve $ref references
    ValidateSpec  bool   // Enable custom validation
    LogLevel      string // Logging level
}
```

### Environment Variables

- `API_TO_MCP_STRICT_MODE` - Enable strict validation
- `API_TO_MCP_LOG_LEVEL` - Set logging level
- `API_TO_MCP_VALIDATE_SPEC` - Enable custom validation

## Troubleshooting

### Common Issues

1. **File Not Found**
   - Check file path
   - Verify file permissions
   - Ensure file exists

2. **Invalid YAML**
   - Check YAML syntax
   - Verify indentation
   - Use YAML validator

3. **Validation Errors**
   - Check required fields
   - Verify constraint logic
   - Review OpenAPI specification

4. **Type Conversion Errors**
   - Check OpenAPI version compatibility
   - Verify schema definitions
   - Review type mappings

### Debug Mode

Enable debug logging to see detailed parsing information:

```go
logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)
parser := NewOpenAPIParser(specPath, logger)
```
