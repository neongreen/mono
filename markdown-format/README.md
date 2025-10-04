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

## Integration with Formatting Tools

See the [examples/](examples/) directory for complete configuration files and sample markdown files demonstrating the integrations.

### treefmt

[treefmt](https://github.com/numtide/treefmt) is a universal code formatter that runs multiple formatters with one command.

To integrate markdown-format with treefmt, you'll need to create a wrapper script since markdown-format outputs to stdout rather than formatting files in place.

1. Create a wrapper script (e.g., `markdown-format-inplace.sh`):

```bash
#!/bin/bash
# Wrapper script to make markdown-format work in-place for treefmt

MARKDOWN_FORMAT="./markdown-format/markdown-format"

for file in "$@"; do
    if [ -f "$file" ]; then
        temp_file=$(mktemp)
        "$MARKDOWN_FORMAT" "$file" > "$temp_file"
        mv "$temp_file" "$file"
    fi
done
```

2. Make the wrapper script executable:

```bash
chmod +x markdown-format-inplace.sh
```

3. Add to your `treefmt.toml`:

```toml
[formatter.markdown-format]
command = "./markdown-format-inplace.sh"
options = []
includes = ["*.md"]
```

4. Run treefmt:

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
