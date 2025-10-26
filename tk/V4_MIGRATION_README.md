# tk v4 Migration Documents

This directory contains the complete migration strategy for transitioning tk from v1/v2 (prefix-based) to v4 (project-based) architecture, as specified in [specs/spec-v4-projects.md](./specs/spec-v4-projects.md).

## 📚 Reading Guide

Start here based on your goal:

### 🚀 Just Tell Me What to Do
**→ Start with**: [V4_MIGRATION_SUMMARY.md](./V4_MIGRATION_SUMMARY.md)
- Executive summary (5 min read)
- My recommendations with rationale
- What happens next
- **Best for**: Quick overview, decision-making

### 🎨 Show Me What Changes
**→ Start with**: [V4_VISUAL_COMPARISON.md](./V4_VISUAL_COMPARISON.md)
- Before/after diagrams (10 min read)
- Collision scenarios explained visually
- CLI command mapping
- Migration examples
- **Best for**: Understanding v4 conceptually

### ✅ What Do I Need to Decide?
**→ Start with**: [V4_MIGRATION_DECISIONS.md](./V4_MIGRATION_DECISIONS.md)
- Decision checklist (5 min read)
- 9 key decisions (5 critical, 4 nice-to-have)
- Options and recommendations
- **Best for**: Making decisions, ready to proceed

### 🔬 I Want Technical Details
**→ Start with**: [MIGRATION_V4_STRATEGY.md](./MIGRATION_V4_STRATEGY.md)
- Full technical strategy (30 min read)
- Three migration paths analyzed
- Implementation phases, risks, testing
- Timeline estimates
- **Best for**: Implementation planning, architecture review

## 📊 Quick Comparison

| Current (v1/v2) | Target (v4) |
|----------------|-------------|
| **Prefixes** (e.g., `tk`, `foo`) | **Projects** (e.g., `prj_01J5Q...`) with aliases |
| Task IDs: `prefix-number-node` | Task IDs: **derived** from `alias-number` |
| Identity: task_uuid + task_id | Identity: **task_uid** (ULID) |
| Numbers: per-(prefix,node) counters | Numbers: **mutable labels** (may collide) |
| Events: `prefix.created`, `task.reprefix` | Events: `project.created`, `task.number.set` |

**Impact**: 🔴 **Breaking change** to data model, events, and CLI

## 🎯 My Recommendation

**Hard Break Migration (Option C)**
- Ship v4 as breaking change with migration tool
- Users: export v1 → import v4, or start fresh
- Timeline: **3 weeks** implementation + testing + docs
- Rationale: Simplest, cleanest, aligns with repo policy

## ✨ What I've Done

1. ✅ Read and analyzed [specs/spec-v4-projects.md](./specs/spec-v4-projects.md)
2. ✅ Reviewed current tk implementation (v1/v2)
3. ✅ Mapped current state → target state
4. ✅ Identified breaking changes and migration complexity
5. ✅ Designed three migration path options
6. ✅ Analyzed risks and mitigation strategies
7. ✅ Created implementation timeline estimates
8. ✅ Documented all decisions needed
9. ✅ Provided recommendations based on repo policies

## ⏭️ Next Steps

### For You (User)
1. Read [V4_MIGRATION_SUMMARY.md](./V4_MIGRATION_SUMMARY.md) or [V4_VISUAL_COMPARISON.md](./V4_VISUAL_COMPARISON.md)
2. Review decisions in [V4_MIGRATION_DECISIONS.md](./V4_MIGRATION_DECISIONS.md)
3. Let me know:
   - Do my recommendations sound good?
   - Any decisions you want to discuss?
   - Should I proceed with implementation?

### For Me (Agent)
Once you decide:
1. Create detailed implementation task breakdown
2. Implement v4 schema + event handlers
3. Build migration tool (`tk migrate export-v1` / `import-v1`)
4. Update all CLI commands to v4 model
5. Write comprehensive tests
6. Document migration process
7. Ship v4.0

## 🗂️ File Index

| File | Purpose | Length | Read Time |
|------|---------|--------|-----------|
| [V4_MIGRATION_SUMMARY.md](./V4_MIGRATION_SUMMARY.md) | Executive summary, start here | 8KB | 5 min |
| [V4_VISUAL_COMPARISON.md](./V4_VISUAL_COMPARISON.md) | Visual guide with diagrams | 12KB | 10 min |
| [V4_MIGRATION_DECISIONS.md](./V4_MIGRATION_DECISIONS.md) | Decision checklist | 7KB | 5 min |
| [MIGRATION_V4_STRATEGY.md](./MIGRATION_V4_STRATEGY.md) | Full technical strategy | 19KB | 30 min |
| **This file** | Reading guide and index | 3KB | 2 min |

## 🔑 Key Insights

### Why v4 is Different
- **v1/v2**: Prefixes are organizational units, numbers are counters
- **v4**: Projects have stable UIDs, numbers are mutable labels
- **Enable**: External integrations (GitHub, Linear, Jira)
- **Embrace**: Offline collisions, distributed workflows

### Why Hard Break is Recommended
- ✅ Simplest implementation (3 weeks vs 7 weeks)
- ✅ Cleanest code (no dual-write complexity)
- ✅ Aligns with repo policy ("backwards compatibility is NOT important")
- ✅ tk is work-in-progress, not production-critical

### What Changes for Users
- CLI: `tk prefix` → `tk project`, `--prefix` → `--project`
- Display: `tk-1-abc123` → `tk-1` (or `tk-1-abc` if collision)
- Semantics: Moving tasks → Renumbering tasks
- Events: New project model, mutable task numbers

## 📞 Questions?

If you need clarification:
- "Explain the collision scenario in more detail"
- "What's the implementation effort for Option B?"
- "Show me an example of external project integration"
- "How does the migration tool work?"

Just ask, and I'll provide more details on any aspect!

## 🚦 Status

**Phase**: ✅ Analysis Complete, ⏸️ Awaiting Decisions

**Waiting for**: Your review and decision on migration path

**Ready to**: Start implementation as soon as you decide

---

Generated by GitHub Copilot Agent  
Date: 2025-10-26  
Task: Analyze spec-v4 and develop migration strategy
