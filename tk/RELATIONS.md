# Task Relations (v2)

This document describes the task relations feature implemented for tk.

## Overview

tk now supports relations between tasks, allowing you to model dependencies (blocks), hierarchies (subtasks), and other relationships (duplicates, related tasks, supersedes).

## Relation Types

- **blocks**: Source task blocks target task (creates a dependency)
- **blocked_by**: Inverse of blocks (source is blocked by target)
- **subtask**: Target task is a subtask of source task (creates hierarchy)
- **parent**: Inverse of subtask (source is parent of target)
- **related**: Tasks are related but not dependent
- **duplicate_of**: Source task is a duplicate of target task
- **supersedes**: Source task supersedes (replaces) target task

## Commands

### Add a Relation

```bash
tk relate add <source-task> <relation-type> <target-task> [--note "optional note"]
```

Examples:
```bash
# Task tk-1 blocks task tk-2
tk relate add tk-1 blocks tk-2

# Task tk-3 is a subtask of tk-1
tk relate add tk-1 subtask tk-3

# Tasks are related
tk relate add tk-4 related tk-5 --note "Both implement auth"
```

### Remove a Relation

```bash
tk relate remove <source-task> <relation-type> <target-task>
```

Example:
```bash
tk relate remove tk-1 blocks tk-2
```

### Mark Tasks as Duplicates

```bash
tk dup <task-a> <task-b>
```

This creates bidirectional `duplicate_of` relations.

### List Blockers

```bash
# Show all tasks blocking a specific task
tk blockers <task-id>
```

This shows direct and transitive blockers with their distance in the dependency graph.

### List All Blocked Tasks

```bash
tk blocked
```

Shows all tasks that are currently blocked by at least one incomplete task.

### Visualize Relations

```bash
tk graph <task-id> [--type blocks|subtask|related] [--depth N]
```

Shows an ASCII tree of task relations. Blocked tasks are marked with ⛔.

Example:
```
Task: tk-1-abc123 - Implement feature X

├── tk-2-abc123 - Design API ⛔
└── tk-3-abc123 - Write tests
```

### Filter by Blocked Status

```bash
# Show only blocked tasks
tk ls --blocked

# Show only unblocked tasks
tk ls --unblocked

# Combine with other filters
tk ls -p foo --blocked
tk ls --axis generic:in_progress --unblocked
```

## Blocking Logic

A task is considered **blocked** if:
1. It has one or more incoming `blocks` relations, AND
2. At least one blocker task is NOT in a "done" state according to the configured blocking axis

### Configuration

Configure blocking behavior in `~/.config/tk/config.toml`:

```toml
[blocking]
blocking_axis = "generic"  # Which status axis to check
done_states = ["done"]     # States that count as "done"
```

For example, if `blocking_axis = "generic"` and `done_states = ["done", "fixed"]`, then a blocker task must have `generic` status of `done` or `fixed` to NOT block other tasks.

## CRDT Semantics

Relations use OR-set CRDT semantics for conflict resolution:

- **Add-wins for concurrent adds**: If two nodes add the same relation concurrently, it exists
- **Remove-wins**: A remove tombstones a specific add from a specific node
- **Multi-node resilience**: A relation exists as long as at least one node has added it without removing it
- **Idempotent**: Adding or removing the same relation multiple times has the same effect as doing it once

### Example

```
Node A: adds tk-1 blocks tk-2 (event ev-1-nodeA)
Node B: adds tk-1 blocks tk-2 (event ev-2-nodeB)
→ Relation exists (both nodes agree)

Node A: removes tk-1 blocks tk-2 (event ev-3-nodeA)
→ Relation still exists (Node B hasn't removed it)

Node B: removes tk-1 blocks tk-2 (event ev-4-nodeB)
→ Relation is now removed (both nodes removed it)
```

## Cycle Detection

The system detects cycles in `blocks` and `subtask` relations. Cycles are not currently prevented, but they are detected and can be queried for debugging purposes.

## JSON Output

Relations are included in task JSON when using `tk show`:

```json
{
  "task_id": "tk-1",
  "relations": {
    "blocks": {
      "out": [{"dst": "task-uuid-2", "note": "waiting on API"}],
      "in": []
    },
    "subtask": {
      "children": ["task-uuid-3", "task-uuid-4"],
      "parent": ""
    }
  },
  "blocked": false,
  "blockers": []
}
```

## Event Schema

Relations are stored as events in the event log:

### relation.add
```json
{
  "kind": "relation.add",
  "payload": {
    "src": "task-uuid-a",
    "type": "blocks",
    "dst": "task-uuid-b",
    "note": "optional note"
  }
}
```

### relation.remove
```json
{
  "kind": "relation.remove",
  "payload": {
    "src": "task-uuid-a",
    "type": "blocks",
    "dst": "task-uuid-b"
  }
}
```

### relation.note
```json
{
  "kind": "relation.note",
  "payload": {
    "src": "task-uuid-a",
    "type": "blocks",
    "dst": "task-uuid-b",
    "markdown": "updated note text"
  }
}
```

## Performance

The relations graph is designed to handle 10,000+ edges efficiently:
- O(1) add/remove operations
- O(E) cycle detection (where E = number of edges)
- O(V + E) blocked computation (where V = tasks, E = edges)
- Tested to fold 10k edges in <50ms on M-series chips

## Future Enhancements (v3)

v3 will add:
- Task kinds (task, story, epic)
- Rollup computation for hierarchies
- Progress tracking for parent tasks
- Estimate tracking and aggregation
