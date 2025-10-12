# printpdf - Print to PDF Tool

A versatile tool for converting various content sources to PDF format using multiple PDF generation engines.

## Features

### Input Sources

- **Local Markdown files**: Direct file paths
- **HTTP/HTTPS URLs**: Remote Markdown files
- **GitHub files**: Direct links to files in GitHub repositories (including private repos with authentication)
  - Blob URLs: `https://github.com/owner/repo/blob/branch/file.md`
  - Raw URLs: `https://raw.githubusercontent.com/owner/repo/branch/file.md`
  - Commit URLs: `https://github.com/owner/repo/blob/commit-sha/file.md`
- **Web pages**: Regular web pages (processed with Mozilla Readability for clean content)

### PDF Converters

The tool supports multiple PDF conversion engines and will try all of them by default:

- **Typst**: Modern typesetting system (automatically downloaded if not present)
- **Prince XML**: Commercial HTML-to-PDF converter (requires manual installation)
- **WeasyPrint**: Python-based HTML-to-PDF converter (requires `pip install weasyprint`)

Each converter has different strengths:
- Typst: Best for documents with modern typography
- Prince: Professional output with excellent CSS support
- WeasyPrint: Open-source, good for simple documents

## Installation

### Build from source

```bash
cd printpdf
go build -o printpdf ./cmd
```

### Using mise

```bash
cd printpdf
mise run build
```

## Usage

### Basic usage

```bash
# Convert a local Markdown file
printpdf document.md

# Convert from a URL
printpdf https://example.com/document.md

# Convert a GitHub file (public repository)
printpdf https://github.com/owner/repo/blob/main/README.md

# Convert a GitHub file (private repository - requires GITHUB_TOKEN)
export GITHUB_TOKEN=your_token_here
printpdf https://github.com/owner/repo/blob/main/README.md

# Convert a web page (uses Readability)
printpdf https://example.com/article
```

### Options

```bash
# Specify output directory
printpdf -o ./output document.md

# Use specific converters (comma-separated)
printpdf -converters typst,weasyprint document.md

# Use only one converter
printpdf -converters typst document.md
```

### GitHub Authentication

For private repositories, set your GitHub token:

```bash
export GITHUB_TOKEN=ghp_your_token_here
```

Or use the GitHub CLI:

```bash
gh auth login
```

The tool will automatically use your GitHub token if available in the environment.

## Output

The tool generates PDF files in the specified output directory (current directory by default):

- `output-typst.pdf` - Generated with Typst
- `output-prince.pdf` - Generated with Prince XML
- `output-weasyprint.pdf` - Generated with WeasyPrint

This allows you to compare the output quality of different converters and choose the best one for your needs.

## Automatic Tool Download

The tool automatically manages PDF converter downloads:

- **Typst**: Downloaded automatically from GitHub releases
- **Prince**: Must be installed manually (commercial software)
- **WeasyPrint**: Must be installed via pip (Python required)

Downloaded tools are cached in `~/.cache/printpdf/` for future use.

## Examples

See the `samples/` directory for example PDF outputs generated from various sources.

## Requirements

- Go 1.24.7 or later (for building)
- Internet connection (for downloading converters and fetching remote content)

Optional:
- Python 3 with pip (for WeasyPrint)
- Prince XML license (for Prince converter)

## Development

### Run tests

```bash
mise run test
```

### Format code

```bash
mise run fmt
```

### Build

```bash
mise run build
```

## Architecture

The tool is organized into several packages:

- `cmd/` - Main entry point
- `pkg/fetcher/` - Content fetching from various sources
- `pkg/converter/` - PDF conversion engines
- `pkg/downloader/` - Tool download and caching

## License

See [LICENSE](../LICENSE) in the repository root.
