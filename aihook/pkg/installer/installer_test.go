package installer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallHook(t *testing.T) {
	tests := []struct {
		name         string
		initialJSON  string
		hookConfig   HookConfig
		expectedJSON map[string]any
	}{
		{
			name:        "install to empty settings",
			initialJSON: "",
			hookConfig: HookConfig{
				EventName: "Stop",
				Matcher:   "stop",
				Command:   "/usr/bin/aihook prevent-stop",
			},
			expectedJSON: map[string]any{
				"hooks": map[string]any{
					"Stop": []any{
						map[string]any{
							"matcher": "stop",
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "/usr/bin/aihook prevent-stop",
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "install to existing settings without hooks",
			initialJSON: `{"model": "sonnet"}`,
			hookConfig: HookConfig{
				EventName: "Stop",
				Matcher:   "stop",
				Command:   "/usr/bin/aihook prevent-stop",
			},
			expectedJSON: map[string]any{
				"model": "sonnet",
				"hooks": map[string]any{
					"Stop": []any{
						map[string]any{
							"matcher": "stop",
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "/usr/bin/aihook prevent-stop",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "install to existing settings with other hooks",
			initialJSON: `{
				"hooks": {
					"PreToolUse": [
						{
							"matcher": "Bash",
							"hooks": [{"type": "command", "command": "echo test"}]
						}
					]
				}
			}`,
			hookConfig: HookConfig{
				EventName: "Stop",
				Matcher:   "stop",
				Command:   "/usr/bin/aihook prevent-stop",
			},
			expectedJSON: map[string]any{
				"hooks": map[string]any{
					"PreToolUse": []any{
						map[string]any{
							"matcher": "Bash",
							"hooks": []any{
								map[string]any{"type": "command", "command": "echo test"},
							},
						},
					},
					"Stop": []any{
						map[string]any{
							"matcher": "stop",
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "/usr/bin/aihook prevent-stop",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "aihook-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Set HOME to temp dir
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", oldHome)

			// Create .claude directory and initial settings if provided
			claudeDir := filepath.Join(tmpDir, ".claude")
			if err := os.MkdirAll(claudeDir, 0755); err != nil {
				t.Fatalf("Failed to create .claude dir: %v", err)
			}

			settingsPath := filepath.Join(claudeDir, "settings.json")
			if tt.initialJSON != "" {
				if err := os.WriteFile(settingsPath, []byte(tt.initialJSON), 0644); err != nil {
					t.Fatalf("Failed to write initial settings: %v", err)
				}
			}

			// Install hook
			if err := InstallHook(tt.hookConfig); err != nil {
				t.Fatalf("InstallHook() error = %v", err)
			}

			// Read result
			data, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("Failed to read settings: %v", err)
			}

			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Failed to parse result JSON: %v", err)
			}

			// Compare hooks section
			resultJSON, _ := json.MarshalIndent(result, "", "  ")
			expectedJSON, _ := json.MarshalIndent(tt.expectedJSON, "", "  ")

			// Deep comparison of the hooks structure
			resultHooks, _ := result["hooks"].(map[string]any)
			expectedHooks, _ := tt.expectedJSON["hooks"].(map[string]any)

			if len(resultHooks) != len(expectedHooks) {
				t.Errorf("Different number of hook events.\nGot: %s\nWant: %s", resultJSON, expectedJSON)
				return
			}

			for eventName := range expectedHooks {
				if _, ok := resultHooks[eventName]; !ok {
					t.Errorf("Missing hook event %s.\nGot: %s", eventName, resultJSON)
				}
			}
		})
	}
}

func TestInstallHookDuplicatePrevention(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "aihook-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set HOME to temp dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	hookConfig := HookConfig{
		EventName: "Stop",
		Matcher:   "stop",
		Command:   "/usr/bin/aihook prevent-stop",
	}

	// Install hook twice
	if err := InstallHook(hookConfig); err != nil {
		t.Fatalf("First InstallHook() error = %v", err)
	}
	if err := InstallHook(hookConfig); err != nil {
		t.Fatalf("Second InstallHook() error = %v", err)
	}

	// Read result
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse result JSON: %v", err)
	}

	// Check that there's only one Stop hook
	hooks := result["hooks"].(map[string]any)
	stopHooks := hooks["Stop"].([]any)
	if len(stopHooks) != 1 {
		t.Errorf("Expected 1 Stop hook, got %d", len(stopHooks))
	}
}
