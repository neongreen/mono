package gopls

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Rename renames a symbol in a Go file using gopls.
// It finds the symbol's position and uses gopls rename to update all references.
func Rename(goplsPath string, filePath string, oldName string, newName string, moduleRoot string) error {
	slog.Debug("Renaming symbol", "file", filePath, "old", oldName, "new", newName)

	// Validate new name
	if newName == "" {
		return fmt.Errorf("new name cannot be empty")
	}

	// Find the symbol's position in the file
	offset, err := findSymbolOffset(filePath, oldName)
	if err != nil {
		return fmt.Errorf("error finding symbol '%s': %w", oldName, err)
	}

	// Run gopls rename
	// Format: gopls rename -w file.go:#offset newName
	positionSpec := fmt.Sprintf("%s:#%d", filePath, offset)

	cmd := exec.Command(goplsPath, "rename", "-w", positionSpec, newName)
	cmd.Dir = moduleRoot

	var combinedOutput bytes.Buffer
	cmd.Stdout = &combinedOutput
	cmd.Stderr = &combinedOutput

	err = runWithTextFileBusyRetry(cmd)
	if err != nil {
		return fmt.Errorf("gopls rename failed: %w\nOutput: %s", err, combinedOutput.String())
	}

	slog.Debug("Successfully renamed symbol", "old", oldName, "new", newName, "output", combinedOutput.String(), "goplsPath", goplsPath)
	return nil
}

// findSymbolOffset finds the byte offset of a symbol's definition in a file.
// This is used to provide gopls with the precise location for renaming.
func findSymbolOffset(filePath string, symbolName string) (int, error) {
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("error reading file: %w", err)
	}

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return 0, fmt.Errorf("error parsing file: %w", err)
	}

	// Find the symbol definition
	var symbolPos token.Pos
	found := false

	ast.Inspect(node, func(n ast.Node) bool {
		if found {
			return false
		}

		switch decl := n.(type) {
		case *ast.FuncDecl:
			// Function or method definition
			if decl.Name.Name == symbolName {
				symbolPos = decl.Name.Pos()
				found = true
				return false
			}
		case *ast.GenDecl:
			// Type, const, var declaration
			for _, spec := range decl.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.Name == symbolName {
						symbolPos = s.Name.Pos()
						found = true
						return false
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name.Name == symbolName {
							symbolPos = name.Pos()
							found = true
							return false
						}
					}
				}
			}
		}
		return true
	})

	if !found {
		return 0, fmt.Errorf("symbol '%s' not found in file", symbolName)
	}

	// Convert position to byte offset
	position := fset.Position(symbolPos)
	return position.Offset, nil
}

// runWithTextFileBusyRetry runs a command and retries if it fails with "text file busy" error.
// This error can occur when a binary was just installed by "go install" and the OS hasn't
// fully released the file handle yet. The function retries up to 3 times with exponential backoff.
// Note: This function uses cmd.Run(), so caller should capture output via cmd.Stdout/Stderr if needed.
func runWithTextFileBusyRetry(cmd *exec.Cmd) error {
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Create a new command for each retry since exec.Cmd can only be used once
		if attempt > 0 {
			// Wait before retrying with exponential backoff (100ms, 200ms, 400ms)
			waitTime := time.Duration(100*(1<<(attempt-1))) * time.Millisecond
			slog.Debug("Retrying command after text file busy error", "attempt", attempt+1, "wait", waitTime)
			time.Sleep(waitTime)

			newCmd := exec.Command(cmd.Path, cmd.Args[1:]...)
			newCmd.Dir = cmd.Dir
			newCmd.Env = cmd.Env
			newCmd.Stdout = cmd.Stdout
			newCmd.Stderr = cmd.Stderr
			cmd = newCmd
		}

		lastErr = cmd.Run()
		if lastErr == nil {
			return nil
		}

		// Check if this is a "text file busy" error
		if !strings.Contains(lastErr.Error(), "text file busy") {
			return lastErr
		}
	}

	return lastErr
}
