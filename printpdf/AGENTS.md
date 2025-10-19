# Agent Guidelines for printpdf

## Code Formatting

**All Go code must be formatted with `go fmt` before work is considered complete.**

Before submitting any changes:
- Run `go fmt ./...` in the printpdf directory
- Ensure all Go files are properly formatted
- This applies to both new and modified Go code

## bd-336 Footnote notes

- Prince footnotes expect the note body to remain inline (for example in a `span`) with `float: footnote`, and rely on `::footnote-marker` and `::footnote-call` for numbering so the marker stays aligned with the text.
- Always debug footnote rendering with the `--keep-artifacts` flag so the intermediate HTML and Typst sources are available in `printpdf-artifacts/` for inspection.
