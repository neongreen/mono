# Postlight Parser WASM Library

A Go library that provides bindings to Postlight Parser functionality using WebAssembly. This library extracts clean article content from web pages without requiring CGO, making it portable and easy to integrate.

## Features

- **No CGO required**: Pure Go implementation with WASM support via wazero
- **Article extraction**: Extracts title, content, author, date, and other metadata
- **Flexible API**: Support for parsing HTML strings and pre-fetched JSON
- **Context support**: All parsing operations accept context for cancellation
- **Comprehensive testing**: Full test coverage with real-world examples

## Installation

```bash
go get github.com/neongreen/mono/lib/postlight-parser-wasm
```

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/neongreen/mono/lib/postlight-parser-wasm"
)

func main() {
    // Create a parser instance
    p, err := parser.New()
    if err != nil {
        log.Fatal(err)
    }
    defer p.Close()

    // HTML content to parse
    html := `<!DOCTYPE html>
<html>
<head>
    <title>My Article</title>
</head>
<body>
    <h1>Welcome</h1>
    <p>This is the main content of the article.</p>
</body>
</html>`

    // Extract article information
    ctx := context.Background()
    article, err := p.Extract(ctx, "https://example.com/article", html)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Title: %s\n", article.Title)
    fmt.Printf("Domain: %s\n", article.Domain)
    fmt.Printf("Word Count: %d\n", article.WordCount)
    fmt.Printf("Content: %s\n", article.Content)
}
```

### Advanced Example with WASM

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/neongreen/mono/lib/postlight-parser-wasm"
)

func main() {
    ctx := context.Background()
    
    // Create parser with WASM runtime
    p, err := parser.NewWithWasm(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer p.Close()

    html := `<html>
<head><title>Advanced Article</title></head>
<body>
    <article>
        <h1>Main Heading</h1>
        <p>Article content goes here...</p>
    </article>
</body>
</html>`

    article, err := p.Extract(ctx, "https://blog.example.com/post", html)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Extracted: %s from %s\n", article.Title, article.Domain)
}
```

### Parsing JSON Results

```go
package main

import (
    "fmt"
    "log"

    "github.com/neongreen/mono/lib/postlight-parser-wasm"
)

func main() {
    p, err := parser.New()
    if err != nil {
        log.Fatal(err)
    }
    defer p.Close()

    // Parse pre-fetched JSON data
    jsonData := []byte(`{
        "title": "Article Title",
        "content": "<p>Content</p>",
        "url": "https://example.com",
        "domain": "example.com",
        "word_count": 10
    }`)

    article, err := p.ExtractFromJSON(jsonData)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Parsed: %s\n", article.Title)
}
```

## API Reference

### Types

#### `Article`

Represents the parsed content from a web page.

```go
type Article struct {
    Title         string  // Article title
    Content       string  // Main article content (HTML)
    Author        *string // Author name (optional)
    DatePublished *string // Publication date (optional)
    LeadImageURL  *string // URL of lead image (optional)
    Dek           *string // Article summary/description (optional)
    URL           string  // Original article URL
    Domain        string  // Domain of the article
    Excerpt       *string // Short excerpt (optional)
    WordCount     int     // Approximate word count
    Direction     string  // Text direction (ltr/rtl)
}
```

### Functions

#### `New() (*Parser, error)`

Creates a new Parser instance using the default implementation.

#### `NewWithWasm(ctx context.Context) (*Parser, error)`

Creates a new Parser instance with WASM runtime support. This provides better isolation and allows for future integration with JavaScript-based parsers.

### Methods

#### `Extract(ctx context.Context, url string, html string) (*Article, error)`

Extracts article information from HTML content.

- `ctx`: Context for cancellation
- `url`: The URL of the article (used for context and relative links)
- `html`: Raw HTML content to parse

Returns the extracted article information or an error.

#### `ExtractFromJSON(data []byte) (*Article, error)`

Parses article information from JSON data.

- `data`: JSON-encoded article data

Returns the parsed article or an error.

#### `Close() error`

Releases resources associated with the parser. Should be called when done using the parser (typically via `defer`).

## Implementation Notes

This library is designed to integrate with Postlight Parser functionality. The current implementation provides:

1. **Go-based parsing**: Fast, native HTML parsing without external dependencies
2. **WASM infrastructure**: Foundation for running JavaScript parsers via wazero
3. **Extensible design**: Easy to swap implementations or add new features

### WASM Integration

The library includes wazero runtime support for running WebAssembly modules. This allows for:

- Running JavaScript-based parsers compiled to WASM
- No CGO dependency - pure Go with WASM
- Better isolation and security
- Cross-platform compatibility

The embedded JavaScript parser (`parser_wasm.js`) can be compiled to WASM and executed through wazero for enhanced parsing capabilities.

## Testing

Run the test suite:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

The library includes comprehensive tests covering:
- Basic HTML parsing
- Error handling
- Complex HTML structures
- Real-world article examples
- JSON parsing
- WASM runtime integration

## Future Enhancements

- [ ] Full Postlight Parser WASM integration
- [ ] Support for custom parser configurations
- [ ] Markdown output format
- [ ] Enhanced metadata extraction
- [ ] Content sanitization options
- [ ] Streaming parser for large documents

## License

See the repository root for license information.
