package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/printpdf/pkg/converter"
	"github.com/neongreen/mono/printpdf/pkg/fetcher"
)

func main() {
	outputDir := flag.String("o", ".", "output directory for generated PDFs")
	converters := flag.String("converters", "all", "comma-separated list of converters to use (typst,prince,weasyprint) or 'all'")
	columns := flag.Int("columns", 1, "number of columns for text layout (e.g., 2 or 3 for newspaper-style layout)")
	orientation := flag.String("orientation", "portrait", "page orientation (portrait or landscape)")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: printpdf [options] <input>\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nInput can be:\n")
		fmt.Fprintf(os.Stderr, "  - Path to a Markdown file\n")
		fmt.Fprintf(os.Stderr, "  - URL to a Markdown file\n")
		fmt.Fprintf(os.Stderr, "  - GitHub file URL (works with private repos if GITHUB_TOKEN is set)\n")
		fmt.Fprintf(os.Stderr, "  - Web page URL (will be processed with Mozilla Readability)\n")
		os.Exit(1)
	}

	input := args[0]

	// Fetch the content
	fmt.Printf("Fetching content from: %s\n", input)
	content, contentType, err := fetcher.Fetch(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching content: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Content type: %s (%d bytes)\n", contentType, len(content))

	// Validate and prepare page options
	if *columns < 1 {
		fmt.Fprintf(os.Stderr, "Error: columns must be at least 1\n")
		os.Exit(1)
	}
	if *orientation != "portrait" && *orientation != "landscape" {
		fmt.Fprintf(os.Stderr, "Error: orientation must be 'portrait' or 'landscape'\n")
		os.Exit(1)
	}

	pageOptions := converter.PageOptions{
		Columns:     *columns,
		Orientation: *orientation,
	}

	// Determine which converters to use
	converterList := converter.ParseConverterList(*converters)
	if len(converterList) == 0 {
		fmt.Fprintf(os.Stderr, "No valid converters specified\n")
		os.Exit(1)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Convert with each converter
	success := false
	for _, conv := range converterList {
		fmt.Printf("\n--- Converting with %s ---\n", conv.Name())

		outputPath := filepath.Join(*outputDir, fmt.Sprintf("output-%s.pdf", conv.Name()))

		err := conv.Convert(content, contentType, outputPath, pageOptions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error with %s: %v\n", conv.Name(), err)
			continue
		}

		// Check file was created
		if info, err := os.Stat(outputPath); err == nil {
			fmt.Printf("✓ Generated %s (%d bytes)\n", outputPath, info.Size())
			success = true
		} else {
			fmt.Fprintf(os.Stderr, "✗ Failed to create %s\n", outputPath)
		}
	}

	if !success {
		fmt.Fprintf(os.Stderr, "\nAll converters failed\n")
		os.Exit(1)
	}

	fmt.Println("\nDone!")
}
