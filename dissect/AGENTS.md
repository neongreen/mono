# Agent Guidelines for Dissect

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
