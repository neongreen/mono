package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/neongreen/mono/lib/svghatch"
)

func main() {
	// Create an SVG showcasing all pattern types
	patterns := []struct {
		name    string
		pattern svghatch.PatternType
	}{
		{"Horizontal", svghatch.PatternHorizontal},
		{"Vertical", svghatch.PatternVertical},
		{"Diagonal Left", svghatch.PatternDiagonalLeft},
		{"Diagonal Right", svghatch.PatternDiagonalRight},
		{"Crosshatch", svghatch.PatternCrosshatch},
		{"Dots", svghatch.PatternDots},
		{"Grid", svghatch.PatternGrid},
	}

	// Generate input SVG with different colored rectangles
	var inputSVG bytes.Buffer
	inputSVG.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="800" height="150">`)

	colors := []string{"#ff0000", "#00ff00", "#0000ff", "#ffff00", "#ff00ff", "#00ffff", "#ffa500"}
	for i, color := range colors {
		x := i*110 + 10
		inputSVG.WriteString(fmt.Sprintf(
			`<rect x="%d" y="20" width="90" height="90" fill="%s"/>`,
			x, color,
		))
		inputSVG.WriteString(fmt.Sprintf(
			`<text x="%d" y="130" font-size="10" text-anchor="middle">%s</text>`,
			x+45, patterns[i].name,
		))
	}
	inputSVG.WriteString(`</svg>`)

	// Create mappings for each pattern
	var mappings []svghatch.ColorMapping
	for i, color := range colors {
		mappings = append(mappings, svghatch.ColorMapping{
			SourceColor: color,
			Pattern:     svghatch.DefaultPatternConfig(patterns[i].pattern),
		})
	}

	// Process SVG
	replacer := svghatch.NewReplacer(mappings)
	var output bytes.Buffer

	err := replacer.Replace(&inputSVG, &output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	err = os.WriteFile("pattern_gallery.svg", output.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated pattern_gallery.svg")
}
