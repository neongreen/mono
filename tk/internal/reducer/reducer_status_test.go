package reducer

import (
	"testing"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestReducer_StatusSet(t *testing.T) {
	test := NewReducerTest(t)

	projectUID := test.CreateProject("test-project")
	task := test.CreateTask("alice", "Test task", projectUID)

	task.SetStatus("in_progress")
	task.AssertStatus("in_progress")
	task.AssertClaimCount("generic", 1)
}

func TestReducer_AuthorityResolution(t *testing.T) {
	test := NewReducerTest(t)

	projectUID := test.CreateProject("test-project")
	task := test.CreateTask("alice", "Test task", projectUID)

	// Agent sets status to done
	task.SetStatusAs("done", "agent", "claude")
	task.AssertStatus("done")

	// Human overrides (higher authority)
	task.SetStatusAs("in_progress", "human", "alice")
	task.AssertStatus("in_progress") // Human wins

	// Verify claims exist and tentative flags are correct
	taskData, _ := test.Reducer().GetTask(task.UID)
	axis := taskData.Axes["generic"]

	var humanClaim, agentClaim *types.Claim
	for i := range axis.Claims {
		if axis.Claims[i].Role == "human" {
			humanClaim = &axis.Claims[i]
		}
		if axis.Claims[i].Role == "agent" {
			agentClaim = &axis.Claims[i]
		}
	}

	if humanClaim == nil || agentClaim == nil {
		t.Fatal("Missing claims")
	}

	if humanClaim.Tentative {
		t.Error("Human claim should not be tentative")
	}

	if !agentClaim.Tentative {
		t.Error("Agent claim should be tentative")
	}
}

func TestGetRoleAuthority(t *testing.T) {
	tests := []struct {
		role     string
		expected int
	}{
		{"human", 5},
		{"qa", 4},
		{"rel", 3},
		{"agent", 2},
		{"bot", 1},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			authority := types.GetRoleAuthority(tt.role)
			if authority != tt.expected {
				t.Errorf("Expected authority %d for role %s, got %d", tt.expected, tt.role, authority)
			}
		})
	}
}
