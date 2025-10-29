# Smoke Test Tool

A command-line tool for running ad-hoc smoke tests of dissect on any GitHub project.

## Overview

The smoke-test tool allows you to quickly test dissect on external Go projects without writing test code. It's useful for:

- Testing dissect on new projects during development
- Validating behavior on specific codebases
- Debugging issues with external projects
- Exploring how dissect handles different code patterns

## Usage

### List Available Projects

```bash
go run ./cmd/smoke-test -list
```

Shows all predefined projects that can be tested.

### Test a Predefined Project

```bash
# Basic test
go run ./cmd/smoke-test -project google/uuid

# With git diff output
go run ./cmd/smoke-test -project google/uuid -diff
```

### Test a Custom Project

```bash
go run ./cmd/smoke-test \
  -url https://github.com/owner/repo.git \
  -commit abc123def \
  -files "file1.go,file2.go" \
  -diff
```

### Command-Line Flags

- `-list` - List all available predefined projects
- `-project <name>` - Test a predefined project (e.g., "google/uuid")
- `-url <url>` - Git clone URL for custom project
- `-commit <sha>` - Git commit SHA to checkout (default: HEAD)
- `-diff` - Show git diff after running dissect

**Note**: The tool processes **all Go files** in the project (except test files, cmd packages, and deeply nested internal packages), not just specific target files.

## Examples

### Example 1: Test google/uuid

```bash
$ go run ./cmd/smoke-test -project google/uuid

Using predefined project: google/uuid
Cloning and testing project...

============================================================
SMOKE TEST RESULTS
============================================================
✓ Project directory: /tmp/dissect_external_google_uuid_1234567890/uuid
✓ Go files before: 23
✓ Go files after: 38
✓ New files created: 15

New files:
  - clone.go
  - dce.go
  - marshal.go
  - new.go
  - new_random.go
  - new_random_from_reader.go
  - new_string.go
  - new_v_6_with_time.go
  - new_v_7_with_time.go
  - node_i_d.go
  - null.go
  - sql.go
  - time.go
  - util_new_random_from_pool.go
  - validate.go

✓ Build passed: true
✓ Tests passed: true

✓ Smoke test passed successfully!

Project preserved at: /tmp/dissect_external_google_uuid_1234567890/uuid
```

### Example 2: Test with Diff

```bash
$ go run ./cmd/smoke-test -project google/uuid -diff

[... output as above ...]

------------------------------------------------------------
GIT DIFF
------------------------------------------------------------
diff --git a/version4.go b/version4.go
index 7697802..6c9c98b 100644
--- a/version4.go
+++ b/version4.go
[... diff content ...]
```

### Example 3: Test Custom Project

```bash
$ go run ./cmd/smoke-test \
  -url https://github.com/rs/xid.git \
  -commit 475c481 \
  -diff

Testing custom project: rs/xid
Cloning and testing project...
(will process all Go files in the project)
[... results ...]
```

## How It Works

1. **Clone**: Clones the project to a temporary directory
2. **Checkout**: Checks out the specified commit
3. **Verify**: Runs `go build` and `go test` before dissect
4. **Dissect**: Runs dissect on the specified files
5. **Validate**: Verifies the project still builds and tests pass
6. **Report**: Shows summary with file counts and optional diff

## Output

The tool provides:

- Success/failure status
- File count before and after dissect
- List of newly created files
- Build and test pass/fail status
- Optional git diff of changes
- Location of preserved project directory

## Implementation

The smoke-test tool calls the `dissect explode` command as an external binary for each Go file found in the project. This ensures it's testing the actual dissect functionality as users would experience it.

**Requirements**: The `dissect` binary must be in your PATH or you must run this from the dissect project directory.

## Adding New Predefined Projects

Edit `pkg/externaltest/projects.go` to add new projects:

```go
"project-name": {
    Name:     "owner/repo",
    URL:      "https://github.com/owner/repo.git",
    Commit:   "commit-sha-here",
    ShowDiff: false,
},
```

All Go files in the project will be processed automatically.

## See Also

- Main testing documentation: `../../TESTING.md`
- External test framework: `../../pkg/externaltest/`
- Integration tests: `../main_test.go`
