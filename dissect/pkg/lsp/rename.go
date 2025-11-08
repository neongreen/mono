package lsp

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
)

// TextDocumentIdentifier identifies a text document.
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// Position represents a position in a text document.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// TextDocumentPositionParams contains parameters for position-based operations.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// RenameParams contains parameters for the rename request.
type RenameParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	NewName      string                 `json:"newName"`
}

// TextEdit represents a textual edit to a document.
type TextEdit struct {
	Range struct {
		Start Position `json:"start"`
		End   Position `json:"end"`
	} `json:"range"`
	NewText string `json:"newText"`
}

// TextDocumentEdit represents edits to a text document.
type TextDocumentEdit struct {
	TextDocument struct {
		URI     string `json:"uri"`
		Version int    `json:"version,omitempty"`
	} `json:"textDocument"`
	Edits []TextEdit `json:"edits"`
}

// WorkspaceEdit represents changes to multiple resources.
type WorkspaceEdit struct {
	DocumentChanges []TextDocumentEdit `json:"documentChanges"`
}

// Rename renames a symbol in a Go file using the LSP server.
// The file must be opened with OpenDocument first.
func (c *Client) Rename(filePath string, oldName string, newName string) error {
	slog.Debug("Renaming symbol via LSP", "file", filePath, "old", oldName, "new", newName)

	// Validate new name
	if newName == "" {
		return fmt.Errorf("new name cannot be empty")
	}

	// Find the symbol's position in the file
	position, err := findSymbolPosition(filePath, oldName)
	if err != nil {
		return fmt.Errorf("error finding symbol '%s': %w", oldName, err)
	}

	// Prepare rename parameters
	params := RenameParams{
		TextDocument: TextDocumentIdentifier{
			URI: "file://" + filePath,
		},
		Position: *position,
		NewName:  newName,
	}

	// Call the rename method
	var result WorkspaceEdit
	if err := c.Call("textDocument/rename", params, &result); err != nil {
		return fmt.Errorf("rename failed: %w", err)
	}

	// Apply the edits to the filesystem
	if err := c.applyWorkspaceEdit(&result); err != nil {
		return fmt.Errorf("failed to apply edits: %w", err)
	}

	// Update all affected documents in LSP
	affectedFiles := make(map[string]struct{})
	for _, docChange := range result.DocumentChanges {
		uri := docChange.TextDocument.URI
		if len(uri) >= 7 && uri[:7] == "file://" {
			affectedFiles[uri[7:]] = struct{}{}
		}
	}

	for filePath := range affectedFiles {
		// Check if document is open
		c.mu.Lock()
		_, isOpen := c.openDocs["file://"+filePath]
		c.mu.Unlock()

		if isOpen {
			if err := c.UpdateDocument(filePath); err != nil {
				slog.Warn("Failed to update document after rename", "file", filePath, "error", err)
			}
		}
	}

	slog.Debug("Successfully renamed symbol", "old", oldName, "new", newName)
	return nil
}

// findSymbolPosition finds the LSP position of a symbol's definition in a file.
func findSymbolPosition(filePath string, symbolName string) (*Position, error) {
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing file: %w", err)
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
		return nil, fmt.Errorf("symbol '%s' not found in file", symbolName)
	}

	// Convert position to LSP format (0-based line and character)
	position := fset.Position(symbolPos)
	return &Position{
		Line:      position.Line - 1,      // LSP uses 0-based lines
		Character: position.Column - 1,     // LSP uses 0-based columns
	}, nil
}

// applyWorkspaceEdit applies a workspace edit to the filesystem.
func (c *Client) applyWorkspaceEdit(edit *WorkspaceEdit) error {
	for _, docChange := range edit.DocumentChanges {
		// Parse the URI to get file path
		uri := docChange.TextDocument.URI
		if len(uri) < 7 || uri[:7] != "file://" {
			return fmt.Errorf("invalid URI: %s", uri)
		}
		filePath := uri[7:]

		// Read the file
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		// Apply edits in reverse order to maintain positions
		// (LSP guarantees edits don't overlap and are sorted)
		lines := splitLines(string(content))
		
		// Apply edits from last to first to preserve positions
		for i := len(docChange.Edits) - 1; i >= 0; i-- {
			edit := docChange.Edits[i]
			lines = applyEdit(lines, edit)
		}

		// Write back
		newContent := joinLines(lines)
		if err := os.WriteFile(filePath, []byte(newContent), 0o644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}

		slog.Debug("Applied edits to file", "file", filePath, "editCount", len(docChange.Edits))
	}

	return nil
}

// splitLines splits content into lines, preserving line endings.
func splitLines(content string) []string {
	var lines []string
	start := 0
	
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			lines = append(lines, content[start:i+1])
			start = i + 1
		}
	}
	
	if start < len(content) {
		lines = append(lines, content[start:])
	}
	
	return lines
}

// joinLines joins lines back into content.
func joinLines(lines []string) string {
	result := ""
	for _, line := range lines {
		result += line
	}
	return result
}

// applyEdit applies a single TextEdit to lines.
func applyEdit(lines []string, edit TextEdit) []string {
	startLine := edit.Range.Start.Line
	startChar := edit.Range.Start.Character
	endLine := edit.Range.End.Line
	endChar := edit.Range.End.Character

	if startLine < 0 || startLine >= len(lines) {
		return lines
	}

	// Single line edit
	if startLine == endLine {
		line := lines[startLine]
		// Remove line ending temporarily
		lineEnd := ""
		if len(line) > 0 && line[len(line)-1] == '\n' {
			lineEnd = "\n"
			line = line[:len(line)-1]
		}
		
		newLine := line[:startChar] + edit.NewText + line[endChar:] + lineEnd
		lines[startLine] = newLine
		return lines
	}

	// Multi-line edit
	result := make([]string, 0, len(lines))
	
	// Lines before the edit
	result = append(result, lines[:startLine]...)
	
	// Bounds check
	if startLine >= len(lines) {
		return lines
	}
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}
	
	// First line of edit
	firstLine := lines[startLine]
	lineEnd := ""
	if len(firstLine) > 0 && firstLine[len(firstLine)-1] == '\n' {
		lineEnd = "\n"
		firstLine = firstLine[:len(firstLine)-1]
	}
	
	// Last line of edit
	lastLine := lines[endLine]
	if len(lastLine) > 0 && lastLine[len(lastLine)-1] == '\n' {
		lastLine = lastLine[:len(lastLine)-1]
	}
	
	// Bounds check for character positions
	if startChar > len(firstLine) {
		startChar = len(firstLine)
	}
	if endChar > len(lastLine) {
		endChar = len(lastLine)
	}
	
	// Create the merged line
	merged := firstLine[:startChar] + edit.NewText + lastLine[endChar:] + lineEnd
	result = append(result, merged)
	
	// Lines after the edit
	if endLine+1 < len(lines) {
		result = append(result, lines[endLine+1:]...)
	}
	
	return result
}
