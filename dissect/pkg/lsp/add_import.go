package lsp

import (
	"fmt"
	"log/slog"
)

// AddImport adds an import to a Go file using the gopls.add_import command.
// The file must be opened with OpenDocument first.
func (c *Client) AddImport(filePath string, importPath string) error {
	slog.Debug("Adding import via LSP", "import", importPath, "file", filePath)

	// Prepare command parameters
	params := ExecuteCommandParams{
		Command: "gopls.add_import",
		Arguments: []interface{}{
			map[string]interface{}{
				"ImportPath": importPath,
				"URI":        "file://" + filePath,
			},
		},
	}

	// Execute the command
	var result interface{}
	if err := c.Call("workspace/executeCommand", params, &result); err != nil {
		return fmt.Errorf("add import command failed: %w", err)
	}

	// Update the document to reflect the changes
	if err := c.UpdateDocument(filePath); err != nil {
		slog.Warn("Failed to update document after adding import", "file", filePath, "error", err)
	}

	slog.Debug("Successfully added import", "import", importPath, "file", filePath)
	return nil
}
