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

## Project Structure

```
lion/
├── main.go                    # CLI entry point
├── internal/
│   ├── extractor/            # AST parsing and comment extraction
│   │   └── extractor.go
│   └── generator/            # Markdown file generation
│       └── generator.go
├── README.md
├── AGENTS.md
└── go.mod
```

## Comment Format

The tool looks for comments starting with `//lion:topic-name`:

```go
//lion:topic-name Optional documentation content
func SomeFunction() {
    // ...
}
```

## Implementation Notes

- Uses Go's `go/ast` package for parsing
- Extracts comments attached to functions, types, constants, and variables
- Skips test files (`*_test.go`)
- Generates one markdown file per topic
- Sorts entries by file and line number for consistency
- Includes source references in generated markdown

## Testing Approach

Tests should cover:
- Comment parsing (various formats and edge cases)
- AST traversal (different Go declarations)
- Markdown generation (formatting and structure)
- File I/O operations

## Future Enhancements

Potential features (not yet implemented):
- Multi-line comment support
- Custom templates for generated markdown
- Cross-references between topics
- Filtering by package or file patterns
- HTML output format
