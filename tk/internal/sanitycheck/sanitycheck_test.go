package sanitycheck

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSanityCheck_EmptyDatabase(t *testing.T) {
	// Create an in-memory database
	db, err := database.OpenDB(":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = db.InitDB()
	require.NoError(t, err)

	// Set database version
	err = db.SetDBVersion(7)
	require.NoError(t, err)

	cfg := &config.Config{}

	// Run sanity check on empty database - should find no differences
	hasDiff := RunSanityCheck(db, cfg)
	assert.False(t, hasDiff, "Empty database should have no differences")
}

func TestRunSanityCheck_MatchingState(t *testing.T) {
	// Create an in-memory database
	db, err := database.OpenDB(":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = db.InitDB()
	require.NoError(t, err)

	// Add columns from migrations (needed for projection code to work)
	// is_synthetic column (v4->v5)
	_, err = db.Exec(`ALTER TABLE projects ADD COLUMN is_synthetic INTEGER DEFAULT 0`)
	require.NoError(t, err)
	// item_kind column (v6->v7)
	_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN item_kind TEXT NOT NULL DEFAULT 'task'`)
	require.NoError(t, err)

	// Set database version
	err = db.SetDBVersion(7)
	require.NoError(t, err)

	cfg := &config.Config{}

	// Create a project and task through events (which updates both events and projections)
	projectUID := types.ProjectUID("test-project-uid")
	taskUID := types.TaskUID("test-task-uid")

	// Insert project.created event
	projectEvent := types.Event{
		ID:        "event-1",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload: mustMarshal(types.ProjectCreatedPayload{
			ProjectUID:  projectUID,
			Type:        types.ProjectTypeLocal,
			Name:        "Test Project",
			Description: "Test Description",
			CreatedBy:   "test",
		}),
	}
	err = db.InsertEvent(projectEvent)
	require.NoError(t, err)

	// Project the event into the database
	err = db.ProjectEvent(projectEvent)
	require.NoError(t, err)

	// Insert task.created event
	taskEvent := types.Event{
		ID:        "event-2",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload: mustMarshal(types.TaskCreatedPayload{
			TaskUID:     string(taskUID),
			ProjectUID:  projectUID.String(),
			Title:       "Test Task",
			CreatedNode: "test-node",
			CreatedBy:   "test",
		}),
	}
	err = db.InsertEvent(taskEvent)
	require.NoError(t, err)

	// Project the event into the database
	err = db.ProjectEvent(taskEvent)
	require.NoError(t, err)

	// Run sanity check - should find no differences since we properly projected
	hasDiff := RunSanityCheck(db, cfg)
	assert.False(t, hasDiff, "Properly projected events should have no differences")
}

func TestRunSanityCheck_MismatchDetection(t *testing.T) {
	// Create an in-memory database
	db, err := database.OpenDB(":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = db.InitDB()
	require.NoError(t, err)

	// Add columns from migrations (needed for projection code to work)
	// is_synthetic column (v4->v5)
	_, err = db.Exec(`ALTER TABLE projects ADD COLUMN is_synthetic INTEGER DEFAULT 0`)
	require.NoError(t, err)
	// item_kind column (v6->v7)
	_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN item_kind TEXT NOT NULL DEFAULT 'task'`)
	require.NoError(t, err)

	// Set database version
	err = db.SetDBVersion(7)
	require.NoError(t, err)

	cfg := &config.Config{}

	// Create a project
	projectUID := types.ProjectUID("test-project-uid")
	projectEvent := types.Event{
		ID:        "event-1",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload: mustMarshal(types.ProjectCreatedPayload{
			ProjectUID:  projectUID,
			Type:        types.ProjectTypeLocal,
			Name:        "Test Project",
			Description: "Test Description",
			CreatedBy:   "test",
		}),
	}
	err = db.InsertEvent(projectEvent)
	require.NoError(t, err)
	err = db.ProjectEvent(projectEvent)
	require.NoError(t, err)

	// Insert task.created event but DON'T project it (simulating a bug)
	taskUID := types.TaskUID("test-task-uid")
	taskEvent := types.Event{
		ID:        "event-2",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload: mustMarshal(types.TaskCreatedPayload{
			TaskUID:     string(taskUID),
			ProjectUID:  projectUID.String(),
			Title:       "Test Task",
			CreatedNode: "test-node",
			CreatedBy:   "test",
		}),
	}
	err = db.InsertEvent(taskEvent)
	require.NoError(t, err)
	// DON'T call db.ProjectEvent(taskEvent) - this creates the mismatch

	// Create a temporary directory for the diff file
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Run sanity check - should detect the missing task in database
	hasDiff := RunSanityCheck(db, cfg)
	assert.True(t, hasDiff, "Should detect missing task in database")

	// Verify diff file was created
	diffPath := filepath.Join(tempDir, ".tk", "state-diff.json")
	assert.FileExists(t, diffPath)

	// Read and verify diff content
	data, err := os.ReadFile(diffPath)
	require.NoError(t, err)

	var comparison StateComparison
	err = json.Unmarshal(data, &comparison)
	require.NoError(t, err)

	assert.Equal(t, 2, comparison.EventCount, "Should have 2 events")
	assert.Equal(t, 1, comparison.ReducerTaskCount, "Reducer should have 1 task")
	assert.Equal(t, 0, comparison.DatabaseTaskCount, "Database should have 0 tasks")
	assert.Len(t, comparison.Differences, 1, "Should have 1 difference")
	assert.Equal(t, "missing_in_db", comparison.Differences[0].Type)
	assert.Equal(t, string(taskUID), comparison.Differences[0].TaskUID)
}

func TestCompareTasks_TitleMismatch(t *testing.T) {
	reducerTask := &types.Task{
		TaskUUID:  "task-1",
		Title:     "Original Title",
		CreatedAt: time.Now(),
	}

	dbTaskVal := dbTask{
		TaskUID:   "task-1",
		Title:     "Different Title",
		CreatedAt: reducerTask.CreatedAt,
	}

	reducerTasks := make(map[string]*types.Task)
	reducerTasks["task-1"] = reducerTask

	dbTasks := make(map[string]*dbTask)
	dbTasks["task-1"] = &dbTaskVal

	diffs := compareTasks(reducerTasks, dbTasks)

	require.Len(t, diffs, 1)
	assert.Equal(t, "field_mismatch", diffs[0].Type)
	assert.Equal(t, "title", diffs[0].Field)
	assert.Equal(t, "Original Title", diffs[0].ReducerVal)
	assert.Equal(t, "Different Title", diffs[0].DatabaseVal)
}

func mustMarshal(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
