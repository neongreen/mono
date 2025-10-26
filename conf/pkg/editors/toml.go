package editors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neongreen/mono/lib/toml"
	tomlv2 "github.com/pelletier/go-toml/v2"
)

// TOMLEditor provides surgical editing of TOML files while preserving formatting
type TOMLEditor struct {
	filePath string
	dryRun   bool
}

// NewTOMLEditor creates a new TOML editor for the specified file
func NewTOMLEditor(filePath string) *TOMLEditor {
	return &TOMLEditor{
		filePath: filePath,
		dryRun:   false,
	}
}

// NewTOMLEditorWithDryRun creates a new TOML editor with dry-run mode
func NewTOMLEditorWithDryRun(filePath string, dryRun bool) *TOMLEditor {
	return &TOMLEditor{
		filePath: filePath,
		dryRun:   dryRun,
	}
}

// SetDryRun enables or disables dry-run mode
func (e *TOMLEditor) SetDryRun(dryRun bool) {
	e.dryRun = dryRun
}

// SetValue sets a value at the specified dotted path, preserving existing formatting
func (e *TOMLEditor) SetValue(path string, value interface{}) error {
	if e.dryRun {
		fmt.Printf("DRY RUN: Would set %s = %v in %s\n", path, value, e.filePath)
		return nil
	}

	// Read existing file content if it exists
	var content []byte
	var err error

	if _, err := os.Stat(e.filePath); err == nil {
		content, err = os.ReadFile(e.filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", e.filePath, err)
		}
	}

	// Parse existing TOML or create new document
	var doc *toml.Document
	if len(content) > 0 {
		doc, err = toml.Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse existing TOML: %w", err)
		}
	} else {
		// Create an empty document
		doc, err = toml.ParseString("")
		if err != nil {
			return fmt.Errorf("failed to create TOML document: %w", err)
		}
	}

	// Set the value at the specified path
	if err := doc.Set(path, value); err != nil {
		return fmt.Errorf("failed to set value at path %s: %w", path, err)
	}

	// Ensure directory exists
	dir := filepath.Dir(e.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the file with preserved formatting
	if err := os.WriteFile(e.filePath, doc.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GetValue retrieves a value at the specified dotted path
func (e *TOMLEditor) GetValue(path string) (interface{}, error) {
	content, err := os.ReadFile(e.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			if e.dryRun {
				fmt.Printf("DRY RUN: File %s does not exist\n", e.filePath)
			}
			return nil, fmt.Errorf("file does not exist: %s", e.filePath)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	doc, err := toml.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	value, err := doc.Get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get value at path %s: %w", path, err)
	}

	if value == nil {
		return nil, fmt.Errorf("path %s does not exist", path)
	}

	return value, nil
}

// UnsetValue removes a value at the specified dotted path
func (e *TOMLEditor) UnsetValue(path string) error {
	if e.dryRun {
		fmt.Printf("DRY RUN: Would unset %s in %s\n", path, e.filePath)
		return nil
	}

	content, err := os.ReadFile(e.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to unset
		}
		return fmt.Errorf("failed to read file: %w", err)
	}

	doc, err := toml.Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse TOML: %w", err)
	}

	if err := doc.Delete(path); err != nil {
		return fmt.Errorf("failed to delete value at path %s: %w", path, err)
	}

	// Write the file
	if err := os.WriteFile(e.filePath, doc.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// PreviewSetValue shows what setting a value would do without actually doing it
func (e *TOMLEditor) PreviewSetValue(path string, value interface{}) (string, error) {
	var preview strings.Builder

	// Check if file exists
	_, err := os.Stat(e.filePath)
	if os.IsNotExist(err) {
		preview.WriteString(fmt.Sprintf("Would create new file: %s\n", e.filePath))
	} else {
		preview.WriteString(fmt.Sprintf("Would modify existing file: %s\n", e.filePath))
	}

	preview.WriteString(fmt.Sprintf("Operation: SET\n"))
	preview.WriteString(fmt.Sprintf("Path: %s\n", path))
	preview.WriteString(fmt.Sprintf("Value: %v (%T)\n", value, value))

	return preview.String(), nil
}

// PreviewUnsetValue shows what unsetting a value would do without actually doing it
func (e *TOMLEditor) PreviewUnsetValue(path string) (string, error) {
	var preview strings.Builder

	// Check if file exists
	_, err := os.Stat(e.filePath)
	if os.IsNotExist(err) {
		preview.WriteString(fmt.Sprintf("File does not exist: %s\n", e.filePath))
		preview.WriteString("Operation: No change needed\n")
		return preview.String(), nil
	}

	// Check if the path currently exists
	_, err = e.GetValue(path)
	if err != nil {
		preview.WriteString(fmt.Sprintf("Path does not exist in %s\n", e.filePath))
		preview.WriteString("Operation: No change needed\n")
		return preview.String(), nil
	}

	preview.WriteString(fmt.Sprintf("Would modify existing file: %s\n", e.filePath))
	preview.WriteString(fmt.Sprintf("Operation: UNSET\n"))
	preview.WriteString(fmt.Sprintf("Path: %s\n", path))

	return preview.String(), nil
}

// GetAllValues reads all values from the TOML file and returns them as a nested map
func (e *TOMLEditor) GetAllValues() (map[string]interface{}, error) {
	content, err := os.ReadFile(e.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty map if file doesn't exist
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse TOML into nested map structure
	var data map[string]interface{}
	if err := tomlv2.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	return data, nil
}
