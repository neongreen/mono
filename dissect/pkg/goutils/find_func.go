package goutils

import (
	"fmt"
	"go/ast"
	"go/token"
)

// FindFunc finds a function in a Go file by its name.
func FindFunc(filePath, funcName string) (*token.FileSet, *ast.FuncDecl, error) {
	fset, node, err := ReadGoFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading Go file %s: %w", filePath, err)
	}
	for _, decl := range node.Decls {
		if fn, isFn := decl.(*ast.FuncDecl); isFn && fn.Name.Name == funcName {
			// Return the function declaration if the name matches
			return fset, fn, nil
		}
	}
	return nil, nil, fmt.Errorf("function %s not found in file %s", funcName, filePath)
}
