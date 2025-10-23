package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Reducer reconstructs task state from events
type Reducer struct {
	tasks map[string]*Task
}

// NewReducer creates a new reducer
func NewReducer() *Reducer {
	return &Reducer{
		tasks: make(map[string]*Task),
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
	default:
		return fmt.Errorf("unknown event kind: %s", e.Kind)
	}
}

func (r *Reducer) applyTaskCreated(e Event) error {
	var payload TaskCreatedPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.created payload: %w", err)
	}

	r.tasks[payload.TaskID] = &Task{
		TaskID:    payload.TaskID,
		Title:     payload.Title,
		Axes:      make(map[string]AxisStatus),
		Notes:     []Note{},
		CreatedBy: payload.CreatedBy,
		CreatedAt: e.CreatedAt, // Use actual creation time from event
	}

	return nil
}

func (r *Reducer) applyTaskStatusSet(e Event) error {
	var payload TaskStatusSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.status.set payload: %w", err)
	}

	task, ok := r.tasks[payload.TaskID]
	if !ok {
		return fmt.Errorf("task not found: %s", payload.TaskID)
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

	task, ok := r.tasks[payload.TaskID]
	if !ok {
		return fmt.Errorf("task not found: %s", payload.TaskID)
	}

	note := Note{
		Markdown:  payload.Markdown,
		Actor:     e.Actor,
		Timestamp: e.CreatedAt, // Use actual creation time from event
	}
	task.Notes = append(task.Notes, note)

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

// GetTask returns the current state of a task
func (r *Reducer) GetTask(taskID string) (*Task, bool) {
	task, ok := r.tasks[taskID]
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
