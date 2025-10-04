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

## Current Test Cases (33 total)

### Basic Formatting
- `simple-paragraph` - Basic paragraph with multiple sentences
- `heading` - Heading formatting
- `heading-followed-by-list` - Heading followed by list
- `emphasis` - Bold and italic text
- `link` - Link formatting
- `inline-code` - Inline code formatting
- `fenced-code-block` - Code blocks

### Lists
- `unordered-list` - Unordered list formatting
- `ordered-list` - Ordered list formatting
- `multiple-sentences-in-list-item` - List items with multiple sentences
- `nested-unordered-lists` - Nested list structures
- `ordered-list-with-more-than-10-items` - Testing double-digit list numbers
- `ordered-list-starting-from-different-number` - Testing lists starting from non-1 numbers
- `unordered-list-with-many-items` - Testing long unordered lists (15+ items)
- `ordered-list-with-nested-unordered-list` - Mixed ordered and unordered lists
- `list-with-code-block-inside` - Code blocks inside list items
- `mixed-inline-elements-in-list` - Lists with bold, italic, code, and links

### List Marker Preservation
- `preserve-dash-marker` - Preserving `-` list markers
- `preserve-asterisk-marker` - Preserving `*` list markers
- `preserve-plus-marker` - Preserving `+` list markers
- `preserve-ordered-list-with-dot` - Preserving `.` in ordered lists
- `preserve-ordered-list-with-paren` - Preserving `)` in ordered lists
- `mixed-markers-create-separate-lists` - Different markers create separate lists
- `list-with-multiple-sentences-preserves-marker` - Marker preservation with multiple sentences

### Blockquotes
- `blockquote` - Basic blockquote formatting
- `blockquote-with-multiple-paragraphs` - Multiple paragraphs in blockquotes
- `nested-blockquotes` - Nested blockquotes
- `deeply-nested-blockquotes` - Deeply nested (3+ levels) blockquotes
- `list-inside-blockquote` - Lists inside blockquotes
- `code-block-in-blockquote` - Code blocks inside blockquotes

### Special Cases
- `abbreviation-in-sentence` - Handling of abbreviations (e.g., i.e., etc.)

### Complex Documents
- `complex-document-structure` - Comprehensive test with various elements combined
- `multiple-blocks-in-sequence` - Multiple different block types in sequence
