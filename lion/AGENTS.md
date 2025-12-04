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

# Build TypeScript helper (required for TypeScript support)
cd lion/ts-helper
npm install
npm run build
```

## Project Structure

```
lion/
├── main.go                    # CLI entry point
├── internal/
│   ├── extractor/            # AST parsing and comment extraction
│   │   ├── extractor.go      # Go extraction using go/ast
│   │   └── typescript.go     # TypeScript extraction via Node.js helper
│   └── generator/            # Markdown file generation
│       └── generator.go
├── ts-helper/                 # TypeScript AST parser helper
│   ├── src/index.ts          # TypeScript extractor using TS compiler API
│   ├── package.json
│   └── tsconfig.json
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
- Skips test files (`*_test.go`)

### TypeScript Extraction
- Uses TypeScript compiler API via a Node.js helper script
- Parses JSDoc comments for `@lion` tags
- Skips test files (`*.test.ts`, `*.spec.ts`, `*.d.ts`)
- Supports functions, classes, interfaces, types, enums, and variables

### Common
- Generates one markdown file per topic
- Sorts entries by file and line number for consistency
- Includes source references in generated markdown

## Testing Approach

Tests should cover:
- Comment parsing (various formats and edge cases)
- AST traversal (different Go and TypeScript declarations)
- Markdown generation (formatting and structure)
- File I/O operations
- Mixed Go and TypeScript projects

## Future Enhancements

Potential features (not yet implemented):
- Custom templates for generated markdown
- Cross-references between topics
- Filtering by package or file patterns
- HTML output format
- JavaScript support (in addition to TypeScript)
