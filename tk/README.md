# tk

tk is my experimental task / personal info tracker for humans and agents.

I want to track thousands of tasks across multiple repositories, chats, emails, feature requests, and hobbies.
I want agents to have access to my data and be able to answer questions like "what promises have I made to people", etc.

*tk* is inspired by [beads](https://github.com/steveyegge/beads), but while beads focuses on helping the agents work better, *tk* focuses on helping my ADHD.
I suppose it helps the agents because they, too, have ADHD.

## This is a personal tool

tk is built for my brain and my usecase of "I want to do everything".
If your are similar to me, you're welcome to use it. If you suggest any ideas, I might try them out.

I try to make *tk* fairly general so that I can try out different things and see what works best.

## Technical Overview

tk is a command-line tool that tracks tasks system-wide using:

- **Event sourcing**: All changes recorded as immutable events in SQLite
- **Offline-first sync**: Sync between machines using immutable segment files (iCloud Drive support)
- **Task relations**: Model dependencies (blocks, subtasks) and hierarchies
- **Multi-valued registers**: Conflicting claims resolved based on role authority
- **No CGO required**: Pure Go SQLite implementation for easy deployment

## Development

- **`tk`**: Globally installed binary (e.g., via `brew install` or `want mono tk@latest`)
- **`tk-dev`**: Development binary built from source (use for testing changes)
  - Build: `mise tk:build`
  - Run: `tk-dev <command>`
  - Always use `tk-dev` when testing local changes to avoid conflicts with global installation

## Installation

### From Source

```bash
cd tk
go build -o tk .
```

### Using mise

```bash
mise run tk
```

## Usage

### Database Location

tk stores its database in `~/.tk/tk.db` by default. The database and directory are created automatically on first use.

**Custom database location:**

You can override the default location using the `TK_DB_PATH` environment variable:

```bash
export TK_DB_PATH=/tmp/tk-test.db
tk new "Test task"  # Creates database at /tmp/tk-test.db
tk ls               # Uses /tmp/tk-test.db
```

This is useful for:
- Testing with isolated databases without affecting your main database
- Running multiple tk instances with separate databases
- Custom database locations (e.g., Dropbox, project-specific databases)

### Projects

Tasks are organized by projects. Each project has:
- A stable **project UID** (e.g., `prj_01J5Q...`) that never changes
- A human-readable **name**
- Per-node **aliases** for easy reference

#### Create a new project

```bash
tk project create "My Project" "Project description" --alias myproj
```

This creates a new project with a stable UID and adds the alias `myproj` on your current node.

#### List projects

```bash
tk project list
```

Shows all projects with their UIDs, names, aliases, and descriptions.

#### Manage project aliases

Add an alias for a project:

```bash
tk project alias add prj_01J5Q... myalias
```

Remove an alias:

```bash
tk project alias remove myalias
```

Aliases are per-node, so different nodes can use the same alias for different projects without conflicts.

#### Create tasks in a project

```bash
tk new "Task title" --project myproj
tk new "Another task" -p tk
```

The `--project` (`-p`) flag accepts either a project alias or a project UID.

### Create a task

Create a task in a project:

```bash
tk new "wire up rc deploy toggle" --project myproj
```

This creates a new task with a unique ID. Tasks are numbered within their project and display as short IDs like `myproj-1` when possible.

### Set task status

```bash
tk mark myproj-1 in_progress
tk mark myproj-2 done
```

You can specify the axis and role:

```bash
tk mark myproj-1 done --axis generic --role agent
```

### Add a note to a task

```bash
tk note myproj-1 "Fixed the deployment toggle"
tk note myproj-2 "Implemented new feature"
```

### View a task

```bash
tk show myproj-1
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

Filter by project:

```bash
tk ls -p myproj
```

Filter by multiple projects:

```bash
tk ls -p myproj -p other
```

Combine filters:

```bash
tk ls -p myproj --axis generic:in_progress
```

Show task aliases:

```bash
tk ls --aliases
```

This displays a table with an additional "Aliases" column showing any aliases for tasks.

## Task Relations

tk supports relations between tasks for modeling dependencies, hierarchies, and other relationships.

### Add Relations

```bash
# Task myproj-1 blocks task myproj-2
tk relate add myproj-1 blocks myproj-2

# Task myproj-3 is a subtask of myproj-1
tk relate add myproj-1 subtask myproj-3 --note "API design"

# Mark tasks as duplicates
tk dup myproj-4 myproj-5
```

### View Relations

```bash
# Show graph of task dependencies
tk graph myproj-1 --type blocks

# List all tasks blocking myproj-2
tk blockers myproj-2

# List all blocked tasks
tk blocked

# Filter tasks by blocked status
tk ls --blocked
tk ls --unblocked
```

See [RELATIONS.md](RELATIONS.md) for complete documentation on task relations.

### Get database path

```bash
tk db path
```

### Initialize database (optional)

The database is created automatically when you first use tk. However, you can explicitly initialize it:

```bash
tk init
```

## Sync

tk supports offline-first sync between machines using immutable event segments.

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
tk push --all icloud    # Export all existing events
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

- `tk push [remote]` - Export local events and push segments to remote
- `tk ingest [remote|file]` - Ingest events from segments
- `tk pull [remote]` - Pull segments from remote
- `tk push [remote]` - Push segments to remote

### Debug sync issues

Use the `events` command to inspect events in the database:

```bash
# List all events
tk events list

# List only prefix.created events
tk events list --kind prefix.created

# Show first 10 events
tk events list --limit 10

# Show detailed event information
tk events show ev-1-abc123

# Show event statistics
tk events stats
```


## Concepts

### Task Identity and Aliases

Each task has two identifiers:
- **Task UUID**: A unique, immutable identifier that never changes (e.g., `task-abc123xyz...`)
- **Task ID**: The current display ID (e.g., `myproj-1`)

Task IDs can have aliases, allowing you to reference tasks by alternative names. Aliases are preserved indefinitely and synced between machines, ensuring old links and references continue to work.

### Events

Every action in tk is recorded as an immutable event in the SQLite database. Events have:
- **ID**: Event ID in format `ev-<number>-<node>` (e.g., `ev-42-abc123`)
- **TS**: Lamport timestamp for ordering
- **Actor**: Username who created the event
- **Role**: Role of the actor (human, agent, bot, qa, rel)
- **Kind**: Event type (project.created, task.created, task.status.set, task.note.add)
- **Payload**: Event-specific data (JSON)

Supported event types:

- `project.created` - A new project was created
- `project.alias.add` - An alias was added to a project
- `project.alias.remove` - An alias was removed from a project
- `task.created` - A new task was created
- `task.number.set` - Task number was assigned or changed
- `task.relocate` - Task was moved to a different project
- `task.title.set` - Task title was changed
- `task.status.set` - Task status was updated
- `task.note.add` - A note was added to a task

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

### Current Features

- **Projects with stable UIDs**: First-class projects with immutable identifiers (`prj_...`)
- **Per-node project aliases**: Flexible naming without conflicts
- **Task UIDs as identity**: Stable task UIDs (`tsk_...`) - numbers are mutable labels
- **Collision-tolerant numbering**: Multiple tasks can share the same number
- **Task relations**: Model dependencies (blocks, subtasks, related, duplicate, supersedes)
- **OR-set CRDT semantics**: Relation conflict resolution
- **Blocked task tracking**: Configurable blocking axis
- **Cycle detection**: For blocks and subtasks
- **Relation visualization**: `tk graph` command
- **Event sourcing**: Stable event IDs (`ev-<number>-<node>`)
- **Offline-first sync**: Immutable segment files
- **iCloud Drive support**: Folder remote for syncing
- **Segment compression**: zstd compression
- **Automatic deduplication**: On event ingest
- **Lamport clock sync**: Distributed timestamp ordering
- **Node collision detection**: Prevent ID conflicts

### Not yet implemented

- Context binding (repo, branch, commit tracking)
- JJ integration
- Custom axes and workflows
- Task hierarchies with rollups (stories, epics, progress tracking)
- External project integrations (GitHub, Linear, Jira)

## Testing

```bash
go test ./...
```

## Development

Run via mise:
```bash
mise run tk new "test task"
mise run tk ls
```
