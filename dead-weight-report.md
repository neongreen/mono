# Dead Weight Code Analysis

Focus: Useless code, duplicate logic, and maintenance burden

## 1. Near-Duplicate Functions (Likely Duplicate Logic)

These functions have very similar names and likely contain duplicate or near-duplicate logic:

### 'init' variations (14 functions)
- init in dissect/cmd/main.go:160
- init in dissect/cmd/main_test.go:22
- init in dissect/cmd/move.go:103
- init in dissect/cmd/version.go:7
- init in lib/version/version.go:20
- init in want/cmd/check.go:28
- init in want/cmd/excalifont.go:28
- init in want/cmd/forget.go:25
- init in want/cmd/json.go:27
- init in want/cmd/list.go:29
- init in want/cmd/md.go:26
- init in want/cmd/mono.go:51
- init in want/cmd/root.go:43
- init in want/cmd/version.go:7

### 'main' variations (6 functions)
- main in dissect/cmd/main.go:167
- main in dissect/cmd/smoke-test/main.go:15
- main in lib/toml/examples/basic/main.go:11
- main in lib/toml/examples/config_manager/main.go:13
- main in prrun/main.go:12
- main in want/main.go:10

### 'TempPackage' variations (3 functions)
- createTempPackage in dissect/pkg/references/find_test.go:672
- createTempPackage in dissect/pkg/symbols/extract_test.go:422
- createTempPackage in dissect/pkg/typeinfo/load_test.go:258

### 'Value' variations (2 functions)
- Value in lib/cli/cli.go:69
- TestValue in lib/cli/cli_test.go:36

### 'CompoundHandler' variations (2 functions)
- getCompoundHandler in want/cmd/handlers.go:49
- getCompoundHandler in want/utils.go:14

### 'ParseFileSpec' variations (2 functions)
- ParseFileSpec in dissect/pkg/parser/filespec.go:22
- TestParseFileSpec in dissect/pkg/parser/filespec_test.go:7

### 'handleExcalifontCommand' variations (2 functions)
- handleExcalifontCommand in want/cmd/excalifont.go:35
- handleExcalifontCommand in want/handlers.go:265

### 'IsDirectory' variations (2 functions)
- IsDirectory in dissect/pkg/parser/batch_move.go:106
- TestIsDirectory in dissect/pkg/parser/batch_move_test.go:227

### 'RenameUnexport' variations (2 functions)
- TestRenameUnexport in dissect/cmd/move_rename_test.go:165
- TestRenameUnexport in dissect/pkg/gopls/rename_test.go:270

### 'ExpandGlobs' variations (2 functions)
- ExpandGlobs in dissect/pkg/parser/batch_move.go:57
- TestExpandGlobs in dissect/pkg/parser/batch_move_test.go:139

### 'AllPaths' variations (2 functions)
- GetAllPaths in lib/configschema/integration_test.go:54
- GetAllPaths in lib/configschema/jsonschema_parser.go:62

### 'NormalizeImports' variations (2 functions)
- NormalizeImports in dissect/pkg/goutils/normalize_imports.go:16
- TestNormalizeImports in dissect/pkg/goutils/normalize_imports_test.go:9

### 'Bytes' variations (2 functions)
- Bytes in lib/toml/toml.go:338
- TestBytes in lib/toml/toml_test.go:720

### 'handleGitHubAsset' variations (2 functions)
- handleGitHubAsset in want/cmd/handlers.go:56
- handleGitHubAsset in want/handlers.go:423

### 'ListReleases' variations (2 functions)
- TestListReleases in lib/ghrelease/api_test.go:273
- ListReleases in lib/ghrelease/ghrelease.go:246

### 'handleJsonCommand' variations (2 functions)
- handleJsonCommand in want/cmd/json.go:34
- handleJsonCommand in want/handlers.go:24

### 'installMonoRelease' variations (2 functions)
- installMonoRelease in want/cmd/mono.go:73
- installMonoRelease in want/mono.go:18

### 'ReplaceMap' variations (2 functions)
- ReplaceMap in lib/toml/apply.go:82
- TestReplaceMap in lib/toml/apply_test.go:293

### 'MoveFile' variations (2 functions)
- TestMoveFile in dissect/cmd/move_file_test.go:14
- MoveFile in dissect/pkg/utils/move_file.go:10

### 'Parse' variations (2 functions)
- Parse in lib/toml/toml.go:52
- TestParse in lib/toml/toml_test.go:9

## 2. Duplicate Test Logic

### Setup Functions (4 total)
**LIKELY DUPLICATE LOGIC**: These should probably be consolidated into a shared test package

- setupTools in dissect/cmd/main_test.go:29
- setupGopls in dissect/pkg/gopls/rename_test.go:14
- setupTestLogger in lib/ghclient/token_test.go:18
- setupTestLogger in lib/ghrelease/ghrelease_test.go:14

### Create/Factory Functions in Tests (9 total)
**LIKELY DUPLICATE LOGIC**: These create test fixtures and are probably duplicated

- createTempModuleForBatch in dissect/cmd/move_batch_test.go:121
- createFileForBatch in dissect/cmd/move_batch_test.go:137
- createTempModule in dissect/pkg/qualify/rewrite_test.go:1170
- createFile in dissect/pkg/qualify/rewrite_test.go:1188
- createTempPackage in dissect/pkg/references/find_test.go:672
- createFileInDir in dissect/pkg/references/find_test.go:698
- createTempPackage in dissect/pkg/symbols/extract_test.go:422
- createTempPackage in dissect/pkg/typeinfo/load_test.go:258
- createFileInDir in dissect/pkg/typeinfo/load_test.go:284

### Mock/Stub Functions (2 total)
**MAINTENANCE BURDEN**: Consider using a mocking framework instead

- installGhStub in lib/ghclient/token_test.go:30
- installGhStub in lib/ghrelease/ghrelease_test.go:26

## 3. Deprecated/Legacy Code Patterns

### TODOs/FIXMEs in Documentation (1 items)
**ACTION**: Address these technical debt items

- GuessGoplsExtractedFileName (func) in dissect/pkg/gopls/guess_extracted_file_name.go:10 - GuessGoplsExtractedFileName guesses the file name that gopls would use for the extracted function.

## 4. Potential Dead Code (Suspiciously Specific Names)

These symbols have very specific names suggesting they might only be used once or for a specific edge case:

Found 5 suspicious symbols:

- TestResult (type) in dissect/pkg/externaltest/externaltest.go:28 - TestResult contains the results of running dissect on an external project
- RunExternalProjectTest (func) in dissect/pkg/externaltest/externaltest.go:47 - RunExternalProjectTest clones a project, runs dissect, and validates the results
- IsTestFile (func) in dissect/pkg/goutils/is_test_file.go:8 - isTestFile determines if a given file path corresponds to a Go test file.
- IsTestFunction (func) in dissect/pkg/goutils/is_test_function.go:12 - Check if the function name starts with "Test", is exported, and is not a method.
- JJSchemaLatest (func) in lib/configschema/jj.go:11 - JJSchemaLatest returns the latest JJ schema

**ACTION**: Review each of these. If they're for specific bugs/workarounds, ensure they're still needed.

## 5. Duplicate GitHub/HTTP Logic

### GitHub-related Functions (5 total)
**LIKELY DUPLICATE LOGIC**: Multiple implementations of GitHub API access

- GetGitHubToken in lib/ghrelease/ghrelease.go:64
- TestGetGitHubToken in lib/ghrelease/ghrelease_test.go:86
- TestGetGitHubToken in prrun/github_test.go:88
- handleGitHubAsset in want/cmd/handlers.go:56
- handleGitHubAsset in want/handlers.go:423

### Token-related Functions (5 total)
**LIKELY DUPLICATE LOGIC**: Token handling should be centralized

- GetToken in lib/ghclient/token.go:16
- TestGetToken in lib/ghclient/token_test.go:70
- GetGitHubToken in lib/ghrelease/ghrelease.go:64
- TestGetGitHubToken in lib/ghrelease/ghrelease_test.go:86
- TestGetGitHubToken in prrun/github_test.go:88

### Auth-related Functions (5 total)
**LIKELY DUPLICATE LOGIC**: Authentication logic should be in one place

- TestNewClientAddsAuthorizationHeader in lib/ghclient/client_test.go:12
- CreateAuthenticatedRequest in lib/ghrelease/ghrelease.go:69
- CreateAuthenticatedRequestWithContext in lib/ghrelease/ghrelease.go:74
- TestCreateAuthenticatedRequest in lib/ghrelease/ghrelease_test.go:166
- TestCreateAuthenticatedRequest in prrun/github_test.go:138

