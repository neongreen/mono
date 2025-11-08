# Dead Weight Code Analysis - ACTIONABLE ISSUES

**Focus**: Duplicate logic, useless code, and real maintenance burden
**Date**: 2025-11-07

---

## Executive Summary

Found **genuine duplicate logic** that creates maintenance burden:
- 3 identical `createTempPackage` functions (copy-pasted test helpers)
- 2 identical `setupTestLogger` + `installGhStub` functions
- 6+ duplicate handler implementations in want/ (stub + real implementation pattern)
- Duplicate test functions testing the same logic

These are **real maintenance problems** - when you fix a bug in one place, you have to remember to fix it in all copies.

---

## PRIORITY 1: Copy-Pasted Test Helper Functions

### Issue #1: Triple-Duplicated `createTempPackage` Function

**Location**: 3 identical copies in dissect test files

```
dissect/pkg/references/find_test.go:672    (25 lines)
dissect/pkg/symbols/extract_test.go:422    (25 lines)
dissect/pkg/typeinfo/load_test.go:258      (25 lines)
```

**What it does**: Creates a temp directory with a Go module and test files

**Problem**: EXACTLY identical implementations. Any bugfix or improvement needs to be made 3 times.

**Impact**:
- 75 lines of duplicated code
- If you fix a bug in one, you must fix it in all three
- Already diverging: only one has `createFileInDir` helper

**Recommendation**:
1. Create `dissect/internal/testhelpers/temp_package.go`
2. Move function there with signature: `func CreateTempPackage(t *testing.T, files map[string]string) string`
3. Update all 3 call sites to use centralized version

**Estimated Effort**: 30 minutes

---

### Issue #2: Double-Duplicated GitHub Test Helpers

**Location**: 2 identical copies

```
lib/ghclient/token_test.go:18    setupTestLogger (12 lines)
lib/ghrelease/ghrelease_test.go:14    setupTestLogger (12 lines)

lib/ghclient/token_test.go:30    installGhStub (30+ lines)
lib/ghrelease/ghrelease_test.go:26    installGhStub (30+ lines)
```

**What they do**:
- `setupTestLogger`: Sets up test logging that captures output
- `installGhStub`: Creates fake `gh` CLI executable for testing

**Problem**: EXACTLY identical implementations copy-pasted between two packages

**Impact**:
- 80+ lines of duplicated code
- Both packages test GitHub functionality - should share helpers
- Bug fixes must be applied twice

**Recommendation**:
1. Create `lib/testhelpers/github.go`
2. Move both functions there
3. Update imports in both test files

**Estimated Effort**: 20 minutes

---

## PRIORITY 2: Duplicate Test Functions

### Issue #3: Duplicate Test Cases

**Tests that are identical or nearly identical:**

```
TestRenameExport
  dissect/cmd/move_rename_test.go:128
  dissect/pkg/gopls/rename_test.go:228

TestRenameUnexport
  dissect/cmd/move_rename_test.go:165
  dissect/pkg/gopls/rename_test.go:270

TestCreateAuthenticatedRequest
  lib/ghrelease/ghrelease_test.go:166
  prrun/github_test.go:138

TestGetGitHubToken
  lib/ghrelease/ghrelease_test.go:86
  prrun/github_test.go:88

TestMoveFileDeeplyNested
  dissect/cmd/move_file_integration_imports_test.go:12
  dissect/pkg/refactor/move_file_test.go:351
```

**Problem**: Same tests in multiple places. When logic changes, tests drift apart.

**Recommendation**:
- **For dissect rename tests**: Keep in pkg/gopls/ (unit tests), remove from cmd/ (integration should test end-to-end, not individual functions)
- **For GitHub tests**: These suggest `prrun` and `lib/ghrelease` have duplicate logic. Investigate if `prrun` should just use `lib/ghrelease` functions.
- **For move file test**: Verify if they're testing different things. If identical, keep integration test only.

**Estimated Effort**: 1-2 hours to review and consolidate

---

## PRIORITY 3: Suspicious Duplicate Handler Pattern in want/

### Issue #4: Stub + Real Implementation Pattern

**Pattern found**: Multiple handlers have TWO implementations:
1. A stub in `want/cmd/handlers.go` that delegates to a function variable
2. A real implementation in `want/handlers.go` or `want/utils.go`

**Examples**:

```
handleGitHubAsset:
  want/cmd/handlers.go:56  (stub calling HandleGitHubAssetFunc)
  want/handlers.go:423     (real implementation)

installMonoRelease:
  want/cmd/mono.go:73      (stub calling InstallMonoReleaseFunc)
  want/mono.go:18          (real implementation)

getCompoundHandler:
  want/cmd/handlers.go:49  (stub calling GetCompoundHandlerFunc)
  want/utils.go:14         (real implementation)

handleJsonCommand:
  want/cmd/json.go:34      (stub)
  want/handlers.go:24      (real)

handleExcalifontCommand:
  want/cmd/excalifont.go:35 (stub)
  want/handlers.go:265      (real)
```

**Questions**:
1. Why the indirection? Is this for testing? Plugin architecture?
2. Are the stubs ever used without the real implementation?
3. Can this be simplified?

**Recommendation**:
1. Review the architecture - is this indirection necessary?
2. If it's for testability, consider dependency injection instead
3. If it's unused, remove stubs and use real implementations directly
4. Add comments explaining WHY this pattern exists

**Estimated Effort**: 2-3 hours to review architecture and decide

---

## PRIORITY 4: Duplicate Type Definitions

### Issue #5: PRInfo Defined in 3 Places

**Locations**:
```
prrun/types.go:20
want/cmd/types.go:142
want/mono.go:15
```

**Problem**: Same type defined 3 times. If you add a field, you need to update all 3.

**Impact**: High - this is a core data structure

**Recommendation**:
1. Check if all 3 definitions are identical
2. If yes: Create `lib/types/pr.go` with canonical definition
3. Update all references
4. Delete duplicates

**Estimated Effort**: 30-45 minutes

---

### Issue #6: FulfillmentPlan and PlanStep in 2 Places

**Locations**:
```
FulfillmentPlan:
  want/cmd/types.go:89
  want/handlers.go:19

PlanStep:
  want/cmd/types.go:80
  want/handlers.go:20
```

**Problem**: Same types in two files in same package

**Recommendation**:
- Keep in `want/cmd/types.go` only
- Remove from `want/handlers.go`
- Update imports

**Estimated Effort**: 10 minutes

---

## PRIORITY 5: Potential Dead Code (Needs Investigation)

### Issue #7: Functions With "Test", "Debug", "Experimental" in Names

**Found**: None in production code (only test names, which is fine)

### Issue #8: Deprecated/Legacy Markers

**Found**: 1 TODO marker

```
dissect/pkg/gopls/guess_extracted_file_name.go:10
  TODO(gopls): reuse the guesser from gotools/gopls/internal/golang/extracttofile.go
```

**Recommendation**: Address this TODO - either implement the reuse or document why it's not done

**Estimated Effort**: 1 hour

---

## NOT ISSUES (False Positives)

These looked suspicious but are actually fine:

### ❌ Token Functions (GetToken, GetGitHubToken)
**Status**: ✅ Not duplicate
- `lib/ghrelease/ghrelease.go:GetGitHubToken()` just calls `ghclient.GetToken()`
- This is delegation, not duplication

### ❌ Mock Implementations in Tests
**Status**: ✅ Acceptable pattern
- `lib/configschema/integration_test.go` has mock implementations
- These intentionally duplicate the interface for testing

### ❌ Test Functions with "Test" prefix
**Status**: ✅ These are test functions
- `TestParse`, `TestValue`, etc. are testing the actual functions
- Not duplicate logic

---

## Summary: Real Issues to Fix

### High Priority (Do First)
1. ✅ **Consolidate createTempPackage** (3 copies → 1) - 30 min
2. ✅ **Consolidate GitHub test helpers** (2 copies → 1) - 20 min
3. ✅ **Fix PRInfo duplication** (3 definitions → 1) - 30 min
4. ✅ **Fix FulfillmentPlan/PlanStep duplication** - 10 min

**Total High Priority: 1.5 hours**

### Medium Priority (Review and Decide)
5. ⚠️  **Review duplicate test functions** - 1-2 hours
6. ⚠️  **Review want/ handler architecture** - 2-3 hours

**Total Medium Priority: 3-5 hours**

### Low Priority
7. ℹ️ **Address TODO in gopls code** - 1 hour

---

## Metrics

**Duplicate Code Found**:
- 155+ lines of exactly duplicated test helpers
- 5 duplicate test functions
- 3 duplicate type definitions
- 5 duplicate handler implementations (questionable pattern)

**Estimated Maintenance Burden**:
- Every bug fix in test helpers must be applied 2-3 times
- Every schema change to PRInfo must be applied 3 times
- Test drift likely already happening

**Estimated Time to Fix High Priority Issues**: 1.5 hours

**Impact**:
- Remove 155+ lines of duplicated code
- Reduce bug-fix locations from 2-3x to 1x
- Prevent future test drift

---

## Action Plan

### Week 1: Kill the Copy-Paste
1. Create `dissect/internal/testhelpers/temp_package.go`
2. Move `createTempPackage` there (from 3 locations)
3. Create `lib/testhelpers/github.go`
4. Move `setupTestLogger` and `installGhStub` there (from 2 locations)
5. Run tests to verify nothing broke

### Week 2: Fix Type Duplication
6. Create `lib/types/pr.go` with canonical `PRInfo`
7. Update all references to use `lib/types.PRInfo`
8. Remove duplicate `FulfillmentPlan`/`PlanStep` from handlers.go

### Week 3: Review Architecture (Optional)
9. Review want/ handler pattern - is it necessary?
10. Review duplicate test functions - keep or consolidate?
11. Address TODO in gopls code

---

## Conclusion

You were right to be concerned about maintenance burden rather than fragmentation. The real issues are:

1. **Copy-pasted test helpers** - These will drift and cause bugs
2. **Duplicate type definitions** - Schema changes require multiple updates
3. **Duplicate test functions** - Unclear if they're testing different things
4. **Questionable architecture in want/** - Stub + implementation pattern may be unnecessary

The good news: The high-priority issues can be fixed in ~1.5 hours total.

The fragmentation (45 single-function files) is annoying but not causing bugs. The duplicate logic IS causing bugs and maintenance burden.
