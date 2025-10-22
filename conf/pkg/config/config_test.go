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

	if len(config.Tools) != 3 {
		t.Errorf("Default config should have 3 tools, got %d", len(config.Tools))
	}

	// Check jj tool config
	jjTool, exists := config.GetTool("jj")
	if !exists {
		t.Error("Default config should include jj tool")
	}
	if jjTool.Name != "jj" {
		t.Errorf("jj tool name should be 'jj', got '%s'", jjTool.Name)
	}
	// GetTool() expands tilde, so check for expanded path
	if !strings.Contains(jjTool.ConfigPath, ".config/jj/config.toml") {
		t.Errorf("jj config path should contain '.config/jj/config.toml', got '%s'", jjTool.ConfigPath)
	}
	if strings.HasPrefix(jjTool.ConfigPath, "~") {
		t.Errorf("jj config path should be expanded, got '%s'", jjTool.ConfigPath)
	}

	// Check mise tool config
	miseTool, exists := config.GetTool("mise")
	if !exists {
		t.Error("Default config should include mise tool")
	}
	if miseTool.Name != "mise" {
		t.Errorf("mise tool name should be 'mise', got '%s'", miseTool.Name)
	}
	// GetTool() expands tilde, so check for expanded path
	if !strings.Contains(miseTool.ConfigPath, ".config/mise/config.toml") {
		t.Errorf("mise config path should contain '.config/mise/config.toml', got '%s'", miseTool.ConfigPath)
	}
	if strings.HasPrefix(miseTool.ConfigPath, "~") {
		t.Errorf("mise config path should be expanded, got '%s'", miseTool.ConfigPath)
	}

	// Check starship tool config
	starshipTool, exists := config.GetTool("starship")
	if !exists {
		t.Error("Default config should include starship tool")
	}
	if starshipTool.Name != "starship" {
		t.Errorf("starship tool name should be 'starship', got '%s'", starshipTool.Name)
	}
	// GetTool() expands tilde, so check for expanded path
	if !strings.Contains(starshipTool.ConfigPath, ".config/starship.toml") {
		t.Errorf("starship config path should contain '.config/starship.toml', got '%s'", starshipTool.ConfigPath)
	}
	if strings.HasPrefix(starshipTool.ConfigPath, "~") {
		t.Errorf("starship config path should be expanded, got '%s'", starshipTool.ConfigPath)
	}
}

func TestExpandPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "expand tilde with slash",
			input:    "~/.config/jj/config.toml",
			expected: homeDir + "/.config/jj/config.toml",
			wantErr:  false,
		},
		{
			name:     "expand just tilde",
			input:    "~",
			expected: homeDir,
			wantErr:  false,
		},
		{
			name:     "no tilde - absolute path",
			input:    "/etc/config.toml",
			expected: "/etc/config.toml",
			wantErr:  false,
		},
		{
			name:     "no tilde - relative path",
			input:    "./config.toml",
			expected: "./config.toml",
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "tilde in middle - no expansion",
			input:    "/path/~/config.toml",
			expected: "/path/~/config.toml",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandPath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ExpandPath() = %v, want %v", result, tt.expected)
			}
		})
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

func TestLoadExpandsTildePaths(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "conf-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override HOME for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create config with tilde paths
	config := &Config{
		Tools: map[string]ToolConfig{
			"jj": {
				Name:       "jj",
				ConfigPath: "~/.config/jj/config.toml",
				SchemaPath: "embedded://jj.json",
			},
			"mise": {
				Name:       "mise",
				ConfigPath: "~/.config/mise/config.toml",
				SchemaPath: "embedded://mise.toml",
			},
		},
	}

	// Save config
	err = config.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config - should expand tilde paths
	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify jj path was expanded
	jjTool, exists := loadedConfig.GetTool("jj")
	if !exists {
		t.Fatal("Loaded config should contain jj tool")
	}

	expectedJJPath := tempDir + "/.config/jj/config.toml"
	if jjTool.ConfigPath != expectedJJPath {
		t.Errorf("jj ConfigPath should be expanded to '%s', got '%s'", expectedJJPath, jjTool.ConfigPath)
	}

	// Verify mise path was expanded
	miseTool, exists := loadedConfig.GetTool("mise")
	if !exists {
		t.Fatal("Loaded config should contain mise tool")
	}

	expectedMisePath := tempDir + "/.config/mise/config.toml"
	if miseTool.ConfigPath != expectedMisePath {
		t.Errorf("mise ConfigPath should be expanded to '%s', got '%s'", expectedMisePath, miseTool.ConfigPath)
	}
}

func TestPerToolConfigLoading(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "conf-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override HOME for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create config directory
	configDir, err := ConfigDir()
	if err != nil {
		t.Fatalf("Failed to get config dir: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create config.toml with metadata only
	configToml := `[tools.jj]
name = 'jj'
config_path = '~/.config/jj/config.toml'
schema_path = 'embedded://jj.json'
`
	configPath, err := ConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(configToml), 0644); err != nil {
		t.Fatalf("Failed to write config.toml: %v", err)
	}

	// Create per-tool file jj.toml with values
	jjToml := `'user.name' = 'Emily'
'user.email' = 'emily@artyom.me'
`
	jjTomlPath := configDir + "/jj.toml"
	if err := os.WriteFile(jjTomlPath, []byte(jjToml), 0644); err != nil {
		t.Fatalf("Failed to write jj.toml: %v", err)
	}

	// Load config
	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values were loaded from per-tool file
	jjTool, exists := config.GetTool("jj")
	if !exists {
		t.Fatal("Config should contain jj tool")
	}

	if jjTool.Values == nil {
		t.Fatal("jj tool should have Values loaded from per-tool file")
	}

	if name, ok := jjTool.Values["user.name"]; !ok || name != "Emily" {
		t.Errorf("Expected user.name to be 'Emily', got %v", name)
	}

	if email, ok := jjTool.Values["user.email"]; !ok || email != "emily@artyom.me" {
		t.Errorf("Expected user.email to be 'emily@artyom.me', got %v", email)
	}
}

func TestPerToolConfigSaving(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "conf-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override HOME for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create config directory
	configDir, err := ConfigDir()
	if err != nil {
		t.Fatalf("Failed to get config dir: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create empty per-tool file to signal we want per-tool format
	jjTomlPath := configDir + "/jj.toml"
	if err := os.WriteFile(jjTomlPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create jj.toml: %v", err)
	}

	// Create config with values
	config := &Config{
		Tools: map[string]ToolConfig{
			"jj": {
				Name:       "jj",
				ConfigPath: "~/.config/jj/config.toml",
				SchemaPath: "embedded://jj.json",
				Values: map[string]interface{}{
					"user.name":  "Test User",
					"user.email": "test@example.com",
				},
			},
		},
	}

	// Save config
	err = config.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify values were written to per-tool file
	jjTomlData, err := os.ReadFile(jjTomlPath)
	if err != nil {
		t.Fatalf("Failed to read jj.toml: %v", err)
	}

	jjTomlStr := string(jjTomlData)
	if !strings.Contains(jjTomlStr, "Test User") {
		t.Error("jj.toml should contain user.name value")
	}
	if !strings.Contains(jjTomlStr, "test@example.com") {
		t.Error("jj.toml should contain user.email value")
	}

	// Verify config.toml doesn't have values section
	configPath, err := ConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}
	configTomlData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config.toml: %v", err)
	}

	configTomlStr := string(configTomlData)
	if strings.Contains(configTomlStr, "Test User") {
		t.Error("config.toml should not contain values (should be in per-tool file)")
	}
}
