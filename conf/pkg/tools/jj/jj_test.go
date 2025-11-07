package jj

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewJJTool(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jj-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal jj config structure
	jjConfigDir := filepath.Join(tempDir, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create jj config dir: %v", err)
	}

	// Create minimal config file
	configContent := `# JJ config for testing
`
	configPath := filepath.Join(jjConfigDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test tool creation
	tool, err := NewJJTool()
	if err != nil {
		t.Fatalf("Failed to create JJ tool: %v", err)
	}

	if tool == nil {
		t.Fatal("Expected JJ tool to be non-nil")
	}

	if tool.IsDryRun() {
		t.Error("Expected dry-run to be false by default")
	}
}

func TestNewJJToolWithDryRun(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jj-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal jj config structure
	jjConfigDir := filepath.Join(tempDir, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create jj config dir: %v", err)
	}

	configContent := `# JJ config for testing
`
	configPath := filepath.Join(jjConfigDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test dry-run tool creation
	tool, err := NewJJToolWithDryRun(true)
	if err != nil {
		t.Fatalf("Failed to create JJ tool with dry-run: %v", err)
	}

	if !tool.IsDryRun() {
		t.Error("Expected dry-run to be true")
	}
}

func TestJJTool_SetDryRun(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jj-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal jj config structure
	jjConfigDir := filepath.Join(tempDir, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create jj config dir: %v", err)
	}

	configContent := `# JJ config for testing
`
	configPath := filepath.Join(jjConfigDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewJJTool()
	if err != nil {
		t.Fatalf("Failed to create JJ tool: %v", err)
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

func TestJJTool_GetConfigPath(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jj-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal jj config structure
	jjConfigDir := filepath.Join(tempDir, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create jj config dir: %v", err)
	}

	configContent := `# JJ config for testing
`
	configPath := filepath.Join(jjConfigDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewJJTool()
	if err != nil {
		t.Fatalf("Failed to create JJ tool: %v", err)
	}

	path := tool.GetConfigPath()
	expectedPath := filepath.Join(tempDir, ".config", "jj", "config.toml")

	if path != expectedPath {
		t.Errorf("Expected config path %s, got %s", expectedPath, path)
	}
}

func TestJJTool_SetAndGetConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jj-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal jj config structure
	jjConfigDir := filepath.Join(tempDir, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create jj config dir: %v", err)
	}

	configContent := `# JJ config for testing
[user]
name = "Test User"
email = "test@example.com"
`
	configPath := filepath.Join(jjConfigDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewJJTool()
	if err != nil {
		t.Fatalf("Failed to create JJ tool: %v", err)
	}

	// Test getting existing config value
	value, err := tool.GetConfig("user.name")
	if err != nil {
		t.Fatalf("Failed to get config value: %v", err)
	}
	if value != "Test User" {
		t.Errorf("Expected 'Test User', got %v", value)
	}

	// Test setting new config value
	err = tool.SetConfig("user.name", "Updated User")
	if err != nil {
		t.Fatalf("Failed to set config value: %v", err)
	}

	// Verify the value was updated
	value, err = tool.GetConfig("user.name")
	if err != nil {
		t.Fatalf("Failed to get updated config value: %v", err)
	}
	if value != "Updated User" {
		t.Errorf("Expected 'Updated User', got %v", value)
	}
}

func TestJJTool_ValidatePath(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jj-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal jj config structure
	jjConfigDir := filepath.Join(tempDir, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create jj config dir: %v", err)
	}

	configContent := `# JJ config for testing
`
	configPath := filepath.Join(jjConfigDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewJJTool()
	if err != nil {
		t.Fatalf("Failed to create JJ tool: %v", err)
	}

	// Test valid paths (based on schema)
	validPaths := []string{
		"user.name",
		"user.email",
		"ui.default-command",
		"ui.paginate",
	}

	for _, path := range validPaths {
		err := tool.SetConfig(path, "test-value")
		if err != nil && err.Error() == "invalid configuration path: "+path {
			t.Errorf("Expected path '%s' to be valid according to schema", path)
		}
	}

	// Test invalid path
	err = tool.SetConfig("invalid.nonexistent.path", "value")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestJJTool_ListCommonSettings(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jj-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create minimal jj config structure
	jjConfigDir := filepath.Join(tempDir, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create jj config dir: %v", err)
	}

	configContent := `# JJ config for testing
`
	configPath := filepath.Join(jjConfigDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewJJTool()
	if err != nil {
		t.Fatalf("Failed to create JJ tool: %v", err)
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
		"user.name":          false,
		"user.email":         false,
		"ui.default-command": false,
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
