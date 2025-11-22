# File Title

## Titles and headings

The generated file title comes from the topic display title:
- Default: Title Case of the topic slug (e.g., "getting-started" -> "Getting Started").
- Override: set title="Custom Title" on any lion marker for that topic; conflicts fail generation.
Entry headings:
- Default: the attached entity name (package <name>, function name, first const/var in the block).
- Override per entry: section="Custom Section"; section="" suppresses the heading entirely (will emit a warning).
- Conflicting section titles within the same comment group fail extraction.

*Source: `lion/internal/generator/generator.go:23`*

