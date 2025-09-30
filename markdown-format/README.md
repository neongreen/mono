# markdown-format

A Go-based markdown formatter that reformats markdown files with one sentence per line.

## Features

- Uses [goldmark](https://github.com/yuin/goldmark), a well-maintained CommonMark-compliant markdown parser
- Formats paragraphs with one sentence per line
- Preserves markdown structure:
  - Headers
  - Lists (both ordered and unordered)
  - Code blocks (fenced and indented)
  - Blockquotes
  - Inline formatting (bold, italic, links, images, inline code)
  - Horizontal rules

## Why one sentence per line?

Formatting markdown with one sentence per line makes it easier to:
- Track changes in version control (git diffs are clearer)
- Review and edit individual sentences
- Collaborate on documents

## Installation

```bash
go build
```

## Usage

Format a markdown file:

```bash
./markdown-format input.md > output.md
```

Read from stdin:

```bash
cat input.md | ./markdown-format - > output.md
```

## Example

Input:
```markdown
# Hello

This is a paragraph. It has multiple sentences. Let's format it!
```

Output:
```markdown
# Hello

This is a paragraph.
It has multiple sentences.
Let's format it!
```

## License

MIT License - See LICENSE file in the repository root.
