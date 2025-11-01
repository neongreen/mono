package payloads

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

// RelationAddPayload is the payload for relation.add events
type RelationAddPayload struct {
	Src  string `json:"src"`  // Source task UUID
	Type string `json:"type"` // blocks|blocked_by|subtask|parent|related|duplicate_of|supersedes
	Dst  string `json:"dst"`  // Destination task UUID
	Note string `json:"note,omitempty"`
}

// RelationRemovePayload is the payload for relation.remove events
type RelationRemovePayload struct {
	Src  string `json:"src"`  // Source task UUID
	Type string `json:"type"` // Relation type
	Dst  string `json:"dst"`  // Destination task UUID
}

// RelationNotePayload is the payload for relation.note events
type RelationNotePayload struct {
	Src      string `json:"src"`      // Source task UUID
	Type     string `json:"type"`     // Relation type
	Dst      string `json:"dst"`      // Destination task UUID
	Markdown string `json:"markdown"` // Note text
}
