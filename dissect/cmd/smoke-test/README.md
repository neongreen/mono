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
- `-files <list>` - Comma-separated list of files to process (required with -url)
- `-diff` - Show git diff after running dissect

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
✓ Go files after: 28
✓ New files created: 5

New files:
  - new.go
  - new_random.go
  - new_random_from_reader.go
  - new_string.go
  - util_new_random_from_pool.go

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
  -files "id.go" \
  -diff

Testing custom project: rs/xid
Cloning and testing project...
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

## Limitations

**Note**: The smoke-test tool currently uses a placeholder ProcessFile function. To test actual dissect functionality, you need to:

1. Build dissect as a library package
2. Import and call ProcessFile from the dissect package
3. Or run dissect as a separate binary

This is intentional to keep the cmd package from being imported as a library.

## Adding New Predefined Projects

Edit `pkg/externaltest/projects.go` to add new projects:

```go
"project-name": {
    Name:        "owner/repo",
    URL:         "https://github.com/owner/repo.git",
    Commit:      "commit-sha-here",
    TargetFiles: []string{"file1.go", "file2.go"},
    ShowDiff:    false,
},
```

## See Also

- Main testing documentation: `../../TESTING.md`
- External test framework: `../../pkg/externaltest/`
- Integration tests: `../main_test.go`
