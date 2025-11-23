# aihook

A validator for Claude Code hooks that enforces shell scripting best practices.

## Overview

`aihook` is a command-line tool designed to validate shell scripts for use in Claude Code PreToolUse hooks. It parses shell syntax and enforces specific rules to ensure code quality and consistency.

## Features

- **PreToolUse Hook**: Validates Bash commands before execution to forbid `cd` invocations outside subshells
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

### Shell Hook

The `shell` subcommand validates shell scripts and ensures all `cd` commands are executed within subshells:

```bash
# Check a shell script from stdin
echo 'cd /tmp' | aihook shell

# With Claude Code hook format output
echo 'cd /tmp' | aihook shell --claude

# Block execution when cd violations are found
echo 'cd /tmp' | aihook shell --claude --block-on-cd
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

### Flags

- `--claude`: Output in JSON format compatible with Claude Code hooks
- `--block-on-cd`: When set, returns exit code 2 for violations (blocks execution). Without this flag, violations are reported but execution is not blocked (exit code 0)

### Exit Codes

- `0`: No violations found (or violations found but `--block-on-cd` not set)
- `2`: Violations detected and `--block-on-cd` flag is set
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

To use `aihook` as a Claude Code PreToolUse hook that validates Bash commands, add this to your `.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "aihook shell --claude --block-on-cd"
          }
        ]
      }
    ]
  }
}
```

The hook will:
- Match any Bash tool invocation (via the `"Bash"` matcher)
- Receive the bash command script on stdin
- Validate that all `cd` commands are in subshells
- With `--block-on-cd`: Return exit code 2 to block execution when violations are found
- Without `--block-on-cd`: Report violations but allow execution (exit code 0)
- Use `--claude` flag to output in the expected JSON format

For more flexible matching, you can use regex patterns like `"Bash.*cd"` to only check Bash commands containing `cd`.

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
