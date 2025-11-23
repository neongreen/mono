# Supported Syntax

## Supported syntax

Supported syntax formats:

1. Single-line marker first:

//lion:topic-name metadata...

// Content lines (optional)

2. Block comment:

/*lion:topic-name metadata...

Multi-line content

*/

Optional metadata on any lion marker:

- title="Custom Title" overrides the topic's display title (file H1 + index link).

- section="Section Title" overrides the heading for that entry; section="" suppresses it.

- Unknown keys stop metadata parsing and are treated as content.

All formats attach documentation to the next declaration (function, type, const, var).

*Source: `lion/internal/extractor/extractor.go:119`*

