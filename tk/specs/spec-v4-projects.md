# tk spec v4

unified project + task model (numbers as labels)

---

# 1 overview

tk is an event-sourced task tracker.
Every change is an immutable event; the current state is a projection.
Spec v4 defines projects as first-class entities and treats task numbers as mutable labels, not identities.

---

# 2 core entities

## 2.1 Project

| field         | meaning                                                                                   |
|---------------|-------------------------------------------------------------------------------------------|
| project_uid   | stable ULID identifier (prj_…)                                                            |
| type          | `local`                                                                                   |
| origin        | structured metadata ({"owner":"neongreen","repo":"mono"} for GitHub)                      |
| name          | human-readable name                                                                       |
| description   | text                                                                                      |
| created_at    | timestamp                                                                                 |
| created_by    | actor                                                                                     |

## 2.2 Project alias

| field         | meaning                                                                                   |
|---------------|-------------------------------------------------------------------------------------------|
| alias         | short string used in CLI (tk, backend, mono, …)                                           |
| node          | node where alias is defined                                                               |
| project_uid   | link to project                                                                           |
| added_by      | actor                                                                                     |

Aliases are per-node and may collide freely.

---

## 2.3 Task

| field         | meaning                                                                                   |
|---------------|-------------------------------------------------------------------------------------------|
| task_uid      | stable ULID (tsk_…) — true identity                                                       |
| project_uid   | owning project                                                                            |
| created_node  | origin node id                                                                            |
| title         | text                                                                                      |
| created_by    | actor                                                                                     |
| created_at    | timestamp                                                                                 |

## 2.4 Task number (label)

| field         | meaning                                                                                   |
|---------------|-------------------------------------------------------------------------------------------|
| project_uid   | scope                                                                                     |
| number        | integer label (not unique)                                                                |
| task_uid      | target                                                                                    |

Multiple tasks may share the same (project_uid, number); collisions are resolved in display only.

---

# 3 identifiers & rendering

| form                 | meaning                                  | example          |
|----------------------|------------------------------------------|------------------|
| tsk_01J5Q…           | stable internal id                       | always unique    |
| foo-7                | alias + number                           | may refer to 1 or many tasks |
| foo-7-abc123         | disambiguated with node hint           | render when collision exists |
| owner/repo#27        | renderer for external project            | GitHub example   |

Display strings are derived views, never stored in events as IDs.

---

# 4 events

| kind                 | essential fields                                                     | meaning                                   |
|----------------------|----------------------------------------------------------------------|-------------------------------------------|
| project.created      | project_uid,type,origin?,name,description,created_by                | declare project                           |
| project.alias.add    | project_uid,alias,node,added_by                                      | define local alias                        |
| project.alias.remove | project_uid,alias,node                                               | remove alias                              |
| task.created         | task_uid,project_uid,proposed_number?,created_node,title,created_by | create task                               |
| task.number.set      | task_uid,project_uid,number,reason                                   | assign or change label                    |
| task.status.set, task.note.add, etc. | —                                                                    | unchanged from earlier spec               |

All events sync across nodes symmetrically.

---

# 5 creation semantics

```
tk new --project <selector> "title"
```

1.  Resolve <selector> → project_uid via alias or explicit uid.
2.  Compute proposed_number = max_seen + 1 (best-effort).
3.  Emit task.created { task_uid, project_uid, proposed_number, … }.
4.  If later a collision is detected, emit task.number.set to renumber (optionally automatic).

Offline nodes may choose the same number; after sync these coexist as distinct tasks.

---

# 6 renumbering & prefix/project changes

*   tk number set <task_ref> <N> → emit task.number.set event.
*   Projects may be renamed or merged without affecting task_uids.
*   After a merge, numbers can be re-assigned in bulk for tidiness.

---

# 7 lookups

Resolution priority:
1.  Exact task_uid.
2.  <alias>-<number> → resolve to tasks within that project.
3.  If multiple tasks share number, show list with node hints.
4.  Numeric alone → ambiguous error with suggestions.

---

# 8 projections (SQL sketch)

```sql
projects (project_uid PK,…)
project_aliases (project_uid,alias,node PK)
tasks (task_uid PK,project_uid,created_node,title,…)
task_numbers (project_uid,number,task_uid) — non-unique
```

---

# 9 renderers

Renderer selects string form by type and origin.

| type      | rule                 | example      |
|-----------|----------------------|--------------|
| local     | <alias>-<number>     | tk-1         |
| local collision | <alias>-<number>-<node> | tk-1-def456  |
| github    | <owner>/<repo>#<number> | neongreen/mono#27 |
| linear    | <team_key>-<number>  | DEV-42       |

---

# 10 sync semantics

All events are replicated verbatim.
Every node eventually learns every project, task, and number assignment.
Collisions in numbers are permitted; UI renders with node hints until renumbered.

---

# 11 commands (summary)

| command                               | effect                                  |
|---------------------------------------|-----------------------------------------|
| tk project create <name> [desc]       | emit project.created                    |
| tk project alias <name> <project>     | emit project.alias.add                  |
| tk new --project <selector> "title"   | emit task.created                       |
| tk number set <task_ref> <N>          | emit task.number.set                    |
| tk ls [--project x]                   | list tasks by renderer strings          |
| tk id <ref>                           | show task_uid, current render           |
| tk events …                           | inspect raw events                      |

---

# 12 external projects

External integrations (Linear, GitHub, Jira) map their issue numbers onto task.number.set.
task_uid is always the internal anchor; remotes are pure views.
Creating tasks “into” a remote project without rights is allowed locally; push may reject and record a sync error event.

---

# 13 authority model

tk core does not enforce permissions.
Any actor may emit events for any known project.
External remotes apply their own authorization.

---

# 14 example timeline

(node A) tk new → task_uid A1 → label tk-1
(node B) tk new → task_uid B1 → label tk-1
(sync)
tk ls → tk-1-A, tk-1-B
(node B) tk number set B1 2
tk ls → tk-1, tk-2

---

# 15 invariants

*   task_uid is the only immutable identifier.
*   Labels (number, alias) may collide and change.
*   All events are append-only and idempotent.
*   Any string form can be regenerated from events.

---

# 16 summary for implementors

*   Store and sync immutable UID-based events.
*   Treat numbers and aliases as mutable labels.
*   Render disambiguated forms when needed.
*   Allow renumber, merge freely.
*   Identity = task_uid, not prefix-number-node.

Result: a simple distributed model that remains fully offline-safe,
permits collisions and renumbering, and keeps user-facing labels pleasant and human.