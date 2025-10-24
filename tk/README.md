# tk - System-Wide Event-Sourced Task Tracker

tk is a command-line tool that tracks tasks system-wide using an append-only event log in a single SQLite database.

## Features

- **Event sourcing**: All task changes are recorded as immutable events
- **Claims-based status**: Multiple actors (human, agent, bot, qa, rel) can make status claims
- **Authority lattice**: Conflicts are resolved based on role authority (human > qa > rel > agent > bot)
- **Multi-valued registers**: Conflicting claims are preserved as tentative/effective
- **SQLite backend**: Durable, inspectable, and portable (pure Go, no CGO required)
- **Automatic setup**: Database is created automatically in `~/.tk/` on first use

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

### Create a task

```bash
tk new "wire up rc deploy toggle"
```

This creates a new task with a unique ID like `tk_01J3XM4NZ2R72`. No initialization is required.

### Set task status

```bash
tk status set tk_01J3XM4NZ2R72 in_progress
```

You can specify the axis and role:

```bash
tk status set tk_01J3XM4NZ2R72 done --axis generic --role agent
```

### Add a note to a task

```bash
tk note tk_01J3XM4NZ2R72 "Fixed the deployment toggle"
```

### View a task

```bash
tk view tk_01J3XM4NZ2R72
```

This shows the current state, all claims (effective and tentative), and notes.

### List tasks

```bash
tk ls
```

Filter by status:

```bash
tk ls --axis generic:in_progress
```

### Get database path

```bash
tk db path
```

### Initialize database (optional)

The database is created automatically when you first use tk. However, you can explicitly initialize it:

```bash
tk init
```

## Concepts

### Events

Every action in tk is recorded as an immutable event in the SQLite database. Events have:
- **ID**: ULID identifier
- **TS**: Lamport timestamp for ordering
- **Actor**: Username who created the event
- **Role**: Role of the actor (human, agent, bot, qa, rel)
- **Kind**: Event type (task.created, task.status.set, task.note.add)
- **Payload**: Event-specific data (JSON)

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

## Status (v0)

This is v0 of tk - a minimal claims tracker. The current implementation includes:

- Event store with SQLite backend in `~/.tk/`
- Automatic database initialization
- Basic task lifecycle (create, status, notes)
- Authority-based claim resolution
- CLI for all core operations

Not yet implemented:
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
