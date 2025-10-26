# tk v4 Migration Guide

This document describes how to migrate from tk v1/v2 (prefix-based) to v4 (project-based).

---

## Executive Summary

### What v4 Changes

**Current (v1/v2)**:
- Tasks organized by **prefixes** (e.g., `tk`, `foo`, `bar`)
- Task IDs: `prefix-number-node` (e.g., `tk-1-abc123`)
- Prefixes are first-class, numbers are per-(prefix,node) counters
- Manual migration required with export/import

**Target (v4)**:
- Tasks organized by **projects** (with stable UIDs like `prj_01J5Q...`)
- Project **aliases** per node (e.g., `tk`, `backend`) - may collide across nodes
- Task IDs: **derived display strings** (e.g., `tk-1`, or `tk-1-abc` if collision)
- Task identity: **task_uid** (ULID, e.g., `tsk_01J5Q...`)
- Numbers are **mutable labels**, not identity - collisions allowed
- **Automatic migration** on first run

**Impact**: 🔴 **Breaking change** - new data model, event schema, CLI commands

**Timeline**: Automatic upgrade on first run of v4 binary (~seconds to minutes)

---

## Migration Approach: Hard Break with Automatic Upgrade

This migration uses a **hard break** strategy with automatic in-place upgrade. The v4 binary:

1. Detects v1/v2 database on startup
2. Creates safety snapshot (`tk.db.v3.bak`)
3. Upgrades schema in-place
4. Backfills v4 events from legacy data
5. Finalizes and marks database as v4

**No manual export/import needed** - everything happens automatically.

---

## Automatic Migration Process

When you first run the v4 binary on a v1/v2 database, migration runs automatically.

### Step 1: Safety Snapshot
1. Copy current `tk.db` → `tk.db.v3.bak`
2. Create `migrate-v4.lock` to prevent concurrent runs

The original database is preserved as a backup you can restore if needed.

### Step 2: Schema Upgrade
Adds new tables in-place:

```sql
projects(project_uid PK, type, origin_json, name, description, created_at, created_by);
project_aliases(project_uid, alias, node, added_by, PRIMARY KEY(alias,node));
tasks(task_uid PK, project_uid, created_node, title, created_at, created_by);
task_numbers(project_uid, number, task_uid);
metadata(key,value);
```

The existing `events` table continues to store all events. New tables are projections of those events.

### Step 3: Backfill → v4 Events

For every legacy record, emits corresponding v4 events:

| legacy | v4 event(s) |
|--------|-------------|
| prefix.created | project.created {type:"local",name:<prefix>} + project.alias.add {alias:<prefix>,node:<this>} |
| task record | task.created {task_uid,project_uid,proposed_number} + task.number.set {number:<old>} |
| task.reprefix | task.relocate {from_project_uid,to_project_uid,number_policy:{mode:"force",number:<new>}} |

Events are emitted in timestamp order and projected immediately.

### Step 4: Finalize
- Set meta.version_major = 4
- Set config.remote_subdir = "tk-v4"
- Drop the lock
- Keep tk.db.v3.bak for rollback

Your database is now v4. Existing segments in your remote remain untouched.

---

## Rollback

If something goes wrong, rollback to v3:

```bash
tk admin rollback-v4
```

This:
1. Restores tk.db.v3.bak as tk.db
2. Resets meta.version_major = 3
3. Lets you use v1/v2 binaries again

The v4 segments in `tk-v4/` remain untouched and can be ignored.

---

## What Happens to Your Data

### Current State (v1/v2)

```
Prefixes:
  tk (node: abc123, desc: "Personal tasks")
  work (node: abc123, desc: "Work tasks")

Tasks:
  tk-1-abc123 → task_uuid_1 "Write docs"
  tk-2-abc123 → task_uuid_2 "Fix bug"
  work-1-abc123 → task_uuid_3 "Review PR"
```

### After Migration (v4)

```
Projects:
  prj_01J5Q1... (name: "Personal tasks", created_by: alice)
  prj_01J5Q2... (name: "Work tasks", created_by: alice)

Aliases (node: abc123):
  "tk" → prj_01J5Q1...
  "work" → prj_01J5Q2...

Tasks:
  tsk_01J5Q1... (project: prj_01J5Q1..., created_node: abc123, title: "Write docs")
  tsk_01J5Q2... (project: prj_01J5Q1..., created_node: abc123, title: "Fix bug")
  tsk_01J5Q3... (project: prj_01J5Q2..., created_node: abc123, title: "Review PR")

Task Numbers:
  (prj_01J5Q1..., 1, tsk_01J5Q1...)  → displays as "tk-1"
  (prj_01J5Q1..., 2, tsk_01J5Q2...)  → displays as "tk-2"
  (prj_01J5Q2..., 1, tsk_01J5Q3...)  → displays as "work-1"

Old IDs Still Work (via aliases):
  tk-1-abc123 → resolves to tsk_01J5Q1... → displays as "tk-1"
  work-1-abc123 → resolves to tsk_01J5Q3... → displays as "work-1"
```

Your tasks are all preserved. Old task IDs still resolve to the correct tasks.

---

## Core Conceptual Changes

### 1. Identity Model

| Aspect | v1/v2 | v4 |
|--------|-------|----|
| Project identity | Prefix string (mutable) | project_uid (immutable ULID) |
| Project naming | Prefix (global scope) | Alias (per-node, may collide) |
| Task identity | Task UUID (immutable) | task_uid (immutable ULID) |
| Task display ID | prefix-number-node | Derived from alias-number (+node hint) |
| Task numbering | Counter per (prefix, node) | Label per (project_uid), may collide |

### 2. Organizational Model

- **v1/v2**: Prefixes are the organizational unit, scoped to nodes
- **v4**: Projects are the organizational unit, aliases are scoped to nodes

### 3. Number Semantics

- **v1/v2**: Numbers are quasi-stable (change only on explicit `mv`)
- **v4**: Numbers are fully mutable labels, collisions allowed and resolved at display time

---

## Visual Comparison

### Task Creation

**v1/v2:**
```bash
tk prefix create foo "Foo project"  # Create organizational unit
tk new --prefix foo "Fix bug"       # Create task
# Result: foo-1-abc123               # Task ID stored as-is
```

**v4:**
```bash
tk project create "Foo project"              # Create project → prj_01J5Q...
tk project alias foo prj_01J5Q...            # Create alias on this node
tk new --project foo "Fix bug"               # Create task → tsk_01J5Q...
# Result: foo-1                               # Display string (derived)
# (or foo-1-abc if another node also has foo-1)
```

### Task Movement / Renumbering

**v1/v2:**
```bash
tk mv foo-1 bar:2
# Emits: task.reprefix {old_prefix: foo, new_prefix: bar, old_number: 1, new_number: 2}
# Result: Task ID changes to bar-2-abc123
# Alias: foo-1-abc123 → bar-2-abc123 (preserved)
```

**v4:**
```bash
# Change number within same project:
tk number set foo-1 2
# Emits: task.number.set {task_uid, project_uid, number: 2}
# Result: Display changes from foo-1 to foo-2
# The old number foo-1 becomes available again for new tasks

# Move to different project (atomic operation):
tk move foo-1 --to bar --force 2
# Emits: task.relocate {from_project_uid, to_project_uid, number_policy}
# Result: Task moved to bar project, numbered as bar-2
```

---

## Collision Scenarios

### v1/v2: Guaranteed Unique per Node

```
Node A: Creates tk-1-nodeA
Node B: Creates tk-1-nodeB
After sync:
  - Both exist as tk-1-nodeA and tk-1-nodeB
  - No collision (node suffix ensures uniqueness)
  - Display always shows full node suffix
```

### v4: Collisions Allowed and Resolved

```
Node A: Creates task → tsk_A, assigns number 1 in project prj_tk
Node B: Creates task → tsk_B, assigns number 1 in project prj_tk
After sync:
  task_numbers table:
    (prj_tk, 1, tsk_A)  ← Both entries exist
    (prj_tk, 1, tsk_B)

  Display:
    tk-1-abc  ← Node A's task (with node hint)
    tk-1-def  ← Node B's task (with node hint)

  Resolution (optional):
    tk number set tk-1-def 2  → Renumber one task
    Result: tk-1 and tk-2 (no more collision)
```

### Alias Collisions (New in v4)

```
Node A: Creates project prj_personal, adds alias "work"
Node B: Creates project prj_company, adds alias "work"

Result:
  - Node A: "work" resolves to prj_personal
  - Node B: "work" resolves to prj_company
  - No conflict! Aliases are per-node.

When using tasks:
  Node A: tk new --project work "Task"  → Creates in prj_personal
  Node B: tk new --project work "Task"  → Creates in prj_company
  
  These are tasks in different projects, not collision.
```

This is **allowed by design** in v4 (aliases are per-node).

---

## New Events

| event | purpose | replaces |
|-------|---------|----------|
| project.created | define project | prefix.created |
| project.alias.add/remove | manage aliases | prefix.created |
| task.created | create task | task.created (with project_uid) |
| task.number.set | assign / change label | task.reprefix (partially) |
| task.relocate | atomic move + renumber | task.reprefix sequence |

The new `task.relocate` event replaces the legacy "move then renumber" sequence. It atomically handles moving a task from one project to another and renumbering within the new project.

---

## Commands That Changed

| v1/v2 | v4 | Notes |
|-------|-----|-------|
| `tk prefix create foo "desc"` | `tk project create "desc"`<br>`tk project alias foo <uid>` | Two commands in v4 |
| `tk prefix list` | `tk project list` | Similar output |
| `tk new --prefix foo "title"` | `tk new --project foo "title"` | `--project` instead of `--prefix` |
| `tk mv foo-1 bar:2` | `tk number set foo-1 2`<br>`tk move foo-1 --to bar --force 2` | v4: renumber or move separately |
| `tk ls --prefix foo` | `tk ls --project foo` | `--project` instead of `--prefix` |
| `tk view foo-1-abc123` | `tk view foo-1`<br>`tk view foo-1-abc` | Node hint only if needed |

Note: `tk mv` is removed. (Optional hidden alias may exist for one release.)

---

## Data Model Comparison

### Database Tables

**v1/v2:**
```sql
events (id, ts, actor, role, kind, payload, ...)
prefixes (prefix, node, description, created_at, created_by, removed)
prefix_counters (prefix, node, last_id)
task_counter (last_id)  -- legacy
```

**v4:**
```sql
events (id, ts, actor, role, kind, payload, ...)
projects (project_uid PK, type, origin_json, name, description, created_at, created_by)
project_aliases (project_uid, alias, node, added_by, PRIMARY KEY(alias, node))
tasks (task_uid PK, project_uid, created_node, title, created_at, created_by, ...)
task_numbers (project_uid, number, task_uid)  -- Note: NOT unique!
```

### Event Schema

**v1/v2:**
```json
{
  "kind": "prefix.created",
  "payload": {
    "prefix": "foo",
    "description": "Foo project tasks",
    "created_by": "alice"
  }
}

{
  "kind": "task.created",
  "payload": {
    "task_uuid": "task-abc123...",
    "task_id": "foo-1-nodeA",
    "title": "Fix bug",
    "created_by": "alice"
  }
}

{
  "kind": "task.reprefix",
  "payload": {
    "task_uuid": "task-abc123...",
    "old_prefix": "foo",
    "new_prefix": "bar",
    "old_number": 1,
    "new_number": 2,
    "old_node": "nodeA"
  }
}
```

**v4:**
```json
{
  "kind": "project.created",
  "payload": {
    "project_uid": "prj_01J5Q...",
    "type": "local",
    "origin": null,
    "name": "Foo project",
    "description": "Foo project tasks",
    "created_by": "alice"
  }
}

{
  "kind": "project.alias.add",
  "payload": {
    "project_uid": "prj_01J5Q...",
    "alias": "foo",
    "node": "nodeA",
    "added_by": "alice"
  }
}

{
  "kind": "task.created",
  "payload": {
    "task_uid": "tsk_01J5Q...",
    "project_uid": "prj_01J5Q...",
    "proposed_number": 1,
    "created_node": "nodeA",
    "title": "Fix bug",
    "created_by": "alice"
  }
}

{
  "kind": "task.relocate",
  "payload": {
    "task_uid": "tsk_01J5Q...",
    "from_project_uid": "prj_foo...",
    "to_project_uid": "prj_bar...",
    "number_policy": {
      "mode": "force",
      "number": 2
    }
  }
}
```

---

## Guardrails and Diagnostics

The v4 binary includes several safety measures:

### Version Guard
- The v4 binary refuses to ingest any segment not under `tk-v4/` or without `"tk_major": 4`.
- This prevents accidental ingestion of v1/v2 segments.

### Automatic Health Check
- `tk doctor` auto-runs post-upgrade and verifies:
  - every task has a valid project,
  - all aliases resolve,
  - reports any label collisions.

### Collision Detection
- `tk conflicts numbers --project <x>` lists label collisions
- Shows which tasks share the same number within a project

### Event Statistics
- `tk events stats` confirms event count parity with pre-migration total
- Helps verify migration completed successfully

---

## Key Benefits of v4

### 1. External Integration Ready
```
# GitHub project
project_uid: prj_github_1
type: github
origin: {"owner": "neongreen", "repo": "mono"}
name: "neongreen/mono"

# Display as: neongreen/mono#123
# Internal: Still uses task_uid for identity
```

### 2. Flexible Numbering
```
# Can renumber tasks without changing identity
tk number set old-task 100  # "Fresh start" numbering
tk number set urgent-1 999  # Priority numbering

# Numbers are just labels, identity is task_uid
```

### 3. Multi-Node Alias Flexibility
```
# Alice's node:
  project alias work → prj_company
  project alias personal → prj_personal

# Bob's node:
  project alias work → prj_opensource
  project alias personal → prj_personal

# Same alias name, different meanings per node
# No conflicts because aliases are per-node
```

### 4. Collision Tolerance
```
# Two people create "urgent-1" offline
# After sync: Both exist as urgent-1-alice and urgent-1-bob
# No data loss, no merge conflicts
# Can renumber later: tk number set urgent-1-bob 2
```

---

## Summary

Migration from v1/v2 to v4 is **automatic** and **safe**:

1. ✅ **No data loss** - all tasks preserved
2. ✅ **Old IDs work** - legacy task IDs resolve via aliases
3. ✅ **Automatic** - happens on first run of v4 binary
4. ✅ **Reversible** - `tk admin rollback-v4` to go back
5. ✅ **Safe** - backup created before migration

Key changes:
- Prefixes → Projects with stable UIDs
- Prefix names → Per-node aliases
- Task IDs (stored) → Task display strings (derived)
- Numbers (counters) → Numbers (labels, may collide)

Result: a simple distributed model that remains fully offline-safe, permits collisions and renumbering, and keeps user-facing labels pleasant and human.
