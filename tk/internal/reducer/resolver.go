package reducer

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/types"
)

// ============================================================================
// Identifier Resolution with Synthetic Project Creation
// ============================================================================
//
// The resolver transforms lax (permissive) identifiers from historical events
// into validated identifiers for use in reducer state.
//
// LEGACY FORMATS HANDLED:
// -----------------------
// Tasks:
//   - Legacy: "123", "foo-5" → Look up in taskByID map
//   - Modern: "tsk_<ulid>" → Validate and use
//
// Projects:
//   - Legacy: "lovable", "my-project" → Look up in projectByName map
//   - Malformed: "abc" (not valid, not in map) → Create synthetic project
//   - Modern: "prj_<ulid>" → Validate and use
//
// SYNTHETIC PROJECT CREATION:
// ---------------------------
// When a project identifier cannot be resolved:
// 1. Generate deterministic UID using counter: prj_00000000000000000000000001
// 2. Create synthetic project in memory (not persisted to database)
// 3. Add to reducer maps (projects, projectByName)
// 4. Mark as temporary
// 5. Continue processing
//
// After replay, warn about non-empty temporary projects (manual cleanup required)
//
// ============================================================================

// resolveTaskUID resolves a lax task identifier to a validated UID string.
func (r *Reducer) resolveTaskUID(lax string) (string, error) {
	// Try as valid UID first
	uid := types.TaskUID(lax)
	if err := uid.Validate(); err == nil {
		return lax, nil
	}

	// Legacy fallback: look up in taskByID
	if resolved, ok := r.taskByID[lax]; ok {
		return resolved, nil
	}

	return "", fmt.Errorf("cannot resolve task identifier %q: not a valid UID and not found in taskByID map", lax)
}

// resolveProjectUID resolves a lax project identifier to a validated UID string.
// Creates synthetic projects for unresolvable identifiers.
func (r *Reducer) resolveProjectUID(lax string) (string, error) {
	// Try as valid UID first
	uid := types.ProjectUID(lax)
	if err := uid.Validate(); err == nil {
		return lax, nil
	}

	// Look up in projectByName (handles names and aliases)
	if resolved, ok := r.projectByName[lax]; ok {
		return resolved, nil
	}

	// Not found - create synthetic project with deterministic UID
	return r.createSyntheticProject(lax), nil
}

// createSyntheticProject creates a temporary project for an unresolvable reference.
// Uses a counter to generate deterministic UIDs so replay is consistent.
// The project exists only in reducer memory - NOT persisted to database.
func (r *Reducer) createSyntheticProject(name string) string {
	// Generate deterministic UID using counter
	r.syntheticProjectCounter++
	uid := fmt.Sprintf("prj_%026d", r.syntheticProjectCounter)

	// Create synthetic project in memory only
	project := &types.Project{
		UID:         types.ProjectUID(uid),
		Type:        types.ProjectTypeLocal,
		Name:        name, // Use corrupt value as name
		Description: "Temporary project auto-created from malformed reference",
		// Note: CreatedAt and CreatedBy are zero values - this is intentional
		// to distinguish synthetic projects from real ones
	}

	r.projects[uid] = project
	r.projectByName[name] = uid
	r.temporaryProjects[uid] = true

	return uid
}
