package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/neongreen/mono/lib/svghatch"
)

func main() {
	// A simple bar chart that would be suitable for B&W printing
	input := `<svg xmlns="http://www.w3.org/2000/svg" width="400" height="300">
  <!-- Chart title -->
  <text x="200" y="20" font-size="16" text-anchor="middle" font-weight="bold">Sales by Quarter</text>
  
  <!-- X-axis -->
  <line x1="50" y1="250" x2="350" y2="250" stroke="black" stroke-width="2"/>
  
  <!-- Y-axis -->
  <line x1="50" y1="50" x2="50" y2="250" stroke="black" stroke-width="2"/>
  
  <!-- Bars -->
  <rect x="80" y="150" width="50" height="100" fill="#ff0000"/>
  <text x="105" y="265" font-size="12" text-anchor="middle">Q1</text>
  
  <rect x="150" y="100" width="50" height="150" fill="#00ff00"/>
  <text x="175" y="265" font-size="12" text-anchor="middle">Q2</text>
  
  <rect x="220" y="80" width="50" height="170" fill="#0000ff"/>
  <text x="245" y="265" font-size="12" text-anchor="middle">Q3</text>
  
  <rect x="290" y="120" width="50" height="130" fill="#ffff00"/>
  <text x="315" y="265" font-size="12" text-anchor="middle">Q4</text>
  
  <!-- Y-axis labels -->
  <text x="40" y="255" font-size="10" text-anchor="end">0</text>
  <text x="40" y="205" font-size="10" text-anchor="end">50</text>
  <text x="40" y="155" font-size="10" text-anchor="end">100</text>
  <text x="40" y="105" font-size="10" text-anchor="end">150</text>
  <text x="40" y="55" font-size="10" text-anchor="end">200</text>
</svg>`

	// Different patterns for each quarter
	mappings := []svghatch.ColorMapping{
		{
			SourceColor: "#ff0000",
			Pattern:     svghatch.DefaultPatternConfig(svghatch.PatternHorizontal),
		},
		{
			SourceColor: "#00ff00",
			Pattern:     svghatch.DefaultPatternConfig(svghatch.PatternVertical),
		},
		{
			SourceColor: "#0000ff",
			Pattern:     svghatch.DefaultPatternConfig(svghatch.PatternDiagonalLeft),
		},
		{
			SourceColor: "#ffff00",
			Pattern:     svghatch.DefaultPatternConfig(svghatch.PatternCrosshatch),
		},
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
	err = os.WriteFile("chart.svg", output.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated chart.svg - bar chart suitable for B&W printing")
}
