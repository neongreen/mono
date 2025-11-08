# Audit Summary: Silent Error Handling in tk

## Overview

This audit was conducted to identify places where warnings and errors are being silently dropped, particularly those that are user-data-critical or business-logic-critical.

## Findings

### 1. ✅ FIXED: Ingest Operations (internal/remote/ingest.go)
**Status**: Already addressed in commit 9e06e07

**Previous Issue**: 
- Segment read failures were silently dropped with `continue`
- Event projection errors were silently dropped

**Fix Applied**:
- Added `SegmentErrors` and `ProjectionErrors` fields to result types
- Errors are now collected and returned to caller
- CLI wrapper displays warnings to stderr

---

### 2. ✅ OK: Sync Operations (internal/remote/sync.go)
**Status**: Already handles errors properly

**Lines 95-96 and 124-125**:
```go
if err := SaveIndexFile(remoteIndexPath, remoteIndex); err != nil {
    // Non-fatal, just log
    fmt.Printf("Warning: failed to save reconstructed index: %v\n", err)
}
```

**Assessment**: These are already printing warnings with `fmt.Printf()`. The errors are non-critical (failure to save a reconstructed index) and users are informed.

---

### 3. ⚠️ NEEDS ATTENTION: Beads Import (internal/import/beads/importer.go)

#### Issue 3a: Silent AddRenumberNote Failures (Line 112-114)
**Severity**: Low (metadata only)

```go
if err := AddRenumberNote(db, taskUID, issue.ID, number); err != nil {
    // Non-fatal, continue
}
```

**Impact**: If adding a renumber note fails, the user won't know. However:
- The actual task data is already imported successfully 
- Renumber notes are informational/metadata
- The issue ID is still tracked in `renumberedIssues` array

**Recommendation**: Add to a warnings array in ImportResult for visibility

#### Issue 3b: Silent Relationship Import Failures (Line 127-130)  
**Severity**: Medium to High (business logic)

```go
count, err := ImportBeadsRelationships(db, issue, taskUID, issueMap)
if err != nil {
    // Non-fatal, continue
    continue
}
```

**Impact**: This is more concerning:
- Relationships between tasks are business-critical data
- Users won't know if relationships failed to import
- Silent data loss of potentially important dependency information

**Recommendation**: Track failed relationships and report them to users

**Suggested Fix**:
```go
// Add to ImportResult type:
type ImportResult struct {
    // ... existing fields ...
    FailedNotes        []string // Failed renumber notes  
    FailedRelationships []string // Failed relationship imports
}

// In the import loop:
var failedNotes []string
var failedRelationships []string

// For renumber notes:
if err := AddRenumberNote(db, taskUID, issue.ID, number); err != nil {
    failedNotes = append(failedNotes, fmt.Sprintf("task %s: %v", issue.ID, err))
}

// For relationships:
count, err := ImportBeadsRelationships(db, issue, taskUID, issueMap)
if err != nil {
    failedRelationships = append(failedRelationships, fmt.Sprintf("task %s: %v", issue.ID, err))
    continue
}
```

Then in cmd/import_beads.go, display warnings similar to the ingest command.

---

## PR #256 Assessment

**PR**: https://github.com/neongreen/mono/pull/256
**Title**: Fix test failures: resolve duplicate getCurrentUser and import cycle
**Status**: Open

### What PR #256 Does

1. **Extracts shared utility to break potential import cycle**:
   - Creates `internal/utils/user.go` with `GetCurrentUser()`
   - Updates 6 subcommand files to use the shared function
   - Removes duplicate implementations in `cmd/project/helpers.go` and `cmd/relate/helpers.go`

2. **Fixes status command registration issue**:
   - Removes duplicate command registration

### Current State (After PR #259 Work)

**Build Status**: ✅ Everything compiles without errors
- No import cycle exists currently
- Code has 3 duplicate `getCurrentUser()` implementations:
  - `cmd/utils.go` (line 28)
  - `cmd/project/helpers.go` (line 8)
  - `cmd/relate/helpers.go` (line 8)

### Is PR #256 Still Needed?

**Answer: Yes, it's still valuable but not urgent**

**Reasons to merge PR #256**:

1. **DRY Principle**: Eliminates code duplication (3 identical functions)
2. **Future-proofing**: Prevents potential import cycles if code structure changes
3. **Maintainability**: Single source of truth for user retrieval logic
4. **Best Practice**: Follows the pattern of extracting utilities to `internal/` packages

**Reasons it's not urgent**:

1. **No Current Breakage**: Code builds and tests pass
2. **No Active Import Cycle**: The current structure doesn't cause problems
3. **Small Impact**: The duplicate code is simple and unlikely to diverge

**Recommendation**: 
- ✅ Merge PR #256 - It's good refactoring that improves code quality
- Priority: Low (cleanup/refactoring)
- No conflict with PR #259 changes
- Can be merged independently

---

## Summary of Recommendations

1. **✅ Ingest errors**: Already fixed in this PR
2. **✅ Sync warnings**: Already handled properly  
3. **⚠️ Beads import**: Should add error tracking for:
   - Failed renumber notes (low priority)
   - Failed relationship imports (medium-high priority)
4. **✅ PR #256**: Should be merged as good code cleanup

## Next Steps

1. Complete review of this PR (#259/#260)
2. Consider creating follow-up issue for beads import error tracking
3. Merge PR #256 independently for code quality improvement
