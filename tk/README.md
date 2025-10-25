# tk - System-Wide Event-Sourced Task Tracker

tk is a command-line tool that tracks tasks system-wide using an append-only event log with offline-first sync.

## Features

- **Event sourcing**: All task changes are recorded as immutable events
- **Prefixes**: Organize tasks with custom prefixes (e.g., `tk-1`, `foo-2`, `bar-3`)
- **Namespace isolation**: Each prefix has its own task numbering, scoped by node ID
- **Claims-based status**: Multiple actors (human, agent, bot, qa, rel) can make status claims
- **Authority lattice**: Conflicts are resolved based on role authority (human > qa > rel > agent > bot)
- **Multi-valued registers**: Conflicting claims are preserved as tentative/effective
- **SQLite backend**: Durable, inspectable, and portable (pure Go, no CGO required)
- **Automatic setup**: Database is created automatically in `~/.tk/` on first use
- **Offline-first sync**: Sync events between machines using immutable segment files (v1)
- **iCloud sync**: Use iCloud Drive as a sync remote for multi-Mac workflows (v1)

## Installation

### From Source

```bash
cd tk
go build -o tk .
```

### Using mise

```bash
mise run //tk:run
```

## Usage

### Database Location

tk stores its database in `~/.tk/tk.db` by default. The database and directory are created automatically on first use.

### Prefixes

tk supports multiple task prefixes, allowing you to organize tasks by project or category. Each prefix has its own namespace for task numbers.

#### Default prefix

When you initialize tk, a default `tk` prefix is automatically created.

#### Create a new prefix

```bash
tk prefix create foo "Tasks for foo project"
tk prefix create bar "Tasks for bar project"
```

#### List prefixes

```bash
tk prefix list
```

The list shows the state of each prefix:
- **explicit**: Created with `tk prefix create`
- **discovered**: Found in task IDs but not explicitly created
- **removed**: Marked as removed with `tk prefix remove`

Show prefixes from all nodes (including synced prefixes):

```bash
tk prefix list --all
```

#### Remove a prefix

Mark a prefix as removed (does not delete tasks):

```bash
tk prefix remove old-prefix
```

This marks the prefix as "removed" in the prefix list but does not delete any tasks using that prefix.

### Move tasks between prefixes

Move a task to a different prefix:

```bash
tk mv tk-1 foo:1
```

This moves task `tk-1` to prefix `foo` with number `1`, creating the new ID `foo-1`. The old ID `tk-1` becomes an alias and can still be used to reference the task.

Move multiple tasks at once:

```bash
tk mv tk-1 foo:1 tk-2 foo:2
```

Move a task and auto-assign the next available number:

```bash
tk mv tk-1 foo --auto
```

Dry run to preview changes:

```bash
tk mv tk-1 foo:1 -n
```

Options:
- `--alias` / `--no-alias`: Create (or don't create) an alias for the old ID (default: create alias)
- `--auto`: Auto-assign next available number on collision
- `--keep-number`: Keep the same number in the new prefix
- `-n` / `--dry-run`: Show what would happen without making changes
- `--on-collision`: What to do on collision: fail, auto, or swap (default: fail)

### Create a task

Create a task with the default `tk` prefix:

```bash
tk new "wire up rc deploy toggle"
```

This creates a new task with a unique ID like `tk-1-abc123` where:
- `tk` is the prefix
- `1` is the task number within that prefix
- `abc123` is your node ID

The task is displayed as `tk-1` when there's no ambiguity (i.e., no other node has created a `tk-1` task).

Create a task with a specific prefix:

```bash
tk new --prefix foo "implement foo feature"
```

This creates a task like `foo-1-abc123`.

### Set task status

```bash
tk status set tk-1 in_progress
tk status set foo-1 done
```

You can specify the axis and role:

```bash
tk status set tk-1 done --axis generic --role agent
```

### Add a note to a task

```bash
tk note tk-1 "Fixed the deployment toggle"
tk note foo-1 "Implemented new feature"
```

### View a task

```bash
tk view tk-1
tk view foo-1
```

This shows the current state, all claims (effective and tentative), and notes.

### List tasks

List all tasks:

```bash
tk ls
```

Filter by status:

```bash
tk ls --axis generic:in_progress
```

Filter by prefix:

```bash
tk ls --prefix foo
```

Filter by multiple prefixes:

```bash
tk ls --prefix foo --prefix bar
```

Combine filters:

```bash
tk ls --prefix foo --axis generic:in_progress
```

Show task aliases:

```bash
tk ls --aliases
```

This displays a table with an additional "Aliases" column showing any aliases for tasks (e.g., old IDs after moving tasks).

### Get database path

```bash
tk db path
```

### Initialize database (optional)

The database is created automatically when you first use tk. However, you can explicitly initialize it:

```bash
tk init
```

## Sync (v1)

tk v1 supports offline-first sync between machines using immutable event segments.

### Node ID

Each tk installation has a unique 6-character node ID. View yours with:

```bash
tk node show
```

Task and event IDs include the node ID (e.g., `tk-1-abc123`, `ev-42-abc123`) to prevent collisions.

### Add a sync remote

To sync between two Macs using iCloud Drive:

```bash
tk remote add icloud folder ~/Library/Mobile\ Documents/com~apple~CloudDocs/tk-events
```

### Sync workflow

Initial sync on Machine A:
```bash
tk export --all icloud  # Export all existing events
tk sync icloud          # Push to iCloud
```

Wait for iCloud to finish uploading, then on Machine B:
```bash
tk remote add icloud folder ~/Library/Mobile\ Documents/com~apple~CloudDocs/tk-events
tk sync icloud          # Pull and sync
```

Regular sync on either machine:
```bash
tk sync icloud
```

The sync command performs: pull → ingest → export → push

### Check sync status

```bash
tk status sync
```

Shows divergence between local and remote segments.

### Individual sync operations

- `tk export [remote]` - Export local events to segments
- `tk ingest [remote|file]` - Ingest events from segments
- `tk pull [remote]` - Pull segments from remote
- `tk push [remote]` - Push segments to remote

## Concepts

### Prefixes

Prefixes are first-class entities that allow you to organize tasks by project, category, or any other grouping. Each prefix:

- Has a description explaining its purpose
- Is scoped to the node that created it (preventing conflicts when syncing)
- Has its own independent task counter (e.g., `foo-1`, `foo-2` are separate from `bar-1`, `bar-2`)
- Can be filtered in the `tk ls` command
- Must be lowercase, 2-20 characters, start with a letter, and contain only letters, digits, and underscores

**Counter Model:**
- Counters are local to `(prefix, node)` and never sync between nodes
- Each node maintains its own counter for each prefix it uses
- Global uniqueness is guaranteed by the node suffix in task IDs
- Example: Node A creates `foo-1-aaa111` and Node B independently creates `foo-1-bbb222`
- Both tasks are globally unique despite having the same prefix and number

When you sync with other machines:
- Prefix metadata (descriptions) is synced via `prefix.created` events
- Their prefixes become visible in `tk prefix list --all`
- Each node maintains its own task numbering for each prefix
- Task IDs include the node suffix to ensure global uniqueness (e.g., `foo-1-abc123`)

**Prefix Naming Rules:**
- 2-20 characters long
- Must start with a lowercase letter (a-z)
- Can contain lowercase letters, digits, and underscores
- No hyphens allowed (reserved for ID parsing)
- Reserved prefixes: ev, event, task, node, remote, sync

**Prefix States:**
- **explicit**: Created with `tk prefix create`, has full metadata
- **discovered**: Found in task IDs but not explicitly created (no description)
- **removed**: Marked as removed with `tk prefix remove`, but tasks still exist

### Task Identity and Aliases

Each task has two identifiers:
- **Task UUID**: A unique, immutable identifier that never changes (e.g., `task-abc123xyz...`)
- **Task ID**: The current display ID (e.g., `foo-1-node123`)

When you move a task between prefixes using `tk mv`, the task UUID stays the same, but the task ID changes. The old task ID becomes an **alias**, allowing you to reference the task by its old ID.

Example:
```bash
# Create a task
tk new "My task"  # Creates tk-1

# Move it to foo prefix
tk mv tk-1 foo:1  # Now foo-1, but tk-1 is an alias

# Both IDs work
tk view tk-1      # Works (using alias)
tk view foo-1     # Works (using current ID)
```

Aliases are preserved indefinitely and synced between machines, ensuring old links and references continue to work.

### Events

Every action in tk is recorded as an immutable event in the SQLite database. Events have:
- **ID**: Event ID in format `ev-<number>-<node>` (e.g., `ev-42-abc123`)
- **TS**: Lamport timestamp for ordering
- **Actor**: Username who created the event
- **Role**: Role of the actor (human, agent, bot, qa, rel)
- **Kind**: Event type (task.created, task.status.set, task.note.add, prefix.created)
- **Payload**: Event-specific data (JSON)

Supported event types:
- `task.created` - A new task was created
- `task.status.set` - Task status was updated
- `task.note.add` - A note was added to a task
- `task.reprefix` - Task was moved to a different prefix
- `task.alias.added` - An alias was added for a task
- `prefix.created` - A new prefix was created
- `prefix.removed` - A prefix was marked as removed

### Claims

A claim is a status assertion made by an actor. Multiple actors can make claims about the same task, and conflicts are resolved based on authority.

### Authority Lattice

Role authority (highest to lowest):
1. **human** - Humans have the highest authority
2. **qa** - QA/testing roles
3. **rel** - Release/deployment roles
4. **agent** - AI agents
5. **bot** - Automated bots

When concurrent claims exist (same timestamp), the claim with the highest authority becomes effective, and lower-authority claims are marked as tentative.

### Axes

Tasks can have multiple status axes. Currently, only the "generic" axis is used, but the system is designed to support workflow-specific axes in future versions.

## Status

### v1 (current)

- Event sourcing with stable event IDs (`ev-<number>-<node>`)
- Task IDs with prefix and node suffix (`<prefix>-<number>-<node>`)
- Task UUIDs for stable identity across moves
- Multiple task prefixes with independent counters
- Prefix management (create, list, remove, filter)
- Task movement between prefixes with alias preservation
- Task aliases for backward compatibility
- Offline-first sync via immutable segment files
- iCloud Drive folder remote support
- Segment files with zstd compression
- Automatic deduplication on ingest
- Lamport clock synchronization
- Node collision detection

### Not yet implemented

- Context binding (repo, branch, commit tracking)
- JJ integration
- Custom axes and workflows
- Relations between tasks (blocks, subtasks)

## Testing

```bash
go test ./...
```

## Development

Format code:
```bash
go fmt ./...
```

Run via mise:
```bash
mise run //tk:run new "test task"
mise run //tk:run ls
```
