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
  - Just run `tk-dev <command>` - it builds automatically when needed
  - Always use `tk-dev` when testing local changes to avoid conflicts with global installation

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

#### Create a new project

```bash
tk project create work "Work tasks"
```

This creates a new project with a stable UID.

#### List projects

```bash
tk project ls
```

Shows all projects with their UIDs, names, and descriptions.

#### Create tasks in a project

```bash
tk new "Task title" --project work
tk new "Another task" -p prj_01J5Q...
```

The `--project` (`-p`) flag accepts either a project name or a project UID.

### Create a task

Create a task in a project:

```bash
tk new "wire up rc deploy toggle" --project work
```

This creates a new task with a unique ID. Tasks are numbered within their project and display as short IDs like `work-1` when possible.

### Set task status

```bash
tk mark work-1 wip
tk mark work-2 done
```

You can specify the role:

```bash
tk mark work-1 done --role agent
```

### Add a note to a task

```bash
tk note work-1 "Fixed the deployment toggle"
tk note work-2 "Implemented new feature"
```

### View a task

```bash
tk show work-1
```

This shows the current state, all claims (effective and tentative), and notes.

### List tasks

List all tasks:

```bash
tk ls
```

Filter by status:

```bash
tk ls --status wip
```

Filter by project:

```bash
tk ls -p work
```

Filter by multiple projects:

```bash
tk ls -p work -p personal
```

Combine filters:

```bash
tk ls -p work --status wip
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
# Task work-1 blocks task work-2
tk relate add work-1 blocks work-2

# Task work-3 is a subtask of work-1
tk relate add work-1 subtask work-3 --note "API design"

# Mark tasks as duplicates
tk dup work-4 work-5
```

### View Relations

```bash
# Show graph of task dependencies
tk graph work-1 --type blocks

# List all tasks blocking work-2
tk blockers work-2

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
tk debug node show
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

Use the `debug events` command to inspect events in the database:

```bash
# List all events
tk debug events list

# List only prefix.created events
tk debug events list --kind prefix.created

# Show first 10 events
tk debug events list --limit 10

# Show detailed event information
tk debug events show ev-1-abc123

# Show event statistics
tk debug events stats
```


## Concepts

### Task Identity and Aliases

Each task has two identifiers:
- **Task UUID**: A unique, immutable identifier that never changes (e.g., `task-abc123xyz...`)
- **Task ID**: The current display ID (e.g., `work-1`)

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
