package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/neongreen/mono/lib/version"
	"github.com/neongreen/mono/lion/internal/extractor"
	"github.com/neongreen/mono/lion/internal/generator"
)

var rootCmd = &cobra.Command{
	Use:   "lion",
	Short: "Documentation extraction tool for Go code",
	Long: `lion extracts documentation from special comments in Go code and generates markdown files.

Use //lion:topic-name comments in your Go code to mark documentation for specific topics.
Run 'lion generate' to create markdown documentation files organized by topic.`,
}

// To run generation:
//
// 1. Navigate to your project directory
// 2. Run: lion generate
// 3. Check ./docs for generated markdown files
//
// You can specify a different directory and output location:
//
//	lion generate ./myproject --output ./documentation
//
//lion:how-to-run-generation section="Run generation"
var generateCmd = &cobra.Command{
	Use:   "generate [directory]",
	Short: "Generate markdown documentation from Go code",
	Long: `Generate markdown documentation by extracting lion comments from Go source files.

The command scans Go files in the specified directory (default: current directory)
and generates markdown files for each documentation topic found.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGenerate,
}

var topicsCmd = &cobra.Command{
	Use:   "topics [directory]",
	Short: "List all documentation topics found in Go code",
	Long:  `Scan Go source files and list all unique documentation topics.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runTopics,
}

var (
	outputDir string
)

func init() {
	generateCmd.Flags().StringVarP(&outputDir, "output", "o", "./docs", "Output directory for generated markdown files")

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(topicsCmd)
	rootCmd.AddCommand(version.NewVersionCommand("lion"))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runGenerate(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	// Extract documentation from Go files
	docs, err := extractor.Extract(dir)
	if err != nil {
		return fmt.Errorf("failed to extract documentation: %w", err)
	}

	// Generate markdown files
	if err := generator.Generate(docs, outputDir); err != nil {
		return fmt.Errorf("failed to generate markdown files: %w", err)
	}

	fmt.Printf("Generated documentation in %s\n", outputDir)
	return nil
}

func runTopics(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	docs, err := extractor.Extract(dir)
	if err != nil {
		return fmt.Errorf("failed to extract documentation: %w", err)
	}

	if len(docs) == 0 {
		fmt.Println("No documentation topics found")
		return nil
	}

	fmt.Println("Documentation topics:")
	for topic := range docs {
		fmt.Printf("  - %s\n", topic)
	}

	return nil
}
