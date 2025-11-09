# Readability WASM - Go Bindings for Mozilla Readability

This library provides Go bindings for article extraction using [Mozilla Readability](https://github.com/mozilla/readability), the same library that powers Firefox Reader View. It extracts clean, structured article content from web pages using WebAssembly.

**New here?** Check out the [Quick Start Guide](QUICKSTART.md) to get running in 5 minutes!

## Current Status

This library uses **Mozilla Readability** compiled to WASM via Javy. While originally inspired by the [Postlight Parser](https://github.com/postlight/parser), we use Readability because it has zero Node.js dependencies and works perfectly in the minimal Javy runtime.

Current capabilities:

- ✅ Extracts article titles, authors, and publication dates
- ✅ Intelligently filters navigation, footers, and ads
- ✅ Extracts main article content with clean HTML
- ✅ Provides word counts and excerpts
- ✅ Parses URLs and extracts domains
- ✅ Handles images embedded in articles
- ✅ Provides clean Go API via WASM/wazero
- ✅ Production-ready Mozilla Readability engine
- ✅ Same parser used in Firefox Reader View

## Why WASM?

By compiling the JavaScript parser to WebAssembly and running it via [wazero](https://wazero.io/), we get:

- **No CGO required** - Pure Go, easy cross-compilation
- **Embedded runtime** - No external dependencies at runtime
- **Sandboxed execution** - WASM provides isolation
- **Portable** - Works everywhere Go works

## Features

- Extract article title, author, content, and metadata
- Parse HTML directly or fetch from URL
- Returns structured data including word count, images, excerpts
- Fast and efficient WASM-based execution

## Installation

First, build the WASM module:

```bash
# Install dependencies (requires Node.js and cargo/rust)
make install

# Install javy (JavaScript to WASM compiler)
cargo install javy-cli

# Build the WASM module
make build-wasm
```

Then use the library in your Go code:

```go
import "github.com/neongreen/mono/lib/readability-wasm"
```

## Usage

### Parse HTML Directly

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/neongreen/mono/lib/readability-wasm"
)

func main() {
    ctx := context.Background()

    // Create a parser instance
    parser, err := readability.NewParser(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer parser.Close(ctx)

    // Parse HTML content
    html := `<html>...</html>` // Your HTML here
    article, err := parser.Parse(ctx, "https://example.com/article", html)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Title: %s\n", article.Title)
    fmt.Printf("Author: %s\n", article.Author)
    fmt.Printf("Content: %s\n", article.Content)
    fmt.Printf("Word Count: %d\n", article.WordCount)
}
```

### Parse from URL

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/neongreen/mono/lib/readability-wasm"
)

func main() {
    ctx := context.Background()

    // Create a parser instance
    parser, err := readability.NewParser(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer parser.Close(ctx)

    // Fetch and parse an article
    article, err := parser.ParseURL(ctx, "https://example.com/article")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Title: %s\n", article.Title)
    fmt.Printf("Author: %s\n", article.Author)
    fmt.Printf("Published: %s\n", article.DatePublished)
    fmt.Printf("Content: %s\n", article.Content)
}
```

## API

### Types

#### `Article`

The main article data structure:

```go
type Article struct {
    Title         string // Article title
    Author        string // Article author
    DatePublished string // Publication date
    Dek           string // Article excerpt/description
    LeadImageURL  string // URL of the lead image
    Content       string // Main content (HTML)
    Excerpt       string // Short excerpt
    WordCount     int    // Word count
    Direction     string // Text direction (ltr/rtl)
    TotalPages    int    // Total pages (for multi-page articles)
    RenderedPages int    // Rendered pages
    NextPageURL   string // Next page URL (for multi-page articles)
    URL           string // Original URL
    Domain        string // Domain name
}
```

#### `Parser`

The main parser type:

```go
type Parser struct {
    // internal fields
}

// NewParser creates a new parser instance
func NewParser(ctx context.Context) (*Parser, error)

// Parse extracts article content from HTML
func (p *Parser) Parse(ctx context.Context, url, html string) (*Article, error)

// ParseURL fetches and parses an article from a URL
func (p *Parser) ParseURL(ctx context.Context, url string) (*Article, error)

// Close releases resources used by the parser
func (p *Parser) Close(ctx context.Context) error
```

## Build System

The library includes a Makefile for building the WASM module:

```bash
make help          # Show available targets
make install       # Install npm dependencies
make build-js      # Bundle JavaScript code
make build-wasm    # Compile to WASM (requires javy)
make clean         # Remove build artifacts
```

## Architecture

```
┌─────────────────┐
│   Go Code       │
│  (Your App)     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  lib/postlight  │
│  (Go bindings)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    wazero       │
│ (WASM runtime)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  parser.wasm    │
│  (Readability)  │
└─────────────────┘
```

The workflow:
1. Go code calls the library
2. Library uses wazero to execute WASM
3. WASM module (compiled from Readability.js) parses the HTML
4. Results are returned to Go as structured data

## Requirements

### Build Time
- Node.js and npm (for bundling JavaScript)
- Rust and cargo (for installing javy)
- javy (`cargo install javy-cli`)

### Runtime
- None! The WASM module is embedded in the Go binary

## Testing

```bash
# Build the WASM module first
make build-wasm

# Run tests
go test -v
```

## License

This library is a Go wrapper around [Mozilla Readability](https://github.com/mozilla/readability), which is Apache 2.0 licensed.

## Credits

- [Mozilla Readability](https://github.com/mozilla/readability) - The excellent article extraction library from Firefox
- [wazero](https://wazero.io/) - Pure Go WebAssembly runtime
- [javy](https://github.com/bytecodealliance/javy) - JavaScript to WASM compiler with promise support
