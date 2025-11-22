package jj

import (
	"os"
	"path/filepath"
	"strings"
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

func TestJJTool_SetConfig_InvalidValueFails(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "jj-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	jjConfigDir := filepath.Join(tempDir, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create jj config dir: %v", err)
	}

	configPath := filepath.Join(jjConfigDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("# config"), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewJJTool()
	if err != nil {
		t.Fatalf("Failed to create JJ tool: %v", err)
	}

	if err := tool.SetConfig("ui.paginate", "not-bool"); err == nil {
		t.Fatalf("Expected validation error for invalid value type")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}
	if strings.Contains(string(content), "not-bool") {
		t.Errorf("Config should not be written when validation fails")
	}
}

func TestJJTool_SetConfig_ForceAllowsInvalidValue(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "jj-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	jjConfigDir := filepath.Join(tempDir, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create jj config dir: %v", err)
	}

	configPath := filepath.Join(jjConfigDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("# config"), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tool, err := NewJJTool()
	if err != nil {
		t.Fatalf("Failed to create JJ tool: %v", err)
	}
	tool.SetForce(true)

	if err := tool.SetConfig("ui.paginate", "not-bool"); err != nil {
		t.Fatalf("Force should bypass validation: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}
	if !strings.Contains(string(content), "not-bool") {
		t.Errorf("Expected config to be written when force is enabled")
	}
}
