package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	OpenAPI OpenAPIConfig `mapstructure:"openapi"`
	MCP     MCPConfig     `mapstructure:"mcp"`
	Filters FilterConfig  `mapstructure:"filters"`
	Logging LoggingConfig `mapstructure:"logging"`
}

// ServerConfig contains server-specific configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// OpenAPIConfig contains OpenAPI-specific configuration
type OpenAPIConfig struct {
	SpecPath string `mapstructure:"spec_path"`
	BaseURL  string `mapstructure:"base_url"`
}

// MCPConfig contains MCP-specific configuration
type MCPConfig struct {
	ServerName string `mapstructure:"server_name"`
	Version    string `mapstructure:"version"`
}

// FilterConfig contains filtering configuration
type FilterConfig struct {
	IncludePaths   []string `mapstructure:"include_paths"`
	ExcludePaths   []string `mapstructure:"exclude_paths"`
	IncludeMethods []string `mapstructure:"include_methods"`
	ExcludeMethods []string `mapstructure:"exclude_methods"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Set default values
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, use defaults
	}

	// Bind environment variables
	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("openapi.spec_path", "./examples/petstore.yaml")
	viper.SetDefault("openapi.base_url", "https://petstore3.swagger.io/api/v3")
	viper.SetDefault("mcp.server_name", "api-to-mcp")
	viper.SetDefault("mcp.version", "1.0.0")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if config.OpenAPI.SpecPath == "" {
		return fmt.Errorf("openapi.spec_path is required")
	}

	// Check if spec file exists
	if _, err := os.Stat(config.OpenAPI.SpecPath); os.IsNotExist(err) {
		return fmt.Errorf("openapi spec file not found: %s", config.OpenAPI.SpecPath)
	}

	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	return nil
}

// GetConfigPath returns the configuration file path
func GetConfigPath() string {
	configPath := os.Getenv("API_TO_MCP_CONFIG")
	if configPath != "" {
		return configPath
	}
	return "config.yaml"
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	config := `server:
  host: localhost
  port: 8080

openapi:
  spec_path: ./examples/petstore.yaml
  base_url: https://petstore3.swagger.io/api/v3

mcp:
  server_name: api-to-mcp
  version: 1.0.0

filters:
  include_paths: []
  exclude_paths: []
  include_methods: []
  exclude_methods: []

logging:
  level: info
  format: json
`

	return os.WriteFile(path, []byte(config), 0644)
}
