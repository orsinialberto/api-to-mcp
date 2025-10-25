# System Architecture Overview

## High-Level Architecture

The API-to-MCP server follows a modular architecture designed for extensibility and maintainability.

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   OpenAPI Spec  │───▶│     Parser      │───▶│   Tool Generator│───▶│  MCP Server     │
│   (JSON/YAML)   │    │   (kin-openapi) │    │   (Custom)      │    │  (JSON-RPC)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │                       │
                                                       ▼                       ▼
                                               ┌─────────────────┐    ┌─────────────────┐
                                               │   HTTP Client   │    │   MCP Tools     │
                                               │   (go-resty)    │    │   (Generated)   │
                                               └─────────────────┘    └─────────────────┘
```

## Core Components

### 1. Parser Module (`internal/parser/`)
- **Responsibility**: Parse and validate OpenAPI specifications
- **Input**: OpenAPI JSON/YAML files
- **Output**: Structured representation of API endpoints
- **Key Features**:
  - Support for OpenAPI 3.0/3.1 and Swagger 2.0
  - Parameter extraction and validation
  - Schema analysis and type mapping

### 2. Tool Generator (`internal/generator/`)
- **Responsibility**: Convert OpenAPI endpoints to MCP tools
- **Input**: Parsed OpenAPI data
- **Output**: MCP tool definitions
- **Key Features**:
  - Automatic tool naming and description generation
  - Parameter type mapping (OpenAPI → MCP)
  - Schema validation and transformation

### 3. MCP Server (`internal/server/`)
- **Responsibility**: Expose MCP tools via JSON-RPC
- **Input**: Generated MCP tools
- **Output**: JSON-RPC API
- **Key Features**:
  - JSON-RPC 2.0 protocol support
  - Dynamic tool registration
  - Request/response handling

### 4. HTTP Client (`internal/utils/`)
- **Responsibility**: Make HTTP calls to target APIs
- **Input**: MCP tool parameters
- **Output**: API responses
- **Key Features**:
  - Authentication support (API keys, Bearer tokens)
  - Error handling and retry logic
  - Response transformation

## Data Flow

1. **Initialization**:
   - Load OpenAPI specification
   - Parse and validate structure
   - Generate MCP tools
   - Start JSON-RPC server

2. **Runtime**:
   - Client sends JSON-RPC request
   - Server routes to appropriate MCP tool
   - Tool parameters mapped to HTTP request
   - HTTP client calls target API
   - Response transformed and returned

## Configuration System

The system uses a hierarchical configuration approach:
- **Default values** in code
- **Configuration files** (YAML/JSON)
- **Environment variables**
- **Command-line flags**

## Error Handling Strategy

- **Parsing errors**: Graceful degradation with detailed error messages
- **HTTP errors**: Retry logic with exponential backoff
- **JSON-RPC errors**: Standard error codes and messages
- **Validation errors**: Clear parameter validation feedback

## Performance Considerations

- **Lazy loading**: Tools generated on-demand
- **Caching**: Response caching for frequently accessed endpoints
- **Connection pooling**: HTTP client connection reuse
- **Concurrent processing**: Goroutine-based request handling
