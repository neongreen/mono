package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
)

// setupTestDB creates a temporary test database with a project
func setupTestDB(t *testing.T) (*database.DB, string, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	if err := db.InitDB(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	if err := db.RunMigrationsIfNeeded(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Set TK_DB_PATH for the test
	oldPath := os.Getenv("TK_DB_PATH")
	os.Setenv("TK_DB_PATH", dbPath)

	// Create a test project
	projectUID := types.NewProjectUID()
	ts, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("Failed to get timestamp: %v", err)
	}

	projectPayload := types.ProjectCreatedPayload{
		ProjectUID:  projectUID.String(),
		Name:        "test",
		Description: "Test project",
		CreatedBy:   "test",
	}

	projectPayloadJSON, err := json.Marshal(projectPayload)
	if err != nil {
		t.Fatalf("Failed to marshal project payload: %v", err)
	}

	clk := &clock.RealClock{}
	projectEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        ts,
		CreatedAt: clk.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload:   projectPayloadJSON,
	}

	if err := db.InsertEvent(projectEvent); err != nil {
		t.Fatalf("Failed to insert project event: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Setenv("TK_DB_PATH", oldPath)
	}

	return db, projectUID.String(), cleanup
}

func TestGetActor(t *testing.T) {
	actor := GetActor()
	if actor == "" {
		t.Error("GetActor should not return empty string")
	}
}

func TestGetDisplayID(t *testing.T) {
	db, projectUID, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a task
	result, err := tasks.Create(db, tasks.CreateParams{
		ProjectUID: types.ProjectUID(projectUID),
		Title:      "Test task",
	}, "test", &clock.RealClock{})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	displayID := GetDisplayID(db, string(result.TaskUID))
	if displayID == "" {
		t.Error("GetDisplayID should not return empty string")
	}
	if displayID == string(result.TaskUID) {
		t.Error("GetDisplayID should return display ID, not UUID")
	}
}

func TestResolveProject(t *testing.T) {
	db, projectUID, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		projectRef  string
		shouldError bool
	}{
		{"existing project", "test", false},
		{"empty defaults to tk", "", true}, // Will fail because "tk" project doesn't exist
		{"non-existent project", "nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolvedUID, err := ResolveProject(db, tt.projectRef)
			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.projectRef == "test" && resolvedUID != projectUID {
					t.Errorf("Expected UID %s, got %s", projectUID, resolvedUID)
				}
			}
		})
	}
}

func TestSetTaskMetadata(t *testing.T) {
	db, projectUID, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a task
	result, err := tasks.Create(db, tasks.CreateParams{
		ProjectUID: types.ProjectUID(projectUID),
		Title:      "Test task",
	}, "test", &clock.RealClock{})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Set metadata
	err = setTaskMetadata(db, string(result.TaskUID), "foo", "bar", "test")
	if err != nil {
		t.Fatalf("Failed to set metadata: %v", err)
	}

	// Verify metadata was set
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	reducer, err := db.GetCachedReducerWithConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to get reducer: %v", err)
	}

	task, ok := reducer.GetTask(string(result.TaskUID))
	if !ok {
		t.Fatal("Task not found in reducer")
	}

	meta, ok := task.Metadata["foo"]
	if !ok {
		t.Error("Metadata key 'foo' not found")
	}
	if meta.Value != "bar" {
		t.Errorf("Expected metadata value 'bar', got %s", meta.Value)
	}
}

func TestMCPToolsRegistration(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	// Test that all tool functions can be called without panicking
	t.Run("CreateTaskTool", func(t *testing.T) {
		tool := CreateTaskTool(db)
		if tool == nil {
			t.Error("CreateTaskTool returned nil")
		}
	})

	t.Run("ListTasksTool", func(t *testing.T) {
		tool := ListTasksTool(db)
		if tool == nil {
			t.Error("ListTasksTool returned nil")
		}
	})

	t.Run("GetTaskTool", func(t *testing.T) {
		tool := GetTaskTool(db)
		if tool == nil {
			t.Error("GetTaskTool returned nil")
		}
	})

	t.Run("UpdateStatusTool", func(t *testing.T) {
		tool := UpdateStatusTool(db)
		if tool == nil {
			t.Error("UpdateStatusTool returned nil")
		}
	})

	t.Run("AddNoteTool", func(t *testing.T) {
		tool := AddNoteTool(db)
		if tool == nil {
			t.Error("AddNoteTool returned nil")
		}
	})

	t.Run("RelateTasksTool", func(t *testing.T) {
		tool := RelateTasksTool(db)
		if tool == nil {
			t.Error("RelateTasksTool returned nil")
		}
	})
}

func TestMCPResourcesRegistration(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("TaskResource", func(t *testing.T) {
		resource := TaskResource(db)
		if resource == nil {
			t.Error("TaskResource returned nil")
		}
	})

	t.Run("AllTasksResource", func(t *testing.T) {
		resource := AllTasksResource(db)
		if resource == nil {
			t.Error("AllTasksResource returned nil")
		}
	})
}
