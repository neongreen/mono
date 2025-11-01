package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/neongreen/mono/printpdf/pkg/converter"
	"github.com/neongreen/mono/printpdf/pkg/fetcher"
	"github.com/neongreen/mono/printpdf/pkg/pdfutil"
)

var (
	outputDir      string
	converters     string
	columns        int
	orientation    string
	margin         string
	marginTop      string
	marginRight    string
	marginBottom   string
	marginLeft     string
	marginInner    string
	marginOuter    string
	zoom           int
	firstPageGuide string
	keepArtifacts  bool
)

var rootCmd = &cobra.Command{
	Use:   "printpdf <input>",
	Short: "Convert Markdown and HTML to PDF using various rendering engines",
	Long: `printpdf converts Markdown files, HTML files, and web pages to PDF format
using various rendering engines (Typst, Prince XML, WeasyPrint).

Input can be:
  - Path to a Markdown file
  - Path to an HTML file
  - URL to a Markdown file
  - GitHub file URL (works with private repos if GITHUB_TOKEN is set)
  - Web page URL (will be processed with Mozilla Readability)

Examples:
  printpdf README.md
  printpdf --converters weasyprint --zoom 120 README.md
  printpdf --margin-top 3cm --margin-left 4cm document.html
  printpdf https://github.com/user/repo/blob/main/README.md`,
	Args: cobra.ExactArgs(1),
	Run:  runConvert,
}

func init() {
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "output directory for generated PDFs")
	rootCmd.Flags().StringVar(&converters, "converters", "all", "comma-separated list of converters to use (typst,prince,weasyprint) or 'all'")
	rootCmd.Flags().IntVar(&columns, "columns", 1, "number of columns for text layout (e.g., 2 or 3 for newspaper-style layout)")
	rootCmd.Flags().StringVar(&orientation, "orientation", "portrait", "page orientation (portrait or landscape)")
	rootCmd.Flags().StringVar(&margin, "margin", "2cm", "page margin (e.g., '2cm', '1in', '20mm')")
	rootCmd.Flags().StringVar(&marginTop, "margin-top", "", "top page margin (overrides the value from --margin when set)")
	rootCmd.Flags().StringVar(&marginRight, "margin-right", "", "right page margin (overrides the value from --margin when set)")
	rootCmd.Flags().StringVar(&marginBottom, "margin-bottom", "", "bottom page margin (overrides the value from --margin when set)")
	rootCmd.Flags().StringVar(&marginLeft, "margin-left", "", "left page margin (overrides the value from --margin when set)")
	rootCmd.Flags().StringVar(&marginInner, "margin-inner", "", "inner margin for booklet printing (binding side; cannot be used with --margin-left or --margin-right)")
	rootCmd.Flags().StringVar(&marginOuter, "margin-outer", "", "outer margin for booklet printing (outer edge; cannot be used with --margin-left or --margin-right)")
	rootCmd.Flags().IntVar(&zoom, "zoom", 100, "zoom percentage for all font sizes (e.g., 80 for 80%, 120 for 120%)")
	rootCmd.Flags().StringVar(&firstPageGuide, "first-page-guide", "", "draw a thin vertical guide on the first page at the given distance from the left edge (e.g., '3cm')")
	rootCmd.Flags().BoolVar(&keepArtifacts, "keep-artifacts", false, "keep intermediate artifacts such as HTML and Typst sources next to the output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runConvert(cmd *cobra.Command, args []string) {
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
	if columns < 1 {
		fmt.Fprintf(os.Stderr, "Error: columns must be at least 1\n")
		os.Exit(1)
	}
	if orientation != "portrait" && orientation != "landscape" {
		fmt.Fprintf(os.Stderr, "Error: orientation must be 'portrait' or 'landscape'\n")
		os.Exit(1)
	}
	if zoom < 1 || zoom > 500 {
		fmt.Fprintf(os.Stderr, "Error: zoom must be between 1 and 500\n")
		os.Exit(1)
	}

	// Validate margin options: cannot use both left/right and inner/outer
	hasLeftRight := marginLeft != "" || marginRight != ""
	hasInnerOuter := marginInner != "" || marginOuter != ""
	if hasLeftRight && hasInnerOuter {
		fmt.Fprintf(os.Stderr, "Error: cannot specify both --margin-left/--margin-right and --margin-inner/--margin-outer at the same time\n")
		os.Exit(1)
	}

	pageOptions := converter.PageOptions{
		Columns:        columns,
		Orientation:    orientation,
		Margin:         margin,
		MarginTop:      marginTop,
		MarginRight:    marginRight,
		MarginBottom:   marginBottom,
		MarginLeft:     marginLeft,
		MarginInner:    marginInner,
		MarginOuter:    marginOuter,
		Zoom:           zoom,
		FirstPageGuide: strings.TrimSpace(firstPageGuide),
	}

	var artifactBase string
	if keepArtifacts {
		artifactBase = filepath.Join(outputDir, "printpdf-artifacts")
		if err := os.MkdirAll(artifactBase, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating artifact directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Determine which converters to use
	converterList := converter.ParseConverterList(converters)
	if len(converterList) == 0 {
		fmt.Fprintf(os.Stderr, "No valid converters specified\n")
		os.Exit(1)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Convert with each converter
	success := false
	for _, conv := range converterList {
		fmt.Printf("\n--- Converting with %s ---\n", conv.Name())

		outputPath := filepath.Join(outputDir, fmt.Sprintf("output-%s.pdf", conv.Name()))

		convOptions := pageOptions
		if keepArtifacts {
			convOptions.KeepIntermediates = true
			convOptions.IntermediateDir = filepath.Join(artifactBase, conv.Name())
			if err := os.MkdirAll(convOptions.IntermediateDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error preparing artifact directory for %s: %v\n", conv.Name(), err)
				os.Exit(1)
			}
		}

		err := conv.Convert(content, contentType, outputPath, convOptions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error with %s: %v\n", conv.Name(), err)
			continue
		}

		// Check file was created and get page count
		if info, err := os.Stat(outputPath); err == nil {
			pageCount, pageErr := pdfutil.CountPages(outputPath)
			if pageErr == nil {
				fmt.Printf("✓ Generated %s (%d bytes, %d pages)\n", outputPath, info.Size(), pageCount)
			} else {
				fmt.Printf("✓ Generated %s (%d bytes)\n", outputPath, info.Size())
			}
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
