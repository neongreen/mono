# Testing Guide

This document explains the testing infrastructure for the markdown-format project.

## Overview

The project uses a dual testing approach:
1. **Inline tests** - Traditional Go table-driven tests embedded in `main_test.go`
2. **File-based tests** - Separate markdown files in the `testdata/` directory

## File-Based Testing

### Directory Structure

```
testdata/
├── README.md                          # Documentation
├── EXAMPLE_NEW_TEST.md                # How to add new tests
├── simple-paragraph.input.md          # Test input
├── simple-paragraph.output.md         # Expected output
├── complex-document-structure.input.md
├── complex-document-structure.output.md
└── ... (more test pairs)
```

### Test Case Naming Convention

- Input files: `<test-name>.input.md`
- Output files: `<test-name>.output.md`
- Test name is derived from the filename (e.g., `simple-paragraph`)

### Adding a New Test

1. Create an input file:
   ```bash
   cat > testdata/my-test.input.md << 'EOF'
   # My Test
   
   This is input. With multiple sentences.
   EOF
   ```

2. Create the expected output file:
   ```bash
   cat > testdata/my-test.output.md << 'EOF'
   # My Test
   
   This is input.
   With multiple sentences.
   EOF
   ```

3. Run the test:
   ```bash
   go test -v -run TestFormatMarkdownFromFiles/my-test
   ```

That's it! No code changes needed.

## Test Output

### Passing Test
```
=== RUN   TestFormatMarkdownFromFiles/my-test
--- PASS: TestFormatMarkdownFromFiles/my-test (0.00s)
```

### Failing Test with Diff
```
=== RUN   TestFormatMarkdownFromFiles/my-test
    main_test.go:733: formatMarkdown() mismatch (-want +got):
          string(
        - 	"Expected output here.\n",
        + 	"Actual output here.\n",
          )
--- FAIL: TestFormatMarkdownFromFiles/my-test (0.00s)
```

The diff clearly shows:
- Lines with `-` prefix: Expected output (what was in `.output.md`)
- Lines with `+` prefix: Actual output (what the formatter produced)

## Running Tests

### Run all tests
```bash
go test -v
```

### Run only file-based tests
```bash
go test -v -run TestFormatMarkdownFromFiles
```

### Run a specific test
```bash
go test -v -run TestFormatMarkdownFromFiles/simple-paragraph
```

### Run with race detection
```bash
go test -race -v
```

### Run with coverage
```bash
go test -cover
```

## Test Categories

The `testdata/` directory contains tests for:

### Basic Formatting
- Paragraphs with sentence splitting
- Headings (H1-H6)
- Emphasis (bold, italic)
- Links
- Inline code
- Code blocks

### Lists
- Ordered lists
- Unordered lists
- Nested lists
- Lists with multiple sentences per item
- Double-digit list numbering (10+)
- Mixed ordered and unordered lists

### List Marker Preservation
- Dash markers (`-`)
- Asterisk markers (`*`)
- Plus markers (`+`)
- Ordered list delimiters (`.` and `)`)

### Blockquotes
- Simple blockquotes
- Multiple paragraphs in blockquotes
- Nested blockquotes
- Deeply nested blockquotes (3+ levels)
- Lists inside blockquotes
- Code blocks inside blockquotes

### Special Cases
- Abbreviation handling (e.g., i.e., etc.)
- Mixed inline elements (bold + italic + code + links)

### Complex Documents
- Documents combining multiple features
- Real-world document structures

## Diff Library

The project uses [github.com/google/go-cmp](https://github.com/google/go-cmp) for generating diffs. This is a widely-used, well-maintained library from Google that provides:

- Clear, readable diffs
- Color-coded output (when supported)
- Configurable comparison options
- Industry-standard diff format

## Best Practices

1. **One test per feature** - Each test should focus on a specific markdown feature
2. **Clear test names** - Use descriptive names like `nested-blockquotes` or `preserve-asterisk-marker`
3. **Keep tests simple** - Test files should be as minimal as possible while still testing the feature
4. **Add edge cases** - Consider boundary conditions (e.g., double-digit list numbers)
5. **Test real-world scenarios** - Include complex documents that combine multiple features

## Backward Compatibility

The existing inline tests in `main_test.go` are maintained for backward compatibility:
- `TestFormatMarkdown` - Basic formatting tests
- `TestSplitIntoSentences` - Sentence splitting logic
- `TestNestedElements` - Nested structure handling
- `TestLargeOrderedList` - List numbering
- `TestComplexNestedStructures` - Complex document structures
- `TestListMarkerPreservation` - Marker preservation logic

These tests ensure that refactoring doesn't break existing functionality.

## Statistics

- **26 file-based test cases** (52 files: input + output pairs)
- **73 total test assertions** (including inline tests)
- **100% pass rate** with race detection enabled
- **Comprehensive coverage** of all markdown features

## Troubleshooting

### Test fails but output looks correct
Check for:
- Trailing whitespace differences
- Line ending differences (LF vs CRLF)
- Extra blank lines at the end

### Test not discovered
Ensure:
- Files follow naming convention: `*.input.md` and `*.output.md`
- Both input and output files exist
- Files are in the `testdata/` directory

### Diff is hard to read
The go-cmp diff shows exact differences including:
- `\n` for newlines
- `\t` for tabs
- Whitespace is visible

## Contributing

When adding new features to the formatter:
1. Add test cases to `testdata/` first (TDD approach)
2. Run tests to see them fail
3. Implement the feature
4. Run tests to see them pass
5. Add edge case tests as needed

## Additional Resources

- [go-cmp documentation](https://pkg.go.dev/github.com/google/go-cmp/cmp)
- [Go testing documentation](https://pkg.go.dev/testing)
- [testdata/ README](testdata/README.md) - More details on test structure
- [testdata/EXAMPLE_NEW_TEST.md](testdata/EXAMPLE_NEW_TEST.md) - Step-by-step example
