# Postlight Parser WASM Library

Go bindings to [Postlight Parser](https://github.com/postlight/parser) using WebAssembly via [wazero](https://wazero.io/).

## Status

**⚠️ IMPORTANT: This library requires a WASM bundle that is not yet built.**

The library provides the API and infrastructure for executing Postlight Parser via WebAssembly, but the actual WASM compilation of Postlight Parser needs to be completed. See [BUILD.md](BUILD.md) for details.

## What is Postlight Parser?

[Postlight Parser](https://github.com/postlight/parser) extracts clean article content from web pages.

## Current Implementation

- ✅ Go API defined with proper types matching Postlight Parser output
- ✅ wazero runtime integration
- ✅ WASI support
- ❌ **WASM bundle** (stub only - see BUILD.md)

See BUILD.md for how to complete the implementation.
