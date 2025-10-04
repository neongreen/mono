# Test Data

This directory contains input and expected output file pairs for testing the markdown formatter.

## Structure

Each test case consists of two files:
- `<test-name>.input.md` - The input markdown file
- `<test-name>.output.md` - The expected formatted output

## Adding New Tests

To add a new test case:

1. Create a new `.input.md` file with the test input
2. Create a corresponding `.output.md` file with the expected formatted output
3. The test name is derived from the filename (without `.input.md` or `.output.md`)
4. Run `go test` to verify the new test passes

Example:
```bash
# Create test files
echo "Test input. With sentences." > testdata/my-test.input.md
echo "Test input.\nWith sentences.\n" > testdata/my-test.output.md

# Run tests
go test -v -run TestFormatMarkdownFromFiles/my-test
```

## Benefits

- **Easy to maintain**: Test cases are separate files, not embedded in Go code
- **Clear diffs**: The `go-cmp` library provides clear, readable diffs when tests fail
- **Version control friendly**: Input and output files can be easily reviewed in pull requests
- **Easy to add**: Just add two files, no code changes needed

## Current Test Cases

- `simple-paragraph` - Basic paragraph with multiple sentences
- `heading` - Heading formatting
- `unordered-list` - Unordered list formatting
- `ordered-list` - Ordered list formatting
- `fenced-code-block` - Code blocks
- `inline-code` - Inline code formatting
- `emphasis` - Bold and italic text
- `link` - Link formatting
- `blockquote` - Blockquote formatting
- `multiple-sentences-in-list-item` - List items with multiple sentences
- `abbreviation-in-sentence` - Handling of abbreviations (e.g., i.e., etc.)
- `nested-blockquotes` - Nested blockquotes
- `nested-unordered-lists` - Nested list structures
- `preserve-dash-marker` - Preserving `-` list markers
- `preserve-asterisk-marker` - Preserving `*` list markers
- `preserve-plus-marker` - Preserving `+` list markers
- `complex-document-structure` - Comprehensive test with various elements
