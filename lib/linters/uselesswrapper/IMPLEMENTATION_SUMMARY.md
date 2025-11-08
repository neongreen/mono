# Uselesswrapper Linter Implementation Summary

## Overview

Successfully created a custom Go static analyzer that detects useless function wrappers in response to feedback on PR #259.

## What Was Built

### Core Analyzer (`uselesswrapper.go`)
- Uses Go's `go/analysis` framework
- Performs AST (Abstract Syntax Tree) analysis
- Detects functions that:
  - Contain only a single return statement
  - Return the result of calling another function
  - Pass all parameters unchanged (in the same order)
  - Add no logic, validation, or error handling

### Test Suite (`uselesswrapper_test.go`, `testdata/`)
- Comprehensive test cases covering:
  - Useless wrappers that should be detected
  - Valid wrappers that should NOT be detected (with error handling, transformations, multiple statements, etc.)
  - Edge cases (methods, parameter reordering, etc.)
- All tests pass

### CLI Tool (`cmd/uselesswrapper/main.go`)
- Standalone executable using `singlechecker.Main`
- Can be run on any Go package or module
- Provides clear, actionable error messages

### Documentation
- **README.md**: Usage guide with examples of detected and valid patterns
- **DETECTION_RESULTS.md**: Analysis of findings in the tk codebase (13 useless wrappers found)

## How to Use

```bash
# From repository root, build the tool
go build -o uselesswrapper ./lib/linters/uselesswrapper/cmd/uselesswrapper

# Run on a package
./uselesswrapper ./path/to/package

# Run on entire module
./uselesswrapper ./...

# Or run without building
go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./...
```

## Monorepo Integration

This linter is part of the monorepo and uses the root `go.mod`:
- No separate module needed
- Shares dependencies with other tools in the repo
- Can be easily imported by other packages in the monorepo

To add as a mise task, add to root `mise.toml`:
```toml
[tasks."lint:uselesswrapper"]
description = "Detect useless function wrappers"
run = "go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./..."
```

## Results on tk Codebase

Found 13 useless wrapper functions:
- 3x `getCurrentUser()` wrappers (the original issue from PR #259)
- 2x type constructor wrappers (`NewProjectRef`, `NewTaskRef`)
- 6x JSON save wrappers across multiple files
- 1x command execution wrapper (`Execute()`)
- 1x test helper wrapper (`slices.Contains()`)

## Integration Options

The analyzer can be:
1. **Run manually** - as shown above
2. **Added to CI/CD** - fail builds when useless wrappers are introduced
3. **Integrated with golangci-lint** - as a custom analyzer plugin
4. **Pre-commit hook** - catch issues before commit

## Technical Details

- **Dependencies**: `golang.org/x/tools/go/analysis`
- **Module**: `github.com/neongreen/mono/lib/linters/uselesswrapper`
- **Go version**: Compatible with Go 1.16+
- **Exit code**: 3 when issues are found (standard for analysis tools)

## Why This Matters

Useless wrappers:
- Add unnecessary indirection
- Create duplicate code across packages
- Increase maintenance burden
- Cause confusion about which function to call
- Provide no abstraction or encapsulation value

The linter helps maintain clean, straightforward code by identifying these patterns early.

## Commits

- `5d56093` - Add uselesswrapper linter to detect unnecessary function wrappers
- `21b0c0c` - Add detection results documentation for uselesswrapper linter
