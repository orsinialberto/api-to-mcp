# API-to-MCP Server

A Go-based server that automatically converts OpenAPI/Swagger specifications into MCP (Model Context Protocol) tools exposed via JSON-RPC.

## Overview

API-to-MCP analyzes OpenAPI schemas and dynamically generates tools that allow interaction with existing REST APIs through the MCP protocol. This enables seamless integration of REST APIs with AI assistants and other MCP-compatible clients.

## Features

- **OpenAPI Parsing**: Supports OpenAPI 3.0/3.1 and Swagger 2.0 specifications
- **Automatic Tool Generation**: Converts REST endpoints to MCP tools
- **JSON-RPC Server**: Exposes tools via JSON-RPC 2.0 protocol
- **Flexible Configuration**: YAML/JSON configuration with environment variable support
- **Filtering**: Include/exclude endpoints and HTTP methods
- **Authentication**: Support for API keys, Bearer tokens, and basic auth
- **Error Handling**: Comprehensive error handling and logging

## Quick Start

### Prerequisites

- Go 1.21 or later
- An OpenAPI/Swagger specification file

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd api-to-mcp
```

2. Install dependencies:
```bash
go mod tidy
```

3. Create a configuration file:
```bash
cp config.example.yaml config.yaml
```

4. Update the configuration with your OpenAPI spec path and API base URL.

5. Run the server:
```bash
# Using default configuration
go run cmd/server/main.go

# Using custom configuration
go run cmd/server/main.go -config /path/to/config.yaml

# Using custom port
go run cmd/server/main.go -port 8081
```

## Usage

### JSON-RPC API

The server exposes the following JSON-RPC methods:

#### List Tools
```json
{
  "jsonrpc": "2.0",
  "method": "tools/list",
  "id": "1"
}
```

#### Call Tool
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "get_pet_by_id",
    "arguments": {
      "petId": 123
    }
  },
  "id": "1"
}
```

## Project Structure

```
api-to-mcp/
├── cmd/server/          # Application entry point
├── internal/            # Private application code
│   ├── parser/         # OpenAPI parsing
│   ├── generator/      # MCP tools generation
│   ├── server/         # JSON-RPC server
│   ├── config/         # Configuration
│   └── utils/          # Utilities
├── pkg/                # Public library code
│   ├── mcp/            # MCP protocol types
│   └── openapi/        # OpenAPI types
├── docs/               # Documentation
├── examples/           # Example OpenAPI specs
└── tests/             # Test files
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/api-to-mcp cmd/server/main.go
```

### Docker

```bash
docker build -t api-to-mcp .
docker run -p 8080:8080 api-to-mcp
```

## Examples

See the `examples/` directory for sample OpenAPI specifications and usage examples.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Support

For issues and questions, please open an issue on GitHub.
