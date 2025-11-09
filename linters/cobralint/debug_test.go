package cobralint_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/neongreen/mono/linters/cobralint"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func TestDebug(t *testing.T) {
	src := `package testcmd

type Command struct {
	Use   string
	Short string
}

var goodCmd = &Command{
	Use:   "good",
	Short: "A command with json flag",
}

var badCmd = &Command{
	Use:   "bad",
	Short: "A command without json flag",
}

func init() {
	// Simulate flag definition
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Create a fake pass
	pass := &analysis.Pass{
		Fset:  fset,
		Files: []*ast.File{f},
	}

	// Create inspector
	insp := inspector.New([]*ast.File{f})
	pass.ResultOf = map[*analysis.Analyzer]interface{}{
		inspect.Analyzer: insp,
	}

	// Run the analyzer
	result, err := cobralint.Analyzer.Run(pass)
	if err != nil {
		t.Fatal(err)
	}

	commands := result.([]cobralint.CommandInfo)
	fmt.Printf("Found %d commands:\n", len(commands))
	for _, cmd := range commands {
		fmt.Printf("  - %s (use: %s, flags: %d)\n", cmd.Name, cmd.Use, len(cmd.Flags))
	}
}
