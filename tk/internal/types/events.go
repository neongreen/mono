package types

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// types.Event Payload Definitions
// Based on tk/specs/v4.md

// projectNamePattern matches valid project names:
// - lowercase letters (a-z) only
// - single dashes allowed between letter groups
// - no leading/trailing dashes
// - no consecutive dashes
var projectNamePattern = regexp.MustCompile(`^[a-z]+(-[a-z]+)*$`)

// ValidateProjectName validates a project name according to the rules:
// - Must be at least 1 character
// - Lowercase letters (a-z) and dashes (-) only
// - No leading or trailing dashes
// - No consecutive dashes
func ValidateProjectName(name string) error {
	if len(name) < 1 {
		return fmt.Errorf("project name cannot be empty")
	}

	if !projectNamePattern.MatchString(name) {
		// Generate a helpful suggestion
		suggestion := strings.ToLower(name)
		// Remove invalid characters
		suggestion = regexp.MustCompile(`[^a-z-]+`).ReplaceAllString(suggestion, "")
		// Replace consecutive dashes with single dash
		suggestion = regexp.MustCompile(`-+`).ReplaceAllString(suggestion, "-")
		// Remove leading/trailing dashes
		suggestion = strings.Trim(suggestion, "-")

		if suggestion != "" && suggestion != name {
			return fmt.Errorf("invalid project name '%s': must be lowercase letters and dashes only, no leading/trailing dashes or consecutive dashes. Try: '%s'", name, suggestion)
		}
		return fmt.Errorf("invalid project name '%s': must be lowercase letters and dashes only, no leading/trailing dashes or consecutive dashes", name)
	}

	return nil
}

// ProjectCreatedPayload is the payload for project.created events
type ProjectCreatedPayload struct {
	ProjectUID  string `json:"project_uid"`
	Type        string `json:"type"`        // local, github, linear, jira
	Name        string `json:"name"`        // human-readable name
	Description string `json:"description"` // text description
	CreatedBy   string `json:"created_by"`  // actor who created it
}

// ProjectAliasAddPayload is the payload for project.alias.add events
// deprecated:v5 track:true reason:"Aliases removed in favor of project names"
type ProjectAliasAddPayload struct {
	ProjectUID string `json:"project_uid"`
	Alias      string `json:"alias"` // deprecated:v5
	Node       string `json:"node"`  // deprecated:v5
	AddedBy    string `json:"added_by"`
}

// ProjectAliasRemovePayload is the payload for project.alias.remove events
// deprecated:v5 track:true reason:"Aliases removed in favor of project names"
type ProjectAliasRemovePayload struct {
	ProjectUID string `json:"project_uid"`
	Alias      string `json:"alias"` // deprecated:v5
	Node       string `json:"node"`  // deprecated:v5
}

// ProjectDeletePayload is the payload for project.delete events
type ProjectDeletePayload struct {
	ProjectUID ProjectUID `json:"project_uid"` // Must be valid ProjectUID (prj_<ulid>)
}

// Validate checks if the ProjectDeletePayload is valid
func (p ProjectDeletePayload) Validate() error {
	if err := p.ProjectUID.Validate(); err != nil {
		return fmt.Errorf("invalid project_uid in ProjectDeletePayload: %w", err)
	}
	return nil
}

// UnmarshalJSON unmarshals and validates a ProjectDeletePayload
func (p *ProjectDeletePayload) UnmarshalJSON(data []byte) error {
	// Use a type alias to avoid infinite recursion
	type Alias ProjectDeletePayload
	aux := &struct{ *Alias }{Alias: (*Alias)(p)}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal ProjectDeletePayload: %w", err)
	}
	// Validate after unmarshaling
	if err := p.Validate(); err != nil {
		return err
	}
	return nil
}

// ProjectNameSetPayload is the payload for project.name.set events
type ProjectNameSetPayload struct {
	ProjectUID ProjectUID `json:"project_uid"` // Must be valid ProjectUID (prj_<ulid>)
	Name       string     `json:"name"`        // New project name (lowercase, dashes, validated)
}

// Validate checks if the ProjectNameSetPayload is valid
func (p ProjectNameSetPayload) Validate() error {
	if err := p.ProjectUID.Validate(); err != nil {
		return fmt.Errorf("invalid project_uid in ProjectNameSetPayload: %w", err)
	}
	if err := ValidateProjectName(p.Name); err != nil {
		return fmt.Errorf("invalid name in ProjectNameSetPayload: %w", err)
	}
	return nil
}

// UnmarshalJSON unmarshals and validates a ProjectNameSetPayload
func (p *ProjectNameSetPayload) UnmarshalJSON(data []byte) error {
	// Use a type alias to avoid infinite recursion
	type Alias ProjectNameSetPayload
	aux := &struct{ *Alias }{Alias: (*Alias)(p)}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal ProjectNameSetPayload: %w", err)
	}
	// Validate after unmarshaling
	if err := p.Validate(); err != nil {
		return err
	}
	return nil
}

// TaskCreatedPayload is the payload for task.created events
type TaskCreatedPayload struct {
	TaskUID        string `json:"task_uid"`
	ProjectUID     string `json:"project_uid"`
	ProposedNumber int64  `json:"proposed_number,omitempty"` // best-effort number
	CreatedNode    string `json:"created_node"`
	Title          string `json:"title"`
	CreatedBy      string `json:"created_by"`
	ItemKind       string `json:"item_kind,omitempty"` // v7+ item kind (task, decision, resource, etc.)
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

// Container event payloads (v6 event-defined capabilities)
// See tk/specs/v6-event-defined-capabilities.md

// ContainerPrimitive represents the built-in container types
type ContainerPrimitive string

const (
	PrimitiveQueue ContainerPrimitive = "queue"
	PrimitiveStack ContainerPrimitive = "stack"
	PrimitiveGroup ContainerPrimitive = "group"
)

// Item kind event payloads (v7 item kinds feature)

// DefineItemKindPayload is the payload for item_kind.define events
type DefineItemKindPayload struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	LLMHint     string `json:"llm_hint,omitempty"`
	CreatedBy   string `json:"created_by"`
}

// DeprecateItemKindPayload is the payload for item_kind.deprecate events
type DeprecateItemKindPayload struct {
	Name string `json:"name"`
}

// DefineContainerKindPayload is the payload for container.kind.define events
type DefineContainerKindPayload struct {
	Name        string             `json:"name"`
	Primitive   ContainerPrimitive `json:"primitive"`
	Description string             `json:"description,omitempty"`
	CreatedBy   string             `json:"created_by"`
}

// DeprecateContainerKindPayload is the payload for container.kind.deprecate events
type DeprecateContainerKindPayload struct {
	Name string `json:"name"`
}

// CreateContainerPayload is the payload for container.create events
type CreateContainerPayload struct {
	ID        string             `json:"id"`
	Primitive ContainerPrimitive `json:"primitive"`
	Kind      string             `json:"kind"`
	Name      string             `json:"name"`
	Metadata  map[string]any     `json:"metadata,omitempty"`
	CreatedBy string             `json:"created_by"`
}

// RenameContainerPayload is the payload for container.rename events
type RenameContainerPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UpdateContainerMetadataPayload is the payload for container.metadata.update events
type UpdateContainerMetadataPayload struct {
	ID       string         `json:"id"`
	Metadata map[string]any `json:"metadata"`
}

// RemoveContainerPayload is the payload for container.remove events
type RemoveContainerPayload struct {
	ID string `json:"id"`
}

// QueuePushPayload is the payload for queue.push events
type QueuePushPayload struct {
	ContainerID string  `json:"container_id"`
	ItemID      TaskUID `json:"item_id"`
}

// QueuePopPayload is the payload for queue.pop events
type QueuePopPayload struct {
	ContainerID string  `json:"container_id"`
	ItemID      TaskUID `json:"item_id"`
}

// StackPushPayload is the payload for stack.push events
type StackPushPayload struct {
	ContainerID string  `json:"container_id"`
	ItemID      TaskUID `json:"item_id"`
}

// StackPopPayload is the payload for stack.pop events
type StackPopPayload struct {
	ContainerID string  `json:"container_id"`
	ItemID      TaskUID `json:"item_id"`
}

// GroupAddPayload is the payload for group.add events
type GroupAddPayload struct {
	ContainerID string  `json:"container_id"`
	ItemID      TaskUID `json:"item_id"`
}

// GroupRemovePayload is the payload for group.remove events
type GroupRemovePayload struct {
	ContainerID string  `json:"container_id"`
	ItemID      TaskUID `json:"item_id"`
}
