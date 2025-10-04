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

### With Go

```bash
go build
```

### With mise

Install using [mise](https://mise.jdx.dev/) with the Go backend:

```bash
mise use -g go:github.com/neongreen/mono/markdown-format@main
```

Or add to your `.mise.toml`:

```toml
[tools]
"go:github.com/neongreen/mono/markdown-format" = "main"
```

## Usage

Format a markdown file to stdout:

```bash
./markdown-format input.md
```

Format a file in place:

```bash
./markdown-format -w input.md
```

Format multiple files in place:

```bash
./markdown-format -w file1.md file2.md file3.md
```

Read from stdin:

```bash
cat input.md | ./markdown-format -
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

## Integration with Formatting Tools

See the [examples/](examples/) directory for complete configuration files and sample markdown files demonstrating the integrations.

### treefmt

[treefmt](https://github.com/numtide/treefmt) is a universal code formatter that runs multiple formatters with one command.

Add to your `treefmt.toml`:

```toml
[formatter.markdown-format]
command = "markdown-format"
options = ["-w"]
includes = ["*.md"]
```

Then run:

```bash
treefmt
```

### dprint

[dprint](https://dprint.dev/) is a pluggable and configurable code formatter.

Since markdown-format uses a custom sentence-per-line format, you'll need to use dprint's process plugin or create a custom wrapper. Here's how to integrate using a process-based approach:

1. Create a wrapper script that handles dprint's stdin/stdout protocol (e.g., `dprint-markdown-format.sh`):

```bash
#!/bin/bash
# Wrapper script for dprint integration

# Read from stdin, format, and output to stdout
/path/to/markdown-format -
```

2. Add to your `dprint.json`:

```json
{
  "plugins": [
    "https://plugins.dprint.dev/exec-0.5.0.json@checksum"
  ],
  "exec": {
    "commands": [{
      "command": "./dprint-markdown-format.sh",
      "exts": ["md"]
    }]
  }
}
```

Note: The exact configuration may vary depending on your dprint version. Refer to the [dprint documentation](https://dprint.dev/plugins/) for the latest plugin configuration format.

3. Run dprint:

```bash
dprint fmt
```

## License

MIT License - See LICENSE file in the repository root.
