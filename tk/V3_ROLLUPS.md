# v3 Rollups Implementation Plan

This document outlines what would be needed to implement v3 rollups for hierarchical task planning.

## Status

**NOT YET IMPLEMENTED** - v2 relations are complete and production-ready. v3 rollups are planned but not required for basic blocking/subtask functionality.

## Goals (from spec)

1. Hierarchical planning with derived progress and status
2. No hardcoded levels, just subtask edges and optional kind
3. Rollups tolerate partial data and conflicts

## What Would Need to Be Added

### 1. Event Types

```go
// TaskKindSetPayload for task.kind.set events
type TaskKindSetPayload struct {
    TaskUUID string `json:"task_uuid"`
    Kind     string `json:"kind"` // task, story, epic
}

// EstimateSetPayload for estimate.set events (optional)
type EstimateSetPayload struct {
    TaskUUID string `json:"task_uuid"`
    Points   *int   `json:"points,omitempty"`
    Hours    *int   `json:"hours,omitempty"`
}
```

### 2. Task Structure Updates

Add to Task type:
```go
type Task struct {
    // ... existing fields ...
    Kind     string   `json:"kind,omitempty"`     // task, story, epic
    Estimate *Estimate `json:"estimate,omitempty"` // optional
    Progress *Progress `json:"progress,omitempty"` // derived from children
}

type Estimate struct {
    Points int `json:"points,omitempty"`
    Hours  int `json:"hours,omitempty"`
}

type Progress struct {
    CountDone  int     `json:"count_done"`
    CountTotal int     `json:"count_total"`
    Percent    float64 `json:"percent"`
    PointsDone int     `json:"points_done,omitempty"`
    PointsTotal int    `json:"points_total,omitempty"`
}
```

### 3. Rollup Computation

Create `rollup.go` with:
```go
// ComputeRollups computes progress for all tasks with children
func ComputeRollups(tasks map[string]*Task, graph *RelationsGraph, config *Config) {
    // For each task with subtask children:
    // 1. Get all children (direct subtasks)
    // 2. Count how many are done (check axis against done_states)
    // 3. Sum estimates if present
    // 4. Compute percentage
    // 5. Derive parent status based on children
}

// DeriveParentStatus computes parent status from children
func DeriveParentStatus(children []*Task, config *Config) string {
    // blocked if any child blocked
    // in_progress if any child in_progress and none blocked
    // done if all children done
}
```

### 4. CLI Commands

```bash
# Set task kind
tk promote <task-id> story|epic|task

# Show rollup summary
tk rollup <task-id>

# Show task tree with status badges
tk tree <task-id> [--wide]

# Set estimate
tk estimate <task-id> --points 5
tk estimate <task-id> --hours 8
```

### 5. Queries

Add to `tk ls`:
```bash
tk ls --kind story
tk ls --kind epic --progress '<100%'
tk ls --kind story --blocked
```

### 6. Configuration

Add to config:
```toml
[rollup]
axis = "generic"
done_states = ["done"]
policy = "any_in_progress"  # or "strict_all_done"
```

### 7. Tests

- Nested trees (3-4 levels) with correct counts
- Mixed child states producing expected parent status
- Cycles flagged and excluded from rollup
- Performance: 1000 tasks with 100 stories in <100ms

## Why Not Implemented Yet

v2 relations provides the foundation for blocking and dependency tracking, which was the primary goal. v3 rollups add hierarchical planning with progress tracking, which is valuable but:

1. Not required for basic blocking/subtask functionality
2. Can be added incrementally without breaking changes
3. The subtask relation structure is already in place
4. The spec says "land v2 first to unblock workflow"

## Implementation Effort

Estimated: 4-6 hours to fully implement v3 rollups with all features, tests, and documentation.

## When to Implement

Consider implementing v3 when:
- Users need to track multi-level hierarchies (tasks → stories → epics)
- Progress tracking for parent tasks is needed
- Estimation and rollup of estimates is valuable
- Current v2 relations are working well in production
