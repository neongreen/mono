// Package uselesswrapper defines an analyzer that detects useless function wrappers.
//
// A useless wrapper is a function that:
//   - Takes parameters and returns values
//   - Contains only a single return statement
//   - The return statement calls another function with the exact same parameters
//   - No transformation, validation, or additional logic is performed
//
// Example of a useless wrapper:
//
//	func getCurrentUser() (string, error) {
//	    return utils.GetCurrentUser()
//	}
//
// This wrapper adds no value and creates unnecessary indirection.
// The caller should directly call utils.GetCurrentUser() instead.
package uselesswrapper

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "uselesswrapper",
	Doc:      "detects useless function wrappers that add no value",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl := n.(*ast.FuncDecl)

		// Skip functions with no body (interfaces, external functions)
		if funcDecl.Body == nil {
			return
		}

		// Skip methods (functions with receivers)
		if funcDecl.Recv != nil {
			return
		}

		// Check if the function body has exactly one statement
		if len(funcDecl.Body.List) != 1 {
			return
		}

		// Check if that statement is a return statement
		retStmt, ok := funcDecl.Body.List[0].(*ast.ReturnStmt)
		if !ok {
			return
		}

		// Check if the return statement has exactly one result
		if len(retStmt.Results) != 1 {
			return
		}

		// Check if the result is a function call
		callExpr, ok := retStmt.Results[0].(*ast.CallExpr)
		if !ok {
			return
		}

		// Now check if the call passes all parameters unchanged
		if !isPassthroughCall(funcDecl, callExpr, pass.TypesInfo) {
			return
		}

		pass.Reportf(funcDecl.Pos(), "useless wrapper function: directly call %s instead",
			formatCallExpr(callExpr))
	})

	return nil, nil
}

// isPassthroughCall checks if a call expression passes all function parameters unchanged
func isPassthroughCall(funcDecl *ast.FuncDecl, callExpr *ast.CallExpr, info *types.Info) bool {
	// Get function parameters
	if funcDecl.Type.Params == nil {
		// Function has no parameters, but we're calling something
		// This might still be a useless wrapper if the call also has no params
		return len(callExpr.Args) == 0
	}

	// Build a map of parameter names
	paramNames := make(map[string]bool)
	var paramList []*ast.Ident
	for _, field := range funcDecl.Type.Params.List {
		for _, name := range field.Names {
			paramNames[name.Name] = true
			paramList = append(paramList, name)
		}
	}

	// If the function has parameters but the call doesn't, it's not a passthrough
	if len(paramList) != len(callExpr.Args) {
		return false
	}

	// Check if all call arguments are identifiers that match the parameters
	for i, arg := range callExpr.Args {
		ident, ok := arg.(*ast.Ident)
		if !ok {
			// Argument is not a simple identifier (e.g., could be a.b, f(), etc.)
			return false
		}

		// Check if this identifier matches the corresponding parameter
		if i < len(paramList) && ident.Name != paramList[i].Name {
			return false
		}
	}

	return true
}

// formatCallExpr formats a call expression for display in the diagnostic message
func formatCallExpr(callExpr *ast.CallExpr) string {
	switch fun := callExpr.Fun.(type) {
	case *ast.Ident:
		return fun.Name + "()"
	case *ast.SelectorExpr:
		if ident, ok := fun.X.(*ast.Ident); ok {
			return ident.Name + "." + fun.Sel.Name + "()"
		}
		return "function"
	default:
		return "function"
	}
}
