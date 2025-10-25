package server

import (
	"fmt"
	"net/http"

	"api-to-mcp/internal/config"
	"api-to-mcp/pkg/mcp"

	"github.com/sirupsen/logrus"
)

// MCPService handles MCP protocol requests
type MCPService struct {
	tools  []mcp.Tool
	config *config.Config
	logger *logrus.Logger
}

// NewMCPService creates a new MCP service
func NewMCPService(tools []mcp.Tool, cfg *config.Config, logger *logrus.Logger) *MCPService {
	return &MCPService{
		tools:  tools,
		config: cfg,
		logger: logger,
	}
}

// ListTools handles the tools/list request
func (s *MCPService) ListTools(r *http.Request, args *struct{}, reply *mcp.ListToolsResponse) error {
	s.logger.Debug("Handling tools/list request")

	// Create response
	reply.JSONRPC = "2.0"
	reply.Result.Tools = s.tools
	reply.ID = "1" // TODO: Extract ID from request

	s.logger.WithField("tool_count", len(s.tools)).Info("Listed available tools")
	return nil
}

// CallTool handles the tools/call request
func (s *MCPService) CallTool(r *http.Request, args *mcp.CallToolParams, reply *mcp.CallToolResponse) error {
	s.logger.WithFields(logrus.Fields{
		"tool_name": args.Name,
		"arguments": args.Arguments,
	}).Debug("Handling tools/call request")

	// Find the tool
	var tool *mcp.Tool
	for _, t := range s.tools {
		if t.Name == args.Name {
			tool = &t
			break
		}
	}

	if tool == nil {
		reply.JSONRPC = "2.0"
		reply.Result = mcp.NewError(mcp.MethodNotFound, fmt.Sprintf("Tool not found: %s", args.Name), nil)
		reply.ID = "1" // TODO: Extract ID from request
		return nil
	}

	// Execute the tool
	result, err := tool.Handler(args.Arguments)
	if err != nil {
		s.logger.WithError(err).Error("Tool execution failed")
		reply.JSONRPC = "2.0"
		reply.Result = mcp.NewError(mcp.InternalError, fmt.Sprintf("Tool execution failed: %v", err), nil)
		reply.ID = "1" // TODO: Extract ID from request
		return nil
	}

	// Return success response
	reply.JSONRPC = "2.0"
	reply.Result = result
	reply.ID = "1" // TODO: Extract ID from request

	s.logger.WithField("tool_name", args.Name).Info("Tool executed successfully")
	return nil
}
