package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"api-to-mcp/internal/config"
	"api-to-mcp/internal/generator"
	"api-to-mcp/internal/parser"
	"api-to-mcp/pkg/mcp"

	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/sirupsen/logrus"
)

// MCPServer represents the MCP server
type MCPServer struct {
	config *config.Config
	tools  []mcp.Tool
	server *http.Server
	logger *logrus.Logger
}

// NewMCPServer creates a new MCP server
func NewMCPServer(cfg *config.Config) (*MCPServer, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	if cfg.Logging.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	// Parse OpenAPI specification
	openAPIParser := parser.NewOpenAPIParser(cfg.OpenAPI.SpecPath, logger)
	spec, err := openAPIParser.ParseSpec()
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	// Generate MCP tools
	toolGenerator := generator.NewMCPToolGenerator(spec, cfg, logger)
	tools, err := toolGenerator.GenerateTools()
	if err != nil {
		return nil, fmt.Errorf("failed to generate MCP tools: %w", err)
	}

	// Create JSON-RPC server
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(json.NewCodec(), "application/json")

	// Register MCP service
	mcpService := NewMCPService(tools, cfg, logger)
	rpcServer.RegisterService(mcpService, "")

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      rpcServer,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &MCPServer{
		config: cfg,
		tools:  tools,
		server: httpServer,
		logger: logger,
	}, nil
}

// Start starts the MCP server
func (s *MCPServer) Start(ctx context.Context) error {
	s.logger.WithFields(logrus.Fields{
		"host": s.config.Server.Host,
		"port": s.config.Server.Port,
	}).Info("Starting MCP server")

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	s.logger.Info("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.WithError(err).Error("Server shutdown failed")
		return err
	}

	s.logger.Info("Server shutdown complete")
	return nil
}

// GetTools returns the list of available tools
func (s *MCPServer) GetTools() []mcp.Tool {
	return s.tools
}

// GetToolByName returns a tool by name
func (s *MCPServer) GetToolByName(name string) (*mcp.Tool, error) {
	for _, tool := range s.tools {
		if tool.Name == name {
			return &tool, nil
		}
	}
	return nil, fmt.Errorf("tool not found: %s", name)
}
