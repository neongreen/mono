# Events

## project.created

Ensures projects exist in reducer state, replacing synthetic placeholders with real

metadata when the canonical creation event arrives.

*Source: `tk/internal/reducer/project.go:21`*

## project.alias.add

Aliases are tracked in projections; reducer ignores payload beyond validation to

keep replay idempotent while projections handle display IDs.

*Source: `tk/internal/reducer/project.go:27`*

## project.alias.remove

Alias removals are handled in projection tables, so reducer performs a lightweight

decode and otherwise no-ops for compatibility.

*Source: `tk/internal/reducer/project.go:33`*

## project.delete

Records a tombstone so later task events referencing the project create synthetic

placeholders instead of reviving deleted projects.

*Source: `tk/internal/reducer/project.go:39`*

## project.name.set

Updates project names in reducer state; if the project does not exist yet a

synthetic project is created so the rename is not lost.

*Source: `tk/internal/reducer/project.go:45`*

## task.created

Creates tasks with deterministic handling for duplicate creation events and

synthetic project creation when upstream data is incomplete.

*Source: `tk/internal/reducer/project.go:51`*

## task.number.set

Applies display numbers to tasks and updates alias maps so task lookups by ID stay

consistent with projection tables.

*Source: `tk/internal/reducer/project.go:57`*

## task.relocate

Moves tasks between projects, ensuring synthetic projects are created when needed

and relations are rebuilt using the reducer graph helpers.

*Source: `tk/internal/reducer/project.go:63`*

## task.title.set

Updates task titles in reducer state, preserving historical metadata and avoiding

extra allocations when titles repeat.

*Source: `tk/internal/reducer/project.go:69`*

## task.status.set

Applies a status update to the reducer task view, keeping history consistent with

projection tables while ignoring unknown statuses for forward compatibility.

*Source: `tk/internal/reducer/reducer.go:26`*

## task.note.add

Appends an immutable note entry to the reducer, keeping timestamps from the event

payload and leaving existing notes untouched.

*Source: `tk/internal/reducer/reducer.go:32`*

## task.delete

Marks tasks as deleted in-memory without removing historical data so later events

can no-op safely when rebuilding from the log.

*Source: `tk/internal/reducer/reducer.go:38`*

## task.meta.set

Updates reducer metadata claims and preserves competing values so resolution logic

stays deterministic with the projection layer.

*Source: `tk/internal/reducer/reducer.go:44`*

## relation.add

Adds a relation edge between tasks and recomputes the relations graph so blockers

and dependents stay in sync with later FinalizeRelations calls.

*Source: `tk/internal/reducer/reducer.go:50`*

## relation.remove

Removes a relation edge if present; missing edges are ignored to keep replay

idempotent when events arrive out of order.

*Source: `tk/internal/reducer/reducer.go:56`*

## relation.note

Attaches free-form notes to relations without altering the graph structure,

allowing multiple notes per relation across the log.

*Source: `tk/internal/reducer/reducer.go:62`*

## task.attachment.add

Tracks attachment metadata on the reducer task so CLI commands can surface

linked artifacts without querying projection tables.

*Source: `tk/internal/reducer/reducer.go:68`*

## task.attachment.remove

Removes attachment references if they exist; missing attachments are ignored to

keep log replay idempotent.

*Source: `tk/internal/reducer/reducer.go:74`*

