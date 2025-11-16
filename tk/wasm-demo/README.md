# tk WebAssembly Demo

This is a browser-based demo of tk that runs entirely in your web browser using WebAssembly.

## Features

- ✨ Run tk commands directly in the browser
- 💾 In-memory SQLite database (data is not persisted between page reloads)
- 🎨 Terminal-like interface with command history
- 🚀 No server-side processing - everything runs client-side

## Quick Start

### Build the demo

```bash
./build.sh
```

This will:
1. Compile tk to WebAssembly (`tk.wasm`)
2. Copy the Go WASM runtime (`wasm_exec.js`)

### Run the demo

```bash
./serve.sh
```

Then open http://localhost:8080 in your browser.

### Try some commands

Once the page loads, try these commands:

```bash
# Initialize the database
init

# Create a project
project create demo "Demo Project"

# Create some tasks
new "Setup development environment" --project demo
new "Write documentation" --project demo
new "Add tests" --project demo

# List all tasks
ls

# Mark a task as in progress
mark demo-1 wip

# Mark a task as done
mark demo-2 done

# Show task details
show demo-1

# Get help
--help
```

## How it works

The demo uses:
- **Go WASM**: tk is compiled to WebAssembly using Go's WASM target
- **modernc.org/sqlite**: Pure Go SQLite implementation (no CGO required)
- **In-memory database**: The database is stored in the browser's memory
- **JavaScript interop**: Go functions are exposed to JavaScript for command execution

## Limitations

- **No persistence**: Data is lost when you refresh the page
- **No file operations**: Commands that work with files won't work
- **No sync**: Remote sync features are not available
- **Limited I/O**: Some terminal features may not work as expected

## Browser Compatibility

Works with modern browsers that support WebAssembly:
- Chrome 57+
- Firefox 52+
- Safari 11+
- Edge 16+

## Development

The main WASM wrapper is in `main.go`. It:
1. Captures stdout/stderr from tk commands
2. Exposes JavaScript functions for command execution
3. Manages the command lifecycle

The HTML interface (`index.html`) provides:
- Terminal-like UI
- Command input and history
- Output display

## Troubleshooting

**WASM module fails to load**: Make sure you built with `./build.sh` and are serving via HTTP (not file://)

**Commands don't work**: Check the browser console for errors. Some commands may not be compatible with the WASM environment.

**Database errors**: The database is in-memory only. Run `init` first to initialize it.
