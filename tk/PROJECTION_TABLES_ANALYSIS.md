# Projection Tables vs Reducer: Command Analysis

This document analyzes which tk commands read from projection tables versus using the reducer for data access.

## Background

tk uses an event-sourced architecture with:
- **Events table**: Source of truth (append-only log)
- **Projection tables**: Denormalized views of data for efficient queries
- **Reducer**: In-memory state built from events for complex queries

The projection tables include:
- `projects` - Project metadata
- `tasks` - Task metadata
- `task_numbers` - Task numbering within projects
- `project_aliases` - Project aliases per node
- `containers` - Queue/stack/group instances
- `container_members` - Membership in containers
- `container_kinds` - Container type definitions
- `item_kinds` - Item type definitions (task, bug, decision, etc.)

## Analysis Summary

### Commands that ONLY use the Reducer (no projection table reads)

These commands get all their data from the reducer:
- `relate-ls` - List relationships
- `relate-blocked` - Show blocked tasks
- `relate-blockers` - Show what blocks a task
- `relate-conflicts` - Show conflicting claims
- `relate-graph` - Display relationship graph
- `attach` - Attachment operations

### Commands that read Projection Tables (may or may not also use reducer)

#### Task commands
- **`ls`** - YES (reads `projects` for project names in queries) + uses reducer
- **`new`** - YES (reads `tasks` to get parent project, validates projects exist)
- **`show`** - YES (reads nothing directly, but uses `RenderTaskDisplayID` which reads `task_numbers`) + uses reducer
- **`mark`** - YES (reads `task_numbers` to get project UID) + uses reducer  
- **`note`** - YES (reads `tasks` to get task title)
- **`history`** - YES (reads `tasks` to get task titles)
- **`query`** - YES (uses `RenderTaskDisplayID` which reads `task_numbers`) + uses reducer

#### Project commands
- **`project-ls`** - YES (reads `projects` and `project_aliases`)
- **`project-rm`** - YES (reads `tasks` to count tasks in project)

#### Container commands (queue/stack/group)
All container list/show/modification commands read from projection tables:
- **`queue-ls`** - YES (reads `containers`, `container_members`)
- **`queue-show`** - YES (reads `containers`, `container_members`)
- **`queue-push/pop`** - YES (reads `containers` to validate)
- **`queue-create/rename/rm`** - YES (reads `containers`, `container_kinds`)
- **`stack-ls`** - YES (reads `containers`, `container_members`)
- **`stack-show`** - YES (reads `containers`, `container_members`)
- **`stack-push/pop`** - YES (reads `containers` to validate)
- **`stack-create/rename/rm`** - YES (reads `containers`, `container_kinds`)
- **`group-ls`** - YES (reads `containers`, `container_members`)
- **`group-show`** - YES (reads `containers`, `container_members`)
- **`group-addtask/rmtask`** - YES (reads `containers` to validate)
- **`group-create/rename/delete`** - YES (reads `containers`, `container_kinds`)

#### Schema commands
- **`schema-ls`** - YES (reads `item_kinds`, `container_kinds`)
- **`schema-export`** - YES (reads `container_kinds`, `item_kinds`)

#### Write-only commands (neither read projection tables nor use reducer)
These commands only write events to the event log:
- `edit` - Modifies task fields
- `mv` - Relocates tasks between projects
- `rm` - Deletes tasks
- `describe` - Sets task descriptions
- `project-create` - Creates new projects
- `project-rename` - Renames projects
- `schema-add` - Defines new schemas
- `relate-add` - Adds relationships
- `relate-rm` - Removes relationships
- `relate-dup` - Marks duplicates

## Key Observations

1. **Most read commands use projection tables**: Almost all commands that display information read from projection tables.

2. **Container operations heavily rely on projection tables**: All queue/stack/group operations read from `containers` and `container_members` tables.

3. **The reducer is primarily used for**:
   - Task status and relationship queries (`ls`, `show`, `mark`)
   - Relationship analysis (`relate-*` commands)
   - Complex filtering and queries
   - Computing derived state (blocked status, relationships)

4. **Common projection table reads**:
   - `task_numbers` - Used by `RenderTaskDisplayID` helper (many commands)
   - `tasks` - Read for task titles, project UIDs, validation
   - `projects` - Read for project metadata, names
   - `containers` + `container_members` - Read by all container commands

5. **Hybrid approach**: Some commands like `ls`, `show`, `mark`, `query` use BOTH:
   - Reducer for complex state (status, relationships, filtering)
   - Projection tables for simple lookups (display IDs, project names)

## Conclusion

**Answer to the original question**: Yes, many "normal" commands read from projection tables, not just write to them:

- **Task commands** that read projections: `ls`, `new`, `show`, `mark`, `note`, `history`, `query`
- **Project commands** that read projections: `project-ls`, `project-rm`
- **Container commands** that read projections: ALL queue/stack/group commands
- **Schema commands** that read projections: `schema-ls`, `schema-export`

The only category of commands that doesn't read projection tables is the **relationship commands** (`relate-*`), which exclusively use the reducer for their data needs.
