# API-to-MCP Server - Development Process

## Project Overview

API-to-MCP is a Go-based server that automatically converts OpenAPI/Swagger specifications into MCP (Model Context Protocol) tools exposed via JSON-RPC. The server analyzes OpenAPI schemas and dynamically generates tools that allow interaction with existing REST APIs through the MCP protocol.

### Technology Stack

**Backend:**
- Go 1.21+
- kin-openapi for OpenAPI parsing
- gorilla/rpc for JSON-RPC server
- go-resty for HTTP client
- viper for configuration
- logrus for logging

**Protocols:**
- JSON-RPC 2.0 for tool exposure
- HTTP/HTTPS for API communication
- OpenAPI 3.0/3.1 specification support

**Configuration:**
- YAML/JSON configuration files
- Environment variable support
- Hot reload capability

### Project Structure

```
api-to-mcp/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point
├── internal/
│   ├── parser/                    # OpenAPI parsing
│   │   ├── parser.go
│   │   └── types.go
│   ├── generator/                  # MCP tools generation
│   │   ├── generator.go
│   │   └── mapper.go
│   ├── server/                     # JSON-RPC server
│   │   ├── server.go
│   │   └── handlers.go
│   ├── config/                     # Configuration
│   │   └── config.go
│   └── utils/                      # Utilities
│       ├── http.go
│       └── validation.go
├── pkg/
│   ├── mcp/                        # MCP protocol types
│   │   ├── types.go
│   │   └── protocol.go
│   └── openapi/                    # OpenAPI types
│       └── types.go
├── docs/                           # Documentation
│   ├── architecture/
│   │   ├── overview.md
│   │   └── diagrams.md
│   ├── development/
│   │   ├── best-practices.md
│   │   ├── phases/
│   │   └── testing/
│   ├── features/
│   ├── integrations/
│   ├── troubleshooting/
│   └── README.md
├── examples/                       # Example OpenAPI specs
├── tests/
├── go.mod
├── go.sum
├── README.md
├── SPECS.md                        # Technical specifications
└── AGENTS.md                       # Development process (this file)
```

## Development Process

### Standard Workflow for Each Feature

1. **Development**
   - Write code following specifications
   - Verify it works correctly
   - Follow Go best practices and conventions

2. **Testing**
   - Write unit tests for the code
   - Write integration tests where needed
   - Run tests and verify they all pass
   - Ensure test coverage is adequate

3. **Manual Testing**
   - Stop and wait for user to perform manual tests
   - User verifies everything works as expected
   - Test with real OpenAPI specifications
   - Verify JSON-RPC responses are correct

4. **Documentation**
   - Write documentation for the feature
   - Update existing documentation files
   - Update SPECS.md if needed
   - Update /docs if needed
   - Add examples if applicable

5. **Commit**
   - Stop and wait for user to say proceed
   - Commit changes with descriptive message

### Behavior Rules

- **One step at a time**: Complete each phase before moving to the next
- **Wait for confirmation**: Don't proceed without user's OK
- **Working code**: Every commit must contain tested and working code
- **Updated documentation**: Always keep documentation synchronized
- **Continuous testing**: Verify everything works before proceeding
- **Go conventions**: Follow standard Go project structure and naming
- **Error handling**: Implement proper error handling and logging
- **Configuration first**: Always make features configurable

### Development Phases

The development follows the phases outlined in SPECS.md:

1. **Setup and Foundation** - Project initialization and basic structure
2. **Parser OpenAPI** - OpenAPI specification parsing and validation
3. **Generation Tools MCP** - Converting OpenAPI endpoints to MCP tools
4. **Server JSON-RPC** - JSON-RPC server implementation
5. **HTTP Client and Integration** - HTTP client for API calls
6. **Configuration and Filters** - Configuration system and filtering
7. **Testing and Validation** - Comprehensive testing suite
8. **Documentation and Deployment** - Documentation and deployment setup
9. **Optimizations and Advanced Features** - Performance and advanced features
10. **Release and Distribution** - Release preparation and distribution

### Key Development Principles

- **Modularity**: Each component should be independent and testable
- **Configuration**: All behavior should be configurable
- **Error Handling**: Robust error handling and logging
- **Performance**: Efficient parsing and tool generation
- **Extensibility**: Easy to add new features and integrations
- **Testing**: Comprehensive test coverage
- **Documentation**: Clear and up-to-date documentation
