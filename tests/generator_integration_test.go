package tests

import (
	"os"
	"testing"

	"api-to-mcp/internal/config"
	"api-to-mcp/internal/generator"
	"api-to-mcp/internal/parser"
	"api-to-mcp/pkg/mcp"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratorWithPetStoreSpec(t *testing.T) {
	// Use the example petstore spec
	specPath := "../examples/petstore.yaml"

	// Check if file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("Petstore spec not found, skipping integration test")
	}

	logger := logrus.New()
	openAPIParser := parser.NewOpenAPIParser(specPath, logger)

	spec, err := openAPIParser.ParseSpec()
	require.NoError(t, err)

	// Create generator
	cfg := &config.Config{
		OpenAPI: config.OpenAPIConfig{
			BaseURL: "https://petstore3.swagger.io/api/v3",
		},
		Filters: config.FilterConfig{},
	}

	gen := generator.NewMCPToolGenerator(spec, cfg, logger)

	// Generate tools
	tools, err := gen.GenerateTools()
	require.NoError(t, err)
	assert.Greater(t, len(tools), 0)

	// Verify we have some expected tools
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
		assert.NotEmpty(t, tool.Name)
		assert.NotEmpty(t, tool.Description)
		assert.NotNil(t, tool.InputSchema)
		assert.NotNil(t, tool.Handler)
	}

	// Check for some expected petstore endpoints
	expectedTools := []string{"addpet", "updatepet", "findpetsbystatus", "getpetbyid", "deletepet"}
	foundExpected := 0
	for _, expected := range expectedTools {
		if toolNames[expected] {
			foundExpected++
		}
	}
	assert.Greater(t, foundExpected, 0, "Should have found some expected petstore tools")

	// Test a specific tool (addPet)
	if toolNames["addpet"] {
		var addPetTool *mcp.Tool
		for _, tool := range tools {
			if tool.Name == "addpet" {
				addPetTool = &tool
				break
			}
		}
		require.NotNil(t, addPetTool)

		// Verify the tool has request body properties
		schema := addPetTool.InputSchema
		assert.Equal(t, "object", schema.Type)
		assert.Greater(t, len(schema.Properties), 0, "Should have request body properties")

		// Check for some expected pet properties
		expectedProperties := []string{"name", "status", "category", "tags"}
		foundProperties := 0
		for _, prop := range expectedProperties {
			if _, exists := schema.Properties[prop]; exists {
				foundProperties++
			}
		}
		assert.Greater(t, foundProperties, 0, "Should have found some expected pet properties")
	}

	// Test error handling with invalid config
	invalidConfig := &config.Config{
		OpenAPI: config.OpenAPIConfig{
			BaseURL: "", // Empty base URL should cause validation error
		},
		Filters: config.FilterConfig{},
	}

	invalidGen := generator.NewMCPToolGenerator(spec, invalidConfig, logger)
	_, err = invalidGen.GenerateTools()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base URL is required")
}
