package main

import (
	"dissect/pkg/commands"
	"dissect/pkg/gopls"
	"dissect/pkg/goutils"
	"fmt"
	"go/ast"
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

Source files can be specified using glob patterns (e.g., *.go, pkg/**/*.go).

Example:
  dissect move source.go:Foo source.go:Bar target.go
  dissect move source.go:Foo,Bar,Baz target.go
  dissect move *.go:Helper target.go
  dissect move pkg/**/*.go:Utility target.go`,
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

		sourcePattern := parts[0]
		identifiers := parts[1]

		// Expand glob pattern
		matches, err := filepath.Glob(sourcePattern)
		if err != nil {
			slog.Error("Invalid glob pattern", "pattern", sourcePattern, "error", err)
			fmt.Fprintf(os.Stderr, "Error: Invalid glob pattern '%s': %v\n", sourcePattern, err)
			os.Exit(1)
		}

		// If no matches, treat it as a literal filename
		if len(matches) == 0 {
			matches = []string{sourcePattern}
		}

		// Split identifiers by comma
		ids := strings.Split(identifiers, ",")
		for _, sourceFile := range matches {
			for _, id := range ids {
				id = strings.TrimSpace(id)
				if id != "" {
					sourceMap[sourceFile] = append(sourceMap[sourceFile], id)
				}
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

// moveIdentifier moves a single identifier from source to target file using AST operations
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
		_, sourceNode, err := goutils.ReadGoFile(sourceFile)
		if err != nil {
			return fmt.Errorf("error reading source file: %w", err)
		}
		packageName := sourceNode.Name.Name

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

	// Remove the temp file when done
	defer os.Remove(tempFile)

	// Parse the temp file to extract the function declaration and imports using AST
	_, tempNode, err := goutils.ReadGoFile(tempFile)
	if err != nil {
		return fmt.Errorf("error parsing temp file: %w", err)
	}

	// Parse the target file
	targetFset, targetNode, err := goutils.ReadGoFile(targetFile)
	if err != nil {
		return fmt.Errorf("error parsing target file: %w", err)
	}

	// Find the function in the temp file
	var funcToMove *ast.FuncDecl
	for _, decl := range tempNode.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			funcToMove = fn
			break
		}
	}

	if funcToMove == nil {
		return fmt.Errorf("no function found in temp file")
	}

	// Merge imports from temp file to target file
	// Build a map of existing imports in target
	existingImports := make(map[string]bool)
	for _, imp := range targetNode.Imports {
		existingImports[imp.Path.Value] = true
	}

	// Add new imports from temp file
	for _, imp := range tempNode.Imports {
		if !existingImports[imp.Path.Value] {
			targetNode.Imports = append(targetNode.Imports, imp)
		}
	}

	// Add the function declaration to the target file
	targetNode.Decls = append(targetNode.Decls, funcToMove)

	// Write the modified target file back using AST
	err = goutils.WriteGoFile(targetFile, targetFset, targetNode)
	if err != nil {
		return fmt.Errorf("error writing target file: %w", err)
	}

	// Run goimports to organize imports and format properly
	err = commands.RunGoimportsOnFile(targetFile)
	if err != nil {
		return fmt.Errorf("error running goimports: %w", err)
	}

	slog.Debug("Merged function into target file", "identifier", identifier, "target", targetFile)

	return nil
}
