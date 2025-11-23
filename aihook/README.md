# aihook

A validator for Claude Code hooks that enforces shell scripting best practices.

## Overview

`aihook` is a command-line tool designed to validate shell scripts for use in Claude Code hooks. It parses shell syntax and enforces specific rules to ensure code quality and consistency.

## Features

- **Stop Hook**: Validates shell scripts to forbid `cd` invocations outside subshells
- **Shell Parser**: Uses `mvdan.cc/sh/v3/syntax` for accurate shell script parsing
- **Claude Code Integration**: Supports `--claude` flag for JSON output compatible with Claude Code hooks
- **Comprehensive Validation**: Handles complex scenarios including:
  - Nested subshells
  - Command substitution (`$(...)`)
  - Conditional statements
  - Loops and functions
  - Here documents

## Installation

### From Source

```bash
go install github.com/neongreen/mono/aihook@latest
```

### Local Development

```bash
go build ./aihook
```

## Usage

### Stop Hook

The `stop` subcommand validates shell scripts and ensures all `cd` commands are executed within subshells:

```bash
# Check a shell script from stdin
echo 'cd /tmp' | aihook stop

# With Claude Code hook format output
echo 'cd /tmp' | aihook stop --claude
```

### Examples

**Bad (will fail validation):**
```bash
cd /tmp && ls
```

**Good (will pass validation):**
```bash
(cd /tmp && ls)
```

### Exit Codes

- `0`: No violations found
- `2`: Violations detected (cd commands outside subshells)
- `1`: Parse error or other failure

### Output Formats

#### Standard Output
```
Found cd commands outside subshells:
  Line 1: 'cd' command found outside subshell

All 'cd' commands must be in a subshell. Example:
  # Bad:  cd /tmp && ls
  # Good: (cd /tmp && ls)
```

#### Claude Code Hook Format (`--claude`)
```json
{
  "exit_code": 2,
  "message": "Found cd commands outside subshells:\n  Line 1: 'cd' command found outside subshell\n\nAll 'cd' commands must be in a subshell. Example:\n  # Bad:  cd /tmp && ls\n  # Good: (cd /tmp && ls)\n"
}
```

## Claude Code Integration

To use `aihook` as a Claude Code Stop hook, add this to your `.claude/settings.json`:

```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "aihook stop --claude"
          }
        ]
      }
    ]
  }
}
```

Note: Stop hooks are lifecycle hooks and don't require a `matcher` field (matcher is only for PreToolUse, PermissionRequest, and PostToolUse hooks).

The Stop hook will read shell commands from stdin when Claude Code invokes it. If you want to validate a specific script file instead, you can use: `aihook stop --claude < /path/to/script.sh`

## Why Forbid cd Outside Subshells?

Using `cd` outside subshells can cause unexpected behavior in shell scripts:

1. **Side Effects**: Changes the current directory for the entire script and subsequent commands
2. **Hard to Debug**: Directory changes can happen far from where they're used
3. **Error Prone**: Easy to forget to change back to the original directory
4. **Not Composable**: Makes scripts harder to use in pipelines or as building blocks

By requiring `cd` in subshells `(cd /path && command)`, you ensure:
- Directory changes are localized
- The original directory is automatically restored
- Scripts are more predictable and safer

## Development

### Running Tests

```bash
go test ./aihook -v
```

### Building

```bash
go build ./aihook
```

## License

See the root LICENSE file in the repository.
