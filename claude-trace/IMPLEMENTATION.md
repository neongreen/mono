# claude-trace Implementation Summary

## Overview

`claude-trace` is a terminal user interface (TUI) tool for reviewing and annotating Claude Code conversation logs. It provides an interactive way to go through session traces and mark them up with tags and notes.

## Architecture

### Project Structure

```
claude-trace/
├── cmd/
│   └── main.go           # Application entry point
├── pkg/
│   ├── storage/          # Trace file management
│   │   ├── discovery.go  # Finds Claude Code trace locations
│   │   └── trace.go      # Load/save traces and annotations
│   └── tui/              # Terminal user interface
│       └── model.go      # Bubbletea TUI model and views
├── traces/               # Sample trace files for testing
├── README.md             # User documentation
└── go.mod                # Go module dependencies
```

## Key Components

### 1. Trace Discovery (`pkg/storage/discovery.go`)

The tool searches for Claude Code traces in multiple platform-specific locations:

**Linux/XDG:**
- `~/.config/Claude/traces`
- `~/.local/share/Claude/traces`
- `~/.local/share/claude-code/traces`

**macOS:**
- `~/Library/Application Support/Claude/traces`
- `~/Library/Application Support/claude-code/traces`

**Windows:**
- `%APPDATA%/Claude/traces`
- `%LOCALAPPDATA%/Claude/traces`

**Fallback:**
- `./traces` (current directory for testing)

The tool automatically finds all available trace locations and loads files from them.

### 2. Trace Management (`pkg/storage/trace.go`)

**Data Structure:**
```go
type Trace struct {
    Path         string            // Full path to trace file
    Name         string            // Filename
    Content      string            // File content
    ModTime      time.Time         // Last modification time
    Annotations  []Annotation      // History of annotations
    Tags         map[string]bool   // Active tags
    FreeformNote string            // User's notes
}
```

**Features:**
- Loads trace files (`.log`, `.json`, `.txt`, `.md`)
- Sorts traces by modification time (newest first)
- Saves annotations as JSON files (`.annotations.json`)
- Loads existing annotations on startup

### 3. Terminal UI (`pkg/tui/model.go`)

Built with [Charm's Bubbletea](https://github.com/charmbracelet/bubbletea) framework.

**Three Modes:**

1. **List Mode** - Browse traces
   - Navigate with arrow keys or vim bindings (j/k)
   - Quick tag application
   - Shows active tags inline

2. **View Mode** - Read trace content
   - Scrollable viewport
   - Tag display
   - Notes display
   - Apply tags while viewing

3. **Annotate Mode** - Add freeform notes
   - Multi-line text input
   - Save with Ctrl+S
   - Cancel with Esc

## Tag System

Five quick-access tags for common annotations:

- **G** (Good) - Sessions that went well
- **S** (Suspicious) - Potential issues
- **F** (Frustration) - Challenging sessions
- **B** (Bug) - Bug-related conversations
- **W** (Win) - Successful outcomes

Tags can be toggled on/off and are saved with timestamp history.

## Annotation Persistence

Annotations are saved as JSON files alongside traces:

```
session-2024-03-15.log              # Original trace
session-2024-03-15.log.annotations.json  # Annotations
```

The annotation file contains:
- All trace metadata
- Active tags
- Freeform notes
- Complete annotation history with timestamps

## Dependencies

- **bubbletea** (v1.3.10) - TUI framework
- **lipgloss** (v1.1.0) - Terminal styling
- **bubbles** (v0.21.0) - TUI components (viewport, textarea)

All dependencies checked for vulnerabilities - none found.

## Usage Flow

1. **Discovery** - Tool searches common locations for traces
2. **Loading** - Loads all trace files and existing annotations
3. **Browsing** - User navigates through traces in list view
4. **Reviewing** - User views trace content in detail
5. **Annotating** - User adds tags and notes
6. **Saving** - Annotations saved to disk (manual with 'S' or automatic with Ctrl+S)

## Testing

The repository includes three sample trace files in `traces/`:
- `session-2024-03-15.log` - Simple successful session
- `session-2024-03-16.log` - Debugging session with frustration
- `session-2024-03-17.log` - Test writing session

These can be used to explore the tool's functionality.

## Future Enhancements

Potential improvements (not implemented):
- Search/filter traces by content or tags
- Export annotations to different formats (markdown, CSV)
- Aggregate statistics (most common tags, session duration analysis)
- Integration with actual Claude Code trace format parsing
- Support for trace format auto-detection
- Multi-trace comparison view
