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

Format files in place (default behavior):

```bash
./markdown-format file1.md file2.md file3.md
```

Check if files are formatted without modifying them:

```bash
./markdown-format --check file1.md file2.md
```

Format from stdin to stdout:

```bash
cat input.md | ./markdown-format
```

Or:

```bash
./markdown-format < input.md > output.md
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
includes = ["*.md"]
```

Then run:

```bash
treefmt
```

### dprint

[dprint](https://dprint.dev/) is a pluggable and configurable code formatter.

To integrate markdown-format with dprint, use the exec plugin. First, install the exec plugin if you haven't already:

```bash
dprint config add exec
```

Then add to your `dprint.json`:

```json
{
  "exec": {
    "commands": [{
      "command": "markdown-format",
      "exts": ["md"]
    }]
  },
  "plugins": [
    "https://plugins.dprint.dev/exec-0.5.0.json@<hash>"
  ]
}
```

The exec plugin will handle passing files to markdown-format for in-place formatting.

Run dprint:

```bash
dprint fmt
```

## License

MIT License - See LICENSE file in the repository root.
