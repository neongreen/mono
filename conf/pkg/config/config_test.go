package config

import (
	"os"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("Default config should not be nil")
	}

	if len(config.Tools) != 2 {
		t.Errorf("Default config should have 2 tools, got %d", len(config.Tools))
	}

	// Check jj tool config
	jjTool, exists := config.GetTool("jj")
	if !exists {
		t.Error("Default config should include jj tool")
	}
	if jjTool.Name != "jj" {
		t.Errorf("jj tool name should be 'jj', got '%s'", jjTool.Name)
	}
	if !strings.Contains(jjTool.ConfigPath, ".jjconfig.toml") {
		t.Errorf("jj config path should contain '.jjconfig.toml', got '%s'", jjTool.ConfigPath)
	}

	// Check mise tool config
	miseTool, exists := config.GetTool("mise")
	if !exists {
		t.Error("Default config should include mise tool")
	}
	if miseTool.Name != "mise" {
		t.Errorf("mise tool name should be 'mise', got '%s'", miseTool.Name)
	}
	if !strings.Contains(miseTool.ConfigPath, "mise/config.toml") {
		t.Errorf("mise config path should contain 'mise/config.toml', got '%s'", miseTool.ConfigPath)
	}
}

func TestConfigPaths(t *testing.T) {
	configDir, err := ConfigDir()
	if err != nil {
		t.Fatalf("Failed to get config dir: %v", err)
	}

	if !strings.Contains(configDir, ".config/conf") {
		t.Errorf("Config dir should contain '.config/conf', got '%s'", configDir)
	}

	configPath, err := ConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	if !strings.Contains(configPath, "config.toml") {
		t.Errorf("Config path should contain 'config.toml', got '%s'", configPath)
	}
}

func TestSetAndGetTool(t *testing.T) {
	config := DefaultConfig()

	// Add a new tool
	newTool := ToolConfig{
		Name:       "git",
		ConfigPath: "/home/user/.gitconfig",
		SchemaPath: "embedded://git.json",
	}

	config.SetTool("git", newTool)

	// Retrieve the tool
	retrievedTool, exists := config.GetTool("git")
	if !exists {
		t.Error("Should be able to retrieve newly set tool")
	}

	if retrievedTool.Name != newTool.Name {
		t.Errorf("Retrieved tool name should be '%s', got '%s'", newTool.Name, retrievedTool.Name)
	}

	if retrievedTool.ConfigPath != newTool.ConfigPath {
		t.Errorf("Retrieved tool config path should be '%s', got '%s'", newTool.ConfigPath, retrievedTool.ConfigPath)
	}
}

func TestSaveAndLoad(t *testing.T) {
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

	// Create and save config
	originalConfig := DefaultConfig()
	originalConfig.SetTool("test", ToolConfig{
		Name:       "test",
		ConfigPath: "/test/path",
		SchemaPath: "embedded://test.json",
	})

	err = originalConfig.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded config matches original
	if len(loadedConfig.Tools) != len(originalConfig.Tools) {
		t.Errorf("Loaded config should have %d tools, got %d", len(originalConfig.Tools), len(loadedConfig.Tools))
	}

	testTool, exists := loadedConfig.GetTool("test")
	if !exists {
		t.Error("Loaded config should contain test tool")
	}

	if testTool.Name != "test" {
		t.Errorf("Test tool name should be 'test', got '%s'", testTool.Name)
	}
}
