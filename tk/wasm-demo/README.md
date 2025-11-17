# tk WebAssembly Demo

This is a browser-based demo of tk that runs entirely in your web browser using WebAssembly.

## Features

- ✨ Run tk commands directly in the browser
- 💾 In-memory SQLite database powered by ncruces/go-sqlite3 WASM
- 🎨 Terminal-like interface with command history
- 🚀 No server-side processing - everything runs client-side
- ⚡ Full tk functionality available (tasks, projects, status tracking, relations, etc.)

## Quick Start

### Option 1: Using Dagger (Recommended)

The easiest way to build and serve the demo:

```bash
# From the repository root
cd .dagger

# Build and serve on port 8080
dagger call project tk wasm-serve up --ports 8080:8080

# Open http://localhost:8080 in your browser
```

See [DAGGER.md](DAGGER.md) for more options.

### Option 2: Manual Build

Build the demo manually using the provided scripts:

```bash
./build.sh
./serve.sh
```

Then open http://localhost:8080 in your browser.

### Try some commands

Once the page loads, try these commands:

```bash
# Initialize the database
init

# Create a project
project-create demo "Demo Project"

# Create some tasks
new "Setup development environment" --project demo
new "Write documentation" --project demo
new "Add tests" --project demo

# List all tasks
ls

# Show a specific task
show demo-1

# Mark a task as in progress
mark demo-1 wip

# Mark a task as done
mark demo-2 done

# List tasks filtered by project
ls --project demo

# Get help on any command
--help
new --help
mark --help
```

## How it works

The demo uses:
- **Go WASM**: tk is compiled to WebAssembly using Go's js/wasm target
- **ncruces/go-sqlite3**: Pure Go SQLite implementation with WASM support (replaces modernc.org/sqlite which doesn't support WASM)
- **Shared in-memory database**: Uses `file:tk?mode=memory&cache=shared` URI to maintain state across command invocations
- **JavaScript interop**: Go functions are exposed to JavaScript for command execution via `syscall/js`
- **Cobra output capture**: Uses Cobra's `SetOut`/`SetErr` methods to capture command output

## Technical Implementation

### WASM Compilation

The project uses Go build tags to conditionally compile different SQLite drivers:

- **Non-WASM** (`driver_default.go`): Uses `modernc.org/sqlite` (pure Go, no CGO)
- **WASM** (`driver_wasm.go`): Uses `github.com/ncruces/go-sqlite3` (WASM-compatible)

### Database Persistence

The WASM demo uses a shared in-memory database with the URI `file:tk?mode=memory&cache=shared`. This ensures:
- Multiple connections share the same in-memory database instance
- Data persists across command invocations within the same browser session
- Each command (init, new, ls, mark, etc.) can see data from previous commands

Without `cache=shared`, each SQLite connection to `:memory:` creates a separate database instance that is destroyed when the connection closes, making it impossible to maintain state across commands.

### Filesystem Adaptations

Since WASM has limited filesystem support:
- Database uses shared in-memory URI (`file:tk?mode=memory&cache=shared`) to persist across commands
- File locking operations are no-ops in WASM (`lock_wasm.go`)
- Directory creation skipped for in-memory databases
- WAL mode disabled for in-memory databases

## Limitations

- **No persistence**: Data is lost when you refresh the page (in-memory database by design)
- **No file operations**: Commands that work with files (attach, import, etc.) won't work
- **No sync**: Remote sync features are not available
- **Limited I/O**: Some terminal features (colors, formatting) may not display correctly
- **Size**: The WASM binary is large (~25MB) due to including the entire tk application

## Known Issues

- Console log output from `init` command appears in browser console rather than terminal
- Some ANSI color codes may not render in the browser terminal

## Browser Compatibility

Works with modern browsers that support WebAssembly:
- Chrome 57+
- Firefox 52+
- Safari 11+
- Edge 16+

## Development

The main WASM wrapper is in `main.go`. It:
1. Sets up the WASM environment (HOME, USER, TK_DB_PATH)
2. Captures stdout/stderr from tk commands using Cobra's output redirection
3. Exposes JavaScript functions (`tkExecute`, `tkInit`) for command execution
4. Manages the command lifecycle and returns results to JavaScript

The HTML interface (`index.html`) provides:
- Terminal-like UI with dark theme
- Command input with history (arrow keys to navigate)
- Output display with success/error styling
- Getting started guide with example commands

## Troubleshooting

**WASM module fails to load**: Make sure you built with `./build.sh` and are serving via HTTP (not file://)

**Commands don't execute**: Check the browser console for errors. The WASM module must be fully loaded (you'll see "✓ tk WASM module loaded successfully!")

**Database errors**: The database is in-memory only. Run `init` first to initialize it after page load.

**Build errors**: Ensure you have Go 1.21+ installed and run from the monorepo root or tk directory.
