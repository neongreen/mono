package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/neongreen/mono/conf/pkg/config"
)

func TestClaudeTool_Integration(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create conf config directory
	configDir := filepath.Join(tmpDir, ".config", "conf")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	claudeConfigPath := filepath.Join(tmpDir, ".config", "claude", "config.json")

	// Create conf config file with Claude tool configuration
	confConfigPath := filepath.Join(configDir, "config.toml")
	tomlData := `[tools.claude]
name = "claude"
path = "` + claudeConfigPath + `"
`
	if err := os.WriteFile(confConfigPath, []byte(tomlData), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create Claude tool instance
	tool, err := NewClaudeTool()
	if err != nil {
		t.Fatalf("Failed to create Claude tool: %v", err)
	}

	// Test setting various configuration values
	testCases := []struct {
		path  string
		value interface{}
	}{
		{"model", "claude-3-5-sonnet-20241022"},
		{"max_tokens", float64(4096)},
		{"temperature", float64(0.7)},
		{"api.key", "sk-ant-test-key"},
		{"api.url", "https://api.anthropic.com"},
	}

	for _, tc := range testCases {
		if err := tool.SetConfig(tc.path, tc.value); err != nil {
			t.Fatalf("SetConfig(%s, %v) failed: %v", tc.path, tc.value, err)
		}
	}

	// Read the file and verify all values were written
	content, err := os.ReadFile(claudeConfigPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify each value
	if data["model"] != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected model to be 'claude-3-5-sonnet-20241022', got %v", data["model"])
	}

	if data["max_tokens"] != float64(4096) {
		t.Errorf("Expected max_tokens to be 4096, got %v", data["max_tokens"])
	}

	if data["temperature"] != float64(0.7) {
		t.Errorf("Expected temperature to be 0.7, got %v", data["temperature"])
	}

	apiMap, ok := data["api"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected api to be a map")
	}

	if apiMap["key"] != "sk-ant-test-key" {
		t.Errorf("Expected api.key to be 'sk-ant-test-key', got %v", apiMap["key"])
	}

	if apiMap["url"] != "https://api.anthropic.com" {
		t.Errorf("Expected api.url to be 'https://api.anthropic.com', got %v", apiMap["url"])
	}

	// Test GetConfig
	modelValue, err := tool.GetConfig("model")
	if err != nil {
		t.Fatalf("GetConfig(model) failed: %v", err)
	}

	if modelValue != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected model to be 'claude-3-5-sonnet-20241022', got %v", modelValue)
	}

	// Test UnsetConfig
	if err := tool.UnsetConfig("temperature"); err != nil {
		t.Fatalf("UnsetConfig(temperature) failed: %v", err)
	}

	// Verify temperature is gone
	_, err = tool.GetConfig("temperature")
	if err == nil {
		t.Errorf("Expected error for unset temperature value")
	}

	// Verify JSON is still well-formed
	content, err = os.ReadFile(claudeConfigPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var finalData map[string]interface{}
	if err := json.Unmarshal(content, &finalData); err != nil {
		t.Fatalf("Failed to parse JSON after unset: %v", err)
	}

	// Verify temperature is not present
	if _, exists := finalData["temperature"]; exists {
		t.Errorf("Expected temperature to be unset")
	}

	// Verify other values still exist
	if finalData["model"] != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected model to still exist, got %v", finalData["model"])
	}
}

func TestClaudeTool_IntegrationWithConfState(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create conf config directory
	configDir := filepath.Join(tmpDir, ".config", "conf")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	claudeConfigPath := filepath.Join(tmpDir, ".config", "claude", "config.json")

	// Create conf config file with Claude tool configuration
	confConfigPath := filepath.Join(configDir, "config.toml")
	tomlData := `[tools.claude]
name = "claude"
path = "` + claudeConfigPath + `"
`
	if err := os.WriteFile(confConfigPath, []byte(tomlData), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load conf configuration
	conf, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load conf config: %v", err)
	}

	// Set values using conf state management
	conf.SetToolValue("claude", "model", "claude-3-5-sonnet-20241022")
	conf.SetToolValue("claude", "api.key", "sk-ant-test")

	// Save conf configuration
	if err := conf.Save(); err != nil {
		t.Fatalf("Failed to save conf config: %v", err)
	}

	// Create Claude tool and apply values
	tool, err := NewClaudeTool()
	if err != nil {
		t.Fatalf("Failed to create Claude tool: %v", err)
	}

	claudeConfig, exists := conf.GetTool("claude")
	if !exists {
		t.Fatalf("Claude config not found")
	}

	if err := tool.SetAllValues(claudeConfig.Values); err != nil {
		t.Fatalf("SetAllValues failed: %v", err)
	}

	// Verify values were written to JSON file
	content, err := os.ReadFile(claudeConfigPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if data["model"] != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected model to be 'claude-3-5-sonnet-20241022', got %v", data["model"])
	}

	apiMap, ok := data["api"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected api to be a map")
	}

	if apiMap["key"] != "sk-ant-test" {
		t.Errorf("Expected api.key to be 'sk-ant-test', got %v", apiMap["key"])
	}
}
