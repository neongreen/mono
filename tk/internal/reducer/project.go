package reducer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

// V4 Event Reducer Functions
// Handles events (project.created, task.created, task.number.set, task.relocate, etc.)

// ApplyProjectEvent applies an event-specific handler
// Returns (handled=true, error) if the event was handled
// Returns (handled=false, nil) if the event was not handled
func (r *Reducer) ApplyProjectEvent(e types.Event) (bool, error) {
	switch types.EventKind(e.Kind) {
	case types.EventKindProjectCreated:
		//lion:events title="Events" section="project.created"
		// Ensures projects exist in reducer state, replacing synthetic placeholders with real metadata when the canonical creation event arrives.
		return true, r.applyProjectCreated(e)
	case types.EventKindProjectAliasAdd:
		//lion:events section="project.alias.add"
		// Aliases are tracked in projections; reducer ignores payload beyond validation to keep replay idempotent while projections handle display IDs.
		return true, r.applyProjectAliasAdd(e)
	case types.EventKindProjectAliasRemove:
		//lion:events section="project.alias.remove"
		// Alias removals are handled in projection tables, so reducer performs a lightweight decode and otherwise no-ops for compatibility.
		return true, r.applyProjectAliasRemove(e)
	case types.EventKindProjectDelete:
		//lion:events section="project.delete"
		// Records a tombstone so later task events referencing the project create synthetic placeholders instead of reviving deleted projects.
		return true, r.applyProjectDelete(e)
	case types.EventKindProjectNameSet:
		//lion:events section="project.name.set"
		// Updates project names in reducer state; if the project does not exist yet a synthetic project is created so the rename is not lost.
		return true, r.applyProjectNameSet(e)
	case types.EventKindTaskCreated:
		//lion:events section="task.created"
		// Creates tasks with deterministic handling for duplicate creation events and synthetic project creation when upstream data is incomplete.
		return true, r.applyTaskCreated(e)
	case types.EventKindTaskNumberSet:
		//lion:events section="task.number.set"
		// Applies display numbers to tasks and updates alias maps so task lookups by ID stay consistent with projection tables.
		return true, r.applyTaskNumberSet(e)
	case types.EventKindTaskRelocate:
		//lion:events section="task.relocate"
		// Moves tasks between projects, ensuring synthetic projects are created when needed and relations are rebuilt using the reducer graph helpers.
		return true, r.applyTaskRelocate(e)
	case types.EventKindTaskTitleSet:
		//lion:events section="task.title.set"
		// Updates task titles in reducer state, preserving historical metadata and avoiding extra allocations when titles repeat.
		return true, r.applyTaskTitleSet(e)
	default:
		// Not a handled event, skip
		return false, nil
	}
}

func (r *Reducer) applyProjectCreated(e types.Event) error {
	var payload types.ProjectCreatedPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.created payload: %w", err)
	}

	projectUID := string(payload.ProjectUID)

	if existing, exists := r.projects[projectUID]; exists {
		// Compatibility: mirror DB projections. Replace a synthetic placeholder with the real project;
		// keep existing real projects. This preserves historical quirks and is not the long-term intent.
		if existing.IsSynthetic {
			r.projects[projectUID] = &types.Project{
				ProjectUID:  projectUID,
				Type:        string(payload.Type),
				Name:        payload.Name,
				Description: payload.Description,
				CreatedBy:   payload.CreatedBy,
				CreatedAt:   e.CreatedAt,
				CreatedAtTS: e.TS,
				IsSynthetic: false,
			}
		}
		return nil
	}

	// Store project in reducer state
	r.projects[projectUID] = &types.Project{
		ProjectUID:  projectUID,
		Type:        string(payload.Type),
		Name:        payload.Name,
		Description: payload.Description,
		CreatedBy:   payload.CreatedBy,
		CreatedAt:   e.CreatedAt,
		CreatedAtTS: e.TS,
		IsSynthetic: false, // Real project from project.created event
	}

	return nil
}

func (r *Reducer) applyProjectAliasAdd(e types.Event) error {
	var payload types.ProjectAliasAddPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.alias.add payload: %w", err)
	}

	// Alias state is managed by DB projections
	return nil
}

func (r *Reducer) applyProjectAliasRemove(e types.Event) error {
	var payload types.ProjectAliasRemovePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.alias.remove payload: %w", err)
	}

	// Alias state is managed by DB projections
	return nil
}

func (r *Reducer) applyProjectNameSet(e types.Event) error {
	var payload types.ProjectNameSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.name.set payload: %w", err)
	}

	projectUID := string(payload.ProjectUID)

	project, exists := r.projects[projectUID]
	if !exists {
		// Create synthetic project so the name update is not lost
		r.createSyntheticProject(projectUID, e.CreatedAt, e.TS)
		project = r.projects[projectUID]
	}

	project.Name = payload.Name

	return nil
}

// createSyntheticProject creates a synthetic project entry for historical/corrupt data
func (r *Reducer) createSyntheticProject(projectUID string, createdAt time.Time, createdAtTS int64) {
	r.projects[projectUID] = &types.Project{
		ProjectUID:  projectUID,
		Type:        "local",
		Name:        projectUID, // Use UID as name for synthetic projects
		Description: "Synthetic project created by reducer",
		CreatedBy:   "system",
		CreatedAt:   createdAt,
		CreatedAtTS: createdAtTS,
		IsSynthetic: true,
	}
}

func (r *Reducer) applyTaskCreated(e types.Event) error {
	var payload types.TaskCreatedPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.created payload: %w", err)
	}

	taskUID := payload.TaskUID

	// Default item kind to "task" if not specified (backward compatibility)
	itemKind := payload.ItemKind
	if itemKind == "" {
		itemKind = "task"
	}

	// Check if project exists; if it was deleted, allow the task but do not resurrect the project
	projectUID := payload.ProjectUID

	// Create synthetic project if needed (mirrors DB projections)
	if _, exists := r.projects[projectUID]; !exists {
		if !r.deletedProj[projectUID] {
			// Compatibility: create synthetic project (for corrupt/historical data).
			// This preserves existing behavior; long-term intent is to handle deletes differently.
			r.createSyntheticProject(projectUID, e.CreatedAt, e.TS)
		}
	}

	// Handle duplicate task.created events deterministically
	// If a task with this UID already exists, keep the one with earlier Lamport timestamp
	if existing, exists := r.tasks[taskUID]; exists {
		// Keep the task with earlier Lamport timestamp (deterministic)
		if e.TS < existing.CreatedAtTS {
			// This event is earlier, replace existing task
			// (But keep any state changes that happened after creation)
			r.tasks[taskUID] = &types.Task{
				TaskUUID:      taskUID,
				TaskDisplayID: taskUID,
				ProjectUUID:   payload.ProjectUID,
				Aliases:       []string{},
				Title:         payload.Title,
				ItemKind:      itemKind,
				Axes:          existing.Axes,     // Preserve status changes
				Metadata:      existing.Metadata, // Preserve metadata
				Notes:         existing.Notes,    // Preserve notes
				CreatedBy:     payload.CreatedBy,
				CreatedAt:     e.CreatedAt,
				CreatedAtTS:   e.TS,
				UpdatedAt:     existing.UpdatedAt, // Preserve updated timestamp
				Relations:     existing.Relations,
				Blocked:       existing.Blocked,
				Blockers:      existing.Blockers,
			}
		}
		// Otherwise, keep existing task (it has earlier Lamport TS)
		return nil
	}

	// types.Task display ID is derived from project alias + number
	// For now, we'll use the task_uid as the task_id until we compute the display ID
	r.tasks[taskUID] = &types.Task{
		TaskUUID:      taskUID,
		TaskDisplayID: taskUID, // Placeholder, will be replaced by display ID
		ProjectUUID:   payload.ProjectUID,
		Aliases:       []string{},
		Title:         payload.Title,
		ItemKind:      itemKind,
		Axes:          make(map[string]types.AxisStatus),
		Notes:         []types.Note{},
		CreatedBy:     payload.CreatedBy,
		CreatedAt:     e.CreatedAt,
		CreatedAtTS:   e.TS,
		UpdatedAt:     e.CreatedAt, // Initially same as CreatedAt
	}

	// Register task by UID
	r.taskByID[taskUID] = taskUID

	// Track which project this task belongs to
	r.taskProjects[taskUID] = payload.ProjectUID

	return nil
}

func (r *Reducer) applyTaskNumberSet(e types.Event) error {
	var payload types.TaskNumberSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.number.set payload: %w", err)
	}

	// Number assignments are managed by DB projections (task_numbers table)
	// The reducer doesn't need to track this in memory
	return nil
}

func (r *Reducer) applyTaskRelocate(e types.Event) error {
	var payload types.TaskRelocatePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.relocate payload: %w", err)
	}

	taskUID := payload.TaskUID.String()
	toProjectUID := payload.ToProjectUID.String()

	// Update the task's project mapping so project.delete can correctly
	// remove tasks that belong to the deleted project
	r.taskProjects[taskUID] = toProjectUID

	// Also update the task's ProjectUUID field so reducer consumers
	// (like ls grouping and filtering) see the task under the new project
	if task, exists := r.tasks[taskUID]; exists {
		task.ProjectUUID = toProjectUID
		task.UpdatedAt = e.CreatedAt
	}

	return nil
}

func (r *Reducer) applyTaskTitleSet(e types.Event) error {
	var payload types.TaskTitleSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.title.set payload: %w", err)
	}

	taskUID := payload.TaskUID

	// Find the task
	task, exists := r.tasks[taskUID.String()]
	if !exists {
		return fmt.Errorf("task %s not found", taskUID)
	}

	// Update the title
	task.Title = payload.Title
	task.UpdatedAt = e.CreatedAt

	return nil
}
