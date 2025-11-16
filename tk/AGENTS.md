# Agent Guidelines for tk

This document describes tk's design philosophy and conventions for AI agents working on the codebase.

## CLI Command Design Philosophy

tk uses a **flat, hyphenated command structure** with no nested subcommands. This design prioritizes clarity, discoverability, and eliminates ambiguity.

### Core Principles

1. **No "folder" commands** - Every command is a leaf node that performs an action
2. **Hyphenated grouping** - Related commands share a prefix: `remote-add`, `remote-ls`, `remote-rm`
3. **Flat namespace** - All commands are at the top level, no nested subcommands
4. **Short aliases for frequent operations** - `tk ls` is the canonical form, but related commands use `tk queue-ls`, `tk remote-ls`, etc.
5. **Breaking changes over backwards compatibility** - We prefer clean design to legacy support

### Why This Design?

**Anti-pattern we avoid:**
```bash
# BAD: Command is both a "folder" and an action
tk sync              # Does this sync, or show help?
tk sync push         # Subcommand
tk sync pull         # Subcommand
```

**Our pattern:**
```bash
# GOOD: Clear, unambiguous
tk sync              # Always syncs
tk sync-status       # Shows sync status
tk push              # Pushes
tk pull              # Pulls
```

### Naming Conventions

**Short names for frequent operations:**
- `ls` not `list`
- `rm` not `remove`
- `mv` not `move`

**Hyphenated names for grouped commands:**
- `remote-add` not `remote add`
- `project-create` not `project create`
- `debug-rebuild` not `debug rebuild`

**Aliases for ergonomics:**
- Frequently used commands get both long and short forms
- Example: `tk dup` and `tk relate-dup` both work
- Implementation lives in the canonical file (`cmd/relate_dup.go`)
- Aliases are registered in the same file

### File Organization

**One command = one file, always:**
```
cmd/
  new.go                    # tk new
  ls.go                     # tk ls
  remote_add.go             # tk remote-add (canonical)
  remote_ls.go              # tk remote-ls
  relate_dup.go             # tk relate-dup (canonical, also registers "tk dup" alias)
  queue_create.go           # tk queue-create
  debug_events_ls.go        # tk debug-events-ls
```

**No subdirectories in cmd/**

All command files live directly in `cmd/`. This makes it easy to find the implementation:
- Want to modify `tk remote-add`? → `cmd/remote_add.go`
- Want to add `tk project-archive`? → Create `cmd/project_archive.go`

### Command Groups

Commands are logically grouped by prefix, but remain flat in the CLI:

**Sync operations:**
```bash
tk sync, tk push, tk pull, tk ingest, tk import-beads, tk sync-status
```

**Remote management:**
```bash
tk remote-add, tk remote-ls, tk remote-rm
```

**Relations:**
```bash
tk relate-add, tk relate-ls, tk relate-rm
tk dup (alias for relate-dup)
tk blockers (alias for relate-blockers)
tk blocked (alias for relate-blocked)
tk graph (alias for relate-graph)
tk conflicts (alias for relate-conflicts)
```

**Projects:**
```bash
tk project-create, tk project-ls, tk project-rename, tk project-rm
```

**Containers (Queues, Stacks, Groups):**
```bash
tk queue-create, tk queue-push, tk queue-pop, tk queue-ls, tk queue-show, tk queue-rename, tk queue-rm
tk stack-create, tk stack-push, tk stack-pop, tk stack-ls, tk stack-show, tk stack-rename, tk stack-rm
tk group-create, tk group-add, tk group-rm, tk group-ls, tk group-show, tk group-rename, tk group-delete
```

**Schema & Metadata:**
```bash
tk schema-add, tk schema-ls, tk schema-export
tk meta-set, tk meta-get, tk meta-ls, tk meta-claims
```

**Debug:**
```bash
tk debug-doctor, tk debug-repair, tk debug-rebuild, tk debug-rebuild-from-remote
tk debug-events-ls, tk debug-events-show, tk debug-events-stats
tk debug-node-show, tk debug-node-regen
tk id (alias for debug-id)
```

### Adding New Commands

When adding a new command:

1. **Choose the canonical name** - Use the full hyphenated form
2. **Create the file** - `cmd/{name}.go` with underscores for hyphens
3. **Register the command** - In `cmd/root.go`
4. **Add aliases if needed** - Register in the same file as the canonical command
5. **Update help text** - Include "See Also" references to related commands

Example:
```go
// cmd/remote_add.go
var remoteAddCmd = &cobra.Command{
    Use:   "remote-add <name> <path>",
    Short: "Add a new remote",
    // ... implementation
}

func init() {
    // Register canonical command
    RootCmd.AddCommand(remoteAddCmd)
}
```

### Migration from Nested Commands

If you find existing code with nested subcommands (e.g., `tk remote add`), flatten it:

**Before:**
```go
// cmd/remote.go
var remoteCmd = &cobra.Command{Use: "remote"}
var remoteAddCmd = &cobra.Command{Use: "add"}
remoteCmd.AddCommand(remoteAddCmd)
```

**After:**
```go
// cmd/remote_add.go
var remoteAddCmd = &cobra.Command{Use: "remote-add"}
RootCmd.AddCommand(remoteAddCmd)
```

### Examples from Popular CLIs

This pattern is used by well-designed tools:

**systemctl:**
```bash
systemctl start, systemctl stop, systemctl list-units, systemctl daemon-reload
```

**rustup:**
```bash
rustup update, rustup toolchain-list, rustup component-add
```

Our design is inspired by these tools, prioritizing:
- **Clarity** - No ambiguity about what a command does
- **Discoverability** - Tab completion shows all commands
- **Consistency** - Same pattern everywhere
- **Simplicity** - No need to learn nested command structure

### When to Add Aliases

Add a short alias when:
1. The command is used **very frequently** (multiple times per day)
2. The full name is **significantly longer** than the alias
3. The alias is **obvious** and doesn't conflict with other commands

Examples of good aliases:
- `tk dup` for `tk relate-dup` ✓ (common operation, clear meaning)
- `tk id` for `tk debug-id` ✓ (frequently needed for scripting)
- `tk blockers` for `tk relate-blockers` ✓ (daily workflow)

Examples of unnecessary aliases:
- `tk rc` for `tk remote-create` ✗ (not obvious)
- `tk pc` for `tk project-create` ✗ (not frequently used enough)

### Help and Documentation

Commands should be self-documenting:

```bash
tk --help                    # Lists all commands
tk remote-add --help         # Help for specific command
```

Use "See Also" sections to link related commands:
```go
SeeAlso: []string{
    "remote-ls",
    "remote-rm",
    "sync",
},
```

---

## Development Workflow

When working on tk:

1. **Use `tk-dev`** - Automatically builds from source
2. **Track your work** - Create tasks with `tk new`
3. **Test thoroughly** - Event sourcing means bugs are permanent in the log
4. **Document changes** - Update help text and this file

## Code Style

- Follow Go conventions
- Use structured logging (`slog`)
- Keep event types immutable (append-only)
- Test with multiple databases using `TK_DB_PATH`

---

*This document should be updated whenever CLI design patterns change.*

---

## Complete Command Reference

After the refactoring, all tk commands use a flat, hyphenated structure. Here's the complete list:

### Core Task Management
```
tk new              - Create a new task
tk ls               - List tasks
tk show             - Show task details
tk mark             - Set task status
tk edit             - Edit task fields (number, title, status)
tk describe         - Change task title
tk note             - Add a note to a task
tk attach           - Attach a file to a task
tk rm               - Delete a task
tk mv               - Move a task to another project
tk history          - Show recent task status changes
```

### Relations & Dependencies
```
# Full commands
tk relate-add       - Add a relation between two tasks
tk relate-ls        - List all relations for a task
tk relate-rm        - Remove a relation between two tasks
tk relate-dup       - Mark two tasks as duplicates
tk relate-blockers  - List blockers for a task
tk relate-blocked   - List all blocked tasks
tk relate-graph     - Show a graph of task relations
tk relate-conflicts - Show relation conflicts for a task
tk task-conflicts   - List task number collisions

# Short aliases (same implementation)
tk dup              - Alias for relate-dup
tk blockers         - Alias for relate-blockers
tk blocked          - Alias for relate-blocked
tk graph            - Alias for relate-graph
tk conflicts        - Alias for relate-conflicts
```

### Projects
```
tk project-create   - Create a new project
tk project-ls       - List all projects
tk project-rename   - Rename a project
tk project-rm       - Delete a project
```

### Sync & Remote
```
tk sync             - Sync with remote (pull → ingest → export → push)
tk push             - Push local segments to remote
tk pull             - Pull segments from remote
tk ingest           - Ingest events from segment files
tk import-beads     - Import issues from beads JSONL format
tk sync-status      - Show sync status for all remotes

tk remote-add       - Add a new remote
tk remote-ls        - List configured remotes
tk remote-rm        - Remove a remote
```

### Containers - Queues (FIFO)
```
tk queue-create     - Create a new queue
tk queue-push       - Push an item onto a queue
tk queue-pop        - Pop an item from a queue
tk queue-ls         - List queues or queue members
tk queue-show       - Show queue details
tk queue-rename     - Rename a queue
tk queue-rm         - Remove a queue (soft delete)
```

### Containers - Stacks (LIFO)
```
tk stack-create     - Create a new stack
tk stack-push       - Push an item onto a stack
tk stack-pop        - Pop an item from a stack
tk stack-ls         - List stacks or stack members
tk stack-show       - Show stack details
tk stack-rename     - Rename a stack
tk stack-rm         - Remove a stack (soft delete)
```

### Containers - Groups (Unordered Sets)
```
tk group-create     - Create a new group
tk group-addtask    - Add an item to a group
tk group-rmtask     - Remove an item from a group
tk group-ls         - List groups or group members
tk group-show       - Show group details
tk group-rename     - Rename a group
tk group-delete     - Remove a group (soft delete)
```

### Schema & Metadata
```
tk schema-add       - Define a new schema kind
tk schema-ls        - List defined schema kinds
tk schema-export    - Export schema definition

tk meta-set         - Set metadata value (creates claim)
tk meta-get         - Get effective metadata value
tk meta-ls          - List all metadata for a task
tk meta-claims      - Show all competing claims for a metadata key
```

### Debug
```
tk debug-doctor     - Verify database health and report issues
tk debug-repair     - UNSAFE: Repair Lamport timestamp corruption
tk debug-rebuild    - Rebuild all projection tables from events
tk debug-rebuild-from-remote  - Rebuild local database from remote segments
tk debug-fix-timestamps       - Reassign Lamport timestamps to events

tk debug-events-ls    - List all events in the database
tk debug-events-show  - Show detailed information about a specific event
tk debug-events-stats - Show statistics about events in the database

tk debug-node-show  - Show the current node ID
tk debug-node-regen - Regenerate the node ID (use with caution)

tk debug-id         - Show canonical identifiers for a task
tk id               - Alias for debug-id
```

### Migration & Logs
```
tk migrate-fix-container-item-ids  - Fix display IDs in container_members
tk migrate-fix-relocate-bug        - Fix orphaned tasks from corrupt events
tk migrate-scan-deprecated         - Scan event log for deprecated field usage

tk log-query        - Query invocation logs with SQL
tk log-search       - Search invocation logs for a pattern
```

### Database & System
```
tk init             - Create a new tk database
tk db-path          - Print the current database path
tk statusline       - Display recently updated tasks
tk mcp              - Run tk as an MCP server
```

### Total Command Count
- **100+ commands** in a completely flat namespace
- **6 commands with aliases** for frequent operations
- **No nested subcommands** - every command is a leaf node
- **Consistent naming** - all related commands share prefixes
