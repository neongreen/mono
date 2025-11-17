package qualify

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

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
		slog.Debug("No unqualified references to fix in file", "file", filePath, "totalRefs", len(refs))
		return nil // Nothing to do
	}

	slog.Debug("Qualifying references", "file", filePath, "unqualifiedCount", len(unqualifiedRefs))

	// Collect all identifiers in the file to detect collisions
	existingIdentifiers, err := CollectIdentifiers(filePath)
	if err != nil {
		return fmt.Errorf("failed to collect identifiers: %w", err)
	}

	// Check if packageName collides with existing identifiers
	actualPackageName := packageName
	if existingIdentifiers[packageName] {
		actualPackageName = generateImportAlias(packageName, existingIdentifiers)
		slog.Info("Package name collision detected, using alias",
			"originalName", packageName,
			"alias", actualPackageName,
			"file", filePath)
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
	// Use alias if actualPackageName differs from original packageName
	if !hasImport {
		importAlias := ""
		if actualPackageName != packageName {
			importAlias = actualPackageName
		}
		if !astutil.AddNamedImport(fset, node, importAlias, importPath) {
			return fmt.Errorf("failed to add import %s", importPath)
		}
	}

	// Create a set of symbol names to qualify
	symbolNames := make(map[string]bool)
	for _, ref := range unqualifiedRefs {
		symbolNames[ref.Ident.Name] = true
	}

	slog.Debug("Symbol names to qualify", "names", symbolNames)

	// Walk the AST and qualify unqualified references
	// We need to replace ast.Ident with ast.SelectorExpr
	modified := false

	// Use astutil.Apply to traverse and modify the AST
	astutil.Apply(node, func(cursor *astutil.Cursor) bool {
		ident, ok := cursor.Node().(*ast.Ident)
		if !ok {
			return true
		}

		// Check if this identifier is one of the symbols we need to qualify
		if !symbolNames[ident.Name] {
			return true
		}

		// Check if it's already part of a selector (qualified)
		parent := cursor.Parent()
		if sel, ok := parent.(*ast.SelectorExpr); ok && sel.Sel == ident {
			// Already qualified
			slog.Debug("Skipping already qualified reference", "name", ident.Name)
			return true
		}

		// Check if this identifier is a definition name (not a type reference)
		switch p := parent.(type) {
		case *ast.FuncDecl:
			// Skip function name in declaration
			if p.Name == ident {
				slog.Debug("Skipping function name in declaration", "name", ident.Name)
				return true
			}
		case *ast.TypeSpec:
			// Skip type name in declaration
			if p.Name == ident {
				slog.Debug("Skipping type name in declaration", "name", ident.Name)
				return true
			}
		case *ast.ValueSpec:
			// Skip variable/const names in declaration (but not type references)
			if slices.Contains(p.Names, ident) {
				slog.Debug("Skipping variable name in declaration", "name", ident.Name)
				return true
			}
		}

		// Only qualify if the parent node type can accept a selector expression
		// Most contexts can, but some (like function names in declarations) cannot
		canQualify := true
		switch p := parent.(type) {
		case *ast.FuncDecl:
			// Don't qualify function names in declarations
			canQualify = false
		case *ast.Field:
			// Don't qualify field names in struct definitions
			if slices.Contains(p.Names, ident) {
				canQualify = false
			}
		case *ast.AssignStmt:
			// Check if ident is on the left side of assignment (definition)
			for _, lhs := range p.Lhs {
				if lhs == ident {
					canQualify = false
					break
				}
			}
		case *ast.KeyValueExpr:
			// Don't qualify field names in struct literals
			// In a KeyValueExpr like "ProjectUID: value", the key is the field name
			if p.Key == ident {
				slog.Debug("Skipping struct literal field name", "name", ident.Name)
				canQualify = false
			}
		}

		if !canQualify {
			slog.Debug("Skipping identifier in non-qualifiable position", "name", ident.Name)
			return true
		}

		// Replace the identifier with a selector expression
		slog.Debug("Qualifying reference", "name", ident.Name, "package", actualPackageName)
		newSelector := &ast.SelectorExpr{
			X:   &ast.Ident{Name: actualPackageName},
			Sel: ident,
		}

		cursor.Replace(newSelector)
		modified = true

		return true
	}, nil)

	if !modified {
		slog.Debug("No AST modifications were made", "file", filePath)
		return nil // Nothing was changed
	}

	slog.Debug("Writing modified file", "file", filePath)

	// Write the file back
	if err := goutils.WriteGoFile(filePath, fset, node); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	slog.Debug("Successfully wrote modified file", "file", filePath)

	return nil
}

// CollectIdentifiers collects all declared identifiers in a Go file.
// This includes variables, functions, types, consts, parameters, and receivers.
func CollectIdentifiers(filePath string) (map[string]bool, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	identifiers := make(map[string]bool)

	ast.Inspect(node, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.FuncDecl:
			// Function name
			if decl.Name != nil {
				identifiers[decl.Name.Name] = true
			}
			// Receiver name
			if decl.Recv != nil {
				for _, field := range decl.Recv.List {
					for _, name := range field.Names {
						identifiers[name.Name] = true
					}
				}
			}
			// Parameters
			if decl.Type.Params != nil {
				for _, field := range decl.Type.Params.List {
					for _, name := range field.Names {
						identifiers[name.Name] = true
					}
				}
			}
			// Results
			if decl.Type.Results != nil {
				for _, field := range decl.Type.Results.List {
					for _, name := range field.Names {
						identifiers[name.Name] = true
					}
				}
			}

		case *ast.GenDecl:
			// Type, const, var declarations
			for _, spec := range decl.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					// Type name
					if s.Name != nil {
						identifiers[s.Name.Name] = true
					}
				case *ast.ValueSpec:
					// Const/var names
					for _, name := range s.Names {
						identifiers[name.Name] = true
					}
				}
			}

		case *ast.AssignStmt:
			// Short variable declarations
			for _, lhs := range decl.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					identifiers[ident.Name] = true
				}
			}
		}

		return true
	})

	return identifiers, nil
}

// generateImportAlias generates a unique import alias that doesn't collide with existing identifiers.
// Strategy: try pkgName, then pkgName_pkg, then pkgName_pkg_, pkgName_pkg__, etc.
func generateImportAlias(pkgName string, existingIdentifiers map[string]bool) string {
	// Try the package name first
	if !existingIdentifiers[pkgName] {
		return pkgName
	}

	// Try pkgName_pkg
	candidate := pkgName + "_pkg"
	if !existingIdentifiers[candidate] {
		return candidate
	}

	// Try pkgName_pkg_, pkgName_pkg__, etc.
	for i := 1; ; i++ {
		candidate = pkgName + "_pkg" + strings.Repeat("_", i)
		if !existingIdentifiers[candidate] {
			return candidate
		}
	}
}
