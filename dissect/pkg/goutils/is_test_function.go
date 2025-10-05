package goutils

import (
	"dissect/pkg/utils"
	"go/ast"
)

// Check if the function name starts with "Test", is exported, and is not a method.
// If yes, return the function name without the "Test" prefix.
// If not, return an empty string.
func IsTestFunction(funcDecl *ast.FuncDecl) string {
	if funcDecl == nil || funcDecl.Name == nil {
		return ""
	}
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		return ""
	}
	// Check if the function is exported, starts with "Test", and is not followed by a lowercase letter.
	if funcDecl.Name.IsExported() && len(funcDecl.Name.Name) > 4 && funcDecl.Name.Name[:4] == "Test" {
		if len(funcDecl.Name.Name) == 4 || (len(funcDecl.Name.Name) > 4 && !utils.IsLower(funcDecl.Name.Name[4])) {
			return funcDecl.Name.Name[4:] // Return the function name without "Test" prefix
		} else {
			return "" // Not a valid test function
		}
	}
	return ""
}
