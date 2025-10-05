package main

import (
	"dissect/pkg/commands"
	"dissect/pkg/gopls"
	"dissect/pkg/goutils"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move <source_file:identifier> [identifiers...] <target_file>",
	Short: "Move specific identifiers (functions) to a target file",
	Long: `Move extracts specific identifiers (functions, methods) from a source file 
and moves them to a target file. The target file will be created if it doesn't exist.

Example:
  dissect move source.go:Foo source.go:Bar target.go
  dissect move source.go:Foo,Bar,Baz target.go`,
	Args: cobra.MinimumNArgs(2),
	Run:  runMove,
}

func runMove(cmd *cobra.Command, args []string) {
	// Parse arguments
	// Last argument is the target file
	targetFile := args[len(args)-1]

	// All other arguments are source specifications (file:identifier or file:id1,id2,id3)
	sourceSpecs := args[:len(args)-1]

	// Parse source specifications into a map of file -> []identifiers
	sourceMap := make(map[string][]string)
	for _, spec := range sourceSpecs {
		// Split by colon
		parts := strings.SplitN(spec, ":", 2)
		if len(parts) != 2 {
			slog.Error("Invalid source specification", "spec", spec)
			fmt.Fprintf(os.Stderr, "Error: Invalid source specification '%s'. Expected format: file:identifier or file:id1,id2\n", spec)
			os.Exit(1)
		}

		sourceFile := parts[0]
		identifiers := parts[1]

		// Split identifiers by comma
		ids := strings.Split(identifiers, ",")
		for _, id := range ids {
			id = strings.TrimSpace(id)
			if id != "" {
				sourceMap[sourceFile] = append(sourceMap[sourceFile], id)
			}
		}
	}

	if len(sourceMap) == 0 {
		slog.Error("No identifiers to move")
		fmt.Fprintln(os.Stderr, "Error: No identifiers specified")
		os.Exit(1)
	}

	// Get absolute paths
	absTargetFile, err := filepath.Abs(targetFile)
	if err != nil {
		slog.Error("Error getting absolute path", "file", targetFile, "error", err)
		fmt.Fprintf(os.Stderr, "Error: Cannot resolve target file path: %v\n", err)
		os.Exit(1)
	}

	// Process each source file
	for sourceFile, identifiers := range sourceMap {
		absSourceFile, err := filepath.Abs(sourceFile)
		if err != nil {
			slog.Error("Error getting absolute path", "file", sourceFile, "error", err)
			fmt.Fprintf(os.Stderr, "Error: Cannot resolve source file path: %v\n", err)
			os.Exit(1)
		}

		// Check if source file exists
		if _, err := os.Stat(absSourceFile); os.IsNotExist(err) {
			slog.Error("Source file does not exist", "file", absSourceFile)
			fmt.Fprintf(os.Stderr, "Error: Source file '%s' does not exist\n", sourceFile)
			os.Exit(1)
		}

		// Find module root
		moduleRoot, err := commands.FindGoModuleRoot(absSourceFile)
		if err != nil {
			slog.Error("Error finding Go module root", "error", err, "file", absSourceFile)
			fmt.Fprintf(os.Stderr, "Error: Cannot find Go module root: %v\n", err)
			os.Exit(1)
		}

		// Move each identifier
		for _, identifier := range identifiers {
			slog.Info("Moving identifier", "identifier", identifier, "from", sourceFile, "to", targetFile)

			err := moveIdentifier(absSourceFile, identifier, absTargetFile, moduleRoot)
			if err != nil {
				slog.Error("Error moving identifier", "identifier", identifier, "error", err)
				fmt.Fprintf(os.Stderr, "Error moving identifier '%s': %v\n", identifier, err)
				os.Exit(1)
			}
		}
	}

	slog.Info("Successfully moved all identifiers")
}

// moveIdentifier moves a single identifier from source to target file
func moveIdentifier(sourceFile string, identifier string, targetFile string, moduleRoot string) error {
	// Check if target file exists
	targetExists := false
	if _, err := os.Stat(targetFile); err == nil {
		targetExists = true
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking target file: %w", err)
	}

	// If target doesn't exist, we need to create it with the right package
	if !targetExists {
		// Get package name from source file
		_, node, err := goutils.ReadGoFile(sourceFile)
		if err != nil {
			return fmt.Errorf("error reading source file: %w", err)
		}
		packageName := node.Name.Name

		// Create target directory if needed
		targetDir := filepath.Dir(targetFile)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("error creating target directory: %w", err)
		}

		// Create target file with package declaration
		err = os.WriteFile(targetFile, []byte(fmt.Sprintf("package %s\n", packageName)), 0644)
		if err != nil {
			return fmt.Errorf("error creating target file: %w", err)
		}
		slog.Debug("Created target file", "file", targetFile, "package", packageName)
	}

	// Use gopls to extract the function to a new file first
	tempFile, err := gopls.ExtractToNewFile(sourceFile, identifier, moduleRoot)
	if err != nil {
		return fmt.Errorf("error extracting identifier: %w", err)
	}
	slog.Debug("Extracted identifier to temp file", "identifier", identifier, "tempFile", tempFile)

	// Now we need to move the content from tempFile to targetFile
	// Read the temp file
	tempContent, err := os.ReadFile(tempFile)
	if err != nil {
		return fmt.Errorf("error reading temp file: %w", err)
	}

	// Read target file to verify it exists
	_, err = os.ReadFile(targetFile)
	if err != nil {
		return fmt.Errorf("error reading target file: %w", err)
	}

	// For now, we'll use a simple approach: append the function to the target file
	// This is a simplified version - a production version would:
	// 1. Parse both files
	// 2. Merge imports properly
	// 3. Add the function declaration
	// 4. Format the result

	// Parse temp file content to extract function and imports
	tempStr := string(tempContent)

	// Remove the temp file
	defer os.Remove(tempFile)

	// Extract only the function declaration from temp file (skip package and imports)
	lines := strings.Split(tempStr, "\n")
	var functionLines []string
	skipImports := false
	inImportBlock := false
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Skip package declaration
		if strings.HasPrefix(trimmed, "package ") {
			continue
		}
		
		// Skip import statements
		if strings.HasPrefix(trimmed, "import (") {
			inImportBlock = true
			skipImports = true
			continue
		}
		if strings.HasPrefix(trimmed, "import ") && !inImportBlock {
			skipImports = true
			continue
		}
		if inImportBlock && trimmed == ")" {
			inImportBlock = false
			skipImports = false
			continue
		}
		if skipImports && inImportBlock {
			continue
		}
		
		// Skip empty lines at the beginning
		if len(functionLines) == 0 && trimmed == "" {
			continue
		}
		
		// Start collecting from the function declaration
		if strings.HasPrefix(trimmed, "func ") || len(functionLines) > 0 {
			functionLines = append(functionLines, line)
		}
		
		// Stop after collecting if we hit the next line after a function
		if len(functionLines) > 0 && i+1 < len(lines) && 
		   strings.TrimSpace(lines[i+1]) == "" && 
		   strings.HasPrefix(trimmed, "}") {
			break
		}
	}

	// Append to target file
	f, err := os.OpenFile(targetFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening target file for append: %w", err)
	}
	defer f.Close()

	// Add a blank line before the function
	_, err = f.WriteString("\n" + strings.Join(functionLines, "\n") + "\n")
	if err != nil {
		return fmt.Errorf("error appending to target file: %w", err)
	}

	// Run goimports to fix imports and formatting
	err = commands.RunGoimportsOnFile(targetFile)
	if err != nil {
		return fmt.Errorf("error running goimports: %w", err)
	}

	slog.Debug("Merged function into target file", "identifier", identifier, "target", targetFile)

	return nil
}
