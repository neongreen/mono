# Command Refactoring Plan

## Complete Old → New Command Mapping

### Core Task Management (unchanged)
```
tk new              → tk new              (cmd/new.go)
tk ls               → tk ls               (cmd/ls.go)
tk show             → tk show             (cmd/show.go)
tk mark             → tk mark             (cmd/mark.go)
tk edit             → tk edit             (cmd/edit.go)
tk describe         → tk describe         (cmd/describe.go)
tk note             → tk note             (cmd/note.go)
tk attach           → tk attach           (cmd/attach.go)
tk rm               → tk rm               (cmd/rm.go)
tk mv               → tk mv               (cmd/mv.go)
tk history          → tk history          (cmd/history.go)
```

### Sync & Remote
```
tk sync             → tk sync             (cmd/sync.go)
tk push             → tk push             (cmd/push.go - NEEDS CREATION)
tk pull             → tk pull             (cmd/pull.go - NEEDS CREATION)
tk ingest           → tk ingest           (cmd/ingest.go)

tk status sync      → tk sync-status      (cmd/status/sync.go → cmd/sync_status.go)

tk remote add       → tk remote-add       (cmd/remote/add.go → cmd/remote_add.go)
tk remote ls        → tk remote-ls        (cmd/remote/ls.go → cmd/remote_ls.go)
tk remote rm        → tk remote-rm        (cmd/remote/rm.go → cmd/remote_rm.go)
```

### Relations & Dependencies
```
tk relate add       → tk relate-add       (cmd/relate/add.go → cmd/relate_add.go)
tk relate ls        → tk relate-ls        (cmd/relate/ls.go → cmd/relate_ls.go)
tk relate remove    → tk relate-rm        (cmd/relate/remove.go → cmd/relate_rm.go)

# Note: keep tk dup as top-level command, implement in relate_dup.go
tk dup              → tk dup + tk relate-dup   (NEW cmd/relate_dup.go, registers both)

tk blockers         → tk blockers + tk relate-blockers  (cmd/blockers.go → cmd/relate_blockers.go, registers both)
tk blocked          → tk blocked + tk relate-blocked    (NEEDS CREATION cmd/relate_blocked.go, registers both)

tk graph            → tk graph + tk relate-graph        (cmd/graph.go → cmd/relate_graph.go, registers both)

tk conflicts        → tk relate-conflicts     (cmd/conflicts.go → cmd/relate_conflicts.go)
tk conflicts numbers → tk task-conflicts      (cmd/conflicts/numbers.go → cmd/task_conflicts.go)
```

### Projects
```
tk project create   → tk project-create   (cmd/project/create.go → cmd/project_create.go)
tk project ls       → tk project-ls       (cmd/project/ls.go → cmd/project_ls.go)
tk project rename   → tk project-rename   (cmd/project/rename.go → cmd/project_rename.go)
tk project rm       → tk project-rm       (cmd/project/rm.go → cmd/project_rm.go)
```

### Containers - Queues
```
tk queue create     → tk queue-create     (cmd/queue/create.go → cmd/queue_create.go)
tk queue push       → tk queue-push       (cmd/queue/push.go → cmd/queue_push.go)
tk queue pop        → tk queue-pop        (cmd/queue/pop.go → cmd/queue_pop.go)
tk queue list       → tk queue-ls         (cmd/queue/list.go → cmd/queue_ls.go)
tk queue show       → tk queue-show       (cmd/queue/show.go → cmd/queue_show.go)
tk queue rename     → tk queue-rename     (cmd/queue/rename.go → cmd/queue_rename.go)
tk queue rm         → tk queue-rm         (cmd/queue/rm.go → cmd/queue_rm.go)
```

### Containers - Stacks
```
tk stack create     → tk stack-create     (cmd/stack/create.go → cmd/stack_create.go)
tk stack push       → tk stack-push       (cmd/stack/push.go → cmd/stack_push.go)
tk stack pop        → tk stack-pop        (cmd/stack/pop.go → cmd/stack_pop.go)
tk stack list       → tk stack-ls         (cmd/stack/list.go → cmd/stack_ls.go)
tk stack show       → tk stack-show       (cmd/stack/show.go → cmd/stack_show.go)
tk stack rename     → tk stack-rename     (cmd/stack/rename.go → cmd/stack_rename.go)
tk stack rm         → tk stack-rm         (cmd/stack/rm.go → cmd/stack_rm.go)
```

### Containers - Groups
```
tk group create     → tk group-create     (cmd/group/create.go → cmd/group_create.go)
tk group add        → tk group-addtask    (cmd/group/add.go → cmd/group_addtask.go)
tk group remove     → tk group-rmtask     (cmd/group/remove.go → cmd/group_rmtask.go)
tk group list       → tk group-ls         (cmd/group/list.go → cmd/group_ls.go)
tk group show       → tk group-show       (cmd/group/show.go → cmd/group_show.go)
tk group rename     → tk group-rename     (cmd/group/rename.go → cmd/group_rename.go)
tk group rm         → tk group-delete     (cmd/group/rm.go → cmd/group_delete.go)
```

### Schema & Metadata
```
tk schema add       → tk schema-add       (cmd/schema/add_kind.go → cmd/schema_add.go)
tk schema list      → tk schema-ls        (cmd/schema/list_kinds.go → cmd/schema_ls.go)
tk schema export    → tk schema-export    (cmd/schema/export.go → cmd/schema_export.go)

tk meta set         → tk meta-set         (cmd/meta/set.go → cmd/meta_set.go)
tk meta get         → tk meta-get         (cmd/meta/get.go → cmd/meta_get.go)
tk meta list        → tk meta-ls          (cmd/meta/list.go → cmd/meta_ls.go)
tk meta claims      → tk meta-claims      (cmd/meta/claims.go → cmd/meta_claims.go)
```

### Debug
```
tk debug doctor     → tk debug-doctor     (cmd/debug/doctor.go → cmd/debug_doctor.go)
tk debug unsafe-repair-timestamps → tk debug-repair (cmd/debug/repair.go → cmd/debug_repair.go)
tk debug rebuild-projections → tk debug-rebuild (cmd/debug/rebuild.go → cmd/debug_rebuild.go)
tk debug rebuild-from-remote → tk debug-rebuild-from-remote (NEEDS CREATION)
tk debug fix-timestamps → tk debug-fix-timestamps (NEEDS CREATION or MOVE)

tk debug events list → tk debug-events-ls  (cmd/debug/events/list.go → cmd/debug_events_ls.go)
tk debug events show → tk debug-events-show (cmd/debug/events/show.go → cmd/debug_events_show.go)
tk debug events stats → tk debug-events-stats (cmd/debug/events/stats.go → cmd/debug_events_stats.go)

tk debug node show  → tk debug-node-show   (cmd/debug/node/show.go → cmd/debug_node_show.go)
tk debug node regen → tk debug-node-regen  (cmd/debug/node/regen.go → cmd/debug_node_regen.go)

tk debug id         → tk id + tk debug-id  (cmd/id.go → cmd/debug_id.go, registers both)
```

### Migration & Logs
```
tk migrate fix-container-item-ids → tk migrate-fix-container-item-ids
  (cmd/migrate/fix_container_item_ids.go → cmd/migrate_fix_container_item_ids.go)

tk migrate fix-relocate-bug → tk migrate-fix-relocate-bug
  (cmd/migrate/fix_relocate_bug.go → cmd/migrate_fix_relocate_bug.go)

tk migrate scan-deprecated → tk migrate-scan-deprecated
  (cmd/migrate/scan_deprecated.go → cmd/migrate_scan_deprecated.go)

tk log query        → tk log-query        (cmd/log/query.go → cmd/log_query.go)
tk log search       → tk log-search       (cmd/log/search.go → cmd/log_search.go)
```

### Database & System
```
tk init             → tk init             (cmd/init.go)
tk db path          → tk db-path          (cmd/path.go → cmd/db_path.go)
tk status           → REMOVED (use tk mark instead)
tk statusline       → tk statusline       (cmd/statusline.go)
tk mcp              → tk mcp              (cmd/mcp.go)
tk version          → tk version          (cmd/version.go)
```

## Files to Remove
```
cmd/status.go       (replaced by tk mark)
cmd/remote.go       (parent command - no longer needed)
cmd/relate.go       (parent command - no longer needed)
cmd/project.go      (parent command - no longer needed)
cmd/queue.go        (parent command - no longer needed)
cmd/stack.go        (parent command - no longer needed)
cmd/group.go        (parent command - no longer needed)
cmd/schema.go       (parent command - no longer needed)
cmd/meta.go         (parent command - no longer needed)
cmd/debug.go        (parent command - no longer needed)
cmd/migrate.go      (parent command - no longer needed)
cmd/log.go          (parent command - no longer needed)
cmd/conflicts.go    (becomes cmd/relate_conflicts.go)
```

## Directories to Remove (after moving files)
```
cmd/conflicts/
cmd/debug/events/
cmd/debug/node/
cmd/debug/
cmd/group/
cmd/log/
cmd/meta/
cmd/migrate/
cmd/project/
cmd/queue/
cmd/relate/
cmd/remote/
cmd/schema/
cmd/stack/
cmd/status/
```

## Commands That Need Creation
```
cmd/push.go         (currently inlined in root.go or missing)
cmd/pull.go         (currently inlined in root.go or missing)
cmd/relate_blocked.go (move from separate file if exists)
cmd/relate_dup.go   (needs to be created, currently might be separate)
```

## Alias Registration

Each canonical command file should register its own aliases. For example:

**cmd/relate_dup.go:**
```go
var relateDupCmd = &cobra.Command{
    Use:   "relate-dup <task1> <task2>",
    Short: "Mark two tasks as duplicates",
    // ...
}

var dupCmd = &cobra.Command{
    Use:   "dup <task1> <task2>",
    Short: "Mark two tasks as duplicates (alias for relate-dup)",
    Run:   relateDupCmd.Run,
}

func init() {
    RootCmd.AddCommand(relateDupCmd)
    RootCmd.AddCommand(dupCmd)  // Short alias
}
```

## Implementation Steps

1. Create new hyphenated command files in cmd/
2. Move code from subdirectory files to new files
3. Update Use: field to use hyphens
4. Register aliases where needed
5. Update root.go to remove parent commands and add new commands
6. Remove old parent command files (remote.go, project.go, etc.)
7. Remove subdirectories
8. Update tests to use new command names
9. Update documentation

## Breaking Changes

- All nested commands become hyphenated top-level commands
- `tk status` removed (use `tk mark`)
- `tk conflicts numbers` → `tk task-conflicts`
- `tk group add/remove` → `tk group-addtask/group-rmtask`
- `tk group rm` → `tk group-delete` (for deleting groups)
