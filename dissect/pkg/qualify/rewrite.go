package qualify

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"

	"github.com/neongreen/mono/dissect/pkg/goutils"
	"github.com/neongreen/mono/dissect/pkg/references"
	"golang.org/x/tools/go/ast/astutil"
)

// QualifyReferences adds package qualifier to unqualified references.
// It also adds the import if missing.
func QualifyReferences(
	filePath string,
	refs []references.Reference,
	packageName string,
	importPath string,
	moduleRoot string,
) error {
	if len(refs) == 0 {
		return nil
	}

	// Filter to only unqualified references for this file
	var unqualifiedRefs []references.Reference
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %w", err)
	}

	for _, ref := range refs {
		absRefPath, err := filepath.Abs(ref.File)
		if err != nil {
			continue
		}
		if absRefPath == absFilePath && !ref.Qualified {
			unqualifiedRefs = append(unqualifiedRefs, ref)
		}
	}

	if len(unqualifiedRefs) == 0 {
		return nil // Nothing to do
	}

	// Read the file
	fset, node, err := goutils.ReadGoFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Check if import already exists
	hasImport := false
	for _, imp := range astutil.Imports(fset, node) {
		for _, spec := range imp {
			path := spec.Path.Value
			if len(path) >= 2 {
				path = path[1 : len(path)-1] // Remove quotes
			}
			if path == importPath {
				hasImport = true
				break
			}
		}
		if hasImport {
			break
		}
	}

	// Add import if missing
	if !hasImport {
		if !astutil.AddNamedImport(fset, node, "", importPath) {
			return fmt.Errorf("failed to add import %s", importPath)
		}
	}

	// Create a map of positions to reference for fast lookup
	refPositions := make(map[token.Pos]references.Reference)
	for _, ref := range unqualifiedRefs {
		refPositions[ref.Pos] = ref
	}

	// Walk the AST and qualify unqualified references
	// We need to replace ast.Ident with ast.SelectorExpr
	modified := false
	
	// Use astutil.Apply to traverse and modify the AST
	astutil.Apply(node, func(cursor *astutil.Cursor) bool {
		ident, ok := cursor.Node().(*ast.Ident)
		if !ok {
			return true
		}

		// Check if this identifier is one we need to qualify
		if _, needsQualifying := refPositions[ident.Pos()]; !needsQualifying {
			return true
		}

		// Check if it's already part of a selector (qualified)
		parent := cursor.Parent()
		if sel, ok := parent.(*ast.SelectorExpr); ok && sel.Sel == ident {
			// Already qualified
			return true
		}

		// Replace the identifier with a selector expression
		newSelector := &ast.SelectorExpr{
			X:   &ast.Ident{Name: packageName},
			Sel: ident,
		}

		cursor.Replace(newSelector)
		modified = true

		return true
	}, nil)

	if !modified {
		return nil // Nothing was changed
	}

	// Write the file back
	if err := goutils.WriteGoFile(filePath, fset, node); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

