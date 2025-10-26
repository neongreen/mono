# V4 Implementation Review - Issue List

**Branch**: `copilot/implement-v4-spec-migration`
**Reviewer**: Claude Code
**Date**: 2025-10-26
**Status**: Requires Major Changes Before Merge

---

## Executive Summary

The v4 implementation demonstrates good understanding of the spec and includes ~2000 lines of code with solid foundational work. However, there are **4 critical** and **3 major** architectural issues that must be resolved before this can be merged.

**Overall Completion**: ~60%
**Recommendation**: Request changes and re-review after fixes

---

## 🔴 CRITICAL ISSUES (Must Fix Before Merge)

### Issue #1: Event Sourcing Architecture Violated

**Severity**: 🔴 Critical
**Files**: `main.go:83-192`, `project_cmd.go:305-366`, `v4_migration.go:305-337`

#### Problem

Events are being written directly to projection tables (projects, tasks, task_numbers) without first being recorded in the `events` table. This breaks the fundamental event sourcing architecture.

**Example violations**:

```go
// main.go:102-107
if err := db.emitTaskCreatedV4Event(string(taskUID), projectUID, proposedNumber, nodeID, title, currentUser, event.CreatedAt.Unix()); err != nil {
    return fmt.Errorf("failed to project task: %w", err)
}
```

```go
// v4_migration.go:305-312
func (d *DB) emitProjectCreatedEvent(projectUID, typ, name, description, createdBy string, createdAt int64) error {
    // Creates and inserts the event
    // For now, we'll add to events table and project the data into projects table
    _, err := d.db.Exec(`
        INSERT INTO projects (project_uid, type, name, description, created_at, created_by)
        VALUES (?, ?, ?, ?, ?, ?)
    `, projectUID, typ, name, description, createdAt, createdBy)
    return err
}
```

#### Impact

- ❌ No append-only event log
- ❌ Cannot replay events
- ❌ Breaks event sourcing guarantees
- ❌ No sync support (events won't propagate to other nodes)
- ❌ No audit trail
- ❌ Cannot troubleshoot issues via event inspection

#### Required Fix

1. **Always write to events table first**:
   ```go
   event := Event{
       ID:        generateEventID(),
       TS:        getNextLamportTimestamp(db),
       CreatedAt: time.Now(),
       Actor:     actor,
       Role:      role,
       Kind:      "task.created",
       Payload:   payloadJSON,
   }

   if err := db.InsertEvent(event); err != nil {
       return err
   }
   ```

2. **Then project from events into tables**:
   - Either: Call reducer to update in-memory state, then persist
   - Or: Have trigger/projection function that reads events and updates tables

3. **Fix all locations**:
   - `createTaskV4()` in main.go
   - `projectCreateCmd.RunE` in project_cmd.go
   - All `emit*Event()` functions in v4_migration.go

#### Acceptance Criteria

- [ ] All v4 events are written to events table with proper ID, TS, payload
- [ ] Projection tables are updated FROM events, not bypassing them
- [ ] Can query events table and see all project.created, task.created, etc. events
- [ ] Event log is complete and can be replayed

#### Mandatory Tests

```go
// Test that events are written to events table
func TestV4EventsWrittenToEventLog(t *testing.T) {
    // Create project
    // Query events table for project.created event
    // Verify event exists with correct payload
}

// Test that projections match events
func TestV4ProjectionsMatchEvents(t *testing.T) {
    // Clear projection tables
    // Replay all events through reducer
    // Verify projection tables are correctly populated
}

// Test event replay
func TestV4EventReplay(t *testing.T) {
    // Create projects and tasks
    // Save event log
    // Clear projection tables
    // Replay events
    // Verify state is identical
}
```

---

### Issue #2: Incomplete Migration - Data Loss Risk

**Severity**: 🔴 Critical
**Files**: `v4_migration.go:150-175`, `v4_migration.go:223-301`

#### Problem

The migration only processes `task.created` events but ignores all other event types. This causes **permanent data loss** during migration.

```go
// v4_migration.go:227-230
rows, err := d.db.Query(`
    SELECT id, ts, created_at, actor, payload
    FROM events
    WHERE kind = 'task.created'  // ❌ ONLY migrating task.created!
    ORDER BY ts, id
`)
```

**Lost data**:
- ❌ All `task.status.set` events (status history lost)
- ❌ All `task.note.add` events (notes lost)
- ❌ All `task.reprefix` events (not properly handled)
- ❌ All `task.alias.added` events
- ❌ All `relation.add/remove` events (if present)

#### Impact

Users will lose:
- Complete status change history
- All task notes and comments
- Task movement history
- Any task aliases they created
- Relationships between tasks

#### Required Fix

1. **Migrate ALL legacy events**, not just task.created:

```go
func (d *DB) migrateAllEventsToV4(nodeID string, actor string) error {
    // Query ALL events
    rows, err := d.db.Query(`
        SELECT id, ts, created_at, actor, role, kind, payload
        FROM events
        ORDER BY ts, id
    `)

    for rows.Next() {
        switch event.Kind {
        case "task.created":
            // Convert to v4 task.created
        case "task.status.set":
            // Already compatible, just verify task_uid mapping
        case "task.note.add":
            // Already compatible, just verify task_uid mapping
        case "task.reprefix":
            // Convert to task.relocate event
        case "task.alias.added":
            // Keep or convert as needed
        case "prefix.created":
            // Already handled
        }
    }
}
```

2. **Create task_uuid → task_uid mapping table** during migration to preserve references

3. **Emit proper v4 equivalents** for legacy events:
   - `task.reprefix` → `task.relocate` with proper number_policy
   - Preserve all other events with updated references

#### Acceptance Criteria

- [ ] Migration processes ALL event types from v1/v2
- [ ] Status history is preserved
- [ ] Notes are preserved
- [ ] Task movement history (reprefix → relocate) is preserved
- [ ] No data loss warnings or errors
- [ ] Event count before migration ≈ event count after migration

#### Mandatory Tests

```go
// Test migration preserves status
func TestV4MigrationPreservesStatus(t *testing.T) {
    // Create v1/v2 DB with task + status changes
    // Migrate to v4
    // Verify all status.set events are preserved
    // Verify status history is intact
}

// Test migration preserves notes
func TestV4MigrationPreservesNotes(t *testing.T) {
    // Create v1/v2 DB with task + notes
    // Migrate to v4
    // Verify all note.add events are preserved
    // Verify notes are retrievable
}

// Test migration handles reprefix
func TestV4MigrationConvertsReprefix(t *testing.T) {
    // Create v1/v2 DB with task.reprefix event
    // Migrate to v4
    // Verify task.relocate event was created
    // Verify old prefix → new project mapping is correct
}

// Test migration event count
func TestV4MigrationEventCount(t *testing.T) {
    // Create v1/v2 DB with N events
    // Migrate to v4
    // Verify events table has >= N events (may add some for backfill)
    // Verify no events were lost
}
```

---

### Issue #3: Missing Core Commands

**Severity**: 🔴 Critical
**Files**: New files needed

#### Problem

Per spec `tk/specs/v4.md:258-270`, several essential commands are not implemented:

| Command | Required | Status |
|---------|----------|--------|
| `tk edit <task> number <N>` | ✅ Yes | ❌ Missing |
| `tk edit <task> title "..."` | ✅ Yes | ❌ Missing |
| `tk mv <task> <target> [--keep\|--auto\|--force N]` | ✅ Yes | ❌ Missing |
| `tk ls --project x` | ✅ Yes | ❌ Missing |
| `tk id <ref>` | ⚠️ Important | ❌ Missing |
| `tk doctor` | ⚠️ Important | ❌ Missing |
| `tk conflicts numbers --project x` | ⚠️ Important | ❌ Missing |

#### Impact

Without these commands:
- ❌ Cannot renumber tasks
- ❌ Cannot edit task titles in v4
- ❌ Cannot move tasks between projects
- ❌ Cannot filter task lists by project
- ❌ Cannot verify migration health
- ❌ Cannot detect/resolve collisions

This makes v4 **unusable in practice**.

#### Required Fix

Implement all missing commands. Minimum requirements:

**1. `tk edit <task> <field> <value>`**

```go
var editCmd = &cobra.Command{
    Use:   "edit <task> <field> <value>",
    Short: "Edit a task field",
    Args:  cobra.ExactArgs(3),
    RunE: func(cmd *cobra.Command, args []string) error {
        taskRef := args[0]
        field := args[1]
        value := args[2]

        // Resolve taskRef → task_uid
        taskUID := resolveTaskReference(taskRef)

        switch field {
        case "number":
            // Emit task.number.set event
        case "title":
            // Emit task.title.set event
        case "status":
            // Emit task.status.set event
        }
    },
}
```

**2. `tk mv <task> <target> [flags]`**

```go
var mvCmd = &cobra.Command{
    Use:   "mv <task> <target-project>",
    Short: "Move task to different project",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Emit task.relocate event with number_policy
    },
}
mvCmd.Flags().String("keep", "", "Keep current number")
mvCmd.Flags().String("auto", "", "Auto-assign new number")
mvCmd.Flags().Int64("force", 0, "Force specific number")
```

**3. `tk ls --project <alias>`**

Update existing ls command to filter by project.

**4. `tk doctor`**

```go
var doctorCmd = &cobra.Command{
    Use:   "doctor",
    Short: "Verify database health",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Check all tasks have valid projects
        // Check all aliases resolve
        // Report collisions
        // Verify event log integrity
    },
}
```

#### Acceptance Criteria

- [ ] `tk edit <task> number N` successfully renumbers task
- [ ] `tk edit <task> title "..."` updates title
- [ ] `tk mv <task> <project>` moves task and handles numbering
- [ ] `tk ls --project foo` shows only tasks in that project
- [ ] `tk doctor` reports health status and collisions
- [ ] All commands emit proper v4 events to event log

#### Mandatory Tests

```go
// Test edit number command
func TestEditNumberCommand(t *testing.T) {
    // Create task with number 1
    // Run: tk edit <task> number 2
    // Verify task.number.set event was emitted
    // Verify task now displays as <project>-2
}

// Test edit title command
func TestEditTitleCommand(t *testing.T) {
    // Create task
    // Run: tk edit <task> title "New Title"
    // Verify task.title.set event was emitted
    // Verify task shows new title
}

// Test mv command
func TestMvCommand(t *testing.T) {
    // Create two projects
    // Create task in project A
    // Run: tk mv <task> projectB --auto
    // Verify task.relocate event was emitted
    // Verify task moved to project B
}

// Test ls --project filter
func TestLsProjectFilter(t *testing.T) {
    // Create two projects with tasks
    // Run: tk ls --project foo
    // Verify only foo tasks are shown
}

// Test doctor command
func TestDoctorCommand(t *testing.T) {
    // Create DB with known issues (orphaned task, collision)
    // Run: tk doctor
    // Verify all issues are reported
}
```

---

### Issue #4: Task Lookup Not Implemented

**Severity**: 🔴 Critical
**Files**: New file needed (e.g., `task_resolver.go`)

#### Problem

Per spec `tk/specs/v4.md:211-218`, task lookup should work with multiple formats:
1. Exact `task_uid` (e.g., `tsk_01J5Q...`)
2. `<alias>-<number>` (e.g., `tk-1`)
3. `<alias>-<number>-<node_hint>` (e.g., `tk-1-abc123`)
4. Numeric alone should error

**Currently**: No task resolution logic exists. All commands that take `<task>` argument will fail.

#### Impact

- ❌ Cannot reference tasks in commands
- ❌ `tk edit`, `tk mv`, `tk view` won't work
- ❌ Display IDs are not rendered with node hints on collision

#### Required Fix

Implement task resolver:

```go
// task_resolver.go

// ResolveTaskReference resolves a user-provided task reference to a task_uid
func ResolveTaskReference(db *DB, ref string) (taskUID string, err error) {
    // 1. Check if it's a direct task_uid
    if strings.HasPrefix(ref, "tsk_") {
        // Verify it exists
        return ref, nil
    }

    // 2. Parse as display ID: <alias>-<number>[-<node_hint>]
    alias, number, nodeHint, err := parseDisplayID(ref)
    if err != nil {
        return "", fmt.Errorf("invalid task reference: %w", err)
    }

    // 3. Resolve alias → project_uid
    projectUID, err := resolveAlias(db, alias)
    if err != nil {
        return "", err
    }

    // 4. Query task_numbers for matches
    rows, err := db.db.Query(`
        SELECT task_uid FROM task_numbers
        WHERE project_uid = ? AND number = ?
    `, projectUID, number)

    taskUIDs := []string{}
    for rows.Next() {
        var uid string
        rows.Scan(&uid)
        taskUIDs = append(taskUIDs, uid)
    }

    // 5. Handle results
    if len(taskUIDs) == 0 {
        return "", fmt.Errorf("task %s not found", ref)
    }

    if len(taskUIDs) == 1 {
        return taskUIDs[0], nil
    }

    // 6. Multiple tasks (collision) - need node hint
    if nodeHint != "" {
        // Filter by node hint
        return filterByNodeHint(db, taskUIDs, nodeHint)
    }

    // 7. Ambiguous - show options
    return "", fmt.Errorf("ambiguous task reference %s, candidates:\n%s",
        ref, formatTaskList(db, taskUIDs))
}

// RenderTaskDisplayID renders a task as a display string
func RenderTaskDisplayID(db *DB, taskUID string) (string, error) {
    // 1. Get task's project_uid and number
    var projectUID string
    var number int64
    err := db.db.QueryRow(`
        SELECT tn.project_uid, tn.number
        FROM task_numbers tn
        WHERE tn.task_uid = ?
    `, taskUID).Scan(&projectUID, &number)

    // 2. Get project alias (prefer current node)
    alias := getProjectAlias(db, projectUID)

    // 3. Check for collision
    hasCollision := checkNumberCollision(db, projectUID, number)

    if !hasCollision {
        return fmt.Sprintf("%s-%d", alias, number), nil
    }

    // 4. Render with node hint
    nodeHint := getTaskNodeHint(db, taskUID)
    return fmt.Sprintf("%s-%d-%s", alias, number, nodeHint), nil
}
```

#### Acceptance Criteria

- [ ] Can resolve `tsk_01J5Q...` → task
- [ ] Can resolve `tk-1` → task (unique number)
- [ ] Can resolve `tk-1-abc123` → task (with node hint)
- [ ] Returns error for ambiguous references with suggestions
- [ ] Display rendering includes node hint only when collision exists
- [ ] Numeric-only reference (e.g., `1`) returns helpful error

#### Mandatory Tests

```go
// Test resolve by task_uid
func TestResolveByTaskUID(t *testing.T) {
    // Create task with known UID
    // Resolve by full UID
    // Verify correct task returned
}

// Test resolve by display ID (unique)
func TestResolveByDisplayIDUnique(t *testing.T) {
    // Create project with alias "tk"
    // Create task with number 1
    // Resolve "tk-1"
    // Verify correct task returned
}

// Test resolve with collision
func TestResolveByDisplayIDCollision(t *testing.T) {
    // Create two tasks with same project + number
    // Resolve "tk-1" (ambiguous)
    // Verify error with suggestions
    // Resolve "tk-1-<node_hint>" (specific)
    // Verify correct task returned
}

// Test render display ID (no collision)
func TestRenderDisplayIDNoCollision(t *testing.T) {
    // Create task tk-1 (unique)
    // Render display ID
    // Verify returns "tk-1" (no node hint)
}

// Test render display ID (with collision)
func TestRenderDisplayIDWithCollision(t *testing.T) {
    // Create two tasks with same number
    // Render display IDs
    // Verify both include node hints: "tk-1-abc", "tk-1-def"
}
```

---

## 🔶 MAJOR ISSUES (High Priority)

### Issue #5: No Segment Versioning / Sync Support

**Severity**: 🔶 Major
**Files**: New files needed in segment reading/writing code

#### Problem

Per spec `tk/specs/v4.md:22-24`:
- remote segments should go to `<remote>/v4/`
- segment header must have `"spec_version": 4`
- ingest should reject non-v4 segments

**Current state**:
- ✅ `remote_subdir` metadata is set to "v4" during migration
- ❌ Segment writing doesn't use v4/ directory
- ❌ No spec_version in segment headers
- ❌ No version guard on segment ingest

#### Impact

- ❌ Cannot sync between v4 nodes
- ❌ Risk of ingesting v1/v2 segments (data corruption)
- ❌ No isolation between v4 and legacy data

#### Required Fix

1. **Update segment writer** to use v4/ directory:
   ```go
   func (s *SegmentWriter) Write() error {
       // Get remote_subdir from metadata
       subdir := db.GetMetadata("remote_subdir") // Should be "v4"

       // Write to <remote>/<subdir>/segment-<id>.zst
       path := filepath.Join(remotePath, subdir, filename)
   }
   ```

2. **Add spec_version to segment headers**:
   ```go
   type SegmentHeader struct {
       SpecVersion int    `json:"spec_version"` // Must be 4
       NodeID      string `json:"node_id"`
       // ... other fields
   }
   ```

3. **Add version guard on ingest**:
   ```go
   func (s *SegmentReader) Validate() error {
       if s.Header.SpecVersion != 4 {
           return fmt.Errorf("rejecting segment: spec_version %d (expected 4)",
               s.Header.SpecVersion)
       }
   }
   ```

4. **Only sync from v4/ directory**:
   ```go
   func (s *Syncer) FetchSegments() error {
       // Only look in <remote>/v4/
       subdir := db.GetMetadata("remote_subdir")
       segments := listSegments(filepath.Join(remotePath, subdir))
   }
   ```

#### Acceptance Criteria

- [ ] Segments written to `<remote>/v4/` directory
- [ ] All segments have `spec_version: 4` in header
- [ ] Reader rejects segments with spec_version != 4
- [ ] Sync only fetches from v4/ directory
- [ ] Legacy v1/v2 segments are ignored

#### Mandatory Tests

```go
// Test segment versioning
func TestSegmentVersioning(t *testing.T) {
    // Create v4 DB
    // Write segment
    // Verify segment is in v4/ directory
    // Verify segment header has spec_version: 4
}

// Test version guard
func TestSegmentVersionGuard(t *testing.T) {
    // Create v1/v2 segment (spec_version: 2)
    // Attempt to ingest
    // Verify rejection with error
}

// Test sync isolation
func TestSyncV4Isolation(t *testing.T) {
    // Create remote with v1/v2 and v4 segments
    // Sync from v4 DB
    // Verify only v4/ segments were fetched
    // Verify v1/v2 segments were ignored
}
```

---

### Issue #6: Missing `tk doctor` Command

**Severity**: 🔶 Major
**Files**: New file needed: `doctor_cmd.go`

#### Problem

Per spec `tk/specs/v4.md:283-287` and `tk/specs/v4-migration.md:413-424`, `tk doctor` should:
- Auto-run after migration
- Verify every task has a valid project
- Check all aliases resolve
- Report label collisions
- Verify event log integrity

**Current state**: Not implemented.

#### Impact

- ❌ Cannot verify migration success
- ❌ No way to detect database corruption
- ❌ Users won't know about collisions
- ❌ No automated health checks

#### Required Fix

Create `doctor_cmd.go`:

```go
var doctorCmd = &cobra.Command{
    Use:   "doctor",
    Short: "Verify database health and integrity",
    RunE: func(cmd *cobra.Command, args []string) error {
        db, err := openExistingDB()
        if err != nil {
            return err
        }
        defer db.Close()

        fmt.Println("Running health checks...")

        issues := []string{}

        // 1. Check all tasks have valid projects
        issues = append(issues, checkOrphanedTasks(db)...)

        // 2. Check all aliases resolve
        issues = append(issues, checkBrokenAliases(db)...)

        // 3. Report label collisions
        collisions := checkNumberCollisions(db)
        if len(collisions) > 0 {
            fmt.Printf("\nFound %d number collisions:\n", len(collisions))
            for _, c := range collisions {
                fmt.Printf("  %s: %v\n", c.DisplayID, c.TaskUIDs)
            }
        }

        // 4. Verify event log integrity
        issues = append(issues, checkEventLogIntegrity(db)...)

        // 5. Report results
        if len(issues) == 0 {
            fmt.Println("✓ All checks passed!")
            return nil
        }

        fmt.Printf("\n✗ Found %d issues:\n", len(issues))
        for _, issue := range issues {
            fmt.Printf("  - %s\n", issue)
        }
        return fmt.Errorf("health check failed")
    },
}

func checkOrphanedTasks(db *DB) []string {
    rows, _ := db.db.Query(`
        SELECT task_uid FROM tasks
        WHERE project_uid NOT IN (SELECT project_uid FROM projects)
    `)
    // Return list of orphaned tasks
}

func checkNumberCollisions(db *DB) []Collision {
    rows, _ := db.db.Query(`
        SELECT project_uid, number, COUNT(*) as cnt
        FROM task_numbers
        GROUP BY project_uid, number
        HAVING cnt > 1
    `)
    // Return list of collisions
}
```

Run automatically after migration in `main.go`:

```go
if needsMigration {
    // ... perform migration ...

    fmt.Println("\nRunning post-migration health check...")
    if err := runDoctor(db); err != nil {
        fmt.Printf("Warning: Health check found issues. Run 'tk doctor' for details.\n")
    }
}
```

#### Acceptance Criteria

- [ ] `tk doctor` command exists
- [ ] Detects orphaned tasks (tasks without valid projects)
- [ ] Detects broken aliases
- [ ] Reports number collisions with display IDs
- [ ] Verifies event log integrity (no gaps in TS, valid JSON)
- [ ] Auto-runs after migration
- [ ] Returns non-zero exit code on failures

#### Mandatory Tests

```go
// Test doctor detects orphaned tasks
func TestDoctorOrphanedTasks(t *testing.T) {
    // Create task with invalid project_uid
    // Run doctor
    // Verify orphaned task is reported
}

// Test doctor detects collisions
func TestDoctorCollisions(t *testing.T) {
    // Create two tasks with same project+number
    // Run doctor
    // Verify collision is reported with both task UIDs
}

// Test doctor on healthy DB
func TestDoctorHealthyDB(t *testing.T) {
    // Create valid DB
    // Run doctor
    // Verify returns success, no issues
}
```

---

### Issue #7: No Collision Detection or Rendering

**Severity**: 🔶 Major
**Files**: `task_resolver.go` (new), update `ls_cmd.go`, `view_cmd.go`

#### Problem

Per spec `tk/specs/v4.md:126-147`, the v4 model allows number collisions:
- Multiple tasks can have same (project_uid, number)
- Display should render with node hints when collision exists
- Users should be able to disambiguate with `<alias>-<number>-<node_hint>`

**Current state**:
- ❌ No collision detection
- ❌ No node hint rendering
- ❌ `tk ls` won't show disambiguated IDs

#### Impact

When two nodes create the same number offline:
- ❌ Both tasks will show as `tk-1` (users can't tell them apart)
- ❌ Referencing `tk-1` will be ambiguous (no error, undefined behavior)
- ❌ No way to select specific task

#### Required Fix

1. **Implement collision detection**:
   ```go
   func HasNumberCollision(db *DB, projectUID string, number int64) bool {
       var count int
       db.db.QueryRow(`
           SELECT COUNT(*) FROM task_numbers
           WHERE project_uid = ? AND number = ?
       `, projectUID, number).Scan(&count)
       return count > 1
   }
   ```

2. **Implement node hint rendering** (part of Issue #4):
   ```go
   func GetTaskNodeHint(db *DB, taskUID string) string {
       var createdNode string
       db.db.QueryRow(`
           SELECT created_node FROM tasks WHERE task_uid = ?
       `, taskUID).Scan(&createdNode)

       // Return last 6 chars for display
       if len(createdNode) > 6 {
           return createdNode[len(createdNode)-6:]
       }
       return createdNode
   }
   ```

3. **Update display rendering**:
   ```go
   func FormatTaskDisplayID(db *DB, taskUID string) string {
       projectUID, number := getTaskProjectAndNumber(db, taskUID)
       alias := getProjectAlias(db, projectUID)

       if HasNumberCollision(db, projectUID, number) {
           nodeHint := GetTaskNodeHint(db, taskUID)
           return fmt.Sprintf("%s-%d-%s", alias, number, nodeHint)
       }

       return fmt.Sprintf("%s-%d", alias, number)
   }
   ```

4. **Update `tk ls` to use proper rendering**

5. **Add `tk conflicts numbers` command** (mentioned in Issue #6)

#### Acceptance Criteria

- [ ] Can detect when multiple tasks share same (project, number)
- [ ] Display IDs include node hint when collision exists
- [ ] `tk ls` shows disambiguated IDs (e.g., `tk-1-abc`, `tk-1-def`)
- [ ] Can resolve `tk-1-abc` to specific task
- [ ] `tk conflicts numbers` lists all collisions

#### Mandatory Tests

```go
// Test collision detection
func TestCollisionDetection(t *testing.T) {
    // Create two tasks with same project+number
    // Check HasNumberCollision returns true
}

// Test node hint rendering
func TestNodeHintRendering(t *testing.T) {
    // Create collision
    // Render display IDs for both tasks
    // Verify both include node hints
    // Verify hints are different
}

// Test ls shows node hints
func TestLsShowsNodeHints(t *testing.T) {
    // Create collision
    // Run tk ls
    // Verify output shows both tasks with node hints
}
```

---

## 🟡 MINOR ISSUES (Should Fix)

### Issue #8: Inconsistent Error Handling

**Severity**: 🟡 Minor
**Files**: `project_cmd.go:365`

#### Problem

```go
func getNextLamportTimestamp(db *DB) int64 {
    var maxTS int64
    err := db.db.QueryRow("SELECT COALESCE(MAX(ts), 0) FROM events").Scan(&maxTS)
    if err != nil {
        return 1  // ❌ Silently returns 1 on error
    }
    return maxTS + 1
}
```

This silently swallows errors, which could lead to:
- Duplicate timestamps
- Event ordering issues
- Hard to debug problems

#### Required Fix

```go
func getNextLamportTimestamp(db *DB) (int64, error) {
    var maxTS int64
    err := db.db.QueryRow("SELECT COALESCE(MAX(ts), 0) FROM events").Scan(&maxTS)
    if err != nil {
        return 0, fmt.Errorf("failed to get max timestamp: %w", err)
    }
    return maxTS + 1, nil
}
```

Update all callers to handle the error.

#### Acceptance Criteria

- [ ] Function returns error instead of swallowing it
- [ ] All callers handle the error properly
- [ ] Database errors are surfaced to user

---

### Issue #9: Missing Input Validation

**Severity**: 🟡 Minor
**Files**: `main.go`, `project_cmd.go`

#### Problem

Functions don't validate required parameters:

```go
func createTaskV4(db *DB, cmd *cobra.Command, title string) error {
    projectFlag, _ := cmd.Flags().GetString("project")
    // No validation that projectFlag is not empty!
    // No validation that title is not empty!
```

#### Required Fix

Add validation:

```go
func createTaskV4(db *DB, cmd *cobra.Command, title string) error {
    if title == "" {
        return fmt.Errorf("title cannot be empty")
    }

    projectFlag, _ := cmd.Flags().GetString("project")
    if projectFlag == "" {
        return fmt.Errorf("--project flag is required")
    }

    // Validate alias format
    if err := Alias(projectFlag).Validate(); err != nil && !strings.HasPrefix(projectFlag, "prj_") {
        return fmt.Errorf("invalid project: %w", err)
    }

    // ... rest of function
}
```

Add similar validation for all commands.

#### Acceptance Criteria

- [ ] All required parameters are validated
- [ ] Empty strings are rejected
- [ ] Invalid formats are rejected with helpful errors
- [ ] Validation errors show usage information

---

### Issue #10: Code Duplication

**Severity**: 🟡 Minor
**Files**: `project_cmd.go:353-365`, `main.go`

#### Problem

Helper functions are duplicated across files:
- `generateEventID()` in project_cmd.go
- `getNextLamportTimestamp()` in project_cmd.go
- Similar logic in main.go

This leads to:
- Maintenance burden
- Risk of divergence
- Code bloat

#### Required Fix

Create `event_helpers.go`:

```go
package main

import (
    "database/sql"
    "fmt"
)

// GenerateEventID creates a new unique event ID
func GenerateEventID() EventID {
    return NewEventID()
}

// GetNextLamportTS gets the next Lamport timestamp from the database
func GetNextLamportTS(db *DB) (int64, error) {
    var maxTS int64
    err := db.db.QueryRow("SELECT COALESCE(MAX(ts), 0) FROM events").Scan(&maxTS)
    if err != nil {
        return 0, fmt.Errorf("failed to get max timestamp: %w", err)
    }
    return maxTS + 1, nil
}

// InsertEventAndProject inserts an event and projects it
func InsertEventAndProject(db *DB, event Event) error {
    // 1. Insert into events table
    if err := db.InsertEvent(event); err != nil {
        return fmt.Errorf("failed to insert event: %w", err)
    }

    // 2. Apply to reducer/projection
    if err := db.ProjectEvent(event); err != nil {
        return fmt.Errorf("failed to project event: %w", err)
    }

    return nil
}
```

Remove duplicates from other files and use these shared helpers.

#### Acceptance Criteria

- [ ] No duplicate helper functions
- [ ] All event operations use shared helpers
- [ ] Consistent behavior across all commands

---

## 📋 MANDATORY TEST REQUIREMENTS

### Test Coverage Requirements

All PRs must include tests. Minimum coverage by category:

#### Unit Tests (Required)

- [ ] **Type validation** (already exists, good)
  - All v4 types (ProjectUID, TaskUID, Alias, TaskNumber, etc.)
  - Edge cases (empty, invalid format, boundary values)

- [ ] **Event payload serialization** (NEW)
  - All v4 event payloads can be marshaled/unmarshaled
  - JSON structure matches spec

- [ ] **Helper functions** (NEW)
  - Event ID generation
  - Lamport timestamp sequencing
  - Task/project lookups

#### Integration Tests (Required)

- [ ] **End-to-end task creation** (NEW)
  ```go
  func TestE2ETaskCreation(t *testing.T) {
      // Create DB
      // Create project with alias
      // Create task
      // Verify:
      //   - Event in events table
      //   - Task in tasks table
      //   - Number in task_numbers table
      //   - Can list task
      //   - Can view task
  }
  ```

- [ ] **End-to-end migration** (expand existing)
  ```go
  func TestE2EMigration(t *testing.T) {
      // Create v1/v2 DB with:
      //   - Multiple prefixes
      //   - Tasks with status changes
      //   - Tasks with notes
      //   - Task movement (reprefix)
      // Migrate to v4
      // Verify:
      //   - All data preserved
      //   - All events migrated
      //   - Can use v4 commands
      //   - Old IDs still work
  }
  ```

- [ ] **Multi-node scenario** (NEW)
  ```go
  func TestMultiNodeCollision(t *testing.T) {
      // Simulate two nodes offline
      // Both create task with same project+number
      // Sync (when implemented)
      // Verify:
      //   - Both tasks exist
      //   - Collision detected
      //   - Display shows node hints
      //   - Can resolve both individually
  }
  ```

#### Command Tests (Required for each command)

Each command must have:

1. **Success case test**
2. **Error case tests** (invalid input, missing data)
3. **Edge case tests**

Example template:

```go
func TestCommandProjectCreate(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        // Run: tk project create "My Project" "Description"
        // Verify: project created, event logged
    })

    t.Run("with alias", func(t *testing.T) {
        // Run: tk project create "My Project" --alias mp
        // Verify: project and alias created
    })

    t.Run("error empty name", func(t *testing.T) {
        // Run: tk project create ""
        // Verify: returns error
    })
}
```

Required command test suites:
- [ ] `tk project create`
- [ ] `tk project list`
- [ ] `tk project alias add/remove`
- [ ] `tk new --project`
- [ ] `tk edit` (number, title, status)
- [ ] `tk mv`
- [ ] `tk ls --project`
- [ ] `tk doctor`
- [ ] `tk admin rollback-v4`

#### Migration Tests (Expand existing)

- [ ] **Empty database migration** (NEW)
- [ ] **Large database migration** (NEW - test performance)
  ```go
  func TestMigrationPerformance(t *testing.T) {
      // Create DB with 10,000 tasks
      // Measure migration time
      // Verify completeness
      // Should complete in <30 seconds
  }
  ```
- [ ] **Corrupted database handling** (NEW)
- [ ] **Rollback after migration** (exists, good)
- [ ] **Re-migration safety** (NEW - verify idempotency)

### Test Organization

Organize tests by file:

```
tk/
  v4_types_test.go       // Type validation tests
  v4_migration_test.go   // Migration tests (expand existing)
  v4_commands_test.go    // NEW: Command integration tests
  v4_resolver_test.go    // NEW: Task resolution tests
  v4_collision_test.go   // NEW: Collision handling tests
  v4_e2e_test.go         // NEW: End-to-end scenarios
```

### Minimum Coverage Targets

- **Overall**: 75% code coverage
- **v4_types.go**: 90% (critical path)
- **v4_migration.go**: 85% (critical path)
- **New commands**: 80%

---

## 📊 PRIORITY MATRIX

| Issue | Severity | Effort | Priority |
|-------|----------|--------|----------|
| #1: Event sourcing architecture | 🔴 Critical | High | **P0** - Fix first |
| #2: Incomplete migration | 🔴 Critical | High | **P0** - Fix first |
| #3: Missing core commands | 🔴 Critical | High | **P0** - Required for usability |
| #4: Task lookup | 🔴 Critical | Medium | **P0** - Blocks commands |
| #5: Segment versioning | 🔶 Major | Medium | **P1** - Before sync |
| #6: Doctor command | 🔶 Major | Low | **P1** - Quality of life |
| #7: Collision handling | 🔶 Major | Medium | **P1** - Part of spec |
| #8: Error handling | 🟡 Minor | Low | **P2** - Nice to have |
| #9: Input validation | 🟡 Minor | Low | **P2** - Nice to have |
| #10: Code duplication | 🟡 Minor | Low | **P2** - Cleanup |

---

## ✅ ACCEPTANCE CHECKLIST

Before requesting re-review, verify:

### Architecture
- [ ] All events written to events table first
- [ ] Projections built FROM events, not bypassing them
- [ ] Event log is complete and replayable
- [ ] No direct writes to projection tables in command handlers

### Migration
- [ ] ALL legacy event types are migrated (not just task.created)
- [ ] Status history preserved
- [ ] Notes preserved
- [ ] Movement history (reprefix → relocate) preserved
- [ ] Migration creates backup
- [ ] Rollback works correctly

### Commands
- [ ] `tk edit <task> number <N>` works
- [ ] `tk edit <task> title "..."` works
- [ ] `tk mv <task> <project>` works with number policies
- [ ] `tk ls --project <alias>` filters correctly
- [ ] `tk doctor` reports health
- [ ] All commands emit events to event log

### Task Resolution
- [ ] Can resolve by task_uid
- [ ] Can resolve by `<alias>-<number>`
- [ ] Can resolve by `<alias>-<number>-<node_hint>`
- [ ] Ambiguous references return helpful errors
- [ ] Display IDs rendered with node hints on collision

### Sync (when implemented)
- [ ] Segments written to `<remote>/v4/`
- [ ] Segments have `spec_version: 4`
- [ ] Ingest rejects non-v4 segments
- [ ] Only syncs from v4/ directory

### Tests
- [ ] All mandatory tests implemented (see section above)
- [ ] Tests pass reliably
- [ ] Coverage >= 75% overall
- [ ] Coverage >= 85% for migration code
- [ ] No flaky tests

### Documentation
- [ ] README updated (already done, good)
- [ ] Code comments for complex logic
- [ ] Migration guide accurate
- [ ] All new commands documented

---

## 📖 RESOURCES

- **Spec**: `tk/specs/v4.md`
- **Migration Spec**: `tk/specs/v4-migration.md`
- **Existing Code**: `tk/*.go` (v1/v2 commands for reference)

---

## 💬 QUESTIONS / CLARIFICATIONS NEEDED

If anything is unclear, please ask about:

1. Event sourcing pattern - should we use in-memory reducer or direct DB projection?
2. Sync implementation - is this in scope for this PR or separate?
3. Performance targets for large migrations?
4. Backward compatibility requirements (e.g., should v4 binary support reading v1/v2 events at all)?

---

**End of Issue List**

*This review was generated by Claude Code on 2025-10-26*
