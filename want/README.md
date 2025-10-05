# want

Interactive task fulfillment tool for macOS.

## ⚠️ Implementation Status

### ✅ Currently Implemented
- [x] **Install tools via mise** - Request and install tools using `want <tool>`
- [x] **Check if tools are already available** - Detects existing installations
- [x] **Dry-run mode** - Preview actions with `--dry-run` flag
- [x] **Compound commands with parameters** - Transform and execute commands
  - [x] **`want json <command>`** - Convert command output to JSON using `jc`
  - [x] **`want md <url>`** - Convert URL to markdown using `markitdown` or `pure.md`
- [x] **Basic CLI interface** - Help, version, and command structure
- [x] **Command stubs** - `list`, `check`, `forget` commands (not functional yet)

### ❌ Not Yet Implemented
- [ ] **Clone git repositories** - `want github.com/user/repo` not working yet
- [ ] **Track requirements** - No persistence of what you've requested
- [ ] **List tracked items** - `want list` shows placeholder message
- [ ] **Check status** - `want check` shows placeholder message
- [ ] **Forget requirements** - `want forget` shows placeholder message
- [ ] **Multiple provider support** - Only mise is supported; no Homebrew, GitHub releases, etc.
- [ ] **Interactive prompts** - No interactive selection when multiple options exist
- [ ] **Preference learning** - No storage of user preferences
- [ ] **Configuration persistence** - No `~/.config/want/` directory created

## Overview

`want` helps you get things you need on your system through CLI commands. It's an interactive assistant that respects your preferences.

```bash
want jujutsu                # Install a tool via mise
want --dry-run jujutsu      # Preview what would be done
want json ps                # Get process list as JSON (installs jc if needed)
want md https://example.com # Convert URL to markdown
want github.com/user/repo   # Clone a repository (NOT YET IMPLEMENTED)
want list                   # Show what you have (NOT YET IMPLEMENTED)
```

## Installation

### With Go

```bash
cd want
go build
```

### With mise

Install using [mise](https://mise.jdx.dev/) with the Go backend:

```bash
mise use -g go:github.com/neongreen/mono/want@main
```

Or add to your `.mise.toml`:

```toml
[tools]
"go:github.com/neongreen/mono/want" = "main"
```

## Usage

### Install a tool (✅ WORKING)

```bash
want jujutsu                    # Install jujutsu via mise
```

If the tool is already available (installed via brew, apt, or any other method), `want` will detect it and skip installation:

```bash
$ want jq
✓ jq is already available
  Location: /usr/bin/jq
```

### Dry-run mode (✅ WORKING)

Preview what would be done without actually doing it:

```bash
want --dry-run jujutsu
want --dry-run json ps
```

### Compound commands with parameters (✅ WORKING)

Transform command output or URLs using specialized tools:

```bash
# Convert command output to JSON
want json ps                    # Get running processes as JSON
want json ls -la                # Get directory listing as JSON
want json df -h                 # Get disk usage as JSON

# Convert URLs to markdown
want md https://example.com     # Convert webpage to markdown
want md https://news.ycombinator.com
```

**How it works:**
- `want json <command>` automatically installs `jc` if needed, then runs `jc <command>`
- `want md <url>` tries to install `markitdown` via pip, or falls back to the `pure.md` web service

### Commands not yet functional (❌ NOT IMPLEMENTED)

```bash
want github.com/user/repo       # Clone a repository - NOT WORKING YET
want list                       # View tracked requirements - NOT WORKING YET
want check                      # Check status of requirements - NOT WORKING YET
want forget <requirement>       # Remove from tracking - NOT WORKING YET
```

## How it works

- CLI-first interaction (not YAML editing)
- Presents multiple options when available
- Learns your preferences over time
- Doesn't reset existing work

Configuration will be stored at `~/.config/want/`.

See [DESIGN.md](DESIGN.md) for complete design.
