package goutils

import (
	"fmt"
	"go/ast"
	"os"
)

// ExtractFunctionText extracts the text of a function declaration including its doc comments
// from a Go source file. It returns the raw text that can be appended to another file.
func ExtractFunctionText(filePath string) (string, error) {
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	// Parse the file
	fset, node, err := ReadGoFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error parsing file %s: %w", filePath, err)
	}

	// Find the first function declaration in the file
	var funcDecl *ast.FuncDecl
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			funcDecl = fn
			break
		}
	}

	if funcDecl == nil {
		return "", fmt.Errorf("no function found in file %s", filePath)
	}

	// Determine the start position - include doc comments if they exist
	startPos := funcDecl.Pos()
	if funcDecl.Doc != nil {
		startPos = funcDecl.Doc.Pos()
	}
	endPos := funcDecl.End()

	// Convert to byte offsets
	start := fset.Position(startPos).Offset
	end := fset.Position(endPos).Offset

	// Extract the text
	funcText := string(content[start:end])

	return funcText, nil
}

// ExtractImports extracts the import declarations from a Go source file.
// It returns them as AST import specs that can be merged into another file.
func ExtractImports(filePath string) ([]*ast.ImportSpec, error) {
	_, node, err := ReadGoFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error parsing file %s: %w", filePath, err)
	}

	return node.Imports, nil
}
