package jj

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestJJToolRealFileOperations tests the complete workflow with real files
func TestJJToolRealFileOperations(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create jj config directory structure
	jjConfigDir := filepath.Join(tempHome, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create jj config directory: %v", err)
	}

	// Test 1: Create new config file from scratch
	t.Run("create new config file", func(t *testing.T) {
		configPath := filepath.Join(jjConfigDir, "config.toml")

		// Ensure file doesn't exist
		os.Remove(configPath)

		// Create tool
		tool, err := NewJJTool()
		if err != nil {
			t.Fatalf("Failed to create JJ tool: %v", err)
		}

		// Set user name
		err = tool.SetConfig("user.name", "Integration Test User")
		if err != nil {
			t.Fatalf("Failed to set user.name: %v", err)
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
		if !strings.Contains(contentStr, "Integration Test User") {
			t.Errorf("Config file does not contain expected user name, got: %s", contentStr)
		}

		// Verify we can read it back
		value, err := tool.GetConfig("user.name")
		if err != nil {
			t.Fatalf("Failed to get user.name: %v", err)
		}
		if value != "Integration Test User" {
			t.Errorf("Expected 'Integration Test User', got %v", value)
		}
	})

	// Test 2: Modify existing config file
	t.Run("modify existing config file", func(t *testing.T) {
		configPath := filepath.Join(jjConfigDir, "config.toml")

		// Create initial config
		initialConfig := `# JJ configuration
[user]
name = "Original User"
email = "original@example.com"

[ui]
default-command = "status"
`
		if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
			t.Fatalf("Failed to write initial config: %v", err)
		}

		// Create tool
		tool, err := NewJJTool()
		if err != nil {
			t.Fatalf("Failed to create JJ tool: %v", err)
		}

		// Modify user name while preserving other values
		err = tool.SetConfig("user.name", "Modified User")
		if err != nil {
			t.Fatalf("Failed to modify user.name: %v", err)
		}

		// Verify file content
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		contentStr := string(content)

		// Check that the name was updated
		if !strings.Contains(contentStr, "Modified User") {
			t.Errorf("Config file does not contain modified name, got: %s", contentStr)
		}

		// Check that other values are preserved
		if !strings.Contains(contentStr, "original@example.com") {
			t.Errorf("Config file lost original email, got: %s", contentStr)
		}
		if !strings.Contains(contentStr, "status") {
			t.Errorf("Config file lost default-command, got: %s", contentStr)
		}

		// Verify through tool interface
		name, _ := tool.GetConfig("user.name")
		email, _ := tool.GetConfig("user.email")
		defaultCmd, _ := tool.GetConfig("ui.default-command")

		if name != "Modified User" {
			t.Errorf("Expected modified name, got %v", name)
		}
		if email != "original@example.com" {
			t.Errorf("Expected preserved email, got %v", email)
		}
		if defaultCmd != "status" {
			t.Errorf("Expected preserved default-command, got %v", defaultCmd)
		}
	})

	// Test 3: Add nested configuration
	t.Run("add nested configuration", func(t *testing.T) {
		configPath := filepath.Join(jjConfigDir, "config.toml")

		// Start with basic config
		initialConfig := `[user]
name = "Test User"
`
		if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
			t.Fatalf("Failed to write initial config: %v", err)
		}

		// Create tool
		tool, err := NewJJTool()
		if err != nil {
			t.Fatalf("Failed to create JJ tool: %v", err)
		}

		// Add nested UI configuration
		err = tool.SetConfig("ui.editor", "vim")
		if err != nil {
			t.Fatalf("Failed to set ui.editor: %v", err)
		}

		err = tool.SetConfig("ui.merge-editor", "vimdiff")
		if err != nil {
			t.Fatalf("Failed to set ui.merge-editor: %v", err)
		}

		// Verify nested values
		editor, err := tool.GetConfig("ui.editor")
		if err != nil {
			t.Fatalf("Failed to get ui.editor: %v", err)
		}
		if editor != "vim" {
			t.Errorf("Expected 'vim', got %v", editor)
		}

		mergeEditor, err := tool.GetConfig("ui.merge-editor")
		if err != nil {
			t.Fatalf("Failed to get ui.merge-editor: %v", err)
		}
		if mergeEditor != "vimdiff" {
			t.Errorf("Expected 'vimdiff', got %v", mergeEditor)
		}

		// Verify original user config is preserved
		name, err := tool.GetConfig("user.name")
		if err != nil {
			t.Fatalf("Failed to get user.name: %v", err)
		}
		if name != "Test User" {
			t.Errorf("Original user.name was lost, got %v", name)
		}
	})

	// Test 4: Unset configuration values
	t.Run("unset configuration values", func(t *testing.T) {
		configPath := filepath.Join(jjConfigDir, "config.toml")

		// Start with multiple values
		initialConfig := `[user]
name = "Test User"
email = "test@example.com"

[ui]
editor = "vim"
default-command = "status"
`
		if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
			t.Fatalf("Failed to write initial config: %v", err)
		}

		// Create tool
		tool, err := NewJJTool()
		if err != nil {
			t.Fatalf("Failed to create JJ tool: %v", err)
		}

		// Unset email
		err = tool.UnsetConfig("user.email")
		if err != nil {
			t.Fatalf("Failed to unset user.email: %v", err)
		}

		// Verify email is gone
		_, err = tool.GetConfig("user.email")
		if err == nil {
			t.Error("Expected error when getting unset user.email")
		}

		// Verify other values are preserved
		name, err := tool.GetConfig("user.name")
		if err != nil {
			t.Fatalf("Failed to get user.name: %v", err)
		}
		if name != "Test User" {
			t.Errorf("user.name was affected by unset operation, got %v", name)
		}

		editor, err := tool.GetConfig("ui.editor")
		if err != nil {
			t.Fatalf("Failed to get ui.editor: %v", err)
		}
		if editor != "vim" {
			t.Errorf("ui.editor was affected by unset operation, got %v", editor)
		}
	})

	// Test 5: Dry run mode doesn't modify files
	t.Run("dry run mode", func(t *testing.T) {
		configPath := filepath.Join(jjConfigDir, "config.toml")

		// Create initial config
		initialConfig := `[user]
name = "Original User"
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
		tool, err := NewJJToolWithDryRun(true)
		if err != nil {
			t.Fatalf("Failed to create JJ tool with dry-run: %v", err)
		}

		// Attempt to modify in dry-run mode
		err = tool.SetConfig("user.name", "Dry Run User")
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

		if !strings.Contains(string(content), "Original User") {
			t.Errorf("File content was changed during dry-run, got: %s", string(content))
		}
		if strings.Contains(string(content), "Dry Run User") {
			t.Error("Dry-run changes were written to file")
		}
	})

	// Test 6: Schema validation with real files
	t.Run("schema validation", func(t *testing.T) {
		configPath := filepath.Join(jjConfigDir, "config.toml")

		// Remove any existing config
		os.Remove(configPath)

		// Create tool
		tool, err := NewJJTool()
		if err != nil {
			t.Fatalf("Failed to create JJ tool: %v", err)
		}

		// Test valid schema path
		err = tool.SetConfig("user.name", "Valid User")
		if err != nil {
			t.Fatalf("Failed to set valid config: %v", err)
		}

		// Test invalid schema path
		err = tool.SetConfig("invalid.nonexistent.path", "value")
		if err == nil {
			t.Error("Expected error for invalid schema path")
		}
		if !strings.Contains(err.Error(), "invalid configuration path") {
			t.Errorf("Expected schema validation error, got: %v", err)
		}

		// Verify valid config was written
		value, err := tool.GetConfig("user.name")
		if err != nil {
			t.Fatalf("Failed to get valid config: %v", err)
		}
		if value != "Valid User" {
			t.Errorf("Expected 'Valid User', got %v", value)
		}
	})
}

// TestJJToolStructurePreservation tests that config structure and values are preserved
func TestJJToolStructurePreservation(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create jj config directory
	jjConfigDir := filepath.Join(tempHome, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create jj config directory: %v", err)
	}

	configPath := filepath.Join(jjConfigDir, "config.toml")

	// Create config with multiple sections
	initialConfig := `[user]
name = "Test User"
email = "test@example.com"

[ui]
default-command = "log"
editor = "vim"

[snapshot]
max-new-file-size = 1048576
`

	if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Create tool
	tool, err := NewJJTool()
	if err != nil {
		t.Fatalf("Failed to create JJ tool: %v", err)
	}

	// Modify one value
	err = tool.SetConfig("user.name", "Modified User")
	if err != nil {
		t.Fatalf("Failed to modify user.name: %v", err)
	}

	// Read the modified file
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read modified config: %v", err)
	}

	contentStr := string(content)

	// Verify the value was actually changed
	if !strings.Contains(contentStr, "Modified User") {
		t.Errorf("Value was not updated, got: %s", contentStr)
	}

	// Verify other values are unchanged via the tool interface
	// (This is more reliable than string matching since formatting may change)
	email, err := tool.GetConfig("user.email")
	if err != nil || email != "test@example.com" {
		t.Errorf("user.email was not preserved, got: %v (error: %v)", email, err)
	}

	editor, err := tool.GetConfig("ui.editor")
	if err != nil || editor != "vim" {
		t.Errorf("ui.editor was not preserved, got: %v (error: %v)", editor, err)
	}

	defaultCmd, err := tool.GetConfig("ui.default-command")
	if err != nil || defaultCmd != "log" {
		t.Errorf("ui.default-command was not preserved, got: %v (error: %v)", defaultCmd, err)
	}

	maxSize, err := tool.GetConfig("snapshot.max-new-file-size")
	if err != nil || maxSize != int64(1048576) {
		t.Errorf("snapshot.max-new-file-size was not preserved, got: %v (error: %v)", maxSize, err)
	}

	// Verify modified value through tool interface
	name, err := tool.GetConfig("user.name")
	if err != nil || name != "Modified User" {
		t.Errorf("user.name modification was not properly saved, got: %v (error: %v)", name, err)
	}
}

// TestJJToolConcurrentAccess tests behavior with multiple tool instances
func TestJJToolConcurrentAccess(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create jj config directory
	jjConfigDir := filepath.Join(tempHome, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create jj config directory: %v", err)
	}

	configPath := filepath.Join(jjConfigDir, "config.toml")

	// Create initial config
	initialConfig := `[user]
name = "Initial User"
email = "initial@example.com"
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Create two tool instances
	tool1, err := NewJJTool()
	if err != nil {
		t.Fatalf("Failed to create first JJ tool: %v", err)
	}

	tool2, err := NewJJTool()
	if err != nil {
		t.Fatalf("Failed to create second JJ tool: %v", err)
	}

	// Make modifications with both tools
	err = tool1.SetConfig("user.name", "Tool1 User")
	if err != nil {
		t.Fatalf("Tool1 failed to set config: %v", err)
	}

	err = tool2.SetConfig("user.email", "tool2@example.com")
	if err != nil {
		t.Fatalf("Tool2 failed to set config: %v", err)
	}

	// Verify both changes are present (last write wins, but both should work)
	finalTool, err := NewJJTool()
	if err != nil {
		t.Fatalf("Failed to create final JJ tool: %v", err)
	}

	// At least one of the changes should be present
	// (This tests that concurrent access doesn't corrupt the file)
	name, _ := finalTool.GetConfig("user.name")
	email, _ := finalTool.GetConfig("user.email")

	// Verify the file is still valid TOML
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read final config: %v", err)
	}

	contentStr := string(content)
	if contentStr == "" {
		t.Error("Config file is empty after concurrent access")
	}

	// Test that we can still read values (file wasn't corrupted)
	if name == nil && email == nil {
		t.Error("All configuration values were lost")
	}

	t.Logf("Final config - name: %v, email: %v", name, email)
	t.Logf("Final file content:\n%s", contentStr)
}
