package cmd

import (
	"os"
	"testing"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/testutil"
	"github.com/neongreen/mono/tk/internal/types"
)

// Wrappers for testutil functions (for backwards compatibility in this package)
var (
	openTempDB       = testutil.OpenTempDB
	seedProject      = testutil.SeedProject
	seedTask         = testutil.SeedTask
	seedTaskWithNode = testutil.SeedTaskWithNode
	mustJSON         = testutil.MustJSON
	marshalPayload   = testutil.MarshalPayload
)

// buildReducerFromDB builds a reducer from all events in the database (cmd-specific helper)
func buildReducerFromDB(t *testing.T, db *database.DB) *reducer.Reducer {
	events, err := db.GetEvents()
	if err != nil {
		t.Fatalf("failed to load events: %v", err)
	}
	r, err := reducer.BuildFromEvents(events)
	if err != nil {
		t.Fatalf("failed to build reducer: %v", err)
	}
	return r
}

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
			got := types.ExtractPrefix(tt.taskID)
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
