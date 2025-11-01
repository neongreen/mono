package types

// Relations represents all relations for a task
type Relations struct {
	Blocks     RelationSet `json:"blocks,omitempty"`     // Tasks this task blocks
	Subtask    RelationSet `json:"subtask,omitempty"`    // Parent/children for subtasks
	Related    RelationSet `json:"related,omitempty"`    // Related tasks
	Duplicate  RelationSet `json:"duplicate,omitempty"`  // Duplicate tasks
	Supersedes RelationSet `json:"supersedes,omitempty"` // Tasks this supersedes
}

// RelationSet represents directional relations
type RelationSet struct {
	Out      []RelationTarget `json:"out,omitempty"`      // Outgoing edges (this task -> others)
	In       []RelationTarget `json:"in,omitempty"`       // Incoming edges (others -> this task)
	Children []string         `json:"children,omitempty"` // For subtask relations
	Parent   string           `json:"parent,omitempty"`   // For subtask relations
}

// RelationTarget represents a relation target
type RelationTarget struct {
	TaskUUID string `json:"dst"` // Destination task UUID
	Note     string `json:"note,omitempty"`
}
