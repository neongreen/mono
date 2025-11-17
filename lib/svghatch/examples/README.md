# svghatch Examples

This directory contains example programs demonstrating various features of the svghatch library.

## Running Examples

Each example is in its own subdirectory. To run an example:

```bash
cd <example-name>
go run *.go
```

The example will generate an SVG file in the same directory that you can view in a web browser or image viewer.

## Available Examples

### basic
Simple demonstration with four colored rectangles replaced with different patterns.

```bash
cd basic && go run basic.go
```

Generates: `basic_output.svg`

### pattern_gallery
Showcase of all 7 available pattern types in a single image.

```bash
cd pattern_gallery && go run pattern_gallery.go
```

Generates: `pattern_gallery.svg`

### custom_spacing
Demonstrates how different spacing values affect the pattern density.

```bash
cd custom_spacing && go run custom_spacing.go
```

Generates: `custom_spacing.svg`

### line_width
Shows the effect of different line widths on diagonal patterns.

```bash
cd line_width && go run line_width.go
```

Generates: `line_width.svg`

### chart
A practical example: bar chart made suitable for black and white printing.

```bash
cd chart && go run chart.go
```

Generates: `chart.svg`

### complex_shapes
Demonstrates that patterns work with various SVG shapes (circles, polygons, paths).

```bash
cd complex_shapes && go run complex_shapes.go
```

Generates: `complex_shapes.svg`

## Creating New Examples

To create a new example:

1. Create a new subdirectory with a descriptive name
2. Add a `main.go` file with package main
3. Import the svghatch library
4. Generate an SVG file as output
5. Update this README with a description

Each example should be self-contained and demonstrate a specific feature or use case.
