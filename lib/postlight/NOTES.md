# Development Notes

## Promise Support Investigation

### Finding: Javy DOES support async/await with event loop!

Tested on 2025-11-09 and confirmed that:
- Javy v7.0.1 supports promises and async/await when built with `-J event-loop=y` flag
- Test confirmed: async functions work correctly in WASM
- This opens the door to using asynchronous JavaScript libraries

### Building with Event Loop

To build WASM with event loop support:

```bash
javy build -J event-loop=y input.js -o output.wasm
```

### Full Postlight Parser Status

Attempted to bundle the full Postlight Parser but encountered **130+ errors** due to Node.js dependencies:

**Missing modules in Javy:**
- `http`, `https` - Network requests
- `fs`, `fs/promises` - File system
- `crypto` - Cryptography
- `buffer`, `stream` - Data handling
- `tls`, `net` - Networking
- `url`, `util`, `events`, `assert` - Core Node.js

**Conclusion:** The full Postlight Parser is designed for Node.js and requires extensive built-in modules that don't exist in Javy's minimal QuickJS environment.

## Alternative Approaches

### 1. Browser-Compatible Parsers
Libraries designed for browsers (like Mozilla's Readability) might work better since they don't depend on Node.js APIs:
- `@mozilla/readability`
- `readability.js`
- `node-readability` (browser fork)

### 2. Custom Lightweight Parser
Build a custom parser using only:
- Basic DOM/HTML parsing (available in WASM)
- Simple content extraction rules
- No external dependencies

### 3. Keep Simplified Parser
Current implementation works well for basic use cases:
- ✅ Title extraction
- ✅ Content extraction
- ✅ Word counting
- ✅ Domain parsing
- ✅ No CGO required
- ✅ Pure Go + WASM

## Recommendations

For production use in a WASM/Javy environment, either:
1. Use a browser-compatible article extraction library
2. Implement custom parsing logic optimized for the environment
3. Accept the simplified parser's limitations for the benefits of no CGO

The current simplified implementation demonstrates the infrastructure works perfectly and is suitable for many use cases.
