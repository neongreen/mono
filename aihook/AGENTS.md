# Agent Guidelines for aihook

## Project Overview

`aihook` is a validator for Claude Code hooks that enforces shell scripting best practices. The primary use case is validating that `cd` commands are always executed within subshells.

## Project Structure

```
aihook/
├── main.go                      # CLI entry point (Cobra setup)
├── pkg/
│   └── validator/
│       ├── validator.go         # Core validation logic
│       └── validator_test.go    # Comprehensive unit tests
├── README.md                    # User documentation
└── AGENTS.md                    # This file
```

**Separation of concerns:**
- `main.go`: Cobra CLI setup, flag handling, output formatting
- `pkg/validator`: Pure validation logic (AST walking, cd detection)
- Tests are in the validator package, testing the pure logic

## Development Guidelines

### Building and Testing

Always run tests and build from the repository root:

```bash
# Run tests
go test ./aihook/...

# Build
go build ./aihook

# Run
./aihook shell < script.sh
```

### Code Structure

- `main.go` (86 lines) - CLI entry point with Cobra framework
  - Command definitions and flag handling
  - Output formatting (regular and --claude JSON)
  - Calls into validator package for logic

- `pkg/validator/validator.go` (106 lines) - Core validation logic
  - `Validator` type with `ValidateScript()` method
  - AST walking to detect cd commands
  - Tracks subshell context (both `(...)` and `$(...)`)
  - `FormatViolations()` for user-friendly error messages

- `pkg/validator/validator_test.go` (330 lines) - Comprehensive tests
  - 41 test scenarios covering all edge cases
  - Tests the validator package directly

### Key Implementation Details

1. **Shell Parsing**: Uses `mvdan.cc/sh/v3/syntax` for accurate shell script parsing
2. **AST Walking**: Traverses the syntax tree to find `cd` commands and track subshell context
3. **Subshell Detection**: Handles both explicit subshells `(...)` and command substitutions `$(...)`
4. **Output Formats**: Supports both human-readable and Claude Code JSON formats

### Testing Philosophy

Tests cover:
- Basic cd detection (inside and outside subshells)
- Complex nesting scenarios
- Command substitution
- Edge cases (loops, conditionals, functions)
- Invalid shell syntax
- Multiple cd commands in one script

All tests must pass before any changes are committed.

### Common Pitfalls

1. **Arithmetic Expansion**: `$((...))` is NOT a subshell, it's arithmetic expansion
2. **Command Substitution**: `$(...)` IS a subshell and cd is allowed inside
3. **Here Documents**: Content inside here-docs should not be parsed as commands

### Claude Code Hook Format

When `--claude` flag is used, output must be JSON with:
- `exit_code`: Integer (0 for success, 2 for violations, 1 for errors)
- `message`: String containing the human-readable message

Example:
```json
{
  "exit_code": 2,
  "message": "Found cd commands outside subshells:\n  Line 1: 'cd' command found outside subshell\n\nAll 'cd' commands must be in a subshell. Example:\n  # Bad:  cd /tmp && ls\n  # Good: (cd /tmp && ls)\n"
}
```

## Adding New Hook Types

When adding new hook types in the future:

1. Add a new subcommand in `main.go`
2. Create a new validator in `pkg/validator` if needed
3. Add comprehensive unit tests in the validator package
4. Update README.md with usage examples
5. Update this AGENTS.md with implementation notes

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `mvdan.cc/sh/v3/syntax` - Shell script parser
- `github.com/neongreen/mono/lib/version` - Shared version command

All dependencies are managed via the repository's root go.mod file.
