package types

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestProjectDeletePayloadValidation tests that ProjectDeletePayload validates properly
func TestProjectDeletePayloadValidation(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid project UID",
			json:      `{"project_uid":"prj_01J5QKF7F8M9N0P1Q2R3S4T5UV"}`,
			wantError: false,
		},
		{
			name:      "invalid - missing prj_ prefix",
			json:      `{"project_uid":"01J5QKF7F8M9N0P1Q2R3S4T5UV"}`,
			wantError: true,
			errorMsg:  "must start with prj_",
		},
		{
			name:      "invalid - not a ULID",
			json:      `{"project_uid":"prj_invalid"}`,
			wantError: true,
			errorMsg:  "invalid project UID ULID part",
		},
		{
			name:      "invalid - empty string",
			json:      `{"project_uid":""}`,
			wantError: true,
			errorMsg:  "must start with prj_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payload ProjectDeletePayload
			err := json.Unmarshal([]byte(tt.json), &payload)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestProjectDeletePayloadValidate tests the Validate() method directly
func TestProjectDeletePayloadValidate(t *testing.T) {
	tests := []struct {
		name      string
		payload   ProjectDeletePayload
		wantError bool
	}{
		{
			name:      "valid",
			payload:   ProjectDeletePayload{ProjectUID: "prj_01J5QKF7F8M9N0P1Q2R3S4T5UV"},
			wantError: false,
		},
		{
			name:      "invalid prefix",
			payload:   ProjectDeletePayload{ProjectUID: "invalid_01J5QKF7F8M9N0P1Q2R3S4T5UV"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}
