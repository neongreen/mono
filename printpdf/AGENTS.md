# Agent Guidelines for printpdf

## Build, Test, and Run Commands

**All commands must be run from the mono repository root.**

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

## Footnote Rendering (bd-336, bd-342, bd-335)

**Implementation Status**: Complete

- Prince footnotes use inline `<span>` elements with `float: footnote` CSS
- Footnote numbering is handled entirely by Prince via `::footnote-call` and `::footnote-marker` pseudo-elements
- The HTML generator creates empty `<sup>` elements - Prince fills them automatically
- Inline footnote spans have their `::footnote-marker` hidden to avoid duplicates
- Always debug footnote rendering with the `--keep-artifacts` flag so the intermediate HTML and Typst sources are available in `printpdf-artifacts/` for inspection
- Before touching footnote code or styling, run `curl -s https://pure.md/https://www.princexml.com/howcome/2022/guides/footnotes/` and re-read the Prince footnote guide

**Key Implementation Details**:
- `normalizeFootnoteCall()` removes the anchor but does NOT add marker text
- CSS includes `sup[id^="fnref:"]::footnote-call` for Prince to generate numbers
- CSS includes `span.printpdf-footnote::footnote-marker { content: ""; }` to hide duplicate markers
- Multi-page footnote test (`footnoteMultiPageTestCase`) verifies layout across page breaks
