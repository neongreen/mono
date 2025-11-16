package lsp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

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
// The source file must be opened with OpenDocument first.
func (c *Client) ExtractToNewFile(filePath string, funcName string) (newFilePath string, err error) {
	slog.Debug("Extracting function to new file via LSP",
		"filePath", filePath, "funcName", funcName)

	// Find the function's position
	fset, fn, err := goutils.FindFunc(filePath, funcName)
	if err != nil {
		return "", fmt.Errorf("error finding function %s: %w", funcName, err)
	}

	startPos := fset.Position(fn.Pos())
	endPos := fset.Position(fn.End())

	// Prepare code action parameters - request actions for the entire function
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
	// Request only refactor.extract code actions
	params.Context.Only = []string{"refactor.extract"}

	// Request code actions
	var actions []CodeAction
	if err := c.Call("textDocument/codeAction", params, &actions); err != nil {
		return "", fmt.Errorf("code action request failed: %w", err)
	}

	slog.Debug("Received code actions", "count", len(actions))

	// Find the "extract to new file" action
	var extractAction *CodeAction
	for i := range actions {
		slog.Debug("Code action", "title", actions[i].Title, "kind", actions[i].Kind)
		// Look for the extract to new file action
		if actions[i].Kind == "refactor.extract.toNewFile" {
			extractAction = &actions[i]
			break
		}
	}

	if extractAction == nil {
		return "", fmt.Errorf("extract to new file action not available for function %s (got %d actions)", funcName, len(actions))
	}

	slog.Debug("Found extract action", "title", extractAction.Title, "kind", extractAction.Kind)

	// Execute the action
	if extractAction.Command != nil {
		// Execute the command
		cmdParams := ExecuteCommandParams{
			Command:   extractAction.Command.Command,
			Arguments: extractAction.Command.Arguments,
		}

		// The result might be a WorkspaceEdit or null
		var result interface{}
		if err := c.Call("workspace/executeCommand", cmdParams, &result); err != nil {
			return "", fmt.Errorf("command execution failed: %w", err)
		}

		slog.Debug("Command executed", "result", result != nil)

		// Gopls applies the edits automatically in -write mode, but we need to find what file was created
		// The command typically doesn't return the workspace edit when executed with -write flag
	} else if extractAction.Edit != nil {
		// Direct edit (less common for this action)
		if err := c.applyWorkspaceEdit(extractAction.Edit); err != nil {
			return "", fmt.Errorf("failed to apply edit: %w", err)
		}
	}

	// Guess the new file path based on gopls's naming convention
	newFileName := guessGoplsFileName(funcName)
	newFilePath = filepath.Join(filepath.Dir(filePath), newFileName)

	// Verify the file was created
	if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
		// Try to find any new .go file in the directory
		dirPath := filepath.Dir(filePath)
		entries, _ := os.ReadDir(dirPath)
		slog.Debug("Searching for new file in directory", "dir", dirPath)
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".go" {
				candidate := filepath.Join(dirPath, entry.Name())
				if candidate != filePath {
					info, _ := entry.Info()
					slog.Debug("Found potential new file", "path", candidate, "modTime", info.ModTime())
				}
			}
		}
		return "", fmt.Errorf("new file not found at expected path: %s", newFilePath)
	}

	// Open the new file in LSP so future operations can work on it
	if err := c.OpenDocument(newFilePath); err != nil {
		slog.Warn("Failed to open new file in LSP", "file", newFilePath, "error", err)
	}

	// Update the source file since it was modified
	if err := c.UpdateDocument(filePath); err != nil {
		slog.Warn("Failed to update source file in LSP", "file", filePath, "error", err)
	}

	slog.Debug("Successfully extracted function", "newFile", newFilePath)
	return newFilePath, nil
}

// guessGoplsFileName guesses the file name gopls will create.
// gopls converts to lowercase and keeps underscores and digits.
func guessGoplsFileName(funcName string) string {
	// Convert to lowercase, keeping underscores and digits
	name := ""
	for _, r := range funcName {
		if r >= 'A' && r <= 'Z' {
			name += string(r + 32) // Convert to lowercase
		} else if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' {
			name += string(r)
		}
	}
	return name + ".go"
}
