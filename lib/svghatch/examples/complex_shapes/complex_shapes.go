package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/neongreen/mono/lib/svghatch"
)

func main() {
	// SVG with various shapes (circles, paths, polygons)
	input := `<svg xmlns="http://www.w3.org/2000/svg" width="400" height="200">
  <!-- Circle -->
  <circle cx="70" cy="100" r="50" fill="red"/>
  
  <!-- Polygon (triangle) -->
  <polygon points="200,50 160,140 240,140" fill="blue"/>
  
  <!-- Path (star) -->
  <path d="M 320,50 L 335,90 L 375,95 L 345,120 L 355,160 L 320,135 L 285,160 L 295,120 L 265,95 L 305,90 Z" fill="green"/>
</svg>`

	// Map different shapes to different patterns
	config1 := svghatch.DefaultPatternConfig(svghatch.PatternDots)
	config1.Spacing = 6.0
	config1.Width = 1.5

	config2 := svghatch.DefaultPatternConfig(svghatch.PatternDiagonalRight)
	config2.Spacing = 4.0

	config3 := svghatch.DefaultPatternConfig(svghatch.PatternGrid)
	config3.Spacing = 8.0

	mappings := []svghatch.ColorMapping{
		{SourceColor: "red", Pattern: config1},
		{SourceColor: "blue", Pattern: config2},
		{SourceColor: "green", Pattern: config3},
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
	err = os.WriteFile("complex_shapes.svg", output.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated complex_shapes.svg - circles, polygons, and paths with patterns")
}
