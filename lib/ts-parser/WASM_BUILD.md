# Building the Tree-Sitter WASM Module

This document explains how the WASM file was produced for the ts-parser library.

**Note**: This documentation is for maintainers only. End users do not need to build the WASM file - it is pre-built and embedded in the library.

## Overview

The library uses a single WASM file:
- `internal/wasm/parser.wasm` - Tree-sitter runtime + TypeScript + TSX grammars

This file contains the tree-sitter parser runtime along with both language grammars, compiled to WASI-compatible WebAssembly using Zig.

## Source Repositories and Pinned Versions

| Component | Repository | Version/Commit |
|-----------|-----------|----------------|
| tree-sitter | https://github.com/tree-sitter/tree-sitter | v0.24.7 |
| tree-sitter-typescript | https://github.com/tree-sitter/tree-sitter-typescript | v0.23.2 |
| Zig compiler | https://ziglang.org | 0.13.0 |

## Prerequisites

### 1. Zig Compiler

Install Zig 0.13.0:

```bash
# Download and extract
curl -sL https://ziglang.org/download/0.13.0/zig-linux-x86_64-0.13.0.tar.xz | tar -xJ

# Add to PATH
export PATH="$PWD/zig-linux-x86_64-0.13.0:$PATH"

# Verify
zig version
# Should output: 0.13.0
```

### 2. Clone Source Repositories

```bash
# Create build directory
mkdir ts-wasm-build && cd ts-wasm-build

# Clone tree-sitter runtime
git clone --depth 1 --branch v0.24.7 https://github.com/tree-sitter/tree-sitter.git ts-runtime

# Clone tree-sitter-typescript grammar
git clone --depth 1 --branch v0.23.2 https://github.com/tree-sitter/tree-sitter-typescript.git ts-typescript
```

## Build Command

Build a single WASM file containing both TypeScript and TSX grammars:

```bash
zig cc --target=wasm32-wasi-musl -mexec-model=reactor \
    -I ts-runtime/lib/include \
    -I ts-runtime/lib/src \
    -I ts-typescript/typescript/src \
    -I ts-typescript/tsx/src \
    -I ts-typescript/common \
    -I ts-typescript \
    ts-runtime/lib/src/lib.c \
    ts-typescript/typescript/src/parser.c \
    ts-typescript/typescript/src/scanner.c \
    ts-typescript/tsx/src/parser.c \
    ts-typescript/tsx/src/scanner.c \
    -o parser.wasm \
    -Oz -fPIC \
    -Wl,--no-entry \
    -Wl,-z -Wl,stack-size=65536 \
    -Wl,--strip-debug \
    -Wl,--export=malloc \
    -Wl,--export=free \
    -Wl,--export=strlen \
    -Wl,--export=ts_parser_new \
    -Wl,--export=ts_parser_parse_string \
    -Wl,--export=ts_parser_set_language \
    -Wl,--export=ts_parser_delete \
    -Wl,--export=ts_language_version \
    -Wl,--export=ts_tree_root_node \
    -Wl,--export=ts_tree_delete \
    -Wl,--export=ts_node_string \
    -Wl,--export=ts_node_child_count \
    -Wl,--export=ts_node_named_child_count \
    -Wl,--export=ts_node_child \
    -Wl,--export=ts_node_named_child \
    -Wl,--export=ts_node_type \
    -Wl,--export=ts_node_start_byte \
    -Wl,--export=ts_node_end_byte \
    -Wl,--export=ts_node_is_error \
    -Wl,--export=ts_node_is_null \
    -Wl,--export=ts_node_parent \
    -Wl,--export=ts_node_next_sibling \
    -Wl,--export=ts_node_prev_sibling \
    -Wl,--export=ts_node_next_named_sibling \
    -Wl,--export=ts_node_prev_named_sibling \
    -Wl,--export=tree_sitter_typescript \
    -Wl,--export=tree_sitter_tsx
```

## Installing WASM File

Copy the built file to this repository:

```bash
cp parser.wasm /path/to/mono/lib/ts-parser/internal/wasm/parser.wasm
```

## Verification

After installing the WASM file, verify the library works:

```bash
CGO_ENABLED=0 go test ./...
```

All tests must pass with CGO disabled.

## Notes on the Build Process

The WASM file is built using Zig's cross-compilation to WASI (WebAssembly System Interface):

- **Target**: `wasm32-wasi-musl` - 32-bit WebAssembly with WASI for system calls and musl libc
- **Execution model**: `reactor` - The module exports functions rather than having a main entry point
- **Optimization**: `-Oz` for size optimization
- **Stack size**: 65536 bytes (64KB)
- **Exports**: Tree-sitter API functions plus both `tree_sitter_typescript` and `tree_sitter_tsx` language functions

The pre-built WASM files from tree-sitter-typescript releases are Emscripten-compiled and expect a JavaScript runtime, so they cannot be used with wazero. The Zig-compiled WASI version works directly with wazero.
