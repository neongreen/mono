package mcp

// Argument types for MCP tools

type CreateTaskArgs struct {
	Title       string            `json:"title" jsonschema:"required,description=Task title"`
	Project     string            `json:"project,omitempty" jsonschema:"description=Project name or alias (default: 'tk')"`
	Status      string            `json:"status,omitempty" jsonschema:"description=Initial status (e.g. 'todo', 'doing', 'done')"`
	Metadata    map[string]string `json:"metadata,omitempty" jsonschema:"description=Additional metadata key-value pairs"`
}

type ListTasksArgs struct {
	Project string `json:"project,omitempty" jsonschema:"description=Filter by project name or alias"`
	Status  string `json:"status,omitempty" jsonschema:"description=Filter by status (e.g. 'todo', 'doing', 'done')"`
	Blocked bool   `json:"blocked,omitempty" jsonschema:"description=Show only blocked tasks"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=Maximum number of tasks to return (default: 50)"`
}

type GetTaskArgs struct {
	TaskID string `json:"task_id" jsonschema:"required,description=Task ID (e.g. 'tk-123' or UUID)"`
}

type UpdateStatusArgs struct {
	TaskID string `json:"task_id" jsonschema:"required,description=Task ID"`
	Status string `json:"status" jsonschema:"required,description=New status (e.g. 'todo', 'doing', 'done')"`
	Axis   string `json:"axis,omitempty" jsonschema:"description=Status axis (default: 'generic')"`
	Role   string `json:"role,omitempty" jsonschema:"description=Role making the claim (default: 'human')"`
}

type AddNoteArgs struct {
	TaskID string `json:"task_id" jsonschema:"required,description=Task ID"`
	Note   string `json:"note" jsonschema:"required,description=Note text (markdown supported)"`
}

type RelateTasksArgs struct {
	SourceID     string `json:"source_id" jsonschema:"required,description=Source task ID"`
	TargetID     string `json:"target_id" jsonschema:"required,description=Target task ID"`
	RelationType string `json:"relation_type,omitempty" jsonschema:"description=Relation type: blocks, blocked_by, subtask, parent, related, duplicate_of, supersedes (default: subtask)"`
	Note         string `json:"note,omitempty" jsonschema:"description=Optional note explaining the relation"`
	// Deprecated fields for backward compatibility
	ParentID string `json:"parent_id,omitempty" jsonschema:"description=Deprecated: use source_id"`
	ChildID  string `json:"child_id,omitempty" jsonschema:"description=Deprecated: use target_id"`
}

type DeleteTaskArgs struct {
	TaskID string `json:"task_id" jsonschema:"required,description=Task ID (e.g. 'tk-123' or UUID)"`
}

type UpdateTitleArgs struct {
	TaskID string `json:"task_id" jsonschema:"required,description=Task ID"`
	Title  string `json:"title" jsonschema:"required,description=New task title"`
}

type SetMetadataArgs struct {
	TaskID string `json:"task_id" jsonschema:"required,description=Task ID"`
	Key    string `json:"key" jsonschema:"required,description=Metadata key"`
	Value  any    `json:"value" jsonschema:"required,description=Metadata value (can be any JSON type)"`
	Role   string `json:"role,omitempty" jsonschema:"description=Role making the claim (default: 'human')"`
}
