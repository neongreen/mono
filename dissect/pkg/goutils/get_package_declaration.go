package goutils

import (
	"fmt"
)

// getPackageDeclaration reads a Go file and extracts its package declaration.
func GetPackageDeclaration(filePath string) (string, error) {
	_, node, err := ReadGoFile(filePath)
	if err != nil {
		return "", err
	}
	if node.Name == nil {
		return "", fmt.Errorf("package declaration not found in file %s", filePath)
	}
	return node.Name.Name, nil
}
