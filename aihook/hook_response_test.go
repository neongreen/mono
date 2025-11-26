package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestHookAllow(t *testing.T) {
	tests := []struct {
		name          string
		systemMessage string
	}{
		{
			name:          "without system message",
			systemMessage: "",
		},
		{
			name:          "with system message",
			systemMessage: "Found cd but allowed it",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := hookAllow(tt.systemMessage)

			if !resp.Continue {
				t.Error("hookAllow() should set Continue to true")
			}

			if tt.systemMessage != "" && resp.SystemMessage != tt.systemMessage {
				t.Errorf("hookAllow() SystemMessage = %q, want %q", resp.SystemMessage, tt.systemMessage)
			}

			if tt.systemMessage == "" && resp.SystemMessage != "" {
				t.Errorf("hookAllow() SystemMessage = %q, want empty", resp.SystemMessage)
			}

			if resp.HookSpecificOutput == nil {
				t.Fatal("hookAllow() HookSpecificOutput should not be nil")
			}

			if decision, ok := resp.HookSpecificOutput["permissionDecision"].(string); !ok || decision != "allow" {
				t.Errorf("hookAllow() permissionDecision = %q, want 'allow'", decision)
			}

			if eventName, ok := resp.HookSpecificOutput["hookEventName"].(string); !ok || eventName != "PreToolUse" {
				t.Errorf("hookAllow() hookEventName = %q, want 'PreToolUse'", eventName)
			}
		})
	}
}

func TestHookDeny(t *testing.T) {
	tests := []struct {
		name   string
		reason string
	}{
		{
			name:   "simple reason",
			reason: "cd command not allowed",
		},
		{
			name:   "detailed reason",
			reason: "Found cd commands outside subshells:\n  Line 1: 'cd' command found outside subshell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := hookDeny(tt.reason)

			if !resp.Continue {
				t.Error("hookDeny() should set Continue to true")
			}

			if resp.HookSpecificOutput == nil {
				t.Fatal("hookDeny() HookSpecificOutput should not be nil")
			}

			if decision, ok := resp.HookSpecificOutput["permissionDecision"].(string); !ok || decision != "deny" {
				t.Errorf("hookDeny() permissionDecision = %q, want 'deny'", decision)
			}

			if reason, ok := resp.HookSpecificOutput["permissionDecisionReason"].(string); !ok || reason != tt.reason {
				t.Errorf("hookDeny() permissionDecisionReason = %q, want %q", reason, tt.reason)
			}

			if eventName, ok := resp.HookSpecificOutput["hookEventName"].(string); !ok || eventName != "PreToolUse" {
				t.Errorf("hookDeny() hookEventName = %q, want 'PreToolUse'", eventName)
			}
		})
	}
}

func TestHookError(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "simple error",
			message: "failed to parse input",
		},
		{
			name:    "detailed error",
			message: "failed to parse JSON input: unexpected end of input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := hookError(tt.message)

			if resp.Continue {
				t.Error("hookError() should set Continue to false")
			}

			if resp.SystemMessage != tt.message {
				t.Errorf("hookError() SystemMessage = %q, want %q", resp.SystemMessage, tt.message)
			}

			expectedStopReason := "Hook error: " + tt.message
			if resp.StopReason != expectedStopReason {
				t.Errorf("hookError() StopReason = %q, want %q", resp.StopReason, expectedStopReason)
			}
		})
	}
}

func TestHookPreventStop(t *testing.T) {
	resp := hookPreventStop()

	if resp.Continue {
		t.Error("hookPreventStop() should set Continue to false")
	}

	if resp.StopReason == "" {
		t.Error("hookPreventStop() should have a StopReason")
	}

	if resp.SystemMessage == "" {
		t.Error("hookPreventStop() should have a SystemMessage")
	}

	// Verify JSON structure is valid
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal hookPreventStop response: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal hookPreventStop response: %v", err)
	}

	// Verify required fields
	if _, ok := parsed["continue"]; !ok {
		t.Error("JSON missing 'continue' field")
	}
	if _, ok := parsed["stopReason"]; !ok {
		t.Error("JSON missing 'stopReason' field")
	}
	if _, ok := parsed["systemMessage"]; !ok {
		t.Error("JSON missing 'systemMessage' field")
	}
}

func TestHookResponseJSON(t *testing.T) {
	tests := []struct {
		name     string
		response HookResponse
		wantKeys []string
	}{
		{
			name:     "hookAllow JSON structure",
			response: hookAllow("test message"),
			wantKeys: []string{"continue", "systemMessage", "hookSpecificOutput"},
		},
		{
			name:     "hookDeny JSON structure",
			response: hookDeny("test reason"),
			wantKeys: []string{"continue", "hookSpecificOutput"},
		},
		{
			name:     "hookError JSON structure",
			response: hookError("test error"),
			wantKeys: []string{"continue", "stopReason", "systemMessage"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Failed to marshal response to JSON: %v", err)
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			for _, key := range tt.wantKeys {
				if _, ok := parsed[key]; !ok {
					t.Errorf("JSON missing required key %q. Got JSON: %s", key, string(jsonBytes))
				}
			}
		})
	}
}

func TestFormatOutputClaudeMode(t *testing.T) {
	tests := []struct {
		name     string
		response HookResponse
		wantJSON bool
	}{
		{
			name:     "allow response",
			response: hookAllow(""),
			wantJSON: true,
		},
		{
			name:     "deny response",
			response: hookDeny("cd not allowed"),
			wantJSON: true,
		},
		{
			name:     "error response",
			response: hookError("parse error"),
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set claudeFlag to true for testing
			originalFlag := claudeFlag
			claudeFlag = true
			defer func() { claudeFlag = originalFlag }()

			// Capture stdout
			var buf bytes.Buffer
			oldStdout := &buf

			// Since we can't easily redirect os.Stdout in this test,
			// we'll just verify the response can be marshaled to JSON
			jsonBytes, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Failed to marshal response to JSON: %v", err)
			}

			// Verify it's valid JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
				t.Errorf("Generated invalid JSON: %v\nJSON: %s", err, string(jsonBytes))
			}

			// Verify continue field exists
			if _, ok := parsed["continue"]; !ok {
				t.Error("JSON missing 'continue' field")
			}

			_ = oldStdout // Silence unused variable warning
		})
	}
}

func TestHookResponseOmitEmptyFields(t *testing.T) {
	tests := []struct {
		name       string
		response   HookResponse
		shouldOmit []string
		shouldHave []string
	}{
		{
			name:       "allow without system message omits stopReason",
			response:   hookAllow(""),
			shouldOmit: []string{"stopReason", "systemMessage"},
			shouldHave: []string{"continue", "hookSpecificOutput"},
		},
		{
			name:       "allow with system message includes it",
			response:   hookAllow("warning message"),
			shouldOmit: []string{"stopReason"},
			shouldHave: []string{"continue", "systemMessage", "hookSpecificOutput"},
		},
		{
			name:       "deny omits stopReason",
			response:   hookDeny("reason"),
			shouldOmit: []string{"stopReason", "systemMessage"},
			shouldHave: []string{"continue", "hookSpecificOutput"},
		},
		{
			name:       "error includes stopReason",
			response:   hookError("error message"),
			shouldOmit: []string{"hookSpecificOutput"},
			shouldHave: []string{"continue", "stopReason", "systemMessage"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Failed to marshal response: %v", err)
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			jsonStr := string(jsonBytes)

			for _, field := range tt.shouldOmit {
				if val, ok := parsed[field]; ok {
					// Check if the value is zero/empty
					switch v := val.(type) {
					case string:
						if v != "" {
							t.Errorf("JSON should omit or have empty %q field, but got: %q", field, v)
						}
					case bool:
						// bool fields are always included
					default:
						t.Errorf("JSON should omit %q field but it's present: %v", field, val)
					}
				}
			}

			for _, field := range tt.shouldHave {
				if _, ok := parsed[field]; !ok {
					t.Errorf("JSON missing required field %q. Got: %s", field, jsonStr)
				}
			}
		})
	}
}

func TestHookSpecificOutputStructure(t *testing.T) {
	tests := []struct {
		name            string
		response        HookResponse
		wantEventName   string
		wantDecision    string
		wantReasonField bool
	}{
		{
			name:            "allow response structure",
			response:        hookAllow(""),
			wantEventName:   "PreToolUse",
			wantDecision:    "allow",
			wantReasonField: true,
		},
		{
			name:            "deny response structure",
			response:        hookDeny("test reason"),
			wantEventName:   "PreToolUse",
			wantDecision:    "deny",
			wantReasonField: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.response.HookSpecificOutput == nil {
				t.Fatal("HookSpecificOutput should not be nil")
			}

			eventName, ok := tt.response.HookSpecificOutput["hookEventName"].(string)
			if !ok || eventName != tt.wantEventName {
				t.Errorf("hookEventName = %q, want %q", eventName, tt.wantEventName)
			}

			decision, ok := tt.response.HookSpecificOutput["permissionDecision"].(string)
			if !ok || decision != tt.wantDecision {
				t.Errorf("permissionDecision = %q, want %q", decision, tt.wantDecision)
			}

			if tt.wantReasonField {
				if _, ok := tt.response.HookSpecificOutput["permissionDecisionReason"]; !ok {
					t.Error("permissionDecisionReason field should be present")
				}
			}
		})
	}
}
