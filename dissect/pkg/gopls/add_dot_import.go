package gopls

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"

	"github.com/neongreen/mono/dissect/pkg/goutils"
)

// AddDotImport adds a dot import to the given file.
func AddDotImport(goplsPath string, filePath string, importPath string, moduleRoot string) error {
	absFilePath := filePath
	if !filepath.IsAbs(filePath) {
		absFilePath = filepath.Join(moduleRoot, filePath)
	}

	// First call AddImport to add the import
	if err := AddImport(goplsPath, absFilePath, importPath, moduleRoot); err != nil {
		return fmt.Errorf("error adding import: %w", err)
	}

	// Now parse the file, find the import statement, and modify it to be a dot import
	fset, node, err := goutils.ReadGoFile(absFilePath)
	if err != nil {
		return fmt.Errorf("error reading Go file: %w", err)
	}

	// Find and modify the import declaration
	modified := false
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			for _, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					// Remove quotes from the import path for comparison
					specPath := importSpec.Path.Value
					if len(specPath) >= 2 && specPath[0] == '"' && specPath[len(specPath)-1] == '"' {
						specPath = specPath[1 : len(specPath)-1]
					}

					// If this is the import we're looking for, make it a dot import
					if specPath == importPath {
						importSpec.Name = &ast.Ident{Name: "."}
						modified = true
						break
					}
				}
			}
			if modified {
				break
			}
		}
	}

	if !modified {
		return fmt.Errorf("import %s not found in file %s", importPath, filePath)
	}

	// Write the modified AST back to the file
	if err := goutils.WriteGoFile(absFilePath, fset, node); err != nil {
		return fmt.Errorf("error writing modified file: %w", err)
	}

	return nil
}
