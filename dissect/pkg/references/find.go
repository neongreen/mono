package references

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/packages"
)

// Reference represents a use of a symbol.
type Reference struct {
	File      string     // file containing the reference
	Pos       token.Pos  // position in file
	Ident     *ast.Ident // the identifier node
	Qualified bool       // whether it's already qualified (e.g., pkg.Symbol)
}

// FindReferences finds all references to given symbols in packages.
// It uses type information to accurately identify symbol references.
func FindReferences(symbolNames []string, pkgs []*packages.Package) ([]Reference, error) {
	// Create a set of symbol names for fast lookup
	symbolSet := make(map[string]bool)
	for _, name := range symbolNames {
		symbolSet[name] = true
	}

	var refs []Reference

	// Iterate through all packages
	for _, pkg := range pkgs {
		// Iterate through all files in the package
		for i, file := range pkg.Syntax {
			filePath := pkg.CompiledGoFiles[i]

			// Walk the AST to find identifier uses
			ast.Inspect(file, func(n ast.Node) bool {
				ident, ok := n.(*ast.Ident)
				if !ok {
					return true
				}

				// Check if this identifier matches one of our target symbols
				if !symbolSet[ident.Name] {
					return true
				}

				// Check if this is a definition (we want uses, not definitions)
				if pkg.TypesInfo != nil && pkg.TypesInfo.Defs[ident] != nil {
					// This is a definition, skip it
					return true
				}

				// Determine if this is a qualified reference
				qualified := false

				// Check if this identifier is part of a selector expression
				// We need to look at the parent node, but ast.Inspect doesn't give us that
				// So we'll check if the identifier is on the right side of a selector
				// by looking at the file's AST structure

				// For now, we'll use a simple heuristic: check if TypesInfo.Uses
				// points to an object in a different package
				if pkg.TypesInfo != nil && pkg.TypesInfo.Uses[ident] != nil {
					obj := pkg.TypesInfo.Uses[ident]
					// If the object's package is different from current package, it's qualified
					if obj.Pkg() != nil && obj.Pkg() != pkg.Types {
						qualified = true
					}
				}

				// Add this reference
				refs = append(refs, Reference{
					File:      filePath,
					Pos:       ident.Pos(),
					Ident:     ident,
					Qualified: qualified,
				})

				return true
			})
		}
	}

	return refs, nil
}
