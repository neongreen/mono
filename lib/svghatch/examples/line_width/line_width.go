package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/neongreen/mono/lib/svghatch"
)

func main() {
	// SVG demonstrating different line widths
	input := `<svg xmlns="http://www.w3.org/2000/svg" width="450" height="150">
  <rect x="10" y="10" width="130" height="130" fill="#ff0000"/>
  <rect x="160" y="10" width="130" height="130" fill="#00ff00"/>
  <rect x="310" y="10" width="130" height="130" fill="#0000ff"/>
</svg>`

	// Create configs with different line widths
	config1 := svghatch.DefaultPatternConfig(svghatch.PatternDiagonalLeft)
	config1.Width = 0.5

	config2 := svghatch.DefaultPatternConfig(svghatch.PatternDiagonalLeft)
	config2.Width = 1.5

	config3 := svghatch.DefaultPatternConfig(svghatch.PatternDiagonalLeft)
	config3.Width = 3.0

	mappings := []svghatch.ColorMapping{
		{SourceColor: "#ff0000", Pattern: config1},
		{SourceColor: "#00ff00", Pattern: config2},
		{SourceColor: "#0000ff", Pattern: config3},
	}

	// Process SVG
	replacer := svghatch.NewReplacer(mappings)
	var output bytes.Buffer

	err := replacer.Replace(bytes.NewBufferString(input), &output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	err = os.WriteFile("line_width.svg", output.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated line_width.svg - shows line widths of 0.5, 1.5, and 3.0")
}
