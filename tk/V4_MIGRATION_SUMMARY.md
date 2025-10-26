# tk v4 Migration Strategy - Executive Summary

## What I've Delivered

I've analyzed the spec-v4-projects.md document and created a comprehensive migration strategy. Here are the deliverables:

### 📄 Documents Created

1. **[MIGRATION_V4_STRATEGY.md](./MIGRATION_V4_STRATEGY.md)** (19KB)
   - Full technical migration strategy
   - Detailed comparison of current vs. target state
   - Three migration path options (Big Bang, Rolling, Hard Break)
   - Implementation phases breakdown
   - Risk analysis and mitigation
   - Testing strategy
   - Timeline estimates

2. **[V4_MIGRATION_DECISIONS.md](./V4_MIGRATION_DECISIONS.md)** (7KB)
   - Quick summary of v4 changes
   - 9 key decisions you need to make (5 critical, 4 nice-to-have)
   - My recommendations for each
   - Next steps

3. **This summary** (you're reading it)

---

## TL;DR - What v4 Changes

**Current (v1/v2)**:
- Tasks organized by **prefixes** (e.g., `tk`, `foo`, `bar`)
- Task IDs: `prefix-number-node` (e.g., `tk-1-abc123`)
- Prefixes are first-class, numbers are per-(prefix,node) counters

**Target (v4)**:
- Tasks organized by **projects** (with stable UIDs like `prj_01J5Q...`)
- Project **aliases** per node (e.g., `tk`, `backend`) - may collide across nodes
- Task IDs: **derived display strings** (e.g., `tk-1`, or `tk-1-abc` if collision)
- Task identity: **task_uid** (ULID, e.g., `tsk_01J5Q...`)
- Numbers are **mutable labels**, not identity - collisions allowed

**Impact**: 🔴 **Breaking change** - new data model, event schema, CLI commands

---

## My Recommendation

Based on the repo's **"backwards compatibility is NOT important"** policy:

**Go with Option C: Hard Break Migration**

**Why**:
- ✅ Simplest to implement (~3 weeks vs ~7 weeks for rolling migration)
- ✅ Cleanest code (no dual-write complexity)
- ✅ Easiest to test and maintain
- ✅ Aligns with repo policy on breaking changes
- ✅ tk is work-in-progress, not production-critical yet

**How it works**:
1. Implement v4 schema + event handlers + CLI
2. Provide migration tool: `tk migrate export-v1` → `tk migrate import-v1`
3. Users export their v1 data, import into v4, or start fresh
4. Ship as v4.0 with clear breaking change notice

**Timeline**: ~3 weeks (vs 7 weeks for rolling migration)

---

## Decisions I Need From You

### Critical (Must Decide)

1. **Migration Path**: Which option?
   - **My rec**: Option C (Hard Break)
   - Alternatives: Option A (Big Bang), Option B (Rolling)

2. **Project UID Generation**: How to convert prefixes → projects?
   - **My rec**: Random ULID + emit events
   - Alternatives: Deterministic hash, Manual

3. **Number Collision Handling**: What if two nodes both have `tk-1`?
   - **My rec**: Keep both (per spec-v4), display as `tk-1-abc` and `tk-1-def`
   - Alternatives: Auto-renumber, Require manual resolution

4. **Backward Compat**: Support old v1 task IDs forever?
   - **My rec**: Yes (via aliases)
   - Alternatives: Phase out immediately, Support for 6-12 months

5. **Sync Strategy** (only if not Hard Break):
   - **My rec**: Separate sync space for v4
   - Alternative: Block mixed-version sync

### Nice-to-Have (Can Use Defaults)

6. Display format for collisions: `tk-1-abc` (my rec)
7. Renumber command: `tk number set` (my rec)
8. External projects: Phase 2 (after v4 stable)
9. Alias uniqueness: Allow duplicates (my rec)

---

## What Happens Next

### If You Agree With My Recommendations

Just say "go ahead with your recommendations" and I'll:

1. Create detailed implementation plan (task breakdown)
2. Start implementation:
   - Week 1: v4 schema + event handlers + reducer
   - Week 2: CLI commands + display renderer + migration tool
   - Week 3: Testing + documentation + integration
3. Deliver v4 implementation ready to ship

### If You Want to Discuss Options

Tell me which decisions you want to explore, and I'll:
- Explain trade-offs in more detail
- Show concrete examples of how each option works
- Discuss implementation complexity differences
- Adjust the plan based on your choices

### If You Have Questions

Ask anything about:
- Why I recommended certain options
- How a specific decision affects users/implementation
- Technical details of any migration path
- Examples of how things would work in v4

---

## Key Insights from Analysis

### What I Found

1. **v4 is a fundamental redesign**, not just an incremental change
   - Shifts from prefix-centric to project-centric model
   - Changes identity model (numbers become labels)
   - Enables future external project integration (GitHub, Linear, Jira)

2. **Current implementation (v2) is solid**
   - Good event-sourcing foundation
   - Task UUIDs already provide stable identity
   - Sync system works well
   - Easy to build v4 on top of this foundation

3. **Migration complexity varies widely**
   - Hard Break: 3 weeks, simple, clean
   - Rolling Migration: 7 weeks, complex, but no downtime
   - Big Bang: 3 weeks, but requires coordination

4. **Spec-v4 is well-designed**
   - Handles offline conflicts elegantly (collision tolerance)
   - Supports distributed scenarios (per-node aliases)
   - Future-proofs for external integrations
   - Follows event-sourcing principles

### Risks Identified

- 🔴 **Sync conflicts** if mixed v1/v4 nodes (mitigated by Hard Break or separate sync space)
- 🟡 **Number collisions** after migration (mitigated by v4's collision tolerance)
- 🟡 **Alias collisions** across nodes (mitigated by per-node scoping)
- 🟢 **Data loss** during migration (mitigated by export/backup tools)

All risks have clear mitigation strategies documented.

---

## Example: What Changes for Users

### Before (v1/v2)
```bash
# Create prefix
tk prefix create foo "Foo project tasks"

# Create task
tk new --prefix foo "Implement feature X"
# → Creates foo-1-abc123

# Move task
tk mv foo-1 bar:1
# → Changes to bar-1-abc123, foo-1 becomes alias

# List tasks
tk ls --prefix foo
```

### After (v4)
```bash
# Create project
tk project create "Foo project tasks"
# → Creates project with UID prj_01J5Q...

# Create alias
tk project alias foo prj_01J5Q...
# → Maps "foo" to project on this node

# Create task
tk new --project foo "Implement feature X"
# → Creates task with UID tsk_01J5Q..., displays as foo-1

# Change number
tk number set foo-1 2
# → Renumbers to foo-2, foo-1 may become available again

# List tasks
tk ls --project foo
```

Key differences:
- `prefix` → `project` (with stable UIDs underneath)
- `mv` → `number set` (renumbering, not moving)
- Display IDs are derived, not stored

---

## Files to Review

1. **Start here**: [V4_MIGRATION_DECISIONS.md](./V4_MIGRATION_DECISIONS.md)
   - Quick checklist of decisions
   - My recommendations
   - Easy to read in 5 minutes

2. **Deep dive**: [MIGRATION_V4_STRATEGY.md](./MIGRATION_V4_STRATEGY.md)
   - Full technical details
   - All three migration paths explained
   - Risk analysis, testing strategy, timelines
   - Read if you want to understand implementation details

3. **Reference**: [specs/spec-v4-projects.md](./specs/spec-v4-projects.md)
   - The original spec (you asked me to read this)
   - I analyzed this to create the migration strategy

---

## Bottom Line

✅ **Migration is feasible** - I've mapped out a clear path

✅ **Recommended approach is simple** - Hard Break with migration tool

✅ **Timeline is reasonable** - 3 weeks for recommended approach

⏸️ **Waiting on your decisions** - Review V4_MIGRATION_DECISIONS.md and let me know

🚀 **Ready to implement** - Once you decide, I can start immediately

---

## Questions for You

1. Have you reviewed [V4_MIGRATION_DECISIONS.md](./V4_MIGRATION_DECISIONS.md)?
2. Do my recommendations (Option C + other defaults) sound good?
3. Do you want to discuss any specific decision in more detail?
4. Should I proceed with implementation, or do you want to adjust the plan?

Let me know how you'd like to proceed!
