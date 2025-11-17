package svghatch

import (
	"bytes"
	"strings"
	"testing"
)

func TestNormalizeColor(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"red", "#ff0000"},
		{"RED", "#ff0000"},
		{"  Red  ", "#ff0000"},
		{"#f00", "#ff0000"},
		{"#FF0000", "#ff0000"},
		{"#ff0000", "#ff0000"},
		{"rgb(255, 0, 0)", "#ff0000"},
		{"rgb(128, 128, 128)", "#808080"},
		{"black", "#000000"},
		{"white", "#ffffff"},
		{"blue", "#0000ff"},
		{"#abc", "#aabbcc"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeColor(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeColor(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDefaultPatternConfig(t *testing.T) {
	config := DefaultPatternConfig(PatternHorizontal)
	if config.Type != PatternHorizontal {
		t.Errorf("Expected type %v, got %v", PatternHorizontal, config.Type)
	}
	if config.Spacing != 5.0 {
		t.Errorf("Expected spacing 5.0, got %f", config.Spacing)
	}
	if config.Width != 1.0 {
		t.Errorf("Expected width 1.0, got %f", config.Width)
	}
	if config.Angle != 45.0 {
		t.Errorf("Expected angle 45.0, got %f", config.Angle)
	}
}

func TestReplacerSimpleRect(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <rect x="10" y="10" width="30" height="30" fill="red"/>
</svg>`

	mappings := []ColorMapping{
		{
			SourceColor: "red",
			Pattern:     DefaultPatternConfig(PatternHorizontal),
		},
	}

	replacer := NewReplacer(mappings)
	var output bytes.Buffer

	err := replacer.Replace(strings.NewReader(input), &output)
	if err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	result := output.String()

	// Check that a pattern definition was added
	if !strings.Contains(result, "<pattern") {
		t.Error("Result should contain pattern definition")
	}

	// Check that the pattern is referenced
	if !strings.Contains(result, "url(#pattern-0)") {
		t.Error("Result should reference pattern-0")
	}

	// Check that original color is replaced
	if strings.Contains(result, `fill="red"`) {
		t.Error("Original red fill should be replaced")
	}
}

func TestReplacerMultipleColors(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <rect x="10" y="10" width="30" height="30" fill="red"/>
  <circle cx="70" cy="70" r="15" fill="blue"/>
</svg>`

	mappings := []ColorMapping{
		{
			SourceColor: "red",
			Pattern:     DefaultPatternConfig(PatternHorizontal),
		},
		{
			SourceColor: "blue",
			Pattern:     DefaultPatternConfig(PatternVertical),
		},
	}

	replacer := NewReplacer(mappings)
	var output bytes.Buffer

	err := replacer.Replace(strings.NewReader(input), &output)
	if err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	result := output.String()

	// Check that both patterns are defined
	if !strings.Contains(result, "pattern-0") {
		t.Error("Result should contain pattern-0")
	}
	if !strings.Contains(result, "pattern-1") {
		t.Error("Result should contain pattern-1")
	}

	// Check that both patterns are referenced
	if !strings.Contains(result, "url(#pattern-0)") {
		t.Error("Result should reference pattern-0")
	}
	if !strings.Contains(result, "url(#pattern-1)") {
		t.Error("Result should reference pattern-1")
	}
}

func TestReplacerWithStyleAttribute(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <rect x="10" y="10" width="30" height="30" style="fill: red; stroke: black"/>
</svg>`

	mappings := []ColorMapping{
		{
			SourceColor: "red",
			Pattern:     DefaultPatternConfig(PatternDots),
		},
	}

	replacer := NewReplacer(mappings)
	var output bytes.Buffer

	err := replacer.Replace(strings.NewReader(input), &output)
	if err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	result := output.String()

	// Check that style contains pattern reference
	if !strings.Contains(result, "url(#pattern-0)") {
		t.Error("Result should reference pattern-0 in style attribute")
	}
}

func TestPatternTypes(t *testing.T) {
	patterns := []PatternType{
		PatternHorizontal,
		PatternVertical,
		PatternDiagonalLeft,
		PatternDiagonalRight,
		PatternCrosshatch,
		PatternDots,
		PatternGrid,
	}

	input := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <rect x="10" y="10" width="30" height="30" fill="red"/>
</svg>`

	for _, pattern := range patterns {
		t.Run(string(pattern), func(t *testing.T) {
			mappings := []ColorMapping{
				{
					SourceColor: "red",
					Pattern:     DefaultPatternConfig(pattern),
				},
			}

			replacer := NewReplacer(mappings)
			var output bytes.Buffer

			err := replacer.Replace(strings.NewReader(input), &output)
			if err != nil {
				t.Fatalf("Replace with pattern %v failed: %v", pattern, err)
			}

			result := output.String()
			if !strings.Contains(result, "<pattern") {
				t.Errorf("Pattern %v should create pattern definition", pattern)
			}
		})
	}
}

func TestPatternConfigSpacing(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <rect x="10" y="10" width="30" height="30" fill="red"/>
</svg>`

	config := DefaultPatternConfig(PatternHorizontal)
	config.Spacing = 10.0

	mappings := []ColorMapping{
		{
			SourceColor: "red",
			Pattern:     config,
		},
	}

	replacer := NewReplacer(mappings)
	var output bytes.Buffer

	err := replacer.Replace(strings.NewReader(input), &output)
	if err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	result := output.String()

	// Check that spacing is reflected in pattern width/height
	if !strings.Contains(result, `width="10.0"`) {
		t.Error("Pattern should have width matching spacing")
	}
}

func TestPatternConfigLineWidth(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <rect x="10" y="10" width="30" height="30" fill="red"/>
</svg>`

	config := DefaultPatternConfig(PatternHorizontal)
	config.Width = 2.0

	mappings := []ColorMapping{
		{
			SourceColor: "red",
			Pattern:     config,
		},
	}

	replacer := NewReplacer(mappings)
	var output bytes.Buffer

	err := replacer.Replace(strings.NewReader(input), &output)
	if err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	result := output.String()

	// Check that line width is reflected in stroke-width
	if !strings.Contains(result, `stroke-width="2.0"`) {
		t.Error("Pattern should have stroke-width matching config width")
	}
}

func TestEmptySVG(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"></svg>`

	mappings := []ColorMapping{
		{
			SourceColor: "red",
			Pattern:     DefaultPatternConfig(PatternHorizontal),
		},
	}

	replacer := NewReplacer(mappings)
	var output bytes.Buffer

	err := replacer.Replace(strings.NewReader(input), &output)
	if err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	// Should succeed even with empty SVG
	result := output.String()
	if !strings.Contains(result, "<svg") {
		t.Error("Result should contain svg tag")
	}
}

func TestInvalidXML(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <rect x="10" y="10" width="30" height="30" fill="red"
</svg>`

	mappings := []ColorMapping{
		{
			SourceColor: "red",
			Pattern:     DefaultPatternConfig(PatternHorizontal),
		},
	}

	replacer := NewReplacer(mappings)
	var output bytes.Buffer

	err := replacer.Replace(strings.NewReader(input), &output)
	if err == nil {
		t.Error("Expected error for invalid XML, got none")
	}
}

func TestHexColorVariants(t *testing.T) {
	input := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <rect x="10" y="10" width="30" height="30" fill="#f00"/>
</svg>`

	mappings := []ColorMapping{
		{
			SourceColor: "#FF0000",
			Pattern:     DefaultPatternConfig(PatternHorizontal),
		},
	}

	replacer := NewReplacer(mappings)
	var output bytes.Buffer

	err := replacer.Replace(strings.NewReader(input), &output)
	if err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	result := output.String()

	// Should match #f00 with #FF0000
	if !strings.Contains(result, "url(#pattern-0)") {
		t.Error("Should match shortened hex color with full hex color")
	}
}
