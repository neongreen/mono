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
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	claudeConfigPath := filepath.Join(tmpDir, ".config", "claude", "config.json")

	// Create conf config file with Claude tool configuration
	confConfigPath := filepath.Join(configDir, "config.toml")
	tomlData := `[tools.claude]
name = "claude"
path = "` + claudeConfigPath + `"
`
	if err := os.WriteFile(confConfigPath, []byte(tomlData), 0o644); err != nil {
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
		value any
	}{
		{"model", "sonnet"},
		{"alwaysThinkingEnabled", true},
		{"outputStyle", "markdown"},
		{"apiKeyHelper", "/path/to/helper.sh"},
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

	var data map[string]any
	if err := json.Unmarshal(content, &data); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify each value
	if data["model"] != "sonnet" {
		t.Errorf("Expected model to be 'sonnet', got %v", data["model"])
	}

	if data["alwaysThinkingEnabled"] != true {
		t.Errorf("Expected alwaysThinkingEnabled to be true, got %v", data["alwaysThinkingEnabled"])
	}

	if data["outputStyle"] != "markdown" {
		t.Errorf("Expected outputStyle to be 'markdown', got %v", data["outputStyle"])
	}

	if data["apiKeyHelper"] != "/path/to/helper.sh" {
		t.Errorf("Expected apiKeyHelper to be '/path/to/helper.sh', got %v", data["apiKeyHelper"])
	}

	// Test GetConfig
	modelValue, err := tool.GetConfig("model")
	if err != nil {
		t.Fatalf("GetConfig(model) failed: %v", err)
	}

	if modelValue != "sonnet" {
		t.Errorf("Expected model to be 'sonnet', got %v", modelValue)
	}

	// Test UnsetConfig
	if err := tool.UnsetConfig("outputStyle"); err != nil {
		t.Fatalf("UnsetConfig(outputStyle) failed: %v", err)
	}

	// Verify outputStyle is gone
	_, err = tool.GetConfig("outputStyle")
	if err == nil {
		t.Errorf("Expected error for unset outputStyle value")
	}

	// Verify JSON is still well-formed
	content, err = os.ReadFile(claudeConfigPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var finalData map[string]any
	if err := json.Unmarshal(content, &finalData); err != nil {
		t.Fatalf("Failed to parse JSON after unset: %v", err)
	}

	// Verify outputStyle is not present
	if _, exists := finalData["outputStyle"]; exists {
		t.Errorf("Expected outputStyle to be unset")
	}

	// Verify model still exists
	if finalData["model"] != "sonnet" {
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
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	claudeConfigPath := filepath.Join(tmpDir, ".config", "claude", "config.json")

	// Create conf config file with Claude tool configuration
	confConfigPath := filepath.Join(configDir, "config.toml")
	tomlData := `[tools.claude]
name = "claude"
path = "` + claudeConfigPath + `"
`
	if err := os.WriteFile(confConfigPath, []byte(tomlData), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load conf configuration
	conf, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load conf config: %v", err)
	}

	// Set values using conf state management
	conf.SetToolValue("claude", "model", "sonnet")
	conf.SetToolValue("claude", "alwaysThinkingEnabled", true)

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

	var data map[string]any
	if err := json.Unmarshal(content, &data); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if data["model"] != "sonnet" {
		t.Errorf("Expected model to be 'sonnet', got %v", data["model"])
	}

	if data["alwaysThinkingEnabled"] != true {
		t.Errorf("Expected alwaysThinkingEnabled to be true, got %v", data["alwaysThinkingEnabled"])
	}
}
