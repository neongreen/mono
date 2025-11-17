package references

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// Reference represents a use of a symbol.
type Reference struct {
	File      string     // file containing the reference
	Pos       token.Pos  // position in file
	Ident     *ast.Ident // the identifier node
	Qualified bool       // whether it's already qualified (e.g., pkg.Symbol)
	Selector  string     // full selector path (e.g., "DB.data" or "obj.field") for field/method accesses
}

// FindReferences finds all references to given symbols in packages.
// It uses type information to accurately identify symbol references.
// It tracks both simple identifiers and field/method accesses via selector expressions.
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

			// Walk the AST to find both identifier uses and selector expressions
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.Ident:
					// Check if this identifier matches one of our target symbols
					if !symbolSet[node.Name] {
						return true
					}

					// Check if this is a definition (we want uses, not definitions)
					if pkg.TypesInfo != nil && pkg.TypesInfo.Defs[node] != nil {
						// This is a definition, skip it
						return true
					}

					// Determine if this is a qualified reference
					qualified := false
					if pkg.TypesInfo != nil && pkg.TypesInfo.Uses[node] != nil {
						obj := pkg.TypesInfo.Uses[node]
						// If the object's package is different from current package, it's qualified
						if obj.Pkg() != nil && obj.Pkg() != pkg.Types {
							qualified = true
						}
					}

					// Add this reference
					refs = append(refs, Reference{
						File:      filePath,
						Pos:       node.Pos(),
						Ident:     node,
						Qualified: qualified,
					})

				case *ast.SelectorExpr:
					// Check if the selector (the field/method name) matches our target symbols
					if !symbolSet[node.Sel.Name] {
						return true
					}

					// Check if this is a definition (skip it if so)
					if pkg.TypesInfo != nil && pkg.TypesInfo.Defs[node.Sel] != nil {
						return true
					}

					// Build the full selector path (e.g., "DB.data" or "obj.field")
					selectorPath := buildSelectorPath(node, pkg)

					// Determine if this is a qualified reference (cross-package)
					qualified := false
					if pkg.TypesInfo != nil && pkg.TypesInfo.Uses[node.Sel] != nil {
						obj := pkg.TypesInfo.Uses[node.Sel]
						if obj.Pkg() != nil && obj.Pkg() != pkg.Types {
							qualified = true
						}
					}

					// Add this reference with selector information
					refs = append(refs, Reference{
						File:      filePath,
						Pos:       node.Sel.Pos(),
						Ident:     node.Sel,
						Qualified: qualified,
						Selector:  selectorPath,
					})
				}

				return true
			})
		}
	}

	return refs, nil
}

// buildSelectorPath constructs the full selector path for a selector expression.
// For example, "db.data" becomes "DB.data" if db has type *DB.
func buildSelectorPath(sel *ast.SelectorExpr, pkg *packages.Package) string {
	if pkg.TypesInfo == nil {
		// Without type info, just use the simple path
		return fmt.Sprintf("?.%s", sel.Sel.Name)
	}

	// Get the type of the base expression (X)
	tv, ok := pkg.TypesInfo.Types[sel.X]
	if !ok {
		return fmt.Sprintf("?.%s", sel.Sel.Name)
	}

	// Get the underlying type (dereference pointers)
	typ := tv.Type
	if ptr, ok := typ.(*types.Pointer); ok {
		typ = ptr.Elem()
	}

	// Get the type name
	var typeName string
	switch t := typ.(type) {
	case *types.Named:
		typeName = t.Obj().Name()
	default:
		// For anonymous types, try to get a representation
		typeName = typ.String()
	}

	return fmt.Sprintf("%s.%s", typeName, sel.Sel.Name)
}
