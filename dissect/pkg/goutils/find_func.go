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

// FindDecl finds any top-level declaration (function, type, interface, const, var) in a Go file by its name.
// Returns the FileSet, the declaration node, and any error.
// This function is used by the move command to locate declarations for AST-based moving.
// See DESIGN.md for implementation details.
func FindDecl(filePath, declName string) (*token.FileSet, ast.Node, error) {
	fset, node, err := ReadGoFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading Go file %s: %w", filePath, err)
	}
	
	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.Name == declName {
				return fset, d, nil
			}
		case *ast.GenDecl:
			// GenDecl can contain multiple specs (type, const, var)
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					// Type or interface declaration
					if s.Name.Name == declName {
						return fset, d, nil
					}
				case *ast.ValueSpec:
					// Const or var declaration
					for _, name := range s.Names {
						if name.Name == declName {
							return fset, d, nil
						}
					}
				}
			}
		}
	}
	
	return nil, nil, fmt.Errorf("declaration %s not found in file %s", declName, filePath)
}
