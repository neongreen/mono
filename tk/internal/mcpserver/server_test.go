package mcpserver

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

func TestMCPServerBasic(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := db.RunMigrationsIfNeeded(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Set TK_DB_PATH for the test
	oldPath := os.Getenv("TK_DB_PATH")
	os.Setenv("TK_DB_PATH", dbPath)
	defer os.Setenv("TK_DB_PATH", oldPath)

	// Create a project
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

	// Create a task
	result, err := tasks.Create(db, tasks.CreateParams{
		ProjectUID: projectUID,
		Title:      "Test task",
	}, "test", clk)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Create MCP server
	server := NewServer(db)
	if server == nil {
		t.Fatal("Failed to create MCP server")
	}

	// Verify that the server has registered the expected tools
	// Note: We can't easily test the actual MCP protocol without a client,
	// but we can verify the server was created successfully
	if server.db == nil {
		t.Fatal("Server database is nil")
	}

	if server.server == nil {
		t.Fatal("Server MCP server is nil")
	}

	t.Logf("Created task: %s", result.DisplayID)
	t.Log("MCP server created successfully")
}
