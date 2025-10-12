# claude-trace

A terminal user interface tool for reviewing and annotating Claude Code conversation logs and traces.

## Overview

`claude-trace` helps you review your Claude Code sessions by providing an interactive TUI to:
- Browse through conversation logs
- Mark traces with quick tags (Good, Suspicious, Frustration, Bug, Win)
- Add freeform notes and annotations
- Save your reviews for future reference

## Installation

### Build from source

```bash
cd claude-trace
go build -o claude-trace ./cmd
```

### Run directly

```bash
go run ./cmd
```

## Usage

The tool automatically searches for Claude Code traces in these locations:
- `~/.claude/projects/` (conversation histories in JSONL format)
- `~/.claude/debug/` (debug logs per session)
- `~/.claude/traces/` (user-created traces)
- `~/.config/Claude/traces` (legacy location)
- `~/Library/Application Support/Claude/traces` (legacy location)
- `~/.local/share/Claude/traces` (legacy location)
- `./traces` (current directory)

### Interactive TUI Mode

Simply run:

```bash
./claude-trace
```

This opens an interactive terminal UI for browsing and annotating traces.

### List Traces

To see where traces are located and how many exist in each location:

```bash
./claude-trace list
```

### Extract Traces

To extract all found traces as structured JSON and rendered Markdown files:

```bash
./claude-trace extract
```

By default, this creates an `extracted-traces` directory with two subdirectories:
- `extracted-traces/json/` - One JSON file per trace with structured data
- `extracted-traces/markdown/` - One Markdown file per trace with human-readable formatting

You can specify a custom output directory:

```bash
./claude-trace extract -o /path/to/output
```

The extracted files include:
- Trace content
- Metadata (path, modification time)
- Tags and annotations
- Freeform notes
- Annotation history with timestamps

This is useful for:
- Reviewing traces in your editor
- Processing traces with other tools
- Creating backups of annotated traces
- Annotating traces in Markdown format manually

### Keyboard Shortcuts

#### List View
- `↑/k`: Move up in the list
- `↓/j`: Move down in the list
- `enter`: View selected trace
- `g`: Toggle "Good" tag
- `s`: Toggle "Suspicious" tag
- `f`: Toggle "Frustration" tag
- `b`: Toggle "Bug" tag
- `w`: Toggle "Win" tag
- `n`: Add/edit freeform notes
- `q`: Quit

#### View Mode
- `↑/↓`: Scroll through trace content
- `g/s/f/b/w`: Toggle tags
- `n`: Add/edit notes
- `S`: Save annotations to disk
- `q` or `esc`: Return to list

#### Annotation Mode
- Type your notes freely
- `ctrl+s`: Save and return to view mode
- `esc`: Cancel and return to view mode

## Trace Discovery

Since Claude Code's actual trace storage location may vary by version and platform, this tool searches multiple locations. The primary locations are:

1. **`~/.claude/projects/`** - Contains conversation histories in JSONL format, organized by project path
2. **`~/.claude/debug/`** - Contains debug logs for each session
3. **`~/.claude/traces/`** - User-created trace files

Legacy locations are also searched for compatibility:
- `~/.config/Claude/traces`
- `~/Library/Application Support/Claude/traces`
- `~/.local/share/Claude/traces`

If your traces are stored elsewhere, you can:

1. Create a symbolic link from one of the searched locations to your actual trace directory
2. Copy your traces to the `./traces` directory in the claude-trace folder
3. Modify the `pkg/storage/discovery.go` file to add your custom location

## Annotations

Annotations are saved as JSON files alongside the original traces with a `.annotations.json` extension. For example:
- Original trace: `session-2024-03-15.log`
- Annotations: `session-2024-03-15.log.annotations.json`

The annotation file contains:
- All applied tags
- Freeform notes
- Timestamp history of annotations

## Example Traces

The repository includes sample traces in the `traces/` directory for testing purposes.

## Features

- **Tag System**: Quick keyboard shortcuts for common annotations
  - `g`: Good - Mark sessions that went well
  - `s`: Suspicious - Flag potential issues
  - `f`: Frustration - Note sessions with challenges
  - `b`: Bug - Identify bug-related sessions
  - `w`: Win - Celebrate successful outcomes

- **Freeform Notes**: Add detailed observations and thoughts about each session

- **Persistent Storage**: Annotations are saved to disk and reloaded when viewing traces again

- **File Format Support**: Automatically detects `.log`, `.json`, `.jsonl`, `.txt`, and `.md` files

## Requirements

- Go 1.24.4 or later
- Terminal with color support (for best experience)

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## License

See [LICENSE](../LICENSE) in the repository root.
