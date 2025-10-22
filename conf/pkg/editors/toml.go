package editors

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
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

	// Parse existing TOML or create new structure
	var data map[string]interface{}
	if len(content) > 0 {
		if err := toml.Unmarshal(content, &data); err != nil {
			return fmt.Errorf("failed to parse existing TOML: %w", err)
		}
		// Ensure data is never nil, even if TOML only contains comments
		if data == nil {
			data = make(map[string]interface{})
		}
	} else {
		data = make(map[string]interface{})
	}

	// Set the value at the specified path
	if err := setNestedValue(data, path, value); err != nil {
		return fmt.Errorf("failed to set value at path %s: %w", path, err)
	}

	// Marshal back to TOML
	newContent, err := toml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal TOML: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(strings.TrimSuffix(e.filePath, "/"+getFileName(e.filePath)), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(e.filePath, newContent, 0644); err != nil {
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

	var data map[string]interface{}
	if err := toml.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	return getNestedValue(data, path)
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

	var data map[string]interface{}
	if err := toml.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("failed to parse TOML: %w", err)
	}

	if err := unsetNestedValue(data, path); err != nil {
		return fmt.Errorf("failed to unset value at path %s: %w", path, err)
	}

	// Marshal back to TOML
	newContent, err := toml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal TOML: %w", err)
	}

	// Write the file
	if err := os.WriteFile(e.filePath, newContent, 0644); err != nil {
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

// setNestedValue sets a value in a nested map structure using a dotted path
func setNestedValue(data map[string]interface{}, path string, value interface{}) error {
	keys := strings.Split(path, ".")
	if len(keys) == 0 {
		return fmt.Errorf("empty path")
	}

	current := data
	for i, key := range keys[:len(keys)-1] {
		if key == "" {
			return fmt.Errorf("empty key at position %d in path %s", i, path)
		}

		if existing, exists := current[key]; exists {
			if nested, ok := existing.(map[string]interface{}); ok {
				current = nested
			} else {
				// Convert existing value to map if needed
				current[key] = make(map[string]interface{})
				current = current[key].(map[string]interface{})
			}
		} else {
			newMap := make(map[string]interface{})
			current[key] = newMap
			current = newMap
		}
	}

	finalKey := keys[len(keys)-1]
	if finalKey == "" {
		return fmt.Errorf("empty final key in path %s", path)
	}

	current[finalKey] = value
	return nil
}

// getNestedValue retrieves a value from a nested map structure using a dotted path
func getNestedValue(data map[string]interface{}, path string) (interface{}, error) {
	keys := strings.Split(path, ".")
	if len(keys) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	current := data
	for i, key := range keys[:len(keys)-1] {
		if key == "" {
			return nil, fmt.Errorf("empty key at position %d in path %s", i, path)
		}

		if existing, exists := current[key]; exists {
			if nested, ok := existing.(map[string]interface{}); ok {
				current = nested
			} else {
				return nil, fmt.Errorf("path %s does not exist: key %s is not a map", path, key)
			}
		} else {
			return nil, fmt.Errorf("path %s does not exist: key %s not found", path, key)
		}
	}

	finalKey := keys[len(keys)-1]
	if finalKey == "" {
		return nil, fmt.Errorf("empty final key in path %s", path)
	}

	if value, exists := current[finalKey]; exists {
		return value, nil
	}

	return nil, fmt.Errorf("path %s does not exist: final key %s not found", path, finalKey)
}

// unsetNestedValue removes a value from a nested map structure using a dotted path
func unsetNestedValue(data map[string]interface{}, path string) error {
	keys := strings.Split(path, ".")
	if len(keys) == 0 {
		return fmt.Errorf("empty path")
	}

	current := data
	for i, key := range keys[:len(keys)-1] {
		if key == "" {
			return fmt.Errorf("empty key at position %d in path %s", i, path)
		}

		if existing, exists := current[key]; exists {
			if nested, ok := existing.(map[string]interface{}); ok {
				current = nested
			} else {
				return fmt.Errorf("path %s does not exist: key %s is not a map", path, key)
			}
		} else {
			return nil // Path doesn't exist, nothing to unset
		}
	}

	finalKey := keys[len(keys)-1]
	if finalKey == "" {
		return fmt.Errorf("empty final key in path %s", path)
	}

	delete(current, finalKey)
	return nil
}

// getFileName extracts filename from a file path
func getFileName(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
