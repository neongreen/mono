package goutils

import (
	"fmt"
)

// updatePackageDeclaration reads a Go file, changes its package declaration in the AST, and writes it back.
func UpdatePackageDeclaration(filePath string, newPackageName string) error {
	fset, node, err := ReadGoFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	// Update the package name in the AST
	if node.Name == nil {
		return fmt.Errorf("package declaration not found in file %s", filePath)
	}
	node.Name.Name = newPackageName

	// Write the modified AST back to the file
	if err := WriteGoFile(filePath, fset, node); err != nil {
		return fmt.Errorf("error writing modified file %s: %w", filePath, err)
	}
	return nil
}
