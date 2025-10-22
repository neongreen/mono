package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIIntegration tests end-to-end CLI functionality by running the actual binary
func TestCLIIntegration(t *testing.T) {
	// Build the binary for testing
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "conf")

	// Build conf binary
	cmd := exec.Command("go", "build", "-o", binaryPath, "./main.go")
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
			name:        "jj without args should show error",
			args:        []string{"jj"},
			expectError: true,
			expectErr:   "config path required",
		},
		{
			name:        "mise without args should show error",
			args:        []string{"mise"},
			expectError: true,
			expectErr:   "accepts between 1 and 2 arg(s), received 0",
		},
		{
			name:        "starship without args should show error",
			args:        []string{"starship"},
			expectError: true,
			expectErr:   "accepts between 1 and 2 arg(s), received 0",
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

	cmd := exec.Command("go", "build", "-o", binaryPath, "./main.go")
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
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create minimal config structures
	jjConfigDir := filepath.Join(tempHome, ".config", "jj")
	os.MkdirAll(jjConfigDir, 0755)

	jjConfig := `# JJ config for testing
[user]
name = "Test User"
`
	os.WriteFile(filepath.Join(jjConfigDir, "config.toml"), []byte(jjConfig), 0644)

	// Build the binary for testing
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "conf")

	cmd := exec.Command("go", "build", "-o", binaryPath, "./main.go")
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
				"All jj configuration settings:",
				"user.name",
				"Type: string",
				"Current value: Test User ✓",
				"Config file:",
			},
		},
		{
			name: "mise list settings",
			args: []string{"mise", "list"},
			expectOut: []string{
				"Common mise configuration settings:",
				"settings.experimental",
				"Type: boolean",
			},
		},
		{
			name: "starship list settings",
			args: []string{"starship", "list"},
			expectOut: []string{
				"Common starship configuration settings:",
				"add_newline",
				"Type: boolean",
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
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Build the binary for testing
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "conf")

	cmd := exec.Command("go", "build", "-o", binaryPath, "./main.go")
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
