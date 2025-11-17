// Package dircheck defines an analyzer that detects file writes without directory existence checks.
//
// This analyzer detects patterns where files are written to paths without ensuring
// the parent directory exists first. This can cause "no such file or directory" errors
// at runtime.
//
// Example of problematic code:
//
//	func writeConfig(path string, data []byte) error {
//	    return os.WriteFile(path, data, 0o644)  // May fail if directory doesn't exist!
//	}
//
// Better approach:
//
//	func writeConfig(path string, data []byte) error {
//	    dir := filepath.Dir(path)
//	    if err := os.MkdirAll(dir, 0o755); err != nil {
//	        return err
//	    }
//	    return os.WriteFile(path, data, 0o644)
//	}
//
// The linter can be suppressed with //nolint:dircheck or //nolint comments.
package dircheck

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "dircheck",
	Doc:      "detects file writes without ensuring parent directory exists",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

// File creation functions that require parent directory to exist
var fileCreationFuncs = map[string][]string{
	"os": {
		"WriteFile",
		"Create",
		"CreateTemp", // When first arg is not empty string
		"OpenFile",
	},
	"io/ioutil": {
		"WriteFile", // Deprecated but still used
	},
}

func run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		callExpr := n.(*ast.CallExpr)

		// Check if this is a file creation function call
		if !isFileCreationCall(callExpr, pass.TypesInfo) {
			return
		}

		// Check if there's a nolint comment
		if hasNolintComment(callExpr, pass) {
			return
		}

		// Check if there's a directory creation in the same function before this call
		if hasPriorDirCreation(callExpr, pass) {
			return
		}

		// Get the function name for the diagnostic
		funcName := getFunctionName(callExpr)
		pass.Reportf(callExpr.Pos(),
			"file write without directory check: ensure parent directory exists before calling %s (use os.MkdirAll or suppress with //nolint:dircheck)",
			funcName)
	})

	return nil, nil
}

// isFileCreationCall checks if a call expression is a file creation function
func isFileCreationCall(callExpr *ast.CallExpr, info *types.Info) bool {
	sel, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Get the package identifier
	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	// Check if the object is from a known package
	obj := info.Uses[pkgIdent]
	if obj == nil {
		return false
	}

	// For PkgName objects (imports), check the imported package path
	pkgName, ok := obj.(*types.PkgName)
	if !ok {
		return false
	}

	pkgPath := pkgName.Imported().Path()
	funcName := sel.Sel.Name

	// Check against known file creation functions
	for knownPkg, funcs := range fileCreationFuncs {
		if pkgPath == knownPkg {
			if slices.Contains(funcs, funcName) {
				return true
			}
		}
	}

	return false
}

// hasNolintComment checks if a call has a nolint comment for dircheck
func hasNolintComment(callExpr *ast.CallExpr, pass *analysis.Pass) bool {
	// Find the file containing this call
	var file *ast.File
	for _, f := range pass.Files {
		if f.Pos() <= callExpr.Pos() && callExpr.End() <= f.End() {
			file = f
			break
		}
	}
	if file == nil {
		return false
	}

	// Check all comments in the file
	for _, cg := range file.Comments {
		// Only check comments that end before or at the start of the call
		if cg.End() <= callExpr.Pos() {
			for _, comment := range cg.List {
				text := comment.Text
				// Support both //nolint and //nolint:dircheck
				if strings.Contains(text, "nolint") &&
					(strings.Contains(text, "dircheck") || !strings.Contains(text, ":")) {
					return true
				}
			}
		}
	}
	return false
}

// hasPriorDirCreation checks if there's an os.MkdirAll call before this file write
// in the same function
func hasPriorDirCreation(callExpr *ast.CallExpr, pass *analysis.Pass) bool {
	// Find the enclosing function
	var enclosingFunc *ast.FuncDecl
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				if funcDecl.Body != nil &&
					funcDecl.Pos() <= callExpr.Pos() &&
					callExpr.End() <= funcDecl.End() {
					enclosingFunc = funcDecl
					return false
				}
			}
			return true
		})
		if enclosingFunc != nil {
			break
		}
	}

	if enclosingFunc == nil || enclosingFunc.Body == nil {
		return false
	}

	// Look for os.MkdirAll calls before this call
	foundMkdirAll := false
	ast.Inspect(enclosingFunc.Body, func(n ast.Node) bool {
		// Stop if we've reached the file creation call
		if n != nil && n.Pos() >= callExpr.Pos() {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		pkgIdent, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		// Check if this is os.MkdirAll
		if pkgIdent.Name == "os" && sel.Sel.Name == "MkdirAll" {
			foundMkdirAll = true
			return false
		}

		return true
	})

	return foundMkdirAll
}

// getFunctionName returns a readable name for the function being called
func getFunctionName(callExpr *ast.CallExpr) string {
	switch fun := callExpr.Fun.(type) {
	case *ast.SelectorExpr:
		if ident, ok := fun.X.(*ast.Ident); ok {
			return ident.Name + "." + fun.Sel.Name
		}
	}
	return "file creation function"
}
