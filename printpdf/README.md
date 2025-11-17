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
  - Commit file URLs: `https://github.com/owner/repo/files/commit-sha/path/to/file.md`
- **Web pages**: Regular web pages (processed with Mozilla Readability for clean content)

### PDF Converters

The tool supports multiple PDF conversion engines and will try all of them by default:

- **Typst**: Modern typesetting system with proper Markdown to Typst conversion (automatically downloaded if not present)
- **Prince XML**: Commercial HTML-to-PDF converter (requires manual installation)
- **WeasyPrint**: Python-based HTML-to-PDF converter (requires `pip install weasyprint`)

Each converter has different strengths:
- Typst: Best for documents with modern typography, excellent for academic papers and technical documents
- Prince: Professional output with excellent CSS support
- WeasyPrint: Open-source, good for documentation and simple layouts

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

# Create a multi-column layout (newspaper style)
printpdf -columns 2 document.md
printpdf -columns 3 document.md

# Use landscape orientation
printpdf -orientation landscape document.md

# Customize page margins
printpdf -margin 1cm document.md
printpdf -margin 1in document.md
printpdf -margin 20mm document.md
printpdf -margin 2cm -margin-top 3cm document.md
printpdf -margin-top 3cm -margin-right 1.5cm -margin-bottom 2.5cm -margin-left 2cm document.md

# Adjust zoom (scales all font sizes proportionally)
printpdf -zoom 80 document.md   # 80% of default size
printpdf -zoom 120 document.md  # 120% of default size

# Draw a vertical guide line 3cm from the left edge of the first page
printpdf -first-page-guide 3cm document.md

# Keep intermediate HTML/Typst files for inspection
printpdf --keep-artifacts document.md

# Combine options
printpdf -columns 2 -margin 1.5cm -zoom 90 document.md
printpdf -orientation landscape -margin 2cm -zoom 110 document.md
```

#### Column Layout

The `-columns` flag allows you to split the document into multiple columns, similar to newspaper layouts:

- Default: 1 column (standard single-column layout)
- `-columns 2`: Two-column layout
- `-columns 3`: Three-column layout
- Works with any number of columns

The implementation varies by converter:
- **Typst**: Uses native `columns` function with automatic text flow
- **WeasyPrint/Prince**: Uses CSS Multi-column Layout

Headings (h1, h2, h3) typically span all columns for better readability.

#### Page Orientation

The `-orientation` flag controls whether the page is in portrait or landscape mode:

- Default: `portrait` (standard vertical orientation)
- `landscape`: Horizontal orientation (useful for wider content)

The implementation:
- **Typst**: Uses `flipped: true` page attribute
- **WeasyPrint/Prince**: Uses CSS `@page { size: A4 landscape; }`

#### Page Margins

The `-margin` flag sets the default margin for all sides of the page:

- Default: `2cm` (2 centimeters on all sides)
- Supports various units: `cm`, `mm`, `in`, `pt`, etc.
- Examples:
  - `-margin 1cm`: 1 centimeter margins
  - `-margin 1in`: 1 inch margins
  - `-margin 20mm`: 20 millimeter margins

Use the side-specific flags to override individual edges while keeping a default for the rest:

- `-margin-top`
- `-margin-right`
- `-margin-bottom`
- `-margin-left`

For example, `printpdf -margin 2cm -margin-top 3cm document.md` increases only the top margin, and `printpdf -margin-top 3cm -margin-right 1.5cm -margin-bottom 2.5cm -margin-left 2cm` sets each side explicitly.

#### Zoom (Font Size Scaling)

The `-zoom` flag scales all font sizes proportionally:

- Default: `100` (100%, no scaling)
- Range: 1-500 (percentage)
- Examples:
  - `-zoom 80`: All fonts at 80% of their normal size
  - `-zoom 120`: All fonts at 120% of their normal size

#### Intermediate Artifacts

Use the `--keep-artifacts` flag when you need to inspect what each converter received. When enabled, printpdf saves the generated HTML and Typst sources in `OUTPUT/printpdf-artifacts/<converter>/` instead of deleting temporary files. Each run creates uniquely named files so you can diff or open them in another tool without re-running the conversion.

This is useful when you want to:
- Make the document more compact without manually adjusting individual font sizes
- Increase readability by making all text larger
- Fit more content on a page by reducing font sizes

The zoom affects:
- **All text elements**: headings, body text, code blocks, lists, etc.
- **Typst**: Adjusts the base font size (11pt by default becomes 8.8pt at 80%, 13.2pt at 120%)
- **HTML/CSS**: Uses `font-size` percentage on the root element, so all relative sizes scale proportionally

#### First Page Guide

Use the `-first-page-guide` flag to draw a thin vertical line on the first page. Provide the distance from the left edge using any supported CSS/Typst length (for example `3cm`, `1in`, `25mm`). The line starts at the top margin, extends to the bottom margin, and is omitted when the flag is not supplied.

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
mise run format
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
