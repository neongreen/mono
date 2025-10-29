# svghatch

A Go library for replacing solid color fills in SVG files with line-based patterns (hatching), making them suitable for black and white printing.

## Features

- Replace solid color fills with various hatching patterns
- Support for multiple pattern types:
  - Horizontal lines
  - Vertical lines
  - Diagonal lines (left and right)
  - Crosshatch
  - Dots
  - Grid
- Configurable pattern spacing and line width
- Handles multiple colors in a single SVG
- Works with various SVG elements (rect, circle, polygon, path, etc.)
- Preserves other SVG attributes and structure

## Installation

```bash
go get github.com/neongreen/mono/lib/svghatch
```

## Usage

### Basic Example

```go
package main

import (
    "bytes"
    "os"
    "github.com/neongreen/mono/lib/svghatch"
)

func main() {
    // Input SVG with colored rectangles
    input := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="100">
      <rect x="10" y="10" width="80" height="80" fill="red"/>
      <rect x="110" y="10" width="80" height="80" fill="blue"/>
    </svg>`

    // Define color-to-pattern mappings
    mappings := []svghatch.ColorMapping{
        {
            SourceColor: "red",
            Pattern:     svghatch.DefaultPatternConfig(svghatch.PatternHorizontal),
        },
        {
            SourceColor: "blue",
            Pattern:     svghatch.DefaultPatternConfig(svghatch.PatternVertical),
        },
    }

    // Create replacer and process
    replacer := svghatch.NewReplacer(mappings)
    var output bytes.Buffer
    
    err := replacer.Replace(bytes.NewBufferString(input), &output)
    if err != nil {
        panic(err)
    }
    
    // Write result
    os.WriteFile("output.svg", output.Bytes(), 0644)
}
```

### Pattern Types

The library supports 7 different pattern types:

```go
svghatch.PatternHorizontal    // Horizontal lines
svghatch.PatternVertical      // Vertical lines
svghatch.PatternDiagonalLeft  // Diagonal lines from top-left to bottom-right
svghatch.PatternDiagonalRight // Diagonal lines from top-right to bottom-left
svghatch.PatternCrosshatch    // Diagonal crosshatch pattern
svghatch.PatternDots          // Dot pattern
svghatch.PatternGrid          // Grid pattern
```

### Custom Pattern Configuration

You can customize pattern spacing and line width:

```go
config := svghatch.DefaultPatternConfig(svghatch.PatternHorizontal)
config.Spacing = 10.0  // Space between lines (default: 5.0)
config.Width = 2.0     // Line width (default: 1.0)

mapping := svghatch.ColorMapping{
    SourceColor: "red",
    Pattern:     config,
}
```

### Color Matching

The library supports various color formats:

- Named colors: `"red"`, `"blue"`, `"green"`, etc.
- Hex colors: `"#FF0000"`, `"#f00"` (3-digit shorthand)
- RGB colors: `"rgb(255, 0, 0)"`

Color matching is case-insensitive and handles format variations automatically.

## Examples

The `examples/` directory contains several demonstration programs:

- **basic.go** - Simple example with four colored rectangles
- **pattern_gallery.go** - Showcase of all 7 pattern types
- **custom_spacing.go** - Demonstrates different spacing values
- **line_width.go** - Demonstrates different line widths
- **chart.go** - Bar chart suitable for B&W printing
- **complex_shapes.go** - Circles, polygons, and paths with patterns

To run the examples:

```bash
cd examples
go run basic.go
go run pattern_gallery.go
# etc.
```

Each example generates an output SVG file that you can view in a browser or image viewer.

## API Reference

### Types

#### `PatternType`

```go
type PatternType string
```

Represents the type of hatching pattern to use.

#### `PatternConfig`

```go
type PatternConfig struct {
    Type    PatternType  // Pattern type to use
    Spacing float64      // Spacing between lines/dots (default: 5)
    Width   float64      // Line width (default: 1)
    Angle   float64      // Angle for diagonal patterns (default: 45)
}
```

Configuration for a hatching pattern.

#### `ColorMapping`

```go
type ColorMapping struct {
    SourceColor string         // Color to replace (e.g., "#FF0000", "red")
    Pattern     PatternConfig  // Pattern to use as replacement
}
```

Maps a source color to a pattern configuration.

### Functions

#### `DefaultPatternConfig`

```go
func DefaultPatternConfig(patternType PatternType) PatternConfig
```

Returns a pattern configuration with default values (spacing: 5.0, width: 1.0, angle: 45.0).

#### `NewReplacer`

```go
func NewReplacer(mappings []ColorMapping) *Replacer
```

Creates a new SVG replacer with the given color-to-pattern mappings.

#### `Replace`

```go
func (r *Replacer) Replace(input io.Reader, output io.Writer) error
```

Reads an SVG from input, replaces colors with patterns, and writes the result to output.

## Use Cases

- Converting colored charts and diagrams for academic papers
- Preparing graphics for black and white printing
- Creating accessible visualizations with distinguishable patterns
- Generating technical documentation with printer-friendly graphics
- Making infographics suitable for photocopying

## Implementation Details

The library:
1. Parses the input SVG using Go's `encoding/xml` package
2. Creates SVG `<pattern>` definitions for each color mapping
3. Walks through all SVG elements and replaces `fill` attributes
4. Handles both direct `fill` attributes and inline `style` attributes
5. Normalizes color values for consistent matching
6. Preserves the original SVG structure and non-color attributes

## Testing

Run the test suite:

```bash
go test ./...
```

The test suite includes:
- Color normalization tests
- Pattern generation tests
- Single and multiple color replacement tests
- Style attribute handling tests
- All pattern type tests
- Edge case tests (empty SVG, invalid XML, etc.)

## License

See the repository LICENSE file.
