package database

import (
	"encoding/json"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectionDeterminism verifies that projections are deterministic regardless of
// the order events are ingested, as long as they're projected in Lamport timestamp order.
func TestProjectionDeterminism(t *testing.T) {
	// Create test events that simulate a realistic scenario
	events := createTestEventSet()

	// Create two databases with same events inserted in different orders
	db1 := createTempDB(t)
	defer db1.Close()

	db2 := createTempDB(t)
	defer db2.Close()

	// Insert events in original order into db1
	for _, e := range events {
		require.NoError(t, db1.InsertEvent(e))
	}

	// Shuffle events for db2
	shuffled := make([]types.Event, len(events))
	copy(shuffled, events)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Insert shuffled events into db2
	for _, e := range shuffled {
		require.NoError(t, db2.InsertEvent(e))
	}

	// Rebuild projections on both databases (projects in Lamport order)
	require.NoError(t, db1.RebuildProjections())
	require.NoError(t, db2.RebuildProjections())

	// Verify projection tables are identical
	assertProjectionsEqual(t, db1, db2)
}

// TestTaskRelocateDeterminism specifically tests that task.relocate events
// with different policies produce deterministic results.
func TestTaskRelocateDeterminism(t *testing.T) {
	db1 := createTempDB(t)
	defer db1.Close()

	db2 := createTempDB(t)
	defer db2.Close()

	// Create a project
	projectUID := types.NewProjectUID()
	createProjectInDB(t, db1, projectUID, "test-project")
	createProjectInDB(t, db2, projectUID, "test-project")

	// Create tasks in different projects
	task1 := types.NewTaskUID()
	task2 := types.NewTaskUID()
	project2UID := types.NewProjectUID()

	createProjectInDB(t, db1, project2UID, "other-project")
	createProjectInDB(t, db2, project2UID, "other-project")

	// Create tasks
	events := []types.Event{
		createTaskEvent(1, task1, projectUID, "Task 1"),
		setTaskNumberEvent(2, task1, projectUID, 1),
		createTaskEvent(3, task2, projectUID, "Task 2"),
		setTaskNumberEvent(4, task2, projectUID, 2),
		// Relocate task1 to project2 with force mode (deterministic)
		createTaskRelocateEvent(5, task1, projectUID, project2UID, "force", 10),
	}

	// Insert in original order
	for _, e := range events {
		require.NoError(t, db1.InsertEvent(e))
		require.NoError(t, db1.ProjectEvent(e))
	}

	// Insert in shuffled order
	shuffled := make([]types.Event, len(events))
	copy(shuffled, events)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	for _, e := range shuffled {
		require.NoError(t, db2.InsertEvent(e))
	}
	require.NoError(t, db2.RebuildProjections())

	// Verify task numbers are identical
	var num1, num2 int64
	require.NoError(t, db1.Db.QueryRow("SELECT number FROM task_numbers WHERE task_uid = ?", task1).Scan(&num1))
	require.NoError(t, db2.Db.QueryRow("SELECT number FROM task_numbers WHERE task_uid = ?", task1).Scan(&num2))
	assert.Equal(t, num1, num2, "task numbers should be identical after relocate")
	assert.Equal(t, int64(10), num1, "task should have number 10 from force mode")
}

// TestRebuildIdempotence verifies that rebuilding projections multiple times
// produces the same result.
func TestRebuildIdempotence(t *testing.T) {
	db := createTempDB(t)
	defer db.Close()

	events := createTestEventSet()
	for _, e := range events {
		require.NoError(t, db.InsertEvent(e))
	}

	// Rebuild three times
	require.NoError(t, db.RebuildProjections())
	snapshot1 := captureProjectionSnapshot(t, db)

	require.NoError(t, db.RebuildProjections())
	snapshot2 := captureProjectionSnapshot(t, db)

	require.NoError(t, db.RebuildProjections())
	snapshot3 := captureProjectionSnapshot(t, db)

	// All snapshots should be identical
	assert.Equal(t, snapshot1, snapshot2, "second rebuild should match first")
	assert.Equal(t, snapshot1, snapshot3, "third rebuild should match first")
}

// Helper functions

func createTempDB(t *testing.T) *DB {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := OpenDB(dbPath)
	require.NoError(t, err)
	require.NoError(t, db.InitDB())
	require.NoError(t, db.SetDBVersion(4))
	return db
}

func createTestEventSet() []types.Event {
	projectUID := types.NewProjectUID()
	task1UID := types.NewTaskUID()
	task2UID := types.NewTaskUID()
	task3UID := types.NewTaskUID()

	return []types.Event{
		createProjectEvent(1, projectUID, "test"),
		addProjectAliasEvent(2, projectUID, "test", "node1"),
		createTaskEvent(3, task1UID, projectUID, "Task 1"),
		setTaskNumberEvent(4, task1UID, projectUID, 1),
		createTaskEvent(5, task2UID, projectUID, "Task 2"),
		setTaskNumberEvent(6, task2UID, projectUID, 2),
		createTaskEvent(7, task3UID, projectUID, "Task 3"),
		setTaskNumberEvent(8, task3UID, projectUID, 3),
	}
}

func createProjectEvent(ts int64, projectUID types.ProjectUID, name string) types.Event {
	payload := types.ProjectCreatedPayload{
		ProjectUID:  projectUID.String(),
		Type:        "local",
		Name:        name,
		Description: "",
		CreatedBy:   "test",
	}
	payloadJSON, _ := json.Marshal(payload)
	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload:   payloadJSON,
	}
}

func addProjectAliasEvent(ts int64, projectUID types.ProjectUID, alias, node string) types.Event {
	payload := types.ProjectAliasAddPayload{
		ProjectUID: projectUID.String(),
		Alias:      alias,
		Node:       node,
		AddedBy:    "test",
	}
	payloadJSON, _ := json.Marshal(payload)
	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindProjectAliasAdd),
		Payload:   payloadJSON,
	}
}

func createTaskEvent(ts int64, taskUID types.TaskUID, projectUID types.ProjectUID, title string) types.Event {
	payload := types.TaskCreatedPayload{
		TaskUID:        taskUID.String(),
		ProjectUID:     projectUID.String(),
		ProposedNumber: 1,
		CreatedNode:    "node1",
		Title:          title,
		CreatedBy:      "test",
	}
	payloadJSON, _ := json.Marshal(payload)
	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   payloadJSON,
	}
}

func setTaskNumberEvent(ts int64, taskUID types.TaskUID, projectUID types.ProjectUID, number int64) types.Event {
	payload := types.TaskNumberSetPayload{
		TaskUID:    taskUID.String(),
		ProjectUID: projectUID.String(),
		Number:     number,
		Reason:     "initial",
	}
	payloadJSON, _ := json.Marshal(payload)
	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindTaskNumberSet),
		Payload:   payloadJSON,
	}
}

func createTaskRelocateEvent(ts int64, taskUID types.TaskUID, fromProjectUID, toProjectUID types.ProjectUID, mode string, number int64) types.Event {
	payload := types.TaskRelocatePayload{
		TaskUID:        taskUID.String(),
		FromProjectUID: fromProjectUID.String(),
		ToProjectUID:   toProjectUID.String(),
		NumberPolicy: types.NumberPolicyPayload{
			Mode:   mode,
			Number: number,
		},
	}
	payloadJSON, _ := json.Marshal(payload)
	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindTaskRelocate),
		Payload:   payloadJSON,
	}
}

func createProjectInDB(t *testing.T, db *DB, projectUID types.ProjectUID, name string) {
	event := createProjectEvent(1, projectUID, name)
	require.NoError(t, db.InsertEvent(event))
	require.NoError(t, db.ProjectEvent(event))
}

type ProjectionSnapshot struct {
	Projects       map[string]string // project_uid -> name
	ProjectAliases map[string]string // alias -> project_uid
	Tasks          map[string]string // task_uid -> title
	TaskNumbers    map[string]int64  // task_uid -> number
}

func captureProjectionSnapshot(t *testing.T, db *DB) ProjectionSnapshot {
	snapshot := ProjectionSnapshot{
		Projects:       make(map[string]string),
		ProjectAliases: make(map[string]string),
		Tasks:          make(map[string]string),
		TaskNumbers:    make(map[string]int64),
	}

	// Capture projects
	rows, err := db.Db.Query("SELECT project_uid, name FROM projects")
	require.NoError(t, err)
	for rows.Next() {
		var uid, name string
		require.NoError(t, rows.Scan(&uid, &name))
		snapshot.Projects[uid] = name
	}
	rows.Close()

	// Capture project aliases
	rows, err = db.Db.Query("SELECT alias, project_uid FROM project_aliases")
	require.NoError(t, err)
	for rows.Next() {
		var alias, uid string
		require.NoError(t, rows.Scan(&alias, &uid))
		snapshot.ProjectAliases[alias] = uid
	}
	rows.Close()

	// Capture tasks
	rows, err = db.Db.Query("SELECT task_uid, title FROM tasks")
	require.NoError(t, err)
	for rows.Next() {
		var uid, title string
		require.NoError(t, rows.Scan(&uid, &title))
		snapshot.Tasks[uid] = title
	}
	rows.Close()

	// Capture task numbers
	rows, err = db.Db.Query("SELECT task_uid, number FROM task_numbers")
	require.NoError(t, err)
	for rows.Next() {
		var uid string
		var number int64
		require.NoError(t, rows.Scan(&uid, &number))
		snapshot.TaskNumbers[uid] = number
	}
	rows.Close()

	return snapshot
}

func assertProjectionsEqual(t *testing.T, db1, db2 *DB) {
	snap1 := captureProjectionSnapshot(t, db1)
	snap2 := captureProjectionSnapshot(t, db2)

	assert.Equal(t, snap1.Projects, snap2.Projects, "projects should be identical")
	assert.Equal(t, snap1.ProjectAliases, snap2.ProjectAliases, "project aliases should be identical")
	assert.Equal(t, snap1.Tasks, snap2.Tasks, "tasks should be identical")
	assert.Equal(t, snap1.TaskNumbers, snap2.TaskNumbers, "task numbers should be identical")
}

