# Agent Guidelines for This Monorepo

This document contains guidelines for AI agents and automated tools working on projects in this monorepo.

## Backwards Compatibility Policy

**Unless explicitly stated otherwise, backwards compatibility is NOT important for ANY project in this repository.**

All tools and projects in this monorepo (`diagram-dsl`, `dissect`, `markdown-format`, `want`, and any future projects) are work in progress and do not have users yet. Breaking changes are acceptable and encouraged if they improve the API or implementation.

### What This Means

When making changes to **any project** in this monorepo:
- ❌ **Do NOT** maintain old APIs for compatibility
- ❌ **Do NOT** add migration guides
- ❌ **Do NOT** add "backwards compatible" claims in documentation
- ❌ **Do NOT** worry about breaking changes
- ✅ **DO** focus on making the best possible API and implementation
- ✅ **DO** update documentation to reflect current state only
- ✅ **DO** remove deprecated/old code completely

### Examples from Past PRs

Here are concrete examples from previous pull requests showing what **NOT** to do:

#### ❌ Example 1: Unnecessary Migration Guide (PR #6)

**Project:** diagram-dsl  
**File:** `diagram-dsl/IMPROVEMENTS.md`

**What was wrong:**
```markdown
## Migration Guide

No migration needed - all changes are backward compatible.

To use new features:
1. `npm install` to get canvas package
2. Use `renderToSVGWithLayout()` for testing
3. Import `LayoutAssertions` for layout tests
4. Run `npm test` to see new tests in action
```

**Why it's wrong:**
- The library has no users yet
- No one needs migration instructions
- This is just extra documentation that needs to be maintained
- Adds confusion about whether compatibility matters

**What to do instead:**
- Don't add migration sections
- Just document how to use the current version

#### ❌ Example 2: Backwards Compatibility Claims (PR #6)

**Project:** diagram-dsl  
**File:** `diagram-dsl/IMPROVEMENTS.md`

**What was wrong:**
```markdown
### Backward Compatibility

All existing APIs remain unchanged:
- `renderToSVG()` still returns a string (uses `renderToSVGWithLayout()` internally)
- All component props remain the same
- All examples work without modifications
```

**Why it's wrong:**
- No users exist to care about API changes
- This adds unnecessary constraints to future development
- Creates false impression that we need to maintain compatibility

**What to do instead:**
- Just document the current API
- If you change APIs, just update the docs

#### ❌ Example 3: Backwards Compatible Claims (PR #8)

**Project:** diagram-dsl  
**File:** `diagram-dsl/SEMANTIC_COMPONENTS_SUMMARY.md`

**What was wrong:**
```markdown
## Testing

All existing tests pass (14/14):
- 7 SVG rendering tests
- 7 layout assertion tests

No breaking changes - backwards compatible.
```

**Why it's wrong:**
- Again, no users to worry about
- "backwards compatible" doesn't add value
- Could discourage making better changes in the future

**What to do instead:**
- Just report that tests pass
- Don't make claims about compatibility

#### ❌ Example 4: Defensive Statements About Backwards Compatibility (PR #6)

**Project:** diagram-dsl  
**File:** `diagram-dsl/AUDIT.md`

**What was wrong:**
```markdown
### ❌ Not Backward Compatibility Code
The old estimation logic is completely replaced - there's no backward compatibility code lingering.
```

**Why it's wrong:**
- Having to defend that there's no backwards compatibility code suggests we're worried about it
- This shouldn't even be a consideration for work-in-progress projects

**What to do instead:**
- Don't mention backwards compatibility at all
- Just describe what was done

### Owner's Direct Feedback

From the repository owner (PR #8, comment from @neongreen):

> "remove the migration guide and any code left for backwards compatibility, this library is still work in progress and it doesn't have any users right now and no compatibility is needed"

### When Backwards Compatibility DOES Matter

The owner will explicitly request backwards compatibility when needed. Until then, assume it doesn't matter.

### Summary

- **Default assumption:** Backwards compatibility is NOT needed for any project
- **Only consider it when:** The owner explicitly asks for it
- **Focus on:** Making the best possible code, not maintaining old code
- **Applies to:** ALL projects in this monorepo (diagram-dsl, dissect, markdown-format, want, etc.)

## Postmortem Requirements

When a bug or issue is discovered after implementation (especially during code review), agents must create a postmortem analysis documenting:

1. **Timeline**: Chronological sequence of events:
   - What was documented/claimed in the implementation
   - What the reviewer found (the actual bug)
   - What tests were missing

2. **Root Cause**: Why the issue occurred

3. **Prevention Measures**: At least one concrete way this could have been caught earlier

4. **Location**: 
   - For project-specific issues: Add to `<project>/AGENTS.md`
   - For cross-cutting concerns: Add to this global `AGENTS.md`

### Example Format

```markdown
### Postmortem: [Brief Title] (YYYY-MM-DD)

**Timeline:**
1. [Initial implementation details]
2. [Review finding]
3. [Missing tests or verification]
4. [Fix applied]

**Root Cause:**
- [Why it happened]

**What Could Have Caught This Earlier:**
1. [Specific action or check]
2. [Another preventive measure]

**Lessons Learned:**
- [Key takeaway 1]
- [Key takeaway 2]
```

### When to Create Postmortems

- When documented functionality doesn't work as claimed
- When tests don't cover documented features
- When assumptions about libraries/APIs are proven wrong
- When edge cases are missed in initial implementation
- When reviewer finds bugs that should have been caught

The goal is continuous improvement: learn from mistakes and build better practices for future work.
