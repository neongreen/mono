package types

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestTaskTitleSetPayloadValidation tests that TaskTitleSetPayload validates properly
func TestTaskTitleSetPayloadValidation(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid payload",
			json:      `{"task_uid":"tsk_01J5QKF7F8M9N0P1Q2R3S4T5UV","title":"My task title"}`,
			wantError: false,
		},
		{
			name:      "invalid task UID - wrong prefix",
			json:      `{"task_uid":"prj_01J5QKF7F8M9N0P1Q2R3S4T5UV","title":"My task"}`,
			wantError: true,
			errorMsg:  "must start with tsk_",
		},
		{
			name:      "invalid task UID - not ULID",
			json:      `{"task_uid":"tsk_invalid","title":"My task"}`,
			wantError: true,
			errorMsg:  "invalid task UID ULID part",
		},
		{
			name:      "invalid title - empty",
			json:      `{"task_uid":"tsk_01J5QKF7F8M9N0P1Q2R3S4T5UV","title":""}`,
			wantError: true,
			errorMsg:  "title cannot be empty",
		},
		{
			name:      "invalid title - whitespace only",
			json:      `{"task_uid":"tsk_01J5QKF7F8M9N0P1Q2R3S4T5UV","title":"   "}`,
			wantError: true,
			errorMsg:  "title cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payload TaskTitleSetPayload
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

// TestProjectNameSetPayloadValidation tests that ProjectNameSetPayload validates properly
func TestProjectNameSetPayloadValidation(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid payload",
			json:      `{"project_uid":"prj_01J5QKF7F8M9N0P1Q2R3S4T5UV","name":"my-project"}`,
			wantError: false,
		},
		{
			name:      "invalid project UID",
			json:      `{"project_uid":"invalid","name":"my-project"}`,
			wantError: true,
			errorMsg:  "must start with prj_",
		},
		{
			name:      "invalid project name - uppercase",
			json:      `{"project_uid":"prj_01J5QKF7F8M9N0P1Q2R3S4T5UV","name":"MyProject"}`,
			wantError: true,
			errorMsg:  "must be lowercase",
		},
		{
			name:      "invalid project name - leading dash",
			json:      `{"project_uid":"prj_01J5QKF7F8M9N0P1Q2R3S4T5UV","name":"-project"}`,
			wantError: true,
			errorMsg:  "no leading/trailing dashes",
		},
		{
			name:      "invalid project name - empty",
			json:      `{"project_uid":"prj_01J5QKF7F8M9N0P1Q2R3S4T5UV","name":""}`,
			wantError: true,
			errorMsg:  "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payload ProjectNameSetPayload
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
