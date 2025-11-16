package lsp

import (
	"fmt"
	"log/slog"
	"os"
)

// TextDocumentItem represents a document that is open in the editor.
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// DidOpenTextDocumentParams contains parameters for textDocument/didOpen.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// DidCloseTextDocumentParams contains parameters for textDocument/didClose.
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// VersionedTextDocumentIdentifier identifies a versioned text document.
type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

// TextDocumentContentChangeEvent represents a change to a text document.
type TextDocumentContentChangeEvent struct {
	Text string `json:"text"` // Full document text
}

// DidChangeTextDocumentParams contains parameters for textDocument/didChange.
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// OpenDocument opens a document in the language server.
// This must be called before performing operations on a file.
func (c *Client) OpenDocument(filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	uri := "file://" + filePath
	
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        uri,
			LanguageID: "go",
			Version:    1,
			Text:       string(content),
		},
	}

	if err := c.Notify("textDocument/didOpen", params); err != nil {
		return fmt.Errorf("failed to open document: %w", err)
	}

	c.mu.Lock()
	c.openDocs[uri] = 1 // Track version
	c.mu.Unlock()

	slog.Debug("Opened document in LSP", "uri", uri)
	return nil
}

// CloseDocument closes a document in the language server.
func (c *Client) CloseDocument(filePath string) error {
	uri := "file://" + filePath

	params := DidCloseTextDocumentParams{
		TextDocument: TextDocumentIdentifier{
			URI: uri,
		},
	}

	if err := c.Notify("textDocument/didClose", params); err != nil {
		return fmt.Errorf("failed to close document: %w", err)
	}

	c.mu.Lock()
	delete(c.openDocs, uri)
	c.mu.Unlock()

	slog.Debug("Closed document in LSP", "uri", uri)
	return nil
}

// UpdateDocument updates the content of an open document.
// The file must have been opened with OpenDocument first.
func (c *Client) UpdateDocument(filePath string) error {
	// Read current file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	uri := "file://" + filePath

	c.mu.Lock()
	version, ok := c.openDocs[uri]
	if !ok {
		c.mu.Unlock()
		return fmt.Errorf("document not open: %s", uri)
	}
	version++
	c.openDocs[uri] = version
	c.mu.Unlock()

	params := DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{
			URI:     uri,
			Version: version,
		},
		ContentChanges: []TextDocumentContentChangeEvent{
			{Text: string(content)}, // Full sync
		},
	}

	if err := c.Notify("textDocument/didChange", params); err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	slog.Debug("Updated document in LSP", "uri", uri, "version", version)
	return nil
}
