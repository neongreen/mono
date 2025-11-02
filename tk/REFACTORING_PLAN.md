# TK Refactoring Plan: Clean Architecture

## Problems Identified

### 1. Helper Functions in Wrong Place
**cmd/** contains reusable building blocks that should be in **internal/**:

**Pure functions** (→ `internal/types`):
- `sortTasks()` - pure function on `[]*types.Task`
- `extractPrefix()` - task ID parsing

**Database operations** (→ `internal/database`):
- `getProjectAliasForTask()` - DB query
- `getAllProjectDisplayNames()` - DB query
- `createTask()` - task creation business logic

**System utilities** (→ `internal/utils` or keep existing):
- `getCurrentUser()` - OS/environment query

**Duplicates to remove**:
- `getNextLamportTimestamp()` in project.go - **DELETE** (use `db.GetNextLamportTS()`)
- `generateEventID()` in project.go - **DELETE** (use `database.GenerateEventID()`)

**Display/CLI-specific** (keep in cmd/ or move to internal/display):
- `colorizeStatus()` - CLI formatting
- `renderTaskTable()` - CLI display  
- `outputTasksJSON()` - CLI output

### 2. Command Reorganization

**Keep useful shortcuts** (user uses these frequently):
- describe, mark, note - keep as-is

**Move admin/debug commands** → consolidate under `tk debug`:
1. **admin.go** → rename to **debug.go**
2. **id.go** → move to `tk debug id`
3. **node.go** → move to `tk debug node`
4. **events.go** → move to `tk debug events`
5. **doctor.go** → move to `tk debug doctor`
6. **conflicts_numbers.go** → move to `tk debug numbers` or remove

**Commands to keep at top level:**
- Core: new, ls, edit, delete/rm, view, mv, describe, mark, note
- Projects: project
- Relations: relate, blockers, graph, conflicts
- Sync: sync, remote, pull, push, export, ingest, status
- Setup: init, db
- Debug: debug (with subcommands: doctor, events, id, node, fix-timestamps, rebuild-from-remote)

## Proposed Structure

```
tk/
├── main.go (package main)
├── cmd/ (package cmd)
│   ├── root.go
│   ├── blockers.go
│   ├── conflicts.go
│   ├── debug.go         # Renamed from admin.go, with subcommands
│   ├── delete.go
│   ├── describe.go      # KEEP - useful shortcut
│   ├── edit.go
│   ├── export.go
│   ├── graph.go
│   ├── ingest.go
│   ├── init.go
│   ├── ls.go
│   ├── mark.go          # KEEP - useful shortcut
│   ├── mv.go
│   ├── new.go
│   ├── note.go          # KEEP - useful shortcut
│   ├── path.go
│   ├── project.go
│   ├── relate.go
│   ├── remote.go
│   ├── status.go
│   ├── sync.go
│   ├── view.go
│   ├── display.go       # CLI-specific display functions
│   └── *_test.go
│
└── internal/
    ├── database/
    │   ├── db.go
    │   ├── projections.go
    │   ├── task_resolver.go
    │   ├── tasks.go          # NEW: createTask, etc.
    │   ├── projects.go       # NEW: project queries
    │   └── *_test.go
    ├── types/
    │   ├── task.go
    │   ├── task_sort.go      # NEW: sortTasks
    │   ├── task_id.go        # NEW: extractPrefix, parseTaskID
    │   └── ...
    └── utils/
        └── user.go            # getCurrentUser

```

## Implementation Steps

### Phase 1: Remove Duplicates
1. Delete `getNextLamportTimestamp()` from project.go
2. Delete `generateEventID()` from project.go
3. Update all callers to use `db.GetNextLamportTS()` and `database.GenerateEventID()`

### Phase 2: Move Pure Functions to internal/types
1. Create `internal/types/task_sort.go` with `SortTasks()`
2. Create `internal/types/task_id.go` with `ExtractPrefix()`
3. Update callers

### Phase 3: Move DB Operations to internal/database
1. Move `getProjectAliasForTask()` → `internal/database/projects.go`
2. Move `getAllProjectDisplayNames()` → `internal/database/projects.go`
3. Move `createTask()` → `internal/database/tasks.go`
4. Update callers

### Phase 4: Consolidate Display Functions
1. Rename `helpers_display.go` → `display.go`
2. Keep in cmd/ since CLI-specific
3. Export what needs to be exported

### Phase 5: Reorganize Debug Commands
1. Rename `cmd/admin.go` → `cmd/debug.go`
2. Move id, node, events, doctor commands as subcommands of debug
3. Delete or consolidate `cmd/conflicts_numbers.go` → `tk debug numbers`
4. Update root.go to register debug command with subcommands

### Phase 6: Rename helpers
1. `helpers_tasks.go` → delete (moved to internal/database)
2. `helpers_utils.go` → delete (moved to internal)
3. `helpers_display.go` → `display.go`
4. `helpers_debug.go` → `debug.go` or merge into admin

## Result

**cmd/** becomes:
- ~20 command files (down from 27, keeping shortcuts)
- debug.go with subcommands: doctor, events, id, node, fix-timestamps, rebuild-from-remote, numbers
- 1 display helper file (display.go, down from 5)
- Just Cobra glue code, no business logic

**internal/** gains:
- `internal/types/task_sort.go`
- `internal/types/task_id.go`
- `internal/database/tasks.go`
- `internal/database/projects.go`

Clean separation: cmd/ = CLI interface, internal/ = reusable logic

