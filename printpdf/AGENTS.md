# Agent Guidelines for printpdf

## Build, Test, and Run Commands

**All commands must be run from the repository root (`/home/user/mono`).**

```bash
# Build
go build ./printpdf

# Test
go test ./printpdf/...

# Run
go run ./printpdf [args...]

# Install (builds and places in $GOPATH/bin)
go install ./printpdf
```

**Important:** Use `go` commands directly. Do not use `mise` for building or running printpdf.

## bd-336 Footnote notes

- Prince footnotes expect the note body to remain inline (for example in a `span`) with `float: footnote`, and rely on `::footnote-marker` and `::footnote-call` for numbering so the marker stays aligned with the text.
- Always debug footnote rendering with the `--keep-artifacts` flag so the intermediate HTML and Typst sources are available in `printpdf-artifacts/` for inspection.
- Before touching footnote code or styling, run `curl -s https://pure.md/https://www.princexml.com/howcome/2022/guides/footnotes/` and re-read the Prince footnote guide.
