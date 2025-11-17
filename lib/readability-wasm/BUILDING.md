# Building the Postlight Parser WASM Module

This document explains how to build the WASM module from the Postlight Parser JavaScript library.

## Prerequisites

### 1. Node.js and npm

Install Node.js (version 18 or later recommended):

```bash
# On Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# On macOS
brew install node

# Verify installation
node --version
npm --version
```

### 2. Rust and Cargo

Install Rust to get cargo (needed for javy):

```bash
# Install rustup
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Follow the prompts, then reload your shell
source $HOME/.cargo/env

# Verify installation
cargo --version
```

### 3. Javy

Install javy, the JavaScript to WASM compiler:

```bash
cargo install javy-cli

# Verify installation
javy --version
```

## Build Process

The build happens in stages:

### 1. Install Dependencies

```bash
cd lib/postlight
make install
```

This installs the npm packages:
- `@postlight/parser` - The actual parser
- `esbuild` - For bundling JavaScript

### 2. Bundle JavaScript

```bash
make build-js
```

This:
- Uses esbuild to bundle `js/index.js` with all dependencies
- Outputs `js/dist/bundle.js`
- The bundle is a single JavaScript file ready for WASM compilation

### 3. Compile to WASM

```bash
make build-wasm
```

This:
- Uses javy to compile `js/dist/bundle.js` to `parser.wasm`
- The WASM file is embedded in the Go binary via `//go:embed`

### Complete Build

To do all steps at once:

```bash
make all
```

Or simply:

```bash
make
```

## Rebuilding

After making changes to the JavaScript code:

```bash
make clean    # Remove old build artifacts
make          # Rebuild everything
```

## Troubleshooting

### "javy: command not found"

Make sure cargo's bin directory is in your PATH:

```bash
export PATH="$HOME/.cargo/bin:$PATH"
```

Add this to your `~/.bashrc` or `~/.zshrc` to make it permanent.

### Build errors with esbuild

Make sure you've run `npm install` in the `js` directory:

```bash
cd js
npm install
cd ..
```

### WASM module too large

The bundled WASM module can be large (several MB) because it includes:
- Postlight Parser
- All its dependencies
- The QuickJS JavaScript engine (from javy)

This is normal. The module is embedded once and cached.

## CI/CD Integration

For automated builds, ensure the CI environment has:

1. Node.js 18+
2. Rust/cargo
3. javy installed

Example GitHub Actions workflow:

```yaml
- name: Install dependencies
  run: |
    # Node.js is usually pre-installed
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    source $HOME/.cargo/env
    cargo install javy-cli

- name: Build WASM module
  run: |
    cd lib/postlight
    make build-wasm
```

## Development Tips

### Testing Changes

After modifying the JavaScript wrapper:

```bash
make build-wasm
go test -v
```

### Debugging

To see the bundled JavaScript output:

```bash
make build-js
cat js/dist/bundle.js
```

To test the WASM module directly with javy:

```bash
echo '{"url":"https://example.com","html":"<html>...</html>"}' | javy run js/dist/bundle.js
```

## File Structure

```
lib/postlight/
├── js/
│   ├── index.js          # JavaScript wrapper
│   ├── build.js          # esbuild configuration
│   ├── package.json      # npm dependencies
│   └── dist/
│       └── bundle.js     # Generated bundle
├── parser.wasm           # Generated WASM module
├── readability.go          # Go bindings
├── Makefile              # Build system
└── README.md             # User documentation
```
