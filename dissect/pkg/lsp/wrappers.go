package lsp

import (
	"fmt"
	"log/slog"
)

// RenameWithClient performs a rename operation using an existing LSP client.
// This is a convenience wrapper that handles document opening/closing.
func RenameWithClient(client *Client, filePath string, oldName string, newName string) error {
	// Open the document if not already open
	uri := "file://" + filePath
	client.mu.Lock()
	_, isOpen := client.openDocs[uri]
	client.mu.Unlock()

	if !isOpen {
		if err := client.OpenDocument(filePath); err != nil {
			return fmt.Errorf("failed to open document: %w", err)
		}
		defer func() {
			if err := client.CloseDocument(filePath); err != nil {
				slog.Warn("Failed to close document", "file", filePath, "error", err)
			}
		}()
	}

	return client.Rename(filePath, oldName, newName)
}

// ExtractToNewFileWithClient performs an extract-to-new-file operation using an existing LSP client.
// This is a convenience wrapper that handles document opening/closing.
func ExtractToNewFileWithClient(client *Client, filePath string, funcName string) (string, error) {
	// Open the document if not already open
	uri := "file://" + filePath
	client.mu.Lock()
	_, isOpen := client.openDocs[uri]
	client.mu.Unlock()

	if !isOpen {
		if err := client.OpenDocument(filePath); err != nil {
			return "", fmt.Errorf("failed to open document: %w", err)
		}
		defer func() {
			if err := client.CloseDocument(filePath); err != nil {
				slog.Warn("Failed to close document", "file", filePath, "error", err)
			}
		}()
	}

	return client.ExtractToNewFile(filePath, funcName)
}

// AddImportWithClient adds an import using an existing LSP client.
// This is a convenience wrapper that handles document opening/closing.
func AddImportWithClient(client *Client, filePath string, importPath string) error {
	// Open the document if not already open
	uri := "file://" + filePath
	client.mu.Lock()
	_, isOpen := client.openDocs[uri]
	client.mu.Unlock()

	if !isOpen {
		if err := client.OpenDocument(filePath); err != nil {
			return fmt.Errorf("failed to open document: %w", err)
		}
		defer func() {
			if err := client.CloseDocument(filePath); err != nil {
				slog.Warn("Failed to close document", "file", filePath, "error", err)
			}
		}()
	}

	return client.AddImport(filePath, importPath)
}
