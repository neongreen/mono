package refactor

import (
	"github.com/neongreen/mono/dissect/pkg/goutils"
	"github.com/neongreen/mono/dissect/pkg/utils"
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
)

// DetermineExtractionTarget determines where we'd like to extract the function to.
func DetermineExtractionTarget(sourceFilePath string, funcDecl *ast.FuncDecl) (
	targetFilePath string,
	targetFuncName string,
	targetPackageName string,
) {
	targetFuncName = funcDecl.Name.Name

	// If the function is a method, we need to extract the receiver type name.
	// Then we will call the file `typename_funcname.go` instead of just `funcname.go`.
	receiverTypeName := goutils.GetReceiverTypeName(funcDecl)

	makeFileName := func(funcName string, receiverTypeName string, isUtil bool, isTestFunc bool) string {
		s := strcase.ToSnake(funcName)
		if isTestFunc {
			s += "_test"
		} else {
			if isUtil {
				s = "util_" + s
			}
			if receiverTypeName != "" {
				// Using ToLower instead of ToSnake because it feels better to transform e.g. FileName to "filename" instead of "file_name"
				s = strings.ToLower(receiverTypeName) + "_" + s
			}
		}
		return s + ".go"
	}

	originalPackage, err := goutils.GetPackageDeclaration(sourceFilePath)
	if err != nil {
		// Fallback or error handling if package declaration cannot be read
		// For now, we'll assume it's "main" or handle the error upstream
		originalPackage = "main"
	}
	targetPackageName = originalPackage

	// If the function is a test (TestSomething) and is in a test file, we move it to something_test.go.
	// IsTestFunction doesn't consider methods to be test functions.
	if testName := goutils.IsTestFunction(funcDecl); goutils.IsTestFile(sourceFilePath) && testName != "" {
		targetFilePath = filepath.Join(
			filepath.Dir(sourceFilePath),
			makeFileName(testName, receiverTypeName, false, true))
	} else
	// If it's a test file but not a test function, we assume it's a helper and move to internal/testutils
	if goutils.IsTestFile(sourceFilePath) {
		targetFuncName = utils.CapitalizeFirstLetter(targetFuncName)
		targetFilePath = filepath.Join(
			filepath.Dir(sourceFilePath), "internal", "testutils",
			makeFileName(targetFuncName, receiverTypeName, false, false))
		targetPackageName = "testutils"
	} else
	// If it's a `main` function, we leave it alone
	if funcDecl.Name.Name == "main" {
		targetFilePath = sourceFilePath
	} else
	// If it's a non-exported function in a non-main package, we move it to util_functionname.go
	if !funcDecl.Name.IsExported() {
		targetFilePath = filepath.Join(
			filepath.Dir(sourceFilePath),
			makeFileName(targetFuncName, receiverTypeName, true, false))
	} else
	// If it's an exported function, doesn't matter where, we move it to a new file in the same directory
	{
		targetFilePath = filepath.Join(
			filepath.Dir(sourceFilePath),
			makeFileName(targetFuncName, receiverTypeName, false, false))
	}
	return targetFilePath, targetFuncName, targetPackageName
}
