package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/neongreen/mono/dissect/pkg/commands"
	"github.com/spf13/cobra"
)

// Symbol represents a Go symbol with its documentation
type Symbol struct {
	Name string `json:"name"`
	Kind string `json:"kind"` // "func", "type", "const", "var", "method"
	Doc  string `json:"doc"`  // Documentation comment
	Line int    `json:"line"` // Line number in the file
}

// FileSymbols represents all symbols in a single file
type FileSymbols struct {
	FilePath string   `json:"file_path"`
	Package  string   `json:"package"`
	Symbols  []Symbol `json:"symbols"`
}

var listCmd = &cobra.Command{
	Use:   "list [paths]...",
	Short: "List all symbols in Go files with their documentation",
	Long: `List command scans Go files and extracts all symbols (functions, types, methods, etc.)
along with their documentation comments. The output is formatted as JSON.

If no paths are provided, it scans all Go files in the current directory and subdirectories.

Examples:
  dissect list                    # List symbols in all Go files
  dissect list ./pkg              # List symbols in pkg directory
  dissect list file1.go file2.go # List symbols in specific files`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no args provided, default to current directory
		if len(args) == 0 {
			args = []string{"."}
		}

		// Get module root
		moduleRoot, err := commands.FindGoModuleRoot(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding Go module root: %v\n", err)
			os.Exit(1)
		}

		// Collect all files to process
		var filesToProcess []string
		for _, path := range args {
			absPath, err := filepath.Abs(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting absolute path for %s: %v\n", path, err)
				continue
			}

			info, err := os.Stat(absPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error accessing path %s: %v\n", absPath, err)
				continue
			}

			if info.IsDir() {
				// Find all Go files in directory
				goFiles, err := commands.FindGoFiles(absPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error finding Go files in %s: %v\n", absPath, err)
					continue
				}
				filesToProcess = append(filesToProcess, goFiles...)
			} else if strings.HasSuffix(info.Name(), ".go") {
				filesToProcess = append(filesToProcess, absPath)
			}
		}

		// Remove duplicates and sort
		uniqueFiles := make(map[string]struct{})
		for _, file := range filesToProcess {
			uniqueFiles[file] = struct{}{}
		}
		filesToProcess = []string{}
		for file := range uniqueFiles {
			filesToProcess = append(filesToProcess, file)
		}
		sort.Strings(filesToProcess)

		// Process each file
		var allFileSymbols []FileSymbols
		for _, filePath := range filesToProcess {
			fileSymbols, err := extractFileSymbols(filePath, moduleRoot)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", filePath, err)
				continue
			}
			allFileSymbols = append(allFileSymbols, fileSymbols)
		}

		// Output as JSON
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(allFileSymbols); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	},
}

// extractFileSymbols extracts all symbols from a Go file
func extractFileSymbols(filePath string, moduleRoot string) (FileSymbols, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return FileSymbols{}, fmt.Errorf("failed to parse file: %w", err)
	}

	// Make path relative to module root for cleaner output
	relPath, err := filepath.Rel(moduleRoot, filePath)
	if err != nil {
		relPath = filePath // Fall back to absolute path if relative path fails
	}

	result := FileSymbols{
		FilePath: relPath,
		Package:  file.Name.Name,
		Symbols:  []Symbol{},
	}

	// Inspect all declarations in the file
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Function or method
			kind := "func"
			if d.Recv != nil {
				kind = "method"
			}

			doc := ""
			if d.Doc != nil {
				doc = d.Doc.Text()
			}

			result.Symbols = append(result.Symbols, Symbol{
				Name: d.Name.Name,
				Kind: kind,
				Doc:  strings.TrimSpace(doc),
				Line: fset.Position(d.Pos()).Line,
			})

		case *ast.GenDecl:
			// Type, const, or var declaration
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					// Type declaration (struct, interface, etc.)
					doc := ""
					if s.Doc != nil {
						doc = s.Doc.Text()
					} else if d.Doc != nil {
						// Sometimes the doc is on the GenDecl instead
						doc = d.Doc.Text()
					}

					result.Symbols = append(result.Symbols, Symbol{
						Name: s.Name.Name,
						Kind: "type",
						Doc:  strings.TrimSpace(doc),
						Line: fset.Position(s.Pos()).Line,
					})

				case *ast.ValueSpec:
					// Const or var declaration
					kind := "var"
					if d.Tok.String() == "const" {
						kind = "const"
					}

					doc := ""
					if s.Doc != nil {
						doc = s.Doc.Text()
					} else if d.Doc != nil {
						doc = d.Doc.Text()
					}

					for _, name := range s.Names {
						result.Symbols = append(result.Symbols, Symbol{
							Name: name.Name,
							Kind: kind,
							Doc:  strings.TrimSpace(doc),
							Line: fset.Position(name.Pos()).Line,
						})
					}
				}
			}
		}
	}

	return result, nil
}
