package goutils

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

// NormalizeImports takes Go source code as a string, parses it,
// ensures all imports are grouped and parenthesized, sorts them,
// and returns the normalized code.
func NormalizeImports(source string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", source, parser.ParseComments)
	if err != nil {
		return "", err
	}

	// Collect all import paths
	var importPaths []string
	var importDecl *ast.GenDecl // To hold the single import declaration (if any)

	// Remove existing import declarations
	var newDecls []ast.Decl
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			for _, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					path := importSpec.Path.Value
					if importSpec.Name != nil {
						path = importSpec.Name.Name + " " + path
					}
					importPaths = append(importPaths, path)
				}
			}
			if importDecl == nil { // Keep the first import declaration to reuse its position
				importDecl = genDecl
			}
		} else {
			newDecls = append(newDecls, decl)
		}
	}
	file.Decls = newDecls

	// Sort import paths
	sort.Strings(importPaths)

	// Create a single, parenthesized import declaration
	if len(importPaths) > 0 {
		var specs []ast.Spec
		for _, path := range importPaths {
			var name *ast.Ident
			actualPath := path
			if parts := strings.SplitN(path, " ", 2); len(parts) == 2 {
				name = ast.NewIdent(parts[0])
				actualPath = parts[1]
			}
			specs = append(specs, &ast.ImportSpec{
				Path: &ast.BasicLit{Kind: token.STRING, Value: actualPath},
				Name: name,
			})
		}

		// Create a new GenDecl for the combined imports
		newImportDecl := &ast.GenDecl{
			Tok:    token.IMPORT,
			Lparen: 1, // Force parenthesized import
			Specs:  specs,
			Rparen: 1, // Force parenthesized import
		}

		// Insert the new import declaration at the beginning of the file, after package
		file.Decls = append([]ast.Decl{newImportDecl}, file.Decls...)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return "", err
	}
	return buf.String(), nil
}
