package goutils

import "go/ast"

// For functions that are methods, get the name of the type it's a method of.
// Otherwise, return an empty string.
func GetReceiverTypeName(funcDecl *ast.FuncDecl) string {
	var receiverTypeName string
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		recvType := funcDecl.Recv.List[0].Type
		switch t := recvType.(type) {
		case *ast.Ident:
			receiverTypeName = t.Name
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				receiverTypeName = ident.Name
			}
		}
	}
	return receiverTypeName
}
