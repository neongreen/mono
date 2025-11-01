package main

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func makeTreeWritableOnCleanup(t *testing.T, root string) {
	t.Helper()

	t.Cleanup(func() {
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			mode := os.FileMode(0o600)
			if d.IsDir() {
				mode = 0o700
			}

			_ = os.Chmod(path, mode)
			return nil
		})
	})
}

// TestCLIIntegration tests end-to-end CLI functionality by running the actual binary
func TestCLIIntegration(t *testing.T) {
	// Build the binary for testing
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "conf")

	// Build conf binary
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "." // Current directory is already cmd/
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build conf binary: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectOut   string
		expectErr   string
	}{
		{
			name:      "help command",
			args:      []string{"--help"},
			expectOut: "smart config manager",
		},
		{
			name:      "jj help",
			args:      []string{"jj", "--help"},
			expectOut: "Get or set configuration values in ~/.config/jj/config.toml",
		},
		{
			name:      "mise help",
			args:      []string{"mise", "--help"},
			expectOut: "Get or set configuration values in ~/.config/mise/config.toml",
		},
		{
			name:      "starship help",
			args:      []string{"starship", "--help"},
			expectOut: "Get or set configuration values in ~/.config/starship.toml",
		},
		{
			name:      "completion help",
			args:      []string{"completion", "--help"},
			expectOut: "Generate shell completion scripts with intelligent schema-aware suggestions",
		},
		{
			name:      "jj without args shows list",
			args:      []string{"jj"},
			expectOut: "SETTING",
		},
		{
			name:      "mise without args shows list",
			args:      []string{"mise"},
			expectOut: "SETTING",
		},
		{
			name:      "starship without args shows list",
			args:      []string{"starship"},
			expectOut: "SETTING",
		},
		{
			name:        "invalid subcommand",
			args:        []string{"invalid"},
			expectError: true,
			expectErr:   "unknown command",
		},
		{
			name:      "completion bash generates script",
			args:      []string{"completion", "bash"},
			expectOut: "# bash completion for conf",
		},
		{
			name:        "completion invalid shell",
			args:        []string{"completion", "invalid"},
			expectError: true,
			expectOut:   "Unsupported shell", // Goes to stdout, not stderr
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create command
			cmd := exec.Command(binaryPath, tt.args...)

			// Capture output
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			// Run command
			err := cmd.Run()

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but command succeeded")
				t.Logf("stdout: %s", stdout.String())
				t.Logf("stderr: %s", stderr.String())
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected success but got error: %v", err)
				t.Logf("stdout: %s", stdout.String())
				t.Logf("stderr: %s", stderr.String())
			}

			// Check output expectations
			if tt.expectOut != "" {
				combined := stdout.String() + stderr.String()
				if !strings.Contains(combined, tt.expectOut) {
					t.Errorf("Expected output to contain %q, got:\nstdout: %s\nstderr: %s",
						tt.expectOut, stdout.String(), stderr.String())
				}
			}

			if tt.expectErr != "" {
				if !strings.Contains(stderr.String(), tt.expectErr) {
					t.Errorf("Expected stderr to contain %q, got: %s", tt.expectErr, stderr.String())
				}
			}
		})
	}
}

// TestCLIDryRunMode tests that dry-run mode works across all commands
func TestCLIDryRunMode(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	makeTreeWritableOnCleanup(t, tempHome)
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create minimal config structures
	jjConfigDir := filepath.Join(tempHome, ".config", "jj")
	miseConfigDir := filepath.Join(tempHome, ".config", "mise")
	starshipConfigDir := filepath.Join(tempHome, ".config")

	os.MkdirAll(jjConfigDir, 0755)
	os.MkdirAll(miseConfigDir, 0755)
	os.MkdirAll(starshipConfigDir, 0755)

	// Create basic config files
	jjConfig := `# JJ config for testing
[user]
name = "Original User"
`
	os.WriteFile(filepath.Join(jjConfigDir, "config.toml"), []byte(jjConfig), 0644)

	miseConfig := `# Mise config for testing
[settings]
experimental = false
`
	os.WriteFile(filepath.Join(miseConfigDir, "config.toml"), []byte(miseConfig), 0644)

	starshipConfig := `# Starship config for testing
add_newline = true
`
	os.WriteFile(filepath.Join(starshipConfigDir, "starship.toml"), []byte(starshipConfig), 0644)

	// Build the binary for testing
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "conf")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build conf binary: %v", err)
	}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "jj dry run set",
			args: []string{"--dry-run", "jj", "user.name", "New User"},
		},
		{
			name: "mise dry run set",
			args: []string{"--dry-run", "mise", "settings.experimental", "true"},
		},
		{
			name: "starship dry run set",
			args: []string{"--dry-run", "starship", "add_newline", "false"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run dry-run command
			cmd := exec.Command(binaryPath, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				t.Errorf("Dry-run command failed: %v", err)
				t.Logf("stdout: %s", stdout.String())
				t.Logf("stderr: %s", stderr.String())
			}

			// Check that output indicates dry-run mode
			combined := stdout.String() + stderr.String()
			if !strings.Contains(combined, "DRY RUN") && !strings.Contains(combined, "dry-run") {
				t.Errorf("Expected dry-run indication in output, got:\nstdout: %s\nstderr: %s",
					stdout.String(), stderr.String())
			}

			// Verify original files weren't modified by reading them
			// This ensures dry-run actually prevents changes
			originalJJConfig, _ := os.ReadFile(filepath.Join(jjConfigDir, "config.toml"))
			if !strings.Contains(string(originalJJConfig), "Original User") {
				t.Error("JJ config was modified during dry-run")
			}

			originalMiseConfig, _ := os.ReadFile(filepath.Join(miseConfigDir, "config.toml"))
			if !strings.Contains(string(originalMiseConfig), "experimental = false") {
				t.Error("Mise config was modified during dry-run")
			}

			originalStarshipConfig, _ := os.ReadFile(filepath.Join(starshipConfigDir, "starship.toml"))
			if !strings.Contains(string(originalStarshipConfig), "add_newline = true") {
				t.Error("Starship config was modified during dry-run")
			}
		})
	}
}

// TestCLIListCommands tests the listing functionality across tools
func TestCLIListCommands(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	makeTreeWritableOnCleanup(t, tempHome)
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create minimal config structures
	jjConfigDir := filepath.Join(tempHome, ".config", "jj")
	starshipConfigDir := filepath.Join(tempHome, ".config")
	os.MkdirAll(jjConfigDir, 0755)
	os.MkdirAll(starshipConfigDir, 0755)

	jjConfig := `# JJ config for testing
[user]
name = "Test User"
`
	starshipConfig := `# Starship config for testing
add_newline = true
`
	os.WriteFile(filepath.Join(jjConfigDir, "config.toml"), []byte(jjConfig), 0644)
	os.WriteFile(filepath.Join(starshipConfigDir, "starship.toml"), []byte(starshipConfig), 0644)

	// Build the binary for testing
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "conf")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build conf binary: %v", err)
	}

	tests := []struct {
		name      string
		args      []string
		expectOut []string
	}{
		{
			name: "jj list settings",
			args: []string{"jj", "--list"},
			expectOut: []string{
				"SETTING",
				"TYPE",
				"VALUE",
				"DESCRIPTION",
				"user.name",
				"string",
				"Test User",
				"Config file:",
			},
		},
		{
			name: "mise list settings",
			args: []string{"mise", "list"},
			expectOut: []string{
				"SETTING",
				"TYPE",
				"DESCRIPTION",
				"EXAMPLE",
				"settings.experimental",
				"boolean",
			},
		},
		{
			name: "starship list settings",
			args: []string{"starship", "list"},
			expectOut: []string{
				"SETTING",
				"TYPE",
				"DESCRIPTION",
				"EXAMPLE",
				"add_newline",
				"boolean",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				t.Errorf("List command failed: %v", err)
				t.Logf("stdout: %s", stdout.String())
				t.Logf("stderr: %s", stderr.String())
				return
			}

			output := stdout.String()
			for _, expected := range tt.expectOut {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got: %s", expected, output)
				}
			}
		})
	}
}

// TestCLIErrorHandling tests various error conditions
func TestCLIErrorHandling(t *testing.T) {
	// Create temporary home directory with invalid configs
	tempHome := t.TempDir()
	makeTreeWritableOnCleanup(t, tempHome)
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Build the binary for testing
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "conf")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build conf binary: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		setup       func()
		expectError bool
		expectErr   string
	}{
		{
			name: "jj invalid path",
			args: []string{"jj", "invalid.nonexistent.path", "value"},
			setup: func() {
				// Create valid jj config dir
				jjConfigDir := filepath.Join(tempHome, ".config", "jj")
				os.MkdirAll(jjConfigDir, 0755)
				os.WriteFile(filepath.Join(jjConfigDir, "config.toml"), []byte("# test"), 0644)
			},
			expectError: true,
			expectErr:   "invalid configuration path",
		},
		{
			name: "jj empty path",
			args: []string{"jj", "", "value"},
			setup: func() {
				// Already set up from previous test
			},
			expectError: true,
			expectErr:   "invalid configuration path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			cmd := exec.Command(binaryPath, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but command succeeded")
					t.Logf("stdout: %s", stdout.String())
					t.Logf("stderr: %s", stderr.String())
				}

				if tt.expectErr != "" && !strings.Contains(stderr.String(), tt.expectErr) {
					t.Errorf("Expected stderr to contain %q, got: %s", tt.expectErr, stderr.String())
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
					t.Logf("stdout: %s", stdout.String())
					t.Logf("stderr: %s", stderr.String())
				}
			}
		})
	}
}

// TestImportCommand tests the import command functionality
func TestImportCommand(t *testing.T) {
	// Build the binary for testing
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "conf")

	// Build conf binary
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build conf binary: %v", err)
	}

	// Setup test home directory
	testHome := t.TempDir()

	// Create test jj config
	jjConfigDir := filepath.Join(testHome, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create jj config dir: %v", err)
	}
	jjConfig := `[user]
name = "Import Test User"
email = "import@example.com"

[snapshot]
max-new-file-size = 2048
`
	if err := os.WriteFile(filepath.Join(jjConfigDir, "config.toml"), []byte(jjConfig), 0644); err != nil {
		t.Fatalf("Failed to write jj config: %v", err)
	}

	// Test dry-run import
	t.Run("import dry-run", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "import", "jj", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME="+testHome)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Errorf("Import dry-run failed: %v", err)
			t.Logf("stdout: %s", stdout.String())
			t.Logf("stderr: %s", stderr.String())
		}

		output := stdout.String()
		if !strings.Contains(output, "Would import: jj.user.name = Import Test User") {
			t.Errorf("Expected dry-run output to show user.name, got: %s", output)
		}
		if !strings.Contains(output, "Would import: jj.user.email = import@example.com") {
			t.Errorf("Expected dry-run output to show user.email, got: %s", output)
		}
		if !strings.Contains(output, "Would import: jj.snapshot.max-new-file-size = 2048") {
			t.Errorf("Expected dry-run output to show max-new-file-size, got: %s", output)
		}
	})

	// Test actual import
	t.Run("import actual", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "import", "jj")
		cmd.Env = append(os.Environ(), "HOME="+testHome)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Errorf("Import failed: %v", err)
			t.Logf("stdout: %s", stdout.String())
			t.Logf("stderr: %s", stderr.String())
		}

		output := stdout.String()
		if !strings.Contains(output, "✓ Imported jj.user.name = Import Test User") {
			t.Errorf("Expected import output to show user.name imported, got: %s", output)
		}
		if !strings.Contains(output, "✓ Import complete") {
			t.Errorf("Expected completion message, got: %s", output)
		}

		// Verify conf state file was created
		confStateFile := filepath.Join(testHome, ".config", "conf", "jj.toml")
		if _, err := os.Stat(confStateFile); os.IsNotExist(err) {
			t.Errorf("Expected conf state file to be created at %s", confStateFile)
		} else {
			// Read and verify content
			content, err := os.ReadFile(confStateFile)
			if err != nil {
				t.Errorf("Failed to read conf state file: %v", err)
			} else {
				contentStr := string(content)
				if !strings.Contains(contentStr, "Import Test User") {
					t.Errorf("Expected conf state to contain imported user name, got: %s", contentStr)
				}
				if !strings.Contains(contentStr, "import@example.com") {
					t.Errorf("Expected conf state to contain imported email, got: %s", contentStr)
				}
			}
		}
	})

	// Test importing jj config with quoted keys (like single-character aliases)
	t.Run("import quoted keys", func(t *testing.T) {
		// Create a new test home for this test
		testHome2 := t.TempDir()

		// Create test jj config with quoted keys
		jjConfigDir2 := filepath.Join(testHome2, ".config", "jj")
		if err := os.MkdirAll(jjConfigDir2, 0755); err != nil {
			t.Fatalf("Failed to create jj config dir: %v", err)
		}

		// Config with quoted single-character aliases (common jj pattern)
		jjConfigWithQuotedKeys := `[user]
name = "Test User"

[aliases]
"." = ["foo"]
".." = ["bar"]
normal = "status"
`
		if err := os.WriteFile(filepath.Join(jjConfigDir2, "config.toml"), []byte(jjConfigWithQuotedKeys), 0644); err != nil {
			t.Fatalf("Failed to write jj config with quoted keys: %v", err)
		}

		// Run import
		cmd := exec.Command(binaryPath, "import", "jj")
		cmd.Env = append(os.Environ(), "HOME="+testHome2)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Errorf("Import with quoted keys failed: %v", err)
			t.Logf("stdout: %s", stdout.String())
			t.Logf("stderr: %s", stderr.String())
		}

		output := stdout.String()

		// Verify the quoted keys are imported correctly
		// The output should show the quoted keys
		if !strings.Contains(output, `aliases."."`) {
			t.Errorf("Expected import output to show aliases.\".\", got: %s", output)
		}
		if !strings.Contains(output, `aliases.".."`) {
			t.Errorf("Expected import output to show aliases.\"..\", got: %s", output)
		}
		if !strings.Contains(output, "aliases.normal") {
			t.Errorf("Expected import output to show aliases.normal, got: %s", output)
		}

		// Verify conf state file contains the imported values with quoted keys
		confStateFile := filepath.Join(testHome2, ".config", "conf", "jj.toml")
		if _, err := os.Stat(confStateFile); os.IsNotExist(err) {
			t.Errorf("Expected conf state file to be created at %s", confStateFile)
		} else {
			content, err := os.ReadFile(confStateFile)
			if err != nil {
				t.Errorf("Failed to read conf state file: %v", err)
			} else {
				contentStr := string(content)
				// The conf state file should contain the values
				// Note: the exact format may vary, but the values should be there
				if !strings.Contains(contentStr, "foo") {
					t.Errorf("Expected conf state to contain 'foo' from aliases.\".\", got: %s", contentStr)
				}
				if !strings.Contains(contentStr, "bar") {
					t.Errorf("Expected conf state to contain 'bar' from aliases.\"..\", got: %s", contentStr)
				}
				if !strings.Contains(contentStr, "status") {
					t.Errorf("Expected conf state to contain 'status' from aliases.normal, got: %s", contentStr)
				}
			}
		}
	})
}

// TestApplyPreservesUnmanagedSettings tests that `conf apply` preserves
// settings in the target config that are not managed by conf
func TestApplyPreservesUnmanagedSettings(t *testing.T) {
	// Build the binary for testing
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "conf")

	// Build conf binary
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build conf binary: %v", err)
	}

	// Setup test home directory
	testHome := t.TempDir()

	// Create jj config with existing user settings
	jjConfigDir := filepath.Join(testHome, ".config", "jj")
	if err := os.MkdirAll(jjConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create jj config dir: %v", err)
	}

	// Initial config has both [user] and [ui] settings
	initialConfig := `# Existing jj config
[user]
name = "Original User"
email = "original@example.com"

[ui]
default-command = "log"
`
	if err := os.WriteFile(filepath.Join(jjConfigDir, "config.toml"), []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write jj config: %v", err)
	}

	// Create conf state directory
	confDir := filepath.Join(testHome, ".config", "conf")
	if err := os.MkdirAll(confDir, 0755); err != nil {
		t.Fatalf("Failed to create conf dir: %v", err)
	}

	// Conf state only manages ui.default-command (NOT user settings)
	confState := `[ui]
default-command = "status"
`
	if err := os.WriteFile(filepath.Join(confDir, "jj.toml"), []byte(confState), 0644); err != nil {
		t.Fatalf("Failed to write conf state: %v", err)
	}

	// Run conf apply jj
	t.Run("apply preserves unmanaged user settings", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "apply", "jj")
		cmd.Env = append(os.Environ(), "HOME="+testHome)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Errorf("Apply failed: %v", err)
			t.Logf("stdout: %s", stdout.String())
			t.Logf("stderr: %s", stderr.String())
		}

		// Read the resulting config
		resultConfig, err := os.ReadFile(filepath.Join(jjConfigDir, "config.toml"))
		if err != nil {
			t.Fatalf("Failed to read result config: %v", err)
		}
		resultStr := string(resultConfig)

		// Verify user settings are preserved
		if !strings.Contains(resultStr, `name = "Original User"`) {
			t.Errorf("Expected user.name to be preserved, got: %s", resultStr)
		}
		if !strings.Contains(resultStr, `email = "original@example.com"`) {
			t.Errorf("Expected user.email to be preserved, got: %s", resultStr)
		}

		// Verify ui.default-command was updated
		if !strings.Contains(resultStr, `default-command = "status"`) {
			t.Errorf("Expected ui.default-command to be updated to 'status', got: %s", resultStr)
		}

		// Verify old value was replaced
		if strings.Contains(resultStr, `default-command = "log"`) {
			t.Errorf("Expected old ui.default-command 'log' to be replaced, got: %s", resultStr)
		}
	})
}
