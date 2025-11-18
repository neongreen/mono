# Agent Guidelines for Dissect

## Build, Test, and Run Commands

**All commands must be run from the mono repository root.**

```bash
# Build
go build ./dissect

# Test
go test ./dissect/...

# Run
go run ./dissect [args...]

# Install (builds and places in $GOPATH/bin)
go install ./dissect
```

**Important:** Use `go` commands directly. Do not use `mise` for building or running dissect.

## Code Manipulation Requirements

### AST-Based Manipulation Only

**All Go code manipulation in the dissect tool MUST use AST (Abstract Syntax Tree) operations.**

### Prohibited Approaches

❌ **String manipulation of Go code is strictly prohibited**

This includes but is not limited to:
- String slicing to extract code snippets
- String concatenation to build Go code
- Text-based search and replace on Go source
- Regex operations on Go source code
- Manual text parsing of Go syntax

### Required Approaches

✅ **AST-based operations are required**

All code manipulation must use the Go AST packages:
- `go/ast` - Abstract syntax tree representation
- `go/parser` - Parsing Go source to AST
- `go/token` - Position information and file sets
- `go/format` - Formatting AST back to source
- `go/printer` - Printing AST nodes

### Rationale

String-based manipulation of Go code is fragile and error-prone because it:

1. **Cannot handle syntax variations**: Go allows multiple valid formatting styles. String matching fails when code is formatted differently.

2. **Breaks with whitespace changes**: String slicing by position breaks when whitespace, comments, or formatting changes.

3. **Cannot preserve semantic information**: Type information, scope, and other semantic data are lost.

4. **Fails on edge cases**: Build tags, compiler directives, and other Go features require proper parsing.

5. **Cannot properly handle comments**: Comments in Go have complex relationships with declarations that only AST preserves correctly.

### Examples

❌ **Prohibited** - String slicing:
```go
content, _ := os.ReadFile(file)
funcText := string(content[start:end])  // String manipulation!
```

✅ **Required** - AST manipulation:
```go
fset := token.NewFileSet()
node, _ := parser.ParseFile(fset, file, nil, parser.ParseComments)
// Work with node.Decls, node.Comments, etc.
format.Node(output, fset, node)
```

### When to Use gopls

The `gopls` language server can be used for complex refactoring operations, but:
- The results must still be processed using AST operations
- Do not extract string snippets from gopls output
- Parse gopls results with `go/parser` and manipulate the resulting AST

### Exception

The only exception is when dealing with non-Go files or when the output is for display purposes only (logging, error messages, etc.).

-------------------------------------------------

## Postmortems

### Postmortem: Double-Star Glob Pattern Bug (2025-10-12)

**Timeline:**

1. **Initial Implementation (Commit 588cf0d)**: Added glob support for file paths using `filepath.Glob()`. Documentation and examples claimed support for `pkg/**/*.go` pattern for recursive directory matching.

2. **Review Finding**: Code reviewer discovered that `filepath.Glob()` does NOT support `**` for recursive directory traversal. The `*` wildcard in Go's `filepath.Glob()` only matches non-separator characters, meaning it cannot cross directory boundaries. The pattern `pkg/**/*.go` would only match immediate subdirectories, not nested ones.

3. **Missing Tests**: No tests were written to verify `**` pattern behavior. The initial test suite only tested single-level glob patterns like `*.go`, which worked correctly with `filepath.Glob()`.

4. **Fix (Commit [current])**: Replaced `filepath.Glob()` with `doublestar.FilepathGlob()` from `github.com/bmatcuk/doublestar/v4`, which properly implements `**` for recursive matching. Added comprehensive test `MoveWithDoubleStarPattern` to verify recursive matching across multiple directory levels.

**Root Cause:**

- Assumption that `filepath.Glob()` supports `**` pattern without verifying in Go documentation
- Documentation was written based on desired behavior rather than tested behavior
- No tests were created to verify the documented functionality

**What Could Have Caught This Earlier:**

1. **Read the documentation**: Check `go doc filepath.Glob` and `go doc filepath.Match` before claiming support for a pattern
2. **Test what you document**: Every example in documentation should have a corresponding test
3. **Manual verification**: Manually test complex patterns (like `**`) before documenting them
4. **Integration tests**: Create test directories with nested structures to verify recursive behavior

**Lessons Learned:**

- Always verify library behavior against documentation before making claims
- Write tests for edge cases and complex patterns, especially when documenting them as supported
- Don't assume glob patterns work the same across all implementations
- Test the actual use case (nested directories) not just simplified scenarios

