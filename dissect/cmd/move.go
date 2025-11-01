package main

// This file implements the "move" command for selectively moving Go declarations
// between files. It uses two different approaches depending on the declaration type:
//
// - Functions: Uses gopls's refactor.extract.toNewFile (via pkg/gopls)
// - Types/Interfaces/Consts/Vars: Manual AST manipulation (moveDeclarationManually)
//
// For detailed design rationale and limitations, see DESIGN.md.

import (
	"bytes"
	"fmt"
	"github.com/neongreen/mono/dissect/pkg/commands"
	"github.com/neongreen/mono/dissect/pkg/gopls"
	"github.com/neongreen/mono/dissect/pkg/goutils"
	"github.com/neongreen/mono/dissect/pkg/refactor"
	"go/ast"
	"go/printer"
	"go/token"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move <source> <target>",
	Short: "Move/rename files or move specific identifiers between files",
	Long: `Move supports two modes of operation:

1. File mode: Move or rename entire files with refactoring support
   dissect move source.go destination.go
   - Moves the physical file
   - Updates package declaration to match new directory
   - Updates all import statements in the codebase
   - Updates package qualifiers in code (e.g., old.Func() → new.Func())

2. Symbol mode: Move specific identifiers (functions, types, interfaces) between files
   dissect move source.go:Foo target.go
   dissect move source.go:Foo,Bar,Baz target.go

Symbol mode supports glob patterns:
  - File patterns: *.go, pkg/**/*.go
  - Identifier patterns: Test*, *Helper, Benchmark*

Glob behavior:
  - If a file doesn't contain a matching identifier, it's skipped (no error)
  - An error is only shown if no identifiers match across all files
  - File globs expand first, then identifier globs match within each file

Examples:
  # File mode
  dissect move source.go destination.go
  dissect move tk/admin_cmd.go tk/cmd/admin.go

  # Symbol mode
  dissect move source.go:Foo source.go:Bar target.go
  dissect move *.go:Helper target.go
  dissect move pkg/**/*.go:Test* target.go
  dissect move file.go:*Helper,Test* target.go
  dissect move source.go:MyType,MyInterface target.go`,
	Args: cobra.MinimumNArgs(2),
	Run:  runMove,
}

func runMove(cmd *cobra.Command, args []string) {
	// Check for required dependencies
	if err := checkDependencies(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Detect mode: file mode (no colons) vs symbol mode (colons present)
	hasColons := false
	for _, arg := range args {
		if strings.Contains(arg, ":") {
			hasColons = true
			break
		}
	}

	// File mode: move/rename entire file
	if !hasColons {
		if len(args) != 2 {
			slog.Error("File mode requires exactly 2 arguments", "args", len(args))
			fmt.Fprintf(os.Stderr, "Error: File mode requires exactly 2 arguments: <source> <target>\n")
			fmt.Fprintf(os.Stderr, "Got %d arguments. Use 'file:identifier' format to move symbols.\n", len(args))
			os.Exit(1)
		}

		sourceFile := args[0]
		targetFile := args[1]

		// Get absolute paths
		absSourceFile, err := filepath.Abs(sourceFile)
		if err != nil {
			slog.Error("Error getting absolute path", "file", sourceFile, "error", err)
			fmt.Fprintf(os.Stderr, "Error: Cannot resolve source file path: %v\n", err)
			os.Exit(1)
		}

		absTargetFile, err := filepath.Abs(targetFile)
		if err != nil {
			slog.Error("Error getting absolute path", "file", targetFile, "error", err)
			fmt.Fprintf(os.Stderr, "Error: Cannot resolve target file path: %v\n", err)
			os.Exit(1)
		}

		// Check if source file exists
		if _, err := os.Stat(absSourceFile); os.IsNotExist(err) {
			slog.Error("Source file does not exist", "file", absSourceFile)
			fmt.Fprintf(os.Stderr, "Error: Source file does not exist: %s\n", sourceFile)
			os.Exit(1)
		}

		// Find module root
		moduleRoot, err := commands.FindGoModuleRoot(absSourceFile)
		if err != nil {
			slog.Error("Error finding Go module root", "error", err, "file", absSourceFile)
			fmt.Fprintf(os.Stderr, "Error: Cannot find Go module root: %v\n", err)
			os.Exit(1)
		}

		// Use refactor.MoveFileWithImportUpdates to move the file and update imports
		slog.Info("Moving file with import updates", "from", sourceFile, "to", targetFile)
		if err := refactor.MoveFileWithImportUpdates(absSourceFile, absTargetFile, moduleRoot); err != nil {
			slog.Error("Error moving file", "error", err)
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		slog.Info("Successfully moved file and updated imports", "from", sourceFile, "to", targetFile)
		return
	}

	// Symbol mode: move specific identifiers between files
	// Parse arguments
	// Last argument is the target file
	targetFile := args[len(args)-1]

	// All other arguments are source specifications (file:identifier or file:id1,id2,id3)
	sourceSpecs := args[:len(args)-1]

	// Parse source specifications into a map of file -> []identifier patterns
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
		identifierPatterns := parts[1]

		// Expand file glob pattern using doublestar for ** support
		matches, err := doublestar.FilepathGlob(sourcePattern)
		if err != nil {
			slog.Error("Invalid glob pattern", "pattern", sourcePattern, "error", err)
			fmt.Fprintf(os.Stderr, "Error: Invalid glob pattern '%s': %v\n", sourcePattern, err)
			os.Exit(1)
		}

		// If no matches, treat it as a literal filename
		if len(matches) == 0 {
			matches = []string{sourcePattern}
		}

		// Split identifier patterns by comma
		idPatterns := strings.Split(identifierPatterns, ",")
		for _, sourceFile := range matches {
			for _, pattern := range idPatterns {
				pattern = strings.TrimSpace(pattern)
				if pattern != "" {
					sourceMap[sourceFile] = append(sourceMap[sourceFile], pattern)
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

	// Track total matches across all files
	totalMatches := 0

	// Process each source file
	for sourceFile, identifierPatterns := range sourceMap {
		absSourceFile, err := filepath.Abs(sourceFile)
		if err != nil {
			slog.Error("Error getting absolute path", "file", sourceFile, "error", err)
			fmt.Fprintf(os.Stderr, "Error: Cannot resolve source file path: %v\n", err)
			os.Exit(1)
		}

		// Check if source file exists
		if _, err := os.Stat(absSourceFile); os.IsNotExist(err) {
			slog.Warn("Source file does not exist, skipping", "file", absSourceFile)
			continue
		}

		// Find module root
		moduleRoot, err := commands.FindGoModuleRoot(absSourceFile)
		if err != nil {
			slog.Error("Error finding Go module root", "error", err, "file", absSourceFile)
			fmt.Fprintf(os.Stderr, "Error: Cannot find Go module root: %v\n", err)
			os.Exit(1)
		}

		// Find matching identifiers in this file
		matchingIdentifiers, err := findMatchingIdentifiers(absSourceFile, identifierPatterns)
		if err != nil {
			slog.Error("Error finding identifiers", "file", sourceFile, "error", err)
			fmt.Fprintf(os.Stderr, "Error finding identifiers in '%s': %v\n", sourceFile, err)
			os.Exit(1)
		}

		if len(matchingIdentifiers) == 0 {
			slog.Debug("No matching identifiers in file", "file", sourceFile, "patterns", identifierPatterns)
			continue
		}

		totalMatches += len(matchingIdentifiers)

		// Move each matched identifier
		for _, identifier := range matchingIdentifiers {
			slog.Info("Moving identifier", "identifier", identifier, "from", sourceFile, "to", targetFile)

			err := moveIdentifier(absSourceFile, identifier, absTargetFile, moduleRoot)
			if err != nil {
				slog.Error("Error moving identifier", "identifier", identifier, "error", err)
				fmt.Fprintf(os.Stderr, "Error moving identifier '%s': %v\n", identifier, err)
				os.Exit(1)
			}
		}
	}

	if totalMatches == 0 {
		slog.Error("No matching identifiers found")
		fmt.Fprintln(os.Stderr, "Error: No identifiers matched the specified patterns")
		os.Exit(1)
	}

	slog.Info("Successfully moved all identifiers", "count", totalMatches)
}

// findMatchingIdentifiers finds all declaration names (functions, types, interfaces, etc.) in a file that match any of the given patterns
func findMatchingIdentifiers(filePath string, patterns []string) ([]string, error) {
	// Parse the file to get all declarations
	_, node, err := goutils.ReadGoFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Compile glob patterns
	var globs []glob.Glob
	for _, pattern := range patterns {
		g, err := glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern '%s': %w", pattern, err)
		}
		globs = append(globs, g)
	}

	// Find all matching declaration names
	var matches []string
	seen := make(map[string]bool)
	for _, decl := range node.Decls {
		var names []string

		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Function or method declaration
			names = append(names, d.Name.Name)
		case *ast.GenDecl:
			// General declaration (type, const, var)
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					// Type or interface declaration
					names = append(names, s.Name.Name)
				case *ast.ValueSpec:
					// Const or var declaration
					for _, name := range s.Names {
						names = append(names, name.Name)
					}
				}
			}
		}

		// Check if any of the names match any pattern
		for _, name := range names {
			for _, g := range globs {
				if g.Match(name) && !seen[name] {
					matches = append(matches, name)
					seen[name] = true
					break
				}
			}
		}
	}

	return matches, nil
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

	// Find the declaration in the source file
	sourceFset, declNode, err := goutils.FindDecl(sourceFile, identifier)
	if err != nil {
		return fmt.Errorf("error finding declaration: %w", err)
	}

	// Check if this is a function - if so, use gopls for extraction
	if _, isFuncDecl := declNode.(*ast.FuncDecl); isFuncDecl {
		return moveFunctionWithGopls(sourceFile, identifier, targetFile, moduleRoot)
	}

	// For non-function declarations (types, interfaces, consts, vars), do manual extraction
	return moveDeclarationManually(sourceFile, identifier, targetFile, sourceFset, declNode)
}

// moveFunctionWithGopls moves a function using gopls's extract refactoring
func moveFunctionWithGopls(sourceFile string, identifier string, targetFile string, moduleRoot string) error {
	// Use gopls to extract the function to a new file first
	tempFile, err := gopls.ExtractToNewFile(sourceFile, identifier, moduleRoot)
	if err != nil {
		return fmt.Errorf("error extracting identifier: %w", err)
	}
	slog.Debug("Extracted identifier to temp file", "identifier", identifier, "tempFile", tempFile)

	// Remove the temp file when done
	defer os.Remove(tempFile)

	// Parse the temp file to extract the function (with comments preserved by gopls)
	tempFset, tempNode, err := goutils.ReadGoFile(tempFile)
	if err != nil {
		return fmt.Errorf("error parsing temp file: %w", err)
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

	// Extract imports from temp file for later merging
	var tempImports []*ast.ImportSpec
	for _, imp := range tempNode.Imports {
		tempImports = append(tempImports, imp)
	}

	// Serialize just the function declaration (not a whole file) using printer
	// This gives us the function with its Doc comments as source text
	var funcBuf bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	if err := cfg.Fprint(&funcBuf, tempFset, funcToMove); err != nil {
		return fmt.Errorf("error serializing function: %w", err)
	}

	// Read the current target file content
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		return fmt.Errorf("error reading target file: %w", err)
	}

	// Append the serialized function to the target file
	// printer.Fprint on a FuncDecl includes the Doc comments automatically
	newContent := string(targetContent) + "\n" + funcBuf.String() + "\n"
	if err := os.WriteFile(targetFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("error writing target file: %w", err)
	}

	// Now reparse the entire target file to get a proper AST
	targetFset, targetNode, err := goutils.ReadGoFile(targetFile)
	if err != nil {
		return fmt.Errorf("error reparsing target file: %w", err)
	}

	// Merge imports from temp file using AST operations
	existingImports := make(map[string]bool)
	for _, imp := range targetNode.Imports {
		existingImports[imp.Path.Value] = true
	}

	for _, imp := range tempImports {
		if !existingImports[imp.Path.Value] {
			targetNode.Imports = append(targetNode.Imports, imp)
		}
	}

	// Write back the target file with merged imports
	if err := goutils.WriteGoFile(targetFile, targetFset, targetNode); err != nil {
		return fmt.Errorf("error writing target file with imports: %w", err)
	}

	// Run goimports to organize imports and format properly
	err = commands.RunGoimportsOnFile(targetFile)
	if err != nil {
		return fmt.Errorf("error running goimports: %w", err)
	}

	slog.Debug("Merged function into target file", "identifier", identifier, "target", targetFile)

	return nil
}

// moveDeclarationManually moves a non-function declaration (type, interface, const, var) manually
// using AST manipulation. This approach is used because gopls doesn't offer code actions for
// these declaration types. Uses goimports for import management.
// See DESIGN.md for detailed explanation and known limitations.
func moveDeclarationManually(sourceFile string, identifier string, targetFile string, sourceFset *token.FileSet, declNode ast.Node) error {
	// Read source file
	sourceFileSet, sourceNode, err := goutils.ReadGoFile(sourceFile)
	if err != nil {
		return fmt.Errorf("error reading source file: %w", err)
	}

	// Find and extract the declaration to move
	var declToMove ast.Decl
	var declIndex int = -1

	for i, decl := range sourceNode.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		// Check if this GenDecl contains our identifier
		for _, spec := range genDecl.Specs {
			found := false
			switch s := spec.(type) {
			case *ast.TypeSpec:
				if s.Name.Name == identifier {
					found = true
				}
			case *ast.ValueSpec:
				for _, name := range s.Names {
					if name.Name == identifier {
						found = true
						break
					}
				}
			}

			if found {
				declToMove = genDecl
				declIndex = i
				break
			}
		}

		if declToMove != nil {
			break
		}
	}

	if declToMove == nil {
		return fmt.Errorf("declaration not found in source file")
	}

	// Serialize the declaration with comments using printer
	var declBuf bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	if err := cfg.Fprint(&declBuf, sourceFileSet, declToMove); err != nil {
		return fmt.Errorf("error serializing declaration: %w", err)
	}

	// Read the current target file content
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		return fmt.Errorf("error reading target file: %w", err)
	}

	// Append the serialized declaration to the target file
	newContent := string(targetContent) + "\n" + declBuf.String() + "\n"
	if err := os.WriteFile(targetFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("error writing target file: %w", err)
	}

	// Remove the declaration from the source file
	sourceNode.Decls = append(sourceNode.Decls[:declIndex], sourceNode.Decls[declIndex+1:]...)

	// Write back the source file
	if err := goutils.WriteGoFile(sourceFile, sourceFileSet, sourceNode); err != nil {
		return fmt.Errorf("error writing source file: %w", err)
	}

	// Run goimports on both files to organize imports and format properly
	if err := commands.RunGoimportsOnFile(targetFile); err != nil {
		return fmt.Errorf("error running goimports on target: %w", err)
	}
	if err := commands.RunGoimportsOnFile(sourceFile); err != nil {
		return fmt.Errorf("error running goimports on source: %w", err)
	}

	slog.Debug("Moved declaration to target file", "identifier", identifier, "target", targetFile)

	return nil
}
