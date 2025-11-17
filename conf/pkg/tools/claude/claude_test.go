package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupTestConfig(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "conf")
	claudeConfigPath := filepath.Join(tmpDir, ".config", "claude", "config.json")

	// Create conf config directory
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Save conf config
	confConfigPath := filepath.Join(configDir, "config.toml")

	// Write as minimal TOML format for testing
	tomlData := `[tools.claude]
name = "claude"
path = "` + claudeConfigPath + `"
`
	if err := os.WriteFile(confConfigPath, []byte(tomlData), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Set config path for testing
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	cleanup := func() {
		os.Setenv("HOME", origHome)
	}

	return claudeConfigPath, cleanup
}

func TestClaudeTool_SetConfig(t *testing.T) {
	claudeConfigPath, cleanup := setupTestConfig(t)
	defer cleanup()

	tool, err := NewClaudeTool()
	if err != nil {
		t.Fatalf("Failed to create Claude tool: %v", err)
	}

	// Set a value
	if err := tool.SetConfig("model", "sonnet"); err != nil {
		t.Fatalf("SetConfig failed: %v", err)
	}

	// Read back and verify
	content, err := os.ReadFile(claudeConfigPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal(content, &data); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if data["model"] != "sonnet" {
		t.Errorf("Expected model to be 'sonnet', got %v", data["model"])
	}
}

func TestClaudeTool_GetConfig(t *testing.T) {
	claudeConfigPath, cleanup := setupTestConfig(t)
	defer cleanup()

	// Create initial config
	data := map[string]any{
		"model":                 "sonnet",
		"alwaysThinkingEnabled": true,
		"outputStyle":           "markdown",
		"apiKeyHelper":          "/path/to/helper.sh",
	}

	os.MkdirAll(filepath.Dir(claudeConfigPath), 0o755)
	content, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(claudeConfigPath, content, 0o644)

	tool, err := NewClaudeTool()
	if err != nil {
		t.Fatalf("Failed to create Claude tool: %v", err)
	}

	// Test getting various values
	tests := []struct {
		name     string
		path     string
		expected any
	}{
		{
			name:     "get model",
			path:     "model",
			expected: "sonnet",
		},
		{
			name:     "get alwaysThinkingEnabled",
			path:     "alwaysThinkingEnabled",
			expected: true,
		},
		{
			name:     "get outputStyle",
			path:     "outputStyle",
			expected: "markdown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tool.GetConfig(tt.path)
			if err != nil {
				t.Fatalf("GetConfig failed: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, value)
			}
		})
	}
}

func TestClaudeTool_UnsetConfig(t *testing.T) {
	claudeConfigPath, cleanup := setupTestConfig(t)
	defer cleanup()

	// Create initial config
	data := map[string]any{
		"model":                 "sonnet",
		"alwaysThinkingEnabled": true,
	}

	os.MkdirAll(filepath.Dir(claudeConfigPath), 0o755)
	content, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(claudeConfigPath, content, 0o644)

	tool, err := NewClaudeTool()
	if err != nil {
		t.Fatalf("Failed to create Claude tool: %v", err)
	}

	// Unset alwaysThinkingEnabled
	if err := tool.UnsetConfig("alwaysThinkingEnabled"); err != nil {
		t.Fatalf("UnsetConfig failed: %v", err)
	}

	// Verify it's gone
	_, err = tool.GetConfig("alwaysThinkingEnabled")
	if err == nil {
		t.Errorf("Expected error for unset value")
	}

	// Verify model still exists
	value, err := tool.GetConfig("model")
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if value != "sonnet" {
		t.Errorf("Expected model to still exist, got %v", value)
	}
}

func TestClaudeTool_SetAllValues(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	tool, err := NewClaudeTool()
	if err != nil {
		t.Fatalf("Failed to create Claude tool: %v", err)
	}

	values := map[string]any{
		"model":                 "sonnet",
		"alwaysThinkingEnabled": true,
		"outputStyle":           "markdown",
	}

	if err := tool.SetAllValues(values); err != nil {
		t.Fatalf("SetAllValues failed: %v", err)
	}

	// Read back and verify
	allValues, err := tool.GetAllValues()
	if err != nil {
		t.Fatalf("GetAllValues failed: %v", err)
	}

	if allValues["model"] != "sonnet" {
		t.Errorf("Expected model to be 'sonnet', got %v", allValues["model"])
	}

	if allValues["alwaysThinkingEnabled"] != true {
		t.Errorf("Expected alwaysThinkingEnabled to be true, got %v", allValues["alwaysThinkingEnabled"])
	}
}

func TestClaudeTool_DryRun(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	tool, err := NewClaudeToolWithDryRun(true)
	if err != nil {
		t.Fatalf("Failed to create Claude tool: %v", err)
	}

	claudeConfigPath := tool.GetConfigPath()

	// SetConfig in dry-run mode should not create the file
	if err := tool.SetConfig("model", "sonnet"); err != nil {
		t.Fatalf("SetConfig failed: %v", err)
	}

	// File should not exist
	if _, err := os.Stat(claudeConfigPath); !os.IsNotExist(err) {
		t.Errorf("File should not exist in dry-run mode")
	}
}

func TestClaudeTool_SchemaValidation(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	tool, err := NewClaudeTool()
	if err != nil {
		t.Fatalf("Failed to create Claude tool: %v", err)
	}

	// Test invalid path should fail
	err = tool.SetConfig("invalid.path", "value")
	if err == nil {
		t.Error("Expected error for invalid path")
	}

	// Test valid path should succeed
	err = tool.SetConfig("model", "sonnet")
	if err != nil {
		t.Errorf("SetConfig with valid path failed: %v", err)
	}

	// Test another invalid path
	err = tool.SetConfig("nonexistent.setting", "value")
	if err == nil {
		t.Error("Expected error for invalid nested path")
	}

	// Test valid path should succeed
	err = tool.SetConfig("alwaysThinkingEnabled", true)
	if err != nil {
		t.Errorf("SetConfig with valid path failed: %v", err)
	}
}
