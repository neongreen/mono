package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/neongreen/mono/aihook/pkg/validator"
	"github.com/neongreen/mono/lib/version"
)

var rootCmd = &cobra.Command{
	Use:   "aihook",
	Short: "Claude Code hook validator",
	Long:  `aihook validates shell commands and code patterns for Claude Code hooks.`,
}

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Shell hook that validates shell commands",
	Long:  `Shell hook that parses shell syntax and forbids 'cd' invocations outside subshells.`,
	RunE:  runShell,
}

var claudeFlag bool
var blockOnCdFlag bool

func init() {
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(version.NewVersionCommand("aihook"))

	shellCmd.Flags().BoolVar(&claudeFlag, "claude", false, "Output in Claude Code hook format")
	shellCmd.Flags().BoolVar(&blockOnCdFlag, "block-on-cd", false, "Block when cd commands are found outside subshells")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runShell(cmd *cobra.Command, args []string) error {
	// Read and parse JSON input from stdin
	// Claude Code hooks pass a full hook event with tool_input as a nested field
	var hookInput struct {
		SessionID      string `json:"session_id"`
		TranscriptPath string `json:"transcript_path"`
		Cwd            string `json:"cwd"`
		PermissionMode string `json:"permission_mode"`
		HookEventName  string `json:"hook_event_name"`
		ToolName       string `json:"tool_name"`
		ToolInput      struct {
			Command     string `json:"command"`
			Timeout     int    `json:"timeout,omitempty"`
			Description string `json:"description,omitempty"`
		} `json:"tool_input"`
		ToolUseID string `json:"tool_use_id"`
	}

	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&hookInput); err != nil {
		return formatOutput(hookError(fmt.Sprintf("failed to parse JSON input: %v", err)))
	}

	// Validate the command string from tool_input
	v := validator.New()
	commandReader := strings.NewReader(hookInput.ToolInput.Command)
	violations, err := v.ValidateScript(commandReader)
	if err != nil {
		return formatOutput(hookError(err.Error()))
	}

	if len(violations) > 0 {
		msg := validator.FormatViolations(violations)
		if blockOnCdFlag {
			return formatOutput(hookDeny(msg))
		}
		return formatOutput(hookAllow(msg))
	}

	return formatOutput(hookAllow(""))
}

// HookResponse represents the output structure for Claude Code hooks
type HookResponse struct {
	Continue         bool                   `json:"continue"`
	StopReason       string                 `json:"stopReason,omitempty"`
	SuppressOutput   bool                   `json:"suppressOutput,omitempty"`
	SystemMessage    string                 `json:"systemMessage,omitempty"`
	HookSpecificOutput map[string]interface{} `json:"hookSpecificOutput,omitempty"`
}

// hookAllow creates a response that allows the tool call
func hookAllow(systemMessage string) HookResponse {
	resp := HookResponse{
		Continue: true,
		HookSpecificOutput: map[string]interface{}{
			"hookEventName":             "PreToolUse",
			"permissionDecision":        "allow",
			"permissionDecisionReason": "Command validated successfully",
		},
	}
	if systemMessage != "" {
		resp.SystemMessage = systemMessage
	}
	return resp
}

// hookDeny creates a response that denies the tool call
func hookDeny(reason string) HookResponse {
	return HookResponse{
		Continue: true,
		HookSpecificOutput: map[string]interface{}{
			"hookEventName":             "PreToolUse",
			"permissionDecision":        "deny",
			"permissionDecisionReason": reason,
		},
	}
}

// hookError creates a response that indicates an error occurred
func hookError(message string) HookResponse {
	return HookResponse{
		Continue:      false,
		StopReason:    "Hook error: " + message,
		SystemMessage: message,
	}
}

// formatOutput formats the output according to the --claude flag
func formatOutput(response HookResponse) error {
	if claudeFlag {
		// Claude Code hook format - always exit with code 0
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(response); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
		return nil
	}

	// Regular output for non-Claude usage
	if response.Continue && response.HookSpecificOutput != nil {
		if decision, ok := response.HookSpecificOutput["permissionDecision"].(string); ok {
			if decision == "allow" {
				if response.SystemMessage != "" {
					fmt.Println(response.SystemMessage)
				} else {
					fmt.Println("OK")
				}
				return nil
			} else if decision == "deny" {
				if reason, ok := response.HookSpecificOutput["permissionDecisionReason"].(string); ok {
					fmt.Fprintln(os.Stderr, reason)
					os.Exit(1)
				}
			}
		}
	}

	if !response.Continue {
		fmt.Fprintln(os.Stderr, response.StopReason)
		os.Exit(1)
	}

	return nil
}
