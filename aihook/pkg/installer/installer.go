package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// HookConfig represents a hook configuration to install
type HookConfig struct {
	EventName string // e.g., "Stop", "PreToolUse"
	Matcher   string // e.g., "stop", "Bash"
	Command   string // the command to run
}

// ClaudeSettings represents the structure of ~/.claude/settings.json
type ClaudeSettings struct {
	Hooks map[string][]HookMatcher `json:"hooks,omitempty"`
	// Include other fields to preserve them during read/write
	OtherFields map[string]any `json:"-"`
}

// HookMatcher represents a hook matcher configuration
type HookMatcher struct {
	Matcher string        `json:"matcher,omitempty"`
	Hooks   []HookCommand `json:"hooks"`
}

// HookCommand represents a single hook command
type HookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// getClaudeSettingsPath returns the path to ~/.claude/settings.json
func getClaudeSettingsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(homeDir, ".claude", "settings.json"), nil
}

// InstallHook installs a hook to the global Claude settings file
func InstallHook(config HookConfig) error {
	settingsPath, err := getClaudeSettingsPath()
	if err != nil {
		return err
	}

	// Ensure the .claude directory exists
	claudeDir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Read existing settings or create new
	settings := make(map[string]any)
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse existing settings: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	// Get or create hooks map
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		hooks = make(map[string]any)
		settings["hooks"] = hooks
	}

	// Get or create the event array
	eventHooks, ok := hooks[config.EventName].([]any)
	if !ok {
		eventHooks = []any{}
	}

	// Create the new hook entry
	newHook := map[string]any{
		"matcher": config.Matcher,
		"hooks": []map[string]any{
			{
				"type":    "command",
				"command": config.Command,
			},
		},
	}

	// Check if a similar hook already exists (same matcher and command)
	hookExists := false
	for _, h := range eventHooks {
		hookMap, ok := h.(map[string]any)
		if !ok {
			continue
		}
		if hookMap["matcher"] == config.Matcher {
			// Check if the command already exists
			hookCommands, ok := hookMap["hooks"].([]any)
			if ok {
				for _, hc := range hookCommands {
					hcMap, ok := hc.(map[string]any)
					if ok && hcMap["command"] == config.Command {
						hookExists = true
						break
					}
				}
			}
		}
		if hookExists {
			break
		}
	}

	if !hookExists {
		eventHooks = append(eventHooks, newHook)
		hooks[config.EventName] = eventHooks
	}

	// Write settings back
	output, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Add trailing newline
	output = append(output, '\n')

	if err := os.WriteFile(settingsPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}
