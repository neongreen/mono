# Agent Guidelines for aihook

## Project Overview

`aihook` is a validator for Claude Code hooks that enforces shell scripting best practices. The primary use case is validating that `cd` commands are always executed within subshells.

## Development Guidelines

### Building and Testing

Always run tests and build from the repository root:

```bash
# Run tests
go test ./aihook -v

# Build
go build ./aihook

# Run
./aihook stop < script.sh
```

### Code Structure

- `main.go` - CLI entry point, command definitions, and core validation logic
- `main_test.go` - Comprehensive unit tests covering all validation scenarios

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
2. Implement the validation logic
3. Add comprehensive unit tests
4. Update README.md with usage examples
5. Update this AGENTS.md with implementation notes

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `mvdan.cc/sh/v3/syntax` - Shell script parser
- `github.com/neongreen/mono/lib/version` - Shared version command

All dependencies are managed via the repository's root go.mod file.
