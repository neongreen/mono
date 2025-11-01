package symbols

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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

// ExtractAllSymbols extracts all symbol names (exported and unexported) from a file without loading the full package.
// This is a lightweight version that doesn't require a pre-loaded package.
// It returns just the symbol names, including both exported and unexported declarations.
func ExtractAllSymbols(filePath string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var symbolNames []string

	// Inspect all declarations in the file
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Function or method (exported and unexported)
			symbolNames = append(symbolNames, d.Name.Name)

		case *ast.GenDecl:
			// Type, const, or var declaration
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					// Type declaration (struct, interface, etc.)
					symbolNames = append(symbolNames, s.Name.Name)

				case *ast.ValueSpec:
					// Const or var declaration
					for _, name := range s.Names {
						symbolNames = append(symbolNames, name.Name)
					}
				}
			}
		}
	}

	return symbolNames, nil
}

// ExtractExportedSymbols extracts exported symbol names from a file without loading the full package.
// This is a lightweight version of FindExportedSymbols that doesn't require a pre-loaded package.
// It returns just the symbol names, which is sufficient for reference qualification.
func ExtractExportedSymbols(filePath string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var symbolNames []string

	// Inspect all declarations in the file
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Function or method
			if d.Name.IsExported() {
				symbolNames = append(symbolNames, d.Name.Name)
			}

		case *ast.GenDecl:
			// Type, const, or var declaration
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					// Type declaration (struct, interface, etc.)
					if s.Name.IsExported() {
						symbolNames = append(symbolNames, s.Name.Name)
					}

				case *ast.ValueSpec:
					// Const or var declaration
					for _, name := range s.Names {
						if name.IsExported() {
							symbolNames = append(symbolNames, name.Name)
						}
					}
				}
			}
		}
	}

	return symbolNames, nil
}
