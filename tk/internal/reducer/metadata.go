package reducer

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/types"
)

// applyTaskMetaSet applies a task.meta.set event
func (r *Reducer) applyTaskMetaSet(e types.Event) error {
	var payload types.TaskMetaSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.meta.set payload: %w", err)
	}

	// Validate that value is valid JSON
	if !json.Valid(payload.Value) {
		return fmt.Errorf("invalid JSON in metadata value for key %s", payload.Key)
	}

	// Get or create task
	task, ok := r.tasks[payload.TaskUUID]
	if !ok {
		return fmt.Errorf("task not found: %s", payload.TaskUUID)
	}

	// Initialize metadata map if needed
	if task.Metadata == nil {
		task.Metadata = make(map[string]types.MetadataStatus)
	}

	// Get or create metadata status for this key
	metaStatus, ok := task.Metadata[payload.Key]
	if !ok {
		metaStatus = types.MetadataStatus{
			Claims: []types.MetadataClaim{},
		}
	}

	// Extract role from event (default to "human" if missing)
	role := e.Role
	if role == "" {
		role = "human"
	}

	// Create new claim
	newClaim := types.MetadataClaim{
		Value:     payload.Value,
		Role:      role,
		Tentative: false, // Will be set by resolution
		TS:        e.TS,
	}

	// Append claim to list
	metaStatus.Claims = append(metaStatus.Claims, newClaim)

	// Resolve effective value using authority lattice
	resolveMetadataValue(&metaStatus)

	// Store back
	task.Metadata[payload.Key] = metaStatus
	task.UpdatedAt = e.CreatedAt

	return nil
}

// resolveMetadataValue determines the effective value and marks tentative claims
// Uses same authority lattice as status axes: human > qa > rel > agent > bot
func resolveMetadataValue(status *types.MetadataStatus) {
	if len(status.Claims) == 0 {
		status.Effective = nil
		return
	}

	// Find highest authority level among all claims
	highestAuthority := 0
	for _, claim := range status.Claims {
		authority := types.GetRoleAuthority(claim.Role)
		if authority > highestAuthority {
			highestAuthority = authority
		}
	}

	// Find latest timestamp among highest authority claims
	var latestTS int64
	for _, claim := range status.Claims {
		authority := types.GetRoleAuthority(claim.Role)
		if authority == highestAuthority && claim.TS > latestTS {
			latestTS = claim.TS
		}
	}

	// Find the effective claim (highest authority, latest timestamp)
	var effectiveClaim *types.MetadataClaim
	for i := range status.Claims {
		claim := &status.Claims[i]
		authority := types.GetRoleAuthority(claim.Role)
		if authority == highestAuthority && claim.TS == latestTS {
			effectiveClaim = claim
			break
		}
	}

	if effectiveClaim == nil {
		// Should never happen, but handle gracefully
		status.Effective = status.Claims[0].Value
		return
	}

	// Set effective value
	status.Effective = effectiveClaim.Value

	// Mark all claims as tentative or not
	for i := range status.Claims {
		claimAuthority := types.GetRoleAuthority(status.Claims[i].Role)

		// A claim is tentative if:
		// 1. It has lower authority than the highest, OR
		// 2. It has same authority but earlier timestamp
		status.Claims[i].Tentative = claimAuthority < highestAuthority ||
			(claimAuthority == highestAuthority && status.Claims[i].TS < latestTS)
	}
}
