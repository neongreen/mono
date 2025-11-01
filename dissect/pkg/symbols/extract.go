package symbols

import (
	"fmt"
	"go/ast"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// ExportedSymbol represents an exported declaration.
type ExportedSymbol struct {
	Name string // The name of the symbol (e.g., "Hello", "Person")
	Kind string // "func", "type", "var", "const", "method"
	Pkg  string // package path where the symbol is defined
}

// FindAllSymbols finds all symbols (exported and unexported) in a file.
// It examines the file within the context of the loaded package to get type information.
func FindAllSymbols(filePath string, pkg *packages.Package) ([]ExportedSymbol, error) {
	return findSymbols(filePath, pkg, false)
}

// FindExportedSymbols finds all exported symbols in a file.
// It examines the file within the context of the loaded package to get type information.
func FindExportedSymbols(filePath string, pkg *packages.Package) ([]ExportedSymbol, error) {
	return findSymbols(filePath, pkg, true)
}

// findSymbols finds symbols in a file, optionally filtering to only exported symbols.
func findSymbols(filePath string, pkg *packages.Package, exportedOnly bool) ([]ExportedSymbol, error) {
	// Get absolute path for comparison
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path: %w", err)
	}

	// Find the AST node for this specific file
	var fileNode *ast.File
	for i, f := range pkg.CompiledGoFiles {
		absCompiledPath, err := filepath.Abs(f)
		if err != nil {
			continue
		}
		if absCompiledPath == absFilePath {
			if i < len(pkg.Syntax) {
				fileNode = pkg.Syntax[i]
				break
			}
		}
	}

	if fileNode == nil {
		return nil, fmt.Errorf("file %s not found in package %s", filePath, pkg.PkgPath)
	}

	var symbols []ExportedSymbol

	// Inspect all declarations in the file
	for _, decl := range fileNode.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Function or method
			if !exportedOnly || d.Name.IsExported() {
				kind := "func"
				if d.Recv != nil {
					kind = "method"
				}
				symbols = append(symbols, ExportedSymbol{
					Name: d.Name.Name,
					Kind: kind,
					Pkg:  pkg.PkgPath,
				})
			}

		case *ast.GenDecl:
			// Type, const, or var declaration
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					// Type declaration (struct, interface, etc.)
					if !exportedOnly || s.Name.IsExported() {
						symbols = append(symbols, ExportedSymbol{
							Name: s.Name.Name,
							Kind: "type",
							Pkg:  pkg.PkgPath,
						})
					}

				case *ast.ValueSpec:
					// Const or var declaration
					kind := "var"
					if d.Tok.String() == "const" {
						kind = "const"
					}
					for _, name := range s.Names {
						if !exportedOnly || name.IsExported() {
							symbols = append(symbols, ExportedSymbol{
								Name: name.Name,
								Kind: kind,
								Pkg:  pkg.PkgPath,
							})
						}
					}
				}
			}
		}
	}

	return symbols, nil
}
