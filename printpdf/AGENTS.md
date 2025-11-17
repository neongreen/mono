# Agent Guidelines for printpdf

## bd-336 Footnote notes

- Prince footnotes expect the note body to remain inline (for example in a `span`) with `float: footnote`, and rely on `::footnote-marker` and `::footnote-call` for numbering so the marker stays aligned with the text.
- Always debug footnote rendering with the `--keep-artifacts` flag so the intermediate HTML and Typst sources are available in `printpdf-artifacts/` for inspection.
- Before touching footnote code or styling, run `curl -s https://pure.md/https://www.princexml.com/howcome/2022/guides/footnotes/` and re-read the Prince footnote guide.
