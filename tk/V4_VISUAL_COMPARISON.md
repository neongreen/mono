# tk v4: Visual Comparison & Key Concepts

Quick visual guide to understand v4 changes. For full details, see:
- **Start here**: [V4_MIGRATION_SUMMARY.md](./V4_MIGRATION_SUMMARY.md)
- **Decisions**: [V4_MIGRATION_DECISIONS.md](./V4_MIGRATION_DECISIONS.md)  
- **Deep dive**: [MIGRATION_V4_STRATEGY.md](./MIGRATION_V4_STRATEGY.md)

---

## Core Concept Change

### Current (v1/v2): Prefix-Centric Model

```
                    ┌─────────────────┐
                    │     Prefix      │  ← First-class entity
                    │   "tk" (node1)  │  ← Scoped to node
                    └─────────────────┘
                            │
            ┌───────────────┼───────────────┐
            ▼               ▼               ▼
      ┌─────────┐     ┌─────────┐     ┌─────────┐
      │  tk-1   │     │  tk-2   │     │  tk-3   │  ← Task IDs
      │ -node1  │     │ -node1  │     │ -node1  │  ← (prefix-num-node)
      └─────────┘     └─────────┘     └─────────┘
           │               │               │
      ┌─────────┐     ┌─────────┐     ┌─────────┐
      │task_uuid│     │task_uuid│     │task_uuid│  ← Immutable identity
      └─────────┘     └─────────┘     └─────────┘
```

**Identity**: task_uuid (immutable) + task_id (partially mutable via `tk mv`)
**Display**: prefix-number-node (e.g., `tk-1-abc123`)
**Numbering**: Counter per (prefix, node) — guaranteed unique per node

---

### Target (v4): Project-Centric Model

```
                    ┌──────────────────┐
                    │     Project      │  ← First-class entity
                    │  prj_01J5Q...    │  ← Stable ULID (immutable)
                    │  type: local     │
                    └──────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
    ┌────────┐         ┌────────┐         ┌────────┐
    │ Alias  │         │ Alias  │         │ Alias  │  ← Per-node aliases
    │ "tk"   │         │ "work" │         │ "mono" │  ← (may collide!)
    │(node1) │         │(node2) │         │(node1) │
    └────────┘         └────────┘         └────────┘
        │
        │  (all point to same project prj_01J5Q...)
        │
    ┌───┴───────────────────────┬───────────────┐
    ▼                           ▼               ▼
┌──────────┐              ┌──────────┐    ┌──────────┐
│ Task     │              │ Task     │    │ Task     │
│tsk_01... │              │tsk_02... │    │tsk_03... │  ← Immutable UIDs
└──────────┘              └──────────┘    └──────────┘
    │                           │               │
    │                           │               │
┌───┴────┐                 ┌────┴───┐      ┌────┴───┐
│Number:1│                 │Number:2│      │Number:3│  ← Mutable labels
│(label) │                 │(label) │      │(label) │  ← (may collide!)
└────────┘                 └────────┘      └────────┘
    │                           │               │
    ▼                           ▼               ▼
Display: tk-1              Display: tk-2    Display: tk-3
(or tk-1-abc if collision) (or tk-2-def)
```

**Identity**: task_uid (immutable ULID)
**Display**: alias-number (derived, e.g., `tk-1`) or alias-number-nodehint (e.g., `tk-1-abc`) if collision
**Numbering**: Label per project — may collide, resolved at display time

---

## What Changes in Practice

### Task Creation

#### v1/v2
```bash
tk prefix create foo "Foo project"  # Create organizational unit
tk new --prefix foo "Fix bug"       # Create task
# Result: foo-1-abc123               # Task ID stored as-is
```

#### v4
```bash
tk project create "Foo project"              # Create project → prj_01J5Q...
tk project alias foo prj_01J5Q...            # Create alias on this node
tk new --project foo "Fix bug"               # Create task → tsk_01J5Q...
# Result: foo-1                               # Display string (derived)
# (or foo-1-abc if another node also has foo-1)
```

---

### Task Movement / Renumbering

#### v1/v2
```bash
tk mv foo-1 bar:2
# Emits: task.reprefix {old_prefix: foo, new_prefix: bar, old_number: 1, new_number: 2}
# Result: Task ID changes to bar-2-abc123
# Alias: foo-1-abc123 → bar-2-abc123 (preserved)
```

#### v4
```bash
# Change project (move to different project):
tk new --project bar "..."  # Would need to create task in target project
# (No "move" between projects in v4; projects are distinct)

# Change number within same project:
tk number set foo-1 2
# Emits: task.number.set {task_uid, project_uid, number: 2}
# Result: Display changes from foo-1 to foo-2
# The old number foo-1 becomes available again for new tasks
```

---

### Collision Scenarios

#### v1/v2: Guaranteed Unique per Node
```
Node A: Creates foo-1-nodeA
Node B: Creates foo-1-nodeB
After sync:
  - Both exist as foo-1-nodeA and foo-1-nodeB
  - No collision (node suffix ensures uniqueness)
  - Display always shows full node suffix
```

#### v4: Collisions Allowed and Resolved
```
Node A: Creates task → tsk_A, assigns number 1 in project prj_foo
Node B: Creates task → tsk_B, assigns number 1 in project prj_foo
After sync:
  task_numbers table:
    (prj_foo, 1, tsk_A)  ← Both entries exist
    (prj_foo, 1, tsk_B)

  Display:
    foo-1-abc  ← Node A's task (with node hint)
    foo-1-def  ← Node B's task (with node hint)

  Resolution (optional):
    tk number set foo-1-def 2  → Renumber one task
    Result: foo-1 and foo-2 (no more collision)
```

---

### Alias Collisions (New in v4)

#### Scenario: Two Nodes Use Same Alias Name

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

## Data Model Comparison

### Database Tables

#### v1/v2
```sql
events (id, ts, actor, role, kind, payload, ...)
prefixes (prefix, node, description, created_at, created_by, removed)
prefix_counters (prefix, node, last_id)
task_counter (last_id)  -- legacy
```

#### v4
```sql
events (id, ts, actor, role, kind, payload, ...)
projects (project_uid PK, type, origin_json, name, description, created_at, created_by)
project_aliases (project_uid, alias, node, added_by, PRIMARY KEY(alias, node))
tasks (task_uid PK, project_uid, created_node, title, created_at, created_by, ...)
task_numbers (project_uid, number, task_uid)  -- Note: NOT unique!
```

---

### Event Schema

#### v1/v2
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

#### v4
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
  "kind": "task.number.set",
  "payload": {
    "task_uid": "tsk_01J5Q...",
    "project_uid": "prj_01J5Q...",
    "number": 2,
    "reason": "renumbered from 1"
  }
}
```

---

## CLI Command Mapping

| v1/v2 | v4 | Notes |
|-------|-----|-------|
| `tk prefix create foo "desc"` | `tk project create "desc"`<br>`tk project alias foo <uid>` | Two commands in v4 |
| `tk prefix list` | `tk project list` | Similar output |
| `tk prefix list --all` | `tk project list --all` | Shows all nodes' projects |
| `tk prefix remove foo` | `tk project alias remove foo` | Just removes alias |
| `tk new --prefix foo "title"` | `tk new --project foo "title"` | `--project` instead of `--prefix` |
| `tk mv foo-1 bar:2` | *No direct equivalent*<br>`tk number set foo-1 2` | v4: renumber only, no project change |
| `tk ls --prefix foo` | `tk ls --project foo` | `--project` instead of `--prefix` |
| `tk view foo-1-abc123` | `tk view foo-1`<br>`tk view foo-1-abc` | Node hint only if needed |
| `tk id foo-1` | `tk id foo-1` | Shows task_uid + current render |

---

## Migration Example

### Before Migration (v1/v2 state)
```
Prefixes:
  tk (node: abc123, desc: "Personal tasks")
  work (node: abc123, desc: "Work tasks")

Tasks:
  tk-1-abc123 → task_uuid_1 "Write docs"
  tk-2-abc123 → task_uuid_2 "Fix bug"
  work-1-abc123 → task_uuid_3 "Review PR"
```

### After Migration (v4 state)
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

---

## Benefits of v4 Model

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

## Key Takeaways

1. **Identity moves from IDs to UIDs**
   - v1/v2: Task ID is `prefix-number-node` (partially mutable)
   - v4: Task UID is `tsk_...` (immutable), display is derived

2. **Organizations move from prefixes to projects**
   - v1/v2: Prefix is a string, scoped to node
   - v4: Project is a UID with metadata, aliases are scoped to node

3. **Numbers become labels**
   - v1/v2: Numbers are counters (sequential, unique per node)
   - v4: Numbers are labels (assigned, may collide, displayed with hints)

4. **Collisions are features, not bugs**
   - v4 embraces offline conflicts
   - Display layer handles disambiguation
   - User can resolve conflicts later with `tk number set`

5. **Migration is substantial**
   - Not just a CLI rename
   - Fundamental data model change
   - Recommended: Hard Break with migration tool

---

## Next Steps

1. **Read** [V4_MIGRATION_SUMMARY.md](./V4_MIGRATION_SUMMARY.md) for executive summary
2. **Decide** on critical options in [V4_MIGRATION_DECISIONS.md](./V4_MIGRATION_DECISIONS.md)
3. **Review** [MIGRATION_V4_STRATEGY.md](./MIGRATION_V4_STRATEGY.md) for technical details
4. **Let me know** your decisions, and I'll implement!
