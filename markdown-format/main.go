package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

var (
	commonAbbreviations = []string{"e.g.", "i.e.", "etc.", "vs.", "cf.", "ex.", "viz.", "approx.", "ca."}
	// sentenceBoundaryRegex matches sentence boundaries (. ! ?) followed by space and uppercase letter
	sentenceBoundaryRegex = regexp.MustCompile(`([.!?])(\s+)([A-Z])`)
	checkFlag             bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "markdown-format [files...]",
		Short: "Format Markdown text",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFormat(args)
		},
	}
	rootCmd.Flags().BoolVar(&checkFlag, "check", false, "check if files are formatted without modifying them")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runFormat(args []string) error {

	// If no arguments, read from stdin and write to stdout
	if len(args) == 0 {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading stdin: %w", err)
		}
		output, err := formatMarkdown(input)
		if err != nil {
			return fmt.Errorf("error formatting markdown: %w", err)
		}
		fmt.Print(string(output))
		return nil
	}

	// Handle file(s) - in-place by default
	hasErrors := false
	for _, filename := range args {
		input, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", filename, err)
		}
		output, err := formatMarkdown(input)
		if err != nil {
			return fmt.Errorf("error formatting %s: %w", filename, err)
		}

		if checkFlag {
			// Check mode: compare without modifying
			if !bytes.Equal(input, output) {
				fmt.Fprintf(os.Stderr, "%s: not formatted\n", filename)
				hasErrors = true
			}
		} else {
			// Default: write in-place
			err = os.WriteFile(filename, output, 0o644)
			if err != nil {
				return fmt.Errorf("error writing file %s: %w", filename, err)
			}
		}
	}

	if hasErrors {
		return fmt.Errorf("unformatted files")
	}
	return nil
}
