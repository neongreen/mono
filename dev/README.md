# dev/

Development scripts for running tools without building them first.

## Purpose

This directory contains wrapper scripts that use `go run` to execute tools from source. This is useful during development when you want to:

- Quickly test changes without running `go build`
- Avoid managing multiple built binaries
- Run tools from any directory in the repo

## Available Scripts

### `ingest-claude-code`

Runs the `ingest-claude-code` tool using `go run`.

**Usage:**
```bash
dev/ingest-claude-code sessions
dev/ingest-claude-code messages
```

**How it works:**
- Uses `go run -C` to build and run from the tool's directory
- Preserves your current working directory
- Passes all arguments through to the tool

## Pattern

Each script follows this pattern:

```bash
#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
tool_dir="$script_dir/../<tool-name>"

exec go run -C "$tool_dir" . "$@"
```

This ensures:
1. The script can be run from anywhere
2. `go run` finds the tool's `go.mod`
3. Your CWD is preserved for the tool's execution
