package testutil

import (
	"testing"

	"github.com/neongreen/mono/tk/internal/types"
)

// NOTE: AssertConvergence and other reducer-dependent helpers moved to reducer package
// to avoid import cycles (testutil/machine.go uses reducer.BuildFromEvents)

// AssertConvergence verifies that two machines have converged to the same state
// MOVED to reducer_test.go to avoid import cycle
/* func AssertConvergence(t *testing.T, machineA, machineB *Machine) {
	t.Helper()

	stateA := machineA.GetState()
	stateB := machineB.GetState()

	// Check task count
	if len(stateA.Tasks()) != len(stateB.Tasks()) {
		t.Fatalf("Convergence failed: machine %s has %d tasks, machine %s has %d tasks",
			machineA.NodeID, len(stateA.Tasks()),
			machineB.NodeID, len(stateB.Tasks()))
	}

	// Check all tasks match
	for uuid, taskA := range stateA.Tasks() {
		taskB, exists := stateB.GetTask(uuid)
		if !exists {
			t.Fatalf("Convergence failed: task %s exists on machine %s but not on machine %s",
				uuid, machineA.NodeID, machineB.NodeID)
		}

		// Check title
		if taskA.Title != taskB.Title {
			t.Fatalf("Convergence failed: task %s has different titles:\n  Machine %s: %q\n  Machine %s: %q",
				uuid, machineA.NodeID, taskA.Title, machineB.NodeID, taskB.Title)
		}

		// Check status
		for axis, statusA := range taskA.Axes {
			statusB, exists := taskB.Axes[axis]
			if !exists {
				t.Fatalf("Convergence failed: task %s has axis %s on machine %s but not on machine %s",
					uuid, axis, machineA.NodeID, machineB.NodeID)
			}

			if statusA.Effective != statusB.Effective {
				t.Fatalf("Convergence failed: task %s axis %s has different effective status:\n  Machine %s: %q\n  Machine %s: %q",
					uuid, axis, machineA.NodeID, statusA.Effective, machineB.NodeID, statusB.Effective)
			}
		}
	}

	t.Logf("✓ Machines %s and %s converged (%d tasks)", machineA.NodeID, machineB.NodeID, len(stateA.Tasks()))
} */

// AssertLamportOrderPreserved verifies that events are ordered by Lamport timestamp
func AssertLamportOrderPreserved(t *testing.T, events []types.Event) {
	t.Helper()

	for i := 1; i < len(events); i++ {
		if events[i].TS < events[i-1].TS {
			t.Fatalf("Lamport order violated: event %d has TS %d but event %d has TS %d",
				i-1, events[i-1].TS, i, events[i].TS)
		}
	}

	t.Logf("✓ Lamport order preserved (%d events)", len(events))
}
