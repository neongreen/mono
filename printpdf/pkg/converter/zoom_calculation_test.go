package converter

import (
	"strings"
	"testing"
)

func TestZoomCalculationsTypst(t *testing.T) {
	markdown := []byte("# Test")

	tests := []struct {
		zoom         int
		expectedSize float64
	}{
		{50, 5.50},   // 11pt * 0.50 = 5.50pt
		{80, 8.80},   // 11pt * 0.80 = 8.80pt
		{100, 11.00}, // 11pt * 1.00 = 11.00pt (default)
		{120, 13.20}, // 11pt * 1.20 = 13.20pt
		{150, 16.50}, // 11pt * 1.50 = 16.50pt
		{200, 22.00}, // 11pt * 2.00 = 22.00pt
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.zoom)), func(t *testing.T) {
			options := PageOptions{
				Columns:     1,
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        tt.zoom,
			}

			result, err := convertMarkdownToTypst(markdown, options)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}

			expectedText := formatFloat(tt.expectedSize) + "pt"
			if !strings.Contains(result, expectedText) {
				t.Errorf("Zoom %d%%: Expected '%s' but not found in output", tt.zoom, expectedText)
			}
		})
	}
}

func TestZoomCalculationsHTML(t *testing.T) {
	markdown := []byte("# Test")

	tests := []struct {
		zoom               int
		expectedPercentage string
	}{
		{50, "font-size: 50%"},
		{80, "font-size: 80%"},
		{100, "font-size: 100%"},
		{120, "font-size: 120%"},
		{150, "font-size: 150%"},
		{200, "font-size: 200%"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.zoom)), func(t *testing.T) {
			options := PageOptions{
				Columns:     1,
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        tt.zoom,
			}

			result, err := convertMarkdownToHTML(markdown, options)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}

			resultStr := string(result)
			if !strings.Contains(resultStr, tt.expectedPercentage) {
				t.Errorf("Zoom %d%%: Expected '%s' but not found in output", tt.zoom, tt.expectedPercentage)
			}
		})
	}
}

func formatFloat(f float64) string {
	// Match the %.2f format used in the actual code
	switch f {
	case 5.50:
		return "5.50"
	case 8.80:
		return "8.80"
	case 11.00:
		return "11.00"
	case 13.20:
		return "13.20"
	case 16.50:
		return "16.50"
	case 22.00:
		return "22.00"
	}
	return ""
}
