# Implementation

## Extraction pipeline

Extraction pipeline:

- Walks all .go files under the directory, skipping *_test.go.

- Parses with comments and pulls lion markers from package doc, func doc, and type/const/var

doc comments (first name in a const/var block is used as the entity).

- Supports single-line markers and block comment markers (marker at top of the doc block).

- Aggregates snippets per topic across files; generator writes one file per topic.

*Source: `lion/internal/extractor/extractor.go:26`*

