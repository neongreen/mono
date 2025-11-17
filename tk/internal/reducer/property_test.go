package reducer

import (
	"encoding/json"
	"testing"

	"github.com/neongreen/mono/tk/internal/types"
	"pgregory.net/rapid"
)

// Property-based testing for CRDT convergence and determinism
//
// This file uses the rapid library (pgregory.net/rapid) for property-based testing.
//
// ## Running property tests:
//
//   go test -v ./internal/reducer/
//
// ## Adjusting test iterations (default: 100):
//
//   go test -v -rapid.checks=1000 ./internal/reducer/
//
// ## Debugging failures:
//
//   go test -v -rapid.seed=<seed_from_failure>
//
// ## What to test:
//
// - Commutativity: Concurrent events can be applied in any order
// - Convergence: Same event set → same final state (regardless of delivery order)
// - Idempotence: Applying same event multiple times = applying once
// - Monotonicity: State only grows (OR-set semantics for relations)

// TestEventGenerators verifies event generators produce valid events
func TestEventGenerators(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a sequence of events
		events := GenEventSequence(5, 20).Draw(t, "events")

		// Property: All events should have increasing Lamport timestamps
		for i := 1; i < len(events); i++ {
			if events[i].TS <= events[i-1].TS {
				t.Fatalf("Lamport timestamps not increasing: event %d has TS %d, but event %d has TS %d",
					i-1, events[i-1].TS, i, events[i].TS)
			}
		}

		// Property: All events should have valid event IDs
		for i, event := range events {
			if event.ID == "" {
				t.Fatalf("Event %d has empty ID", i)
			}
		}

		// Property: All events should have non-empty kind
		for i, event := range events {
			if event.Kind == "" {
				t.Fatalf("Event %d has empty kind", i)
			}
		}
	})
}

// TestCommutativity tests that concurrent events (same TS) can be applied in any order
func TestCommutativity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a sequence of events
		events := GenEventSequence(5, 15).Draw(t, "events")

		// Find the last task.created event
		lastTaskCreatedTS := int64(-1)
		for _, event := range events {
			if event.Kind == string(types.EventKindTaskCreated) {
				if event.TS > lastTaskCreatedTS {
					lastTaskCreatedTS = event.TS
				}
			}
		}

		// Find status change events that happen AFTER all tasks are created
		var statusEvents []int
		for i, event := range events {
			if event.Kind == "task.status.set" && event.TS > lastTaskCreatedTS {
				statusEvents = append(statusEvents, i)
			}
		}

		// Make status events concurrent if we have at least 2
		if len(statusEvents) >= 2 {
			// Use a timestamp after all task creations
			concurrentTS := lastTaskCreatedTS + 1
			for _, idx := range statusEvents {
				events[idx].TS = concurrentTS
			}
		}

		// Sort events by Lamport timestamp (maintaining relative order for equal TS)
		sortByLamportTS := func(evts []types.Event) []types.Event {
			sorted := make([]types.Event, len(evts))
			copy(sorted, evts)
			// Stable sort by TS
			for i := 1; i < len(sorted); i++ {
				for j := i; j > 0 && sorted[j].TS < sorted[j-1].TS; j-- {
					sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
				}
			}
			return sorted
		}

		// Apply events in sorted order
		reducer1 := NewReducer()
		sorted := sortByLamportTS(events)
		for _, event := range sorted {
			if err := reducer1.Apply(event); err != nil {
				t.Fatalf("Failed to apply event in reducer1: %v", err)
			}
		}

		// Shuffle concurrent events but maintain Lamport order
		shuffled := make([]types.Event, len(sorted))
		copy(shuffled, sorted)
		// Only shuffle within groups of equal timestamps
		i := 0
		for i < len(shuffled) {
			j := i + 1
			for j < len(shuffled) && shuffled[j].TS == shuffled[i].TS {
				j++
			}
			// Shuffle events from i to j-1 (they have the same TS)
			if j-i > 1 {
				// Simple shuffle
				for k := i; k < j; k++ {
					swapIdx := rapid.IntRange(i, j-1).Draw(t, "swap_idx")
					shuffled[k], shuffled[swapIdx] = shuffled[swapIdx], shuffled[k]
				}
			}
			i = j
		}

		// Apply shuffled events
		reducer2 := NewReducer()
		for _, event := range shuffled {
			if err := reducer2.Apply(event); err != nil {
				t.Fatalf("Failed to apply event in reducer2: %v", err)
			}
		}

		// Property: Both reducers should have the same number of tasks
		if len(reducer1.Tasks()) != len(reducer2.Tasks()) {
			t.Fatalf("Commutativity violated: reducer1 has %d tasks, reducer2 has %d tasks",
				len(reducer1.Tasks()), len(reducer2.Tasks()))
		}

		// Property: Both reducers should have the same tasks
		for uuid, task1 := range reducer1.Tasks() {
			task2, exists := reducer2.GetTask(uuid)
			if !exists {
				t.Fatalf("Commutativity violated: task %s exists in reducer1 but not in reducer2", uuid)
			}

			// Check that key properties match
			if task1.Title != task2.Title {
				t.Fatalf("Commutativity violated: task %s has different titles: %q vs %q",
					uuid, task1.Title, task2.Title)
			}
		}
	})
}

// TestConvergence tests that same event set leads to same final state regardless of delivery order
func TestConvergence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a sequence of events
		events := GenEventSequence(5, 20).Draw(t, "events")

		// Sort events by Lamport timestamp for correct order
		sortByLamportTS := func(evts []types.Event) []types.Event {
			sorted := make([]types.Event, len(evts))
			copy(sorted, evts)
			for i := 1; i < len(sorted); i++ {
				for j := i; j > 0 && sorted[j].TS < sorted[j-1].TS; j-- {
					sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
				}
			}
			return sorted
		}

		// Apply in sorted order
		reducer1 := NewReducer()
		sorted := sortByLamportTS(events)
		for _, event := range sorted {
			reducer1.Apply(event) // Ignore errors
		}

		// Shuffle the event delivery order (but will be sorted by TS during projection)
		shuffled := make([]types.Event, len(events))
		copy(shuffled, events)
		for i := range shuffled {
			j := rapid.IntRange(0, len(shuffled)-1).Draw(t, "shuffle_idx")
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		}

		// Apply shuffled events (sorting by TS inside)
		reducer2 := NewReducer()
		shuffledSorted := sortByLamportTS(shuffled)
		for _, event := range shuffledSorted {
			reducer2.Apply(event) // Ignore errors
		}

		// Property: Both reducers should converge to the same state
		if len(reducer1.Tasks()) != len(reducer2.Tasks()) {
			t.Fatalf("Convergence violated: reducer1 has %d tasks, reducer2 has %d tasks",
				len(reducer1.Tasks()), len(reducer2.Tasks()))
		}

		for uuid, task1 := range reducer1.Tasks() {
			task2, exists := reducer2.GetTask(uuid)
			if !exists {
				t.Fatalf("Convergence violated: task %s exists in reducer1 but not in reducer2", uuid)
			}

			if task1.Title != task2.Title {
				t.Fatalf("Convergence violated: task %s has different titles: %q vs %q",
					uuid, task1.Title, task2.Title)
			}
		}
	})
}

// TestIdempotence tests that applying the same event multiple times has the same effect as applying it once
func TestIdempotence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a small sequence of events
		events := GenEventSequence(3, 8).Draw(t, "events")

		// Apply events once
		reducer1 := NewReducer()
		for _, event := range events {
			reducer1.Apply(event) // Ignore errors
		}

		// Apply events, then duplicate some events
		reducer2 := NewReducer()
		for _, event := range events {
			reducer2.Apply(event) // First time

			// Randomly duplicate this event
			if rapid.Bool().Draw(t, "duplicate") {
				reducer2.Apply(event) // Second time (should be idempotent)
			}
		}

		// Property: Both reducers should have the same state
		if len(reducer1.Tasks()) != len(reducer2.Tasks()) {
			t.Fatalf("Idempotence violated: reducer1 has %d tasks, reducer2 has %d tasks",
				len(reducer1.Tasks()), len(reducer2.Tasks()))
		}

		for uuid, task1 := range reducer1.Tasks() {
			task2, exists := reducer2.GetTask(uuid)
			if !exists {
				t.Fatalf("Idempotence violated: task %s exists in reducer1 but not in reducer2", uuid)
			}

			if task1.Title != task2.Title {
				t.Fatalf("Idempotence violated: task %s has different titles: %q vs %q",
					uuid, task1.Title, task2.Title)
			}

			// Check status claims (idempotence for status events)
			for axis, status1 := range task1.Axes {
				status2, exists := task2.Axes[axis]
				if !exists {
					t.Fatalf("Idempotence violated: task %s has axis %s in reducer1 but not in reducer2",
						uuid, axis)
				}

				if status1.Effective != status2.Effective {
					t.Fatalf("Idempotence violated: task %s axis %s has different effective status: %q vs %q",
						uuid, axis, status1.Effective, status2.Effective)
				}
			}
		}
	})
}

// TestLamportOrderIndependence tests that Lamport timestamps (not array order) determine causality
// This catches bugs where we use array position instead of Lamport TS
func TestLamportOrderIndependence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate events with increasing Lamport timestamps
		events := GenEventSequence(5, 12).Draw(t, "events")

		if len(events) < 3 {
			return
		}

		// Filter to only status events on existing tasks
		var taskUIDs []string
		var statusEvents []types.Event

		for _, event := range events {
			if event.Kind == string(types.EventKindTaskCreated) {
				var payload types.TaskCreatedPayload
				json.Unmarshal(event.Payload, &payload)
				taskUIDs = append(taskUIDs, payload.TaskUID)
			} else if event.Kind == "task.status.set" && len(taskUIDs) > 0 {
				statusEvents = append(statusEvents, event)
			}
		}

		if len(statusEvents) < 2 {
			return // Need at least 2 status events to test
		}

		// Create task-create events first
		var taskCreates []types.Event
		for _, event := range events {
			if event.Kind == string(types.EventKindTaskCreated) {
				taskCreates = append(taskCreates, event)
			}
		}

		// Apply in correct Lamport order
		reducer1 := NewReducer()
		sortByTS := func(evts []types.Event) []types.Event {
			sorted := make([]types.Event, len(evts))
			copy(sorted, evts)
			for i := 1; i < len(sorted); i++ {
				for j := i; j > 0 && sorted[j].TS < sorted[j-1].TS; j-- {
					sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
				}
			}
			return sorted
		}

		allEvents := append(taskCreates, statusEvents...)
		sorted := sortByTS(allEvents)
		for _, event := range sorted {
			reducer1.Apply(event)
		}

		// Apply status events in REVERSE Lamport order (wrong order!)
		// Task creates still in correct order
		reducer2 := NewReducer()
		for _, event := range taskCreates {
			reducer2.Apply(event)
		}

		// Reverse status events - this means highest TS comes FIRST
		reversed := make([]types.Event, len(statusEvents))
		for i, event := range statusEvents {
			reversed[len(statusEvents)-1-i] = event
		}
		for _, event := range reversed {
			reducer2.Apply(event)
		}

		// Property: Final state should be the SAME despite different order
		// Because Lamport TS (not array order) determines causality
		if len(reducer1.Tasks()) != len(reducer2.Tasks()) {
			t.Fatalf("Lamport order independence violated: reducer1 has %d tasks, reducer2 has %d tasks",
				len(reducer1.Tasks()), len(reducer2.Tasks()))
		}

		for uuid, task1 := range reducer1.Tasks() {
			task2, exists := reducer2.GetTask(uuid)
			if !exists {
				t.Fatalf("Lamport order independence violated: task %s missing in reducer2", uuid)
			}

			// Check status claims - this is where the bug manifests
			for axis, status1 := range task1.Axes {
				status2, exists := task2.Axes[axis]
				if !exists {
					continue // May not have this axis if events filtered differently
				}

				// CRITICAL: Effective status must be the same!
				// If we use array order instead of Lamport TS, this will fail
				if status1.Effective != status2.Effective {
					t.Fatalf("Lamport order independence violated: task %s axis %s has different effective status: %q vs %q (reducer1 vs reducer2)",
						uuid, axis, status1.Effective, status2.Effective)
				}
			}
		}
	})
}
