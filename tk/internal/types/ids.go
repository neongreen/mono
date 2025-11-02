package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/oklog/ulid/v2"
)

// Type Definitions
// Based on tk/specs/v4-types.md

// validatePrefixedULID validates a ULID with a specific prefix
func validatePrefixedULID(s, prefix, typeName string) error {
	if !strings.HasPrefix(s, prefix) {
		return fmt.Errorf("invalid %s: must start with %s", typeName, prefix)
	}
	ulidPart := strings.TrimPrefix(s, prefix)
	if _, err := ulid.Parse(ulidPart); err != nil {
		return fmt.Errorf("invalid %s ULID part: %w", typeName, err)
	}
	return nil
}

// ProjectUID is a stable, immutable identifier for a project
type ProjectUID string

// NewProjectUID creates a new project UID
func NewProjectUID() ProjectUID {
	return ProjectUID("prj_" + ulid.Make().String())
}

// Validate checks if the ProjectUID is valid
func (p ProjectUID) Validate() error {
	return validatePrefixedULID(string(p), "prj_", "project UID")
}

func (p ProjectUID) String() string { return string(p) }

// ProjectRef is an unresolved project reference that could be a ProjectUID, alias, or project name
type ProjectRef string

// NewProjectRef creates an unresolved project reference from a string
func NewProjectRef(s string) ProjectRef {
	return ProjectRef(s)
}

// IsProjectUID checks if this reference looks like a ProjectUID (starts with "prj_")
func (r ProjectRef) IsProjectUID() bool {
	return strings.HasPrefix(string(r), "prj_")
}

func (r ProjectRef) String() string { return string(r) }

// Alias is a per-node short name for projects
type Alias string

var aliasPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Validate checks if the Alias is valid
func (a Alias) Validate() error {
	s := string(a)
	if len(s) < 2 || len(s) > 20 {
		return fmt.Errorf("alias must be 2-20 characters")
	}
	if !aliasPattern.MatchString(s) {
		return fmt.Errorf("alias must contain only alphanumeric, dash, or underscore")
	}
	return nil
}

func (a Alias) String() string { return string(a) }

// ProjectType represents the category of project
type ProjectType string

const (
	ProjectTypeLocal  ProjectType = "local"
	ProjectTypeGithub ProjectType = "github"
	ProjectTypeLinear ProjectType = "linear"
	ProjectTypeJira   ProjectType = "jira"
)

// Validate checks if the ProjectType is valid
func (p ProjectType) Validate() error {
	switch p {
	case ProjectTypeLocal, ProjectTypeGithub, ProjectTypeLinear, ProjectTypeJira:
		return nil
	default:
		return fmt.Errorf("invalid project type: %s", p)
	}
}

// TaskUID is a stable, immutable identifier for a task
type TaskUID string

// NewTaskUID creates a new task UID
func NewTaskUID() TaskUID {
	return TaskUID("tsk_" + ulid.Make().String())
}

// Validate checks if the TaskUID is valid
func (t TaskUID) Validate() error {
	return validatePrefixedULID(string(t), "tsk_", "task UID")
}

func (t TaskUID) String() string { return string(t) }

// TaskRef is an unresolved task reference that could be a TaskUID, project-number, or alias-number
type TaskRef string

// NewTaskRef creates an unresolved task reference from a string
func NewTaskRef(s string) TaskRef {
	return TaskRef(s)
}

// IsTaskUID checks if this reference looks like a TaskUID (starts with "tsk_")
func (r TaskRef) IsTaskUID() bool {
	return strings.HasPrefix(string(r), "tsk_")
}

func (r TaskRef) String() string { return string(r) }

// TaskNumber is a mutable label within a project (not identity)
type TaskNumber int64

// Validate checks if the TaskNumber is valid
func (n TaskNumber) Validate() error {
	if n <= 0 {
		return fmt.Errorf("task number must be positive")
	}
	return nil
}

func (n TaskNumber) String() string { return strconv.FormatInt(int64(n), 10) }

// NodeID is a unique identifier for a device/instance
type NodeID string

// NewNodeID creates a new node ID
func NewNodeID() NodeID {
	return NodeID(ulid.Make().String())
}

// Validate checks if the NodeID is valid
func (n NodeID) Validate() error {
	if _, err := ulid.Parse(string(n)); err != nil {
		return fmt.Errorf("invalid node ID: %w", err)
	}
	return nil
}

// Short returns last 6 characters for display hints
func (n NodeID) Short() string {
	s := string(n)
	if len(s) > 6 {
		return s[len(s)-6:]
	}
	return s
}

func (n NodeID) String() string { return string(n) }

// DisplayID is a human-friendly task identifier (derived view)
type DisplayID string

// Parse extracts components from a display ID
// Format: <alias>-<number> or <alias>-<number>-<nodeHint>
// The alias can contain hyphens, so we parse from right to left
func (d DisplayID) Parse() (alias string, number int64, nodeHint string, err error) {
	parts := strings.Split(string(d), "-")
	if len(parts) < 2 {
		return "", 0, "", fmt.Errorf("invalid display ID format")
	}

	// Parse from right to left
	// Try to parse the last segment as the node hint, second-to-last as number
	// If that fails, try to parse second-to-last as number (no node hint)

	lastIdx := len(parts) - 1
	secondLastIdx := lastIdx - 1

	// Try parsing second-to-last as number (assuming last is node hint)
	if secondLastIdx > 0 {
		num, parseErr := strconv.ParseInt(parts[secondLastIdx], 10, 64)
		if parseErr == nil {
			// Success - we have alias-number-nodeHint format
			alias = strings.Join(parts[:secondLastIdx], "-")
			number = num
			nodeHint = parts[lastIdx]
			return alias, number, nodeHint, nil
		}
	}

	// Try parsing last segment as number (no node hint)
	num, parseErr := strconv.ParseInt(parts[lastIdx], 10, 64)
	if parseErr != nil {
		return "", 0, "", fmt.Errorf("invalid number in display ID: %w", parseErr)
	}

	// Success - we have alias-number format
	alias = strings.Join(parts[:lastIdx], "-")
	number = num
	nodeHint = ""
	return alias, number, nodeHint, nil
}

func (d DisplayID) String() string { return string(d) }

// EventID is a unique identifier for an event
type EventID string

// NewEventID creates a new event ID
func NewEventID() EventID {
	return EventID(ulid.Make().String())
}

// Validate checks if the EventID is valid
func (e EventID) Validate() error {
	if _, err := ulid.Parse(string(e)); err != nil {
		return fmt.Errorf("invalid event ID: %w", err)
	}
	return nil
}

func (e EventID) String() string { return string(e) }

// EventKind represents the type of event
type EventKind string

const (
	EventKindProjectCreated     EventKind = "project.created"
	EventKindProjectAliasAdd    EventKind = "project.alias.add"
	EventKindProjectAliasRemove EventKind = "project.alias.remove"
	EventKindProjectDelete      EventKind = "project.delete"
	EventKindTaskCreated        EventKind = "task.created"
	EventKindTaskNumberSet      EventKind = "task.number.set"
	EventKindTaskRelocate       EventKind = "task.relocate"
	EventKindTaskStatusSet      EventKind = "task.status.set"
	EventKindTaskNoteAdd        EventKind = "task.note.add"
	EventKindTaskTitleSet       EventKind = "task.title.set"
	EventKindTaskDelete         EventKind = "task.delete"
	EventKindTaskMetaSet        EventKind = "task.meta.set"
	EventKindRelationAdd        EventKind = "relation.add"
	EventKindRelationRemove     EventKind = "relation.remove"
	EventKindRelationNote       EventKind = "relation.note"
)

type eventKindIndex int

const (
	eventKindProjectCreatedIndex eventKindIndex = iota
	eventKindProjectAliasAddIndex
	eventKindProjectAliasRemoveIndex
	eventKindProjectDeleteIndex
	eventKindTaskCreatedIndex
	eventKindTaskNumberSetIndex
	eventKindTaskRelocateIndex
	eventKindTaskStatusSetIndex
	eventKindTaskNoteAddIndex
	eventKindTaskTitleSetIndex
	eventKindTaskDeleteIndex
	eventKindTaskMetaSetIndex
	eventKindRelationAddIndex
	eventKindRelationRemoveIndex
	eventKindRelationNoteIndex
	eventKindCount
)

var AllEventKinds = [...]EventKind{
	eventKindProjectCreatedIndex:     EventKindProjectCreated,
	eventKindProjectAliasAddIndex:    EventKindProjectAliasAdd,
	eventKindProjectAliasRemoveIndex: EventKindProjectAliasRemove,
	eventKindProjectDeleteIndex:      EventKindProjectDelete,
	eventKindTaskCreatedIndex:        EventKindTaskCreated,
	eventKindTaskNumberSetIndex:      EventKindTaskNumberSet,
	eventKindTaskRelocateIndex:       EventKindTaskRelocate,
	eventKindTaskStatusSetIndex:      EventKindTaskStatusSet,
	eventKindTaskNoteAddIndex:        EventKindTaskNoteAdd,
	eventKindTaskTitleSetIndex:       EventKindTaskTitleSet,
	eventKindTaskDeleteIndex:         EventKindTaskDelete,
	eventKindTaskMetaSetIndex:        EventKindTaskMetaSet,
	eventKindRelationAddIndex:        EventKindRelationAdd,
	eventKindRelationRemoveIndex:     EventKindRelationRemove,
	eventKindRelationNoteIndex:       EventKindRelationNote,
}

var (
	_ [int(eventKindCount) - len(AllEventKinds)]struct{}
	_ [len(AllEventKinds) - int(eventKindCount)]struct{}
)

// TaskLabel is a complete task reference including project and number
type TaskLabel struct {
	ProjectUID ProjectUID
	Number     TaskNumber
}

// Validate checks if the TaskLabel is valid
func (l TaskLabel) Validate() error {
	if err := l.ProjectUID.Validate(); err != nil {
		return err
	}
	return l.Number.Validate()
}

// NumberPolicy represents policy for assigning task numbers when moving/relocating
type NumberPolicy struct {
	Mode   string     // "keep", "auto", "force", "fail"
	Number TaskNumber // Only used when Mode == "force"
}

// Validate checks if the NumberPolicy is valid
func (p NumberPolicy) Validate() error {
	switch p.Mode {
	case "keep", "auto", "force", "fail":
		if p.Mode == "force" {
			return p.Number.Validate()
		}
		return nil
	default:
		return fmt.Errorf("invalid number policy mode: %s", p.Mode)
	}
}
