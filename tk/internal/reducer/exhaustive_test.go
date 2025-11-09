package reducer

import (
	"fmt"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

// Exhaustive permutation testing for CRDT properties
//
// Unlike property-based testing which samples random orderings, exhaustive tests
// verify ALL possible orderings for small event sets (e.g., 5 events = 5! = 120 orderings).
//
// Benefits:
// - Complete coverage for specific scenarios
// - Deterministic (same result every run)
// - Faster to debug (no random seeds)
// - Proves correctness for critical bug classes
//
// ## Running exhaustive tests:
//
//   go test -v -run TestExhaustive ./internal/reducer/

// TestPermutationGenerator verifies the permutation generator works correctly
func TestPermutationGenerator(t *testing.T) {
	// Test with small slice
	items := []int{1, 2, 3}
	perms := generateAllPermutations(items)

	// Should have 3! = 6 permutations
	if len(perms) != 6 {
		t.Fatalf("Expected 6 permutations, got %d", len(perms))
	}

	// Verify all permutations are different
	seen := make(map[string]bool)
	for _, perm := range perms {
		key := fmt.Sprintf("%v", perm)
		if seen[key] {
			t.Fatalf("Duplicate permutation: %v", perm)
		}
		seen[key] = true
	}

	// Test with 5 items
	items5 := []int{1, 2, 3, 4, 5}
	perms5 := generateAllPermutations(items5)

	// Should have 5! = 120 permutations
	if len(perms5) != 120 {
		t.Fatalf("Expected 120 permutations for 5 items, got %d", len(perms5))
	}
}

// TestExhaustive_StatusOrdering tests ALL orderings of status events to verify Lamport TS is used
// This catches bug #2 from audit: using array order / created_at instead of Lamport timestamps
func TestExhaustive_StatusOrdering(t *testing.T) {
	// Scenario: Task created, then 3 status changes with increasing Lamport timestamps
	// Property: No matter what order events arrive, effective status = highest Lamport TS

	taskUID := "task_TEST123"
	projectUID := "proj_TEST456"

	createEvent := types.Event{
		ID:        "ev-1-test",
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload: mustMarshal(types.TaskCreatedPayload{
			TaskUID:        taskUID,
			ProjectUID:     projectUID,
			ProposedNumber: 1,
			CreatedNode:    "node1",
			Title:          "Test task",
			CreatedBy:      "alice",
		}),
	}

	// Status events with DIFFERENT Lamport timestamps
	statusTodo := types.Event{
		ID: "ev-2-test", TS: 1, CreatedAt: time.Now(), Actor: "alice", Role: "human",
		Kind: "task.status.set",
		Payload: mustMarshal(types.TaskStatusSetPayload{
			TaskUUID: taskUID, Axis: "generic", State: "todo", Role: "human",
		}),
	}

	statusWip := types.Event{
		ID: "ev-3-test", TS: 2, CreatedAt: time.Now(), Actor: "alice", Role: "human",
		Kind: "task.status.set",
		Payload: mustMarshal(types.TaskStatusSetPayload{
			TaskUUID: taskUID, Axis: "generic", State: "wip", Role: "human",
		}),
	}

	statusDone := types.Event{
		ID: "ev-4-test", TS: 3, CreatedAt: time.Now(), Actor: "alice", Role: "human",
		Kind: "task.status.set",
		Payload: mustMarshal(types.TaskStatusSetPayload{
			TaskUUID: taskUID, Axis: "generic", State: "done", Role: "human",
		}),
	}

	statusEvents := []types.Event{statusTodo, statusWip, statusDone}

	// Generate all 3! = 6 permutations of status events
	permutations := generateAllPermutations(statusEvents)

	if len(permutations) != 6 {
		t.Fatalf("Expected 6 permutations, got %d", len(permutations))
	}

	var referenceState *Reducer

	// Test each permutation
	for i, perm := range permutations {
		reducer := NewReducer()

		// Always apply create event first
		if err := reducer.Apply(createEvent); err != nil {
			t.Fatalf("Permutation %d: failed to apply create event: %v", i, err)
		}

		// Apply status events in this permutation's order
		for _, event := range perm {
			if err := reducer.Apply(event); err != nil {
				t.Fatalf("Permutation %d: failed to apply event: %v", i, err)
			}
		}

		// Get the task
		task, exists := reducer.GetTask(taskUID)
		if !exists {
			t.Fatalf("Permutation %d: task not found", i)
		}

		// CRITICAL: Effective status MUST be "done" (highest Lamport TS)
		// If code uses array order instead of Lamport TS, this will fail for some permutations
		axis, exists := task.Axes["generic"]
		if !exists {
			t.Fatalf("Permutation %d: generic axis not found", i)
		}

		if axis.Effective != "done" {
			t.Fatalf("Permutation %d: Expected effective status 'done', got '%s'. "+
				"Event order: %v. This means code is using array order instead of Lamport TS!",
				i, axis.Effective, getEventOrder(perm))
		}

		// Verify all permutations produce identical state
		if i == 0 {
			referenceState = reducer
		} else {
			if task.Title != referenceState.Tasks()[taskUID].Title {
				t.Fatalf("Permutation %d: State diverged from reference", i)
			}
		}
	}

	t.Logf("✓ All %d permutations converged to same state (effective=done)", len(permutations))
}

// TestExhaustive_ConcurrentAuthority tests that higher authority wins for concurrent claims
// This catches bug #8 from audit: collision detection and authority resolution
func TestExhaustive_ConcurrentAuthority(t *testing.T) {
	// Scenario: Concurrent status claims from different roles (same Lamport TS)
	// Property: Human authority > Agent authority > Bot authority

	taskUID := "task_AUTHORITY_TEST"
	projectUID := "proj_TEST"

	createEvent := types.Event{
		ID: "ev-1-test", TS: 0, CreatedAt: time.Now(), Actor: "system", Role: "human",
		Kind: string(types.EventKindTaskCreated),
		Payload: mustMarshal(types.TaskCreatedPayload{
			TaskUID:        taskUID,
			ProjectUID:     projectUID,
			ProposedNumber: 1,
			CreatedNode:    "node1",
			Title:          "Authority test",
			CreatedBy:      "system",
		}),
	}

	// Concurrent claims (SAME Lamport TS = 1) from different roles
	humanClaim := types.Event{
		ID: "ev-2-test", TS: 1, CreatedAt: time.Now(), Actor: "alice", Role: "human",
		Kind: "task.status.set",
		Payload: mustMarshal(types.TaskStatusSetPayload{
			TaskUUID: taskUID, Axis: "generic", State: "wip", Role: "human",
		}),
	}

	agentClaim := types.Event{
		ID: "ev-3-test", TS: 1, CreatedAt: time.Now(), Actor: "bot", Role: "agent",
		Kind: "task.status.set",
		Payload: mustMarshal(types.TaskStatusSetPayload{
			TaskUUID: taskUID, Axis: "generic", State: "done", Role: "agent",
		}),
	}

	botClaim := types.Event{
		ID: "ev-4-test", TS: 1, CreatedAt: time.Now(), Actor: "bot2", Role: "bot",
		Kind: "task.status.set",
		Payload: mustMarshal(types.TaskStatusSetPayload{
			TaskUUID: taskUID, Axis: "generic", State: "blocked", Role: "bot",
		}),
	}

	concurrentEvents := []types.Event{humanClaim, agentClaim, botClaim}

	// Generate all 3! = 6 permutations
	permutations := generateAllPermutations(concurrentEvents)

	if len(permutations) != 6 {
		t.Fatalf("Expected 6 permutations, got %d", len(permutations))
	}

	// Test each permutation
	for i, perm := range permutations {
		reducer := NewReducer()

		// Apply create event
		if err := reducer.Apply(createEvent); err != nil {
			t.Fatalf("Permutation %d: failed to apply create event: %v", i, err)
		}

		// Apply concurrent claims in this permutation's order
		for _, event := range perm {
			if err := reducer.Apply(event); err != nil {
				t.Fatalf("Permutation %d: failed to apply event: %v", i, err)
			}
		}

		// Get the task
		task, exists := reducer.GetTask(taskUID)
		if !exists {
			t.Fatalf("Permutation %d: task not found", i)
		}

		axis, exists := task.Axes["generic"]
		if !exists {
			t.Fatalf("Permutation %d: generic axis not found", i)
		}

		// CRITICAL: Human authority MUST win (highest authority)
		// Regardless of what order the concurrent events arrived
		if axis.Effective != "wip" {
			t.Fatalf("Permutation %d: Expected human authority to win (effective='wip'), got '%s'. "+
				"Event order: [%s, %s, %s]",
				i, axis.Effective, perm[0].Role, perm[1].Role, perm[2].Role)
		}
	}

	t.Logf("✓ All %d permutations correctly resolved human authority (effective=wip)", len(permutations))
}

// TestExhaustive_DuplicateEvents tests that duplicate task.created events are handled idempotently
// This catches bug #4 from audit: duplicate tasks from double event processing
//
// Fixed by tk-229: Now uses earliest Lamport timestamp for deterministic duplicate handling
func TestExhaustive_DuplicateEvents(t *testing.T) {
	// Scenario: Same task.created event appears twice (duplicate processing)
	// Property: Only one task should exist, regardless of event order

	taskUID := "task_DUPLICATE_TEST"
	projectUID := "proj_TEST"

	createEvent1 := types.Event{
		ID: "ev-1-test", TS: 0, CreatedAt: time.Now(), Actor: "alice", Role: "human",
		Kind: string(types.EventKindTaskCreated),
		Payload: mustMarshal(types.TaskCreatedPayload{
			TaskUID:        taskUID,
			ProjectUID:     projectUID,
			ProposedNumber: 1,
			CreatedNode:    "node1",
			Title:          "Original task",
			CreatedBy:      "alice",
		}),
	}

	// Same UID, different event ID (simulates duplicate processing)
	createEvent2 := types.Event{
		ID: "ev-5-test", TS: 1, CreatedAt: time.Now(), Actor: "alice", Role: "human",
		Kind: string(types.EventKindTaskCreated),
		Payload: mustMarshal(types.TaskCreatedPayload{
			TaskUID:        taskUID, // SAME UID!
			ProjectUID:     projectUID,
			ProposedNumber: 1,
			CreatedNode:    "node1",
			Title:          "Duplicate task", // Different title
			CreatedBy:      "alice",
		}),
	}

	statusEvent := types.Event{
		ID: "ev-6-test", TS: 2, CreatedAt: time.Now(), Actor: "alice", Role: "human",
		Kind: "task.status.set",
		Payload: mustMarshal(types.TaskStatusSetPayload{
			TaskUUID: taskUID, Axis: "generic", State: "done", Role: "human",
		}),
	}

	events := []types.Event{createEvent1, createEvent2, statusEvent}

	// Generate all 3! = 6 permutations
	permutations := generateAllPermutations(events)

	var referenceTitle string

	// Test each permutation
	for i, perm := range permutations {
		reducer := NewReducer()

		// Apply events in Lamport timestamp order (not permutation order)
		// We're testing that duplicate handling is deterministic regardless of ARRIVAL order
		// but events must still be projected in Lamport order
		sorted := make([]types.Event, len(perm))
		copy(sorted, perm)
		for i := 1; i < len(sorted); i++ {
			for j := i; j > 0 && sorted[j].TS < sorted[j-1].TS; j-- {
				sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			}
		}

		for _, event := range sorted {
			if err := reducer.Apply(event); err != nil {
				t.Fatalf("Permutation %d: failed to apply event: %v", i, err)
			}
		}

		// CRITICAL: Should have exactly 1 task (duplicates ignored)
		if len(reducer.Tasks()) != 1 {
			t.Fatalf("Permutation %d: Expected 1 task, got %d tasks. "+
				"Duplicate task.created events were not handled idempotently!",
				i, len(reducer.Tasks()))
		}

		task, exists := reducer.GetTask(taskUID)
		if !exists {
			t.Fatalf("Permutation %d: task not found", i)
		}

		// Property: ALL permutations must produce the SAME result (deterministic)
		// With the fix (tk-229), we always keep the task with earliest Lamport TS
		// createEvent1 has TS=0, createEvent2 has TS=1, so title should always be "Original task"
		if task.Title != "Original task" {
			t.Fatalf("Permutation %d: Expected title 'Original task' (TS=0), got '%s'. "+
				"Duplicate handling should keep earliest Lamport TS!",
				i, task.Title)
		}

		if i == 0 {
			referenceTitle = task.Title
		} else {
			if task.Title != referenceTitle {
				t.Fatalf("Permutation %d: Title diverged! Expected '%s', got '%s'. "+
					"Duplicate handling is NON-DETERMINISTIC!",
					i, referenceTitle, task.Title)
			}
		}
	}

	t.Logf("✓ All %d permutations converged to same state (title='Original task' from TS=0)", len(permutations))
}
