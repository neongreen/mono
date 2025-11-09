#!/bin/bash
# Script to check if all build dependencies are installed

set -e

echo "Checking build dependencies for Postlight Parser WASM module..."
echo

# Check Node.js
if command -v node >/dev/null 2>&1; then
    NODE_VERSION=$(node --version)
    echo "✓ Node.js installed: $NODE_VERSION"
else
    echo "✗ Node.js not found"
    echo "  Install from: https://nodejs.org/"
    exit 1
fi

# Check npm
if command -v npm >/dev/null 2>&1; then
    NPM_VERSION=$(npm --version)
    echo "✓ npm installed: $NPM_VERSION"
else
    echo "✗ npm not found"
    echo "  Usually comes with Node.js"
    exit 1
fi

# Check cargo
if command -v cargo >/dev/null 2>&1; then
    CARGO_VERSION=$(cargo --version)
    echo "✓ cargo installed: $CARGO_VERSION"
else
    echo "✗ cargo not found"
    echo "  Install from: https://rustup.rs/"
    exit 1
fi

# Check javy
if command -v javy >/dev/null 2>&1; then
    JAVY_VERSION=$(javy --version)
    echo "✓ javy installed: $JAVY_VERSION"
else
    echo "✗ javy not found"
    echo "  Install with: cargo install javy-cli"
    exit 1
fi

echo
echo "All dependencies are installed!"
echo
echo "Next steps:"
echo "  1. cd lib/readability-wasm"
echo "  2. make build-wasm"
echo "  3. go test -v"
