package refactor

import (
	"fmt"
	"go/ast"
	"go/types"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

// UnexportedDependency represents an unexported symbol that a file depends on.
type UnexportedDependency struct {
	Name         string // Symbol name
	Kind         string // "func", "var", "type", "const", "method"
	DefiningFile string // File where symbol is defined
}

// analyzeMoveDependencies checks if the file being moved depends on unexported symbols
// from the source package. Returns the list of unexported symbols referenced.
func analyzeMoveDependencies(sourceFile string, sourcePkg *packages.Package) ([]UnexportedDependency, error) {
	if sourcePkg == nil || sourcePkg.TypesInfo == nil {
		return nil, fmt.Errorf("package or type info is nil")
	}

	// Get absolute path for comparison
	absSourceFile, err := filepath.Abs(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path: %w", err)
	}

	// Find the AST node for this specific file
	var fileNode *ast.File
	for i, f := range sourcePkg.CompiledGoFiles {
		absCompiledPath, err := filepath.Abs(f)
		if err != nil {
			continue
		}
		if absCompiledPath == absSourceFile {
			if i < len(sourcePkg.Syntax) {
				fileNode = sourcePkg.Syntax[i]
				break
			}
		}
	}

	if fileNode == nil {
		return nil, fmt.Errorf("file %s not found in package %s", sourceFile, sourcePkg.PkgPath)
	}

	// Track symbols defined in this file (we don't want to report self-references)
	definedSymbols := make(map[string]bool)
	for _, decl := range fileNode.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name != nil {
				definedSymbols[d.Name.Name] = true
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name != nil {
						definedSymbols[s.Name.Name] = true
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name != nil {
							definedSymbols[name.Name] = true
						}
					}
				}
			}
		}
	}

	// Find all identifiers used in the file
	unexportedDeps := make(map[string]UnexportedDependency)

	ast.Inspect(fileNode, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		// Look up the object this identifier refers to
		obj := sourcePkg.TypesInfo.Uses[ident]
		if obj == nil {
			// This might be a definition, not a use
			return true
		}

		// Check if the object is from the same package
		if obj.Pkg() == nil || obj.Pkg() != sourcePkg.Types {
			// From different package or universe scope
			return true
		}

		// Check if it's unexported
		if !isUnexported(obj.Name()) {
			// Exported, OK to reference from another package
			return true
		}

		// Check if it's defined in this file (self-reference)
		if definedSymbols[obj.Name()] {
			// Don't report self-references
			return true
		}

		// This is an unexported symbol from the source package!
		// Find which file it's defined in
		definingFile := "unknown"
		if pos := obj.Pos(); pos.IsValid() {
			position := sourcePkg.Fset.Position(pos)
			definingFile = filepath.Base(position.Filename)

			// Check if it's defined in the same file (e.g., parameter, receiver, local var)
			definingFileAbs, _ := filepath.Abs(position.Filename)
			if definingFileAbs == absSourceFile {
				// This is a local symbol (parameter, receiver, local variable)
				// Don't report as a dependency
				return true
			}
		}

		// Determine the kind
		kind := determineKind(obj)

		// Add to map (using name as key to deduplicate)
		key := obj.Name()
		if _, exists := unexportedDeps[key]; !exists {
			unexportedDeps[key] = UnexportedDependency{
				Name:         obj.Name(),
				Kind:         kind,
				DefiningFile: definingFile,
			}
		}

		return true
	})

	// Convert map to slice and sort for consistent output
	deps := make([]UnexportedDependency, 0, len(unexportedDeps))
	for _, dep := range unexportedDeps {
		deps = append(deps, dep)
	}
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Name < deps[j].Name
	})

	slog.Debug("Analyzed dependencies", "file", filepath.Base(sourceFile), "unexportedCount", len(deps))

	return deps, nil
}

// isUnexported returns true if the name starts with a lowercase letter.
func isUnexported(name string) bool {
	if name == "" {
		return false
	}
	r := []rune(name)[0]
	return unicode.IsLower(r)
}

// determineKind determines the kind of an object (func, var, type, const, method).
func determineKind(obj any) string {
	switch obj.(type) {
	case *types.Func:
		// Check if it's a method by seeing if it has a receiver
		if fn, ok := obj.(*types.Func); ok {
			sig, ok := fn.Type().(*types.Signature)
			if ok && sig.Recv() != nil {
				return "method"
			}
		}
		return "func"
	case *types.Var:
		return "var"
	case *types.TypeName:
		return "type"
	case *types.Const:
		return "const"
	default:
		return "unknown"
	}
}

// formatDependencyError creates a helpful error message for unexported dependencies.
func formatDependencyError(deps []UnexportedDependency, sourceRelPath, targetRelPath string) error {
	var b strings.Builder

	b.WriteString("cannot move file: it references unexported symbols from source package\n\n")
	b.WriteString("  Referenced unexported symbols:\n")

	for _, dep := range deps {
		fmt.Fprintf(&b, "    - %s (%s in %s)\n", dep.Name, dep.Kind, dep.DefiningFile)
	}

	targetPkg := filepath.Dir(targetRelPath)
	b.WriteString(fmt.Sprintf("\n  These symbols cannot be accessed from the target package (%s).\n\n", targetPkg))

	b.WriteString("  To fix this, choose one of these options:\n\n")

	b.WriteString("  Option 1: Export the symbols\n")
	b.WriteString("    Make these symbols accessible by capitalizing their first letter, then retry:\n")
	b.WriteString(fmt.Sprintf("      dissect move %s %s\n\n", sourceRelPath, targetRelPath))

	b.WriteString("  Option 2: Move the dependent symbols together\n")
	b.WriteString("    Move all symbols that the file depends on:\n")
	symbolList := strings.Builder{}
	symbolList.WriteString(sourceRelPath)
	for _, dep := range deps {
		symbolList.WriteString(",")
		symbolList.WriteString(dep.Name)
	}
	b.WriteString(fmt.Sprintf("      dissect move %s %s\n", symbolList.String(), targetRelPath))

	return fmt.Errorf("%s", b.String())
}
