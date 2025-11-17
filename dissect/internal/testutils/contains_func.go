package testutils

// ContainsFunc checks if a function exists in code
func ContainsFunc(code string, funcName string) bool {
	funcDecl := "func " + funcName + "("
	return ContainsString(code, funcDecl)
}
