# tk Concepts

This document explains the core concepts and architecture of tk.

## Task Identity

Each task has two identifiers:
- **Task UID**: A unique, immutable identifier that never changes (e.g., `tsk_01J5Q...`)
- **Task ID**: The current display ID (e.g., `myproj-1`)

The Task UID is the true identity of a task and never changes. The Task ID is a human-friendly display format that combines the project name (or UID) with the task number.

## Events

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

## Claims

A claim is a status assertion made by an actor. Multiple actors can make claims about the same task, and conflicts are resolved based on authority.

For example:
- An agent might mark a task as "done"
- A human might later mark it as "in_progress"
- The human's claim takes precedence due to higher authority

## Authority Lattice

Role authority (highest to lowest):
1. **human** - Humans have the highest authority
2. **qa** - QA/testing roles
3. **rel** - Release/deployment roles
4. **agent** - AI agents
5. **bot** - Automated bots

When concurrent claims exist (same timestamp), the claim with the highest authority becomes effective, and lower-authority claims are marked as tentative.

## Axes

Tasks can have multiple status axes. Currently, only the "generic" axis is used, but the system is designed to support workflow-specific axes in future versions.

An axis represents a dimension of task state. For example:
- **generic** - General task status (todo, in_progress, done)
- **qa** (future) - QA status (untested, testing, passed, failed)
- **deploy** (future) - Deployment status (undeployed, staging, production)

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
