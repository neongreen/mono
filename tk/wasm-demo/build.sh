#!/bin/bash
set -e

echo "Building tk for WebAssembly..."

# Get the directory where this script is located
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Build WASM from monorepo root
cd "$DIR/.."
GOOS=js GOARCH=wasm go build -o wasm-demo/tk.wasm ./wasm-demo/main.go

# Copy wasm_exec.js from Go installation
GOROOT=$(go env GOROOT)
WASM_EXEC=$(find "$GOROOT" -name "wasm_exec.js" | head -1)
if [ -z "$WASM_EXEC" ]; then
    echo "Error: Could not find wasm_exec.js in Go installation"
    exit 1
fi
cp "$WASM_EXEC" "$DIR/"

echo "✓ Build complete!"
echo ""
echo "Files generated:"
echo "  - tk.wasm (WebAssembly binary - $(du -h "$DIR/tk.wasm" | cut -f1))"
echo "  - wasm_exec.js (Go WASM runtime)"
echo ""
echo "To run the demo:"
echo "  cd $DIR && ./serve.sh"
echo ""
echo "Or use Python:"
echo "  cd $DIR && python3 -m http.server 8080"
echo ""
echo "Then open: http://localhost:8080"
