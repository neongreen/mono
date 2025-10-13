package converter

import (
"fmt"
"testing"
)

func TestPrintTypstOutput(t *testing.T) {
markdown := []byte(`# Test Document

## Section 1

This is body text to test zoom.

### Subsection 1.1

- Item 1
- Item 2
`)

// Test Typst with zoom 80%
options := PageOptions{
Columns:     1,
Orientation: "portrait",
Margin:      "1cm",
Zoom:        80,
}

typstOutput, err := convertMarkdownToTypst(markdown, options)
if err != nil {
t.Fatalf("Error: %v", err)
}

fmt.Println("\n=== Typst output with zoom 80% and margin 1cm ===")
fmt.Println(typstOutput)
}

func TestPrintHTMLOutput(t *testing.T) {
markdown := []byte(`# Test Document

## Section 1

This is body text to test zoom.
`)

// Test HTML with zoom 120%
options := PageOptions{
Columns:     1,
Orientation: "portrait",
Margin:      "3cm",
Zoom:        120,
}

htmlOutput, err := convertMarkdownToHTML(markdown, options)
if err != nil {
t.Fatalf("Error: %v", err)
}

fmt.Println("\n=== HTML output with zoom 120% and margin 3cm ===")
fmt.Println(string(htmlOutput[:1000]))
fmt.Println("...")
}
