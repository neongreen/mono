package config

import (
	"os"
	"testing"
)

func TestConfig_SetToolValue(t *testing.T) {
	config := &Config{
		Tools: make(map[string]ToolConfig),
	}

	// Test setting a tool value
	config.SetToolValue("jj", "user.name", "Test User")

	// Verify the tool was created
	tool, exists := config.Tools["jj"]
	if !exists {
		t.Fatal("Expected tool 'jj' to be created")
	}

	// Verify the value was set
	if tool.Values == nil {
		t.Fatal("Expected tool values to be initialized")
	}

	value, exists := tool.Values["user.name"]
	if !exists {
		t.Fatal("Expected value 'user.name' to be set")
	}

	if value != "Test User" {
		t.Errorf("Expected value to be 'Test User', got %v", value)
	}
}

func TestConfig_GetToolValue(t *testing.T) {
	config := &Config{
		Tools: map[string]ToolConfig{
			"jj": {
				Name: "jj",
				Values: map[string]interface{}{
					"user.name":  "Test User",
					"user.email": "test@example.com",
				},
			},
		},
	}

	// Test getting existing value
	value, exists := config.GetToolValue("jj", "user.name")
	if !exists {
		t.Fatal("Expected value to exist")
	}
	if value != "Test User" {
		t.Errorf("Expected 'Test User', got %v", value)
	}

	// Test getting non-existent value
	_, exists = config.GetToolValue("jj", "non.existent")
	if exists {
		t.Error("Expected value to not exist")
	}

	// Test getting value from non-existent tool
	_, exists = config.GetToolValue("nonexistent", "user.name")
	if exists {
		t.Error("Expected value to not exist for non-existent tool")
	}
}

func TestConfig_UnsetToolValue(t *testing.T) {
	config := &Config{
		Tools: map[string]ToolConfig{
			"jj": {
				Name: "jj",
				Values: map[string]interface{}{
					"user.name":  "Test User",
					"user.email": "test@example.com",
				},
			},
		},
	}

	// Test unsetting existing value
	config.UnsetToolValue("jj", "user.name")

	_, exists := config.GetToolValue("jj", "user.name")
	if exists {
		t.Error("Expected value to be unset")
	}

	// Verify other values are still there
	value, exists := config.GetToolValue("jj", "user.email")
	if !exists || value != "test@example.com" {
		t.Error("Expected other values to remain")
	}

	// Test unsetting from non-existent tool (should not panic)
	config.UnsetToolValue("nonexistent", "user.name")
}

func TestConfig_SetShim(t *testing.T) {
	config := &Config{}

	// Test setting shims
	config.SetShim("ll", "ls -la")
	config.SetShim("gst", "git status")

	if config.Shims == nil {
		t.Fatal("Expected shims to be initialized")
	}

	if config.Shims["ll"] != "ls -la" {
		t.Errorf("Expected 'll' shim to be 'ls -la', got %v", config.Shims["ll"])
	}

	if config.Shims["gst"] != "git status" {
		t.Errorf("Expected 'gst' shim to be 'git status', got %v", config.Shims["gst"])
	}
}

func TestConfig_GetShim(t *testing.T) {
	config := &Config{
		Shims: map[string]string{
			"ll":  "ls -la",
			"gst": "git status",
		},
	}

	// Test getting existing shim
	command, exists := config.GetShim("ll")
	if !exists {
		t.Fatal("Expected shim to exist")
	}
	if command != "ls -la" {
		t.Errorf("Expected 'ls -la', got %v", command)
	}

	// Test getting non-existent shim
	_, exists = config.GetShim("nonexistent")
	if exists {
		t.Error("Expected shim to not exist")
	}

	// Test getting from nil shims
	config.Shims = nil
	_, exists = config.GetShim("ll")
	if exists {
		t.Error("Expected shim to not exist when Shims is nil")
	}
}

func TestConfig_UnsetShim(t *testing.T) {
	config := &Config{
		Shims: map[string]string{
			"ll":  "ls -la",
			"gst": "git status",
		},
	}

	// Test unsetting existing shim
	config.UnsetShim("ll")

	_, exists := config.GetShim("ll")
	if exists {
		t.Error("Expected shim to be unset")
	}

	// Verify other shims are still there
	command, exists := config.GetShim("gst")
	if !exists || command != "git status" {
		t.Error("Expected other shims to remain")
	}

	// Test unsetting from nil shims (should not panic)
	config.Shims = nil
	config.UnsetShim("nonexistent")
}

func TestConfig_DeclarativeStateSaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "conf-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create config with declarative state
	config := &Config{
		Tools: map[string]ToolConfig{
			"jj": {
				Name:       "jj",
				ConfigPath: "/test/jj/config.toml",
				Values: map[string]interface{}{
					"user.name":  "Test User",
					"user.email": "test@example.com",
				},
			},
			"mise": {
				Name:       "mise",
				ConfigPath: "/test/mise/config.toml",
				Values: map[string]interface{}{
					"settings.experimental": true,
					"settings.jobs":         4,
				},
			},
		},
		Shims: map[string]string{
			"ll":  "ls -la",
			"gst": "git status",
		},
	}

	// Save using existing method
	if err := config.Save(); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load using existing method
	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify tools were loaded correctly
	jjTool, exists := loadedConfig.GetTool("jj")
	if !exists {
		t.Fatal("Expected jj tool to be loaded")
	}

	// Verify tool values
	userName, exists := loadedConfig.GetToolValue("jj", "user.name")
	if !exists || userName != "Test User" {
		t.Errorf("Expected user.name to be 'Test User', got %v", userName)
	}

	experimental, exists := loadedConfig.GetToolValue("mise", "settings.experimental")
	if !exists || experimental != true {
		t.Errorf("Expected settings.experimental to be true, got %v", experimental)
	}

	// Verify shims were loaded correctly
	llCommand, exists := loadedConfig.GetShim("ll")
	if !exists || llCommand != "ls -la" {
		t.Errorf("Expected ll shim to be 'ls -la', got %v", llCommand)
	}

	gstCommand, exists := loadedConfig.GetShim("gst")
	if !exists || gstCommand != "git status" {
		t.Errorf("Expected gst shim to be 'git status', got %v", gstCommand)
	}

	// Verify tool metadata is preserved
	if jjTool.ConfigPath != "/test/jj/config.toml" {
		t.Errorf("Expected jj config path to be preserved")
	}
}
