# tk WASM Demo - Dagger Functions

This directory includes Dagger functions for building and serving the tk WASM demo.

## Available Functions

### WasmBuild

Builds the tk WASM binary and returns a directory with the compiled assets.

**Usage:**
```bash
# Export the built assets to a local directory
dagger call project tk wasm-build export --path=./output

# The output directory will contain:
# - tk.wasm (WebAssembly binary ~25MB)
# - wasm_exec.js (Go WASM runtime)
# - index.html (Web interface)
```

### WasmServe

Builds the tk WASM demo and serves it on the specified port using Python's HTTP server.

**Usage:**
```bash
# Serve the demo on port 8080 (default)
dagger call project tk wasm-serve up --ports 8080:8080

# Serve on a custom port
dagger call project tk wasm-serve --port 3000 up --ports 3000:3000

# Once running, open http://localhost:8080 in your browser
```

## Implementation Details

The Dagger functions are defined in `.dagger/project_tk.go`:

- **WasmBuild**: Compiles tk to WebAssembly using `GOOS=js GOARCH=wasm`, extracts `wasm_exec.js` from the Go installation, and bundles everything with the HTML interface.

- **WasmServe**: Builds the assets using `WasmBuild`, then creates a Python container that serves the files on the specified port.

## Examples

### Quick Demo

```bash
# Build and serve in one command
cd .dagger
dagger call project tk wasm-serve up --ports 8080:8080

# In another terminal or browser, visit:
# http://localhost:8080
```

### Build Only

```bash
# Just build the WASM assets
cd .dagger
dagger call project tk wasm-build export --path=../tk/wasm-demo/dist

# Serve with your own HTTP server
cd ../tk/wasm-demo/dist
python3 -m http.server 8080
```

### Custom Port

```bash
# Serve on port 3000
cd .dagger
dagger call project tk wasm-serve --port 3000 up --ports 3000:3000
```

## Requirements

- Dagger CLI installed
- Docker (required by Dagger)
- Go 1.24.7+ (managed by Dagger container)

## Notes

- The WASM build uses an in-memory SQLite database (`:memory:`)
- The build process caches Go modules and build artifacts for faster subsequent builds
- The service runs in a Python container for simplicity
- Port forwarding is required to access the service from your local machine
