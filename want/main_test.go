package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestParseMonoProjectVersion(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantProject    string
		wantVersion    string
		wantListFormat bool
	}{
		{
			name:           "project with version",
			input:          "printpdf@main.1",
			wantProject:    "printpdf",
			wantVersion:    "main.1",
			wantListFormat: false,
		},
		{
			name:           "project without version",
			input:          "printpdf",
			wantProject:    "printpdf",
			wantVersion:    "",
			wantListFormat: true,
		},
		{
			name:           "project with pr version",
			input:          "dissect@pr-42.1",
			wantProject:    "dissect",
			wantVersion:    "pr-42.1",
			wantListFormat: false,
		},
		{
			name:           "project with pr version without suffix",
			input:          "want@pr-100",
			wantProject:    "want",
			wantVersion:    "pr-100",
			wantListFormat: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.Split(tt.input, "@")

			gotProject := parts[0]
			if gotProject != tt.wantProject {
				t.Errorf("project = %v, want %v", gotProject, tt.wantProject)
			}

			if len(parts) == 1 {
				if !tt.wantListFormat {
					t.Errorf("expected list format but got version format")
				}
				return
			}

			if len(parts) == 2 {
				gotVersion := parts[1]
				if gotVersion != tt.wantVersion {
					t.Errorf("version = %v, want %v", gotVersion, tt.wantVersion)
				}
			}
		})
	}
}

func TestFormatMonoTag(t *testing.T) {
	tests := []struct {
		name    string
		project string
		version string
		want    string
	}{
		{
			name:    "main release",
			project: "printpdf",
			version: "main.1",
			want:    "printpdf--main.1",
		},
		{
			name:    "pr release",
			project: "dissect",
			version: "pr-42.1",
			want:    "dissect--pr-42.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.project + "--" + tt.version
			if got != tt.want {
				t.Errorf("tag = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractVersionFromTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		project string
		want    string
	}{
		{
			name:    "main release",
			tag:     "printpdf--main.1",
			project: "printpdf",
			want:    "main.1",
		},
		{
			name:    "pr release",
			tag:     "dissect--pr-42.1",
			project: "dissect",
			want:    "pr-42.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := tt.project + "--"
			got := strings.TrimPrefix(tt.tag, prefix)
			if got != tt.want {
				t.Errorf("version = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePRNumber(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		wantPR    int
		wantError bool
	}{
		{
			name:      "pr with suffix",
			version:   "pr-42.1",
			wantPR:    42,
			wantError: false,
		},
		{
			name:      "pr without suffix",
			version:   "pr-100",
			wantPR:    100,
			wantError: false,
		},
		{
			name:      "not a pr version",
			version:   "main.1",
			wantPR:    0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.HasPrefix(tt.version, "pr-") {
				if !tt.wantError {
					t.Errorf("expected error for non-PR version %s", tt.version)
				}
				return
			}

			prStr := strings.TrimPrefix(tt.version, "pr-")
			parts := strings.Split(prStr, ".")
			var prNumber int
			n, err := fmt.Sscanf(parts[0], "%d", &prNumber)

			if tt.wantError && err == nil && n == 1 {
				t.Errorf("expected error but got none")
			}
			if !tt.wantError && (err != nil || n != 1) {
				t.Errorf("unexpected error: %v, n=%d", err, n)
			}
			if !tt.wantError && prNumber != tt.wantPR {
				t.Errorf("PR number = %v, want %v", prNumber, tt.wantPR)
			}
		})
	}
}

func TestCreateGoBuildCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantPath string // Either "go" or "mise"
	}{
		{
			name:     "build with output flag",
			args:     []string{"build", "-o", "/tmp/test", "."},
			wantPath: "", // Will be either "go" or "mise" depending on PATH
		},
		{
			name:     "simple build",
			args:     []string{"build"},
			wantPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createGoBuildCommand(tt.args...)

			// Check that we got a command
			if cmd == nil {
				t.Fatalf("createGoBuildCommand returned nil")
			}

			// Verify the command is either "go" or "mise exec go@..."
			if cmd.Path != "" {
				basePath := cmd.Path
				if !strings.Contains(basePath, "go") && !strings.Contains(basePath, "mise") {
					t.Errorf("unexpected command path: %s", basePath)
				}
			}

			// If go is available, should use go directly
			if isToolAvailable("go") {
				// When go is in PATH, args should be the original args
				if len(cmd.Args) < 2 {
					t.Fatalf("expected at least 2 args, got %d: %v", len(cmd.Args), cmd.Args)
				}
				// First arg is the command name (go or full path to go)
				if !strings.Contains(cmd.Args[0], "go") {
					t.Errorf("expected first arg to contain 'go', got %s", cmd.Args[0])
				}
				// Rest should match our input args
				for i, arg := range tt.args {
					if cmd.Args[i+1] != arg {
						t.Errorf("arg[%d] = %v, want %v", i+1, cmd.Args[i+1], arg)
					}
				}
			} else {
				// When go is not in PATH, should use mise exec if mise is available
				if isMiseAvailable() {
					if !strings.Contains(cmd.Args[0], "mise") {
						t.Errorf("expected to use mise when go not available and mise is available, got %s", cmd.Args[0])
					}
					// Check for "exec", "go@<version>", "--", "go" in args
					expectedArgs := []string{"mise", "exec", fmt.Sprintf("go@%s", goVersion), "--", "go"}
					if len(cmd.Args) < len(expectedArgs) {
						t.Fatalf("expected at least %d args for mise exec, got %d", len(expectedArgs), len(cmd.Args))
					}
					// Verify the go version matches
					if !strings.Contains(cmd.Args[2], goVersion) {
						t.Errorf("expected go version %s in args, got %v", goVersion, cmd.Args)
					}
				} else {
					// If neither go nor mise is available, should still return go command
					// (will fail with clear error when executed)
					if !strings.Contains(cmd.Args[0], "go") {
						t.Errorf("expected to use go when neither go nor mise available, got %s", cmd.Args[0])
					}
				}
			}
		})
	}
}

func TestIsHexString(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{
			name: "valid short commit sha",
			str:  "abc1234",
			want: true,
		},
		{
			name: "valid full commit sha",
			str:  "abc1234567890abcdef1234567890abcdef12345",
			want: true,
		},
		{
			name: "uppercase hex",
			str:  "ABC1234",
			want: true,
		},
		{
			name: "mixed case hex",
			str:  "AbC1234",
			want: true,
		},
		{
			name: "non-hex characters",
			str:  "xyz1234",
			want: false,
		},
		{
			name: "branch name",
			str:  "feature-branch",
			want: false,
		},
		{
			name: "version number",
			str:  "main.5",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isHexString(tt.str)
			if got != tt.want {
				t.Errorf("isHexString(%q) = %v, want %v", tt.str, got, tt.want)
			}
		})
	}
}

func TestCreateGoBuildCommandIntegration(t *testing.T) {
	// This test verifies the command structure without executing it
	cmd := createGoBuildCommand("build", "-o", "/tmp/test", ".")

	if cmd == nil {
		t.Fatal("createGoBuildCommand returned nil")
	}

	// The command should be executable
	_, err := exec.LookPath(cmd.Path)
	if err != nil {
		// If we can't find the exact path, check if the command name is valid
		cmdName := cmd.Args[0]
		if cmdName != "go" && cmdName != "mise" {
			t.Errorf("unexpected command name: %s", cmdName)
		}
	}
}

func TestGetBuildPath(t *testing.T) {
	tests := []struct {
		name        string
		projectDir  string
		want        string
		description string
	}{
		{
			name:        "project with cmd subdirectory",
			projectDir:  "../conf",
			want:        "./cmd",
			description: "conf has Go files in cmd subdirectory",
		},
		{
			name:        "project without cmd subdirectory",
			projectDir:  "../want",
			want:        ".",
			description: "want has Go files in root directory",
		},
		{
			name:        "another project with cmd",
			projectDir:  "../dissect",
			want:        "./cmd",
			description: "dissect has Go files in cmd subdirectory",
		},
		{
			name:        "project without cmd - prrun",
			projectDir:  "../prrun",
			want:        ".",
			description: "prrun has Go files in root directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getBuildPath(tt.projectDir)
			if got != tt.want {
				t.Errorf("getBuildPath(%q) = %v, want %v (%s)", tt.projectDir, got, tt.want, tt.description)
			}
		})
	}
}

func TestToolRegistry(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		wantExists   bool
		wantAutomatic bool
	}{
		{
			name:         "mise tool exists",
			toolName:     "mise",
			wantExists:   true,
			wantAutomatic: true,
		},
		{
			name:         "uv tool exists",
			toolName:     "uv",
			wantExists:   true,
			wantAutomatic: false,
		},
		{
			name:         "jc tool exists and is automatic",
			toolName:     "jc",
			wantExists:   true,
			wantAutomatic: true,
		},
		{
			name:         "markitdown tool exists",
			toolName:     "markitdown",
			wantExists:   true,
			wantAutomatic: true,
		},
		{
			name:       "nonexistent tool",
			toolName:   "nonexistent-tool-xyz",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, exists := ToolRegistry[tt.toolName]
			if exists != tt.wantExists {
				t.Errorf("ToolRegistry[%q] exists = %v, want %v", tt.toolName, exists, tt.wantExists)
			}
			if exists && tool.InstallStep.Automatic != tt.wantAutomatic {
				t.Errorf("ToolRegistry[%q].InstallStep.Automatic = %v, want %v", tt.toolName, tool.InstallStep.Automatic, tt.wantAutomatic)
			}
		})
	}
}

func TestBuildToolInstallationPlan(t *testing.T) {
	tests := []struct {
		name             string
		toolName         string
		wantStepsMin     int
		wantStepsMax     int
		expectMiseSteps  bool
	}{
		{
			name:            "jc installation (requires mise)",
			toolName:        "jc",
			wantStepsMin:    1,
			wantStepsMax:    4, // Could include mise installation steps
			expectMiseSteps: true,
		},
		{
			name:            "markitdown installation (requires mise)",
			toolName:        "markitdown",
			wantStepsMin:    1,
			wantStepsMax:    4,
			expectMiseSteps: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if tool is already available (can't test installation plan)
			if isToolAvailable(tt.toolName) {
				t.Skipf("Tool %s is already available, skipping installation plan test", tt.toolName)
			}

			steps := buildToolInstallationPlan(tt.toolName)
			
			if len(steps) < tt.wantStepsMin || len(steps) > tt.wantStepsMax {
				t.Errorf("buildToolInstallationPlan(%q) returned %d steps, want between %d and %d",
					tt.toolName, len(steps), tt.wantStepsMin, tt.wantStepsMax)
			}

			// Check that steps are properly structured
			for i, step := range steps {
				if step.Description == "" {
					t.Errorf("Step %d has empty Description", i)
				}
				if step.Command == "" {
					t.Errorf("Step %d has empty Command", i)
				}
			}
		})
	}
}

func TestBuildMiseInstallationSteps(t *testing.T) {
	steps := buildMiseInstallationSteps()

	// Should have at least 1 step (mise installation)
	// May have more if shell activation is needed
	if len(steps) < 1 {
		t.Errorf("buildMiseInstallationSteps() returned %d steps, want at least 1", len(steps))
	}

	// First step should be mise installation
	if steps[0].Type != "install" {
		t.Errorf("First step type = %v, want 'install'", steps[0].Type)
	}

	// Check that all steps have proper structure
	for i, step := range steps {
		if step.Description == "" {
			t.Errorf("Step %d has empty Description", i)
		}
		if step.Command == "" {
			t.Errorf("Step %d has empty Command", i)
		}
		if step.Type == "" {
			t.Errorf("Step %d has empty Type", i)
		}
	}
}

func TestBuildShellConfigStep(t *testing.T) {
	step := buildShellConfigStep()

	if step.Type != "configure" {
		t.Errorf("buildShellConfigStep() Type = %v, want 'configure'", step.Type)
	}

	if step.Automatic != true {
		t.Errorf("buildShellConfigStep() Automatic = %v, want true", step.Automatic)
	}

	if step.Description == "" {
		t.Error("buildShellConfigStep() has empty Description")
	}

	if step.Command == "" {
		t.Error("buildShellConfigStep() has empty Command")
	}

	// Command should contain "mise activate"
	if !strings.Contains(step.Command, "mise activate") {
		t.Errorf("buildShellConfigStep() Command should contain 'mise activate', got %q", step.Command)
	}
}

func TestBuildManualActivationStep(t *testing.T) {
	step := buildManualActivationStep()

	if step.Type != "configure" {
		t.Errorf("buildManualActivationStep() Type = %v, want 'configure'", step.Type)
	}

	if step.Automatic != false {
		t.Errorf("buildManualActivationStep() Automatic = %v, want false", step.Automatic)
	}

	if step.Description == "" {
		t.Error("buildManualActivationStep() has empty Description")
	}

	if step.Command == "" {
		t.Error("buildManualActivationStep() has empty Command")
	}

	// Command should contain "mise activate"
	if !strings.Contains(step.Command, "mise activate") {
		t.Errorf("buildManualActivationStep() Command should contain 'mise activate', got %q", step.Command)
	}
}
