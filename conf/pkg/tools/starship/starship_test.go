package starship

import (
	"os"
	"path/filepath"
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

func TestStarshipTool_ListCommonSettings(t *testing.T) {
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

	settings := tool.ListCommonSettings()

	if len(settings) == 0 {
		t.Error("Expected some common settings to be returned")
	}

	// Verify structure of common settings
	for i, setting := range settings {
		if setting.Path == "" {
			t.Errorf("Setting %d has empty path", i)
		}
		if setting.Description == "" {
			t.Errorf("Setting %d has empty description", i)
		}
		if setting.Type == "" {
			t.Errorf("Setting %d has empty type", i)
		}
		if setting.Example == "" {
			t.Errorf("Setting %d has empty example", i)
		}
	}

	// Check for some expected common settings
	expectedPaths := map[string]bool{
		"format":                   false,
		"add_newline":              false,
		"command_timeout":          false,
		"character.success_symbol": false,
	}

	for _, setting := range settings {
		if _, exists := expectedPaths[setting.Path]; exists {
			expectedPaths[setting.Path] = true
		}
	}

	for path, found := range expectedPaths {
		if !found {
			t.Errorf("Expected common setting '%s' not found", path)
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
