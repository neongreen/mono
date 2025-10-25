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
	TaskUUID  string `json:"task_uuid"`  // Canonical immutable UUID
	TaskID    string `json:"task_id"`    // Display ID (prefix-number-node)
	Title     string `json:"title"`
	CreatedBy string `json:"created_by"`
}

// TaskStatusSetPayload is the payload for task.status.set events
type TaskStatusSetPayload struct {
	TaskUUID string `json:"task_uuid,omitempty"` // New field for UUID-based updates
	TaskID   string `json:"task_id"`             // Legacy field, still required for now
	Axis     string `json:"axis"`                // e.g. "generic"
	State    string `json:"state"`               // e.g. "in_progress", "done", "blocked"
	Role     string `json:"role"`                // human / agent / bot / qa / rel
}

// TaskNoteAddPayload is the payload for task.note.add events
type TaskNoteAddPayload struct {
	TaskUUID string `json:"task_uuid,omitempty"` // New field for UUID-based updates
	TaskID   string `json:"task_id"`             // Legacy field, still required for now
	Markdown string `json:"markdown"`
}

// PrefixCreatedPayload is the payload for prefix.created events
type PrefixCreatedPayload struct {
	Prefix      string `json:"prefix"`
	Description string `json:"description"`
	CreatedBy   string `json:"created_by"`
}

// PrefixDescriptionSetPayload is the payload for prefix.description.set events
type PrefixDescriptionSetPayload struct {
	Prefix      string `json:"prefix"`
	Description string `json:"description"`
}

// PrefixAliasAddedPayload is the payload for prefix.alias.added events
type PrefixAliasAddedPayload struct {
	Prefix string `json:"prefix"`
	Alias  string `json:"alias"`
}

// PrefixRemovedPayload is the payload for prefix.removed events
type PrefixRemovedPayload struct {
	Prefix string `json:"prefix"`
}

// TaskReprefixPayload is the payload for task.reprefix events
type TaskReprefixPayload struct {
	TaskUUID   string `json:"task_uuid"`
	OldPrefix  string `json:"old_prefix"`
	NewPrefix  string `json:"new_prefix"`
	OldNumber  int64  `json:"old_number"`
	NewNumber  int64  `json:"new_number"`
	OldNode    string `json:"old_node"`
	Reason     string `json:"reason,omitempty"`
}

// TaskAliasAddedPayload is the payload for task.alias.added events
type TaskAliasAddedPayload struct {
	TaskUUID string `json:"task_uuid"`
	AliasID  string `json:"alias_id"`
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
	TaskUUID  string                `json:"task_uuid"`            // Canonical immutable UUID
	TaskID    string                `json:"task_id"`              // Current display ID
	Aliases   []string              `json:"aliases,omitempty"`    // Previous IDs (when task was moved)
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
