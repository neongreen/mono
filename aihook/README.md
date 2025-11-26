# aihook

A toolkit for Claude Code hooks including validators and behavior modifiers.

## Overview

`aihook` is a command-line tool that provides various hooks for Claude Code. It can validate shell scripts and modify Claude's behavior (like preventing premature stopping).

## Features

- **Shell Validation Hook**: Validates Bash commands before execution to forbid `cd` invocations outside subshells
- **Prevent-Stop Hook**: Prevents Claude from stopping prematurely, forcing it to continue working
- **Global Install**: Use `--install` flag to add hooks to `~/.claude/settings.json` automatically
- **Shell Parser**: Uses `mvdan.cc/sh/v3/syntax` for accurate shell script parsing
- **Claude Code Integration**: Supports JSON output compatible with Claude Code hooks

## Installation

### From Source

```bash
go install github.com/neongreen/mono/aihook@latest
```

### Local Development

```bash
go build -o /tmp/aihook ./aihook
```

## Usage

### Global --install Flag

All hooks support the `--install` flag which adds the hook to your global Claude settings (`~/.claude/settings.json`) instead of running it:

```bash
# Install the prevent-stop hook
aihook prevent-stop --install

# Install the shell hook
aihook shell --install

# Install shell hook with block-on-cd behavior
aihook shell --install --block-on-cd
```

### Prevent-Stop Hook

The `prevent-stop` subcommand prevents Claude from stopping prematurely:

```bash
# Run as a hook (reads JSON input from stdin)
echo '{}' | aihook prevent-stop

# Install to global Claude settings
aihook prevent-stop --install
```

Output when running:
```json
{
  "continue": false,
  "stopReason": "Keep working! Don't stop until the task is fully complete.",
  "systemMessage": "You are not done yet. Continue working on the task until it is fully complete. Do not stop prematurely."
}
```

### Shell Hook

The `shell` subcommand validates shell scripts and ensures all `cd` commands are executed within subshells:

```bash
# Check a shell script from stdin
echo 'cd /tmp' | aihook shell

# With Claude Code hook format output
echo 'cd /tmp' | aihook shell --claude

# Block execution when cd violations are found
echo 'cd /tmp' | aihook shell --claude --block-on-cd

# Install to global Claude settings
aihook shell --install --block-on-cd
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

Global flags:
- `--install`: Install hook to global Claude settings (`~/.claude/settings.json`) instead of running it

Shell hook flags:
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
  "continue": true,
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "deny",
    "permissionDecisionReason": "Found cd commands outside subshells..."
  }
}
```

## Claude Code Integration

### Manual Configuration

To use `aihook` as a Claude Code hook, add this to your `.claude/settings.json`:

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
    ],
    "Stop": [
      {
        "matcher": "stop",
        "hooks": [
          {
            "type": "command",
            "command": "aihook prevent-stop"
          }
        ]
      }
    ]
  }
}
```

### Automatic Installation

Use the `--install` flag to automatically add hooks to your global Claude settings:

```bash
# Install prevent-stop hook
aihook prevent-stop --install

# Install shell hook with cd blocking
aihook shell --install --block-on-cd
```

The hooks will be added to `~/.claude/settings.json`, creating the file if it doesn't exist.

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
go test ./aihook/... -v
```

### Building

```bash
go build -o /tmp/aihook ./aihook
```

## License

See the root LICENSE file in the repository.
