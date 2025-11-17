# internal/database

Database layer providing schema, connections, queries, and event projection.

## Purpose

This package owns the SQLite database and provides low-level operations. It does NOT contain business logic for task operations - that lives in `internal/tasks`.

## What Belongs Here

**Schema and migrations**:
- Table definitions
- Index creation
- Schema versioning

**Database connections and management**:
- Opening/closing DB
- Transaction handling
- Node ID management
- Lamport timestamp generation

**Read operations (queries)**:
- Task/project resolution (ResolveTaskReference, ResolveProjectRef)
- Rendering display IDs
- Querying task/project state
- Building projections

**Event projection methods**:
- `ProjectTaskCreatedEvent()` - applies task.created to tables
- `ProjectTaskTitleSetEvent()` - applies task.title.set to tables
- All other `Project*Event()` methods
- These are called BY internal/tasks operations

## What Does NOT Belong Here

- **Business logic** for task operations → `internal/tasks`
  - Creating events
  - Validation logic
  - Collision handling
  - Orchestrating multi-event operations
  
- **CRDT merge logic** → `internal/reducer`, `internal/relations`

- **CLI concerns** → `cmd/`

## Bright-Line Rule

**"Who creates the events?"**
- This package: Provides projection methods to apply events
- internal/tasks: Creates events and calls projection methods

**"Who queries the database?"**
- This package: All SQL queries live here
- Other packages: Call database functions, don't write SQL

## Related Work

- **tk-171**: Extracting operations from database/ to tasks/
- See `internal/tasks/README.md` for what belongs there

