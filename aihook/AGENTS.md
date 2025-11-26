# Agent Guidelines for aihook

## Project Overview

`aihook` is a toolkit for Claude Code hooks that provides validators and behavior modifiers. It includes:
- Shell validation hook to enforce `cd` commands are always executed within subshells
- Prevent-stop hook to prevent Claude from stopping prematurely
- Global `--install` flag to add hooks to `~/.claude/settings.json`

## Project Structure

```
aihook/
├── main.go                      # CLI entry point (Cobra setup)
├── pkg/
│   ├── installer/
│   │   ├── installer.go         # Hook installation to ~/.claude/settings.json
│   │   └── installer_test.go    # Installer tests
│   └── validator/
│       ├── validator.go         # Core validation logic
│       └── validator_test.go    # Comprehensive unit tests
├── hook_response_test.go        # Tests for hook response functions
├── integration_test.go          # Integration tests
├── README.md                    # User documentation
└── AGENTS.md                    # This file
```

**Separation of concerns:**
- `main.go`: Cobra CLI setup, flag handling, output formatting, hook commands
- `pkg/validator`: Pure validation logic (AST walking, cd detection)
- `pkg/installer`: Installation logic for adding hooks to Claude settings
- Tests are in each package, testing the pure logic

## Development Guidelines

### Building and Testing

Always run tests and build from the repository root:

```bash
# Run tests
go test ./aihook/...

# Build
go build -o /tmp/aihook ./aihook

# Run
/tmp/aihook shell < script.sh
/tmp/aihook prevent-stop --install
```

### Code Structure

- `main.go` - CLI entry point with Cobra framework
  - Command definitions (shell, prevent-stop)
  - Global `--install` flag
  - Output formatting (regular and --claude JSON)
  - Hook response functions

- `pkg/validator/validator.go` - Core validation logic
  - `Validator` type with `ValidateScript()` method
  - AST walking to detect cd commands
  - Tracks subshell context (both `(...)` and `$(...)`)
  - `FormatViolations()` for user-friendly error messages

- `pkg/installer/installer.go` - Hook installation
  - `InstallHook()` function to add hooks to `~/.claude/settings.json`
  - Handles creating the settings file if it doesn't exist
  - Prevents duplicate hook installation

### Available Commands

1. **`shell`**: Validates Bash commands to ensure `cd` is only used in subshells
   - `--claude`: Output in JSON format
   - `--block-on-cd`: Deny execution when violations found

2. **`prevent-stop`**: Prevents Claude from stopping prematurely
   - Returns `continue: false` with a message telling Claude to keep working

3. **`--install` (global flag)**: Installs the hook to `~/.claude/settings.json`

### Key Implementation Details

1. **Shell Parsing**: Uses `mvdan.cc/sh/v3/syntax` for accurate shell script parsing
2. **AST Walking**: Traverses the syntax tree to find `cd` commands and track subshell context
3. **Subshell Detection**: Handles both explicit subshells `(...)` and command substitutions `$(...)`
4. **Output Formats**: Supports both human-readable and Claude Code JSON formats
5. **Hook Installation**: Modifies `~/.claude/settings.json` to add hooks

### Testing Philosophy

Tests cover:
- Basic cd detection (inside and outside subshells)
- Complex nesting scenarios
- Command substitution
- Edge cases (loops, conditionals, functions)
- Invalid shell syntax
- Multiple cd commands in one script
- Hook installation to settings file
- Duplicate installation prevention
- Hook response JSON structure

All tests must pass before any changes are committed.

### Common Pitfalls

1. **Arithmetic Expansion**: `$((...))` is NOT a subshell, it's arithmetic expansion
2. **Command Substitution**: `$(...)` IS a subshell and cd is allowed inside
3. **Here Documents**: Content inside here-docs should not be parsed as commands

### Claude Code Hook Response Format

For PreToolUse hooks:
```json
{
  "continue": true,
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow|deny",
    "permissionDecisionReason": "..."
  }
}
```

For Stop hooks (prevent-stop):
```json
{
  "continue": false,
  "stopReason": "Keep working!",
  "systemMessage": "Continue working on the task..."
}
```

## Adding New Hook Types

When adding new hook types:

1. Add a new subcommand in `main.go`
2. Create a new validator in `pkg/validator` if needed
3. Add support in `pkg/installer` for the new event type
4. Add comprehensive unit tests
5. Update README.md with usage examples
6. Update this AGENTS.md with implementation notes

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `mvdan.cc/sh/v3/syntax` - Shell script parser
- `github.com/neongreen/mono/lib/version` - Shared version command

All dependencies are managed via the repository's root go.mod file.
