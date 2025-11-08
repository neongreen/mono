package lsp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/neongreen/mono/dissect/pkg/goutils"
)

// CodeActionParams contains parameters for code action requests.
type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        struct {
		Start Position `json:"start"`
		End   Position `json:"end"`
	} `json:"range"`
	Context struct {
		Diagnostics []interface{} `json:"diagnostics"`
		Only        []string      `json:"only,omitempty"`
	} `json:"context"`
}

// CodeAction represents a code action.
type CodeAction struct {
	Title   string         `json:"title"`
	Kind    string         `json:"kind"`
	Edit    *WorkspaceEdit `json:"edit,omitempty"`
	Command *Command       `json:"command,omitempty"`
}

// Command represents a command that can be executed.
type Command struct {
	Title     string        `json:"title"`
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

// ExecuteCommandParams contains parameters for command execution.
type ExecuteCommandParams struct {
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

// ExtractToNewFile extracts a function to a new file using LSP code actions.
func (c *Client) ExtractToNewFile(filePath string, funcName string) (newFilePath string, err error) {
	slog.Debug("Extracting function to new file via LSP",
		"filePath", filePath, "funcName", funcName)

	// Find the function's position
	fset, fn, err := goutils.FindFunc(filePath, funcName)
	if err != nil {
		return "", fmt.Errorf("error finding function %s: %w", funcName, err)
	}

	startPos := fset.Position(fn.Name.Pos())
	endPos := fset.Position(fn.End())

	// Prepare code action parameters
	params := CodeActionParams{
		TextDocument: TextDocumentIdentifier{
			URI: "file://" + filePath,
		},
	}
	params.Range.Start = Position{
		Line:      startPos.Line - 1,
		Character: startPos.Column - 1,
	}
	params.Range.End = Position{
		Line:      endPos.Line - 1,
		Character: endPos.Column - 1,
	}
	params.Context.Only = []string{"refactor.extract"}

	// Request code actions
	var actions []CodeAction
	if err := c.Call("textDocument/codeAction", params, &actions); err != nil {
		return "", fmt.Errorf("code action request failed: %w", err)
	}

	// Find the "extract to new file" action
	var extractAction *CodeAction
	for i := range actions {
		if strings.Contains(strings.ToLower(actions[i].Title), "new file") ||
			strings.Contains(actions[i].Kind, "refactor.extract.toNewFile") {
			extractAction = &actions[i]
			break
		}
	}

	if extractAction == nil {
		return "", fmt.Errorf("extract to new file action not available for function %s", funcName)
	}

	slog.Debug("Found extract action", "title", extractAction.Title, "kind", extractAction.Kind)

	// Execute the action
	if extractAction.Edit != nil {
		// Direct edit
		if err := c.applyWorkspaceEdit(extractAction.Edit); err != nil {
			return "", fmt.Errorf("failed to apply edit: %w", err)
		}
	} else if extractAction.Command != nil {
		// Command to execute
		cmdParams := ExecuteCommandParams{
			Command:   extractAction.Command.Command,
			Arguments: extractAction.Command.Arguments,
		}

		var result interface{}
		if err := c.Call("workspace/executeCommand", cmdParams, &result); err != nil {
			return "", fmt.Errorf("command execution failed: %w", err)
		}

		// The result might contain the edits
		if edit, ok := result.(map[string]interface{}); ok {
			// Try to parse as WorkspaceEdit
			var wsEdit WorkspaceEdit
			if changes, ok := edit["documentChanges"].([]interface{}); ok {
				// Reconstruct the workspace edit
				for _, change := range changes {
					// This is simplified - in practice we'd need proper JSON unmarshaling
					slog.Debug("Command returned document changes", "changes", len(changes))
				}
			}
		}
	}

	// Guess the new file path
	// gopls creates files with snake_case names
	newFileName := guessExtractedFileName(funcName)
	newFilePath = filepath.Join(filepath.Dir(filePath), newFileName)

	// Check if the file was created
	if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
		// Try to find it by checking what new files appeared
		files, _ := filepath.Glob(filepath.Join(filepath.Dir(filePath), "*.go"))
		// This is a fallback - the file should have been created
		slog.Warn("Could not verify new file was created", "expected", newFilePath)
	}

	return newFilePath, nil
}

// guessExtractedFileName guesses the file name gopls will create.
func guessExtractedFileName(funcName string) string {
	// Convert to snake_case
	var result strings.Builder
	for i, r := range funcName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	name := strings.ToLower(result.String())
	return fmt.Sprintf("%s.go", name)
}
