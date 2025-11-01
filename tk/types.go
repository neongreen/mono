package main

import (
	"encoding/json"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
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

// TaskStatusSetPayload is the payload for task.status.set events

// New field for UUID-based updates
// Legacy field, still required for now
// e.g. "generic"
// e.g. "in_progress", "done", "blocked"
// human / agent / bot / qa / rel

// TaskNoteAddPayload is the payload for task.note.add events

// New field for UUID-based updates
// Legacy field, still required for now

// RelationAddPayload is the payload for relation.add events

// Source task UUID
// blocks|blocked_by|subtask|parent|related|duplicate_of|supersedes
// Destination task UUID

// RelationRemovePayload is the payload for relation.remove events

// Source task UUID
// Relation type
// Destination task UUID

// RelationNotePayload is the payload for relation.note events

// Source task UUID
// Relation type
// Destination task UUID
// Note text

// Claim represents a status assertion by an actor

// AxisStatus represents the status claims for a single axis

// Task represents the current state of a task, derived from events
type Task struct {
	TaskUUID  string                      `json:"task_uuid"`         // Canonical immutable UUID
	TaskID    string                      `json:"task_id"`           // Current display ID
	Aliases   []string                    `json:"aliases,omitempty"` // Previous IDs (when task was moved)
	Title     string                      `json:"title"`
	Axes      map[string]types.AxisStatus `json:"axes"`
	Notes     []types.Note                `json:"notes"`
	CreatedBy string                      `json:"created_by"`
	CreatedAt time.Time                   `json:"created_at"`
	Relations *types.Relations            `json:"relations,omitempty"` // Task relations
	Blocked   bool                        `json:"blocked,omitempty"`   // Is this task blocked
	Blockers  []types.Blocker             `json:"blockers,omitempty"`  // List of blocking tasks
}

// Blocker represents a task that blocks another

// Distance in dependency graph

// Note represents a note on a task

// Role authority levels (higher is more authoritative)
var roleAuthority = map[string]int{
	"human": 5,
	"qa":    4,
	"rel":   3,
	"agent": 2,
	"bot":   1,
}
