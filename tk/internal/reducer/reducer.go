package reducer

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/neongreen/mono/tk/internal/relations"
	"github.com/neongreen/mono/tk/internal/sync"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
)

// Reducer reconstructs task state from events
type Reducer struct {
	tasks        map[string]*types.Task    // Key: task UUID
	taskByID     map[string]string         // Key: task ID (current or alias) -> Value: task UUID
	taskProjects map[string]string         // Key: task UUID -> Value: project UID
	relations    *relations.RelationsGraph // Relations graph
}

// NewReducer creates a new reducer
func NewReducer() *Reducer {
	return &Reducer{
		tasks:        make(map[string]*types.Task),
		taskByID:     make(map[string]string),
		taskProjects: make(map[string]string),
		relations:    relations.NewRelationsGraph(),
	}
}

// Relations returns the relations graph
func (r *Reducer) Relations() *relations.RelationsGraph {
	return r.relations
}

// Tasks returns the tasks map
func (r *Reducer) Tasks() map[string]*types.Task {
	return r.tasks
}

// Apply applies an event to update the task state
func (r *Reducer) Apply(e types.Event) error {
	// Try project events first
	handled, err := r.ApplyProjectEvent(e)
	if err != nil {
		return err
	}
	// If project handler processed this event, don't run legacy handlers
	if handled {
		return nil
	}

	// Handle shared events (status, notes, relations, delete)
	switch e.Kind {
	case "task.status.set":
		return r.applyTaskStatusSet(e)
	case "task.note.add":
		return r.applyTaskNoteAdd(e)
	case "task.delete":
		return r.applyTaskDelete(e)
	case "project.delete":
		return r.applyProjectDelete(e)
	case "relation.add":
		return r.applyRelationAdd(e)
	case "relation.remove":
		return r.applyRelationRemove(e)
	case "relation.note":
		return r.applyRelationNote(e)
	default:
		// Unknown events are ignored for forward compatibility
		return nil
	}
}

func (r *Reducer) applyTaskStatusSet(e types.Event) error {
	var payload types.TaskStatusSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.status.set payload: %w", err)
	}

	// Resolve task ID to UUID
	taskUUID := payload.TaskUUID
	if taskUUID == "" {
		// Legacy event - look up UUID by task ID
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

	return nil
}

func (r *Reducer) applyTaskNoteAdd(e types.Event) error {
	var payload types.TaskNoteAddPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.note.add payload: %w", err)
	}

	// Resolve task ID to UUID
	taskUUID := payload.TaskUUID
	if taskUUID == "" {
		// Legacy event - look up UUID by task ID
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

	note := types.Note{
		Markdown:  payload.Markdown,
		Actor:     e.Actor,
		Timestamp: e.CreatedAt, // Use actual creation time from event
	}
	task.Notes = append(task.Notes, note)

	return nil
}

func (r *Reducer) applyTaskDelete(e types.Event) error {
	var payload types.TaskDeletePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.delete payload: %w", err)
	}

	r.removeTaskFromMaps(payload.TaskUUID)
	return nil
}

// removeTaskFromMaps removes a task from all internal maps and relations
func (r *Reducer) removeTaskFromMaps(taskUUID string) {
	// Remove task from tasks map
	delete(r.tasks, taskUUID)

	// Remove all task ID mappings that point to this UUID
	for taskID, uuid := range r.taskByID {
		if uuid == taskUUID {
			delete(r.taskByID, taskID)
		}
	}

	// Remove project association
	delete(r.taskProjects, taskUUID)

	// Remove all relations involving this task
	r.relations.RemoveTaskRelations(taskUUID)
}

func (r *Reducer) applyProjectDelete(e types.Event) error {
	var payload types.ProjectDeletePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.delete payload: %w", err)
	}

	projectUID := payload.ProjectUID

	// Find all tasks in this project and delete them
	tasksToDelete := make([]string, 0)
	for taskUUID, taskProjectUID := range r.taskProjects {
		if taskProjectUID == projectUID {
			tasksToDelete = append(tasksToDelete, taskUUID)
		}
	}

	// Delete each task using the helper
	for _, taskUUID := range tasksToDelete {
		r.removeTaskFromMaps(taskUUID)
	}

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

// GetTask returns the current state of a task by ID (supports task ID or UUID)
func (r *Reducer) GetTask(idOrUUID string) (*types.Task, bool) {
	// Try direct UUID lookup first
	task, ok := r.tasks[idOrUUID]
	if ok {
		return task, true
	}

	// Try looking up by task ID
	taskUUID, ok := r.taskByID[idOrUUID]
	if !ok {
		return nil, false
	}

	task, ok = r.tasks[taskUUID]
	return task, ok
}

// GetAllTasks returns all tasks
func (r *Reducer) GetAllTasks() []*types.Task {
	tasks := make([]*types.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// applyRelationAdd adds a relation edge
func (r *Reducer) applyRelationAdd(e types.Event) error {
	var payload types.RelationAddPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal relation.add payload: %w", err)
	}

	// Extract node from event ID (format: ev-<number>-<node>)
	node := ""
	if len(e.ID) > 0 {
		parts := utils.SplitEventID(e.ID)
		if len(parts) == 3 {
			node = parts[2]
		}
	}

	r.relations.AddRelation(payload.Src, payload.Type, payload.Dst, payload.Note, e.ID, node, e.TS)
	return nil
}

// applyRelationRemove removes a relation edge
func (r *Reducer) applyRelationRemove(e types.Event) error {
	var payload types.RelationRemovePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal relation.remove payload: %w", err)
	}

	// Extract node from event ID
	node := ""
	if len(e.ID) > 0 {
		parts := utils.SplitEventID(e.ID)
		if len(parts) == 3 {
			node = parts[2]
		}
	}

	r.relations.RemoveRelation(payload.Src, payload.Type, payload.Dst, e.ID, node, e.TS)
	return nil
}

// applyRelationNote sets a note on a relation
func (r *Reducer) applyRelationNote(e types.Event) error {
	var payload types.RelationNotePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal relation.note payload: %w", err)
	}

	r.relations.SetRelationNote(payload.Src, payload.Type, payload.Dst, payload.Markdown)
	return nil
}

// FinalizeRelations builds relations for all tasks and computes blocked status
func (r *Reducer) FinalizeRelations(config *sync.Config) {
	// Build relations for all tasks
	for uuid, task := range r.tasks {
		task.Relations = r.relations.BuildTaskRelations(uuid)
	}

	// Compute blocked status
	utils.ComputeBlocked(r.relations, r.tasks, config.Blocking.BlockingAxis, config.Blocking.DoneStates)
}

// BuildFromEvents builds the current state from a list of events
func BuildFromEvents(events []types.Event) (*Reducer, error) {
	reducer := NewReducer()
	for _, e := range events {
		if err := reducer.Apply(e); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", e.ID, err)
		}
	}
	return reducer, nil
}

// BuildFromEventsWithConfig builds the current state from events and finalizes relations
func BuildFromEventsWithConfig(events []types.Event, config *sync.Config) (*Reducer, error) {
	reducer := NewReducer()
	for _, e := range events {
		if err := reducer.Apply(e); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", e.ID, err)
		}
	}
	reducer.FinalizeRelations(config)
	return reducer, nil
}

// V4 Event Reducer Functions
// Handles events (project.created, task.created, task.number.set, task.relocate, etc.)

// ApplyProjectEvent applies an event-specific handler
// Returns (handled=true, error) if the event was handled
// Returns (handled=false, nil) if the event was not handled
func (r *Reducer) ApplyProjectEvent(e types.Event) (bool, error) {
	switch types.EventKind(e.Kind) {
	case types.EventKindProjectCreated:
		return true, r.applyProjectCreated(e)
	case types.EventKindProjectAliasAdd:
		return true, r.applyProjectAliasAdd(e)
	case types.EventKindProjectAliasRemove:
		return true, r.applyProjectAliasRemove(e)
	case types.EventKindTaskCreated:
		return true, r.applyTaskCreated(e)
	case types.EventKindTaskNumberSet:
		return true, r.applyTaskNumberSet(e)
	case types.EventKindTaskRelocate:
		return true, r.applyTaskRelocate(e)
	case types.EventKindTaskTitleSet:
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

	// Project state is managed by DB projections, not in-memory reducer
	// This is just for completeness
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

func (r *Reducer) applyTaskCreated(e types.Event) error {
	var payload types.TaskCreatedPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.created payload: %w", err)
	}

	taskUID := payload.TaskUID

	// Guard against duplicate task.created for same UID
	if _, exists := r.tasks[taskUID]; exists {
		return nil
	}

	// types.Task display ID is derived from project alias + number
	// For now, we'll use the task_uid as the task_id until we compute the display ID
	r.tasks[taskUID] = &types.Task{
		TaskUUID:  taskUID,
		TaskID:    taskUID, // Placeholder, will be replaced by display ID
		Aliases:   []string{},
		Title:     payload.Title,
		Axes:      make(map[string]types.AxisStatus),
		Notes:     []types.Note{},
		CreatedBy: payload.CreatedBy,
		CreatedAt: e.CreatedAt,
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

	// Update the task's project mapping so project.delete can correctly
	// remove tasks that belong to the deleted project
	r.taskProjects[payload.TaskUID] = payload.ToProjectUID

	return nil
}

func (r *Reducer) applyTaskTitleSet(e types.Event) error {
	var payload types.TaskTitleSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.title.set payload: %w", err)
	}

	taskUID := payload.TaskUID

	// Find the task
	task, exists := r.tasks[taskUID]
	if !exists {
		return fmt.Errorf("task %s not found", taskUID)
	}

	// Update the title
	task.Title = payload.Title

	return nil
}
