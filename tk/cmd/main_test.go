package cmd

import (
	"os"
	"testing"
)

func TestExtractPrefix(t *testing.T) {
	tests := []struct {
		name   string
		taskID string
		want   string
	}{
		{
			name:   "standard format",
			taskID: "proj-42-node123",
			want:   "proj",
		},
		{
			name:   "single part",
			taskID: "proj",
			want:   "proj",
		},
		{
			name:   "two parts",
			taskID: "proj-42",
			want:   "proj",
		},
		{
			name:   "empty",
			taskID: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPrefix(tt.taskID)
			if got != tt.want {
				t.Errorf("extractPrefix(%v) = %v, want %v", tt.taskID, got, tt.want)
			}
		})
	}
}

func TestGetCurrentUser(t *testing.T) {
	user, err := getCurrentUser()
	if err != nil {
		t.Skipf("getCurrentUser() error = %v (skipping test)", err)
	}
	if user == "" {
		t.Error("getCurrentUser() returned empty string")
	}

	// Should match environment USER or USERNAME
	envUser := os.Getenv("USER")
	if envUser == "" {
		envUser = os.Getenv("USERNAME")
	}

	if envUser != "" && user != envUser {
		t.Errorf("getCurrentUser() = %v, want %v", user, envUser)
	}
}

func TestColorizeStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{
			name:   "wip",
			status: "wip",
		},
		{
			name:   "done",
			status: "done",
		},
		{
			name:   "fixed",
			status: "fixed",
		},
		{
			name:   "other",
			status: "todo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorizeStatus(tt.status)
			// Just verify it returns something
			if result == "" {
				t.Errorf("colorizeStatus(%v) returned empty string", tt.status)
			}
		})
	}
}
