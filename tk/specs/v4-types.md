# tk v4 Type Definitions

This document defines all ID types and data structures for tk v4.

---

## Overview

tk v4 uses strongly-typed identifiers for all entities. This document specifies:
- Format for each ID type
- Examples
- Validation rules
- Go newtype definitions

---

## Project Types

### ProjectUID

**Format**: `prj_` + ULID  
**Example**: `prj_01J5QKF7F8M9N0P1Q2R3S4T5`  
**Purpose**: Stable, immutable identifier for a project

**Validation**:
- Must start with `prj_`
- Rest must be valid ULID (26 characters)
- Case-sensitive

**Go Definition**:
```go
type ProjectUID string

func NewProjectUID() ProjectUID {
    return ProjectUID("prj_" + ulid.Make().String())
}

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
```

### Alias

**Format**: String (2-20 characters, alphanumeric + dash/underscore)  
**Example**: `tk`, `backend`, `work-projects`  
**Purpose**: Per-node short name for projects

**Validation**:
- Length: 2-20 characters
- Characters: alphanumeric, dash (-), underscore (_)
- Case-sensitive
- Per-node scoped (collisions allowed)

**Go Definition**:
```go
type Alias string

func (a Alias) Validate() error {
    s := string(a)
    if len(s) < 2 || len(s) > 20 {
        return fmt.Errorf("alias must be 2-20 characters")
    }
    matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", s)
    if !matched {
        return fmt.Errorf("alias must contain only alphanumeric, dash, or underscore")
    }
    return nil
}

func (a Alias) String() string { return string(a) }
```

### ProjectType

**Format**: String enum  
**Values**: `local`, `github`, `linear`, `jira`  
**Purpose**: Category of project

**Go Definition**:
```go
type ProjectType string

const (
    ProjectTypeLocal  ProjectType = "local"
    ProjectTypeGithub ProjectType = "github"
    ProjectTypeLinear ProjectType = "linear"
    ProjectTypeJira   ProjectType = "jira"
)

func (p ProjectType) Validate() error {
    switch p {
    case ProjectTypeLocal, ProjectTypeGithub, ProjectTypeLinear, ProjectTypeJira:
        return nil
    default:
        return fmt.Errorf("invalid project type: %s", p)
    }
}
```

---

## Task Types

### TaskUID

**Format**: `tsk_` + ULID  
**Example**: `tsk_01J5QKF7F8M9N0P1Q2R3S4T5`  
**Purpose**: Stable, immutable identifier for a task (true identity)

**Validation**:
- Must start with `tsk_`
- Rest must be valid ULID
- Case-sensitive

**Go Definition**:
```go
type TaskUID string

func NewTaskUID() TaskUID {
    return TaskUID("tsk_" + ulid.Make().String())
}

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
```

### TaskNumber

**Format**: Integer (64-bit)  
**Example**: `1`, `42`, `1001`  
**Purpose**: Mutable label within a project (not identity)

**Validation**:
- Must be positive (> 0)
- No upper bound
- Not unique (collisions allowed)
- Resolved at display time

**Go Definition**:
```go
type TaskNumber int64

func (n TaskNumber) Validate() error {
    if n <= 0 {
        return fmt.Errorf("task number must be positive")
    }
    return nil
}

func (n TaskNumber) String() string { return strconv.FormatInt(int64(n), 10) }
```

---

## Node Types

### NodeID

**Format**: ULID  
**Example**: `01J5QKF7F8M9N0P1Q2R3S4T5`  
**Purpose**: Unique identifier for a device/instance

**Validation**:
- Valid ULID (26 characters)
- Case-sensitive

**Go Definition**:
```go
type NodeID string

func NewNodeID() NodeID {
    return NodeID(ulid.Make().String())
}

func (n NodeID) Validate() error {
    if _, err := ulid.Parse(string(n)); err != nil {
        return fmt.Errorf("invalid node ID: %w", err)
    }
    return nil
}

func (n NodeID) Short() string {
    // Return last 6 characters for display hints
    s := string(n)
    if len(s) > 6 {
        return s[len(s)-6:]
    }
    return s
}

func (n NodeID) String() string { return string(n) }
```

---

## Display String Types

These are derived views, never stored in the database.

### DisplayID

**Format**: `<alias>-<number>` or `<alias>-<number>-<node_hint>`  
**Example**: `tk-1`, `tk-1-abc123`  
**Purpose**: Human-friendly task identifier

**Components**:
- `alias`: Project alias on this node
- `number`: Task number within project
- `node_hint`: Last 6 chars of NodeID (only when collision exists)

**Validation**:
- Parses as `<alias>-<number>` or `<alias>-<number>-<node_hint>`
- Alias must be valid
- Number must be positive integer

**Go Definition**:
```go
type DisplayID string

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
```

### ExternalDisplayID

**Format**: `<owner>/<repo>#<number>`  
**Example**: `neongreen/mono#123`  
**Purpose**: Display format for external projects (future)

**Go Definition**:
```go
type ExternalDisplayID string

func (d ExternalDisplayID) Parse() (owner, repo string, number int64, err error) {
    // Parse <owner>/<repo>#<number>
    // Implementation deferred to external project support
    return "", "", 0, fmt.Errorf("not implemented")
}

func (d ExternalDisplayID) String() string { return string(d) }
```

---

## Event Types

### EventID

**Format**: ULID  
**Example**: `01J5QKF7F8M9N0P1Q2R3S4T5`  
**Purpose**: Unique identifier for an event

**Go Definition**:
```go
type EventID string

func NewEventID() EventID {
    return EventID(ulid.Make().String())
}

func (e EventID) Validate() error {
    if _, err := ulid.Parse(string(e)); err != nil {
        return fmt.Errorf("invalid event ID: %w", err)
    }
    return nil
}

func (e EventID) String() string { return string(e) }
```

### EventKind

**Format**: String  
**Values**: See event specifications in v4.md  
**Purpose**: Type of event

**Go Definition**:
```go
type EventKind string

const (
    EventKindProjectCreated   EventKind = "project.created"
    EventKindProjectAliasAdd  EventKind = "project.alias.add"
    EventKindProjectAliasRemove EventKind = "project.alias.remove"
    EventKindTaskCreated      EventKind = "task.created"
    EventKindTaskNumberSet    EventKind = "task.number.set"
    EventKindTaskRelocate     EventKind = "task.relocate"
    EventKindTaskStatusSet    EventKind = "task.status.set"
    // ... etc
)

func (k EventKind) Validate() error {
    // Check against known event kinds
    return nil
}
```

---

## Composite Types

### TaskLabel

**Purpose**: Complete task reference including project and number

**Go Definition**:
```go
type TaskLabel struct {
    ProjectUID ProjectUID
    Number     TaskNumber
}

func (l TaskLabel) Validate() error {
    if err := l.ProjectUID.Validate(); err != nil {
        return err
    }
    return l.Number.Validate()
}
```

### NumberPolicy

**Purpose**: Policy for assigning task numbers when moving/relocating

**Go Definition**:
```go
type NumberPolicy struct {
    Mode  string  // "keep", "auto", "force"
    Number TaskNumber // Only used when Mode == "force"
}

func (p NumberPolicy) Validate() error {
    switch p.Mode {
    case "keep", "auto", "force":
        // Valid modes
        if p.Mode == "force" {
            return p.Number.Validate()
        }
        return nil
    default:
        return fmt.Errorf("invalid number policy mode: %s", p.Mode)
    }
}
```

---

## Validation Rules Summary

| Type | Prefix | Format | Uniqueness | Mutable |
|------|--------|--------|------------|---------|
| ProjectUID | `prj_` | ULID | Global unique | No |
| TaskUID | `tsk_` | ULID | Global unique | No |
| NodeID | none | ULID | Global unique | No |
| Alias | none | String 2-20 chars | Per-node only | Yes |
| TaskNumber | none | Int > 0 | Per-project only | Yes |
| DisplayID | none | `alias-number[-node_hint]` | Derived | No |
| EventID | none | ULID | Global unique | No |

---

## Type Relationships

```
ProjectUID (prj_...)
  └─> Alias[] (per-node, may collide)
  └─> TaskUID[] (tsk_..., many tasks)
       └─> TaskLabel{ProjectUID, TaskNumber} (label, not identity)
            └─> DisplayID (derived: alias-number[-node_hint])

NodeID (01J5Q...)
  └─> Aliases[NodeID] (aliases per node)
```

---

## Usage Examples

### Creating IDs
```go
// Create a new project
projectUID := NewProjectUID()
// -> prj_01J5QKF7F8M9N0P1Q2R3S4T5

// Create a new task
taskUID := NewTaskUID()
// -> tsk_01J5QKF7F8M9N0P1Q2R3S4T5

// Create alias
alias := Alias("tk")
if err := alias.Validate(); err != nil {
    // handle error
}

// Assign task number (label)
number := TaskNumber(1)
if err := number.Validate(); err != nil {
    // handle error
}
```

### Rendering Display IDs
```go
// When no collision:
DisplayID: "tk-1"

// When collision exists:
DisplayID: "tk-1-abc123"  // abc123 = last 6 chars of NodeID
```

### Parsing Display IDs
```go
display := DisplayID("tk-1-abc123")
alias, number, nodeHint, err := display.Parse()
// alias = "tk", number = 1, nodeHint = "abc123"
```
