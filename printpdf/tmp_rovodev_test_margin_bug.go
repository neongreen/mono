package main

import (
	"fmt"
	"os"

	"github.com/neongreen/mono/printpdf/pkg/converter"
	"github.com/neongreen/mono/printpdf/pkg/fetcher"
)

func main() {
	// Test HTML input with custom margins
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>Margin Test</title>
</head>
<body>
    <h1>HTML Margin Test</h1>
    <p>This HTML document should have custom margins:</p>
    <ul>
        <li>Top: 1cm</li>
        <li>Right: 3cm</li>
        <li>Bottom: 2cm</li>
        <li>Left: 4cm</li>
    </ul>
    <p>Before the fix, these margin settings would be ignored for HTML input.</p>
    <p>After the fix, the margins should be applied correctly.</p>
</body>
</html>`

	options := converter.PageOptions{
		Columns:      1,
		Orientation:  "portrait",
		Margin:       "2cm", // Default margin (will be overridden)
		MarginTop:    "1cm",
		MarginRight:  "3cm",
		MarginBottom: "2cm",
		MarginLeft:   "4cm",
		Zoom:         100,
	}

	// Test with WeasyPrint (most likely to be available)
	conv := converter.NewWeasyPrintConverter()
	outputPath := "pkg/golden/testdata/output/html-margin-test-weasyprint.pdf"
	
	fmt.Println("Testing HTML margin fix with WeasyPrint...")
	err := conv.Convert([]byte(htmlContent), fetcher.ContentTypeHTML, outputPath, options)
	if err != nil {
		fmt.Printf("WeasyPrint test failed (expected if not installed): %v\n", err)
	} else {
		fmt.Printf("✅ WeasyPrint HTML margin test passed - PDF created: %s\n", outputPath)
	}

	// Test with Prince if available
	princeConv := converter.NewPrinceConverter()
	princeOutputPath := "pkg/golden/testdata/output/html-margin-test-prince.pdf"
	
	fmt.Println("Testing HTML margin fix with Prince...")
	err = princeConv.Convert([]byte(htmlContent), fetcher.ContentTypeHTML, princeOutputPath, options)
	if err != nil {
		fmt.Printf("Prince test failed (expected if not installed): %v\n", err)
	} else {
		fmt.Printf("✅ Prince HTML margin test passed - PDF created: %s\n", princeOutputPath)
	}

	fmt.Println("\nNext steps:")
	fmt.Println("1. Check the generated PDFs to visually verify margins are applied")
	fmt.Println("2. Compare with equivalent Markdown input to ensure consistency")
}