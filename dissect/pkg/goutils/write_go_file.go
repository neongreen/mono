package goutils

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
)

// WriteGoFile writes an AST back to a Go file, formatting it.
func WriteGoFile(filePath string, fset *token.FileSet, node *ast.File) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", filePath, err)
	}
	defer file.Close()

	if err := format.Node(file, fset, node); err != nil {
		return fmt.Errorf("error formatting and writing file %s: %w", filePath, err)
	}
	return nil
}
