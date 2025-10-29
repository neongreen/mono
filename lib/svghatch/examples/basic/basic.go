package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/neongreen/mono/lib/svghatch"
)

func main() {
	// Simple SVG with colored rectangles
	input := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
  <rect x="10" y="10" width="80" height="80" fill="red"/>
  <rect x="110" y="10" width="80" height="80" fill="blue"/>
  <rect x="10" y="110" width="80" height="80" fill="green"/>
  <rect x="110" y="110" width="80" height="80" fill="yellow"/>
</svg>`

	// Create color mappings
	mappings := []svghatch.ColorMapping{
		{
			SourceColor: "red",
			Pattern:     svghatch.DefaultPatternConfig(svghatch.PatternHorizontal),
		},
		{
			SourceColor: "blue",
			Pattern:     svghatch.DefaultPatternConfig(svghatch.PatternVertical),
		},
		{
			SourceColor: "green",
			Pattern:     svghatch.DefaultPatternConfig(svghatch.PatternDiagonalLeft),
		},
		{
			SourceColor: "yellow",
			Pattern:     svghatch.DefaultPatternConfig(svghatch.PatternDots),
		},
	}

	// Create replacer and process SVG
	replacer := svghatch.NewReplacer(mappings)
	var output bytes.Buffer

	err := replacer.Replace(bytes.NewBufferString(input), &output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	err = os.WriteFile("basic_output.svg", output.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated basic_output.svg")
}
