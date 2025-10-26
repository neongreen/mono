package mise

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neongreen/mono/conf/pkg/config"
)

func TestNewMiseTool(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mise-tool-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override HOME to use temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create conf config with mise tool configuration
	conf := config.DefaultConfig()
	conf.SetTool("mise", config.ToolConfig{
		Name:       "mise",
		ConfigPath: filepath.Join(tempDir, ".config", "mise", "config.toml"),
		SchemaPath: "embedded://mise.schema",
	})

	err = conf.Save()
	if err != nil {
		t.Fatalf("Failed to save conf config: %v", err)
	}

	// Test creating mise tool
	miseTool, err := NewMiseTool()
	if err != nil {
		t.Fatalf("Failed to create mise tool: %v", err)
	}

	if miseTool == nil {
		t.Fatal("Mise tool should not be nil")
	}

	if miseTool.configPath == "" {
		t.Error("Mise tool should have config path set")
	}
}

func TestMiseTool_SetAndGetConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mise-tool-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override HOME to use temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create conf config
	conf := config.DefaultConfig()
	conf.SetTool("mise", config.ToolConfig{
		Name:       "mise",
		ConfigPath: filepath.Join(tempDir, ".config", "mise", "config.toml"),
		SchemaPath: "embedded://mise.schema",
	})
	err = conf.Save()
	if err != nil {
		t.Fatalf("Failed to save conf config: %v", err)
	}

	// Create mise tool
	miseTool, err := NewMiseTool()
	if err != nil {
		t.Fatalf("Failed to create mise tool: %v", err)
	}

	// Test setting valid configuration
	err = miseTool.SetConfig("settings.experimental", true)
	if err != nil {
		t.Fatalf("Failed to set settings.experimental: %v", err)
	}

	err = miseTool.SetConfig("tools.node", "20")
	if err != nil {
		t.Fatalf("Failed to set tools.node: %v", err)
	}

	// Test getting configuration
	experimental, err := miseTool.GetConfig("settings.experimental")
	if err != nil {
		t.Fatalf("Failed to get settings.experimental: %v", err)
	}
	if experimental != true {
		t.Errorf("Expected true, got %v", experimental)
	}

	nodeVersion, err := miseTool.GetConfig("tools.node")
	if err != nil {
		t.Fatalf("Failed to get tools.node: %v", err)
	}
	if nodeVersion != "20" {
		t.Errorf("Expected '20', got %v", nodeVersion)
	}
}

func TestMiseTool_UnsetConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mise-tool-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override HOME to use temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create conf config
	conf := config.DefaultConfig()
	conf.SetTool("mise", config.ToolConfig{
		Name:       "mise",
		ConfigPath: filepath.Join(tempDir, ".config", "mise", "config.toml"),
		SchemaPath: "embedded://mise.schema",
	})
	err = conf.Save()
	if err != nil {
		t.Fatalf("Failed to save conf config: %v", err)
	}

	// Create mise tool
	miseTool, err := NewMiseTool()
	if err != nil {
		t.Fatalf("Failed to create mise tool: %v", err)
	}

	// Set a value first
	err = miseTool.SetConfig("settings.verbose", true)
	if err != nil {
		t.Fatalf("Failed to set settings.verbose: %v", err)
	}

	// Verify it was set
	value, err := miseTool.GetConfig("settings.verbose")
	if err != nil {
		t.Fatalf("Failed to get settings.verbose: %v", err)
	}
	if value != true {
		t.Errorf("Expected true, got %v", value)
	}

	// Unset the value
	err = miseTool.UnsetConfig("settings.verbose")
	if err != nil {
		t.Fatalf("Failed to unset settings.verbose: %v", err)
	}

	// Verify it was unset
	_, err = miseTool.GetConfig("settings.verbose")
	if err == nil {
		t.Error("Should get error for unset config")
	}
}

func TestMiseTool_ValidatePath(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mise-tool-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override HOME to use temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create conf config
	conf := config.DefaultConfig()
	conf.SetTool("mise", config.ToolConfig{
		Name:       "mise",
		ConfigPath: filepath.Join(tempDir, ".config", "mise", "config.toml"),
		SchemaPath: "embedded://mise.schema",
	})
	err = conf.Save()
	if err != nil {
		t.Fatalf("Failed to save conf config: %v", err)
	}

	// Create mise tool
	miseTool, err := NewMiseTool()
	if err != nil {
		t.Fatalf("Failed to create mise tool: %v", err)
	}

	// Test valid paths
	validPaths := []string{
		"",
		"tools",
		"settings",
		"env",
		"tasks",
		"tools.node",
		"settings.experimental",
		"env.NODE_ENV",
		"tasks.dev.run",
	}

	for _, path := range validPaths {
		if !miseTool.ValidatePath(path) {
			t.Errorf("Path '%s' should be valid", path)
		}
	}

	// Test invalid paths
	invalidPaths := []string{
		"completely.invalid.path",
		"unknown.section",
	}

	for _, path := range invalidPaths {
		if miseTool.ValidatePath(path) {
			t.Errorf("Path '%s' should be invalid", path)
		}
	}
}

func TestMiseTool_CompletionOptions(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mise-tool-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override HOME to use temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create conf config
	conf := config.DefaultConfig()
	conf.SetTool("mise", config.ToolConfig{
		Name:       "mise",
		ConfigPath: filepath.Join(tempDir, ".config", "mise", "config.toml"),
		SchemaPath: "embedded://mise.schema",
	})
	err = conf.Save()
	if err != nil {
		t.Fatalf("Failed to save conf config: %v", err)
	}

	// Create mise tool
	miseTool, err := NewMiseTool()
	if err != nil {
		t.Fatalf("Failed to create mise tool: %v", err)
	}

	// Test top-level completion
	options := miseTool.GetCompletionOptions("")
	if len(options) == 0 {
		t.Error("Should get completion options for top level")
	}

	// Check for expected top-level options
	var foundTasks, foundSettings bool
	for _, option := range options {
		if option.Name == "tasks" {
			foundTasks = true
		}
		if option.Name == "settings" {
			foundSettings = true
		}
	}

	if !foundTasks {
		t.Error("Should find 'tasks' in top-level options")
	}
	if !foundSettings {
		t.Error("Should find 'settings' in top-level options")
	}

	// Test nested completion
	settingsOptions := miseTool.GetCompletionOptions("settings")
	if len(settingsOptions) == 0 {
		t.Error("Should get completion options for settings")
	}
}

func TestMiseTool_ListCommonSettings(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mise-tool-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override HOME to use temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create conf config
	conf := config.DefaultConfig()
	conf.SetTool("mise", config.ToolConfig{
		Name:       "mise",
		ConfigPath: filepath.Join(tempDir, ".config", "mise", "config.toml"),
		SchemaPath: "embedded://mise.schema",
	})
	err = conf.Save()
	if err != nil {
		t.Fatalf("Failed to save conf config: %v", err)
	}

	// Create mise tool
	miseTool, err := NewMiseTool()
	if err != nil {
		t.Fatalf("Failed to create mise tool: %v", err)
	}

	// Test listing common settings
	settings := miseTool.ListCommonSettings()
	if len(settings) == 0 {
		t.Error("Should get some common settings")
	}

	// Check for expected settings
	var foundExperimental bool
	for _, setting := range settings {
		if setting.Path == "settings.experimental" {
			foundExperimental = true
			if setting.Type != "boolean" {
				t.Errorf("settings.experimental should be boolean type, got %s", setting.Type)
			}
			if setting.Description == "" {
				t.Error("settings.experimental should have description")
			}
			if setting.Example == "" {
				t.Error("settings.experimental should have example")
			}
		}
	}

	if !foundExperimental {
		t.Error("Should find settings.experimental in common settings")
	}
}

func TestMiseTool_ConfigPath(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mise-tool-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override HOME to use temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create conf config
	conf := config.DefaultConfig()
	configPath := filepath.Join(tempDir, ".config", "mise", "config.toml")
	conf.SetTool("mise", config.ToolConfig{
		Name:       "mise",
		ConfigPath: configPath,
		SchemaPath: "embedded://mise.schema",
	})
	err = conf.Save()
	if err != nil {
		t.Fatalf("Failed to save conf config: %v", err)
	}

	// Create mise tool
	miseTool, err := NewMiseTool()
	if err != nil {
		t.Fatalf("Failed to create mise tool: %v", err)
	}

	// Test config path
	if miseTool.GetConfigPath() != configPath {
		t.Errorf("Expected config path '%s', got '%s'", configPath, miseTool.GetConfigPath())
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{"single", []string{"single"}},
		{"dotted.path", []string{"dotted", "path"}},
		{"deeply.nested.path.here", []string{"deeply", "nested", "path", "here"}},
	}

	for _, test := range tests {
		result := splitPath(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("For input '%s', expected %d parts, got %d", test.input, len(test.expected), len(result))
			continue
		}

		for i, part := range result {
			if part != test.expected[i] {
				t.Errorf("For input '%s', expected part %d to be '%s', got '%s'", test.input, i, test.expected[i], part)
			}
		}
	}
}
