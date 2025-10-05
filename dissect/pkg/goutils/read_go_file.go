package goutils

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// ReadGoFile reads a Go file and parses it into an AST.
func ReadGoFile(filePath string) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing file %s: %w", filePath, err)
	}
	return fset, node, nil
}
