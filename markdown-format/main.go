package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
)

// Common abbreviations that should not be treated as sentence boundaries
var commonAbbreviations = []string{"e.g.", "i.e.", "etc.", "vs.", "cf.", "ex.", "viz.", "approx.", "ca."}

// sentenceBoundaryRegex matches sentence boundaries (. ! ?) followed by space and uppercase letter
var sentenceBoundaryRegex = regexp.MustCompile(`([.!?])(\s+)([A-Z])`)

func main() {
	checkFlag := flag.Bool("check", false, "check if files are formatted without modifying them")
	flag.Parse()

	args := flag.Args()

	// If no arguments, read from stdin and write to stdout
	if len(args) == 0 {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
		output, err := formatMarkdown(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting markdown: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(string(output))
		return
	}

	// Handle file(s) - in-place by default
	hasErrors := false
	for _, filename := range args {
		input, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
			os.Exit(1)
		}
		output, err := formatMarkdown(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting %s: %v\n", filename, err)
			os.Exit(1)
		}

		if *checkFlag {
			// Check mode: compare without modifying
			if !bytes.Equal(input, output) {
				fmt.Fprintf(os.Stderr, "%s: not formatted\n", filename)
				hasErrors = true
			}
		} else {
			// Default: write in-place
			err = os.WriteFile(filename, output, 0o644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file %s: %v\n", filename, err)
				os.Exit(1)
			}
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}
