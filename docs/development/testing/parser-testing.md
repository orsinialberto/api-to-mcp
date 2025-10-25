# Parser Testing Strategy

This document outlines the testing strategy for the OpenAPI parser component.

## Testing Overview

The parser testing is organized into three levels:
1. **Unit Tests** - Individual function testing
2. **Integration Tests** - End-to-end parsing
3. **Performance Tests** - Load and stress testing

## Unit Testing

### Test Structure

```go
func TestFunctionName_Scenario_ExpectedResult(t *testing.T) {
    // Arrange
    // Act  
    // Assert
}
```

### Test Categories

#### 1. Parser Creation
- Test parser initialization
- Verify logger assignment
- Check spec path setting

#### 2. File Operations
- Test file existence checking
- Test file reading
- Test error handling for missing files

#### 3. OpenAPI Parsing
- Test valid OpenAPI specifications
- Test invalid specifications
- Test malformed YAML/JSON

#### 4. Type Conversion
- Test parameter conversion
- Test schema conversion
- Test constraint handling

#### 5. Validation
- Test validation rules
- Test error reporting
- Test constraint validation

### Test Data Management

#### Test Files
Store test OpenAPI specifications in `testdata/` directory:

```
testdata/
├── valid-spec.yaml          # Valid OpenAPI spec
├── invalid-spec.yaml        # Invalid OpenAPI spec
├── malformed-yaml.yaml      # Malformed YAML
├── missing-fields.yaml      # Missing required fields
└── constraint-violations.yaml # Constraint violations
```

#### Test Utilities

```go
func createTempSpec(t *testing.T, content string) string {
    tempDir := t.TempDir()
    specPath := filepath.Join(tempDir, "test-spec.yaml")
    
    err := os.WriteFile(specPath, []byte(content), 0644)
    require.NoError(t, err)
    
    return specPath
}
```

## Integration Testing

### Test Scenarios

#### 1. Real OpenAPI Specifications
- Test with Pet Store API
- Test with GitHub API
- Test with Stripe API

#### 2. Complex Specifications
- Multiple endpoints
- Nested schemas
- Reference resolution
- Multiple servers

#### 3. Error Scenarios
- Non-existent files
- Invalid formats
- Validation failures

### Integration Test Structure

```go
func TestIntegration_ParseRealSpec(t *testing.T) {
    // Use real OpenAPI specification
    specPath := "../../examples/petstore.yaml"
    
    // Check if file exists
    if _, err := os.Stat(specPath); os.IsNotExist(err) {
        t.Skip("Spec file not found")
    }
    
    // Parse and validate
    parser := NewOpenAPIParser(specPath, logger)
    spec, err := parser.ParseSpec()
    
    require.NoError(t, err)
    assert.NotNil(t, spec)
    // Additional assertions...
}
```

## Performance Testing

### Load Testing

```go
func TestPerformance_ParseLargeSpec(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test in short mode")
    }
    
    // Create large OpenAPI specification
    largeSpec := generateLargeSpec(1000) // 1000 endpoints
    
    start := time.Now()
    spec, err := parser.ParseSpec()
    duration := time.Since(start)
    
    require.NoError(t, err)
    assert.Less(t, duration, 5*time.Second) // Should parse in < 5s
}
```

### Memory Testing

```go
func TestMemory_ParseLargeSpec(t *testing.T) {
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Parse large specification
    spec, err := parser.ParseSpec()
    require.NoError(t, err)
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    memoryUsed := m2.Alloc - m1.Alloc
    assert.Less(t, memoryUsed, uint64(50*1024*1024)) // < 50MB
}
```

## Test Coverage

### Coverage Goals
- **Unit Tests**: 90%+ coverage
- **Integration Tests**: Cover all major scenarios
- **Error Cases**: Test all error conditions

### Coverage Measurement

```bash
# Run tests with coverage
go test -cover ./internal/parser/...

# Generate coverage report
go test -coverprofile=coverage.out ./internal/parser/...
go tool cover -html=coverage.out
```

## Test Data Generation

### Valid Specifications

```go
func generateValidSpec() string {
    return `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: Success`
}
```

### Invalid Specifications

```go
func generateInvalidSpec() string {
    return `openapi: 3.0.0
info:
  # Missing title
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: Success`
}
```

## Mocking and Stubs

### Logger Mocking

```go
type MockLogger struct {
    entries []logrus.Entry
}

func (m *MockLogger) WithFields(fields logrus.Fields) *logrus.Entry {
    return &logrus.Entry{}
}

func (m *MockLogger) Info(args ...interface{}) {
    // Mock implementation
}
```

### File System Mocking

```go
type MockFileSystem struct {
    files map[string][]byte
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
    if _, exists := m.files[name]; !exists {
        return nil, os.ErrNotExist
    }
    return &mockFileInfo{}, nil
}
```

## Test Organization

### File Structure

```
internal/parser/
├── parser.go              # Main parser code
├── parser_test.go         # Unit tests
├── validation.go          # Validation code
├── validation_test.go     # Validation tests
└── testdata/              # Test data files
    ├── valid-spec.yaml
    ├── invalid-spec.yaml
    └── complex-spec.yaml

tests/
├── integration_test.go    # Integration tests
└── performance_test.go    # Performance tests
```

### Test Naming

- **Unit Tests**: `TestFunctionName_Scenario_ExpectedResult`
- **Integration Tests**: `TestIntegration_Scenario`
- **Performance Tests**: `TestPerformance_Scenario`

## Continuous Integration

### GitHub Actions

```yaml
name: Parser Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go test -v ./internal/parser/...
      - run: go test -v ./tests/...
      - run: go test -cover ./internal/parser/...
```

### Test Commands

```bash
# Run all tests
go test ./...

# Run parser tests only
go test ./internal/parser/...

# Run integration tests
go test ./tests/...

# Run with coverage
go test -cover ./internal/parser/...

# Run performance tests
go test -run=TestPerformance ./tests/...

# Run tests with race detection
go test -race ./internal/parser/...
```

## Best Practices

### Test Design
1. **Arrange-Act-Assert** pattern
2. **One assertion per test** when possible
3. **Descriptive test names**
4. **Independent tests** (no dependencies)

### Test Data
1. **Use realistic data**
2. **Include edge cases**
3. **Test both valid and invalid inputs**
4. **Use table-driven tests** for similar scenarios

### Error Testing
1. **Test all error conditions**
2. **Verify error messages**
3. **Test error recovery**
4. **Test timeout scenarios**

### Performance
1. **Set performance benchmarks**
2. **Test memory usage**
3. **Test with large datasets**
4. **Monitor test execution time**
