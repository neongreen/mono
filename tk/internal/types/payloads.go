package types

import "encoding/json"

// TaskStatusSetPayload is the payload for task.status.set events
type TaskStatusSetPayload struct {
	TaskUUID string `json:"task_uuid,omitempty"` // New field for UUID-based updates
	TaskID   string `json:"task_id"`             // Legacy: only for reading old events. See tk-190 for removal plan.
	Axis     string `json:"axis"`                // e.g. "generic"
	State    string `json:"state"`               // e.g. "in_progress", "done", "blocked"
	Role     string `json:"role"`                // human / agent / bot / qa / rel
}

// TaskNoteAddPayload is the payload for task.note.add events
type TaskNoteAddPayload struct {
	TaskUUID string `json:"task_uuid,omitempty"` // New field for UUID-based updates
	TaskID   string `json:"task_id"`             // Legacy: only for reading old events. See tk-190 for removal plan.
	Markdown string `json:"markdown"`
}

// TaskDeletePayload is the payload for task.delete events
type TaskDeletePayload struct {
	TaskUUID string `json:"task_uuid"`
}

// TaskMetaSetPayload is the payload for task.meta.set events
type TaskMetaSetPayload struct {
	TaskUUID string          `json:"task_uuid"`
	TaskID   string          `json:"task_id"`
	Key      string          `json:"key"`
	Value    json.RawMessage `json:"value"` // JSON value (number, string, array, object, etc.)
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

// TaskAttachmentAddPayload is the payload for task.attachment.add events
type TaskAttachmentAddPayload struct {
	TaskUUID       string `json:"task_uuid"`
	AttachmentID   string `json:"attachment_id"`   // Attachment ID (e.g., "att-1")
	AttachmentHash string `json:"attachment_hash"` // SHA256 hash of content
	Filename       string `json:"filename"`
	Description    string `json:"description,omitempty"`
	MimeType       string `json:"mime_type,omitempty"`
	Size           int64  `json:"size"`
}

// TaskAttachmentRemovePayload is the payload for task.attachment.remove events
type TaskAttachmentRemovePayload struct {
	TaskUUID     string `json:"task_uuid"`
	AttachmentID string `json:"attachment_id"` // Attachment ID
}
