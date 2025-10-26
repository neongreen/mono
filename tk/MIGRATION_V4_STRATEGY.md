# tk spec-v4 Migration Strategy

## Executive Summary

This document outlines the migration strategy from tk's current implementation (v1/v2 with prefixes) to spec-v4 (unified project + task model). The migration is **substantial** as it fundamentally changes the identity and organizational model from prefixes to projects.

## Current State (v1/v2)

### What We Have
- **Prefixes**: First-class organizational entities (e.g., `tk`, `foo`, `bar`)
  - Each prefix has its own namespace
  - Prefix metadata stored per-node (prefix, description, created_by, created_at)
  - Counters are per (prefix, node) pair
  - Prefixes can be discovered, explicit, or removed
- **Task IDs**: `<prefix>-<number>-<node>` (e.g., `tk-1-abc123`)
  - Prefix component is mutable (via `tk mv`)
  - Number component is mutable (via `tk mv`)
  - Node component is immutable
- **Task UUIDs**: Immutable canonical identifiers (e.g., `task-abc123xyz...`)
- **Task Aliases**: Old IDs preserved when tasks are moved
- **Event Types**:
  - `task.created`, `task.status.set`, `task.note.add`
  - `task.reprefix`, `task.alias.added`
  - `prefix.created`, `prefix.removed`, `prefix.description.set`, `prefix.alias.added`
  - `relation.add`, `relation.remove`, `relation.note` (v2)
- **Relations (v2)**: blocks, subtasks, related, duplicates, supersedes
- **Sync (v1)**: Immutable segment files with folder remotes

### Database Schema
```sql
-- Current v1/v2 schema
events (id, ts, created_at, actor, role, kind, payload, ctx, ...)
prefixes (prefix, node, description, created_at, created_by, removed)
prefix_counters (prefix, node, last_id)
task_counter (last_id)  -- legacy
event_counter (last_id)
event_id_map (rowid, event_id)
metadata (key, value)
```

## Target State (spec-v4)

### What We Need
- **Projects**: First-class entities with stable UIDs
  - `project_uid` (ULID, e.g., `prj_01J5Q...`)
  - `type` (local, github, linear, jira)
  - `origin` (structured metadata, e.g., `{"owner":"neongreen","repo":"mono"}`)
  - `name`, `description`, `created_at`, `created_by`
- **Project Aliases**: Per-node short names (e.g., `tk`, `backend`, `mono`)
  - `(alias, node)` → `project_uid`
  - Aliases may collide freely across nodes
- **Task IDs**: Display strings, not stored identifiers
  - Internal: `task_uid` (ULID, e.g., `tsk_01J5Q...`)
  - Display: `<alias>-<number>` or `<alias>-<number>-<node_hint>` for collisions
  - External: `owner/repo#number` for GitHub, etc.
- **Task Numbers**: Mutable labels, not identity
  - `(project_uid, number)` → `task_uid` (non-unique, may have collisions)
  - Numbers can be reassigned via `task.number.set` events
- **Event Types** (new/changed):
  - `project.created`
  - `project.alias.add`, `project.alias.remove`
  - `task.created` (with `project_uid`, optional `proposed_number`)
  - `task.number.set` (reassign/change task number)

### Database Schema (target)
```sql
-- Target v4 schema
events (id, ts, created_at, actor, role, kind, payload, ctx, ...)
projects (project_uid PK, type, origin_json, name, description, created_at, created_by)
project_aliases (project_uid, alias, node, added_by, PRIMARY KEY(alias, node))
tasks (task_uid PK, project_uid, created_node, title, created_at, created_by, ...)
task_numbers (project_uid, number, task_uid)  -- non-unique, may have duplicates
-- ... rest unchanged (event_counter, metadata, etc.)
```

## Key Conceptual Changes

### 1. Identity Model
| Aspect | Current (v1/v2) | Target (v4) |
|--------|----------------|-------------|
| Project identity | Prefix string (mutable via `reprefix`) | `project_uid` (immutable ULID) |
| Project naming | Prefix (2-20 chars, global scope) | Alias (per-node, may collide) |
| Task identity | Task UUID (immutable) | `task_uid` (immutable ULID) |
| Task display ID | `prefix-number-node` (partially mutable) | Derived from `alias-number` (+node hint if needed) |
| Task numbering | Counter per (prefix, node) | Label per (project_uid), may collide |

### 2. Organizational Model
- **Current**: Prefixes are the organizational unit, scoped to nodes
- **Target**: Projects are the organizational unit, aliases are scoped to nodes

### 3. Number Semantics
- **Current**: Numbers are quasi-stable (change only on explicit `mv`)
- **Target**: Numbers are fully mutable labels, collisions allowed and resolved at display time

## Breaking Changes

### Data Model
1. **Prefix → Project**: Prefixes become projects, but the semantics change significantly
2. **Task IDs**: Display format changes from `prefix-number-node` to `alias-number` (or `alias-number-nodehint`)
3. **Numbering**: Move from per-(prefix,node) counters to per-project labels with collision tolerance

### Event Schema
1. **New events**: `project.created`, `project.alias.add/remove`, `task.number.set`
2. **Changed events**: `task.created` now references `project_uid` and includes `proposed_number`
3. **Deprecated events**: `prefix.created`, `prefix.removed`, `task.reprefix` (replaced by `task.number.set`)

### CLI Commands
1. **Removed**: `tk prefix create/list/remove`
2. **New**: `tk project create/list`, `tk project alias add/remove`, `tk number set`
3. **Changed**: `tk new` now uses `--project <selector>` instead of `--prefix`
4. **Changed**: `tk mv` fundamentally changes (now emits `task.number.set` instead of `task.reprefix`)

### Sync Compatibility
- **v4 segments are NOT compatible with v1/v2 databases**
- Mixed-version sync will cause issues
- Migration must be coordinated across all nodes

## Migration Phases

### Phase 1: Add v4 Schema Alongside v1/v2 (Dual-Write)
**Goal**: Support both data models simultaneously during transition

**Changes**:
1. Add new tables: `projects`, `project_aliases`, `task_numbers`
2. Keep existing tables: `prefixes`, `prefix_counters`
3. Dual-write to both schemas:
   - When `prefix.created` event → also create project + alias
   - When `task.created` event → also write to `task_numbers`
4. Add migration flag in config: `migration.v4_mode = "dual_write"`

**Testing**:
- Verify both schemas stay in sync
- Test round-trip: v1 event → v4 projection → v1 display

### Phase 2: Backfill Historical Data
**Goal**: Convert existing prefixes → projects, populate task_numbers

**Process**:
1. Run migration command: `tk migrate v4-backfill`
2. For each prefix in `prefixes`:
   - Generate `project_uid` (deterministic from prefix+node? or random?)
   - Create `project.created` event
   - Create `project.alias.add` event mapping old prefix name → project_uid
3. For each task:
   - Extract (prefix, number) from current task_id
   - Resolve prefix → project_uid
   - Create `task_numbers` entry (project_uid, number, task_uid)
4. Verify: All tasks resolvable via new schema

**Decisions Needed**:
- **Q1**: How to generate project_uid for existing prefixes?
  - Option A: Deterministic (hash of prefix+node) — reproducible but complex
  - Option B: Random ULID per node — simpler, requires explicit sync of backfill events
  - **Recommendation**: Option B (emit `project.created` events, let sync handle it)

### Phase 3: v4 Read Mode
**Goal**: Start using v4 schema for reads, keep v1 for compatibility

**Changes**:
1. Set config: `migration.v4_mode = "read_v4"`
2. CLI commands use v4 schema for queries:
   - `tk ls` reads from `task_numbers` + `project_aliases`
   - `tk view` renders using v4 display format
3. CLI commands still accept v1-style IDs for backwards compat
4. Event emission: Still dual-write both v1 and v4 events

**Testing**:
- All queries work with v4 schema
- v1-style task IDs still resolve correctly

### Phase 4: v4 Write Mode (Breaking Change)
**Goal**: Switch to v4-only event emission

**Changes**:
1. Set config: `migration.v4_mode = "write_v4"`
2. Stop emitting v1 events:
   - `tk prefix create` → `project.created` + `project.alias.add` only
   - `tk new` → `task.created` with `project_uid`
   - `tk mv` → `task.number.set` (not `task.reprefix`)
3. Update CLI:
   - `tk prefix` commands replaced by `tk project`
   - `tk mv` semantics change to number reassignment
4. **WARNING**: Segments generated in this phase are NOT compatible with v1/v2 nodes

**Coordination Required**:
- All nodes must upgrade to Phase 4 before syncing
- Or: Use separate sync space for v4 (`tk-events-v4/`)

### Phase 5: v4 Cleanup
**Goal**: Remove v1/v2 code and schema

**Changes**:
1. Drop tables: `prefixes`, `prefix_counters`, `task_counter`
2. Remove v1 event handlers from reducer
3. Remove v1 CLI commands
4. Set config: `migration.v4_mode = "native_v4"` (or remove flag entirely)

**Timeline**: After all nodes migrated, 30+ day grace period

## Migration Path Options

### Option A: Big Bang Migration
**Approach**: All nodes stop sync, migrate together, restart sync with v4

**Steps**:
1. Coordinate downtime window with all users
2. Stop sync on all nodes
3. Run `tk migrate v4` on each node (includes all phases 1-4)
4. Verify migrations successful
5. Update sync space or clear old segments
6. Resume sync

**Pros**: Clean, no dual-write complexity, no mixed-version issues
**Cons**: Requires coordination, potential data loss if migration fails

### Option B: Rolling Migration with Separate Sync Space
**Approach**: Gradual migration, v4 nodes sync to new space

**Steps**:
1. Deploy Phase 1-2 to all nodes (dual-write, backfill)
2. Let sync distribute backfill events
3. Each node independently upgrades to Phase 3 (read v4)
4. Coordinate switch to Phase 4:
   - Create new sync remote: `tk remote add icloud-v4 folder ~/iCloud/tk-events-v4`
   - Switch config to write_v4 mode
   - Run `tk export --all icloud-v4`
5. After all nodes migrated, deprecate old sync space

**Pros**: Gradual, can pause/rollback, less risky
**Cons**: Complex, requires separate sync space, dual-write overhead

### Option C: Hard Break (Recommended for Early Stage)
**Approach**: Treat v4 as a new system, optional manual migration

**Steps**:
1. Release v4 as breaking change
2. Provide migration tool: `tk migrate export-v1` → JSON dump
3. Provide migration tool: `tk migrate import-v1 <dump.json>` (rewrites as v4 events)
4. Users manually export from v1, import to v4
5. Or: Users start fresh with v4

**Pros**: Simplest implementation, clean break, no dual-write
**Cons**: Manual migration effort, can't incrementally adopt

## Decisions Needed from User

### Critical Decisions

1. **Migration Path**: Which option?
   - [ ] Option A: Big Bang (coordinate all nodes)
   - [ ] Option B: Rolling (separate sync space)
   - [ ] Option C: Hard Break (recommended if tk is not yet production-critical)

2. **Project UID Generation**: For backfilling prefixes → projects
   - [ ] Deterministic (hash of prefix + node)
   - [ ] Random ULID (emit events, sync distributes)
   - [ ] Manual (user creates projects, tool maps prefixes → projects)

3. **Number Collision Strategy**: When backfilling, what if multiple nodes have same (prefix, number)?
   - [ ] Keep all as separate task_numbers entries (per spec-v4)
   - [ ] Auto-renumber to avoid collisions
   - [ ] Warn and require manual resolution

4. **Backward Compatibility**: How long to support v1 task IDs?
   - [ ] Phase them out in Phase 4 (immediate)
   - [ ] Support forever via alias resolution
   - [ ] Support for N months, then deprecate

5. **Sync Strategy**: 
   - [ ] Require all nodes migrate before any node syncs v4 events
   - [ ] Use separate sync space (icloud-v4) during transition
   - [ ] Accept that mixed-version sync will break (hard break)

### Nice-to-Have Decisions

6. **Display Format**: For collision resolution
   - [ ] `alias-number-node` (spec suggests last 6 chars of node ULID)
   - [ ] `alias-number-abc` (current 6-char node ID)
   - [ ] `alias-number-@user` (user who created task)

7. **Alias Management**: Can users have multiple aliases for same project?
   - [ ] Yes (per spec, aliases are per-node and can collide)
   - [ ] No (enforce uniqueness per node)

8. **External Projects**: When to implement?
   - [ ] Phase 1 (GitHub integration)
   - [ ] Phase 2 (after v4 stable)
   - [ ] Phase 3 (much later)

9. **Number Reassignment UI**:
   - [ ] `tk number set <task> <N>` (per spec)
   - [ ] `tk renumber <task> <N>` (more explicit)
   - [ ] `tk mv <task> <alias>:<N>` (reuse mv syntax)

## Implementation Complexity

### Low Complexity (1-2 days)
- Add v4 schema tables
- Implement `project.created`, `project.alias.add` event handlers
- Implement `task.number.set` event handler
- Add basic projection: projects, aliases, task_numbers

### Medium Complexity (3-5 days)
- Dual-write mode implementation
- Backfill command for existing data
- Update CLI to use v4 schema
- Renderer for v4 display strings (with collision resolution)
- Migration between modes (dual_write → read_v4 → write_v4)

### High Complexity (5-10 days)
- Rolling migration with separate sync spaces
- Compatibility layer for v1 task ID resolution in v4
- Comprehensive testing of all migration paths
- Documentation and migration guides
- Handling edge cases (orphaned tasks, missing projects, etc.)

## Risks & Mitigation

### Risk 1: Data Loss During Migration
**Likelihood**: Medium  
**Impact**: Critical  
**Mitigation**: 
- Backup database before migration: `tk export --all backup`
- Test migration on copy of production data
- Provide rollback procedure

### Risk 2: Sync Conflicts in Mixed-Version Environment
**Likelihood**: High (if not coordinated)  
**Impact**: High  
**Mitigation**:
- Use separate sync space for v4
- Add version check in sync code (block v1 ↔ v4 sync)
- Clear migration coordination docs

### Risk 3: Number Collisions After Migration
**Likelihood**: Medium  
**Impact**: Low (spec-v4 tolerates this)  
**Mitigation**:
- Implement collision detection in display layer
- Provide `tk conflicts` command to show collisions
- Document how to resolve (use `tk number set`)

### Risk 4: Alias Collisions
**Likelihood**: Low (single user) to High (multi-user)  
**Impact**: Medium  
**Mitigation**:
- Document that aliases are per-node and may collide
- Implement clear error messages when alias resolves to multiple projects
- Require full project_uid for ambiguous operations

## Testing Strategy

### Unit Tests
- [ ] v4 event handlers (project.created, task.number.set, etc.)
- [ ] Dual-write mode: events produce both v1 and v4 projections
- [ ] Backfill logic: prefix → project conversion
- [ ] Display rendering: collision detection and node hints

### Integration Tests
- [ ] Migration path: v1 → dual-write → read_v4 → write_v4
- [ ] Round-trip: create task in v4, export, ingest, verify
- [ ] Collision scenarios: two nodes create same project/number
- [ ] Alias resolution: multiple aliases, ambiguous queries

### Sync Tests
- [ ] v4 segments sync correctly
- [ ] Version mismatch detected and blocked
- [ ] Separate sync spaces work independently

### Migration Tests (Critical)
- [ ] Backfill existing database with 100+ tasks, 5+ prefixes
- [ ] Verify all tasks resolvable after migration
- [ ] Verify sync works after migration
- [ ] Test rollback from each phase

## Documentation Requirements

### User-Facing
1. **Migration Guide**: Step-by-step for each migration path
2. **v4 Concepts**: Explain projects, aliases, task numbers
3. **CLI Changes**: Command mapping (v1 → v4)
4. **FAQ**: Common migration issues and solutions

### Developer-Facing
1. **Migration Architecture**: Phase diagrams, data flow
2. **Event Schema Changes**: Before/after comparison
3. **Testing Migration**: How to test locally
4. **Rollback Procedures**: How to revert each phase

## Timeline Estimate

### Aggressive (Hard Break - Option C)
- Week 1: Schema + basic v4 event handlers
- Week 2: CLI commands + renderer + backfill tool
- Week 3: Testing + documentation
- **Total**: 3 weeks

### Conservative (Rolling Migration - Option B)
- Week 1-2: Dual-write implementation + schema
- Week 3: Backfill + migration command
- Week 4: CLI updates + renderer
- Week 5: Sync coordination + version checking
- Week 6-7: Testing + documentation + rollout
- **Total**: 7 weeks

## Recommendation

Given that tk is **work in progress** (per AGENTS.md: "backwards compatibility is NOT important"):

**Recommended Path**: **Option C - Hard Break**

**Rationale**:
1. Cleanest implementation (no dual-write complexity)
2. Fastest to deliver (3 weeks vs 7 weeks)
3. Easier to test and maintain
4. Aligns with repo policy: "breaking changes are acceptable and encouraged"
5. Users can export → import or start fresh

**Recommended Phasing**:
1. Implement v4 schema and event handlers (1 week)
2. Implement backfill tool for migration: `tk migrate to-v4` (3 days)
3. Update all CLI commands to v4 model (4 days)
4. Test migration thoroughly (3 days)
5. Document migration process (2 days)
6. Ship as v4.0 with breaking change notice (1 day)

**Migration Tool Design**:
```bash
# Export current state to JSON
tk migrate export-v1 > tk-backup.json

# Initialize new v4 database (or upgrade existing)
tk migrate to-v4

# Or: import from another node's export
tk migrate import-v1 tk-backup.json
```

## Next Steps

1. **User Decision**: Choose migration path (A/B/C) and answer critical decisions
2. **Implementation Plan**: Create detailed task breakdown for chosen path
3. **Prototype**: Build minimal v4 schema + event handlers
4. **Validate**: Test with subset of real data
5. **Full Implementation**: Complete chosen migration path
6. **Release**: Ship v4 with migration guide

---

## Appendix: Event Mapping (v1 → v4)

### Prefix → Project
```
# v1
prefix.created {prefix, description, created_by}

# v4 (emits TWO events)
project.created {project_uid, type: "local", name, description, created_by}
project.alias.add {project_uid, alias: <prefix>, node, added_by}
```

### Task Creation
```
# v1
task.created {task_uuid, task_id: "tk-1-abc123", title, created_by}

# v4
task.created {task_uid, project_uid, proposed_number: 1, created_node, title, created_by}
(optionally followed by)
task.number.set {task_uid, project_uid, number: 1, reason: "initial"}
```

### Task Move
```
# v1
task.reprefix {task_uuid, old_prefix: "tk", new_prefix: "foo", old_number: 1, new_number: 2, ...}
task.alias.added {task_uuid, alias_id: "tk-1-abc123"}

# v4
task.number.set {task_uid, project_uid, number: 2, reason: "moved from tk-1"}
(number labels are the identity now, no separate move event needed)
(if changing projects, emit another task.number.set with new project_uid)
```

## Appendix: Collision Resolution Examples

### Scenario 1: Two Nodes Create Same Task Number
```
Node A: task.created {task_uid: tsk_A, project_uid: prj_tk, proposed_number: 1}
Node B: task.created {task_uid: tsk_B, project_uid: prj_tk, proposed_number: 1}

After sync, task_numbers table:
(prj_tk, 1, tsk_A)
(prj_tk, 1, tsk_B)

Display:
tk-1-abc     # Node A's task (with node hint)
tk-1-def     # Node B's task (with node hint)

Resolution:
tk number set tsk_B 2   # Renumber one of them
```

### Scenario 2: Alias Collision
```
Node A: project.alias.add {project_uid: prj_personal, alias: "work", node: abc123}
Node B: project.alias.add {project_uid: prj_company, alias: "work", node: def456}

Both nodes have local alias "work" pointing to different projects.

Resolution (per spec): This is allowed! Aliases are per-node.
- Node A's "work" resolves to prj_personal
- Node B's "work" resolves to prj_company
- No conflict, both coexist
```
