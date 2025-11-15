# tk Usage Guide

This guide covers how to use tk for day-to-day task tracking.

## Database Location

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

## Projects

Tasks are organized by projects. Each project has:
- A stable **project UID** (e.g., `prj_01J5Q...`) that never changes
- A human-readable **name**

### Create a new project

```bash
tk project create work "Work tasks"
tk project create my-project "My personal project"
```

**Important:** Project names must be lowercase letters and dashes only (e.g., `work`, `my-project`, `backend`).

### List projects

```bash
tk project ls
```

Shows all projects with their UIDs, names, and descriptions.

### Create tasks in a project

```bash
tk new "Task title" --project work
tk new "Another task" -p tk
```

The `--project` (`-p`) flag accepts either a project name or a project UID.

## Tasks

### Create a task

Create a task in a project:

```bash
tk new "wire up rc deploy toggle" --project work
```

This creates a new task with a unique ID. Tasks are numbered within their project and display as short IDs like `work-1` when possible.

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

See [../RELATIONS.md](../RELATIONS.md) for complete documentation on task relations.

## Database Management

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

# List only project.created events
tk events list --kind project.created

# Show first 10 events
tk events list --limit 10

# Show detailed event information
tk events show ev-1-abc123

# Show event statistics
tk events stats
```
