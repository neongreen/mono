package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Reducer reconstructs task state from events
type Reducer struct {
	tasks    map[string]*Task  // Key: task UUID
	taskByID map[string]string // Key: task ID (current or alias) -> Value: task UUID
}

// NewReducer creates a new reducer
func NewReducer() *Reducer {
	return &Reducer{
		tasks:    make(map[string]*Task),
		taskByID: make(map[string]string),
	}
}

// Apply applies an event to update the task state
func (r *Reducer) Apply(e Event) error {
	switch e.Kind {
	case "task.created":
		return r.applyTaskCreated(e)
	case "task.status.set":
		return r.applyTaskStatusSet(e)
	case "task.note.add":
		return r.applyTaskNoteAdd(e)
	case "task.reprefix":
		return r.applyTaskReprefix(e)
	case "task.alias.added":
		return r.applyTaskAliasAdded(e)
	case "prefix.created":
		// Prefix events are handled by the DB projector, not the reducer
		return nil
	case "prefix.description.set":
		// Prefix events are handled by the DB projector, not the reducer
		return nil
	case "prefix.alias.added":
		// Prefix events are handled by the DB projector, not the reducer
		return nil
	case "prefix.removed":
		// Prefix events are handled by the DB projector, not the reducer
		return nil
	default:
		// Unknown events are ignored for forward compatibility
		return nil
	}
}

func (r *Reducer) applyTaskCreated(e Event) error {
	var payload TaskCreatedPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.created payload: %w", err)
	}

	// Support both old events (without task_uuid) and new events (with task_uuid)
	taskUUID := payload.TaskUUID
	if taskUUID == "" {
		// Legacy event without UUID - use task_id as UUID
		taskUUID = payload.TaskID
	}

	r.tasks[taskUUID] = &Task{
		TaskUUID:  taskUUID,
		TaskID:    payload.TaskID,
		Aliases:   []string{},
		Title:     payload.Title,
		Axes:      make(map[string]AxisStatus),
		Notes:     []Note{},
		CreatedBy: payload.CreatedBy,
		CreatedAt: e.CreatedAt, // Use actual creation time from event
	}

	// Register the task ID in the lookup map
	r.taskByID[payload.TaskID] = taskUUID

	return nil
}

func (r *Reducer) applyTaskStatusSet(e Event) error {
	var payload TaskStatusSetPayload
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
		axis = AxisStatus{
			Claims: []Claim{},
		}
	}

	// Add new claim
	newClaim := Claim{
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

func (r *Reducer) applyTaskNoteAdd(e Event) error {
	var payload TaskNoteAddPayload
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

	note := Note{
		Markdown:  payload.Markdown,
		Actor:     e.Actor,
		Timestamp: e.CreatedAt, // Use actual creation time from event
	}
	task.Notes = append(task.Notes, note)

	return nil
}

func (r *Reducer) applyTaskReprefix(e Event) error {
	var payload TaskReprefixPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.reprefix payload: %w", err)
	}

	task, ok := r.tasks[payload.TaskUUID]
	if !ok {
		return fmt.Errorf("task UUID not found: %s", payload.TaskUUID)
	}

	// Construct old and new task IDs
	oldTaskID := fmt.Sprintf("%s-%d-%s", payload.OldPrefix, payload.OldNumber, payload.OldNode)
	newTaskID := fmt.Sprintf("%s-%d-%s", payload.NewPrefix, payload.NewNumber, payload.OldNode)

	// Update the task's current ID
	task.TaskID = newTaskID

	// Update the lookup map: remove old ID, add new ID
	delete(r.taskByID, oldTaskID)
	r.taskByID[newTaskID] = payload.TaskUUID

	return nil
}

func (r *Reducer) applyTaskAliasAdded(e Event) error {
	var payload TaskAliasAddedPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.alias.added payload: %w", err)
	}

	task, ok := r.tasks[payload.TaskUUID]
	if !ok {
		return fmt.Errorf("task UUID not found: %s", payload.TaskUUID)
	}

	// Add alias to the task
	task.Aliases = append(task.Aliases, payload.AliasID)

	// Register the alias in the lookup map
	r.taskByID[payload.AliasID] = payload.TaskUUID

	return nil
}

// resolveEffectiveStatus resolves which claim is effective based on authority
func (r *Reducer) resolveEffectiveStatus(axis *AxisStatus) {
	if len(axis.Claims) == 0 {
		return
	}

	// Group claims by state to find concurrent claims
	stateGroups := make(map[string][]Claim)
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
	var concurrentClaims []Claim
	for i, claim := range axis.Claims {
		if claim.TS == latestTS {
			concurrentClaims = append(concurrentClaims, axis.Claims[i])
		}
	}

	// Sort by authority (highest first)
	sort.Slice(concurrentClaims, func(i, j int) bool {
		return GetRoleAuthority(concurrentClaims[i].Role) > GetRoleAuthority(concurrentClaims[j].Role)
	})

	// The highest authority claim is effective
	effectiveClaim := concurrentClaims[0]
	axis.Effective = effectiveClaim.State

	// Mark all claims as tentative or not
	highestAuthority := GetRoleAuthority(effectiveClaim.Role)
	for i := range axis.Claims {
		claimAuthority := GetRoleAuthority(axis.Claims[i].Role)
		// A claim is tentative if it has lower authority than the effective claim
		// OR if it's not the latest claim with that state
		axis.Claims[i].Tentative = claimAuthority < highestAuthority ||
			axis.Claims[i].TS < latestTS
	}
}

// GetTask returns the current state of a task by ID (supports task ID or UUID)
func (r *Reducer) GetTask(idOrUUID string) (*Task, bool) {
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
func (r *Reducer) GetAllTasks() []*Task {
	tasks := make([]*Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// BuildFromEvents builds the current state from a list of events
func BuildFromEvents(events []Event) (*Reducer, error) {
	reducer := NewReducer()
	for _, e := range events {
		if err := reducer.Apply(e); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", e.ID, err)
		}
	}
	return reducer, nil
}
