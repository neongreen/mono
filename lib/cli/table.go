package cli

import (
	"io"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
)

// NewTable returns a go-pretty table.Writer configured with the monorepo defaults.
// It mirrors output to stdout, uses the light style, separates rows, and hides the outer border.
func NewTable(output io.Writer) table.Writer {
	t := table.NewWriter()
	if output == nil {
		output = os.Stdout
	}
	t.SetOutputMirror(output)

	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Style().Options.DrawBorder = false

	// Adjust to terminal width when possible so columns wrap nicely
	if file, ok := output.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		if width, _, err := term.GetSize(int(file.Fd())); err == nil && width > 0 {
			t.SetAllowedRowLength(width - 2) // leave a small margin
		}
	}

	return t
}
