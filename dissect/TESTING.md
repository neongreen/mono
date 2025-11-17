# Testing Guide

This document explains the testing infrastructure for the dissect project.

## Overview

The project uses file-based integration testing with TOML configuration files. Each test case defines input Go files and expected output files, making tests easy to read, write, and review.

## Running Tests

### Run All Tests

```bash
go test ./... -v
```

### Run Specific Test

```bash
go test ./cmd -v -run TestAllDissectIntegration/simple.toml
```

### Run Tests with Race Detection

```bash
go test ./... -v -race
```

### Run Tests in Parallel

Tests automatically run in parallel using `t.Parallel()` for better performance.

### Run External Project Test

The external project integration test clones a real Go project and runs dissect on it:

```bash
go test ./cmd -v -run TestExternalProjectIntegration
```

This test is skipped in short mode. To skip it:

```bash
go test ./cmd -v -short
```

**Note:** This test requires an internet connection to clone the external repository.

## Test Structure

### Directory Layout

```
dissect/
├── cmd/
│   ├── main_test.go         # Integration test runner
│   └── ...
├── tests/                   # Test cases
│   ├── simple.toml
│   ├── with_test.toml
│   ├── internal_extraction.toml
│   ├── no_refactor.toml
│   ├── ignore_file.toml
│   └── v2_import_test.toml
└── pkg/
    ├── testutils/           # Test utilities
    └── goutils/
        └── normalize_imports_test.go  # Unit tests
```

## File-Based Testing

### TOML Test Format

Each test is a TOML file with two sections:

```toml
# Input files before running dissect
[files_in]
"go.mod" = '''
module example.com/testproject
go 1.20
'''

"main.go" = '''
package main

func main() {
    foo()
}

func foo() {
    // implementation
}
'''

# Expected output files after running dissect
[files_out]
"go.mod" = '''
module example.com/testproject
go 1.20
'''

"main.go" = '''
package main

func main() {
    foo()
}
'''

"util_foo.go" = '''
package main

func foo() {
    // implementation
}
'''
```

### Test Naming Convention

- Test files: `<test-name>.toml`
- Test name becomes the subtest name in Go
- Use descriptive names like `simple.toml`, `with_test.toml`

## Test Cases

### `simple.toml`

Tests basic function extraction from a simple Go file with multiple functions.

**Validates:**
- Regular function extraction (`foo()` → `util_foo.go`)
- Function with parameters (`barQuix(x int)` → `util_bar_quix.go`)
- Method extraction (`(*MyStruct).Baz()` → `mystruct_baz.go`)
- Type definitions remain in original file
- `main()` function stays in place

### `with_test.toml`

Tests extraction of functions from test files.

**Validates:**
- Test file detection (`*_test.go`)
- Test helper function extraction
- Test functions (`Test*`) remain in original file
- Benchmark functions (`Benchmark*`) remain in original file
- Helper functions are extracted to `internal/testutils`

### `internal_extraction.toml`

Tests extraction of test helpers to `internal/testutils` package.

**Validates:**
- Test helpers in `package main_test` → `internal/testutils`
- Package declaration updates
- Import path resolution
- Multiple helper functions

### `no_refactor.toml`

Tests files that should not be refactored.

**Validates:**
- Files with only `main()` are skipped
- Files with only one function are skipped
- Appropriate skip reasons are logged

### `ignore_file.toml`

Tests respecting `.gitignore` patterns.

**Validates:**
- Files matching `.gitignore` patterns are not processed
- Other files in the same directory are processed normally

### `v2_import_test.toml`

Tests handling of Go modules with `/v2` or higher version suffixes.

**Validates:**
- Correct import path generation for versioned modules
- Module path parsing with version suffixes

## External Project Integration Tests

The `TestExternalProjects` suite validates dissect on real-world Go projects using a reusable testing framework.

### What It Tests

This test suite:
1. Clones external Go projects at specific commits
2. Runs dissect on **all Go files** in the project (except test files, cmd packages, and deeply nested internal packages)
3. Verifies the project still compiles with `go build`
4. Verifies the project's test suite still passes with `go test`

### Purpose

This test ensures that:
- dissect works correctly on real production code
- The refactored code maintains correctness when processing entire codebases
- All imports and dependencies are handled properly
- The test suite continues to pass after refactoring

### Current Test Projects

**google/uuid**
- Small, focused single-package library
- Processes all main package Go files (15 new files created)
- Full test suite passes after refactoring

**segmentio/ksuid**
- K-Sortable Unique Identifier library
- Processes all package files (13 new files created)
- Project compiles and tests pass

### Running the Tests

```bash
# Run all external project tests
go test ./cmd -v -run TestExternalProjects

# Run a specific project test
go test ./cmd -v -run TestExternalProjects/google

# Skip external project tests (use short mode)
go test ./cmd -v -short

# Show git diff of changes made by dissect
DISSECT_SHOW_DIFF=1 go test ./cmd -v -run TestExternalProjects
```

### Adding New Projects

Projects are defined in `pkg/externaltest/projects.go`:

```go
"project-name": {
    Name:        "owner/repo",
    URL:         "https://github.com/owner/repo.git",
    Commit:      "commit-sha",
    TargetFiles: []string{"file1.go", "file2.go"},
    ShowDiff:    false,
}
```

### Smoke Testing Tool

For ad-hoc testing on any GitHub project, use the `smoke-test` tool:

```bash
# List available predefined projects
go run ./cmd/smoke-test -list

# Test a predefined project
go run ./cmd/smoke-test -project google/uuid

# Test a predefined project with diff
go run ./cmd/smoke-test -project google/uuid -diff

# Test a custom project
go run ./cmd/smoke-test \
  -url https://github.com/owner/repo.git \
  -commit abc123 \
  -files "file1.go,file2.go" \
  -diff
```

The smoke-test tool:
- Clones the project to a temporary directory
- Runs dissect on specified files
- Validates compilation and tests
- Optionally shows git diff
- Preserves the project directory for inspection

### Technical Details

- **Framework**: Reusable `pkg/externaltest` package
- **gopls requirement**: Automatically installs gopls and goimports if not present
- **Test duration**: ~1-20 seconds per project (depends on size)
- **Internet required**: Yes (to clone repositories)
- **Preserved directories**: `/tmp/dissect_external_*` for debugging

### Debugging

The tests preserve cloned repositories for inspection:

```bash
# Find test directories
ls -td /tmp/dissect_external_* | head -5

# Examine a specific test
cd /tmp/dissect_external_google_uuid_*/uuid/

# View git diff
git diff

# List new files created
git ls-files --others --exclude-standard
```

## Test Execution Flow

The test runner (`cmd/main_test.go`) executes each test as follows:

1. **Setup**
   - Create temporary directory
   - Write input files from TOML
   - Run `go mod tidy` to initialize module

2. **Execution**
   - Run `dissect` on all Go files
   - Process functions iteratively

3. **Normalization**
   - Run `go fmt` on output
   - Normalize imports for consistent comparison
   - Apply same normalization to expected output

4. **Verification**
   - Compare actual output with expected output
   - Check file contents match exactly
   - Verify no unexpected files were created
   - Run `go build` to ensure code compiles

5. **Cleanup**
   - Log all output files for debugging
   - Leave temporary directory for manual inspection

## Test Utilities

### `pkg/testutils/compare_directories.go`

Compares actual output directory with expected files:

```go
func CompareDirectories(t *testing.T, expectedFiles map[string]string, actualDir string) error
```

**Features:**
- Compares file contents
- Reports missing files
- Reports unexpected files
- Provides detailed diffs

### `pkg/goutils/normalize_imports.go`

Normalizes import formatting for consistent comparison:

```go
func NormalizeImports(content string) (string, error)
```

**Normalization:**
- Groups imports consistently
- Removes blank lines in import blocks
- Standardizes formatting

## Adding a New Test

1. **Create TOML file**

   ```bash
   cat > tests/my_test.toml << 'EOF'
   [files_in]
   "go.mod" = '''
   module example.com/test
   go 1.20
   '''
   
   "main.go" = '''
   package main
   
   func main() {}
   func foo() {}
   '''
   
   [files_out]
   "go.mod" = '''
   module example.com/test
   go 1.20
   '''
   
   "main.go" = '''
   package main
   
   func main() {}
   '''
   
   "util_foo.go" = '''
   package main
   
   func foo() {}
   '''
   EOF
   ```

2. **Run the test**

   ```bash
   go test ./cmd -v -run TestAllDissectIntegration/my_test.toml
   ```

3. **Debug if needed**

   Check the temporary directory logged in test output:
   ```
   /tmp/dissect_my_test.toml_<random>/
   ```

## Debugging Tests

### View Test Output

Tests log detailed information:
- Temporary directory location
- Files created from test data
- Processing steps
- Output directory state

Enable debug logging:
```bash
go test ./cmd -v -run TestAllDissectIntegration/simple.toml 2>&1 | grep DEBUG
```

### Inspect Temporary Directory

Tests preserve temporary directories for debugging:
```bash
ls -la /tmp/dissect_simple.toml_*/
```

### Compare Files Manually

```bash
cd /tmp/dissect_simple.toml_<random>/
diff -u <(cat expected.go) <(cat actual.go)
```

## Unit Tests

Some packages have unit tests for specific functions:

### `pkg/goutils/normalize_imports_test.go`

Tests import normalization logic:
- Empty imports
- Single imports
- Multiple imports
- Grouped imports

Run unit tests:
```bash
go test ./pkg/goutils -v
```

## Test Coverage

Generate coverage report:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

View coverage by package:
```bash
go test ./... -cover
```

## Best Practices

1. **One test per scenario** - Each TOML file tests a specific scenario
2. **Clear test names** - Use descriptive names like `with_test.toml` or `internal_extraction.toml`
3. **Minimal test cases** - Keep test files as small as possible while still testing the feature
4. **Test edge cases** - Consider boundary conditions and special cases
5. **Test real-world scenarios** - Include complex cases that combine multiple features
6. **Document test purpose** - Add comments in TOML files explaining what's being tested

## Statistics

Current test suite:
- **6 integration test cases** (TOML files)
- **100% pass rate** with parallel execution
- **Comprehensive coverage** of core functionality:
  - Function extraction
  - Test file handling
  - Import management
  - Package declaration updates
  - Build verification

## Continuous Integration

Tests are designed to run in CI environments:
- No external dependencies beyond Go toolchain
- Deterministic output
- Fast execution with parallel tests
- Clear error messages

## Troubleshooting

### Test Fails with "go build failed"

Check the temporary directory for build errors:
```bash
cd /tmp/dissect_<test>_<random>/
go build ./...
```

### Test Fails with "Directory comparison failed"

Compare expected vs actual files:
```bash
# Check test output for file differences
go test ./cmd -v -run <test> 2>&1 | grep "expected"
```

### gopls Not Found

Ensure gopls is installed:
```bash
go install golang.org/x/tools/gopls@latest
```

### Import Normalization Issues

Check import formatting:
```bash
goimports -w .
```

## Contributing

When adding new functionality:

1. Write a test case in TOML format
2. Run the test to see it fail
3. Implement the feature
4. Verify the test passes
5. Add edge cases if needed

## Additional Resources

- [Go testing documentation](https://golang.org/pkg/testing/)
- [TOML specification](https://toml.io/)
- [gopls documentation](https://github.com/golang/tools/tree/master/gopls)
- [plan.md](plan.md) - Original testing plan
