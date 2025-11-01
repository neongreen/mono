package gopls

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	"os/exec"
)

// Rename renames a symbol in a Go file using gopls.
// It finds the symbol's position and uses gopls rename to update all references.
func Rename(filePath string, oldName string, newName string) error {
	slog.Debug("Renaming symbol", "file", filePath, "old", oldName, "new", newName)

	// Validate new name
	if newName == "" {
		return fmt.Errorf("new name cannot be empty")
	}

	// Find the symbol's position in the file
	offset, err := findSymbolOffset(filePath, oldName)
	if err != nil {
		return fmt.Errorf("error finding symbol '%s': %w", oldName, err)
	}

	// Run gopls rename
	// Format: gopls rename -w file.go:#offset newName
	positionSpec := fmt.Sprintf("%s:#%d", filePath, offset)

	cmd := exec.Command("gopls", "rename", "-w", positionSpec, newName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gopls rename failed: %w\nOutput: %s", err, string(output))
	}

	slog.Debug("Successfully renamed symbol", "old", oldName, "new", newName, "output", string(output))
	return nil
}

// findSymbolOffset finds the byte offset of a symbol's definition in a file.
// This is used to provide gopls with the precise location for renaming.
func findSymbolOffset(filePath string, symbolName string) (int, error) {
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("error reading file: %w", err)
	}

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return 0, fmt.Errorf("error parsing file: %w", err)
	}

	// Find the symbol definition
	var symbolPos token.Pos
	found := false

	ast.Inspect(node, func(n ast.Node) bool {
		if found {
			return false
		}

		switch decl := n.(type) {
		case *ast.FuncDecl:
			// Function or method definition
			if decl.Name.Name == symbolName {
				symbolPos = decl.Name.Pos()
				found = true
				return false
			}
		case *ast.GenDecl:
			// Type, const, var declaration
			for _, spec := range decl.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.Name == symbolName {
						symbolPos = s.Name.Pos()
						found = true
						return false
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name.Name == symbolName {
							symbolPos = name.Pos()
							found = true
							return false
						}
					}
				}
			}
		}
		return true
	})

	if !found {
		return 0, fmt.Errorf("symbol '%s' not found in file", symbolName)
	}

	// Convert position to byte offset
	position := fset.Position(symbolPos)
	return position.Offset, nil
}
