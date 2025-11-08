# database/ vs tasks/ Package Boundaries

## Current State (Unclear)

**database/** contains mix of:
- Schema, migrations, DB operations
- Business logic (CreateTask, ResolveTaskReference)
- Projection logic (ProjectTaskCreatedEvent, etc.)

**tasks/** contains:
- Business logic for edit, mark, move operations
- Extracted from cmd/ for testability

## The Confusion

- `database.CreateTask` exists - should it move to `tasks.Create`?
- Where does resolution logic belong?
- Where do projection methods belong?

## Proposed Boundary

### database/ should contain:
- DB schema and migrations
- Low-level DB operations (InsertEvent, QueryRow wrappers)
- **Projection/reduction** (ProjectTaskCreatedEvent - applies events to DB state)
- Node ID management
- Lamport timestamp management

### tasks/ should contain:
- **Business logic** for task operations
- Event creation (what events to emit, in what order)
- Validation, collision detection
- Computing derived values
- NO projection logic (that's database's job)

### Resolution: Where does it belong?

**Option A**: Resolution is business logic → move to tasks/ or separate package
**Option B**: Resolution is DB query logic → keep in database/

Currently in database/task_resolver.go. Needs decision.

## Action Items

1. Decide: Move CreateTask from database → tasks, or keep in database?
2. If moving: Update all callers, remove database/tasks.go
3. Document decision in both package READMEs
4. Consider: Should resolution be separate package? (internal/resolution/)

Related: tk-171, tk-182

