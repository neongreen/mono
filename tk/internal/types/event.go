package types

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
