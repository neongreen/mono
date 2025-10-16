package converter

import (
	"strings"
	"testing"
)

func TestTypstConverterWithCustomMargin(t *testing.T) {
	markdown := []byte("# Test\n\nSome text")

	tests := []struct {
		name           string
		options        PageOptions
		expectedMargin string
	}{
		{
			name: "default margin",
			options: PageOptions{
				Columns:     1,
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        100,
			},
			expectedMargin: "margin: 2cm",
		},
		{
			name: "custom margin 1cm",
			options: PageOptions{
				Columns:     1,
				Orientation: "portrait",
				Margin:      "1cm",
				Zoom:        100,
			},
			expectedMargin: "margin: 1cm",
		},
		{
			name: "custom margin 20mm",
			options: PageOptions{
				Columns:     1,
				Orientation: "portrait",
				Margin:      "20mm",
				Zoom:        100,
			},
			expectedMargin: "margin: 20mm",
		},
		{
			name: "per-edge margins",
			options: PageOptions{
				Columns:      1,
				Orientation:  "portrait",
				Margin:       "2cm",
				MarginTop:    "3cm",
				MarginRight:  "1.5cm",
				MarginBottom: "2.5cm",
				MarginLeft:   "2cm",
				Zoom:         100,
			},
			expectedMargin: "margin: (top: 3cm, right: 1.5cm, bottom: 2.5cm, left: 2cm)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertMarkdownToTypst(markdown, tt.options)
			if err != nil {
				t.Fatalf("convertMarkdownToTypst failed: %v", err)
			}

			if !strings.Contains(result, tt.expectedMargin) {
				t.Errorf("Expected margin '%s' not found in output:\n%s", tt.expectedMargin, result)
			}
		})
	}
}

func TestTypstConverterWithCustomZoom(t *testing.T) {
	markdown := []byte("# Test\n\nSome text")

	tests := []struct {
		name             string
		options          PageOptions
		expectedFontSize string
	}{
		{
			name: "default zoom 100%",
			options: PageOptions{
				Columns:     1,
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        100,
			},
			expectedFontSize: "size: 11.00pt",
		},
		{
			name: "zoom 80%",
			options: PageOptions{
				Columns:     1,
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        80,
			},
			expectedFontSize: "size: 8.80pt",
		},
		{
			name: "zoom 120%",
			options: PageOptions{
				Columns:     1,
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        120,
			},
			expectedFontSize: "size: 13.20pt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertMarkdownToTypst(markdown, tt.options)
			if err != nil {
				t.Fatalf("convertMarkdownToTypst failed: %v", err)
			}

			if !strings.Contains(result, tt.expectedFontSize) {
				t.Errorf("Expected font size '%s' not found in output:\n%s", tt.expectedFontSize, result)
			}
		})
	}
}

func TestHTMLConverterWithCustomMargin(t *testing.T) {
	markdown := []byte("# Test\n\nSome text")

	tests := []struct {
		name           string
		options        PageOptions
		expectedMargin string
	}{
		{
			name: "default margin",
			options: PageOptions{
				Columns:     1,
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        100,
			},
			expectedMargin: "margin: 2cm",
		},
		{
			name: "custom margin 1in",
			options: PageOptions{
				Columns:     1,
				Orientation: "portrait",
				Margin:      "1in",
				Zoom:        100,
			},
			expectedMargin: "margin: 1in",
		},
		{
			name: "per-edge margins",
			options: PageOptions{
				Columns:      1,
				Orientation:  "portrait",
				Margin:       "1in",
				MarginTop:    "3cm",
				MarginRight:  "1.25in",
				MarginBottom: "2.5cm",
				MarginLeft:   "2cm",
				Zoom:         100,
			},
			expectedMargin: "margin: 3cm 1.25in 2.5cm 2cm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertMarkdownToHTML(markdown, tt.options)
			if err != nil {
				t.Fatalf("convertMarkdownToHTML failed: %v", err)
			}

			resultStr := string(result)
			if !strings.Contains(resultStr, tt.expectedMargin) {
				t.Errorf("Expected margin '%s' not found in output", tt.expectedMargin)
			}
		})
	}
}

func TestHTMLConverterWithCustomZoom(t *testing.T) {
	markdown := []byte("# Test\n\nSome text")

	tests := []struct {
		name        string
		options     PageOptions
		expectedCSS string
	}{
		{
			name: "default zoom 100%",
			options: PageOptions{
				Columns:     1,
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        100,
			},
			expectedCSS: "font-size: 100%",
		},
		{
			name: "zoom 80%",
			options: PageOptions{
				Columns:     1,
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        80,
			},
			expectedCSS: "font-size: 80%",
		},
		{
			name: "zoom 120%",
			options: PageOptions{
				Columns:     1,
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        120,
			},
			expectedCSS: "font-size: 120%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertMarkdownToHTML(markdown, tt.options)
			if err != nil {
				t.Fatalf("convertMarkdownToHTML failed: %v", err)
			}

			resultStr := string(result)
			if !strings.Contains(resultStr, tt.expectedCSS) {
				t.Errorf("Expected CSS '%s' not found in output", tt.expectedCSS)
			}
		})
	}
}

func TestBodyCSSNoMaxWidth(t *testing.T) {
	markdown := []byte("# Test")

	options := PageOptions{
		Columns:     1,
		Orientation: "portrait",
		Margin:      "0cm",
		Zoom:        100,
	}

	result, err := convertMarkdownToHTML(markdown, options)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	resultStr := string(result)

	// Check that body no longer has max-width
	if strings.Contains(resultStr, "max-width: 800px") {
		t.Errorf("Body should not have max-width constraint when using custom margins")
	}

	// Check that body no longer has auto margins
	if strings.Contains(resultStr, "margin: 40px auto") {
		t.Errorf("Body should not have auto margins when using custom margins")
	}
}
