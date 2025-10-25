package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"api-to-mcp/internal/config"
	"api-to-mcp/internal/server"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	port := flag.Int("port", 8080, "Server port")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override port if specified via flag
	if *port != 8080 {
		cfg.Server.Port = *port
	}

	// Create MCP server
	mcpServer, err := server.NewMCPServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down server...")
		cancel()
	}()

	// Start the server
	fmt.Printf("Starting API-to-MCP server on port %d\n", cfg.Server.Port)
	if err := mcpServer.Start(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
