# internal/tasks

Business logic for task operations, extracted from CLI commands for testability and reuse.

## Purpose

Commands in `cmd/` should be thin wrappers that parse arguments, call operations from this package, and format output. Business logic lives here to make it testable and reusable.

## What Belongs Here

Task operations that modify state through events:
- Creating, editing, moving, deleting tasks
- Setting status, title, number, metadata
- Adding notes
- Validation and collision detection
- Computing derived values (next available number, etc.)

## What Does NOT Belong Here

- Event projection/CRDT logic → `internal/reducer`, `internal/relations`
- Database schema → `internal/database`
- CLI parsing/formatting → `cmd/`, `internal/display`
- Configuration → `internal/config`

## Design Principles

1. **Event-sourced**: All changes go through events
2. **Deterministic**: Same inputs → same events (especially "auto" modes)
3. **No I/O**: Don't print output or read stdin
4. **Accept UUIDs**: Let callers do resolution
5. **Accept actor**: Caller provides actor string

## Related Work

- **tk-171**: Parent task for extracting business logic
- **tk-185**: Add unit tests for operations here
- **tk-182, tk-183, tk-184**: Extract remaining commands (new, note, rm)
- **History**: Initial extraction in commit c45affbe (edit, mark, move)

