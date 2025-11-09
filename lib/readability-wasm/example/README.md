# Postlight Parser Example

This example demonstrates how to use the Postlight Parser library.

## Building

First, build the WASM module in the parent directory:

```bash
cd ..
make build-wasm
cd example
```

Then build the example:

```bash
go build -o parser-example
```

## Usage

Parse an article from a URL:

```bash
./parser-example -url https://example.com/article
```

Output as JSON:

```bash
./parser-example -url https://example.com/article -json
```

Parse HTML content directly:

```bash
./parser-example -url https://example.com -html "<html>...</html>"
```

## Options

- `-url <URL>` - URL to parse (required)
- `-html <HTML>` - HTML content to parse (optional, will fetch URL if not provided)
- `-json` - Output as JSON instead of human-readable format
