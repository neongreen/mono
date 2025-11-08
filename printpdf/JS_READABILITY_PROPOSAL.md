# Proposal: Multiple Readability Engine Support for printpdf

## Overview

Add support for actual JavaScript-based readability engines (Mozilla Readability, Defuddle, Postlight Parser, etc.) while keeping the custom Go implementation as a fallback.

## Why JavaScript Engines?

The best readability implementations are in JavaScript:
- **@mozilla/readability** - The original, battle-tested implementation
- **kepano/defuddle** - Modern fork with improvements
- **@postlight/parser** - Commercial-grade parser (Mercury)
- **@akira108sys/html-rewriter-readability** - Alternative implementation

Implementing these algorithms from scratch in Go is complex and won't match the quality of the original JS implementations.

## Implementation Options

### Option 1: External Node.js (Recommended)

**Approach:**
- Bundle JS libraries at build time with esbuild/webpack
- Embed bundles in Go binary using `//go:embed`
- Write bundles to temp directory at runtime
- Shell out to `node` command to run extraction
- Fall back to custom Go implementation if Node.js unavailable

**Pros:**
- Simple implementation
- Works with any npm package
- No CGO dependency
- Graceful degradation (falls back to Go)
- Node.js already in monorepo environment

**Cons:**
- Requires Node.js installed on user's machine
- Process spawn overhead (~50-100ms)
- Temp file I/O

**Code structure:**
```go
type ReadabilityEngine interface {
    Extract(html []byte) ([]byte, error)
    IsAvailable() bool
}

type NodeJSEngine struct {
    engineName string
    bundlePath string
}

type CustomGoEngine struct {
    // Current implementation
}
```

### Option 2: QuickJS with CGO

**Approach:**
- Use `github.com/buke/quickjs-go`
- Bundle JS libraries at build time
- Embed in Go binary
- Run in-process via QuickJS

**Pros:**
- Embedded (no external dependencies)
- Fast (in-process)
- Modern JS support (ES2020)

**Cons:**
- Requires CGO (complicates cross-compilation)
- Larger binary size
- Less flexibility for updates

### Option 3: Goja (Pure Go)

**Approach:**
- Use `github.com/dop251/goja`
- Bundle JS with ES5 target
- Run in-process

**Pros:**
- Pure Go (no CGO)
- Cross-platform
- Embedded

**Cons:**
- ES5.1 only (many modern packages won't work)
- Slower than V8/QuickJS
- May require transpiling npm packages

## Recommended Approach: Node.js with Fallback

### Architecture

```
printpdf/
├── cmd/
│   └── main.go                 # Add --readability-engine flag
├── pkg/
│   └── fetcher/
│       ├── fetcher.go          # Use engine registry
│       ├── readability.go      # Custom Go (fallback)
│       ├── readability_js.go   # JS engine runner
│       └── engines/
│           ├── registry.go     # Engine registry
│           ├── mozilla.go
│           ├── defuddle.go
│           ├── postlight.go
│           └── custom.go
└── readability-bundles/        # Built JS bundles (embedded)
    ├── mozilla.bundle.js
    ├── defuddle.bundle.js
    ├── postlight.bundle.js
    └── html-rewriter.bundle.js
```

### Build Process

1. **At Development Time:**
   ```bash
   # In readability-bundles/
   npm install @mozilla/readability jsdom
   # Create extract-mozilla.js wrapper
   # Bundle with esbuild
   ```

2. **At Build Time:**
   ```go
   //go:embed readability-bundles/*.bundle.js
   var readabilityBundles embed.FS
   ```

3. **At Runtime:**
   - Check if Node.js available
   - Extract appropriate bundle to temp file
   - Run: `node bundle.js < input.html > output.html`
   - Parse result
   - Clean up temp files

### CLI Interface

```bash
# Default (auto-detect best available)
printpdf https://example.com/article

# Specific engine
printpdf --readability-engine mozilla https://example.com/article
printpdf --readability-engine defuddle https://example.com/article
printpdf --readability-engine postlight https://example.com/article
printpdf --readability-engine custom https://example.com/article

# List available
printpdf --list-readability-engines
# Output:
#   mozilla (available via Node.js)
#   defuddle (available via Node.js)
#   postlight (available via Node.js)
#   custom (built-in Go implementation)

# Disable JS engines (force Go)
printpdf --readability-engine custom https://example.com/article
```

### Configuration File

```toml
[readability]
# Primary engine to use
engine = "mozilla"

# Fallback if primary unavailable or fails
fallback = "custom"

# Additional options
[readability.options]
# Engine-specific options can be added here
```

### Error Handling

```
1. Try primary engine (e.g., mozilla)
2. If unavailable/fails, try fallback (custom Go)
3. If all fail, return error with helpful message
```

### Engine Priority (Auto mode)

```
1. Mozilla Readability (if Node.js available)
2. Custom Go implementation (always available)
```

## Implementation Plan

### Phase 1: Infrastructure (1-2 hours)
- [ ] Create engine registry interface
- [ ] Refactor current Go implementation to use engine interface
- [ ] Add Node.js availability check
- [ ] Add temp file management

### Phase 2: Mozilla Readability (2-3 hours)
- [ ] Create readability-bundles/ directory
- [ ] Install @mozilla/readability + jsdom
- [ ] Create wrapper script (extract-mozilla.js)
- [ ] Bundle with esbuild
- [ ] Embed in Go binary
- [ ] Implement NodeJSEngine for mozilla
- [ ] Add tests

### Phase 3: Additional Engines (1-2 hours each)
- [ ] Add Defuddle support
- [ ] Add Postlight Parser support
- [ ] Add html-rewriter-readability support

### Phase 4: CLI & Config (1 hour)
- [ ] Add --readability-engine flag
- [ ] Add --list-readability-engines command
- [ ] Add configuration file support
- [ ] Update help text and README

### Phase 5: Testing & Documentation (2 hours)
- [ ] Add integration tests for each engine
- [ ] Add comparison tests (verify outputs are reasonable)
- [ ] Update documentation
- [ ] Add troubleshooting guide

**Total Estimate: 8-12 hours**

## Trade-offs

### Node.js Approach
✅ Best compatibility with JS ecosystem
✅ Simple to maintain
✅ Graceful degradation
❌ Requires Node.js on user machine
❌ Process spawn overhead

### QuickJS Approach
✅ Fully embedded
✅ No runtime dependencies
❌ CGO complexity
❌ Cross-compilation challenges
❌ Less flexible

### Recommendation
Start with Node.js approach because:
1. Node.js already used in monorepo
2. Simpler to implement and maintain
3. Better compatibility
4. Easy to test and debug
5. Can add QuickJS later if needed

## Success Criteria

1. Mozilla Readability works with Node.js
2. Falls back to custom Go implementation gracefully
3. CLI flag allows engine selection
4. All existing tests pass
5. New tests for each engine
6. Documentation updated
7. Performance acceptable (<200ms overhead)

## Open Questions

1. Should we cache extracted content?
2. Should we support custom JS engines (user-provided)?
3. Should we vendor the JS bundles or build them on-demand?
4. What's the minimum Node.js version we should support?
5. Should we add telemetry to track which engines are used?
