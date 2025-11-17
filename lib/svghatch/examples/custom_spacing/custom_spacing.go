package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/neongreen/mono/lib/svghatch"
)

func main() {
	// SVG demonstrating different spacing values
	input := `<svg xmlns="http://www.w3.org/2000/svg" width="450" height="150">
  <rect x="10" y="10" width="130" height="130" fill="#ff0000"/>
  <rect x="160" y="10" width="130" height="130" fill="#00ff00"/>
  <rect x="310" y="10" width="130" height="130" fill="#0000ff"/>
</svg>`

	// Create configs with different spacing
	config1 := svghatch.DefaultPatternConfig(svghatch.PatternHorizontal)
	config1.Spacing = 3.0

	config2 := svghatch.DefaultPatternConfig(svghatch.PatternHorizontal)
	config2.Spacing = 8.0

	config3 := svghatch.DefaultPatternConfig(svghatch.PatternHorizontal)
	config3.Spacing = 15.0

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
	err = os.WriteFile("custom_spacing.svg", output.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated custom_spacing.svg - shows spacing of 3, 8, and 15")
}
