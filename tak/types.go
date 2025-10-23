package main

import (
	"encoding/json"
	"time"
)

// Event represents an immutable event in the event log
type Event struct {
	ID        string          `json:"id"`         // Event ID
	TS        int64           `json:"ts"`         // Lamport timestamp
	CreatedAt time.Time       `json:"created_at"` // Actual creation time
	Actor     string          `json:"actor"`      // Username
	Role      string          `json:"role"`       // human / agent / bot / qa / rel
	Kind      string          `json:"kind"`       // e.g. task.created, status.set, note.add
	Payload   json.RawMessage `json:"payload"`    // event-specific data
	Ctx       json.RawMessage `json:"ctx"`        // contextual info (repo, branch, commit)
	RepoUUID  string          `json:"repo_uuid"`  // optional
	Branch    string          `json:"branch"`     // optional
	Commit    string          `json:"commit"`     // optional
	JJOpID    string          `json:"jj_op_id"`   // optional
}

// TaskCreatedPayload is the payload for task.created events
type TaskCreatedPayload struct {
	TaskID    string `json:"task_id"`
	Title     string `json:"title"`
	CreatedBy string `json:"created_by"`
}

// TaskStatusSetPayload is the payload for task.status.set events
type TaskStatusSetPayload struct {
	TaskID string `json:"task_id"`
	Axis   string `json:"axis"`  // e.g. "generic"
	State  string `json:"state"` // e.g. "in_progress", "done", "blocked"
	Role   string `json:"role"`  // human / agent / bot / qa / rel
}

// TaskNoteAddPayload is the payload for task.note.add events
type TaskNoteAddPayload struct {
	TaskID   string `json:"task_id"`
	Markdown string `json:"markdown"`
}

// Claim represents a status assertion by an actor
type Claim struct {
	State     string `json:"state"`
	Role      string `json:"role"`
	Tentative bool   `json:"tentative"`
	TS        int64  `json:"ts"`
}

// AxisStatus represents the status claims for a single axis
type AxisStatus struct {
	Effective string  `json:"effective"`
	Claims    []Claim `json:"claims"`
}

// Task represents the current state of a task, derived from events
type Task struct {
	TaskID    string                `json:"task_id"`
	Title     string                `json:"title"`
	Axes      map[string]AxisStatus `json:"axes"`
	Notes     []Note                `json:"notes"`
	CreatedBy string                `json:"created_by"`
	CreatedAt time.Time             `json:"created_at"`
}

// Note represents a note on a task
type Note struct {
	Markdown  string    `json:"markdown"`
	Actor     string    `json:"actor"`
	Timestamp time.Time `json:"timestamp"`
}

// Role authority levels (higher is more authoritative)
var roleAuthority = map[string]int{
	"human": 5,
	"qa":    4,
	"rel":   3,
	"agent": 2,
	"bot":   1,
}

// GetRoleAuthority returns the authority level for a role
func GetRoleAuthority(role string) int {
	if auth, ok := roleAuthority[role]; ok {
		return auth
	}
	return 0
}
