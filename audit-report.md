# Codebase Audit Report

## 1. Duplicate Symbol Names (Potential Code Duplication)

- **FulfillmentPlan (type)** appears in 2 locations:
  - want/cmd/types.go:89
  - want/handlers.go:19

- **GetAllPaths (method)** appears in 2 locations:
  - lib/configschema/integration_test.go:54
  - lib/configschema/jsonschema_parser.go:62

- **GetAllSettingsWithInfo (method)** appears in 2 locations:
  - lib/configschema/integration_test.go:66
  - lib/configschema/jsonschema_parser.go:198

- **GetPropertyInfo (method)** appears in 2 locations:
  - lib/configschema/integration_test.go:62
  - lib/configschema/jsonschema_parser.go:174

- **PRInfo (type)** appears in 3 locations:
  - prrun/types.go:20
  - want/cmd/types.go:142
  - want/mono.go:15

- **PlanStep (type)** appears in 2 locations:
  - want/cmd/types.go:80
  - want/handlers.go:20

- **String (method)** appears in 2 locations:
  - dissect/pkg/parser/filespec.go:51
  - lib/toml/toml.go:333

- **TestCreateAuthenticatedRequest (func)** appears in 2 locations:
  - lib/ghrelease/ghrelease_test.go:166
  - prrun/github_test.go:138

- **TestGetGitHubToken (func)** appears in 2 locations:
  - lib/ghrelease/ghrelease_test.go:86
  - prrun/github_test.go:88

- **TestMoveFileDeeplyNested (func)** appears in 2 locations:
  - dissect/cmd/move_file_integration_imports_test.go:12
  - dissect/pkg/refactor/move_file_test.go:351

- **TestRenameExport (func)** appears in 2 locations:
  - dissect/cmd/move_rename_test.go:128
  - dissect/pkg/gopls/rename_test.go:228

- **TestRenameUnexport (func)** appears in 2 locations:
  - dissect/cmd/move_rename_test.go:165
  - dissect/pkg/gopls/rename_test.go:270

- **ValidatePath (method)** appears in 2 locations:
  - lib/configschema/integration_test.go:58
  - lib/configschema/jsonschema_parser.go:88

## 2. Undocumented Code

Total undocumented exported symbols: 17

Files with >3 undocumented symbols:
- dissect/cmd/process_file.go (5 undocumented)

## 3. Test Utilities (Potential for Consolidation)

Test utility functions found in multiple locations:

**dissect/cmd/internal/testutils:**
- ContainsFunc (dissect/cmd/internal/testutils/contains_func.go)
- ContainsString (dissect/cmd/internal/testutils/contains_string.go)
- FindSubstring (dissect/cmd/internal/testutils/find_substring.go)

**dissect/pkg/goutils:**
- IsTestFile (dissect/pkg/goutils/is_test_file.go)
- IsTestFunction (dissect/pkg/goutils/is_test_function.go)
- TestNormalizeImports (dissect/pkg/goutils/normalize_imports_test.go)

**dissect/pkg/testutils:**
- CompareDirectories (dissect/pkg/testutils/compare_directories.go)

## 4. Small Files (Potential for Consolidation)

Files with exactly 1 exported symbol (45 total):

- dissect/cmd/internal/testutils/contains_func.go: ContainsFunc (func)
- dissect/cmd/internal/testutils/contains_string.go: ContainsString (func)
- dissect/cmd/internal/testutils/find_substring.go: FindSubstring (func)
- dissect/pkg/commands/find_go_module_root.go: FindGoModuleRoot (func)
- dissect/pkg/commands/get_full_import_path.go: GetFullImportPath (func)
- dissect/pkg/commands/get_module_name.go: GetModuleName (func)
- dissect/pkg/commands/run_command.go: RunCommand (func)
- dissect/pkg/commands/run_go_build.go: RunGoBuild (func)
- dissect/pkg/commands/run_go_command.go: RunGoCommand (func)
- dissect/pkg/commands/run_go_list_no_args.go: RunGoListNoArgs (func)
- dissect/pkg/commands/run_go_mod_tidy.go: RunGoModTidy (func)
- dissect/pkg/commands/run_goimports_on_directory.go: RunGoimportsOnDirectory (func)
- dissect/pkg/commands/run_goimports_on_file.go: RunGoimportsOnFile (func)
- dissect/pkg/gopls/add_dot_import.go: AddDotImport (func)
- dissect/pkg/gopls/add_import.go: AddImport (func)
- dissect/pkg/gopls/extract_to_new_file.go: ExtractToNewFile (func)
- dissect/pkg/gopls/guess_extracted_file_name.go: GuessGoplsExtractedFileName (func)
- dissect/pkg/gopls/rename.go: Rename (func)
- dissect/pkg/goutils/get_package_declaration.go: GetPackageDeclaration (func)
- dissect/pkg/goutils/get_receiver_type_name.go: GetReceiverTypeName (func)
- dissect/pkg/goutils/is_test_file.go: IsTestFile (func)
- dissect/pkg/goutils/is_test_function.go: IsTestFunction (func)
- dissect/pkg/goutils/normalize_imports.go: NormalizeImports (func)
- dissect/pkg/goutils/read_go_file.go: ReadGoFile (func)
- dissect/pkg/goutils/should_refactor.go: ShouldRefactor (func)
- dissect/pkg/goutils/update_package_declaration.go: UpdatePackageDeclaration (func)
- dissect/pkg/goutils/write_go_file.go: WriteGoFile (func)
- dissect/pkg/refactor/dependencies.go: UnexportedDependency (type)
- dissect/pkg/refactor/determine_extraction_target.go: DetermineExtractionTarget (func)
- dissect/pkg/testutils/compare_directories.go: CompareDirectories (func)
- dissect/pkg/utils/capitalize_first_letter.go: CapitalizeFirstLetter (func)
- dissect/pkg/utils/delete_file.go: DeleteFile (func)
- dissect/pkg/utils/hash_string.go: HashString (func)
- dissect/pkg/utils/is_lower.go: IsLower (func)
- dissect/pkg/utils/move_file.go: MoveFile (func)
- lib/configschema/claude.go: ClaudeSchema (func)
- lib/configschema/mise.go: MiseSchema (func)
- lib/configschema/starship.go: StarshipSchema (func)
- want/cmd/excalifont.go: HandleExcalifontCommandFunc (var)
- want/cmd/json.go: HandleJsonCommandFunc (var)
- want/cmd/md.go: HandleMarkdownCommandFunc (var)
- want/cmd/root.go: Execute (func)
- want/mono.go: PRInfo (type)
- want/tools.go: ToolRegistry (var)
- want/utils.go: CompoundHandler (type)

## 5. Similar Function Names (Potential Duplication)

**Find*** prefix (9 functions):
- FindAllSymbols (dissect/pkg/symbols/extract.go:22)
- FindDecl (dissect/pkg/goutils/find_func.go:28)
- FindExportedSymbols (dissect/pkg/symbols/extract.go:28)
- FindFunc (dissect/pkg/goutils/find_func.go:10)
- FindGoFiles (dissect/pkg/commands/find_go_files.go:18)
- FindGoModuleRoot (dissect/pkg/commands/find_go_module_root.go:10)
- FindPlatformAsset (lib/ghrelease/ghrelease.go:121)
- FindReferences (dissect/pkg/references/find.go:24)
- FindSubstring (dissect/cmd/internal/testutils/find_substring.go:4)

**Get*** prefix (18 functions):
- GetAllPaths (lib/configschema/integration_test.go:54)
- GetAllPaths (lib/configschema/jsonschema_parser.go:62)
- GetAllSettingsWithInfo (lib/configschema/integration_test.go:66)
- GetAllSettingsWithInfo (lib/configschema/jsonschema_parser.go:198)
- GetCompletionOptions (lib/configschema/jsonschema_parser.go:25)
- GetCurrentPlatform (lib/ghrelease/ghrelease.go:40)
- GetFullImportPath (dissect/pkg/commands/get_full_import_path.go:8)
- GetGitHubToken (lib/ghrelease/ghrelease.go:64)
- GetModuleName (dissect/pkg/commands/get_module_name.go:10)
- GetPackageDeclaration (dissect/pkg/goutils/get_package_declaration.go:8)
- ... and 8 more

**Test*** prefix (236 functions):
- TestAddingNewKeysPreservesStyle (lib/toml/preservation_test.go:331)
- TestAllDissectIntegration (dissect/cmd/main_test.go:228)
- TestAnalyzeMoveDependenciesExportedIgnored (dissect/pkg/refactor/dependencies_test.go:157)
- TestAnalyzeMoveDependenciesMultipleTypes (dissect/pkg/refactor/dependencies_test.go:80)
- TestAnalyzeMoveDependenciesSelfReferencesIgnored (dissect/pkg/refactor/dependencies_test.go:213)
- TestAnalyzeMoveDependenciesUnit (dissect/pkg/refactor/dependencies_test.go:12)
- TestApplyMap (lib/toml/apply_test.go:154)
- TestArrayOfInlineTables (lib/toml/edge_cases_test.go:89)
- TestArrayOfTables (lib/toml/advanced_test.go:65)
- TestArrayOfTablesPreservation (lib/toml/edge_cases_test.go:185)
- ... and 226 more

**Run*** prefix (8 functions):
- RunCommand (dissect/pkg/commands/run_command.go:12)
- RunExternalProjectTest (dissect/pkg/externaltest/externaltest.go:47)
- RunGoBuild (dissect/pkg/commands/run_go_build.go:8)
- RunGoCommand (dissect/pkg/commands/run_go_command.go:8)
- RunGoListNoArgs (dissect/pkg/commands/run_go_list_no_args.go:9)
- RunGoModTidy (dissect/pkg/commands/run_go_mod_tidy.go:8)
- RunGoimportsOnDirectory (dissect/pkg/commands/run_goimports_on_directory.go:11)
- RunGoimportsOnFile (dissect/pkg/commands/run_goimports_on_file.go:9)

## 6. Potential Library Candidates

**Command Execution** (16 functions):
- GitHubWorkflowRun (prrun/types.go:28)
- GitHubWorkflowRunsResponse (prrun/types.go:36)
- HandleExcalifontCommandFunc (want/cmd/excalifont.go:33)
- HandleJsonCommandFunc (want/cmd/json.go:32)
- HandleMarkdownCommandFunc (want/cmd/md.go:31)
- NewVersionCommand (lib/version/version.go:77)
- RunCommand (dissect/pkg/commands/run_command.go:12)
- RunExternalProjectTest (dissect/pkg/externaltest/externaltest.go:47)
- RunGoBuild (dissect/pkg/commands/run_go_build.go:8)
- RunGoCommand (dissect/pkg/commands/run_go_command.go:8)
- RunGoListNoArgs (dissect/pkg/commands/run_go_list_no_args.go:9)
- RunGoModTidy (dissect/pkg/commands/run_go_mod_tidy.go:8)
- RunGoimportsOnDirectory (dissect/pkg/commands/run_goimports_on_directory.go:11)
- RunGoimportsOnFile (dissect/pkg/commands/run_goimports_on_file.go:9)
- TestMoveCommand (dissect/cmd/move_command_test.go:12)
- ... and 1 more

**File Operations** (72 functions):
- DeleteFile (dissect/pkg/utils/delete_file.go:9)
- ExtractToNewFile (dissect/pkg/gopls/extract_to_new_file.go:16)
- FileSpec (dissect/pkg/parser/filespec.go:9)
- FileSymbols (dissect/cmd/list.go:27)
- FindGoFiles (dissect/pkg/commands/find_go_files.go:18)
- GetAllPaths (lib/configschema/integration_test.go:54)
- GetAllPaths (lib/configschema/jsonschema_parser.go:62)
- GetFullImportPath (dissect/pkg/commands/get_full_import_path.go:8)
- GuessGoplsExtractedFileName (dissect/pkg/gopls/guess_extracted_file_name.go:10)
- IsTestFile (dissect/pkg/goutils/is_test_file.go:8)
- MoveBatchFiles (dissect/pkg/refactor/move_file.go:488)
- MoveFile (dissect/pkg/utils/move_file.go:10)
- MoveFileWithImportUpdates (dissect/pkg/refactor/move_file.go:73)
- ParseFileSpec (dissect/pkg/parser/filespec.go:22)
- Path (lib/cli/cli.go:64)
- ... and 57 more

**String Utilities** (15 functions):
- ContainsString (dissect/cmd/internal/testutils/contains_string.go:4)
- FormatValueToString (lib/toml/toml.go:435)
- HashString (dissect/pkg/utils/hash_string.go:8)
- ParseString (lib/toml/toml.go:61)
- String (dissect/pkg/parser/filespec.go:51)
- String (lib/toml/toml.go:333)
- TestFileSpec_String (dissect/pkg/parser/filespec_test.go:138)
- TestFormatFunctions (lib/cli/cli_test.go:56)
- TestFormatMonoTag (want/main_test.go:73)
- TestGetPlatformBinaryName_DoubleDashFormat (prrun/github_test.go:57)
- TestIsHexString (want/main_test.go:190)
- TestMoveWithDifferentFormatting (dissect/cmd/move_ast_requirements_test.go:20)
- TestMultilineStrings (lib/toml/advanced_test.go:91)
- TestNumberFormats (lib/toml/advanced_test.go:140)
- TestQualifyReferences_PreservesFormatting (dissect/pkg/qualify/rewrite_test.go:476)

**Go Tooling** (5 functions):
- FindGoModuleRoot (dissect/pkg/commands/find_go_module_root.go:10)
- GoListPackage (dissect/pkg/commands/find_go_files.go:12)
- RunGoBuild (dissect/pkg/commands/run_go_build.go:8)
- RunGoListNoArgs (dissect/pkg/commands/run_go_list_no_args.go:9)
- RunGoModTidy (dissect/pkg/commands/run_go_mod_tidy.go:8)

# Deep Analysis - Additional Issues

## Directory Structure Analysis

### Packages with few symbols (potential for consolidation):
- **testutils**: 4 symbols across 4 files
  - dissect/cmd/internal/testutils/contains_func.go
  - dissect/cmd/internal/testutils/contains_string.go
  - dissect/cmd/internal/testutils/find_substring.go
  - dissect/pkg/testutils/compare_directories.go

## Test Helper Functions (Scattered)

Test helper functions that could be centralized:
- setupTools in dissect/cmd/main_test.go:29
- createTempModuleForBatch in dissect/cmd/move_batch_test.go:121
- setupGopls in dissect/pkg/gopls/rename_test.go:14
- createTempModule in dissect/pkg/qualify/rewrite_test.go:1170
- createTempPackage in dissect/pkg/references/find_test.go:672
- createTempPackage in dissect/pkg/symbols/extract_test.go:422
- createTempPackage in dissect/pkg/typeinfo/load_test.go:258
- setupTestLogger in lib/ghclient/token_test.go:18
- installGhStub in lib/ghclient/token_test.go:30
- setupTestLogger in lib/ghrelease/ghrelease_test.go:14
- installGhStub in lib/ghrelease/ghrelease_test.go:26

## Mock/Stub Implementations

Mock/stub implementations found:
- mockLogger (type) in dissect/cmd/smoke-test/main.go:137
- installGhStub (func) in lib/ghclient/token_test.go:30
- installGhStub (func) in lib/ghrelease/ghrelease_test.go:26

**Recommendation**: Consider consolidating these into a shared testing package or using a mocking framework.

## Init Functions (Legacy Setup Code)

Found 14 init functions:
- dissect/cmd/main.go:160 (package: main)
- dissect/cmd/main_test.go:22 (package: main_test)
- dissect/cmd/move.go:103 (package: main)
- dissect/cmd/version.go:7 (package: main)
- lib/version/version.go:20 (package: version)
- want/cmd/check.go:28 (package: cmd)
- want/cmd/excalifont.go:28 (package: cmd)
- want/cmd/forget.go:25 (package: cmd)
- want/cmd/json.go:27 (package: cmd)
- want/cmd/list.go:29 (package: cmd)
- want/cmd/md.go:26 (package: cmd)
- want/cmd/mono.go:51 (package: cmd)
- want/cmd/root.go:43 (package: cmd)
- want/cmd/version.go:7 (package: cmd)

**Note**: Init functions can make code harder to test and reason about. Consider explicit initialization where possible.

## Types That Appear in Multiple Packages

Common type patterns across packages (may indicate need for shared types):

**PRInfo** (3 occurrences):
- main (prrun/types.go:20)
- cmd (want/cmd/types.go:142)
- main (want/mono.go:15)

## Potential Legacy/Deprecated Code Patterns

### Temporary code
Pattern: 'Temp' (5 matches)

- createTempModuleForBatch in dissect/cmd/move_batch_test.go:121
- createTempModule in dissect/pkg/qualify/rewrite_test.go:1170
- createTempPackage in dissect/pkg/references/find_test.go:672
- createTempPackage in dissect/pkg/symbols/extract_test.go:422
- createTempPackage in dissect/pkg/typeinfo/load_test.go:258

## Over-Fragmented Code

### Directories with many single-purpose files (>10 files):
- **dissect/pkg/goutils**: 11 files (consider consolidating related functionality)
- **dissect/cmd**: 19 files (consider consolidating related functionality)
- **dissect/pkg/commands**: 11 files (consider consolidating related functionality)
- **prrun**: 11 files (consider consolidating related functionality)
- **want/cmd**: 12 files (consider consolidating related functionality)

### Files with common prefixes (candidates for consolidation):

**dissect/cmd/move*** pattern (12 files):
- dissect/cmd/move_ast_requirements_test.go
- dissect/cmd/move_batch_test.go
- dissect/cmd/move_command_test.go
- dissect/cmd/move_command_with_comments_test.go
- dissect/cmd/move_edge_cases_test.go
- dissect/cmd/move_file_integration_basic_test.go
- dissect/cmd/move_file_integration_content_test.go
- dissect/cmd/move_file_integration_imports_test.go
- ... and 4 more

**dissect/pkg/commands/run*** pattern (7 files):
- dissect/pkg/commands/run_command.go
- dissect/pkg/commands/run_go_build.go
- dissect/pkg/commands/run_go_command.go
- dissect/pkg/commands/run_go_list_no_args.go
- dissect/pkg/commands/run_go_mod_tidy.go
- dissect/pkg/commands/run_goimports_on_directory.go
- dissect/pkg/commands/run_goimports_on_file.go

