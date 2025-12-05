# Agent Guidelines for lion

## Building and Testing

```bash
# Build from repository root
go build ./lion

# Run tests
go test ./lion/...

# Run the tool
go run ./lion generate
go run ./lion topics

# Format code
goimports -w lion/
```

TypeScript support requires `tsc` and `node` to be available in PATH (no additional setup needed).

## Project Structure

```
lion/
├── main.go                    # CLI entry point
├── internal/
│   ├── extractor/            # AST parsing and comment extraction
│   │   ├── extractor.go      # Go extraction using go/ast
│   │   └── typescript.go     # TypeScript extraction (embedded, uses tsc at runtime)
│   └── generator/            # Markdown file generation
│       └── generator.go
├── README.md
├── AGENTS.md
└── docs/
```

## Comment Formats

### Go Comments

The tool looks for comments starting with `//lion:topic-name`:

```go
//lion:topic-name Optional documentation content
func SomeFunction() {
    // ...
}
```

### TypeScript Comments

TypeScript uses JSDoc-style comments with `@lion` tags:

```typescript
/**
 * @lion topic-name
 * Documentation content here.
 */
function someFunction() {
    // ...
}
```

## Implementation Notes

### Go Extraction
- Uses Go's `go/ast` package for parsing
- Extracts comments attached to functions, types, constants, and variables
- Supports inline comments inside function bodies
- Supports single-line (`//lion:`) and block (`/*lion: */`) formats
- Skips test files (`*_test.go`)

### TypeScript Extraction
- Uses an embedded TypeScript extractor script compiled with `tsc` at runtime
- Uses the TypeScript compiler API for AST parsing
- Parses JSDoc comments for `@lion` tags (multi-line format required for declarations)
- Supports inline single-line comments inside function bodies (`// @lion` or `//lion:`)
- Skips test files (`*.test.ts`, `*.spec.ts`, `*.d.ts`)
- Supports functions, classes, interfaces, types, enums, and variables

### Feature Parity

| Feature | Go | TypeScript |
|---------|----|----|
| Topic extraction | ✅ | ✅ |
| Metadata (title, section) | ✅ | ✅ |
| Entity name extraction | ✅ | ✅ |
| Multi-line content | ✅ | ✅ |
| Test file skipping | ✅ | ✅ |
| Inline function comments | ✅ | ✅ |
| Single-line declaration comments | ✅ | ❌ (requires JSDoc) |
| Package-level docs | ✅ | ❌ |

### Common
- Generates one markdown file per topic
- Sorts entries by file and line number for consistency
- Includes source references in generated markdown
- Unified output for mixed Go/TypeScript projects

## Testing Approach

Tests should cover:
- Comment parsing (various formats and edge cases)
- AST traversal (different Go and TypeScript declarations)
- Markdown generation (formatting and structure)
- File I/O operations
- Mixed Go and TypeScript projects
- Feature parity between Go and TypeScript extraction
- Inline comments inside function bodies

## Future Enhancements

Potential features (not yet implemented):
- Custom templates for generated markdown
- Cross-references between topics
- Filtering by package or file patterns
- HTML output format
- JavaScript support (in addition to TypeScript)
