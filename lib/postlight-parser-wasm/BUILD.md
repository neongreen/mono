# Building the Postlight Parser WASM Module

This document explains how to build the WebAssembly module for Postlight Parser.

## Overview

This library provides Go bindings to the [Postlight Parser](https://github.com/postlight/parser) JavaScript library via WebAssembly using [wazero](https://wazero.io/). To work properly, it requires a WASM-compiled version of Postlight Parser.

## Current Status

**The WASM bundle is not yet implemented.** This is a non-trivial task that requires:

1. Bundling the Postlight Parser JavaScript library with all dependencies
2. Creating a wrapper that exposes the parse function
3. Compiling to WebAssembly using a JavaScript-to-WASM compiler
4. Integrating with the Go code via wazero

## Prerequisites

- Node.js 18+ and npm/yarn
- [javy](https://github.com/bytecodealliance/javy) - JavaScript to WebAssembly compiler (recommended)
  OR
- [wizer](https://github.com/bytecodealliance/wizer) - WebAssembly pre-initializer
- wasm-opt (from [binaryen](https://github.com/WebAssembly/binaryen)) for optimization

## Implementation Options

### Option 1: Using Javy (Recommended)

[Javy](https://github.com/bytecodealliance/javy) compiles JavaScript to WebAssembly using QuickJS.

```bash
# Install javy
cargo install javy-cli

# Bundle Postlight Parser with dependencies
cd /tmp
npm install @postlight/parser
# Create wrapper script that uses the library

# Compile to WASM
javy compile parser_wrapper.js -o parser.wasm
```

### Option 2: Using a JavaScript Engine WASM Build

Alternatively, use a pre-built JavaScript engine (like QuickJS) compiled to WASM and load the Postlight Parser code into it.

### Option 3: Direct Integration (Most Complex)

1. Build Postlight Parser as a standalone bundle
2. Use Emscripten or similar to compile Node.js/V8 to WASM
3. Load and execute the bundle

## Steps to Complete Implementation

### 1. Create the JavaScript Wrapper

Create a wrapper that:
- Imports Postlight Parser
- Exposes a `parse(url, html)` function
- Returns JSON results
- Handles errors appropriately

Example structure:
```javascript
const Parser = require('@postlight/parser');

async function parse(inputJSON) {
    const { url, html } = JSON.parse(inputJSON);
    const result = await Parser.parse(url, { html: html });
    return JSON.stringify(result);
}

// Export for WASM
exports.parse = parse;
```

### 2. Bundle Dependencies

Use a bundler like webpack or rollup to create a single JavaScript file with all dependencies:

```bash
npm install @postlight/parser
npx webpack --entry ./parser_wrapper.js --output parser_bundle.js --target node
```

### 3. Compile to WASM

```bash
javy compile parser_bundle.js -o parser.wasm
```

### 4. Integrate with Go

Update `parser.go` to:
- Load the WASM module
- Call the parse function with input JSON
- Parse the JSON result into the Article struct

## Challenges

1. **Size**: The bundled WASM will be large (~2-5MB) due to Postlight Parser and dependencies
2. **Async**: Postlight Parser is async, but WASM execution is synchronous
3. **HTTP**: Postlight Parser expects to fetch URLs, but in WASM it needs pre-fetched HTML
4. **Dependencies**: Postlight Parser has many dependencies that need to be bundled

## Alternative Approach

If WASM compilation proves too complex, consider:

1. Creating a separate Node.js microservice that runs Postlight Parser
2. Having the Go library communicate with it via HTTP/gRPC
3. Or: using a simpler HTML parser library that's easier to compile to WASM

## Testing

Once the WASM module is built:

```bash
cd lib/postlight-parser-wasm
go test -v
```

## References

- [Postlight Parser](https://github.com/postlight/parser)
- [Javy - JavaScript to WebAssembly](https://github.com/bytecodealliance/javy)
- [wazero - WebAssembly runtime for Go](https://wazero.io/)
- [Building JavaScript for WebAssembly](https://bytecodealliance.org/articles/making-javascript-run-fast-on-webassembly)
