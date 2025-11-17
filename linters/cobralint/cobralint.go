// Package cobralint provides an analyzer for Cobra commands.
//
// It extracts structured information about Cobra commands and allows
// running custom checkers to enforce project-specific conventions.
//
// Example checkers:
//   - Require all commands to have a --json flag
//   - Enforce naming conventions
//   - Ensure proper documentation
package cobralint

import (
	"go/ast"
	"go/token"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// CommandInfo represents a Cobra command extracted from the code.
type CommandInfo struct {
	Name               string         // Variable name (e.g., "lsCmd")
	Use                string         // Command use string (e.g., "ls")
	Flags              []FlagInfo     // Flags attached to this command
	Pos                token.Pos      // Position in source code
	VarDecl            *ast.ValueSpec // The variable declaration
	IsRoot             bool           // Whether this is the root command
	ExemptFromJSONFlag *ExemptionInfo // JSON flag exemption with reason
}

// ExemptionInfo represents an exemption from the JSON flag requirement.
type ExemptionInfo struct {
	Reason string // Why this command is exempt from requiring --json
}

// FlagInfo represents a flag attached to a command.
type FlagInfo struct {
	Name      string    // Flag name (e.g., "json")
	ShortName string    // Short flag name (e.g., "j")
	FlagType  string    // Flag type (e.g., "Bool", "String")
	Pos       token.Pos // Position where flag is defined
}

// Checker defines an interface for command checkers.
type Checker interface {
	// Check runs the checker on a command and reports issues via the analysis pass.
	Check(pass *analysis.Pass, cmd *CommandInfo)
	// Name returns the name of the checker.
	Name() string
}

var Analyzer = &analysis.Analyzer{
	Name:       "cobralint",
	Doc:        "enforces conventions for Cobra commands",
	Run:        run,
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	ResultType: reflect.TypeFor[[]CommandInfo](),
}

// EnabledCheckers is the list of checkers to run.
// This can be modified to enable/disable specific checkers.
var EnabledCheckers = []Checker{
	&RequireJSONFlagChecker{},
}

func run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// First pass: find all cobra.Command definitions
	commands := extractCommands(pass, inspect)

	// Second pass: find flag definitions for each command
	for i := range commands {
		commands[i].Flags = extractFlags(pass, inspect, commands[i].Name)
	}

	// Run all enabled checkers
	for _, cmd := range commands {
		for _, checker := range EnabledCheckers {
			checker.Check(pass, &cmd)
		}
	}

	return commands, nil
}

// extractCommands finds all cobra.Command variable declarations in the code.
func extractCommands(pass *analysis.Pass, inspect *inspector.Inspector) []CommandInfo {
	var commands []CommandInfo

	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		genDecl := n.(*ast.GenDecl)

		// Only interested in variable declarations
		if genDecl.Tok != token.VAR {
			return
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			// Check if this is a cobra.Command
			if !isCobraCommand(valueSpec) {
				continue
			}

			// Extract command info
			for i, name := range valueSpec.Names {
				if i >= len(valueSpec.Values) {
					continue
				}

				cmdInfo := CommandInfo{
					Name:    name.Name,
					Pos:     name.Pos(),
					VarDecl: valueSpec,
					IsRoot:  name.Name == "RootCmd",
				}

				// Extract the Use field from the command literal
				val := valueSpec.Values[i]
				// Handle &cobra.Command{...}
				if unary, ok := val.(*ast.UnaryExpr); ok && unary.Op == token.AND {
					val = unary.X
				}
				if compLit, ok := val.(*ast.CompositeLit); ok {
					cmdInfo.Use = extractUseField(compLit)
				}

				// Check for exemption directive in comments
				cmdInfo.ExemptFromJSONFlag = extractExemption(pass, genDecl)

				commands = append(commands, cmdInfo)
			}
		}
	})

	return commands
}

// extractExemption checks for a cobralint:exemptjson directive in comments above the declaration.
// The directive format is: // cobralint:exemptjson reason: <explanation>
func extractExemption(pass *analysis.Pass, genDecl *ast.GenDecl) *ExemptionInfo {
	if genDecl.Doc == nil {
		return nil
	}

	// Regular expression to match: cobralint:exemptjson reason: <text>
	exemptRe := regexp.MustCompile(`cobralint:exemptjson\s+reason:\s*(.+)`)

	for _, comment := range genDecl.Doc.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimSpace(text)

		if matches := exemptRe.FindStringSubmatch(text); matches != nil {
			return &ExemptionInfo{
				Reason: strings.TrimSpace(matches[1]),
			}
		}
	}

	return nil
}

// isCobraCommand checks if a ValueSpec is a cobra.Command.
func isCobraCommand(spec *ast.ValueSpec) bool {
	// Check explicit type annotation: var foo *cobra.Command
	if spec.Type != nil {
		switch typ := spec.Type.(type) {
		case *ast.StarExpr:
			if sel, ok := typ.X.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "cobra" && sel.Sel.Name == "Command" {
						return true
					}
				}
			}
		case *ast.SelectorExpr:
			if ident, ok := typ.X.(*ast.Ident); ok {
				if ident.Name == "cobra" && typ.Sel.Name == "Command" {
					return true
				}
			}
		}
	}

	// Check value initialization: var foo = &cobra.Command{}
	if len(spec.Values) > 0 {
		if slices.ContainsFunc(spec.Values, isCobraCommandValue) {
			return true
		}
	}

	return false
}

// isCobraCommandValue checks if an expression is a cobra.Command literal.
func isCobraCommandValue(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.UnaryExpr:
		if e.Op == token.AND {
			return isCobraCommandValue(e.X)
		}
	case *ast.CompositeLit:
		if sel, ok := e.Type.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				return ident.Name == "cobra" && sel.Sel.Name == "Command"
			}
		}
	}
	return false
}

// extractUseField extracts the Use field value from a cobra.Command composite literal.
func extractUseField(compLit *ast.CompositeLit) string {
	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Use" {
			continue
		}

		// Extract string value
		if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			// Remove quotes
			return strings.Trim(lit.Value, "\"")
		}
	}
	return ""
}

// extractFlags finds all flag definitions for a given command.
func extractFlags(pass *analysis.Pass, inspect *inspector.Inspector, cmdName string) []FlagInfo {
	var flags []FlagInfo

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		callExpr := n.(*ast.CallExpr)

		// Look for cmdName.Flags().FlagType("name", ...)
		// This matches patterns like: lsCmd.Flags().Bool("json", false, "...")
		if !isFlagDefinition(callExpr, cmdName) {
			return
		}

		flag := extractFlagInfo(callExpr)
		if flag.Name != "" {
			flags = append(flags, flag)
		}
	})

	return flags
}

// isFlagDefinition checks if a call expression is a flag definition for the given command.
func isFlagDefinition(call *ast.CallExpr, cmdName string) bool {
	// Pattern: cmdName.Flags().FlagType(...)
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Check if this is a call to Flags()
	flagsCall, ok := sel.X.(*ast.CallExpr)
	if !ok {
		return false
	}

	flagsSel, ok := flagsCall.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if flagsSel.Sel.Name != "Flags" {
		return false
	}

	// Check if it's called on our command
	ident, ok := flagsSel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == cmdName
}

// extractFlagInfo extracts flag information from a flag definition call.
func extractFlagInfo(call *ast.CallExpr) FlagInfo {
	sel := call.Fun.(*ast.SelectorExpr)
	flagType := sel.Sel.Name

	var flagName, shortName string

	// Determine which argument contains the flag name
	// For *Var methods (BoolVar, StringVar, etc.), the first arg is the destination variable
	// and the second arg is the flag name
	// For regular methods (Bool, String, etc.), the first arg is the flag name
	nameArgIndex := 0
	shortNameArgIndex := 1

	if strings.HasSuffix(flagType, "Var") || strings.HasSuffix(flagType, "VarP") {
		nameArgIndex = 1
		shortNameArgIndex = 2
	}

	// Extract flag name from the appropriate argument
	if len(call.Args) > nameArgIndex {
		if lit, ok := call.Args[nameArgIndex].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			flagName = strings.Trim(lit.Value, "\"")
		}
	}

	// For *P methods (StringP, BoolP, BoolVarP, etc.), extract the short name
	if strings.Contains(flagType, "P") && len(call.Args) > shortNameArgIndex {
		if lit, ok := call.Args[shortNameArgIndex].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			shortName = strings.Trim(lit.Value, "\"")
		}
	}

	return FlagInfo{
		Name:      flagName,
		ShortName: shortName,
		FlagType:  flagType,
		Pos:       call.Pos(),
	}
}
