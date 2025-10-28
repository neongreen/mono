package main

// V4 Event Payload Definitions
// Based on tk/specs/v4.md

// ProjectCreatedPayload is the payload for project.created events
type ProjectCreatedPayload struct {
	ProjectUID  string `json:"project_uid"`
	Type        string `json:"type"`        // local, github, linear, jira
	Name        string `json:"name"`        // human-readable name
	Description string `json:"description"` // text description
	CreatedBy   string `json:"created_by"`  // actor who created it
}

// ProjectAliasAddPayload is the payload for project.alias.add events
type ProjectAliasAddPayload struct {
	ProjectUID string `json:"project_uid"`
	Alias      string `json:"alias"`
	Node       string `json:"node"`
	AddedBy    string `json:"added_by"`
}

// ProjectAliasRemovePayload is the payload for project.alias.remove events
type ProjectAliasRemovePayload struct {
	ProjectUID string `json:"project_uid"`
	Alias      string `json:"alias"`
	Node       string `json:"node"`
}

// TaskCreatedV4Payload is the payload for task.created events (v4)
type TaskCreatedV4Payload struct {
	TaskUID        string `json:"task_uid"`
	ProjectUID     string `json:"project_uid"`
	ProposedNumber int64  `json:"proposed_number,omitempty"` // best-effort number
	CreatedNode    string `json:"created_node"`
	Title          string `json:"title"`
	CreatedBy      string `json:"created_by"`
}

// TaskNumberSetPayload is the payload for task.number.set events
type TaskNumberSetPayload struct {
	TaskUID    string `json:"task_uid"`
	ProjectUID string `json:"project_uid"`
	Number     int64  `json:"number"`
	Reason     string `json:"reason,omitempty"` // e.g., "collision resolved", "manual renumber"
}

// TaskRelocatePayload is the payload for task.relocate events
type TaskRelocatePayload struct {
	TaskUID        string              `json:"task_uid"`
	FromProjectUID string              `json:"from_project_uid"`
	ToProjectUID   string              `json:"to_project_uid"`
	NumberPolicy   NumberPolicyPayload `json:"number_policy"`
}

// NumberPolicyPayload is embedded in TaskRelocatePayload
type NumberPolicyPayload struct {
	Mode   string `json:"mode"`             // keep, auto, force, fail
	Number int64  `json:"number,omitempty"` // used when mode == "force"
}

// TaskTitleSetPayload is the payload for task.title.set events
type TaskTitleSetPayload struct {
	TaskUID string `json:"task_uid"`
	Title   string `json:"title"`
}
