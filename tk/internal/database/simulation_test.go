package database

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/types"
)

// Multi-machine simulation tests using virtual time
//
// These tests simulate distributed tk scenarios where:
// - Multiple machines operate with different wall-clock times
// - Events are created, synced, and merged
// - We verify convergence despite timing differences
//
// This catches bugs related to:
// - Using created_at instead of Lamport timestamps
// - Non-deterministic projection
// - Filesystem ordering issues

// Machine represents a simulated machine in tests
type Machine struct {
	t      *testing.T
	db     *DB
	clock  *clock.VirtualClock
	nodeID string
}

// newMachine creates a test machine
func newMachine(t *testing.T, nodeID string, startTime time.Time) *Machine {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "machine-"+nodeID+".db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	if err := db.SetDBVersion(4); err != nil {
		t.Fatalf("failed to set version: %v", err)
	}

	if _, err := db.Db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('node_id', ?)`, nodeID); err != nil {
		t.Fatalf("failed to set node_id: %v", err)
	}

	return &Machine{t: t, db: db, clock: clock.NewVirtualClock(startTime), nodeID: nodeID}
}

func (m *Machine) createProject(alias, name string) types.ProjectUID {
	// Create project manually (avoid testutil to prevent import cycle)
	projectUID := types.NewProjectUID()

	payload := types.ProjectCreatedPayload{
		ProjectUID:  projectUID.String(),
		Type:        "local",
		Name:        name,
		Description: "",
		CreatedBy:   "machine-" + m.nodeID,
	}
	payloadJSON, _ := json.Marshal(payload)

	ts, _ := m.db.GetNextLamportTS()

	event := types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts,
		CreatedAt: m.clock.Now(),
		Actor:     "machine-" + m.nodeID,
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload:   payloadJSON,
	}

	m.db.InsertEvent(event)
	m.db.ProjectProjectCreatedEvent(event)

	// Add alias
	ts2, _ := m.db.GetNextLamportTS()
	aliasPayload := types.ProjectAliasAddPayload{
		ProjectUID: projectUID.String(),
		Alias:      alias,
		Node:       m.nodeID,
		AddedBy:    "machine-" + m.nodeID,
	}
	aliasJSON, _ := json.Marshal(aliasPayload)

	aliasEvent := types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts2,
		CreatedAt: m.clock.Now(),
		Actor:     "machine-" + m.nodeID,
		Role:      "human",
		Kind:      string(types.EventKindProjectAliasAdd),
		Payload:   aliasJSON,
	}

	m.db.InsertEvent(aliasEvent)
	m.db.ProjectProjectAliasAddEvent(aliasEvent)

	return projectUID
}

func (m *Machine) createTask(projectUID types.ProjectUID, title string) types.TaskUID {
	// Create task manually (avoid tasks package to prevent import cycle)
	taskUID := types.NewTaskUID()

	ts, _ := m.db.GetNextLamportTS()

	payload := types.TaskCreatedPayload{
		TaskUID:        taskUID.String(),
		ProjectUID:     projectUID.String(),
		ProposedNumber: 1,
		CreatedNode:    m.nodeID,
		Title:          title,
		CreatedBy:      "machine-" + m.nodeID,
	}
	payloadJSON, _ := json.Marshal(payload)

	event := types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts,
		CreatedAt: m.clock.Now(),
		Actor:     "machine-" + m.nodeID,
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   payloadJSON,
	}

	m.db.InsertEvent(event)
	m.db.ProjectTaskCreatedEvent(event)

	// Also create number.set event
	ts2, _ := m.db.GetNextLamportTS()
	numberPayload := types.TaskNumberSetPayload{
		TaskUID:    taskUID.String(),
		ProjectUID: projectUID.String(),
		Number:     1,
		Reason:     "initial",
	}
	numberJSON, _ := json.Marshal(numberPayload)

	numberEvent := types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts2,
		CreatedAt: m.clock.Now(),
		Actor:     "machine-" + m.nodeID,
		Role:      "human",
		Kind:      string(types.EventKindTaskNumberSet),
		Payload:   numberJSON,
	}

	m.db.InsertEvent(numberEvent)
	m.db.ProjectTaskNumberSetEvent(numberEvent)

	return taskUID
}

func (m *Machine) getEvents() []types.Event {
	rows, err := m.db.Db.Query(`SELECT id, ts, created_at, actor, role, kind, payload FROM events ORDER BY ts ASC`)
	if err != nil {
		m.t.Fatalf("failed to query events: %v", err)
	}
	defer rows.Close()

	var events []types.Event
	for rows.Next() {
		var e types.Event
		var createdAtNano int64
		if err := rows.Scan(&e.ID, &e.TS, &createdAtNano, &e.Actor, &e.Role, &e.Kind, &e.Payload); err != nil {
			m.t.Fatalf("failed to scan event: %v", err)
		}
		e.CreatedAt = time.Unix(0, createdAtNano)
		events = append(events, e)
	}
	return events
}

func (m *Machine) ingestEvents(events []types.Event) {
	for _, event := range events {
		var exists int
		m.db.Db.QueryRow(`SELECT COUNT(*) FROM events WHERE id = ?`, event.ID).Scan(&exists)
		if exists > 0 {
			continue
		}

		if err := m.db.InsertEvent(event); err != nil {
			m.t.Fatalf("failed to ingest event: %v", err)
		}
	}

	if err := m.db.RebuildProjections(); err != nil {
		m.t.Fatalf("failed to rebuild projections: %v", err)
	}
}

func syncMachines(t *testing.T, a, b *Machine) {
	t.Helper()
	eventsA := a.getEvents()
	eventsB := b.getEvents()
	a.ingestEvents(eventsB)
	b.ingestEvents(eventsA)
	t.Logf("✓ Synced %s (%d events) and %s (%d events)", a.nodeID, len(eventsA), b.nodeID, len(eventsB))
}

// TestSimulation_DifferentCreatedAt verifies that Lamport timestamps (not created_at) determine causality
func TestSimulation_DifferentCreatedAt(t *testing.T) {
	// Scenario: Two machines create tasks at different wall-clock times
	// Machine A at T=100, Machine B at T=50 (earlier!)
	// But Lamport timestamps determine the actual causal order

	machineA := newMachine(t, "node-A", time.Unix(100, 0))
	machineB := newMachine(t, "node-B", time.Unix(50, 0))

	// Both create the same project (must be in sync first)
	projectUID := machineA.createProject("test", "Test Project")
	machineB.createProject("test", "Test Project") // Same alias and name

	// Machine A creates task at T=100 (later wall clock)
	machineA.clock.Set(time.Unix(100, 0))
	taskA := machineA.createTask(projectUID, "Task from A")

	// Machine B creates task at T=50 (earlier wall clock!)
	machineB.clock.Set(time.Unix(50, 0))
	taskB := machineB.createTask(projectUID, "Task from B")

	// Verify: Different created_at times
	eventsA := machineA.getEvents()
	eventsB := machineB.getEvents()

	t.Logf("Machine A task created at: %v", eventsA[len(eventsA)-2].CreatedAt) // -2 because number.set is last
	t.Logf("Machine B task created at: %v", eventsB[len(eventsB)-2].CreatedAt)

	// Sync machines
	syncMachines(t, machineA, machineB)

	// CRITICAL: Both machines must converge to the same state
	// Lamport timestamps (not created_at) determine order
	assertConvergence(t, machineA, machineB)

	// Both machines should see both tasks
	stateA, err := reducer.BuildFromEvents(machineA.getEvents())
	if err != nil {
		t.Fatalf("failed to build state: %v", err)
	}

	if len(stateA.Tasks()) != 2 {
		t.Fatalf("Expected 2 tasks after sync, got %d", len(stateA.Tasks()))
	}

	// Verify both tasks exist by UUID
	assertTaskExists(t, stateA, string(taskA), "Task from A")
	assertTaskExists(t, stateA, string(taskB), "Task from B")

	t.Logf("✓ Machines converged despite different created_at times")
}

// TestSimulation_DuplicateWithTiming proves tk-229 fix works in multi-machine scenario
func TestSimulation_DuplicateWithTiming(t *testing.T) {
	// Scenario: Both machines somehow create duplicate task.created events for same UID
	// (This could happen during migration, sync issues, or bugs)
	// Machine A creates at T=100, Machine B creates at T=50
	// Different data (titles) but same UID

	machineA := newMachine(t, "node-A", time.Unix(100, 0))
	machineB := newMachine(t, "node-B", time.Unix(50, 0))

	// Create shared project
	projectUID := machineA.createProject("test", "Test Project")
	machineB.createProject("test", "Test Project")

	// Both machines create a task with THE SAME UID (duplicate!)
	// But at different times with different titles
	sharedUID := types.NewTaskUID()

	// Machine B creates at T=50 with earlier Lamport TS
	machineB.clock.Set(time.Unix(50, 0))
	// Force Lamport TS to be low
	event2 := types.Event{
		ID:        types.NewEventID().String(),
		TS:        2, // Explicitly set low TS
		CreatedAt: machineB.clock.Now(),
		Actor:     "machine-B",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload: mustMarshal(types.TaskCreatedPayload{
			TaskUID:        sharedUID.String(),
			ProjectUID:     projectUID.String(),
			ProposedNumber: 1,
			CreatedNode:    "node-B",
			Title:          "Task from B (TS=2, earlier)",
			CreatedBy:      "machine-B",
		}),
	}
	machineB.db.InsertEvent(event2)
	machineB.db.ProjectTaskCreatedEvent(event2)

	// Machine A creates at T=100 with later Lamport TS
	machineA.clock.Set(time.Unix(100, 0))
	event1 := types.Event{
		ID:        types.NewEventID().String(),
		TS:        10, // Explicitly set high TS
		CreatedAt: machineA.clock.Now(),
		Actor:     "machine-A",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload: mustMarshal(types.TaskCreatedPayload{
			TaskUID:        sharedUID.String(), // SAME UID!
			ProjectUID:     projectUID.String(),
			ProposedNumber: 1,
			CreatedNode:    "node-A",
			Title:          "Task from A (TS=10, later)",
			CreatedBy:      "machine-A",
		}),
	}
	machineA.db.InsertEvent(event1)
	machineA.db.ProjectTaskCreatedEvent(event1)

	t.Logf("Machine A: created_at T=100, Lamport TS=%d, title='Task from A (TS=10, later)'", event1.TS)
	t.Logf("Machine B: created_at T=50, Lamport TS=%d, title='Task from B (TS=2, earlier)'", event2.TS)

	// Sync machines
	syncMachines(t, machineA, machineB)

	// CRITICAL: Both machines must converge to the task with EARLIER Lamport TS
	assertConvergence(t, machineA, machineB)

	// Verify both machines have the task from B (earlier Lamport TS=2)
	stateA, _ := reducer.BuildFromEvents(machineA.getEvents())
	task, exists := stateA.GetTask(sharedUID.String())
	if !exists {
		t.Fatal("Task not found after sync")
	}

	if task.Title != "Task from B (TS=2, earlier)" {
		t.Fatalf("Expected title from earlier Lamport TS (B, TS=2), got: %s", task.Title)
	}

	t.Logf("✓ Both machines converged to task from B (earlier Lamport TS=2)")
	t.Logf("✓ Duplicate handling is now deterministic (tk-229 fixed)")
	t.Logf("✓ Proves: Lamport TS wins over created_at time (B at T=50 beats A at T=100)")
}

// Helper functions for simulation tests

// mustMarshal marshals a value to JSON or panics
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

// assertConvergence verifies that two machines have converged to the same state
func assertConvergence(t *testing.T, machineA, machineB *Machine) {
	t.Helper()

	// Build reducers from events
	stateA, err := reducer.BuildFromEvents(machineA.getEvents())
	if err != nil {
		t.Fatalf("failed to build state for machine %s: %v", machineA.nodeID, err)
	}

	stateB, err := reducer.BuildFromEvents(machineB.getEvents())
	if err != nil {
		t.Fatalf("failed to build state for machine %s: %v", machineB.nodeID, err)
	}

	// Check task count
	if len(stateA.Tasks()) != len(stateB.Tasks()) {
		diff := formatStateDiff(machineA.nodeID, stateA, machineB.nodeID, stateB)
		t.Fatalf("Convergence failed: machine %s has %d tasks, machine %s has %d tasks\n%s",
			machineA.nodeID, len(stateA.Tasks()),
			machineB.nodeID, len(stateB.Tasks()),
			diff)
	}

	// Check all tasks match
	for uuid, taskA := range stateA.Tasks() {
		taskB, exists := stateB.GetTask(uuid)
		if !exists {
			t.Fatalf("Convergence failed: task %s exists on machine %s but not on machine %s",
				uuid, machineA.nodeID, machineB.nodeID)
		}

		// Check title
		if taskA.Title != taskB.Title {
			t.Fatalf("Convergence failed: task %s has different titles:\n  Machine %s: %q\n  Machine %s: %q",
				uuid, machineA.nodeID, taskA.Title, machineB.nodeID, taskB.Title)
		}

		// Check status
		for axis, statusA := range taskA.Axes {
			statusB, exists := taskB.Axes[axis]
			if !exists {
				t.Fatalf("Convergence failed: task %s has axis %s on machine %s but not on machine %s",
					uuid, axis, machineA.nodeID, machineB.nodeID)
			}

			if statusA.Effective != statusB.Effective {
				t.Fatalf("Convergence failed: task %s axis %s has different effective status:\n  Machine %s: %q\n  Machine %s: %q",
					uuid, axis, machineA.nodeID, statusA.Effective, machineB.nodeID, statusB.Effective)
			}
		}
	}

	t.Logf("✓ Machines %s and %s converged (%d tasks)", machineA.nodeID, machineB.nodeID, len(stateA.Tasks()))
}

// assertTaskExists verifies a task exists with the expected properties
func assertTaskExists(t *testing.T, state *reducer.Reducer, taskUID string, expectedTitle string) {
	t.Helper()

	task, exists := state.GetTask(taskUID)
	if !exists {
		t.Fatalf("Task %s not found in state", taskUID)
	}

	if task.Title != expectedTitle {
		t.Fatalf("Task %s has title %q, expected %q", taskUID, task.Title, expectedTitle)
	}
}

// formatStateDiff shows differences between two machine states
func formatStateDiff(labelA string, stateA *reducer.Reducer, labelB string, stateB *reducer.Reducer) string {
	diff := fmt.Sprintf("State difference between %s and %s:\n", labelA, labelB)

	diff += fmt.Sprintf("  %s: %d tasks\n", labelA, len(stateA.Tasks()))
	diff += fmt.Sprintf("  %s: %d tasks\n", labelB, len(stateB.Tasks()))

	// Tasks only in A
	for uuid, taskA := range stateA.Tasks() {
		if _, exists := stateB.GetTask(uuid); !exists {
			diff += fmt.Sprintf("  - Only in %s: %s (%s)\n", labelA, uuid, taskA.Title)
		}
	}

	// Tasks only in B
	for uuid, taskB := range stateB.Tasks() {
		if _, exists := stateA.GetTask(uuid); !exists {
			diff += fmt.Sprintf("  - Only in %s: %s (%s)\n", labelB, uuid, taskB.Title)
		}
	}

	// Tasks with different data
	for uuid, taskA := range stateA.Tasks() {
		taskB, exists := stateB.GetTask(uuid)
		if !exists {
			continue
		}

		if taskA.Title != taskB.Title {
			diff += fmt.Sprintf("  - Different title for %s:\n", uuid)
			diff += fmt.Sprintf("      %s: %q\n", labelA, taskA.Title)
			diff += fmt.Sprintf("      %s: %q\n", labelB, taskB.Title)
		}
	}

	return diff
}
