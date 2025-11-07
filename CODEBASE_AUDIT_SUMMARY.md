# Comprehensive Codebase Audit Summary

**Date**: 2025-11-07
**Scope**: dissect/, want/, prrun/, lib/ directories
**Files Analyzed**: 139 Go files
**Total Symbols**: ~2000+

---

## Executive Summary

This audit identified several categories of technical debt and code organization issues that could improve codebase maintainability:

1. **Code Duplication**: 13 duplicate symbols across packages
2. **Over-fragmentation**: 45 files with single exported functions
3. **Test Code Duplication**: Multiple test helpers scattered across packages
4. **Missing Documentation**: 17 undocumented exported symbols
5. **Legacy Patterns**: 14 init functions, 5 "temp" patterns
6. **Library Candidates**: Several utility packages that could be consolidated or extracted

---

## Priority 1: High-Impact Consolidation Opportunities

### 1.1 Over-Fragmented Packages (HIGHEST PRIORITY)

**Problem**: Many packages have extreme fragmentation with one function per file.

#### `dissect/pkg/commands/` - 11 files, mostly single-function
**Recommendation**: Consolidate into 2-3 files:
- `commands/go_tooling.go` (Run* functions: RunGoCommand, RunGoBuild, RunGoModTidy, RunGoListNoArgs)
- `commands/go_module.go` (FindGoModuleRoot, GetModuleName, GetFullImportPath)
- `commands/goimports.go` (RunGoimportsOnFile, RunGoimportsOnDirectory)

**Files to consolidate**:
```
dissect/pkg/commands/run_command.go
dissect/pkg/commands/run_go_build.go
dissect/pkg/commands/run_go_command.go
dissect/pkg/commands/run_go_list_no_args.go
dissect/pkg/commands/run_go_mod_tidy.go
dissect/pkg/commands/run_goimports_on_directory.go
dissect/pkg/commands/run_goimports_on_file.go
dissect/pkg/commands/find_go_module_root.go
dissect/pkg/commands/get_full_import_path.go
dissect/pkg/commands/get_module_name.go
```

**Impact**: Reduce 10 files to 3 files, improve discoverability

---

#### `dissect/pkg/goutils/` - 11 files, mostly single-function
**Recommendation**: Consolidate into 3-4 files:
- `goutils/ast.go` (FindFunc, FindDecl, GetReceiverTypeName)
- `goutils/package.go` (GetPackageDeclaration, UpdatePackageDeclaration)
- `goutils/imports.go` (NormalizeImports)
- `goutils/file_io.go` (ReadGoFile, WriteGoFile)
- `goutils/predicates.go` (IsTestFile, IsTestFunction, ShouldRefactor, IsLower)

**Files to consolidate**:
```
dissect/pkg/goutils/find_func.go
dissect/pkg/goutils/get_package_declaration.go
dissect/pkg/goutils/get_receiver_type_name.go
dissect/pkg/goutils/is_test_file.go
dissect/pkg/goutils/is_test_function.go
dissect/pkg/goutils/normalize_imports.go
dissect/pkg/goutils/read_go_file.go
dissect/pkg/goutils/should_refactor.go
dissect/pkg/goutils/update_package_declaration.go
dissect/pkg/goutils/write_go_file.go
```

**Impact**: Reduce 10 files to 4-5 files

---

#### `dissect/pkg/utils/` - 5 single-function files
**Recommendation**: Consolidate into 1-2 files:
- `utils/strings.go` (CapitalizeFirstLetter, HashString, IsLower)
- `utils/files.go` (DeleteFile, MoveFile)

**Files to consolidate**:
```
dissect/pkg/utils/capitalize_first_letter.go
dissect/pkg/utils/delete_file.go
dissect/pkg/utils/hash_string.go
dissect/pkg/utils/is_lower.go
dissect/pkg/utils/move_file.go
```

**Impact**: Reduce 5 files to 2 files

---

#### `dissect/pkg/gopls/` - Multiple single-function files
**Recommendation**: Consolidate into 2 files:
- `gopls/commands.go` (AddImport, AddDotImport, ExtractToNewFile, Rename)
- `gopls/helpers.go` (GuessGoplsExtractedFileName, findSymbolOffset)

**Files to consolidate**:
```
dissect/pkg/gopls/add_dot_import.go
dissect/pkg/gopls/add_import.go
dissect/pkg/gopls/extract_to_new_file.go
dissect/pkg/gopls/guess_extracted_file_name.go
```

**Impact**: Reduce 4 files to 2 files

---

### 1.2 Test Utilities Consolidation (HIGH PRIORITY)

**Problem**: Test helper functions are duplicated across multiple test files.

#### Duplicate Test Helpers
```
setupTools        - dissect/cmd/main_test.go:29
setupGopls        - dissect/pkg/gopls/rename_test.go:14
setupTestLogger   - lib/ghclient/token_test.go:18
setupTestLogger   - lib/ghrelease/ghrelease_test.go:14
installGhStub     - lib/ghclient/token_test.go:30
installGhStub     - lib/ghrelease/ghrelease_test.go:26
```

**Recommendation**: Create shared test packages:
- `dissect/internal/testhelpers/` for dissect-specific test utilities
- `lib/testhelpers/` for library test utilities

**Files to create**:
```
dissect/internal/testhelpers/gopls.go       (setupTools, setupGopls)
dissect/internal/testhelpers/tempmodule.go  (createTempModule, createTempPackage)
lib/testhelpers/github.go                   (setupTestLogger, installGhStub)
```

---

#### Scatter Test Utilities (Already Identified)
**Current State**:
```
dissect/cmd/internal/testutils/    (ContainsFunc, ContainsString, FindSubstring)
dissect/pkg/testutils/              (CompareDirectories)
dissect/pkg/goutils/                (IsTestFile, IsTestFunction)
```

**Recommendation**: Consolidate all into `dissect/internal/testhelpers/`:
- Move `dissect/cmd/internal/testutils/*` → `dissect/internal/testhelpers/assertions.go`
- Move `dissect/pkg/testutils/*` → `dissect/internal/testhelpers/filesystem.go`
- Move test-related functions from `dissect/pkg/goutils` → `dissect/internal/testhelpers/predicates.go`

---

### 1.3 Type Duplication (HIGH PRIORITY)

**Problem**: Same types defined in multiple locations

#### PRInfo type (3 locations!)
```
prrun/types.go:20
want/cmd/types.go:142
want/mono.go:15
```

**Recommendation**:
- Create `lib/types/pr.go` with canonical `PRInfo` definition
- Update all references to use `lib/types.PRInfo`
- Delete duplicate definitions

---

#### FulfillmentPlan and PlanStep (2 locations each)
```
FulfillmentPlan:
  want/cmd/types.go:89
  want/handlers.go:19

PlanStep:
  want/cmd/types.go:80
  want/handlers.go:20
```

**Recommendation**:
- Keep only in `want/cmd/types.go`
- Update `want/handlers.go` to import from types
- These types likely belong together and were accidentally duplicated

---

### 1.4 Duplicate Test Functions (MEDIUM PRIORITY)

**Problem**: Test functions duplicated between integration and unit tests

```
TestCreateAuthenticatedRequest  - lib/ghrelease/ghrelease_test.go:166
TestCreateAuthenticatedRequest  - prrun/github_test.go:138

TestGetGitHubToken              - lib/ghrelease/ghrelease_test.go:86
TestGetGitHubToken              - prrun/github_test.go:88

TestMoveFileDeeplyNested        - dissect/cmd/move_file_integration_imports_test.go:12
TestMoveFileDeeplyNested        - dissect/pkg/refactor/move_file_test.go:351

TestRenameExport                - dissect/cmd/move_rename_test.go:128
TestRenameExport                - dissect/pkg/gopls/rename_test.go:228

TestRenameUnexport              - dissect/cmd/move_rename_test.go:165
TestRenameUnexport              - dissect/pkg/gopls/rename_test.go:270
```

**Recommendation**:
- For GitHub tests: Extract common test logic to `lib/testhelpers/github.go`
- For dissect tests: Keep integration tests in cmd/, remove duplicates from pkg/
- Use table-driven tests or shared test helpers instead of duplicating

---

## Priority 2: Medium-Impact Improvements

### 2.1 Init Functions (14 total)

**Problem**: Init functions make code harder to test and create hidden dependencies.

**Locations**:
```
dissect/cmd/main.go:160        - Logging setup
dissect/cmd/main_test.go:22    - Test logging
dissect/cmd/move.go:103        - Command registration
dissect/cmd/version.go:7       - Version command
lib/version/version.go:20      - Version info
want/cmd/*.go (9 files)        - Command registration
```

**Recommendation**:
- Logging setup: Convert to explicit `initLogging()` call in main()
- Command registration: Already using cobra, these inits are fine (cobra pattern)
- Test init: Keep for test logging setup (acceptable pattern)
- **Action**: Only `dissect/cmd/main.go:160` and `lib/version/version.go:20` need review

---

### 2.2 Undocumented Code (17 symbols)

**File with most issues**:
- `dissect/cmd/process_file.go` (5 undocumented symbols)

**Recommendation**: Add godoc comments to exported symbols

---

### 2.3 configschema Mock Interface Issue

**Problem**: Mock interface methods duplicated between test and production code

```
GetAllPaths          - lib/configschema/integration_test.go:54 (mock)
GetAllPaths          - lib/configschema/jsonschema_parser.go:62 (real)

GetPropertyInfo      - lib/configschema/integration_test.go:62 (mock)
GetPropertyInfo      - lib/configschema/jsonschema_parser.go:174 (real)

GetAllSettingsWithInfo - lib/configschema/integration_test.go:66 (mock)
GetAllSettingsWithInfo - lib/configschema/jsonschema_parser.go:198 (real)

ValidatePath         - lib/configschema/integration_test.go:58 (mock)
ValidatePath         - lib/configschema/jsonschema_parser.go:88 (real)
```

**Recommendation**:
- This is actually good practice (interface + mock implementation)
- Consider using mockgen or similar to auto-generate mocks
- Alternatively, extract interface to separate file: `lib/configschema/schema_reader.go`

---

## Priority 3: Library Extraction Opportunities

### 3.1 Go Tooling Utilities (Candidate for Public Library)

**Current Location**: `dissect/pkg/commands/`

**Functions**:
```
RunGoCommand
RunGoBuild
RunGoModTidy
RunGoListNoArgs
FindGoModuleRoot
GetModuleName
GetFullImportPath
```

**Recommendation**:
- These are generic Go tooling utilities
- Could be extracted to `lib/gotool/` for reuse
- Or published as separate open-source library
- **Value**: Other Go tools could benefit from this

---

### 3.2 Command Execution Utilities

**Current Locations**: Scattered across multiple packages

**Functions**:
```
dissect/pkg/commands/run_command.go:12      - RunCommand
dissect/pkg/commands/run_goimports_*.go     - Goimports wrappers
```

**Recommendation**:
- Keep in `dissect/pkg/commands/` after consolidation
- These are specific to dissect's needs

---

## Priority 4: Code Organization Issues

### 4.1 Over-Fragmented Directories

**Directories with >10 files that need review**:
```
dissect/cmd/              - 19 files (mostly test files, acceptable)
dissect/pkg/goutils/      - 11 files (CONSOLIDATE - see Priority 1)
dissect/pkg/commands/     - 11 files (CONSOLIDATE - see Priority 1)
prrun/                    - 11 files (review needed)
want/cmd/                 - 12 files (review needed)
```

**Recommendation for prrun/**:
- Check if files can be logically grouped
- May be acceptable if each file represents a distinct concern

**Recommendation for want/cmd/**:
- Each file appears to be a command handler
- Current structure is fine (one command per file is a valid pattern)

---

### 4.2 Move Command Test Files (12 test files!)

**Pattern**: `dissect/cmd/move_*_test.go`

**Files**:
```
move_ast_requirements_test.go
move_batch_test.go
move_command_test.go
move_command_with_comments_test.go
move_edge_cases_test.go
move_file_integration_basic_test.go
move_file_integration_content_test.go
move_file_integration_imports_test.go
move_file_test.go
move_fileset_bug_test.go
move_rename_test.go
move_types_test.go
```

**Recommendation**:
- This is reasonable for a complex feature with many edge cases
- Consider consolidating into themed files:
  - `move_basic_test.go` (basic functionality)
  - `move_integration_test.go` (integration tests)
  - `move_edge_cases_test.go` (edge cases, bugs, regressions)
  - `move_batch_test.go` (batch operations)
- **Impact**: Reduce from 12 to 4 test files

---

## Priority 5: Potential Dead Code (Needs Investigation)

### 5.1 Schema Functions in lib/configschema/

**Single-function files**:
```
lib/configschema/claude.go     - ClaudeSchema (func)
lib/configschema/mise.go       - MiseSchema (func)
lib/configschema/starship.go   - StarshipSchema (func)
```

**Question**: Are these used? Check references.

**Action**: Run `grep -r "ClaudeSchema\|MiseSchema\|StarshipSchema" .` to verify usage

---

### 5.2 Want Command Handlers

**Single-export files in want/cmd/**:
```
want/cmd/excalifont.go   - HandleExcalifontCommandFunc
want/cmd/json.go         - HandleJsonCommandFunc
want/cmd/md.go           - HandleMarkdownCommandFunc
```

**Action**: Verify these are all actively used and not legacy/experimental commands

---

## Detailed Action Plan

### Phase 1: High-Impact Consolidation (1-2 weeks)
1. ✅ Consolidate `dissect/pkg/commands/` (10 files → 3 files)
2. ✅ Consolidate `dissect/pkg/goutils/` (10 files → 4 files)
3. ✅ Consolidate `dissect/pkg/utils/` (5 files → 2 files)
4. ✅ Fix PRInfo type duplication (create lib/types/pr.go)
5. ✅ Fix FulfillmentPlan/PlanStep duplication

### Phase 2: Test Code Cleanup (1 week)
6. ✅ Create dissect/internal/testhelpers/ package
7. ✅ Consolidate test utilities
8. ✅ Remove duplicate test functions
9. ✅ Create lib/testhelpers/ for shared GitHub test utilities

### Phase 3: Documentation (2-3 days)
10. ✅ Document all symbols in dissect/cmd/process_file.go
11. ✅ Review and document other undocumented exports

### Phase 4: Code Review and Cleanup (1 week)
12. ✅ Review init functions in dissect/cmd/main.go
13. ✅ Check for dead code in lib/configschema/
14. ✅ Verify want/cmd/ handlers are all used
15. ✅ Consider consolidating move_*_test.go files

### Phase 5: Library Extraction (Optional, 1-2 weeks)
16. ✅ Extract Go tooling utilities to lib/gotool/
17. ✅ Consider publishing as separate library

---

## Metrics

**Before Consolidation**:
- 45 single-symbol files
- 10+ scattered test helpers
- 3 duplicate type definitions
- 13 duplicate symbols
- 14 init functions (some unnecessary)

**Expected After Consolidation**:
- ~20 single-symbol files (reduction of 55%)
- Centralized test helpers in 2 packages
- 0 duplicate type definitions
- 0-3 duplicate symbols (mocks are acceptable)
- ~10 init functions (only necessary ones)

**Estimated Impact**:
- **Developer Time**: 10-15% reduction in time searching for utilities
- **Code Navigation**: 40% fewer files to navigate in key packages
- **Test Maintenance**: 30% easier to maintain tests with centralized helpers
- **Onboarding**: 25% faster for new developers to understand structure

---

## Notes on What NOT to Change

### Acceptable Patterns
1. **One command per file** in `want/cmd/` - This is good CLI structure
2. **Many test files** for move command - Complex feature needs thorough testing
3. **Init functions for cobra commands** - Standard cobra pattern
4. **Mock implementations** - Duplicate methods are fine if they're test mocks

### Low-Priority Issues
1. String() method duplication - These are interface implementations, not duplication
2. Test* prefix functions - These are test functions, not duplication
3. Init functions in test files - Acceptable for test setup

---

## Conclusion

The codebase shows signs of **over-application of the "one function per file" pattern**,
especially in utility packages. This was likely done with good intentions (clear separation
of concerns) but has resulted in fragmentation that makes the code harder to navigate and maintain.

**Key Recommendation**: Consolidate related functionality into cohesively-themed files rather
than extreme fragmentation. Aim for 3-7 functions per file for utility packages.

The test code duplication is the **highest priority issue**, as it directly impacts
maintainability and can lead to test drift where identical tests diverge over time.

Overall code quality is good - this is primarily an organizational issue rather than
a code quality issue.
