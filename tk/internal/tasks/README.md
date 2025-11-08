# internal/tasks

This package contains extracted business logic for task operations, separated from CLI concerns.

## Purpose

Commands in `cmd/` should be thin wrappers that:
1. Parse arguments and flags
2. Open database connection
3. Call operations from this package
4. Format and display output

Business logic for task operations lives here, making it:
- Testable without CLI machinery
- Reusable across different interfaces (CLI, API, TUI)
- Easier to reason about and maintain

## What Belongs Here

**Task operations** that modify task state through events:
- Creating tasks
- Editing task fields (title, status, number)
- Moving tasks between projects
- Adding notes
- Deleting tasks
- Setting metadata

**Pure business logic** for task operations:
- Validation of operation parameters
- Computing derived values (e.g., next available number)
- Collision detection
- State transitions

## What Does NOT Belong Here

- Database schema definitions → `internal/database`
- Event projection/reduction logic → `internal/reducer`
- CRDT merge logic → `internal/reducer`, `internal/relations`
- CLI argument parsing → `cmd/`
- Display/formatting → `cmd/`, `internal/display`
- Configuration → `internal/config`

## Current Files

### edit.go
- `EditField(db, taskUID, field, value, actor)` - Edit any task field
- `EditTitle(db, taskUID, title, actor)` - Set task title
- `EditStatus(db, taskUID, status, actor)` - Set task status on generic axis
- `EditNumber(db, taskUID, number, actor)` - Set task number

### mark.go
- `Mark(db, taskUID, opts, actor)` - Set or unset task status on any axis
- `MarkOptions` - Options for status changes (axis, state, role)

### move.go
- `Move(db, taskUID, toProjectUID, opts, actor)` - Relocate task to different project
- `MoveOptions` - Options for move (keep/auto/force number, collision handling)
- **Key feature**: Resolves "auto" mode deterministically at event creation time

### helpers.go
- `GetProjectForTask(db, taskUID)` - Get project UID for a task
- `GetProjectAndNumberForTask(db, taskUID)` - Get project UID and number
- `CheckNumberCollision(db, projectUID, number, excludeTaskUID)` - Check if number exists

## Design Principles

1. **Event-sourced operations**: All state changes go through events
2. **Deterministic**: Same inputs produce same events (especially for "auto" modes)
3. **No I/O mixing**: Operations don't print output or read stdin
4. **Return errors**: Let callers decide how to handle/display errors
5. **Accept UUIDs**: Operations work with UUIDs, let callers do resolution
6. **Accept actor**: Caller provides actor string (don't call GetCurrentUser here)

## Testing Strategy

Operations in this package should have:
- Unit tests that don't require real database (use temp DB)
- Tests for edge cases and error conditions
- Tests for determinism (same inputs → same events)
- Tests independent of CLI

See: tk-185 (unit tests for these operations)

## History

This package was created as part of refactoring to improve testability:
- Initial extraction: Commit c45affbe (Claude, 12 hours ago)
  - Moved edit, mark, mv operations from cmd/ to internal/tasks/
  - Reduced cmd files by 25-71%
  - Created internal/query/ for filtering logic

Related tasks:
- tk-161: Overall testability refactoring epic
- tk-171: Extract business logic from commands (parent task for this work)
- tk-173: Add reducer tests (CRDT/distributed systems logic - high priority)
- tk-182, tk-183, tk-184: Extract remaining commands (new, note, rm)
- tk-185: Add unit tests for operations in this package

## What's Next

Commands still needing extraction:
- `cmd/new.go` → `tasks.Create()`
- `cmd/note.go` → `tasks.AddNote()`
- `cmd/rm.go` → `tasks.Delete()`
- `cmd/describe.go` - already uses `tasks.EditTitle()` but could be cleaner
- Project operations → potential `internal/projects/` package

Once extraction is complete, add comprehensive unit tests (tk-185).

