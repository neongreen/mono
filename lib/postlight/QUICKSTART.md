# Quick Start Guide

Get up and running with the Postlight Parser library in 5 minutes.

## TL;DR

```bash
# 1. Install dependencies
cargo install javy-cli

# 2. Build WASM module
cd lib/postlight
make build-wasm

# 3. Use in your Go code
go run example/main.go -url https://example.com/article
```

## Step-by-Step

### 1. Check Dependencies

Run the dependency checker:

```bash
./scripts/check-deps.sh
```

This will tell you what's missing. Install any missing dependencies:

- **Node.js**: https://nodejs.org/
- **Rust/Cargo**: https://rustup.rs/
- **Javy**: `cargo install javy-cli`

### 2. Build the WASM Module

```bash
cd lib/postlight
make
```

This will:
- Install npm packages
- Bundle the JavaScript code
- Compile to WASM
- Create `parser.wasm` (embedded in your Go binary)

### 3. Try the Example

```bash
cd example
go build -o parser-example
./parser-example -url https://blog.golang.org/go1.18
```

You should see the parsed article content!

### 4. Use in Your Code

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/neongreen/mono/lib/postlight"
)

func main() {
    ctx := context.Background()

    // Create parser
    parser, err := postlight.NewParser(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer parser.Close(ctx)

    // Parse an article
    article, err := parser.ParseURL(ctx, "https://example.com/article")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Title: %s\n", article.Title)
    fmt.Printf("Author: %s\n", article.Author)
    fmt.Printf("Content: %s\n", article.Content)
}
```

## Troubleshooting

### Build fails with "javy: command not found"

Install javy:

```bash
cargo install javy-cli
```

Make sure `~/.cargo/bin` is in your PATH.

### Build takes a long time

The first build takes longer because it:
- Downloads npm packages
- Bundles all JavaScript code
- Compiles to WASM

Subsequent builds are faster.

### WASM module seems large

The parser.wasm file is several MB because it includes:
- Postlight Parser + dependencies
- QuickJS JavaScript engine

This is normal and expected.

### Tests fail

Make sure you've built the WASM module first:

```bash
make build-wasm
```

Then run tests:

```bash
go test -v
```

## Next Steps

- Read [README.md](README.md) for full API documentation
- Read [BUILDING.md](BUILDING.md) for detailed build information
- Check out [example/](example/) for more usage examples
- Explore the Postlight Parser documentation: https://github.com/postlight/parser

## Getting Help

If you run into issues:

1. Check [BUILDING.md](BUILDING.md) for detailed build instructions
2. Run `./scripts/check-deps.sh` to verify dependencies
3. Try `make clean && make` to rebuild from scratch
4. Check the Postlight Parser GitHub issues

## Pro Tips

### Parser Reuse

Create one parser and reuse it for multiple articles:

```go
parser, _ := postlight.NewParser(ctx)
defer parser.Close(ctx)

for _, url := range urls {
    article, _ := parser.ParseURL(ctx, url)
    // process article
}
```

### Custom HTTP Client

For more control over HTTP requests, fetch HTML yourself:

```go
resp, _ := http.Get(url)
html, _ := io.ReadAll(resp.Body)
article, _ := parser.Parse(ctx, url, string(html))
```

This lets you:
- Set custom headers
- Handle authentication
- Use a custom HTTP client
- Cache responses
- Handle rate limiting

### JSON Output

The Article struct has JSON tags for easy serialization:

```go
article, _ := parser.ParseURL(ctx, url)
json.NewEncoder(os.Stdout).Encode(article)
```
