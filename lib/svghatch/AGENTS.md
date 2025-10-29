# svghatch Agent Guidelines

## Library Purpose

svghatch is an internal Go library for replacing solid color fills in SVG files with line-based patterns (hatching). It makes colored SVGs suitable for black and white printing.

## Project Status

Alpha - API is stable but may change based on usage patterns. Currently used internally, no external users yet.

## Code Style

- Follow standard Go conventions
- Use `go fmt` before committing
- Error messages should use `%w` for wrapping to preserve error chains
- Pattern generation functions should be self-contained and testable

## Testing

- All public functions should have unit tests
- Test with various SVG structures (rect, circle, path, polygon)
- Test color format variations (hex, rgb, named colors)
- Test edge cases (empty SVG, invalid XML, missing colors)

## Pattern Types

The library supports 7 pattern types:
1. Horizontal lines
2. Vertical lines
3. Diagonal left (top-left to bottom-right)
4. Diagonal right (top-right to bottom-left)
5. Crosshatch (diagonal lines in both directions)
6. Dots
7. Grid (horizontal and vertical lines)

New pattern types should follow the same structure and be added to the PatternType enum.

## SVG Parsing

The library uses a simplified XML parsing approach:
- Parses SVG into SVGNode structure with Children slice
- Does not use full DOM parsing for performance
- Preserves unknown attributes and structure
- Pattern definitions are added to <defs> element

When modifying SVG parsing:
- Test with real-world SVG files
- Ensure namespace handling is preserved
- Keep the parser lightweight

## Examples

Each example should be in its own subdirectory under examples/ to avoid "multiple main declarations" errors. Examples should:
- Be self-contained (single .go file)
- Generate output SVG files
- Print confirmation message when done
- Use clear, descriptive names

## Dependencies

The library depends only on the Go standard library:
- encoding/xml for SVG parsing
- io for stream handling
- Standard formatting and error handling packages

Keep it this way - do not add external dependencies without strong justification.
