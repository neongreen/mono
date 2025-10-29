package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/oklog/ulid/v2"
)

// Type Definitions
// Based on tk/specs/v4-types.md

// ProjectUID is a stable, immutable identifier for a project
type ProjectUID string

// NewProjectUID creates a new project UID
func NewProjectUID() ProjectUID {
	return ProjectUID("prj_" + ulid.Make().String())
}

// Validate checks if the ProjectUID is valid
func (p ProjectUID) Validate() error {
	if !strings.HasPrefix(string(p), "prj_") {
		return fmt.Errorf("invalid project UID: must start with prj_")
	}
	ulidPart := strings.TrimPrefix(string(p), "prj_")
	if _, err := ulid.Parse(ulidPart); err != nil {
		return fmt.Errorf("invalid project UID ULID part: %w", err)
	}
	return nil
}

func (p ProjectUID) String() string { return string(p) }

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
	if !strings.HasPrefix(string(t), "tsk_") {
		return fmt.Errorf("invalid task UID: must start with tsk_")
	}
	ulidPart := strings.TrimPrefix(string(t), "tsk_")
	if _, err := ulid.Parse(ulidPart); err != nil {
		return fmt.Errorf("invalid task UID ULID part: %w", err)
	}
	return nil
}

func (t TaskUID) String() string { return string(t) }

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
func (d DisplayID) Parse() (alias string, number int64, nodeHint string, err error) {
	parts := strings.Split(string(d), "-")
	if len(parts) < 2 || len(parts) > 3 {
		return "", 0, "", fmt.Errorf("invalid display ID format")
	}

	alias = parts[0]
	number, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid number in display ID: %w", err)
	}

	if len(parts) == 3 {
		nodeHint = parts[2]
	}

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
	EventKindTaskCreated        EventKind = "task.created"
	EventKindTaskNumberSet      EventKind = "task.number.set"
	EventKindTaskRelocate       EventKind = "task.relocate"
	EventKindTaskStatusSet      EventKind = "task.status.set"
	EventKindTaskNoteAdd        EventKind = "task.note.add"
	EventKindTaskTitleSet       EventKind = "task.title.set"
	EventKindRelationAdd        EventKind = "relation.add"
	EventKindRelationRemove     EventKind = "relation.remove"
	EventKindRelationNote       EventKind = "relation.note"
)

type eventKindIndex int

const (
	eventKindProjectCreatedIndex eventKindIndex = iota
	eventKindProjectAliasAddIndex
	eventKindProjectAliasRemoveIndex
	eventKindTaskCreatedIndex
	eventKindTaskNumberSetIndex
	eventKindTaskRelocateIndex
	eventKindTaskStatusSetIndex
	eventKindTaskNoteAddIndex
	eventKindTaskTitleSetIndex
	eventKindRelationAddIndex
	eventKindRelationRemoveIndex
	eventKindRelationNoteIndex
	eventKindCount
)

var AllEventKinds = [...]EventKind{
	eventKindProjectCreatedIndex:     EventKindProjectCreated,
	eventKindProjectAliasAddIndex:    EventKindProjectAliasAdd,
	eventKindProjectAliasRemoveIndex: EventKindProjectAliasRemove,
	eventKindTaskCreatedIndex:        EventKindTaskCreated,
	eventKindTaskNumberSetIndex:      EventKindTaskNumberSet,
	eventKindTaskRelocateIndex:       EventKindTaskRelocate,
	eventKindTaskStatusSetIndex:      EventKindTaskStatusSet,
	eventKindTaskNoteAddIndex:        EventKindTaskNoteAdd,
	eventKindTaskTitleSetIndex:       EventKindTaskTitleSet,
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
