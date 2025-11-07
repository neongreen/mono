package mise

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMiseToolRealFileOperations tests the complete workflow with real files
func TestMiseToolRealFileOperations(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create mise config directory structure
	miseConfigDir := filepath.Join(tempHome, ".config", "mise")
	if err := os.MkdirAll(miseConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create mise config directory: %v", err)
	}

	// Test 1: Create new config file from scratch
	t.Run("create new config file", func(t *testing.T) {
		configPath := filepath.Join(miseConfigDir, "config.toml")

		// Ensure file doesn't exist
		os.Remove(configPath)

		// Create tool
		tool, err := NewMiseTool()
		if err != nil {
			t.Fatalf("Failed to create Mise tool: %v", err)
		}

		// Set experimental feature
		err = tool.SetConfig("settings.experimental", true)
		if err != nil {
			t.Fatalf("Failed to set settings.experimental: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}

		// Verify content
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "experimental") {
			t.Errorf("Config file does not contain expected setting, got: %s", contentStr)
		}

		// Verify we can read it back
		value, err := tool.GetConfig("settings.experimental")
		if err != nil {
			t.Fatalf("Failed to get settings.experimental: %v", err)
		}
		if value != true {
			t.Errorf("Expected true, got %v", value)
		}
	})

	// Test 2: Modify existing config file
	t.Run("modify existing config file", func(t *testing.T) {
		configPath := filepath.Join(miseConfigDir, "config.toml")

		// Create initial config
		initialConfig := `# Mise configuration
[settings]
experimental = false
jobs = 4
verbose = true

[env]
NODE_ENV = "development"
`
		if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
			t.Fatalf("Failed to write initial config: %v", err)
		}

		// Create tool
		tool, err := NewMiseTool()
		if err != nil {
			t.Fatalf("Failed to create Mise tool: %v", err)
		}

		// Modify experimental setting while preserving other values
		err = tool.SetConfig("settings.experimental", true)
		if err != nil {
			t.Fatalf("Failed to modify settings.experimental: %v", err)
		}

		// Verify through tool interface
		experimental, err := tool.GetConfig("settings.experimental")
		if err != nil {
			t.Fatalf("Failed to get settings.experimental: %v", err)
		}
		if experimental != true {
			t.Errorf("Expected true, got %v", experimental)
		}

		// Check that other values are preserved
		jobs, err := tool.GetConfig("settings.jobs")
		if err != nil {
			t.Fatalf("Failed to get settings.jobs: %v", err)
		}
		if jobs != int64(4) {
			t.Errorf("Expected 4, got %v", jobs)
		}

		verbose, err := tool.GetConfig("settings.verbose")
		if err != nil {
			t.Fatalf("Failed to get settings.verbose: %v", err)
		}
		if verbose != true {
			t.Errorf("Expected true, got %v", verbose)
		}

		nodeEnv, err := tool.GetConfig("env.NODE_ENV")
		if err != nil {
			t.Fatalf("Failed to get env.NODE_ENV: %v", err)
		}
		if nodeEnv != "development" {
			t.Errorf("Expected 'development', got %v", nodeEnv)
		}
	})

	// Test 3: Add nested configuration
	t.Run("add nested configuration", func(t *testing.T) {
		configPath := filepath.Join(miseConfigDir, "config.toml")

		// Start with basic config
		initialConfig := `[settings]
experimental = true
`
		if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
			t.Fatalf("Failed to write initial config: %v", err)
		}

		// Create tool
		tool, err := NewMiseTool()
		if err != nil {
			t.Fatalf("Failed to create Mise tool: %v", err)
		}

		// Add new settings
		err = tool.SetConfig("settings.jobs", 8)
		if err != nil {
			t.Fatalf("Failed to set settings.jobs: %v", err)
		}

		err = tool.SetConfig("settings.verbose", false)
		if err != nil {
			t.Fatalf("Failed to set settings.verbose: %v", err)
		}

		// Add environment variables section
		err = tool.SetConfig("env.PATH", "/custom/bin")
		if err != nil {
			t.Fatalf("Failed to set env.PATH: %v", err)
		}

		// Verify all values
		jobs, err := tool.GetConfig("settings.jobs")
		if err != nil || jobs != int64(8) {
			t.Errorf("Expected jobs=8, got %v (error: %v)", jobs, err)
		}

		verbose, err := tool.GetConfig("settings.verbose")
		if err != nil || verbose != false {
			t.Errorf("Expected verbose=false, got %v (error: %v)", verbose, err)
		}

		path, err := tool.GetConfig("env.PATH")
		if err != nil || path != "/custom/bin" {
			t.Errorf("Expected PATH='/custom/bin', got %v (error: %v)", path, err)
		}

		// Verify original experimental setting is preserved
		experimental, err := tool.GetConfig("settings.experimental")
		if err != nil || experimental != true {
			t.Errorf("Original experimental setting was lost, got %v (error: %v)", experimental, err)
		}
	})

	// Test 4: Unset configuration values
	t.Run("unset configuration values", func(t *testing.T) {
		configPath := filepath.Join(miseConfigDir, "config.toml")

		// Start with multiple values
		initialConfig := `[settings]
experimental = true
jobs = 4
verbose = true

[env]
NODE_ENV = "development"
DEBUG = "true"
`
		if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
			t.Fatalf("Failed to write initial config: %v", err)
		}

		// Create tool
		tool, err := NewMiseTool()
		if err != nil {
			t.Fatalf("Failed to create Mise tool: %v", err)
		}

		// Unset verbose setting
		err = tool.UnsetConfig("settings.verbose")
		if err != nil {
			t.Fatalf("Failed to unset settings.verbose: %v", err)
		}

		// Verify verbose is gone
		_, err = tool.GetConfig("settings.verbose")
		if err == nil {
			t.Error("Expected error when getting unset settings.verbose")
		}

		// Verify other values are preserved
		experimental, err := tool.GetConfig("settings.experimental")
		if err != nil || experimental != true {
			t.Errorf("settings.experimental was affected by unset operation, got %v (error: %v)", experimental, err)
		}

		jobs, err := tool.GetConfig("settings.jobs")
		if err != nil || jobs != int64(4) {
			t.Errorf("settings.jobs was affected by unset operation, got %v (error: %v)", jobs, err)
		}

		nodeEnv, err := tool.GetConfig("env.NODE_ENV")
		if err != nil || nodeEnv != "development" {
			t.Errorf("env.NODE_ENV was affected by unset operation, got %v (error: %v)", nodeEnv, err)
		}
	})

	// Test 5: Dry run mode doesn't modify files
	t.Run("dry run mode", func(t *testing.T) {
		configPath := filepath.Join(miseConfigDir, "config.toml")

		// Create initial config
		initialConfig := `[settings]
experimental = false
`
		if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
			t.Fatalf("Failed to write initial config: %v", err)
		}

		// Get initial file modification time
		initialStat, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("Failed to stat config file: %v", err)
		}

		// Create tool with dry-run enabled
		tool, err := NewMiseToolWithDryRun(true)
		if err != nil {
			t.Fatalf("Failed to create Mise tool with dry-run: %v", err)
		}

		// Attempt to modify in dry-run mode
		err = tool.SetConfig("settings.experimental", true)
		if err != nil {
			t.Fatalf("Dry-run SetConfig failed: %v", err)
		}

		// Verify file was not modified
		finalStat, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("Failed to stat config file after dry-run: %v", err)
		}

		if finalStat.ModTime() != initialStat.ModTime() {
			t.Error("File was modified during dry-run")
		}

		// Verify content is unchanged
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		if !strings.Contains(string(content), "experimental = false") {
			t.Errorf("File content was changed during dry-run, got: %s", string(content))
		}
		if strings.Contains(string(content), "experimental = true") {
			t.Error("Dry-run changes were written to file")
		}
	})

	// Test 6: Various data types
	t.Run("various data types", func(t *testing.T) {
		configPath := filepath.Join(miseConfigDir, "config.toml")

		// Remove any existing config
		os.Remove(configPath)

		// Create tool
		tool, err := NewMiseTool()
		if err != nil {
			t.Fatalf("Failed to create Mise tool: %v", err)
		}

		// Test boolean
		err = tool.SetConfig("settings.experimental", true)
		if err != nil {
			t.Fatalf("Failed to set boolean: %v", err)
		}

		// Test integer
		err = tool.SetConfig("settings.jobs", 6)
		if err != nil {
			t.Fatalf("Failed to set integer: %v", err)
		}

		// Test string
		err = tool.SetConfig("env.NODE_ENV", "production")
		if err != nil {
			t.Fatalf("Failed to set string: %v", err)
		}

		// Verify all types
		experimental, err := tool.GetConfig("settings.experimental")
		if err != nil || experimental != true {
			t.Errorf("Boolean not preserved: %v (error: %v)", experimental, err)
		}

		jobs, err := tool.GetConfig("settings.jobs")
		if err != nil || jobs != int64(6) {
			t.Errorf("Integer not preserved: %v (error: %v)", jobs, err)
		}

		nodeEnv, err := tool.GetConfig("env.NODE_ENV")
		if err != nil || nodeEnv != "production" {
			t.Errorf("String not preserved: %v (error: %v)", nodeEnv, err)
		}
	})
}

// TestMiseToolConfigPathValidation tests path validation behavior
func TestMiseToolConfigPathValidation(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create mise config directory
	miseConfigDir := filepath.Join(tempHome, ".config", "mise")
	if err := os.MkdirAll(miseConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create mise config directory: %v", err)
	}

	configPath := filepath.Join(miseConfigDir, "config.toml")

	// Create basic config
	initialConfig := `[settings]
experimental = false
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Create tool
	tool, err := NewMiseTool()
	if err != nil {
		t.Fatalf("Failed to create Mise tool: %v", err)
	}

	// Test valid common paths (mise is flexible, most paths should work)
	validPaths := []string{
		"settings.experimental",
		"settings.jobs",
		"settings.verbose",
		"env.NODE_ENV",
		"env.PATH",
		"tasks.build.cmd",
	}

	for _, path := range validPaths {
		err := tool.SetConfig(path, "test-value")
		if err != nil {
			t.Errorf("Expected path '%s' to be valid, got error: %v", path, err)
		}
	}

	// Test invalid path (empty string)
	err = tool.SetConfig("", "value")
	if err == nil {
		t.Error("Expected error for empty path")
	}
}

// TestMiseToolConcurrentAccess tests behavior with multiple tool instances
func TestMiseToolConcurrentAccess(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create mise config directory
	miseConfigDir := filepath.Join(tempHome, ".config", "mise")
	if err := os.MkdirAll(miseConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create mise config directory: %v", err)
	}

	configPath := filepath.Join(miseConfigDir, "config.toml")

	// Create initial config
	initialConfig := `[settings]
experimental = false
jobs = 2

[env]
NODE_ENV = "test"
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Create two tool instances
	tool1, err := NewMiseTool()
	if err != nil {
		t.Fatalf("Failed to create first Mise tool: %v", err)
	}

	tool2, err := NewMiseTool()
	if err != nil {
		t.Fatalf("Failed to create second Mise tool: %v", err)
	}

	// Make modifications with both tools
	err = tool1.SetConfig("settings.experimental", true)
	if err != nil {
		t.Fatalf("Tool1 failed to set config: %v", err)
	}

	err = tool2.SetConfig("settings.jobs", 8)
	if err != nil {
		t.Fatalf("Tool2 failed to set config: %v", err)
	}

	// Verify the file is still valid TOML
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read final config: %v", err)
	}

	contentStr := string(content)
	if contentStr == "" {
		t.Error("Config file is empty after concurrent access")
	}

	// Create new tool to read final state
	finalTool, err := NewMiseTool()
	if err != nil {
		t.Fatalf("Failed to create final Mise tool: %v", err)
	}

	// Test that we can still read values (file wasn't corrupted)
	experimental, _ := finalTool.GetConfig("settings.experimental")
	jobs, _ := finalTool.GetConfig("settings.jobs")
	nodeEnv, _ := finalTool.GetConfig("env.NODE_ENV")

	if experimental == nil && jobs == nil && nodeEnv == nil {
		t.Error("All configuration values were lost")
	}

	// At least the original NODE_ENV should be preserved
	if nodeEnv != "test" {
		t.Errorf("Original NODE_ENV was lost, got: %v", nodeEnv)
	}

	t.Logf("Final config - experimental: %v, jobs: %v, NODE_ENV: %v", experimental, jobs, nodeEnv)
	t.Logf("Final file content:\n%s", contentStr)
}
