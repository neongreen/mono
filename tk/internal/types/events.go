package types

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// types.Event Payload Definitions
// Based on tk/specs/v4.md

// ============================================================================
// Event Payload Type Pattern: Lax vs Strict
// ============================================================================
//
// PROBLEM:
// Historical tk databases contain events with malformed identifiers:
// - Project UIDs like "lovable" (name/alias) instead of "prj_<ulid>"
// - Task IDs like "123" (old format) instead of "tsk_<ulid>"
//
// SOLUTION: Lax/Strict Pattern
// -----------------------------
// For events where validation has been tightened, we use two types:
//
// 1. *PayloadLax (e.g., TaskNumberSetPayloadLax)
//    - Used when READING events from the log
//    - Fields are permissive (plain strings, etc.)
//    - Accepts legacy/malformed data
//    - Superset of what strict type accepts
//
// 2. *Payload (e.g., TaskNumberSetPayload)
//    - Used when WRITING/EMITTING new events
//    - Fields are strict (typed UIDs, validation)
//    - Only accepts properly formatted data
//    - Subset of what lax type accepts
//
// INVARIANTS:
// -----------
// 1. New code NEVER emits lax events - always use strict types
// 2. Lax types ONLY for reading historical events
// 3. Lax type MUST be superset of strict type
// 4. After resolution, TaskUID/ProjectUID are always valid ULID format
//
// WORKFLOW:
// ---------
// Reading events (reducer):
//   Event JSON → Unmarshal to Lax → Resolve identifiers → Use validated strings
//
// Creating events (tk new, tk mv):
//   Build Strict payload → Marshal to JSON → Insert into events table
//
// ============================================================================

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
	ProjectUID  ProjectUID  `json:"project_uid"` // Must be valid ProjectUID (prj_<ulid>)
	Type        ProjectType `json:"type"`        // Project type (local, github, linear, jira)
	Name        string      `json:"name"`        // Project name (lowercase, dashes, validated)
	Description string      `json:"description"` // Text description (any string)
	CreatedBy   string      `json:"created_by"`  // Actor who created it (any string)
}

// Validate checks if the ProjectCreatedPayload is valid
func (p ProjectCreatedPayload) Validate() error {
	if err := p.ProjectUID.Validate(); err != nil {
		return fmt.Errorf("invalid project_uid in ProjectCreatedPayload: %w", err)
	}
	if err := p.Type.Validate(); err != nil {
		return fmt.Errorf("invalid type in ProjectCreatedPayload: %w", err)
	}
	if err := ValidateProjectName(p.Name); err != nil {
		return fmt.Errorf("invalid name in ProjectCreatedPayload: %w", err)
	}
	// Description and CreatedBy can be any string
	return nil
}

// UnmarshalJSON unmarshals and validates a ProjectCreatedPayload
func (p *ProjectCreatedPayload) UnmarshalJSON(data []byte) error {
	// Use a type alias to avoid infinite recursion
	type Alias ProjectCreatedPayload
	aux := &struct{ *Alias }{Alias: (*Alias)(p)}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal ProjectCreatedPayload: %w", err)
	}
	// Validate after unmarshaling
	if err := p.Validate(); err != nil {
		return err
	}
	return nil
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

// ========== TaskNumberSetPayload ==========

// TaskNumberSetPayloadLax is used when reading events from the log.
// Accepts legacy/malformed project UIDs and task IDs for backward compatibility.
type TaskNumberSetPayloadLax struct {
	TaskUID    string `json:"task_uid"`    // Could be: "tsk_<ulid>" OR legacy "123"
	ProjectUID string `json:"project_uid"` // Could be: "prj_<ulid>" OR legacy "lovable"
	Number     int64  `json:"number"`
	Reason     string `json:"reason,omitempty"`
}

// TaskNumberSetPayload is used when creating new events.
// Enforces strict validation - only proper ULID-based identifiers allowed.
type TaskNumberSetPayload struct {
	TaskUID    TaskUID    `json:"task_uid"`         // Must be valid TaskUID (tsk_<ulid>)
	ProjectUID ProjectUID `json:"project_uid"`      // Must be valid ProjectUID (prj_<ulid>)
	Number     int64      `json:"number"`           // Task number (must be positive)
	Reason     string     `json:"reason,omitempty"` // e.g., "collision resolved", "manual renumber"
}

// Validate checks if the TaskNumberSetPayload is valid
func (p TaskNumberSetPayload) Validate() error {
	if err := p.TaskUID.Validate(); err != nil {
		return fmt.Errorf("invalid task_uid in TaskNumberSetPayload: %w", err)
	}
	if err := p.ProjectUID.Validate(); err != nil {
		return fmt.Errorf("invalid project_uid in TaskNumberSetPayload: %w", err)
	}
	if p.Number <= 0 {
		return fmt.Errorf("invalid number in TaskNumberSetPayload: must be positive, got %d", p.Number)
	}
	return nil
}

// UnmarshalJSON unmarshals and validates a TaskNumberSetPayload
func (p *TaskNumberSetPayload) UnmarshalJSON(data []byte) error {
	// Use a type alias to avoid infinite recursion
	type Alias TaskNumberSetPayload
	aux := &struct{ *Alias }{Alias: (*Alias)(p)}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal TaskNumberSetPayload: %w", err)
	}
	// Validate after unmarshaling
	if err := p.Validate(); err != nil {
		return err
	}
	return nil
}

// ========== TaskRelocatePayload ==========

// TaskRelocatePayloadLax is used when reading events from the log.
type TaskRelocatePayloadLax struct {
	TaskUID        string              `json:"task_uid"`
	FromProjectUID string              `json:"from_project_uid"`
	ToProjectUID   string              `json:"to_project_uid"`
	NumberPolicy   NumberPolicyPayload `json:"number_policy"`
}

// TaskRelocatePayload is used when creating new events.
type TaskRelocatePayload struct {
	TaskUID        TaskUID             `json:"task_uid"`         // Must be valid TaskUID (tsk_<ulid>)
	FromProjectUID ProjectUID          `json:"from_project_uid"` // Must be valid ProjectUID (prj_<ulid>)
	ToProjectUID   ProjectUID          `json:"to_project_uid"`   // Must be valid ProjectUID (prj_<ulid>)
	NumberPolicy   NumberPolicyPayload `json:"number_policy"`    // Number assignment policy
}

// Validate checks if the TaskRelocatePayload is valid
func (p TaskRelocatePayload) Validate() error {
	if err := p.TaskUID.Validate(); err != nil {
		return fmt.Errorf("invalid task_uid in TaskRelocatePayload: %w", err)
	}
	if err := p.FromProjectUID.Validate(); err != nil {
		return fmt.Errorf("invalid from_project_uid in TaskRelocatePayload: %w", err)
	}
	if err := p.ToProjectUID.Validate(); err != nil {
		return fmt.Errorf("invalid to_project_uid in TaskRelocatePayload: %w", err)
	}
	if err := p.NumberPolicy.Validate(); err != nil {
		return fmt.Errorf("invalid number_policy in TaskRelocatePayload: %w", err)
	}
	return nil
}

// UnmarshalJSON unmarshals and validates a TaskRelocatePayload
func (p *TaskRelocatePayload) UnmarshalJSON(data []byte) error {
	// Use a type alias to avoid infinite recursion
	type Alias TaskRelocatePayload
	aux := &struct{ *Alias }{Alias: (*Alias)(p)}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal TaskRelocatePayload: %w", err)
	}
	// Validate after unmarshaling
	if err := p.Validate(); err != nil {
		return err
	}
	return nil
}

// NumberPolicyPayload is embedded in TaskRelocatePayload
type NumberPolicyPayload struct {
	Mode   string `json:"mode"`             // keep, auto, force, fail
	Number int64  `json:"number,omitempty"` // used when mode == "force" (must be positive)
}

// Validate checks if the NumberPolicyPayload is valid
func (p NumberPolicyPayload) Validate() error {
	switch p.Mode {
	case "keep", "auto", "force", "fail":
		if p.Mode == "force" && p.Number <= 0 {
			return fmt.Errorf("invalid number in NumberPolicyPayload: when mode is 'force', number must be positive, got %d", p.Number)
		}
		return nil
	default:
		return fmt.Errorf("invalid mode in NumberPolicyPayload: must be one of [keep, auto, force, fail], got %q", p.Mode)
	}
}

// TaskTitleSetPayload is the payload for task.title.set events
type TaskTitleSetPayload struct {
	TaskUID TaskUID `json:"task_uid"` // Must be valid TaskUID (tsk_<ulid>)
	Title   string  `json:"title"`    // Task title (non-empty, trimmed)
}

// Validate checks if the TaskTitleSetPayload is valid
func (p TaskTitleSetPayload) Validate() error {
	if err := p.TaskUID.Validate(); err != nil {
		return fmt.Errorf("invalid task_uid in TaskTitleSetPayload: %w", err)
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return fmt.Errorf("invalid title in TaskTitleSetPayload: title cannot be empty")
	}
	return nil
}

// UnmarshalJSON unmarshals and validates a TaskTitleSetPayload
func (p *TaskTitleSetPayload) UnmarshalJSON(data []byte) error {
	// Use a type alias to avoid infinite recursion
	type Alias TaskTitleSetPayload
	aux := &struct{ *Alias }{Alias: (*Alias)(p)}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal TaskTitleSetPayload: %w", err)
	}
	// Validate after unmarshaling
	if err := p.Validate(); err != nil {
		return err
	}
	return nil
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
