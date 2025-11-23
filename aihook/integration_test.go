package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestShellCommandIntegration tests the full shell command flow including JSON input/output
func TestShellCommandIntegration(t *testing.T) {
	tests := []struct {
		name               string
		command            string
		blockOnCd          bool
		wantDecision       string
		wantContinue       bool
		wantReasonContains string
	}{
		{
			name:         "allow simple command",
			command:      "ls -la",
			blockOnCd:    false,
			wantDecision: "allow",
			wantContinue: true,
		},
		{
			name:         "allow cd in subshell",
			command:      "(cd /tmp && ls)",
			blockOnCd:    true,
			wantDecision: "allow",
			wantContinue: true,
		},
		{
			name:               "deny cd outside subshell with block flag",
			command:            "cd /tmp",
			blockOnCd:          true,
			wantDecision:       "deny",
			wantContinue:       true,
			wantReasonContains: "cd commands outside subshells",
		},
		{
			name:               "allow cd outside subshell without block flag (warning)",
			command:            "cd /tmp",
			blockOnCd:          false,
			wantDecision:       "allow",
			wantContinue:       true,
			wantReasonContains: "",
		},
		{
			name:               "deny multiple cd commands",
			command:            "cd /tmp && cd /home",
			blockOnCd:          true,
			wantDecision:       "deny",
			wantContinue:       true,
			wantReasonContains: "cd commands outside subshells",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create hook input JSON
			hookInput := map[string]interface{}{
				"session_id":      "test-session",
				"transcript_path": "/tmp/test-transcript",
				"cwd":             "/tmp",
				"permission_mode": "ask",
				"hook_event_name": "PreToolUse",
				"tool_name":       "Bash",
				"tool_input": map[string]interface{}{
					"command": tt.command,
				},
				"tool_use_id": "test-tool-use-id",
			}

			inputJSON, err := json.Marshal(hookInput)
			if err != nil {
				t.Fatalf("Failed to create input JSON: %v", err)
			}

			// Set up stdin and stdout
			oldStdin := os.Stdin
			oldStdout := os.Stdout
			defer func() {
				os.Stdin = oldStdin
				os.Stdout = oldStdout
			}()

			// Create pipe for stdin
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdin = r

			// Write input JSON to stdin
			go func() {
				w.Write(inputJSON)
				w.Close()
			}()

			// Capture stdout
			outR, outW, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create output pipe: %v", err)
			}
			os.Stdout = outW

			// Set flags
			claudeFlag = true
			blockOnCdFlag = tt.blockOnCd

			// Run the shell command
			err = runShell(shellCmd, []string{})

			// Close output writer and restore stdout
			outW.Close()
			os.Stdout = oldStdout

			// Read captured output
			var outBuf bytes.Buffer
			outBuf.ReadFrom(outR)
			output := outBuf.String()

			if err != nil {
				t.Fatalf("runShell() returned error: %v", err)
			}

			// Parse output JSON
			var response HookResponse
			if err := json.Unmarshal([]byte(output), &response); err != nil {
				t.Fatalf("Failed to parse output JSON: %v\nOutput: %s", err, output)
			}

			// Verify continue field
			if response.Continue != tt.wantContinue {
				t.Errorf("Continue = %v, want %v", response.Continue, tt.wantContinue)
			}

			// Verify hook-specific output
			if response.HookSpecificOutput == nil {
				if tt.wantDecision != "" {
					t.Fatal("HookSpecificOutput is nil, expected decision")
				}
				return
			}

			decision, ok := response.HookSpecificOutput["permissionDecision"].(string)
			if !ok {
				t.Fatal("permissionDecision not found or not a string")
			}

			if decision != tt.wantDecision {
				t.Errorf("permissionDecision = %q, want %q", decision, tt.wantDecision)
			}

			// Verify reason if specified
			if tt.wantReasonContains != "" {
				reason, ok := response.HookSpecificOutput["permissionDecisionReason"].(string)
				if !ok {
					t.Fatal("permissionDecisionReason not found or not a string")
				}
				if !strings.Contains(reason, tt.wantReasonContains) {
					t.Errorf("permissionDecisionReason should contain %q, got: %q", tt.wantReasonContains, reason)
				}
			}

			// Verify hookEventName
			eventName, ok := response.HookSpecificOutput["hookEventName"].(string)
			if !ok || eventName != "PreToolUse" {
				t.Errorf("hookEventName = %q, want 'PreToolUse'", eventName)
			}
		})
	}
}

// TestShellCommandInvalidJSON tests error handling for invalid JSON input
func TestShellCommandInvalidJSON(t *testing.T) {
	// Set up stdin with invalid JSON
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r

	go func() {
		w.Write([]byte("invalid json"))
		w.Close()
	}()

	// Capture stdout
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create output pipe: %v", err)
	}
	os.Stdout = outW

	claudeFlag = true

	// Run the shell command
	err = runShell(shellCmd, []string{})

	outW.Close()
	os.Stdout = oldStdout

	var outBuf bytes.Buffer
	outBuf.ReadFrom(outR)
	output := outBuf.String()

	if err != nil {
		t.Fatalf("runShell() returned error: %v", err)
	}

	// Parse output JSON
	var response HookResponse
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("Failed to parse output JSON: %v\nOutput: %s", err, output)
	}

	// Should have continue=false for errors
	if response.Continue {
		t.Error("Expected continue=false for invalid JSON input")
	}

	// Should have a stopReason
	if response.StopReason == "" {
		t.Error("Expected stopReason for invalid JSON input")
	}
}
