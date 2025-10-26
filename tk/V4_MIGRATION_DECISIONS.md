# spec-v4 Migration: Decisions Required

This document lists the key decisions needed to proceed with the v4 migration. See [MIGRATION_V4_STRATEGY.md](./MIGRATION_V4_STRATEGY.md) for full details.

## Quick Summary

**What v4 Changes**:
- Replaces **prefixes** with **projects** (with stable UIDs)
- Replaces **prefix names** with **project aliases** (per-node, may collide)
- Makes **task numbers** into mutable labels (not part of identity)
- Changes task display from `prefix-number-node` to `alias-number` (or `alias-number-nodehint` for collisions)

**Impact**: This is a **breaking change** to the data model, event schema, and CLI.

---

## Critical Decisions

### 1. Migration Path

Which migration approach should we use?

**[ ] Option A: Big Bang Migration**
- All nodes stop sync, migrate together, restart sync with v4
- Pros: Clean, no dual-write complexity
- Cons: Requires coordination, downtime
- Timeline: ~3 weeks
- Best for: Small number of nodes, can coordinate

**[ ] Option B: Rolling Migration**
- Gradual migration, nodes sync to separate v4 space during transition
- Dual-write to both v1 and v4 schemas during transition
- Pros: No downtime, can pause/rollback
- Cons: Complex, separate sync space, dual-write overhead
- Timeline: ~7 weeks
- Best for: Many nodes, can't coordinate downtime

**[ ] Option C: Hard Break (RECOMMENDED)**
- Release v4 as breaking change with migration tool
- Provide export/import: `tk migrate export-v1` → `tk migrate import-v1`
- Users manually migrate or start fresh
- Pros: Simplest implementation, clean break, fastest
- Cons: Manual migration effort
- Timeline: ~3 weeks
- Best for: Work-in-progress tools (per AGENTS.md: "backwards compatibility is NOT important")

**Recommendation**: **Option C (Hard Break)** based on repo policy

---

### 2. Project UID Generation

When converting existing prefixes to projects, how should we generate project UIDs?

**[ ] Random ULID (RECOMMENDED)**
- Emit `project.created` events during migration
- Let sync distribute them to other nodes
- Simple, follows normal event flow
- Cons: Requires coordination (all nodes must run backfill)

**[ ] Deterministic (hash of prefix + creating node)**
- Same prefix → same project_uid on all nodes
- No need to sync project creation events
- Cons: Complex, conflicts if multiple nodes claim to have "created" a prefix

**[ ] Manual**
- User explicitly creates projects
- Tool interactively maps old prefixes to new projects
- Cons: Tedious for many prefixes

**Recommendation**: **Random ULID** (simpler, follows event-sourcing model)

---

### 3. Number Collision Handling

When backfilling, multiple nodes may have created the same `(prefix, number)`. What should we do?

**[ ] Keep All (RECOMMENDED - per spec-v4)**
- Allow collisions in `task_numbers` table
- Display with node hints: `tk-1-abc`, `tk-1-def`
- User can manually renumber: `tk number set tk-1-def 2`
- Pros: True to spec-v4, no data loss
- Cons: May confuse users initially

**[ ] Auto-Renumber**
- Detect collisions during backfill
- Automatically reassign numbers to avoid collisions
- Emit `task.number.set` events
- Pros: No collisions after migration
- Cons: Changes task numbers, may break external references

**[ ] Warn and Require Manual Resolution**
- Detect collisions, halt migration
- User resolves manually before proceeding
- Cons: Tedious, blocks migration

**Recommendation**: **Keep All** (spec-v4 is designed to handle this)

---

### 4. Backward Compatibility for v1 Task IDs

After migration, how long should old v1-style task IDs (`tk-1-abc123`) work?

**[ ] Phase Out Immediately**
- After migration, only v4 IDs work (`tk-1`, `tk-1-abc` for collisions)
- Old IDs in external docs/links break
- Pros: Clean break, simpler code
- Cons: Breaks existing references

**[ ] Support Forever (RECOMMENDED)**
- Maintain task aliases forever
- Old IDs resolve to current task
- Pros: Links don't break, gradual transition
- Cons: More complex resolver

**[ ] Support for 6-12 Months**
- Deprecation period, then remove
- Pros: Balance between clean break and compatibility
- Cons: Need to track deprecation timeline

**Recommendation**: **Support Forever** (low cost, high user value)

---

### 5. Sync Strategy During Migration

If not choosing Hard Break, how should sync work during transition?

**[ ] Block Mixed-Version Sync**
- Add version check in sync code
- Nodes must all upgrade before any can sync v4 events
- Clear error message if versions mismatch
- Pros: Prevents corruption
- Cons: Requires coordination

**[ ] Separate Sync Space**
- v4 nodes sync to `tk-events-v4/` folder
- v1 nodes continue syncing to `tk-events/`
- After all migrated, deprecate old space
- Pros: Can migrate incrementally
- Cons: Requires manual sync space management

**Recommendation**: If not Hard Break, use **Separate Sync Space**

---

## Nice-to-Have Decisions

### 6. Display Format for Collisions

When displaying colliding task numbers, what format?

- **[ ]** `alias-number-node` (e.g., `tk-1-abc123`) — matches current format
- **[ ]** `alias-number-abc` (e.g., `tk-1-abc`) — shorter node hint
- **[ ]** `alias-number@user` (e.g., `tk-1@alice`) — human-friendly

**Recommendation**: `alias-number-abc` (shorter, still unique)

---

### 7. Number Reassignment Command

What command to change a task's number?

- **[ ]** `tk number set <task> <N>` (per spec-v4)
- **[ ]** `tk renumber <task> <N>` (more explicit)
- **[ ]** `tk mv <task> <alias>:<N>` (reuse existing mv syntax, but change semantics)

**Recommendation**: `tk number set` (follows spec-v4)

---

### 8. External Projects

When to implement GitHub/Linear/Jira project support?

- **[ ]** Phase 1 (with v4 migration)
- **[ ]** Phase 2 (after v4 stable, 1-2 months)
- **[ ]** Phase 3 (much later, 6+ months)

**Recommendation**: **Phase 2** (get v4 stable first)

---

### 9. Alias Uniqueness

Can a node have multiple aliases for the same project?

- **[ ]** Yes (per spec, aliases can collide)
- **[ ]** No (enforce unique alias per project per node)

**Recommendation**: **Yes** (follows spec-v4, more flexible)

---

## My Recommendations (Quick Start)

If you want to proceed quickly and align with repo policies:

1. **Migration Path**: Option C (Hard Break)
2. **Project UID Generation**: Random ULID
3. **Number Collisions**: Keep All (per spec)
4. **Backward Compat**: Support Forever (via aliases)
5. **Sync Strategy**: N/A (hard break)
6. **Display Format**: `alias-number-abc`
7. **Renumber Command**: `tk number set`
8. **External Projects**: Phase 2
9. **Alias Uniqueness**: Yes (allow duplicates)

**Estimated Timeline**: 3 weeks for full implementation + testing + docs

---

## Next Steps

1. Review both documents:
   - [MIGRATION_V4_STRATEGY.md](./MIGRATION_V4_STRATEGY.md) - Full technical details
   - This document - Decision checklist

2. Make decisions on Critical Decisions (#1-5)

3. Confirm or adjust Nice-to-Have decisions (#6-9)

4. I'll create a detailed implementation plan based on your choices

5. Begin implementation

---

## Questions?

If you need clarification on any decision:
- Ask about trade-offs for specific options
- Ask for examples of how a decision affects users
- Ask about implementation complexity for any option
- Ask about testing strategies

Just let me know which decisions you've made, or if you want to discuss any option in more detail.
