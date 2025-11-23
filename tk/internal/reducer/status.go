package reducer

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/neongreen/mono/tk/internal/types"
)

func (r *Reducer) applyTaskStatusSet(e types.Event) error {
	var payload types.TaskStatusSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.status.set payload: %w", err)
	}

	// Resolve task ID to UUID
	taskUUID := payload.TaskUUID
	if taskUUID == "" {
		// Legacy event fallback - only for reading old pre-UUID events
		// See tk-190: can be removed after running 'tk migrate compact-remote' on all machines
		var ok bool
		taskUUID, ok = r.taskByID[payload.TaskID]
		if !ok {
			return fmt.Errorf("task not found: %s", payload.TaskID)
		}
	}

	task, ok := r.tasks[taskUUID]
	if !ok {
		return fmt.Errorf("task UUID not found: %s", taskUUID)
	}

	// Get or create axis status
	axis, ok := task.Axes[payload.Axis]
	if !ok {
		axis = types.AxisStatus{
			Claims: []types.Claim{},
		}
	}

	// Add new claim
	newClaim := types.Claim{
		State:     payload.State,
		Role:      payload.Role,
		Tentative: false,
		TS:        e.TS,
	}
	axis.Claims = append(axis.Claims, newClaim)

	// Resolve effective status based on authority
	r.resolveEffectiveStatus(&axis)

	task.Axes[payload.Axis] = axis
	task.UpdatedAt = e.CreatedAt

	return nil
}

// resolveEffectiveStatus resolves which claim is effective based on authority
func (r *Reducer) resolveEffectiveStatus(axis *types.AxisStatus) {
	if len(axis.Claims) == 0 {
		return
	}

	// Group claims by state to find concurrent claims
	stateGroups := make(map[string][]types.Claim)
	for _, claim := range axis.Claims {
		stateGroups[claim.State] = append(stateGroups[claim.State], claim)
	}

	// Find the claim with highest authority among the latest claims
	// First, find the latest timestamp
	latestTS := int64(0)
	for _, claim := range axis.Claims {
		if claim.TS > latestTS {
			latestTS = claim.TS
		}
	}

	// Get all claims at the latest timestamp (concurrent claims)
	var concurrentClaims []types.Claim
	for i, claim := range axis.Claims {
		if claim.TS == latestTS {
			concurrentClaims = append(concurrentClaims, axis.Claims[i])
		}
	}

	// Sort by authority (highest first)
	sort.Slice(concurrentClaims, func(i, j int) bool {
		return types.GetRoleAuthority(concurrentClaims[i].Role) > types.GetRoleAuthority(concurrentClaims[j].Role)
	})

	// The highest authority claim is effective
	effectiveClaim := concurrentClaims[0]
	axis.Effective = effectiveClaim.State

	// Mark all claims as tentative or not
	highestAuthority := types.GetRoleAuthority(effectiveClaim.Role)
	for i := range axis.Claims {
		claimAuthority := types.GetRoleAuthority(axis.Claims[i].Role)
		// A claim is tentative if it has lower authority than the effective claim
		// OR if it's not the latest claim with that state
		axis.Claims[i].Tentative = claimAuthority < highestAuthority ||
			axis.Claims[i].TS < latestTS
	}
}
