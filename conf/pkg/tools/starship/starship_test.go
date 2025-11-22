package starship

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewStarshipTool(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "starship-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal starship config structure
	starshipConfigDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(starshipConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create starship config dir: %v", err)
	}

	// Create minimal config file
	configContent := `# Starship config for testing
add_newline = true
`
	configPath := filepath.Join(starshipConfigDir, "starship.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test tool creation
	tool, err := NewStarshipTool()
	if err != nil {
		t.Fatalf("Failed to create starship tool: %v", err)
	}

	if tool == nil {
		t.Fatal("Expected starship tool to be non-nil")
	}

	if tool.IsDryRun() {
		t.Error("Expected dry-run to be false by default")
	}
}

func TestNewStarshipToolWithDryRun(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "starship-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal starship config structure
	starshipConfigDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(starshipConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create starship config dir: %v", err)
	}

	configContent := `# Starship config for testing
add_newline = true
`
	configPath := filepath.Join(starshipConfigDir, "starship.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test dry-run tool creation
	tool, err := NewStarshipToolWithDryRun(true)
	if err != nil {
		t.Fatalf("Failed to create starship tool with dry-run: %v", err)
	}

	if !tool.IsDryRun() {
		t.Error("Expected dry-run to be true")
	}
}

func TestStarshipTool_SetDryRun(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "starship-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal starship config structure
	starshipConfigDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(starshipConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create starship config dir: %v", err)
	}

	configContent := `# Starship config for testing
add_newline = true
`
	configPath := filepath.Join(starshipConfigDir, "starship.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewStarshipTool()
	if err != nil {
		t.Fatalf("Failed to create starship tool: %v", err)
	}

	// Test setting dry-run to true
	tool.SetDryRun(true)
	if !tool.IsDryRun() {
		t.Error("Expected dry-run to be true")
	}

	// Test setting dry-run to false
	tool.SetDryRun(false)
	if tool.IsDryRun() {
		t.Error("Expected dry-run to be false")
	}
}

func TestStarshipTool_GetConfigPath(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "starship-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal starship config structure
	starshipConfigDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(starshipConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create starship config dir: %v", err)
	}

	configContent := `# Starship config for testing
add_newline = true
`
	configPath := filepath.Join(starshipConfigDir, "starship.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewStarshipTool()
	if err != nil {
		t.Fatalf("Failed to create starship tool: %v", err)
	}

	path := tool.GetConfigPath()
	expectedPath := filepath.Join(tempDir, ".config", "starship.toml")

	if path != expectedPath {
		t.Errorf("Expected config path %s, got %s", expectedPath, path)
	}
}

func TestStarshipTool_SetAndGetConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "starship-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal starship config structure
	starshipConfigDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(starshipConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create starship config dir: %v", err)
	}

	configContent := `# Starship config for testing
add_newline = true
command_timeout = 500
`
	configPath := filepath.Join(starshipConfigDir, "starship.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewStarshipTool()
	if err != nil {
		t.Fatalf("Failed to create starship tool: %v", err)
	}

	// Test getting existing config value
	value, err := tool.GetConfig("add_newline")
	if err != nil {
		t.Fatalf("Failed to get config value: %v", err)
	}
	if value != true {
		t.Errorf("Expected true, got %v", value)
	}

	// Test setting new config value
	err = tool.SetConfig("add_newline", false)
	if err != nil {
		t.Fatalf("Failed to set config value: %v", err)
	}

	// Verify the value was updated
	value, err = tool.GetConfig("add_newline")
	if err != nil {
		t.Fatalf("Failed to get updated config value: %v", err)
	}
	if value != false {
		t.Errorf("Expected false, got %v", value)
	}
}

func TestStarshipTool_UnsetConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "starship-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal starship config structure
	starshipConfigDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(starshipConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create starship config dir: %v", err)
	}

	configContent := `# Starship config for testing
add_newline = true
command_timeout = 500
`
	configPath := filepath.Join(starshipConfigDir, "starship.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewStarshipTool()
	if err != nil {
		t.Fatalf("Failed to create starship tool: %v", err)
	}

	// Verify value exists before unsetting
	_, err = tool.GetConfig("command_timeout")
	if err != nil {
		t.Fatalf("Expected config value to exist before unsetting: %v", err)
	}

	// Test unsetting config value
	err = tool.UnsetConfig("command_timeout")
	if err != nil {
		t.Fatalf("Failed to unset config value: %v", err)
	}

	// Verify the value was removed
	_, err = tool.GetConfig("command_timeout")
	if err == nil {
		t.Error("Expected error when getting unset config value")
	}
}

func TestStarshipTool_ValidatePath(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "starship-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal starship config structure
	starshipConfigDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(starshipConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create starship config dir: %v", err)
	}

	configContent := `# Starship config for testing
add_newline = true
`
	configPath := filepath.Join(starshipConfigDir, "starship.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewStarshipTool()
	if err != nil {
		t.Fatalf("Failed to create starship tool: %v", err)
	}

	// Test valid paths (starship is flexible, most paths should be valid)
	validPaths := []string{
		"add_newline",
		"format",
		"command_timeout",
		"character.success_symbol",
		"git_branch.format",
		"directory.truncation_length",
	}

	for _, path := range validPaths {
		err := tool.SetConfig(path, "test-value")
		if err != nil && err.Error() == "invalid configuration path: "+path {
			t.Errorf("Expected path '%s' to be valid", path)
		}
	}

	// Test invalid path (empty string)
	err = tool.SetConfig("", "value")
	if err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestStarshipTool_PreviewSetConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "starship-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal starship config structure
	starshipConfigDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(starshipConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create starship config dir: %v", err)
	}

	configContent := `# Starship config for testing
add_newline = true
`
	configPath := filepath.Join(starshipConfigDir, "starship.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewStarshipTool()
	if err != nil {
		t.Fatalf("Failed to create starship tool: %v", err)
	}

	// Test preview set config
	preview, err := tool.PreviewSetConfig("add_newline", false)
	if err != nil {
		t.Fatalf("Failed to preview set config: %v", err)
	}

	if preview == "" {
		t.Error("Expected non-empty preview")
	}

	// Check that preview contains expected information
	expectedStrings := []string{"Operation: SET", "Path: add_newline", "Value: false"}
	for _, expected := range expectedStrings {
		if !containsString(preview, expected) {
			t.Errorf("Expected preview to contain '%s', got: %s", expected, preview)
		}
	}
}

func TestStarshipTool_NestedConfiguration(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "starship-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal starship config structure
	starshipConfigDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(starshipConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create starship config dir: %v", err)
	}

	configContent := `# Starship config for testing
add_newline = true

[character]
success_symbol = "[➜](bold green)"
error_symbol = "[➜](bold red)"
`
	configPath := filepath.Join(starshipConfigDir, "starship.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewStarshipTool()
	if err != nil {
		t.Fatalf("Failed to create starship tool: %v", err)
	}

	// Test getting nested config value
	value, err := tool.GetConfig("character.success_symbol")
	if err != nil {
		t.Fatalf("Failed to get nested config value: %v", err)
	}
	if value != "[➜](bold green)" {
		t.Errorf("Expected '[➜](bold green)', got %v", value)
	}

	// Test setting nested config value
	err = tool.SetConfig("character.error_symbol", "[✗](bold red)")
	if err != nil {
		t.Fatalf("Failed to set nested config value: %v", err)
	}

	// Verify the nested value was updated
	value, err = tool.GetConfig("character.error_symbol")
	if err != nil {
		t.Fatalf("Failed to get updated nested config value: %v", err)
	}
	if value != "[✗](bold red)" {
		t.Errorf("Expected '[✗](bold red)', got %v", value)
	}

	// Test setting entirely new nested config
	err = tool.SetConfig("git_branch.symbol", "🌱 ")
	if err != nil {
		t.Fatalf("Failed to set new nested config value: %v", err)
	}

	// Verify the new nested value was set
	value, err = tool.GetConfig("git_branch.symbol")
	if err != nil {
		t.Fatalf("Failed to get new nested config value: %v", err)
	}
	if value != "🌱 " {
		t.Errorf("Expected '🌱 ', got %v", value)
	}
}

func TestStarshipTool_ListAllSettings(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "starship-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal starship config structure
	starshipConfigDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(starshipConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create starship config dir: %v", err)
	}

	configContent := `# Starship config for testing
add_newline = true
command_timeout = 1000

[character]
success_symbol = "[➜](bold green)"
`
	configPath := filepath.Join(starshipConfigDir, "starship.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewStarshipTool()
	if err != nil {
		t.Fatalf("Failed to create starship tool: %v", err)
	}

	// Test listing all settings
	settings, err := tool.ListAllSettings()
	if err != nil {
		t.Fatalf("Failed to list all settings: %v", err)
	}

	// Should have many settings from schema
	if len(settings) < 50 {
		t.Errorf("Expected at least 50 settings from schema, got %d", len(settings))
	}

	// Check that we have some expected top-level settings
	expectedSettings := map[string]bool{
		"format":          false,
		"add_newline":     false,
		"command_timeout": false,
	}

	for _, setting := range settings {
		if _, exists := expectedSettings[setting.Path]; exists {
			expectedSettings[setting.Path] = true

			// Verify setting has required fields
			if setting.Type == "" {
				t.Errorf("Setting %s has empty type", setting.Path)
			}
		}
	}

	// Verify expected settings were found
	for path, found := range expectedSettings {
		if !found {
			t.Errorf("Expected setting %s not found in all settings", path)
		}
	}

	// Check that current values are populated for configured settings
	var addNewlineFound, commandTimeoutFound bool
	for _, setting := range settings {
		if setting.Path == "add_newline" {
			addNewlineFound = true
			if !setting.IsSet {
				t.Error("Expected add_newline to be marked as set")
			}
			if setting.CurrentValue != true {
				t.Errorf("Expected add_newline current value to be true, got %v", setting.CurrentValue)
			}
		}
		if setting.Path == "command_timeout" {
			commandTimeoutFound = true
			if !setting.IsSet {
				t.Error("Expected command_timeout to be marked as set")
			}
			if setting.CurrentValue != int64(1000) {
				t.Errorf("Expected command_timeout current value to be 1000, got %v", setting.CurrentValue)
			}
		}
	}

	if !addNewlineFound {
		t.Error("add_newline setting not found")
	}
	if !commandTimeoutFound {
		t.Error("command_timeout setting not found")
	}

	// Verify that unset settings have IsSet=false
	for _, setting := range settings {
		if setting.Path == "scan_timeout" { // This one is not set in our test config
			if setting.IsSet {
				t.Error("Expected scan_timeout to not be marked as set")
			}
			if setting.CurrentValue != nil {
				t.Error("Expected scan_timeout CurrentValue to be nil")
			}
			// Should have a default value from schema
			if setting.Default == nil {
				t.Error("Expected scan_timeout to have a default value from schema")
			}
		}
	}
}

func TestLookupValueByPath(t *testing.T) {
	data := map[string]any{
		"top_level": "value1",
		"nested": map[string]any{
			"key1": "value2",
			"key2": map[string]any{
				"deep": "value3",
			},
		},
		"number": int64(42),
	}

	tests := []struct {
		name     string
		path     string
		expected any
	}{
		{"top level key", "top_level", "value1"},
		{"nested key", "nested.key1", "value2"},
		{"deeply nested key", "nested.key2.deep", "value3"},
		{"number value", "number", int64(42)},
		{"non-existent key", "nonexistent", nil},
		{"non-existent nested", "nested.nonexistent", nil},
		{"empty path", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lookupValueByPath(data, tt.path)
			if result != tt.expected {
				t.Errorf("lookupValueByPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestStarshipTool_ValidateValueAndForce(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "starship-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	starshipConfigDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(starshipConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create starship config dir: %v", err)
	}

	configPath := filepath.Join(starshipConfigDir, "starship.toml")
	if err := os.WriteFile(configPath, []byte("command_timeout = 500\n"), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewStarshipTool()
	if err != nil {
		t.Fatalf("Failed to create starship tool: %v", err)
	}

	if err := tool.SetConfig("command_timeout", "fast"); err == nil {
		t.Fatalf("Expected validation error for invalid value type")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}
	if strings.Contains(string(content), "fast") {
		t.Fatalf("Config should not be written when validation fails")
	}

	tool.SetForce(true)
	if err := tool.SetConfig("command_timeout", "fast"); err != nil {
		t.Fatalf("Force should bypass validation: %v", err)
	}

	content, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}
	if !strings.Contains(string(content), "fast") {
		t.Errorf("Expected config to be written when force is enabled")
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{"simple path", "a.b.c", []string{"a", "b", "c"}},
		{"single segment", "single", []string{"single"}},
		{"empty path", "", nil},
		{"quoted segment", `a."b.c".d`, []string{"a", "b.c", "d"}},
		{"nested path", "character.success_symbol", []string{"character", "success_symbol"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitPath(tt.path)
			if len(result) != len(tt.expected) {
				t.Errorf("splitPath(%q) returned %d parts, want %d", tt.path, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitPath(%q)[%d] = %q, want %q", tt.path, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) && (s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				stringContains(s, substr))))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
